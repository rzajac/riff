package riff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// IDINFO represents LIST sub-chunk ID "INFO".
const IDINFO uint32 = 0x494e464f

// The sub-chunk INFO text labels.
// Source: http://bwfmetaedit.sourceforge.net/listinfo.html
const (
	// LabIARL is an Archival location.
	// Indicates where the subject of the file is archived.
	//
	// String value: "IARL"
	LabIARL uint32 = 0x4941524c

	// LabIART is an Artist.
	// Lists the artist of the original subject of the file.
	//
	// String value: "IART"
	LabIART uint32 = 0x49415254

	// LabICMS is the name of the person or organization that
	// commissioned the subject of the file.
	//
	// String value: "ICMS"
	LabICMS uint32 = 0x49434d53

	// LabICMT provides general comments about the file or the subject of the
	// file. If the comment is several sentences long, end each sentence with a
	// period. Do not include newline characters.
	//
	// String value: "ICMT"
	LabICMT uint32 = 0x49434d54

	// LabICOP records the copyright information for the file. For example,
	// Copyright Encyclopedia International 1991. If there are multiple
	// copyrights, separate them by a semicolon followed by a space.
	//
	// String value: "ICOP"
	LabICOP uint32 = 0x49434f50

	// LabICRD Specifies the date the subject of the file was created. List
	// dates in year-month-day format, padding one-digit months and days with a
	// zero on the left. For example, 1553-05-03 for May 3, 1553. The year
	// should always be given using four digits.
	//
	// String value: "ICRD"
	LabICRD uint32 = 0x49435244

	// LabIENG stores the name of the engineer who worked on the file. If there
	// are multiple engineers, separate the names by a semicolon and a blank.
	//
	// String value: "IENG"
	LabIENG uint32 = 0x49454e47

	// LabIGNR Describes the genre, such as jazz, classical, rock, etc.
	//
	// String value: "IGNR"
	LabIGNR uint32 = 0x49474e52

	// LabIKEY Provides a list of keywords that refer to the file or subject
	// of the file. Separate multiple keywords with a semicolon and a blank.
	//
	// String value: "IKEY"
	LabIKEY uint32 = 0x494b4559

	// LabIMED Describes the original subject of the file, such as record, CD
	// and so forth.
	//
	// String value: "IMED"
	LabIMED uint32 = 0x494d4544

	// LabINAM stores the title of the subject of the file.
	//
	// String value: "INAM"
	LabINAM uint32 = 0x494e414d

	// LabIPRD specifies the name of the title the file was originally intended
	// for (Album).
	//
	// String value: "IPRD"
	LabIPRD uint32 = 0x49505244

	// LabITRK
	//
	// String value: "ITRK"
	LabITRK uint32 = 0x4954524b

	// LabISBJ describes the contents of the file, such as ListItems Management.
	//
	// String value: "ISBJ"
	LabISBJ uint32 = 0x4953424a

	// LabISFT identifies the name of the software package used to create the
	// file, such as Audacity 1.3.9.
	//
	// String value: "ISFT"
	LabISFT uint32 = 0x746c7374

	// LabISRC identifies the name of the person or organization who supplied
	// the original subject of the file.
	//
	// String value: "ISRC"
	LabISRC uint32 = 0x49535243

	// LabISRF identifies the original form of the material that was digitized,
	// such as record, sampling CD, TV soundtrack and so forth.
	// This is not necessarily the same as IMED.
	//
	// String value: "ISRF"
	LabISRF uint32 = 0x49535246

	// LabITCH identifies the technician who sampled the subject file.
	//
	// String value: "ITCH"
	LabITCH uint32 = 0x49544348
)

// ChunkINFO represents INFO sub-chunk of the LIST chunk.
// Allows RIFF files to be "tagged" with information falling into
// a number of predefined categories.
type ChunkINFO struct {
	// Chunk ID (label ID).
	id uint32

	// Chunk size in bytes.
	// The ID and extra padding byte is not counted in the chunk size.
	size uint32

	// Label text.
	text []byte
}

// INFOMake returns [IDMaker] function for [ChunkINFO] instances.
func INFOMake(_ bool) IDMaker {
	return func(id uint32) Chunk {
		return INFO(id)
	}
}

// INFO returns a new instance of [ChunkINFO].
func INFO(id uint32) *ChunkINFO {
	return &ChunkINFO{
		id: id,
	}
}

func (ch *ChunkINFO) ID() uint32     { return ch.id }
func (ch *ChunkINFO) Size() uint32   { return ch.size }
func (ch *ChunkINFO) Type() uint32   { return 0 }
func (ch *ChunkINFO) Multi() bool    { return true }
func (ch *ChunkINFO) Chunks() Chunks { return nil }
func (ch *ChunkINFO) Raw() bool      { return false }

// Text returns INFO text.
func (ch *ChunkINFO) Text() io.Reader {
	return bytes.NewReader(TrimZeroRight(ch.text))
}

func (ch *ChunkINFO) ReadFrom(r io.Reader) (int64, error) {
	var sum int64

	if err := binary.Read(r, le, &ch.size); err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, ch.id), err)
	}
	sum += 4

	ch.text = grow(ch.text, int(ch.size))
	in, err := io.ReadFull(r, ch.text)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, ch.id), err)
	}

	// If the length of text bytes is odd, it means the padding byte was added
	// to the end.
	n, err := ReadPaddingIf(r, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDINFO, ch.id), err)
	}

	return sum, nil
}

func (ch *ChunkINFO) WriteTo(w io.Writer) (int64, error) {
	var sum int64

	n, err := WriteIDAndSize(w, ch.id, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, ch.id), err)
	}

	in, err := w.Write(ch.text)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, ch.id), err)
	}

	n, err = WritePaddingIf(w, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(IDINFO, ch.id), err)
	}

	return sum, nil
}

func (ch *ChunkINFO) Reset() {
	ch.size = 0
	ch.text = ch.text[:0]
}

// InfoLabel returns human-readable INFO sub-chunk label.
//
// nolint: cyclop
func InfoLabel(lab uint32) string {
	switch lab {
	case LabIARL:
		return "archival location"
	case LabIART:
		return "artist"
	case LabICMS:
		return "commissioned"
	case LabICMT:
		return "comments"
	case LabICOP:
		return "copyright"
	case LabICRD:
		return "creation date"
	case LabIENG:
		return "engineer"
	case LabIGNR:
		return "genre"
	case LabIKEY:
		return "keywords"
	case LabIMED:
		return "original medium"
	case LabINAM:
		return "title"
	case LabIPRD:
		return "album"
	case LabITRK:
		return "track"
	case LabISBJ:
		return "subject"
	case LabISFT:
		return "software"
	case LabISRC:
		return "source"
	case LabISRF:
		return "source form"
	case LabITCH:
		return "technician"
	default:
		return Uint32(lab).String()
	}
}
