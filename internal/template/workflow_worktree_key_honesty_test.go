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
// @MX:NOTE the production scan mirrors the load shape of the in-repo
// precedent internal/config/shipped_key_reader_test.go (packages.Config
// without the Tests flag, ./... pattern, Selections route); only the scoped
// fixture-mode load adds NeedDeps|NeedImports so internal/config is resolvable
// as an import. Extraction to a shared helper is deliberately deferred (plan
// D2) until a second consumer exists.

import (
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
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
}

// reservedWorktreeKeys declares the keys whose reader status is "reserved: no
// production reader" per the WorkflowWorktreeConfig doc comment and the
// shipped workflow.yaml template. Declared independently of the expectation
// table so that a table naming a file under a reserved key is itself invalid
// (REQ-005) rather than silently re-classifying the key as read.
var reservedWorktreeKeys = map[string]bool{
	"session_name_pattern": true,
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
	loadErrors   []string // type + parse + list errors, package and cause named
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

	// REQ-007 — a typed-or-parsed-broken tree must fail naming package and
	// cause, never pass on a possibly-incomplete index.
	if len(idx.loadErrors) > 0 {
		t.Fatalf("REQ-007: load errors (type/parse/list) across scanned packages — reader index may be incomplete:\n  %s",
			strings.Join(idx.loadErrors, "\n  "))
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
	for _, key := range slices.Sorted(maps.Keys(reservedWorktreeKeys)) {
		if entry, ok := expectedWorktreeReaders[key]; ok && len(entry) > 0 {
			t.Errorf("REQ-005: reserved key %q must have an empty expectation; the table names %v", key, entry)
		}
	}

	// REQ-003 — set equality per key, named findings in both directions.
	for _, key := range slices.Sorted(maps.Keys(expectedWorktreeReaders)) {
		field, ok := keyToField[key]
		if !ok {
			// REQ-006 (bidirectional completeness): a table key naming no
			// live struct field is itself a defect — REQ-006's field→entry
			// direction above cannot see it.
			t.Errorf("REQ-006 orphan key: %q names no field of the live %s struct — remove the entry or add the field it expected", key, worktreeConfigTypeName)
			continue
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

// Load modes. The production scan uses the shipped-key precedent's exact
// mode set; the scoped fixture-mode load adds NeedDeps|NeedImports so that
// internal/config — a dependency, not a root, under an explicit
// single-package pattern — is reachable with populated Types through the
// Imports graph.
const (
	prodScanMode    = packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedFiles
	fixtureScanMode = prodScanMode | packages.NeedDeps | packages.NeedImports
)

// loadWorktreeScan loads the given patterns with the precedent's mode set and
// returns the computed reader index plus the live struct fields.
func loadWorktreeScan(t *testing.T, patterns ...string) (*readerIndex, []worktreeField) {
	t.Helper()
	return loadWorktreeScanMode(t, prodScanMode, patterns...)
}

// loadWorktreeScanWithDeps is the fixture-mode load: same scanner, mode
// extended with NeedDeps|NeedImports for scoped explicit patterns.
func loadWorktreeScanWithDeps(t *testing.T, patterns ...string) (*readerIndex, []worktreeField) {
	t.Helper()
	return loadWorktreeScanMode(t, fixtureScanMode, patterns...)
}

// loadWorktreeScanMode runs the scan: no Tests flag (test files are
// structurally absent from the index), root packages only in the scan loop,
// and fail-closed error collection (type + parse + list errors, each named by
// package and cause).
func loadWorktreeScanMode(t *testing.T, mode packages.LoadMode, patterns ...string) (*readerIndex, []worktreeField) {
	t.Helper()

	root := findRepoRoot(t)
	cfg := &packages.Config{
		Mode: mode,
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
		// Fail-closed error collection (REQ-007 + audit repair): pkg.Errors
		// is the superset (list + parse + type); pkg.TypeErrors is recorded
		// first so REQ-007's named surface leads. Deduped by text.
		seenErr := make(map[string]bool, len(pkg.Errors)+len(pkg.TypeErrors))
		addErr := func(msg string) {
			if !seenErr[msg] {
				seenErr[msg] = true
				idx.loadErrors = append(idx.loadErrors, msg)
			}
		}
		for _, e := range pkg.TypeErrors {
			addErr(pkg.PkgPath + ": " + e.Error())
		}
		for _, e := range pkg.Errors {
			addErr(pkg.PkgPath + ": " + e.Error())
		}
		if pkg.TypesInfo == nil {
			continue
		}
		idx.pkgsScanned++
		for _, file := range pkg.Syntax {
			if file == nil || !file.Pos().IsValid() {
				continue
			}
			tf := pkg.Fset.File(file.Pos())
			if tf == nil {
				continue
			}
			// Attribute by Fset, not by GoFiles index alignment — the
			// alignment diverges under cgo-generated syntax entries.
			abs := tf.Name()
			// REQ-008 — belt-and-braces on top of the structural exclusion:
			// the production load carries no test files to begin with.
			if strings.HasSuffix(abs, "_test.go") {
				continue
			}
			idx.filesScanned++
			scanFileForWorktreeFieldReads(file, pkg, fieldObjs, repoRel(root, abs), idx)
		}
	}

	// Non-vacuity — an empty index proves nothing (acceptance §D.11).
	t.Logf("reader index (%v): %d packages, %d files scanned, %d load errors",
		patterns, idx.pkgsScanned, idx.filesScanned, len(idx.loadErrors))
	if idx.pkgsScanned == 0 || idx.filesScanned == 0 {
		t.Fatalf("reader index is empty (%d packages, %d files scanned) — wrong Dir or pattern; refusing a vacuous pass",
			idx.pkgsScanned, idx.filesScanned)
	}
	return idx, fields
}

// liveWorktreeFields resolves WorkflowWorktreeConfig's live fields — Go name,
// yaml key and *types.Var identity — from the loaded config package, so that
// REQ-006 completeness and the scan both run against the real struct rather
// than a hardcoded field list. The config package is looked up across the
// loaded package graph (roots and their transitive imports): a scoped scan
// such as the fixture characterization loads the fixture package as the only
// root, with internal/config reachable as an import.
func liveWorktreeFields(t *testing.T, pkgs []*packages.Package) []worktreeField {
	t.Helper()
	var configPkg *packages.Package
	seen := map[*packages.Package]bool{}
	var walk func(*packages.Package)
	walk = func(p *packages.Package) {
		if p == nil || seen[p] {
			return
		}
		seen[p] = true
		if p.PkgPath == worktreeConfigPkgPath && p.Types != nil && configPkg == nil {
			configPkg = p
		}
		for _, imp := range p.Imports {
			walk(imp)
		}
	}
	for _, pkg := range pkgs {
		walk(pkg)
	}
	if configPkg == nil {
		t.Fatalf("%s.%s not found among loaded packages — cannot resolve field objects",
			worktreeConfigPkgPath, worktreeConfigTypeName)
		return nil
	}
	obj := configPkg.Types.Scope().Lookup(worktreeConfigTypeName)
	if obj == nil {
		t.Fatalf("%s not declared in %s — cannot resolve field objects",
			worktreeConfigTypeName, worktreeConfigPkgPath)
		return nil
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		t.Fatalf("%s in %s is not a named type", worktreeConfigTypeName, worktreeConfigPkgPath)
		return nil
	}
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		t.Fatalf("%s in %s does not have a struct underlying type", worktreeConfigTypeName, worktreeConfigPkgPath)
		return nil
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

// scanFileForWorktreeFieldReads walks one file and records every
// type-resolved read of a tracked field, attributed to relFile. Resolution is
// by field-object identity (Selections[sel].Obj()), which is alias-proof by
// construction: a local copy of the struct resolves to the same field object
// as the original accessor chain.
func scanFileForWorktreeFieldReads(file *ast.File, pkg *packages.Package, fieldObjs map[*types.Var]string, relFile string, idx *readerIndex) {
	// ast.Inspect parent stack: push on every visited node, pop on the
	// trailing f(nil) that closes it — stack[len(stack)-2] is sel's parent.
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

// Guard findings iterate table keys in sorted order (slices.Sorted(maps.Keys))
// so output is deterministic across runs — the idiom the sibling guard tests
// use (internal/cli/huh_v1_guard_test.go).

// TestWorkflowWorktreeKeyHonestyAliasFixture is the alias characterization
// (M2): the same scan path, pointed at the testdata fixture package by
// explicit pattern, proves alias-copy reads are attributed (AC-005a), plain
// `=` writes are excluded (AC-005b), compound assignments are reads
// (AC-005c) and the legacy text accessor string is absent from the alias
// fixture source (AC-005d — the t682 text scan's blind spot, recorded as
// evidence).
func TestWorkflowWorktreeKeyHonestyAliasFixture(t *testing.T) {
	const (
		fixturePattern = "./internal/template/testdata/worktreekeyaliasprobe"
		fixtureDir     = "internal/template/testdata/worktreekeyaliasprobe"
	)
	idx, _ := loadWorktreeScanWithDeps(t, fixturePattern)

	if len(idx.loadErrors) > 0 {
		t.Fatalf("load errors in the fixture package — fixture scan is unreliable:\n  %s",
			strings.Join(idx.loadErrors, "\n  "))
	}

	// AC-005a — alias-copy read attributed.
	if !idx.readers["AutoCleanup"][fixtureDir+"/aliasprobe.go"] {
		t.Errorf("AC-005a: aliasprobe.go reads AutoCleanup through a local alias copy but was not attributed as a reader")
	}
	// AC-005b — plain-`=` write excluded (REQ-002).
	if idx.readers["AutoCleanup"][fixtureDir+"/writeonly.go"] {
		t.Errorf("AC-005b: writeonly.go touches AutoCleanup only as a plain `=` write — it must not be classified as a reader")
	}
	// AC-005c — compound assignment is a read (REQ-002).
	if !idx.readers["SessionNamePattern"][fixtureDir+"/compoundassign.go"] {
		t.Errorf("AC-005c: compoundassign.go consumes SessionNamePattern via += and must be classified as a reader")
	}

	// AC-005d — the retained t682 text scan demonstrably reports NO reader
	// here: neither legacy accessor form matches the alias fixture's source
	// (comments included), so a text scan finds nothing — while the AST scan
	// above attributes the read (AC-005a). The comparison is the recorded
	// blind spot.
	src, err := os.ReadFile(filepath.Join(findRepoRoot(t), fixtureDir, "aliasprobe.go"))
	if err != nil {
		t.Fatalf("read aliasprobe.go: %v", err)
	}
	t682TextScanDetectsReader := strings.Contains(string(src), ".Workflow.Worktree.AutoCleanup") ||
		strings.Contains(string(src), ".Worktree.AutoCleanup")
	if t682TextScanDetectsReader {
		t.Errorf("AC-005d: aliasprobe.go matches a legacy accessor string — the alias-copy shape must not spell the accessor")
	}
	if !idx.readers["AutoCleanup"][fixtureDir+"/aliasprobe.go"] {
		t.Errorf("AC-005d: the AST scan must attribute the alias read the t682 text scan cannot see (zero text matches above)")
	}
}
