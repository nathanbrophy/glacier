// SPDX-License-Identifier: Apache-2.0

package cli

import "runtime/debug"

// readBuildVersion reads the version string from runtime/debug.ReadBuildInfo.
// Returns "(devel)" when no version information is available.
func readBuildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "(devel)"
}
