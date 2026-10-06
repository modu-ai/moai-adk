package hook

import (
	"os"
	"path/filepath"
	"testing"
)

// t262 regression: the wrapper-copy contract — the deployed root .sh plus its
// template originals, all byte-identical — is guarded only for stop-goal
// (TestStopGoalWrapperCopiesStayIdentical), and the six wrappers below
// therefore drift silently: the t262 regression left stale copies that
// no test caught and only a hand-run cmp repaired.
//
// Card t1540 retired the template-.sh twin member of the contract: each hook
// now ships exactly ONE template original (the .sh.tmpl, rendered at deploy),
// removing the pair-drift hazard class the 2026-08-15 incident created. The
// contract that remains is two-copy — the deployed root .sh this repo
// executes and the template original `moai update` deploys (the .sh.tmpl for
// the three wrapped events below; the template .sh itself for the tmpl-less
// wrappers). CLAUDE.local.md §2.3 still binds: a fix landing on one copy only
// is either silently reverted on the next update or never reaches users at
// all — edit every copy in the same commit.

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
			filepath.Join(templateDir, "handle-agent-hook.sh.tmpl"),
		}},
		{name: "handle-task-completed.sh", copies: []string{
			filepath.Join(rootDir, "handle-task-completed.sh"),
			filepath.Join(templateDir, "handle-task-completed.sh.tmpl"),
		}},
		{name: "handle-teammate-idle.sh", copies: []string{
			filepath.Join(rootDir, "handle-teammate-idle.sh"),
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
