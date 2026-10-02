package settings

// llmoverrides_test.go — SPEC-WEB-AGENTFM-RESTORE-001 M3: unit tests for the
// llm.profile / llm.agent_overrides write seams (REQ-AFR-003/004).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"gopkg.in/yaml.v3"
)

const llmOverridesFixture = `llm:
  mode: ""
  # a comment the write must preserve
  harness: "claude"
  glm:
    base_url: x
`

func TestWriteLLMProfileSplicesAndGates(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := llmOverridesFixture + "  profile: \"\"\n  agent_overrides: {}\n"
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteLLMProfile(root, "max"); err != nil {
		t.Fatalf("WriteLLMProfile: %v", err)
	}
	got := readFileAt(t, filepath.Join(dir, "llm.yaml"))
	if !strings.Contains(got, `profile: "max"`) {
		t.Errorf("profile not spliced:\n%s", got)
	}
	if !strings.Contains(got, "# a comment the write must preserve") {
		t.Errorf("comment lost by the profile splice:\n%s", got)
	}
	if strings.Contains(got, "performance_tier") {
		t.Errorf("the retired alias must never be written:\n%s", got)
	}

	// No-op gate: an equal value leaves the file byte-identical (mtime too).
	info, _ := os.Stat(filepath.Join(dir, "llm.yaml"))
	if err := WriteLLMProfile(root, "max"); err != nil {
		t.Fatal(err)
	}
	info2, _ := os.Stat(filepath.Join(dir, "llm.yaml"))
	if !info.ModTime().Equal(info2.ModTime()) {
		t.Error("an equal profile submission rewrote the file (mtime moved) — the no-op gate failed")
	}
}

func TestWriteLLMAgentOverridesBlockSplice(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := llmOverridesFixture + "  profile: \"\"\n  agent_overrides: {}\n"
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	ov := map[string]config.ModelEffort{
		"manager-develop": {Model: "opus", Effort: "xhigh"},
		"manager-todo":    {Model: "haiku", Effort: "low"},
	}
	if err := WriteLLMAgentOverrides(root, ov); err != nil {
		t.Fatalf("WriteLLMAgentOverrides: %v", err)
	}
	got := readFileAt(t, filepath.Join(dir, "llm.yaml"))
	for _, want := range []string{
		"# a comment the write must preserve",
		"agent_overrides:",
		"    manager-develop:",
		"        model: opus",
		"        effort: xhigh",
		"    manager-todo:",
		"        model: haiku",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("block write lost %q:\n%s", want, got)
		}
	}

	// Byte-identical no-op: the same map again must not touch the file.
	info, _ := os.Stat(filepath.Join(dir, "llm.yaml"))
	if err := WriteLLMAgentOverrides(root, ov); err != nil {
		t.Fatal(err)
	}
	info2, _ := os.Stat(filepath.Join(dir, "llm.yaml"))
	if !info.ModTime().Equal(info2.ModTime()) {
		t.Error("an equal override map rewrote the file — the no-op gate failed")
	}

	// Clear: passing the map minus one entry removes that entry and keeps the
	// sibling.
	if err := WriteLLMAgentOverrides(root, map[string]config.ModelEffort{
		"manager-todo": {Model: "haiku", Effort: "low"},
	}); err != nil {
		t.Fatal(err)
	}
	got = readFileAt(t, filepath.Join(dir, "llm.yaml"))
	if strings.Contains(got, "manager-develop") {
		t.Errorf("cleared entry survived:\n%s", got)
	}
	if !strings.Contains(got, "manager-todo") {
		t.Errorf("sibling override lost on clear:\n%s", got)
	}

	// All clear: the block collapses to the template's flow-map shape.
	if err := WriteLLMAgentOverrides(root, nil); err != nil {
		t.Fatal(err)
	}
	got = readFileAt(t, filepath.Join(dir, "llm.yaml"))
	if !strings.Contains(got, "agent_overrides: {}") {
		t.Errorf("empty map must collapse the block to {}: %s", got)
	}
	if !strings.Contains(got, "# a comment the write must preserve") {
		t.Errorf("comment lost by the block rewrite:\n%s", got)
	}
}

func TestWriteLLMAgentOverridesAbsentKeyAndFile(t *testing.T) {
	t.Run("absent key inserts under the root", func(t *testing.T) {
		root := t.TempDir()
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(llmOverridesFixture), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := WriteLLMAgentOverrides(root, map[string]config.ModelEffort{
			"manager-git": {Model: "sonnet", Effort: "low"},
		}); err != nil {
			t.Fatal(err)
		}
		got := readFileAt(t, filepath.Join(dir, "llm.yaml"))
		if !strings.Contains(got, "agent_overrides:") || !strings.Contains(got, "manager-git") {
			t.Errorf("absent key not inserted:\n%s", got)
		}
		if !strings.Contains(got, "# a comment the write must preserve") {
			t.Errorf("comment lost:\n%s", got)
		}
	})
	t.Run("absent file creates a minimal document", func(t *testing.T) {
		root := t.TempDir()
		if err := WriteLLMAgentOverrides(root, map[string]config.ModelEffort{
			"manager-git": {Model: "sonnet", Effort: "low"},
		}); err != nil {
			t.Fatal(err)
		}
		got := readFileAt(t, filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
		if !strings.Contains(got, "llm:") || !strings.Contains(got, "manager-git") {
			t.Errorf("greenfield write malformed:\n%s", got)
		}
	})
}

func readFileAt(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestSnapshotRestoreLLMYAML — F3: the two-step write pair's rollback
// primitives (sync-audit card t1411). A snapshot of an existing file restores
// byte-identically; a greenfield snapshot (existed=false) restores by
// removing the created file.
func TestSnapshotRestoreLLMYAML(t *testing.T) {
	t.Run("existing file restores byte-identically", func(t *testing.T) {
		root := t.TempDir()
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		seed := "llm:\n  profile: \"max\"\n"
		if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(seed), 0o644); err != nil {
			t.Fatal(err)
		}

		snap, existed, err := SnapshotLLMYAML(root)
		if err != nil || !existed {
			t.Fatalf("SnapshotLLMYAML = (%d bytes, existed=%v, %v)", len(snap), existed, err)
		}
		if string(snap) != seed {
			t.Fatalf("snapshot bytes = %q, want the seed", snap)
		}

		// A later write mutates the file; the restore undoes it.
		if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte("llm:\n  profile: \"high\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := RestoreLLMYAML(root, snap, existed); err != nil {
			t.Fatalf("RestoreLLMYAML: %v", err)
		}
		if got := readFileAt(t, filepath.Join(dir, "llm.yaml")); got != seed {
			t.Errorf("restore = %q, want the seed byte-identical", got)
		}
	})
	t.Run("greenfield snapshot removes the created file on restore", func(t *testing.T) {
		root := t.TempDir()
		snap, existed, err := SnapshotLLMYAML(root)
		if err != nil || existed {
			t.Fatalf("greenfield SnapshotLLMYAML = (existed=%v, %v); want false, nil", existed, err)
		}
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte("llm:\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := RestoreLLMYAML(root, snap, existed); err != nil {
			t.Fatalf("RestoreLLMYAML: %v", err)
		}
		if _, statErr := os.Stat(filepath.Join(dir, "llm.yaml")); !os.IsNotExist(statErr) {
			t.Error("the created llm.yaml survived the greenfield restore")
		}
	})
}

// ── F1 regression tests (sync-audit verdict, card t1411) ────────────────────
//
// The splice's key-search and body-scope loops must treat blank and comment
// lines as PART of the llm block: only a zero-indent CONTENT line terminates
// it. The shipped template llm.yaml carries blank lines between child keys
// (lines 4/12/14/24), so a loop that breaks on the first zero-indent line —
// blank included — stops before the agent_overrides key and takes the
// absent-key insertion path, writing a DUPLICATE key that makes the document
// unparseable (yaml.v3: mapping key "agent_overrides" already defined).

// realTemplateLLMYAML reads the shipped template llm.yaml bytes — the real
// deployment shape a fresh project's console save hits first.
func realTemplateLLMYAML(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "template", "templates", ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatalf("read shipped template llm.yaml: %v", err)
	}
	// Precondition: the deployment shape carries blank lines INSIDE the llm
	// block and the agent_overrides key — otherwise this fixture measures
	// nothing (verification-completeness §1.1).
	s := string(b)
	blankInside := false
	sawKey := false
	for _, l := range strings.Split(s, "\n") {
		if strings.HasPrefix(l, "  agent_overrides:") {
			sawKey = true
		}
		if l == "" {
			blankInside = true
		}
	}
	if !sawKey || !blankInside {
		t.Fatalf("template fixture drifted: agent_overrides key present=%v, blank line present=%v — the F1 shape is gone, rewrite this fixture", sawKey, blankInside)
	}
	return b
}

// parseLLMDoc parses doc and returns the llm root mapping — a duplicate
// agent_overrides key makes yaml.v3 Unmarshal FAIL, which is exactly the
// disk corruption F1 produced.
func parseLLMDoc(t *testing.T, doc string) map[string]any {
	t.Helper()
	var top map[string]any
	if err := yaml.Unmarshal([]byte(doc), &top); err != nil {
		t.Fatalf("spliced llm.yaml does not parse (duplicate key = the F1 corruption): %v\ndoc:\n%s", err, doc)
	}
	llm, _ := top["llm"].(map[string]any)
	if llm == nil {
		t.Fatalf("spliced llm.yaml has no llm root:\n%s", doc)
	}
	return llm
}

func TestWriteLLMAgentOverridesRealTemplateShape(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), realTemplateLLMYAML(t), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteLLMAgentOverrides(root, map[string]config.ModelEffort{
		"manager-git": {Model: "sonnet", Effort: "low"},
	}); err != nil {
		t.Fatalf("WriteLLMAgentOverrides: %v", err)
	}

	got := readFileAt(t, filepath.Join(dir, "llm.yaml"))
	llm := parseLLMDoc(t, got)
	ov, _ := llm["agent_overrides"].(map[string]any)
	git, _ := ov["manager-git"].(map[string]any)
	if git == nil || git["model"] != "sonnet" || git["effort"] != "low" {
		t.Errorf("override did not land on the real-template shape: %v", ov)
	}
	if _, still := llm["profile"]; !still {
		t.Errorf("the profile sibling key was lost by the splice:\n%s", got)
	}
}

func TestWriteLLMAgentOverridesBlankLineInsideBlock(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "llm:\n  mode: \"\"\n\n  profile: \"\"\n\n  agent_overrides: {}\n\n  glm:\n    base_url: x\n"
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteLLMAgentOverrides(root, map[string]config.ModelEffort{
		"manager-todo": {Model: "haiku", Effort: "low"},
	}); err != nil {
		t.Fatalf("WriteLLMAgentOverrides: %v", err)
	}

	got := readFileAt(t, filepath.Join(dir, "llm.yaml"))
	llm := parseLLMDoc(t, got)
	ov, _ := llm["agent_overrides"].(map[string]any)
	if len(ov) != 1 {
		t.Errorf("agent_overrides entries = %v, want exactly the one pin", ov)
	}
	if _, ok := llm["glm"]; !ok {
		t.Errorf("the glm sibling key was lost:\n%s", got)
	}
}

func TestWriteLLMAgentOverridesColumn0CommentInsideBlock(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "llm:\n  mode: \"\"\n# a stray comment at column zero — comments never end a YAML block\n  profile: \"\"\n  agent_overrides: {}\n  glm:\n    base_url: x\n"
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteLLMAgentOverrides(root, map[string]config.ModelEffort{
		"manager-git": {Model: "sonnet", Effort: "low"},
	}); err != nil {
		t.Fatalf("WriteLLMAgentOverrides: %v", err)
	}

	got := readFileAt(t, filepath.Join(dir, "llm.yaml"))
	llm := parseLLMDoc(t, got)
	ov, _ := llm["agent_overrides"].(map[string]any)
	if _, ok := ov["manager-git"]; !ok {
		t.Errorf("override missing after splice past a column-0 comment:\n%s", got)
	}
	if !strings.Contains(got, "# a stray comment at column zero") {
		t.Errorf("the column-0 comment was dropped:\n%s", got)
	}
}

func TestWriteLLMAgentOverridesEndToEndTemplateFile(t *testing.T) {
	// The verdict's arm (d): the PUBLIC seam end-to-end on the real template
	// bytes — after the write, the DISK file must parse with the pin landed.
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), realTemplateLLMYAML(t), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteLLMAgentOverrides(root, map[string]config.ModelEffort{
		"manager-develop": {Model: "opus", Effort: "xhigh"},
	}); err != nil {
		t.Fatalf("WriteLLMAgentOverrides: %v", err)
	}

	got := readFileAt(t, filepath.Join(dir, "llm.yaml"))
	llm := parseLLMDoc(t, got)
	ov, _ := llm["agent_overrides"].(map[string]any)
	dev, _ := ov["manager-develop"].(map[string]any)
	if dev == nil || dev["model"] != "opus" || dev["effort"] != "xhigh" {
		t.Errorf("the end-to-end write did not land the pin on the real-template shape: %v", ov)
	}
}
