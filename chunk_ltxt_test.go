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

func Test_ChunkLTXT_ReadFrom_TextLenEven(t *testing.T) {
	// --- Given ---
	src := ltxtChunkTextLenEven(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := LTXT()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(28), n)
	assert.Exactly(t, IDltxt, ch.ID())
	assert.Exactly(t, uint32(24), ch.Size())
	assert.Exactly(t, uint32(1), ch.CuePointID)
	assert.Exactly(t, uint32(2), ch.SamLen)
	assert.Exactly(t, uint32(3), ch.PurID)
	assert.Exactly(t, uint16(4), ch.Country)
	assert.Exactly(t, uint16(5), ch.Language)
	assert.Exactly(t, uint16(6), ch.Dialect)
	assert.Exactly(t, uint16(7), ch.CodePage)
	assert.Exactly(t, []byte{'a', 'b', 'c', 0}, ch.text)
	assert.Exactly(t, []byte{'a', 'b', 'c'}, kit.ReadAll(t, ch.Text()))
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

	assert.Exactly(t, int64(28), n)
	assert.Exactly(t, IDltxt, ch.ID())
	assert.Exactly(t, uint32(23), ch.Size())
	assert.Exactly(t, uint32(1), ch.CuePointID)
	assert.Exactly(t, uint32(2), ch.SamLen)
	assert.Exactly(t, uint32(3), ch.PurID)
	assert.Exactly(t, uint16(4), ch.Country)
	assert.Exactly(t, uint16(5), ch.Language)
	assert.Exactly(t, uint16(6), ch.Dialect)
	assert.Exactly(t, uint16(7), ch.CodePage)
	assert.Exactly(t, []byte{'a', 'b', 0}, ch.text)
	assert.Exactly(t, []byte{'a', 'b'}, kit.ReadAll(t, ch.Text()))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkLTXT_ReadFrom_Errors(t *testing.T) {
	// Reading less then 12 bytes should always result in an error.
	for i := 1; i < 12; i++ {
		// --- Given ---
		src := ltxtChunkTextLenEven(t)
		test.Skip4B(t, src) // Skip chunk ID.

		// --- When ---
		_, err := LTXT().ReadFrom(io.LimitReader(src, int64(i)))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
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

func Test_ChunkLTXT_WriteTo_Errors(t *testing.T) {
	// Writing less then 32 bytes should always result in an error.
	for i := 32; i > 0; i-- {
		// --- Given ---
		src := ltxtChunkTextLenOdd(t)
		test.Skip4B(t, src) // Skip chunk ID.

		ch := LTXT()
		_, err := ch.ReadFrom(src)
		assert.NoError(t, err, "i=%d", i)

		// --- When ---
		dst := &bytes.Buffer{}
		_, err = ch.WriteTo(kit.ErrWriter(dst, i, nil))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
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
	assert.Exactly(t, IDltxt, ch.ID())
	assert.Exactly(t, uint32(0), ch.Size())
	assert.Exactly(t, uint32(0), ch.CuePointID)
	assert.Exactly(t, uint32(0), ch.SamLen)
	assert.Exactly(t, uint32(0), ch.PurID)
	assert.Exactly(t, uint16(0), ch.Country)
	assert.Exactly(t, uint16(0), ch.Language)
	assert.Exactly(t, uint16(0), ch.Dialect)
	assert.Exactly(t, uint16(0), ch.CodePage)
	assert.Exactly(t, []byte{}, ch.text)
}
