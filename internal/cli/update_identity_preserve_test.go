package cli

// Card t1139 — `moai update` must not reset the project/user identity or a
// wizard-patched key to the template default.
//
// The 3-way section merge takes its BASE from the persistent template snapshot
// (.moai/cache/template-snapshot). When that snapshot records the tree AFTER the
// init wizard patches or AFTER an update's restore, BASE equals the user's own
// value, the merge reads "the user did not customize this", and the NEW template
// render wins. These tests drive the real `moai init` followed by the real
// template-sync update cycle, so they bind to the actual write order rather than
// to a hand-built snapshot.
//
// prepareSafeInitHome sets environment variables and the update helper chdirs,
// so no test in this file may run in parallel.

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

const (
	identityProjectName = "p-gpt"
	identityUserName    = "goos"
)

// initIdentityProject runs the real runInit against a fresh t.TempDir() project
// named identityProjectName, with the wizard answering identityUserName and
// LSP enabled (the wizard-patched key the template ships as false).
func initIdentityProject(t *testing.T) string {
	t.Helper()
	prepareSafeInitHome(t)

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	wizResult := &wizard.WizardResult{
		ConversationLang:          "en",
		UserName:                  identityUserName,
		AgentWiring:               "claude",
		AutonomyTier:              "semi-auto",
		LSPEnabled:                true,
		EnforceQuality:            true,
		CoverageExemptionsEnabled: false,
		DesignEnabled:             true,
		ClaudeDesignEnabled:       true,
	}
	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) { return wizResult, nil }
	t.Cleanup(func() { runWizardFn = origWizard })

	projectDir := filepath.Join(t.TempDir(), identityProjectName)
	cmd := newInitTestCmd()
	// An absolute positional argument names the project by the whole path
	// (init.go uses the argument verbatim), so the name is pinned by flag.
	if err := cmd.Flags().Set("name", identityProjectName); err != nil {
		t.Fatalf("set --name: %v", err)
	}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v (stderr: %s)", err, errBuf.String())
	}
	return projectDir
}

// runForcedTemplateSyncAt drives the real template-sync update cycle with
// --force (the version check is skipped, backup → clean → deploy → restore
// still runs), from inside root.
func runForcedTemplateSyncAt(t *testing.T, root string) string {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if chErr := os.Chdir(root); chErr != nil {
		t.Fatalf("chdir to fixture: %v", chErr)
	}
	defer func() { _ = os.Chdir(origDir) }()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("config", false, "")
	_ = cmd.Flags().Set("force", "true")
	_ = cmd.Flags().Set("yes", "true")
	var buf, errBuf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())

	if syncErr := runTemplateSyncWithReporter(cmd, nil, true); syncErr != nil {
		t.Fatalf("runTemplateSyncWithReporter: %v\noutput: %s\nstderr: %s", syncErr, buf.String(), errBuf.String())
	}
	return buf.String()
}

// sectionValue resolves a dotted key path inside a YAML file (sections file or
// snapshot file) and returns the leaf value.
func sectionValue(t *testing.T, file string, path ...string) any {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal %s: %v", file, err)
	}
	var cur any = doc
	for _, key := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("%s: path %v reached a non-mapping at %q", file, path, key)
		}
		cur, ok = m[key]
		if !ok {
			t.Fatalf("%s: key %q missing (path %v)", file, key, path)
		}
	}
	return cur
}

func sectionsFile(root, name string) string {
	return filepath.Join(root, ".moai", "config", "sections", name)
}

func snapshotFile(root, name string) string {
	return filepath.Join(backup.SnapshotDir(root), "sections", name)
}

// assertIdentityKeys checks the three keys the defect resets.
func assertIdentityKeys(t *testing.T, root, stage string) {
	t.Helper()
	if got := sectionValue(t, sectionsFile(root, "project.yaml"), "project", "name"); got != identityProjectName {
		t.Errorf("%s: project.name = %v, want %q", stage, got, identityProjectName)
	}
	if got := sectionValue(t, sectionsFile(root, "user.yaml"), "user", "name"); got != identityUserName {
		t.Errorf("%s: user.name = %v, want %q", stage, got, identityUserName)
	}
	if got := sectionValue(t, sectionsFile(root, "lsp.yaml"), "lsp", "enabled"); got != true {
		t.Errorf("%s: lsp.enabled = %v, want true", stage, got)
	}
}

// TestUpdateForce_PreservesInitIdentityAndWizardKeys is the t1139 primary
// reproduction: after a real init, a forced update must keep project.name,
// user.name, and the wizard-patched lsp.enabled — and must keep them across a
// SECOND update too.
func TestUpdateForce_PreservesInitIdentityAndWizardKeys(t *testing.T) {
	root := initIdentityProject(t)

	// Positive control: init actually wrote the three values, so a failure
	// below is the update losing them, not init never writing them.
	assertIdentityKeys(t, root, "after init")
	if t.Failed() {
		t.Fatalf("init did not produce the expected identity keys; the update assertions would be vacuous")
	}

	runForcedTemplateSyncAt(t, root)
	assertIdentityKeys(t, root, "after update #1")

	runForcedTemplateSyncAt(t, root)
	assertIdentityKeys(t, root, "after update #2")
}

// TestUpdateForce_SnapshotIsTheDeployedRender pins the snapshot invariant the
// defect violated: after init and after each update the snapshot records what
// the template render produced, not the user's patched/restored value. The
// template ships lsp.enabled: false; the user has true.
func TestUpdateForce_SnapshotIsTheDeployedRender(t *testing.T) {
	root := initIdentityProject(t)

	if got := sectionValue(t, snapshotFile(root, "lsp.yaml"), "lsp", "enabled"); got != false {
		t.Errorf("after init: snapshot lsp.enabled = %v, want false (the template render, before the wizard patch)", got)
	}

	runForcedTemplateSyncAt(t, root)
	if got := sectionValue(t, snapshotFile(root, "lsp.yaml"), "lsp", "enabled"); got != false {
		t.Errorf("after update: snapshot lsp.enabled = %v, want false (the deployed render, not the restored user value)", got)
	}
	// The live file still carries the user's value (positive control that the
	// snapshot and the live tree are genuinely different at this key).
	if got := sectionValue(t, sectionsFile(root, "lsp.yaml"), "lsp", "enabled"); got != true {
		t.Errorf("after update: live lsp.enabled = %v, want true", got)
	}
}

// TestUpdateForce_UserCustomizationSurvivesTwoUpdates is the second-cycle
// poisoning case: a key the user edited after init survives the first update
// (BASE = template render) and must also survive the second. With a snapshot
// written after the restore, the second update's BASE is the user value and the
// template default wins.
func TestUpdateForce_UserCustomizationSurvivesTwoUpdates(t *testing.T) {
	root := initIdentityProject(t)

	qualityPath := sectionsFile(root, "quality.yaml")
	data, err := os.ReadFile(qualityPath)
	if err != nil {
		t.Fatalf("read quality.yaml: %v", err)
	}
	edited := replaceOnce(t, string(data), "test_coverage_target: 85", "test_coverage_target: 91")
	if err := os.WriteFile(qualityPath, []byte(edited), 0o644); err != nil {
		t.Fatalf("write quality.yaml: %v", err)
	}

	for i := 1; i <= 2; i++ {
		runForcedTemplateSyncAt(t, root)
		if got := sectionValue(t, qualityPath, "constitution", "test_coverage_target"); got != 91 {
			t.Errorf("after update #%d: test_coverage_target = %v, want the user's 91", i, got)
		}
	}
}

// TestUpdateForce_TemplateChangedKeyStillPropagates guards the other half of
// the merge contract: a key the user never changed while the template did must
// take the new template value. Simulated by giving the live file AND the
// snapshot the same "previous template" value for a key the current template
// renders differently.
func TestUpdateForce_TemplateChangedKeyStillPropagates(t *testing.T) {
	root := initIdentityProject(t)

	const oldTemplateValue = ".moai/previous-template-rules"
	const currentTemplateValue = ".moai/config/astgrep-rules"
	for _, p := range []string{sectionsFile(root, "lsp.yaml"), snapshotFile(root, "lsp.yaml")} {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		patched := replaceOnce(t, string(data), `rules_dir: "`+currentTemplateValue+`"`, `rules_dir: "`+oldTemplateValue+`"`)
		if err := os.WriteFile(p, []byte(patched), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	runForcedTemplateSyncAt(t, root)
	if got := sectionValue(t, sectionsFile(root, "lsp.yaml"), "lsp", "delegate_to_astgrep", "rules_dir"); got != currentTemplateValue {
		t.Errorf("rules_dir = %v, want the new template value %q (user never customized it)", got, currentTemplateValue)
	}
}
