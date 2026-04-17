// Package examplecheck enforces REN-127's acceptance rule that every
// exported identifier in format/, theme/, and component/ is covered by
// at least one godoc Example* function — so pkg.go.dev renders a usage
// block for each symbol downstream consumers depend on.
//
// The rule is a test, not a shell script, so it parses with go/parser
// alone and adds no new module dependencies.
package examplecheck

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// themeStyleConstructors are the theme/ style-constructor functions that
// the Sub-3 plan deliberately exempts: one representative ExampleHeader
// stands in for the family, and the package-level Example sweeps the
// rest of the style/palette/map surface. Keep the rationale here so a
// future maintainer does not "fix" the gap by demanding an example per
// constructor (which would bloat godoc without adding coverage).
var themeStyleConstructors = map[string]bool{
	"Header":           true,
	"StatLabel":        true,
	"StatValue":        true,
	"StatValueAccent":  true,
	"StatValueTeal":    true,
	"TableHeader":      true,
	"TableRow":         true,
	"TableRowSelected": true,
	"Muted":            true,
	"Dimmed":           true,
	"HelpBar":          true,
	"HelpKey":          true,
	"HelpDesc":         true,
	"CardBorder":       true,
	"SectionTitle":     true,
	"SpinnerStyle":     true,
	"ErrorText":        true,
	"TabActive":        true,
	"TabInactive":      true,
	"TabDisabled":      true,
	"TabBar":           true,
	"TabSeparator":     true,
	"LogFollow":        true,
	"LogPaused":        true,
	"LogFooterRow":     true,
	"LogBody":          true,
}

// themePackageLevelCovered are exported vars/maps covered by the
// package-level theme.Example rather than a dedicated ExampleName.
var themePackageLevelCovered = map[string]bool{
	"BgPrimary":           true,
	"BgSecondary":         true,
	"BgTertiary":          true,
	"Surface":             true,
	"SurfaceRaised":       true,
	"SurfaceBorder":       true,
	"SurfaceBorderBright": true,
	"Accent":              true,
	"AccentDim":           true,
	"Teal":                true,
	"TealDim":             true,
	"Blue":                true,
	"StatusSuccess":       true,
	"StatusWarning":       true,
	"StatusError":         true,
	"TextPrimary":         true,
	"TextSecondary":       true,
	"TextTertiary":        true,
	"ActivityColors":      true,
	"ActivityIcons":       true,
}

func TestExportedSymbolsHaveExamples(t *testing.T) {
	t.Parallel()

	repoRoot := repoRoot(t)

	cases := []struct {
		pkgName string
		dir     string
	}{
		{"format", filepath.Join(repoRoot, "format")},
		{"theme", filepath.Join(repoRoot, "theme")},
		{"component", filepath.Join(repoRoot, "component")},
	}

	for _, tc := range cases {
		t.Run(tc.pkgName, func(t *testing.T) {
			exports, examples, hasPackageExample := parsePackage(t, tc.dir)

			var missing []string
			for _, name := range exports {
				if covered(tc.pkgName, name, examples, hasPackageExample) {
					continue
				}
				missing = append(missing, name)
			}

			if len(missing) > 0 {
				sort.Strings(missing)
				for _, name := range missing {
					t.Errorf("%s.%s: no Example%s found (and not covered by package-level Example)", tc.pkgName, name, name)
				}
			}
		})
	}
}

func covered(pkg, name string, examples map[string]bool, hasPackageExample bool) bool {
	if examples["Example"+name] {
		return true
	}
	// Example<Name>_variant also counts as covering <Name>.
	for ex := range examples {
		if strings.HasPrefix(ex, "Example"+name+"_") {
			return true
		}
	}
	if pkg == "theme" {
		if themeStyleConstructors[name] && examples["ExampleHeader"] {
			return true
		}
		if themePackageLevelCovered[name] && hasPackageExample {
			return true
		}
	}
	// component/ exports a single interface, Component, whose canonical
	// godoc entry is the package-level Example implementing it (Sub-4
	// deliberately did not add a separate ExampleComponent).
	if pkg == "component" && name == "Component" && hasPackageExample {
		return true
	}
	return false
}

// parsePackage reads every .go file in dir, returning exported top-level
// symbol names (types, funcs, vars, consts), the set of Example*
// function names, and whether the package defines a bare `Example`.
func parsePackage(t *testing.T, dir string) (exports []string, examples map[string]bool, hasPackageExample bool) {
	t.Helper()

	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}

	examples = map[string]bool{}
	seen := map[string]bool{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		isTest := strings.HasSuffix(entry.Name(), "_test.go")
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				name := d.Name.Name
				if isTest {
					if strings.HasPrefix(name, "Example") {
						examples[name] = true
						if name == "Example" {
							hasPackageExample = true
						}
					}
					continue
				}
				if d.Recv != nil {
					continue
				}
				if !seen[name] {
					seen[name] = true
					exports = append(exports, name)
				}
			case *ast.GenDecl:
				if isTest {
					continue
				}
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() && !seen[s.Name.Name] {
							seen[s.Name.Name] = true
							exports = append(exports, s.Name.Name)
						}
					case *ast.ValueSpec:
						for _, n := range s.Names {
							if n.IsExported() && !seen[n.Name] {
								seen[n.Name] = true
								exports = append(exports, n.Name)
							}
						}
					}
				}
			}
		}
	}

	sort.Strings(exports)
	return exports, examples, hasPackageExample
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// This file: <root>/internal/examplecheck/examplecheck_test.go
	return filepath.Join(filepath.Dir(file), "..", "..")
}
