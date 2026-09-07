// published_skills_deploy_test.go — SPEC-CODEX-COMMAND-SKILLS-001 M3:
// deploy-side distribution (AC-009), update-mode user-file safety (R-011),
// and mirror/published coexistence (AC-008).
package template

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// publishedSkillsNamesMatchTree guards the embedded publishedSkillNames
// list against drift from the committed published tree. The chain is:
// emitter output == committed tree (the drift check) == this list (here) —
// a command added without extending the list would deploy unprotected.
func TestPublishedSkillsNamesMatchTree(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("templates", ".agents", "skills"))
	if err != nil {
		t.Fatalf("read committed published tree: %v", err)
	}
	seen := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		seen[e.Name()] = true
		if _, listed := publishedSkillNames[e.Name()]; !listed {
			t.Errorf("committed published skill %q missing from publishedSkillNames — it would deploy without update-mode protection", e.Name())
		}
	}
	for name := range publishedSkillNames {
		if !seen[name] {
			t.Errorf("publishedSkillNames carries %q with no committed directory under templates/.agents/skills/", name)
		}
	}
	if len(seen) != 16 {
		t.Errorf("committed published tree holds %d directories, want 16", len(seen))
	}
}

func TestPublishedSkillPathScope(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{".agents/skills/moai-plan/SKILL.md", true},
		{".agents/skills/moai-todo/SKILL.md", true},
		// Mirror entry shapes: not published-skill paths.
		{".agents/skills/moai-workflow-tdd", false},
		{".agents/skills/moai-workflow-tdd/SKILL.md", false},
		{".agents/skills/moai/SKILL.md", false},
		// §24.4 scope is untouched.
		{".claude/skills/moai-plan/SKILL.md", false},
		{".claude/commands/moai/plan.md.tmpl", false},
	}
	for _, tc := range cases {
		if got := isPublishedSkillPath(tc.path); got != tc.want {
			t.Errorf("isPublishedSkillPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

// TestPublishedSkillInitDistribution is AC-009's distribution half: a fresh
// t.TempDir() project deployed from the embedded template FS carries all 16
// published skills at .agents/skills/moai-<command>/SKILL.md, tracked
// template-managed.
func TestPublishedSkillInitDistribution(t *testing.T) {
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("manifest Load: %v", err)
	}
	deployer := NewDeployer(embedded)
	if err := deployer.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("Deploy: %v", err)
	}

	commands := []string{"clean", "codemaps", "e2e", "feedback", "fix", "gate",
		"goal", "harness", "loop", "mx", "plan", "project", "review", "run",
		"sync", "todo"}
	if len(commands) != 16 {
		t.Fatalf("fixture inventory drifted: %d commands", len(commands))
	}
	for _, cmd := range commands {
		rel := ".agents/skills/moai-" + cmd + "/SKILL.md"
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("published skill %s missing after init: %v", rel, err)
			continue
		}
		if !strings.Contains(string(data), "name: moai-"+cmd+"\n") {
			t.Errorf("%s: deployed bytes lack the moai-%s identity", rel, cmd)
		}
		entry, found := mgr.GetEntry(rel)
		if !found {
			t.Errorf("%s: not tracked in manifest", rel)
			continue
		}
		if entry.Provenance != manifest.TemplateManaged {
			t.Errorf("%s: provenance = %v, want %v", rel, entry.Provenance, manifest.TemplateManaged)
		}
	}
}

// publishedSkillFixtureFS builds a small tree carrying one published skill
// plus one ordinary file.
func publishedSkillFixtureFS(body string) fstest.MapFS {
	return fstest.MapFS{
		".agents/skills/moai-plan/SKILL.md": &fstest.MapFile{Data: []byte(body)},
		"CLAUDE.md":                         &fstest.MapFile{Data: []byte("# directive\n")},
	}
}

func setupManifestProject(t *testing.T) (string, manifest.Manager) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("manifest Load: %v", err)
	}
	return root, mgr
}

// TestPublishedSkillUpdateProtectsUserOwned is R-011/AC-009's update-mode
// half: where a user-owned entry pre-occupies a published-skill path, a
// forceUpdate deploy leaves it untouched and reports the skip.
func TestPublishedSkillUpdateProtectsUserOwned(t *testing.T) {
	root, mgr := setupManifestProject(t)
	userBody := "# the user's own skill\n"
	publishedRel := ".agents/skills/moai-plan/SKILL.md"
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, publishedRel)), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, publishedRel), []byte(userBody), 0o644); err != nil {
		t.Fatalf("write user file: %v", err)
	}

	deployer := NewDeployerWithForceUpdate(publishedSkillFixtureFS("# publisher bytes\n"), true)
	res, err := deployer.(ResultDeployer).DeployWithResult(context.Background(), root, mgr, nil)
	if err != nil {
		t.Fatalf("DeployWithResult: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, publishedRel))
	if err != nil {
		t.Fatalf("read published path: %v", err)
	}
	if string(got) != userBody {
		t.Errorf("update overwrote a user-owned published-skill file: %q", got)
	}
	if len(res.ProtectedSkips) != 1 || res.ProtectedSkips[0] != publishedRel {
		t.Errorf("ProtectedSkips = %v, want [%s]", res.ProtectedSkips, publishedRel)
	}
	entry, found := mgr.GetEntry(publishedRel)
	if !found {
		t.Fatalf("skip not recorded in manifest")
	}
	if entry.Provenance != manifest.UserCreated {
		t.Errorf("provenance after skip = %v, want %v", entry.Provenance, manifest.UserCreated)
	}
}

// TestPublishedSkillUpdateOverwritesTemplateManaged pins the other half of
// the R-011 scope: our own template-managed published entries still
// overwrite normally in update mode — that is how updates deliver content.
func TestPublishedSkillUpdateOverwritesTemplateManaged(t *testing.T) {
	root, mgr := setupManifestProject(t)

	initDeployer := NewDeployer(publishedSkillFixtureFS("# version A\n"))
	if err := initDeployer.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("init Deploy: %v", err)
	}

	publishedRel := ".agents/skills/moai-plan/SKILL.md"
	updateDeployer := NewDeployerWithForceUpdate(publishedSkillFixtureFS("# version B\n"), true)
	if err := updateDeployer.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("update Deploy: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, publishedRel))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "# version B\n" {
		t.Errorf("template-managed published entry not refreshed by update: %q", got)
	}
	if entry, _ := mgr.GetEntry(publishedRel); entry.Provenance != manifest.TemplateManaged {
		t.Errorf("provenance = %v, want %v", entry.Provenance, manifest.TemplateManaged)
	}
}

// TestUpdateForceStillOverwritesOutsideScope documents the scope boundary:
// outside the published-skill namespace, forceUpdate keeps its existing
// overwrite-everything behavior (no other path's semantics change).
func TestUpdateForceStillOverwritesOutsideScope(t *testing.T) {
	root, mgr := setupManifestProject(t)
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("# user edit\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	deployer := NewDeployerWithForceUpdate(publishedSkillFixtureFS("x"), true)
	res, err := deployer.(ResultDeployer).DeployWithResult(context.Background(), root, mgr, nil)
	if err != nil {
		t.Fatalf("DeployWithResult: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "# directive\n" {
		t.Errorf("forceUpdate semantics changed outside the published scope: %q", got)
	}
	if len(res.ProtectedSkips) != 0 {
		t.Errorf("unexpected protected skips outside scope: %v", res.ProtectedSkips)
	}
}

// TestMirrorCommandSkillCoexistence is AC-008: mirror creation neither
// clobbers a published directory (even under a colliding name, the mirror
// skips a non-symlink occupant) nor is disturbed by published entries
// sitting next to its own.
func TestMirrorCommandSkillCoexistence(t *testing.T) {
	root := t.TempDir()
	canonicalRel := filepath.Join(".claude", "skills", "moai-workflow-tdd", "SKILL.md")
	publishedRel := filepath.Join(".agents", "skills", "moai-plan", "SKILL.md")
	publishedBody := "# published command skill\n"

	for _, rel := range []string{canonicalRel, publishedRel} {
		if err := os.MkdirAll(filepath.Join(root, filepath.Dir(rel)), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		body := "# canonical skill\n"
		if rel == publishedRel {
			body = publishedBody
		}
		if err := os.WriteFile(filepath.Join(root, rel), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	d := NewDeployer(fstest.MapFS{}).(*deployer)
	entries := d.mirrorSkills(root, []string{"moai-workflow-tdd", "moai-plan"})

	modes := map[string]MirrorMode{}
	for _, e := range entries {
		modes[e.Skill] = e.Mode
	}
	// The canonical skill's mirror materializes (symlink, or copy where
	// links are unavailable) — published entries did not block it.
	if m := modes["moai-workflow-tdd"]; m != MirrorModeSymlink && m != MirrorModeCopy {
		t.Errorf("canonical mirror mode = %v, want symlink or copy", m)
	}
	// The published directory occupies its path as a non-symlink entry: the
	// mirror must skip and report, never remove or overwrite.
	if m := modes["moai-plan"]; m != MirrorModeSkipped {
		t.Errorf("published-path mirror mode = %v, want skipped", m)
	}
	got, err := os.ReadFile(filepath.Join(root, publishedRel))
	if err != nil {
		t.Fatalf("read published skill: %v", err)
	}
	if string(got) != publishedBody {
		t.Errorf("mirror run modified the published skill: %q", got)
	}
}
