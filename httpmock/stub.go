// SPDX-License-Identifier: Apache-2.0

package httpmock

import (
	"net/http"
	"regexp"
	"strings"
)

// Stub is the chained builder for a single registered request expectation.
type Stub struct {
	transport   *Transport
	method      string // upper-cased HTTP verb; "" matches any
	pathExact   string
	pathPrefix  string
	pathRegex   *regexp.Regexp
	useRegex    bool // Regex() was called; pathExact holds raw pattern
	queryParams map[string]string
	headers     map[string]string
	bodyMatcher BodyMatcher
	timesMin    int // -1 = no lower bound
	timesMax    int // -1 = no upper bound
	hitCount    int // protected by Transport.mu
	responder   Responder
}

// Method restricts the stub to requests with the given HTTP method (case-insensitive).
func (s *Stub) Method(m string) *Stub {
	s.method = strings.ToUpper(m)
	return s
}

// Path restricts the stub to requests whose URL path equals p exactly.
func (s *Stub) Path(p string) *Stub {
	s.pathExact = p
	s.pathPrefix = ""
	s.useRegex = false
	s.pathRegex = nil
	return s
}

// PathPrefix restricts the stub to requests whose URL path starts with p.
// Mutually exclusive with Regex (panics at Respond time if both are set).
func (s *Stub) PathPrefix(p string) *Stub {
	s.pathPrefix = p
	s.pathExact = ""
	s.useRegex = false
	s.pathRegex = nil
	return s
}

// Regex reinterprets the most recently set Path value as a regular expression.
// Must be called after Path. Mutually exclusive with PathPrefix.
func (s *Stub) Regex() *Stub {
	s.useRegex = true
	return s
}

// Query ANDs a query-parameter constraint onto the stub.
func (s *Stub) Query(name, value string) *Stub {
	if s.queryParams == nil {
		s.queryParams = make(map[string]string)
	}
	s.queryParams[name] = value
	return s
}

// Header ANDs a header constraint onto the stub. Name is canonicalized.
func (s *Stub) Header(name, value string) *Stub {
	if s.headers == nil {
		s.headers = make(map[string]string)
	}
	s.headers[http.CanonicalHeaderKey(name)] = value
	return s
}

// Body restricts the stub to requests satisfying the given BodyMatcher.
func (s *Stub) Body(matcher BodyMatcher) *Stub {
	s.bodyMatcher = matcher
	return s
}

// Times sets the exact expected call count. Panics if n <= 0.
func (s *Stub) Times(n int) *Stub {
	if n <= 0 {
		panic("httpmock: Times: n must be > 0")
	}
	s.timesMin = n
	s.timesMax = n
	return s
}

// AtLeast sets a minimum expected call count.
func (s *Stub) AtLeast(n int) *Stub {
	s.timesMin = n
	s.timesMax = -1
	return s
}

// AtMost sets a maximum expected call count.
func (s *Stub) AtMost(n int) *Stub {
	s.timesMin = -1
	s.timesMax = n
	return s
}

// AnyTimes removes all call-count expectations.
func (s *Stub) AnyTimes() *Stub {
	s.timesMin = -1
	s.timesMax = -1
	return s
}

// Never asserts the stub must not be matched.
func (s *Stub) Never() *Stub {
	s.timesMin = 0
	s.timesMax = 0
	return s
}

// Respond sets the Responder. Calling twice overrides the first (last wins).
// Panics if PathPrefix and Regex are both set, or if the regex fails to compile.
func (s *Stub) Respond(r Responder) *Stub {
	if s.pathPrefix != "" && s.useRegex {
		panic("httpmock: stub: PathPrefix and Regex are mutually exclusive")
	}
	if s.useRegex && s.pathRegex == nil {
		re, err := regexp.Compile(s.pathExact)
		if err != nil {
			panic("httpmock: stub: regex does not compile: " + err.Error())
		}
		s.pathRegex = re
	}
	s.responder = r
	return s
}

// matches reports whether all predicates match req.
func (s *Stub) matches(req *http.Request, body []byte, contentType string) bool {
	if s.method != "" && req.Method != s.method {
		return false
	}
	path := req.URL.Path
	if s.pathRegex != nil {
		if !s.pathRegex.MatchString(path) {
			return false
		}
	} else if s.pathPrefix != "" {
		if !strings.HasPrefix(path, s.pathPrefix) {
			return false
		}
	} else if s.pathExact != "" {
		if path != s.pathExact {
			return false
		}
	}
	if len(s.queryParams) > 0 {
		q := req.URL.Query()
		for name, want := range s.queryParams {
			if q.Get(name) != want {
				return false
			}
		}
	}
	if len(s.headers) > 0 {
		for name, want := range s.headers {
			if req.Header.Get(name) != want {
				return false
			}
		}
	}
	if s.bodyMatcher != nil {
		if !s.bodyMatcher.Match(body, contentType) {
			return false
		}
	}
	return true
}
