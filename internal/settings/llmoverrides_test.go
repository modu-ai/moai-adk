package settings

// llmoverrides_test.go — SPEC-WEB-AGENTFM-RESTORE-001 M3: unit tests for the
// llm.profile / llm.agent_overrides write seams (REQ-AFR-003/004).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
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
