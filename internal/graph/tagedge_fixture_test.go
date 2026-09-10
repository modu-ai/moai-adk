package graph

import (
	"os"
	"path/filepath"
	"testing"
)

// tagFixture builds the shared @MX-tag fixture tree for the tag-edge tests
// (plan.md §D): a Go module carrying all six standalone tag kinds —
// function-body-anchored, nested-closure, file-scope, and a DEBT pair with
// and without an @MX:UPGRADE sub-line (rot state must never leak into an
// edge either way) — plus one dynamic-language file for scanner coverage.
func tagFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, src string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/tags\n\ngo 1.26\n")

	// anchored.go: a tag INSIDE a function body (the REQ-MTE-002 join case)
	// and a nested-closure tag (innermost range wins, acceptance §D.2).
	write("internal/app/anchored.go", `package app

import "fmt"

func Anchored() {
	// @MX:NOTE: [AUTO] anchored inside the function body
	fmt.Println("anchored")
}

func Nested() {
	func() {
		// @MX:TODO: resolve the nested closure question
		_ = 1
	}()
}

func Contract() {
	// @MX:ANCHOR: [AUTO] invariant contract marker
	// @MX:REASON: body-anchored so the range join can resolve it
	_ = 1
}
`)

	// filescope.go: a tag at file scope (no enclosing range → self-edge).
	write("internal/app/filescope.go", `package app

// @MX:NOTE: [AUTO] file-scope note, outside any function range
const Version = "v1"
`)

	// warned.go: a WARN tag inside a function body.
	write("internal/app/warned.go", `package app

func Warned() {
	// @MX:WARN: [AUTO] danger zone in the body
	// @MX:REASON: goroutine below is unbounded
	go func() {}()
}
`)

	// debts.go: DEBT with an @MX:UPGRADE trigger + TODO + LEGACY-with-SPEC.
	write("internal/app/debts.go", `package app

func Debted() {
	// @MX:DEBT: in-memory cache, no eviction
	// @MX:CEILING: < 10k entries
	// @MX:UPGRADE: switch to LRU past the ceiling
	_ = map[string]int{}
}

func Todos() {
	// @MX:TODO: wire the exporter
	_ = 1
}

func Legacies() {
	// @MX:LEGACY: no SPEC coverage
	// @MX:SPEC:SPEC-MTE-FIXTURE-001
	_ = 1
}
`)

	// rot.go: DEBT WITHOUT an @MX:UPGRADE sub-line (RotRisk="no-trigger").
	// Its edge bytes must equal the triggered DEBT's key set (AC-MTE-007).
	write("internal/app/rot.go", `package app

func Rotting() {
	// @MX:DEBT: rotting cache, no upgrade trigger
	// @MX:CEILING: < 10k entries
	_ = map[string]int{}
}
`)

	// Dynamic-language coverage: the scanner reads non-Go files too.
	write("scripts/tool.py", `# @MX:NOTE: python-side note
def tool():
    pass
`)
	return root
}

// tagFixtureTagCount is the number of standalone tags tagFixture carries.
// Kept as a constant so the kind-domain and occurrence-count assertions name
// the expected population explicitly.
const tagFixtureTagCount = 10

// Fixture line anchors (1-indexed) for the endpoint tests — kept next to the
// fixture content so an edit to the content and to the expectation move
// together.
const (
	anchoredNoteLine   = 6  // inside Anchored() in anchored.go
	nestedTodoLine     = 12 // inside the closure inside Nested()
	fileScopeNoteLine  = 3  // filescope.go, outside any range
	anchoredAnchorLine = 18 // inside Contract()
	warnedWarnLine     = 4  // inside Warned()
	debtedDebtLine     = 4  // inside Debted() (with @MX:UPGRADE)
	rottingDebtLine    = 4  // inside Rotting() (no @MX:UPGRADE)
)
