---
id: SPEC-HOOK-STOP-PARSE-CAP-001
title: "Claude Stop 파싱 실패 차단의 moai 자체 상한 — 호스트 상한에 기대지 않는 루프 한계와 사유 문구 개정"
version: "0.1.0"
status: draft
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
- 대안 D2(호스트 상한 값을 읽어 Claude Stop 을 면제)는 환경 변수가 거부를 통과로 바꾸는 스위치가 되어 REQ-HSF-001 의 금지 절과 부딪히고 발견 2 도 막지 못해 채택하지 않았다. D3(문서화만)도 채택하지 않았다.

### A.4 왜 moai 가 호스트 프로세스를 열쇠로 쓰는가

파싱이 실패한 페이로드에서는 `session_id` 도 `stop_hook_active` 도 읽을 수 없다(선행 SPEC §F.2). 남는 식별자는 훅 프로세스 자신의 환경뿐이다. 배포되는 Stop 훅 설정은 `bash -c '… exec bash "$0"'` 로 래퍼를 부르고(`.claude/settings.json` `Stop` 항목), 래퍼 `.claude/hooks/moai/handle-stop.sh` 는 `exec moai hook stop` 으로 끝난다. 두 단계 모두 `exec` 이므로 moai 프로세스의 부모는 호스트(Claude Code) 프로세스다. 같은 가정을 SessionStart 의 프로필 임대 등록이 이미 쓰고 있다(`internal/hook/session_start.go` `registerProfileLease` 의 `os.Getppid()`). 이 판독은 이 트리의 파일 내용으로 확인했고, 실제 프로세스 트리를 관측하지는 않았다.

---

## §B 용어

| 용어 | 뜻 |
|---|---|
| 파싱 실패 Stop | Claude 하네스(`--harness` 미지정 또는 `claude`)에서 `moai hook stop` 이 stdin 을 파싱하지 못한 호출. 선행 SPEC REQ-HSF-001 의 「`ReadInput` 이 오류를 돌려주는 모든 경우」와 같다 |
| 호스트 프로세스 | 훅 프로세스의 부모 프로세스. §A.4 의 `exec` 사슬에서는 Claude Code 프로세스다 |
| 셈 기록 | 호스트 프로세스 하나에 대응하는 상태 기록. 연속 파싱 실패 Stop 의 횟수와 마지막 갱신 시각을 담는다 |
| 상태 영역 | 셈 기록이 놓이는 프로젝트 로컬 디렉터리. 프로젝트 루트는 선행 SPEC 의 영속 기록과 같은 방식으로 정한다(구체 경로는 plan.md §C) |
| N | 차단을 유지하는 연속 파싱 실패 Stop 의 최대 횟수. 8 로 고정한다 |
| 상한 해제 응답 | N 을 넘은 파싱 실패 Stop 에 내는 무의견 기본 출력(stdout `{}`, exit 0) |
| 만료 시간 | 마지막 갱신 뒤 이 시간이 지난 셈 기록은 없는 것으로 본다. 값은 plan.md §B 가 정한다 |

---

## §C 요구사항 (GEARS)

- **REQ-SPC-001** (Event-driven): **When** Claude 하네스에서 파싱 실패 Stop 이 발생하면, the hook dispatcher **shall** 그 호출의 호스트 프로세스에 대응하는 셈 기록의 연속 횟수를 1 늘리고 마지막 갱신 시각을 현재로 바꾼 뒤, 늘어난 횟수로 REQ-SPC-002 와 REQ-SPC-003 중 하나를 적용한다. 셈 기록이 없으면 횟수 1 에서 시작한다.
- **REQ-SPC-002** (State-driven): **While** 늘어난 횟수가 N(8) 이하이면, the hook dispatcher **shall** 선행 SPEC REQ-HSF-001·004·008 이 정한 Claude Stop fail-closed 거부를 그대로 낸다 — 디스패치 없음, exit 0, stderr 한 줄, 파싱 실패 fail-closed 키의 영속 기록 한 건. 사유 문구만 REQ-SPC-010 을 따른다.
- **REQ-SPC-003** (State-driven): **While** 늘어난 횟수가 N 을 넘으면, the hook dispatcher **shall** 상한 해제 응답(stdout `{}`, exit 0)을 내고, 디스패치하지 않으며, 이벤트·하네스·파싱 오류 원인·연속 횟수·N 을 담은 stderr 한 줄과, 기존 세 키(`stdin-parse-fail-closed`·`stdin-parse-exempt`·`hook-fault`)와 구분되는 상한 해제 전용 키로 영속 기록 한 건을 남긴다. 횟수는 계속 늘어나며, REQ-SPC-005 의 초기화나 REQ-SPC-007 의 만료가 일어나기 전까지 이후의 파싱 실패 Stop 도 상한 해제 응답을 받는다 — 파싱할 수 없는 Stop 에서는 턴 경계를 알아낼 방법이 없기 때문이다.
- **REQ-SPC-004** (Unwanted): The hook dispatcher **shall not** 차단 유지와 상한 해제를 가르는 데 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` 나 그 밖의 환경 변수·설정 키를 읽는다. N 은 코드 안의 이름 붙은 상수 하나이며, 이 판정을 바꾸는 실행 시점 스위치를 두지 않는다(선행 SPEC REQ-HSF-001 의 금지 절과 같은 원칙).
- **REQ-SPC-005** (Event-driven): **When** Claude 하네스의 `moai hook stop` 이 stdin 을 파싱하는 데 성공하면, the hook dispatcher **shall** 디스패치 전에 그 호스트 프로세스의 셈 기록을 지운다. 이후 Stop 핸들러의 결과(차단·통과·오류)는 셈에 영향을 주지 않는다.
- **REQ-SPC-006** (Unwanted): The hook dispatcher **shall not** Stop 이외의 이벤트 호출(PreToolUse·PostToolUse 등 도구 사용에 딸린 호출을 포함하며, 그 호출의 파싱 성공·실패를 가리지 않는다)로 셈 기록을 늘리거나 지운다.
- **REQ-SPC-007** (Event-driven): **When** 셈 기록에 접근할 때 그 기록의 마지막 갱신 시각이 만료 시간보다 오래됐으면, the hook dispatcher **shall** 그 기록을 없는 것으로 보고 지우며, 같은 접근에서 상태 영역에 남은 다른 호스트 프로세스의 만료된 셈 기록도 지운다. 정리는 훅의 응답을 바꾸지 않고, 정리 실패는 stderr 한 줄로만 남는다.
- **REQ-SPC-008** (Event-driven): **When** 호스트 프로세스를 식별할 수 없거나(부모 프로세스 id 가 1 이하) 셈 기록을 읽고 쓸 수 없으면(상태 영역 생성·쓰기 실패), the hook dispatcher **shall** 상한을 적용하지 않고 REQ-SPC-002 의 fail-closed 거부를 내며, 셈을 적용하지 못한 사유를 담은 stderr 한 줄을 남긴다. 내용을 해석할 수 없는 셈 기록은 없는 것으로 보고 새 기록으로 덮어쓴다. 셈 기록의 쓰기는 읽는 쪽이 반쯤 쓰인 내용을 보지 않는 방식으로 한다.
- **REQ-SPC-009** (Where): **Where** 호출이 `--harness codex` 모드이면, the hook dispatcher **shall** 선행 SPEC 의 동작 — (Codex, Stop) 면제(REQ-HSF-012·013), Codex fail-closed 출력의 바이트, 그 사유 문구 — 을 바꾸지 않으며, 셈 기록을 만들거나 읽지 않는다.
- **REQ-SPC-010** (Where): **Where** 호출이 Claude 하네스이면, the fail-closed 거부의 사유 문구 **shall** 다음 고정 문자열과 바이트 단위로 같다 — `fail-closed: hook stdin could not be parsed as JSON. Do not edit hook scripts or settings files to get past this; stop and tell a human operator (.moai/docs/hook-stdin-fail-closed.md)`. 이 문자열은 선행 SPEC REQ-HSF-010 의 (a) `fail-closed` 표시, (b) 원인 고정 문구, (c) 운영자 문서 식별자를 그대로 싣고, 거기에 훅·설정 파일을 편집하지 말고 사람에게 알리라는 지시를 더한다. 복구 절차(moai 갱신, `disableAllHooks` 등 훅 비활성화 방법)와 파싱에 실패한 페이로드에서 유래한 내용은 싣지 않는다. 이 요구는 Claude 하네스의 네 결정 이벤트(PreToolUse·PermissionRequest·Stop·UserPromptSubmit)와 `moai hook agent` 의 Claude 모드 fail-closed 에 적용된다.
- **REQ-SPC-011** (Unwanted): The 상한 해제 응답의 stderr 줄과 영속 기록 **shall not** 파싱에 실패한 페이로드의 원문이나 그 일부를 싣는다. 입력에서 유래해 허용되는 정보는 선행 SPEC REQ-HSF-011 과 같다(파싱 오류 메시지, stdin 바이트 수).
- **REQ-SPC-012** (Ubiquitous): The 운영자 문서 `hook-stdin-fail-closed.md`(템플릿 원본 `internal/template/templates/.moai/docs/`) **shall** Claude Stop 의 moai 자체 상한 — N 값, N 을 넘은 뒤 상한 해제 응답이 이어진다는 것, 파싱에 성공한 Stop 과 만료만이 셈을 되돌린다는 것, 셈 기록이 놓이는 곳 — 과, 이 상한이 호스트 상한 값과 무관하다는 사실을 적는다.
- **REQ-SPC-013** (Ubiquitous): The 코드 주석 **shall** 측정 사실과 일치한다 — `internal/codexadapter/stop_cap.go` `HostLacksStopBlockCap` 주석의 「Claude Code 는 측정되지 않았다」 문장과 `internal/cli/hook_stdin_failclosed.go` 의 Stop 루프 `@MX:WARN`(현재 「호스트 상한만이 루프를 묶는다」)을, t1230·t1272 측정(기본 상한은 도구 사용 없이 이어진 차단에만 걸린다는 해석, 200 주입 조건)과 이 SPEC 의 moai 자체 상한을 반영해 고친다. 이 수정은 run-phase 과제다.

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
| 부모 프로세스 id 재사용 | 호스트가 끝난 뒤 같은 id 를 받은 다른 호스트가 남은 셈 기록을 이어받을 수 있다. 그러면 새 세션의 첫 파싱 실패 Stop 이 N 보다 이르게 상한 해제 응답을 받는다 | 만료(REQ-SPC-007)가 재사용 창을 만료 시간으로 묶는다. 프로세스 시작 시각 같은 신원 정보를 열쇠에 더하는 방안은 plan.md §C 에 선택지로 적고 이 SPEC 의 요구로 두지 않는다 |
| 한 호스트 안의 여러 Stop 발생원 | in-process 팀원처럼 한 호스트 프로세스가 여러 대화의 Stop 을 낼 수 있다면 셈이 섞인다 — 한쪽의 파싱 성공이 다른 쪽 셈을 지우거나, 두 쪽 실패가 합산돼 이르게 해제된다 | 잔여 위험으로 기록한다. in-process 팀원의 Stop 이 같은 부모 프로세스에서 `moai hook stop` 으로 오는지는 재지 않았다 |
| 같은 프로젝트의 동시 세션 | 호스트 프로세스가 다르면 셈 기록이 따로 있으므로 서로 영향이 없다. 같은 호스트의 Stop 두 개가 동시에 셈 기록을 읽고 쓰면 증가 한 번이 사라질 수 있다 | 증가가 사라지는 쪽은 차단이 한 번 더 이어지는 방향(fail-closed 쪽)이다. 잠금은 두지 않는다 |
| `exec` 가 아닌 래퍼 | 사용자가 래퍼를 `exec` 없이 고치면 moai 의 부모가 호출마다 새 셸이 되어 셈이 매번 1 에서 시작한다 | 그 경우 상한이 걸리지 않고 오늘의 동작(호스트 상한만)으로 돌아간다 — fail-closed 쪽이다. t1272 의 LIVE 훅 래퍼(`hook.sh`)도 파이프로 moai 를 불러 이 모양이었으므로, run-phase 의 LIVE 재측정은 `exec` 형 래퍼로 한다(plan.md M4) |
| 셈 기록 조작 | 셈 기록은 프로젝트 상태 영역의 파일이므로 쓰기 권한이 있는 모델이 큰 횟수를 써 넣어 Stop 차단을 즉시 해제할 수 있다 | 풀리는 것은 Stop 한 게이트이고, 도구 호출을 막는 결정 이벤트 거부는 영향이 없다. 설정 편집으로 모든 훅을 끄는 면(`disableAllHooks`)이 이미 있으므로 새로 여는 면은 그보다 좁다. 잔여 위험으로 기록한다 |
| 상한 해제 뒤의 지속 해제 | N 을 넘은 뒤에는 파싱 성공이나 만료 전까지 그 호스트의 파싱 실패 Stop 이 계속 통과한다 — Stop 에 걸린 가드 핸들러가 그동안 실행되지 않는다 | 파싱할 수 없는 Stop 에서 턴 경계를 알아낼 방법이 없으므로 D1 을 문자 그대로 적용한 결과다. 매 해제가 stderr 와 전용 기록으로 소리를 낸다(REQ-SPC-003) |
| 증거의 폭 | 호스트 상한 관측은 Claude Code 2.1.283, `-p` 비대화형, haiku 한 모델, 팔당 1회다. 「도구 사용이 셈을 초기화한다」는 해석이다 | 이 SPEC 의 상한은 호스트 상한의 의미와 무관하게 동작하도록 요구되므로(REQ-SPC-004) 해석이 틀려도 설계는 유효하다. 업스트림이 상한 의미를 바꾸면 A.1 표가 낡는다 |
| 사유 문구의 효과 | 새 사유 문구가 모델의 훅 편집 시도를 실제로 줄이는지는 재지 않았다 | 효과 측정은 plan.md M4 의 LIVE 재측정에서 관측 항목으로 둔다(차단 AC 아님) |
