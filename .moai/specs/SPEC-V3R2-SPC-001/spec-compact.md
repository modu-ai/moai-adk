# spec-compact — SPEC-V3R2-SPC-001

One-page reference card. For full text see `spec.md`, `research.md`, `plan.md`, `acceptance.md`, `tasks.md`.

## Identity

| Field | Value |
|-------|-------|
| ID | SPEC-V3R2-SPC-001 |
| Title | EARS + hierarchical acceptance criteria |
| Phase | v3.0.0 — Phase 5 — Harness + Evaluator |
| Priority | P0 Critical |
| Status (spec.md) | draft |
| Status (plan.md) | audit-ready |
| Breaking | true (BC-V3R2-011) |
| Lifecycle | spec-anchored |

## Goal in 3 lines

Upgrade SPEC acceptance-criteria from flat list to hierarchical (parent → child) tree per Agent-as-a-Judge DevAI shape (55 → 365 sub-requirement ratio). Maximum tree depth = 3. Children inherit parent Given. 100% back-compat: legacy flat ACs auto-wrap as 1-child parents.

## REQ Index (18 total)

| Modality | REQ IDs |
|----------|---------|
| Ubiquitous (5.1) | 001, 002, 003, 004, 005, 006, 007 |
| Event-driven (5.2) | 010, 011, 012, 013 |
| State-driven (5.3) | 020, 021 |
| Optional (5.4) | 030, 031 |
| Complex (5.5) | 040, 041, 042 |

## AC Index (17 total — 14 from spec.md §6 retained as flat baseline + 3 plan-phase additions; 4 of 17 self-demonstrate hierarchy)

| AC ID | Self-demonstrates hierarchy? | Mapped REQs |
|-------|-------------------------------|-------------|
| AC-SPC-001-01 | YES (3 children .a/.b/.c) | 001, 005, 007, 011 |
| AC-SPC-001-02 | YES (2 children .a/.b) | 010, 020 |
| AC-SPC-001-03 | flat | 006 |
| AC-SPC-001-04 | flat | 012 |
| AC-SPC-001-05 | flat | 003, 021 |
| AC-SPC-001-06 | flat | 013 |
| AC-SPC-001-07 | flat | 030 |
| AC-SPC-001-08 | flat | 031 |
| AC-SPC-001-09 | YES (3 children .a/.b/.c) | 010, 040 |
| AC-SPC-001-10 | flat | 041 |
| AC-SPC-001-11 | flat | 020 |
| AC-SPC-001-12 | flat | 004 |
| AC-SPC-001-13 | flat | 004, 042 |
| AC-SPC-001-14 | YES (3 children .a/.b/.c) | 001, 003 |
| AC-SPC-001-15 | flat | 002 |
| AC-SPC-001-16 | flat | 007 |
| AC-SPC-001-17 | flat | 042 |

## Files Touched in Run Phase

| File | Purpose | Milestone |
|------|---------|-----------|
| `internal/spec/parser_test.go` | Add `BenchmarkParse365Leaves` | M2 |
| `internal/spec/testdata/perf-365-leaves/spec.md` | Perf fixture | M2 |
| `internal/spec/testdata/flat-format-malformed-*/spec.md` | Frontmatter edge cases | M2 |
| `internal/cli/spec_view.go` | `--shape-trace` audit | M3 |
| `internal/cli/spec_view_test.go` | CLI test for trace fields | M3 |
| `.claude/rules/moai/workflow/spec-workflow.md` | Hierarchical schema block | M3 |
| `.claude/skills/moai-workflow-spec/SKILL.md` | Hierarchical AC subsection | M3 |
| `.claude/rules/moai/core/zone-registry.md` | CONST-V3R2-001 cross-link | M3 |
| `.moai/specs/SPEC-V3R2-MIG-001/spec.md` | Handoff note | M4 |
| `.moai/specs/SPEC-V3R2-SPC-001/canary-v2-reparse.txt` | Canary evidence | M5 |
| `.moai/specs/SPEC-V3R2-SPC-001/progress.md` | Final status update | M5 |
| `internal/spec/ears.go` | `@MX:NOTE` annotation only | M5 |

## Files Already on Main (NOT touched in run phase)

- `internal/spec/ears.go` (165 LOC) — Acceptance struct, MaxDepth, GenerateChildID, Depth, InheritGiven, IsLeaf, CountLeaves, ValidateDepth, ExtractRequirementMappings, ValidateRequirementMappings.
- `internal/spec/errors.go` (45 LOC) — DuplicateAcceptanceID, MaxDepthExceeded, DanglingRequirementReference, MissingRequirementMapping.
- `internal/spec/parser.go` (350 LOC) — full hierarchical parser with auto-wrap.
- `internal/spec/lint.go` (857 LOC) — REQ↔AC coverage walks tree.
- `internal/cli/spec_view.go` (158 LOC) — `moai spec view --shape-trace` command.

## Adjacent SPECs

| Type | SPEC | Relationship |
|------|------|--------------|
| Blocked by | SPEC-V3R2-CON-001 | EARS modality FROZEN (no conflict; SPC-001 amends shape only) |
| Gates | SPEC-V3R2-CON-002 | 5-layer amendment safety gate |
| Blocks | SPEC-V3R2-SPC-003 | SPEC linter consumes hierarchical schema |
| Blocks | SPEC-V3R2-HRN-002 | Sprint Contract durable per-leaf state |
| Blocks | SPEC-V3R2-HRN-003 | evaluator-active per-leaf scoring |
| Blocks | SPEC-V3R2-MIG-001 | Optional cosmetic AC rewrite |

## Quality Gates Summary

Plan-phase: research ≥25 anchors (48 achieved), all 18 REQs mapped, ≥3 ACs hierarchical, plan-auditor PASS at iteration ≤2.

Run-phase: `go test ./internal/spec/...` green, `go test -bench BenchmarkParse365Leaves` <500ms, Canary evidence committed, HumanOversight approval recorded.

Sync-phase: spec-workflow.md amended, SKILL.md updated, zone-registry cross-linked, CHANGELOG updated.

## Hierarchical Schema (canonical example)

```
- AC-<DOMAIN>-<NNN>-<NN>: Given <parent context>
  - AC-<DOMAIN>-<NNN>-<NN>.a: When variant-A, Then result-A. (maps REQ-...)
  - AC-<DOMAIN>-<NNN>-<NN>.b: Given <override>, When variant-B, Then result-B. (maps REQ-...)
    - AC-<DOMAIN>-<NNN>-<NN>.b.i: When sub-case 1, Then result-1. (maps REQ-...)
    - AC-<DOMAIN>-<NNN>-<NN>.b.ii: When sub-case 2, Then result-2. (maps REQ-...)
- AC-<DOMAIN>-<NNN>-<NN+1>: Given ..., When ..., Then ... (maps REQ-...)   # flat still valid
```

## Risks (top 3)

| Risk | Severity | Mitigation milestone |
|------|----------|-----------------------|
| CON-002 paperwork incomplete | HIGH | M5 — Canary + HumanOversight |
| `--shape-trace` field emission gap | MEDIUM | M3 — audit + CLI test |
| 365-leaf perf unmeasured | LOW | M2 — benchmark |

End of compact.
