package riff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// IDsmpl represents "smpl" chunk ID.
const IDsmpl uint32 = 0x736d706c

// SMPLChunkSize represents the size of smpl chunk static part in bytes.
// Does not count ID, SampleLoops and sampleData bytes.
const SMPLChunkSize uint32 = 36

// SampleLoopCntSize represents the size of single sample loop in bytes.
const SampleLoopCntSize uint32 = 24

// smplStatic represents chunk static data (always there).
// This struct is defined separately to allow for binary
// decoding / encoding in one call to binary.Read / binary.Write.
type smplStatic struct {
	// The manufacturer field specifies the MIDI Manufacturer's
	// Association (MMA) Manufacturer code for the sampler intended
	// to receive this file's waveform. Each manufacturer of a MIDI
	// product is assigned a unique ID which identifies the company.
	// If no particular manufacturer is to be specified, a value of
	// 0 should be used.
	//
	// The value is stored with some extra information to enable
	// translation to the value used in a MIDI System Exclusive
	// transmission to the sampler. The high byte indicates the number
	// of low-order bytes (1 or 3) that are valid for the manufacturer
	// code. For example, the value for Digidesign will be 0x01000013 (0x13)
	// and the value for Microsoft will be 0x30000041 (0x00, 0x00, 0x41).
	//
	// See the MIDI Manufacturers List for a list:
	// https://www.midi.org/specifications-old/item/manufacturer-id-numbers
	Manufacturer uint32

	// The product field specifies the MIDI model ID defined by the
	// manufacturer corresponding to the Manufacturer field. Contact
	// the manufacturer of the sampler to get the model ID. If no
	// particular manufacturer's product is to be specified, a value
	// of 0 should be used.
	Product uint32

	// The sample period specifies the duration of time that passes during
	// the playback of one sample in nanoseconds (normally equal to
	// 1 / Samplers Per Second, where Samples Per Second is the value
	// found in the ChunkFMT chunk).
	SamplePeriod uint32

	// The MIDI unity note value has the same meaning as the instrument
	// chunk's MIDI Unshifted Note field which specifies the musical note
	// at which the sample will be played at it's original sample
	// rate (the sample rate specified in the ChunkFMT chunk).
	MIDIUnityNote uint32

	// The MIDI pitch fraction specifies the fraction of a semitone up
	// from the specified MIDI unity note field. A value of 0x80000000
	// means 1/2 semitone (50 cents), and a value of 0x00000000 means no
	// fine-tuning between semitones.
	MIDIPitchFraction uint32

	// The SMPTE format specifies the Society of Motion Pictures and
	// Television E time format used in the following SMPTE Offset field.
	// If a value of 0 is set, SMPTE Offset should also be set to 0.
	// SMPTE Format:
	//
	//     0 - no SMPTE offset
	//     24 - 24 frames per second
	//     25 - 25 frames per second
	//     29 - 30 frames per second with frame dropping (30 drop)
	//     30 - 30 frames per second
	//
	SMPTEFormat uint32

	// The SMPTE Offset value specifies the time offset to be used for
	// the synchronization / calibration to the first sample in the waveform.
	// This value uses a format of 0xhhmmssff where hh is a signed value
	// that specifies the number of hours (-23 to 23), mm is an unsigned
	// value that specifies the number of minutes (0 to 59), ss is an
	// unsigned value that specifies the number of seconds (0 to 59)
	// and ff is an unsigned value that specifies the number of
	// frames (0 to -1).
	SMPTEOffset uint32

	// The sample loops field specifies the number of Sample Loop definitions
	// in the following list. This value may be set to 0, meaning that no
	// sample loops follow.
	SampleLoopCnt uint32

	// The sampler data value specifies the number of bytes that will follow
	// this chunk (including the entire sample loop list). This value is
	// greater than 0 when an application needs to save additional
	// information. This value is reflected in this chunk data size value.
	SamplerDataCnt uint32
}

// ChunkSMPL represents smpl chunk.
//
// Source:
// https://sites.google.com/site/musicgapi/technical-documents/wav-file-format
type ChunkSMPL struct {
	// Chunk size in bytes.
	// The ID and extra padding byte is not counted in the chunk size.
	size uint32

	smplStatic

	// A list of sample loops is simply a set of consecutive loop descriptions.
	// The sample loops do not have to be in any particular order because
	// each sample loop associated cue point position is used to determine
	// the play order. The sampler chunk is optional.
	SampleLoops []*SampleLoop

	// Optional sampler specific data.
	sampleData []byte
}

// SMPLMake is a [Maker] function for creating [ChunkSMPL] instances.
func SMPLMake() Chunk { return SMPL() }

// SMPL returns a new instance of [ChunkSMPL].
func SMPL() *ChunkSMPL {
	return &ChunkSMPL{}
}

func (ch *ChunkSMPL) ID() uint32     { return IDsmpl }
func (ch *ChunkSMPL) Size() uint32   { return ch.size }
func (ch *ChunkSMPL) Type() uint32   { return 0 }
func (ch *ChunkSMPL) Multi() bool    { return false }
func (ch *ChunkSMPL) Chunks() Chunks { return nil }
func (ch *ChunkSMPL) Raw() bool      { return false }

// SamplerData returns reader for sampler specific data.
func (ch *ChunkSMPL) SamplerData() io.Reader {
	return bytes.NewReader(ch.sampleData)
}

func (ch *ChunkSMPL) ReadFrom(r io.Reader) (int64, error) {
	var sum int64
	if err := binary.Read(r, le, &ch.size); err != nil {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDsmpl), err)
	}
	sum += 4

	if ch.size < SMPLChunkSize {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDsmpl), ErrTooShort)
	}

	if err := binary.Read(r, le, &ch.smplStatic); err != nil {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDsmpl), err)
	}
	sum += int64(SMPLChunkSize)

	// We trust size more than SamplerDataCnt.
	extra := int(ch.size) - int(SMPLChunkSize) - int(ch.SampleLoopCnt*SampleLoopCntSize)
	if extra < 0 {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDsmpl), ErrChunkSizeMismatch)
	}

	for i := 0; i < int(ch.SampleLoopCnt); i++ {
		loop := sampleLoopPool.Get().(*SampleLoop) // nolint: forcetypeassert
		loop.Reset()
		if err := binary.Read(r, le, loop); err != nil {
			return 0, fmt.Errorf(errFmtDecode, Uint32(IDsmpl), err)
		}
		ch.SampleLoops = append(ch.SampleLoops, loop)
		sum += int64(SampleLoopCntSize)
	}

	ch.sampleData = grow(ch.sampleData, int(extra))
	in, err := io.ReadFull(r, ch.sampleData)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDdata), err)
	}

	n, err := ReadPaddingIf(r, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDsmpl), err)
	}

	return sum, nil
}

func (ch *ChunkSMPL) WriteTo(w io.Writer) (int64, error) {
	var sum int64

	size := SMPLChunkSize +
		uint32(len(ch.SampleLoops))*SampleLoopCntSize +
		uint32(len(ch.sampleData))

	ch.SampleLoopCnt = uint32(len(ch.SampleLoops))
	ch.SamplerDataCnt = ch.SampleLoopCnt*24 + uint32(len(ch.sampleData))

	n, err := WriteIDAndSize(w, IDsmpl, size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), err)
	}

	if err = binary.Write(w, le, ch.smplStatic); err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), err)
	}
	sum += int64(SMPLChunkSize)

	for i := 0; i < len(ch.SampleLoops); i++ {
		if err = binary.Write(w, le, ch.SampleLoops[i]); err != nil {
			return sum, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), err)
		}
		sum += int64(SampleLoopCntSize)
	}

	lsd := len(ch.sampleData)
	if lsd > 0 {
		if err = binary.Write(w, le, ch.sampleData); err != nil {
			return sum, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), err)
		}
		sum += int64(lsd)
	}

	n, err = WritePaddingIf(w, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDsmpl), err)
	}

	return sum, nil
}

func (ch *ChunkSMPL) Reset() {
	ch.size = 0
	ch.Manufacturer = 0
	ch.Product = 0
	ch.SamplePeriod = 0
	ch.MIDIUnityNote = 0
	ch.MIDIPitchFraction = 0
	ch.SMPTEFormat = 0
	ch.SMPTEOffset = 0
	ch.SampleLoopCnt = 0
	ch.SamplerDataCnt = 0
	for _, sl := range ch.SampleLoops {
		sampleLoopPool.Put(sl)
	}
	ch.SampleLoops = ch.SampleLoops[:0]
	ch.sampleData = ch.sampleData[:0]
}
