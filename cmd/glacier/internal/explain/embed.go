// SPDX-License-Identifier: Apache-2.0

package explain

import "embed"

// FS holds the embedded topic Markdown files under topics/.
// Each file contains YAML front matter followed by the topic body.
//
//go:embed topics/*.md
var FS embed.FS
