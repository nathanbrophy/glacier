// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	cligen "github.com/nathanbrophy/glacier/cli/gen"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/report"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/sdkerr"
	"github.com/nathanbrophy/glacier/internal/safefile"
)

// NewCmd is the parent for new <package|command|option>.
//
// +glacier:command name=new parent=glacier
type NewCmd struct{}

// Run shows help when invoked bare.
func (c *NewCmd) Run(_ context.Context) error {
	return nil
}

// NewPackageCmd creates a new Go package skeleton.
//
// +glacier:command name=package parent=new
type NewPackageCmd struct {
	// Name is the package directory name.
	//
	// +glacier:positional
	Name string

	// DryRun prints what would be created without writing files.
	//
	// +glacier:default false
	DryRun bool

	// Force overwrites existing files.
	//
	// +glacier:default false
	Force bool

	// Pkg overrides the Go package name (defaults to Name).
	Pkg string
}

// Run creates a new package skeleton under Name/.
func (c *NewPackageCmd) Run(ctx context.Context) error {
	report.Status(report.Calm, "glacier new package")

	if c.Name == "" {
		return &sdkerr.ErrInvalidName{Kind: "package", Name: "", Why: "name is required"}
	}

	pkgName := c.Pkg
	if pkgName == "" {
		pkgName = c.Name
	}

	// Validate name.
	if err := validateIdentifier("package", c.Name); err != nil {
		return err
	}

	// Find module root.
	root, err := findModuleRoot(".")
	if err != nil {
		wd, _ := os.Getwd()
		return &sdkerr.ErrNoModule{Cwd: wd}
	}

	files := map[string]string{
		filepath.Join(c.Name, "doc.go"):          fmt.Sprintf("// SPDX-License-Identifier: Apache-2.0\n\n// Package %s ...\npackage %s\n", pkgName, pkgName),
		filepath.Join(c.Name, c.Name+".go"):      fmt.Sprintf("// SPDX-License-Identifier: Apache-2.0\n\npackage %s\n", pkgName),
		filepath.Join(c.Name, c.Name+"_test.go"): fmt.Sprintf("// SPDX-License-Identifier: Apache-2.0\n\npackage %s_test\n\nimport \"testing\"\n\nfunc TestPlaceholder(t *testing.T) {\n\tt.Skip(\"placeholder: remove me\")\n}\n", pkgName),
	}

	if c.DryRun {
		report.Status(report.Thinking, "dry run: would create:")
		for rel := range files {
			fmt.Fprintf(os.Stdout, "  %s\n", filepath.Join(root, rel))
		}
		return nil
	}

	for rel, content := range files {
		relFwd := filepath.ToSlash(rel)
		if writeErr := safefile.WriteFileAtomic(root, relFwd, []byte(content), 0o644); writeErr != nil {
			if !c.Force {
				report.Status(report.Err, "write failed: "+writeErr.Error())
				return &exitCodeError{code: exitScaffoldFailed, cause: writeErr}
			}
		}
	}

	report.Status(report.Confident, "on solid ground.")
	return nil
}

// NewCommandCmd creates a new +glacier:command struct and re-runs cligen.
//
// +glacier:command name=command parent=new
type NewCommandCmd struct {
	// Name is the command name.
	//
	// +glacier:positional
	Name string

	// Parent is the parent command name.
	Parent string

	// DryRun prints what would be created without writing files.
	//
	// +glacier:default false
	DryRun bool

	// Force overwrites existing files.
	//
	// +glacier:default false
	Force bool
}

// Run creates a new command stub and regenerates the CLI wiring.
//
// The new file is written into the user's command directory, discovered by
// walking from the working directory up to the module root and looking for
// the directory containing zz_generated_cli.go. The package declaration is
// taken from a sibling .go file so the new struct slots into the existing
// command tree.
func (c *NewCommandCmd) Run(ctx context.Context) error {
	report.Status(report.Calm, "glacier new command")

	if c.Name == "" {
		return &sdkerr.ErrInvalidName{Kind: "command", Name: "", Why: "name is required"}
	}
	if err := validateIdentifier("command", c.Name); err != nil {
		return err
	}

	root, err := findModuleRoot(".")
	if err != nil {
		wd, _ := os.Getwd()
		return &sdkerr.ErrNoModule{Cwd: wd}
	}

	cmdDir, pkgName, err := findCommandDir(root)
	if err != nil {
		return &exitCodeError{code: exitScaffoldFailed, cause: err}
	}

	parent := c.Parent
	if parent == "" {
		parent = "root"
	}

	typeName := strings.ToUpper(c.Name[:1]) + c.Name[1:] + "Cmd"
	content := fmt.Sprintf(
		"// SPDX-License-Identifier: Apache-2.0\n\npackage %s\n\nimport \"context\"\n\n// %s implements the %s command.\n//\n// +glacier:command name=%s parent=%s\ntype %s struct{}\n\n// Run implements cli.Command.\nfunc (c *%s) Run(_ context.Context) error {\n\treturn nil\n}\n",
		pkgName, typeName, c.Name, c.Name, parent, typeName, typeName,
	)

	outFile := filepath.Join(cmdDir, strings.ToLower(c.Name)+".go")
	relOutFile, _ := filepath.Rel(root, outFile)

	if c.DryRun {
		report.Status(report.Thinking, "dry run: would create "+relOutFile)
		fmt.Fprintln(os.Stdout, content)
		return nil
	}

	if err := safefile.WriteFileAtomic(root, filepath.ToSlash(relOutFile), []byte(content), 0o644); err != nil {
		return &exitCodeError{code: exitScaffoldFailed, cause: err}
	}

	// Re-run codegen for the user's command tree.
	if err := cligen.Generate(cligen.Options{Pattern: "./..."}); err != nil {
		report.Status(report.Alarmed, "codegen after new command: "+err.Error())
	}

	report.Status(report.Confident, "on solid ground at "+relOutFile)
	return nil
}

// findCommandDir locates the user's command directory under root. It returns
// the absolute path of the directory that contains zz_generated_cli.go and
// the package name read from a sibling .go file. Returns an error if no such
// directory is found.
func findCommandDir(root string) (dir, pkgName string, err error) {
	var found string
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		// Skip vendor and version control directories.
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git" || info.Name() == "node_modules") {
			return filepath.SkipDir
		}
		if !info.IsDir() && info.Name() == "zz_generated_cli.go" {
			found = filepath.Dir(path)
			return filepath.SkipAll
		}
		return nil
	})
	if walkErr != nil {
		return "", "", fmt.Errorf("new command: walk %q: %w", root, walkErr)
	}
	if found == "" {
		return "", "", fmt.Errorf("new command: no zz_generated_cli.go found under %q (run cligen first)", root)
	}
	// Read package name from any .go file in the directory.
	entries, err := os.ReadDir(found)
	if err != nil {
		return "", "", fmt.Errorf("new command: read %q: %w", found, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || e.Name() == "zz_generated_cli.go" {
			continue
		}
		data, rerr := os.ReadFile(filepath.Join(found, e.Name()))
		if rerr != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "package ") {
				return found, strings.TrimSpace(strings.TrimPrefix(line, "package ")), nil
			}
		}
	}
	return "", "", fmt.Errorf("new command: could not determine package name in %q", found)
}

// NewOptionCmd appends a functional-option constructor in the chosen package.
//
// +glacier:command name=option parent=new
type NewOptionCmd struct {
	// TypeName is the option target type (e.g. ServerConfig).
	//
	// +glacier:positional
	TypeName string

	// Pkg is the package path where the option will be added (default: current).
	Pkg string

	// DryRun prints the generated code without writing it.
	//
	// +glacier:default false
	DryRun bool

	// Force overwrites existing files.
	//
	// +glacier:default false
	Force bool
}

// Run appends a new functional-option constructor to options.go in Pkg.
func (c *NewOptionCmd) Run(_ context.Context) error {
	report.Status(report.Calm, "glacier new option")

	if c.TypeName == "" {
		return &sdkerr.ErrInvalidName{Kind: "option", Name: "", Why: "TypeName is required"}
	}
	if err := validateIdentifier("option", c.TypeName); err != nil {
		return err
	}

	pkg := c.Pkg
	if pkg == "" {
		pkg = "."
	}

	// Resolve the package directory's actual Go package name from a sibling
	// .go file so the generated code compiles into the existing tree.
	pkgName, err := readPkgName(pkg)
	if err != nil {
		return &exitCodeError{code: exitScaffoldFailed, cause: err}
	}

	optionTypeName := c.TypeName + "Option"
	funcName := "With" + c.TypeName

	src := fmt.Sprintf(
		"// SPDX-License-Identifier: Apache-2.0\n\npackage %s\n\n// %s configures %s.\ntype %s func(*%s)\n\n// %s is a placeholder option constructor.\nfunc %s() %s {\n\treturn func(_ *%s) {}\n}\n",
		pkgName,
		optionTypeName, c.TypeName,
		optionTypeName, c.TypeName,
		funcName, funcName, optionTypeName, c.TypeName,
	)

	// Format with gofmt.
	formatted, err := format.Source([]byte(src))
	if err != nil {
		formatted = []byte(src)
	}

	if c.DryRun {
		report.Status(report.Thinking, "dry run: would append to "+pkg+"/options.go")
		fmt.Fprintln(os.Stdout, string(formatted))
		return nil
	}

	optionsFile := filepath.Join(pkg, "options.go")
	existing, readErr := os.ReadFile(optionsFile)
	var out []byte
	if readErr == nil {
		// Append after last closing brace.
		out = append(bytes.TrimRight(existing, "\n"), '\n')
		out = append(out, '\n')
		// Only append the type and func, not the package declaration.
		lines := strings.Split(string(formatted), "\n")
		var body []string
		inPkg := false
		for _, l := range lines {
			if strings.HasPrefix(l, "package ") {
				inPkg = true
				continue
			}
			if inPkg {
				body = append(body, l)
			}
		}
		out = append(out, []byte(strings.Join(body, "\n"))...)
	} else {
		out = formatted
	}

	if writeErr := os.WriteFile(optionsFile, out, 0o644); writeErr != nil {
		return &exitCodeError{code: exitScaffoldFailed, cause: writeErr}
	}

	report.Status(report.Confident, "on solid ground.")
	return nil
}

// validateIdentifier returns an ErrInvalidName if s is not a valid Go identifier.
func validateIdentifier(kind, s string) error {
	if s == "" {
		return &sdkerr.ErrInvalidName{Kind: kind, Name: s, Why: "must not be empty"}
	}
	for i, r := range s {
		if i == 0 && !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_') {
			return &sdkerr.ErrInvalidName{Kind: kind, Name: s, Why: "must start with a letter or underscore"}
		}
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return &sdkerr.ErrInvalidName{Kind: kind, Name: s, Why: fmt.Sprintf("invalid character %q at position %d", r, i)}
		}
	}
	return nil
}

// readPkgName returns the Go package name declared in any .go file in dir.
// Returns the directory's basename as a fallback when no .go files exist.
func readPkgName(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("new option: read %q: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		// Skip generated files; their package is correct but we prefer source files.
		if strings.HasPrefix(e.Name(), "zz_generated_") {
			continue
		}
		data, rerr := os.ReadFile(filepath.Join(dir, e.Name()))
		if rerr != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "package ") {
				name := strings.TrimSpace(strings.TrimPrefix(line, "package "))
				// Strip any trailing comment.
				if idx := strings.Index(name, "//"); idx >= 0 {
					name = strings.TrimSpace(name[:idx])
				}
				return name, nil
			}
		}
	}
	// Fallback: use the directory's basename as a package name guess.
	return filepath.Base(dir), nil
}

// findModuleRoot walks parent directories to find the nearest go.mod.
func findModuleRoot(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(abs, "go.mod")); err == nil {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", fmt.Errorf("new: no go.mod found")
		}
		abs = parent
	}
}
