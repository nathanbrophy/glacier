// SPDX-License-Identifier: Apache-2.0

package fixture

import "github.com/nathanbrophy/glacier/errs"

// ErrPathRejected is returned (wrapped) when a test-data path fails safefile
// validation :  e.g. it contains "..", is absolute, or is a Windows UNC path.
var ErrPathRejected = errs.Sentinel("fixture: path rejected")
