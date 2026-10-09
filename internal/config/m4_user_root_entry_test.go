package config

// m4_user_root_entry_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M4: the
// user-root entry kind (REQ-GRD-001). The loader's repository-relative
// contract is respected WITH the prefix: an absolute path is rejected
// whether or not it carries "user-root:", and .. segments stay rejected.

import (
	"strings"
	"testing"
)

func TestParseZoneEntryUserRootKinds(t *testing.T) {
	cases := []struct {
		raw         string
		wantSub     ZoneEntryKind
		wantPattern string
	}{
		{"user-root:claude-agents/moai/", ZoneDir, "claude-agents/moai/"},
		{"user-root:claude-skills/", ZoneDir, "claude-skills/"},
		{"user-root:claude-agents/moai/plan-auditor.md", ZoneExact, "claude-agents/moai/plan-auditor.md"},
		{"user-root:claude-agents/moai/*", ZonePrefix, "claude-agents/moai/"},
		{"user-root:agents-skills/", ZoneDir, "agents-skills/"},
	}
	for _, c := range cases {
		e, err := parseZoneEntry(c.raw)
		if err != nil {
			t.Fatalf("parseZoneEntry(%q): %v", c.raw, err)
		}
		if e.Kind != ZoneUserRoot {
			t.Fatalf("%q: kind = %v, want ZoneUserRoot", c.raw, e.Kind)
		}
		if e.Sub != c.wantSub {
			t.Errorf("%q: sub = %v, want %v", c.raw, e.Sub, c.wantSub)
		}
		if e.Pattern != c.wantPattern {
			t.Errorf("%q: pattern = %q, want %q", c.raw, e.Pattern, c.wantPattern)
		}
	}
}

func TestParseZoneEntryUserRootRejections(t *testing.T) {
	for _, raw := range []string{
		"user-root:/etc/passwd",              // absolute under the prefix — still rejected
		"user-root:claude-agents/C:/x",       // drive-letter absolute
		"user-root:claude-agents/../secrets", // .. escape
		"user-root:/",                        // bare absolute
		"user-root:",                         // no slug
		"user-root:CLAude-agents/",           // slug must be [a-z0-9-]
	} {
		if _, err := parseZoneEntry(raw); err == nil {
			t.Fatalf("parseZoneEntry(%q) accepted — the repository-relative contract must hold under the user-root prefix", raw)
		} else if !strings.Contains(err.Error(), raw) {
			t.Errorf("the rejection does not name the entry: %v", err)
		}
	}
}

func TestZoneUserRootMatch(t *testing.T) {
	dir, err := parseZoneEntry("user-root:claude-agents/moai/")
	if err != nil {
		t.Fatal(err)
	}
	if !dir.Match(FoldZoneText("user-root:claude-agents/moai/plan-auditor.md")) {
		t.Error("the dir entry missed its namespaced form")
	}
	if dir.Match(FoldZoneText("claude-agents/moai/plan-auditor.md")) {
		t.Error("the dir entry matched a NON-namespaced form — repo entries must not see user-root forms")
	}
	if dir.Match(FoldZoneText("user-root:agents-skills/x")) {
		t.Error("the dir entry crossed into another root's namespace")
	}
	exact, err := parseZoneEntry("user-root:claude-agents/moai/plan-auditor.md")
	if err != nil {
		t.Fatal(err)
	}
	if exact.Match(FoldZoneText("user-root:claude-agents/moai/other.md")) {
		t.Error("the exact entry matched a different file")
	}
}
