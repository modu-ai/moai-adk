---
id: SPEC-HOOK-STDIN-FAILCLOSED-001
title: "Plan — 훅 stdin 파싱 실패의 결정 이벤트 fail-closed"
version: "0.3.3"
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
- **Q1 — Stop·UserPromptSubmit 의 fail-closed 포함 여부와 탈출 장치 메커니즘: 판정됨 — 선택지 A1 + 메커니즘 (i).**
  - 출처: 운영자 판정, t1152 레인의 AskUserQuestion, 2026-09-24.
  - 판정 내용(동작): 결정 이벤트 4개(PreToolUse·PermissionRequest·Stop·UserPromptSubmit)를 두 하네스 모두에서 fail-closed 로 한다. **예외는 Codex 하네스의 Stop 하나** — 관측 경로(`{}` + exit 0)에 stderr 한 줄과 면제 기록을 더한다(spec.md REQ-HSF-012). 근거는 Q2 측정이다: Codex 에 상한이 없으므로 여기서 차단하면 파싱 실패가 지속되는 동안 턴이 끝나지 않는다.
  - 판정 내용(표현): 예외는 두 번째 이벤트 목록이 아니라 `internal/codexadapter` 안의 **이름 붙은 술어 하나**로 표현한다 — 「Codex 호스트에는 Stop 연속 차단 상한이 없다」(spec.md REQ-HSF-013). 향후 Codex 가 상한을 도입하면 그 술어 하나만 다시 보면 된다. 술어의 식별자 이름은 run-phase 가 정한다.
  - 판정 내용(탈출 장치): 후보 (i) — **셸 환경 변수 하나**. 이름은 run-phase 에서 `internal/config/envkeys.go` 상수로 정의한다. 같은 키가 `$CLAUDE_PROJECT_DIR/.claude/settings.json` 또는 `.claude/settings.local.json` 의 `env` 블록에 선언돼 있으면 환경 변수를 **무시**한다. 존재하는 설정 파일을 읽지 못하면 활성화를 **인정하지 않는다**(fail-closed). 파일 부재는 「선언 없음」으로 본다 — 이 구분은 판정 문구를 구현 가능하게 옮긴 manager-spec 의 해석으로 시작했고, 0.3.1 에서 운영자가 확인했다(아래 「부재 해석 확인」, spec.md REQ-HSF-016). 거부 사유는 키 이름을 싣지 않는다(REQ-HSF-010).
  - 기록된 잔여 면: 사용자 범위 `~/.claude/settings.json` 의 `env`. 호스트가 이 블록을 훅 환경으로 전파하면 셸 환경 변수와 구별되지 않는다(spec.md §F.2).
  - **부재 해석 확인: 판정됨 (2026-09-24).** 출처: 운영자 판정, t1152 레인의 AskUserQuestion, plan-audit 2회차 N1 후속. 내용: 설정 파일이 **없으면** 「선언 없음」으로 보고, 탈출 장치를 셸 환경 변수로 인정할 수 있다. 조건 두 가지 — (1) 「추가 후 제거」 우회(설정 `env` 에 키 추가 → 호스트 반영 → 키 삭제 또는 파일 삭제 → 파손 입력)를 잔여 위험으로 기록한다(spec.md §F.2 「추가 후 제거」 행 — 키 삭제 변형은 부재 해석과 무관하게 존재하고, 파일 삭제 변형은 부재 해석이 추가로 연다). (2) run 시작 시 호스트 동작을 측정한다(§C Pre-flight 5). 호스트가 키·파일을 지운 뒤에도 그 `env` 값을 훅 환경에 유지한다고 나오면, 구현을 시작하지 않고 운영자에게 돌아가 재판정을 받는다. 이 조건을 집행하는 판정 규칙은 §C Pre-flight 5 에 있다 — 0.3.2 에서 삭제 전 관측을 판정 전제로 두고, 값이 어느 경로로도 오지 않는 「전달 없음」을 따로 닫도록 고쳤다(plan-audit 3회차 N8).
  - 판정 범위 밖의 선택지(기록만, 채택 아님): 한 세션에서 한 번이라도 선언이 관측된 키는 이후 선언이 사라져도 계속 불인정하는 강화. 파싱 실패 상태에서는 stdin 의 `session_id` 를 읽을 수 없어 stdin 밖의 세션 열쇠와 상태 저장이 필요하다 — 선택지 A2 와 같은 미측정 전제다. Pre-flight 5 가 「유지한다」로 나와 재판정할 때 후보로 올린다.
  - 판정 전 선택지(기록):
    - (A) 네 이벤트 전부 fail-closed + 탈출 장치 — Q2 에 조건부. Codex Stop 상한이 확인되지 않으면 (A1) Codex Stop 만 관측 경로로 두고 예외를 이름 붙은 술어로 표현, 또는 (A2) MoAI 자체 상한(세션 열쇠가 stdin 밖에 있어야 하며 미측정). **A1 이 채택됐다.**
    - (B) PreToolUse·PermissionRequest 만 fail-closed, Stop·UserPromptSubmit 은 fail-open + 기록. 결정 집합의 부분집합을 새로 정의해야 해 단일 목록 제약과 긴장. 채택되지 않았다.
    - 탈출 장치 후보: (i) 환경 변수 + 프로젝트·로컬 `env` 선언 시 무시 — **채택**; (ii) 프로젝트 밖 사용자 범위 파일(예: `~/.moai/` 아래) — 채택되지 않았다; `.moai/config/sections/*.yaml` 키는 출처 제약에 정면으로 걸려 애초에 후보가 아니었다. 이전 판이 이 면을 「ConfigChange 감시의 몫」으로 넘겼던 것은 0.2.0 에서 철회했다.
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
3. 설치된 codex-cli 버전을 기록한다(`codex --version`). Q2 측정 버전 0.156.1 과 다르면 Q2 를 같은 방법으로 다시 재고, 상한이 생겼다면 REQ-HSF-013 의 술어와 spec.md §B.2 행 3 을 먼저 개정한다.
4. 기준선 측정(구현 전, 별도 커밋): 수리 전 트리에서 (a) `pre-tool` 에 파손 4 형태를 넣었을 때 stdout 이 `{}` 이고 exit 0 이며 디스패치가 0회임을, (b) 관측 하위 명령 22개의 파싱 실패 출력(20개 `{}`, worktree-create·worktree-remove 빈 stdout)을, (c) `moai hook agent` 의 결정 매핑 action(`x-validation`, `x-pre-transformation`, `x-pre-implementation`, 미지 action `foo`)과 관측 매핑 action(`x-verification`, `x-post-transformation`, `x-post-implementation`, `x-completion`)의 파싱 실패 출력(모두 `{}` exit 0)과, 유효한 stdin(`{}`) + `--harness bogus` 로 `x-validation` 을 실행했을 때의 종료 코드와 stdout(plan-audit 2회차에서 codex 백엔드가 exit 0 + PreToolUse allow 출력을 재현했다 — 이 트리에서 다시 잰다; AC-HSF-012 신규 거부의 기준선), 같은 유효한 stdin(`{}`) + `--harness bogus` 로 관측 매핑 action `x-verification`·`x-completion` 을 실행했을 때의 종료 코드와 stdout(수리 전에는 exit 0 예상; AC-HSF-013 신규 거부의 기준선)을 재현하고 progress.md 에 기록한다 — 재현이 수리 커밋보다 앞선 커밋에 있어야 순서가 git 이력으로 증명된다(`verification-claim-integrity.md` §2.3).
5. **호스트의 설정 `env` 전파·유지 측정 (운영자 판정 Q1 의 조건, spec.md §F.2 「추가 후 제거」).** 구현 전에, 4 와 같은 자리에서 수행한다. 0.3.2 에서 plan-audit 3회차 N8·N9 에 따라 다시 짰다 — 종전 판은 삭제 전 관측을 요구하지 않아, 호스트가 설정 `env` 를 세션 시작 때만 적용하면 공허하게 「유지하지 않는다」가 나올 수 있었다.
   - 설정: 개발 프로젝트가 아니라 `/tmp` 아래 격리 프로젝트에서 한다(설정 파일을 실제로 고치는 측정이라 `CLAUDE.local.md` §13 과 같은 원칙). 훅이 호출될 때마다 자기 프로세스 환경의 탈출 장치 키 값과, 그 순간 두 설정 파일의 존재 여부·`env` 선언 여부를 한 줄씩 기록하는 기록형 훅을 두고 Claude Code 를 비대화형으로 실행한다. 양성 대조를 뺀 모든 실행에서 셸 환경에는 키를 두지 않아, 훅이 본 값은 설정 출처에서 온 것뿐이게 한다. 세션 안에서 일어나는 키 추가·삭제와 파일 삭제는 세션 안의 도구 호출(모델이 실행하는 셸 명령)로 한다 — 위협 모형이 「모델이 설정 파일을 쓰고 지운다」이기 때문이다. 그 편집·삭제가 권한 확인에 막히지 않도록 모든 실행은 격리 프로젝트에서 `--permission-mode bypassPermissions` 로 띄운다 — 그래도 삭제가 실행되지 않은 변형은 아래 판정 전제에 따라 「미측정」이다. 선언이 있었는지는 타이밍이 아니라 각 기록 줄의 선언 여부 필드로 가른다.
   - **실행 1회의 정의:** `claude -p` 프로세스 호출 하나. 그 호출이 이어 받는 세션(`--resume`·`--continue`)이 있어도 프로세스가 하나면 1회이고, 한 시나리오가 프로세스 둘을 쓰면 2회로 센다. 아래 설계는 모든 시나리오를 프로세스 하나로 끝내도록 짰다.
   - **실행 목록 — 두 전파 모형을 모두 덮는다:**
     - C 양성 대조: 셸 환경에만 키를 두고 1회. 기록형 훅이 값을 봐야 계측기가 동작한 것이다. 못 보면 아래 어떤 실행의 결과도 판정에 쓰지 않는다(전체 「미측정」).
     - A 같은 세션 전파: 설정 파일에 키 없이 세션을 시작하고, 세션 안에서 키를 추가한 뒤 훅이 여러 번 호출되게 하고, 이어서 A1 은 키를 삭제, A2 는 설정 파일을 삭제한 뒤 다시 훅이 호출되게 한다. 각 1회. A1 에서 선언이 있는 동안의 훅 호출이 값을 한 번도 보지 못하면 같은 세션 전파는 값을 전달하지 않는 것이므로 A2 는 돌리지 않는다.
     - B 세션 시작 전파: 실행 전에 설정 파일의 `env` 에 키를 선언해 두고 새 `claude -p` 프로세스를 띄운다(「키 추가 → 그 값을 들고 새 세션 시작」). 세션 안의 훅 호출로 값을 먼저 확인한 뒤, B1 은 키를 삭제, B2 는 설정 파일을 삭제하고 이어지는 훅 호출을 본다. 각 1회.
   - **[HARD] 판정 전제 — 삭제 전 관측과 삭제 뒤 기록:** 한 변형 실행의 「삭제 뒤」 관측은, 같은 실행에서 삭제 **전에** 선언이 있는 상태의 훅 호출이 설정 출처의 값을 본 경우에만 판정에 쓴다. 삭제 전에 값을 보지 못한 실행은 결코 「유지하지 않는다」의 근거가 아니다. 그 실행의 결과는 선언이 있는 동안의 기록이 한 줄 이상 있고 모두 값이 없으면 그 경로의 「전달 없음」, 그런 기록이 없으면(상한 도달, 편집 거부, 훅 미호출) 「미측정」이다. 또 삭제 뒤 관측은 같은 실행에 삭제 **뒤** 기록 — 선언 여부 필드가 선언이 사라졌음(A1·B1: 키 없음, A2·B2: 파일 없음)을 보이는 훅 호출 기록 — 이 한 줄 이상 있을 때만 판정에 쓴다. 이 필드가 삭제가 실제로 실행됐다는 확인이다. 삭제 전에 값을 봤더라도 삭제가 실행되지 않았거나(거부·오류) 삭제 뒤 기록이 0건이면, 그 변형은 「유지한다」의 근거도 「유지하지 않는다」의 근거도 아닌 「미측정」이다. 경로의 확정: 어느 변형에서든 선언이 있는 동안 값을 봤으면 그 경로는 「전달」, 그렇지 않고 그 경로에서 돌린 변형(A 는 A1, B 는 B1·B2)마다 선언이 있는 동안의 기록이 한 줄 이상 있으면 「전달 없음」, 나머지는 「미확정」이다.
   - **판정 — 앞의 것이 우선한다:**
     1. 「유지한다」 — 값을 전달한 경로의 어느 변형에서든, 선언이 사라진 뒤의 훅 호출에 값이 남아 있다.
     2. 「미측정」 — 양성 대조가 값을 보지 못했거나, 값을 전달한 경로의 어느 변형이 판정 전제(삭제 전 관측과 삭제 뒤 기록)를 채우지 못했거나, A·B 가운데 어느 한 경로라도 「전달」과 「전달 없음」 가운데 하나로 확정되지 않았다.
     3. 「유지하지 않는다」 — A·B 두 경로가 모두 확정됐고, 값을 전달한 경로가 하나 이상 있으며, 그 경로마다 두 변형 모두 판정 전제(삭제 전 관측과 삭제 뒤 기록)를 채운 채 삭제 뒤 값이 없다. 판정은 실제로 값을 전달한 경로에서만 내리며, 전달하지 않은 경로의 실행은 판정에 쓰지 않는다.
     4. 「전달 없음 — 설정 `env` 가 훅 환경에 닿지 않는다」 — 양성 대조가 값을 봤고, A·B 두 경로가 모두 「전달 없음」으로 확정됐다 — A1·B1·B2 모두 선언이 있는 동안의 훅 호출 기록이 한 줄 이상 있는데 어느 것도 값을 보지 못했다. 이 경우 「추가 후 제거」 우회는 이 호스트·버전에서 성립 조건이 없다(값이 애초에 훅 환경에 오지 않는다). 결과는 버전에 묶인 관측이므로, REQ-HSF-007 의 출처 제약은 그대로 두고(버전이 바뀌면 전파가 생길 수 있다) spec.md §F.2 「추가 후 제거」 행의 성립 조건을 「이 버전에서 관측되지 않음」으로 적는다.
   - **정지 조건:** 「유지한다」 또는 「미측정」이면 구현을 시작하지 않고 리드를 거쳐 운영자 재판정을 받는다(메커니즘 (i) 자체의 재검토 포함). 「유지하지 않는다」 또는 「전달 없음」이면 진행하고, 판정·Claude Code 버전·실행별 형태와 기록(삭제 전 관측 줄과 삭제 뒤 기록 줄 포함)·선언한 상한과 실제로 쓴 실행 수를 progress.md 에 남긴다.
   - **상한 — 제안값이다.** manager-spec 이 정했고 아직 누구도 승인하지 않았다. Kickoff 때 운영자가 함께 승인한다(§F M0). 먼저 선언하고 그 안에서 끝낸다.
     - 실행 예산: 최대 6회 = C 1 + A1 1 + A2 1 + B1 1 + B2 1 + 재시도 1. A2 를 건너뛰면 5회 안에서 끝난다. 재시도는 전체에서 1회뿐이며, 쓴 뒤에는 더 돌리지 않는다.
     - 실행마다 `timeout -k 10 300 claude -p --max-turns 8 …` 형태로 건다. 벽시계 300초가 지나면 TERM 을 보내고, 10초 유예 뒤에도 살아 있으면 KILL 한다. `timeout` 은 이 호스트에 있다(2026-09-24 관측: `command -v timeout gtimeout` → `/opt/homebrew/bin/timeout`, `/opt/homebrew/bin/gtimeout`; `timeout --version` → `timeout (GNU coreutils) 9.9`; `timeout --help` 에 `-k, --kill-after=DURATION`). run 착수 시 다시 확인하고, 없으면 같은 의미(TERM, 유예, KILL)의 외부 래퍼를 쓰되 그 형태를 progress.md 에 적는다.
     - `--max-turns` 는 숨은 플래그다. Claude Code 2.1.281 의 `claude --help` 에는 나오지 않고(`claude --help | grep -cE 'max-turns'` → `0`), 설치 바이너리 `/Users/goos/.local/share/claude/versions/2.1.281` 에는 `"--max-turns <turns>"` 정의가 있다(2026-09-24 관측). 실행 시 이 플래그가 거부되면 그 실행은 예산 1회를 쓴 「미측정」으로 기록하고, 이후 실행은 플래그 없이 건다 — 구속력 있는 상한은 언제나 바깥의 벽시계다.
     - 상한에 닿은 실행은 「미측정」으로 기록한다.
   - 이 항목이 재지 않는 것: Codex 호스트의 설정면(`.codex/` 아래)이 훅 환경에 변수를 공급하는지 — spec.md §F.2 「출처 제약의 잔여 면」 행에 이미 미측정 잔여 면으로 남아 있다.

## §D 제약

- `internal/hook` 패키지를 수정하지 않는다. 이음매는 `internal/cli` 에 둔다.
- 결정 이벤트 목록을 새로 나열하지 않는다(REQ-HSF-002).
- 템플릿(`internal/template/templates/`) 변경 없음 — 동작 변경은 Go 바이너리 안에서만 일어난다. 문서 반영(hooks-system.md, 운영자 문서의 탈출 장치 절)은 sync-phase 에서 판단한다.
- 환경 변수 이름은 `internal/config/envkeys.go` 상수로 정의한다(하드코딩 금지).
- 탈출 장치 환경 변수를 켜는 테스트는 `t.Setenv` 를 쓰므로 비병렬로 둔다.
- 테스트는 `t.TempDir()` 과 `CLAUDE_PROJECT_DIR` 로 기록면을 격리한다. 전체 스위트(`go test ./...`)를 로컬에서 돌리지 않는다.

## §E 자기 검증

run 완료 보고는 acceptance.md 의 AC 마다 명령과 원문 출력을 붙인다. 특히 AC-HSF-003(단일 목록)은 동작 등가 테스트, Go AST 검사(두 진입점 모두), 다섯 변이(판정 호출 무력화 → 동작 테스트 RED, 중복 `switch` 삽입 → AST 검사 RED, 면제 술어 무력화 → 동작 테스트 RED, `runAgentHook` 의 리터럴 비교 대체 → AST 검사 RED, 판정 호출을 남긴 채 결과를 버리고 다른 술어로 분기 → AST 검사 RED) 모두를 요구한다 — 하나만으로는 「목록이 없다」를 증명하지 못한다.

## §F 마일스톤 (결정의 번복 가능성 순)

### M0 — Kickoff 차단 질문 판정 (Priority High, 사람) — 판정됨, Kickoff 는 대기

Q2(측정)·Q1(A1 + 메커니즘 (i))·Q3(포함)이 2026-09-24 에 판정됐다(§B.1). Q3 포함에 따른 REQ·AC 추가는 0.3.0 개정에 담겼다. 남은 것은 **Implementation Kickoff Approval 자체**이며, 리드 지시에 따라 t1099 가 develop 에 착지한 뒤 요청한다. 그때 spec.md §F.1 의 심볼 diff 와 §F.1 의 3(codex 버전, `runAgentHook` 매핑, 설정 파일 형식)으로 설계 전제를 다시 확인한다. 0.3.1 에서 운영자가 파일 부재 해석을 확인했고(§B.1 Q1), 그 조건인 호스트 `env` 전파·유지 측정은 run 시작 시 §C Pre-flight 5 가 수행한다 — 「유지한다」 또는 「미측정」이면 run 은 거기서 멈춘다. **Kickoff 요청에는 Pre-flight 5 의 상한(실행 예산 최대 6회, 실행당 `timeout -k 10 300` 벽시계와 `--max-turns 8`)을 승인 대상으로 함께 올린다** — 라이브 모델 실행의 비용·범위 선언인데 manager-spec 이 정한 제안값이고 아직 승인되지 않았기 때문이다(0.3.2, plan-audit 3회차 N9). 추적 항목 Q4~Q8 은 Kickoff 를 막지 않는다: Q5·Q8 은 run 초반 측정, Q4·Q6 은 판정 기록, Q7 은 대안을 택할 경우에만 REQ 개정.

### M1 — 결정 이벤트의 fail-closed 동작 (Priority High)

- 파싱 실패 분기에서 `codexadapter.IsDecisionBearing(event)` 로 갈라, 결정 이벤트는 하네스별 fail-closed 경로로, 관측 이벤트는 기존 경로로 보낸다.
- **Codex Stop 면제 술어**(REQ-HSF-012·013): `internal/codexadapter` 안에 「Codex 호스트에는 Stop 연속 차단 상한이 없다」를 나타내는 술어 하나를 두고, 디스패처는 `IsDecisionBearing(ev) && !<술어>(harness, ev)` 형태의 조합으로만 fail-closed 여부를 정한다. 술어 본문은 Stop 하나만 이름으로 가리키므로 AC-HSF-003(b2) 의 「결정 이벤트 식별자 2개 이상」 조건에 걸리지 않는다. 주석에 Q2 증거 경로·codex 버전·관측 범위와 「Codex 가 상한을 도입하면 여기를 다시 본다」를 적는다. 면제 쌍은 `{}` + exit 0, stderr 한 줄, 면제 전용 키의 `RecordDiscards` 기록 한 건을 낸다.
- Codex: `writeCodexFailClosed` 재사용. **재사용 범위** — 현재 서명(`writeCodexFailClosed(event, cause error)`)은 기록 키 `hook-fault`, Reason 고정 문구, `ContentLength = len(causeText)` 를 박아 두었고(`fabc33812:internal/cli/hook_codex_failclosed.go:44-49`), stderr 미러(`RecordDiscards` 의 `codex-adapter: dropped %q on %s (%d bytes): %s`)에는 하네스 모드·파싱 오류 원인이 없다. 그대로 쓰면 REQ-HSF-008(파싱 실패 전용 키, 하네스 모드·원인이 든 stderr 한 줄)과 REQ-HSF-008(b) 의 stdin 바이트 수를 만족하지 못한다. **권장: 작은 확장**
  - 기록 필드(키·길이·Reason)를 인자로 받는 형태로 서명을 넓히되, t1099 M2c 의 기존 호출부는 같은 값(`hook-fault`, `len(causeText)`, 기존 Reason)을 넘겨 동작이 바뀌지 않게 한다. t1099 의 fault injection 테스트가 확장 후에도 통과해야 한다.
  - 하네스 모드와 파싱 오류 원인이 든 stderr 한 줄은 CLI 호출부가 `RecordDiscards` 미러와 **별도로** 쓴다. 미러 형식은 바꾸지 않는다.
  - stdin 바이트 수는 `internal/hook` 을 건드리지 않고 CLI 쪽에서 `os.Stdin` 을 계수 리더로 감싸 `ReadInput` 에 넘겨 얻는다.
  - 주의: 이 함수의 소유자는 t1099 다. t1099 착지 전에 이 확장이 필요해지면 t1099 레인과 조율하고, 착지 후라면 이 브랜치에서 확장하되 `@MX:ANCHOR`(fan_in) 를 갱신한다. 확장 대신 파싱 실패 전용 작성기를 새로 두는 것도 허용되지만, 그 경우에도 출력은 `TranslateCodex` 를 거쳐야 한다(REQ-HSF-003).
- Claude: 번역 표 HarnessClaude fatal_error 행에서 렌더링. `TranslateCodex` 를 하네스 인자를 받는 형태로 일반화할지, CLI 쪽에서 `Lookup` + `Render` 를 직접 부를지는 구현 판단이다 — 어느 쪽이든 표를 거친다. 일반화한다면 t1099 의 `@MX:ANCHOR`(fan_in) 를 갱신한다.
- 사유 문구: `fail-closed` 표시, stdin 파싱 실패 고정 문구, 운영자 문서 식별자를 run-phase 상수로 정의한다. 탈출 장치 식별자는 싣지 않는다(REQ-HSF-010).
- Stop 경로에 `@MX:WARN`(REQ-HSF-009): Claude 쪽은 호스트 상한 의존과 그 근거가 미측정 독트린이라는 사실, Codex 쪽은 면제와 Q2 측정 근거를 적는다.

### M1b — `runAgentHook` 의 같은 처리 (Priority High)

- action → 이벤트 매핑(`internal/cli/hook.go:469-481`)과 하네스 모드 판정을 `ReadInput`(`:455`) **앞으로** 옮긴다. 매핑은 `args[0]` 만 쓰므로 stdin 이 필요 없다.
- 파싱 실패 분기(`:455-463`)에서 `IsDecisionBearing(event)` 로 갈라, 결정 매핑 action 은 M1 과 같은 fail-closed 경로로(같은 렌더링·같은 사유 상수·같은 기록 키), 관측 매핑 action 은 현재 동작(stderr 경고 + `HookProtocol.WriteOutput` 의 `{}` + exit 0)으로 보낸다. 매핑 switch 는 그대로 두되 결정 이벤트 목록 역할을 하지 않는다 — 판정은 여전히 `IsDecisionBearing` 한 곳이다.
- 탈출 장치(M2)와 기록(M2)은 `runHookEvent` 와 같은 함수를 공유한다. 두 진입점이 같은 파싱 실패 처리 함수를 부르게 하면 형제 수리 누락을 구조로 막을 수 있다 — 구조는 구현 판단이다.

### M2 — 탈출 장치와 기록 (Priority High)

- 탈출 장치(Q1 메커니즘 (i), REQ-HSF-007·016). 기본 꺼짐. 판정 순서: (1) 환경 변수가 켜짐 값인가 → 아니면 꺼짐. (2) `$CLAUDE_PROJECT_DIR/.claude/settings.json`·`.claude/settings.local.json` 각각에 대해 — 부재면 「선언 없음」, 존재하는데 읽기·JSON 해석 실패면 **불인정**(stderr 한 줄), 해석되면 `env` 블록에 같은 키가 있는지 본다 → 있으면 **무시**. (3) 두 파일 모두 선언 없음이면 적용 — `{}` + exit 0, stderr 한 줄, 탈출 적용 전용 키의 기록 한 건. 환경 변수 이름은 `internal/config/envkeys.go` 상수. `.moai/config/` 는 읽지 않는다.
- 운영자 문서(sync-phase)에 활성화 절차와 잔여 면(사용자 범위 `~/.claude/settings.json` 의 `env`)을 적는다. 거부 사유에는 문서 식별자만 싣는다.
- stderr 한 줄 + `RecordDiscards` 기록 한 건(파싱 실패 전용 키, stdin 바이트 수). 페이로드 원문은 싣지 않는다(REQ-HSF-011). 재사용 범위는 M1 과 같다. 기록 키는 셋이 새로 생긴다 — 파싱 실패 fail-closed, Codex Stop 면제, 탈출 장치 적용 — 모두 `hook-fault` 와 서로 구분된다.

### M3 — 하네스 모드 판정 순서 이동 (Priority Medium)

- `harnessModeIsCodex(cmd)` 를 `ReadInput` 앞으로 옮긴다. `runHookEvent` 에서 잘못된 `--harness` 값의 거부는 그대로 0 이 아닌 종료다 — 다만 이제 stdin 을 읽기 전에 거부된다. `runAgentHook` 은 지금 `--harness` 를 읽지 않으므로(`harnessModeIsCodex` 의 비테스트 호출처는 `internal/cli/hook.go:292` 하나), 이 함수에서는 하네스 판독과 잘못된 값의 거부가 **새 동작**이다 — 유효한 stdin 과 잘못된 `--harness` 의 agent 호출이 성공에서 실패로 바뀐다(REQ-HSF-005, AC-HSF-012). `validateCodexHarnessEvent` 는 페이로드가 필요하므로 파싱 성공 뒤에 남는다.

### M4 — 테스트 갱신과 추가 (Priority Medium)

- acceptance.md §D 의 기존 테스트 갱신 — `TestRunHookEvent_MalformedStdinGraceful`(`internal/cli/hook_protocol_fix_test.go:60`)와 `TestRunAgentHook_ReadInputError`(`internal/cli/misc_coverage_test.go:355`, action `test-validation` → PreToolUse, 주석 `:372-374` 가 「default output」을 약속).
- 결정 이벤트 × 하네스 × 파손 4 형태의 표 기반 테스트: fail-closed 28 경우(Claude 16, Codex 12)와 Codex Stop 면제 4 경우, 관측 22 × 하네스 2 × 파손 4 형태(176 경우) 특성 테스트, 단일 목록 등가 테스트(면제 술어를 거친 기대 집합)와 AST 검사, `runAgentHook` action 8종 × 하네스 2 × 파손 4 형태(64 경우) 테스트, 탈출 장치 양성 테스트와 출처 제약·판독 실패 음성 테스트.

## §G 금지 패턴

- 결정 이벤트를 `[]hook.EventType{…}` 리터럴, 맵 리터럴, `switch` 의 case 나열, `||` 비교 사슬로 다시 나열하기.
- fail-closed JSON 을 `fmt.Sprintf` 나 map 리터럴로 손 조립하기(번역 표 우회).
- 거부를 exit 2 로 내기 — stdout JSON 이 무시되어 사유가 사라진다(`internal/cli/hook.go:362-365`).
- 파싱 실패 원문 페이로드를 기록·사유에 싣기.
- 탈출 장치의 식별자나 활성화 절차를 거부 사유에 싣기.
- Codex Stop 면제를 이벤트 목록이나 하네스별 이벤트 목록으로 표현하기(REQ-HSF-013 의 술어 하나만 허용).
- `runAgentHook` 의 파싱 실패 처리를 `runHookEvent` 와 다른 렌더링·다른 사유 문구로 따로 구현하기.
- 모델이 쓸 수 있는 설정면(`.claude/settings*.json` 의 `env`, `.moai/config/`)에서 탈출 장치 활성화를 읽기.
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
- `internal/cli/hook.go:447`·`:455-463`·`:469-481` (`runAgentHook`, 이 트리 `44dfc25fd` 기준), `internal/cli/hook.go:45` (`--harness` PersistentFlags)
