package riff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// IDltxt represents LIST sub-chunk ID "ltxt".
const IDltxt uint32 = 0x6C747874

// ltxtStatic represents chunk static data (always there).
// This struct is defined separately to allow for binary
// decoding / encoding in one call to binary.Read / binary.Write.
type ltxtStatic struct {
	// Cue Point ID specifies the starting sample point that corresponds to
	// this text label by providing the ID of a Cue Point defined in the
	// Cue Point List. The ID that associates this label with a Cue Point
	// must be unique to all other note chunk Cue Point IDs.
	CuePointID uint32

	// The sample length defines how many samples from the cue point the
	// region or section spans.
	SamLen uint32

	// The purpose field specifies what the text is used for. For example
	// a value of "scrp" means script text, and "capt" means close-caption.
	// There are several more purpose IDs, but they are meant to be used
	// with other types of RIFF files (not usually found in WAVE files).
	PurID uint32

	// Information about the location used by the text and is typically
	// used for queries to obtain information from the operating system.
	Country uint16

	// Information about the language used by the text and is typically
	// used for queries to obtain information from the operating system.
	Language uint16

	// Information about the dialect used by the text and is typically
	// used for queries to obtain information from the operating system.
	Dialect uint16

	// Information about the code page used by the text and is typically
	// used for queries to obtain information from the operating system.
	CodePage uint16
}

// ChunkLTXT represents labeled text chunk. It is used to associate a
// text label with a region or section of waveform data. This information
// is often displayed in marked regions of a waveform in digital audio editors.
type ChunkLTXT struct {
	// Chunk size in bytes.
	// The ID and extra padding byte is not counted in the chunk size.
	size uint32

	ltxtStatic

	// Text is a null terminated string of characters. If the number of
	// characters in the string is not even, padding must be appended to
	// the string. The appended padding is not considered in the note
	// chunk's chunk size field.
	text []byte
}

// LTXTMake is a Maker function for creating ChunkLTXT instances.
func LTXTMake() Chunk { return LTXT() }

// LTXT returns new instance of ChunkLTXT.
func LTXT() *ChunkLTXT {
	return &ChunkLTXT{}
}

func (ch *ChunkLTXT) ID() uint32     { return IDltxt }
func (ch *ChunkLTXT) Size() uint32   { return ch.size }
func (ch *ChunkLTXT) Type() uint32   { return 0 }
func (ch *ChunkLTXT) Multi() bool    { return true }
func (ch *ChunkLTXT) Chunks() Chunks { return nil }
func (ch *ChunkLTXT) Raw() bool      { return false }

// Text returns reader for text field.
func (ch *ChunkLTXT) Text() io.Reader {
	return bytes.NewReader(TrimZeroRight(ch.text))
}

func (ch *ChunkLTXT) ReadFrom(r io.Reader) (int64, error) {
	var sum int64
	if err := binary.Read(r, le, &ch.size); err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, IDltxt), err)
	}
	sum += 4

	if err := binary.Read(r, le, &ch.ltxtStatic); err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, IDltxt), err)
	}
	sum += 20

	tl := int(ch.size - 20) // Subtract ltxtStatic fields size.
	ch.text = grow(ch.text, tl)
	in, err := io.ReadFull(r, ch.text)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, IDltxt), err)
	}

	// If length of text bytes is odd it means the
	// padding byte was added to the end.
	n, err := ReadPaddingIf(r, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, IDltxt), err)
	}
	return sum, nil
}

func (ch *ChunkLTXT) WriteTo(w io.Writer) (int64, error) {
	var sum int64

	n, err := WriteIDAndSize(w, IDltxt, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, IDltxt), err)
	}

	if err = binary.Write(w, le, ch.ltxtStatic); err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, IDltxt), err)
	}
	sum += 20

	in, err := w.Write(ch.text)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, IDltxt), err)
	}

	n, err = WritePaddingIf(w, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, IDltxt), err)
	}

	return sum, nil
}

func (ch *ChunkLTXT) Reset() {
	ch.size = 0
	ch.CuePointID = 0
	ch.SamLen = 0
	ch.PurID = 0
	ch.Country = 0
	ch.Language = 0
	ch.Dialect = 0
	ch.CodePage = 0
	ch.text = ch.text[:0]
}
