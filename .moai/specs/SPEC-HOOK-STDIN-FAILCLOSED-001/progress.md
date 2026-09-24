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

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
