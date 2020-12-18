package riff

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
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
			assert.Exactly(t, tc.idS, got, "test %s", tc.idS)
		})
	}
}

func Test_Uint32_Read(t *testing.T) {
	// --- Given ---
	dst := make([]byte, 4)

	// --- When ---
	n, err := Uint32(unkID).Read(dst)

	// --- Then ---
	assert.ErrorIs(t, err, io.EOF)
	assert.Exactly(t, 4, n)
	assert.Exactly(t, []byte{0x41, 0x42, 0x43, 0x44}, dst)
}

func Test_Uint32_Read_Error(t *testing.T) {
	// --- Given ---
	dst := make([]byte, 3)

	// --- When ---
	n, err := Uint32(unkID).Read(dst)

	// --- Then ---
	assert.EqualError(t, err, "buffer too small for uint32: length too short")
	assert.ErrorIs(t, err, ErrTooShort)
	assert.Exactly(t, 0, n)
}
