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

// markerMock is the +glacier:mock marker string as it appears in doc comments.
const markerMock = "+glacier:mock"

// loadMode is the exact set of go/packages load mode flags needed for discovery.
const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedImports |
	packages.NeedTypes |
	packages.NeedSyntax

// DiscoveredInterface represents an interface type annotated with +glacier:mock.
type DiscoveredInterface struct {
	// PkgPath is the full Go import path of the package (e.g. "github.com/example/app/store").
	PkgPath string
	// PkgDir is the absolute filesystem directory of the package.
	PkgDir string
	// PkgName is the Go package name (the identifier after "package").
	PkgName string
	// TypeName is the unqualified interface type name (e.g. "Repo").
	TypeName string
	// Methods is the ordered list of methods on the interface.
	Methods []DiscoveredMethod
}

// DiscoveredMethod describes one method on a mocked interface.
type DiscoveredMethod struct {
	// Name is the exported method name.
	Name string
	// Params holds the parameter types in order.
	Params []DiscoveredParam
	// Results holds the result types in order.
	Results []DiscoveredResult
}

// DiscoveredParam is one parameter position.
type DiscoveredParam struct {
	// TypeExpr is the Go source expression for the type (e.g. "context.Context", "string").
	TypeExpr string
}

// DiscoveredResult is one result position.
type DiscoveredResult struct {
	// TypeExpr is the Go source expression for the type.
	TypeExpr string
}

// Discoverer discovers interfaces annotated with +glacier:mock in a set of Go
// packages. The production implementation uses go/packages; tests may inject a
// fake to avoid network and filesystem I/O.
//
// +glacier:mock
type Discoverer interface {
	// Discover returns all interfaces annotated with +glacier:mock in the
	// packages matching pattern. Only packages sharing the given module prefix
	// are considered; modulePrefix may be empty to disable the filter.
	Discover(pattern, modulePrefix string, logger *slog.Logger) ([]DiscoveredInterface, error)
}

// pkgDiscoverer is the real go/packages-backed Discoverer implementation.
type pkgDiscoverer struct{}

// Discover implements Discoverer using go/packages to load the requested
// pattern and walk every interface declaration carrying a +glacier:mock
// marker.
func (pkgDiscoverer) Discover(pattern, modulePrefix string, logger *slog.Logger) ([]DiscoveredInterface, error) {
	cfg := &packages.Config{
		Mode: loadMode,
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, fmt.Errorf("mockgen: load packages %q: %w", pattern, err)
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("mockgen: package %q has errors: %v", pkg.PkgPath, pkg.Errors[0])
		}
	}

	var results []DiscoveredInterface
	for _, pkg := range pkgs {
		if modulePrefix != "" && !strings.HasPrefix(pkg.PkgPath, modulePrefix) {
			continue
		}
		if pkg.Types == nil {
			continue
		}
		pkgDir := pkgDirOf(pkg)
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
			iface, ok := named.Underlying().(*types.Interface)
			if !ok {
				continue
			}
			if !hasMarker(pkg, name, markerMock) {
				continue
			}
			// Skip generic interfaces. Mock generation for type-parameterized
			// interfaces is not supported in v0; the emitter would need to
			// preserve type-param declarations and generic instantiations.
			// Document via spec amendment if a real use case appears.
			if named.TypeParams() != nil && named.TypeParams().Len() > 0 {
				if logger != nil {
					// Debug-level: this is an expected v0 limitation, not a problem.
					logger.Debug("mockgen: skipping generic interface (type parameters not supported)",
						"type", name, "pkg", pkg.PkgPath)
				}
				continue
			}
			methods := extractMethods(iface)
			results = append(results, DiscoveredInterface{
				PkgPath:  pkg.PkgPath,
				PkgDir:   pkgDir,
				PkgName:  pkg.Name,
				TypeName: name,
				Methods:  methods,
			})
		}
	}
	return results, nil
}

// pkgDirOf returns the directory of the first Go file in pkg, or "".
func pkgDirOf(pkg *packages.Package) string {
	if len(pkg.GoFiles) == 0 {
		return ""
	}
	f := pkg.GoFiles[0]
	// Find directory: trim everything after the last separator.
	if idx := strings.LastIndexAny(f, `/\`); idx >= 0 {
		return f[:idx+1]
	}
	return "."
}

// hasMarker reports whether the named type in pkg has a doc comment containing
// the given marker string on its own line.
func hasMarker(pkg *packages.Package, typeName, marker string) bool {
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
				// Check GenDecl doc and TypeSpec doc.
				for _, cg := range []*ast.CommentGroup{gd.Doc, ts.Doc} {
					if cg == nil {
						continue
					}
					for _, c := range cg.List {
						text := strings.TrimPrefix(c.Text, "// ")
						text = strings.TrimSpace(text)
						if text == marker {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// extractMethods returns the methods of iface in alphabetical order by name.
func extractMethods(iface *types.Interface) []DiscoveredMethod {
	n := iface.NumMethods()
	methods := make([]DiscoveredMethod, 0, n)
	for i := range n {
		fn := iface.Method(i)
		sig, ok := fn.Type().(*types.Signature)
		if !ok {
			continue
		}
		methods = append(methods, DiscoveredMethod{
			Name:    fn.Name(),
			Params:  extractParams(sig.Params()),
			Results: extractResults(sig.Results()),
		})
	}
	return methods
}

// extractParams converts a *types.Tuple into []DiscoveredParam.
func extractParams(tuple *types.Tuple) []DiscoveredParam {
	if tuple == nil {
		return nil
	}
	params := make([]DiscoveredParam, tuple.Len())
	for i := range tuple.Len() {
		params[i] = DiscoveredParam{TypeExpr: types.TypeString(tuple.At(i).Type(), nil)}
	}
	return params
}

// extractResults converts a *types.Tuple into []DiscoveredResult.
func extractResults(tuple *types.Tuple) []DiscoveredResult {
	if tuple == nil {
		return nil
	}
	results := make([]DiscoveredResult, tuple.Len())
	for i := range tuple.Len() {
		results[i] = DiscoveredResult{TypeExpr: types.TypeString(tuple.At(i).Type(), nil)}
	}
	return results
}
