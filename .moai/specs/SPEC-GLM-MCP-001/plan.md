# SPEC-GLM-MCP-001 구현 계획

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-01 | manager-spec | 초기 작성 — 5개 마일스톤, 5개 리스크 카탈로그, 6개 OQ 결정 표 |

## 개요

본 계획은 SPEC-GLM-MCP-001 의 구현 — `moai glm tools enable|disable [vision|websearch|webreader|all]` subcommand 를 통해 Z.AI 의 공식 `@z_ai/mcp-server` 패키지를 Claude Code MCP 서버로 등록 — 의 마일스톤, 기술적 접근, 리스크, 의존성을 정의한다.

본 문서는 시간 추정 (예: "2 days") 대신 **우선순위 라벨** 과 **단계 간 순서** 로 진행을 표현한다 (CLAUDE.local.md §14 + agent-common-protocol.md §Time Estimation 준수).

## 기술적 접근 (High-Level)

핵심 구현은 다음 4 컴포넌트로 구성된다:

1. **명령 라우터 확장** — `internal/cli/glm.go` 또는 신규 `internal/cli/glm_tools.go` 에 `tools` subcommand 추가. cobra 또는 기존 명령 디스패치 패턴을 따름.
2. **Node.js 버전 검증** — `exec.LookPath("node")` + `exec.Command("node", "--version")` 으로 최소 v22.0.0 검증. 부재/구버전 시 graceful fail.
3. **JSON 머지 로직** — `~/.claude.json` (또는 `.mcp.json` for project scope) 을 안전하게 read → unmarshal → mcpServers 키만 부분 갱신 → marshal → 백업 후 atomic write. 다른 MCP 엔트리 (`context7`, `sequential-thinking`, `moai-lsp`, 사용자 정의) 는 절대 손대지 않음.
4. **토큰 재사용** — 기존 `loadGLMKey()` 헬퍼 (`~/.moai/.env.glm` 의 `GLM_AUTH_TOKEN`) 를 재사용하여 사용자가 토큰을 재입력하지 않도록 함. 토큰 부재 시 친절한 안내.

테스트 측면에서는 기존 `internal/cli/glm_test.go` 의 패턴 (table-driven, t.TempDir() 격리, `loadGLMKey()` 모킹) 을 따른다. 신규 테스트 함수는 `TestGLMToolsEnable*`, `TestGLMToolsDisable*` prefix 로 grep/search 가능하게 명명한다.

문서 동기화는 `internal/template/templates/.claude/rules/moai/core/settings-management.md` (템플릿) 와 `.claude/rules/moai/core/settings-management.md` (루트 미러본) 두 파일에 동일한 짧은 노트를 추가하는 단순 미러 작업이며, Template-First Rule (CLAUDE.local.md §2) 에 따라 템플릿 본을 먼저 수정한다.

## Open Questions Resolution

research.md §7 의 6개 OQ 에 대한 plan-stage 결정:

| OQ | 결정 | 근거 |
|----|------|------|
| OQ-1 (별도 vs `all`) | **둘 다 지원** — `tools enable [vision\|websearch\|webreader\|all]` | 사용자 통제권 ↑, 단순성 미세 손실은 허용 |
| OQ-2 (user vs project scope) | **user scope 기본**, `--scope project` 옵션 | Z.AI 공식 가이드 일관 + 다중 프로젝트 사용자 편의 |
| OQ-3 (기존 엔트리 처리) | **idempotent skip (토큰 일치)** + **`--force` 안내 (토큰 불일치)** | R1 위험 완화 + 사용자 데이터 보호 |
| OQ-4 (Lite 플랜 처리) | **경고 출력 + 진행** (사전 검증 API 비싸거나 불안정 가능) | EX-2 와 일관, 런타임 에러는 Z.AI 가 반환 |
| OQ-5 (Mainland China v0.1?) | **제외 (EX-1)**, v0.2 로 연기 | 단순화 + 검증 부담 감소 |
| OQ-6 (sample test 자동 실행) | **v0.1 에서 제외 (EX-6)**, 향후 `--test` 플래그 | 첫 실행 ~10s 지연 회피 |

## Milestones

### M1 — Path Discovery + Existing Pattern Review (Priority: High)

**목표**: 변경 대상 파일의 정확한 경로/라인 범위 식별, `internal/cli/glm.go` 의 기존 명령 디스패치 패턴 + `internal/cli/glm_test.go` 의 기존 테스트 패턴 식별.

**작업 항목:**

- `internal/cli/glm.go` 의 명령 라우팅 (cobra 등) 패턴 확인
- `internal/cli/github_auth.go` 의 `moai github auth glm` subcommand 구조 분석 (참조 패턴)
- `internal/cli/glm.go:loadGLMKey()` (또는 동등 함수) 의 정확한 위치, 시그니처, 테스트 헬퍼 확인
- `internal/cli/glm_test.go` 의 기존 케이스가 사용하는 테스트 컨벤션 식별: t.TempDir() 사용 여부, HOME 디렉토리 모킹 방식, JSON fixture 위치
- `~/.claude.json` 의 표준 구조 (`mcpServers` 객체 + 사용자 정의 엔트리 보존 정책) 사례 확인 (SPEC-CC2122-MCP-001 참조)
- `.claude/rules/moai/core/settings-management.md` 와 그 템플릿 미러본의 현재 구조 (섹션 헤더, MCP 관련 기존 설명) 확인하여 노트 삽입 지점 선정

**Done 기준:**

- 수정 대상 파일 모두 정확한 경로/라인 범위가 식별되어 PR 설명에 인용 가능
- 신규 테스트가 따라야 할 기존 컨벤션이 명확히 결정됨
- 기존 `loadGLMKey()` 재사용 가능성 / 보강 필요성 판단

### M2 — JSON Merge Helper + Node.js Version Detection (Priority: High)

**목표**: `~/.claude.json` 의 안전 머지 로직 + Node.js 버전 검증 헬퍼를 패키지로 분리하여 재사용성 + 테스트 용이성 확보.

**작업 항목:**

- 신규 헬퍼 (위치: `internal/cli/glm_tools.go` 또는 `internal/integration/zaimcp/`):
  - `mergeMCPServer(configPath string, name string, entry MCPEntry) error` — atomic write + 백업
  - `removeMCPServer(configPath string, name string) error` — 부분 제거, 다른 엔트리 보존
  - `detectNodeVersion() (semver.Version, error)` — node `--version` 파싱
- 백업 정책: `~/.claude.json.bak-<ISO timestamp>` 생성 (REQ-GMC-005). idempotent skip 시 백업 생략
- atomic write: 임시 파일 → `os.Rename` 패턴 (POSIX atomic guarantee)
- 사용자 정의 엔트리 보존 검증: 기존 `mcpServers` 의 모든 키 중 `zai-mcp-server` 외에는 변경하지 않음 (REQ-GMC-010)
- TDD 사이클: 헬퍼 테스트 먼저 작성 (RED) → 구현 (GREEN)

**Done 기준:**

- 신규 헬퍼 함수가 `internal/cli/...` 또는 `internal/integration/zaimcp/...` 에 추가됨
- 헬퍼별 unit test 가 `t.TempDir()` 격리로 작성됨
- 사용자 정의 MCP 엔트리 보존 검증 테스트 (REQ-GMC-010 매핑) 통과
- Node.js 버전 검증의 3가지 케이스 (정상, 부재, 구버전) 테스트 통과

### M3 — Subcommand Routing + Token Reuse Integration (Priority: High)

**목표**: `moai glm tools enable|disable [도구명]` subcommand 를 명령 라우터에 추가하고, M2 의 헬퍼 + 기존 `loadGLMKey()` 를 통합한다.

**작업 항목:**

- `glm tools enable` 명령:
  - Step 1: `loadGLMKey()` 호출 → 토큰 부재 시 REQ-GMC-007 에러 메시지 + exit code 비-0
  - Step 2: `detectNodeVersion()` 호출 → < v22 시 REQ-GMC-009 에러 메시지 + exit code 비-0
  - Step 3: 기존 `~/.claude.json` 읽고 `zai-mcp-server` 엔트리 검사 → REQ-GMC-006 (a) idempotent skip 또는 (b) `--force` 안내
  - Step 4: `mergeMCPServer()` 호출 → 백업 + atomic write
  - Step 5: 활성화된 도구 목록 + Pro 플랜 권장 + Claude Code 재시작 안내 출력
- `glm tools disable` 명령:
  - Step 1: `~/.claude.json` 읽기
  - Step 2: `removeMCPServer()` 호출
  - Step 3: 제거된 도구 목록 출력
- `--scope project` 옵션 처리: configPath 를 `.mcp.json` (project root) 으로 전환 (REQ-GMC-008)
- 부분 도구 비활성화 모델 결정: v0.1 에서는 `vision|websearch|webreader|all` 모두 동일하게 zai-mcp-server 엔트리 전체를 추가/제거 (단순성). 도구별 활성/비활성 토글은 v0.2 후보
- TDD: 명령 통합 테스트 (`TestGLMToolsEnable*`, `TestGLMToolsDisable*` prefix)

**Done 기준:**

- `moai glm tools enable vision` / `moai glm tools enable all` 이 정상 작동
- `moai glm tools disable all` 이 정상 작동, 다른 MCP 엔트리 보존 (REQ-GMC-010)
- `--scope project` 옵션이 `.mcp.json` 에 기록 (REQ-GMC-008)
- 명령 통합 테스트 ≥ 6개 추가 (각 REQ-GMC-003 ~ 010 매핑)
- `go test -count=1 -race ./internal/cli/...` 통과

### M4 — Documentation Sync (Priority: Medium)

**목표**: `settings-management.md` 두 파일 (템플릿 + 루트 미러본) 에 Z.AI MCP 통합 노트 + `moai glm tools` 사용 가이드를 추가한다.

**작업 항목:**

- `internal/template/templates/.claude/rules/moai/core/settings-management.md` 에 노트 추가 (Template-First Rule, CLAUDE.local.md §2):
  - Z.AI MCP 통합의 의미 (Vision / Web Search / Web Reader)
  - `moai glm tools enable [도구명]` / `moai glm tools disable [도구명]` 사용법
  - Node.js >= v22 + GLM_AUTH_TOKEN 사전 조건
  - Pro 플랜 권장 (Vision / WebSearch)
  - SPEC-CC2122-MCP-001 (`alwaysLoad`) 와의 관계 — zai-mcp-server 는 lazy-load 권장 (스키마 크기 미상, 사용 빈도 미상이므로 conservative default)
- `.claude/rules/moai/core/settings-management.md` 에 동일한 노트 미러링
- 두 파일의 변경분이 동일한지 `diff` 로 검증
- CHANGELOG.md 에 v0.1 항목 추가 (사용자 알림)

**Done 기준:**

- 두 `settings-management.md` 파일에 Z.AI MCP 노트 부착됨, 두 사본 동일 내용
- CHANGELOG.md 에 SPEC-GLM-MCP-001 항목 추가
- 16개 언어 중립성 위반 없음 (CLAUDE.local.md §15) — `internal/template/templates/` 변경분이 특정 언어 편향이 없는지 확인

### M5 — Final Quality Gates + Pre-merge Validation (Priority: High)

**목표**: 모든 품질 게이트 통과 + plan-auditor 사후 감사 준비.

**작업 항목:**

- `go test ./...` (전체) 통과 — cascading failure 검출 (CLAUDE.local.md §6)
- `go test -race ./...` 통과 — concurrency 안전성
- `go vet ./...` 무경고
- `golangci-lint run` 무경고 (가능한 경우)
- `make build` 성공 + `internal/template/embedded.go` 가 최신 템플릿 반영
- 통합 검증 (수동, GLM 토큰 보유 시):
  - `moai glm tools enable all` 실행 → `~/.claude.json` 내용 검증
  - Claude Code 재시작 → MCP 도구 호출 시도 (Vision OCR ping)
  - `moai glm tools disable all` 실행 → 엔트리 제거 검증
  - `moai cc` 모드 전환 후에도 엔트리 유지 확인 (R5 매핑)
- 커밋 메시지 초안: `feat(glm): add 'moai glm tools enable/disable' for Z.AI MCP integration (SPEC-GLM-MCP-001)`

**Done 기준:**

- 모든 자동화 게이트 (vet/test/lint/build) 통과
- 수동 통합 검증 (가능한 경우) 성공
- PR 설명에 SPEC-GLM-MCP-001 명시, REQ ↔ GWT 매핑 표 포함
- plan-auditor 의 사전 감사(`/moai run` Phase 0.5) 통과 준비

## Dependencies

- **상위 의존**: SPEC-GLM-001 (호환성 자동화) — 이미 `completed`. 본 SPEC 은 SPEC-GLM-001 의 환경변수 정책 위에서 작동.
- **하위 의존 후보**:
  - v0.2: Mainland China (`Z_AI_MODE=ZHIPU`) 지원 SPEC
  - v0.2: 도구별 부분 enable/disable 모델 (Vision 만 활성, WebSearch 비활성)
  - v0.2: GLM Coding Plan 등급 사전 검증 (Pro vs Lite)
- **외부 의존**:
  - `@z_ai/mcp-server >= v0.1.2` 가 npm 에 publish 되어 있어야 함 — 이미 확인됨
  - Claude Code 가 `~/.claude.json` 의 `mcpServers` 키를 인식해야 함 — 공식 메커니즘
  - Node.js >= v22.0.0 가 사용자 환경에 설치되어 있어야 함 — REQ-GMC-009 에 의해 검증됨

## Risks and Mitigations

| ID | 리스크 | 영향도 | 완화 전략 | REQ 매핑 |
|----|--------|--------|-----------|----------|
| R1 | 사용자 `~/.claude.json` 에 이미 `zai-mcp-server` 엔트리 존재 → 덮어쓰기 vs 보존 | High | REQ-GMC-006 (a) idempotent skip, (b) `--force` 안내 + 차이 요약. M3 에서 차이 검사 테스트 작성 | REQ-GMC-006 |
| R2 | 사용자 GLM Coding Plan Lite ($3) 보유 → Vision MCP 호출 시 401/403 | Medium | enable 시 명확한 경고 출력 ("Pro 플랜 이상 필요"). 런타임 에러는 Z.AI 응답 그대로 노출 (EX-2). 추후 v0.2 에서 사전 API 검증 검토 | EX-2 |
| R3 | Node.js < 22 → `npx @z_ai/mcp-server` 실패 | High | M2 의 `detectNodeVersion()` 로 사전 검증. 부재/구버전 시 REQ-GMC-009 에 따라 graceful fail + 설치 가이드 출력 | REQ-GMC-009 |
| R4 | `~/.claude.json` (user scope) vs `.mcp.json` (project scope) 결정 모호 | Medium | OQ-2 결정: user scope 기본, `--scope project` 옵트인 (REQ-GMC-008). Z.AI 공식 가이드와 일관 | REQ-GMC-008 |
| R5 | `moai cc` ↔ `moai glm` 모드 전환 시 MCP 설정 영향 (회귀 가능성) | Medium | EX-4 결정: 모드 전환 시 자동 enable/disable 하지 *않음*. 단순성 + 예측 가능성 우선. M5 수동 통합 검증으로 모드 전환 후 엔트리 유지 확인 | EX-4 |
| R6 | 사용자 정의 MCP 엔트리 (`context7`, `sequential-thinking`, `moai-lsp`, 또는 사용자 추가) 손상 | High | REQ-GMC-010 명시적 보존. M2 의 `mergeMCPServer()` / `removeMCPServer()` 가 다른 엔트리 변경 금지. 회귀 검증 테스트 필수 | REQ-GMC-010 |
| R7 | atomic write 실패 (디스크 풀, 권한 문제) → `~/.claude.json` 손상 | High | M2 에서 백업 생성 (`~/.claude.json.bak-<ts>`) + atomic rename. 실패 시 명확한 복구 가이드 출력 | REQ-GMC-005 |
| R8 | `make build` 누락으로 `internal/template/embedded.go` 가 stale | Medium | M5 에서 `make build` 명시적 실행. PR 체크리스트 (CLAUDE.local.md §4) 준수 | M5 |
| R9 | 16개 언어 중립성 위반 (특정 언어 편향 도입) | Low | 본 SPEC 은 영어 도움말 메시지 + 한국어 사용자 메시지. `internal/template/templates/` 의 변경분은 영어 + 16개 언어 동등 사례. CLAUDE.local.md §15 verify | CLAUDE.local.md §15 |

## Out of Scope (Plan-Level)

- Mainland China 엔드포인트 지원 (EX-1, v0.2 후보)
- GLM 플랜 등급 사전 API 검증 (EX-2)
- 다른 MCP 엔트리 (`context7`, `sequential-thinking`, `moai-lsp`) 정책 변경 (EX-3)
- 모드 전환 시 자동 enable/disable (EX-4)
- `~/.claude.json` 스키마 강화 검증 (EX-5)
- 사전 sample test 자동 실행 (EX-6)
- 도구별 부분 enable/disable 모델 (v0.2 후보)
- 4개국어 docs-site (`docs-site/`) 동기화 — 본 SPEC 의 변경은 내부 룰 문서 (`.claude/rules/moai/core/settings-management.md`) 에 한정. 공식 사용자 가이드 갱신 여부는 별도 판단 (CLAUDE.local.md §17.3 ko-canonical 정책 검토 후 필요 시 후속 PR 분리)

## 16개 언어 중립성 검증 체크리스트

본 SPEC 의 변경이 CLAUDE.local.md §15 (Template Language Neutrality) 를 위반하지 않는지 확인:

- [ ] `internal/template/templates/.claude/rules/moai/core/settings-management.md` 의 노트가 특정 언어를 PRIMARY 로 배치하지 않음
- [ ] 16개 언어 (go/python/typescript/javascript/rust/java/kotlin/csharp/ruby/php/elixir/cpp/scala/r/flutter/swift) 가 동등 수준
- [ ] Node.js 22 prerequisite 는 *Z.AI MCP 패키지 실행* 의 외부 요구사항이며, 사용자 프로젝트 언어 선택과 무관함을 노트에 명시
- [ ] 본 통합은 *moai-adk-go 의 내부 CLI 기능* 이며, 16개 언어 사용자 모두에게 동일하게 작동

## Approval / Next Step

본 plan.md 와 spec.md, acceptance.md 가 plan-auditor 의 승인을 받으면, manager-cycle 또는 manager-tdd 가 `/moai run SPEC-GLM-MCP-001` 으로 위임받아 M1 → M5 순으로 구현한다. 본 SPEC 단계에서는 코드/테스트/문서 어떤 파일도 수정하지 않는다 (사용자 명시 제약).

manager-cycle 위임 시 권장 순서:
1. M1 (path discovery)
2. M2 (helpers, TDD RED → GREEN)
3. M3 (subcommand routing, TDD)
4. M4 (docs sync)
5. M5 (quality gates + plan-auditor pre-flight)
