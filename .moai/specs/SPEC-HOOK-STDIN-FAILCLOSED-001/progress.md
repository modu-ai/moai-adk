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

### Kickoff 판정 (2026-09-24)

- 출처: t1152 레인에서 운영자가 AskUserQuestion 에 직접 답했다(2026-09-24). 리드가 앞서 보낸 「Kickoff 는 이 지시로 갈음」 전달문은 리드가 철회했으며, 이 판정의 출처가 아니다.
- 판정 대상 트리: 브랜치 `WT-hook-stdin-failclosed`, HEAD `6cad9840f`(0.3.3, plan-audit 5회차 PASS-WITH-DEBT 0.91 — `.moai/reports/t1152/plan-audit-iter5.md`, 로컬 증거).
- 운영자 답:
  - (a) 측정을 시작하기 전에 N17·N18 을 닫는다 — 0.3.4 개정(아래)에서 닫았다.
  - (b) 측정 조건을 제안대로 승인한다 — 최대 6회, 실행당 `timeout -k 10 300` 과 `--max-turns 8`, 격리 `/tmp` 프로젝트에서 `--permission-mode bypassPermissions`, 「유지한다」 또는 「미측정」이면 멈추고 다시 판정받는다.
  - (c) N15·N16 은 run 에서 테스트를 더해 닫는다 — `--harness bogus` 거부를 agent action 8개 모두에 단언하고, `runAgentHook` 의 `IsDecisionBearing` 결과를 뒤집거나 무력화하는 변이가 AC-HSF-012 를 RED 로 만들어야 한다.
- 구현 커밋은 t1099 가 develop 에 착지하고 이 브랜치가 그것을 흡수한 뒤 시작한다(변경 없음). 그 시점의 전제 재확인은 spec.md §F.1 의 2·3 이다.
- 리드 조건(plan.md §C Pre-flight 5 「격리 확인」에 기록): 첫 실행 전에 증거 파일이 모든 실행의 명령 원문과 격리(작업 디렉터리, 건드리는 설정 파일, `HOME`·`CLAUDE_CONFIG_DIR` 류 변수)를 보여야 하며, 보이지 못하면 실행하지 않는다.

### 개정 0.3.4 (2026-09-24)

- 계기: 위 Kickoff 판정. 처분 요약은 spec.md HISTORY 0.3.4 행. 개정 기준 트리: HEAD `6cad9840f`.
- N17: plan.md §C Pre-flight 5 에 변형별 설정 파일 배치를 정했다. 기록형 훅은 모든 변형에서 `.claude/settings.json` 에 `PreToolUse`·`PostToolUse`(matcher `*`)로 등록하고 이 파일은 고치거나 지우지 않는다. 탈출 장치 키는 `.claude/settings.local.json` 에만 선언하고, A1·B1 은 그 키를, A2·B2 는 그 파일을 지운다. C 는 셸 환경에만 키를 둔다. 결과가 `settings.local.json` 면과 비대화형 `-p` 형태의 관측이라는 한정도 적었다.
- 설정 병합 전제의 확인 수준: 공식 문서 판독만(모델 실행 없음). <https://code.claude.com/docs/en/hooks> — 「Hook entries merge across settings levels rather than replacing each other」; <https://code.claude.com/docs/en/settings> 「When edits take effect」 — 설정 파일을 감시해 실행 중인 세션에 다시 적재한다. 이 호스트(Claude Code 2.1.281)에서 `settings.local.json` 삭제 뒤에도 `settings.json` 의 훅이 불리는지는 재지 않았다 — 성립하지 않으면 해당 변형은 「미측정」으로 떨어진다(판정 규칙 불변).
- N18: plan.md §F M0 의 승인 대상에 격리 `/tmp` 프로젝트에서의 `--permission-mode bypassPermissions` 를 넣고 승인됨으로 표시했다.
- 이번 개정은 라이브 모델·에이전트 실행을 하지 않았다. REQ 수 16, AC 수 13 + GATE 유지. 열린 선택 결함: N15·N16(run 에서 닫음, 위 판정 (c)).

### Pre-flight 5 재판정 (2026-09-25)

- 측정 판정: **「유지한다」 (plan.md §C 5 판정 1).** 증거 `.moai/reports/t1152/preflight5.md` §1·§2 (로컬 증거, gitignored). 측정 주체와 실행 경로: 가드 거부로 레인이 `claude -p` 를 직접 띄우지 못해 운영자가 실행기 `/tmp/t1152-pf5/run.sh` 를 실행했고, 레인이 결과 파일을 읽어 판정했다(증거 파일 §1). Claude Code 2.1.281, 비대화형 `claude -p`, 격리 `/tmp` 프로젝트, `--permission-mode bypassPermissions`, 실행당 `timeout -k 10 300`·`--max-turns 8`.
- 판정 근거(요지 — 원문 기록은 §E.2 「Pre-flight 5 결과」): 양성 대조 C 가 탐침을 봤고, A1 에서 키를 지운 뒤 기록 3줄(`local_declares:false`)이 탐침 `on` 이었다. 셸 환경에는 키가 없었다.
- 돌리지 않은 변형: A2·B1·B2. 판정 규칙상 판정 1 이 우선해 결론이 A1 하나로 섰기 때문이다. 예산 사용 2/6 — 모델 호출 0회로 끝난 인증 실패 4건은 2026-09-24 운영자 판정 두 건(첫 건, 이어 둘째·셋째 건)과 그 판정의 적용(2026-09-25 넷째 건)으로 예산에서 뺐다(증거 파일 §1, `/tmp/t1152-pf5/logs/runs.txt`).
- 정지 조건 발동: 구현을 시작하지 않고 운영자 재판정을 받았다.
- **운영자 재판정 (2026-09-25, t1152 레인 AskUserQuestion): 탈출 장치 없음 — 결정 이벤트는 파싱 실패 시 언제나 거부한다.** Codex Stop 면제는 그대로다. 지속적인 파싱 실패(예: 호스트 형식 변화)는 moai 갱신이나 훅 비활성화(`disableAllHooks`)로 복구하고 운영자 문서에 적는다. 거부 사유는 원인(stdin 파싱 실패)과 운영자 문서 식별자만 싣는다.

### 개정 0.4.0 (2026-09-25)

- 계기: 위 재판정. 처분 요약은 spec.md HISTORY 0.4.0 행. 개정 기준 트리: 브랜치 `WT-hook-stdin-failclosed`, HEAD `39ee312cf`.
- 삭제(번호 유지, 「삭제됨 (0.4.0, Pre-flight 5 결과)」 표시): REQ-HSF-007·016, AC-HSF-006 [RETIRED]·AC-HSF-010 [RETIRED].
- 개정: REQ-HSF-001(탈출 장치 조건절 제거, 실행 시점 스위치 금지 절 추가), REQ-HSF-010(사유는 원인·문서 식별자만, 복구 절차 미기재), REQ-HSF-011(탈출 적용 기록면 삭제), AC-HSF-001(Given 정리, (e4) 를 카나리·복구 절차 문구 부재로), AC-HSF-003((b3) AST 검사·(c6) 변이 신설), AC-HSF-007·011·012(Given·키 목록 정리), acceptance.md §1·§E, spec.md §D(Out of Scope — 탈출 장치 추가)·§F.1 3(c)·§F.2(잠김과 복구 경로, 유지 측정 결과, `disableAllHooks` 잔여 위험), plan.md §B.1 Q1 재판정·§C 5 완료 기록·§D·§E·M0·M1b·M2·M4·§G·§H.
- REQ: 16 번호 중 살아 있는 것 14(Tier M 상한 16). AC: 13 번호 중 살아 있는 것 11 + GATE.
- 이번 개정은 코드·테스트·라이브 모델 실행을 하지 않았다. 새로 생긴 전제: 파싱 실패 처리 경로가 환경 변수·설정 파일을 읽지 않는다는 AC-HSF-003(b3) 의 범위를 t1099 흡수 트리에서 먼저 잰다(spec.md §F.1 3(c)).

### 개정 0.4.1 (2026-09-25)

- 계기: plan-audit 6회차 FAIL 0.87(`.moai/reports/t1152/plan-audit-iter6.md`, HEAD `719c884df` 측정) — 차단 결함 N19·N20·N21, 선택 결함 N22·N23·N24. 반복 상한을 넘은 회차라 STOP 신호였다.
- **운영자 판정 (2026-09-25, t1152 레인 AskUserQuestion): 명시적 연장.** N19~N21(선택으로 N23)을 고친 뒤, N19~N21 만 보는 좁은 7회차 감사를 받는다.
- 처분: N19 — AC-HSF-003(b3) 에 기록 싱크 루트 해석 허용 목록을 명시(`resolveHookProjectRoot` 호출, 인자가 `config.EnvClaudeProjectDir` 인 `os.Getenv`; 이 트리 `internal/cli/hook.go:681-690` 에서 본문을 확인), 범위를 `internal/cli` 안 직접 호출의 전이적 폐포로 넓힘(N22 함께 닫음), spec.md §F.1 3(c) 를 맞춤. N20 — 삭제 AC 식별자 모든 출현 바로 뒤에 `[RETIRED]`, 짧은 별칭 5개를 정식 식별자로 풀어 씀. N21 — AC-HSF-001(e4) 를 사유 문자열 전체의 바이트 단위 일치로, (d)·AC-HSF-002(d) 의 기대 바이트를 테스트가 조립한 사유로 계산, 복구 안내 덧붙임 필수 RED 변이 추가, 카나리·`disableAllHooks` 부재를 (e5) 로. N23 — plan.md M0 의 미래형 Pre-flight 5 문장 삭제. N24(선택)는 열어 둔다.
- AC 계수(manager-docs § B12 계수기): 수정 전 `go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1 -v` → `absent-from-snapshot .moai/specs/SPEC-HOOK-STDIN-FAILCLOSED-001/acceptance.md: COUNT 18`; 수정 뒤 같은 명령 → `COUNT 11`, `--- PASS`. 선언(살아 있는 AC 11 + GATE — GATE 는 계수 패턴에 걸리지 않는다)과 일치.
- 이번 개정은 코드·테스트 코드·라이브 모델 실행을 하지 않았다. REQ·AC 수 변화 없음.

## §E.2 Run-phase Evidence

### Pre-flight 3·4 기준선 (구현 전, 2026-09-25)

범위: plan.md §C 의 3·4 항목만 수행했다. 1(t1099 착지 확인)·2(심볼 diff)·5(설정 `env` 전파 측정)는 이번에 하지 않았다. Go 소스·테스트·템플릿·규칙 파일은 건드리지 않았고, 라이브 모델·에이전트 실행(`claude -p`, `codex exec`)도 하지 않았다.

**[HARD] 이 기준선은 t1099 를 흡수하기 전에 잰 것이다.** t1099 는 `internal/cli/hook.go` 의 디스패치 오류 경로를 바꾸므로, run 착수 시 t1099 를 흡수한 병합 트리에서 같은 측정을 다시 하고 이 기록과 대조해야 한다. 이 표를 병합 트리의 수리 전 기준선으로 그대로 인용하지 않는다.

#### 측정 대상 빌드

- 트리 HEAD: `0aa6b14f0bbbd9a371e1dae97c2bc5b588441468` (브랜치 `WT-hook-stdin-failclosed`, 빌드 시점 `git status --short` 빈 출력)
- 빌드: 이 트리에서 `go build -o <scratch>/pf4/moai-t1152 ./cmd/moai` → `build_exit=0`. 툴체인 `go version go1.26.8 darwin/arm64`. 모든 측정은 이 바이너리를 경로로 직접 호출했다(설치본 `moai` 를 쓰지 않음, `verification-claim-integrity.md` §2.2). ldflags 없이 빌드했으므로 바이너리의 `version` 출력에는 커밋이 박혀 있지 않다 — 판정 빌드의 커밋은 위 HEAD 로 귀속한다.
- `<scratch>` = `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go--claude-worktrees-develop/a99d4f42-56e7-4b79-a6cc-04004cb1829e/scratchpad` (머신 로컬 스크래치 — 아래 판정에 쓴 명령과 출력은 이 파일에 옮겨 적었다).

#### Pre-flight 3 — codex-cli 버전

```
$ command -v codex; codex --version
/Users/goos/.local/bin/codex
codex-cli 0.156.1
```

Q2 측정 버전(0.156.1)과 같다. Q2 재측정은 필요 없고, REQ-HSF-013 술어·spec.md §B.2 행 3 개정 사유도 생기지 않았다. run 착수 시 버전이 바뀌어 있으면 그때 plan.md §C 3 에 따라 다시 잰다.

#### Pre-flight 4 — 측정 방법

- 각 호출마다 `<scratch>/pf4/proj.XXXXXX` 아래 새 임시 디렉터리를 만들고, 작업 디렉터리와 `CLAUDE_PROJECT_DIR`·`MOAI_PROJECT_DIR` 를 모두 그 디렉터리로 두었다. `HOME` 은 바꾸지 않았다. 셸에서 물려받은 `MOAI_KANBAN_BACKEND`·`MOAI_KANBAN_SETTINGS_INJECTED`·`MOAI_FACTORY_WORKER`·`MOAI_FACTORY_WORKERS` 는 러너 안에서 `unset` 했다.
- 호출 형태: `( cd "$d" && CLAUDE_PROJECT_DIR="$d" MOAI_PROJECT_DIR="$d" "$BIN" hook <하위 명령> [인자] < <페이로드> > <label>.stdout 2> <label>.stderr )`. 각 줄의 `rc` 는 그 종료 코드, `stdout(nB)` 는 바이트 수와 앞 200바이트, `stderr` 는 앞 220바이트, `canary_in_out` 은 stdout·stderr 에 `CANARY` 문자열이 있는지, `files_in_projdir` 는 호출 뒤 임시 디렉터리 안 파일 수다.
- 페이로드(acceptance.md §0 의 파손 4 형태, 스크래치 러너가 생성 — 인라인 heredoc 아님):
  - malformed: `{"broken":"CANARY-MALFORMED` (27 B)
  - truncated: `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"note":"CANARY-TRUNCATED",` — 카나리 뒤에서 자름, 닫는 중괄호 없음 (91 B)
  - oversize: `{"hook_event_name":"PreToolUse","tool_input":{"note":"CANARY-OVERSIZE","content":"` + `a` × 6 MiB + `"}}` (6,291,541 B)
  - depth: `{"hook_event_name":"PreToolUse","tool_input":{"note":"CANARY-DEPTH","x":` + `[`×10001 + `]`×10001 + `}}` (20,076 B)
  - depth9000(대조군, acceptance.md §0 의 run 착수 시 대조를 미리 한 번 잰 것): 같은 구성, 깊이 9000 (18,078 B)
  - valid: `{}` (2 B)
- 관측 한계: **디스패치 횟수는 바깥에서 관측할 수 없다.** 프로세스 경계에서 보이는 것은 stdout `{}` 와 stderr 의 `emitting default output` 경고뿐이며, 이것이 파싱 실패 분기(디스패치 없음)를 탔다는 관측 가능한 표지다. 디스패치 0회의 직접 관측은 run 의 스파이 레지스트리 테스트(AC-HSF-001 등) 몫이다.

#### Pre-flight 4 — 결과 (판정에 쓴 출력 원문, stdout `{}` 는 줄바꿈 포함 3 B)

(a) `pre-tool` × 파손 4 형태 + 깊이 9000 대조군:

```
a_pretool_malformed | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
a_pretool_truncated | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
a_pretool_oversize | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
a_pretool_depth | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: invalid character '[' exceeded max depth); emitting default output  | canary_in_out=no | files_in_projdir=0
a_pretool_depth9000 | rc=0 | stdout(83B)={"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}  | stderr= | canary_in_out=no | files_in_projdir=0
```

(b) 관측 하위 명령 22개 × malformed:

```
b_session-start | rc=0 | stdout(3B)={}  | stderr=moai hook SessionStart: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_post-tool | rc=0 | stdout(3B)={}  | stderr=moai hook PostToolUse: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_session-end | rc=0 | stdout(3B)={}  | stderr=moai hook SessionEnd: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_subagent-stop | rc=0 | stdout(3B)={}  | stderr=moai hook SubagentStop: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_compact | rc=0 | stdout(3B)={}  | stderr=moai hook PreCompact: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_post-tool-failure | rc=0 | stdout(3B)={}  | stderr=moai hook PostToolUseFailure: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_notification | rc=0 | stdout(3B)={}  | stderr=moai hook Notification: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_subagent-start | rc=0 | stdout(3B)={}  | stderr=moai hook SubagentStart: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_teammate-idle | rc=0 | stdout(3B)={}  | stderr=moai hook TeammateIdle: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_task-completed | rc=0 | stdout(3B)={}  | stderr=moai hook TaskCompleted: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_worktree-create | rc=0 | stdout(0B)= | stderr=moai hook WorktreeCreate: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_worktree-remove | rc=0 | stdout(0B)= | stderr=moai hook WorktreeRemove: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_post-compact | rc=0 | stdout(3B)={}  | stderr=moai hook PostCompact: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_instructions-loaded | rc=0 | stdout(3B)={}  | stderr=moai hook InstructionsLoaded: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_stop-failure | rc=0 | stdout(3B)={}  | stderr=moai hook StopFailure: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_config-change | rc=0 | stdout(3B)={}  | stderr=moai hook ConfigChange: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_task-created | rc=0 | stdout(3B)={}  | stderr=moai hook TaskCreated: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_cwd-changed | rc=0 | stdout(3B)={}  | stderr=moai hook CwdChanged: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_file-changed | rc=0 | stdout(3B)={}  | stderr=moai hook FileChanged: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_elicitation | rc=0 | stdout(3B)={}  | stderr=moai hook Elicitation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_elicitation-result | rc=0 | stdout(3B)={}  | stderr=moai hook ElicitationResult: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
b_permission-denied | rc=0 | stdout(3B)={}  | stderr=moai hook PermissionDenied: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
```

`pre-tool` 외 결정 이벤트 하위 명령 × malformed, 그리고 `pre-tool --harness codex` × malformed:

```
d_permission-request | rc=0 | stdout(3B)={}  | stderr=moai hook PermissionRequest: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
d_stop | rc=0 | stdout(3B)={}  | stderr=moai hook Stop: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
d_user-prompt-submit | rc=0 | stdout(3B)={}  | stderr=moai hook UserPromptSubmit: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
d_pretool_codex | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
```

(c) `hook agent <action>` × malformed, 그리고 유효한 stdin `{}` + `--harness bogus`:

```
c_agent_x-validation | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-validation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-pre-transformation | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-pre-transformation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-pre-implementation | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-pre-implementation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_foo | rc=0 | stdout(3B)={}  | stderr=moai hook agent foo: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-verification | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-verification: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-post-transformation | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-post-transformation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-post-implementation | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-post-implementation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-completion | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-completion: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_bogus_x-validation | rc=0 | stdout(83B)={"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}  | stderr= | canary_in_out=no | files_in_projdir=0
c_agent_bogus_x-verification | rc=0 | stdout(55B)={"hookSpecificOutput":{"hookEventName":"PostToolUse"}}  | stderr= | canary_in_out=no | files_in_projdir=0
c_agent_bogus_x-completion | rc=0 | stdout(3B)={}  | stderr= | canary_in_out=no | files_in_projdir=0
```

#### 계획의 예상과 대조

| 항목 | plan.md §C 4 의 예상 | 관측 | 일치 |
|---|---|---|---|
| (a) `pre-tool` 파손 4 형태 | stdout `{}`, exit 0, 디스패치 0 | 4 형태 모두 stdout `{}`·rc=0·stderr 파싱 실패 경고. 디스패치 횟수는 바깥에서 관측 불가(위 관측 한계) | 예(디스패치 0 은 표지로만) |
| (b) 관측 하위 명령 22개 | 20개 `{}`, worktree-create·worktree-remove 빈 stdout | 20개 `{}`, 두 개 0 B, 전부 rc=0 | 예 |
| (c) 결정 매핑 action 4종 × malformed | `{}` exit 0 | 4종 모두 `{}` rc=0 | 예 |
| (c) 관측 매핑 action 4종 × malformed | `{}` exit 0 | 4종 모두 `{}` rc=0 | 예 |
| (c) `x-validation` + `{}` + `--harness bogus` | exit 0 + PreToolUse allow (plan-audit 2회차 codex 재현) | rc=0, `permissionDecision:"allow"` | 예 — AC-HSF-012 신규 거부의 기준선 |
| (c) `x-verification`·`x-completion` + `{}` + `--harness bogus` | exit 0 | 둘 다 rc=0. `x-verification` 은 `{"hookSpecificOutput":{"hookEventName":"PostToolUse"}}`, `x-completion` 은 `{}` | 예 — AC-HSF-013 신규 거부의 기준선 |
| `permission-request`·`stop`·`user-prompt-submit` × malformed | (수리 대상 — 수리 전 `{}` 예상) | 셋 다 `{}` rc=0 | 예 |
| `pre-tool --harness codex` × malformed | `{}` (AC-HSF-004 가 적은 수리 전 값) | `{}` rc=0 | 예 |

예상에서 어긋난 것은 없다. 덧붙여 기록할 관측:

- malformed·truncated·oversize 세 형태의 stderr 원인 문자열이 모두 `unexpected end of JSON input` 으로 같다. oversize 는 5 MiB 상한에서 잘려 조기 종료로 보고된다(acceptance.md §0 의 서술과 일치). depth 만 `invalid character '[' exceeded max depth` 로 다르다.
- 깊이 9000 대조군은 파싱에 성공해 디스패치됐고(PreToolUse allow 출력, 경고 없음), depth 형태의 실패가 깊이 상한 때문임을 뒷받침한다. acceptance.md §0 은 이 대조를 `ReadInput` 수준에서 run 착수 시 재라고 하므로, 이 CLI 수준 관측은 그 대조를 대신하지 않는다.
- 모든 경우에서 카나리 문자열이 stdout·stderr 에 나타나지 않았다(수리 전 경고 문자열은 페이로드 원문을 싣지 않는다).
- `--harness bogus` 는 `hook agent` 에서 경고 없이 받아들여진다 — 수리 전 `runAgentHook` 에는 하네스 판정이 없다.

#### 격리 확인

- 모든 호출 뒤 임시 프로젝트 디렉터리 안 파일 수가 0이었다(위 각 줄 `files_in_projdir=0`).
- 측정 뒤 이 워크트리에서 `git status --short` → 빈 출력(저장소에 쓰인 것 없음).

#### 미측정·잔여

- 디스패치 횟수(위 관측 한계). run 의 스파이 테스트 몫.
- `--harness bogus` + malformed stdin 조합(AC-HSF-004·AC-HSF-012 의 「판정이 읽기보다 앞선다」 관측)은 plan.md §C 4 의 목록에 없어 재지 않았다.
- Pre-flight 1·2·5 는 이번 범위 밖이다. 특히 1(t1099 착지)이 성립하기 전에는 구현을 시작하지 않는다.

### Pre-flight 5 결과 (2026-09-25) — 「유지한다」

기록자: manager-spec (0.4.0 개정, 오케스트레이터 재위임). 측정 원문은 레인이 증거 파일에 옮겨 적은 것을 판독했고, 이 개정에서 측정을 다시 하지 않았다.

- 증거: `.moai/reports/t1152/preflight5.md` (로컬 증거, gitignored). 실행기와 훅 기록: `/tmp/t1152-pf5/` (머신 로컬 스크래치 — 판정에 쓴 줄은 아래에 옮겼다).
- 선언 상한(승인됨): 최대 6회, 실행당 `timeout -k 10 300`·`--max-turns 8`, 모델 `haiku`, 격리 `HOME=/tmp/t1152-pf5/home`·`CLAUDE_CONFIG_DIR=/tmp/t1152-pf5/cfg`, `env -i` 로 상속 환경 제거. 실제 사용: 2/6 (C 1, A1 1). 인증 실패 4건(모델 호출 0회)은 운영자 판정으로 예산 제외.
- 격리 차이(증거 파일 §1 기록): 격리 `HOME` 에 macOS 키체인 경로 심볼릭 링크 하나(`/tmp/t1152-pf5/home/Library/Keychains` → 사용자 키체인)를 더했다. 자격 증명 저장소이며 설정 파일이 아니다.

C (양성 대조, 셸 환경에만 키) — 11:57:48–11:58:00, exit 0:

```
{"ts":"11:57:57","event":"PreToolUse","tool":"Bash","probe":"on","local_exists":true,"local_declares":false}
{"ts":"11:57:58","event":"PostToolUse","tool":"Bash","probe":"on","local_exists":true,"local_declares":false}
```

A1 (같은 세션: `settings.local.json` `env` 에 키 추가 → 키 삭제, 셸 환경에는 키 없음) — 11:59:46–12:00:09, exit 0:

```
{"ts":"11:59:52","event":"PreToolUse","tool":"Bash","probe":"","local_exists":true,"local_declares":false}
{"ts":"11:59:56","event":"PostToolUse","tool":"Bash","probe":"on","local_exists":true,"local_declares":true}
{"ts":"11:59:58","event":"PreToolUse","tool":"Bash","probe":"on","local_exists":true,"local_declares":true}
{"ts":"11:59:58","event":"PostToolUse","tool":"Bash","probe":"on","local_exists":true,"local_declares":true}
{"ts":"12:00:01","event":"PreToolUse","tool":"Bash","probe":"on","local_exists":true,"local_declares":true}
{"ts":"12:00:04","event":"PostToolUse","tool":"Bash","probe":"on","local_exists":true,"local_declares":false}
{"ts":"12:00:06","event":"PreToolUse","tool":"Bash","probe":"on","local_exists":true,"local_declares":false}
{"ts":"12:00:06","event":"PostToolUse","tool":"Bash","probe":"on","local_exists":true,"local_declares":false}
```

- 판정 전제: 삭제 전 관측 충족(선언 중 탐침 `on` 4줄), 삭제 뒤 기록 충족(`local_declares:false` 3줄). 경로 A = 「전달」.
- 판정: **1 「유지한다」** — 키를 지운 뒤에도 그 세션의 훅 환경에 값이 남았다. A2·B1·B2 는 판정 1 이 우선해 돌리지 않았다.
- 처분: 정지 → 운영자 재판정(§E.1 「Pre-flight 5 재판정」) → spec 0.4.0 에서 탈출 장치 제거.
- 한계: 이 호스트·Claude Code 2.1.281·비대화형 `-p`·`settings.local.json` 면의 키 삭제 변형 한 번의 관측이다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
