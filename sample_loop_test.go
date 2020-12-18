package riff

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Exactly(t, uint32(0), sl.CuePointID)
	assert.Exactly(t, uint32(0), sl.Type)
	assert.Exactly(t, uint32(0), sl.Start)
	assert.Exactly(t, uint32(0), sl.End)
	assert.Exactly(t, uint32(0), sl.Fraction)
	assert.Exactly(t, uint32(0), sl.PlayCnt)
}
