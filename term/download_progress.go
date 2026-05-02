// SPDX-License-Identifier: Apache-2.0

package term

import "io"

// DownloadProgress wraps an io.Reader, updating an embedded *Progress as bytes
// are read. It is intended for use with httpc.WithProgress and analogous callers.
//
// Because DownloadProgress embeds *Progress, all Progress methods are promoted:
// Run(ctx), Close(), Done(), Set(n), Increment(n), and Render(). A caller can
// use dp.Run(ctx) for self-contained animation, or inject a shared Animator via
// WithProgressAnimator to coordinate multiple animations.
type DownloadProgress struct {
	*Progress
	Source io.Reader // invariant: non-nil
}

// NewDownloadProgress wraps r with a DownloadProgress backed by a Progress
// configured with total = contentLength. Pass contentLength = -1 for unknown length
// (indeterminate mode).
//
// To inject a shared Animator pass WithProgressAnimator(a) as one of opts.
// When injected, dp.Run(ctx) is a no-op; the caller drives the shared Animator.
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
// When Source reaches EOF, Done is called automatically so the animation
// terminates without the caller needing to call Done or Close explicitly.
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
