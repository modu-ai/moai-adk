package hook

// user_root_symlink_siblings_test.go — card t1591 (SPEC-USERASSET-DEPLOY-GUARD-001),
// codex round-3 finding (P1): an operation on a directory ABOVE a managed asset's link
// target was not judged. The link-prefix match compared the candidate with the target
// for equality only, so rm -rf of the target's parent passed.
//
// The table measures the sibling cases of that defect on the current tree: the upper
// directories (parent, grandparent, install root), the target itself, the link and the
// real SKILL.md, lower directories with and without a tracked file, and the mv and
// rename forms of each. Every denied family carries an allowed control, so a repair
// that widens the protection to unrelated paths turns a control red.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUserRootSymlinkSiblings(t *testing.T) {
	// siblingRow is one measured case: a Bash command (shell rows) or a Write/Edit
	// (file rows), with the verdict it must receive. want is true when denied.
	type siblingRow struct {
		id, subject string
		command     string
		tool, file  string
		want        bool
	}

	root := t.TempDir()  // the project the hook guards
	home := t.TempDir()  // the user home that holds the install roots
	outer := t.TempDir() // the grandparent of the link target
	mid := filepath.Join(outer, "mid")
	realDemo := filepath.Join(mid, "demo") // the managed asset's real directory: the link target
	skillsDir := filepath.Join(home, ".claude", "skills")
	linkPath := filepath.Join(skillsDir, "demo")

	writeFile := func(p string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

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

	// The managed asset: ~/.claude/skills/demo links to <outer>/mid/demo. Tracked files
	// sit behind the link, one at the top and one in a lower directory.
	writeFile(filepath.Join(realDemo, "SKILL.md"))
	writeFile(filepath.Join(realDemo, "refs", "notes.md"))
	// Untracked paths: a lower directory, a file in the target, and an unrelated sibling.
	writeFile(filepath.Join(realDemo, "scratch", "log.txt"))
	writeFile(filepath.Join(realDemo, "notes.md"))
	writeFile(filepath.Join(mid, "other", "file.txt"))
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkSkipWindows(t, realDemo, linkPath)
	// A tracked ordinary asset under the install root (no link): the control that the
	// user-root protection is live. An untracked directory there is the rename control.
	writeFile(filepath.Join(skillsDir, "moai-alpha", "SKILL.md"))
	writeFile(filepath.Join(skillsDir, "scratch-skill", "notes.md"))

	// The manifest records the LINK paths, the keys the tracked judgment reads.
	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version": 1, "files": {"claude-skills/demo/SKILL.md": {"sha256": "ab"}, "claude-skills/demo/refs/notes.md": {"sha256": "ef"}, "claude-skills/moai-alpha/SKILL.md": {"sha256": "cd"}}}`
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	prevGetwd := zoneGetwd
	zoneGetwd = func() (string, error) { return root, nil }
	t.Cleanup(func() { zoneGetwd = prevGetwd })
	prevHome := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = prevHome })

	realSkill := filepath.Join(realDemo, "SKILL.md")
	linkSkill := filepath.Join(linkPath, "SKILL.md")
	lowerRefs := filepath.Join(realDemo, "refs")
	lowerScratch := filepath.Join(realDemo, "scratch")
	notes := filepath.Join(realDemo, "notes.md")
	other := filepath.Join(mid, "other")
	scratchSkill := filepath.Join(skillsDir, "scratch-skill")

	shellRow := func(id, command string, want bool) siblingRow {
		command = zoneBashText(command)
		return siblingRow{id: id, subject: command, command: command, want: want}
	}
	fileRow := func(id, tool, file string, want bool) siblingRow {
		return siblingRow{id: id, subject: tool + " " + file, tool: tool, file: file, want: want}
	}

	rows := []siblingRow{
		// upper directories: removing an ancestor of the link target removes the asset
		shellRow("R1", "rm -rf "+mid, true),
		shellRow("R2", "rm -rf "+realDemo, true),
		shellRow("R6", "rm -rf "+outer, true),
		shellRow("R1r", "rm -rf "+skillsDir, true),
		shellRow("R11", "rm -rf /", true),
		// the tracked file, through its real path and through the link
		shellRow("R3", "rm -f "+realSkill, true),
		shellRow("R3l", "rm -f "+linkSkill, true),
		// lower directories: denied when a tracked file sits under them, allowed when none does
		shellRow("R5", "rm -rf "+lowerRefs, true),
		shellRow("R5c", "rm -rf "+lowerScratch, false),
		// an unrelated sibling of the target under the same parent (control)
		shellRow("R4", "rm -rf "+other, false),
		// mv: the tracked file out of the target, the target, its parent, and the link
		shellRow("R7", "mv "+realSkill+" "+filepath.Join(outer, "moved.md"), true),
		shellRow("R7c", "mv "+notes+" "+filepath.Join(realDemo, "notes2.md"), false),
		shellRow("R8", "mv "+realDemo+" "+filepath.Join(mid, "demo2"), true),
		shellRow("R9", "mv "+mid+" "+filepath.Join(outer, "mid2"), true),
		shellRow("R10", "mv "+linkPath+" "+filepath.Join(skillsDir, "demo2"), true),
		shellRow("R4m", "mv "+other+" "+filepath.Join(mid, "other2"), false),
		// the rename utility on the tracked link, and on an untracked directory (control)
		shellRow("R10r", "rename demo demo2 "+linkPath, true),
		shellRow("R10c", "rename scratch-skill scratch-skill2 "+scratchSkill, false),
		// Write and Edit: the file-tool arm, over the same user-root forms
		fileRow("W1", "Write", realSkill, true),
		fileRow("W2", "Edit", linkSkill, true),
		fileRow("W3", "Write", notes, false),
	}

	h := &preToolHandler{policy: DefaultSecurityPolicy(), projectDir: root}
	denied := func(t *testing.T, row siblingRow) bool {
		t.Helper()
		if row.tool != "" {
			sentinel, _ := h.checkProtectedZone(harnessLearnerIdentity, row.tool, row.file)
			return sentinel != ""
		}
		raw, err := json.Marshal(map[string]string{"command": row.command})
		if err != nil {
			t.Fatal(err)
		}
		return h.checkProtectedZoneShell(harnessLearnerIdentity, raw) != ""
	}
	verdict := map[bool]string{true: "denied", false: "allowed"}
	for _, row := range rows {
		t.Run(row.id, func(t *testing.T) {
			if got := denied(t, row); got != row.want {
				t.Errorf("row %s (%s): expected %s, observed %s", row.id, row.subject, verdict[row.want], verdict[got])
			}
		})
	}
}
