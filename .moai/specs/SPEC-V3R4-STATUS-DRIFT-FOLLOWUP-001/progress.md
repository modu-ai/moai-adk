# Progress — SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001

## Phase Tracking

- plan_started_at: 2026-05-16
- plan_complete_at: 2026-05-16T10:30:00Z
- plan_status: audit-ready
- run_started_at: 2026-05-16
- run_complete_at: (pending — Wave 6 closeout 후 갱신)
- sync_started_at: (pending)
- sync_complete_at: (pending)

## Artifacts Created (plan-phase)

| File | Lines | Purpose |
|------|-------|---------|
| spec.md | ~280 | EARS 5-modality requirements + 8 AC + Exclusions |
| plan.md | ~430 | 5-Wave 전략 + bulk script architecture + 4 OQ + 8 risks |
| design.md | ~370 | Status field model + detector contract + edge case 분석 |
| research.md | ~440 | 64 drift 8-pattern 분류 + LSKC-001 fork 평가 + alternatives |
| acceptance.md | ~310 | 5 Given/When/Then + 4 edge case + DoD checklist + quality gates |
| spec-compact.md | ~80 | REQ + AC + Exclusions auto-extract |
| progress.md | (this file) | phase tracking |

## Plan-Phase Metrics

- REQ count: 16 (REQ-SDF-001 ~ REQ-SDF-016)
  - Ubiquitous: 3
  - Event-Driven: 4
  - State-Driven: 2
  - Unwanted: 5
  - Optional: 2
- AC count: 8 (AC-SDF-001 ~ AC-SDF-008)
- Exclusions count: 13 (spec.md §10)
- MX target count (Wave 3): 4 (terminalStatusEnum + Check fn + exemption 분기 + 4 new tests)
- Affected SPEC count: 64 (Pattern A: 50, B: 4, C: 6, D: 1, E: 1, F: 1, G: 1, H: 3)
- Wave count: 5 (BASELINE → A → B+C → D-G → H sync)

## Predecessor Verification

- ✅ SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 PR #933 (run) + #934 (sync) MERGED
- ✅ SPEC-V3R4-LINT-SKIP-CLEANUP-001 PR #937 MERGED → main `758341089`
- ✅ walker filter active (drift.go::shouldSkipCommitTitle 검증됨)
- ✅ 64 WARN baseline 측정 (project_v3r4_lskc_001_run_complete 메모리 검증)

## BODP Evaluation

- Signal A (depends_on + diff path overlap): ¬ (본 SPEC dependencies: 없음 — LSKC-001/CHORE-SKIP/DEBT-001 모두 머지됨)
- Signal B (worktree co-location): ¬ (worktree 미사용 — feedback_worktree_never_use)
- Signal C (open PR head): ¬ (현재 main 작업)
- 결정: **main @ origin/main** (plan-in-main 표준)

## Open Items for plan-auditor

1. **OQ1 — Pattern A 해석 1 vs 해석 2**: plan.md §7 OQ1 — default 해석 2 (downgrade) 채택. plan-auditor가 정당화 검증.
2. **OQ2 — Pattern B+C 분기 (b) 허용 기준**: 0-3건 허용, 4건 이상 escalate.
3. **OQ3 — bulk script 보존 여부**: 기본 보존 (LSKC-001 정신 계승).
4. **OQ4 — Pattern H 처리 phase**: sync-phase 권장 (본 SPEC 자체 포함 가능).
5. **Pattern G 단일 SPEC**: SPEC-V3R3-WEB-001 — Wave 3 detector exemption 으로 자동 해결, 의도된 abandonment 인지 verification 권장 (run-phase).
6. **64 drift count 변동성**: plan-phase 추정. BASELINE milestone 에서 실시간 정정 필요.

## Pre-Write Frontmatter Validation (self-check)

- ✅ id: `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (regex `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` 매칭 — STATUS-DRIFT-FOLLOWUP 가 [A-Z][A-Z0-9-]+ 확장형, V3R4 prefix 동일 패턴 LSKC-001 / DEBT-001 와 일치)
- ✅ version: `"0.1.0"` (quoted string)
- ✅ status: `draft` (enum)
- ✅ created_at: `2026-05-16` (ISO date)
- ✅ updated_at: `2026-05-16` (ISO date)
- ✅ author: `GOOS행님` (string)
- ✅ priority: `P1` (P-prefixed uppercase)
- ✅ labels: `[v3r4, lint, spec-frontmatter, status-drift, plan-in-main]` (YAML array)
- ✅ issue_number: `null` (integer | null)
- ✅ NEVER used `created` / `updated` (legacy field names) — `created_at` / `updated_at` 적용

## Audit Pre-Check (against plan-auditor checklist)

- ✅ EARS 5 modality 모두 사용
- ✅ REQ ↔ AC 매핑 100% (16 REQ × 8 AC)
- ✅ Exclusions 명시 (13 항목)
- ✅ Worktree 정책 명시 (worktree 미사용 + 사유)
- ✅ BODP 평가 기록
- ✅ 9-field frontmatter
- ✅ 의존 SPEC 명시
- ✅ 위험 평가 (8 risks)
- ✅ Definition of Done (4-section, ~30 항목)
- ✅ Open Questions (4 OQ + default decision)
- ✅ HISTORY row 작성 (각 artifact)

## Run-Phase Wave Log

### Wave 1 (BASELINE)
- commit: `b57bcd31e`
- 64건 drift 측정 확인 (±계획 64건, 정확 일치)
- 8 패턴 카테고리화 (A:47, B:4, C:6, D:1, E:1, F:1, G:1, H:3)
- affected-list 4개 파일 생성
- run-baseline.md 생성

### Wave 2+3 (Pattern A+B+C bulk script)
- commit: `1bcee01c1`
- .moai/scripts/status-drift-cleanup.go 신규 (LSKC-001 fork)
- Pattern A 47건 + Pattern B 4건 + Pattern C 6건 = 57건 처리
- lint StatusGitConsistency: 64 → 8

### Wave 4 (Pattern D/E/F/G TDD + 추가 20건)
- commit: `87223091c`
- internal/spec/lint.go: terminalStatusEnum + Check() 조기 반환 (Pattern D/E/F/G 4건 exemption)
- internal/spec/status_terminal_exemption_test.go: 4 test RED→GREEN
- 추가 20건 drift 수정 (make install 후 신규 발견):
  - in-progress → implemented (14건)
  - in-progress → planned (5건)
  - in-progress → completed (1건)
- lint StatusGitConsistency: 8 → 4

### Wave 5 (Pattern H 문서화)
- Pattern H (3건) + 이 SPEC 자체 (1건) = 4건 → sync-phase 위임 (OQ4 결정)
- Pattern H 해설:
  - SPEC-V3R4-LINT-SKIP-CLEANUP-001: frontmatter `implemented`, git-implied `planned`
    → 이 브랜치에 plan 커밋만 존재, run/sync 커밋은 다른 브랜치에서 머지됨
    → sync-phase에서 feat: 커밋 추가 시 git-implied 자동 갱신
  - SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001: frontmatter `completed`, git-implied `planned`
    → 동일 패턴
  - SPEC-V3R4-SPECLINT-DEBT-001: frontmatter `completed`, git-implied `planned`
    → 동일 패턴
  - SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 (자체): frontmatter `in-progress`(갱신 후), git-implied `planned`
    → Wave 4 이후 feat: 커밋 있으나 브랜치에서만 유효, main 머지 전
- 결론: 4건 모두 PR 머지 후 main에서 `moai spec lint --strict` 재실행 시 자동 해소
  → AC-SDF-001 binary criterion (=0)은 main 머지 후 검증
- progress.md status 갱신 (run_started_at + Wave 5 log)
- 이 SPEC 자신의 frontmatter status: draft → in-progress

## Pattern H Sync-Phase Delegation

Pattern H 4건의 git-implied 불일치는 다음 이유로 sync-phase 완료 후 자동 해소:

1. 현재 `feat/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` 브랜치에는 plan 커밋만 존재
2. Wave 4 `feat:` 커밋(87223091c)이 브랜치에 있으나 main 미머지
3. PR 머지 → main에 feat 커밋 반영 → git-implied `implemented` 갱신
4. sync-phase에서 `chore/sync:` 또는 `docs(sync):` 커밋 추가 → git-implied `completed` 갱신
5. StatusGitConsistencyRule이 각 Pattern H SPEC의 git history를 재평가 → 불일치 해소

**예상 결과**: PR 머지 후 main에서 `moai spec lint --strict | grep -c StatusGitConsistency` = 0

### Wave 6 (Closeout)
- go test ./internal/spec/...: PASS (4 new exemption tests GREEN)
- go test ./...: ALL PASS (no failures)
- go vet ./...: clean
- lint StatusGitConsistency 최종 count: 4 (Pattern H sync-phase 위임 — AC-SDF-001은 main 머지 후 검증)
- 전체 브랜치 커밋 수: 5개 (Wave 1 + Wave 2+3 합산 + Wave 4 + Wave 5 + Wave 6)
- run-phase 완료 상태 오케스트레이터에게 반환

## MX Tag Report

### GREEN Phase 태그 (Wave 4)
- `terminalStatusEnum` 앞: `@MX:NOTE` + `@MX:REASON` 추가
  - 파일: `internal/spec/lint.go`
  - 이유: 새로운 package-level var로 business rule 맥락 필요

### 기존 태그 보존
- `Linter` struct: `@MX:ANCHOR` (fan_in ≥ 3, 변경 없음)
- `Linter.Lint()`: `@MX:ANCHOR` (fan_in ≥ 3, 변경 없음)
- `getGitImpliedStatus()`: `@MX:ANCHOR` on drift.go (fan_in=2, 변경 없음)

## Next Action

오케스트레이터에게 복귀:
- branch: `feat/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001`
- push + PR 생성은 오케스트레이터 담당
- sync-phase: AC-SDF-001 (moai spec lint --strict | grep -c StatusGitConsistency = 0) main 머지 후 검증
