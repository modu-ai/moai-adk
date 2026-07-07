---
id: SPEC-INTERNAL-ARCH-001
status: draft
updated: 2026-07-08
---

# SPEC-INTERNAL-ARCH-001 — Implementation Plan

## §A Context

internal/ 전수 감사(2026-07-08)의 검증 findings 6건을 행위 보존 구조 리팩터링으로 해소한다. 본 SPEC은 **카탈로그 내 최고 회귀 위험** SPEC이다: 대상이 CLI 진입 경로(update/hook), 전역 의존성 배선(deps singleton), config 해석(전 커맨드가 소비)에 걸쳐 있고, 변경 자체는 "아무 행위도 바꾸지 않아야" 성공이다.

- 요구사항 SSOT: spec.md §B (REQ-ARCH-001..008)
- AC SSOT: acceptance.md
- 설계 SSOT: design.md
- 증거 SSOT: research.md

## §B Known Issues (착수 전 인지 사항)

1. **baseline RED**: 2026-07-08 실측 기준 `go test ./...`에 2건 FAIL 잔존 (`internal/spec` TestCloseSubjectDoctrineAmendment/AC-DLC-011, `internal/statusline` TestBuild_WritesContextUsageWithSessionID env flaky). sibling SPEC-INTERNAL-TEST-001이 복구 담당 — **복구 전 M1 착수 금지** (REQ-ARCH-008).
2. **SPEC-INTERNAL-TEST-001 미저작**: depends_on은 forward reference 상태. run-phase 진입 전 해당 SPEC의 completed 여부를 기계 확인해야 한다.
3. **병렬 세션 레이스 이력**: 본 리포는 공유 체크아웃 동시 커밋 레이스가 반복 실측된 환경이다. `internal/cli`(update.go 포함)는 과거 병렬 세션 접촉 이력이 있는 영역 — spawn 전 겹침 확인 필수.
4. **loader 카운트 delta**: 감사는 "~17 near-identical section loaders"라 했으나 실측은 13개(`load*Section`, loader.go L115-306). 감사가 언급한 `loadDesignSection`은 현존하지 않는다(추정: config-diet 계열 후속 작업으로 축소). REQ-ARCH-005의 목표(테이블 주도 붕괴)는 delta와 무관하게 성립 — research.md §B에 delta 기록.
5. **`flattenStruct` 소재 정정**: 감사 anchor는 loader.go:562였으나 실제는 `resolver.go:562` (loader.go는 474줄). resolver 은퇴(M4) 시 doctor 표시용 flatten 로직의 존치/이전 결정 필요 — design.md §D.

## §C Pre-flight (run-phase 진입 전 체크리스트)

- [ ] SPEC-INTERNAL-TEST-001 `status: completed` 확인 (`grep "^status:" .moai/specs/SPEC-INTERNAL-TEST-001/spec.md`)
- [ ] `go test ./...` exit 0 실측 (AC-ARCH-001 gate — 실행 출력 verbatim 기록)
- [ ] pre-spawn sync check: `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` → `0 0` 또는 `0 N`만 허용
- [ ] `git diff --cached --stat` + `git status --short`로 병렬 세션의 `internal/cli`, `internal/core`, `internal/config` 접촉 여부 확인 — 겹침 시 quiesce 대기
- [ ] 격리 worktree 사용 권고 (과거 병렬 run-phase 레이스 흡수 선례) — 단, worktree가 병렬 main 전진을 가릴 수 있으므로 landing 직전 rebase 재확인
- [ ] baseline 산출물 캡처: `moai --help` 및 주요 서브커맨드 help 출력 파일 저장, `wc -l` 기준선(update.go 3172 / hook.go 1182 / loader.go 474 / manager.go 438 / resolver.go 1156) 기록

## §D Constraints

- **cycle_type=ddd 명시 위임**: 전역 quality.yaml pin은 `tdd`이나, 본 SPEC은 기존 코드 행위 보존이 목적이므로 orchestrator는 "Use the manager-develop subagent ... (cycle_type=ddd)"로 명시 위임한다.
- **RED commit 금지**: 모든 commit은 green-to-green (REQ-ARCH-007). characterization test 추가 commit → 구조 변경 commit 순서.
- **milestone당 커밋 분리**: findings 간 커밋 혼합 금지 — 각 milestone은 독립 revert 가능해야 한다.
- **specific-path commit**: 병렬 세션 방어를 위해 `git add -A` 금지, 대상 경로 명시 staging만 허용.
- **시간 추정 금지**: 우선순위 라벨과 순서만 사용.

## §E Self-Verification (run-phase 종료 시 manager-develop 의무)

- E1: acceptance.md 전체 AC PASS/FAIL matrix (커맨드 + verbatim 출력)
- E2: `go build ./...` (host) + `GOOS=windows go build ./...` cross-build exit 0
- E3: 변경 패키지 커버리지 — 리팩터링 전 대비 하락 없음 (`go test -cover` 실측 인용)
- E4: `moai --help` byte-diff — baseline 캡처 대비 0 diff
- E5: `golangci-lint run` — NEW finding 0
- E6: milestone별 커밋 SHA 목록 + green-to-green 증거 (각 커밋에서 suite green)
- E7: 미검증 항목(Gaps) 명시 열거 — verification-claim-integrity §3.4

## §F Milestones (우선순위 기반 — 시간 추정 없음)

| MS | 대상 (REQ) | Priority | 종속성 | 착지 단위 |
|----|-----------|----------|--------|-----------|
| **M0** | green baseline gate (REQ-ARCH-008) | Critical | SPEC-INTERNAL-TEST-001 completed | gate만 — 커밋 없음 |
| **M1** | DI seam (REQ-ARCH-001) | High | M0 | seam 패키지 신설 → pilot 1개 subpackage 이행 → `var deps` 전역 제거 |
| **CHECKPOINT-1** | M1 사후 재평가 | Critical | M1 | 아래 CHECKPOINT 절 참조 |
| **M2** | update.go / hook.go 파일 분할 (REQ-ARCH-002) | Medium | M1 + CHECKPOINT-1 PASS | characterization 선행 → update.go 분할(3-4 커밋) → hook.go 분할(2 커밋) |
| **M3** | internal/core 해체 (REQ-ARCH-003) | High | M0 (M1과 독립) | 스텁 제거 1커밋 → subpackage별 이동 3커밋 (각 커밋에 call-site 갱신 포함) |
| **M4** | config 단일 pipeline (REQ-ARCH-004) | Medium | M0 (독립) | doctor/runtime 정합 characterization → doctor_config.go를 ConfigManager 기반 재구축 → resolver.go 제거 |
| **M5** | loader table-driven (REQ-ARCH-005) | Medium | M4 이후 권장 (동일 패키지 churn 회피; 논리적 종속 없음) | sectionSpec 레지스트리 → load 붕괴 → save/get/set 붕괴 |
| **M6** | env-var 문서 정정 (REQ-ARCH-006) | Low | 없음 — 어느 시점이든 선행 착지 가능 | 문서 1커밋 |

**권장 실행 순서**: M0 → M6(선착지 가능) → M1 → CHECKPOINT-1 → {M3 ∥ M4→M5} → M2. M2를 마지막에 두는 이유: 최대 diff(3,172줄 재배치)라 병렬 세션 겹침 창을 최소화하고, M1 seam의 안정화를 먼저 관찰한다.

### CHECKPOINT-1 (SPEC-CLI-SUBPKG-SPLIT-001 REQ-CSS-010 선례 승계)

M1 완료 후, M2 착수 전 orchestrator가 재평가한다:

- seam 도입이 실제로 update/hook 관심사 분리를 저해하는 잔여 결합을 남겼는가?
- M1 과정에서 발견된 새 coupling axis가 있는가?
- 판정: **PROCEED** (M2 진행) / **RESCOPE** (M2 범위 축소 후 진행) / **STOP** (M1+M3..M6만 ship, M2는 별도 SPEC 이관 — SUBPKG-SPLIT M1-only close 선례와 동일 패턴)

STOP 판정은 실패가 아니다 — 선례상 정당한 종결 경로이며, 이 경우 spec.md의 REQ-ARCH-002를 `partially_superseded_by` 후속 SPEC으로 이관 처리한다.

## §G Anti-Patterns (금지)

- 감사 findings를 재검증 없이 결함으로 단정하고 코드를 수정 (defect-claim은 도구 검증 선행 — 본 SPEC은 저작 시점 재검증 완료, run-phase에서 anchor 재확인 의무)
- characterization 없이 update.go merge/path-guard 로직 이동 (SEC-HARDEN 가드 가족은 보안 회귀 표면)
- 파일 분할과 동작 수정을 한 커밋에 혼합
- `internal/core` 이동 시 3개 subpackage를 한 커밋에 일괄 이동 (revert 단위 파괴)
- resolver.go 제거를 doctor_config.go 재구축 **이전에** 수행 (중간 상태에서 doctor 기능 상실 = 행위 변경)
- 임의 시점 `git add -A` / 광역 커밋

## §H Cross-References + 감사 권고

- **plan-auditor 독립 감사 필수 권고**: Tier L + 최고 회귀 위험 — Phase 0.5 plan audit gate를 skip하지 않는다.
- **sync-auditor 구현 후 감사 필수 권고**: 4-dimension scoring에서 특히 Functionality(행위 보존 증거)와 Consistency(green-to-green 이력)를 중점 검증.
- 선례: SPEC-CLI-SUBPKG-SPLIT-001 (M1 run commit 0ee246ad9, re-sequence c9abe2ddb, close 85adab898) — agentlint 추출 시 `moai --help` byte-identical 검증 기법 재사용.
- design.md §F — 이행 순서 상세 + rollback 전략.
