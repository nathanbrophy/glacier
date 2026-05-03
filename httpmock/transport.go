// SPDX-License-Identifier: Apache-2.0

package httpmock

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/errs"
	"github.com/nathanbrophy/glacier/option"
)

// transportConfig holds construction-time settings for Transport.
type transportConfig struct {
	// defaultStatus == 0 → strict mode (unmatched → ErrNoRouteMatch).
	defaultStatus int
	// logger is never nil; defaults to slog.Default().
	logger *slog.Logger
}

// Transport implements http.RoundTripper. All mutable state is protected by mu.
type Transport struct {
	mu       sync.RWMutex
	stubs    []*Stub
	recorded []*http.Request
	closed   bool
	cfg      transportConfig
}

// New constructs a fresh Transport with no registered stubs.
// Strict mode is the default: unmatched requests return ErrNoRouteMatch.
func New(opts ...option.Option[transportConfig]) *Transport {
	cfg, err := option.Apply(opts)
	if err != nil {
		//glacier:nolint=panic-in-library programmer error: option misuse surfaces at construction.
		panic("httpmock: New: " + err.Error())
	}
	if cfg.logger == nil {
		cfg.logger = slog.Default()
	}
	return &Transport{cfg: cfg}
}

// NewWithT constructs a Transport and registers Transport.Verify at t.Cleanup.
func NewWithT(t assert.TB, opts ...option.Option[transportConfig]) *Transport {
	rt := New(opts...)
	t.Cleanup(func() { rt.Verify(t) })
	return rt
}

// RoundTrip implements http.RoundTripper. Match, increment, and respond happen
// in a single write-lock critical section (§23.14).
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil && req.Body != http.NoBody {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, errs.Wrap(err, "httpmock: roundtrip: read body")
		}
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.recorded = append(t.recorded, req)

	if t.closed {
		return nil, ErrNoRouteMatch
	}

	contentType := ""
	if req.Header != nil {
		contentType = req.Header.Get("Content-Type")
	}

	for i, s := range t.stubs {
		if !s.matches(req, bodyBytes, contentType) {
			continue
		}
		s.hitCount++
		if s.responder == nil {
			return nil, &ScriptError{Step: i, Cause: errors.New("stub missing responder")}
		}
		t.cfg.logger.Debug("httpmock: stub matched",
			"method", req.Method,
			"path", req.URL.Path,
			"stub_index", i,
		)
		return s.responder.Respond(req)
	}

	if t.cfg.defaultStatus == 0 {
		return nil, ErrNoRouteMatch
	}
	return &http.Response{
		StatusCode: t.cfg.defaultStatus,
		Status:     http.StatusText(t.cfg.defaultStatus),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(nil)),
		Request:    req,
	}, nil
}

// OnRequest returns a new Stub builder. The stub is registered immediately;
// Respond finalizes its configuration.
func (t *Transport) OnRequest() *Stub {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := &Stub{transport: t, timesMin: -1, timesMax: -1}
	t.stubs = append(t.stubs, s)
	return s
}

// RequestsTo returns all recorded requests whose path matches pattern.
// Glob matching is used when pattern contains '*'.
func (t *Transport) RequestsTo(pattern string) []*http.Request {
	t.mu.RLock()
	defer t.mu.RUnlock()
	isGlob := containsWildcard(pattern)
	var out []*http.Request
	for _, r := range t.recorded {
		path := r.URL.Path
		if isGlob {
			if matched, _ := filepath.Match(pattern, path); matched {
				out = append(out, r)
			}
		} else if path == pattern {
			out = append(out, r)
		}
	}
	return out
}

// AllRequests returns every recorded request in arrival order (copy).
func (t *Transport) AllRequests() []*http.Request {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]*http.Request, len(t.recorded))
	copy(out, t.recorded)
	return out
}

// Verify checks Times/AtLeast/AtMost/Never expectations on all stubs.
func (t *Transport) Verify(tb assert.TB) {
	tb.Helper()
	t.mu.RLock()
	defer t.mu.RUnlock()
	for i, s := range t.stubs {
		if s.timesMin == -1 && s.timesMax == -1 {
			continue
		}
		hit := s.hitCount
		if s.timesMin >= 0 && hit < s.timesMin {
			tb.Errorf("httpmock: stub[%d]: expected at least %d call(s), got %d", i, s.timesMin, hit)
		}
		if s.timesMax >= 0 && hit > s.timesMax {
			tb.Errorf("httpmock: stub[%d]: expected at most %d call(s), got %d", i, s.timesMax, hit)
		}
	}
}

// Close marks the transport closed. Subsequent RoundTrip calls return
// ErrNoRouteMatch. Idempotent.
func (t *Transport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}

func containsWildcard(s string) bool {
	for i := range len(s) {
		if s[i] == '*' {
			return true
		}
	}
	return false
}

// StrictDefault configures the transport to return ErrNoRouteMatch for unmatched
// requests. This is the default; the option makes strict intent explicit.
func StrictDefault() option.Option[transportConfig] {
	return option.OptionFunc[transportConfig](func(c *transportConfig) error {
		c.defaultStatus = 0
		return nil
	})
}

// LenientMode configures the transport to return an empty 404 for unmatched requests.
func LenientMode() option.Option[transportConfig] {
	return option.OptionFunc[transportConfig](func(c *transportConfig) error {
		c.defaultStatus = http.StatusNotFound
		return nil
	})
}

// WithDefaultStatus configures the transport to return an empty response with
// the given status for unmatched requests. Panics if status ∉ [100,599].
func WithDefaultStatus(status int) option.Option[transportConfig] {
	if status < 100 || status > 599 {
		//glacier:nolint=panic-in-library programmer error: out-of-range status is documented as a panic precondition.
		panic("httpmock: WithDefaultStatus: status must be in [100, 599]")
	}
	return option.OptionFunc[transportConfig](func(c *transportConfig) error {
		c.defaultStatus = status
		return nil
	})
}

// WithLogger injects the slog.Logger for stub-match trace events.
// Panics if l is nil.
func WithLogger(l *slog.Logger) option.Option[transportConfig] {
	if l == nil {
		//glacier:nolint=panic-in-library programmer error: nil logger is documented as a panic precondition.
		panic("httpmock: WithLogger: logger must not be nil")
	}
	return option.OptionFunc[transportConfig](func(c *transportConfig) error {
		c.logger = l
		return nil
	})
}
