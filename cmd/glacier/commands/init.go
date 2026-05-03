// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/nathanbrophy/glacier/cmd/glacier/internal/figgen"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/mascots"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/report"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/sdkerr"
	"github.com/nathanbrophy/glacier/internal/safefile"
	"github.com/nathanbrophy/glacier/term"
)

// InitCmd scaffolds a new Glacier project.
//
// +glacier:command name=init parent=glacier
type InitCmd struct {
	// Dir is the target directory (default: current directory).
	//
	// +glacier:positional
	Dir string

	// Name is the module/app name (e.g. github.com/acme/myapp).
	Name string

	// Template selects the project template.
	//
	// +glacier:choices library-only|cli-app|both
	// +glacier:default cli-app
	Template string

	// License selects the license to include.
	//
	// +glacier:choices Apache-2.0|MIT|BSD-3-Clause|none
	// +glacier:default Apache-2.0
	License string

	// Mascot selects the project mascot.
	//
	// +glacier:choices polar_bear|penguin|owl|fox|otter|raccoon
	// +glacier:default polar_bear
	Mascot string

	// NoGit suppresses git init.
	//
	// +glacier:default false
	NoGit bool

	// Yes skips interactive prompts, using defaults.
	//
	// +glacier:short y
	// +glacier:default false
	Yes bool

	// Force overwrites files in a non-empty directory.
	//
	// +glacier:default false
	Force bool

	// tplFS is the template filesystem. Defaults to templatesFS().
	// Injected by tests via withTemplateFS.
	tplFS fs.FS

	// out is the writer for the success message. nil means os.Stdout.
	// Injected by tests to avoid os.Stdout races with parallel tests.
	out io.Writer
}

// withTemplateFS returns a copy of c with the given fs.FS injected.
// Used by tests to exercise template rendering against an in-memory FS.
func (c *InitCmd) withTemplateFS(fsys fs.FS) *InitCmd {
	cp := *c
	cp.tplFS = fsys
	return &cp
}

// withOut returns a copy of c with the given writer injected.
func (c *InitCmd) withOut(w io.Writer) *InitCmd {
	cp := *c
	cp.out = w
	return &cp
}

// stdout returns the effective output writer.
func (c *InitCmd) stdout() io.Writer {
	if c.out != nil {
		return c.out
	}
	return os.Stdout
}

// templateFS returns the injected FS, or the embedded default.
func (c *InitCmd) templateFS() fs.FS {
	if c.tplFS != nil {
		return c.tplFS
	}
	return templatesFS()
}

// templateData holds the values interpolated into every scaffold template.
type templateData struct {
	// AppName is the last path segment of ModulePath (e.g. "myapp").
	AppName string
	// ModulePath is the full Go module path (e.g. "github.com/acme/myapp").
	ModulePath string
	// License is the SPDX identifier chosen at init time.
	License string
	// Mascot is the stable mascot ID (e.g. "polar_bear").
	Mascot string
	// MascotKaomoji is the single-line kaomoji for the chosen mascot.
	MascotKaomoji string
	// GoVersion is the host's go version (e.g. "1.26.0") from runtime.Version.
	GoVersion string
	// GlacierVersion is the running SDK version prefixed with "v".
	GlacierVersion string
	// BannerLines holds the figgen-rendered banner for the app name.
	BannerLines []string
}

// tplEntry maps a template filename (in the embed.FS) to its output path
// relative to the target directory. The output path may contain template
// variables resolved at render time.
type tplEntry struct {
	// src is the path within the sub-FS (relative to the template tree root).
	src string
	// dst is the output path relative to the scaffold target directory.
	// Unlike src, dst is computed by substituting AppName where needed.
	dst string
}

// Run scaffolds a new Glacier project.
func (c *InitCmd) Run(ctx context.Context) error {
	report.Status(report.Calm, "glacier init")

	dir := c.Dir
	if dir == "" {
		dir = "."
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("init: resolve dir: %w", err)
	}

	// Collect parameters interactively or from flags.
	data, err := c.collectParams(ctx, absDir)
	if err != nil {
		return err
	}

	// Determine which template tree to use.
	tplName := c.Template
	if tplName == "" {
		tplName = "cli-app"
	}

	entries, err := buildEntries(tplName, data.AppName)
	if err != nil {
		return &exitCodeError{code: exitScaffoldFailed, cause: err}
	}

	// Check for collisions unless --force.
	if !c.Force {
		if collErr := checkCollisions(absDir, entries); collErr != nil {
			return collErr
		}
	} else if !c.Yes {
		// --force without --yes: prompt once for confirmation.
		ok, promptErr := term.Confirm(ctx, "Overwrite existing files?")
		if promptErr != nil || !ok {
			return &exitCodeError{code: exitInterrupted, cause: fmt.Errorf("init: overwrite cancelled")}
		}
	}

	// Create the target directory.
	if mkErr := os.MkdirAll(absDir, 0o755); mkErr != nil {
		return fmt.Errorf("init: mkdir: %w", mkErr)
	}

	// Write scaffolded files.
	if writeErr := c.writeScaffold(absDir, entries, data, tplName); writeErr != nil {
		report.Status(report.Err, "scaffold failed: "+writeErr.Error())
		return &exitCodeError{code: exitScaffoldFailed, cause: writeErr}
	}

	// Write the figgen banner.
	bannerContent := strings.Join(data.BannerLines, "\n") + "\n"
	if writeErr := safefile.WriteFileAtomic(absDir, "assets/banner.txt", []byte(bannerContent), 0o644); writeErr != nil {
		return fmt.Errorf("init: write assets/banner.txt: %w", writeErr)
	}

	// Run git init unless suppressed or .git already exists.
	if !c.NoGit {
		if _, statErr := os.Stat(filepath.Join(absDir, ".git")); os.IsNotExist(statErr) {
			gitCmd := exec.CommandContext(ctx, "git", "init", absDir)
			gitCmd.Stderr = os.Stderr
			if gitErr := gitCmd.Run(); gitErr != nil {
				report.Status(report.Alarmed, "git init failed: "+gitErr.Error())
			}
		}
	}

	// Print success box.
	content := fmt.Sprintf(
		"Project %q is ready.\n\nNext steps:\n  cd %s\n  glacier generate\n  glacier test\n",
		data.AppName, dir,
	)
	box := term.Box(content,
		term.WithTitle("ʕ⌐■-■ʔ  all set."),
		term.WithRoundedCorners(),
		term.WithPadding(1, 2, 1, 2),
	)
	fmt.Fprintln(c.stdout(), box)
	return nil
}

// collectParams gathers scaffold parameters, prompting if not --yes.
func (c *InitCmd) collectParams(ctx context.Context, absDir string) (templateData, error) {
	name := c.Name
	mascotID := c.Mascot
	if mascotID == "" {
		mascotID = "polar_bear"
	}
	license := c.License
	if license == "" {
		license = "Apache-2.0"
	}

	if !c.Yes {
		// Interactive prompts.
		if name == "" {
			prompted, err := term.Prompt(ctx, "Module name (e.g. github.com/you/app): ")
			if err != nil {
				return templateData{}, err
			}
			name = strings.TrimSpace(prompted)
		}

		// Mascot selection.
		all := mascots.All()
		chosen, err := term.Select(ctx, "Choose a mascot:", all, func(m mascots.Mascot) string {
			return m.Kaomoji + " " + m.Display
		})
		if err == nil {
			mascotID = chosen.ID
		}
	}

	if name == "" {
		// Derive from directory name.
		name = filepath.Base(absDir)
	}

	// App name is the last path segment of the module name.
	appName := name
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		appName = name[idx+1:]
	}

	m := mascots.Get(mascotID)
	bannerLines := figgen.Render(appName)

	// Extract numeric Go version (strip "go" prefix, e.g. "go1.26.0" -> "1.26.0").
	goVer := strings.TrimPrefix(runtime.Version(), "go")
	// Keep only MAJOR.MINOR.PATCH.
	if parts := strings.SplitN(goVer, " ", 2); len(parts) > 0 {
		goVer = parts[0]
	}

	return templateData{
		AppName:        appName,
		ModulePath:     name,
		License:        license,
		Mascot:         m.ID,
		MascotKaomoji:  m.Kaomoji,
		GoVersion:      goVer,
		GlacierVersion: "v" + Version,
		BannerLines:    bannerLines,
	}, nil
}

// buildEntries returns the list of template-to-output-path mappings for the
// given template name. AppName is substituted into output paths that need it.
func buildEntries(tplName, appName string) ([]tplEntry, error) {
	cmdDir := filepath.Join("cmd", appName)

	cliAppEntries := []tplEntry{
		{src: "cli-app/go.mod.tmpl", dst: "go.mod"},
		{src: "cli-app/gitignore.tmpl", dst: ".gitignore"},
		{src: "cli-app/README.md.tmpl", dst: "README.md"},
		{src: "cli-app/CLAUDE.md.tmpl", dst: "CLAUDE.md"},
		{src: "cli-app/LICENSE.tmpl", dst: "LICENSE"},
		{src: "cli-app/cmd_main.go.tmpl", dst: filepath.Join(cmdDir, "main.go")},
		{src: "cli-app/cmd_serve.go.tmpl", dst: filepath.Join(cmdDir, "serve.go")},
		{src: "cli-app/cmd_zz_generated_cli.go.tmpl", dst: filepath.Join(cmdDir, "zz_generated_cli.go")},
		{src: "cli-app/assets_mascot.go.tmpl", dst: filepath.Join("assets", "mascot.go")},
	}

	libOnlyEntries := []tplEntry{
		{src: "library-only/go.mod.tmpl", dst: "go.mod"},
		{src: "library-only/gitignore.tmpl", dst: ".gitignore"},
		{src: "library-only/README.md.tmpl", dst: "README.md"},
		{src: "library-only/pkg_example.go.tmpl", dst: filepath.Join(appName, appName+".go")},
	}

	switch tplName {
	case "cli-app":
		return cliAppEntries, nil
	case "library-only":
		return libOnlyEntries, nil
	case "both":
		// Union: cli-app entries plus the library package file.
		entries := append([]tplEntry{}, cliAppEntries...)
		entries = append(entries, tplEntry{
			src: "both/pkg_example.go.tmpl",
			dst: filepath.Join(appName, appName+".go"),
		})
		return entries, nil
	default:
		return nil, fmt.Errorf("init: unknown template %q", tplName)
	}
}

// checkCollisions returns an error if any output file already exists.
func checkCollisions(absDir string, entries []tplEntry) error {
	var conflicts []string
	for _, e := range entries {
		if _, err := os.Stat(filepath.Join(absDir, e.dst)); err == nil {
			conflicts = append(conflicts, e.dst)
		}
	}
	if len(conflicts) > 0 {
		return &sdkerr.ErrAlreadyInitialized{Path: absDir, Conflict: conflicts}
	}
	return nil
}

// writeScaffold renders each template entry and writes it to absDir.
func (c *InitCmd) writeScaffold(absDir string, entries []tplEntry, data templateData, tplName string) error {
	fsys := c.templateFS()
	for _, e := range entries {
		tmplBytes, err := fs.ReadFile(fsys, e.src)
		if err != nil {
			return fmt.Errorf("read template %s: %w", e.src, err)
		}
		rendered, err := renderTemplate(string(tmplBytes), data)
		if err != nil {
			return fmt.Errorf("render %s: %w", e.src, err)
		}
		// Normalize to forward slashes for safefile.
		dstFwd := filepath.ToSlash(e.dst)
		if writeErr := safefile.WriteFileAtomic(absDir, dstFwd, rendered, 0o644); writeErr != nil {
			return fmt.Errorf("write %s: %w", e.dst, writeErr)
		}
		report.Status(report.Calm, "  ✓ "+e.dst)
	}
	return nil
}

// renderTemplate parses and executes a Go text template with data.
func renderTemplate(tmplStr string, data templateData) ([]byte, error) {
	t, err := template.New("").Parse(tmplStr)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
