package riff

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/kit/iokit"
	"github.com/ctx42/testing/pkg/kit/memfs"

	"github.com/rzajac/riff/internal/test"
)

func Test_ChunkRAWC_RAWC(t *testing.T) {
	// --- When ---
	ch := RAWC(IDUNKN, LoadData)

	// --- Then ---
	assert.Equal(t, uint32(IDUNKN), ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Equal(t, uint32(0), ch.Type())
	assert.True(t, ch.Multi())
	assert.Len(t, 0, ch.Chunks())
	assert.True(t, ch.Raw())
}

func Test_ChunkRAWC_ReadFrom(t *testing.T) {
	// --- Given ---
	src := &bytes.Buffer{}
	test.WriteUint32LE(t, src, 3)    // Chunk size (4).
	src.Write([]byte{'A', 'B', 'C'}) // Chunk data (*).
	src.WriteByte(0)                 // Padding byte (1).

	// --- When ---
	ch := RAWC(IDUNKN, LoadData)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, int64(8), n)
	assert.Equal(t, uint32(3), ch.Size())
	assert.Equal(t, uint32(0), ch.Type())
	assert.Equal(t, []byte{'A', 'B', 'C'}, ch.data)
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkRAWC_ReadFrom_ErrUnexpectedEOF(t *testing.T) {
	// --- Given ---
	src := &bytes.Buffer{}
	src.Write([]byte{0, 1, 2}) // Too short chunk size.

	// --- When ---
	ch := RAWC(IDUNKN, LoadData)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
	assert.ErrorContain(t, "error decoding RAWC:ABCD chunk: ", err)
	assert.Equal(t, int64(0), n)
}

func Test_ChunkRAWC_ReadFrom_ErrorReadingData(t *testing.T) {
	// --- Given ---
	src := &bytes.Buffer{}
	test.WriteUint32LE(t, src, 10)   // Chunk size (4).
	src.Write([]byte{'A', 'B', 'C'}) // Chunk data (*).
	src.WriteByte(0)                 // Padding byte (1).

	// --- When ---
	ch := RAWC(IDUNKN, LoadData)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
	assert.ErrorContain(t, "error decoding RAWC:ABCD chunk: ", err)
	assert.Equal(t, int64(8), n)
}

func Test_ChunkRAWC_ReadFrom_ErrorReadingPadding(t *testing.T) {
	// --- Given ---
	src := &bytes.Buffer{}
	test.WriteUint32LE(t, src, 3)    // Chunk size (4).
	src.Write([]byte{'A', 'B', 'C'}) // Chunk data (*).

	// --- When ---
	ch := RAWC(IDUNKN, LoadData)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.ErrorIs(t, io.ErrUnexpectedEOF, err)
	assert.ErrorContain(t, "error decoding RAWC:ABCD chunk: ", err)
	assert.Equal(t, int64(7), n)
}

func Test_ChunkRAWC_ReadFrom_SkipData_SeekAvailable(t *testing.T) {
	// --- Given ---
	src := &bytes.Buffer{}
	test.WriteUint32LE(t, src, 3)    // Chunk size (4).
	src.Write([]byte{'A', 'B', 'C'}) // Chunk data (*).
	src.WriteByte(0)                 // Padding byte (1).

	// --- When ---
	ch := RAWC(IDUNKN, SkipData)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, int64(8), n)
	assert.Equal(t, uint32(3), ch.Size())
	assert.Nil(t, ch.data)
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkRAWC_ReadFrom_SkipData_SeekNotAvailable(t *testing.T) {
	// --- Given ---
	tmp := &memfs.File{}
	test.WriteUint32LE(t, tmp, 3)                  // Chunk size (4).
	test.WriteBytes(t, tmp, []byte{'A', 'B', 'C'}) // Chunk data (*).
	test.WriteByte(t, tmp, 0)                      // Padding byte (1).
	tmp.SeekStart()                                // Seek to buffer start.

	src := &bytes.Buffer{} // bytes.Buffer doesn't have Seek method.
	test.WriteTo(t, src, tmp)

	// --- When ---
	ch := RAWC(IDUNKN, SkipData)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, int64(8), n)
	assert.Equal(t, uint32(3), ch.Size())
	assert.Nil(t, ch.data)
}

func Test_ChunkRAWC_Write_WithoutPadding(t *testing.T) {
	// --- Given ---
	ch := RAWC(IDUNKN, LoadData)
	ch.size = 2
	ch.data = []byte{0, 1}

	// --- When ---
	dst := &bytes.Buffer{}
	n, err := ch.WriteTo(dst)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, int64(10), n)

	exp := []byte{
		0x41, 0x42, 0x43, 0x44, // ID.
		0x2, 0x0, 0x0, 0x0, // Size.
		0x0, 0x1, // Data.
	}
	assert.Equal(t, exp, dst.Bytes())
}

func Test_ChunkRAWC_Write_WithPadding(t *testing.T) {
	// --- Given ---
	ch := RAWC(IDUNKN, LoadData)
	ch.size = 3
	ch.data = []byte{0, 1, 2}

	// --- When ---
	dst := &bytes.Buffer{}
	n, err := ch.WriteTo(dst)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, int64(12), n)

	exp := []byte{
		0x41, 0x42, 0x43, 0x44, // ID.
		0x3, 0x0, 0x0, 0x0, // Size.
		0x0, 0x1, 0x2, // Data.
		0x0, // Padding.
	}
	assert.Equal(t, exp, dst.Bytes())
}

func Test_ChunkRAWC_Write_ErrorWritingID(t *testing.T) {
	// --- Given ---
	ch := RAWC(IDUNKN, LoadData)
	ch.size = 3
	ch.data = []byte{0, 1, 2}

	buf := &bytes.Buffer{}
	dst := iokit.ErrWriter(buf, 3)

	// --- When ---
	n, err := ch.WriteTo(dst)

	// --- Then ---
	assert.ErrorIs(t, iokit.ErrWrite, err)
	assert.ErrorContain(t, "error encoding RAWC:ABCD chunk: ", err)
	assert.Equal(t, int64(0), n)

	exp := []byte{0x41, 0x42, 0x43}
	assert.Equal(t, exp, buf.Bytes())
}

func Test_ChunkRAWC_Write_ErrorWritingData(t *testing.T) {
	// --- Given ---
	ch := RAWC(IDUNKN, LoadData)
	ch.size = 3
	ch.data = []byte{0, 1, 2}

	buf := &bytes.Buffer{}
	dst := iokit.ErrWriter(buf, 10)

	// --- When ---
	n, err := ch.WriteTo(dst)

	// --- Then ---
	assert.ErrorIs(t, iokit.ErrWrite, err)
	assert.ErrorContain(t, "error encoding RAWC:ABCD chunk: ", err)
	assert.Equal(t, int64(10), n)

	exp := []byte{
		0x41, 0x42, 0x43, 0x44, // ID.
		0x3, 0x0, 0x0, 0x0, // Size.
		0x0, 0x1, // Data.
	}
	assert.Equal(t, exp, buf.Bytes())
}

func Test_ChunkRAWC_Write_ErrorWritingPadding(t *testing.T) {
	// --- Given ---
	ch := RAWC(IDUNKN, LoadData)
	ch.size = 3
	ch.data = []byte{0, 1, 2}

	buf := &bytes.Buffer{}
	dst := iokit.ErrWriter(buf, 12)

	// --- When ---
	n, err := ch.WriteTo(dst)

	// --- Then ---
	assert.ErrorIs(t, iokit.ErrWrite, err)
	assert.ErrorContain(t, "error encoding RAWC:ABCD chunk: ", err)
	assert.Equal(t, int64(12), n)

	exp := []byte{
		0x41, 0x42, 0x43, 0x44, // ID.
		0x3, 0x0, 0x0, 0x0, // Size.
		0x0, 0x1, 0x2, // Data.
		0x0, // Padding
	}
	assert.Equal(t, exp, buf.Bytes())
}

func Test_ChunkRAWC_Write_SkipData(t *testing.T) {
	// --- Given ---
	ch := RAWC(IDUNKN, SkipData)
	ch.size = 3

	dst := &bytes.Buffer{}

	// --- When ---
	n, err := ch.WriteTo(dst)

	// --- Then ---
	assert.True(t, errors.Is(err, ErrSkipDataMode))
	assert.Equal(t, int64(0), n)
}

func Test_ChunkRAWC_Reset(t *testing.T) {
	// --- Given ---
	ch := RAWC(IDUNKN, LoadData)
	ch.size = 3
	ch.data = []byte{0, 1, 2}

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Equal(t, uint32(IDUNKN), ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Len(t, 0, ch.data)
}
