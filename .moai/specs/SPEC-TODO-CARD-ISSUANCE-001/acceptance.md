# SPEC-TODO-CARD-ISSUANCE-001 — 수용 기준

Tier L. 출시 차단으로 분류한 모든 기준은 `verification-completeness.md` §2 와 §2.1 에 따라 **RED-now 셀**(읽기 전용 단일 호출 명령, 그 명령의 verbatim stdout, 종료 코드, 고정된 트리)과 **green-path 셀**(어느 마일스톤이 뒤집는지와 통과했을 때의 출력)을 짝으로 가진다. 도착 시점에 붉을 수 없는 기준은 **회귀 가드**로 표시하고 출시 차단으로 기록하지 않는다. Given-When-Then 은 검증 층의 형식이고 요구는 `spec.md` 의 GEARS 다. 모델 행위라서 기계 검사가 닿지 않는 조항은 *(doctrine-only)* 로 적는다.

**문서 수준 핀.** 아래 원장의 모든 행은 트리 `2de0a2cb6`(전체 SHA `2de0a2cb613b04765a1554f86685a3b48e0be806`, 로컬 develop 팁)에서 측정했다. 측정 시점의 워킹 트리는 깨끗했고(`git status --short` 빈 출력) 이 SPEC 자신의 파일은 아직 없었다. 핀이 따로 적히지 않은 기준은 이 핀에 묶인다. 판정 도구(`git`, `ls`, `wc`, `go test -list`)는 트리가 아니라 설치된 일반 명령이다 — `moai` 바이너리는 RED 행에 쓰지 않았다(이 세션은 레인 표지가 있어 `moai todo add` 가 거절된다).

**상시 점화 규칙(모든 기준이 인용).** `verification-completeness.md` §1.2·§1.3 에 맞춰 기준마다 네 가지를 적는다. 공통 항은 여기서 한 번만 정의한다.

- **W(언제 돌아야 하나)**: 소유 마일스톤의 종료 시점과 sync-audit 시점. 마일스톤 이전에 돌리면 구조적으로 항상 초록이거나 항상 빨강이다.
- **V(붉음을 누가 보나)**: 마일스톤을 소유한 레인이 마일스톤 게이트에서 실패한 `go test` 종료 코드 1 로 본다. CI 는 같은 테스트를 푸시마다 실패한 패키지로 본다. `git grep -c -F` 형 기준은 빈 출력과 종료 코드 1 이다.
- **S(멈춘 것을 독자가 어떻게 아나)**: 시험 이름 기준은 `go test <패키지> -list '<앵커된 패턴>'` 가 출력한 이름 수가 지정한 수와 같은지를 짝으로 본다. 이름이 바뀌거나 지워지면 짝의 `git grep -c -F` 가 비어 종료 코드 1 이고 `-list` 수가 모자란다. 이 짝은 마일스톤 종료 때 `progress.md` §E.2 에 기록되고 sync-audit 에서 다시 돈다.

각 기준 끝의 "점화" 줄은 W·V·S 를 가리키고, 붉어지는 **입력**(R)만 기준별로 적는다.

## 증거 원장 (RED-now 관측)

열: Id, 명령(단일 호출, 파이프·리디렉션·`&&`·`;`·서브셸 없음), 그 명령의 verbatim stdout, 종료 코드(별도 필드), 붉은 이유. stdout 이 비면 `(empty)`; `ls` 가 실패하면 오류 문장은 stderr 로 나가므로 따로 적는다. 행 id 는 `L<n>`(RED-now), `C<n>`(대조군), `G<n>`(맥락·가드)다.

| Id | 명령 | verbatim stdout | 종료 | 붉은 이유 |
|---|---|---|---|---|
| L1 | `ls .moai/reports/t1454` | (empty); stderr `ls: .moai/reports/t1454: No such file or directory` | 1 | 기준선 디렉터리가 아직 없다 — M0 가 만든다 |
| L2 | `git grep -c -F "TestTodoAddPresentationStderrOnly" -- internal/cli` | (empty) | 1 | 제시 stderr 시험이 아직 없다 |
| L3 | `git grep -c -F "TestIssuanceNeighborsIncludeArchivedAndDropped" -- internal/kanban` | (empty) | 1 | 보관·dropped 를 포함하는 이웃 조회와 시험이 아직 없다 |
| L4 | `git grep -c -F "TestTodoAddPresentationProbeOutsideLock" -- internal/cli` | (empty) | 1 | 락 밖 탐침 시험이 없다 |
| L5 | `git grep -c -F "TestTodoAddPresentationDryRunWritesNothing" -- internal/cli` | (empty) | 1 | `--dry-run` 시험이 없다 |
| L6 | `git grep -c -F "dryRun" -- internal/cli/todo.go` | (empty) | 1 | `add` 인자 스캔에 `--dry-run` 이 없다 |
| L7 | `git grep -c -F "TestTodoAddMCPCarriesPresentation" -- internal/cli` | (empty) | 1 | MCP 가 제시를 싣는 시험이 없다 |
| L8 | `git grep -c -F "issuance:TEXT:0:NULL" -- internal/kanban/backlog_schema_freeze_test.go` | (empty) | 1 | 동결 튜플 시험이 새 컬럼을 아직 고정하지 않는다 |
| L9 | `git grep -c -F "TestBacklogIssuanceArchiveRestoreRoundTrip" -- internal/kanban` | (empty) | 1 | 보관·복원 왕복 시험이 없다 |
| L10 | `git grep -c -F "TestTodoDropStoresReason" -- internal/cli` | (empty) | 1 | drop 사유 속성 저장 시험이 없다 |
| L11 | `git grep -c -F "TestFindingDispositionRecordOnly" -- internal/kanban` | (empty) | 1 | finding 처분과 그 시험이 없다 |
| L12 | `git grep -c -F "TestRelationOntologyMapsLegacyKinds" -- internal/kanban` | (empty) | 1 | 일곱 종류 어휘와 매핑 시험이 없다 |
| L13 | `git grep -c -F "TestTodoRelateConstraintRefusalByteIdentical" -- internal/cli` | (empty) | 1 | 종류별 제약 거절 시험이 없다 |
| L14 | `git grep -c -F "TestRelationConstraintRefusesCycle" -- internal/kanban` | (empty) | 1 | 순환 제약 시험이 없다 |
| L15 | `git grep -c -F "TestCardGTDResolverBothDirections" -- internal/kanban` | (empty) | 1 | 카드↔GTD 해석기가 없다 |
| L16 | `git grep -c -F "TestTodoTraceTransitiveDeterministic" -- internal/cli` | (empty) | 1 | `trace` 동사와 시험이 없다 |
| L17 | `git grep -c -F "TestGraphCardFileEdgesDeterministic" -- internal/graph` | (empty) | 1 | card→file 간선 층과 시험이 없다 |
| L18 | `git grep -c -F "TestFactoryNextBundleSerialLane" -- internal/cli` | (empty) | 1 | 묶음 직렬 레인 시험이 없다 |
| L19 | `git grep -c -F "TestTodoMergeRecordsAndDrops" -- internal/cli` | (empty) | 1 | 병합 동사와 시험이 없다 |
| L20 | `git grep -c -F "TestFactoryAssignBundleHubChain" -- internal/cli` | (empty) | 1 | 허브 체인 시험이 없다 |
| L21 | `git rev-list --first-parent --count -E -i --grep='^merge[( :]+(card )?t1453' HEAD` | `0` | 0 | t1453 병합이 HEAD 의 first-parent 역사에 없다 — **M5 게이트가 닫혀 있다**(종료 코드는 0 이므로 출력의 수를 읽는다) |
| L22 | `ls .claude/rules/moai/workflow/card-issuance.md` | (empty); stderr `ls: .claude/rules/moai/workflow/card-issuance.md: No such file or directory` | 1 | 발행 규칙 파일이 없다 |
| L23 | `ls .moai/reports/t1454/m5-draft` | (empty); stderr `ls: .moai/reports/t1454/m5-draft: No such file or directory` | 1 | M5 초안 디렉터리가 없다 |
| L24 | `wc -m .claude/skills/moai/workflows/gtd.md` | `   40037 .claude/skills/moai/workflows/gtd.md` | 0 | `gtd.md` 가 40,000자 한도를 37자 넘는다 |
| L25 | `git grep -c -F "TestTodoGraphViewRendersRelations" -- internal/web` | (empty) | 1 | 웹 그래프 보기와 시험이 없다 |
| L26 | `git grep -c -F "amendment_of" -- .moai/specs/SPEC-TODO-ANALYSIS-001/spec.md` | (empty) | 1 | 대상 SPEC 이 아직 개정 선언을 갖지 않는다 |
| L27 | `git grep -c -F "TestTodoMergeNeverInvokedByAnalysis" -- internal/cli` | (empty) | 1 | 분석 경로가 병합을 못 부른다는 시험이 없다 |

**대조군**(같은 경로·같은 탐침이 존재하는 것을 맞히므로 위 빈 행이 측정된 부재임을 보인다).

| Id | 명령 | verbatim stdout | 종료 | 대조하는 행 |
|---|---|---|---|---|
| C1 | `git grep -c -F "TestTodoAddRefusesExactDuplicate" -- internal/cli` | `internal/cli/todo_analysis_add_test.go:2` | 0 | L2, L4~L5, L7, L10, L13, L16, L18~L20, L27 |
| C2 | `git grep -c -F "TestClassifyCardText" -- internal/kanban` | `internal/kanban/backlog_analysis_test.go:2` | 0 | L3, L9, L11, L12, L14, L15 |
| C3 | `git grep -c -F "scan.pick" -- internal/cli/todo.go` | `internal/cli/todo.go:2` | 0 | L6 |
| C4 | `git grep -c -F "lease_expires_at:TEXT:0:NULL" -- internal/kanban/backlog_schema_freeze_test.go` | `internal/kanban/backlog_schema_freeze_test.go:2` | 0 | L8 |
| C5 | `git grep -c -F "func TestBuild" -- internal/graph/graph_test.go` | `internal/graph/graph_test.go:3` | 0 | L17 |
| C6 | `git grep -c -F "func Test" -- internal/web/todo_route_test.go` | `internal/web/todo_route_test.go:6` | 0 | L25 |
| C7 | `git grep -c -F "TestSD_AC014_MCPMatchesCLIWithProjectRoot" -- internal/cli` | `internal/cli/factory_m3_test.go:1` | 0 | L7(MCP 패리티 시험은 이미 있다) |
| C8 | `git rev-list --first-parent --count -E -i --grep='^merge[( :]+(card )?t1448' HEAD` | `1` | 0 | L21(선택자가 눈먼 것이 아니다) |
| C9 | `git rev-list --first-parent --count -E -i --grep='^merge[( :]+(card )?t1344' HEAD` | `1` | 0 | L21(두 번째 양성 대조) |
| C10 | `git grep -c -F "amendment_of" -- .moai/specs/SPEC-DRIFT-CLOSE-BODY-001/spec.md` | `.moai/specs/SPEC-DRIFT-CLOSE-BODY-001/spec.md:1` | 0 | L26 |
| C11 | `ls .claude/rules/moai/workflow/kanban-dispatch-detail.md` | `-rw-r--r--@ 1 goos  staff  43403 Oct  3 11:30 .claude/rules/moai/workflow/kanban-dispatch-detail.md` | 0 | L22(`ls` 가 같은 디렉터리에서 존재하는 파일은 맞힌다) |

**맥락·가드 행**(RED 셀이 아니다 — 도착 시 초록이거나 정보용).

| Id | 명령 | verbatim stdout | 종료 | 쓰임 |
|---|---|---|---|---|
| G1 | `git diff --name-only 2de0a2cb6 -- internal/kanban/backlog_analysis.go` | (empty) | 0 | AC-TCI-007 의 도착 시 상태: 분류기 파일이 핀에서 변하지 않았다 |
| G2 | `go test ./internal/cli -list '^TestTodoListJSON_GoldenByteIdentity$'` | `TestTodoListJSON_GoldenByteIdentity` · `ok  	github.com/modu-ai/moai-adk/internal/cli	1.528s` | 0 | AC-TCI-009 의 골든 시험이 존재하고 선택된다(선택 수 1) |
| G3 | `go test ./internal/kanban -list '^TestClassifyCardText$'` | `TestClassifyCardText` · `ok  	github.com/modu-ai/moai-adk/internal/kanban	0.409s` | 0 | AC-TCI-007 의 고정 시험이 선택된다. 같은 방식으로 `^TestNormalizeCardText$`·`^TestTokenSetJaccard$`(kanban), `^TestTodoAddRefusesExactDuplicate$`·`^TestTodoAddNearDuplicateRecordsOnly$`·`^TestSD_AC014_MCPMatchesCLIWithProjectRoot$`·`^TestTodoAdd_PrintsIDAndPosition$`(cli)를 각각 한 번씩 `-list` 해 이름 한 줄과 `ok` 를 관측했다(종료 0) |
| G4 | `git cat-file -s 2de0a2cb6:.claude/rules/moai/workflow/kanban-dispatch.md` | `28301` | 0 | AC-TCI-021 상시 로드 바이트 기준(핀) |
| G5 | `wc -m .claude/rules/moai/workflow/kanban-dispatch.md` | `   28092 .claude/rules/moai/workflow/kanban-dispatch.md` | 0 | 같은 파일의 문자 기준(핀) |
| G6 | `wc -m .claude/rules/moai/workflow/kanban-dispatch-detail.md` | `   43138 .claude/rules/moai/workflow/kanban-dispatch-detail.md` | 0 | 이미 40,000자 초과 — 증가 금지 기준(핀) |
| G7 | `cmp .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `.claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md differ: char 25113, line 181` | 1 | 분기된 쌍의 기존 분기(보존 대상) |
| G8 | `cmp .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md` | (empty) | 0 | 바이트 동일 쌍(동일 유지 대상) |
| G9 | `git merge-base develop HEAD` | `2de0a2cb613b04765a1554f86685a3b48e0be806` | 0 | 측정 시점의 `CARD_BASE`. 읽는 시점에 다시 구한다 — 핀하지 않는다 |
| G10 | `git rev-list -n 1 --first-parent --grep='^merge(t1448)' HEAD` | `4315f0d0e9742b4e3741eb20efc1614f7f395238` | 0 | t1448 병합(게이트 대조군의 정체) |

기준선 재측정 요약은 `research.md` §3(figure 35개, git 쪽 헤드라인 전부 재현). 위 행은 SPEC 파일이 생기기 **전**에 측정했고, 이 SPEC 의 파일(미추적)은 `git grep` 에 잡히지 않으므로 행의 출력은 SPEC 작성 뒤에도 같다.

## AC-TCI-001 — 기준선이 추적되는 형태로 반입되고 재현 가능하다 (REQ-TCI-001)

**Covers**: maps REQ-TCI-001

Release-blocking.

**Given** M0 가 끝난 트리, `research.md` §3 의 figure 35개(GB 16, QB 10, SB 9)
**When** `.moai/reports/t1454/baseline/` 를 읽는다
**Then** (a) `baseline.md` 가 있고, (b) `figure:`·`command:`·`tree:` 줄의 수가 같고 35 이상이다(세 번의 `git grep -c -F` 가 같은 수를 낸다), (c) git 쪽 헤드라인 figure(GB02, GB05, GB06, GB10, GB13, GB14)를 같은 명령으로 다시 돌린 값이 기록과 같거나 각 차이가 `drift:` 줄에 설명돼 있다, (d) 끝의 "임계값" 표의 각 값 옆에 측정 명령이 있다.

- **RED-now:** L1(디렉터리 없음, 종료 1; 대조군 C11 형태).
- **Green path:** M0 — `ls .moai/reports/t1454/baseline` 가 `baseline.md` 와 스크립트를 나열한다.
- **Mutant probe:** (MU-1) 디렉터리만 만들고 비어 있음 → (a)(b) 실패. (MU-2) 이 문서의 숫자를 베끼되 명령이 없음 → `command:` 수가 0 이라 (b) 실패. (MU-3) 명령은 있으나 출력이 재현되지 않는 값 → (c) 가 git 쪽 값에서 잡는다. 잔여: 큐 쪽 값은 비공개 DB 위의 스냅숏이라 제3자가 재실행할 수 없다 — `queue-snapshot:` 줄로 귀속만 하고 재현 주장을 하지 않는다(수용된 잔여).
- **점화:** W 는 M0 종료와 sync-audit, R 은 디렉터리 삭제나 `command:` 없는 figure, V·S 는 공통(+ `ls` 종료 1).

## AC-TCI-002 — 제시는 stderr 로만 나가고 stdout 은 기계 줄 그대로다 (REQ-TCI-002)

**Covers**: maps REQ-TCI-002

Release-blocking.

**Given** 큐에 live 카드 L, `[DROPPED — <reason>] ` 접두사를 단 dropped 카드 D, 보관 카드 A 가 있고 셋 모두 새 본문 N 과 표시 하한 이상으로 겹친다
**When** `add "<N>"` 이 실행된다(`go test` 도구, 레인 변수 제거)
**Then** (a) stdout 은 정확히 `t<id> <pos>\n`(바이트 비교), (b) stderr 에 이웃 줄이 3개 이하로 나오고 각 줄이 id·상태(`live|dropped|archived`)·점수·`measure=`·본문 앞 60자를 담으며, (c) D 의 줄은 접두사를 뺀 본문을 보이고 사유를 따로 밝히며, (d) L·D·A 의 id 가 모두 나타난다.

- **RED-now:** L2(대조군 C1).
- **Green path:** M1 — `go test ./internal/cli -run '^TestTodoAddPresentationStderrOnly$' -count=1 -v` 가 `--- PASS`, 같은 패턴 `-list` 가 이름 1개.
- **Mutant probe:** (MU-4) 제시를 stdout 에 출력 → (a) 실패. (MU-5) live 만 조회 → (d) 에서 D·A 부재로 실패. (MU-6) 접두사를 벗기지 않음 → (c) 실패. (MU-7) 상위 4개를 출력 → 줄 수 단언 실패.
- **점화:** W 는 M1 종료, R 은 stdout 에 한 줄이라도 더 쓰는 변경, V·S 공통.

## AC-TCI-003 — 제시의 출처 네 가지와 읽기 전용 조회 (REQ-TCI-002, -006)

**Covers**: maps REQ-TCI-002, REQ-TCI-006

Release-blocking.

**Given** 출처별 고정 입력(보관·dropped 이웃, 경로를 담은 열린 카드, `status: completed` 인 SPEC 디렉터리 고정 입력과 `status: draft` 인 하나, 예상 파일이 없는 신규 카드)
**When** 조회가 실행된다
**Then** (a) 이웃은 live·dropped·archived 를 모두 후보로 삼고 dropped 접두사를 벗기며 상한 3·표시 하한을 지킨다, (b) 같은 구성요소의 열린 카드는 공유 구성요소 키(경로 앞 두 마디)가 있을 때만 나온다, (c) 완료 SPEC 조회는 `completed` 만 포함하고 `draft` 는 제외하며 `heuristic` 표지를 단다, (d) 예상 파일이 없는 신규 카드의 겹침 항목은 `unmeasured` 이고 `none` 이 아니다, (e) 모든 항목이 `source` 와 `measure` 를 가진다.

필수 시험 6개(green path 가 전부 포함, `-list` 수 6): `TestIssuanceNeighborsIncludeArchivedAndDropped`, `TestIssuanceNeighborsStripDropPrefix`, `TestIssuanceNeighborsLimitAndFloor`, `TestIssuanceSameComponentOpenCards`, `TestIssuanceCompletedSpecCoverage`, `TestIssuanceInFlightOverlapUnmeasured`.

- **RED-now:** L3(대조군 C2).
- **Green path:** M1.
- **Mutant probe:** (MU-8) 이웃을 `rec.Items` 에서만 뽑음 → (a) 에서 보관 후보 부재. (MU-9) 하한 무시 → 소음 고정 입력이 나와 실패. (MU-10) 완료 SPEC 조회가 draft 포함 → (c) 실패. (MU-11) 겹침을 비었을 때 `none` 으로 표기 → (d) 실패. (MU-12) 이웃 조회가 `ClassifyCardText` 호출 → AC-TCI-007 의 파일 불변 가드가 잡는다.
- **점화:** W 는 M1 종료, R 은 후보 집합이 줄어드는 변경, V·S 공통.

## AC-TCI-004 — 제시는 admit 을 바꾸지 않고 락 밖에서 돈다 (REQ-TCI-003)

**Covers**: maps REQ-TCI-003

Release-blocking; **연언**이라 공허하지 않다.

**Given** 같은 큐와 같은 본문, 제시 출처 하나가 실패하거나 시간 상한을 넘는 주입(시험 이음매)
**When** `add` 가 실행된다
**Then** (a) 종료 코드와 큐 파일이 제시를 끈 실행과 같고(새 카드 행과 분석기가 원래 기록하던 소견만 다르다), (b) 정확 중복 거절은 여전히 유일한 거절이고 그때 큐 파일은 바이트 동일이며, (c) 제시 탐침이 대기하는 동안 다른 프로세스의 `Mutate` 가 끝난다(락을 쥐지 않는다), (d) 시간 상한을 넘은 항목은 `unmeasured (time bound)` 로 출력되고 admit 은 성공한다.

필수 시험 3개: `TestTodoAddPresentationProbeOutsideLock`, `TestTodoAddPresentationNeverBlocks`, `TestTodoAddPresentationTimeBound`.

- **RED-now:** L4(대조군 C1).
- **Green path:** M1.
- **Mutant probe:** (MU-13) 탐침을 `Mutate` 안으로 옮김 → (c) 에서 두 번째 쓰기가 막혀 실패. (MU-14) 탐침 오류를 add 오류로 전파 → (a) 실패. (MU-15) 상한 없음 → (d) 의 막힘 이음매에서 시험 시간 초과.
- **점화:** W 는 M1 종료, R 은 탐침이 락 안으로 들어가는 변경, V 는 시험 실패(시간 초과 포함), S 공통.

## AC-TCI-005 — `--dry-run` 은 아무것도 쓰지 않는다 (REQ-TCI-004)

**Covers**: maps REQ-TCI-004

Release-blocking.

**Given** 큐 파일 F 와 `meta.last_seq` 값
**When** `add --dry-run "<본문>"` 이 실행된다(정확 중복 본문 한 번 포함)
**Then** (a) 종료 0, (b) F 가 바이트 동일, (c) `last_seq` 불변(id 소비 없음), (d) stdout 에 `<id> <pos>` 줄이 없고 stderr 에 제시와 `dry-run: nothing was written` 줄이 있으며, (e) 정확 중복이면 `a real add would refuse: t<N> already holds this card` 를 출력하고 여전히 종료 0, (f) 알 수 없는 플래그는 기존처럼 `unknown flag` 로 거절된다(회귀).

필수 시험: `TestTodoAddPresentationDryRunWritesNothing`, `TestTodoAddDryRunFlagParsed`(`scanTodoAddArgs` 가 `--dry-run` 을 아는 것과 폴스루가 플래그를 받지 않는 것 둘 다).

- **RED-now:** L5, L6(대조군 C1, C3).
- **Green path:** M1.
- **Mutant probe:** (MU-16) id 를 소비한 뒤 되돌림 → (c) 실패. (MU-17) dry-run 이 소견을 기록 → (b) 실패. (MU-18) 중복 본문에서 비제로 종료 → (e) 실패.
- **점화:** W 는 M1 종료, R 은 dry-run 경로에서 `Mutate` 를 부르는 변경, V·S 공통.

## AC-TCI-006 — MCP 와 engage 도 제시를 싣고 기존 계약은 그대로다 (REQ-TCI-005)

**Covers**: maps REQ-TCI-005

Release-blocking.

**Given** 이웃이 있는 큐, 그리고 이웃이 없는 빈 큐
**When** MCP `todo_add` 와 `gtd engage` 가 카드를 admit 한다
**Then** (a) MCP 결과 텍스트의 첫 줄은 `<id> <pos>` 이고 그 뒤 빈 줄과 제시가 이어지며, (b) 이웃이 없을 때 MCP 결과 텍스트는 CLI stdout 과 같고(기존 `TestSD_AC014_MCPMatchesCLIWithProjectRoot` 가 그대로 통과), (c) engage 는 제시를 stderr 로 내고 소견을 기록하지 않으며 거절하지 않는다.

필수 시험: `TestTodoAddMCPCarriesPresentation`, `TestGTDEngagePresentationRecordsNothing`.

- **RED-now:** L7(대조군 C7: 패리티 시험은 이미 존재).
- **Green path:** M1.
- **Mutant probe:** (MU-19) 제시를 첫 줄 앞에 둠 → (a) 실패. (MU-20) engage 가 `near-duplicate` 소견을 기록 → (c) 실패. (MU-21) MCP 가 제시가 비었을 때도 빈 줄을 덧붙임 → (b) 실패.
- **점화:** W 는 M1 종료, R 은 MCP 텍스트 형태 변경, V·S 공통.

## AC-TCI-007 — 분류기와 분석 동작은 변하지 않는다 (REQ-TCI-006)

**Covers**: maps REQ-TCI-006

**회귀 가드**(도착 시 초록이라 출시 차단이 아니다; 도착 시 상태는 G1).

**Given** 핀 `2de0a2cb6`
**When** 작업이 끝난 뒤 `git diff --name-only 2de0a2cb6 -- internal/kanban/backlog_analysis.go` 를 읽고 고정 시험을 돌린다
**Then** (a) diff 출력이 비어 있다(파일 불변, 이 SPEC 은 같은 파일에 함수를 더하지 않는다), (b) `TestClassifyCardText`, `TestNormalizeCardText`, `TestTokenSetJaccard`, `TestTodoAddRefusesExactDuplicate`, `TestTodoAddNearDuplicateRecordsOnly`, `TestTodoAnalysisNeverReordersQueue` 가 각각 `-list` 로 선택되고 통과한다.

- **RED-now:** 해당 없음 — 도착 시 초록(G1, G3). 회귀 가드다.
- **Mutant probe:** (MU-22) `ClassifyCardText` 가 보관 카드를 포함하도록 바꿈 → (a) 의 diff 가 비지 않아 잡힌다. (MU-23) 같은 파일에 새 함수를 더함 → (a) 가 잡는다.
- **점화:** W 는 M1 종료와 sync-audit, R 은 그 파일의 어떤 수정, V 는 diff 출력, S 는 `-list` 짝. 핀은 재베이스 때 다시 잰다.

## AC-TCI-008 — 스키마는 가산적이고 이주·parity·보관 왕복이 지켜진다 (REQ-TCI-007, -008, -009)

**Covers**: maps REQ-TCI-007, REQ-TCI-008, REQ-TCI-009

Release-blocking.

**Given** 옛 컬럼 집합으로 만든 DB 고정 입력, 새 속성을 가진 카드와 가지지 않은 카드
**When** 새 엔진이 열고, 순수 읽기가 읽고, 보관·복원이 오간다
**Then** (a) `items`·`archived_items` 가 새 컬럼을 갖고 기존 행의 해시가 같으며, (b) 새 컬럼이 없는 DB 를 순수 읽기가 읽어도 속성은 nil 이고 DB 파일 바이트가 같으며(DDL 없음), (c) 동결 튜플 시험이 두 표의 마지막 튜플로 `issuance:TEXT:0:NULL` 을 고정하고 이 변경이 같은 커밋에서 시험 문자열을 갱신하며, (d) 보관 → 복원 뒤 속성이 같고, (e) parity 단언이 새 컬럼을 포함한다.

필수 시험: `TestBacklogIssuanceColumnRetrofit`, `TestBacklogIssuancePureReaderNoDDL`, `TestBacklogIssuanceArchiveRestoreRoundTrip`, `TestBacklogParityCoversIssuance`, 그리고 갱신된 동결 시험(기존 `TestTodoHistoryAddsNoSchemaChange`, `TestSchemaFreezeRecordsTransitionStamps`).

- **RED-now:** L8, L9(대조군 C4, C2). 동결 튜플 리터럴은 시험이 컬럼을 고정하지 않으면 도착한 트리에서 빈다.
- **Green path:** M2.
- **Mutant probe:** (MU-24) 컬럼을 `items` 에만 추가 → 보관 왕복(d)과 (a) 가 `archived_items` 에서 실패. (MU-25) 컬럼을 v1→v2 재구성 목록에 넣음 → v1 DB 재구성 시험이 컬럼을 잃는다(기존 카드 t1310 순서 시험). (MU-26) retrofit 을 버전 조정보다 앞에서 실행 → 같은 시험이 잡는다. (MU-27) parity 가 새 컬럼을 무시 → (e) 실패.
- **점화:** W 는 M2 종료, R 은 컬럼 목록 변경, V·S 공통.

## AC-TCI-009 — `list --json` 골든은 바이트 동일하다 (REQ-TCI-007)

**Covers**: maps REQ-TCI-007

**회귀 가드**(G2 로 도착 시 초록: 시험이 존재하고 선택된다).

**Given** 새 속성을 하나도 갖지 않은 카드들
**When** `list --json` 이 렌더된다
**Then** 출력이 `internal/cli/testdata/ac_tst_012_golden.json` 과 바이트 동일하다(`TestTodoListJSON_GoldenByteIdentity`).

- **RED-now:** 해당 없음(도착 시 초록).
- **Mutant probe:** (MU-28) 새 필드를 `omitempty` 없이 직렬화 → 골든이 `"issuance":null` 로 달라져 실패.
- **점화:** W 는 M2 종료, R 은 직렬화 형태 변경, V·S 공통.

## AC-TCI-010 — drop 사유가 속성에 저장되고 닫힘 시각은 접근자로 읽힌다 (REQ-TCI-010)

**Covers**: maps REQ-TCI-010

Release-blocking.

**Given** 카드 둘, 하나는 drop 하고 하나는 `done` 한다. 스탬프가 없는 과거 카드 하나
**When** `moai todo drop <id> "<reason>"` 과 접근자가 실행된다
**Then** (a) 사유가 속성에 저장되고, (b) 텍스트 접두사 `[DROPPED — <reason>] ` 도 그대로 쓰이며(기존 독자가 불변), (c) 접근자가 보관 카드는 `archived_at`, dropped 카드는 `dropped_at` 에서 시각을 내고, (d) 스탬프가 없으면 "unknown" 을 낸다.

필수 시험: `TestTodoDropStoresReason`, `TestTodoDropKeepsTextPrefix`, `TestCardClosedAtAccessor`.

- **RED-now:** L10(대조군 C1).
- **Green path:** M2.
- **Mutant probe:** (MU-29) 사유를 접두사에만 둠 → (a) 실패. (MU-30) 접두사를 제거 → (b) 실패. (MU-31) 스탬프 없는 카드에 0 시각을 냄 → (d) 실패.
- **점화:** W 는 M2 종료, R 은 drop 경로 변경, V·S 공통.

## AC-TCI-011 — finding 처분은 기록만 한다 (REQ-TCI-011)

**Covers**: maps REQ-TCI-011

Release-blocking.

**Given** 소견 하나와 카드 둘, 처분 동사
**When** 처분(accept·merge·reject)을 기록한다
**Then** (a) 처분이 소견에만 저장되고 카드 필드·소견의 관계·큐 순서·픽업 필터가 바이트 동일하며, (b) 기존 소견은 처분이 없고 표시도 그대로이며, (c) 알 수 없는 처분 값은 아무것도 쓰지 않고 거절된다.

필수 시험: `TestFindingDispositionRecordOnly`, `TestTodoRelateDispositionVerb`, `TestFindingDispositionAbsentForLegacy`.

- **RED-now:** L11(대조군 C2).
- **Green path:** M2.
- **Mutant probe:** (MU-32) `merge` 처분이 카드를 drop → (a) 실패. (MU-33) `reject` 가 소견을 삭제 → (a) 실패. (MU-34) 기존 소견에 기본 처분을 채움 → (b) 실패.
- **점화:** W 는 M2 종료, R 은 처분 효과가 카드 필드를 건드리는 변경, V·S 공통.

## AC-TCI-012 — 일곱 종류 어휘와 읽기 시 매핑, 옛 독자는 불변이다 (REQ-TCI-012, -017)

**Covers**: maps REQ-TCI-012, REQ-TCI-017

Release-blocking.

**Given** 옛 소견 여덟 종류와 `gtd_relations` 아홉 종류를 담은 고정 입력
**When** 해석기가 읽고 옛 독자(`list`, `why`, `export`, 픽업 필터, 자동 순위의 near-duplicate 경로)가 렌더한다
**Then** (a) 읽기 결과는 일곱 종류만 쓰고 매핑 표(depends·blocks→blocks, near-duplicate·duplicate-forced→duplicates, replaces→supersedes, contains·absorbs·conflicts→relates-to 와 한정어)를 지키며, (b) 저장된 소견 행의 바이트가 읽기 전후 같고, (c) 옛 독자의 출력이 변경 전 골든과 같다.

필수 시험: `TestRelationOntologyMapsLegacyKinds`, `TestRelationLegacyReadersByteIdentical`.

- **RED-now:** L12(대조군 C2).
- **Green path:** M3.
- **Mutant probe:** (MU-35) 매핑이 저장 행을 다시 씀(`replaces`→`supersedes`) → (b) 실패. (MU-36) `conflicts` 를 매핑에서 빠뜨림 → (a) 실패. (MU-37) 새 어휘가 옛 독자의 표시 문자열을 바꿈 → (c) 실패.
- **점화:** W 는 M3 종료, R 은 소견 읽기·렌더 변경, V·S 공통.

## AC-TCI-013 — 종류별 제약은 거절하고 파일을 건드리지 않는다 (REQ-TCI-013)

**Covers**: maps REQ-TCI-013

Release-blocking.

**Given** 큐 파일과 제약 위반 쓰기들: 자기 간선, `blocks` 순환(기존 가드 유지), `parent-of` 순환, 카드의 두 번째 부모, 두 번째 `merged-into` 대상, 같은 `duplicates` 쌍의 역순 재기록
**When** 쓰기 동사가 실행된다
**Then** (a) 각 위반이 종류와 쌍을 이름으로 대는 메시지로 거절되고 큐 파일이 바이트 동일이며, (b) `duplicates` 는 쌍이 정규화돼 역순 재기록이 두 번째 레코드를 만들지 않고, (c) 합법적 쓰기는 통과한다.

필수 시험: `TestTodoRelateConstraintRefusalByteIdentical`(cli), `TestRelationConstraintRefusesCycle`, `TestRelationConstraintCardinality`(kanban).

- **RED-now:** L13, L14(대조군 C1, C2).
- **Green path:** M3.
- **Mutant probe:** (MU-38) `blocks` 에만 순환 검사 → `parent-of` 순환이 통과해 실패. (MU-39) 카디널리티를 메모리에서만 검사하고 영속 상태와 어긋남 → 두 번째 쓰기 시험이 잡는다. (MU-40) 거절 전에 일부를 씀 → 바이트 동일 단언 실패.
- **점화:** W 는 M3 종료, R 은 쓰기 동사 검증 순서 변경, V·S 공통.

## AC-TCI-014 — 카드↔GTD 해석기 (REQ-TCI-014)

**Covers**: maps REQ-TCI-014

Release-blocking.

**Given** engage 된 GTD 항목(관계 보유)과 engage 되지 않은 카드(고정 입력, 라이브 `gtd_relations` 는 0행이라 고정 입력이 필요하다)
**When** 해석기가 양방향으로 읽는다
**Then** (a) `tN → gtd 항목 id → tN` 이 왕복하고, (b) engage 된 적 없는 카드는 GTD 쪽이 비며, (c) `todo why <tN>` 이 GTD 관계 줄을 `source` 표지와 함께 보이고, (d) `gtd_relations` 에 쓰는 새 경로가 없고 `moai gtd organize` 동작이 바이트 동일이다.

필수 시험: `TestCardGTDResolverBothDirections`, `TestCardGTDResolverAddsNoGTDWritePath`, `TestTodoWhyShowsGTDRelations`.

- **RED-now:** L15(대조군 C2).
- **Green path:** M3.
- **Mutant probe:** (MU-41) 해석기가 `gtd_relations` 에 쓴다 → (d) 실패. (MU-42) 한 방향만 구현 → (a) 실패.
- **점화:** W 는 M3 종료, R 은 `gtd_items.card_id` 연결 변경, V·S 공통.

## AC-TCI-015 — `trace` 는 결정적이고 순환에서 끝나며 쓰지 않는다 (REQ-TCI-015)

**Covers**: maps REQ-TCI-015

Release-blocking.

**Given** `A parent-of B parent-of C` 와 `A blocks C`, 그리고 레거시 순환 데이터(`X blocks Y`, `Y blocks X`)
**When** `trace A [--kind parent-of] [--depth 1]` 와 `trace X` 가 실행된다
**Then** (a) 도달한 노드가 (깊이, 종류, id) 순으로 같은 출력을 두 번 내고, (b) `--depth 1` 은 B 까지만, `--kind parent-of` 는 `blocks` 간선을 따르지 않으며, (c) 순환 입력에서 끝난다, (d) 큐 파일이 바이트 동일, (e) 레인 세션에서 거절되지 않는다.

필수 시험: `TestTodoTraceTransitiveDeterministic`, `TestTodoTraceBreaksCycles`, `TestTodoTraceLaneReadOnlyAllowed`, `TestTodoTraceWritesNothing`.

- **RED-now:** L16(대조군 C1).
- **Green path:** M3.
- **Mutant probe:** (MU-43) 깊이 무시 → (b) 실패. (MU-44) 맵 순회 순서 의존 → (a) 의 이중 실행이 달라 실패. (MU-45) 방문 집합 없음 → (c) 시험 시간 초과.
- **점화:** W 는 M3 종료, R 은 탐색 순서·방문 처리 변경, V·S 공통.

## AC-TCI-016 — 그래프 간선은 커밋된 증거에서만 나오고 결정적이다 (REQ-TCI-016)

**Covers**: maps REQ-TCI-016

Release-blocking.

**Given** `merge(t5): …` 와 `Merge card t6 …` 제목의 병합이 있고 각각 파일을 바꾼 고정 git 저장소, 그리고 큐 DB 가 있는/없는 두 환경
**When** `graph build` 가 두 번 돌고 `graph check` 가 읽는다
**Then** (a) `card-file` 간선(`Source: t5`, `Target: <경로>`)이 `(kind, source, target, line)` 순으로 나오고, (b) 두 빌드의 출력이 바이트 동일, (c) 큐 DB 의 유무가 출력을 바꾸지 않으며 미착지 카드 id·예상 파일·소견이 출력에 없고, (d) 새 카드 귀속 병합이 착지하면 `graph check` 가 stale 로 읽는다.

필수 시험: `TestGraphCardFileEdgesDeterministic`, `TestGraphCardFileEdgesCarryNoQueueState`, `TestGraphCheckNoticesCardFileSource`.

- **RED-now:** L17(대조군 C5).
- **Green path:** M3.
- **Mutant probe:** (MU-46) 큐에서 예상 파일을 읽어 간선에 넣음 → (c) 실패. (MU-47) 시각·정렬 불안정 → (b) 실패. (MU-48) 출처 지문 미등록 → (d) 실패.
- **점화:** W 는 M3 종료와 `graph-freshness` CI, R 은 간선 층 입력 변경, V·S 공통.

## AC-TCI-017 — 묶음은 한 레인에 직렬로 실린다 (REQ-TCI-018)

**Covers**: maps REQ-TCI-018

Release-blocking.

**Given** 팩토리 런과 레인 둘, 묶음 B 의 세 카드 `b1 b2 b3`(레인 `lane-1`)와 묶음이 아닌 대기 카드 `u1`
**When** `factory next` 가 반복된다
**Then** (a) `lane-1` 이 `b1` 을 임대하고, `b1` 이 병합되기 전에는 `b2` 를 받지 못하며(선행 미병합이 오류가 아니라 건너뜀), (b) `lane-2` 는 `b2`·`b3` 를 받지 못하고, (c) `b1` 이 로컬 병합된 뒤 `lane-1` 의 다음 임대가 `u1` 보다 `b2` 를 우선하며, (d) 묶음이 아닌 카드의 직렬 슬롯 의미와 선택 결과는 변경 전 골든과 같다.

필수 시험: `TestFactoryNextBundleSerialLane`, `TestFactoryBundleKeepsSerialSlot`, `TestFactoryAssignBundleOrderGuard`.

- **RED-now:** L18(대조군 C1).
- **Green path:** M4.
- **Mutant probe:** (MU-49) 레인 친화를 무시해 다른 레인이 받음 → (b) 실패. (MU-50) 묶음 멤버를 병렬 임대 → (a) 실패. (MU-51) 직렬 슬롯 의미 변경 → (d) 실패.
- **점화:** W 는 M4 종료, R 은 선택 호 변경, V·S 공통.

## AC-TCI-018 — 병합 동사는 운영자 호출 전용이고 기록하고 닫는다 (REQ-TCI-019)

**Covers**: maps REQ-TCI-019

Release-blocking.

**Given** queued 카드 `a`·`b`, picked 카드 `c`, 그리고 분석 경로
**When** `merge a b`, `merge a c`, `merge b a`(순환), 레인 세션의 `merge`, 그리고 `add`·`analyze`·`relate` 실행
**Then** (a) `merge a b` 뒤 `a` 의 본문에 `[merged from b] …` 절이 붙고 `b` 가 사유 `merged into a` 로 drop 되며 `merged-into` 소견(`b → a`)이 기록되고, (b) picked 카드·이미 병합된 카드·순환은 거절되고 큐 파일이 바이트 동일이며, (c) 레인 세션에서는 `todoRefuseLaneMutation` 이 거절하고, (d) `add`·`analyze`·`relate`·분석기가 병합 함수를 호출하지 않는다는 것이 호출 그래프 시험으로 강제된다.

필수 시험: `TestTodoMergeRecordsAndDrops`, `TestTodoMergeRefusesPickedAndCycles`, `TestTodoMergeNeverInvokedByAnalysis`, `TestTodoMergeRefusedInLane`.

- **RED-now:** L19, L27(대조군 C1).
- **Green path:** M4.
- **Mutant probe:** (MU-52) 병합이 큐 순서도 바꿈 → (a) 의 순서 단언 실패. (MU-53) `relate` 가 병합을 부름 → (d) 실패. (MU-54) 병합이 관계 기록 없이 drop 만 함 → (a) 실패.
- **점화:** W 는 M4 종료와 sync-audit, R 은 병합 함수의 호출자 추가, V·S 공통.

## AC-TCI-019 — 허브 파일을 담은 카드는 체인으로 직렬화된다 (REQ-TCI-020)

**Covers**: maps REQ-TCI-020

Release-blocking.

**Given** 기준선 기록의 허브 목록(고정 입력으로 `internal/template/catalog.yaml` 포함)과 예상 파일에 그 경로를 담은 열린 카드 둘
**When** 둘째 카드의 팩토리 레코드가 만들어지고 선택이 돈다
**Then** (a) 둘째의 `after` 가 첫째로 채워지고 첫째가 로컬 병합되기 전에는 임대되지 않으며, (b) keep-set 판정이 파일 겹침을 읽지 않는다는 것이 시험으로 고정되고(`TestFactoryKeepSetReadsNoFileOverlap`), (c) 허브가 아닌 경로만 겹치는 카드는 체인이 되지 않는다.

필수 시험: `TestFactoryAssignBundleHubChain`, `TestHubFileListFromBaseline`, `TestFactoryKeepSetReadsNoFileOverlap`.

- **RED-now:** L20(대조군 C1).
- **Green path:** M4.
- **Mutant probe:** (MU-55) 겹침을 keep-set 안에서 읽음 → (b) 실패. (MU-56) 비허브 겹침도 체인 → (c) 실패.
- **점화:** W 는 M4 종료, R 은 허브 목록·레코드 생성 시점 변경, V·S 공통.

## AC-TCI-020 — M5 모드 선택은 게이트 명령이 하고 정확히 한 모드가 성립한다 (REQ-TCI-021, -022)

**Covers**: maps REQ-TCI-021, REQ-TCI-022

Release-blocking. 이 SPEC 이 항상 초록으로 만들 수 있도록 두 모드의 합집합이 아니라 **게이트가 가리킨 모드**가 성립하는 것이 기준이다.

**Given** M5 시작 직전의 게이트 세 명령 — 대상 `git rev-list --first-parent --count -E -i --grep='^merge[( :]+(card )?t1453' HEAD`, 대조군 t1448·t1344 (L21, C8, C9). 대조군이 하나라도 0 이면 게이트는 "미측정 = 미충족"이다
**When** M5 가 끝난다
**Then** 정확히 한 모드가 성립한다. **Mode A**(대상 ≥ 1, 대조군 둘 모두 ≥ 1): `.claude/rules/moai/workflow/card-issuance.md` 와 템플릿 사본이 있고 여섯 섹션(카드 크기, 후속 지적 규칙, 파생 깊이, 동시 진행 한도, 발행 체크리스트, 부채 대장)을 가지며 각 수치가 기준선 기록의 값과 같다(`TestCardIssuanceRuleValuesMatchBaseline`). **Mode B**(그 외): 여섯 규칙 파일과 템플릿 사본이 카드 브랜치의 병합 기준(읽는 시점에 `git merge-base develop HEAD`)에 대해 변하지 않았고(`git diff --name-only <그 SHA> -- <여섯 파일과 사본>` 이 비어 있다), `.moai/reports/t1454/m5-draft/` 가 삽입 위치 앵커와 함께 있고, `progress.md` §E.2 에 게이트 세 명령의 출력과 후속 카드 문안이 있다. 두 모드의 산출물이 함께 있으면(편집과 초안) 실패다.

- **RED-now:** L21(게이트 0 — 정보용으로 Mode B 선택을 가리킨다), L22(규칙 파일 부재), L23(초안 부재). 두 모드의 산출물이 모두 도착 시 없으므로 붉다.
- **Green path:** M5 — 게이트가 열려 있으면 Mode A, 닫혀 있으면 Mode B. 핀의 게이트 값은 닫힘이므로 지금 읽으면 Mode B 가 성립해야 한다.
- **Mutant probe:** (MU-57) 게이트가 닫혔는데 규칙 파일을 편집(Mode A 산출물) → Mode B 의 diff 가 비지 않아 실패. (MU-58) 게이트 명령을 대조군 없이 읽어 눈먼 선택자로 열림 판정 → 대조군 읽기 기록 부재로 실패. (MU-59) 초안에 앵커가 없음 → 앵커 검사 실패. (MU-60) 수치를 기준선 기록 대신 SPEC 에서 복사 → 값 일치 시험 실패.
- **점화:** W 는 M5 시작 직전과 첫 커밋 직전, 그리고 sync-audit, R 은 t1453 착지로 게이트 값이 바뀌는 사건, V 는 게이트 출력과 종료 코드, S 는 게이트 읽기를 `progress.md` §E.2 에 매번 기록하는 것(읽지 않으면 모드 선택의 근거가 없다).

## AC-TCI-021 — Mode A 의 예산·미러·생성물 정합 (REQ-TCI-021)

**Covers**: maps REQ-TCI-021

Release-blocking, **Mode A 에서만**. Mode B 에서는 "해당 없음(Mode B)"으로 기록하고 AC-TCI-020 의 Mode B 가 같은 파일을 건드리지 않는다는 것을 보증한다.

**Given** Mode A 가 성립한 M5 커밋
**When** 크기와 가드를 읽는다
**Then** (a) `wc -m .claude/skills/moai/workflows/gtd.md` ≤ 40000(도착 시 40037), (b) 상시 로드 `kanban-dispatch.md` 의 문자 수와 바이트 수가 병합 기준 blob 보다 크지 않다(기준은 읽는 시점에 `git merge-base develop HEAD` 로 구하고 핀하지 않는다, 핀 값은 G4·G5), (c) `kanban-dispatch-detail.md` 가 43,138자(G6)보다 크지 않다, (d) 바이트 동일 쌍(`gtd.md`, `kanban-dispatch-detail.md`, `kanban-dispatch-mechanics.md`, `manager-todo.md`)은 `cmp` 가 같고, 분기된 쌍(`kanban-dispatch.md`, `sync-auditor.md`)은 분기 hunk 수가 편집 전과 같다(G7 에서 `kanban-dispatch.md` 는 1 hunk), (e) `card-issuance.md` 가 40,000자 이하이고 `paths:` 패턴 어느 것도 `kanban-dispatch*` 글롭과 자기 매칭하지 않으며, (f) `make agents-emit` 으로 생성 Codex TOML 이 갱신되고 `catalog.yaml` 해시가 생성기로 다시 만들어져 `TestManifestHashFormat`·`TestCatalogHashCoversSkillSubfiles`·`TestGoldenCommittedArtifactsMatchEmission` 이 통과하며, (g) 템플릿 가드(카드 id 누출, 중립성, 경로 고정, 문서 정합, `TestDeclaredRuleMirrorForks`)가 통과한다.

- **RED-now:** L24(`gtd.md` 40,037자 — 한도 초과, 붉은 이유는 이 작업이 `gtd.md` 를 고치며 한도 안으로 되돌려야 하기 때문).
- **Green path:** M5 Mode A — `wc -m` 가 40000 이하를 출력.
- **Mutant probe:** (MU-61) `gtd.md` 를 한도에 맞추려 findings-source 열거 문장을 삭제 → 문서 정합 시험 실패. (MU-62) 로컬만 편집 → (d) 의 `cmp` 실패. (MU-63) 상시 로드 스텁에 내용을 이동해 예산 충족 → (b) 실패. (MU-64) 템플릿 사본에 카드 id 추가 → 누출 시험 실패. (MU-65) 카탈로그 미재생성 → 해시 시험 두 개 실패(AUTO-PICK 계획이 관측한 같은 모양).
- **점화:** W 는 M5 종료, R 은 규칙 파일 증가, V·S 공통(+ `wc -m` 출력).

## AC-TCI-022 — 웹 관계 그래프 보기 (REQ-TCI-023)

**Covers**: maps REQ-TCI-023

Release-blocking.

**Given** live·dropped·보관 카드와 각 종류의 관계를 담은 고정 큐, 그리고 노드 상한을 넘는 큰 고정 큐
**When** `GET /todo?view=graph` 와 `POST /todo?view=graph` 가 요청된다
**Then** (a) 200 이고 모든 고정 카드(보관·dropped 포함)의 노드와 각 관계 종류의 간선이 서버 렌더 HTML/SVG 로 나온다, (b) POST 는 405 이고 큐 락 파일·큐 파일이 바뀌지 않는다, (c) 큰 큐는 상한을 넘으면 "N개 생략" 문구와 함께 상한 개수의 노드만 낸다, (d) 응답에 외부 호스트 참조(`http`로 시작하는 `src`/`href`)가 없고 그래프가 참조하는 자산은 모두 임베드 파일 시스템에 있으며, (e) 새 문자열은 모든 로케일에 키가 있다(i18n 거버넌스 시험 통과), (f) 같은 입력은 같은 출력이다.

필수 시험: `TestTodoGraphViewRendersRelations`, `TestTodoGraphViewReadOnly`, `TestTodoGraphViewBounded`, `TestTodoGraphAssetsEmbedded`.

- **RED-now:** L25(대조군 C6).
- **Green path:** M6.
- **Mutant probe:** (MU-66) 보기 렌더가 락을 잡음 → (b) 실패. (MU-67) CDN 링크 → (d) 실패. (MU-68) 노드 무제한 → (c) 실패. (MU-69) 로케일 키 누락 → (e) 실패.
- **점화:** W 는 M6 종료, R 은 보기 렌더의 쓰기·네트워크 추가, V·S 공통.

## AC-TCI-023 — 기존 `/todo` 는 그대로다 (REQ-TCI-024)

**Covers**: maps REQ-TCI-024

**회귀 가드**(M6 시작 시 변경 전 렌더러에 대해 먼저 쓰므로 도착 시 초록).

**Given** `view` 매개변수 없는 `GET /todo` 와 같은 큐
**When** M6 이후 렌더한다
**Then** 응답 바이트가 M6 시작 시 기록한 기준과 같고 기존 `todo_route_test.go`·정렬·상세·live 시험이 통과한다(`TestTodoPageUnchangedWithoutViewParam`).

- **RED-now:** 해당 없음(도착 시 초록).
- **Mutant probe:** (MU-70) 기본 보기에 그래프 패널을 끼움 → 기준 바이트가 달라 실패.
- **점화:** W 는 M6 종료와 sync-audit, R 은 기본 렌더 변경, V·S 공통.

## AC-TCI-024 — 대상 SPEC 개정이 sync 에서 선언된다 (Amendments A1~A6)

**Covers**: maps REQ-TCI-012, REQ-TCI-018, REQ-TCI-019, REQ-TCI-023 (개정 행의 근거 요구)

Release-blocking; sync 단계에서 확인한다.

**Given** sync 단계, 결정 D2·D5·D6·D14 의 판정에 따라 적용으로 남은 개정 행(`spec.md` `## Amendments`)
**When** manager-spec 재위임으로 대상 SPEC 의 개정 커밋이 만들어진다
**Then** (a) 적용된 대상 SPEC 의 `spec.md` 가 `amendment_of:`(자기 참조)와 `## Amendments` 절을 가지며 절이 `prior_completed_sha` 로 이 SPEC 의 행과 같은 값을 적고(ANALYSIS `b6716a748`, AUTO-PICK `4293b2d73`, RELATION-PICKUP-FILTER `dcf744fdc`, GTD-AUTONOMY `5ec516165ef5c5e31cdf372b28e6e538f8347719`, WEB-TODO-QUEUE `f1e71db4b`), (b) 상태가 `completed → in-progress → completed`(sync 커밋)로 전이하고 `moai spec audit` 가 그 SPEC 에 `SyncStatusDrift` 를 내지 않으며, (c) 적용하지 않기로 한 조건부 행(A1·A5 등)은 "미적용"으로 기록돼 대상 SPEC 이 편집되지 않는다.

- **RED-now:** L26(대조군 C10: `amendment_of` 를 가진 SPEC 이 존재).
- **Green path:** sync 단계의 개정 커밋.
- **Mutant probe:** (MU-71) 대상 SPEC 본문을 `amendment_of` 없이 편집 → `moai spec audit` 가 잡는다. (MU-72) `prior_completed_sha` 가 progress.md 의 값과 다름 → (a) 실패. (MU-73) 미적용이어야 할 조건부 행을 적용 → (c) 실패.
- **점화:** W 는 sync 단계, R 은 대상 SPEC 의 무선언 편집, V 는 `moai spec audit` 와 manager-spec 의 개정 커밋 검토, S 는 대상 SPEC 이 `completed` 인데 이 SPEC 의 개정 행이 적용 기록 없이 남는 상태를 sync-audit 가 읽는 것.

## 변이 탐침 요약

| 변이 | 위반 | 잡는 기준 |
|---|---|---|
| MU-1~3 빈 디렉터리, 명령 없는 숫자, 재현되지 않는 값 | REQ-TCI-001 | AC-TCI-001 |
| MU-4~7 stdout 출력, live 만 조회, 접두사 미제거, 상위 4개 | REQ-TCI-002 | AC-TCI-002 |
| MU-8~12 live 만 이웃, 하한 무시, draft SPEC 포함, 빈 겹침을 none, 분류기 호출 | REQ-TCI-002, -006 | AC-TCI-003, -007 |
| MU-13~15 락 안 탐침, 탐침 오류 전파, 상한 없음 | REQ-TCI-003 | AC-TCI-004 |
| MU-16~18 id 소비, 소견 기록, 중복에서 비제로 종료 | REQ-TCI-004 | AC-TCI-005 |
| MU-19~21 제시를 첫 줄 앞에, engage 소견 기록, 빈 제시에 빈 줄 | REQ-TCI-005 | AC-TCI-006 |
| MU-22~23 분류기 변경, 같은 파일에 함수 추가 | REQ-TCI-006 | AC-TCI-007 |
| MU-24~27 한 표에만 컬럼, 재구성 목록에 포함, retrofit 순서, parity 무시 | REQ-TCI-007~009 | AC-TCI-008 |
| MU-28 `omitempty` 없음 | REQ-TCI-007 | AC-TCI-009 |
| MU-29~31 접두사에만 사유, 접두사 제거, 0 시각 | REQ-TCI-010 | AC-TCI-010 |
| MU-32~34 merge 처분이 drop, reject 가 삭제, 기본 처분 | REQ-TCI-011 | AC-TCI-011 |
| MU-35~37 저장 행 재작성, 매핑 누락, 표시 변경 | REQ-TCI-012, -017 | AC-TCI-012 |
| MU-38~40 `blocks` 만 순환 검사, 메모리 카디널리티, 거절 전 일부 쓰기 | REQ-TCI-013 | AC-TCI-013 |
| MU-41~42 해석기가 씀, 한 방향 | REQ-TCI-014 | AC-TCI-014 |
| MU-43~45 깊이 무시, 비결정 순회, 방문 집합 없음 | REQ-TCI-015 | AC-TCI-015 |
| MU-46~48 큐에서 간선 생성, 비결정 출력, 지문 미등록 | REQ-TCI-016 | AC-TCI-016 |
| MU-49~51 레인 친화 무시, 병렬 임대, 슬롯 의미 변경 | REQ-TCI-018 | AC-TCI-017 |
| MU-52~54 순서 변경, `relate` 가 병합 호출, 관계 없이 drop | REQ-TCI-019 | AC-TCI-018 |
| MU-55~56 keep-set 이 겹침 읽음, 비허브도 체인 | REQ-TCI-020 | AC-TCI-019 |
| MU-57~60 게이트 닫힘에 편집, 대조군 없는 판정, 앵커 없는 초안, 수치 복사 | REQ-TCI-021, -022 | AC-TCI-020 |
| MU-61~65 열거 문장 삭제, 로컬만 편집, 스텁 이동, 카드 id 추가, 카탈로그 미재생성 | REQ-TCI-021 | AC-TCI-021 |
| MU-66~69 락 획득, CDN, 무제한, 로케일 누락 | REQ-TCI-023 | AC-TCI-022 |
| MU-70 기본 보기에 패널 | REQ-TCI-024 | AC-TCI-023 |
| MU-71~73 무선언 편집, SHA 불일치, 미적용 행 적용 | Amendments | AC-TCI-024 |

**도착 시 통과하는데 수용하는 변이**(잡히지 않는 것과 이유):

| 변이 | 위반 | 잡히지 않는 이유와 수용 |
|---|---|---|
| MU-74 큐 쪽 baseline figure 의 값을 바꿔 적음 | REQ-TCI-001 | 비공개 DB 스냅숏이라 제3자가 재실행할 수 없다 — `queue-snapshot:` 귀속만 하고 재현은 주장하지 않는다 |
| MU-75 `--dry-run` 이 레인 세션에서 쓰이지 않음 | REQ-TCI-004 | `add` 는 레인 가드가 거절한다(읽기 전용 목록에 없다) — 관찰했고 정책 의도 여부는 미확인(공백 13) |
| MU-76 선행 미병합 후보가 오류가 아니라 건너뛰는 현재 동작이 이미 옳은 경우 | REQ-TCI-018 | M4 의 첫 시험이 현재 동작을 관측해 기록한다 — 읽기 추정이라 RED 일 수도 초록일 수도 있다 |
| MU-77 웹 그래프가 큰 큐에서 느림 | REQ-TCI-023 | 노드 상한과 결정성만 시험하고 브라우저 렌더 시간은 재지 않았다(공백 6) |

개수: 변이 77개 명명 — 73개는 기준이 잡고 4개(MU-74~77)는 이유와 함께 수용한다.

## 경계 사례

- 한국어 본문처럼 공백 없이 이어진 토큰은 토큰 집합이 작아져 Jaccard 가 흔들린다 — 제시는 점수와 척도를 항상 함께 적는다.
- dropped 카드 본문의 접두사가 대시 변형(`—` 가 아닌 `-`)이면 `stripTodoDropMarker` 가 벗기지 못할 수 있다 — 벗기지 못한 본문은 접두사를 단 채 비교되고 점수가 낮게 나온다(보수적).
- 정규화 본문이 같은 카드가 live 와 보관 양쪽에 있으면 둘 다 `exact` 로 보이되 상위 3개 상한은 지킨다.
- git 이 없거나 탐침이 실패하면 해당 항목이 `unmeasured` 이고 admit 은 성공한다.
- `--dry-run` 과 `--pick`/`--force` 를 함께 주면 dry-run 이 우선이고 쓰기는 없다(둘 중 하나가 무시됐음을 stderr 에 한 줄 적는다).
- 폴스루(`moai todo <단어…>`)는 플래그를 받지 않으므로 `--dry-run` 은 `add` 서브커맨드에서만 쓴다.
- 워크트리 세션에서도 큐 루트는 primary 체크아웃으로 해석되므로 제시는 같은 큐를 읽는다.
- 시험은 라이브 `moai todo add` 가 아니라 `go test` 도구로만 한다(레인 가드가 라이브 호출을 거절하고, 시험 도구는 레인 변수를 지운다).

## 품질 게이트

- TRUST 5: 시험을 먼저 쓰고 RED 를 관측한 뒤 GREEN(회귀 가드 AC-TCI-007·009·023 만 예외), 변경 함수 커버리지 ≥ 85%, CI 가 쓰는 버전의 `golangci-lint` 청결, `GOOS=windows GOARCH=amd64 go build` 를 변경 패키지에 한해 통과.
- 문서 게이트: `moai spec lint SPEC-TODO-CARD-ISSUANCE-001` 과 `--strict` 청결(plan 단계), M5 Mode A 의 템플릿 가드 전부 통과.
- MX: 새 고팬인 함수(관계 정규화·제약 검증, 읽기 전용 이웃 조회)에 `@MX:ANCHOR`(`@MX:REASON` 필수), 탐침 병렬 코드가 있으면 `@MX:WARN`, 상수에 `@MX:NOTE`.

## 완료 정의

1. AC-TCI-001~006, -008, -010~022, -024 가 통과하고(출시 차단 21개, 단 -021 은 Mode A 에서만, 아니면 "해당 없음(Mode B)"으로 기록하고 -020 이 모드를 보증한다) RED-now 셀이 관측되고 green-path 출력이 `progress.md` §E.2 에 기록된다. AC-TCI-007·009·023 은 회귀 가드로 유지된다.
2. M0 기준선 커밋이 어느 변경 커밋보다 앞서 커밋 그래프에 있다(`verification-claim-integrity.md` §2.3).
3. 각 마일스톤의 시험 선택 수(`-list`)와 `git grep -c -F` 짝이 `progress.md` §E.2 에 있다.
4. M5 커밋 본문(Mode A)에 상시 로드 전후 바이트·문자와 이 규칙이 필요 없는 세션의 비용 문장이 있다.
5. 완료 보고가 보존된 분기(`kanban-dispatch.md` 한 문장, `sync-auditor.md` 18줄), 게이트 모드(A/B)와 Mode B 의 경우 "요구는 충족했으나 규칙은 아직 효력이 없다"는 사실, 그리고 열린 잔여 위험(spec §G)을 명시한다.
6. `TestACCounterFullCorpusMatchesBaseline` 이 통과한다(이 SPEC 의 acceptance.md 는 스냅숏에 없어 보고만 되고 실패하지 않는다; 기준선 재생성이 필요하면 그것은 plan 커밋의 몫이다).
