package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
	"gopkg.in/yaml.v3"
)

// update_gate_yaml_preserve_test.go — the gate.yaml update-survival contract
// (SPEC-PRECOMMIT-GATE-SCOPE-001 REQ-006 / AC-004, plus AC-008's existing-install
// hook refresh). gate.yaml sits inside the .moai/config root that
// CleanMoaiManagedPaths deletes wholesale before template redeploy; the update
// pipeline's Backup → Clean Managed Paths → Deploy Templates → Restore Settings
// cycle (3-way node merge, SPEC-UPDATE-YAML-PRESERVE-001) is its only
// protection. These tests drive the REAL template-sync cycle through
// runTemplateSyncAt (the same driver the llm.yaml preservation contract uses)
// and pin, at parsed-key + comment-presence granularity, that hand-edited AND
// web-written values survive a `moai update`.
//
// Fixture honesty: the user fixture is derived from the REAL embedded template
// gate.yaml bytes; every edit applies with must-replace semantics so template
// drift fails the fixture builder instead of yielding a vacuous green.

const (
	// gateUserMarkerComment rides a USER-ADDED key (absent from the template),
	// so the merge retains it with its comment — the same placement rule the
	// llm.yaml contract measured (a comment on a template-carried key is
	// dropped with the template's own key comment).
	gateUserMarkerComment = "# T461-GATE-PRESERVE-MARKER: user note pinned above a user-added key"

	gateUserMarkerKey   = "t461_user_marker_key"
	gateUserMarkerVal   = "observed"
	gateUserStepDisable = "mypy"
)

// readEmbeddedGateYAML returns the REAL embedded template gate.yaml bytes.
func readEmbeddedGateYAML(t *testing.T) []byte {
	t.Helper()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	data, err := fs.ReadFile(fsys, ".moai/config/sections/gate.yaml")
	if err != nil {
		t.Fatalf("read embedded gate.yaml: %v", err)
	}
	return data
}

// buildGateUserYAML derives a user-edited gate.yaml from the embedded template:
// skip_tests re-pinned, a disabled_steps entry added, the pre-commit opt-in
// flipped ON (the value a `moai web` save writes), and a marker comment +
// user-added key inserted. Every needle must appear exactly once in the
// template or the fixture builder fails (template drift ≠ silent vacuity).
func buildGateUserYAML(t *testing.T, base []byte) []byte {
	t.Helper()
	s := string(base)
	s = replaceOnce(t, s, "  skip_tests: false", "  skip_tests: true")
	s = replaceOnce(t, s, "  disabled_steps: {}",
		"  disabled_steps:\n    "+gateUserStepDisable+": false\n"+
			"  "+gateUserMarkerComment+"\n"+
			"  "+gateUserMarkerKey+": "+gateUserMarkerVal)
	s = replaceOnce(t, s, "  pre_commit:\n    enabled: false", "  pre_commit:\n    enabled: true")
	return []byte(s)
}

// makeGatePreserveFixture builds a t.TempDir() project wired for the
// template-sync update cycle, carrying the user-edited gate.yaml.
func makeGatePreserveFixture(t *testing.T, gateYAML []byte) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"system:\n  template_version: \"0.0.0\"\n")
	writeTestFile(t, root, ".moai/manifest.json", "{}\n")
	if gateYAML != nil {
		writeTestFile(t, root, ".moai/config/sections/gate.yaml", string(gateYAML))
	}
	return root
}

// gateYAMLRaw reads the fixture's gate.yaml as text.
func gateYAMLRaw(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "gate.yaml"))
	if err != nil {
		t.Fatalf("read gate.yaml: %v", err)
	}
	return string(data)
}

// TestUpdateGateYAMLPreservesUserValues is AC-004: hand-edited AND web-written
// gate.yaml values (skip_tests, disabled_steps, pre_commit.enabled) survive the
// update cycle; the shipped template defaults do not overwrite them.
func TestUpdateGateYAMLPreservesUserValues(t *testing.T) {
	root := makeGatePreserveFixture(t, buildGateUserYAML(t, readEmbeddedGateYAML(t)))

	runTemplateSyncAt(t, root)

	raw := gateYAMLRaw(t, root)
	var doc struct {
		Gate struct {
			Enabled       bool            `yaml:"enabled"`
			SkipTests     bool            `yaml:"skip_tests"`
			DisabledSteps map[string]bool `yaml:"disabled_steps"`
			PreCommit     struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"pre_commit"`
		} `yaml:"gate"`
	}
	if err := yaml.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("unmarshal post-update gate.yaml: %v\n%s", err, raw)
	}
	if !doc.Gate.SkipTests {
		t.Errorf("skip_tests reverted to template default — user value lost:\n%s", raw)
	}
	v, ok := doc.Gate.DisabledSteps[gateUserStepDisable]
	if !ok || v {
		t.Errorf("disabled_steps entry %q lost or re-enabled (%v, present=%t):\n%s", gateUserStepDisable, v, ok, raw)
	}
	if !doc.Gate.PreCommit.Enabled {
		t.Errorf("pre_commit.enabled reverted to default — the web-written opt-in was lost:\n%s", raw)
	}
	if !strings.Contains(raw, gateUserMarkerComment) {
		t.Errorf("user marker comment lost:\n%s", raw)
	}
	if !strings.Contains(raw, gateUserMarkerKey+": "+gateUserMarkerVal) {
		t.Errorf("user-added key lost:\n%s", raw)
	}
	// Template keys are still delivered: the merge must not turn the file into
	// a bare user document.
	if !strings.Contains(raw, "ast_grep_gate:") {
		t.Errorf("template-carried ast_grep_gate section missing after update:\n%s", raw)
	}
}

// TestUpdateGateYAMLDeliversNewPreCommitKey pins the default-off delivery: a
// project whose gate.yaml predates the pre_commit key receives the new key
// (default false) from the template after update.
func TestUpdateGateYAMLDeliversNewPreCommitKey(t *testing.T) {
	base := string(readEmbeddedGateYAML(t))
	// Simulate a pre-existing install: strip the pre_commit block from the
	// template bytes to stand in for the older shipped gate.yaml.
	idx := strings.Index(base, "  pre_commit:\n")
	if idx < 0 {
		t.Fatal("embedded template gate.yaml carries no pre_commit block — fixture premise broken")
	}
	end := strings.Index(base[idx:], "\n  ast_grep_gate:")
	if end < 0 {
		t.Fatal("embedded template gate.yaml shape drifted — ast_grep_gate anchor missing")
	}
	oldYAML := base[:idx] + base[idx+end+1:]

	root := makeGatePreserveFixture(t, []byte(oldYAML))
	runTemplateSyncAt(t, root)

	if !strings.Contains(gateYAMLRaw(t, root), "pre_commit:") {
		t.Errorf("template's new pre_commit key was not delivered to the existing install:\n%s", gateYAMLRaw(t, root))
	}
}

// TestUpdateReplacesMarkerHookWithCurrentContent is AC-008: an existing
// install whose pre-commit hook carries the MoAI marker (an older shipped
// version) is overwritten with the current hook content by `moai update` —
// no separate migration step.
func TestUpdateReplacesMarkerHookWithCurrentContent(t *testing.T) {
	root := makeGatePreserveFixture(t, nil)
	// An older MoAI hook: same marker line, no MOAI_PRECOMMIT marker export.
	oldHook := "#!/bin/sh\n# MoAI-ADK pre-commit hook — older shipped version\nif ! moai gate; then exit 1; fi\n"
	writeTestFile(t, root, ".git/hooks/pre-commit", oldHook)

	runTemplateSyncAt(t, root)

	data, err := os.ReadFile(filepath.Join(root, ".git", "hooks", "pre-commit"))
	if err != nil {
		t.Fatalf("read post-update pre-commit hook: %v", err)
	}
	got := string(data)
	if got == oldHook {
		t.Error("pre-commit hook was not replaced by the update cycle")
	}
	if !strings.Contains(got, "MOAI_PRECOMMIT=1 moai gate") {
		t.Errorf("replaced hook does not carry the current marker-exporting gate invocation:\n%s", got)
	}
}
