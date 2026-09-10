package cli

// todo_temp_guard_test.go — SPEC-TODO-HOME-TEMP-GUARD-001 M2, the command-path
// half of AC-THG-005: a refusal must not read as a silent success.
//
// The console half (the pure resolver stays silent and writes nothing) lives in
// internal/kanban; this file owns the surface only the command has — guidance
// on stderr, and an exit status that says the run continued.
//
// These tests mutate CLAUDE_PROJECT_DIR and the userHomeDirFn seam, both
// process-global, so they do not run in parallel.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestTempOriginGuidance_NamesRootsAndContinues — AC-THG-005.
//
// Three things the guidance must carry, and all three are asserted: the
// matched temp root, the substitute root the run continues against, and the
// fact of continuing (exit 0 plus a working queue). A refusal that printed
// nothing would be indistinguishable from the silent home-queue creation it
// replaced; one that printed only "refused" would leave the operator unable to
// find their cards.
func TestTempOriginGuidance_NamesRootsAndContinues(t *testing.T) {
	dir := t.TempDir() // not a git repository, and a temporary origin by location
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	canaryHome := t.TempDir()
	orig := userHomeDirFn
	userHomeDirFn = func() (string, error) { return canaryHome, nil }
	t.Cleanup(func() { userHomeDirFn = orig })

	matched, isTemp := kanban.TempOriginReason(dir)
	if !isTemp {
		t.Fatalf("precondition: %q must classify as a temporary origin", dir)
	}

	out, errOut, err := runTodo(t, "add", "t536 guidance probe card")
	if err != nil {
		t.Fatalf("the run did not continue: %v (stderr: %s)", err, errOut)
	}

	if !strings.Contains(errOut, matched) {
		t.Errorf("guidance does not name the matched temp root %q:\n%s", matched, errOut)
	}
	if !strings.Contains(errOut, dir) {
		t.Errorf("guidance does not name the substitute root %q the run continues against:\n%s", dir, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Errorf("the command produced no stdout, so it did not actually do the work it continued into")
	}

	// The run continued against the project-local queue, and no home queue was
	// created on the way.
	root := resolveTodoQueueRoot()
	if root != dir {
		t.Errorf("queue root = %q, want the substitute root %q", root, dir)
	}
	rec, loadErr := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root)).Load()
	if loadErr != nil {
		t.Fatalf("load the project-local queue: %v", loadErr)
	}
	if len(rec.Items) == 0 {
		t.Errorf("the card landed nowhere: the queue at %s is empty", kanban.BacklogPathForRoot(root))
	}
	if entries, readErr := os.ReadDir(filepath.Join(canaryHome, ".moai", "db")); readErr == nil && len(entries) > 0 {
		t.Errorf("canary HOME polluted: %d entr(ies) under %s",
			len(entries), filepath.Join(canaryHome, ".moai", "db"))
	}
}

// TestTempOriginGuidance_SilentOnNonTemporaryBase — the guidance's other half.
//
// A non-temporary non-git base keeps its home fallback (AC-THG-003), so no
// refusal happens and nothing is said. Without this, guidance that fired
// unconditionally would pass the test above while shouting at every user.
func TestTempOriginGuidance_SilentOnNonTemporaryBase(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	canaryHome := t.TempDir()
	orig := userHomeDirFn
	userHomeDirFn = func() (string, error) { return canaryHome, nil }
	t.Cleanup(func() { userHomeDirFn = orig })
	declareNonTemporaryQueueBase(t)
	assertQueueSeamHeard(t, dir)

	_, errOut, err := runTodo(t, "add", "t536 non-temporary probe card")
	if err != nil {
		t.Fatalf("add on a non-temporary base failed: %v (stderr: %s)", err, errOut)
	}
	if strings.Contains(errOut, "no home queue was created") {
		t.Errorf("guidance fired on a non-temporary base — the guard's trigger is a temporary origin, not the absence of git:\n%s", errOut)
	}
}

func TestTempOriginGuidance_SilentWithExplicitAbsoluteMOAIHome(t *testing.T) {
	dir := t.TempDir()
	homeRoot := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	t.Setenv("MOAI_HOME", homeRoot)

	_, errOut, err := runTodo(t, "add", "explicit home probe")
	if err != nil {
		t.Fatalf("add with explicit MOAI_HOME failed: %v (stderr: %s)", err, errOut)
	}
	if strings.Contains(errOut, "no home queue was created") {
		t.Errorf("explicit absolute MOAI_HOME was falsely reported as refused:\n%s", errOut)
	}
	root := resolveTodoQueueRoot()
	queuePath := todoBacklogPath(root)
	if !strings.HasPrefix(queuePath, filepath.Join(homeRoot, "db")+string(os.PathSeparator)) {
		t.Errorf("queue path = %q, want beneath explicit MOAI_HOME %q", queuePath, homeRoot)
	}
}
