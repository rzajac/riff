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

func listChunkType_INFO(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDLIST))             // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 16)                    // ( 4) 4 - Chunk size
	test.WriteUint32BE(t, src, IDINFO)                // ( 8) 4 - Type
	test.ReadFrom(t, src, Uint32(LabIART))            // (12) 4 - Chunk ID
	test.WriteUint32LE(t, src, 4)                     // (16) 4 - Chunk size
	test.WriteBytes(t, src, []byte{'a', 'b', 'c', 0}) // (20) 4 - Text
	// Total length: 24
	return src
}

func listChunkType_adtl(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDLIST))             // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 52)                    // ( 4) 4 - Chunk size
	test.WriteUint32BE(t, src, IDadtl)                // ( 8) 4 - Type
	test.ReadFrom(t, src, Uint32(IDlabl))             // (12) 4 - Chunk ID
	test.WriteUint32LE(t, src, 8)                     // (16) 4 - Chunk size
	test.WriteUint32LE(t, src, 123)                   // (20) 4 - Cue ID
	test.WriteBytes(t, src, []byte{'a', 'b', 'c', 0}) // (24) 4 - Text
	test.ReadFrom(t, src, Uint32(IDltxt))             // (28) 4 - Chunk ID
	test.WriteUint32LE(t, src, 24)                    // (32) 4 - Chunk size
	test.WriteUint32LE(t, src, 1)                     // (36) 4 - PID
	test.WriteUint32LE(t, src, 2)                     // (40) 4 - SamLen
	test.WriteUint32LE(t, src, 3)                     // (44) 4 - PurID
	test.WriteUint16LE(t, src, 4)                     // (48) 2 - Country
	test.WriteUint16LE(t, src, 5)                     // (50) 2 - Language
	test.WriteUint16LE(t, src, 6)                     // (52) 2 - Dialect
	test.WriteUint16LE(t, src, 7)                     // (54) 2 - CodePage
	test.WriteBytes(t, src, []byte{'a', 'b', 'c', 0}) // (56) 4 - Text
	// Total length: 60
	return src
}

func listChunkType_unknown(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDLIST))             // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 16)                    // ( 4) 4 - Chunk size
	test.WriteUint32BE(t, src, IDUNKN)                // ( 8) 4 - Type
	test.ReadFrom(t, src, Uint32(LabIART))            // (12) 4 - Chunk ID
	test.WriteUint32LE(t, src, 4)                     // (16) 4 - Chunk size
	test.WriteBytes(t, src, []byte{'a', 'b', 'c', 0}) // (20) 4 - Text
	// Total length: 24
	return src
}

func Test_ChunkLIST_LIST(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(LoadData))

	// --- When ---
	ch := LIST(LoadData, reg)

	// --- Then ---
	assert.Exactly(t, IDLIST, ch.ID())
	assert.Exactly(t, uint32(0), ch.Size())
	assert.Exactly(t, uint32(0), ch.Type())
	assert.True(t, ch.Multi())
	assert.Nil(t, ch.Chunks())
	assert.False(t, ch.Raw())
}

func Test_ChunkLIST_Type_INFO(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(LoadData))

	src := listChunkType_INFO(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := LIST(LoadData, reg)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(20), n)
	assert.Exactly(t, IDINFO, ch.Type())
	require.Len(t, ch.Chunks(), 1)

	sub := ch.Chunks()[0]
	assert.IsType(t, &ChunkINFO{}, sub)
	assert.Exactly(t, LabIART, sub.ID())
}

func Test_ChunkLIST_Type_adtl(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(LoadData))

	src := listChunkType_adtl(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := LIST(LoadData, reg)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(56), n)
	assert.Exactly(t, IDadtl, ch.Type())
	require.Len(t, ch.Chunks(), 2)

	sub := ch.Chunks()[0]
	assert.IsType(t, &ChunkLABL{}, sub)
	assert.Exactly(t, IDlabl, sub.ID())

	sub = ch.Chunks()[1]
	assert.IsType(t, &ChunkLTXT{}, sub)
	assert.Exactly(t, IDltxt, sub.ID())
}

func Test_ChunkLIST_Type_unknown(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(LoadData))

	src := listChunkType_unknown(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := LIST(LoadData, reg)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(20), n)
	assert.Exactly(t, IDUNKN, ch.Type())
	require.Len(t, ch.Chunks(), 1)

	sub := ch.Chunks()[0]
	assert.IsType(t, &ChunkRAWC{}, sub)
	assert.Exactly(t, LabIART, sub.ID())
}

func Test_ChunkLIST_ReadFrom_Errors(t *testing.T) {
	// Reading less then 20 bytes should always result in an error.
	for i := 1; i < 20; i++ {
		// --- Given ---
		reg := NewRegistry(RAWCMake(LoadData))
		src := listChunkType_INFO(t)
		test.Skip4B(t, src) // Skip chunk ID.

		// --- When ---
		_, err := LIST(LoadData, reg).ReadFrom(io.LimitReader(src, int64(i)))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
}

func Test_ChunkLIST_WriteTo(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(*testing.T) io.Reader
	}{
		{"listChunkType_INFO", 24, listChunkType_INFO},
		{"listChunkType_adtl", 60, listChunkType_adtl},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			reg := NewRegistry(RAWCMake(LoadData))

			src := tc.ch(t)
			test.Skip4B(t, src) // Skip chunk ID.

			ch := LIST(LoadData, reg)
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

func Test_ChunkLIST_WriteTo_Errors(t *testing.T) {
	// Writing less then 60 bytes should always result in an error.
	for i := 60; i > 0; i-- {
		// --- Given ---
		reg := NewRegistry(RAWCMake(LoadData))

		src := listChunkType_adtl(t)
		test.Skip4B(t, src) // Skip chunk ID.

		ch := LIST(LoadData, reg)
		_, err := ch.ReadFrom(src)
		assert.NoError(t, err, "i=%d", i)

		// --- When ---
		dst := &bytes.Buffer{}
		_, err = ch.WriteTo(kit.ErrWriter(dst, i, nil))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
}

func Test_ChunkLIST_Reset(t *testing.T) {
	reg := NewRegistry(RAWCMake(LoadData))

	src := listChunkType_INFO(t)
	test.Skip4B(t, src) // Skip chunk ID.

	ch := LIST(LoadData, reg)
	_, err := ch.ReadFrom(src)
	require.NoError(t, err)

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Exactly(t, uint32(0), ch.Size())
	assert.Exactly(t, uint32(0), ch.ListType)
	assert.Len(t, ch.Chunks(), 0)
}
