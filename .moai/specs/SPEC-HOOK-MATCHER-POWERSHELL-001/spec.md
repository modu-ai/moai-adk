---
id: SPEC-HOOK-MATCHER-POWERSHELL-001
title: "Run the shell-command guard hooks for the PowerShell tool, not only for Bash"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template/templates/.claude/settings.json.tmpl, .claude/settings.json, internal/hook (pre_tool.go, post_tool.go, evidence_writer.go), internal/cli/hook.go"
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

## §A Background

### A.1 What the vendor documentation says (fetched 2026-09-26)

- `https://code.claude.com/docs/en/tools-reference` § PowerShell tool: "Match `Bash|PowerShell` in hooks that inspect shell commands; the PowerShell hook input section explains why matching `Bash` alone is not enough." The tool is enabled with `CLAUDE_CODE_USE_POWERSHELL_TOOL=1`; its name in permission rules and configuration is `PowerShell`.
- `https://code.claude.com/docs/en/hooks` § Matcher patterns: a matcher containing only letters, digits, `_`, `-`, spaces, `,` and `|` is evaluated as an exact string or a `|`/`,`-separated list of exact strings. `Bash|PowerShell` is therefore an exact two-name list; the same page shows `"matcher": "Bash|PowerShell"` in a PreToolUse example.
- Hook input shape for the PowerShell tool: the lane reported (2026-09-26) that the hooks page gives `tool_name: "PowerShell"` with `tool_input.command` shaped like Bash. The manager-spec fetch of the same page on 2026-09-26 did NOT surface a literal PowerShell input example (only the Bash one). The `tool_input.command` field and the `tool_response` shape for PowerShell are therefore **UNVERIFIED** and are captured in the run-phase LIVE measurement (REQ-HMP-015) before any requirement relies on them beyond fail-safe handling.

### A.2 What the template registers (measured 2026-09-26, base `efc807b81`)

`grep -n '"matcher"' internal/template/templates/.claude/settings.json.tmpl` (local `.claude/settings.json` identical at the same lines):

| Line | Event | Matcher | Command | Relevance |
|---|---|---|---|---|
| 58 | PreToolUse | `Write\|Edit\|Bash` | `handle-pre-tool.sh` | **The gap** — PowerShell calls never reach it |
| 69 / 80 / 91 | PreToolUse | `Agent\|Task`, `SendMessage\|TaskStop`, `AskUserQuestion` | `handle-pre-tool.sh` | Not shell tools |
| 129 | PostToolUse | `Write\|Edit\|MultiEdit\|EnterWorktree\|ExitWorktree` | `handle-post-tool.sh` (+ status-transition hooks) | Omits Bash as well — see site 5/6 below |
| ~135 (opt-in) | PostToolUse | none (match all) | `handle-harness-observe.sh` | Already receives PowerShell |
| others | SessionStart, PreCompact, ConfigChange, StopFailure, FileChanged, … | non-tool matchers | — | Not tool events |

`internal/template/agent_model_matcher_test.go` asserts the exact string `"Write|Edit|Bash"` is present and "must not be renamed or widened" (its purpose: the Agent matcher must be a separate block). Widening the line-58 matcher breaks that guard; this is decision D1.

### A.3 Why widening the matcher alone does nothing

Every Go consumer of the hook payload branches on the literal tool name `"Bash"`. A PowerShell payload routed to the same handler falls through every branch and is allowed with no check. Each site was read and classified:

- **(a)** inspects shell-command text whose meaning is the same under PowerShell (git, cloud CLIs, test runners run identically) → should accept PowerShell.
- **(b)** relies on POSIX shell syntax that PowerShell does not share → explicit decision.
- **(c)** not a hook shell-command consumer.

| # | Site (base `efc807b81`) | Function / region | What it inspects | Class | Decision in this SPEC |
|---|---|---|---|---|---|
| 1 | `internal/hook/pre_tool.go:451` | `(*preToolHandler).Handle` — Bash block: `--no-verify` defense + `checkBashCommand` | Substring `git commit` + `--no-verify`; `dangerousRemovalTarget` (POSIX `rm` structural parse); regex deny/ask lists over `substituteQuotedArguments` output | (a) for `--no-verify` and the regex lists (the deny list already carries `Remove-Item`, `Clear-Content`, `rd /s /q` patterns — written for PowerShell/cmd text); **(b)** for the quote-collapse step (POSIX quoting: `\` escapes inside `"…"`, PowerShell uses backtick) and for the `rm` structural parse | Accept PowerShell. Deny/ask stay deny-on-positive-match. Quote-collapse misparse is policy D2 (REQ-HMP-010) and residual risk. `Remove-Item` root removal is already enforced by Claude Code's built-in check (SPEC-POWERSHELL-DENY-PARITY-001 §A.2) |
| 2 | `internal/hook/pre_tool.go:534` | `Handle` → `checkBranchState` (`branch_guard.go:648`, `matchBranchStateCommand`) | `git switch/checkout/branch/reset/stash/rebase/merge` patterns after heredoc/comment/quote substitution | (a) git syntax identical; **(b)** for PowerShell indirection (`& git …`, `Invoke-Expression`/`iex "git …"`, `Start-Process git`) the POSIX-flavoured parser cannot see | Accept PowerShell; unclassifiable → D2 |
| 3 | `internal/hook/pre_tool.go:552` | `Handle` → `checkIntegrationLock` (`integration_lock_guard.go:58`) | `git merge` in the release tree | (a); indirection → (b) | Accept PowerShell; unclassifiable → D2 |
| 4 | `internal/hook/pre_tool.go:569` | `Handle` → `checkSlotLease` (`slot_lease_guard.go:55`) | Configured heavy-command regexes | (a) | Accept PowerShell |
| 5 | `internal/hook/post_tool.go:238` | `(*postToolHandler).Handle` → `logEvidence` | Test-command evidence | (a); **unreachable from the Claude template** (PostToolUse matcher omits Bash — pre-existing); reachability via the Codex adapter unmeasured | Convert to the shared predicate; do NOT widen the PostToolUse matcher (D3) |
| 6 | `internal/hook/post_tool.go:248` | `Handle` → `maybeZeroExecutionAdvisory` | Zero-execution test advisory | Same as 5 | Same as 5 |
| 7 | `internal/hook/evidence_writer.go:184` | `maybeZeroExecutionAdvisory` | Same | (a) | Accept PowerShell |
| 8 | `internal/hook/evidence_writer.go:469` | `buildEvidenceRecord` switch → `buildBashRecord` / `classifyTestCommand` | Test-runner prefix + `tool_response` exit code/text | (a) for the command prefix; `tool_response` shape for PowerShell UNVERIFIED (§A.1) | Accept PowerShell; an unrecognized response shape yields no pass/fail signal (existing fall-through), never a fabricated pass |
| 9 | `internal/hook/evidence_writer.go:650` | `LogBashEvidence` | Same | (a) | Accept PowerShell |
| 10 | `internal/cli/hook.go:788` | `runHarnessObserve` (opt-in, no matcher → already receives PowerShell) | Routes Bash to `LogBashEvidence` | (a) | Accept PowerShell |
| 11 | `internal/permission/stack.go:460` | `IsWriteOperation` (plan-mode write detection in `resolver.go:182`) | Command prefixes | (c) — permission resolver, not a hook consumer | Out of scope; recorded as sibling (D4) |
| 12 | `internal/hook/normalize.go:65,74` | doc comments only | — | (c) | None |
| 13 | `internal/codexadapter/**` (develop) | — | `grep -rn '"Bash"'` → 0 matches on base | (c) on base | None; re-measure after absorbing t1099 (§C) |

## §B Requirements (GEARS)

### B.1 Registration

- **REQ-HMP-001** — The distributed settings template shall register `handle-pre-tool.sh` for PreToolUse calls of the PowerShell tool in addition to Bash, in the matcher form selected at decision D1.
- **REQ-HMP-002** — The local `.claude/settings.json` shall carry the same PowerShell registration as the template, and the embedded template shall be rebuilt so that the binary distributes the change.
- **REQ-HMP-003** — The delivery shall not add Bash or PowerShell to the PostToolUse `handle-post-tool.sh` matcher, and shall not modify the `PowerShell(...)` / `Bash(...)` permission rules delivered by SPEC-POWERSHELL-DENY-PARITY-001.

### B.2 Shared shell-tool predicate

- **REQ-HMP-004** — The hook package shall decide whether a tool call is a shell-command call through one shared predicate that answers true for exactly the tool names `Bash` and `PowerShell`, and every site of §A.3 classified (a) shall use it.
- **REQ-HMP-005** — The hook package and the hook CLI shall not compare a tool name against the literal `"Bash"` outside that shared predicate.

### B.3 Guard behavior on PowerShell calls

- **REQ-HMP-006** — When a PowerShell PreToolUse call carries a command string, the `--no-verify` defense and the destructive deny/ask lists shall return the same decision as for a Bash call carrying identical command text.
- **REQ-HMP-007** — Where the branch guard is enabled, when a PowerShell command matches a branch-state pattern in the primary checkout and the agent is not exempt, the pre-tool hook shall deny it with the same `BRANCH_GUARD_VIOLATION:` reason as the Bash path.
- **REQ-HMP-008** — Where the integration lock or the slot lease is enabled, when a PowerShell command matches the guarded pattern while another live session holds the record, the pre-tool hook shall deny it with the same reason as the Bash path.
- **REQ-HMP-009** — When a PowerShell tool input is not parseable JSON or lacks a string `command` field, the pre-tool and post-tool hooks shall not deny, shall not panic, and shall return the same output as for a Bash input with the same defect.
- **REQ-HMP-010** — When a PowerShell command names `git` through an indirection construct the guard parsers do not classify (call operator `&`, `Invoke-Expression` / `iex`, `Start-Process`, `-EncodedCommand`), the branch guard and the integration lock shall apply the unclassifiable-command policy selected at decision D2, and the destructive deny list shall keep its deny-on-positive-match behavior regardless of D2.

### B.4 Evidence path

- **REQ-HMP-011** — When a PowerShell test-runner command reaches the evidence path (sites 5–10), the hook shall record the same evidence kind as for the identical Bash command, and when the PowerShell `tool_response` shape is not recognized, the hook shall record no pass/fail signal rather than a pass.

### B.5 Guards against regression

- **REQ-HMP-012** — A template test shall fail when any (event, hook command) pair registered with a matcher that names `Bash` has no registration of the same command in the same event whose matcher names `PowerShell`, unless the pair is on a declared exclusion list with a reason; a stale exclusion entry shall also fail the test.
- **REQ-HMP-013** — A source test shall fail when a `"Bash"` tool-name comparison appears in `internal/hook` or the hook CLI outside the shared predicate.
- **REQ-HMP-014** — Every hook test this SPEC adds or modifies shall run with an isolated `MOAI_HOME` under a per-test temporary directory, and shall not read or write the real MoAI home, its lease database, or the repository's `.moai/state`.

### B.6 Live evidence

- **REQ-HMP-015** — The run-phase shall attempt one isolated LIVE measurement outside the repository showing whether the hook fires for a PowerShell tool call under the new registration and not under the old one, and when any validity condition of that measurement does not hold, the run-phase shall record the outcome as INCONCLUSIVE and deliver without a LIVE claim.
- **REQ-HMP-016** — When the LIVE measurement captures a PowerShell hook payload, the run-phase shall record its `tool_name`, `tool_input` keys, and `tool_response` keys in the evidence file before claiming REQ-HMP-006 or REQ-HMP-011 hold on the live path.

## §C Constraints and collision points

- **Base and absorption.** Authored on `efc807b81`. Local `develop` had advanced to `044fb91c7` at authoring time. Run-phase M0 absorbs local `develop` and re-measures every line number in §A.3 before editing.
- **Concurrent lanes touching the same package (measured 2026-09-26 with `git diff --name-only develop...<branch>`):**
  - card t1099 — SPEC-DUAL-HARNESS-HOOK-PARITY-001, branch `WT-dual-harness-parity-rebuild` (`e722a1493`): edits `internal/cli/hook.go` (hunks near `init` ~L93, `runHookEvent` ~L331–376, `runHarnessObserveStop` ~L823–933) and `internal/codexadapter/**` (including a new `translate.go`).
  - card t1152 — SPEC-HOOK-STDIN-FAILCLOSED-001, branch `WT-hook-stdin-failclosed` (`ed1759b65`): edits `internal/cli/hook.go` (`hookCmd`/`init` ~L34–120, `runHookEvent` ~L269–360, `runAgentHook` ~L451–540) and `internal/codexadapter/**`.
  - Neither branch edits `internal/hook/pre_tool.go`, `internal/hook/post_tool.go`, `internal/hook/evidence_writer.go`, or `settings.json.tmpl` (0 files in their diffs).
- **Regions this SPEC edits:** `pre_tool.go` `(*preToolHandler).Handle` — the four `ToolName == "Bash"` conditions at L451/534/552/569 only; `post_tool.go` `(*postToolHandler).Handle` L238/L248 conditions only; `evidence_writer.go` `maybeZeroExecutionAdvisory`, `buildEvidenceRecord` switch, `LogBashEvidence`; `internal/cli/hook.go` `runHarnessObserve` L788 condition only (adjacent to t1099's `runHarnessObserveStop` hunk starting ~L823); one new predicate file in `internal/hook`; the PreToolUse block of the template and local settings.
- **Codex adapter.** If t1099's `translate.go` lands before run and maps a Codex shell tool onto `tool_name`, site 13 is re-classified in M0 and the predicate's closed set is revisited (it stays `Bash`, `PowerShell` unless the operator widens it).
- **Test environment.** Hook verification runs env-scrubbed and scoped: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && MOAI_HOME="<scratch dir>" go test -count=1 ./internal/hook/... ./internal/template/... -run '<SPEC tests>'`, with each test also setting `MOAI_HOME` to its own `t.TempDir()` (REQ-HMP-014). No full-suite local run.

## §D Out of Scope

### Out of Scope — PowerShell grammar

- No PowerShell parser is added. Guards keep their POSIX-flavoured tokenizers; PowerShell-specific indirection is handled only by the D2 policy.

### Out of Scope — PostToolUse matcher widening

- `handle-post-tool.sh` stays registered for `Write|Edit|MultiEdit|EnterWorktree|ExitWorktree`. Its Bash branches (sites 5–6) are converted for parity only; making Bash/PowerShell reach it is a separate decision (D3).

### Out of Scope — Hook runtime on Windows without Git Bash

- Every hook entry runs `bash -c …`. On Windows without Git Bash no hook runs for any tool. That is a separate card candidate (D4), not addressed here.

### Out of Scope — Permission rules and the plan-mode write classifier

- `PowerShell(...)` / `Bash(...)` permission rules (t1211) are untouched.
- `internal/permission/stack.go` `IsWriteOperation` (site 11) is recorded as a sibling gap only (D4).

### Out of Scope — Widening the shell-tool set

- Tool names other than `Bash` and `PowerShell` (for example a future Codex shell tool name) are not added without an operator decision.
