---
id: SPEC-HOOK-STOP-PARSE-CAP-001
title: "Claude Stop 파싱 실패 차단의 moai 자체 상한 — 호스트 상한에 기대지 않는 루프 한계와 사유 문구 개정"
version: "0.2.1"
status: completed
created: 2026-09-26
updated: 2026-09-26
author: manager-spec (card t1272)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/codexadapter"
lifecycle: spec-anchored
tags: "hook, fail-closed, stdin, stop, loop-cap, claude"
tier: M
related_specs:
  - SPEC-HOOK-STDIN-FAILCLOSED-001
  - SPEC-INFINITE-GOAL-001
  - SPEC-FACTORY-MODE-001
---

# SPEC-HOOK-STOP-PARSE-CAP-001 — Claude Stop 파싱 실패 차단의 moai 자체 상한

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-26 | manager-spec (card t1272) | 최초 plan-phase 초안. 근거는 t1230 판정서와 이 카드의 LIVE 재현(`.moai/reports/t1272/verdict.md`, 로컬 증거)이고, 설계는 리드 판정 D1(2026-09-26)과 사유 문구 개정 지시를 따른다. 기준 트리 `e464fd5d0`. SPEC-HOOK-STDIN-FAILCLOSED-001 의 낡은 측정 서술(F6)은 같은 카드에서 그 SPEC 0.4.3 으로 정정했고, 그 SPEC 의 REQ-HSF-010 문구는 Claude 하네스에 한해 이 SPEC 의 REQ-SPC-010 이 대체한다. |
| 0.2.0 | 2026-09-26 | manager-spec (card t1272) | plan-audit 1차(FAIL 0.79, `.moai/reports/t1272/plan-audit.md`) 수리. 리드 판정(2026-09-26): G1 — 만료 60분, 이름 붙은 상수 하나. D3 — 셈 열쇠를 `CLAUDE_CODE_SESSION_ID`(비어 있지 않을 때) 우선, 없으면 정본 세션 소유자 해석으로 바꾸고(REQ-SPC-014 신설), 훅 환경의 그 변수 존재를 실경로에서 재는 AC-SPC-016 을 더했다. D1 — 대체 경로를 원시 부모 pid 가 아닌 래퍼 셸을 건너뛰는 해석으로 고치고, 합성 프로세스 트리 차단 AC(AC-SPC-007 (iii))와 배포 래퍼 사슬 경유 M5 로 바꿨으며, §F 에 Windows 기대 동작을 적었다. D2 — 기록 파일 이름을 열쇠의 일방향 해시로 정하고(REQ-SPC-015 신설) 정리 대상을 그 이름 형식의 일반 파일로 좁혔다. D4·D5·D7·D9·D10 반영, D8 은 D3 으로 해소. REQ 13→15, AC 15→16(옛 AC-SPC-007·008 의 범위를 열쇠 해석과 기록 사용 불가로 재편). |
| 0.2.1 | 2026-09-26 | manager-spec (card t1272) | plan-audit 2차(FAIL 0.73, `.moai/reports/t1272/plan-audit-iter2.md`) 차단 결함 N1~N4 만 수리했다. 설계(D1·D3·G1)는 바꾸지 않았다. N1 — M5a 실행을 `unset CLAUDE_CODE_SESSION_ID MOAI_SESSION_PID && …` 한 번의 복합 호출로 바꾸고, AC-SPC-016 이 로그 값과 같은 실행의 `session_id` 가 같은지를 판정하게 했다. N2 — AC-SPC-016 (b)(유효 stdin 경로에서 관측할 출력이 명세에 없는 열쇠 종류 stderr)를 없앴다. 열쇠 종류 매핑은 AC-SPC-007 (i)(ii) 가 고정한다. N3 — AC-SPC-007 (iii) 의 필수 RED 변이를 「주입된 이음매 안에서 래퍼 셸을 건너뛰지 않는 구현」으로 다시 정의하고, `os.Getppid` 는 두 구현 파일 대상 `grep -c` 정적 검사로 따로 잡는다. N4 — REQ-SPC-004 의 「차단 쪽으로만 기운다」를 양방향 서술로 고치고 §F 에 「열쇠 공유」 행을 더했다. REQ·AC 개수 변화 없음. |

---

## §A 배경

### A.1 무엇이 측정됐는가

SPEC-HOOK-STDIN-FAILCLOSED-001(이하 「선행 SPEC」)은 훅 stdin 을 파싱하지 못한 결정 이벤트를 fail-closed 로 거부한다. Claude 하네스의 Stop 도 거부 대상이며, 반복 차단의 한계는 호스트의 Stop 연속 차단 상한(`CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`, 기본 8)이 정한다고 **가정**했다. 두 카드가 그 가정을 쟀다.

| 측정 | 조건 | 관측 | 출처 |
|---|---|---|---|
| t1230 2차 | Claude Code 2.1.283, `claude -p --model haiku`, 모조 Stop 훅(JSON block·exit 2 두 팔), 상한 변수 unset | 두 팔 모두 훅이 9번 실행돼 9번 차단한 뒤 턴이 끝났다(`num_turns 10`) | `.moai/reports/t1230/verdict.md` |
| t1230 1차 | 같은 조건에 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200` 누수 | 20회 모두 차단, `--max-turns 20` 에서 `error_max_turns` — 「20회」는 기록 원본이 없는 추론값이다 | 같은 판정서 F1 |
| t1272 A팔 | 실제 `moai hook stop` 경로에 파손 stdin, `CAP=200` 명시, `--max-turns 30` | Stop 23회, 도구 사용 없이 이어진 마지막 차단 17회 이상에서 끊기지 않고 `error_max_turns`(num_turns 31) | `.moai/reports/t1272/verdict.md` |
| t1272 B팔 | 같은 경로, 상한 변수 unset | Stop 16회, 도구 사용 없이 이어진 마지막 차단 9회 뒤 `success`(num_turns 26) | 같은 판정서 |

증거 파일은 모두 `.moai/reports/` 아래의 로컬 증거이며 저장소에 커밋되지 않는다.

### A.2 세 가지 발견

1. **200 이 주입된 세션에서는 호스트가 루프를 끊지 않는다.** 런처는 칸반·팩토리 세션과, 실행 시점에 무한(`MaxTurns == 0`) goal 이 걸린 세션에 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200` 을 무조건 주입한다(`internal/cli/launcher_blockcap_infinite.go` `DefaultRaisedStopHookBlockCap`). 그 세션에서 파싱 실패가 지속되면 Stop 차단이 최대 200회까지 이어질 수 있다. t1272 A팔은 30턴 상한에서 잘렸으므로 200 까지 가는지는 재지 않았다.
2. **기본 상한조차 「연속 차단 8회」가 아니라 「도구 사용 없이 이어진 차단」을 세는 것으로 보인다.** B팔에서 Stop 은 16회 불렸고, 도구 사용이 끼어들 때마다 셈이 처음부터 다시 시작된 모양이다. 따라서 모델이 도구를 계속 쓰는 한 기본 상한에서도 루프 길이에 위쪽 한계가 없다. 이것은 두 실행의 순서 기록에서 읽은 **해석**이며 업스트림 구현이나 changelog 로 확인하지 않았다.
3. **fail-closed 사유를 받은 모델은 훅을 끄려고 한다.** 두 팔 모두 모델이 `.claude/settings.json` 과 훅 스크립트를 읽은 뒤 편집·삭제를 시도했고, `-p` 기본 권한 모드여서 거부됐다. `bypassPermissions` 로 도는 레인에서는 그 편집이 실제로 이뤄질 수 있다(재현하지 않았다 — 권한 모드에서 추론한 위험). 선행 SPEC 이 없앤 실행 시점 스위치(REQ-HSF-001)를 모델이 설정 편집으로 되살리는 모양이다.

### A.3 리드 판정 (2026-09-26, 구속력 있음)

- **D1**: moai 가 연속된 stdin 파싱 실패 Stop 을 스스로 센다. 열쇠는 부모 프로세스 id(호스트 프로세스)와 그에 대응하는 상태 기록이다. 셈이 N=8 을 넘으면 무의견 기본 출력으로 답해 차단을 멈추고 그 사실을 기록한다. 요건: 도구 사용이 셈을 초기화하지 않는다, 파싱에 성공한 Stop 은 초기화한다, 상태 기록은 호스트 프로세스별이며 만료와 정리 경로를 갖는다, 호스트 상한 값에 의존하지 않는다.
- **사유 문구 개정**: fail-closed 사유는 모델에게 훅·설정 파일을 편집하지 말고 사람에게 알리라고 말해야 한다. 선행 SPEC REQ-HSF-010 의 고정 문구를 이 SPEC 이 대체한다. **Claude 쪽만** 바꾸며, Codex 번역 틀은 건드리지 않는다(카드 t1233 의 범위).
- **수리 판정 (2026-09-26, 구속력 있음)**: (G1) 만료 시간은 60분이며 이름 붙은 상수 하나로 둔다. (D3) 셈 열쇠는 `CLAUDE_CODE_SESSION_ID` 가 있고 비어 있지 않으면 그 값, 없거나 비어 있으면 정본 세션 소유자 해석(`session.ResolveOwnerPID`)의 결과다. `/clear` 뒤 세션 id 가 바뀌면 셈은 이월되지 않는 것이 의도이며, 바뀌는지는 측정 항목이다. (D1) 대체 경로에 원시 부모 pid 를 쓰지 않는다. (D2) 정리는 기록 파일 이름 형식에 정확히 맞는 일반 파일만 지운다. 위 D1 문단의 「부모 프로세스 id(호스트 프로세스)」와 「호스트 프로세스별」은 이 판정에 따라 「셈 열쇠별」로 읽는다.
- 대안 D2(호스트 상한 값을 읽어 Claude Stop 을 면제)는 환경 변수가 거부를 통과로 바꾸는 스위치가 되어 REQ-HSF-001 의 금지 절과 부딪히고 발견 2 도 막지 못해 채택하지 않았다. D3(문서화만)도 채택하지 않았다.

### A.4 셈 열쇠를 무엇으로 정하는가

파싱이 실패한 페이로드에서는 `session_id` 도 `stop_hook_active` 도 읽을 수 없다(선행 SPEC §F.2). 남는 식별자는 훅 프로세스 자신의 환경과 프로세스 조상이다.

1. **세션 id 환경 변수 (1순위).** 저장소는 `CLAUDE_CODE_SESSION_ID` 를 「Claude Code 가 자신이 띄우는 모든 하위 프로세스(Bash 도구, 훅, stdio MCP 서버)의 환경에 찍는 세션 UUID이며, 훅 stdin 의 `session_id` 와 같다」고 문서화한다(`internal/config/envkeys.go` `EnvClaudeCodeSessionID`). 이 값을 열쇠로 쓰면 PID 재사용, 래퍼 셸 개수, `/clear` 뒤 이월 문제가 함께 줄어든다. **다만 훅 프로세스 환경에 이 변수가 있다는 것은 문서 서술일 뿐 이 카드에서 관측하지 않았다** — plan-audit 은 Bash 도구의 하위 프로세스에서만 값을 봤다. 실경로(`moai hook stop`)에서의 존재는 AC-SPC-016 이 잰다.
2. **세션 소유자 프로세스 (대체).** 변수가 없거나 비어 있으면 저장소의 정본 해석 `session.ResolveOwnerPID`(`internal/session/session_pid.go`)가 돌려주는 프로세스를 열쇠로 쓴다. 이 해석은 moai 에서 위로 올라가며 래퍼 셸(`sh`·`bash`·`zsh`·`cmd`·`powershell` 등)을 건너뛰고 가장 가까운 비래퍼 조상을 고른다. 그 파일의 주석이 적듯 「런타임의 `sh -c` 와 래퍼의 `exec` 가 접히는지에 따라 세션과 moai 사이에 셸이 0개 이상 끼인다」. 원시 `os.Getppid()` 는 그 셸 하나를 열쇠로 삼아 호출마다 다른 값을 낼 수 있으므로 쓰지 않는다. 해석하지 못하면 `(0, false)` 이며, 이 SPEC 은 그것을 「열쇠 없음」으로 다룬다(REQ-SPC-008).

배포되는 Stop 훅 설정은 `bash -c '… exec bash "$0"'` 로 래퍼를 부르고(`internal/template/templates/.claude/settings.json.tmpl` Stop 항목), 래퍼 `handle-stop.sh` 는 `exec moai hook stop` 으로 끝난다. macOS/Linux 에서는 정적 판독상 이 사슬이 접혀 moai 의 부모가 호스트 프로세스가 될 가능성이 높지만, 실제 프로세스 트리는 관측하지 않았다. 대체 경로가 셸을 건너뛰므로 이 판독이 틀려도 열쇠는 바뀌지 않는다.

---

## §B 용어

| 용어 | 뜻 |
|---|---|
| 파싱 실패 Stop | Claude 하네스(`--harness` 미지정 또는 `claude`)에서 `moai hook stop` 이 stdin 을 파싱하지 못한 호출. 선행 SPEC REQ-HSF-001 의 「`ReadInput` 이 오류를 돌려주는 모든 경우」와 같다 |
| 셈 열쇠 | 셈 기록 하나를 고르는 식별자. 세션 열쇠(`CLAUDE_CODE_SESSION_ID` 값) 또는 프로세스 열쇠(세션 소유자 프로세스) 중 하나이며 REQ-SPC-014 가 고른다. 두 종류는 값이 같아 보여도 서로 다른 열쇠다 |
| 세션 소유자 프로세스 | `session.ResolveOwnerPID` 가 돌려주는 프로세스 — 래퍼 셸을 건너뛴 가장 가까운 조상(§A.4). 이 SPEC 에서 「호스트 프로세스」는 이것을 뜻한다 |
| 셈 기록 | 셈 열쇠 하나에 대응하는 상태 기록. 연속 파싱 실패 Stop 의 횟수, 마지막 갱신 시각, 열쇠 종류를 담는다 |
| 기록 파일 이름 | 셈 열쇠에서 일방향 해시로 만든 고정 길이·고정 문자 집합의 이름(REQ-SPC-015). 구체 형식은 plan.md §C |
| 상태 영역 | 셈 기록이 놓이는 프로젝트 로컬 디렉터리. 프로젝트 루트는 선행 SPEC 의 영속 기록과 같은 방식으로 정한다(구체 경로는 plan.md §C) |
| N | 차단을 유지하는 연속 파싱 실패 Stop 의 최대 횟수. 8 로 고정한다 |
| 상한 해제 응답 | N 을 넘은 파싱 실패 Stop 에 내는 무의견 기본 출력(stdout `{}`, exit 0) |
| 만료 시간 | 마지막 갱신 뒤 이 시간이 지난 셈 기록은 없는 것으로 본다. **60분**이며 이름 붙은 상수 하나로 둔다(리드 판정 G1, 위치는 plan.md §B.1) |

---

## §C 요구사항 (GEARS)

- **REQ-SPC-001** (Event-driven): **When** Claude 하네스에서 파싱 실패 Stop 이 발생하면, the hook dispatcher **shall** 그 호출의 셈 열쇠(REQ-SPC-014)에 대응하는 셈 기록의 연속 횟수를 1 늘리고 마지막 갱신 시각을 현재로 바꾼 뒤, 늘어난 횟수로 REQ-SPC-002 와 REQ-SPC-003 중 하나를 적용한다. 셈 기록이 없으면 횟수 1 에서 시작한다.
- **REQ-SPC-002** (State-driven): **While** 늘어난 횟수가 N(8) 이하이면, the hook dispatcher **shall** 선행 SPEC REQ-HSF-001·004·008 이 정한 Claude Stop fail-closed 거부를 그대로 낸다 — 디스패치 없음, exit 0, stderr 한 줄, 파싱 실패 fail-closed 키의 영속 기록 한 건. 사유 문구만 REQ-SPC-010 을 따른다.
- **REQ-SPC-003** (State-driven): **While** 늘어난 횟수가 N 을 넘으면, the hook dispatcher **shall** 상한 해제 응답(stdout `{}`, exit 0)을 내고, 디스패치하지 않으며, 이벤트·하네스·파싱 오류 원인·연속 횟수·N·열쇠 종류(세션/프로세스)를 담은 stderr 한 줄과, 기존 세 키(`stdin-parse-fail-closed`·`stdin-parse-exempt`·`hook-fault`)와 구분되는 상한 해제 전용 키로 영속 기록 한 건을 남긴다. 횟수는 계속 늘어나며, REQ-SPC-005 의 초기화나 REQ-SPC-007 의 만료가 일어나기 전까지 이후의 파싱 실패 Stop 도 상한 해제 응답을 받는다 — 파싱할 수 없는 Stop 에서는 턴 경계를 알아낼 방법이 없기 때문이다.
- **REQ-SPC-004** (Unwanted): The hook dispatcher **shall not** 차단 유지와 상한 해제를 가르거나 N·만료 시간의 값을 정하는 데 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` 나 그 밖의 환경 변수·설정 키를 읽는다. N 과 만료 시간은 각각 코드 안의 이름 붙은 상수 하나이며, 이 판정을 바꾸는 실행 시점 스위치를 두지 않는다(선행 SPEC REQ-HSF-001 의 금지 절과 같은 원칙). 이 경로에서 허용되는 환경 변수 읽기는 REQ-SPC-014 의 열쇠 선택 하나뿐이며, 그것은 어느 기록을 쓸지만 고르고 문턱은 정하지 않는다. 다만 열쇠 선택은 두 방향으로 셈에 영향을 준다. 값이 호출마다 달라지는 방향은 셈이 늘 1 이 되어 차단이 이어지는 쪽(fail-closed)이다. 서로 다른 대화가 한 값을 공유하는 방향은 두 대화의 파싱 실패를 한 셈에 합쳐 상한 해제를 앞당기는 쪽이다 — 중첩 세션이 바깥 세션의 `CLAUDE_CODE_SESSION_ID` 를 물려받아 덮어쓰지 않는 경우와, 상속된 `MOAI_SESSION_PID` 가 조상을 가리켜 후손 세션들이 한 프로세스 열쇠로 모이는 경우(`internal/session/session_pid.go:118-123` 이 KNOWN RESIDUAL 로 적은 후손 사례)가 여기에 해당한다. 이 방향은 §F 「열쇠 공유」 행의 잔여 위험이다.
- **REQ-SPC-005** (Event-driven): **When** Claude 하네스의 `moai hook stop` 이 stdin 을 파싱하는 데 성공하면, the hook dispatcher **shall** 디스패치 전에 그 호출의 셈 열쇠에 대응하는 셈 기록을 지운다. 기록이 없으면 아무 파일도, 상태 영역 디렉터리도 만들지 않는다. 삭제에 실패하면 응답을 바꾸지 않고 실패 사유를 담은 stderr 한 줄만 남긴다. 셈 열쇠를 정할 수 없으면 아무것도 하지 않는다. 이후 Stop 핸들러의 결과(차단·통과·오류)는 셈에 영향을 주지 않는다.
- **REQ-SPC-006** (Unwanted): The hook dispatcher **shall not** Stop 이외의 이벤트 호출(PreToolUse·PostToolUse 등 도구 사용에 딸린 호출을 포함하며, 그 호출의 파싱 성공·실패를 가리지 않는다)로 셈 기록을 늘리거나 지운다.
- **REQ-SPC-007** (Event-driven): **When** 셈 기록에 접근할 때 그 기록의 마지막 갱신 시각이 만료 시간(60분)보다 오래됐으면, the hook dispatcher **shall** 그 기록을 없는 것으로 보고 지우며, 같은 접근에서 상태 영역에 남은 다른 열쇠의 만료된 셈 기록도 지운다. 정리 대상은 이름이 기록 파일 이름 형식(REQ-SPC-015)에 정확히 맞고 심볼릭 링크가 아닌 일반 파일로 한정한다 — 심볼릭 링크는 따라가지도 지우지도 않고, 형식에 맞지 않는 이름의 파일과 내용을 해석할 수 없는 다른 열쇠의 파일은 남긴다. 상태 영역 자체가 심볼릭 링크이거나 디렉터리가 아니면 정리를 하지 않는다. 정리는 훅의 응답을 바꾸지 않고, 정리 실패는 stderr 한 줄로만 남는다.
- **REQ-SPC-008** (Event-driven): **When** 셈 열쇠를 정할 수 없거나(REQ-SPC-014 의 두 순위 모두 값을 주지 못함) 셈 기록을 읽고 쓸 수 없으면(상태 영역 생성·쓰기 실패, 상태 영역이 심볼릭 링크이거나 디렉터리가 아님, 자기 기록 자리가 일반 파일이 아님), the hook dispatcher **shall** 상한을 적용하지 않고 REQ-SPC-002 의 fail-closed 거부를 내며, 셈을 적용하지 못한 사유를 담은 stderr 한 줄을 남긴다. 이때 심볼릭 링크의 대상은 읽거나 고치지 않는다. 일반 파일이지만 내용을 해석할 수 없는 자기 기록은 없는 것으로 보고 새 기록으로 덮어쓴다. 셈 기록의 쓰기는 읽는 쪽이 반쯤 쓰인 내용을 보지 않는 방식으로 한다.
- **REQ-SPC-009** (Where): **Where** 호출이 `--harness codex` 모드이면, the hook dispatcher **shall** 선행 SPEC 의 동작 — (Codex, Stop) 면제(REQ-HSF-012·013), Codex fail-closed 출력의 바이트, 그 사유 문구 — 을 바꾸지 않으며, 셈 기록을 만들거나 읽지 않는다.
- **REQ-SPC-010** (Where): **Where** 호출이 Claude 하네스이면, the fail-closed 거부의 사유 문구 **shall** 다음 고정 문자열과 바이트 단위로 같다 — `fail-closed: hook stdin could not be parsed as JSON. Do not edit hook scripts or settings files to get past this; stop and tell a human operator (.moai/docs/hook-stdin-fail-closed.md)`. 이 문자열은 선행 SPEC REQ-HSF-010 의 (a) `fail-closed` 표시, (b) 원인 고정 문구, (c) 운영자 문서 식별자를 그대로 싣고, 거기에 훅·설정 파일을 편집하지 말고 사람에게 알리라는 지시를 더한다. 복구 절차(moai 갱신, `disableAllHooks` 등 훅 비활성화 방법)와 파싱에 실패한 페이로드에서 유래한 내용은 싣지 않는다. 이 요구는 Claude 하네스의 네 결정 이벤트(PreToolUse·PermissionRequest·Stop·UserPromptSubmit)와 `moai hook agent` 의 Claude 모드 fail-closed 에 적용된다.
- **REQ-SPC-011** (Unwanted): The 상한 해제 응답의 stderr 줄과 영속 기록 **shall not** 파싱에 실패한 페이로드의 원문이나 그 일부를 싣는다. 입력에서 유래해 허용되는 정보는 선행 SPEC REQ-HSF-011 과 같다(파싱 오류 메시지, stdin 바이트 수).
- **REQ-SPC-012** (Ubiquitous): The 운영자 문서 `hook-stdin-fail-closed.md`(템플릿 원본 `internal/template/templates/.moai/docs/`) **shall** Claude Stop 의 moai 자체 상한 — N 값, N 을 넘은 뒤 상한 해제 응답이 이어진다는 것, 파싱에 성공한 Stop 과 만료(60분)만이 셈을 되돌린다는 것, 셈 열쇠의 선택 규칙(세션 id 우선, 없으면 세션 소유자 프로세스), 셈 기록이 놓이는 곳 — 과, 이 상한이 호스트 상한 값과 무관하다는 사실을 적는다.
- **REQ-SPC-013** (Ubiquitous): The 코드 주석 **shall** 측정 사실과 일치한다 — `internal/codexadapter/stop_cap.go` `HostLacksStopBlockCap` 주석의 「Claude Code 는 측정되지 않았다」 문장과 `internal/cli/hook_stdin_failclosed.go` 의 Stop 루프 `@MX:WARN`(현재 「호스트 상한만이 루프를 묶는다」)을, t1230·t1272 측정(기본 상한은 도구 사용 없이 이어진 차단에만 걸린다는 해석, 200 주입 조건)과 이 SPEC 의 moai 자체 상한을 반영해 고친다. 이 수정은 run-phase 과제다.
- **REQ-SPC-014** (Ubiquitous): The hook dispatcher **shall** 셈 열쇠를 다음 순서로 정한다 — (1) 훅 프로세스 환경의 `CLAUDE_CODE_SESSION_ID` 가 있고 앞뒤 공백을 뺀 값이 비어 있지 않으면 그 값을 세션 열쇠로, (2) 그렇지 않으면 저장소의 정본 세션 소유자 해석이 돌려준 프로세스를 프로세스 열쇠로. 원시 부모 프로세스 id 를 열쇠로 쓰지 않으며, 두 종류의 열쇠는 값이 같아도 서로 다른 셈 기록에 대응한다.
- **REQ-SPC-015** (Ubiquitous): The 셈 기록의 파일 이름 **shall** 열쇠 종류와 열쇠 값으로부터 일방향 해시로 만든 고정 길이·고정 문자 집합의 이름이며, 상태 영역 바로 아래에 놓인다. 세션 id 에 든 어떤 문자(`/`·`\`·`..`·NUL·제어 문자 포함)도 파일 경로를 상태 영역 밖으로 옮기거나 경로 구성 요소를 늘리지 못한다. 기록과 stderr 줄에는 열쇠 종류만 싣고 세션 id 원문은 싣지 않는다.

---

## §D 제외 범위

아래 항목은 이 SPEC 의 범위 밖이다.

### Out of Scope — Codex 쪽 변경

- Codex 번역 틀(`TranslateCodex` 가 사유를 감싸는 고정 문구)과 Codex 하네스의 사유 문구. 카드 t1233 의 범위다. 이 SPEC 은 Codex 쪽 출력 바이트를 바꾸지 않는다(REQ-SPC-009).
- (Codex, Stop) 면제 술어 `HostLacksStopBlockCap` 의 판정 논리. 주석만 고친다(REQ-SPC-013).

### Out of Scope — 호스트 상한과 런처 주입

- `DefaultRaisedStopHookBlockCap = 200` 의 값이나 주입 조건(칸반·팩토리·무한 goal) 변경. 그 주입은 무한 goal 루프를 위한 의도된 교환이며, 이 SPEC 은 그 값과 무관하게 파싱 실패 루프만 묶는다.
- 호스트 상한의 의미(도구 사용이 셈을 초기화하는지)를 업스트림에서 확인하거나 바꾸는 일.

### Out of Scope — 다른 이벤트의 루프

- SubagentStop 은 관측 이벤트로 남는다(선행 SPEC §B). PreToolUse·PermissionRequest·UserPromptSubmit 의 fail-closed 거부에는 상한을 두지 않는다 — 그 셋은 턴 종료를 미루는 게이트가 아니라 도구 호출·프롬프트를 막는 게이트이며, 거부가 턴을 끝없이 잇지 않는다.
- 파싱에 **성공한** Stop 의 반복 차단(예: stop-goal 평가기의 block). 그것은 `stop_hook_active` 와 goal 엔진의 상한이 다룬다.

### Out of Scope — 설정 파일 쓰기 방어

- 모델이 `.claude/settings*.json`·훅 스크립트·셈 기록 자체를 편집하는 것을 막는 가드. 이 SPEC 은 사유 문구로 그 시도를 줄이려 할 뿐 막지 않는다(선행 SPEC §D 「설정 파일 쓰기 방어」와 같은 경계). 발견 3 의 bypass 세션 위험은 §F 의 잔여 위험으로 남는다.

### Out of Scope — 세션 id 의 해석과 보정

- `CLAUDE_CODE_SESSION_ID` 의 형식(UUID 여부) 검증과, 그 값과 페이로드 `session_id` 의 일치 확인. 파싱 실패에서는 페이로드를 읽을 수 없고, 형식은 이 SPEC 이 보장받은 계약이 아니다 — 값은 해시로만 쓰인다(REQ-SPC-015).
- `/clear` 뒤에도 환경 변수가 바뀌지 않는 경우의 셈 이월 보정. 바뀌는지는 측정 항목이며(plan.md M5c), 바뀌지 않으면 §F 의 잔여 위험으로 남는다.

### Out of Scope — 선행 SPEC 의 상태와 AC

- 선행 SPEC 의 `status: completed` 는 유지한다. 그 SPEC 의 acceptance.md 는 고치지 않는다 — Claude 하네스 사유 문구를 바이트로 비교하는 AC-HSF-001(d)·(e4) 와 AC-HSF-013 의 Claude 모드 행의 기대값은 이 SPEC 의 run-phase 가 테스트에서 REQ-SPC-010 문자열로 바꾸며, 그 추적은 이 SPEC 의 acceptance.md 가 소유한다.

---

## §E 선행 SPEC 과의 관계

| SPEC | 관계 | 내용 |
|---|---|---|
| SPEC-HOOK-STDIN-FAILCLOSED-001 (completed) | **부분 대체·확장** | REQ-HSF-001 의 Claude Stop 거부에 「N 초과 시 상한 해제」 예외를 더하고(REQ-SPC-001~003), REQ-HSF-010 의 사유 문구를 Claude 하네스에 한해 대체한다(REQ-SPC-010). 그 SPEC 0.4.3 이 REQ-HSF-010 에 대체 표시와 이 SPEC 을 가리키는 포인터를 달았고, §F.2·REQ-HSF-009·plan.md Q8·§D 의 낡은 측정 서술을 t1230·t1272 결과로 정정했다 |
| SPEC-INFINITE-GOAL-001 / SPEC-FACTORY-MODE-001 | **소비, 변경 없음** | 두 SPEC 이 도입한 200 주입이 발견 1 의 조건이다. 이 SPEC 은 그 주입을 바꾸지 않는다 |

---

## §F 전제와 위험

| 위험 | 설명 | 대응 |
|---|---|---|
| 프로세스 열쇠의 PID 재사용 — 해제 상태 이어받기 | 세션 열쇠를 쓸 수 없어 프로세스 열쇠로 떨어진 경우에만 해당한다. 호스트가 끝난 뒤 60분 안에 같은 id 를 받은 다른 호스트는 남은 셈 기록을 이어받는다. 그 기록이 상한 해제 상태(횟수 9 이상)였다면 새 호스트의 **첫** 파싱 실패 Stop 이 차단 없이 무의견 응답을 받는다 — 파싱 성공이나 만료 전까지 Stop 한 게이트가 풀린 채로 시작하는 fail-open 쪽 결과다 | 잔여 위험으로 명시하며, 이 동작은 AC-SPC-006 대조가 「같은 열쇠, 횟수 9, 59분 전 기록 → 첫 호출이 해제」로 고정해 우연이 아니라 진술된 동작으로 둔다. 창은 만료 시간(60분)으로 묶이고, 세션 열쇠가 있으면 생기지 않는다. 프로세스 지문을 기록에 더하는 보강은 plan.md §C 5 의 선택지다 |
| `/clear` 뒤의 이월 | 세션 열쇠에서 `/clear` 뒤 `CLAUDE_CODE_SESSION_ID` 가 바뀌면 새 기록에서 1 부터 센다(의도). 바뀌지 않으면 해제 상태를 포함한 셈이 새 대화로 이월된다. 프로세스 열쇠에서는 같은 호스트 프로세스이므로 늘 이월된다 | 바뀌는지는 plan.md M5c 의 측정 항목(AC-SPC-015 (iii), 관측)이다. 측정 전까지 이월 가능성은 잔여 위험이다 |
| 한 호스트 안의 여러 Stop 발생원 | in-process 팀원처럼 한 호스트 프로세스가 여러 대화의 Stop 을 낸다면, 프로세스 열쇠에서는 셈이 섞인다. 세션 열쇠에서는 팀원 훅의 환경에 팀원 자신의 세션 id 가 찍히는지에 달려 있다 | 잔여 위험으로 기록한다. 팀원 훅 환경의 세션 id 는 재지 않았다 |
| 열쇠 공유 — 셈이 합쳐져 해제가 앞당겨짐 | 서로 다른 대화가 같은 셈 열쇠를 쓰면 두 대화의 파싱 실패가 한 셈 기록에 쌓여, 어느 한쪽이 자기 대화에서 N 번을 채우기 전에 상한 해제를 받는다 — fail-open 쪽이다. 현실적인 경로는 둘이다. (a) 세션 열쇠: 중첩 `claude -p` 가 바깥 세션의 `CLAUDE_CODE_SESSION_ID` 를 환경으로 물려받고 자기 값으로 덮어쓰지 않는 경우. (b) 프로세스 열쇠: `session.ResolveOwnerPID` 가 상속된 `MOAI_SESSION_PID` 를 조상이면 받아들이므로, 한 세션의 후손 세션들이 같은 프로세스 열쇠로 모이는 경우(`session_pid.go:118-123` KNOWN RESIDUAL) | 두 대화가 한 열쇠에서 동시에 파싱 실패 상태여야 하므로 가능성은 낮고, 새로 열리는 악용 면은 없다(훅 환경을 바꾸려면 설정 편집이 필요하고 그 면에는 이미 `disableAllHooks` 가 있다). (a) 는 직접 관측하지 않는다 — M5a 는 상속을 먼저 지우고 재므로, AC-SPC-016 의 값 일치 판정은 호스트가 자기 세션 id 를 찍는지만 보이며 (a) 는 그로부터의 간접 추정에 그친다. 잔여 위험으로 기록한다 |
| 같은 열쇠의 동시 Stop | 열쇠가 다르면 셈 기록이 따로 있어 서로 영향이 없다. 같은 열쇠의 Stop 두 개가 동시에 셈 기록을 읽고 쓰면 증가 한 번이 사라질 수 있다 | 증가가 사라지는 쪽은 차단이 한 번 더 이어지는 방향(fail-closed 쪽)이다. 잠금은 두지 않는다 |
| 셸이 아닌 래퍼 | 대체 경로는 래퍼 셸을 건너뛰므로 `exec` 여부는 열쇠를 바꾸지 않는다. 다만 사용자가 래퍼 목록에 없는 프로그램(예: 인터프리터)으로 moai 를 부르면 그 프로그램이 호출마다 새 프로세스라 프로세스 열쇠가 매번 달라진다 | 그 경우 셈이 늘 1 이라 오늘의 동작(호스트 상한만)으로 돌아간다 — fail-closed 쪽이다. 세션 열쇠가 있으면 래퍼 모양과 무관하다 |
| Windows | 세션 열쇠가 있으면 다른 플랫폼과 같다. 대체 경로에서는 Git Bash 의 `exec` 가 프로세스를 대체하지 않고 에뮬레이션하므로 호스트와 moai 사이에 `bash` 프로세스가 남는다. 해석은 Toolhelp32 로 조상을 읽고 `bash`·`sh`·`cmd`·`powershell`·`pwsh` 를 건너뛰므로 호스트를 찾을 것으로 **기대**한다 | Windows 에서는 재지 않았다. 조상 탐색이 실패하면 「열쇠 없음」이 되어 차단이 유지된다(REQ-SPC-008, 오늘의 동작). 완료 정의에 `GOOS=windows` 빌드가 있다 |
| 셈 기록 조작 | 셈 기록은 프로젝트 상태 영역의 파일이므로 쓰기 권한이 있는 모델이 큰 횟수를 써 넣어 Stop 차단을 즉시 해제할 수 있다. 상태 영역이나 기록을 심볼릭 링크로 바꿔 moai 의 삭제 경로를 다른 곳으로 돌리려는 시도도 가능하다 | 풀리는 것은 Stop 한 게이트이고, 도구 호출을 막는 결정 이벤트 거부는 영향이 없다. 설정 편집으로 모든 훅을 끄는 면(`disableAllHooks`)이 이미 있으므로 새로 여는 면은 그보다 좁다. 삭제 경로는 링크를 따라가지 않고 이름 형식에 맞는 일반 파일만 지우므로(REQ-SPC-007) 상태 영역 밖을 지우지 못하며, 링크로 바뀐 상태 영역은 상한 없이 차단을 유지한다(REQ-SPC-008). 잔여 위험으로 기록한다 |
| 상한 해제 뒤의 지속 해제 | N 을 넘은 뒤에는 파싱 성공이나 만료 전까지 그 호스트의 파싱 실패 Stop 이 계속 통과한다 — Stop 에 걸린 가드 핸들러가 그동안 실행되지 않는다 | 파싱할 수 없는 Stop 에서 턴 경계를 알아낼 방법이 없으므로 D1 을 문자 그대로 적용한 결과다. 매 해제가 stderr 와 전용 기록으로 소리를 낸다(REQ-SPC-003) |
| 증거의 폭 | 호스트 상한 관측은 Claude Code 2.1.283, `-p` 비대화형, haiku 한 모델, 팔당 1회다. 「도구 사용이 셈을 초기화한다」는 해석이다 | 이 SPEC 의 상한은 호스트 상한의 의미와 무관하게 동작하도록 요구되므로(REQ-SPC-004) 해석이 틀려도 설계는 유효하다. 업스트림이 상한 의미를 바꾸면 A.1 표가 낡는다 |
| 사유 문구의 효과 | 새 사유 문구가 모델의 훅 편집 시도를 실제로 줄이는지는 재지 않았다 | 효과 측정은 plan.md M5b 의 LIVE 재측정에서 관측 항목으로 둔다(차단 AC 아님) |
