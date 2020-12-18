package riff

import (
	"bytes"
	"io"
	"testing"
	"time"

	kit "github.com/rzajac/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rzajac/riff/internal/test"
)

// dataChunkEven constructs data chunk with even number of data bytes.
func dataChunkEven(t *testing.T) io.Reader {
	data := []byte{
		0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f,
	}

	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDdata)) // (0)  4 - Chunk ID
	test.WriteUint32LE(t, src, 16)        // (4)  4 - Chunk size
	test.WriteBytes(t, src, data)         // (8) 16 - Data
	// Total length: 8+16=24
	return src
}

// dataChunkOdd constructs data chunk with odd number of data bytes.
func dataChunkOdd(t *testing.T) io.Reader {
	data := []byte{
		0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e,
	}

	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDdata)) // ( 0)  4 - Chunk ID
	test.WriteUint32LE(t, src, 15)        // ( 4)  4 - Chunk size
	test.WriteBytes(t, src, data)         // ( 8) 15 - Data
	test.WriteByte(t, src, 0)             // (24)  1 - Padding byte
	// Total length: 8+16=24
	return src
}

func Test_ChunkDATA_DATA_SkipDataMode(t *testing.T) {
	// --- When ---
	ch := DATA(SkipData)

	// --- Then ---
	assert.Exactly(t, IDdata, ch.ID())
	assert.Exactly(t, uint32(0), ch.Size())
	assert.Exactly(t, uint32(0), ch.Type())
	assert.False(t, ch.Multi())
	assert.Nil(t, ch.Chunks())
	assert.False(t, ch.Raw())
	assert.Nil(t, ch.data)
	assert.Exactly(t, []byte{}, kit.ReadAll(t, ch.Data()))
}

func Test_ChunkDATA_DATA_LoadDataMode(t *testing.T) {
	// --- When ---
	ch := DATA(LoadData)

	// --- Then ---
	assert.Exactly(t, IDdata, ch.ID())
	assert.Exactly(t, uint32(0), ch.Size())
	assert.Exactly(t, uint32(0), ch.Type())
	assert.False(t, ch.Multi())
	assert.Nil(t, ch.Chunks())
	assert.False(t, ch.Raw())
	assert.NotNil(t, ch.data)
	assert.Exactly(t, []byte{}, kit.ReadAll(t, ch.Data()))
}

func Test_ChunkDATA_SetData_Even(t *testing.T) {
	// --- Given ---
	ch := DATA(LoadData)

	// --- When ---
	err := ch.SetData([]byte{0, 1, 2, 3})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, uint32(4), ch.Size())
	assert.Exactly(t, []byte{0, 1, 2, 3}, kit.ReadAll(t, ch.Data()))
}

func Test_ChunkDATA_SetData_Odd(t *testing.T) {
	// --- Given ---
	ch := DATA(LoadData)

	// --- When ---
	err := ch.SetData([]byte{0, 1, 2})

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, uint32(3), ch.Size())
	assert.Exactly(t, []byte{0, 1, 2}, kit.ReadAll(t, ch.Data()))
}

func Test_ChunkDATA_SetData_SkipDataMode_Error(t *testing.T) {
	// --- Given ---
	ch := DATA(SkipData)

	// --- When ---
	err := ch.SetData([]byte{0, 1, 2})

	// --- Then ---
	assert.ErrorIs(t, err, ErrSkipDataMode)
	assert.Exactly(t, uint32(0), ch.Size())
	assert.Exactly(t, []byte{}, kit.ReadAll(t, ch.Data()))
}

func Test_ChunkDATA_ReadFrom_Even(t *testing.T) {
	// --- Given ---
	src := dataChunkEven(t)
	test.Skip4B(t, src) // Skip chunk ID.

	ch := DATA(LoadData)

	// --- When ---
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(20), n)
	assert.Exactly(t, uint32(16), ch.Size())

	exp := []byte{
		0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f,
	}
	assert.Exactly(t, exp, kit.ReadAll(t, ch.Data()))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkDATA_ReadFrom_Odd(t *testing.T) {
	// --- Given ---
	src := dataChunkOdd(t)
	test.Skip4B(t, src) // Skip chunk ID.

	ch := DATA(LoadData)

	// --- When ---
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(20), n)
	assert.Exactly(t, uint32(15), ch.Size())

	exp := []byte{
		0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e,
	}
	assert.Exactly(t, exp, kit.ReadAll(t, ch.Data()))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkDATA_ReadFrom_Errors(t *testing.T) {
	// Reading less then 20 bytes should always result in an error.
	for i := 1; i < 20; i++ {
		// --- Given ---
		src := dataChunkOdd(t)
		test.Skip4B(t, src) // Skip chunk ID.

		// --- When ---
		_, err := DATA(LoadData).ReadFrom(io.LimitReader(src, int64(i)))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
}

func TestChunkDATA_WriteTo(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(*testing.T) io.Reader
	}{
		{"dataChunkEven", 24, dataChunkEven},
		{"dataChunkOdd", 24, dataChunkOdd},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			test.Skip4B(t, src) // Skip chunk ID.

			ch := DATA(LoadData)
			_, err := ch.ReadFrom(src)
			require.NoError(t, err, tc.testN)

			// --- When ---
			dst := &bytes.Buffer{}
			n, err := ch.WriteTo(dst)

			// --- Then ---
			assert.NoError(t, err, tc.testN)
			assert.Exactly(t, tc.n, n, tc.testN)

			exp := kit.ReadAll(t, tc.ch(t))
			assert.Exactly(t, exp, dst.Bytes(), tc.testN)
		})
	}
}

func Test_ChunkDATA_WriteTo_Errors(t *testing.T) {
	// Writing less then 24 bytes should always result in an error.
	for i := 24; i > 0; i-- {
		// --- Given ---
		src := dataChunkOdd(t)
		test.Skip4B(t, src) // Skip chunk ID.

		ch := DATA(LoadData)
		_, err := ch.ReadFrom(src)
		assert.NoError(t, err, "i=%d", i)

		// --- When ---
		dst := &bytes.Buffer{}
		_, err = ch.WriteTo(kit.ErrWriter(dst, i, nil))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
}

func Test_ChunkDATA_WriteTo_SkipData(t *testing.T) {
	// --- Given ---
	src := dataChunkOdd(t)
	test.Skip4B(t, src) // Skip chunk ID.

	ch := DATA(SkipData)
	_, err := ch.ReadFrom(src)
	require.NoError(t, err)

	// --- When ---
	n, err := ch.WriteTo(&bytes.Buffer{})

	// --- Then ---
	assert.ErrorIs(t, err, ErrSkipDataMode)
	assert.Exactly(t, int64(0), n)
}

func Test_ChunkDATA_Duration(t *testing.T) {
	// --- Given ---
	ch := DATA(SkipData)
	ch.size = 88200
	ch.data = bytes.Repeat([]byte{0}, 88200)

	// --- When ---
	d := ch.Duration(88200)

	// --- Then ---
	assert.Exactly(t, time.Second, d)
}

func Test_ChunkDATA_Reset(t *testing.T) {
	// --- Given ---
	ch := DATA(LoadData)
	ch.size = 88200
	err := ch.SetData(bytes.Repeat([]byte{0}, 88200))
	assert.NoError(t, err)

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Exactly(t, uint32(0), ch.Size())
	assert.Len(t, ch.data, 0)
}
