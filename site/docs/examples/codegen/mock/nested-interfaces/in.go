//go:build glacier_codegen_fixture

// Package rw defines read-write interfaces for streaming data.
package rw

// Reader reads bytes.
type Reader interface {
	// Read reads up to len(p) bytes into p.
	Read(p []byte) (n int, err error)
}

// ReadWriter reads, writes, and closes a byte stream.
//
// +glacier:mock
type ReadWriter interface {
	Reader

	// Write writes len(p) bytes from p.
	Write(p []byte) (n int, err error)

	// Close closes the stream.
	Close() error
}
