package cli

// m7_gate46_47_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 gate rounds 46/47,
// cli family:
//
//   - 46-3: the resume checkpoint's completeness arm reads BOTH deployment
//     faces. A project carrying llm.yaml AND .mcp.json is COMPLETE — a
//     stale user-global journal (the journal carries no project identity)
//     must not hijack it into a resume that would consume another
//     project's pending intent. A project still missing .mcp.json (the M6
//     window) stays resumable.
//   - 47-4: the resume applies the wizard's Jev answer — a resume that
//     skipped it flipped an explicit wizard selection back to false.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

func corruptUserManifestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte("{ corrupt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func TestInitResumeCheckpointReadsBothDeploymentFaces(t *testing.T) {
	home := corruptUserManifestHome(t)

	// The template-shaped llm.yaml — no deployment_mode key: the M6 window
	// (deployed, post-deploy setup not finished) stays resumable.
	deployed := t.TempDir()
	sections := filepath.Join(deployed, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("harness: claude\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !initResumeCheckpoint(home, deployed) {
		t.Fatal("a deployed project without the post-deploy marker must stay resumable (the M6 window)")
	}

	// The post-path RAN (ApplyDeployMode wrote deployment_mode — the one
	// key the template never ships): COMPLETE — the stale user-global
	// journal must not route it into a resume (gate 46-3).
	completed := t.TempDir()
	sections = filepath.Join(completed, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("harness: claude\ndeployment_mode: local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if initResumeCheckpoint(home, completed) {
		t.Fatal("a COMPLETE project was routed into resume — the stale user-global journal would be consumed as this project's pending intent (gate 46-3)")
	}

	// Undeployed: no llm.yaml — never resumable.
	if initResumeCheckpoint(home, t.TempDir()) {
		t.Fatal("an undeployed project was routed into resume")
	}

	// No interruption (healthy manifest): not resumable regardless.
	healthyHome := t.TempDir()
	if initResumeCheckpoint(healthyHome, deployed) {
		t.Fatal("a healthy user store routed a deployed project into resume")
	}
}

func TestInitResumeAppliesJevSelection(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	// The ensure-failure injection: a corrupt USER manifest makes
	// InstallPreserveSelection fail after the executor finished deploying.
	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte("{ corrupt\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	seamsForInit(t)
	// The wizard ANSWERED yes to Jev (gate 47-4: the resume must keep it)
	// and DECLINED participation (gate 50-3: the resume must record the
	// decline — Asked:true, Enabled:false).
	t.Setenv("CI", "")
	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{AutonomyTier: "automatic", JevEnabled: true, ParticipationEnabled: false}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	projectDir := filepath.Join(t.TempDir(), "jev-resume-proj")

	cmd1 := newInitTestCmd()
	var out1, err1b strings.Builder
	cmd1.SetOut(&out1)
	cmd1.SetErr(&err1b)
	if err := runInit(cmd1, []string{projectDir}); err == nil {
		t.Skip("the ensure-failure injection did not bite — the corrupt-manifest injection needs re-aiming")
	}

	cmd2 := newInitTestCmd()
	var out2, err2b strings.Builder
	cmd2.SetOut(&out2)
	cmd2.SetErr(&err2b)
	if err := runInit(cmd2, []string{projectDir}); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}

	// The Jev selection survived the resume: workflow.yaml's jev.enabled
	// reads true (the shipped default is false).
	wf, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		t.Fatalf("workflow.yaml missing after resume: %v", err)
	}
	if !jevEnabledInWorkflowYAML(string(wf)) {
		t.Fatalf("the resume lost the wizard's Jev selection (gate 47-4) — workflow.yaml jev block:\n%s", wf)
	}

	// Gate 50-3: the participation DECLINE was recorded (Asked, disabled) —
	// a resume without the step left the consent invisible.
	p := config.ReadUserParticipation()
	if !p.Asked {
		t.Fatalf("the resume never recorded the participation answer (Asked=false) — gate 50-3")
	}
	if p.Enabled {
		t.Fatalf("the wizard's participation DECLINE flipped to enabled after the resume — gate 50-3")
	}
}

// jevEnabledInWorkflowYAML reads the jev: block's enabled: value — a small
// targeted scan (the shipped default carries enabled: false).
func jevEnabledInWorkflowYAML(content string) bool {
	lines := strings.Split(content, "\n")
	inJev := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "jev:"):
			inJev = true
		case inJev && strings.HasPrefix(line, " ") && strings.HasPrefix(trimmed, "enabled:"):
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "enabled:")) == "true"
		case inJev && trimmed != "" && !strings.HasPrefix(line, " "):
			return false // the jev block ended
		}
	}
	return false
}

// TestJournalClearedAfterResume pins the paired invariant: the resume that
// consumes the interruption clears the journal (the next run sees no
// checkpoint).
func TestJournalClearedAfterResume(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte("{ corrupt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A stale pending journal from the failed attempt.
	if err := userassets.WriteJournal(userassets.JournalPath(home), &userassets.PendingJournal{BundlesSelection: []string{"extras"}}); err != nil {
		t.Fatal(err)
	}

	seamsForInit(t)
	projectDir := filepath.Join(t.TempDir(), "journal-clear-proj")

	cmd1 := newInitTestCmd()
	var out1, err1b strings.Builder
	cmd1.SetOut(&out1)
	cmd1.SetErr(&err1b)
	if err := runInit(cmd1, []string{projectDir}); err == nil {
		t.Skip("the ensure-failure injection did not bite")
	}

	cmd2 := newInitTestCmd()
	var out2, err2b strings.Builder
	cmd2.SetOut(&out2)
	cmd2.SetErr(&err2b)
	if err := runInit(cmd2, []string{projectDir}); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}

	if _, err := os.Stat(userassets.JournalPath(home)); !os.IsNotExist(err) {
		t.Fatalf("the pending journal survived the resume (stat err = %v) — the next run would re-enter the checkpoint", err)
	}
}
