---
id: SPEC-HOOK-STDIN-FAILCLOSED-001
title: "훅 stdin 파싱 실패 시 결정 이벤트 fail-closed — 관측 이벤트의 기존 fail-open 보존"
version: "0.3.4"
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
| 0.2.0 | 2026-09-24 | manager-spec (card t1152) | plan-audit 1회차 FAIL(0.66, `.moai/reports/t1152/plan-audit.md`) 반영. D1: 관측 이벤트의 보존 출력을 이벤트별로 정정(worktree-create·worktree-remove 는 빈 stdout, 나머지 20개는 `{}`) — REQ-HSF-006, §B.2 행 15·16. D3·D7: REQ-HSF-007 에 탈출 장치 출처 제약 추가, REQ-HSF-010 을 「활성화 절차 대신 운영자 문서 식별자」로 개정, §F.2 의 존재하지 않는 「ConfigChange 감시」 지목 삭제. D4: §A.3 에 중첩 깊이 경로 추가, 5 MiB 경로를 미측정 가설로 낮춤. D8: REQ-HSF-001 에 탈출 장치 조건절 합성. D9: REQ-HSF-008·011 에 기록 키 구분·바이트 수를 명시. D10: Stop 차단 상한 근거를 JSON block 경로로 보강(REQ-HSF-009, §F.2). D11: events.go 줄 번호 정정. D13: runAgentHook 재현을 §D 에 근거로 기록. 판정 대상 질문 분류는 plan.md §B. |
| 0.3.0 | 2026-09-24 | manager-spec (card t1152) | Kickoff 차단 질문 셋을 판정 결과로 닫았다(plan.md §B.1). Q2: 레인 측정(`.moai/reports/t1152/q2-codex-stop-cap.md`, codex-cli 0.156.1)에서 Codex `exec` 가 Stop 의 JSON block 을 191회 연속 받아들이고 스스로 끊지 않았다 — 상한 없음(관측 범위 한정). Q1: 운영자 판정 A1 — 결정 이벤트 4개를 두 하네스에서 fail-closed 로 하되 Codex 하네스의 Stop 만 관측 경로로 둔다. 예외는 `internal/codexadapter` 안의 이름 붙은 술어 하나로 표현한다(REQ-HSF-012·013 신설, REQ-HSF-001·003·009 개정, §B.2 행 3·§B.3·§B.4 갱신). 탈출 장치 메커니즘: 셸 환경 변수 하나, 프로젝트·로컬 설정 `env` 선언 시 무시, 설정 판독 실패 시 불인정(REQ-HSF-007 개정, REQ-HSF-016 신설). Q3: 운영자 판정 — `runAgentHook` 을 범위에 넣는다(REQ-HSF-014·015 신설, §D 에서 이동). Claude 쪽 JSON block 상한은 여전히 미측정 독트린으로 두고 잔여 위험·추적 항목으로 남긴다. Kickoff Approval 은 아직 요청 전이다(§F.1). |
| 0.3.1 | 2026-09-24 | manager-spec (card t1152) | plan-audit 2회차 FAIL(0.86, `.moai/reports/t1152/plan-audit-iter2.md`)의 차단 결함 N1~N3 과 선택 결함 N4·N7 반영. N1: 파일 부재 = 「선언 없음」 해석을 운영자가 확인했다(plan.md §B.1 Q1, 2026-09-24). §F.2 에 「추가 후 제거」 잔여 행(키 삭제·파일 삭제 두 변형, 성립 조건은 호스트의 `env` 전파·유지 방식)을 추가하고, 완화를 과대 서술하던 「완화는 출처 제약 하나에 걸려 있다」를 고쳤다. 호스트 측정은 plan.md §C Pre-flight 5(턴·시간 상한 선언, 「유지한다」면 운영자 재판정). N2: REQ-HSF-005 를 진입점별로 나눠, `runAgentHook` 의 하네스 판독과 잘못된 `--harness` 거부가 새 동작임을 명시. N3: REQ-HSF-014 의 「판정은 `IsDecisionBearing` 한 곳」 절을 acceptance.md AC-HSF-003(b1)·(c4) 로 단언. N4: REQ-HSF-016 의 부정 의무와 긍정 의무를 두 절로 분리(REQ 수 16 유지). N7: codex 모드 agent 경로가 현재 배포 래퍼에서 도달하지 않음을 §F.2 와 plan.md 추적 항목 Q9 에 기록(범위 변경 없음). |
| 0.3.2 | 2026-09-24 | manager-spec (card t1152) | plan-audit 3회차 FAIL(0.88, `.moai/reports/t1152/plan-audit-iter3.md`)의 차단 결함 N8·N9·N10 과 선택 결함 N11·N12 반영. 운영자가 4회차 감사로 연장했다(2026-09-24, t1152 레인 AskUserQuestion). N8: plan.md §C Pre-flight 5 를 다시 짰다 — 삭제 뒤 관측은 삭제 전에 설정 출처의 값이 훅 환경에서 관측된 경우에만 판정에 쓰고, 같은 세션 전파와 세션 시작 전파를 모두 재며, 어느 경로로도 값이 오지 않으면 「전달 없음」을 세 번째 결과로 따로 닫는다. §F.2 「추가 후 제거」 행과 acceptance.md §E 를 그 규칙에 맞췄다. N9: 실행 1회를 `claude -p` 프로세스 하나로 정의하고 예산을 최대 6회로 다시 세었으며, `timeout -k 10 300` 강제 종료, 숨은 `--max-turns` 플래그 거부 시 처리, 상한이 Kickoff 때 운영자 승인을 받을 제안값임을 적었다. N10: AC-HSF-013 에 관측 매핑 action 의 잘못된 `--harness` 거부 단언을 더했다(REQ-HSF-005). N11: AC-HSF-003(b1) 이 판정 호출 결과가 분기 조건식에 쓰이는지를 보게 하고 변이 (c5) 를 더했다. N12: 「stdin 보다 먼저」 증거에 `ReadInput` 호출 0 회를 더했다. REQ 수 16 유지. |
| 0.3.3 | 2026-09-24 | manager-spec (card t1152) | plan-audit 4회차 FAIL(0.89, `.moai/reports/t1152/plan-audit-iter4.md`)의 차단 결함 N13·N14 반영. N13: plan.md §C Pre-flight 5 의 판정 전제에 삭제 뒤 기록 요건을 더했다 — 선언이 사라졌음을 보이는 훅 기록이 한 줄 이상 있어야 하고, 삭제가 실행되지 않았거나 삭제 뒤 기록이 0건인 변형은 「미측정」이다. 실행은 `--permission-mode bypassPermissions` 로 띄운다. N14: 판정 2의 셋째 절을 「어느 한 경로라도 확정되지 않았다」로 고치고 판정 3에 두 경로 확정을 전제로 넣어, 네 결과가 모든 경우를 하나씩 덮게 했다. §F.2 「추가 후 제거」 행과 acceptance.md §E 를 맞췄다. 선택 결함 N15·N16·N17 은 열린 채로 둔다. REQ 수 16 유지. |
| 0.3.4 | 2026-09-24 | manager-spec (card t1152) | Implementation Kickoff Approval 을 기록했다(운영자, t1152 레인 AskUserQuestion, 2026-09-24 — 리드 전달문은 출처가 아니다). 승인 내용: N17·N18 을 측정 전에 닫는다, plan.md §C Pre-flight 5 의 측정 조건(최대 6회, 실행당 `timeout -k 10 300`·`--max-turns 8`, 격리 `/tmp` 프로젝트에서 `--permission-mode bypassPermissions`, 「유지한다」·「미측정」이면 정지·재판정)을 제안대로 승인한다, N15·N16 은 run 에서 테스트를 더해 닫는다. 구현 커밋은 t1099 가 develop 에 착지·흡수된 뒤 시작한다(변경 없음). N17: Pre-flight 5 에 변형별 설정 파일 배치를 정했다 — 기록형 훅은 삭제되지 않는 `.claude/settings.json` 에, 탈출 장치 키는 `.claude/settings.local.json` 에 두고 A2·B2 는 후자만 지운다. N18: 권한 모드를 M0 승인 목록에 넣었다. 리드 조건: 첫 실행 전에 증거 파일이 모든 실행의 명령 원문과 격리(작업 디렉터리·설정 파일·`HOME`·`CLAUDE_CONFIG_DIR`)를 보여야 하고, 보이지 못하면 실행하지 않는다. §F.1 의 3 을 갱신했다. REQ 수 16, AC 수 13 + GATE 유지. |

---

## §A 배경

### A.1 결함 — 파싱 실패가 모든 이벤트에서 기본 출력 + exit 0 이 된다

`internal/cli/hook.go:272-280` (`runHookEvent`) 은 `deps.HookProtocol.ReadInput(os.Stdin)` 이 오류를 돌려주면 stderr 에 경고 한 줄을 쓰고 `writeHookOutput(event, nil, &hook.HookOutput{})` 를 반환한다. 결과는 exit 0 이며 stdout 은 대부분의 이벤트에서 `{}` 다(WorktreeCreate·WorktreeRemove 는 빈 stdout — §B.2). 이 분기는 **이벤트를 가리지 않는다.** PreToolUse·PermissionRequest·Stop·UserPromptSubmit 처럼 호스트의 다음 동작을 바꾸는 결정 이벤트도 똑같이 `{}` 를 낸다.

호스트는 `{}` 를 「이 훅은 의견 없음」으로 읽는다. PreToolUse 에서 의견 없음은 호스트 자신의 권한 흐름으로 넘어간다는 뜻이고, 무확인 권한 모드에서는 그것이 곧 허용이다. 따라서 **stdin 페이로드가 깨지면 그 이벤트에 등록된 모든 가드 핸들러가 한 번도 실행되지 않은 채 우회된다**(fail-open). 두 하네스(Claude Code, Codex) 모두 같은 분기를 지난다.

순서 문제도 있다. 이 `ReadInput` 오류 분기는 `harnessCodex, herr := harnessModeIsCodex(cmd)` (`internal/cli/hook.go:292`) **보다 먼저** 실행되므로, 분기가 발화하는 시점에는 호출이 Codex 모드인지조차 알 수 없다. `harnessModeIsCodex` 는 stdin 이 아니라 cobra 플래그(`--harness`)를 읽으므로(`internal/cli/hook_harness_codex.go:30-41`) 원리상 `ReadInput` 앞으로 옮길 수 있다.

같은 관찰이 t1099 에도 기록돼 있다. `fabc33812:.moai/specs/SPEC-DUAL-HARNESS-HOOK-PARITY-001/progress.md:294-298` 의 Residual risk 항목이 「stdin 이 유효한 JSON 이 아니면 `--harness codex` 의 결정 이벤트에서도 `{}` 가 exit 0 으로 나가며, Codex 가 그것을 allow 로 해석할 수 있다 — 운영자 판정이 필요하다」고 적었다. 이 SPEC 이 그 판정의 결과물이다.

### A.2 기존 설계 의도 — `6a3603274` 가 일부러 만든 fail-open

이 분기는 실수가 아니라 의도된 설계다. 커밋 `6a3603274` (2026-07-03, `fix(hook): Go hook 프로토콜 공식 스펙 정합 ...`) 의 메시지는 다음과 같다(원문 인용):

> protocol.go: stdin LimitReader 5MB + 잘린/불량 JSON 우아한 처리 (cobra usage 노이즈 + exit 1 → stderr 경고 + 기본 출력 + exit 0)

같은 커밋이 `internal/hook/protocol.go:17` 에 `maxHookInputBytes = 5 << 20` 을 두고, 그 주석에 「cap 을 넘는 페이로드는 잘리고 JSON 파싱에 실패하며, 호출자가 우아하게 처리한다(default output, exit 0)」고 적었다. `internal/cli/hook.go:274-277` 의 주석도 같은 의도를 말한다: 깨진 stdin 이 훅 파이프라인을 실패시키면 cobra 가 usage 노이즈와 exit 1 을 내고, 훅이 관찰하는 도구가 가짜 훅 실패를 드러낸다.

이 의도는 **관측 이벤트에 대해서는 여전히 옳다.** PostToolUse 가 깨진 페이로드 때문에 exit 1 을 내면 이미 실행된 도구 호출이 실패로 보이고, 운영자는 존재하지 않는 결함을 쫓는다. 관측 이벤트의 기본 출력은 막아야 할 것을 흘려보내지 않는다 — 애초에 막을 것이 없기 때문이다.

**결정 이벤트에서는 이 의도가 보안 가드와 충돌한다.** 결정 이벤트의 `{}` 는 「이상 없음」이 아니라 「가드가 판단하지 않았는데 판단한 것처럼 통과」다. 이 SPEC 은 `6a3603274` 의 의도를 **결정 이벤트에 한해서만** 바꾼다: 파이프라인을 깨뜨리지 않는다는 원칙(exit 1·usage 노이즈 금지)은 그대로 두고, 출력만 `{}` 에서 명시적 거부로 바꾼다. 거부도 JSON 이고 exit 0 이므로, `6a3603274` 가 없애려 한 증상(usage 노이즈, exit 1, 가짜 도구 실패)은 되살아나지 않는다.

### A.3 위협 모델 — 누가 stdin 을 깨뜨릴 수 있는가

훅 stdin 은 호스트가 만든다. 그러나 호스트가 만드는 페이로드의 **내용 일부는 모델이 정한다**: PreToolUse 페이로드의 `tool_input` 은 모델이 호출한 도구의 인자 그대로다. 파손 경로는 넷이다.

| 경로 | 누가 일으키는가 | 비용·근거 | 상태 |
|---|---|---|---|
| 중첩 깊이 초과 | 모델 — `tool_input` 안에 배열·객체를 깊게 중첩한 도구 호출. 프롬프트 인젝션을 받은 모델도 포함 | 약 20 KB. `encoding/json` 은 중첩 깊이 10000 을 넘으면 파싱을 거부한다. `ReadInput` 은 `normalizeHookInput(data)` (`internal/hook/protocol.go:58`) 를 부르고, 그 첫 단계가 `map[string]json.RawMessage` 로의 `json.Unmarshal` 이다(`internal/hook/normalize.go:85-88`). 측정(2026-09-24, 저장소 밖 scratchpad 의 독립 Go 프로그램, go1.26.8, 같은 대상 타입): 깊이 9000 → `bytes=18052 err=<nil>`, 깊이 10000 → `bytes=20052 err=invalid character '[' exceeded max depth`, 깊이 10001 → `bytes=20054` 같은 오류. plan-audit 의 독립 측정과 일치한다 | 파싱 실패까지는 **측정됨**. 호스트가 그런 `tool_input` 을 그대로 훅에 넘기는지(예: 자유 형식 객체 인자를 받는 MCP 도구)는 **미측정** |
| 5 MiB 초과 페이로드 | 모델 — 예: 본문이 5 MiB 를 넘는 Write 호출, 긴 Bash 인자 | `io.LimitReader` 가 잘라(`protocol.go:44`) JSON 이 깨지고 파싱 오류가 된다(`protocol.go:58-60`). 코드 연쇄는 판독으로 확인 | **미측정 가설**. 모델 한 턴의 출력 상한을 감안하면 단일 도구 호출로 5 MiB 에 이르기 어려울 수 있다. 실현성은 plan.md 추적 항목 Q5 |
| 호스트 형식 변화 | 호스트 버전 차이 — `normalizeHookInput` 이 새 형식을 거부 | 비악의적이지만 지속적이다. 한 번 발생하면 매 호출마다 재발한다 | 과거 사례에 근거한 가정 |
| 전송 중 절단 | 훅 래퍼·파이프 결함 | 비악의적, 간헐적 | 가정 |

모델이 제어하는 경로 가운데 **가장 싼 것은 중첩 깊이 초과**다. 이 경로가 호스트에서 성립하면, 모델은 가드가 거부했을 도구 호출을 인자 하나를 깊게 중첩하는 것만으로 가드 전체를 건너뛰고 실행할 수 있다. 어느 경로가 실제로 가장 중요한지는 호스트 측 측정 전까지 단정하지 않는다. 코드 판독과 측정으로 확인한 것은 「파싱에 실패하는 입력이 있고, 파싱 실패면 결정 이벤트도 `{}` 가 된다」까지다. run-phase AC-HSF-001·002 가 이 연쇄를 네 파손 형태 모두로 프로세스 내에서 재현한다.

fail-closed 가 중요한 이유는 비대칭에 있다. 잘못된 거부는 소리가 난다 — 모델과 사용자가 거부 사유를 보고, 원인을 추적할 수 있다. 잘못된 통과는 소리가 나지 않는다 — 가드가 실행되지 않았다는 사실이 어디에도 드러나지 않는다.

### A.4 t1099 와의 관계 — 결정 이벤트 목록은 하나만 둔다

card t1099 (SPEC-DUAL-HARNESS-HOOK-PARITY-001, 브랜치 `WT-dual-harness-parity-rebuild`) 가 결정 이벤트 집합, 정규화된 결정 어휘, 하네스별 번역 표, Codex fault writer 를 도입했다. **이 코드는 아직 develop 에 없다**(`git merge-base --is-ancestor fabc33812 HEAD` → exit 1, 이 트리에서 측정). 이 SPEC 은 그것을 고정 SHA `fabc33812` 로만 인용한다.

관측(2026-09-24, plan-audit 측정): t1099 브랜치는 `fabc33812` 이후 `fe4fd9d4d` 로 전진했으나, 이 SPEC 이 인용하는 파일(`decision.go`, `translate.go`, `hook_codex_failclosed.go`, `hook.go`, `diagnostics.go`, `output.go`)의 diff 는 두 커밋 사이에 0 이었다(같은 명령의 전체 diff 는 비어 있지 않아 양성 대조가 섰다). 고정점은 `fabc33812` 로 유지하며, 착지 시점 재확인은 §F.1 이 맡는다.

| 심볼 | 위치 (`fabc33812`) | 이 SPEC 에서의 역할 |
|---|---|---|
| `DecisionBearingEvents()` | `internal/codexadapter/decision.go:77-79` | 결정 이벤트 집합의 **유일한** 원천. 값: PreToolUse, PermissionRequest, Stop, UserPromptSubmit |
| `IsDecisionBearing(ev)` | `internal/codexadapter/translate.go:19-27` | 분류 판정 함수 |
| 번역 표 `buildTranslationTable` | `internal/codexadapter/decision.go:92-131` (HarnessClaude 행 `:103-108`, HarnessCodex 행 `:114-117`) | 하네스 × 이벤트 × 결정 → 출력 형태 |
| `Lookup(h, ev, d)` / `Render(ev, o, reason)` | `decision.go:139`, `decision.go:171` | 표 행 조회와 바이트 렌더링 |
| `TranslateCodex(ev, d, reason)` | `translate.go:42` | Codex 렌더링 단일 진입점 |
| `writeCodexFailClosed(event, cause)` | `internal/cli/hook_codex_failclosed.go:34-58` | Codex fault → fail-closed deny 작성기 + `RecordDiscards` 기록. 기록 키 `hook-fault`(`:20`), 기록 길이 `len(causeText)`, Reason 고정 문구(`:44-49`) |
| `RecordDiscards` / `Discard` | `internal/codexadapter/diagnostics.go:24`, `output.go:77-82` | 영속 기록면. `Discard` 는 내용이 아니라 길이만 싣는 설계다 |

[HARD] 이 SPEC 은 **두 번째 결정 이벤트 목록을 만들지 않는다.** Codex 측은 `codexadapter.IsDecisionBearing` 과 `writeCodexFailClosed` / `TranslateCodex(ev, DecisionFatalError, reason)` 를 재사용한다(재사용 범위는 plan.md M1·M2). Claude 측도 같은 집합과 같은 번역 표의 HarnessClaude 행(`Lookup(HarnessClaude, ev, DecisionFatalError)` → `Render`)에서 출력을 얻는다 — 손으로 JSON 을 짜지 않는다.

`fabc33812` 의 번역 표에서 fatal_error 열은 두 하네스 모두 네 이벤트에서 `OutcomeDeny` 다(`decision.go:103-108`, `:114-117`). 따라서 두 하네스의 이벤트 집합이 같고, 이 SPEC 에는 하네스별로 다른 목록이 필요하지 않다. 하네스별 차이는 단 하나 — Codex 하네스의 Stop — 이며, 그것은 이벤트 목록이 아니라 「Codex 호스트에는 Stop 연속 차단 상한이 없다」는 이름 붙은 술어 하나로 표현한다(REQ-HSF-013, §B.3).

---

## §B 이벤트 분류표

### B.1 판정 기준

- **결정 이벤트 (fail-closed)**: `codexadapter.IsDecisionBearing(ev) == true` 인 이벤트. 다른 기준을 두지 않는다.
- **관측 이벤트 (현재 출력 보존)**: 그 밖의 모든 이벤트.

「Claude 호스트가 차단을 받아들이는가」(`.claude/rules/moai/core/hooks-system.md:388` 의 Can Block 목록)는 분류 기준이 **아니다.** Can Block 이면서 관측으로 남는 이벤트가 있으며, 각각의 사유를 표에 적었다. 그 이벤트들을 결정 집합에 넣는 일은 `DecisionBearingEvents()` 자체를 바꾸는 일이고, 그러면 Codex 쪽 번역 표·패리티 검증도 함께 바뀌므로 이 SPEC 의 범위 밖이다(§D).

### B.2 전체 표 — `internal/hook/types.go` 의 이벤트 상수 30개

열 설명: `sub` = `moai hook` 하위 명령 (`internal/cli/hook.go:53-78`, 모두 `runHookEvent` 로 들어간다, `:90`). `Codex` = `internal/codexadapter/events.go:69-84` 의 EventTable 행(A = adapted, U = recognized/unadapted, — = 행 없음). `CB` = hooks-system.md:388 의 Claude Can Block.

「보존」은 **이벤트별 현재 출력**을 그대로 둔다는 뜻이다: stdout `{}` + exit 0 이 기본이고, WorktreeCreate·WorktreeRemove 두 이벤트만 **빈 stdout** + exit 0 이다. `writeHookOutput` 이 그 두 이벤트에서 `input == nil` 이면 아무것도 쓰지 않고 반환하기 때문이다(`internal/cli/hook.go:389-392`). 파싱 실패 분기는 `input` 을 nil 로 넘긴다(`:279`).

| # | 이벤트 (types.go:line) | sub (hook.go:line) | Codex | CB | 분류 | 파싱 실패 시 출력 — Claude | 파싱 실패 시 출력 — Codex | 관측으로 두는 사유 |
|---|---|---|---|---|---|---|---|---|
| 1 | PreToolUse (:25) | pre-tool (:54) | A | Y | **결정** | `hookSpecificOutput.permissionDecision:"deny"` + 사유, exit 0 | 같은 형태 (`TranslateCodex` fatal_error), exit 0 | — |
| 2 | PermissionRequest (:55) | permission-request (:63) | U | Y | **결정** | `hookSpecificOutput.decision.behavior:"deny"` + `message`, exit 0 | 같은 형태, exit 0 | — |
| 3 | Stop (:34) | stop (:57) | A | Y | **결정** | `{"decision":"block","reason":…}`, exit 0 | **면제** — `{}` exit 0 + stderr 한 줄 + 면제 기록 (REQ-HSF-012). 근거: Codex 에 Stop 연속 차단 상한이 없음(plan.md Q2 측정) | — (Claude 는 결정, Codex 만 면제) |
| 4 | UserPromptSubmit (:52) | user-prompt-submit (:62) | A | Y | **결정** | `{"decision":"block","reason":…}`, exit 0 | 같은 형태, exit 0 | — |
| 5 | SessionStart (:22) | session-start (:53) | A | N | 관측 | `{}` exit 0 (보존) | `{}` exit 0 (보존) | 차단 불가 이벤트 |
| 6 | PostToolUse (:28) | post-tool (:55) | A | N (JSON block 은 사후 피드백) | 관측 | `{}` 보존 | `{}` 보존 | 도구가 이미 실행됨 — 막을 동작이 없다. `6a3603274` 가 지키려던 바로 그 경우 |
| 7 | SessionEnd (:31) | session-end (:56) | A | N | 관측 | `{}` 보존 | `{}` 보존 | 차단 불가 |
| 8 | SubagentStop (:37) | subagent-stop (:66) | A | Y | 관측 | `{}` 보존 | `{}` 보존 | 차단 = 서브에이전트 계속 작업. 위험 동작을 막지 않고, 파싱 실패 상태에서 `stop_hook_active` 를 읽을 수 없어 루프 위험만 더한다 (plan Q6) |
| 9 | PreCompact (:40) | compact (:58) | U | Y | 관측 | `{}` 보존 | `{}` 보존 | 차단 = 압축 거부. 막을 위해가 없고 거부는 컨텍스트 고갈을 앞당긴다 |
| 10 | PostToolUseFailure (:43) | post-tool-failure (:59) | — | N | 관측 | `{}` 보존 | — | 차단 불가 |
| 11 | Notification (:46) | notification (:60) | — | N | 관측 | `{}` 보존 | — | 차단 불가 |
| 12 | SubagentStart (:49) | subagent-start (:61) | A | N | 관측 | `{}` 보존 | `{}` 보존 | 차단 불가 (hooks-system.md:388) |
| 13 | TeammateIdle (:58) | teammate-idle (:64) | — | Y | 관측 | `{}` 보존 | — | 차단 = 팀원 계속 작업. 위험 동작 게이트가 아니다 (plan Q6) |
| 14 | TaskCompleted (:61) | task-completed (:65) | — | Y | 관측 | `{}` 보존 | — | 차단 = 완료 거부. 무결성 게이트이지 위험 동작 게이트가 아니다 (plan Q6) |
| 15 | WorktreeCreate (:65) | worktree-create (:67) | — | Y | 관측 | **빈 stdout** exit 0 (보존, hook.go:389-392) | — | 설정에 미등록(`coverage_table.go` IsActive:false). 이 이벤트는 빈 출력이 이미 생성 중단이다 |
| 16 | WorktreeRemove (:69) | worktree-remove (:68) | — | N | 관측 | **빈 stdout** exit 0 (보존, hook.go:389-392) | — | 차단 불가 |
| 17 | PostCompact (:73) | post-compact (:69) | U | N | 관측 | `{}` 보존 | `{}` 보존 | 차단 불가 |
| 18 | InstructionsLoaded (:77) | instructions-loaded (:70) | — | N | 관측 | `{}` 보존 | — | 차단 불가 |
| 19 | StopFailure (:81) | stop-failure (:71) | — | N | 관측 | `{}` 보존 | — | 차단 불가 |
| 20 | ConfigChange (:85) | config-change (:72) | — | Y | 관측 | `{}` 보존 | — | MoAI 핸들러가 무조건 빈 출력(hooks-system.md:105) — 현재 가드가 없다 |
| 21 | TaskCreated (:89) | task-created (:73) | — | Y | 관측 | `{}` 보존 | — | 관측 전용(RETIRE-OBS-ONLY), HOI opt-in 게이트 |
| 22 | CwdChanged (:93) | cwd-changed (:74) | — | N | 관측 | `{}` 보존 | — | 차단 불가 |
| 23 | FileChanged (:97) | file-changed (:75) | — | N | 관측 | `{}` 보존 | — | 차단 불가 |
| 24 | Elicitation (:101) | elicitation (:76) | — | Y | 관측 | `{}` 보존 | — | RETIRE-OBS-ONLY, 설정 미등록 |
| 25 | ElicitationResult (:105) | elicitation-result (:77) | — | Y | 관측 | `{}` 보존 | — | RETIRE-OBS-ONLY, 설정 미등록 |
| 26 | PermissionDenied (:110) | permission-denied (:78) | — | N | 관측 | `{}` 보존 | — | 차단 불가 |
| 27 | PostToolBatch (:114) | 없음 | — | Y | 관측 (도달 불가) | `runHookEvent` 에 도달하지 않음 | — | 하위 명령이 없어 이 경로를 지나지 않는다 |
| 28 | UserPromptExpansion (:118) | 없음 | — | Y | 관측 (도달 불가) | 도달하지 않음 | — | 같음 |
| 29 | MessageDisplay (:122) | 없음 | — | N | 관측 (도달 불가) | 도달하지 않음 | — | 같음 |
| 30 | Setup (:133) | 없음 | — | N | 관측 (도달 불가) | 도달하지 않음 | — | 인식 전용 |

Codex 전용 `Interrupt` (`internal/codexadapter/events.go:45`) 는 MoAI 디스패처 대응이 없어 이 경로에 들어오지 않는다.

「파싱 실패 시 출력 — Codex」 열의 「—」는 Codex EventTable 에 행이 없다는 뜻이다. `--harness codex` 로 그 하위 명령을 부르면 파싱 실패 분기가 먼저 발화하므로(§A.1), 두 하네스 모드에서 관측 이벤트의 출력은 Claude 열과 같다. AC-HSF-005 는 22개 하위 명령 전부를 두 모드로 잰다.

### B.3 집계

- 결정 이벤트: **4** (두 하네스 동일 — `DecisionBearingEvents()` 하나에서 나온다)
- 파싱 실패 시 실제로 fail-closed 가 발화하는 (하네스, 이벤트) 쌍: **7** — Claude 4, Codex 3. 빠지는 한 쌍은 (Codex, Stop) 이며, 이 면제는 결정 이벤트 목록을 줄이는 것이 아니라 REQ-HSF-013 의 술어 하나가 그 쌍에만 참을 돌려주는 것으로 생긴다. 두 번째 이벤트 목록은 생기지 않는다
- 관측 이벤트: **26** — 그중 `runHookEvent` 로 들어오는 하위 명령 **22**(출력 `{}` 20, 빈 stdout 2), 하위 명령이 없어 도달하지 않는 **4**
- Can Block 이지만 관측으로 두는 이벤트: **11** (SubagentStop, PreCompact, TeammateIdle, TaskCompleted, WorktreeCreate, ConfigChange, TaskCreated, Elicitation, ElicitationResult, PostToolBatch, UserPromptExpansion) — 각각의 사유는 표에 있다. 이들 중 무엇이든 결정 집합으로 옮기려면 `DecisionBearingEvents()` 를 바꾸는 별도 카드가 필요하다.

### B.4 결정 이벤트의 출력 형태 (fatal_error 행)

아래 렌더링은 fail-closed 가 발화하는 7쌍에 적용된다. (Codex, Stop) 은 면제 쌍이라 렌더링하지 않고 `{}` 를 낸다(REQ-HSF-012). 네 이벤트 모두 번역 표의 fatal_error 열이 `OutcomeDeny` 이고, `Render` 는 하네스를 인자로 받지 않으므로 두 하네스의 **바이트 형태가 같다**(`fabc33812:internal/codexadapter/decision.go:171-211`). 차이는 사유 문구뿐이다(Codex 는 `TranslateCodex` 가 `MoAI <event> hook failed, so the call was denied fail-closed: <cause>` 로 감싼다, `translate.go:57-61`).

| 이벤트 | 렌더링 결과 | 거부 필드 | exit |
|---|---|---|---|
| PreToolUse | `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"<사유>"}}` | `hookSpecificOutput.permissionDecision == "deny"` | 0 |
| PermissionRequest | `{"hookSpecificOutput":{"hookEventName":"PermissionRequest","decision":{"behavior":"deny","message":"<사유>"}}}` | `hookSpecificOutput.decision.behavior == "deny"` | 0 |
| Stop | `{"decision":"block","reason":"<사유>"}` | `decision == "block"` | 0 |
| UserPromptSubmit | `{"decision":"block","reason":"<사유>"}` | `decision == "block"` | 0 |

exit 0 인 이유: 이 트리의 `internal/cli/hook.go:362-365` 이 기록한 원칙(「JSON deny 는 설계상 exit 0 이다. exit 2 에서는 stdout JSON 이 무시되어 deny 가 사라진다」)과, `writeCodexFailClosed` 의 「returns nil so the process exits 0 and Codex reads the deny」(`fabc33812:internal/cli/hook_codex_failclosed.go:30-33`)를 따른다.

「거부 필드」 열은 AC-HSF-003 의 관측 술어다. `{}` 와 빈 stdout 은 어느 거부 필드도 갖지 않는다.

---

## §C 요구사항 (GEARS)

- **REQ-HSF-001** (Where + Event-driven): **Where** 탈출 장치(REQ-HSF-007)가 꺼져 있으면(기본값), **When** 훅 디스패처가 결정 이벤트의 stdin 을 파싱하지 못하면(`ReadInput` 이 오류를 돌려주는 모든 경우 — 형식 불량, 잘림, 5 MiB 상한 절단, 중첩 깊이 초과 포함), the hook dispatcher **shall** 그 이벤트의 fail-closed 거부를 stdout 에 쓰고 exit 0 으로 끝내며, 어떤 핸들러에도 디스패치하지 않는다. 단 REQ-HSF-013 의 술어가 참인 (하네스, 이벤트) 쌍은 REQ-HSF-012 를 따른다.
- **REQ-HSF-002** (Ubiquitous): The hook dispatcher **shall** 어떤 이벤트가 결정 이벤트인지를 `codexadapter.IsDecisionBearing` 한 곳으로만 판정한다 — `internal/cli` 와 `internal/hook` 에 결정 이벤트를 나열한 두 번째 목록(슬라이스·배열·맵 리터럴, 여러 결정 이벤트를 나열한 `switch` 나 `||` 비교 사슬)을 두지 않는다.
- **REQ-HSF-003** (Where): **Where** 호출이 `--harness codex` 모드이면, the hook dispatcher **shall** 파싱 실패의 fail-closed 출력을 t1099 의 Codex fault 경로(`writeCodexFailClosed` → `TranslateCodex(ev, DecisionFatalError, reason)`)로 만든다 — REQ-HSF-012 의 면제 쌍은 제외한다. 재사용 범위(서명 확장 허용 여부)는 plan.md M1 이 정한다.
- **REQ-HSF-004** (Where): **Where** 호출이 Claude 모드(`--harness` 미지정 또는 `claude`)이면, the hook dispatcher **shall** 파싱 실패의 fail-closed 출력을 같은 번역 표의 HarnessClaude fatal_error 행(`Lookup(HarnessClaude, ev, DecisionFatalError)` + `Render`)에서 얻으며, 출력 JSON 을 별도로 손으로 조립하지 않는다.
- **REQ-HSF-005** (Ubiquitous): The hook dispatcher **shall** 하네스 모드를 stdin 을 읽기 **전에** 결정하여, 파싱 실패 분기가 발화하는 시점에 하네스 모드가 이미 알려져 있게 하고, 잘못된 `--harness` 값은 stdin 을 읽기 전에 0 이 아닌 종료로 거부한다. 두 진입점의 현재 동작은 다르다. `runHookEvent` 는 지금도 잘못된 값을 0 이 아닌 종료로 거부하므로(`internal/cli/hook.go:292`), 이 요구는 그 거부를 stdin 판독 앞으로 옮길 뿐이다. `runAgentHook` 은 지금 `--harness` 를 읽지 않아 잘못된 값을 무시하고 성공한다(`harnessModeIsCodex` 의 비테스트 호출처는 `hook.go:292` 하나뿐). 따라서 `runAgentHook` 의 하네스 판독과 잘못된 값의 거부는 이 요구가 **새로 도입하는 동작**이며, 유효한 stdin 과 잘못된 `--harness` 로 부른 agent 호출은 성공에서 실패로 바뀐다 — 보존 동작이 아니라 의도된 계약 변경이다. 이 거부는 action 이 결정 이벤트로 매핑되든 관측 이벤트로 매핑되든 모든 agent 호출에 적용된다(REQ-HSF-014·015, AC-HSF-012·013).
- **REQ-HSF-006** (Event-driven): **When** 관측 이벤트의 stdin 을 파싱하지 못하면, the hook dispatcher **shall** 이벤트별 현재 동작을 두 하네스 모드 모두에서 그대로 유지한다 — stderr 경고 한 줄, exit 0, 디스패치 없음, 그리고 stdout 은 worktree-create·worktree-remove 에서 빈 문자열, 나머지 20개 하위 명령에서 `{}`. `6a3603274` 의 의도를 관측 이벤트에서 보존한다.
- **REQ-HSF-007** (Where + Unwanted): **Where** 호스트를 띄운 셸 환경에 탈출 장치 환경 변수(이름은 run-phase 에서 `internal/config/envkeys.go` 상수로 정한다)가 켜짐 값으로 설정돼 있고, 같은 키가 `$CLAUDE_PROJECT_DIR/.claude/settings.json` 과 `$CLAUDE_PROJECT_DIR/.claude/settings.local.json` 의 `env` 블록 어디에도 선언돼 있지 않으면, the hook dispatcher **shall** fail-closed 가 발화했을 결정 이벤트의 파싱 실패에 `{}` + exit 0 을 내되, 탈출 장치가 적용됐다는 사실을 stderr 한 줄과 영속 기록 한 건(파싱 실패 fail-closed 키·면제 키·`hook-fault` 와 구분되는 전용 키)에 남긴다. 탈출 장치는 기본값이 꺼짐이다. The hook dispatcher **shall not** 탈출 장치의 활성화를 환경 변수 외의 출처에서 읽거나, 두 설정 파일 가운데 하나라도 `env` 블록에 같은 키를 선언한 상태의 환경 변수를 활성화로 인정한다 — 그 선언은 호스트가 `env` 블록을 훅 환경으로 전파해 모델이 쓴 값이 환경 변수로 보일 수 있다는 신호이기 때문이다. `.moai/config/` 아래 파일은 활성화 출처가 아니다. 잔여 면(사용자 범위 `~/.claude/settings.json` 의 `env`)은 §F.2 에 적었다.
- **REQ-HSF-008** (Event-driven): **When** 파싱 실패로 fail-closed 가 발화하면, the hook dispatcher **shall** (a) 이벤트 이름·하네스 모드·파싱 오류 원인·「fail-closed」임을 담은 stderr 한 줄을 쓰고, (b) t1099 의 `codexadapter.RecordDiscards` 기록 경로를 재사용해 영속 기록을 한 건 남기며, 그 기록은 핸들러 fault 기록(키 `hook-fault`)과 구분되는 파싱 실패 전용 키와, 디스패처가 관측한 stdin 바이트 수를 싣는다. `runAgentHook` 경로의 fail-closed 도 같다(REQ-HSF-014).
- **REQ-HSF-009** (Ubiquitous): The Stop 이벤트의 fail-closed 출력 **shall** `stop_hook_active` 의 값에 의존하지 않는다 — 파싱이 실패한 페이로드에서 그 값을 읽을 수 없기 때문이다. Claude 하네스에서 반복 차단의 상한은 호스트의 Stop 차단 상한이 정하며, 그 의존 사실과 근거 등급을 코드 주석(@MX:WARN)에 남긴다. Claude 쪽 상한이 JSON `decision:"block"` + exit 0 차단에도 걸린다는 근거는 저장소 독트린(`goal-directive.md:13`, 그 대상인 stop-goal 평가기의 출력 형태 `internal/cli/hook_stop_goal.go:120-121`)뿐이며 **측정되지 않았다** — 잔여 위험으로 남기고 추적한다(§F.2, plan.md 추적 항목 Q8). Codex 하네스에는 상한이 없다는 측정 결과(plan.md Q2)에 따라 Codex 의 Stop 은 fail-closed 하지 않는다(REQ-HSF-012).
- **REQ-HSF-010** (Ubiquitous): The fail-closed 출력의 사유 문구 **shall** 비어 있지 않으며, (a) `fail-closed` 라는 표시, (b) 원인이 stdin 파싱 실패라는 고정 문구, (c) 운영자 문서 식별자(고정 문자열)를 싣는다. 사유 문구는 모델이 읽는 면이므로(PreToolUse 의 `permissionDecisionReason` 등), 탈출 장치의 식별자(환경 변수 이름)와 활성화 절차는 싣지 않는다 — 활성화 절차는 운영자 문서에만 둔다. 빈 사유는 Codex 에서 거부 자체를 무효로 만든다(`fabc33812:decision.go:171-175`).
- **REQ-HSF-011** (Unwanted): The fail-closed 출력, 면제·탈출 장치 적용의 stderr 줄과 영속 기록 **shall not** 파싱에 실패한 페이로드의 원문이나 그 일부를 싣는다. 입력에서 유래한 정보로 허용되는 것은 파싱 오류 메시지(`encoding/json` 오류가 인용하는 한 글자·숫자 리터럴 수준)와 stdin 바이트 수뿐이다. 모델이 제어하는 `tool_input` 이 사유면·기록면으로 새어 나가지 않게 하기 위함이다.
- **REQ-HSF-012** (Where + Event-driven): **Where** 호출이 `--harness codex` 모드이고 이벤트가 Stop 이면, **When** stdin 을 파싱하지 못하면, the hook dispatcher **shall** fail-closed 거부 대신 stdout `{}` 와 exit 0 을 내고, 디스패치하지 않으며, 면제가 적용됐다는 사실(이벤트·하네스 모드·파싱 오류 원인·면제 사유가 「호스트에 Stop 연속 차단 상한 없음」이라는 것)을 담은 stderr 한 줄과, 면제 전용 키(파싱 실패 fail-closed 키·`hook-fault` 와 구분)로 영속 기록 한 건을 남긴다. 근거는 plan.md Q2 측정(`.moai/reports/t1152/q2-codex-stop-cap.md`)이다: codex-cli 0.156.1 의 `codex exec` 는 Stop 의 JSON block 을 191회 연속 받아들이며 스스로 끊지 않았으므로, 여기서 fail-closed 하면 파싱 실패가 지속되는 동안 턴이 끝나지 않는다. Claude 하네스의 Stop 은 이 요구의 대상이 아니다 — REQ-HSF-001 대로 차단한다.
- **REQ-HSF-013** (Ubiquitous): The (Codex, Stop) 면제 **shall** `internal/codexadapter` 안에 있는 이름 붙은 술어 **하나**로 표현되며, 그 술어는 「Codex 호스트에는 Stop 연속 차단 상한이 없다」는 사실을 나타낸다. 디스패처는 fail-closed 여부를 `IsDecisionBearing` 과 이 술어의 조합으로만 판정하고, 이벤트 목록을 따로 두지 않는다. 술어의 주석은 Q2 측정 근거(codex 버전·관측 범위)를 인용하고, 향후 Codex 가 상한을 도입하면 이 술어 하나가 재검토 지점임을 적는다.
- **REQ-HSF-014** (Event-driven): **When** `moai hook agent <action>` 의 action 이 결정 이벤트로 매핑되고(현재 코드 기준: 접미사 `-validation`·`-pre-transformation`·`-pre-implementation` 과 그 밖의 미지 action 은 PreToolUse — `internal/cli/hook.go:469-481`) 그 stdin 을 파싱하지 못하면, the agent hook dispatcher **shall** `runHookEvent` 의 같은 이벤트·같은 하네스 모드와 같은 fail-closed 출력, exit 0, 디스패치 없음, REQ-HSF-008 의 stderr·영속 기록을 낸다. action → 이벤트 매핑과 하네스 모드는 stdin 을 읽기 전에 결정한다. 결정 이벤트 판정은 여기서도 `IsDecisionBearing` 한 곳이다(REQ-HSF-002).
- **REQ-HSF-015** (Event-driven): **When** `moai hook agent <action>` 의 action 이 관측 이벤트로 매핑되고(현재 코드 기준: `-verification`·`-post-transformation`·`-post-implementation` 은 PostToolUse, `-completion` 은 SubagentStop) 그 stdin 을 파싱하지 못하면, the agent hook dispatcher **shall** 현재 동작 — stderr 경고 한 줄, stdout `{}`, exit 0, 디스패치 없음 — 을 유지한다.
- **REQ-HSF-016** (Event-driven + Unwanted): **When** 탈출 장치 환경 변수가 켜짐 값으로 설정돼 있으나 두 설정 파일 가운데 존재하는 것이 읽히지 않거나 JSON 으로 해석되지 않으면, the hook dispatcher **shall not** 탈출 장치를 활성화로 인정한다. **When** 같은 조건이 성립하면, the hook dispatcher **shall** fail-closed 거부를 그대로 내고, 그 불인정 사실을 stderr 한 줄로 남긴다. 파일이 존재하지 않는 것은 판독 실패가 아니라 「선언 없음」으로 취급한다 — 설정 파일이 없는 프로젝트에서도 정당한 경로가 동작해야 하기 때문이며, 이 해석은 운영자가 2026-09-24 에 확인했다(plan.md §B.1 Q1). 이 해석이 추가로 여는 「파일 삭제」 우회 변형은 §F.2 「추가 후 제거」 행에 잔여 위험으로 적었다.

---

## §D 제외 범위

아래 항목은 이 SPEC 의 범위 밖이다.

### Out of Scope — 결정 이벤트 집합의 변경

- `codexadapter.DecisionBearingEvents()` 에 이벤트를 더하거나 빼는 일. §B.3 의 Can Block 11개 이벤트 중 어느 것이든 fail-closed 로 옮기려면 별도 카드가 집합 자체를 바꿔야 한다. 그 변경은 Codex 번역 표와 패리티 검증에도 파급된다.
- 번역 표의 행·열 값(Outcome) 변경. 이 SPEC 은 t1099 표를 소비만 한다.

### Out of Scope — 다른 stdin 진입점

- (0.3.0 에서 이동) `runAgentHook` (`internal/cli/hook.go:447`, 파싱 실패 분기 `:455-463`) 은 운영자 판정 Q3 에 따라 **범위 안으로 옮겼다** — REQ-HSF-014·015. 이 절에 남는 것은 아래 진입점뿐이다.
- `runSpecStatus` (`:526-535`, 파싱 실패 시 exit 1), `runSessionStartCompact` (`:560-568`), harness-observe 계열 `readNormalizedHookInput` (`:707-713`), `security-turn` / codex review gate / multi review gate 의 자체 stdin 처리.
- 빈 stdin 을 성공으로 처리하는 `ReadInput` 의 동작(`protocol.go:52-56`). 빈 페이로드로 핸들러가 기본 입력을 받아 실행되는 것은 파싱 실패와 다른 경로다(plan.md 추적 항목 Q4).

### Out of Scope — 프로토콜 계층

- `internal/hook/protocol.go` 의 `ReadInput` · `maxHookInputBytes` · `normalizeHookInput` 변경. 5 MiB 상한과 `encoding/json` 의 깊이 상한은 그대로 둔다. 이 SPEC 은 오류를 받은 **호출자**의 처리만 바꾼다.
- `internal/hook` 패키지의 어떤 파일도 수정하지 않는다(t1099 와 같은 원칙: 이음매는 CLI 계층에 둔다).

### Out of Scope — 핸들러 실패 경로

- 디스패치 오류·타임아웃·출력 매핑 실패의 fail-closed. 그것은 t1099 M2c 가 이미 다룬다(`fabc33812:internal/cli/hook.go:338-345`). Claude 하네스의 디스패치 오류 경로(현재 exit 1)도 이 SPEC 에서 바꾸지 않는다.
- Stop 체인의 루프 상한 설계(t1099 `@MX:WARN` 이 M2d 의 몫으로 명시한 것). MoAI 자체 Stop 상한(plan.md Q1 의 선택지 A2)은 채택되지 않았다 — Codex 의 Stop 은 면제 술어로 처리한다(REQ-HSF-012·013).

### Out of Scope — 상한의 측정과 Codex 의 다른 동작 형태

- Claude 호스트의 Stop 상한이 JSON `decision:"block"` + exit 0 에도 걸리는지의 측정. 이 SPEC 은 그것을 독트린 근거로만 두고 잔여 위험으로 추적한다(plan.md 추적 항목 Q8). Kickoff 를 막지 않는다.
- Codex 대화형 TUI, 기본 설정 외의 Codex 설정, `gpt-6-astra` 외 모델에서의 Stop 상한 동작. Q2 측정은 비대화형 `codex exec` 와 기본 설정만 잰다.

### Out of Scope — 설정 파일 쓰기 방어

- 모델이 `.claude/settings*.json` 이나 `.moai/config/` 를 편집하는 것 자체를 막는 가드. 이 SPEC 은 그 면에서 온 값을 탈출 장치로 인정하지 않을 뿐이다(REQ-HSF-007). `internal/hook/pre_tool.go` 에 그런 가드가 있는지는 plan-audit 의 grep 1회로 찾지 못했을 뿐 부재가 확정되지 않았다(§F.2).

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
     internal/codexadapter/diagnostics.go \
     internal/codexadapter/output.go \
     internal/cli/hook_codex_failclosed.go \
     internal/cli/hook.go
   ```

   특히 (a) `DecisionBearingEvents()` 의 원소, (b) HarnessClaude·HarnessCodex 의 fatal_error 열 Outcome, (c) `writeCodexFailClosed` 의 서명과 기록 키, (d) `Render` 의 네 이벤트 출력 형태, (e) `Discard` 필드와 `RecordDiscards` 의 stderr 미러 형식이 §A.4·§B.4 와 같은지 확인한다. 하나라도 다르면 §B 와 acceptance.md 를 먼저 고친다(D-NEW-1 경로).

3. **Kickoff Approval 은 받았고(0.3.4), 구현은 t1099 착지 뒤에 시작한다.** 운영자가 2026-09-24 t1152 레인의 AskUserQuestion 에서 승인했다(판정 내용은 plan.md §F M0, 기록은 progress.md §E.1 「Kickoff 판정」). 구현 커밋은 t1099 가 develop 에 착지하고 이 브랜치가 그것을 흡수한 뒤 시작한다. 그 시점에 위 2의 심볼 diff 와 함께 다음 설계 전제를 다시 확인하고 progress.md 에 남긴다: (a) 설치된 codex-cli 버전이 Q2 측정 버전(0.156.1)과 같은지, 다르면 같은 방법으로 Q2 를 다시 재는지, (b) `runAgentHook` 의 action → 이벤트 매핑(`internal/cli/hook.go:469-481`)이 바뀌지 않았는지, (c) `.claude/settings.json`·`.claude/settings.local.json` 의 위치·`env` 블록 형식이 REQ-HSF-007 의 전제와 같은지.

### F.2 위험

| 위험 | 설명 | 대응 |
|---|---|---|
| Stop 루프 — Claude | 파싱 실패가 지속되면 매 Stop 이 차단된다. 파싱 실패 상태에서는 `stop_hook_active` 도 `session_id` 도 읽을 수 없어 가드식 조기 반환이나 세션별 횟수 계산이 불가능하다 | 호스트 상한(8회 연속 차단 후 해제, `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`)이 루프를 끊는다고 **가정**한다. hooks-system.md:213 은 exit 2 차단을 서술하고, JSON `decision:"block"` + exit 0 에도 상한이 걸린다는 근거는 `goal-directive.md:13` 과 그 평가기의 출력 형태(`hook_stop_goal.go:120-121`)뿐이다 — 저장소 독트린이며 **측정되지 않았다**. 잔여 위험으로 남기고 plan.md 추적 항목 Q8 로 추적한다. Kickoff 를 막지 않는다 |
| Stop 루프 — Codex | 측정(plan.md Q2): codex-cli 0.156.1 `codex exec` 는 191회 연속 block 을 받아들이며 스스로 끊지 않았다. 두 번째 호출부터 `stop_hook_active: true` 를 넘겨 주지만 파싱 실패 상태에서는 읽을 수 없다 | Codex 의 Stop 은 fail-closed 에서 면제한다(REQ-HSF-012). 관측 범위는 「기본 설정·비대화형·191회/약 10분 동안 상한 미발동」이지 무한의 증명이 아니다. 향후 Codex 가 상한을 도입하면 REQ-HSF-013 의 술어 하나가 재검토 지점이다 |
| Codex Stop 면제의 대가 | (Codex, Stop) 에서 파싱 실패 시 Stop 에 걸린 가드 핸들러가 실행되지 않고 턴이 끝난다 — 이 한 쌍에서는 fail-open 이 남는다 | Stop 은 위험한 도구 호출을 막는 게이트가 아니라 턴 종료를 미루는 게이트다. 면제는 stderr 한 줄과 면제 전용 기록으로 소리를 낸다(REQ-HSF-012). 무기한 턴 잠김보다 싸다는 것이 운영자 판정 A1 의 근거다 |
| UserPromptSubmit 잠김 | 호스트 형식 변화처럼 지속적인 원인이면 사용자의 모든 프롬프트가 차단되어 세션을 쓸 수 없다 | 차단 사유가 파싱 실패와 운영자 문서 식별자를 알린다(REQ-HSF-010). 운영자는 문서에서 탈출 장치를 찾아 켠다. 잠김은 소리가 나고, 우회는 소리가 나지 않는다 — 이 비대칭이 fail-closed 를 택하는 근거다 |
| 호스트 형식 변화의 파급 | `6a3603274` 가 막으려던 상황, 즉 새 호스트 버전에서 페이로드 형식이 바뀌어 매 호출이 파싱 실패하는 경우, 결정 이벤트 전부가 거부된다 | 관측 이벤트는 보존되므로 세션 시작·종료·관측 경로는 영향이 없다. 탈출 장치로 복구 가능. 기록(REQ-HSF-008)이 원인 진단의 근거가 된다 |
| 탈출 장치의 남용·자가 활성화 | 탈출 장치가 켜진 상태로 남으면 이 SPEC 이 없는 것과 같다. 모델이 설정 파일의 `env` 블록을 편집해 스스로 켜는 경로도 있다 | 기본 꺼짐, 적용될 때마다 stderr·영속 기록. 모델이 쓸 수 있는 프로젝트·로컬 설정면에서 온 값은 활성화로 인정하지 않는다(REQ-HSF-007, AC-HSF-010). **이 면을 감시하는 ConfigChange 가드는 존재하지 않는다** — §B.2 행 20 이 적었듯 ConfigChange 핸들러는 무조건 빈 출력을 낸다. 따라서 이 면의 완화는 출처 제약뿐이고, 그 출처 제약도 훅 실행 시점의 설정 파일 상태를 한 번 볼 뿐이다. 출처 제약이 막는 것은 설정면에 선언이 **남아 있는** 상태의 자가 활성화이며, 선언을 넣었다가 지우는 순서는 막지 못한다(다음 행) |
| 추가 후 제거 — 출처 제약의 시점 공백 | 출처 제약은 훅이 실행되는 순간의 두 설정 파일 상태만 본다. 순서: 모델이 설정 `env` 블록에 탈출 장치 키를 켜짐 값으로 추가 → 호스트가 그 값을 훅 환경에 반영 → 키를 삭제하거나(변형 1, 키 삭제) 설정 파일 자체를 삭제 → 이어지는 파손 입력. 이때 호스트가 반영했던 값을 훅 환경에 계속 유지하면, 디스패처는 「선언 없음 + 환경 변수 켜짐」을 보고 탈출 장치를 인정한다. 변형 1 은 파일 부재 해석과 무관하게 존재하고, 파일 삭제 변형은 부재를 「선언 없음」으로 보는 해석(REQ-HSF-016)이 추가로 연다 — 파일을 쓰고 지울 수 있는 모델에게 두 변형은 같은 능력이다. **성립 조건은 호스트가 설정 `env` 를 언제 훅 환경으로 전파하고, 키·파일을 지운 뒤에도 그 값을 유지하는지이며, 측정되지 않았다(가설).** AC-HSF-006 (p1)·(p2) 는 이 순서가 남긴 최종 상태와 정당한 활성화를 구별하지 못한다 | 운영자 판정(2026-09-24, plan.md §B.1 Q1): 부재 = 선언 없음 해석을 유지하고 이 순서를 잔여 위험으로 기록한다. run 착수 시 plan.md §C Pre-flight 5 가 두 전파 모형 — 세션 안에서 키를 추가하는 경우와 키를 선언한 채 새 세션을 시작하는 경우 — 을 선언된 실행·시간 상한(Kickoff 때 운영자 승인) 안에서 잰다. 삭제 뒤의 관측은 같은 실행에서 삭제 **전에** 설정 출처의 값이 훅 환경에서 관측됐고, 삭제 **뒤에** 선언이 사라졌음을 보이는 훅 기록이 한 줄 이상 있을 때만 판정에 쓴다 — 값이 애초에 오지 않은 실행이나 삭제가 실행되지 않은 실행을 「유지하지 않는다」로 읽지 않기 위해서다. 두 경로 가운데 어느 하나라도 확정되지 않으면 「미측정」이다. 「유지한다」(또는 미측정)로 나오면 구현을 시작하지 않고 운영자 재판정을 받는다 — 메커니즘 (i) 자체의 재검토를 포함한다. 어느 경로로도 값이 훅 환경에 오지 않으면(「전달 없음」) 이 순서는 그 호스트·버전에서 성립 조건이 없으므로 진행하되, 버전에 묶인 관측으로 기록하고 출처 제약은 그대로 둔다. 탈출 장치는 적용될 때마다 stderr·영속 기록을 남기므로, 이 순서가 성립해도 적용 사실은 소리를 낸다. 운영자 판정 범위 밖의 선택지(기록만, 채택 아님): 한 세션에서 한 번이라도 선언이 관측된 키는 이후 선언이 사라져도 계속 불인정하는 강화. 파싱 실패 상태에서는 stdin 의 `session_id` 를 읽을 수 없으므로 stdin 밖의 세션 열쇠와 상태 저장이 필요하다 — 선택지 A2 와 같은 미측정 전제다 |
| 출처 제약의 잔여 면 | 메커니즘(셸 환경 변수 + 프로젝트·로컬 `env` 선언 시 무시)으로도 막지 못하는 면: **사용자 범위 `~/.claude/settings.json` 의 `env`**. 호스트가 그 블록을 훅 환경으로 전파하면 디스패처는 그 값을 셸 환경 변수와 구별하지 못하고, 모델의 파일 도구가 그 경로에 쓸 수 있는 권한 설정이면 모델이 스스로 켤 수 있다. Codex 하네스의 프로젝트 설정면(`.codex/` 아래)이 훅 환경에 변수를 공급하는 경로가 있는지도 측정하지 않았다. 설정 `env` 블록이 훅 프로세스 환경으로 언제 전파되는지(편집 즉시인지, 세션 재시작 후인지)도 측정하지 않았다 | 잔여 면으로 명시하고 판정 기록(plan.md §B.1)에 남긴다. 탈출 장치가 적용될 때마다 stderr·영속 기록이 남으므로 적용 사실은 소리를 낸다. 설정 파일 쓰기 자체를 막는 가드의 존재 여부는 미확정(§D) |
| 설정 판독 실패의 처리 | 두 설정 파일 가운데 존재하는 것이 읽히지 않으면 「선언 없음」을 확인할 수 없다 | 활성화를 인정하지 않는다(REQ-HSF-016). 파일 부재만은 「선언 없음」으로 본다 — 이 구분은 운영자 판정 문구(「판독 불가 시 불인정」)를 구현 가능하게 옮긴 manager-spec 의 해석으로 시작했고, 운영자가 2026-09-24 에 확인했다(plan.md §B.1 Q1). 이 해석이 여는 파일 삭제 변형은 「추가 후 제거」 행에 있다 |
| codex 모드 agent 경로의 도달성 | AC-HSF-012·013 의 절반(`--harness codex` 로 `moai hook agent` 를 부르는 32 경우)은 현재 배포 설정에서 도달하지 않는 경로를 검증한다. 배포 래퍼 `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh:47`(`.sh.tmpl:47`, 로컬 `.claude/hooks/moai/handle-agent-hook.sh:47` 도 같음)은 `moai hook agent "$1"` 을 `--harness` 없이 부른다. `internal/template/templates/.codex/` 에는 `agents/` 만 있고 훅 호출이 없으며, Codex 쪽 훅 명령 행은 `internal/cli/codex_readiness.go:74` 의 `moai hook --harness codex` 다 — 이 트리에서 `hook agent` 를 부르는 Codex 설정 생성 경로는 grep 으로 찾지 못했다(부재 확정은 아니다). 또 성공 경로의 agent 호출은 Codex 번역을 거치지 않으므로, 이 경로가 도달한다면 실패 경로만 Codex 렌더링을 하는 비대칭이 된다 | 범위는 바꾸지 않는다(운영자 판정 Q3 의 포함 범위 유지). plan.md 추적 항목 Q9 로 기록한다 |
| 사유 문구를 통한 우회 절차 노출 | 거부 사유는 모델이 읽는다. 사유에 활성화 절차가 실리면 방어 대상에게 우회 방법을 알려 주는 셈이다 | REQ-HSF-010 이 사유에 운영자 문서 식별자만 싣게 하고, AC-HSF-001(e)·002(e) 가 탈출 장치 식별자의 부재를 단언한다 |
| 기존 테스트의 의도 역전 | `TestRunHookEvent_MalformedStdinGraceful` 이 pre-tool 로 `6a3603274` 의 의도를 단언한다 | acceptance.md §D 에 갱신 대상과 갱신 방향을 적었다 |
