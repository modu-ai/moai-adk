package backup

import (
	"strings"
	"testing"
)

// Issue #1267 — llm.agent_overrides did not survive `moai update`.
//
// The template ships `agent_overrides: {}` as an empty placeholder. DeepMerge3Way
// iterated only the NEW map's keys, so once it recursed into that empty
// placeholder there were no keys left to walk and every user-authored entry was
// dropped. A 3-way merge has the information to tell the two cases apart:
//
//	old-only AND absent from base -> the user added it       -> preserve
//	old-only AND present in base  -> the template removed it -> drop
//
// These tests pin both directions, so a fix that simply carries every old key
// forward (resurrecting keys the template deliberately retired) fails too.
func TestDeepMerge3Way_PreservesUserAddedNestedKeys(t *testing.T) {
	tpl := []byte("llm:\n  profile: medium\n  agent_overrides: {}\n")
	user := []byte("llm:\n  profile: medium\n  agent_overrides:\n    manager-develop:\n      model: sonnet\n      effort: medium\n")

	got, err := MergeYAML3Way(tpl, user, tpl)
	if err != nil {
		t.Fatalf("MergeYAML3Way: %v", err)
	}
	if !strings.Contains(string(got), "manager-develop") {
		t.Fatalf("user-added agent_overrides entry was dropped\n--- merged ---\n%s", string(got))
	}
	if !strings.Contains(string(got), "sonnet") {
		t.Fatalf("user-added override value was dropped\n--- merged ---\n%s", string(got))
	}
}

// A key the template deliberately retired (present in old AND base, absent from
// new) must still be dropped — otherwise the fix resurrects deprecated config.
func TestDeepMerge3Way_StillDropsTemplateRemovedKeys(t *testing.T) {
	base := []byte("llm:\n  profile: medium\n  retired_key: keep-me-not\n")
	user := []byte("llm:\n  profile: medium\n  retired_key: keep-me-not\n")
	tpl := []byte("llm:\n  profile: medium\n")

	got, err := MergeYAML3Way(tpl, user, base)
	if err != nil {
		t.Fatalf("MergeYAML3Way: %v", err)
	}
	if strings.Contains(string(got), "retired_key") {
		t.Fatalf("template-removed key was resurrected\n--- merged ---\n%s", string(got))
	}
}

// A user value that diverges from the template default must still win — the
// pre-existing 3-way contract, kept as a regression guard for the fix.
func TestDeepMerge3Way_UserScalarStillWins(t *testing.T) {
	base := []byte("llm:\n  profile: medium\n")
	user := []byte("llm:\n  profile: low\n")
	tpl := []byte("llm:\n  profile: medium\n")

	got, err := MergeYAML3Way(tpl, user, base)
	if err != nil {
		t.Fatalf("MergeYAML3Way: %v", err)
	}
	if !strings.Contains(string(got), "profile: low") {
		t.Fatalf("user scalar override was lost\n--- merged ---\n%s", string(got))
	}
}
