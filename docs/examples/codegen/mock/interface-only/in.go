//go:build glacier_codegen_fixture

// Package store defines the data-access interface for the application.
package store

import "context"

// Store is the data-access interface.
//
// +glacier:mock
type Store interface {
	// Get retrieves a record by ID.
	Get(ctx context.Context, id string) (string, error)

	// Put writes a record by ID.
	Put(ctx context.Context, id string, value string) error
}
