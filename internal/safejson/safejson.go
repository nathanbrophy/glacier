// SPDX-License-Identifier: Apache-2.0

// Package safejson wraps encoding/json with security guards for untrusted input.
// It caps input size at MaxFileSize and nesting depth at MaxDepth before
// delegating to the standard library decoder.
package safejson

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

const (
	// MaxFileSize is the maximum allowed JSON byte count (1 MiB).
	MaxFileSize int64 = 1 << 20

	// MaxDepth is the maximum allowed JSON nesting depth.
	MaxDepth = 32
)

// ErrTooLarge is returned when the JSON input exceeds MaxFileSize.
var ErrTooLarge = errors.New("safejson: input too large (max 1 MiB)")

// ErrDepthExceeded is returned when the JSON nesting depth exceeds MaxDepth.
var ErrDepthExceeded = errors.New("safejson: nesting depth exceeded (max 32)")

// Decode reads at most MaxFileSize bytes from r, validates the nesting depth,
// and decodes JSON into dst. Unknown fields are not rejected; use DecodeStrict
// if that behaviour is required.
//
// Returns ErrTooLarge or ErrDepthExceeded for structural violations; otherwise
// returns a standard json error.
func Decode(r io.Reader, dst any) error {
	data, err := readCapped(r)
	if err != nil {
		return err
	}
	if err := checkDepth(data); err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// DecodeStrict is like Decode but also rejects unknown fields.
func DecodeStrict(r io.Reader, dst any) error {
	data, err := readCapped(r)
	if err != nil {
		return err
	}
	if err := checkDepth(data); err != nil {
		return err
	}
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// readCapped reads from r up to MaxFileSize bytes and returns an error if the
// stream is longer.
func readCapped(r io.Reader) ([]byte, error) {
	limited := io.LimitReader(r, MaxFileSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > MaxFileSize {
		return nil, ErrTooLarge
	}
	return data, nil
}

// checkDepth scans JSON tokens and returns ErrDepthExceeded if the nesting
// depth of objects or arrays exceeds MaxDepth.
func checkDepth(data []byte) error {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	depth := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			// io.EOF or any scan error :  stop; let Unmarshal report parse errors.
			break
		}
		switch tok {
		case json.Delim('{'), json.Delim('['):
			depth++
			if depth > MaxDepth {
				return ErrDepthExceeded
			}
		case json.Delim('}'), json.Delim(']'):
			depth--
		}
	}
	return nil
}
