# SPEC-DEPLOY-STEP-OPTIN-001 — 인수 기준

## §A 관측 규율 [HARD]

1. **문서 단위 트리 고정.** 이 문서의 모든 셀은 트리 SHA `8249bff8c3933d647b4f99e59048595f9808e83a`(짧은 형식 `8249bff8c3`)에서 측정했다. 셀 단위 고정이 없는 기준은 이 고정을 따른다(verification-completeness §2.1). AC-DSO-006–008 은 t1882 가 착지한 트리에서 다시 측정해 다시 고정한다(plan.md §E).
2. **RED-now 셀의 네 요소.** 명령, 원문 stdout, 종료 코드, 트리 SHA 를 research.md 의 증거 원장(L01–L63)이 한 항목씩 가진다. 이 문서의 RED-now 셀은 원장 번호를 인용한다. 원장이 운반자다. 모든 원장 명령은 단일 호출, 읽기 전용이다.
3. **분류.** `RB` 는 릴리스 차단 기준이다(RED-now 가 현재 트리에서 재실행 가능하고 이유가 적혀 있다). `RG` 는 회귀 가드다(현재 GREEN 이거나 RED-now 가 재실행 불가). 회귀 가드는 통과로 기록하지 않는다.
4. **RED 의 이유.** 모든 RB 셀은 왜 빨간지를 한 줄로 적는다. 그 이유는 이 작업이 건드리는 파일에서만 풀린다(구현 밖 파일을 고쳐야만 풀리는 RED 는 채택하지 않는다).
5. **0건 스윕.** 녹색 경로의 `go test` 는 `-v` 로 실행하고 `--- PASS` 줄 수가 아래 §D 의 기대 개수와 같아야 한다. `[no tests to run]` 은 통과가 아니다.
6. **go test 의 성격.** `go test` 는 빌드 캐시와 임시 디렉터리를 쓰므로 엄격한 의미의 읽기 전용이 아니다. 실행 전후 `git status --short` 가 같아야 한다. 자원 규칙에 따라 변경 패키지로 범위를 좁히고 패키지 전체 `go test ./...` 는 쓰지 않는다.
7. **이 plan 단계에서 실행한 것.** 구현 테스트는 아직 없으므로 녹색 경로는 약속이다. 모든 RED-now 셀은 원장 명령의 실제 출력이다. `go test` 와 `go build` 는 plan 단계에서 실행하지 않았다.

## §B AC 요약표

| AC | 요구 | 분류 | RED-now (원장) | 녹색 경로 |
|---|---|---|---|---|
| AC-DSO-001 | REQ-DSO-001, REQ-DSO-002, REQ-DSO-018 | RB | L47 | M2 |
| AC-DSO-002 | REQ-DSO-002, REQ-DSO-003 | RB | L47 | M2 |
| AC-DSO-003 | REQ-DSO-001, REQ-DSO-004 | RB | L47 | M2 |
| AC-DSO-004 | REQ-DSO-005 | RB | L59, L41 | M2 |
| AC-DSO-005 | REQ-DSO-006 | RG | (RED 불가, 아래) | M3 |
| AC-DSO-006 | REQ-DSO-007 | RB | L03, L08 | M3 |
| AC-DSO-007 | REQ-DSO-008, REQ-DSO-004 | RB | L03, L08 | M3 |
| AC-DSO-008 | REQ-DSO-009 | RB | L03, L05, L08 | M3 |
| AC-DSO-009 | REQ-DSO-010 | RB | L11, L12, L45, L46 | M3 |
| AC-DSO-010 | REQ-DSO-011 | RB | L57, L58 | M3 |
| AC-DSO-011 | REQ-DSO-012 | RB | L52 | M3 |
| AC-DSO-012 | REQ-DSO-013 | RB | L54 | M3 |
| AC-DSO-013 | REQ-DSO-013 | RB | L54, L59 | M3 |
| AC-DSO-014 | REQ-DSO-014 | RG | (현재 GREEN, 아래) | M5 |
| AC-DSO-015 | REQ-DSO-018 | RG | (현재 GREEN, 아래) | M6 |
| AC-DSO-016 | REQ-DSO-015 | RB | L31, L47 | M4 |
| AC-DSO-017 | REQ-DSO-015, REQ-DSO-016 | RB | L47 | M4 |
| AC-DSO-018 | REQ-DSO-015 | RB | L29, L30 | M4 |
| AC-DSO-019 | REQ-DSO-015, REQ-DSO-013 | RB | L47 | M4 |
| AC-DSO-020 | REQ-DSO-016 | RB 일부 + RG 일부 | L47 / L26–L28 | M4 |
| AC-DSO-021 | REQ-DSO-017 | RG | (RED 불가, 아래) | M7 |
| AC-DSO-022 | REQ-DSO-007, REQ-DSO-008, REQ-DSO-009, REQ-DSO-015 | RB | L03, L08, L47 | M4 이후 |

요구 18개, 인수 기준 22개(Tier L 상한 각 25).

## §C AC 상세

각 항목은 Given-When-Then, RED-now 셀(원장 인용과 이유), 녹색 경로 셀(마일스톤과 통과 출력), 변이체 탐침으로 구성된다.

### AC-DSO-001 — 설정이 없는 프로젝트는 일곱 단계가 모두 수동이다 (REQ-DSO-001, REQ-DSO-002, REQ-DSO-018)

- **Given** `.moai` 트리가 없거나, `workflow.yaml` 이 없거나, 파일은 있으나 단계 블록이 없는 프로젝트가 있다.
- **When** 해석기가 프로젝트 루트로 불린다.
- **Then** 일곱 단계가 모두 `manual`, 출처 `default`, 이유 `absent` 다. 새로 `moai init` 한 프로젝트(원격 있음·없음 각각)도 같다.
- **RED-now** L47: 해석기도 단계 키도 비테스트 Go 코드 어디에도 없다(`grep` 종료 코드 1, stdout 비어 있음). 이유: 해석기가 없다. 양성 대조 L14 가 같은 `internal` 트리에서 기존 토큰을 찾는다(스윕 범위가 비어 있지 않다).
- **녹색 경로** M2: `go test -count=1 -v -run '^TestStagePolicy_DefaultsAllManual$' ./internal/config` 가 서브테스트 합계 5개(`no-moai-dir`, `no-workflow-file`, `no-stage-block`, `init-with-remote`, `init-without-remote`)를 `--- PASS` 로 보고하고 `ok` 로 끝난다.
- **변이체 탐침** "파일이 없으면 auto" 변이체는 `no-workflow-file` 에서 빨개진다. "항상 manual" 변이체는 AC-DSO-003 의 양성 사례에서 빨개진다. `init-with-remote` 는 원격 유무로 `auto` 를 고르는 변이체를 잡는다(L21).

### AC-DSO-002 — 누락·잘못된·부분 설정은 해당 단계만 수동이다 (REQ-DSO-002, REQ-DSO-003)

- **Given** 표 형태 입력: 단계 블록 없음, 블록이 매핑이 아님, 단계 값이 null·`true`·`yes`·`1`·`"Auto"`·`"auto "`·빈 문자열·알 수 없는 단어, 알 수 없는 단계 이름, 같은 단계 키가 두 번, YAML 파싱 불가 파일, 읽기 불가 파일, 그리고 유효한 `push: auto` 옆에 잘못된 `merge: yes`.
- **When** 해석기가 불린다.
- **Then** 잘못된 단계는 서로 다른 이유 코드와 함께 `manual` 이고, 유효한 `auto` 는 그대로 `auto` 이며(부분 설정), 매핑이 아니거나 파싱 불가한 경우 일곱 단계가 모두 `manual` 이다. 알 수 없는 단계 이름은 무시되고 이유가 보고된다.
- **RED-now** L47: 해석기가 없다. 이유: 같은 이유로 빨갛다. 현재 이웃 키는 반대로 열린 쪽으로 푼다(L13: 값이 틀려도 `card`, L17: 읽기 실패에도 켜짐).
- **녹색 경로** M2: `TestStagePolicy_FailClosedMatrix` 가 15개 서브테스트 `--- PASS`(위 Given 의 열거 순서대로 15개).
- **변이체 탐침** truthy 문자열(`true`, `yes`, `1`)을 `auto` 로 읽는 변이체, 잘못된 값 하나가 형제 단계의 유효한 `auto` 까지 지우는 변이체(결정 Q5 가 블록 전체 수동으로 정해지면 기대값이 뒤집히며 그때는 이 항목을 Q5 판정에 맞춰 고친다), 파싱 실패에서 패닉하는 변이체가 모두 각자의 서브테스트에서 빨개진다.

### AC-DSO-003 — 단계는 서로 독립이다 (REQ-DSO-001, REQ-DSO-004)

- **Given** 일곱 단계 중 하나만 `auto` 인 설정이 일곱 가지(단계마다 하나).
- **When** 해석기가 불린다.
- **Then** 지정한 단계만 `auto` 이고 나머지 여섯은 `manual` 이다.
- **RED-now** L47. 이유: 해석기가 없다.
- **녹색 경로** M2: `TestStagePolicy_EachStageIndependent` 가 7개 서브테스트 `--- PASS`.
- **변이체 탐침** `pr_creation: auto` 가 `push` 를 함께 여는 변이체(상위 단계 암묵 허용)는 `pr_creation` 서브테스트에서 `push` 가 `auto` 로 나와 빨개진다. "첫 auto 가 이후 모든 단계를 연다"는 변이체도 같다.

### AC-DSO-004 — 출처·override·revision 을 돌려준다 (REQ-DSO-005)

- **Given** (a) 설정 없음, (b) project 계층(`workflow.yaml`)만 `push: auto`, (c) project 가 `push: auto` 이고 local 계층이 `push: manual`, (d) (c) 에서 project 의 무관한 키만 바뀜, (e) (b) 에서 `push` 값이 `manual` 로 바뀜, (f) project 파일이 파싱 불가이고 local 계층이 `push: auto`.
- **When** 해석기가 각 입력으로 불린다.
- **Then** (a) 출처 `default`. (b) 출처 `project`, 값 `auto`. (c) 값 `manual`, 출처 `local-override`, project 값도 함께 반환된다. (d) revision 이 변하지 않는다. (e) revision 이 바뀐다. (f) `push` 는 `auto`(출처 `local-override`), 나머지 여섯은 `manual`(project 쪽 이유 `file-unparseable`).
- **RED-now** L59: 전환·기록·해석 어디에도 policy revision 이 없다(종료 코드 1, 빈 stdout). L41: 기존 8단 해석기의 호출자는 doctor 하나뿐이고 local 경로를 CWD 상대로 읽는다(L40). 이유: 이 작업이 만드는 해석기에만 있는 출력이다.
- **녹색 경로** M2: `TestStagePolicy_SourceAndRevision` 이 6개 서브테스트 `--- PASS`.
- **변이체 탐침** revision 이 상수인 변이체는 (e) 에서, project 계층만으로 계산하는 변이체는 (c) 변형(local 만 바뀜)에서, 단계와 무관한 키까지 해시에 넣는 변이체는 (d) 에서 빨개진다.

### AC-DSO-005 — 단계 키를 읽는 코드는 해석기뿐이다 (REQ-DSO-006)

- **Given** 비테스트 Go 코드 전체.
- **When** 정적 검사가 단계 키 이름과 `Workflow` 구조체의 단계 필드를 읽는 지점을 센다.
- **Then** 해석기 파일 밖의 직접 읽기는 0건이다.
- **분류 RG.** 단계 키가 아직 없으므로 오늘은 0건으로 통과한다(L47, 빈 스윕). 빈 스윕의 통과는 아무것도 단언하지 못하므로 RED-now 를 만들 수 없다. 이 기준은 회귀 가드로만 채택하고 통과로 기록하지 않는다.
- **녹색 경로** M3: `TestStagePolicy_NoDirectReaders` 와 양성 대조 `TestStagePolicy_NoDirectReaders_FlagsSeededReader` 가 `--- PASS` 2개. 후자는 시험 안에서 만든 읽기 지점이 있는 임시 파일을 검사기에 넣어 검사기가 반드시 실패하는 것을 관측한다.
- **변이체 탐침** 검사기가 아무것도 스윕하지 않는 변이체는 양성 대조에서 빨개진다. 호출자가 `cfg.Workflow` 필드를 직접 읽는 변이체는 주 검사에서 빨개진다.

### AC-DSO-006 — push 는 push 단계가 auto 일 때만 한다 (REQ-DSO-007)

- **Given** github-flow, 카드 `merge-ready`, 유효한 lease, `push: manual`. 기존 시험 설비를 쓴다: gh 는 패키지 변수 `factoryGH`(`factory_card_pr.go:50`)를 교체해 호출을 기록하고, 원격은 임시 bare 저장소다(L62, L63). push 는 가짜가 아니라 실제 `git push` 이므로 "push 0회"는 bare 저장소에 카드 브랜치 ref 가 생기지 않았다는 관측으로 확인한다.
- **When** `factoryDeliverByPR` 가 실행된다.
- **Then** bare 저장소에 카드 브랜치 ref 가 없고, 카드 레코드의 상태와 version 이 그대로이며, 오류 메시지에 `push` 단계 이름과 켜는 방법이 있다. `push: auto` 에서는 기존 시험(`TestFactoryCompleteGitHubFlowPR` 계열)이 그대로 통과한다.
- **RED-now** L03: `push` 호출이 `factory_card_pr.go:395` 에서 조건 없이 일어난다. L08: 이 파일에 정책을 읽는 토큰이 없다(종료 코드 1). L09: 유일한 게이트는 워크플로 종류다. 이유: 정책 판정이 없다. t1882 착지 뒤 트리에서 같은 명령을 다시 실행해 다시 고정한다.
- **녹색 경로** M3: `TestFactoryDeliverByPR_PushManualStopsBeforePush` 와 `TestFactoryDeliverByPR_PushAutoStillPushes` 가 `--- PASS`. 같은 L08 명령이 종료 코드 0 과 정책 호출 지점 한 줄 이상을 낸다.
- **변이체 탐침** 정책 확인을 push 뒤에 두는 변이체는 "bare 저장소에 ref 없음" 단언에서 빨개진다. 멈추면서 상태를 `merging` 으로 옮기는 변이체는 상태 불변 단언에서 빨개진다. 정책 확인을 `verifyLease` 앞이 아니라 readiness 앞에 한 번만 두는 변이체는 AC-DSO-012 가 잡는다.

### AC-DSO-007 — PR 생성은 pr_creation 이 auto 일 때만 하고 push 를 대신하지 않는다 (REQ-DSO-008, REQ-DSO-004)

- **Given** (a) `push: auto`, `pr_creation: manual`. (b) `push: manual`, `pr_creation: auto`, 원격에 카드 브랜치가 없다.
- **When** 전달 본체가 실행된다.
- **Then** (a) bare 저장소에 카드 브랜치 ref 가 생기고 `pr create` 호출은 0회이며 카드는 pr-open 으로 가지 않는다. (b) ref 도 `pr create` 도 없고 메시지가 `push` 단계 이름을 댄다.
- **RED-now** L03(`:410` 의 `pr create` 가 조건 없음), L08. 이유: 정책 판정이 없다.
- **녹색 경로** M3: `TestFactoryDeliverByPR_PRManualStopsAfterPush`, `TestFactoryDeliverByPR_PRAutoWithPushManualRefuses` 가 `--- PASS` 2개.
- **변이체 탐침** `pr_creation: auto` 가 push 를 대신 수행하는 변이체는 (b) 에서 bare 저장소에 ref 가 생겨 빨개진다.

### AC-DSO-008 — merge 요청은 merge 가 auto 일 때만 한다 (REQ-DSO-009)

- **Given** `push: auto`, `pr_creation: auto`, `merge: manual`.
- **When** 전달 본체가 PR 을 만든(또는 읽은) 뒤 pr-open 기록 지점에 도달한다.
- **Then** `pr merge` 호출이 0회이고 카드는 `pr-open` 으로 기록된다. `merge: auto` 에서는 요청이 `--match-head-commit` 으로 검증된 tip 에 묶여 한 번 나간다.
- **RED-now** L05: `"--auto"` 병합 요청이 `:444` 에 있다. L03, L08. 이유: 조건 없는 요청이다. t1882 가 요청을 없애거나 자기 키 뒤로 옮기면 이 RED 의 대상이 바뀐다. 그 경우 착지 트리에서 다시 측정하고 요구 문장(병합 요청은 merge 단계가 auto 일 때만)에 맞는 새 RED 를 고정한다(결정 Q11).
- **녹색 경로** M3: `TestFactoryDeliverByPR_MergeManualRequestsNoAutoMerge`, `TestFactoryDeliverByPR_MergeAutoPinsVerifiedTip` 가 `--- PASS` 2개.
- **변이체 탐침** `merge: manual` 에서도 `--auto` 를 보내는 변이체, `merge: auto` 에서 `--match-head-commit` 을 빼는 변이체가 각각 빨개진다.

### AC-DSO-009 — 자가 배차는 dispatch 가 auto 일 때만 한다 (REQ-DSO-010)

- **Given** (a) 설정 없음. (b) `dispatch: auto`. (c) `dispatch: auto` 와 `--no-auto-dispatch`. (d) `dispatch: manual` 인 프로젝트에서 레인 세션이 인자 없이 `moai factory next`, 같은 세션이 `moai factory next --card <id>`.
- **When** 레인 실행 선택(`laneDispatchSelection`), 세션 시작 레인 규칙 선택, `factory next` 가 실행된다.
- **Then** (a) 수동 스탬프·수동 규칙. (b) 자동 스탬프·다음 카드 규칙. (c) 수동. (d) 인자 없는 호출은 `refused dispatch-manual` 한 줄과 종료 코드 4 로 거절되고 `--card` 지명은 허용된다.
- **RED-now** L11: 선택 함수가 `--no-auto-dispatch` 가 없으면 항상 auto 를 돌려준다(`:323`). L12: 기본값이 코드에 박혀 있고 설정 키가 아니다. L45: 인자 없는 `factory next` 는 우선순위 순서로 집는다. L46: 스탬프를 읽는 두 곳은 안내문 선택일 뿐이다. 이유: 설정과 이어진 판정이 없다.
- **녹색 경로** M3: `TestLaneDispatchSelection_FollowsStagePolicy`(3개 서브테스트: (a)(b)(c)), `TestFactoryNext_BareRefusedWhenDispatchManual`, `TestFactoryNext_NominatedAllowedWhenDispatchManual`, `TestSessionStartLaneRule_ManualWhenStageManual` 가 `--- PASS`.
- **변이체 탐침** 스탬프만 바꾸고 `factory next` 를 막지 않는 변이체는 (d) 에서 빨개진다. `--card` 지명까지 막는 변이체도 (d) 의 두 번째 호출에서 빨개진다(결정 Q14 가 달라지면 기대값을 그에 맞춘다).

### AC-DSO-010 — 자동 카드 생성은 todo_creation 이 auto 일 때만 한다 (REQ-DSO-011)

- **Given** (a) `todo_creation` 기본(수동), (b) `todo_creation: auto`.
- **When** 자동 생산자가 `moai todo add --auto-source project-phase14 "<text>"` 를 부르고, 운영자가 `moai todo add "<text>"` 를 부른다.
- **Then** (a) 자동 호출은 `refused todo-creation-manual` 한 줄과 비영 종료 코드로 거절되고 큐가 변하지 않는다. 운영자 호출은 (a)(b) 모두 성공한다. (b) 자동 호출이 성공한다.
- **RED-now** L57: `--auto-source` 표지가 코드에 없다(빈 스윕이 아니다: 양성 대조 L58 이 같은 파일의 `add` 선언을 찾는다). L58: `add` 는 `DisableFlagParsing: true` 라 수동 파싱이다. 이유: 자동 생산자를 구별하는 길이 없다. L13·L15: 키가 없거나 틀리면 카드 발급 쪽으로 풀린다.
- **녹색 경로** M3: `TestTodoAdd_AutoSourceRefusedWhenStageManual`, `TestTodoAdd_AutoSourceAllowedWhenStageAuto`, `TestTodoAdd_OperatorAddIsNeverGated`, `TestProjectPhase14_DocGatesCardOnTodoCreationStage`(스킬 산문이 해석기 읽기 동사를 부른다고 단언) 가 `--- PASS` 4개.
- **변이체 탐침** 운영자 호출까지 거절하는 변이체, 표지가 있어도 통과시키는 변이체, 표지 이름을 조금 달리 쓴 호출을 운영자 호출로 취급하는 변이체(잘못 쓴 표지는 거절한다)가 각각 빨개진다.

### AC-DSO-011 — 리뷰 지적 수리 요청은 review_repair 가 auto 일 때만 한다 (REQ-DSO-012)

- **Given** codex review gate 가 켜져 있고(`enabled: true`) fresh fail 영수증이 있으며 차단 횟수가 C 미만이다. (a) `review_repair: manual`, (b) `review_repair: auto`.
- **When** 턴 종료 게이트가 실행된다.
- **Then** (a) 차단 결정 없이 허용되고 systemMessage 가 `수리 요청 꺼짐` 상태 이름을 댄다. 검토는 실행되고 판정은 영수증에 기록된다. (b) 현재와 같이 한 번 차단한다(SPEC-REVIEW-GATE-REPAIR-001 REQ-RG-003).
- **RED-now** L52: 턴을 잇는 차단 결정이 `codex_review_gate.go:426` 에 있고 단계 정책 판정이 없다. L47: 단계 키가 없다. L23/L53: 템플릿은 게이트를 끄고 이 저장소 배포본은 켠다(로컬 설정). 이유: 차단이 정책과 무관하게 일어난다.
- **녹색 경로** M3: `TestReviewGate_RepairManualAllowsWithStateMessage`, `TestReviewGate_RepairAutoBlocksAsBefore` 가 `--- PASS` 2개.
- **변이체 탐침** manual 에서 검토 자체를 건너뛰는 변이체(검토는 실행돼야 한다), 메시지 없이 허용하는 변이체(SPEC-REVIEW-GATE-REPAIR-001 REQ-RG-007 위반), auto 에서 차단을 없애는 변이체가 각각 빨개진다.

### AC-DSO-012 — 끄면 다음 경계에서 멈추고 진행 중 작업은 죽이지 않는다 (REQ-DSO-013)

- **Given** `push: auto`, `pr_creation: auto` 로 시작한 전달 본체가 push 를 마쳤고, PR 생성 직전에(`pr view` 가 "PR 없음"을 답하는 시점에) 교체한 `factoryGH` 가 `workflow.yaml` 의 `pr_creation` 을 `manual` 로 바꿔 쓴다.
- **When** 본체가 PR 생성 경계에 도달한다.
- **Then** `pr create` 가 0회이고, 이미 일어난 push 는 되돌려지지 않으며(원격에 남음), lease·워크트리·증거 파일은 그대로이고, 카드는 `merging` 에 머문다. 같은 카드로 `complete` 를 다시 부르고 정책이 다시 `auto` 이면 push 를 반복하지 않고 PR 생성부터 이어간다.
- **RED-now** L54: 네 번의 원격 변경 앞 재확인은 lease·version·만료뿐이고 정책을 다시 읽지 않는다. 이유: 경계마다 다시 판정하는 지점이 없다.
- **녹색 경로** M3: `TestFactoryDeliverByPR_FlipToManualBetweenStepsStopsAtNextBoundary` 가 `--- PASS`.
- **변이체 탐침** 정책을 본체 시작에서 한 번만 읽는 변이체는 중간에 바뀐 값을 못 봐서 빨개진다. 끄면서 lease 나 워크트리를 지우는 변이체는 보존 단언에서 빨개진다.

### AC-DSO-013 — 켜는 일은 지난 경계를 되살리지 않고 판정 revision 이 기록된다 (REQ-DSO-013)

- **Given** (a) 이미 `pr-open` 인 카드, (b) `merge` 가 `manual` 에서 `auto` 로 바뀐다.
- **When** (a) 카드에 `complete` 가 다시 불린다.
- **Then** push·PR 생성은 반복되지 않는다. 병합 요청은 이 호출이 병합 경계에 처음 도달한 것으로 판정될 때만 나간다(경계 통과 여부는 레코드 상태로 판정). 전환 기록마다 그 전환을 허가한 판정의 revision 이 남는다.
- **RED-now** L59: 전환 기록에 revision 이 없다(빈 스윕이 아니다: L54 의 같은 트리 검색이 호출 지점을 찾는다). L54. 이유: 판정과 전환이 이어져 있지 않다.
- **녹색 경로** M3: `TestFactoryDeliverByPR_FlipToAutoDoesNotReplayCompletedSteps`, `TestFactoryTransition_RecordsAuthorizingPolicyRevision` 가 `--- PASS` 2개.
- **변이체 탐침** 정책이 `auto` 가 되면 처음부터 다시 도는 변이체, revision 을 기록하지 않는 변이체가 각각 빨개진다.

### AC-DSO-014 — update 는 명시 설정을 덮어쓰지 않는다 (REQ-DSO-014)

- **Given** 낡은 `workflow.yaml`: (a) 단계 블록 없음, (b) `push: auto` 와 같은 줄 주석·알 수 없는 키·`manual` 로 명시한 단계, (c) project 와 local 계층에 서로 다른 값. 새 템플릿은 주석 블록만 싣는다.
- **When** update 병합과 배치 복원이 실행된다.
- **Then** (a) 병합 뒤에도 일곱 단계가 모두 `manual`, 출처 `default`. (b) 값·주석·알 수 없는 키·명시 `manual` 이 그대로이며 출처 `project`. (c) local 계층 파일이 바이트 단위로 같고 값은 `local-override`.
- **분류 RG.** 현재 트리에서 이미 GREEN 이다. 템플릿에 없는 사용자 키는 모두 보존된다(L34). 이 기준은 새 작업이 만들어내는 동작이 아니라 새 작업이 깨뜨리면 안 되는 동작이므로 RED-now 가 없고 통과로 기록하지 않는다.
- **녹색 경로** M5: `TestUpdate_KeepsExplicitStageValuesCommentsUnknownKeys`, `TestUpdate_AbsentStageBlockStaysManual`, `TestUpdate_LocalTierUntouched` 가 `--- PASS` 3개.
- **변이체 탐침** 템플릿이 단계 값을 싣고 base 와 같은 사용자 값을 새 값으로 바꾸는 변이체(L33 의 규칙이 그대로 작동하는 경우)는 (b) 의 명시 `manual` 항목을 다른 값으로 바꾼 템플릿 변형에서 빨개진다. 이 변이체는 AC-DSO-015 도 함께 깨뜨린다.

### AC-DSO-015 — 템플릿은 단계 값을 싣지 않고 사본은 일치한다 (REQ-DSO-018)

- **Given** 템플릿의 `workflow.yaml`·git-strategy 템플릿·변경된 스킬/규칙/문서 파일과 그 프로젝트 사본.
- **When** 정적 시험과 diff 가 실행된다.
- **Then** (a) 템플릿 워크플로 파일의 단계 블록은 모두 주석이다(단계 값 0개). (b) 해석기 코드 기본값이 일곱 단계 모두 `manual` 이다. (c) `.claude/`·`.agents/`·`.codex/` 아래 변경된 템플릿 파일은 프로젝트 사본과 바이트 단위로 같다(`diff -q` 종료 코드 0). `.moai/config/sections/*.yaml` 은 운영자 자신의 파일이라 다르며(L56) 대신 주석 블록의 존재(헤더 줄 1회)를 양쪽에서 확인한다.
- **분류 RG.** (a) 는 오늘 이미 참이다(L48 빈 스윕, 단계 키 0). (c) 는 변경이 없는 지금 참이다(L55). RED-now 를 만들 수 없는 전방 가드이므로 통과로 기록하지 않는다.
- **녹색 경로** M2(a)(b), M6(c): `TestTemplateWorkflowShipsNoStageValues`, `TestStagePolicy_CodeDefaultsAllManual`, 그리고 변경 파일 목록에 대한 `diff -q` 단일 호출이 종료 코드 0.
- **변이체 탐침** 템플릿이 `push: manual` 한 줄을 싣는 변이체, 코드 기본값 하나가 `auto` 인 변이체, 사본 한 파일에 공백 하나를 더한 변이체가 각각 빨개진다.

### AC-DSO-016 — 웹이 단계별 효력·출처·override·영향·revision 을 보여 준다 (REQ-DSO-015)

- **Given** project 가 `push: auto`, local 계층이 `push: manual`, 나머지 단계는 기본.
- **When** `GET /settings` 가 루프백 Host 로 불린다.
- **Then** 일곱 단계 행이 각각 효력(`auto`/`manual`), 출처 라벨(기본/프로젝트/로컬 override), override 표시(project 값과 효력이 다르면 표시), 영향 설명 문구, 페이지 하나의 정책 revision 을 보여 준다. 읽기 모델이 같은 해석기 출력을 쓴다.
- **RED-now** L31: 웹 읽기 모델은 문자열 스칼라 맵만 돌려주고 출처·override 가 없다. L47: 단계 필드가 스키마에 없다. 이유: 보여 줄 모델이 없다.
- **녹색 경로** M4: `TestStageFields_SchemaCarriesSevenStages`(settings), `TestSettingsPage_ShowsStageEffectiveSourceOverrideImpact`(web) 가 `--- PASS` 2개.
- **변이체 탐침** 디스크 스칼라를 그대로 보여 줘 override 를 숨기는 변이체는 효력 `manual` 단언에서 빨개진다. 영향 설명이 비어 있는 변이체도 빨개진다.

### AC-DSO-017 — 저장 전 검증이 잘못된 입력을 원자적으로 거부한다 (REQ-DSO-015, REQ-DSO-016)

- **Given** 저장 폼: (a) 알 수 없는 값(`maybe`), (b) 같은 단계 필드를 두 번 보냄, (c) 알 수 없는 단계 이름, (d) 유효한 `push: auto` 와 잘못된 `merge: maybe` 를 함께 보냄.
- **When** `POST /save` 가 보내진다.
- **Then** 모든 경우에 디스크가 바뀌지 않고(파일 바이트 동일), 응답은 필드별 오류를 보이며, (d) 에서는 유효한 `push: auto` 도 저장되지 않는다(저장은 전부 또는 전무).
- **RED-now** L47: 단계 필드가 없다. L28 은 검증 실패가 쓰기를 멈춘다는 기존 동작을 보이며 이 기준이 그 위에 쌓인다. 이유: 단계 값의 검증 규칙이 없다.
- **녹색 경로** M4: `TestSave_RejectsInvalidStageValueAtomically`, `TestSave_UnknownOrDuplicateStageRejected` 가 `--- PASS`(서브테스트 합계 4).
- **변이체 탐침** 유효한 필드만 먼저 쓰고 잘못된 필드를 거부하는 변이체는 (d) 에서 파일이 바뀌어 빨개진다.

### AC-DSO-018 — 저장 성공은 재조회한 정책과 revision 을 응답하고 감사 기록을 남긴다 (REQ-DSO-015)

- **Given** 정상 저장 폼: `push: auto` 로 변경.
- **When** `POST /save` 가 처리된다.
- **Then** 응답이 디스크에서 다시 읽은 일곱 단계의 효력·출처와 새 revision 을 담는다(요청에서 되돌려 준 값이 아니다). 감사 기록 한 줄이 추가된다: 시각, 채널(`web`), 호출 주체, 단계별 변경 전후 값, 결과(`saved`/`rejected`), 새 revision. 거부된 저장도 `rejected` 와 이유로 기록된다.
- **RED-now** L29: revision 개념이 없다. L30: 저장 감사 기록을 쓰는 코드가 없다. 이유: 두 가지 모두 이 작업이 만든다.
- **녹색 경로** M4: `TestSave_ResponseCarriesReReadPolicyAndRevision`, `TestSave_WritesAuditRecord`, `TestSave_RejectedSaveIsAudited` 가 `--- PASS` 3개.
- **변이체 탐침** 요청값을 그대로 응답하는 변이체(디스크 재조회 아님)는 저장 직후 외부에서 파일을 바꾼 시험 변형에서 빨개진다. 변경 전 값을 기록하지 않는 변이체, 거부를 기록하지 않는 변이체가 각각 빨개진다.

### AC-DSO-019 — 즉시 중지는 모든 단계를 수동으로 쓰고 진행 중 작업은 건드리지 않는다 (REQ-DSO-015, REQ-DSO-013)

- **Given** 여러 단계가 `auto` 이고 전달 본체가 실행 중이다(시험 이음매).
- **When** 웹의 stop-all 동작이 실행된다.
- **Then** project 계층의 일곱 단계가 한 번의 원자적 쓰기로 `manual` 이 되고 감사 기록 한 줄이 남는다. 실행 중이던 본체는 죽지 않고 다음 경계에서 멈춘다(AC-DSO-012). local override 가 `auto` 인 단계는 효력이 여전히 `auto` 이며 화면이 그 사실을 알린다.
- **RED-now** L47. 이유: stop-all 동작이 없다.
- **녹색 경로** M4: `TestStopAll_SetsAllManualAndAudits`, `TestStopAll_LocalOverrideStillWinsAndIsDisclosed` 가 `--- PASS` 2개.
- **변이체 탐침** 일곱 중 일부만 쓰는 변이체, 쓰기를 여러 번 나눠 중간 상태가 보이는 변이체, 실행 중 프로세스를 종료하는 변이체가 각각 빨개진다.

### AC-DSO-020 — 웹 저장의 안전 경계 (REQ-DSO-016)

- **Given** 단계 저장 폼.
- **When** (a) 비루프백 Host, (b) `Sec-Fetch-Site: cross-site`, (c) local override 가 고정한 단계를 project 에서 바꾸려는 저장, (d) 동시에 두 저장, (e) local 계층 파일을 쓰려는 요청이 도착한다.
- **Then** (a)(b) 는 403 으로 거부되고 디스크 불변. (c) 저장은 project 값만 쓰고 효력은 local 값 그대로이며 응답이 그 사실(`overridden-by-local`)을 알린다. (d) 두 저장이 직렬화되어 하나의 최종 상태와 두 개의 감사 기록이 남고 어느 쪽도 다른 쪽의 쓰기를 되돌리지 않는다. (e) 웹에는 local 계층을 쓰는 경로가 없다.
- **분류** (a)(b)(d) 는 기존 동작의 회귀 가드다(L26–L28: 게이트와 직렬화가 이미 있고 `host_gate_test.go`, `save_serialization_test.go` 가 있다). (c)(e) 는 RB 이며 RED-now 는 L47(단계 필드가 없고 local 계층 개념이 웹에 없다, L41·L42).
- **녹색 경로** M4: `TestSave_NonLoopbackHostDeniedForStageForm`, `TestSave_CrossSiteDeniedForStageForm`, `TestSave_ProjectWriteUnderLocalOverrideIsDisclosed`, `TestSave_ConcurrentStageSavesSerialized`, `TestSave_CannotWriteLocalTier` 가 `--- PASS` 5개.
- **변이체 탐침** 단계 저장을 게이트 앞 경로로 노출하는 변이체, local 값을 덮어쓰는 변이체, 직렬화 잠금을 빼는 변이체가 각각 빨개진다.

### AC-DSO-021 — release/deploy 단계는 선언되고 새 게시 호출은 게이트를 거친다 (REQ-DSO-017)

- **Given** 비테스트 Go 코드 전체.
- **When** 정적 검사가 `gh release create|upload|edit`, 태그 생성(`git tag -a|-s`), 배포 도구 호출 지점을 센다.
- **Then** 모든 지점이 해석기의 `release_deploy` 확인 호출과 같은 함수 안에 있다. 현재는 지점이 0개다.
- **분류 RG.** 오늘의 배포본 Go 코드에 게시 호출이 없으므로(L51, 주석 두 줄만) 검사는 빈 스윕으로 통과한다. RED-now 를 만들 수 없다. 통과로 기록하지 않는다.
- **녹색 경로** M7: `TestNoUngatedPublishCallSites` 와 양성 대조 `TestNoUngatedPublishCallSites_FlagsSeededCallSite` 가 `--- PASS` 2개. 결정 Q4 가 첫 조각 제외로 정해지면 이 기준과 M7 은 후속 SPEC 으로 옮긴다.
- **변이체 탐침** 검사기가 호출 패턴을 못 찾는 변이체는 양성 대조에서 빨개진다.

### AC-DSO-022 — 웹 저장에서 런타임 효력까지 끝에서 끝으로 이어진다 (REQ-DSO-007, REQ-DSO-008, REQ-DSO-009, REQ-DSO-015)

- **Given** 임시 bare 원격과 기존 `factoryGH` 이음매(L62, L63), 빈 프로젝트.
- **When** 웹 저장으로 `push: auto`, `pr_creation: auto` 만 켜고(`merge` 는 기본), 같은 프로젝트에서 전달 본체를 실행한다.
- **Then** 디스크에 두 단계가 `auto` 로 쓰이고, 런타임 해석기가 같은 값을 돌려주며, bare 저장소에 카드 브랜치 ref 가 하나 생기고 gh 호출 기록에 `pr create` 1회가 있으며 `pr merge`·태그·배포 호출은 0회다. 웹 응답의 revision 이 런타임 판정의 revision 과 같다.
- **RED-now** L03·L08·L47. 이유: 웹 저장도 해석기도 단계 정책 분기도 없다.
- **녹색 경로** M4 이후(M3 와 M4 가 모두 끝난 뒤): `TestStagePolicy_WebSaveToRuntimeEndToEnd` 가 `--- PASS`.
- **변이체 탐침** 웹이 다른 키 위치에 쓰는 변이체(런타임이 못 읽음), 런타임이 캐시된 낡은 값을 쓰는 변이체, `pr_creation: auto` 가 merge 요청까지 보내는 변이체가 각각 빨개진다.

## §D 필요한 시험 목록 (이름, 위치, 기준)

| 시험 | 위치 | AC | 서브테스트 수 |
|---|---|---|---|
| `TestStagePolicy_DefaultsAllManual` | `internal/config/stage_policy_test.go` | 001 | 5 |
| `TestStagePolicy_FailClosedMatrix` | 같은 파일 | 002 | 15 |
| `TestStagePolicy_EachStageIndependent` | 같은 파일 | 003 | 7 |
| `TestStagePolicy_SourceAndRevision` | 같은 파일 | 004 | 6 |
| `TestStagePolicy_NoDirectReaders`, `TestStagePolicy_NoDirectReaders_FlagsSeededReader` | 같은 파일 | 005 | 1+1 |
| `TestStagePolicy_CodeDefaultsAllManual` | 같은 파일 | 015 | 7 |
| `TestTemplateWorkflowShipsNoStageValues` | `internal/template/stage_policy_template_test.go` | 015 | 1 |
| `TestFactoryDeliverByPR_PushManualStopsBeforePush`, `..._PushAutoStillPushes` | `internal/cli/factory_card_pr_stage_test.go` | 006 | 1+1 |
| `TestFactoryDeliverByPR_PRManualStopsAfterPush`, `..._PRAutoWithPushManualRefuses` | 같은 파일 | 007 | 1+1 |
| `TestFactoryDeliverByPR_MergeManualRequestsNoAutoMerge`, `..._MergeAutoPinsVerifiedTip` | 같은 파일 | 008 | 1+1 |
| `TestFactoryDeliverByPR_FlipToManualBetweenStepsStopsAtNextBoundary` | 같은 파일 | 012 | 1 |
| `TestFactoryDeliverByPR_FlipToAutoDoesNotReplayCompletedSteps` | 같은 파일 | 013 | 1 |
| `TestFactoryTransition_RecordsAuthorizingPolicyRevision` | `internal/homestate/card_transition_policy_test.go` | 013 | 1 |
| `TestLaneDispatchSelection_FollowsStagePolicy` | `internal/cli/factory_stage_dispatch_test.go` | 009 | 3 |
| `TestFactoryNext_BareRefusedWhenDispatchManual`, `TestFactoryNext_NominatedAllowedWhenDispatchManual` | 같은 파일 | 009 | 1+1 |
| `TestSessionStartLaneRule_ManualWhenStageManual` | `internal/hook/session_start_factory_stage_test.go` | 009 | 1 |
| `TestTodoAdd_AutoSourceRefusedWhenStageManual`, `..._AutoSourceAllowedWhenStageAuto`, `TestTodoAdd_OperatorAddIsNeverGated` | `internal/cli/todo_stage_test.go` | 010 | 1+1+1 |
| `TestProjectPhase14_DocGatesCardOnTodoCreationStage` | `internal/template/project_phase14_stage_test.go` | 010 | 1 |
| `TestReviewGate_RepairManualAllowsWithStateMessage`, `TestReviewGate_RepairAutoBlocksAsBefore` | `internal/cli/codex_review_gate_stage_test.go` | 011 | 1+1 |
| `TestUpdate_KeepsExplicitStageValuesCommentsUnknownKeys`, `TestUpdate_AbsentStageBlockStaysManual`, `TestUpdate_LocalTierUntouched` | `internal/cli/update/backup/stage_policy_merge_test.go` | 014 | 1+1+1 |
| `TestStageFields_SchemaCarriesSevenStages` | `internal/settings/stage_fields_test.go` | 016 | 1 |
| `TestSettingsPage_ShowsStageEffectiveSourceOverrideImpact` | `internal/web/stage_policy_handler_test.go` | 016 | 1 |
| `TestSave_RejectsInvalidStageValueAtomically`, `TestSave_UnknownOrDuplicateStageRejected` | 같은 파일 | 017 | 2+2 |
| `TestSave_ResponseCarriesReReadPolicyAndRevision`, `TestSave_WritesAuditRecord`, `TestSave_RejectedSaveIsAudited` | 같은 파일 | 018 | 1+1+1 |
| `TestStopAll_SetsAllManualAndAudits`, `TestStopAll_LocalOverrideStillWinsAndIsDisclosed` | 같은 파일 | 019 | 1+1 |
| `TestSave_NonLoopbackHostDeniedForStageForm`, `TestSave_CrossSiteDeniedForStageForm`, `TestSave_ProjectWriteUnderLocalOverrideIsDisclosed`, `TestSave_ConcurrentStageSavesSerialized`, `TestSave_CannotWriteLocalTier` | 같은 파일 | 020 | 1+1+1+1+1 |
| `TestNoUngatedPublishCallSites`, `..._FlagsSeededCallSite` | `internal/config/stage_policy_publish_scan_test.go` | 021 | 1+1 |
| `TestStagePolicy_WebSaveToRuntimeEndToEnd` | `internal/cli/stage_policy_e2e_test.go` | 022 | 1 |

최상위 시험 함수는 47개다(위 표의 함수 이름을 하나씩 센 값). 위 표에 없는 시험 이름은 만들지 않는다. 기존 시험은 이 작업이 기대값을 바꾸지 않고 통과해야 한다: `internal/cli/factory_card_pr_test.go`(전체), `internal/cli/factory_autodispatch_test.go`, `internal/config/project_continuation_test.go`, `internal/config/todo_enabled_test.go`, `internal/web/host_gate_test.go`, `internal/settings/write_safety_test.go`, `internal/cli/init_quiet_wizard_test.go`, `internal/cli/update/backup/node_merge_test.go`. 단 `factory_card_pr_test.go` 는 push/PR/merge 가 기본 수동이 된 결과 정책을 `auto` 로 두는 준비 코드를 더해야 하며, 그 기대값 변경은 M3 의 승인 파일 목록에 있다(plan.md §C).

## §E 입력 메모의 후보 기준과의 대응

입력 메모(`.moai/reports/t1895/OPTIN-DESIGN.md`)의 AC-01–07 은 가설 원천이었다. 측정 가능한 형태로 다시 썼고 대응은 다음과 같다.

| 메모 | 이 문서 |
|---|---|
| AC-01 기본 수동·잘못된·부분 설정 | AC-DSO-001, 002, 015 |
| AC-02 단계별 독립 판정 | AC-DSO-003, 006, 007, 008, 009, 010, 011 |
| AC-03 웹이 같은 해석기의 값·출처·override·영향·적용 경계를 보이고 revision·감사 기록 | AC-DSO-004, 016, 018 |
| AC-04 끄면 다음 경계에서 멈추고 진행 중 작업·증거·lease 를 지우지 않음 | AC-DSO-012, 013, 019 |
| AC-05 update 가 명시 값·주석·unknown key·local override 를 보존 | AC-DSO-014 |
| AC-06 비허용 Host/origin, 중복·잘못된 필드, 동시 저장 | AC-DSO-017, 020 |
| AC-07 웹→저장 API→디스크→런타임 유효값, PR 만 허용 시 merge/tag/deploy 0회 | AC-DSO-022 |

메모에 없던 기준: AC-DSO-005(단일 읽기 지점), AC-DSO-021(release/deploy 선언과 정적 검사).

## §F 엣지 케이스와 완료 정의

엣지 케이스: 단계 블록 안의 YAML 앵커·별칭(해석기는 별칭을 펼치지 않고 값이 문자열이 아니면 수동), 한 파일에 같은 키가 두 번(수동, 이유 `duplicate`), 대소문자·공백이 다른 `auto`(수동), 빈 프로젝트 루트 문자열(수동), 동시 읽기 중 파일이 교체됨(원자적 rename 이므로 이전 또는 이후 중 하나를 읽는다), 윈도우 경로 구분자.

품질 게이트: 변경 패키지 단위 커버리지 85% 이상, `gofmt`, `golangci-lint` 변경 패키지 범위, `GOOS=windows GOARCH=amd64 go build` 는 CI 에서(plan 단계에서 실행하지 않았다).

완료 정의: (1) §D 의 모든 시험이 `--- PASS` 로 기대 개수를 채웠고 0건 스윕이 없다. (2) RB 기준마다 원장 RED-now 명령이 녹색 경로의 새 출력으로 바뀐 것을 같은 명령으로 관측했다. (3) 변이체 탐침마다 변이체를 실제로 만들어 빨개지는 것을 한 번 관측했다(progress.md §E.2). (4) 템플릿-사본 비교가 종료 코드 0. (5) 제품 수준 결정 행이 모두 열려 있지 않다(decision-index.md).
