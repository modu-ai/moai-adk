---
id: SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001
title: "Implementation Plan — Session-ID attribution dead-feature repair"
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

# Implementation Plan — SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001

## §A. Context

This plan implements the 3-pillar repair defined in spec.md §B.1. The defect's
essence (spec.md §A.1): write path exists but orchestrator-visible read path
is absent, and the registry runtime is empty. The user-confirmed direction is
**A (full repair)** preserving the multi-session coordination Layer 2 race-attribution
purpose.

**Tier: M** (standard). **cycle_type: tdd** (RED-GREEN-REFACTOR; per
`quality.yaml` `development_mode: tdd`). Files in scope: 4-6 edit + test.
Risk: Med (compaction UUID-loss limitation persists; runtime non-exposure
root barrier — documented as residual risk in spec.md §F).

## §B. Known Issues (carried forward from diagnosis)

| ID | Issue | File:line (HEAD 12e20d190) | Severity |
|----|-------|----------------------------|----------|
| K1 | No `current`/`whoami`/`uuid` query subcommand | `internal/cli/session.go:43-47` (5 subcommands: register/heartbeat/deregister/list/purge) | High |
| K2 | `list` cannot identify "which entry is me" | `internal/cli/session.go:116-155` (returns full entry list) | High |
| K3 | Write path exists but gated on `input.SessionID != "" && input.ProjectDir != ""` | `internal/hook/session_start.go:66` (gate), `:804` (`reg.Register` call) | Med |
| K4 | `FormatStderrReminder` is others-only, returns empty when `others == 0` | `internal/session/registry.go:424` (func), `:431-432` (empty return) | Info (NOT a defect — correct behavior, do not repurpose) |
| K5 | No `additionalContext` injection in SessionStart hook | `internal/hook/session_start.go` (zero matches, grep-confirmed) | High |
| K6 | Registry runtime empty: `.moai/state/active-sessions.json` absent, `list --json` returns `[]` exit 0 | `.moai/state/active-sessions.json` (ls exit 1) | High |
| K7 | 4 distinct `source_session_id` variants on canonical doctrine surfaces (1 canonical + 3 non-canonical); 14 in the broader-context sweep including `.moai/docs/` + memory | see research.md §C | Med |
| K8 | `session.id` only in hook stdin JSON, not runtime-exposed to LLM | `internal/cli/hook.go:675,699,788` | Info (root barrier, documented) |

## §C. Pre-flight Checks

Before any implementation work:

1. Confirm HEAD is `12e20d190` or later (no regressions in session subsystem since).
2. Re-grep K1-K8 citations against current HEAD (spec.md/research.md citations
   are baseline-attributed to `12e20d190`; if HEAD shifts, re-verify).
3. Confirm `handle-session-start.sh` 3-tier fallback behavior (PATH → `~/go/bin/moai`
   → `~/.local/bin/moai` → silent `exit 0`) is intact — this is the P1
   investigation target.
4. Confirm the 4 canonical-doctrine-surface variants are still enumerated
   identically (research.md §C — the 4-variant grep EXCLUDING `.moai/specs/`).
   The broader-context sweep (14 variants including `.moai/docs/` + memory)
   is informational only and does NOT gate AC-FBC-003.

## §D. Constraints (recap from spec.md §D)

- Behavior preservation of existing 5 subcommands (REQ-MSC-001).
- Hook timeout ≤ 5s (additionalContext is O(1) map assignment).
- Byte-parity of canonical fallback across both SSOT surfaces.
- Test isolation via `t.TempDir()` — no writes to real registry.
- Single canonical fallback string anchor.

## §E. Self-Verification (manager-develop deliverable)

The implementer (manager-develop) MUST self-verify the following at run-phase completion (§E.1-E.5 in progress.md will be populated by manager-develop; only §E.1 plan-phase audit-ready signal is owned by manager-spec):

- E.1: plan-phase audit-ready signal — populated by manager-spec at plan-phase close (this file + spec.md + acceptance.md + progress.md all emitted).
- E.2: run-phase evidence (test run output, coverage, build) — populated by manager-develop.
- E.3: run-phase audit-ready signal — populated by manager-develop.
- E.4: sync-phase audit-ready signal — populated by manager-docs.
- E.5: Mx-phase audit-ready signal — populated by manager-docs.

## §F. Milestones (priority-ordered; NO time estimates per agent-common-protocol.md § Time Estimation)

### M1 — P1 Write-path reliability investigation (GATING PREDECESSOR)

**Priority: P0 (gating). cycle_type: tdd (characterization tests first).**

Tasks:
1. Write characterization test reproducing the empty-registry state: simulate
   a SessionStart hook invocation with `input.SessionID != ""` and assert
   whether `reg.Register` is called (RED — expected: it IS called per L804,
   so the test should PASS; if it FAILS, the bug is in the call path itself).
2. Write characterization test for `input.SessionID == ""` scenario (RED —
   expected: Register is NOT called; assert the hook emits no warning today,
   then GREEN with REQ-WPR-003 warning added).
3. Audit `handle-session-start.sh` 3-tier fallback: enumerate the conditions
   under which all 3 tiers fail and the script silent-exits (the `exit 0` at
   the final line). Determine if Claude Code silently drops the hook result.
4. Document the P1 root cause(s) in research.md §D (append-only). The root
   cause MUST be empirically reproduced (test or observed runtime state),
   not assumed.

**Exit criteria:** research.md §D documents the root cause(s); K6
(empty registry) is either reproduced deterministically OR explained as an
environmental artifact (e.g., the diagnostic session ran in a context where
SessionStart hook was bypassed).

**Gating rule (REQ-WPR-004):** M2-M6 milestones MUST NOT be marked
ready-to-start until M1 root cause is documented.

### M2 — `moai session current` subcommand (P2 Stage 1)

**Priority: P1. cycle_type: tdd.**

**M2→M3 dependency note (D4):** the AC-RDP-002 happy-path (UUID resolution
via the side-channel file) is only satisfiable after the M3 side-channel
write lands (the SessionStart `additionalContext` injection). At M2
implementation time, the side-channel does not yet exist, so M2 `current`
MAY initially implement only the degradation path (AC-RDP-003 exit 1 +
diagnostic, AC-RDP-006 canonical fallback) and defer AC-RDP-002 GREEN until
M3. The P1 investigation outcome (research.md §D) determines whether the
side-channel write is even possible in the empty-SessionID scenario (see
AC-RDP-002 P1-outcome-conditional note). If P1 reveals the side-channel is
also `input.SessionID`-gated, M2 ships only the degradation path and
AC-RDP-002 GREEN moves to post-M3.

Tasks:
1. RED: write specification test for `moai session current` returning the
   orchestrator's UUID when the runtime exposes it (via the side-channel file
   that M1/P2-Stage2 writes, OR via env var if available).
2. RED: write specification test for `current` returning exit 1 with stderr
   diagnostic when runtime does NOT expose session.id (REQ-RDP-003).
3. RED: write specification test for `current` returning the canonical
   fallback string when registry is empty (REQ-RDP-006).
4. GREEN: implement `newSessionCurrentCmd()` in `internal/cli/session.go`,
   register it alongside the existing 5 subcommands (L43-47).
5. REFACTOR: extract UUID-resolution logic into a testable helper.

**Exit criteria:** all 3 RED tests pass (AC-RDP-002 GREEN MAY be deferred
to post-M3 per the dependency note above); existing 5 subcommands unaffected
(REQ-MSC-001); `moai session --help` lists 6 subcommands.

### M3 — SessionStart hook `additionalContext` injection (P2 Stage 2)

**Priority: P1. cycle_type: tdd.**

Tasks:
1. RED: write specification test that SessionStart hook response includes
   `additionalContext` (or `hookSpecificOutput`) carrying the orchestrator's
   own UUID when `input.SessionID != ""`.
2. RED: write specification test that the injection is strictly additive —
   existing behavior (Register → Purge → Query → stderr surface) unchanged
   (REQ-RDP-005).
3. GREEN: add `additionalContext` map population in the SessionStart hook
   response, after the existing Register step (L804). Use the `input.SessionID`
   as the UUID source.
4. REFACTOR: extract the injection into a small helper for testability.

**Exit criteria:** 2 RED tests pass; hook timeout stays ≤ 5s (additionalContext
is O(1)); existing SessionStart tests pass unchanged.

### M4 — P3 Fallback doctrine canonicalization (session-handoff.md)

**Priority: P2. cycle_type: tdd (doctrine-edit characterization).**

Tasks:
1. Enumerate all 4 canonical-doctrine-surface variants (research.md §C) and
   classify each:
   - Canonical fallback → keep verbatim.
   - Happy-path UUID-source template slot (`<UUID>` at L77/L159/L689,
     `<orchestrator-uuid-from-current-turn>` at moai.md L653) → rewrite to
     `source_session_id: <UUID from moai session current>` for symmetry with
     the P2 `current` deliverable (D3 disposition).
   - Illustrative-example placeholder (`<orchestrator-uuid-here>` at L96) →
     collapse to the canonical fallback.
2. Edit `.claude/rules/moai/workflow/session-handoff.md` Block 2 environment
   fallback clause to use the canonical string verbatim (REQ-FBC-001, REQ-FBC-002).
3. Edit the template mirror at
   `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`
   identically (byte-parity with the local doctrine file per CLAUDE.local.md
   §2 Template-First Rule).
4. Verify the 3 non-canonical variants are eliminated on canonical doctrine
   surfaces (the 4-variant grep command in research.md §C, EXCLUDING
   `.moai/specs/`, returns only the canonical fallback + the rewritten
   happy-path slot).

**Exit criteria:** single canonical fallback string in both local and
template session-handoff.md; canonical-surface grep for the 3 non-canonical
variants returns 0 matches in their original form.

### M5 — P3 Fallback doctrine canonicalization (moai.md §8)

**Priority: P2. cycle_type: tdd (doctrine-edit characterization).**

Tasks:
1. Edit `.claude/output-styles/moai/moai.md` §8 Block 2 checklist item to use
   the identical canonical string (REQ-FBC-002 byte-parity with M4).
2. Edit the template mirror at
   `internal/template/templates/.claude/output-styles/moai/moai.md` identically.
3. Verify byte-parity between session-handoff.md Block 2 and moai.md §8
   (the canonical string is identical across both surfaces).

**Exit criteria:** canonical string present in both SSOT surfaces and both
template mirrors; 4 files in byte-parity for the canonical string token.

### M6 — Update resume template citation (REQ-MSC-003) + sync-phase prep

**Priority: P2. cycle_type: tdd.**

Tasks:
1. Edit session-handoff.md Block 2 to cite `moai session current` as the
   primary UUID source (REQ-MSC-003).
2. Edit moai.md §8 Block 2 checklist identically.
3. Update the corresponding template mirrors.
4. Run the full test suite + lint + vet to confirm no regressions.
5. Prepare sync-phase artifacts (CHANGELOG entry, frontmatter transition
   `draft → planned` per the Status Transition Ownership Matrix).

**Exit criteria:** `go test ./...` PASS; `go vet ./...` clean; `golangci-lint`
clean; doctrine surfaces cite `moai session current` + canonical fallback;
plan-auditor verdict PASS or PASS-WITH-DEBT (≥ 0.80 Tier M threshold).

## §G. Anti-Patterns to Avoid

- **AP-1: Repurposing `FormatStderrReminder` as the self-UUID surface.** That
  function's semantics (warn about OTHER sessions, empty return when
  `others == 0`) are correct as-is (REQ-MSC-002). Do NOT change its return
  signature or behavior.
- **AP-2: Skipping M1 P1 investigation and jumping to M2.** Without M1, the
  registry stays empty and M2's `current` returns nothing useful.
  REQ-WPR-004 explicitly gates M2-M6 on M1.
- **AP-3: Hand-writing a NEW fallback variant in doctrine edits.** Every
  doctrine edit referencing the fallback MUST use the canonical string
  verbatim (REQ-FBC-004 / AC-FBC-005). This is exactly the drift that
  produced the 4 doctrine-surface variants (and 14 broader-context strings)
  in the first place.
- **AP-4: Using `sed -n 'NNN,MMM'p` windowed reads to verify line counts.**
  Per `feedback_windowed_grep_undercount_authoring`, use content-anchored
  `grep -n` instead — windowed reads undercount by excluding boundary lines.
- **AP-5: Trusting the baseline "9 variants" count OR the iter-1 "14
  variants" count without re-checking the path scope.** The independent
  canonical-surface re-grep (research.md §C, EXCLUDING `.moai/specs/`) found
  4 variants; the broader-context sweep (INCLUDING `.moai/docs/` + memory)
  found 14. Every count claim MUST be re-run with the documented path scope,
  not copied. The iter-1 plan-auditor further flagged the original §C
  command as self-referentially polluted (`.moai/specs/` in path list → 16
  not 14); canonical-surface measurements MUST exclude `.moai/specs/`
  (`feedback_self_referential_grep_pollution`).
- **AP-6: Modifying the 5 existing subcommands' semantics.** REQ-MSC-001
  forbids this; the new `current` subcommand is purely additive.

## §H. Cross-References

- spec.md: `SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001/spec.md` (requirements, scope, exclusions).
- acceptance.md: `SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001/acceptance.md` (AC matrix, Given-When-Then).
- research.md: `SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001/research.md` (5-section evidence format, 14-variant enumeration, P1 investigation scaffold).
- Status Transition Ownership Matrix: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `draft → planned` transition owned by CI/hook (NOT this agent).
- Predecessor: `SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001` (track 2, completed).
- Predecessor: `SPEC-V3R6-MULTI-SESSION-COORD-001` (Layer 1 primitives).
