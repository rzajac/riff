// Package riff provides low level tools for working with files in Resource
// Interchange File Format (RIFF).
package riff

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// IDRIFF represents "RIFF" chunk ID.
const IDRIFF uint32 = 0x52494646

// RIFF file types.
// Supported file types as defined in [RIFF] chunk.
const (
	// TypeWAVE represents the "WAVE" file type.
	TypeWAVE uint32 = 0x57415645

	// TypeAVI represents the "AVI " file type.
	TypeAVI uint32 = 0x41564920

	// TypeRMID represents the "RMID" file type.
	TypeRMID uint32 = 0x524d4944
)

// Maker is a function signature for instantiating chunk decoder.
type Maker func() Chunk

// IDMaker is a function signature for instantiating chunk decoders for id.
type IDMaker func(id uint32) Chunk

// RIFF represents a file in Resource Interchange File Format.
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

	// Main chunk processing options.
	opts Opts
}

// New returns new instance of Riff with all "out-of-the-box" chunk decoders
// registered.
func New(opts ...OptsFn) *RIFF {
	reg := NewRegistry(RAWCMake(opts...))

	// Register "out of the box" chunk decoders.
	reg.Register(IDfmt, FMTMake(opts...))
	reg.Register(IDdata, DATAMake(opts...))
	reg.Register(IDLIST, LISTMake(reg, opts...))
	reg.Register(IDsmpl, SMPLMake(opts...))

	return Bare(reg, opts...)
}

// Bare returns a new instance of [RIFF] without any chunk decoders registered.
// If reg is set to nil, it will be created with the default raw chunk decoder
// (ChunkRAWC) set to skip data.
func Bare(reg *Registry, opts ...OptsFn) *RIFF {
	if reg == nil {
		reg = NewRegistry(RAWCMake())
	}
	rif := &RIFF{
		chunks: make([]Chunk, 0, 4),
		reg:    reg,
		opts:   NewOpts(opts...),
	}
	return rif
}

// Compose returns a new instance of [RIFF] based on [Chunks].
// This method allows you to create a copy of the original RIFF with
// modifications.
func Compose(chs Chunks) *RIFF {
	rif := Bare(nil)
	rif.Modify(chs)

	return rif
}

func (rif *RIFF) ID() uint32     { return IDRIFF }
func (rif *RIFF) Size() uint32   { return rif.size }
func (rif *RIFF) Type() uint32   { return rif.riffType }
func (rif *RIFF) Multi() bool    { return false }
func (rif *RIFF) Chunks() Chunks { return rif.chunks }
func (rif *RIFF) Raw() bool      { return false }

func (rif *RIFF) SetType(t uint32) { rif.riffType = t }

// IsRegistered returns true if decoder for id is registered.
func (rif *RIFF) IsRegistered(id uint32) bool {
	return rif.reg.Has(id)
}

func (rif *RIFF) ReadFrom(r io.Reader) (int64, error) {
	rif.Reset()

	var err error
	var id uint32

	if err = ReadChunkID(r, &id); err != nil {
		return 0, err
	}

	if id != IDRIFF {
		return 4, ErrNotRIFF
	}

	cr, size, err := rif.opts.ChunkReader(r)
	if err != nil {
		return 4, fmt.Errorf(errFmtDecode, Uint32(IDRIFF), err)
	}
	cr.setOffset(int(cr.Count()) + 4) // IDRIFF was already read.

	// Main chunk RIFF size should never be odd. Since it's a small infraction,
	// we can simply round it up.
	rif.size = RealSize(size)

	if err = binary.Read(cr, be, &rif.riffType); err != nil {
		return cr.Count(), fmt.Errorf(errFmtDecode, Uint32(IDRIFF), err)
	}

	for {
		if err = ReadChunkID(cr, &id); err != nil {
			if !errors.Is(err, io.EOF) {
				return cr.Count(), fmt.Errorf(errFmtReadingRIFF, cr.Count(), err)
			}
			break
		}

		_, err = rif.decodeChunk(id, cr)
		if err != nil {
			return cr.Count(), fmt.Errorf(errFmtReadingRIFF, cr.Count(), err)
		}
	}
	// EOF reached.
	readSize := uint32(cr.Count() - 8)

	if !rif.opts.IgnoreDeclaredSize() && rif.size > readSize {
		return cr.Count(), fmt.Errorf(
			"RIFF declared %d bytes, decoder read %d bytes: %w",
			rif.size,
			readSize,
			io.ErrUnexpectedEOF,
		)
	}

	if rif.opts.IgnoreDeclaredSize() && rif.size != readSize {
		// Size needs to be corrected.
		rif.size = readSize
	}

	return cr.Count(), nil
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

	if err = binary.Write(w, be, rif.riffType); err != nil {
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

// Modify set a new set of the chunks.
func (rif *RIFF) Modify(chs Chunks) {
	rif.chunks = chs
	// Recalculate chunks size.
	rif.size = 4 + rif.chunks.Size()
}
