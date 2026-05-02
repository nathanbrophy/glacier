// SPDX-License-Identifier: Apache-2.0

package term

import "io"

// DownloadProgress wraps an io.Reader, updating an embedded *Progress as bytes
// are read. It is intended for use with httpc.WithProgress and analogous callers.
type DownloadProgress struct {
	*Progress
	Source io.Reader // invariant: non-nil
}

// NewDownloadProgress wraps r with a DownloadProgress backed by a Progress
// configured with total = contentLength. Pass contentLength = -1 for unknown length
// (indeterminate mode).
func NewDownloadProgress(r io.Reader, contentLength int64, label string, opts ...ProgressOption) *DownloadProgress {
	allOpts := make([]ProgressOption, 0, len(opts)+1)
	if label != "" {
		allOpts = append(allOpts, WithProgressLabel(label))
	}
	allOpts = append(allOpts, opts...)
	return &DownloadProgress{
		Progress: NewProgress(contentLength, allOpts...),
		Source:   r,
	}
}

// Read implements io.Reader. Transparent to callers: reads from Source, then
// calls p.Increment(n) with the byte count. Updates speed and ETA sliding window.
// Goroutine-safe.
func (d *DownloadProgress) Read(p []byte) (int, error) {
	n, err := d.Source.Read(p)
	if n > 0 {
		d.Increment(int64(n))
	}
	if err == io.EOF {
		d.Done()
	}
	return n, err
}
