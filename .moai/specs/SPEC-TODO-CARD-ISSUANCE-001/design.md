# SPEC-TODO-CARD-ISSUANCE-001 — 설계

이 문서는 구조와 형식을 정한다. 코드는 쓰지 않는다. 이름은 **작업 이름**이며 실행 단계가 저장소의 명명 관례에 맞춰 바꿀 수 있다(수용 기준이 문자 그대로 요구하는 시험 이름만 고정이다). 근거 측정은 `research.md` §3, 선택지 표는 §11, 기준선 운반체는 §12, M5 게이트 술어는 §13. 요구(`spec.md`)는 관측 가능한 행동만 말하고 저장 경로·순회 방식 같은 구현 방법은 이 문서가 소유한다.

## §1 구조 개요

| 층 | 패키지 | 이 SPEC 이 더하거나 바꾸는 것 |
|---|---|---|
| 순수 조회·모형 | `internal/kanban` | 읽기 전용 이웃·구성요소·완료 SPEC 조회, 발행 속성·처분 타입, 관계 어휘 정규화·제약·카드↔GTD 해석기 |
| 저장 | `internal/kanban`(SQLite 엔진) | `items`·`archived_items`·`findings`·`archived_findings` 의 가산 컬럼, 읽기/쓰기/parity |
| 명령 | `internal/cli` | `add` 의 제시와 `--dry-run`·발행 플래그, `todo trace`, `todo merge`, 처분 동사, `gtd engage` 제시, MCP `todo_add` |
| 그래프 | `internal/graph` | 카드 귀속 병합의 변경 파일 간선 층, 출처 지문 |
| 팩토리 | `internal/homestate`, `internal/cli` | `cards` 행의 묶음 속성, 적재 동사, 선택 호의 묶음 예약, 임베드 허브 파일 목록 |
| 웹 | `internal/web` | 관계 그래프 읽기 이음매와 보기 |
| 규칙 | `.claude/rules`, `.claude/skills`, `.claude/agents`, 템플릿 사본 | M5 |
| 기준선 | `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/` | M0 — 추적되는 기록과 재현 스크립트, 허브 목록(§12) |

의존 방향은 한쪽이다 — `kanban`(순수) ← `cli`/`web`/`graph`/`homestate`. 새 조회 함수는 입력으로 `BacklogRecord` 스냅숏을 받고 쓰지 않는다. 쓰기는 기존 `Mutate` 경로를 그대로 쓴다.

## §2 `add` 경로의 데이터 흐름

**현재**: 인자 파싱(`scanTodoAddArgs`) → 쓰기 전 검증(분류 JSON 파일) → `runTodoAddAppendRoot`: stale-store 공개(stderr) → `Mutate{ appendAnalyzedCard(분류·정확 중복 거절·near 소견·Jev) ; 분류 적용 ; 재정렬 }` → stdout `"<id> <position>\n"`.

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
| 유사 카드 상위 3 | live(dropped 포함, 접두사 제거) + 보관 카드 | `measure=token-set-jaccard`, 점수, 상태(`live`·`dropped`·`archived` 중 하나), 본문 앞 60자, dropped 이면 사유 |
| 같은 구성요소의 열린 카드 | 열린 카드(queued·picked·hold)의 본문 경로와 예상 파일 | `measure=component`, 구성요소 키, 카드 id·상태 |
| 진행 중 레인과의 예상 파일 겹침 | 신규 카드의 예상 파일 × 진행 중 레인 카드(picked, 레인 배정 있음)의 예상 파일 — 아래 "예상 파일" 정의 | `measure=file-overlap`; 공유 경로마다 한 항목(진행 중 카드 id·레인·경로). 입력이 있는 비교에서 겹침이 없으면 `none`(줄을 쓰지 않는다), 후보에 입력이 없거나 비교할 진행 중 카드에 입력이 하나도 없으면 `unmeasured` 와 이유. 판정은 진행 중 카드 집합 단위라 입력이 없는 카드는 비교에서 빠지고 보고되지 않으며, 경로 일치는 정규화한 경로 전체의 일치다(파일 이름·문자열 접두사가 아니다) |

**예상 파일(겹침 항목의 입력).** 카드의 예상 파일은 (i) 본문이 이름으로 대는 저장소 경로 토큰(읽기 전용 추출 — 저장하지 않는다, D13)과 (ii) M2 이후 `issuance.files` 의 합집합이다. 진행 중 레인 카드는 여기에 (iii) 그 레인 브랜치의 변경 파일(`git diff --name-only`, 시간 상한 안에서만; §3.4)을 더한다. 오늘은 `files` 속성이 어느 카드에도 없고 본문 경로가 열린 카드 37장 중 12장에만 있으므로(SB06) 대부분 `unmeasured` 로 나가며, 양성 겹침은 고정 입력 시험으로 관측한다(`acceptance.md` AC-TCI-002 (e), AC-TCI-003 (f)(g)). 허브 체인(§7.2)은 (ii) 만 읽는다 — 본문 추출은 재현율이 낮고(SB06 32%) 저장되지 않는 추론이라 체인을 만들지 않는다.
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

git 탐침(진행 중 레인 브랜치와 병합 기준 사이의 `git diff --name-only`)은 레인 브랜치 하나에 약 0.25초이고(SB08) `WT-` 브랜치가 353개(이터레이션 2 재측정 356), 워크트리가 74개(재측정 79)다. 진행 중 레인 수만큼 직렬로 돌면 `add` 가 눈에 띄게 느려진다. 그래서 (a) 탐침은 `Mutate` 밖에서만, (b) 항목당·전체에 시간 상한, (c) 예상 파일이 있는 카드는 git 을 부르지 않고 필드끼리 비교, (d) 상한 초과는 `unmeasured (time bound)` 로 표기한다. 상한값은 M1 이 진행 중 레인 수와 탐침 지연을 먼저 재서 정한다(`plan.md` §F.11). 탐침은 락을 쥔 채 돌지 않는다는 것이 시험 가능한 성질이다(탐침 대기 중에 다른 프로세스의 `Mutate` 가 끝나야 한다).

### 3.5 분류기와의 분리

제시용 조회는 `ClassifyCardText` 를 부르지 않고 `kanban` 에 새로 둔 읽기 전용 함수다. 분류기는 dropped 카드를 거절 대상에서 제외한다는 독트린(`backlog_analysis.go` 의 함수 주석)과 고정 테스트를 가지고 있어 건드리지 않는다. 제시는 dropped 카드도 **알림 대상**으로 포함하되 카드를 거절하거나 소견을 기록하지 않는다. 카드 본문이 요구했던 `ClassifyCardText` 와 `todoTriageSymbols` 의 재사용은 이 설계가 **의도적으로 하지 않는다**(`spec.md` §A.3 정정 7): 전자는 위 독트린 때문에, 후자는 기호를 최대 4개만 내는 상한 때문에 경로 겹침 용도에 맞지 않기 때문이다. 척도(`NormalizeCardText`·`TokenSetJaccard`)만 재사용한다.

## §4 스키마와 이주

### 4.1 형태 (D4 작업 기본값)

| 표 | 추가 | 형태 |
|---|---|---|
| `items`, `archived_items` | `issuance` | nullable TEXT, JSON 객체. 없음 = NULL |
| `findings`, `archived_findings` | `disposition` | nullable TEXT(`accept`/`merge`/`reject`). 없음 = NULL |

`issuance` JSON 의 키(모두 선택): `spawned_by`(카드 id), `origin`(닫힌 집합: `operator`, `leader`, `audit-finding`, `follow-up`, `ci-repair`, `standing`, `external`, `split`), `size_lines`(예상 제품 줄 수, 정수), `files`(예상 파일 경로 배열), `drop_reason`(문자열). 카드가 어느 키도 갖지 않으면 컬럼은 NULL 이고 직렬화에서 사라진다(`omitempty`). 닫힌 `origin` 집합은 카드 머리말 분류(`research.md` §3.2 의 02 스크립트 분류)를 겹치지 않는 값으로 줄인 것이다.

REQ-TCI-007 은 이 형태를 행동으로만 말한다(없음은 빈 값이 아니라 없음, 속성이 없는 카드는 이전과 같게 저장·직렬화·내보내짐). "가산 컬럼 경로로 추가하고 v1→v2 재구성 목록에는 넣지 않는다", "없음은 SQL NULL 과 nil" 같은 구현 방법은 이 절과 §4.2·§4.4 가 소유한다.

### 4.2 닿는 곳 (모두 한 마일스톤 M2 의 한 커밋 묶음)

카드 표 둘과 finding 표 둘은 **같은 규율**을 받는다 — 이주, 순수 읽기, 보관 왕복, parity, 동결 튜플, 큐 병합. 아래 줄 번호는 이터레이션 2 가 `1894984c3` 에서 `git grep` 으로 읽은 위치다(이터레이션 3 의 핀 `ad02a5677` 과 제품 경로가 같다 — `acceptance.md` G20 이 두 커밋 사이 `internal`·`cmd`·`.claude` 의 변화 없음을 읽었다).

| 닿는 곳 | 위치 | 이유 |
|---|---|---|
| 타입 | `backlog_store.go` 의 `BacklogItem`·`BacklogFinding`(:208)·`BacklogArchivedFinding`(:310)·`BacklogArchiveEntry`(:328) | 필드 추가, 포인터와 `omitempty`. **보관 finding 은 별도 타입**이라 `BacklogFinding` 만 고치면 처분이 보관 때 사라진다 |
| 가산 컬럼 | `backlog_sqlite.go` 의 `ensureSchema` 마지막 가산 패스, `ensureColumn` | 버전 조정 뒤에 retrofit(카드 t1310 순서). 카드 표 둘은 `issuance`, finding 표 둘은 `disposition`. `backlogItemsTableColumns` 에는 넣지 않음. `archived_findings` 의 DDL 은 `:177` |
| 쓰기 | `backlog_migrate.go` 의 `INSERT INTO items(...)`(:586)·`INSERT INTO archived_items(...)`(:469), `INSERT INTO findings(...)`(:595)·`INSERT INTO archived_findings(...)`(:477) | 열 목록 갱신 — 네 곳 모두 |
| 읽기 | `readSnapshot`(live 카드 :86, live finding :200)·`readArchive`(보관 카드 :243, 보관 finding :379)와 `columnExpr`(:75) | 없는 컬럼은 NULL 로 읽어 DDL 없이 순수 읽기 유지. finding 읽기는 오늘 `columnExpr` 를 쓰지 않는 고정 SELECT 목록이므로 새 컬럼을 `columnExpr` 로 읽게 바꿔야 한다 |
| parity | `assertBacklogParity`(:904) | 새 컬럼 두 개를 네 표 모두에서 포함 |
| 동결 튜플 | `backlog_schema_freeze_test.go` 의 `wantItemsColumns`·`wantArchivedItemsColumns` | 카드 표 둘은 마지막 튜플로 `issuance:TEXT:0:NULL` 추가, 같은 커밋. **finding 표 둘은 오늘 열 튜플이 동결돼 있지 않다**(시험은 `wantTables` 로 표 이름만 고정한다) — 이 변경이 `findings`·`archived_findings` 의 마지막 튜플 `disposition:TEXT:0:NULL` 고정을 새로 더한다 |
| 보관·복원 | `ArchiveCard`·`RestoreCard` | 항목 통째 복사라 카드 필드는 따라가고, finding 은 카드와 함께 `BacklogArchivedFinding` 으로 이동하므로 처분이 따라가는지 시험 |
| 큐 병합·이주 | `todo_queue_merge.go`(`rewriteArchivedFindings` :359 외), `todo_merge_procedure.go`(:333-376) | 항목·소견 복사·재매핑 경로가 두 컬럼을 나르는지(`TestQueueMergeCarriesIssuanceAndDisposition`) |
| JSON | `todoJSONProjection`(`todo_claim.go`), `omitempty` | 새 필드가 없는 카드는 골든과 바이트 동일 |
| export | `todo_export.go` | 새 필드를 어떻게 다루는지(공백) |
| drop | `todo_drop.go` | 사유 속성 저장, 텍스트 접두사 유지 |

### 4.3 `add` 의 새 플래그와 닫힘 시각

`add` 서브커맨드(폴스루 아님)에 `--parent <id>`, `--origin <값>`, `--size-lines <n>`, `--files <경로,…>` 를 둔다(D13: 명시 입력만, 본문에서 추론해 저장하지 않는다). 파싱은 `scanTodoAddArgs` 의 알려진 긴 플래그 분기에 더한다. 검증은 모두 쓰기 전에 하고, 틀리면 아무것도 쓰지 않으며 id 도 소비하지 않는다(분류 JSON 파일의 선례): (a) `--parent`·`--origin` 은 한 번만 받는다(둘째는 플래그 이름을 대는 메시지로 거절 — 조용한 덮어쓰기 금지), (b) `--origin` 은 닫힌 집합 안, (c) `--parent` 는 큐에 **존재하는** 카드 id 여야 한다 — live·dropped·**보관 카드 모두 존재로 본다**(후속 카드는 닫힌 카드에서 나오는 것이 정상 흐름이다). 닫힘 시각(D9)은 저장하지 않고 접근자 하나가 `archived_at`(보관)과 `dropped_at`(dropped)에서 읽으며 둘 다 없으면 "unknown" 이다. 과거 카드는 소급하지 않는다(보관 76/1,047, dropped 1/90 만 스탬프가 있다 — QB09).

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

**쓰는 길.** `todo relate` 는 오늘 여섯 이름(`contains`·`absorbs`·`replaces`·`conflicts`·`blocks`·`depends`)을 받는다. 이 SPEC 은 새 쓰기 이름 `duplicates`·`supersedes`·`relates-to` 를 더해 기존 `findings` 어휘를 확장한다(오래된 행은 그대로 읽힌다, REQ-TCI-017). `merged-into` 는 `todo merge`(M4)만 쓰고 `relate` 는 쓰지 않는다 — 접는 동사가 접는 기록의 유일한 작성자여야 "분석은 접지 않는다"는 불변식이 유지된다. `parent-of`·`follow-up-of` 는 쓰지 않고 속성에서 투영한다(D10). 그래서 `relate` 의 쓰기 가능 집합은 오늘의 여섯 이름(`BacklogSemanticRelations`)에 `duplicates`·`supersedes`·`relates-to` 셋만 더한 아홉이고, 일곱 종류 어휘 전체가 쓰기 가능 집합이 되어서는 안 된다 — `relate` 에 준 `parent-of`·`follow-up-of`·`merged-into` 는 오늘처럼 `--relation must be one of … (got "<값>")` 로 거절돼야 하고(`internal/cli/todo_relate.go:63-65`), 이 불변식을 `TestTodoRelateRefusesProjectionKinds` 가 붙든다(`acceptance.md` AC-TCI-013 (f)).

### 5.3 종류별 제약

감사 지적 D5 에 따라 표는 **쓰는 동사**를 열로 둔다 — 어떤 동사로도 만들 수 없는 위반은 요구·기준에 넣지 않고 이유를 적는다.

| 종류 | 쓰는 동사 | 자기 간선 | 순환 | 카디널리티 | 대칭 정규화 |
|---|---|---|---|---|---|
| `blocks` | `relate`(`blocks`·`depends`) | 거절 | 거절(기존 `WaitsOnClosesCycle` 재사용) | n:m | 없음 |
| `duplicates` | `relate`(새 이름) | 거절 | 해당 없음(무방향) | n:m | 쌍을 (작은 id, 큰 id) 로 정규화, 같은 쌍 중복 기록 없음 |
| `supersedes` | `relate`(새 이름; 옛 `replaces` 는 읽기 시 `supersedes` 로 센다) | 거절 | 거절 | n:m | 없음 |
| `relates-to` | `relate`(새 이름) | 거절 | 해당 없음 | n:m | 쌍 정규화 |
| `merged-into` | `todo merge` 만 | 거절 | 닫히거나 이미 병합된 대상은 거절 — 순환이 만들어질 길이 없다 | 흡수된 카드당 대상 ≤ 1(이미 병합된 `<from>` 거절) | 없음 |
| `parent-of` | 없음(add 속성의 투영) | 해당 없음 | 구조적으로 도달 불가 | 자식당 부모 ≤ 1 은 스칼라와 `--parent` 중복 거절 | — |
| `follow-up-of` | 없음(add 속성의 투영) | 해당 없음 | 구조적으로 도달 불가 | 카드당 원점 ≤ 1 은 스칼라와 `--origin` 중복 거절 | — |

`parent-of`·`follow-up-of` 는 속성에서 투영하므로 **관계 동사에는 이 종류의 순환이나 둘째 부모를 만들 길이 없다**. 순환이 구조적으로 도달할 수 없는 이유: 속성은 `add` 시점에 새 카드에만 설정되고, 새 카드는 후손이 없으며, `--parent` 는 이미 존재하는(따라서 더 작은 id 의) 카드여야 한다. 그래서 REQ-TCI-013 과 AC-TCI-013 은 이 위반들을 시험하지 않고, 대신 실제로 위반 입력을 받는 쓰기 동사의 검사를 요구한다 — `relate` 의 자기 간선·`blocks` 순환·`supersedes` 순환, `add` 의 둘째 `--parent`/`--origin`·존재하지 않는 `--parent`·닫힌 집합 밖 `--origin`. 거절은 어느 동사든 관계나 속성을 쓰기 전에 하며 파일을 바이트 그대로 남긴다.

### 5.4 해석기

카드 id ↔ GTD 항목 id 를 `gtd_items.card_id`(engage 가 설정)로 양방향 매핑한다. 읽기만 하고 `gtd_relations` 에 쓰지 않으며 `gtd organize` 는 바뀌지 않는다. 한 번의 읽기가 카드의 소견, 속성 투영, 연결된 GTD 항목의 `gtd_relations` 를 한 목록으로 돌려주고 각 행은 `source`(`findings`·`field`·`gtd`)를 단다. 한 번도 engage 되지 않은 카드는 GTD 쪽이 비어 있다(스냅숏에서 `gtd_items` 0행, `gtd_relations` 0행이므로 오늘의 실질 내용은 소견과 속성뿐이다).

### 5.5 `moai todo trace`

`trace <id> [--kind <k>]… [--depth <n>]`: 해석기가 낸 관계 목록 위의 방문 집합 기반 너비 우선 탐색, 방문한 노드는 다시 들어가지 않으므로 레거시 데이터의 순환에서도 끝난다. 출력은 `(깊이, 종류, 시작 → 도착, 한정어, 소스)` 한 줄씩, 정렬은 (깊이, 종류, id)로 결정적이다. 쓰기 없음. `todoLaneReadOnlyVerbs` 에 등록한다.

### 5.6 소견 처분

처분은 소견의 선택 컬럼 `disposition`(`accept`·`merge`·`reject`)이고 설정은 명시 동사(작업 이름 `moai todo dispose <index> <값>`; 인덱스는 `todo why` 가 출력하는 것)로만 한다. 처분은 **기록만** 한다 — 카드·순서·픽업 필터·표시를 바꾸지 않는다(D11 작업 기본값). 오래된 소견은 처분이 없다. 처분 컬럼은 `findings` 와 `archived_findings` 두 표에 같은 규율로 있으므로(§4.2) 카드가 보관돼도 처분은 소견과 함께 간다.

## §6 카드→파일 간선의 운반체 (D3 분석)

사실. `.moai/project/graph/` 는 `.gitignore:344` 로 추적 제외이고 `edges.jsonl` 은 CI(`graph-freshness.yml`)에서 `moai graph build` 로 새로 만들어진다 — CI 에는 큐 DB 가 없다. 파일 머리 주석의 계약은 "같은 트리에서 두 번 돌리면 바이트 동일, 시각 없음"이다. MCP 그래프 도구는 `code-call` 간선만 읽는다. GTD 투영은 `~/.moai/db/<key>/todo/gtd-edges.jsonl` 에 0600 비공개로 쓴다.

작업 설계(D3 작업 기본값 = 커밋된 증거에서만): 카드 귀속 병합 커밋과 그 병합이 가져온 변경 파일에서 `Edge{Kind: "card-file", Source: "t<N>", Target: <경로>}` 를 만든다. 큐를 읽지 않으므로 큐의 유무가 출력을 바꾸지 않는다. 신선도는 `SourceFingerprintsForEdges` 에 "카드 귀속 병합 목록"의 지문을 더해 새 병합이 착지하면 stale 로 읽힌다(specs·reports 지문과 같은 성질). 열린 카드의 예상 파일(큐 비공개)은 `moai graph` 에 싣지 않고 M1·M4 가 큐에서 직접 읽는다.

**순회와 귀속 (감사 D2·D13 의 같은 맹점).** 이터레이션 1 설계는 "first-parent 병합"만 읽었다. 카드 브랜치가 develop 을 `git merge develop` 으로 흡수하면 develop 의 카드 병합은 흡수 병합의 두 번째 부모 쪽에만 있어서 첫 부모 경로만 걷는 층은 그것을 놓친다(§13 이 같은 현상을 고정 SHA 로 보인다). 그래서 층은 **HEAD 에서 어느 부모 경로로든 닿는 병합**을 본다. 이때 흡수 병합이 카드의 기여로 잘못 셈하지 않도록 귀속은 새 정규식을 만들지 않고 기존의 단일 귀속 지점 `subjectAttribution`(`prlink_landed.go:188`)을 재사용한다 — 그 함수는 이미 "통합 대상이 해석된 통합 브랜치가 아닌 병합 제목은 어떤 카드도 귀속하지 않는다", "한 그룹이 두 개 이상의 서로 다른 카드 토큰을 담으면 아무것도 귀속하지 않는다", "착지하지 않았다는 표지는 아무것도 귀속하지 않는다"는 규칙을 가진다. 변경 파일은 각 병합 커밋의 **첫 부모 대비** diff 다 — develop 의 카드 병합은 첫 부모가 병합 직전의 develop 이므로 diff 가 카드의 기여이고, 그 병합이 어느 경로로 닿았는지와 무관하다. 어떤 제목 형태가 어떤 귀속을 내는지는 이 설계가 읽지 못했다(`subjectAttribution` 의 제목 형태 열거를 이 반복이 모두 따라가지는 않았다) — M3 의 첫 RED 시험이 흡수 병합의 두 번째 부모에서만 닿는 카드 병합과 흡수 방향 제목을 한 고정 저장소에 두고 실제 귀속을 관측해 순회 선택을 정한다.

**불변식의 범위 (감사 D13).** 출력은 커밋된 트리와 **HEAD 에서 도달 가능한 이력의 함수**다. 카드 귀속 병합은 트리가 아니라 이력에서 나오므로 "같은 트리면 같은 출력"은 이력이 같을 때만 참이다. 얕은 클론에서는 이력이 일부라 같은 입력이 아니다. `graph-freshness.yml` 은 `fetch-depth: 0` 으로 체크아웃하지만 `pull_request` 실행은 합성 병합 참조를 체크아웃하고 그 이력은 카드 브랜치와 다르다 — **로컬과 CI 가 같은 출력을 내는지는 측정하지 못했다**(미검증, `spec.md` §G). 이 설계는 그 일치를 약속하지 않는다.

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
| 허브 체인 | 레코드 생성 시점(`RecordPicked`)에 새 카드의 `files` 속성(명시 입력)이 **임베드된 허브 목록**(§7.4)의 경로를 담고 같은 허브 파일을 `files` 속성에 담은 열린 카드가 있으면 `HintAfter` 를 그 카드로 채운다 — 본문 경로 추출은 읽지 않는다(§3.1) |

직렬 슬롯 의미(전군 단일 슬롯)는 그대로다. keep-set 은 파일 겹침을 읽지 않는다 — 허브 체인은 레코드 생성의 입력이지 keep-set 이나 선택의 입력이 아니다. 임대 선택 호가 묶음 속성을 읽는 것은 REQ-TAU-013 의 "새 입력이 임대 경로를 고치지 않는다"와 맞지 않으므로 개정 행 A3 에 기록한다. 선택 호 변경은 비묶음 카드의 동작이 같음을 골든으로 증명한다.

### 7.3 병합 동사

`moai todo merge <into> <from>`: `<into>` 본문 끝에 `[merged from <from>] <from 본문>` 절을 덧붙이고, `<from>` 을 `merged into <into>` 사유로 drop 하고(접두사와 `drop_reason`), `merged-into` 소견(소스 `agent`, `relate` 와 같은 소스 집합 — 새 소스를 더하지 않아 `gtd.md` 의 소스 열거 문장과 그 테스트에 영향이 없다)을 기록한다. 거절: 둘 중 하나가 picked, `<from>` 이 이미 병합됨(둘째 `merged-into` 대상을 만들지 않는다), `<into>` 가 닫히거나 이미 병합됨(순환 불가), 순환, 존재하지 않는 id. 레인 세션에서는 `todoRefuseLaneMutation` 이 막고(읽기 전용 목록에 없다), 분석기·`analyze`·`relate` 는 이 함수를 부르지 않는다는 것을 호출 그래프 테스트로 강제한다(기존 "코드 모양이 강제한다" 원칙의 새 형태).

### 7.4 허브 목록의 출하 형태 (D1 대응)

허브 체인은 출하되는 제품 동작이므로 그 입력은 사용자 프로젝트에 없는 SPEC·reports 경로를 읽을 수 없다. 입력은 `internal/homestate/hub_files.txt`(한 줄에 경로 하나, `#` 주석 허용)를 `go:embed` 로 묶은 데이터 파일이고, M0 가 만드는 추적되는 `baseline/hub-files.txt` 의 경로 열이다. 두 사본을 시험이 대조하고 목록의 각 경로를 기준선이 기록한 단일 호출 명령으로 다시 잰다(§12). 목록의 경로가 존재하지 않는 프로젝트에서는 체인이 만들어지지 않는다. 그러나 목록의 경로가 사용자 프로젝트에도 있는지는 **검증하지 못했다** — 후보에는 흔한 Go 배치인 `internal/config/defaults.go` 가 들어 있고(`git ls-files -- internal/config/defaults.go` 가 이 저장소의 경로를 출력한다), 이 저장소는 다른 프로젝트를 볼 수 없다. 그런 프로젝트에서 두 열린 카드가 `--files` 로 같은 경로를 명시하면 이 저장소의 측정에서 나온 직렬화가 적용된다 — 영향은 두 카드를 한 줄로 세우는 순서 지정이고 거절이 아니며, 입력은 명시 `files` 속성뿐이다(본문 추출은 체인을 만들지 않는다, §3.1). 수용된 잔여이고 사용자별 허브 설정은 범위 밖이다(`spec.md` §G).

## §8 M5 규칙 설계

**한 가지 읽기(감사 D4).** 수치는 **로컬 전용 규칙에만** 쓰고 배포 사본은 **메커니즘만** 쓴다. 이 문장 하나를 `spec.md` REQ-TCI-021, `plan.md` §F.8, `acceptance.md` AC-TCI-020·021 이 같은 뜻으로 인용한다. 이터레이션 1 은 REQ 에서는 "규칙 사본에 수치가 있다"고, plan·design 에서는 "수치는 사본에 쓰지 않고 기준선 기록을 가리킨다"고 서로 어긋나게 적었다.

배치(D1·D7·D15): 새 규칙은 **두 파일**이다.

- `.claude/rules/moai/workflow/card-issuance.md` — 배포 규칙. 경로 한정(`paths:` = `**/.claude/skills/moai/workflows/gtd.md,**/.claude/agents/moai/manager-todo.md`; `kanban-dispatch*` 글롭과 자기 매칭하지 않는 이름), 템플릿 미러 있음, 40,000자 이하. **메커니즘만** — 여섯 섹션이 각 한도의 존재와 읽는 법을 말하고 값은 "프로젝트가 측정해 정한 값"이라 부른다.
- `.claude/rules/local/card-issuance-thresholds.md` — 로컬 전용(미러 없음, 이 저장소에서는 추적됨 — `.claude/rules/local/` 의 다른 규칙과 같다, G17). **수치만** — 각 값은 기준선 기록의 임계값 표와 같다(`TestCardIssuanceRuleValuesMatchBaseline` 가 두 추적 파일을 대조한다). `paths:` 는 배포 규칙과 같아 두 파일이 함께 적재된다.

수치를 배포 사본에 쓰지 않는 이유와 대안:

| 읽기 | 판정 | 이유 |
|---|---|---|
| 수치를 배포 사본에 카드 경로 없이 쓴다 | 반려 | 이 저장소 큐의 분포에서 나온 값(제품 줄 1,604 등)이 사용자 프로젝트에 배포되고, 값이 바뀔 때마다 템플릿 사본·`catalog.yaml` 해시를 다시 만들어야 한다 |
| 수치는 쓰지 않고 `.moai/reports/t1454/…` 기준선을 가리킨다(이터레이션 1) | 반려 | 템플릿 사본에 새 `t1454` 가 들어가 `card_id_leak_test.go` 의 (경로, 카드 id) 기준선 쌍 목록을 어기고, 사용자 프로젝트에 없고 이 저장소에서도 추적되지 않는 경로를 가리킨다 |
| **수치는 로컬 전용 규칙에만, 배포 사본은 메커니즘만** | **채택** | 배포 사본에 카드 id·`.moai/` 경로·측정값이 없고(`TestCardIssuanceTemplateIsMechanismOnly`), 값은 추적되는 한 곳에 있어 CI 가 기준선과 대조한다. 사용자 프로젝트는 자기 측정으로 자기 값을 정한다 |

섹션:

- **카드 크기**: 하한 트리거(작고 파일이 적은 카드는 같은 주 파일의 열린 카드와 묶음 후보)와 상한(넘으면 분할 후보). 값은 로컬 전용 규칙에, 도출은 기준선의 분포에.
- **후속 지적 규칙**: 카드가 만든 결함은 병합 전 카드 안에서 고친다. 이미 있던 결함은 구성요소별 부채 대장에 올린다(D12). `sync-auditor.md` 의 blocking/optional 분류와 `PASS-WITH-DEBT` 와 이어진다.
- **파생 깊이 상한**과 **동시 진행 한도**.
- **발행 체크리스트**: 발행 전 제시를 읽었는가(정확 일치·같은 구성요소·진행 중 겹침·완료 SPEC), 크기 하한·상한, 부모·출처 입력, 묶음 후보.
- **부채 대장**: 큐의 카드(`origin` 값 `follow-up` 과 구분되는 부채 표지), 구성요소 노드로 묶음 조회.

**발행 세션 도달성 (감사 D14).** 경로 한정 규칙은 그 경로의 파일을 여는 세션에서만 적재된다. 카드를 발행하는 세션(리더·운영자·`moai todo add` 를 부르는 세션)은 `gtd.md` 나 `manager-todo.md` 를 열지 않을 수 있다. 그래서 도달은 두 장치가 맡는다 — (i) 상시 로드 `kanban-dispatch.md` 가 규칙을 이름으로 가리키는 한 문장(기존 문장 압축으로 얻은 자리, 순증가 0)이 모든 세션에 도달하고, (ii) `card-issuance.md` 의 `paths:` 를 `workflow_rule_paths_pinned_test.go` 의 고정 목록에 등록해 경로가 바뀌면 시험이 붉어진다(점화 §1.3 — 조용히 멈추는 점검을 막는다). 그리고 규칙과 무관하게 발행 세션에 닿는 장치는 M1 의 제시 자체다 — `add` 가 stderr 로 내는 겹침 제시는 어떤 세션이 부르든 나간다. 이 도달은 AC-TCI-021 (h) 가 검증한다.

미러: 바이트 동일 쌍은 동일 편집, 분기된 쌍(`kanban-dispatch.md`, `sync-auditor.md`)은 분기 보존(§plan F.8). 템플릿 사본에는 카드 id·SPEC·REQ id·날짜·`.moai/` 경로·측정 수치를 쓰지 않는다. 기계 검사의 범위: 측정 수치 검사(`TestCardIssuanceTemplateIsMechanismOnly`)는 새 `card-issuance.md` 사본 하나에만 걸리고, 편집되는 다섯 사본은 카드 id·중립성 가드까지만 기계가 보며 숫자 임계값은 M5 커밋의 `git diff` 리뷰가 읽는다(수용 잔여, MU-121). 압축이 카드 id 기준선 짝(`gtd.md t696`, `kanban-dispatch.md t1330`, `kanban-dispatch-detail.md t133`·`t224`)의 문장을 지우면 같은 커밋에서 `card_id_leak_test.go` 의 기준선 항목을 지워야 한다(`plan.md` §D).

## §9 웹 보기 설계

`/todo?view=graph`. 읽기 이음매는 `readTodoQueue` 옆에 두고(`LoadPure` 와 읽기 전용 GTD 투영 소스 `GTDProjectionSource` 를 쓴다) 보관 카드·관계·해석기 결과를 `GraphVM{Nodes, Edges, Omitted}` 로 돌려준다. 노드 상한을 넘으면 열린 카드와 연결된 노드를 우선하고 나머지는 "N개 생략" 문구로 표기한다. 레이아웃은 입력이 같으면 출력이 같은 결정적 배치(연결 요소별, 파생·선행 방향의 깊이를 행으로)이고 SVG 는 Templ 이 서버에서 그린다. 새 JS 는 만들지 않는다(데이터는 서버 렌더). 스타일은 임베드 목록에 이미 있는 `console.css` 에 더하므로 임베드 목록은 바뀌지 않거나, 새 자산 파일을 만들면 `assets.go` 의 `//go:embed` 목록에 넣는다. 새 문자열은 모든 로케일 키를 가진다(`i18n_governance_test.go`). GET 전용이고 락·쓰기가 없으며 외부 요청이 없다(SPEC-WEB-TODO-QUEUE-001 의 규칙 유지). 락이 없다는 것은 락 파일의 바이트·mtime 으로 증언하지 않는다 — advisory flock 은 락 파일에 흔적을 남기지 않으므로 큐 읽기는 락이 없다는 `backlog_store.go` 머리 주석에 기대어, 저장소 자신의 `BacklogStore.Mutate` 로 락을 쥔 채 `GET /todo?view=graph` 가 2초(락 도우미의 대기 예산 3.3초보다 짧다) 안에 200 을 내는 시험(`TestTodoGraphViewDoesNotWaitOnQueueLock`)과 웹 패키지 비시험 소스의 락 토큰 어휘 검사(`TestTodoGraphViewReadOnly`)로 증언한다. 어휘 검사는 락을 시도만 하는 보기를 닫고, 한계는 토큰을 피해 이름을 바꾼 시도다(`acceptance.md` MU-122). 기존 표는 바뀌지 않는다(REQ-TCI-024).

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
| (i) M5 를 마지막에 두고 게이트 명령이 판정 | t1453 병합 수 0(핀), t1448 대조군 1; t1453 의 plan M4 가 같은 규칙 파일을 절체 시점에 고친다(브랜치 끝 `2a5f9c91c`의 plan.md). 게이트 술어는 어느 부모 경로로든 닿는 조상이다(§13) | 게이트가 안 열리면 규칙이 효력 없이 SPEC 이 닫힐 수 있다 |
| (ii) 지금 편집, t1453 이 흡수 | 규칙 파일 여섯과 사본 | 두 카드가 같은 파일을 만든 충돌, 템플릿 미러 분기 위에 편집이 쌓임 |
| (iii) M5 를 후속 SPEC 으로 분리 | Tier L 의 상한과 요구 수(24/25)에 여유가 없다 | 요구 REQ-TCI-021/022 를 옮겨야 한다 |
| 미충족 인도물 (a) 추적되는 초안만 내고 요구 문면대로 닫음 | REQ-TCI-022 | 규칙 미효력 상태가 완료로 읽힘 — 완료 보고에 명시 |
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
| (ii) 커밋된 증거(카드 귀속 병합)에서만 `edges.jsonl` | 큐의 유무와 무관한 출력, 비공개 상태 없음. 불변식은 (트리, 도달 가능한 이력) 위에서만 말한다(§6) | 열린 카드의 file→card 질의 불가, git 증거 추출 시간 미측정, CI 일치 미검증 |
| (iii) 둘 다 | | 두 운반체의 유지 |

사실: `.gitignore:344` 가 산출물을 추적 제외한다. 작업 기본값: (ii).

### D4 — M2 스키마 형태

| 선택지 | 근거 | 비용 |
|---|---|---|
| (i) JSON 컬럼 하나 | 분류 선례(REQ-TCD-002), 닿는 곳(두 표의 `ensureColumn`, INSERT·SELECT, parity, 동결 튜플 문자열 둘, 투영)을 한 번씩만 바꿈 | SQL 조회는 `json_extract`, 컬럼별 NOT NULL·CHECK 불가 |
| (ii) 속성별 컬럼(5개) | 임대·스탬프 선례 | 닿는 곳이 속성 수만큼 곱해짐 |

큐는 `LoadPure` 로 통째로 메모리에 올려 Go 에서 읽고 보관 1,047행에 SQL 속성 조회가 없다. finding 처분은 별도 두 표라 어느 쪽이든 컬럼 하나가 는다. 작업 기본값: (i).

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
| (ii) 값은 M0 기준선 기록(추적됨)에서 읽음, SPEC 은 기록과 측정 명령만 가리킴 | GB03·GB04·GB09·QB06·SB03·SB07 등 | M0 가 값 도출의 판단(어느 분위수)을 맡음 |
| (iii) 배포 규칙은 메커니즘만, 값은 로컬 전용 규칙 | 배포 사본에 이 저장소의 수치가 새지 않음 | 규칙 파일이 하나 더 |

작업 기본값: (ii)와 (iii)의 합 — 값은 추적되는 기준선에서 읽고, 값을 싣는 곳은 로컬 전용 규칙 한 곳이며, 배포 사본은 메커니즘만 쓴다(§8, 감사 D4).

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
| (i) 카드 속성이 원천, 관계는 투영 | 스칼라라 부모 ≤ 1 이 자동, 관계 동사가 쓰지 않으므로 순환·둘째 부모 위반이 관계 쓰기 경로에 없다(§5.3) | 관계 저장소에 `parent-of` 가 안 쌓임 |
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

## §12 기준선 운반체와 출하 입력 (감사 지적 D1·D4 대응)

### 12.1 이터레이션 2 가 관측한 사실

- `git check-ignore -v .moai/reports/t1454/baseline/x.txt` 는 `.gitignore:235:.moai/reports/*` 를 내고 종료 0 이다(`acceptance.md` G14). 이터레이션 1 이 "추적되는 형태"라 부른 기준선·M5 초안은 그 아래에 놓였으므로 어느 클론·CI 도 볼 수 없었다.
- 같은 위치의 감사 보고서는 디스크에 있고(`ls` 종료 0, G12) 추적되지 않는다(`git ls-files --error-unmatch` 종료 1, G13). "디스크에 있다"와 "추적된다"는 다른 사실이다.
- `.moai/docs/audit-artifact-convention.md` § Committing 은 감사 산출물이 로컬 파일이고 어떤 규약도 그것을 통합 브랜치로 강제하지 않는다고 하며, `.gitignore` 의 `.moai/reports/*` 는 그 정책이다. 이 SPEC 은 보고서를 트리에 강제로 넣는 길도 무시 규칙을 넓히는 길도 택하지 않는다.
- 새 후보 경로는 무시되지 않는다(`git check-ignore -v` 종료 1: G15 기준선, G16 임베드 데이터 파일, G17 로컬 전용 규칙, G18 M5 초안).

### 12.2 운반체 선택지

| 후보 | 추적 | 클론·CI 가시 | 출하 코드가 읽어도 되는가 | 판정 |
|---|---|---|---|---|
| (a) `.moai/reports/t1454/baseline/`(이터레이션 1) | 아니오(`.gitignore:235`) | 아니오 | 아니오 | 반려 — 순서 증인이 존재할 수 없고 시험이 어떤 클론에도 없는 기록을 읽는다 |
| (b) `.gitignore` 예외를 더해 reports 를 추적 | 예 | 예 | 아니오 | 반려 — 규약이 무시 규칙 확대를 금하고 다른 보고서까지 추적 대상이 된다 |
| (c) `git add -f` 로 한 파일을 강제 추적 | 예 | 예 | 아니오 | 반려 — 규약이 강제 추가를 금하고, 스테이징 규율에서 조용히 빠질 수 있다 |
| (d) SPEC 디렉터리 `baseline/` | 예 | 예 | 아니오(사용자 프로젝트에 SPEC 경로가 없다) | **채택** — 기록·재현 스크립트·허브 목록의 자리. 이 SPEC 은 이미 증거 문서(`research.md`)와 스크립트 부록을 이 디렉터리에 추적한다 |
| (e) `internal/**/testdata/` | 예 | 예 | 시험만 | 기록의 자리로는 반려 — 기록은 SPEC 과 함께 읽혀야 한다 |
| (f) 임베드 데이터 파일 `internal/homestate/hub_files.txt` | 예 | 예 | **예**(`go:embed`, 런타임 경로 없음) | **출하 입력으로 채택** — M0 의 `baseline/hub-files.txt` 에서 경로 열을 복사한 사본 |
| (g) Go 상수 | 예 | 예 | 예 | 반려 — 출처 머리 주석과 갱신 절차가 코드에 섞이고 데이터 대조 시험이 어렵다 |
| (h) 설정 키(`.moai/config/sections/`) | 예 | 예 | 예 | 반려 — 구성 구조체·기본값·템플릿 패리티를 새로 늘린다. 사용자 설정 키는 범위 밖 |

### 12.3 시험이 하는 일 (REQ-TCI-020 의 근거 사슬)

- `TestHubFileListFromBaseline` — 임베드 목록의 경로 집합이 추적되는 `baseline/hub-files.txt` 의 경로 집합과 같다. 기준선 파일이 없으면 **실패**한다(건너뛰지 않는다 — 건너뛰는 시험은 목록을 아무것도 대조하지 않은 채 통과로 읽힌다).
- `TestHubFileListMatchesMeasuringCommand` — 기준선이 적은 각 경로의 단일 호출 측정 명령(`git log --first-parent --since=<S> --until=<U> --format=%H <develop 팁 SHA> -- <path>` 의 출력 줄 수; 통합 브랜치 자체를 읽으므로 first-parent 가 맞다)을 기록된 develop 팁 SHA 에서 다시 돌려 기록된 개수와 비교한다. 그 SHA 를 클론이 갖지 않으면 사유를 출력하고 건너뛴다 — 건너뜀은 통과가 아니라 공백이며 레인의 전체 이력 클론에서는 통과해야 한다. 건너뜀은 `ok` 로 출력돼 묻지 않으면 보이지 않으므로(`verification-completeness.md` §1.3) 환경 변수 `CI` 가 있는 실행에서는 건너뛰지 않고 실패한다 — CI 의 test 작업은 `fetch-depth: 0` 이다(`.github/workflows/ci.yml:131`). 이 시험은 목록에 **오른** 경로가 기록 개수를 만족하는지(건전성)만 잰다. 목록에서 **빠진** 허브(완전성)는 M0 스크립트 실행 기록에만 기댄다 — 수용된 공백이다.
- `TestHubFileMeasurementSkipPolicy` — 위 건너뜀/실패 결정이 순수 함수이고 표로 읽는다: SHA 없음 + `CI` 미설정 → 건너뜀, SHA 없음 + `CI` 설정 → 실패, SHA 있음 → 실행. 함수만 읽으면 측정 시험이 함수를 거치지 않고 SHA 가 없으면 무조건 건너뛰어도 이 시험은 통과하므로(이터레이션 3 감사 D32), 측정 시험의 본문을 건너뜀·실패·실행을 기록하는 작은 보고 이음매 하나를 받는 함수로 두고 이 시험이 그 함수를 같은 세 입력으로 직접 구동해 결과를 관측한다.
- `TestHubFileLoaderReadsNoProjectPath` — 목록 적재 코드가 `.moai` 문자열 리터럴을 읽지 않고 임베드 데이터만 쓴다는 정적 검사. 경로를 조각으로 조립하는 적재는 이 검사를 피한다.
- `TestHubFileLoaderIgnoresProjectTree` — 동작 검사: 빈 임시 디렉터리와, `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/hub-files.txt`·`.moai/reports/` 아래에 표식 경로를 심은 임시 디렉터리를 각각 작업 디렉터리로 삼아 적재해 두 결과가 같고 임베드 목록과 같으며 표식이 없다. 한계: 작업 디렉터리와 리터럴만 본다 — 실행 파일 위치나 `$HOME` 에서 경로를 만드는 적재는 둘 다 통과한다(수용; 적재가 `go:embed` 한 함수라는 설계와 리뷰가 닫는다).

### 12.4 순서 증인

기준선 커밋 `B` 는 `baseline/` 아래 파일만 담는 **자기 커밋**이고 제품 경로를 바꾸는 어떤 커밋보다 앞서야 한다. 증인은 커밋 그래프이고 읽는 명령은 `acceptance.md` AC-TCI-001 의 순서 증인 1~5 와 완료 정의 #2 다(기준선 커밋 찾기, 자기 커밋 증명, B 뒤의 변경 커밋이 모두 B 의 후손임을 보이는 두 질의의 `--full-history` SHA 목록 일치, 측정 트리 `card-head:` 대 B 의 부모, **B 보다 앞선 제품 커밋이 0 개임**). 증인 3 은 범위 `B^..HEAD` 만 보므로 B 앞의 제품 커밋을 보지 못한다 — 증인 5(`git rev-list --count --grep=t1454 <B>^ -- internal cmd pkg scripts .claude .codex` = 0)가 앞쪽을 본다. **증인 3 은 개수가 아니라 SHA 목록을 비교한다**(이터레이션 3 감사 D26): 곁가지의 제품 커밋을 카드 id 가 적힌 병합으로 들이면 첫 질의는 그 커밋을, 둘째 질의(`--ancestry-path`)는 병합 커밋을 세어 개수만 같다. 그리고 `--full-history` 를 쓴다 — 기본 이력 단순화는 병합이 한쪽 부모와 트리가 같으면 다른 부모를 따라가지 않아 곁가지의 제품 커밋을 통째로 가리고(곁가지에서 넣었다 되돌린 경우), 카드 id 를 적은 정당한 병합 하나를 첫 질의에서만 가려 정당한 이력에 거짓으로 실패시키기도 한다. 두 현상은 스크래치 저장소와 실제 이력(카드 t1460 대역)에서 실행해 관측했다(`acceptance.md` C37~C40 과 "이터레이션 4 가 스크래치 저장소에서 실행한 순서 증인 탐침" 표). 제품 커밋은 카드 id 를 메시지에 담고 `internal`·`cmd`·`pkg`·`scripts`·`.claude`·`.codex` 를 바꾸는 커밋이고 `.moai/` 만 바꾸는 SPEC·기준선 커밋은 세지 않는다 — 경로 필터는 정의의 일부다(필터가 없으면 `t1454` 를 적은 다른 카드의 SPEC plan 커밋이 센다: `acceptance.md` C27). 기준선이 무시되는 경로에 있으면 이 증인이 존재할 수 없다. 카드 id 없는 제품 커밋은 증인 3·5 가 못 보는 수용된 잔여다(MU-110).

### 12.5 M5 초안

Mode B 의 초안(`anchors.md`, `card-issuance.md`, `card-issuance-thresholds.md`)도 같은 이유로 `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/` 에 둔다 — "리더가 적용할 수 있는 초안"이 추적되지 않으면 리더의 클론에 없다.

### 12.6 수치의 자리

값은 `baseline.md` 의 임계값 표에서 읽고(REQ-TCI-001), 값을 싣는 곳은 `.claude/rules/local/card-issuance-thresholds.md` 하나뿐이다(§8, REQ-TCI-021). 기준선과 로컬 규칙 둘 다 추적되므로 `TestCardIssuanceRuleValuesMatchBaseline` 이 CI 에서 둘을 대조할 수 있다.

## §13 M5 게이트 술어 (감사 지적 D2 대응)

### 13.1 이터레이션 1 게이트의 눈먼 곳

이터레이션 1 은 `git rev-list --first-parent --count … HEAD` 로 t1453 병합을 읽었다. 카드 브랜치는 측정 전에 `git merge develop` 으로 develop 을 흡수한다(`plan.md` §C.3). 흡수 병합의 첫 부모는 카드 브랜치의 이전 끝이고 두 번째 부모가 develop 이다 — develop 의 병합 커밋(t1453 이 착지하면 생기는 `merge(t1453)`)은 두 번째 부모 쪽에만 있으므로 `--first-parent` 는 t1453 이 착지한 뒤에도 0 을 읽는다. 양성 대조 t1448·t1344 는 카드 브랜치가 갈라지기 **전**의 커밋이라 첫 부모 경로에 남아 1 을 읽으므로 이 맹점을 드러내지 못한다.

### 13.2 고정 SHA 로 본 맹점

`WT-github-flow-default` 브랜치의 두 부모 병합 `b05c3be9049822851b4cc2088cd3fc011fe39899`(`acceptance.md` G11)는 첫 부모 `5e31abbcb3d09a03fc782bcfc842d8221b508dba` 와 두 번째 부모 `46be0b8c87c76591cff7ec46b007571dafd74ca5`(`merge(t1407)` 제목의 develop 병합)를 가진다. `46be0b8c8…` 은 첫 부모의 조상이 아니고(C17, 종료 1) 병합의 조상이다(C18, 종료 0). 같은 선택자를 t1407 에 걸면 `--first-parent` 는 0(C19), 모든 부모를 걷는 형태는 1(C20)이다. 이 SHA 들은 미푸시 브랜치의 객체이므로 그 객체가 있는 클론에서만 재현된다.

### 13.3 술어와 판독

술어를 말로 하면 **"t1453 의 착지 병합 — 병합 제목 줄이 develop 을 그 브랜치에 흡수한다고 하지 않는 병합 — 을 기록한 커밋이 HEAD 의 조상이다"**(어느 부모 경로로든). 판독은 `plan.md` §F.8 의 다섯 가지다 — (1) `--first-parent` 없는 제목 커밋 수 `T` 에서 그 가운데 병합 제목 줄에 `absorb` 가 든 흡수 방향 병합 수 `A`(`--grep` 하나에 카드 id 와 `absorb` 를 같은 패턴으로 적는다 — git 이 정규식을 메시지의 줄 단위로 맞추므로 같은 줄이다)를 뺀 착지 후보, 그리고 `T ≥ 1` 인데 `T − A = 0` 이면 닫힘으로 읽기 전에 판독 2 의 목록을 읽는다, (2) 후보 목록에서 `absorb` 없는 줄의 SHA 를 고정하고 같은 줄 형태의 `git rev-list --no-walk --count -E -i --grep='^merge[( :]+(card )?t1453[^0-9].*absorb' <S>` 가 0 이며 `git merge-base --is-ancestor <S> HEAD`, (3) 양성 대조 t1448·t1344(그리고 t1448 의 `A` 가 0, 본문에만 `absorb` 가 든 착지 t1439 의 T 가 1 이고 `A` 가 0), (4) 같은 명령의 `--first-parent` 형태보다 엄격히 큰 일반 병합 수(이식 가능한 런타임 대조: 핀에서 333 대 214), (5) 제목 형태가 다를 때의 직접 SHA 판독. 판독 4 가 "눈먼 선택자 방지"를, 판독 1 의 `A` 가 "착지 없이 열림 방지"를 맡는다 — 선택자가 첫 부모 경로 밖의 커밋을 본다는 증거가 없으면 게이트는 "미측정 = 미충족"이다.

**흡수 방향 병합이 왜 문제인가.** 이터레이션 2 의 선택자는 제목 grep 이라 `merge(t1453): absorb … into WT-github-flow-default` 도 센다. 미푸시 브랜치 `WT-github-flow-default` 가 그런 커밋 다섯 개를 이미 갖고 있고(`acceptance.md` G23), 카드 브랜치가 미착지 코드에 기대려고 그 브랜치를 병합하면(kanban-dispatch 의 새 카드 조항) 이 다섯이 HEAD 에서 닿는다. 같은 두 질의를 흡수 병합 `b05c3be90…` 에 걸면 `T` 가 5, `A` 가 5 라 후보가 0 이고(C28·C29), 실제 착지 병합(t1448)은 `A` 가 0 이다(C30). 이터레이션 2 설계 §13.4 의 "그런 커밋은 이 카드의 브랜치에 존재하지 않는다고 가정한다"는 가정은 그 브랜치에 실제로 존재한다는 관측 앞에서 폐기했다.

**`absorb` 를 어디서 찾는가(이터레이션 3 감사 D27).** 이터레이션 3 의 `A` 는 `--all-match --grep=absorb` 라 메시지 **전체**에서 그 낱말을 찾았다. 병합 본문에 "absorbed tree" 같은 구절이 있는 진짜 착지는 `A` 에 세어져 후보가 0 이 되고 게이트가 닫힌 채 읽힌다 — 같은 착지 t1439(`d53e6ca59…`)에서 T=1, 옛 `A`=1(C35), 새 `A`=0(C34)이다. 핀 `ad02a5677` 에서 카드 id 를 담은 병합 제목 333개 가운데 제목 줄에 `absorb` 가 든 것이 76개, 본문에만 든 착지가 28개이며 옛 `A` 는 둘을 합한 104 였다(`python3` 로 줄 단위로 읽은 값). 새 `A` 는 `--grep='^merge[( :]+(card )?t1453[^0-9].*absorb'` 하나다 — git 의 `--grep` 은 정규식을 메시지의 줄 단위로 맞추므로 `.*` 가 줄바꿈을 건너지 못하고 병합 제목(첫 줄)에 `absorb` 가 있어야 센다. 판독 2 의 점검도 같은 형태로 맞췄다(옛 `--grep=absorb` 는 본문에 그 낱말이 든 착지 커밋에서 1 이라 착지를 흡수 방향으로 읽는다: C36).

### 13.4 한계

- 판독 1 은 `merge(t1453)`·`Merge card t1453` 같은 제목 형태를 가정한다. t1453 이 다른 제목으로, 또는 빨리감기·스쿼시로 착지하면 판독 1 이 0 으로 남아 게이트가 닫힌 채 읽히고(안전한 방향: Mode B) 판독 5 가 보완한다. 착지 병합의 **제목 줄**에 `absorb` 라는 낱말이 들어 있으면 `A` 가 그것을 흡수 병합으로 세어 후보에서 빼므로 역시 닫힌 쪽(안전한 방향)이다 — 핀에서 제목 줄에 `absorb` 가 든 76개 가운데 착지로 읽히는 제목이 적어도 셋(`584cfc1a5`, `615d18c1f`, `4c3b1653c`; 제목 모양만 읽은 추정)이라 드물지 않으므로 `T ≥ 1` 인데 `T − A = 0` 이면 목록을 읽게 했다. 본문에만 든 낱말은 세지 않는다. 선택자의 `^` 는 메시지의 어느 줄 머리에나 맞으므로 본문 줄이 `merge(t1453)` 로 시작하면 `T` 에 들어오는데 핀의 12,444개 커밋에서는 카드 id 를 가리지 않는 같은 패턴으로 본 그런 커밋이 0개였다(`python3` 줄 단위 읽기, 원장 행 아님).
- 고정 SHA 대조 행(C17~C20, C28·C29·C31)은 미푸시 브랜치의 객체가 필요하다. 이식 가능한 대조는 판독 4 의 부등식이다.
- 착지가 되돌려진(revert) 경우는 원래 병합이 조상으로 남아 게이트가 열린 채로 읽힌다 — 이 이력에 되돌림 사례가 없어 측정하지 못한 추정이다. 판독 2 가 SHA 를 `progress.md` 에 적어 리더가 그 출처를 읽을 수 있다.
- 후보 목록에서 `absorb` 없는 줄을 고르는 것(판독 2)은 사람이 읽는 단계다 — 목록이 여러 줄이면 각 줄의 SHA 에 점검 질의를 건다.
