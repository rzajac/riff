package riff

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
)

// StrToID converts four byte ASCII string to uint32.
// If string is shorter then 4 bytes it will be padded with spaces.
// If string is longer then four bytes it will be trimmed to 4 bytes.
func StrToID(s string) uint32 {
	b := []byte(s)
	if len(b) > 4 {
		b = b[:4]
	}
	if len(b) < 4 {
		b = append(b, bytes.Repeat([]byte{' '}, 4-len(b))...)
	}
	var id uint32
	for _, b := range b {
		id = (id << 8) | uint32(b)
	}
	return id
}

// ReadChunkID reads four byte chunk ID.
// It returns io.ErrUnexpectedEOF if there are less then four bytes in
// the reader.
func ReadChunkID(r io.Reader, id *uint32) error {
	if err := binary.Read(r, be, id); err != nil {
		return err
	}
	return nil
}

// ReadChunkSize reads four byte chunk size.
// It returns io.ErrUnexpectedEOF if there are less then four bytes in
// the reader.
func ReadChunkSize(r io.Reader) (uint32, error) {
	var size uint32
	if err := binary.Read(r, le, &size); err != nil {
		return 0, err
	}
	return size, nil
}

// LimitedRead reads size bytes to buf from r. It returns io.ErrUnexpectedEOF
// if number of read bytes is less then size.
func LimitedRead(src io.Reader, size uint32, dst io.ReaderFrom) error {
	n, err := dst.ReadFrom(io.LimitReader(src, int64(size)))
	if err != nil {
		return err
	}
	if n < int64(size) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// SkipN skips n bytes from reader r.
// If reader implements io.Seeker SkipN will use it to skip n bytes,
// otherwise SkipN will read n bytes and discard them.
func SkipN(r io.Reader, n uint32) error {
	// Check if r implements Seeker so we can just skip n bytes.
	if skr, ok := r.(io.Seeker); ok {
		_, err := skr.Seek(int64(n), io.SeekCurrent)
		if err != nil {
			return err
		}
		return nil
	}

	// If we cannot seek we read data to black hole.
	rf := ioutil.Discard.(io.ReaderFrom)
	if err := LimitedRead(r, n, rf); err != nil {
		return err
	}
	return nil
}

// RealSize returns size increased by padding byte if necessary.
func RealSize(size uint32) uint32 {
	if size%2 != 0 {
		return size + 1
	}
	return size
}

// ReadPaddingIf reads chunk padding byte from r if size is odd. Returns 1
// if byte was read zero otherwise. It returns read errors unless it
// was io.EOF error.
func ReadPaddingIf(r io.Reader, size uint32) (int64, error) {
	if size%2 == 0 {
		return 0, nil
	}

	n, err := io.CopyN(ioutil.Discard, r, 1)
	if err != nil {
		if err == io.EOF && n != 1 {
			err = io.ErrUnexpectedEOF
		}
		return 0, err
	}
	return n, nil
}

// WriteIDAndSize writes chunk id and size to writer w.
func WriteIDAndSize(w io.Writer, id, size uint32) (int64, error) {
	var sum int64
	if err := binary.Write(w, be, id); err != nil {
		return 0, err
	}
	sum += 4
	if err := binary.Write(w, le, size); err != nil {
		return sum, err
	}
	sum += 4
	return sum, nil
}

// WritePaddingIf writes padding byte (0) to writer w if size is odd.
func WritePaddingIf(w io.Writer, size uint32) (int64, error) {
	if size%2 == 1 {
		n, err := w.Write([]byte{0})
		return int64(n), err
	}
	return 0, nil
}

// TrimZeroRight removes all zero bytes from the end of the slice.
func TrimZeroRight(b []byte) []byte {
	for i := len(b) - 1; i > 0; i-- {
		if b[i] > 0 {
			return b[:i+1]
		}
	}
	return b[:0]
}

// linkids expects two chunk IDs and returns their ASCII representation
// concatenated with ':' character.
func linkids(id1, id2 uint32) string {
	return Uint32(id1).String() + ":" + Uint32(id2).String()
}

// grow grows byte slice b so it can fit n bytes.
func grow(b []byte, n int) []byte {
	if cap(b) >= n {
		return b[:n]
	}

	tmp := make([]byte, n)
	copy(tmp, b)
	return tmp
}
