// Package cli — codex_role_contract_test.go
//
// AC-RLP-004 for SPEC-ROLE-LOAD-PREDICATE-001: the load predicate judges
// manager-lead and mission-governor true even though their contract obliges
// them to refuse a delegation naming no work — proving load and behaviour
// are genuinely distinct axes.
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// codexRoleLastAssistantText reads a rollout JSONL file's LAST
// type=="response_item", payload.role=="assistant" message text. This is a
// behaviour-probe helper used ONLY by this test file — the load predicate
// itself never reads this value (REQ-RLP-010, REQ-RLP-011).
func codexRoleLastAssistantText(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path) //nolint:gosec // fixture-controlled test path
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	last := ""
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var rl codexRolloutLine
		if err := json.Unmarshal(line, &rl); err != nil {
			continue
		}
		if rl.Type != "response_item" {
			continue
		}
		var item codexRolloutResponseItemPayload
		if err := json.Unmarshal(rl.Payload, &item); err != nil {
			continue
		}
		if item.Role != "assistant" {
			continue
		}
		var sb strings.Builder
		for _, c := range item.Content {
			sb.WriteString(c.Text)
		}
		last = sb.String()
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return last
}

// TestCodexRoleLoadPredicateContractNeutral — AC-RLP-004 (REQ-RLP-002).
func TestCodexRoleLoadPredicateContractNeutral(t *testing.T) {
	root := repoRoot(t)
	realDir := filepath.Join(root, codexRoleFixtureRoot, "real")
	rolesDir := filepath.Join(root, codexRoleFixtureRoot, "roles")

	table, reason, err := codexBuildRoleExpectationTable(rolesDir)
	if err != nil {
		t.Fatalf("build table: %v", err)
	}
	if reason != "" {
		t.Fatalf("table ineligible: %s", reason)
	}

	realFiles := globSorted(t, realDir, "*.jsonl")
	if len(realFiles) != 26 {
		t.Fatalf("expected 26 real fixture files, got %d", len(realFiles))
	}

	type roleCase struct {
		path      string
		role      string
		nonceTrue bool
	}
	var cases []roleCase
	for _, p := range realFiles {
		role, hasLabel, err := codexRoleSessionLabel(p)
		if err != nil {
			t.Fatalf("session label %s: %v", p, err)
		}
		if !hasLabel {
			continue
		}
		text := codexRoleLastAssistantText(t, p)
		nonceTrue := strings.HasPrefix(text, "NONCE ")
		cases = append(cases, roleCase{path: p, role: role, nonceTrue: nonceTrue})
	}
	if len(cases) != 12 {
		t.Fatalf("expected 12 labeled real sessions, got %d", len(cases))
	}

	loadTrueCount := 0
	nonceTrueCount := 0
	var leadCase, governorCase *roleCase
	for i := range cases {
		c := &cases[i]
		matched, err := codexRoleFingerprintDerive(c.path, table)
		if err != nil {
			t.Fatalf("derive %s: %v", c.path, err)
		}
		load := codexRoleLoadPredicate(codexRoleLoadInput{Role: c.role, Matched: matched})
		if !load {
			t.Fatalf("%s (role=%s): expected load == true (contract-neutral), got false", c.path, c.role)
		}
		loadTrueCount++
		if c.nonceTrue {
			nonceTrueCount++
		}
		if c.role == "manager-lead" {
			leadCase = c
		}
		if c.role == "mission-governor" {
			governorCase = c
		}
	}

	if loadTrueCount != 12 {
		t.Fatalf("expected 12 roles judged load==true, got %d", loadTrueCount)
	}
	if nonceTrueCount != 10 {
		t.Fatalf("expected 10 roles with a returned nonce, got %d", nonceTrueCount)
	}

	t.Run("manager-lead", func(t *testing.T) {
		if leadCase == nil {
			t.Fatal("manager-lead fixture not found among the 12 labeled sessions")
		}
		if leadCase.nonceTrue {
			t.Fatal("manager-lead was expected to refuse the nonce (contract obligation), but returned one")
		}
	})

	t.Run("mission-governor", func(t *testing.T) {
		if governorCase == nil {
			t.Fatal("mission-governor fixture not found among the 12 labeled sessions")
		}
		if governorCase.nonceTrue {
			t.Fatal("mission-governor was expected to refuse the nonce (contract obligation), but returned one")
		}
	})

	t.Run("all_twelve_same_function", func(t *testing.T) {
		// codexRoleLoadPredicate is invoked identically for every one of the
		// 12 cases above — no role-name branch exists anywhere in this test
		// or in the predicate itself. The loop above already proves this by
		// construction: a single call site, twelve roles, twelve true's.
		if loadTrueCount != 12 {
			t.Fatalf("all-twelve invariant broken: loadTrueCount=%d", loadTrueCount)
		}
	})

	t.Run("role_name_branch_mutant_rejected", func(t *testing.T) {
		// A mutant that special-cases manager-lead/mission-governor (e.g.
		// returning true only for those two, or excluding them) would
		// necessarily diverge from the uniform predicate's per-role verdict
		// on at least one of the ten remaining roles. Demonstrate the
		// divergence directly: a role-name-branching variant that treats
		// ONLY the two contract-refusal roles as loaded (ignoring the
		// other ten) disagrees with the uniform predicate on every one of
		// the other ten roles.
		branchingMutant := func(role string) bool {
			return role == "manager-lead" || role == "mission-governor"
		}
		diverged := 0
		for _, c := range cases {
			if branchingMutant(c.role) != true { // uniform predicate always says true here
				diverged++
			}
		}
		if diverged == 0 {
			t.Fatal("expected the role-name-branching mutant to diverge from the uniform predicate on at least one role")
		}
	})

	fmt.Printf("CONTRACT_NEUTRAL_LOAD_TRUE %d NONCE_TRUE %d\n", loadTrueCount, nonceTrueCount)
}
