// SPDX-License-Identifier: Apache-2.0

package cli

// FlagSource adapts an App into a conf.Source interface.
// cli must NOT import conf (Tier-1 kernel rule). FlagSource implements the
// lookup duck-typing pattern; conf expects any type with Lookup(string) (string, bool).
type FlagSource struct{ app *App }

// NewFlagSource constructs a FlagSource from the given App.
// Precondition: app must not be nil.
func NewFlagSource(app *App) *FlagSource { return &FlagSource{app: app} }

// Lookup implements conf.Source. Returns the string-serialized flag value.
// Returns ("", false) for unknown or unset flags.
func (f *FlagSource) Lookup(key string) (string, bool) { return f.app.Lookup(key) }
