package template

// SPEC-INIT-HARNESS-001 M2 — harnessFS unit tests: the codex-only deployer's
// template listing hides claude-only surfaces, keeps the universal set, and
// re-homes the catalog; a real deployment materializes .agents/skills as
// REAL directories with zero symlinks (AC-IH-004 remap integrity).

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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
		// DEPLOY-TARGET name, not the template-source name. `ListTemplates`
		// strips the `.tmpl` suffix (deployer.go), so the mirror — which ships
		// as `AGENTS.md.tmpl` to stay out of Codex's filename-keyed discovery
		// inside this repo (card t925) — appears here as `AGENTS.md`.
		// The two layers are easy to confuse: `harnessFS.Stat` in
		// apply_harness_test.go sees template-SOURCE names and must be given
		// `AGENTS.md.tmpl`, while this listing sees deploy targets.
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

// TestCodexOnlyForceUpdateVariant deploys through the update-path constructor
// (force-update semantics) and re-checks the codex-only file set.
func TestCodexOnlyForceUpdateVariant(t *testing.T) {
	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded: %v", err)
	}
	d, err := NewCodexOnlyDeployerWithRendererAndForceUpdate(cat, NewRenderer(embedded))
	if err != nil {
		t.Fatalf("new codex-only force-update deployer: %v", err)
	}
	rd, ok := d.(ResultDeployer)
	if !ok {
		t.Fatal("force-update codex deployer must satisfy ResultDeployer")
	}

	root := t.TempDir()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	// A template context is REQUIRED, not incidental: the deployer only strips
	// the `.tmpl` suffix on the render branch, which it takes when a renderer
	// AND a context are both present. Passing nil here would deploy the mirror
	// verbatim as `AGENTS.md.tmpl` and the assertion below would report a
	// regression that production does not have — both production deploy sites
	// (initializer.go, mirror_notice.go) pass a context (card t925).
	if err := rd.Deploy(context.Background(), root, mgr, NewTemplateContext()); err != nil {
		t.Fatalf("deploy: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Errorf("AGENTS.md missing from force-update codex deploy: %v", err)
	}
	// ...and the template-source name must NOT survive into the project.
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md.tmpl")); err == nil {
		t.Error("AGENTS.md.tmpl leaked into the deployed project — the .tmpl suffix must be stripped")
	}
	if _, err := os.Stat(filepath.Join(root, ".claude")); err == nil {
		t.Error(".claude/ deployed by the force-update codex deployer")
	}
}

// TestCodexOnlyHiddenPathsErrNotExist pins the Open/Stat hiding contract
// directly: hidden paths answer fs.ErrNotExist through both entry points.
func TestCodexOnlyHiddenPathsErrNotExist(t *testing.T) {
	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded: %v", err)
	}
	d, err := NewCodexOnlyDeployerWithRenderer(cat, NewRenderer(embedded))
	if err != nil {
		t.Fatalf("new deployer: %v", err)
	}
	for _, p := range []string{".claude/skills/moai/SKILL.md", "CLAUDE.md", ".mcp.json", ".claudeignore", ".moai/status_line.sh"} {
		// ExtractTemplate wraps ErrTemplateNotFound — the hidden contract is
		// "reads as absent", whichever error shape the entry point wraps.
		if _, err := d.ExtractTemplate(p); err == nil {
			t.Errorf("hidden path %q readable via ExtractTemplate — not hidden", p)
		} else if !errors.Is(err, ErrTemplateNotFound) && !os.IsNotExist(err) {
			t.Errorf("hidden path %q: unexpected error shape: %v", p, err)
		}
		entries := d.ListTemplates()
		for _, listed := range entries {
			if listed == p {
				t.Errorf("hidden path %q visible in listing", p)
			}
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

func TestHarnessProfilesResolveSharedReferences(t *testing.T) {
	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatal(err)
	}
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	renderer := NewRenderer(embedded)
	profiles := []struct {
		name        string
		newDeployer func() (Deployer, error)
		wantClaude  bool
		wantCodex   bool
	}{
		{"claude", func() (Deployer, error) { return NewClaudeHarnessDeployerWithRenderer(cat, renderer) }, true, false},
		{"gpt", func() (Deployer, error) { return NewCodexOnlyDeployerWithRenderer(cat, renderer) }, false, true},
		{"both", func() (Deployer, error) { return NewDualHarnessDeployerWithRenderer(cat, renderer) }, true, true},
	}
	for _, tc := range profiles {
		t.Run(tc.name, func(t *testing.T) {
			d, err := tc.newDeployer()
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			owned := filepath.Join(root, ".moai", "policies", "README.md")
			if err := os.MkdirAll(filepath.Dir(owned), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(owned, []byte("user policy note\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			mgr := manifest.NewManager()
			if _, err := mgr.Load(root); err != nil {
				t.Fatal(err)
			}
			if err := d.Deploy(context.Background(), root, mgr, NewTemplateContext()); err != nil {
				t.Fatal(err)
			}
			if got, err := os.ReadFile(owned); err != nil || string(got) != "user policy note\n" {
				t.Errorf("user policy note changed: bytes=%q error=%v", got, err)
			}
			for _, rel := range []string{".moai/policies/core/moai-constitution.md", ".moai/workflows/plan.md", "AGENTS.md"} {
				if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
					t.Errorf("required shared reference %s: %v", rel, err)
				}
			}
			if _, err := os.Stat(filepath.Join(root, "CLAUDE.md")); (err == nil) != tc.wantClaude {
				t.Errorf("CLAUDE.md presence = %v, want %v", err == nil, tc.wantClaude)
			}
			if _, err := os.Stat(filepath.Join(root, ".codex", "agents", "moai", "manager-develop.toml")); (err == nil) != tc.wantCodex {
				t.Errorf("Codex agent presence = %v, want %v", err == nil, tc.wantCodex)
			}
			if tc.wantCodex {
				// Only concrete file references are checked. Placeholders and
				// directory examples do not pretend to be required read targets.
				concreteRef := regexp.MustCompile(`(?:\.moai/(?:policies|workflows)|\.agents/skills)/[A-Za-z0-9_./-]+\.(?:md|yaml|json|toml)`)
				checkedRefs := 0
				checkReferences := func(path string, data []byte) {
					for _, line := range strings.Split(string(data), "\n") {
						lower := strings.ToLower(line)
						if !strings.Contains(lower, "read ") && !strings.Contains(lower, "load ") {
							continue
						}
						for _, ref := range concreteRef.FindAllString(line, -1) {
							checkedRefs++
							if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(ref))); err != nil {
								t.Errorf("%s requires absent %s: %v", strings.TrimPrefix(path, root), ref, err)
							}
						}
					}
				}
				body, err := os.ReadFile(filepath.Join(root, ".agents", "skills", "moai", "SKILL.md"))
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(body), ".claude/rules/moai/") || strings.Contains(string(body), ".claude/skills/moai/workflows/") {
					t.Error("Codex dispatcher contains unavailable Claude reference")
				}
				if !strings.Contains(string(body), ".moai/policies/") || !strings.Contains(string(body), ".moai/workflows/") {
					t.Error("Codex dispatcher does not link shared references")
				}
				role, err := os.ReadFile(filepath.Join(root, ".codex", "agents", "moai", "manager-develop.toml"))
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(role), ".claude/rules/moai/") || strings.Contains(string(role), ".claude/skills/") {
					t.Error("Codex role contains unavailable Claude reference")
				}
				err = filepath.WalkDir(filepath.Join(root, ".codex", "agents"), func(path string, entry fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if entry.IsDir() || !strings.HasSuffix(path, ".toml") {
						return nil
					}
					data, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					if strings.Contains(string(data), ".claude/rules/moai/") || strings.Contains(string(data), ".claude/skills/") {
						t.Errorf("Codex role %s contains unavailable Claude reference", filepath.Base(path))
					}
					checkReferences(path, data)
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
				err = filepath.WalkDir(filepath.Join(root, ".agents", "skills"), func(path string, entry fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if entry.IsDir() || !strings.HasSuffix(path, ".md") {
						return nil
					}
					data, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					if strings.Contains(string(data), ".claude/rules/moai/") || strings.Contains(string(data), ".claude/skills/") {
						t.Errorf("Codex skill %s contains unavailable Claude reference", strings.TrimPrefix(path, root))
					}
					checkReferences(path, data)
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
				for _, rel := range []string{".moai/policies", ".moai/workflows"} {
					err := filepath.WalkDir(filepath.Join(root, rel), func(path string, entry fs.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if entry.IsDir() || !strings.HasSuffix(path, ".md") {
							return nil
						}
						data, err := os.ReadFile(path)
						if err != nil {
							return err
						}
						checkReferences(path, data)
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
				}
				t.Logf("checked %d concrete read/load references", checkedRefs)
			}
		})
	}
}
