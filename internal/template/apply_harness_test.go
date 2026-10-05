package template

// SPEC-INIT-HARNESS-001 — ApplyHarness unit tests (the llm.yaml patch write
// path behind init/update persistence, REQ-IH-002/010).

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeHarnessFixture lays a llm.yaml with the given harness line into a temp
// project and returns its path.
func writeHarnessFixture(t *testing.T, harnessLine string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "llm:\n" + harnessLine + "  glm_env_var: \"GLM_API_KEY\"\n"
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestApplyHarness_ReplacesQuotedValue covers the plain replace path: a
// quoted template value is rewritten to the new unquoted value.
func TestApplyHarness_ReplacesQuotedValue(t *testing.T) {
	root := writeHarnessFixture(t, "  harness: \"claude\"\n")
	if err := ApplyHarness(root, "gpt"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "harness: gpt") {
		t.Errorf("harness value not replaced:\n%s", data)
	}
	if strings.Contains(string(data), "harness: \"claude\"") {
		t.Errorf("old value survived:\n%s", data)
	}
}

// TestApplyHarness_NoOpWhenAlreadyEqual pins the byte-identity contract the
// update first-deploy test found: an already-correct value — QUOTED style
// included — must be left byte-identical, not re-styled.
func TestApplyHarness_NoOpWhenAlreadyEqual(t *testing.T) {
	for _, line := range []string{"  harness: \"claude\"\n", "  harness: claude\n"} {
		root := writeHarnessFixture(t, line)
		before, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		if err := ApplyHarness(root, "claude"); err != nil {
			t.Fatalf("apply(%q): %v", line, err)
		}
		after, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		if string(before) != string(after) {
			t.Errorf("no-op violated for %q:\nbefore:\n%s\nafter:\n%s", line, before, after)
		}
	}
}

// TestApplyHarness_InsertsWhenAbsent covers the legacy-config insertion path.
func TestApplyHarness_InsertsWhenAbsent(t *testing.T) {
	root := writeHarnessFixture(t, "  glm_env_var: \"GLM_API_KEY\"\n")
	if err := ApplyHarness(root, "both"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "harness: both") {
		t.Errorf("harness key not inserted:\n%s", data)
	}
}

// TestApplyHarness_RejectsOutOfSet covers the closed-set guard.
func TestApplyHarness_RejectsOutOfSet(t *testing.T) {
	root := writeHarnessFixture(t, "  harness: \"claude\"\n")
	if err := ApplyHarness(root, "gemini"); err == nil {
		t.Error("out-of-set harness value must error, not write")
	}
	if err := ApplyHarness(root, ""); err == nil {
		t.Error("empty harness value must error, not write")
	}
}

// TestApplyHarness_AbsentFileIsNoOp covers the graceful missing-file branch.
func TestApplyHarness_AbsentFileIsNoOp(t *testing.T) {
	if err := ApplyHarness(t.TempDir(), "gpt"); err != nil {
		t.Errorf("absent llm.yaml must be a graceful no-op, got: %v", err)
	}
}

// TestHarnessFSStat covers the wrapper's Stat path: a remapped catalog path
// stats through to the catalog root, and a hidden path answers fs.ErrNotExist
// (the deployer itself walks via ReadDir, so Stat needs its own pin).
func TestHarnessFSStat(t *testing.T) {
	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	base, err := SlimFS(embeddedRaw, cat)
	if err != nil {
		t.Fatalf("slim base: %v", err)
	}
	catalogRoot, err := fs.Sub(embeddedRaw, "templates/"+CanonicalSkillsRelDir)
	if err != nil {
		t.Fatalf("catalog root: %v", err)
	}
	h, err := newHarnessFS(base, catalogRoot)
	if err != nil {
		t.Fatalf("new harnessFS: %v", err)
	}

	if _, statErr := h.Stat(".agents/skills/moai-workflow-tdd/SKILL.md"); statErr != nil {
		t.Errorf("remapped catalog path Stat failed: %v", statErr)
	}
	if _, statErr := h.Stat("CLAUDE.md"); !errors.Is(statErr, fs.ErrNotExist) {
		t.Errorf("hidden path Stat = %v, want fs.ErrNotExist", statErr)
	}
	if _, statErr := h.Stat(".claude/skills"); !errors.Is(statErr, fs.ErrNotExist) {
		t.Errorf("hidden dir Stat = %v, want fs.ErrNotExist", statErr)
	}
	// The contract mirror ships under its template-source name `AGENTS.md.tmpl`
	// (card t925 — the suffix keeps it out of Codex's filename-keyed discovery
	// in this repo; the deployer strips it so a user project gets `AGENTS.md`).
	// This wrapper sees template-source names, so that is the name to stat.
	if _, statErr := h.Stat("AGENTS.md.tmpl"); statErr != nil {
		t.Errorf("underlying path Stat failed: %v", statErr)
	}
}
