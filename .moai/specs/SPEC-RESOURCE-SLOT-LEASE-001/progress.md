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

## §E.3 Run-phase Audit-Ready Signal

_<run 단계 대기>_

## §E.4 Sync-phase Audit-Ready Signal

_<sync 단계 대기>_
