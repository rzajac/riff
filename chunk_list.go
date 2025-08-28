package riff

import (
	"encoding/binary"
	"fmt"
	"io"
)

// IDLIST represents "LIST" chunk ID.
const IDLIST uint32 = 0x4c495354

// ListTypeSize represents the size of list type in bytes.
const ListTypeSize uint32 = 4

// IDs of sub-chunks of the LIST chunk.
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

	// Chunk processing options.
	opts Opts
}

// LISTMake returns [Maker] function for creating [ChunkLIST] instances.
func LISTMake(reg *Registry, opts ...OptsFn) Maker {
	return func() Chunk {
		return LIST(reg, opts...)
	}
}

// LIST returns a new instance of [ChunkLIST].
func LIST(reg *Registry, opts ...OptsFn) *ChunkLIST {
	ch := &ChunkLIST{
		reg:  reg,
		opts: NewOpts(opts...),
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
	cr, size, err := ch.opts.ChunkReader(r)
	if err != nil {
		return 0, fmt.Errorf(errFmtDecode, Uint32(IDLIST), err)
	}
	ch.size = size

	if ch.size < ListTypeSize {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDLIST), ErrTooShort)
	}

	if err := binary.Read(cr, be, &ch.ListType); err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDLIST), err)
	}

	var mkr IDMaker
	switch ch.ListType {
	case IDINFO:
		mkr = INFOMake(WithOpts(ch.opts))
	case IDadtl:
		ch.reg.Register(IDlabl, LABLMake(WithOpts(ch.opts)))
		ch.reg.Register(IDltxt, LTXTMake(WithOpts(ch.opts)))
		mkr = RAWCMake(WithOpts(ch.opts))

	default:
		mkr = RAWCMake(WithOpts(ch.opts))
	}

	for {
		if cr.Count()-4 >= int64(ch.size) {
			return cr.Count(), fmt.Errorf(errFmtDecode, linkids(IDLIST, ch.ListType), ErrChunkSizeMismatch)
		}

		var id uint32
		if err = ReadChunkID(cr, &id); err != nil {
			return cr.Count(), fmt.Errorf(errFmtDecode, linkids(IDLIST, ch.ListType), err)
		}

		dec := ch.reg.GetNoRaw(id)
		if dec == nil {
			dec = mkr(id)
		}
		dec.Reset()

		_, err = dec.ReadFrom(cr)
		if err != nil {
			return cr.Count(), fmt.Errorf(errFmtDecode, linkids(IDLIST, id), err)
		}
		ch.chunks = append(ch.chunks, dec)

		// Break the loop if we read all bytes declared in size.
		if cr.Count()-4 == int64(ch.size) {
			break
		}
	}

	return cr.Count(), nil
}

func (ch *ChunkLIST) WriteTo(w io.Writer) (int64, error) {
	if !ch.opts.LoadData() {
		return 0, ErrSkipDataMode
	}

	var sum int64
	size := ch.chunks.Size() + 4 // Add four bytes for the list type.

	n, err := WriteIDAndSize(w, IDLIST, size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDLIST), err)
	}

	if err = binary.Write(w, be, ch.ListType); err != nil {
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

// Modify set a new set of the chunks.
func (ch *ChunkLIST) Modify(chs Chunks) {
	ch.chunks = chs
	// Recalculate chunks size.
	ch.size = 4 + ch.chunks.Size()
}
