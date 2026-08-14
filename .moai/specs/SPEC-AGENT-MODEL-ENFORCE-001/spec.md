---
id: SPEC-AGENT-MODEL-ENFORCE-001
title: "서브에이전트 spawn 시 per-agent 모델 프로파일 기계적 관측·집행 — 침묵 폴백 종식"
version: "0.1.0"
status: completed
created: 2026-08-08
updated: 2026-08-14
author: manager-spec
priority: P1
phase: "v3.1.0"
module: "internal/hook"
lifecycle: spec-anchored
tags: "hook, pretooluse, agent-spawn, model-profile, observability, fail-open, opt-in-gate"
tier: M
era: V3R6
issue_number: 1376
related_specs: [SPEC-MODEL-PROFILE-MATRIX-001, SPEC-MODEL-PROFILE-MATRIX-002, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001, SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-08 | manager-spec | 최초 draft. issue #1376 신고 3건 + 오케스트레이터 실측 브리프 + 본 SPEC 저작 중 추가 실측 3건(F4/F5/F6) 기반. 관측→권고→(opt-in)차단 계층 설계. |

---

## §A Context & Motivation

`moai model profile --json`(`internal/cli/model.go`)은 활성 프로파일과 per-agent 오버라이드를 해석해 11개 리테인드 에이전트 각각의 `{model, effort}`를 산출한다. 해석기 `template.ResolveAgentModelEffort`는 정상 동작한다. 그러나 **그 산출물이 실제 spawn에 반영됐는지를 확인하는 기계적 표면이 하나도 없다.** 프로파일은 계산되지만 아무도 검사하지 않는다.

### §A.1 실측 Findings

전체 명령과 출력은 acceptance.md §B에 있다. 라인 앵커는 실측 시점 기준이며 run-phase에서 content-token으로 재검증한다.

| # | 결함 | 실측 근거 | 처분 |
|---|------|-----------|------|
| F1 | **해석기는 있으나 검사자가 없다** | `internal/cli/model.go`가 `--json` 리포트를 방출하고 `internal/template/profile_matrix.go`가 33셀 SSOT를 보유한다. spawn 인자를 읽어 이 값과 대조하는 코드는 리포지터리 전체에 0건 | 관측 계층 신설(M2) |
| F2 | **PreToolUse가 Agent를 보지 않는다** | `internal/template/templates/.claude/settings.json.tmpl`의 유일한 PreToolUse 항목은 `"matcher": "Write|Edit|Bash"`. Agent/Task는 어떤 PreToolUse 훅에도 관측되지 않는다 | 별도 matcher 블록 추가(M2) |
| F3 | **PostToolUse Agent 분기는 모델을 읽지 않는다** | `internal/hook/post_tool.go`의 `input.ToolName == "Agent" \|\| input.ToolName == "Task"` 분기는 `logTaskMetrics(input)`만 호출하며 입출력 크기 텔레메트리만 다룬다. `model` 필드를 읽는 코드는 없다 | 해당 분기는 본 SPEC의 관측 지점으로 채택하지 **않는다**(§A.3 D2) |
| F4 | **그 분기는 애초에 도달 불가다 (신고자 미식별)** | `handle-post-tool.sh` → `moai hook post-tool`은 PostToolUse의 `"matcher": "Write\|Edit\|MultiEdit"` 블록에만 등록돼 있다. matcher 없는 두 번째 PostToolUse 블록은 `handle-harness-observe.sh`(→ `moai hook harness-observe`)로, 다른 핸들러이며 opt-in 게이트 하위다. 즉 F3의 Agent 분기는 배포 배선에서 **결코 발화하지 않는다** | (b) 안의 전제를 무효화 — §A.3 D2 |
| F5 | **규칙이 구조적으로 시야에 없다 (신고자 미식별)** | `.claude/rules/moai/development/model-policy.md`는 frontmatter `paths: "**/.claude/agents/**"`를 갖는다. 일반 `/moai run` 세션은 에이전트 파일을 건드리지 않으므로 이 규칙은 로드되지 **않는다**. 실패는 "LLM이 규칙을 잊었다"가 아니라 "규칙이 애초에 시야에 없다" | 규칙 가시성 분해(M5) |
| F6 | **주입은 실제로 거의 일어나지 않는다 (신규 실측 — 정량)** | 로컬 세션 트랜스크립트 54개 전수 스캔: Agent/Task tool_use 호출 **156건** 중 `model` 인자를 담은 것은 **1건**(0.6%). 관측된 인자 키 집합은 `{description, subagent_type, run_in_background, prompt, name, model(1), isolation(2)}`. 유일한 `model` 보유 호출의 값은 문자열 별칭 `"opus"` | 판정 논리를 "불일치"가 아니라 **"누락"** 중심으로 설계(§A.3 D1) |
| F7 | **폴백의 폭발 반경** | `.claude/agents/moai/*.md` 11개 파일 중 10개가 `model: inherit`, `manager-git`만 `model: sonnet`. 주입이 누락되면 10개 에이전트가 전부 부모 세션 모델로 침묵 강등된다 | 관측 대상은 리테인드 카탈로그 전체 |

### §A.2 F6이 문제 정의를 바꾼다

브리프와 신고서는 암묵적으로 "주입된 모델이 프로파일과 **다를** 수 있다"(불일치)를 상정했다. 실측은 더 단순하고 더 나쁜 상태를 보여준다: 주입이 **거의 없다**(155/156 누락). 따라서 spawn 페이로드의 `model`을 해석된 프로파일과 비교하는 판정은 실질적으로 "선언값 부재 vs 해석값 존재"의 대조로 수렴한다. 이는 결함이 아니라 설계 입력이다 — 판정 논리는 세 가지 결과(`ok` / `missing` / `mismatch`)를 모두 표현해야 하며, 지배적 케이스는 `missing`이다.

### §A.3 훅이 볼 수 있는 것과 볼 수 없는 것 (정직성 계약)

본 SPEC은 검증되지 않은 능력을 전제하지 않는다.

- **검증됨**: Agent 도구는 `model` 인자를 **받는다** — 트랜스크립트에 실제 값(`"opus"`)을 담은 호출이 존재한다. 즉 주입 채널 자체는 실존한다.
- **검증됨**: `HookInput.ToolInput`은 `json.RawMessage`로, Claude Code가 보내는 `tool_input`을 그대로 담는다(`internal/hook/types.go`). 파싱 제약은 없다.
- **검증되지 않음**: **PreToolUse 훅이 Agent/Task 도구에 대해 실제로 발화하는지**, 그리고 발화한다면 그 stdin `tool_input`이 트랜스크립트에서 관측된 인자 집합을 그대로 담는지. 현재 matcher가 Agent를 제외하므로 이에 대한 관측 증거는 **0건**이다.
- **검증되지 않음**: `HookInput.Model` 필드는 SessionStart 전용으로 문서화돼 있다. Agent PreToolUse 이벤트가 이 필드를 채우는지는 미확인이며, 본 SPEC은 이를 전제하지 않는다.

따라서 M1은 기능 구현이 아니라 **측정**이다. M2 이후의 모든 요구사항은 M1이 산출한 실제 페이로드 픽스처 위에서만 성립한다.

### §A.4 effort는 집행 대상이 아니다

`model-policy.md`의 채널 표는 Agent 도구에 `effort` 파라미터가 **존재하지 않는다**고 기록한다. 표준 sub-agent 경로에서 effort의 유일한 전달 채널은 에이전트 파일의 frontmatter `effort:`이다. 따라서 본 SPEC의 집행 범위는 `model` 단일 축이며, effort는 §C에서 명시적으로 배제된다.

---

## §B Requirements (GEARS)

### §B.1 M1 — 페이로드 실측 게이트

- **REQ-AME-001**: 구현은 Agent/Task 도구 spawn에 대한 PreToolUse stdin 페이로드를 **실제로 1건 이상 포획**해야 한다(shall). 포획 결과는 `internal/hook/testdata/` 하위의 픽스처 파일로 커밋되어야 한다.
- **REQ-AME-002**: 포획 결과는 다음 세 질문에 대한 관측된 답을 산출물에 기록해야 한다(shall): (a) PreToolUse가 Agent/Task에 대해 발화하는가, (b) `tool_input`이 `subagent_type`을 담는가, (c) `tool_input`이 `model` 키를 담을 수 있는가.
- **REQ-AME-003**: **When** M1 측정 결과 PreToolUse가 Agent/Task에 대해 발화하지 않는 것이 관측되면, 구현은 M2 이후를 진행하지 않고 대체 관측 지점(SubagentStart / PostToolUse 배선 확장)으로 설계를 재라우팅하고 그 사유를 산출물에 기록해야 한다(shall). 관측되지 않은 능력 위에 후속 밀스톤을 쌓지 **않아야 한다(shall not)**.

### §B.2 M2 — 관측 계층 (기본 활성, 차단 없음)

- **REQ-AME-010**: PreToolUse 훅 설정은 Agent/Task를 관측하는 **별도 matcher 블록**을 가져야 한다(shall). 기존 `Write|Edit|Bash` 블록의 matcher 문자열은 변경되지 **않아야 한다(shall not)** — Write/Edit/Bash 경로의 지연·타임아웃 특성이 바뀌지 않아야 한다.
- **REQ-AME-011**: **When** Agent 또는 Task 도구에 대한 PreToolUse 이벤트가 감지되면, 훅 핸들러는 `tool_input`에서 에이전트 식별자(`subagent_type`)와 선언된 모델(`model`, 부재 가능)을 추출해야 한다(shall).
- **REQ-AME-012**: 핸들러는 추출한 에이전트 식별자를 `template.ResolveAgentModelEffort`에 전달해 해석된 모델 별칭을 얻어야 하며(shall), 프로파일 매트릭스를 재선언하거나 모델 별칭을 리터럴로 하드코딩하지 **않아야 한다(shall not)**.
- **REQ-AME-013**: 핸들러는 각 spawn에 대해 `.moai/logs/agent-model-audit.jsonl`에 한 줄의 구조화 레코드를 추가해야 한다(shall). 레코드는 최소한 타임스탬프, session_id, 에이전트 식별자, 선언 모델(부재 시 빈 값), 해석 모델, 판정(`ok` / `missing` / `mismatch` / `unmapped`)을 담아야 한다.
- **REQ-AME-014**: 판정은 다음과 같이 정의되어야 한다(shall): 해석 결과가 매핑되지 않은 에이전트(`inherit` 센티널)면 `unmapped`; 해석 모델이 구체적 별칭인데 선언이 부재하면 `missing`; 선언과 해석이 다르면 `mismatch`; 같으면 `ok`.
- **REQ-AME-015**: M2 단계에서 훅은 **어떤 경우에도** deny 또는 ask 결정을 반환하지 **않아야 한다(shall not)** — 판정 결과와 무관하게 allow 폴스루해야 한다(shall).
- **REQ-AME-016**: **When** 로그 파일 쓰기가 실패하면(권한 부재, 디스크 부족, 프로젝트 루트 미해석), 핸들러는 오류를 전파하지 않고 allow 폴스루해야 한다(shall) — 관측 실패가 spawn을 막지 **않아야 한다(shall not)**.

### §B.3 M3 — 권고 계층

- **REQ-AME-020**: **When** 판정이 `missing` 또는 `mismatch`인 spawn이 감지되면, 훅은 비차단 권고 메시지를 방출해야 한다(shall). 메시지는 에이전트 이름, 해석된 모델 별칭, 그리고 취해야 할 시정(해당 별칭을 per-spawn `model` 인자로 주입)을 담아야 한다.
- **REQ-AME-021**: 권고 방출은 훅의 종료 코드를 바꾸지 **않아야 한다(shall not)** — allow 폴스루가 유지되어야 한다.
- **REQ-AME-022**: 훅은 `AskUserQuestion` 또는 `mcp__askuser__*`를 호출하지 **않아야 한다(shall not)**. 사용자 상호작용은 오케스트레이터 전용 채널이다.

### §B.4 M4 — opt-in 차단 게이트

- **REQ-AME-030**: 설정은 `workflow.agent_model_guard.enabled` 불리언 키를 제공해야 하며(shall), 배포 기본값은 `false`여야 한다. 이는 `workflow.branch_guard.enabled`의 선례를 그대로 따른다.
- **REQ-AME-031**: **Where** `workflow.agent_model_guard.enabled`가 `false`인 동안, 훅은 차단 판정을 산출하지 **않아야 한다(shall not)**. 관측(M2)과 권고(M3)는 게이트와 무관하게 계속 동작해야 한다(shall).
- **REQ-AME-032**: **Where** 게이트가 `true`인 동안, **When** 판정이 `mismatch`인 spawn이 감지되면 훅은 deny 결정과 `AGENT_MODEL_VIOLATION:` 접두사를 가진 사유를 반환해야 한다(shall).
- **REQ-AME-033**: 차단은 **긍정적 증거에만** 발화해야 한다(shall) — 에이전트 식별자를 추출했고, 해석이 매핑됐고, 선언 모델이 존재하며, 두 값이 다른 경우에 한한다. 그 외 모든 불확실(페이로드 파싱 실패, 에이전트 미매핑, 설정 미해석, 프로젝트 루트 미해석)에서는 allow 폴스루해야 한다(shall).
- **REQ-AME-034**: 게이트가 `true`이더라도 판정 `missing`은 차단되지 **않아야 한다(shall not)** — 155/156 누락 실측(§A.1 F6)을 감안하면 `missing` 차단은 모든 spawn을 사실상 봉쇄한다. `missing`은 권고 대상에 머물러야 한다.
- **REQ-AME-035**: 게이트가 `false`인 경로에서는 프로파일 해석 이외의 추가 비용(파일 I/O를 유발하는 설정 재로드 등)이 발생하지 **않아야 한다(shall not)**.

### §B.5 M5 — 규칙 가시성

- **REQ-AME-040**: per-spawn 모델 주입 의무를 기술하는 규칙 조각은 세션에 **항상 로드되는** 표면에 존재해야 한다(shall) — 즉 `paths:` frontmatter로 스코프되지 않아야 한다.
- **REQ-AME-041**: **Where** 항상 로드되는 조각을 신설하는 경우, 그 조각은 압축된 스텁이어야 하며(shall) 기존 `model-policy.md` 본문(27.5KB)을 통째로 항상 로드로 전환하지 **않아야 한다(shall not)** — 현재 항상 로드 규칙 총량은 13개 파일 197KB이며 전체 전환은 이를 14% 증가시킨다.
- **REQ-AME-042**: 신설 스텁은 상세 규칙(`model-policy.md`)을 교차 참조해야 하며(shall), 프로파일 매트릭스 셀 값이나 모델 별칭 목록을 재선언하지 **않아야 한다(shall not)**.
- **REQ-AME-043**: `model-policy.md`의 기존 `paths:` 스코프는 유지되어야 한다(shall) — 에이전트 파일 편집 시의 상세 규칙 로드 동작은 회귀하지 않아야 한다.

### §B.6 M6 — 횡단 마감

- **REQ-AME-050**: **When** `.claude/` 하위 파일에 대한 편집이 발생하면, 동일 커밋이 `internal/template/templates/` 하위의 대응 미러 편집과 `make build` 재생성 산출물을 함께 담아야 한다(shall).
- **REQ-AME-051**: `internal/template/templates/` 하위로 배포되는 산출물은 SPEC ID, REQ/AC 토큰, 내부 작업 날짜, commit SHA를 포함하지 **않아야 한다(shall not)**.
- **REQ-AME-052**: 신설 훅 배선은 셸 래퍼 → `moai hook <event>` 서브커맨드 형태를 따라야 하며(shall), `$CLAUDE_PROJECT_DIR`는 항상 따옴표로 감싸야 하고 타임아웃은 5초여야 한다.
- **REQ-AME-053**: `internal/hook` 패키지의 구문 커버리지는 90% 이상이어야 한다(shall).
- **REQ-AME-054**: 신설 로그 파일 `.moai/logs/agent-model-audit.jsonl`은 SessionEnd 로그 age-out 대상에 등록되어야 한다(shall) — 무한 성장하지 않아야 한다.

---

## §C Out of Scope

본 SPEC이 **만들지 않는** 것. 각 항목은 의도적 배제이며 누락이 아니다.

### Out of Scope — effort 집행

- Agent 도구에는 `effort` 파라미터가 존재하지 않으므로(§A.4) effort 값의 spawn-time 집행은 불가능하다. frontmatter `effort:` 값의 검증·재작성, GLM effort 오버레이 변경은 본 SPEC 범위 밖이다.

### Out of Scope — 오케스트레이터 자동 주입

- 훅이 spawn 인자를 **수정**해 모델을 자동 주입하는 것. PreToolUse는 allow/deny/ask 결정 채널이지 인자 변형 채널이 아니며, 그런 변형은 관측 가능성 없는 침묵 개입이다. 본 SPEC은 관측·권고·차단만 다룬다.

### Out of Scope — 프로파일 매트릭스 내용 변경

- 33셀 `defaultProfileMatrix`의 셀 값, 에이전트 그룹 분류, `llm.profiles` 스키마. 본 SPEC은 해석기의 **소비자**이지 소유자가 아니다.

### Out of Scope — 하네스·사용자 생성 에이전트 집행

- `.claude/agents/harness/` 하위 생성 스페셜리스트는 리테인드 카탈로그 밖이라 해석기가 `inherit`을 반환한다. 이들은 판정 `unmapped`으로 기록만 되고 권고·차단 대상이 아니다.

### Out of Scope — PostToolUse 배선 확장

- F4가 드러낸 "PostToolUse Agent 분기 도달 불가" 문제의 해소(`matcher`에 Agent 추가, `logTaskMetrics` 부활). 이는 텔레메트리 위생 문제이며 별도 SPEC 소관이다. 단, M1 측정 결과 PreToolUse가 Agent에 발화하지 않는 경우 REQ-AME-003의 재라우팅 대상으로 승격될 수 있다.

### Out of Scope — 다른 훅 이벤트의 모델 관측

- SubagentStart / SubagentStop 이벤트에 모델 필드를 추가하는 업스트림 프로토콜 변경 요청.

---

## §D 성공 기준

1. Agent/Task spawn에 대한 실제 PreToolUse 페이로드가 픽스처로 커밋되어 있고, 그 관측 결과가 산출물에 기록되어 있다.
2. 모든 Agent/Task spawn이 `.moai/logs/agent-model-audit.jsonl`에 판정과 함께 한 줄씩 기록된다.
3. 판정 `missing` / `mismatch`에 대해 비차단 권고가 방출된다.
4. `workflow.agent_model_guard.enabled` 기본값이 `false`이고, 그 상태에서 어떤 spawn도 차단되지 않는다.
5. 게이트 활성 상태에서 `mismatch`만 차단되고, 불확실 케이스는 전부 allow 폴스루한다.
6. per-spawn 모델 주입 의무가 항상 로드되는 표면에서 읽힌다.
7. `internal/hook` 커버리지가 90% 이상이고, 훅 코드에 `AskUserQuestion` 참조가 0건이다.

---

## §E 참조

- `internal/cli/model.go` — `moai model profile --json` 해석기 표면
- `internal/template/profile_matrix.go` — `ResolveAgentModelEffort` / `ProfileMatrixAgents` / 33셀 SSOT
- `internal/hook/types.go` — `HookInput` 프로토콜 구조체
- `internal/hook/pre_tool.go` — PreToolUse 핸들러 + 브랜치 가드 호출 지점
- `internal/hook/branch_guard.go` — fail-open + 감사 로그 + 센티널 접두사 선례
- `internal/hook/post_tool.go` — 기존 Agent/Task 분기(도달 불가, §A.1 F4)
- `internal/config/types.go` / `defaults.go` — `BranchGuardConfig` opt-in 게이트 선례
- `internal/template/templates/.claude/settings.json.tmpl` — PreToolUse matcher 배선
- `.claude/rules/moai/development/model-policy.md` — 채널 표(effort 파라미터 부재) + `paths:` 스코프
- `internal/verify/subagent_boundary_test.go` — 훅 AskUserQuestion 금지 가드
- `CLAUDE.local.md` §2(Template-First) / §6(커버리지) / §7(훅 정책) / §25(템플릿 중립성)
