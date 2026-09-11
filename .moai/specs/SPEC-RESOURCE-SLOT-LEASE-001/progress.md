# SPEC-RESOURCE-SLOT-LEASE-001 — progress (card t607)

plan 단계 산출물을 트리 `c4ce42eca` @ `WT-heavy-test-slot`(worktree t607)에서 작성했다. Tier M. Status: in-progress(run 단계 M1 커밋에서 전환).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- 산출물(Tier M): spec.md, plan.md, acceptance.md, progress.md(이 파일).
- SPEC ID 정규식 검사 — 실행한 명령과 출력:
  `ID="SPEC-RESOURCE-SLOT-LEASE-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`
- ID 고유성: 이 워크트리 `.moai/specs/`에서 `RESOURCE`·`SLOT`·`LEASE` 이름을 가진 디렉터리는 `SPEC-SYNC-SHA-SLOT-FORMAT-001` 하나뿐이며 ID가 다르다.
- 프론트매터: 12개 정식 필드 + `tier: M`. `status: draft`. `phase`는 릴리스 대상(`"v3.2.0 target"`)이다.
- 요구사항: REQ-RSL-001..016(GEARS, IF/THEN 없음, Tier M 상한 16 이내).
- 인수 기준: AC-RSL-001..016(Tier M 상한 16 이내). 릴리스 차단 기준의 RED-now 셀은 acceptance.md 증거 원장(0.2.0 기준 EL-1..EL-8)을 인용한다.
- Out of Scope: `### Out of Scope — <주제>` H3 7개, 각각 `-` 항목 보유.
- spec lint — 이 트리(`c4ce42eca`)에서 빌드한 바이너리를 경로로 호출했다(설치본 `ed71054d3`은 HEAD의 조상이라 뒤처진 빌드다: `git merge-base --is-ancestor ed71054d3 HEAD` → 종료 코드 0):
  - `go build -o <scratch>/moai ./cmd/moai` → 성공
  - `<scratch>/moai spec lint --strict .moai/specs/SPEC-RESOURCE-SLOT-LEASE-001` → `✓ No findings — all SPEC documents are valid`
  - 음성 대조군: 같은 산출물의 사본에 `phase: plan`을 심고 같은 바이너리로 lint → `ERROR FrontmatterPhaseInvalid ... 1 error(s)`. lint가 이 디렉터리 모양을 실제로 읽는다는 증거다.
  - 참고: 요구 문장을 한국어로만 썼던 첫 초안은 설치본 lint에서 `ModalityUnjudged` 경고 16건을 받았다. 정본 요구 문장을 GEARS 영어로 바꾸고 한국어 설명을 하위 항목으로 옮긴 뒤 위 결과가 나왔다.
- M6(레인 문서 반영) 게이트 상태: `git merge-base --is-ancestor WT-acquire-branch-record develop` → 종료 코드 1(`WT-acquire-branch-record` = `f680dab46`, `develop` = `eb50af5a8`). 게이트 닫힘.

### plan-audit 1회차 수리 (v0.2.0)

plan-auditor 1회차는 FAIL 0.78이었다(Tier M 기준 0.80, 보고서 `.moai/reports/t607/plan-audit-iter1.md`, 커밋 `e50cfea93`). blocking D1-D6과 optional D7-D12를 모두 반영했다. 아래 측정은 이번 실행에서 트리 `e50cfea93`에 대해 했다.

- 코드 트리 불변 확인: `git diff --stat c4ce42eca HEAD -- internal cmd pkg` → 출력 없음. 문서 핀 `c4ce42eca`의 코드 측정이 그대로 유효하다.
- D1: 훅 루트 해석은 `internal/hook/path_resolve.go`:78-94에서 `CLAUDE_PROJECT_DIR` → `os.Getwd()`만 쓴다(직접 읽음). §B.3 정정, REQ-RSL-008(가드 루트 정규화), AC-RSL-016 추가.
- D2: 출처 `internal/kanban/integration_lock_cross_test.go`:46-63(500ms 풀림 타임아웃)과 :180-183(부모 pid 고정)을 직접 읽고 AC-RSL-001a를 다시 썼다.
- D4: lane 규칙 `.claude/rules/local/gitflow-lane-protocol.md` §8을 읽고 그 형태(`CARD_BASE=$(git merge-base develop HEAD)` + 대조군 + 병합 전 전용)로 바꿨다.
- D6: 언어 도구 토큰 목록의 기준선과 대조군을 쟀다 — `/usr/bin/grep -nwiE '<TOOL_TOKENS>' internal/template/templates/.moai/config/sections/workflow.yaml` → 출력 없음, 종료 코드 1(EL-7). 같은 목록으로 `/usr/bin/grep -cwiE '<TOOL_TOKENS>' internal/template/templates/.claude/rules/moai/languages/python.md` → `7`, 종료 코드 0(EL-8).
- D7: OQ-1은 감사 보고서 Evidence E3의 측정으로 닫았다. 런타임 버전은 E3에 없어 이번 실행의 `claude --version` → `2.1.268 (Claude Code)`을 참고로만 적었다.
- D9: M6 게이트 재측정 — `git merge-base --is-ancestor WT-acquire-branch-record develop` → 종료 코드 1. 이 시점 `git rev-parse --short develop` → `ac6c42c2d`(움직이는 ref의 측정 시점 값).
- spec lint(0.2.0): 트리 HEAD `e50cfea93`에서 `go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=e50cfea93" -o <scratch>/moai-e50 ./cmd/moai`로 빌드하고 경로로 호출했다. 바이너리의 `version` 출력에 커밋 `e50cfea93`가 찍힌다. `<scratch>/moai-e50 spec lint --strict .moai/specs/SPEC-RESOURCE-SLOT-LEASE-001` → `✓ No findings — all SPEC documents are valid`. 음성 대조군: 0.2.0 산출물 사본에 `phase: plan`을 심고 같은 바이너리로 lint → `1 error(s), 0 warning(s)`.
- 예산: REQ 16 / AC 16 유지. 옛 REQ-RSL-008(pid 0)은 REQ-RSL-005로, 옛 AC-RSL-002(뮤턴트 관측)는 AC-RSL-001b로, 옛 AC-RSL-016(CLI·교차 플랫폼)은 AC-RSL-003c로 흡수했다.

## §E.2 Run-phase Evidence

### M1 — RED 테스트와 API 표면 고정 (cycle_type=tdd)

- 기준 트리: HEAD `f95fade41`(로컬 develop `ac6c42c2d` 흡수 병합) 위의 작업 트리. 측정은 모두 M1 커밋에 들어간 파일 상태에서 했다. 예외 하나: kanban 회귀 실행(아래 R2)은 `slot_lease_cross_test.go`의 `fmt.Fprintf` 세 줄을 `_, _ =`로 바꾸기 전에 돌렸다. 바뀐 곳은 건너뛴 테스트만 부르는 헬퍼이며, 바꾼 뒤 lint와 kanban RED는 다시 돌렸다.
- 리드 제약(cli 슬롯): `internal/cli`를 컴파일·링크하는 명령은 하나도 돌리지 않았다. 돌린 패키지는 모두 `go list -test -deps <pkg> | grep -c 'moai-adk/internal/cli$'` → `0`을 확인했다(kanban 0, hook 0, config 0).

#### 스텁 방식 (compile-RED를 피한 이유)

Go에서 없는 심볼을 부르는 테스트는 패키지 전체의 테스트 빌드를 깨뜨린다. 그래서 테스트가 부르는 심볼만 스텁으로 선언했다. 스텁은 결정(이름, 기록 스키마, 경로)만 담고 동작은 없다.

| 파일 | 스텁 내용 | 동작 |
|---|---|---|
| `internal/kanban/slot_lease.go` | `SlotLease`/`SlotLeaseDisplacement`/`SlotLeaseRequest` 타입, 센티널 6개와 판정 함수, `AcquireSlotLease`·`ReleaseSlotLease`·`ReadSlotLease`·`ValidateSlotResourceName`·`ParseSlotLeaseMaxDuration`·`ResolveSlotLeaseRoot`, 테스트 끼어들기 훅 | 모든 연산이 `errSlotLeaseNotImplemented`(어떤 센티널도 아님), 메서드는 영값. Acquire는 nil 가드된 테스트 훅만 부른다 |
| `internal/hook/slot_lease_guard.go` | `slotLeaseViolationPrefix`, `checkSlotLease(input, hookRoot, cfg, advisory io.Writer)` | 항상 허용, 아무것도 쓰지 않음. `pre_tool.go` 배선은 M4 |
| `internal/config/types.go` | `WorkflowConfig.SlotLease` + `SlotLeaseConfig{Enabled, DefaultMaxDuration, Resources}` + `SlotLeaseResourceConfig{Commands, Invalid}` | 기본값(`defaults.go`)과 항목별 관대한 디코딩은 M4 |

결과: **이번 M1의 모든 RED는 assertion-RED다.** compile-RED인 테스트는 없다. 단, `internal/cli/slot_test.go`는 cli 슬롯 제약 때문에 컴파일조차 확인하지 않았다(아래 "미검증").

#### 고정한 결정 (테스트가 단언하는 것)

- 기록 경로 `<root>/.moai/state/slot-leases/<resource>.json`, 변경 락 `<resource>.mutation.lock`, 감사 로그 `<root>/.moai/logs/slot-lease-audit.jsonl`.
- 기록 키: `resource`, `session_id`, `session_name`, `pid`, `pid_source`(`session-owner`), `command`, `acquired_at`(RFC3339 UTC), `max_duration`(Go duration 문자열), `expires_at`(= 획득 + 상한), `displaced{session_id, session_name, pid, acquired_at, reason, at}`. `session_name`·`command`는 생략 시 빈 문자열로 **키가 남는다**.
- 감사 어휘: kanban 쪽 `event` ∈ {`acquire`, `refuse`, `takeover`(+`reason` stale/expired/force), `release`}, 가드 쪽 `event` ∈ {`guard-deny`, `allow-self`, `allow-stale`, `allow-expired`, `allow-unheld`, `fail-open`(+`reason` 비어 있지 않음)}. AC-RSL-011의 "기대 감사 사유"는 가드 줄의 `event` 필드다.
- 설정 키 `workflow.slot_lease.{enabled, default_max_duration, resources.<name>.commands}`, 기본값 `enabled: false`, 기본 상한 `30m`, 자원 없음.
- 상한 입력: 0·음수는 `ErrSlotLeaseBoundInvalid`, 문자열 `""`·`"0"`·`"-5m"`·`"abc"`·`"5"`도 같은 오류, `"5m"`은 통과(양성 대조).
- 자원 이름: `../escape`, `a/b`, `a\b`, `""`, 65자, `Demo`, `a b`, `..` 거부(획득·해제·조회·검증 모두) + 임시 루트 **바깥 부모까지** 파일 목록 불변. `demo`, `heavy-test-1`, 64자는 통과(양성 대조).
- **N1(plan-audit 2회차) 반영**: 공유 루트 해석 함수를 `kanban.ResolveSlotLeaseRoot(start)` 하나로 두고(리드 권고대로 cli 밖 패키지), primary / primary 하위 디렉터리 / 링크된 워크트리 → primary, 비저장소 → 오류(스텁 오류가 아니어야 함)를 kanban 테스트로 고정했다. CLI가 이 함수를 쓰는지는 `TestSlotCLI_RootNormalizesWorktreeProjectDir`(CLAUDE_PROJECT_DIR=워크트리 → 기록이 primary에 생김)가 고정한다.
- **N4 반영**: 정규화 불가(no-git) 행의 fail-open 감사 줄 위치를 "정규화 전 훅 루트 아래 `.moai/logs/slot-lease-audit.jsonl`"로 테스트에 고정했다.
- **N3**: AC-RSL-014의 토큰 파일 + `grep -f` 형태는 M5에서 판정 명령을 쓸 때 반영한다(M1 범위에 해당 판정 없음).

#### RED 증거 (원문 발췌, 전체 출력은 추적 경로에 반출)

반출 경로: `.moai/reports/t607/m1-red/m1-kanban-red.txt`, `m1-hook-red.txt`, `m1-config-red.txt`.

R-K (kanban): `go test ./internal/kanban/ -run '^(TestSlotLease|TestResolveSlotLeaseRoot)' -count=1 -v` → 종료 코드 1, FAIL 줄 32개.

```text
    slot_lease_cross_test.go:191: control: starts=2 (A: RESULT=started SESSION=lane-a | B: RESULT=started SESSION=lane-b)
    slot_lease_cross_test.go:215: lease: acquired=0 refused=0 busy=0 other=2 (A: RESULT=error SESSION=lane-a | B: RESULT=error SESSION=lane-b)
--- FAIL: TestSlotLease_ControlGroupTwoSessions (0.06s)
    --- PASS: TestSlotLease_ControlGroupTwoSessions/probe_then_start (0.03s)
    --- FAIL: TestSlotLease_ControlGroupTwoSessions/lease (0.03s)
    slot_lease_test.go:175: acquire on a free resource: slot lease: not implemented (M1 stub)
    slot_lease_test.go:258: acquire over a live foreign holder: err = slot lease: not implemented (M1 stub), want the held sentinel
    slot_lease_test.go:315: acquire under a contended mutation lock: err = slot lease: not implemented (M1 stub), want the busy sentinel
    slot_lease_test.go:333: acquire over a live holder with a free mutation lock: err = slot lease: not implemented (M1 stub), want the held sentinel
    slot_lease_test.go:757: ResolveSlotLeaseRoot(non-repo) answered with the M1 stub error, not a resolution failure: slot lease: not implemented (M1 stub)
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.696s
```

R-H (hook): `go test ./internal/hook/ -run '^TestSlotLeaseGuard_' -count=1 -v` → 종료 코드 1.

```text
    slot_lease_guard_test.go:232: decision = "", want deny
    slot_lease_guard_test.go:254: no audit line written, want last event "allow-self"
    slot_lease_guard_test.go:312: no "[moai:slot-lease] advisory:" line on the advisory stream; a silent allow makes the guard look enforcing while it is not. got: ""
    slot_lease_guard_test.go:405: enabled guard did not deny a live foreign holder through the handler (decision "allow") — the disabled rows above assert nothing until this passes
    slot_lease_guard_test.go:441: hook root in a linked worktree: decision = "", want deny — the guard read the worktree's empty state instead of the primary's record
FAIL	github.com/modu-ai/moai-adk/internal/hook	9.227s
```

(wt-deny의 실패 문구는 M4 이후 뮤턴트를 겨냥한 설명이다. M1에서 실제 원인은 가드 스텁이 모든 호출을 허용하는 것이다.)

R-C (config): `go test ./internal/config/ -run '^(TestDefaults_SlotLeaseDisabled|TestSlotLeaseConfig_)' -count=1 -v` → 종료 코드 1.

```text
    workflow_slot_lease_test.go:43: Workflow.SlotLease.DefaultMaxDuration = "", want a duration string equal to 30m
--- FAIL: TestDefaults_SlotLeaseDisabled (0.00s)
    --- PASS: TestSlotLeaseConfig_KeyShape/round_trip (0.00s)
    --- FAIL: TestSlotLeaseConfig_KeyShape/absent_key_keeps_defaults (0.00s)
2026/09/12 04:04:50 WARN failed to load workflow config, using defaults error="parse workflow.yaml: config: invalid YAML syntax"
    workflow_slot_lease_test.go:112: slot_lease.enabled = false after loading a malformed resource entry — the entry would vanish into the quiet disabled path instead of reaching the guard's fail-open report
FAIL	github.com/modu-ai/moai-adk/internal/config	0.443s
```

R-C의 마지막 두 줄이 plan.md §B4가 경고한 결함을 현재 로더에서 그대로 재현한다: 자원 항목 하나의 타입 오류가 workflow 섹션 전체를 기본값(= 꺼짐)으로 떨어뜨린다.

#### AC별 RED 상태

| AC | 테스트 | M1 상태 | 실패 이유(올바른 이유인가) |
|---|---|---|---|
| AC-RSL-001a | `TestSlotLease_ControlGroupTwoSessions` | FAIL(lease) / PASS(probe_then_start) | 표면 없음 → 두 자식 모두 `RESULT=error ... not implemented`. 대조 갈래 `starts=2`는 양성 대조로 통과해야 맞다 |
| AC-RSL-002 | `TestSlotLeaseBusy_IsNotHeld` | FAIL 3/3 | busy·held 센티널 대신 스텁 오류 |
| AC-RSL-003a/b | `TestSlotLease_RecordsAllFields` | FAIL 2/2 | 획득 미구현 |
| AC-RSL-004 | `TestSlotLease_RefusesLiveForeignHolderWritesNothing` | FAIL | held 센티널 없음 |
| AC-RSL-005 | `TestSlotLease_Liveness` | FAIL 3/3 | 획득 미구현 |
| AC-RSL-006 | `TestSlotLease_DeclaredBound` | FAIL 4/4 | 조회·획득·상한 파싱 미구현 |
| AC-RSL-007 | `TestSlotLease_ForceRecordsDisplaced` | FAIL | 획득 미구현 |
| AC-RSL-008 | `TestSlotLease_Release` | FAIL 3/3 | 해제 미구현 |
| AC-RSL-009 | `TestSlotLease_RejectsInvalidResourceNames` | FAIL | 이름 검증 미구현(양성 대조도 실패) |
| AC-RSL-010 | `TestSlotLeaseGuard_DisabledNeverReadsNorDenies`, `TestDefaults_SlotLeaseDisabled` | FAIL | 꺼짐 6조합은 PASS(스텁이 원래 허용·무음), **양성 대조 2개가 FAIL**이라 테스트 전체가 RED. 기본 상한 `""` |
| AC-RSL-011 | `TestSlotLeaseGuard_DenyMatrix` | FAIL 6/9 | base·multi 거부 안 됨, 허용 행 4개 감사 줄 없음. n-enabled·n-match·n-quoted는 "허용 + 감사 줄 없음"이 스텁과 같아 M1에서 PASS — 이 세 행의 판별력은 M4 뮤턴트 표에서 확보한다 |
| AC-RSL-012 | `TestSlotLeaseGuard_FailOpen`, `TestSlotLeaseConfig_MalformedResourceKeepsEnabled` | FAIL 5/5 + FAIL | 안내·fail-open 줄 없음, 로더가 섹션 전체를 버림 |
| AC-RSL-013(a) | `TestSlotLease_SeparateFromIntegrationWindow` | FAIL 1/2 | 슬롯 쪽은 획득 미구현. 통합 쪽 하위 테스트는 기존 동작이라 PASS(회귀 가드) |
| AC-RSL-016 + N1 | `TestSlotLeaseGuard_NormalizesWorktreeRootToPrimary`, `TestResolveSlotLeaseRoot_NormalizesToPrimary` | FAIL 3/3, FAIL 4/4 | 가드·해석 함수 미구현 |
| (형태 가드) | `TestSlotLeaseConfig_KeyShape/round_trip` | PASS | M1 스텁 구조체 자체가 이 결정이므로 초록이 맞다 |

#### 미검증 — cli 슬롯 대기 (written, RED-unverified)

`internal/cli/slot_test.go` 7개 테스트는 작성만 했고 **컴파일도 실행도 하지 않았다**(리드 제약). 확인한 것은 `gofmt -l` 무출력(구문 파싱 통과)뿐이다. cli 슬롯에서 돌릴 명령:

```bash
go test ./internal/cli/ -run '^(TestSlotCLI_AcquireJSONFields|TestSlotCLI_OmittedNameAndCommand|TestSlotCLI_StatusJSONFree|TestSlotCLI_HelpListsVerbs|TestSlotCLI_ForceReportsDisplaced|TestSlotCLI_RefusesWithoutSessionID|TestSlotCLI_RootNormalizesWorktreeProjectDir)$' -count=1 -v -timeout 600s
```

예상 RED: `slot` 명령이 없으므로 `findSlotCmd`의 `no \`moai slot\` command registered on the root command`(assertion-RED). 이 파일은 생성자 심볼 대신 `rootCmd`에서 이름으로 명령을 찾으므로 M3 전에도 패키지 빌드를 깨지 않도록 설계했지만, 그 주장 자체가 아직 관측되지 않은 가설이다.

#### 회귀·빌드·린트

- R1 `go test ./internal/config/ -skip '^(TestDefaults_SlotLeaseDisabled|TestSlotLeaseConfig_)' -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/config	4.410s`
- R2 `go test ./internal/kanban/ -skip '^(TestSlotLease|TestResolveSlotLeaseRoot)' -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/kanban	155.451s`
- R3 `go test ./internal/hook/ -run 'IntegrationLock|BranchGuard' -count=1 -v` → `ok  	github.com/modu-ai/moai-adk/internal/hook	20.480s`, `--- PASS` 159줄
- `go build ./internal/kanban/ ./internal/hook/ ./internal/config/` → 종료 코드 0; `GOOS=windows GOARCH=amd64` 같은 명령 → 0; `go vet ./internal/hook/ ./internal/config/` → 0; `GOOS=windows GOARCH=amd64 go vet ./internal/kanban/ ./internal/hook/ ./internal/config/` → 0
- `golangci-lint run --timeout=5m ./internal/kanban/... ./internal/hook/... ./internal/config/...` → `0 issues.`(첫 실행의 errcheck 3건은 새 테스트 헬퍼의 것이었고 수정했다)

#### M2·M4로 넘기는 사항

- M4: `internal/config/cache.go`의 `configCacheSchemaVersion`을 올려야 한다. 그 파일 주석이 필드 추가 때 올리라고 요구한다(옛 캐시가 새 필드를 영값으로 덮는다). M1은 필드만 선언했고 캐시 버전은 건드리지 않았다.
- M4: `SlotLeaseResourceConfig`의 항목별 관대한 디코딩(잘못된 항목은 `Invalid`만 채우고 `enabled`는 유지).
- M2: AC-RSL-001b 뮤턴트 관측, AC-RSL-002·005·006 뮤턴트 짝. M4: AC-RSL-010·011·016 뮤턴트 표.
- 계획 대비 편차: 없음. 추가한 것 — AC-RSL-013(a)와 `TestSlotLeaseConfig_KeyShape`, `TestSlotCLI_RefusesWithoutSessionID`를 M1에 함께 썼다(모두 plan.md M2-M4 범위의 판정을 앞당겨 고정한 것).

### M2 — 임대 핵심 (internal/kanban)

- 커밋: `28d58376b`(구현), `c331bc589`(기록 쓰기 함수가 자기 디렉터리를 만들도록 수정 — 아래 뮤턴트 참조). 측정 트리: HEAD `c331bc589`, tree `2c7fa687ec85420372f293cbcd1f92b3d362d301`.
- 구현: `internal/kanban/slot_lease.go`, `slot_lease_mutation_unix.go`(Unix 잔재 정리 없음), `slot_lease_mutation_windows.go`(`clearStaleLockAtPath` 재사용). 판정표: 보유자 없음·자기 세션 → 획득(재획득은 상한 재시작) / 소유자 사라짐 → `stale` 인수 / 상한 경과 → `expired` 인수 / 강제 → `force` 인수 / 그 외 → held 오류 + `refuse` 감사, 기록 불변.
- M1 테스트 수정 한 곳: `TestResolveSlotLeaseRoot_NormalizesToPrimary/not_a_repository`의 "스텁 오류가 아닐 것" 조항은 스텁 심볼이 사라져 "오류가 해석하지 못한 디렉터리를 이름으로 댈 것"으로 바꿨다. 새 가장자리 테스트 `TestSlotLease_UnreadableRecordAndInputEdges`(손상 기록은 비어 있음이 아님, 강제로만 정리, 강제 해제 감사 사유, 입력 가장자리)를 추가했다 — 구현 뒤에 쓴 커버리지 보강이며 RED 단계를 거치지 않았다.

GREEN: `go test ./internal/kanban/ -run '^(TestSlotLease|TestResolveSlotLeaseRoot)' -count=1 -v` → 종료 코드 0, `--- PASS` 40줄, `--- FAIL` 0줄(전체 출력 `.moai/reports/t607/m2/m2-kanban-green.txt`).

```text
    slot_lease_cross_test.go:191: control: starts=2 (A: RESULT=started SESSION=lane-a | B: RESULT=started SESSION=lane-b)
    slot_lease_cross_test.go:215: lease: acquired=1 refused=1 busy=0 other=0 (A: RESULT=acquired SESSION=lane-a | B: RESULT=held SESSION=lane-b)
ok  	github.com/modu-ai/moai-adk/internal/kanban	4.686s
```

뮤턴트(각각 파일 수정 → 해당 테스트 실행 → 되돌림, 되돌린 뒤 `git diff --exit-code --stat -- internal/kanban/slot_lease.go` → 출력 없음, 종료 코드 0 = 커밋본과 동일):

| 뮤턴트 | 변형 | 실행 | 관측(원문) |
|---|---|---|---|
| AC-RSL-001b | `withSlotLeaseMutation(projectRoot, req.Resource, decide)` → `decide()` 한 줄 | `-run '^TestSlotLease_ControlGroupTwoSessions$'` 종료 코드 1 | `lease: acquired=2 refused=0 busy=0 other=0 (A: RESULT=acquired SESSION=lane-a \| B: RESULT=acquired SESSION=lane-b)` / `--- FAIL: TestSlotLease_ControlGroupTwoSessions/lease`. 대조 갈래는 같은 실행에서 `starts=2`, PASS |
| AC-RSL-001b 복구 | 되돌림 | 같은 명령 종료 코드 0 | `m2-mutant-001b-restored.txt` |
| AC-RSL-002 | busy 경로가 `ErrSlotLeaseHeld`를 감쌈 | `-run '^TestSlotLeaseBusy_IsNotHeld$'` 종료 코드 1 | `acquire under a contended mutation lock: err = slot lease: resource held by another session (waited 1.65s): kanban board lock held, want the busy sentinel` |
| AC-RSL-005 | `Stale()`이 pid를 보지 않음(항상 살아 있음) | `-run '^TestSlotLease_Liveness$'` 종료 코드 1 | `--- FAIL: TestSlotLease_Liveness/stale_takeover` — `acquire over a dead owner without --force: slot lease: resource held by another session ...` |
| AC-RSL-006 | 판정표의 `Expired` 가지 삭제 | `-run '^TestSlotLease_DeclaredBound$'` 종료 코드 1 | `--- FAIL: TestSlotLease_DeclaredBound/expired_takeover` — `acquire over an expired live owner without --force: ... held by another session` |

001b 첫 시도에서 알게 된 것: 원래 디렉터리 생성이 `withSlotLeaseMutation` 안에만 있어서, 한 줄 되돌림이 임계 구역과 함께 mkdir까지 없앴다(쓰기가 실패해 `other=2`가 됐을 것이다). 그래서 뮤턴트를 돌리기 전에 `writeSlotLease`가 자기 디렉터리를 만들도록 고치고(`c331bc589`) 테스트를 다시 통과시킨 뒤 뮤턴트를 돌렸다. 되돌림 한 줄이 임계 구역만 끄도록 한 것이다.

회귀·빌드·린트(M2):
- `go test ./internal/kanban/ -skip '^(TestSlotLease|TestResolveSlotLeaseRoot)' -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/kanban	153.868s`(측정 트리 `28d58376b` 작업 트리; 이후 변경은 `writeSlotLease`의 mkdir 한 곳)
- `GOOS=windows GOARCH=amd64 go build ./internal/kanban/` → 0, 같은 조건 `go vet` → 0, 네이티브 `go vet ./internal/kanban/` → 0
- `golangci-lint run --timeout=5m ./internal/kanban/...` → `0 issues.`
- 커버리지: `go test ./internal/kanban/ -run '^(TestSlotLease|TestResolveSlotLeaseRoot)' -count=1 -coverprofile=<scratch>/m2-cover.out` 후 프로파일에서 `slot_lease.go` 줄만 합산 → `slot_lease.go statements=189 covered=163 pct=86.2%`(`28d58376b` 작업 트리 기준, mkdir 한 줄 추가 전).

### M4 — 설정 키와 PreToolUse 가드 (internal/config, internal/hook)

- 커밋: `dff5dee4a`(구현, tree `f8dfe502339c9f610757e7cdabb7b744945a53c8`). 뒤이은 증거 커밋에 설정 가장자리 테스트 `TestSlotLeaseConfig_NonMappingEntriesAreMarked`(구현 뒤 커버리지 보강, RED를 거치지 않음)와 뮤턴트 실행기가 들어간다.
- 변경: `defaults.go`에 `DefaultSlotLeaseMaxDuration = "30m"`(정의는 이 한 곳)과 `SlotLease{Enabled: false, DefaultMaxDuration}`. `slot_lease_config.go`에 자원 항목별 관대한 `UnmarshalYAML`(맵이 아닌 항목·목록이 아닌 `commands`·문자열이 아닌 원소는 `Invalid`만 채우고 섹션은 살린다). `cache.go`의 `configCacheSchemaVersion` 3 → 4. `pre_tool.go`에서 통합 가드 바로 뒤에 `checkSlotLease` 배선 — `slotLeaseConfig()`가 꺼짐(nil 제공자·nil 설정 포함)을 돌려주면 `h.projectRoot()`조차 부르지 않는다. 가드는 매칭된 호출에서만 `kanban.ResolveSlotLeaseRoot`로 루트를 정규화한다(CLI와 같은 함수, N1).

GREEN:
- `go test ./internal/hook/ -run '^TestSlotLeaseGuard_' -count=1 -v` → 종료 코드 0, `ok  	github.com/modu-ai/moai-adk/internal/hook	9.434s`. DenyMatrix 9행, FailOpen 5행, DisabledNeverReadsNorDenies 6조합 + 양성 대조 2개, NormalizesWorktreeRootToPrimary 3행 모두 `--- PASS`(`.moai/reports/t607/m4/m4-hook-green.txt`).
- `go test ./internal/config/ -run '^(TestDefaults_SlotLeaseDisabled|TestSlotLeaseConfig_)' -count=1 -v` → 종료 코드 0, 테스트 4개 `--- PASS`(`.moai/reports/t607/m4/m4-config-green.txt`).

뮤턴트 표(`.moai/reports/t607/m4/m4_mutants.py` — 실행기가 파일을 스크래치 백업으로 `cp` → 치환 한 곳(정확히 한 번 일치 단언) → 해당 테스트 → 백업에서 `cp`로 복구 → `filecmp` 바이트 비교. 실행 뒤 `git diff --quiet -- internal/hook/slot_lease_guard.go` → 종료 코드 0, `cmp <backup> internal/hook/slot_lease_guard.go` → 종료 코드 0):

| 뮤턴트 | 변형 | 실패 행(원문) |
|---|---|---|
| 설정 검사 삭제 | `!cfg.Enabled ||` 제거 | `--- FAIL: TestSlotLeaseGuard_DenyMatrix/n-enabled` |
| 패턴 검사 삭제 | `if true \|\| re.MatchString(...)` | `n-match`, `n-quoted` |
| 따옴표 제거 삭제 | `scrubbed := command` | `n-quoted` |
| 세션 비교 삭제 | `allow-self` 가지 제거 | `n-self` |
| 생존 검사 삭제 | `allow-stale` 가지 제거 | `n-alive` |
| 만료 검사 무효화 | `case false && lease.Expired(now):` | `n-bound` |
| 보유자 있음 검사 삭제 | `allow-unheld` 가지 제거 | `n-held`, `multi` |
| 첫 자원만 판정 | `range matched[:1]` | `multi` |
| AC-RSL-016 정규화 제거 | `root, err := hookRoot, error(nil)` | `NormalizesWorktreeRootToPrimary/wt-deny`, `/no-git` |
| AC-RSL-016 실패 시 비정규화 루트로 진행 | 오류 가지가 `root = hookRoot` | `NormalizesWorktreeRootToPrimary/no-git` |

모든 뮤턴트는 종료 코드 1이었고 복구 비교는 모두 `True`였다. 만료 뮤턴트의 첫 형태(가지 삭제)는 `now`가 쓰이지 않아 **컴파일이 실패했다** — `slot_lease_guard.go:127:2: declared and not used: now`, FAIL 행 0. 이 실행은 아무것도 재지 않았으므로 판정에 쓰지 않고, 가지를 `false &&`로 무효화하는 형태로 다시 돌려 `n-bound` 실패를 관측했다(`summary.tsv`에 두 줄이 모두 남아 있다).

회귀·빌드·린트(M4, 트리 `dff5dee4a`):
- `go test ./internal/config/ -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/config	3.204s`
- `go test ./internal/hook/ -run 'IntegrationLock|BranchGuard|PreTool|SlotLease' -count=1 -v` → `ok  	github.com/modu-ai/moai-adk/internal/hook	24.830s`, `--- PASS` 306줄, `--- FAIL` 0줄
- `golangci-lint run --timeout=5m ./internal/kanban/... ./internal/hook/... ./internal/config/...` → `0 issues.`
- `GOOS=windows GOARCH=amd64 go build` / `go vet`(kanban·hook·config) → 0 / 0, 네이티브 `go vet` → 0
- 커버리지(해당 테스트만, 프로파일에서 파일 줄 합산): `slot_lease_guard.go statements=72 covered=68 pct=94.4%`, `slot_lease_config.go statements=25 covered=25 pct=100.0%`

### M3 — CLI `moai slot` (internal/cli) — cli 슬롯 안에서 검증

리드가 heavy-test 슬롯을 lane-10에 넘겼다(코디네이터 전달). 첫 cli 컴파일 전에 `ps -eo pid,command | grep -E 'go (test|build|vet)'`로 확인했을 때 다른 go 프로세스는 없었다(종료 코드 1). 그 뒤 슬롯 사용 중 한 번 다른 레인의 `go test ./internal/cli -run ^TestDefaultLaunchSelectsNativeGateway ...` 프로세스(pid 50142, 명령줄이 `/tmp/t649-default-red.txt`로 출력 — t649 레인)가 관측됐고, `lsof`로 cwd를 보려 했을 때는 이미 끝나 있었다(2026-09-11T19:40:18Z UTC). 우리 쪽 go 명령은 끝까지 한 번에 하나씩만 돌렸다.

1. **M1 CLI RED 먼저**: `go test ./internal/cli/ -run '^TestSlotCLI_' -count=1 -v -timeout 600s` → 종료 코드 1. 컴파일은 됐고(M1의 "컴파일 미검증" 가설 확인), 7개 모두 `slot_test.go:NNN: no \`moai slot\` command registered on the root command`로 실패 — assertion-RED(`.moai/reports/t607/m3/m3-cli-red.txt`).
2. **config 한 키 읽기 함수**: `TestLoadSlotLeaseDefaultMaxDuration`을 먼저 쓰고 실행 → `internal/config/workflow_slot_lease_test.go:162:14: undefined: LoadSlotLeaseDefaultMaxDuration`(compile-RED). 그다음 `internal/config/loader_slot_lease.go`(`LoadGitFlowDevelopBranch` 방식, 실패 시 `DefaultSlotLeaseMaxDuration`) → 하위 5개 PASS.
3. **구현**: `internal/cli/slot.go` — `moai slot acquire|status|release`. 루트는 `CLAUDE_PROJECT_DIR`(없으면 cwd)를 `kanban.ResolveSlotLeaseRoot`로 정규화(N1 — 가드와 같은 함수), git 저장소 밖이면 시작 디렉터리 그대로. 세션 id는 통합 창의 `integrationSessionID` 재사용, 소유자 pid는 `session.ResolveOwnerPID()`(풀리지 않으면 0). 종료 코드: held 3, busy 4(재시도 안내 포함), 그 외 1. 두 오류 모두 kanban 센티널이 `errors.Is`로 닿는다. 상태 조회는 빈 이름·명령을 `(not given)`으로 표시하고, `--resource` 없이 부르면 기록된 자원을 모두 나열한다.
4. **cli 가드에 걸린 회귀와 수리**: 전체 cli 패키지 실행에서 `TestSessionPIDStamp_NotSetFromHooks`가 실패했다 — `hook sources must not set MOAI_SESSION_PID (a hook PID is dead on arrival): [.../internal/hook/slot_lease_guard_test.go]`. 원인은 M1에서 쓴 hook 테스트의 환경 비움 한 줄(`t.Setenv` 세션 pid 변수)이다. 이 가드는 `internal/hook` 아래 모든 `.go`(테스트 포함)에서 그 변수 이름을 금지한다. 가드는 소유자 pid를 해석하지 않으므로 그 줄을 지우고 이유를 주석으로 남겼다. 수리 뒤 `TestSessionPIDStamp_NotSetFromHooks` PASS.

GREEN(트리 = M3 커밋):
- `go test ./internal/cli/ -run '^TestSlotCLI_' -count=1 -v -timeout 600s` → 종료 코드 0, 11개 `--- PASS`(M1 7개 + 표면 4개 `TestSlotCLI_HeldAndBusyCarryDistinctExitCodes`·`StatusStatesAndList`·`ReleaseRoundTripAndRefusals`·`InputRefusals` — 표면 4개는 구현 뒤 커버리지 보강으로 썼다), `ok  	github.com/modu-ai/moai-adk/internal/cli	1.931s`(`.moai/reports/t607/m3/m3-cli-green.txt`).
- `go test ./internal/cli/ -run '^(TestSessionPIDStamp_|TestSlotCLI_|TestRootCmd_|TestIntegration)' -count=1 -v -timeout 600s` → 종료 코드 0, `--- PASS` 37, `--- FAIL` 0(`m3-cli-targeted.txt`).
- `go test ./internal/config/ -run '^(TestLoadSlotLeaseDefaultMaxDuration|TestDefaults_SlotLeaseDisabled|TestSlotLeaseConfig_)' -count=1 -v` → 종료 코드 0.
- `go test ./internal/hook/ -run '^TestSlotLeaseGuard_' -count=1` → `ok ... 9.288s`(환경 비움 한 줄 제거 뒤 재실행).
- `go build ./internal/cli/ ./cmd/moai/` → 0, `go build ./...` → 0, `GOOS=windows GOARCH=amd64 go build ./...` → 0, `go vet ./internal/cli/ ./internal/hook/` → 0, `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` → 0.
- `golangci-lint run --timeout=10m ./internal/cli/ ./internal/hook/...` → `0 issues.`(그 앞 실행 `./internal/cli/ ./internal/config/... ./internal/hook/... ./internal/kanban/...` → `0 issues.`)
- 커버리지: `slot.go statements=140 covered=134 pct=95.7%`(TestSlotCLI_ 11개 기준).

미완(Gap):
- **전체 cli 패키지는 끝까지 돌지 못했다.** `go test ./internal/cli/ -count=1 -timeout 600s` → `panic: test timed out after 10m0s`, 그때 돌던 테스트 `TestSyncGitSpecStatuses_NoSpecIDsInGitLog (0s)`, `FAIL ... 600.769s`. 시한 전에 실패한 테스트는 위 4번 하나뿐이었고 이미 고쳤다(`m3-cli-full-suite-excerpt.txt`). 10분 안에 끝나지 않는 것이 이 머신의 패키지 크기 문제인지, 그 테스트가 멈춘 것인지는 가리지 못했다. 전체 판정은 develop push 뒤 CI 몫이다.

## §E.3 Run-phase Audit-Ready Signal

_<run 단계 대기>_

## §E.4 Sync-phase Audit-Ready Signal

_<sync 단계 대기>_
