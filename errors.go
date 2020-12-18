package riff

import (
	"errors"
)

// Errors.
var (
	// ErrNotRIFF is returned when file is not in the RIFF format.
	ErrNotRIFF = errors.New("not RIFF file")

	// ErrTooShort is returned when chunk or field is shorter then its
	// defined length.
	ErrTooShort = errors.New("length too short")

	// ErrSkipDataMode is returned when decoder in SkipData mode
	// is used in write context (e.x. calling WriteTo method).
	ErrSkipDataMode = errors.New("decoder in meta only mode used in write context")
)

// Error format strings.
const (
	// errFmtDecode format string for chunk decoding errors.
	errFmtDecode = "error decoding %s chunk: %w"

	// errFmtEncode format string for chunk encoding errors.
	errFmtEncode = "error encoding %s chunk: %w"
)
