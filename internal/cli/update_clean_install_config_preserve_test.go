// Package cli — update_clean_install_config_preserve_test.go
//
// Config-preservation tests for the v2→v3 clean-reinstall path
// (SPEC-UPDATE-REINSTALL-LOOP-001 R3, REQ-RIL-007/008/009/010,
// AC-RIL-005/006/007/010).
//
// Root cause (issue #1084): the clean-reinstall Step 5 force-deploy overwrites
// .claude/settings.json and every .moai/config/sections/*.yaml with template
// defaults, bypassing the normal `moai update` path's 3-way merge — silently
// dropping user customizations (effortLevel, permissions, model, operator name,
// conversation language, brand tokens). These tests reproduce the clobber via
// an overwritingDeployer and assert user config survives the reinstall,
// matching the normal-path protection.

package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// overwritingDeployer is a template.Deployer test double that reproduces the
// force-deploy clobber: on Deploy() it overwrites .claude/settings.json and the
// user/language/design section YAMLs with template-default content, exactly the
// wholesale-overwrite behavior the real forceUpdate deployer performs.
type overwritingDeployer struct {
	deployCalls int
}

func (d *overwritingDeployer) Deploy(ctx context.Context, projectRoot string, mgr manifest.Manager, tmplCtx *template.TemplateContext) error {
	d.deployCalls++
	clobber := map[string]string{
		".claude/settings.json":               `{"model": "sonnet", "effortLevel": "medium"}` + "\n",
		".moai/config/sections/user.yaml":     "user:\n  name: \"\"\n",
		".moai/config/sections/language.yaml": "language:\n  conversation_language: \"en\"\n",
		".moai/config/sections/design.yaml":   "design:\n  default_framework: \"next.js\"\n",
	}
	for rel, content := range clobber {
		abs := filepath.Join(projectRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (d *overwritingDeployer) ListTemplates() []string { return nil }
func (d *overwritingDeployer) ValidateAll(ctx context.Context, c *template.TemplateContext) error {
	return nil
}
func (d *overwritingDeployer) ExtractTemplate(name string) ([]byte, error) { return nil, nil }

// makeConfigPreserveFixture builds a v2 project (so the clean-reinstall body
// runs) whose user config carries clearly-custom values that differ from the
// embedded template defaults:
//   - settings.json: model=opus, effortLevel=high, a user-added permission
//   - user.yaml:     name=GOOS-CUSTOM-NAME  (base: "{{.UserName}}")
//   - language.yaml: conversation_language=ko  (base: "{{.ConversationLanguage}}")
//   - design.yaml:   default_framework=flutter (base ships: "next.js")
func makeConfigPreserveFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	// v2 signal: v2.* version so the clean-reinstall activates.
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v2.16.1\n")
	// A deprecated path so Signal 3 also fires (belt-and-suspenders trigger).
	writeTestFile(t, root, ".claude/agents/moai/manager-strategy.md", "retired\n")

	// User-customized config (the values that MUST survive).
	writeTestFile(t, root, ".claude/settings.json",
		`{"model": "opus", "effortLevel": "high", "permissions": {"allow": ["Bash(mytool:*)"]}}`+"\n")
	writeTestFile(t, root, ".moai/config/sections/user.yaml",
		"user:\n    name: GOOS-CUSTOM-NAME\n")
	writeTestFile(t, root, ".moai/config/sections/language.yaml",
		"language:\n    conversation_language: ko\n")
	writeTestFile(t, root, ".moai/config/sections/design.yaml",
		"design:\n    default_framework: flutter\n")

	// A PRESERVE-inventory seed so the reinstall has inventory work too.
	writeTestFile(t, root, ".moai/specs/SPEC-USER-CFG/spec.md", "user spec\n")
	return root
}

// runConfigPreserveReinstall runs the clean-reinstall with the clobbering
// deployer and returns the project root for post-condition assertions.
func runConfigPreserveReinstall(t *testing.T, root string) {
	t.Helper()
	deployer := &overwritingDeployer{}
	migrate := &stubMigrateRunner{}
	if _, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              io.Discard,
		Deployer:         deployer,
		RunMigrateAgency: migrate.Run,
	}); err != nil {
		t.Fatalf("runCleanReinstall: %v", err)
	}
	if deployer.deployCalls != 1 {
		t.Fatalf("overwritingDeployer.deployCalls = %d; want 1 (deploy must run to reproduce the clobber)", deployer.deployCalls)
	}
}

// yamlNestedString reads projectRoot/rel, unmarshals it, and returns the value
// at parent.child (both string keys). Fails the test on any error.
func yamlNestedString(t *testing.T, root, rel, parent, child string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal %s: %v", rel, err)
	}
	sub, ok := m[parent].(map[string]any)
	if !ok {
		t.Fatalf("%s: parent key %q missing or not a map (content: %s)", rel, parent, data)
	}
	v, ok := sub[child]
	if !ok {
		t.Fatalf("%s: child key %q missing under %q (content: %s)", rel, child, parent, data)
	}
	s, _ := v.(string)
	return s
}

// TestCleanReinstall_SettingsJSONUserKeysPreserved covers AC-RIL-005: user keys
// in .claude/settings.json (effortLevel, a user-added permission) survive the
// clean-reinstall's force-deploy clobber.
func TestCleanReinstall_SettingsJSONUserKeysPreserved(t *testing.T) {
	root := makeConfigPreserveFixture(t)
	runConfigPreserveReinstall(t, root)

	data, err := os.ReadFile(filepath.Join(root, ".claude/settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `"high"`) {
		t.Errorf("settings.json lost user effortLevel=high after clean-reinstall; content: %s", content)
	}
	if !strings.Contains(content, "mytool") {
		t.Errorf("settings.json lost user-added permission (mytool) after clean-reinstall; content: %s", content)
	}
	// The template default effortLevel (medium) must NOT have replaced the user's.
	if strings.Contains(content, `"medium"`) {
		t.Errorf("settings.json shows template default effortLevel=medium; user value was clobbered; content: %s", content)
	}
}

// TestCleanReinstall_ModelPinDoesNotDowngrade covers AC-RIL-010: a user-set
// higher-capability model (opus) survives; the template "model": "sonnet" pin
// does NOT silently downgrade it.
func TestCleanReinstall_ModelPinDoesNotDowngrade(t *testing.T) {
	root := makeConfigPreserveFixture(t)
	runConfigPreserveReinstall(t, root)

	data, err := os.ReadFile(filepath.Join(root, ".claude/settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "opus") {
		t.Errorf("settings.json lost user model=opus after clean-reinstall (silent downgrade); content: %s", content)
	}
	if strings.Contains(content, "sonnet") {
		t.Errorf("settings.json was downgraded to the template model pin (sonnet); content: %s", content)
	}
}

// TestCleanReinstall_AllSectionsYAMLPreserved covers AC-RIL-006: user-populated
// values across MULTIPLE sections/*.yaml (user.yaml name, language.yaml
// conversation_language, design.yaml default_framework) survive — issue #1084
// explicitly reports language.yaml / design.yaml loss, so the AC exercises more
// than user.yaml.
func TestCleanReinstall_AllSectionsYAMLPreserved(t *testing.T) {
	root := makeConfigPreserveFixture(t)
	runConfigPreserveReinstall(t, root)

	if got := yamlNestedString(t, root, ".moai/config/sections/user.yaml", "user", "name"); got != "GOOS-CUSTOM-NAME" {
		t.Errorf("user.yaml name = %q; want GOOS-CUSTOM-NAME (blanked to template default?)", got)
	}
	if got := yamlNestedString(t, root, ".moai/config/sections/language.yaml", "language", "conversation_language"); got != "ko" {
		t.Errorf("language.yaml conversation_language = %q; want ko (reset to template default en?)", got)
	}
	if got := yamlNestedString(t, root, ".moai/config/sections/design.yaml", "design", "default_framework"); got != "flutter" {
		t.Errorf("design.yaml default_framework = %q; want flutter (reset to shipped default next.js?)", got)
	}
}

// TestCleanReinstall_MatchesNormalPathProtection covers AC-RIL-007: the
// clean-reinstall's config handling is equivalent to the normal-path 3-way
// merge — user keys are preserved even when the deploy actively overwrites the
// files with template defaults. This asserts the clean path is NOT a
// lower-protection bypass: a wholesale clobber does not lose user config.
func TestCleanReinstall_MatchesNormalPathProtection(t *testing.T) {
	root := makeConfigPreserveFixture(t)
	runConfigPreserveReinstall(t, root)

	// Every user customization across settings.json + all three section files
	// survives the clobbering deploy — the same protection the normal
	// `moai update` path provides via backup.RestoreMoaiConfig + MergeUserFiles.
	settings, err := os.ReadFile(filepath.Join(root, ".claude/settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	if !strings.Contains(string(settings), "opus") || !strings.Contains(string(settings), "mytool") {
		t.Errorf("clean-reinstall provided lower protection than the normal path: settings.json lost user keys; content: %s", settings)
	}
	if got := yamlNestedString(t, root, ".moai/config/sections/user.yaml", "user", "name"); got != "GOOS-CUSTOM-NAME" {
		t.Errorf("clean-reinstall provided lower protection than the normal path: user.yaml name = %q", got)
	}
}
