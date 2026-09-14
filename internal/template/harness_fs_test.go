package template

// SPEC-INIT-HARNESS-001 M2 — harnessFS unit tests: the codex-only deployer's
// template listing hides claude-only surfaces, keeps the universal set, and
// re-homes the catalog; a real deployment materializes .agents/skills as
// REAL directories with zero symlinks (AC-IH-004 remap integrity).

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// newCodexOnlyTestDeployer builds the codex-only deployer against the real
// embedded tree (the production path under test).
func newCodexOnlyTestDeployer(t *testing.T) Deployer {
	t.Helper()
	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	d, err := NewCodexOnlyDeployerWithRenderer(cat, NewRenderer(embedded))
	if err != nil {
		t.Fatalf("new codex-only deployer: %v", err)
	}
	return d
}

// TestCodexOnlyDeployerWalkIntegrity asserts the template listing: hidden set
// absent, universal set present, catalog re-homed under .agents/skills.
func TestCodexOnlyDeployerWalkIntegrity(t *testing.T) {
	d := newCodexOnlyTestDeployer(t)

	list := d.ListTemplates()
	if len(list) == 0 {
		t.Fatal("codex-only deployer lists zero templates")
	}
	seen := make(map[string]bool, len(list))
	for _, p := range list {
		seen[p] = true
		if p == ".claude" || strings.HasPrefix(p, ".claude/") || p == "CLAUDE.md" || p == ".mcp.json" ||
			p == ".claudeignore" || p == ".moai/status_line.sh" || p == ".moai/status_line.sh.tmpl" {
			t.Errorf("hidden path %q visible in codex-only template listing (REQ-IH-005)", p)
		}
	}
	for _, want := range []string{
		".moai/config/sections/llm.yaml",
		"AGENTS.md",
		".gitignore",
		".agents/skills/moai-workflow-tdd/SKILL.md", // remapped catalog skill
		".agents/skills/moai-plan/SKILL.md",         // published skill
	} {
		if !seen[want] {
			t.Errorf("%q missing from codex-only template listing", want)
		}
	}
}

// TestCodexOnlyDeployerRealDeployment deploys into a temp dir and checks the
// remapped skills are REAL directories (no symlinks) with readable content
// (AC-IH-004).
func TestCodexOnlyDeployerRealDeployment(t *testing.T) {
	d := newCodexOnlyTestDeployer(t)

	root := t.TempDir()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	skillsDir := filepath.Join(root, ".agents", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("read .agents/skills: %v", err)
	}
	symlinks := 0
	for _, e := range entries {
		if e.Type()&fs.ModeSymlink != 0 {
			symlinks++
			t.Errorf("symlink at .agents/skills/%s — codex-only must materialize real directories (REQ-IH-006)", e.Name())
		}
	}
	if symlinks > 0 {
		t.Errorf("symlink count = %d, want 0", symlinks)
	}

	// Every catalog skill from the embedded .claude/skills listing is present
	// as a real directory under .agents/skills with a readable SKILL.md.
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded: %v", err)
	}
	catalogEntries, err := fs.ReadDir(embedded, filepath.ToSlash(filepath.Join(".claude", "skills")))
	if err != nil {
		t.Fatalf("read catalog listing: %v", err)
	}
	for _, e := range catalogEntries {
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(skillsDir, e.Name(), "SKILL.md"))
		if err != nil {
			t.Errorf("remapped catalog skill %s unreadable: %v", e.Name(), err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("remapped catalog skill %s: SKILL.md empty", e.Name())
		}
	}

	// The claude-only surfaces did not deploy.
	for _, rel := range []string{".claude", "CLAUDE.md", ".mcp.json", ".claudeignore", ".moai/status_line.sh"} {
		if _, err := os.Stat(filepath.Join(root, rel)); err == nil {
			t.Errorf("%s deployed by codex-only deployer (REQ-IH-005 violation)", rel)
		}
	}
}

// TestCodexOnlyRemapSkipsOccupiedTarget covers REQ-IH-007: a codex-only
// remap whose destination path is already occupied by a user-owned
// non-symlink entry is left untouched and reported, not overwritten.
func TestCodexOnlyRemapSkipsOccupiedTarget(t *testing.T) {
	d := newCodexOnlyTestDeployer(t)

	root := t.TempDir()
	// A user-owned file occupies a remapped destination path.
	occupied := filepath.Join(root, ".agents", "skills", "moai-workflow-tdd", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(occupied), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(occupied, []byte("user-owned bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	rd, ok := d.(ResultDeployer)
	if !ok {
		t.Fatal("codex-only deployer must implement ResultDeployer")
	}
	res, err := rd.DeployWithResult(context.Background(), root, mgr, nil)
	if err != nil {
		t.Fatalf("deploy: %v", err)
	}

	// The user-owned file survives byte-identically.
	data, err := os.ReadFile(occupied)
	if err != nil {
		t.Fatalf("occupied file vanished: %v", err)
	}
	if string(data) != "user-owned bytes" {
		t.Error("occupied file was overwritten by the remap (REQ-IH-007 violation)")
	}
	// And the skip is reported somewhere observable.
	if len(res.ProtectedSkips) == 0 && len(res.Warnings()) == 0 {
		t.Error("occupied remap target: neither ProtectedSkips nor Warnings recorded the skip")
	}
}
