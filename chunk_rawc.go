package riff

import (
	"bytes"
	"fmt"
	"io"
)

// idRAWC represents raw chunk ID used only in error reporting from RAWC chunk.
const idRAWC uint32 = 0x52415743

// ChunkRAWC is a decoder for unknown (not registered) chunks.
type ChunkRAWC struct {
	// Chunk ID.
	id uint32

	// Chunk size in bytes.
	// The ID and extra padding byte is not counted in the chunk size.
	size uint32

	// Buffer to read the chunk data into.
	data []byte

	// Chunk processing options.
	opts Opts
}

// RAWCMake returns [IDMaker] function for creating [ChunkRAWC] instances.
func RAWCMake(opts ...OptsFn) IDMaker {
	return func(id uint32) Chunk {
		return RAWC(id, opts...)
	}
}

// RAWC returns a new instance of [ChunkRAWC] for given ID.
func RAWC(id uint32, opts ...OptsFn) *ChunkRAWC {
	ch := &ChunkRAWC{
		id:   id,
		opts: NewOpts(opts...),
	}

	if ch.opts.LoadData() {
		ch.data = make([]byte, 0, 1<<8)
	}

	return ch
}

func (ch *ChunkRAWC) ID() uint32     { return ch.id }
func (ch *ChunkRAWC) Size() uint32   { return ch.size }
func (ch *ChunkRAWC) Type() uint32   { return 0 }
func (ch *ChunkRAWC) Multi() bool    { return true }
func (ch *ChunkRAWC) Chunks() Chunks { return nil }
func (ch *ChunkRAWC) Raw() bool      { return true }

func (ch *ChunkRAWC) Body() io.Reader {
	return bytes.NewReader(ch.data)
}

func (ch *ChunkRAWC) ReadFrom(r io.Reader) (int64, error) {
	cr, size, err := ch.opts.ChunkReader(r)
	if err != nil {
		return 0, fmt.Errorf(errFmtDecode, linkids(idRAWC, ch.id), err)
	}
	ch.size = size

	if !ch.opts.loadData {
		rs := RealSize(ch.size) // Skip padding byte if present.
		if err := cr.SkipN(rs); err != nil {
			return cr.Count(), fmt.Errorf(errFmtDecode, linkids(idRAWC, ch.id), err)
		}
		return cr.Count(), nil
	}

	ch.data = grow(ch.data, int(ch.size))
	_, err = io.ReadFull(cr, ch.data)
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, linkids(idRAWC, ch.id), err)
	}

	_, err = ReadPaddingIf(cr, ch.size)
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, linkids(idRAWC, ch.id), err)
	}

	return cr.Count(), nil
}

func (ch *ChunkRAWC) WriteTo(w io.Writer) (int64, error) {
	if !ch.opts.LoadData() {
		return 0, ErrSkipDataMode
	}

	var sum int64

	n, err := WriteIDAndSize(w, ch.id, uint32(len(ch.data)))
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(idRAWC, ch.id), err)
	}

	in, err := w.Write(ch.data)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(idRAWC, ch.id), err)
	}

	n, err = WritePaddingIf(w, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, linkids(idRAWC, ch.id), err)
	}

	return sum, nil
}

func (ch *ChunkRAWC) Reset() {
	ch.size = 0
	ch.data = ch.data[:0]
}
