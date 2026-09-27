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

// codexRoleUniformityViolations checks pred against every role in roles with
// three matched sets built from the role list itself: {} (nothing matched),
// {role} (only this role), and every OTHER role. A uniform predicate returns
// false, true, false respectively for every role. It returns one entry per
// violation, naming the role and the matched set that exposed it.
func codexRoleUniformityViolations(pred func(codexRoleLoadInput) bool, roles []string) []string {
	var out []string
	for _, r := range roles {
		others := map[string]bool{}
		for _, o := range roles {
			if o != r {
				others[o] = true
			}
		}
		checks := []struct {
			label   string
			matched map[string]bool
			want    bool
		}{
			{"empty", map[string]bool{}, false},
			{"self", map[string]bool{r: true}, true},
			{"others", others, false},
		}
		for _, c := range checks {
			if got := pred(codexRoleLoadInput{Role: r, Matched: c.matched}); got != c.want {
				out = append(out, fmt.Sprintf("%s/%s=%v", r, c.label, got))
			}
		}
	}
	return out
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
		// Run the PRODUCTION predicate through a uniformity check over all
		// twelve roles: a role is loaded exactly when its own name is in the
		// matched set, and never otherwise. A predicate that special-cases a
		// role by name — forcing it true, or forcing it false — violates the
		// check for that role (card t1260: the earlier form only inspected a
		// lambda defined in this test, so a branch planted in the production
		// predicate passed here).
		roles := make([]string, 0, len(cases))
		for _, c := range cases {
			roles = append(roles, c.role)
		}
		if v := codexRoleUniformityViolations(codexRoleLoadPredicate, roles); len(v) != 0 {
			t.Fatalf("codexRoleLoadPredicate is not uniform across roles: %v", v)
		}

		// Positive controls: the same check rejects both branch directions.
		forceTrue := func(in codexRoleLoadInput) bool {
			if in.Role == "manager-lead" || in.Role == "mission-governor" {
				return true
			}
			return in.Matched[in.Role]
		}
		forceFalse := func(in codexRoleLoadInput) bool {
			if in.Role == "manager-lead" || in.Role == "mission-governor" {
				return false
			}
			return in.Matched[in.Role]
		}
		if v := codexRoleUniformityViolations(forceTrue, roles); len(v) == 0 {
			t.Fatal("uniformity check failed to reject a force-true role-name branch")
		}
		if v := codexRoleUniformityViolations(forceFalse, roles); len(v) == 0 {
			t.Fatal("uniformity check failed to reject a force-false role-name branch")
		}
	})

	fmt.Printf("CONTRACT_NEUTRAL_LOAD_TRUE %d NONCE_TRUE %d\n", loadTrueCount, nonceTrueCount)
}
