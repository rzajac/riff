package riff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// ChunkLABL represents "labl" chunk which is always contained inside of
// an associated "LIST" chunk. It's used to associate a text label with
// a Cue Point. This information is often displayed next to markers or
// flags in digital audio editors.
type ChunkLABL struct {
	// Chunk size in bytes.
	// The ID and extra padding byte is not counted in the chunk size.
	size uint32

	// The Cue Point ID specifies the sample point that corresponds to
	// this text label by providing the ID of a Cue Point defined in the
	// Cue Point List. The ID that associates this label with a Cue Point
	// must be unique to all other label Cue Point IDs.
	CuePointID uint32

	// The label is a null terminated string of characters. If the number of
	// characters in the string is not even, padding must be appended to
	// the string. The appended padding is not considered in the label
	// chunk's chunk size field.
	label []byte
}

// LABLMake is a Maker function for creating ChunkLABL instances.
func LABLMake() Chunk { return LABL() }

// LABL returns new instance of ChunkLABL.
func LABL() *ChunkLABL {
	return &ChunkLABL{}
}

func (ch *ChunkLABL) ID() uint32     { return IDlabl }
func (ch *ChunkLABL) Size() uint32   { return ch.size }
func (ch *ChunkLABL) Type() uint32   { return 0 }
func (ch *ChunkLABL) Multi() bool    { return true }
func (ch *ChunkLABL) Chunks() Chunks { return nil }
func (ch *ChunkLABL) Raw() bool      { return false }

// Label returns label.
func (ch *ChunkLABL) Label() io.Reader {
	return bytes.NewReader(TrimZeroRight(ch.label))
}

func (ch *ChunkLABL) ReadFrom(r io.Reader) (int64, error) {
	var sum int64
	if err := binary.Read(r, le, &ch.size); err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, IDlabl), err)
	}
	sum += 4

	if err := binary.Read(r, le, &ch.CuePointID); err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, IDlabl), err)
	}
	sum += 4

	ch.label = grow(ch.label, int(ch.size-4)) // Subtract pid field size.
	in, err := io.ReadFull(r, ch.label)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, IDlabl), err)
	}

	// If length of label bytes is odd it means the
	// padding byte was added to the end.
	n, err := ReadPaddingIf(r, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, IDlabl), err)
	}

	return sum, nil
}

func (ch *ChunkLABL) WriteTo(w io.Writer) (int64, error) {
	var sum int64

	n, err := WriteIDAndSize(w, IDlabl, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, IDlabl), err)
	}

	if err := binary.Write(w, le, ch.CuePointID); err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, IDlabl), err)
	}
	sum += 4

	in, err := w.Write(ch.label)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, IDlabl), err)
	}

	n, err = WritePaddingIf(w, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, IDlabl), err)
	}

	return sum, nil
}

func (ch *ChunkLABL) Reset() {
	ch.size = 0
	ch.CuePointID = 0
	ch.label = ch.label[:0]
}
