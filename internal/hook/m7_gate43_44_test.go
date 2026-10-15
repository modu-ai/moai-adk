package hook

// m7_gate43_44_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 gate rounds 43/44:
//
//   - 43-1 [P1]: the manifest protection must arm BOTH tool paths — a
//     Write/Edit replacing ~/.moai/user-assets.json with {"files":{}} used
//     to sail through (only the Bash path judged the manifest forms), and
//     an emptied manifest unregisters every managed asset, after which the
//     installer overwrites them freely.
//   - 43-2 [P2]: the protection is the manifest ITSELF and its containing
//     directory — the former ~/.moai/** prefix arm over-protected (an
//     untracked user note under ~/.moai was denied a deletion the normal
//     tracked-ness judgment should decide).
//   - 44-1 [P1]: the symlink dual forms are PAIRED BY BASIS — unresolved
//     candidate against the unresolved root, resolved candidate against
//     the resolved root. Mixing loses the tracked key when the ROOT itself
//     is a symlink: the link-spelled candidate never prefix-matched the
//     resolved root, so a manifest key recorded under the link path
//     (demo/SKILL.md) read as unregistered real-demo/SKILL.md.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func symlinkSkipWindows(t *testing.T, oldname, newname string) {
	t.Helper()
	if err := os.Symlink(oldname, newname); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink creation unavailable on %s: %v", runtime.GOOS, err)
		}
		t.Fatalf("symlink %s -> %s: %v", newname, oldname, err)
	}
}

// gate43Fixture builds the guarded project + user home the manifest
// protection arms judge: a project overlay declaring the managed roots
// (the zone loads OK) and a user manifest tracking one skill file.
func gate43Fixture(t *testing.T) (root, home string) {
	t.Helper()
	root = t.TempDir()
	home = t.TempDir()
	overlayDir := filepath.Join(root, ".moai", "project")
	if err := os.MkdirAll(overlayDir, 0o755); err != nil {
		t.Fatal(err)
	}
	overlay := `version: 1
categories:
  managed_assets:
    runtime_paths:
      - user-root:claude-skills/
      - user-root:claude-agents/moai/
`
	if err := os.WriteFile(filepath.Join(overlayDir, "protected-zone.yaml"), []byte(overlay), 0o644); err != nil {
		t.Fatal(err)
	}
	// A managed user-side twin the user-root arm can judge (the control).
	skill := filepath.Join(home, ".claude", "skills", "moai-alpha", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill, []byte("managed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version": 1, "files": {"claude-skills/moai-alpha/SKILL.md": {"sha256": "ab"}}}`
	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	prevHome := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = prevHome })
	prevGetwd := zoneGetwd
	zoneGetwd = func() (string, error) { return root, nil }
	t.Cleanup(func() { zoneGetwd = prevGetwd })
	return root, home
}

// writeDecision drives one identity Write call through the hook entry.
func writeDecision(t *testing.T, root, rawPath string) (string, string) {
	t.Helper()
	h := &preToolHandler{policy: DefaultSecurityPolicy(), projectDir: root}
	raw, err := json.Marshal(map[string]string{"file_path": rawPath})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return h.checkHarnessFrozenZoneFromInput(harnessLearnerIdentity, "Write", raw)
}

func TestWriteToolCannotReplaceUserManifest(t *testing.T) {
	root, _ := gate43Fixture(t)

	// 43-1: the emptied-manifest replacement is denied on the Write path —
	// no tracked-ness filter (the manifest IS the evidence).
	sentinel, reason := writeDecision(t, root, "~/.moai/user-assets.json")
	if sentinel == "" {
		t.Fatal("a Write replacing the user manifest was ALLOWED — the manifest protection covers only the Bash path (gate 43-1)")
	}
	if reason != "" && !json.Valid([]byte(reason)) {
		// the reason is free text; only its presence is asserted above
		_ = reason
	}

	// The absolute and $HOME spellings earn the same denial.
	for _, p := range []string{"$HOME/.moai/user-assets.json"} {
		if s, _ := writeDecision(t, root, p); s == "" {
			t.Fatalf("the %q spelling of the manifest replacement was allowed", p)
		}
	}

	// 43-2 control: an UNTRACKED sibling under ~/.moai stays writable — the
	// narrowed protection covers the manifest and its directory, nothing
	// wider.
	if s, _ := writeDecision(t, root, "~/.moai/user-note.txt"); s != "" {
		t.Fatalf("an untracked user note under ~/.moai was denied: %s (the over-protection the narrowing removes)", s)
	}

	// The tracked managed asset is still denied (the user-root arm intact).
	if s, _ := writeDecision(t, root, "~/.claude/skills/moai-alpha/SKILL.md"); s == "" {
		t.Fatal("a tracked managed asset was allowed — the user-root arm broke alongside the manifest changes")
	}
}

func TestUserManifestProtectFormsArePrecise(t *testing.T) {
	home := t.TempDir()
	prevHome := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = prevHome })

	for _, p := range []string{
		"~/.moai/user-assets.json",
		"$HOME/.moai/user-assets.json",
		"${HOME}/.moai/user-assets.json",
		filepath.Join(home, ".moai", "user-assets.json"),
		"~/.moai",  // the containing directory itself
		"~/.moai/", // trailing-slash spelling of the same deletion
		// Gate round 46-1: the case-alias names the same inode on a
		// case-insensitive filesystem — the resolver returns the caller's
		// spelling, so the FOLD is the only normalization that holds.
		"~/.MOAI/user-assets.json",
		"~/.MOAI",
	} {
		if forms := userManifestProtectForms(p); len(forms) == 0 {
			t.Fatalf("candidate %q produced no protected form — the manifest or its directory escaped protection", p)
		}
	}
	for _, p := range []string{
		"~/.moai/user-note.txt", // an untracked sibling — the normal judgment
		"~/.moai/subdir/x",      // a deeper path — the normal judgment
		"~/.moai/subdir/",       // a deeper directory — the normal judgment
		"~/.claude/skills/x.md", // outside .moai entirely
	} {
		if forms := userManifestProtectForms(p); len(forms) != 0 {
			t.Fatalf("candidate %q was over-protected: %v (gate 43-2 — only the manifest and its containing directory)", p, forms)
		}
	}
}

func TestDualFormsPairedByBasis(t *testing.T) {
	home := t.TempDir()
	// The ROOT is a symlink: ~/.claude -> real-claude (the dotfile-manager
	// shape), and INSIDE it a leaf symlink: skills/demo -> skills/real-demo
	// (the installed-symlink shape). The manifest tracks the LINK path.
	realClaude := filepath.Join(home, "real-claude")
	if err := os.MkdirAll(filepath.Join(realClaude, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkSkipWindows(t, realClaude, filepath.Join(home, ".claude"))
	realDemo := filepath.Join(realClaude, "skills", "real-demo")
	if err := os.MkdirAll(realDemo, 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkSkipWindows(t, realDemo, filepath.Join(realClaude, "skills", "demo"))
	if err := os.WriteFile(filepath.Join(realDemo, "SKILL.md"), []byte("managed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The manifest records the LINK path — the key the dual-form pairing
	// must keep reachable.
	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version": 1, "files": {"claude-skills/demo/SKILL.md": {"sha256": "ab"}}}`
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	prevHome := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = prevHome })

	// The candidate is spelled through BOTH links — the shape a tracked
	// write through the installed path arrives with. Form displays are
	// case-folded (the comparison basis the entry match uses), so the
	// tracked key is judged on the folded form.
	forms := userRootZoneForms(filepath.Join(home, ".claude", "skills", "demo", "SKILL.md"))
	want := config.FoldZoneText("user-root:claude-skills/demo/SKILL.md")
	found := false
	for _, f := range forms {
		if f.Folded == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("the link-path form %q is missing from %v — the unresolved candidate was judged against the RESOLVED root and lost the tracked key (gate 44-1)", want, forms)
	}
	if !userRootTracksAny(home, want) {
		t.Fatal("the link-path form read untracked — the tracked key is lost")
	}

	// Gate round 50-1: the REVERSE direction — the candidate arrives as
	// the REAL path while the manifest tracks the LINK path. The tracked
	// key must resolve to its real location for the comparison, or a
	// write through the real path sails through as unregistered.
	realForms := userRootZoneForms(filepath.Join(realDemo, "SKILL.md"))
	realWant := config.FoldZoneText("user-root:claude-skills/real-demo/SKILL.md")
	foundReal := false
	for _, f := range realForms {
		if f.Folded == realWant {
			foundReal = true
		}
	}
	if !foundReal {
		t.Fatalf("the real-path form %q is missing from %v (gate 50-1 reverse mapping)", realWant, realForms)
	}
	if !userRootTracksAny(home, realWant) {
		t.Fatal("the REAL-path form of a link-tracked file read untracked — the reverse dual-form mapping is missing (gate 50-1)")
	}

	// Gate round 51-1: the containment check applies on the REAL path too —
	// rm -rf real-demo (the directory CONTAINING the tracked file via its
	// link) is as destructive as a direct file delete, and equality-only
	// reverse mapping let it through.
	if !userRootTracksAny(home, "user-root:claude-skills/real-demo") {
		t.Fatal("the real-path DIRECTORY containing a tracked file read untracked — the real-path containment is missing (gate 51-1)")
	}
}

// TestSymlinkedMoaiManifestStaysProtected — gate round 46-2: ~/.moai as a
// SYMLINK to a directory outside the home relocates the manifest's REAL
// path beyond the home prefix; the link-path spelling must keep the
// manifest judged.
func TestReviewManifestResolvedAliases(t *testing.T) {
	home := t.TempDir()
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outside, "moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "moai", "user-assets.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	symlinkSkipWindows(t, filepath.Join(outside, "moai"), filepath.Join(home, ".moai"))

	prevHome := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = prevHome })

	for _, p := range []string{
		"~/.moai/user-assets.json",
		filepath.Join(home, ".moai", "user-assets.json"),
		"~/.moai",
		// Gate round 48-1: the symlink+dot combination escapes the naive
		// link-path rest — the dot drops textually and the rest matches.
		// RAW concatenation — Join would clean the dot away (the round-34
		// lesson).
		"~/.moai/./user-assets.json",
		home + "/.moai/./user-assets.json",
		"~/.moai/.",
	} {
		if forms := userManifestProtectForms(p); len(forms) == 0 {
			t.Fatalf("candidate %q escaped protection — the symlinked ~/.moai lost its manifest (gate 46-2 / 48-1)", p)
		}
	}
	// Gate round 50-2: the manifest's REAL-path ancestors are protected —
	// rm -rf outside deletes the symlinked manifest just as surely as the
	// file itself. An unrelated directory stays outside the protection.
	if forms := userManifestProtectForms(outside); len(forms) == 0 {
		t.Fatal("rm -rf on the manifest's real-path ANCESTOR escaped protection (gate 50-2)")
	}
	unrelated := t.TempDir()
	if forms := userManifestProtectForms(unrelated); len(forms) != 0 {
		t.Fatalf("an unrelated directory was over-protected: %v (gate 50-2)", forms)
	}
}

// TestUserRootPhysicalDotDot — gate round 48-3, RECONCILED with gate
// round 50-1: a `..` segment after a symlink keeps real-path semantics —
// the verdict follows the FILESYSTEM the spelling lands on, never the
// textually cleaned key. Under 50-1's reverse mapping the tracked key
// resolves to its real location, so the link+.. spelling that lands on the
// tracked asset's real file (moai-alpha/../moi-alpha/SKILL.md through the
// link) is DENIED — it modifies the managed asset — and the deny reason
// names the RESOLVED real-path form, not a textual collapse. A `..`
// spelling landing on a genuinely untracked file stays ALLOWED.
func TestUserRootPhysicalDotDot(t *testing.T) {
	root, home := gate43Fixture(t)

	// Re-shape the tracked skill as a symlink: skills/moai-alpha ->
	// skills/nested/moai-alpha (the manifest key keeps the LINK path), and
	// add an UNTRACKED twin for the allowed arm.
	skillsDir := filepath.Join(home, ".claude", "skills")
	nested := filepath.Join(skillsDir, "nested", "moai-alpha")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(skillsDir, "moai-alpha", "SKILL.md"), filepath.Join(nested, "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(skillsDir, "moai-alpha")); err != nil {
		t.Fatal(err)
	}
	symlinkSkipWindows(t, nested, filepath.Join(skillsDir, "moai-alpha"))
	untrackedDir := filepath.Join(skillsDir, "untracked-twin")
	if err := os.MkdirAll(untrackedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(untrackedDir, "SKILL.md"), []byte("the user's own\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The direct link-path spelling stays denied (the tracked key).
	if s, _ := writeDecision(t, root, filepath.Join(skillsDir, "moai-alpha", "SKILL.md")); s == "" {
		t.Fatal("the link-path spelling of a tracked file was allowed")
	}

	// The link + `..` spelling lands on the tracked asset's REAL file (the
	// link resolves before `..` applies — EvalSymlinks measured) — denied,
	// and the deny reason names the resolved form. RAW concatenation —
	// Join would clean the `..` away before the hook sees it (the
	// round-34 lesson).
	spelled := skillsDir + "/moai-alpha/../moai-alpha/SKILL.md"
	if s, reason := writeDecision(t, root, spelled); s == "" {
		t.Fatal("a link+.. spelling landing on the tracked asset's real file was allowed (gate 50-1 reverse mapping)")
	} else if !strings.Contains(reason, "nested/moai-alpha") {
		t.Fatalf("the deny did not name the RESOLVED real-path form — a textual collapse decided it: %s (gate 48-3)", reason)
	}

	// A `..` spelling landing on a genuinely UNTRACKED file stays allowed:
	// moai-alpha (link) / .. -> skills, then untracked-twin/SKILL.md.
	healthy := skillsDir + "/moai-alpha/../untracked-twin/SKILL.md"
	if s, reason := writeDecision(t, root, healthy); s != "" {
		t.Fatalf("a link+.. spelling landing on an untracked file was DENIED: %s / reason: %s (gate 48-3 — real-path semantics)", s, reason)
	}
	if _, statErr := os.Stat(filepath.Join(nested, "SKILL.md")); statErr != nil {
		t.Fatalf("the real file vanished: %v", statErr)
	}
}

// TestDotSegmentAliasInterpretsTrackedKey — gate round 47-2: a dot-segment
// spelling of a tracked link path is normalized TEXTUALLY (path.Clean, no
// symlink walk) so the alias interprets the same tracked key — the former
// drop left skills/./moai-alpha/SKILL.md unjudged.
func TestDotSegmentAliasInterpretsTrackedKey(t *testing.T) {
	root, home := gate43Fixture(t)

	want := config.FoldZoneText("user-root:claude-skills/moai-alpha/SKILL.md")
	forms := userRootZoneForms(home + "/.claude/skills/./moai-alpha/SKILL.md")
	found := false
	for _, f := range forms {
		if f.Folded == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("the dot-segment spelling lost the tracked key: forms = %v (gate 47-2)", forms)
	}
	if !userRootTracksAny(home, want) {
		t.Fatal("the normalized alias read untracked")
	}
	if s, _ := writeDecision(t, root, home+"/.claude/skills/./moai-alpha/SKILL.md"); s == "" {
		t.Fatal("a Write through a dot-segment spelling of a tracked path was ALLOWED (gate 47-2)")
	}
}
