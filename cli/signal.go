// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"

	"github.com/nathanbrophy/glacier/internal/sigh"
)

// installSignals wraps ctx with signal-based cancellation via internal/sigh.
func installSignals(ctx context.Context) (context.Context, context.CancelFunc) {
	return sigh.Notify(ctx)
}
