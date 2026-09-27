# Plan — SPEC-HOOK-STOP-PARSE-CAP-001

- 카드: t1272 (Class C). 브랜치 `WT-stop-parse-cap`, 기준 트리 `e464fd5d0`. 0.2.0 수리본의 판독 트리는 `6c0da65d5`(`e464fd5d0` 이후 문서 커밋 둘뿐, 코드 동일).
- Tier: **M**. 판단 근거 — 바뀌는 파일이 대략 7~9개(`internal/cli/hook_stdin_failclosed.go`, `internal/cli/hook.go`, 셈 기록 구현 파일 1개와 그 테스트, 상수를 두는 `internal/config/defaults.go`, `internal/codexadapter/stop_cap.go` 주석, 운영자 문서 템플릿 원본과 로컬 사본, 필요하면 `internal/session` 의 테스트 이음매)이고, 예상 변경량은 테스트 포함 400~700줄 수준이다. 설계 결정(열쇠, 만료, 파일 이름, 실패 방향)이 여럿이라 AC 를 별도 파일로 두는 편이 감사에 유리하다. 5개 산출물이 필요한 헌법급 변경은 아니다.

---

## §A 맥락

선행 SPEC-HOOK-STDIN-FAILCLOSED-001 은 Claude Stop 의 파싱 실패 루프를 호스트 상한이 끊는다고 가정했다. t1230 과 t1272 의 측정(spec.md §A.1)이 그 가정이 성립하지 않는 두 조건을 보였다 — 200 주입 세션, 그리고 도구 사용이 끼어드는 루프. 리드 판정 D1 과 수리 판정(spec.md §A.3)에 따라 moai 가 셈 열쇠(세션 id 우선, 없으면 세션 소유자 프로세스)별로 셈을 두고 N=8 을 넘으면 차단을 멈춘다. 같은 카드에서 사유 문구를 바꿔 모델이 훅·설정 파일을 편집하지 말고 사람에게 알리도록 한다.

현재 코드의 사실(이 트리에서 판독):

- `answerStdinParseFailure`(`internal/cli/hook_stdin_failclosed.go`)는 상태가 없다. 같은 입력이면 몇 번째 호출이든 같은 출력을 낸다. 이 트리에서 빌드된 바이너리로 파싱 실패 Stop 한 번을 실행한 결과는 acceptance.md §E 증거 원장 L1 에 있다.
- 사유 문구는 `stdinParseFailClosedReason()` 한 함수가 두 하네스에 공통으로 준다. REQ-SPC-009·010 을 함께 만족하려면 사유가 하네스별로 갈라져야 한다 — Codex 는 기존 문자열, Claude 는 새 문자열.
- 파싱 성공 경로는 `runHookEvent`(`internal/cli/hook.go`)의 `ReadInput` 뒤에 있다. REQ-SPC-005 의 초기화는 이 지점에서 Claude·Stop 일 때만 일어난다.
- `session.ResolveOwnerPID`(`internal/session/session_pid.go`)는 `MOAI_SESSION_PID` 스탬프(자기 조상을 가리킬 때만) → 래퍼 셸을 건너뛴 가장 가까운 조상 순으로 해석하고, 실패하면 `(0, false)` 를 돌려준다. 조상 표는 패키지 변수 `procInfo` 이음매로 읽으며, 기존 테스트(`session_pid_test.go`)가 합성 트리로 그 이음매를 바꿔 쓴다. 이 이음매는 비공개라 `internal/cli` 의 테스트에서는 직접 바꿀 수 없다(§C 6).
- 이 경로의 코드는 현재 세션 id 환경 변수도 세션 소유자 해석도 쓰지 않는다 — 원장 L5.

---

## §B 질문

### B.1 판정된 질문 (Kickoff 전 판정 완료)

- **만료 시간 — 판정됨(리드, 2026-09-26): 60분.** REQ-SPC-007. 근거: 파싱 실패 Stop 사이의 간격은 한 턴의 작업 시간이다. t1272 측정에서 23회 Stop 이 138초 안에 났지만, 긴 빌드·테스트를 도는 턴은 수십 분이 걸릴 수 있다. 만료가 턴 간격보다 짧으면 매 Stop 마다 셈이 1 로 돌아가 상한이 걸리지 않는다(fail-closed 쪽). 길수록 프로세스 열쇠의 PID 재사용 창이 넓어진다(spec.md §F 첫 행).
  - **상수 위치**: 값은 이름 붙은 상수 하나로 두고, 매직 넘버를 쓰지 않는다(CLAUDE.local.md §14 — 임계값은 `internal/config/defaults.go` 단일 원천). 같은 파일의 훅 관련 기본값(`DefaultHookDispatcherTimeout` 등) 옆에 둔다. N=8 도 같은 곳에 이름 붙은 상수로 둔다(REQ-SPC-004). 이름은 run-phase 가 정하되, 이름 예: `DefaultStopParseCapLimit = 8`, `DefaultStopParseCapExpiry = 60 * time.Minute`. 어느 쪽도 설정 키나 환경 변수로 노출하지 않는다.
- **셈 열쇠 — 판정됨(리드, 2026-09-26, D3 채택).** `CLAUDE_CODE_SESSION_ID` 가 있고 비어 있지 않으면 1순위, 없거나 비어 있으면 `session.ResolveOwnerPID`(D1). `/clear` 뒤 세션 id 가 바뀌면 셈을 이월하지 않는 것이 의도이며, 바뀌는지는 M5c 에서 잰다.

### B.2 판정이 필요 없는 설계 선택 — 근거와 함께 기록

- **열쇠를 정할 수 없을 때의 방향: 차단 유지(fail-closed).** REQ-SPC-008. 근거 셋. (1) 오늘의 동작과 같으므로 리드 판정 밖의 사용자 가시 변화가 없다. (2) 식별 실패를 해제로 바꾸면 「식별 정보가 없다」는 조건이 곧 거부를 통과로 바꾸는 길이 된다 — 선행 SPEC REQ-HSF-001 의 금지 절(거부를 통과로 바꾸는 스위치 금지)과 같은 위험이다. (3) 식별 실패는 드문 경로다: 세션 id 변수가 없고, 동시에 조상 탐색이 실패해야 한다(조상 표를 읽지 못하는 플랫폼, 또는 호스트가 이미 죽어 init 으로 재부모된 경우). 대가: 그 경우 루프는 호스트 상한만으로 묶인다(선행 SPEC 의 상태).
- **기록 파일 이름: 엄격한 검증이 아니라 해시.** REQ-SPC-015. 이름은 `sha256("session:" + 세션 id)` 또는 `sha256("process:" + 10진 pid)` 의 소문자 16진 64자에 `.json` 을 붙인 것이다 — 정규식 `^[0-9a-f]{64}\.json$`. 근거: (1) 세션 id 의 형식은 문서화된 계약이 아니다(관측값이 UUID 일 뿐). 문자 허용 목록으로 검증하면 형식이 바뀌는 날 모든 세션이 조용히 대체 경로로 떨어지거나, 허용 목록의 빈틈(`..` 조합, 유니코드 구분자, 길이)을 따로 막아야 한다. (2) 해시는 입력이 무엇이든 경로가 고정 길이·고정 문자 집합이 되므로 `/`·`\`·`..`·NUL·제어 문자가 경로에 닿을 수 없음이 구성으로 보장된다 — 정리 경로(REQ-SPC-007)가 지울 이름 형식도 이 정규식 하나로 정확해진다. (3) 종류 접두어가 세션 id `"123"` 과 pid 123 을 다른 기록으로 가른다. 대가: 파일 이름만 보고 어느 세션인지 알 수 없다 — 기록 안에 열쇠 종류만 남기고 세션 id 원문은 남기지 않는다(필요 없는 값을 저장하지 않는다).
- **링크와 비정상 항목: 따라가지 않고, 쓸 수 없으면 차단 유지.** 상태 영역이 링크이거나 자기 기록 자리가 일반 파일이 아니면 상한을 적용하지 않는다(REQ-SPC-008). 정리는 이름 형식에 맞는 일반 파일만(`Lstat` 판정) 지운다(REQ-SPC-007). 모델이 이 영역에 쓸 수 있으므로 삭제 경로가 상태 영역 밖으로 나가지 않게 하는 것이 목적이다.
- **셈 기록 손상 시 방향: 없는 것으로 보고 1 부터.** 역시 fail-closed 쪽이다. 손상된 기록을 큰 값으로 해석하는 선택지는 두지 않는다.
- **상한 해제 뒤 지속 해제.** REQ-SPC-003. 파싱할 수 없는 Stop 에서 턴 경계를 읽을 수 없으므로(`stop_hook_active` 판독 불가) 「턴마다 N 번」은 구현할 수 없다. D1 의 「연속 파싱 실패 Stop」을 문자 그대로 적용한 결과이며, 되돌리는 것은 파싱 성공과 만료뿐이다.
- **N=8 의 근거.** 리드 판정 값이며 호스트 기본 상한(8)과 같다. 호스트 기본 상한이 이미 걸리는 무도구 루프에서는 두 상한이 거의 같은 시점에 걸린다 — t1230·t1272 B팔에서 훅이 9번째 호출까지 차단했으므로, moai 의 9번째 상한 해제와 호스트의 9번째 종료가 겹친다(해석, 순서는 미측정).

---

## §C 착수 전 점검 (Pre-flight)

run-phase 착수 시 다음을 다시 재고 progress.md §E.2 에 남긴다.

1. `git rev-parse HEAD` 와 `git merge-base develop HEAD` — 기준 트리가 `6c0da65d5` 에서 움직였으면 아래 판독을 다시 한다.
2. `internal/cli/hook_stdin_failclosed.go` 의 `answerStdinParseFailure` 서명·분기 순서가 plan 시점과 같은지. 다르면 D-NEW-1 경로로 spec 을 먼저 고친다.
3. `internal/template/templates/.claude/settings.json.tmpl` 의 Stop 항목과 `internal/template/templates/.claude/hooks/moai/handle-stop.sh.tmpl` 이 여전히 셸 래퍼 + `exec moai hook stop` 모양인지 기록한다. 대체 경로는 래퍼 셸을 건너뛰므로 이것은 blocker 가 아니라 M5b 설계의 입력이다. 다만 래퍼가 셸이 아닌 프로그램으로 바뀌었으면 spec.md §F 「셸이 아닌 래퍼」 행이 현실이 되므로 보고한다.
4. 셈 기록 위치: `<프로젝트 루트>/.moai/state/stop-parse-cap/<64자 16진>.json`(§B.2 의 이름 규칙). 프로젝트 루트는 기존 영속 기록과 같은 `resolveHookProjectRoot()` 로 정한다. `.moai/state/` 는 런타임 관리 영역이며 저장소에 커밋되지 않는다 — plan 시점 확인: `git check-ignore -v .moai/state/stop-parse-cap/x.json` → `.gitignore:398:.moai/state/	.moai/state/stop-parse-cap/x.json`. 착수 시 다시 확인한다.
5. 열쇠 보강 선택지(요구 아님, 프로세스 열쇠에만 해당): `homestate.ProbeProcessIdentity(pid)` 가 돌려주는 프로세스 지문을 셈 기록에 함께 적고, 지문이 다르면 기록을 없는 것으로 본다. PID 재사용 시 해제 상태를 이어받는 위험(spec.md §F 첫 행)을 줄이지만 매 파싱 실패 Stop 마다 프로세스 조회 비용이 든다. 채택 여부는 run-phase 가 비용을 재고 정하며, 채택하면 AC-SPC-006 의 대조(해제 상태 이어받기)를 지문 일치 조건 아래로 옮기는 D-NEW-1 개정이 필요하다.
6. 합성 프로세스 트리 시험(AC-SPC-007 (iii))의 이음매: `session.ResolveOwnerPID` 의 조상 표 이음매(`procInfo`, `pidIsAlive`)는 비공개다. run-phase 는 **실제 해석의 조상 탐색이 합성 표 위에서 돌도록** 이음매를 연다 — 예: `internal/session` 에 조상 표를 인자로 받는 해석 변형을 두고 `ResolveOwnerPID` 가 그것을 부르게 하거나, 테스트 전용 공개 설정 함수를 둔다. 어느 쪽이든 `ResolveOwnerPID` 의 동작은 바꾸지 않으며(기존 `session_pid_test.go` 통과), 해석 결과를 상수로 돌려주는 가짜 해석기로 시험을 대신하지 않는다 — 그렇게 하면 래퍼 셸을 건너뛰지 않는 이음매 안 변이(AC-SPC-007 (iii) 필수 RED)를 잡지 못한다. 원시 `os.Getppid()` 변이는 주입된 표를 읽지 않으므로 이 시험으로는 잡히지 않고, AC-SPC-007 (iii) 의 두 파일 대상 `grep -c 'os.Getppid'` 정적 검사가 잡는다.

---

## §D 제약

- `internal/hook` 패키지를 수정하지 않는다(선행 SPEC 과 같은 원칙 — 이음매는 CLI 계층).
- `internal/session` 은 §C 6 의 테스트 이음매 외에는 바꾸지 않는다. 래퍼 이름 목록과 해석 순서는 그대로다.
- Codex 쪽 출력 바이트와 사유 문구를 바꾸지 않는다(REQ-SPC-009, 카드 t1233 범위).
- 테스트는 셈 기록 위치를 `t.TempDir()` 아래로 돌린다. 열쇠는 `CLAUDE_CODE_SESSION_ID` 를 `t.Setenv` 로 정하거나(그 테스트는 병렬로 돌리지 않는다) 대체 경로의 조상 표를 주입해 정한다. 실제 프로세스 트리는 테스트 프로세스들이 공유하므로 병렬 테스트가 같은 셈 기록을 건드리지 않게 한다. 러너 자신의 환경에 `CLAUDE_CODE_SESSION_ID` 가 있을 수 있으므로(이 저장소의 레인 세션이 그렇다) 대체 경로 시험은 그 변수를 명시적으로 비운다.
- 전체 스위트를 로컬에서 돌리지 않는다. 영향 패키지(`./internal/cli/...`, `./internal/codexadapter/...`, `./internal/session/...`, `./internal/config/...`)만 돌리고 전체 판정은 CI 에 맡긴다. `internal/cli` 전체 스위트는 무거운 실행이므로 `moai slot acquire` 로 자원 임대를 먼저 잡는다.
- 운영자 문서는 템플릿 원본을 먼저 고치고 `make build` 로 임베드한다(Template-First).

---

## §E 자기 검증

run-phase 완료 보고는 acceptance.md 의 AC 별로 명령·출력 원문·트리 SHA 를 싣는다. TDD 이므로 AC-SPC-002·004·006·007 의 RED 는 구현 전 실패 출력 원문으로 남긴다(현재 코드에는 셈이 없어 9번째 호출도 거부하고, 열쇠 해석이 없어 셈 기록이 생기지 않는다 — 그 실패가 RED 다).

---

## §F 마일스톤 (결정의 번복 가능성 순)

### M1 — 셈 열쇠·기록의 모양 (Priority High)

- 열쇠 선택(REQ-SPC-014): 세션 id 환경 변수 → `session.ResolveOwnerPID` → 없음. 열쇠 공급을 한 곳에 두고 테스트가 환경 변수와 조상 표를 주입할 수 있게 한다(§C 6).
- 기록 파일 이름(REQ-SPC-015, §B.2), 필드(연속 횟수, 마지막 갱신 시각, 열쇠 종류, 선택 시 프로세스 지문), 위치(§C 4), 쓰기 방식(임시 파일 + 이름 바꾸기로 반쯤 쓰인 내용 방지 — REQ-SPC-008), 만료 판정(60분 상수), 만료 정리(REQ-SPC-007 — 이름 형식에 맞는 일반 파일만, 링크 비추종), 링크·비정상 항목 처리(REQ-SPC-008).
- 상수 두 개(N, 만료)를 `internal/config/defaults.go` 에 둔다(§B.1).
- 구현 파일 이름은 `internal/cli/hook_stop_parse_cap.go` 로 한다 — AC-SPC-005 의 정적 검사가 이 이름을 가리킨다. 이름을 바꾸면 D-NEW-1 로 AC-SPC-005 를 함께 고친다.
- 이 마일스톤이 먼저인 이유: 열쇠와 만료는 사용자 가시 동작(언제 해제되고 언제 다시 차단되는가, 어떤 대화끼리 셈을 나누는가)을 정하며, 뒤집히면 모든 테스트가 바뀐다.

### M2 — 사유 문구 하네스 분리 (Priority High)

- Claude 하네스 사유를 REQ-SPC-010 의 고정 문자열로, Codex 하네스 사유는 기존 문자열 그대로 둔다.
- 기존 테스트의 기대 사유를 하네스별 리터럴 둘로 가른다(acceptance.md §D 의 목록 — `hook_stdin_failclosed_test.go`, `hook_protocol_fix_test.go:127`, `misc_coverage_test.go:382`). 공유 도우미 `expectedFailClosedReason()` 하나를 새 문구로 바꾸는 식의 수정은 Codex 기대값까지 함께 움직여 Codex 변경을 가리므로 하지 않는다.
- 사용자 가시 문구이므로 M3 보다 앞에 둔다.

### M3 — 상한 분기와 초기화 배선 (Priority High)

- `answerStdinParseFailure` 의 Claude·Stop 거부 분기 앞에 셈 증가와 N 판정을 둔다. N 이하는 기존 거부, 초과는 상한 해제 응답 + 전용 키 기록 + stderr 한 줄(열쇠 종류 포함).
- `runHookEvent` 의 파싱 성공 지점에서 Claude·Stop 일 때 셈 기록을 지운다(REQ-SPC-005 — 없으면 아무것도 만들지 않고, 삭제 실패는 stderr 한 줄). 다른 이벤트는 셈에 손대지 않는다(REQ-SPC-006).
- 판정에 호스트 상한 변수를 읽지 않는다(REQ-SPC-004) — AC-SPC-005 의 정적 검사가 잡는다.

### M4 — 주석·운영자 문서 정정 (Priority Medium)

- `internal/codexadapter/stop_cap.go` `HostLacksStopBlockCap` 주석의 「Claude Code is not covered here … has not been measured」 문장을 t1230·t1272 측정과 이 SPEC 의 moai 자체 상한으로 고친다(REQ-SPC-013).
- `internal/cli/hook_stdin_failclosed.go` 의 Stop 루프 `@MX:WARN`·`@MX:REASON` 을 「호스트 상한 + moai 자체 상한 N=8」로 고친다.
- 운영자 문서 템플릿 원본과 로컬 사본에 REQ-SPC-012 의 내용을 더한다(sync-phase 로 넘겨도 된다 — 그 경우 sync 의 manager-docs 가 소유).

### M5 — LIVE 측정 (M5a 는 Priority High·차단, M5b·M5c 는 Priority Low·운영자 예산)

세 측정 모두 **배포 래퍼 사슬 그대로** 돈다: 격리 `/tmp` 프로젝트에 이 워크트리에서 빌드한 바이너리로 설치한 `.claude/settings.json` Stop 항목과 `.claude/hooks/moai/handle-stop.sh` 를 쓰고(손으로 쓴 별도 래퍼 금지), `PATH` 맨 앞 디렉터리에 `moai` 대리 스크립트를 둔다. 래퍼의 `command -v moai` 가 대리 스크립트를 고르고, 대리 스크립트는 `exec` 로 워크트리 빌드 바이너리가 되므로 호스트 → 셸 → 래퍼 → moai 사슬이 배포본과 같다. 훅 stderr 는 프로젝트 안으로 돌린다(`MOAI_HOOK_STDERR_LOG=$CLAUDE_PROJECT_DIR/.moai/logs/hook-stderr.log` — 래퍼의 허용 목록 안). 증거는 `.moai/reports/t1272/` 아래(gitignored). 공통 격리: `--permission-mode bypassPermissions` 는 격리 `/tmp` 프로젝트에서만, 모델은 haiku.

대리 스크립트의 모양(`#!/bin/sh`): 첫 두 인자가 `hook stop` 이면 `CLAUDE_CODE_SESSION_ID` 의 존재 여부(`set-nonempty` / `set-empty` / `unset`)와 값, `$$`, `$PPID` 를 측정 로그 한 줄로 남긴 뒤 `exec <빌드 바이너리> "$@"` 한다. M5b 에서만 그 `exec` 의 stdin 을 파손 파일로 바꾼다. 다른 인자의 호출은 그대로 통과시킨다. 대리 스크립트는 `exec` 전에 같은 프로세스이므로 이 로그가 곧 `moai hook stop` 프로세스의 환경이다.

**착수 전 상한 선언** — 측정마다 실행 전에 progress.md §E.2 에 다시 적는다. 선언을 넘기면 멈추고 재판정을 받는다.

| 측정 | 실행 수 | 턴 상한 | 벽시계 상한 | 재실행 |
|---|---|---|---|---|
| M5a — 훅 환경의 세션 id (AC-SPC-016) | 1 | `--max-turns 2` | `timeout -k 10 120` | 없음 |
| M5b — 파손 stdin 루프 A(`CAP=200`)·B(unset) (AC-SPC-015 (i)(ii)) | 2 | 각 `--max-turns 30` | 각 `timeout -k 10 300` | 없음 |
| M5c — `/clear` 전후 세션 id (AC-SPC-015 (iii)) | 대화형 1 세션, 프롬프트 최대 3개(`/clear` 제외) | 프롬프트당 1턴 | 세션 전체 10분 | 없음 |

- **M5a** — 한 번의 복합 호출로 실행한다: `unset CLAUDE_CODE_SESSION_ID MOAI_SESSION_PID && timeout -k 10 120 claude -p "reply with ok" --max-turns 2 --model haiku --output-format json > <증거 경로>/m5a-result.json`. 레인 세션 안의 Bash 는 두 변수를 바깥 세션 값으로 갖고 있으므로(plan-audit 2차 N1 측정), 따로 부른 `unset` 은 다음 호출에 닿지 않는다 — 정리와 실행이 같은 호출이어야 한다. 유효 stdin 통과, Stop 이 한 번 이상 불린다. 판정 자료는 대리 스크립트 측정 로그의 상태·값과, 같은 실행의 결과 JSON `session_id` 다(AC-SPC-016). `set-nonempty` 값이 그 `session_id` 와 같아야 통과이고, 존재만으로는 통과가 아니다. 이 측정은 run-phase 완료를 막는다(리드 판정 D3 의 [HARD]).
- **M5b** — 기대: 두 팔 모두 파싱 실패 Stop 이 9번째에서 `{}` 를 받고 턴이 끝난다. 관측 항목(차단 아님): 새 사유 문구를 받은 모델의 훅·설정 편집 시도 횟수.
- **M5c** — `-p` 는 `/clear` 를 한 대화 안에서 재현하지 못하므로 운영자가 대화형으로 돈다: 프롬프트 1 → Stop 관측 → `/clear` → 프롬프트 2 → Stop 관측. 두 Stop 의 측정 로그에서 세션 id 가 바뀌었는지 기록한다. 기대(의도): 바뀐다 → 셈이 이월되지 않는다. 바뀌지 않으면 spec.md §F 「`/clear` 뒤의 이월」이 확인된 잔여 위험이 된다.
- M5b·M5c 는 모델 사용량과 운영자 시간에 달려 있으므로 막혀도 M1~M4 의 완료를 막지 않는다(Gap 보고). M5a 가 막히면 Gap 이 아니라 blocker 로 리드에게 보고한다.

---

## §G 금지 패턴

- 호스트 상한 값을 읽어 판정하는 구현(D2 부활).
- 페이로드의 `session_id` 로 열쇠를 삼는 구현 — 파싱 실패에서는 읽을 수 없다. (환경 변수 `CLAUDE_CODE_SESSION_ID` 는 REQ-SPC-014 의 1순위로 허용된다.)
- 원시 `os.Getppid()` 를 열쇠로 쓰는 구현 — 셸이 끼면 호출마다 열쇠가 바뀐다.
- 세션 id 원문을 파일 경로에 넣는 구현, 정리 경로에서 링크를 따라가거나 이름 형식 밖의 파일을 지우는 구현.
- 손상·판독 불가·열쇠 없음을 해제 쪽으로 해석하는 구현.
- 사유 문구에 복구 절차(`disableAllHooks`, moai 갱신 방법)를 넣는 것.
- Codex 사유 문구·번역 틀을 함께 바꾸는 것, 또는 두 하네스의 기대 사유를 공유 도우미 하나에서 만드는 테스트.

---

## §H 교차 참조

- `.moai/specs/SPEC-HOOK-STDIN-FAILCLOSED-001/spec.md` — REQ-HSF-001·009·010·011·012·013, §F.2 (0.4.3 정정본)
- `.moai/reports/t1272/verdict.md`, `.moai/reports/t1230/verdict.md`, `.moai/reports/t1272/plan-audit.md` — 로컬 증거
- `internal/session/session_pid.go` — 세션 소유자 해석(대체 경로)
- `internal/config/envkeys.go` `EnvClaudeCodeSessionID` — 세션 id 환경 변수의 문서화
- `internal/cli/launcher_blockcap_infinite.go` — 200 주입 조건
- `.claude/rules/moai/workflow/goal-directive.md` — 호스트 상한의 기존 서술
