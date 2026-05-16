# SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 Progress

## Run-Phase Summary

**Branch**: `feat/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002`
**Started**: 2026-05-16
**Wave 3 Verification**: PASS — `0 error(s), 0 warning(s)`

## Wave 1: Analysis (Complete)

Category A (5건, forward drift — frontmatter sync-up):
- A1 SPEC-GLM-MCP-001: in-progress → completed
- A2 SPEC-STATUSLINE-001: in-progress → implemented
- A3 SPEC-V3R2-WF-002: in-progress → implemented
- A4 SPEC-V3R4-CATALOG-001: implemented → completed
- A5 SPEC-WORKTREE-002: implemented → completed

Category B (12건, suspect downgrade — per-SPEC analysis):
- B1 SPEC-V3R2-ORC-003: lint.skip (bulk-closure PR #926 override)
- B2 SPEC-V3R2-RT-001: lint.skip (chore sweep #930 override)
- B3 SPEC-V3R2-RT-007: sync(spec) commit + lint.skip (walker main-only)
- B4 SPEC-V3R2-SPC-002: lint.skip (bulk-closure PR #926 override)
- B5 SPEC-V3R2-SPC-003: lint.skip (backfill commit override)
- B6 SPEC-V3R2-WF-003: sync(spec) commit + lint.skip (walker main-only)
- B7 SPEC-V3R3-CI-AUTONOMY-001: lint.skip (bulk-closure PR #927 override)
- B8 SPEC-V3R3-HARNESS-LEARNING-001: lint.skip (feat commit override)
- B9 SPEC-V3R3-PROJECT-HARNESS-001: sync(spec) commit + lint.skip (walker main-only)
- B10 SPEC-V3R3-RETIRED-AGENT-001: lint.skip (PR #856 docs(sync) override)
- B11 SPEC-V3R3-RETIRED-DDD-001: lint.skip (HARNESS-001 closeout override)
- B12 SPEC-V3R4-LINT-SKIP-CLEANUP-001: lint.skip (chore(spec) PR #937 override)

## Wave 2: Apply (Complete)

Commits on branch:
- `c4993f252` chore(spec): Wave 2-A — Category A 5건 frontmatter sync-up
- `dc314ff1b` sync(spec): SPEC-V3R2-RT-007 — B3 sync commit
- `d4d4eabf6` sync(spec): SPEC-V3R2-WF-003 — B6 sync commit
- `f35a2c984` sync(spec): SPEC-V3R3-PROJECT-HARNESS-001 — B9 sync commit
- `86b9a2070` chore(spec): Wave 2-B-skip — Category B 9건 lint.skip 추가
- `a44bcc7a2` chore(spec): lint.skip 포맷 수정 — flat → nested YAML
- `af5ea0c39` chore(spec): B3/B6/B9 lint.skip 추가 (walker main-only 스캔 제약)
- `a567bbe74` chore(spec): FOLLOWUP-002 자체 status in-progress + lint.skip

## Wave 3: Verify (Complete)

```
moai spec lint --strict 2>&1 | tail -1
✓ No findings — all SPEC documents are valid
```

Result: 0 error(s), 0 warning(s) — AC-SDF002-X-001 PASS

## Key Discoveries

1. **lint.skip nested YAML format**: `lint.skip: [...]` is invalid YAML. Must use `lint:\n  skip:\n    - ...`
2. **walker main-only scan**: `getGitImpliedStatus` calls `git log main --grep=...`. Feature branch commits are invisible until PR merge. B3/B6/B9 sync(spec) commits only take effect post-merge.
3. **FOLLOWUP-002 plan commit body**: Plan commit `7e9da02c5` body lists all B SPEC IDs, causing walker to find it via `--grep`. But `commitMatchesSPECID` filters by title-only word-boundary, so plan commit is correctly filtered for individual B SPECs.
4. **18th warning**: FOLLOWUP-002 itself (draft vs planned) = 18th warning, not 17th. Resolved by status→in-progress + lint.skip.

## Acceptance Criteria Status

| AC ID | Description | Status |
|-------|-------------|--------|
| AC-SDF002-A-001 | A1-A5 frontmatter status sync-up | PASS |
| AC-SDF002-B-001 | B1-B12 remediation (lint.skip / sync-commit) | PASS |
| AC-SDF002-X-001 | `moai spec lint --strict` 0 error 0 warning | PASS |

## Next Phase

Sync phase: Update FOLLOWUP-002 status to `completed`, remove lint.skip from FOLLOWUP-002, generate CHANGELOG entry, open sync PR.
