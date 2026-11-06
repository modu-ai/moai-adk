// update_profile_reach_test.go — card t1702 (gaps 1-2) at the cli surface:
// ensureGlobalSettingsEnv reaching profile config roots, the update phase's
// profile-root reach + roots-reached summary line, and the Skill Reach
// doctor check. Home-isolated through the package seams; nothing here writes
// the developer's real HOME.
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/template"
)

// TestEnsureGlobalSettingsEnv_ProfileConfigRootsReach pins gap 2: the same
// moai-managed settings cleanup applied to ~/.claude reaches every real
// profile dir too, and never touches junk dirs (card t1702). Reassigns a
// package-level seam, so it MUST NOT call t.Parallel().
func TestEnsureGlobalSettingsEnv_ProfileConfigRootsReach(t *testing.T) {
	tmp := t.TempDir()

	origFn := userHomeDirFn
	userHomeDirFn = func() (string, error) { return tmp, nil }
	t.Cleanup(func() { userHomeDirFn = origFn })

	profDir := filepath.Join(tmp, ".moai", "claude-profiles", "work")
	if err := os.MkdirAll(profDir, 0o755); err != nil {
		t.Fatalf("create profile dir: %v", err)
	}
	profSettings := filepath.Join(profDir, "settings.json")
	if err := os.WriteFile(profSettings, []byte(`{
  "env": {"PATH": "stale", "USER_KEY": "keep-me"},
  "teammateMode": "auto",
  "outputStyle": "moai"
}`), 0o644); err != nil {
		t.Fatalf("write profile settings: %v", err)
	}
	// The profile's own hooks dir + a user-owned sibling that must survive.
	profHooks := filepath.Join(profDir, "hooks", "moai")
	if err := os.MkdirAll(profHooks, 0o755); err != nil {
		t.Fatalf("create profile hooks dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(profHooks, "handle.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write profile hook: %v", err)
	}
	siblingHook := filepath.Join(profDir, "hooks", "user-hook.sh")
	if err := os.WriteFile(siblingHook, []byte("#!/bin/sh\n# user-owned\n"), 0o755); err != nil {
		t.Fatalf("write sibling hook: %v", err)
	}
	// A junk dir must not be touched at all.
	junkSettings := filepath.Join(tmp, ".moai", "claude-profiles", "Junk Dir", "settings.json")
	if err := os.MkdirAll(filepath.Dir(junkSettings), 0o755); err != nil {
		t.Fatalf("create junk dir: %v", err)
	}
	if err := os.WriteFile(junkSettings, []byte(`{"env":{"PATH":"junk"}}`), 0o644); err != nil {
		t.Fatalf("write junk settings: %v", err)
	}

	if err := ensureGlobalSettingsEnv(); err != nil {
		t.Fatalf("ensureGlobalSettingsEnv: %v", err)
	}

	if _, err := os.Stat(profHooks); !os.IsNotExist(err) {
		t.Errorf("profile hooks/moai should be removed, stat err = %v", err)
	}
	if _, err := os.Stat(siblingHook); err != nil {
		t.Errorf("user-owned sibling hook %s must survive: %v", siblingHook, err)
	}
	data, err := os.ReadFile(profSettings)
	if err != nil {
		t.Fatalf("read profile settings: %v", err)
	}
	var cleaned map[string]any
	if err := json.Unmarshal(data, &cleaned); err != nil {
		t.Fatalf("parse cleaned profile settings: %v", err)
	}
	envMap, _ := cleaned["env"].(map[string]any)
	if envMap == nil {
		t.Fatalf("env section lost: %s", data)
	}
	if got := envMap["PATH"]; got == nil || got == "stale" {
		t.Errorf("moai-managed PATH key not refreshed in the profile settings (got %v)", got)
	}
	if envMap["USER_KEY"] != "keep-me" {
		t.Error("user-added env key did not survive in the profile settings")
	}
	if _, ok := cleaned["teammateMode"]; ok {
		t.Error("moai-managed teammateMode not removed from the profile settings")
	}
	if cleaned["outputStyle"] != "moai" {
		t.Error("unknown top-level key did not survive in the profile settings")
	}
	// Defaults the cleanup ensures must land in the profile root too.
	if envMap["CLAUDE_DISABLE_PATH_WARNING"] != "1" {
		t.Error("default CLAUDE_DISABLE_PATH_WARNING not written to the profile settings")
	}
	// Junk dir untouched.
	junkRaw, err := os.ReadFile(junkSettings)
	if err != nil || string(junkRaw) != `{"env":{"PATH":"junk"}}` {
		t.Errorf("junk dir settings were modified (err = %v, body = %s)", err, junkRaw)
	}
	// The default root has no settings file — none is created.
	if _, err := os.Stat(filepath.Join(tmp, ".claude", "settings.json")); !os.IsNotExist(err) {
		t.Errorf("the default root's settings.json should not be created, stat err = %v", err)
	}
}

// TestRunUserAssetUpdatePhase_ReachesProfileAndNamesRoots pins gap 1 at the
// phase level: a profile dir present before the update phase receives the
// core skills/agents, and the summary names the roots reached (card t1702).
func TestRunUserAssetUpdatePhase_ReachesProfileAndNamesRoots(t *testing.T) {
	home := t.TempDir()
	profDir := filepath.Join(home, ".moai", "claude-profiles", "work")
	if err := os.MkdirAll(profDir, 0o755); err != nil {
		t.Fatalf("create profile dir: %v", err)
	}

	var out bytes.Buffer
	if err := runUserAssetUpdatePhase(home, &out); err != nil {
		t.Fatalf("update phase: %v", err)
	}

	if _, err := os.Stat(filepath.Join(profDir, "skills", "moai", "SKILL.md")); err != nil {
		t.Fatalf("profile session cannot resolve skill moai after update: %v", err)
	}
	summary := out.String()
	if !strings.Contains(summary, "user-asset roots reached: 4 base roots + 2 profile roots (work)") {
		t.Errorf("summary missing the roots-reached line; got:\n%s", summary)
	}
}

// TestRunUserAssetUpdatePhase_RootsLineWithoutProfiles pins the count-aware
// no-profile form of the summary line.
func TestRunUserAssetUpdatePhase_RootsLineWithoutProfiles(t *testing.T) {
	home := t.TempDir()

	var out bytes.Buffer
	if err := runUserAssetUpdatePhase(home, &out); err != nil {
		t.Fatalf("update phase: %v", err)
	}
	if !strings.Contains(out.String(), "user-asset roots reached: 4 base roots (no Claude profile dirs)") {
		t.Errorf("summary missing the no-profile roots line; got:\n%s", out.String())
	}
}

// TestCheckProfileSkillReach pins the doctor check's observable behavior: a
// warn naming the active root while core skills are unresolvable, an OK once
// every core skill's SKILL.md exists under <active>/skills (card t1702). The
// message carries the caller's DISPLAY form — never the resolved absolute
// path (golden hermeticity).
func TestCheckProfileSkillReach(t *testing.T) {
	home := t.TempDir()
	active := filepath.Join(home, ".moai", "claude-profiles", "work")

	check := checkProfileSkillReach(active, "claude-profiles/work", home, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("with nothing installed: status = %v, want warn (message = %q)", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "NOT resolvable from claude-profiles/work") {
		t.Errorf("warn message should name the display root; got %q", check.Message)
	}
	if strings.Contains(check.Message, active) {
		t.Errorf("warn message must not embed the resolved absolute path; got %q", check.Message)
	}

	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load embedded catalog: %v", err)
	}
	for _, e := range cat.Catalog.Core.Skills {
		dir := filepath.Join(active, "skills", e.Name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("x"), 0o644); err != nil {
			t.Fatalf("write SKILL.md for %s: %v", e.Name, err)
		}
	}
	check = checkProfileSkillReach(active, "claude-profiles/work", home, false)
	if check.Status != uikit.CheckOK {
		t.Errorf("with every core skill installed: status = %v, want ok (message = %q)", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "core skill(s) resolvable from claude-profiles/work") {
		t.Errorf("ok message should name the display root; got %q", check.Message)
	}
}

// TestDoctorSkillReachUnreadableProfileStoreWarns pins the round-4 decision
// (card t1702): when the active root's installability cannot be judged
// because the profile store is unreadable, the Skill Reach check warns with
// the enumeration failure — it must not claim the root is "outside moai's
// install roots" on unobserved evidence. Skips on Windows and under
// effective root (0o000 unobservable).
func TestDoctorSkillReachUnreadableProfileStoreWarns(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not enforceable on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("effective root reads through 0o000 modes; the permission arm is unobservable")
	}
	home := t.TempDir()
	active := filepath.Join(home, ".moai", "claude-profiles", "work")
	if err := os.MkdirAll(active, 0o755); err != nil {
		t.Fatalf("create profile dir: %v", err)
	}
	store := filepath.Join(home, ".moai", "claude-profiles")
	if err := os.Chmod(store, 0o000); err != nil {
		t.Fatalf("chmod store: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(store, 0o755) })

	check := checkProfileSkillReach(active, "claude-profiles/work", home, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("status = %v, want warn (message = %q)", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "cannot confirm") || !strings.Contains(check.Message, "unreadable") {
		t.Errorf("warn should name the enumeration failure: %q", check.Message)
	}
	if strings.Contains(check.Message, "outside moai's install roots") {
		t.Errorf("an unjudgeable root must not be claimed outside the install roots: %q", check.Message)
	}
}

// TestUpdateGlobalSettingsEnv_UnreadableProfileStoreErrors pins the
// round-4 update-side decision (card t1702): the settings cleanup
// propagates a profile-store enumeration failure instead of silently
// cleaning ~/.claude only — the skipped profile-root cleanups stay visible.
// Reassigns a package-level seam, so it MUST NOT call t.Parallel(). Skips
// on Windows and under effective root (0o000 unobservable).
func TestUpdateGlobalSettingsEnv_UnreadableProfileStoreErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not enforceable on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("effective root reads through 0o000 modes; the permission arm is unobservable")
	}
	tmp := t.TempDir()
	origFn := userHomeDirFn
	userHomeDirFn = func() (string, error) { return tmp, nil }
	t.Cleanup(func() { userHomeDirFn = origFn })

	store := filepath.Join(tmp, ".moai", "claude-profiles")
	if err := os.MkdirAll(filepath.Join(store, "work"), 0o755); err != nil {
		t.Fatalf("create profile dir: %v", err)
	}
	if err := os.Chmod(store, 0o000); err != nil {
		t.Fatalf("chmod store: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(store, 0o755) })

	err := ensureGlobalSettingsEnv()
	if err == nil {
		t.Fatal("ensureGlobalSettingsEnv must propagate an unreadable profile store, not clean the default root only")
	}
	if !strings.Contains(err.Error(), store) {
		t.Errorf("error must name the store path %s; got %v", store, err)
	}
}

// TestDoctorSkillReachCustomConfigDirGuidance pins repair 3 (card t1702
// audit P2): when the ACTIVE CLAUDE_CONFIG_DIR is a user's own config dir
// — neither ~/.claude nor a MoAI profile dir — the warning must NOT advise
// 'moai update' (which cannot reach that root); it must name what WOULD
// fix it. Pre-fix this fails behaviorally: every warning advises
// 'moai update' regardless of reachability.
func TestDoctorSkillReachCustomConfigDirGuidance(t *testing.T) {
	// A temp home whose profile base holds no such dir: the active root is
	// the user's OWN config dir — outside moai's install families.
	check := checkProfileSkillReach("/home/someone/own-claude", "my-own-claude", t.TempDir(), false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("status = %v, want warn (message = %q)", check.Status, check.Message)
	}
	if strings.Contains(check.Message, "run 'moai update'") {
		t.Errorf("advice is not actionable for a config dir outside moai's install roots: %q", check.Message)
	}
	// home-guard:allow asserts the wording of the doctor advice, not a home access
	if !strings.Contains(check.Message, "claude-profiles") || !strings.Contains(check.Message, "~/.claude") {
		t.Errorf("guidance must name moai's managed root families: %q", check.Message)
	}
}

// TestUserAssetRootsLineConsumesInstallSnapshot pins the final-round repair
// (card t1702 audit round 2): the roots-reached line describes the INSTALL's
// validated snapshot, not a post-hoc disk re-enumeration — a profile dir
// created after the snapshot must never be reported as reached.
// Pre-fix the line re-enumerates at render time and claims "late"; the RED
// assertion below fails on that. Post-fix the line renders from the
// install result's root list and "late" never appears.
func TestUserAssetRootsLineConsumesInstallSnapshot(t *testing.T) {
	home := t.TempDir()
	base := filepath.Join(home, ".moai", "claude-profiles")
	if err := os.MkdirAll(filepath.Join(base, "early"), 0o755); err != nil {
		t.Fatalf("create early profile: %v", err)
	}

	inst, err := newUserAssetInstaller(home)
	if err != nil {
		t.Fatalf("installer: %v", err)
	}
	res, err := inst.Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	// The profile appears after the snapshot, before the summary render.
	if err := os.MkdirAll(filepath.Join(base, "late"), 0o755); err != nil {
		t.Fatalf("create late profile: %v", err)
	}

	// Post-fix the line renders from the install result's snapshot; the
	// late profile is absent from it and never claimed as reached.
	line := userAssetRootsLine(res)
	if !strings.Contains(line, "early") {
		t.Errorf("line must name the snapshot-reached profile; got %q", line)
	}
	if strings.Contains(line, "late") {
		t.Errorf("line must not claim a profile created after the snapshot was reached; got %q", line)
	}
}
