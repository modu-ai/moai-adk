// branch_protection_parity_test.go: CI guard for branch-protection template
// parity. Verifies that the local authoritative copy
// (.github/branch-protection.json.gtmpl) and the user-distributed template
// mirror (internal/template/templates/.github/branch-protection.json.gtmpl)
// stay byte-identical. The two files were byte-identical at Wave-3 hardening
// time (2026-06-24) but had NO parity assertion, risking silent drift: a future
// edit to one tree without the other would ship a stale or divergent
// branch-protection rule to user projects via moai init / moai update.
//
// Sentinel on failure: BRANCH_PROTECTION_PARITY_DRIFT
// Origin: Wave-3 .github hardening (#20).
//
// Design notes:
//   - This is a _test.go file under internal/template/ (NOT under
//     internal/template/templates/), so it does NOT trigger the template-
//     neutrality CI guard (.github/workflows/template-neutrality-check.yaml
//     paths filter) and is exempt per CLAUDE.local.md §25 (_test.go files are
//     not user-distributed template content).
//   - The test reads both files relative to the project root (located via the
//     shared findProjectRootForMirrorTest helper) and asserts byte equality,
//     naming both paths in the failure message so the fix (cp one onto the
//     other) is self-evident from CI output.
//   - Mirrors the byte-parity pattern established by TestRuleTemplateMirrorDrift
//     (rule_template_mirror_test.go) but scoped to a single high-value file pair
//     rather than an allowlist — branch-protection is one file, not a set.
//
// @MX:NOTE: [AUTO] fan_in=2 — guards the branch-protection template mirror
// invariant. Touching either tree's branch-protection.json.gtmpl without the
// other breaks user-project branch-protection parity.
package template_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// branchProtectionRelPath is the repo-relative path of the authoritative local
// copy (source of truth for branch-protection rule).
const branchProtectionRelPath = ".github/branch-protection.json.gtmpl"

// branchProtectionMirrorRelPath is the repo-relative path of the template
// mirror that go:embed ships to user projects via moai init / moai update.
const branchProtectionMirrorRelPath = "internal/template/templates/.github/branch-protection.json.gtmpl"

// TestBranchProtectionParity asserts that the local authoritative
// branch-protection.json.gtmpl and its template mirror are byte-identical.
//
// On drift, emits BRANCH_PROTECTION_PARITY_DRIFT with both absolute paths and
// byte counts so the CI log shows exactly which file to copy onto which.
func TestBranchProtectionParity(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)

	srcPath := filepath.Join(projectRoot, branchProtectionRelPath)
	mirrorPath := filepath.Join(projectRoot, branchProtectionMirrorRelPath)

	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("BRANCH_PROTECTION_PARITY_DRIFT: source file unreadable %s: %v", srcPath, err)
	}

	mirrorContent, mirrorErr := os.ReadFile(mirrorPath)
	if mirrorErr != nil {
		if os.IsNotExist(mirrorErr) {
			t.Errorf(
				"BRANCH_PROTECTION_PARITY_DRIFT: source %s has no mirror at %s; "+
					"run 'cp %s %s' and stage both files",
				branchProtectionRelPath, mirrorPath, srcPath, mirrorPath,
			)
			return
		}
		t.Fatalf("BRANCH_PROTECTION_PARITY_DRIFT: mirror file unreadable %s: %v", mirrorPath, mirrorErr)
	}

	if !bytes.Equal(srcContent, mirrorContent) {
		t.Errorf(
			"BRANCH_PROTECTION_PARITY_DRIFT: source %s differs from its mirror %s "+
				"(source %d bytes, mirror %d bytes); run 'cp %s %s' and stage both files",
			branchProtectionRelPath, mirrorPath,
			len(srcContent), len(mirrorContent),
			srcPath, mirrorPath,
		)
	}
}
