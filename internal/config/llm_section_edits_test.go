package config

// llm_section_edits_test.go — the edits LLMSectionEdits hands the yamlpatch seam
// (REQ-RC30-007): a removed key becomes a removal, a quoted empty value a quoted
// replacement, a file that does not exist yet gets an upsert for every scalar,
// and a change the seam cannot express is an error that names the path.

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// findLLMEdit returns the edit whose path is exactly path, and whether one exists.
func findLLMEdit(edits []yamlpatch.KeyEdit, path ...string) (yamlpatch.KeyEdit, bool) {
	for _, e := range edits {
		if len(e.Path) != len(path) {
			continue
		}
		match := true
		for i := range path {
			if e.Path[i] != path[i] {
				match = false
				break
			}
		}
		if match {
			return e, true
		}
	}
	return yamlpatch.KeyEdit{}, false
}

func TestLLMSectionEditsRemovalAndQuotedValue(t *testing.T) {
	t.Parallel()
	persisted := LLMConfig{TeamMode: "glm", ClaudeBin: "/opt/claude"}
	edits, err := LLMSectionEdits(&persisted, LLMConfig{})
	if err != nil {
		t.Fatalf("LLMSectionEdits: %v", err)
	}
	if len(edits) != 2 {
		t.Fatalf("got %d edits, want 2 (team_mode replaced, claude_bin removed): %+v", len(edits), edits)
	}
	teamMode, ok := findLLMEdit(edits, "llm", "team_mode")
	if !ok || teamMode.Value != "" || !teamMode.Quoted || teamMode.Remove {
		t.Errorf("team_mode edit = %+v (found %v), want a quoted empty replacement", teamMode, ok)
	}
	claudeBin, ok := findLLMEdit(edits, "llm", "claude_bin")
	if !ok || !claudeBin.Remove {
		t.Errorf("claude_bin edit = %+v (found %v), want a removal", claudeBin, ok)
	}
}

func TestLLMSectionEditsNilPersistedUpsertsEveryScalar(t *testing.T) {
	t.Parallel()
	edits, err := LLMSectionEdits(nil, LLMConfig{TeamMode: "glm"})
	if err != nil {
		t.Fatalf("LLMSectionEdits: %v", err)
	}
	if teamMode, ok := findLLMEdit(edits, "llm", "team_mode"); !ok || teamMode.Value != "glm" || teamMode.Quoted {
		t.Errorf("team_mode upsert = %+v (found %v), want plain glm", teamMode, ok)
	}
	if mode, ok := findLLMEdit(edits, "llm", "mode"); !ok || mode.Value != "" || !mode.Quoted {
		t.Errorf("mode upsert = %+v (found %v), want a quoted empty value", mode, ok)
	}
	for _, e := range edits {
		if e.Remove {
			t.Errorf("a new file has nothing to remove, got %+v", e)
		}
	}
}

func TestLLMSectionEditsUnchangedSectionIsEmpty(t *testing.T) {
	t.Parallel()
	section := LLMConfig{TeamMode: "glm", ClaudeBin: "/opt/claude", GLM: GLMSettings{BaseURL: "https://example.test"}}
	edits, err := LLMSectionEdits(&section, section)
	if err != nil {
		t.Fatalf("LLMSectionEdits: %v", err)
	}
	if len(edits) != 0 {
		t.Errorf("an unchanged section produced edits: %+v", edits)
	}
}

// TestDiffLLMNodesReportsUnexpressedChange pins the error path: a sequence that
// changed has no scalar edit, so the diff names its path rather than guessing.
func TestDiffLLMNodesReportsUnexpressedChange(t *testing.T) {
	t.Parallel()
	var base, now yaml.Node
	if err := yaml.Unmarshal([]byte("a: [x]\n"), &base); err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal([]byte("a: [y]\n"), &now); err != nil {
		t.Fatal(err)
	}
	var edits []yamlpatch.KeyEdit
	err := diffLLMNodes([]string{"llm"}, base.Content[0], now.Content[0], &edits)
	if err == nil {
		t.Fatal("a changed sequence must be an error, got nil")
	}
	if want := "llm.a: change cannot be written through the yamlpatch seam"; err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}
