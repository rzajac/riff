package riff

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Uint32 helper type wrapping uint32 and adding methods.
type Uint32 uint32

// String returns ASCII representation of the uint32 integer.
func (d Uint32) String() string {
	return string(byte(d>>24&0x000000FF)) +
		string(byte(d>>16&0x000000FF)) +
		string(byte(d>>8&0x000000FF)) +
		string(byte(d&0x000000FF))
}

func (d Uint32) Read(p []byte) (int, error) {
	if len(p) < 4 {
		return 0, fmt.Errorf("buffer too small for uint32: %w", ErrTooShort)
	}
	binary.BigEndian.PutUint32(p, uint32(d))
	return 4, io.EOF
}
