package config

// Card t1702, iter5 F2 [P1]: the shipped `user-root:claude-skills/` and
// `user-root:claude-agents/` entries must cover the profile-derived roots the
// zone resolver emits as `user-root:<base>@<profile>/...`. The earlier parity
// test stopped at form generation and the tracked-file judgement and never
// matched a form against a SHIPPED entry, which is how the gap went unseen: a
// profile root's tracked files could be removed by Write, Edit or Bash.

import (
	"os"
	"path/filepath"
	"testing"
)

func shippedZoneEntries(t *testing.T) []ZoneEntry {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "template", "templates", ".moai", "config", "sections", "protected-zone.yaml"))
	if err != nil {
		t.Fatalf("read the shipped protected-zone manifest: %v", err)
	}
	entries, err := ParseProtectedZone(data, true, "shipped protected-zone.yaml")
	if err != nil {
		t.Fatalf("parse the shipped manifest: %v", err)
	}
	return entries
}

func anyUserRootEntryMatches(entries []ZoneEntry, form string) bool {
	for _, e := range entries {
		if e.Kind == ZoneUserRoot && e.Match(form) {
			return true
		}
	}
	return false
}

func TestShippedUserRootEntries_CoverProfileDerivedRoots(t *testing.T) {
	entries := shippedZoneEntries(t)
	userRootEntries := 0
	for _, e := range entries {
		if e.Kind == ZoneUserRoot {
			userRootEntries++
		}
	}
	if userRootEntries == 0 {
		t.Fatalf("the shipped manifest carries no user-root entry; the sweep below would assert nothing")
	}

	cases := []struct {
		form string
		want bool
		why  string
	}{
		// Controls: the fixed roots were covered before this repair and must stay so.
		{"user-root:claude-skills/moai/skill.md", true, "fixed skills root, a file"},
		{"user-root:claude-agents/manager-spec.md", true, "fixed agents root, a file"},
		{"user-root:claude-skills", true, "fixed skills root itself (a directory candidate arrives without the slash)"},
		// The defect: profile-derived roots.
		{"user-root:claude-skills@work/moai/skill.md", true, "profile skills root, a file"},
		{"user-root:claude-agents@work/manager-spec.md", true, "profile agents root, a file"},
		{"user-root:claude-skills@work", true, "profile skills root itself (rm -rf of the root)"},
		{"user-root:claude-skills@work/", true, "profile skills root with the trailing slash"},
		{"user-root:claude-agents@a-b/x/y.md", true, "a profile name with a hyphen"},
		// Not covered: only the two profile-derivable bases are.
		{"user-root:agents-skills@work/x.md", false, "agents-skills has no per-profile form"},
		{"user-root:codex-agents@work/x.md", false, "codex-agents has no per-profile form"},
		{"user-root:claude-skills-evil/x.md", false, "a different slug that merely shares a prefix"},
		{"user-root:claude-skills-evil@work/x.md", false, "a different base with a profile suffix"},
		{"user-root:unknown@work/x.md", false, "an unknown base"},
		{"user-root:@work/x.md", false, "an empty base"},
		{"user-root:claude-skills@/x.md", false, "an empty profile name"},
	}
	for _, c := range cases {
		if got := anyUserRootEntryMatches(entries, c.form); got != c.want {
			t.Errorf("%-52s covered = %v, want %v (%s)", c.form, got, c.want, c.why)
		}
	}
}

// An overlay entry is declared with a slug of [a-z0-9-] only, so a profile form
// cannot be declared separately; the base entry has to be what covers it.
func TestZoneUserRootMatch_ProfileFormUsesTheBaseEntry(t *testing.T) {
	e := ZoneEntry{Kind: ZoneUserRoot, Sub: ZoneDir, Pattern: "claude-skills/moai-demo/"}
	if !e.Match("user-root:claude-skills@work/moai-demo/skill.md") {
		t.Errorf("a narrower base entry must cover the same subtree under a profile root")
	}
	if e.Match("user-root:claude-skills@work/other/skill.md") {
		t.Errorf("a narrower base entry must not cover a different subtree under a profile root")
	}
	if !e.Match("user-root:claude-skills@work") {
		t.Errorf("removing the profile root removes the protected subtree below it (ancestor rule)")
	}
}
