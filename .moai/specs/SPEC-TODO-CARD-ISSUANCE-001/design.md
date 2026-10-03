# SPEC-TODO-CARD-ISSUANCE-001 — 설계

이 문서는 구조와 형식을 정한다. 코드는 쓰지 않는다. 이름은 **작업 이름**이며 실행 단계가 저장소의 명명 관례에 맞춰 바꿀 수 있다(수용 기준이 문자 그대로 요구하는 시험 이름만 고정이다). 근거 측정은 `research.md` §3, 선택지 표는 §11.

## §1 구조 개요

| 층 | 패키지 | 이 SPEC 이 더하거나 바꾸는 것 |
|---|---|---|
| 순수 조회·모형 | `internal/kanban` | 읽기 전용 이웃·구성요소·완료 SPEC 조회, 발행 속성·처분 타입, 관계 어휘 정규화·제약·카드↔GTD 해석기 |
| 저장 | `internal/kanban`(SQLite 엔진) | `items`·`archived_items`·`findings`·`archived_findings` 의 가산 컬럼, 읽기/쓰기/parity |
| 명령 | `internal/cli` | `add` 의 제시와 `--dry-run`, `todo trace`, `todo merge`, 처분 동사, `gtd engage` 제시, MCP `todo_add` |
| 그래프 | `internal/graph` | 카드 귀속 병합의 변경 파일 간선 층, 출처 지문 |
| 팩토리 | `internal/homestate`, `internal/cli` | `cards` 행의 묶음 속성, 적재 동사, 선택 호의 묶음 예약 |
| 웹 | `internal/web` | 관계 그래프 읽기 이음매와 보기 |
| 규칙 | `.claude/rules`, `.claude/skills`, `.claude/agents`, 템플릿 사본 | M5 |

의존 방향은 한쪽이다 — `kanban`(순수) ← `cli`/`web`/`graph`/`homestate`. 새 조회 함수는 입력으로 `BacklogRecord` 스냅숏을 받고 쓰지 않는다. 쓰기는 기존 `Mutate` 경로를 그대로 쓴다.

## §2 `add` 경로의 데이터 흐름

**현재**: 인자 파싱(`scanTodoAddArgs`) → 쓰기 전 검증(분류 JSON 파일) → `runTodoAddAppendRoot`: stale-store 공개(stderr) → `Mutate{ appendAnalyzedCard(분류·정확 중복 거절·near 소견·Jev) ; 분류 적용 ; 재정렬 }` → stdout `"<id> <pos>\n"`.

**새 흐름**(작업안):

1. 인자 파싱이 `--dry-run` 과 새 플래그(§4.3)를 받는다. 폴스루 경로는 플래그를 받지 않으므로 제시만 있고 `--dry-run` 은 없다.
2. **락 밖**: `LoadPure` 스냅숏을 읽는다.
3. **락 밖**: 제시를 계산한다 — 이웃(보관·dropped 포함), 구성요소, 완료 SPEC, 진행 중 레인 겹침. git·SPEC 디렉터리 같은 외부 읽기는 시간 상한 안에서만 한다(§3.4).
4. `--dry-run` 이면 제시를 stderr 로 내고 "dry-run: nothing was written" 한 줄을 더해 exit 0 으로 끝낸다. 정확 중복이면 "a real add would refuse: t<N> already holds this card" 를 내고 exit 0.
5. 아니면 **기존 `Mutate`** 로 admit 한다(분석기 불변).
6. admit 결과가 확정되면 제시를 stderr 로 낸다(거절이어도 낸다 — 거절 이유를 보강한다).
7. stdout 한 줄 `<id> <pos>`.

제시 계산이 스냅숏 시점이라 동시 admit 으로 생긴 카드를 놓칠 수 있다 — 제시는 조언이고 admit 의 정확성에 관여하지 않으므로 허용한다. 제시가 실패하거나 시간 상한을 넘기면 해당 항목만 `unmeasured` 로 나가고 admit 은 영향받지 않는다(REQ-TCI-003).

**MCP.** `handleTodoAdd` 가 같은 본문(`runTodoAddAppendRoot`)을 부르되, 버려지던 stderr 버퍼의 제시를 결과 텍스트의 첫 줄(`<id> <pos>`) 뒤에 한 빈 줄을 두고 덧붙인다. 제시가 비면 결과 텍스트는 CLI stdout 과 같다 — 기존 패리티 테스트의 고정 입력("mcp parity card", 빈 큐)에서 제시가 비므로 그대로 통과한다.

**engage.** `gtd engage` 의 CLI 계층이 admit 이 성공한 뒤 같은 조회를 읽기 전용으로 불러 stderr 로 낸다. 분석기·소견 기록·거절은 없다. 엔진 계층의 `EngageInput.DryRun` 이 이미 있으므로 그 선례와 맞춰 engage 의 `--dry-run` 은 이 SPEC 의 범위 밖이다.

## §3 제시 설계

### 3.1 입력과 항목

| 항목 | 입력 | 척도·표지 |
|---|---|---|
| 유사 카드 상위 3 | live(dropped 포함, 접두사 제거) + 보관 카드 | `measure=token-set-jaccard`, 점수, 상태(`live|dropped|archived`), 본문 앞 60자, dropped 이면 사유 |
| 같은 구성요소의 열린 카드 | 열린 카드(queued·picked·hold)의 본문 경로와 예상 파일 | `measure=component`, 구성요소 키, 카드 id·상태 |
| 진행 중 레인과의 예상 파일 겹침 | 신규 카드의 예상 파일 × 진행 중 레인 카드의 예상 파일(없으면 레인 브랜치의 변경 파일) | `measure=file-overlap`, 겹친 경로. 입력이 없으면 `unmeasured` 와 이유 |
| 이미 덮는 완료 SPEC | `status: completed` 인 SPEC 의 `card:`·`module:`·제목·태그 | `measure=spec-heuristic`, 표지 `heuristic` |

### 3.2 척도와 하한

기존 `NormalizeCardText`·`TokenSetJaccard` 를 그대로 쓴다(재사용, 새 척도 없음). 표시 하한은 `kanban` 패키지의 이름 있는 상수이고 **값은 M0 의 기준선 기록에서 읽는다**(작업 기본값 0.30, `plan.md` §F.11). 하한 미만의 어휘 이웃은 일반 `add` 에서는 출력하지 않고 `--dry-run` 에서만 상위 3개를 점수와 함께 보인다. 정규화 본문이 같은 카드(점수 1.0)는 하한과 무관하게 항상 보이며 라벨이 `exact` 다 — 이것이 이 제시에서 가장 확실한 항목이다(live 가 아닌 보관·dropped 카드와의 일치는 `ClassifyCardText` 가 못 보던 것이다).

근거(`research.md` §3.3): 하한 없이 항상 상위 3개를 보이면 추가의 91% 이상에서 점수 0.1 안팎의 소음이 나가고(SB02), 하한 0.3 으로 막으면 알림이 8.4%(최근 4.9%)로 줄지만 기록된 관계 상대의 재현율이 6% 로 떨어진다(SB03, SB05). 어휘 이웃은 의미 중복 판정의 대체물이 아니라 보조 신호이며, 의미 판정은 기존 Jev 소비자(발행 시 기록 전용)에 남는다(범위 밖 항목).

### 3.3 출력 형식 (작업안)

stderr, 항목마다 한 줄. 값이 없는 항목은 줄을 쓰지 않는다. 머리 줄과 항목 줄:

```
issuance: 겹침 제시 (읽기 전용, 아무것도 막지 않는다)
  similar   t812   archived  0.46  measure=token-set-jaccard  "<본문 앞 60자>"
  similar   t903   dropped   0.41  measure=token-set-jaccard  "<본문 앞 60자>" (사유: <reason>)
  exact     t77    archived  1.00  measure=normalized-equal   "<본문 앞 60자>"
  component t1411  picked    internal/cli  measure=component
  overlap   t1414  lane-2    internal/cli/todo.go  measure=file-overlap
  overlap   unmeasured (new card carries no expected files)
  spec      SPEC-TODO-ANALYSIS-001  completed  measure=spec-heuristic  heuristic
```

### 3.4 시간 상한과 락 규율

git 탐침(진행 중 레인 브랜치와 병합 기준 사이의 `git diff --name-only`)은 레인 브랜치 하나에 약 0.25초이고(SB08) `WT-` 브랜치가 353개, 워크트리가 74개다. 진행 중 레인 수만큼 직렬로 돌면 `add` 가 눈에 띄게 느려진다. 그래서 (a) 탐침은 `Mutate` 밖에서만, (b) 항목당·전체에 시간 상한, (c) 예상 파일이 있는 카드는 git 을 부르지 않고 필드끼리 비교, (d) 상한 초과는 `unmeasured (time bound)` 로 표기한다. 상한값은 M1 이 진행 중 레인 수와 탐침 지연을 먼저 재서 정한다(`plan.md` §F.11). 탐침은 락을 쥔 채 돌지 않는다는 것이 시험 가능한 성질이다(탐침 대기 중에 다른 프로세스의 `Mutate` 가 끝나야 한다).

### 3.5 분류기와의 분리

제시용 조회는 `ClassifyCardText` 를 부르지 않고 `kanban` 에 새로 둔 읽기 전용 함수다. 분류기는 dropped 카드를 거절 대상에서 제외한다는 독트린(`backlog_analysis.go` 의 함수 주석)과 고정 테스트를 가지고 있어 건드리지 않는다. 제시는 dropped 카드도 **알림 대상**으로 포함하되 카드를 거절하거나 소견을 기록하지 않는다.

## §4 스키마와 이주

### 4.1 형태 (D4 작업 기본값)

| 표 | 추가 | 형태 |
|---|---|---|
| `items`, `archived_items` | `issuance` | nullable TEXT, JSON 객체. 없음 = NULL |
| `findings`, `archived_findings` | `disposition` | nullable TEXT(`accept`/`merge`/`reject`). 없음 = NULL |

`issuance` JSON 의 키(모두 선택): `spawned_by`(카드 id), `origin`(닫힌 집합: `operator`, `leader`, `audit-finding`, `follow-up`, `ci-repair`, `standing`, `external`, `split`), `size_lines`(예상 제품 줄 수, 정수), `files`(예상 파일 경로 배열), `drop_reason`(문자열). 카드가 어느 키도 갖지 않으면 컬럼은 NULL 이고 직렬화에서 사라진다(`omitempty`). 닫힌 `origin` 집합은 카드 머리말 분류(`research.md` §3.2 의 02 스크립트 분류)를 겹치지 않는 값으로 줄인 것이다.

### 4.2 닿는 곳 (모두 한 마일스톤 M2 의 한 커밋 묶음)

| 닿는 곳 | 위치 | 이유 |
|---|---|---|
| 타입 | `backlog_store.go` 의 `BacklogItem`·`BacklogFinding` | 필드 추가, 포인터와 `omitempty` |
| 가산 컬럼 | `backlog_sqlite.go` 의 `ensureSchema` 마지막 가산 패스, `ensureColumn` | 버전 조정 뒤에 retrofit(카드 t1310 순서), `backlogItemsTableColumns` 에는 넣지 않음 |
| 쓰기 | `backlog_migrate.go` 의 `INSERT INTO items(...)`·`INSERT INTO archived_items(...)`(소견 표 두 곳 포함) | 열 목록 갱신 |
| 읽기 | `readSnapshot`·`readArchive`, `columnExpr` | 없는 컬럼은 NULL 로 읽어 DDL 없이 순수 읽기 유지 |
| parity | `assertBacklogParity` | 새 컬럼 포함 |
| 동결 튜플 | `backlog_schema_freeze_test.go` 의 `wantItemsColumns`·`wantArchivedItemsColumns` | 마지막 튜플로 `issuance:TEXT:0:NULL` 추가(소견 표가 동결돼 있으면 같은 방식), 같은 커밋 |
| 보관·복원 | `ArchiveCard`·`RestoreCard` | 항목 통째 복사라 필드는 따라가고, 소견 처분도 소견과 함께 이동하는지 시험 |
| 큐 병합·이주 | `todo_queue_merge.go`, `todo_merge_procedure.go` | 항목·소견 복사·재매핑 경로가 새 컬럼을 나르는지 |
| JSON | `todoJSONProjection`(`todo_claim.go`), `omitempty` | 새 필드가 없는 카드는 골든과 바이트 동일 |
| export | `todo_export.go` | 새 필드를 어떻게 다루는지(공백) |
| drop | `todo_drop.go` | 사유 속성 저장, 텍스트 접두사 유지 |

### 4.3 `add` 의 새 플래그와 닫힘 시각

`add` 서브커맨드(폴스루 아님)에 `--parent <id>`, `--origin <값>`, `--size-lines <n>`, `--files <경로,…>` 를 둔다(D13: 명시 입력만, 본문에서 추론해 저장하지 않는다). 파싱은 `scanTodoAddArgs` 의 알려진 긴 플래그 분기에 더하고, 닫힌 집합 검증(`origin`)과 존재 검증(`--parent` 의 카드 id)은 쓰기 전에 한다 — 틀리면 아무것도 쓰지 않는다(분류 JSON 파일의 선례). 닫힘 시각(D9)은 저장하지 않고 접근자 하나가 `archived_at`(보관)과 `dropped_at`(dropped)에서 읽으며 둘 다 없으면 "unknown" 이다. 과거 카드는 소급하지 않는다(보관 76/1,047, dropped 1/90 만 스탬프가 있다 — QB09).

### 4.4 이주 순서

(1) 열기 → DDL → 버전 조정 → 기존 가산 패스들 → **새 가산 패스**(마지막). (2) 기존 DB 를 새 엔진으로 열면 컬럼이 생기고 기존 행은 그대로(행 해시 동일). (3) 새 DB 를 옛 엔진으로 여는 경로는 지원하지 않는다(가산 컬럼 선례와 같다). (4) 순수 읽기는 DDL 을 돌리지 않고 컬럼이 없으면 NULL 로 읽는다.

## §5 관계 모델

### 5.1 노드

| 노드 | 식별 | 원천 |
|---|---|---|
| Card | `tN` | 큐(live·보관) |
| Component | 경로 앞 두 마디(`internal/cli`) | 예상 파일·본문 경로에서 파생, 저장하지 않음 |
| File | 저장소 상대 경로 | 예상 파일, 카드 귀속 병합의 변경 파일 |
| Commit | SHA | 카드 귀속 병합 커밋 |
| Spec | `SPEC-…` | `card:`·`depends_on`·카드의 `spec_id` |
| Finding | `finding:<subject>:<related>:<relation>:<source>` | 소견 한 건 |

### 5.2 종류와 읽기 시 매핑 (저장 행은 다시 쓰지 않는다)

| 어휘 종류 | 방향 | 읽는 곳 |
|---|---|---|
| `blocks` | 선행 → 후행 | 기존 `blocks`(그대로), `depends`(`WaitsOnOf` 가 방향 정규화), `gtd_relations.depends_on`(방향 반전) |
| `duplicates` | 대칭 | `near-duplicate`, `duplicate-forced`(점수·소스·강제 표지를 한정어로 보존) |
| `parent-of` | 부모 → 자식 | `issuance.spawned_by` 가 있고 `origin = split` 인 카드(투영), `gtd_relations.part_of` |
| `follow-up-of` | 자식 → 원점 | `issuance.spawned_by` 가 있고 `origin ∈ {follow-up, audit-finding, ci-repair}` 인 카드(투영) |
| `merged-into` | 흡수된 → 흡수한 | `todo merge` 가 기록하는 소견 |
| `supersedes` | 새 → 옛 | `replaces`, `gtd_relations.supersedes`·`replaces` |
| `relates-to` | 대칭 | `contains`, `absorbs`, `conflicts`(원래 이름을 한정어로), `gtd_relations.related_to`·`supported_by`·`contains`·`absorbs`·`conflicts` |

새로 기록하는 관계는 기존 `findings` 어휘를 확장해 쓴다(`duplicates`, `merged-into`, `supersedes`, `relates-to`; `parent-of`·`follow-up-of` 는 쓰지 않고 속성에서 투영 — D10). 오래된 행은 그대로 읽힌다(REQ-TCI-017).

### 5.3 종류별 제약

| 종류 | 자기 간선 | 순환 | 카디널리티 | 대칭 정규화 |
|---|---|---|---|---|
| `blocks` | 거절 | 거절(기존 `WaitsOnClosesCycle` 재사용) | n:m | 없음 |
| `duplicates` | 거절 | 해당 없음(무방향) | n:m | 쌍을 (작은 id, 큰 id) 로 정규화, 같은 쌍 중복 기록 없음 |
| `parent-of` | 거절 | 거절 | 자식당 부모 ≤ 1 | 없음 |
| `follow-up-of` | 거절 | 거절 | 카드당 원점 ≤ 1 | 없음 |
| `merged-into` | 거절 | 거절 | 흡수된 카드당 대상 ≤ 1, 대상 카드는 닫힌 카드 불가 | 없음 |
| `supersedes` | 거절 | 거절 | n:m | 없음 |
| `relates-to` | 거절 | 해당 없음 | n:m | 쌍 정규화 |

`parent-of`·`follow-up-of` 는 속성에서 투영하므로 "부모 ≤ 1"은 스칼라 필드로 자동 성립하고, 속성을 쓰는 시점(`add --parent`)에 순환(자기 조상을 부모로 삼기)을 거절한다. 거절은 어느 종류든 쓰기 동사가 관계를 쓰기 전에 하며 파일을 바이트 그대로 남긴다.

### 5.4 해석기

카드 id ↔ GTD 항목 id 를 `gtd_items.card_id`(engage 가 설정)로 양방향 매핑한다. 읽기만 하고 `gtd_relations` 에 쓰지 않으며 `gtd organize` 는 바뀌지 않는다. 한 번의 읽기가 카드의 소견, 속성 투영, 연결된 GTD 항목의 `gtd_relations` 를 한 목록으로 돌려주고 각 행은 `source`(`findings`·`field`·`gtd`)를 단다. 한 번도 engage 되지 않은 카드는 GTD 쪽이 비어 있다(스냅숏에서 `gtd_items` 0행, `gtd_relations` 0행이므로 오늘의 실질 내용은 소견과 속성뿐이다).

### 5.5 `moai todo trace`

`trace <id> [--kind <k>]… [--depth <n>]`: 해석기가 낸 관계 목록 위의 방문 집합 기반 너비 우선 탐색, 방문한 노드는 다시 들어가지 않으므로 레거시 데이터의 순환에서도 끝난다. 출력은 `(깊이, 종류, 시작 → 도착, 한정어, 소스)` 한 줄씩, 정렬은 (깊이, 종류, id)로 결정적이다. 쓰기 없음. `todoLaneReadOnlyVerbs` 에 등록한다.

### 5.6 소견 처분

처분은 소견의 선택 컬럼 `disposition`(`accept`·`merge`·`reject`)이고 설정은 명시 동사(작업 이름 `moai todo dispose <index> <값>`; 인덱스는 `todo why` 가 출력하는 것)로만 한다. 처분은 **기록만** 한다 — 카드·순서·픽업 필터·표시를 바꾸지 않는다(D11 작업 기본값). 오래된 소견은 처분이 없다.

## §6 카드→파일 간선의 운반체 (D3 분석)

사실. `.moai/project/graph/` 는 `.gitignore:344` 로 추적 제외이고 `edges.jsonl` 은 CI(`graph-freshness.yml`)에서 `moai graph build` 로 새로 만들어진다 — CI 에는 큐 DB 가 없다. 파일 머리 주석의 계약은 "같은 트리에서 두 번 돌리면 바이트 동일, 시각 없음"이다. MCP 그래프 도구는 `code-call` 간선만 읽는다. GTD 투영은 `~/.moai/db/<key>/todo/gtd-edges.jsonl` 에 0600 비공개로 쓴다.

작업 설계(D3 작업 기본값 = 커밋된 증거에서만): 카드 귀속 first-parent 병합 커밋(`subjectAttribution` 이 인정하는 제목 형태)과 그 병합이 가져온 변경 파일에서 `Edge{Kind: "card-file", Source: "t<N>", Target: <경로>}` 를 만든다. 큐를 읽지 않으므로 같은 트리 → 같은 출력이고, CI 와 로컬이 같은 간선을 만든다. 신선도는 `SourceFingerprintsForEdges` 에 "카드 귀속 병합 목록"의 지문을 더해 새 병합이 착지하면 stale 로 읽힌다(specs·reports 지문과 같은 성질). 열린 카드의 예상 파일(큐 비공개)은 `moai graph` 에 싣지 않고 M1·M4 가 큐에서 직접 읽는다.

한계: 열린 카드의 file→card 질의는 그래프로 못 한다. 병합 커밋이 두 카드를 함께 이름 붙인 경우는 귀속하지 않는다(`subjectAttribution` 의 의도). 제목만 보므로 카드 이름이 곧 인도의 증거는 아니다. 증거 추출 시간은 측정하지 못했고(공백 5) 증분 캐시(병합 SHA 키)가 후보다.

## §7 묶음과 병합

### 7.1 왜 지금 도구로는 안 되는가

`assign --after` 의 이전 카드 병합 가드는 **할당 간선**에서 돈다(`guardAssign` → `predecessorMerged`). 그래서 이전 카드가 병합되기 전에는 다음 카드를 레인에 할당할 수 없다. 직렬은 레인별이 아니라 전군 슬롯이다. 임대 선택 호 (a)(b)(b2)(c)는 첫 후보에서 `return` 하고, 후보가 거절되면 `factoryNextClaimRefused` 가 경쟁(stale·holder)이 아닌 것을 실제 오류로 돌려준다 — 선행 미병합 후보가 앞에 서면 호출이 오류로 끝날 수 있다는 읽기 추정이 있다(M4 가 관측한다).

### 7.2 설계 (D14 작업 기본값)

| 훅 | 변경 |
|---|---|
| `cards` 행 | 묶음 식별·순번·레인 속성(`ALTER TABLE … ADD COLUMN` 목록에 추가) |
| 적재 동사 | `moai factory bundle <lane> <card>…`: 각 카드의 레코드를 만들고(`RecordPicked`), 묶음 속성과 `HintAfter`(앞 멤버)를 채우고, 첫 멤버만 레인에 할당한다 |
| 선택 호 | (1) 묶음 멤버가 선행 미병합이면 오류 대신 건너뛴다, (2) 묶음 레인이 다르면 건너뛴다, (3) 레인은 자기 묶음의 다음 멤버를 소유자 없는 picked 카드보다 우선한다 |
| 허브 체인 | 레코드 생성 시점(`RecordPicked`)에 새 카드의 예상 파일이 허브 파일을 담고 같은 허브 파일을 담은 열린 카드가 있으면 `HintAfter` 를 그 카드로 채운다 |

직렬 슬롯 의미(전군 단일 슬롯)는 그대로다. keep-set 은 파일 겹침을 읽지 않는다 — 허브 체인은 레코드 생성의 입력이지 keep-set 이나 선택의 입력이 아니다. 임대 선택 호가 묶음 속성을 읽는 것은 REQ-TAU-013 의 "새 입력이 임대 경로를 고치지 않는다"와 맞지 않으므로 개정 행 A3 에 기록한다. 선택 호 변경은 비묶음 카드의 동작이 같음을 골든으로 증명한다.

### 7.3 병합 동사

`moai todo merge <into> <from>`: `<into>` 본문 끝에 `[merged from <from>] <from 본문>` 절을 덧붙이고, `<from>` 을 `merged into <into>` 사유로 drop 하고(접두사와 `drop_reason`), `merged-into` 소견(소스 `agent`, `relate` 와 같은 소스 집합 — 새 소스를 더하지 않아 `gtd.md` 의 소스 열거 문장과 그 테스트에 영향이 없다)을 기록한다. 거절: 둘 중 하나가 picked, `<from>` 이 이미 병합됨, 순환, 존재하지 않는 id. 레인 세션에서는 `todoRefuseLaneMutation` 이 막고(읽기 전용 목록에 없다), 분석기·`analyze`·`relate` 는 이 함수를 부르지 않는다는 것을 호출 그래프 테스트로 강제한다(기존 "코드 모양이 강제한다" 원칙의 새 형태).

## §8 M5 규칙 설계

배치(D1·D7·D15): 새 규칙은 `.claude/rules/moai/workflow/card-issuance.md`(경로 한정, `**/.claude/skills/moai/workflows/gtd.md,**/.claude/agents/moai/manager-todo.md`; `kanban-dispatch*` 글롭과 자기 매칭하지 않는 이름)에 둔다. 상시 로드 `kanban-dispatch.md` 는 순증가 0 의 한 줄 포인터만. 수치는 배포 사본에 쓰지 않고 기준선 기록 또는 로컬 전용 규칙을 가리킨다. 섹션:

- **카드 크기**: 하한 트리거(작고 파일이 적은 카드는 같은 주 파일의 열린 카드와 묶음 후보)와 상한(넘으면 분할 후보). 값은 기준선 기록의 분포에서.
- **후속 지적 규칙**: 카드가 만든 결함은 병합 전 카드 안에서 고친다. 이미 있던 결함은 구성요소별 부채 대장에 올린다(D12). `sync-auditor.md` 의 blocking/optional 분류와 `PASS-WITH-DEBT` 와 이어진다.
- **파생 깊이 상한**과 **동시 진행 한도**.
- **발행 체크리스트**: 발행 전 제시를 읽었는가(정확 일치·같은 구성요소·진행 중 겹침·완료 SPEC), 크기 하한·상한, 부모·출처 입력, 묶음 후보.
- **부채 대장**: 큐의 카드(`origin` 값 `follow-up` 과 구분되는 부채 표지), 구성요소 노드로 묶음 조회.

미러: 바이트 동일 쌍은 동일 편집, 분기된 쌍(`kanban-dispatch.md`, `sync-auditor.md`)은 분기 보존(§plan F.8). 템플릿 사본에는 카드 id·SPEC·REQ id·날짜를 쓰지 않는다.

## §9 웹 보기 설계

`/todo?view=graph`. 읽기 이음매는 `readTodoQueue` 옆에 두고(`LoadPure` 와 읽기 전용 GTD 투영 소스 `GTDProjectionSource` 를 쓴다) 보관 카드·관계·해석기 결과를 `GraphVM{Nodes, Edges, Omitted}` 로 돌려준다. 노드 상한을 넘으면 열린 카드와 연결된 노드를 우선하고 나머지는 "N개 생략" 문구로 표기한다. 레이아웃은 입력이 같으면 출력이 같은 결정적 배치(연결 요소별, 파생·선행 방향의 깊이를 행으로)이고 SVG 는 Templ 이 서버에서 그린다. 새 JS 는 만들지 않는다(데이터는 서버 렌더). 스타일은 임베드 목록에 이미 있는 `console.css` 에 더하므로 임베드 목록은 바뀌지 않거나, 새 자산 파일을 만들면 `assets.go` 의 `//go:embed` 목록에 넣는다. 새 문자열은 모든 로케일 키를 가진다(`i18n_governance_test.go`). GET 전용이고 락·쓰기가 없으며 외부 요청이 없다(SPEC-WEB-TODO-QUEUE-001 의 규칙 유지). 기존 표는 바뀌지 않는다(REQ-TCI-024).

## §10 가드·제약 목록 (변경이 지켜야 하는 것)

- 분류기·정규화·Jaccard·임계값과 그 테스트(`TestClassifyCardText`, `TestNormalizeCardText`, `TestTokenSetJaccard`), 정확 중복 거절·`--force`·near 기록·비재정렬·`analyze` 멱등 테스트.
- `add` stdout 한 줄과 대시 본문·폴스루·`--pick` 테스트, stale-store 공개, 동시 add 8 프로세스, `--pick` 한 번의 잠금 쓰기.
- MCP 패리티 `TestSD_AC014_MCPMatchesCLIWithProjectRoot`.
- 스키마 동결 테스트, `list --json` 골든, v1→v2 재구성 순서(카드 t1310).
- 보관·복원·큐 병합·이주 경로.
- `ValidateGTDRelation`·`WaitsOnOf`·`FindingsBlocking`·픽업 필터 테스트, `TestAutoRank*`.
- 직렬 슬롯·keep-set·임대 선택 테스트(`TestFactoryNext*`).
- `moai graph` 결정성·신선도 테스트, `graph-freshness` CI.
- 웹: `todo_route_test.go`·정렬·상세·live 테스트, i18n 거버넌스, `appjs_fire_*` 가드.
- 템플릿: `card_id_leak_test.go`, `workflow_rule_paths_pinned_test.go`, `contract_mode_guided_test.go`, `jev_auto_exception_test.go`, `todo_skill_doc_parity_test.go`·`todo_classify_doc_parity_test.go`, `TestDeclaredRuleMirrorForks`, 카탈로그 해시 두 테스트, `TestGoldenCommittedArtifactsMatchEmission`.
- 상시 로드 토큰 가드와 40,000자 훅.

## §11 결정 선택지와 근거 (D1~D15)

`recommendation_mode: pull` 이므로 `decision-index.md` 는 권고하지 않는다. 아래 "작업 기본값"은 계획이 진행하는 값일 뿐이며 리더가 감사 근거로 판정한다.

### D1 — M5 순서와 게이트 미충족 시 인도물

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) M5 를 마지막에 두고 게이트 명령이 판정 | t1453 병합 수 0(핀), t1448 대조군 1; t1453 의 plan M4 가 같은 규칙 파일을 절체 시점에 고친다(브랜치 끝 `2a5f9c91c`의 plan.md) | 게이트가 안 열리면 규칙이 효력 없이 SPEC 이 닫힐 수 있다 |
| (ii) 지금 편집, t1453 이 흡수 | 규칙 파일 여섯과 사본 | 두 카드가 같은 파일을 만든 충돌, 템플릿 미러 분기 위에 편집이 쌓임 |
| (iii) M5 를 후속 SPEC 으로 분리 | Tier L 의 상한과 요구 수(24/25)에 여유가 없다 | 요구 REQ-TCI-021/022 를 옮겨야 한다 |
| 미충족 인도물 (a) 초안만 내고 요구 문면대로 닫음 | REQ-TCI-022 | 규칙 미효력 상태가 완료로 읽힘 — 완료 보고에 명시 |
| 미충족 인도물 (b) SPEC 을 in-progress 로 붙듦 | 수명주기 | 카드가 열려 있음 |

작업 기본값: (i)과 (a).

### D2 — 두 관계 저장소 통합

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) 단일 저장소(소견을 `gtd_relations` 로 이주) | 식별 공간·저장 형태·보관 의미·쓰기 경로·어휘(8 대 9)·중복 제거 키가 모두 다르고, `ValidateGTDRelation` 은 대상이 `gtd_items` 에 있어야 한다("missing_target") | 보관·복원·큐 병합·이주 코드 전체, 저장 행 재작성 |
| (ii) 저장 분리 + 읽기 해석기 + 어휘 확장 | 저장 행 불변, `gtd_items.card_id` 가 연결 고리 | 두 저장소가 영구히 남음 |
| (iii) 읽기 시점 합집합만 | 새 어휘가 없음 | 새 관계(병합·부모)를 쓸 곳이 없다 |

근거 수치: `gtd_relations` 0행, `findings`+`archived_findings` 고유 125쌍. 작업 기본값: (ii).

### D3 — 카드→파일 간선 운반체

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) 비공개 `card-edges.jsonl` 사이드카 | GTD 투영 선례(`gtd_private.go`), 큐 상태에서 만든다 | 오늘 소비자가 없다(MCP 도구는 `code-call` 만 읽는다) |
| (ii) 커밋된 증거(카드 귀속 병합)에서만 `edges.jsonl` | 동일 트리 동일 출력, CI 와 로컬 일치, 비공개 상태 없음 | 열린 카드의 file→card 질의 불가, git 증거 추출 시간 미측정 |
| (iii) 둘 다 | | 두 운반체의 유지 |

사실: `.gitignore:344` 가 산출물을 추적 제외한다. 작업 기본값: (ii).

### D4 — M2 스키마 형태

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) JSON 컬럼 하나 | 분류 선례(REQ-TCD-002), 닿는 곳(두 표의 `ensureColumn`, INSERT·SELECT, parity, 동결 튜플 문자열 둘, 투영)을 한 번씩만 바꿈 | SQL 조회는 `json_extract`, 컬럼별 NOT NULL·CHECK 불가 |
| (ii) 속성별 컬럼(5개) | 임대·스탬프 선례 | 닿는 곳이 속성 수만큼 곱해짐 |

큐는 `LoadPure` 로 통째로 메모리에 올려 Go 에서 읽고 보관 1,047행에 SQL 속성 조회가 없다. 작업 기본값: (i).

### D5 — MCP 전달과 engage

| 선택지 | 근거 | 비용 |
|---|---|---|
| MCP (i) 결과 텍스트에 덧붙임 | `sdCallTool` 이 모든 TextContent 를 이어 붙여 CLI stdout 과 비교하는데, 고정 입력에서 제시가 비어 기존 테스트가 통과한다 | 제시가 있을 때 결과 텍스트 ≠ CLI stdout(문서화된 차이) |
| MCP (ii) `StructuredContent` | 텍스트 계약 불변 | Claude Code 가 어떻게 보이는지 미측정 — 안 보이면 알림이 사라진다 |
| MCP (iii) 별도 미리보기 도구 | `todo_add` 불변 | 호출자가 먼저 불러야 보인다 |
| engage (a) 제시만(분석기 비합류) | engage 는 분석·분류를 거치지 않는다; REQ-TA-010 문면 불변 | engage 는 소견을 기록하지 않음 |
| engage (b) 분석기 합류 | 정확 중복 거절·소견 | REQ-TA-010 개정(A1), engage 가 거절할 수 있음 |
| engage (c) 변경 없음 | | engage 만 제시가 없다 |

작업 기본값: MCP (i), engage (a).

### D6 — 병합 독트린 개정 범위

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) 운영자 호출 동사 하나, 나머지는 계속 금지 | ANALYSIS §B.1(눈에 안 보이는 변경이 문제), 흡수는 이미 `drop`+`edit` 의 운영자 조합으로 가능 | 문면 개정(A2), 레인 가드 목록에 넣지 않기 |
| (ii) 리더·레인에도 허용 | | "큐는 운영자가 쓴다" 원칙과 충돌 |
| (iii) 동사 없이 관계 기록만 개선 | | 카드가 요구한 경로가 없음 |

작업 기본값: (i).

### D7 — 수치 임계값

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) SPEC 에 값 고정 | | 큐가 변해 측정이 낡고, 카드의 "저장소 실측" 요구와 어긋남 |
| (ii) 값은 M0 기준선 기록에서 읽음, SPEC 은 기록과 측정 명령만 가리킴 | GB03·GB04·GB09·QB06·SB03·SB07 등 | M0 가 값 도출의 판단(어느 분위수)을 맡음 |
| (iii) 배포 규칙은 메커니즘만, 값은 로컬 전용 규칙 | 배포 사본에 이 저장소의 수치가 새지 않음 | 규칙 파일이 하나 더 |

작업 기본값: (ii).

### D8 — 유사도 척도와 소음

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) 기존 `TokenSetJaccard` + 표시 하한 | 재사용; 하한 0.3 에서 알림 8.4%(최근 4.9%) | 기록된 관계 재현율 6% |
| (ii) idf 가중 척도 신설 | 하한 없음 재현율 65%(대 52%), 하한 0.3 에서 알림 5.7% | 새 척도와 문서 빈도 계산 |
| (iii) 하한 없이 항상 상위 3개 | 재현율 52~65% | 알림이 91% 이상에서 점수 0.1 안팎의 소음 |
| (iv) 확실한 것은 항상, 어휘 이웃은 `--dry-run` 에서만 | | 일반 `add` 에서 어휘 이웃이 안 보임 |

작업 기본값: (i) 하한과 (iv)의 `--dry-run` 전용 하한 미만 표시를 합친 형태.

### D9 — closed-at

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) 저장하지 않고 접근자 | `archived_at` 76/1,047, `dropped_at` 1/90 — 컬럼을 더해도 과거는 NULL | 접근자의 "unknown" 폴백 |
| (ii) 컬럼 추가 | 전이 시점에 한 자리에서 쓴다 | 기존 두 스탬프와 중복 |
| (iii) 과거는 git 착지 커밋 시각으로 소급 | 착지 964장은 git 에 있다 | 소급 쓰기가 범위 밖 항목과 충돌 |

작업 기본값: (i).

### D10 — 파생 부모의 단일 원천

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) 카드 속성이 원천, 관계는 투영 | 스칼라라 부모 ≤ 1 이 자동 | 관계 저장소에 `parent-of` 가 안 쌓임 |
| (ii) 관계가 원천 | | 카디널리티·순환을 매번 검사 |
| (iii) 둘 다 쓰고 검사 | | 불일치 가능 |

작업 기본값: (i).

### D11 — finding 처분의 효과

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) 기록만 | REQ-RPF-001 이 `blocks`/`depends` 소견의 존재를 미해결 술어로 정함 | 처분이 표시를 안 바꿈 |
| (ii) 처분 소견을 `list` 와 `machine-only` 표지에서 제외 | | 표시 변경, 골든 영향 |
| (iii) 처분이 소견을 보관으로 이동 | | 소견 생애 주기를 바꿈 |

작업 기본값: (i).

### D12 — 부채 대장 운반체

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) 큐의 카드(부채 표지)와 구성요소 노드 | 새 저장소 없음; 큐가 이미 카드·관계·그래프를 운반 | 큐는 비공개·머신 로컬 |
| (ii) 추적되는 구성요소별 문서 | 검토 가능 | 새 규약·쓰기 주체 |
| (iii) 감사 판정서의 표 | | 흩어짐 |

작업 기본값: (i).

### D13 — `add` 새 플래그의 입력

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) 명시 플래그만 | 열린 카드의 본문 경로 32%, 파생 머리말 추론이 큐에 없는 부모 54건을 가리킴; ANALYSIS §B.1 | 입력하지 않으면 필드가 빈다 |
| (ii) 본문에서 추론해 저장 | | 오추론이 눈에 안 보이는 변경으로 남음 |
| (iii) 둘 다 | | |

작업 기본값: (i); 추론은 제시의 읽기 전용 제안으로만.

### D14 — 묶음 메커니즘과 임대 경로 개정

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) 적재 동사만, 다음 카드 할당은 이전 병합 뒤 사람이 | 임대 경로 불변 | 자동 연속이 없다 — 카드의 "정식 경로"에 못 미침 |
| (ii) 팩토리 `cards` 에 묶음 컬럼 + 선택 호가 묶음 예약·선행 미병합 건너뛰기 | `guardAssign` 이 할당 간선에서 이전 병합을 요구해 미리 할당 불가; 선택 호는 첫 후보에서 반환 | 임대 경로 개정(A3), `factory_card.go`(1,800줄대)·homestate 스키마 |
| (iii) 큐 쪽 발행 속성에 묶음을 싣고 선택 호가 읽음 | 큐는 이미 선택 호가 읽음 | 큐 스키마에 팩토리 개념 유입 |

작업 기본값: (ii).

### D15 — 신규 경로 한정 파일의 40,000자 한도

이미 정해짐: SPEC-INSTRUCTION-BUDGET-SCOPE-001 § 1 Context 와 HISTORY 0.1.0. 이 SPEC 은 따른다.
