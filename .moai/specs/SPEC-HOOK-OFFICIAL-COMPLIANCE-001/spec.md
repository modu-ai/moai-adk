---
id: SPEC-HOOK-OFFICIAL-COMPLIANCE-001
title: "Template Hooks Official-Doc Compliance (32 gaps, 8 Recs)"
version: "0.2.0"
status: in-progress
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: P1
phase: "v14.4.0"
module: "internal/template/templates/.claude/hooks/moai, internal/template/templates/.claude/settings.json.tmpl, .claude/rules/moai/development, .claude/rules/moai/core"
lifecycle: spec-anchored
tier: L
tags: "hook, official-docs, compliance, json-contract, async, timeout, doctrine, template-first"
depends_on: []
related_specs: [SPEC-HOOK-SESSIONSTART-PROBE-001, SPEC-HOOK-FACTFORCE-ADVISORY-001, SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001, SPEC-V3R6-HOOK-ASYNC-EXPAND-001]
---

# SPEC-HOOK-OFFICIAL-COMPLIANCE-001 — Template Hooks Official-Doc Compliance

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-10 | 0.1.0 | Initial plan-phase draft. Tier L. Converts the 2026-07-10 hooks improvement plan report (`.moai/reports/hooks-improvement-plan-20260710.html` — 32 gaps across 8 groups, Top 8 prioritized recommendations) into an actionable GEARS specification. Primary source = the HTML report (SSOT for the audit). Authoritative baseline = `code.claude.com/docs/en/hooks`. Dedup against 6 overlapping SPECs completed (see plan.md §D): SESSIONSTART-PROBE-001 (probe done, residual WorktreeCreate/PermissionRequest fail-open), FACTFORCE-ADVISORY-001 (exit-0 done, residual MOAI_HOOK_STDERR_LOG + state-file escaping), DIVECC-HOOK-FAILURE-MODE-AUDIT-001 (doctrine file exists, 8 stale points unfixed), HOOK-CONTRACT-FIX-001 (WorktreeCreate plain-text, NOT these gates — no overlap), ASYNC-EXPAND-001 (FileChanged/ConfigChange/TaskCreated/Notification async done, NOT UserPromptSubmit/Stop/SubagentStop — no overlap), HOOK-DEADCODE-001 (Go dead code, NOT shell defects — no overlap). 21 REQs across 8 milestones M1-M8. M1 = HIGH severity (JSON contract + PreToolUse exit-code). | manager-spec |
| 2026-07-10 | 0.1.0 | Plan-auditor PASS-WITH-DEBT (0.86) revision. 3 fixes applied: D1 windowed-grep false-pass hazard (acceptance.md AC-HOC-001 anchored to reject-path printf window lines 40-50); D2 REQ-HOC-010 OOS Go-handler-split alternative reworded to non-normative "SHALL DOCUMENT (not implement)" framing (spec.md); D4 GAP-STT-05 added to plan.md §D dedup summary table as DONE (resolved by SESSIONSTART-PROBE-001 probe). 4 documented as Known Audit Debt (acceptance.md §E): D3 dead python3 block in AC-HOC-017, D5 REQ "Event-detected" vs "Event-driven" heading labels, D6 loose AC-HOC-019/020 timeout boundary, D7 ACs hybrid Given/When/Then vs canonical GEARS-AC form. No REQ/AC renumbering; status stays `draft`. | manager-spec |

---

## §A. Problem Statement

### §A.1 What this SPEC is

This SPEC converts the audit findings in `.moai/reports/hooks-improvement-plan-20260710.html` (the "Report") into an implementable specification. The Report audited 39 files (35 template hook files under `internal/template/templates/.claude/hooks/moai/` + 3 rule docs + settings.json.tmpl wiring) against the official Claude Code hooks documentation at `code.claude.com/docs/en/hooks` and identified 32 gaps (2 HIGH, 8 MEDIUM, 15 LOW, 7 INFO) plus 8 prioritized recommendations.

The audit's overall compliance score is **0.80** (8-group weighted average). The Report's §7 methodology limitations section is binding on this SPEC: (a) Go code was NOT audited — several gaps are hypotheses pending `internal/hook/*.go` inspection; (b) 2 of 8 audit groups (tool + lifecycle-core) were analyzed by the orchestrator directly due to a dynamic-workflow 429 rate-limit — their gaps (GAP-TOOL-*, GAP-LC2-*) carry the same authority but note the methodological difference.

### §A.2 The two HIGH-severity defects (M1)

**GAP-GD-01 / GAP-STT-02** — `team-ac-verify.sh` (TaskCompleted event) emits `{"decision":"block","reason":...,"ledger_note":...}` on its reject path. The official TaskCompleted JSON contract is `{"continue":false,"stopReason":...}` (or exit 2 + stderr). The `decision` field is documented only for PostToolUse/Stop/SubagentStop/UserPromptSubmit/ConfigChange/PreCompact/PostToolBatch — NOT TaskCompleted. Verified at `internal/template/templates/.claude/hooks/moai/team-ac-verify.sh:45`.

**GAP-TOOL-01** — `handle-pre-tool.sh.tmpl` uses `printf '%s' "$payload" | moai hook pre-tool 2>>"$MOAI_HOOK_STDERR_LOG"; exit 0` (lines 55-56, 61-62, 67-68), hardcoding `exit 0` on all 3 resolution branches. This overrides the moai binary's exit 2 (PreToolUse reject), and the `2>>` redirect swallows stderr — the channel Claude Code uses to feed the reject reason back to the model. The JSON `permissionDecision:deny` path (b) is preserved via stdout passthrough; the exit-2 path (a) is broken. Verified at `internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl:55-72`.

Both HIGH defects are latent silent-disable failures: the current lenient runtime tolerates them, but a runtime tightening JSON-schema validation would silently disable the entire quality-enforcement surface with no signal.

### §A.3 Dedup against existing SPECs (summary)

See plan.md §D for the full per-Rec dedup verdict. Summary:

| Rec | Overlap SPEC | Verdict |
|-----|-------------|---------|
| #1 (JSON contract + PreToolUse) | HOOK-CONTRACT-FIX-001 | **NEW** — that SPEC covers WorktreeCreate plain-text, NOT these gates |
| #2 (doctrine refresh) | DIVECC-HOOK-FAILURE-MODE-AUDIT-001 | **NEW** — that SPEC created hook-independence.md; this UPDATES it (8 stale points) |
| #3 (async 3 taps) | HOOK-ASYNC-EXPAND-001 | **NEW** — that SPEC covers FileChanged/ConfigChange/TaskCreated/Notification; explicitly excludes UserPromptSubmit/Stop/SubagentStop (§6.3) |
| #4 (timeout headroom) | — | **NEW** |
| #5 (FileChanged matcher + ConfigChange) | — | **NEW** (requires Go verification) |
| #6 (fail-open semantics) | SESSIONSTART-PROBE-001 | **PARTIAL** — probe DONE (completed); WorktreeCreate fail-closed + PermissionRequest warning = residual NEW |
| #7 (input hardening) | FACTFORCE-ADVISORY-001 | **PARTIAL** — exit-0 rewrite DONE (completed); MOAI_HOOK_STDERR_LOG allowlist + gateguard state-file escaping = residual NEW |
| #8 (coverage + defects) | HOOK-DEADCODE-001 | **NEW** — that SPEC is Go dead code (agents/+lifecycle/, dual_parse.go); Rec #8 is shell/settings defects (no collision) |

---

## §B. Goal

Bring the template hooks layer into compliance with `code.claude.com/docs/en/hooks` across 8 recommendation areas, prioritized by severity: (1) fix the two HIGH-severity silent-disable defects on the blocking quality gates and PreToolUse reject path; (2) refresh stale doctrine in a single pass; (3) add async to the 3 remaining synchronous observation taps; (4) give real-work blocking gates timeout headroom; (5) resolve the FileChanged matcher ambiguity and confirm the ConfigChange block channel via Go inspection; (6) correct per-event fail-open semantics for WorktreeCreate (fail-closed) and PermissionRequest (security-negative priority); (7) harden unvalidated inputs; (8) close coverage holes and shell/code defects.

---

## §C. Functional Requirements (GEARS)

> GEARS notation per `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format. Domain token: `HOC` (Hook Official Compliance). Subjects are generalized (the wrapper, the gate, the doctrine, the matcher).

### §C.1 M1 — Blocking gate JSON contract + PreToolUse exit-code (HIGH)

#### REQ-HOC-001 (Event-detected, TaskCompleted JSON contract)

**When** the `team-ac-verify.sh` gate rejects a task completion, the gate **shall** emit exit 0 with stdout JSON `{"continue":false,"stopReason":"AC verification failed: ...","ledger_note":"..."}` — NOT `{"decision":"block",...}`. The `decision` field **shall not** be used for the TaskCompleted event (it is documented only for PostToolUse/Stop/SubagentStop/UserPromptSubmit/ConfigChange/PreCompact/PostToolBatch). The `ledger_note` **shall** be carried as a sidecar field alongside the official `continue`/`stopReason` pair. (GAP-GD-01, GAP-STT-02)

#### REQ-HOC-002 (Event-driven, Stop JSON nesting)

**When** the `sync-phase-quality-gate.sh` gate blocks a Stop event, the gate **shall** emit the decision/reason inside a `hookSpecificOutput` object containing a `hookEventName:"Stop"` field, as `{"hookSpecificOutput":{"hookEventName":"Stop","decision":"block","reason":"..."},"systemMessage":"..."}`. The gate **shall not** emit `decision` as a bare top-level field. (GAP-GD-02)

#### REQ-HOC-003 (Ubiquitous, PreToolUse exit-code passthrough)

The `handle-pre-tool.sh.tmpl` wrapper **shall** propagate the moai binary's exit code to Claude Code on every resolution branch, replacing the hardcoded `exit 0` after the `printf | moai hook pre-tool` invocation. The wrapper **shall** preserve stderr as the Claude-feedback channel for the exit-2 reject path — the `2>>"$MOAI_HOOK_STDERR_LOG"` redirect **shall** be removed or made event-conditional for PreToolUse so that the reject reason reaches the model. The JSON `permissionDecision:deny` stdout path **shall** remain preserved via clean stdout passthrough. (GAP-TOOL-01)

#### REQ-HOC-004 (Ubiquitous, SessionStart probe hookEventName)

The SessionStart moai-resolvability probe in `handle-session-start.sh.tmpl` (line 44) **shall** emit `{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"..."}}` — the `hookEventName:"SessionStart"` field is mandatory and **shall not** be omitted. The current `{"hookSpecificOutput":{"additionalContext":"..."}}` form (missing `hookEventName`) is non-compliant. (GAP-LC2-01)

#### REQ-HOC-005 (Ubiquitous, Ledger Closure doctrine mirror)

The `agent-common-protocol.md` § Ledger Closure clause (b) **shall** encode the corrected TaskCompleted reject form (`continue:false` + `stopReason`), NOT the current incorrect `decision:block` form. The doctrine **shall** stay byte-consistent with the `team-ac-verify.sh` implementation after REQ-HOC-001 lands. (GAP-STT-02 doctrine mirror)

### §C.2 M2 — Doctrine refresh (MEDIUM, single pass)

#### REQ-HOC-006 (Ubiquitous, hooks-system.md 8-point refresh)

The `hooks-system.md` doctrine file **shall** be refreshed in a single editing pass addressing 8 stale points: (1) PreCompact row "Can Block" → `Yes` (exit 2 / `decision:block+reason`); (2) SubagentStop stdout column → add `decision:block+reason | additionalContext`; (3) Notification matcher list → official 8 values (+ SessionEnd 6 / StopFailure 10 rows); (4) FileChanged row → note "literal filenames, NOT regex/glob"; (5) UserPromptSubmit stdout → add `decision:block+reason`, `sessionTitle`, `suppressOriginalPrompt`; PermissionRequest → add `decision.behavior`, `updatedInput`, `updatedPermissions`; (6) event-specific handler-type column (Elicitation/ElicitationResult = command+http+mcp_tool only); (7) "exit 2 is universal" → per-event enum; (8) PostToolUse async constraint (only `additionalContext` deliverable). (GAP-LC-01, GAP-STT-03, GAP-OBS-04, GAP-CEW-03, GAP-PE-02, GAP-PE-03, GAP-GD-06, GAP-TOOL-04)

#### REQ-HOC-007 (Ubiquitous, hook-independence.md timeout correction)

The `hook-independence.md` cross-tab row (g) **shall** show the sync-phase-quality-gate.sh timeout as `60s (Stop)`, correcting the stale `10s` value. The doctrine **shall** agree with both the settings.json.tmpl wiring (line 139: 60) and hooks-system.md. (GAP-GD-03)

### §C.3 M3 — Async observation taps (MEDIUM)

#### REQ-HOC-008 (Capability gate, async for 3 observation taps)

**Where** the `harness-observe` hook is registered for UserPromptSubmit, Stop, or SubagentStop in `settings.json.tmpl`, the registration **shall** carry `"async": true`, aligning with the existing PostToolUse observe tap. The observation tap is non-deterministic, so async loses nothing and removes I/O latency from the user-prompt path and the turn-end (death-spiral-risk) path. (GAP-OBS-01, GAP-OBS-02)

#### REQ-HOC-009 (Capability gate, InstructionsLoaded async consideration)

**Where** the InstructionsLoaded event is registered, the registration **shall** carry `"async": true` to align with the official "runs asynchronously for observability purposes" guidance. This is lower priority than REQ-HOC-008 (InstructionsLoaded fires on lazy loads, not every user prompt). (GAP-LC2-02)

### §C.4 M4 — Timeout headroom (MEDIUM)

#### REQ-HOC-010 (Ubiquitous, TeammateIdle/TaskCompleted timeout)

The TeammateIdle and TaskCompleted blocking quality gates in `settings.json.tmpl` **shall** carry a timeout sufficient to run the LSP-error-threshold gate under a non-anomalous codebase — raised from the current 5s toward the sync-phase-quality-gate 60s precedent. The raise **shall** be documented. If the full raise is deemed excessive, the run-phase **SHALL DOCUMENT** (not implement — Go handler logic changes are Out of Scope per §G) the fast-pre-check + async-deferred full-LSP-verification split as a deferred Go-side follow-up. (GAP-STT-01)

#### REQ-HOC-011 (Ubiquitous, PreCompact timeout)

The PreCompact hook in `settings.json.tmpl` (line 30) **shall** carry a timeout sufficient for the session-memo filesystem side-effect that PostCompact depends on for context recovery — raised from 5s toward the SessionStart 30s class. (GAP-LC-03)

### §C.5 M5 — Matcher resolution + Go verification (MEDIUM)

#### REQ-HOC-012 (Event-detected, FileChanged matcher ambiguity resolution)

**When** the FileChanged matcher `.env|.envrc|.gitignore` (pipe-delimited) is encountered, the SPEC **shall** resolve whether the runtime treats FileChanged matchers as literal filenames (official contract) or as regex. **Where** the runtime honors literal-only, the pipe is NOT alternation and the matcher is malformed — the settings.json.tmpl **shall** register individual FileChanged entries per filename instead. **Where** the runtime treats it as regex, the doctrine **shall** document the shipping runtime's interpretation. The resolution **shall** be grounded in Go inspection of the FileChanged handler / matcher engine, NOT assumption. (GAP-GD-04)

#### REQ-HOC-013 (Event-detected, ConfigChange block channel Go verification)

**When** the Go ConfigChange handler blocks a config change, the handler **shall** emit the block via stdout JSON (`decision:block+reason`, `hookEventName:ConfigChange`), NOT via exit 2 + stderr. The blanket `2>>"$MOAI_HOOK_STDERR_LOG"` redirect in `handle-config-change.sh.tmpl` would otherwise swallow a user-visible block reason. The SPEC **shall** verify which channel the Go handler uses via `internal/hook/*.go` inspection before adjusting the wrapper. (GAP-CEW-02)

### §C.6 M6 — Fail-open semantics correction (MEDIUM)

#### REQ-HOC-014 (Event-detected, WorktreeCreate fail-closed)

**When** the `handle-worktree-create.sh.tmpl` wrapper's missing-binary fallback fires (all 3 resolution tiers fail), the wrapper **shall not** silently `exit 0` with empty stdout — because under the active-creator contract, empty-stdout + exit 0 is creation-ABORT, not graceful skip. The wrapper **shall** either (a) carry a header comment stating it MUST NOT be registered in settings.json unless the moai binary is guaranteed resolvable, and correct the misleading "Claude Code handles missing hooks gracefully" comment; OR (b) emit a non-zero exit + clear stderr diagnostic on the missing-binary branch to make the abort diagnosable. (GAP-CEW-01)

#### REQ-HOC-015 (Event-detected, PermissionRequest fail-open warning)

**When** the `handle-permission-request.sh.tmpl` wrapper's missing-binary fallback fires, the wrapper's exit 0 (= ALLOW per official semantics) **shall** be flagged as a security-negative silent-allow. Because the per-wrapper edit is NOT the right surface (31 wrappers share the chain), the SPEC **shall** implement the standing SessionStart moai-resolvability probe's priority dimension: the probe (already shipped per SPEC-HOOK-SESSIONSTART-PROBE-001) **shall** warn explicitly when the PermissionRequest wrapper would fail-open, so the security-negative case gets priority over the observer-event silent-degradation. (GAP-PE-01)

### §C.7 M7 — Input hardening (LOW)

#### REQ-HOC-016 (Ubiquitous, MOAI_HOOK_STDERR_LOG allowlist)

The shared wrapper pattern that reads `MOAI_HOOK_STDERR_LOG` **shall** validate the value against an allowlist prefix (`$HOME/.moai/logs` or `$CLAUDE_PROJECT_DIR/.moai/logs` subtree) before using it in `mkdir -p`, `mv -f`, or `2>>` redirection. **Where** the value does not match the allowlist, the wrappers **shall** fall back to a default log path. This is applied once in the shared wrapper pattern, not per-wrapper. (GAP-STT-04, GAP-TOOL-02)

#### REQ-HOC-017 (Ubiquitous, gateguard state-file escaping)

The `gateguard-fact-force.sh` state-file write (line 107) **shall** escape `$session_id` / `$file_path` / `$tool_name` interpolation using the same awk-based escaping already used for its `systemMessage` path, OR **shall** record the state as plain `key=value` lines (the state file is checked for existence only, not parsed as JSON — JSON is unnecessary). (GAP-GD-08)

### §C.8 M8 — Coverage holes + shell/code defects (LOW)

#### REQ-HOC-018 (Ubiquitous, MultiEdit matcher extension)

The PostToolUse status-transition matcher in `settings.json.tmpl` (line 104) **shall** carry `Write|Edit|MultiEdit` (adding MultiEdit), so that MultiEdit on a SPEC artifact does not bypass the status-transition-ownership audit log. The matcher **shall** agree with the `status-transition-ownership.sh` case statement (line 54), which already handles Write|Edit|MultiEdit. (GAP-GD-05)

#### REQ-HOC-019 (Ubiquitous, csharp glob fix)

The `sync-phase-quality-gate.sh` `detect_language` csharp branch **shall** replace the dead `[ -f "$root/*.csproj" ]` glob (which never expands inside `[ -f ]`) with `find "$root" -maxdepth 1 -name '*.csproj' -print -quit` (the already-working fallback form). (GAP-GD-07)

#### REQ-HOC-020 (Ubiquitous, agent-hook exec form)

The agent-frontmatter `handle-agent-hook.sh` invocation **shall** use exec form (`args=["{action}"]`) or quote `"{action}"` at the invocation layer, eliminating the shell word-splitting risk for action tokens containing whitespace. (GAP-OBS-03)

#### REQ-HOC-021 (Ubiquitous, compact naming + comment fix)

The `handle-compact.sh.tmpl` line-7 stale self-referencing comment (naming `compact.sh` instead of the deployed `handle-compact.sh`) **shall** be corrected. A rename to `handle-pre-compact.sh.tmpl` (aligning with the `handle-<event-name>` convention) **shall** be evaluated; if the rename is taken, `settings.json.tmpl` + the `moai hook` subcommand mapping **shall** be updated atomically. (GAP-LC-02)

---

## §D. Acceptance Criteria (summary)

See [acceptance.md](./acceptance.md) for the full binary AC matrix (AC-HOC-001 through AC-HOC-036), each with verification command, expected output, and traceability mapping.

---

## §E. Constraints

### §E.1 Template-First Rule (CLAUDE.local.md §2)

Every change to a template hook file MUST be made in `internal/template/templates/` first, then mirrored to the live `.claude/hooks/moai/` counterpart. The doctrine files (`hooks-system.md`, `hook-independence.md`, `agent-common-protocol.md`) have a template mirror that MUST stay byte-identical — enforced by `internal/template/rule_template_mirror_test.go`. A deviation between template and live is a CI failure.

### §E.2 Go code inspection is in-scope as a verification requirement (NOT a wrapper-only fix)

The Report's §7 methodology limitations section is binding: Go code was NOT audited. Several gaps (GAP-CEW-02 ConfigChange block channel, GAP-TOOL-01 PreToolUse exit-code path, GAP-GD-04 FileChanged matcher resolution, GAP-LC2-01 SessionStart probe) are hypotheses pending `internal/hook/*.go` inspection. The run-phase MUST inspect the relevant Go handlers and confirm which JSON path / exit-code path the handler uses before adjusting the wrapper. The run-phase MUST NOT adjust a wrapper on the assumption that the Go handler uses a particular path — it MUST observe the Go source first.

### §E.3 429-rate-limit methodological caveat (Report §7)

Two of 8 audit groups (tool + lifecycle-core) were analyzed by the orchestrator directly in the main context due to a dynamic-workflow 429 rate-limit, not via the structured schema-based automated audit. Their gaps (GAP-TOOL-01..04, GAP-LC2-01..02) carry the same authority but were produced by a different methodology. The plan.md notes this methodological difference; the run-phase treats these gaps identically to the schema-audited ones.

### §E.4 PRESERVE — opt-in infrastructure untouched

The following opt-in infrastructure is PRESERVE (byte-identical, per SPEC-V3R6-HOOK-CONTRACT-FIX-001 REQ-HCF-007): `handle-worktree-create.sh` / `handle-worktree-remove.sh` wrappers + their CLI subcommands + the `worktreeCreateHandler` / `worktreeRemoveHandler` registration. REQ-HOC-014 adds a header comment / guard to the WorktreeCreate wrapper but does NOT remove the wrapper or its registration.

### §E.5 Forbidden operations

- `git reset --hard` (use `--keep` per CLAUDE.local.md §23.5)
- `--no-verify` on commit
- `--amend` (use new commits)
- force-push to main
- Removing any opt-in wrapper or CLI subcommand (PRESERVE per §E.4)
- AskUserQuestion call from any `internal/hook/` or wrapper code (subagent boundary)

---

## §F. Risks

| # | Risk | Likelihood / Impact | Mitigation |
|---|------|---------------------|------------|
| R1 | Runtime tightens JSON-schema validation, silently disabling both blocking quality gates before M1 lands | L / H | M1 is the first milestone (HIGH priority); the existing lenient runtime tolerates the current non-compliant output |
| R2 | Go handler uses a different exit-code/JSON path than the wrapper adjustment assumes | M / H | §E.2 Go inspection is a hard pre-condition; run-phase MUST observe Go source before wrapper edit |
| R3 | Template-mirror drift after doctrine refresh | M / M | §E.1 rule_template_mirror_test.go CI guard; run-phase verifies parity |
| R4 | FileChanged matcher resolution requires runtime test the run-phase cannot perform | M / M | REQ-HOC-012 falls back to "document the shipping runtime's interpretation" if runtime test is infeasible |
| R5 | Raising TeammateIdle/TaskCompleted timeout to 60s slows the no-op fast path | L / L | REQ-HOC-010 allows the fast-pre-check + async-deferred split as an alternative |
| R6 | handle-pre-tool.sh.tmpl exit-code passthrough interacts with the Bash Risk-Amplifier WARN-only design (which is intentionally FAIL-OPEN) | M / M | REQ-HOC-003 preserves the JSON `permissionDecision` stdout path; the exit-2 path is additive (restored, not newly-blocking for the WARN path) |

---

## §G. Out of Scope

### Out of Scope — Single mega-dispatcher consolidation

- Unifying 31 shell wrappers under one `moai hook dispatch <event>` entry-point (the long-term architectural change deferred by SPEC-V3R6-HOOK-ASYNC-EXPAND-001 §6.1) — out of scope; this SPEC keeps the per-event wrapper topology.

### Out of Scope — Go handler logic changes

- Modifying the decision logic inside `internal/hook/*.go` handlers — out of scope. Go code is INSPECTED (verification requirement §E.2) but NOT modified. If Go inspection reveals a handler-side defect, it is recorded as a finding and deferred to a separate SPEC.

### Out of Scope — Re-registration of dormant gates

- Wiring `team-ac-verify.sh` into settings.json.tmpl as a live TaskCompleted hook (GAP-GD-09 INFO) — out of scope. The gate's dormancy is documented as intentional (team-mode activation is a separate concern owned by the team-mode SPEC lineage).

### Out of Scope — moai-binary resolution chain redesign

- Editing the 30 non-WorktreeCreate/non-PermissionRequest wrappers' shared 3-tier resolution chain — out of scope. The standing SessionStart probe (SPEC-HOOK-SESSIONSTART-PROBE-001, completed) surfaces the silent simultaneous degradation; this SPEC does not edit the shared chain.

### Out of Scope — prompt/agent/HTTP/mcp_tool handler types

- Adopting the non-command handler types (prompt, agent, HTTP, mcp_tool) flagged as "not adopted" in the Report §5 — out of scope. The command thin-wrapper architecture delegates decisions to the Go binary; adopting other handler types is a separate architectural initiative.

### Out of Scope — pre-tool 1MB payload cap raise

- GAP-TOOL-03 (the 1MB `head -c 1048576` cap in handle-pre-tool.sh.tmpl) is flagged as INFO-adjacent; the cap is intentionally protective. Raising it or adding a truncation flag is out of scope (low priority, defer to a future SPEC).

---

## §H. Cross-References

- **Primary source (SSOT for the audit)**: `.moai/reports/hooks-improvement-plan-20260710.html`
- **Authoritative baseline**: `code.claude.com/docs/en/hooks` (2026-07-10 fetch)
- **Frontmatter SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- **Template-First Rule**: CLAUDE.local.md §2
- **Template-neutrality CI guard**: `.github/workflows/template-neutrality-check.yaml`
- **Dedup partners**: SPEC-HOOK-SESSIONSTART-PROBE-001 (probe done), SPEC-HOOK-FACTFORCE-ADVISORY-001 (exit-0 done), SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001 (doctrine file exists), SPEC-V3R6-HOOK-CONTRACT-FIX-001 (WorktreeCreate plain-text — no overlap), SPEC-V3R6-HOOK-ASYNC-EXPAND-001 (4-event async — no overlap), SPEC-HOOK-DEADCODE-001 (Go dead code — no overlap)
