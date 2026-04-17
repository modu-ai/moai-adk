---
id: SPEC-OPUS47-COMPAT-001
version: 0.1.1
status: draft
created: 2026-04-17
updated: 2026-04-17
author: GOOS행님
priority: critical
issue_number: 671
---

# Research: SPEC-OPUS47-COMPAT-001

## 개요

본 research 문서는 Plan 단계에서 수행된 전수 감사 결과를 보존한다. Anthropic 공식 문서, moai-adk-go 현재 상태 검증, 충돌 해소 논리를 단일 출처로 모아 Run 단계 실행 시 참조하기 위함이다.

## A. Opus 4.7 프롬프트 철학 5원칙 (Anthropic 공식 인용)

### 원칙 1 — effort 기본값 xhigh 유지

- **공식 출처**: `code.claude.com/docs/en/model-config`
- **인용**: "On Opus 4.7, the default effort is **xhigh** for all plans and providers."
- **의미**: Opus 4.7은 별도 지정 없이도 xhigh 수준 추론을 기본 제공. moai-adk-go는 xhigh를 명시적으로 표기해 의도를 고정하고, 다른 모델(Sonnet/Haiku) 사용 시 자동 폴백(`high`)되도록 설계.

### 원칙 2 — 1턴 몰빵 (Literal Instruction Following)

- **공식 출처**: `platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7`
- **인용**: "More literal instruction following. The model will not silently generalize... will not infer requests you didn't make." / "If you want a hard limit, say so."
- **의미**: 핑퐁 없이 1턴에 완전 정보 세트를 전달해야 함. "double-check X before returning" 류 스캐폴딩은 Opus 4.7에서 불필요하며 때로 혼선 유발.

### 원칙 3 — Extended Thinking 고정 예산 삭제

- **공식 출처**: `platform.claude.com/docs/en/build-with-claude/adaptive-thinking`
- **인용**: "Extended thinking budgets are **removed** in Claude Opus 4.7. Setting `thinking.budget_tokens` returns a **400 error**. Adaptive thinking is the **only** supported thinking mode."
- **의미**: `thinking.budget_tokens` 지정은 Opus 4.7에서 400 에러를 반환. `moai-workflow-thinking/SKILL.md`에서 "thinking budget N tokens" 류 고정 수치 지시를 제거해야 함.

### 원칙 4 — 서브에이전트 자동 스폰 감소

- **공식 출처**: Anthropic 공식 (Opus 4.7 release notes)
- **인용**: "Fewer subagents spawned by default. Steerable through prompting."
- **의미**: Opus 4.7은 기본적으로 서브에이전트를 덜 스폰함. MoAI가 orchestrator로서 명시적 팬아웃(`Agent()` 호출)을 제공해야 함 → `moai-constitution.md`에 "명시적 팬아웃" 원칙 추가 필요.

### 원칙 5 — 툴 호출 감소, 추론 증가

- **공식 출처**: Anthropic 공식
- **인용**: "Fewer tool calls by default, using reasoning more. Raising effort increases tool usage."
- **의미**: Opus 4.7은 툴 호출을 줄이고 내부 추론을 더 사용. `effort: xhigh`는 필요 시 툴 호출 재증가 레버. `agent-common-protocol.md`에 "툴 호출 전 추론" 가이드 추가.

### 결정적 권고 (Anthropic)

- **인용**: "If existing prompts have mitigations in these areas (e.g. 'double-check X before returning'), try removing that scaffolding and re-baselining."
- **의미**: 기존 Opus 4.6 시대 스캐폴딩은 **제거**하고 재기준선 수립이 권장됨. REQ-OC-005의 근거.

## B. Claude Code v2.1.110 High-impact 4건

### B-1. `/doctor` MCP scope 중복 경고

- **현상**: Claude Code v2.1.110부터 `/doctor` 명령이 `.mcp.json`, `settings.json`, user config 3 원천에서 MCP 서버 중복 정의를 탐지해 경고.
- **moai-adk-go 현 상태**: `internal/cli/doctor.go`에 MCP 중복 감지 로직 부재.
- **대응**: `runDiagnosticChecks` 내부에 `checkMCPScopeDuplicates` 추가 (REQ-OC-003).

### B-2. Bash tool 최대 timeout 강제 (600,000ms)

- **현상**: Claude Code v2.1.110은 Bash tool timeout 최대값을 600,000ms(10분)으로 강제.
- **moai-adk-go 현 상태**: `settings.json.tmpl` hooks timeout은 개별 설정(기본 5s). Bash tool 자체 timeout 상한 검증은 없음.
- **대응**: `settings.json` 주석으로 상한 명시, `agent-common-protocol.md`에 문서화 (REQ-OC-003).

### B-3. PermissionRequest hook `updatedInput` deny 재검증

- **현상**: Claude Code v2.1.110은 hook이 `updatedInput`을 반환할 때 deny 패턴과 재대조 수행. 이전에는 updatedInput이 deny 필터를 우회 가능.
- **moai-adk-go 현 상태**: `internal/hook/permission_request.go`에 재검증 로직 없음.
- **대응**: updatedInput 반환 시 기존 deny 패턴 매칭 함수 재호출 (REQ-OC-003).

### B-4. `setMode:'bypassPermissions'` + `disableBypassPermissionsMode`

- **현상**: Claude Code v2.1.110은 settings.json에 `disableBypassPermissionsMode: true` 설정 시 `setMode('bypassPermissions')` 호출을 차단. 엔터프라이즈 보안 요구 대응.
- **moai-adk-go 현 상태**: `settings.json.tmpl`에 해당 키 없음. 기본 동작은 `false`(허용)로 간주.
- **대응**: `settings.json.tmpl`에 명시적으로 `disableBypassPermissionsMode: false` 추가, 문서화 (REQ-OC-003).

## C. Claude Code v2.1.111 High-impact 2건

### C-1. Opus 4.7(`claude-opus-4-7`) + effort 5단계

- **현상**: v2.1.111부터 `claude-opus-4-7` 모델 ID 지원, effort 필드가 `low`/`medium`/`high`/`xhigh`/`max` 5단계로 확장.
- **moai-adk-go 현 상태**:
  - `internal/profile/preferences.go:26`의 `Model` 주석이 `claude-opus-4-6`
  - `internal/template/model_policy.go:16-86`의 `ModelPolicy`가 3단계(`high`/`medium`/`low`)
  - `internal/cli/profile_setup_translations.go:102-262`에 Opus 4.7 선택지 번역 없음
  - `internal/cli/launcher.go:489`에 4.7 모델 ID 라우팅 분기 없음
- **대응**: 위 4개 파일 모두 수정 (REQ-OC-001).

### C-2. Windows `CLAUDE_ENV_FILE` 정상화

- **현상**: v2.1.111부터 Windows 환경에서 `CLAUDE_ENV_FILE` 환경변수가 정상 인식됨. 이전 버전에서는 Windows에서 무시됨.
- **moai-adk-go 현 상태**:
  - `internal/config/envkeys.go:70-72`에 `CLAUDE_ENV_FILE` 상수는 정의됨
  - 하지만 `internal/hook/session_start.go`에 Windows 분기로 해당 상수를 사용하는 코드 없음
  - 참고: 플랫폼 분기 reference 패턴은 `internal/cli/glm.go:489` `injectGLMEnvForTeam`(정의부) 및 `internal/hook/glm_tmux.go:73` `ensureTmuxGLMEnv`(정의부). `session_start.go:72-95`는 이들의 호출부임.
- **대응**: `session_start.go`에 Windows 분기 추가, `CLAUDE_ENV_FILE` 주입 (REQ-OC-004).

## D. 현 프로젝트 상태 검증 결과

### D-1. `.moai/config/sections/llm.yaml` (실측)

```yaml
llm:
    mode: ""
    team_mode: ""
    glm_env_var: GLM_API_KEY
    performance_tier: ""
    claude_models:
        high: ""       # 빈 문자열 — Opus 4.7 설정 전무
        medium: ""
        low: ""
    glm:
        base_url: https://api.z.ai/api/anthropic
        models:
            high: glm-5.1
            medium: glm-4.7
            low: glm-4.5-air
    default_model: sonnet
    quality_model: opus
    speed_model: haiku
```

**분석**: `claude_models.high/medium/low`가 모두 빈 문자열. Opus 4.7 모델 ID 및 effort 필드 전무. P0 작업으로 `high: claude-opus-4-7` 지정 필요.

### D-2. `.moai/config/sections/quality.yaml` (실측)

```yaml
constitution:
    development_mode: tdd
    enforce_quality: true
    test_coverage_target: 85
    ...
    lsp_quality_gates:
        enabled: true
        run:
            max_errors: 0
```

**분석**: LSP quality gate 엄격(run: max_errors: 0). effort 정책 없음. 본 SPEC이 `session_effort_default: xhigh` 추가 대상.

### D-3. `internal/template/model_policy.go` (실측 line 16-86)

3단계 `ModelPolicy` 구조 확인:
- `ModelPolicyHigh` / `ModelPolicyMedium` / `ModelPolicyLow`
- `agentModelMap map[string][3]string` — 각 agent에 대해 `[high, medium, low]` 매핑
- `GetAgentModel(policy, agentName)` 함수 제공

**분석**: 5단계 effort로 확장 시 `[3]string → [5]string` 변경 또는 effort 별도 매핑 도입(Option B 권장, plan.md 참조).

### D-4. `internal/profile/preferences.go` (실측 line 14-44)

```go
type ProfilePreferences struct {
    ...
    ModelPolicy string `yaml:"model_policy,omitempty"` // "high", "medium", "low"
    Model       string `yaml:"model,omitempty"`        // e.g. "claude-opus-4-6"
    ...
}
```

**분석**: `Model` 필드 주석이 `claude-opus-4-6`. Opus 4.7 업데이트 대상. `EffortLevel` 필드 신규 추가 필요.

### D-5. `.claude/rules/moai/development/skill-authoring.md`

- **현 문구 (오류 상태)**: `effort: Session effort override: low, medium, high, max (max is Opus 4.6 only)`
- **수정 필요**: `xhigh` 추가 및 `max is Opus 4.6 only` 문구 삭제 또는 재정의. Opus 4.7에서도 `max`는 지원됨.

### D-6. 28개 agent frontmatter 전수 조사

- 결과: **28개 agent 모두 frontmatter에 `effort` 필드 미설정**
- v2.1.111 신규 필드이므로 정상. 다만 즉시 할당 필요.
- P0에서 6개 reasoning agent 우선 설정, 나머지 22개는 기본값(`high`) 의존.

#### Agent Inventory (28개 검증, 실측 Glob 결과)

실측 경로: `internal/template/templates/.claude/agents/**/*.md`
실측 명령: `Glob("internal/template/templates/.claude/agents/**/*.md")` (iteration 2 검증)

**Moai 카테고리 (22개, `internal/template/templates/.claude/agents/moai/`)**:

1. `manager-spec.md` (P0 reasoning 대상)
2. `manager-strategy.md` (P0 reasoning 대상)
3. `manager-ddd.md`
4. `manager-tdd.md`
5. `manager-docs.md`
6. `manager-git.md`
7. `manager-quality.md`
8. `manager-project.md`
9. `expert-backend.md`
10. `expert-frontend.md`
11. `expert-security.md` (P0 reasoning 대상)
12. `expert-devops.md`
13. `expert-testing.md`
14. `expert-debug.md`
15. `expert-performance.md`
16. `expert-refactoring.md` (P0 reasoning 대상)
17. `builder-agent.md`
18. `builder-skill.md`
19. `builder-plugin.md`
20. `plan-auditor.md` (P0 reasoning 대상)
21. `evaluator-active.md` (P0 reasoning 대상)
22. `researcher.md`

**Agency 카테고리 (6개, `internal/template/templates/.claude/agents/agency/`)**:

23. `planner.md`
24. `copywriter.md`
25. `designer.md`
26. `builder.md`
27. `evaluator.md`
28. `learner.md`

**합계**: 22 + 6 = **28개** 확정.

**P0 대상(6개 reasoning agent)** 확인: manager-spec, manager-strategy, plan-auditor, evaluator-active, expert-security, expert-refactoring. 나머지 22개는 P0 외(기본값 `high` 의존).

**참고 (CLAUDE.md Section 4 카탈로그)**: "Manager 8 + Expert 8 + Builder 3 + Evaluator 2 + Agency 6" = 27개. `researcher.md`는 2026-04 이전에 추가되었으나 CLAUDE.md Section 4 카탈로그 업데이트가 누락된 상태(별도 정비 필요). 본 SPEC은 실측 28개를 기준으로 함.

## E. 전수 패치 파일 리스트 (27개)

### P0 — 15 files (Opus 4.7 호환성 + 프롬프트 철학)

1. `internal/profile/preferences.go`
2. `internal/template/model_policy.go`
3. `internal/cli/profile_setup_translations.go`
4. `internal/cli/launcher.go:489`
5. `internal/template/templates/.moai/config/sections/quality.yaml`
6. `internal/template/templates/.moai/config/sections/llm.yaml`
7. `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`
8. `internal/template/templates/.claude/rules/moai/development/skill-authoring.md`
9. `internal/template/templates/.claude/agents/moai/manager-spec.md`
10. `internal/template/templates/.claude/agents/moai/manager-strategy.md`
11. `internal/template/templates/.claude/agents/moai/plan-auditor.md`
12. `internal/template/templates/.claude/agents/moai/evaluator-active.md`
13. `internal/template/templates/.claude/agents/moai/expert-security.md`
14. `internal/template/templates/.claude/agents/moai/expert-refactoring.md`
15. `internal/template/templates/.claude/skills/moai-workflow-thinking/SKILL.md`

### P1 — 8 files (v2.1.110 High 대응)

16. `internal/cli/doctor.go`
17. `internal/hook/permission_request.go`
18. `internal/hook/session_start.go`
19. `internal/template/templates/.claude/settings.json`
20. `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
21. `internal/template/templates/.claude/rules/moai/core/moai-constitution.md`
22. `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`
23. `internal/template/templates/.moai/config/sections/harness.yaml`

### P2 — 4 files (문서 레이어)

24. `internal/template/templates/.claude/rules/moai/development/coding-standards.md`
25. `CLAUDE.local.md`
26. `internal/template/templates/CLAUDE.md` (§12)
27. `CHANGELOG.md`

## F. 중요 상충 해소

### F-1. `--deepthink` (Sequential Thinking MCP) vs Opus 4.7 Adaptive Thinking

- **상황**: Opus 4.7은 `thinking.budget_tokens` 설정 시 400 error 반환. 반면 MoAI의 `--deepthink` 플래그는 Sequential Thinking MCP를 통해 동작.
- **기술 검증**:
  - Sequential Thinking MCP는 `server_tool_use` content type으로 동작 (별도 도구 호출 경로)
  - Anthropic의 `thinking.budget_tokens`는 Extended Thinking API 직접 파라미터
  - **두 메커니즘은 독립적**이며 서로 영향 없음
- **결론**: `--deepthink`는 Opus 4.7에서도 정상 동작. Sequential Thinking MCP 관련 지시는 유지.
- **주의**: 단, agent 본문에 "thinking budget N tokens" 류 고정 예산 지시가 있다면 Opus 4.7 Extended Thinking 직접 호출 시 무시되므로 제거 대상(REQ-OC-005).

### F-2. `effort` 5단계 vs 기존 `ModelPolicy` 3단계

- **상황**: Opus 4.7의 effort(`low`/`medium`/`high`/`xhigh`/`max`)와 moai-adk-go의 ModelPolicy(`high`/`medium`/`low`)는 층위가 다름.
- **설계 결정 (plan.md Option B)**:
  - `ModelPolicy`: 사용자 요금제(Plus $20 / Max $100 / Max $200) 기반 모델 선택 정책
  - `Effort`: Opus 4.7 내부 reasoning 강도 조절 (Opus 4.7 전용, 타 모델은 무시)
  - 두 개념 분리 유지, `GetAgentEffort(agentName)` 신규 함수 도입

## G. MX Tag Plan (Phase 3.5)

본 SPEC 실행 시 추가할 @MX 태그 계획:

### @MX:ANCHOR (fan_in >= 3 또는 public API boundary)

- `internal/profile/preferences.go`의 `ProfilePreferences` 구조체 — 여러 모듈이 참조
- `internal/template/model_policy.go`의 `ModelPolicy` 상수군 및 `agentModelMap` — 템플릿 배포 로직의 핵심
- `internal/template/model_policy.go`의 `GetAgentModel` 함수 — 외부 호출 다수

### @MX:WARN (위험 영역)

- `internal/cli/doctor.go` MCP scope 검증 함수 — 파일 시스템 읽기 + JSON 파싱, 실패 시 사일런트 fallback 위험
- `@MX:REASON` 필수: "MCP 중복 감지 실패 시 false-negative가 보안 경고 누락으로 이어질 수 있음"

### @MX:NOTE (맥락 전달)

- `internal/hook/session_start.go` Windows 분기 — 플랫폼별 동작 차이 설명
- `internal/profile/preferences.go` `EffortLevel` 필드 — Opus 4.7 전용이며 타 모델에서 무시됨을 주석

### @MX:TODO (미완성)

- P2 문서 항목(CHANGELOG.md, CLAUDE.md §12 등)은 Run phase에서 GREEN 완료 후 해결 예정

## H. Exclusions 근거

### H-1. Vertex AI / AWS Bedrock 미지원

- moai-adk-go의 현 설계는 Anthropic 공식 API + GLM(z.ai) 두 provider만 지원
- Enterprise provider 요청이 사용자로부터 접수된 바 없음
- Vertex/Bedrock 지원은 별도 SPEC으로 분리해야 할 범위

### H-2. `/tui`, `/focus`, Push notification 등

- Claude Code 런타임 레벨 UX 기능
- moai-adk-go는 CLI orchestrator이므로 이러한 런타임 기능 중복 구현 불필요

### H-3. `plugin_errors` stream-json

- MoAI는 Claude Code plugin을 배포하지 않음
- plugin 관련 응답 형식 처리는 범위 외

### H-4. TRACEPARENT/TRACESTATE 분산 트레이싱

- moai-adk-go는 headless 모드 미지원 결정 이미 존재
- 분산 트레이싱은 해당 결정과 상충

### H-5. CLAUDE_CODE_ENABLE_AWAY_SUMMARY opt-out

- Claude Code 런타임 기능
- moai-adk-go 개입 불가

### H-6. `/ultrareview` vs `/moai review` 관계 정리

- 사용자가 명시적으로 "3회 수동 제한으로 관계 정리 안 함" 지시
- 본 SPEC 범위 외

### H-7. 기존 Opus 4.6/Sonnet 4.6/Haiku 4.5 제거

- NFR-1(하위 호환성) 필수
- 제거 시 기존 사용자 마이그레이션 비용 과다

### H-8. 실제 코드 구현

- 본 SPEC은 Plan 산출물
- 실제 수정은 `/moai run SPEC-OPUS47-COMPAT-001` 에서 수행

### H-9. Sequential Thinking MCP 비활성화

- F-1 검증으로 Opus 4.7 Adaptive Thinking과 무관함 확인
- 비활성화 시 기존 `--deepthink` 워크플로우 전체 손실

## I. Reference Implementations (코드 참조, 실측 검증 완료)

- **ModelPolicy 3단계 패턴**: `internal/template/model_policy.go:16-86`
  - 5단계로 확장 시 `agentModelMap[string][3]string → map[string][5]string` 또는 `map[string]AgentModelEntry` 구조체 도입
- **SessionStart hook 호출 흐름**: `internal/hook/session_start.go:72-95`
  - 72-95 라인은 함수 **호출부**이며 정의부가 아님에 유의
  - 해당 라인에서 호출되는 함수 및 정의 위치:
    - `ensureGLMCredentials` → 정의부 `internal/hook/session_start.go:157`
    - `ensureTmuxGLMEnv` → 정의부 `internal/hook/glm_tmux.go:73` (macOS/Linux tmux 전용)
    - `ensureTeammateMode` → 정의부 `internal/hook/session_start.go:274`
- **플랫폼 분기/환경 주입 모범 사례**: `internal/cli/glm.go:489`의 `injectGLMEnvForTeam` 정의부
  - GLM 모드에서 settings.local.json을 수정하여 환경변수를 주입하는 패턴
  - Windows `CLAUDE_ENV_FILE` 주입 구현 시 동일 스타일(함수 정의 + runtime.GOOS 가드)로 신규 함수 도입 권장
- **Doctor 진단 함수 구조**: `internal/cli/doctor.go:118`의 `runDiagnosticChecks`
  - 기존 체크 함수들(예: `checkGoVersion`, `checkGitInstalled`)과 동일 패턴으로 `checkMCPScopeDuplicates` 추가

## J. 공식 외부 참조 URL

- **Opus 4.7 Release Notes**: `https://platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7`
- **Model Config (effort)**: `https://code.claude.com/docs/en/model-config`
- **Adaptive Thinking**: `https://platform.claude.com/docs/en/build-with-claude/adaptive-thinking`
- **Claude Code Changelog**: `https://code.claude.com/docs/en/changelog` (v2.1.110, v2.1.111)

주의: 본 URL들은 Context Dump 시점 검증된 출처. Run phase 실행 시 WebFetch로 재확인 권장.

## K. Risk 요약

| Risk ID | Phase | Severity | Mitigation |
|---------|-------|----------|------------|
| R-P0-1  | P0    | Medium   | `agentModelMap` 확장 시 Go 컴파일러가 누락 차단 + CI로 검증 |
| R-P0-2  | P0    | Low      | Agent 본문 리라이트 시 워크플로우 스텝 유지, 스캐폴딩만 제거 (feedback_agent_refactor_constraint 준수) |
| R-P0-3  | P0    | Medium   | `make build` 후 `embedded.go` git diff 체크를 CI에 추가 |
| R-P1-1  | P1    | High     | `session_start.go` 기존 함수 시그니처 불변, Windows 분기는 별도 함수 + 회귀 테스트 |
| R-P1-2  | P1    | Medium   | `settings.json.tmpl` 변경 블록을 SPEC-GLM-001과 분리 |
| R-P1-3  | P1    | Low      | MCP 중복은 warning only, exit code 0 유지 |
| R-P2-1  | P2    | Low      | CHANGELOG 버전 bump는 v2.12.0 tag 직전 최종 확정 |
