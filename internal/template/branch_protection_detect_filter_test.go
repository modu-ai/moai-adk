package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBranchProtectionDetectFilterCoversParityInput (AC-CI-012,
// SPEC-CI-VERDICT-INTEGRITY-001): the branch-protection parity guard
// (branch_protection_parity_test.go) reads .github/branch-protection.json.gtmpl
// — if a PR edits that file alone, the ci.yml go_code detect filter must
// still route the PR to the Go test jobs, or the parity guard never
// executes in CI. This correspondence guard fails when the parity test's
// read input disappears from the detect filter.
func TestBranchProtectionDetectFilterCoversParityInput(t *testing.T) {
	t.Parallel()

	// Walk up from the package dir (tests run with cwd = internal/template/)
	// to the repo root, mirroring findProjectRootForMirrorTest's worktree
	// behavior without coupling to its defining file.
	projectRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		t.Fatalf("resolve project root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err != nil {
		t.Fatalf("project root %s has no go.mod (cwd=%s): %v", projectRoot, ".", err)
	}

	ciYml := filepath.Join(projectRoot, ".github", "workflows", "ci.yml")
	raw, err := os.ReadFile(ciYml)
	if err != nil {
		t.Fatalf("read %s: %v", ciYml, err)
	}
	content := string(raw)

	// Same literal the parity test reads (branchProtectionRelPath) — kept
	// inline to avoid coupling this guard to the other test file's symbols.
	const parityInput = ".github/branch-protection.json.gtmpl"
	if !strings.Contains(content, parityInput) {
		t.Fatalf("detect-filter correspondence violated: %s go_code filter does not cover %s — a PR editing it alone would skip the branch-protection parity guard (AC-CI-012)",
			ciYml, parityInput)
	}

	// The parity test itself must still exist and still read the covered
	// input — otherwise the filter entry guards nothing.
	parityTest := filepath.Join(projectRoot, "internal", "template",
		"branch_protection_parity_test.go")
	paritySrc, err := os.ReadFile(parityTest)
	if err != nil {
		t.Fatalf("read %s: %v", parityTest, err)
	}
	if !strings.Contains(string(paritySrc), parityInput) {
		t.Fatalf("correspondence violated: %s no longer reads %s — the detect-filter entry guards nothing and must be re-derived",
			parityTest, parityInput)
	}
}
