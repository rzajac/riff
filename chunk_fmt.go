package riff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// IDfmt represents "fmt " chunk ID.
const IDfmt uint32 = 0x666d7420

// FMTChunkSize represents the size of fmt chunk in bytes
// without ID and extra formatting bytes.
const FMTChunkSize uint32 = 16

// Compression codes.
const (
	CompNone uint16 = 0x0000 // Uncompressed PCM file.
	CompPCM  uint16 = 0x0001 // Microsoft Pulse Code Modulation (PCM).
)

// fmtStatic represents chunk static data (always there).
// This struct is defined separately to allow for binary
// decoding / encoding in one call to binary.Read / binary.Write.
type fmtStatic struct {
	// Compression code.
	//
	// See Comp* constants.
	CompCode uint16

	// Channel count indicates how many separate
	// audio signals are encoded in the data chunk.
	// Values: mono = 1, stereo = 2, etc.
	ChannelCnt uint16

	// Number of samples taken per second at each channel.
	// This value is unaffected by the number of channels.
	SampleRate uint32

	// How many bytes of waveform data must be streamed to a D/A converter
	// per second to play the waveform data.
	// AvgByteRate = SampleRate * BlockAlign.
	AvgByteRate uint32

	// The number of bytes per sample.
	// BlockAlign = round(BitsPerSample / 8) * ChannelCnt
	BlockAlign uint16

	// This value specifies the number of bits used to define each sample.
	// This value is usually 8, 16, 24 or 32. If the number of bits is
	// not byte aligned (a multiple of 8) then the number of bytes used
	// per sample is rounded up to the nearest byte size, and the unused
	// bytes are set to 0 and ignored.
	BitsPerSample uint16
}

// ChunkFMT represents a format chunk containing information about
// how the waveform data is stored.
type ChunkFMT struct {
	// Chunk size in bytes.
	// The ID and extra padding byte is not counted in the chunk size.
	size uint32

	fmtStatic

	// Extra bytes.
	// It doesn't exist if the compression code is 0 (uncompressed PCM file)
	// but may exist and have any value for other compression types
	// depending on what compression information is needed to decode the
	// waveform data.
	extra []byte

	// For some reason, even though there are no extra bytes, the
	// size of 0 is still written after the main 16-byte chunk. This value
	// controls this behavior. When set to true, the zero size will be written.
	WriteZeroExtra bool
}

func (ch *ChunkFMT) ID() uint32     { return IDfmt }
func (ch *ChunkFMT) Size() uint32   { return ch.size }
func (ch *ChunkFMT) Type() uint32   { return 0 }
func (ch *ChunkFMT) Multi() bool    { return false }
func (ch *ChunkFMT) Chunks() Chunks { return nil }
func (ch *ChunkFMT) Raw() bool      { return false }

// FMTMake is a Maker function for creating ChunkFMT instances.
func FMTMake() Chunk { return FMT() }

// FMT returns new instance of ChunkFMT.
func FMT() *ChunkFMT {
	return &ChunkFMT{
		size: FMTChunkSize,
	}
}

// Extra returns reader for extra format bytes.
func (ch *ChunkFMT) Extra() io.Reader {
	return bytes.NewReader(ch.extra)
}

// SetExtra set extra format bytes.
func (ch *ChunkFMT) SetExtra(extra []byte) {
	el := len(extra)
	ch.extra = grow(ch.extra, el)
	copy(ch.extra, extra)

	ch.size = FMTChunkSize
	if el > 0 {
		ch.size += 2 + RealSize(uint32(el))
	}
	if el == 0 && ch.WriteZeroExtra {
		ch.size += 2
	}
}

// Duration returns file duration given data size ds.
func (ch *ChunkFMT) Duration(ds uint32) time.Duration {
	dur := float64(ds) / float64(ch.AvgByteRate)
	dur *= float64(time.Second)
	return time.Duration(dur)
}

func (ch *ChunkFMT) ReadFrom(r io.Reader) (int64, error) {
	var sum int64
	if err := binary.Read(r, le, &ch.size); err != nil {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDfmt), err)
	}
	sum += 4

	if ch.size < FMTChunkSize {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDfmt), ErrTooShort)
	}

	if err := binary.Read(r, le, &ch.fmtStatic); err != nil {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDfmt), err)
	}
	sum += 16

	if ch.size > FMTChunkSize {
		// The first uint16 value in the byte slice is the uint16 length of
		// extra format bytes that will not be part of the ch.extra.
		//
		// If this value is not word-aligned (a multiple of 2),
		// padding should be added to the end of this data to word align it,
		// but the value should remain non-aligned.
		var es uint16
		if err := binary.Read(r, le, &es); err != nil {
			return sum, fmt.Errorf(errFmtDecode, Uint32(IDfmt), err)
		}
		sum += 2

		extra := int64(ch.size-FMTChunkSize) - 2 // Subtract uint16 size.
		if extra == 0 {
			ch.WriteZeroExtra = true
		}

		ch.extra = grow(ch.extra, int(es))
		in, err := io.ReadFull(r, ch.extra)
		sum += int64(in)
		if err != nil {
			return sum, fmt.Errorf(errFmtDecode, Uint32(IDfmt), err)
		}

		// If the length of extra format bytes is odd, it means the padding
		// byte was added to the end.
		n, err := ReadPaddingIf(r, uint32(es))
		sum += n
		if err != nil {
			return sum, fmt.Errorf(errFmtDecode, Uint32(IDfmt), err)
		}
	}

	n, err := ReadPaddingIf(r, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDfmt), err)
	}

	return sum, nil
}

func (ch *ChunkFMT) WriteTo(w io.Writer) (int64, error) {
	var sum int64

	size := uint32(16)
	eln := len(ch.extra)
	if eln > 0 || ch.WriteZeroExtra {
		// Adding 2 for extra format bytes size.
		size += RealSize(uint32(eln)) + 2
	}

	n, err := WriteIDAndSize(w, IDfmt, size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDfmt), err)
	}

	if err = binary.Write(w, le, ch.fmtStatic); err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDfmt), err)
	}
	sum += 16

	if eln > 0 || ch.WriteZeroExtra {
		// Write size of extra bytes.
		if err = binary.Write(w, le, uint16(eln)); err != nil {
			return sum, fmt.Errorf(errFmtEncode, Uint32(IDfmt), err)
		}
		sum += 2

		in, err := w.Write(ch.extra)
		sum += int64(in)
		if err != nil {
			return sum, fmt.Errorf(errFmtEncode, Uint32(IDfmt), err)
		}

		n, err = WritePaddingIf(w, uint32(eln))
		sum += n
		if err != nil {
			return sum, fmt.Errorf(errFmtEncode, Uint32(IDfmt), err)
		}
	}

	n, err = WritePaddingIf(w, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDfmt), err)
	}

	return sum, nil
}

func (ch *ChunkFMT) Reset() {
	ch.size = 16
	ch.CompCode = 0
	ch.ChannelCnt = 0
	ch.SampleRate = 0
	ch.AvgByteRate = 0
	ch.BlockAlign = 0
	ch.BitsPerSample = 0
	ch.extra = ch.extra[:0]
	ch.WriteZeroExtra = false
}
