package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rzajac/riff"
)

func main() {
	if len(os.Args) < 3 {
		printHelp()
		os.Exit(1)
	}

	fil, err := os.Open(os.Args[2])
	checkErr(err)
	defer fil.Close()

	switch os.Args[1] {
	case "raw-chunk-print":
		err = rawChunkPrint(fil)

	case "fmt-chunk-print":
		err = fmtChunkPrint(fil)

	case "data-chunk-print":
		err = dataChunkPrint(fil)

	default:
		printHelp()
	}

	checkErr(err)
}

// rawChunkPrint prints fmt chunk as slice of bytes.
func rawChunkPrint(fil *os.File) error {
	// Don't register any decoders and load all the data.
	rif := riff.New(riff.LoadData)

	// Decode chunks.
	if _, err := rif.ReadFrom(fil); err != nil {
		return err
	}

	// Get all sub-chunks.
	chs := rif.Chunks()

	// Check the file has chunk we are interested in.
	if chs.Count(riff.IDfmt) == 0 {
		return fmt.Errorf(
			"chunk %s not present in the file %s\n",
			riff.Uint32(riff.IDfmt),
			fil.Name(),
		)
	}

	chi := chs.First(riff.IDfmt)

	// We did not register any chunk decoders so we
	// are sure riff.ChunkRAWC will be used.
	ch := chi.(*riff.ChunkRAWC)

	// Read chunk and print it.
	buf := make([]byte, ch.Size())
	if _, err := io.ReadFull(ch.Body(), buf); err != nil {
		return err
	}

	fmt.Println(buf)

	return nil
}

// fmtChunkPrint prints values defined in fmt chunk.
func fmtChunkPrint(fil *os.File) error {
	// Load only metadata (default).
	rif := riff.New(riff.SkipData)

	// Decode chunks.
	if _, err := rif.ReadFrom(fil); err != nil {
		return err
	}

	// Get all sub-chunks.
	chs := rif.Chunks()

	// Check the file has chunk we are interested in.
	if chs.Count(riff.IDfmt) == 0 {
		return fmt.Errorf(
			"chunk %s not present in the file %s\n",
			riff.Uint32(riff.IDfmt),
			fil.Name(),
		)
	}

	// Get the first fmt chunk.
	// There should be only one of those in RIFF file,
	// decoding would error out if there were more then one
	// because ChunkFMT returns false from Multi() method.
	chi := chs.First(riff.IDfmt)

	// We registered decoder ourselves so we are sure of its type.
	ch := chi.(*riff.ChunkFMT)

	fmt.Printf("Chunk fmt\n")
	fmt.Printf(" - Compression Code: %#04x\n", ch.CompCode)
	fmt.Printf(" - Channel Count: %d\n", ch.ChannelCnt)
	fmt.Printf(" - Sample Rate: %d\n", ch.SampleRate)
	fmt.Printf(" - Average Bit Rate: %d\n", ch.AvgByteRate)
	fmt.Printf(" - Block Align: %d\n", ch.BlockAlign)
	fmt.Printf(" - Bits Per Sample: %d\n", ch.BitsPerSample)

	extra, err := ioutil.ReadAll(ch.Extra())
	if err != nil {
		return err
	}
	fmt.Printf(" - Extra fmt Bytes: %v\n", extra)
	fmt.Println()

	return nil
}

// dataChunkPrint set data chunk.
func dataChunkPrint(fil *os.File) error {
	rif := riff.New(riff.LoadData)

	if _, err := rif.ReadFrom(fil); err != nil {
		return err
	}

	// Get all sub-chunks.
	chs := rif.Chunks()

	// Check the file has chunk we are interested in.
	if chs.Count(riff.IDdata) == 0 {
		return fmt.Errorf(
			"chunk %s not present in the file %s\n",
			riff.Uint32(riff.IDdata),
			fil.Name(),
		)
	}

	// Get the data chunk.
	// There should be only one of those in RIFF file,
	// decoding would error out if there were more then one
	// because ChunkDATA returns false from Multi() method.
	chi := chs.First(riff.IDdata)

	// We registered decoder ourselves so we are sure of its type.
	ch := chi.(*riff.ChunkDATA)

	buf := make([]byte, 10)
	if _, err := io.ReadFull(ch.Data(), buf); err != nil {
		return err
	}
	fmt.Printf("first 10 bytes of data chunk: %v\n", buf)

	return nil
}

func registerCustom(fil *os.File) error {
	// Create registry with raw parsers skipping data.
	reg := riff.NewRegistry(riff.RAWCMake(riff.SkipData))

	// Register parsers for chunks we are interested in.

	// Parse fmt and LIST chunk(s).
	reg.Register(riff.IDfmt, riff.FMTMake)
	reg.Register(riff.IDLIST, riff.LISTMake(riff.LoadData, reg))
	// Skip reading data in data chunk.
	reg.Register(riff.IDdata, riff.DATAMake(riff.SkipData))

	// Not registered chunks will be decoded by RAWChunk.

	rif := riff.Bare(reg)
	if _, err := rif.ReadFrom(fil); err != nil {
		return err
	}

	ch := rif.Chunks().First(riff.IDLIST)
	_ = ch

	// Work with the chunk.

	return nil
}

func printHelp() {
	hlp := `%s action 
action:
  raw-chunk-print [file]   print fmt chunk as raw bytes
  fmt-chunk-print [file]   print human readable values from fmt chunk
  data-chunk-print [file]  print first few bytes of data chunk
`
	fmt.Printf(hlp, filepath.Base(os.Args[0]))
}

// checkErr prints error message and exits program if err is not nil.
func checkErr(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
