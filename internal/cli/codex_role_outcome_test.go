// Package cli — codex_role_outcome_test.go
//
// Card t1260 (t1171 sync-audit F5): a session record in which the scan sees
// no developer response_item at all must read as "not measured", distinct
// from a measured record whose role body is absent. The load predicate alone
// returns false for both, so the distinction lives in codexRoleLoadOutcome.
//
// This file deliberately prints nothing: the AC-RLP judges reject any test
// output carrying the not-measured token.
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCodexRoleLoadOutcomeSeparatesUnmeasured(t *testing.T) {
	root := repoRoot(t)
	rolesDir := filepath.Join(root, codexRoleFixtureRoot, "roles")
	table, reason, err := codexBuildRoleExpectationTable(rolesDir)
	if err != nil {
		t.Fatalf("build table: %v", err)
	}
	if reason != "" {
		t.Fatalf("table ineligible: %s", reason)
	}

	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
		return p
	}
	// Labeled session with an assistant reply but no developer item.
	noDeveloper := write("no-developer.jsonl",
		`{"type":"session_meta","payload":{"source":{"subagent":{"thread_spawn":{"agent_role":"builder-harness"}}}}}`+"\n"+
			`{"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"NONCE x"}]}}`+"\n")
	empty := write("empty.jsonl", "")
	n3 := filepath.Join(root, codexRoleFixtureRoot, "synthetic", "n3-label-without-body.jsonl")

	observe := func(p string) codexRoleFingerprintObservation {
		obs, err := codexRoleFingerprintObserve(p, table)
		if err != nil {
			t.Fatalf("observe %s: %v", p, err)
		}
		return obs
	}

	t.Run("zero_developer_items", func(t *testing.T) {
		obs := observe(noDeveloper)
		if obs.DeveloperItems != 0 || len(obs.Matched) != 0 {
			t.Fatalf("expected an unmeasured observation, got %+v", obs)
		}
		if got := codexRoleLoadOutcome("builder-harness", obs); got != codexRoleLoadOutcomeNotRun {
			t.Fatalf("outcome = %q, want %q", got, codexRoleLoadOutcomeNotRun)
		}
	})

	t.Run("empty_record", func(t *testing.T) {
		if got := codexRoleLoadOutcome("builder-harness", observe(empty)); got != codexRoleLoadOutcomeNotRun {
			t.Fatalf("outcome = %q, want %q", got, codexRoleLoadOutcomeNotRun)
		}
	})

	t.Run("measured_body_absent", func(t *testing.T) {
		obs := observe(n3)
		if obs.DeveloperItems == 0 {
			t.Fatal("N3 fixture must still carry developer items for this arm to be a measured FAIL")
		}
		if got := codexRoleLoadOutcome("builder-harness", obs); got != codexRoleLoadOutcomeFail {
			t.Fatalf("outcome = %q, want %q", got, codexRoleLoadOutcomeFail)
		}
	})

	t.Run("measured_body_present", func(t *testing.T) {
		passes := 0
		for _, p := range globSorted(t, filepath.Join(root, codexRoleFixtureRoot, "real"), "*.jsonl") {
			role, hasLabel, err := codexRoleSessionLabel(p)
			if err != nil {
				t.Fatalf("session label %s: %v", p, err)
			}
			if !hasLabel {
				continue
			}
			if got := codexRoleLoadOutcome(role, observe(p)); got != codexRoleLoadOutcomePass {
				t.Fatalf("%s (role=%s): outcome = %q, want %q", p, role, got, codexRoleLoadOutcomePass)
			}
			passes++
		}
		if passes != 12 {
			t.Fatalf("expected 12 measured passes, got %d", passes)
		}
	})

	t.Run("predicate_alone_conflates", func(t *testing.T) {
		// Control: the boolean predicate gives the same answer for the
		// unmeasured and the measured-absent records — the reason the
		// outcome carries a third value.
		a := codexRoleLoadPredicate(codexRoleLoadInput{Role: "builder-harness", Matched: observe(noDeveloper).Matched})
		b := codexRoleLoadPredicate(codexRoleLoadInput{Role: "builder-harness", Matched: observe(n3).Matched})
		if a || b {
			t.Fatalf("expected false for both records, got unmeasured=%v measured-absent=%v", a, b)
		}
	})
}
