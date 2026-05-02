// SPDX-License-Identifier: Apache-2.0

package concur

import "github.com/nathanbrophy/glacier/errs"

var (
	// ErrCancelled is returned by any blocking concur operation when the caller's
	// context is cancelled. errors.Is(err, context.Canceled) holds.
	ErrCancelled = errs.Sentinel("concur: cancelled")

	// ErrInvalidPermits is returned by Semaphore operations when n is
	// non-positive or exceeds capacity.
	ErrInvalidPermits = errs.Sentinel("concur: invalid permits")
)
