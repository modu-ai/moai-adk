package cli

// init_resume_early_ensure_test.go — card t1591 (SPEC-USERASSET-DEPLOY-GUARD-001),
// finding B, re-aimed by card t1619. A fresh init interrupted after the executor must
// leave a resume checkpoint, so the rerun resumes instead of refusing with "already
// initialized". The user-asset ensure runs before the executor now (card t1587): a
// failing ensure leaves no project state at all (see
// TestEnsureFailureBeforeExecutorLeavesNoProjectState). The interruption this finding
// describes is therefore the post-executor window.
//
// The failing run leaves the project half deployed (llm.yaml without deployment_mode)
// and the harness selection record, both on disk when the window fails. No journal is
// written: the window is not an ensure, so the checkpoint rests on the init marker alone.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

// TestInitResumeAfterEarlyEnsureFailureIsResumable reproduces finding B at the
// post-executor window. The first run fails there, after the executor, with the init
// marker and the selection record on disk. The second run uses the real window. After
// the failure the checkpoint must report resumable and the rerun must not be refused. A
// completed resume must leave no checkpoint behind.
func TestInitResumeAfterEarlyEnsureFailureIsResumable(t *testing.T) {
	isolatedInitHome(t)
	projectDir := filepath.Join(t.TempDir(), "early-ensure-proj")
	injectInitWindowFailure(t)

	if err := resumeRun(t, projectDir, "both"); err == nil || !strings.Contains(err.Error(), "initialization interrupted after the executor") {
		t.Fatalf("run 1 did not fail at the post-executor window (the injection did not bite): %v", err)
	}

	// The state the finding describes: no journal, the selection record on disk,
	// and the project half deployed without the post-deploy setup.
	home := userHomeDirOrEmpty()
	if _, err := os.Stat(userassets.JournalPath(home)); !os.IsNotExist(err) {
		t.Fatalf("precondition: a pending-install journal exists after the window failure (stat err = %v)", err)
	}
	if got, ok := readResumeWiring(projectDir); !ok || got != agentWiringBoth {
		t.Fatalf("precondition: the failed run left no selection record (ok=%v, got=%q, want both)", ok, got)
	}
	if llm := readResumeLLMYAML(t, projectDir); strings.Contains(llm, "deployment_mode:") {
		t.Fatalf("precondition: the failed run reached the post-deploy setup (llm.yaml carries deployment_mode):\n%s", llm)
	}

	if !initResumeCheckpoint(home, projectDir) {
		t.Errorf("the post-executor window failure left no resume checkpoint: initResumeCheckpoint reports not resumable")
	}
	err2 := resumeRun(t, projectDir, "")
	if err2 != nil && strings.Contains(err2.Error(), "already initialized") {
		t.Fatalf("the rerun after the window failure is refused with %q", firstLineStr(err2.Error()))
	}
	if err2 != nil {
		t.Fatalf("run 2 (the rerun) failed: %v", err2)
	}
	if initResumeCheckpoint(home, projectDir) {
		t.Errorf("a completed resume left a resume checkpoint behind: the next run would re-enter the resume")
	}
}
