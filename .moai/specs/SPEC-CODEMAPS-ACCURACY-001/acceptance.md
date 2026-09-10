# acceptance — SPEC-CODEMAPS-ACCURACY-001

## §D AC Matrix

> 판정 명령은 실행 가능한 형태로 기재한다. 행번호 좌표는 develop `65196a5a7` 기준 작성 시점 값이며 run-phase 판정은 **post-absorb 재측정 좌표**(M0, progress.md §E.2 기록분)로 대체 판정한다 (REQ-CMA-008).

| AC | 심각도 | 요구 귀속 | 판정 기준 (이항 판정 가능 형태) |
|----|--------|-----------|--------------------------------|
| AC-CMA-001 | MUST | REQ-CMA-001 | 인용 경로 부재 집합이 열거된 부정 인용과 정확히 일치 |
| AC-CMA-002 | MUST | REQ-CMA-003/004 | modules.md가 `internal/kanban`을 서술하고 `internal/factory`·`internal/bodp` 양성 절 부재 |
| AC-CMA-003 | MUST | REQ-CMA-005 | data-flow.md에 `ListActive` 토큰 0건, 실제 API 서명 존재 |
| AC-CMA-004 | MUST | REQ-CMA-006 | known-5 경고 노트 5개 잔존 |
| AC-CMA-005 | MUST | REQ-CMA-002 | citations 검사 red/green/면제 3방향 성립 (뮤턴트 포함) |
| AC-CMA-006 | MUST | REQ-CMA-007 | 스킬 양본에 재발 방지 지시 삽입 (Template-First) |
| AC-CMA-007 | SHOULD | REQ-CMA-008 | post-absorb 재측정 기록 존재 |
| AC-CMA-008 | MUST | REQ-CMA-009 | F1 정정 기록 존재 + t432 트리 무결 |
| AC-CMA-009 | MUST | REQ-CMA-010 | 완료 판정 근거가 관측이지 재스탬프가 아님 |

### AC-CMA-001 — 부재 집합 불변식
- **Given** 수정 완료 후의 post-absorb 트리, **When** codemaps 6개 문서에서 `(internal|pkg|cmd)/[A-Za-z0-9_/.-]+` 토큰을 추출해 정리 규칙(t432: 후행 슬래시 제거·`.go` 복원·`cmd/moai/main`→`cmd/moai/main.go`)을 적용하고 존재 검사를 실행하면, **Then** 부재 집합은 blockquote 행에서 인용된 경로의 집합(§1.1 P1–P5 + P7 bodp 노트, 형식 정렬 후)과 정확히 일치한다 — 그 외 부재 0건.
- **Given** 위와 동일, **When** 부재 집합의 각 원소가 속한 행이 blockquote(`>` 접두)인지 검사하면, **Then** 전원 blockquote 행이다 (양성 부재 0).

### AC-CMA-002 — factory→kanban 재작성 + bodp 부정 노트
- **Given** modules.md, **When** `grep -n '### internal/factory' modules.md`를 실행하면 **Then** 0행이고, `grep -n '### internal/kanban' modules.md`는 1행 이상이다.
- **Given** `### internal/kanban` 절, **When** 절이 인용하는 모든 파일 경로(`internal/kanban/record.go`·`revision.go`·`integration_lock.go`·진입점 `internal/cli/factory.go` 등)를 존재 검사하면 **Then** 전부 실존한다.
- **Given** modules.md·dependencies.md, **When** `grep -n '^### internal/bodp' modules.md`와 양성(비-blockquote) `internal/bodp` 서술을 검색하면 **Then** 0건이고, blockquote 형식의 bodp 제거 기록 노트는 1건 존재한다 (D3).

### AC-CMA-003 — ListActive API 정정
- **Given** data-flow.md, **When** `grep -n 'ListActive' data-flow.md`를 실행하면 **Then** 0행이다 (mermaid 노드 197행·흐름 단계 214행·인터페이스 블록 356행 전부 수정됨 — post-absorb 재측정 좌표 기준).
- **Given** "Registry (Session)" 인터페이스 블록, **When** 블록 내용을 `internal/session/registry.go`의 실제 서명(`Register(sessionID, specID, phase string) error`, `Heartbeat(sessionID string) error`, `Deregister(sessionID string) error`, `Query(optSpecID string) ([]Entry, error)` — 리시버 메서드) 및 패키지 함수 `QueryActiveWork(optSpecID string) ([]Entry, error)`와 대조하면, **Then** 식별자·파라미터·반환 타입이 전부 일치한다 (`Session` 타입 인용 0건).

### AC-CMA-004 — known-5 노트 보존
- **Given** modules.md, **When** `internal/design`·`internal/migrate`·`internal/state`·`internal/research`·`internal/evaluator` 각 토큰의 modules.md 내 출현을 검색하면, **Then** 각 1건 이상이 blockquote 경고 노트 형태로 잔존하고 (D4) 비-blockquote(양성) 출현은 0건이다.

### AC-CMA-005 — citations 검사 3방향 (뮤턴트 포함)
- **Given** `internal/graph`·`internal/cli`에 citations 축이 구현된 상태, **When** 테스트 픽스처에 양성 부재 경로 1개를 포함한 codemaps 모방 문서로 `CheckFreshness`(또는 검사 함수)를 실행하면, **Then** citations 행 verdict=stale이고 `CheckResult.Failed()`가 true — CLI에서 exit 1.
- **Given** blockquote 행에만 부재 경로가 있는 모방 문서, **When** 동일하게 실행하면, **Then** citations 행 verdict=fresh (면제 성립 — D2).
- **Given** M2 완료 후의 실제 codemaps 6개 문서, **When** `moai graph check`를 실행하면, **Then** exit 0 (citations 행 포함 전 레이어 관측; 신선도 레이어 red는 별개 축이므로 citations 행만 판정 대상으로 읽는다 — 필요시 `--json`으로 citations 행 verdict만 판정).
- **뮤턴트**: 수정된 modules.md에 가짜 양성 유령(예: `internal/zzz-phantom`을 비-blockquote 제목으로) 주입 → 검사 red → 원복 → green. 이 왕복을 관측 기록으로 남긴다.

### AC-CMA-006 — 스킬 재발 방지 지시 (Template-First)
- **Given** 템플릿 정본 `internal/template/templates/.claude/skills/moai/workflows/codemaps.md`, **When** Phase 4 절을 검사하면, **Then** (1) 실행 가능한 존재 검증 수단으로 `moai graph check`(citations 행)를 명시하는 문구와 (2) 부정 인용은 blockquote 표기로 한다는 규약이 존재한다.
- **Given** 로컬 미러 `.claude/skills/moai/workflows/codemaps.md`, **When** 동일 검사를 실행하면, **Then** 동일 문구가 존재한다 (양본 동일 변경) — `make build` 성공 로그가 근거.

### AC-CMA-007 — post-absorb 재측정 기록 (SHOULD)
- **Given** run-phase 개시, **When** progress.md §E.2를 읽으면, **Then** 흡수 커밋 좌표와 §1.1 8행·§1.2 3지점의 재측정 명령+출력이 기록되어 있다.

### AC-CMA-008 — F1 정정 + t432 무결
- **Given** `.moai/reports/t304/` 산출물, **When** F1 정정 기록(t432 보고서 §3.1 제목 "26항목" vs 27행 표의 불일치)을 검색하면, **Then** 1건 존재한다.
- **Given** t432 워크트리, **When** `git -C .claude/worktrees/t432 status --short`를 실행하면, **Then** 본 카드 소행 변경 0건 (기존 lane-7 로컬 파일 제외 — 본 카드가 쓴 흔적만 판정).

### AC-CMA-009 — 완료 판정의 근거 계약
- **Given** run-phase 완료 보고, **When** 판정 근거를 검사하면, **Then** AC-CMA-001~006의 관측된 명령 출력이 근거로 제시되며, 신선도 게이트 green(스탬프·described-source-diff)은 정확성 근거로 인용되지 않는다 (REQ-CMA-010).

## §D.1 심각도 정의
- **MUST**: 미충족 시 카드 종결 불가. AC-001~005·008·009.
- **SHOULD**: 미충족 시 종결 가능하나 Gaps에 기록. AC-007(병합 순서상 대부분 자동으로 수행되나 기록 형식이 판정 대상).

## §D.2 간접 검증 (뮤턴트)
- AC-CMA-005의 뮤턴트 왕복(가짜 유령 주입 → red → 원복 → green)은 검사가 공허하게 초록일 수 없음을 폐쇄한다 — Vacuous-green 방지.
- AC-CMA-003의 경우 registry.go 실서명 대조가 곧 근원 검증이다 (문서↔코드 대조, grep 토큰 존재만이 아님 — t432 Residual-risk 교훈: "HIT 판정은 토큰 존재 수준" 한계를 서명 대조로 넘어선다).

## §D.3 종결 관문 (Definition of Done)
1. AC-CMA-001~006·008·009 전부 MUST/달성 관측 기록 존재.
2. `go test ./internal/graph/... ./internal/cli/...` 초록 (본 카드 변경 패키지 한정 — 전체 스위트는 CI).
3. `make build` 성공 (M3 스킬 편집 후).
4. 기존 가드 통과: mirror-parity·template 중립성·lint/vet (변경 범위 패키지).
5. 커밋은 레인 세션이 수행; 본 plan-phase 산출물은 커밋하지 않는다.

## §D.4 Forward-looking
- citations 축이 안정化되면 `.moai/project/*.md`(product/structure/tech)로의 확장이 별도 카드 후보가 된다 — 본 SPEC은 codemaps 6개 파일로 한정 (Out of Scope).
- codemaps 재생성 replay(스킬 지시가 실제 재생성 산출물을 규약 준수로 유지하는지)은 t432 스타일 재생성 카드에서 관측한다 — 본 카드는 스킬 본문 삽입(AC-CMA-006)까지만 검증한다.
