package contract

import "testing"

func TestMatchGlob(t *testing.T) {
	const contractPath = ".moai/specs/SPEC-FIXTURE-001/contract.yaml"
	cases := []struct {
		pattern, name string
		want          bool
	}{
		// design.md § Glob Semantics — the coverage examples.
		{".moai/**", contractPath, true},
		{".moai/specs/**", contractPath, true},
		{".moai/specs/SPEC-FIXTURE-001/**", contractPath, true},
		{".moai/specs/*.md", contractPath, false},
		{".moai/specs/SPEC-FIXTURE-001/contract.yaml", contractPath, true},
		{".moai/specs/*/contract.yaml", contractPath, true},
		{".moai/*/contract.yaml", contractPath, false},

		// ** matches zero or more whole segments.
		{"**/CLAUDE.md", "CLAUDE.md", true},
		{"**/CLAUDE.md", "a/b/CLAUDE.md", true},
		{"**/CLAUDE.md", "a/CLAUDE.md.bak", false},
		{"**/CLAUDE.md", "a/xCLAUDE.md", false},
		{"**", "anything/at/all", true},
		{"**", "x", true},
		{"a/**/b", "a/b", true},
		{"a/**/b", "a/x/y/b", true},
		{"a/**/b", "a/x/y/c", false},
		{"a/**", "a", true},
		{"a/**", "ab", false},
		{"**/x/**", "p/x/q", true},
		{"**/x/**", "p/y/q", false},

		// * stays inside one segment; ? matches one non-separator character.
		{"internal/*", "internal/a", true},
		{"internal/*", "internal/a/b", false},
		{"internal/*.go", "internal/a.go", true},
		{"internal/?.go", "internal/a.go", true},
		{"internal/?.go", "internal/ab.go", false},
		{"a?b", "a/b", false},
		{"*", "", true},
		{"a/*/c", "a//c", true},

		// Literal and malformed patterns.
		{"internal/fixture/x.go", "internal/fixture/x.go", true},
		{"internal/fixture/x.go", "internal/fixture/y.go", false},
		{"[", "[", false},
		{"a/[", "a/[", false},
		{"a/b", "a/b/c", false},
		{"a/b/c", "a/b", false},
	}
	for _, tc := range cases {
		if got := MatchGlob(tc.pattern, tc.name); got != tc.want {
			t.Errorf("MatchGlob(%q, %q) = %v, want %v", tc.pattern, tc.name, got, tc.want)
		}
	}
}
