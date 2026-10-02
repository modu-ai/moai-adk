# acceptance.md — SPEC-SESSION-ANCHOR-ATTR-001

> **판정 규칙(전 AC 공통)**: 검증 명령의 판정은 exit 코드가 아니라 **관측된 스윕 내용**으로 한다. `go test` 출력에 `[no tests to run]`이 보이면 그 명령은 무엇을 통과시켰는지 알 수 없으므로 **실패로 처리**한다. 모든 `-run` 선택자는 실제 Go 테스트 함수명(prefix) 기준이며, 저작 시점(2026-09-30, 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1339`, HEAD `a1ed20c0b`)에 본 문서에 적은 명령을 실제 실행해 비어 있지 않은 스윕을 확인했다 — 실행 관측치는 각 셀과 §D.0에 인용.

## §D.0 채택표 (RED-now / green-path — 신규 동작 AC용, [HARD] 두-셀 규칙)

| AC | 구분 | RED-now 셀 (명령 · 관측 출력 · exit · 측정 트리 — 2026-09-30 본 실행) | green-path 셀 (전환 마일스톤) |
|----|------|----------------------------------------------------------------------|-------------------------------|
| AC-001 | 신규 동작 (행 필드) | `grep 'WorktreeGuardRefusal' /Users/goos/MoAI/moai-adk-go/.moai/harness/usage-log.jsonl \| grep -c session_id` → `0` (전체 4,366행, exit 1) — session_id 실은 행 0건; primary 체크아웃 런타임 파일 | M1 — failure_observer 행에 session_id·cwd·트리 경로 추가 |
| AC-002 | 신규 동작 (unknown 마커) | `go test ./internal/hook/ -run '^TestClassifyError_GuardRefusal_UnknownMarkers$'` → `ok … [no tests to run]`, exit 0 — 대상 테스트 미존재 (빈 스윕 = RED의 증거) | M1 — `TestClassifyError_GuardRefusal_UnknownMarkers` 신설 후 §D 매트릭스 명령이 비지 않는 스윕으로 전환 |
| AC-003 | 신규 동작 (감사 행) | `ls /Users/goos/MoAI/moai-adk-go/.moai/logs/ \| grep -ci relocat` → 매치 0, exit 1 — 재배치 감사 로그 부재 | M2 — RelocateSession 감사 행 + 감사 로그 신설 |
| AC-004 | 신규 동작 (소유 플래그) | `grep -c 'anchor_relocation_guard' internal/config/defaults.go` → `0` — 플래그 판정·config 키 모두 미존재 | M2 — REQ-SAA-004 케이스표 구현 |
| AC-005 | 신규 동작 (opt-in blocking) | `grep -c 'anchor_relocation_guard' internal/config/defaults.go` → `0`, exit 1 — config 키 미존재 | M2 — `workflow.anchor_relocation_guard.enabled` 키 + 두-팔 테스트 |
| AC-006 | **regression-guard** (전환 없음) | 현행 GREEN이 곧 기준 (본 실행): `go test ./internal/session/ -run '^(TestRelocateSession_UpdatesCwd\|TestRelocateSession_MissingEntryNoOp\|TestRelocateSession_EmptyArgsReject)$'` → 3 RUN, `ok`; `go test ./internal/hook/ -run '^(TestRelocateReachesPrimaryRegistryWhenEveryCandidateIsInsideOneWorktree\|TestRelocateFailOpen\|TestRelocateCandidatesStopAtHomeInPassTwo)$'` → 8 RUN, `ok` | — GREEN 유지가 판정 |
| AC-007 | 신규 동작 (트레이스 on) | `grep -rn 'MOAI_ANCHOR_TRACE' internal/ --include='*.go' \| wc -l` → `0` | M3 — 트레이스 경로 + envkeys 상수 + 신설 테스트 |
| AC-008 | 신규 동작 (트레이스 off, 부정 경로) | AC-007과 동일 RED — 트레이스 부재 상태가 곧 기본 동작. 단, green-path 전환 후에는 AC-007을 **양성 대조**(같은 fixture, 게이트 on)로 병행해야 0출력이 증거가 된다 | M3 |
| AC-009 | **regression-guard** (전환 없음) | 현행 GREEN 기준 (본 실행): `TestLiveAnchoredSessions` 6 RUN · `TestRefuseMutation` 1 RUN · `TestClassifyError_GuardRefusal\|TestFormatMessage_GuardRefusal` 8 RUN — 전부 `ok`, exit 0 | — GREEN 유지가 판정 |
| AC-010 | 신규 동작 (문서 4요소) | `grep -c 't1339' internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md` → `0`, exit 1 — 4요소 미문서화 | M4 — 템플릿 선수정 + `make build` 동기 |

## §D AC Matrix

| AC | 요구 | 시나리오 (Given-When-Then) | 검증 방법 (실행 확인된 선택자 + 빈-스윕 가드) |
|----|------|---------------------------|-----------|
| AC-001 | REQ-SAA-001 | **Given** 훅 층이 도구 실패를 `WorktreeGuardRefusal`로 분류하고, **When** failure observer가 로그 행을 기록하면, **Then** 행에 `session_id`, 해상 cwd, 거부 인용 트리 경로가 `subject`·`context_hash`와 함께 존재한다. | `go test ./internal/hook/ -run '^(TestClassifyError_GuardRefusal\|TestClassifyError_GuardRefusal_BeatsOOM\|TestClassifyError_GuardRefusal_AnchorOnly\|TestFormatMessage_GuardRefusal_NamesTheGap)$' -v` — 저작 시 관측: **8 RUN, ok, exit 0**. 행-필드 단정은 M1에서 이 테스트군에 추가한다. 가드: `=== RUN` 행 수 ≥ 1, `[no tests to run]` 출력 시 실패. |
| AC-002 | REQ-SAA-002 | **Given** 기록 시점에 세션 식별자 또는 트리 경로를 해상할 수 없고, **When** 행을 기록하면, **Then** `session_id: "unknown"` / `cwd: "unknown"` 마커로 행이 유지되고 행 유실은 0이다. | green-path 전환 후: `go test ./internal/hook/ -run '^TestClassifyError_GuardRefusal_UnknownMarkers$' -v` — 비지 않는 스윕 + 마커 단정. RED-now §D.0. 가드 동일. |
| AC-003 | REQ-SAA-003 | **Given** 세션 항목이 레지스트리에 존재하고, **When** `RelocateSession`이 cwd를 재작성하면, **Then** 감사 행(session_id, 이전 cwd, 새 cwd, 트리거 훅, timestamp)이 추가된다. | `go test ./internal/session/ -run '^(TestRelocateSession_UpdatesCwd\|TestRelocateSession_MissingEntryNoOp\|TestRelocateSession_EmptyArgsReject)$' -v` — 저작 시 관측: **3 RUN, ok, exit 0**. 감사 행 필드 단정은 M2에서 추가. 가드 동일. |
| AC-004 | REQ-SAA-004 | **Given** 재배치 목표 트리의 git worktree lock이 REQ-SAA-004 소유 비교 규칙의 "other live card" 케이스(lock reason card-id 불일치 + holder 생존)에 해당하고, **When** 재배치가 실행되면, **Then** advisory 플래그 행이 감사 로그에 기록되고 재배치는 계속 진행된다. **Given** "unreadable" 케이스(lock reason 파싱 실패)이면 **Then** 플래그 없이 `owner: unreadable` 감사 행만 남는다(fail-closed 판독 준수). **Given** "registry-only" 케이스(lock 부재)이면 **Then** advisory 플래그 + 진행. **Given** "self-owned"(lock pid == 세션 항목 pid)이면 **Then** 플래그 없음. | green-path 전환 후: M2 신설 테스트(`-run 'TestRelocateOwnershipFlag'` 계열) — 케이스표 4팔 각각 단정. RED-now §D.0. 가드 동일. |
| AC-005 | REQ-SAA-005 | **Given** `workflow.anchor_relocation_guard.enabled: true`(opt-in)이고, **When** 플래그된 재배치가 발생하면, **Then** 재배치가 거부되고 거부가 감사 로그에 기록된다. **Given** false(기본)이면 **Then** AC-004의 기본 동작과 동일(advisory+진행). | green-path 전환 후: M2 신설 두-팔 테스트 — config on/off 각 팔 실행, off 팔 출력이 AC-004 기본 케이스와 동일함을 단정. RED-now §D.0. 가드 동일. |
| AC-006 | REQ-SAA-006 | **Given** 어떤 재배치 시나리오든, **When** 후보 탐색이 실행되면, **Then** 두-패스 순서(상향 보행 → primary 레지스트리)와 fail-open 오류 동작이 기존과 동일하다. | `go test ./internal/session/ -run '^(TestRelocateSession_UpdatesCwd\|TestRelocateSession_MissingEntryNoOp\|TestRelocateSession_EmptyArgsReject)$' -v` (저작 시 **3 RUN, ok**) + `go test ./internal/hook/ -run '^(TestRelocateReachesPrimaryRegistryWhenEveryCandidateIsInsideOneWorktree\|TestRelocateFailOpen\|TestRelocateCandidatesStopAtHomeInPassTwo)$' -v` (저작 시 **8 RUN, ok**) — **기존 테스트 파일 무수정** GREEN. regression-guard(§D.0). 가드 동일. |
| AC-007 | REQ-SAA-007 | **Given** `MOAI_ANCHOR_TRACE=1`이 설정되고, **When** 앵커 의사결정(checkBranchState 앵커 읽기, RelocateSession, AnchorDecision 판정)이 발생하면, **Then** 의사결정당 1행의 상세 로그가 session_id+pid+cwd+timestamp와 함께 출력된다. | green-path 전환 후: M3 신설 테스트(`-run 'TestAnchorTrace'` 계열) — 행당 필드 단정. RED-now §D.0. 가드 동일. |
| AC-008 | REQ-SAA-008 | **Given** `MOAI_ANCHOR_TRACE`가 unset 또는 falsy이고, **When** 앵커 의사결정이 발생하면, **Then** 트레이스 출력은 0이다. | green-path 전환 후: M3 신설 부정-경로 테스트 — 출력 0 단정, **AC-007 양성 대조 병행 필수**(같은 fixture, 게이트 on — 0출력만으로는 부재의 증거가 안 됨). RED-now §D.0. |
| AC-009 | REQ-SAA-009, REQ-SAA-010, REQ-SAA-011, REQ-SAA-012, REQ-SAA-013 | **Given** 본 SPEC의 모든 변경이 착지한 뒤, **When** 회귀 게이트를 실행하면, **Then** (a) `worktree_guard_refusal_test.go` 변이 2건 살생 유지, (b) `checkBranchState` git-context는 `input.CWD` 판독 유지, (c) 감사 로그 디렉터리는 `CLAUDE_PROJECT_DIR` 앵커 유지, (d) env 상수는 envkeys.go 선언만 존재, (e) `LiveAnchoredSessions` 읽기 경로·처분 가드 시맨틱 관련 기존 테스트 무수정 GREEN (REQ-SAA-011 — R1 게이트), (f) `RefuseMutationFromNonCanonicalTree`/`MOAI_HOME` 변이 대상 관련 기존 테스트 무수정 GREEN + 계측기가 Go 훅 층에만 배치됨 (REQ-SAA-013). | (a)-(c) `go test ./internal/hook/ -run '^(TestClassifyError_GuardRefusal\|TestClassifyError_GuardRefusal_BeatsOOM\|TestClassifyError_GuardRefusal_AnchorOnly\|TestFormatMessage_GuardRefusal_NamesTheGap)$'` (저작 시 **8 RUN, ok**) + 기존 branch-guard 테스트 무수정 GREEN. (d) `grep -rn 'MOAI_ANCHOR_TRACE' internal/ --include='*.go'` — envkeys.go 선언 외 0건. (e) `go test ./internal/session/ -run '^(TestLiveAnchoredSessions_CallerRegistrySource\|TestLiveAnchoredSessions_KeepsLivePIDWithStaleHeartbeat\|TestLiveAnchoredSessions_DropsDeadPIDStaleHeartbeat\|TestLiveAnchoredSessions_KeepsFreshHeartbeatDespiteDeadPID\|TestLiveAnchoredSessions_FiltersForeignHostAndCWD\|TestLiveAnchoredSessions_MissingRegistryFailsOpen)$' -v` (저작 시 관측: **6 RUN, ok**). (f) `go test ./internal/homestate/ -run '^TestRefuseMutationFromBareLinkedWorktree$' -v` (저작 시 관측: **1 RUN, ok**) + 배치 검증은 W1~W3 변경 파일 목록이 internal/hook·internal/session·internal/config로 한정됨을 `git diff --stat`으로 확인(훅 사본 경로 부재). regression-guard(§D.0). |
| AC-010 | REQ-SAA-014 | **Given** W4 문서 갱신의 템플릿 선수정이 착지하고, **When** `make build` 후 로컬 동기가 완료되면, **Then** worktree-integration-ops.md에 (a) 복구 절차(ExitWorktree keep + 재-EnterWorktree), (b) 런타임 경계 선언, (c) 무음 경로-없는-명령 위험(t741), (d) 정정된 t1337/t1339 기록(재시작 없이 지속) 4요소가 존재한다. | 4요소 presence grep (템플릿·로컬 양면) + **일치 판정 기제**: `diff internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md .claude/rules/moai/workflow/worktree-integration-ops.md` → 차이 0 (Template-First 동기의 기계적 증거) + `go test ./internal/template/ -run '^TestTemplateNeutralityAudit$' -v` (저작 시 관측: **7 RUN, ok** — 하위 검사 포함; `…C8Preserve` 변형은 W4 착지 시 패키지 전체 스위트가 커버). RED-now §D.0. |

## §D.1 엣지 케이스

- 세션 식별 불가 환경(훅 페이로드에 session_id 부재) → AC-002의 unknown 마커 경로.
- 감사 로그 디렉터리 기록 불가(권한/디스크) → fail-open: 재배치 자체는 계속. (구현 참고 노트 — REQ 아님: 기록 실패는 훅 프로세스 stderr로 1회 출력한다. 런타임이 이 스트림을 `~/.moai/logs/hook-stderr.log`로 수집하는 것이 관측된 표면이며, 신규 통지 기능을 REQ로 요구하지 않는다.)
- lock reason 판독 불가 트리 → REQ-SAA-004 "unreadable" 케이스: 기존 fail-closed("treated as anchored") 준수 — 플래그 오판(생존 카드 아님 판정) 금지.
- 동일 트리 재배치(from == to) → 감사 행은 기록하되 플래그 없음.
- Windows(pid 프로브 무측정) → 트레이스/감사 행은 pid 값을 그대로 기록, 생존 판정 분기는 기존 heartbeat 폴백 유지.

## §D.2 품질 게이트 (TRUST 5 / 회귀 게이트)

- 기존 hook·session 패키지 테스트 전체 GREEN 유지 (baseline 대비).
- `worktree_guard_refusal_test.go` 살생 변이 2건 유지 — REQ-SAA-010의 기계적 증거.
- `go test ./internal/template/... -run '^TestTemplateNeutralityAudit$'` GREEN (W4 템플릿 변경 대상; prefix 형태는 AC-010 셀 참조).
- E4 grep: env 상수 하드코딩 0건.
- 린트: CI 판 golangci-lint 버전으로 판정 (`feedback_lane_lint_must_use_ci_golangci_version`).

## §D.3 Definition of Done

- AC-001..AC-010 전부 PASS — 신규 동작 AC는 §D.0 green-path 전환 후, regression-guard AC(006·009)는 무수정 GREEN 유지로 (실행 증거 인용).
- REQ-SAA-001..014 전 항목 대응 AC 존재(AC-009가 REQ-SAA-009~013 불변식군을 (a)-(f)로 집행 가능하게 커버).
- 심각도 논거(spec.md §A.3)와 구현 산출물 간 모순 0.
- plan-audit PASS (0.80 이상).
