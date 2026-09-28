# SPEC-TODO-HOLD-STATE-001 — 인수조건

> 각 항목은 Given-When-Then 이며 이진 판정 가능하다. 각 AC 는 머리에 **Covers** 로 자기가 덮는
> `REQ-THS-XXX` 를 인용한다. "테스트" 는 run 페이즈가 `internal/kanban`·`internal/cli`·
> `internal/statusline`·`internal/web` 에 두는 이름 붙은 Go 테스트이며, 판정은 그 테스트의 실제
> 실행 출력으로 한다. 문서 기재는 파일 존재·본문 grep 으로 판정한다.
>
> **RED-now / green-path 규율.** 모든 AC 는 말미의 채택 셀 표에서 RED-now 판별(측정 트리 `8a969dfc0`
> 기준)과 green 으로 뒤집는 마일스톤을 함께 적는다. 지금 실패하는 것이 무엇인지 명시되지 않은 AC 는
> 기각된다 — 통과만 가능한 AC 는 드리프트를 못 잡는다.

## AC-THS-001 — 네 상태 열거와 CHECK 구속

**Covers**: maps REQ-THS-001

**Given** 새로 초기화된 백로그 저장소가 있고
**When** 네 상태 값 `queued`·`picked`·`dropped`·`hold` 를 각각 가진 행을 `items` 에 삽입할 때
**Then** 네 삽입이 모두 성공하고, 다섯째 임의 값(`"held"`, `"onhold"`, `""`)의 삽입은 CHECK 위반으로
실패한다. 테스트는 테이블 스키마를 읽어 CHECK 튜플에 네 리터럴이 정확히 있는지도 단언한다(값 하나를
빼거나 다섯째 값을 몰래 넣으면 실패).

## AC-THS-002 — v1 → v2 마이그레이션 라운드트립: 행과 아카이브가 보존된다

**Covers**: maps REQ-THS-002

**Given** v1 스탬프(`schema_version = "1"`) 데이터베이스에 상태를 섞은 카드 N행과 `archived_items`
행 M개, `findings` 행을 심어 두고
**When** 현재 바이너리로 저장소를 열 때
**Then** (a) 열기가 성공하고 스탬프가 `"2"` 로 바뀐다, (b) N행 전원이 id·text·added_at·spec_id·
state·landing 전 필드 동일하게 존재한다, (c) `archived_items` M행과 `findings` 가 바이트 동일하게
보존된다, (d) 마이그레이션은 멱등하다 — 같은 데이터베이스를 다시 열어도 행 수와 스탬프가 변하지
않는다.

## AC-THS-003 — 패리티 실패 시 원본 무손상

**Covers**: maps REQ-THS-003

**Given** 패리티 검증이 실패하도록 변이된 마이그레이션 경로(예: 행 수 불일치를 유발하는 변형)와 v1
데이터베이스가 있고
**When** 저장소 열기를 시도할 때
**Then** 열기가 실패로 끝나고, 원본 데이터베이스 파일이 마이그레이션 시도 전 바이트와 동일함이
파일 해시로 단언되며, 데이터베이스와 레거시 아티팩트 어느 쪽도 삭제·덮어쓰기·격리되지 않는다.

## AC-THS-004 — 커밋과 스탬프가 한 트랜잭션에

**Covers**: maps REQ-THS-004

**Given** v1 데이터베이스가 있고
**When** 마이그레이션 커밋 직후 저장소를 닫고 `meta` 를 직접 조회할 때
**Then** `schema_version` 이 `"2"` 이고 `items` 의 CHECK 가 넷 값이다. 변이 시험: 스탬프를 `"1"` 로
되돌린 채 두면 다음 열기에서 마이그레이션이 다시 시도되어 성공해야 하고(멱등), 스탬프를 알 수 없는
값으로 바꾸면 AC-THS-005 의 거절이 관측된다 — 스탬프 없이 CHECK 만 넓힌 중간 상태는 어느 테스트도
통과하지 못한다.

## AC-THS-005 — 구(舊) 바이너리가 신(新) 데이터베이스를 거절한다

**Covers**: maps REQ-THS-005

**Given** `schema_version = "2"` 로 스탬프된 데이터베이스와, 이전 스탬프만 아는 바이너리(테스트는
현재 코드의 버전 스위치를 이전 값으로 고정한 변형 또는 reader 경로의 버전 상수 주입으로 재현)가 있고
**When** 그 바이너리로 저장소를 열 때
**Then** 열기가 `unsupported schema_version` 을 사유로 하는 `ErrBacklogCorrupt` 로 실패하고, `items`
행이 하나도 읽히지 않으며, 파일이 쓰이지 않는다(수정 시각·바이트 동일성 단언). 반대 방향 양성 대조:
같은 바이너리가 `"1"` 스탬프 데이터베이스는 정상 연다.

## AC-THS-006 — hold 동사: queued → hold

**Covers**: maps REQ-THS-006

**Given** queued 카드 한 장이 있고
**When** `moai todo hold <id>` 를 실행할 때
**Then** 카드 상태가 `hold` 로 바뀌고, id·text·added_at·spec_id·landing 이 실행 전과 동일하며, stdout
한 줄이 id 와 텍스트 접두어를 운반한다. `--expect <prefix>` 불일치 시는 쓰기 없이 거절된다(drop 과
같은 `Mutate` 계약).

## AC-THS-007 — hold 거절: queued 가 아닌 카드

**Covers**: maps REQ-THS-007

**Given** picked·dropped·hold 상태의 카드가 각각 있을 때(표 테스트 3행)
**When** 각각에 `moai todo hold <id>` 를 실행할 때
**Then** 세 경우 모두 거절이고 오류 메시지가 카드의 현재 상태와 회복 동사(picked → unpick 먼저)를
이름하며, 저장소 레코드가 실행 전 바이트와 동일하다.

## AC-THS-008 — unhold 동사: hold → queued, 본문 무손상

**Covers**: maps REQ-THS-008

**Given** hold 상태의 카드 한 장이 있고
**When** `moai todo unhold <id>` 를 실행할 때
**Then** 상태가 `queued` 로 돌아가고 카드 본문이 바이트 동일하며(마커가 존재하지 않음이 본문 전체
비교로 단언된다), stdout 한 줄이 id 와 텍스트 접두어를 운반한다. hold → unhold → hold 를 두 번
되풀이해도 레코드가 동일하게 수렴한다.

## AC-THS-009 — unhold 거절: hold 가 아닌 카드

**Covers**: maps REQ-THS-009

**Given** queued·picked·dropped 상태의 카드가 각각 있을 때
**When** 각각에 `moai todo unhold <id>` 를 실행할 때
**Then** 세 경우 모두 거절이고 쓰기는 없다.

## AC-THS-010 — 행위자 경계: 임대 경로에 hold 표면이 없다

**Covers**: maps REQ-THS-010

**Given** 현재 바이너리의 전체 todo/lease 명령 표면이 있고
**When** (a) `--help` 전수와 (b) 소스 grep 으로 임대·레인 경로(factory next, `-f` lane claim,
`--auto` 실행, autodone)에서 `hold` 상태를 설정하거나 해제하는 코드 경로를 찾을 때
**Then** (a) hold/unhold 는 연산자가 실행하는 최상위 todo 동사로만 존재하고 임대 경로의 플래그·
환경·설정으로는 도달할 수 없으며, (b) `BacklogStateHold` 를 대입하는 문장이 임대·레인 경로 파일에
0건이다(양성 대조: `todo hold` 구현 파일에는 1건 이상). 리드 자가 승격 금지 규정은 본 AC 로
변경되지 않는다.

## AC-THS-011 — 선형 선택 술어의 양(陽) 열거 전환 (변이 시험 가능)

**Covers**: maps REQ-THS-011, REQ-THS-012

**Given** M3 스윕이 산출한 실현목록의 각 actionable 표면(next bare, `next <n>`, machine lease,
`--auto`, autodone 후보 스캔)에 대해
**When** (a) 소스 grep 이 해당 표면의 상태 비교를 검사하고, (b) 변이 시험으로 다섯째 상태 값
`"future_state"` 를 가진 카드를 심어 각 선택 경로를 실행할 때
**Then** (a) 모든 actionable 선택 술어가 허용 상태를 긍정 열거한다(`State != X` 형태의 부정 필터가
0건 — 양성 대조: 드롭 게임 같은 긍정 switch 는 세고, 부정 필터는 세지 않는다), (b) 다섯째 상태의
카드는 어떤 actionable 경로로도 선택·종료되지 않는다. (b)가 RED 를 내는 방식이 변이 기반이므로,
새 상태가 추가돼 스윕을 안 돌리면 이 AC 가 실패한다 — 드리프트 탐지의 축이다.

## AC-THS-012 — 기계 임대자는 queued 만 임대한다

**Covers**: maps REQ-THS-013

**Given** queued 2장·hold 1장·picked 1장이 섞인 큐가 있고
**When** 기계 임대 선택(factory next, Codex `-f` lane claim 경로의 선택 함수)을 실행할 때
**Then** 선택 집합은 queued 2행 정확히이고, hold 카드는 구성상 건너뛰어진다(순서를 뒤섞은 5회
실행에서 동일). 변이: hold 카드의 상태를 queued 로 되돌리면 선택 집합이 3행으로 늘어나는 양성
대조를 같은 테스트가 단언한다 — 임대자가 hold 를 무시하는 게 아니라 보는지 증명하는 유일한 방법.

## AC-THS-013 — hold 카드에 대한 pick 거절

**Covers**: maps REQ-THS-014

**Given** hold 상태 카드 한 장이 있고
**When** `moai todo next <id>` 로 픽을 시도할 때
**Then** 거절이고 오류 메시지가 unhold 를 회복 동사로 이름하며, 쓰기는 없다. 양성 대조: 같은 카드를
`unhold` 한 뒤 같은 픽은 성공한다.

## AC-THS-014 — autodone 후보에서 hold 제외

**Covers**: maps REQ-THS-015

**Given** landed-by-ref 조건을 만족하는 hold 카드 한 장과 같은 조건의 queued 카드 한 장이 있고
**When** auto-done 후보 스캔(`planAutoDone` 경로)을 실행할 때
**Then** 결과 집합에 queued 카드만 있고 hold 카드에 대한 outcome 이 없다. RED-now: 현재 트리에서
hold 가 없어도 이 AC 의 `!= Queued && != Picked` 스캔은 넷째 상태를 조용히 삼키는 모양이므로, M1 이
상태를 추가하는 순간 스캔이 hold 를 후보로 삼아 본 AC 가 RED 를 내야 하고 M3 이 뒤집는다.

## AC-THS-015 — list·--json 의 진실성

**Covers**: maps REQ-THS-016

**Given** hold 카드 한 장·queued 카드 한 장이 있고
**When** `moai todo list`(텍스트)와 `moai todo list --json` 을 실행할 때
**Then** 텍스트 뷰는 hold 카드를 `hold` 상태 표기와 함께 렌더링하고(숨김도 queued 로의 위장도 없음),
`--json` 레코드가 `"state":"hold"` 를 운반한다. `--json` 전체 레코드가 gtd.md:48 의 계약 — 상태
필드로 기계 필터링 가능 — 를 유지한다.

## AC-THS-016 — 상태줄 카운트가 hold 를 세지 않는다

**Covers**: maps REQ-THS-017

**Given** picked 1·queued 2·hold 1이 섞인 큐가 있고
**When** `BacklogCountsForRoot` / 상태줄 backlog 카운트를 실행할 때
**Then** `Picked == 1`, `Queued == 2` 이고 hold 카드는 어느 카운트에도 더해지지 않는다. 변이:
hold 카드를 queued 로 바꾸면 `Queued == 3` (양성 대조 — 카운트가 hold 를 '볼 줄 아는지').

## AC-THS-017 — 웹 콘솔 상태 표기

**Covers**: maps REQ-THS-018

**Given** hold 카드가 포함된 큐로 웹 콘솔 큐 뷰 VM(`todo_queue_read.go` 경로)을 만들 때
**When** 해당 카드의 `TodoItemVM` 을 검사할 때
**Then** `State` 필드가 리터럴 `"hold"` 이고 렌더링된 화면 마크업에 그 값이 나타난다.

## AC-THS-018 — 문서 표면과 템플릿 미러 동기화

**Covers**: maps REQ-THS-019

**Given** M5 이후의 문서 표면이 있고
**When** (a) `.claude/skills/moai/workflows/gtd.md`, (b) `.claude/commands/moai/todo.md`,
(c) 각각의 `internal/template/templates/**` 미러를 검사할 때
**Then** 네 파일 전원이 `hold` 상태와 `todo hold`·`todo unhold` 동사를 기재하고, 로컬 사본과 미러의
해당 절이 내용 동일하며(diff 단언), gtd.md 의 JSON 예시가 `"hold"` 상태 값을 포함한다. 미러 누락은
(a)만 통과해도 본 AC 를 실패시킨다.

## AC-THS-019 — 마이그레이션이 아카이브·landing 증거를 흔들지 않는다 (라운드트립 확장)

**Covers**: maps REQ-THS-002, REQ-THS-004

**Given** landing 증거가 기록된 카드와 `todo pr` 레코드가 붙은 아카이브 행이 있는 v1 데이터베이스가
있고
**When** 마이그레이션이 실행된 뒤 `todo list --json`, `todo history`, statusline 카운트를 실행할 때
**Then** landing 증거와 아카이브 조회 결과가 마이그레이션 전 조회와 동일하고, SPEC-TODO-DESTRUCTIVE-GUARD-001
의 저장 형상 단언 테스트가 무수정으로 통과한다(마이그레이션이 형제 SPEC 의 저장 계약을 깨지 않는다는
형제-테스트 재실행 증거).

## 채택 셀 (RED-now / green path)

측정 트리 `8a969dfc0`. `hold` 상태·동사·마이그레이션이 아직 없으므로 대부분의 RED-now 사유는
"대상 상태·동사·테스트 부재" 다. 이 부재가 **올바른 이유의 RED** 인 근거: 부재는 본 SPEC 의 run 이
만드는 것이고, 부재 아래에서도 이미 깨져 있는 행동(주석 두 행)은 별도로 식별해 둔다.

| AC | RED-now 판별 (측정 트리 기준) | green 마일스톤 |
|----|-------------------------------|----------------|
| AC-THS-001 | `grep -n "BacklogStateHold" internal/kanban/backlog_store.go` → 0건 | M1 |
| AC-THS-002..005 | `grep -n "items_new\|schema_version = \"2\"\|backlogSchemaVersion = \"2\"" internal/kanban/` → 0건 (마이그레이션 부재) | M1 |
| AC-THS-006..009 | `go run . todo hold x` → `Error: unknown command "hold"` | M2 |
| AC-THS-010 | 표면 부재로 조건부 통과 — M2 에서 hold 동사가 생긴 뒤에야 의미를 가진다(거절 경로 부재의 RED: `todo.go:920` 이 held 픽을 받아들이는 것, §B.5) | M2 |
| AC-THS-011 | **이미 깨져 있음**: `todo.go:903`의 `State != Queued`, `todo_autodone.go:283`의 부정 disjunction — 넷째 상태가 생기는 순간 조용히 삼킨다 (grep 2건) | M3 |
| AC-THS-012 | 임대 경로가 queued 만 고르는 것은 우연일 뿐 핀돼 있지 않음 — 변이 양성 대조 테스트 부재 (`grep -c "queued" internal/cli/factory_card.go` 는 계약이 아니다) | M3 |
| AC-THS-013 | **이미 깨져 있음**: `todo.go:920` 을 보면 held 픽 거절 경로가 없다 (거절 분기 부재) | M2 |
| AC-THS-014 | **이미 깨져 있음**: `todo_autodone.go:283`이 넷째 상태 후보를 삼키는 모양 (상태 추가 시 즉시 RED) | M3 |
| AC-THS-015..017 | `"hold"` 상태 값 부재 — list·statusline·web 테스트가 넷째 값을 단언할 수 없다 | M4 |
| AC-THS-018 | `grep -rn '"hold"' .claude/skills/moai/workflows/gtd.md` → 0건 | M5 |
| AC-THS-019 | 마이그레이션 부재로 재실행 자체가 불가 — M1 착지 후 형제 테스트 재실행이 green 증거 | M1 |
