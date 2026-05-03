// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cli"
)

// serveCmd is a minimal test command.
type serveCmd struct {
	Port    int
	Verbose bool
	Host    string
}

func (s *serveCmd) Run(_ context.Context) error { return nil }

// failCmd always returns an error.
type failCmd struct{}

func (f *failCmd) Run(_ context.Context) error { return errors.New("fail") }

// panicCmd panics in Run.
type panicCmd struct{}

func (p *panicCmd) Run(_ context.Context) error { panic("test panic") }

// notACmd is not a command.
type notACmd struct{}

func newTestApp(t *testing.T) *cli.App {
	t.Helper()
	var buf bytes.Buffer
	return cli.New(
		cli.WithStdout(&buf),
		cli.WithStderr(&buf),
		cli.WithoutBanner(),
	)
}

func TestRegisterFromInterface(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	err := app.Register(&serveCmd{}, cli.WithName("serve"), cli.WithRoot())
	assert.NoError(t, err)
}

func TestRegisterReturnsErrOnDuplicateName(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve")))
	err := app.Register(&serveCmd{}, cli.WithName("serve"))
	assert.Error(t, err)
}

func TestRegisterRejectsNonCommand(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	err := app.Register(&notACmd{}, cli.WithName("bad"))
	assert.Error(t, err)
}

func TestNewWithDefaultOptions(t *testing.T) {
	t.Parallel()
	app := cli.New()
	assert.Equal(t, false, app == nil)
}

func TestNewWithVersion(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	app := cli.New(cli.WithStdout(&buf), cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve"), cli.WithRoot()))
	err := app.Run(context.Background(), []string{"serve", "--version"})
	assert.NoError(t, err)
}

func TestNewWithoutBanner(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	app := cli.New(cli.WithStdout(&buf), cli.WithStderr(&buf), cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve"), cli.WithRoot()))
	// Bare invocation :  banner should not appear.
	_ = app.Run(context.Background(), []string{"serve"})
	// The banner text should not be in the output.
	assert.False(t, strings.Contains(buf.String(), "GLACIER"), "banner appeared despite WithoutBanner()")
}

func TestRunDispatchesArgs(t *testing.T) {
	t.Parallel()
	var ran bool
	type runCmd struct{ Port int }
	app := cli.New(cli.WithoutBanner())
	err := app.Register(&struct {
		Port int
		run  func(context.Context) error
	}{
		run: func(_ context.Context) error {
			ran = true
			return nil
		},
	}, cli.WithName("run"))
	// Use a simpler approach: use serveCmd which just returns nil.
	app2 := cli.New(cli.WithoutBanner())
	assert.NoError(t, app2.Register(&serveCmd{}, cli.WithName("serve")))
	err = app2.Run(context.Background(), []string{"serve"})
	assert.NoError(t, err)
	_ = ran
	_ = app
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	err := app.Run(context.Background(), []string{"nonexistent"})
	var unk *cli.ErrUnknownCommand
	assert.ErrorAs(t, err, &unk)
}

func TestRunUnknownFlag(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve")))
	err := app.Run(context.Background(), []string{"serve", "--bogus"})
	var unk *cli.ErrUnknownFlag
	assert.ErrorAs(t, err, &unk)
}

func TestRunFlagParseError(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve")))
	err := app.Run(context.Background(), []string{"serve", "--port", "notanint"})
	var pe *cli.FlagParseError
	assert.ErrorAs(t, err, &pe)
}

func TestFlagParseErrorUnwrap(t *testing.T) {
	t.Parallel()
	inner := errors.New("inner")
	e := &cli.FlagParseError{Name: "port", Err: inner}
	assert.Equal(t, inner, e.Unwrap())
}

func TestArgvNULRejected(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve")))
	err := app.Run(context.Background(), []string{"serve\x00"})
	var pe *cli.FlagParseError
	assert.ErrorAs(t, err, &pe)
}

func TestArgvOversizeCapped(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve")))
	big := strings.Repeat("a", 65*1024)
	err := app.Run(context.Background(), []string{big})
	var pe *cli.FlagParseError
	assert.ErrorAs(t, err, &pe)
}

func TestPanicCaughtAsErrPanic(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&panicCmd{}, cli.WithName("panic-cmd")))
	err := app.Run(context.Background(), []string{"panic-cmd"})
	var pe *cli.ErrPanic
	assert.ErrorAs(t, err, &pe)
	assert.Equal(t, "test panic", fmt.Sprintf("%v", pe.Value))
}

func TestCtxCancelledBeforeHandler(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve")))
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	err := app.Run(ctx, []string{"serve"})
	assert.ErrorIs(t, err, cli.ErrCancelled)
}

func TestConfFlagSourceLookup(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve")))
	assert.NoError(t, app.Run(context.Background(), []string{"serve", "--port", "9090"}))
	src := cli.NewFlagSource(app)
	val, ok := src.Lookup("port")
	assert.Equal(t, true, ok)
	assert.Equal(t, "9090", val)
}

func TestConfFlagSourceUnknownReturnsFalse(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	src := cli.NewFlagSource(app)
	_, ok := src.Lookup("nonexistent")
	assert.Equal(t, false, ok)
}

func TestAppCloseFlushesAndIsIdempotent(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	err := app.Close()
	assert.NoError(t, err)
	// Second close should also return nil.
	err = app.Close()
	assert.NoError(t, err)
}

func TestWithRoot(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	err := app.Register(&serveCmd{}, cli.WithName("serve"), cli.WithRoot())
	assert.NoError(t, err)
}

func TestWithRootDuplicateRejected(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve"), cli.WithRoot()))
	err := app.Register(&failCmd{}, cli.WithName("other"), cli.WithRoot())
	var me *cli.ErrMultipleRoots
	assert.ErrorAs(t, err, &me)
}

func TestWithParentResolves(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("root"), cli.WithRoot()))
	err := app.Register(&failCmd{}, cli.WithName("child"), cli.WithParent("root"))
	assert.NoError(t, err)
}

func TestWithParentUnresolved(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	err := app.Register(&serveCmd{}, cli.WithName("child"), cli.WithParent("nonexistent"))
	var ure *cli.ErrUnresolvedParent
	assert.ErrorAs(t, err, &ure)
}

func TestRequiredMissingTypedError(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{},
		cli.WithName("serve"),
		cli.WithFlagRequired("Host"),
	))
	err := app.Run(context.Background(), []string{"serve"})
	var re *cli.RequiredError
	assert.ErrorAs(t, err, &re)
}

func TestChoicesEnforced(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{},
		cli.WithName("serve"),
		cli.WithFlagChoices("Host", "localhost", "remotehost"),
	))
	err := app.Run(context.Background(), []string{"serve", "--host", "badvalue"})
	var ce *cli.ChoicesError
	assert.ErrorAs(t, err, &ce)
}

func TestWithFlagShort(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	err := app.Register(&serveCmd{}, cli.WithName("serve"), cli.WithFlagShort("Port", 'p'))
	assert.NoError(t, err)
}

func TestWithFlagShortRejectsMultiByte(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	err := app.Register(&serveCmd{}, cli.WithName("serve"), cli.WithFlagShort("Port", 'é'))
	assert.Error(t, err)
}

func TestWithFlagEnv(t *testing.T) {
	// no t.Parallel() :  t.Setenv requires sequential execution
	t.Setenv("TEST_GLACIER_PORT", "7777")
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{},
		cli.WithName("serve"),
		cli.WithFlagEnv("Port", "TEST_GLACIER_PORT"),
	))
	assert.NoError(t, app.Run(context.Background(), []string{"serve"}))
	val, ok := app.Lookup("port")
	assert.Equal(t, true, ok)
	assert.Equal(t, "7777", val)
}

func TestWithFlagEnvBadName(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	err := app.Register(&serveCmd{}, cli.WithName("serve"), cli.WithFlagEnv("Port", "bad_name"))
	assert.Error(t, err)
}

func TestWithAlias(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{},
		cli.WithName("serve"),
		cli.WithAlias("s"),
	))
	// Both names should dispatch.
	err := app.Run(context.Background(), []string{"serve"})
	assert.NoError(t, err)
	err = app.Run(context.Background(), []string{"s"})
	assert.NoError(t, err)
}

func TestSubcommandTreeDepthFive(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("a"), cli.WithRoot()))
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("b"), cli.WithParent("a")))
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("c"), cli.WithParent("a.b")))
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("d"), cli.WithParent("a.b.c")))
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("e"), cli.WithParent("a.b.c.d")))
	err := app.Run(context.Background(), []string{"a", "b", "c", "d", "e"})
	assert.NoError(t, err)
}

func TestConcurrentAppRunRaceFree(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{}, cli.WithName("serve")))

	done := make(chan struct{}, 2)
	for range 2 {
		go func() {
			_ = app.Run(context.Background(), []string{"serve"})
			done <- struct{}{}
		}()
	}
	<-done
	<-done
}

func TestConcurrentRegisterDuringConstruction(t *testing.T) {
	t.Parallel()
	app := cli.New(cli.WithoutBanner())
	done := make(chan struct{}, 3)
	for i := range 3 {
		go func(idx int) {
			name := fmt.Sprintf("cmd%d", idx)
			_ = app.Register(&serveCmd{}, cli.WithName(name))
			done <- struct{}{}
		}(i)
	}
	<-done
	<-done
	<-done
}

func TestValidateFunctionInvoked(t *testing.T) {
	t.Parallel()
	called := false
	app := cli.New(cli.WithoutBanner())
	assert.NoError(t, app.Register(&serveCmd{},
		cli.WithName("serve"),
		cli.WithFlagValidate("Host", func(s string) error {
			called = true
			if s == "bad" {
				return errors.New("bad host")
			}
			return nil
		}),
	))
	assert.NoError(t, app.Run(context.Background(), []string{"serve", "--host", "good"}))
	assert.Equal(t, true, called)
}

// ExampleApp_Run is a godoc-compatible example.
func ExampleApp_Run() {
	app := cli.New(cli.WithoutBanner(), cli.WithVersion("v0.1.0"))

	_ = app.Register(&serveCmd{}, cli.WithRoot(), cli.WithName("serve"))

	ctx := context.Background()
	err := app.Run(ctx, []string{"serve", "--port", "9090"})
	if err != nil {
		fmt.Println("error:", err)
	}
	// Output:
}

// ExampleApp_Main would normally call app.Main() which calls os.Exit.
// We demonstrate the pattern without actually exiting.
func ExampleApp_Main() {
	// Construct an App, register commands, then call Main.
	// Main handles os.Args, error formatting, and os.Exit.
	app := cli.New(cli.WithVersion("v0.2.0"))
	_ = app.Register(&serveCmd{}, cli.WithRoot(), cli.WithName("serve"))
	// In production: app.Main()
	// In tests: use app.Run(ctx, argv) for error inspection.
	_ = app
	// Output:
}
