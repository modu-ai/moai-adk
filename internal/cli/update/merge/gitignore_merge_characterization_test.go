package merge

// Issue #1094 self-reproduction attempt: the reporter claims MergeGitignoreFile
// deletes the "User Custom Patterns (preserved by moai update)" block. These
// characterization tests probe the edge cases most likely to lose user lines.
// Cases that do NOT reproduce a loss are kept as regression guards pinning the
// preservation behavior.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeGitignore writes template content to a temp .gitignore (simulating the
// freshly deployed template) and returns its path.
func writeGitignore(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".gitignore")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func mergedContent(t *testing.T, template, userBackup string) string {
	t.Helper()
	path := writeGitignore(t, template)
	if err := MergeGitignoreFile(path, []byte(userBackup)); err != nil {
		t.Fatalf("MergeGitignoreFile: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// Case (a): a user line that byte-collides with a template line is deduplicated
// (not re-appended) while distinct user lines survive.
func TestMergeGitignore_UserLineCollidingWithTemplate_Deduplicated(t *testing.T) {
	t.Parallel()
	got := mergedContent(t,
		"node_modules/\n.moai/cache/\n",
		"node_modules/\nmy-secret-dir/\n")
	if strings.Count(got, "node_modules/") != 1 {
		t.Errorf("colliding line duplicated:\n%s", got)
	}
	if !strings.Contains(got, "my-secret-dir/") {
		t.Errorf("distinct user line LOST (would reproduce #1094):\n%s", got)
	}
}

// Case (b): '!' negation lines are user patterns like any other and must survive.
func TestMergeGitignore_NegationLinesPreserved(t *testing.T) {
	t.Parallel()
	got := mergedContent(t,
		"dist/\n*.log\n",
		"*.env\n!.env.example\n")
	for _, want := range []string{"*.env", "!.env.example"} {
		if !strings.Contains(got, want) {
			t.Errorf("negation-path user line %q LOST (would reproduce #1094):\n%s", want, got)
		}
	}
}

// Case (c): re-merge path — the user backup already carries a
// "User Custom Patterns" header from a PREVIOUS moai update run. The pattern
// lines under the old header must be re-collected under the new header
// (the old header comment itself is regenerated, not duplicated).
func TestMergeGitignore_ReMergeWithExistingHeader_PatternsSurvive(t *testing.T) {
	t.Parallel()
	userBackup := "dist/\n\n# User Custom Patterns (preserved by moai update)\nmy-tool-output/\n*.scratch\n"
	got := mergedContent(t, "dist/\n*.log\n", userBackup)
	for _, want := range []string{"my-tool-output/", "*.scratch"} {
		if !strings.Contains(got, want) {
			t.Errorf("re-merge dropped user pattern %q (would reproduce #1094):\n%s", want, got)
		}
	}
	if strings.Count(got, "# User Custom Patterns") != 1 {
		t.Errorf("expected exactly one User Custom Patterns header, got %d:\n%s",
			strings.Count(got, "# User Custom Patterns"), got)
	}
}

// Case (c2): double re-merge — running the merge twice in a row must be
// idempotent (no pattern loss, no header accumulation).
func TestMergeGitignore_DoubleReMerge_Idempotent(t *testing.T) {
	t.Parallel()
	template := "dist/\n"
	first := mergedContent(t, template, "dist/\ncustom-a/\ncustom-b/\n")

	// Second update: fresh template deploy, backup is the previously merged file.
	got := mergedContent(t, template, first)
	for _, want := range []string{"custom-a/", "custom-b/"} {
		if strings.Count(got, want) != 1 {
			t.Errorf("second merge: pattern %q count = %d, want 1:\n%s",
				want, strings.Count(got, want), got)
		}
	}
	if strings.Count(got, "# User Custom Patterns") != 1 {
		t.Errorf("second merge: header count = %d, want 1:\n%s",
			strings.Count(got, "# User Custom Patterns"), got)
	}
}

// Case (d): CRLF line endings in the user backup — trimming must still match
// template lines (dedup) and preserve distinct patterns.
func TestMergeGitignore_CRLFBackup_PatternsSurvive(t *testing.T) {
	t.Parallel()
	got := mergedContent(t,
		"dist/\n*.log\n",
		"dist/\r\ncrlf-custom/\r\n")
	if !strings.Contains(got, "crlf-custom/") {
		t.Errorf("CRLF user pattern LOST (would reproduce #1094):\n%s", got)
	}
	// The colliding "dist/" (with \r) must be deduplicated, not re-appended.
	if strings.Count(got, "dist/") != 1 {
		t.Errorf("CRLF colliding line duplicated:\n%s", got)
	}
}

// Case (e): trailing-whitespace variant of a template line is deduplicated;
// a trailing-whitespace variant of a genuinely custom line survives.
func TestMergeGitignore_TrailingWhitespaceVariants(t *testing.T) {
	t.Parallel()
	got := mergedContent(t,
		"dist/\n",
		"dist/   \nspaced-custom/  \n")
	if strings.Count(got, "dist/") != 1 {
		t.Errorf("trailing-whitespace collide not deduplicated:\n%s", got)
	}
	if !strings.Contains(got, "spaced-custom/") {
		t.Errorf("trailing-whitespace custom pattern LOST (would reproduce #1094):\n%s", got)
	}
}

// Observed limitation (documented, not a #1094 reproduction): user COMMENT
// lines (e.g. "# my IDE stuff" annotations above custom patterns) are never
// carried over — only non-comment pattern lines are preserved. This pins the
// current behavior so any future change is deliberate.
func TestMergeGitignore_UserCommentAnnotationsNotCarried(t *testing.T) {
	t.Parallel()
	got := mergedContent(t,
		"dist/\n",
		"# my IDE annotations\n.idea-custom/\n")
	if !strings.Contains(got, ".idea-custom/") {
		t.Errorf("pattern under user comment LOST (would reproduce #1094):\n%s", got)
	}
	if strings.Contains(got, "# my IDE annotations") {
		t.Errorf("behavior changed: user comment annotations now carried over — update this characterization:\n%s", got)
	}
}

// Empty-additions path: when every user line is already in the template, the
// deployed template is left untouched (no header is appended).
func TestMergeGitignore_NoAdditions_TemplateUntouched(t *testing.T) {
	t.Parallel()
	template := "dist/\n*.log\n"
	got := mergedContent(t, template, "dist/\n*.log\n")
	if got != template {
		t.Errorf("template mutated with no user additions:\ngot:  %q\nwant: %q", got, template)
	}
}
