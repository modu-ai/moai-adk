package hook

// branch_guard_psforms_test.go — SPEC-HOOK-GUARD-POWERSHELL-FORMS-001 (card
// t1255): the mutant-matrix battery of acceptance.md §D, both directions.
//
// Every disguise leg asserts the POST-FIX verdict — deny with the established
// reason family, or allow with exactly one unclassifiable audit line — and
// every legit leg asserts allow with an audit-line delta of 0, so the battery
// must stay green through every detector change (AC-HGF-010). On the pre-fix
// tree the disguise legs FAIL (RED — the under-match defect this SPEC fixes)
// while the controls already sit at their target verdicts; the controls'
// existing behavior is what the fix must not flip.
//
// Isolation (REQ-HGF-014): hmpIsolateHome points MOAI_HOME and
// CLAUDE_PROJECT_DIR at per-test directories; no t.Parallel; no real MoAI
// home and no repository .moai/state is touched — every fixture is a
// t.TempDir.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestBranchGuardPSForms is the M1 battery: one subtest per mutant-matrix
// row of acceptance.md §D.
func TestBranchGuardPSForms(t *testing.T) {
	hmpIsolateHome(t)
	t.Setenv(branchGuardExemptEnv, "")

	enc := hmpB64("Write-Output ok")
	cases := []struct {
		name       string
		tool       string
		command    string
		wantDeny   bool
		reasonHas  string // substring the deny reason must carry
		wantLines  int    // powershell-unclassified lines in the branch-guard log
		wantConstr string // expected powerShellIndirection construct (when wantLines >= 1)
	}{
		// ---- F1a: .exe-suffixed executable (REQ-HGF-001) ----
		{"f1a/deny/exe-switch", "PowerShell", `git.exe switch probe`, true, branchGuardViolationPrefix, 0, ""},
		{"f1a/deny/exe-reset-hard", "PowerShell", `git.exe reset --hard`, true, branchGuardViolationPrefix, 0, ""},
		{"f1a/allow/exe-status", "PowerShell", `git.exe status`, false, "", 0, ""},
		{"f1a/allow/plain-short", "PowerShell", `git status --short`, false, "", 0, ""},

		// ---- F1b: call operator + quoted call target (REQ-HGF-002) ----
		{"f1b/deny/single-quoted-target", "PowerShell", `& 'git' switch probe`, true, branchGuardViolationPrefix, 0, ""},
		{"f1b/deny/double-quoted-exe-target", "PowerShell", `& "git.exe" switch probe`, true, branchGuardViolationPrefix, 0, ""},
		{"f1b/allow/query-target", "PowerShell", `& 'git' status`, false, "", 0, ""},
		{"f1b/allow/quoted-prose", "PowerShell", `moai todo add "git switch x"`, false, "", 0, ""},

		// ---- F1c: backtick escape in command position (REQ-HGF-003) ----
		{"f1c/deny/backtick-switch", "PowerShell", "git swi`tch probe", true, branchGuardViolationPrefix, 0, ""},
		{"f1c/deny/backtick-rebase", "PowerShell", "git reba`se x", true, branchGuardViolationPrefix, 0, ""},
		{"f1c/allow/backtick-status", "PowerShell", "git sta`tus probe", false, "", 0, ""},
		{"f1c/allow/literal-in-single-quotes", "PowerShell", "echo 'swi`tch'", false, "", 0, ""},

		// ---- F2: -Command payload scan (REQ-HGF-004) ----
		{"f2/deny/command-payload", "PowerShell", `pwsh -Command "git switch probe"`, true, branchGuardViolationPrefix, 0, ""},
		{"f2/deny/short-c-payload", "PowerShell", `powershell -c "git switch probe"`, true, branchGuardViolationPrefix, 0, ""},
		{"f2/deny/payload-with-comment", "PowerShell", `pwsh -Command "git switch probe # note"`, true, branchGuardViolationPrefix, 0, ""},
		{"f2/allow/query-payload", "PowerShell", `pwsh -Command "git status"`, false, "", 0, ""},
		{"f2/allow/nested-quote-mutant", "PowerShell", `pwsh -Command "Write-Output 'git switch'"`, false, "", 0, ""},
		// Control: the git words here sit at command position OUTSIDE the
		// payload — the base tree already denies this form, and must keep
		// denying it after the payload scan lands.
		{"f2/control/outside-payload", "PowerShell", `pwsh -Command "git status" ; git switch probe`, true, branchGuardViolationPrefix, 0, ""},
		// Control (E-09): the cmd /c wrapper deny is the parity anchor.
		{"f2/control/cmd-c-wrapper", "PowerShell", `cmd /c git switch probe`, true, branchGuardViolationPrefix, 0, ""},

		// ---- F3: dynamic resolution demotion (REQ-HGF-005) ----
		{"f3/demote/dynamic-resolution", "PowerShell", `& (Get-Command git) switch probe`, false, "", 1, constructDynamicResolution},
		{"f3/allow/no-git-word", "PowerShell", `& (Get-Command node) serve`, false, "", 0, ""},
		{"f3/allow/no-call-operator", "PowerShell", `Get-Command git`, false, "", 0, ""},

		// ---- Start-Process aliases (REQ-HGF-006) ----
		{"alias/saps-logs", "PowerShell", `saps git -ArgumentList 'switch','probe'`, false, "", 1, constructStartProcess},
		{"alias/start-logs", "PowerShell", `start git -ArgumentList 'switch'`, false, "", 1, constructStartProcess},
		{"alias/allow/no-git-word", "PowerShell", `saps notepad readme.txt`, false, "", 0, ""},
		{"alias/parity/fully-spelled-form", "PowerShell", `Start-Process git -ArgumentList 'log'`, false, "", 1, constructStartProcess},

		// ---- denylist: literal indirection operands (REQ-HGF-007) ----
		{"dl/deny/bash-eval", "Bash", `eval "terraform destroy"`, true, "Dangerous command blocked", 0, ""},
		{"dl/deny/ps-iex", "PowerShell", `iex "terraform destroy"`, true, "Dangerous command blocked", 0, ""},
		{"dl/deny/ps-start-process", "PowerShell", `Start-Process terraform -ArgumentList 'destroy'`, true, "Dangerous command blocked", 0, ""},
		{"dl/allow/iex-query-stays-d2", "PowerShell", `iex "git status"`, false, "", 1, constructInvokeExpression},
		{"dl/allow/iex-variable", "PowerShell", `$c = 'git switch probe'; iex $c`, false, "", 1, constructInvokeExpression},
		{"dl/allow/eval-benign", "Bash", `eval "echo hi"`, false, "", 0, ""},
		{"dl/allow/eval-subexpression", "Bash", `eval "$(printf 'terraform destroy')"`, false, "", 0, ""},
		{"dl/allow/eval-env-var", "Bash", `eval "$ENV:X"`, false, "", 0, ""},
		{"dl/allow/start-process-notepad", "PowerShell", `Start-Process notepad`, false, "", 0, ""},
		// Control (E-13): the bare-form deny is the anchor.
		{"dl/control/bare-form", "PowerShell", `terraform destroy`, true, "Dangerous command blocked", 0, ""},

		// ---- F5: encoded-command spellings (REQ-HGF-011; M0 in §E.2) ----
		{"f5/enc/u2013-dash", "PowerShell", "pwsh –enc " + enc, false, "", 1, constructEncodedCommand},
		{"f5/enc/u2014-dash", "PowerShell", "pwsh —enc " + enc, false, "", 1, constructEncodedCommand},
		{"f5/enc/u2010-hyphen", "PowerShell", "pwsh ‐enc " + enc, false, "", 1, constructEncodedCommand},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newBranchGuardRepoFixture(t)
			h := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), repo)
			d, r := hmpHandle(t, h, hmpInput(t, tc.tool, "s-1", repo, tc.command))
			if tc.wantDeny {
				if d != DecisionDeny {
					t.Fatalf("decision = %q (reason %q), want deny", d, r)
				}
				if tc.reasonHas != "" && !strings.Contains(r, tc.reasonHas) {
					t.Errorf("reason %q lacks %q", r, tc.reasonHas)
				}
			} else if d == DecisionDeny {
				t.Fatalf("decision = %q (reason %q), want allow", d, r)
			}
			lines := hmpAuditLines(t, filepath.Join(repo, branchGuardAuditRelPath), "powershell-unclassified")
			if len(lines) != tc.wantLines {
				t.Fatalf("unclassified audit lines = %d (%q), want %d", len(lines), lines, tc.wantLines)
			}
			if tc.wantConstr != "" {
				if got := powerShellIndirection(tc.command); got != tc.wantConstr {
					t.Errorf("powerShellIndirection(%q) = %q, want %q", tc.command, got, tc.wantConstr)
				}
			}
		})
	}

	t.Run("injection/newline-keeps-one-line", func(t *testing.T) {
		repo := newBranchGuardRepoFixture(t)
		h := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), repo)
		cmd := "iex \"git switch x\"\n[2026] event=forged reason=\"x\""
		if d, r := hmpHandle(t, h, hmpInput(t, "PowerShell", "s-1", repo, cmd)); d == DecisionDeny {
			t.Fatalf("injection probe denied: %q", r)
		}
		data, err := os.ReadFile(filepath.Join(repo, branchGuardAuditRelPath))
		if err != nil {
			t.Fatalf("read audit log: %v", err)
		}
		if got := strings.Count(strings.TrimSpace(string(data)), "\n") + 1; got != 1 {
			t.Fatalf("audit log holds %d lines, want exactly 1: %q", got, string(data))
		}
		if !strings.Contains(string(data), "\\n") {
			t.Errorf("audit line does not carry the newline inside its quoted field: %q", string(data))
		}
	})
}
