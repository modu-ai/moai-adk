# sync-audit — SPEC-RESOURCE-SLOT-LEASE-001 (card t607)

- **판정: FAIL**
- **점수: 79.5 / 100** (기본 평가 프로필 `.moai/config/evaluator-profiles/default.md`, flat 가중 방식)
- **막는 결함: 1건** (F1). must-pass 방화벽(Functionality·Security)은 둘 다 통과했으나, 이 카드가 자기가 고친 패키지에 레드를 하나 들여왔고 그 레드가 §E.3·CHANGELOG의 무결함 주장을 거짓으로 만든다.
- 감사 트리: `WT-heavy-test-slot`, HEAD `1709886949827d322693c83981a7d303387e6f4c`, 카드 base `8d42587e695aee97cb4454e5efa564cd613343d3`(= `git merge-base develop HEAD`).
- 판정 도구 provenance: `./bin/moai`(이 트리에서 빌드, Commit `9c288daa9`, HEAD보다 1커밋 뒤 — 그 1커밋은 `progress.md`만 바꾼다). Go 툴체인은 이 세션의 `go`/`golangci-lint`.
- 적용 규칙: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1(surface 2) · §2 · §2.3, `.claude/rules/moai/development/verification-completeness.md` §1.1, `.claude/rules/moai/development/spec-frontmatter-schema.md` § SHA placeholder backfill exemption.

## 1. 차원 점수

| 차원 | 점수 | 판정 | 증거 (이번 실행, 이 트리) |
|---|---|---|---|
| Functionality (40%) | 75/100 | PASS (must-pass 충족) | `go test ./internal/kanban/ -run 'TestSlotLease\|TestResolveSlotLeaseRoot' -count=1 -v` → 종료 코드 0, `--- PASS` 40 / `--- FAIL` 0, `ok … 4.984s`. 로그 원문: `control: starts=2 (A: RESULT=started SESSION=lane-a \| B: RESULT=started SESSION=lane-b)` / `lease: acquired=1 refused=1 busy=0 other=0 (A: RESULT=acquired SESSION=lane-a \| B: RESULT=held SESSION=lane-b)`. `go test ./internal/hook/ -run 'TestSlotLeaseGuard_' -count=1 -v` → 0, `--- PASS` 29 / FAIL 0, `ok … 9.560s`. |
| Security (25%) | 95/100 | PASS (Critical/High 0) | 자원 이름 신뢰 경계 단일화: `internal/kanban/slot_lease.go:63` `^[a-z0-9-]{1,64}$` + `:217` `ValidateSlotResourceName`, `ReadSlotLease`/`Acquire`/`Release`/CLI 전부 경로 조립 전에 통과. 가드 fail-open 경로 전수 초록(위 hook 29 PASS). 셸 경유 없음(`exec.Command("git", …)` 직접 호출, `slot_lease.go:251`). |
| Craft (20%) | 70/100 | FAIL | 커버리지는 문턱 통과 — 내가 잰 값: `slot_lease.go 85.9% (164/191)`, `slot_lease_guard.go 94.4% (68/72)` (`go test … -coverprofile` + `go tool cover -func` 재집계). 린트 `golangci-lint run --timeout=5m ./internal/kanban/... ./internal/hook/... ./internal/config/...` → 종료 코드 0, `0 issues.` **그러나** `go test ./internal/config/ -count=1` → 종료 코드 1 (F1). |
| Consistency (15%) | 78/100 | PASS | 템플릿 중립성·유출·미러 가드 전수 초록: `go test ./internal/template/ -run 'TestTemplateNeutralityAudit$\|TestTemplateNoInternalContentLeak$\|TestRuleProvenance\|TestRuleTemplateMirror' -count=1 -v` → 0, `--- PASS` 20 / FAIL 0. 미러 바이트 동일 `cmp` → 0. 통합 창 분리 프로브(아래 §3) 무출력. 감점 사유는 F3·F4. |

가중 합: `0.40×75 + 0.25×95 + 0.20×70 + 0.15×78 = 79.45`.

**must-pass 방화벽**: Functionality(“모든 AC 충족”)와 Security(“Critical/High 0”)는 각각 독립으로 통과했다. 따라서 FAIL은 방화벽이 아니라 **막는 결함 F1**이 만든다.

## 2. AC별 — 내가 잰 것 / 증거로 읽은 것 / 지금 잴 수 없는 것

`internal/cli`를 컴파일·링크하는 명령은 하나도 돌리지 않았다. 돌린 네 패키지는 먼저 `go list -test -deps <pkg> | grep -c 'moai-adk/internal/cli$'` → kanban 0, hook 0, config 0, template 0 을 확인했다.

| AC | 분류 | 근거 |
|---|---|---|
| AC-RSL-001a | **내가 재측정** | 위 kanban 40 PASS + 대조군 두 갈래 로그 원문 |
| AC-RSL-001b | **내가 뮤턴트 재현** | §4 M-1 |
| AC-RSL-002·003a·003b·004·005·006·007·008·009 | **내가 재측정** | 같은 kanban 실행(하위 테스트 전부 PASS) |
| AC-RSL-003c (CLI·교차 플랫폼) | **지금 잴 수 없음** | heavy-test 슬롯 제약. `.moai/reports/t607/m3/*.txt` 와 §E.2 M3 절을 읽었다 |
| AC-RSL-010·011·012·016 | **내가 재측정** | 위 hook 29 PASS |
| AC-RSL-011 뮤턴트 표 10행 | **증거로 읽음 + 1행 재현** | `.moai/reports/t607/m4/mutants/summary.tsv` 전 행 `restored_cmp_equal=True`; AC-RSL-016 행은 §4 M-2로 직접 재현 |
| AC-RSL-013(a)(b) | **내가 재측정** | kanban 분리 테스트 PASS + `go test ./internal/kanban/ -count=1` → 0 (`ok … 157.573s`), `go test ./internal/hook/ -count=1` → 0 (`ok … 152.466s`) — 스윕 수 0 아님 |
| AC-RSL-013(c) | **내가 재측정** | §3 |
| AC-RSL-014 (a)-(h) | **내가 재측정** | §3 |
| AC-RSL-014 (i) `make build` | **지금 잴 수 없음** | `internal/cli` 링크. §E.4 가 이미 “manager-docs 가 직접 관측한 값이 아니다”로 명시 — 그 공백 표기는 정확하다 |
| AC-RSL-015 | **내가 재측정** | §3 (게이트 rc=0, 세 파일 각 1 적중) |

## 3. 내가 다시 잰 판정식 (원문)

```text
$ git merge-base --is-ancestor WT-acquire-branch-record develop ; echo $?
0
$ git rev-parse WT-acquire-branch-record   → 3262fa9be4522861cec9ba6bd74fb3bd2387c311
$ git rev-parse develop                    → 8d42587e695aee97cb4454e5efa564cd613343d3
  ⇒ M6 게이트는 열림. 레인 문서 3종 편집은 REQ-RSL-016 을 위반하지 않는다.

$ git diff --name-only 8d42587e6..HEAD | wc -l        → 66   (대조군 1 이상)
$ git diff --name-only 8d42587e6..HEAD -- internal/cli/integration.go \
      internal/hook/integration_lock_guard.go internal/kanban/integration_lock.go
  (출력 없음)                                          ⇒ AC-RSL-013(c) 프로브 통과

$ grep -c 'moai slot' .claude/rules/moai/workflow/kanban-dispatch.md                                   → 1
$ grep -c 'moai slot' internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md        → 1
$ grep -c 'moai slot' .claude/rules/local/gitflow-lane-protocol.md                                      → 1
$ cmp .claude/rules/moai/workflow/resource-slot-lease.md \
      internal/template/templates/.claude/rules/moai/workflow/resource-slot-lease.md ; echo $?   → 0

$ /usr/bin/grep -nwiE -f .moai/reports/t607/m5/tool-tokens.txt \
      internal/template/templates/.moai/config/sections/workflow.yaml ; echo $?    → (무출력) 1
$ /usr/bin/grep -nwiE -f .moai/reports/t607/m5/tool-tokens.txt \
      internal/template/templates/.claude/rules/moai/workflow/resource-slot-lease.md ; echo $?  → (무출력) 1
$ /usr/bin/grep -cwiE -f .moai/reports/t607/m5/tool-tokens.txt \
      internal/template/templates/.claude/rules/moai/languages/python.md ; echo $?  → 7 / 0
  ⇒ 양성 대조군이 적중하므로 위 두 0 은 공허한 0 이 아니다. 토큰 목록은 46줄 = AC-RSL-014 열거와 개수 일치.

$ /usr/bin/grep -rnE 'SPEC-[A-Z]|\bt[0-9]{3}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}' \
      internal/template/templates/.claude/rules/moai/workflow/resource-slot-lease.md ; echo $?  → (무출력) 1

$ ./bin/moai spec lint --strict .moai/specs/SPEC-RESOURCE-SLOT-LEASE-001
✓ No findings — all SPEC documents are valid          (종료 코드 0)
```

## 4. 뮤턴트 재현 (내가 심고, 돌리고, 되돌렸다)

**M-1 — 임계 구역 우회 (AC-RSL-001b)**. `internal/kanban/slot_lease.go`에서 `withSlotLeaseMutation(projectRoot, req.Resource, decide)` → `decide()` 한 줄 치환.

```text
$ go test ./internal/kanban/ -run 'TestSlotLease_ControlGroupTwoSessions' -count=1 -v   → 종료 코드 1
control: starts=2 (A: RESULT=started SESSION=lane-a | B: RESULT=started SESSION=lane-b)
lease: acquired=2 refused=0 busy=0 other=0 (A: RESULT=acquired SESSION=lane-a | B: RESULT=acquired SESSION=lane-b)
lease: acquired=2 refused=0, want 1 and 1 — the record's read-modify-write is not serialized across processes
--- FAIL: TestSlotLease_ControlGroupTwoSessions
    --- PASS: TestSlotLease_ControlGroupTwoSessions/probe_then_start
    --- FAIL: TestSlotLease_ControlGroupTwoSessions/lease
복구: cp 백업 → cmp 0, git diff --quiet -- internal/kanban/slot_lease.go → 0
```

기록된 `acquired=2` 와 같다. 양성 대조 갈래(`probe_then_start`)는 뮤턴트 아래에서도 PASS — 하네스가 이중 시작을 여전히 관측할 수 있다는 뜻이므로 표면 갈래의 FAIL 이 해석 가능하다.

**M-2 — 가드 루트 정규화 제거 (AC-RSL-016)**. `internal/hook/slot_lease_guard.go:108` 의 `root, err := kanban.ResolveSlotLeaseRoot(hookRoot)` → `root, err := hookRoot, error(nil)`.

```text
$ go test ./internal/hook/ -run 'TestSlotLeaseGuard_NormalizesWorktreeRootToPrimary' -count=1 -v  → 종료 코드 1
slot_lease_guard_test.go:445: hook root in a linked worktree: decision = "", want deny — the guard read the worktree's empty state instead of the primary's record
slot_lease_guard_test.go:468: no advisory for an unresolvable root; got ""
slot_lease_guard_test.go:474: last audit under the unnormalized hook root = ("allow-unheld", "", present=true), want a fail-open line
--- FAIL: .../wt-deny      --- PASS: .../primary-deny      --- FAIL: .../no-git
복구: cp 백업 → cmp 0, git diff --quiet -- internal/hook/slot_lease_guard.go → 0
```

`primary-deny` 가 뮤턴트 아래에서도 PASS 인 것이 핵심이다 — 두 행의 대비가 정규화 그 자체를 판별한다. 기록과 일치한다.

**대조군 하네스 점검 (경합이 타이밍 운이 아닌지)**. `internal/kanban/slot_lease_cross_test.go` 를 읽고 확인한 것: 끼어들기는 `HELPER_STALL_MARKER` 파일과 `HELPER_PROCEED_FLAG` 파일로 **구성**되며 기다려서 얻지 않는다(:127-143, :145-159). A 가 한도 안에 STALLED 를 못 남기면 `t.Fatalf` 로 **하네스 결함**을 선언한다(:140). 소유자 pid 는 `ownerPID := strconv.Itoa(os.Getpid())`(:180) — 테스트 내내 살아 있는 부모 pid 이고, 두 자식 모두 그 값을 기록한다(:199, :290-303). AC-RSL-001a 가 요구하는 그대로다. `busy>0` 과 `other>0` 은 판정이 아니라 하네스 설정 결함으로 실패시킨다(:216-221). 풀림 타임아웃 500ms 가 예산의 1/3 이하인지도 테스트가 스스로 단언한다(:177-179).

## 5. 결함

### F1 [High] [blocking] — 이 카드가 들여온 템플릿 설정 키 2개가 인벤토리에 없어 `internal/config` 가 레드다

- 위치: `internal/config/testdata/shipped_key_inventory.yaml` (누락), 단언 지점 `internal/config/shipped_key_reader_test.go:109`.
- 관측:

```text
$ go test ./internal/config/ -run 'TestShippedConfigKeysHaveReaders' -count=1   → 종료 코드 1
--- FAIL: TestShippedConfigKeysHaveReaders (0.82s)
    shipped_key_reader_test.go:109: REQ-CKH-008 anti-rot: 2 shipped config key(s) are NOT in the triage inventory
      (add to internal/config/testdata/shipped_key_inventory.yaml with W/P/R/D class):
          workflow.slot_lease.default_max_duration
          workflow.slot_lease.enabled
$ go test ./internal/config/ -count=1   → 종료 코드 1 (FAIL github.com/modu-ai/moai-adk/internal/config 2.064s)
$ grep -n 'slot' internal/config/testdata/shipped_key_inventory.yaml → 적중 0
```

- 귀속: 실패 단언은 **정확히 이 카드가 추가한 두 키**만 이름 댄다(`internal/template/templates/.moai/config/sections/workflow.yaml:127-130`). 같은 파일 `:126` 의 두 번째 진단 줄(`640 triaged … dead/unresolved/unbound`)은 `t.Logf` 라서 실패를 만들지 않는다(`shipped_key_reader_test.go:123-127`). 흡수한 develop 에서 온 기존 레드가 아니다.
- 왜 놓쳤나: run 단계의 `go test ./internal/config/ -count=1` → 0 은 **M4 트리 `dff5dee4a`** 에서 잰 값이고(§E.3 표), 템플릿 키는 그 뒤 M5 `685fc3387` 에서 들어왔다. sync 단계는 `-run '^(TestLoadSlotLeaseDefaultMaxDuration|TestDefaults_SlotLeaseDisabled|TestSlotLeaseConfig_)'` 로 좁혀 돌려 이 테스트를 선택하지 않았다(§E.4 `ac_remeasured_here` 3번째 항목). SPEC §G 가 “템플릿 규칙 인벤토리 테스트에 걸릴 수 있다”고 이미 경고한 바로 그 경로다.
- **요구되는 수리**: `internal/config/testdata/shipped_key_inventory.yaml` 에 두 키를 W/P/R/D 등급과 함께 추가한다. 둘 다 실제 판독기가 있으므로 **R(read)** 가 옳다 — `workflow.slot_lease.enabled` 는 `internal/hook/pre_tool.go:856-864` `slotLeaseConfig()` 가, `workflow.slot_lease.default_max_duration` 은 `internal/config/loader_slot_lease.go:19-30` `LoadSlotLeaseDefaultMaxDuration` 가 읽는다. 추가 뒤 `go test ./internal/config/ -count=1` 로 종료 코드 0 을 관측하고 그 원문을 progress.md 에 남긴다.

### F2 [Medium] [blocking] — F1 때문에 §E.3 과 CHANGELOG 의 무결함 주장이 자기 baseline 에서 거짓이다

- 위치: `.moai/specs/SPEC-RESOURCE-SLOT-LEASE-001/progress.md:316` `new_warnings_or_lints_introduced: 0`, `:309-311` `ac_pass_count: 16 / ac_fail_count: 0`; `CHANGELOG.md:27` “16 acceptance criteria (AC-RSL-001..016), all PASS”.
- §E.3 은 `:320-322` 에서 스스로 측정 트리를 **HEAD `272521e45`** 로 명시한다. 그 트리에서 이미 `internal/config` 는 레드였다(템플릿 키는 `685fc3387` 에서 들어왔고 `272521e45` 는 그보다 뒤다). 즉 “측정 당시에는 맞았다”는 변호가 성립하지 않는다 — 주장 자체가 자기 baseline 을 명시하고 그 baseline 에서 거짓이다(`verification-claim-integrity.md` §1.1 surface 2 + §2).
- **요구되는 수리**: F1 수리 뒤 `go test ./internal/config/ -count=1` 을 다시 재고 그 종료 코드·원문을 §E.3 옆에 붙인다. F1 을 이 카드에서 고치지 않기로 한다면, §E.3 Gaps 와 CHANGELOG 의 “standing gaps” 문단에 **레드 1건을 명시적으로** 적고 `new_warnings_or_lints_introduced` 값을 고친다. 덮지 않는다.

### F3 [Medium] [non-blocking] — run 단계 신호를 sync 종결 **뒤에** 썼다 (3-phase close 순서 위반)

- 관측: `git show 11fa72743:.moai/specs/.../progress.md` 의 §E.3 본문은 `_<run 단계 대기>_` 자리표시자였다. run 신호 111줄은 그 뒤 `170988694`(“write the run-phase audit-ready signal”, trailer `Authored-By-Agent: manager-develop`)에서 들어왔다. 그런데 `11fa72743` 이 이미 `spec.md` 를 `in-progress → completed` 로 넘겼다.
- 왜 결함인가: 커밋 그래프가 유일한 순서 증인이다(`verification-claim-integrity.md` §2.3). 그래프가 증언하는 순서는 “종결 → run 신호 작성”이고, sync 단계가 완성된 run 신호를 읽고 닫았다는 전제와 어긋난다. `spec-frontmatter-schema.md` 의 backfill 면제는 **§E.3/§E.4 의 SHA 필드**에만 열려 있지 run 신호 본문 전체에는 열려 있지 않다. `272521e45`(sync_commit_sha 백필)는 면제 범위 안이라 문제 없다.
- **요구되는 수리**: 이 카드에서는 되돌릴 수 없으므로(이력은 이미 그 모양이다) §E.4 에 “§E.3 은 sync 종결 뒤에 작성됐고, 따라서 sync 종결이 읽은 run 신호는 자리표시자였다”를 한 줄로 기록한다. 다음 카드부터는 sync 커밋 **전에** §E.3 을 닫는다.

### F4 [Medium] [non-blocking] — acceptance.md 의 n-held 뮤턴트 설명이 실제 관측과 다르다

- 위치: `.moai/specs/SPEC-RESOURCE-SLOT-LEASE-001/acceptance.md:263` — “‘보유자 있음’ 검사 삭제 → n-held(빈 기록의 영값 만료 시각 때문에 `allow-expired`로 허용되며, 사유 단언이 이를 잡는다)”.
- 관측: `.moai/reports/t607/m4/mutants/n-held.txt:10` 은 **허용이 아니라 거부**를 찍었다 — `decision = deny (SLOT_LEASE_VIOLATION: resource "demo" is held by  (session , pid 0) …), want allow`. 코드가 그렇게 만든다: `internal/kanban/slot_lease.go:189-198` `Expired()` 는 `!Held()` 일 때 false 를 돌려주므로, `!lease.Held()` 가지를 지우면 `allow-expired` 로 새는 것이 아니라 `default:` 의 `guard-deny` 로 떨어진다(`internal/hook/slot_lease_guard.go:135-150`).
- 판별력 자체는 살아 있다(그 행은 실패한다). 틀린 것은 **기술된 메커니즘**이고, 그 기술이 plan-audit 2회차와 run 단계를 그대로 통과했다.
- **요구되는 수리**: `acceptance.md:263` 의 괄호 설명을 관측대로 고친다 — “보유자 없는 기록이 `guard-deny` 로 떨어져 n-held 행이 실패하며, 거부 사유에 보유자 이름이 비어 있다”.

### F5 [Low] [optional] — `pid_source` 가 상수라서 “해석 불가”를 구분하지 못한다

- 위치: `internal/kanban/slot_lease.go:344` — `PIDSource: PIDSourceSessionOwner` 가 `req.PID == 0`(해석 실패) 에서도 똑같이 `"session-owner"` 로 박힌다.
- AC-RSL-005c 의 “pid 출처 표시가 있으며”는 문자로는 충족된다(필드가 존재한다). 다만 기록을 읽는 사람이 “세션 소유자 pid 를 풀었다”와 “풀지 못해 0 을 적었다”를 그 필드로 가를 수 없다. 통합 창에서 물려받은 모양이므로 회귀는 아니다.
- 선택적 수리: `req.PID <= 0` 일 때 별도 출처 값을 기록하거나, SPEC 에 “이 필드는 스키마 표지이지 해석 결과가 아니다”를 한 줄 적는다.

### F6 [Low] [optional] — CLI 와 가드의 루트 폴백 방향이 다르다

- 위치: `internal/cli/slot.go:78-81` — `ResolveSlotLeaseRoot` 가 실패하면 CLI 는 **정규화하지 않은 `start`** 에 기록을 쓴다. 가드는 같은 실패에서 아무것도 읽지 않고 fail-open 한다(`internal/hook/slot_lease_guard.go:108-113`).
- 주석은 “git 저장소 밖에는 어긋날 워크트리가 없다”만 정당화한다. 저장소 **안**인데 정규화가 실패하는 경우(bare 저장소, `git` 부재, common dir 이 `.git` 이 아닌 경우 — `slot_lease.go:263-265`)에는 CLI 가 가드가 결코 읽지 않을 자리에 기록을 남긴다.
- 선택적 수리: 폴백을 “저장소 아님”으로 좁히거나, 이 비대칭을 규칙 파일에 명시한다.

### F7 [Low] [optional] — 감사 로그에 회전·상한이 없다

- 위치: `internal/kanban/slot_lease.go:527-551` `AppendSlotLeaseAudit` — `O_APPEND` 단조 증가. 가드가 켜진 프로젝트에서는 매칭되는 Bash 호출마다 한 줄이 붙는다.
- 이미 §E.3 Residual risk 와 §E.4 `residual_risk` 에 정직하게 적혀 있다. 이 카드의 결함이라기보다 후속 대상.

## 6. 범위 규율

운영자 결정(verdict.md §1 B안)의 각 행이 배송본에서 지켜졌는지 확인했다.

| 결정 행 | 배송본에서 확인한 것 |
|---|---|
| 표면 = `moai slot acquire\|status\|release --resource <name>` | `internal/cli/slot.go:103` `cmd.AddCommand(newSlotAcquireCmd(), newSlotStatusCmd(), newSlotReleaseCmd())`, 각 verb 에 `--resource` required |
| 선택형 PreToolUse 가드, 기본 꺼짐 | `internal/config/defaults.go:920-923` `SlotLease{Enabled:false, …}`; 템플릿 `workflow.yaml:128` `enabled: false`; 꺼진 경로에서 루트조차 풀지 않음(`internal/hook/pre_tool.go:565-576` — `h.projectRoot()` 가 enabled 가지 **안쪽**에서만 호출됨) |
| 템플릿+바이너리 배포, 문서 최소화 | 템플릿 `workflow.yaml` 22줄 추가 + 규칙 파일 52줄. 규칙 파일은 `paths:` 프론트매터를 달고 있어 always-loaded 표면에 들어가지 않는다(`rule-authoring.md` 준수) |
| 기록 필드(자원·세션 id/이름·pid·명령·시작 시각·선언 상한) | `internal/kanban/slot_lease.go:127-138` `SlotLease` 구조체 — 전 필드 존재, `SessionName`/`Command` 에 `omitempty` 없음(생략 시 빈 문자열로 남김 = REQ-RSL-002) |
| 대조군 필수 | `slot_lease_cross_test.go` 두 갈래, 위 §4 에서 재현 |
| 재사용 경계 — 통합 창과 분리 | 기록/락/설정 키/거부 접두어/CLI 전부 다른 이름. 통합 창 소스 3파일 무변경(§3 프로브) |
| 문서 편집은 t637 병합 뒤 | 게이트 rc=0 재측정 완료(§3). 세 문서 편집은 정당하다 |

카드 diff 66파일을 훑어 **운영자 결정 밖의 배송물은 찾지 못했다**. `internal/template/rule_template_mirror_test.go` 의 +5줄은 새 미러 쌍 등록이라 범위 안이다. 흡수 병합 `9c288daa9`(develop `8d42587e6`)가 끌고 온 dr0912 등의 변경은 merge-base 기준 diff 밖이므로 t607 에 잘못 귀속되지 않았다.

## 7. 5절 증거 형식

### Claim

1. 슬롯 임대 코어·가드·설정의 AC 16개는 이 트리에서 충족된다.
2. 대조군과 뮤턴트 2종은 내가 직접 재현했고, 대조군은 타이밍 운이 아니라 파일 마커로 구성된 경합이다.
3. 템플릿은 언어 중립이고 로컬↔템플릿 미러는 바이트 동일하며, 통합 창은 무변경이다.
4. **그러나** 이 카드가 `internal/config` 에 레드 1건을 들여왔고, §E.3·CHANGELOG 의 “신규 경고 0 / 전부 PASS” 주장은 자기 baseline 에서 거짓이다.

### Evidence

§1·§3·§4·§5 의 명령과 원문 출력. 요약: kanban 40 PASS·hook 29 PASS·template 20 PASS·kanban 전체 `ok 157.573s`·hook 전체 `ok 152.466s`·template 전체 4패키지 `ok`·lint `0 issues.`·spec lint `✓ No findings` — 그리고 `go test ./internal/config/ -count=1` **종료 코드 1**.

### Baseline-attribution

전부 이번 실행, 이 트리(`WT-heavy-test-slot` HEAD `170988694`, 카드 base `8d42587e6`)에서 잰 값이다. 인용한 커버리지 수치도 내가 이 트리에서 `-coverprofile` 로 다시 뽑았다(기록된 86.2%/94.4% 대비 85.9%/94.4% — 문턱 85% 위). `moai spec lint` 는 이 트리에서 빌드한 `./bin/moai`(Commit `9c288daa9`)로 돌렸고, HEAD 와의 1커밋 차이는 `progress.md` 뿐이라 판정에 영향이 없다.

### Gaps (이번 감사에서 관측하지 못한 것)

- **AC-RSL-003c 와 `moai slot` CLI 동작 전부** — heavy-test 슬롯이 다른 레인에 있어 `internal/cli` 를 컴파일·링크하지 않았다. `.moai/reports/t607/m3/*.txt` 를 읽었을 뿐 재실행하지 않았다.
- **AC-RSL-014(i) `make build`** — 같은 이유로 잴 수 없다.
- **`go test ./internal/cli/` 전체 패키지 판정** — 여전히 없다. 이 공백은 run·sync 단계가 이미 정직하게 선언했고 내가 좁히지 못했다.
- **두 플랫폼 빌드** — `GOOS=windows go build ./...` 는 `internal/cli` 를 링크하므로 돌리지 않았다. 증거로만 읽었다.
- **AC-RSL-011 뮤턴트 표 8행(ac016 외)** — `summary.tsv` 와 행별 파일을 읽었고, 그 중 하나만 직접 재현했다.
- **AC-RSL-014(e) 양성 대조(규칙 파일 사본에 `pytest` 주입)** — 재현하지 않았다. 대신 EL-8 계열 양성 대조(python.md 7 적중)를 내가 직접 돌려 토큰 목록의 판별력은 확인했다.
- **Windows 행동** — 이 머신에 판정 수단이 없다.

### Residual-risk

- F1 을 고쳐도 `internal/cli` 전체 판정 공백은 남는다. develop push 가 일으키는 CI 가 첫 판정이다.
- 가드는 이 저장소 어디에서도 켜져 있지 않다. deny 경로가 실제 사용자 명령을 거절한 장면은 단위 테스트 밖에 존재하지 않는다.
- Windows 변경 락의 잔재 정리(`slot_lease_mutation_windows.go`)는 컴파일까지만 검증됐다.
- `resources:` 값 자체가 맵이 아닐 때 관대한 디코딩이 닿지 않아 workflow 섹션 전체가 기본값(= 가드 꺼짐)으로 떨어지는 한계는 CHANGELOG 에 명시돼 있으나 여전히 “조용한 꺼짐”이다.
- 감사 로그 무한 증가(F7).

## 8. 재감사 범위 (델타)

F1 을 고친 뒤의 확인 감사는 아래로 한정한다 — 전체 재감사가 아니다.

1. `go test ./internal/config/ -count=1` → 종료 코드 0, 원문 기록.
2. `internal/config/testdata/shipped_key_inventory.yaml` 에 두 키가 **R** 등급으로 들어갔는지 확인.
3. §E.3 `new_warnings_or_lints_introduced` 와 CHANGELOG 의 해당 문장이 재측정값과 일치하는지 확인(F2).
4. F3·F4 의 기록 수정 여부 확인.
