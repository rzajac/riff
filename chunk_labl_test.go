package riff

import (
	"bytes"
	"io"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	kit "github.com/rzajac/testkit"

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
	assert.Equal(t, IDlabl, ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Equal(t, uint32(0), ch.Type())
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

	assert.Equal(t, int64(12), n)
	assert.Equal(t, IDlabl, ch.ID())
	assert.Equal(t, uint32(8), ch.Size())
	assert.Equal(t, uint32(123), ch.CuePointID)
	assert.Equal(t, []byte{'a', 'b', 'c', 0}, ch.label)
	assert.Equal(t, []byte{'a', 'b', 'c'}, must.Value(io.ReadAll(ch.Label())))
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

	assert.Equal(t, int64(12), n)
	assert.Equal(t, IDlabl, ch.ID())
	assert.Equal(t, uint32(7), ch.Size())
	assert.Equal(t, uint32(123), ch.CuePointID)
	assert.Equal(t, []byte{'a', 'b', 0}, ch.label)
	assert.Equal(t, []byte{'a', 'b'}, must.Value(io.ReadAll(ch.Label())))
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
		if !assert.Error(t, err) {
			t.Logf("errro i=%d", i)
		}
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
			assert.NoError(t, err)

			// --- When ---
			dst := &bytes.Buffer{}
			n, err := ch.WriteTo(dst)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.n, n)

			exp := must.Value(io.ReadAll(tc.ch(t)))
			assert.Equal(t, exp, dst.Bytes())
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
		if !assert.NoError(t, err) {
			t.Logf("errro i=%d", i)
		}

		// --- When ---
		dst := &bytes.Buffer{}
		_, err = ch.WriteTo(kit.ErrWriter(dst, i, nil))

		// --- Then ---
		if !assert.Error(t, err) {
			t.Logf("errro i=%d", i)
		}
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
	assert.Equal(t, uint32(IDlabl), ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Equal(t, uint32(0), ch.CuePointID)
	assert.Equal(t, []byte{}, ch.label)
	assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.Label())))
}
