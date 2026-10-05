# SPEC-RELATION-PICKUP-FILTER-001 — 인수조건

검증 계층. 각 AC는 RED-now(현행 실패의 측정 근거)와 green 경로(명령 + 기대
출력)를 한 쌍으로 적는다 — card t1343 게이트 요건. 명령의 모듈 경로는
`github.com/modu-ai/moai-adk`(go.mod). 테스트 이름 패턴은 run-phase M1·M2·M4에서
이 표의 패턴에 고정한다. 모든 테스트는 `t.TempDir()` 저장소를 쓴다(운영자 라이브
큐 접촉 금지 — plan.md §C).

## AC-RPF-001 — 관계가 남은 queued 카드는 픽업 후보에서 제외된다

**Covers**: maps REQ-RPF-001

- **Given** 큐에 queued 카드 A·B가 있고, `A depends B` 관계가 기록돼 있으며,
  **When** `autoPickTargets`가 후보를 선택하면,
  **Then** 반환 목록에 A가 없고 B만 있다 (b) — B는 아무 관계가 없으므로 후보다.
- **RED-now**: `internal/cli/todo_auto.go:162-166` 의 queued 분기는 상태만 보고
  무조건 추가한다 — 관계 소비 코드가 트리에 없다(비테스트 grep 0적중, spec.md
  §A.2 F-2). A가 오늘 후보에 오는 것은 코드 구조상 확정.
- **green**: `go test -timeout 30m -count=1 ./internal/cli/ -run
  'TestAutoPickTargets'` → `ok github.com/modu-ai/moai-adk/internal/cli`.
  릴리스 차단 기준: 이 명령의 `ok` 출력.

## AC-RPF-002 — 방향: blocks 는 related 를, depends 는 subject 를 막는다

**Covers**: maps REQ-RPF-002

- **Given** `A blocks B`가 기록된 큐에서 **When** 픽업을 고르면 **Then** B만
  제외된다(A는 계속 후보). **Given** `A depends B`가 기록된 큐에서 **When**
  픽업을 고르면 **Then** A만 제외된다(B는 계속 후보).
- **RED-now**: F-2와 동일 — 방향 해석 코드가 없다. 한 방향만 맞춘 구현은 이
  AC의 두 시나리오 중 하나를 실패시키게 설계된다(plan.md E-2).
- **green**: 같은 명령(AC-RPF-001)에 양방향 테이블 케이스로 흡수 → `ok`.

## AC-RPF-003 — 선행 카드가 done 되면 후보로 돌아온다

**Covers**: maps REQ-RPF-003

- **Given** `A depends B`로 A가 제외 중이고, **When** B가 `done` 되어
  `ArchiveCard`가 B를 이름대는 관계를 아카이브 항목으로 옮기면(기존 동작 —
  `backlog_store.go:308-341`), **When** 다시 픽업을 고르면 **Then** A가 후보다.
  별도의 관계 정리(bookkeeping)는 없다.
- **RED-now**: 이 AC는 기존 해소 경로(F-4)를 회귀 고정으로 묶는다 — 필터가
  "관계 존재"만 보는 한 done 뒤 자동 풀림이 따라온다. 상태 스캔 구현이 끼어들면
  이 AC가 그 결함을 잡는다.
- **green**: `go test -timeout 30m -count=1 ./internal/cli/ -run
  'TestAutoPickTargets'` → `ok`.

## AC-RPF-004 — 건너뛴 카드는 표식이 남고 종료 코드는 0이다

**Covers**: maps REQ-RPF-004

- **Given** queued 카드 A·B가 모두 관계로 막혀 있고, **When** `todo --auto`
  주기가 후보를 찾으면, **Then** 카드마다 한 줄의 표식 비발견(non-finding)이
  카드 id·관계·막는 선행 카드 id를 이름대며 인쇄되고, 주기는 "no eligible card"
  문구와 함께 exit 0으로 끝난다.
- **RED-now**: 표식이 존재하지 않는다 — 필터 자체가 없으므로 건너뛰기도 없다
  (M3가 함께 착지).
- **green**: `go test -timeout 30m -count=1 ./internal/cli/ -run
  'TestAutoPickTargets|TestRunAutoCycle'` → `ok`, 표식 문구 단정 포함.

## AC-RPF-005 — 순환 관계 기록은 거부된다

**Covers**: maps REQ-RPF-005

- **Given** `A depends B`가 기록돼 있고, **When** `todo relate B A --relation
  depends`를 실행하면, **Then** 명령은 0이 아닌 종료 코드로 실패하고 오류가
  양쪽 카드 id를 이름대며, 큐 기록은 변경되지 않는다(관계 수 불변).
- **RED-now**: `runTodoRelate`(`todo_relate.go:60-94`)의 검사는 관계 어휘·
  자기관계·카드 존재·튜플 중복 4가지뿐이다 — 그래프 순회가 없어 오늘 이
  명령은 성공한다(F-3, 코드 인용).
- **green**: `go test -timeout 30m -count=1 ./internal/cli/ -run
  'TestTodoRelate'` → `ok`; 수동 재현은 임시 저장소에서 `todo relate` 2회 —
  2회째 비영 종료 + 관계 수 1 유지.

## AC-RPF-006 — 3-순환은 거부되고 열린 사슬은 허용된다

**Covers**: maps REQ-RPF-005

- **Given** `A depends B`, `B depends C`가 기록돼 있을 때 `C depends A`는
  **거부**된다(A→B→C→A). 같은 큐에서 `C depends D`(새 카드 D)는 **허용**된다.
  그리고 `A depends B`가 있을 때 `B blocks A`는 같은 waits-on 변(A→B)의 다른
  어휘 표기일 뿐 순환이 아니므로 **허용**된다(plan.md B.3).
- **RED-now**: F-3 — 순환 판정 코드가 없다.
- **green**: `go test -timeout 30m -count=1 ./internal/cli/ -run
  'TestTodoRelate'` → `ok` (3-순환·열린 사슬·역철자 케이스 포함).

## AC-RPF-007 — 경계 회귀 고정: 구조 arm 은 그대로, 다른 관계는 무관

**Covers**: maps REQ-RPF-006, REQ-RPF-007

- **Given** 소유자가 죽은 picked 카드가 관계로 막혀 있어도 구조(rescue) arm은
  그것을 건드리지 않는다 (REQ-RPF-006). **Given** `contains` 관계만 있는
  queued 카드는 계속 후보다 (REQ-RPF-006). **Given** GTD 테이블
  (`gtd_relations`)의 관계는 픽업 판정에 읽히지 않는다 (REQ-RPF-007).
- **RED-now**: 필터가 없어 오늘 아무것도 막히지 않는다 — 이 AC는 green 뒤
  미래의 과잉 구현을 잡는 핀이다.
- **green**: `go test -timeout 30m -count=1 ./internal/cli/ ./internal/kanban/`
  → 두 패키지 `ok`.

## 채택 셀 — RED-now 판별과 green 마일스톤

| AC | RED-now 근거 | green 마일스톤 |
|---|---|---|
| AC-RPF-001 | todo_auto.go:162-166 무조건 append (코드 인용) | M1 |
| AC-RPF-002 | 방향 해석 부재 (F-2) | M1 |
| AC-RPF-003 | ArchiveCard 관계 이동 기존 동작 — 회귀 고정 (F-4) | M1 |
| AC-RPF-004 | 표식 부재 (필터 부재의 귀결) | M3 |
| AC-RPF-005 | runTodoRelate 검사 4종, 그래프 순회 없음 (F-3) | M2 |
| AC-RPF-006 | 동일 (F-3) | M2 |
| AC-RPF-007 | 핀 — 과잉 구현 감지용 | M4 |

## 품질 게이트와 완료 정의 (Definition of Done)

- 변경 패키지 전부: `go vet ./internal/cli/... ./internal/kanban/...` →
  출력 없음·exit 0. `golangci-lint run ./internal/cli/... ./internal/kanban/...`
  → 0 findings(CI 판 golangci 버전 준수). `go test -timeout 30m -count=1`로
  두 패키지 → `ok` ×2. 커버리지는 `quality.yaml` `test_coverage_target`(85%)
  기준 유지.
- DoD: AC-RPF-001..007 전부 green + B-I-1의 두 doctrine 주석 블록이 새 소비자를
  이름대고 + 관련 문서 표면(list/why/web)이 불문이라는 점 확인 + 위 게이트
  전부 통과. RED-first 증거(각 마일스톤의 RED 실패 출력)가 progress.md §E.2에
  남는다.
