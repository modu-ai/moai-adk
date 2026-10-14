// profile_reach_test.go — card t1702 (gaps 1-2): install reach into Claude
// profile config roots.
//
// A profile-launched session (`moai cc -p <name>`) runs with
// CLAUDE_CONFIG_DIR=~/.moai/claude-profiles/<name>, so the only skill/agent
// roots that session can see are the ones INSIDE its own profile dir. These
// tests pin that a standard install reaches them, that a profile's own
// differing copy is never overwritten, and that junk directories under the
// profile base are ignored. Every test runs against a temp HOME (plan §D).
package userassets

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/modu-ai/moai-adk/internal/template"
)

// profileFixture builds a minimal core-only catalog (skill "moai" + agent
// "manager-x") over a MapFS source — the dispatcher-skill shape the
// profile-reach defect is about ("Unknown skill: moai").
type profileFixture struct {
	home    string
	cat     *template.Catalog
	src     fstest.MapFS
	skillV1 []byte
	agentV1 []byte
	tomlV1  []byte
}

func newProfileFixture(t *testing.T) *profileFixture {
	t.Helper()
	f := &profileFixture{
		home:    t.TempDir(),
		skillV1: []byte("---\nname: moai\ndescription: dispatcher\n---\nskill body v1\n"),
		agentV1: []byte("---\nname: manager-x\n---\nagent body v1\n"),
		tomlV1:  []byte("name = \"manager-x\"\n"),
	}
	f.cat = &template.Catalog{
		Version: "test",
		Catalog: template.CatalogSections{
			Core: template.TierSection{
				Skills: []template.Entry{
					{Name: "moai", Tier: template.TierCore, Path: "templates/.claude/skills/moai/", Version: "1.0.0"},
				},
				Agents: []template.Entry{
					{Name: "manager-x", Tier: template.TierCore, Path: "templates/.claude/agents/moai/manager-x.md", Version: "1.0.0"},
				},
			},
		},
	}
	f.src = fstest.MapFS{
		".claude/skills/moai/SKILL.md":      &fstest.MapFile{Data: f.skillV1},
		".agents/skills/moai/SKILL.md":      &fstest.MapFile{Data: f.skillV1},
		".claude/agents/moai/manager-x.md":  &fstest.MapFile{Data: f.agentV1},
		".codex/agents/moai/manager-x.toml": &fstest.MapFile{Data: f.tomlV1},
	}
	return f
}

func (f *profileFixture) installer(t *testing.T) *Installer {
	t.Helper()
	return &Installer{
		Home:        f.home,
		Catalog:     f.cat,
		Source:      f.src,
		MoaiVersion: "v3.2.0-test",
		Now:         func() time.Time { return time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC) },
	}
}

// profileBase is the temp home's profile store.
func (f *profileFixture) profileBase() string {
	return filepath.Join(f.home, ".moai", "claude-profiles")
}

// makeProfileDir creates one real profile directory under the base.
func (f *profileFixture) makeProfileDir(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join(f.profileBase(), name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create profile dir %s: %v", name, err)
	}
	return dir
}

// TestProfileSessionSkillReach is the card t1702 RED test: after a standard
// install, a session whose CLAUDE_CONFIG_DIR points at a profile dir must be
// able to resolve the core skill moai and the core agent from the profile's
// own skills/agents directories.
func TestProfileSessionSkillReach(t *testing.T) {
	t.Parallel()
	f := newProfileFixture(t)
	profDir := f.makeProfileDir(t, "work")

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if res.Installed == 0 {
		t.Fatal("first install installed nothing")
	}

	skillBytes, err := os.ReadFile(filepath.Join(profDir, "skills", "moai", "SKILL.md"))
	if err != nil {
		t.Fatalf("profile session cannot resolve skill moai after install: %v", err)
	}
	if string(skillBytes) != string(f.skillV1) {
		t.Errorf("profile skill bytes differ: %q", skillBytes)
	}
	agentBytes, err := os.ReadFile(filepath.Join(profDir, "agents", "manager-x.md"))
	if err != nil {
		t.Fatalf("profile session cannot resolve agent manager-x after install: %v", err)
	}
	if string(agentBytes) != string(f.agentV1) {
		t.Errorf("profile agent bytes differ: %q", agentBytes)
	}

	// The manifest tracks the profile-root copies under their own slugs.
	m, err := Load(ManifestPath(f.home))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if _, ok := m.Files["claude-skills@work/moai/SKILL.md"]; !ok {
		t.Error("manifest missing the claude-skills@work record for moai/SKILL.md")
	}
	if _, ok := m.Files["claude-agents@work/manager-x.md"]; !ok {
		t.Error("manifest missing the claude-agents@work record for manager-x.md")
	}
}

// TestProfileJunkDirsIgnored pins the junk-dir filter: names containing
// spaces, the reserved sentinel, hidden entries, and non-directories are
// never enumerated as profile roots, and an install never creates anything
// inside them (card t1702 gap 1).
func TestProfileJunkDirsIgnored(t *testing.T) {
	t.Parallel()
	f := newProfileFixture(t)
	base := f.profileBase()
	for _, junk := range []string{"My Profile", "__no_such_profile__", ".hidden", "trailing.space "} {
		if err := os.MkdirAll(filepath.Join(base, junk), 0o755); err != nil {
			t.Fatalf("create junk dir %q: %v", junk, err)
		}
	}
	// A plain file (not a directory) under the base is never a profile.
	if err := os.WriteFile(filepath.Join(base, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write junk file: %v", err)
	}

	f.makeProfileDir(t, "work")
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("Install: %v", err)
	}

	for _, r := range ResolveRoots(f.home) {
		slug := string(r.Slug)
		if i := indexByte(slug, '@'); i >= 0 && !ValidProfileDirName(slug[i+1:]) {
			t.Errorf("root %q derived from an invalid profile name", r.Slug)
		}
	}
	for _, junk := range []string{"My Profile", "__no_such_profile__", ".hidden", "trailing.space "} {
		for _, sub := range []string{"skills", "agents"} {
			abs := filepath.Join(base, junk, sub)
			if _, err := os.Stat(abs); !os.IsNotExist(err) {
				t.Errorf("junk dir %q gained an install root at %s (stat err = %v)", junk, abs, err)
			}
		}
	}
	if got := RealProfileDirNames(f.home); len(got) != 1 || got[0] != "work" {
		t.Errorf("RealProfileDirNames = %v, want [work]", got)
	}
}

// indexByte is a tiny helper so the junk test can split a slug without
// importing strings for one call.
func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// TestProfileOwnVersionPreserved pins the collision rule: a profile dir that
// already holds ITS OWN differing copy of a shipped skill keeps that copy,
// and the run reports the collision (the same policy class the existing
// install applies to user customizations, REQ-010) (card t1702 gap 1).
func TestProfileOwnVersionPreserved(t *testing.T) {
	t.Parallel()
	f := newProfileFixture(t)
	profDir := f.makeProfileDir(t, "work")
	ownSkill := filepath.Join(profDir, "skills", "moai", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(ownSkill), 0o755); err != nil {
		t.Fatalf("mkdir profile skill dir: %v", err)
	}
	ownBytes := []byte("---\nname: moai\ndescription: my own tuning\n---\nuser's own body\n")
	if err := os.WriteFile(ownSkill, ownBytes, 0o644); err != nil {
		t.Fatalf("write profile's own skill: %v", err)
	}

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if got, err := os.ReadFile(ownSkill); err != nil || string(got) != string(ownBytes) {
		t.Errorf("profile's own skill was overwritten (err = %v, body = %q)", err, got)
	}
	found := false
	for _, c := range res.Collisions {
		if c == "claude-skills@work/moai/SKILL.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("collision on the profile's own skill not reported; collisions = %v", res.Collisions)
	}
}

// TestProfileTrackedDivergencePreserved pins the REQ-023 arm on a
// profile-root key: a tracked profile copy edited AFTER install is preserved
// in place, the shipped replacement is backed up under the profile slug, and
// the divergence is reported (card t1702 gap 1).
func TestProfileTrackedDivergencePreserved(t *testing.T) {
	t.Parallel()
	f := newProfileFixture(t)
	profDir := f.makeProfileDir(t, "work")
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("Install: %v", err)
	}
	ownSkill := filepath.Join(profDir, "skills", "moai", "SKILL.md")
	edited := []byte("---\nname: moai\n---\nuser edited after install\n")
	if err := os.WriteFile(ownSkill, edited, 0o644); err != nil {
		t.Fatalf("edit profile skill: %v", err)
	}

	// The refresh pass (what `moai update`'s user-asset phase runs first)
	// classifies the edited copy through the truth table.
	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if got, err := os.ReadFile(ownSkill); err != nil || string(got) != string(edited) {
		t.Errorf("edited profile copy was not preserved (err = %v, body = %q)", err, got)
	}
	found := false
	for _, d := range res.Divergences {
		if d == "claude-skills@work/moai/SKILL.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("divergence on the edited profile copy not reported; divergences = %v", res.Divergences)
	}
	if _, err := os.Stat(filepath.Join(BackupHome(f.home), "claude-skills@work", "moai", "SKILL.md")); err != nil {
		t.Errorf("shipped replacement not backed up under the profile slug: %v", err)
	}
}

// TestClaudeConfigRoots pins the config-root enumeration shared by gap 2 and
// doctor: ~/.claude first, then the real profile dirs sorted by name; junk
// is never returned.
func TestClaudeConfigRoots(t *testing.T) {
	t.Parallel()
	f := newProfileFixture(t)
	f.makeProfileDir(t, "zeta")
	f.makeProfileDir(t, "alpha")
	if err := os.MkdirAll(filepath.Join(f.profileBase(), "Bad Name"), 0o755); err != nil {
		t.Fatalf("create junk dir: %v", err)
	}

	got := ClaudeConfigRoots(f.home)
	want := []string{
		filepath.Join(f.home, ".claude"),
		filepath.Join(f.profileBase(), "alpha"),
		filepath.Join(f.profileBase(), "zeta"),
	}
	if len(got) != len(want) {
		t.Fatalf("ClaudeConfigRoots = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ClaudeConfigRoots[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestValidProfileDirNameAndRootKind unit-pins the two helpers the slug
// scheme rests on.
func TestValidProfileDirNameAndRootKind(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]bool{
		"work": true, "Work-1": true, "a_b": true,
		"": false, ".": false, "..": false, ".hidden": false,
		"my profile": false, "__no_such_profile__": false,
		"sub/dir": false, `sub\dir`: false, "/abs": false,
	} {
		if got := ValidProfileDirName(name); got != want {
			t.Errorf("ValidProfileDirName(%q) = %v, want %v", name, got, want)
		}
	}
	if got := RootKind("claude-skills@work"); got != RootClaudeSkills {
		t.Errorf("RootKind(claude-skills@work) = %q, want claude-skills", got)
	}
	if got := RootKind(RootCodexAgents); got != RootCodexAgents {
		t.Errorf("RootKind(codex-agents) = %q, want codex-agents", got)
	}
}
