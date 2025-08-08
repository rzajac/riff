package riff

import (
	"bytes"
	"os"
	"testing"

	"github.com/ctx42/memfs/pkg/memfs"
	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/kit"
	"github.com/ctx42/testing/pkg/must"
)

func Test_RIFF_New(t *testing.T) {
	// --- When ---
	rif := New(LoadData)

	// --- Then ---
	assert.Equal(t, IDRIFF, rif.ID())
	assert.Equal(t, uint32(0), rif.Size())
	assert.Equal(t, uint32(0), rif.Type())
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
	assert.Equal(t, IDRIFF, rif.ID())
	assert.Equal(t, uint32(0), rif.Size())
	assert.Equal(t, uint32(0), rif.Type())
	assert.False(t, rif.Multi())

	assert.False(t, rif.IsRegistered(IDfmt))
	assert.False(t, rif.IsRegistered(IDdata))
}

func Test_RIFF_Bare_NilRegistry(t *testing.T) {
	// --- When ---
	rif := Bare(nil)

	// --- Then ---
	assert.Equal(t, IDRIFF, rif.ID())
	assert.Equal(t, uint32(0), rif.Size())
	assert.Equal(t, uint32(0), rif.Type())
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
			fil := must.Value(os.Open(tc.pth))

			// --- When ---
			n, err := rif.ReadFrom(fil)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.size, n)
			assert.Equal(t, uint32(tc.size-8), rif.Size())
			assert.Equal(t, tc.typ, rif.Type())
			assert.Len(t, tc.chunks, rif.Chunks())
		})
	}
}

func Test_RIFF_CorrectingSize(t *testing.T) {
	// --- Given ---
	rif := New(LoadData)
	_, err := rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
	assert.NoError(t, err)

	// --- When ---
	dst := &bytes.Buffer{}
	n, err := rif.WriteTo(dst)

	// --- Then ---
	assert.NoError(t, err)

	assert.Equal(t, int64(27074), n)
	assert.Equal(t, uint32(27074), rif.Size()+8)
	assert.Equal(t, "ebdf4fefcbbb804b44cddbc50521ba29f833e05c", kit.SHA1Reader(dst))
}

func Test_RIFF_WriteTo_SmokeTest(t *testing.T) {
	rif := New(LoadData)

	tt := []struct {
		pth  string
		size int64
		hash string
	}{
		{"testdata/11k16bitpcm.wav", 304578, "f183c8045848f7fe821f1f26fdffaeb794baad6c"},
		{"testdata/11k8bitpcm.wav", 152312, "51218ead7323766530094ea37211cdc190bade2b"},
		{"testdata/11kadpcm.wav", 77252, "27f7d01150347f039d1997c4ef820a10bf9cbc0a"},
		{"testdata/11kgsm.wav", 31000, "5d7f26bac9d3270c424446936360b0ea04f3ccd9"},
		{"testdata/11kulaw.wav", 152326, "22699d04d0d30adf1a9137b79bb87354f5b72683"},
		{"testdata/32bit.wav", 352844, "46bc96f0953df4f6f36053e65733f5e32d276167"},
		{"testdata/8bit.wav", 88244, "ef28c1a4fc75b0e4630499b1899bef244c837424"},
		{"testdata/8k16bitpcm.wav", 221026, "666c63ca5610b6d8cabbaf5f8b84427657f0ec50"},
		{"testdata/8k8bitpcm.wav", 110532, "25046c120e524475fae9417c2cc938e1bbe95231"},
		{"testdata/8kadpcm.wav", 56072, "9387acf987806693dd2f14feaaaa49ae2a79030a"},
		{"testdata/8kcelp.wav", 8350, "f7645b8f9bed2f397f426daf02273a1731652f4c"},
		{"testdata/8kgsm.wav", 22550, "35cd3a68a9ab8e706a2ca8ac75a8d9b7c70f4545"},
		{"testdata/8kmp316.wav", 27386, "4652808d259bcf19d129bbfd2817e56466269808"},
		{"testdata/8kmp38.wav", 13584, "71b7463d5e66e418c242a56a04ba15b3de329822"},
		{"testdata/8ksbc12.wav", 19658, "6e7e5f4f2854f29e9ea87070a94a04b253020c49"},
		{"testdata/8ktruespeech.wav", 14842, "b8d289588a9a195042fd3482b12cb3a659cae26b"},
		{"testdata/8kulaw.wav", 110546, "f8f7ff5311da5a59a91f41ac679fbc23a8381a99"},
		{"testdata/bass.wav", 143786, "5c547df0f3d24da873d92e28ca3ecd22b83ad92b"},
		{"testdata/dirty-kick-24b441k.wav", 132136, "61b6cebfe073f34a81bc4352e69b5a044b07e188"},
		{"testdata/flloop.wav", 434838, "d12b046588af5475d846dd4a72f957cf03a33591"},
		{"testdata/junkKick.wav", 83084, "5974547313bcb8804618a3048f6902a429d62a4e"},
		{"testdata/kick-16b441k.wav", 31692, "1d7dbd0fe12ce2f8ec33ef5f90e271125a83ea94"},
		{"testdata/kick.wav", 9012, "cfa0812a881c3d3f2a2783b5b7a6d2ba62f7a1aa"},
		{"testdata/listChunkInHeader.wav", 104196, "9d722b63acf00c0031a5aaf7b3f73291321aafe3"},
		{"testdata/listinfo.wav", 176720, "8dd085ecf49c1ca781857dd1f26b16b636b7ab07"},
		{"testdata/misaligned-chunk.wav", 3441572, "dda435791002e772da033e3e7f09a3854c9e8d76"},
		{"testdata/padded24b.wav", 24908, "25b88e9afdc533be201b4393bc489cbfa61ec2e7"},
		{"testdata/sample.avi", 230264, "8f30db3104fafec017241b63fbba6588ee8cd5b4"},
		{"testdata/sample.rmi", 29640, "aa4cdab911e956439f8553772b56cd01993e9e12"},
	}

	for _, tc := range tt {
		t.Run(tc.pth, func(t *testing.T) {
			// --- Given ---
			_, err := rif.ReadFrom(must.Value(os.Open(tc.pth)))
			assert.NoError(t, err)

			// --- When ---
			buf := &bytes.Buffer{}
			n, err := rif.WriteTo(buf)

			// --- Then ---
			assert.NoError(t, err)

			// nolint: gocritic
			// fil := kit.CreateFile(t, filepath.Join("tmp", filepath.Base(tc.pth)))
			// content := kit.ReadAll(t, buf)
			// fil.Write(content)
			// buf = bytes.NewBuffer(content)

			assert.Equal(t, tc.size, n)
			size := int64(RealSize(rif.Size()) + 8)
			assert.Equal(t, tc.size, size)
			assert.Equal(t, tc.hash, kit.SHA1Reader(buf))
		})
	}
}

func Test_Compose(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		rif := New(SkipData)
		_, _ = rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
		chs := rif.Chunks()

		// --- When ---
		nw := Compose(chs.Remove(IDfmt))

		// --- Then ---
		assert.Equal(t, 14, len(chs.IDs()))
		assert.Equal(t, 13, len(nw.Chunks()))
	})

	t.Run("size ok", func(t *testing.T) {
		// --- Given ---
		rif := New(SkipData)
		_, _ = rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
		chs := rif.Chunks()

		// --- When ---
		nw := Compose(chs)

		// --- Then ---
		assert.Equal(t, rif.Size(), nw.Size())
		assert.Equal(t, len(chs.IDs()), len(nw.Chunks().IDs()))
	})
}

func Test_Modify(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		rif := New(SkipData)
		_, _ = rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
		s := rif.Size()
		chs := rif.Chunks()
		f := chs.First(IDfmt)

		// --- When ---
		rif.Modify(chs.Remove(IDfmt))

		// --- Then ---
		assert.Equal(t, s-(f.Size()+8), rif.Size())
		assert.Equal(t, 13, len(rif.Chunks().IDs()))
	})
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
			buf := must.Value(memfs.NewFile("file"))
			_, err := buf.ReadFrom(must.Value(os.Open(tc.pth)))
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
