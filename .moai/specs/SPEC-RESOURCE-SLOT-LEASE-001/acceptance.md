---
id: SPEC-RESOURCE-SLOT-LEASE-001
title: "Acceptance criteria — resource slot lease (card t607)"
version: "0.2.0"
created: 2026-09-12
updated: 2026-09-12
author: manager-spec
tier: M
---

# Acceptance Criteria — SPEC-RESOURCE-SLOT-LEASE-001

**문서 수준 측정 트리 핀: `c4ce42eca`** (`WT-heavy-test-slot`, worktree t607). 자체 핀이 없는 모든 RED-now 셀은 이 트리에 묶인다. 0.2.0에서 추가한 원장 항목 EL-7·EL-8은 `e50cfea93`에서 쟀고, 두 트리 사이의 코드 차이는 없다(`git diff --stat c4ce42eca e50cfea93 -- internal cmd pkg` → 출력 없음, 종료 코드 0).

작성 규율: 모든 기준은 Given-When-Then 형식이고, 판정 명령과 기대 출력을 가진다. 릴리스 차단 기준은 두 셀(RED-now + green path)을 짝으로 가진다. RED-now의 네 요소(명령, 원문 출력, 종료 코드, 트리 SHA)는 아래 §D.0 증거 원장에 두고 셀은 원장 id로 인용한다.

테스트 격리: 모든 테스트는 `t.TempDir()` 아래에서 돌고 `CLAUDE_PROJECT_DIR` + `GIT_CEILING_DIRECTORIES`를 고정하며, `CLAUDE_CODE_SESSION_ID`와 `MOAI_SESSION_PID`를 고정하거나 비운다(레인 환경에 두 값이 설정돼 있어, 비우지 않으면 실제 세션 id와 pid가 기록에 박힌다).

스윕 규율: `go test` 판정은 스윕 수를 먼저 확인한다 — `-v` 출력에 기준이 이름 댄 테스트의 `--- PASS` 줄이 실제로 있어야 하며, `[no tests to run]`이나 해당 줄 부재는 통과가 아니라 공백이다.

## §D.0 증거 원장

```text
EL-1  command : /usr/bin/grep -rlE 'SlotLease|slot_lease|slot-lease' internal cmd pkg
      stdout  : (empty)
      exit    : 1
      tree    : c4ce42eca
      meaning : 임대 표면(코드·설정 키·템플릿 키)이 어디에도 없다. internal/ 아래이므로 internal/template/templates 도 스캔 범위에 들어간다.

EL-2  command : /usr/bin/grep -rlE 'IntegrationLock' internal/kanban/integration_lock.go
      stdout  : internal/kanban/integration_lock.go
      exit    : 0
      tree    : c4ce42eca
      meaning : EL-1의 대조군. 같은 grep 바이너리와 옵션이 존재하는 토큰을 잡는다(셸의 grep 래퍼가 조용히 건너뛰는 경우를 배제하려고 /usr/bin/grep을 쓴다).

EL-3  command : /usr/bin/grep -rl 'review_gate' internal/template/templates/.moai/config/sections
      stdout  : internal/template/templates/.moai/config/sections/workflow.yaml
      exit    : 0
      tree    : c4ce42eca
      meaning : EL-1의 두 번째 대조군. 재귀 스캔이 점(.)으로 시작하는 템플릿 디렉터리 안까지 들어간다.

EL-4  command : /usr/bin/grep -rn 'Use: *"slot' internal/cli
      stdout  : (empty)
      exit    : 1
      tree    : c4ce42eca
      meaning : `moai slot` cobra 명령이 없다.

EL-5  command : git merge-base --is-ancestor WT-acquire-branch-record develop
      stdout  : (empty)
      exit    : 1
      tree    : c4ce42eca
      refs    : 측정 시점(트리 c4ce42eca) 값 — WT-acquire-branch-record = f680dab46, develop = eb50af5a8.
                develop은 움직이는 ref다. 감사 시점에는 e82ef5565, 0.2.0 작성 시점에는 ac6c42c2d였고,
                0.2.0 작성 시점에 같은 명령을 다시 실행해도 종료 코드는 1이었다.
      meaning : M6 진입 조건이 거짓이다. 레인 문서 편집은 보류된다. 게이트 명령은 움직이는 ref 자체를 판정하므로 SHA를 핀하지 않는다.

EL-6  command : go test ./internal/template/ -run 'TestTemplateNeutralityAudit$|TestTemplateNoInternalContentLeak$' -count=1 -v 2>&1 | /usr/bin/grep -E '^(--- |ok|FAIL)'
      stdout  : --- PASS: TestTemplateNoInternalContentLeak (0.66s)
                --- PASS: TestTemplateNeutralityAudit (0.00s)
                ok  	github.com/modu-ai/moai-adk/internal/template	1.071s
      exit    : 0
      tree    : c4ce42eca
      meaning : 템플릿 중립성·유출 가드의 기준선이 초록이고 스윕 수가 2다. 파이프를 쓴 필터 명령이므로 단일 호출 형식의 RED 셀이 아니라 회귀 가드의 기준선으로만 인용한다.

EL-7  command : /usr/bin/grep -nwiE '<TOOL_TOKENS>' internal/template/templates/.moai/config/sections/workflow.yaml
      stdout  : (empty)
      exit    : 1
      tree    : e50cfea93
      meaning : 언어 도구 토큰 목록(§D.2 AC-RSL-014의 TOOL_TOKENS, 명시 열거)이 현재 템플릿 workflow.yaml 어디에도 없다. 새 slot_lease 블록이 들어간 뒤에도 0이어야 한다.

EL-8  command : /usr/bin/grep -cwiE '<TOOL_TOKENS>' internal/template/templates/.claude/rules/moai/languages/python.md
      stdout  : 7
      exit    : 0
      tree    : e50cfea93
      meaning : EL-7의 양성 대조군. 같은 토큰 목록과 같은 grep 옵션이 도구 토큰을 담은 파일에서 적중한다. EL-7의 0은 목록이 무력해서 나온 값이 아니다.
```

`<TOOL_TOKENS>`는 AC-RSL-014에 적힌 토큰 목록 문자열 그대로다(원장에서는 줄 길이 때문에 이름으로 가리킨다).

## §D AC 행렬

| AC | 요구사항 | 종류 | 뒤집는 마일스톤 |
|---|---|---|---|
| AC-RSL-001 | REQ-RSL-003 | 대조군 + RED-first(교차 프로세스) + 뮤턴트 관측(001b) | M2 |
| AC-RSL-002 | REQ-RSL-004 | RED-first(busy ≠ held, 양방향) | M2 |
| AC-RSL-003 | REQ-RSL-002, REQ-RSL-001 | RED-first(기록 필드 + 생략 필드 + CLI 표면) | M2, M3 |
| AC-RSL-004 | REQ-RSL-004 | RED-first | M2 |
| AC-RSL-005 | REQ-RSL-005 | RED-first + 뮤턴트 짝 | M2 |
| AC-RSL-006 | REQ-RSL-006 | RED-first + 뮤턴트 짝 | M2 |
| AC-RSL-007 | REQ-RSL-007 | RED-first | M2 |
| AC-RSL-008 | REQ-RSL-009 | RED-first | M2 |
| AC-RSL-009 | REQ-RSL-010 | RED-first(입력 검증) | M2 |
| AC-RSL-010 | REQ-RSL-012 | RED-first(기본 꺼짐) | M4 |
| AC-RSL-011 | REQ-RSL-011, REQ-RSL-013 | RED-first + 복합 조건 뮤턴트 표 | M4 |
| AC-RSL-012 | REQ-RSL-014 | RED-first(fail-open) | M4 |
| AC-RSL-013 | REQ-RSL-001 | 회귀 가드(분리) | M2-M5 |
| AC-RSL-014 | REQ-RSL-015 | 회귀 가드 + RED-first(템플릿 키·중립성) | M5 |
| AC-RSL-015 | REQ-RSL-016 | 게이트 가지 판정 | M6 |
| AC-RSL-016 | REQ-RSL-008 | RED-first(가드 루트 정규화) + 뮤턴트 짝 | M4 |

0.1.0 대비 번호 이동: 옛 AC-RSL-002(뮤턴트 관측)는 AC-RSL-001b로, 옛 AC-RSL-016(CLI·교차 플랫폼)은 AC-RSL-003c로 옮겼다. 비운 두 번호에 busy≠held(AC-RSL-002)와 가드 루트 정규화(AC-RSL-016)를 넣었다.

## §D.1 심각도

- MUST-PASS(릴리스 차단): AC-RSL-001(001a), 002, 003, 004, 005, 006, 007, 008, 009, 010, 011, 012, 014, 016.
- MUST-PASS(관측 기록): AC-RSL-001b — 되돌림 뮤턴트의 실패 출력이 progress.md §E.2에 원문으로 남아야 한다.
- MUST-PASS(회귀 가드): AC-RSL-013 — RED-now가 없는 보존 불변식이며 뮤턴트 짝으로 의미를 확보한다.
- MUST-PASS(게이트 가지): AC-RSL-015 — run 단계 종료 시점의 게이트 판정이 정한 가지 하나를 만족해야 한다.

## §D.2 Given-When-Then 기준

### AC-RSL-001 — 대조군: 확인-후-시작은 둘 다 시작하고, 임대는 하나를 거절한다

#### AC-RSL-001a — 두 갈래 대조군

**Given** 임시 프로젝트 루트 하나와, 서로 다른 세션 id를 가진 자식 프로세스 두 개(A, B). 두 자식이 기록하는 소유자 pid는 **부모 테스트 프로세스의 pid**로 고정한다(테스트 내내 살아 있는 pid). 자식이 자기 pid를 기록하면 쓰고 곧 종료해 기록이 스테일이 되고, 다른 자식이 정당하게 인수해 락이 옳아도 이중 성공이 나기 때문이다(출처 `internal/kanban/integration_lock_cross_test.go`:180-183).
**When** 자식 A는 테스트 전용 끼어들기 지점에서 멈추고 STALLED 표지를 남긴다. 부모는 STALLED를 확인한 뒤 B를 띄우고, **B의 완료 또는 풀림 타임아웃 중 먼저 오는 쪽**에 A를 푼다. 풀림 타임아웃은 변경 락 대기 예산의 **1/3 이하**다(출처 `integration_lock_cross_test.go`:46-63 — 500ms, 예산 1.65s의 30.3%). 경합은 기다려서 얻지 않고 구성한다.
- (a) 대조 갈래: 두 자식이 표면 없이 "자원 사용 흔적이 없는지 확인 → 시작 흔적 기록"을 수행하고, A는 확인 직후 멈춘다. 락이 없으므로 B가 곧 끝나고, A는 B 완료 뒤 풀린다.
- (b) 표면 갈래: 두 자식이 같은 자원에 임대 획득을 수행하고, A는 판정과 쓰기 사이(변경 락 임계 구역 안)에서 멈춘다. 올바른 직렬화 아래에서 B는 락을 기다리므로 A는 풀림 타임아웃에 풀리고, B는 남은 예산 안에 A가 쓴 기록을 읽는다.

**Then** (a) 시작 수가 2이고, (b) 정확히 한 자식이 획득에 성공하고 다른 한 자식이 "보유 중(held)" 오류로 거절되며, 남은 기록의 보유자는 성공한 자식의 세션 id이고 기록이 스테일이 아니다. (b)에서 어느 자식이든 결과가 busy면 **하네스 설정 결함**으로 실패한다 — 판정이 아니라 구성 오류다. A가 STALLED 표지를 대기 한도 안에 남기지 않으면 끼어들기가 구성되지 않은 것이므로 역시 하네스 결함으로 실패한다.

- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_ControlGroupTwoSessions$' -count=1 -v`
- 기대 출력: `--- PASS: TestSlotLease_ControlGroupTwoSessions`, 하위 테스트 두 개(`probe_then_start` / `lease`)의 `--- PASS` 줄, 로그 `control: starts=2`와 `lease: acquired=1 refused=1 busy=0`.
- **RED-now:** 표면이 없다(EL-1, 대조군 EL-2·EL-3). 표면 갈래는 컴파일되지 않거나 실패한다. M1에서 이 테스트를 먼저 쓰고 관측한 실패 출력이 채택 증거다. 대조 갈래는 양성 대조군으로, 표면 전후 모두 통과해야 한다 — 이 갈래가 실패하면 하네스가 이중 시작을 관측할 수 없다는 뜻이므로 표면 갈래의 통과도 해석할 수 없다.
- **Green path:** M2의 원자적 획득이 표면 갈래를 뒤집는다.

#### AC-RSL-001b — 뮤턴트: 직렬화를 끄면 표면 갈래가 이중 보유로 실패한다

**Given** M2 완료 트리에서 자원별 변경 락 임계 구역을 한 줄 되돌림으로 끈 임시 변형(임계 구역 없이 판정 함수를 직접 호출).
**When** AC-RSL-001a의 판정 명령을 실행한다.
**Then** 테스트가 실패하고, 실패 출력에 `lease: acquired=2 refused=0`(서로 다른 두 세션이 모두 성공, 기록은 스테일 아님)이 나타난다. 변형을 되돌린 뒤 같은 명령이 다시 통과한다.

- 기대 출력: 변형 상태에서 `--- FAIL: TestSlotLease_ControlGroupTwoSessions`와 이중 보유 로그, 복구 후 `--- PASS`.
- **RED-now:** 해당 없음(뮤턴트 관측). 이 관측이 없으면 001a의 통과가 락 때문인지 구성의 우연 때문인지 가를 수 없다.
- **Green path:** M2 종료 시 수행하고, 두 출력을 progress.md §E.2에 원문으로 남긴다. 변형은 커밋하지 않는다.

### AC-RSL-002 — busy는 held가 아니다(양방향)

**Given** 자원 `demo`의 기록이 있고(보유자 `s-1`, 살아 있는 pid, 만료 전), 테스트가 그 자원의 변경 락을 대기 예산보다 오래 쥐고 있다.
**When** (a) 세션 `s-2`가 획득을 시도하고, (b) `s-1`이 해제를 시도한다. 이어서 (c) 테스트가 변경 락을 놓은 뒤 `s-2`가 `--force` 없이 다시 획득한다.
**Then** (a)와 (b)의 오류는 busy 판정 함수가 참이고 held 판정 함수가 거짓이며, 두 경우 모두 기록 파일의 SHA-256이 시도 전후로 같다. (c)의 오류는 held 판정 함수가 참이고 busy 판정 함수가 거짓이다.

- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLeaseBusy_IsNotHeld$' -count=1 -v` (하위 테스트 `acquire_busy`, `release_busy`, `held_is_not_busy`)
- 기대 출력: 상위와 하위 세 개의 `--- PASS` 줄.
- **RED-now:** EL-1 — busy 오류도 판정 함수도 없다.
- **뮤턴트 짝:** busy를 held로 감싸는 변형은 (a)에서, held를 busy로 감싸는 변형은 (c)에서 실패한다.
- **Green path:** M2.

### AC-RSL-003 — 기록 필드, 생략 필드, CLI 표면

- **003a (모든 필드):** **Given** 비어 있는 자원 `demo`. **When** 세션 id `s-1`, 이름 `lane-a`, 명령 텍스트 `heavy-suite`, 상한 5분으로 획득한다. **Then** 디스크 기록과 `--json` 출력 모두에 자원 이름, 세션 id, 세션 이름, 소유자 pid, pid 출처 표시, 명령, 획득 시각(RFC3339 UTC), 선언 상한, 만료 시각(= 획득 시각 + 상한)이 있다. 소유자 pid는 획득 CLI 프로세스 자신의 pid가 아니다.
- **003b (생략 필드):** **Given** 비어 있는 자원 `demo`. **When** `--name`과 `--command` 없이 획득한다. **Then** 획득은 성공하고, 디스크 기록에 세션 이름·명령 키가 **빈 문자열**로 있으며(키 누락도 대체값도 아니다), `moai slot status --resource demo`의 사람용 출력이 두 필드를 `(not given)`으로 보여 준다.
- **003c (CLI 표면과 교차 플랫폼):** **Given** M3 완료 트리. **When** 두 플랫폼으로 빌드하고, 빈 자원에 `moai slot status --resource demo --json`을 실행하고, `moai slot --help`를 본다. **Then** 두 빌드가 성공하고, 상태 JSON이 자원 이름과 보유 여부 `false`를 담으며, 도움말이 세 동작(acquire, status, release)을 보여 준다.
- 판정 명령:
  - `go test ./internal/kanban/ -run 'TestSlotLease_RecordsAllFields$' -count=1 -v`
  - `go test ./internal/cli/ -run 'TestSlotCLI_AcquireJSONFields$|TestSlotCLI_OmittedNameAndCommand$|TestSlotCLI_StatusJSONFree$|TestSlotCLI_HelpListsVerbs$' -count=1 -v`
  - `go build ./...` → 종료 코드 0
  - `GOOS=windows GOARCH=amd64 go build ./...` → 종료 코드 0
- 기대 출력: 이름 댄 다섯 테스트 각각의 `--- PASS` 줄, 두 빌드의 종료 코드 0.
- **RED-now:** EL-1, EL-4 — 기록 타입도 CLI 명령도 없다.
- **Green path:** M2(기록), M3(CLI 출력·도움말).

### AC-RSL-004 — 살아 있는 다른 보유자가 있으면 거절하고 아무것도 쓰지 않는다

**Given** 자원 `demo`의 보유자가 살아 있는 프로세스(테스트 프로세스 자신)를 소유자 pid로 가진 세션 `s-1`이고 만료 전이다.
**When** 세션 `s-2`가 `--force` 없이 획득한다.
**Then** held 오류가 반환되고(held 판정 함수 참), 오류 문구에 보유자 이름이 있으며, 기록 파일의 SHA-256이 획득 시도 전후로 같고, 감사 로그에 `refuse` 한 줄이 생긴다.

- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_RefusesLiveForeignHolderWritesNothing$' -count=1 -v`
- 기대 출력: `--- PASS` 줄.
- **RED-now:** EL-1.
- **Green path:** M2.

### AC-RSL-005 — 생존 판정: 스테일은 인수되고, 살아 있는 보유자와 pid 0은 거절한다

- **005a (스테일 인수):** **Given** 소유자 pid가 이미 종료된 더미 자식 프로세스의 pid인 기록(만료 전). **When** 다른 세션이 `--force` 없이 획득한다. **Then** 성공하고, 새 기록의 `displaced`에 이전 세션 id와 사유 `stale`이 있으며 감사 로그에 `takeover` 한 줄(사유 `stale`)이 있다.
- **005b (살아 있는 보유자, 짝):** **Given** 소유자 pid가 살아 있는 기록(만료 전). **When** 같은 획득. **Then** 거절된다.
- **005c (pid 0 보수적 판정):** **Given** 소유자 pid를 풀 수 없게 만든 획득(pid 해석 입력을 비움)으로 생긴 기록. **When** 기록을 읽고, 다른 세션이 같은 획득을 한다. **Then** 기록의 pid는 0이고(획득 프로세스 pid가 아니다) pid 출처 표시가 있으며, 획득은 거절되고, 상태 조회는 스테일이 아니라고 보고한다.
- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_Liveness$' -count=1 -v` (하위 테스트 `stale_takeover`, `live_refused`, `pid_zero_live`)
- 기대 출력: 상위와 하위 세 개의 `--- PASS` 줄.
- **RED-now:** EL-1.
- **뮤턴트 짝:** "항상 인수 가능" 뮤턴트는 005b·005c에서, "pid를 보지 않음(항상 살아 있음)" 뮤턴트는 005a에서, "풀 수 없으면 자기 pid 기록" 뮤턴트는 005c에서 실패한다.
- **Green path:** M2.

### AC-RSL-006 — 선언 상한: 지나면 살아 있어도 인수되고, 지나기 전에는 거절한다

- **006a (만료 인수):** **Given** 소유자 pid가 살아 있고 만료 시각이 이미 지난 기록. **When** 다른 세션이 `--force` 없이 획득한다. **Then** 성공하고 `displaced` 사유가 `expired`이며, 획득 전 상태 조회는 만료를 보고한다.
- **006b (만료 전, 짝):** **Given** 소유자 pid가 살아 있고 만료 전인 기록. **When** 같은 획득. **Then** 거절된다.
- **006c (보유자 재획득):** **Given** 세션 `s-1`이 보유한 만료 전 기록. **When** `s-1`이 다시 획득한다. **Then** 성공하고 만료 시각이 재획득 시각 + 상한으로 새로 잡힌다(이전 값보다 늦다).
- **006d (상한 입력 검증):** **When** 상한 0, 음수, 해석 불가 문자열로 획득한다. **Then** 모두 오류이고 기록이 생기지 않는다.
- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_DeclaredBound$' -count=1 -v` (하위 테스트 네 개)
- 기대 출력: 상위와 하위 네 개의 `--- PASS` 줄.
- **RED-now:** EL-1.
- **뮤턴트 짝:** "만료 무시" 뮤턴트는 006a에서, "살아 있는 보유자를 모두 만료로 취급" 뮤턴트는 006b에서 실패한다. 시간은 테스트가 주입하는 시계로 제어한다(실제 대기 금지).
- **Green path:** M2.

### AC-RSL-007 — 강제 인수는 밀어낸 보유자를 기록과 감사 로그에 남긴다

**Given** 살아 있고 만료 전인 보유자 `s-1`.
**When** 세션 `s-2`가 `--force`로 획득한다.
**Then** 성공하고, 새 기록의 `displaced`에 `s-1`, 이전 pid, 사유 `force`, 시각이 있으며, 감사 로그에 `takeover` 한 줄(사유 `force`)이 있고, CLI 출력에 밀어낸 보유자가 표시된다.

- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_ForceRecordsDisplaced$' -count=1 -v` 그리고 `go test ./internal/cli/ -run 'TestSlotCLI_ForceReportsDisplaced$' -count=1 -v`
- 기대 출력: 두 `--- PASS` 줄.
- **RED-now:** EL-1, EL-4. (통합 창은 밀어낸 보유자를 표준 출력에만 낸다 — `internal/cli/integration.go`:282-286. 기록에 남기는 것은 이 SPEC의 새 의무다.)
- **Green path:** M2, M3.

### AC-RSL-008 — 해제는 보유자만, 빈 해제는 오류

**Given** 보유자 `s-1`.
**When** (a) `s-2`가 `--force` 없이 해제, (b) `s-1`이 해제, (c) 보유자가 없는 상태에서 다시 해제.
**Then** (a) "다른 세션 보유" 오류이고 기록 불변, (b) 기록 파일이 사라지고 감사 로그에 `release` 한 줄, (c) "보유 없음" 오류.

- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_Release$' -count=1 -v`
- 기대 출력: `--- PASS` 줄과 하위 세 개.
- **RED-now:** EL-1.
- **Green path:** M2.

### AC-RSL-009 — 허용되지 않는 자원 이름은 거부하고 아무 파일도 쓰지 않는다

**Given** 임시 프로젝트 루트.
**When** 자원 이름 `../escape`, `a/b`, `a\b`, 빈 문자열, 65자 이름, 대문자·공백 포함 이름으로 획득·해제·상태를 요청한다.
**Then** 모두 오류이고, 임시 루트 전체에서 요청 전후의 파일 목록이 같다(임대 디렉터리 안팎 모두).

- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_RejectsInvalidResourceNames$' -count=1 -v`
- 기대 출력: `--- PASS` 줄.
- **RED-now:** EL-1.
- **Green path:** M2.

### AC-RSL-010 — 가드 기본 꺼짐: 기록을 읽지 않고 거부하지 않는다

**Given** 매칭 패턴이 설정된 자원 `demo`와 (a) 살아 있는 다른 보유자 기록 또는 (b) 손상된 기록, 그리고 세 가지 설정 상태: (i) `workflow.slot_lease.enabled` 키 없음, (ii) `enabled: false`, (iii) 설정 제공자가 설정을 돌려주지 못함(nil).
**When** 매칭되는 Bash 명령으로 PreToolUse를 호출한다.
**Then** 여섯 조합 모두 허용이고, 표준 오류에 안내가 없으며, 감사 로그 파일이 생기지 않는다. (b)에서 안내가 나오면 꺼진 경로가 기록을 읽었다는 증거이므로 실패다.

- 판정 명령: `go test ./internal/hook/ -run 'TestSlotLeaseGuard_DisabledNeverReadsNorDenies$' -count=1 -v` 그리고 `go test ./internal/config/ -run 'TestDefaults_SlotLeaseDisabled$' -count=1 -v`
- 기대 출력: 두 `--- PASS` 줄.
- **RED-now:** EL-1.
- **뮤턴트 짝:** 설정 확인보다 기록 읽기를 먼저 하는 뮤턴트는 (b)에서, nil 설정을 fail-open 안내로 처리하는 뮤턴트는 (iii)에서 실패한다.
- **Green path:** M4.

### AC-RSL-011 — 가드 거부 복합 조건: 각 항이 제 몫의 실패 사례와 감사 사유를 가진다

**Given** 설정 `enabled: true`, 자원 `demo`의 패턴 목록, 호출 세션 `s-2`, 훅 루트는 기록이 있는 primary 루트.
**When** 표의 각 행 입력으로 PreToolUse를 호출한다.
**Then** 거부 행은 사유가 `SLOT_LEASE_VIOLATION:`으로 시작하고 자원 이름과 보유자를 담는다. 허용 행은 허용되고, 감사 로그의 마지막 줄 사유가 표의 기대 사유와 같다("감사 줄 없음" 행은 감사 로그 줄 수가 호출 전후로 같다).

| 행 | 설정 켜짐 | 패턴 일치 | 보유자 ≠ 호출자 | 보유자 생존 | 만료 전 | 기대 | 기대 감사 사유 |
|---|---|---|---|---|---|---|---|
| base | 예 | 예 | 예 | 예 | 예 | 거부 | `guard-deny` |
| n-enabled | 아니오 | 예 | 예 | 예 | 예 | 허용 | 감사 줄 없음 |
| n-match | 예 | 아니오 | 예 | 예 | 예 | 허용 | 감사 줄 없음 |
| n-quoted | 예 | 따옴표 안에서만 | 예 | 예 | 예 | 허용 | 감사 줄 없음 |
| n-self | 예 | 예 | 아니오(보유자 = 호출자) | 예 | 예 | 허용 | `allow-self` |
| n-alive | 예 | 예 | 예 | 아니오(스테일) | 예 | 허용 | `allow-stale` |
| n-bound | 예 | 예 | 예 | 예 | 아니오(만료) | 허용 | `allow-expired` |
| n-held | 예 | 예 | 보유자 없음 | — | — | 허용 | `allow-unheld` |
| multi | 예 | 두 자원 중 하나만 거부 조건 | 예 | 예 | 예 | 거부(그 자원 이름) | `guard-deny` |

- 판정 명령: `go test ./internal/hook/ -run 'TestSlotLeaseGuard_DenyMatrix$' -count=1 -v`
- 기대 출력: 상위와 행마다 하위 `--- PASS` 줄(9개).
- **뮤턴트 표(각 항의 검사를 하나씩 지운 변형이 실패하는 행):** 설정 검사 삭제 → n-enabled, 패턴 검사 삭제 → n-match, 따옴표 제거 삭제 → n-quoted, 세션 비교 삭제 → n-self, 생존 검사 삭제 → n-alive, 만료 검사 삭제 → n-bound, "보유자 있음" 검사 삭제 → n-held(빈 기록의 영값 만료 시각 때문에 `allow-expired`로 허용되며, 사유 단언이 이를 잡는다), 자원별 판정을 첫 자원만 보도록 바꾸기 → multi. 허용 행마다 사유를 단언하므로, 한 허용 경로를 지워도 다른 허용 경로로 새는 변형이 사유 불일치로 실패한다.
- **RED-now:** EL-1.
- **Green path:** M4.

### AC-RSL-012 — 가드 fail-open: 켜진 뒤의 불확실성은 허용하고 흔적을 남긴다

**Given** 설정 `enabled: true`(참으로 읽힘)와 매칭 패턴.
**When** (a) 손상된 기록, (b) 호출 세션 id가 빈 입력, (c) 컴파일되지 않는 패턴(`(`), (d) 프로젝트 루트가 빈 입력, (e) 자원 항목의 명령 목록이 패턴 문자열 목록으로 해석되지 않는 설정(예: 목록 대신 숫자, 목록 원소가 문자열이 아님)으로 매칭 명령을 호출한다.
**Then** 다섯 경우 모두 허용이고, 각각 표준 오류에 `[moai:slot-lease] advisory:` 줄이 있으며, (d)를 뺀 네 경우는 감사 로그에 `fail-open` 한 줄이 사유와 함께 남는다((d)는 로그를 쓸 루트가 없으므로 표준 오류만 요구한다). (e)에서 `enabled`는 참으로 유지돼야 한다 — 잘못된 자원 항목 하나가 설정 전체를 꺼짐으로 떨어뜨리면 이 경우는 조용한 꺼짐(REQ-RSL-012)으로 사라지므로 실패다.

- 판정 명령: `go test ./internal/hook/ -run 'TestSlotLeaseGuard_FailOpen$' -count=1 -v` 그리고 `go test ./internal/config/ -run 'TestSlotLeaseConfig_MalformedResourceKeepsEnabled$' -count=1 -v`
- 기대 출력: 가드 테스트의 상위와 하위 다섯 개 `--- PASS` 줄, 설정 테스트의 `--- PASS` 줄.
- **RED-now:** EL-1.
- **Green path:** M4.

### AC-RSL-013 — 통합 창과의 분리(회귀 가드)

**Given** M5 완료 트리, 카드 브랜치가 아직 develop에 병합되지 않은 상태.
**When** (a) 슬롯 획득을 해도 통합 기록 파일이 생기지 않고, 통합 획득을 해도 슬롯 임대 디렉터리가 생기지 않는지 확인한다. (b) 통합 창의 기존 테스트를 돌린다. (c) 통합 창 소스 파일이 이 카드에서 바뀌지 않았는지 `.claude/rules/local/gitflow-lane-protocol.md` §8 형태로 확인한다.
**Then** (a) 두 방향 모두 상대 파일이 없고, (b) 기존 테스트가 모두 통과하며, (c) 대조군이 1 이상이고 프로브 출력이 없다.

- 판정 명령:
  - `go test ./internal/kanban/ -run 'TestSlotLease_SeparateFromIntegrationWindow$' -count=1 -v`
  - `go test ./internal/kanban/ -run 'IntegrationLock' -count=1 -v`
  - `go test ./internal/hook/ -run 'IntegrationLock' -count=1 -v`
  - (c) 읽는 시점에 기준을 다시 구한다(값을 핀하지 않는다):
    ```bash
    CARD_BASE=$(git merge-base develop HEAD)
    git diff --name-only "$CARD_BASE"..HEAD | wc -l                                   # 대조군: 1 이상
    git diff --name-only "$CARD_BASE"..HEAD -- internal/cli/integration.go internal/hook/integration_lock_guard.go   # 프로브: 출력 없음
    ```
- 기대 출력: 첫째 명령 `--- PASS`, 둘째·셋째 명령은 `--- FAIL` 없이 `ok`이며 스윕 수가 0이 아니고, (c)의 대조군은 1 이상, 프로브는 출력이 없다. **대조군이 0이면 "변경 없음"이 아니라 "측정 불가"로 보고한다.** 흡수 기준 ref는 이 저장소 절차의 흡수 대상인 로컬 `develop`이다 — develop을 흡수한 뒤에도 merge-base가 마지막으로 흡수한 develop 커밋에 머물러, 다른 카드가 그 파일들을 고친 커밋은 범위에 들어오지 않는다. 기반 함수를 매개변수화해야 해서 `internal/kanban/integration_lock_mutation.go`가 바뀌는 경우, 그 파일은 프로브 범위에 넣지 않되 둘째 명령의 전체 통과로 동작 불변을 보인다.
- **유효 시점:** (c)는 **병합 전에만** 유효하다. 병합 뒤에는 merge-base가 카드 tip 자신이 되어 범위가 비고 판정이 공허하게 통과한다. 병합 뒤의 근거는 병합 트리와 카드 브랜치 트리의 동일성이다(`git rev-parse <merge>^{tree}` = 병합 전 재측정에 쓴 카드 tip의 `git rev-parse <card-tip>^{tree}`).
- **RED-now:** 해당 없음(보존 불변식). 기준선: 통합 창 테스트는 현재 트리에 존재한다(`internal/kanban/integration_lock_test.go`, `integration_lock_cross_test.go`).
- **뮤턴트 짝:** 슬롯 기록을 통합 기록 경로에 쓰는 변형은 (a)에서 실패한다.
- **Green path:** M2-M5 동안 유지.

### AC-RSL-014 — 템플릿 배포: 키는 꺼진 채로 싣고, 언어 중립이며, 유출 가드를 통과한다

**Given** M5 완료 트리와 `make build`로 다시 컴파일한 바이너리.
**When** 템플릿 설정과 새 규칙 파일을 검사하고 중립성·유출 테스트를 돌린다.
**Then** 아래 판정이 모두 기대대로 나온다.

`TOOL_TOKENS` — 16개 지원 프로그래밍 언어의 대표 도구 토큰 명시 목록(대소문자 무시, 단어 경계 `-w`):

```text
go test|go build|pytest|unittest|tox|npm|npx|yarn|pnpm|jest|vitest|mocha|tsc|cargo|rustc|mvn|maven|junit|gradle|kotlinc|dotnet|msbuild|xunit|nunit|rspec|rake|bundle exec|minitest|phpunit|composer|pest|mix test|exunit|ctest|cmake|make test|gtest|sbt|scalatest|rscript|testthat|flutter|dart|swift test|xcodebuild|swiftpm
```

언어별 대응: go(go test, go build), python(pytest, unittest, tox), typescript(tsc, jest, vitest), javascript(npm, npx, yarn, pnpm, mocha), rust(cargo, rustc), java(mvn, maven, junit), kotlin(gradle, kotlinc), csharp(dotnet, msbuild, xunit, nunit), ruby(rspec, rake, bundle exec, minitest), php(phpunit, composer, pest), elixir(mix test, exunit), cpp(ctest, cmake, make test, gtest), scala(sbt, scalatest), r(rscript, testthat), flutter(flutter, dart), swift(swift test, xcodebuild, swiftpm).

- 판정 명령:
  - (a) 키와 기본값: `/usr/bin/grep -n -A2 'slot_lease:' internal/template/templates/.moai/config/sections/workflow.yaml` → `enabled: false` 줄 포함, 종료 코드 0
  - (b) 자리표시자 구조: `/usr/bin/grep -c -E 'resources: \{\}|<command-regex>' internal/template/templates/.moai/config/sections/workflow.yaml` → 2 이상
  - (c) 언어 도구 토큰 부재(설정): `/usr/bin/grep -nwiE '<TOOL_TOKENS>' internal/template/templates/.moai/config/sections/workflow.yaml` → 출력 없음, 종료 코드 1
  - (d) 언어 도구 토큰 부재(규칙): `/usr/bin/grep -nwiE '<TOOL_TOKENS>' internal/template/templates/.claude/rules/moai/workflow/resource-slot-lease.md` → 출력 없음, 종료 코드 1
  - (e) 양성 대조: 새 규칙 파일을 스크래치 경로로 복사하고 그 사본 끝에 `pytest` 한 줄을 붙인 뒤 (d)와 같은 명령을 사본에 실행 → 적중 1줄 이상, 종료 코드 0. 원본은 건드리지 않는다. 이 적중이 없으면 (d)의 무출력은 해석할 수 없다(EL-8이 설정 쪽 목록의 대조군이다).
  - (f) 내부 토큰 부재: `/usr/bin/grep -rnE 'SPEC-[A-Z]|\bt[0-9]{3}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}' internal/template/templates/.claude/rules/moai/workflow/resource-slot-lease.md` → 출력 없음, 종료 코드 1
  - (g) `go test ./internal/template/ -run 'TestTemplateNeutralityAudit$|TestTemplateNoInternalContentLeak$|TestRuleProvenance' -count=1 -v` → `--- FAIL` 없음, 이름 댄 테스트의 `--- PASS` 줄 존재
  - (h) `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run 'TestTemplateNoInternalContentLeak$' -count=1 -v` → `--- PASS`
  - (i) `make build` → 종료 코드 0
- **RED-now:** 템플릿 키 부재는 EL-1(스캔 범위에 템플릿 포함, 점 디렉터리 대조군 EL-3). (c)는 현재도 무출력이다(EL-7) — 새 블록이 언어 토큰을 들여오지 않았음을 지키는 회귀 가드이며, 목록의 판별력은 EL-8과 (e)가 보인다. 중립성·유출 테스트는 현재 초록이다(EL-6) — 회귀 가드다.
- **뮤턴트 짝:** 새 규칙 파일에 SPEC ID 한 줄을 넣은 임시 변형에서 유출 테스트가 실패함을 관측하고 되돌린다(출력은 progress.md §E.2). 템플릿 자리표시자를 특정 언어 명령으로 바꾼 변형은 (c)에서 적중해 실패한다.
- **Green path:** M5.

### AC-RSL-015 — 레인 문서 반영은 게이트를 따른다

**Given** run 단계 종료 시점에 `git merge-base --is-ancestor WT-acquire-branch-record develop`의 종료 코드와 두 ref의 그 시점 SHA를 progress.md에 기록한 상태. 카드 브랜치는 아직 develop에 병합되지 않았다.
**When** 기록된 종료 코드로 가지를 고른다.
**Then**
- **게이트 닫힘(종료 코드 ≠ 0):** 이 카드의 커밋이 세 문서를 바꾸지 않았다. 판정은 `.claude/rules/local/gitflow-lane-protocol.md` §8 형태로, 읽는 시점에 기준을 다시 구한다:
  ```bash
  CARD_BASE=$(git merge-base develop HEAD)
  git diff --name-only "$CARD_BASE"..HEAD | wc -l        # 대조군: 1 이상, 0이면 "측정 불가"
  git diff --name-only "$CARD_BASE"..HEAD -- .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/local/gitflow-lane-protocol.md   # 프로브: 출력 없음
  ```
  progress.md에 후속 카드로 넘긴다는 기록이 있다.
- **게이트 열림(종료 코드 0):** 세 파일 각각에서 `/usr/bin/grep -c 'moai slot' <file>` → 1 이상. 템플릿판 `kanban-dispatch.md`에 대해 AC-RSL-014(g)(h)의 유출 테스트가 통과한다.
- **유효 시점:** 닫힘 가지의 판정은 병합 전에만 유효하다. 병합 뒤의 근거는 병합 트리와 카드 브랜치 트리의 동일성이다.
- **RED-now:** EL-5 — 현재 게이트는 닫혀 있다(종료 코드 1). 이 기준은 run 단계 종료 시점의 판정으로 가지를 정한다.
- **Green path:** M6.

### AC-RSL-016 — 가드 루트 정규화: 워크트리 훅 루트에서도 primary 기록을 읽는다

**Given** 임시 디렉터리 안의 git 저장소(primary)와 그 저장소의 링크된 워크트리 하나(`git worktree add`로 테스트 안에서 만든다). primary의 `.moai/state/slot-leases/demo.json`에 살아 있는 다른 보유자(`s-1`, 부모 테스트 프로세스 pid, 만료 전) 기록이 있고, 워크트리의 `.moai/state/slot-leases/`에는 기록이 없다. 설정 `enabled: true`와 `demo`의 매칭 패턴, 호출 세션 `s-2`.
**When** 훅 프로젝트 루트를 **링크된 워크트리 경로**로 두고(훅 루트 해석이 워크트리를 내놓는 조건 — `CLAUDE_PROJECT_DIR`를 워크트리로 설정) 매칭 명령으로 PreToolUse를 호출한다.
**Then** 거부된다(사유가 `SLOT_LEASE_VIOLATION:`으로 시작하고 보유자 `s-1`을 담는다).

| 행 | 훅 루트 | 기록 위치 | git common dir | 기대 |
|---|---|---|---|---|
| wt-deny | 링크된 워크트리 | primary | 풀림 | 거부 |
| primary-deny | primary | primary | 풀림 | 거부 |
| no-git | git 저장소가 아닌 디렉터리 | (없음) | 풀리지 않음 | 허용 + 표준 오류 안내 + 감사 `fail-open` |

- 판정 명령: `go test ./internal/hook/ -run 'TestSlotLeaseGuard_NormalizesWorktreeRootToPrimary$' -count=1 -v` (하위 테스트 세 개)
- 기대 출력: 상위와 하위 세 개의 `--- PASS` 줄.
- **뮤턴트 행:** 루트 정규화를 지우고 훅 루트에서 바로 기록을 읽는 변형은 wt-deny에서 **허용**을 내므로(워크트리에 기록이 없어 "보유자 없음") 실패한다. primary-deny는 이 변형에서도 거부이므로, 두 행의 대비가 정규화 자체를 판별한다. 정규화 실패 시 비정규화 루트로 넘어가 읽는 변형은 no-git에서 안내 없이 허용해 실패한다.
- **RED-now:** EL-1 — 가드도 정규화도 없다. 근거가 되는 코드 사실: 훅 루트 해석은 `CLAUDE_PROJECT_DIR` → `os.Getwd()`뿐이다(`internal/hook/path_resolve.go`:78-94, 트리 `c4ce42eca`).
- **Green path:** M4.

## §D.3 추적성

| 요구사항 | 인수 기준 |
|---|---|
| REQ-RSL-001 | AC-RSL-003c, AC-RSL-013 |
| REQ-RSL-002 | AC-RSL-003a, AC-RSL-003b |
| REQ-RSL-003 | AC-RSL-001a, AC-RSL-001b |
| REQ-RSL-004 | AC-RSL-002, AC-RSL-004 |
| REQ-RSL-005 | AC-RSL-005 |
| REQ-RSL-006 | AC-RSL-006 |
| REQ-RSL-007 | AC-RSL-007 |
| REQ-RSL-008 | AC-RSL-016 |
| REQ-RSL-009 | AC-RSL-008 |
| REQ-RSL-010 | AC-RSL-009 |
| REQ-RSL-011 | AC-RSL-011 |
| REQ-RSL-012 | AC-RSL-010 |
| REQ-RSL-013 | AC-RSL-011 |
| REQ-RSL-014 | AC-RSL-012 |
| REQ-RSL-015 | AC-RSL-014 |
| REQ-RSL-016 | AC-RSL-015 |

## §D.4 간접 검증 (숨기지 않고 밝힘)

- **Windows 동작.** 로컬 증거는 `GOOS=windows` 컴파일까지다. Windows 변경 락 잔재 정리의 행동 관측은 develop push 뒤 CI의 Windows 경로가 처음이다.
- **서브에이전트 세션 id.** plan-audit 1회차 E3의 관측(서브에이전트 = 부모 세션 id)은 한 세션·한 런타임 한 번이다. 단위 테스트는 훅 입력을 직접 만들므로 런타임이 서브에이전트에 어떤 `session_id`를 주는지는 보지 못한다.
- **워크트리 훅 루트가 실제로 나오는 세션 형태.** AC-RSL-016은 `CLAUDE_PROJECT_DIR`를 워크트리로 두어 조건을 만든다. 어떤 실제 세션 형태(예: `moai cc -w`로 시작한 세션)에서 훅 환경이 워크트리를 가리키는지는 재지 않았다. 정규화는 그 조건이 생기든 아니든 결과가 같도록 요구한다.
- **부하 비용.** 어떤 기준도 무거운 실행이 겹칠 때의 부하를 재지 않는다(범위 밖).

## §D.5 종료 게이트

- AC-RSL-001a·001b의 출력(대조 갈래 `starts=2`, 표면 갈래 `acquired=1 refused=1 busy=0`, 뮤턴트 `acquired=2`)이 progress.md §E.2에 원문으로 있다.
- 모든 MUST-PASS 기준의 판정 명령 출력이 스윕 수와 함께 기록돼 있다.
- AC-RSL-013(c)와 AC-RSL-015 닫힘 가지는 **병합 전에** `CARD_BASE=$(git merge-base develop HEAD)`를 읽는 시점에 다시 구해 잰 결과(대조군 수 포함)가 기록돼 있다. 리터럴 기준 SHA를 미리 잡아 두는 방식은 쓰지 않는다. 병합 뒤에는 병합 트리와 카드 브랜치 트리의 동일성(`git rev-parse <merge>^{tree}`와 카드 tip의 트리)이 기록돼 있다.
- M6 게이트 판정(명령, 종료 코드, 두 ref의 그 시점 SHA)이 기록돼 있다.

## §D.6 완료 정의

- M1-M5 완료, M6은 게이트 가지에 따라 완료 또는 후속 카드 이관.
- 영향 패키지(`internal/kanban`, `internal/cli`, `internal/hook`, `internal/config`, `internal/template`) 테스트가 `-run` 필터 기준으로 통과하고, 새 코드 커버리지가 85% 이상이다(`go test -cover ./internal/<pkg>/...` 원문 출력으로 제시).
- `golangci-lint run` 신규 경고 없음(기존 기준선과 구분해 보고).
- 두 플랫폼 빌드 통과.
- 템플릿 변경 뒤 `make build` 통과.
- 모든 커밋 메시지에 카드 id(`t607`)가 있다.

## §D.7 앞으로 볼 점 (이 카드의 게이트 아님)

- 감사 로그의 `allow-unheld` 빈도가 높으면 보유 강제 모드(OQ-2)의 필요성을 다시 본다.
- 선언 상한 초과 인수(`expired`)가 정상 실행을 자주 넘기면 기본 상한(OQ-3)을 다시 잡는다.
- 통합 창 가드의 루트 정규화는 후속 카드 후보다.
- 가드를 이 저장소 로컬 설정에서 켤지는 운영자 결정이다.
