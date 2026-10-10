package cli

// init_resume_early_ensure_test.go — card t1591 (SPEC-USERASSET-DEPLOY-GUARD-001),
// finding B: a fresh init whose user-asset ensure fails before any journal exists
// must leave a resume checkpoint, so the rerun resumes instead of refusing with
// "already initialized".
//
// The failing run leaves the project half deployed (llm.yaml without
// deployment_mode) and the harness selection record, which runInit writes before
// the ensure. A lock refusal writes no journal and the user manifest stays healthy,
// so initResumeCheckpoint finds no interruption evidence and the rerun is refused.

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

// TestInitResumeAfterEarlyEnsureFailureIsResumable reproduces finding B. The first
// run's ensure fails the way a lock acquisition refusal does, before any journal is
// written. The second run uses the real ensure. After the failure the checkpoint must
// report resumable and the rerun must not be refused. A completed resume must leave
// no checkpoint behind.
func TestInitResumeAfterEarlyEnsureFailureIsResumable(t *testing.T) {
	isolatedInitHome(t)
	projectDir := filepath.Join(t.TempDir(), "early-ensure-proj")
	calls := 0
	stubEnsure(t, func(homeDir string, selection []string, out io.Writer) error {
		calls++
		if calls == 1 {
			return errors.New("simulated early ensure failure: user lock acquisition refused")
		}
		return ensureUserAssetsLocked(homeDir, selection, out)
	})

	if err := resumeRun(t, projectDir, "both"); err == nil || !strings.Contains(err.Error(), "user-asset install failed") {
		t.Fatalf("run 1 did not fail at the user-asset ensure (the injection did not bite): %v", err)
	}

	// The state the finding describes: no journal, the selection record on disk,
	// and the project half deployed without the post-deploy setup.
	home := userHomeDirOrEmpty()
	if _, err := os.Stat(userassets.JournalPath(home)); !os.IsNotExist(err) {
		t.Fatalf("precondition: a pending-install journal exists after an early ensure failure (stat err = %v)", err)
	}
	if got, ok := readResumeWiring(projectDir); !ok || got != agentWiringBoth {
		t.Fatalf("precondition: the failed run left no selection record (ok=%v, got=%q, want both)", ok, got)
	}
	if llm := readResumeLLMYAML(t, projectDir); strings.Contains(llm, "deployment_mode:") {
		t.Fatalf("precondition: the failed run reached the post-deploy setup (llm.yaml carries deployment_mode):\n%s", llm)
	}

	if !initResumeCheckpoint(home, projectDir) {
		t.Errorf("the early ensure failure left no resume checkpoint: initResumeCheckpoint reports not resumable")
	}
	err2 := resumeRun(t, projectDir, "")
	if err2 != nil && strings.Contains(err2.Error(), "already initialized") {
		t.Fatalf("the rerun after an early ensure failure is refused with %q", firstLineStr(err2.Error()))
	}
	if err2 != nil {
		t.Fatalf("run 2 (the rerun) failed: %v", err2)
	}
	if initResumeCheckpoint(home, projectDir) {
		t.Errorf("a completed resume left a resume checkpoint behind: the next run would re-enter the resume")
	}
}
