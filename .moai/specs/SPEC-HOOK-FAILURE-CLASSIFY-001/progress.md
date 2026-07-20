---
id: SPEC-HOOK-FAILURE-CLASSIFY-001
title: "Progress — PostToolUseFailure nested-error classification"
version: "0.1.0"
status: draft
created: 2026-07-17
updated: 2026-07-17
author: manager-spec
tier: S
---

# Progress — SPEC-HOOK-FAILURE-CLASSIFY-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifact set authored (Tier S LEAN): `spec.md` (AC inline §D) + `plan.md` + `progress.md`.
- SPEC ID pre-write self-check: `decomposition: SPEC ✓ | HOOK ✓ | FAILURE ✓ | CLASSIFY ✓ | 001 ✓ → PASS` (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`, executed → PASS). No collision in `.moai/specs/`.
- Root causes verified against source at HEAD `efd423b88`: (1) `classifyError` reads only top-level `input.Error`; (2) `validateInput` sets `session_id = "unknown"` → `trace-unknown.jsonl`.
- 6 REQ (GEARS) / 8 AC. Reuse candidate identified: `decodeToolResponse` (evidence_writer.go).
- Open run-phase investigation (NOT a product-decision clarification): session_id resolution source (M2 — likely `transcript_path`, evidence pending).
- No `[NEEDS CLARIFICATION]` markers — all choices resolved from code evidence.

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
