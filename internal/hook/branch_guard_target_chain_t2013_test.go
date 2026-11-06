package hook

// branch_guard_target_chain_t2013_test.go — tracked item 1 (source t2021 static
// review): is `cd <worktree> && cd <primary> && git reset --hard` allowed? A
// table over the real git fixture (primary checkout on main, deny_commits_on
// [main], one real linked worktree) and the real preToolHandler, for BOTH hook
// cwds. Each cell is observed twice: through the whole handler (what the user
// sees) and through checkBranchState alone (the BranchGuard's own verdict,
// because the generic dangerous-command check asks about `git reset --hard`
// before the BranchGuard runs and so hides it; commit verbs go through
// checkProtectedCommit). The rows with hook cwd =
// primary checkout are asserted: a mutating git verb that really lands in the
// primary checkout must never be allowed. The rows with hook cwd = linked
// worktree assert the narrow repair: any literal `cd <abs>` naming the primary
// checkout makes the command primary-judged (cells a, b, f deny), while a
// command whose only cd target is the worktree stays allowed (cells e, i).

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBranchGuardChainedCdDecisionTable(t *testing.T) {
	skipWindowsPOSIXCommandText(t)
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	handler := newTargetGuardHandler(fx)

	// lands: where the verb really runs, as the shell reads the command.
	const (
		landsPrimary  = "primary"
		landsWorktree = "worktree"
	)
	// wtGuard is the BranchGuard-level expectation for the rows whose hook cwd
	// is the linked worktree (narrow repair of tracked item 1): a literal
	// `cd <abs>` naming the primary checkout anywhere in the command makes the
	// command primary-judged; a command whose only cd target is the worktree
	// stays allowed. "" means observed only.
	const (
		wtDeny  = "deny"
		wtAllow = "allow"
	)
	cells := []struct {
		name    string
		command string
		lands   string
		wtGuard string
	}{
		{"a: cd wt && cd primary && git reset --hard", "cd " + fx.release + " && cd " + fx.primary + " && git reset --hard", landsPrimary, wtDeny},
		{"b: cd wt && cd primary && git commit", "cd " + fx.release + " && cd " + fx.primary + " && git commit -m x", landsPrimary, wtDeny},
		{"c: cd primary && git reset --hard", "cd " + fx.primary + " && git reset --hard", landsPrimary, ""},
		{"d: git reset --hard", "git reset --hard", landsPrimary, ""},
		{"e: cd wt && git reset --hard", "cd " + fx.release + " && git reset --hard", landsWorktree, wtAllow},
		{"f: cd wt && cd primary && git switch other", "cd " + fx.release + " && cd " + fx.primary + " && git switch other", landsPrimary, wtDeny},
		{"g: cd primary && git switch other", "cd " + fx.primary + " && git switch other", landsPrimary, ""},
		{"h: git switch other", "git switch other", landsPrimary, ""},
		{"i: cd wt && git switch other", "cd " + fx.release + " && git switch other", landsWorktree, wtAllow},
	}

	for _, cwdCase := range []struct {
		label string
		cwd   string
	}{
		{"cwd=primary", fx.primary},
		{"cwd=worktree", fx.release},
	} {
		for _, cell := range cells {
			decision, reason := affObserve(t, handler, cwdCase.cwd, cell.command)
			sentinel := strings.HasPrefix(reason, branchGuardViolationPrefix+":")

			toolInput, err := json.Marshal(map[string]string{"command": cell.command})
			if err != nil {
				t.Fatal(err)
			}
			guardInput := &HookInput{
				SessionID: "sess-t2013-chain", HookEventName: "PreToolUse", ToolName: "Bash",
				CWD: cwdCase.cwd, ToolInput: json.RawMessage(toolInput),
			}
			guardDecision, guardReason := checkBranchState(guardInput, fx.primary)
			if guardDecision == "" {
				guardDecision, guardReason = checkProtectedCommit(guardInput, fx.primary, []string{"main"})
			}
			guardDenies := guardDecision == DecisionDeny && strings.HasPrefix(guardReason, branchGuardViolationPrefix+":")
			t.Logf("%s | %s -> handler=%q sentinel=%t | BranchGuard-alone deny=%t", cwdCase.label, cell.name, decision, sentinel, guardDenies)

			if cwdCase.cwd != fx.primary {
				switch cell.wtGuard {
				case wtDeny:
					if !guardDenies {
						t.Errorf("%s | %s: BranchGuard alone = %q / %q, want deny with the sentinel (a literal cd names the primary checkout)", cwdCase.label, cell.name, guardDecision, guardReason)
					}
				case wtAllow:
					if guardDenies || sentinel {
						t.Errorf("%s | %s is denied by the BranchGuard, want no BranchGuard deny (its only cd target is the worktree)", cwdCase.label, cell.name)
					}
				}
				continue
			}
			if cell.lands == landsWorktree {
				if guardDenies || sentinel {
					t.Errorf("%s | %s is denied by the BranchGuard, want no BranchGuard deny (the verb runs in the worktree)", cwdCase.label, cell.name)
				}
				continue
			}
			if decision != DecisionDeny && decision != DecisionAsk {
				t.Errorf("%s | %s = %q, want deny or ask (the verb lands in the primary checkout)", cwdCase.label, cell.name, decision)
			}
			if !guardDenies {
				t.Errorf("%s | %s: BranchGuard alone = %q / %q, want deny with the sentinel (the verb lands in the primary checkout)", cwdCase.label, cell.name, guardDecision, guardReason)
			}
		}
	}
}
