// SPDX-License-Identifier: Apache-2.0

package commands

// Exit codes for the glacier binary per spec 0032 D-S27.
const (
	exitSuccess        = 0
	exitGeneric        = 1
	exitUsage          = 2
	exitGenerateFailed = 64
	exitLintFindings   = 65
	exitTestsFailed    = 66
	exitScaffoldFailed = 67
	exitVersionCheck   = 68
	exitCodegenDrift   = 69
	exitSubprocess     = 70
	exitInterrupted    = 130
	exitTerminated     = 143
)
