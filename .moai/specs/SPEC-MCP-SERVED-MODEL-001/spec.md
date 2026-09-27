---
id: SPEC-MCP-SERVED-MODEL-001
title: "MCP 위임 도구의 서빙 모델 관측 — glm_task·codex_task 결과와 작업 기록에 요청 모델과 서빙 모델을 함께 남기기"
version: "0.1.1"
status: completed
created: 2026-09-28
updated: 2026-09-28
author: manager-spec (card t1284)
priority: P2
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "mcp, glm-task, codex-task, served-model, delegation, record-plus-warn"
tier: S
era: V3R6
related_specs: [SPEC-SERVED-MODEL-AUDIT-001]
---

# SPEC: MCP 위임 도구의 서빙 모델 관측

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-28 | manager-spec (card t1284) | 최초 draft. 전제 실측 `.moai/reports/t1284/premise.md` 와 plan 단계 코드 판독(§A.2)에 근거. 형제 SPEC-SERVED-MODEL-AUDIT-001 의 설계 (c) 가운데 "기록하고 경고하되 거부하지 않는다" 자세만 MCP 위임 경로로 옮긴다. |
| 0.1.1 | 2026-09-28 | manager-spec (card t1284) | plan-audit(`.moai/reports/t1284/plan-audit.md`) 반영: D1 AC-MSM-003 을 설정 기반 경로로 재작성(codex_task 에 `model` 인자를 더하지 않음), D2 비문자열 `model` 행과 디코드 비실패 조항, D3 import 판독을 diff 기반 grep 으로 교체, D4 REQ-MSM-005 정규화 범위 한정, D5 codex 경고 부재 관측식 정의, D7 백그라운드 실패 기록 행 추가. |

---

## §A 배경

### §A.1 결함

SPEC-SERVED-MODEL-AUDIT-001 은 서브에이전트 트랜스크립트에서 서빙 모델을 관측하게 했다. 그러나 MoAI MCP 서버의 위임 도구 두 개는 그 관측 밖에 있다.

- **glm_task**: 요청에는 모델을 싣지만 z.ai 응답의 `model` 필드는 버려진다. 응답은 `glmMessagesResponse`(`internal/cli/mcp_glm.go:142-147`)로 디코드되는데 이 구조체에는 `Content` 만 있다. 동기 결과의 `model`(`internal/cli/glm_task.go:175`)과 백그라운드 작업 기록의 `model`(`internal/cli/glm_jobs.go:127-129`, 주석 "the GLM model id the task was sent to")은 둘 다 **요청 모델**이다.
- **codex_task**: 결과(`CodexTaskResult`, `internal/cli/codex_task.go:201`)에도 작업 기록(`CodexJobRecord`, `internal/cli/codex_jobs.go:135`)에도 모델 필드가 없다.

그래서 "이 출력은 어떤 모델이 만들었는가"라는 질문에 두 도구 모두 요청값만 답하거나 아무것도 답하지 않는다.

### §A.2 plan 단계 코드 판독 (base `e9577de4f`)

- **전제 실측**: `.moai/reports/t1284/premise.md` — t1282 서빙 모델 코드가 이 카드의 base 에 포함됨(`merge-base --is-ancestor 1971f5501 HEAD` → YES), 두 위임 경로 모두 서빙 모델을 관측하지 않음.
- **RED-now 프로브**: `git grep -n -E 'served_model|ServedModel' -- internal/cli/glm_task.go internal/cli/glm_jobs.go internal/cli/mcp_glm.go internal/cli/codex_task.go internal/cli/codex_jobs.go` → 출력 없음, exit 1. 같은 패턴의 양성 대조 `git grep -c -E 'served_model|ServedModel' -- internal/hook/served_model.go` → `internal/hook/served_model.go:20`. 도구는 발화하고, 대상 파일에는 서빙 모델 개념이 없다.
- **glm_audit 공유**: `glmMessagesResponse` 는 `callGLMTask`(`glm_task.go:337`)와 `parseGLMReview`(`mcp_glm.go:343`) 두 곳에서만 쓰인다(`git grep -n glmMessagesResponse`).
- **codex 프로토콜**: codex_task 는 `--json` 이벤트 스트림이 아니라 codex app-server JSON-RPC 세션(`initialize` → `thread/start` 또는 `thread/resume` → `turn/start`)을 쓴다(`internal/cli/mcp_codex.go:772-817`). `thread/start` 응답에서는 `result.thread.id` 만 읽고(`extractThreadID`, `:1120`), 턴 알림에서는 `turn/started`·`item/completed`·`turn/completed` 만 읽는다(`awaitCodexTurnReview`, `:1287`). 모델 식별자를 읽는 코드는 없다. 테스트 픽스처의 `thread/start` 응답도 전부 `{"id":2,"result":{"thread":{"id":"tid-fake"}}}` 모양이다(`codex_jobs_test.go:44` 외). 요청 쪽에서는 `resolveCodexModelEffort`(`mcp_codex.go:223`)가 모델을 정하면 그 값을 `thread/start` 파라미터 `model` 로 보낸다(`:805-807`).
- **GLM 모델 id 모양**: `internal/config/defaults.go:224-226` — z.ai 는 `[1m]` 접미사가 붙은 id 를 모르는 모델로 거부하므로 GLM 모델 id 는 접미사 없는 bare id 다.

### §A.3 확인하지 않은 것 (plan 단계 Gap)

- z.ai 응답 봉투가 최상위 `model` 필드를 **실제로** 채워 보내는지, 채운다면 요청 id 를 그대로 되돌려주는지는 이 머신에서 실측하지 않았다. Anthropic 호환 응답 봉투가 `model` 을 싣는다는 것은 호환 형식의 전제일 뿐이다. 그래서 응답에 `model` 이 없을 때의 동작(REQ-MSM-004)을 요구사항으로 둔다.
- codex app-server 응답이 서빙 모델을 담는지는 코드와 테스트에서 확립되지 않는다. 확립되지 않은 필드를 읽는 코드를 만들지 않고, 서빙 모델을 `unknown` 으로 명시 기록한다(REQ-MSM-006).

---

## §B 용어

- **요청 모델(requested model)**: 도구가 백엔드에 보낸 모델 id. glm_task 는 호출자 지정값 또는 해석된 기본값, codex_task 는 `thread/start` 에 실어 보낸 값(보내지 않았으면 빈 값 — codex 자체 설정 기본값이 적용되며 MoAI 는 그 값을 모른다).
- **서빙 모델(served model)**: 백엔드 응답이 스스로 밝힌 모델 id.
- **서빙 경고(served-model warning)**: 요청 모델과 서빙 모델이 다르거나 서빙 모델이 없을 때 결과에 붙는 사람이 읽는 문장.

---

## §C 요구사항 (GEARS)

### §C.1 GLM 위임 경로

- **REQ-MSM-001** (Ubiquitous) — The GLM task call shall take the served model of a completed call from the top-level `model` field of the z.ai response envelope, trimmed of surrounding whitespace, and shall treat an absent, non-string, or empty field as no served model without failing the decode of the rest of the envelope.
- **REQ-MSM-002** (Event-driven) — **When** a foreground `glm_task` call completes, the tool result shall carry both the requested model and the served model, and the existing requested-model field shall keep its current meaning.
- **REQ-MSM-003** (Event-driven) — **When** a background GLM job reaches the completed status, its durable job record — and therefore the job-status output that renders that record — shall carry the served model alongside the requested model.
- **REQ-MSM-004** (Event-driven) — **When** a completed GLM call's served model differs from its requested model under case-insensitive comparison, or the completed call carries no served model, the foreground result or the background job record shall carry a served-model warning naming the requested model and the served model (or its absence); any note the result already carries shall be preserved.
- **REQ-MSM-005** (Unwanted) — The GLM served-model observation shall not change the status, output, or error of any task, shall not refuse or fail a call on a served-model ground, shall not attach a served-model warning to a call that did not complete, and shall not remove suffixes (such as `[1m]`) from, rewrite, or alias-match either model id before comparing them, beyond trimming surrounding whitespace (REQ-MSM-001) and the case-insensitive comparison (REQ-MSM-004).

### §C.2 Codex 위임 경로

- **REQ-MSM-006** (Ubiquitous) — The `codex_task` result and the codex job record shall carry the requested model the session sent in its thread request (empty when none was sent) and shall carry the literal served model `unknown`, because no response the tool consumes is established to carry a model identifier.
- **REQ-MSM-007** (Unwanted) — The codex delegation path shall not report a served model equal to the requested model, shall not derive a served model from any field whose presence the code and tests have not established, and shall not attach a served-model warning to every call on account of the structural `unknown`.

### §C.3 범위 고정

- **REQ-MSM-008** (Unwanted) — The change shall not alter the `glm_audit` result contract (the shared review output returned by the GLM, codex, and multi-backend audit tools), shall not add a doctor check, and shall not add any gate, refusal, or spawn denial.

---

## §D 인수 기준 요약

Tier S 이지만 카드 지시에 따라 인수 기준 본문은 `acceptance.md` 에 둔다. 요약:

| AC | REQ | 요지 |
|----|-----|------|
| AC-MSM-001 | REQ-MSM-001, REQ-MSM-002, REQ-MSM-004, REQ-MSM-005 | glm_task 동기 결과 표: 일치·불일치·부재·대소문자만 다름·접미사 차이·실패 호출·숫자형 `model` |
| AC-MSM-002 | REQ-MSM-003, REQ-MSM-004 | 백그라운드 GLM 작업 기록과 job-status 출력에 요청·서빙 모델과 경고, 실패 기록에는 경고 없음 |
| AC-MSM-003 | REQ-MSM-006, REQ-MSM-007 | codex_task 결과와 작업 기록: 설정 기반 요청 모델 기록, 서빙 `unknown`, 경고 없음(관측식 명시) |
| AC-MSM-004 | REQ-MSM-005, REQ-MSM-008 | glm_audit 판정 불변(응답에 문자열·숫자형 `model` 이 있어도), 기존 GLM·codex 테스트 회귀 없음 |
| AC-MSM-005 | REQ-MSM-001..008 (재측정) | 재측정 묶음(형제 가드·spec 패키지·CI 판 린트) |

---

## §E 비기능 요구사항

- **fail-open**: 서빙 모델 관측은 어떤 경우에도 호출을 실패시키지 않는다 — REQ-MSM-005.
- **하위 호환**: 기존 `model` 키의 의미(요청 모델)는 바꾸지 않는다. 새 정보는 새 필드로 더한다. 이전에 쓰인 작업 기록 파일은 새 필드 없이 그대로 읽혀야 한다.
- **템플릿 중립성**: 이 변경은 `internal/cli` Go 코드와 테스트에 한정되며 `internal/template/templates/**` 를 건드리지 않는다.

---

## §F 제외 범위 (Exclusions)

이 절은 이번 SPEC 에서 만들지 않는 것을 적는다.

### Out of Scope — 강제 수단과 사후 스캔

- 서빙 모델 불일치를 이유로 호출을 거부·재시도하거나 결과 채택을 거부하는 것 — 기록과 경고만 한다.
- `moai doctor` 에 MCP 작업 기록을 훑는 점검을 더하는 것 — 형제 SPEC 의 사후 스캔은 트랜스크립트 판독기라 이 기록에 재사용할 수 없다(plan.md §B.4).

### Out of Scope — 인접 경로

- `glm_audit`·`codex_audit`·`audit_multi` 의 서빙 모델 관측 — 세 도구가 공유하는 리뷰 출력 계약을 바꾸는 일이라 별도 카드 몫이다(plan.md §B.2).
- codex rollout 파일(`~/.codex/sessions/**`)을 읽어 서빙 모델을 복원하는 것 — app-server 세션이 쓰는 rollout 과 이 세션의 대응이 확립되지 않았다.
- `glm_job_result` 출력에 모델 필드를 더하는 것 — 기록은 job-status 로 이미 보인다.
- `codex_task` 도구에 `model` 인자를 더하는 것 — codex 요청 모델은 지금처럼 프로젝트 llm 설정 해석값에서만 나온다(REQ-MSM-006 은 이미 보낸 값을 기록할 뿐 새 입력을 요구하지 않는다).

### Out of Scope — 형제 SPEC 의 부채

- t1282 부채 F1(서브에이전트 서빙 모델 비교기가 Claude `[1m]` 접미사를 정규화하지 않아 생기는 거짓 drift) — 이 카드의 비교는 그 비교기를 쓰지 않으므로 이 카드의 판정에 영향이 없다. 결정과 근거는 plan.md §B.1.
