---
id: SPEC-HOOK-STDIN-FAILCLOSED-001
title: "훅 stdin 파싱 실패 시 결정 이벤트 fail-closed — 관측 이벤트의 기존 fail-open 보존"
version: "0.1.0"
status: draft
created: 2026-09-24
updated: 2026-09-24
author: manager-spec (card t1152)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/codexadapter"
lifecycle: spec-anchored
tags: "hook, fail-closed, stdin, security, codex, dual-harness"
tier: M
related_specs:
  - SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001
  - SPEC-HOOK-FAILURE-CLASSIFY-001
  - SPEC-CODEX-WIRING-001
---

# SPEC-HOOK-STDIN-FAILCLOSED-001 — 훅 stdin 파싱 실패의 결정 이벤트 fail-closed

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-24 | manager-spec (card t1152) | 최초 plan-phase 초안. 결함은 이 트리(HEAD `60017eb83`)의 코드 판독으로 확인했다. t1099 의 결정 이벤트 집합·번역 표·fault writer 는 아직 develop 에 없어 `fabc33812` 고정 SHA 로만 인용한다(§F 전제). |

---

## §A 배경

### A.1 결함 — 파싱 실패가 모든 이벤트에서 `{}` + exit 0 이 된다

`internal/cli/hook.go:272-280` (`runHookEvent`) 은 `deps.HookProtocol.ReadInput(os.Stdin)` 이 오류를 돌려주면 stderr 에 경고 한 줄을 쓰고 `writeHookOutput(event, nil, &hook.HookOutput{})` 를 반환한다. 결과는 stdout `{}` 와 exit 0 이며, 이 분기는 **이벤트를 가리지 않는다.** PreToolUse·PermissionRequest·Stop·UserPromptSubmit 처럼 호스트의 다음 동작을 바꾸는 결정 이벤트도 똑같이 `{}` 를 낸다.

호스트는 `{}` 를 「이 훅은 의견 없음」으로 읽는다. PreToolUse 에서 의견 없음은 호스트 자신의 권한 흐름으로 넘어간다는 뜻이고, 무확인 권한 모드에서는 그것이 곧 허용이다. 따라서 **stdin 페이로드가 깨지면 그 이벤트에 등록된 모든 가드 핸들러가 한 번도 실행되지 않은 채 우회된다**(fail-open). 두 하네스(Claude Code, Codex) 모두 같은 분기를 지난다.

순서 문제도 있다. 이 `ReadInput` 오류 분기는 `harnessCodex, herr := harnessModeIsCodex(cmd)` (`internal/cli/hook.go:292`) **보다 먼저** 실행되므로, 분기가 발화하는 시점에는 호출이 Codex 모드인지조차 알 수 없다. `harnessModeIsCodex` 는 stdin 이 아니라 cobra 플래그(`--harness`)를 읽으므로(`internal/cli/hook_harness_codex.go:30-41`) 원리상 `ReadInput` 앞으로 옮길 수 있다.

같은 관찰이 t1099 에도 기록돼 있다. `fabc33812:.moai/specs/SPEC-DUAL-HARNESS-HOOK-PARITY-001/progress.md:294-298` 의 Residual risk 항목이 「stdin 이 유효한 JSON 이 아니면 `--harness codex` 의 결정 이벤트에서도 `{}` 가 exit 0 으로 나가며, Codex 가 그것을 allow 로 해석할 수 있다 — 운영자 판정이 필요하다」고 적었다. 이 SPEC 이 그 판정의 결과물이다.

### A.2 기존 설계 의도 — `6a3603274` 가 일부러 만든 fail-open

이 분기는 실수가 아니라 의도된 설계다. 커밋 `6a3603274` (2026-07-03, `fix(hook): Go hook 프로토콜 공식 스펙 정합 ...`) 의 메시지는 다음과 같다(원문 인용):

> protocol.go: stdin LimitReader 5MB + 잘린/불량 JSON 우아한 처리 (cobra usage 노이즈 + exit 1 → stderr 경고 + 기본 출력 + exit 0)

같은 커밋이 `internal/hook/protocol.go:17` 에 `maxHookInputBytes = 5 << 20` 을 두고, 그 주석에 「cap 을 넘는 페이로드는 잘리고 JSON 파싱에 실패하며, 호출자가 우아하게 처리한다(default output, exit 0)」고 적었다. `internal/cli/hook.go:274-277` 의 주석도 같은 의도를 말한다: 깨진 stdin 이 훅 파이프라인을 실패시키면 cobra 가 usage 노이즈와 exit 1 을 내고, 훅이 관찰하는 도구가 가짜 훅 실패를 드러낸다.

이 의도는 **관측 이벤트에 대해서는 여전히 옳다.** PostToolUse 가 깨진 페이로드 때문에 exit 1 을 내면 이미 실행된 도구 호출이 실패로 보이고, 운영자는 존재하지 않는 결함을 쫓는다. 관측 이벤트에서 `{}` 는 막아야 할 것을 흘려보내지 않는다 — 애초에 막을 것이 없기 때문이다.

**결정 이벤트에서는 이 의도가 보안 가드와 충돌한다.** 결정 이벤트의 `{}` 는 「이상 없음」이 아니라 「가드가 판단하지 않았는데 판단한 것처럼 통과」다. 이 SPEC 은 `6a3603274` 의 의도를 **결정 이벤트에 한해서만** 바꾼다: 파이프라인을 깨뜨리지 않는다는 원칙(exit 1·usage 노이즈 금지)은 그대로 두고, 출력만 `{}` 에서 명시적 거부로 바꾼다. 거부도 JSON 이고 exit 0 이므로, `6a3603274` 가 없애려 한 증상(usage 노이즈, exit 1, 가짜 도구 실패)은 되살아나지 않는다.

### A.3 위협 모델 — 누가 stdin 을 깨뜨릴 수 있는가

훅 stdin 은 호스트가 만든다. 그러나 호스트가 만드는 페이로드의 **내용 일부는 모델이 정한다**: PreToolUse 페이로드의 `tool_input` 은 모델이 호출한 도구의 인자 그대로다. 파손 경로는 셋이다.

| 경로 | 누가 일으키는가 | 성격 |
|---|---|---|
| 5 MiB 초과 페이로드 | 모델 — 예: 본문이 5 MiB 를 넘는 Write 호출, 긴 Bash 인자. 프롬프트 인젝션을 받은 모델도 포함 | `io.LimitReader` 가 잘라 JSON 이 깨지고 파싱 오류가 된다(`protocol.go:43-45` → `:59-67`). **모델이 제어 가능한 경로이며 가장 중요한 위협이다** |
| 호스트 형식 변화 | 호스트 버전 차이 — `normalizeHookInput` 이 새 형식을 거부 | 비악의적이지만 지속적이다. 한 번 발생하면 매 호출마다 재발한다 |
| 전송 중 절단 | 훅 래퍼·파이프 결함 | 비악의적, 간헐적 |

첫 번째 경로가 성립하면, 모델은 가드가 거부했을 도구 호출을 페이로드를 부풀리는 것만으로 가드 전체를 건너뛰고 실행할 수 있다. 이 경로의 실현 가능성 — 호스트가 5 MiB 를 넘는 `tool_input` 을 실제로 훅에 넘기는가 — 은 **이 plan 에서 측정하지 않았다**(plan.md 미해결 질문 Q5). 코드 판독으로 확인한 것은 「넘기면 파싱 오류가 되고, 파싱 오류면 `{}` 가 된다」까지다. run-phase AC-HSF-001 이 이 연쇄를 프로세스 내에서 재현한다.

fail-closed 가 중요한 이유는 비대칭에 있다. 잘못된 거부는 소리가 난다 — 모델과 사용자가 거부 사유를 보고, 원인을 추적할 수 있다. 잘못된 통과는 소리가 나지 않는다 — 가드가 실행되지 않았다는 사실이 어디에도 드러나지 않는다.

### A.4 t1099 와의 관계 — 결정 이벤트 목록은 하나만 둔다

card t1099 (SPEC-DUAL-HARNESS-HOOK-PARITY-001, 브랜치 `WT-dual-harness-parity-rebuild`) 가 결정 이벤트 집합, 정규화된 결정 어휘, 하네스별 번역 표, Codex fault writer 를 도입했다. **이 코드는 아직 develop 에 없다**(`git merge-base --is-ancestor fabc33812 HEAD` → exit 1, 이 트리에서 측정). 이 SPEC 은 그것을 고정 SHA `fabc33812` 로만 인용한다.

| 심볼 | 위치 (`fabc33812`) | 이 SPEC 에서의 역할 |
|---|---|---|
| `DecisionBearingEvents()` | `internal/codexadapter/decision.go:77-79` | 결정 이벤트 집합의 **유일한** 원천. 값: PreToolUse, PermissionRequest, Stop, UserPromptSubmit |
| `IsDecisionBearing(ev)` | `internal/codexadapter/translate.go:19-27` | 분류 판정 함수 |
| 번역 표 `buildTranslationTable` | `internal/codexadapter/decision.go:92-131` (HarnessClaude 행 `:103-108`, HarnessCodex 행 `:114-117`) | 하네스 × 이벤트 × 결정 → 출력 형태 |
| `Lookup(h, ev, d)` / `Render(ev, o, reason)` | `decision.go:139`, `decision.go:171` | 표 행 조회와 바이트 렌더링 |
| `TranslateCodex(ev, d, reason)` | `translate.go:42` | Codex 렌더링 단일 진입점 |
| `writeCodexFailClosed(event, cause)` | `internal/cli/hook_codex_failclosed.go` | Codex fault → fail-closed deny 작성기 + `RecordDiscards` 기록 |

[HARD] 이 SPEC 은 **두 번째 결정 이벤트 목록을 만들지 않는다.** Codex 측은 `codexadapter.IsDecisionBearing` 과 `writeCodexFailClosed` / `TranslateCodex(ev, DecisionFatalError, reason)` 를 그대로 재사용한다. Claude 측도 같은 집합과 같은 번역 표의 HarnessClaude 행(`Lookup(HarnessClaude, ev, DecisionFatalError)` → `Render`)에서 출력을 얻는다 — 손으로 JSON 을 짜지 않는다.

`fabc33812` 의 번역 표에서 fatal_error 열은 두 하네스 모두 네 이벤트에서 `OutcomeDeny` 다(`decision.go:103-108`, `:114-117`). 따라서 두 하네스의 이벤트 집합이 같고, 이 SPEC 에는 하네스별로 다른 목록이 필요하지 않다.

---

## §B 이벤트 분류표

### B.1 판정 기준

- **결정 이벤트 (fail-closed)**: `codexadapter.IsDecisionBearing(ev) == true` 인 이벤트. 다른 기준을 두지 않는다.
- **관측 이벤트 (기존 `{}` + exit 0 보존)**: 그 밖의 모든 이벤트.

「Claude 호스트가 차단을 받아들이는가」(`.claude/rules/moai/core/hooks-system.md:388` 의 Can Block 목록)는 분류 기준이 **아니다.** Can Block 이면서 관측으로 남는 이벤트가 있으며, 각각의 사유를 표에 적었다. 그 이벤트들을 결정 집합에 넣는 일은 `DecisionBearingEvents()` 자체를 바꾸는 일이고, 그러면 Codex 쪽 번역 표·패리티 검증도 함께 바뀌므로 이 SPEC 의 범위 밖이다(§D).

### B.2 전체 표 — `internal/hook/types.go` 의 이벤트 상수 30개

열 설명: `sub` = `moai hook` 하위 명령 (`internal/cli/hook.go:53-78`, 모두 `runHookEvent` 로 들어간다, `:90`). `Codex` = `internal/codexadapter/events.go:71-84` 의 EventTable 행(A = adapted, U = recognized/unadapted, — = 행 없음). `CB` = hooks-system.md:388 의 Claude Can Block.

| # | 이벤트 (types.go:line) | sub (hook.go:line) | Codex | CB | 분류 | 파싱 실패 시 출력 — Claude | 파싱 실패 시 출력 — Codex | 관측으로 두는 사유 |
|---|---|---|---|---|---|---|---|---|
| 1 | PreToolUse (:25) | pre-tool (:54) | A | Y | **결정** | `hookSpecificOutput.permissionDecision:"deny"` + 사유, exit 0 | 같은 형태 (`TranslateCodex` fatal_error), exit 0 | — |
| 2 | PermissionRequest (:55) | permission-request (:63) | U | Y | **결정** | `hookSpecificOutput.decision.behavior:"deny"` + `message`, exit 0 | 같은 형태, exit 0 | — |
| 3 | Stop (:34) | stop (:57) | A | Y | **결정** | `{"decision":"block","reason":…}`, exit 0 | 같은 형태, exit 0 | — |
| 4 | UserPromptSubmit (:52) | user-prompt-submit (:62) | A | Y | **결정** | `{"decision":"block","reason":…}`, exit 0 | 같은 형태, exit 0 | — |
| 5 | SessionStart (:22) | session-start (:53) | A | N | 관측 | `{}` exit 0 (보존) | `{}` exit 0 (보존) | 차단 불가 이벤트 |
| 6 | PostToolUse (:28) | post-tool (:55) | A | N (JSON block 은 사후 피드백) | 관측 | 보존 | 보존 | 도구가 이미 실행됨 — 막을 동작이 없다. `6a3603274` 가 지키려던 바로 그 경우 |
| 7 | SessionEnd (:31) | session-end (:56) | A | N | 관측 | 보존 | 보존 | 차단 불가 |
| 8 | SubagentStop (:37) | subagent-stop (:66) | A | Y | 관측 | 보존 | 보존 | 차단 = 서브에이전트 계속 작업. 위험 동작을 막지 않고, 파싱 실패 상태에서 `stop_hook_active` 를 읽을 수 없어 루프 위험만 더한다 (plan Q6) |
| 9 | PreCompact (:40) | compact (:58) | U | Y | 관측 | 보존 | 보존 | 차단 = 압축 거부. 막을 위해가 없고 거부는 컨텍스트 고갈을 앞당긴다 |
| 10 | PostToolUseFailure (:43) | post-tool-failure (:59) | — | N | 관측 | 보존 | — | 차단 불가 |
| 11 | Notification (:46) | notification (:60) | — | N | 관측 | 보존 | — | 차단 불가 |
| 12 | SubagentStart (:49) | subagent-start (:61) | A | N | 관측 | 보존 | 보존 | 차단 불가 (hooks-system.md:388) |
| 13 | TeammateIdle (:58) | teammate-idle (:64) | — | Y | 관측 | 보존 | — | 차단 = 팀원 계속 작업. 위험 동작 게이트가 아니다 (plan Q6) |
| 14 | TaskCompleted (:61) | task-completed (:65) | — | Y | 관측 | 보존 | — | 차단 = 완료 거부. 무결성 게이트이지 위험 동작 게이트가 아니다 (plan Q6) |
| 15 | WorktreeCreate (:65) | worktree-create (:67) | — | Y | 관측 | 보존 | — | 설정에 미등록(`coverage_table.go` IsActive:false). 이 이벤트는 빈 출력이 이미 생성 중단이다 |
| 16 | WorktreeRemove (:69) | worktree-remove (:68) | — | N | 관측 | 보존 | — | 차단 불가 |
| 17 | PostCompact (:73) | post-compact (:69) | U | N | 관측 | 보존 | 보존 | 차단 불가 |
| 18 | InstructionsLoaded (:77) | instructions-loaded (:70) | — | N | 관측 | 보존 | — | 차단 불가 |
| 19 | StopFailure (:81) | stop-failure (:71) | — | N | 관측 | 보존 | — | 차단 불가 |
| 20 | ConfigChange (:85) | config-change (:72) | — | Y | 관측 | 보존 | — | MoAI 핸들러가 무조건 빈 출력(hooks-system.md:105) — 현재 가드가 없다 |
| 21 | TaskCreated (:89) | task-created (:73) | — | Y | 관측 | 보존 | — | 관측 전용(RETIRE-OBS-ONLY), HOI opt-in 게이트 |
| 22 | CwdChanged (:93) | cwd-changed (:74) | — | N | 관측 | 보존 | — | 차단 불가 |
| 23 | FileChanged (:97) | file-changed (:75) | — | N | 관측 | 보존 | — | 차단 불가 |
| 24 | Elicitation (:101) | elicitation (:76) | — | Y | 관측 | 보존 | — | RETIRE-OBS-ONLY, 설정 미등록 |
| 25 | ElicitationResult (:105) | elicitation-result (:77) | — | Y | 관측 | 보존 | — | RETIRE-OBS-ONLY, 설정 미등록 |
| 26 | PermissionDenied (:110) | permission-denied (:78) | — | N | 관측 | 보존 | — | 차단 불가 |
| 27 | PostToolBatch (:114) | 없음 | — | Y | 관측 (도달 불가) | `runHookEvent` 에 도달하지 않음 | — | 하위 명령이 없어 이 경로를 지나지 않는다 |
| 28 | UserPromptExpansion (:118) | 없음 | — | Y | 관측 (도달 불가) | 도달하지 않음 | — | 같음 |
| 29 | MessageDisplay (:122) | 없음 | — | N | 관측 (도달 불가) | 도달하지 않음 | — | 같음 |
| 30 | Setup (:133) | 없음 | — | N | 관측 (도달 불가) | 도달하지 않음 | — | 인식 전용 |

Codex 전용 `Interrupt` (`internal/codexadapter/events.go:45`) 는 MoAI 디스패처 대응이 없어 이 경로에 들어오지 않는다.

### B.3 집계

- 결정 이벤트: **4** (두 하네스 동일 — `DecisionBearingEvents()` 하나에서 나온다)
- 관측 이벤트: **26** — 그중 `runHookEvent` 로 들어오는 하위 명령 **22**, 하위 명령이 없어 도달하지 않는 **4**
- Can Block 이지만 관측으로 두는 이벤트: **11** (SubagentStop, PreCompact, TeammateIdle, TaskCompleted, WorktreeCreate, ConfigChange, TaskCreated, Elicitation, ElicitationResult, PostToolBatch, UserPromptExpansion) — 각각의 사유는 표에 있다. 이들 중 무엇이든 결정 집합으로 옮기려면 `DecisionBearingEvents()` 를 바꾸는 별도 카드가 필요하다.

### B.4 결정 이벤트의 출력 형태 (fatal_error 행)

네 이벤트 모두 번역 표의 fatal_error 열이 `OutcomeDeny` 이고, `Render` 는 하네스를 인자로 받지 않으므로 두 하네스의 **바이트 형태가 같다**(`fabc33812:internal/codexadapter/decision.go:171-211`). 차이는 사유 문구뿐이다(Codex 는 `TranslateCodex` 가 `MoAI <event> hook failed, so the call was denied fail-closed: <cause>` 로 감싼다, `translate.go:57-61`).

| 이벤트 | 렌더링 결과 | exit |
|---|---|---|
| PreToolUse | `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"<사유>"}}` | 0 |
| PermissionRequest | `{"hookSpecificOutput":{"hookEventName":"PermissionRequest","decision":{"behavior":"deny","message":"<사유>"}}}` | 0 |
| Stop | `{"decision":"block","reason":"<사유>"}` | 0 |
| UserPromptSubmit | `{"decision":"block","reason":"<사유>"}` | 0 |

exit 0 인 이유: 이 트리의 `internal/cli/hook.go:362-366` 이 기록한 원칙(「JSON deny 는 설계상 exit 0 이다. exit 2 에서는 stdout JSON 이 무시되어 deny 가 사라진다」)과, `writeCodexFailClosed` 의 「returns nil so the process exits 0 and Codex reads the deny」를 따른다.

---

## §C 요구사항 (GEARS)

- **REQ-HSF-001** (Event-driven): **When** 훅 디스패처가 결정 이벤트의 stdin 을 파싱하지 못하면(형식 불량, 잘림, 5 MiB 초과로 인한 절단 포함), the hook dispatcher **shall** 그 이벤트의 fail-closed 거부를 stdout 에 쓰고 exit 0 으로 끝내며, 어떤 핸들러에도 디스패치하지 않는다.
- **REQ-HSF-002** (Ubiquitous): The hook dispatcher **shall** 어떤 이벤트가 결정 이벤트인지를 `codexadapter.IsDecisionBearing` 한 곳으로만 판정한다 — `internal/cli` 와 `internal/hook` 에 결정 이벤트를 나열한 두 번째 목록을 두지 않는다.
- **REQ-HSF-003** (Where): **Where** 호출이 `--harness codex` 모드이면, the hook dispatcher **shall** 파싱 실패의 fail-closed 출력을 t1099 의 Codex fault 경로(`writeCodexFailClosed` → `TranslateCodex(ev, DecisionFatalError, reason)`)로 만든다.
- **REQ-HSF-004** (Where): **Where** 호출이 Claude 모드(`--harness` 미지정 또는 `claude`)이면, the hook dispatcher **shall** 파싱 실패의 fail-closed 출력을 같은 번역 표의 HarnessClaude fatal_error 행(`Lookup(HarnessClaude, ev, DecisionFatalError)` + `Render`)에서 얻으며, 출력 JSON 을 별도로 손으로 조립하지 않는다.
- **REQ-HSF-005** (Ubiquitous): The hook dispatcher **shall** 하네스 모드를 stdin 을 읽기 **전에** 결정하여, 파싱 실패 분기가 발화하는 시점에 하네스 모드가 이미 알려져 있게 한다. 잘못된 `--harness` 값은 지금과 같이 0 이 아닌 종료로 거부된다.
- **REQ-HSF-006** (Event-driven): **When** 관측 이벤트의 stdin 을 파싱하지 못하면, the hook dispatcher **shall** 현재 동작(stderr 경고 한 줄, stdout `{}`, exit 0, 디스패치 없음)을 두 하네스 모두에서 그대로 유지한다 — `6a3603274` 의 의도를 관측 이벤트에서 보존한다.
- **REQ-HSF-007** (Where): **Where** 운영자가 fail-closed 탈출 장치를 켜면, the hook dispatcher **shall** 결정 이벤트의 파싱 실패에도 `{}` + exit 0 을 내되, 탈출 장치가 적용됐다는 사실을 stderr 와 영속 기록에 남긴다. 탈출 장치는 기본값이 꺼짐이다.
- **REQ-HSF-008** (Event-driven): **When** 파싱 실패로 fail-closed 가 발화하면, the hook dispatcher **shall** 이벤트 이름·하네스 모드·파싱 오류 원인·「fail-closed」임을 담은 stderr 한 줄을 쓰고, t1099 의 `codexadapter.RecordDiscards` 기록 경로를 재사용해 영속 기록을 한 건 남긴다.
- **REQ-HSF-009** (Ubiquitous): The Stop 이벤트의 fail-closed 출력 **shall** `stop_hook_active` 의 값에 의존하지 않는다 — 파싱이 실패한 페이로드에서 그 값을 읽을 수 없기 때문이다. 반복 차단의 상한은 호스트의 Stop 차단 상한이 정하며, 그 의존 사실을 코드 주석(@MX:WARN)에 남긴다.
- **REQ-HSF-010** (Ubiquitous): The fail-closed 출력의 사유 문구 **shall** 비어 있지 않으며, 파싱 실패가 원인이라는 것과 탈출 장치를 켜는 방법을 알 수 있게 한다. 빈 사유는 Codex 에서 거부 자체를 무효로 만든다(`fabc33812:decision.go:171-175`).
- **REQ-HSF-011** (Ubiquitous): The fail-closed 출력 **shall not** 파싱에 실패한 페이로드의 원문을 사유 문구나 영속 기록에 그대로 싣는다 — 원인 오류 메시지와 바이트 길이만 싣는다. 모델이 제어하는 `tool_input` 이 기록면으로 새어 나가지 않게 하기 위함이다.

---

## §D 제외 범위

아래 항목은 이 SPEC 의 범위 밖이다.

### Out of Scope — 결정 이벤트 집합의 변경

- `codexadapter.DecisionBearingEvents()` 에 이벤트를 더하거나 빼는 일. §B.3 의 Can Block 11개 이벤트 중 어느 것이든 fail-closed 로 옮기려면 별도 카드가 집합 자체를 바꿔야 한다. 그 변경은 Codex 번역 표와 패리티 검증에도 파급된다.
- 번역 표의 행·열 값(Outcome) 변경. 이 SPEC 은 t1099 표를 소비만 한다.

### Out of Scope — 다른 stdin 진입점

- `runAgentHook` (`internal/cli/hook.go:447-463`) 의 같은 모양 fail-open 분기. 에이전트 훅 액션의 다수가 PreToolUse 로 매핑되므로 같은 결함 계열이지만, 운영자 판정 전까지는 범위 밖으로 둔다(plan.md Q3 — 포함을 권장).
- `runSpecStatus` (`:526-535`, 파싱 실패 시 exit 1), `runSessionStartCompact` (`:560-568`), harness-observe 계열 `readNormalizedHookInput` (`:707-713`), `security-turn` / codex review gate / multi review gate 의 자체 stdin 처리.
- 빈 stdin 을 성공으로 처리하는 `ReadInput` 의 동작(`protocol.go:50-54`). 빈 페이로드로 핸들러가 기본 입력을 받아 실행되는 것은 파싱 실패와 다른 경로다(plan.md Q4).

### Out of Scope — 프로토콜 계층

- `internal/hook/protocol.go` 의 `ReadInput` · `maxHookInputBytes` · `normalizeHookInput` 변경. 5 MiB 상한은 그대로 둔다. 이 SPEC 은 오류를 받은 **호출자**의 처리만 바꾼다.
- `internal/hook` 패키지의 어떤 파일도 수정하지 않는다(t1099 와 같은 원칙: 이음매는 CLI 계층에 둔다).

### Out of Scope — 핸들러 실패 경로

- 디스패치 오류·타임아웃·출력 매핑 실패의 fail-closed. 그것은 t1099 M2c 가 이미 다룬다(`fabc33812:internal/cli/hook.go:339-345`). Claude 하네스의 디스패치 오류 경로(현재 exit 1)도 이 SPEC 에서 바꾸지 않는다.
- Stop 체인의 루프 상한 설계(t1099 `@MX:WARN` 이 M2d 의 몫으로 명시한 것).

---

## §E 선행 SPEC 과의 관계

| SPEC | 관계 | 근거 |
|---|---|---|
| SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001 (completed) | **보완** | 그 SPEC 은 `.claude/hooks/moai/` 셸 래퍼 층의 공유 실패 모드(바이너리 해석 사슬, `--skip-hook`)를 감사하고 `hook-independence.md` 독트린을 만들었으며, 훅 스크립트를 수정하지 않는다고 명시했다(REQ-DIVECC-012). 이 SPEC 은 그 아래층인 Go 디스패처의 stdin 파싱 실패 모드를 다룬다 — 같은 「방어 계층이 한 조건에서 함께 무너진다」 관점의 다른 층이다. 충돌 없음. 새로 생기는 실패 모드(파싱 실패 → 결정 이벤트 전체 거부)는 그 독트린의 분류 대상이므로, sync-phase 에서 `hook-independence.md` 카탈로그 반영 여부를 판단한다 |
| SPEC-HOOK-FAILURE-CLASSIFY-001 (completed) | **보완, 영역 겹침 없음** | 그 SPEC 은 PostToolUseFailure 핸들러의 분류 입력과 trace 파일명의 session_id 를 고쳤다 — 파싱이 **성공한** 페이로드의 필드 처리다. REQ-HFC-006 의 fail-open 은 `tool_response` 필드 모양에 관한 것이다. 이 SPEC 은 파싱이 **실패한** 경우만 다루며, PostToolUseFailure 는 관측 이벤트로 남는다(§B 행 10) |
| SPEC-DUAL-HARNESS-HOOK-PARITY-001 (t1099, 미병합) | **의존** | 결정 이벤트 집합·번역 표·fault writer 를 재사용한다. 그 SPEC 의 progress.md 가 남긴 Residual risk 가 이 SPEC 의 발단이다 |
| SPEC-CODEX-WIRING-001 | **소비** | `--harness codex` 모드와 `validateCodexHarnessEvent` 를 정의했다. 이 SPEC 은 하네스 모드 판정의 위치만 앞당기며 그 의미는 바꾸지 않는다 |

---

## §F 전제와 위험

### F.1 [HARD] run-phase 착수 전제

1. **t1099 가 develop 에 착지하고, 그 develop 을 이 브랜치(`WT-hook-stdin-failclosed`)가 흡수한 뒤에만** run-phase 를 시작한다. 현재 `fabc33812` 는 이 트리의 조상이 아니다.
2. t1099 는 병합 전에 `IsDecisionBearing` / `DecisionBearingEvents` / `TranslateCodex` / `writeCodexFailClosed` / 번역 표를 바꿀 수 있다. 따라서 run 착수 시 다음을 **다시 재고** 결과를 progress.md §E.2 에 기록한다:

   ```bash
   git diff fabc33812 <착지 병합 커밋> -- \
     internal/codexadapter/decision.go \
     internal/codexadapter/translate.go \
     internal/cli/hook_codex_failclosed.go \
     internal/cli/hook.go
   ```

   특히 (a) `DecisionBearingEvents()` 의 원소, (b) HarnessClaude·HarnessCodex 의 fatal_error 열 Outcome, (c) `writeCodexFailClosed` 의 서명과 기록 키, (d) `Render` 의 네 이벤트 출력 형태가 §B.4 와 같은지 확인한다. 하나라도 다르면 §B 와 acceptance.md 를 먼저 고친다(D-NEW-1 경로).

### F.2 위험

| 위험 | 설명 | 대응 |
|---|---|---|
| Stop 루프 | 파싱 실패가 지속되면 매 Stop 이 차단된다. 파싱 실패 상태에서는 `stop_hook_active` 를 읽을 수 없어 가드식 조기 반환도 불가능하다 | Claude 는 호스트 상한(8회 연속 차단 후 해제, `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`, hooks-system.md:213)이 루프를 끊는다. Codex 쪽 상한은 이 plan 에서 확인하지 못했다(plan Q2). 탈출 장치(REQ-HSF-007)가 최후의 수단이다 |
| UserPromptSubmit 잠김 | 호스트 형식 변화처럼 지속적인 원인이면 사용자의 모든 프롬프트가 차단되어 세션을 쓸 수 없다 | 차단 사유에 탈출 장치 안내를 싣는다(REQ-HSF-010). 잠김은 소리가 나고, 우회는 소리가 나지 않는다 — 이 비대칭이 fail-closed 를 택하는 근거다 |
| 호스트 형식 변화의 파급 | `6a3603274` 가 막으려던 상황, 즉 새 호스트 버전에서 페이로드 형식이 바뀌어 매 호출이 파싱 실패하는 경우, 결정 이벤트 전부가 거부된다 | 관측 이벤트는 보존되므로 세션 시작·종료·관측 경로는 영향이 없다. 탈출 장치로 즉시 복구 가능. 기록(REQ-HSF-008)이 원인 진단의 근거가 된다 |
| 탈출 장치의 남용 | 탈출 장치가 켜진 상태로 남으면 이 SPEC 이 없는 것과 같다 | 기본 꺼짐, 적용될 때마다 stderr·영속 기록. 모델이 켤 수 있는 면인지는 plan Q1 에서 판단한다 |
| 기존 테스트의 의도 역전 | `TestRunHookEvent_MalformedStdinGraceful` 이 pre-tool 로 `6a3603274` 의 의도를 단언한다 | acceptance.md §D 에 갱신 대상과 갱신 방향을 적었다 |
