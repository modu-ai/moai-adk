package template

// retired_keys_reship_test.go — SPEC-WEB-AGENTFM-RESTORE-001 M1: the
// re-shipped console keys must re-enter ShippedRetiredModelKeys (membership =
// strip exclusion — retired_model_keys.go reads the embedded template) while
// the still-retired keys stay strippable (REQ-AFR-008, plan §F M1).

import "testing"

// TestShippedRetiredModelKeys_IncludesReshippedConsoleKeys pins the seam
// membership for the two keys the restored console surface writes
// (llm.profile, llm.agent_overrides): while the embedded template ships them,
// `moai update` must NOT strip them. performance_tier / profiles /
// harness_agents stay outside the template and therefore outside this set.
func TestShippedRetiredModelKeys_IncludesReshippedConsoleKeys(t *testing.T) {
	shipped, err := ShippedRetiredModelKeys()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"llm.profile", "llm.agent_overrides"} {
		if !shipped[key] {
			t.Errorf("shipped set lacks %q — the template does not re-ship the console key yet, so moai update would strip a key the console writes", key)
		}
	}
	for _, key := range []string{"llm.performance_tier", "llm.profiles", "llm.harness_agents"} {
		if shipped[key] {
			t.Errorf("shipped set carries %q — that key stays retired and strippable (plan §D.1 REQ-AMI-013 narrowing)", key)
		}
	}
}
