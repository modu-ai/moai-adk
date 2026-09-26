# Progress — SPEC-HOOK-MATCHER-POWERSHELL-001

## §E.1 Plan-phase Audit-Ready Signal

- Card t1224. Artifacts (Tier M): spec.md, plan.md, acceptance.md, progress.md. Status: draft. Version 0.2.0 (plan-audit iter-1 FAIL 0.77 resolved; report `.moai/reports/plan-audit/SPEC-HOOK-MATCHER-POWERSHELL-001-review-1.md`).
- Requirements: 16 GEARS (REQ-HMP-001..016; REQ-HMP-002 redefined to the pre-tool wrapper in 0.2.0). ACs: 13 matrix rows, 13 Given-When-Then scenarios.
- SPEC ID self-check: `SPEC-HOOK-MATCHER-POWERSHELL-001` → PASS; uniqueness: 0 existing directories at authoring.
- Base: `efc807b81`; local develop at `35ab8cff3` when 0.2.0 was written; M0 absorbs.
- Open decisions: D1 matcher form (operator), D2 unclassifiable-command policy with per-guard log destinations (operator), D3 PostToolUse dead branches (lead), D4 sibling cards (operator), D5 LIVE arm C with binary provenance (lead), D6 wrapper Risk-Amplifier on PowerShell (operator) — plan.md § Open Decisions.
- Unverified premise: PowerShell hook payload carries `tool_input.command` and a Bash-shaped `tool_response`. Transcript-level t1211 evidence supports it; hook-level capture runs in M0 (arms A/B) before M1.
- Collision state: t1152 (SPEC-HOOK-STDIN-FAILCLOSED-001) landed in develop, `status: completed`; REQ-HMP-009 defers to its fail-closed answer. t1099 (`WT-dual-harness-parity-rebuild` `e722a1493`) edits `internal/cli/hook.go`, `hook_test.go`, `internal/codexadapter/**`, `internal/template/obligations*`; it does not edit `internal/hook/*`, the settings template, or `.claude/hooks/*` (measured 2026-09-26).
- Awaiting plan-audit iter-2 and Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

Card t1224, cycle_type=tdd, manager-develop. Operator Implementation Kickoff Approval granted; decisions D1=A, D2=A, D6=A applied as resolved in spec.md 0.2.2. All test commands ran env-scrubbed (`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && …`) inside `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1224`. Raw logs sit under `.moai/reports/t1224/` (gitignored, machine-local); the deciding lines are carried below.

### M0 — absorption and re-measure

- Pre-merge HEAD `d2d89320c`. `git merge-base --is-ancestor efc807b81 develop; echo $?` → `0` (local develop contains t1211's `efc807b81`). Local develop was `553e224f3`; `git merge develop` → no conflicts; merge commit `6636e696d` (parents `d2d89320c 553e224f3`). `BASE = git merge-base develop HEAD` → `553e224f33ea0de095afa12b7b59e1e6a6102216`.
- t1099 collision re-check: t1099 branch is now `WT-dual-harness-parity-rebuild` at `d96189e42` (was `e722a1493` at plan time). `git diff --name-only develop...WT-dual-harness-parity-rebuild | grep -E 'internal/hook/|settings.json|\.claude/hooks|cli/hook.go'` → `internal/cli/hook.go` only; its hunks there sit in `init`, `runHookEvent`, and `runHarnessObserveStop` (hunk `@@ -823,8 +858,15 @@` rewrites the `runHarnessObserveStop` body), not in `runHarnessObserve`. No overlap with the one line this card changes (`hook.go:822`). `git grep '"Bash"\|"PowerShell"\|shell' WT-dual-harness-parity-rebuild -- internal/codexadapter` → 0 lines; site 13 stays class (c). Branch `WT-dual-harness-recovery` (`8dfdab19f`) touches none of the four paths.
- Sites re-measured on the merged tree (spec §A.3 line numbers on base `efc807b81` → merged tree): pre_tool.go 451/534/552/569 (unchanged), post_tool.go 238/248, evidence_writer.go 184/469/650, internal/cli/hook.go `822` (develop number, as spec §C predicted), wrapper tmpl `:44`, local wrapper `:39`, settings template/local `"matcher": "Write|Edit|Bash"` at line 58. `internal/permission/stack.go:460` (site 11) unchanged, out of scope.
- Indirection classification (fixes the REQ-HMP-010 set) — `TestHMPIndirectionClassification` on the unchanged matchers:
  ```
  "& git switch x"                                   branch-guard=true integration-lock=false
  "& git merge --no-ff WT-x"                         branch-guard=true integration-lock=true
  "iex \"git switch x\""                             branch-guard=false integration-lock=false
  "Invoke-Expression 'git merge --no-ff WT-x'"       branch-guard=false integration-lock=false
  "Start-Process git -ArgumentList 'switch','x'"     branch-guard=false integration-lock=false
  "Start-Process git -ArgumentList 'merge','WT-x'"   branch-guard=false integration-lock=false
  "pwsh -EncodedCommand ZwBpAHQAIABzAHcAaQB0AGMAaAAgAHgA" branch-guard=false integration-lock=false
  ```
  So the call operator is already classified (REQ-HMP-007/008 path); Invoke-Expression/iex, Start-Process and -EncodedCommand are the REQ-HMP-010 set.
- Bash controls for REQ-HMP-009 (`TestHMPDefectiveToolInput`, pre-tool with branch guard + integration lock + slot lease enabled; post-tool default handler), identical for all five `tool_input` values `"x"`, `[1]`, `null`, `{}`, `{"command": 42}`:
  ```
  Bash control: pre={"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}} post={"hookSpecificOutput":{"hookEventName":"PostToolUse"}}
  ```
- Encoded-command spellings with pwsh: **REFUSED** (see Gaps). `pwsh` is at `/usr/local/bin/pwsh`; `claude --version` → `2.1.283 (Claude Code)`; the pwsh version probe itself was refused.
- LIVE arms A/B: **REFUSED** (see Gaps). Scratch directories and settings for both arms were created under `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/aa4a8059-2c32-4b20-bdde-000da988d1ca/scratchpad/t1224-live/{A,B}/` (arm B adds a PostToolUse logger to capture `tool_response` keys); nothing ran. Arm B is therefore INCONCLUSIVE and M1–M3 proceeded on the REQ-HMP-009 fail-safe design with no live-path claim for REQ-HMP-006/011 (plan M0 / AC-HMP-012 branch).

### RED evidence (captured before GREEN)

- M1 — `go test -count=1 ./internal/template/ -run 'TestHMP'` → exit 1:
  ```
  --- FAIL: TestHMPMatcherParityTemplate (0.00s)
          hmp_powershell_matcher_test.go:177: Bash registration has no PowerShell counterpart: PreToolUse .claude/hooks/moai/handle-pre-tool.sh
  --- FAIL: TestHMPMatcherParityLocal (0.00s)
      hmp_powershell_matcher_test.go:214: no PreToolUse registration of .claude/hooks/moai/handle-pre-tool.sh names PowerShell: ["Agent|Task" "AskUserQuestion" "SendMessage|TaskStop" "Write|Edit|Bash"]
  --- FAIL: TestHMPWrapperScopeComment (0.00s)
  ```
  (TestHMPMatcherParityTemplate failed once per platform darwin/linux/windows. `TestHMPWrapperRiskWarningBashOnly` passed on base: the wrapper was already Bash-only, which D6=A keeps — it is a pin, not a RED-then-GREEN criterion.)
- M2 — `go test -count=1 -v ./internal/hook/ -run 'TestHMP'` → exit 1, `shell-tool literal hits: 10` (cli/hook.go:822, evidence_writer.go:184/469/650, post_tool.go:238/248, pre_tool.go:451/534/552/569), each reported as `shell-tool literal outside internal/hook/shell_tool.go`. The wrapper hit was matched by the D6 exclusion (no stale-exclusion failure, so the scan saw the line).
- M3/M4 — same package run (M4 detector stubbed to return `""` so everything compiled): `--- FAIL:` TestHMPDenyAskNoVerifyParity (all 5 cases), TestHMPBranchGuardParity, TestHMPIntegrationLockParity, TestHMPSlotLeaseParity, TestHMPEvidenceParity, TestHMPEvidenceWritePaths/PowerShell, TestHMPPowerShellIndirectionDetector, TestHMPUnclassifiedBranchGuard (4 constructs + call-operator + deny-list subtests), TestHMPUnclassifiedIntegrationLock (2 constructs); `internal/cli` → `--- FAIL: TestHMPHarnessObservePowerShellEvidence/PowerShell` with `--- PASS: …/Bash`. Passing on base by design: TestHMPDefectiveToolInput (both tools allow), TestHMPTruncatedStdinFailClosed (tool-independent), TestHMPIndirectionClassification (the measurement).

### Commits (branch `WT-hook-matcher-powershell`, unpushed)

| SHA | Subject |
|---|---|
| `6636e696d` | merge(t1224): absorb local develop 553e224f3 before run (card t1224) |
| `02463b5ff` | feat(SPEC-HOOK-MATCHER-POWERSHELL-001): M1 register the pre-tool hook for PowerShell (card t1224) — carries `status: draft → in-progress` |
| `ec5285014` | feat(SPEC-HOOK-MATCHER-POWERSHELL-001): M2-M3 route shell-tool branches through IsShellTool (card t1224) |
| `2006d3cd3` | feat(SPEC-HOOK-MATCHER-POWERSHELL-001): M4 allow and log unclassifiable PowerShell indirection (card t1224) |

Pre-existing tests modified by this card: none (every new test is in a new file), so the AC-HMP-011 "listed modified tests" set is empty.

### AC matrix (HEAD `2006d3cd3`)

Verification command for every TestHMP row: `go test -count=1 ./internal/hook/ ./internal/cli/ ./internal/template/ -run 'TestHMP' -v` → exit 0, 23 `--- PASS: TestHMP*` top-level lines, `ok github.com/modu-ai/moai-adk/internal/hook 8.589s`, `ok …/internal/cli 0.807s`, `ok …/internal/template 0.751s`.

| AC | Status | Evidence |
|---|---|---|
| AC-HMP-001 | PASS | TestHMPMatcherParityTemplate (darwin/linux/windows), TestHMPMatcherParityLocal (template and local PreToolUse matcher sets for handle-pre-tool.sh equal, one names PowerShell). `make build` → exit 0, last line `go build -ldflags "… Commit=6636e696d …" -o bin/moai ./cmd/moai` (built on the M1–M4 working tree before commit; template bytes unchanged since) |
| AC-HMP-002 | PASS | `git diff --name-only 553e224f3..HEAD \| wc -l` → `22` (non-empty control); `git diff -U0 553e224f3..HEAD -- internal/template/templates/.claude/settings.json.tmpl .claude/settings.json \| grep '^@@'` → `@@ -64,0 +65,11 @@` and `@@ -63,0 +64,11 @@` — pure 11-line insertions after the `Write\|Edit\|Bash` block inside the PreToolUse array; no PostToolUse or `permissions` line changed |
| AC-HMP-003 | PASS | RED above; GREEN TestHMPMatcherParityTemplate; TestHMPMatcherParityMutants: m1 (PowerShell registration removed) and m2 (synthetic fixture, PowerShell on a second script) each name `no PowerShell counterpart: PreToolUse .claude/hooks/moai/handle-pre-tool.sh`; stale-exclusion subtest reports |
| AC-HMP-004 | PASS | RED 10 Go hits + wrapper hit (above); HEAD `shell-tool literal hits: 1` (inside `internal/hook/shell_tool.go`); TestHMPSourceGuardPositiveControls flags `const x = "Bash"` and `strings.EqualFold(name, "bash")`, ignores a comment, reports a stale exclusion; wrapper positive control flags an unexcluded Bash-only match |
| AC-HMP-005 | PASS | TestHMPDenyAskNoVerifyParity: `terraform destroy`, `Remove-Item -Recurse -Force C:\`, `git commit --no-verify -m x` (deny), `git push --force origin feature/x`, `git reset --hard HEAD~1` (ask) — PowerShell (decision, reason) equals Bash for each; RED before conversion |
| AC-HMP-006 | PASS | TestHMPBranchGuardParity: primary-checkout PowerShell `git switch -c probe` → deny, reason prefix `BRANCH_GUARD_VIOLATION:`, equal to the Bash reason; same payload with cwd in a linked worktree → not denied |
| AC-HMP-007 | PASS | TestHMPIntegrationLockParity (`git merge --no-ff WT-x`, live foreign hold) and TestHMPSlotLeaseParity (`go test ./internal/cli/...`, resource regex `\bgo\s+test\b`, live foreign lease) — PowerShell deny with the Bash reason; both RED before conversion |
| AC-HMP-008 | PASS | TestHMPDefectiveToolInput: five wrongly-typed inputs, pre and post outputs equal to the Bash control, no panic; TestHMPTruncatedStdinFailClosed: truncated pre-tool stdin naming `PowerShell` and `Bash` both return the SPEC-HOOK-STDIN-FAILCLOSED-001 deny containing `fail-closed`, byte-identical stdout |
| AC-HMP-009 | PASS | TestHMPUnclassifiedBranchGuard: `iex "git switch probe"`, `Start-Process git …`, `pwsh -EncodedCommand <b64 "git status">`, `powershell -enc <b64>` → allowed, exactly one `powershell-unclassified` line in `.moai/logs/branch-guard-audit.log`, reason field contains `unclassifiable`, no decoded payload in the line; Bash `eval "git switch probe"` → allowed, no line; `& git switch probe` → deny; linked worktree → no line; `iex terraform destroy` → deny from the destructive list. TestHMPUnclassifiedIntegrationLock: `iex "git merge --no-ff WT-x"`, `pwsh -enc <b64>` → allowed, one line in `.moai/logs/integration-lock-audit.log`; unheld window → no line; Bash eval control → no log file |
| AC-HMP-010 | PASS | TestHMPEvidenceParity: PowerShell `go test ./x/...` record equals Bash (ok, pass, fail, outcome, zero-execution); response shapes `{"result": 7}`, `[1, 2]`, `42` record no pass; zero-execution advisory identical. TestHMPEvidenceWritePaths: LogBashEvidence and the post-tool handler each write one record for PowerShell as for Bash. TestHMPHarnessObservePowerShellEvidence (CLI site 10) |
| AC-HMP-011 | PASS | TestHMPTestIsolation → `TestHMP functions checked: 23`, 0 violations; every TestHMP* calls `hmpIsolateHome` (helper body verified to contain `t.Setenv("MOAI_HOME", t.TempDir())` in all three packages); positive control `TestHMPBad` with `t.Parallel()` → 2 violations on 1 function |
| AC-HMP-012 | INCONCLUSIVE | Validity condition failed before any arm ran: the worktree-isolation guard refused both `claude -p … --tools PowerShell` (arm A) and `pwsh` (see Gaps). No payload captured, no LIVE claim made; `.moai/reports/t1224/live/` holds no arm output |
| AC-HMP-013 | PASS | TestHMPWrapperRiskWarningBashOnly (rendered template and local copy): PowerShell `a \| b \| c \| d \| e \| f \| g` → no `[moai:bash-risk]` on stderr or in the log; Bash control → warning present. TestHMPWrapperScopeComment: both copies contain `matchers "Write\|Edit\|Bash" and "PowerShell"` |

### Quality gates (HEAD `2006d3cd3` unless noted)

- `go test -count=1 ./internal/hook/...` → exit 0 (`ok …/internal/hook 443.090s` plus all 10 sub-packages ok) — run on the M1–M4 working tree before commit.
- `go test -count=1 ./internal/template/...` → exit 0 (`ok …/internal/template 70.910s`, agentemit, commandemit ok) — same tree.
- `go test -count=1 ./internal/cli/ -run 'Hook|Harness|HMP|StdinFailClosed|Settings|PreTool|Evidence'` → exit 0 (`ok …/internal/cli 80.561s`) — same tree.
- `go vet ./internal/hook/... ./internal/cli/... ./internal/template/...` → exit 0.
- `golangci-lint run ./internal/hook/... ./internal/template/... ./internal/cli/...` → exit 0, `0 issues.` (re-run on HEAD).
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `gofmt -l internal/hook internal/cli internal/template` → empty after formatting `hmp_source_guard_test.go`.
- Coverage of the new code under `go test -run TestHMP -coverprofile ./internal/hook/`: `IsShellTool` 100%, `isPowerShellTool` 100%, `powerShellIndirection` 100%, `encodedCommandSegment` 100%, `isEncodedCommandParameter` 100%, `isScriptParameter` 66.7%, `appendUnclassifiedAudit` 66.7% (uncovered: mkdir/open/write failure branches), `recordBranchGuardUnclassified` 70.0%.

### Gaps

- **Worktree-isolation guard refusals (the LIVE measurements).** Verbatim: `pwsh -NoProfile -NonInteractive -Command '$PSVersionTable…'` → "this command runs pwsh in a plain command; what it reads or is handed as shell text cannot be shown not to run git. Refusing to run it"; `cd <scratch>/A && CLAUDE_CODE_USE_POWERSHELL_TOOL=1 timeout 180 claude -p … --tools PowerShell …` → "this command runs claude with the text PowerShell in a plain command, so what it runs cannot be shown not to be git. Refusing to run it". Not worked around (no script file, no subshell). Consequences: arms A, B, C and the pwsh `-EncodedCommand` spelling measurement were not run; the hook-level PowerShell payload shape (`tool_input` keys, `tool_response` keys) remains UNVERIFIED (REQ-HMP-016 not claimed); the encoded-command detector accepts `-e`, `-ec`, and every prefix of `EncodedCommand` from `-en`, with `-`, `--`, or `/`, on the documented alias set rather than a measured one (over-matching costs one audit line, never a deny).
- **Recipe deviation noted for the lead.** The dispatched recipe carried `--safe-mode`; `claude --help` states it starts "with all customizations (CLAUDE.md, skills, installed plugins, hooks, …) disabled", which would make every arm vacuous for a hook measurement. The prepared arms follow plan M5 and omit it. This did not come into play because both commands were refused.
- Package-level coverage of `internal/hook` and `internal/cli` was not measured this run (the full `internal/hook` suite alone takes ~7 minutes).
- The full `internal/hook/...` and `internal/template/...` runs and `make build` were taken on the working tree before the commits; HEAD differs from that tree only by the `gofmt` whitespace fix in `hmp_source_guard_test.go` and the `spec.md` status line, and the TestHMP set plus lint were re-run on HEAD.

### Residual risk

- The POSIX quote collapse still misreads PowerShell backslash paths and backtick escapes; the deny/ask lists can under-match such text (never newly over-deny). Recorded in spec §A.3 site 1.
- Indirection outside the measured set is not detected: a quoted call target (`& 'git' switch x`), aliases such as `saps`/`start` for Start-Process, and a pwsh path given in quotes. These fall through as allow with no audit line.
- PostToolUse sites 5–6 remain unreachable from the Claude template (matcher omits both shell tools, D3); their conversion is parity-only.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: 2006d3cd3
run_status: complete-with-live-inconclusive
ac_pass_count: 12
ac_fail_count: 0
ac_inconclusive: [AC-HMP-012]
preserve_list_post_run_count: "agent_model_matcher_test.go untouched; PostToolUse matcher and permissions untouched"
l44_pre_commit_fetch: "not run — lanes do not push; local develop absorbed at 553e224f3"
l44_post_push_fetch: "n/a — no push (lead batch-pushes develop)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: "make build exit 0"
  windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
total_run_phase_files: 16  # code/config/test files; plus spec.md (status) and progress.md
m1_to_mN_commit_strategy: "3 feature commits (M1, M2-M3, M4) after one develop-absorption merge; unpushed"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
