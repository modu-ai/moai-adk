# Tasks — SPEC-V3R4-SPECLINT-DEBT-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-15 | manager-spec | 초기 draft. plan.md 의 3 Wave / 11 task (T-SLD-001 ~ T-SLD-011) 일람표. 각 task 의 wave 배치, dependency, priority, 추정 effort 분류 (S/M/L), reference REQ/AC, MX 태그 표기. |

---

## 1. Task 일람표

| Task ID    | Wave | Category              | Priority | Effort | Dependencies        | REQ Reference | AC Reference  | MX Tags            |
|------------|------|-----------------------|----------|--------|---------------------|---------------|---------------|--------------------|
| T-SLD-001  | 1    | FrontmatterInvalid    | P0       | M      | -                   | REQ-SLD-001   | AC-SLD-001    | @MX:NOTE           |
| T-SLD-002  | 1    | MissingDependency     | P0       | S      | -                   | REQ-SLD-003   | AC-SLD-003    | -                  |
| T-SLD-003  | 1    | DependencyCycle       | P0       | S      | -                   | REQ-SLD-004   | AC-SLD-004    | -                  |
| T-SLD-004  | 1    | ModalityMalformed     | P0       | S      | -                   | REQ-SLD-005   | AC-SLD-005    | -                  |
| T-SLD-005  | 1    | MissingExclusions     | P0       | S      | T-SLD-001           | REQ-SLD-002   | AC-SLD-002    | -                  |
| T-SLD-006  | 1    | CoverageIncomplete    | P0       | L      | -                   | REQ-SLD-006   | AC-SLD-006    | @MX:WARN @MX:REASON |
| T-SLD-007  | 2    | StatusGitConsistency  | P1       | L      | Wave 1 완료          | REQ-SLD-007   | AC-SLD-007    | @MX:WARN           |
| T-SLD-008  | 2    | OrphanBCID            | P1       | S      | -                   | REQ-SLD-008   | AC-SLD-008    | -                  |
| T-SLD-009  | 3    | 통합 lint 검증         | P0       | S      | T-SLD-001~T-SLD-008 | REQ-SLD-009   | Gate G1/G2/G3 | -                  |
| T-SLD-010  | 3    | plan-auditor review   | P0       | S      | T-SLD-009           | REQ-SLD-010   | Gate G6       | -                  |
| T-SLD-011  | 3    | PR 생성 + CI GREEN     | P0       | M      | T-SLD-010           | REQ-SLD-009   | AC-SLD-009 / Gate G4 | -          |

**Effort 분류**:
- **S** (Small): 단일 SPEC 1-2 라인 편집, < 30분.
- **M** (Medium): 2-10 SPEC 편집 또는 자동화 스크립트 작성, 1-3시간.
- **L** (Large): 다수 SPEC 일괄 처리 또는 case-by-case 분석, 4시간+.

---

## 2. Wave 1 Task 상세

### T-SLD-001 (FrontmatterInvalid, P0, M)

- **Action**: SPEC-V3R2-RT-001 (7 필드) + SPEC-V3R4-HARNESS-002 (4 필드) frontmatter 보충.
- **Procedure**: 인접 SPEC 컨벤션 + git log + 본문 분석.
- **Output**: 2개 SPEC frontmatter 완전화.
- **Plan reference**: plan.md §2 T-SLD-001.
- **Verify**: `moai spec lint --strict 2>&1 | grep -c "FrontmatterInvalid"` → 0.

### T-SLD-002 (MissingDependency, P0, S)

- **Action**: SPEC-V3R2-RT-005 의 depends_on 에서 `SPEC-V3R2-SCH-001` 제거 + SPEC-V3R3-COV-001 의 depends_on 에서 `SPEC-V3R3-ARCH-003` 제거.
- **Plan reference**: plan.md §2 T-SLD-002.
- **Verify**: `moai spec lint --strict 2>&1 | grep -c "MissingDependency"` → 0.

### T-SLD-003 (DependencyCycle, P0, S)

- **Action**: SPEC-V3R2-RT-004 의 depends_on 에서 `SPEC-V3R2-RT-005` 제거 (백엣지 차단).
- **Plan reference**: plan.md §2 T-SLD-003.
- **Verify**: `moai spec lint --strict 2>&1 | grep -c "DependencyCycle"` → 0.

### T-SLD-004 (ModalityMalformed, P0, S)

- **Action**: SPEC-V3R2-SPC-003 line 95 REQ-SPC-003-041 본문에 `SHALL` 키워드 삽입.
- **Plan reference**: plan.md §2 T-SLD-004.
- **Verify**: `grep -A 2 "REQ-SPC-003-041" .moai/specs/SPEC-V3R2-SPC-003/spec.md | grep -c "SHALL"` ≥ 1.

### T-SLD-005 (MissingExclusions, P0, S)

- **Action**: SPEC-V3R4-HARNESS-002 의 `## Out of Scope` 섹션에 최소 1 항목 추가.
- **Plan reference**: plan.md §2 T-SLD-005.
- **Verify**: `moai spec lint --strict 2>&1 | grep -c "MissingExclusions"` → 0.

### T-SLD-006 (CoverageIncomplete, P0, L) @MX:WARN

- **Action**: 25+ 미참조 REQ 마다 결정 (a) AC 추가 / (b) REQ 제거 / (c) 기존 AC 에 reference 추가.
- **Plan reference**: plan.md §2 T-SLD-006 + §5.2 결정 트리.
- **Verify**: `moai spec lint --strict 2>&1 | grep -c "CoverageIncomplete"` → 0.
- **@MX:WARN**: 실제 케이스가 50+ 일 가능성. 60건 초과 시 sub-task 분할 (T-SLD-006a/b/c).
- **@MX:REASON**: CoverageIncomplete 가 가장 case-specific 분석을 요구하므로 single task 추정 effort 부정확.

---

## 3. Wave 2 Task 상세

### T-SLD-007 (StatusGitConsistency, P1, L) @MX:WARN

- **Action**:
  1. `scripts/spec-status-sync.go` 자동화 스크립트 작성.
  2. `--dry-run` 으로 변경 사항 검토.
  3. False-positive 수동 검토.
  4. `--apply` 로 batch update.
- **Plan reference**: plan.md §3 T-SLD-007 + §5.3 status 추론 로직.
- **Verify**: `moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"` → ≤ 5.
- **@MX:WARN**: false-positive rate > 30% 시 수동 batch 전환.

### T-SLD-008 (OrphanBCID, P1, S)

- **Action**: SPEC-V3R3-ARCH-007 frontmatter 에서 `bc_id` 필드 제거 또는 `bc_id: []`.
- **Plan reference**: plan.md §3 T-SLD-008.
- **Verify**: `moai spec lint --strict 2>&1 | grep -c "OrphanBCID"` → 0.

---

## 4. Wave 3 Task 상세

### T-SLD-009 (통합 lint 검증, P0, S)

- **Action**: `moai spec lint --strict` 재실행 + 카테고리별 카운트 측정 + 보고서 작성 (`lint-final.md`).
- **Plan reference**: plan.md §4 T-SLD-009.
- **Verify**: `moai spec lint --strict; echo $?` → 0.

### T-SLD-010 (plan-auditor self-review, P0, S)

- **Action**: orchestrator 가 plan-auditor 를 호출하여 본 SPEC 의 3종 문서 audit.
- **Plan reference**: plan.md §4 T-SLD-010.
- **Verify**: plan-auditor 점수 ≥ 0.85 또는 PASS.

### T-SLD-011 (Run PR 생성 + CI GREEN + 머지, P0, M)

- **Action**:
  1. `feat/SPEC-V3R4-SPECLINT-DEBT-001` 브랜치에 변경 commit.
  2. `gh pr create --base main --title ... --body ...`.
  3. GitHub Actions `spec-lint` workflow 대기.
  4. `gh pr checks` 로 GREEN 확인.
  5. admin squash merge.
- **Plan reference**: plan.md §4 T-SLD-011.
- **Verify**: PR 머지 완료 + main HEAD `moai spec lint --strict` exit 0.

---

## 5. Cross-cutting Notes

### 5.1 Commit 분할 전략 (T-SLD-011 reference)

운영 결정 (plan.md OQ4) 후 적용:

- **Option A (단일 commit)**: 모든 변경을 단일 commit. reviewer 부담 크나 SPEC 의 의미 (메타데이터 batch 정정) 가 명확.
- **Option B (카테고리별 5 commit)**:
  - commit 1: T-SLD-001 (frontmatter 보충).
  - commit 2: T-SLD-002 + T-SLD-003 (depends_on 정정 + cycle 차단).
  - commit 3: T-SLD-004 + T-SLD-005 (modality + exclusions).
  - commit 4: T-SLD-006 (coverage 일괄).
  - commit 5: T-SLD-007 + T-SLD-008 (status + bc_id).

추천: **Option B** (categorical commits) — diff 검토 시 카테고리별 분리되어 reviewer 부담 완화.

### 5.2 Parallel 실행 가능성

다음 task pair 들은 dependency 가 없으므로 worktree 또는 단일 세션 내 병렬 실행 가능:

- T-SLD-001, T-SLD-002, T-SLD-003, T-SLD-004 — 서로 독립 (4 task 병렬).
- T-SLD-006, T-SLD-008 — 서로 독립.

다음은 sequential 만 가능:

- T-SLD-005 ← T-SLD-001 (frontmatter 정정 후 본문 검토 수월).
- T-SLD-007 ← Wave 1 전체 완료 (ERROR 잔존 시 자동화 동작 미정).
- T-SLD-009 ← T-SLD-001 ~ T-SLD-008.
- T-SLD-010 ← T-SLD-009.
- T-SLD-011 ← T-SLD-010.

### 5.3 Worktree 전략

본 SPEC 의 run-phase 는 worktree `~/.moai/worktrees/moai-adk-go/SPEC-V3R4-SPECLINT-DEBT-001` 에서 실행. plan-phase 는 main checkout (plan-in-main doctrine PR #822 준수).

---

## 6. References

- spec.md §3 (REQ-SLD-001 ~ REQ-SLD-010)
- acceptance.md (AC-SLD-001 ~ AC-SLD-010, Gate G1 ~ G6)
- plan.md (Wave 1 / Wave 2 / Wave 3)
- research.md §5 (자동화 vs 수동 분류)
- `.claude/rules/moai/workflow/mx-tag-protocol.md` (MX 태그 규칙)
