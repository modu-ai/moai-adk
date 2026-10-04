package template

// deployer_mode_test.go — the M2 deploy-mode split tests (SPEC-INIT-SHRINK-001
// AC-001 (b) / AC-002 (a) / AC-006). The fixtures use small in-memory trees
// shaped like the real one (skills + commands + a rules file), so the mode
// semantics are pinned without the embedded tree's size.

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// modeTestFS builds a template tree shaped like the real one: a catalog
// skill (.claude/skills/moai-fake), a published command skill under
// .agents/skills (a DIFFERENT name — the real tree's catalog and published
// sets never collide), one command, one rule, and one config render.
func modeTestFS() fstest.MapFS {
	return fstest.MapFS{
		".claude/skills/moai-fake/SKILL.md":     &fstest.MapFile{Data: []byte("# fake skill\n")},
		".claude/skills/moai-fake/ref.md":       &fstest.MapFile{Data: []byte("ref\n")},
		".claude/skills/moai-fake/second.md":    &fstest.MapFile{Data: []byte("second\n")},
		".claude/commands/moai/fake.md":         &fstest.MapFile{Data: []byte("# fake command\n")},
		".claude/rules/moai/core/rule.md":       &fstest.MapFile{Data: []byte("rule\n")},
		".agents/skills/moai-fake-cmd/SKILL.md": &fstest.MapFile{Data: []byte("# published command skill\n")},
		".moai/config/sections/llm.yaml":        &fstest.MapFile{Data: []byte("llm:\n  harness: claude\n")},
	}
}

// modeDeploy constructs a renderer-less deployer over modeTestFS with the
// given options and deploys into a fresh temp root.
func modeDeploy(t *testing.T, opts ...DeployerOption) (string, DeployResult) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("manifest Load: %v", err)
	}
	dep := NewDeployer(modeTestFS(), opts...)
	res, err := dep.(ResultDeployer).DeployWithResult(context.Background(), root, mgr, nil)
	if err != nil {
		t.Fatalf("DeployWithResult: %v", err)
	}
	return root, *res
}

// deployedFiles lists the files a deploy wrote, project-root-relative slash
// paths.
func deployedFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk deployed tree: %v", err)
	}
	return files
}

// TestDeployerModeSplitsFileSet is AC-001 (b): in plugin mode the deploy
// walk and ListTemplates carry no .claude/skills/** or .claude/commands/**
// file while everything else the tree carries still deploys; local mode
// (the zero value) deploys the whole tree byte-for-byte as today.
func TestDeployerModeSplitsFileSet(t *testing.T) {
	t.Run("local-deploys-everything", func(t *testing.T) {
		root, _ := modeDeploy(t)
		files := deployedFiles(t, root)
		for _, want := range []string{
			".claude/skills/moai-fake/SKILL.md",
			".claude/commands/moai/fake.md",
			".claude/rules/moai/core/rule.md",
			".moai/config/sections/llm.yaml",
		} {
			if !containsPath(files, want) {
				t.Errorf("local deploy missing %s (deployed: %v)", want, files)
			}
		}
	})

	t.Run("plugin-excludes-skills-and-commands", func(t *testing.T) {
		root, _ := modeDeploy(t, WithDeployMode(DeployModePlugin))
		files := deployedFiles(t, root)
		for _, f := range files {
			if strings.HasPrefix(f, ".claude/skills/") || strings.HasPrefix(f, ".claude/commands/") {
				t.Errorf("plugin deploy wrote dropped component %s", f)
			}
		}
		for _, want := range []string{
			".claude/rules/moai/core/rule.md",
			".moai/config/sections/llm.yaml",
		} {
			if !containsPath(files, want) {
				t.Errorf("plugin deploy dropped a kept component %s (deployed: %v)", want, files)
			}
		}
	})

	t.Run("plugin-listtemplates-excludes-dropped", func(t *testing.T) {
		dep := NewDeployer(modeTestFS(), WithDeployMode(DeployModePlugin))
		for _, name := range dep.ListTemplates() {
			if strings.HasPrefix(name, ".claude/skills/") || strings.HasPrefix(name, ".claude/commands/") {
				t.Errorf("plugin ListTemplates carries dropped component %s", name)
			}
		}
		found := false
		for _, name := range dep.ListTemplates() {
			if name == ".claude/rules/moai/core/rule.md" {
				found = true
			}
		}
		if !found {
			t.Error("plugin ListTemplates lost a kept component")
		}
	})
}

// TestEmbeddedSkillAndCommandSourcesRetained is AC-002 (a): the embedded
// template tree keeps every skill and command source it carried at the base
// tree (they remain the plugin derivation source and the local-deploy
// source — REQ-002), and catalog.yaml parses with the pinned tier counts.
func TestEmbeddedSkillAndCommandSourcesRetained(t *testing.T) {
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	skills := 0
	commands := 0
	err = fs.WalkDir(embedded, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skill directories are counted at their top-level entry (the P-02
		// pin counts directories, not SKILL.md files — three dirs carry
		// none).
		if rel, ok := strings.CutPrefix(path, ".claude/skills/"); ok {
			if !strings.Contains(rel, "/") && d.IsDir() {
				skills++
			}
			return nil
		}
		if strings.HasPrefix(path, ".claude/commands/moai/") && !d.IsDir() {
			commands++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded tree: %v", err)
	}
	if skills != 38 {
		t.Errorf("embedded skill directory count = %d, want 38", skills)
		// Plan-pin note: spec.md P-02/L-24 recorded 41, measured through the
		// session shell's `ls` alias (`ls -la`), whose output adds the
		// total line and the . and .. entries to the 38 real directories.
		// /bin/ls and find agree on 38; this pin is the unpolluted count.
	}
	if commands != 17 {
		t.Errorf("embedded command count = %d, want 17 (the base-tree pin, P-03)", commands)
	}

	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	// The P-02 pin counts `tier: <x>` markers: 36 core (skills AND agents),
	// 13 optional-pack, 1 harness-generated.
	tiers := map[string]int{}
	packTier := 0
	for _, entry := range cat.AllEntries() {
		switch entry.Tier {
		case "core", "harness-generated":
			tiers[entry.Tier]++
		default:
			packTier++ // FormatOptionalPackTier(<pack>) values
		}
	}
	if tiers["core"] != 36 {
		t.Errorf("core tier entry count = %d, want 36", tiers["core"])
	}
	if packTier != 13 {
		t.Errorf("optional-pack tier entry count = %d, want 13", packTier)
	}
	if tiers["harness-generated"] != 1 {
		t.Errorf("harness-generated tier entry count = %d, want 1", tiers["harness-generated"])
	}
}

// TestCodexMirrorFollowsDeployMode is AC-006: plugin mode creates no
// .agents/skills symlink entries (MirrorPolicyNone) — or, where the Codex
// execution verification is not produced, deploys the mirror as REAL
// directory copies rendered from the embedded tree, never symlinks
// (MirrorPolicyRehome); local mode mirrors exactly as today.
func TestCodexMirrorFollowsDeployMode(t *testing.T) {
	t.Run("local-mirrors-as-today", func(t *testing.T) {
		root, result := modeDeploy(t)
		// P-11's two legal forms: a symlink where the OS allows it, a real
		// directory copy where it does not (sandboxed test filesystems
		// refuse symlink). What must NOT happen is the entry being absent.
		assertMirrorEntry(t, root, "moai-fake")
		_ = result
	})

	t.Run("plugin-mirror-none-deploys-no-mirror", func(t *testing.T) {
		root, _ := modeDeploy(t, WithDeployMode(DeployModePlugin), WithPluginMirrorPolicy(MirrorPolicyNone))
		if _, err := os.Lstat(filepath.Join(root, ".agents", "skills")); !os.IsNotExist(err) {
			t.Errorf(".agents/skills exists under plugin+none mode: %v", err)
		}
	})

	t.Run("plugin-mirror-rehome-deploys-real-copies", func(t *testing.T) {
		root, result := modeDeploy(t, WithDeployMode(DeployModePlugin), WithPluginMirrorPolicy(MirrorPolicyRehome))
		mirrorSkill := filepath.Join(root, ".agents", "skills", "moai-fake")
		info, err := os.Lstat(mirrorSkill)
		if err != nil {
			t.Fatalf("re-homed mirror entry missing: %v", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Fatal("re-homed mirror entry is a symlink — it would dangle once .claude/skills is not deployed (REQ-006/D-8)")
		}
		if !info.IsDir() {
			t.Fatalf("re-homed mirror entry is not a directory: %v", info.Mode())
		}
		// The copy carries the skill's files.
		if _, err := os.Stat(filepath.Join(mirrorSkill, "SKILL.md")); err != nil {
			t.Errorf("re-homed copy missing SKILL.md: %v", err)
		}
		if _, err := os.Stat(filepath.Join(mirrorSkill, "ref.md")); err != nil {
			t.Errorf("re-homed copy missing ref.md: %v", err)
		}
		for _, e := range result.SkillMirrors {
			if e.Mode == MirrorModeSymlink {
				t.Errorf("plugin+rehome recorded a symlink entry for %s", e.Skill)
			}
		}
	})
}

// assertMirrorEntry asserts the mirror entry for skill exists (either of
// P-11's materialization forms).
func assertMirrorEntry(t *testing.T, root, skill string) {
	t.Helper()
	path := filepath.Join(root, ".agents", "skills", skill)
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("mirror entry %s missing: %v", skill, err)
	}
}

// containsPath reports whether the slash-path list holds path.
func containsPath(paths []string, path string) bool {
	for _, p := range paths {
		if p == path {
			return true
		}
	}
	return false
}
