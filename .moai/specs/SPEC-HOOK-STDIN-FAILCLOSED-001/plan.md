---
id: SPEC-HOOK-STDIN-FAILCLOSED-001
title: "Plan — 훅 stdin 파싱 실패의 결정 이벤트 fail-closed"
version: "0.2.0"
created: 2026-09-24
author: manager-spec (card t1152)
---

# Plan — SPEC-HOOK-STDIN-FAILCLOSED-001

마일스톤은 **되돌리기 어려운 결정부터** 배치했다. 사람이 먼저 봐야 할 것은 M0(Kickoff 차단 질문 판정)과 M1(Stop·UserPromptSubmit 의 동작 결정)이며, 순서 이동·테스트 갱신 같은 기계적 작업은 뒤에 둔다. 우선순위 표기만 쓰고 기간 추정은 하지 않는다.

0.2.0 개정: plan-audit 1회차(`.moai/reports/t1152/plan-audit.md`)의 D0·D2·D3·D9·D10·D12 를 반영했다. 명확화 필요 표식은 Kickoff 차단 질문(Q2·Q1·Q3)에만 남기고, 나머지(Q4~Q7)는 표식 없는 추적 항목으로 옮겼다.

## §A 맥락

spec.md §A 참조. 요약: `internal/cli/hook.go:272-280` 의 stdin 파싱 실패 분기가 결정 이벤트에서도 `{}` + exit 0 을 내 가드 전체가 우회된다. t1099(`fabc33812`)의 결정 이벤트 집합과 번역 표를 재사용해 결정 이벤트 4개만 fail-closed 로 바꾸고, 관측 이벤트 26개는 `6a3603274` 의 의도대로 이벤트별 현재 출력을 보존한다(하위 명령 22개 중 20개 `{}`, worktree-create·worktree-remove 는 빈 stdout).

## §B 질문

### B.1 Kickoff 차단 질문 — Implementation Kickoff Approval 전에 판정

순서가 의미를 갖는다. Q1 의 옵션 A 가 Codex 에서 안전한지는 Q2 의 답에 달려 있으므로 Q2 를 먼저 둔다.

- **Q2** [NEEDS CLARIFICATION: Codex 의 Stop 연속 차단 상한 유무] — Claude 는 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`(기본 8)이 있다. 그 상한이 JSON `decision:"block"` + exit 0 에도 걸린다는 근거는 저장소 독트린(`goal-directive.md:13`, 그 평가기의 출력 형태 `internal/cli/hook_stop_goal.go:120-121`)이며, `hooks-system.md:213` 자체는 exit 2 차단을 서술한다 — 이 plan 은 JSON block 경로를 측정하지 않았다. Codex 에 대응하는 상한이 있는지는 확인하지 못했다. 없다면 Codex 에서 파싱 실패가 지속될 때 Stop 이 무기한 차단된다. t1099 의 `@MX:WARN`(`fabc33812:internal/cli/hook_codex_failclosed.go:28-29`)이 Stop 루프 상한을 「M2d Stop 체인의 몫」으로 넘겼다. 판정 방법은 둘 중 하나다:
  - (a) Kickoff 전에 측정한다: Codex 문서 또는 Codex 세션에서 Stop 훅이 연속 block 을 낼 때의 동작, 그리고 Claude 에서 JSON block + exit 0 이 상한에 걸리는지를 재고 그 명령·출력을 progress.md 에 남긴다.
  - (b) 측정 없이 Q1 을 아래 조건부 형태로 판정하고, 측정은 run Pre-flight 로 옮긴다.
- **Q1** [NEEDS CLARIFICATION: Stop·UserPromptSubmit 을 fail-closed 에 포함할지, 그리고 탈출 장치의 활성화 메커니즘] — 선택지:
  - **(A) 네 이벤트 전부 fail-closed + 탈출 장치** (권장, **Q2 에 조건부**). 근거: `DecisionBearingEvents()` 를 그대로 쓰므로 목록이 하나로 유지된다. 잠김은 소리가 나고(차단 사유가 모델·사용자에게 보인다), 우회는 소리가 나지 않는다. Claude 의 Stop 루프는 호스트 상한이 끊는다(Q2 의 근거 등급 참조). 비용: 지속적인 파싱 실패(호스트 형식 변화)에서 UserPromptSubmit 이 매 프롬프트를 차단해 운영자가 탈출 장치를 켤 때까지 세션을 쓸 수 없다.
    - 조건: **Q2 에서 Codex Stop 상한이 확인되지 않으면**, Codex 의 Stop 은 다음 중 하나로 처리한다 — (A1) Codex 하네스의 Stop 만 fail-closed 에서 제외하고 관측 경로(`{}` + 기록)로 둔다. 이 경우 「하네스 × 이벤트」 예외 하나가 생기므로 REQ-HSF-002 의 단일 목록과 긴장한다 — 예외는 이벤트 목록이 아니라 「Codex 에 Stop 상한이 없다」는 이름 붙은 술어로 `codexadapter` 안에 두고 REQ·AC 를 D-NEW-1 개정으로 추가한다. (A2) MoAI 자체 상한을 둔다. 다만 파싱 실패 상태에서는 `session_id` 를 읽을 수 없어 세션별 횟수를 셀 열쇠가 stdin 밖(환경 변수 등)에 있어야 하며, 그런 열쇠가 Codex 훅 환경에 있는지 측정하지 않았다 — A2 를 택하면 그 측정이 선행한다.
  - **(B) PreToolUse·PermissionRequest 만 fail-closed, Stop·UserPromptSubmit 은 fail-open + 기록**. 근거: 이 둘은 위험한 **동작**을 막는 게이트가 아니라 진행을 막는 게이트이고, 지속 실패의 잠김 비용이 크다. Q2 결과와 무관하게 안전하다. 비용: 결정 집합의 부분집합을 코드 어딘가에 새로 정의해야 한다 — 리드 제약(「두 번째 목록 금지」)과 긴장한다. 택한다면 부분집합은 번역 표의 Outcome 이나 별도 목록이 아니라 「이벤트가 동작 게이트인가」라는 이름 붙은 술어로 `codexadapter` 안에 둬야 하며, 그 술어의 정당화를 spec.md §B 표에 적어야 한다.
  - **탈출 장치의 활성화 메커니즘** — REQ-HSF-007 의 출처 제약(모델이 쓸 수 있는 `.claude/settings.json`·`.claude/settings.local.json` 의 `env` 블록과 `.moai/config/` 아래 파일에서 온 값은 무효)을 만족해야 한다. 후보:
    - (i) 환경 변수(이름은 `internal/config/envkeys.go` 상수) + **프로젝트·로컬 설정 파일의 `env` 블록에 같은 키가 선언돼 있으면 무시**. 디스패처가 두 설정 파일을 읽어 그 키의 존재를 확인한다. 정당한 경로는 호스트를 띄우는 셸 환경이다. 모델이 Bash 도구로 설정한 변수는 훅 프로세스에 닿지 않는다(훅 프로세스는 호스트가 띄운다). **권장.** 잔여 면: 사용자 범위 `~/.claude/settings.json` 의 `env`(REQ 가 명시한 면 밖), 그리고 설정 파일 판독 실패 시의 처리(무시가 아니라 「활성화 불인정」으로 닫아야 한다).
    - (ii) 프로젝트 밖 사용자 범위 파일(예: `~/.moai/` 아래). 모델의 파일 도구가 그 경로에 쓸 수 있는 권한 설정이면 출처 제약이 약해진다.
    - `.moai/config/sections/*.yaml` 키는 출처 제약에 정면으로 걸리므로 후보가 아니다.
    - 어느 쪽이든 거부 사유에는 그 식별자를 싣지 않는다(REQ-HSF-010). 활성화 절차는 운영자 문서에만 둔다. 이전 판이 이 면을 「ConfigChange 감시의 몫」으로 넘겼던 것은 철회한다 — spec.md §B.2 행 20 이 적었듯 ConfigChange 핸들러는 무조건 빈 출력이고 그런 감시는 존재하지 않는다.
- **Q3** [NEEDS CLARIFICATION: `runAgentHook` 포함 여부] — `internal/cli/hook.go:447-463` 에 같은 모양의 fail-open 분기가 있고, `-validation` / `-pre-transformation` / `-pre-implementation` 액션과 미지 액션은 PreToolUse 로 매핑된다. plan-audit 에서 codex 백엔드가 이 트리에서 `foo-validation` 에 파손 stdin 을 넣어 `rc=0, stdout={}` 를 재현했다(이 plan 저자가 직접 실행한 것은 아니다). 같은 결함 계열의 형제 경로를 남겨 두면 수리가 절반이 된다. **포함을 권장**하지만, 원 카드 범위 밖이므로 운영자 판정 사항이다. 포함 시 REQ·AC 를 추가하는 D-NEW-1 개정이 run 착수 전에 필요하다.

### B.2 추적 항목 — Kickoff 를 막지 않는다

- **Q4 빈 stdin 경로** — `ReadInput` 은 빈 stdin 을 오류가 아니라 기본 입력으로 처리한다(`internal/hook/protocol.go:52-56`). 결정 이벤트에서 빈 페이로드로 가드가 기본 입력을 받아 실행되는데, 가드가 빈 `tool_name` 을 허용으로 처리하면 같은 우회가 된다. 이 SPEC 은 다루지 않는다. 별도 카드로 측정할지는 리드가 판단하고 판정만 기록한다.
- **Q5 5 MiB 경로의 호스트 실현성** — 호스트(Claude Code, Codex)가 5 MiB 를 넘는 `tool_input` 을 훅 stdin 으로 실제 넘기는지 측정하지 않았다. 수리의 정당성과 무관하다: 모델이 제어하는 더 싼 경로(중첩 깊이 초과, 약 20 KB)의 파싱 실패가 측정됐다(spec.md §A.3). 측정은 sync-phase `--security --deep` 렌즈의 위협 등급 판단용으로 run 초반에 수행하고 progress.md 에 기록한다. 같은 자리에서 「호스트가 깊게 중첩된 `tool_input` 을 그대로 넘기는가」도 잰다.
- **Q6 Can Block 11개** — spec.md §B.3 의 이벤트 가운데 결정 집합으로 옮겨야 할 것이 있는지. 이 SPEC 은 옮기지 않는다. 옮기려면 `DecisionBearingEvents()` 를 바꾸는 별도 카드가 필요하다 — 판정만 기록한다.
- **Q7 Claude 하네스의 영속 기록면** — REQ-HSF-008 은 `codexadapter.RecordDiscards` 재사용을 요구하는데, 그 기록면의 경로가 `.moai/logs/codex-adapter.jsonl` 이라 Claude 하네스 사건이 「codex-adapter」 이름의 파일에 쌓인다. **기록면 하나를 재사용하는 쪽을 권장**(새 기록면을 만들면 관측 지점이 둘로 갈린다). 대안(Claude 경로는 stderr 만 남긴다)을 택하면 REQ-HSF-008(b) 와 AC-HSF-007(b) 를 먼저 개정해야 한다 — 판정 기록만으로는 대안을 택할 수 없다.

## §C 착수 전 점검 (Pre-flight)

1. t1099 착지 확인: `git merge-base --is-ancestor <t1099 착지 커밋> HEAD` → exit 0.
2. spec.md §F.1 의 심볼 diff 를 실행하고 결과를 progress.md §E.2 에 붙인다. 달라진 것이 있으면 구현 전에 spec.md §B·acceptance.md 를 먼저 고친다.
3. Q2 를 (b) 로 판정했다면 그 측정을 여기서 수행하고, Q1 의 조건부 분기 가운데 어느 쪽이 적용되는지 progress.md 에 기록한다.
4. 기준선 측정(구현 전, 별도 커밋): 수리 전 트리에서 (a) `pre-tool` 에 파손 4 형태를 넣었을 때 stdout 이 `{}` 이고 exit 0 이며 디스패치가 0회임을, (b) 관측 하위 명령 22개의 파싱 실패 출력(20개 `{}`, worktree-create·worktree-remove 빈 stdout)을 재현하고 progress.md 에 기록한다 — 재현이 수리 커밋보다 앞선 커밋에 있어야 순서가 git 이력으로 증명된다(`verification-claim-integrity.md` §2.3).

## §D 제약

- `internal/hook` 패키지를 수정하지 않는다. 이음매는 `internal/cli` 에 둔다.
- 결정 이벤트 목록을 새로 나열하지 않는다(REQ-HSF-002).
- 템플릿(`internal/template/templates/`) 변경 없음 — 동작 변경은 Go 바이너리 안에서만 일어난다. 문서 반영(hooks-system.md, 운영자 문서의 탈출 장치 절)은 sync-phase 에서 판단한다.
- 환경 변수 이름은 `internal/config/envkeys.go` 상수로 정의한다(하드코딩 금지).
- 테스트는 `t.TempDir()` 과 `CLAUDE_PROJECT_DIR` 로 기록면을 격리한다. 전체 스위트(`go test ./...`)를 로컬에서 돌리지 않는다.

## §E 자기 검증

run 완료 보고는 acceptance.md 의 AC 마다 명령과 원문 출력을 붙인다. 특히 AC-HSF-003(단일 목록)은 동작 등가 테스트, Go AST 검사, 두 변이(판정 호출 무력화 → 동작 테스트 RED, 중복 `switch` 삽입 → AST 검사 RED) 넷을 모두 요구한다 — 하나만으로는 「목록이 없다」를 증명하지 못한다.

## §F 마일스톤 (결정의 번복 가능성 순)

### M0 — Kickoff 차단 질문 판정 (Priority High, 사람)

Q2 → Q1 → Q3 순서로 판정한다. Q2 를 측정으로 닫지 못하면 Q1 은 조건부 형태(옵션 A 의 A1/A2 중 무엇을 택하는지까지)로 판정하고, Q2 측정은 Pre-flight 3 으로 넘긴다. Q1 에는 탈출 장치 메커니즘 선택이 포함된다. Q3 을 포함으로 판정하면 D-NEW-1 개정이 run 착수 전에 착지해야 한다. 추적 항목 Q4~Q7 은 Kickoff 를 막지 않는다: Q5 는 run 초반 측정, Q4·Q6 은 판정 기록, Q7 은 대안을 택할 경우에만 REQ 개정.

### M1 — 결정 이벤트의 fail-closed 동작 (Priority High)

- 파싱 실패 분기에서 `codexadapter.IsDecisionBearing(event)` 로 갈라, 결정 이벤트는 하네스별 fail-closed 경로로, 관측 이벤트는 기존 경로로 보낸다.
- Codex: `writeCodexFailClosed` 재사용. **재사용 범위** — 현재 서명(`writeCodexFailClosed(event, cause error)`)은 기록 키 `hook-fault`, Reason 고정 문구, `ContentLength = len(causeText)` 를 박아 두었고(`fabc33812:internal/cli/hook_codex_failclosed.go:44-49`), stderr 미러(`RecordDiscards` 의 `codex-adapter: dropped %q on %s (%d bytes): %s`)에는 하네스 모드·파싱 오류 원인이 없다. 그대로 쓰면 REQ-HSF-008(파싱 실패 전용 키, 하네스 모드·원인이 든 stderr 한 줄)과 REQ-HSF-008(b) 의 stdin 바이트 수를 만족하지 못한다. **권장: 작은 확장**
  - 기록 필드(키·길이·Reason)를 인자로 받는 형태로 서명을 넓히되, t1099 M2c 의 기존 호출부는 같은 값(`hook-fault`, `len(causeText)`, 기존 Reason)을 넘겨 동작이 바뀌지 않게 한다. t1099 의 fault injection 테스트가 확장 후에도 통과해야 한다.
  - 하네스 모드와 파싱 오류 원인이 든 stderr 한 줄은 CLI 호출부가 `RecordDiscards` 미러와 **별도로** 쓴다. 미러 형식은 바꾸지 않는다.
  - stdin 바이트 수는 `internal/hook` 을 건드리지 않고 CLI 쪽에서 `os.Stdin` 을 계수 리더로 감싸 `ReadInput` 에 넘겨 얻는다.
  - 주의: 이 함수의 소유자는 t1099 다. t1099 착지 전에 이 확장이 필요해지면 t1099 레인과 조율하고, 착지 후라면 이 브랜치에서 확장하되 `@MX:ANCHOR`(fan_in) 를 갱신한다. 확장 대신 파싱 실패 전용 작성기를 새로 두는 것도 허용되지만, 그 경우에도 출력은 `TranslateCodex` 를 거쳐야 한다(REQ-HSF-003).
- Claude: 번역 표 HarnessClaude fatal_error 행에서 렌더링. `TranslateCodex` 를 하네스 인자를 받는 형태로 일반화할지, CLI 쪽에서 `Lookup` + `Render` 를 직접 부를지는 구현 판단이다 — 어느 쪽이든 표를 거친다. 일반화한다면 t1099 의 `@MX:ANCHOR`(fan_in) 를 갱신한다.
- 사유 문구: `fail-closed` 표시, stdin 파싱 실패 고정 문구, 운영자 문서 식별자를 run-phase 상수로 정의한다. 탈출 장치 식별자는 싣지 않는다(REQ-HSF-010).
- Stop 경로에 `@MX:WARN`(REQ-HSF-009): 호스트 상한 의존 사실과 Q2 결과(근거 등급 포함)를 적는다.

### M2 — 탈출 장치와 기록 (Priority High)

- Q1 판정에 따른 탈출 장치. 기본 꺼짐. 출처 제약(REQ-HSF-007): 모델이 쓸 수 있는 설정면에서 온 값은 활성화로 인정하지 않는다. 켜졌을 때도 stderr·기록을 남긴다.
- stderr 한 줄 + `RecordDiscards` 기록 한 건(파싱 실패 전용 키, stdin 바이트 수). 페이로드 원문은 싣지 않는다(REQ-HSF-011). 재사용 범위는 M1 과 같다.

### M3 — 하네스 모드 판정 순서 이동 (Priority Medium)

- `harnessModeIsCodex(cmd)` 를 `ReadInput` 앞으로 옮긴다. 잘못된 `--harness` 값의 거부는 그대로 0 이 아닌 종료다 — 다만 이제 stdin 을 읽기 전에 거부된다. `validateCodexHarnessEvent` 는 페이로드가 필요하므로 파싱 성공 뒤에 남는다.

### M4 — 테스트 갱신과 추가 (Priority Medium)

- acceptance.md §D 의 기존 테스트 갱신.
- 결정 4 × 하네스 2 × 파손 4 형태(32 경우)의 표 기반 테스트, 관측 22 × 하네스 2 × 파손 4 형태(176 경우) 특성 테스트, 단일 목록 등가 테스트와 AST 검사, 탈출 장치 출처 제약 음성 테스트.

## §G 금지 패턴

- 결정 이벤트를 `[]hook.EventType{…}` 리터럴, 맵 리터럴, `switch` 의 case 나열, `||` 비교 사슬로 다시 나열하기.
- fail-closed JSON 을 `fmt.Sprintf` 나 map 리터럴로 손 조립하기(번역 표 우회).
- 거부를 exit 2 로 내기 — stdout JSON 이 무시되어 사유가 사라진다(`internal/cli/hook.go:362-365`).
- 파싱 실패 원문 페이로드를 기록·사유에 싣기.
- 탈출 장치의 식별자나 활성화 절차를 거부 사유에 싣기.
- 모델이 쓸 수 있는 설정면(`.claude/settings*.json` 의 `env`, `.moai/config/`)에서 탈출 장치 활성화를 읽기.
- 관측 이벤트의 출력·종료 코드를 바꾸기.

## §H 교차 참조

- spec.md §A.3 (위협 모델), §B (분류표), §F (전제·위험)
- `fabc33812:internal/codexadapter/decision.go`, `translate.go`, `diagnostics.go`, `output.go`, `internal/cli/hook_codex_failclosed.go`
- `.claude/rules/moai/core/hooks-system.md:213` (Stop 차단 상한 — exit 2 서술), `:388` (Can Block 목록)
- `.claude/rules/moai/workflow/goal-directive.md:13`, `internal/cli/hook_stop_goal.go:120-121` (JSON block 경로의 상한 근거)
- `6a3603274` (보존 대상 의도)
- `.moai/reports/t1152/plan-audit.md` (0.2.0 개정 근거, 로컬 증거)
