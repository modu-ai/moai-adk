---
id: SPEC-OPUS47-COMPAT-001
version: 0.1.0
status: ready-for-run
created: 2026-04-17
updated: 2026-04-17
author: manager-strategy
priority: critical
parent_spec: SPEC-OPUS47-COMPAT-001
total_tasks: 30
---

# Task Decomposition: SPEC-OPUS47-COMPAT-001

SPEC: SPEC-OPUS47-COMPAT-001
Methodology: TDD (RED-GREEN-REFACTOR per quality.yaml)
Phases: P0 (15 files / 16 tasks) → P1 (8 files / 9 tasks) → P2 (4 files / 5 tasks)
Architecture Decision: Option B (ModelPolicy + Effort 분리, GetAgentEffort 신규 함수 도입)

## Task Conventions

- 각 T-XXX는 단일 RED-GREEN-REFACTOR 사이클로 완료 가능한 단위
- Status 값: `pending` | `in_progress` | `completed` | `blocked`
- Planned Files 컬럼은 Drift Guard에서 사용 — 실제 작업 파일과 일치해야 함
- 의존성(Dependencies)은 선행 task ID 또는 `-`(없음)
- T-RED 표기는 RED 단계(테스트 작성), T-IMPL 표기는 GREEN/REFACTOR 단계(구현)
- 템플릿 파일 수정 task 완료 후 T-022(make build)에서 일괄 embedded.go 재생성

---

## Phase P0 — 코어 호환성 + 프롬프트 철학 (16 tasks)

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-001-RED | `ProfilePreferences` Opus 4.7 + `EffortLevel` 필드 YAML round-trip 실패 테스트 작성 | REQ-OC-001 | - | `internal/profile/preferences_test.go` | pending |
| T-001-IMPL | `ProfilePreferences.Model` 주석 갱신(`claude-opus-4-7`) + `EffortLevel string` 필드 추가, `migrateLegacyFields` 수정 | REQ-OC-001 | T-001-RED | `internal/profile/preferences.go` | pending |
| T-002-RED | `GetAgentEffort(agentName) string` 미존재 / `xhigh` `max` 상수 미존재 컴파일 실패 테스트 작성 | REQ-OC-001 | - | `internal/template/model_policy_test.go` | pending |
| T-002-IMPL | `ModelPolicy` 상수 유지(`high`/`medium`/`low`) + 신규 `agentEffortMap map[string]string` 도입, `GetAgentEffort` 함수 신규 추가, `claude-opus-4-7` 모델 ID 상수 정의 | REQ-OC-001 | T-002-RED | `internal/template/model_policy.go` | pending |
| T-003-RED | `profile_setup_translations`에 Opus 4.7 + xhigh/max 라벨 부재 검증 테스트 | REQ-OC-001 | - | `internal/cli/profile_setup_translations_test.go` | pending |
| T-003-IMPL | 한국어/영문 번역 엔트리에 `claude-opus-4-7` 모델 라벨 + 5단계 effort 라벨(`low`/`medium`/`high`/`xhigh`/`max`) 추가 | REQ-OC-001 | T-003-RED, T-002-IMPL | `internal/cli/profile_setup_translations.go` | pending |
| T-004-RED | `launcher.go`의 `claude-opus-4-7` 모델 ID 라우팅 부재 테스트(현 로직 분기 검증) | REQ-OC-001 | - | `internal/cli/launcher_test.go` | pending |
| T-004-IMPL | `claude-opus-4-7` 모델 ID 분기 + `EffortLevel`을 `CLAUDE_CODE_EFFORT_LEVEL`(or 결정된 env 키)로 child 프로세스에 전달 | REQ-OC-001 | T-004-RED, T-001-IMPL, T-002-IMPL | `internal/cli/launcher.go` | pending |
| T-005-IMPL | `quality.yaml`에 `session_effort_default: xhigh` 키 추가 (테스트는 T-006과 통합) | REQ-OC-001 | - | `internal/template/templates/.moai/config/sections/quality.yaml` | pending |
| T-006-IMPL | `llm.yaml`의 `claude_models.high: claude-opus-4-7` 매핑 + effort 매핑 예시 주석 추가 | REQ-OC-001 | T-002-IMPL | `internal/template/templates/.moai/config/sections/llm.yaml` | pending |
| T-007-IMPL | `agent-authoring.md`에 `effort` 5단계 스펙(Opus 4.7 전용 주석 포함) 추가 | REQ-OC-002 | - | `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | pending |
| T-008-IMPL | `skill-authoring.md`에서 `max is Opus 4.6 only` 문구 삭제, `xhigh` 추가, Opus 4.7 노트 갱신 | REQ-OC-002 | - | `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` | pending |
| T-009-IMPL | `manager-spec.md` frontmatter `effort: xhigh` 추가 + 본문 1턴 몰빵 리라이트(스캐폴딩 제거, 워크플로우 스텝 6개 유지) | REQ-OC-002, REQ-OC-005 | T-007-IMPL | `internal/template/templates/.claude/agents/moai/manager-spec.md` | pending |
| T-010-IMPL | `manager-strategy.md` frontmatter `effort: xhigh` + 명시적 팬아웃 원칙 추가 + 스캐폴딩 제거(워크플로우 7-step 유지) | REQ-OC-002, REQ-OC-005 | T-007-IMPL | `internal/template/templates/.claude/agents/moai/manager-strategy.md` | pending |
| T-011-IMPL | `plan-auditor.md` / `evaluator-active.md` / `expert-security.md` / `expert-refactoring.md` 4개 파일에 `effort: high` frontmatter + 스캐폴딩 제거(파일별 워크플로우 스텝 보존) | REQ-OC-002, REQ-OC-005 | T-007-IMPL | `internal/template/templates/.claude/agents/moai/plan-auditor.md`, `internal/template/templates/.claude/agents/moai/evaluator-active.md`, `internal/template/templates/.claude/agents/moai/expert-security.md`, `internal/template/templates/.claude/agents/moai/expert-refactoring.md` | pending |
| T-012-IMPL | `moai-workflow-thinking/SKILL.md`에서 `thinking budget N tokens` 류 고정 예산 지시 제거, "Adaptive Thinking is Opus 4.7's only supported thinking mode" 원칙 추가, Sequential Thinking MCP 지시 보존 | REQ-OC-005 | - | `internal/template/templates/.claude/skills/moai-workflow-thinking/SKILL.md` | pending |
| T-013-VERIFY | `make build` 실행 → `embedded.go` 재생성 → `git diff --exit-code internal/template/embedded.go` 검증 + `go test -race ./internal/profile/... ./internal/template/... ./internal/cli/...` 통과 | REQ-OC-001, REQ-OC-002, REQ-OC-005 | T-001-IMPL ~ T-012-IMPL | `internal/template/embedded.go`(generated) | pending |

---

## Phase P1 — v2.1.110 Runtime 대응 (9 tasks)

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-014-RED | `doctor.go` MCP scope 중복 감지 미존재 실패 테스트(픽스처: `.mcp.json` + `settings.json` 동일 서버명, 다른 scope) | REQ-OC-003 | - | `internal/cli/doctor_test.go` | pending |
| T-014-IMPL | `runDiagnosticChecks` 내부에 `checkMCPScopeDuplicates` 함수 추가, `.mcp.json` + `settings.json` + user config 3 원천 파싱(graceful skip), warning only(exit 0) | REQ-OC-003 | T-014-RED | `internal/cli/doctor.go` | pending |
| T-015-RED | `permission_request.go`의 `updatedInput` deny 재검증 미수행 실패 테스트 | REQ-OC-003 | - | `internal/hook/permission_request_test.go` | pending |
| T-015-IMPL | `updatedInput` 반환 시 기존 deny 패턴 매칭 함수 재호출하여 재검증, 우회 차단 | REQ-OC-003 | T-015-RED | `internal/hook/permission_request.go` | pending |
| T-016-RED | `session_start.go` Windows 분기 부재 + `runtime.GOOS=="windows"` 시 `CLAUDE_ENV_FILE` 미주입 실패 테스트 + macOS/Linux 회귀 테스트(기존 `ensureTmuxGLMEnv` 호출 유지) | REQ-OC-004 | - | `internal/hook/session_start_test.go` | pending |
| T-016-IMPL | `injectCLAUDEEnvFile` 함수 신규 추가(`internal/cli/glm.go:489` `injectGLMEnvForTeam` 스타일), `runtime.GOOS=="windows"` 가드 내 호출, `envkeys.go:70-72` 상수 사용, 파일 부재 시 graceful skip. macOS/Linux 경로 무영향 | REQ-OC-004 | T-016-RED | `internal/hook/session_start.go` | pending |
| T-017-IMPL | `settings.json` 템플릿에 `disableBypassPermissionsMode: false` 키 추가, Bash timeout 600,000ms 상한 문서화 주석 추가 (JSON validity 유지, SPEC-GLM-001 블록과 분리) | REQ-OC-003 | - | `internal/template/templates/.claude/settings.json` | pending |
| T-018-IMPL | `agent-common-protocol.md`에 Bash timeout 600,000ms 상한 문서화 + 명시적 팬아웃 원칙 추가; `agent-authoring.md`에도 Bash timeout 상한 문서화 추가; `moai-constitution.md`에 Opus 4.7 원칙 4(팬아웃 명시) + 원칙 5(툴 호출↓ 추론↑) 섹션 추가; `worktree-integration.md`에 v2.1.110 (MCP doctor) + v2.1.111 (CLAUDE_ENV_FILE) 최소 버전 추가 | REQ-OC-003 | T-007-IMPL | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`, `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`, `internal/template/templates/.claude/rules/moai/core/moai-constitution.md`, `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | pending |
| T-019-IMPL | `harness.yaml`에 level↔effort 매핑 추가(thorough→xhigh, standard→high, minimal→medium), 본 SPEC `model_upgrade_review` 실행 기록 | REQ-OC-002 | T-002-IMPL | `internal/template/templates/.moai/config/sections/harness.yaml` | pending |
| T-020-VERIFY | `make build` → `embedded.go` 재생성 + `go test -race ./internal/cli/... ./internal/hook/...` 전체 통과 + `GOOS=windows go build ./...` cross-compile 검증 | REQ-OC-003, REQ-OC-004 | T-014-IMPL ~ T-019-IMPL | `internal/template/embedded.go`(generated) | pending |

---

## Phase P2 — 문서 레이어 (5 tasks)

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-021-IMPL | `CLAUDE.md §12`에 UltraThink vs Adaptive Thinking 용어 정리 (Sequential Thinking MCP `server_tool_use`와 Adaptive Thinking `thinking.budget_tokens` 차이 명시) | REQ-OC-005 | T-012-IMPL | `internal/template/templates/CLAUDE.md` | pending |
| T-022-IMPL | `coding-standards.md`에 v2.1.111 read-only Bash 자동 분류 + `/less-permission-prompts` 가이드 추가 | REQ-OC-003 | - | `internal/template/templates/.claude/rules/moai/development/coding-standards.md` | pending |
| T-023-IMPL | `CLAUDE.local.md`에 `settings.local.json` 분리 원칙 + `OTEL_LOG_RAW_API_BODIES` 프로덕션 금지 경고 추가 (로컬 전용, 템플릿 외) | GOAL-2 | - | `CLAUDE.local.md` | pending |
| T-024-IMPL | `CHANGELOG.md`에 v2.12.0 항목 추가: "Claude Code v2.1.110+ 권장, Opus 4.7 + effort 5단계 + 프롬프트 철학 적용" 한국어 작성 | GOAL-2 | T-013-VERIFY, T-020-VERIFY | `CHANGELOG.md` | pending |
| T-025-VERIFY | 최종 검증: `make build` → `git diff --exit-code internal/template/embedded.go` + `go test -race ./...` 전체 통과 + `golangci-lint run ./...` 0 errors + `rg "max is Opus 4.6 only" internal/template/templates/`가 0건 + `rg -l "double-check\|verify \\d+ times\|explicitly confirm before" internal/template/templates/.claude/agents/moai/{manager-spec,manager-strategy,plan-auditor,evaluator-active,expert-security,expert-refactoring}.md`가 0건 + `rg "thinking\\.budget_tokens\|thinking budget \\d+" internal/template/templates/.claude/skills/moai-workflow-thinking/`가 0건 + GWT-1~8 시나리오 수동 검증 + DoD 체크리스트 전체 통과 | ALL | T-001 ~ T-024 | (verification) | pending |

---

## Coverage Verification (REQ → Task Mapping)

| EARS Requirement | Tasks Covering | Total |
|------------------|----------------|-------|
| REQ-OC-001 (모델/effort 카탈로그) | T-001-RED, T-001-IMPL, T-002-RED, T-002-IMPL, T-003-RED, T-003-IMPL, T-004-RED, T-004-IMPL, T-005-IMPL, T-006-IMPL, T-013-VERIFY | 11 |
| REQ-OC-002 (Opus 4.7 프롬프트 철학) | T-007-IMPL, T-008-IMPL, T-009-IMPL, T-010-IMPL, T-011-IMPL, T-013-VERIFY, T-019-IMPL | 7 |
| REQ-OC-003 (v2.1.110 Runtime) | T-014-RED, T-014-IMPL, T-015-RED, T-015-IMPL, T-017-IMPL, T-018-IMPL, T-020-VERIFY, T-022-IMPL | 8 |
| REQ-OC-004 (Windows CLAUDE_ENV_FILE) | T-016-RED, T-016-IMPL, T-020-VERIFY | 3 |
| REQ-OC-005 (스캐폴딩 제거) | T-009-IMPL, T-010-IMPL, T-011-IMPL, T-012-IMPL, T-021-IMPL, T-025-VERIFY | 6 |

**모든 REQ-OC-001~005가 최소 3개 이상 task로 cover됨**.

---

## Sequencing Notes

1. **Phase 의존성**: P0 → P1 → P2 (순차). P0의 Go 코드(T-001~T-004)는 템플릿(T-005~T-012)보다 먼저 완료해야 `harness.yaml`/`llm.yaml` 검증 가능.
2. **Embedded.go 재생성**: T-013-VERIFY(P0 종료), T-020-VERIFY(P1 종료), T-025-VERIFY(P2 종료) 3회 실행. 중간 commit 시점마다 `git diff --exit-code internal/template/embedded.go` 통과 필수.
3. **Windows 회귀 가드**: T-016-RED는 macOS/Linux 기존 동작(`ensureTmuxGLMEnv`, `injectGLMEnvForTeam`) 보존을 명시적으로 검증해야 함(R-P1-1 mitigation).
4. **`agent-authoring.md` 2회 수정**: T-007-IMPL(effort 스펙)과 T-018-IMPL(Bash timeout 문서화) 두 번 수정. 같은 파일 동시 편집이므로 T-007 → T-018 순차 의존.
5. **`moai-constitution.md` Opus 4.7 원칙 추가**: T-018-IMPL에 통합(원칙 4 팬아웃 + 원칙 5 툴호출↓). 별도 task 분리 불필요.

---

## Risk Mitigation Tasks

R-P0-1 (`agentModelMap` 컴파일 에러) → T-002-RED가 컴파일 실패 자체를 검증 (Go 컴파일러가 차단)
R-P0-2 (Agent 본문 워크플로우 축소) → T-009/T-010/T-011 description에 "워크플로우 스텝 N개 유지" 명시 (feedback_agent_refactor_constraint 준수)
R-P0-3 (`make build` 누락) → T-013/T-020/T-025 모두 `git diff --exit-code internal/template/embedded.go` 검증
R-P1-1 (`session_start.go` macOS/Linux 회귀) → T-016-RED에 회귀 테스트 명시
R-P1-2 (`settings.json` SPEC-GLM-001 충돌) → T-017-IMPL description에 "SPEC-GLM-001 블록과 분리" 명시
R-P1-3 (MCP false-positive) → T-014-IMPL이 warning only로 명시
R-P2-1 (CHANGELOG 버전 충돌) → T-024-IMPL이 P0/P1 완료 후 마지막 실행

---

## TodoWrite 동기화

본 tasks.md는 `/moai run SPEC-OPUS47-COMPAT-001` 실행 시 manager-tdd가 TodoWrite로 변환하여 in_progress 추적한다. 30개 task 모두 `pending` 상태로 시작.
