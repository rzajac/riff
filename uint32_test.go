package riff

import (
	"io"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Uint32_String(t *testing.T) {
	tt := []struct {
		idS string
		idI uint32
	}{
		{"RIFF", 0x52494646},
		{"fmt ", IDfmt},
		{"data", IDdata},
		{"WAVE", 0x57415645},
		{"LIST", IDLIST},
	}

	for _, tc := range tt {
		t.Run(tc.idS, func(t *testing.T) {
			// --- When ---
			got := Uint32(tc.idI).String()

			// --- Then ---
			assert.Equal(t, tc.idS, got)
		})
	}
}

func Test_Uint32_Read(t *testing.T) {
	// --- Given ---
	dst := make([]byte, 4)

	// --- When ---
	n, err := Uint32(IDUNKN).Read(dst)

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte{0x41, 0x42, 0x43, 0x44}, dst)
}

func Test_Uint32_Read_Error(t *testing.T) {
	// --- Given ---
	dst := make([]byte, 3)

	// --- When ---
	n, err := Uint32(IDUNKN).Read(dst)

	// --- Then ---
	assert.ErrorEqual(t, "buffer too small for uint32: length too short", err)
	assert.ErrorIs(t, ErrTooShort, err)
	assert.Equal(t, 0, n)
}
