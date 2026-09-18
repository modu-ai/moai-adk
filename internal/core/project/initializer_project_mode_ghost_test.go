package project

// initializer_project_mode_ghost_test.go — SPEC-INIT-UPDATE-CONSISTENCY-001
// REQ-ICU-001 (F8): project.mode is a ghost configuration key. The template
// shipped it, the init pipeline wrote and patched it, and no Go component
// anywhere reads it. These tests assert the key's ABSENCE (not merely a
// false-ish value) so the ghost cannot quietly grow back.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// TestWritePhase1Configs_DoesNotPersistProjectMode pins the write-path half of
// REQ-ICU-001: the init pipeline shall not persist a project.mode key. After
// the ghost removal, WritePhase1Configs must not create project.yaml at all —
// the file's other keys are written by the template deploy / fallback path,
// never by the Page-3 expansion.
//
// Pre-fix behavior: writeProjectModeYAML created project.yaml containing
// "mode: personal" on this fresh path, which is exactly the assertion below.
func TestWritePhase1Configs_DoesNotPersistProjectMode(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	opts := InitOptions{ProjectRoot: root}
	result := &InitResult{}
	if err := WritePhase1Configs(opts, result); err != nil {
		t.Fatalf("WritePhase1Configs: %v", err)
	}

	if _, err := os.Stat(filepath.Join(sectionsDir, defs.ProjectYAML)); err == nil {
		data, _ := os.ReadFile(filepath.Join(sectionsDir, defs.ProjectYAML))
		t.Errorf("WritePhase1Configs created project.yaml — project.mode is a ghost key with no Go reader (REQ-ICU-001); got:\n%s", data)
	}
}

// TestProjectYAMLTemplateCarriesNoModeKey pins the template half of REQ-ICU-001:
// the shipped project.yaml must not carry a `mode:` key under the project
// mapping. Asserting the KEY's absence (rather than its value) is deliberate —
// a just-false assertion was the F1-class miss this SPEC exists to close.
func TestProjectYAMLTemplateCarriesNoModeKey(t *testing.T) {
	t.Parallel()
	// go test runs with the package directory as cwd, so the repo-relative
	// template path resolves against the tree under test.
	templatePath := filepath.Join("..", "..", "template", "templates", ".moai", "config", "sections", string(defs.ProjectYAML)+".tmpl")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", templatePath, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed == "mode: personal" || bytes.HasPrefix([]byte(trimmed), []byte("mode:")) {
			t.Errorf("project.yaml.tmpl still carries a mode key (REQ-ICU-001): %q", trimmed)
		}
	}
}
