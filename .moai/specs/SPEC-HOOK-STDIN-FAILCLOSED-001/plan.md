---
id: SPEC-HOOK-STDIN-FAILCLOSED-001
title: "Plan — 훅 stdin 파싱 실패의 결정 이벤트 fail-closed"
version: "0.4.2"
created: 2026-09-24
author: manager-spec (card t1152)
---

# Plan — SPEC-HOOK-STDIN-FAILCLOSED-001

마일스톤은 **되돌리기 어려운 결정부터** 배치했다. 사람이 먼저 봐야 할 것은 M0(Kickoff 차단 질문 판정 — 0.3.0 에서 판정됨)과 M1(결정 이벤트의 동작, Codex Stop 면제 술어)이며, 순서 이동·테스트 갱신 같은 기계적 작업은 뒤에 둔다. 우선순위 표기만 쓰고 기간 추정은 하지 않는다.

0.2.0 개정: plan-audit 1회차(`.moai/reports/t1152/plan-audit.md`)의 D0·D2·D3·D9·D10·D12 를 반영했다. 명확화 필요 표식은 Kickoff 차단 질문(Q2·Q1·Q3)에만 남기고, 나머지(Q4~Q7)는 표식 없는 추적 항목으로 옮겼다.

0.3.0 개정: Kickoff 차단 질문 셋을 판정 결과로 닫았다(§B.1). 명확화 필요 표식은 이제 하나도 남지 않는다. Kickoff Approval 자체는 아직 요청 전이며, 리드 지시에 따라 t1099 착지 후 요청한다(§F M0, spec.md §F.1 의 3).

0.3.1 개정: plan-audit 2회차(`.moai/reports/t1152/plan-audit-iter2.md`, FAIL 0.86)의 차단 결함 N1~N3 과 선택 결함 N4·N7 을 반영했다. §B.1 Q1 에 파일 부재 해석의 운영자 확인과 「추가 후 제거」 잔여 위험의 처분을 기록하고, §C 에 agent 경로의 잘못된 `--harness` 기준선(4(c) 안)과 호스트 `env` 전파·유지 측정(Pre-flight 5)을 추가했으며, M3 의 `runAgentHook` 서술과 §B.2 추적 항목 Q9 를 고쳤다.

0.3.2 개정: plan-audit 3회차(`.moai/reports/t1152/plan-audit-iter3.md`, FAIL 0.88)의 차단 결함 N8·N9·N10 과 선택 결함 N11·N12 를 반영했다. §C Pre-flight 5 를 다시 짰다 — 삭제 전 관측을 판정 전제로 두고, 같은 세션 전파와 세션 시작 전파를 모두 재며, 「전달 없음」을 세 번째 결과로 따로 적었다(N8). 실행 1회를 `claude -p` 프로세스 하나로 정의하고 예산을 다시 세었으며, `timeout -k` 강제 종료와 숨은 `--max-turns` 플래그의 거부 처리를 적고, 상한을 Kickoff 때 운영자가 승인할 제안값으로 표시했다(N9, §F M0). Pre-flight 4(c) 에 관측 매핑 action 의 잘못된 `--harness` 기준선을 더했다(N10).

0.3.3 개정: plan-audit 4회차(`.moai/reports/t1152/plan-audit-iter4.md`, FAIL 0.89)의 차단 결함 N13·N14 를 반영했다. §C Pre-flight 5 의 판정 전제에 삭제 뒤 기록 요건을 더하고(삭제가 실행되지 않았거나 삭제 뒤 기록이 없는 변형은 「미측정」), 경로의 확정을 정의해 판정 네 결과가 모든 경우를 하나씩 덮게 했으며, 실행의 권한 모드를 정했다.

0.3.4 개정: Implementation Kickoff Approval 을 기록했다(§F M0, progress.md §E.1 「Kickoff 판정」). plan-audit 5회차(`.moai/reports/t1152/plan-audit-iter5.md`, PASS-WITH-DEBT 0.91)의 선택 결함 가운데 운영자가 측정 전에 닫으라고 한 N17·N18 을 반영했다 — §C Pre-flight 5 에 변형별 설정 파일 배치와 격리 확인(리드 조건)을 더하고, M0 승인 목록에 권한 모드를 넣었다. N15·N16 은 운영자 판정대로 run 에서 테스트를 더해 닫는다.

0.4.0 개정: §C Pre-flight 5 를 수행했고 결과는 「유지한다」였다(`.moai/reports/t1152/preflight5.md`). 정지 조건에 따라 운영자 재판정을 받았다 — **탈출 장치 없이 결정 이벤트는 언제나 거부한다**(Codex Stop 면제 유지, 2026-09-25). §B.1 Q1 에 재판정을 기록하고 폐기된 탈출 장치 판정을 표시했으며, §C 5 를 완료 기록으로 줄이고, M2 에서 탈출 장치를 빼고 복구 절차의 문서화를 넣었다. §D·§E·§G·M1b·M4 를 맞췄다.

0.4.1 개정: plan-audit 6회차(`.moai/reports/t1152/plan-audit-iter6.md`, FAIL 0.87)의 N23 을 반영해 M0 의 미래형 Pre-flight 5 문장을 지웠다. 차단 결함 N19~N21 은 acceptance.md(AC-HSF-001(e4)·(e5), AC-HSF-003(b3))와 spec.md §F.1 3(c) 에서 고쳤다. 마일스톤 순서·범위는 그대로다.

0.4.2 개정: §C 3 에 리드 판정(2026-09-26)을 기록했다 — run 시점 설치본 codex-cli 0.157.0 에서 Q2 를 다시 재지 않고 0.156.1 에 묶인 관측으로 남긴다(spec.md §F.2 「Q2 의 버전 한정」). 0.157.0 재측정은 리드의 후속 카드 몫이다. 마일스톤 순서·범위는 그대로다.

## §A 맥락

spec.md §A 참조. 요약: `internal/cli/hook.go:272-280` 의 stdin 파싱 실패 분기가 결정 이벤트에서도 `{}` + exit 0 을 내 가드 전체가 우회된다. t1099(`fabc33812`)의 결정 이벤트 집합과 번역 표를 재사용해 결정 이벤트 4개만 fail-closed 로 바꾸되 Codex 하네스의 Stop 은 면제하고(0.3.0 판정), 같은 결함이 있는 `runAgentHook` 도 함께 고치며, 관측 이벤트 26개는 `6a3603274` 의 의도대로 이벤트별 현재 출력을 보존한다(하위 명령 22개 중 20개 `{}`, worktree-create·worktree-remove 는 빈 stdout).

## §B 질문

### B.1 Kickoff 차단 질문 — 판정됨 (2026-09-24)

세 질문 모두 판정됐다. 각 항목은 판정 결과, 출처, 판정 전 선택지(기록용 원문 요지) 순으로 적는다. 순서는 원래대로 Q2 → Q1 → Q3 이다 — Q1 의 Codex 분기가 Q2 의 답에 달려 있었기 때문이다.

- **Q2 — Codex 의 Stop 연속 차단 상한 유무: 판정됨 — 상한 없음(관측).**
  - 출처: t1152 레인 세션의 직접 측정, 2026-09-24 12:39–12:50 KST, codex-cli 0.156.1. 증거: `.moai/reports/t1152/q2-codex-stop-cap.md` (로컬 증거, gitignored — 커밋되지 않는다).
  - 방법: 격리된 `CODEX_HOME` 에 사용자 범위 Stop 훅 하나(stdin 을 기록하고 `{"decision":"block","reason":…}` 출력, exit 0)를 두고 `codex exec` 를 실행했다.
  - 관측: Stop 훅이 **191회 연속** 차단했고 Codex 는 약 10분 동안 스스로 멈추지 않았다. 측정자가 프로세스를 종료했다. `stop_hook_active` 는 1번째 호출에서 `false`, 2~191번째에서 `true` 였다(`jq -r '.stop_hook_active' … | sort | uniq -c` → `1 false` / `190 true`). 음성 대조(차단 없는 Stop 훅)는 1회로 끝났다.
  - 관측 범위(한계): 기본 설정만, 비대화형 `exec` 만, 모델 `gpt-6-astra` 만. 「191회·약 10분 동안 상한 미발동」이지 무한의 증명이 아니다. 대화형 TUI 와 사용자 설정의 상한 키 유무는 재지 않았다.
  - Claude 쪽: JSON `decision:"block"` + exit 0 이 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` 에 걸리는지는 **여전히 미측정**이며 저장소 독트린(`goal-directive.md:13`, `internal/cli/hook_stop_goal.go:120-121`)에만 근거한다. Kickoff 차단 사유로 두지 않고 잔여 위험(spec.md §F.2)과 추적 항목 Q8 로 옮겼다.
  - 판정 전 선택지(기록): (a) Kickoff 전에 측정 — **이것이 실행됐다(Codex 쪽)**; (b) 측정 없이 Q1 을 조건부로 판정하고 측정을 run Pre-flight 로 넘긴다.
- **Q1 — Stop·UserPromptSubmit 의 fail-closed 포함 여부와 탈출 장치 메커니즘: 판정됨 — 선택지 A1. 탈출 장치는 0.4.0 재판정으로 없앴다(아래 「재판정」).**
  - 출처: 운영자 판정, t1152 레인의 AskUserQuestion, 2026-09-24.
  - 판정 내용(동작): 결정 이벤트 4개(PreToolUse·PermissionRequest·Stop·UserPromptSubmit)를 두 하네스 모두에서 fail-closed 로 한다. **예외는 Codex 하네스의 Stop 하나** — 관측 경로(`{}` + exit 0)에 stderr 한 줄과 면제 기록을 더한다(spec.md REQ-HSF-012). 근거는 Q2 측정이다: Codex 에 상한이 없으므로 여기서 차단하면 파싱 실패가 지속되는 동안 턴이 끝나지 않는다.
  - 판정 내용(표현): 예외는 두 번째 이벤트 목록이 아니라 `internal/codexadapter` 안의 **이름 붙은 술어 하나**로 표현한다 — 「Codex 호스트에는 Stop 연속 차단 상한이 없다」(spec.md REQ-HSF-013). 향후 Codex 가 상한을 도입하면 그 술어 하나만 다시 보면 된다. 술어의 식별자 이름은 run-phase 가 정한다.
  - **재판정 (2026-09-25) — 탈출 장치 없음.** 출처: 운영자 판정, t1152 레인의 AskUserQuestion. 계기: §C Pre-flight 5 의 결과 「유지한다」 — Claude Code 2.1.281 `claude -p` 에서 세션 중 `settings.local.json` `env` 에 추가한 키가 다음 훅부터 훅 환경에 전달되고, 키를 지운 뒤에도 남았다(A1). 따라서 아래 메커니즘 (i) 의 출처 제약은 「추가 후 제거」로 우회된다. 판정 내용: (1) 결정 이벤트는 파싱 실패 시 언제나 거부한다 — 실행 시점 스위치를 두지 않는다(spec.md REQ-HSF-001 금지 절, REQ-HSF-007·016 삭제). (2) Codex Stop 면제는 그대로다. (3) 지속적인 파싱 실패(예: 호스트 형식 변화)는 moai 갱신이나 훅 비활성화(`disableAllHooks`)로 복구하며 운영자 문서에 적는다. (4) 거부 사유는 원인(stdin 파싱 실패)과 운영자 문서 식별자만 싣는다(REQ-HSF-010). 아래 「판정 내용(탈출 장치)」·「부재 해석 확인」·「판정 범위 밖의 선택지」는 이 재판정으로 폐기됐으며 기록으로만 남긴다.
  - [0.4.0 에서 폐기] 판정 내용(탈출 장치): 후보 (i) — **셸 환경 변수 하나**. 이름은 run-phase 에서 `internal/config/envkeys.go` 상수로 정의한다. 같은 키가 `$CLAUDE_PROJECT_DIR/.claude/settings.json` 또는 `.claude/settings.local.json` 의 `env` 블록에 선언돼 있으면 환경 변수를 **무시**한다. 존재하는 설정 파일을 읽지 못하면 활성화를 **인정하지 않는다**(fail-closed). 파일 부재는 「선언 없음」으로 본다 — 이 구분은 판정 문구를 구현 가능하게 옮긴 manager-spec 의 해석으로 시작했고, 0.3.1 에서 운영자가 확인했다(아래 「부재 해석 확인」, spec.md REQ-HSF-016). 거부 사유는 키 이름을 싣지 않는다(REQ-HSF-010).
  - [0.4.0 에서 폐기] 기록된 잔여 면: 사용자 범위 `~/.claude/settings.json` 의 `env`. 호스트가 이 블록을 훅 환경으로 전파하면 셸 환경 변수와 구별되지 않는다(spec.md §F.2).
  - [0.4.0 에서 폐기] **부재 해석 확인: 판정됨 (2026-09-24).** 출처: 운영자 판정, t1152 레인의 AskUserQuestion, plan-audit 2회차 N1 후속. 내용: 설정 파일이 **없으면** 「선언 없음」으로 보고, 탈출 장치를 셸 환경 변수로 인정할 수 있다. 조건 두 가지 — (1) 「추가 후 제거」 우회(설정 `env` 에 키 추가 → 호스트 반영 → 키 삭제 또는 파일 삭제 → 파손 입력)를 잔여 위험으로 기록한다(spec.md §F.2 「추가 후 제거」 행 — 키 삭제 변형은 부재 해석과 무관하게 존재하고, 파일 삭제 변형은 부재 해석이 추가로 연다). (2) run 시작 시 호스트 동작을 측정한다(§C Pre-flight 5). 호스트가 키·파일을 지운 뒤에도 그 `env` 값을 훅 환경에 유지한다고 나오면, 구현을 시작하지 않고 운영자에게 돌아가 재판정을 받는다. 이 조건을 집행하는 판정 규칙은 §C Pre-flight 5 에 있다 — 0.3.2 에서 삭제 전 관측을 판정 전제로 두고, 값이 어느 경로로도 오지 않는 「전달 없음」을 따로 닫도록 고쳤다(plan-audit 3회차 N8).
  - [0.4.0 에서 폐기] 판정 범위 밖의 선택지(기록만, 채택 아님 — 재판정에서도 채택되지 않았다): 한 세션에서 한 번이라도 선언이 관측된 키는 이후 선언이 사라져도 계속 불인정하는 강화. 파싱 실패 상태에서는 stdin 의 `session_id` 를 읽을 수 없어 stdin 밖의 세션 열쇠와 상태 저장이 필요하다 — 선택지 A2 와 같은 미측정 전제다. Pre-flight 5 가 「유지한다」로 나와 재판정할 때 후보로 올린다.
  - 판정 전 선택지(기록):
    - (A) 네 이벤트 전부 fail-closed + 탈출 장치 — Q2 에 조건부. Codex Stop 상한이 확인되지 않으면 (A1) Codex Stop 만 관측 경로로 두고 예외를 이름 붙은 술어로 표현, 또는 (A2) MoAI 자체 상한(세션 열쇠가 stdin 밖에 있어야 하며 미측정). **A1 이 채택됐다.**
    - (B) PreToolUse·PermissionRequest 만 fail-closed, Stop·UserPromptSubmit 은 fail-open + 기록. 결정 집합의 부분집합을 새로 정의해야 해 단일 목록 제약과 긴장. 채택되지 않았다.
    - 탈출 장치 후보: (i) 환경 변수 + 프로젝트·로컬 `env` 선언 시 무시 — **채택(0.3.0), 0.4.0 재판정으로 폐기**; (ii) 프로젝트 밖 사용자 범위 파일(예: `~/.moai/` 아래) — 채택되지 않았다; `.moai/config/sections/*.yaml` 키는 출처 제약에 정면으로 걸려 애초에 후보가 아니었다. 이전 판이 이 면을 「ConfigChange 감시의 몫」으로 넘겼던 것은 0.2.0 에서 철회했다.
- **Q3 — `runAgentHook` 포함 여부: 판정됨 — 포함.**
  - 출처: 운영자 판정, t1152 레인의 AskUserQuestion, 2026-09-24.
  - 판정 내용: `runAgentHook`(`internal/cli/hook.go:447`, 파싱 실패 분기 `:455-463`)을 범위에 넣는다. action 이 결정 이벤트로 매핑되는 경우(`:469-481` 의 switch — 접미사 `-validation`·`-pre-transformation`·`-pre-implementation` 과 `default` 의 미지 action 이 모두 PreToolUse)는 `runHookEvent` 와 같은 fail-closed 출력을 내고, 관측 이벤트로 매핑되는 경우(`-verification`·`-post-transformation`·`-post-implementation` → PostToolUse, `-completion` → SubagentStop)는 현재 동작을 유지한다(spec.md REQ-HSF-014·015). 매핑은 이 트리(`44dfc25fd`)의 코드를 읽어 확인했다. 현재 이 함수는 `--harness` 플래그를 읽지 않는다 — 플래그는 `hookCmd` 의 PersistentFlags(`internal/cli/hook.go:45`)라 `agent` 하위 명령에도 붙어 있으므로, 하네스 모드 판정을 stdin 앞으로 두는 REQ-HSF-005 를 여기에도 적용한다 — `runAgentHook` 에서는 하네스 판독과 잘못된 값의 거부가 **새 동작**이다. 배포 래퍼(`internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh:47`)는 `moai hook agent "$1"` 을 `--harness` 없이 부르므로, codex 모드 agent 경로는 현재 배포 설정에서 도달하지 않는다(추적 항목 Q9).
  - 판정 전 선택지(기록): 포함(권장) / 제외하고 별도 카드. 포함이 채택돼 이 0.3.0 개정이 run 착수 전에 필요한 D-NEW-1 성격의 REQ·AC 추가를 담는다.

### B.2 추적 항목 — Kickoff 를 막지 않는다

- **Q4 빈 stdin 경로** — `ReadInput` 은 빈 stdin 을 오류가 아니라 기본 입력으로 처리한다(`internal/hook/protocol.go:52-56`). 결정 이벤트에서 빈 페이로드로 가드가 기본 입력을 받아 실행되는데, 가드가 빈 `tool_name` 을 허용으로 처리하면 같은 우회가 된다. 이 SPEC 은 다루지 않는다. 별도 카드로 측정할지는 리드가 판단하고 판정만 기록한다.
- **Q5 5 MiB 경로의 호스트 실현성** — 호스트(Claude Code, Codex)가 5 MiB 를 넘는 `tool_input` 을 훅 stdin 으로 실제 넘기는지 측정하지 않았다. 수리의 정당성과 무관하다: 모델이 제어하는 더 싼 경로(중첩 깊이 초과, 약 20 KB)의 파싱 실패가 측정됐다(spec.md §A.3). 측정은 sync-phase `--security --deep` 렌즈의 위협 등급 판단용으로 run 초반에 수행하고 progress.md 에 기록한다. 같은 자리에서 「호스트가 깊게 중첩된 `tool_input` 을 그대로 넘기는가」도 잰다.
- **Q6 Can Block 11개** — spec.md §B.3 의 이벤트 가운데 결정 집합으로 옮겨야 할 것이 있는지. 이 SPEC 은 옮기지 않는다. 옮기려면 `DecisionBearingEvents()` 를 바꾸는 별도 카드가 필요하다 — 판정만 기록한다.
- **Q8 Claude 호스트의 JSON block 상한** — `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` 가 exit 2 가 아닌 JSON `decision:"block"` + exit 0 차단에도 걸리는지 측정하지 않았다. 근거는 저장소 독트린뿐이다(Q2 의 Claude 쪽 항목). 걸리지 않는다면 Claude 하네스의 Stop fail-closed 도 Codex 와 같은 무기한 잠김 위험을 갖는다. Kickoff 를 막지 않는다 — 운영자 판정 A1 이 Claude Stop 을 fail-closed 에 두었다. 측정은 run 초반(Q5 와 같은 자리)에 Q2 와 같은 방식(격리 설정, 기록형 Stop 훅, 연속 block 횟수 계수)으로 수행하고 progress.md 에 남긴다. 상한이 없다고 나오면 REQ-HSF-013 의 술어를 하네스 인자로 일반화할지 운영자 판정을 다시 받는다.
- **Q9 codex 모드 `moai hook agent` 경로의 도달성과 비대칭** — 배포 래퍼 `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh:47`(`.sh.tmpl:47` 도 같음)은 `--harness` 없이 부르고, `internal/template/templates/.codex/` 에는 `agents/` 만 있으며, 이 트리에서 `hook agent` 를 부르는 Codex 설정 생성 경로를 grep 으로 찾지 못했다(Codex 쪽 명령 행은 `internal/cli/codex_readiness.go:74` 의 `moai hook --harness codex`, 부재 확정은 아니다). 따라서 AC-HSF-012·013 의 codex 모드 32 경우는 현재 도달하지 않는 경로를 검증하고, 성공 경로의 agent 호출은 Codex 번역을 거치지 않아 도달한다면 실패 경로만 Codex 렌더링을 하는 비대칭이 된다. **범위 변경 없음**(Q3 판정의 포함 범위 유지). 도달 경로를 만들지, 성공 경로를 Codex 번역에 태울지는 별도 카드의 판단이며 판정만 기록한다.
- **Q7 Claude 하네스의 영속 기록면** — REQ-HSF-008 은 `codexadapter.RecordDiscards` 재사용을 요구하는데, 그 기록면의 경로가 `.moai/logs/codex-adapter.jsonl` 이라 Claude 하네스 사건이 「codex-adapter」 이름의 파일에 쌓인다. **기록면 하나를 재사용하는 쪽을 권장**(새 기록면을 만들면 관측 지점이 둘로 갈린다). 대안(Claude 경로는 stderr 만 남긴다)을 택하면 REQ-HSF-008(b) 와 AC-HSF-007(b) 를 먼저 개정해야 한다 — 판정 기록만으로는 대안을 택할 수 없다.

## §C 착수 전 점검 (Pre-flight)

1. t1099 착지 확인: `git merge-base --is-ancestor <t1099 착지 커밋> HEAD` → exit 0.
2. spec.md §F.1 의 심볼 diff 를 실행하고 결과를 progress.md §E.2 에 붙인다. 달라진 것이 있으면 구현 전에 spec.md §B·acceptance.md 를 먼저 고친다.
3. 설치된 codex-cli 버전을 기록한다(`codex --version`). Q2 측정 버전 0.156.1 과 다르면 Q2 를 같은 방법으로 다시 재고, 상한이 생겼다면 REQ-HSF-013 의 술어와 spec.md §B.2 행 3 을 먼저 개정한다. — 0.4.2: run 시점 설치본은 0.157.0 이었다. 리드 판정(2026-09-26)으로 이 run 에서는 재측정하지 않고 버전에 묶인 관측으로 남기며, 0.157.0 재측정은 후속 카드가 맡는다(spec.md §F.2 「Q2 의 버전 한정」).
4. 기준선 측정(구현 전, 별도 커밋): 수리 전 트리에서 (a) `pre-tool` 에 파손 4 형태를 넣었을 때 stdout 이 `{}` 이고 exit 0 이며 디스패치가 0회임을, (b) 관측 하위 명령 22개의 파싱 실패 출력(20개 `{}`, worktree-create·worktree-remove 빈 stdout)을, (c) `moai hook agent` 의 결정 매핑 action(`x-validation`, `x-pre-transformation`, `x-pre-implementation`, 미지 action `foo`)과 관측 매핑 action(`x-verification`, `x-post-transformation`, `x-post-implementation`, `x-completion`)의 파싱 실패 출력(모두 `{}` exit 0)과, 유효한 stdin(`{}`) + `--harness bogus` 로 `x-validation` 을 실행했을 때의 종료 코드와 stdout(plan-audit 2회차에서 codex 백엔드가 exit 0 + PreToolUse allow 출력을 재현했다 — 이 트리에서 다시 잰다; AC-HSF-012 신규 거부의 기준선), 같은 유효한 stdin(`{}`) + `--harness bogus` 로 관측 매핑 action `x-verification`·`x-completion` 을 실행했을 때의 종료 코드와 stdout(수리 전에는 exit 0 예상; AC-HSF-013 신규 거부의 기준선)을 재현하고 progress.md 에 기록한다 — 재현이 수리 커밋보다 앞선 커밋에 있어야 순서가 git 이력으로 증명된다(`verification-claim-integrity.md` §2.3).
5. **호스트의 설정 `env` 전파·유지 측정 — 완료(2026-09-25), 결과 「유지한다」 → 탈출 장치 제거.** 운영자 판정 Q1 의 조건이었던 측정이다. 절차 원문(변형 C·A1·A2·B1·B2, 판정 전제, 판정 네 결과, 정지 조건, 승인된 상한)은 0.3.4 판(커밋 `39ee312cf` 의 이 파일)에 있다.
   - 증거: `.moai/reports/t1152/preflight5.md` (로컬 증거, gitignored). 실행 기록: `/tmp/t1152-pf5/logs/` (머신 로컬 스크래치).
   - 실행: C(양성 대조) — 셸 환경의 탐침을 훅 기록 2줄 모두에서 봤다. A1(같은 세션, 키 추가 → 키 삭제) — 선언 중 기록 4줄이 탐침 `on`, 키를 지운 뒤 기록 3줄(`local_declares:false`)도 탐침 `on`, 셸 환경에는 키 없음(첫 기록 `probe:""`). 판정 전제(삭제 전 관측·삭제 뒤 기록)를 채웠다.
   - 판정: 1 「유지한다」. 판정 규칙상 1 이 우선하므로 A2·B1·B2 는 돌리지 않았다. 예산 사용 2/6 — 모델 호출 0회로 끝난 인증 실패 4건은 운영자 판정으로 예산에서 뺐다(증거 파일 §1).
   - 처분: 정지 조건대로 구현을 시작하지 않고 운영자 재판정을 받았다 — §B.1 Q1 「재판정」. 탈출 장치가 없어져 측정 대상도 사라졌으므로 run 에서 다시 재지 않는다.
   - 한계: 이 호스트·이 버전·비대화형 `-p`·`settings.local.json` 면의 키 삭제 변형 한 번의 관측이다. 결론(출처 제약이 우회된다)은 한 변형의 유지로 선다.

## §D 제약

- `internal/hook` 패키지를 수정하지 않는다. 이음매는 `internal/cli` 에 둔다.
- 결정 이벤트 목록을 새로 나열하지 않는다(REQ-HSF-002).
- 템플릿(`internal/template/templates/`) 변경 없음 — 동작 변경은 Go 바이너리 안에서만 일어난다. 문서 반영(hooks-system.md, 운영자 문서의 복구 절 — moai 갱신과 훅 비활성화)은 sync-phase 에서 판단한다.
- 결정 이벤트의 fail-closed 를 끄는 실행 시점 스위치(환경 변수·설정 파일·`.moai/config/` 키)를 두지 않는다(REQ-HSF-001, 0.4.0).
- 테스트는 `t.TempDir()` 과 `CLAUDE_PROJECT_DIR` 로 기록면을 격리한다. 전체 스위트(`go test ./...`)를 로컬에서 돌리지 않는다.

## §E 자기 검증

run 완료 보고는 acceptance.md 의 AC 마다 명령과 원문 출력을 붙인다. 특히 AC-HSF-003(단일 목록)은 동작 등가 테스트, Go AST 검사(두 진입점 모두), 여섯 변이(판정 호출 무력화 → 동작 테스트 RED, 중복 `switch` 삽입 → AST 검사 RED, 면제 술어 무력화 → 동작 테스트 RED, `runAgentHook` 의 리터럴 비교 대체 → AST 검사 RED, 판정 호출을 남긴 채 결과를 버리고 다른 술어로 분기 → AST 검사 RED, 파싱 실패 경로에 환경 변수 스위치 삽입 → AST 검사 (b3) RED) 모두를 요구한다 — 하나만으로는 「목록이 없다」를 증명하지 못한다.

## §F 마일스톤 (결정의 번복 가능성 순)

### M0 — Kickoff 차단 질문 판정 (Priority High, 사람) — 판정됨, Kickoff 승인됨(2026-09-24)

Q2(측정)·Q1(A1 + 메커니즘 (i))·Q3(포함)이 2026-09-24 에 판정됐다(§B.1). **0.4.0: Pre-flight 5 결과 「유지한다」에 따라 운영자가 2026-09-25 에 탈출 장치 없음으로 재판정했다(§B.1 Q1 「재판정」) — 이 절의 메커니즘 (i)·부재 해석·Pre-flight 5 관련 서술은 기록으로만 남는다.** Q3 포함에 따른 REQ·AC 추가는 0.3.0 개정에 담겼다. **Implementation Kickoff Approval 은 2026-09-24 에 받았다(0.3.4).** 출처는 t1152 레인에서 운영자가 AskUserQuestion 에 직접 한 답이다. 리드가 앞서 보낸 「Kickoff 는 이 지시로 갈음」 전달문은 리드가 철회했으며 출처가 아니다. 승인 내용: (a) N17·N18 을 측정 전에 닫는다(0.3.4 에서 닫음); (b) 측정 조건을 제안대로 승인한다 — 최대 6회, 실행당 `timeout -k 10 300` 과 `--max-turns 8`, 격리 `/tmp` 프로젝트에서 `--permission-mode bypassPermissions`, 「유지한다」·「미측정」이면 멈추고 다시 판정받는다; (c) N15·N16 은 run 에서 테스트를 더해 닫는다 — `--harness bogus` 거부를 agent action 8개 모두에 단언하고(N15), `runAgentHook` 의 `IsDecisionBearing` 결과를 뒤집거나 무력화하는 변이가 AC-HSF-012 를 RED 로 만들어야 한다(N16). 구현 커밋은 t1099 가 develop 에 착지하고 이 브랜치가 그것을 흡수한 뒤 시작한다(변경 없음). 그때 spec.md §F.1 의 심볼 diff 와 §F.1 의 3(codex 버전, `runAgentHook` 매핑, 설정 파일 형식)으로 설계 전제를 다시 확인한다. **Kickoff 승인 대상에는 Pre-flight 5 의 상한(실행 예산 최대 6회, 실행당 `timeout -k 10 300` 벽시계와 `--max-turns 8`)과 격리 `/tmp` 프로젝트에서의 `--permission-mode bypassPermissions` 를 함께 올렸고, 둘 다 승인됐다** — 라이브 모델 실행의 비용·범위 선언이며 manager-spec 이 정한 제안값이었기 때문이다(0.3.2 N9, 0.3.4 N18). 추적 항목 Q4~Q8 은 Kickoff 를 막지 않는다: Q5·Q8 은 run 초반 측정, Q4·Q6 은 판정 기록, Q7 은 대안을 택할 경우에만 REQ 개정.

### M1 — 결정 이벤트의 fail-closed 동작 (Priority High)

- 파싱 실패 분기에서 `codexadapter.IsDecisionBearing(event)` 로 갈라, 결정 이벤트는 하네스별 fail-closed 경로로, 관측 이벤트는 기존 경로로 보낸다.
- **Codex Stop 면제 술어**(REQ-HSF-012·013): `internal/codexadapter` 안에 「Codex 호스트에는 Stop 연속 차단 상한이 없다」를 나타내는 술어 하나를 두고, 디스패처는 `IsDecisionBearing(ev) && !<술어>(harness, ev)` 형태의 조합으로만 fail-closed 여부를 정한다. 술어 본문은 Stop 하나만 이름으로 가리키므로 AC-HSF-003(b2) 의 「결정 이벤트 식별자 2개 이상」 조건에 걸리지 않는다. 주석에 Q2 증거 경로·codex 버전·관측 범위와 「Codex 가 상한을 도입하면 여기를 다시 본다」를 적는다. 면제 쌍은 `{}` + exit 0, stderr 한 줄, 면제 전용 키의 `RecordDiscards` 기록 한 건을 낸다.
- Codex: `writeCodexFailClosed` 재사용. **재사용 범위** — 현재 서명(`writeCodexFailClosed(event, cause error)`)은 기록 키 `hook-fault`, Reason 고정 문구, `ContentLength = len(causeText)` 를 박아 두었고(`fabc33812:internal/cli/hook_codex_failclosed.go:44-49`), stderr 미러(`RecordDiscards` 의 `codex-adapter: dropped %q on %s (%d bytes): %s`)에는 하네스 모드·파싱 오류 원인이 없다. 그대로 쓰면 REQ-HSF-008(파싱 실패 전용 키, 하네스 모드·원인이 든 stderr 한 줄)과 REQ-HSF-008(b) 의 stdin 바이트 수를 만족하지 못한다. **권장: 작은 확장**
  - 기록 필드(키·길이·Reason)를 인자로 받는 형태로 서명을 넓히되, t1099 M2c 의 기존 호출부는 같은 값(`hook-fault`, `len(causeText)`, 기존 Reason)을 넘겨 동작이 바뀌지 않게 한다. t1099 의 fault injection 테스트가 확장 후에도 통과해야 한다.
  - 하네스 모드와 파싱 오류 원인이 든 stderr 한 줄은 CLI 호출부가 `RecordDiscards` 미러와 **별도로** 쓴다. 미러 형식은 바꾸지 않는다.
  - stdin 바이트 수는 `internal/hook` 을 건드리지 않고 CLI 쪽에서 `os.Stdin` 을 계수 리더로 감싸 `ReadInput` 에 넘겨 얻는다.
  - 주의: 이 함수의 소유자는 t1099 다. t1099 착지 전에 이 확장이 필요해지면 t1099 레인과 조율하고, 착지 후라면 이 브랜치에서 확장하되 `@MX:ANCHOR`(fan_in) 를 갱신한다. 확장 대신 파싱 실패 전용 작성기를 새로 두는 것도 허용되지만, 그 경우에도 출력은 `TranslateCodex` 를 거쳐야 한다(REQ-HSF-003).
- Claude: 번역 표 HarnessClaude fatal_error 행에서 렌더링. `TranslateCodex` 를 하네스 인자를 받는 형태로 일반화할지, CLI 쪽에서 `Lookup` + `Render` 를 직접 부를지는 구현 판단이다 — 어느 쪽이든 표를 거친다. 일반화한다면 t1099 의 `@MX:ANCHOR`(fan_in) 를 갱신한다.
- 사유 문구: `fail-closed` 표시, stdin 파싱 실패 고정 문구, 운영자 문서 식별자를 run-phase 상수로 정의한다. 그 밖의 안내(복구 절차 등)는 싣지 않는다(REQ-HSF-010, 0.4.0).
- Stop 경로에 `@MX:WARN`(REQ-HSF-009): Claude 쪽은 호스트 상한 의존과 그 근거가 미측정 독트린이라는 사실, Codex 쪽은 면제와 Q2 측정 근거를 적는다.

### M1b — `runAgentHook` 의 같은 처리 (Priority High)

- action → 이벤트 매핑(`internal/cli/hook.go:469-481`)과 하네스 모드 판정을 `ReadInput`(`:455`) **앞으로** 옮긴다. 매핑은 `args[0]` 만 쓰므로 stdin 이 필요 없다.
- 파싱 실패 분기(`:455-463`)에서 `IsDecisionBearing(event)` 로 갈라, 결정 매핑 action 은 M1 과 같은 fail-closed 경로로(같은 렌더링·같은 사유 상수·같은 기록 키), 관측 매핑 action 은 현재 동작(stderr 경고 + `HookProtocol.WriteOutput` 의 `{}` + exit 0)으로 보낸다. 매핑 switch 는 그대로 두되 결정 이벤트 목록 역할을 하지 않는다 — 판정은 여전히 `IsDecisionBearing` 한 곳이다.
- 기록(M2)은 `runHookEvent` 와 같은 함수를 공유한다. 두 진입점이 같은 파싱 실패 처리 함수를 부르게 하면 형제 수리 누락을 구조로 막을 수 있다 — 구조는 구현 판단이다.

### M2 — 기록과 복구 절차 문서화 (Priority High)

- 0.4.0: 탈출 장치를 없앴다(§B.1 Q1 「재판정」). 파싱 실패 처리 경로는 환경 변수·설정 파일을 읽지 않는다 — acceptance.md AC-HSF-003(b3)·(c6).
- 운영자 문서(sync-phase)에 복구 절을 둔다: 지속적인 파싱 실패(예: 호스트 형식 변화)에서 (1) `moai update` 로 바이너리를 갱신하거나 (2) 훅을 끈다(Claude Code `disableAllHooks`; Codex 쪽 방법은 이 SPEC 에서 재지 않았다). 잠김을 푼 뒤 훅을 다시 켜는 절차도 함께 적는다. 거부 사유에는 문서 식별자만 싣는다(REQ-HSF-010).
- stderr 한 줄 + `RecordDiscards` 기록 한 건(파싱 실패 전용 키, stdin 바이트 수). 페이로드 원문은 싣지 않는다(REQ-HSF-011). 재사용 범위는 M1 과 같다. 기록 키는 둘이 새로 생긴다 — 파싱 실패 fail-closed, Codex Stop 면제 — 둘 다 `hook-fault` 와 서로 구분된다(0.3.x 의 셋째 키 「탈출 장치 적용」은 0.4.0 에서 없어졌다).

### M3 — 하네스 모드 판정 순서 이동 (Priority Medium)

- `harnessModeIsCodex(cmd)` 를 `ReadInput` 앞으로 옮긴다. `runHookEvent` 에서 잘못된 `--harness` 값의 거부는 그대로 0 이 아닌 종료다 — 다만 이제 stdin 을 읽기 전에 거부된다. `runAgentHook` 은 지금 `--harness` 를 읽지 않으므로(`harnessModeIsCodex` 의 비테스트 호출처는 `internal/cli/hook.go:292` 하나), 이 함수에서는 하네스 판독과 잘못된 값의 거부가 **새 동작**이다 — 유효한 stdin 과 잘못된 `--harness` 의 agent 호출이 성공에서 실패로 바뀐다(REQ-HSF-005, AC-HSF-012). `validateCodexHarnessEvent` 는 페이로드가 필요하므로 파싱 성공 뒤에 남는다.

### M4 — 테스트 갱신과 추가 (Priority Medium)

- acceptance.md §D 의 기존 테스트 갱신 — `TestRunHookEvent_MalformedStdinGraceful`(`internal/cli/hook_protocol_fix_test.go:60`)와 `TestRunAgentHook_ReadInputError`(`internal/cli/misc_coverage_test.go:355`, action `test-validation` → PreToolUse, 주석 `:372-374` 가 「default output」을 약속).
- 결정 이벤트 × 하네스 × 파손 4 형태의 표 기반 테스트: fail-closed 28 경우(Claude 16, Codex 12)와 Codex Stop 면제 4 경우, 관측 22 × 하네스 2 × 파손 4 형태(176 경우) 특성 테스트, 단일 목록 등가 테스트(면제 술어를 거친 기대 집합)와 AST 검사, `runAgentHook` action 8종 × 하네스 2 × 파손 4 형태(64 경우) 테스트, 파싱 실패 경로의 환경 변수·파일 판독 부재 AST 검사(AC-HSF-003(b3)).

## §G 금지 패턴

- 결정 이벤트를 `[]hook.EventType{…}` 리터럴, 맵 리터럴, `switch` 의 case 나열, `||` 비교 사슬로 다시 나열하기.
- fail-closed JSON 을 `fmt.Sprintf` 나 map 리터럴로 손 조립하기(번역 표 우회).
- 거부를 exit 2 로 내기 — stdout JSON 이 무시되어 사유가 사라진다(`internal/cli/hook.go:362-365`).
- 파싱 실패 원문 페이로드를 기록·사유에 싣기.
- 원인·운영자 문서 식별자 외의 안내(복구 절차, `disableAllHooks` 등)를 거부 사유에 싣기.
- Codex Stop 면제를 이벤트 목록이나 하네스별 이벤트 목록으로 표현하기(REQ-HSF-013 의 술어 하나만 허용).
- `runAgentHook` 의 파싱 실패 처리를 `runHookEvent` 와 다른 렌더링·다른 사유 문구로 따로 구현하기.
- 결정 이벤트의 fail-closed 를 끄는 스위치를 어떤 출처(환경 변수, `.claude/settings*.json`, `.moai/config/`)에서든 읽기(0.4.0).
- 관측 이벤트의 출력·종료 코드를 바꾸기.

## §H 교차 참조

- spec.md §A.3 (위협 모델), §B (분류표), §F (전제·위험)
- `fabc33812:internal/codexadapter/decision.go`, `translate.go`, `diagnostics.go`, `output.go`, `internal/cli/hook_codex_failclosed.go`
- `.claude/rules/moai/core/hooks-system.md:213` (Stop 차단 상한 — exit 2 서술), `:388` (Can Block 목록)
- `.claude/rules/moai/workflow/goal-directive.md:13`, `internal/cli/hook_stop_goal.go:120-121` (JSON block 경로의 상한 근거)
- `6a3603274` (보존 대상 의도)
- `.moai/reports/t1152/plan-audit.md` (0.2.0 개정 근거, 로컬 증거)
- `.moai/reports/t1152/q2-codex-stop-cap.md` (0.3.0 Q2 판정 근거, 로컬 증거 — gitignored)
- `.moai/reports/t1152/plan-audit-iter2.md` (0.3.1 개정 근거, 로컬 증거)
- `.moai/reports/t1152/plan-audit-iter3.md` (0.3.2 개정 근거, 로컬 증거)
- `.moai/reports/t1152/plan-audit-iter4.md` (0.3.3 개정 근거, 로컬 증거)
- `.moai/reports/t1152/preflight5.md` (0.4.0 개정 근거 — Pre-flight 5 결과, 로컬 증거)
- `internal/cli/hook.go:447`·`:455-463`·`:469-481` (`runAgentHook`, 이 트리 `44dfc25fd` 기준), `internal/cli/hook.go:45` (`--harness` PersistentFlags)
