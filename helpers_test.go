package riff

import (
	"bytes"
	"io"
	"testing"

	"github.com/rzajac/flexbuf"
	kit "github.com/rzajac/testkit"
	"github.com/stretchr/testify/assert"
)

func Test_StrToID(t *testing.T) {
	tt := []struct {
		id  string
		exp uint32
	}{
		{"RIFF", IDRIFF},
		{"fmt ", IDfmt},
		{"data", IDdata},
		{"fact", 0x66616374},
		{"WAVE", TypeWAVE},
		{"LIST", IDLIST},
		{"INFO", IDINFO},
		{"id3 ", 0x69643320},
		{"IARL", 0x4941524c},
		{"IART", 0x49415254},
		{"ICMS", 0x49434d53},
		{"ICMT", 0x49434d54},
		{"ICOP", 0x49434f50},
		{"ICRD", 0x49435244},
		{"IENG", 0x49454e47},
		{"IGNR", 0x49474e52},
		{"IKEY", 0x494b4559},
		{"IMED", 0x494d4544},
		{"INAM", 0x494e414d},
		{"IPRD", 0x49505244},
		{"ITRK", 0x4954524b},
		{"ISBJ", 0x4953424a},
		{"ISFT", 0x49534654},
		{"ISRC", 0x49535243},
		{"ISRF", 0x49535246},
		{"ITCH", 0x49544348},
		{"tlst", 0x746c7374},
		{"JUNK", IDJUNK},
		{"bext", 0x62657874},
		{"ABCD", IDUNKN},
		{"AVI ", TypeAVI},
		{"RMID", TypeRMID},

		{"ITCHxx", 0x49544348},
		{"I", 0x49202020},
	}

	for _, tc := range tt {
		t.Run(tc.id, func(t *testing.T) {
			// --- When ---
			got := StrToID(tc.id)

			// --- Then ---
			assert.Exactly(t, tc.exp, got, "test %s", tc.id)
		})
	}
}

func Test_ReadChunkID(t *testing.T) {
	// --- Given ---
	src := bytes.NewReader([]byte{0x41, 0x42, 0x43, 0x44})

	// --- When ---
	var id uint32
	err := ReadChunkID(src, &id)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, uint32(0x41424344), id)
}

func Test_ReadChunkID_Error(t *testing.T) {
	// --- Given ---
	buf := []byte{0x41, 0x42, 0x43, 0x44}
	src := kit.ErrReader(bytes.NewReader(buf), 3, nil)

	// --- When ---
	var id uint32
	err := ReadChunkID(src, &id)

	// --- Then ---
	assert.ErrorIs(t, err, kit.ErrTestError)
	assert.Exactly(t, uint32(0), id)
}

func Test_ReadChunkSize(t *testing.T) {
	// --- Given ---
	src := bytes.NewReader([]byte{0x10, 0x00, 0x00, 0x00})

	// --- When ---
	size, err := ReadChunkSize(src)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, uint32(0x10), size)
}

func Test_ReadChunkSize_Error(t *testing.T) {
	// --- Given ---
	buf := []byte{0x10, 0x00, 0x00}
	src := kit.ErrReader(bytes.NewReader(buf), 3, nil)

	// --- When ---
	size, err := ReadChunkSize(src)

	// --- Then ---
	assert.ErrorIs(t, err, kit.ErrTestError)
	assert.Exactly(t, uint32(0), size)
}

func Test_LimitedRead(t *testing.T) {
	// --- Given ---
	src := bytes.NewReader([]byte{0, 1, 2, 3})
	dst := &flexbuf.Buffer{}

	// --- When ---
	err := LimitedRead(src, 3, dst)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, []byte{0, 1, 2}, kit.ReadAllFromStart(t, dst))
}

func Test_LimitedRead_ErrUnexpectedEOF(t *testing.T) {
	// --- Given ---
	src := bytes.NewReader([]byte{0, 1, 2, 3})
	dst := &flexbuf.Buffer{}

	// --- When ---
	err := LimitedRead(src, 30, dst)

	// --- Then ---
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func Test_LimitedRead_Error(t *testing.T) {
	// --- Given ---
	src := kit.ErrReader(bytes.NewReader([]byte{0, 1, 2}), 2, nil)
	dst := &flexbuf.Buffer{}

	// --- When ---
	err := LimitedRead(src, 3, dst)

	// --- Then ---
	assert.ErrorIs(t, err, kit.ErrTestError)
}

func Test_SkipN_BlackHole(t *testing.T) {
	// --- Given ---
	buf := &bytes.Buffer{}
	for i := 0; i < 30; i++ {
		buf.WriteByte(byte(i))
	}

	// --- When ---
	err := SkipN(buf, 20)

	// --- Then ---
	assert.NoError(t, err)
	exp := []byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29}
	assert.Exactly(t, exp, kit.ReadAll(t, buf))
}

func Test_SkipN_BlackHole_Error(t *testing.T) {
	// --- Given ---
	buf := &bytes.Buffer{}
	for i := 0; i < 30; i++ {
		buf.WriteByte(byte(i))
	}
	r := kit.ErrReader(buf, 10, nil)

	// --- When ---
	err := SkipN(r, 20)

	// --- Then ---
	assert.ErrorIs(t, err, kit.ErrTestError)
	exp := []byte{
		0x0a, 0x0b, 0x0c, 0x0d, 0x0e,
		0x0f, 0x10, 0x11, 0x12, 0x13,
		0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d,
	}
	assert.Exactly(t, exp, kit.ReadAll(t, buf))
}

func Test_SkipN_Seek(t *testing.T) {
	// --- Given ---
	buf := &flexbuf.Buffer{}
	for i := 0; i < 30; i++ {
		_ = buf.WriteByte(byte(i))
	}
	buf.SeekStart()

	// --- When ---
	err := SkipN(buf, 20)

	// --- Then ---
	assert.NoError(t, err)
	exp := []byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29}
	assert.Exactly(t, exp, kit.ReadAll(t, buf))
}

func Test_RealSize(t *testing.T) {
	assert.Exactly(t, uint32(124), RealSize(123))
	assert.Exactly(t, uint32(124), RealSize(124))
}

func Test_ReadPaddingIf_OddSize(t *testing.T) {
	// --- Given ---
	src := bytes.NewReader([]byte{0, 1})

	// --- When ---
	n, err := ReadPaddingIf(src, 3)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(1), n)
}

func Test_ReadPaddingIf_EvenSize(t *testing.T) {
	// --- Given ---
	src := bytes.NewReader([]byte{0, 1})

	// --- When ---
	n, err := ReadPaddingIf(src, 4)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(0), n)
}

func Test_ReadPaddingIf_ErrUnexpectedEOF(t *testing.T) {
	// --- Given ---
	src := bytes.NewReader([]byte{})

	// --- When ---
	n, err := ReadPaddingIf(src, 3)

	// --- Then ---
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
	assert.Exactly(t, int64(0), n)
}

func Test_ReadPaddingIf_HandleEOF(t *testing.T) {
	// --- Given ---
	src := bytes.NewReader([]byte{0})

	// --- When ---
	n, err := ReadPaddingIf(src, 3)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(1), n)
}

func Test_WriteIDAndSize(t *testing.T) {
	// --- Given ---
	dst := &bytes.Buffer{}

	// --- When ---
	n, err := WriteIDAndSize(dst, IDRIFF, 124)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(8), n)

	exp := []byte{0x52, 0x49, 0x46, 0x46, 0x7c, 0x0, 0x0, 0x0}
	assert.Exactly(t, exp, dst.Bytes())
}

func Test_WriteIDAndSize_ErrorWritingID(t *testing.T) {
	// --- Given ---
	buf := &bytes.Buffer{}
	dst := kit.ErrWriter(buf, 3, nil)

	// --- When ---
	n, err := WriteIDAndSize(dst, IDRIFF, 124)

	// --- Then ---
	assert.ErrorIs(t, err, kit.ErrTestError)
	assert.Exactly(t, int64(0), n)

	exp := []byte{0x52, 0x49, 0x46}
	assert.Exactly(t, exp, buf.Bytes())
}

func Test_WriteIDAndSize_ErrorWritingSize(t *testing.T) {
	// --- Given ---
	buf := &bytes.Buffer{}
	dst := kit.ErrWriter(buf, 5, nil)

	// --- When ---
	n, err := WriteIDAndSize(dst, IDRIFF, 124)

	// --- Then ---
	assert.ErrorIs(t, err, kit.ErrTestError)
	assert.Exactly(t, int64(4), n)

	exp := []byte{0x52, 0x49, 0x46, 0x46, 0x7c}
	assert.Exactly(t, exp, buf.Bytes())
}

func Test_WritePaddingIf_OddSize(t *testing.T) {
	// --- Given ---
	dst := &bytes.Buffer{}

	// --- When ---
	n, err := WritePaddingIf(dst, 123)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(1), n)

	exp := []byte{0x00}
	assert.Exactly(t, exp, dst.Bytes())
}

func Test_WritePaddingIf_EvenSize(t *testing.T) {
	// --- Given ---
	dst := &bytes.Buffer{}

	// --- When ---
	n, err := WritePaddingIf(dst, 124)

	// --- Then ---
	assert.NoError(t, err)
	assert.Exactly(t, int64(0), n)
	assert.Exactly(t, []byte(nil), dst.Bytes())
}

func Test_WritePaddingIf_Error(t *testing.T) {
	// --- Given ---
	buf := &bytes.Buffer{}
	dst := kit.ErrWriter(buf, 0, nil)

	// --- When ---
	n, err := WritePaddingIf(dst, 123)

	// --- Then ---
	assert.ErrorIs(t, err, kit.ErrTestError)
	assert.Exactly(t, int64(0), n)
	assert.Exactly(t, []byte(nil), buf.Bytes())
}

func Test_TrimZeroRight(t *testing.T) {
	tt := []struct {
		testN string

		in  []byte
		exp []byte
	}{
		{"1", []byte{}, []byte{}},
		{"2", []byte{0}, []byte{}},
		{"3", []byte{0, 0}, []byte{}},
		{"4", []byte{'a', 'b'}, []byte{'a', 'b'}},
		{"5", []byte{'a', 'b', 0}, []byte{'a', 'b'}},
		{"6", []byte{'a', 'b', 0, 0}, []byte{'a', 'b'}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			assert.Exactly(t, tc.exp, TrimZeroRight(tc.in), "test %s", tc.testN)
		})
	}
}

func Test_linkids(t *testing.T) {
	// --- When ---
	got := linkids(idRAWC, IDRIFF)

	// --- Then ---
	assert.Exactly(t, "RAWC:RIFF", got)
}
