---
id: SPEC-HOOK-STDIN-FAILCLOSED-001
title: "Plan — 훅 stdin 파싱 실패의 결정 이벤트 fail-closed"
version: "0.1.0"
created: 2026-09-24
author: manager-spec (card t1152)
---

# Plan — SPEC-HOOK-STDIN-FAILCLOSED-001

마일스톤은 **되돌리기 어려운 결정부터** 배치했다. 사람이 먼저 봐야 할 것은 M0(미해결 질문 판정)과 M1(Stop·UserPromptSubmit 의 동작 결정)이며, 순서 이동·테스트 갱신 같은 기계적 작업은 뒤에 둔다. 우선순위 표기만 쓰고 기간 추정은 하지 않는다.

## §A 맥락

spec.md §A 참조. 요약: `internal/cli/hook.go:272-280` 의 stdin 파싱 실패 분기가 결정 이벤트에서도 `{}` + exit 0 을 내 가드 전체가 우회된다. t1099(`fabc33812`)의 결정 이벤트 집합과 번역 표를 재사용해 결정 이벤트 4개만 fail-closed 로 바꾸고, 관측 이벤트 26개는 `6a3603274` 의 의도대로 보존한다.

## §B 미해결 질문 — Implementation Kickoff Approval 전에 판정

- **Q1** [NEEDS CLARIFICATION: Stop·UserPromptSubmit 을 fail-closed 에 포함할지] — 선택지:
  - **(A) 네 이벤트 전부 fail-closed + 탈출 장치** (권장). 근거: `DecisionBearingEvents()` 를 그대로 쓰므로 목록이 하나로 유지된다. 잠김은 소리가 나고(차단 사유가 모델·사용자에게 보인다), 우회는 소리가 나지 않는다. Claude 의 Stop 루프는 호스트 상한(8회, hooks-system.md:213)이 끊는다. 비용: 지속적인 파싱 실패(호스트 형식 변화)에서 UserPromptSubmit 이 매 프롬프트를 차단해 세션이 탈출 장치를 켤 때까지 쓸 수 없다.
  - **(B) PreToolUse·PermissionRequest 만 fail-closed, Stop·UserPromptSubmit 은 fail-open + 기록**. 근거: 이 둘은 위험한 **동작**을 막는 게이트가 아니라 진행을 막는 게이트이고, 지속 실패의 잠김 비용이 크다. 비용: 결정 집합의 부분집합을 코드 어딘가에 새로 정의해야 한다 — 리드 제약(「두 번째 목록 금지」)과 긴장한다. 택한다면 부분집합은 번역 표의 Outcome 이나 별도 목록이 아니라 「이벤트가 동작 게이트인가」라는 이름 붙은 술어로 `codexadapter` 안에 둬야 하며, 그 술어의 정당화를 spec.md §B 표에 적어야 한다.
  - 탈출 장치의 형태: 환경 변수(이름은 run-phase 에서 `internal/config/envkeys.go` 에 정의) vs `.moai/config/sections/*.yaml` 키. 환경 변수는 훅 프로세스 환경에서만 읽히므로 모델이 Bash 도구로 켤 수 없다(훅 프로세스는 호스트가 띄운다). YAML 키는 모델이 Edit 로 켤 수 있는 면이다. **환경 변수를 권장한다.** 다만 `settings.json` 의 `env` 블록도 모델이 편집 가능한 면이라는 점은 남는다 — 그 면은 ConfigChange 감시의 몫이다.
- **Q2** [NEEDS CLARIFICATION: Codex 의 Stop 연속 차단 상한] — Claude 는 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`(기본 8)이 있다. Codex 에 대응하는 상한이 있는지 이 plan 에서 확인하지 못했다. 없다면 Codex 에서 파싱 실패가 지속될 때 Stop 이 무한 차단된다. t1099 의 `@MX:WARN` 이 Stop 루프 상한을 「M2d Stop 체인의 몫」으로 넘겼으므로, t1099 착지분에 그 상한이 들어왔는지 §F 재확인에서 함께 본다.
- **Q3** [NEEDS CLARIFICATION: `runAgentHook` 포함 여부] — `internal/cli/hook.go:447-463` 에 같은 모양의 fail-open 분기가 있고, `-validation` / `-pre-transformation` / `-pre-implementation` 액션과 미지 액션은 PreToolUse 로 매핑된다. 같은 결함 계열의 형제 경로를 남겨 두면 수리가 절반이 된다. **포함을 권장**하지만, 원 카드 범위 밖이므로 운영자 판정 사항이다. 포함 시 REQ·AC 를 추가하는 D-NEW-1 개정이 필요하다.
- **Q4** [NEEDS CLARIFICATION: 빈 stdin 경로] — `ReadInput` 은 빈 stdin 을 오류가 아니라 기본 입력으로 처리한다(`protocol.go:50-54`). 결정 이벤트에서 빈 페이로드로 가드가 기본 입력을 받아 실행되는데, 가드가 빈 `tool_name` 을 허용으로 처리하면 같은 우회가 된다. 이 SPEC 은 다루지 않는다. 별도 카드로 측정할지 판정이 필요하다.
- **Q5** [NEEDS CLARIFICATION: 5 MiB 경로의 호스트 실현성] — 호스트(Claude Code, Codex)가 5 MiB 를 넘는 `tool_input` 을 훅 stdin 으로 실제 넘기는지 측정하지 않았다. 이 SPEC 의 수리는 실현성과 무관하게 옳지만(파싱 실패는 다른 경로로도 생긴다), 위협 등급 산정과 sync-phase `--security --deep` 렌즈의 판단에는 이 측정이 필요하다.
- **Q6** Can Block 이지만 관측으로 둔 11개 이벤트(spec.md §B.3) 가운데 결정 집합으로 옮겨야 할 것이 있는지. 이 SPEC 은 옮기지 않는다. 옮기려면 `DecisionBearingEvents()` 를 바꾸는 별도 카드가 필요하다 — 판정만 기록한다.
- **Q7** Claude 하네스의 영속 기록면. REQ-HSF-008 은 `codexadapter.RecordDiscards` 재사용을 요구하는데, 그 기록면의 경로가 `.moai/logs/codex-adapter.jsonl` 이라 Claude 하네스 사건이 「codex-adapter」 이름의 파일에 쌓인다. **기록면 하나를 재사용하는 쪽을 권장**(새 기록면을 만들면 관측 지점이 둘로 갈린다). 이름 불일치가 받아들일 수 없다면 대안은 Claude 경로에서 stderr 만 남기는 것이다.

## §C 착수 전 점검 (Pre-flight)

1. t1099 착지 확인: `git merge-base --is-ancestor <t1099 착지 커밋> HEAD` → exit 0.
2. spec.md §F.1 의 심볼 diff 를 실행하고 결과를 progress.md §E.2 에 붙인다. 달라진 것이 있으면 구현 전에 spec.md §B·acceptance.md 를 먼저 고친다.
3. 기준선 측정(구현 전, 별도 커밋): 수리 전 트리에서 `pre-tool` 에 `{"broken` 을 넣었을 때 stdout 이 `{}` 이고 exit 0 이며 디스패치가 0회임을 재현하고 progress.md 에 기록한다 — 재현이 수리 커밋보다 앞선 커밋에 있어야 순서가 git 이력으로 증명된다(`verification-claim-integrity.md` §2.3).

## §D 제약

- `internal/hook` 패키지를 수정하지 않는다. 이음매는 `internal/cli` 에 둔다.
- 결정 이벤트 목록을 새로 나열하지 않는다(REQ-HSF-002).
- 템플릿(`internal/template/templates/`) 변경 없음 — 동작 변경은 Go 바이너리 안에서만 일어난다. 문서 반영(hooks-system.md 등)이 필요하면 sync-phase 에서 판단한다.
- 환경 변수 이름은 `internal/config/envkeys.go` 상수로 정의한다(하드코딩 금지).
- 테스트는 `t.TempDir()` 과 `CLAUDE_PROJECT_DIR` 로 기록면을 격리한다. 전체 스위트(`go test ./...`)를 로컬에서 돌리지 않는다.

## §E 자기 검증

run 완료 보고는 acceptance.md 의 AC 마다 명령과 원문 출력을 붙인다. 특히 AC-HSF-003(단일 목록)은 동작 등가 테스트와 정적 grep 둘 다를 요구한다 — 하나만으로는 「목록이 없다」를 증명하지 못한다.

## §F 마일스톤 (결정의 번복 가능성 순)

### M0 — 미해결 질문 판정 (Priority High, 사람)

Q1·Q3 는 구현 형태를 바꾸므로 Kickoff 전에 판정한다. Q2·Q5 는 측정 과제로 run 초반에 수행해 progress.md 에 기록한다. Q4·Q6·Q7 은 판정만 기록한다.

### M1 — 결정 이벤트의 fail-closed 동작 (Priority High)

- 파싱 실패 분기에서 `codexadapter.IsDecisionBearing(event)` 로 갈라, 결정 이벤트는 하네스별 fail-closed 경로로, 관측 이벤트는 기존 경로로 보낸다.
- Codex: `writeCodexFailClosed(event, <파싱 오류>)` 재사용.
- Claude: 번역 표 HarnessClaude fatal_error 행에서 렌더링. `TranslateCodex` 를 하네스 인자를 받는 형태로 일반화할지, CLI 쪽에서 `Lookup` + `Render` 를 직접 부를지는 구현 판단이다 — 어느 쪽이든 표를 거친다. 일반화한다면 t1099 의 `@MX:ANCHOR`(fan_in) 를 갱신한다.
- Stop 경로에 `@MX:WARN`(REQ-HSF-009): 호스트 상한 의존 사실과 Q2 결과를 적는다.

### M2 — 탈출 장치와 기록 (Priority High)

- Q1 판정에 따른 탈출 장치. 기본 꺼짐. 켜졌을 때도 stderr·기록을 남긴다.
- stderr 한 줄 + `RecordDiscards` 기록 한 건. 페이로드 원문은 싣지 않는다(REQ-HSF-011).

### M3 — 하네스 모드 판정 순서 이동 (Priority Medium)

- `harnessModeIsCodex(cmd)` 를 `ReadInput` 앞으로 옮긴다. 잘못된 `--harness` 값의 거부는 그대로 0 이 아닌 종료다 — 다만 이제 stdin 을 읽기 전에 거부된다. `validateCodexHarnessEvent` 는 페이로드가 필요하므로 파싱 성공 뒤에 남는다.

### M4 — 테스트 갱신과 추가 (Priority Medium)

- acceptance.md §D 의 기존 테스트 갱신.
- 결정 4 × 하네스 2 × 파손 3 형태의 표 기반 테스트, 관측 22 × 하네스 2 특성 테스트, 단일 목록 등가 테스트.

## §G 금지 패턴

- 결정 이벤트를 `[]hook.EventType{…}` 리터럴이나 `switch` 로 다시 나열하기.
- fail-closed JSON 을 `fmt.Sprintf` 나 map 리터럴로 손 조립하기(번역 표 우회).
- 거부를 exit 2 로 내기 — stdout JSON 이 무시되어 사유가 사라진다(`internal/cli/hook.go:362-366`).
- 파싱 실패 원문 페이로드를 기록·사유에 싣기.
- 관측 이벤트의 출력·종료 코드를 바꾸기.

## §H 교차 참조

- spec.md §B (분류표), §F (전제·위험)
- `fabc33812:internal/codexadapter/decision.go`, `translate.go`, `internal/cli/hook_codex_failclosed.go`
- `.claude/rules/moai/core/hooks-system.md:213` (Stop 차단 상한), `:388` (Can Block 목록)
- `6a3603274` (보존 대상 의도)
