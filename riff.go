// Package riff provides low level tools for working with files in Resource
// Interchange File Format (RIFF).
package riff

import (
	"encoding/binary"
	"fmt"
	"io"
)

// IDRIFF represents "RIFF" chunk ID.
const IDRIFF uint32 = 0x52494646

// RIFF file types.
// Supported file types as defined in RIFF chunk.
const (
	// TypeWAVE represents "WAVE" file type.
	TypeWAVE uint32 = 0x57415645

	// TypeAVI represents "AVI " file type.
	TypeAVI uint32 = 0x41564920

	// TypeRMID represents "RMID" file type.
	TypeRMID uint32 = 0x524d4944
)

// Maker is a function signature for instantiating chunk decoder.
type Maker func() Chunk

// IDMaker is a function signature for instantiating chunk decoders for id.
type IDMaker func(id uint32) Chunk

// RIFF represents file in Resource Interchange File Format.
type RIFF struct {
	// Chunk size in bytes.
	// The ID and extra padding byte is not counted in the chunk size.
	size uint32

	// Determines the type of the resource (e.g.: WAVE).
	riffType uint32

	// List of decoded file chunks in order they appeared in the file.
	chunks Chunks

	// Registered chunk decoders.
	reg *Registry

	// Controls how chunks are processed. If set to false then only the
	// metadata about the chunks are read the rest is skipped. This improves
	// performance in cases when user is only interested in metadata and isn't
	// going to modify or write the RIFF file.
	// It's up to chunk decoder to decide what is considered data vs metadata.
	// By default it is set to false.
	load bool
}

const (
	// LoadData is RIFF constructor option instructing decoders to load chunk's
	// metadata and data.
	LoadData bool = true

	// SkipData is RIFF constructor option instructing decoders to skip chunk's
	// data and load only metadata.
	SkipData bool = false
)

// New returns new instance of Riff with all "out of the box" chunk decoders
// registered.
func New(load bool) *RIFF {
	reg := NewRegistry(RAWCMake(load))

	// Register "out of the box" chunk decoders.
	reg.Register(IDfmt, FMTMake)
	reg.Register(IDdata, DATAMake(load))
	reg.Register(IDLIST, LISTMake(load, reg))
	reg.Register(IDsmpl, SMPLMake)

	return Bare(reg)
}

// Bare returns new instance of RIFF without any chunk decoders registered.
// If reg is set to nil it will be created with default raw chunk decoder
// (ChunkRAWC) set to skip data.
func Bare(reg *Registry) *RIFF {
	if reg == nil {
		reg = NewRegistry(RAWCMake(SkipData))
	}
	rif := &RIFF{
		chunks: make([]Chunk, 0, 4),
		reg:    reg,
	}
	return rif
}

func (rif *RIFF) ID() uint32     { return IDRIFF }
func (rif *RIFF) Size() uint32   { return rif.size }
func (rif *RIFF) Type() uint32   { return rif.riffType }
func (rif *RIFF) Multi() bool    { return false }
func (rif *RIFF) Chunks() Chunks { return rif.chunks }
func (rif *RIFF) Raw() bool      { return false }

// IsRegistered returns true if decoder for id is registered.
func (rif *RIFF) IsRegistered(id uint32) bool {
	return rif.reg.Has(id)
}

func (rif *RIFF) ReadFrom(r io.Reader) (int64, error) {
	rif.Reset()

	var err error
	var sum int64
	var id uint32

	if err = ReadChunkID(r, &id); err != nil {
		return 0, err
	}
	sum += 4

	if id != IDRIFF {
		return sum, ErrNotRIFF
	}

	if rif.size, err = ReadChunkSize(r); err != nil {
		return sum, err
	}
	sum += 4

	if err := binary.Read(r, be, &rif.riffType); err != nil {
		return sum, fmt.Errorf(errFmtDecode, Uint32(IDRIFF), err)
	}
	sum += 4

	var n int64
	for {
		if err = ReadChunkID(r, &id); err != nil {
			break
		}
		sum += 4

		n, err = rif.decodeChunk(id, r)
		sum += n
		if err != nil {
			break
		}
	}

	// Size needs to be corrected.
	if err == io.EOF && rif.size != uint32(sum-8) {
		rif.size = uint32(sum - 8)
	}

	if err == io.EOF {
		return sum, nil
	}

	return sum, fmt.Errorf("error reading chunk ID: %w", err)
}

func (rif *RIFF) WriteTo(w io.Writer) (int64, error) {
	var sum int64

	// Recalculate chunks size and add RIFF type.
	rif.size = 4 + rif.chunks.Size()

	n, err := WriteIDAndSize(w, IDRIFF, rif.size)
	sum += n
	if err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDRIFF), err)
	}

	if err := binary.Write(w, be, rif.riffType); err != nil {
		return sum, fmt.Errorf(errFmtEncode, Uint32(IDRIFF), err)
	}
	sum += 4

	n, err = rif.chunks.WriteTo(w)
	sum += n
	if err != nil {
		return sum, err
	}

	return sum, nil
}

// Reset resets instance so it can be reused.
func (rif *RIFF) Reset() {
	for _, ch := range rif.chunks {
		rif.reg.Put(ch)
	}
	rif.chunks = rif.chunks[:0]
}

// decodeChunk decodes a chunk with id.
func (rif *RIFF) decodeChunk(id uint32, r io.Reader) (int64, error) {
	if rif.chunks.Count(id) > 0 && !rif.chunks.First(id).Multi() {
		return 0, fmt.Errorf("chunk %s (0x%x) already seen", Uint32(id), id)
	}
	dec := rif.reg.Get(id)
	dec.Reset()
	n, err := dec.ReadFrom(r)
	if err != nil {
		return n, err
	}
	rif.chunks = append(rif.chunks, dec)
	return n, nil
}
