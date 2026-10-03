package template

// SPEC-TPL-AST-GUARD-001 — AST-based workflow.worktree.* key-honesty guard.
//
// Re-implements PR #1707's bidirectional reader contract on true go/types
// resolution: for every documented workflow.worktree.* key, the set of
// production files reading the matching WorkflowWorktreeConfig field must
// exactly equal the expectation table below, and the table itself must stay
// complete against the live struct and honest about reserved keys.
//
// Reads are type-resolved (TypesInfo.Selections[*ast.SelectorExpr].Obj()),
// never text-matched: the scan sees through local alias copies, method bodies
// and multi-line expressions by construction — the defect class the retained
// t682 text guard (internal/config/workflow_key_honesty_test.go, unchanged)
// cannot see.
//
// Read definition (REQ-002, by exclusion): every type-resolved occurrence of
// a WorkflowWorktreeConfig field is a read EXCEPT the field selector appearing
// as the direct left-hand operand of a plain `=` assignment. Compound
// assignments (op=), ++/-- operands and &x.F address-taking are reads.
//
// @MX:NOTE the scanner mirrors the load shape of the in-repo precedent
// internal/config/shipped_key_reader_test.go (packages.Config without the
// Tests flag, ./... pattern, Selections route); extraction to a shared helper
// is deliberately deferred (plan D2) until a second consumer exists.

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

const (
	// worktreeConfigPkgPath is the import path of the package that owns
	// WorkflowWorktreeConfig.
	worktreeConfigPkgPath = "github.com/modu-ai/moai-adk/internal/config"
	// worktreeConfigTypeName is the tracked struct.
	worktreeConfigTypeName = "WorkflowWorktreeConfig"
)

// expectedWorktreeReaders is the expectation table (contract A): for each
// shipped template key, the complete set of production files that read the
// matching config field. It mirrors the reader-status doc comment on
// WorkflowWorktreeConfig (internal/config/types.go) — when the reader map
// moves, update that comment in the same commit (plan §G maintenance rule).
var expectedWorktreeReaders = map[string][]string{
	"auto_create":          {"internal/cli/worktree_advisory.go"},
	"auto_merge":           {"internal/cli/session_worktree_automerge.go"},
	"auto_cleanup":         {"internal/cli/session_worktree.go", "internal/cli/session_worktree_prmerge.go"},
	"session_name_pattern": {},
	"tmux_preferred":       {},
}

// reservedWorktreeKeys declares the keys whose reader status is "reserved: no
// production reader" per the WorkflowWorktreeConfig doc comment and the
// shipped workflow.yaml template. Declared independently of the expectation
// table so that a table naming a file under a reserved key is itself invalid
// (REQ-005) rather than silently re-classifying the key as read.
var reservedWorktreeKeys = map[string]bool{
	"session_name_pattern": true,
	"tmux_preferred":       true,
}

// mandatoryAutoCleanupReaders names the two load-bearing auto-cleanup sites
// that must stay in the auto_cleanup expectation entry (REQ-004). The check
// evaluates the table itself, before and without consulting the computed
// scan, so a same-commit removal of a read and of its table entry still
// fails.
var mandatoryAutoCleanupReaders = []string{
	"internal/cli/session_worktree.go",
	"internal/cli/session_worktree_prmerge.go",
}

// worktreeField carries one live field of WorkflowWorktreeConfig.
type worktreeField struct {
	name    string // Go field name, e.g. AutoCleanup
	yamlKey string // yaml tag key, e.g. auto_cleanup
	obj     *types.Var
}

// readerIndex is the computed reader map of one scan: field name → set of
// repo-relative files carrying a type-resolved read.
type readerIndex struct {
	readers      map[string]map[string]bool
	pkgsScanned  int
	filesScanned int
	typeErrors   []string
}

// TestWorkflowWorktreeKeyHonesty is the production honesty guard: the
// computed reader set of every WorkflowWorktreeConfig field must exactly
// equal its expectation-table entry, the table must cover the live struct
// (REQ-006), reserved keys must stay empty (REQ-005), the auto_cleanup pair
// is asserted on the table itself (REQ-004), and a type-broken tree fails
// with the cause named (REQ-007) instead of passing on a possibly-incomplete
// index.
func TestWorkflowWorktreeKeyHonesty(t *testing.T) {
	idx, fields := loadWorktreeScan(t, "./...")

	// REQ-007 — a typed-but-broken tree must fail naming package and error.
	if len(idx.typeErrors) > 0 {
		t.Fatalf("REQ-007: type errors across scanned packages — reader index may be incomplete:\n  %s",
			strings.Join(idx.typeErrors, "\n  "))
	}

	// REQ-006 — table completeness against the live struct.
	keyToField := make(map[string]string, len(fields))
	for _, f := range fields {
		if f.yamlKey == "" {
			t.Errorf("REQ-006: field %s carries no yaml tag — the expectation table cannot address it", f.name)
			continue
		}
		if _, ok := expectedWorktreeReaders[f.yamlKey]; !ok {
			t.Errorf("REQ-006: field %s (key %q) has no expectation-table entry — add one when the field lands", f.name, f.yamlKey)
			continue
		}
		keyToField[f.yamlKey] = f.name
	}

	// REQ-004 — independent table-content assertion on auto_cleanup,
	// evaluated on the table without consulting the computed scan.
	have := make(map[string]bool, len(mandatoryAutoCleanupReaders))
	for _, file := range expectedWorktreeReaders["auto_cleanup"] {
		have[file] = true
	}
	for _, want := range mandatoryAutoCleanupReaders {
		if !have[want] {
			t.Errorf("REQ-004: auto_cleanup expectation must name mandatory reader %s (both auto-cleanup sites gate worktree removal)", want)
		}
	}

	// REQ-005 — reserved-key table validity, independent of the scan.
	for _, key := range sortedStringKeys(reservedWorktreeKeys) {
		if entry, ok := expectedWorktreeReaders[key]; ok && len(entry) > 0 {
			t.Errorf("REQ-005: reserved key %q must have an empty expectation; the table names %v", key, entry)
		}
	}

	// REQ-003 — set equality per key, named findings in both directions.
	for _, key := range sortedStringKeys(expectedWorktreeReaders) {
		field, ok := keyToField[key]
		if !ok {
			continue // REQ-006 already reported the unmatched key
		}
		expSet := make(map[string]bool, len(expectedWorktreeReaders[key]))
		for _, file := range expectedWorktreeReaders[key] {
			expSet[file] = true
		}
		computed := idx.readers[field]
		for _, want := range expectedWorktreeReaders[key] {
			if !computed[want] {
				t.Errorf("REQ-003 dropped reader: key %q (field %s) — expected reader %s no longer reads the field", key, field, want)
			}
		}
		for got := range computed {
			if !expSet[got] {
				t.Errorf("REQ-003 unnamed reader: key %q (field %s) — %s reads the field but the expectation table does not name it", key, field, got)
			}
		}
	}
}

// loadWorktreeScan loads the given patterns with the shipped-key precedent's
// config (no Tests flag — test files are structurally absent from the index)
// and returns the computed reader index plus the live struct fields.
func loadWorktreeScan(t *testing.T, patterns ...string) (*readerIndex, []worktreeField) {
	t.Helper()

	root := findRepoRoot(t)
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedFiles,
		Dir:  root,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		t.Fatalf("go/packages.Load(%v) failed: %v", patterns, err)
	}

	fields := liveWorktreeFields(t, pkgs)
	fieldObjs := make(map[*types.Var]string, len(fields))
	for _, f := range fields {
		fieldObjs[f.obj] = f.name
	}

	idx := &readerIndex{readers: map[string]map[string]bool{}}
	for _, pkg := range pkgs {
		for _, terr := range pkg.TypeErrors {
			idx.typeErrors = append(idx.typeErrors, pkg.PkgPath+": "+terr.Error())
		}
		if pkg.TypesInfo == nil {
			continue
		}
		idx.pkgsScanned++
		for i, file := range pkg.Syntax {
			if i >= len(pkg.GoFiles) {
				continue
			}
			// REQ-008 — belt-and-braces on top of the structural exclusion:
			// the production load carries no test files to begin with.
			if strings.HasSuffix(pkg.GoFiles[i], "_test.go") {
				continue
			}
			idx.filesScanned++
			scanFileForWorktreeFieldReads(file, pkg, fieldObjs, repoRel(root, pkg.GoFiles[i]), idx)
		}
	}

	// Non-vacuity — an empty index proves nothing (acceptance §D.11).
	t.Logf("reader index (%v): %d packages, %d files scanned, %d type errors",
		patterns, idx.pkgsScanned, idx.filesScanned, len(idx.typeErrors))
	if idx.pkgsScanned == 0 || idx.filesScanned == 0 {
		t.Fatalf("reader index is empty (%d packages, %d files scanned) — wrong Dir or pattern; refusing a vacuous pass",
			idx.pkgsScanned, idx.filesScanned)
	}
	return idx, fields
}

// liveWorktreeFields resolves WorkflowWorktreeConfig's live fields — Go name,
// yaml key and *types.Var identity — from the loaded config package, so that
// REQ-006 completeness and the scan both run against the real struct rather
// than a hardcoded field list.
func liveWorktreeFields(t *testing.T, pkgs []*packages.Package) []worktreeField {
	t.Helper()
	for _, pkg := range pkgs {
		if pkg.PkgPath != worktreeConfigPkgPath || pkg.Types == nil {
			continue
		}
		obj := pkg.Types.Scope().Lookup(worktreeConfigTypeName)
		if obj == nil {
			continue
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}
		st, ok := named.Underlying().(*types.Struct)
		if !ok {
			continue
		}
		out := make([]worktreeField, 0, st.NumFields())
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			if !f.Exported() {
				continue
			}
			tag := reflect.StructTag(st.Tag(i)).Get("yaml")
			out = append(out, worktreeField{
				name:    f.Name(),
				yamlKey: strings.SplitN(tag, ",", 2)[0],
				obj:     f,
			})
		}
		return out
	}
	t.Fatalf("%s.%s not found among loaded packages — cannot resolve field objects",
		worktreeConfigPkgPath, worktreeConfigTypeName)
	return nil
}

// scanFileForWorktreeFieldReads walks one file and records every
// type-resolved read of a tracked field, attributed to relFile. Resolution is
// by field-object identity (Selections[sel].Obj()), which is alias-proof by
// construction: a local copy of the struct resolves to the same field object
// as the original accessor chain.
func scanFileForWorktreeFieldReads(file *ast.File, pkg *packages.Package, fieldObjs map[*types.Var]string, relFile string, idx *readerIndex) {
	var stack []ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		stack = append(stack, n)
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		selection := pkg.TypesInfo.Selections[sel]
		if selection == nil || selection.Kind() != types.FieldVal {
			return true
		}
		fieldVar, ok := selection.Obj().(*types.Var)
		if !ok {
			return true
		}
		field, ok := fieldObjs[fieldVar]
		if !ok {
			return true
		}
		if isPlainAssignLHS(stack, sel) { // REQ-002's sole exclusion
			return true
		}
		if idx.readers[field] == nil {
			idx.readers[field] = map[string]bool{}
		}
		idx.readers[field][relFile] = true
		return true
	})
}

// isPlainAssignLHS reports whether sel is the direct left-hand operand of a
// plain `=` assignment — the one non-read shape REQ-002 excepts. Compound
// assignments, ++/-- operands and address-taking resolve through other parent
// nodes and stay reads.
func isPlainAssignLHS(stack []ast.Node, sel *ast.SelectorExpr) bool {
	if len(stack) < 2 {
		return false
	}
	assign, ok := stack[len(stack)-2].(*ast.AssignStmt)
	if !ok || assign.Tok != token.ASSIGN {
		return false
	}
	for _, lhs := range assign.Lhs {
		if lhs == sel {
			return true
		}
	}
	return false
}

// findRepoRoot returns the module root derived from this file's location —
// the shipped_key_reader_test.go precedent's mechanism.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file = <repo root>/internal/template/workflow_worktree_key_honesty_test.go
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// repoRel renders an absolute path repo-root-relative with forward slashes.
func repoRel(root, abs string) string {
	if rel, err := filepath.Rel(root, abs); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(abs)
}

// sortedStringKeys returns the map's keys in sorted order so guard findings
// are deterministic across runs.
func sortedStringKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// The fixture-mode characterization (TestWorkflowWorktreeKeyHonestyAliasFixture)
// lands with the testdata fixture package (M2 of plan.md §F).
