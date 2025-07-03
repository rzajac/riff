package riff

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_SampleLoop_Reset(t *testing.T) {
	// --- Given ---
	sl := &SampleLoop{}
	sl.CuePointID = 1
	sl.Type = 2
	sl.Start = 3
	sl.End = 4
	sl.Fraction = 5
	sl.PlayCnt = 6

	// --- When ---
	sl.Reset()

	// --- Then ---
	assert.Equal(t, uint32(0), sl.CuePointID)
	assert.Equal(t, uint32(0), sl.Type)
	assert.Equal(t, uint32(0), sl.Start)
	assert.Equal(t, uint32(0), sl.End)
	assert.Equal(t, uint32(0), sl.Fraction)
	assert.Equal(t, uint32(0), sl.PlayCnt)
}
