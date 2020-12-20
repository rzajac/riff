package riff

import (
	"encoding/binary"
	"fmt"
	"io"
)

// IDLIST represents "LIST" chunk ID.
const IDLIST uint32 = 0x4c495354

// IDs of sub-chunks of LIST chunk.
const (
	// IDlabl represents LIST sub-chunk ID "labl".
	IDlabl uint32 = 0x6C61626C

	// IDadtl represents sub-chunk ID "adtl" of the LIST chunk.
	IDadtl uint32 = 0x6164746C
)

// ChunkLIST represents LIST chunk.
type ChunkLIST struct {
	// Chunk size in bytes.
	// The ID and extra padding byte is not counted in the chunk size.
	size uint32

	// List type.
	ListType uint32

	// Sub chunks.
	chunks Chunks

	// Registered chunk decoders.
	reg *Registry

	// When set to false decoder will try to skip reading the data.
	load bool
}

// LISTMake returns Maker function for creating ChunkLIST instances.
func LISTMake(load bool, reg *Registry) Maker {
	return func() Chunk {
		return LIST(load, reg)
	}
}

// LIST returns new instance of ChunkLIST.
func LIST(load bool, reg *Registry) *ChunkLIST {
	ch := &ChunkLIST{
		load: load,
		reg:  reg,
	}
	return ch
}

func (ch *ChunkLIST) ID() uint32     { return IDLIST }
func (ch *ChunkLIST) Size() uint32   { return ch.size }
func (ch *ChunkLIST) Type() uint32   { return ch.ListType }
func (ch *ChunkLIST) Multi() bool    { return true }
func (ch *ChunkLIST) Chunks() Chunks { return ch.chunks }
func (ch *ChunkLIST) Raw() bool      { return false }

func (ch *ChunkLIST) ReadFrom(r io.Reader) (int64, error) {
	var sum int64

	if err := binary.Read(r, le, &ch.size); err != nil {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDLIST), err)
	}
	sum += 4

	if err := binary.Read(r, be, &ch.ListType); err != nil {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDLIST), err)
	}
	sum += 4

	var mkr IDMaker
	switch ch.ListType {
	case IDINFO:
		mkr = INFOMake(ch.load)
	case IDadtl:
		ch.reg.Register(IDlabl, LABLMake)
		ch.reg.Register(IDltxt, LTXTMake)
		mkr = RAWCMake(ch.load)

	default:
		mkr = RAWCMake(ch.load)
	}

	var n int64
	var id uint32
	var err error

	for {
		if sum-4 >= int64(ch.size) {
			return sum, fmt.Errorf("invalid LIST chunk")
		}

		if err = ReadChunkID(r, &id); err != nil {
			return sum, err
		}
		sum += 4

		dec := ch.reg.GetNoRaw(id)
		if dec == nil {
			dec = mkr(id)
		}
		dec.Reset()

		n, err = dec.ReadFrom(r)
		sum += n
		if err != nil {
			return sum, fmt.Errorf(errFmtDecode, linkids(IDLIST, id), err)
		}
		ch.chunks = append(ch.chunks, dec)

		// Break the loop if we read all bytes declared in size.
		if sum-4 == int64(ch.size) {
			break
		}
	}

	n, err = ReadPaddingIf(r, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtDecode, linkids(IDLIST, ch.ListType), err)
	}

	return sum, nil
}

func (ch *ChunkLIST) WriteTo(w io.Writer) (int64, error) {
	if ch.load == SkipData {
		return 0, ErrSkipDataMode
	}

	var sum int64
	size := ch.chunks.Size() + 4 // Add four bytes for list type.

	n, err := WriteIDAndSize(w, IDLIST, size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDLIST), err)
	}

	if err := binary.Write(w, be, ch.ListType); err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDLIST), err)
	}
	sum += 4

	n, err = ch.chunks.WriteTo(w)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDLIST), err)
	}

	n, err = WritePaddingIf(w, size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDLIST), err)
	}

	return sum, nil
}

func (ch *ChunkLIST) Reset() {
	ch.size = 0
	ch.ListType = 0
	for _, dec := range ch.chunks {
		ch.reg.Put(dec)
	}
	ch.chunks = ch.chunks[:0]
}
