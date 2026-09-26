package template

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Obligation is one row of the dual-harness obligation registry
// (SPEC-DUAL-HARNESS-HOOK-PARITY-001 REQ-HPR-020, design.md §D4): a required
// behavior, where each harness applies it, and the test that verifies it.
type Obligation struct {
	ID       string
	Required bool
	// ClaudePath and CodexPath are application paths (a rendered artifact or
	// a registered subcommand) or a marker: "UNSUPPORTED:<evidence-ref>",
	// "blocked:<reason>", or "unverified:<reason>".
	ClaudePath string
	CodexPath  string
	// Check is the Go test function that verifies the obligation.
	Check string
	// AC is the acceptance criterion the obligation answers to.
	AC string
}

// ObligationRegistry is the parsed registry file.
type ObligationRegistry struct {
	Obligations []Obligation
}

// PathMarker classifies an application path.
type PathMarker string

const (
	// MarkerNone: a concrete application path.
	MarkerNone PathMarker = ""
	// MarkerUnsupported: the host cannot express the obligation; the value is
	// the evidence reference.
	MarkerUnsupported PathMarker = "UNSUPPORTED"
	// MarkerBlocked: the path depends on unfinished work; the value names it.
	MarkerBlocked PathMarker = "blocked"
	// MarkerUnverified: the path exists but is not verified; the value says why.
	MarkerUnverified PathMarker = "unverified"
)

var pathMarkers = []PathMarker{MarkerUnsupported, MarkerBlocked, MarkerUnverified}

// ParseApplicationPath splits an application path into its marker and value.
// A marker must carry a non-empty reference: a bare "blocked:" would let a row
// claim a known gap without saying what it is.
func ParseApplicationPath(s string) (PathMarker, string, error) {
	for _, m := range pathMarkers {
		prefix := string(m) + ":"
		if strings.HasPrefix(s, prefix) {
			ref := strings.TrimSpace(strings.TrimPrefix(s, prefix))
			if ref == "" {
				return m, "", fmt.Errorf("marker %s carries no reference", m)
			}
			return m, ref, nil
		}
	}
	return MarkerNone, s, nil
}

// obligationRow is the on-disk row shape. Required is a pointer so an absent
// key is an error rather than a silent false.
type obligationRow struct {
	ID         string `yaml:"id"`
	Required   *bool  `yaml:"required"`
	ClaudePath string `yaml:"claude_path"`
	CodexPath  string `yaml:"codex_path"`
	Check      string `yaml:"check"`
	AC         string `yaml:"ac"`
}

// ParseObligations decodes and validates a registry document. It refuses
// unknown keys, a row without an id, a row without an explicit required flag
// or acceptance id, a duplicate id, and a marker without a reference. Whether
// every required row is covered is CheckObligationCoverage's job, so a row
// missing a path or a check still loads and is then reported by id.
func ParseObligations(data []byte) (*ObligationRegistry, error) {
	var doc struct {
		Obligations []obligationRow `yaml:"obligations"`
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("obligation registry: %w", err)
	}
	reg := &ObligationRegistry{}
	seen := map[string]bool{}
	for i, r := range doc.Obligations {
		if strings.TrimSpace(r.ID) == "" {
			return nil, fmt.Errorf("obligation registry: row %d has no id", i)
		}
		if seen[r.ID] {
			return nil, fmt.Errorf("obligation registry: duplicate id %q", r.ID)
		}
		seen[r.ID] = true
		if r.Required == nil {
			return nil, fmt.Errorf("obligation registry: %s: required is not set", r.ID)
		}
		if strings.TrimSpace(r.AC) == "" {
			return nil, fmt.Errorf("obligation registry: %s: ac is not set", r.ID)
		}
		for field, p := range map[string]string{"claude_path": r.ClaudePath, "codex_path": r.CodexPath} {
			if _, _, err := ParseApplicationPath(p); err != nil {
				return nil, fmt.Errorf("obligation registry: %s: %s: %w", r.ID, field, err)
			}
		}
		reg.Obligations = append(reg.Obligations, Obligation{
			ID: r.ID, Required: *r.Required, ClaudePath: r.ClaudePath, CodexPath: r.CodexPath, Check: r.Check, AC: r.AC,
		})
	}
	return reg, nil
}

// CoverageResolver answers the two lookups the coverage check needs.
type CoverageResolver interface {
	// TestExists reports whether a test function with this name exists.
	TestExists(name string) bool
	// PathExists reports whether a concrete application path resolves for
	// the harness ("claude" or "codex").
	PathExists(harness, path string) bool
}

// CoverageViolation names one obligation and what it lacks.
type CoverageViolation struct {
	ID      string
	Problem string
}

func (v CoverageViolation) String() string { return v.ID + ": " + v.Problem }

// @MX:NOTE: [AUTO] obligation coverage predicate (AC-HPR-018); a required row passing here without a path, marker, or resolvable check widens the parity claim silently
// @MX:SPEC: SPEC-DUAL-HARNESS-HOOK-PARITY-001
// CheckObligationCoverage reports every required obligation that lacks a
// Claude path, a Codex path (a marker counts as a path), or a check that
// resolves to an existing test; and every concrete path that does not
// resolve. An empty registry is itself a violation, so the check cannot pass
// vacuously (REQ-HPR-021).
func CheckObligationCoverage(reg *ObligationRegistry, r CoverageResolver) []CoverageViolation {
	if reg == nil || len(reg.Obligations) == 0 {
		return []CoverageViolation{{ID: "(registry)", Problem: "the registry declares no obligations"}}
	}
	var out []CoverageViolation
	for _, o := range reg.Obligations {
		if !o.Required {
			continue
		}
		for _, side := range []struct{ harness, field, path string }{
			{"claude", "claude_path", o.ClaudePath},
			{"codex", "codex_path", o.CodexPath},
		} {
			if strings.TrimSpace(side.path) == "" {
				out = append(out, CoverageViolation{o.ID, side.field + " is empty (give a path or an UNSUPPORTED/blocked/unverified marker)"})
				continue
			}
			marker, value, err := ParseApplicationPath(side.path)
			if err != nil {
				out = append(out, CoverageViolation{o.ID, side.field + ": " + err.Error()})
				continue
			}
			if marker == MarkerNone && !r.PathExists(side.harness, value) {
				out = append(out, CoverageViolation{o.ID, side.field + " " + value + " does not resolve"})
			}
		}
		switch {
		case strings.TrimSpace(o.Check) == "":
			out = append(out, CoverageViolation{o.ID, "check is empty"})
		case !r.TestExists(o.Check):
			out = append(out, CoverageViolation{o.ID, "check " + o.Check + " names no existing test"})
		}
	}
	return out
}

// TestIndex is a static set of test function names.
type TestIndex map[string]bool

// TestExists reports whether name was indexed.
func (idx TestIndex) TestExists(name string) bool { return idx[name] }

// IndexTestFunctions parses every _test.go file under root and indexes the
// top-level test functions (func TestXxx(t *testing.T)). It never runs a test:
// the coverage check needs existence, and running would put a test suite
// inside a hook-adjacent check. Hidden directories, testdata, and vendor are
// skipped.
func IndexTestFunctions(root string) (TestIndex, error) {
	idx := TestIndex{}
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (strings.HasPrefix(name, ".") || name == "testdata" || name == "vendor" || name == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if perr != nil {
			return fmt.Errorf("index tests: %w", perr)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") {
				continue
			}
			if isTestingTParam(fn) {
				idx[fn.Name.Name] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return idx, nil
}

// isTestingTParam reports whether fn takes exactly one *testing.T parameter.
func isTestingTParam(fn *ast.FuncDecl) bool {
	params := fn.Type.Params.List
	if len(params) != 1 || len(params[0].Names) > 1 {
		return false
	}
	star, ok := params[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "testing" && sel.Sel.Name == "T"
}
