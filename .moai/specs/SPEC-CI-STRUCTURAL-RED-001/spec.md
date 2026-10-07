---
id: SPEC-CI-STRUCTURAL-RED-001
title: "CI structural red unblock — windows build guard + codemaps stamp re-anchor"
version: "0.1.0"
status: draft
created: 2026-10-07
updated: 2026-10-07
author: lane-6 (MoAI factory)
priority: P1
phase: "v3.2.0"
module: "internal/runtime, .moai/project/codemaps"
lifecycle: spec-anchored
tags: "ci, windows, build-tag, codemaps, stamp, merge-blocker"
card: t1563
tier: S
---

# SPEC-CI-STRUCTURAL-RED-001 — CI structural red unblock

## Background

Two main-tip defects make every pull-request merge-ref build red,
blocking the whole merge queue (card t1563; the dispatch's original
sg-guard premise was measured FALSE — the cited run's guard passed
`PASS: ast-grep 0.40.5`; the actual failures follow).

- **D1 windows build**: `internal/runtime/audit_counter_fifo_test.go`
  (landed with card t1500, #1775) calls `syscall.Mkfifo`, which does not
  exist on windows — the file carries a runtime `GOOS` skip, but the
  compile fails before any runtime check runs. Measured at main tip
  cad44a751: `GOOS=windows go vet ./internal/runtime/` →
  `undefined: syscall.Mkfifo`.
- **D5 graph-freshness**: `.moai/project/codemaps/provenance.json` on
  origin/main names commit `faa179e7e…`, which lives only on the
  un-merged branch `WT-self-improve-protected-zone` (card t1510's tree).
  The stamp-reachability guard (SPEC-STAMP-REACHABILITY-001, incident
  0d15864ae90b) correctly fails every merge-ref checkout that cannot
  reach it.

## Requirements (EARS)

- REQ-CIS-001: WHEN the windows toolchain compiles the repository test
  sources, THE SYSTEM SHALL compile every package without undefined
  POSIX-only symbols.
- REQ-CIS-002: WHEN the FIFO regression guard cannot run (windows,
  FIFOs POSIX-only), THE SYSTEM SHALL keep the test name discoverable
  via a same-name windows twin that skips with the reason, so the swept
  test set stays visible on every platform.
- REQ-CIS-003: WHEN the codemaps provenance names a commit, THAT commit
  SHALL be reachable from main's history (no orphan-bound stamps; the
  t279 incident class).

## Acceptance Criteria

- AC-CIS-001: `GOOS=windows go vet ./internal/runtime/` exits 0 on this
  tree. (RED at base cad44a751 — measured 2026-10-07, this session.)
- AC-CIS-002: `GOOS=windows go test ./internal/runtime/ -run
  TestCountAuditRoundsDoesNotBlockOnFifoEvidence` reports the test as a
  skip, not an absence.
- AC-CIS-003: `git merge-base --is-ancestor $(jq -r .commit_sha
  .moai/project/codemaps/provenance.json) HEAD` exits 0, and the stamped
  commit is this branch's tip (reachable after merge).

## 3.x Out of Scope

- The sg guard (measured innocent), the data race in
  `TestSessionStartMemoryBudget_JoinBound` (flaky, CI-conditions —
  follow-up card recommended), PR-specific parity/doctrine fixes for
  card t1542's own branch (repaired on that branch separately).
