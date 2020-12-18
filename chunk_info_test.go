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

func infoChunkTextLenEven(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(LabIART))            // (0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 4)                     // (4) 4 - Chunk size
	test.WriteBytes(t, src, []byte{'a', 'b', 'c', 0}) // (8) 4 - Text
	// Total length: 8+4=12
	return src
}

func infoChunkTextLenOdd(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(LabIART))       // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 3)                // ( 4) 4 - Chunk size
	test.WriteBytes(t, src, []byte{'a', 'b', 0}) // ( 8) 4 - Text
	test.WriteByte(t, src, 0)                    // (11) 1 - Padding byte
	// Total length: 8+3+1=12
	return src
}

func infoChunkThree(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(LabIART))            // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 4)                     // ( 4) 4 - Chunk size
	test.WriteBytes(t, src, []byte{'a', 'r', 't', 0}) // ( 8) 4 - Text
	test.ReadFrom(t, src, Uint32(LabICOP))            // (12) 4 - Chunk ID
	test.WriteUint32LE(t, src, 3)                     // (16) 4 - Chunk size
	test.WriteBytes(t, src, []byte{'c', 'p', 0})      // (20) 3 - Text
	test.WriteByte(t, src, 0)                         // (23) 1 - Padding byte
	test.ReadFrom(t, src, Uint32(LabICMT))            // (24) 4 - Chunk ID
	test.WriteUint32LE(t, src, 4)                     // (28) 4 - Chunk size
	test.WriteBytes(t, src, []byte{'c', 'm', 't', 0}) // (32) 4 - Text
	// Total length: 8+4+8+3+1+8+4=36
	return src
}

func Test_ChunkINFO_ReadFrom_TextLenEven(t *testing.T) {
	// --- Given ---
	src := infoChunkTextLenEven(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := INFO(LabIART)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(8), n)
	assert.Exactly(t, LabIART, ch.ID())
	assert.Exactly(t, uint32(4), ch.Size())
	assert.Exactly(t, []byte{'a', 'b', 'c', 0}, ch.text)
	assert.Exactly(t, []byte{'a', 'b', 'c'}, kit.ReadAll(t, ch.Text()))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkINFO_ReadFrom_TextLenOdd(t *testing.T) {
	// --- Given ---
	src := infoChunkTextLenOdd(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := INFO(LabIART)
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(8), n)
	assert.Exactly(t, LabIART, ch.ID())
	assert.Exactly(t, uint32(3), ch.Size())
	assert.Exactly(t, []byte{'a', 'b', 0}, ch.text)
	assert.Exactly(t, []byte{'a', 'b'}, kit.ReadAll(t, ch.Text()))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkINFI_ReadFrom_Errors(t *testing.T) {
	// Reading less then 8 bytes should always result in an error.
	for i := 1; i < 8; i++ {
		// --- Given ---
		src := infoChunkTextLenEven(t)
		test.Skip4B(t, src) // Skip chunk ID.

		// --- When ---
		_, err := INFO(LabIART).ReadFrom(io.LimitReader(src, int64(i)))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
}

func Test_ChunkINFO_WriteTo(t *testing.T) {
	tt := []struct {
		testN string

		id uint32
		n  int64
		ch func(*testing.T) io.Reader
	}{
		{"infoChunkTextLenEven", LabIART, 12, infoChunkTextLenEven},
		{"infoChunkTextLenOdd", LabIART, 12, infoChunkTextLenOdd},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			test.Skip4B(t, src) // Skip chunk ID.

			ch := INFO(tc.id)
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

func Test_ChunkINFO_WriteTo_Errors(t *testing.T) {
	// Writing less then 12 bytes should always result in an error.
	for i := 12; i > 0; i-- {
		// --- Given ---
		src := infoChunkTextLenEven(t)
		test.Skip4B(t, src) // Skip chunk ID.

		ch := INFO(LabIART)
		_, err := ch.ReadFrom(src)
		assert.NoError(t, err, "i=%d", i)

		// --- When ---
		dst := &bytes.Buffer{}
		_, err = ch.WriteTo(kit.ErrWriter(dst, i, nil))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
}

func Test_ChunkINFO_Reset(t *testing.T) {
	// --- Given ---
	ch := INFO(LabIART)
	ch.size = 3
	ch.text = []byte{'a', 'b', 0}

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Exactly(t, uint32(LabIART), ch.ID())
	assert.Exactly(t, uint32(0), ch.Size())
	assert.Exactly(t, []byte{}, ch.text)
	assert.Exactly(t, []byte{}, kit.ReadAll(t, ch.Text()))
}

func Test_ChunkINFO_InfoLabel(t *testing.T) {
	tt := []struct {
		label uint32
		str   string
	}{
		{LabIARL, "archival location"},
		{LabIART, "artist"},
		{LabICMS, "commissioned"},
		{LabICMT, "comments"},
		{LabICOP, "copyright"},
		{LabICRD, "creation date"},
		{LabIENG, "engineer"},
		{LabIGNR, "genre"},
		{LabIKEY, "keywords"},
		{LabIMED, "original medium"},
		{LabINAM, "title"},
		{LabIPRD, "album"},
		{LabITRK, "track"},
		{LabISBJ, "subject"},
		{LabISFT, "software"},
		{LabISRC, "source"},
		{LabISRF, "source form"},
		{LabITCH, "technician"},
		{IDINFO, "INFO"},
	}

	for _, tc := range tt {
		t.Run(tc.str, func(t *testing.T) {
			t.Parallel()
			got := InfoLabel(tc.label)
			assert.Exactly(t, tc.str, got, "test %s", tc.str)
		})
	}
}
