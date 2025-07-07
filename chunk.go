package riff

import (
	"encoding/binary"
	"io"
)

// Chunk represents Resource Interchange File Format (RIFF) chunk decoder.
type Chunk interface {
	// ID returns a four-byte (uint32) ASCII identifier of the chunk.
	ID() uint32

	// Size returns chunk size in bytes. The chunk ID, size and extra
	// padding byte (if present) is not counted in the returned value.
	// If Chunk has been edited, this method should return the new size.
	Size() uint32

	// Type returns chunk type. Returns zero if the chunk doesn't have
	// the type field.
	Type() uint32

	// Multi returns true if there can be more than one chunk with a given ID
	// in the RIFF file.
	Multi() bool

	// Chunks returns all the sub-chunks of the chunk or empty slice if the
	// chunk doesn't support sub-chunks.
	Chunks() Chunks

	// Raw returns true if the chunk was decoded using ChunkRAWC
	// or ChunkRAWB decoder.
	Raw() bool

	// ReadFrom reads bytes from r decoding and validating a chunk.
	// It expects r to be in a position right after chunk ID.
	// It returns the actual number of bytes read and error if any.
	// If an error is returned, the number of read bytes might not be accurate.
	ReadFrom(r io.Reader) (int64, error)

	// WriteTo writes chunk (encodes) to w. Returns the number of bytes written
	// and error if any.
	WriteTo(w io.Writer) (int64, error)

	// Reset resets the chunk so it can be reused.
	Reset()
}

// List of most popular chunk IDs.
const (
	// IDJUNK represents "JUNK" chunk ID.
	IDJUNK uint32 = 0x4a554e4b

	// IDID3 represents "ID3 " chunk ID.
	IDID3 uint32 = 0x49443320
)

// Convenience imports.
var (
	le = binary.LittleEndian
	be = binary.BigEndian
)
