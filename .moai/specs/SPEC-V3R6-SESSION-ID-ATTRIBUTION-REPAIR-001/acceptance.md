---
id: SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001
title: "Acceptance Criteria — Session-ID attribution dead-feature repair"
version: "0.1.0"
status: in-progress
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli/session.go; internal/hook/session_start.go; internal/session/registry.go"
lifecycle: spec-anchored
tags: "session, attribution, multi-session, coordination, race-attribution, doctrine"
era: V3R6
---

# Acceptance Criteria — SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001

## §A. Acceptance Criteria Matrix

### §A.1 P1 — Write-path reliability (gating predecessor)

**AC-WPR-001 (MUST, M1)** — `moai session doctor` exists and reports
registry file existence + current session presence.

- **Given** a host with at least one active Claude Code session
- **When** the user runs `moai session doctor`
- **Then** the command reports (a) whether `.moai/state/active-sessions.json`
  exists, (b) whether the current process's session_id is in it, (c) if
  absent, the likely root cause (empty `input.SessionID`, hook silent-exit,
  or write failure).
- **Traceability:** REQ-WPR-001.

**AC-WPR-002 (MUST, M1)** — Empty-list diagnostic explains root causes.

- **Given** `moai session list --json` returns `[]` on a host with an active session
- **When** the user runs the diagnostic
- **Then** the output enumerates the likely root causes (empty
  `input.SessionID`, hook wrapper fallback silent-exit, registry write
  failure).
- **Traceability:** REQ-WPR-002.

**AC-WPR-003 (MUST, M1)** — Empty-SessionID warning emitted.

- **Given** the SessionStart hook handler receives `input.SessionID == ""`
- **When** the handler runs
- **Then** a warning is emitted to stderr (non-blocking, hook still exits 0).
- **Traceability:** REQ-WPR-003.

**AC-WPR-004 (MUST, M1 — GATE)** — M1 root cause documented before M2 starts.

- **Given** the P1 investigation is complete
- **When** the implementer marks M2 ready-to-start
- **Then** research.md §D contains an empirically-reproduced root cause
  (test or observed runtime state, not assumption).
- **Traceability:** REQ-WPR-004.

### §A.2 P2 — Read-path addition

**AC-RDP-001 (MUST, M2)** — `moai session current` subcommand exists.

- **Given** the CLI is built
- **When** the user runs `moai session --help`
- **Then** `current` is listed alongside register/heartbeat/deregister/list/purge
  (6 subcommands total).
- **Traceability:** REQ-RDP-001.

**AC-RDP-002 (MUST, M2)** — `current` returns UUID when runtime exposes it.

- **Given** the runtime exposes `session.id` (via the SessionStart side-channel
  file or env var)
- **When** the user runs `moai session current`
- **Then** the command outputs the UUID and exits 0.
- **Traceability:** REQ-RDP-002.
- **P1-outcome-conditional note (D5 de-risk):** the happy-path Given clause
  presupposes the SessionStart side-channel write succeeds, which itself
  depends on the P1 investigation outcome (M1). If P1 reveals the side-channel
  write is ALSO gated by `input.SessionID != ""` (the same gate at
  `internal/hook/session_start.go:66` that governs the registry Register
  call), then in the empty-SessionID scenario AC-RDP-002's happy-path is
  unsatisfiable until the M3 additionalContext injection is made
  unconditional (runs even when `input.SessionID == ""`, which would require
  an alternative UUID source). In that case, defer AC-RDP-002 GREEN to
  post-M3 and rely on AC-RDP-003 (exit 1 + diagnostic) + AC-RDP-006
  (canonical fallback) as the degradation contract for M2. The P1 outcome
  recorded in research.md §D determines which branch applies.

**AC-RDP-003 (MUST, M2)** — `current` returns exit 1 + diagnostic when UUID unavailable.

- **Given** the runtime does NOT expose `session.id` to the CLI subprocess
- **When** the user runs `moai session current`
- **Then** the command exits 1 with a stderr diagnostic pointing to the
  SessionStart-hook injection path as the alternative.
- **Traceability:** REQ-RDP-003.

**AC-RDP-004 (MUST, M3)** — SessionStart hook injects UUID via additionalContext.

- **Given** a SessionStart hook invocation with `input.SessionID != ""`
- **When** the hook handler runs
- **Then** the hook response includes `additionalContext` (or
  `hookSpecificOutput`) carrying the orchestrator's own UUID.
- **Traceability:** REQ-RDP-004.

**AC-RDP-005 (MUST, M3)** — Injection is strictly additive.

- **Given** the existing SessionStart behavior (Register → Purge → Query →
  stderr surface)
- **When** the `additionalContext` injection is added
- **Then** all existing SessionStart tests pass unchanged; the hook timeout
  stays ≤ 5s.
- **Traceability:** REQ-RDP-005.

**AC-RDP-006 (SHOULD, M2)** — `current` degrades to canonical fallback.

- **Given** the registry is empty (P1 unresolved or runtime non-exposure)
- **When** the user runs `moai session current`
- **Then** the command emits the canonical fallback string
  (REQ-FBC-001) rather than an opaque error.
- **Traceability:** REQ-RDP-006. SHOULD because exit 1 + diagnostic
  (AC-RDP-003) is also acceptable; fallback emission is a nicer-to-have.

### §A.3 P3 — Fallback doctrine canonicalization

**AC-FBC-001 (MUST, M4+M5)** — Exactly one canonical fallback string.

- **Given** the doctrine surfaces (session-handoff.md Block 2 + moai.md §8)
- **When** the implementer greps for `source_session_id: <...>` patterns
- **Then** exactly ONE canonical string appears:
  `source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>`.
- **Traceability:** REQ-FBC-001, REQ-FBC-002.

**AC-FBC-002 (MUST, M4+M5)** — Byte-parity across SSOT surfaces.

- **Given** the canonical string is defined
- **When** the implementer compares the string in
  `.claude/rules/moai/workflow/session-handoff.md` Block 2 vs
  `.claude/output-styles/moai/moai.md` §8
- **Then** the two strings are byte-identical.
- **Traceability:** REQ-FBC-002.

**AC-FBC-003 (MUST, M4+M5)** — 3 non-canonical variants eliminated on
canonical doctrine surfaces.

- **Given** the 4 variants enumerated in research.md §C (1 canonical + 3
  non-canonical on canonical doctrine surfaces: rules tree + output-styles +
  template mirrors)
- **When** the implementer greps the canonical doctrine surfaces (the
  load-bearing 4-variant command in research.md §C, which EXCLUDES
  `.moai/specs/` to avoid self-referential pollution) for
  `source_session_id: <...>` patterns after M4+M5
- **Then** only two distinct spellings remain: the canonical fallback string
  (AC-FBC-001) and the rewritten happy-path UUID-source slot
  `source_session_id: <UUID from moai session current>` (D3 disposition for
  variants #2 and #10). The third non-canonical variant
  (`<orchestrator-uuid-here>` at session-handoff.md L96 illustrative example)
  is collapsed to the canonical fallback. The 3 non-canonical variants from
  research.md §C each return 0 matches in their original form.
- **Traceability:** REQ-FBC-003. Note: the broader-context sweep (14 variants
  including `.moai/docs/` + memory) is NOT in elimination scope; historical
  memory entries are append-only per spec.md §G.

**AC-FBC-004 (MUST, M4+M5)** — Template mirrors in parity.

- **Given** CLAUDE.local.md §2 Template-First Rule
- **When** the implementer compares local doctrine vs template mirror
- **Then** the canonical string is byte-identical across both pairs:
  - `.claude/rules/moai/workflow/session-handoff.md` ↔ `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`
  - `.claude/output-styles/moai/moai.md` ↔ `internal/template/templates/.claude/output-styles/moai/moai.md`
- **Traceability:** REQ-FBC-002, CLAUDE.local.md §2.

**AC-FBC-005 (SHOULD, M4+M5)** — Future-edit doctrine citation discipline.

- **Given** a doctrine edit in M4/M5 that references the fallback pattern
  (e.g., an edit to session-handoff.md Block 2 or moai.md §8 Pre-emit
  self-check item 9)
- **When** the implementer greps the edited doctrine surface for the literal
  canonical fallback string
  (`grep -F 'source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>' <edited-file>`)
- **Then** the grep returns ≥1 match AND every `source_session_id: <...>`
  reference in the edited file that means "the fallback" cites the canonical
  string verbatim (no paraphrase, no abbreviation, no new variant spelling).
  No new `source_session_id: <...>` variant spelling is introduced by the
  edit.
- **Traceability:** REQ-FBC-004. SHOULD because this is a forward-looking
  discipline check; the core P3 outcome is AC-FBC-001/002/003 (canonical
  string present + byte-parity + 3 non-canonical eliminated).

### §A.4 Cross-cutting — multi-session coordination preservation

**AC-MSC-001 (MUST, M2+M3)** — Existing 5 subcommands behaviorally unchanged.

- **Given** the existing `moai session register/heartbeat/deregister/list/purge` semantics
- **When** the implementer runs the existing test suite
- **Then** all pre-existing session tests pass unchanged.
- **Traceability:** REQ-MSC-001.

**AC-MSC-002 (MUST, M3)** — `FormatStderrReminder` semantics unchanged.

- **Given** `internal/session/registry.go:424-448` `FormatStderrReminder`
- **When** the implementer runs the registry tests
- **Then** the function still returns empty when `others == 0` (L431-432)
  and the existing tests pass.
- **Traceability:** REQ-MSC-002.

**AC-MSC-003 (SHOULD, M6)** — Resume template cites `moai session current`.

- **Given** the resume template in session-handoff.md Block 2 + moai.md §8
- **When** the implementer reads Block 2
- **Then** the text cites `moai session current` as the primary UUID source
  and the canonical fallback as the degradation path.
- **Traceability:** REQ-MSC-003. SHOULD because the citation is a
  documentation nicety; the core repair (P1+P2+P3 canonicalization) does
  not depend on it.

## §B. Severity Classification

| AC ID | Severity | Rationale |
|-------|----------|-----------|
| AC-WPR-001 | MUST | P1 diagnostic is the gating predecessor; without it M2-M6 cannot start. |
| AC-WPR-002 | MUST | Empty-list root-cause explanation is the user-facing surface of P1. |
| AC-WPR-003 | MUST | Empty-SessionID warning is the observable signal that the write path was bypassed. |
| AC-WPR-004 | MUST (GATE) | M1 root cause documentation is the explicit gate for M2-M6 (REQ-WPR-004). |
| AC-RDP-001 | MUST | The `current` subcommand is the core P2 deliverable. |
| AC-RDP-002 | MUST | UUID resolution when runtime exposes it is the happy path. |
| AC-RDP-003 | MUST | Exit 1 + diagnostic when unavailable is the degradation contract. |
| AC-RDP-004 | MUST | additionalContext injection is the mechanism that surfaces UUID to the orchestrator. |
| AC-RDP-005 | MUST | Strictly-additive injection preserves existing behavior (REQ-RDP-005). |
| AC-RDP-006 | SHOULD | Canonical-fallback emission is nicer-to-have; exit 1 is also acceptable. |
| AC-FBC-001 | MUST | Single canonical string is the core P3 deliverable. |
| AC-FBC-002 | MUST | Byte-parity across SSOT surfaces prevents drift recurrence. |
| AC-FBC-003 | MUST | 3 non-canonical variants eliminated on canonical doctrine surfaces is the measurable P3 outcome. |
| AC-FBC-004 | MUST | Template-mirror parity preserves the Template-First Rule. |
| AC-FBC-005 | SHOULD | Future-edit doctrine citation discipline is forward-looking; core P3 outcome is AC-FBC-001/002/003. Traces to REQ-FBC-004 (added iter-2 to close the orphan-REQ gap flagged by plan-auditor iter-1 D2). |
| AC-MSC-001 | MUST | Existing subcommands unchanged is the behavior-preservation constraint. |
| AC-MSC-002 | MUST | FormatStderrReminder unchanged prevents the AP-1 anti-pattern. |
| AC-MSC-003 | SHOULD | Resume template citation is documentation; core repair does not depend on it. |

**Totals: 18 ACs — 15 MUST + 3 SHOULD.**

## §C. Edge Cases

1. **Headless `-p` invocation without hooks.** The SessionStart hook is
   bypassed; `additionalContext` is never injected; `moai session current`
   returns exit 1 + diagnostic (AC-RDP-003). This is the documented residual
   risk (spec.md §F.3), not a defect.

2. **Compaction mid-session.** After `/clear`, the `additionalContext` is
   lost; the orchestrator no longer "remembers" its UUID. The doctrine
   (M6) instructs the orchestrator to re-query `moai session current` before
   emitting a resume. This is the documented residual risk (spec.md §F.2).

3. **Multiple concurrent sessions on the same host.** `moai session list`
   returns N entries; the orchestrator cannot identify its own entry. This is
   the K2 defect that `moai session current` (M2) resolves — `current` reads
   the side-channel file written by THIS session's own SessionStart hook, not
   the full registry.

4. **`input.SessionID == ""` on first-turn activation.** Some Claude Code
   activation paths may emit an empty session_id. AC-WPR-003 requires the
   hook to emit a warning; the registry write is bypassed (K3). P1
   investigation (M1) determines whether this is the root cause of the empty
   registry (K6).

5. **Handle-session-start.sh 3-tier fallback silent-exit.** All 3 tiers fail
   (PATH, `~/go/bin/moai`, `~/.local/bin/moai`) and the script exits 0. Claude
   Code may silently drop the hook result. P1 investigation (M1) determines
   whether this contributes to the empty registry.

## §D. Quality Gate Criteria (Definition of Done)

- [ ] All 15 MUST ACs PASS (verified by test output, not assertion).
- [ ] 3 SHOULD ACs PASS or explicitly deferred with rationale.
- [ ] `go test ./...` PASS (full suite, no skipped tests in scope).
- [ ] `go vet ./...` clean.
- [ ] `golangci-lint run --timeout=2m` clean.
- [ ] Coverage for `internal/cli/session.go` ≥ 85% (package-level baseline).
- [ ] Coverage for `internal/hook/session_start.go` not decreased from baseline.
- [ ] Byte-parity verification: canonical fallback string identical across 4 files (2 local + 2 template).
- [ ] plan-auditor verdict ≥ 0.80 (Tier M threshold) PASS or PASS-WITH-DEBT.
- [ ] progress.md §E.2/E.3 populated by manager-develop with verbatim test output.
- [ ] No non-canonical `source_session_id: <...>` variant remains in doctrine surfaces (grep returns only canonical).

## §E. Indirect Verification (Sentinel Checks)

These are indirect signals that the repair is complete, not direct ACs:

- **Sentinel-1:** After this SPEC completes, the NEXT paste-ready resume
  message emitted by the orchestrator SHOULD carry a real UUID (or the single
  canonical fallback), not one of the 3 non-canonical doctrine-surface
  variants. This is observable in the wild, not a test.
- **Sentinel-2:** `moai session list --json` on a host with an active session
  SHOULD return a non-empty array (if P1 root cause is "SessionStart was
  bypassed in the diagnostic context", this may remain `[]` in some
  legitimate contexts — not a hard failure).

## §F. Forward-Looking Checks (Out of This SPEC's Scope)

- **FL-1:** A future Go lint rule that detects doctrine drift across SSOT
  surfaces (the deferred track 2 deliverable). This is a SEPARATE independent
  SPEC; do not include it here.
- **FL-2:** Runtime-level session.id exposure (upstream Claude Code request).
  Out of MoAI-ADK's control; documented as residual risk.
