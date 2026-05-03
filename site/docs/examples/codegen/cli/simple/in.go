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

// ServeCmd starts the HTTP server.
//
// +glacier:command name=serve parent=myapp
type ServeCmd struct {
	// Port is the TCP port to listen on.
	//
	// +glacier:default 8080
	Port int

	// Config is the path to the config file.
	//
	// +glacier:positional
	Config string
}

// Run starts the server on the configured port.
func (c *ServeCmd) Run(_ context.Context) error {
	// implementation goes here
	return nil
}
