package config

// llm_save_branches_test.go — card t1619 (SPEC-UPDATE-REPAIR-RC30-001 U-05,
// REQ-RC30-007): the error branches of the llm Save path that
// llm_save_patch_test.go and llm_save_seam_test.go do not reach. Each test
// names the llm_save.go or protected_zone.go lines it exercises.
//
// The baseline is injected through m.llmBaseline. Save diffs against the typed
// rendering taken at load (resetSectionDirtyLocked in manager.go), so a
// malformed or differently shaped baseline is reachable only from a
// package-internal test.

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// saveLLMExpectingError loads root, applies mutate to the llm section, replaces
// the in-memory llm baseline when baseline is non-nil, and returns the error
// that Save reports.
func saveLLMExpectingError(t *testing.T, root string, baseline []byte, mutate func(*LLMConfig)) error {
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
	if baseline != nil {
		m.llmBaseline = baseline
	}
	return m.Save()
}

// TestSaveLLMRejectsMultilineScalarChange covers the scalar refusal of a
// changed value with no yamlpatch form: a multi-line string is rendered as a
// block scalar, so llmQuoting refuses it, and the refusal is returned through
// the edit builder and Save (llm_save.go lines 37, 85, 142, 188). Save names
// the field and leaves llm.yaml unchanged.
func TestSaveLLMRejectsMultilineScalarChange(t *testing.T) {
	root, path := writeLLMSaveFixture(t, llmSaveFixture)
	err := saveLLMExpectingError(t, root, nil, func(l *LLMConfig) {
		l.GLMEnvVar = "FIRST_LINE\nSECOND_LINE"
	})
	if err == nil {
		t.Fatal("Save accepted a multi-line value that yamlpatch cannot express")
	}
	if !strings.Contains(err.Error(), "llm.glm_env_var") {
		t.Errorf("refusal does not name the field: %v", err)
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != llmSaveFixture {
		t.Errorf("llm.yaml changed after a refused save:\n%s", got)
	}
}

// TestSaveLLMRejectsUnparseableBaseline covers the YAML decode refusal of the
// baseline (llm_save.go lines 33 and 95): the baseline cannot be parsed, so
// Save refuses the patch rather than diffing against unknown content.
func TestSaveLLMRejectsUnparseableBaseline(t *testing.T) {
	root, _ := writeLLMSaveFixture(t, llmSaveFixture)
	err := saveLLMExpectingError(t, root, []byte("llm: [unclosed\n"), func(l *LLMConfig) {
		l.Harness = "codex"
	})
	if err == nil || !strings.Contains(err.Error(), "llm baseline") {
		t.Fatalf("Save err = %v, want an llm baseline refusal", err)
	}
}

// TestSaveLLMRejectsBaselineWithoutMapping covers the two structure refusals of
// the baseline (llm_save.go lines 98 and 102): a document whose root is not a
// mapping, and a mapping that carries no llm section.
func TestSaveLLMRejectsBaselineWithoutMapping(t *testing.T) {
	cases := []struct {
		name     string
		baseline string
		want     string
	}{
		{"root is a sequence", "- item\n", "no top-level mapping"},
		{"no llm mapping", "other: 1\n", "no llm mapping"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root, _ := writeLLMSaveFixture(t, llmSaveFixture)
			err := saveLLMExpectingError(t, root, []byte(c.baseline), func(l *LLMConfig) {
				l.Harness = "codex"
			})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Save err = %v, want a refusal containing %q", err, c.want)
			}
		})
	}
}

// TestSaveLLMRejectsKindChangeUnderSection covers the container-kind refusal of
// diffLLMNodes (llm_save.go line 115): the baseline holds a scalar at llm.glm
// where the typed section now renders a mapping, which no scalar patch can
// express. The refusal names the path.
func TestSaveLLMRejectsKindChangeUnderSection(t *testing.T) {
	root, _ := writeLLMSaveFixture(t, llmSaveFixture)
	err := saveLLMExpectingError(t, root, []byte("llm:\n  glm: scalar\n"), func(l *LLMConfig) {
		l.Harness = "codex"
	})
	if err == nil || !strings.Contains(err.Error(), "llm.glm") {
		t.Fatalf("Save err = %v, want a refusal naming llm.glm", err)
	}
}

// TestSaveLLMRejectsMultilineValueUnderNewKey covers the upsert path: a key the
// baseline lacks is appended scalar by scalar, and a block scalar among those
// leaves is refused. The refusal is carried back through each enclosing level
// (llm_save.go lines 132, 162, and 169).
func TestSaveLLMRejectsMultilineValueUnderNewKey(t *testing.T) {
	root, _ := writeLLMSaveFixture(t, llmSaveFixture)
	err := saveLLMExpectingError(t, root, []byte("llm:\n  harness: claude\n"), func(l *LLMConfig) {
		l.GLM.Models.High = "glm-5.2\nglm-5.3"
	})
	if err == nil || !strings.Contains(err.Error(), "llm.glm.models.high") {
		t.Fatalf("Save err = %v, want a refusal naming llm.glm.models.high", err)
	}
}

// TestDiffLLMNodesAcceptsEqualSequence covers the equal-container branch of
// diffLLMNodes (llm_save.go line 148): a sequence that is unchanged yields no
// edit and no error. The typed rendering never emits a sequence, so the input
// is a direct yaml.Node pair, as in TestDiffLLMNodesReportsUnexpressedChange.
func TestDiffLLMNodesAcceptsEqualSequence(t *testing.T) {
	var base, now yaml.Node
	if err := yaml.Unmarshal([]byte("a: [x]\n"), &base); err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal([]byte("a: [x]\n"), &now); err != nil {
		t.Fatal(err)
	}
	var edits []yamlpatch.KeyEdit
	if err := diffLLMNodes([]string{"llm"}, base.Content[0], now.Content[0], &edits); err != nil {
		t.Fatalf("an unchanged sequence must not be an error: %v", err)
	}
	if len(edits) != 0 {
		t.Errorf("an unchanged sequence produced edits: %+v", edits)
	}
}

// TestAppendLLMLeavesRefusesSequence covers the default branch of
// appendLLMLeaves (llm_save.go line 174): a key the baseline lacks, whose value
// is a sequence, has no scalar upsert and is refused with its path.
func TestAppendLLMLeavesRefusesSequence(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("[x]\n"), &doc); err != nil {
		t.Fatal(err)
	}
	var edits []yamlpatch.KeyEdit
	err := appendLLMLeaves([]string{"llm", "a"}, doc.Content[0], &edits)
	if err == nil {
		t.Fatal("a sequence under an absent key must be an error, got nil")
	}
	if want := "llm.a: change cannot be written through the yamlpatch seam"; err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// TestParseZoneEntryRejectsWildcardInsideUserRoot covers the wildcard refusal in
// the user-root branch (protected_zone.go line 335): a glob is legal only as the
// trailing * of an entry, never inside its path.
func TestParseZoneEntryRejectsWildcardInsideUserRoot(t *testing.T) {
	for _, raw := range []string{"user-root:claude-agents/a*b", "user-root:claude-agents/a?b"} {
		_, err := parseZoneEntry(raw)
		if err == nil || !strings.Contains(err.Error(), "wildcard is legal only") {
			t.Errorf("parseZoneEntry(%q) err = %v, want the wildcard refusal", raw, err)
		}
	}
}
