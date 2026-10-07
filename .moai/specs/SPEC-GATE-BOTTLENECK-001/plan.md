---
id: SPEC-GATE-BOTTLENECK-001
title: "Plan — turn-end review gate bottleneck"
version: "0.1.0"
created: 2026-10-07
updated: 2026-10-07
---

# Plan — SPEC-GATE-BOTTLENECK-001

## A.1 Milestones

- **M1 — tree-keyed verdict reuse (REQ-GBN-001, AC-GBN-001/002/003)**:
  landed (a6cf619b2) — the gate consults the receipt store over the
  resolved scope's key before the RPC; a fresh receipt's verdict is
  reused (skip counted); a miss records a receipt for the next turn.
  All cache paths fail-open.
- **M2 — conditional reproduction contract (REQ-GBN-003)**: cache-hit
  path skips the whole review (landed with M1). The live-review
  zero-findings arm needs a reproduction directive in
  `reviewRequestParams` — gated on confirming the codex-side support.
- **M3 — delayed block (REQ-GBN-002, AC-GBN-004)**: Stop allows and
  starts the review in the background (reuse `moai verify codex-review`);
  enforcement moves to a NEW next-turn-entry hook
  (UserPromptSubmit/PreToolUse wiring in settings.json.tmpl + the hook
  manifest M4). `async:true` is not adopted (it surrenders the block).

## A.5 PRESERVE

- `.github/workflows/**`, the sg guard, the codex reviewer itself.
- The Codex Stop chain's receipt consumption (shared store — do not
  change the receipt shape).

## A.6 Verification plan

- Cache-hit Stop wall-clock vs the pre-change synchronous review
  (AC-GBN-003) — measured per the card's verification clause.
- `TestReviewGate_NoEditTurnAllows` stays green (no no-edit
  self-allow regression).
- Every new path's fail-open is covered by the cache tests'
  failure arms.
