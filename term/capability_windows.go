// SPDX-License-Identifier: Apache-2.0

//go:build windows

package term

import (
	"golang.org/x/term"
)

// isTTY reports whether fd refers to a console on Windows.
func isTTY(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}

// termSize returns the console dimensions on Windows.
func termSize(fd uintptr) (width, height int) {
	w, h, err := term.GetSize(int(fd))
	if err != nil {
		return 0, 0
	}
	return w, h
}
