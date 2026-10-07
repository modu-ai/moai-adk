package hook

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestProtectedZone groups the protected-zone guard tests as named subtests of one
// top-level test so a selector that matches nothing cannot pass for a sweep: every
// subtest prints its own --- PASS line and reports how many rows it swept.
func TestProtectedZone(t *testing.T) {
	t.Run("Normalization", testZoneNormalization)
	t.Run("FileTools", testZoneFileTools)
	t.Run("ShellMutation", testZoneShellMutation)
	t.Run("ShellQuoting", testZoneShellQuoting)
	t.Run("ManifestStates", testZoneManifestStates)
	t.Run("NonRegression", testZoneNonRegression)
	t.Run("DenyReason", testZoneDenyReason)
	t.Run("NoManifestReadForOthers", testZoneNoManifestReadForOthers)
	t.Run("AuditRow", testZoneAuditRow)
	t.Run("BaselineCovered", testZoneBaselineCovered)
	t.Run("Liveness", testZoneLiveness)
}

// testZoneNormalization is the platform-independent table: the pure lexical
// function carries its own absolute-path rule, so a drive-letter path is judged
// the same on a POSIX host as on Windows. The table is data, never conditional
// on runtime.GOOS.
func testZoneNormalization(t *testing.T) {
	cases := []struct {
		name       string
		root, in   string
		wantRel    string
		wantInside bool
	}{
		{"posix absolute", "/proj", "/proj/.claude/hooks/x", ".claude/hooks/x", true},
		{"posix relative", "/proj", ".claude/hooks/x", ".claude/hooks/x", true},
		{"dot prefix", "/proj", "./.claude/hooks/x", ".claude/hooks/x", true},
		{"dot-dot detour", "/proj", "docs/../.claude/hooks/x", ".claude/hooks/x", true},
		{"mixed case keeps display case", "/proj", "/PROJ/.CLAUDE/Hooks/x", ".CLAUDE/Hooks/x", true},
		{"windows drive absolute", "C:/proj", `C:\proj\.claude\hooks\x`, ".claude/hooks/x", true},
		{"windows drive mixed case", "C:/proj", `c:\PROJ\.CLAUDE\Hooks\x`, ".CLAUDE/Hooks/x", true},
		{"windows backslash dot prefix", "C:/proj", `.\.claude\hooks\x`, ".claude/hooks/x", true},
		{"windows root with backslashes", `C:\proj`, `C:\proj\a\b`, "a/b", true},
		{"windows drive on posix root is outside", "/proj", `C:\proj\.claude\hooks\x`, "", false},
		{"unc prefix is outside", "C:/proj", `//host/share/x`, "", false},
		{"unc prefix backslashes is outside", "C:/proj", `\\host\share\x`, "", false},
		{"other absolute", "/proj", "/other/x", "", false},
		{"sibling sharing a prefix is outside", "/proj", "/projx/a", "", false},
		{"escape via dot-dot", "/proj", "../x", "", false},
		{"escape via nested dot-dot", "/proj", "a/../../x", "", false},
		{"absolute dot-dot out of root", "/proj", "/proj/../x", "", false},
		{"root itself", "/proj", "/proj", ".", true},
		{"root with trailing slash", "/proj/", "/proj/a", "a", true},
		{"filesystem root as project root", "/", "/a/b", "a/b", true},
		{"empty root leaves absolute outside", "", "/a/b", "", false},
		{"empty path", "/proj", "", "", false},
		{"nfd korean path is nfc in the result", "/proj", "docs/\u1112\u1161\u11ab\u1100\u1173\u11af.md", "docs/한글.md", true},
		{"nfc korean root", "/한글", "/한글/a", "a", true},
	}
	swept := 0
	for _, c := range cases {
		swept++
		rel, inside := zoneLexicalRel(c.root, c.in)
		if rel != c.wantRel || inside != c.wantInside {
			t.Errorf("%s: zoneLexicalRel(%q, %q) = (%q, %v), want (%q, %v)",
				c.name, c.root, c.in, rel, inside, c.wantRel, c.wantInside)
		}
	}

	// every spelling of one hooks file folds to the same comparison form and is matched by
	// the same directory entry, which is the whole point of normalizing before matching
	entry, err := config.ParseProtectedZone([]byte("version: 1\ncategories:\n  x:\n    paths: [\".claude/hooks/\"]\n"), false, "t.yaml")
	if err != nil || len(entry) != 1 {
		t.Fatalf("fixture entry: %v", err)
	}
	for _, c := range []struct{ root, in string }{
		{"/proj", "/proj/.claude/hooks/x"},
		{"/proj", "./.claude/hooks/x"},
		{"/proj", "docs/../.claude/hooks/x"},
		{"/proj", ".CLAUDE/Hooks/x"},
		{"C:/proj", `C:\proj\.claude\hooks\x`},
		{"C:/proj", `.\.claude\hooks\x`},
	} {
		swept++
		rel, inside := zoneLexicalRel(c.root, c.in)
		if !inside || !entry[0].Match(config.FoldZoneText(rel)) {
			t.Errorf("%q under %q: rel=%q inside=%v, want a match of .claude/hooks/", c.in, c.root, rel, inside)
		}
	}
	for _, c := range []struct{ root, in string }{
		{"C:/proj", `//host/share/.claude/hooks/x`},
		{"/proj", "/projx/.claude/hooks/x"},
		{"/proj", "../.claude/hooks/x"},
	} {
		swept++
		if rel, inside := zoneLexicalRel(c.root, c.in); inside {
			t.Errorf("%q under %q is inside the zone as %q, want outside", c.in, c.root, rel)
		}
	}

	if swept < 33 {
		t.Fatalf("swept %d rows, want at least 33", swept)
	}
	t.Logf("swept=%d", swept)
}
