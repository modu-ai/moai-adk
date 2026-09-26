---
id: SPEC-HOOK-MATCHER-POWERSHELL-001
title: "Run the shell-command guard hooks for the PowerShell tool, not only for Bash"
version: "0.2.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template/templates/.claude/settings.json.tmpl, internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl, .claude/settings.json, .claude/hooks/moai/handle-pre-tool.sh, internal/hook (pre_tool.go, post_tool.go, evidence_writer.go, integration_lock_guard.go), internal/cli/hook.go"
lifecycle: spec-anchored
tags: "hooks, powershell, windows, matcher, pre-tool, branch-guard, evidence, template, t1224"
era: V3R6
tier: M
related_specs: [SPEC-POWERSHELL-DENY-PARITY-001, SPEC-DUAL-HARNESS-HOOK-PARITY-001, SPEC-HOOK-STDIN-FAILCLOSED-001]
---

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-26 | manager-spec | Initial draft for card t1224 (origin: t1211 / SPEC-POWERSHELL-DENY-PARITY-001 open decision D3, "hook matcher coverage of the PowerShell tool"). Per-site classification of every Bash tool-name branch; unclassifiable-command policy, collision points with t1099 / t1152, and isolated hook-test environment added per lead instruction. |
| 0.2.0 | 2026-09-26 | manager-spec | Plan-audit iter-1 FAIL 0.77 (`.moai/reports/plan-audit/SPEC-HOOK-MATCHER-POWERSHELL-001-review-1.md`) resolved. D1: shell wrapper added as site 14 (class b), REQ-HMP-002 redefined to the wrapper (its former local-copy/rebuild content folded into REQ-HMP-001), REQ-HMP-013 extended to `.sh`/`.sh.tmpl`, new operator decision D6. D2: REQ-HMP-009 rewritten against the landed fail-closed dispatcher (t1152 completed on develop), §C collision text refreshed. D4: hook identity defined as the script path in REQ-HMP-012. D6: `-EncodedCommand` is unclassifiable regardless of payload (REQ-HMP-010). D7: t1211 transcript evidence cited in §A.1; payload capture moved ahead of M3. D11: source guard widened to any `"bash"` literal. D12: slot-lease indirection behaviour stated. D14: SPEC locations noted. REQ count unchanged (16). |

## §A Background

### A.1 What the vendor documentation says (fetched 2026-09-26)

- `https://code.claude.com/docs/en/tools-reference` § PowerShell tool: "Match `Bash|PowerShell` in hooks that inspect shell commands; the PowerShell hook input section explains why matching `Bash` alone is not enough." The tool is enabled with `CLAUDE_CODE_USE_POWERSHELL_TOOL=1`; its name in permission rules and configuration is `PowerShell`.
- `https://code.claude.com/docs/en/hooks` § Matcher patterns: a matcher containing only letters, digits, `_`, `-`, spaces, `,` and `|` is evaluated as an exact string or a `|`/`,`-separated list of exact strings. `Bash|PowerShell` is therefore an exact two-name list; the same page shows `"matcher": "Bash|PowerShell"` in a PreToolUse example.
- Hook input shape for the PowerShell tool: the lane reported (2026-09-26) that the hooks page gives `tool_name: "PowerShell"` with `tool_input.command` shaped like Bash. The manager-spec fetch of the same page on 2026-09-26 did NOT surface a literal PowerShell input example (only the Bash one).
- **Supporting local evidence (transcript-level, not hook-payload-level).** The t1211 M1 LIVE runs (`.moai/reports/t1211/m1/A.jsonl`, `B.jsonl`, committed in this tree) record PowerShell `tool_use.input` = `{"command": "git clean -fdx", "description": "…"}` — the same keys as Bash — and `tool_use_result` keys `["interrupted", "isImage", "stderr", "stdout"]`. This is the model's tool call as seen in the stream, not the JSON a hook receives on stdin; it narrows the risk but does not prove the hook payload shape. The hook payload's `tool_input` keys and `tool_response` keys therefore remain **UNVERIFIED** until the LIVE capture of REQ-HMP-016, which runs in plan M0 (before any per-site conversion relies on the shape beyond fail-safe handling).

### A.2 What the template registers (measured 2026-09-26, base `efc807b81`)

`grep -n '"matcher"' internal/template/templates/.claude/settings.json.tmpl` (local `.claude/settings.json` identical at the same lines):

| Line | Event | Matcher | Hook script (final `args` element) | Relevance |
|---|---|---|---|---|
| 58 | PreToolUse | `Write\|Edit\|Bash` | `handle-pre-tool.sh` | **The gap** — PowerShell calls never reach it |
| 69 / 80 / 91 | PreToolUse | `Agent\|Task`, `SendMessage\|TaskStop`, `AskUserQuestion` | `handle-pre-tool.sh` | Not shell tools |
| 129 | PostToolUse | `Write\|Edit\|MultiEdit\|EnterWorktree\|ExitWorktree` | `handle-post-tool.sh` (+ status-transition hooks) | Omits Bash as well — see site 5/6 below |
| ~135 (opt-in) | PostToolUse | none (match all) | `handle-harness-observe.sh` | Already receives PowerShell |
| others | SessionStart, PreCompact, ConfigChange, StopFailure, FileChanged, … | non-tool matchers | — | Not tool events |

Every template hook entry has `"command": "bash"` and carries the hook script path as the final element of `args` (`${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/<script>`). The `command` field is therefore not an identity; REQ-HMP-012 keys on the script path.

`internal/template/agent_model_matcher_test.go` asserts the exact string `"Write|Edit|Bash"` is present and "must not be renamed or widened" (its purpose: the Agent matcher must be a separate block), and builds a RED fixture on the same literal. Widening the line-58 matcher breaks both; this is decision D1.

### A.3 Why widening the matcher alone does nothing

Every consumer of the hook payload branches on the literal tool name `Bash` — the Go handlers and the shell wrapper that fronts them. A PowerShell payload routed to the same handler falls through every branch and is allowed with no check. Each site was read and classified:

- **(a)** inspects shell-command text whose meaning is the same under PowerShell (git, cloud CLIs, test runners run identically) → should accept PowerShell.
- **(b)** relies on POSIX shell syntax that PowerShell does not share → explicit decision.
- **(c)** not a hook shell-command consumer.

| # | Site (base `efc807b81`) | Function / region | What it inspects | Class | Decision in this SPEC |
|---|---|---|---|---|---|
| 1 | `internal/hook/pre_tool.go:451` | `(*preToolHandler).Handle` — Bash block: `--no-verify` defense + `checkBashCommand` | Substring `git commit` + `--no-verify`; `dangerousRemovalTarget` (POSIX `rm` structural parse); regex deny/ask lists over `substituteQuotedArguments` output | (a) for `--no-verify` and the regex lists (the deny list already carries `Remove-Item`, `Clear-Content`, `rd /s /q` patterns — written for PowerShell/cmd text); **(b)** for the quote-collapse step (POSIX quoting: `\` escapes inside `"…"`, PowerShell uses backtick) and for the `rm` structural parse | Accept PowerShell. Deny/ask stay deny-on-positive-match. Quote-collapse misparse is policy D2 (REQ-HMP-010) and residual risk. `Remove-Item` root removal is already enforced by Claude Code's built-in check (SPEC-POWERSHELL-DENY-PARITY-001 §A.2) |
| 2 | `internal/hook/pre_tool.go:534` | `Handle` → `checkBranchState` (`branch_guard.go:648`, `matchBranchStateCommand`) | `git switch/checkout/branch/reset/stash/rebase/merge` patterns after heredoc/comment/quote substitution | (a) git syntax identical; **(b)** for PowerShell indirection (`& git …`, `Invoke-Expression`/`iex "git …"`, `Start-Process git`, encoded commands) the POSIX-flavoured parser may not see | Accept PowerShell; unclassifiable → D2 |
| 3 | `internal/hook/pre_tool.go:552` | `Handle` → `checkIntegrationLock` (`integration_lock_guard.go:58`) | `git merge` in the release tree | (a); indirection → (b) | Accept PowerShell; unclassifiable → D2 |
| 4 | `internal/hook/pre_tool.go:569` | `Handle` → `checkSlotLease` (`slot_lease_guard.go:55`, also runs `substituteQuotedArguments` at :63) | Configured heavy-command regexes | (a) | Accept PowerShell. Under indirection it keeps its existing fail-open, match-on-scrubbed-text behaviour; no D2 policy applies (its patterns are operator-configured command regexes, not git branch-state patterns) |
| 5 | `internal/hook/post_tool.go:238` | `(*postToolHandler).Handle` → `logEvidence` | Test-command evidence | (a); **unreachable from the Claude template** (PostToolUse matcher omits Bash — pre-existing); reachability via the Codex adapter unmeasured | Convert to the shared predicate; do NOT widen the PostToolUse matcher (D3) |
| 6 | `internal/hook/post_tool.go:248` | `Handle` → `maybeZeroExecutionAdvisory` | Zero-execution test advisory | Same as 5 | Same as 5 |
| 7 | `internal/hook/evidence_writer.go:184` | `maybeZeroExecutionAdvisory` | Same | (a) | Accept PowerShell |
| 8 | `internal/hook/evidence_writer.go:469` | `buildEvidenceRecord` switch → `buildBashRecord` / `classifyTestCommand` | Test-runner prefix + `tool_response` exit code/text | (a) for the command prefix; `tool_response` shape for PowerShell UNVERIFIED at hook level (§A.1) | Accept PowerShell; an unrecognized response shape yields no pass/fail signal (existing fall-through), never a fabricated pass |
| 9 | `internal/hook/evidence_writer.go:650` | `LogBashEvidence` | Same | (a) | Accept PowerShell |
| 10 | `internal/cli/hook.go:788` (develop `35ab8cff3`: `:822`) | `runHarnessObserve` (opt-in, no matcher → already receives PowerShell) | Routes Bash to `LogBashEvidence` | (a) | Accept PowerShell |
| 11 | `internal/permission/stack.go:460` | `IsWriteOperation` (plan-mode write detection in `resolver.go:182`) | Command prefixes | (c) — permission resolver, not a hook consumer | Out of scope; recorded as sibling (D4) |
| 12 | `internal/hook/normalize.go:65,74` | doc comments only | — | (c) | None |
| 13 | `internal/codexadapter/**` (develop) | — | `grep -rn '"Bash"'` → 0 matches on base | (c) on base | None; re-measure after absorbing t1099 (§C) |
| 14 | `internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl:44` (local `.claude/hooks/moai/handle-pre-tool.sh:39`) | Bash Risk-Amplifier warn block, gated by `grep -q '"tool_name"[[:space:]]*:[[:space:]]*"Bash"'`; comment at :43 says "matcher scope is Write\|Edit\|Bash" | Counts `\|\|`, `&&`, `\|`, `;`, backtick, `$(` in the command and warns on stderr above a soft cap of 5 (warn-only, never blocks) | **(b)** — backtick is PowerShell's escape and line-continuation character and `$(` is a PowerShell subexpression, so the same count means something different on PowerShell text | Operator decision D6 (REQ-HMP-002). The comment is updated in the same delivery whatever D6 selects |

## §B Requirements (GEARS)

### B.1 Registration

- **REQ-HMP-001** — The distributed settings template and the local `.claude/settings.json` shall register `handle-pre-tool.sh` for PreToolUse calls of the PowerShell tool in addition to Bash, in the matcher form selected at decision D1, and the embedded template shall be rebuilt so that the binary distributes the change.
- **REQ-HMP-002** — The pre-tool shell wrapper (template `handle-pre-tool.sh.tmpl` and its local copy `handle-pre-tool.sh`) shall apply its Bash Risk-Amplifier subcommand-count warning to PowerShell calls exactly as decision D6 selects, and its matcher-scope comment shall name the matcher delivered by REQ-HMP-001.
- **REQ-HMP-003** — The delivery shall not add Bash or PowerShell to the PostToolUse `handle-post-tool.sh` matcher, and shall not modify the `PowerShell(...)` / `Bash(...)` permission rules delivered by SPEC-POWERSHELL-DENY-PARITY-001.

### B.2 Shared shell-tool predicate

- **REQ-HMP-004** — The hook package shall decide whether a tool call is a shell-command call through one shared predicate that answers true for exactly the tool names `Bash` and `PowerShell`, and every Go site of §A.3 classified (a) shall use it.
- **REQ-HMP-005** — The hook package and the hook CLI shall not carry a string literal equal to `bash` in any letter case outside that shared predicate, except entries on the declared exclusion list of REQ-HMP-013.

### B.3 Guard behavior on PowerShell calls

- **REQ-HMP-006** — When a PowerShell PreToolUse call carries a command string, the `--no-verify` defense and the destructive deny/ask lists shall return the same decision as for a Bash call carrying identical command text.
- **REQ-HMP-007** — Where the branch guard is enabled, when a PowerShell command matches a branch-state pattern in the primary checkout and the agent is not exempt, the pre-tool hook shall deny it with the same `BRANCH_GUARD_VIOLATION:` reason as the Bash path.
- **REQ-HMP-008** — Where the integration lock or the slot lease is enabled, when a PowerShell command matches the guarded pattern while another live session holds the record, the pre-tool hook shall deny it with the same reason as the Bash path.
- **REQ-HMP-009** — When a PreToolUse or PostToolUse payload has a parseable envelope, `tool_name` `PowerShell`, and a `tool_input` that is not a JSON object or has no string `command` field, the pre-tool and post-tool hooks shall not panic and shall return the same output as for the payload that differs only by `tool_name` `Bash`. The delivery shall not change the fail-closed answer that SPEC-HOOK-STDIN-FAILCLOSED-001 (REQ-HSF-001) gives an unparseable decision-event stdin; that answer is tool-independent, because the dispatcher rejects the stdin before any tool name is read.
- **REQ-HMP-010** — When a PowerShell command contains an indirection construct that the existing branch-guard or integration-lock parser does not classify (the set is fixed by the plan M0 measurement over: call operator `& git …`, `Invoke-Expression` / `iex` with an argument naming `git`, `Start-Process` naming `git`, and any `pwsh` / `powershell` invocation carrying `-EncodedCommand` or its short forms `-enc`, `-ec`, `-e` in any letter case), the branch guard and the integration lock shall apply the unclassifiable-command policy selected at decision D2. An encoded-command argument is unclassifiable regardless of its payload and shall not be decoded. The destructive deny list shall keep its deny-on-positive-match behavior regardless of D2, and the slot lease shall keep its existing behaviour (§A.3 site 4).

### B.4 Evidence path

- **REQ-HMP-011** — When a PowerShell test-runner command reaches the evidence path (sites 5–10), the hook shall record the same evidence kind as for the identical Bash command, and when the PowerShell `tool_response` shape is not recognized, the hook shall record no pass/fail signal rather than a pass.

### B.5 Guards against regression

- **REQ-HMP-012** — A template test shall fail when any (event, hook script) pair registered with a matcher that names `Bash` has no registration of the same hook script in the same event whose matcher names `PowerShell`, unless the pair is on a declared exclusion list with a reason; a stale exclusion entry shall also fail the test. The hook script of a registration is the final element of its `args` array with the `${CLAUDE_PROJECT_DIR}/` prefix removed, or the `command` value when `args` is absent.
- **REQ-HMP-013** — A source test shall fail when (i) a non-comment string literal equal to `bash` in any letter case appears in a non-test Go file of `internal/hook` or in `internal/cli/hook.go` outside the shared predicate file, or (ii) a `tool_name` match naming `Bash` in `internal/template/templates/.claude/hooks/moai/*.sh.tmpl` or `.claude/hooks/moai/*.sh` does not also name `PowerShell` — in both cases unless the hit is on a declared exclusion list with a reason; a stale exclusion entry shall also fail the test.
- **REQ-HMP-014** — Every hook test this SPEC adds or modifies shall set `MOAI_HOME` to a per-test temporary directory, shall not call `t.Parallel`, and shall not read or write the real MoAI home, its lease database, or the repository's `.moai/state`.

### B.6 Live evidence

- **REQ-HMP-015** — The run-phase shall attempt one isolated LIVE measurement outside the repository showing whether the hook fires for a PowerShell tool call under the new registration and not under the old one, and whether the branch guard denies a PowerShell branch switch when the hook runs the binary built from this branch. When any validity condition does not hold, the run-phase shall record INCONCLUSIVE and deliver without a LIVE claim. When every validity condition holds and an expected observable is absent or contradicted, the run-phase shall record FAIL naming the arm and the observable and shall return a blocker report instead of delivering.
- **REQ-HMP-016** — When the LIVE measurement captures a PowerShell hook payload, the run-phase shall record its `tool_name`, `tool_input` keys, and `tool_response` keys in the evidence file before any per-site conversion relies on that shape and before claiming REQ-HMP-006 or REQ-HMP-011 hold on the live path.

## §C Constraints and collision points

- **Base and absorption.** Authored on `efc807b81`. Local `develop` was at `35ab8cff3` when 0.2.0 was written. Run-phase M0 absorbs local `develop` and re-measures every line number in §A.3 before editing (site 10 already moved to develop `internal/cli/hook.go:822`).
- **Landed: card t1152 — SPEC-HOOK-STDIN-FAILCLOSED-001 (`status: completed` on develop; absent from this tree until M0 absorption).** develop `internal/cli/hook.go` `runHookEvent` (~L280–293 at audit time) answers an unparseable decision-event stdin with a fail-closed deny via `answerStdinParseFailure`. Behavioural interaction: REQ-HMP-009 defers to REQ-HSF-001 for the envelope-unparseable case and covers only the parseable-envelope case. An unparseable `tool_input` inside a parseable envelope is unreachable (`HookInput.ToolInput` is a `json.RawMessage` in `internal/hook/types.go:218`, which must be valid JSON for the envelope to decode), so only wrongly-typed values (string, array, number, null, object without a string `command`) occur.
- **Concurrent: card t1099 — SPEC-DUAL-HARNESS-HOOK-PARITY-001 (`status: in-progress` on branch `WT-dual-harness-parity-rebuild` at `e722a1493`; absent from this tree and from develop).** `git diff --name-only develop...WT-dual-harness-parity-rebuild` (2026-09-26) lists `internal/cli/hook.go`, `internal/cli/hook_test.go`, other `internal/cli/*codex*` files, `internal/codexadapter/**` (including a new `translate.go`), and `internal/template/obligations*`; it lists no `internal/hook/*`, no `settings.json.tmpl`, and no `.claude/hooks/*`. Its `hook.go` hunks fall in `init`, `runHookEvent`, and `runHarnessObserveStop`.
- **Regions this SPEC edits:** `pre_tool.go` `(*preToolHandler).Handle` — the four tool-name conditions only; `post_tool.go` `(*postToolHandler).Handle` two conditions only; `evidence_writer.go` `maybeZeroExecutionAdvisory`, `buildEvidenceRecord` switch, `LogBashEvidence`; `integration_lock_guard.go` only if D2=(A) adds its audit line; `internal/cli/hook.go` `runHarnessObserve` condition only (adjacent to t1099's `runHarnessObserveStop` hunk); one new predicate file in `internal/hook`; the PreToolUse block of the template and local settings; the Risk-Amplifier block of the template and local pre-tool wrapper. New CLI-level tests go in a new `_test.go` file, not in `internal/cli/hook_test.go` (t1099 edits it).
- **Codex adapter.** If t1099's `translate.go` lands before run and maps a Codex shell tool onto `tool_name`, site 13 is re-classified in M0 and the predicate's closed set is revisited (it stays `Bash`, `PowerShell` unless the operator widens it).
- **Test environment.** Hook verification runs env-scrubbed and scoped: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/hook/... ./internal/template/... ./internal/cli/... -run '<SPEC tests>'`. Isolation of the real MoAI home is the tests' own obligation (REQ-HMP-014) and is not supplied by the command line, so a test that forgets it is caught by the static check of AC-HMP-011 rather than masked by a process-level override. No full-suite local run.

## §D Out of Scope

### Out of Scope — PowerShell grammar

- No PowerShell parser is added. Guards keep their POSIX-flavoured tokenizers; PowerShell-specific indirection is handled only by the D2 policy, and encoded commands are never decoded.

### Out of Scope — PostToolUse matcher widening

- `handle-post-tool.sh` stays registered for `Write|Edit|MultiEdit|EnterWorktree|ExitWorktree`. Its Bash branches (sites 5–6) are converted for parity only; making Bash/PowerShell reach it is a separate decision (D3).

### Out of Scope — Hook runtime on Windows without Git Bash

- Every hook entry runs `bash -c …`. On Windows without Git Bash no hook runs for any tool. That is a separate card candidate (D4), not addressed here.

### Out of Scope — Permission rules and the plan-mode write classifier

- `PowerShell(...)` / `Bash(...)` permission rules (t1211) are untouched.
- `internal/permission/stack.go` `IsWriteOperation` (site 11) is recorded as a sibling gap only (D4).

### Out of Scope — Widening the shell-tool set

- Tool names other than `Bash` and `PowerShell` (for example a future Codex shell tool name) are not added without an operator decision.

### Out of Scope — The stdin parse-failure policy

- The fail-closed answer to an unparseable decision-event stdin belongs to SPEC-HOOK-STDIN-FAILCLOSED-001 and is not changed here.
