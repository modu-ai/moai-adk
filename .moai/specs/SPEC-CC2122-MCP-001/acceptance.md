# SPEC-CC2122-MCP-001 인수 기준

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-30 | manager-spec | 초기 작성 — GWT-1 ~ GWT-5 시나리오 + Edge Cases + Quality Gate + Definition of Done |

본 문서는 SPEC-CC2122-MCP-001 의 구현이 완료되었음을 검증하기 위한 Given-When-Then(GWT) 시나리오와 Definition of Done 을 정의한다. 각 시나리오는 `internal/template/settings_test.go` 의 테이블 드리븐(또는 sub-test) 구조로 자동화된다.

## Given-When-Then 시나리오

### GWT-1: `context7` 엔트리에 `alwaysLoad: true` 부착 검증

**Given:** `internal/template/templates/.mcp.json.tmpl` 이 본 SPEC 의 변경 후 상태이고, 표준 템플릿 컨텍스트(예: `Platform="darwin"` 또는 `Platform="linux"`) 로 `moai init` 또는 `moai update` 가 렌더링하여 `.mcp.json` 파일을 생성했다.
**When:** 결과 `.mcp.json` 의 JSON 트리에서 `mcpServers.context7` 엔트리를 검사한다.
**Then:** `mcpServers.context7` 객체는 `"alwaysLoad": true` 키-값 쌍을 포함해야 한다. 다른 모든 기존 필드(`command`, `args`, `$comment`) 는 손상되지 않은 채 보존되어야 한다.

### GWT-2: `sequential-thinking` 엔트리에 `alwaysLoad: true` 부착 검증

**Given:** GWT-1 과 동일한 렌더링 컨텍스트.
**When:** 결과 `.mcp.json` 의 `mcpServers["sequential-thinking"]` 엔트리를 검사한다.
**Then:** `mcpServers["sequential-thinking"]` 객체는 `"alwaysLoad": true` 키-값 쌍을 포함해야 한다. `command`, `args`, `$comment` 는 변경 없이 보존되어야 한다.

### GWT-3: `moai-lsp` 엔트리에 `alwaysLoad` 부재 (또는 `false`) 검증

**Given:** GWT-1 과 동일한 렌더링 컨텍스트.
**When:** 결과 `.mcp.json` 의 `mcpServers["moai-lsp"]` 엔트리를 검사한다.
**Then:** `mcpServers["moai-lsp"]` 객체에는 `alwaysLoad` 키가 부재해야 한다. 만약 어떤 이유로 키가 존재한다면 그 값은 `false` 여야 한다 (lazy-load 정책 유지). `command`, `args`, `timeout`, `$comment` 는 변경 없이 보존되어야 한다.

### GWT-4: 기존 필드 회귀 없음 — 베이스라인 비교

**Given:** SPEC 적용 전의 `.mcp.json.tmpl` 렌더링 결과를 베이스라인 JSON 으로 보유하거나 메모리 내 fixture 로 정의한다. 베이스라인은 본 SPEC 변경 직전의 3개 mcpServers 엔트리 + `staggeredStartup` 블록 + `$schema` 필드를 포함한다.
**When:** 본 SPEC 적용 후의 렌더링 결과 JSON 트리를 베이스라인과 필드 단위로 비교한다 (단, `alwaysLoad` 의 `context7`/`sequential-thinking` 추가는 제외).
**Then:** 모든 3개 mcpServers 엔트리에서 `command`, `args`, `timeout` (존재하는 경우), `$comment` 필드가 베이스라인과 정확히 일치해야 한다. 윈도우/유닉스 플랫폼 분기 모두에서 동일하게 검증된다. `staggeredStartup` 블록과 `$schema` 필드도 변경되지 않아야 한다.

### GWT-5: 테스트 스위트 + 회귀 없음

**Given:** `internal/template/settings_test.go` 에 본 SPEC 으로 추가된 신규 테스트 케이스 (`TestMCPTemplateAlwaysLoad*` prefix 의 함수 ≥ 3개) 가 작성되어 있고, 기존 테스트 스위트도 그대로 유지된다.
**When:** 다음 명령들을 차례로 실행한다:
- `go test -count=1 ./internal/template/...`
- `go test -race ./internal/template/...`
- `go test ./...` (전체, cascading failure 검출용)
**Then:**
- 모든 명령이 `PASS` 한다.
- 신규 추가된 `TestMCPTemplateAlwaysLoad*` 케이스 ≥ 3개가 명시적으로 PASS 출력됨 (각각 REQ-001 / REQ-002 / REQ-003 에 매핑).
- 기존 `internal/template/settings_test.go` 의 모든 케이스가 회귀 없이 통과한다.
- `go test -race` 가 race detector 경고 없이 clean 통과한다.
- `go test ./...` 전체에서 다른 패키지의 회귀가 발생하지 않는다.

## Edge Cases (테스트 추가 권장)

GWT-1 ~ GWT-5 외에 구현 단계에서 명시적으로 다뤄야 할 추가 엣지 케이스:

- **윈도우 플랫폼 렌더링 (`Platform="windows"`):** `{{- if eq .Platform "windows"}}` 분기에서도 `alwaysLoad: true` 가 `context7` 와 `sequential-thinking` 두 엔트리에 동일하게 부착되어야 한다. 명령 분기 (`cmd.exe /c npx ...`) 는 변경 없이 보존된다.
- **JSON 파싱 유효성:** 렌더링 결과 문자열을 `encoding/json` 으로 디코드 시 에러 없이 성공해야 한다. 키 순서, trailing comma, 따옴표 짝 등 JSON 문법이 깨지지 않았음을 보증.
- **`alwaysLoad` 값 타입 검증:** `alwaysLoad` 는 boolean `true` 여야 하며, 문자열 `"true"` 또는 숫자 `1` 등이 되어서는 안 된다 (Claude Code 런타임 호환).
- **기존 사용자의 `.mcp.json` 보호:** 사용자가 `.mcp.json` 을 수정한 경우, `moai update` 가 그 파일을 임의로 덮어쓰지 않아야 한다 (Protected Directories 정책, CLAUDE.local.md §2). 이는 본 SPEC 의 직접 검증 대상은 아니나, 변경이 사용자 데이터를 손상시키지 않는다는 일반 원칙을 침해해서는 안 된다.
- **두 `settings-management.md` 파일 동일성:** 루트 (`.claude/rules/moai/core/settings-management.md`) 와 템플릿 미러본 (`internal/template/templates/.claude/rules/moai/core/settings-management.md`) 의 v2.1.119 노트 부분이 정확히 동일한 텍스트인지 `diff` 로 검증.

## Quality Gate Criteria

본 SPEC 의 인수 조건은 다음을 모두 충족해야 한다:

- [ ] **EARS 준수**: spec.md 의 5개 REQ 가 모두 EARS 패턴(WHEN/IF/Ubiquitous) 으로 표현됨
- [ ] **GWT-1 ~ GWT-5**: 5개 시나리오 모두 자동화된 테스트로 통과
- [ ] **REQ-매핑**: REQ-001 → GWT-1 + GWT-2 + 신규 테스트 (a)(b); REQ-002 → GWT-3 + 신규 테스트 (c); REQ-003 → GWT-4 + 회귀 검증; REQ-004 → 두 `settings-management.md` 파일 노트 부착 + diff 검증; REQ-005 → 신규 테스트 함수 ≥ 3개 존재 + GWT-5 통과
- [ ] **하위 호환**: GWT-3 통과로 `moai-lsp` lazy-load 유지, 그리고 v2.1.119 미만 사용자에 대해서는 `alwaysLoad` 가 단순 무시되므로 자연스럽게 호환됨 (별도 검증 불필요)
- [ ] **회귀 없음**: GWT-4, GWT-5 통과로 기존 필드 보존 + 다른 패키지 회귀 없음 보장
- [ ] **race-free**: `go test -race ./internal/template/...` 통과
- [ ] **vet/lint**: `go vet ./...` 무경고, `golangci-lint run` 무경고 (가능한 경우)
- [ ] **빌드**: `make build` 성공으로 `internal/template/embedded.go` 가 최신 템플릿을 반영
- [ ] **문서 동기화**: 두 `settings-management.md` 파일에 v2.1.119 `alwaysLoad` 노트 부착, 두 사본 동일 (REQ-004)
- [ ] **Template-First Rule 준수**: CLAUDE.local.md §2 의 verification (모든 신규 `.claude/`/`.moai/` 변경에 대응되는 `internal/template/templates/` 변경 존재) 통과

## Definition of Done

다음 조건이 모두 만족되면 SPEC-CC2122-MCP-001 은 완료(`status: completed`)로 간주한다:

1. `internal/template/templates/.mcp.json.tmpl` 의 `context7` 와 `sequential-thinking` 엔트리에 `"alwaysLoad": true` 가 추가되었고, `moai-lsp` 엔트리는 변경되지 않았다.
2. `internal/template/settings_test.go` 에 `TestMCPTemplateAlwaysLoad*` (또는 동등 명명의) 신규 테스트 함수 ≥ 3개가 추가되어 REQ-001 / REQ-002 / REQ-003 을 각각 검증한다.
3. 위 GWT-1 ~ GWT-5 시나리오가 모두 자동화된 테스트로 작성·통과됨.
4. `go test ./...` 전체 통과, `go vet ./...` 무경고, `go test -race ./internal/template/...` 통과.
5. `make build` 이후 `internal/template/embedded.go` 가 최신 템플릿을 반영한다.
6. `.claude/rules/moai/core/settings-management.md` 와 `internal/template/templates/.claude/rules/moai/core/settings-management.md` 두 파일에 v2.1.119 `alwaysLoad` 의미 + 사용 권장 기준 노트가 동일하게 부착됨 (REQ-004).
7. PR 이 main 에 머지되어 `merged_pr` / `merged_commit` 정보가 spec.md 의 frontmatter 에 기록됨.
8. plan-auditor 의 사후 감사(post-merge audit)에서 통과 판정.

## 검증 방법 요약

| 시나리오 | 검증 도구 | 자동화 여부 |
|---------|-----------|-----------|
| GWT-1, GWT-2 | `go test ./internal/template/...` (`TestMCPTemplateAlwaysLoadOnContext7`, `TestMCPTemplateAlwaysLoadOnSequentialThinking` 또는 통합 테스트의 sub-tests) | 자동화 |
| GWT-3 | `go test ./internal/template/...` (`TestMCPTemplateAlwaysLoadAbsentOnMoaiLSP`) | 자동화 |
| GWT-4 | `go test ./internal/template/...` (`TestMCPTemplateExistingFieldsPreserved` — 베이스라인 비교 또는 필드별 명시적 검증) | 자동화 |
| GWT-5 | `go test -count=1 -race ./internal/template/...` + `go test ./...` | 자동화 |
| Edge cases (윈도우 분기, JSON 파싱) | 위와 동일 테스트 파일에 추가 | 자동화 |
| REQ-004 (두 `settings-management.md` 동일성) | `diff .claude/rules/moai/core/settings-management.md internal/template/templates/.claude/rules/moai/core/settings-management.md` 의 v2.1.119 노트 섹션 비교 | 수동 또는 CI 스크립트로 자동화 가능 |
| Definition of Done #7, #8 | manager-git PR 머지 후 plan-auditor 호출 | 수동 트리거 (자동 검증) |

## REQ ↔ GWT 매핑 표

| REQ | 매핑 GWT | 매핑 신규 테스트 케이스 (예시 명명) |
|-----|---------|--------------------------------|
| REQ-001 (context7 + sequential-thinking 에 `alwaysLoad: true`) | GWT-1, GWT-2 | `TestMCPTemplateAlwaysLoadOnContext7`, `TestMCPTemplateAlwaysLoadOnSequentialThinking` |
| REQ-002 (moai-lsp 에 `alwaysLoad` 부재 또는 false) | GWT-3 | `TestMCPTemplateAlwaysLoadAbsentOnMoaiLSP` |
| REQ-003 (기존 필드 회귀 없음) | GWT-4 | `TestMCPTemplateExistingFieldsPreserved` (또는 GWT-1/2/3 의 회귀 검증 통합) |
| REQ-004 (두 `settings-management.md` 노트) | (테스트 외) 수동 또는 CI 스크립트 | `diff` 비교 + reviewer 시각 검증 |
| REQ-005 (테스트 ≥ 3개 존재 + 통과) | GWT-5 | 위 신규 테스트 3개 모두 + `go test ./...` 통과 |
