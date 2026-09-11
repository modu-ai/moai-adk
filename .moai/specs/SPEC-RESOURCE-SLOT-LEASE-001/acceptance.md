---
id: SPEC-RESOURCE-SLOT-LEASE-001
title: "Acceptance criteria — resource slot lease (card t607)"
version: "0.1.0"
created: 2026-09-12
updated: 2026-09-12
author: manager-spec
tier: M
---

# Acceptance Criteria — SPEC-RESOURCE-SLOT-LEASE-001

**문서 수준 측정 트리 핀: `c4ce42eca`** (`WT-heavy-test-slot`, worktree t607). 자체 핀이 없는 모든 RED-now 셀은 이 트리에 묶인다.

작성 규율: 모든 기준은 Given-When-Then 형식이고, 판정 명령과 기대 출력을 가진다. 릴리스 차단 기준은 두 셀(RED-now + green path)을 짝으로 가진다. RED-now의 네 요소(명령, 원문 출력, 종료 코드, 트리 SHA)는 아래 §D.0 증거 원장에 두고 셀은 원장 id로 인용한다. 모든 테스트는 `t.TempDir()` 아래에서 `CLAUDE_PROJECT_DIR` + `GIT_CEILING_DIRECTORIES`를 고정해 돈다. `go test` 판정은 스윕 수를 먼저 확인한다 — `-v` 출력에 기준이 이름 댄 테스트의 `--- PASS` 줄이 실제로 있어야 하며, `[no tests to run]`이나 해당 줄 부재는 통과가 아니라 공백이다.

## §D.0 증거 원장 (측정 트리 `c4ce42eca`)

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
      tree    : c4ce42eca  (WT-acquire-branch-record = f680dab46, develop = eb50af5a8)
      meaning : M6 진입 조건이 거짓이다. 레인 문서 편집은 보류된다.

EL-6  command : go test ./internal/template/ -run 'TestTemplateNeutralityAudit$|TestTemplateNoInternalContentLeak$' -count=1 -v 2>&1 | /usr/bin/grep -E '^(--- |ok|FAIL)'
      stdout  : --- PASS: TestTemplateNoInternalContentLeak (0.66s)
                --- PASS: TestTemplateNeutralityAudit (0.00s)
                ok  	github.com/modu-ai/moai-adk/internal/template	1.071s
      exit    : 0
      tree    : c4ce42eca
      meaning : 템플릿 중립성·유출 가드의 기준선이 초록이고 스윕 수가 2다. 파이프를 쓴 필터 명령이므로 단일 호출 형식의 RED 셀이 아니라 회귀 가드의 기준선으로만 인용한다.
```

## §D AC 행렬

| AC | 요구사항 | 종류 | 뒤집는 마일스톤 |
|---|---|---|---|
| AC-RSL-001 | REQ-RSL-003 | 대조군 + RED-first(교차 프로세스) | M2 |
| AC-RSL-002 | REQ-RSL-003 | 뮤턴트 관측(한 줄 되돌림) | M2 |
| AC-RSL-003 | REQ-RSL-002 | RED-first | M2, M3 |
| AC-RSL-004 | REQ-RSL-004 | RED-first | M2 |
| AC-RSL-005 | REQ-RSL-005, REQ-RSL-008 | RED-first + 뮤턴트 짝 | M2 |
| AC-RSL-006 | REQ-RSL-006 | RED-first + 뮤턴트 짝 | M2 |
| AC-RSL-007 | REQ-RSL-007 | RED-first | M2 |
| AC-RSL-008 | REQ-RSL-009 | RED-first | M2 |
| AC-RSL-009 | REQ-RSL-010 | RED-first(입력 검증) | M2 |
| AC-RSL-010 | REQ-RSL-012 | RED-first(기본 꺼짐) | M4 |
| AC-RSL-011 | REQ-RSL-011, REQ-RSL-013 | RED-first + 복합 조건 뮤턴트 표 | M4 |
| AC-RSL-012 | REQ-RSL-014 | RED-first(fail-open) | M4 |
| AC-RSL-013 | REQ-RSL-001 | 회귀 가드(분리) | M2-M5 |
| AC-RSL-014 | REQ-RSL-015 | 회귀 가드 + RED-first(템플릿 키) | M5 |
| AC-RSL-015 | REQ-RSL-016 | 게이트 가지 판정 | M6 |
| AC-RSL-016 | REQ-RSL-001 | RED-first(CLI + 교차 플랫폼) | M3 |

## §D.1 심각도

- MUST-PASS(릴리스 차단): AC-RSL-001, 003, 004, 005, 006, 007, 008, 009, 010, 011, 012, 014, 016.
- MUST-PASS(관측 기록): AC-RSL-002 — 되돌림 뮤턴트의 실패 출력이 progress.md §E.2에 원문으로 남아야 한다.
- MUST-PASS(회귀 가드): AC-RSL-013 — RED-now가 없는 보존 불변식이며 뮤턴트 짝으로 의미를 확보한다.
- MUST-PASS(게이트 가지): AC-RSL-015 — run 단계 종료 시점의 게이트 판정이 정한 가지 하나를 만족해야 한다.

## §D.2 Given-When-Then 기준

### AC-RSL-001 — 대조군: 확인-후-시작은 둘 다 시작하고, 임대는 하나를 거절한다

**Given** 임시 프로젝트 루트 하나와, 서로 다른 세션 id를 가진 자식 프로세스 두 개(A, B). 자식 A는 테스트 전용 끼어들기 지점에서 멈추고, 부모는 B가 자기 동작을 끝낸 뒤 A를 풀어 준다(경합은 기다려서 얻지 않고 구성한다).
**When** (a) 대조 갈래: 두 자식이 표면 없이 "자원 사용 흔적이 없는지 확인 → 시작 흔적 기록"을 수행하고, A는 확인 직후 멈춘다. (b) 표면 갈래: 두 자식이 같은 자원에 임대 획득을 수행하고, A는 판정과 쓰기 사이에서 멈춘다.
**Then** (a) 시작 수가 2이고, (b) 정확히 한 자식이 획득에 성공하고 다른 한 자식이 "보유 중" 오류로 거절되며, 남은 기록의 보유자는 성공한 자식의 세션 id다.

- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_ControlGroupTwoSessions$' -count=1 -v`
- 기대 출력: `--- PASS: TestSlotLease_ControlGroupTwoSessions`, 하위 테스트 두 개(`probe_then_start` / `lease`)의 `--- PASS` 줄, 로그 `control: starts=2`와 `lease: acquired=1 refused=1`.
- **RED-now:** 표면이 없다(EL-1, 대조군 EL-2·EL-3). 표면 갈래는 컴파일되지 않거나 실패한다. M1에서 이 테스트를 먼저 쓰고 관측한 실패 출력이 채택 증거다. 대조 갈래는 양성 대조군으로, 표면 전후 모두 통과해야 한다 — 이 갈래가 실패하면 하네스가 이중 시작을 관측할 수 없다는 뜻이므로 표면 갈래의 통과도 해석할 수 없다.
- **Green path:** M2의 원자적 획득이 표면 갈래를 뒤집는다.
- **짝 뮤턴트:** AC-RSL-002.

### AC-RSL-002 — 뮤턴트: 직렬화를 끄면 표면 갈래가 이중 보유로 실패한다

**Given** M2 완료 트리에서 자원별 변경 락 임계 구역을 한 줄 되돌림으로 끈 임시 변형(임계 구역 없이 판정 함수를 직접 호출).
**When** AC-RSL-001의 판정 명령을 실행한다.
**Then** 테스트가 실패하고, 실패 출력에 `lease: acquired=2 refused=0`(서로 다른 두 세션이 모두 성공)이 나타난다. 변형을 되돌린 뒤 같은 명령이 다시 통과한다.

- 판정 명령: AC-RSL-001과 같다. 변형 전후 두 번 실행한다.
- 기대 출력: 변형 상태에서 `--- FAIL: TestSlotLease_ControlGroupTwoSessions`와 이중 보유 로그, 복구 후 `--- PASS`.
- **RED-now:** 해당 없음(뮤턴트 관측 기준). 이 기준이 없으면 AC-RSL-001의 통과가 락 때문인지, 테스트 구성의 우연 때문인지 가를 수 없다.
- **Green path:** M2 종료 시 수행하고, 두 출력을 progress.md §E.2에 원문으로 남긴다. 변형은 커밋하지 않는다.

### AC-RSL-003 — 획득은 모든 기록 필드를 남긴다

**Given** 비어 있는 자원 `demo`.
**When** 세션 id `s-1`, 이름 `lane-a`, 명령 텍스트 `heavy-suite`, 상한 5분으로 획득한다.
**Then** 디스크 기록과 `--json` 출력 모두에 자원 이름, 세션 id, 세션 이름, 소유자 pid, pid 출처 표시, 명령, 획득 시각(RFC3339 UTC), 선언 상한, 만료 시각(= 획득 시각 + 상한)이 있다. 소유자 pid는 획득 CLI 프로세스 자신의 pid가 아니다.

- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_RecordsAllFields$' -count=1 -v` 그리고 `go test ./internal/cli/ -run 'TestSlotCLI_AcquireJSONFields$' -count=1 -v`
- 기대 출력: 두 테스트 각각 `--- PASS` 줄.
- **RED-now:** EL-1, EL-4 — 기록 타입도 CLI 명령도 없다.
- **Green path:** M2(기록), M3(CLI 출력).

### AC-RSL-004 — 살아 있는 다른 보유자가 있으면 거절하고 아무것도 쓰지 않는다

**Given** 자원 `demo`의 보유자가 살아 있는 프로세스(테스트 프로세스 자신)를 소유자 pid로 가진 세션 `s-1`이고 만료 전이다.
**When** 세션 `s-2`가 `--force` 없이 획득한다.
**Then** "보유 중" 오류가 반환되고(오류 판정 함수로 식별 가능), 오류 문구에 보유자 이름이 있으며, 기록 파일의 SHA-256이 획득 시도 전후로 같고, 감사 로그에 `refuse` 한 줄이 생긴다.

- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_RefusesLiveForeignHolderWritesNothing$' -count=1 -v`
- 기대 출력: `--- PASS` 줄.
- **RED-now:** EL-1.
- **Green path:** M2.

### AC-RSL-005 — 생존 판정: 스테일은 인수되고, 살아 있는 보유자와 pid 0은 거절한다

- **005a (스테일 인수):** **Given** 소유자 pid가 이미 종료된 더미 자식 프로세스의 pid인 기록(만료 전). **When** 다른 세션이 `--force` 없이 획득한다. **Then** 성공하고, 새 기록의 `displaced`에 이전 세션 id와 사유 `stale`이 있으며 감사 로그에 `takeover` 한 줄(사유 `stale`)이 있다.
- **005b (살아 있는 보유자, 짝):** **Given** 소유자 pid가 살아 있는 기록(만료 전). **When** 같은 획득. **Then** 거절된다.
- **005c (pid 0 보수적 판정):** **Given** pid 0과 pid 출처 표시가 있는 기록(만료 전). **When** 같은 획득. **Then** 거절되고, 상태 조회는 스테일이 아니라고 보고한다.
- 판정 명령: `go test ./internal/kanban/ -run 'TestSlotLease_Liveness$' -count=1 -v` (하위 테스트 `stale_takeover`, `live_refused`, `pid_zero_live`)
- 기대 출력: 상위와 하위 세 개의 `--- PASS` 줄.
- **RED-now:** EL-1.
- **뮤턴트 짝:** "항상 인수 가능" 뮤턴트는 005b·005c에서 실패하고, "pid를 보지 않음(항상 살아 있음)" 뮤턴트는 005a에서 실패한다.
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

**Given** 설정 키 `workflow.slot_lease.enabled`가 없는 설정, 자원 `demo`에 매칭 패턴 설정, 그리고 (a) 살아 있는 다른 보유자 기록 또는 (b) 손상된 기록.
**When** 매칭되는 Bash 명령으로 PreToolUse를 호출한다.
**Then** 두 경우 모두 허용이고, 표준 오류에 안내가 없으며, 감사 로그 파일이 생기지 않는다. (b)에서 안내가 나오면 꺼진 경로가 기록을 읽었다는 증거이므로 실패다.

- 판정 명령: `go test ./internal/hook/ -run 'TestSlotLeaseGuard_DisabledNeverReadsNorDenies$' -count=1 -v` 그리고 `go test ./internal/config/ -run 'TestDefaults_SlotLeaseDisabled$' -count=1 -v`
- 기대 출력: 두 `--- PASS` 줄.
- **RED-now:** EL-1.
- **뮤턴트 짝:** 설정 확인보다 기록 읽기를 먼저 하는 뮤턴트는 (b)에서 실패한다.
- **Green path:** M4.

### AC-RSL-011 — 가드 거부 복합 조건: 각 항이 제 몫의 실패 사례를 가진다

**Given** 설정 `enabled: true`, 자원 `demo`의 패턴 목록, 호출 세션 `s-2`.
**When** 표의 각 행 입력으로 PreToolUse를 호출한다.
**Then** 기준 행만 거부되고(사유가 `SLOT_LEASE_VIOLATION:`으로 시작하며 자원 이름과 보유자를 담는다), 나머지 행은 모두 허용된다.

| 행 | 설정 켜짐 | 패턴 일치 | 보유자 ≠ 호출자 | 보유자 생존 | 만료 전 | 기대 |
|---|---|---|---|---|---|---|
| base | 예 | 예 | 예 | 예 | 예 | 거부 |
| n-enabled | 아니오 | 예 | 예 | 예 | 예 | 허용 |
| n-match | 예 | 아니오 | 예 | 예 | 예 | 허용 |
| n-quoted | 예 | 따옴표 안에서만 | 예 | 예 | 예 | 허용 |
| n-self | 예 | 예 | 아니오(보유자 = 호출자) | 예 | 예 | 허용 |
| n-alive | 예 | 예 | 예 | 아니오(스테일) | 예 | 허용 |
| n-bound | 예 | 예 | 예 | 예 | 아니오(만료) | 허용 |
| n-held | 예 | 예 | 보유자 없음 | — | — | 허용 + 감사 `allow-unheld` |
| multi | 예 | 두 자원 중 하나만 거부 조건 | 예 | 예 | 예 | 거부(그 자원 이름) |

- 판정 명령: `go test ./internal/hook/ -run 'TestSlotLeaseGuard_DenyMatrix$' -count=1 -v`
- 기대 출력: 상위와 행마다 하위 `--- PASS` 줄(9개).
- **뮤턴트 표(각 항의 검사를 하나씩 지운 변형이 실패하는 행):** 설정 검사 삭제 → n-enabled, 패턴 검사 삭제 → n-match, 따옴표 제거 삭제 → n-quoted, 세션 비교 삭제 → n-self, 생존 검사 삭제 → n-alive, 만료 검사 삭제 → n-bound, "보유자 있음" 검사 삭제 → n-held, 자원별 판정을 첫 자원만 보도록 바꾸기 → multi. 모든 항이 자기만의 실패 행을 가지므로, 한 항을 빼먹은 구현은 이 표를 통과할 수 없다.
- **RED-now:** EL-1.
- **Green path:** M4.

### AC-RSL-012 — 가드 fail-open: 불확실하면 허용하고 흔적을 남긴다

**Given** 설정 `enabled: true`와 매칭 패턴.
**When** (a) 손상된 기록, (b) 호출 세션 id가 빈 입력, (c) 컴파일되지 않는 패턴(`(`), (d) 프로젝트 루트가 빈 입력으로 매칭 명령을 호출한다.
**Then** 네 경우 모두 허용이고, 각각 표준 오류에 `[moai:slot-lease] advisory:` 줄이 있으며, (d)를 뺀 세 경우는 감사 로그에 `fail-open` 한 줄이 사유와 함께 남는다((d)는 로그를 쓸 루트가 없으므로 표준 오류만 요구한다).

- 판정 명령: `go test ./internal/hook/ -run 'TestSlotLeaseGuard_FailOpen$' -count=1 -v`
- 기대 출력: 상위와 하위 네 개의 `--- PASS` 줄.
- **RED-now:** EL-1.
- **Green path:** M4.

### AC-RSL-013 — 통합 창과의 분리(회귀 가드)

**Given** M5 완료 트리.
**When** (a) 슬롯 획득을 해도 통합 기록 파일이 생기지 않고, 통합 획득을 해도 슬롯 임대 디렉터리가 생기지 않는지 확인한다. (b) 통합 창의 기존 테스트를 돌린다. (c) 통합 창 소스 파일이 바뀌지 않았는지 확인한다.
**Then** (a) 두 방향 모두 상대 파일이 없고, (b) 기존 테스트가 모두 통과하며, (c) 차이가 없다.

- 판정 명령:
  - `go test ./internal/kanban/ -run 'TestSlotLease_SeparateFromIntegrationWindow$' -count=1 -v`
  - `go test ./internal/kanban/ -run 'IntegrationLock' -count=1 -v`
  - `go test ./internal/hook/ -run 'IntegrationLock' -count=1 -v`
  - `git diff --stat "$BASELINE_SHA" HEAD -- internal/cli/integration.go internal/hook/integration_lock_guard.go` — `BASELINE_SHA`는 run 단계 첫 커밋 전에 `git merge-base develop HEAD`로 잡아 progress.md에 기록한 값이다.
- 기대 출력: 첫째 명령 `--- PASS`, 둘째·셋째 명령은 `--- FAIL` 없이 `ok`이며 스윕 수가 0이 아니고, 넷째 명령은 출력이 없다. 기반 함수를 매개변수화해야 해서 `internal/kanban/integration_lock_mutation.go`가 바뀌는 경우, 그 파일은 넷째 명령 범위에 넣지 않되 둘째 명령의 전체 통과로 동작 불변을 보인다.
- **RED-now:** 해당 없음(보존 불변식). 기준선: 통합 창 테스트는 현재 트리에서 존재한다(`internal/kanban/integration_lock_test.go`, `integration_lock_cross_test.go`).
- **뮤턴트 짝:** 슬롯 기록을 통합 기록 경로에 쓰는 변형은 (a)에서 실패한다.
- **Green path:** M2-M5 동안 유지.

### AC-RSL-014 — 템플릿 배포: 키는 꺼진 채로 싣고, 중립성과 유출 가드를 통과한다

**Given** M5 완료 트리와 `make build`로 다시 컴파일한 바이너리.
**When** 템플릿 설정과 새 규칙 파일을 검사하고 중립성·유출 테스트를 돌린다.
**Then** 템플릿 `workflow.yaml`에 `slot_lease` 블록이 있고 그 `enabled`가 `false`이며, 새 규칙 파일에 SPEC ID·카드 id·날짜 토큰이 없고, 중립성·유출(일반·strict)·규칙 출처 테스트가 모두 통과한다.

- 판정 명령:
  - `/usr/bin/grep -n -A2 'slot_lease:' internal/template/templates/.moai/config/sections/workflow.yaml` → `enabled: false` 줄 포함, 종료 코드 0
  - `/usr/bin/grep -rnE 'SPEC-[A-Z]|\bt[0-9]{3}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}' internal/template/templates/.claude/rules/moai/workflow/resource-slot-lease.md` → 출력 없음, 종료 코드 1
  - `go test ./internal/template/ -run 'TestTemplateNeutralityAudit$|TestTemplateNoInternalContentLeak$|TestRuleProvenance' -count=1 -v` → `--- FAIL` 없음, 이름 댄 테스트의 `--- PASS` 줄 존재
  - `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run 'TestTemplateNoInternalContentLeak$' -count=1 -v` → `--- PASS`
  - `make build` → 종료 코드 0
- **RED-now:** 템플릿 키 부재는 EL-1(스캔 범위에 템플릿 포함, 점 디렉터리 대조군 EL-3). 중립성·유출 테스트는 현재 초록이다(EL-6) — 이 부분은 회귀 가드다.
- **뮤턴트 짝:** 새 규칙 파일에 SPEC ID 한 줄을 넣은 임시 변형에서 유출 테스트가 실패함을 관측하고 되돌린다(출력은 progress.md §E.2). 이 관측이 없으면 새 파일이 유출 가드의 스캔 범위 안에 있다는 것을 알 수 없다.
- **Green path:** M5.

### AC-RSL-015 — 레인 문서 반영은 게이트를 따른다

**Given** run 단계 종료 시점에 `git merge-base --is-ancestor WT-acquire-branch-record develop`의 종료 코드와 두 ref의 SHA를 progress.md에 기록한 상태.
**When** 기록된 종료 코드로 가지를 고른다.
**Then**
- **게이트 닫힘(종료 코드 ≠ 0):** 이 SPEC의 커밋이 세 문서를 바꾸지 않았다. `git diff --name-only "$BASELINE_SHA" HEAD -- .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/local/gitflow-lane-protocol.md` → 출력 없음. progress.md에 후속 카드로 넘긴다는 기록이 있다.
- **게이트 열림(종료 코드 0):** 세 파일 각각에서 `/usr/bin/grep -c 'moai slot' <file>` → 1 이상. 템플릿판 `kanban-dispatch.md`에 대해 AC-RSL-014의 유출 테스트가 통과한다.
- **RED-now:** EL-5 — 현재 게이트는 닫혀 있다(종료 코드 1). 이 기준은 run 단계 종료 시점의 판정으로 가지를 정한다.
- **Green path:** M6.

### AC-RSL-016 — CLI 표면과 교차 플랫폼 빌드

**Given** M3 완료 트리.
**When** 두 플랫폼으로 빌드하고, 빈 자원에 `moai slot status --resource demo --json`을 실행한다.
**Then** 두 빌드가 성공하고, 상태 JSON이 자원 이름과 보유 여부 `false`를 담는다. `moai slot` 도움말이 세 동작(acquire, status, release)을 보여 준다.

- 판정 명령:
  - `go build ./...` → 종료 코드 0
  - `GOOS=windows GOARCH=amd64 go build ./...` → 종료 코드 0
  - `go test ./internal/cli/ -run 'TestSlotCLI_StatusJSONFree$|TestSlotCLI_HelpListsVerbs$' -count=1 -v` → 두 `--- PASS` 줄
- **RED-now:** EL-4 — `slot` 명령이 없다.
- **Green path:** M3.

## §D.3 추적성

| 요구사항 | 인수 기준 |
|---|---|
| REQ-RSL-001 | AC-RSL-013, AC-RSL-016 |
| REQ-RSL-002 | AC-RSL-003 |
| REQ-RSL-003 | AC-RSL-001, AC-RSL-002 |
| REQ-RSL-004 | AC-RSL-004 |
| REQ-RSL-005 | AC-RSL-005 |
| REQ-RSL-006 | AC-RSL-006 |
| REQ-RSL-007 | AC-RSL-007 |
| REQ-RSL-008 | AC-RSL-005 |
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
- **서브에이전트 세션 id(OQ-1).** 단위 테스트는 훅 입력을 직접 만들므로 실제 런타임이 서브에이전트에 어떤 `session_id`를 주는지 보지 못한다. M4에서 실제 훅 입력으로 잰 결과를 progress.md에 남긴다.
- **부하 비용.** 어떤 기준도 무거운 실행이 겹칠 때의 부하를 재지 않는다(범위 밖).

## §D.5 종료 게이트

- AC-RSL-001, 002의 출력(대조 갈래 `starts=2`, 표면 갈래 `acquired=1 refused=1`, 뮤턴트 `acquired=2`)이 progress.md §E.2에 원문으로 있다.
- 모든 MUST-PASS 기준의 판정 명령 출력이 스윕 수와 함께 기록돼 있다.
- `BASELINE_SHA`가 첫 run 커밋 전에 잡혀 기록돼 있다.
- M6 게이트 판정(명령, 종료 코드, 두 ref SHA)이 기록돼 있다.

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
- 가드를 이 저장소 로컬 설정에서 켤지는 운영자 결정이다.
