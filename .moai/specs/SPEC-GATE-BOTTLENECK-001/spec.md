---
id: SPEC-GATE-BOTTLENECK-001
title: "Turn-end review gate bottleneck — tree-keyed reuse, delayed block, conditional reproduction"
version: "0.1.0"
status: draft
created: 2026-10-07
updated: 2026-10-07
author: lane-6 (MoAI factory)
priority: P1
phase: "v3.2.0"
module: "internal/cli, internal/template/templates/.claude/hooks/moai"
lifecycle: spec-anchored
tags: "review-gate, cache, async, stop-hook, bottleneck"
card: t1575
tier: S
---

# SPEC-GATE-BOTTLENECK-001 — turn-end review gate bottleneck

## Background (measured, card t1575)

The Claude turn-end review gate (`HandleCodexReviewGate`) runs a SYNCHRONOUS
codex review on every Stop — up to 900 s — and re-reviews the FULL tree on
every turn even when the tree state is identical to the last judgment. The
operator measured 11 of 12 Stop hooks blocked, 14 m 13 s total (leader survey
2026-10-07). The tree-keyed receipt infrastructure the CODEX Stop chain
already consumes (`verify.Key` = HEAD + porcelain digest, `ReceiptState`,
`RecordReceipt`/`CheckReceipt`) exists but the CLAUDE gate never consults it.

## Scope split (card t1575 §3 — dedupe with t1383)

The lane-turn CARD-DIFF scope axis belongs to t1383. This SPEC owns the
CACHE axis (tree-keyed verdict reuse) and the ASYNC axis (delayed block)
only.

## Requirements (EARS)

- REQ-GBN-001 (cache): WHEN the review receipt for the resolved scope's
  tree key (`verify.Key` / card-scope binding) is fresh in the receipt
  store, THE GATE SHALL reuse its verdict without an RPC and record the
  skip; a stale or absent receipt SHALL fall through to a live review that
  then records a receipt.
- REQ-GBN-002 (delayed block): WHEN no fresh receipt exists at Stop, THE
  GATE SHALL allow the turn, start the review in the background, and leave
  the verdict to be enforced at the NEXT turn entry; a fresh FAIL at turn
  entry SHALL block. `async:true` SHALL NOT be used (it surrenders the
  block capability).
- REQ-GBN-003 (conditional reproduction): WHEN the review is skipped on a
  cache hit, or the review finds zero findings, THE reproduction suite
  SHALL be skipped, stated in the review-request contract.
- REQ-GBN-004 (invariants): every path above SHALL preserve the fail-open
  posture (a missing/inconclusive reviewer allows) and SHALL NOT allow a
  no-edit session to self-approve a failing verdict (the tree key covers
  HEAD + porcelain, so an untouched tree cannot shed a fail).

## Acceptance Criteria

- AC-GBN-001 (REQ-GBN-001): two consecutive Stop invocations over an unchanged tree make
  exactly ONE codex RPC; the second reuses the receipt verdict and records
  the skip.
- AC-GBN-002 (REQ-GBN-001, REQ-GBN-004): a stored FAIL verdict for the current tree key blocks at the
  gate without an RPC; editing a tracked file (porcelain change) makes the
  key stale and the next Stop runs a live review.
- AC-GBN-003 (REQ-GBN-001): the wall-clock of a cache-hit Stop is bounded by local
  receipt lookup (sub-second class), versus the pre-change synchronous
  review.
- AC-GBN-004 (REQ-GBN-002, REQ-GBN-004): fail-open holds on every new path: a corrupt/absent cache
  entry and a failed background start both allow.

## 3. Non-Goals

### 3.1 Out of Scope

- The card-diff scope axis (t1383). The Codex Stop chain (already
  receipt-based). Changing the codex reviewer itself.
