package cli

// init_resume_profile_sync_test.go — Item D (card t1591, the finding on 01e510335):
// the resume runs the same profile synchronization as a healthy init.
//
// A healthy init writes the profile preferences into the project config after the
// template deploy, and the wizard's user name and conversation language take
// precedence. The resume dispatch returned before that call, so a resumed run never
// carried the current preferences or the wizard's identity answers into
// .moai/config/sections/ (user.yaml and language.yaml).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
)

// resumeWizardIdentity makes the resume's wizard answer the user name and the
// conversation language, overriding the seamsForInit stub until the test ends. The
// harness answer stays empty, so the selection record decides the wiring.
func resumeWizardIdentity(t *testing.T, userName, conversationLang string) {
	t.Helper()
	prev := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{
			UserName:         userName,
			ConversationLang: conversationLang,
			AutonomyTier:     config.AutonomyTierAutomatic,
		}, nil
	}
	t.Cleanup(func() { runWizardFn = prev })
}

// TestInitResumeSyncsWizardNameAndLanguage — Item D. The seed run deploys the project
// with no identity answers. The resume's wizard answers a user name and a conversation
// language. After the resume, the project config must carry the wizard's values, as
// the profile sync of a healthy init leaves them.
func TestInitResumeSyncsWizardNameAndLanguage(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "sync-resume-proj")
	resumeSeedRun(t, projectDir, "claude")

	const wantName, wantLang = "Resumed Operator", "ja"
	seeded, err := config.NewConfigManager().LoadRaw(projectDir)
	if err != nil {
		t.Fatalf("precondition: the seed run left an unreadable project config: %v", err)
	}
	if seeded.User.Name == wantName || seeded.Language.ConversationLanguage == wantLang {
		t.Fatalf("precondition: the seed run already carries the resume's identity (name %q, language %q)",
			seeded.User.Name, seeded.Language.ConversationLanguage)
	}

	resumeWizardIdentity(t, wantName, wantLang)
	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}

	after, err := config.NewConfigManager().LoadRaw(projectDir)
	if err != nil {
		t.Fatalf("read the project config after the resume: %v", err)
	}
	if after.User.Name != wantName {
		t.Errorf("the resume did not sync the wizard's user name into user.yaml: got %q, want %q", after.User.Name, wantName)
	}
	if after.Language.ConversationLanguage != wantLang {
		t.Errorf("the resume did not sync the wizard's conversation language into language.yaml: got %q, want %q",
			after.Language.ConversationLanguage, wantLang)
	}
}

// TestInitResumeProfileSyncFailureKeepsMarker — Item D. A sync failure is a post-step
// failure: the resume returns the error and the resume marker stays in place, as the
// other post-steps do, so the next run still resumes. The failure is injected with a
// malformed user.yaml, and the resume's wizard answers a user name, so the sync must
// patch that file and the patch refuses it. A sync with nothing to write reads nothing
// and cannot fail, which is why the wizard answer is part of the injection.
func TestInitResumeProfileSyncFailureKeepsMarker(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "sync-failure-proj")
	resumeSeedRun(t, projectDir, "claude")

	userYAML := filepath.Join(projectDir, ".moai", "config", "sections", "user.yaml")
	if err := os.WriteFile(userYAML, []byte("user:\n  name: [unclosed\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	resumeWizardIdentity(t, "Resumed Operator", "")
	err := resumeRun(t, projectDir, "")
	if err == nil || !strings.Contains(err.Error(), "profile sync") {
		t.Fatalf("the resume did not surface the profile-sync failure (err = %v): the sync did not run as a post-step", err)
	}
	if _, statErr := os.Stat(filepath.Join(projectDir, ".moai", "resume-pending")); statErr != nil {
		t.Fatalf("the resume marker did not stay in place after the sync failure: %v", statErr)
	}
}
