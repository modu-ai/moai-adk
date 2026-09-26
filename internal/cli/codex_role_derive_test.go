// Package cli — codex_role_derive_test.go
//
// AC-RLP-003 for SPEC-ROLE-LOAD-PREDICATE-001: fingerprint derivation, the
// label-only selection predicate, and the seven-arm discriminating power
// table (S/P/N1/N2/N3/N4/N5).
package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"
)

const codexRoleFixtureRoot = "internal/cli/testdata/codex-rollouts-t1171"

// TestCodexRoleBodyFingerprintDerive — AC-RLP-003 (REQ-RLP-001, 004, 007, 015).
func TestCodexRoleBodyFingerprintDerive(t *testing.T) {
	root := repoRoot(t)
	realDir := filepath.Join(root, codexRoleFixtureRoot, "real")
	synthDir := filepath.Join(root, codexRoleFixtureRoot, "synthetic")
	rolesDir := filepath.Join(root, codexRoleFixtureRoot, "roles")
	otherVersionDir := filepath.Join(root, codexRoleFixtureRoot, "roles-other-version")

	realFiles := globSorted(t, realDir, "*.jsonl")
	if len(realFiles) != 26 {
		t.Fatalf("expected 26 real fixture files, got %d", len(realFiles))
	}
	n3 := filepath.Join(synthDir, "n3-label-without-body.jsonl")
	n4 := filepath.Join(synthDir, "n4-crossed-roles.jsonl")

	pool := make([]string, 0, len(realFiles)+2)
	pool = append(pool, realFiles...)
	pool = append(pool, n3, n4)
	if len(pool) != 28 {
		t.Fatalf("expected a 28-file selection pool, got %d", len(pool))
	}

	table, reason, err := codexBuildRoleExpectationTable(rolesDir)
	if err != nil {
		t.Fatalf("build table: %v", err)
	}
	if reason != "" {
		t.Fatalf("table ineligible: %s", reason)
	}
	if len(table.ByRole) != 12 {
		t.Fatalf("expected 12 roles in the expectation table, got %d", len(table.ByRole))
	}

	otherTable, reason, err := codexBuildRoleExpectationTable(otherVersionDir)
	if err != nil {
		t.Fatalf("build other-version table: %v", err)
	}
	if reason != "" {
		t.Fatalf("other-version table ineligible: %s", reason)
	}

	// Selection: label presence ONLY (§A.0, REQ-RLP-015) — never body
	// agreement with the expectation table.
	var labeled []string
	labelOf := map[string]string{}
	for _, p := range pool {
		role, hasLabel, err := codexRoleSessionLabel(p)
		if err != nil {
			t.Fatalf("session label %s: %v", p, err)
		}
		if hasLabel {
			labeled = append(labeled, p)
			labelOf[p] = role
		}
	}

	t.Run("S_selection_by_label", func(t *testing.T) {
		if len(labeled) != 14 {
			t.Fatalf("expected 14 label-bearing sessions out of 28, got %d", len(labeled))
		}
		if labelOf[n3] == "" || labelOf[n4] == "" {
			t.Fatal("N3 and N4 must both carry a label and therefore be selected")
		}
	})

	realLabeled := func() []string {
		out := make([]string, 0, 12)
		for _, p := range labeled {
			if p == n3 || p == n4 {
				continue
			}
			out = append(out, p)
		}
		return out
	}()
	if len(realLabeled) != 12 {
		t.Fatalf("expected 12 real labeled sessions, got %d", len(realLabeled))
	}

	t.Run("P_all_twelve", func(t *testing.T) {
		positiveMatches := 0
		for _, p := range realLabeled {
			role := labelOf[p]
			matched, err := codexRoleFingerprintDerive(p, table)
			if err != nil {
				t.Fatalf("derive %s: %v", p, err)
			}
			if len(matched) != 1 || !matched[role] {
				t.Fatalf("%s (role=%s): expected matched set == {%s}, got %v", p, role, role, matched)
			}
			positiveMatches++
		}
		if positiveMatches != 12 {
			t.Fatalf("expected 12 positive matches, got %d", positiveMatches)
		}
		fmt.Printf("FINGERPRINT_POSITIVE_MATCHES %d\n", positiveMatches)
	})

	t.Run("N1_version_mismatch", func(t *testing.T) {
		count := 0
		for _, p := range realLabeled {
			role := labelOf[p]
			matched, err := codexRoleFingerprintDerive(p, otherTable)
			if err != nil {
				t.Fatalf("derive %s: %v", p, err)
			}
			if codexRoleLoadPredicate(codexRoleLoadInput{Role: role, Matched: matched}) {
				t.Fatalf("%s: expected false against a different-version expectation table, got true", p)
			}
			count++
		}
		if count != 12 {
			t.Fatalf("expected 12 version-mismatch checks, got %d", count)
		}
	})

	t.Run("N2_absent_roles", func(t *testing.T) {
		emptyDir := t.TempDir()
		emptyTable, reason, err := codexBuildRoleExpectationTable(emptyDir)
		if err != nil {
			t.Fatalf("build empty table: %v", err)
		}
		if reason != codexRoleTableReasonEmptyDir {
			t.Fatalf("expected empty_dir reason, got %q", reason)
		}
		count := 0
		for _, p := range realLabeled {
			role := labelOf[p]
			matched, err := codexRoleFingerprintDerive(p, emptyTable)
			if err != nil {
				t.Fatalf("derive %s: %v", p, err)
			}
			if codexRoleLoadPredicate(codexRoleLoadInput{Role: role, Matched: matched}) {
				t.Fatalf("%s: expected false against an empty expectation table, got true", p)
			}
			count++
		}
		if count != 12 {
			t.Fatalf("expected 12 absent-role checks, got %d", count)
		}
	})

	t.Run("N3_label_without_body", func(t *testing.T) {
		role := labelOf[n3]
		if role == "" {
			t.Fatal("N3 must carry a label")
		}
		matched, err := codexRoleFingerprintDerive(n3, table)
		if err != nil {
			t.Fatalf("derive: %v", err)
		}
		if codexRoleLoadPredicate(codexRoleLoadInput{Role: role, Matched: matched}) {
			t.Fatal("N3 (label present, matching body item removed) must judge false")
		}
	})

	t.Run("N4_crossed_roles", func(t *testing.T) {
		role := labelOf[n4]
		if role == "" {
			t.Fatal("N4 must carry a label")
		}
		matched, err := codexRoleFingerprintDerive(n4, table)
		if err != nil {
			t.Fatalf("derive: %v", err)
		}
		if codexRoleLoadPredicate(codexRoleLoadInput{Role: role, Matched: matched}) {
			t.Fatal("N4 (label A, body B) must judge false")
		}
	})

	t.Run("N5_real_parent_sessions", func(t *testing.T) {
		var unlabeled []string
		for _, p := range pool {
			if p == n3 || p == n4 {
				continue
			}
			_, hasLabel, err := codexRoleSessionLabel(p)
			if err != nil {
				t.Fatalf("session label %s: %v", p, err)
			}
			if !hasLabel {
				unlabeled = append(unlabeled, p)
			}
		}
		if len(unlabeled) != 14 {
			t.Fatalf("expected 14 unlabeled (parent) sessions, got %d", len(unlabeled))
		}
		falseCount := 0
		for _, p := range unlabeled {
			matched, err := codexRoleFingerprintDerive(p, table)
			if err != nil {
				t.Fatalf("derive %s: %v", p, err)
			}
			anyTrue := false
			for role := range table.ByRole {
				if codexRoleLoadPredicate(codexRoleLoadInput{Role: role, Matched: matched}) {
					anyTrue = true
					break
				}
			}
			if anyTrue {
				t.Fatalf("%s: a real parent session must judge false against every role", p)
			}
			falseCount++
		}
		fmt.Printf("PARENT_SESSIONS_FALSE %d\n", falseCount)
	})

	t.Run("hybrid_selection_mutant_rejected", func(t *testing.T) {
		// A selection predicate requiring BOTH label presence AND body
		// agreement with the expectation table would select only the 12
		// real matching sessions — excluding N3 (label, body absent) and N4
		// (label, body crossed) — collapsing 14 to 12. REQ-RLP-015 forbids
		// exactly this: selection is by label alone.
		hybridCount := 0
		for _, p := range labeled {
			role := labelOf[p]
			matched, err := codexRoleFingerprintDerive(p, table)
			if err != nil {
				t.Fatalf("derive %s: %v", p, err)
			}
			if matched[role] {
				hybridCount++
			}
		}
		if hybridCount != 12 {
			t.Fatalf("expected the body-agreement-combined selector to select 12, got %d", hybridCount)
		}
		if hybridCount == len(labeled) {
			t.Fatal("hybrid selector must diverge from label-only selection (14), but it matched all 14")
		}
	})

	fmt.Printf("SELECTED_BY_LABEL %d OF %d\n", len(labeled), len(pool))
}

func globSorted(t *testing.T, dir, pattern string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		t.Fatalf("glob %s/%s: %v", dir, pattern, err)
	}
	sort.Strings(matches)
	return matches
}
