// SPDX-License-Identifier: Apache-2.0

package obs

import (
	"context"

	glacierErrs "github.com/nathanbrophy/glacier/errs"
)

// Shutdown flushes pending spans and metrics and releases resources. Idempotent.
func (p *Provider) Shutdown(ctx context.Context) error {
	var collected []error
	p.once.Do(func() {
		if p.tracer != nil {
			if err := p.tracer.Shutdown(ctx); err != nil {
				collected = append(collected, err)
			}
		}
		if p.meter != nil {
			if err := p.meter.Shutdown(ctx); err != nil {
				collected = append(collected, err)
			}
		}
	})
	return glacierErrs.Join(collected...)
}
