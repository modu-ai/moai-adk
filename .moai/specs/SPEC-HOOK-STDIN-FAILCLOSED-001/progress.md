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

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
