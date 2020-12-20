package test

import (
	"encoding/binary"
	"io"
	"testing"
)

// WriteUint32LE writes v encoded using little endian to dst.
// Calls t.Fatal() on error.
func WriteUint32LE(t *testing.T, dst io.Writer, v uint32) {
	t.Helper()
	if err := binary.Write(dst, binary.LittleEndian, &v); err != nil {
		t.Fatal(err)
	}
}

// WriteUint32BE writes v encoded using big endian to dst.
// Calls t.Fatal() on error.
func WriteUint32BE(t *testing.T, dst io.Writer, v uint32) {
	t.Helper()
	if err := binary.Write(dst, binary.BigEndian, &v); err != nil {
		t.Fatal(err)
	}
}

// WriteUint16LE writes v encoded using little endian to dst.
// Calls t.Fatal() on error.
func WriteUint16LE(t *testing.T, dst io.Writer, v uint16) {
	t.Helper()
	if err := binary.Write(dst, binary.LittleEndian, &v); err != nil {
		t.Fatal(err)
		return
	}
}

// ReadFrom writes to dst from src. Calls t.Fatal() on error.
func ReadFrom(t *testing.T, dst io.ReaderFrom, src io.Reader) int64 {
	t.Helper()
	n, err := dst.ReadFrom(src)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// WriteTo writes from src to dst. Calls t.Fatal() on error.
func WriteTo(t *testing.T, dst io.Writer, src io.WriterTo) int64 {
	t.Helper()
	n, err := src.WriteTo(dst)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// WriteBytes writes p to dst. Calls t.Fatal() on error.
func WriteBytes(t *testing.T, dst io.Writer, p []byte) int {
	t.Helper()
	n, err := dst.Write(p)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// WriteByte writes byte c to dst. Calls t.Fatal() on error.
func WriteByte(t *testing.T, dst io.ByteWriter, c byte) {
	t.Helper()
	if err := dst.WriteByte(c); err != nil {
		t.Fatal(err)
	}
}

// Skip4B reads exactly 4 bytes from r and discards them.
// Calls t.Fatal() on error.
func Skip4B(t *testing.T, r io.Reader) {
	t.Helper()
	n, err := r.Read(make([]byte, 4))
	if err != nil {
		t.Fatal(n)
	}
	if n != 4 {
		t.Fatalf("expected to read 4 bytes")
	}
}

// IsAllRead returns true if all bytes from the underlying buffer
// have been read.
func IsAllRead(r io.Reader) bool {
	_, err := r.Read(make([]byte, 1))
	return err == io.EOF
}
