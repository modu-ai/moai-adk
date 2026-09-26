---
id: SPEC-HOOK-GUARD-POWERSHELL-FORMS-001
title: "Acceptance — Classify the PowerShell tool-call forms the shell guards pass without record (t1224 debts F1-F5)"
version: "0.1.0"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
tier: M
---

# acceptance.md — SPEC-HOOK-GUARD-POWERSHELL-FORMS-001

## §A Scope of verification

14 criteria: AC-HGF-001..012 release-blocking (two-cell), AC-HGF-013..014 regression-guard. The RED-now evidence carrier is the ledger in §B (per `.claude/rules/moai/development/verification-completeness.md` §2.1 — command, verbatim stdout, exit code, tree SHA as separate fields; the ledger is the recommended carrier).

## §B Evidence ledger (measured on this tree, card base `19b5321c1`, 2026-09-26)

**Command** (single invocation, no pipes/chains):
```
go test -count=1 -overlay /tmp/t1255-overlay.json ./internal/hook/ -run TestAuditProbe -v -o /dev/null
```
with `/tmp/t1255-overlay.json` = `{"Replace":{"<worktree>/internal/hook/zz_audit_probe_test.go":"<worktree>/.moai/reports/t1224/followup-probe_test.go.txt"}}` (overlay content is one Replace entry; the probe file is committed under `.moai/reports/t1224/` in this tree).

**Exit code**: `0` (`ok github.com/modu-ai/moai-adk/internal/hook 6.197s`).
**Tree SHA**: `19b5321c1` (short SHA of the clean card base; `git branch --show-current` = `WT-powershell-guard-debt`).

Verbatim stdout rows (log prefixes preserved as captured; `zz_audit_probe_test.go` is the overlay's target name):

| ID | Verbatim row |
|---|---|
| E-01 | `zz_audit_probe_test.go:48: BRANCH "git.exe switch probe"                        -> decision="allow" unclassifiedLines=0 construct=""` |
| E-02 | `zz_audit_probe_test.go:48: BRANCH "& 'git' switch probe"                        -> decision="allow" unclassifiedLines=0 construct=""` |
| E-03 | `zz_audit_probe_test.go:48: BRANCH "& \"git.exe\" switch probe"               -> decision="allow" unclassifiedLines=0 construct=""` |
| E-04 | `zz_audit_probe_test.go:48: BRANCH "git swi`tch probe"                           -> decision="allow" unclassifiedLines=0 construct=""` |
| E-05 | `zz_audit_probe_test.go:48: BRANCH "pwsh -Command \"git switch probe\""          -> decision="allow" unclassifiedLines=0 construct=""` |
| E-06 | `zz_audit_probe_test.go:48: BRANCH "powershell -c \"git switch probe\""       -> decision="allow" unclassifiedLines=0 construct=""` |
| E-07 | `zz_audit_probe_test.go:48: BRANCH "& (Get-Command git) switch probe"            -> decision="allow" unclassifiedLines=0 construct=""` |
| E-08 | `zz_audit_probe_test.go:48: BRANCH "saps git -ArgumentList 'switch','probe'"     -> decision="allow" unclassifiedLines=0 construct=""` |
| E-09 | `zz_audit_probe_test.go:48: BRANCH "cmd /c git switch probe"                     -> decision="deny" unclassifiedLines=0 construct=""` (control — must stay) |
| E-10 | `zz_audit_probe_test.go:48: BRANCH "$c = 'git switch probe'; iex $c"             -> decision="allow" unclassifiedLines=1 construct="invoke-expression"` (D2 baseline — must stay) |
| E-11 | `zz_audit_probe_test.go:27: DENYLIST Bash       "eval \"terraform destroy\""                       -> "allow" ""` |
| E-12 | `zz_audit_probe_test.go:27: DENYLIST PowerShell "iex \"terraform destroy\""                        -> "allow" ""` |
| E-13 | `zz_audit_probe_test.go:27: DENYLIST PowerShell "terraform destroy"                                -> "deny" "Dangerous command blocked: (?i)terraform\\s+destroy"` (bare control — must stay) |
| E-14 | `zz_audit_probe_test.go:27: DENYLIST PowerShell "Start-Process terraform -ArgumentList 'destroy'"  -> "allow" ""` |
| E-15 | `zz_audit_probe_test.go:58: ILOCK "iex \"git status\""                     -> decision="allow" unclassifiedLines=1` (F4 behavior evidence) |
| E-16 | `zz_audit_probe_test.go:58: ILOCK "Start-Process git -ArgumentList 'log'"  -> decision="allow" unclassifiedLines=1` (F4 behavior evidence) |
| E-17 | F4 current comment (`sed -n '29,32p' internal/hook/powershell_indirection.go`, exit 0, tree `19b5321c1`): `// integrationLockAuditRelPath is the integration lock's audit log, relative to` / `// the handler's project root. It holds only unclassifiable-command lines; the` / `// guard's other fail-open paths keep writing their stderr advisories.` |

## §C Acceptance criteria (two-cell)

| AC | Requirement | RED-now (cell 1) | Green path (cell 2) |
|---|---|---|---|
| AC-HGF-001 | REQ-HGF-001 | E-01: `git.exe switch probe` allows today. Red because the branch-state scan does not accept the `.exe`-suffixed executable spelling — not because the command is benign (control E-09 proves the scan denies the plain form). | M2 flips it: committed test asserts deny with the `BRANCH_GUARD_VIOLATION:` family; scoped selector `go test -count=1 ./internal/hook/ -run TestBranchGuardPSForms` exits 0. |
| AC-HGF-002 | REQ-HGF-002 | E-02 + E-03: call-operator + quoted-target forms allow today (quote collapse blanks the target). | M2: both forms deny; `& 'git' status` stays allow (mutant matrix D-legit). |
| AC-HGF-003 | REQ-HGF-003 | E-04: backtick-split subcommand allows today. | M2: ``git swi`tch probe`` reaches the same verdict as `git switch probe` (deny); ``git sta`tus probe`` stays allow. |
| AC-HGF-004 | REQ-HGF-004 | E-05 + E-06: `-Command`/`-c` payloads allow today. | M2: both deny; `pwsh -Command "git status"` stays allow; nested-quote mutant `pwsh -Command "Write-Output 'git switch'"` stays allow; control E-09 unchanged. |
| AC-HGF-005 | REQ-HGF-005 | E-07: dynamic-resolution form allows with zero audit lines today. | M3: same command allows with exactly one unclassifiable line (construct recorded); `& (Get-Command node) serve` stays allow with zero lines. |
| AC-HGF-006 | REQ-HGF-006 | E-08: `saps` alias allows with zero audit lines today. | M3: `saps git -ArgumentList 'switch','probe'` logs exactly one `start-process` line — byte-shape identical to the fully spelled form's line; `saps notepad readme.txt` stays allow with zero lines. |
| AC-HGF-007 | REQ-HGF-007 | E-12: `iex "terraform destroy"` allows today (bare control E-13 denies). | M4: denies with the E-13 reason family; `iex "git status"` stays allow + one D2 line; `iex $c` stays allow. |
| AC-HGF-008 | REQ-HGF-007 | E-14: `Start-Process terraform -ArgumentList 'destroy'` allows today. | M4: denies with the E-13 reason family; `Start-Process git -ArgumentList 'log'` stays allow + one D2 line. |
| AC-HGF-009 | REQ-HGF-007 | E-11: Bash `eval "terraform destroy"` allows today. | M4: denies with the E-13 reason family; `eval "echo hi"` and `eval "$(printf 'terraform destroy')"` (non-literal) stay allow. |
| AC-HGF-010 | REQ-HGF-008 + REQ-HGF-014 | Healthy battery RED-side is structural: at `19b5321c1` the battery passes but has no committed, named test — a regression could land unnoticed (absence of a failure signal is not a guard). | M2-M4 deliver the battery as a committed table-driven test: every §D legit leg asserts allow AND zero new audit lines; it must stay green through every detector change. The test file carries the REQ-HGF-014 isolation contract (each test sets `MOAI_HOME` to a per-test temp dir, no `t.Parallel`, no real MoAI home or repo `.moai/state` access), verified by the same scoped run. |
| AC-HGF-011 | REQ-HGF-009 | E-17 + E-15/E-16: the comment claims a narrower scope than the measured logging behavior. | M5: the comment states the measured scope (every unclassifiable PowerShell command under a live foreign hold, merge-shaped or not, is recorded); verbatim comment quoted in progress.md §E.2. |
| AC-HGF-012 | REQ-HGF-010 + REQ-HGF-011 | F5 is unmeasured today: no accepted-spelling record exists for `-EncodedCommand` (verdict Gap, line 47), and the detector's coverage of `/`-prefixed and U+2013 forms is unknown. | M0 produces the accepted-set table under the declared caps, recorded in progress.md §E.2; M5 pins the detector per REQ-HGF-011 (measured-accepted → detected, or documented superset with the over-match rationale; measured-rejected → not required). pwsh absent → INCONCLUSIVE record names the gap; detector unchanged. |
| AC-HGF-013 (regression-guard) | REQ-HGF-013 | Already green at `19b5321c1` (INJECTION probe row: newline command yields exactly one audit line, command verbatim in its quoted field). Classification: regression-guard — its RED cannot be red by construction. | Committed test keeps the injection row green through M2-M4; any new audit-line path reproduces the one-line contract. |
| AC-HGF-014 (regression-guard) | REQ-HGF-012 | Trivially satisfied at the base (no card commits yet); classification: regression-guard over the card's commit range. | At close: `git diff --name-only <base>..HEAD -- internal/template/ .claude/` is empty — the verifiable form of the "Go code only" template assessment. |

## §D Mutant matrix (both-way — card-mandated)

Disguise leg = the form must be caught (deny) or recorded (one audit line). Legit leg = the same shape carrying no guarded pattern must not change. Every row is one table-driven test case in the M1 test file.

| Family | Disguise leg (expected after fix) | Legit leg (must stay allow, audit-line delta 0 unless stated) |
|---|---|---|
| F1a exe-suffix | `git.exe switch probe` → deny; `git.exe reset --hard` → deny | `git.exe status` → allow; `git status --short` → allow |
| F1b quoted call target | `& 'git' switch probe` → deny; `& "git.exe" switch probe` → deny | `& 'git' status` → allow; `moai todo add "git switch x"` → allow (quoted prose) |
| F1c backtick | ``git swi`tch probe`` → deny; ``git reba`se x`` → deny | ``git sta`tus probe`` → allow; `echo 'swi`tch'` (single-quoted literal) → allow |
| F2 -Command payload | `pwsh -Command "git switch probe"` → deny; `powershell -c "git switch probe"` → deny; `pwsh -Command "git switch probe # note"` → deny | `pwsh -Command "git status"` → allow; `pwsh -Command "Write-Output 'git switch'"` → allow (nested-quote mutant); `pwsh -Command "git status" ; git switch probe` (outside payload) → deny |
| F3 dynamic | `& (Get-Command git) switch probe` → allow + exactly 1 unclassifiable line | `& (Get-Command node) serve` → allow, 0 lines; `Get-Command git` alone → allow, 0 lines |
| alias | `saps git -ArgumentList 'switch','probe'` → allow + 1 `start-process` line; `start git -ArgumentList 'switch'` → allow + 1 line | `saps notepad readme.txt` → allow, 0 lines; `Start-Process git -ArgumentList 'log'` → allow + 1 line (parity) |
| denylist iex | `iex "terraform destroy"` → deny (E-13 reason family) | `iex "git status"` → allow + 1 D2 line; `iex $c` → allow (+ existing D2 line) |
| denylist Start-Process | `Start-Process terraform -ArgumentList 'destroy'` → deny | `Start-Process git -ArgumentList 'log'` → allow + 1 D2 line; `Start-Process notepad` → allow |
| denylist eval | `eval "terraform destroy"` → deny | `eval "echo hi"` → allow; `eval "$(printf 'terraform destroy')"` → allow (non-literal fail-open); `eval "$ENV:X"` → allow |
| controls (unchanged) | `cmd /c git switch probe` → deny (E-09); bare `terraform destroy` → deny (E-13) | `git status` in a worktree fixture → allow; `iex "git status"` → allow + 1 line (E-10) |

Note: run-phase may adjust the literal spelling of a disguise case (e.g. a different backtick-split subcommand or alias argument shape) as long as both legs of the row keep their verdicts and the change is recorded in progress.md.

## §E F5 measurement plan (M0)

- **Location**: scratch directory under `/tmp`, outside every worktree — the worktree session guard refuses `pwsh`; the measurement never runs inside one.
- **Recipe** (verbatim from `.moai/reports/t1224/verdict.md` lines 100-103):
```
B64=$(printf 'Write-Output ok' | iconv -t UTF-16LE | base64)
for s in -e -ec -en -enc -enco -encodedc -EncodedCommand -ENC --EncodedCommand /enc –enc; do
  printf '%s -> ' "$s"; pwsh -NoProfile -NonInteractive $s "$B64" 2>&1 | head -1
done
```
  (`–enc` is written with the Unicode en-dash U+2013; the loop body is the recipe's own — the outer run wraps each iteration in `timeout 60`.)
- **Caps declared up front** (t1152 lesson): 12 candidate spellings, one pass, wall-clock ≤ 10 min, ≤ 2 turns, foreground only, per-invocation `timeout 60`, no background load.
- **Accepted set**: a spelling is accepted iff pwsh prints `ok`. The table (spelling → accepted/rejected) is recorded in progress.md §E.2 and compared against `isEncodedCommandParameter` behavior per spelling.
- **Pin rule** (REQ-HGF-011): every accepted spelling must be detected, or the detector keeps a documented superset citing the over-match policy (one audit line, never a deny); rejected spellings impose nothing. pwsh unavailable → INCONCLUSIVE record, no detector change.

## §F Quality gates and Definition of Done

- All release-blocking ACs (001-012) PASS with two-cell evidence; regression-guards (013-014) verified at close.
- `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/hook/ -run 'TestBranchGuardPSForms'` → exit 0 (plus the pre-existing family selectors it must not break: `-run 'TestHMP|TestBranchGuard'`).
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run internal/hook/...` no NEW issues.
- Template diff empty over the card range (AC-HGF-014); F5 record present (AC-HGF-012); F4 comment aligned (AC-HGF-011).
- Every commit carries the card id `t1255`; RED outputs (E8) captured before GREEN.

## §G Edge cases

- Unicode dash U+2013 vs ASCII hyphen in parameter prefixes — byte-level distinction, M0 decides.
- Nested quoting inside a `-Command` payload (the `Write-Output 'git switch'` mutant) — the in-payload collapse is what keeps it allow.
- `start` alias over-match — bounded by the git-word gate; any `start` without a git word stays silent.
- Newline injection into audit lines — one-line contract (AC-HGF-013).
- Absorbed-base drift — line-number citations re-measured after `git merge develop` (plan §C).
- `cmd /c` control and bare-deny controls must never flip — they are the parity anchors the whole matrix leans on.
