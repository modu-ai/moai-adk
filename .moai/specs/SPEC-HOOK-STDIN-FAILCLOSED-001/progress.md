# Progress — SPEC-HOOK-STDIN-FAILCLOSED-001

## §E.1 Plan-phase Audit-Ready Signal

- 작성: manager-spec (card t1152), 브랜치 `WT-hook-stdin-failclosed`, 기준 트리 HEAD `60017eb83`
- 산출물: spec.md, plan.md, acceptance.md, progress.md (Tier M)
- SPEC ID 형식 검사: `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`; `.moai/specs/SPEC-HOOK-STDIN-FAILCLOSED-001` 사전 부재 확인(`ls` → No such file or directory)
- t1099 조상 여부: `git merge-base --is-ancestor fabc33812 HEAD` → exit 1 (미착지). run 착수 전제는 spec.md §F.1
- 미해결(0.1.0 시점 기록): plan.md §B Q1~Q7 (Q1·Q3 은 Kickoff 전 판정 필요) — 0.2.0 에서 Q2 를 Kickoff 차단으로 올렸다(아래 개정 기록)
- 측정하지 않은 전제: 호스트가 5 MiB 초과 `tool_input` 을 훅에 넘기는지(Q5), Codex 의 Stop 연속 차단 상한(Q2), Claude 호스트가 네 가지 거부 JSON 을 문서대로 존중하는지(저장소 독트린 `hooks-system.md` 와 `internal/hook/types.go` 에 근거, 공식 문서 재확인 안 함)

### 개정 0.2.0 (2026-09-24)

- 계기: plan-audit 1회차 FAIL 0.66 (`.moai/reports/t1152/plan-audit.md`, 로컬 증거). 결함 D0~D13 을 spec.md·plan.md·acceptance.md 에 반영했다. 처분 요약은 spec.md HISTORY 0.2.0 행.
- 개정 기준 트리: 브랜치 `WT-hook-stdin-failclosed`, HEAD `d2ddaae7e`. t1099 인용은 계속 `fabc33812` 고정. plan-audit 관측(2026-09-24): t1099 는 `fe4fd9d4d` 로 전진했으나 인용 파일 6개의 diff 는 0.
- 이번 개정에서 직접 측정한 것: 중첩 깊이 파싱 실패(저장소 밖 scratchpad 의 독립 Go 프로그램, go1.26.8, `map[string]json.RawMessage` 대상) — 깊이 9000 `bytes=18052 err=<nil>`, 깊이 10000 `bytes=20052 err=invalid character '[' exceeded max depth`, 깊이 10001 `bytes=20054` 같은 오류. 줄 번호 재확인: `internal/cli/hook.go:389-392`(worktree 빈 출력), `internal/hook/normalize.go:85-88`, `internal/hook/protocol.go:44`·`:52-56`·`:58`, `internal/codexadapter/events.go:69-84`, `internal/cli/hook_protocol_fix_test.go:57-59`, `internal/cli/hook_stop_goal.go:120-121`, `fabc33812:internal/cli/hook_codex_failclosed.go:20`·`:28-29`·`:34-58`·`:44-49`, `fabc33812:internal/codexadapter/output.go:77-82`, `fabc33812:internal/codexadapter/diagnostics.go:24`.
- 직접 측정하지 않고 plan-audit 에서 인용한 것: worktree 두 이벤트의 빈 stdout 실행 재현, 중복 `switch` 변이 미검출, `runAgentHook` 의 `rc=0, stdout={}` 재현(모두 codex 백엔드 실행).
- 남은 Kickoff 차단 질문: plan.md §B.1 Q2 → Q1(탈출 장치 메커니즘 포함) → Q3. Q4~Q7 은 추적 항목(표식 없음).

### 개정 0.3.0 (2026-09-24)

- 계기: Kickoff 차단 질문 셋의 판정. 판정 기록은 plan.md §B.1(「판정됨」, 판정 전 선택지 보존), 처분 요약은 spec.md HISTORY 0.3.0 행.
  - Q2: t1152 레인 세션의 직접 측정(2026-09-24 12:39–12:50 KST, codex-cli 0.156.1). 증거 `.moai/reports/t1152/q2-codex-stop-cap.md`(로컬, gitignored — 이 커밋에 포함하지 않음). 결과: `codex exec` 가 Stop 의 JSON block 을 191회 연속 받아들이고 스스로 끊지 않음(`stop_hook_active` false 1 / true 190). 관측 범위: 기본 설정·비대화형·`gpt-6-astra`·191회/약 10분. manager-spec 은 이 측정을 재실행하지 않았고 증거 파일을 판독만 했다.
  - Q1: 운영자 판정(t1152 레인 AskUserQuestion) — A1. 결정 이벤트 4개를 두 하네스에서 fail-closed, (Codex, Stop) 만 면제 술어 하나로 관측 경로. 탈출 장치는 셸 환경 변수 하나 + 프로젝트·로컬 `env` 선언 시 무시 + 판독 실패 시 불인정.
  - Q3: 운영자 판정(같은 채널) — `runAgentHook` 포함.
- 신설 REQ: REQ-HSF-012·013·014·015·016. 개정 REQ: 001·003·005·007·008·009·010·011. 신설 AC: AC-HSF-011·012·013. 개정 AC: 002·003·006·007·008·010·GATE, §1 대응표, §D 갱신 대상 테스트(`TestRunAgentHook_ReadInputError`, `TestRunAgentHook_AllActionSuffixes` 추가).
- 개정 기준 트리: 브랜치 `WT-hook-stdin-failclosed`, HEAD `44dfc25fd`. t1099 인용은 계속 `fabc33812` 고정.
- 이번 개정에서 이 트리에서 다시 확인한 줄 번호: `internal/cli/hook.go:45`(`--harness` PersistentFlags), `:447`(`runAgentHook`), `:455-463`(파싱 실패 분기), `:469-481`(action → 이벤트 매핑 switch, 대입 `:472`·`:479`), `:54`·`:57`·`:62`·`:63`(하위 명령 표의 결정 이벤트 4행); `internal/cli/misc_coverage_test.go:355`·`:372-374`; `internal/cli/coverage_fixes_test.go:810`·`:848`·`:852`. AC-HSF-003(b) 기준선 grep 을 HEAD `44dfc25fd` 에서 다시 실행해 6행을 확인했다.
- 명확화 필요 표식: 0개. plan.md §B.1 의 세 표식을 모두 판정 기록으로 바꿨고, 대괄호 표식 문자열을 SPEC 디렉터리 전체에서 grep 한 결과가 비어 있음을 커밋 전에 확인했다.
- **Kickoff Approval: 아직 요청 전.** 리드 지시에 따라 t1099 가 develop 에 착지한 뒤 요청하며, 그때 spec.md §F.1 의 2(심볼 diff)·3(codex 버전, `runAgentHook` 매핑, 설정 파일 형식)으로 설계 전제를 다시 확인한다.
- 남은 미측정 전제: Claude 호스트의 JSON block Stop 상한(plan.md 추적 항목 Q8, 독트린 근거만), 사용자 범위 `~/.claude/settings.json` `env` 의 전파(잔여 면), Codex 가 `moai hook agent` 를 부르는 경로의 존재, 파일 부재를 「선언 없음」으로 보는 해석(REQ-HSF-016)이 운영자 의도와 일치하는지.

### 개정 0.3.1 (2026-09-24)

- 계기: plan-audit 2회차 FAIL 0.86 (`.moai/reports/t1152/plan-audit-iter2.md`, 로컬 증거). 차단 결함 N1(major)·N2·N3(minor)과 선택 결함 N4·N7 을 반영했다. N5·N6 은 기록만 요구해 손대지 않았다. 처분 요약은 spec.md HISTORY 0.3.1 행.
- 운영자 판정(2026-09-24, t1152 레인 AskUserQuestion): (1) N1~N3 을 고친 뒤 그 결함만 보는 3회차 감사로 진행한다. (2) 파일 부재 = 「선언 없음」 해석을 확정한다 — 탈출 장치를 셸 환경 변수로 인정할 수 있다. 다만 「추가 후 제거」 우회를 잔여 위험으로 기록하고 run 시작 시 호스트 동작을 재며, 호스트가 제거된 `env` 값을 훅 환경에 유지하면 운영자에게 돌아가 재판정을 받는다. 기록 위치: plan.md §B.1 Q1 「부재 해석 확인」, spec.md §F.2 「추가 후 제거」 행, plan.md §C Pre-flight 5.
- 개정 기준 트리: 브랜치 `WT-hook-stdin-failclosed`, HEAD `f5c3d9aa8`. t1099 인용은 계속 `fabc33812` 고정(이번 개정은 t1099 인용을 새로 만들거나 고치지 않았다).
- 이번 개정에서 이 트리에서 다시 확인한 것: `grep -n harnessModeIsCodex internal/cli/*.go` 의 비테스트 호출처는 `internal/cli/hook.go:292` 하나(정의 `hook_harness_codex.go:30`); `internal/cli/hook.go:45`(`--harness` PersistentFlags), `:455-463`(`runAgentHook` 파싱 실패 분기, `--harness` 판독 없음), `:469-481`(매핑 switch); 배포 래퍼 `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh:47`·`.sh.tmpl:47`·로컬 `.claude/hooks/moai/handle-agent-hook.sh:47` 모두 `exec moai hook agent "$1"`(`--harness` 없음); `internal/template/templates/.codex/` 는 `agents/` 만 담고 `moai hook`·`--harness` 문자열이 없다; `internal/cli/codex_readiness.go:74` `codexHarnessCommand = "moai hook --harness codex"`.
- 직접 측정하지 않고 plan-audit 에서 인용한 것: 유효한 stdin + `--harness bogus` 로 `moai hook agent x-validation` 이 exit 0 + PreToolUse allow 를 낸다는 재현(codex 백엔드 실행). Pre-flight 4(c) 에서 다시 잰다.
- 측정하지 않은 전제(새로 명시): 호스트가 설정 `env` 를 언제 훅 환경으로 전파하고 키·파일 삭제 뒤에도 유지하는지 — 「추가 후 제거」 우회의 성립 조건. run 시작 시 plan.md §C Pre-flight 5 가 선언된 상한(실행 4회, 실행당 `--max-turns 8`, 벽시계 300초) 안에서 잰다. 이번 개정은 라이브 모델·에이전트 실행을 하지 않았다.
- 0.3.0 의 미측정 목록에서 닫힌 것: 「파일 부재를 선언 없음으로 보는 해석이 운영자 의도와 일치하는지」 — 운영자가 확인했다.
- REQ 수 16(Tier M 상한 16, N4 는 REQ 안에서 절을 나눠 수를 늘리지 않았다). AC 수 13 + GATE. 명확화 필요 표식 0개(커밋 전 grep).
- **Kickoff Approval: 여전히 요청 전.** t1099 착지 후 요청하며, run 시작 시 Pre-flight 5 결과가 「유지한다」 또는 「미측정」이면 구현 전에 멈춘다.

### 개정 0.3.2 (2026-09-24)

- 계기: plan-audit 3회차 FAIL 0.88 (`.moai/reports/t1152/plan-audit-iter3.md`, 로컬 증거). 차단 결함 N8(major)·N9·N10(minor)과 선택 결함 N11·N12 를 반영했다. 처분 요약은 spec.md HISTORY 0.3.2 행. 1·2회차에서 닫힌 항목(D0~D13, N1~N7)은 다시 열지 않았다.
- **운영자 판정(2026-09-24, t1152 레인 AskUserQuestion): 3회차 하드 상한을 넘어 4회차 감사로 연장한다.** 이 개정은 그 4회차를 위한 것이다.
- 개정 기준 트리: 브랜치 `WT-hook-stdin-failclosed`, HEAD `92bdf4a8f`. t1099 인용은 계속 `fabc33812` 고정(이번 개정은 t1099 인용을 새로 만들거나 고치지 않았다).
- N8: plan.md §C Pre-flight 5 를 다시 짰다. 삭제 뒤 관측은 같은 실행에서 삭제 전에 설정 출처의 값이 훅 환경에서 관측된 경우에만 판정에 쓴다. 같은 세션 전파(A1 키 삭제·A2 파일 삭제)와 세션 시작 전파(B1·B2)를 모두 재고, 판정은 값을 실제로 전달한 경로에서만 내린다. 어느 경로로도 값이 오지 않으면 「전달 없음」으로 따로 닫는다(진행, 버전에 묶인 관측으로 기록, 출처 제약 유지). 「유지한다」·「미측정」이면 정지하고 운영자 재판정.
- N9 — **이번 개정이 선언하는 상한(제안값, manager-spec 작성, 미승인 — Kickoff 때 운영자가 승인한다)**: 실행 1회 = `claude -p` 프로세스 호출 하나(이어 받는 세션 포함). 예산 최대 6회(양성 대조 1, A1 1, A2 1, B1 1, B2 1, 재시도 1; A2 를 건너뛰면 5회). 실행마다 `timeout -k 10 300 claude -p --max-turns 8 …`. `--max-turns` 가 거부되면 그 실행은 예산 1회를 쓴 「미측정」이고 이후 실행은 벽시계 상한만으로 건다. 0.3.1 의 「실행 4회」 선언을 대체한다(0.3.1 기록은 그대로 둔다).
- 이번 개정에서 이 호스트에서 직접 관측한 것(모두 `--help`·`--version`·바이너리 판독만, 모델 실행 없음): `command -v timeout gtimeout` → `/opt/homebrew/bin/timeout`, `/opt/homebrew/bin/gtimeout`; `timeout --version` → `timeout (GNU coreutils) 9.9`; `timeout --help` 에 `-k, --kill-after=DURATION`; `claude --version` → `2.1.281 (Claude Code)`; `claude --help | grep -cE 'max-turns'` → `0`; `readlink -f /Users/goos/.local/bin/claude` → `/Users/goos/.local/share/claude/versions/2.1.281`; 그 바이너리에 `grep -a -o -E -- '"--max-turns <turns>"'` → `"--max-turns <turns>"`.
- N10: AC-HSF-013 에 유효한 stdin + `--harness bogus` 로 `x-verification`·`x-completion` 의 거부 단언과 RED 변이(하네스 판독을 결정 매핑에만 두는 변이)를 더하고, Pre-flight 4(c) 기준선에 같은 입력을 넣었다. 수리 전 트리에서 exit 0 이라는 예상은 측정하지 않았다 — Pre-flight 4(c) 에서 잰다.
- N11: AC-HSF-003(b1) 에 「판정 호출 결과가 fail-closed 분기 조건식에 쓰인다」를 더하고 인자 조건을 부정 조건으로 좁혔으며, 변이 (c5)(`_ = codexadapter.IsDecisionBearing(event)` + `event != hook.EventPostToolUse && event != hook.EventSubagentStop` 분기)를 더했다. 식별자 존재 확인: `internal/hook/types.go:28`·`:37`. 이 변이가 다른 기준을 모두 통과한다는 것은 문언 대조로 추론했고 실행하지 않았다.
- N12: AC-HSF-004·012·013 에 `deps.HookProtocol` 스파이의 `ReadInput` 호출 0 회를 직접 관측으로 더했다. 두 진입점이 모두 `deps.HookProtocol.ReadInput(os.Stdin)` 을 부른다는 것을 이 트리에서 확인했다(`internal/cli/hook.go:272`, `:455`).
- REQ 수 16(Tier M 상한 16) 유지. AC 수 13 + GATE 유지. 명확화 필요 표식 0개(커밋 전 grep).
- **Kickoff Approval: 여전히 요청 전.** t1099 착지 후 요청하며, 그때 Pre-flight 5 의 상한을 승인 대상으로 함께 올린다.

### 개정 0.3.3 (2026-09-24)

- 계기: plan-audit 4회차 FAIL 0.89 (`.moai/reports/t1152/plan-audit-iter4.md`, 로컬 증거). 차단 결함 N13(major)·N14(minor)만 고쳤다. 처분 요약은 spec.md HISTORY 0.3.3 행. 이전 회차에서 닫힌 항목은 다시 열지 않았다.
- **운영자 판정(2026-09-24, t1152 레인 AskUserQuestion): N13·N14 를 고친 뒤, 그 두 결함을 고친 문장만 대조하는 좁은 5회차 확인 감사로 진행한다.**
- 개정 기준 트리: 브랜치 `WT-hook-stdin-failclosed`, HEAD `203809812`.
- N13: plan.md §C Pre-flight 5 판정 전제에 삭제 뒤 기록 요건을 더했다 — 선언 여부 필드가 선언이 사라졌음을 보이는 훅 기록이 한 줄 이상 있어야 하며, 이 필드가 삭제 실행의 확인이다. 삭제가 실행되지 않았거나 삭제 뒤 기록이 0건인 변형은 「미측정」. 편집 거부를 줄이려고 실행을 `--permission-mode bypassPermissions` 로 띄운다(격리 `/tmp` 프로젝트, 삭제 주체는 위협 모형대로 모델). spec.md §F.2 「추가 후 제거」 행, acceptance.md §E 완료 조건을 맞췄다.
- N14: 판정 2의 셋째 절을 「A·B 가운데 어느 한 경로라도 확정되지 않았다」로 고치고, 판정 3에 「두 경로가 모두 확정됐고」를, 판정 4에 「두 경로가 모두 「전달 없음」으로 확정됐다」를 넣었다. 경로의 확정(「전달」·「전달 없음」·「미확정」)을 판정 전제 안에서 정의했다.
- 열린 선택 결함(이번 개정에서 손대지 않음): N15(거부 단언이 8개 agent action 가운데 3개만 표본), N16(`runAgentHook` 파싱 실패 분기에서 판정 결과를 뒤집는 동작 변이 부재), N17(기록형 훅의 등록 이벤트 미기재, 결과가 `-p` 형태에 묶인다는 한정 미기재 — 권한 모드 부분은 N13 수정으로 정했다).
- 이번 개정은 라이브 모델·에이전트 실행을 하지 않았다. REQ 수 16, AC 수 13 + GATE 유지.
- **Kickoff Approval: 여전히 요청 전.**

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
