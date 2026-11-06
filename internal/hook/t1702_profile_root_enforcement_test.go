package hook

// Card t1702, iter5 F2 [P1], end to end: the SHIPPED protected-zone manifest,
// parsed as the deny sites parse it, must make zoneUserRootCovered refuse a
// manifest-tracked file under a PROFILE root exactly as it does under the
// fixed roots. The earlier parity test built the profile forms and judged them
// tracked but never matched them against a shipped entry, so a profile root's
// tracked files stayed removable by Write, Edit and Bash.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func shippedZoneLoad(t *testing.T) config.ProtectedZoneLoad {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "template", "templates", ".moai", "config", "sections", "protected-zone.yaml"))
	if err != nil {
		t.Fatalf("read the shipped manifest: %v", err)
	}
	entries, err := config.ParseProtectedZone(data, true, "shipped protected-zone.yaml")
	if err != nil {
		t.Fatalf("parse the shipped manifest: %v", err)
	}
	return config.ProtectedZoneLoad{State: config.ZoneStateOK, Zone: config.ProtectedZone{Entries: entries}}
}

func TestZoneUserRootCovered_ProfileRootTrackedFilesAreRefusedLikeFixedRoots(t *testing.T) {
	home := t.TempDir()
	t1619SetHome(t, home)

	fixedSkill := filepath.Join(home, ".claude", "skills", "moai-a", "skill.md")
	profileSkill := filepath.Join(home, ".moai", "claude-profiles", "work", "skills", "moai-b", "skill.md")
	profileAgent := filepath.Join(home, ".moai", "claude-profiles", "work", "agents", "manager-x.md")
	untrackedInProfile := filepath.Join(home, ".moai", "claude-profiles", "work", "skills", "own-skill", "skill.md")
	for _, p := range []string{fixedSkill, profileSkill, profileAgent, untrackedInProfile} {
		t1619WriteFile(t, p, "x\n")
	}
	t1619WriteFile(t, filepath.Join(home, ".moai", "user-assets.json"),
		`{"files": {"claude-skills/moai-a/skill.md": {}, "claude-skills@work/moai-b/skill.md": {}, "claude-agents@work/manager-x.md": {}}}`)

	load := shippedZoneLoad(t)

	cases := []struct {
		name string
		path string
		want bool
	}{
		{"fixed root, tracked (control)", fixedSkill, true},
		{"profile skills root, tracked", profileSkill, true},
		{"profile agents root, tracked", profileAgent, true},
		{"profile skills root, the tracked file's directory", filepath.Dir(profileSkill), true},
		// The profile's own files that the manifest does not track stay writable:
		// the protection is for what the install put there, nothing else.
		{"profile root, a file the manifest does not track", untrackedInProfile, false},
	}
	for _, c := range cases {
		if _, display, covered := zoneUserRootCovered(c.path, load); covered != c.want {
			t.Errorf("%s: covered = %v (display %q), want %v", c.name, covered, display, c.want)
		}
	}
}

// The same profile file, once its profile directory is gone from the store, is
// no longer enumerated as a root and the judgment falls back to "nothing of it
// is on disk to protect" - the posture userRootDirRels documents.
func TestZoneUserRootCovered_AJunkProfileNameIsNotARoot(t *testing.T) {
	home := t.TempDir()
	t1619SetHome(t, home)
	junk := filepath.Join(home, ".moai", "claude-profiles", "has space", "skills", "moai-b", "skill.md")
	t1619WriteFile(t, junk, "x\n")
	t1619WriteFile(t, filepath.Join(home, ".moai", "user-assets.json"),
		`{"files": {"claude-skills@has space/moai-b/skill.md": {}}}`)
	if _, display, covered := zoneUserRootCovered(junk, shippedZoneLoad(t)); covered {
		t.Errorf("a directory with a junk profile name was treated as a protected root (display %q)", display)
	}
}
