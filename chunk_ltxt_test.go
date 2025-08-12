package riff

import (
	"bytes"
	"io"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/kit/iokit"
	"github.com/ctx42/testing/pkg/must"

	"github.com/rzajac/riff/internal/test"
)

func ltxtChunkTextLenEven(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDltxt))             // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 24)                    // ( 4) 4 - Chunk size
	test.WriteUint32LE(t, src, 1)                     // ( 8) 4 - PID
	test.WriteUint32LE(t, src, 2)                     // (12) 4 - SamLen
	test.WriteUint32LE(t, src, 3)                     // (16) 4 - PurID
	test.WriteUint16LE(t, src, 4)                     // (20) 2 - Country
	test.WriteUint16LE(t, src, 5)                     // (22) 2 - Language
	test.WriteUint16LE(t, src, 6)                     // (24) 2 - Dialect
	test.WriteUint16LE(t, src, 7)                     // (26) 2 - CodePage
	test.WriteBytes(t, src, []byte{'a', 'b', 'c', 0}) // (28) 4 - Text
	// Total length: 8+4*3+2*4+4=32
	return src
}

func ltxtChunkTextLenOdd(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDltxt))        // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 23)               // ( 4) 4 - Chunk size
	test.WriteUint32LE(t, src, 1)                // ( 8) 4 - PID
	test.WriteUint32LE(t, src, 2)                // (12) 4 - SamLen
	test.WriteUint32LE(t, src, 3)                // (16) 4 - PurID
	test.WriteUint16LE(t, src, 4)                // (20) 2 - Country
	test.WriteUint16LE(t, src, 5)                // (22) 2 - Language
	test.WriteUint16LE(t, src, 6)                // (24) 2 - Dialect
	test.WriteUint16LE(t, src, 7)                // (26) 2 - CodePage
	test.WriteBytes(t, src, []byte{'a', 'b', 0}) // (28) 3 - Text
	test.WriteByte(t, src, 0)                    // (31) 1 - Padding byte
	// Total length: 8+4*3+2*4+3+1=32
	return src
}

func Test_ChunkLTXT_LTXT(t *testing.T) {
	// --- When ---
	ch := LTXT()

	// --- Then ---
	assert.Equal(t, IDltxt, ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Equal(t, uint32(0), ch.Type())
	assert.True(t, ch.Multi())
	assert.Nil(t, ch.Chunks())
	assert.False(t, ch.Raw())
}

func Test_ChunkLTXT_ReadFrom_TextLenEven(t *testing.T) {
	// --- Given ---
	src := ltxtChunkTextLenEven(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := LTXT()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Equal(t, int64(28), n)
	assert.Equal(t, IDltxt, ch.ID())
	assert.Equal(t, uint32(24), ch.Size())
	assert.Equal(t, uint32(1), ch.CuePointID)
	assert.Equal(t, uint32(2), ch.SamLen)
	assert.Equal(t, uint32(3), ch.PurID)
	assert.Equal(t, uint16(4), ch.Country)
	assert.Equal(t, uint16(5), ch.Language)
	assert.Equal(t, uint16(6), ch.Dialect)
	assert.Equal(t, uint16(7), ch.CodePage)
	assert.Equal(t, []byte{'a', 'b', 'c', 0}, ch.text)
	assert.Equal(t, []byte{'a', 'b', 'c'}, must.Value(io.ReadAll(ch.Text())))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkLTXT_ReadFrom_TextLenOdd(t *testing.T) {
	// --- Given ---
	src := ltxtChunkTextLenOdd(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := LTXT()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Equal(t, int64(28), n)
	assert.Equal(t, IDltxt, ch.ID())
	assert.Equal(t, uint32(23), ch.Size())
	assert.Equal(t, uint32(1), ch.CuePointID)
	assert.Equal(t, uint32(2), ch.SamLen)
	assert.Equal(t, uint32(3), ch.PurID)
	assert.Equal(t, uint16(4), ch.Country)
	assert.Equal(t, uint16(5), ch.Language)
	assert.Equal(t, uint16(6), ch.Dialect)
	assert.Equal(t, uint16(7), ch.CodePage)
	assert.Equal(t, []byte{'a', 'b', 0}, ch.text)
	assert.Equal(t, []byte{'a', 'b'}, must.Value(io.ReadAll(ch.Text())))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkLTXT_ReadFrom_Errors(t *testing.T) {
	// Reading less than 12 bytes should always result in an error.
	for i := 1; i < 12; i++ {
		// --- Given ---
		src := ltxtChunkTextLenEven(t)
		test.Skip4B(t, src) // Skip chunk ID.

		// --- When ---
		_, err := LTXT().ReadFrom(io.LimitReader(src, int64(i)))

		// --- Then ---
		assert.Error(t, err)
	}
}

func Test_ChunkLTXT_ReadFrom_TooShortError(t *testing.T) {
	// --- Given ---
	src := &bytes.Buffer{}
	test.WriteUint32LE(t, src, 19)

	// --- When ---
	ch := LTXT()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.ErrorIs(t, ErrTooShort, err)
	assert.ErrorContain(t, "INFO:ltxt chunk", err)
	assert.Equal(t, int64(4), n)
}

func Test_ChunkLTXT_WriteTo(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(*testing.T) io.Reader
	}{
		{"ltxtChunkTextLenEven", 32, ltxtChunkTextLenEven},
		{"ltxtChunkTextLenOdd", 32, ltxtChunkTextLenOdd},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			test.Skip4B(t, src) // Skip chunk ID.

			ch := LTXT()
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

func Test_ChunkLTXT_WriteTo_Errors(t *testing.T) {
	// Writing less then 32 bytes should always result in an error.
	for i := 32; i > 0; i-- {
		// --- Given ---
		src := ltxtChunkTextLenOdd(t)
		test.Skip4B(t, src) // Skip chunk ID.

		ch := LTXT()
		_, err := ch.ReadFrom(src)
		if !assert.NoError(t, err) {
			t.Logf("errro i=%d", i)
		}

		// --- When ---
		dst := &bytes.Buffer{}
		_, err = ch.WriteTo(iokit.ErrWriter(dst, i))

		// --- Then ---
		if !assert.Error(t, err) {
			t.Logf("errro i=%d", i)
		}
	}
}

func Test_ChunkLTXT_Reset(t *testing.T) {
	// --- Given ---
	ch := LTXT()
	ch.size = 3
	ch.CuePointID = 1
	ch.SamLen = 2
	ch.PurID = 3
	ch.Country = 4
	ch.Language = 5
	ch.Dialect = 6
	ch.CodePage = 7
	ch.text = []byte{'a', 'b', 0}

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Equal(t, IDltxt, ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Equal(t, uint32(0), ch.CuePointID)
	assert.Equal(t, uint32(0), ch.SamLen)
	assert.Equal(t, uint32(0), ch.PurID)
	assert.Equal(t, uint16(0), ch.Country)
	assert.Equal(t, uint16(0), ch.Language)
	assert.Equal(t, uint16(0), ch.Dialect)
	assert.Equal(t, uint16(0), ch.CodePage)
	assert.Equal(t, []byte{}, ch.text)
}
