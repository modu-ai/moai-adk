package cli

// init_resume_precedence_test.go — the restore precedence of the init resume,
// card t1591 (the two residuals left after findings 1 and 2).
//
// Item A: the wizard's agent_wiring default is pre-filled from the recorded
// selection when the resume checkpoint applies, so accepting the default keeps
// the recorded harness instead of the shipped claude default. A changed answer wins.
//
// Item B: the legacy GPT-layout fallback applies only when no explicit --llm was
// given, and never over a record that exists.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
)

// writeResumeCompletenessLLM writes the llm.yaml a failed run deployed. It has no
// deployment_mode key, so the post-deploy setup is unfinished and the checkpoint's
// completeness arm holds.
func writeResumeCompletenessLLM(t *testing.T, projectDir string) {
	t.Helper()
	sections := filepath.Join(projectDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("harness: claude\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// resumeWizardAnswer makes the resume's wizard answer agent_wiring with
// answer(rootFlag), overriding the seamsForInit stub until the test ends.
func resumeWizardAnswer(t *testing.T, answer func(rootFlag string) string) {
	t.Helper()
	prev := runWizardFn
	runWizardFn = func(rootFlag, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{AgentWiring: answer(rootFlag), AutonomyTier: config.AutonomyTierAutomatic}, nil
	}
	t.Cleanup(func() { runWizardFn = prev })
}

// TestResumeAgentWiringDefault_RecordPrefillsOnlyWithCheckpoint — Item A. The
// record pre-fills the wizard's agent_wiring default only when the resume
// checkpoint applies. A record without a checkpoint is stale and pre-fills nothing.
func TestResumeAgentWiringDefault_RecordPrefillsOnlyWithCheckpoint(t *testing.T) {
	cases := []struct {
		name       string
		checkpoint bool   // the interruption checkpoint applies
		record     string // the recorded selection, "" for none
		want       string // resumeAgentWiringDefault's result
	}{
		{"checkpoint and record pre-fill the record", true, "both", "both"},
		{"checkpoint without a record pre-fills nothing", true, "", ""},
		{"a record without a checkpoint is stale", false, "both", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			if tc.checkpoint {
				home = corruptUserManifestHome(t)
			}
			projectDir := filepath.Join(t.TempDir(), "default-proj")
			writeResumeCompletenessLLM(t, projectDir)
			if tc.record != "" {
				if err := persistResumeWiring(projectDir, agentWiring(tc.record)); err != nil {
					t.Fatal(err)
				}
			}

			got := resumeAgentWiringDefault(home, projectDir)
			if got != tc.want {
				t.Fatalf("resumeAgentWiringDefault = %q, want %q", got, tc.want)
			}
			// The wizard's own default follows the same rule: the record when the
			// checkpoint applies, the shipped claude default otherwise.
			wantDefault := tc.want
			if wantDefault == "" {
				wantDefault = "claude"
			}
			q := wizard.QuestionByID(wizard.InitQuestionsWithAgentWiring(projectDir, got), "agent_wiring")
			if q == nil {
				t.Fatal("agent_wiring is absent from the init question set")
			}
			if q.Default != wantDefault {
				t.Fatalf("the wizard's agent_wiring default = %q, want %q", q.Default, wantDefault)
			}
		})
	}
}

// TestInitResumeAcceptedWizardDefaultKeepsRecord — Item A, the reported symptom.
// The failing run selects --llm both. The resume's wizard displays the recorded
// selection as its default and the operator accepts it, so the answer is the
// displayed default. The resume must restore both, not the claude default.
func TestInitResumeAcceptedWizardDefaultKeepsRecord(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "accepted-default-proj")
	resumeSeedRun(t, projectDir, "both")

	resumeWizardAnswer(t, func(rootFlag string) string {
		displayed := resumeAgentWiringDefault(userHomeDirOrEmpty(), rootFlag)
		return wizard.QuestionByID(wizard.InitQuestionsWithAgentWiring(rootFlag, displayed), "agent_wiring").Default
	})
	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}
	if llm := readResumeLLMYAML(t, projectDir); !hasHarness(llm, "both") {
		t.Fatalf("accepting the resume's wizard default discarded the recorded both selection — llm.yaml:\n%s", llm)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".codex", "config.toml")); err != nil {
		t.Fatalf("the accepted default skipped the Codex wiring of the recorded both selection: %v", err)
	}
}

// requireGPTLayout asserts the deployed layout is the codex-only one the legacy
// fallback reads: AGENTS.md present, the .claude/ surface absent.
func requireGPTLayout(t *testing.T, projectDir string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(projectDir, "AGENTS.md")); err != nil {
		t.Fatalf("precondition: the gpt seed run did not deploy AGENTS.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".claude")); err == nil {
		t.Fatal("precondition: the gpt seed run deployed a .claude/ surface, so the layout is not the gpt layout")
	}
}

// resumeForgetRecord removes the selection record the failing run wrote, so the
// resume has no record and only the layout, the flag, or the wizard can decide.
func resumeForgetRecord(t *testing.T, projectDir string) {
	t.Helper()
	if _, err := os.Stat(resumeWiringPath(projectDir)); err != nil {
		t.Fatalf("precondition: the failing run did not write the selection record: %v", err)
	}
	clearResumeWiring(projectDir)
}

// TestInitResumeExplicitLLMClaudeRefusedOverGPTLayout — Item B, restated by Item C.
// An explicit --llm claude on the resume is the current invocation's answer, so the
// legacy layout fallback does not force gpt over it. The answer is then checked
// against the deployed GPT layout: with no record the layout is the deployed harness,
// and a claude request over it is refused (card t1591 Item C).
func TestInitResumeExplicitLLMClaudeRefusedOverGPTLayout(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "explicit-claude-proj")
	resumeSeedRun(t, projectDir, "gpt")
	requireGPTLayout(t, projectDir)
	resumeForgetRecord(t, projectDir)
	llmBefore := readResumeLLMYAML(t, projectDir)

	manifestBefore := readUserManifestForTest(t)
	err := resumeRun(t, projectDir, "claude")
	requireResumeRefused(t, err, projectDir, llmBefore, manifestBefore, "", "gpt", "claude")
}

// TestInitResumeChangedWizardAnswerOverRecordedGPTIsRefused — Item A, restated by
// Item C. While a record exists, the legacy layout fallback does not override a
// changed wizard answer. The answer is still a harness change on the resume: over the
// recorded gpt, claude is refused rather than applied over the gpt files.
func TestInitResumeChangedWizardAnswerOverRecordedGPTIsRefused(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "changed-answer-proj")
	resumeSeedRun(t, projectDir, "gpt")
	requireGPTLayout(t, projectDir)
	llmBefore := readResumeLLMYAML(t, projectDir)

	resumeWizardAnswer(t, func(string) string { return "claude" })
	manifestBefore := readUserManifestForTest(t)
	err := resumeRun(t, projectDir, "")
	requireResumeRefused(t, err, projectDir, llmBefore, manifestBefore, "gpt", "gpt", "claude")
}

// TestInitResumeLegacyFallbackAppliesWithoutRecordOrFlag — a guard on the legacy
// fallback itself. With no record, no --llm, and no wizard answer, the deployed
// GPT layout still decides (gate round 51-3).
func TestInitResumeLegacyFallbackAppliesWithoutRecordOrFlag(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "legacy-fallback-proj")
	resumeSeedRun(t, projectDir, "gpt")
	requireGPTLayout(t, projectDir)
	resumeForgetRecord(t, projectDir)

	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}
	if llm := readResumeLLMYAML(t, projectDir); !hasHarness(llm, "gpt") {
		t.Fatalf("the legacy layout fallback did not restore gpt over a deployed GPT layout with no record — llm.yaml:\n%s", llm)
	}
}

// TestInitResumeLegacyFallbackHoldsAgainstAcceptedDefault — a guard. With no
// record and no --llm, an accepted claude default still yields to the GPT layout.
// The wizard's claude default and an explicit claude answer look the same in the
// result, and the fallback has always read the claude default as no selection.
func TestInitResumeLegacyFallbackHoldsAgainstAcceptedDefault(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "accepted-claude-proj")
	resumeSeedRun(t, projectDir, "gpt")
	requireGPTLayout(t, projectDir)
	resumeForgetRecord(t, projectDir)

	resumeWizardAnswer(t, func(string) string { return "claude" })
	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}
	if llm := readResumeLLMYAML(t, projectDir); !hasHarness(llm, "gpt") {
		t.Fatalf("an accepted claude default over a GPT layout with no record dropped the legacy gpt wiring — llm.yaml:\n%s", llm)
	}
}
