# SPEC-GLM-MCP-001 인수 기준

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-01 | manager-spec | 초기 작성 — REQ-GMC-001 ~ 010 별 GWT 시나리오 (총 22개), Edge Cases, Quality Gate, Definition of Done |

본 문서는 SPEC-GLM-MCP-001 의 구현이 완료되었음을 검증하기 위한 Given-When-Then(GWT) 시나리오와 Definition of Done 을 정의한다. 각 시나리오는 `internal/cli/glm_test.go` 또는 `internal/cli/glm_tools_test.go` 의 테이블 드리븐 / sub-test 구조로 자동화된다.

본 acceptance.md 는 **single source of truth** 이며, spec.md 의 "Acceptance Summary" 섹션은 본 파일의 간단 요약일 뿐이다. SPEC-CI-MULTI-LLM-001 audit defect D5 ("acceptance defer to external file" anti-pattern) 를 회피하기 위해 본 파일에 모든 시나리오를 명시적으로 작성한다.

## REQ ↔ GWT 매핑 표

| REQ | 매핑 GWT | 핵심 검증 대상 |
|-----|---------|---------------|
| REQ-GMC-001 (Ubiquitous, subcommand 존재 + idempotent) | GWT-1, GWT-2 | `tools enable/disable` subcommand 라우팅 + 반복 호출 안전성 |
| REQ-GMC-002 (Ubiquitous, SPEC-GLM-001 호환) | GWT-3 | DISABLE_BETAS / DISABLE_PROMPT_CACHING 정책 보존 |
| REQ-GMC-003 (Event-Driven, enable 전체 경로) | GWT-4, GWT-5 | `~/.claude.json` 에 정확한 엔트리 추가 + 사용자 안내 출력 |
| REQ-GMC-004 (Event-Driven, disable) | GWT-6, GWT-7 | 엔트리 제거 + 다른 엔트리 보존 |
| REQ-GMC-005 (Event-Driven, 백업) | GWT-8, GWT-9 | 백업 파일 생성 + idempotent skip 시 백업 생략 |
| REQ-GMC-006 (Event-Driven, 기존 엔트리 처리) | GWT-10, GWT-11 | 토큰 일치 시 idempotent skip + 토큰 불일치 시 `--force` 안내 |
| REQ-GMC-007 (State-Driven, 토큰 부재) | GWT-12 | 토큰 부재 시 enable 거부 + 가이드 출력 |
| REQ-GMC-008 (Optional, project scope) | GWT-13 | `--scope project` 시 `.mcp.json` 에 기록 |
| REQ-GMC-009 (Unwanted, Node.js 부재/구버전) | GWT-14, GWT-15 | 부재/구버전 시 graceful fail + `~/.claude.json` 변경 없음 |
| REQ-GMC-010 (Unwanted, 사용자 정의 엔트리 보존) | GWT-16, GWT-17 | enable/disable 모두에서 다른 mcpServers 엔트리 보존 |
| Edge Cases | GWT-18 ~ GWT-22 | 모드 전환, atomic write, JSON 파싱, locale, 명령 인자 검증 |

## Given-When-Then 시나리오

### GWT-1: `tools enable` subcommand 라우팅 + idempotent

**Given:** 사용자가 정상 환경 (Node.js >= 22, GLM_AUTH_TOKEN 설정됨, `~/.claude.json` 의 mcpServers 에 zai-mcp-server 부재) 에 있다.
**When:** `moai glm tools enable vision` 을 실행한다. 이어서 동일 명령을 두 번째로 실행한다.
**Then:**
- 첫 실행: 종료 코드 0, `~/.claude.json` 의 `mcpServers.zai-mcp-server` 엔트리 추가
- 두 번째 실행: 종료 코드 0, "이미 활성화됨" 메시지, `~/.claude.json` 내용 변경 없음 (idempotent skip)

### GWT-2: `tools disable` subcommand 라우팅 + idempotent

**Given:** GWT-1 의 첫 실행 직후 상태 (zai-mcp-server 엔트리 존재).
**When:** `moai glm tools disable all` 을 실행한다. 이어서 동일 명령을 두 번째로 실행한다.
**Then:**
- 첫 실행: 종료 코드 0, zai-mcp-server 엔트리 제거됨
- 두 번째 실행: 종료 코드 0, "활성화된 zai-mcp-server 엔트리 없음" 메시지, `~/.claude.json` 변경 없음 (idempotent skip)

### GWT-3: SPEC-GLM-001 호환성 (REQ-GMC-002)

**Given:** 사용자가 `moai glm` 모드에서 작업 중이며, settings.local.json 에 `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1` 와 `DISABLE_PROMPT_CACHING=1` 가 SPEC-GLM-001 에 의해 설정되어 있다.
**When:** `moai glm tools enable all` 을 실행하고, 그 후 `moai cc` 로 모드 전환 후 `moai glm` 으로 다시 전환한다.
**Then:**
- `moai glm tools enable` 동작은 settings.local.json 의 env 키들을 *변경하지 않음*
- `moai cc` → `moai glm` 모드 전환 후에도 SPEC-GLM-001 의 정책이 정상 작동 (기존 회귀 테스트 통과)
- zai-mcp-server 엔트리는 모드 전환 영향을 받지 않음 (사용자 명시 disable 까지 유지)

### GWT-4: enable 의 전체 경로 — 엔트리 정확성 (REQ-GMC-003)

**Given:** Node.js v22.5.0 설치됨, `~/.moai/.env.glm` 에 `GLM_AUTH_TOKEN=test-glm-key-abc123`, `~/.claude.json` 에 zai-mcp-server 부재.
**When:** `moai glm tools enable vision` 을 실행한다.
**Then:** `~/.claude.json` 의 `mcpServers.zai-mcp-server` 엔트리는 정확히 다음을 포함:
- `command: "npx"`
- `args: ["-y", "@z_ai/mcp-server@latest"]`
- `env.Z_AI_API_KEY: "test-glm-key-abc123"`
- `env.Z_AI_MODE: "ZAI"`

JSON 파싱 후 위 4개 필드가 모두 검증되어야 한다. 다른 필드 (timeout 등) 는 본 SPEC 범위 밖이므로 포함 여부 무관.

### GWT-5: enable 의 사용자 안내 출력 (REQ-GMC-003)

**Given:** GWT-4 와 동일.
**When:** `moai glm tools enable all` 을 실행한다.
**Then:** stdout 에 다음 정보가 모두 포함됨:
- 활성화된 도구 목록 (Vision / Web Search / Web Reader)
- "Pro 플랜 ($9/월) 이상 필요" 또는 동등한 plan tier 안내
- "Claude Code 재시작 필요" 안내
- 비활성화 명령 (`moai glm tools disable all`) 안내

### GWT-6: disable 의 엔트리 제거 (REQ-GMC-004)

**Given:** `~/.claude.json` 의 mcpServers 에 zai-mcp-server 엔트리가 존재함.
**When:** `moai glm tools disable all` 을 실행한다.
**Then:** `~/.claude.json` 의 mcpServers 객체에서 zai-mcp-server 키가 제거되고, mcpServers 객체 자체는 (다른 엔트리가 없으면) 빈 객체로 유지되거나, 다른 엔트리가 있으면 그대로 보존됨.

### GWT-7: disable 후 다른 mcpServers 엔트리 보존 (REQ-GMC-004)

**Given:** `~/.claude.json` 의 mcpServers 에 다음 4개 엔트리 존재: `context7`, `sequential-thinking`, `moai-lsp`, `zai-mcp-server`.
**When:** `moai glm tools disable all` 을 실행한다.
**Then:** mcpServers 객체에 정확히 3개 엔트리 (`context7`, `sequential-thinking`, `moai-lsp`) 만 남고, 각 엔트리의 `command`, `args`, `env`, `timeout`, `$comment` 등 모든 필드가 변경 없이 보존됨.

### GWT-8: 백업 파일 생성 (REQ-GMC-005)

**Given:** `~/.claude.json` 의 mcpServers 에 zai-mcp-server 부재, Node.js + 토큰 정상.
**When:** `moai glm tools enable vision` 을 실행하고, 명령 종료 후 `~/.claude.json` 동일 디렉토리를 검사한다.
**Then:** `~/.claude.json.bak-<ISO timestamp>` 패턴의 백업 파일이 정확히 1개 생성됨. 백업 내용은 명령 실행 *직전* 상태의 `~/.claude.json` 와 동일.

### GWT-9: idempotent skip 시 백업 생략 (REQ-GMC-005)

**Given:** `~/.claude.json` 에 zai-mcp-server 엔트리가 이미 존재하고 토큰이 일치함 (GWT-1 의 두 번째 실행 시나리오).
**When:** `moai glm tools enable vision` 을 실행한다.
**Then:** `~/.claude.json.bak-*` 패턴의 백업 파일이 *생성되지 않음* (idempotent skip 은 변경이 없으므로 백업 불필요).

### GWT-10: 기존 엔트리 + 토큰 일치 → idempotent skip (REQ-GMC-006 (a))

**Given:** `~/.claude.json` 에 zai-mcp-server 엔트리가 이미 존재하며, `env.Z_AI_API_KEY` 가 `~/.moai/.env.glm` 의 `GLM_AUTH_TOKEN` 과 동일.
**When:** `moai glm tools enable all` 을 실행한다.
**Then:**
- 종료 코드 0
- "이미 활성화됨" 메시지 출력
- `~/.claude.json` 변경 없음 (mtime 변경 없음 또는 동일 내용)

### GWT-11: 기존 엔트리 + 토큰 불일치 → `--force` 안내 (REQ-GMC-006 (b))

**Given:** `~/.claude.json` 에 zai-mcp-server 엔트리가 존재하며, `env.Z_AI_API_KEY` 가 `~/.moai/.env.glm` 의 `GLM_AUTH_TOKEN` 과 *다름*.
**When:** `moai glm tools enable all` 을 실행한다 (--force 없이).
**Then:**
- 종료 코드 비-0
- 차이 요약 출력 (예: "기존: TOKEN_A...XYZ, 새 토큰: TOKEN_B...ABC")
- "덮어쓰려면 --force 사용" 안내 출력
- `~/.claude.json` 변경 없음 (사용자 데이터 보호)

### GWT-12: 토큰 부재 → enable 거부 (REQ-GMC-007)

**Given:** `~/.moai/.env.glm` 파일이 존재하지 않거나 `GLM_AUTH_TOKEN` 이 빈 문자열.
**When:** `moai glm tools enable vision` 을 실행한다.
**Then:**
- 종료 코드 비-0
- "GLM 토큰 미설정" 에러 메시지
- 토큰 등록 가이드 (`moai github auth glm <token>` 또는 동등한 명령) 출력
- `~/.claude.json` 변경 없음 (어떤 엔트리도 추가/수정되지 않음)

### GWT-13: `--scope project` 옵션 (REQ-GMC-008)

**Given:** Node.js + 토큰 정상, 프로젝트 루트 (`pwd`) 에 `.mcp.json` 부재.
**When:** `moai glm tools enable vision --scope project` 을 실행한다.
**Then:**
- 프로젝트 루트의 `.mcp.json` 에 `mcpServers.zai-mcp-server` 엔트리가 생성됨 (GWT-4 와 동일한 4개 필드 포함)
- `~/.claude.json` 의 mcpServers 는 *변경되지 않음* (user scope 미터치)

### GWT-14: Node.js 부재 (REQ-GMC-009)

**Given:** PATH 에 `node` 명령이 없음 (`exec.LookPath("node")` 가 에러 반환).
**When:** `moai glm tools enable all` 을 실행한다.
**Then:**
- 종료 코드 비-0
- "Node.js not found" 에러 메시지
- 최소 요구 버전 (`>= v22.0.0`) 안내
- 설치 가이드 (예: `https://nodejs.org/` 또는 nvm 권장) 출력
- `~/.claude.json` 변경 없음

### GWT-15: Node.js 구버전 (REQ-GMC-009)

**Given:** PATH 에 `node` 가 존재하나 버전이 `v18.20.4` (< v22.0.0).
**When:** `moai glm tools enable vision` 을 실행한다.
**Then:**
- 종료 코드 비-0
- 감지된 버전 (`v18.20.4`) + 최소 요구 버전 (`>= v22.0.0`) 출력
- 업그레이드 가이드 출력
- `~/.claude.json` 변경 없음

### GWT-16: 사용자 정의 엔트리 보존 — enable (REQ-GMC-010)

**Given:** `~/.claude.json` 의 mcpServers 에 사용자 정의 엔트리 `my-custom-server` (사용자가 직접 추가) + `context7`, `sequential-thinking` 가 존재. zai-mcp-server 부재.
**When:** `moai glm tools enable vision` 을 실행한다.
**Then:**
- mcpServers 객체에 4개 엔트리: `my-custom-server`, `context7`, `sequential-thinking`, `zai-mcp-server`
- `my-custom-server`, `context7`, `sequential-thinking` 의 모든 필드 (`command`, `args`, `env`, etc.) 가 변경 없이 *완전히* 보존됨
- 새 엔트리 `zai-mcp-server` 만 추가됨

### GWT-17: 사용자 정의 엔트리 보존 — disable (REQ-GMC-010)

**Given:** `~/.claude.json` 의 mcpServers 에 `my-custom-server`, `context7`, `zai-mcp-server` 3개 존재.
**When:** `moai glm tools disable all` 을 실행한다.
**Then:**
- mcpServers 객체에 정확히 2개 엔트리: `my-custom-server`, `context7`
- `my-custom-server` 와 `context7` 의 모든 필드가 변경 없이 보존됨
- `zai-mcp-server` 만 제거됨

## Edge Cases (자동화 테스트 권장)

### GWT-18: 모드 전환 후 zai-mcp-server 엔트리 유지 (R5 검증)

**Given:** `moai glm tools enable all` 으로 zai-mcp-server 활성화된 상태.
**When:** `moai cc` 로 모드 전환 → 작업 → `moai glm` 으로 재전환.
**Then:** `~/.claude.json` 의 mcpServers.zai-mcp-server 엔트리가 모든 모드 전환 후에도 *그대로 유지됨* (사용자 명시 `disable` 까지 유지). EX-4 정책 검증.

### GWT-19: atomic write 실패 시 복구 (R7 검증)

**Given:** `~/.claude.json` 이 존재하고, 디스크 풀 또는 권한 문제로 atomic write 가 실패하는 시뮬레이션 (테스트에서는 mock).
**When:** `moai glm tools enable vision` 을 실행한다.
**Then:**
- 종료 코드 비-0
- 명확한 에러 메시지 출력 (디스크 풀 또는 권한 문제)
- `~/.claude.json` 의 *원본 내용* 이 손상 없이 보존됨 (atomic rename 의 atomicity 보장)
- 백업 파일 (`~/.claude.json.bak-*`) 이 존재하면 사용자에게 복구 가이드 출력

### GWT-20: JSON 파싱 유효성

**Given:** `moai glm tools enable all` 실행 후 결과 `~/.claude.json`.
**When:** 결과 파일을 `encoding/json` 으로 디코드한다.
**Then:** 파싱이 에러 없이 성공. JSON 문법 (trailing comma, 따옴표 짝, 키 중복 없음) 이 유효.

### GWT-21: locale 출력 언어 (i18n)

**Given:** `.moai/config/sections/language.yaml` 의 `conversation_language: ko` (현재 프로젝트 설정).
**When:** `moai glm tools enable vision` 을 실행하여 stdout 메시지를 캡처한다.
**Then:**
- 사용자 안내 메시지 (`"Vision 활성화됨"`, `"Pro 플랜 권장"`, `"Claude Code 재시작 필요"`) 가 한국어로 출력됨
- 에러 메시지 (errno, exec error 등) 는 영어 그대로 (CLAUDE.local.md §language policy 준수)

### GWT-22: 명령 인자 검증

**Given:** 사용자가 잘못된 인자로 명령 실행.
**When:**
- (a) `moai glm tools enable foo` (도구명 오타)
- (b) `moai glm tools enable` (인자 없음)
- (c) `moai glm tools enable vision websearch` (옵션 2개 동시)
**Then:**
- (a): "알 수 없는 도구명: foo. 지원: vision, websearch, webreader, all" 에러 + 사용법 출력
- (b): 사용법 출력 + 종료 코드 비-0 (또는 기본값으로 `all` 적용 — 결정은 M3 에서, 어느 쪽이든 일관)
- (c): "한 번에 하나의 도구명만 또는 'all' 사용" 에러 + 사용법 출력 (또는 v0.1 에서는 모두 enable 처리 — 결정은 M3)

## 성능 게이트

| 메트릭 | 기준 | 측정 방법 |
|--------|------|-----------|
| `moai glm tools enable` 명령 지연 | < 5초 (npx 첫 다운로드 제외) | 테스트에서 `time` 측정 |
| `moai glm tools disable` 명령 지연 | < 1초 | 테스트에서 `time` 측정 |
| `~/.claude.json` 파일 크기 증가 | < 500 bytes 추가 (단일 zai-mcp-server 엔트리) | 백업 파일과 신규 파일 크기 비교 |
| atomic write race window | 0 (POSIX rename atomicity 보장) | OS 가 보장, 테스트 불요 |

> **Note**: npx 첫 다운로드 (~5MB `@z_ai/mcp-server` 패키지) 는 사용자가 실제로 도구를 *호출* 할 때 발생하며, `moai glm tools enable` 자체에서는 발생하지 않음 (Claude Code 의 lazy MCP loading 정책 + SPEC-CC2122-MCP-001 의 alwaysLoad=false 기본값에 의해).

## 품질 게이트 기준

본 SPEC 의 인수 조건은 다음을 모두 충족해야 한다:

- [ ] **EARS 준수**: spec.md 의 10개 REQ-GMC-001 ~ 010 이 모두 EARS 패턴으로 표현됨 (Ubiquitous 2 + Event-Driven 4 + State-Driven 1 + Optional 1 + Unwanted 2)
- [ ] **GWT-1 ~ GWT-22**: 모든 시나리오가 자동화된 테스트로 통과 (manual 검증 가능 케이스는 수동 + 가이드)
- [ ] **REQ ↔ GWT 매핑**: 본 acceptance.md 상단 매핑 표의 모든 REQ 가 ≥ 1개 GWT 로 검증됨
- [ ] **하위 호환**: GWT-3 (SPEC-GLM-001 호환성) + GWT-7, GWT-16, GWT-17 (다른 mcpServers 엔트리 보존) 통과
- [ ] **회귀 없음**: 기존 `internal/cli/glm_test.go` 의 모든 테스트 + 다른 패키지 테스트 통과 (`go test ./...`)
- [ ] **race-free**: `go test -race ./internal/cli/...` 통과
- [ ] **vet/lint**: `go vet ./...` 무경고, `golangci-lint run` 무경고 (가능한 경우)
- [ ] **빌드**: `make build` 성공으로 `internal/template/embedded.go` 가 최신 템플릿 반영
- [ ] **테스트 커버리지**: 신규 패키지 (`internal/cli/glm_tools.go` 또는 `internal/integration/zaimcp/`) ≥ 85% 라인 커버리지 (CLAUDE.local.md §6 Coverage Targets)
- [ ] **문서 동기화**: 두 `settings-management.md` 파일 (템플릿 + 루트 미러본) 에 Z.AI MCP 노트 부착, 두 사본 동일 내용 (Template-First Rule, CLAUDE.local.md §2)
- [ ] **CHANGELOG.md** 에 v0.1 항목 추가
- [ ] **16개 언어 중립성**: CLAUDE.local.md §15 verification 통과

## Definition of Done

다음 조건이 모두 만족되면 SPEC-GLM-MCP-001 은 완료(`status: completed`)로 간주한다:

1. `moai glm tools enable [vision|websearch|webreader|all]` subcommand 가 작동하며 REQ-GMC-001 ~ 010 을 모두 충족.
2. `moai glm tools disable [vision|websearch|webreader|all]` subcommand 가 작동하며 다른 mcpServers 엔트리를 손상하지 않음.
3. `~/.moai/.env.glm` 의 `GLM_AUTH_TOKEN` 이 자동 재사용되어 사용자가 토큰을 재입력하지 않음.
4. Node.js < v22 환경에서 graceful fail + `~/.claude.json` 무변경.
5. 기존 zai-mcp-server 엔트리가 사용자 정의 토큰을 가질 때 `--force` 안내로 데이터 보호.
6. GWT-1 ~ GWT-22 시나리오가 모두 자동화된 테스트로 작성·통과됨 (수동 검증 케이스는 명시적으로 수동임을 표시).
7. `go test ./...` 전체 통과, `go vet ./...` 무경고, `go test -race ./internal/cli/...` 통과.
8. `make build` 이후 `internal/template/embedded.go` 가 최신 템플릿 반영.
9. `internal/template/templates/.claude/rules/moai/core/settings-management.md` 와 `.claude/rules/moai/core/settings-management.md` 두 파일에 Z.AI MCP 통합 노트 + `moai glm tools` 사용 가이드가 동일하게 부착됨.
10. CHANGELOG.md 에 v0.1 항목 추가.
11. PR 이 main 에 머지되어 `merged_pr` / `merged_commit` 정보가 spec.md frontmatter 에 기록됨.
12. plan-auditor 의 사후 감사(post-merge audit)에서 통과 판정.
13. 사용자 수동 통합 검증 (가능한 경우): `moai glm tools enable all` → Claude Code 재시작 → Vision MCP 도구 호출 → `moai glm tools disable all` → 엔트리 제거 확인.

## 검증 방법 요약

| 시나리오 그룹 | 검증 도구 | 자동화 여부 |
|--------------|-----------|-----------|
| GWT-1 ~ GWT-2 (subcommand idempotent) | `go test ./internal/cli/...` (`TestGLMToolsEnableIdempotent`, `TestGLMToolsDisableIdempotent`) | 자동화 |
| GWT-3 (SPEC-GLM-001 호환) | 기존 GLM 회귀 테스트 + 신규 sub-test | 자동화 |
| GWT-4 ~ GWT-5 (enable 정확성) | 신규 테스트 + JSON 디코드 후 필드 검증 | 자동화 |
| GWT-6 ~ GWT-7 (disable + 보존) | 신규 테스트 + 다중 엔트리 fixture | 자동화 |
| GWT-8 ~ GWT-9 (백업 정책) | 신규 테스트 + `t.TempDir()` 격리 | 자동화 |
| GWT-10 ~ GWT-11 (기존 엔트리 처리) | 신규 테스트 + 토큰 mock | 자동화 |
| GWT-12 (토큰 부재) | 신규 테스트 + `~/.moai/.env.glm` mock | 자동화 |
| GWT-13 (`--scope project`) | 신규 테스트 + 프로젝트 루트 mock | 자동화 |
| GWT-14 ~ GWT-15 (Node.js 부재/구버전) | 신규 테스트 + PATH mock + 버전 출력 mock | 자동화 |
| GWT-16 ~ GWT-17 (사용자 정의 엔트리 보존) | 신규 테스트 + 다중 mcpServers fixture | 자동화 |
| GWT-18 (모드 전환) | 통합 테스트 또는 수동 검증 (CLAUDE.local.md §13 GLM Integration Testing 정책 — dev 프로젝트에서 실행 금지, `/tmp/test-project` 사용) | 부분 자동화 |
| GWT-19 (atomic write 실패) | 신규 테스트 + write error mock | 자동화 |
| GWT-20 (JSON 파싱 유효성) | 신규 테스트 + `encoding/json` 검증 | 자동화 |
| GWT-21 (locale 출력) | 신규 테스트 + language.yaml mock | 자동화 |
| GWT-22 (인자 검증) | 신규 테스트 + 명령 인자 변형 | 자동화 |
| Definition of Done #11, #12 | manager-git PR 머지 후 plan-auditor 호출 | 수동 트리거 (자동 검증) |
| Definition of Done #13 | 사용자 수동 검증 (실제 Z.AI 토큰 + 실제 Claude Code 세션 필요) | 수동 |
