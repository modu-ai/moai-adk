package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// zoneSevenCats renders the seven required categories with empty lists, the form
// a valid file at the shipped path must carry.
func zoneSevenCats() string {
	var b strings.Builder
	for _, c := range ProtectedZoneRequiredCategories {
		b.WriteString("  " + c + ":\n    paths: []\n")
	}
	return b.String()
}

func zoneShipped(extra string) string {
	return "version: 1\ncategories:\n" + zoneSevenCats() + extra
}

func writeZoneFile(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// swept counts the table rows a subtest actually evaluated, so an empty table
// cannot pass for a sweep.
func zoneSweptFloor(t *testing.T, swept, floor int) {
	t.Helper()
	if swept < floor {
		t.Fatalf("swept %d rows, want at least %d", swept, floor)
	}
	t.Logf("swept=%d", swept)
}

func TestProtectedZone(t *testing.T) {
	t.Run("Load", func(t *testing.T) {
		swept := 0

		// absent: neither file exists
		got := LoadProtectedZone(t.TempDir())
		swept++
		if got.State != ZoneStateAbsent || len(got.Zone.Entries) != 0 {
			t.Errorf("absent: state=%q entries=%d, want absent/0", got.State, len(got.Zone.Entries))
		}

		// valid shipped manifest: entries, kinds, runtime flag, source, order
		root := t.TempDir()
		writeZoneFile(t, root, ProtectedZoneShippedRel, "version: 1\ncategories:\n"+
			"  safety_guards:\n    paths: [\".claude/hooks/\"]\n    runtime_paths: [\".claude/settings.json\"]\n"+
			"  gate_policy:\n    paths: [\".moai/config/sections/\"]\n"+
			"  auditor:\n    paths: [\".claude/skills/moai-*\", \"**/CLAUDE.md\"]\n"+
			"  regression_tests:\n    paths: []\n  apply_rollback:\n    paths: []\n  budgets: {}\n  logs:\n    paths: []\n")
		got = LoadProtectedZone(root)
		swept++
		if got.State != ZoneStateOK {
			t.Fatalf("valid: state=%q err=%v", got.State, got.Err)
		}
		type want struct {
			cat, pat string
			kind     ZoneEntryKind
			runtime  bool
		}
		wants := []want{
			{"safety_guards", ".claude/hooks/", ZoneDir, false},
			{"safety_guards", ".claude/settings.json", ZoneExact, true},
			{"gate_policy", ".moai/config/sections/", ZoneDir, false},
			{"auditor", ".claude/skills/moai-", ZonePrefix, false},
			{"auditor", "claude.md", ZoneBaseGlob, false},
		}
		if len(got.Zone.Entries) != len(wants) {
			t.Fatalf("entries=%d want %d: %+v", len(got.Zone.Entries), len(wants), got.Zone.Entries)
		}
		for i, w := range wants {
			e := got.Zone.Entries[i]
			swept++
			if e.Category != w.cat || e.Pattern != w.pat || e.Kind != w.kind || e.Runtime != w.runtime || e.Source != ProtectedZoneShippedRel {
				t.Errorf("entry %d = %+v, want %+v", i, e, w)
			}
		}

		// the manifest is what is read: entries a hard-coded list cannot know
		root = t.TempDir()
		writeZoneFile(t, root, ProtectedZoneShippedRel, zoneShipped("  probe_docs:\n    paths: [\"docs/\"]\n"))
		got = LoadProtectedZone(root)
		swept++
		found := false
		for _, e := range got.Zone.Entries {
			if e.Category == "probe_docs" && e.Pattern == "docs/" {
				found = true
			}
		}
		if got.State != ZoneStateOK || !found {
			t.Errorf("shipped docs/ entry not loaded: state=%q found=%v", got.State, found)
		}

		// present-but-unreadable: a directory where the file should be
		root = t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(ProtectedZoneShippedRel)), 0o755); err != nil {
			t.Fatal(err)
		}
		got = LoadProtectedZone(root)
		swept++
		if got.State != ZoneStateInvalid || got.InvalidFile != ProtectedZoneShippedRel {
			t.Errorf("unreadable: state=%q file=%q", got.State, got.InvalidFile)
		}
		zoneSweptFloor(t, swept, 9)
	})

	t.Run("Overlay", func(t *testing.T) {
		swept := 0

		// overlay only: the shipped file is absent, the overlay still loads
		root := t.TempDir()
		writeZoneFile(t, root, ProtectedZoneOverlayRel, "version: 1\ncategories:\n  probe_docs:\n    paths: [\"docs/\"]\n")
		got := LoadProtectedZone(root)
		swept++
		if got.State != ZoneStateOK || len(got.Zone.Entries) != 1 || got.Zone.Entries[0].Source != ProtectedZoneOverlayRel {
			t.Errorf("overlay only: %+v", got)
		}

		// overlay adds, never replaces (the overlay-cannot-narrow unit mirror)
		root = t.TempDir()
		writeZoneFile(t, root, ProtectedZoneShippedRel, zoneShipped("  probe_base:\n    paths: [\"base_dir/\"]\n"))
		writeZoneFile(t, root, ProtectedZoneOverlayRel, "version: 1\ncategories:\n  probe_docs:\n    paths: [\"docs/\"]\n")
		got = LoadProtectedZone(root)
		swept++
		var order []string
		for _, e := range got.Zone.Entries {
			order = append(order, e.Pattern)
		}
		if got.State != ZoneStateOK || strings.Join(order, ",") != "base_dir/,docs/" {
			t.Errorf("union order = %v, state=%q (want shipped first, then overlay)", order, got.State)
		}

		// an overlay may not narrow: an unknown key invalidates the overlay, and the
		// failing file is the overlay
		root = t.TempDir()
		writeZoneFile(t, root, ProtectedZoneShippedRel, zoneShipped("  probe_base:\n    paths: [\"base_dir/\"]\n"))
		writeZoneFile(t, root, ProtectedZoneOverlayRel, "version: 1\nexclude:\n  - base_dir/\ncategories:\n  probe_docs:\n    paths: [\"docs/\"]\n")
		got = LoadProtectedZone(root)
		swept++
		if got.State != ZoneStateInvalid || got.InvalidFile != ProtectedZoneOverlayRel {
			t.Errorf("overlay exclude: state=%q file=%q", got.State, got.InvalidFile)
		}

		// a zero-byte overlay is valid and adds nothing; an overlay is not bound by the seven
		root = t.TempDir()
		writeZoneFile(t, root, ProtectedZoneShippedRel, zoneShipped(""))
		writeZoneFile(t, root, ProtectedZoneOverlayRel, "")
		got = LoadProtectedZone(root)
		swept++
		if got.State != ZoneStateOK {
			t.Errorf("zero-byte overlay: state=%q err=%v", got.State, got.Err)
		}

		// an invalid overlay with a valid shipped manifest fails closed
		root = t.TempDir()
		writeZoneFile(t, root, ProtectedZoneShippedRel, zoneShipped(""))
		writeZoneFile(t, root, ProtectedZoneOverlayRel, "::: not [valid yaml")
		got = LoadProtectedZone(root)
		swept++
		if got.State != ZoneStateInvalid || got.InvalidFile != ProtectedZoneOverlayRel {
			t.Errorf("garbage overlay: state=%q file=%q", got.State, got.InvalidFile)
		}

		// both invalid: the shipped file is reported first
		root = t.TempDir()
		writeZoneFile(t, root, ProtectedZoneShippedRel, "::: not [valid yaml")
		writeZoneFile(t, root, ProtectedZoneOverlayRel, "::: not [valid yaml")
		got = LoadProtectedZone(root)
		swept++
		if got.State != ZoneStateInvalid || got.InvalidFile != ProtectedZoneShippedRel {
			t.Errorf("both invalid: state=%q file=%q", got.State, got.InvalidFile)
		}
		zoneSweptFloor(t, swept, 6)
	})

	t.Run("Validation", func(t *testing.T) {
		swept := 0
		long := strings.Repeat("a", 33)
		cases := []struct {
			name    string
			shipped bool
			doc     string
			valid   bool
		}{
			{"valid shipped with extra category", true, zoneShipped("  extra_cat:\n    paths: [\"x/\"]\n"), true},
			{"duplicate entry is valid", true, zoneShipped("  extra_cat:\n    paths: [\"x/\", \"x/\"]\n"), true},
			{"empty category list is valid", true, zoneShipped("  extra_cat:\n    paths: []\n"), true},
			{"basename glob with class", true, zoneShipped("  extra_cat:\n    paths: [\"**/*_test.go\", \"**/[a-c]x.md\"]\n"), true},
			{"raw prefix", true, zoneShipped("  extra_cat:\n    paths: [\"internal/harness/retention*\"]\n"), true},
			{"exact path", true, zoneShipped("  extra_cat:\n    paths: [\".claude/settings.json\"]\n"), true},
			{"middle wildcard", true, zoneShipped("  extra_cat:\n    paths: [\"a/*/b\"]\n"), false},
			{"question mark outside basename glob", true, zoneShipped("  extra_cat:\n    paths: [\"a?b\"]\n"), false},
			{"class outside basename glob", true, zoneShipped("  extra_cat:\n    paths: [\"a[b]\"]\n"), false},
			{"basename glob containing slash", true, zoneShipped("  extra_cat:\n    paths: [\"**/a/b\"]\n"), false},
			{"basename glob empty pattern", true, zoneShipped("  extra_cat:\n    paths: [\"**/\"]\n"), false},
			{"basename glob bad class", true, zoneShipped("  extra_cat:\n    paths: [\"**/[a\"]\n"), false},
			{"bare star prefix", true, zoneShipped("  extra_cat:\n    paths: [\"*\"]\n"), false},
			{"dot-dot segment", true, zoneShipped("  extra_cat:\n    paths: [\"a/../b\"]\n"), false},
			{"absolute entry", true, zoneShipped("  extra_cat:\n    paths: [\"/etc/x\"]\n"), false},
			{"drive absolute entry", true, zoneShipped("  extra_cat:\n    paths: [\"C:/x\"]\n"), false},
			{"backslash entry", true, zoneShipped("  extra_cat:\n    paths: [\"a\\\\b\"]\n"), false},
			{"empty entry", true, zoneShipped("  extra_cat:\n    paths: [\"\"]\n"), false},
			{"unknown top-level key", true, "version: 1\nexclude: []\ncategories:\n" + zoneSevenCats(), false},
			{"unknown category key", true, zoneShipped("  extra_cat:\n    paths: []\n    deny: [\"x\"]\n"), false},
			{"wrong version", true, "version: 2\ncategories:\n" + zoneSevenCats(), false},
			{"missing version", true, "categories:\n" + zoneSevenCats(), false},
			{"category name too long", true, zoneShipped("  " + long + ":\n    paths: []\n"), false},
			{"category name upper case", true, zoneShipped("  Bad:\n    paths: []\n"), false},
			{"category name non-ascii", true, zoneShipped("  카테고리:\n    paths: []\n"), false},
			{"category name 32 bytes", true, zoneShipped("  " + long[:32] + ":\n    paths: []\n"), true},
			{"duplicate category", true, zoneShipped("  logs:\n    paths: []\n"), false},
			{"paths is a scalar", true, zoneShipped("  extra_cat:\n    paths: x\n"), false},
			{"categories is a list", true, "version: 1\ncategories: [a]\n", false},
			{"shipped zero-byte", true, "", false},
			{"shipped empty categories", true, "version: 1\ncategories: {}\n", false},
			{"shipped one category only", true, "version: 1\ncategories:\n  logs:\n    paths: [\"x/\"]\n", false},
			{"shipped missing one of seven", true, "version: 1\ncategories:\n" + strings.Replace(zoneSevenCats(), "  budgets:\n    paths: []\n", "", 1), false},
			{"overlay zero-byte", false, "", true},
			{"overlay comment-only", false, "# nothing\n", true},
			{"overlay one category", false, "version: 1\ncategories:\n  probe:\n    paths: [\"docs/\"]\n", true},
			{"overlay unknown key", false, "version: 1\nexclude: [x]\ncategories: {}\n", false},
			{"shipped second YAML document", true, "version: 1\ncategories:\n" + zoneSevenCats() + "---\nx: [", false},
		}
		for _, c := range cases {
			swept++
			entries, err := ParseProtectedZone([]byte(c.doc), c.shipped, "x.yaml")
			if (err == nil) != c.valid {
				t.Errorf("%s: valid=%v err=%v", c.name, err == nil, err)
				continue
			}
			if err != nil && entries != nil {
				t.Errorf("%s: invalid manifest returned entries", c.name)
			}
		}

		// folding: an entry is stored NFC + ASCII lower case, and Match is a plain comparison
		entries, err := ParseProtectedZone([]byte(zoneShipped("  extra_cat:\n    paths: [\".Claude/Hooks/\", \"**/*_TEST.go\", \"Docs/한글.md\"]\n")), true, "x.yaml")
		swept++
		if err != nil {
			t.Fatal(err)
		}
		e := map[string]ZoneEntry{}
		for _, en := range entries {
			e[en.Raw] = en
		}
		for raw, tc := range map[string]struct {
			rel  string
			want bool
		}{
			".Claude/Hooks/": {".claude/hooks/moai/x.sh", true},
			"**/*_TEST.go":   {"a/b/c_test.go", true},
			"Docs/한글.md":     {"docs/한글.md", true},
		} {
			swept++
			if got := e[raw].Match(FoldZoneText(tc.rel)); got != tc.want {
				t.Errorf("Match(%q) for %q = %v, want %v", tc.rel, raw, got, tc.want)
			}
		}
		// NFD input folds to the same NFC form as an NFC entry
		swept++
		if FoldZoneText("한글") != FoldZoneText("\u1112\u1161\u11ab\u1100\u1173\u11af") {
			t.Errorf("NFD and NFC forms of the same Korean text fold differently")
		}
		// exact entries do not behave as prefixes; prefixes are raw strings; basename globs match only the final segment
		swept++
		exact := ZoneEntry{Kind: ZoneExact, Pattern: "a/b.txt"}
		dirE := ZoneEntry{Kind: ZoneDir, Pattern: "a/"}
		pre := ZoneEntry{Kind: ZonePrefix, Pattern: "a/retention"}
		glob := ZoneEntry{Kind: ZoneBaseGlob, Pattern: "*_test.go"}
		if exact.Match("a/b.txt/x") || !exact.Match("a/b.txt") || dirE.Match("a") || !dirE.Match("a/x") ||
			!pre.Match("a/retention_x.go") || pre.Match("a/other") || glob.Match("x_test.go/y") || !glob.Match("x/y_test.go") {
			t.Errorf("entry kind semantics diverge from the grammar")
		}
		// normalization: a dot-prefixed entry is stored cleaned, so a target written
		// the ordinary way still matches it (merge-gate round 1 P2-1, observed red first)
		entries, err = ParseProtectedZone([]byte(zoneShipped("  extra_cat:\n    paths: [\"./Docs/\", \"./x/./y/\"]\n")), true, "x.yaml")
		swept++
		if err != nil {
			t.Fatal(err)
		}
		if entries[0].Pattern != "docs/" || !entries[0].Match(FoldZoneText("docs/a.txt")) {
			t.Errorf("dot-prefixed dir entry: pattern=%q, want docs/ matching docs/a.txt", entries[0].Pattern)
		}
		if entries[1].Pattern != "x/y/" || !entries[1].Match(FoldZoneText("x/y/a.txt")) {
			t.Errorf("interior-dot dir entry: pattern=%q, want x/y/ matching x/y/a.txt", entries[1].Pattern)
		}
		// a manifest holds exactly one YAML document; anything after the first is rejected
		// (merge-gate round 1 P2-2, observed red first)
		_, err = ParseProtectedZone([]byte("version: 1\ncategories:\n  probe:\n    paths: [\"docs/\"]\n---\ngarbage: [\n"), false, "x.yaml")
		swept++
		if err == nil {
			t.Errorf("overlay with a second YAML document: valid, want invalid (err=%v)", err)
		}
		zoneSweptFloor(t, swept, 40)
	})
}
