package hook

import (
	"encoding/json"
	"testing"
)

// TestSubstituteQuotedArguments_CollapsesSpans pins the helper's own contract: a
// quoted span collapses to a single non-flag placeholder word, and unquoted text
// is untouched. The placeholder — rather than blank whitespace — is what keeps a
// quoted branch name from un-guarding its command, and the surrounding spaces
// are what stop neighbouring tokens from fusing into an accidental match.
func TestSubstituteQuotedArguments_CollapsesSpans(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		command string
		want    string
	}{
		{"no quotes untouched", "git switch main", "git switch main"},
		{"double-quoted span collapsed", `moai todo add "git switch main"`, "moai todo add  X "},
		{"single-quoted span collapsed", `moai todo add 'git switch main'`, "moai todo add  X "},
		{"apostrophe inside double quotes", `echo "it's fine"`, "echo  X "},
		{"unpaired quote left alone", `git switch main # don't`, `git switch main # don't`},
		{"tokens do not fuse across a span", `a"x"b`, "a X b"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := substituteQuotedArguments(tc.command); got != tc.want {
				t.Fatalf("substituteQuotedArguments(%q) = %q, want %q", tc.command, got, tc.want)
			}
		})
	}
}

// TestBranchStatePatterns_QuotedArgumentsAreData is the regression arm for the
// scan-scope defect: branch-state text carried inside a quoted argument is data,
// not a command, and must not match. The command actually being invoked in each
// case is `moai todo add`, `git commit`, or `echo` — none of which change branch
// state.
func TestBranchStatePatterns_QuotedArgumentsAreData(t *testing.T) {
	t.Parallel()
	cases := []string{
		`moai todo add "document why git switch main is refused"`,
		`moai todo add 'git checkout -b feat/x should be denied'`,
		`git commit -m "revert the git switch main change"`,
		`git commit -m "explain git reset --hard in the runbook"`,
		`echo "git stash"`,
		`gh pr create --body "the guard denies git rebase in the primary checkout"`,
	}
	for _, command := range cases {
		command := command
		t.Run(command, func(t *testing.T) {
			t.Parallel()
			suffix, matched := matchBranchStateCommand(command)
			if matched {
				t.Fatalf("matchBranchStateCommand(%q) = (%q, true), want false — quoted text is data, not a command", command, suffix)
			}
		})
	}
}

// TestBranchStatePatterns_RealCommandsStillDenyAroundQuotes is the falsification
// arm for the test above. Blanking quoted spans must not blunt the guard: a real
// branch-state invocation still matches, including when it sits beside or after
// a quoted argument. A helper that blanked too much would vacuously pass the
// data table above; this pins the deny side.
func TestBranchStatePatterns_RealCommandsStillDenyAroundQuotes(t *testing.T) {
	t.Parallel()
	cases := []string{
		`git switch main`,
		`git switch "main"`,
		`git checkout -b "feat/quoted-branch"`,
		`git commit -m "wip" && git switch main`,
		`git branch -D 'old-feature'`,
		`git reset --hard origin/main`,
	}
	for _, command := range cases {
		command := command
		t.Run(command, func(t *testing.T) {
			t.Parallel()
			if _, matched := matchBranchStateCommand(command); !matched {
				t.Fatalf("matchBranchStateCommand(%q) = false, want true — real branch-state command must still deny", command)
			}
		})
	}
}

// TestCheckBranchState_ExemptionAxesAreIndependent verifies each documented
// exemption axis fires ON ITS OWN when its value reaches the hook, so a failure
// observed in production is attributable to the value not arriving rather than
// to the axis being unimplemented.
//
// The axes are read from different places and a tool-spawned subagent can supply
// neither: agent_type arrives in the payload only for a main-thread
// `claude --agent` launch, and the env var is read from the hook process's own
// environment, which an `export` inside the guarded command cannot reach.
func TestCheckBranchState_ExemptionAxesAreIndependent(t *testing.T) {
	command := `{"command":"git branch -D feature/x"}`

	t.Run("AgentTypeAxis_alone_exempts", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "")
		input := &HookInput{AgentType: "manager-git", ToolInput: json.RawMessage(command)}
		if !isExemptAgent(input) {
			t.Fatalf("agent_type=manager-git alone did not exempt; the identity axis is unreachable in code, not just in the payload")
		}
	})

	t.Run("EnvAxis_alone_exempts", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "1")
		input := &HookInput{AgentType: "", ToolInput: json.RawMessage(command)}
		if !isExemptAgent(input) {
			t.Fatalf("%s=1 alone did not exempt; the env axis is unreachable in code, not just in the process environment", branchGuardExemptEnv)
		}
	})

	t.Run("NeitherAxis_leaves_command_matchable", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "")
		input := &HookInput{AgentType: "manager-develop", ToolInput: json.RawMessage(command)}
		if isExemptAgent(input) {
			t.Fatalf("neither axis supplied but isExemptAgent returned true")
		}
		if _, matched := matchBranchStateCommand("git branch -D feature/x"); !matched {
			t.Fatalf("branch-state command no longer matches; the deny path would be vacuous")
		}
	})
}
