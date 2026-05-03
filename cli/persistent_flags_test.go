// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cli"
)

// rootWithGlobals is a minimal root command that exposes the same
// persistent-flag shape as GlacierCmd: a quiet/verbose ladder, a no-color
// switch, and a string profile path. It records ApplyRoot invocations and
// the runFn last-seen flag values so tests can verify both the parse
// outcome and the lifecycle hook contract.
type rootWithGlobals struct {
	Quiet      bool
	Verbose    bool
	NoColor    bool
	Profile    string
	applied    atomic.Int32
	applyErr   error
	seenQuiet  bool
	seenVerb   bool
	seenColor  bool
	seenProf   string
	wantApplyN int32
}

func (r *rootWithGlobals) Run(_ context.Context) error { return nil }

func (r *rootWithGlobals) ApplyRoot(_ context.Context) error {
	r.applied.Add(1)
	r.seenQuiet = r.Quiet
	r.seenVerb = r.Verbose
	r.seenColor = r.NoColor
	r.seenProf = r.Profile
	return r.applyErr
}

// childCmd is a subcommand with no flag overlap; its Run records that it
// was reached so tests can confirm dispatch went past the root.
type childCmd struct {
	Patterns string
	ran      atomic.Int32
}

func (c *childCmd) Run(_ context.Context) error { c.ran.Add(1); return nil }

func newRootChildApp(t *testing.T) (*cli.App, *rootWithGlobals, *childCmd) {
	t.Helper()
	var buf bytes.Buffer
	app := cli.New(cli.WithStdout(&buf), cli.WithStderr(&buf), cli.WithoutBanner())
	root := &rootWithGlobals{}
	child := &childCmd{}
	assert.NoError(t, app.Register(root, cli.WithRoot(), cli.WithName("glacier")))
	assert.NoError(t, app.Register(child, cli.WithName("child")))
	return app, root, child
}

func TestPersistentFlags(t *testing.T) {
	t.Parallel()
	type expect struct {
		applyN  int32
		quiet   bool
		verbose bool
		noColor bool
		profile string
		ranKid  int32
	}
	rows := []struct {
		name string
		argv []string
		want expect
	}{
		{
			name: "verbose_before_subcommand",
			argv: []string{"--verbose", "child"},
			want: expect{applyN: 1, verbose: true, ranKid: 1},
		},
		{
			name: "verbose_after_subcommand",
			argv: []string{"child", "--verbose"},
			want: expect{applyN: 1, verbose: true, ranKid: 1},
		},
		{
			name: "verbose_around_subcommand",
			argv: []string{"--quiet", "child", "--profile", "out"},
			want: expect{applyN: 1, quiet: true, profile: "out", ranKid: 1},
		},
		{
			name: "string_flag_inline_value",
			argv: []string{"child", "--profile=out.cpu"},
			want: expect{applyN: 1, profile: "out.cpu", ranKid: 1},
		},
		{
			name: "bare_root_no_flags",
			argv: []string{},
			want: expect{applyN: 1},
		},
		{
			name: "double_dash_terminates_peeling",
			argv: []string{"--", "--verbose", "child"},
			want: expect{applyN: 0}, // root not reached: positional after -- treated as cmd path "--verbose"; resolves to ErrUnknownCommand path
		},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			app, root, child := newRootChildApp(t)
			err := app.Run(context.Background(), tc.argv)
			if tc.name == "double_dash_terminates_peeling" {
				// Either ErrUnknownCommand or no-op root depending on splitArgv;
				// the contract we care about is that -- prevents peeling.
				assert.Equal(t, false, root.seenVerb)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want.applyN, root.applied.Load())
			assert.Equal(t, tc.want.quiet, root.seenQuiet)
			assert.Equal(t, tc.want.verbose, root.seenVerb)
			assert.Equal(t, tc.want.noColor, root.seenColor)
			assert.Equal(t, tc.want.profile, root.seenProf)
			assert.Equal(t, tc.want.ranKid, child.ran.Load())
		})
	}
}

func TestApplyRootErrorAbortsDispatch(t *testing.T) {
	t.Parallel()
	app, root, child := newRootChildApp(t)
	root.applyErr = errors.New("global flags rejected")

	err := app.Run(context.Background(), []string{"--verbose", "child"})
	assert.Error(t, err)
	assert.Equal(t, "global flags rejected", err.Error())
	assert.Equal(t, int32(1), root.applied.Load())
	assert.Equal(t, int32(0), child.ran.Load())
}

func TestPersistentFlagsLookupSeesRootValues(t *testing.T) {
	t.Parallel()
	app, _, _ := newRootChildApp(t)
	assert.NoError(t, app.Run(context.Background(), []string{"--verbose", "child"}))
	v, ok := app.Lookup("verbose")
	assert.True(t, ok)
	assert.Equal(t, "true", v)
}

func TestRootFlagInlineValuePeeled(t *testing.T) {
	t.Parallel()
	app, root, child := newRootChildApp(t)
	assert.NoError(t, app.Run(context.Background(), []string{"--profile=cpu.out", "child"}))
	assert.Equal(t, "cpu.out", root.seenProf)
	assert.Equal(t, int32(1), child.ran.Load())
}

func TestRootFlagSeparateValuePeeled(t *testing.T) {
	t.Parallel()
	app, root, child := newRootChildApp(t)
	assert.NoError(t, app.Run(context.Background(), []string{"--profile", "cpu.out", "child"}))
	assert.Equal(t, "cpu.out", root.seenProf)
	assert.Equal(t, int32(1), child.ran.Load())
}

// TestSubcommandUnknownFlagStillErrors guards that peeling does not silently
// swallow flags the subcommand doesn't know about. Any flag that isn't a
// root flag and isn't a subcommand flag must still surface ErrUnknownFlag.
func TestSubcommandUnknownFlagStillErrors(t *testing.T) {
	t.Parallel()
	app, _, _ := newRootChildApp(t)
	err := app.Run(context.Background(), []string{"child", "--bogus"})
	assert.Error(t, err)
	var unk *cli.ErrUnknownFlag
	assert.True(t, errors.As(err, &unk))
}
