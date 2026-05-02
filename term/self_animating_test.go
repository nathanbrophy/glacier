// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/term"
)

// TestSpinnerRunSelfContained verifies that Spinner.Run blocks until ctx cancels.
func TestSpinnerRunSelfContained(t *testing.T) {
	t.Parallel()

	sp := term.Spinner("loading", term.WithSpinnerFrames([]string{"-", "\\", "|", "/"}))

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- sp.Run(ctx)
	}()

	select {
	case err := <-done:
		// ErrCancelled is expected when ctx is cancelled.
		if err == nil {
			t.Error("Spinner.Run returned nil, want ErrCancelled")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Spinner.Run did not return after ctx cancellation")
	}
}

// TestSpinnerRunWithInjectedAnimator verifies that Run is a no-op when an
// Animator is injected via WithSpinnerAnimator.
func TestSpinnerRunWithInjectedAnimator(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := term.NewAnimator(logger,
		term.WithWriter(io.Discard),
		term.WithRefreshRate(10*time.Millisecond),
	)

	// Spinner is registered into a when WithSpinnerAnimator is applied.
	sp := term.Spinner("shared", term.WithSpinnerAnimator(a))

	// Run must return immediately (it is a no-op for injected animators).
	runDone := make(chan error, 1)
	go func() {
		runDone <- sp.Run(context.Background())
	}()

	select {
	case err := <-runDone:
		if err != nil {
			t.Errorf("Spinner.Run with injected animator = %v, want nil", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Spinner.Run with injected animator did not return immediately")
	}

	// Close the spinner and then cancel the animator.
	_ = sp.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = a.Run(ctx)
}

// TestProgressRunSelfContained verifies that Progress.Run terminates when Done is called.
func TestProgressRunSelfContained(t *testing.T) {
	t.Parallel()

	p := term.NewProgress(100, term.WithProgressLabel("test"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- p.Run(ctx)
	}()

	// Give Run time to start and enter its frame loop.
	time.Sleep(50 * time.Millisecond)

	// Calling Done marks the progress complete; the internal animator should
	// detect allDone and exit Run.
	p.Done()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Progress.Run = %v, want nil", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Progress.Run did not return after Done()")
	}
}

// TestProgressRunWithInjectedAnimator verifies that Run is a no-op when an
// Animator is injected via WithProgressAnimator.
func TestProgressRunWithInjectedAnimator(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := term.NewAnimator(logger,
		term.WithWriter(io.Discard),
		term.WithRefreshRate(10*time.Millisecond),
	)

	p := term.NewProgress(100, term.WithProgressAnimator(a))

	runDone := make(chan error, 1)
	go func() {
		runDone <- p.Run(context.Background())
	}()

	select {
	case err := <-runDone:
		if err != nil {
			t.Errorf("Progress.Run with injected animator = %v, want nil", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Progress.Run with injected animator did not return immediately")
	}

	p.Done()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = a.Run(ctx)
}

// TestStatusBarRunSelfContained verifies that StatusBar.Run blocks until ctx cancels.
func TestStatusBarRunSelfContained(t *testing.T) {
	t.Parallel()

	sb := term.NewStatusBar()
	sb.SetSection("phase", "running")

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- sb.Run(ctx)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Error("StatusBar.Run returned nil, want ErrCancelled")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("StatusBar.Run did not return after ctx cancellation")
	}
}

// TestStatusBarRunWithInjectedAnimator verifies that Run is a no-op when an
// Animator is injected via WithStatusBarAnimator.
func TestStatusBarRunWithInjectedAnimator(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := term.NewAnimator(logger,
		term.WithWriter(io.Discard),
		term.WithRefreshRate(10*time.Millisecond),
	)

	sb := term.NewStatusBar(term.WithStatusBarAnimator(a))
	sb.SetSection("x", "y")

	runDone := make(chan error, 1)
	go func() {
		runDone <- sb.Run(context.Background())
	}()

	select {
	case err := <-runDone:
		if err != nil {
			t.Errorf("StatusBar.Run with injected animator = %v, want nil", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("StatusBar.Run with injected animator did not return immediately")
	}

	_ = sb.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = a.Run(ctx)
}

// TestDownloadProgressRunSelfContained verifies that DownloadProgress.Run terminates
// when the underlying reader reaches EOF (which calls Done internally).
func TestDownloadProgressRunSelfContained(t *testing.T) {
	t.Parallel()

	data := bytes.Repeat([]byte("x"), 1024)
	r := bytes.NewReader(data)

	dp := term.NewDownloadProgress(r, int64(len(data)), "download-test")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runDone := make(chan error, 1)
	go func() {
		runDone <- dp.Run(ctx)
	}()

	// Give Run time to start, then drain the reader to trigger EOF -> Done.
	time.Sleep(50 * time.Millisecond)
	_, _ = io.ReadAll(dp)

	select {
	case err := <-runDone:
		if err != nil {
			t.Errorf("DownloadProgress.Run = %v, want nil", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("DownloadProgress.Run did not return after EOF")
	}
}

// TestDownloadProgressRunWithInjectedAnimator verifies that Run is a no-op when
// an Animator is injected via WithProgressAnimator.
func TestDownloadProgressRunWithInjectedAnimator(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := term.NewAnimator(logger,
		term.WithWriter(io.Discard),
		term.WithRefreshRate(10*time.Millisecond),
	)

	data := []byte("hello")
	r := bytes.NewReader(data)
	dp := term.NewDownloadProgress(r, int64(len(data)), "injected", term.WithProgressAnimator(a))

	runDone := make(chan error, 1)
	go func() {
		runDone <- dp.Run(context.Background())
	}()

	select {
	case err := <-runDone:
		if err != nil {
			t.Errorf("DownloadProgress.Run with injected animator = %v, want nil", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("DownloadProgress.Run with injected animator did not return immediately")
	}

	dp.Done()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = a.Run(ctx)
}

// TestPrimitiveCloseIdempotent verifies that Close called twice returns nil for
// all four primitives.
func TestPrimitiveCloseIdempotent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		close func() (error, error)
	}{
		{
			name: "SpinnerAnimator",
			close: func() (error, error) {
				sp := term.Spinner("test")
				return sp.Close(), sp.Close()
			},
		},
		{
			name: "Progress",
			close: func() (error, error) {
				p := term.NewProgress(100)
				return p.Close(), p.Close()
			},
		},
		{
			name: "StatusBar",
			close: func() (error, error) {
				sb := term.NewStatusBar()
				return sb.Close(), sb.Close()
			},
		},
		{
			name: "DownloadProgress",
			close: func() (error, error) {
				dp := term.NewDownloadProgress(bytes.NewReader(nil), 0, "")
				return dp.Close(), dp.Close()
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e1, e2 := tc.close()
			if e1 != nil {
				t.Errorf("%s: first Close() = %v, want nil", tc.name, e1)
			}
			if e2 != nil {
				t.Errorf("%s: second Close() = %v, want nil (idempotent)", tc.name, e2)
			}
		})
	}
}

// TestSpinnerCloseStopsRender verifies that after Close, Render returns done=true.
func TestSpinnerCloseStopsRender(t *testing.T) {
	t.Parallel()

	sp := term.Spinner("test")
	_ = sp.Close()
	_, done := sp.Render()
	if !done {
		t.Error("Spinner.Render() done=false after Close()")
	}
}

// TestProgressCloseStopsRender verifies that after Close, Render returns done=true.
func TestProgressCloseStopsRender(t *testing.T) {
	t.Parallel()

	p := term.NewProgress(100)
	_ = p.Close()
	_, done := p.Render()
	if !done {
		t.Error("Progress.Render() done=false after Close()")
	}
}

// TestSpinnerSatisfiesAnimation verifies *SpinnerAnimator satisfies Animation.
func TestSpinnerSatisfiesAnimation(t *testing.T) {
	t.Parallel()

	var _ term.Animation = term.Spinner("check")

	// Verify it can be added to a shared Animator without type assertions.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := term.NewAnimator(logger,
		term.WithWriter(io.Discard),
		term.WithRefreshRate(10*time.Millisecond),
	)
	sp := term.Spinner("label")
	h := a.Add(sp)
	_ = sp.Close() // mark done so animator exits naturally
	h.Cancel()
}
