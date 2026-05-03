// SPDX-License-Identifier: Apache-2.0

// Package ghreleases fetches release metadata from the GitHub Releases API
// using the Glacier httpc package. It does not implement caching; callers
// are responsible for cache arbitration (e.g. the version command).
package ghreleases

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/nathanbrophy/glacier/errs"
	"github.com/nathanbrophy/glacier/httpc"
	"github.com/nathanbrophy/glacier/option"
)

// Release is the decoded response from the GitHub /releases/latest endpoint.
type Release struct {
	// TagName is the git tag (e.g. "v1.2.3").
	TagName string `json:"tag_name"`
	// Name is the human-readable release title.
	Name string `json:"name"`
	// HTMLURL is the browser URL for the release page.
	HTMLURL string `json:"html_url"`
	// PublishedAt is when the release was published.
	PublishedAt time.Time `json:"published_at"`
}

// ReleaseFetcher fetches the latest release for a GitHub repository.
//
// +glacier:mock
type ReleaseFetcher interface {
	// Latest returns the most recent published release for the given
	// "owner/repo" string. Returns an error on network failure, HTTP error,
	// JSON decode failure, or tag/URL validation failure.
	Latest(ctx context.Context, repo string) (Release, error)
}

// tagRe validates tag names before display. Accepts vMAJOR.MINOR.PATCH with
// an optional pre-release suffix (e.g. v1.2.3-rc.1).
var tagRe = regexp.MustCompile(`^v\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?$`)

// urlPrefix is the required prefix for html_url validation.
const urlPrefix = "https://github.com/"

// releaseAPIURL returns the GitHub API URL for the latest release of repo.
func releaseAPIURL(repo string) string {
	return "https://api.github.com/repos/" + repo + "/releases/latest"
}

// clientConfig holds construction-time options for the default client.
type clientConfig struct {
	httpClient *httpc.Client
}

// Option is a functional option for NewClient. It dogfoods the framework's
// option.Option[T] pattern so callers can compose multiple options uniformly
// across Glacier packages.
type Option = option.Option[clientConfig]

// WithHTTPClient replaces the httpc.Client used for GitHub API requests.
// The primary use-case is test injection via httpmock.Transport.
func WithHTTPClient(c *httpc.Client) Option {
	return option.OptionFunc[clientConfig](func(cfg *clientConfig) error {
		cfg.httpClient = c
		return nil
	})
}

// defaultClient is the concrete implementation of ReleaseFetcher.
type defaultClient struct {
	hc *httpc.Client
}

// NewClient returns a ReleaseFetcher backed by the given options.
// When no WithHTTPClient option is provided, httpc.Default is used.
func NewClient(opts ...Option) ReleaseFetcher {
	cfg, err := option.Apply(opts)
	if err != nil {
		// ghreleases options never return errors at v0; this branch is unreachable.
		panic("ghreleases: option.Apply returned unexpected error: " + err.Error())
	}
	if cfg.httpClient == nil {
		cfg.httpClient = httpc.Default
	}
	return &defaultClient{hc: cfg.httpClient}
}

// Latest fetches the latest release for repo (format "owner/repo") from the
// GitHub Releases API. Returns an error on non-200 status, JSON decode failure,
// or tag/URL validation failure.
//
// Does NOT implement caching; that responsibility belongs to the caller.
func (c *defaultClient) Latest(ctx context.Context, repo string) (Release, error) {
	if !strings.Contains(repo, "/") {
		return Release{}, fmt.Errorf("ghreleases: invalid repo %q: must be owner/repo format", repo)
	}
	url := releaseAPIURL(repo)
	rel, _, err := httpc.GetWith[Release](c.hc, ctx, url)
	if err != nil {
		// Preserve HTTP 403 rate-limit so the caller can distinguish it from
		// a generic network error.
		var statusErr *httpc.StatusError
		if errors.As(err, &statusErr) && statusErr.Status == http.StatusForbidden {
			return Release{}, &RateLimitError{Cause: err}
		}
		return Release{}, errs.Wrap(err, "ghreleases: fetch")
	}
	if !tagRe.MatchString(rel.TagName) {
		return Release{}, fmt.Errorf("ghreleases: unexpected tag_name format: %q", rel.TagName)
	}
	if !strings.HasPrefix(rel.HTMLURL, urlPrefix) {
		return Release{}, fmt.Errorf("ghreleases: unexpected html_url prefix: %q", rel.HTMLURL)
	}
	return rel, nil
}

// RateLimitError is returned by Latest when the GitHub API responds with HTTP 403.
// Callers use errors.As to distinguish rate-limit from generic network failure.
type RateLimitError struct {
	Cause error
}

// Error implements error.
func (e *RateLimitError) Error() string {
	return "ghreleases: rate limited (HTTP 403): " + e.Cause.Error()
}

// Unwrap supports errors.Is / errors.As traversal.
func (e *RateLimitError) Unwrap() error { return e.Cause }

// Latest is the package-level convenience function using the default httpc.Client.
// Kept for backward compatibility; prefer NewClient for testable code.
func Latest(ctx context.Context, repo string) (Release, error) {
	return NewClient().Latest(ctx, repo)
}
