## RIFF file decoder and encoder

[![Go Report Card](https://goreportcard.com/badge/github.com/rzajac/riff)](https://goreportcard.com/report/github.com/rzajac/riff)
[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/rzajac/riff)


Package provides low level tools for working with files in Resource Interchange
File Format (RIFF).

Supported chunks:

* RIFF
    * data
    * fmt
    * LIST
        * INFO
        * adtl
            * labl
            * ltxt
    * sampl

Package provides a way to register custom decoders for chunks not yet supported.

## Installation

```
go get github.com/rzajac/riff
```

## Examples

See more examples in [_examples](_examples) directory.

### Print fmt chunk info

```go
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/rzajac/riff"
)

func main() {
	// Open file.
	src, err := os.Open("path")
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	// Create RIFF instance, read only metadata.
	rif := riff.New()

	// Decode chunks.
	_, err = rif.ReadFrom(src)
	if err != nil {
		log.Fatal(err)
	}

	// Get all sub-chunks.
	chs := rif.Chunks()

	// Check the file has chunk we are interested in.
	if chs.Count(riff.IDfmt) == 0 {
		fmt.Printf(
			"chunk %s not present in the file %s\n",
			riff.Uint32(riff.IDfmt),
			src.Name(),
		)
		return
	}

	// Get the first fmt chunk.
	// There should be only one of those in RIFF file,
	// decoding would error out if there were more than one.
	chi := chs.First(riff.IDfmt)

	// Cast to concrete type.
	ch := chi.(*riff.ChunkFMT)

	fmt.Printf("Chunk fmt\n")
	fmt.Printf(" - Compression Code: %#04x\n", ch.CompCode)
	fmt.Printf(" - Channel Count: %d\n", ch.ChannelCnt)
	fmt.Printf(" - Sample Rate: %d\n", ch.SampleRate)
	fmt.Printf(" - Average Bit Rate: %d\n", ch.AvgByteRate)
	fmt.Printf(" - Block Align: %d\n", ch.BlockAlign)
	fmt.Printf(" - Bits Per Sample: %d\n", ch.BitsPerSample)

	extra, err := io.ReadAll(ch.Extra())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(" - Extra fmt Bytes: %v\n", extra)
	fmt.Println()
}
```

### Reuse instance when decoding multiple files.

```go
package main

import (
	"log"
	"os"

	"github.com/rzajac/riff"
)

func main() {
	// Create RIFF instance, read only metadata.
	rif := riff.New()

	for _, pth := range []string{"pth1", "pth2", "pth3"} {
		src, err := os.Open(pth)
		if err != nil {
			log.Fatal(err)
		}

		_, err = rif.ReadFrom(src)
		if err != nil {
			log.Fatal(err)
		}

		// Work on RIFF instance.

		err = src.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
}
```

### Register custom chunk decoders.

```go
package main

import (
	"github.com/rzajac/riff"
)

func main()  {
	// Create registry and tell it what to use to handle unknown chunks.
	reg := riff.NewRegistry(riff.RAWCMake())

	// Register one or more custom decoders.
	reg.Register(riff.IDfmt, riff.FMTMake())

	// Create RIFF decoder.
	_ = riff.Bare(reg)
}
```

In the example above only "fmt " chunks will be decoded. The rest will be 
skipped by `ChunkRAWC` decoder.

### Save edits

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rzajac/riff"
)

func main() {
	// Open file.
	src, err := os.Open("path")
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	// Create RIFF instance, load chunks' content.
	rif := riff.New(riff.WithLoadData())

	// Decode chunks.
	_, err = rif.ReadFrom(src)
	if err != nil {
		log.Fatal(err)
	}

	// Get all sub-chunks.
	chs := rif.Chunks()

	// Check the file has chunk we are interested in.
	if chs.Count(riff.IDfmt) == 0 {
		fmt.Printf(
			"chunk %s not present in the file %s\n",
			riff.Uint32(riff.IDfmt),
			src.Name(),
		)
		return
	}

	// Get the first fmt chunk.
	// There should be only one of those in RIFF file,
	// decoding would error out if there were more than one.
	chi := chs.First(riff.IDfmt)

	// Cast to concrete type.
	ch := chi.(*riff.ChunkFMT)

	// Edit values.
	ch.SampleRate = 44100

	dst, err := os.Create("new_file")
	if err != nil {
		log.Fatal(err)
	}

	_, err = rif.WriteTo(dst)
	if err != nil {
		log.Fatal(err)
	}
}
```

## FAQ

### Can I reuse RIFF instance for multiple files?

Yes.

### Is the library thread safe?

No. It's up to the user to handle concurrency.

### How do I create custom decoder for chunk X?

Implement interface [Chunk](chunk.go) and register it. See example above.

### How unsupported chunks are handled?

Not supported chunks are decoded and encoded by `ChunkRAWC` so it's still
possible to decode -> edit -> encode a file with not supported chunks. 

## License

BSD-2-Clause