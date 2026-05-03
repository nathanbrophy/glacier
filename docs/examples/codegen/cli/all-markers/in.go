//go:build glacier_codegen_fixture

// Package main is the entry point for the myapp CLI.
package main

import (
	"context"
	"errors"
)

// MyApp is the root command.
//
// +glacier:command name=myapp
// +glacier:root
type MyApp struct{}

// Run is a no-op; subcommands do the work.
func (MyApp) Run(_ context.Context) error { return nil }

// DeployCmd deploys the application. It demonstrates every supported marker.
//
// +glacier:command name=deploy parent=myapp
type DeployCmd struct {
	// Env is the target environment.
	//
	// +glacier:required
	// +glacier:choices staging|production
	// +glacier:short e
	// +glacier:env DEPLOY_ENV
	Env string

	// Parallel controls the number of concurrent deployments.
	//
	// +glacier:default 2
	Parallel int

	// DryRun previews changes without applying them.
	//
	// +glacier:default false
	DryRun bool

	// Strategy is the rollout strategy.
	//
	// +glacier:choices rolling|blue-green|canary
	// +glacier:default rolling
	// +glacier:validate validateStrategy
	Strategy string

	// OldFlag is kept for backward compatibility.
	//
	// +glacier:deprecated Use --strategy instead.
	OldFlag string

	// Target is the positional deployment target (e.g. a cluster name).
	//
	// +glacier:positional
	Target string
}

// Run deploys the application.
func (c *DeployCmd) Run(_ context.Context) error {
	// implementation goes here
	return nil
}

// validateStrategy rejects unknown strategy strings.
func validateStrategy(s string) error {
	switch s {
	case "rolling", "blue-green", "canary":
		return nil
	}
	return errors.New("unknown strategy: " + s)
}
