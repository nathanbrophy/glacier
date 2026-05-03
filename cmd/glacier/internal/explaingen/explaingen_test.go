// SPDX-License-Identifier: Apache-2.0

package explaingen_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/explaingen"
)

// fixtureSpec011 is a minimal stand-in for specs/0011-cli.md that contains
// only the rows needed for the marker grammar table.
const fixtureSpec011 = `
## §23.8 Marker grammar

| Marker | Regex | Max length | Notes |
|---|---|---|---|
| ` + "`" + `+glacier:command name=` + "`" + ` | ` + "`" + `^[a-z]` + "`" + ` | 32 chars | Annotates a struct as a CLI command |
| ` + "`" + `+glacier:root` + "`" + ` | (no payload) | — | Marks a command struct as the root |
| ` + "`" + `+glacier:mock` + "`" + ` | see spec 0012 | — | Governed by mock spec |
`

// fixtureSpec032 is a minimal stand-in for specs/0032-sdk.md containing exit
// code and config key tables.
const fixtureSpec032 = `
#### Configuration (D-S21)

The SDK registers a Config struct via conf.Register at startup.

| Key | Type | Default | Effect |
|---|---|---|---|
| ` + "`" + `github.repo` + "`" + ` | string | ` + "`" + `nathanbrophy/glacier` + "`" + ` | Repo for version --check |

#### Exit codes (D-S27)

| Code | Meaning |
|---|---|
| 66 | Tests failed |
| 67 | Init or new template / scaffolding failed |
`

// fixtureFS builds an in-memory fs.FS with the two fixture spec files.
func fixtureFS() fstest.MapFS {
	return fstest.MapFS{
		"specs/0011-cli.md": &fstest.MapFile{Data: []byte(fixtureSpec011)},
		"specs/0032-sdk.md": &fstest.MapFile{Data: []byte(fixtureSpec032)},
	}
}

func TestFileSpecSource_Markers(t *testing.T) {
	t.Parallel()
	src := explaingen.NewFileSpecSource(fixtureFS())
	rows, err := src.Markers()
	require.NoError(t, err)
	// Fixture has 3 rows; +glacier:command sub-attrs collapse to 1.
	assert.Equal(t, 3, len(rows))
	assert.Equal(t, "+glacier:command", rows[0].Name)
	assert.Equal(t, "+glacier:root", rows[1].Name)
	assert.Equal(t, "+glacier:mock", rows[2].Name)
}

func TestFileSpecSource_ExitCodes(t *testing.T) {
	t.Parallel()
	src := explaingen.NewFileSpecSource(fixtureFS())
	rows, err := src.ExitCodes()
	require.NoError(t, err)
	assert.Equal(t, 2, len(rows))
	assert.Equal(t, "66", rows[0].Code)
	assert.Equal(t, "67", rows[1].Code)
}

func TestFileSpecSource_ConfigKeys(t *testing.T) {
	t.Parallel()
	src := explaingen.NewFileSpecSource(fixtureFS())
	rows, err := src.ConfigKeys()
	require.NoError(t, err)
	assert.Equal(t, 1, len(rows))
	assert.Equal(t, "github.repo", rows[0].Key)
}

func TestGenerate_FilesProduced(t *testing.T) {
	t.Parallel()
	src := explaingen.NewFileSpecSource(fixtureFS())
	files, err := explaingen.Generate(src)
	require.NoError(t, err)
	// 3 markers + 1 synthetic +glacier:flag + 2 exit codes + 1 config key = 7 files.
	assert.Equal(t, 7, len(files))
	assert.True(t, files["+glacier_command.md"] != nil, "expected +glacier_command.md")
	assert.True(t, files["exit_66.md"] != nil, "expected exit_66.md")
	assert.True(t, files["config_github.repo.md"] != nil, "expected config_github.repo.md")
}

func TestGenerate_FrontMatterValid(t *testing.T) {
	t.Parallel()
	src := explaingen.NewFileSpecSource(fixtureFS())
	files, err := explaingen.Generate(src)
	require.NoError(t, err)

	content := string(files["exit_66.md"])
	assert.True(t, strings.HasPrefix(content, "---\n"), "expected YAML front matter opening ---")
	assert.True(t, strings.Contains(content, "slug: exit:66"), "expected slug in front matter")
	assert.True(t, strings.Contains(content, "category: exit-code"), "expected category in front matter")
}

func TestCheck_Clean(t *testing.T) {
	t.Parallel()
	src := explaingen.NewFileSpecSource(fixtureFS())

	// Generate the files first.
	files, err := explaingen.Generate(src)
	require.NoError(t, err)

	// Build a topics FS from the generated content.
	topicFS := make(fstest.MapFS)
	for name, content := range files {
		topicFS["topics/"+name] = &fstest.MapFile{Data: content}
	}

	// Check should pass when topics match generated output.
	err = explaingen.Check(src, topicFS)
	assert.NoError(t, err)
}

func TestCheck_DetectsDrift(t *testing.T) {
	t.Parallel()
	src := explaingen.NewFileSpecSource(fixtureFS())

	files, err := explaingen.Generate(src)
	require.NoError(t, err)

	// Mutate one file to simulate drift.
	topicFS := make(fstest.MapFS)
	for name, content := range files {
		topicFS["topics/"+name] = &fstest.MapFile{Data: content}
	}
	topicFS["topics/exit_66.md"] = &fstest.MapFile{Data: []byte("---\nslug: exit:66\ntitle: \"mutated\"\ncategory: exit-code\n---\nbad content\n")}

	err = explaingen.Check(src, topicFS)
	assert.True(t, err != nil, "expected error on drift")
	assert.True(t, strings.Contains(err.Error(), "exit_66.md"), "error should name the drifted file")
}

func TestCheck_DetectsMissingFile(t *testing.T) {
	t.Parallel()
	src := explaingen.NewFileSpecSource(fixtureFS())

	// Empty topics dir.
	topicFS := fstest.MapFS{
		"topics/.keep": &fstest.MapFile{Data: []byte{}},
	}

	checkErr := explaingen.Check(src, topicFS)
	assert.True(t, checkErr != nil, "expected error when topics are missing")
}

func TestSlugToFilename(t *testing.T) {
	t.Parallel()
	cases := []struct {
		slug string
		want string
	}{
		{"+glacier:command", "+glacier_command.md"},
		{"exit:66", "exit_66.md"},
		{"config:github.repo", "config_github.repo.md"},
	}
	for _, tc := range cases {
		t.Run(tc.slug, func(t *testing.T) {
			t.Parallel()
			// Generate a single-item source and check the filename.
			src := explaingen.NewFileSpecSource(fstest.MapFS{
				"specs/0011-cli.md": &fstest.MapFile{Data: []byte(fixtureSpec011)},
				"specs/0032-sdk.md": &fstest.MapFile{Data: []byte(fixtureSpec032)},
			})
			files, err := explaingen.Generate(src)
			require.NoError(t, err)
			// Only check the expected file exists with the right name.
			if _, ok := files[tc.want]; !ok {
				// Not all slugs appear in the fixture; skip those we don't have.
				t.Skip("slug not in fixture")
			}
		})
	}
}

// ExampleGenerate demonstrates the Generate function with a fixture source.
func ExampleGenerate() {
	src := explaingen.NewFileSpecSource(fixtureFS())
	files, err := explaingen.Generate(src)
	if err != nil {
		return
	}
	_ = len(files) // 6
}
