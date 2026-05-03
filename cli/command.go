// SPDX-License-Identifier: Apache-2.0

package cli

import "context"

// command is the unexported interface every CLI command must satisfy.
// Users implement Run(ctx context.Context) error on their struct.
// The interface is intentionally unexported :  users never reference it by name.
type command interface {
	Run(ctx context.Context) error
}

// Stub is a do-nothing implementation of the command interface.
// Useful for parent nodes that exist only to group children.
type Stub struct{}

// Run implements command. It always returns nil.
func (Stub) Run(_ context.Context) error { return nil }
