---
id: SPEC-V3R6-LEGACY-CLEANUP-001
title: "Progress — v2.x agency keyword residual cleanup"
version: "0.1.0"
status: draft
created_at: 2026-05-23
updated_at: 2026-05-23
author: manager-spec
tier: M
---

# Progress — SPEC-V3R6-LEGACY-CLEANUP-001

## Status

| Phase | Status | Commit | Notes |
|-------|--------|--------|-------|
| plan | in-progress (creating 4 artifacts) | TBD | This commit |
| run M1 (backup + skills/rule) | not started | — | 6 files: backup dir + 5 source files |
| run M2 (docs-site ko + en) | not started | — | 10 files |
| run M3 (docs-site ja + zh) | not started | — | 10 files |
| run M4 (root markdown + verification) | not started | — | 6 files |
| sync | not started | — | CHANGELOG `[Unreleased]` entry only (B12 rule) |

## Milestone Tracker

### M1 — Backup + Skills + Rule

- [ ] T1.1: Create backup directory + copy 31 in-scope files
- [ ] T1.2: Generate manifest.json with path + sha256 + bytes
- [ ] T1.3: Inspect + classify 5 skill/rule files
- [ ] T1.4: Apply surgical edits
- [ ] T1.5: Spot-verify PRESERVE sample (5 files)
- [ ] T1.6: M1 commit

### M2 — docs-site ko + en

- [ ] T2.1: Inspect + classify 10 docs-site files (ko + en)
- [ ] T2.2: Apply ko edits
- [ ] T2.3: Mirror en edits
- [ ] T2.4: Parity tracker note
- [ ] T2.5: Hugo build verification
- [ ] T2.6: M2 commit

### M3 — docs-site ja + zh

- [ ] T3.1: Mirror ja edits (with translation quality review)
- [ ] T3.2: Mirror zh edits (with translation quality review)
- [ ] T3.3: Cross-locale parity verification
- [ ] T3.4: Hugo build verification
- [ ] T3.5: Global symmetric count grep
- [ ] T3.6: M3 commit

### M4 — Root markdown + verification

- [ ] T4.1: CHANGELOG pre-v3.0 vs v3.0+ classification
- [ ] T4.2: CHANGELOG surgical edits
- [ ] T4.3: CLAUDE.md edits
- [ ] T4.4: 4-locale README parity edits
- [ ] T4.5: Final 5-cmd verification batch
- [ ] T4.6: M4 commit

## Acceptance Tracker

| AC | Status | Verification | Linked REQ |
|----|--------|--------------|------------|
| AC-LCL-001 | pending | Backup dir + manifest 31 entries | REQ-LCL-001/002 |
| AC-LCL-002 | pending | PRESERVE SHA256 (10 sample) | REQ-LCL-004 |
| AC-LCL-003 | pending | Keyword count ≤5 | REQ-LCL-013 |
| AC-LCL-004 | pending | Hugo exit 0 | REQ-LCL-011 |
| AC-LCL-005 | pending | go test PASS delta = 0 | REQ-LCL-012 |
| AC-LCL-006 | pending | 4-locale symmetric count | REQ-LCL-009/010 |
| AC-LCL-007 | pending | CHANGELOG pre-v3.0 SHA256 | REQ-LCL-007 |
| AC-LCL-008 | pending | Manifest SHA256 self-check | REQ-LCL-002 |
| AC-LCL-009 | pending | 0 .go file modifications | §C exclusion #1 |
| AC-LCL-010 | pending | 0 template mirror modifications | §C exclusion #2 |
| AC-LCL-011 | pending | Locale file count unchanged | REQ-LCL-009 |

## Notes

### Plan-phase deviations from spawn prompt §C

- **File count**: Spawn prompt §C claimed 31 files (root MD 4 + skills 4 + rule 1 + docs-site 22). Actual grep verification produced **31 files but with different breakdown** (root MD 6 + skills 4 + rule 1 + docs-site 20). See spec.md §A.1 + §A.1.7 for full reconciliation.
- **docs-site distribution**: Spawn prompt §C claimed ja/zh missing `code-based-path.md` (asymmetric); actual grep shows `code-based-path.md` exists in all 4 locales but does NOT contain `agency` keyword. The agency-keyword distribution is symmetric (5 per locale × 4 = 20).
- **HEAD reference**: Spawn prompt §A.4 referenced `731aa0df5`; actual plan-phase entry HEAD is `87dd61564` (parallel session race, L9 reinforced).

### Partial verification disclosure (§A.3)

Per-file inspection of all 31 in-scope files was **NOT performed during plan-phase**. The 4 replacement categories framework is established; per-file semantic judgment is deferred to run-phase milestones M1-M4.

### Follow-up SPEC candidates (§A.6)

Documented in spec.md §A.6:
- SPEC-V3R6-LEGACY-CLEANUP-002 (template mirror cascade — 7 files)
- SPEC-V3R6-LEGACY-CLEANUP-003 (production Go code audit — 19 files)
- SPEC-V3R6-LEGACY-CLEANUP-004 (master design doc cleanup)
- SPEC-V3R6-LEGACY-CLEANUP-005 (historical SPEC archive consolidation — 38+ files)

## Lessons Captured (post-merge candidates)

- (TBD post-run-phase) Lesson candidates will be documented after milestone M4 completes.
