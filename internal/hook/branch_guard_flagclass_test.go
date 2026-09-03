package hook

// branch_guard_flagclass_test.go — SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001
// M1: the permanent synthetic measurement matrix driving matchBranchStateCommand
// with the full flag-form matrix of acceptance.md §D.1 (mutation forms M-01..M-33
// → deny, query forms Q-01..Q-17 → allow), the whole-token discrimination pairs
// (P-01/P-02), and the edge cases (§D.3 E-1..E-6).
//
// Expectations are the doctrine-based TARGET (post-M2) classifications; §D.1 is
// the normative expectation authority. On the pre-fix tree every row whose
// expectation is deny and whose pre-fix behavior was allow FAILs (RED — the
// under-match defect this SPEC fixes).
//
// @MX:SPEC: SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001
// @MX:TEST: M1 measurement matrix — classification authority for the git branch matcher

import (
	"encoding/json"
	"testing"
)

// TestBranchGuardFlagClassMatrix is the M1 permanent measurement matrix.
// Cell IDs trace 1:1 to acceptance.md §D.1/§D.3 rows.
func TestBranchGuardFlagClassMatrix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		id      string
		command string
		want    bool // true = deny (pattern matches); false = allow
	}{
		// ---- Mutation forms (§D.1 M-01..M-33) → deny ----
		{"M-01", "git branch feature", true},                             // bare creation
		{"M-02", "git branch newbranch oldstart", true},                  // creation at start point
		{"M-03a", "git branch -d old", true},                             // delete
		{"M-03b", "git branch -D old", true},                             // force delete
		{"M-04a", "git branch -m renamed", true},                         // move
		{"M-04b", "git branch -M renamed", true},                         // force move
		{"M-05a", "git branch -c copied", true},                          // copy
		{"M-05b", "git branch -C copied", true},                          // force copy
		{"M-06", "git branch -f topic abc1234", true},                    // force pointer rewrite
		{"M-07", "git branch -f topic", true},                            // force create at HEAD
		{"M-08", "git branch --force topic abc1234", true},               // long force
		{"M-09", "git branch -df old", true},                             // combined cluster
		{"M-10", "git branch -fm renamed", true},                         // combined cluster
		{"M-11", "git branch -vD old", true},                             // combined cluster (mid)
		{"M-12", "git branch -u origin/main topic", true},                // set upstream
		{"M-13", "git branch --set-upstream-to=origin/main topic", true}, // attached value
		{"M-14", "git branch --set-upstream origin/main topic", true},    // removed-in-git form, fail-closed-safe pin
		{"M-15", "git branch --unset-upstream topic", true},
		{"M-16", "git branch -t topic origin/main", true}, // track
		{"M-17", "git branch --track topic origin/main", true},
		{"M-18", "git branch --no-track topic origin/main", true},
		{"M-19", "git branch --edit-description topic", true},
		{"M-20", "git branch --delete old", true},
		{"M-21a", "git branch --move renamed", true},
		{"M-21b", "git branch --copy copied", true},
		{"M-22", "git branch --set-upstream-to origin/main topic", true}, // space-separated value
		{"M-23", "git branch --track=direct topic abc1234", true},        // attached value
		{"M-24", "git branch -umain topic", true},                        // attached short value
		{"M-25", "git branch -v vbranch", true},                          // flag + name = create (measured)
		{"M-26", "git branch -- cr2", true},                              // option-prefixed creation
		{"M-27", "git branch -q qbranch", true},                          // option-prefixed creation
		{"M-28", "git branch --no-force nfbranch", true},                 // option-prefixed creation
		{"M-29", "git branch -vt vtbranch main", true},                   // cluster containing t
		{"M-30", "git branch --create-reflog crbranch", true},            // modifier + creation operand
		{"M-31", "git branch --color colprobe", true},                    // attached-only flag; space token = positional
		{"M-32", "git branch --abbrev 12 abbranch", true},                // attached-only flag; positionals
		{"M-33", "git branch -vux main x", true},                         // mid-cluster u

		// ---- Query forms (§D.1 Q-01..Q-17) → allow ----
		{"Q-01", "git branch", false},
		{"Q-02a", "git branch --list", false},
		{"Q-02b", "git branch --list develop -v", false},
		{"Q-03a", "git branch -v", false},
		{"Q-03b", "git branch -vv", false},
		{"Q-04a", "git branch -a", false},
		{"Q-04b", "git branch -r", false},
		{"Q-05", "git branch --show-current", false},
		{"Q-06", "git branch --contains HEAD", false},
		{"Q-07", "git branch --merged main", false},
		{"Q-08", "git branch --no-merged main", false},
		{"Q-09", "git branch --points-at HEAD", false},
		{"Q-10", "git branch --format %(refname)", false},
		{"Q-11", "git branch --sort=-committerdate", false},
		{"Q-12a", "git branch -q", false},
		{"Q-12b", "git branch -i", false},
		{"Q-13", "git branch --sort committerdate", false}, // space-separated value consumed
		{"Q-14", "git branch --no-contains HEAD", false},   // space-separated value consumed
		{"Q-15a", "git branch -l lpattern", false},         // short list selector
		{"Q-15b", "git branch -l foo bar", false},          // variadic list mode
		{"Q-16", "git branch --contains HEAD main", false}, // filter selector → pattern, not creation
		{"Q-17", "git branch --color=always", false},       // attached-only value consumed, no positional

		// ---- Whole-token discrimination pairs (§D.1 P-01/P-02) ----
		// P-01: --format (query) vs --force (mutation) — no prefix matching.
		{"P-01a", "git branch --format %(refname)", false},
		{"P-01b", "git branch --force topic abc1234", true},
		// P-02: embedded letters in query long flags must not trip a character class.
		{"P-02a", "git branch --contains HEAD", false},
		{"P-02b", "git branch --merged main", false},
		{"P-02c", "git branch --no-merged main", false},

		// ---- Edge cases (§D.3 E-1..E-6) ----
		{"E-1", "git branch -d old -v", true},                           // mutating flag dominates
		{"E-2", "git branch --set-upstream-to=origin/main topic", true}, // = value not a creation operand
		{"E-3a", "git branch -F topic", true},                           // case-fold of known mutation flag
		{"E-3b", "git branch --FORCE topic abc1234", true},              // case-fold of known mutation flag
		{"E-4", "git branch -D 'old-feature'", true},                    // quoted operand still denies
		{"E-5a", "git branch --dele old", false},                        // abbreviation prefix → unknown → allow
		{"E-5b", "git branch -z foo", false},                            // unknown flag → allow (fail-open)
		{"E-6", "git branch -f x y && git status", true},                // compound command segment
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.id+" "+tc.command, func(t *testing.T) {
			t.Parallel()
			suffix, matched := matchBranchStateCommand(tc.command)
			if matched != tc.want {
				t.Fatalf("matchBranchStateCommand(%q) = (%q, %v), want matched=%v (§D.1 %s)",
					tc.command, suffix, matched, tc.want, tc.id)
			}
			if matched && suffix == "" {
				t.Fatalf("matched but suffix empty for %q", tc.command)
			}
		})
	}
}

// --- M3: end-to-end extensions (checkBranchState layer, real primary fixture) ---

// TestBranchGuardFlagClass_EndToEndDenial is the M3.1 end-to-end matrix: the
// flag-class mutation forms exercised through checkBranchState in a REAL
// primary checkout (the context the guard denies in), extending the shape of
// TestBranchGuard_BranchInquiryFormsInPrimary. Every command MUST deny with
// the BRANCH_GUARD_VIOLATION sentinel and the `git branch` suffix.
//
// Non-parallel: t.Setenv mutates the process-global exemption env var.
func TestBranchGuardFlagClass_EndToEndDenial(t *testing.T) {
	requireGit(t)
	repo := newBranchGuardRepoFixture(t)
	t.Setenv(branchGuardExemptEnv, "")

	commands := []string{
		"git branch -f topic abc1234",                    // M-06 force pointer rewrite
		"git branch --force topic abc1234",               // M-08 long force
		"git branch -df old",                             // M-09 combined cluster
		"git branch -fm renamed",                         // M-10 combined cluster
		"git branch -vD old",                             // M-11 mid-cluster
		"git branch -vux main x",                         // M-33 mid-cluster u
		"git branch -u origin/main topic",                // M-12 upstream
		"git branch --set-upstream-to=origin/main topic", // M-13 attached value
		"git branch --unset-upstream topic",              // M-15
		"git branch -t topic origin/main",                // M-16 track
		"git branch --edit-description topic",            // M-19
	}
	for _, command := range commands {
		command := command
		t.Run(command, func(t *testing.T) {
			input := &HookInput{
				ToolName:  "Bash",
				CWD:       repo,
				ToolInput: json.RawMessage(`{"command": "` + command + `"}`),
			}
			decision, reason := checkBranchState(input, repo)
			if decision != DecisionDeny {
				t.Fatalf("checkBranchState(%q) decision = %q, want %q", command, decision, DecisionDeny)
			}
			const wantPrefix = "BRANCH_GUARD_VIOLATION: git branch"
			if len(reason) < len(wantPrefix) || reason[:len(wantPrefix)] != wantPrefix {
				t.Fatalf("checkBranchState(%q) reason = %q, want prefix %q", command, reason, wantPrefix)
			}
		})
	}
}

// TestBranchGuardFlagClass_QueryAllowlistEndToEnd is the M3.2 query-allowlist
// regression, exercised end-to-end at the checkBranchState layer in the same
// real primary fixture: the value-taking query flags (and the filter-mode
// positional) MUST stay allowed alongside the existing pins in
// branch_guard_test.go (AC-WBG-F-003 no-over-match-regression arm). Includes
// the N3 run-gate debt case `git branch -al <pattern>`: `-l` anywhere in a
// short cluster selects list/filter mode, so the trailing positional is a
// pattern, not a creation operand.
//
// Non-parallel: t.Setenv mutates the process-global exemption env var.
func TestBranchGuardFlagClass_QueryAllowlistEndToEnd(t *testing.T) {
	requireGit(t)
	repo := newBranchGuardRepoFixture(t)
	t.Setenv(branchGuardExemptEnv, "")

	commands := []string{
		"git branch --merged main",        // Q-07
		"git branch --no-merged main",     // Q-08
		"git branch --points-at HEAD",     // Q-09
		"git branch --format %(refname)",  // Q-10
		"git branch --sort committerdate", // Q-13 space-separated value
		"git branch --list develop -v",    // Q-02b (re-pinned end-to-end)
		"git branch --contains HEAD main", // Q-16 filter-mode positional
		"git branch -al develop",          // N3: -l mid-cluster selects list mode
	}
	for _, command := range commands {
		command := command
		t.Run(command, func(t *testing.T) {
			input := &HookInput{
				ToolName:  "Bash",
				CWD:       repo,
				ToolInput: json.RawMessage(`{"command": "` + command + `"}`),
			}
			decision, reason := checkBranchState(input, repo)
			if decision != "" || reason != "" {
				t.Fatalf("checkBranchState(%q) = (%q, %q), want (\"\", \"\") — query form must stay allowed", command, decision, reason)
			}
		})
	}
}

// TestBranchGuardFlagClass_SubagentNegativePath is the D7 (audit iteration 1)
// negative-path condition test for REQ-WBG-F-008 / AC-WBG-F-008. It pins the
// guard LOGIC only and does NOT prove the PreToolUse payload shape.
//
// Documented conditions (REQ-WBG-F-008) — both exemption axes are unreachable
// from a tool-spawned subagent:
//
//   - Env axis (uncontested): MOAI_BRANCH_GUARD_EXEMPT is read from the hook
//     process's own environment, spawned BEFORE the guarded command runs, so
//     exporting it inside the command is a no-op. This test holds it unset,
//     which is the subagent-shaped condition.
//   - AgentType axis (CONTESTED — left contested by this card, see below):
//     branch_guard.go:30-33 holds that agent_type arrives only for a
//     main-thread `claude --agent <name>` launch, while
//     .claude/rules/moai/core/hooks-system.md:114 states all hook events
//     include agent_type when triggered from a subagent context. This test
//     supplies a subagent-SHAPED HookInput (AgentType zero-valued) and pins
//     that the deny stands for it — an assertion about isExemptAgent's logic,
//     NOT about which shape the runtime actually delivers.
//
// CONTESTED-AXIS CAPTURE OUTCOME (D12, audit iteration 2): the mandated
// capture of one real tool-spawned PreToolUse payload is IMPRACTICABLE in this
// card's environment — the nested `claude -p` probe is refused by the runtime
// worktree-session guard (verbatim refusal observed twice: "this command runs
// claude ... in a plain command, so what it runs cannot be shown not to be
// git"), and the hook's own trace logging (internal/hook/trace) records no
// agent_type/agent_id field, so no existing log can decide the axis either.
// The axis therefore REMAINS CONTESTED; a future capture contradicting the
// guard's reading routes to a doc-reconciliation blocker report, never a
// silent re-classification.
//
// Non-parallel: t.Setenv mutates the process-global exemption env var.
func TestBranchGuardFlagClass_SubagentNegativePath(t *testing.T) {
	requireGit(t)
	repo := newBranchGuardRepoFixture(t)
	t.Setenv(branchGuardExemptEnv, "")

	input := &HookInput{
		ToolName:  "Bash",
		CWD:       repo,
		ToolInput: json.RawMessage(`{"command": "git branch -f x y"}`),
		// AgentType deliberately left zero-valued: subagent-shaped payload.
	}
	decision, reason := checkBranchState(input, repo)
	if decision != DecisionDeny {
		t.Fatalf("checkBranchState(subagent-shaped git branch -f) decision = %q, want %q — deny stands when neither exemption axis fires", decision, DecisionDeny)
	}
	const wantPrefix = "BRANCH_GUARD_VIOLATION: git branch"
	if len(reason) < len(wantPrefix) || reason[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("checkBranchState(subagent-shaped git branch -f) reason = %q, want prefix %q", reason, wantPrefix)
	}
}
