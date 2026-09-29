# SPEC-TODO-CLASSIFY-DISPATCH-001 — 인수조건

> 각 항목은 Given-When-Then 이며 이진 판정 가능하다. 각 AC 는 머리에 **Covers** 로 자기가 덮는
> `REQ-TCD-XXX` 를 인용한다. "테스트" 는 run 페이즈가 `internal/kanban`·`internal/cli`·
> `internal/hook` 에 두는 이름 붙은 Go 테스트이며, 판정은 그 테스트의 실제 실행 출력으로 한다.
> 문서·부재 기재는 파일 존재·본문 grep 으로 판정한다.
>
> **RED-now / green-path 규율.** 모든 AC 는 말미 채택 셀 표에서 RED-now 판별(측정 트리 `145c3d98c`
> 기준)과 green 으로 뒤집는 마일스톤을 함께 적는다. 지금 실패하는 것이 무엇인지 명시되지 않은 AC 는
> 기각된다. 회귀 가드로 선언된 행은 예외다 — 기존 코드가 이미 요구 행동을 달성 중인 경우 그 AC 는
> 해당 마일스톤 시점부터 GREEN 이어야 하고, green 의 이유가 "올바른 이유" 임을 변이·양성 대조로
> 시험한다. 전체 14/14 (Tier M 상한 16 이내).

## AC-TCD-001 — 생성 시 분류 기록 E2E

**Covers**: maps REQ-TCD-001

**Given** 분류 decider 가 판정을 반환하도록 연결된 프로젝트가 있고
**When** `moai todo add "<text>"` 로 카드를 추가할 때
**Then** 큐 레코드의 그 카드에 `priority`·`blocked`·`mode`·`decider`·`classified_at`·`reason` 를
포함한 분류가 기록돼 있고, 같은 잠금 쓰기 안에서 큐가 우선순위 순으로 재정렬돼 있다. 테스트는
`--json` 큐 판독으로 필드값을 단언한다.

## AC-TCD-002 — 첨가 필드의 바이트 동일 라운드트립

**Covers**: maps REQ-TCD-002, REQ-TCD-014

**Given** 이 SPEC 이전 형식으로 기록된 큐 파일(SQLite 와 JSON 쌍)이 있고
**When** 새 바이너리로 열고 아무 카드도 추가하지 않고 다시 쓸 때
**Then** 큐 파일은 스키마 마이그레이션을 포함해 시도 전 바이트와 동일하다(분류 없는 카드의
marshal 이 byte-identical). 그리고 같은 저장소에 카드를 하나 추가하면, 그 카드에만 분류가
나타나고 기존 카드의 분류 필드는 계속 부재(`null`/`NULL`)다.

## AC-TCD-003 — decider 부재 시 기본값 승격

**Covers**: maps REQ-TCD-003

**Given** 분류 decider 가 unavailable/failure 를 반환하도록 변이된 상태가 있고
**When** `moai todo add "<text>"` 를 실행할 때
**Then** 카드는 승인되고 분류는 `priority=normal, blocked=false, mode=serial,
decider=default` 이며, stderr 에 fallback 통지가 정확히 1행 출력된다. 양성 대조로 decider 가
정상일 때는 실제 판정값이 기록됨을 같은 테스트가 확인한다 — 기본값이 "항상 찍히는 값" 이 아니라
실패 경로만의 값임을 변이로 분리한다.

## AC-TCD-004 — 명시적 분류 입력의 검증과 jev 거부

**Covers**: maps REQ-TCD-004

**Given** 닫힌 값집합에 맞는 분류 JSON 과, 집합 밖 값(`priority: "urgent"`, `mode: "serially"`,
`decider: "jev"`)을 담은 JSON 들이 있고
**When** 각각 `todo add --classification-file` 로 넣을 때
**Then** 유효한 JSON 은 그대로 기록되고 공급 decider 신원이 남으며, 집합 밖 값과 `decider: "jev"`
는 usage 오류(exit 2)로 거부되고 큐에 아무것도 쓰이지 않는다.

## AC-TCD-005 — 정렬 키: blocked 우선 · priority · 삽입순 안정성

**Covers**: maps REQ-TCD-005

**Given** 빈 큐에 순서대로 A(normal)·B(high)·C(low)·D(normal, blocked=true) 를 추가할 때
**When** 네 추가가 모두 끝난 뒤 큐 순서를 읽을 때
**Then** 순서는 B, A, C, D 다 — blocked 는 모든 non-blocked 뒤로 가라앉고, 같은 rank 안에서는
삽입 순서가 유지된다. rank 를 뒤섞는 추가 하나를 변이로 넣으면(예: D 를 non-blocked 로) 순서가
그에 따라 바뀜도 함께 단언한다.

## AC-TCD-006 — 위치 출력과 렌더가 정렬 순서를 말한다

**Covers**: maps REQ-TCD-006

**Given** AC-TCD-005 의 정렬된 큐가 있고
**When** 다음 카드를 추가해 `<id> <position>` 이 출력되고 `moai todo list` 와 웹 큐 판독을 할 때
**Then** add 가 찍는 position 은 정렬된 순서 기준 1-based 위치이고, list 와 웹 렌더는 같은 순서로
행을 출력한다.

## AC-TCD-007 — factory next 는 최상위 우선순위를 임대한다

**Covers**: maps REQ-TCD-007

**Given** low 우선순위 queued 카드와 high 우선순위 queued 카드, 그리고 blocked queued 카드가 큐에 있고
**When** 레인 세션이 `moai factory next` 를 실행할 때
**Then** 임대되는 카드는 high 카드고, blocked 카드는 자동 선택되지 않는다(연속 `next` 로도
승격되지 않음을 반복 실행으로 단언한다).

## AC-TCD-008 — serial 카드의 상호 독점·우선순위 서빙·비행 중 parallel 선택

**Covers**: maps REQ-TCD-008

**Given** serial 카드 둘(우선순위 high·low)과 parallelizable 카드 하나가 큐에 있고
**When** 레인 A 가 high serial 카드를 임대한 뒤 각 단계가 실행될 때
**Then** (a) 레인 B 의 `factory next` 는 parallelizable 카드를 임대한다 — serial 카드 비행 중에도
parallelizable 선택은 영향받지 않는다, (b) high serial 이 양으로 열거된 terminal 상태에 도달하기
전에는 어떤 레인도 low serial 카드를 임대하지 못하고, parallelizable 후보가 남아 있지 않은 상태의
`next` 는 기존 no-card exit 계약(exit 3)으로 거부된다, (c) high serial 이 terminal 에 도달한 뒤
low serial 이 임대된다 — serial 카드는 우선순위 순서대로 한 번에 하나씩 서빙된다. 테스트는
terminal 열거에서 하나를 빼는 변이에 실패해야 한다(양적 열거 규율).

## AC-TCD-009 — parallelizable 동시 다중 레인 임대

**Covers**: maps REQ-TCD-009

**Given** 서로 다른 parallelizable 카드 둘이 큐에 있고
**When** 레인 A 와 레인 B 가 `factory next` 를 동시에(고루틴 경합으로) 실행할 때
**Then** 두 레인 모두 성공하고 서로 다른 카드를 임대하며, 두 레인 모두 재시도 없이는 물론
재시도를 포함해도 각자 정확히 한 카드를 가진다.

## AC-TCD-010 — 중복 디스패치 음성: 같은 카드는 두 번 임대되지 않는다

**Covers**: maps REQ-TCD-009 (회귀 가드)

**Given** 같은 queued 카드를 두 레인이 동시에 임대 시도하도록 경합을 구성하고
**When** 경합이 마친 뒤 factory 기록을 읽을 때
**Then** 그 카드는 정확히 한 레인의 holder 만 가지며, 버전 검사에 진 레인은 다른 카드로
수렴하거나 no-card 로 끝난다. 이 AC 는 t1240 중복 방지 기계 위의 회귀 가드로, t1240 흡수 시점부터
GREEN 이어야 하고 green 의 이유가 버전 검사임을 holder 불일치 변이로 시험한다.

## AC-TCD-011 — status 가 레인↔카드·mode·priority 를 보여준다

**Covers**: maps REQ-TCD-010

**Given** 레인 A 가 serial high 카드를, 레인 B 가 parallelizable normal 카드를 임대 중이고
**When** `moai factory status`(텍스트와 `--json`)를 읽을 때
**Then** 각 카드 행에 holder 레인과 `mode`, `priority` 가 함께 출력되고, 두 출력 형태가 같은
값을 말한다.

## AC-TCD-012 — -f 레인 자동 디스패치 기본값과 옵트아웃

**Covers**: maps REQ-TCD-011

**Given** 팩토리가 실행 중이고
**When** `moai cc -f lane` 으로 레인을 시작할 때
**Then** 레인 부트스트랩이 자동 디스패치(self-dispatch 루프 진입) 지시를 운반한다. 같은 시작을
`--no-auto-dispatch` 로 하면 부트스트랩이 수동 모드를 운반한다. 테스트는 두 시작의 부트스트랩
페이로드를 스냅샷 비교한다.

## AC-TCD-013 — scripts/jev 제품 경로 부재 (양성 대조 동반)

**Covers**: maps REQ-TCD-012

**Given** 제품 경로 집합 `internal/ pkg/ cmd/ internal/template/templates/` 가 있고
**When** `scripts/jev` 참조(임포트·exec·경로 상수)를 grep 할 때
**Then** 적중 0건이다. 같은 grep 이 통제 경로(실제로 참조가 존재하는 로컬 전용 문서)에서는
적중함을 양성 대조로 확인한다 — 대조 없는 0 은 부재의 증거가 아니다.

## AC-TCD-014 — 흡수 순서 진입 전제의 기계 판정

**Covers**: maps REQ-TCD-013

**Given** run 페이즈 진입 시점이 있고
**When** 로컬 develop 의 `internal/cli` 에서 `newFactoryNextCommand`·`newFactoryCompleteCommand`
심볼 존재를 검사할 때
**Then** 두 심볼이 모두 존재하면 M3 이 진행 가능하고, 하나라도 부재하면 run 진입은 흡수 전제
위반으로 차단된다. 계획 문서(spec.md REQ-TCD-013, plan.md §D.1)에 흡수 원천
(`WT-factory-self-dispatch`)과 검사 명령이 기재돼 있음을 문서 grep 으로 단언한다.

## 채택 셀 — RED-now 판별과 green 마일스톤

| AC | RED-now (145c3d98c 기준) | green 마일스톤 |
|---|---|---|
| AC-TCD-001 | RED — 분류 필드·기록이 없다 | M2 |
| AC-TCD-002 | GREEN-at-M1 회귀 가드 — 첨가 규율 자체가 요구 사항(현행 저장소가 바이트 보존을 달성), 변이로 시험 | M1 |
| AC-TCD-003 | RED — fallback 경로가 없다 | M2 |
| AC-TCD-004 | RED — 입력 표면이 없다 | M2 |
| AC-TCD-005 | RED — 정렬이 없다(삽입순) | M2 |
| AC-TCD-006 | RED — 위치가 삽입순 | M2 |
| AC-TCD-007 | RED — t1240 흡수 전 `next` 자체가 develop 에 없다(전제 차단, AC-TCD-014 참조) | M3 |
| AC-TCD-008 | RED — mode 개념 부재 | M3 |
| AC-TCD-009 | t1240 흡수 후 RED — 동시 임대는 가능하나 AC pin 이 없다 | M3 |
| AC-TCD-010 | GREEN-at-M3 회귀 가드 — t1240 버전 검사가 이미 달성, holder 변이로 시험 | M3 |
| AC-TCD-011 | RED — mode/priority 표시 부재 | M3 |
| AC-TCD-012 | RED — 기본값·플래그 부재(부트스트랩은 항상 자동 지시를 운반, 조건화 안 됨) | M4 |
| AC-TCD-013 | GREEN-at-M1 회귀 가드 — 오늘도 참조 0건, 양성 대조로 시험 | M1 |
| AC-TCD-014 | RED — 전제가 문서·기계 검사로 존재하지 않는다 | M3 (문서는 M1 착지) |
