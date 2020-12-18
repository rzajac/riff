package riff

import (
	"io"
)

// Chunks represents slice of sub-chunks.
type Chunks []Chunk

// First returns first chunk with given ID.
func (chs Chunks) First(id uint32) Chunk {
	for _, ch := range chs {
		if ch.ID() == id {
			return ch
		}
	}
	return nil
}

// Count returns number of chunks with given id.
func (chs Chunks) Count(id uint32) int {
	var cnt int
	for _, ch := range chs {
		if ch.ID() == id {
			cnt++
		}
	}
	return cnt
}

// IDs returns chunk IDs in order they were seen in the file.
func (chs Chunks) IDs() []uint32 {
	ids := make([]uint32, len(chs))
	for i, ch := range chs {
		ids[i] = ch.ID()
	}
	return ids
}

// Size returns size (with padding bytes) of all the chunks in the collection.
func (chs Chunks) Size() uint32 {
	var size uint32
	for _, ch := range chs {
		size += RealSize(ch.Size()) + 8 // Add 8 for chunk ID and size fields.
	}
	return size
}

// WriteTo writes all the chunks in the collection to w.
func (chs Chunks) WriteTo(w io.Writer) (n int64, err error) {
	var sum int64
	for _, ch := range chs {
		n, err = ch.WriteTo(w)
		sum += n
		if err != nil {
			return sum, err
		}
	}
	return sum, nil
}
