package cli

// migrate_ownership_gate_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M10 (card
// t1865), AC-OWR-19 (REQ-OWR-014): the removal switch stays off in production.
//
// The switch is the package variable removalEnabled (migrate_project_assets.go).
// The only way production code could turn it on — or hand it to a verdict
// environment turned on — is an assignment, so the scan below reads the syntax
// of every non-test file of the package and reports every statement that could
// change it:
//
//   - an assignment (=, :=, op=) or a range clause that names the variable;
//   - an assignment to a field named removalEnabled (the verdict environment's
//     copy of the switch);
//   - a var declaration of it other than the one package-level declaration whose
//     value is the literal false;
//   - taking its address;
//   - a composite-literal field removalEnabled whose value is anything but a read
//     of the package variable itself.
//
// What the scan cannot see (stated, not hidden): a write through reflection or
// unsafe, a go:linkname, and a switch turned on by a process outside this
// package. Config files and environment variables cannot reach the variable
// without one of the statements above.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const removalSwitchName = "removalEnabled"

// switchScan is what the scan found in a set of sources.
type switchScan struct {
	files        int
	violations   []string
	declarations int // package-level `var removalEnabled = false`
	mentions     int // mentions of the package variable that are not declarations or keyed field names
	keyedCopies  int // composite-literal fields handed the package variable
}

// scanRemovalSwitch parses each source (file name -> text) and records every
// statement that could change the switch.
func scanRemovalSwitch(sources map[string]string) (switchScan, error) {
	var res switchScan
	names := make([]string, 0, len(sources))
	for name := range sources {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, name, sources[name], parser.SkipObjectResolution)
		if err != nil {
			return res, fmt.Errorf("parse %s: %w", name, err)
		}
		res.files++
		scanRemovalSwitchFile(fset, f, &res)
	}
	return res, nil
}

func isSwitchIdent(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == removalSwitchName
}

func isSwitchSelector(e ast.Expr) bool {
	sel, ok := e.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == removalSwitchName
}

func scanRemovalSwitchFile(fset *token.FileSet, f *ast.File, res *switchScan) {
	flag := func(pos token.Pos, what string) {
		res.violations = append(res.violations, fmt.Sprintf("%s: %s", fset.Position(pos), what))
	}
	// The one allowed declaration: a package-level var with a single value, false.
	allowedDecl := map[*ast.ValueSpec]bool{}
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs := spec.(*ast.ValueSpec)
			if len(vs.Names) == 1 && vs.Names[0].Name == removalSwitchName && len(vs.Values) == 1 {
				if id, isIdent := vs.Values[0].(*ast.Ident); isIdent && id.Name == "false" {
					allowedDecl[vs] = true
					res.declarations++
				}
			}
		}
	}
	// Idents that sit in a position already judged (a declaration name, a keyed
	// field name, a selector's field) are not mentions of the variable.
	judged := map[*ast.Ident]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ValueSpec:
			for _, id := range x.Names {
				if id.Name == removalSwitchName {
					judged[id] = true
					if !allowedDecl[x] {
						flag(id.Pos(), "a declaration of "+removalSwitchName+" other than the package-level `= false`")
					}
				}
			}
		case *ast.AssignStmt:
			for _, lhs := range x.Lhs {
				if isSwitchIdent(lhs) || isSwitchSelector(lhs) {
					flag(lhs.Pos(), "an assignment to "+removalSwitchName)
				}
			}
		case *ast.RangeStmt:
			for _, e := range []ast.Expr{x.Key, x.Value} {
				if e != nil && (isSwitchIdent(e) || isSwitchSelector(e)) {
					flag(e.Pos(), "a range clause that assigns "+removalSwitchName)
				}
			}
		case *ast.IncDecStmt:
			if isSwitchIdent(x.X) || isSwitchSelector(x.X) {
				flag(x.Pos(), "an increment or decrement of "+removalSwitchName)
			}
		case *ast.UnaryExpr:
			if x.Op == token.AND && (isSwitchIdent(x.X) || isSwitchSelector(x.X)) {
				flag(x.Pos(), "the address of "+removalSwitchName+" taken")
			}
		case *ast.SelectorExpr:
			if x.Sel.Name == removalSwitchName {
				judged[x.Sel] = true
			}
		case *ast.KeyValueExpr:
			if key, ok := x.Key.(*ast.Ident); ok && key.Name == removalSwitchName {
				judged[key] = true
				if isSwitchIdent(x.Value) {
					res.keyedCopies++
				} else {
					flag(x.Pos(), "a composite-literal field "+removalSwitchName+" that is not a read of the package variable")
				}
			}
		}
		return true
	})
	ast.Inspect(f, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == removalSwitchName && !judged[id] {
			res.mentions++
		}
		return true
	})
}

// gateSources reads every non-test Go file of the package.
func gateSources(t *testing.T) map[string]string {
	t.Helper()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list the package sources: %v", err)
	}
	out := map[string]string{}
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		data, readErr := os.ReadFile(p)
		if readErr != nil {
			t.Fatalf("read %s: %v", p, readErr)
		}
		out[p] = string(data)
	}
	return out
}

func TestRemovalSwitchStaysOffInProduction(t *testing.T) {
	t.Run("DefaultIsOff", func(t *testing.T) {
		if removalEnabled {
			t.Fatalf("%s is true at the start of the test: the shipped default must be false", removalSwitchName)
		}
	})

	t.Run("NoProductionStatementChangesTheSwitch", func(t *testing.T) {
		res, err := scanRemovalSwitch(gateSources(t))
		if err != nil {
			t.Fatal(err)
		}
		// Reach: the scan saw the package, its single declaration and its uses.
		if res.files < 100 {
			t.Fatalf("the scan read %d non-test files; the package holds far more, so the scan did not reach it", res.files)
		}
		if res.declarations != 1 {
			t.Errorf("found %d package-level `var %s = false` declarations, want exactly 1", res.declarations, removalSwitchName)
		}
		if res.mentions == 0 || res.keyedCopies == 0 {
			t.Errorf("the scan found no mention (%d) or keyed hand-over (%d) of the switch: it is not looking at the code that uses it", res.mentions, res.keyedCopies)
		}
		for _, v := range res.violations {
			t.Errorf("production code can change the removal switch: %s", v)
		}
		t.Logf("scanned %d non-test files: %d declaration, %d mentions, %d keyed hand-overs, %d violations", res.files, res.declarations, res.mentions, res.keyedCopies, len(res.violations))
	})

	// The scan must go red in front of each wrong implementation, and stay quiet
	// on the shape the shipped code has.
	t.Run("ScanMutants", func(t *testing.T) {
		clean := "package p\n\nvar removalEnabled = false\n\ntype env struct{ removalEnabled bool }\n\nfunc use() env { if removalEnabled { return env{} }; return env{removalEnabled: removalEnabled} }\n"
		res, err := scanRemovalSwitch(map[string]string{"clean.go": clean})
		if err != nil || len(res.violations) != 0 || res.declarations != 1 || res.keyedCopies != 1 {
			t.Fatalf("the shipped shape is not clean: %+v (%v)", res, err)
		}
		cases := []struct{ name, src string }{
			{"InitSetsItTrue", "package p\n\nvar removalEnabled = false\n\nfunc init() { removalEnabled = true }\n"},
			{"NonConstantAssignment", "package p\n\nvar removalEnabled = false\n\nfunc load(cfg struct{ Enabled bool }) { removalEnabled = cfg.Enabled }\n"},
			{"DeclarationDefaultTrue", "package p\n\nvar removalEnabled = true\n"},
			{"DeclarationFromConfig", "package p\n\nvar removalEnabled = readConfig()\n\nfunc readConfig() bool { return false }\n"},
			{"ShortDeclarationShadow", "package p\n\nvar removalEnabled = false\n\nfunc f() { removalEnabled := true; _ = removalEnabled }\n"},
			{"AddressTaken", "package p\n\nvar removalEnabled = false\n\nfunc f() { p := &removalEnabled; *p = true }\n"},
			{"EnvFieldAssigned", "package p\n\nvar removalEnabled = false\n\ntype env struct{ removalEnabled bool }\n\nfunc f(e *env) { e.removalEnabled = true }\n"},
			{"EnvFieldLiteralTrue", "package p\n\nvar removalEnabled = false\n\ntype env struct{ removalEnabled bool }\n\nfunc f() env { return env{removalEnabled: true} }\n"},
			{"EnvFieldLiteralFromConfig", "package p\n\nvar removalEnabled = false\n\ntype env struct{ removalEnabled bool }\n\nfunc f(on bool) env { return env{removalEnabled: on} }\n"},
			{"OpAssign", "package p\n\nvar removalEnabled = false\n\nfunc f() { removalEnabled = removalEnabled || true }\n"},
			{"RangeAssigns", "package p\n\nvar removalEnabled = false\n\nfunc f(xs []bool) { for _, removalEnabled = range xs {} }\n"},
			{"SecondPackageLevelDeclaration", "package p\n\nvar removalEnabled = false\nvar removalEnabled = true\n"},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				r, err := scanRemovalSwitch(map[string]string{"mutant.go": c.src})
				if err != nil {
					t.Fatalf("the mutant does not parse: %v", err)
				}
				if len(r.violations) == 0 {
					t.Errorf("the scan stayed green in front of the %s mutant", c.name)
				}
			})
		}
	})
}
