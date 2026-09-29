---
id: SPEC-LANE-NOTICE-DIET-001
title: "Lead·lane join-notice multi-locale diet — compress the lane join notice and its embedded standing spawn authority to core form"
version: "0.1.0"
status: in-progress
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/hook"
lifecycle: spec-anchored
tags: "hook, session-start, i18n, kanban, factory, lane-notice, spawn-authority, diet"
tier: S
related_specs: [SPEC-AC-LOCALE-TOKEN-001]
---

# SPEC-LANE-NOTICE-DIET-001 — Lead·lane Join-Notice Multi-Locale Diet

## A. History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-29 | manager-spec | Initial plan-phase draft (card t1335). |

## B. Background

The lane join notice — the SessionStart additionalContext a factory lane or kanban
companion session reads at startup — is assembled in exactly two places:

- `internal/hook/session_start_factory.go` `factoryLaneNotice` — localized
  `laneJoin` / `laneJoinNoCount` line + `"\n\n" + laneSpawnAuthority`
- `internal/hook/session_start_kanban.go` `kanbanCompanionNotice` — localized
  `companionJoin` line + `"\n\n" + laneSpawnAuthority`

`laneSpawnAuthority` (`internal/hook/lane_spawn_authority.go:38`) is an English-only
const (const line 624 bytes, string ≈597 bytes, 3 sentences — measured at `7bef423c0`) carrying: (1) the standing spawn grant with an
inline specialist-mapping parenthetical, (2) the depth-1 seal, (3) a bootstrap-placement
sentence born from the tk8hce incident (card t224). The join lines themselves are
already single-sentence in all 4 locales.

The diet compresses the notice to core form: identity tokens stay (mode + lane
label), the specialist mapping moves into a doctrine pointer, the depth-1 seal
stays explicit, and pure narration shrinks — without reducing the standing grant
(card t224 regression, invariant 1).

Measured at tree `7bef423c0`, branch `WT-bootstrap-notice-diet`.

## C. Requirements (GEARS)

- REQ-LND-001 (Ubiquitous): The lane join notice shall grant a **standing spawn
  authority** — the lane session may use the Agent tool to spawn the phase-required
  specialist without asking the leader or the operator first. The grant must remain
  carried in the join notice itself (bootstrap context), never reduced to a
  pointer-only stub that grants nothing.
- REQ-LND-002 (Ubiquitous): The lane join notice shall carry the depth-1 seal
  explicitly — agents the lane spawns are leaf workers and must not spawn further
  agents.
- REQ-LND-003 (Ubiquitous): The standing spawn authority shall name the **Status
  Transition Ownership Matrix** as the specialist-mapping pointer instead of
  inlining the per-phase mapping (plan-phase artifacts to manager-spec /
  implementation to manager-develop / sync-phase docs to manager-docs), keeping
  the auditors tail ("plus the workflow chain's prescribed auditors") in the
  pointer sentence — the pointer target names the three specialists but no
  auditors, so the tail is load-bearing.
- REQ-LND-004 (Ubiquitous): The standing spawn authority shall remain English-only —
  no locale rows in either i18n table (`session_start_factory_i18n.go`,
  `session_start_kanban_i18n.go`); the join line stays localized in en/ko/ja/zh.
- REQ-LND-005 (Ubiquitous): The authority text shall reach core form by moving
  the specialist mapping to the pointer (mechanically gated by AC-LND-001) and
  by keeping the bootstrap-placement clause (authority is part of the lane's
  bootstrap context, not granted or revoked by peer messages) as a clause, not
  a standalone sentence. The former 2-sentence budget is deliberately NOT a
  gated deliverable (audit D5): a sentence-delimiter or byte-cap discriminator
  is a grammar-shaped check — the unsoundness
  `.claude/rules/moai/development/verification-completeness.md` §2 [HARD]
  names — while the real compression deliverables (mapping → pointer,
  placement sentence → clause) are already mechanically pinned.
- REQ-LND-006 (Event-driven): **When** the session label does not parse
  (`SplitFactoryLaneLabel` / `SplitCompanionLabel` returns `ok=false`), the notice
  builders shall emit the empty string — fail-open, no notice, no error.
- REQ-LND-007 (Ubiquitous): The localized join lines shall preserve their format
  mechanics: explicit `%[n]s` argument-order pins (en/ja/ko count-first, zh
  label-first), no leading or trailing newlines in any i18n field, and the
  count / no-count factory variant split.
- REQ-LND-008 (Ubiquitous): The notice diet shall not modify the out-of-scope
  surfaces: factory `laneNextCardRule` / `laneOwnedCardRule`, all leader-notice
  i18n fields (`leaderHeader`, `leaderIdentity`, `leaderManual`, `entryGuide`,
  `leaderClasses`, `leaderStagger`, …), `session_stale_run.go`, the launcher, and
  `internal/cli/` entirely.

## D. Acceptance Criteria (summary)

Full two-cell AC matrix with the evidence ledger: `acceptance.md`.

- AC-LND-001 — the authority const drops the inline specialist-mapping
  parenthetical (release-blocking).
- AC-LND-002 — the assertion sites in `lane_spawn_authority_test.go` no longer
  assert the removed specialist-name markers and assert the compressed markers
  instead, including the two grant-verb markers that pin the standing grant
  mechanically (release-blocking).
- AC-LND-003 — the full notice test set stays green on the compressed notice
  (regression guard; baseline observed green at `7bef423c0`).

## E. Non-Functional Constraints

- Code comments in the implementation follow `code_comments: en`
  (`.moai/config/sections/language.yaml`).
- Harness level: minimal (Tier S). Development mode: tdd
  (`.moai/config/sections/quality.yaml`).
- The authority stays an English-only const per the two-audience rule
  (additionalContext is agent-facing; see the comment in `lane_spawn_authority.go`).

## F. Out of Scope

### Out of Scope — leader-notice i18n fields

- All leader-side fields in `session_start_factory_i18n.go` and
  `session_start_kanban_i18n.go` (`leaderHeader`, `leaderIdentity`, `leaderManual`,
  `entryGuide`, `leaderClasses`, `leaderStagger`, `laneNextCardRule`,
  `laneOwnedCardRule`, and siblings) stay byte-identical.

### Out of Scope — non-notice surfaces

- `internal/hook/session_stale_run.go` (stale-run advisory) is untouched.
- The launcher and `internal/cli/` are untouched; the join strings appear in no
  launcher/CLI file (measured: zero grep hits at `7bef423c0`).
- The `.claude/hooks/moai/handle-*.sh` wrappers are unaffected (the notice is Go
  source compiled into the moai binary; no template mirror applies).
- No run id is added to the join line (RESOLVED — operator decision 2026-09-29:
  keep as-is; see `plan.md` §I).

### Out of Scope — doctrine files

- The Status Transition Ownership Matrix prose in
  `.claude/rules/moai/development/spec-frontmatter-schema.md` and the four doctrine
  files naming it (`kanban-dispatch.md`, `agent-common-protocol.md`,
  `moai-constitution.md`, `manager-lead.md`) are read as the pointer target, never
  edited by this SPEC.
