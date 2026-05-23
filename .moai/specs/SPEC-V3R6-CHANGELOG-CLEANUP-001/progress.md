# SPEC-V3R6-CHANGELOG-CLEANUP-001 — Progress

## Phase 0 — Plan-Phase Header

- SPEC ID: SPEC-V3R6-CHANGELOG-CLEANUP-001
- Title: CHANGELOG.md `[Unreleased]` 환각·중복 엔트리 정리
- Tier: S (Simple — 2 files, ~25 LOC net)
- Status: draft
- Plan-phase created: 2026-05-23
- Plan-phase author: MoAI Maintainer (manager-spec direct delegation)
- plan-auditor verdict: pending (orchestrator-side plan-auditor invocation 후 채워질 예정)
- Expected line 34-39 sha256 (Phase 0 / M0 pre-flight, run-phase 시작 시 기록):

```
expected_line_34_39_sha256: 044c70f72f577a5cc5ec0a9df45b948bca134d131f2efb2deca3d84b397a77f7
```

(Phase 0 pre-flight 는 위 fenced block 내 라인을 실제 64-hex sha256 으로 교체. AC-CHL-004 의 `grep -E '^expected_line_34_39_sha256:' progress.md | awk '{print $2}'` 가 그 라인의 두 번째 토큰을 추출하여 검증.)

## Phase 0 — Pre-flight Findings (orchestrator-side verified)

본 SPEC 의 plan-phase 는 orchestrator 가 working-tree HEAD `97a36b5a2` (main) 에 대해 다음 사실을 사전 검증한 결과를 받았다. plan-phase 는 어떤 코드도 재조사하지 않고 이 사실을 ground truth 로 사용했다.

### Hallucination catalogue (line 65, 7-row drift)

§A.1 spec.md 의 7-row 표 참조. 모든 환각 사실은 working-tree 검증 완료.

### Sibling SPEC AC count cross-check (acceptance.md SSOT)

| SPEC | CHANGELOG line | Claimed | acceptance.md SSOT | Drift? |
|------|----------------|---------|---------------------|--------|
| HOOK-OBSERVE-OPT-IN-001 | 57 | 7/7 | 7 | NO |
| HOOK-ASYNC-EXPAND-001 | 59 | 12/12 | **8** | **YES (-4)** |
| HOOK-CWD-LEAK-AUDIT-001 | 61 | 8/8 | **7** | **YES (-1)** |
| CI-BASELINE-DRIFT-001 | 63 | 8/8 | 8 | NO |
| SESSION-HANDOFF-AUTO-001 | 65 | 10/10 | 10 | (deletion target) |
| SESSION-HANDOFF-AUTO-001 | 34 | 10/10 | 10 | NO (PRESERVE) |

## Milestone Tracking

| Milestone | Status | Started | Completed | Commit | Notes |
|-----------|--------|---------|-----------|--------|-------|
| M0 — Pre-flight baseline capture | pending | _N/A_ | _N/A_ | _N/A_ | Run-phase 시작 시 baseline `sed -n '34,39p' CHANGELOG.md \| sha256sum` 캡처, expected_line_34_39_sha256 필드 본 progress.md 에 기록 |
| M1 — Line 65 hallucination deletion | pending | _N/A_ | _N/A_ | _N/A_ | CHANGELOG.md 라인 65-anchored entry 단일 contiguous range deletion |
| M2 — Sibling AC count reconciliation | pending | _N/A_ | _N/A_ | _N/A_ | CHANGELOG.md 라인 59 (12/12 → 8/8) + 라인 61 (8/8 → 7/7) atomic edit |
| M3 — B12 standing-rule guard insertion | pending | _N/A_ | _N/A_ | _N/A_ | manager-develop-prompt-template.md Section B 끝에 B12 bullet 삽입 |

## AC Verification Matrix

| AC | Description | Verification Command | Status | Actual |
|----|-------------|----------------------|--------|--------|
| AC-CHL-001 | Line 65 hallucination 7-row 카탈로그 모두 제거 | `grep -c 'internal/handoff/{package,atomic_write,parser}' CHANGELOG.md` → 0 (+4 sub-conditions) | pending | _N/A_ |
| AC-CHL-002 | Sibling AC 카운트 acceptance.md SSOT 일치 | `grep -c 'AC 12/12 PASS' CHANGELOG.md` → 0 + 4 sibling 매치 | pending | _N/A_ |
| AC-CHL-003 | manager-develop-prompt-template B12 가드 삽입 | `grep -c 'B12. Sync-phase CHANGELOG emission'` → ≥1 (+ 3 sub-conditions) | pending | _N/A_ |
| AC-CHL-004 | Line 34-39 byte-identical 보존 (sha256) | `sha256sum <(sed -n '34,39p' CHANGELOG.md)` matches baseline | pending | _N/A_ |
| AC-CHL-005 | Diff scope 2 files only, 5 sibling 0 changes | `git diff --name-only HEAD~3..HEAD` matches expected list + sibling 0 changes | pending | _N/A_ |

## Sync-phase Evidence

_(placeholder — sync-phase 진입 시 manager-docs 가 CHANGELOG.md `[Unreleased]` 에 본 SPEC 엔트리 추가 + 본 progress.md 에 sync commit SHA + evidence summary 기록)_

## Verification Batch Results (orchestrator-side, post-implementation)

_(placeholder — run-phase 완료 후 orchestrator 가 §D Quality Gate 5개 명령 parallel batch 실행 결과 기록)_
