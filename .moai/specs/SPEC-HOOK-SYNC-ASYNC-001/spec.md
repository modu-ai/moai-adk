---
id: SPEC-HOOK-SYNC-ASYNC-001
title: "Hook sync/async cleanup: three async conversions, codex-review-entry gate-ON shell fast path, timeout doctrine correction, and the sync-hook measurement protocol"
version: "1.0.3"
status: completed
created: 2026-10-15
updated: 2026-10-22
author: manager-spec
priority: P1
phase: "v3.2.0"
module: ".claude/hooks/moai, internal/hook, internal/template/templates/.claude"
lifecycle: spec-anchored
tags: "hooks, async, timeout, template-first, performance, measurement"
tier: M
related_specs: [SPEC-GATE-BOTTLENECK-001, SPEC-REVIEW-GATE-REPAIR-001, SPEC-HARNESS-RATCHET-REWIRE-001, SPEC-CODEX-GATE-SCOPE-001, SPEC-CODEX-REVIEW-OWNERSHIP-001, SPEC-CODEX-GATE-SCOPING-001]
---

# SPEC-HOOK-SYNC-ASYNC-001

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-10-15 | 1.0.0 | Initial draft. Authored at plan phase for card t1780 (operator directive 2026-10-15, Class C, Template-First, P1). Carries board decisions d-20261015T030xx (codex-review-entry stays synchronous) and d-20261015T024551Z-78aa (measure every synchronous hook 10x, table under .moai/reports/t1780/). |
| 2026-10-15 | 1.0.1 | Plan-audit r1 repairs (verdict `.moai/reports/t1780/plan-audit.md`, FAIL 0.8125 on D1-D4 blocking): REQ-HSA-006 kick narrowed to cache-miss states only — fresh pass/inconclusive exits 0 WITHOUT kicking, mirroring `codex_review_gate.go:234-244` (D1); tree-scope/primary-scope skip policy added to the REQ-HSA-006 Where-clause (D2); REQ-HSA-003 gains the `hooks-system.md` PostToolUseFailure stdout-row refresh (D7). AC-HSA-017 added per audit D6 (only AC-set addition; every existing REQ/AC id unchanged). |
| 2026-10-15 | 1.0.2 | Plan-audit r2 repairs (delta FAIL, r1 fixes verified resolved): REQ-HSA-006(a) early kick now carries the two kick-region preconditions — binary-age (decidable-behind suppresses, binlag leniency) and nothing-reviewable (change presence + config-only narrowing), shell-carried where judgeable, ALLOW-without-early-kick otherwise (R2-D3); REQ-HSA-007 corrected — an unreadable/unparseable snapshot is receipt-ABSENT (the shared `verify.LoadReceipt` reading) and takes the REQ-HSA-006(a) miss arm, removed from the fall-through enumeration (R2-D2, Go authority read). Id sets unchanged. |
| 2026-10-15 | 1.0.3 | Plan-audit r3 repairs (delta FAIL, r2 fixes verified resolved; ceiling-policy auto-delta): REQ-HSA-006(a) binary-age gains the HEAD-match pre-check FIRST — a binary commit matching HEAD is StatusFresh (`binlag.go:139-141`), the early kick proceeds; only a non-matching ancestor-hit is behind (`merge-base --is-ancestor` succeeds on equality and would mis-class the exactly-current binary) (R3-D1); AC-HSA-015's exactly-one bound scoped to the same kick bucket, with the bucket-roll case recorded as designed behavior (R3-D2); §D.6 TTL-boundary corrected — exactly-24h is FRESH under the strict `>` comparison (R3-O3). Id sets unchanged. |

## 1. Overview

The leader's full survey of `.claude/settings.json` (41 command-hook registrations including the
statusLine offset, ~30 real hook entries) decided a sync/async cleanup of the hook layer. Three
registrations convert to `async: true` — their results only need to reach the next turn as
`additionalContext`, so paying a synchronous slot for them is waste:

1. **ConfigChange** (`handle-config-change.sh`, timeout 5, no handler output).
2. **SubagentStop** (`chain-event.sh`, timeout 5, ledger append only, fail-open).
3. **PostToolUseFailure** (`handle-post-tool-failure.sh`, timeout 5, failure record +
   SystemMessage — the SystemMessage emission lives at `internal/hook/post_tool_failure.go:108-110`).
   This one converts **AND** moves its message to `additionalContext`, because an async hook
   delivers ONLY `additionalContext`; it cannot deliver `systemMessage`.

One registration stays synchronous by decision: **UserPromptSubmit
`handle-codex-review-entry.sh`** (board decision d-20261015T030xx — decision item ④, closed; this
SPEC does not reopen it). It gains a **gate-ON shell fast path**: a pure-shell receipt-freshness
probe that resolves the common ALLOW states without paying the synchronous `moai` binary cold
start, falling through to the Go handler only when a fresh-FAIL receipt needs enforcement.

The timeout doctrine in `hooks-system.md` (live + template mirror) is corrected: per the official
Claude Code hooks documentation, 30 seconds is the **default** for UserPromptSubmit command hooks
with **no documented per-event maximum** — not a cap. The template's UserPromptSubmit timeout
values are aligned to the corrected doctrine.

Finally, per board decision d-20261015T024551Z-78aa, **every synchronous hook registration is
measured 10 times under the current machine load** (no artificial load), with per-run timings
recorded to `.moai/reports/t1780/`, and prescriptions follow a fixed order: (1) shell fast-paths,
(2) timeout raises within per-event runtime caps, (3) async only for hooks the evidence table
passes.

### Code-verified semantics note (recorded so run phase designs toward the live code)

The card directive describes the entry hook's purpose as "blocking the prompt on a fresh FAIL".
The live entry handler carries SPEC-REVIEW-GATE-REPAIR-001 REQ-RG-001's no-block amendment
(verified in `internal/cli/codex_review_delay.go` `HandleCodexReviewEntry`): a fresh FAIL never
blocks a prompt; its finding detail is delivered as `additionalContext` once per review key
(REQ-RG-002, delivery record per key). The DECIDED item ④ — stays synchronous, gains the shell
fast path — holds under either semantics (the enforcement is turn-timing-critical and async
delivery would arrive one turn late, useless for a verdict about the previous turn's tree). This
SPEC records the code-verified semantics so the run phase implements the fall-through toward the
handler that exists, not toward a block the handler deliberately does not perform.

## 2. Requirements (GEARS)

### Async conversions

- **REQ-HSA-001** — The ConfigChange registration of `handle-config-change.sh` shall carry
  `"async": true` in both the live settings and the template settings, so the handler's
  continue-only, empty-output behavior costs no synchronous turn time.

- **REQ-HSA-002** — The SubagentStop registration of `chain-event.sh` shall carry `"async": true`
  in both settings surfaces; the chain-ledger append is fail-open and emits no decision output,
  so next-turn `additionalContext` delivery loses nothing.

- **REQ-HSA-003** — The PostToolUseFailure registration of `handle-post-tool-failure.sh` shall
  carry `"async": true` in both settings surfaces, **and** the Go handler
  (`internal/hook/post_tool_failure.go`) shall emit the classified failure message as
  `additionalContext` (via `hookSpecificOutput.additionalContext`) instead of `systemMessage`,
  because an async hook delivers only `additionalContext`. The REQ-HRR-001
  `tool_failure:<tool>:<category>` usage-log recording and the contract-mode escalation
  observation shall remain unchanged, and the `hooks-system.md` stdin/stdout reference row for
  PostToolUseFailure (live and template mirrors) shall be updated to the new output field in the
  same change set, so the doctrine does not go stale against the converted handler.

### Decision item ④ — codex-review-entry (DECIDED, not reopened)

- **REQ-HSA-004** — The UserPromptSubmit registration of `handle-codex-review-entry.sh` shall
  remain synchronous. Async is forbidden for this hook (board decision d-20261015T030xx; card
  t1780 directive).

- **REQ-HSA-005** — While the review gate is disabled or unreadable, the entry wrapper shall exit
  0 before any binary resolution (the existing gate-OFF shell fast path, unchanged).

- **REQ-HSA-006** — Where the review gate is enabled, the session resolves to tree scope, AND
  neither entry-side skip policy applies — the `tree_scope: skip` config and the
  primary-checkout `primary_scope: skip` default, the two sibling policies the Go entry handler
  applies at `internal/cli/codex_review_delay.go:217` so "the two halves skip together" (a
  `WT-` card session is never skipped — REQ-CRO-004) — the entry wrapper shall probe the
  codex-review receipt store in pure shell — snapshot path
  `.moai/state/verify/snapshots/<head12>-<digest>.json` (head truncated to 12 chars, `:` → `-`),
  key = `sha256(porcelain-v2 status || 0x00 || git diff HEAD || per-untracked-file digests)` per
  `internal/verify/key.go`, entry `check_id == "codex-review"`, freshness TTL 24h
  (`codexReviewCacheTTL`) — and:
  - (a) **When** a cache-MISS state is observed (no receipt, or a stale receipt past the TTL),
    the wrapper shall exit 0 immediately (ALLOW) without a synchronous handler invocation, and
    shall fire the detached background review kick (`moai verify codex-review --project-root
    <dir>`, output appended to `.moai/logs/codex-review-bg.log`) guarded by the same kick marker
    (`.moai/state/verify/codex-review/<head>-<digest>.kick-<bucket>`, bucket = epoch /
    `DefaultCodexReviewGateTimeout` = 900s) the Stop gate uses on the same miss states
    (`codex_review_gate.go:249-270` — only miss states reach the kick), so the kick arm of
    REQ-GBN-002 is preserved without double-kicking against the Stop side. The early kick
    additionally carries the two preconditions that kick region sits behind
    (`codex_review_delay.go:214-225`), each carried in shell where judgeable and taking the
    ALLOW-without-early-kick disposition otherwise (the kick is accelerated, never widened):
    the binary-age policy — only a DECIDABLE behind verdict suppresses the early kick, judged
    in shell via the resolved binary's version provenance, a HEAD-match pre-check FIRST (the
    binary commit matching HEAD is NOT behind — StatusFresh, `binlag.go:139-141` — and the
    early kick proceeds; the ancestry probe alone would mis-class an exactly-current binary,
    since `merge-base --is-ancestor` succeeds on equality), then
    `git merge-base --is-ancestor <bin-commit> HEAD` — a non-matching ancestor-hit is behind —
    mirroring binlag's leniency contract
    (fresh / divergent / ahead / not-applicable all proceed — `codex_review_binary_age.go:33-37`);
    and the nothing-reviewable detector (tree-scope form) — change presence plus the
    config-only narrowing, the latter mirrored from the Go-owned runtime-config prefix list
    under a golden test so the shell probe only narrows, never widens, the detector's answer
    (`codex_review_scope.go:268`);
  - (b) **When** a fresh pass or inconclusive verdict is observed (five bound fields equal +
    within TTL, verdict not fail), the wrapper shall exit 0 immediately WITHOUT kicking —
    mirroring the Stop gate's clear-and-reuse disposition (`codex_review_gate.go:234-244`: a
    fresh passing review is never re-kicked, so an idle passing tree spends no review budget);
  - (c) **When** a fresh-FAIL receipt is observed (five bound fields equal + within TTL + verdict
    fail), the wrapper shall fall through to the synchronous `moai hook codex-review-entry` so
    the Go handler owns the verdict delivery.

- **REQ-HSA-007** — **When** any shell-probe dependency is absent or any probe step fails
  (missing `git` or sha256 tool, card-scope session on a `WT-` branch, an unapplied skip policy
  per the REQ-HSA-006 Where-clause, unresolvable moai binary), the wrapper shall fall through to
  the current behavior — the synchronous Go handler where the binary resolves, else exit 0 —
  never a block and never a synthetic verdict. An unreadable or unparseable snapshot file is NOT
  a failure of this kind: it is receipt-ABSENT by the shared store's own reading
  (`verify.LoadReceipt`, `internal/verify/receipt.go:113-117` — every read failure is absence,
  so a partially written receipt can never be read as valid or failing) and takes the
  REQ-HSA-006(a) miss arm. The fast path may skip the binary only on positive evidence of
  an ALLOW state; uncertainty always falls through.

- **REQ-HSA-008** — The shell fast path shall introduce no new shared failure mode per
  `hook-independence.md` §7 (authoring checklist): new dependencies limited to `git` and one
  sha256 tool (`shasum` or `openssl`), every absent dependency degrading per REQ-HSA-007, and the
  shell-computed receipt key pinned equal to the Go `verify.Key` output by a test.

### Keep-sync invariants

- **REQ-HSA-009** — The registrations in the keep-sync table below shall remain synchronous with
  their current decision semantics; none of them gains async in this SPEC.

| Registration | Event | Why it stays synchronous |
|---|---|---|
| `handle-pre-tool.sh` (5 matcher kinds) | PreToolUse | Pre-execution decision/risk-tier signal; the block window is the point |
| permission handlers | PermissionRequest, PermissionDenied | allow/deny/`{retry:true}` decisions consumed in-turn |
| Stop-family blocking gates (`handle-stop.sh`, `sync-phase-quality-gate.sh`) | Stop | turn-end blocking decisions; block-cap semantics are synchronous |
| `handle-subagent-stop.sh` | SubagentStop | decision output consumed this turn |
| `handle-user-prompt-submit.sh` | UserPromptSubmit | blocking + `sessionTitle` + inbox `additionalContext` needed this turn |
| session-start wrappers (2 kinds) | SessionStart | bootstrap context must arrive this session; async delivery lands next turn, useless |
| `handle-compact.sh` | PreCompact | memo save must complete BEFORE compaction runs |
| PostCompact handler | PostCompact | memo restoration reaches the turn as `systemMessage` |
| `handle-stop-failure.sh` | StopFailure | immediate context at error time |
| CwdChanged / FileChanged handlers | CwdChanged, FileChanged | `CLAUDE_ENV_FILE` writes must complete before the env is used |
| SessionEnd handler | SessionEnd | process runs during exit; an async hook may not run at all at exit |
| `handle-teammate-idle.sh` / `handle-task-completed.sh` | TeammateIdle, TaskCompleted | blocking keep-working / reject decisions |

Already async (recorded, no change): PostToolUse 4 entries, harness-observe 4 entries, Stop
security 2 entries, InstructionsLoaded.

### Timeout doctrine correction

- **REQ-HSA-010** — The timeout section of `hooks-system.md` (live
  `.claude/rules/moai/core/hooks-system.md` AND template mirror
  `internal/template/templates/.claude/rules/moai/core/hooks-system.md`) shall document the
  official UserPromptSubmit semantics: 30 seconds is the **default** for UserPromptSubmit command
  hooks with **no documented per-event maximum** — headroom above 30s exists where a hook needs
  it, within the general hook timeout ceiling — replacing both the table row's
  "15s | 30s (reduced max)" framing and the runtime-ceiling bullet's "lowers `command` / `http` /
  `mcp_tool` hooks to 30s" cap wording. The row's shipped-value cell shall state the value
  actually shipped by this SPEC (30s).

- **REQ-HSA-011** — The UserPromptSubmit registrations (`handle-user-prompt-submit.sh`,
  `handle-codex-review-entry.sh`) shall carry `timeout: 30` in the template settings and in the
  live settings of every tree this card lands on. Observed three-way drift at plan time: local
  primary live settings already at 30 (raised by the leader), this card worktree's live copy at
  15/5, template at 5/5.

### Measurement protocol

- **REQ-HSA-012** — When the run phase begins, every synchronous hook registration shall be
  measured 10 times under the current machine load (no artificial load), with per-run timings
  recorded into a measurement table committed at `.moai/reports/t1780/hook-sync-timings.md`;
  prescription order is fixed: (1) shell fast-paths first, (2) then timeout raises within the
  per-event runtime caps, (3) async only for hooks the evidence table passes. The
  UserPromptSubmit hard-timeout count under multi-lane load shall be 0.

## 3. Acceptance Criteria

Tier M — the full AC matrix with Given-When-Then scenarios, RED-now / green-path two-cell
adoptions, and the evidence ledger (per `verification-completeness.md` §2) lives in
[acceptance.md](./acceptance.md).

## 4. Constraints

- **C1 — t1775 sequencing (HARD).** Run-phase edits to the template files
  `settings.json.tmpl` and `hooks-system.md` (and, by Template-First duality, their live-tree
  mirrors) are blocked until card t1775 (lane-36, editing the SAME template files) lands on main
  and main is absorbed into this card's branch. Plan artifacts are authored now; the run phase
  begins only after the t1775 gate passes (plan.md M2).
- **C2 — Template-First.** Every live-tree change has its template mirror; live and template land
  in the same commit pair so neither surface drifts.
- **C3 — REQ-GBN-002 preserved.** The fresh-FAIL enforcement path stays synchronous; async is
  forbidden for `handle-codex-review-entry.sh`.
- **C4 — No artificial load.** The measurement protocol runs under the current machine load only.
- **C5 — Hook independence.** The shell fast path obeys `hook-independence.md` §7; no new shared
  failure mode; fail-open to the CURRENT behavior, never to a block.
- **C6 — no-block semantics respected.** The entry-side fresh-FAIL handling follows the live
  REQ-RG-001/002 amendment (additionalContext delivery, once per key) — the fast path changes
  WHO decides (shell probe vs binary) and WHEN (skip the binary on ALLOW states), not the
  handler's decision semantics.

## 5. Out of Scope

### Out of Scope — decision item ④ reopening

- The synchronous registration of `handle-codex-review-entry.sh` and the existing gate-OFF
  fast path's shape are decided (board d-20261015T030xx) and are not renegotiated by this SPEC.

### Out of Scope — Stop-gate kick arm changes

- The Stop gate's synchronous block arm and its kick dedup mechanics are untouched; the entry
  shell kick composes with them through the shared kick marker and changes nothing on the Stop
  side.

### Out of Scope — speculative timeout raises

- No timeout value changes without the measurement table's evidence; raises beyond the corrected
  doctrine's per-event defaults are prescribed only by the fixed prescription order.

### Out of Scope — async conversions beyond the three decided

- Every registration not named in REQ-HSA-001..003 stays synchronous; the keep-sync table is the
  boundary. New async candidates discovered by the measurement table become follow-up cards, not
  scope growth here.

### Out of Scope — runtime async-delivery mechanism changes

- Claude Code's async hook delivery contract (additionalContext-only, next turn) is upstream
  behavior; this SPEC adapts handlers to it and does not modify or document around it.

## 6. Cross-references

- Card t1780 (operator directive 2026-10-15) — the controlling card; board decisions
  d-20261015T030xx and d-20261015T024551Z-78aa.
- SPEC-GATE-BOTTLENECK-001 — REQ-GBN-002 (delayed verdict, entry + kick arms).
- SPEC-REVIEW-GATE-REPAIR-001 — REQ-RG-001/002 (entry no-block amendment, per-key delivery).
- SPEC-HARNESS-RATCHET-REWIRE-001 — REQ-HRR-001 (PostToolUseFailure failure recording; preserved
  unchanged by REQ-HSA-003).
- SPEC-CODEX-GATE-SCOPE-001 — the card/tree scope discriminator the fast path falls through on.
- `.claude/rules/moai/development/hook-independence.md` — authoring checklist (§7).
- `.claude/rules/moai/development/verification-completeness.md` §2 — two-cell AC discipline.
- `.claude/rules/moai/core/hooks-system.md` — the timeout doctrine this SPEC corrects.
