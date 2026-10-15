package hook

// role_rules_stamp_test.go — card t1619, codex review P2 (role rules version stamp):
// the template_version stamp lives under the moai section of system.yaml. A trailing
// YAML comment is not part of the value, and a commented-out occurrence is not the
// stamp. The stamp is read as YAML, not located by a string search.

import "testing"

func TestRoleRulesVersionSkewReadsNestedMoaiStamp(t *testing.T) {
	skewed, rule, ok := detectRoleRuleVersionSkew("moai:\n  template_version: v3.0.0 # 설명\n", "v3.1.3")
	if !ok || !skewed || rule != "v3.0.0" {
		t.Fatalf("nested stamp with a trailing comment: skewed=%v rule=%q ok=%v, want skewed, rule v3.0.0, ok", skewed, rule, ok)
	}

	skewed, rule, ok = detectRoleRuleVersionSkew("# template_version: v9.9.9\nmoai:\n  template_version: v3.1.3\n", "v3.1.3")
	if !ok || skewed || rule != "v3.1.3" {
		t.Fatalf("a commented-out occurrence must not be the stamp: skewed=%v rule=%q ok=%v, want no skew, rule v3.1.3", skewed, rule, ok)
	}

	if _, _, ok := detectRoleRuleVersionSkew("hook:\n  x: 1\n", "v3.1.3"); ok {
		t.Fatalf("a system.yaml without a moai stamp must be skipped (ok=false)")
	}
}
