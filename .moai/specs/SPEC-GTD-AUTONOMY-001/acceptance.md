---
id: SPEC-GTD-AUTONOMY-001
status: completed
created: 2026-09-15
updated: 2026-09-15
---

# SPEC-GTD-AUTONOMY-001 수락 기준

## §A. 판정 원칙

- 각 기준은 현재 worktree와 격리된 `MOAI_HOME`에서 관측한 명령·출력·저장 상태로 판정한다.
- exit 0, agent idle, 파일 존재 또는 grep 부재만으로 의미적 PASS를 주장하지 않는다.
- 외부 부작용이 포함된 기준은 side effect 대상의 authoritative readback까지 요구한다.
- 모든 테스트는 요구사항 위반 mutant 또는 명시적 negative fixture를 포함한다.

## §B. Given-When-Then 기준

**AC-GTD-001 [regression-guard].** Given 기존 카드·finding·archive 데이터가 있는 schema-version 1 fixture, When 신규 바이너리가 list/pick/drop/done/undone을 왕복 실행한 뒤 저장소를 다시 연다, Then 기존 DB 경로·카드 ID·순서·상태·finding·보관 결과가 기준 snapshot과 동일하다. 현재 보존 동작이라 재현 가능한 pre-implementation failure가 아니며 release-blocking RED로 분류하지 않는다.

**AC-GTD-002 [release-blocking; RED-GTD-002; GREEN-GTD-002].** Given 동일한 격리 DB snapshot 두 개, When §C의 19개 verb×주요 flag matrix를 `moai gtd`와 `moai todo`로 각각 실행한다, Then 각 셀의 exit status, stdout/stderr 채널, 선택 카드, ID·순서·relation·finding·landing·history와 최종 logical DB state가 같다.

**AC-GTD-003 [release-blocking; RED-GTD-003; GREEN-GTD-003].** Given command·skill·template·completion·4개 언어 문서 생성 fixture, When 배포 생성기를 실행한다, Then GTD canonical 표면과 todo compatibility 표면이 모두 존재하며 호환 표면은 같은 handler 또는 정식 표면으로의 얇은 전달임이 구조 검사와 golden으로 확인된다.

**AC-GTD-004 [regression-guard].** Given GTD 다섯 절차를 모두 사용한 project fixture, When board와 queue state를 조회한다, Then 허용된 개발 단계와 기존 queue state 집합에는 Capture·Clarify·Organize·Reflect·Engage가 추가되지 않는다. 기존 개발 단계·queue enum 보존을 감시하는 기준이라 release-blocking RED로 분류하지 않는다.

**AC-GTD-005 [release-blocking; RED-GTD-005; GREEN-GTD-005].** Given 허용 출처, 미허용 출처, 중복 event, 민감도 불명 입력, When Capture를 실행한다, Then 허용된 고유 입력만 출처·민감도·event identity와 함께 inbox에 저장되고 카드나 실행 약속은 생성되지 않는다.

**AC-GTD-006 [release-blocking; RED-GTD-006; GREEN-GTD-006].** Given 실행 가능·참고·보류·폐기 후보와 목표·완료 조건·권한이 모호한 후보, When Clarify를 실행한다, Then 각 disposition과 근거가 저장되고 모호하거나 비신뢰인 후보의 카드 발행은 거절된다.

**AC-GTD-007 [release-blocking; RED-GTD-007; GREEN-GTD-007].** Given project/action/reference/waiting/scheduled 항목과 정상·누락 대상·순환 관계 fixture, When Organize를 실행한다, Then 정상 분류와 관계는 상태를 보존하고 누락 대상과 `depends_on` cycle은 명시적 이유로 거절된다.

**AC-GTD-008 [release-blocking; RED-GTD-008; GREEN-GTD-008].** Given 완료·취소·재개방·근거 revision 변경·다음 행동 없는 project fixture, When Reflect를 실행한다, Then stale·재검토·blocked successor가 결정적으로 표시되고 취소 항목은 dependency 완료로 취급되지 않는다.

**AC-GTD-009 [release-blocking; RED-GTD-009; GREEN-GTD-009].** Given next, waiting, scheduled, blocked, stale, lane-conflicting 후보, When Engage dry-run과 승인된 실행을 각각 수행한다, Then dry-run은 부작용 없이 결정과 거절 이유만 반환하고 실행은 fresh·unblocked·policy-authorized 후보에만 queue operation을 만든다.

**AC-GTD-010 [release-blocking; RED-GTD-010; GREEN-GTD-010].** Given 기존 schema-version 1 DB, When migration을 처음과 반복해서 실행한다, Then 기존 live/archive 표와 schema-version 의미는 보존되고 GTD metadata와 표만 추가되며 두 번째 실행은 논리 상태를 바꾸지 않는다.

**AC-GTD-011 [release-blocking; RED-GTD-011; GREEN-GTD-011].** Given old→new, new→old→new, backup/restore, export/import, WAL checkpoint, vacuum, archive/reopen matrix, When 각 왕복을 실행한다, Then 기존 카드와 GTD item·relation·policy·operation identity가 명시된 fixture snapshot과 동일하고 손실 fixture는 배포 gate를 FAIL시킨다.

**AC-GTD-012 [release-blocking; RED-GTD-012; GREEN-GTD-012].** Given 신규 관계와 기존 `contains/absorbs/replaces/conflicts`의 양방향 fixture, note bytes, archive/restore snapshot, When 관계 validator·실행 가능성·old/new round-trip을 실행한다, Then 미완료 `depends_on`만 차단하고 참고 관계는 비차단이며 기존 네 관계의 source→target 방향, note bytes, logical meaning, archive와 restore 결과는 변경 전 snapshot과 동일하다.

**AC-GTD-013 [release-blocking; RED-GTD-013; GREEN-GTD-013].** Given 같은 logical revision의 행 순서 변형, privacy 불명, 다른 OS account, repo/template/Git/log/telemetry/export/backup 수집기, opt-in backup, revoke와 item delete fixture, When 홈 DB sibling projection을 생성·조회·수집·삭제한다, Then 디렉터리 `0700`, projection·sidecar·temp `0600`, 허용 local reader만 성공하고 기본 노출면은 모두 0건이며 명시 opt-in backup만 포함되고 revoke는 신규 접근 차단, delete는 다음 atomic rebuild에서 node·edge 제거, privacy 불명은 부분 파일 없이 거절된다.

**AC-GTD-014 [release-blocking; RED-GTD-014; GREEN-GTD-014].** Given 완전한 위임 계약과 목표·완료 증거·금지선 중 하나씩 누락한 계약, When 승인·봉인을 시도한다, Then 완전한 계약만 version과 content hash를 얻고 누락 계약은 실행 가능한 mission이 되지 않는다.

**AC-GTD-015 [release-blocking; RED-GTD-015; GREEN-GTD-015].** Given valid, stale snapshot, scope-expanding, unauthorized, evidence-missing Proposal/Decision, When deterministic validator가 평가한다, Then valid operation만 prepared receipt를 얻고 나머지는 side effect 전에 stable reason code로 거절된다.

**AC-GTD-016 [release-blocking; RED-GTD-016; GREEN-GTD-016].** Given 한국어 자연어와 셸 메타문자가 포함된 `/moai:goal --auto` 임무, When mission을 생성한다, Then 문자는 data로 보존되고 shell 및 기존 condition parser는 호출되지 않으며 저장 상태는 `mission_mode=auto`이고 기존 progression mode 값은 변하지 않는다.

**AC-GTD-017 [release-blocking; RED-GTD-017; GREEN-GTD-017].** Given 승인된 auto mission의 범위 안 선택과 새 권한이 필요한 범위 밖 선택, When supervisor가 각각 다음 행동을 결정한다, Then 범위 안에서는 `AskUserQuestion` 호출 없이 진행하고 범위 밖에서는 후속 부작용 없이 blocked report를 저장한다.

**AC-GTD-018 [release-blocking; RED-GTD-018; GREEN-GTD-018].** Given advisor 권고, governor Decision, auditor FAIL, policy denial이 충돌하는 fixture, When 다음 행동을 판정한다, Then policy denial과 auditor FAIL이 우선하고 advisor는 권한을 얻지 못하며 governor는 파일·큐·git을 직접 변경하지 않는다.

**AC-GTD-019 [release-blocking; RED-GTD-019; GREEN-GTD-019].** Given full(start+reconnect+replace+유효 credential), partial(reconnect 불가 또는 replace 불가), unsupported, credential-expired, lease-loss/takeover runtime fixture, When 시작·재연결·교체·재시작을 각각 수행한다, Then full만 durable로 같은 mission/contract/snapshot hash에서 재개하고 partial은 지원 범위 밖 transition 전에 `active-session-only`, unsupported와 복구 불가 credential은 즉시 `active-session-only`, replace는 이전 owner effect 중단 확인 뒤 실행되며 lease takeover는 하나의 owner만 같은 last-reconciled snapshot lineage를 재개한다.

**AC-GTD-020 [release-blocking; RED-GTD-020; GREEN-GTD-020].** Given 카드 발행·dispatch·commit·local develop merge·batch push·release branch·release PR·main merge의 부작용 성공 직후 receipt 기록 전 종료 fixture, When 동일 mission을 재시작한다, Then authoritative readback으로 이미 적용된 stable operation은 reconciled되고 중복 호출되지 않으며 미적용이 확인된 operation만 같은 ID로 재시도된다.

**AC-GTD-021 [release-blocking; RED-GTD-021; GREEN-GTD-021].** Given 하나의 capture item에 대한 허용·정책 거절·publish 직후 crash·dispatch 직후 crash·두 manager 동시 fixture, When `Capture→Clarify→Organize→sealed delegation→publish→pick→dispatch`를 끝까지 실행한다, Then 허용 흐름은 동일 item/card/operation identity와 순서로 dispatch 하나를 만들고, 거절은 publish 전에 멈추며, restart와 동시 실행은 authoritative readback·UNIQUE key·lane ownership으로 같은 카드의 두 번째 publish/pick/dispatch를 만들지 않는다.

**AC-GTD-022 [release-blocking; RED-GTD-022; GREEN-GTD-022].** Given `local_develop_base_sha`에서 launcher-entered `WT-gtd-autonomy`로 수행한 카드 작업, When delivery를 수행한다, Then lane은 explicit-path commit만 만들고 push/card PR은 0건이며 manager-git만 single local develop integration worktree lease 아래 `--no-ff` merge하여 `card_head_sha`와 `local_develop_merge_sha`를 기록한다.

**AC-GTD-023 [release-blocking; RED-GTD-023; GREEN-GTD-023].** Given local develop merge 여러 개와 stale owner/lease/SHA/audit/CI/review/protection mutants, When lead delivery를 수행한다, Then manager-git의 한 번 batch push 후 `origin_develop_sha` readback과 그 SHA의 CI PASS가 있어야 verified develop에서 `release_head_sha`가 생성되고 release→main PR만 허용되며, 모든 stale mutant와 lane push/card PR은 main merge 전에 거절되고 `main_landed_sha` ancestry가 확인된다.

**AC-GTD-024 [release-blocking; RED-GTD-024; GREEN-GTD-024].** Given 외부 문서에 정책 변경 명령, mission shell metacharacter, advisor의 범위 확대 권고, 도구 출력 속 지시문 fixture, When Clarify부터 executor까지 처리한다, Then 어떤 문자열도 실행되거나 sealed policy를 변경하지 않고 validator 허용 목록 밖 operation은 생성되지 않는다.

**AC-GTD-025 [release-blocking; RED-GTD-025; GREEN-GTD-025].** Given running mission의 사용자 철회, 모든 완료 증거 충족, agent idle, lease 만료 fixture, When supervisor가 종료 상태를 판정한다, Then 철회는 새 작업을 막고 in-flight effect를 조정하며 완료는 authoritative evidence가 모두 충족된 경우에만 기록되고 idle·lease만으로 완료되지 않는다.

## §C. todo ↔ gtd 호환 매트릭스

각 행은 동일한 초기 DB 복제본에서 `todo`와 `gtd`를 실행하여 exit, stdout, stderr 분리, live/archive state, card ID/order, findings, relations, landing, history를 비교한다. `—`는 해당 verb의 고유 flag가 없다는 뜻이며 root 공통 help/error 동작은 전 행에 적용한다.

| Verb | 주요 flag·형태 | 추가 상태 판정 |
|---|---|---|
| add | `--pick`, `--force`, 자연어 fallthrough | ID·position, add+pick 원자성 |
| list | `--json`, `--dropped`, `--limit` | 순서·finding·truncation 채널 |
| done | `--expect`, `--require-landed` | item/finding archive와 landing verdict |
| undone | — | 같은 ID·finding·관계 restore |
| next | bare, ID, `--spec`, `--expect` | read-only 후보와 atomic pick 분리 |
| unpick | — | queued 복귀, ID·added_at 보존 |
| edit | `--expect` | 본문만 변경, mismatch byte 불변 |
| move | `--top`, `--bottom`, `--before`, `--after` | 정렬과 ID 불변 |
| drop | `--expect` | reason·dropped state |
| undrop | `--expect` | queued 복귀와 reason history |
| analyze | — | findings와 분석 출력 |
| relate | `--relation`, `--note` | 네 기존 relation 방향·note bytes |
| unrelate | index | 대상 relation만 제거 |
| why | — | finding·relation 설명 |
| pr | optional ID, `--json` | PR evidence 조회·출력 |
| landed | `--sha`, `--ref`, `--clear` | landing record·provenance |
| auto-done | `--fetch`, `--dry-run`, `--json` | dry-run 무변경, archive 대상 동일 |
| export-json | — | logical export 동등성 |
| history | optional ID, `--limit` | event order·truncation 동일 |

## §D. RED-now Evidence Ledger

문서 수준 tree SHA: `b45c813493751106cd6619dce60fba1c61e70583`. 이 SHA는 아래 RED-GTD-002~025 중 criterion-level pin이 없는 모든 entry에 직접 적용된다. 모든 명령은 이 worktree에서 한 번의 read-only invocation으로 실행했고 raw stdout은 빈 byte sequence였다. exit `1`은 대상 디렉터리를 정상 탐색했으나 계획된 고유 심볼이 아직 없다는 뜻이므로, 신규 capability 부재라는 각 criterion의 옳은 RED다.

```text
id: RED-GTD-002
command: rg -n 'func NewGTDCommand' internal/cli
raw_stdout: ""
exit_code: 1
right_red_reason: canonical gtd command constructor가 아직 없다.
```
```text
id: RED-GTD-003
command: rg -n 'GTDCanonicalSurfaceGolden' internal/template
raw_stdout: ""
exit_code: 1
right_red_reason: canonical/compatibility 생성물 golden guard가 아직 없다.
```
```text
id: RED-GTD-005
command: rg -n 'func CaptureGTDItem' internal/kanban
raw_stdout: ""
exit_code: 1
right_red_reason: Capture 경계가 아직 없다.
```
```text
id: RED-GTD-006
command: rg -n 'func ClarifyGTDItem' internal/kanban
raw_stdout: ""
exit_code: 1
right_red_reason: Clarify 경계가 아직 없다.
```
```text
id: RED-GTD-007
command: rg -n 'func OrganizeGTDItem' internal/kanban
raw_stdout: ""
exit_code: 1
right_red_reason: Organize 경계가 아직 없다.
```
```text
id: RED-GTD-008
command: rg -n 'func ReflectGTDState' internal/kanban
raw_stdout: ""
exit_code: 1
right_red_reason: Reflect 경계가 아직 없다.
```
```text
id: RED-GTD-009
command: rg -n 'func EngageGTDItem' internal/kanban
raw_stdout: ""
exit_code: 1
right_red_reason: Engage 경계가 아직 없다.
```
```text
id: RED-GTD-010
command: rg -n 'func MigrateGTDSchema' internal/kanban
raw_stdout: ""
exit_code: 1
right_red_reason: additive GTD migration이 아직 없다.
```
```text
id: RED-GTD-011
command: rg -n 'TestGTDLegacyRoundTrip' internal/kanban
raw_stdout: ""
exit_code: 1
right_red_reason: old/new 왕복 검증기가 아직 없다.
```
```text
id: RED-GTD-012
command: rg -n 'func ValidateGTDRelation' internal/kanban
raw_stdout: ""
exit_code: 1
right_red_reason: 신규·기존 relation 통합 validator가 아직 없다.
```
```text
id: RED-GTD-013
command: rg -n 'func BuildPrivateGTDProjection' internal/graph
raw_stdout: ""
exit_code: 1
right_red_reason: 권한·노출 정책을 집행하는 private projection builder가 아직 없다.
```
```text
id: RED-GTD-014
command: rg -n 'func SealMissionContract' internal
raw_stdout: ""
exit_code: 1
right_red_reason: 완전성 검사와 hash 봉인을 수행하는 contract 경계가 아직 없다.
```
```text
id: RED-GTD-015
command: rg -n 'func ValidateMissionDecision' internal
raw_stdout: ""
exit_code: 1
right_red_reason: deterministic mission decision validator가 아직 없다.
```
```text
id: RED-GTD-016
command: rg -n 'func NewAutoMissionCommand' internal/cli
raw_stdout: ""
exit_code: 1
right_red_reason: condition parser와 분리된 auto mission command가 아직 없다.
```
```text
id: RED-GTD-017
command: rg -n 'func SuppressAutoMissionQuestions' internal
raw_stdout: ""
exit_code: 1
right_red_reason: sealed scope 안 무질문·밖 blocked 경계가 아직 없다.
```
```text
id: RED-GTD-018
command: rg -n '^name: mission-governor$' .claude/agents/moai
raw_stdout: ""
exit_code: 1
right_red_reason: 비쓰기 mission-governor agent가 아직 없다.
```
```text
id: RED-GTD-019
command: rg -n 'func ProbeMissionRuntimeCapability' internal
raw_stdout: ""
exit_code: 1
right_red_reason: full/partial/unsupported runtime probe가 아직 없다.
```
```text
id: RED-GTD-020
command: rg -n 'func ReconcileMissionOperation' internal
raw_stdout: ""
exit_code: 1
right_red_reason: stable operation readback/reconciliation 경계가 아직 없다.
```
```text
id: RED-GTD-021
command: rg -n 'TestGTDAutonomyEndToEnd' internal
raw_stdout: ""
exit_code: 1
right_red_reason: capture부터 dispatch까지 identity·순서·exactly-once를 묶는 E2E가 아직 없다.
```
```text
id: RED-GTD-022
command: rg -n 'func IntegrateCardIntoLocalDevelop' internal
raw_stdout: ""
exit_code: 1
right_red_reason: repo-local no-ff integration 경계가 아직 없다.
```
```text
id: RED-GTD-023
command: rg -n 'func ValidateDevelopBatchMerge' internal
raw_stdout: ""
exit_code: 1
right_red_reason: develop batch push부터 main release PR까지 SHA gate가 아직 없다.
```
```text
id: RED-GTD-024
command: rg -n 'TestMissionInputCannotEscalatePolicy' internal
raw_stdout: ""
exit_code: 1
right_red_reason: 비신뢰 입력의 policy escalation mutant 검증이 아직 없다.
```
```text
id: RED-GTD-025
command: rg -n 'func FinalizeOrRevokeMission' internal
raw_stdout: ""
exit_code: 1
right_red_reason: revoke/completion authoritative state 경계가 아직 없다.
```

## §E. Green Path Ledger

각 명령은 구현 milestone에서 실제로 생성할 정확한 test name을 고정한다. 기대 출력은 `=== RUN <name>`, `--- PASS: <name>`, package `ok`이며, `=== RUN` 1개 이상과 `[no tests to run]`/`[no test files]` 부재를 함께 확인해야 한다. 명령 exit는 0이어야 한다.

| ID | Milestone | 실제 테스트 명령 | 위반 mutant |
|---|---|---|---|
| GREEN-GTD-002 | M1 | `go test -v -run '^TestGTDAllTodoVerbsParity$' ./internal/cli` | gtd의 `done --expect`만 다른 handler로 연결 |
| GREEN-GTD-003 | M1/M9 | `go test -v -run '^TestGTDCanonicalSurfaceGolden$' ./internal/template` | 한 locale 또는 emitted todo wrapper만 독립 본문 사용 |
| GREEN-GTD-005 | M3 | `go test -v -run '^TestCaptureGTDItem$' ./internal/kanban` | duplicate event가 두 item 생성 |
| GREEN-GTD-006 | M3 | `go test -v -run '^TestClarifyGTDItem$' ./internal/kanban` | 권한 불명 항목을 publishable로 분류 |
| GREEN-GTD-007 | M3 | `go test -v -run '^TestOrganizeGTDItem$' ./internal/kanban` | depends_on cycle 허용 |
| GREEN-GTD-008 | M3 | `go test -v -run '^TestReflectGTDState$' ./internal/kanban` | 취소를 dependency 완료로 처리 |
| GREEN-GTD-009 | M3 | `go test -v -run '^TestEngageGTDItem$' ./internal/kanban` | dry-run이 queue를 변경 |
| GREEN-GTD-010 | M2 | `go test -v -run '^TestMigrateGTDSchemaIdempotent$' ./internal/kanban` | 두 번째 migration이 revision 증가 |
| GREEN-GTD-011 | M2 | `go test -v -run '^TestGTDLegacyRoundTrip$' ./internal/kanban` | old writer 이후 GTD relation 소실 |
| GREEN-GTD-012 | M4 | `go test -v -run '^TestValidateGTDRelationCompatibility$' ./internal/kanban` | conflicts 방향 반전 또는 note 재직렬화 |
| GREEN-GTD-013 | M4 | `go test -v -run '^TestBuildPrivateGTDProjectionPrivacy$' ./internal/graph` | 0644 sidecar 또는 default export 포함 |
| GREEN-GTD-014 | M5 | `go test -v -run '^TestSealMissionContractCompleteness$' ./internal/mission` | 금지선 없는 contract 승인 |
| GREEN-GTD-015 | M5 | `go test -v -run '^TestValidateMissionDecision$' ./internal/mission` | stale snapshot 허용 |
| GREEN-GTD-016 | M6 | `go test -v -run '^TestNewAutoMissionCommandTreatsTextAsData$' ./internal/cli` | mission text를 condition parser로 전달 |
| GREEN-GTD-017 | M6 | `go test -v -run '^TestAutoMissionQuestionBoundary$' ./internal/mission` | scope 밖에서 질문 후 계속 실행 |
| GREEN-GTD-018 | M6 | `go test -v -run '^TestMissionGovernorPermissionBoundary$' ./internal/mission` | governor에 write 또는 git 권한 부여 |
| GREEN-GTD-019 | M7 | `go test -v -run '^TestMissionRuntimeCapabilityMatrix$' ./internal/mission` | partial runtime을 durable로 표시 |
| GREEN-GTD-020 | M7/M8 | `go test -v -run '^TestReconcileMissionOperationCrashCuts$' ./internal/mission` | invoked receipt를 무조건 재호출 |
| GREEN-GTD-021 | M8 | `go test -v -run '^TestGTDAutonomyEndToEnd$' ./internal/mission` | 두 manager가 같은 card를 두 번 dispatch |
| GREEN-GTD-022 | M8 | `go test -v -run '^TestIntegrateCardIntoLocalDevelop$' ./internal/mission` | lane push/card PR 또는 fast-forward merge 허용 |
| GREEN-GTD-023 | M8 | `go test -v -run '^TestValidateDevelopBatchMerge$' ./internal/mission` | 다른 origin/develop SHA의 CI 재사용 |
| GREEN-GTD-024 | M5/M6 | `go test -v -run '^TestMissionInputCannotEscalatePolicy$' ./internal/mission` | 외부 문서 명령이 allowlist 변경 |
| GREEN-GTD-025 | M8 | `go test -v -run '^TestFinalizeOrRevokeMission$' ./internal/mission` | idle 또는 lease expiry만으로 completed |

## §F. Edge Cases

- 두 관리자가 같은 event와 operation을 동시에 처리한다.
- archive와 reopen 사이에 근거 revision 또는 policy version이 바뀐다.
- graph projection 임시 파일 작성 뒤 atomic publish 전에 프로세스가 종료된다.
- merge API 응답이 유실됐지만 원격에는 merge가 반영되어 있다.
- session API는 존재하지만 reconnect가 지원되지 않거나 자격증명이 만료됐다.
- todo와 gtd 중 하나의 help/completion만 갱신된 mutant가 주입된다.
- advisor와 governor가 동일한 실패 결정을 반복해 정체 한도를 소진한다.

## §G. Quality Gates

- 23개 release-blocking AC가 RED-now와 green-path pair를 가지며 2개 regression-guard는 기존 보존 동작의 비회귀 evidence를 가진다.
- 영향 Go 패키지의 unit/integration/race 테스트가 현재 worktree에서 PASS한다.
- migration과 crash-cut 테스트는 격리 `MOAI_HOME`만 사용한다.
- 신규 Go 경계의 statement coverage는 프로젝트 Tier L 기준 85% 이상이며 안전 경로는 위반 mutant를 검출한다.
- 현재 `origin_develop_sha`의 CI와 release PR head의 필수 CI·review·audit가 모두 PASS하고 `main_landed_sha`를 readback한다.
- template source와 emitted artifact, 4개 언어 링크·help·completion golden이 일치한다.
- 일반 goal과 Kanban/Factory mode의 질문·승인 계약 회귀 테스트가 PASS한다.

## §H. Definition of Done

- 23개 release-blocking AC가 nonempty sweep으로 PASS하고 AC-GTD-001·004 regression-guard에서 보존 동작의 악화가 없다.
- M1~M9 산출물이 현재 대상 SHA에 존재하고 독립 sync audit가 PASS다.
- 기존 큐 fixture에서 ID·순서·상태·보관·복원 회귀가 없다.
- `--auto`와 기존 progression mode가 별도 상태와 경로로 동작한다.
- 중단·재시작·중복 이벤트·동시 manager에서 중복 부작용이 없다.
- local develop `--no-ff` merge, `origin_develop_sha` CI, release→main PR, `main_landed_sha`와 ancestry를 모두 확인했다.
- Gaps와 residual risk가 progress.md §E.4의 sync evidence에 기록되었다.
