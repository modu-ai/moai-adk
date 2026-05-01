---
id: SPEC-GLM-MCP-001
artifact: plan
version: "0.1.0"
created_at: 2026-05-01
parent_spec: spec.md
methodology: TDD (per quality.yaml development_mode default)
---

# SPEC-GLM-MCP-001 Implementation Plan

## Status

- Plan-audit verdict: **PASS** (score 0.91)
- Plan-audit report: `.moai/reports/plan-audit/SPEC-GLM-MCP-001-review-1.md`
- Methodology: **TDD (RED-GREEN-REFACTOR)** (≥85% coverage per commit, per CLAUDE.local.md §6)

## Milestone Overview

세 마일스톤으로 분할 (각 마일스톤은 독립 PR 후보):

| Milestone | 범위 | 우선순위 | 대표 REQ | 의존성 |
|-----------|------|----------|----------|--------|
| M1 | `moai glm tools enable\|disable\|status` 기본 명령 + registrar 핵심 로직 | High | REQ-GMC-001/002/007 | 없음 |
| M2 | 자격증명 재사용, Node.js 검사, 충돌 처리, scope 옵션 | High | REQ-GMC-003/004/005/008 | M1 |
| M3 | Pro 플랜 검사, 모드 전환 무영향 검증, 비대화형 + 원자적 쓰기 | Medium | REQ-GMC-006/009/010 | M1, M2 |

각 마일스톤은 RED → GREEN → REFACTOR 사이클로 진행되며, 마일스톤 종료 시 `go test ./...` (race + coverage) 와 `golangci-lint run` 이 모두 green 이어야 한다.

## M1 — Core Subcommand Skeleton

**목표**: `moai glm tools` 서브커맨드 트리 + registrar 의 read/write 핵심 로직.

### M1 Tasks

1. **TDD-RED**: `internal/cli/cmd/glm_tools_test.go` 에 GWT-01, GWT-04, GWT-16 (status 출력 정확성) 테스트 작성. 실패 확인.
2. **TDD-GREEN**: `internal/cli/cmd/glm_tools.go` 에 cobra 서브커맨드 (`enable`, `disable`, `status`) 추가. `internal/glm/mcp/registrar.go` 에 settings 의 `mcpServers` 키 추가/제거/조회 함수 작성. 정규 키 3종을 단일 진실 공급원으로 정의 (REQ-GMC-001).
3. **TDD-REFACTOR**: 중복 제거, 키 상수 추출, 에러 wrap (`fmt.Errorf("...: %w", err)`).
4. **검증**: M1 종료 시 `go test ./internal/cli/... ./internal/glm/...` line coverage ≥85%, `golangci-lint run` green.

### M1 Output

- 새 파일: `internal/cli/cmd/glm_tools.go`, `internal/cli/cmd/glm_tools_test.go`, `internal/glm/mcp/registrar.go`, `internal/glm/mcp/registrar_test.go`
- 수정 파일: `internal/cli/cmd/glm.go` (서브커맨드 라우팅 1줄 추가)

## M2 — Credentials, Node Check, Conflicts, Scope

**목표**: M1 위에 R1 (충돌), R3 (Node 부재), R4 (scope) mitigation 추가.

### M2 Tasks

1. **REQ-GMC-003 — 자격증명 재사용** (R 무관, 신규 동작)
   - GWT-07, GWT-08 RED → GREEN.
   - `~/.moai/.env.glm` 파서 재사용 (SPEC-GLM-001 의 기존 함수 호출).
   - D1 정제: Z.AI 패키지가 요구하는 정확한 환경변수명 목록 확정. `@z_ai/mcp-server` README 또는 source code 확인 (Run phase 첫 작업).

2. **REQ-GMC-004 — Node.js 부재 graceful (R3 mitigation)**
   - GWT-09, GWT-10 RED → GREEN.
   - `exec.LookPath("node")`, `exec.LookPath("npx")` 양쪽 검사. 메시지 3요소(부재 사실, 권장 버전, 설치 가이드 URL) 포함.
   - 인터페이스 주입으로 테스트 가능하게 만든다 (D3 mock injection mechanism — `type LookPathFunc func(string) (string, error)`).

3. **REQ-GMC-005 — 충돌 처리 (R1 mitigation)**
   - GWT-11, GWT-12, GWT-13 RED → GREEN.
   - 정규 키 (`glm-vision`/`glm-websearch`/`glm-webreader`) 만 충돌 검사 대상. 비표준 키는 건드리지 않음 (EX-4).
   - 대화형 옵션 3종 (보존 / 백업 후 덮어쓰기 / 취소) — 백업 키 형식: `<key>.bak.<unix-epoch>` (D2 에서 timestamp 형식 확정 권장).

4. **REQ-GMC-008 — scope 옵션 (R4 결정 반영)**
   - GWT-17, GWT-18 RED → GREEN.
   - `--scope user|project` 플래그. 기본값 `user`. project scope 는 프로젝트 루트의 `.mcp.json` 에 작성.

### M2 Output

- M1 파일들 확장 + 신규 helper (`internal/glm/mcp/credentials.go`, `internal/glm/mcp/node_check.go`).
- M2 종료 시 GWT-01 ~ GWT-13, GWT-16 ~ GWT-18 통과 (총 16/22 GWT 커버).

## M3 — Pro Plan, Mode Switch, Non-Interactive Atomicity

**목표**: M1+M2 위에 R2 (Pro 플랜), R5 (모드 전환), 그리고 비대화형 환경 안전성 추가.

### M3 Tasks

1. **REQ-GMC-006 — Pro 플랜 검사 (R2 mitigation)**
   - GWT-14, GWT-15 RED → GREEN.
   - 검출 mechanism: Z.AI API `GET /v1/account/tier` 또는 동등 endpoint 호출 (정확한 endpoint 는 D2 에서 확정). 검출 실패 시 안전하게 안내 메시지로 fallback (REQ-GMC-006 의 (c)).
   - 호출 자체는 mockable 인터페이스로 분리 (`type TierChecker interface`).

2. **REQ-GMC-009 — 모드 전환 무영향 (R5 mitigation)**
   - GWT-19, GWT-20 RED → GREEN.
   - `moai cc`, `moai cg` 의 settings 쓰기 코드를 점검하여 `mcpServers` 키를 절대 삭제하지 않음을 검증. 기존 코드가 이미 그 동작을 보장하면 회귀 테스트만 추가.
   - D3 정제: `moai cc`/`moai cg` 호출 시점에 외부 API 호출이 일어난다면, 단위 테스트는 `httptest.Server` 또는 함수 인터페이스 주입으로 mock.

3. **REQ-GMC-010 — 비대화형 + 원자적 쓰기**
   - GWT-21, GWT-22 RED → GREEN.
   - `isatty.IsTerminal(os.Stdin.Fd())` 로 대화형 여부 감지. 비대화형 + 충돌 발생 시 즉시 비-0 종료.
   - 원자적 쓰기: temp file 작성 → fsync → `os.Rename` (atomic on POSIX). 실패 시 temp file 삭제, 원본 보존.
   - D2 정제: M3 의 GWT-22 결과 명세 (정확한 atomic rename 구현 + 실패 시뮬레이션 mechanism) 를 본 plan.md 에 명시 — 향후 v0.2 업데이트.

### M3 Output

- M1+M2 파일들 + 신규 `internal/glm/mcp/tier_check.go`, `internal/glm/mcp/atomic_write.go`.
- 22/22 GWT 통과.

## Refinements (Plan-audit Minor Refinements)

Plan-audit 가 지적한 비차단 minor refinements 3건. Run phase 에서 또는 v0.2 SPEC 갱신 시 정리.

### D1: REQ-GMC-006 (b) 환경변수 매핑 필드 열거

**현재 상태**: REQ-GMC-003 은 `Z_AI_API_KEY` 와 "패키지가 요구하는 호환 변수" 로 모호하게 표현.

**해결 방안 (Run phase 첫 작업)**:
- `@z_ai/mcp-server` 패키지의 `package.json` 또는 README 확인.
- 실제 요구 환경변수 목록 (예: `Z_AI_API_KEY`, `ZHIPU_API_KEY`, `GLM_API_KEY`) 을 정확히 식별.
- spec.md REQ-GMC-003 의 모호 표현을 정확한 변수명 목록으로 교체 (v0.2 또는 patch).
- 단위 테스트 GWT-07 의 assertion 도 그에 맞게 정확한 키 검사.

**원래 표시**: 비차단 — 현재 SPEC 표현으로도 단위 테스트 작성 가능. Run phase 첫 일에 패키지 docs 확인 후 spec.md patch 권장.

### D2: GWT-22 의 M3 atomic write 결과 명세

**현재 상태**: GWT-22 는 "정확한 mechanism 은 D2 에서 확정" 으로 끝남.

**해결 방안 (M3 시점)**:
- POSIX: `os.WriteFile(tmpPath, data, 0644)` → `os.Rename(tmpPath, finalPath)` (atomic on same filesystem).
- Windows: `os.Rename` 도 atomic 보장됨 (NTFS).
- 실패 시뮬레이션 mechanism: 인터페이스 `type FileWriter interface { WriteFile(path string, data []byte, perm fs.FileMode) error }` 주입 → 테스트에서 `errFakeDiskFull` 반환하는 fake 사용.
- spec.md plan.md 양쪽에 atomic write 의 정확한 단계 명시 (v0.2).

**관련 GWT**: GWT-22 (그리고 GWT-21 의 비대화형 실패 경로).

### D3: GWT-19/모킹 mechanism

**현재 상태**: GWT-19 의 "mock injection mechanism 으로 외부 API 호출은 모킹" 표현이 모호.

**해결 방안 (M3 시점)**:
- `moai cc`/`moai cg` 의 외부 API 호출 코드(현재 어디서 일어나는지 grep 으로 확인) 를 인터페이스로 추상화.
- 인터페이스 예시: `type ModeSwitcher interface { Switch(target Mode) error }`. 기본 구현은 실제 호출, 테스트는 mock 구현 사용.
- GWT-19, GWT-20 의 단위 테스트는 인터페이스 주입으로 외부 API 호출 회피.
- spec.md REQ-GMC-009 의 "외부 API 호출은 모킹" 표현을 정확한 인터페이스 이름으로 교체 (v0.2).

## Risk Mitigations Summary

이슈 본문의 R1~R5 mitigations 요약 (각 항목별로 어떤 마일스톤에서 어떤 REQ 가 처리하는지):

| Risk | 설명 | 처리 마일스톤 | 처리 REQ | mitigation |
|------|------|---------------|----------|-----------|
| R1 | 기존 `mcpServers.zai-mcp-server` 등 사용자 키 덮어쓰기 vs 보존 | M2 | REQ-GMC-005, EX-4 | 정규 키만 충돌 검사. 비표준 키는 건드리지 않음. 대화형 3옵션 (보존/백업/취소). |
| R2 | GLM Coding Plan tier (Lite vs Pro) — Vision 은 Pro 이상 | M3 | REQ-GMC-006 | Pro 플랜 검사 + 검출 실패 시 안내 메시지 fallback. websearch/webreader 는 영향 없음. |
| R3 | Node.js 부재 graceful 감지 | M2 | REQ-GMC-004 | `exec.LookPath` 로 `node`+`npx` 검사. 3요소 메시지 (부재/권장 버전/설치 가이드). settings 변경 없음. |
| R4 | `~/.claude.json` (user scope) vs 프로젝트 `.mcp.json` 결정 | M2 | REQ-GMC-008 | **결정**: 기본값은 user scope (`~/.claude.json`). `--scope project` 명시 시 프로젝트 `.mcp.json` 사용. |
| R5 | settings.json 모드 전환 동기화 | M3 | REQ-GMC-009, GWT-19/20 | 모드 전환은 `mcpServers` 키를 건드리지 않음. 회귀 테스트로 보장. |

## Exclusions Reaffirmed

이슈 본문의 EX-1, EX-6 + 본 SPEC 에서 추가한 EX-2~EX-5, EX-7:

- **EX-1**: Mainland China endpoint variant — v0.2 (별도 SPEC 가능).
- **EX-2**: 자체 Vision MCP 서버 구현 — `@z_ai/mcp-server` thin wrapper 만.
- **EX-3**: GLM 토큰 자동 갱신 — SPEC-GLM-001 책임.
- **EX-4**: 비표준 사용자 키 자동 정리 — 보존 (REQ-GMC-005 후반부).
- **EX-5**: Pro 미가입 시 자동 fallback 모델 선택 — 명확한 에러만 (REQ-GMC-006).
- **EX-6**: 실 API 키 통합 테스트 — v0.2 (Z_AI_TEST_API_KEY 환경변수 보호).
- **EX-7**: `moai glm tools` 외 진입점 (슬래시 커맨드, GUI, 자동 탐지) — 범위 밖.

## Quality Gates (per Milestone)

각 마일스톤 종료 시 다음을 모두 통과해야 다음 마일스톤 또는 PR 진행:

- [ ] `go test -race ./...` (full suite) green
- [ ] `golangci-lint run` green
- [ ] `go vet ./...` green
- [ ] 신규 패키지 line coverage ≥85%
- [ ] CLAUDE.local.md §6 의 test isolation 규칙 준수 (`t.TempDir()` 사용, `t.Setenv("HOME", ...)` 금지)
- [ ] CLAUDE.local.md §13 의 GLM dev project 통합 테스트 금지 규칙 준수
- [ ] CLAUDE.local.md §15 의 16-language neutrality (템플릿 변경 시) 준수

## Branching and PR Strategy

- 본 작업 브랜치: `claude/issue-756-20260501-0519` (이미 존재).
- M1, M2, M3 각 마일스톤은 별도 PR 으로 분할 가능 (선택). 또는 단일 PR (`feat/SPEC-GLM-MCP-001-zai-mcp-integration`) 로 한꺼번에 머지.
- Merge strategy: feature 브랜치이므로 **squash** 머지 (CLAUDE.local.md §18.3).
- Conventional Commit prefix: `feat(glm): add Z.AI MCP server integration`.

## Documentation Sync (Sync phase)

Run phase 완료 후 `/moai sync SPEC-GLM-MCP-001` 호출 시 다음 문서가 동기화되어야 한다 (CLAUDE.local.md §17 4개국어 동기화 규칙):

- `docs-site/content/{ko,en,ja,zh}/reference/cli/glm-tools.md` — `moai glm tools` 명령 reference.
- `.claude/rules/moai/core/settings-management.md` — Z.AI MCP 통합 노트.
- `internal/template/templates/.claude/rules/moai/core/settings-management.md` — 템플릿 미러본.
- `CHANGELOG.md` — 다음 minor release 항목.

## Open Questions (Run phase 시작 전 확인 필요)

본 plan.md 작성 시점에 확정되지 않은 사항. Run phase 첫 단계에서 확인 후 spec.md/plan.md patch:

1. **OQ1 (D1 관련)**: `@z_ai/mcp-server` 패키지의 정확한 환경변수 키 이름 — Run phase 첫 작업.
2. **OQ2**: Z.AI Pro 티어 검출 endpoint 정확한 경로 (`/v1/account/tier` 가설).
3. **OQ3**: `moai cc`/`moai cg` 의 settings 쓰기 코드 위치 (grep 으로 확인 후 인터페이스 주입 적용 — D3).
4. **OQ4**: 백업 키 timestamp 형식 — unix epoch 권장 (간결, 정렬 가능). ISO-8601 도 가능.

## References

- Parent SPEC: `spec.md` (10 REQ, EARS 분포)
- Acceptance: `acceptance.md` (22 GWT 시나리오)
- Plan-audit report: `.moai/reports/plan-audit/SPEC-GLM-MCP-001-review-1.md` (PASS, 0.91)
- Related SPECs: `SPEC-GLM-001` (base GLM compatibility), `SPEC-LSPMCP-001`, `SPEC-CC2122-MCP-001`
- Issue: #756
- CLAUDE.local.md §13 (GLM integration testing rules), §17 (docs-site 4-language sync), §18 (Enhanced GitHub Flow)
