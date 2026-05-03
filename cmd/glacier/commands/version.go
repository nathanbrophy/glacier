// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/nathanbrophy/glacier/cache"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/ghreleases"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/report"
	"github.com/nathanbrophy/glacier/httpc"
)

// VersionCmd prints version information.
//
// +glacier:command name=version parent=glacier
type VersionCmd struct {
	// Check fetches the latest release from GitHub and compares with the running version.
	//
	// +glacier:default false
	Check bool

	// Strict causes --check to exit non-zero when the GitHub endpoint is unreachable.
	//
	// +glacier:default false
	Strict bool

	// JSON emits version information as JSON to stdout.
	//
	// +glacier:default false
	JSON bool

	// fetcher is the release-fetching dependency. nil means use the default.
	// Populated by tests via withFetcher.
	fetcher ghreleases.ReleaseFetcher

	// stdout is the writer for human-readable and JSON output.
	// nil means os.Stdout; injected by tests to avoid global os.Stdout mutation.
	stdout io.Writer

	// releaseCache is the layered (mem -> disk) cache for release lookups.
	// nil means use the default constructed at <UserCacheDir>/glacier/.
	// Tests inject an in-memory-only cache for determinism.
	releaseCache cache.Cache[ghreleases.Release]
}

// withFetcher returns a copy of c with the given fetcher injected.
// Used by tests to swap in a mock or httpmock-backed client without network calls.
func (c *VersionCmd) withFetcher(f ghreleases.ReleaseFetcher) *VersionCmd {
	cp := *c
	cp.fetcher = f
	return &cp
}

// withWriter returns a copy of c with the given writer injected.
func (c *VersionCmd) withWriter(w io.Writer) *VersionCmd {
	cp := *c
	cp.stdout = w
	return &cp
}

// withCache returns a copy of c with the given release cache injected.
// Used by tests to use a deterministic in-memory cache without disk I/O.
func (c *VersionCmd) withCache(rc cache.Cache[ghreleases.Release]) *VersionCmd {
	cp := *c
	cp.releaseCache = rc
	return &cp
}

// resolveCache returns the configured cache, or constructs the default
// layered cache (in-memory primary backed by a per-user disk cache).
// The default TTL comes from sdkCfg().VersionCheck.CacheTTL (24h per D-S21).
func (c *VersionCmd) resolveCache() cache.Cache[ghreleases.Release] {
	if c.releaseCache != nil {
		return c.releaseCache
	}
	ttl := sdkCfg().VersionCheck.CacheTTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	mem := cache.New[ghreleases.Release](cache.WithDefaultTTL(ttl))

	// Best-effort disk cache: if UserCacheDir or NewDisk fails, fall back to
	// memory-only. The cache is supposed to degrade silently per spec 0033.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return mem
	}
	disk, err := cache.NewDisk[ghreleases.Release](filepath.Join(cacheDir, "glacier"), cache.WithDefaultTTL(ttl))
	if err != nil {
		return mem
	}
	return cache.NewLayered(mem, disk)
}

// out returns the effective output writer.
func (c *VersionCmd) out() io.Writer {
	if c.stdout != nil {
		return c.stdout
	}
	return os.Stdout
}

// releaseFetcher returns the injected fetcher, or the default client.
func (c *VersionCmd) releaseFetcher() ghreleases.ReleaseFetcher {
	if c.fetcher != nil {
		return c.fetcher
	}
	return ghreleases.NewClient(ghreleases.WithHTTPClient(httpc.Default))
}

// versionOutput holds the data printed by version (D-S63 JSON schema).
type versionOutput struct {
	Version   string        `json:"version"`
	BuildTime string        `json:"build_time,omitempty"`
	GoVersion string        `json:"go_version"`
	OS        string        `json:"os"`
	Arch      string        `json:"arch"`
	Latest    *latestOutput `json:"latest,omitempty"`
}

// latestOutput is the optional "latest" object in the JSON schema (D-S63).
type latestOutput struct {
	Tag         string `json:"tag"`
	PublishedAt string `json:"published_at,omitempty"`
	HTMLURL     string `json:"html_url,omitempty"`
	Stale       bool   `json:"stale"`
}

// Run prints version information to stdout. With --check it also fetches the
// latest release from GitHub. With --json it emits machine-readable output.
func (c *VersionCmd) Run(ctx context.Context) error {
	report.Status(report.Calm, "glacier version")

	version := resolveVersion()

	info := versionOutput{
		Version:   version,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	// Populate build time from build info.
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.committer.date" || s.Key == "vcs.time" {
				info.BuildTime = s.Value
				break
			}
		}
	}

	if c.Check && sdkCfg().VersionCheck.Enabled {
		rel, err := c.checkLatest(ctx)
		if err != nil {
			var rateLimitErr *ghreleases.RateLimitError
			// 403 rate-limit follows the same graceful-degradation path as offline per spec D-S62.
			_ = errors.As(err, &rateLimitErr)

			if c.Strict || sdkCfg().VersionCheck.Strict {
				report.Status(report.Err, "version check failed: "+err.Error())
				return &exitCodeError{code: exitVersionCheck, cause: err}
			}
			report.Status(report.Alarmed, "latest: unknown (offline)")
		} else {
			info.Latest = &latestOutput{
				Tag:         rel.TagName,
				PublishedAt: rel.PublishedAt.UTC().Format("2006-01-02T15:04:05Z"),
				HTMLURL:     rel.HTMLURL,
				Stale:       false,
			}
		}
	}

	w := c.out()
	if c.JSON {
		return printVersionJSON(w, info)
	}
	printVersionText(w, info)
	return nil
}

// checkLatest performs the version check, consulting the cache first.
//
// Cache strategy (D-S22): a layered Cache[ghreleases.Release] with an
// in-memory primary and a per-user disk backing under <UserCacheDir>/glacier/.
// Default TTL is sdkCfg().VersionCheck.CacheTTL (24h). Concurrent misses on
// the same key collapse onto a single GitHub fetch via the cache's built-in
// singleflight.
func (c *VersionCmd) checkLatest(ctx context.Context) (ghreleases.Release, error) {
	repo := sdkCfg().GitHub.Repo
	rc := c.resolveCache()
	return rc.GetOrLoad(ctx, "github:"+repo, func(ctx context.Context) (ghreleases.Release, error) {
		return c.releaseFetcher().Latest(ctx, repo)
	})
}

// printVersionText writes human-readable version information to w (D-S60).
func printVersionText(w io.Writer, out versionOutput) {
	report.Status(report.Confident, "glacier "+out.Version)
	if out.BuildTime != "" {
		fmt.Fprintf(w, "  build: %s\n", out.BuildTime)
	}
	fmt.Fprintf(w, "  go:    %s\n", out.GoVersion)
	fmt.Fprintf(w, "  os:    %s/%s\n", out.OS, out.Arch)
	if out.Latest != nil {
		fmt.Fprintf(w, "ʕ⌐■-■ʔ latest: %s (released %s)\n", out.Latest.Tag, out.Latest.PublishedAt)
		fmt.Fprintf(w, "  upgrade: go install github.com/nathanbrophy/glacier/cmd/glacier@latest\n")
	}
}

// printVersionJSON writes JSON version information to w (D-S63).
func printVersionJSON(w io.Writer, out versionOutput) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// resolveVersion returns the binary version. It prefers the ldflags-injected
// Version variable from cmd/glacier/version.go; falls back to the module pseudo-version.
func resolveVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			return bi.Main.Version
		}
	}
	return Version
}

// exitCodeError carries a specific exit code for errors that map to non-standard codes.
// It implements cli.ExitCoder so cli.App.Main propagates the embedded code to os.Exit.
type exitCodeError struct {
	code  int
	cause error
}

// Error implements error.
func (e *exitCodeError) Error() string { return e.cause.Error() }

// Unwrap returns the wrapped cause for errors.Is/errors.As traversal.
func (e *exitCodeError) Unwrap() error { return e.cause }

// ExitCode returns the explicit exit code carried by this error.
func (e *exitCodeError) ExitCode() int { return e.code }
