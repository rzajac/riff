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

func smplWithoutLoopsNoData(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDsmpl)) // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 36)        // ( 4) 4 - Chunk size
	test.WriteUint32LE(t, src, 1)         // ( 8) 4 - Manufacturer
	test.WriteUint32LE(t, src, 2)         // (12) 4 - Product
	test.WriteUint32LE(t, src, 3)         // (16) 4 - SamplePeriod
	test.WriteUint32LE(t, src, 4)         // (20) 4 - MIDIUnityNote
	test.WriteUint32LE(t, src, 5)         // (24) 4 - MIDIPitchFraction
	test.WriteUint32LE(t, src, 6)         // (28) 4 - SMPTEFormat
	test.WriteUint32LE(t, src, 7)         // (32) 4 - SMPTEOffset
	test.WriteUint32LE(t, src, 0)         // (36) 4 - SampleLoopCnt
	test.WriteUint32LE(t, src, 0)         // (40) 4 - SamplerData
	// Total length: 8+36=44
	return src
}

func smplWithoutLoopsWithData(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDsmpl))    // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 39)           // ( 4) 4 - Chunk size
	test.WriteUint32LE(t, src, 1)            // ( 8) 4 - Manufacturer
	test.WriteUint32LE(t, src, 2)            // (12) 4 - Product
	test.WriteUint32LE(t, src, 3)            // (16) 4 - SamplePeriod
	test.WriteUint32LE(t, src, 4)            // (20) 4 - MIDIUnityNote
	test.WriteUint32LE(t, src, 5)            // (24) 4 - MIDIPitchFraction
	test.WriteUint32LE(t, src, 6)            // (28) 4 - SMPTEFormat
	test.WriteUint32LE(t, src, 7)            // (32) 4 - SMPTEOffset
	test.WriteUint32LE(t, src, 0)            // (36) 4 - SampleLoopCnt
	test.WriteUint32LE(t, src, 3)            // (40) 4 - SamplerData
	test.WriteBytes(t, src, []byte{0, 1, 2}) // (44) 3 - Data
	test.WriteByte(t, src, 0)                // (47) 1 - Padding byte
	// Total length: 8+36+3+1=48
	return src
}

func smplWithLoopsNoData(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDsmpl)) // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 60)        // ( 4) 4 - Chunk size
	test.WriteUint32LE(t, src, 1)         // ( 8) 4 - Manufacturer
	test.WriteUint32LE(t, src, 2)         // (12) 4 - Product
	test.WriteUint32LE(t, src, 3)         // (16) 4 - SamplePeriod
	test.WriteUint32LE(t, src, 4)         // (20) 4 - MIDIUnityNote
	test.WriteUint32LE(t, src, 5)         // (24) 4 - MIDIPitchFraction
	test.WriteUint32LE(t, src, 6)         // (28) 4 - SMPTEFormat
	test.WriteUint32LE(t, src, 7)         // (32) 4 - SMPTEOffset
	test.WriteUint32LE(t, src, 1)         // (36) 4 - SampleLoopCnt
	test.WriteUint32LE(t, src, 24)        // (40) 4 - SamplerData
	test.WriteUint32LE(t, src, 9)         // (44) 4 - CuePointID
	test.WriteUint32LE(t, src, 10)        // (48) 4 - Type
	test.WriteUint32LE(t, src, 11)        // (52) 4 - Start
	test.WriteUint32LE(t, src, 12)        // (56) 4 - End
	test.WriteUint32LE(t, src, 13)        // (60) 4 - Fraction
	test.WriteUint32LE(t, src, 14)        // (64) 4 - PlayCnt
	// Total length: 8+36+24=68
	return src
}

func smplWithLoopsWithData(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDsmpl))    // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 63)           // ( 4) 4 - Chunk size
	test.WriteUint32LE(t, src, 1)            // ( 8) 4 - Manufacturer
	test.WriteUint32LE(t, src, 2)            // (12) 4 - Product
	test.WriteUint32LE(t, src, 3)            // (16) 4 - SamplePeriod
	test.WriteUint32LE(t, src, 4)            // (20) 4 - MIDIUnityNote
	test.WriteUint32LE(t, src, 5)            // (24) 4 - MIDIPitchFraction
	test.WriteUint32LE(t, src, 6)            // (28) 4 - SMPTEFormat
	test.WriteUint32LE(t, src, 7)            // (32) 4 - SMPTEOffset
	test.WriteUint32LE(t, src, 1)            // (36) 4 - SampleLoopCnt
	test.WriteUint32LE(t, src, 27)           // (40) 4 - SamplerData
	test.WriteUint32LE(t, src, 9)            // (44) 4 - CuePointID
	test.WriteUint32LE(t, src, 10)           // (48) 4 - Type
	test.WriteUint32LE(t, src, 11)           // (52) 4 - Start
	test.WriteUint32LE(t, src, 12)           // (56) 4 - End
	test.WriteUint32LE(t, src, 13)           // (60) 4 - Fraction
	test.WriteUint32LE(t, src, 14)           // (64) 4 - PlayCnt
	test.WriteBytes(t, src, []byte{0, 1, 2}) // (68) 3 - Data
	test.WriteByte(t, src, 0)                // (71) 1 - Padding byte
	// Total length: 8+36+24+3+1=72
	return src
}

// smplInvalidSize declares one SampleLoop but chunk size declares only 36 bytes.
func smplInvalidSize(t *testing.T) io.Reader {
	src := &bytes.Buffer{}
	test.ReadFrom(t, src, Uint32(IDsmpl)) // ( 0) 4 - Chunk ID
	test.WriteUint32LE(t, src, 36)        // ( 4) 4 - Chunk size
	test.WriteUint32LE(t, src, 1)         // ( 8) 4 - Manufacturer
	test.WriteUint32LE(t, src, 2)         // (12) 4 - Product
	test.WriteUint32LE(t, src, 3)         // (16) 4 - SamplePeriod
	test.WriteUint32LE(t, src, 4)         // (20) 4 - MIDIUnityNote
	test.WriteUint32LE(t, src, 5)         // (24) 4 - MIDIPitchFraction
	test.WriteUint32LE(t, src, 6)         // (28) 4 - SMPTEFormat
	test.WriteUint32LE(t, src, 7)         // (32) 4 - SMPTEOffset
	test.WriteUint32LE(t, src, 1)         // (36) 4 - SampleLoopCnt
	test.WriteUint32LE(t, src, 0)         // (40) 4 - SamplerData
	// Total length: 8+36=44
	return src
}

func Test_ChunkSMPL_SMPL(t *testing.T) {
	// --- When ---
	ch := SMPL()

	// --- Then ---
	assert.Equal(t, IDsmpl, ch.ID())
	assert.Equal(t, uint32(0), ch.Size())
	assert.Equal(t, uint32(0), ch.Type())
	assert.False(t, ch.Multi())
	assert.Nil(t, ch.Chunks())
	assert.False(t, ch.Raw())

	assert.Equal(t, uint32(0), ch.Manufacturer)
	assert.Equal(t, uint32(0), ch.Product)
	assert.Equal(t, uint32(0), ch.SamplePeriod)
	assert.Equal(t, uint32(0), ch.MIDIUnityNote)
	assert.Equal(t, uint32(0), ch.MIDIPitchFraction)
	assert.Equal(t, uint32(0), ch.SMPTEFormat)
	assert.Equal(t, uint32(0), ch.SMPTEOffset)
	assert.Equal(t, uint32(0), ch.SampleLoopCnt)
	assert.Equal(t, uint32(0), ch.SamplerDataCnt)
	assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.SamplerData())))
	assert.Len(t, 0, ch.SampleLoops)
}

func Test_ChunkSMPL_ReadFrom_WithoutLoops(t *testing.T) {
	// --- Given ---
	src := smplWithoutLoopsNoData(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := SMPL()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Equal(t, int64(40), n)
	assert.Equal(t, uint32(36), ch.Size())
	assert.Equal(t, uint32(1), ch.Manufacturer)
	assert.Equal(t, uint32(2), ch.Product)
	assert.Equal(t, uint32(3), ch.SamplePeriod)
	assert.Equal(t, uint32(4), ch.MIDIUnityNote)
	assert.Equal(t, uint32(5), ch.MIDIPitchFraction)
	assert.Equal(t, uint32(6), ch.SMPTEFormat)
	assert.Equal(t, uint32(7), ch.SMPTEOffset)
	assert.Equal(t, uint32(0), ch.SampleLoopCnt)
	assert.Equal(t, uint32(0), ch.SamplerDataCnt)
	assert.Len(t, 0, ch.SampleLoops)
	assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.SamplerData())))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkSMPL_ReadFrom_WithoutLoopsWithData(t *testing.T) {
	// --- Given ---
	src := smplWithoutLoopsWithData(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := SMPL()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Equal(t, int64(44), n)
	assert.Equal(t, uint32(39), ch.Size())
	assert.Equal(t, uint32(1), ch.Manufacturer)
	assert.Equal(t, uint32(2), ch.Product)
	assert.Equal(t, uint32(3), ch.SamplePeriod)
	assert.Equal(t, uint32(4), ch.MIDIUnityNote)
	assert.Equal(t, uint32(5), ch.MIDIPitchFraction)
	assert.Equal(t, uint32(6), ch.SMPTEFormat)
	assert.Equal(t, uint32(7), ch.SMPTEOffset)
	assert.Equal(t, uint32(0), ch.SampleLoopCnt)
	assert.Equal(t, uint32(3), ch.SamplerDataCnt)
	assert.Len(t, 0, ch.SampleLoops)
	assert.Equal(t, []byte{0, 1, 2}, must.Value(io.ReadAll(ch.SamplerData())))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkSMPL_ReadFrom_WithLoopsWithData(t *testing.T) {
	// --- Given ---
	src := smplWithLoopsWithData(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := SMPL()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Equal(t, int64(68), n)
	assert.Equal(t, uint32(63), ch.Size())
	assert.Equal(t, uint32(1), ch.Manufacturer)
	assert.Equal(t, uint32(2), ch.Product)
	assert.Equal(t, uint32(3), ch.SamplePeriod)
	assert.Equal(t, uint32(4), ch.MIDIUnityNote)
	assert.Equal(t, uint32(5), ch.MIDIPitchFraction)
	assert.Equal(t, uint32(6), ch.SMPTEFormat)
	assert.Equal(t, uint32(7), ch.SMPTEOffset)
	assert.Equal(t, uint32(1), ch.SampleLoopCnt)
	assert.Equal(t, uint32(27), ch.SamplerDataCnt)
	assert.Len(t, 1, ch.SampleLoops)

	assert.Equal(t, uint32(9), ch.SampleLoops[0].CuePointID)
	assert.Equal(t, uint32(10), ch.SampleLoops[0].Type)
	assert.Equal(t, uint32(11), ch.SampleLoops[0].Start)
	assert.Equal(t, uint32(12), ch.SampleLoops[0].End)
	assert.Equal(t, uint32(13), ch.SampleLoops[0].Fraction)
	assert.Equal(t, uint32(14), ch.SampleLoops[0].PlayCnt)

	assert.Equal(t, []byte{0, 1, 2}, must.Value(io.ReadAll(ch.SamplerData())))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkSMPL_ReadFrom_WithLoops(t *testing.T) {
	// --- Given ---
	src := smplWithLoopsNoData(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := SMPL()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.NoError(t, err)

	assert.Equal(t, int64(64), n)
	assert.Equal(t, uint32(60), ch.Size())
	assert.Equal(t, uint32(1), ch.Manufacturer)
	assert.Equal(t, uint32(2), ch.Product)
	assert.Equal(t, uint32(3), ch.SamplePeriod)
	assert.Equal(t, uint32(4), ch.MIDIUnityNote)
	assert.Equal(t, uint32(5), ch.MIDIPitchFraction)
	assert.Equal(t, uint32(6), ch.SMPTEFormat)
	assert.Equal(t, uint32(7), ch.SMPTEOffset)
	assert.Equal(t, uint32(1), ch.SampleLoopCnt)
	assert.Equal(t, uint32(24), ch.SamplerDataCnt)
	assert.Len(t, 1, ch.SampleLoops)
	assert.Equal(t, uint32(9), ch.SampleLoops[0].CuePointID)
	assert.Equal(t, uint32(10), ch.SampleLoops[0].Type)
	assert.Equal(t, uint32(11), ch.SampleLoops[0].Start)
	assert.Equal(t, uint32(12), ch.SampleLoops[0].End)
	assert.Equal(t, uint32(13), ch.SampleLoops[0].Fraction)
	assert.Equal(t, uint32(14), ch.SampleLoops[0].PlayCnt)
	assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.SamplerData())))
	assert.True(t, test.IsAllRead(src))
}

func Test_ChunkSMPL_ReadFrom_Errors(t *testing.T) {
	// Reading less than 64 bytes should always result in an error.
	for i := 1; i < 64; i++ {
		// --- Given ---
		src := smplWithLoopsNoData(t)
		test.Skip4B(t, src) // Skip chunk ID.

		// --- When ---
		_, err := SMPL().ReadFrom(io.LimitReader(src, int64(i)))

		// --- Then ---
		assert.Error(t, err)
	}
}

func Test_ChunkSMPL_ReadFrom_TooShortError(t *testing.T) {
	// --- Given ---
	src := &bytes.Buffer{}
	test.WriteUint32LE(t, src, 35)

	// --- When ---
	ch := SMPL()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.ErrorIs(t, ErrTooShort, err)
	assert.ErrorContain(t, "smpl chunk", err)
	assert.Equal(t, int64(4), n)
}

func Test_ChunkSMPL_ReadFrom_InvalidSizeError(t *testing.T) {
	// --- Given ---
	src := smplInvalidSize(t)
	test.Skip4B(t, src) // Skip chunk ID.

	// --- When ---
	ch := SMPL()
	n, err := ch.ReadFrom(src)

	// --- Then ---
	assert.ErrorIs(t, ErrChunkSizeMismatch, err)
	assert.ErrorContain(t, "smpl chunk", err)
	assert.Equal(t, int64(40), n)
}

func Test_ChunkSMPL_Reset(t *testing.T) {
	// --- Given ---
	ch := SMPL()
	ch.size = 1
	ch.Manufacturer = 2
	ch.Product = 3
	ch.SamplePeriod = 4
	ch.MIDIUnityNote = 5
	ch.MIDIPitchFraction = 6
	ch.SMPTEFormat = 7
	ch.SMPTEOffset = 8
	ch.SampleLoopCnt = 9
	ch.SamplerDataCnt = 10
	ch.SampleLoops = []*SampleLoop{
		{
			CuePointID: 11,
			Type:       12,
			Start:      13,
			End:        14,
			Fraction:   15,
			PlayCnt:    16,
		},
	}
	ch.sampleData = []byte{0, 1, 2}

	// --- When ---
	ch.Reset()

	// --- Then ---
	assert.Equal(t, uint32(0), ch.Manufacturer)
	assert.Equal(t, uint32(0), ch.Product)
	assert.Equal(t, uint32(0), ch.SamplePeriod)
	assert.Equal(t, uint32(0), ch.MIDIUnityNote)
	assert.Equal(t, uint32(0), ch.MIDIPitchFraction)
	assert.Equal(t, uint32(0), ch.SMPTEFormat)
	assert.Equal(t, uint32(0), ch.SMPTEOffset)
	assert.Equal(t, uint32(0), ch.SampleLoopCnt)
	assert.Equal(t, uint32(0), ch.SamplerDataCnt)
	assert.Len(t, 0, ch.SampleLoops)
	assert.Equal(t, []byte{}, must.Value(io.ReadAll(ch.SamplerData())))
}

func Test_ChunkSMPL_WriteTo(t *testing.T) {
	tt := []struct {
		testN string

		n  int64
		ch func(*testing.T) io.Reader
	}{
		{"smplWithoutLoopsNoData", 44, smplWithoutLoopsNoData},
		{"smplWithoutLoopsWithData", 48, smplWithoutLoopsWithData},
		{"smplWithLoopsNoData", 68, smplWithLoopsNoData},
		{"smplWithLoopsWithData", 72, smplWithLoopsWithData},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			src := tc.ch(t)
			test.Skip4B(t, src) // Skip chunk ID.

			ch := SMPL()
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

func Test_ChunkSMPL_WriteTo_Errors(t *testing.T) {
	// Writing less then 64 bytes should always result in an error.
	for i := 64; i > 0; i-- {
		// --- Given ---
		src := smplWithLoopsNoData(t)
		test.Skip4B(t, src) // Skip chunk ID.

		ch := SMPL()
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
