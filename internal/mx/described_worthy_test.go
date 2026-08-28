package mx

import "testing"

// AC-GFC-001 — the described-worthy predicate over repo-relative slash paths.
// Every rejected case carries the branch that rejects it AND a witness path
// differing only by that branch's property, which MUST be admitted: without
// the witness a predicate could reject for the wrong reason and still pass.
func TestDescribedWorthy(t *testing.T) {
	cases := []struct {
		path    string
		admit   bool
		branch  string
		witness string // admitted variant isolating `branch` (rejected cases only)
	}{
		{path: "internal/astgrep/coverage_matrix.go", admit: true, branch: "admitted — .go, not a test, no testdata segment"},
		{path: "internal/astgrep/coverage_matrix_test.go", admit: false, branch: "_test.go suffix", witness: "internal/astgrep/coverage_matrix.go"},
		{path: "internal/astgrep/testdata/rule-tests/x.yml", admit: false, branch: "not .go", witness: "internal/astgrep/rule_tests.go"},
		{path: "internal/hook/testdata/nav/a.go", admit: false, branch: "testdata path segment", witness: "internal/hook/nav/a.go"},
		{path: "internal/template/templates/.moai/config/astgrep-rules/go/concurrency.yml", admit: false, branch: "not .go", witness: "internal/template/templates/concurrency.go"},

		// Segment equality, not substring (acceptance.md §D.2 edge case).
		{path: "internal/foo/testdatax/a.go", admit: true, branch: "admitted — testdatax is not the testdata segment"},
		{path: "testdata/a.go", admit: false, branch: "testdata path segment", witness: "a.go"},
		{path: "internal/a/testdata/b/c/d.go", admit: false, branch: "testdata path segment", witness: "internal/a/b/c/d.go"},
		{path: "internal/x_test.go", admit: false, branch: "_test.go suffix", witness: "internal/x.go"},
		{path: "cmd/tool/main.go", admit: true, branch: "admitted"},
		{path: "README.md", admit: false, branch: "not .go", witness: "readme.go"},
		{path: "internal/mx/_test.go", admit: false, branch: "_test.go suffix", witness: "internal/mx/t.go"},
	}

	for _, c := range cases {
		if got := IsDescribedWorthy(c.path); got != c.admit {
			t.Errorf("IsDescribedWorthy(%q) = %v, want %v (branch: %s)", c.path, got, c.admit, c.branch)
		}
		if c.admit {
			continue
		}
		if !IsDescribedWorthy(c.witness) {
			t.Errorf("witness %q for %q must be admitted — rejection of %q is not attributable to %q",
				c.witness, c.path, c.path, c.branch)
		}
	}
}
