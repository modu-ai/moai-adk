// Package merge — gitignore_marker_block_test.go
//
// Regression guards for the .gitignore merge losing user comments and ordering
// (issues #1415 / #1377 residual).
//
// MergeGitignoreFile did line-set subtraction: it skipped blank and `#`-prefixed
// lines when collecting user additions and then re-emitted a fresh
// "User Custom Patterns (preserved by moai update)" header. The tool therefore
// regenerated the very block it advertised as preserved — comments were lost and
// ordering was destroyed. Ordering is semantic in gitignore: a `!negation` placed
// after an unrelated rule behaves differently from one placed before it.

package merge

import (
	"os"
	"strings"
	"testing"
)

const testUserMarker = "# User Custom Patterns (preserved by moai update)"

// readGitignore reads the merged file back.
func readGitignore(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestMergeGitignoreFile_MarkerBlockPreservesCommentsAndOrder is the primary
// issue #1415 guard: when the backup carries the marker line, everything below
// it is carried over verbatim — user comments intact, original order intact.
func TestMergeGitignoreFile_MarkerBlockPreservesCommentsAndOrder(t *testing.T) {
	const template = "# MoAI template\nnode_modules/\n.env\n"
	userBackup := "# MoAI template\nnode_modules/\n.env\n" +
		"\n" + testUserMarker + "\n" +
		"# @MX:WARN keep this\n" +
		".moai/*\n" +
		"!.moai/specs/\n" +
		"\n" +
		"# build outputs\n" +
		"dist/\n"

	path := writeGitignore(t, template)
	if err := MergeGitignoreFile(path, []byte(userBackup)); err != nil {
		t.Fatalf("MergeGitignoreFile: %v", err)
	}
	got := readGitignore(t, path)

	// (1) the user's comments survive
	for _, want := range []string{"# @MX:WARN keep this", "# build outputs"} {
		if !strings.Contains(got, want) {
			t.Errorf("user comment %q lost by the merge\ngot:\n%s", want, got)
		}
	}

	// (2) ordering survives: `.moai/*` must still precede its negation
	iPattern := strings.Index(got, ".moai/*")
	iNegation := strings.Index(got, "!.moai/specs/")
	if iPattern < 0 || iNegation < 0 {
		t.Fatalf("user patterns missing from merged output\ngot:\n%s", got)
	}
	if iPattern > iNegation {
		t.Errorf("ordering destroyed: `.moai/*` must precede `!.moai/specs/`\ngot:\n%s", got)
	}

	// (3) the marker appears exactly once
	if n := strings.Count(got, testUserMarker); n != 1 {
		t.Errorf("marker line appears %d times, want 1\ngot:\n%s", n, got)
	}
}

// TestMergeGitignoreFile_MarkerBlockIsIdempotent asserts a second merge over the
// already-merged result neither duplicates the marker nor duplicates patterns.
func TestMergeGitignoreFile_MarkerBlockIsIdempotent(t *testing.T) {
	const template = "# MoAI template\nnode_modules/\n"
	userBackup := template + "\n" + testUserMarker + "\n# mine\n.moai/*\n!.moai/specs/\n"

	path := writeGitignore(t, template)
	if err := MergeGitignoreFile(path, []byte(userBackup)); err != nil {
		t.Fatalf("first merge: %v", err)
	}
	first := readGitignore(t, path)

	// Second pass: redeploy the template, then merge the FIRST result back as the
	// backup — exactly what a second `moai update` does.
	if err := os.WriteFile(path, []byte(template), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := MergeGitignoreFile(path, []byte(first)); err != nil {
		t.Fatalf("second merge: %v", err)
	}
	second := readGitignore(t, path)

	if second != first {
		t.Errorf("merge is not idempotent\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if n := strings.Count(second, testUserMarker); n != 1 {
		t.Errorf("marker duplicated after second merge (%d occurrences)\ngot:\n%s", n, second)
	}
	if n := strings.Count(second, ".moai/*"); n != 1 {
		t.Errorf("pattern `.moai/*` duplicated after second merge (%d occurrences)\ngot:\n%s", n, second)
	}
}

// TestMergeGitignoreFile_MarkerBlockDropsTemplateDuplicates asserts that a
// non-comment line already present in the template is not re-emitted, while the
// surrounding structure is left intact.
func TestMergeGitignoreFile_MarkerBlockDropsTemplateDuplicates(t *testing.T) {
	const template = "# MoAI template\nnode_modules/\n.env\n"
	userBackup := template + "\n" + testUserMarker + "\n" +
		"# my section\n" +
		"node_modules/\n" + // duplicate of a template line
		"coverage/\n"

	path := writeGitignore(t, template)
	if err := MergeGitignoreFile(path, []byte(userBackup)); err != nil {
		t.Fatalf("MergeGitignoreFile: %v", err)
	}
	got := readGitignore(t, path)

	if n := strings.Count(got, "node_modules/"); n != 1 {
		t.Errorf("template duplicate `node_modules/` re-emitted (%d occurrences)\ngot:\n%s", n, got)
	}
	if !strings.Contains(got, "# my section") {
		t.Errorf("surrounding comment dropped along with the duplicate\ngot:\n%s", got)
	}
	if !strings.Contains(got, "coverage/") {
		t.Errorf("non-duplicate user pattern `coverage/` lost\ngot:\n%s", got)
	}
}

// TestMergeGitignoreFile_NoMarkerFallsBackToSubtraction asserts the pre-existing
// set-subtraction behaviour still applies to a backup with no marker — the
// first-time upgrader path.
func TestMergeGitignoreFile_NoMarkerFallsBackToSubtraction(t *testing.T) {
	const template = "# MoAI template\nnode_modules/\n"
	userBackup := "# MoAI template\nnode_modules/\ncoverage/\n.idea/\n"

	path := writeGitignore(t, template)
	if err := MergeGitignoreFile(path, []byte(userBackup)); err != nil {
		t.Fatalf("MergeGitignoreFile: %v", err)
	}
	got := readGitignore(t, path)

	if !strings.Contains(got, testUserMarker) {
		t.Errorf("fallback path did not emit the user-patterns marker\ngot:\n%s", got)
	}
	for _, want := range []string{"coverage/", ".idea/"} {
		if !strings.Contains(got, want) {
			t.Errorf("fallback path lost user pattern %q\ngot:\n%s", want, got)
		}
	}
}

// TestMergeGitignoreFile_MarkerBlockEmptyIsNoOp asserts a marker block whose
// content adds nothing leaves the file untouched.
func TestMergeGitignoreFile_MarkerBlockEmptyIsNoOp(t *testing.T) {
	const template = "# MoAI template\nnode_modules/\n"
	userBackup := template + "\n" + testUserMarker + "\n"

	path := writeGitignore(t, template)
	if err := MergeGitignoreFile(path, []byte(userBackup)); err != nil {
		t.Fatalf("MergeGitignoreFile: %v", err)
	}
	if got := readGitignore(t, path); got != template {
		t.Errorf("empty marker block was not a no-op\nwant:\n%s\ngot:\n%s", template, got)
	}
}
