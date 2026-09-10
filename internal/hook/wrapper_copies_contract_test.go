package hook

import (
	"os"
	"path/filepath"
	"testing"
)

// t262 regression: the three-copy contract — root deployed .sh + template .sh
// + template .sh.tmpl, all byte-identical — is guarded only for stop-goal
// (TestStopGoalWrapperCopiesStayIdentical), and the pair-drift guard
// (internal/template TestHookWrapperPairParity) compares the .sh/.sh.tmpl pair
// inside the template tree but never the deployed root copy. The six wrappers
// below therefore drift silently: the t262 regression left stale copies that
// no test caught and only a hand-run cmp repaired.
//
// This guard extends the stop-goal identity contract to those six. Three-copy
// wrappers carry .sh + .sh.tmpl + root .sh; tmpl-less wrappers carry no .tmpl,
// so their contract is the two .sh copies (template source + deployed root).
// CLAUDE.local.md §2.3: the .tmpl is what `moai update` deploys and the root
// copy is what this repo executes, so a fix landing on one copy only is either
// silently reverted on the next update or never reaches users at all — edit
// every copy in the same commit.

type wrapperContract struct {
	name   string
	copies []string
}

// hookWrapperContracts lists the six wrappers the stop-goal identity test does
// not cover, with every shipped copy of each. The copy lists are explicit
// rather than discovered on disk so that a deleted copy fails the test (read
// error) instead of silently shrinking the swept set.
func hookWrapperContracts() []wrapperContract {
	const (
		rootDir     = ".claude/hooks/moai"
		templateDir = "internal/template/templates/.claude/hooks/moai"
	)
	return []wrapperContract{
		{name: "handle-agent-hook.sh", copies: []string{
			filepath.Join(rootDir, "handle-agent-hook.sh"),
			filepath.Join(templateDir, "handle-agent-hook.sh"),
			filepath.Join(templateDir, "handle-agent-hook.sh.tmpl"),
		}},
		{name: "handle-task-completed.sh", copies: []string{
			filepath.Join(rootDir, "handle-task-completed.sh"),
			filepath.Join(templateDir, "handle-task-completed.sh"),
			filepath.Join(templateDir, "handle-task-completed.sh.tmpl"),
		}},
		{name: "handle-teammate-idle.sh", copies: []string{
			filepath.Join(rootDir, "handle-teammate-idle.sh"),
			filepath.Join(templateDir, "handle-teammate-idle.sh"),
			filepath.Join(templateDir, "handle-teammate-idle.sh.tmpl"),
		}},
		{name: "handle-session-start-compact.sh", copies: []string{
			filepath.Join(rootDir, "handle-session-start-compact.sh"),
			filepath.Join(templateDir, "handle-session-start-compact.sh"),
		}},
		{name: "status-transition-ownership.sh", copies: []string{
			filepath.Join(rootDir, "status-transition-ownership.sh"),
			filepath.Join(templateDir, "status-transition-ownership.sh"),
		}},
		{name: "sync-phase-quality-gate.sh", copies: []string{
			filepath.Join(rootDir, "sync-phase-quality-gate.sh"),
			filepath.Join(templateDir, "sync-phase-quality-gate.sh"),
		}},
	}
}

// TestHookWrapperCopiesStayIdentical locks the same Template-First contract
// TestStopGoalWrapperCopiesStayIdentical locks for stop-goal, extended to the
// six wrappers the t262 regression drifted silently.
func TestHookWrapperCopiesStayIdentical(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := filepath.Join(cwd, "..", "..")

	for _, wc := range hookWrapperContracts() {
		t.Run(wc.name, func(t *testing.T) {
			first, err := os.ReadFile(filepath.Join(repoRoot, wc.copies[0]))
			if err != nil {
				t.Fatalf("read %s: %v", wc.copies[0], err)
			}
			for _, rel := range wc.copies[1:] {
				b, err := os.ReadFile(filepath.Join(repoRoot, rel))
				if err != nil {
					t.Fatalf("read %s: %v", rel, err)
				}
				if string(b) != string(first) {
					t.Errorf("%s differs from %s — all shipped copies must stay byte-identical (CLAUDE.local.md §2.3: moai update deploys the .tmpl and re-syncs the root copy, so a one-copy edit is reverted or never reaches users)",
						rel, wc.copies[0])
				}
			}
		})
	}
}
