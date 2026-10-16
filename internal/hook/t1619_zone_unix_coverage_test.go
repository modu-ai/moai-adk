//go:build unix

// t1619_zone_unix_coverage_test.go — card t1619 bundle A: the unix-only arms of
// the changed-line closure for the hook package's user-root protection helpers.
// A symlink that points at itself makes zoneResolve fail closed at its depth
// bound, so every "cannot resolve" arm is driven by a real link on disk. The
// test also covers the $HOME spellings and the non-regular read. No production
// code is changed.
package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// t1619SelfLoop creates path as a symlink to itself. Resolving it fails at the
// symlink depth bound, which is the fail-closed resolution the helpers guard.
func t1619SelfLoop(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.Symlink(path, path); err != nil {
		t.Fatalf("symlink %s -> itself: %v", path, err)
	}
}

// TestUserRootZoneFormsExpandsHomeSpellings covers the $HOME and ${HOME}
// spellings of a candidate. Each spelling resolves to the same tracked file
// under the claude-skills root, and the folded user-root form is emitted.
func TestUserRootZoneFormsExpandsHomeSpellings(t *testing.T) {
	home := t.TempDir()
	t1619SetHome(t, home)
	t1619WriteFile(t, filepath.Join(home, ".claude", "skills", "moai-x", "SKILL.md"), "x\n")
	const want = "user-root:claude-skills/moai-x/skill.md"
	for _, cand := range []string{
		"~/.claude/skills/moai-x/SKILL.md",
		"$HOME/.claude/skills/moai-x/SKILL.md",
		"${HOME}/.claude/skills/moai-x/SKILL.md",
	} {
		if !t1619HasDisplay(userRootZoneForms(cand), want) {
			t.Fatalf("userRootZoneForms(%q) lacks %s", cand, want)
		}
	}
}

// TestUserRootZoneFormsSkipsRootThatCannotResolve covers the root loop's
// unresolvable-root exit. The claude-skills root is a self-referencing link, so
// it is skipped, while the claude-agents root still yields its form.
func TestUserRootZoneFormsSkipsRootThatCannotResolve(t *testing.T) {
	home := t.TempDir()
	t1619SetHome(t, home)
	t1619SelfLoop(t, filepath.Join(home, ".claude", "skills"))
	agent := filepath.Join(home, ".claude", "agents", "manager-x.md")
	t1619WriteFile(t, agent, "x\n")

	forms := userRootZoneForms(agent)
	if !t1619HasDisplay(forms, "user-root:claude-agents/manager-x.md") {
		t.Fatalf("the resolvable claude-agents root lost its form: %+v", forms)
	}
	for _, f := range forms {
		if strings.HasPrefix(f.Display, "user-root:claude-skills") {
			t.Fatalf("a root that cannot resolve produced a form: %q", f.Display)
		}
	}
}

// TestUserRootZoneFormsSkipsCandidateThatCannotResolve covers the per-root exit
// for a candidate whose resolved spelling fails. The unresolved link-path alias
// is still judged, and no resolved form is emitted.
func TestUserRootZoneFormsSkipsCandidateThatCannotResolve(t *testing.T) {
	home := t.TempDir()
	t1619SetHome(t, home)
	skills := filepath.Join(home, ".claude", "skills")
	loop := filepath.Join(skills, "loopdir")
	t1619SelfLoop(t, loop)

	forms := userRootZoneForms(filepath.Join(loop, "SKILL.md"))
	if !t1619HasDisplay(forms, "user-root:claude-skills/loopdir/skill.md") {
		t.Fatalf("the link-path alias of an unresolvable candidate was not judged: %+v", forms)
	}
}

// TestUserRootTracksAnyFailsClosedOnUnresolvableKeys covers the tracked-ness
// judgment's real-path pass. An unresolvable candidate fails closed. A manifest
// key that cannot resolve, or whose shape is malformed, is passed over, and an
// untracked candidate with a real location is reported as untracked.
func TestUserRootTracksAnyFailsClosedOnUnresolvableKeys(t *testing.T) {
	home := t.TempDir()
	skills := filepath.Join(home, ".claude", "skills")
	loop := filepath.Join(skills, "loopdir")
	t1619SelfLoop(t, loop)
	t1619WriteFile(t, filepath.Join(skills, "real.md"), "real\n")
	t1619WriteFile(t, filepath.Join(skills, "untracked.md"), "untracked\n")
	t1619WriteFile(t, filepath.Join(home, ".moai", "user-assets.json"),
		`{"files": {"claude-skills/loopdir/skill.md": {}, "claude-skills/real.md": {}, "bogus-slug/x.md": {}, "bare-key": {}}}`)

	t.Run("candidate under an unresolvable link fails closed", func(t *testing.T) {
		if userRootTracksAny(home, "user-root:claude-skills/loopdir/other.md") {
			t.Fatal("a candidate whose real location cannot resolve read as tracked")
		}
	})
	t.Run("malformed and unresolvable manifest keys are passed over", func(t *testing.T) {
		if userRootTracksAny(home, "user-root:claude-skills/untracked.md") {
			t.Fatal("an untracked file read as tracked past malformed or unresolvable keys")
		}
	})
}

// TestUserManifestProtectFormsFailsClosedOnUnresolvableTargets covers the
// manifest-protection exits for a target that cannot resolve. A candidate that
// is itself a self-referencing link, and a home that cannot resolve for a
// candidate outside it, both yield no protection form.
func TestUserManifestProtectFormsFailsClosedOnUnresolvableTargets(t *testing.T) {
	t.Run("candidate is a self-referencing link", func(t *testing.T) {
		home := t.TempDir()
		t1619SetHome(t, home)
		t1619SelfLoop(t, filepath.Join(home, "loopfile"))
		if forms := userManifestProtectForms("~/loopfile"); forms != nil {
			t.Fatalf("userManifestProtectForms of an unresolvable target = %+v, want nil", forms)
		}
	})
	t.Run("home cannot resolve and the candidate is outside it", func(t *testing.T) {
		loopHome := filepath.Join(t.TempDir(), "loophome")
		t1619SelfLoop(t, loopHome)
		t1619SetHome(t, loopHome)
		outside := filepath.Join(t.TempDir(), "outside", "file.txt")
		if forms := userManifestProtectForms(outside); forms != nil {
			t.Fatalf("userManifestProtectForms outside an unresolvable home = %+v, want nil", forms)
		}
	})
}

// TestUserManifestProtectFormsResolvesManifestDirThroughAlias covers the
// resolved-path arm. A link outside the home that points at the home's .moai
// directory reaches the manifest through its real location, and the manifest
// protection form is emitted.
func TestUserManifestProtectFormsResolvesManifestDirThroughAlias(t *testing.T) {
	home := t.TempDir()
	t1619SetHome(t, home)
	t1619WriteFile(t, filepath.Join(home, ".moai", "user-assets.json"), "{}\n")
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(filepath.Join(home, ".moai"), alias); err != nil {
		t.Fatalf("symlink alias -> .moai: %v", err)
	}

	forms := userManifestProtectForms(filepath.Join(alias, "user-assets.json"))
	if len(forms) != 1 || forms[0].Display != ".moai/user-assets.json" {
		t.Fatalf("userManifestProtectForms through an alias = %+v, want the manifest form", forms)
	}
}

// TestReadRegularFileRefusesNonRegularPath covers the non-regular exit of the
// zone's manifest read. A directory opens but is not a regular file, so it
// reads as no evidence. A regular file reads through the same path.
func TestReadRegularFileRefusesNonRegularPath(t *testing.T) {
	dir := t.TempDir()
	if data, ok := readRegularFile(dir); ok || data != nil {
		t.Fatalf("readRegularFile(directory) = (%q, %v), want no evidence", data, ok)
	}
	file := filepath.Join(dir, "f.json")
	t1619WriteFile(t, file, "payload")
	if data, ok := readRegularFile(file); !ok || string(data) != "payload" {
		t.Fatalf("readRegularFile(regular file) = (%q, %v), want the payload", data, ok)
	}
}
