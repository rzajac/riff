package riff

import (
	"bytes"
	"testing"

	"github.com/rzajac/flexbuf"
	kit "github.com/rzajac/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RIFF_New(t *testing.T) {
	// --- When ---
	rif := New(LoadData)

	// --- Then ---
	assert.Exactly(t, IDRIFF, rif.ID())
	assert.Exactly(t, uint32(0), rif.Size())
	assert.Exactly(t, uint32(0), rif.Type())
	assert.False(t, rif.Multi())

	assert.True(t, rif.IsRegistered(IDfmt))
	assert.True(t, rif.IsRegistered(IDdata))
	assert.False(t, rif.IsRegistered(0))
}

func Test_RIFF_Bare(t *testing.T) {
	// --- When ---
	reg := NewRegistry(RAWCMake(LoadData))
	rif := Bare(reg)

	// --- Then ---
	assert.Exactly(t, IDRIFF, rif.ID())
	assert.Exactly(t, uint32(0), rif.Size())
	assert.Exactly(t, uint32(0), rif.Type())
	assert.False(t, rif.Multi())

	assert.False(t, rif.IsRegistered(IDfmt))
	assert.False(t, rif.IsRegistered(IDdata))
}

func Test_RIFF_Bare_NilRegistry(t *testing.T) {
	// --- When ---
	rif := Bare(nil)

	// --- Then ---
	assert.Exactly(t, IDRIFF, rif.ID())
	assert.Exactly(t, uint32(0), rif.Size())
	assert.Exactly(t, uint32(0), rif.Type())
	assert.False(t, rif.Multi())

	assert.False(t, rif.IsRegistered(IDfmt))
	assert.False(t, rif.IsRegistered(IDdata))
}

func Test_RIFF_ReadFrom_SmokeTest(t *testing.T) {
	rif := New(SkipData)

	tt := []struct {
		pth    string
		typ    uint32
		size   int64
		chunks int
	}{
		{"testdata/11k16bitpcm.wav", TypeWAVE, 304578, 2},
		{"testdata/11k8bitpcm.wav", TypeWAVE, 152312, 2},
		{"testdata/11kadpcm.wav", TypeWAVE, 77252, 3},
		{"testdata/11kgsm.wav", TypeWAVE, 31000, 3},
		{"testdata/11kulaw.wav", TypeWAVE, 152326, 3},
		{"testdata/32bit.wav", TypeWAVE, 352844, 2},
		{"testdata/8bit.wav", TypeWAVE, 88244, 2},
		{"testdata/8k16bitpcm.wav", TypeWAVE, 221026, 2},
		{"testdata/8k8bitpcm.wav", TypeWAVE, 110532, 2},
		{"testdata/8kadpcm.wav", TypeWAVE, 56072, 3},
		{"testdata/8kcelp.wav", TypeWAVE, 8350, 3},
		{"testdata/8kgsm.wav", TypeWAVE, 22550, 3},
		{"testdata/8kmp316.wav", TypeWAVE, 27386, 3},
		{"testdata/8kmp38.wav", TypeWAVE, 13584, 3},
		{"testdata/8ksbc12.wav", TypeWAVE, 19658, 3},
		{"testdata/8ktruespeech.wav", TypeWAVE, 14842, 3},
		{"testdata/8kulaw.wav", TypeWAVE, 110546, 3},
		{"testdata/bass.wav", TypeWAVE, 143786, 2},
		{"testdata/bwf.wav", TypeWAVE, 27074, 14},
		{"testdata/dirty-kick-24b441k.wav", TypeWAVE, 132136, 3},
		{"testdata/flloop.wav", TypeWAVE, 434838, 7},
		{"testdata/junkKick.wav", TypeWAVE, 83084, 9},
		{"testdata/kick-16b441k.wav", TypeWAVE, 31692, 3},
		{"testdata/kick.wav", TypeWAVE, 9012, 2},
		{"testdata/listChunkInHeader.wav", TypeWAVE, 104196, 4},
		{"testdata/listinfo.wav", TypeWAVE, 176720, 4},
		{"testdata/misaligned-chunk.wav", TypeWAVE, 3441572, 8},
		{"testdata/padded24b.wav", TypeWAVE, 24908, 5},
		{"testdata/sample16bit.wav", TypeWAVE, 882160, 3},
		{"testdata/sample.avi", TypeAVI, 230264, 4}, // --
		{"testdata/sample.rmi", TypeRMID, 29640, 4}, // --
		{"testdata/sample.wav", TypeWAVE, 54002, 2},
	}

	for _, tc := range tt {
		t.Run(tc.pth, func(t *testing.T) {
			// --- Given ---
			fil := kit.OpenFile(t, tc.pth)

			// --- When ---
			n, err := rif.ReadFrom(fil)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.pth)
			assert.Exactly(t, tc.size, n, "test %s", tc.pth)
			assert.Exactly(t, uint32(tc.size-8), rif.Size(), "test %s", tc.pth)
			assert.Exactly(t, tc.typ, rif.Type(), "test %s", tc.pth)
			assert.Len(t, rif.Chunks(), tc.chunks, "test %s", tc.pth)
		})
	}
}

func Test_RIFF_CorrectingSize(t *testing.T) {
	// --- Given ---
	rif := New(LoadData)
	_, err := rif.ReadFrom(kit.OpenFile(t, "testdata/bwf.wav"))
	require.NoError(t, err)

	// --- When ---
	dst := &bytes.Buffer{}
	n, err := rif.WriteTo(dst)

	// --- Then ---
	assert.NoError(t, err)

	assert.Exactly(t, int64(27074), n)
	assert.Exactly(t, uint32(27074), rif.Size()+8)
	assert.Exactly(t, "67cfc04cee2ad37e90e480a26fa45c0e", kit.MD5Reader(t, dst))
}

func Test_RIFF_WriteTo_SmokeTest(t *testing.T) {
	rif := New(LoadData)

	tt := []struct {
		pth  string
		size int64
		hash string
	}{
		{"testdata/11k16bitpcm.wav", 304578, "1fffa675b2467f77c691843e7d096595"},
		{"testdata/11k8bitpcm.wav", 152312, "5bc47c5d45b3c50f6cd90939b9d4d80e"},
		{"testdata/11kadpcm.wav", 77252, "6b90b2661f2c9171a8cbcf23b88ba404"},
		{"testdata/11kgsm.wav", 31000, "e16651ac93f206192e12a75fa5a69d02"},
		{"testdata/11kulaw.wav", 152326, "b7fd8d4ba39edc80112f565aacb2e3cb"},
		{"testdata/32bit.wav", 352844, "14c4936b9e2f28de8489af4ced6d1f05"},
		{"testdata/8bit.wav", 88244, "b23f0c1392ff580281ec3ff2cf66ef21"},
		{"testdata/8k16bitpcm.wav", 221026, "2c313f0691f872d50d71399b78318fe0"},
		{"testdata/8k8bitpcm.wav", 110532, "2616254a680948b9f5cd4cad6cd64af2"},
		{"testdata/8kadpcm.wav", 56072, "e47bf1f4c3f987af80e3e8f576b8bd1b"},
		{"testdata/8kcelp.wav", 8350, "be2acd3a5da4a1c56a4d2fd8a254665b"},
		{"testdata/8kgsm.wav", 22550, "a9ada6656389e38b488e1eb77d7310b4"},
		{"testdata/8kmp316.wav", 27386, "77766ae68f490737c6105c1630dc6c07"},
		{"testdata/8kmp38.wav", 13584, "fc83a0bda8a45f78413eba43d9f07e78"},
		{"testdata/8ksbc12.wav", 19658, "b4228c93ed192aec8531946f215c194d"},
		{"testdata/8ktruespeech.wav", 14842, "c412c81374599b7a165d943067c55ebf"},
		{"testdata/8kulaw.wav", 110546, "b0e107862bb4b8e0a0a3821f27c0bd93"},
		{"testdata/bass.wav", 143786, "db23c035ed961fe32c63806b95cd3b5a"},
		{"testdata/dirty-kick-24b441k.wav", 132136, "987fa9f0c9b328ee9a9fe2801274b5fd"},
		{"testdata/flloop.wav", 434838, "6d553975207ed98e10dc313bd415db86"},
		{"testdata/junkKick.wav", 83084, "076be6fd56e33a658dcc9fcedbab5a41"},
		{"testdata/kick-16b441k.wav", 31692, "f49af03be043e796c699be1da22d6d7d"},
		{"testdata/kick.wav", 9012, "2f69f2d444206336866ffe37699f614a"},
		{"testdata/listChunkInHeader.wav", 104196, "80690b46dea414c349353d41e66b1a7c"},
		{"testdata/listinfo.wav", 176720, "4d145a1b64a8e1d348bb3a5937cbf849"},
		{"testdata/misaligned-chunk.wav", 3441572, "d5382f19e0fb746e30aaf4e9fa33ce53"},
		{"testdata/padded24b.wav", 24908, "227863fe5043172b12171dbd93b03873"},
		{"testdata/sample.avi", 230264, "bc5558ae9465ef6addb308317a99a6df"},
		{"testdata/sample.rmi", 29640, "df2f2f43af049d401bc3ebc73bf79045"},
	}

	for _, tc := range tt {
		t.Run(tc.pth, func(t *testing.T) {
			// --- Given ---
			_, err := rif.ReadFrom(kit.OpenFile(t, tc.pth))
			require.NoError(t, err, "test %s", tc.pth)

			// --- When ---
			buf := &bytes.Buffer{}
			n, err := rif.WriteTo(buf)

			// --- Then ---
			assert.NoError(t, err, "test %s", tc.pth)

			// fil := kit.CreateFile(t, filepath.Join("tmp", filepath.Base(tc.pth)))
			// content := kit.ReadAll(t, buf)
			// fil.Write(content)
			// buf = bytes.NewBuffer(content)

			assert.Exactly(t, tc.size, n, "test %s", tc.pth)
			size := int64(RealSize(rif.Size()) + 8)
			assert.Exactly(t, tc.size, size, "test %s", tc.pth)
			assert.Exactly(t, tc.hash, kit.MD5Reader(t, buf), "test %s", tc.pth)
		})
	}
}

func Benchmark_RIFFReuse(b *testing.B) {
	tt := []struct {
		pth string
	}{
		{"testdata/11k16bitpcm.wav"},
		{"testdata/11k8bitpcm.wav"},
		{"testdata/11kadpcm.wav"},
		{"testdata/11kgsm.wav"},
		{"testdata/11kulaw.wav"},
		{"testdata/32bit.wav"},
		{"testdata/8bit.wav"},
		{"testdata/8k16bitpcm.wav"},
		{"testdata/8k8bitpcm.wav"},
		{"testdata/8kadpcm.wav"},
		{"testdata/8kcelp.wav"},
		{"testdata/8kgsm.wav"},
		{"testdata/8kmp316.wav"},
		{"testdata/8kmp38.wav"},
		{"testdata/8ksbc12.wav"},
		{"testdata/8ktruespeech.wav"},
		{"testdata/8kulaw.wav"},
		{"testdata/bass.wav"},
		{"testdata/bwf.wav"},
		{"testdata/dirty-kick-24b441k.wav"},
		{"testdata/flloop.wav"},
		{"testdata/junkKick.wav"},
		{"testdata/kick-16b441k.wav"},
		{"testdata/kick.wav"},
		{"testdata/listChunkInHeader.wav"},
		{"testdata/listinfo.wav"},
		{"testdata/misaligned-chunk.wav"},
		{"testdata/padded24b.wav"},
		{"testdata/sample16bit.wav"},
		{"testdata/sample.avi"},
		{"testdata/sample.rmi"},
		{"testdata/sample.wav"},
	}

	rif := New(LoadData)

	for _, tc := range tt {
		b.Run(tc.pth, func(b *testing.B) {
			b.StopTimer()
			buf := &flexbuf.Buffer{}
			_, err := buf.ReadFrom(kit.OpenFile(b, tc.pth))
			if err != nil {
				b.Fatal(err)
			}

			b.StartTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				buf.SeekStart()
				if _, err := rif.ReadFrom(buf); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
