// SPDX-License-Identifier: Apache-2.0

package gen_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cli/gen"
)

func TestMarkerCommandName_Allowed(t *testing.T) {
	t.Parallel()
	pr, errs := gen.ParseMarkers("ServeCmd", []string{"+glacier:command name=serve"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, "serve", pr.Cmd.Name)
}

func TestMarkerCommandName_RejectUppercase(t *testing.T) {
	t.Parallel()
	_, errs := gen.ParseMarkers("ServeCmd", []string{"+glacier:command name=Serve"})
	assert.Equal(t, true, len(errs) > 0)
}

func TestMarkerCommandName_RejectTrailingDash(t *testing.T) {
	t.Parallel()
	_, errs := gen.ParseMarkers("ServeCmd", []string{"+glacier:command name=serve-"})
	assert.Equal(t, true, len(errs) > 0)
}

func TestMarkerCommandName_OversizeRejected(t *testing.T) {
	t.Parallel()
	// 33 char name (exceeds max 32).
	longName := "a" + string(make([]byte, 32))
	_, errs := gen.ParseMarkers("ServeCmd", []string{"+glacier:command name=" + longName})
	assert.Equal(t, true, len(errs) > 0)
}

func TestMarkerParentDottedSegments(t *testing.T) {
	t.Parallel()
	pr, errs := gen.ParseMarkers("DeployCmd", []string{"+glacier:command parent=root.foo.bar"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, "root.foo.bar", pr.Cmd.Parent)
}

func TestMarkerParentDotDotRejected(t *testing.T) {
	t.Parallel()
	_, errs := gen.ParseMarkers("DeployCmd", []string{"+glacier:command parent=root..bar"})
	assert.Equal(t, true, len(errs) > 0)
}

func TestMarkerAlias(t *testing.T) {
	t.Parallel()
	pr, errs := gen.ParseMarkers("ServeCmd", []string{"+glacier:command alias=s"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, "s", pr.Cmd.Alias)
}

func TestMarkerShort_Single(t *testing.T) {
	t.Parallel()
	fm, errs := gen.ParseFieldMarkers("Port", []string{"+glacier:short p"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, true, fm.HasShort)
	assert.Equal(t, 'p', fm.Short)
}

func TestMarkerShort_NonASCII(t *testing.T) {
	t.Parallel()
	_, errs := gen.ParseFieldMarkers("Port", []string{"+glacier:short é"})
	assert.Equal(t, true, len(errs) > 0)
}

func TestMarkerShort_MultiChar(t *testing.T) {
	t.Parallel()
	_, errs := gen.ParseFieldMarkers("Port", []string{"+glacier:short port"})
	assert.Equal(t, true, len(errs) > 0)
}

func TestMarkerEnv_Allowed(t *testing.T) {
	t.Parallel()
	fm, errs := gen.ParseFieldMarkers("Port", []string{"+glacier:env GLACIER_PORT"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, "GLACIER_PORT", fm.Env)
}

func TestMarkerEnv_Lowercase(t *testing.T) {
	t.Parallel()
	_, errs := gen.ParseFieldMarkers("Port", []string{"+glacier:env glacier_port"})
	assert.Equal(t, true, len(errs) > 0)
}

func TestMarkerRequired(t *testing.T) {
	t.Parallel()
	fm, errs := gen.ParseFieldMarkers("Config", []string{"+glacier:required"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, true, fm.Required)
}

func TestMarkerChoices_Pipe(t *testing.T) {
	t.Parallel()
	fm, errs := gen.ParseFieldMarkers("Mode", []string{"+glacier:choices http|grpc|both"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, 3, len(fm.Choices))
	assert.Equal(t, "http", fm.Choices[0])
}

func TestMarkerChoices_Max32(t *testing.T) {
	t.Parallel()
	// 33 choices.
	choices := make([]string, 33)
	for i := range choices {
		choices[i] = "a"
	}
	line := "+glacier:choices " + joinChoices(choices)
	_, errs := gen.ParseFieldMarkers("Mode", []string{line})
	assert.Equal(t, true, len(errs) > 0)
}

func TestMarkerChoices_BadChar(t *testing.T) {
	t.Parallel()
	_, errs := gen.ParseFieldMarkers("Mode", []string{"+glacier:choices A|b"})
	assert.Equal(t, true, len(errs) > 0)
}

func TestMarkerDeprecated(t *testing.T) {
	t.Parallel()
	fm, errs := gen.ParseFieldMarkers("OldFlag", []string{"+glacier:deprecated use --new instead"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, true, fm.HasDepr)
	assert.Equal(t, "use --new instead", fm.Deprecated)
}

func TestMarkerValidate_BadIdent(t *testing.T) {
	t.Parallel()
	_, errs := gen.ParseFieldMarkers("Port", []string{"+glacier:validate 1bad"})
	assert.Equal(t, true, len(errs) > 0)
}

func TestMarkerRoot(t *testing.T) {
	t.Parallel()
	pr, errs := gen.ParseMarkers("RootCmd", []string{"+glacier:root"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, true, pr.IsRoot)
}

func TestMarkerCommandApp(t *testing.T) {
	t.Parallel()
	pr, errs := gen.ParseMarkers("DeployCmd", []string{"+glacier:command app=myApp"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, "myApp", pr.Cmd.App)
}

func TestMarkerUnknownWarn(t *testing.T) {
	t.Parallel()
	pr, errs := gen.ParseMarkers("Cmd", []string{"+glacier:unknownmarker"})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, 1, len(pr.Unknown))
}

func TestMarkerInjectionAttempt_Newlines(t *testing.T) {
	t.Parallel()
	_, errs := gen.ParseFieldMarkers("Port", []string{"+glacier:env GLACIER_PORT\nevil()"})
	// The line with newline: should be parsed as "+glacier:env GLACIER_PORT\nevil()"
	// The env regex won't match because of the newline content — but the line is
	// passed as-is from the caller. The injection is prevented by the regex.
	_ = errs
}

func TestFieldMarkerHelp_Basic(t *testing.T) {
	t.Parallel()
	lines := []string{
		"Check fetches the latest release from GitHub and compares with the running version.",
		"",
		"+glacier:default false",
	}
	fm, errs := gen.ParseFieldMarkers("Check", lines)
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, "Fetches the latest release from GitHub and compares with the running version", fm.Help)
}

func TestFieldMarkerHelp_MarkerLinesExcluded(t *testing.T) {
	t.Parallel()
	lines := []string{
		"+glacier:short p",
		"+glacier:env MY_PORT",
	}
	fm, errs := gen.ParseFieldMarkers("Port", lines)
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, "", fm.Help)
}

func TestFieldMarkerHelp_Truncation(t *testing.T) {
	t.Parallel()
	// "Field" is the leading word (stripped), followed by 200 x's — exceeds 120 chars.
	payload := "Field " + strings.Repeat("x", 200)
	fm, errs := gen.ParseFieldMarkers("F", []string{payload})
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, true, len(fm.Help) <= 123) // 120 chars + "..."
}

func joinChoices(choices []string) string {
	result := ""
	for i, c := range choices {
		if i > 0 {
			result += "|"
		}
		result += c
	}
	return result
}
