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

# Acceptance Criteria: SPEC-OPUS47-COMPAT-001

## Primary Scenarios (Given / When / Then)

### Scenario GWT-1 — REQ-OC-001: Opus 4.7 프로파일 생성

- **Given**: 사용자가 신규 MoAI 프로젝트에서 `moai profile` 명령으로 대화형 설정을 시작하고, moai-adk-go 바이너리는 본 SPEC의 P0 변경사항이 적용된 버전이다.
- **When**: 사용자가 모델 선택 단계에서 `claude-opus-4-7`을 선택하고 effort 단계에서 `xhigh`를 선택한 뒤 저장을 완료한다.
- **Then**:
  - `~/.moai/claude-profiles/<name>/preferences.yaml` 파일이 생성된다
  - 파일 내용에 `model: claude-opus-4-7` 라인이 존재한다
  - 파일 내용에 `effort_level: xhigh` (또는 동등한 YAML 키) 라인이 존재한다
  - `moai cc` 실행 시 Claude Code 런처에 해당 모델 ID가 전달된다
  - 기존 `claude-opus-4-6` 선택지도 여전히 나타난다 (하위 호환성)

### Scenario GWT-2 — REQ-OC-002: Agent frontmatter effort 필드 전파

- **Given**: `internal/template/templates/.claude/agents/moai/manager-spec.md`에 `effort: xhigh`가 추가되고 `make build`가 수행되었다.
- **When**: 사용자가 `moai init test-project`로 신규 프로젝트를 초기화한다.
- **Then**:
  - `test-project/.claude/agents/moai/manager-spec.md` 파일이 배포된다
  - 해당 파일의 YAML frontmatter에 `effort: xhigh` 필드가 정확히 존재한다
  - manager-strategy는 `effort: xhigh`, plan-auditor/evaluator-active/expert-security/expert-refactoring는 `effort: high`로 설정되어 있다
  - 나머지 22개 agent(research.md D-6 Agent Inventory 기준)의 frontmatter에는 `effort` 키가 존재하지 않는다 (런타임에서 Opus 4.7 기본값 `xhigh`, 타 모델은 `high`로 내부 폴백 처리됨은 단위 테스트로 별도 검증)

### Scenario GWT-3 — REQ-OC-002: skill-authoring.md 문구 정정

- **Given**: P0 적용 전 `skill-authoring.md`는 `effort: Session effort override: low, medium, high, max (max is Opus 4.6 only)` 문구를 포함한다.
- **When**: P0 적용 후 `make build`가 수행되고 `internal/template/embedded.go`가 재생성된다.
- **Then**:
  - `rg "max is Opus 4.6 only" internal/template/templates/`가 0건을 반환한다
  - `skill-authoring.md`에 `effort` 5단계(`low`/`medium`/`high`/`xhigh`/`max`)가 명시되며 Opus 4.7 주석이 포함된다
  - 기존 다른 스킬 문서에서도 동일 문구가 잔존하지 않는다

### Scenario GWT-4 — REQ-OC-003: moai doctor MCP 중복 경고

- **Given**: 테스트 프로젝트의 `.mcp.json`에 `context7` 서버가 `project` scope로 등록되어 있고, 동시에 `.claude/settings.json`에도 `context7` 서버가 `local` scope로 등록되어 있다.
- **When**: 사용자가 `moai doctor`를 실행한다.
- **Then**:
  - 실행 결과에 MCP scope 중복 관련 경고 메시지가 출력된다
  - 경고 메시지는 중복된 서버 이름(`context7`)과 두 스코프 출처를 명시한다
  - exit code는 0이다 (warning만 발생, 차단 금지)
  - `.mcp.json`만 존재하고 `settings.json`에 MCP 설정이 없는 경우 경고는 발생하지 않는다

### Scenario GWT-5 — REQ-OC-003: settings.json 정책 키 + Bash timeout 문서화

- **Given**: P1 적용 후 `internal/template/templates/.claude/settings.json`과 `agent-authoring.md`/`agent-common-protocol.md`가 업데이트되었다.
- **When**: 사용자가 `moai init test-project`를 실행하거나 `moai update`를 실행한다.
- **Then**:
  - 렌더링된 `test-project/.claude/settings.json`에 `disableBypassPermissionsMode` 키가 존재한다 (기본값 `false`, 하위 호환성)
  - JSON validity가 유지되며 `jq . settings.json` 파싱이 성공한다
  - `agent-authoring.md`에서 `rg "600,?000ms"` 또는 `rg "Bash tool timeout ceiling"` 검색 시 1건 이상 매치(Bash tool timeout 상한 600,000ms가 명시적으로 문서화됨)
  - `agent-common-protocol.md`에서도 동일 내용이 문서화됨(`rg "600,?000ms" internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` ≥ 1건)
  - 주의: Bash timeout 상한의 실제 enforcement는 Claude Code 런타임이 수행하며, moai-adk-go 수준에서는 문서화만 수행한다(REQ-OC-003(c) 정의와 일치, Exclusions 정책 준수)

### Scenario GWT-6 — REQ-OC-004: Windows CLAUDE_ENV_FILE 주입

- **Given**: `GOOS=windows` 로 빌드된 moai 바이너리를 Windows 환경(또는 Wine/VM)에서 실행하며, `envkeys.go`의 `CLAUDE_ENV_FILE` 상수에 대응하는 환경 파일이 존재한다.
- **When**: Claude Code가 `moai hook session-start`를 호출한다.
- **Then**:
  - 훅 실행 결과로 `CLAUDE_ENV_FILE` 환경변수가 자식 프로세스(Claude Code)에 주입된다
  - 주입된 값은 `envkeys.go:70-72`에 정의된 상수와 일치한다
  - macOS/Linux에서 동일 훅 실행 시 기존 `injectGLMEnvForTeam`/`ensureTmuxGLMEnv` 동작이 변경 없이 유지된다 (회귀 없음)
  - 환경 파일이 존재하지 않는 경우 graceful skip(에러 없이 진행)한다

### Scenario GWT-7 — REQ-OC-005: 프롬프트 스캐폴딩 제거

- **Given**: 본 SPEC P0 적용 전 `manager-spec.md` 본문에 "double-check your output before returning" 또는 "verify 2-3 times" 류 문구가 포함되어 있다.
- **When**: P0 적용 후 `make build`가 수행된다.
- **Then**:
  - `rg -l "double-check|verify \d+ times|explicitly confirm before" internal/template/templates/.claude/agents/moai/manager-spec.md`가 **0건**을 반환한다(하드 기준)
  - manager-strategy, plan-auditor, evaluator-active, expert-security, expert-refactoring 5개 파일에서도 동일 rg 검색 시 각각 **0건** 반환(하드 기준)
  - 워크플로우 핵심 스텝(Step 1-6 등)은 삭제되지 않고 유지된다 (feedback_agent_refactor_constraint 준수). 워크플로우 스텝 수는 P0 적용 전과 동일하거나 더 많아야 한다.

### Scenario GWT-8 — REQ-OC-005: Adaptive Thinking 고정 예산 지시 제거

- **Given**: P0 적용 전 `moai-workflow-thinking/SKILL.md` 또는 관련 모듈에 "thinking budget 10000 tokens" 또는 "set thinking.budget_tokens to N" 류 지시가 존재한다.
- **When**: P0 적용 후 `make build` 수행.
- **Then**:
  - `rg -l "thinking\.budget_tokens|thinking budget \d+" internal/template/templates/.claude/skills/moai-workflow-thinking/`가 0건을 반환한다
  - SKILL.md에 "Adaptive Thinking is Opus 4.7's only supported thinking mode" 같은 원칙 설명이 존재한다
  - Sequential Thinking MCP(`--deepthink`, `server_tool_use` 경로) 관련 지시는 온전히 유지된다 (Context Dump §F 충돌 해소 검증)

## Edge Cases

### EC-1 — Opus 4.6 사용자의 기존 프로파일 보존

- **Given**: 사용자가 본 SPEC 적용 전 이미 `model: claude-opus-4-6`으로 프로파일을 저장해둔 상태이다.
- **When**: 사용자가 본 SPEC P0가 적용된 새 바이너리로 `moai cc`를 실행한다.
- **Then**: 기존 preferences.yaml이 변경되지 않고 Opus 4.6 세션이 정상 시작된다. 4.7로의 강제 마이그레이션은 발생하지 않는다.

### EC-2 — effort 필드 미지원 모델에서 무시 처리

- **Given**: 사용자의 프로파일이 `model: claude-sonnet-4-6`이고 agent frontmatter에 `effort: xhigh`가 설정되어 있다.
- **When**: Claude Code 세션이 시작되고 해당 agent가 호출된다.
- **Then**: Sonnet 4.6은 effort 필드를 지원하지 않으므로 해당 값이 무시되고 정상 동작한다. moai-adk-go가 warning을 출력하거나 차단하지 않는다.

### EC-3 — GLM 모드에서 effort 필드 처리

- **Given**: 사용자가 `moai glm` 모드로 실행 중이다 (`GLM_API_KEY` 환경 주입, `glm-5.1` 모델).
- **When**: agent에 `effort: xhigh`가 설정되어 있다.
- **Then**: moai-adk-go 레벨에서 에러가 발생하지 않는다. 구체적으로: (1) `moai glm` 실행 시 non-zero exit이 발생하지 않고, (2) `llm.yaml`의 `glm.models` 매핑이 그대로 적용되며, (3) settings.local.json 주입 과정에서 effort 필드로 인한 파싱 실패가 없다. GLM 엔드포인트(`https://api.z.ai/api/anthropic`)의 응답 처리는 외부 시스템 동작이므로 본 SPEC 테스트 대상 외.

### EC-4 — MCP scope 중복 의도적 override 케이스

- **Given**: 사용자가 `.mcp.json`에 `project` scope로 `context7`을 정의하고, 개인 개발 환경에서만 `.claude/settings.local.json`에 `local` scope로 override를 추가한 상태이다.
- **When**: `moai doctor` 실행.
- **Then**: 경고는 발생하되, 메시지에 "local scope override는 의도적일 수 있음"이라는 안내가 포함된다. exit code 0 유지.

### EC-5 — macOS에서 CLAUDE_ENV_FILE 주입 비활성

- **Given**: macOS 환경에서 `moai hook session-start`를 실행한다.
- **When**: 훅 핸들러가 `runtime.GOOS`를 확인한다.
- **Then**: `darwin` 분기이므로 `injectCLAUDEEnvFile` 신규 함수는 호출되지 않는다. 기존 tmux GLM env 주입 경로만 실행된다.

## Quality Gates

### QG-1 — LSP Quality Gate (run phase)

- `.moai/config/sections/quality.yaml`의 `lsp_quality_gates.run.max_errors: 0` 기준 충족
- `go vet ./...` 0 errors
- `golangci-lint run` 0 errors
- `go test -race ./...` 전체 통과
- `go test -cover ./internal/profile/... ./internal/template/... ./internal/cli/... ./internal/hook/...` 85%+ coverage 유지

### QG-2 — Template Integrity

- `make build` 후 `git diff --exit-code internal/template/embedded.go` 통과 (재생성 누락 방지)
- `internal/template/templates/` 변경 파일 수와 `embedded.go` 반영 파일 수 일치

### QG-3 — Backward Compatibility

- 기존 Opus 4.6/Sonnet 4.6/Haiku 4.5 모델이 카탈로그에 여전히 존재
- 기존 `ModelPolicy` 3단계(`high`/`medium`/`low`) API 시그니처 유지(`IsValidModelPolicy`, `ValidModelPolicies`, `GetAgentModel`)
- 기존 `.moai/config/sections/llm.yaml` 사용자 설정 마이그레이션 없이 동작

### QG-4 — Language Neutrality (NFR-3)

- `internal/template/templates/` 전체에서 특정 언어를 "PRIMARY"로 배치한 코드 없음
- 16개 지원 언어(go/python/typescript/.../flutter)가 동등 수준으로 유지됨
- 본 SPEC으로 인해 특정 언어만 enabled 되고 나머지가 planned로 격하되지 않음

### QG-5 — SPEC vs Implementation Boundary

- 본 SPEC 단계에서 실제 Go 코드·YAML·Markdown 파일 수정이 발생하지 않았음 (Plan-Run 경계 유지)
- 수정은 `/moai run SPEC-OPUS47-COMPAT-001` 단계에서만 수행

## Definition of Done (DoD)

### DoD Checklist

- [ ] spec.md의 5개 EARS 요구사항(REQ-OC-001 ~ REQ-OC-005)이 모두 acceptance 시나리오에 매핑되었다
- [ ] Primary Scenarios GWT-1 ~ GWT-8 모두 테스트 통과
- [ ] Edge Cases EC-1 ~ EC-5 모두 검증 완료
- [ ] Quality Gates QG-1 ~ QG-5 모두 통과
- [ ] P0/P1/P2 각 phase 별로 독립 commit 가능한 상태
- [ ] Opus 4.7 모델 카탈로그에 `claude-opus-4-7` 등록 확인
- [ ] 6개 reasoning agent의 frontmatter에 `effort` 필드 설정 확인
- [ ] `skill-authoring.md`에서 "max is Opus 4.6 only" 문구 제거 확인
- [ ] `moai doctor` MCP 중복 감지 동작 확인
- [ ] Windows `CLAUDE_ENV_FILE` 주입 확인(VM 또는 `GOOS=windows` cross-compile 테스트)
- [ ] `settings.json`에 `disableBypassPermissionsMode` 키 존재 확인
- [ ] `moai-workflow-thinking/SKILL.md`에서 고정 예산 지시 제거 확인
- [ ] `make build` 실행 및 `embedded.go` 재생성 커밋 포함
- [ ] `go test -race ./...` 전체 통과
- [ ] `golangci-lint run` 0 errors
- [ ] GitHub Issue 생성 완료 및 `issue_number` frontmatter 업데이트
- [ ] CHANGELOG.md v2.12.0 항목 추가(P2)

### Traceability Matrix

| EARS 요구사항 | Primary Scenarios | Edge Cases | Quality Gates |
|---------------|-------------------|------------|---------------|
| REQ-OC-001    | GWT-1             | EC-1, EC-3 | QG-1, QG-3    |
| REQ-OC-002    | GWT-2, GWT-3      | EC-2       | QG-2, QG-4    |
| REQ-OC-003    | GWT-4, GWT-5      | EC-4       | QG-1, QG-2    |
| REQ-OC-004    | GWT-6             | EC-5       | QG-1          |
| REQ-OC-005    | GWT-7, GWT-8      | -          | QG-2          |

## Rollback Criteria

다음 중 하나라도 발생 시 본 SPEC 적용을 즉시 revert하고 대안 설계를 수립한다:

- Opus 4.6 사용자의 기존 `preferences.yaml`이 자동 마이그레이션 과정에서 손상
- `session_start.go` Windows 분기 추가가 macOS/Linux tmux GLM 주입을 깨뜨림 (R-P1-1 현실화)
- `disableBypassPermissionsMode: false` 기본값이 실제로는 `true`로 해석되어 기존 bypassPermissions 워크플로우 전면 차단
- `make build` 후 `embedded.go` 크기가 비이성적으로 증가(10% 초과)하여 바이너리 사이즈 영향 과다
