package cli

// glm_legacy_alias_test.go — the scope of the DR-2 alias removal in saveLLMSection
// (REQ-RC30-007 named exception). A real write removes the legacy alias keys only
// at llm.glm.models; a same-named key under another parent survives.

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// TestLegacyLLMAliasKeysAreExactlyOpusSonnetHaiku pins the named slice that the
// writer removes. It holds exactly opus, sonnet, and haiku, and the removals built
// from it target llm.glm.models and nothing else.
func TestLegacyLLMAliasKeysAreExactlyOpusSonnetHaiku(t *testing.T) {
	t.Parallel()
	if want := []string{"opus", "sonnet", "haiku"}; !slices.Equal(legacyLLMAliasKeys, want) {
		t.Fatalf("legacyLLMAliasKeys = %v, want exactly %v", legacyLLMAliasKeys, want)
	}
	edits := legacyLLMAliasEdits()
	if len(edits) != 3 {
		t.Fatalf("legacyLLMAliasEdits returned %d edits, want 3", len(edits))
	}
	for i, key := range []string{"opus", "sonnet", "haiku"} {
		if !edits[i].Remove || !slices.Equal(edits[i].Path, []string{"llm", "glm", "models", key}) {
			t.Errorf("edit %d = %+v, want a removal at llm.glm.models.%s", i, edits[i], key)
		}
	}
}

// TestSaveLLMSectionRemovesOnlyTheLegacyAliasKeys runs the writer on an llm.yaml
// that carries the three legacy alias keys at llm.glm.models, plus same-named keys
// elsewhere. The three alias keys are gone. The nested same-named key, the
// document-root key, deployment_mode, future_flag, the models the writer keeps, and
// the header comment all survive.
func TestSaveLLMSectionRemovesOnlyTheLegacyAliasKeys(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sectionsDir, "llm.yaml")
	src := "# llm header kept\n" +
		"opus: root-keep\n" +
		"llm:\n" +
		"  team_mode: cg\n" +
		"  deployment_mode: plugin\n" +
		"  future_flag: keep-me\n" +
		"  custom:\n" +
		"    opus: nested-keep\n" +
		"  glm:\n" +
		"    models:\n" +
		"      high: glm-5.3-flash\n" +
		"      medium: glm-5.3-flash\n" +
		"      low: glm-5.3-flash\n" +
		"      fable: glm-5.3\n" +
		"      opus: glm-5.2\n" +
		"      sonnet: glm-5.2\n" +
		"      haiku: glm-5.2\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := persistTeamMode(root, "glm"); err != nil {
		t.Fatalf("persistTeamMode: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var tree map[string]any
	if err := yaml.Unmarshal(raw, &tree); err != nil {
		t.Fatalf("parse rewritten llm.yaml: %v\n%s", err, raw)
	}
	llm, _ := tree["llm"].(map[string]any)
	custom, _ := llm["custom"].(map[string]any)
	glm, _ := llm["glm"].(map[string]any)
	models, _ := glm["models"].(map[string]any)
	for _, alias := range []string{"opus", "sonnet", "haiku"} {
		if _, ok := models[alias]; ok {
			t.Errorf("legacy alias %q survived at llm.glm.models:\n%s", alias, raw)
		}
	}
	if keys := slices.Sorted(maps.Keys(models)); !slices.Equal(keys, []string{"fable", "high", "low", "medium"}) {
		t.Errorf("llm.glm.models keys = %v, want exactly [fable high low medium]", keys)
	}
	if custom["opus"] != "nested-keep" {
		t.Errorf("same-named key under llm.custom did not survive: %v\n%s", custom["opus"], raw)
	}
	if tree["opus"] != "root-keep" {
		t.Errorf("document-root key opus changed: %v\n%s", tree["opus"], raw)
	}
	if llm["deployment_mode"] != "plugin" || llm["future_flag"] != "keep-me" {
		t.Errorf("unknown keys did not survive: deployment_mode=%v future_flag=%v\n%s", llm["deployment_mode"], llm["future_flag"], raw)
	}
	if llm["team_mode"] != "glm" {
		t.Errorf("team_mode = %v, want glm", llm["team_mode"])
	}
	if !strings.Contains(string(raw), "# llm header kept") {
		t.Errorf("header comment lost:\n%s", raw)
	}
}
