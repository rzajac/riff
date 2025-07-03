package riff

import (
	"os"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
)

func Test_Chunks_First(t *testing.T) {
	// --- Given ---
	rif := New(SkipData)
	_, err := rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
	assert.NoError(t, err)
	chs := rif.Chunks()

	// --- When ---
	ch := chs.First(IDfmt)

	// --- Then ---
	assert.Equal(t, uint32(IDfmt), ch.ID())
	assert.Equal(t, uint32(0x10), ch.Size())
}

func Test_Chunks_First_NotPresent(t *testing.T) {
	// --- Given ---
	rif := New(SkipData)
	_, err := rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
	assert.NoError(t, err)
	chs := rif.Chunks()

	// --- When ---
	ch := chs.First(IDUNKN)

	// --- Then ---
	assert.Nil(t, ch)
}

func Test_Chunks_Size(t *testing.T) {
	// --- Given ---
	rif := New(SkipData)
	_, err := rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
	assert.NoError(t, err)
	chs := rif.Chunks()

	// --- When ---
	size := chs.Size()

	// --- Then ---
	// The 27074 is the file size, we add 12 for ID, type and size fields.
	assert.Equal(t, uint32(27074), size+12)
}

func Test_Chunks_IDs(t *testing.T) {
	// --- Given ---
	rif := New(SkipData)
	_, err := rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
	assert.NoError(t, err)
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
	assert.Equal(t, exp, ids)
}

func Test_Chunks_Count(t *testing.T) {
	// --- Given ---
	rif := New(SkipData)
	_, err := rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
	assert.NoError(t, err)
	chs := rif.Chunks()

	// --- When ---
	cnt := chs.Count(IDJUNK)

	// --- Then ---
	assert.Equal(t, 7, cnt)
}

func Test_Chunks_Remove(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		rif := New(SkipData)
		_, err := rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
		assert.NoError(t, err)

		el := len(rif.Chunks().IDs())

		// --- When ---
		chs := rif.Chunks().Remove(IDfmt).IDs()

		// --- Then ---
		assert.Equal(t, el-1, len(chs))
		assert.Equal(t, 14, el)
	})

	t.Run("right order", func(t *testing.T) {
		// --- Given ---
		rif := New(SkipData)
		_, _ = rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
		chsIDs := rif.Chunks().IDs()

		// --- When ---
		chs := rif.Chunks().Remove(IDfmt).IDs()

		// --- Then ---
		for i := 1; i < len(chs); i++ {
			assert.Equal(t, chsIDs[i+1], chs[i])
		}
		assert.Equal(t, len(chsIDs)-1, len(chs))
		assert.Equal(t, 14, len(chsIDs))
	})

	t.Run("key does not exist", func(t *testing.T) {
		// --- Given ---
		rif := New(SkipData)
		_, _ = rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
		chsIDs := rif.Chunks().IDs()
		id := uint32(1)

		// --- When ---
		chs := rif.Chunks().Remove(id).IDs()

		// --- Then ---
		assert.Equal(t, len(chsIDs), len(chs))
	})
}
