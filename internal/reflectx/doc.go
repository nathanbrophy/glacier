// SPDX-License-Identifier: Apache-2.0

// Package reflectx provides reflection helpers shared by mock, fluent, and
// conf. It centralizes the one-time-setup reflection work that those packages
// cache per type, ensuring reflection overhead is paid at construction rather
// than on hot paths. The package is internal to the Glacier module and may not
// be imported outside it.
package reflectx
