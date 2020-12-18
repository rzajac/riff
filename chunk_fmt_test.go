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

// fmtChunkWithoutExtraBytes constructs fmt chunk without extra bytes.
func fmtChunkWithoutExtraBytes(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDfmt)) // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 16)       // ( 4) 4 - Chunk size
	test.WriteUint16LE(t, src, CompPCM)  // ( 6) 2 - CompCode
	test.WriteUint16LE(t, src, 1)        // ( 8) 2 - ChannelCnt
	test.WriteUint32LE(t, src, 44100)    // (10) 4 - SampleRate
	test.WriteUint32LE(t, src, 88200)    // (14) 4 - AvgByteRate
	test.WriteUint16LE(t, src, 2)        // (18) 2 - BlockAlign
	test.WriteUint16LE(t, src, 16)       // (20) 2 - BitsPerSample
	// Total length: 8+16=24
	return src
}

// fmtChunkWithExtraBytesEven constructs fmt chunk with extra bytes of
// even length.
func fmtChunkWithExtraBytesEven(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDfmt))  // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 16+4)      // ( 4) 4 - Chunk size
	test.WriteUint16LE(t, src, CompPCM)   // ( 6) 2 - CompCode
	test.WriteUint16LE(t, src, 1)         // ( 8) 2 - ChannelCnt
	test.WriteUint32LE(t, src, 44100)     // (10) 4 - SampleRate
	test.WriteUint32LE(t, src, 88200)     // (14) 4 - AvgByteRate
	test.WriteUint16LE(t, src, 2)         // (18) 2 - BlockAlign
	test.WriteUint16LE(t, src, 16)        // (20) 2 - BitsPerSample
	test.WriteUint16LE(t, src, 2)         // (22) 2 - ExtraBytes
	test.WriteBytes(t, src, []byte{0, 1}) // (24) 2 - Extra bytes
	// Total length: 8+16+2+2=28
	return src
}

// fmtChunkWithExtraBytesOdd constructs fmt chunk with extra bytes of
// odd length.
func fmtChunkWithExtraBytesOdd(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDfmt))     // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 16+2+4)       // ( 4) 4 - Chunk size
	test.WriteUint16LE(t, src, CompPCM)      // ( 6) 2 - CompCode
	test.WriteUint16LE(t, src, 1)            // ( 8) 2 - ChannelCnt
	test.WriteUint32LE(t, src, 44100)        // (10) 4 - SampleRate
	test.WriteUint32LE(t, src, 88200)        // (14) 4 - AvgByteRate
	test.WriteUint16LE(t, src, 2)            // (18) 2 - BlockAlign
	test.WriteUint16LE(t, src, 16)           // (20) 2 - BitsPerSample
	test.WriteUint16LE(t, src, 3)            // (22) 2 - ExtraBytes
	test.WriteBytes(t, src, []byte{0, 1, 2}) // (24) 3 - Extra bytes
	test.WriteByte(t, src, 0)                // (27) 1 - Padding byte
	// Total length: 8+16+2+4=30
	return src
}

// fmtChunkWithExtraOfZeroLen constructs fmt chunk with extra bytes of
// zero length.
func fmtChunkWithExtraOfZeroLen(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDfmt)) // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 16+2)     // ( 4) 4 - Chunk size
	test.WriteUint16LE(t, src, CompPCM)  // ( 6) 2 - CompCode
	test.WriteUint16LE(t, src, 1)        // ( 8) 2 - ChannelCnt
	test.WriteUint32LE(t, src, 44100)    // (10) 4 - SampleRate
	test.WriteUint32LE(t, src, 88200)    // (14) 4 - AvgByteRate
	test.WriteUint16LE(t, src, 2)        // (18) 2 - BlockAlign
	test.WriteUint16LE(t, src, 16)       // (20) 2 - BitsPerSample
	test.WriteUint16LE(t, src, 0)        // (22) 2 - ExtraBytes
	// Total length: 8+16+2+0=26
	return src
}

func Test_ChunkFMT_FMT(t *testing.T) {
	// --- When ---
	ch := FMT()

	// --- Then ---
	assert.Exactly(t, IDfmt, ch.ID())
	assert.Exactly(t, uint32(16), ch.Size())
	assert.Exactly(t, uint32(0), ch.Type())
	assert.False(t, ch.Multi())
	assert.Nil(t, ch.Chunks())
	assert.False(t, ch.Raw())
	assert.False(t, ch.WriteZeroExtra)
}

func Test_ChunkFMT_ReadFrom(t *testing.T) {
	// --- Given ---
	src := fmtChunkWithoutExtraBytes(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := FMT()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(20), n)
	assert.Exactly(t, uint32(16), ch.Size())
	assert.Exactly(t, CompPCM, ch.CompCode)
	assert.Exactly(t, uint16(1), ch.ChannelCnt)
	assert.Exactly(t, uint32(44100), ch.SampleRate)
	assert.Exactly(t, uint32(88200), ch.AvgByteRate)
	assert.Exactly(t, uint16(2), ch.BlockAlign)
	assert.Exactly(t, uint16(16), ch.BitsPerSample)
	assert.Exactly(t, []byte(nil), ch.extra)
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkFMT_ReadFrom_ExtraBytesEven(t *testing.T) {
	// --- Given ---
	src := fmtChunkWithExtraBytesEven(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := FMT()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(24), n)
	assert.Exactly(t, uint32(20), ch.Size())
	assert.Exactly(t, CompPCM, ch.CompCode)
	assert.Exactly(t, uint16(1), ch.ChannelCnt)
	assert.Exactly(t, uint32(44100), ch.SampleRate)
	assert.Exactly(t, uint32(88200), ch.AvgByteRate)
	assert.Exactly(t, uint16(2), ch.BlockAlign)
	assert.Exactly(t, uint16(16), ch.BitsPerSample)
	assert.Exactly(t, []byte{0, 1}, ch.extra)
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkFMT_ReadFrom_ExtraBytesOdd(t *testing.T) {
	// --- Given ---
	src := fmtChunkWithExtraBytesOdd(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := FMT()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(26), n)
	assert.Exactly(t, uint32(22), ch.Size())
	assert.Exactly(t, CompPCM, ch.CompCode)
	assert.Exactly(t, uint16(1), ch.ChannelCnt)
	assert.Exactly(t, uint32(44100), ch.SampleRate)
	assert.Exactly(t, uint32(88200), ch.AvgByteRate)
	assert.Exactly(t, uint16(2), ch.BlockAlign)
	assert.Exactly(t, uint16(16), ch.BitsPerSample)
	assert.Exactly(t, []byte{0, 1, 2}, ch.extra)
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkFMT_ReadFrom_Errors(t *testing.T) {
	// Reading less then 26 bytes should always result in an error.
	for i := 1; i < 26; i++ {
		// --- Given ---
		src := fmtChunkWithExtraBytesOdd(t)
		test.Skip4B(t, src) // Skip chunk ID.

		// --- When ---
		_, err := FMT().ReadFrom(io.LimitReader(src, int64(i)))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
}

func Test_ChunkFMT_ReadFrom_ExtraBytesOfZeroLength(t *testing.T) {
	// --- Given ---
	src := fmtChunkWithExtraOfZeroLen(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := FMT()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(22), n)
	assert.Exactly(t, uint32(18), ch.Size())
	assert.Exactly(t, CompPCM, ch.CompCode)
	assert.Exactly(t, uint16(1), ch.ChannelCnt)
	assert.Exactly(t, uint32(44100), ch.SampleRate)
	assert.Exactly(t, uint32(88200), ch.AvgByteRate)
	assert.Exactly(t, uint16(2), ch.BlockAlign)
	assert.Exactly(t, uint16(16), ch.BitsPerSample)
	assert.Exactly(t, []byte(nil), ch.extra)
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkFMT_ReadFrom_SizeLessThen16Error(t *testing.T) {
	// --- Given ---
	src := &bytes.Buffer{}
	test.WriteUint32LE(t, src, 15)

	// --- When ---
	ch := FMT()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	kit.AssertErrPrefix(t, err, "error decoding fmt  chunk: ")
	assert.Exactly(t, int64(4), n)
}

func Test_ChunkFMT_ReadFrom_LimitReaderError(t *testing.T) {
	// --- Given ---
	src := fmtChunkWithExtraBytesOdd(t)
	test.Skip4B(t, src) // Skip chunk ID.
	src = kit.ErrReader(src, 24, nil)

	// --- When ---
	ch := FMT()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.ErrorIs(t, err, kit.ErrTestError)
	assert.Exactly(t, int64(24), n)
}

func Test_ChunkFMT_Reset(t *testing.T) {
	// --- Given ---
	src := fmtChunkWithExtraBytesOdd(t)
	test.Skip4B(t, src) // Skip chunk ID.

	ch := FMT()
	_, err := ch.ReadFrom(src)
	require.NoError(t, err)

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Exactly(t, uint32(16), ch.Size())
	assert.Exactly(t, CompNone, ch.CompCode)
	assert.Exactly(t, uint16(0), ch.ChannelCnt)
	assert.Exactly(t, uint32(0), ch.SampleRate)
	assert.Exactly(t, uint32(0), ch.AvgByteRate)
	assert.Exactly(t, uint16(0), ch.BlockAlign)
	assert.Exactly(t, uint16(0), ch.BitsPerSample)
	assert.False(t, ch.WriteZeroExtra)
	assert.Len(t, ch.extra, 0)
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkFMT_WriteTo(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(*testing.T) io.Reader
	}{
		{"fmtChunkWithoutExtraBytes", 24, fmtChunkWithoutExtraBytes},
		{"fmtChunkWithExtraBytesEven", 28, fmtChunkWithExtraBytesEven},
		{"fmtChunkWithExtraBytesOdd", 30, fmtChunkWithExtraBytesOdd},
		{"fmtChunkWithExtraOfZeroLen", 26, fmtChunkWithExtraOfZeroLen},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			test.Skip4B(t, src) // Skip chunk ID.

			ch := FMT()
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

func Test_ChunkFMT_WriteTo_Errors(t *testing.T) {
	// Writing less then 30 bytes should always result in an error.
	for i := 30; i > 0; i-- {
		// --- Given ---
		src := fmtChunkWithExtraBytesOdd(t)
		test.Skip4B(t, src) // Skip chunk ID.

		ch := FMT()
		_, err := ch.ReadFrom(src)
		assert.NoError(t, err, "i=%d", i)

		// --- When ---
		dst := &bytes.Buffer{}
		_, err = ch.WriteTo(kit.ErrWriter(dst, i, nil))

		// --- Then ---
		assert.Error(t, err, "i=%d", i)
	}
}

func Test_ChunkFMT_SetExtra(t *testing.T) {
	tt := []struct {
		testN string

		ch      func(*testing.T) io.Reader
		extra   []byte
		expSize uint32
	}{
		{"fmtWithoutExtraBytes1", fmtChunkWithoutExtraBytes, []byte{}, 16},
		{"fmtWithoutExtraBytes2", fmtChunkWithoutExtraBytes, []byte{5, 6}, 20},
		{"fmtWithoutExtraBytes3", fmtChunkWithoutExtraBytes, []byte{5, 6, 7}, 22},

		{"fmtWithExtraBytesEven1", fmtChunkWithExtraBytesEven, []byte{}, 16},
		{"fmtWithExtraBytesEven2", fmtChunkWithExtraBytesEven, []byte{5, 6}, 20},
		{"fmtWithExtraBytesEven3", fmtChunkWithExtraBytesEven, []byte{5, 6, 7}, 22},

		{"fmtWithExtraBytesOdd1", fmtChunkWithExtraBytesOdd, []byte{}, 16},
		{"fmtWithExtraBytesOdd2", fmtChunkWithExtraBytesOdd, []byte{5, 6}, 20},
		{"fmtWithExtraBytesOdd3", fmtChunkWithExtraBytesOdd, []byte{5, 6, 7}, 22},

		{"fmtWithExtraOfZeroLen1", fmtChunkWithExtraOfZeroLen, []byte{}, 18},
		{"fmtWithExtraOfZeroLen2", fmtChunkWithExtraOfZeroLen, []byte{5, 6}, 20},
		{"fmtWithExtraOfZeroLen3", fmtChunkWithExtraOfZeroLen, []byte{5, 6, 7}, 22},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			test.Skip4B(t, src)

			ch := FMT()
			_, err := ch.ReadFrom(src)
			assert.NoError(t, err, tc.testN)

			// --- When ---
			ch.SetExtra(tc.extra)

			// --- Then ---
			assert.Exactly(t, tc.extra, kit.ReadAll(t, ch.Extra()), tc.testN)
			assert.Exactly(t, tc.expSize, ch.Size(), tc.testN)
		})
	}
}

func Test_ChunkFMT_CreateAndWrite(t *testing.T) {
	tt := []struct {
		testN string

		ch    func(*testing.T) io.Reader
		n     int64
		extra []byte
	}{
		{"fmtChunkWithoutExtraBytes", fmtChunkWithoutExtraBytes, 24, []byte{}},
		{"fmtChunkWithExtraBytesEven", fmtChunkWithExtraBytesEven, 28, []byte{0, 1}},
		{"fmtChunkWithExtraBytesOdd", fmtChunkWithExtraBytesOdd, 30, []byte{0, 1, 2}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			ch := FMT()
			ch.CompCode = CompPCM
			ch.ChannelCnt = 1
			ch.SampleRate = 44100
			ch.AvgByteRate = 88200
			ch.BlockAlign = 2
			ch.BitsPerSample = 16

			// --- When ---
			// Setting format extra bytes.
			ch.SetExtra(tc.extra)
			assert.Exactly(t, uint32(tc.n-8), ch.Size(), tc.testN)

			// Writing the chunk.
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

func Test_ChunkFMT_Duration(t *testing.T) {
	// --- Given ---
	src := fmtChunkWithoutExtraBytes(t)
	test.Skip4B(t, src)

	ch := FMT()
	_, err := ch.ReadFrom(src)
	assert.NoError(t, err)

	// --- When ---
	d := ch.Duration(88200)

	// --- Then ---
	assert.Exactly(t, time.Second, d)
}
