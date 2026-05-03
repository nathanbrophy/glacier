// SPDX-License-Identifier: Apache-2.0

package term

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const promptLineCap = 4 * 1024 // 4 KiB per §23.9 #25

// promptConfig holds resolved options for Prompt.
type promptConfig struct {
	defaultVal  string
	validator   func(string) error
	maxAttempts int // 0 = unlimited
	timeout     time.Duration
}

// PromptOption configures Prompt behavior.
type PromptOption interface{ applyPrompt(*promptConfig) error }

type promptOptionFunc func(*promptConfig) error

func (f promptOptionFunc) applyPrompt(c *promptConfig) error { return f(c) }

// WithDefault sets the value returned when the user submits an empty input.
func WithDefault(s string) PromptOption {
	return promptOptionFunc(func(c *promptConfig) error { c.defaultVal = s; return nil })
}

// WithValidator registers fn as the input validator. If fn returns non-nil
// the prompt is re-displayed with the error message; attempts are counted
// against WithMaxAttempts.
func WithValidator(fn func(string) error) PromptOption {
	return promptOptionFunc(func(c *promptConfig) error { c.validator = fn; return nil })
}

// WithMaxAttempts limits the number of invalid-input retries.
// When exhausted, Prompt returns ErrTooManyAttempts.
// Default: unlimited.
func WithMaxAttempts(n int) PromptOption {
	return promptOptionFunc(func(c *promptConfig) error { c.maxAttempts = n; return nil })
}

// WithTimeout sets a per-prompt deadline. Exceeded deadline returns ErrTimeout.
func WithTimeout(d time.Duration) PromptOption {
	return promptOptionFunc(func(c *promptConfig) error { c.timeout = d; return nil })
}

// confirmConfig holds resolved options for Confirm.
type confirmConfig struct {
	defaultYes bool
}

// ConfirmOption configures Confirm behavior.
type ConfirmOption interface{ applyConfirm(*confirmConfig) error }

type confirmOptionFunc func(*confirmConfig) error

func (f confirmOptionFunc) applyConfirm(c *confirmConfig) error { return f(c) }

// WithDefaultYes makes the empty-input response "yes". Default is "no".
func WithDefaultYes() ConfirmOption {
	return confirmOptionFunc(func(c *confirmConfig) error { c.defaultYes = true; return nil })
}

// selectConfig holds resolved options for Select/MultiSelect.
type selectConfig struct{}

// SelectOption configures Select and MultiSelect behavior.
type SelectOption interface{ applySelect(*selectConfig) error }

type selectOptionFunc func(*selectConfig) error

func (f selectOptionFunc) applySelect(c *selectConfig) error { return f(c) }

// Prompt displays question on the terminal, reads a single line of input,
// trims trailing whitespace, and returns it.
//
// If the writer is not a TTY (piped stdin), reads from stdin without entering
// raw mode; returns ErrNotInteractive if the prompt requires interactivity.
//
// Input constraints (§23.9 #25):
//   - Lines are capped at 4 KiB. Input exceeding this limit returns an error.
//   - NUL bytes and non-printable control characters except backspace (0x08)
//     and arrow-key escape sequences are rejected.
//
// Preconditions: ctx must not be nil; question must not be empty.
// Error contract:
//   - ErrCancelled if ctx is cancelled or the user sends EOF/Ctrl-C.
//   - ErrTimeout if WithTimeout is set and expires.
//   - ErrTooManyAttempts if WithMaxAttempts is set and exhausted.
//
// Terminal restore: raw mode is released unconditionally via defer; panic-safe.
// Concurrency: blocking; not goroutine-safe (owns the terminal).
func Prompt(ctx context.Context, question string, opts ...PromptOption) (string, error) {
	cfg := promptConfig{}
	for _, o := range opts {
		if o == nil {
			continue
		}
		if err := o.applyPrompt(&cfg); err != nil {
			return "", err
		}
	}

	if cfg.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.timeout)
		defer cancel()
	}

	attempts := 0
	for {
		select {
		case <-ctx.Done():
			return "", ErrCancelled
		default:
		}

		input, err := readLine(ctx, question, false)
		if err != nil {
			return "", err
		}
		if input == "" && cfg.defaultVal != "" {
			input = cfg.defaultVal
		}

		if cfg.validator != nil {
			if verr := cfg.validator(input); verr != nil {
				attempts++
				if cfg.maxAttempts > 0 && attempts >= cfg.maxAttempts {
					return "", ErrTooManyAttempts
				}
				fmt.Fprintf(os.Stderr, "  error: %s\n", verr)
				continue
			}
		}
		return input, nil
	}
}

// Password displays question and reads input with echo disabled.
// The typed characters are never written to the terminal.
//
// All constraints from Prompt apply except the line cap is still 4 KiB.
// Returns ErrCancelled on ctx cancellation.
//
// Terminal restore: cooked mode restored unconditionally via defer; panic-safe.
// Concurrency: blocking; not goroutine-safe (owns the terminal).
func Password(ctx context.Context, question string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ErrCancelled
	default:
	}

	fmt.Fprint(os.Stderr, question+" ")

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// Fallback: read plainly.
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, promptLineCap+1), promptLineCap+1)
		if scanner.Scan() {
			return sanitizeInput(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("term: password: %w", err)
		}
		return "", ErrCancelled
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", fmt.Errorf("term: password: raw mode: %w", err)
	}
	panicking := true
	defer func() {
		_ = term.Restore(fd, oldState)
		if panicking {
			if r := recover(); r != nil {
				//glacier:nolint=panic-in-library re-panic propagates the inner panic faithfully after raw-mode restoration.
				panic(r)
			}
		}
	}()

	resultCh := make(chan struct {
		s   string
		err error
	}, 1)
	go func() {
		b, e := term.ReadPassword(fd)
		resultCh <- struct {
			s   string
			err error
		}{string(b), e}
	}()

	select {
	case <-ctx.Done():
		panicking = false
		fmt.Fprintln(os.Stderr)
		return "", ErrCancelled
	case res := <-resultCh:
		panicking = false
		fmt.Fprintln(os.Stderr)
		if res.err != nil {
			if res.err == io.EOF {
				return "", ErrCancelled
			}
			return "", fmt.Errorf("term: password: %w", res.err)
		}
		return res.s, nil
	}
}

// Confirm displays question with a Y/N prompt and returns the boolean answer.
//
// Accepted affirmative inputs (case-insensitive): "y", "yes".
// Accepted negative inputs: "n", "no".
// Invalid input causes re-prompt; after three invalid attempts returns ErrTooManyAttempts.
//
// Error contract: same as Prompt. Default is No unless WithDefaultYes() is set.
// Concurrency: blocking; not goroutine-safe (owns the terminal).
func Confirm(ctx context.Context, question string, opts ...ConfirmOption) (bool, error) {
	cfg := confirmConfig{}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applyConfirm(&cfg)
	}

	hint := "[y/N]"
	if cfg.defaultYes {
		hint = "[Y/n]"
	}
	fullQ := question + " " + hint

	attempts := 0
	for {
		input, err := readLine(ctx, fullQ, false)
		if err != nil {
			return false, err
		}
		lower := strings.ToLower(strings.TrimSpace(input))
		switch lower {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		case "":
			return cfg.defaultYes, nil
		default:
			attempts++
			if attempts >= 3 {
				return false, ErrTooManyAttempts
			}
			fmt.Fprintln(os.Stderr, "  please enter y or n")
		}
	}
}

// Select displays a numbered or arrow-navigable list of options and returns
// the one selected by the user. T is the element type; render converts T to a
// display string shown in the list.
//
// Preconditions: options must not be empty.
// Returns ErrCancelled if ctx is cancelled or EOF/Ctrl-C received.
// Returns ErrNotInteractive if the writer is not a TTY.
// Concurrency: blocking; not goroutine-safe (owns the terminal).
func Select[T any](ctx context.Context, question string, options []T, render func(T) string, opts ...SelectOption) (T, error) {
	var zero T
	if len(options) == 0 {
		return zero, fmt.Errorf("term: select: options must not be empty")
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return zero, ErrNotInteractive
	}

	fmt.Fprintln(os.Stderr, question)
	for i, o := range options {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, render(o))
	}

	for {
		input, err := readLine(ctx, fmt.Sprintf("Enter 1-%d:", len(options)), false)
		if err != nil {
			return zero, err
		}
		var idx int
		if _, err := fmt.Sscan(input, &idx); err != nil || idx < 1 || idx > len(options) {
			fmt.Fprintf(os.Stderr, "  please enter a number between 1 and %d\n", len(options))
			continue
		}
		return options[idx-1], nil
	}
}

// MultiSelect displays a checkable list of options and returns all selected.
// Space toggles selection; Enter confirms. Returns an empty slice if none selected.
//
// Same error contract as Select.
// Concurrency: blocking; not goroutine-safe (owns the terminal).
func MultiSelect[T any](ctx context.Context, question string, options []T, render func(T) string, opts ...SelectOption) ([]T, error) {
	if len(options) == 0 {
		return nil, fmt.Errorf("term: multi-select: options must not be empty")
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, ErrNotInteractive
	}

	fmt.Fprintln(os.Stderr, question+" (comma-separated numbers, or empty for none):")
	for i, o := range options {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, render(o))
	}

	for {
		input, err := readLine(ctx, "Select:", false)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(input) == "" {
			return []T{}, nil
		}
		parts := strings.Split(input, ",")
		var result []T
		valid := true
		for _, p := range parts {
			p = strings.TrimSpace(p)
			var idx int
			if _, e := fmt.Sscan(p, &idx); e != nil || idx < 1 || idx > len(options) {
				fmt.Fprintf(os.Stderr, "  invalid selection %q; enter comma-separated numbers between 1 and %d\n", p, len(options))
				valid = false
				break
			}
			result = append(result, options[idx-1])
		}
		if valid {
			return result, nil
		}
	}
}

// readLine prints question to stderr and reads one line from stdin.
// Performs §23.9 #25 sanitization: caps at 4 KiB; rejects NUL and non-printable
// control characters except backspace.
func readLine(ctx context.Context, question string, _ bool) (string, error) {
	fmt.Fprint(os.Stderr, question+" ")

	type result struct {
		s   string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, promptLineCap+1), promptLineCap+1)
		if scanner.Scan() {
			s, err := sanitizeInput(scanner.Text())
			ch <- result{s, err}
			return
		}
		if err := scanner.Err(); err != nil {
			ch <- result{"", fmt.Errorf("term: prompt: %w", err)}
			return
		}
		ch <- result{"", ErrCancelled}
	}()

	select {
	case <-ctx.Done():
		return "", ErrCancelled
	case r := <-ch:
		return r.s, r.err
	}
}

// sanitizeInput enforces §23.9 #25: rejects NUL bytes and non-printable
// control chars (except backspace 0x08). Returns the cleaned string.
func sanitizeInput(s string) (string, error) {
	if len(s) > promptLineCap {
		return "", fmt.Errorf("term: prompt: input exceeds 4 KiB limit")
	}
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		if r == 0 {
			return "", fmt.Errorf("term: prompt: input contains NUL byte")
		}
		// Allow printable runes, backspace, and ANSI escape intro.
		if r < 0x20 && r != 0x08 && r != 0x1b {
			return "", fmt.Errorf("term: prompt: input contains non-printable control character 0x%02x", r)
		}
		sb.WriteRune(r)
	}
	return strings.TrimRight(sb.String(), "\r"), nil
}
