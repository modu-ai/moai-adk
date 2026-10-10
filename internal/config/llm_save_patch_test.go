package config

// llm_save_patch_test.go — U-05 (REQ-RC30-007): the branches of the llm Save
// path beyond the AC-RC30-06 round-trip. unchanged, nested scalar, upserted
// key, and upserted map entry are RED while the patch path is off: the typed
// rewrite changes bytes and drops comments. The two fallback cases keep the
// typed rewrite, so they are regression guards that pass either way.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// llmSaveFixture is the llm.yaml shape of U-05: a header comment, a trailing
// comment, a key the typed struct does not model, and the GLM tier models.
const llmSaveFixture = "# llm header comment\n" +
	"llm:\n" +
	"  harness: claude\n" +
	"  team_mode: claude  # team note\n" +
	"  deployment_mode: plugin\n" +
	"  future_flag: keep-me\n" +
	"  glm:\n" +
	"    models:\n" +
	"      high: glm-5.2\n"

// writeLLMSaveFixture seeds a fresh project root with an llm.yaml body and
// returns the root and the file path.
func writeLLMSaveFixture(t *testing.T, body string) (root, path string) {
	t.Helper()
	root = t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(sectionsDir, "llm.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, path
}

// saveLLMMutation loads root, applies mutate to the llm section, and saves.
func saveLLMMutation(t *testing.T, root string, mutate func(*LLMConfig)) {
	t.Helper()
	m := NewConfigManager()
	if _, err := m.LoadRaw(root); err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	llm := m.Get().LLM
	mutate(&llm)
	if err := m.SetSection("llm", llm); err != nil {
		t.Fatalf("SetSection(llm): %v", err)
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
}

// reloadLLMSection returns the llm section that a fresh LoadRaw of root reads.
func reloadLLMSection(t *testing.T, root string) LLMConfig {
	t.Helper()
	m := NewConfigManager()
	if _, err := m.LoadRaw(root); err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	return m.Get().LLM
}

func TestLLMSavePatchBranches(t *testing.T) {
	t.Run("unchanged_section_leaves_bytes_alone", func(t *testing.T) {
		t.Parallel()
		root, path := writeLLMSaveFixture(t, llmSaveFixture)
		saveLLMMutation(t, root, func(*LLMConfig) {})
		if got := readFile(t, path); got != llmSaveFixture {
			t.Errorf("a Save with no change rewrote llm.yaml:\n%s", got)
		}
	})

	t.Run("nested_scalar_patched_in_place", func(t *testing.T) {
		t.Parallel()
		root, path := writeLLMSaveFixture(t, llmSaveFixture)
		saveLLMMutation(t, root, func(c *LLMConfig) { c.GLM.Models.High = "glm-5.3" })
		want := strings.Replace(llmSaveFixture, "high: glm-5.2", "high: glm-5.3", 1)
		if got := readFile(t, path); got != want {
			t.Errorf("nested change was not a one-line patch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("upserted_key_keeps_comments_and_unmodelled_keys", func(t *testing.T) {
		t.Parallel()
		root, path := writeLLMSaveFixture(t, llmSaveFixture)
		saveLLMMutation(t, root, func(c *LLMConfig) { c.AgentOverridesConsume = true })
		got := readFile(t, path)
		for _, keep := range []string{"# llm header comment", "# team note", "deployment_mode: plugin", "future_flag: keep-me", "agent_overrides_consume: true"} {
			if !strings.Contains(got, keep) {
				t.Errorf("upsert lost %q:\n%s", keep, got)
			}
		}
		if !reloadLLMSection(t, root).AgentOverridesConsume {
			t.Error("upserted key did not round-trip as true")
		}
	})

	t.Run("upserted_map_entry_keeps_comments_and_unmodelled_keys", func(t *testing.T) {
		t.Parallel()
		root, path := writeLLMSaveFixture(t, llmSaveFixture)
		saveLLMMutation(t, root, func(c *LLMConfig) {
			c.AgentOverrides = map[string]ModelEffort{"manager-develop": {Model: "opus", Effort: "high"}}
		})
		got := readFile(t, path)
		for _, keep := range []string{"# llm header comment", "# team note", "deployment_mode: plugin", "future_flag: keep-me", "manager-develop:"} {
			if !strings.Contains(got, keep) {
				t.Errorf("map upsert lost %q:\n%s", keep, got)
			}
		}
		if got := reloadLLMSection(t, root).AgentOverrides["manager-develop"]; got != (ModelEffort{Model: "opus", Effort: "high"}) {
			t.Errorf("upserted map entry did not round-trip: %+v", got)
		}
	})

	t.Run("quoted_value_falls_back_and_round_trips", func(t *testing.T) {
		t.Parallel()
		root, path := writeLLMSaveFixture(t, llmSaveFixture+"  profile: high\n")
		saveLLMMutation(t, root, func(c *LLMConfig) { c.Profile = "" })
		if strings.Contains(readFile(t, path), "profile: high") {
			t.Error("cleared profile survived the Save")
		}
		if got := reloadLLMSection(t, root).Profile; got != "" {
			t.Errorf("profile after fallback Save = %q, want empty", got)
		}
	})

	t.Run("removed_key_falls_back_and_round_trips", func(t *testing.T) {
		t.Parallel()
		root, path := writeLLMSaveFixture(t, llmSaveFixture+"  claude_bin: /opt/claude\n")
		saveLLMMutation(t, root, func(c *LLMConfig) { c.ClaudeBin = "" })
		if strings.Contains(readFile(t, path), "claude_bin") {
			t.Error("cleared claude_bin survived the Save")
		}
		if got := reloadLLMSection(t, root).ClaudeBin; got != "" {
			t.Errorf("claude_bin after fallback Save = %q, want empty", got)
		}
	})
}
