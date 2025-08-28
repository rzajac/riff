package riff

import (
	"encoding/binary"
	"io"
)

// OptsFn represent signature of option function.
type OptsFn func(*Opts)

// WithSizeLimit is an option setting maximal size of chunk in bytes.
func WithSizeLimit(size uint32) OptsFn {
	return func(o *Opts) { o.sizeLimit = size }
}

// WithLoadData is an option setting that causes chunks' content to be loaded.
func WithLoadData() OptsFn {
	return func(o *Opts) { o.loadData = true }
}

// WithIgnoreSize is an option setting that ignores declare chunk size and allows
// to read more bytes. It also allows to fix main chunk size declaration.
func WithIgnoreSize() OptsFn {
	return func(o *Opts) { o.ignoreDeclaredSize = true }
}

// WithOpts applies all values from given options.
func WithOpts(opts Opts) OptsFn {
	return func(o *Opts) {
		o.sizeLimit = opts.sizeLimit
		o.ignoreDeclaredSize = opts.ignoreDeclaredSize
		o.loadData = opts.loadData
	}
}

// Opts represents options for chunks handling.
type Opts struct {
	// Determines external maximal size of chunk in bytes. Zero value means
	// unlimited size.
	sizeLimit uint32

	// When enabled allows to read more data than indicated in chunk declaration.
	// By default, it is set to false.
	ignoreDeclaredSize bool

	// Controls how chunks are processed. If set to false, then only the
	// metadata about the chunks are read the rest is skipped. This improves
	// performance in cases when a user is only interested in metadata and
	// isn't going to modify or write the RIFF file.
	// It's up to the chunk decoder to decide what is considered data vs.
	// metadata.
	// By default, it is set to false.
	loadData bool
}

// LoadData getter.
func (o *Opts) LoadData() bool {
	return o.loadData
}

// IgnoreDeclaredSize getter.
func (o *Opts) IgnoreDeclaredSize() bool {
	return o.ignoreDeclaredSize
}

// NewOpts crates Opts instance.
func NewOpts(fns ...OptsFn) Opts {
	opts := Opts{}
	for _, fn := range fns {
		fn(&opts)
	}
	return opts
}

// ChunkReader reads first 4 bytes from reader as chunk size. Checks size
// constrains and return reader which should be used to read chunk's content.
func (o *Opts) ChunkReader(r io.Reader) (*ChunkReader, uint32, error) {
	var size uint32
	if err := binary.Read(r, le, &size); err != nil {
		return nil, 0, err
	}

	if o.sizeLimit != 0 && o.sizeLimit < RealSize(size) {
		return nil, 0, ErrTooLarge
	}

	cr := NewChunkReader(r)
	cr.setOffset(4) // Already read chunk size.

	if !o.IgnoreDeclaredSize() {
		cr.setLimit(int64(RealSize(size)))
	}
	if o.IgnoreDeclaredSize() && o.sizeLimit != 0 {
		cr.setLimit(int64(o.sizeLimit))
	}

	return cr, size, nil
}

// ChunkReader wraps an io.Reader, tracks the number of bytes read and allows to
// set read limit.
type ChunkReader struct {
	src    io.Reader // Original reader (may implement io.Seeker).
	rdr    io.Reader // Main reader (may have limit).
	cnt    int64     // Number of read bytes from reader.
	offset int64     // Shift for [Count] func.
}

// NewChunkReader creates a new ChunkReader.
func NewChunkReader(r io.Reader) *ChunkReader {
	return &ChunkReader{src: r, rdr: r}
}

// setOffset add shift to returned value of [Count].
func (cr *ChunkReader) setOffset(n int) {
	cr.offset = int64(n)
}

// setLimit constrains allowed number of bytes to read.
func (cr *ChunkReader) setLimit(n int64) {
	cr.rdr = io.LimitReader(cr.src, n)
}

// Read implements the io.Reader interface.
func (cr *ChunkReader) Read(p []byte) (n int, err error) {
	n, err = cr.rdr.Read(p)
	cr.cnt += int64(n)
	return n, err
}

// SkipN skips n bytes from reader.
// If the source implements [io.Seeker] SkipN will use it to skip n bytes,
// otherwise SkipN will read n bytes and discard them.
func (cr *ChunkReader) SkipN(n uint32) error {
	num := int64(n)

	if !cr.canRead(num) {
		return io.ErrUnexpectedEOF
	}

	// Handle nested chunk readers.
	if pcr, ok := cr.src.(*ChunkReader); ok {
		err := pcr.SkipN(n)
		if err != nil {
			return err
		}
		cr.moveCursors(num)
		return nil
	}

	// When src implements Seeker, we can just skip n bytes.
	if skr, ok := cr.src.(io.Seeker); ok {
		_, err := skr.Seek(num, io.SeekCurrent)
		if err != nil {
			return err
		}
		cr.moveCursors(num)
		return nil
	}

	// If we cannot seek, we read data to black hole.
	read, err := io.CopyN(io.Discard, cr, num)
	if err != nil {
		return err
	}
	if read < num {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// canRead checks whether the reader will not reach the limit when trying to read
// n bytes.
func (cr *ChunkReader) canRead(n int64) bool {
	lr, ok := cr.rdr.(*io.LimitedReader)
	if !ok {
		return true
	}
	return lr.N >= n
}

// moveCursors pushes cursors forward n bytes. Shift on reader must be done
// separately.
func (cr *ChunkReader) moveCursors(n int64) {
	// Add skipped bytes to counter.
	cr.cnt += n

	lr, ok := cr.rdr.(*io.LimitedReader)
	if ok {
		// Subtract skipped bytes from reader limit.
		lr.N -= n
	}
}

// Count return number of read bytes.
func (cr *ChunkReader) Count() int64 {
	return cr.cnt + cr.offset
}
