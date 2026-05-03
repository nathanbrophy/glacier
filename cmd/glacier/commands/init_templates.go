// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"embed"
	"io/fs"
)

// embeddedTemplates holds the scaffold template files. The embed pattern
// includes all files under templates/ including subdirectories.
//
//go:embed templates/cli-app templates/library-only templates/both
var embeddedTemplates embed.FS

// templatesFS returns the root of the embedded templates directory as an fs.FS.
func templatesFS() fs.FS {
	sub, err := fs.Sub(embeddedTemplates, "templates")
	if err != nil {
		// embed.FS Sub never fails for a path present in the embed directive.
		panic("init: templates sub-fs: " + err.Error())
	}
	return sub
}
