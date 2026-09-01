package graph

import (
	"os"
	"path/filepath"
	"testing"
)

// tierFixture builds a Go module exercising every confidence tier:
//
//	wire.go   package wire — declares A, B; imports internal/helper
//	sib.go    package wire — declares Sibling, Dup, S2 (no imports)
//	helper    package helper — declares Shared, Dup, method Do
//	far       package far — declares Remote (never imported by wire)
//
// Expected tiers from wire.go's A: B=extracted (T1 same file),
// Shared/Dup/Do=extracted (T2 import evidence, Dup also in sibling+far),
// Sibling=intra-package (T3, declared in sib.go), Remote/Nowhere=inferred
// (T4). From sib.go's S2: B=intra-package (T3, cross-file, zero imports).
func tierFixture(t *testing.T) string {
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
	write("go.mod", "module example.com/tier\n\ngo 1.26\n")
	write("internal/wire/wire.go", `package wire

import "example.com/tier/internal/helper"

func A() {
	B()
	helper.Shared()
	Remote()
	Dup()
	Sibling()
	Nowhere()
	h := helper.Helper{}
	h.Do()
}

func B() {
	helper.Shared()
}
`)
	write("internal/wire/sib.go", `package wire

func Sibling() {}

func Dup() {}

func S2() {
	B()
}
`)
	write("internal/helper/helper.go", `package helper

func Shared() {}

func Dup() {}

func Other() {}

type Helper struct{}

func (Helper) Do() {}
`)
	write("internal/far/far.go", `package far

func Remote() {}
`)
	// Non-Go case (repair-round D1): a python caller invokes a name declared
	// in a DIFFERENT directory via a specifier T2 can never resolve — the
	// edge must fall through to T4 inferred, never promote.
	write("internal/pyapp/app.py", `def caller():
    helper_fn()


def local_one():
    pass
`)
	write("internal/pylib/worker.py", `def helper_fn():
    pass
`)
	return root
}

// goldenFixture builds the deterministic tree the committed doc-edge golden
// (testdata/edges-doc-golden.jsonl) was generated from (plan.md §G): every
// doc-derived kind (import, mx-spec, spec-depends, report-milestone,
// milestone-card) plus code-import lines; code-call lines are excluded from
// the golden comparison. NEVER hand-edit the golden — regenerate on a new
// base when this fixture changes, naming the base SHA.
func goldenFixture(t *testing.T) string {
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
	write("go.mod", "module example.com/fix\n\ngo 1.26\n")
	write(".moai/project/codemaps/dependencies.md",
		"```mermaid\ngraph TD\n    a[\"internal/alpha\"] --> b[\"internal/beta\"]\n```\n")
	write(".moai/specs/SPEC-GF-001/spec.md",
		"---\nid: SPEC-GF-001\ntitle: \"t\"\ndepends_on: [SPEC-GF-DEP-001]\n---\n")
	write(".moai/reports/demo.md", `# Demo

## Card Cross-Check

| milestone | scope | card |
|---|---|---|
| S1 | demo scope | t999 |
`)
	write("internal/alpha/alpha.go", `package alpha

import "fmt"

func Alpha() {
	fmt.Println("alpha")
}
`)
	write("internal/beta/beta.go", `package beta

import (
	"fmt"

	"example.com/fix/internal/alpha"
)

// @MX:NOTE: [AUTO] beta demo edge fixture
// @MX:SPEC:SPEC-GF-001
func Beta() {
	alpha.Alpha()
	Nowhere()
	fmt.Println("beta")
}
`)
	// Tag-kind golden coverage (SPEC-MX-TAG-EDGES-001 §B.5): every mx-*
	// kind — a body-anchored ANCHOR/WARN/TODO pair, a file-scope LEGACY, and
	// a DEBT pair with and without an @MX:UPGRADE sub-line (identical edge
	// key sets either way).
	write("internal/beta/tags.go", `package beta

// @MX:LEGACY: [AUTO] file-scope legacy marker

func Anchored() {
	// @MX:ANCHOR: [AUTO] golden fixture anchor
	// @MX:REASON: body-anchored for the range join
	_ = 1
}

func Warned() {
	// @MX:WARN: [AUTO] golden fixture warn
	// @MX:REASON: danger below
	go func() {}()
}

func Debted() {
	// @MX:DEBT: golden fixture debt, with trigger
	// @MX:CEILING: < 10k entries
	// @MX:UPGRADE: swap to LRU past the ceiling
	_ = map[string]int{}
}

func Rotting() {
	// @MX:DEBT: golden fixture debt, no trigger
	// @MX:CEILING: < 10k entries
	_ = map[string]int{}
}

func Todoed() {
	// @MX:TODO: golden fixture todo
	_ = 1
}
`)
	write("scripts/fix.py", `# @MX:NOTE: golden fixture python note
def fix():
    pass
`)
	return root
}
