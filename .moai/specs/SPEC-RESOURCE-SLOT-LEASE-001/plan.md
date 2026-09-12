---
id: SPEC-RESOURCE-SLOT-LEASE-001
title: "Implementation plan — resource slot lease (card t607)"
version: "0.2.0"
created: 2026-09-12
updated: 2026-09-12
author: manager-spec
tier: M
---

# Implementation Plan — SPEC-RESOURCE-SLOT-LEASE-001

Tier: **M**. 기준 트리: `c4ce42eca` @ `WT-heavy-test-slot`(worktree t607).

## §A 맥락과 Tier 판정

**Tier M.** 새 코드는 한 패키지의 임대 핵심, CLI 명령 하나, 가드 하나, 설정 키 하나, 템플릿 두 파일(설정 키 + 최소 규칙)이다. 테스트까지 합치면 약 12-15개 파일, 600-1000 LOC로 추정한다(추정이며 측정이 아니다).

- Tier S가 아닌 이유: (1) 대조군이 프로세스를 가로지르는 테스트다(자식 프로세스 두 개 + 구성된 끼어들기). (2) 소비 표면이 둘이다(CLI, PreToolUse 가드). (3) 플랫폼이 둘이다(Unix flock, Windows 원자적 생성 + 잔재 정리). (4) 템플릿 배포와 중립성 가드가 걸려 있다. (5) 게이트 걸린 문서 마일스톤이 있다.
- Tier L이 아닌 이유: 설계 공간은 두 선행 SPEC이 이미 탐색했고(통합 창의 소유자 앵커와 변경 락 직렬화), 이 SPEC은 그 방식을 새 기록에 적용한다. research.md가 필요할 만한 미지의 영역이 없다. 필요한 기준 측정은 spec.md §B와 acceptance.md 증거 원장에 있다.

마일스톤은 **되돌리기 어려운 결정부터** 둔다. 디스크에 남는 기록 스키마, 설정 키 모양, 가드 판정 의미가 먼저이고, 기계적인 배선·문서 작업이 뒤다.

## §B 다섯 가지 설계 질문에 대한 결정

### B1. 가드는 무엇을 매칭하는가 (결정: 자원별 설정 패턴 목록)

- 설정 `workflow.slot_lease.resources.<name>.commands`에 자원마다 **정규식(RE2) 목록**을 둔다. 가드는 따옴표 구간을 지운 명령 문자열(`substituteQuotedArguments` 재사용)에서 각 패턴을 검색한다. 복합 명령(`cd x && <heavy>`)도 검색으로 잡힌다.
- 고정 목록을 두지 않는 이유: 고정 목록은 곧 특정 언어의 명령 목록이 되어 템플릿 중립성과 충돌한다. 설정 패턴 목록은 사용자가 자기 생태계의 명령을 적는 방식이라 언어 중립이다.
- Bash 호출의 자원 귀속: 명령이 자원 R의 패턴 하나라도 만족하면 그 호출은 R에 귀속된다. 여러 자원에 귀속되면 자원마다 따로 판정하고, 하나라도 거부 조건을 만족하면 거부한다(사유는 첫 거부 자원을 이름으로 댄다).
- 보유자 자신의 명령 통과: 훅 입력의 `session_id`가 기록된 보유 세션 id와 같으면 허용한다. 서브에이전트 호출은 이 런타임에서 부모 세션 id로 온다(§E OQ-1, 닫힘). 그래서 한 세션 안의 병렬 무거운 실행은 직렬화되지 않는다(spec.md Exclusions).
- 보유자가 없는 자원에 매칭되면 허용하고 감사 로그에 `allow-unheld` 한 줄만 남긴다(표준 오류는 쓰지 않는다 — 불확실성이 아니라 정상 경로라서 매 호출 안내는 소음이다).
- 허용 경로마다 감사 사유를 따로 남긴다: `allow-self`(보유자 자신), `allow-stale`(보유자 사라짐), `allow-expired`(상한 경과), `allow-unheld`(보유자 없음). 허용 경로가 여럿이라, 한 경로의 검사를 지워도 다른 경로로 허용되는 뮤턴트를 사유 단언이 잡는다(AC-RSL-011). 패턴 불일치는 가드가 자원에 귀속하지 않은 호출이므로 감사 줄을 남기지 않는다.

### B2. 임대 상태는 어디에 두는가 (결정: primary 체크아웃 `.moai/state/slot-leases/`)

- 기록: `<primary>/.moai/state/slot-leases/<resource>.json` — 자원마다 파일 하나. 다른 자원끼리는 경합하지 않는다.
- 변경 락: 같은 디렉터리의 `<resource>.mutation.lock` — 자원마다 따로. 파일 이름 줄기는 통합 창(`integration-lock`, `integration-mutation.lock`)과 겹치지 않는다.
- 루트 해석(CLI): 통합 창과 같은 순서(`CLAUDE_PROJECT_DIR` → git common dir의 부모 → cwd, `internal/cli/integration.go`:46-65). 모든 워크트리의 CLI 호출이 한 기록을 본다.
- 루트 해석(가드): 훅의 프로젝트 루트 해석은 `CLAUDE_PROJECT_DIR` → `os.Getwd()` 두 단계뿐이다(`internal/hook/path_resolve.go`:78-94). 그대로 쓰면 훅 환경이 링크된 워크트리를 가리키는 세션에서 가드가 워크트리의 빈 디렉터리를 읽는다. 따라서 가드는 훅이 푼 루트에서 `git rev-parse --git-common-dir`의 부모를 다시 구해 CLI와 같은 primary 루트로 정규화한 뒤 기록을 읽는다(REQ-RSL-008). 구할 수 없으면 기록을 읽지 않고 안내와 함께 허용한다(fail-open, 감사 사유 `fail-open: root-unresolved`). 권장 구현은 CLI와 가드가 **같은 해석 함수 하나**를 부르는 것이다 — 두 경로가 따로 풀면 같은 어긋남이 다시 생긴다. 이 정규화는 hook 경로에서 git 하위 프로세스 하나를 추가하므로, 꺼진 경로와 패턴 불일치 경로에서는 실행하지 않는다(매칭된 호출에서만).
- 감사 로그: `<primary>/.moai/logs/slot-lease-audit.jsonl`. 한 줄 = 한 사건(acquire / refuse / takeover(stale|expired|force) / release / guard-deny / guard-allow-* / fail-open).
- 구현 위치(run 단계 결정, 권장안): `internal/kanban`에 통합 창 옆 파일로 둔다. 변경 락 기반(`acquireBoardLockImpl`, 대기 예산 `boardLockWaitBudget`, Windows 잔재 정리)이 이 패키지의 비공개 함수·상수라서, 다른 패키지에 두면 그 기반을 공개하거나 복제해야 한다. 생존 판정 `FactoryProcessAlive`는 공개 함수라(`internal/kanban/factory_alive_unix.go`:22, `factory_alive_windows.go`:28) 배치 결정의 근거가 아니다. **공유하는 것은 기반뿐이고, 창·기록·락 파일은 공유하지 않는다.**

### B3. 선언 상한의 의미 (결정: 상한이 지나면 만료 = 인수 가능)

- 획득 시 `--max-duration <dur>`(생략 시 설정 `workflow.slot_lease.default_max_duration`, 기본값은 `internal/config/defaults.go`에 한 번만 정의)로 상한을 받는다. 0 이하나 해석 불가 값은 거부한다.
- 상한이 지나면 소유자가 살아 있어도 **만료**로 읽는다. 다른 세션의 획득은 `--force` 없이 넘겨받고 사유 `expired`로 기록한다. 가드는 만료된 보유자에 대해 허용한다. 보유자 자신의 재획득은 상한을 새로 시작한다.
- 통합 창과 다르게 판단한 이유(비대칭의 크기가 다르다):
  - 통합 창에서 거짓 "비었음"의 비용은 한 트리에 두 병합이 동시에 들어가는 **데이터 손상**이다. 그래서 통합 창에는 시간 기반 해제가 없다.
  - 슬롯에서 거짓 "비었음"의 비용은 두 무거운 실행이 **겹치는 것**(느려짐·불안정성)이다. 손상이 아니다. 반면 거짓 "살아 있음"의 비용은 해제를 잊은 살아 있는 세션이 자원을 무기한 붙잡는 것이다. 소유자 pid가 세션 pid라서 이 경우가 흔하다.
  - 상한은 보유자 **자신의 선언**이다. 선언을 넘긴 보유를 인수하는 것은 남의 판단이 아니라 보유자의 약속을 집행하는 것이다.
- 비용: 선언 상한보다 오래 걸린 정상 실행은 넘겨질 수 있다. 보유자가 상한을 넉넉히 잡거나 재획득으로 연장하는 것이 대응이다.
- 기본값(권장): 30분. 이 값은 측정이 아니라 선택이다. 레인 운영에서 가장 무거운 패키지 테스트에 약 10분 타임아웃을 잡는 관행이 있다고 알려져 있으나, 이 트리의 규칙·문서·Makefile·CI 워크플로에서 그 값을 찾지 못했다(`/usr/bin/grep -rn -E 'timeout[= ]600s|-timeout 600s|-timeout=600s' .claude/rules .moai/docs Makefile .github/workflows` → 출력 없음). 따라서 "10분의 3배"라는 근거는 확인되지 않은 전제로만 둔다. §E OQ-3에서 조정할 수 있다.

### B4. 가드의 fail-open (결정: 모든 불확실성은 허용 + 감사 한 줄 + 표준 오류 안내)

- 허용하는 불확실성(`enabled`가 참으로 읽힌 뒤): 기록을 읽을 수 없음(손상 JSON), 호출 세션 id 없음, 프로젝트 루트 없음, 공유 루트로 정규화할 수 없음, 컴파일되지 않는 패턴, 명령 목록을 패턴 문자열 목록으로 해석할 수 없는 자원 항목. 각각 감사 로그에 `fail-open`과 사유를 남기고 표준 오류에 `[moai:slot-lease] advisory: ...`를 쓴다.
- 설정 자체를 읽을 수 없으면 `enabled`를 알 수 없으므로 꺼짐으로 읽고 조용히 넘어간다(REQ-RSL-012, 통합 가드 `integrationLockEnabled`와 같은 관례). 이 경로는 fail-open 안내를 내지 않는다 — 설정 없는 프로젝트에서 매 Bash 호출마다 안내가 나오는 해석을 막으려는 것이다.
- 해석할 수 없는 자원 항목이 켜진 경로에서 실제로 도달 가능하려면, 자원 구획 하나가 잘못됐다고 워크플로 설정 전체의 로드가 실패해서는 안 된다. 따라서 자원 구획은 관대하게 해석한다(잘못된 항목은 그 항목만 무효로 표시하고 `enabled`는 그대로 읽는다). 로더가 파일 전체를 거부하는 방식을 택하면 이 경우는 "설정 읽기 실패 = 꺼짐"으로 떨어져 조용히 사라지므로, 그 방식은 택하지 않는다.
- CLI와의 비대칭은 통합 창과 같다. CLI는 "들어가도 되나"를 묻는 호출이므로 읽을 수 없는 기록을 **오류**로 돌려준다(빈 자원으로 읽지 않는다). 가드는 뜨거운 도구 경로에 있으므로 같은 사실에 **허용**으로 답한다.
- 꺼진 경로(`enabled` 거짓/미설정)에서는 기록도 설정 패턴도 읽지 않는다. 감사 로그도 쓰지 않는다.

### B5. Tier와 마일스톤 (결정: Tier M, 여섯 마일스톤, M6은 게이트)

아래 §F를 본다.

## §C 사전 점검 (run 단계 시작 전)

```bash
git rev-parse --short HEAD
git branch --show-current
/usr/bin/grep -rlE 'SlotLease|slot_lease|slot-lease' internal cmd pkg          # 여전히 없어야 한다 (EL-1)
/usr/bin/grep -rlE 'IntegrationLock' internal/kanban/integration_lock.go      # 대조군 (EL-2)
git merge-base --is-ancestor WT-acquire-branch-record develop                  # M6 게이트 (EL-5)
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
/usr/bin/grep -rn 'rules/moai/workflow' internal/template --include='*_test.go' -l   # 규칙 인벤토리 테스트 존재 확인
```

마지막 줄은 새 템플릿 규칙 파일이 걸릴 인벤토리·예산 테스트를 찾기 위한 것이다. 걸리는 테스트가 있으면 M5에서 그 테스트의 기대 집합을 함께 갱신한다.

## §D 구현 제약

- 통합 창 파일(`integration_lock*.go`, `internal/cli/integration.go`, `internal/hook/integration_lock_guard.go`)의 동작을 바꾸지 않는다. 공유 기반 함수를 자원별 경로로 쓰려고 매개변수화하는 것은 허용하되, 통합 창의 기존 테스트가 그대로 통과해야 한다(AC-RSL-013).
- 기본값·임계값은 `internal/config/defaults.go`에 한 번만 둔다. 환경변수 이름이 필요하면 `internal/config/envkeys.go`에 상수로 둔다.
- 테스트는 `t.TempDir()` 아래에서만 돌고, 자식 프로세스는 `t.Cleanup`에 등록한 kill과 외부 데드라인으로 묶는다. 끝에 붙인 `kill`은 정리가 아니다.
- 테스트 환경 고정·비움 목록: `CLAUDE_PROJECT_DIR`, `GIT_CEILING_DIRECTORIES`, `CLAUDE_CODE_SESSION_ID`, `MOAI_SESSION_PID`. 레인 환경에는 뒤의 둘이 설정돼 있고 세션 id·소유자 pid 해석이 이 둘을 먼저 읽으므로, 비우지 않으면 테스트가 실제 세션 id와 pid를 기록에 박는다.
- 템플릿 파일에는 SPEC ID, 날짜, SHA, 카드 id, 특정 언어 명령을 넣지 않는다.
- `go test ./...`를 로컬에서 돌리지 않는다. 영향받는 패키지만 `-run` 필터로 돌리고 스윕 수를 확인한다.

## §E 열린 질문

- **OQ-1 (닫힘) — 서브에이전트 Bash 호출의 `session_id`.** plan-audit 1회차가 쟀다(`.moai/reports/t607/plan-audit-iter1.md` Evidence E3). 명령: 감사자 서브에이전트 Bash 호출 `date -u +%FT%T; echo t607-oq1-probe` → `2026-09-11T15:31:44`. 직후 워크트리 trace 로그의 PreToolUse 줄은 `"session_id":"432c40c9-1b0f-40c5-9e9e-3b22db04b9b2"`였고, 서브에이전트 Bash 환경의 `CLAUDE_CODE_SESSION_ID`도 같은 값이었다. 즉 서브에이전트는 부모 세션으로 읽힌다. E3에는 런타임 버전이 없다. 이 개정 작성 시점의 `claude --version` → `2.1.268 (Claude Code)`(같은 런타임이라는 증거는 아니다). 결론: 보유자 서브에이전트가 거부되는 경로는 이 런타임에서 없고, 대신 **한 세션 안의 병렬 무거운 실행은 직렬화하지 않는다**(spec.md Exclusions). M4에서는 재측정을 요구하지 않되, 런타임 버전이 바뀐 뒤 서브에이전트 거부가 관측되면 이 결론을 다시 연다.
- **OQ-2 (범위 질문) — 보유 강제 모드.** 보유자가 없을 때 매칭 명령을 거부하는 모드는 범위 밖으로 두었다(spec.md Exclusions). 감사 로그의 `allow-unheld` 빈도가 그 필요성을 재는 근거가 된다.
- **OQ-3 (값 질문) — 기본 상한 30분.** B3의 근거는 선택이며 측정이 아니다. 운영자가 다른 값을 원하면 `defaults.go` 한 곳만 바뀐다.
- **OQ-4 (재론 가능한 결정) — 만료 시 무강제 인수.** B3은 이 SPEC이 내린 결정이다. 감사자나 운영자가 통합 창과 같은 "시간 기반 해제 없음"을 원하면 REQ-RSL-006과 AC-RSL-006을 바꿔야 한다.

## §F 마일스톤 (되돌리기 어려운 결정 먼저)

### M1 — 결정 고정과 RED: 대조군과 판정 테스트 (Priority: High)

- 기록 스키마(필드 이름·타입·만료 시각 도출), 상태 경로, 설정 키 모양(`enabled`, `default_max_duration`, `resources.<name>.commands`), 가드 판정 표를 먼저 코드 주석이 아닌 테스트로 고정한다.
- 대조군 테스트(AC-RSL-001)를 쓴다. 한 테스트 함수 안에 두 갈래:
  - 대조 갈래 — 표면 없이 확인-후-시작: 자식 A는 확인 직후 테스트 전용 끼어들기 지점에서 멈추고 STALLED 표지를 남긴다. 부모는 A의 STALLED를 본 뒤 B를 띄우고, B의 완료 또는 풀림 타임아웃 중 **먼저 오는 쪽**에 A를 푼다. 여기서는 락이 없어 B가 곧 끝나므로 A는 B 완료 뒤 풀리고, 두 자식 모두 시작한다(시작 수 = 2). 이 갈래는 하네스가 이중 시작을 **관측할 수 있음**을 보이는 양성 대조군이며 표면 전후 모두 통과해야 한다.
  - 표면 갈래 — 같은 끼어들기를 획득의 판정과 쓰기 사이, 곧 변경 락 임계 구역 안에 둔다. 올바른 직렬화 아래에서 B는 변경 락을 기다리느라 끝날 수 없으므로 "B 완료만 기다리기"는 교착한다. 그래서 A는 **B의 완료 또는 풀림 타임아웃 중 먼저 오는 쪽**에 풀린다. 풀림 타임아웃은 변경 락 대기 예산(`boardLockWaitBudget`)의 **1/3 이하**로 둔다 — 코드 방식 출처는 `internal/kanban/integration_lock_cross_test.go`:46-63(500ms, 예산 1.65s의 30.3%)이다. 그러면 A가 풀린 뒤에도 B에게 예산의 2/3 이상이 남아 B는 A가 쓴 기록을 읽고 "보유 중"으로 거절된다. 결과가 busy면 하네스 설정 결함으로 실패시킨다(판정이 아니라 구성 오류다).
  - 두 자식이 기록하는 소유자 pid는 **부모 테스트 프로세스의 pid**로 고정한다(출처 `integration_lock_cross_test.go`:180-183). 자식이 자기 pid를 기록하면 쓰고 곧 종료해 기록이 스테일이 되고, 다른 자식이 정당하게 인수해 이중 성공이 난다 — 락이 옳아도 그렇다.
- 이 단계에서 표면 갈래는 표면이 없으므로 실패(또는 컴파일 불가)해야 하고, 그 실패 출력이 RED 증거다.
- 판정 테스트(스테일/만료/강제/pid 0/해제/이름 검증)도 이 단계에서 RED로 둔다.

### M2 — 임대 핵심 (Priority: High)

- 획득·상태·해제 핵심과 자원별 변경 락 임계 구역. 읽기는 임계 구역 **안에서** 한다(대기한 호출자는 앞선 변경이 발행한 상태로 판정해야 한다).
- 스테일(pid 사라짐) / 만료(상한 경과) / 강제 인수 시 새 기록에 `displaced`(이전 보유자, 사유, 시각)를 담고 감사 로그에 한 줄 남긴다.
- pid 0 기록은 살아 있음으로 읽되 상한 만료는 적용된다.
- 기록 쓰기는 고유 임시 파일 + rename. 자원 이름은 허용 문자 집합(권장: 소문자 영숫자와 `-`, 1-64자)으로 검증한 뒤에만 경로를 만든다.
- Windows 기반의 잔재 정리를 자원별 락 경로에 적용한다.
- 변경 락이 대기 예산 내내 경합하면 held와 구별되는 busy 오류를 돌려주고 기록을 바꾸지 않는다. busy 판정 함수와 held 판정 함수는 서로의 오류에 대해 거짓이다(통합 창의 `ErrIntegrationLockBusy` / `ErrIntegrationLockHeld` 구별과 같은 방식, `integration_lock_mutation.go`:43-58).
- `--name`·`--command`를 생략하면 빈 값으로 기록한다.
- AC-RSL-001(표면 갈래 + 001b 뮤턴트), 002, 003a/b, 004, 005, 006, 007, 008, 009가 여기서 초록이 된다.

### M3 — CLI `moai slot` (Priority: High)

- `moai slot acquire --resource <name> [--name <lane>] [--command <text>] [--max-duration <dur>] [--force] [--session <id>] [--json]`, `moai slot status [--resource <name>] [--json]`, `moai slot release --resource <name> [--force] [--session <id>] [--json]`.
- 세션 id 해석과 소유자 pid 해석은 통합 창의 해석 함수를 재사용한다. 세션 id가 없으면 오류로 보고한다(지어내지 않는다).
- 획득 출력은 밀어낸 보유자를 항상 보고한다. 거절은 0이 아닌 종료 코드와 보유자 정보를 낸다. busy는 held와 다른 문구와 종료 코드로 낸다.
- 상태 조회는 빈 세션 이름·명령을 `(not given)`으로 보여 준다.
- AC-RSL-003c(교차 플랫폼 빌드 + status JSON + 도움말)가 여기서 초록이 된다.

### M4 — 설정 키와 PreToolUse 가드 (Priority: High)

- `internal/config/types.go`에 `workflow.slot_lease` 구조 선언, `defaults.go`에 `Enabled: false`와 기본 상한.
- 가드: 꺼진 경로는 즉시 반환(기록·패턴 미조회). 켜진 경로는 §B1·§B4의 판정표대로. 거부 접두어 `SLOT_LEASE_VIOLATION:`. `pre_tool.go`에서 통합 가드 바로 뒤에 배선.
- 가드 루트 정규화(§B2): 매칭된 호출에서 훅 루트를 git common dir의 부모로 정규화한 뒤 기록을 읽는다. 정규화 실패는 fail-open.
- AC-RSL-010, 011, 012, 016이 여기서 초록이 된다.

### M5 — 템플릿 배포와 로컬 사본 (Priority: Medium)

- 템플릿 `internal/template/templates/.moai/config/sections/workflow.yaml`에 `slot_lease` 블록 추가: `enabled: false`, `default_max_duration`, `resources: {}` 그리고 언어를 가리키지 않는 자리표시자 예시 주석. 자리표시자는 `<resource-name>`과 `<command-regex>` 두 토큰으로만 쓴다 — AC-RSL-014가 이 구조와 언어 도구 토큰 부재를 기계로 확인한다.
- 새 최소 규칙 `internal/template/templates/.claude/rules/moai/workflow/resource-slot-lease.md`: `paths:`로 범위를 좁혀(설정 파일과 규칙 자신) 항상 로드되지 않게 한다. 내용은 세 동작, 선언 상한 의미, 가드 opt-in 방법, fail-open 성질뿐이다. SPEC ID·날짜·SHA·카드 id·특정 언어 명령 금지.
- `make build`로 임베드를 다시 컴파일한다. 로컬 `.claude/rules/moai/workflow/resource-slot-lease.md` 사본을 템플릿과 맞춘다. 로컬 `.moai/config/sections/workflow.yaml`에서 켤지는 운영자 결정이며 이 마일스톤은 켜지 않는다.
- AC-RSL-013(분리 회귀), AC-RSL-014(템플릿 중립성·유출)가 여기서 확인된다.

### M6 — 게이트: 레인 문서 반영 (Priority: Low, 진입 조건 필수)

- **진입 조건**: `git merge-base --is-ancestor WT-acquire-branch-record develop`이 종료 코드 0. 판정 명령과 결과(종료 코드, 두 ref의 SHA)를 progress.md에 먼저 기록한다.
- 진입 조건이 참이면: 흡수(로컬 develop) 후 로컬·템플릿 `kanban-dispatch.md`와 로컬 `gitflow-lane-protocol.md`에 슬롯 임대 사용법을 짧게 추가한다(병합 창과 별개임을 명시). 템플릿판은 중립성 규칙을 따른다.
- 진입 조건이 거짓이면: 세 파일을 건드리지 않고, 문서 반영을 후속 카드로 넘긴다는 사실과 판정 증거를 progress.md에 기록한다. SPEC은 M1-M5로 닫을 수 있다(AC-RSL-015의 "게이트 닫힘" 가지).
- "세 파일을 건드리지 않았다"는 판정은 `.claude/rules/local/gitflow-lane-protocol.md` §8 형태로만 잰다: 읽는 시점에 `CARD_BASE=$(git merge-base develop HEAD)`를 다시 구하고(값을 핀하지 않는다), 대조군 `git diff --name-only "$CARD_BASE"..HEAD | wc -l`이 1 이상일 때만 판정한다(0이면 "측정 불가"). 이 판정은 병합 전에만 유효하다. 병합 뒤의 근거는 병합 트리와 카드 브랜치 트리의 동일성이다.

## §G 명명된 반패턴

- **확인-후-행동을 락 대신 쓰기.** 기록 존재 여부를 먼저 보고 나서 쓰는 모든 경로는 이 SPEC이 닫으려는 결함이다. 읽기는 임계 구역 안에 있어야 한다.
- **통합 창 재사용.** `moai integration`에 `--resource`를 붙이거나 통합 기록에 필드를 더하는 것. 기각된 C안이다.
- **훅 루트를 그대로 믿기.** 훅이 푼 루트에서 바로 기록을 읽는 것. 링크된 워크트리 세션에서 가드가 조용히 꺼진다(REQ-RSL-008).
- **리터럴 기준 SHA로 범위 판정.** "이 카드가 무엇을 바꿨나"를 run 시작 때 핀한 SHA로 재는 것. develop을 흡수하는 순간 남의 커밋이 범위에 들어온다(lane 규칙 §8).
- **busy를 held로 보고.** 경합을 "다른 세션이 보유 중"으로 알리면 레인이 판을 잘못 읽는다(REQ-RSL-004).
- **특정 언어 명령을 기본 패턴으로 싣기.** 템플릿 중립성 위반이다.
- **pid 대신 CLI 프로세스 pid 기록.** 통합 창이 이미 한 번 겪은 결함이다. 기록 즉시 스테일로 읽힌다.
- **강제 인수를 표준 출력에만 남기기.** 기록과 감사 로그에 남겨야 한다(REQ-RSL-007).
- **불확실성에서 거부하기.** 가드가 손상된 JSON 하나로 모든 무거운 명령을 막으면 보호하려던 작업 전체를 멈춘다.
- **운 좋은 경합을 기다리는 대조군.** 끼어들기는 테스트 전용 지점으로 구성한다. 기다려서 얻는 경합은 멈출 기준이 없다.
- **게이트 전 문서 편집.** M6 진입 조건 전에 세 문서를 고치는 것.

## §H 교차 참조

- `.moai/reports/t607/verdict.md` — 운영자 B안 결정과 사전 판단 공백의 정본 기록
- `.moai/specs/SPEC-INTEGRATION-LOCK-LIVENESS-001/` — 세션 소유자 pid 앵커(코드 방식 출처)
- `.moai/specs/SPEC-INTEGRATION-LOCK-ATOMIC-001/` — 변경 락 직렬화와 구성된 끼어들기 교차 프로세스 테스트(코드 방식 출처)
- `internal/kanban/integration_lock.go`, `integration_lock_mutation.go`, `integration_lock_cross_test.go`
- `internal/cli/integration.go`, `internal/hook/integration_lock_guard.go`, `internal/hook/pre_tool.go`
- `internal/config/types.go`, `internal/config/defaults.go`
- `.github/workflows/template-neutrality-check.yaml`, `internal/template/internal_content_leak_test.go`, `internal/template/template_neutrality_audit_test.go`
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — opt-in + fail-open 가드 원칙
