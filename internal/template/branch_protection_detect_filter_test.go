package template

import (
	"os"
	"path/filepath"
	"regexp"
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
	const sharedFilter = ".github/test-input-filters.yml"
	filterRef := regexp.MustCompile(`(?m)^\s+filters:\s*` + regexp.QuoteMeta(sharedFilter) + `\s*$`)
	if !filterRef.MatchString(content) {
		t.Fatalf("%s does not consume the canonical filter %s", ciYml, sharedFilter)
	}
	filterPath := filepath.Join(projectRoot, filepath.FromSlash(sharedFilter))
	raw, err = os.ReadFile(filterPath)
	if err != nil {
		t.Fatalf("read %s: %v", filterPath, err)
	}
	content = string(raw)
	lines := strings.Split(content, "\n")

	// Same literal the parity test reads (branchProtectionRelPath) — kept
	// inline to avoid coupling this guard to the other test file's symbols.
	const parityInput = ".github/branch-protection.json.gtmpl"
	// GATE fix: bare strings.Contains passes even when the filter ENTRY is
	// commented out (the path string survives in the comment). Assert a
	// LIST-ENTRY line: optional indent, a `- ` dash, then the quoted path —
	// a commented entry (`# - '...'`) starts with # and cannot match.
	entryRe := regexp.MustCompile(`(?m)^\s*-\s*['"]` + regexp.QuoteMeta(parityInput) + `['"]\s*$`)
	// GATE P2 (round 4): the entry must live INSIDE the go_code filter
	// block — an entry moved under another filter (or to file top level)
	// routes the edit to the wrong job set while a whole-file match still
	// passes. Extract the go_code block by indent scan and match the entry
	// against that block only. The block's `go_code:` key carries no value
	// (the `go_code: ${{ ... }}` output mapping is a different line shape).
	goStart, goIndent := -1, -1
	for i, ln := range lines {
		trimmed := strings.TrimLeft(ln, " ")
		if strings.HasPrefix(trimmed, "go_code:") && strings.TrimSpace(strings.TrimPrefix(trimmed, "go_code:")) == "" {
			goStart, goIndent = i, len(ln)-len(trimmed)
			break
		}
	}
	if goStart < 0 {
		t.Fatalf("go_code filter not found in %s — the detect filter is renamed or moved; re-derive this guard", filterPath)
	}
	var goBlock []string
	for _, ln := range lines[goStart+1:] {
		if strings.TrimSpace(ln) == "" {
			continue // blank lines do not end a YAML block
		}
		if len(ln)-len(strings.TrimLeft(ln, " ")) <= goIndent {
			break
		}
		goBlock = append(goBlock, ln)
	}
	if !entryRe.MatchString(strings.Join(goBlock, "\n")) {
		t.Fatalf("detect-filter correspondence violated: %s go_code filter does not cover %s — a PR editing it alone would skip the branch-protection parity guard (AC-CI-012)",
			filterPath, parityInput)
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
