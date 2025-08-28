package riff

import (
	"bytes"
	"io"
	"testing"

	"github.com/ctx42/memfs/pkg/memfs"
	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/kit/iokit"
	"github.com/ctx42/testing/pkg/mock"
	"github.com/ctx42/testing/pkg/must"
	"github.com/rzajac/riff/internal/test"
)

func Test_NewOpts(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		// --- When ---
		opts := NewOpts()

		// --- Then ---
		assert.Equal(t, uint32(0), opts.sizeLimit)
		assert.False(t, opts.loadData)
		assert.False(t, opts.ignoreDeclaredSize)
		assert.False(t, opts.LoadData())
		assert.False(t, opts.IgnoreDeclaredSize())
	})

	t.Run("WithSizeLimit", func(t *testing.T) {

		// --- When ---
		opts := NewOpts(WithSizeLimit(1024))

		// --- Then ---
		assert.Equal(t, uint32(1024), opts.sizeLimit)
	})

	t.Run("WithLoadData", func(t *testing.T) {
		// --- When ---
		opts := NewOpts(WithLoadData())

		// --- Then ---
		assert.True(t, opts.loadData)
		assert.True(t, opts.LoadData())
	})

	t.Run("WithIgnoreSize", func(t *testing.T) {
		// --- When ---
		opts := NewOpts(WithIgnoreSize())

		// --- Then ---
		assert.True(t, opts.ignoreDeclaredSize)
		assert.True(t, opts.IgnoreDeclaredSize())
	})
}

func Test_Opts_ChunkReader_Read(t *testing.T) {
	t.Run("read whole chunk", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.WriteUint32LE(t, src, 8) // size 8
		test.WriteUint32LE(t, src, 1) // some data
		test.WriteUint32LE(t, src, 2) // some data

		opts := NewOpts()
		cr, size := must.Values(opts.ChunkReader(src))

		// --- When ---
		n, err := io.CopyN(io.Discard, cr, 8)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(8), n)

		assert.Equal(t, uint32(8), size)
		assert.Equal(t, int64(12), cr.Count())
	})

	t.Run("hit chunk declared size", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.WriteUint32LE(t, src, 4) // size 4
		test.WriteUint32LE(t, src, 1) // some data
		test.WriteUint32LE(t, src, 2) // some data

		opts := NewOpts()
		cr, size := must.Values(opts.ChunkReader(src))

		// --- When ---
		n, err := io.CopyN(io.Discard, cr, 5)

		// --- Then ---
		assert.ErrorIs(t, io.EOF, err)
		assert.Equal(t, int64(4), n)

		assert.Equal(t, uint32(4), size)
		assert.Equal(t, int64(8), cr.Count())
	})

	t.Run("limit supports ReadAll behaviour", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.WriteUint32LE(t, src, 4) // size 4
		test.WriteUint32LE(t, src, 1) // some data
		test.WriteUint32LE(t, src, 2) // some data
		test.WriteUint32LE(t, src, 3) // some data

		opts := NewOpts()
		cr, size := must.Values(opts.ChunkReader(src))

		// --- When ---
		have, err := io.ReadAll(cr)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, []byte{1, 0, 0, 0}, have)

		// Consecutive ReadAll call does not return error.
		have, err = io.ReadAll(cr)
		assert.NoError(t, err)
		assert.Empty(t, have)

		n, err := io.CopyN(io.Discard, cr, 1)
		assert.ErrorIs(t, io.EOF, err)
		assert.Equal(t, int64(0), n)

		assert.Equal(t, uint32(4), size)
		assert.Equal(t, int64(8), cr.Count())
	})

	t.Run("read whole chunk with size limit", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.WriteUint32LE(t, src, 8) // size 8
		test.WriteUint32LE(t, src, 1) // some data
		test.WriteUint32LE(t, src, 2) // some data

		opts := NewOpts(WithSizeLimit(8))
		cr, size := must.Values(opts.ChunkReader(src))

		// --- When ---
		n, err := io.CopyN(io.Discard, cr, 8)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(8), n)

		assert.Equal(t, uint32(8), size)
		assert.Equal(t, int64(12), cr.Count())
	})

	t.Run("hit chunk max size limit", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.WriteUint32LE(t, src, 8) // size 8
		test.WriteUint32LE(t, src, 1) // some data
		test.WriteUint32LE(t, src, 2) // some data

		opts := NewOpts(WithSizeLimit(7))

		// --- When ---
		cr, size, err := opts.ChunkReader(src)

		// --- Then ---
		assert.ErrorIs(t, ErrTooLarge, err)
		assert.Equal(t, uint32(0), size)
		assert.Nil(t, cr)
	})

	t.Run("ignore size declaration", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.WriteUint32LE(t, src, 4) // size 4
		test.WriteUint32LE(t, src, 1) // some data
		test.WriteUint32LE(t, src, 2) // some data

		opts := NewOpts(WithIgnoreSize())
		cr, size := must.Values(opts.ChunkReader(src))

		// --- When ---
		n, err := io.CopyN(io.Discard, cr, 8)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(8), n)

		assert.Equal(t, uint32(4), size)
		assert.Equal(t, int64(12), cr.Count())
	})

	t.Run("read size error", func(t *testing.T) {
		// --- Given ---
		src := iokit.NewReaderMock(t)
		src.OnRead(mock.Any).Return(0, iokit.ErrRead)

		opts := NewOpts()

		// --- When ---
		cr, size, err := opts.ChunkReader(src)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrRead, err)
		assert.Equal(t, uint32(0), size)
		assert.Nil(t, cr)
	})

	t.Run("read content error", func(t *testing.T) {
		// --- Given ---
		src := iokit.NewReaderMock(t)
		fn := mock.MatchBy(func(p []byte) bool {
			p[0] = 80 // chunk size
			return true
		})
		src.OnRead(fn).Return(4, nil).Once()
		src.OnRead(mock.Any).Return(0, iokit.ErrRead).Once()

		opts := NewOpts()
		cr, size := must.Values(opts.ChunkReader(src))

		// --- When ---
		n, err := io.CopyN(io.Discard, cr, 8)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrRead, err)
		assert.Equal(t, int64(0), n)

		assert.Equal(t, uint32(80), size)
		assert.Equal(t, int64(4), cr.Count())
	})

	t.Run("nested readers", func(t *testing.T) {
		// --- Given ---
		buf := bytes.NewBuffer(nil)
		test.WriteUint32LE(t, buf, 30) // chunk size
		for i := 0; i < 30; i++ {
			buf.WriteByte(byte(i))
		}

		opts := NewOpts()

		// --- When ---
		cr0, _ := must.Values(opts.ChunkReader(buf))
		must.Value(io.CopyN(io.Discard, cr0, 4))
		assert.Equal(t, int64(8), cr0.Count())

		cr1, _ := must.Values(opts.ChunkReader(cr0))
		must.Value(io.CopyN(io.Discard, cr1, 4))
		assert.Equal(t, int64(16), cr0.Count())
		assert.Equal(t, int64(8), cr1.Count())

		cr2, _ := must.Values(opts.ChunkReader(cr1))
		must.Value(io.CopyN(io.Discard, cr2, 4))
		assert.Equal(t, int64(24), cr0.Count())
		assert.Equal(t, int64(16), cr1.Count())
		assert.Equal(t, int64(8), cr2.Count())

		_, err := io.CopyN(io.Discard, cr2, 4)

		// --- Then ---
		assert.NoError(t, err)

		assert.Equal(t, int64(28), cr0.Count())
		assert.Equal(t, int64(20), cr1.Count())
		assert.Equal(t, int64(12), cr2.Count())
	})
}

func Test_Opts_ChunkReader_SkipN(t *testing.T) {
	t.Run("no seeker skips by reading", func(t *testing.T) {
		// --- Given ---
		buf := bytes.NewBuffer(nil)
		test.WriteUint32LE(t, buf, 30) // chunk size
		for i := 0; i < 30; i++ {
			buf.WriteByte(byte(i))
		}

		opts := NewOpts()
		cr, _ := must.Values(opts.ChunkReader(buf))

		// --- When ---
		err := cr.SkipN(20)

		// --- Then ---
		assert.NoError(t, err)
		exp := []byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29}
		assert.Equal(t, exp, must.Value(io.ReadAll(buf)))

		assert.Equal(t, int64(24), cr.Count())
	})

	t.Run("seeker ok", func(t *testing.T) {
		// --- Given ---
		buf := &memfs.File{}
		test.WriteUint32LE(t, buf, 30) // chunk size
		for i := 0; i < 30; i++ {
			_ = buf.WriteByte(byte(i))
		}
		buf.SeekStart()

		opts := NewOpts()
		cr, _ := must.Values(opts.ChunkReader(buf))
		assert.Equal(t, int64(4), iokit.Offset(buf))

		// --- When ---
		err := cr.SkipN(20)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(24), iokit.Offset(buf))

		exp := []byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29}
		assert.Equal(t, exp, must.Value(io.ReadAll(buf)))

		assert.Equal(t, int64(24), cr.Count())
	})

	t.Run("no seeker read error", func(t *testing.T) {
		// --- Given ---
		src := iokit.NewReaderMock(t)

		fn := mock.MatchBy(func(p []byte) bool {
			p[0] = 80 // chunk size
			return true
		})
		src.OnRead(fn).Return(4, nil).Once()
		src.OnRead(mock.Any).Return(0, iokit.ErrRead).Once()

		opts := NewOpts()
		cr, _ := must.Values(opts.ChunkReader(src))

		// --- When ---
		err := cr.SkipN(20)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrRead, err)

		assert.Equal(t, int64(4), cr.Count())
	})

	t.Run("seek error", func(t *testing.T) {
		// --- Given ---
		src := iokit.NewReadSeekerMock(t)

		fn := mock.MatchBy(func(p []byte) bool {
			p[0] = 80 // chunk size
			return true
		})
		src.OnRead(fn).Return(4, nil).Once()

		src.OnSeek(int64(20), io.SeekCurrent).
			Return(int64(0), iokit.ErrRead).Once()

		opts := NewOpts()
		cr, _ := must.Values(opts.ChunkReader(src))

		// --- When ---
		err := cr.SkipN(20)

		// --- Then ---
		assert.ErrorIs(t, iokit.ErrRead, err)

		assert.Equal(t, int64(4), cr.Count())
	})

	t.Run("no seeker multiple skips", func(t *testing.T) {
		// --- Given ---
		buf := bytes.NewBuffer(nil)
		test.WriteUint32LE(t, buf, 30) // chunk size
		for i := 0; i < 30; i++ {
			buf.WriteByte(byte(i))
		}

		opts := NewOpts()
		cr, _, err := opts.ChunkReader(buf)
		assert.NoError(t, err)

		// --- When ---
		err = cr.SkipN(4)
		assert.NoError(t, err)
		assert.Equal(t, int64(8), cr.Count())

		err = cr.SkipN(4)
		assert.NoError(t, err)
		assert.Equal(t, int64(12), cr.Count())

		err = cr.SkipN(4)
		assert.NoError(t, err)
		assert.Equal(t, int64(16), cr.Count())
	})

	t.Run("seeker multiple skips", func(t *testing.T) {
		// --- Given ---
		buf := &memfs.File{}
		test.WriteUint32LE(t, buf, 30) // chunk size
		for i := 0; i < 30; i++ {
			_ = buf.WriteByte(byte(i))
		}
		buf.SeekStart()

		opts := NewOpts()
		cr, _, err := opts.ChunkReader(buf)
		assert.NoError(t, err)

		// --- When ---
		err = cr.SkipN(4)
		assert.NoError(t, err)
		assert.Equal(t, int64(8), cr.Count())

		err = cr.SkipN(4)
		assert.NoError(t, err)
		assert.Equal(t, int64(12), cr.Count())

		err = cr.SkipN(4)
		assert.NoError(t, err)
		assert.Equal(t, int64(16), cr.Count())
	})

	t.Run("seeker skips reduces limit", func(t *testing.T) {
		// --- Given ---
		buf := &memfs.File{}
		test.WriteUint32LE(t, buf, 16) // chunk size
		for i := 0; i < 30; i++ {
			_ = buf.WriteByte(byte(i))
		}
		buf.SeekStart()

		opts := NewOpts()
		cr, _, err := opts.ChunkReader(buf)
		assert.NoError(t, err)

		// --- When ---
		err = cr.SkipN(10)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(14), cr.Count())

		// Reads all up to chunk limit.
		exp := []byte{10, 11, 12, 13, 14, 15}
		assert.Equal(t, exp, must.Value(io.ReadAll(cr)))

		// Next reads returns EOF.
		_, err = io.CopyN(io.Discard, cr, 1)
		assert.ErrorIs(t, io.EOF, err)

		assert.Equal(t, int64(20), cr.Count())
	})

	t.Run("nested readers", func(t *testing.T) {
		// --- Given ---
		buf := &memfs.File{}
		test.WriteUint32LE(t, buf, 30) // chunk size
		for i := 0; i < 30; i++ {
			_ = buf.WriteByte(byte(i))
		}
		buf.SeekStart()

		opts := NewOpts()

		// --- When ---
		cr0, _ := must.Values(opts.ChunkReader(buf))
		must.Nil(cr0.SkipN(4))
		assert.Equal(t, int64(8), cr0.Count())

		cr1, _ := must.Values(opts.ChunkReader(cr0))
		must.Nil(cr1.SkipN(4))
		assert.Equal(t, int64(16), cr0.Count())
		assert.Equal(t, int64(8), cr1.Count())

		cr2, _ := must.Values(opts.ChunkReader(cr1))
		must.Nil(cr2.SkipN(4))
		assert.Equal(t, int64(24), cr0.Count())
		assert.Equal(t, int64(16), cr1.Count())
		assert.Equal(t, int64(8), cr2.Count())

		_, err := io.CopyN(io.Discard, cr2, 8)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(32), cr0.Count())
		assert.Equal(t, int64(24), cr1.Count())
		assert.Equal(t, int64(16), cr2.Count())

		err = cr2.SkipN(3)
		assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
	})
}
