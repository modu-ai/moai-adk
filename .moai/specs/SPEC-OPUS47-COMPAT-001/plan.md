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

# Plan: SPEC-OPUS47-COMPAT-001

## Overview

Opus 4.7 호환성 및 프롬프트 철학 적용을 **P0(코어 호환성) → P1(v2.1.110 런타임 대응) → P2(문서 레이어)** 3단계로 분해한다. 각 단계는 독립 commit 가능하며, P0 완료 시점에서 Opus 4.7 사용자는 즉시 이득을 본다. P1/P2는 점진 롤아웃 가능.

## Priority-Based Phases

### Phase P0 — 코어 호환성 + 프롬프트 철학 (Priority: critical, 15 files)

**Scope**: REQ-OC-001, REQ-OC-002, REQ-OC-005

**Objective**: Opus 4.7 모델 카탈로그 등록, `effort` 5단계 지원, 6개 핵심 reasoning agent의 프롬프트 철학 적용.

**File list with [DELTA] markers**:

| # | File | DELTA | Change |
|---|------|-------|--------|
| 1 | `internal/profile/preferences.go` | [MODIFY] | `Model` 주석 갱신 (`claude-opus-4-7`), `EffortLevel string` 필드 신규 추가 (line 26 근처) |
| 2 | `internal/template/model_policy.go` | [MODIFY] | `ModelPolicy` 상수에 `xhigh`, `max` 추가; `agentModelMap`을 5단계 또는 effort 별도 매핑 도입 (line 16-86) |
| 3 | `internal/cli/profile_setup_translations.go` | [MODIFY] | Opus 4.7 + xhigh 번역 엔트리 추가 (line 102-262) |
| 4 | `internal/cli/launcher.go` | [MODIFY] | `claude-opus-4-7` 모델 ID 라우팅 분기 (line 489) |
| 5 | `internal/template/templates/.moai/config/sections/quality.yaml` | [MODIFY] | `session_effort_default: xhigh` 필드 추가 |
| 6 | `internal/template/templates/.moai/config/sections/llm.yaml` | [MODIFY] | `claude_models.high: claude-opus-4-7`, effort 매핑 예시 추가 |
| 7 | `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | [MODIFY] | `effort` 5단계 스펙 + Opus 4.7 주석 |
| 8 | `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` | [MODIFY] | `max is Opus 4.6 only` 문구 삭제, `xhigh` 레벨 추가 |
| 9 | `internal/template/templates/.claude/agents/moai/manager-spec.md` | [MODIFY] | `effort: xhigh` frontmatter + 본문 1턴 몰빵 리라이트 + 스캐폴딩 제거 |
| 10 | `internal/template/templates/.claude/agents/moai/manager-strategy.md` | [MODIFY] | `effort: xhigh` + 팬아웃 원칙 명시 + 스캐폴딩 제거 |
| 11 | `internal/template/templates/.claude/agents/moai/plan-auditor.md` | [MODIFY] | `effort: high` |
| 12 | `internal/template/templates/.claude/agents/moai/evaluator-active.md` | [MODIFY] | `effort: high` |
| 13 | `internal/template/templates/.claude/agents/moai/expert-security.md` | [MODIFY] | `effort: high` |
| 14 | `internal/template/templates/.claude/agents/moai/expert-refactoring.md` | [MODIFY] | `effort: high` |
| 15 | `internal/template/templates/.claude/skills/moai-workflow-thinking/SKILL.md` | [MODIFY] | Adaptive Thinking 재정의, "thinking budget N tokens" 고정 수치 제거 |

**Reference implementations**:

- ModelPolicy 3단계 패턴: `internal/template/model_policy.go:16-86` (기존 구조 참조하여 5단계 확장)
- agentModelMap 구조: `internal/template/model_policy.go:44-67` ([5]string 또는 별도 effort 매핑)
- Profile preferences YAML 필드 추가 패턴: `internal/profile/preferences.go:14-44` (기존 필드 주석·yaml 태그 패턴 준용)

**Implementation order**:

1. Go 코드 먼저 (파일 1-4): `profile/preferences.go` → `template/model_policy.go` → `cli/profile_setup_translations.go` → `cli/launcher.go`
2. 테스트: `go test ./internal/profile/... ./internal/template/... ./internal/cli/...`
3. 템플릿 설정 파일 (파일 5-6): `quality.yaml`, `llm.yaml`
4. Rules 문서 (파일 7-8): `agent-authoring.md`, `skill-authoring.md`
5. Agent 본문 (파일 9-14): 6개 reasoning agent 순차 리라이트
6. Skill 본문 (파일 15): `moai-workflow-thinking/SKILL.md`
7. `make build` → `embedded.go` 재생성 검증
8. `go test ./...` 전체 통과 확인

**Technical constraints**:

- `ModelPolicy` 상수 확장 시 기존 `IsValidModelPolicy`, `ValidModelPolicies` 함수 시그니처 유지 (호환성)
- `agentModelMap` 구조 변경 시 `GetAgentModel` 시그니처는 유지하되 내부 구현만 확장
- Agent frontmatter 수정 시 YAML 파싱 충돌 방지 — `effort` 필드는 기존 필드(name, description, tools 등) 이후에 배치
- `skill-authoring.md` 수정 시 기존 다른 스킬 문서에 `max is Opus 4.6 only` 참조가 없는지 Grep으로 확인

**Risks**:

- **R-P0-1** (Medium): `agentModelMap [3]string → [5]string` 확장 시 초기값 설정 누락으로 컴파일 에러. **Mitigation**: Go 컴파일러가 차단하므로 CI에서 자동 감지. 모든 agent entry에 5개 값 명시.
- **R-P0-2** (Low): 6개 agent 본문 리라이트 시 기존 워크플로우 축소 리스크(memory feedback_agent_refactor_constraint 참조). **Mitigation**: 프롬프트 스캐폴딩만 제거, 워크플로우 스텝은 유지.
- **R-P0-3** (Medium): `make build` 누락으로 `embedded.go` 구버전 유지. **Mitigation**: CI에서 `make build` 후 `git diff --exit-code internal/template/embedded.go` 체크.

### Phase P1 — v2.1.110 Runtime 대응 (Priority: high, 8 files)

**Scope**: REQ-OC-003, REQ-OC-004

**Objective**: Claude Code v2.1.110의 4건 High-impact 변경 및 Windows CLAUDE_ENV_FILE 정상화.

**File list with [DELTA] markers**:

| # | File | DELTA | Change |
|---|------|-------|--------|
| 16 | `internal/cli/doctor.go` | [MODIFY] | MCP scope 중복 감지 로직 추가 (runDiagnosticChecks 내부) |
| 17 | `internal/hook/permission_request.go` | [MODIFY] | `updatedInput` deny 재검증 로직 |
| 18 | `internal/hook/session_start.go` | [MODIFY] | Windows `CLAUDE_ENV_FILE` 주입 분기 (기존 macOS/Linux 경로 영향 없음) |
| 19 | `internal/template/templates/.claude/settings.json` | [MODIFY] | `disableBypassPermissionsMode: false` 키 추가, Bash timeout 상한(600,000ms) 문서화 주석 |
| 20 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | [MODIFY] | Bash timeout 600,000ms 상한 문서화(Claude Code 런타임 강제), 팬아웃 원칙 추가 |
| 20a | `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | [MODIFY] | Bash timeout 600,000ms 상한 문서화(agent 저작 시 참조) |
| 21 | `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` | [MODIFY] | Opus 4.7 원칙 4(팬아웃 명시), 원칙 5(툴 호출 가이드) 섹션 추가 |
| 22 | `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | [MODIFY] | v2.1.110 최소 버전(MCP doctor), v2.1.111(CLAUDE_ENV_FILE) 추가 |
| 23 | `internal/template/templates/.moai/config/sections/harness.yaml` | [MODIFY] | level↔effort 매핑(thorough→xhigh), 본 SPEC `model_upgrade_review` 실행 기록 |

**Reference implementations**:

- `doctor.go` 진단 패턴: 기존 `runDiagnosticChecks`(`internal/cli/doctor.go:118`) 함수 구조를 참조하여 `checkMCPScopeDuplicates` 같은 함수 추가
- 플랫폼 분기/환경 주입 패턴: `internal/cli/glm.go:489`의 `injectGLMEnvForTeam` 정의부와 `internal/hook/glm_tmux.go:73`의 `ensureTmuxGLMEnv` 정의부 구조를 준용하여 Windows `CLAUDE_ENV_FILE` 주입을 `session_start.go`에 신규 추가. `session_start.go:72-95`는 해당 함수들의 호출부이며 정의부가 아님에 유의
- `settings.json` 템플릿 필드 추가 패턴: 기존 `permissions` 블록 구조 참조

**Implementation order**:

1. `doctor.go` (단독 실행 명령, 다른 코드와 의존성 낮음)
2. `permission_request.go` (hook 격리)
3. `session_start.go` (Windows 분기 — 기존 경로 회귀 방지 테스트 필수)
4. `settings.json` + rules 3종 (문서 레이어)
5. `harness.yaml` 업데이트
6. `make build` → `embedded.go` 재생성
7. `go test ./internal/cli/... ./internal/hook/...` 통과 확인

**Technical constraints**:

- `doctor.go` MCP 감지는 `.mcp.json`, `settings.json`, `~/.config/claude/settings.json`(user scope) 3 원천을 모두 파싱해야 함 — 파일 부재 시 graceful skip
- `permission_request.go` 재검증 로직은 기존 deny 패턴 매칭 함수를 재사용(중복 구현 금지)
- `session_start.go` Windows 분기는 `runtime.GOOS == "windows"` 가드 내에서만 실행, POSIX 경로에 영향 없음
- `settings.json` 변경은 `settings.json.tmpl` 렌더링 후에도 JSON validity 유지

**Risks**:

- **R-P1-1** (High): `session_start.go` 변경이 기존 tmux GLM env 주입(teammate_mode_regression 이슈 재발 가능)을 깨뜨릴 리스크. **Mitigation**: 기존 `injectGLMEnvForTeam` 함수 시그니처/동작 불변 유지, Windows 분기는 별도 함수로 추가. macOS/Linux 회귀 테스트 필수.
- **R-P1-2** (Medium): `settings.json.tmpl` 수정이 SPEC-GLM-001의 `DISABLE_BETAS` 주입과 merge conflict. **Mitigation**: 같은 블록 동시 수정 금지, 별도 키 블록으로 분리.
- **R-P1-3** (Low): `doctor.go` MCP 중복 감지가 false-positive(같은 서버를 의도적으로 다른 스코프에서 override하는 경우)를 생성. **Mitigation**: warning 레벨로만 리포트, 차단 금지.

### Phase P2 — 문서 레이어 (Priority: medium, 4+ files)

**Scope**: GOAL-2 보강, 사용자 마이그레이션 가이드

**Objective**: 사용자 대면 문서(CLAUDE.local.md, coding-standards.md, CHANGELOG.md)에 v2.1.111/Opus 4.7 변경사항 반영.

**File list with [DELTA] markers**:

| # | File | DELTA | Change |
|---|------|-------|--------|
| 24 | `internal/template/templates/.claude/rules/moai/development/coding-standards.md` | [MODIFY] | v2.1.111 read-only Bash 자동 분류, `/less-permission-prompts` 가이드 |
| 25 | `CLAUDE.local.md` | [MODIFY] | `settings.local.json` 분리 원칙 + `OTEL_LOG_RAW_API_BODIES` 프로덕션 금지 경고 |
| 26 | `internal/template/templates/CLAUDE.md` §12 | [MODIFY] | UltraThink vs Adaptive Thinking 용어 정리(Sequential Thinking MCP와 구분) |
| 27 | `CHANGELOG.md` | [MODIFY] | v2.12.0 항목 추가 ("Claude Code v2.1.110+ 권장, Opus 4.7 프롬프트 철학 적용") |

**Implementation order**:

1. `CLAUDE.md §12` 용어 정리 먼저(다른 문서가 참조)
2. `coding-standards.md` 업데이트
3. `CLAUDE.local.md` (로컬 전용, 템플릿 아님)
4. `CHANGELOG.md` 마지막(P0/P1 완료 후 배포 전)

**Technical constraints**:

- `CLAUDE.md §12` 수정 시 Sequential Thinking MCP 동작(`server_tool_use`)과 Adaptive Thinking(`thinking.budget_tokens`) 차이를 명확히 서술
- `CHANGELOG.md`는 Conventional Commits 스타일 유지, 한국어 작성

**Risks**:

- **R-P2-1** (Low): `CHANGELOG.md` 버전 번호 충돌(다른 SPEC과 merge 순서). **Mitigation**: release/v2.11.0 → v2.12.0 bump 시점에 최종 확정.

## Technical Approach

### Effort 필드 설계

5단계 매핑 전략은 2가지 중 하나로 선택(P0 구현 시 결정):

**Option A**: `agentModelMap`을 `[5]string`으로 확장
- 장점: 기존 구조와 유사
- 단점: 모든 agent에 대해 5개 값 기입 필요 — Sonnet/Haiku는 `xhigh`/`max`가 의미 없으므로 중복 값 다수 발생

**Option B**: `ModelPolicy`(high/medium/low)와 `Effort`(low/medium/high/xhigh/max) 분리 — 권장
- 장점: Opus 4.7만 effort 적용, 타 모델은 `high`로 자동 폴백
- 단점: 두 개념 분리로 초기 복잡도 증가

**결정 근거**: Option B가 NFR-1(하위 호환성) 및 NFR-3(언어 중립성)을 더 잘 만족. `GetAgentModel(policy, agentName)`은 유지하고, 신규 `GetAgentEffort(agentName) string`을 추가하는 방향 권장.

### Opus 4.7 프롬프트 철학 5원칙 적용 매트릭스

| 원칙 | 적용 대상 | 조치 |
|------|----------|------|
| 1. effort 기본값 xhigh | manager-spec, manager-strategy | frontmatter에 `effort: xhigh` |
| 2. 1턴 몰빵(literal instruction) | 6개 reasoning agent | 본문에 "double-check/verify N times" 스캐폴딩 제거 |
| 3. Extended Thinking 고정 예산 제거 | moai-workflow-thinking/SKILL.md | "thinking budget N tokens" 지시 삭제 |
| 4. 서브에이전트 자동 스폰 감소 | moai-constitution.md | 팬아웃 명시 원칙 추가(MoAI가 명시적 팬아웃 제공) |
| 5. 툴 호출↓ 추론↑ | agent-common-protocol.md | 툴 호출 전 추론 단계 권장 가이드 추가 |

### Windows CLAUDE_ENV_FILE 주입 흐름

```
SessionStart hook (internal/hook/session_start.go)
  ├─ runtime.GOOS == "darwin" || "linux"
  │   └─ ensureTmuxGLMEnv (정의: internal/hook/glm_tmux.go:73, 호출: session_start.go:92, 불변)
  └─ runtime.GOOS == "windows"
      └─ injectCLAUDEEnvFile (신규 — glm.go:489 injectGLMEnvForTeam 스타일로 신규 정의)
          ├─ Read envkeys.go CLAUDE_ENV_FILE constant (internal/config/envkeys.go:70-72)
          ├─ Check file existence (graceful skip)
          └─ Set process env for child Claude Code process
```

Reference 함수 정의 위치 요약(실측):

- `injectGLMEnvForTeam`: `internal/cli/glm.go:489`
- `ensureTmuxGLMEnv`: `internal/hook/glm_tmux.go:73`
- `ensureGLMCredentials`: `internal/hook/session_start.go:157`
- `ensureTeammateMode`: `internal/hook/session_start.go:274`
- `runDiagnosticChecks`: `internal/cli/doctor.go:118`

## Dependencies

- **Upstream dependencies**: 없음 (독립 실행 가능)
- **Downstream impacts**:
  - `SPEC-GLM-001` (GLM settings.json 자동 주입)과 `settings.json.tmpl` 동시 수정 주의
  - `SPEC-CC297-001` (Claude Code 2.9.7)과 `session_start.go` 동시 수정 주의
- **External dependencies**:
  - Claude Code v2.1.110 이상 (runtime 대응 검증용)
  - Opus 4.7 접근 가능 Anthropic API 키(수동 검증용, CI에서는 mock)

## Testing Strategy

- **Unit tests**:
  - `internal/template/model_policy_test.go`: Effort 5단계 라우팅, 모델 ID 검증
  - `internal/profile/preferences_test.go`: YAML round-trip (Opus 4.7 + xhigh)
  - `internal/cli/doctor_test.go`: MCP scope 중복 감지 (픽스처 기반)
  - `internal/hook/session_start_test.go`: Windows 분기(`runtime.GOOS` override) + macOS/Linux 회귀
- **Integration tests**:
  - `moai init` → Opus 4.7 선택 → `preferences.yaml` 내용 검증
  - `moai update` → 기존 프로젝트에 템플릿 변경 적용, 회귀 없음 확인
  - `make build` 후 `embedded.go`와 `internal/template/templates/` diff 정합성
- **Manual verification**:
  - Windows VM에서 `moai hook session-start` 실행 후 `CLAUDE_ENV_FILE` 환경변수 확인
  - Opus 4.7 세션에서 `effort: xhigh`로 지정된 manager-spec 호출 시 reasoning tier 반영 확인

## Rollout Plan

1. **release/v2.11.0 브랜치 내 작업 (현재 브랜치 유지)**: P0/P1/P2 모든 변경을 release/v2.11.0 위에서 commit
2. **P0 완료 후 내부 검증**: macOS + Opus 4.7 조합으로 주요 플로우 smoke test
3. **P1 완료 후 Windows 검증**: VM/CI에서 `GOOS=windows` 빌드 통과 + hook 동작 확인
4. **P2 완료 후 CHANGELOG 작성** → v2.12.0 tag 후보
5. **PR 생성** (manager-git 위임): release/v2.11.0 → main, issue_number는 GitHub Issue 생성 후 SPEC frontmatter 업데이트

## MX Tag Plan (Phase 3.5)

- `@MX:ANCHOR`: `internal/profile/preferences.go`의 `ProfilePreferences` 구조체(fan_in 높음), `internal/template/model_policy.go`의 `ModelPolicy` 상수군 및 `agentModelMap`
- `@MX:WARN`: `internal/cli/doctor.go` MCP scope 검증 함수(파일 시스템 읽기 + JSON 파싱, 실패 시 사일런트 fallback 위험) — `@MX:REASON` 필수
- `@MX:NOTE`: `internal/hook/session_start.go` Windows `CLAUDE_ENV_FILE` 분기 (플랫폼별 동작 차이)
- `@MX:TODO`: P2 문서 항목은 Run phase에서 GREEN 완료 후 해결 예정

## Post-Implementation Review Items

- Opus 4.6/Sonnet 4.6 사용자가 `moai update` 수행 시 `preferences.yaml`이 손상되지 않는가?
- 28개 agent 전체를 조사해 `effort: xhigh`/`high` 미설정 agent가 Opus 4.7 세션에서도 기본값 `xhigh`로 정상 동작하는가?
- `disableBypassPermissionsMode: false` 기본값이 기존 `bypassPermissions` 워크플로우를 차단하지 않는가?
- `moai-workflow-thinking/SKILL.md`에서 고정 예산 지시 제거 후 agent들이 Sequential Thinking MCP를 여전히 올바르게 호출하는가?
