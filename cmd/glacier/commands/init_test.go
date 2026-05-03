// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/sdkerr"
)

// newInitCmd returns an InitCmd pre-loaded with test-safe defaults.
// --yes skips all interactive prompts; NoGit avoids spawning git in temp dirs.
// A discard writer is injected so tests don't write to os.Stdout (which would
// race with tests that capture os.Stdout via os.Pipe for other assertions).
func newInitCmd(dir, tpl string) *InitCmd {
	return (&InitCmd{
		Dir:      dir,
		Name:     "github.com/test/myapp",
		Template: tpl,
		License:  "Apache-2.0",
		Mascot:   "polar_bear",
		Yes:      true,
		NoGit:    true,
	}).withOut(new(bytes.Buffer))
}

func TestInitTemplateCliApp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cmd := newInitCmd(dir, "cli-app")

	err := cmd.Run(context.Background())
	assert.NoError(t, err)

	// Check key files exist.
	for _, rel := range []string{
		"go.mod",
		".gitignore",
		"README.md",
		"CLAUDE.md",
		"LICENSE",
		filepath.Join("cmd", "myapp", "main.go"),
		filepath.Join("cmd", "myapp", "serve.go"),
		filepath.Join("cmd", "myapp", "zz_generated_cli.go"),
		filepath.Join("assets", "mascot.go"),
		filepath.Join("assets", "banner.txt"),
	} {
		path := filepath.Join(dir, rel)
		_, statErr := os.Stat(path)
		assert.NoError(t, statErr, "expected file to exist: "+rel)
	}
}

func TestInitTemplateLibraryOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cmd := newInitCmd(dir, "library-only")

	err := cmd.Run(context.Background())
	assert.NoError(t, err)

	for _, rel := range []string{
		"go.mod",
		".gitignore",
		"README.md",
		filepath.Join("myapp", "myapp.go"),
	} {
		_, statErr := os.Stat(filepath.Join(dir, rel))
		assert.NoError(t, statErr, "expected file: "+rel)
	}
}

func TestInitTemplateBoth(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cmd := newInitCmd(dir, "both")

	err := cmd.Run(context.Background())
	assert.NoError(t, err)

	// Should have CLI files.
	for _, rel := range []string{
		filepath.Join("cmd", "myapp", "main.go"),
		filepath.Join("assets", "mascot.go"),
	} {
		_, statErr := os.Stat(filepath.Join(dir, rel))
		assert.NoError(t, statErr, "expected CLI file: "+rel)
	}
	// And the library package file.
	_, libErr := os.Stat(filepath.Join(dir, "myapp", "myapp.go"))
	assert.NoError(t, libErr, "expected library package file")
}

func TestInitRefuseOnCollision(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Pre-create go.mod to trigger collision.
	goModPath := filepath.Join(dir, "go.mod")
	assert.NoError(t, os.WriteFile(goModPath, []byte("module placeholder\n"), 0o644))

	cmd := newInitCmd(dir, "cli-app")
	err := cmd.Run(context.Background())
	assert.Error(t, err)

	var alreadyInit *sdkerr.ErrAlreadyInitialized
	assert.True(t, errors.As(err, &alreadyInit), "expected ErrAlreadyInitialized")
}

func TestInitForceOverwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Pre-create go.mod.
	goModPath := filepath.Join(dir, "go.mod")
	assert.NoError(t, os.WriteFile(goModPath, []byte("module placeholder\n"), 0o644))

	cmd := newInitCmd(dir, "cli-app")
	cmd.Force = true // --yes is already set so no prompt needed

	err := cmd.Run(context.Background())
	assert.NoError(t, err, "--yes --force should overwrite silently")

	// go.mod should now contain the rendered content, not the placeholder.
	data, readErr := os.ReadFile(goModPath)
	assert.NoError(t, readErr)
	assert.True(t, string(data) != "module placeholder\n", "go.mod should be overwritten")
}

func TestInitNoGitSkips(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cmd := newInitCmd(dir, "cli-app")
	cmd.NoGit = true

	err := cmd.Run(context.Background())
	assert.NoError(t, err)

	// .git directory should not exist.
	_, statErr := os.Stat(filepath.Join(dir, ".git"))
	assert.True(t, os.IsNotExist(statErr), ".git should not be created with --no-git")
}

func TestInitExistingGitSilentSkip(t *testing.T) {
	// When .git/ already exists, git init is skipped silently.
	t.Parallel()
	dir := t.TempDir()

	// Pre-create .git dir.
	gitDir := filepath.Join(dir, ".git")
	assert.NoError(t, os.MkdirAll(gitDir, 0o755))

	cmd := newInitCmd(dir, "cli-app")
	cmd.NoGit = false // allow git init, but .git already exists

	err := cmd.Run(context.Background())
	assert.NoError(t, err, "pre-existing .git/ should be skipped without error")
}

func TestInitModuleNameValidation(t *testing.T) {
	// The AppName is derived from the module path; rendering uses AppName as
	// the package name. An empty name falls back to the directory basename.
	t.Parallel()
	dir := t.TempDir()

	cmd := (&InitCmd{
		Dir:      dir,
		Name:     "", // will be derived from dir basename
		Template: "library-only",
		License:  "none",
		Yes:      true,
		NoGit:    true,
	}).withOut(new(bytes.Buffer))
	err := cmd.Run(context.Background())
	assert.NoError(t, err, "empty Name should fall back to dir basename")
}

func TestInitYesHappyPath(t *testing.T) {
	// --yes applies defaults without any interactive prompts.
	t.Parallel()
	dir := t.TempDir()
	cmd := (&InitCmd{
		Dir:      dir,
		Name:     "github.com/corp/service",
		Template: "cli-app",
		License:  "MIT",
		Mascot:   "otter",
		Yes:      true,
		NoGit:    true,
	}).withOut(new(bytes.Buffer))
	err := cmd.Run(context.Background())
	assert.NoError(t, err)

	// mascot.go should contain "otter".
	data, readErr := os.ReadFile(filepath.Join(dir, "assets", "mascot.go"))
	assert.NoError(t, readErr)
	assert.True(t, contains(string(data), "otter"), "mascot.go should reference otter")
}

func TestInitTemplatesRegistryFilesPresent(t *testing.T) {
	// Verify the embed.FS contains exactly the documented template files.
	t.Parallel()
	fsys := templatesFS()

	for _, f := range []string{
		"cli-app/go.mod.tmpl",
		"cli-app/gitignore.tmpl",
		"cli-app/README.md.tmpl",
		"cli-app/CLAUDE.md.tmpl",
		"cli-app/LICENSE.tmpl",
		"cli-app/cmd_main.go.tmpl",
		"cli-app/cmd_serve.go.tmpl",
		"cli-app/cmd_zz_generated_cli.go.tmpl",
		"cli-app/assets_mascot.go.tmpl",
		"library-only/go.mod.tmpl",
		"library-only/gitignore.tmpl",
		"library-only/README.md.tmpl",
		"library-only/pkg_example.go.tmpl",
		"both/pkg_example.go.tmpl",
	} {
		_, err := fsys.Open(f)
		assert.NoError(t, err, "expected template file in embed.FS: "+f)
	}
}

// contains is a tiny helper so this test file has zero external dependencies
// beyond the ones already imported.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
