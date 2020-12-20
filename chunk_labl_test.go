package riff

import (
	"bytes"
	"io"
	"testing"

	kit "github.com/rzajac/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rzajac/riff/internal/test"
)

func lablChunkTextLenEven(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDlabl))             // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 8)                     // ( 4) 4 - Chunk size
	test.WriteUint32LE(t, src, 123)                   // ( 8) 4 - Cue ID
	test.WriteBytes(t, src, []byte{'a', 'b', 'c', 0}) // (12) 4 - Text
	// Total length: 8+4+4=16
	return src
}

func lablChunkTextLenOdd(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDlabl))        // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 7)                // ( 4) 4 - Chunk size
	test.WriteUint32LE(t, src, 123)              // ( 8) 4 - Cue ID
	test.WriteBytes(t, src, []byte{'a', 'b', 0}) // (12) 3 - Text
	test.WriteByte(t, src, 0)                    // (15) 1 - Padding byte
	// Total length: 8+4+3+1=16
	return src
}

func Test_ChunkLABL_LABL(t *testing.T) {
	// --- When ---
	ch := LABL()

	// --- Then ---
	assert.Exactly(t, IDlabl, ch.ID())
	assert.Exactly(t, uint32(0), ch.Size())
	assert.Exactly(t, uint32(0), ch.Type())
	assert.True(t, ch.Multi())
	assert.Nil(t, ch.Chunks())
	assert.False(t, ch.Raw())
}

func Test_ChunkLABL_ReadFrom_TextLenEven(t *testing.T) {
	// --- Given ---
	src := lablChunkTextLenEven(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := LABL()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(12), n)
	assert.Exactly(t, IDlabl, ch.ID())
	assert.Exactly(t, uint32(8), ch.Size())
	assert.Exactly(t, uint32(123), ch.CuePointID)
	assert.Exactly(t, []byte{'a', 'b', 'c', 0}, ch.label)
	assert.Exactly(t, []byte{'a', 'b', 'c'}, kit.ReadAll(t, ch.Label()))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkLABL_ReadFrom_TextLenOdd(t *testing.T) {
	// --- Given ---
	src := lablChunkTextLenOdd(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := LABL()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(12), n)
	assert.Exactly(t, IDlabl, ch.ID())
	assert.Exactly(t, uint32(7), ch.Size())
	assert.Exactly(t, uint32(123), ch.CuePointID)
	assert.Exactly(t, []byte{'a', 'b', 0}, ch.label)
	assert.Exactly(t, []byte{'a', 'b'}, kit.ReadAll(t, ch.Label()))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkLABL_ReadFrom_Errors(t *testing.T) {
	// Reading less then 12 bytes should always result in an error.
	for i := 1; i < 12; i++ {
		// --- Given ---
		src := lablChunkTextLenEven(t)
		test.Skip4B(t, src) // Skip chunk ID.

		// --- When ---
		_, err := LABL().ReadFrom(io.LimitReader(src, int64(i)))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
}

func Test_ChunkLABL_WriteTo(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(*testing.T) io.Reader
	}{
		{"lablChunkTextLenEven", 16, lablChunkTextLenEven},
		{"lablChunkTextLenOdd", 16, lablChunkTextLenOdd},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			test.Skip4B(t, src) // Skip chunk ID.

			ch := LABL()
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

func Test_ChunkLABL_WriteTo_Errors(t *testing.T) {
	// Writing less then 16 bytes should always result in an error.
	for i := 16; i > 0; i-- {
		// --- Given ---
		src := lablChunkTextLenOdd(t)
		test.Skip4B(t, src) // Skip chunk ID.

		ch := LABL()
		_, err := ch.ReadFrom(src)
		assert.NoError(t, err, "i=%d", i)

		// --- When ---
		dst := &bytes.Buffer{}
		_, err = ch.WriteTo(kit.ErrWriter(dst, i, nil))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
}

func Test_ChunkLABL_Reset(t *testing.T) {
	// --- Given ---
	ch := LABL()
	ch.size = 3
	ch.CuePointID = 123
	ch.label = []byte{'a', 'b', 0}

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Exactly(t, uint32(IDlabl), ch.ID())
	assert.Exactly(t, uint32(0), ch.Size())
	assert.Exactly(t, uint32(0), ch.CuePointID)
	assert.Exactly(t, []byte{}, ch.label)
	assert.Exactly(t, []byte{}, kit.ReadAll(t, ch.Label()))
}
