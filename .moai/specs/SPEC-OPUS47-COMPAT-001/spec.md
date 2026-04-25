---
id: SPEC-OPUS47-COMPAT-001
version: 0.2.0
status: completed
created: 2026-04-17
updated: 2026-04-24
author: GOOS행님
priority: critical
issue_number: 671
merged_pr: [672, 673]
merged_commits: [07525c7ae, 4a00aa304]
---

# SPEC-OPUS47-COMPAT-001: Claude Code v2.1.110/111 + Opus 4.7 프롬프트 철학 적용

## HISTORY

| Version | Date       | Author    | Change                                                                                                                                                   |
| ------- | ---------- | --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 0.1.0   | 2026-04-17 | GOOS행님 | 초기 SPEC 작성 — Opus 4.7 호환성 및 프롬프트 철학 5원칙 시스템 적용                                                                                      |
| 0.1.1   | 2026-04-17 | GOOS행님 | plan-auditor iteration 1 반영: Reference 경로 정정(injectGLMEnvForTeam, ensureTmuxGLMEnv), REQ-OC-003(c) 문서화로 약화, DELTA 마커 통일, EARS 라벨 정정 |
| 0.2.0   | 2026-04-24 | plan-auditor | 후행 감사 반영: status draft → completed (PR #672, #673 merge 반영, commits 07525c7ae & 4a00aa304). D-OPUS47-2 (builder-agent.md `effort` 필드 scope 미정합) 인지됨 — REQ-OC-002 enumeration에 builder-agent 포함 의도 확인 필요. |

## Context (Background)

Anthropic이 2026년 4월 Opus 4.7(`claude-opus-4-7`)을 출시하며 **5원칙 프롬프트 철학**을 공식화했다. 동시에 Claude Code v2.1.110/111은 runtime 레벨에서 4건의 High-impact 변경(MCP scope 중복 경고, Bash timeout 상한 강제, PermissionRequest updatedInput deny 재검증, bypassPermissions 정책)과 프로파일 레벨에서 2건의 변경(Opus 4.7 + effort 5단계, Windows CLAUDE_ENV_FILE 정상화)을 도입했다.

moai-adk-go는 현재:
- `internal/template/model_policy.go`가 3단계(`high`/`medium`/`low`) 매핑만 지원 → Opus 4.7의 `effort` 5단계(`low`/`medium`/`high`/`xhigh`/`max`) 미반영
- `internal/profile/preferences.go`의 `Model` 필드 주석이 `claude-opus-4-6`에 고정
- `internal/cli/launcher.go`, `internal/cli/profile_setup_translations.go`가 4.7 모델 ID 미라우팅
- 28개 agent 전원 frontmatter에 `effort` 필드 미설정
- `.moai/config/sections/llm.yaml`의 `claude_models.{high,medium,low}`가 모두 빈 문자열
- `skill-authoring.md`가 `max is Opus 4.6 only`로 명기(4.7 대비 오류 상태)
- `internal/hook/session_start.go`에 Windows CLAUDE_ENV_FILE 주입 분기 없음
- `internal/cli/doctor.go`가 MCP scope 중복 감지 미지원
- `internal/hook/permission_request.go`가 updatedInput deny 재검증 미구현

본 SPEC은 상기 격차를 P0/P1/P2 3단계로 시스템적으로 해소한다. Opus 4.6/Sonnet 4.6/Haiku 4.5는 **하위 호환 유지**하며, `effort` 5단계는 Opus 4.7에서만 활성(타 모델은 `high`로 자동 폴백).

## Goals

- [GOAL-1] Opus 4.7(`claude-opus-4-7`) + `effort` 5단계를 프로파일/템플릿/에이전트 전 레이어에 반영한다
- [GOAL-2] Opus 4.7 프롬프트 철학 5원칙(xhigh 기본, 1턴 몰빵, Adaptive Thinking 전용, 서브에이전트↓, 툴 호출↓)을 에이전트 본문과 rules에 명문화한다
- [GOAL-3] Claude Code v2.1.110 High-impact 4건과 v2.1.111 프로파일 2건을 runtime/hook 레이어에 적용한다
- [GOAL-4] 기존 Opus 4.6/Sonnet 4.6/Haiku 4.5 사용자에게 영향 없이 점진 도입이 가능한 구조로 설계한다

## Requirements (EARS Format)

### REQ-OC-001 — 모델 카탈로그 및 프로파일 갱신 [Ubiquitous + MODIFY]

The system **shall** support `claude-opus-4-7` as a first-class model in profile preferences, launcher model routing, and template model policy with full 5-level `effort` field support (`low`/`medium`/`high`/`xhigh`/`max`).

**DELTA markers:**
- [MODIFY] `internal/profile/preferences.go` — `Model` 필드 주석을 `claude-opus-4-7`로 갱신, `EffortLevel` 필드 신규 추가
- [MODIFY] `internal/template/model_policy.go` — `ModelPolicy` 상수에 `xhigh`/`max` 추가, `agentModelMap`을 `[5]string`으로 확장 또는 별도 effort 매핑 도입
- [MODIFY] `internal/cli/profile_setup_translations.go:102-262` — Opus 4.7 + xhigh 선택지 번역 엔트리 추가(한국어/영문)
- [MODIFY] `internal/cli/launcher.go:489` — `claude-opus-4-7` 모델 ID 라우팅 분기 추가

**Acceptance scope:** 사용자가 `moai profile` 대화창에서 Opus 4.7 선택 후 `effort: xhigh` 지정 시, `~/.moai/claude-profiles/<name>/preferences.yaml`에 `model: claude-opus-4-7`와 `effort_level: xhigh`가 기록되며, `moai cc` 실행 시 해당 값이 Claude Code에 전달된다.

### REQ-OC-002 — Opus 4.7 프롬프트 철학 5원칙 명문화 [Event-Driven + MODIFY]

**When** the system deploys templates via `moai init` or `moai update`, the system **shall** apply Opus 4.7 prompt philosophy principles to all critical reasoning agents (`manager-spec`, `manager-strategy`, `plan-auditor`, `evaluator-active`, `expert-security`, `expert-refactoring`) by setting `effort: xhigh` (or `high`) in frontmatter and rewriting prompt bodies to follow the "one-turn fully-loaded" principle.

**DELTA markers:**
- [MODIFY] `internal/template/templates/.claude/agents/moai/manager-spec.md` — `effort: xhigh`, 1턴 몰빵 리라이트
- [MODIFY] `internal/template/templates/.claude/agents/moai/manager-strategy.md` — `effort: xhigh`, 팬아웃 명시
- [MODIFY] `internal/template/templates/.claude/agents/moai/plan-auditor.md` — `effort: high`
- [MODIFY] `internal/template/templates/.claude/agents/moai/evaluator-active.md` — `effort: high`
- [MODIFY] `internal/template/templates/.claude/agents/moai/expert-security.md` — `effort: high`
- [MODIFY] `internal/template/templates/.claude/agents/moai/expert-refactoring.md` — `effort: high`
- [MODIFY] `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` — `effort` 5단계 스펙 + Opus 4.7 주석
- [MODIFY] `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` — `max is Opus 4.6 only` 문구 수정, `xhigh` 추가
- [MODIFY] `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` — Opus 4.7 원칙 4(팬아웃 명시), 원칙 5(툴 호출 가이드) 명문화
- [MODIFY] `internal/template/templates/.claude/skills/moai-workflow-thinking/SKILL.md` — Adaptive Thinking 원칙 재정의, 고정 예산 지시("thinking budget N tokens" 류) 제거

**Acceptance scope:** `make build` 후 내장 템플릿(`internal/template/embedded.go`)에서 상기 6개 에이전트의 frontmatter `effort` 필드가 존재하며, 본문에서 "double-check X before returning" 류 스캐폴딩이 제거된다. `skill-authoring.md`에서 `max is Opus 4.6 only` 문자열이 더 이상 존재하지 않는다.

### REQ-OC-003 — v2.1.110 Runtime 대응 [State-Driven + MODIFY]

**While** a MoAI project is running under Claude Code v2.1.110 or later, the system **shall** (a) detect and warn on MCP scope duplicates across `.mcp.json` + `settings.json` + user config in `moai doctor`, (b) re-validate `updatedInput` in PermissionRequest hook to prevent deny-pattern bypass, (c) document the Claude Code Bash tool timeout ceiling (600,000ms) in `agent-authoring.md` and `agent-common-protocol.md` so that agents do not specify exceeding values (enforcement itself is handled by Claude Code runtime, not moai-adk-go), and (d) respect `disableBypassPermissionsMode` policy in settings.json for `setMode:'bypassPermissions'` gating.

**DELTA markers:**
- [MODIFY] `internal/cli/doctor.go` — `runDiagnosticChecks` 내부에 MCP scope 중복 감지 로직 추가 (reference: `internal/cli/doctor.go` 기존 진단 함수 패턴)
- [MODIFY] `internal/hook/permission_request.go` — `updatedInput` deny 재검증 로직 추가
- [MODIFY] `internal/template/templates/.claude/settings.json` — `disableBypassPermissionsMode` 정책 필드 추가, Bash timeout 상한(600000ms) 문서화 주석
- [MODIFY] `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` — Bash timeout 600,000ms 상한을 문서화(Claude Code 런타임 강제 사항이므로 moai는 agent 저작 시 참조용 문서만 유지), 팬아웃 원칙 추가
- [MODIFY] `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` — Bash timeout 600,000ms 상한 문서화(동일 근거)
- [MODIFY] `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` — v2.1.110 최소 버전 요구사항(MCP doctor) 추가

**Acceptance scope:** `moai doctor` 실행 시 중복 MCP 서버 정의(`.mcp.json`과 `settings.json`이 같은 이름을 다른 스코프로 등록)가 경고로 리포트된다. `settings.json.tmpl` 렌더링 결과에서 `disableBypassPermissionsMode: false`(기본값) 키가 존재한다. `agent-authoring.md`와 `agent-common-protocol.md`에 Bash tool timeout 상한(600,000ms) 문서가 추가되며, 본 한계는 Claude Code 런타임이 강제하는 사항이므로 moai-adk-go 수준에서는 문서화만 수행한다.

### REQ-OC-004 — Windows CLAUDE_ENV_FILE 정상화 [Optional + MODIFY]

**Where** the host operating system is Windows, the system **shall** inject `CLAUDE_ENV_FILE` environment variable via `SessionStart` hook so that `internal/config/envkeys.go:70-72` defined constants are actually used by Claude Code sessions.

**DELTA markers:**
- [MODIFY] `internal/hook/session_start.go` — Windows 분기 추가, `CLAUDE_ENV_FILE` 환경변수 주입 (reference patterns: `internal/cli/glm.go:489` `injectGLMEnvForTeam`와 `internal/hook/glm_tmux.go:73` `ensureTmuxGLMEnv` 정의부가 플랫폼 분기/환경 주입 모범 사례. `session_start.go:72-95`는 해당 함수들의 호출부임)

**Context (no DELTA — unchanged):**
- `internal/config/envkeys.go:70-72` — `CLAUDE_ENV_FILE` 상수는 이미 정의되어 있으므로 변경 불필요(Reference only)

**Acceptance scope:** Windows 환경(또는 `GOOS=windows` 빌드)에서 `moai hook session-start` 실행 시 `CLAUDE_ENV_FILE` 값이 프로세스 환경변수로 설정된다. macOS/Linux에서는 기존 동작(tmux GLM env 주입)이 영향받지 않는다.

### REQ-OC-005 — 프롬프트 스캐폴딩 제거 원칙 [Unwanted + MODIFY]

**If** an agent prompt body contains Opus 4.6 era mitigation scaffolding (e.g., "double-check X before returning", "verify N times", "explicitly confirm before proceeding" repeated patterns), **then** the system **shall not** preserve these instructions in Opus 4.7-targeted agents (those with `effort: xhigh` or `effort: high`); instead the system **shall** delete the scaffolding and re-baseline the agent's behavior based on Opus 4.7's improved literal instruction following.

**DELTA markers:**
- [MODIFY] Scope covers the 6 agents listed in REQ-OC-002 plus any agent whose `effort` is set to `xhigh`/`high` during REQ-OC-002 deployment
- [MODIFY] `internal/template/templates/.claude/skills/moai-workflow-thinking/SKILL.md` — "thinking budget N tokens" 류 고정 예산 지시 제거 (Adaptive Thinking 원칙과 충돌)

**Acceptance scope:** `make build` 이후 대상 6개 에이전트 본문에서 "double-check"/"verify N times"/"explicitly confirm" 패턴이 **0건**(하드 기준)으로 제거된다. `moai-workflow-thinking/SKILL.md`에서 "thinking budget" 고정 수치 지시가 제거된다.

## Non-Functional Requirements (NFR)

### NFR-1 — 하위 호환성 (Backward Compatibility)

- Opus 4.6/Sonnet 4.6/Haiku 4.5 모델은 카탈로그에서 제거하지 않는다
- `effort` 필드 미설정 에이전트는 현행 동작 그대로 유지(Opus 4.7 세션에서는 기본값 `xhigh`, 타 모델 세션에서는 무시)
- 기존 `.moai/config/sections/llm.yaml` 사용자 설정(빈 문자열 포함)은 마이그레이션 없이 동작

### NFR-2 — Template-First 준수

- 모든 변경은 `internal/template/templates/` 하위에 선 반영 후 `make build`로 `embedded.go` 재생성
- Local `.claude/` 직접 수정 금지
- `make build` 실행 없이 커밋 금지

### NFR-3 — 언어 중립성

- 새로 추가되는 agent frontmatter의 `effort` 필드는 16개 언어 템플릿 전체에 동일 적용
- 특정 언어(Go 포함)에 편향된 예시를 `internal/template/templates/**` 본문에 삽입 금지

### NFR-4 — LSP Quality Gate 통과

- `internal/` 하위 Go 코드 변경은 `.moai/config/sections/quality.yaml`의 `lsp_quality_gates.run.max_errors: 0` 기준을 만족
- `go vet ./...`, `golangci-lint run`, `go test -race ./...` 모두 통과

## Exclusions (What NOT to Build)

- [EXCLUDE] **Vertex AI / AWS Bedrock 지원 추가**: moai-adk-go는 Claude(Anthropic 공식) + GLM(z.ai) 전용. Enterprise provider 요청 없음
- [EXCLUDE] **`/tui`, `/focus`, Push notification, Remote Control 등 v2.1.110 UX 기능**: Claude Code 런타임 레벨 기능이며 moai-adk-go 범위 외
- [EXCLUDE] **`plugin_errors` stream-json 구조 구현**: moai는 plugin을 배포하지 않으므로 해당 응답 형식 처리 불필요
- [EXCLUDE] **TRACEPARENT/TRACESTATE SDK 분산 트레이싱**: headless 모드 미지원 결정과 상충
- [EXCLUDE] **`CLAUDE_CODE_ENABLE_AWAY_SUMMARY` opt-out 구현**: Claude Code 런타임 기능
- [EXCLUDE] **`/ultrareview`와 `/moai review` 관계 정리**: 사용자가 "3회 수동 제한으로 관계 정리 안 함"으로 명시
- [EXCLUDE] **기존 Opus 4.6 / Sonnet 4.6 / Haiku 4.5 모델 제거**: 하위 호환 유지 필수 (NFR-1)
- [EXCLUDE] **실제 Go 코드·YAML·Markdown 파일의 구현 변경**: 본 SPEC은 Plan 단계 산출물이며, 실제 수정은 `/moai run SPEC-OPUS47-COMPAT-001`에서 수행
- [EXCLUDE] **Sequential Thinking MCP(`--deepthink`) 비활성화**: `server_tool_use` 경로로 동작하므로 Opus 4.7 `thinking.budget_tokens` 400 error와 무관(Context Dump §F 검증 완료)

## Scope Validation

- **요구사항 모듈**: 5개 (REQ-OC-001 ~ REQ-OC-005) — SPEC 규칙 "요구사항 모듈 5개 이하" 준수
- **EARS 패턴 커버리지**: Ubiquitous(REQ-OC-001), Event-Driven(REQ-OC-002), State-Driven(REQ-OC-003), Optional(REQ-OC-004), Unwanted(REQ-OC-005) — **5 유형 전원 사용**
- **Composite domain**: `OPUS47-COMPAT` 단일 (승인된 최대 2 도메인 제약 준수)
- **Exclusions 엔트리**: 9개 (최소 1개 필수 기준 초과 만족)

## Traceability

- **Related SPECs**:
  - `SPEC-GLM-001` (GLM Compatibility, 기존): `DISABLE_BETAS` + `DISABLE_PROMPT_CACHING` 자동화와 본 SPEC의 `disableBypassPermissionsMode`는 독립적이나 settings.json 같은 파일 수정이므로 merge 순서 주의
  - `SPEC-CC297-001` (Claude Code 2.9.7 대응, 기존): v2.1.x 계열 호환성 기반
- **External references**:
  - `platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7` (Opus 4.7 literal instruction following)
  - `code.claude.com/docs/en/model-config` (effort xhigh default)
  - `platform.claude.com/docs/en/build-with-claude/adaptive-thinking` (고정 예산 400 error)
- **Code references** (실측 검증 완료, iteration 2):
  - `internal/template/model_policy.go:16-86` (ModelPolicy 3단계 → 5단계 확장 대상)
  - `internal/profile/preferences.go:26` (Model 필드 Opus 4.6 → 4.7)
  - `internal/cli/launcher.go:489` (모델 ID 라우팅)
  - `internal/hook/session_start.go:72-95` (Windows 분기 추가 위치 — 호출부)
  - `internal/cli/glm.go:489` (`injectGLMEnvForTeam` 정의부, 플랫폼 분기 reference 패턴)
  - `internal/hook/glm_tmux.go:73` (`ensureTmuxGLMEnv` 정의부, macOS/Linux tmux 처리 reference)
  - `internal/hook/session_start.go:157` (`ensureGLMCredentials` 정의부)
  - `internal/hook/session_start.go:274` (`ensureTeammateMode` 정의부)
  - `internal/cli/doctor.go:118` (`runDiagnosticChecks` 정의부, MCP scope 중복 감지 추가 위치)

## References

- Plan: [plan.md](./plan.md)
- Acceptance: [acceptance.md](./acceptance.md)
- Research: [research.md](./research.md)
- Compact: [spec-compact.md](./spec-compact.md)
