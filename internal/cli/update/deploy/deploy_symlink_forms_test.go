package deploy

// Five-form bipolar symlink fixtures for CleanMoaiManagedPaths
// (SPEC-CLI-CLEAN-SYMLINK-001, plan M2). Every fixture uses the SAME link
// form as the product path it verifies (REQ-CSL-010, t81 D2): FX-1 fixtures
// are directory links, FX-2 is a file link, FX-3 are dangling links at the
// three placements (non-glob root / glob-match name / .moai/config root),
// FX-4 and FX-5 are link-free controls on the must-NOT-flag poles. Every
// link assertion combines at least two independent observation axes
// (REQ-CSL-011, t81 D4) — link existence, target integrity, backup content,
// progress output — never a bare backup-count number.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// --- FX-3: dangling links, remaining two placements (AC-CSL-002/003) ---

// TestCleanMoaiManagedPaths_DanglingSymlinkAtGlobMatchName is AC-CSL-002: a
// dangling link whose name matches the .claude/skills/moai* glob used to be
// a silent permanent no-op (Run D) — nothing removed it and no message named
// it. Axes: link existence + progress line + sibling skill redeployability.
func TestCleanMoaiManagedPaths_DanglingSymlinkAtGlobMatchName(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir)
	tmplSkill := filepath.Join(skillsDir, "moai")
	if err := os.MkdirAll(tmplSkill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmplSkill, "SKILL.md"), []byte("user copy"), 0o644); err != nil {
		t.Fatal(err)
	}
	danglingRel := filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai-dangling-custom")
	makeSymlink(t, filepath.Join(root, "t173-also-gone"), filepath.Join(root, danglingRel))

	var out bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out, preCleanTestFS(".claude/skills/moai/SKILL.md")); err != nil {
		t.Errorf("clean failed: %v", err)
	}

	// Then 1: the dangling link is gone (previously it survived forever).
	if _, lerr := os.Lstat(filepath.Join(root, danglingRel)); !os.IsNotExist(lerr) {
		t.Errorf("glob-match dangling symlink still present after clean (Lstat err = %v)", lerr)
	}
	// Then 2: a progress line names the link path and its symlink form.
	if !symlinkProgressLine(out.String(), danglingRel) {
		t.Errorf("progress output has no line naming %s as a symlink:\n%s", danglingRel, out.String())
	}
	// Then 3: the sibling template skill was removed by the same glob sweep
	// (an MkdirAll-only check would pass vacuously on a surviving dir) and
	// redeployment lands it again (≥1 file, real directory).
	if _, serr := os.Lstat(tmplSkill); !os.IsNotExist(serr) {
		t.Errorf("sibling template skill not removed by the glob sweep: %v", serr)
	}
	deploySimMkdirAll(t, tmplSkill)
	if err := os.WriteFile(filepath.Join(tmplSkill, "SKILL.md"), []byte("template:skill"), 0o644); err != nil {
		t.Errorf("redeploy write failed: %v", err)
	}
	if fi, serr := os.Lstat(tmplSkill); serr != nil || !fi.Mode().IsDir() {
		t.Errorf("template skill not redeployable as a real directory: %v", serr)
	}
}

// TestCleanMoaiManagedPaths_DanglingSymlinkAtConfigRoot is AC-CSL-003 — the
// eighth removal root (.moai/config, no pre-check) shares backupThenRemove,
// so the dangling disposition is inherited (dossier gap 4: code-trace only
// until this test). Axes: link existence + redeploy + EEXIST-free deploy +
// the config root's own progress line (Then 4 pins that an implementation
// naming only the seven ManagedCleanTargets roots cannot pass).
func TestCleanMoaiManagedPaths_DanglingSymlinkAtConfigRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, defs.MoAIDir), 0o755); err != nil {
		t.Fatal(err)
	}
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	makeSymlink(t, filepath.Join(root, "t173-config-gone"), configDir)

	var out bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out, preCleanTestFS(".moai/config/config.yaml")); err != nil {
		t.Errorf("clean failed: %v", err)
	}

	// Then 1: the link is removed.
	if _, lerr := os.Lstat(configDir); !os.IsNotExist(lerr) {
		t.Errorf("config dangling symlink still present after clean (Lstat err = %v)", lerr)
	}
	// Then 3 (checked via the deploy simulation below) + Then 2: .moai/config
	// is redeployable as a real directory with template content.
	deploySimMkdirAll(t, configDir)
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("template:config"), 0o644); err != nil {
		t.Errorf("redeploy write failed: %v", err)
	}
	if fi, serr := os.Lstat(configDir); serr != nil || !fi.Mode().IsDir() {
		t.Errorf(".moai/config not redeployed as a real directory: %v", serr)
	}
	// Then 4: the config root gets its own dangling-naming line.
	if !symlinkProgressLine(out.String(), filepath.Join(defs.MoAIDir, defs.ConfigSubdir)) {
		t.Errorf("progress output has no line naming .moai/config as a symlink:\n%s", out.String())
	}
}

// TestCleanMoaiManagedPaths_DanglingSymlinkAtFileRoot closes dossier gap 3
// (§D.3 first edge): settings.json as a dangling link. The clean step must
// remove the link and name it; the deploy side (atomicWriteFile's rename)
// then succeeds because the destination is a plain absent path.
func TestCleanMoaiManagedPaths_DanglingSymlinkAtFileRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, defs.ClaudeDir), 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(root, defs.ClaudeDir, defs.SettingsJSON)
	makeSymlink(t, filepath.Join(root, "t173-settings-gone"), settingsPath)

	var out bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out, preCleanTestFS(".claude/settings.json.tmpl")); err != nil {
		t.Errorf("clean failed: %v", err)
	}

	if _, lerr := os.Lstat(settingsPath); !os.IsNotExist(lerr) {
		t.Errorf("file-root dangling symlink still present after clean (Lstat err = %v)", lerr)
	}
	if !symlinkProgressLine(out.String(), filepath.Join(defs.ClaudeDir, defs.SettingsJSON)) {
		t.Errorf("progress output has no line naming .claude/settings.json as a symlink:\n%s", out.String())
	}
	// Deploy-side simulation: the written settings.json is a real file.
	if err := os.WriteFile(settingsPath, []byte("{}"), 0o644); err != nil {
		t.Errorf("deploy-side write failed: %v", err)
	}
	if fi, serr := os.Lstat(settingsPath); serr != nil || fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf("settings.json is not a real file after redeploy: %v", serr)
	}
}

// --- FX-1: live directory links (AC-CSL-004, both placements) ---

// TestCleanMoaiManagedPaths_LiveDirectorySymlinkRoots is AC-CSL-004. The
// fixture form is a DIRECTORY link (never a file link — REQ-CSL-010: the two
// forms ride different branches). Axes: link-vs-real-directory distinction +
// outside target integrity (sentinel content) + progress line, with the
// zero-backup count as an auxiliary axis only (t81 D4: WalkDir-skip makes a
// bare count assertion hollow).
func TestCleanMoaiManagedPaths_LiveDirectorySymlinkRoots(t *testing.T) {
	for _, tc := range []struct {
		name     string
		relRoot  string // clean-root path relative to the project root
		tmplPath string // one template file under the root
	}{
		{
			name:     "non-glob root",
			relRoot:  filepath.Join(defs.ClaudeDir, defs.RulesMoaiSubdir),
			tmplPath: ".claude/rules/moai/NOTICE.md",
		},
		{
			name:     "glob match name",
			relRoot:  filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai-livelink"),
			tmplPath: ".claude/skills/moai/SKILL.md",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outsideRoot := t.TempDir()
			outsideDir := filepath.Join(outsideRoot, "outside-tree")
			if err := os.MkdirAll(outsideDir, 0o755); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(outsideDir, "sentinel.md")
			if err := os.WriteFile(sentinel, []byte("UNMANAGED-SENTINEL-v1"), 0o644); err != nil {
				t.Fatal(err)
			}

			root := t.TempDir()
			rootPath := filepath.Join(root, tc.relRoot)
			if err := os.MkdirAll(filepath.Dir(rootPath), 0o755); err != nil {
				t.Fatal(err)
			}
			makeSymlink(t, outsideDir, rootPath)

			var out bytes.Buffer
			if err := CleanMoaiManagedPaths(root, &out, preCleanTestFS(tc.tmplPath)); err != nil {
				t.Errorf("clean failed: %v", err)
			}

			// Axis 1: the link is gone and the root is redeployable as a
			// REAL directory (deploy simulation).
			if _, lerr := os.Lstat(rootPath); !os.IsNotExist(lerr) {
				t.Errorf("directory symlink still present after clean (Lstat err = %v)", lerr)
			}
			deploySimMkdirAll(t, rootPath)
			if fi, serr := os.Lstat(rootPath); serr != nil || !fi.Mode().IsDir() {
				t.Errorf("root not redeployed as a real directory: %v", serr)
			}

			// Axis 2: the outside target and its sentinel are untouched.
			data, rerr := os.ReadFile(sentinel)
			if rerr != nil {
				t.Fatalf("outside sentinel lost: %v", rerr)
			}
			if string(data) != "UNMANAGED-SENTINEL-v1" {
				t.Errorf("outside sentinel content changed: %q", data)
			}

			// Axis 3: the progress output names the link path + symlink form.
			if !symlinkProgressLine(out.String(), tc.relRoot) {
				t.Errorf("progress output has no line naming %s as a symlink:\n%s", tc.relRoot, out.String())
			}

			// Auxiliary axis: nothing under the link root reached the
			// pre-clean backup (zero-backup continuity, REQ-CSL-003). A run
			// that produced no backup tree at all satisfies the axis
			// vacuously — the backup tree is only materialized when a copy
			// actually lands somewhere.
			backups, _ := filepath.Glob(filepath.Join(root, defs.BackupsDir, "*", preCleanBackupSubdir))
			for _, bk := range backups {
				var underRoot int
				_ = filepath.WalkDir(bk, func(p string, d os.DirEntry, err error) error {
					if err == nil && !d.IsDir() && strings.Contains(p, tc.relRoot) {
						underRoot++
					}
					return nil
				})
				if underRoot != 0 {
					t.Errorf("backup tree carries %d file(s) under the directory-link root %s", underRoot, tc.relRoot)
				}
			}
		})
	}
}

// --- FX-2: live file link (AC-CSL-005, Run B) ---

// TestCleanMoaiManagedPaths_LiveFileSymlinkSettings is AC-CSL-005. The
// fixture is a FILE link (a directory link would ride FX-1's branch —
// REQ-CSL-010). Axes: backup content (bytes read through the link) + final
// real-file state + progress line + the restored-content flow Run B measured.
func TestCleanMoaiManagedPaths_LiveFileSymlinkSettings(t *testing.T) {
	outsideRoot := t.TempDir()
	outsideFile := filepath.Join(outsideRoot, "outside-settings.json")
	if err := os.WriteFile(outsideFile, []byte("OUTSIDE-SETTINGS-v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, defs.ClaudeDir), 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(root, defs.ClaudeDir, defs.SettingsJSON)
	makeSymlink(t, outsideFile, settingsPath)

	var out bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out, preCleanTestFS(".claude/settings.json.tmpl")); err != nil {
		t.Errorf("clean failed: %v", err)
	}

	// Axis 1: the pre-clean backup holds the target bytes read THROUGH the
	// link (the template carries only settings.json.tmpl, never the exact
	// path, so the linked file counts as unmanaged).
	backup := filepath.Join(preCleanBackupRoot(t, root), defs.ClaudeDir, defs.SettingsJSON)
	data, rerr := os.ReadFile(backup)
	if rerr != nil {
		t.Fatalf("linked settings.json not backed up: %v", rerr)
	}
	if string(data) != "OUTSIDE-SETTINGS-v1" {
		t.Errorf("backup content = %q, want the target bytes", data)
	}

	// Axis 2: the link is gone; deployment writes a REAL file at the root.
	if fi, lerr := os.Lstat(settingsPath); lerr == nil && fi.Mode()&os.ModeSymlink != 0 {
		t.Error("settings.json is still a symlink after clean")
	}
	// Axis 4: the restored-content flow Run B measured — the backed-up user
	// bytes come back as the final content (3-way merge input).
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Errorf("deploy-side write failed: %v", err)
	}
	final, ferr := os.ReadFile(settingsPath)
	if ferr != nil || string(final) != "OUTSIDE-SETTINGS-v1" {
		t.Errorf("final settings.json = %q (err %v), want the restored user bytes", final, ferr)
	}
	if fi, lerr := os.Lstat(settingsPath); lerr != nil || fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf("settings.json is not a real file after redeploy: %v", lerr)
	}

	// Axis 3: the progress output names the link path + symlink form.
	if !symlinkProgressLine(out.String(), filepath.Join(defs.ClaudeDir, defs.SettingsJSON)) {
		t.Errorf("progress output has no line naming .claude/settings.json as a symlink:\n%s", out.String())
	}
}

// TestBackupThenRemove_LiveFileSymlinkTemplateCarried covers the file-link
// sub-disposition where the template carries the exact relative path: the
// link is removed with the note but no backup (deployment rewrites the file,
// so its only copy is never at stake — same rule as a real carried file).
// Direct unit call because no clean root hits this shape today.
func TestBackupThenRemove_LiveFileSymlinkTemplateCarried(t *testing.T) {
	outsideRoot := t.TempDir()
	outsideFile := filepath.Join(outsideRoot, "carried.txt")
	if err := os.WriteFile(outsideFile, []byte("carried"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	linkPath := filepath.Join(root, "carried-link")
	makeSymlink(t, outsideFile, linkPath)

	backedUp, note, err := backupThenRemove(linkPath, "carried-link",
		filepath.Join(root, "bk"), preCleanTestFS("carried-link"))
	if err != nil {
		t.Fatalf("backupThenRemove: %v", err)
	}
	if backedUp != 0 {
		t.Errorf("backedUp = %d, want 0 (template carries the path)", backedUp)
	}
	if !strings.Contains(note, "symlink") {
		t.Errorf("note = %q, want a symlink-naming note", note)
	}
	if _, lerr := os.Lstat(linkPath); !os.IsNotExist(lerr) {
		t.Errorf("link still present (Lstat err = %v)", lerr)
	}
	if _, serr := os.Stat(outsideFile); serr != nil {
		t.Errorf("outside target lost: %v", serr)
	}
}

// --- error path: symlink target stat failure ---

// TestCleanMoaiManagedPaths_SymlinkTargetStatError pins the failure
// attribution when the form-classification Stat cannot resolve the target
// for a reason other than absence (EACCES): the run fails loudly naming the
// symlink target instead of guessing a disposition.
func TestCleanMoaiManagedPaths_SymlinkTargetStatError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	locked := filepath.Join(root, "locked")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locked, "target"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, locked)
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}

	agentsDir := filepath.Join(root, defs.ClaudeDir, defs.AgentsMoaiSubdir)
	if err := os.MkdirAll(filepath.Dir(agentsDir), 0o755); err != nil {
		t.Fatal(err)
	}
	makeSymlink(t, filepath.Join(locked, "target"), agentsDir)

	var out bytes.Buffer
	err := CleanMoaiManagedPaths(root, &out, preCleanTestFS())
	if err == nil {
		t.Fatal("expected an error from the unresolvable symlink target, got nil")
	}
	if !strings.Contains(err.Error(), "stat symlink target") {
		t.Errorf("error does not attribute the symlink-target stat: %v", err)
	}
}

// --- FX-4 / FX-5: controls on the must-NOT-flag poles ---

// TestCleanMoaiManagedPaths_RealDirectoryControlNoSymlinkLines is AC-CSL-006
// (FX-4): a link-free managed root with unmanaged content must keep the
// current branch semantics exactly — backup present, redeploy normal, and NO
// symlink token anywhere in the output (a link line in a link-free run would
// be a false positive).
func TestCleanMoaiManagedPaths_RealDirectoryControlNoSymlinkLines(t *testing.T) {
	root := t.TempDir()
	hooksDir := filepath.Join(root, defs.ClaudeDir, defs.HooksMoaiSubdir)
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "user-hook.sh"), []byte("user hook"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out, preCleanTestFS(".claude/hooks/moai/handle-x.sh")); err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	// Then 1: the unmanaged file reached the pre-clean backup.
	backup := filepath.Join(preCleanBackupRoot(t, root), defs.ClaudeDir, defs.HooksMoaiSubdir, "user-hook.sh")
	if data, rerr := os.ReadFile(backup); rerr != nil || string(data) != "user hook" {
		t.Errorf("unmanaged hook not backed up intact (err %v)", rerr)
	}
	// Then 2: no symlink-named line in a link-free run.
	if strings.Contains(out.String(), "symlink") {
		t.Errorf("symlink token in a link-free run (false positive):\n%s", out.String())
	}
	// Then 3: the root was removed and redeployment lands normally.
	if _, serr := os.Lstat(hooksDir); !os.IsNotExist(serr) {
		t.Errorf("hooks root not removed: %v", serr)
	}
	deploySimMkdirAll(t, hooksDir)
	if err := os.WriteFile(filepath.Join(hooksDir, "handle-x.sh"), []byte("template:hook"), 0o644); err != nil {
		t.Errorf("redeploy write failed: %v", err)
	}
}

// TestCleanMoaiManagedPaths_UserOwnedNamespaceUntouched is AC-CSL-007 (FX-5,
// Run C): hns-* user-owned paths — including a dangling link INSIDE them —
// are outside the clean set entirely. This is the must-not-flag pole that
// proves AC-CSL-002's removal is scoped to the managed namespace.
func TestCleanMoaiManagedPaths_UserOwnedNamespaceUntouched(t *testing.T) {
	root := t.TempDir()
	mine := filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir, "hns-mine")
	if err := os.MkdirAll(mine, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mine, "SKILL.md"), []byte("HNS-USER-OWNED-v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeSymlink(t, filepath.Join(root, "t173-gone"), filepath.Join(mine, "badlink"))

	var out bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out, preCleanTestFS()); err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	// Then 1: the directory and its content are intact.
	data, rerr := os.ReadFile(filepath.Join(mine, "SKILL.md"))
	if rerr != nil || string(data) != "HNS-USER-OWNED-v1" {
		t.Errorf("user-owned skill not intact (err %v, content %q)", rerr, data)
	}
	// Then 2: the user-owned dangling link SURVIVES (it is not in the clean
	// set — only managed-namespace links are removed).
	if fi, lerr := os.Lstat(filepath.Join(mine, "badlink")); lerr != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("user-owned dangling link did not survive: %v", lerr)
	}
	// Then 3: nothing under hns-mine reached any backup path.
	var hits int
	_ = filepath.WalkDir(filepath.Join(root, defs.BackupsDir), func(p string, d os.DirEntry, err error) error {
		if err == nil && strings.Contains(p, "hns-mine") {
			hits++
		}
		return nil
	})
	if hits != 0 {
		t.Errorf("backup tree carries %d path(s) under hns-mine", hits)
	}
	if strings.Contains(out.String(), "hns-mine") {
		t.Errorf("progress output names the user-owned namespace:\n%s", out.String())
	}
}
