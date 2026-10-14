package hook

// user_root_symlink_real_path_test.go — card t1591 (SPEC-USERASSET-DEPLOY-GUARD-001),
// codex round-2 finding 2: a managed asset reached through a symlink is protected by
// its real path too.
//
// The manifest tracks the LINK path (~/.claude/skills/demo/SKILL.md), and the link
// points to a directory outside the install roots. A Write or Edit to the real file
// behind the link matched no user-root form, so the tracked-containment check was
// never reached and the write was allowed.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserRootRealPathOfSymlinkedManagedAssetIsProtected(t *testing.T) {
	root := t.TempDir()    // the project the hook guards
	home := t.TempDir()    // the user home that holds the install roots
	outside := t.TempDir() // a directory outside the install roots that a managed link points to

	// The loaded protection set: the user-root entry for the claude-skills root.
	overlayDir := filepath.Join(root, ".moai", "project")
	if err := os.MkdirAll(overlayDir, 0o755); err != nil {
		t.Fatal(err)
	}
	overlay := `version: 1
categories:
  managed_assets:
    runtime_paths:
      - user-root:claude-skills/
`
	if err := os.WriteFile(filepath.Join(overlayDir, "protected-zone.yaml"), []byte(overlay), 0o644); err != nil {
		t.Fatal(err)
	}

	// The managed asset: ~/.claude/skills/demo is a symlink to a directory outside the
	// install roots. The real SKILL.md sits behind it.
	realDemo := filepath.Join(outside, "demo")
	if err := os.MkdirAll(realDemo, 0o755); err != nil {
		t.Fatal(err)
	}
	realSkill := filepath.Join(realDemo, "SKILL.md")
	if err := os.WriteFile(realSkill, []byte("managed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	skillsDir := filepath.Join(home, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkSkipWindows(t, realDemo, filepath.Join(skillsDir, "demo"))

	// A tracked ordinary asset under the install root (no link), the control that the
	// user-root protection is live.
	ordinary := filepath.Join(skillsDir, "moai-alpha", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(ordinary), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ordinary, []byte("managed\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// An ordinary file outside every install root: not a managed asset.
	unrelated := filepath.Join(outside, "notes", "todo.md")
	if err := os.MkdirAll(filepath.Dir(unrelated), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unrelated, []byte("a user's note\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The manifest records the LINK path — the key the tracked judgment reads.
	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version": 1, "files": {"claude-skills/demo/SKILL.md": {"sha256": "ab"}, "claude-skills/moai-alpha/SKILL.md": {"sha256": "cd"}}}`
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	prevGetwd := zoneGetwd
	zoneGetwd = func() (string, error) { return root, nil }
	t.Cleanup(func() { zoneGetwd = prevGetwd })
	prevHome := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = prevHome })

	h := &preToolHandler{policy: DefaultSecurityPolicy(), projectDir: root}
	denied := func(tool, path string) bool {
		t.Helper()
		sentinel, _ := h.checkProtectedZone(harnessLearnerIdentity, tool, path)
		return sentinel == SentinelHarnessFrozenProtectedZone
	}

	// CONTROL: the tracked ordinary asset under the root is denied, so the user-root
	// protection and its config are live. A dead guard would make the defect arm
	// below meaningless.
	if !denied("Write", ordinary) {
		t.Fatalf("control arm broke: a Write to the tracked ordinary asset %s was not denied — the user-root guard is not live, so the defect arm below proves nothing", ordinary)
	}

	// DEFECT ARM: the real SKILL.md behind the symlinked managed asset is denied for
	// both tools, as the link-path spelling already is.
	for _, tool := range []string{"Write", "Edit"} {
		if !denied(tool, realSkill) {
			t.Errorf("%s to the real path %s behind the symlinked managed asset was allowed — the tracked asset ~/.claude/skills/demo/SKILL.md is not protected by its real path", tool, realSkill)
		}
	}

	// CONTROL: an unrelated file outside every install root stays editable.
	if denied("Write", unrelated) {
		t.Errorf("a Write to the unrelated outside file %s was denied — the real-path protection over-protects a file that is not a managed asset", unrelated)
	}
}
