package riff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Registry_Register_Get(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(LoadData))

	// --- When ---
	reg.Register(IDfmt, FMTMake)

	// --- Then ---
	assert.IsType(t, &ChunkFMT{}, reg.Get(IDfmt))
	assert.IsType(t, &ChunkRAWC{}, reg.Get(IDUNKN))
}

func Test_Registry_Get_Reuse(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(LoadData))
	reg.Register(IDfmt, FMTMake)

	// --- When ---
	ch0 := reg.Get(IDfmt)
	reg.Put(ch0)
	ch1 := reg.Get(IDfmt)

	// --- Then ---
	assert.Same(t, ch0, ch1)
}

func Test_Registry_Has(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(LoadData))
	reg.Register(IDfmt, FMTMake)

	// --- Then ---
	assert.True(t, reg.Has(IDfmt))
	assert.False(t, reg.Has(IDLIST))
}

func Test_Registry_GetNoRaw(t *testing.T) {
	// --- Given ---
	reg := NewRegistry(RAWCMake(LoadData))

	// --- When ---
	ch0 := reg.GetNoRaw(IDfmt)

	// --- Then ---
	assert.Nil(t, ch0)
}
