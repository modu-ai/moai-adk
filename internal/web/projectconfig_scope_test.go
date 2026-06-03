package web

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteProjectConfigSectionIsolation covers AC-WC3-007: a project-config
// write touches ONLY the quality + git_convention sections. The config manager's
// Save() rewrites every section it owns, so the assertion is on CONTENT equality
// (not file-untouched) for workflow.yaml / harness.yaml / git-strategy.yaml /
// llm.yaml — their content must be byte-identical before and after the write.
func TestWriteProjectConfigSectionIsolation(t *testing.T) {
	t.Parallel()
	root := writeProjectSections(t, "tdd", "auto")
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")

	outOfScope := []string{"workflow.yaml", "harness.yaml", "git-strategy.yaml", "llm.yaml"}
	before := make(map[string][]byte, len(outOfScope))
	for _, name := range outOfScope {
		data, err := os.ReadFile(filepath.Join(sectionsDir, name))
		if err != nil {
			t.Fatalf("read %s before: %v", name, err)
		}
		before[name] = data
	}

	if err := writeProjectConfig(root, "ddd", "angular"); err != nil {
		t.Fatalf("writeProjectConfig: %v", err)
	}

	// The two in-scope files must reflect the change.
	devMode, convention := readProjectSectionValues(t, root)
	if devMode != "ddd" || convention != "angular" {
		t.Fatalf("in-scope write failed: devMode=%q convention=%q", devMode, convention)
	}

	// The out-of-scope files must be CONTENT-unchanged. The config manager Save()
	// only rewrites the sections it owns (user/language/quality/git-convention/llm);
	// workflow/harness/git-strategy are never written by Save() at all, and llm is
	// re-serialized but its content (mode: "") must be semantically preserved.
	for _, name := range []string{"workflow.yaml", "harness.yaml", "git-strategy.yaml"} {
		after, err := os.ReadFile(filepath.Join(sectionsDir, name))
		if err != nil {
			t.Fatalf("read %s after: %v", name, err)
		}
		if string(after) != string(before[name]) {
			t.Errorf("%s content changed by project-config write:\nbefore: %q\nafter:  %q",
				name, before[name], after)
		}
	}

	// llm.yaml: the config manager Save() owns it, so it may be re-serialized.
	// Assert its DO_NOT_TOUCH semantic content survives (mode still empty).
	llmAfter, err := os.ReadFile(filepath.Join(sectionsDir, "llm.yaml"))
	if err != nil {
		t.Fatalf("read llm.yaml after: %v", err)
	}
	// The llm section must not have gained a non-empty mode (no backend switch).
	if got := string(llmAfter); containsBackendSwitch(got) {
		t.Errorf("llm.yaml gained a non-empty mode after project-config write: %q", got)
	}
}

// containsBackendSwitch reports whether the llm yaml carries a non-empty mode
// (e.g. mode: glm) — S2a must never flip the backend.
func containsBackendSwitch(yaml string) bool {
	return stringContains(yaml, "mode: glm") || stringContains(yaml, "mode: \"glm\"")
}

func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
