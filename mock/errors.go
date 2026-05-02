// SPDX-License-Identifier: Apache-2.0

package mock

import "github.com/nathanbrophy/glacier/errs"

// ErrUnexpectedCall is the sentinel returned (and wrapped) when a method is
// called with no matching expectation in strict mode. Tests may use
// errors.Is(err, mock.ErrUnexpectedCall) to detect these failures.
var ErrUnexpectedCall = errs.Sentinel("mock: unexpected call")

// ErrTypeMismatch is the sentinel used when a return-value type does not match
// the method's declared return type. It is wrapped in the panic message
// produced at Return registration time.
var ErrTypeMismatch = errs.Sentinel("mock: type mismatch")
