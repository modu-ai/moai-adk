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

### Run 착수 전제 재확인 (2026-09-26, 병합 트리 `5fcc615a9`)

기록자: manager-develop (card t1152, run-phase). 측정 트리: 브랜치 `WT-hook-stdin-failclosed`, HEAD `5fcc615a94af7f1a5a3f23a03c7747c36c4fb3ab`, `git status --short` 빈 출력.

#### spec.md §F.1 의 1 — t1099 흡수 형태

- **리드 판정으로 형태를 바꿔 충족했다.** spec.md §F.1 의 1 은 「t1099 가 develop 에 착지하고 그 develop 을 이 브랜치가 흡수한 뒤」를 요구한다. 리드가 2026-09-26 cross-session 메시지로 t1099 전체 착지를 기다리지 않고 M2a~M2c 끝 커밋 `c2d06a518` 을 흡수하도록 판정했고, 병합 커밋 `5fcc615a9` 가 그것을 흡수했다. 증거: `.moai/reports/t1152/t1099-partial-absorb.md` (로컬 증거). 따라서 이 전제는 「t1099 가 develop 에 있다」가 아니라 「t1099 M2a~M2c 를 병합 `5fcc615a9` 로 흡수했다」로 충족됐다.
- `git merge-base --is-ancestor c2d06a518 HEAD` → exit 0 (흡수됨).
- `git merge-base --is-ancestor fabc33812 HEAD` → exit 1. `fabc33812` 은 `c2d06a518` 뒤의 M2c 문서 마감 커밋이라 조상이 아니다. 인용 대상 코드 파일이 같은지는 아래 blob 비교로 본다.
- 잔여 위험(판정서 인용): t1099 가 develop 에 착지하기 전에 이 카드가 병합 창을 받으면 t1099 `c2d06a518` 까지의 조상 전부가 이 카드를 통해 develop 에 먼저 들어간다. 리드에게 창 요청 전에 알린다.

#### spec.md §F.1 의 2 — 인용 심볼 diff

```
$ git diff --stat fabc33812 HEAD -- internal/codexadapter/decision.go internal/codexadapter/translate.go internal/codexadapter/diagnostics.go internal/codexadapter/output.go internal/cli/hook_codex_failclosed.go internal/cli/hook.go
(출력 없음, exit 0)
```

양성 대조(같은 두 끝점의 전체 diff 는 비어 있지 않다):

```
$ git diff --shortstat fabc33812 HEAD
 214 files changed, 31658 insertions(+), 555 deletions(-)
```

blob 대조 — 두 트리의 여섯 파일 blob 이 모두 같다(`git ls-tree fabc33812 …` 와 `git ls-tree HEAD …` 출력이 줄마다 일치):

```
100644 blob 6df9dc72dae630ac7e1c5f46e929f046e2407e2d	internal/cli/hook.go
100644 blob 7061bf45583513bd7e81cbf241223cf0b3da0e5f	internal/cli/hook_codex_failclosed.go
100644 blob 49223d76f5e60bbd2bffe489c897b7f53d92f537	internal/codexadapter/decision.go
100644 blob 0b324d67427e61c1c03a0d0e9a27212d106b57ee	internal/codexadapter/diagnostics.go
100644 blob cc15e2b9c785de222274c3b3ee6663b4cd4cb4e9	internal/codexadapter/output.go
100644 blob 2736fe6fcebf2d4b732eb17f57aab9bbfa2b636b	internal/codexadapter/translate.go
```

AC-HSF-009 항목별 판정 — blob 이 같으므로 spec.md §A.4·§B.4 가 `fabc33812` 에서 인용한 내용과 모두 **동일**:

| 항목 | 판정 | 근거(이 트리 판독) |
|---|---|---|
| (a) `DecisionBearingEvents()` 원소 | 동일 | `decision.go:78` — PreToolUse, PermissionRequest, Stop, UserPromptSubmit |
| (b) fatal_error 열 Outcome | 동일 | 두 하네스 네 이벤트 모두 다섯째 칸 `row(OutcomeDeny)` |
| (c) `writeCodexFailClosed` 서명·기록 키 | 동일 | `writeCodexFailClosed(event hook.EventType, cause error) error`, 키 `hookFaultDiscardKey = "hook-fault"` |
| (d) 네 이벤트의 `Render` 형태 | 동일 | `decision.go` `Render` 본문 — PreToolUse `permissionDecision`, PermissionRequest `decision.behavior`/`message`, Stop·UserPromptSubmit `decision:"block"`/`reason` |
| (e) `Discard` 필드와 stderr 미러 | 동일 | `Discard{Event, Key, ContentLength, Reason}`, 미러 `codex-adapter: dropped %q on %s (%d bytes): %s` |

spec 개정 필요 없음(D-NEW-1 경로 미발동).

#### spec.md §F.1 의 3 — 설계 전제

(a) codex-cli 버전:

```
$ codex --version; command -v codex
codex-cli 0.157.0
/Users/goos/.local/bin/codex
```

**Q2 측정 버전(0.156.1)과 다르다.** plan.md §C 3 은 버전이 다르면 Q2 를 같은 방법으로 다시 재라고 한다. Q2 재측정은 `codex exec` 라이브 모델 실행이고, 이번 위임은 그 실행의 상한(turn·벽시계)을 선언·승인받지 않았으므로 **돌리지 않았다** — 미측정(Gap). 바이너리 문자열만 판독했다(측정 아님, 가설 수준): `grep -a -o -E 'stop_hook[a-z_]{0,40}'` → `stop_hook_active`·`stop_hooks` 만, `block_cap|max_stop|stop_block|consecutive_block` 패턴 0건. 면제는 REQ-HSF-013 의 술어 하나에 모여 있으므로, 재측정에서 상한이 나오면 그 술어와 spec.md §B.2 행 3 만 다시 보면 된다. 재측정 여부는 리드 판정 사항으로 올린다.

(b) `runAgentHook` 의 action → 이벤트 매핑: 이 트리 `internal/cli/hook.go:476-486` 의 switch 가 spec 인용(`44dfc25fd` 기준 `:469-481`)과 **내용이 같다**(줄 번호만 이동). 접미사 `-validation`·`-pre-transformation`·`-pre-implementation` → `hook.EventPreToolUse`, `-verification`·`-post-transformation`·`-post-implementation` → `hook.EventPostToolUse`, `-completion` → `hook.EventSubagentStop`, `default` → `hook.EventPreToolUse`.

(c) AC-HSF-003(b3) 범위의 판독(흡수 트리, 구현 전):

- `runHookEvent` 파싱 실패 분기(`hook.go:273-280`): `fmt.Fprintf(os.Stderr, …)`, `writeHookOutput(event, nil, …)`. `writeHookOutput`(`hook.go:394-415`) 본문: `fmt.Fprintln(os.Stdout, …)` 와 메서드 `deps.HookProtocol.WriteOutput` — 환경·파일 판독 없음.
- `runAgentHook` 파싱 실패 분기(`hook.go:462-469`): `fmt.Fprintf`, 메서드 `WriteOutput` — 판독 없음.
- `writeCodexFailClosed` 본문(`hook_codex_failclosed.go:34-58`): `codexadapter.TranslateCodex`·`codexadapter.RecordDiscards`(다른 패키지 — 범위 밖), `resolveHookProjectRoot()`(허용 목록 (i)), `os.Stdout.Write`, `fmt.Fprintf`.
- `resolveHookProjectRoot`(`hook.go:687-696`): `os.Getenv(config.EnvClaudeProjectDir)` + `os.Getwd` — 허용 목록 (i) 로 본문에 들어가지 않는다.
- 판정: **허용 목록 밖 판독 없음.** 정지 조건 미발동.

#### 기존 테스트 영향 검색 (acceptance.md §D 재실행)

`grep -rn 'broken\|invalid JSON\|not valid json' internal/cli/*_test.go` → 108행(`… | wc -l` → `108`). 전체를 판독해 훅 경로에 해당하는 것만 추리면:

```
internal/cli/coverage_test.go:71:				return nil, errors.New("invalid JSON")
internal/cli/hook_protocol_fix_test.go:71:	swapStdinString(t, `{"broken`)
internal/cli/hook_harness_classify_test.go:155:		`{this is not valid json`,
internal/cli/multi_review_gate_wiring_test.go:7://   - the cobra RunE fail-OPEN contract (a broken stdin never traps Stop),
```

`coverage_test.go:71` 은 `TestRunHookEvent_ReadInputError`(`post-tool`, 관측 — 변경 불필요), `hook_protocol_fix_test.go:71` 은 갱신 대상 `TestRunHookEvent_MalformedStdinGraceful`, 나머지 둘은 `harness-classify`·`multi-review-gate` 자체 stdin 처리로 spec.md §D 범위 밖이다. 결정 이벤트에 파손 stdin 을 넣고 `{}` 를 단언하는 새 테스트는 흡수 트리에 없다. acceptance.md §D 표 밖의 `TestRunAgentHook_ReadInputError`(`misc_coverage_test.go:355`)는 그대로 있다.

#### AC-HSF-003(b) 기준선 grep (흡수 트리)

```
$ grep -rnE 'hook\.Event(PreToolUse|PermissionRequest|Stop|UserPromptSubmit)\b' internal/cli --include='*.go' | grep -v _test.go
internal/cli/hook.go:54:		{"pre-tool", "Handle pre-tool-use event", hook.EventPreToolUse},
internal/cli/hook.go:57:		{"stop", "Handle stop event", hook.EventStop},
internal/cli/hook.go:62:		{"user-prompt-submit", "Handle user prompt submit event", hook.EventUserPromptSubmit},
internal/cli/hook.go:63:		{"permission-request", "Handle permission request event", hook.EventPermissionRequest},
internal/cli/hook.go:478:		event = hook.EventPreToolUse
internal/cli/hook.go:485:		event = hook.EventPreToolUse
```

6행 — 하위 명령 표 4행 + 매핑 대입 2행, spec 기준선과 같은 구성(대입 줄 번호만 `:472`·`:479` → `:478`·`:485`). t1099 흡수가 네 식별자의 새 출현을 더하지 않았다.

### Pre-flight 4 기준선 재측정 (병합 트리 `5fcc615a9`, 구현 전, 2026-09-26)

**이 표가 수리 전 기준선이다.** 위 「Pre-flight 3·4 기준선 (2026-09-25)」은 t1099 흡수 전 트리(`0aa6b14f0`)를 쟀으므로 대조용으로만 남긴다. 이 기준선은 수리 커밋보다 앞선 커밋에 들어간다(`verification-claim-integrity.md` §2.3).

#### 측정 대상 빌드

- 트리 HEAD `5fcc615a94af7f1a5a3f23a03c7747c36c4fb3ab`, 빌드 직전 `git status --short` 빈 출력.
- 빌드: `go build -o <scratch>/pf4/moai-t1152 ./cmd/moai` → 오류 출력 없이 종료. 모든 호출은 이 바이너리를 경로로 직접 불렀다(설치본 미사용, `verification-claim-integrity.md` §2.2). ldflags 없이 빌드했으므로 판정 빌드의 커밋은 위 HEAD 로 귀속한다.
- `<scratch>` = `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t1152/a3f6b5c6-2d81-4d44-a4fe-0c85054f0469/scratchpad` (머신 로컬 — 판정에 쓴 줄은 아래에 옮겼다). 러너 `pf4/run.sh`.

#### 측정 방법

2026-09-25 기준선과 같다(각 호출마다 새 임시 프로젝트 디렉터리, 그곳을 작업 디렉터리·`CLAUDE_PROJECT_DIR`·`MOAI_PROJECT_DIR` 로, `MOAI_KANBAN_BACKEND`·`MOAI_KANBAN_SETTINGS_INJECTED`·`MOAI_FACTORY_WORKER`·`MOAI_FACTORY_WORKERS` 는 러너 안에서 `unset`). 페이로드는 러너가 새로 만들었다 — malformed 27 B, truncated 91 B, oversize 6,291,541 B, depth(깊이 10001) 20,076 B, depth9000 18,078 B, valid `{}` 2 B. 이번에 더한 것: 결정 하위 명령 셋의 `--harness codex` 형태, agent action 8종 모두의 유효 stdin + `--harness bogus`(N15), `--harness bogus` + malformed 조합 두 건(종전 미측정 항목). 디스패치 횟수는 프로세스 밖에서 관측할 수 없다(종전과 같은 관측 한계).

#### 결과 (판정에 쓴 출력 원문)

```
## (a) pre-tool x 4 forms + depth9000
a_pretool_malformed | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
a_pretool_truncated | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
a_pretool_oversize | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
a_pretool_depth | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: invalid character '[' exceeded max depth); emitting default output  | canary_in_out=no | files_in_projdir=0
a_pretool_depth9000 | rc=0 | stdout(83B)={"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}  | stderr= | canary_in_out=no | files_in_projdir=0
## (b) observation subcommands x malformed
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
## (d) other decision subcommands x malformed, both harness modes
d_permission-request | rc=0 | stdout(3B)={}  | stderr=moai hook PermissionRequest: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
d_permission-request_codex | rc=0 | stdout(3B)={}  | stderr=moai hook PermissionRequest: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
d_stop | rc=0 | stdout(3B)={}  | stderr=moai hook Stop: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
d_stop_codex | rc=0 | stdout(3B)={}  | stderr=moai hook Stop: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
d_user-prompt-submit | rc=0 | stdout(3B)={}  | stderr=moai hook UserPromptSubmit: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
d_user-prompt-submit_codex | rc=0 | stdout(3B)={}  | stderr=moai hook UserPromptSubmit: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
d_pretool_codex | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
## (c) agent actions x malformed
c_agent_x-validation | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-validation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-pre-transformation | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-pre-transformation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-pre-implementation | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-pre-implementation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_foo | rc=0 | stdout(3B)={}  | stderr=moai hook agent foo: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-verification | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-verification: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-post-transformation | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-post-transformation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-post-implementation | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-post-implementation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
c_agent_x-completion | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-completion: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
## (c) agent valid stdin + --harness bogus (all 8 actions)
c_agent_bogus_x-validation | rc=0 | stdout(83B)={"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}  | stderr= | canary_in_out=no | files_in_projdir=0
c_agent_bogus_x-pre-transformation | rc=0 | stdout(83B)={"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}  | stderr= | canary_in_out=no | files_in_projdir=0
c_agent_bogus_x-pre-implementation | rc=0 | stdout(83B)={"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}  | stderr= | canary_in_out=no | files_in_projdir=0
c_agent_bogus_foo | rc=0 | stdout(83B)={"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}  | stderr= | canary_in_out=no | files_in_projdir=0
c_agent_bogus_x-verification | rc=0 | stdout(55B)={"hookSpecificOutput":{"hookEventName":"PostToolUse"}}  | stderr= | canary_in_out=no | files_in_projdir=0
c_agent_bogus_x-post-transformation | rc=0 | stdout(55B)={"hookSpecificOutput":{"hookEventName":"PostToolUse"}}  | stderr= | canary_in_out=no | files_in_projdir=0
c_agent_bogus_x-post-implementation | rc=0 | stdout(55B)={"hookSpecificOutput":{"hookEventName":"PostToolUse"}}  | stderr= | canary_in_out=no | files_in_projdir=0
c_agent_bogus_x-completion | rc=0 | stdout(3B)={}  | stderr= | canary_in_out=no | files_in_projdir=0
## (e) --harness bogus + malformed (AC-HSF-004/012 ordering baseline)
e_pretool_bogus_malformed | rc=0 | stdout(3B)={}  | stderr=moai hook PreToolUse: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
e_agent_bogus_malformed | rc=0 | stdout(3B)={}  | stderr=moai hook agent x-validation: invalid stdin JSON (hook: invalid JSON input: unexpected end of JSON input); emitting default output  | canary_in_out=no | files_in_projdir=0
```

#### 2026-09-25 기준선과 대조

- 두 기준선에 공통인 행은 모두 같은 값이다 — t1099 흡수가 파싱 실패 경로의 출력을 바꾸지 않았다(흡수된 M2c 는 디스패치 이후 경로만 건드린다).
- 새로 잰 것: (1) `--harness codex` 결정 하위 명령 셋이 모두 `{}` rc=0 — AC-HSF-002 의 수리 전 RED 근거. (2) agent 8 action 전부 유효 stdin + `--harness bogus` 에서 rc=0 — 결정 매핑 넷은 PreToolUse allow, 관측 매핑 셋은 PostToolUse 빈 형태, `x-completion` 은 `{}`. N15 의 수리 전 기준선. (3) **`--harness bogus` + malformed 가 두 진입점 모두에서 rc=0 + stdin 경고** — 하네스 판정이 stdin 파싱 뒤에 있다는 수리 전 관측이다(AC-HSF-004·012 의 RED 근거). `pre-tool` 에서는 유효 stdin 이면 `invalid --harness value` 로 거부되지만(hook.go:292), 파싱 실패 분기가 그보다 먼저 반환한다.
- 카나리는 모든 경우에 stdout·stderr 에 나타나지 않았고, 모든 임시 프로젝트 디렉터리의 파일 수가 0이었다.

#### 미측정·잔여

- 디스패치 횟수(프로세스 밖 관측 불가 — 스파이 테스트 몫).
- depth 대조군의 `ReadInput` 수준 측정(acceptance.md §0)은 구현 테스트에서 한다. 여기서는 CLI 수준(깊이 9000 이 파싱되어 디스패치됨)만 봤다.
- codex-cli 0.157.0 에서의 Q2 재측정(위 §F.1 3(a)).

### 구현 (M1~M4, 2026-09-26)

- 기준선 커밋 `9b6296a63`(위 두 절) → 구현 커밋 `53ec0d5a8`. 기준선이 수리보다 앞선 커밋에 있다(`verification-claim-integrity.md` §2.3).
- 변경 파일: `internal/codexadapter/stop_cap.go`(술어 `HostLacksStopBlockCap`, 신규), `internal/codexadapter/stop_cap_test.go`(신규), `internal/codexadapter/translate.go`(`@MX:ANCHOR` fan_in 3→4 주석만), `internal/cli/hook_stdin_failclosed.go`(신규 — 공유 파싱 실패 처리 `answerStdinParseFailure`, 작성기 `writeFailClosedDeny`, 사유·기록 키 상수, stdin 바이트 계수기), `internal/cli/hook.go`, `internal/cli/hook_stdin_failclosed_test.go`·`hook_stdin_failclosed_ast_test.go`(신규), `internal/cli/hook_protocol_fix_test.go`·`misc_coverage_test.go`(acceptance.md §D 갱신 대상 둘). `internal/hook/**` 변경 없음.
- 설계 결정:
  - **`writeCodexFailClosed` 는 손대지 않았다.** plan.md M1 이 허용한 「파싱 실패 전용 작성기」 경로를 택했다 — 새 `writeFailClosedDeny` 가 Codex 는 `TranslateCodex`, Claude 는 `Lookup(HarnessClaude, ev, DecisionFatalError)` + `Render` 로 렌더링한다. 이유: `hook-fault` 키 호출부의 동작을 바이트 단위로 보존하고, t1099 소유 파일(`hook_codex_failclosed.go`)의 diff 를 0으로 둬 병합 위험을 줄인다. 대가: 기록·stdout 쓰기 약 15줄이 두 작성기에 겹친다.
  - 술어 이름은 `HostLacksStopBlockCap(h Harness, ev hook.EventType) bool` — (Codex, Stop) 에만 참.
  - 사유 상수: `stdinParseFailureCause = "hook stdin could not be parsed as JSON"`, `stdinParseFailClosedDocID = "moai-doc:hook-stdin-fail-closed"`. 사유 = `"fail-closed: " + cause + " (" + docID + ")"`. 운영자 문서 본문은 sync-phase 몫(spec.md §D).
  - 기록 키: 파싱 실패 fail-closed `stdin-parse-fail-closed`, Codex Stop 면제 `stdin-parse-exempt` (둘 다 `hook-fault` 와 다름). 기록면은 Q7 권장대로 `codexadapter.RecordDiscards`(`.moai/logs/codex-adapter.jsonl`) 하나를 두 하네스가 쓴다.
  - 하위 명령 표를 `init` 안 지역 변수에서 패키지 변수 `hookEventSubcommands` 로 옮겼다 — 테스트가 관측 이벤트 22개를 손으로 적지 않고 디스패처와 같은 표에서 계산하기 위함(acceptance.md §0).
  - `runAgentHook` 의 action → 이벤트 매핑을 함수 `agentActionEvent` 로 추출해 stdin 앞으로 옮겼다(내용 불변).

#### RED (구현 전, 행동 테스트)

측정 트리: HEAD `9b6296a63` + 미커밋 구조 준비(표 이동, `agentActionEvent` 추출, 상수 파일, 테스트 파일). 파싱 실패 분기는 수리 전 그대로였다. 명령: `unset MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test -count=1 -run StdinFailClosed ./internal/cli/` → exit 1.

```
--- FAIL: TestStdinFailClosed_ClaudeDecisionEvents (0.11s)
--- FAIL: TestStdinFailClosed_CodexDecisionEvents (0.08s)
--- FAIL: TestStdinFailClosed_ObservedDecisionSetMatchesTheSource (0.02s)
--- FAIL: TestStdinFailClosed_HarnessDecidedBeforeStdin (0.00s)
--- FAIL: TestStdinFailClosed_StopIgnoresStopHookActive (0.00s)
--- FAIL: TestStdinFailClosed_CodexStopExempt (0.02s)
--- FAIL: TestStdinFailClosed_AgentDecisionActions (0.18s)
--- FAIL: TestStdinFailClosed_AgentRejectsBogusHarness (0.02s)
--- FAIL: TestStdinFailClosed_AgentObservationActionsPreserved (0.33s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.863s
```

실패 사유(중복 제거):

```
hook_stdin_failclosed_test.go:439: stdout = "{}", want a fail-closed deny (neither empty nor {})
hook_stdin_failclosed_test.go:472: stdout = "{}", want a fail-closed deny (neither empty nor {})
hook_stdin_failclosed_test.go:508: claude: observed fail-closed set map[], want map[PermissionRequest:true PreToolUse:true Stop:true UserPromptSubmit:true]
hook_stdin_failclosed_test.go:522: RunE = <nil>, want an invalid --harness value error
hook_stdin_failclosed_test.go:591: claude Stop stdout = "{}\n", want decision:block regardless of stop_hook_active
hook_stdin_failclosed_test.go:631: (c) no exemption stderr line:
hook_stdin_failclosed_test.go:695: (c) stdout = {}
hook_stdin_failclosed_test.go:712: bogus+malformed RunE = <nil>, want an invalid --harness value error
hook_stdin_failclosed_test.go:731: RunE = <nil>, want an invalid --harness value error
hook_stdin_failclosed_test.go:779: x-verification bogus RunE = <nil>, want an invalid --harness value error
```

녹색으로 남은 것: `TestStdinFailClosed_ObservationEventsPreserved`(특성 기준 — 수리 전에도 녹색이어야 한다)와 `TestStdinFailClosed_DepthControl`. `AgentObservationActionsPreserved` 는 보존 32경우가 아니라 `--harness bogus` 새 거부 단언(`:779`)에서만 실패했다. 술어 테스트: 스텁(`return false`)에서 `go test -count=1 -run TestHostLacksStopBlockCap ./internal/codexadapter/` → `stop_cap_test.go:24: exempt pairs = [], want exactly [codex/Stop]`, `FAIL`. AST 테스트는 구현 뒤에 썼으므로 수리 전 RED 가 없다 — 대신 아래 변이 (c2)·(c4)·(c5)·(c6) 이 RED 를 관측한다.

#### GREEN (구현 커밋 `53ec0d5a8` 의 트리 — 커밋 전 워킹 트리에서 측정, 이후 코드 변경 없음)

```
$ unset MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test -count=1 -v -coverprofile=<scratch>/cover_stdin.out -run StdinFailClosed ./internal/cli/
--- PASS: TestStdinFailClosed_SingleDecisionListAST (0.09s)
--- PASS: TestStdinFailClosed_ASTCheckerDetectsItsTargets (0.00s)
--- PASS: TestStdinFailClosed_ClaudeDecisionEvents (0.10s)
--- PASS: TestStdinFailClosed_CodexDecisionEvents (0.09s)
--- PASS: TestStdinFailClosed_ObservedDecisionSetMatchesTheSource (0.06s)
--- PASS: TestStdinFailClosed_HarnessDecidedBeforeStdin (0.00s)
--- PASS: TestStdinFailClosed_ObservationEventsPreserved (1.23s)
--- PASS: TestStdinFailClosed_StopIgnoresStopHookActive (0.01s)
--- PASS: TestStdinFailClosed_CodexStopExempt (0.05s)
--- PASS: TestStdinFailClosed_AgentDecisionActions (0.34s)
--- PASS: TestStdinFailClosed_AgentRejectsBogusHarness (0.02s)
--- PASS: TestStdinFailClosed_AgentObservationActionsPreserved (0.22s)
--- PASS: TestStdinFailClosed_DepthControl (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.856s	coverage: 5.8% of statements
```

`=== RUN` 293행(최상위 13 + 하위 280: Claude 16, Codex 12, 관측 176, Codex Stop 4, agent 결정 32, bogus 8, agent 관측 32). 위 표의 시간은 마지막 실행의 값이다(최종 실행 `final_stdin.txt`, exit 0, `=== RUN` 293).

```
$ unset MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test -count=1 -v -run Hook ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	61.235s      (=== RUN 214, --- FAIL 0)
--- PASS: TestRunAgentHook_AllActionSuffixes (0.00s)
--- PASS: TestRunHookEvent_ReadInputError (0.00s)
--- PASS: TestHookFaultInjection (0.01s)
--- PASS: TestRunHookEvent_MalformedStdinGraceful (0.00s)
--- PASS: TestRunAgentHook_ReadInputError (0.00s)
```

(마지막 다섯 줄은 같은 명령의 앞선 실행 `run_hook.txt` 에서 발췌 — 코드는 그 뒤 `translate.go` 주석 한 줄만 바뀌었다. 최종 실행은 `--- FAIL` 0건.)

```
$ unset MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test -count=1 -run Codex ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	144.411s
$ go test -count=1 -cover ./internal/codexadapter/...
ok  	github.com/modu-ai/moai-adk/internal/codexadapter	0.570s	coverage: 88.4% of statements
```

(`-run Codex` 는 변이 실행 전·`translate.go` 주석 변경 전에 돌렸다. 변이 뒤 파일 복원은 `cmp` 로 바이트 일치를 확인했다.)

#### 필수 RED 변이 (각각 적용 → 관측 → 복원, 복원 후 `cmp` 바이트 일치 확인)

러너: `<scratch>/mut/run.sh <변이> <-run 필터> ./internal/cli/`(또는 codexadapter), 변이 정의 `<scratch>/mut/apply.py`.

| 변이 | 요구 출처 | 결과(원문 발췌) |
|---|---|---|
| 사유 끝에 ` — update moai or turn hooks off` 덧붙임 | AC-HSF-001 (0.4.1) | `test_exit=1`; `hook_stdin_failclosed_test.go:447: (e4) reason = "fail-closed: hook stdin could not be parsed as JSON (moai-doc:hook-stdin-fail-closed) — update moai or turn hooks off", want exactly …` (같은 실행에서 (d) 바이트 비교도 실패) |
| (c1) `IsDecisionBearing` 결과를 `false` 로 | AC-HSF-003 | `:508: claude: observed fail-closed set map[], want map[PermissionRequest:true PreToolUse:true Stop:true UserPromptSubmit:true]` |
| (c2) 여러 줄 `switch` 중복 목록 삽입 | AC-HSF-003 | `hook_stdin_failclosed_ast_test.go:419: (b2) hook_stdin_failclosed.go:139:2: switch case lists lists EventPreToolUse,EventStop` |
| (c3) 면제 술어 건너뛰기 | AC-HSF-003 | `:508: codex: observed fail-closed set map[PermissionRequest:true PreToolUse:true Stop:true UserPromptSubmit:true], want map[PermissionRequest:true PreToolUse:true UserPromptSubmit:true]` |
| (c4) `runAgentHook` 을 `event == hook.EventPreToolUse` 한 번 비교로(행동은 같게) | AC-HSF-003 | `-run StdinFailClosed` 전체에서 **`SingleDecisionListAST` 만** FAIL: `(b1) runAgentHook: decision-event comparison in the parse-failure scope: == EventPreToolUse` / `(b1) runAgentHook: no codexadapter.IsDecisionBearing result selects a branch …` |
| (c5) `_ = codexadapter.IsDecisionBearing(event)` + 관측 식별자 부정 조건으로 분기 | AC-HSF-003 | 역시 `SingleDecisionListAST` 만 FAIL: `(b1) runAgentHook: no codexadapter.IsDecisionBearing result selects a branch in the parse-failure scope` |
| (c6) 공유 처리 함수에 `if os.Getenv("MOAI_T1152_MUTANT") != "" { 기본 출력 }` | AC-HSF-003 | `SingleDecisionListAST` 만 FAIL: `(b3) hook_stdin_failclosed.go:79:5: os.Getenv in the parse-failure path` |
| stdin 계수기가 페이로드 앞 256바이트를 stderr 에 씀 | AC-HSF-007 | `:448: stderr carries the {depth,malformed,oversize,truncated} canary — payload leaked` (네 형태 모두) |
| 술어를 늘 거짓으로 | AC-HSF-011 | `:621: (b) stdout = "{\"decision\":\"block\",\"reason\":\"MoAI Stop hook failed, so the call was denied fail-closed: …` — (b) 에서 `t.Fatalf` 로 멈춰 (c) 는 이 실행에서 따로 보고되지 않았다 |
| (1) 관측 매핑에도 fail-closed 적용 | AC-HSF-013 | `:756: RunE = fail-closed for PostToolUse: no claude fatal_error translation row, want nil` (PostToolUse·SubagentStop × 두 하네스) |
| (2) `--harness` 판독을 결정 매핑 action 에서만 | AC-HSF-013 | `-run StdinFailClosed` 에서 `AgentRejectsBogusHarness`·`AgentObservationActionsPreserved` 만 FAIL(`AgentDecisionActions` 는 통과): `:779: x-verification bogus RunE = <nil>, want an invalid --harness value error` |
| N16 — `runAgentHook` 경로의 판정 결과 반전(`!` 제거) | Kickoff 판정 (c) | `:695: (c) stdout = {}` — AC-HSF-012 RED |

N15(`--harness bogus` 거부를 agent action 8개 모두에 단언)는 `TestStdinFailClosed_AgentRejectsBogusHarness` 가 닫는다 — 수리 전 RED(`:731`), 수리 후 8개 모두 PASS.

#### AC 판정표

| AC | 판정 | 결정 근거(이 실행, 트리 `53ec0d5a8`) |
|---|---|---|
| AC-HSF-001 | PASS | `TestStdinFailClosed_ClaudeDecisionEvents` 16경우 PASS — (a) nil, (b) 디스패치 0, (c) 단일 JSON, (d) `Render(ev, Lookup(HarnessClaude…).Outcome, 기대 사유)` 바이트 일치, (e1)~(e5) — 기대 사유는 테스트가 두 상수와 고정 틀로 조립. 필수 변이 RED 관측 |
| AC-HSF-002 | PASS | `TestStdinFailClosed_CodexDecisionEvents` 12경우 PASS — `TranslateCodex(ev, DecisionFatalError, 기대 사유)` 바이트 일치. PermissionRequest 포함(파싱 실패가 Codex 이벤트 검증보다 앞선다) |
| AC-HSF-003 | PASS | (a) `ObservedDecisionSetMatchesTheSource` PASS(두 하네스 집합을 원천에서 계산), (b1)(b2)(b3) `SingleDecisionListAST` PASS, 검사기 자체의 탐지력 `ASTCheckerDetectsItsTargets` PASS, 변이 (c1)~(c6) RED 관측 |
| AC-HSF-004 | PASS | `HarnessDecidedBeforeStdin` — bogus+malformed: `invalid --harness value`, 파싱 경고 없음, `ReadInput` 0회; codex+malformed: Codex 형태 |
| AC-HSF-005 | PASS | `ObservationEventsPreserved` 176경우 PASS — stderr 경고 정확히 한 줄, stdout `{}\n`(worktree 둘 0 B), 기록 0건. 수리 전 기준선(위 표)과 같은 바이트 |
| AC-HSF-007 | PASS | fail-closed 28경우(`Claude/CodexDecisionEvents`)와 agent 32경우(`AgentDecisionActions`)에서 `assertFailClosedObservability` — stderr 줄(이벤트·`harness <mode>`·원인·fail-closed), 기록 1건(키 `stdin-parse-fail-closed`, `content_length` = 관측 바이트 수, oversize 는 `5<<20`), 카나리 부재. 필수 변이 RED |
| AC-HSF-008 | PASS | `StopIgnoresStopHookActive` PASS(두 하네스 × `stop_hook_active` true/false). `grep -n '@MX:WARN' internal/cli/hook_stdin_failclosed.go` → `64:// @MX:WARN: [AUTO] a Stop that fails to parse is blocked under Claude on every turn — … assumed from repository doctrine and unmeasured; …`. plan.md §B.1 Q2 에 판정·증거 경로 있음. codex-cli 버전 기록은 위 §F.1 3(a)(0.157.0). 추적 항목 Q8(Claude JSON block 상한)은 **미측정이며 사유를 기록한다**: 라이브 `claude -p` 실행이 필요하고 이번 위임에 turn·벽시계 상한 승인이 없다(AC 는 「측정 결과 또는 미측정 사유」를 요구) |
| AC-HSF-009 | PASS | 위 「spec.md §F.1 의 2」 — 명령, 빈 출력, 양성 대조, blob 대조, (a)~(e) 모두 「동일」 |
| AC-HSF-011 | PASS | `CodexStopExempt` 4경우 PASS — `{}`, 면제 stderr 줄(`no Stop block cap`), 기록 1건(키 `stdin-parse-exempt`), 카나리 부재, `--harness` 없으면 block. 술어 선언부 주석에 Q2 증거 경로·codex-cli 0.156.1·관측 범위·재검토 지점(`internal/codexadapter/stop_cap.go`). 필수 변이 RED((b) 관측) |
| AC-HSF-012 | PASS | `AgentDecisionActions` 32경우 + bogus+malformed(`ReadInput` 0회) + `AgentRejectsBogusHarness`(유효 stdin, 8 action) PASS. N16 변이 RED |
| AC-HSF-013 | PASS | `AgentObservationActionsPreserved` 32경우 보존 + `x-verification`·`x-completion` bogus 거부(`ReadInput` 0회). 변이 (1)·(2) RED |
| AC-HSF-GATE | PASS(범위 한정) | 아래 품질 게이트. acceptance.md 가 적은 `-run 'Hook\|Stdin\|FailClosed\|AgentHook'` 결합 정규식 대신 단어 필터 `Hook`·`StdinFailClosed`·`Codex` 를 따로 돌렸다(`=== RUN` 214 / 293 / 비-v 실행) — 선택 범위는 합집합으로 같거나 넓다 |

#### 품질 게이트 (이 실행)

```
$ go build ./... ; echo build_exit=$?
build_exit=0
$ GOOS=windows GOARCH=amd64 go build ./... ; echo winbuild_exit=$?
winbuild_exit=0
$ go vet ./internal/cli/ ./internal/codexadapter/ ; echo vet_exit=$?
vet_exit=0
$ golangci-lint run ./internal/cli/... ./internal/codexadapter/... ; echo lint_exit=$?
0 issues.
lint_exit=0
$ gofmt -l <변경 Go 파일 9개>
(출력 없음)
```

커버리지: `internal/codexadapter` 88.4%(패키지 전체). `internal/cli/hook_stdin_failclosed.go` 파일 단위 32문 중 28문 = 87.5%(`-run StdinFailClosed` 프로필에서 계산; 미커버는 렌더·쓰기·기록 실패 분기). `internal/cli` 패키지 전체 커버리지는 필터 실행이라 의미가 없다(5.8%) — 전 패키지 판정은 CI 몫.

#### `internal/cli/hook.go` 편집 범위 (t1099 M2d 충돌 위험 판단용)

`git diff -U0 5fcc615a9 53ec0d5a8 -- internal/cli/hook.go` 의 새 파일 기준 줄:

- `:37-73` — 하위 명령 표를 패키지 변수 `hookEventSubcommands` 로 이동(추가), `init` 안 `:84-85` 가 그것을 참조(삭제 33줄 → 2줄).
- `:278-288` — `runHookEvent` 에서 `harnessModeIsCodex` 를 `ReadInput` 앞으로, stdin 계수기.
- `:290-296` — `runHookEvent` 파싱 실패 분기 → `answerStdinParseFailure`.
- `:304-307` — 종전 하네스 판독 자리(주석만 남김). `validateCodexHarnessEvent` 호출과 그 아래 디스패치 경로(M2d 가 바꾸는 `Dispatch` 호출 주변)는 **건드리지 않았다.**
- `:472-493` — `runAgentHook` 에서 매핑·하네스 판독을 stdin 앞으로, 파싱 실패 분기.
- `:523-542` — `agentActionEvent` 추가.

충돌 탐침(텍스트 수준): `git merge-tree --write-tree --name-only HEAD e7e3b3813`(t1099 M2d 끝) → 트리 `3708a2db6…` 만 출력(충돌 파일 없음). `git merge-tree --write-tree --name-only HEAD e722a1493`(t1099 현재 tip) → 트리 `76cf7e22d…` 만 출력. 병합 트리의 `hook.go` 에서 `harnessModeIsCodex` 선언은 진입점마다 1회(`:301`, `:506`), `newCodexStopChain` 분기(`:371`)는 파싱 성공 뒤에 있다. 병합 트리의 컴파일·테스트는 하지 않았다(Gap).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: 53ec0d5a8
run_status: audit-ready
baseline_commit_sha: 9b6296a63
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: "internal/hook diff 0; writeCodexFailClosed/hook-fault unchanged"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows_amd64: pass
total_run_phase_files: 11
m1_to_mN_commit_strategy: "baseline commit (9b6296a63) then one implementation commit (53ec0d5a8) for M1-M4"
pushed: false
open_gaps:
  - "codex-cli 0.157.0 != Q2 version 0.156.1; Q2 re-measure (live) not run"
  - "Q8 Claude JSON block cap not measured"
  - "merged tree with t1099 tip not compiled"
  - "internal/cli full package suite not run (CI)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
