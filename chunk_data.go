package riff

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

// IDdata represents "data" chunk ID.
const IDdata uint32 = 0x64617461

// ChunkDATA represents 'data' chunk decoder.
type ChunkDATA struct {
	// Chunk size in bytes.
	// The ID and extra padding byte is not counted in the chunk size.
	size uint32

	// Buffer data is read to.
	data []byte

	// Chunk processing options.
	opts Opts
}

func (ch *ChunkDATA) ID() uint32     { return IDdata }
func (ch *ChunkDATA) Size() uint32   { return ch.size }
func (ch *ChunkDATA) Type() uint32   { return 0 }
func (ch *ChunkDATA) Multi() bool    { return false }
func (ch *ChunkDATA) Chunks() Chunks { return nil }
func (ch *ChunkDATA) Raw() bool      { return false }

// DATAMake returns Maker function for ChunkDATA instances.
func DATAMake(opts ...OptsFn) Maker {
	return func() Chunk {
		return DATA(opts...)
	}
}

// DATA returns new instance of ChunkDATA. If load is false
// the chunk data will not be loaded into memory. It's used to reduce
// memory footprint of the decoder if only metadata is of interest.
func DATA(opts ...OptsFn) *ChunkDATA {
	ch := &ChunkDATA{
		opts: NewOpts(opts...),
	}
	if ch.opts.LoadData() {
		ch.data = make([]byte, 0, 1<<15)
	}
	return ch
}

// Data returns reader for data. If in SkipData mode, an empty reader is
// returned.
func (ch *ChunkDATA) Data() io.Reader {
	return bytes.NewReader(ch.data)
}

// SetData set data bytes. If will return ErrSkipDataMode if in SkipData mode.
func (ch *ChunkDATA) SetData(data []byte) error {
	if !ch.opts.LoadData() {
		return ErrSkipDataMode
	}
	l := len(data)
	ch.data = grow(ch.data, l)
	copy(ch.data, data)
	ch.size = uint32(l)
	return nil
}

// Duration returns file duration given average bit rate abr.
func (ch *ChunkDATA) Duration(abr uint32) time.Duration {
	dur := float64(ch.size) / float64(abr)
	dur *= float64(time.Second)
	return time.Duration(dur)
}

func (ch *ChunkDATA) ReadFrom(r io.Reader) (int64, error) {
	cr, size, err := ch.opts.ChunkReader(r)
	if err != nil {
		return 0, fmt.Errorf(errFmtDecode, Uint32(IDdata), err)
	}
	ch.size = size

	if !ch.opts.LoadData() {
		rs := RealSize(ch.size) // Skip padding byte if present.
		if err := cr.SkipN(rs); err != nil {
			return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDdata), err)
		}
		return cr.Count(), nil
	}

	ch.data = grow(ch.data, int(ch.size))
	_, err = io.ReadFull(cr, ch.data)
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDdata), err)
	}

	_, err = ReadPaddingIf(cr, ch.size)
	if err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDdata), err)
	}

	return cr.Count(), nil
}

func (ch *ChunkDATA) WriteTo(w io.Writer) (int64, error) {
	if !ch.opts.LoadData() {
		return 0, ErrSkipDataMode
	}

	var sum int64

	n, err := WriteIDAndSize(w, IDdata, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDdata), err)
	}

	in, err := w.Write(ch.data)
	sum += int64(in)
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDdata), err)
	}

	n, err = WritePaddingIf(w, ch.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDdata), err)
	}

	return sum, nil
}

// Reset resets the chunk so it can be reused.
func (ch *ChunkDATA) Reset() {
	ch.size = 0
	ch.data = ch.data[:0]
}
