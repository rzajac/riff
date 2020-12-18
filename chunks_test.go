package riff

import (
	"testing"

	kit "github.com/rzajac/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Chunks_First(t *testing.T) {
	// --- Given ---
	rif := New(SkipData)
	_, err := rif.ReadFrom(kit.OpenFile(t, "testdata/bwf.wav"))
	require.NoError(t, err)
	chs := rif.Chunks()

	// --- When ---
	ch := chs.First(IDfmt)

	// --- Then ---
	assert.Exactly(t, uint32(IDfmt), ch.ID())
	assert.Exactly(t, uint32(0x10), ch.Size())
}

func Test_Chunks_First_NotPresent(t *testing.T) {
	// --- Given ---
	rif := New(SkipData)
	_, err := rif.ReadFrom(kit.OpenFile(t, "testdata/bwf.wav"))
	require.NoError(t, err)
	chs := rif.Chunks()

	// --- When ---
	ch := chs.First(unkID)

	// --- Then ---
	assert.Nil(t, ch)
}

func Test_Chunks_Size(t *testing.T) {
	// --- Given ---
	rif := New(SkipData)
	_, err := rif.ReadFrom(kit.OpenFile(t, "testdata/bwf.wav"))
	require.NoError(t, err)
	chs := rif.Chunks()

	// --- When ---
	size := chs.Size()

	// --- Then ---
	// The 27074 is the file size, we add 12 for ID, type and size fields.
	assert.Exactly(t, uint32(27074), size+12)
}

func Test_Chunks_IDs(t *testing.T) {
	// --- Given ---
	rif := New(SkipData)
	_, err := rif.ReadFrom(kit.OpenFile(t, "testdata/bwf.wav"))
	require.NoError(t, err)
	chs := rif.Chunks()

	// --- When ---
	ids := chs.IDs()

	// --- Then ---
	exp := []uint32{
		0x62657874, // bext
		IDfmt,      // fmt
		IDdata,     // data
		0x4146416e, // AFAn
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDJUNK,     // JUNK
		IDLIST,     // LIST
		0x41466d64, // AFmd
		IDID3,      // ID3+
	}
	assert.Exactly(t, exp, ids)
}

func Test_Chunks_Count(t *testing.T) {
	// --- Given ---
	rif := New(SkipData)
	_, err := rif.ReadFrom(kit.OpenFile(t, "testdata/bwf.wav"))
	require.NoError(t, err)
	chs := rif.Chunks()

	// --- When ---
	cnt := chs.Count(IDJUNK)

	// --- Then ---
	assert.Exactly(t, 7, cnt)
}
