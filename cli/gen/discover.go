// SPDX-License-Identifier: Apache-2.0

package gen

import (
	"fmt"
	"go/ast"
	"go/types"
	"log/slog"
	"strings"

	"golang.org/x/tools/go/packages"
)

// loadMode is the exact set of load mode flags used by discover.
// Test row 129 asserts this constant.
const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedImports |
	packages.NeedTypes |
	packages.NeedSyntax

// DiscoveredCommand represents a struct type that implements cli.Command.
type DiscoveredCommand struct {
	PkgPath   string // e.g. "github.com/example/app"
	PkgDir    string // absolute directory of the package
	TypeName  string // e.g. "ServeCmd"
	ParseResult *ParseResult
}

// discoverCommands walks packages matching pattern and returns all types that
// implement Run(ctx context.Context) error. Only packages within the same
// module (sharing modulePrefix) are considered.
func discoverCommands(pattern, modulePrefix string, logger *slog.Logger) ([]DiscoveredCommand, error) {
	cfg := &packages.Config{
		Mode: loadMode,
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, fmt.Errorf("gen: load packages %q: %w", pattern, err)
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("gen: package %q has errors: %v", pkg.PkgPath, pkg.Errors[0])
		}
	}

	var results []DiscoveredCommand

	for _, pkg := range pkgs {
		// Enforce module-prefix: only include packages from the same module.
		if modulePrefix != "" && !strings.HasPrefix(pkg.PkgPath, modulePrefix) {
			continue
		}

		pkgDir := ""
		if len(pkg.GoFiles) > 0 {
			// Use the directory of the first Go file.
			f := pkg.GoFiles[0]
			pkgDir = f[:strings.LastIndex(f, "/")+1]
			if pkgDir == "" {
				pkgDir = f[:strings.LastIndexByte(f, '\\')+1]
			}
		}

		if pkg.Types == nil {
			continue
		}

		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if obj == nil {
				continue
			}
			tn, ok := obj.(*types.TypeName)
			if !ok {
				continue
			}
			named, ok := tn.Type().(*types.Named)
			if !ok {
				continue
			}
			// Check if the type (or its pointer) has Run(context.Context) error.
			if !implementsCommand(named) && !implementsCommand(types.NewPointer(named)) {
				continue
			}

			// Parse markers from the type's doc comment.
			docLines := extractDocLines(pkg, name)
			pr, errs := ParseMarkers(name, docLines)
			for _, e := range errs {
				if logger != nil {
					logger.Warn("marker parse error", "type", name, "error", e)
				}
			}

			// Skip types that have no +glacier:command or +glacier:root marker.
			// Types that merely happen to implement Run(ctx context.Context) error
			// (e.g. term.Animator, test helpers) are not commands and must not
			// generate a registration entry.
			if pr.Cmd.Name == "" && !pr.IsRoot {
				continue
			}

			// Also parse field-level markers.
			parseFieldMarkers(pkg, named, pr)

			results = append(results, DiscoveredCommand{
				PkgPath:     pkg.PkgPath,
				PkgDir:      pkgDir,
				TypeName:    name,
				ParseResult: pr,
			})
		}
	}

	return results, nil
}

// implementsCommand checks whether t has method Run(ctx context.Context) error.
func implementsCommand(t types.Type) bool {
	ms := types.NewMethodSet(t)
	sel := ms.Lookup(nil, "Run")
	if sel == nil {
		return false
	}
	fn, ok := sel.Obj().(*types.Func)
	if !ok {
		return false
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return false
	}
	// Params: (context.Context)
	if sig.Params().Len() != 1 {
		return false
	}
	param := sig.Params().At(0)
	paramType := param.Type()
	// Check that the param implements context.Context (an interface).
	// We do a name-based check for simplicity.
	if !isContextType(paramType) {
		return false
	}
	// Results: (error)
	if sig.Results().Len() != 1 {
		return false
	}
	result := sig.Results().At(0)
	return isErrorType(result.Type())
}

// isContextType checks whether t is context.Context.
func isContextType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		// Could be an interface type directly.
		if iface, ok2 := t.Underlying().(*types.Interface); ok2 {
			return iface.NumMethods() >= 4 // rough heuristic for context.Context
		}
		return false
	}
	obj := named.Obj()
	return obj.Name() == "Context" && obj.Pkg() != nil && obj.Pkg().Path() == "context"
}

// isErrorType checks whether t is the error interface.
func isErrorType(t types.Type) bool {
	if named, ok := t.(*types.Named); ok {
		return named.Obj().Name() == "error"
	}
	// error is a built-in interface, not a named type.
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return iface.NumMethods() == 1 && iface.Method(0).Name() == "Error"
}

// extractDocLines extracts doc comment lines for a named type in the package.
func extractDocLines(pkg *packages.Package, typeName string) []string {
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if ts.Name.Name != typeName {
					continue
				}
				var lines []string
				if gd.Doc != nil {
					for _, c := range gd.Doc.List {
						text := strings.TrimPrefix(c.Text, "// ")
						text = strings.TrimPrefix(text, "//")
						lines = append(lines, text)
					}
				}
				if ts.Doc != nil {
					for _, c := range ts.Doc.List {
						text := strings.TrimPrefix(c.Text, "// ")
						text = strings.TrimPrefix(text, "//")
						lines = append(lines, text)
					}
				}
				return lines
			}
		}
	}
	return nil
}

// parseFieldMarkers extracts per-field +glacier: markers from the struct type.
func parseFieldMarkers(pkg *packages.Package, named *types.Named, pr *ParseResult) {
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return
	}
	for i := range st.NumFields() {
		field := st.Field(i)
		if !field.Exported() {
			continue
		}
		// Find the field in the AST to get doc comments.
		docLines := extractFieldDocLines(pkg, named.Obj().Name(), field.Name())
		if len(docLines) == 0 {
			continue
		}
		fm, errs := ParseFieldMarkers(field.Name(), docLines)
		for _, e := range errs {
			_ = e // caller handles via logger
		}
		if hasAnyMarker(fm) {
			pr.Fields[field.Name()] = fm
		}
	}
}

// hasAnyMarker returns true if fm has any marker or description set.
func hasAnyMarker(fm *FieldMarker) bool {
	return fm.HasShort || fm.Env != "" || fm.Required || len(fm.Choices) > 0 ||
		fm.HasDepr || fm.Validate != "" || fm.HasDefault || fm.Help != ""
}

// extractFieldDocLines extracts doc comment lines for a field of a named struct.
func extractFieldDocLines(pkg *packages.Package, typeName, fieldName string) []string {
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name.Name != typeName {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range st.Fields.List {
					for _, name := range field.Names {
						if name.Name != fieldName {
							continue
						}
						if field.Doc == nil {
							return nil
						}
						var lines []string
						for _, c := range field.Doc.List {
							text := strings.TrimPrefix(c.Text, "// ")
							text = strings.TrimPrefix(text, "//")
							lines = append(lines, text)
						}
						return lines
					}
				}
			}
		}
	}
	return nil
}
