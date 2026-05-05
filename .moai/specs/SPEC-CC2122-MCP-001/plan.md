# SPEC-CC2122-MCP-001 구현 계획

## 개요

본 계획은 Claude Code v2.1.119 의 `alwaysLoad: true` MCP 필드를 `internal/template/templates/.mcp.json.tmpl` 의 `context7` 와 `sequential-thinking` 엔트리에 통합하는 작업의 마일스톤, 기술적 접근, 리스크, 의존성을 정의한다.

이 문서는 시간 추정(예: "2 days") 대신 **우선순위 라벨** 과 **단계 간 순서** 로 진행을 표현한다 (CLAUDE.local.md §14 + agent-common-protocol.md §Time Estimation 준수).

## 기술적 접근 (High-Level)

`.mcp.json.tmpl` 은 Go template 형식의 단순 JSON 파일이다. 변경은 두 mcpServers 엔트리(`context7`, `sequential-thinking`)에 `"alwaysLoad": true` 한 줄을 추가하는 작은 보조-필드 수정이다. 기존의 `command`, `args`, `timeout`, `$comment`, 플랫폼 분기 (`{{- if eq .Platform "windows"}}`) 는 절대 손대지 않는다.

테스트 측면에서는 기존 `internal/template/settings_test.go` 에 이미 `.mcp.json.tmpl` 렌더링 검증 패턴이 존재할 가능성이 높으므로, M1 에서 그 패턴을 식별하고 동일 컨벤션(테이블 드리븐 또는 sub-test 구조)을 따른다. 신규 테스트 함수는 `TestMCPTemplateAlwaysLoad*` prefix 로 grep/search 가능하게 명명한다.

문서 동기화는 두 파일 (`settings-management.md` 루트 + 템플릿 미러본) 에 동일한 짧은 노트를 추가하는 단순 미러 작업이며, Template-First Rule (CLAUDE.local.md §2) 에 따라 템플릿 본을 먼저 수정하고 루트 사본을 동기화하거나 동시에 작성한다.

## Milestones

### M1 — Path Discovery + Existing Test Pattern Review (Priority: High)

**목표**: 변경 대상 파일의 정확한 라인 범위와 `settings_test.go` 의 기존 `.mcp.json.tmpl` 검증 패턴을 식별한다.

**작업 항목:**

- `internal/template/templates/.mcp.json.tmpl` 의 현재 3개 엔트리(`context7`, `sequential-thinking`, `moai-lsp`) 의 정확한 라인 범위 확인
- `internal/template/settings_test.go` 내 `.mcp.json` / `.mcp.json.tmpl` 관련 기존 테스트 함수 확인 (Grep `mcp` 또는 `MCP` 케이스)
- 기존 테스트가 사용하는 패턴 식별: 테이블 드리븐 vs sub-test, JSON 파싱 후 필드 검증 vs 문자열 매칭
- 기존 테스트 컨벤션의 helper / fixture (예: 템플릿 렌더링 헬퍼, JSON 디코드 헬퍼) 위치 파악
- `.claude/rules/moai/core/settings-management.md` 와 그 템플릿 미러본의 현재 구조 (섹션 헤더, MCP 관련 기존 설명) 확인하여 노트 삽입 지점 선정

**Done 기준:**

- 수정 대상 4개 파일 모두 정확한 라인 범위가 식별되어 PR 설명에 인용 가능
- 신규 테스트가 따라야 할 기존 컨벤션이 명확히 결정됨 (예: "table-driven with `tests := []struct{...}` pattern")

### M2 — Template Update (Priority: High)

**목표**: `.mcp.json.tmpl` 의 `context7` 와 `sequential-thinking` 두 엔트리에 `"alwaysLoad": true` 를 추가한다. `moai-lsp` 는 변경하지 않는다.

**작업 항목:**

- `context7` 엔트리에 `"alwaysLoad": true` 한 줄 추가 (`$comment` 라인 바로 뒤 또는 엔트리 마지막 — 기존 JSON 의 키 순서 컨벤션을 따름)
- `sequential-thinking` 엔트리에 동일하게 한 줄 추가
- `moai-lsp` 엔트리는 손대지 않음 (REQ-002 명시적 제외)
- JSON 문법 유효성 확인 (trailing comma 위치, 따옴표 짝)
- 플랫폼 분기 (`{{- if eq .Platform "windows"}}` ... `{{- else}}` ... `{{- end}}`) 는 그대로 유지 (REQ-003 회귀 방지)
- `make build` 로 `internal/template/embedded.go` 재생성

**Done 기준:**

- `internal/template/templates/.mcp.json.tmpl` 변경분이 `context7` 와 `sequential-thinking` 두 엔트리에만 국한됨 (`git diff` 로 검증 가능)
- `make build` 성공, `internal/template/embedded.go` 재생성됨
- `go build ./internal/template/...` 통과

### M3 — Test Coverage (TDD: RED → GREEN) (Priority: High)

**목표**: `internal/template/settings_test.go` 에 REQ-001/002/003 을 각각 검증하는 신규 테스트 케이스 최소 3개를 TDD 사이클(RED → GREEN)로 추가한다.

**작업 항목:**

- M1 에서 식별한 기존 컨벤션을 따라 신규 함수 추가:
  - `TestMCPTemplateAlwaysLoadOnContext7` (또는 통합 함수의 sub-test) — REQ-001 (a) 검증: 렌더링된 `.mcp.json` 의 `context7` 엔트리에 `alwaysLoad: true` 존재
  - `TestMCPTemplateAlwaysLoadOnSequentialThinking` — REQ-001 (b) 검증: `sequential-thinking` 엔트리에 `alwaysLoad: true` 존재
  - `TestMCPTemplateAlwaysLoadAbsentOnMoaiLSP` — REQ-002 검증: `moai-lsp` 엔트리에 `alwaysLoad` 키가 부재하거나 `false`
  - (선택 통합) `TestMCPTemplateExistingFieldsPreserved` — REQ-003 검증: 3개 엔트리 모두에서 `command`, `args`, `timeout` (해당 시), `$comment` 가 보존됨
- TDD 순서: 테스트를 먼저 작성하고 (M2 작업 전에는 RED), M2 적용 후 GREEN 확인. 본 plan 에서는 M2 가 M3 보다 먼저 기재되어 있으나, 실제 구현자는 자유롭게 RED-first 로 순서를 뒤집어도 된다.
- 테스트는 두 플랫폼 분기 (windows / non-windows) 모두를 커버 (M1 에서 발견한 기존 패턴이 plat-aware 이면 따름)
- `go test -count=1 ./internal/template/...` 통과 확인 (캐시 disable, CLAUDE.local.md §6 Go Test Execution Rules)

**Done 기준:**

- 신규 테스트 함수 ≥ 3개가 `TestMCPTemplateAlwaysLoad*` 또는 `TestMCPTemplateExistingFieldsPreserved` 명명으로 추가됨
- 각 테스트가 REQ-001 / REQ-002 / REQ-003 중 하나 이상에 매핑됨 (acceptance.md 검증 표 참조)
- `go test ./internal/template/...` 전체 통과 (신규 + 기존, 회귀 없음)
- `go test -race ./internal/template/...` 통과

### M4 — Documentation Sync + Final Validation (Priority: High)

**목표**: `settings-management.md` 두 파일 (루트 + 템플릿 미러본) 에 v2.1.119 `alwaysLoad` 노트를 추가하고, 전체 품질 게이트를 통과시킨다.

**작업 항목:**

- 두 `settings-management.md` 파일에 짧은 섹션/노트 블록 추가 (REQ-004):
  - 노트 내용: (a) `alwaysLoad: true` 의 pre-load 의미, (b) `false`/부재 시 lazy-load 하위 호환, (c) 권장 기준 — 소형 + 빈번 → `true`, 대형 + 비상시 → `false`/부재
  - moai-adk 의 현재 사례 ("`context7`, `sequential-thinking` 은 `true`, `moai-lsp` 는 부재") 를 한 줄 예시로 포함하여 독자 이해를 돕는다
  - 노트의 길이는 5-10 문장 정도로 제한 (문서 비대화 방지, Karpathy Simplicity First 준수)
- Template-First Rule (CLAUDE.local.md §2) 준수: `internal/template/templates/.claude/rules/moai/core/settings-management.md` 를 먼저 또는 동시에 수정
- 두 파일의 변경분이 동일한지 `diff` 로 검증
- 최종 게이트:
  - `go test ./...` (전체) 통과 — cascading failure 검출 (CLAUDE.local.md §6)
  - `go vet ./...` 통과
  - `golangci-lint run` 무경고 (가능한 경우)
  - `make build` 성공 (embedded.go 최신)
- 커밋 메시지 초안: `feat(mcp): add alwaysLoad: true to context7 + sequential-thinking MCP entries (SPEC-CC2122-MCP-001)`

**Done 기준:**

- 두 `settings-management.md` 파일에 v2.1.119 `alwaysLoad` 노트 부착됨, 두 사본이 동일 내용
- `make build` 후 `git status` 에 `internal/template/embedded.go` 변경 포함 (자동 재생성 결과)
- 모든 게이트 (vet/test/lint/build) 통과
- PR 설명에 SPEC-CC2122-MCP-001 명시, REQ-001 ~ REQ-005 와 GWT-1 ~ GWT-5 의 매핑 표 포함

## Dependencies

- **상위 의존**: 없음. 본 SPEC 은 v2.1.119 호환 통합 시리즈의 일부로 SPEC-CC2122-HOOK-001 (이미 main 적용) 및 SPEC-CC2122-STATUSLINE-001 (별도 진행) 과 병렬 관계이며, 코드 수준의 직접 의존은 없다.
- **하위 의존 후보**: 향후 사용자가 추가하는 자체 MCP 서버에 `alwaysLoad` 자동 권장 로직(예: `moai doctor` 가 사용 빈도 분석 후 제안)을 도입한다면 본 SPEC 의 결정 표(소형/빈번 vs 대형/비상시)를 입력으로 사용할 수 있다. 그러나 그 작업은 별도 SPEC 이다.
- **외부 의존**: Claude Code v2.1.119 이상에서 `.mcp.json` 의 `alwaysLoad` 필드를 실제로 인식하는지 — 릴리스 노트로 확인됨. 통합 테스트는 사용자 환경에서 수동으로 1회 검증할 수 있으나, 본 SPEC 의 자동화된 테스트는 템플릿 렌더링 수준에서만 검증한다 (런타임 동작 검증은 Claude Code 의 책임).

## Risks and Mitigations

| 리스크 | 영향도 | 완화 전략 |
|--------|--------|-----------|
| `moai-lsp` 엔트리에 실수로 `alwaysLoad: true` 가 추가됨 | High | M2 의 `git diff` 검토 + M3 의 `TestMCPTemplateAlwaysLoadAbsentOnMoaiLSP` 자동화 테스트로 회귀 즉시 검출. REQ-002 에 명시적으로 금지. |
| JSON 문법 오류 (trailing comma, 따옴표 누락) 로 `.mcp.json` 파싱 실패 | High | M2 후 `go test` 의 템플릿 렌더링 + JSON 파싱 테스트가 통과해야 함. 파싱 실패는 즉시 RED 가 됨. |
| 플랫폼 분기 (`{{- if eq .Platform "windows"}}`) 의 위치/구조가 깨짐 | Medium | REQ-003 회귀 검증 + M3 의 `TestMCPTemplateExistingFieldsPreserved` 가 윈도우/유닉스 양쪽 분기를 모두 검증. |
| 두 `settings-management.md` 파일의 내용이 어긋남 (Template-First 위반) | Medium | M4 에서 `diff` 로 두 파일 동일성 검증. CLAUDE.local.md §2 의 verification 체크리스트 준수. |
| `make build` 누락으로 `internal/template/embedded.go` 가 stale | High | M2 와 M4 모두에서 `make build` 명시적 실행. PR 체크리스트의 "Templates regenerated" 항목 준수 (CLAUDE.local.md §4). |
| 기존 사용자가 `moai update` 실행 시 `.mcp.json` 머지 충돌 | Low | 본 SPEC 범위 밖. `moai update` 의 기존 머지 정책에 위임. 사용자가 직접 수정한 `.mcp.json` 은 보호됨 (CLAUDE.local.md §2 Protected Directories 정책 준수). |
| 템플릿 언어 중립성 위반 (특정 언어 편향 도입) | Low | 본 SPEC 은 JSON 한 줄 추가 + 영어 노트 한 단락이며, 16개 언어 사용자 모두에게 동등하게 적용됨. CLAUDE.local.md §15 위반 없음. |

## Out of Scope (Plan-Level)

- `moai-lsp` 의 `alwaysLoad` 정책 변경 — 본 SPEC 은 `false`/부재 유지를 명시 (REQ-002).
- `staggeredStartup` 블록 (`enabled`, `delayMs`, `connectionTimeout`) 변경 — 별도 SPEC 분리.
- 사용자 정의 MCP 서버에 대한 자동 `alwaysLoad` 추천/주입 — 별도 SPEC 후보.
- `.claude/settings.json` 의 MCP 관련 키 (예: `enabledMcpjsonServers`) 수정 — 본 SPEC 범위 밖.
- `.mcp.json` 스키마 강화 검증 (boolean type check 등) — 별도 SPEC 후보.
- 기존 사용자의 `.mcp.json` 자동 마이그레이션 도구 — 본 SPEC 범위 밖.
- 4개국어 docs-site (`docs-site/`) 동기화 — 본 SPEC 의 변경은 내부 룰 문서 (`.claude/rules/moai/core/settings-management.md`) 에 한정되며, docs-site 사용자 가이드 갱신 여부는 별도 판단 (CLAUDE.local.md §17.3 ko-canonical 정책 검토 후 필요 시 후속 PR 분리).

## Approval / Next Step

본 plan.md 와 spec.md, acceptance.md 가 plan-auditor 의 승인을 받으면, manager-cycle 또는 manager-tdd 가 `/moai run SPEC-CC2122-MCP-001` 으로 위임받아 M1 → M4 순으로 구현한다. 본 SPEC 단계에서는 코드/테스트/문서 어떤 파일도 수정하지 않는다 (사용자 명시 제약).
