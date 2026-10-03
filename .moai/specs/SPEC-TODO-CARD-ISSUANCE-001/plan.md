# SPEC-TODO-CARD-ISSUANCE-001 — 계획

Tier L. 카드 t1454, 계획 시작 HEAD `2de0a2cb6`(전체 SHA `2de0a2cb613b04765a1554f86685a3b48e0be806`), 워크트리 `.claude/worktrees/t1454`, 브랜치 `WT-card-issuance-overlap-graph`. 개발 방식은 TDD(`.moai/config/sections/quality.yaml` 의 `constitution.development_mode: tdd`)다 — 마일스톤마다 RED 테스트를 먼저 커밋하고 GREEN 으로 넘긴다. 시간 추정은 쓰지 않는다. 우선순위 라벨과 순서("A 를 마치고 B 를 시작")만 쓴다.

문서 구성: 결정이 바뀔 가능성이 큰 순서(§F.0)가 먼저 나오고, 마일스톤은 의존 순서(실행 순서)로 이어진다. 기계적인 작업(M0 의 측정 복사, M5 의 미러 정합)은 뒤에 있다.

## §A 배경

- **Epic 참조.** 운영자가 선언한 Epic 은 없다. 이 SPEC 은 단독 SPEC 이다(`.claude/rules/moai/development/sprint-round-naming.md` 의 단일 SPEC 허용 형태). 같은 주제(카드 큐의 품질)의 인접 SPEC 은 SPEC-TODO-AUTO-PICK-001(completed, 카드 t1448)과 SPEC-GITHUB-FLOW-DEFAULT-001(in-progress, 카드 t1453, **이 트리에는 없다** — `WT-github-flow-default` 브랜치에만 있다).
- **용어.** 마일스톤은 M0~M6 이다(`Milestone`). M5 만 게이트가 있고 나머지는 의존 순서로 실행한다.
- **무엇을 하는가.** `spec.md` §A.2 와 §C.
- **측정 근거.** `research.md` §3. 카드가 인용한 git 쪽 수치는 이 트리에서 전부 재현됐고, 큐 쪽 수치는 스냅숏 위에서 다시 쟀으며, 발행 시점 제시의 설계를 바꾼 새 측정 두 가지가 있다 — 어휘 겹침이 약한 변별자라는 것(SB01~SB05), 예상 파일 입력이 거의 비어 있다는 것(SB06~SB07).

## §B 알려진 문제

1. **낡은 전제 여섯 가지**는 `spec.md` §A.3 에서 바로잡았다(near-duplicate 기록 위치, t1349 완료, `assign --after` 의 정체, 병합 금지 문면의 위치, t1448/t1453 착지 상태, 추적되지 않는 `edges.jsonl`).
2. **`assign --after` 의 구조적 한계.** 이전 카드의 병합 가드는 할당 간선(`homestate/card_transition.go` 의 `guardAssign`)에서 돈다. 이전 카드가 병합되기 전에는 다음 카드를 레인에 할당할 수 없어서, "같은 레인에 다음 카드를 미리 걸어 둔다"는 지금의 도구로 표현되지 않는다(`research.md` §5.4).
3. **선택 호의 오류 처리(읽기 추정, 미확인).** 선행 미병합 `after` 후보를 만난 임대 선택 호가 건너뛰지 않고 오류로 끝날 수 있다 — M4 가 첫 RED 테스트로 확인한다(`research.md` §10 항목 9).
4. **규칙 파일의 예산과 분기.** `gtd.md` 40,037자(37자 초과), `kanban-dispatch-detail.md` 43,138자(3,138자 초과), `kanban-dispatch.md` 는 상시 로드 28,092자. 로컬과 템플릿의 `kanban-dispatch.md`·`sync-auditor.md` 사본은 이미 갈라져 있고 어느 가드에도 등록돼 있지 않다.
5. **같은 파일을 만지는 다른 카드**: t1453(절체 시점에 같은 규칙 파일을 고친다, 미착지), t1450·t1452·t1359(대기) — `research.md` §8.
6. **레인 가드.** 레인 세션은 `add --dry-run` 도 거절된다(`todoRefuseLaneMutation` 의 읽기 전용 동사 목록에 `add` 가 없다).

## §C 사전 점검 (M0 시작 전, 결과는 `progress.md` §E.2 에 기록)

1. 핀과 청결: `git rev-parse HEAD` 가 계획 시작 SHA 와 같은지(다르면 재측정·재고정), `git status --short` 가 비었는지.
2. 판정 도구를 트리에서 빌드해 **경로로** 호출한다: `go build -o <scratch>/moai ./cmd/moai`. 설치된 `moai` 는 쓰지 않는다 — 측정 인용에는 트리 HEAD 와 판정 빌드의 커밋을 함께 적는다(`verification-claim-integrity.md` §2.2).
3. 병렬 세션 점검(`gitflow-lane-protocol.md` §11 의 비교): `git fetch origin develop` 가 끝난 뒤 `git rev-list --count --left-right origin/develop...develop`. 카드 브랜치는 develop 을 병합으로 흡수한 뒤(`git merge develop`, 카드 트리 안에서) 측정을 다시 고정한다.
4. M5 게이트 명령과 대조군을 한 번 읽어 둔다(§F.8). 지금 읽으면 게이트는 닫혀 있다(t1453 병합 수 0, t1448·t1344 대조군 1).
5. 큐 스냅숏을 scratchpad 로 복사하고(원본 DB 를 읽기 전용 플래그로 여는 것은 샌드박스에서 실패했다) 기준선 스크립트를 복사한다.
6. 레인 환경: 이 트리의 세션은 레인 표지가 있어 라이브 `moai todo add` 는 거절된다. add 경로 검증은 `go test` 로만 한다(테스트 도구가 레인 변수를 지운다, `sdClearLaneEnv`).
7. 계획 단계에서 이미 한 일: `TestACCounterFullCorpusMatchesBaseline` 이 새 SPEC 을 "absent-from-snapshot"으로 보고하고 실패하지 않음을 확인한다(`progress.md` §G).

## §D 제약

- **허브 파일 직렬.** 이 카드 자신의 실행에서도 `internal/cli/todo.go`(M1·M2·M3·M4 가 모두 만진다)와 `internal/template/catalog.yaml` 은 허브다. 마일스톤을 순서대로 한 작성자가 진행하므로 겹치지 않는다. `catalog.yaml` 은 템플릿 아티팩트를 고친 마일스톤마다 같은 커밋에서 생성기로 다시 만든다.
- **순서 귀속.** 기준선 산출물(M0)은 자기 커밋으로 착지하고, 그것이 측정한 변경 커밋보다 앞서야 한다(`verification-claim-integrity.md` §2.3). 커밋 그래프만이 순서의 증인이다.
- **검증은 마일스톤이 만지는 패키지로 한정**한다. `go test ./...` 를 돌리지 않는다(`gitflow-lane-protocol.md` §8). 시험 명령은 레인 변수를 한 번의 복합 호출에서 지운다: `SCRUB = unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED &&`. 모든 시험 이름은 `-run` 에서 하나씩 `^TestName$` 로 앵커하고(여러 개는 `^TestA$|^TestB$` 로 가지마다 쓴다), 같은 패턴에 `go test <패키지> -list '<패턴>'` 를 짝지어 출력된 이름 수가 지정한 시험 수와 같은지 본다 — 0개를 고르고도 `ok` 가 나오는 선택자를 증거로 쓰지 않는다(`verification-completeness.md` §1.1).
- **예산.** 상시 로드 `kanban-dispatch.md` 는 순증가 0, 건드린 모든 규칙 파일은 40,000자 이하이거나 이전보다 크지 않아야 한다(D15). 1,000바이트를 넘는 상시 로드 증가는 `rule-authoring.md` 의 크기·비용 진술을 커밋 본문에 쓴다.
- **템플릿 가드.** 새 `t####` 를 템플릿 사본에 넣지 않는다(`card_id_leak_test.go`), 템플릿 사본에는 SPEC·REQ id·날짜·9자 이상 16진 단어가 없어야 한다(`todo_skill_doc_parity_test.go` 의 중립성 정규식, 바이트 동일 쌍 전체에 사실상 적용), `kanban-dispatch-detail.md` 의 `paths:` 는 바꾸지 않는다(`workflow_rule_paths_pinned_test.go`).
- **레인 가드.** 새 읽기 동사 `trace` 는 `todoLaneReadOnlyVerbs` 에 넣고, `merge`·`bundle` 은 넣지 않는다.
- **분석 불변식.** 분석기·`analyze`·`relate` 는 카드를 접지 못한다 — 코드 모양과 테스트로 계속 강제한다(`TestTodoMergeNeverInvokedByAnalysis`).

## §E 자기 검증

- `moai spec lint SPEC-TODO-CARD-ISSUANCE-001`, `--strict` — 명령과 종료 코드는 `progress.md` §G 에 기록한다.
- 요구·기준 개수: 요구 24 ≤ 25, 기준 24 ≤ 25, 모듈 5 ≤ 5(Tier L 상한). 모든 요구가 기준 하나 이상에서 인용된다.
- RED-now: 출시 차단 21개 기준은 각각 핀 트리에서 실행한 원장 행을 가진다(`acceptance.md`).
- SPEC ID 점검은 Bash 로 실행해 `PASS` 를 출력했다.

## §F 마일스톤

### F.0 결정이 바뀔 가능성 순서 (가장 바뀔 것부터)

| 순위 | 결정 | 바뀌면 다시 해야 하는 것 | 마일스톤 |
|---|---|---|---|
| 1 | D5 MCP 전달·engage 합류, D8 유사도 척도·소음, D13 플래그 입력 방식(사용자가 직접 보는 흐름) | 제시 렌더러·플래그 파싱·테스트 기대값 | M1, M2 |
| 2 | D4 스키마 형태, D9 closed-at, D10 부모 원천, D11 처분 효과(데이터 모형) | 컬럼·읽기/쓰기 목록·동결 튜플 문자열·JSON 투영 | M2 |
| 3 | D2 저장소 통합, D3 간선 운반체, D14 묶음 메커니즘(새 형식·인터페이스) | 해석기·간선 층·팩토리 `cards` 컬럼과 선택 호 | M3, M4 |
| 4 | D1 M5 순서, D6 병합 독트린 범위, D12 부채 대장 운반체(독트린) | 규칙 문면·개정 행 | M4, M5 |
| 5 | D7 수치 임계값 | 값 하나가 바뀌면 기준 리터럴만 | M0 에서 확정 |
| — | D15(이미 정해짐) | — | M5 |

### F.1 실행 순서 (의존)

M0 → M1 → M2 → M3 → M4 → M6 → M5(게이트). 한 마일스톤을 마치고 다음을 시작한다.

- M1 은 기존 저장소만으로 제시를 만든다. 진행 중 레인 겹침 입력은 예상 파일 필드가 생기면(M2) 측정값이 되고, 그 전에는 `unmeasured` 로 나간다(REQ-TAU-013 의 열린 입력 집합과 같은 모양).
- M2 가 속성과 처분을 추가한다. M3 은 M2 의 부모·출처 속성을 읽어 `parent-of`/`follow-up-of` 를 투영한다. M4 는 M2 의 예상 파일과 M3 의 `merged-into` 를 쓴다. M6 은 M3 의 해석기를 읽는다.
- M5 는 t1453 착지에 걸려 있어 마지막에 둔다(D1). M6 은 M5 와 독립이다.

### F.2 M0 — 기준선 반입과 재측정 (Priority High, 기계적)

- **목적**: 카드의 측정과 이 계획의 새 측정을 추적되는 형태로 옮기고 실행 단계에서 다시 잰다. 이후 모든 임계값은 이 기록에서 읽는다(REQ-TCI-001).
- **선행**: §C 사전 점검.
- **파일(모두 신규)**: `.moai/reports/t1454/baseline/baseline.md`(figure 행), `.moai/reports/t1454/baseline/` 아래 스크립트와 출력(카드의 `lib.py`·`01`~`06`, 이 계획의 부록 A 다섯 개), 재현 절차 `README.md`.
- **작업**: (1) 카드 스크립트를 복사하고 원본이 사라졌다면 `research.md` §3.1~§3.4 의 방법 서술로 재구성한다. (2) `fp.tsv`·`allc.tsv` 를 실행 시점 트리에서 다시 만들고 스크립트를 돌려 `research.md` §3 의 figure 35개(GB 16 + QB 10 + SB 9)를 다시 잰다. 각 행은 `figure:`, `command:`, `output:`, `tree:` 줄을 가진다. (3) 핵심 git 쪽 헤드라인이 `research.md` §3.1 과 다르면 `drift:` 줄에 이유를 적는다. (4) `baseline.md` 끝에 "임계값" 표를 둔다 — §F.11 의 작업 기본값을 재측정한 분포에서 다시 도출하고 각 값에 측정 명령을 붙인다. (5) 허브 파일 목록(만진 서로 다른 카드 수 순위)과 허브 임계값을 정한다.
- **RED/GREEN**: RED 는 `ls .moai/reports/t1454` 가 exit 1(원장 L1). GREEN 은 디렉터리가 있고 `figure:`·`command:`·`tree:` 줄 수가 같다.
- **검증 명령**: `ls .moai/reports/t1454/baseline`, `git grep -c -F "figure:" -- .moai/reports/t1454/baseline/baseline.md`, 같은 형태로 `command:`, `tree:`; 세 수가 같고 35 이상이다.
- **종료 증거**: AC-TCI-001. **되돌리기**: 커밋 되돌리기(코드 변경 없음). **커밋**: 이 마일스톤의 커밋은 다른 어떤 실행 커밋보다 앞선다.

### F.3 M1 — 발행 시점 제시 (Priority High, 사용자가 보는 흐름이라 바뀔 가능성이 가장 크다)

- **목적**: `add` 가 막지 않고 알린다(REQ-TCI-002~006).
- **파일**: `internal/kanban/backlog_issuance.go`(신규, 읽기 전용 이웃·구성요소·완료 SPEC 조회와 표시 하한·상한 상수), `internal/cli/todo_issuance.go`(신규, 제시 조립과 렌더, 탐침 시간 상한), `internal/cli/todo.go`(`scanTodoAddArgs`·`todoAddScan` 에 `--dry-run`, 잠금 밖에서 제시를 계산·출력), `internal/cli/mcp_todo.go`(결과 텍스트에 제시 덧붙임), `internal/cli/gtd.go`(engage 가 제시만 출력), 각 `_test.go`.
- **RED 테스트(먼저 쓴다)**: kanban — `TestIssuanceNeighborsIncludeArchivedAndDropped`, `TestIssuanceNeighborsStripDropPrefix`, `TestIssuanceNeighborsLimitAndFloor`, `TestIssuanceSameComponentOpenCards`, `TestIssuanceCompletedSpecCoverage`, `TestIssuanceInFlightOverlapUnmeasured`. cli — `TestTodoAddPresentationStderrOnly`, `TestTodoAddPresentationNeverBlocks`, `TestTodoAddPresentationProbeOutsideLock`, `TestTodoAddPresentationTimeBound`, `TestTodoAddPresentationDryRunWritesNothing`, `TestTodoAddMCPCarriesPresentation`, `TestGTDEngagePresentationRecordsNothing`.
- **구현 요점**: 이웃 조회는 `LoadPure` 스냅숏을 입력으로 받는 순수 함수이고 `ClassifyCardText` 를 쓰지 않는다(분류기는 dropped 를 건너뛰는 독트린을 지킨다). dropped 본문은 `[DROPPED — ` 접두사를 벗기고 사유는 따로 보인다. 보관 카드는 `rec.Archived` 에서 읽는다(추가 질의 없음). 출력은 stderr 에 상위 3개를 점수·척도·상태와 함께 적고, 표시 하한 미만의 어휘 이웃은 `--dry-run` 에서만 보인다(D8 작업 기본값). 진행 중 레인 겹침은 `rec.Runtime.Assignments` 의 레인 카드와 예상 파일을 비교하고, 예상 파일이 없으면 `unmeasured` 다. git 탐침은 잠금 밖에서, 시간 상한 안에서만 돌고 상한을 넘으면 `unmeasured (time bound)` 로 출력한다. 시퀀스는 (스냅숏 읽기 → 제시 계산 → `Mutate` 로 admit → 제시 출력 → stdout 한 줄)이다.
- **검증 명령(패키지 한정)**: `SCRUB go test ./internal/kanban -run '^TestIssuanceNeighborsIncludeArchivedAndDropped$' -count=1 -v` 등 위 이름 각각(위 -list 짝). 회귀: `^TestClassifyCardText$`, `^TestNormalizeCardText$`, `^TestTokenSetJaccard$`(kanban), `^TestTodoAddRefusesExactDuplicate$`, `^TestTodoAddNearDuplicateRecordsOnly$`, `^TestTodoAnalysisNeverReordersQueue$`, `^TestTodoAdd_PrintsIDAndPosition$`, `^TestSD_AC014_MCPMatchesCLIWithProjectRoot$`(cli). 정적: `go vet ./internal/kanban/... ./internal/cli/...`, CI 가 쓰는 버전의 `golangci-lint run ./internal/kanban/... ./internal/cli/...`.
- **종료 증거**: AC-TCI-002~007. **되돌리기**: 커밋 되돌리기(스키마 변경 없음).

### F.4 M2 — 카드 스키마 (Priority High, 되돌리기 비용이 가장 크다)

- **목적**: 발행 속성과 finding 처분을 가산적으로 추가한다(REQ-TCI-007~011).
- **파일**: `internal/kanban/backlog_store.go`(타입), `backlog_sqlite.go`(`ensureColumn` 가산 경로를 `ensureSchema` 의 **마지막** retrofit 로, `backlogItemsTableColumns` 에는 넣지 않는다), `backlog_migrate.go`(INSERT 두 곳, `readSnapshot`/`readArchive` 와 `columnExpr`, `assertBacklogParity`), `backlog_schema_freeze_test.go`(동결 컬럼 튜플 문자열 둘에 같은 커밋에서 새 튜플을 추가), `todo_queue_merge.go`·`todo_merge_procedure.go`(항목 복사 경로가 새 컬럼을 나르는지), `internal/cli/todo.go`(`add` 의 명시 플래그), `todo_drop.go`(사유 속성 저장, 접두사 유지), `todo_claim.go`(`todoJSONProjection` 또는 `omitempty` 로 골든 무변), `todo_export.go`(새 필드 처리 확인).
- **RED 테스트**: `TestBacklogIssuanceColumnRetrofit`, `TestBacklogIssuancePureReaderNoDDL`, `TestBacklogIssuanceArchiveRestoreRoundTrip`, `TestBacklogParityCoversIssuance`, 동결 테스트(기존 `TestTodoHistoryAddsNoSchemaChange`·`TestSchemaFreezeRecordsTransitionStamps` 를 새 튜플 기대로 고친다), `TestTodoDropStoresReason`, `TestTodoDropKeepsTextPrefix`, `TestCardClosedAtAccessor`, `TestFindingDispositionRecordOnly`, `TestTodoRelateDispositionVerb`, `TestFindingDispositionAbsentForLegacy`.
- **구현 요점(D4 작업 기본값)**: `items`·`archived_items` 에 nullable TEXT 컬럼 하나(`issuance`, JSON: 스폰한 카드, 출처, 크기 추정 줄 수, 예상 파일, drop 사유), `findings`·`archived_findings` 에 nullable TEXT 컬럼 하나(`disposition`). 없음은 SQL NULL·nil 이고 `omitempty` 로 직렬화한다. closed-at 은 저장하지 않고 접근자 하나가 `archived_at`·`dropped_at` 에서 읽는다(D9). 과거 카드는 소급하지 않는다. `add` 의 새 플래그는 폴스루 경로가 플래그를 받지 않으므로 `add` 서브커맨드에만 있다(D13).
- **검증 명령**: `SCRUB go test ./internal/kanban -run '^TestBacklogIssuanceColumnRetrofit$' -count=1 -v` 등(-list 짝). 회귀: `^TestTodoListJSON_GoldenByteIdentity$`(cli), 기존 동결·retrofit 테스트(`internal/kanban`), 큐 병합·이주 테스트. 정적 검사 M1 과 같다.
- **종료 증거**: AC-TCI-008~011 과 회귀 가드 AC-TCI-009. **되돌리기**: 커밋 되돌리기. 컬럼은 DB 에 남아도 읽는 쪽이 모르면 무해하다(가산 컬럼 선례). 단, 되돌리기 전에 쓴 데이터는 읽히지 않는다.

### F.5 M3 — 관계 모델과 그래프 (Priority High)

- **목적**: 일곱 종류 어휘, 종류별 제약, 해석기, 추이 질의, card→file 간선(REQ-TCI-012~017).
- **파일**: `internal/kanban/backlog_relation.go`(신규: 어휘 정규화·매핑·제약·해석기), `backlog_store.go`(순환 가드 도우미 일반화 — `WaitsOnOf` 는 그대로 두고 호출한다), `internal/cli/todo_relate.go`(새 종류 수용, 처분 동사, 제약 거절), `internal/cli/todo_trace.go`(신규), `internal/cli/todo.go`(서브커맨드 등록과 `todoLaneReadOnlyVerbs` 에 `trace`), `internal/graph/card_file.go`(신규: 카드 귀속 병합 커밋의 변경 파일 간선), `graph.go`(층 호출과 정렬), `meta.go`(지문 출처 추가), `internal/cli/graph.go`(간선 종류 집계 루프).
- **RED 테스트**: `TestRelationOntologyMapsLegacyKinds`, `TestRelationLegacyReadersByteIdentical`, `TestRelationConstraintRefusesCycle`, `TestRelationConstraintCardinality`, `TestTodoRelateConstraintRefusalByteIdentical`, `TestCardGTDResolverBothDirections`, `TestCardGTDResolverAddsNoGTDWritePath`, `TestTodoWhyShowsGTDRelations`, `TestTodoTraceTransitiveDeterministic`, `TestTodoTraceBreaksCycles`, `TestTodoTraceLaneReadOnlyAllowed`, `TestTodoTraceWritesNothing`, `TestGraphCardFileEdgesDeterministic`, `TestGraphCardFileEdgesCarryNoQueueState`, `TestGraphCheckNoticesCardFileSource`.
- **구현 요점**: 저장된 소견 행은 다시 쓰지 않는다 — 읽기 시 매핑(design §5)과 새 종류는 기존 `findings` 어휘 확장으로 기록한다. `parent-of`/`follow-up-of` 는 M2 속성에서 읽기 투영(D10). 해석기는 `gtd_items.card_id` 로 양방향이고 `gtd_relations` 에는 쓰지 않는다. 간선 층은 `moai graph build` 안에서 카드 귀속 first-parent 병합과 그 변경 파일에서만 만들고(큐를 읽지 않는다) 출처 지문에 병합 목록을 넣는다.
- **검증 명령**: kanban·cli·graph 패키지 각각 위 이름. 회귀: 기존 `todo why`·`list`·`export`·픽업 필터·`TestAutoRank*` 계열(cli), `internal/graph` 의 기존 `TestBuild*`·`TestCheck*`, GTD 관련 기존 테스트(`internal/kanban` gtd 계열).
- **종료 증거**: AC-TCI-012~016. **되돌리기**: 커밋 되돌리기(저장 행 불변이므로 데이터 되돌리기 없음).

### F.6 M4 — 묶음·병합·허브 직렬 (Priority High, 임대 경로를 건드린다)

- **목적**: 묶음 경로, 운영자 호출 병합 동사, 허브 파일 체인(REQ-TCI-018~020).
- **파일**: `internal/homestate/factory.go`(`cards` 에 묶음 컬럼을 `ALTER TABLE ... ADD COLUMN` 목록 방식으로 추가 — D14 기본값), `card_record.go`·`card_picked.go`·`card_transition.go`, `internal/cli/factory_card.go`(적재 동사와 선택 호의 묶음 예약·선행 미병합 건너뛰기), `internal/cli/todo_merge.go`(신규), `internal/cli/todo.go`(서브커맨드 등록).
- **RED 테스트**: `TestTodoMergeRecordsAndDrops`, `TestTodoMergeRefusesPickedAndCycles`, `TestTodoMergeNeverInvokedByAnalysis`, `TestTodoMergeRefusedInLane`, `TestFactoryNextBundleSerialLane`, `TestFactoryBundleKeepsSerialSlot`, `TestFactoryAssignBundleOrderGuard`, `TestFactoryAssignBundleHubChain`, `TestHubFileListFromBaseline`, `TestFactoryKeepSetReadsNoFileOverlap`. 첫 테스트는 선행 미병합 `after` 후보가 선택 호 맨 앞에 설 때의 현재 동작을 관측으로 기록한다(§B.3).
- **구현 요점**: 병합은 `<into>` 본문에 `<from>` 본문을 명시 절로 덧붙이고, `merged-into` 관계를 기록하고, `<from>` 을 drop 사유와 함께 닫는다. picked·이미 병합·순환은 거절한다. 레인과 분석 경로에서는 호출할 수 없다. 묶음은 팩토리 레코드의 묶음 식별·순번 컬럼과 적재 동사로 만들고, 선택 호는 (a) 묶음 선행이 미병합인 멤버를 건너뛰고 (b) 묶음 레인이 다음 멤버를 우선 받고 (c) 다른 레인에는 묶음 멤버를 주지 않는다. 직렬 슬롯 의미와 keep-set 은 그대로다. 허브 체인은 기준선 기록의 허브 목록과 두 열린 카드의 예상 파일 교차로 생성한다.
- **검증 명령**: `SCRUB go test ./internal/homestate -run '^TestFactory...$' -count=1 -v`(해당하는 homestate 테스트), `./internal/cli` 의 위 이름들. 회귀: 기존 `TestFactoryNext*`·`TestSD_AC014_...`, 직렬 슬롯·keep-set 테스트.
- **종료 증거**: AC-TCI-017~019. **되돌리기**: 커밋 되돌리기. 컬럼은 가산이라 남아도 무해하다.

### F.7 M6 — `moai web` 관계 그래프 보기 (Priority Medium)

- **목적**: `/todo?view=graph`(REQ-TCI-023~024).
- **파일**: `internal/web/todo_queue_read.go`(`readTodoQueue` 옆에 노드·간선을 읽는 이음매 — 보관 카드·관계·`gtd_relations` 를 `LoadPure`/읽기 전용 소스로), `todo_view.go`(VM 확장), `screens.templ`(그래프 패널, 생성물 `screens_templ.go`), `assets.go`(임베드 목록에 새 자산 이름), `assets/`(새 CSS·SVG 필요 시), `assets/i18n.js`(키), 테스트.
- **RED 테스트**: `TestTodoGraphViewRendersRelations`, `TestTodoGraphViewReadOnly`, `TestTodoGraphViewBounded`, `TestTodoGraphAssetsEmbedded`, 회귀 `TestTodoPageUnchangedWithoutViewParam`(M6 시작 시 변경 전 렌더러에 대해 먼저 쓴다 — 도착 시 초록인 가드).
- **구현 요점**: 서버 렌더 Templ, GET 전용, 쓰기·락 없음, 외부 요청 없음, 노드 수 상한(값은 M0 측정에서 — 관계에 이름이 오른 카드 수와 열린 카드 수가 하한 근거), 결정적 레이아웃(같은 입력 같은 출력). 새 문자열은 모든 로케일에 둔다.
- **검증 명령**: `SCRUB go test ./internal/web -run '^TestTodoGraphViewRendersRelations$' -count=1 -v` 등, i18n 거버넌스 테스트, 저장소의 templ 생성 대상(`make templ-generate`, Makefile 의 `templ-generate`)으로 생성물을 다시 만든 뒤 생성물이 커밋에 포함됐는지 확인.
- **종료 증거**: AC-TCI-022~023. **되돌리기**: 커밋 되돌리기(자산·키 포함).

### F.8 M5 — 규칙 개정 (Priority High, 게이트됨)

**게이트 명령.** M5 는 카드 브랜치가 develop 을 흡수한 뒤, 아래 세 명령을 **한 번씩 따로** 읽는다(각각 단일 호출, 파이프 없음).

1. `git rev-list --first-parent --count -E -i --grep='^merge[( :]+(card )?t1453' HEAD` — 대상. 1 이상이면 열림.
2. `git rev-list --first-parent --count -E -i --grep='^merge[( :]+(card )?t1448' HEAD` — 양성 대조군. 핀에서 1.
3. `git rev-list --first-parent --count -E -i --grep='^merge[( :]+(card )?t1344' HEAD` — 두 번째 양성 대조군. 핀에서 1.

대조군 중 하나라도 0 이면 선택자가 눈이 먼 것이므로 게이트는 "미측정 = 미충족"이고 보고서에 그렇게 적는다. 세 명령은 모두 종료 코드 0 을 내므로 **출력된 수를 읽어야** 한다(수를 읽지 않고 종료 코드만 보는 것이 이 게이트의 함정이다). 핀에서의 값: 대상 0, 대조군 1·1. 게이트는 M5 시작 직전과 M5 첫 커밋 직전에 다시 읽는다(탐침은 썩는다).

**Mode A (게이트 열림, REQ-TCI-021).**

1. 시작 측정: 카드 브랜치 기준 `git merge-base develop HEAD` 를 읽는 시점에 다시 구하고(핀하지 않는다), 대상 파일 여섯과 템플릿 사본의 크기(문자·바이트)를 다시 잰다. 그 사이 t1450·t1452·t1359·t1453 이 바꾼 것이 있는지 `git diff --name-only <그 SHA> -- <여섯 파일>` 로 읽는다.
2. 배치: 새 규칙은 `.claude/rules/moai/workflow/card-issuance.md`(경로 한정, 상시 로드 아님; `paths:` 는 `**/.claude/skills/moai/workflows/gtd.md,**/.claude/agents/moai/manager-todo.md` 로 한정하고 `kanban-dispatch*` 글롭과 자기 매칭하는 이름을 피한다 — 이름 함정은 SPEC-INSTRUCTION-BUDGET-SCOPE-001 이 기록했다)에 둔다. 섹션: 카드 크기, 후속 지적 규칙, 파생 깊이, 동시 진행 한도, 발행 체크리스트, 부채 대장. 수치는 배포 규칙에 쓰지 않고 로컬 전용 규칙(`.claude/rules/local/` 의 새 파일)에 두거나 기준선 기록을 가리킨다(D7 선택지 (iii); 작업 기본값은 (ii) 기록을 가리킴).
3. 상시 로드 `kanban-dispatch.md` 는 한 문장 포인터를 기존 문장을 압축해 얻은 자리로 넣는다(순증가 0자·0바이트). `gtd.md` 는 병합 독트린 두 문장(`:63` 행과 `:114` 문장)을 개정하고 40,000자 이하로 맞춘다 — 그 압축이 내용의 이동이면 도달성(스킬 본문이 경로 트리거 없이 적재되는 경우)을 확인한다(`rule-loading-budget.md`: 도달성 확인 없는 바이트 절감은 성능 결과가 아니다). `sync-auditor.md` 에는 후속 지적 규칙과 부채 대장 문장을, `manager-todo.md` 에는 후속 지적·파생 깊이 문장을 더한다.
4. 미러: 바이트 동일 쌍(`gtd.md`, `kanban-dispatch-detail.md`, `kanban-dispatch-mechanics.md`, `manager-todo.md`)은 같은 편집으로 동일하게 유지한다. **분기된 두 쌍**(`kanban-dispatch.md`: 로컬에만 있는 `moai worktree sweep` 한 문장, `sync-auditor.md`: 18줄 차이)은 분기를 흡수하지 않고 **보존**한다 — 편집은 두 사본에 같은 텍스트로 하고, 분기 hunk 수가 편집 전과 같음을 `diff` 로 보인다(AUTO-PICK 의 선례). 두 쌍을 `declaredForkedPairs` 에 등록할지는 이 SPEC 이 정하지 않는다(범위 밖; 후속 항목으로 보고).
5. 생성물: `sync-auditor.md`·`manager-todo.md` 를 고쳤으면 `make agents-emit` 으로 `.codex/agents/moai/*.toml`(저장소 루트와 템플릿 하위)을 다시 만들고, 템플릿 아티팩트를 고친 뒤 마지막에 `go run ./internal/template/scripts/gen-catalog-hashes.go --all` 로 `catalog.yaml` 해시를 다시 만들어 같은 커밋에 스테이징한다.
6. 가드: `go test ./internal/template -run '^TestManifestHashFormat$'`, `'^TestCatalogHashCoversSkillSubfiles$'`, `'^TestDeclaredRuleMirrorForks$'`, 카드 id 누출 테스트, `workflow_rule_paths_pinned_test.go`, `todo_skill_doc_parity_test.go` 의 `TestTodoSkillDocumentsJevSource`, `./internal/cli` 의 `TestAutoRank*` 계열과 `TestTodoSkillDocumentsClassification`, `./internal/spec` 의 `TestACCounterFullCorpusMatchesBaseline`, `jev_auto_exception_test.go`. 각각 `-list` 로 선택된 수를 확인한다.
7. 커밋 본문: 상시 로드 파일의 측정 전후 바이트·문자와, 이 규칙이 필요 없는 세션이 치르는 비용(`rule-authoring.md` (c))을 적는다.

**Mode B (게이트 닫힘, REQ-TCI-022).** 여섯 규칙 파일과 템플릿 사본은 편집하지 않는다. `.moai/reports/t1454/m5-draft/` 에 각 규칙 문장과 **삽입 위치 앵커**(파일, 앵커 문장, 앞/뒤)를 담은 초안을 쓰고, `progress.md` §E.2 에 게이트 세 명령의 출력과 리더가 발행할 후속 카드 문안을 적는다. 이 경우 완료 보고는 "요구는 충족했으나 규칙은 아직 효력이 없다"를 명시한다. D1 이 SPEC 을 in-progress 로 붙들라고 정하면 sync 는 M5 적용 뒤에 한다.

### F.9 위험과 완화

| # | 위험 | 마일스톤 | 완화 |
|---|---|---|---|
| R1 | 어휘 이웃이 소음이거나 재현율이 낮아 알림이 무시된다 | M1 | 점수·척도를 항상 표기, 하한 아래는 `--dry-run` 전용(D8), 확실한 일치를 먼저 보임; 하한은 M0 재측정으로 확정 |
| R2 | 제시가 `add` 를 느리게 한다(git 탐침 0.25초/브랜치) | M1 | 락 밖·시간 상한·`unmeasured (time bound)`; 탐침 수 상한은 M0 가 잰 진행 중 레인 수에서 |
| R3 | 새 컬럼이 동결 튜플·JSON 골든·보관/복원·큐 병합을 깬다 | M2 | JSON 컬럼 하나, `omitempty`, 동결 테스트를 같은 커밋에서 갱신, 큐 병합 복사 경로 RED 테스트 |
| R4 | retrofit 순서 오류로 v1→v2 재구성이 새 컬럼을 떨어뜨린다 | M2 | `ensureSchema` 의 마지막 가산 패스, `backlogItemsTableColumns` 에 넣지 않음(카드 t1310 순서) |
| R5 | 저장소 통합이 보관·복원 의미를 흔든다 | M3 | 저장 행 불변, 읽기 해석기만(D2 기본값), 읽기 시 매핑 |
| R6 | git 증거 간선이 `graph build` 를 느리게 하거나 비결정적이 된다 | M3 | 병합 SHA 키 증분 캐시 후보, 정렬·시각 없음, 결정성 테스트, 비용은 M3 첫 측정 |
| R7 | 묶음이 임대 경로를 깨거나 직렬 슬롯 의미를 바꾼다 | M4 | 선택 호 변경을 묶음 예약·선행 미병합 건너뛰기로 한정, 비묶음 카드 골든 테스트, A3 개정 행 |
| R8 | 병합 동사가 "분석은 접지 않는다" 불변식을 우회로로 연다 | M4 | 레인·분석·`relate` 에서 호출 불가를 테스트로 강제, 운영자 전용 |
| R9 | M5 압축이 내용을 경로 한정 파일로 옮겨 도달 불가를 만든다 | M5 | 도달성 확인, 상시 로드 순증 0 |
| R10 | t1453 이 같은 파일을 고쳐 충돌한다 | M5 | 게이트가 선행, 시작 직전 재측정과 diff 확인, Mode B |
| R11 | 웹 그래프가 큰 큐에서 무겁다 | M6 | 노드 상한(M0 측정 근거), 결정적 레이아웃, 서버 렌더 |
| R12 | 템플릿 가드(카드 id·중립성·카탈로그 해시)가 늦게 터진다 | M5 | 편집마다 가드 실행, 카탈로그 재생성을 같은 커밋에 |
| R13 | 분기된 미러 쌍이 흡수돼 의도치 않은 로컬 문장이 배포된다 | M5 | 분기 보존, hunk 수 증명 |

### F.10 @MX 태그 계획 (`mx_plan`)

`mx_plan`
- **ANCHOR**(fan_in ≥ 3 예상): (1) 관계 어휘 정규화·제약 검증 함수(`relate`·`merge`·`trace`·웹 읽기 이음매·그래프 층이 부른다), (2) 읽기 전용 이웃 조회 함수(`add`·`--dry-run`·MCP·engage 가 부른다), (3) `WaitsOnOf`(기존 앵커 유지, 새 호출자를 앵커의 `@MX:REASON` fan_in 수에 반영). 앵커마다 `@MX:REASON` 필수.
- **WARN**: (1) git 탐침을 동시에 돌리는 경우(고루틴이 있다면) — 시간 상한과 취소 맥락 필수, (2) 선택 호의 묶음 예약 분기(복잡도 증가 시), (3) 그래프 층의 git 증거 추출(순회 복잡도).
- **NOTE**: 표시 하한·이웃 상한·구성요소 깊이·허브 임계값·탐침 시간 상한 상수(각 상수에 근거 측정 id 를 적는다), 어휘 매핑 표, drop 접두사 규약(`[DROPPED — ` 형태).
- **TODO**: 처분이 `list` 표시에 반영되는지(D11 이 정하면), 성능 측정이 끝나지 않은 탐침 병렬화.
- **DEBT**: `closed_at` 접근자의 "unknown" 폴백(`@MX:CEILING` 소급 불가, `@MX:UPGRADE` D9 가 컬럼을 고르면). 태그 문구 언어는 `language.yaml` 의 `code_comments`(en)를 따른다.

### F.11 작업 기본값 요약 (계획이 진행하는 값, 판정이 아님)

모든 수치는 `research.md` §3 의 측정에서 왔다. **M0 재측정이 이 값을 다시 도출하며, 기준과 규칙은 이 표가 아니라 기준선 기록을 가리킨다.**

| 항목 | 작업 값 | 근거 측정 |
|---|---|---|
| 유사 카드 상한 | 3 | 카드 본문이 정한 값(고정) |
| 표시 하한(token-set Jaccard) | 0.30 — 추가의 8.4%(최근 4.9%)에 알림, 기록된 관계 재현율 6% | SB03, SB05 |
| 구성요소 깊이 | 경로 앞 두 마디 — 깊이 3 은 공유 쌍 0 | SB07 |
| 카드 크기 하한 트리거 | 예상 제품 줄 50 미만이면서 제품 파일 3개 이하(착지 카드의 29%)는 같은 주 파일의 열린 카드와 묶거나 이유를 적는다 | GB09, GB13 |
| 카드 크기 상한 | 제품 줄 1,604 초과 또는 제품 파일 21 초과(착지 카드의 상위 10%) | GB03, GB04 |
| 파생 깊이 상한 | 2 — depth ≥ 3 은 71장(6.0%) | QB06 |
| 동시 진행 한도 | 16 — 한 3시간 구간에 커밋한 서로 다른 카드의 P90 대리 지표(중앙 5, 최대 41) | QB10(스냅숏 시점 picked 16) |
| 허브 임계값 | 72시간 창 최대 겹침 순위의 상위(`catalog.yaml` 20, `defaults.go` 12, `todo.go` 10, `kanban-dispatch.md` 10) — 정확한 컷은 M0 | GB15 |
| git 탐침 시간 상한 | M1 이 진행 중 레인 수와 탐침 지연 분포를 먼저 재서 정한다 | SB08 |
| 웹 그래프 노드 상한 | 관계에 이름이 오른 카드 176장과 열린 카드 37장의 합집합 규모에서 M6 가 정한다 | QB04, QB02 |

결정 열다섯 개의 작업 기본값은 `spec.md` §B 표, 선택지와 근거는 `design.md` §11, 열린 질문은 `decision-index.md`.

### F.12 커밋 구조

1. M0 기준선(코드 변경 없음) — 가장 먼저.
2. 마일스톤마다 RED 시험 커밋, GREEN 커밋(필요하면 리팩터링 커밋). 모든 커밋 메시지에 카드 id(`t1454`)를 넣는다(브랜치 이름이 카드를 식별하지 않는다).
3. M5 Mode A 는 규칙·템플릿 사본·생성 Codex TOML·`catalog.yaml` 을 한 커밋으로(생성기 출력 상한).
4. sync 단계: 대상 SPEC 개정 행(A1~A6 중 적용분)은 각 대상 SPEC 별 커밋으로(manager-spec 재위임, 종결 커밋 제목은 전체 SPEC id 하나).

## §G 안티패턴

- 어휘 이웃이 의미 중복 판정을 대체한다고 약속하는 것. 표시는 "참고용"과 "확실함"을 구분한다.
- 진행 중 레인 겹침 입력이 비었는데 `none` 을 찍는 것. 읽지 못한 입력은 항상 `unmeasured` 다(REQ-TAU-013 의 규율).
- 제시 계산을 `Mutate` 안에서 하거나 제시 실패를 admit 실패로 바꾸는 것.
- 게이트 명령의 종료 코드만 읽고 출력된 수를 읽지 않는 것(0 을 내도 "열림"이 아니다).
- 갈라진 미러 쌍을 "고치는 김에" 바이트 동일로 합치는 것(AUTO-PICK 의 보존 선례).
- 규칙 파일에 수치를 SPEC 에서 복사해 붙이는 것 — 값은 기준선 기록에서 읽는다.
- `-run` 에 다중 선택자를 `^(A|B)$` 로 쓰는 것(lint 가 권하는 형태가 선택을 비운다) — 가지마다 `^TestName$` 를 쓴다.

## §H 상호참조

`spec.md`(요구·개정 행), `acceptance.md`(기준·RED 원장·변이 탐침), `design.md`(구조·표·선택지), `research.md`(측정·재독 인용·공백), `decision-index.md`, `progress.md`. 규칙: `verification-completeness.md`, `verification-claim-integrity.md`, `spec-frontmatter-schema.md`, `rule-authoring.md`, `rule-loading-budget.md`, `sprint-round-naming.md`, `kanban-dispatch.md`, `gitflow-lane-protocol.md`.
