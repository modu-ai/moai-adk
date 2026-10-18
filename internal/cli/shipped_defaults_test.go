package cli

// shipped_defaults_test.go — card t1866: what a fresh project receives, and what an
// update leaves alone. The runs are in-process and sit behind prepareSafeInitHome,
// whose real-home fingerprint fails the test if the real user folders changed,
// so "the shipped defaults came out" is observed without touching them.

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestInit_FreshProjectCarriesTheShippedDefaults runs the real init command
// function non-interactively and reads what it wrote. The literals are
// deliberate: the Go constants and the embedded template have their own tests, and
// this one reads the deployed files.
func TestInit_FreshProjectCarriesTheShippedDefaults(t *testing.T) {
	projectDir, _ := quietWizardRunInit(t, t.TempDir(), nil, nil)
	sections := filepath.Join(projectDir, defs.MoAIDir, defs.SectionsSubdir)

	var reportDoc struct {
		Report struct {
			Format string `yaml:"format"`
		} `yaml:"report"`
	}
	if err := quietWizardReadYAML(filepath.Join(sections, defs.ReportYAML), &reportDoc); err != nil {
		t.Fatalf("read report.yaml: %v", err)
	}
	if reportDoc.Report.Format != "artifact" {
		t.Errorf("report.yaml report.format = %q, want artifact", reportDoc.Report.Format)
	}

	cfg, err := config.NewLoader().Load(filepath.Join(projectDir, defs.MoAIDir))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	for slot, want := range map[string]string{
		"high":   "max",
		"medium": "high",
		"low":    "max",
		"fable":  "high",
	} {
		var got string
		switch slot {
		case "high":
			got = cfg.LLM.GLM.Effort.High
		case "medium":
			got = cfg.LLM.GLM.Effort.Medium
		case "low":
			got = cfg.LLM.GLM.Effort.Low
		case "fable":
			got = cfg.LLM.GLM.Effort.Fable
		}
		if got != want {
			t.Errorf("llm.glm.effort.%s = %q, want %q", slot, got, want)
		}
	}
}

// TestUpdate_KeepsUserValuesForTheDefaultsT1866Changed drives the real update
// cycle over a project whose files hold the pre-t1866 defaults as the user's own
// choices (effort.low=low, effort.fable=max, report.format=md) and a profile that
// stored permission_mode acceptEdits: none of them may move to the new defaults.
func TestUpdate_KeepsUserValuesForTheDefaultsT1866Changed(t *testing.T) {
	isolateProfileStore(t)
	if err := profile.WritePreferences("default", profile.ProfilePreferences{PermissionMode: "acceptEdits"}); err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	root := makeLLMPreserveFixture(t, false, false)
	llm := string(embeddedLLMYAMLBytes(t))
	llm = replaceOnce(t, llm, "      low: \"max\"\n      fable: \"high\"", "      low: \"low\"\n      fable: \"max\"")
	writeTestFile(t, root, ".moai/config/sections/llm.yaml", llm)
	writeTestFile(t, root, ".moai/config/sections/report.yaml", "report:\n  format: md\n")

	runTemplateSyncAt(t, root)

	if got := llmYAMLString(t, root, "llm", "glm", "effort", "low"); got != "low" {
		t.Errorf("llm.glm.effort.low = %q after update, want the user's low", got)
	}
	if got := llmYAMLString(t, root, "llm", "glm", "effort", "fable"); got != "max" {
		t.Errorf("llm.glm.effort.fable = %q after update, want the user's max", got)
	}
	var reportDoc struct {
		Report struct {
			Format string `yaml:"format"`
		} `yaml:"report"`
	}
	if err := quietWizardReadYAML(filepath.Join(root, ".moai/config/sections/report.yaml"), &reportDoc); err != nil {
		t.Fatalf("read report.yaml: %v", err)
	}
	if reportDoc.Report.Format != "md" {
		t.Errorf("report.yaml report.format = %q after update, want the user's md", reportDoc.Report.Format)
	}

	prefs, err := profile.ReadPreferences("default")
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if prefs.PermissionMode != "acceptEdits" {
		t.Errorf("profile permission_mode = %q after update, want the stored acceptEdits", prefs.PermissionMode)
	}
}
