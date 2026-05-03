//go:build glacier_codegen_fixture

// Package main is the entry point for the myapp CLI.
package main

import (
	"context"
)

// MyApp is the root command.
//
// +glacier:command name=myapp
// +glacier:root
type MyApp struct{}

// Run is a no-op; subcommands do the work.
func (MyApp) Run(_ context.Context) error { return nil }

// StartCmd starts the server.
//
// +glacier:command name=start parent=myapp
type StartCmd struct {
	// Port is the TCP port to listen on.
	//
	// +glacier:default 8080
	// +glacier:short s
	Port int
}

// Run starts the server.
func (c *StartCmd) Run(_ context.Context) error {
	// implementation goes here
	return nil
}

// StopCmd stops the server gracefully.
//
// +glacier:command name=stop parent=myapp
type StopCmd struct {
	// Timeout is the graceful-shutdown timeout in seconds.
	//
	// +glacier:default 30
	Timeout int
}

// Run stops the server.
func (c *StopCmd) Run(_ context.Context) error {
	// implementation goes here
	return nil
}
