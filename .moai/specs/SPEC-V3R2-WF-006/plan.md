# Plan: SPEC-V3R2-WF-006 — Output Styles Alignment (MoAI, Einstein)

- SPEC: SPEC-V3R2-WF-006
- Wave: 1 (Leaf, no dependencies — 병렬 CON-001 / EXT-001 / WF-001)
- Methodology: TDD (quality.yaml default)
- Created: 2026-04-24
- Updated: 2026-04-25 (v1.1.0 — plan-audit follow-up)
- Author: manager-spec

이 문서는 spec.md의 15개 REQ(Ubiquitous 5 / Event-Driven 3 / State-Driven 2 / Optional 2 / Complex 3)를 구현하기 위한 phase/작업/롤아웃 계획이다. `.moai/reports/plan-audit/V3R2-triage-audit-2026-04-24.md` 에서 WF-006은 PASS 판정이며 "leaf SPEC (no deps)"로 분류되었다.

Scope 요약: (a) Claude Code output-style frontmatter 스키마 검증 테스트 codify, (b) loading precedence(`project > user > "MoAI"` default) 정책 문서화, (c) template/local byte-identity drift guard를 CI로 강제. Go loader 구현은 본 SPEC 범위 외(SPEC-V3R2-EXT-002).

---

## 0. 실행 전 확인 (Precheck)

- [ ] `make build && make install` 최신 바이너리 탑재 (템플릿 변경 시 필수 — CLAUDE.local.md §2 Template-First)
- [ ] `go test ./internal/template/... -count=1` 현재 baseline 녹색 확인 (`commands_audit_test.go` 참조 패턴)
- [ ] `diff -rq .claude/output-styles/moai internal/template/templates/.claude/output-styles/moai` 출력이 비어 있는지 확인 (현재 byte-identical 전제)
- [ ] `grep -rn '"outputStyle"' .claude/settings.json internal/template/templates/.claude/settings.json.tmpl` 로 기본값이 `"MoAI"`로 일치하는지 확인
- [ ] `dependencies: []` 재확인 — Wave 1 leaf, 병렬 착수 가능

---

## 1. Phase Breakdown

| Phase | 목표 | 주요 산출물 | 가드레일 |
|-------|------|-------------|----------|
| P1. Schema test | Output style frontmatter 스키마 검증 테스트 신설 | `internal/template/output_styles_audit_test.go` | `commands_audit_test.go` 패턴 준수, 외부 의존 0 |
| P2. Drift guard | template↔local byte-identity 검증 테스트 | 동일 파일 또는 `output_styles_drift_test.go` | make build 상에서도 green |
| P3. Precedence 문서화 | `outputStyle` 해석 precedence 명문화 | `settings-management.md` 섹션 확장 + template mirror | 공식 Claude Code 동작과 정합 |
| P4. 기본값/Fallback 명시 | unknown 스타일 → "MoAI" fallback + 경고 로그 정책 문서 | `settings-management.md` fallback 섹션 | 본 SPEC은 정책 확립만(구현은 EXT-002) |
| P5. BC guard (3rd style) | 3번째 style 추가 차단 테스트 | `output_styles_audit_test.go` count assertion | CI에서 실패 시 `OUTPUT_STYLE_UNVERIFIED` 메시지 |
| P6. Verification | 전체 테스트 + drift 재확인 + docs lint | 테스트 리포트 | TRUST 5 gate |

각 Phase는 독립 커밋. Phase 단위로 `go test ./internal/template/... -race` 녹색 유지.

---

## 2. Files to Create / Modify

### 2.1 신규 생성

| 경로 | 목적 | 예상 LOC |
|------|------|----------|
| `internal/template/output_styles_audit_test.go` | 프론트매터 스키마 + 스타일 개수 + MoAI/Einstein boolean 계약 + drift 검증 (ALL-IN-ONE) | ~280 |

설계 결정: `commands_audit_test.go` (225 LOC, 2개 Test 함수)와 동일 구조·스타일로 단일 파일 내에 테스트 함수 복수 배치. 별도 `output_styles_drift_test.go` 분리 안 함 (패키지 내 타 감사 테스트 정렬).

### 2.2 수정

| 경로 | 수정 내용 | 근거 REQ |
|------|-----------|----------|
| `.claude/rules/moai/core/settings-management.md` | `outputStyle` precedence 섹션 신설: project > user > "MoAI" default; unknown fallback 정책; frontmatter 스키마 계약 요약 | REQ-WF006-004, 008, 009, 015 |
| `internal/template/templates/.claude/rules/moai/core/settings-management.md` | 위와 동일 내용 (템플릿 동기화 — Template-First 원칙) | 동일 |
| `.moai/specs/SPEC-V3R2-WF-006/acceptance.md` | 본 SPEC의 DoD/Traceability (별도 작업) | N/A |
| `.moai/specs/SPEC-V3R2-WF-006/tasks.md` | 본 SPEC의 Task 분해 (별도 작업) | N/A |

### 2.3 건드리지 않음 (Out of Scope)

- `.claude/output-styles/moai/moai.md` body — spec.md §1.2 Non-Goals (body rewrite 금지)
- `.claude/output-styles/moai/einstein.md` body — 동일
- `.claude/settings.json` `outputStyle` 기본값 "MoAI" — 이미 일치, 변경 불필요
- `internal/template/templates/.claude/settings.json.tmpl` outputStyle 라인 — 동일
- Go loader/파서 구현 — SPEC-V3R2-EXT-002 범위 (REQ-WF006-012는 "Where future" optional로 선언만)
- `/config` UI — Non-Goal
- 3번째 style 추가 — Non-Goal (CI에서 차단하는 감사 테스트만 작성)

---

## 3. Go Package Design

### 3.1 패키지 배치

기존 `package template` (경로: `internal/template/`) 내에 테스트 파일만 추가. 별도 프로덕션 패키지 신설 없음. 이유:

- 본 SPEC은 감사/문서 확립이 목표이며 런타임 loader는 EXT-002 영역
- `commands_audit_test.go` 가 이미 임베디드 템플릿을 `fs.WalkDir`로 순회하는 패턴을 확립하고 있어 동일 패턴 재사용 가능
- `parseFrontmatterAndBody` / `countNonEmptyLines` 헬퍼(현재 `commands_audit_test.go`에 정의)는 export 하지 않되, 본 신규 파일에서는 자체 frontmatter 파서(필요 시 복제 또는 소규모 util)로 한정. 동일 패키지 내부이므로 기존 unexported 함수를 공유 호출 가능.

### 3.2 테스트 파일 구조 (초안)

```go
// internal/template/output_styles_audit_test.go
package template

import (
    "bufio"
    "bytes"
    "io/fs"
    "strings"
    "testing"
)

// TestOutputStylesFrontmatterSchema verifies REQ-WF006-001, REQ-WF006-005.
func TestOutputStylesFrontmatterSchema(t *testing.T) { ... }

// TestOutputStylesExactlyTwo enforces REQ-WF006-002 and REQ-WF006-014
// (blocks 3rd style without schema pass).
func TestOutputStylesExactlyTwo(t *testing.T) { ... }

// TestOutputStylesTemplateLiveParity enforces REQ-WF006-003, REQ-WF006-010.
// Fails with OUTPUT_STYLE_DRIFT error message when the two trees diverge.
func TestOutputStylesTemplateLiveParity(t *testing.T) { ... }

// TestOutputStylesEncoding verifies UTF-8 + LF-only (REQ assumption §4).
func TestOutputStylesEncoding(t *testing.T) { ... }
```

검증 항목 (함수별):

1. `TestOutputStylesFrontmatterSchema`
   - 각 `.md` 파일에 `name` (string, non-empty), `description` (string, non-empty), `keep-coding-instructions` (boolean literal `true`/`false`) 존재
   - `name == "MoAI"` → `keep-coding-instructions == true`
   - `name == "Einstein"` → `keep-coding-instructions == false`
   - 위반 시 에러 메시지에 `OUTPUT_STYLE_SCHEMA_ERROR` 접두사 포함 (REQ-WF006-007, 013)

2. `TestOutputStylesExactlyTwo`
   - embedded templates 의 `.claude/output-styles/moai/` 하위 파일 수 == 2
   - 파일 이름 집합 == `{moai.md, einstein.md}`
   - 초과 시 `OUTPUT_STYLE_UNVERIFIED` 에러 메시지 (REQ-WF006-014)

3. `TestOutputStylesTemplateLiveParity`
   - live 경로(`.claude/output-styles/moai/*.md`)와 template(embedded) byte-by-byte 동일
   - **주의**: live 파일은 프로젝트 루트 기준 읽기. OPEN-A RESOLVED (v1.1.0) — `runtime.Caller(0)` ascent 로 프로젝트 루트(`.moai/` 마커) 탐지 후 `filepath.Join(root, ".claude/output-styles/moai")` 로 접근. 자세한 알고리즘은 §8 OPEN-A 참조.
   - 불일치 시 `OUTPUT_STYLE_DRIFT` 에러 메시지 (REQ-WF006-010)
   - sandbox / non-standard checkout 환경에서 `.moai/` 마커 부재 시 `t.Skip("live tree not available")` graceful fallback.

4. `TestOutputStylesEncoding`
   - UTF-8 BOM 없음, CR 없음(LF only) — spec.md §4 Assumption 5 검증

### 3.3 설계 원칙

1. **Pure test-only**: 프로덕션 패키지 영향 0. `make build` 산출물 크기 변화 0.
2. **하드코딩 금지**: `"MoAI"`, `"Einstein"`, `"moai.md"`, `"einstein.md"` 같은 리터럴은 테스트 내부 const 블록에 단일 원천으로 배치 (CLAUDE.local.md §14).
3. **테스트 격리**: `t.Parallel()` 적용, 파일 시스템 수정 0, `t.TempDir()` 불필요(read-only 검증).
4. **16개 언어 중립**: 본 SPEC은 MoAI-ADK 내부 플랫폼 감사 테스트로 사용자 프로젝트 언어 선택과 무관 (audit 보고서 MP-4 "N/A" 판정).

---

## 4. Test Strategy (TDD)

### 4.1 RED-GREEN-REFACTOR 사이클

이 SPEC은 "검증 테스트 자체를 작성"하는 메타 성격이라 TDD는 다음과 같이 적용한다:

- **RED (가설 검증 단계)**: 의도적으로 규칙을 어긴 임시 fixture(`output-styles/tmp-invalid.md` in-memory via fs.Sub)로 실패 경로 확인 → production fixture 수정 없이 subtest 내에서 synthetic 입력으로 테스트. 단, 임베디드 템플릿 fs는 mutable 불가이므로 subtest에서 `fs.WalkDir` 대신 in-memory string 기반 reference parsing 경로를 별도 호출하여 실패 조건을 트리거.
- **GREEN**: 실제 live/template 파일을 대상으로 감사 테스트 통과.
- **REFACTOR**: 리터럴 상수 추출, 에러 메시지 포맷 통일(`OUTPUT_STYLE_SCHEMA_ERROR | OUTPUT_STYLE_DRIFT | OUTPUT_STYLE_UNVERIFIED`).

### 4.2 단위 테스트 매트릭스

| Test 함수 | 케이스 | 근거 REQ |
|-----------|--------|----------|
| `TestOutputStylesFrontmatterSchema` (table-driven) | moai / einstein 각각 name / description / keep-coding-instructions 존재·타입·값 | REQ-001, 005, 007, 013 |
| `TestOutputStylesFrontmatterSchema_Synthetic` | missing keep-coding-instructions / non-bool / missing name → 내부 파서가 각각 오류 반환 | REQ-007, 013 |
| `TestOutputStylesExactlyTwo` | 파일 수·이름 집합 | REQ-002, 014 |
| `TestOutputStylesTemplateLiveParity` | live vs template byte diff | REQ-003, 010 |
| `TestOutputStylesEncoding` | UTF-8 valid + LF only | §4 Assumption |

커버리지 목표: 신규 테스트 파일 자체의 line 커버리지는 측정 대상 아님(테스트 코드). 대신 본 SPEC 성공 지표는 "spec.md AC 12개 중 12개 대응 테스트 존재".

### 4.3 통합/수동 검증 (수동)

Precedence (REQ-004, 006, 015) 및 unknown fallback(REQ-008) 은 Claude Code 세션 행위에 의존 — Go 테스트로 재현 불가. 따라서:

- **문서화로 고정** (P3, P4): `settings-management.md` 에 precedence 표 + fallback 규칙 기술.
- **수동 수용 시나리오** (acceptance.md AC-WF006-05/06/07/08) 는 "문서 존재 + 후속 EXT-002 loader 테스트에서 자동화" 로 분할.
- EXT-002가 구현되는 시점에 REQ-012 가 활성화되며, 그때 Go-level precedence 테스트가 추가된다. 본 SPEC은 "정책 확립 + 감사 테스트" 에 한정.

### 4.4 테스트 격리 규칙 (CLAUDE.local.md §6)

- [HARD] 파일 시스템 수정 없음 → `t.TempDir()` 불필요
- [HARD] `t.Setenv("HOME", ...)` 금지
- [HARD] OTEL env 조작 금지
- 테이블 드리븐 + `t.Run(tt.name, ...)` 일관 적용
- `t.Parallel()` 기본 활성화 (`commands_audit_test.go` 와 동일)

---

## 5. Rollout Plan

### 5.1 단계

1. **Phase 1 커밋 (Schema test)**: `output_styles_audit_test.go` 신규 작성, 3개 Test 함수 (Schema / ExactlyTwo / Encoding) 포함. 프로덕션 영향 0.
2. **Phase 2 커밋 (Drift guard)**: `TestOutputStylesTemplateLiveParity` 추가. 동일 파일 내. `make build && go test` 녹색 확인.
3. **Phase 3 커밋 (Precedence docs)**: `settings-management.md` live + template 동시 수정. `make build` 필수.
4. **Phase 4 커밋 (Fallback docs)**: unknown/default fallback 섹션 추가.
5. **Phase 5 커밋 (BC guard)**: `TestOutputStylesExactlyTwo` 가 이미 P1에 포함되었으면 skip. 누락 시 이 단계에서 보강.
6. **Phase 6 커밋 (Verification)**: `go test ./... -race -count=1`, `go vet ./...`, `golangci-lint run ./internal/template/...`, `make build && make install`.

각 Phase는 Conventional Commits 형식 (`test(template): ...`, `docs(settings): ...`, `chore: ...`). 커밋 본문 한국어 사용 근거: `.moai/config/sections/language.yaml` 의 `git_commit_messages: ko` 설정. 이 설정은 로컬 language.yaml 파일에 코드로 선언되어 있으며, 본 SPEC 은 프로젝트 전역 규칙을 따를 뿐 별도 정책을 신설하지 않는다.

### 5.2 v3.0 하위 Wave 위치

- Wave 1 leaf — CON-001 / EXT-001 / WF-001과 병렬 가능 (파일 겹침 없음).
- SPEC-V3R2-EXT-002(Go output-style loader)는 본 SPEC 머지 후 시작 (REQ-WF006-012 prerequisite — frontmatter 스키마 consume).
- **Watch-item (downgrade from hard guard, v1.1.0)**: CON-001 의 `settings-management.md` 편집 여부 — 2026-04-25 boundary audit 결과 CON-001 은 `settings-management|output-styles|outputStyle` 를 zero-reference 함이 독립 grep 으로 확인됨 (audit 보고서 §7). 따라서 병렬 실행은 안전하며, T4 머지 시점에만 재확인하여 충돌 발생 시에만 rebase 한다. "sequential merge 강제" 하지 않는다.

### 5.3 모니터링 포인트

- 첫 1주: `OUTPUT_STYLE_DRIFT` 실패 발생 빈도 (template mirror 갱신 누락 탐지).
- 3번째 style 추가 PR 시도 발생 시 `OUTPUT_STYLE_UNVERIFIED` 가 잘 차단되는지.
- 향후 Claude Code 공식 스키마 변경 공지 모니터링 (Risk #1).

---

## 6. Rollback Plan

### 6.1 기준

- 회귀 징후: 감사 테스트가 false-positive로 실패하여 정상 PR을 차단하는 경우.

### 6.2 롤백 절차

1. **Phase 6 롤백 불요**: verification은 read-only.
2. **Phase 1-2 (테스트 파일) 롤백**: `git revert <commit>` 로 `output_styles_audit_test.go` 제거. 프로덕션 영향 0.
3. **Phase 3-4 (문서) 롤백**: `settings-management.md` 이전 버전 복원. template mirror 동기화 (`make build`).
4. **완전 철회**: `go.mod` 변경 없음, 임베디드 바이너리 size 변화 없음 → 흔적 없이 철회 가능.
5. [HARD] CLAUDE.local.md `feedback_revert_runtime_config.md` 참조: 런타임 config(`llm.yaml` 등)는 revert 대상 제외. 본 SPEC은 런타임 config 수정 없음.

### 6.3 데이터 보존

- Output style body 파일(moai.md / einstein.md)은 본 SPEC에서 수정하지 않음 → 사용자 혹은 upstream 수정 내역 보존.
- 템플릿/live 모두 읽기만 수행 → 데이터 손실 위험 0.

---

## 7. Risks & Mitigations (plan-side)

| 리스크 | 영향 | 완화 |
|--------|------|------|
| 임베디드 fs와 live 파일 경로 읽기 비대칭 | drift 테스트가 CI 환경에서만 green | 테스트는 `os.Getwd` → `filepath.Abs` 로 live 경로 확정; embedded는 `EmbeddedTemplates()` 사용 (상대경로 의존 0) |
| 기존 YAML frontmatter 파서가 boolean 타입을 문자열로 저장 | `keep-coding-instructions: true` 검증 false positive | 별도 파서 헬퍼로 raw 라인을 `"true"`/`"false"` 리터럴 체크 (`parseFrontmatterAndBody` 는 문자열 저장 후 `val == "true" or "false"` 로 명시 비교) |
| Claude Code 공식 스키마 변경 (keys 추가/삭제) | 테스트가 obsolete | 본 SPEC의 스키마를 "최소 집합" (name/description/keep-coding-instructions) 으로 고정하고 기타 keys 는 tolerate — 엄격하지 않은 superset 허용 명시 |
| Einstein body의 Context7 MCP 의존 | 런타임 fallback 실패 | body 수정 범위 외(Non-Goal); Einstein body 내 soft-fallback 문구는 현재 그대로 보존 |
| live 파일 권한/심볼릭 링크 | 테스트 파일 읽기 실패 | `os.ReadFile` 오류를 `t.Fatalf` 로 명시 — 환경 문제와 drift 구분 |
| user-level override 기대하는 사용자 | UX 혼란 | REQ-004 precedence를 `settings-management.md` 에 명문화 + 예시 2개 포함 |

---

## 8. Open Questions

**OPEN-A — RESOLVED (v1.1.0, 2026-04-25)**: `TestOutputStylesTemplateLiveParity` 의 live-tree 경로 탐색 전략은 `runtime.Caller(0)` ascent 로 확정.

구현 알고리즘:

1. `_, thisFile, _, _ := runtime.Caller(0)` 로 테스트 소스 파일 절대 경로 획득 (`.../internal/template/output_styles_audit_test.go`).
2. `dir := filepath.Dir(thisFile)` 에서 시작해 `.moai/` 디렉터리가 존재하는 상위 디렉터리까지 `filepath.Dir(dir)` 반복 (최대 8 levels 상향 — 극단적인 non-standard layout 방어).
3. `.moai/` 를 찾지 못하면 `t.Skip("live tree not available: .moai marker not found up to 8 levels")` — CI sandboxed checkout 등의 예외 환경에서 graceful degradation.
4. 찾은 프로젝트 루트에서 `filepath.Join(root, ".claude/output-styles/moai")` 로 live dir 접근.

**Rationale**:
- runtime.Caller 는 빌드 타임 `-trimpath` 적용 시에도 테스트 컨텍스트에서 절대 경로를 반환 (Go test binary 는 `-trimpath` 미적용이 기본).
- `.moai/` 마커는 main repo 와 Claude Code native worktree (`.claude/worktrees/abc123/.moai/` 존재) 양쪽에서 동작 — worktree 는 프로젝트 루트에 `.moai/` 를 동기화한다는 MoAI Worktree 전제(worktree-integration.md)에 의존.
- `os.Getwd()` 와 달리 test CWD 가 변형되어도 동작 (`go test ./internal/template/...` vs `go test` from repo root 양쪽 안전).
- env 의존 없음 — CI env 설정 불요.

**Task T0 smoke 검증**: 아래 프로토타입을 `go run` 또는 `go test -run TestRuntimeCallerAscent` 로 실행해 프로젝트 루트가 `.moai/` 포함 디렉터리로 올바르게 잡히는지 확인 (T0 DoD 참조).

```go
// OPEN-A 검증 스케치 (T0 smoke only, production 코드에는 포함 금지)
_, this, _, _ := runtime.Caller(0)
dir := filepath.Dir(this)
for i := 0; i < 8; i++ {
    if _, err := os.Stat(filepath.Join(dir, ".moai")); err == nil {
        fmt.Println("resolved root:", dir)
        break
    }
    parent := filepath.Dir(dir)
    if parent == dir { break }
    dir = parent
}
```

**OPEN-B**: 본 SPEC의 스키마 테스트는 "frontmatter에 명시된 3개 key만 엄격 검증" 이며 **추가 key(예: `model`, `temperature`) 가 있을 때 허용/거부** 정책이 spec.md에 명시되지 않았다.

**Plan 제안**: spec.md §7 Constraints "Frontmatter 스키마 변경은 breaking change" 라는 문장에서 추정하면 "엄격 superset" (추가 key 허용, 필수 key 결락 금지) 가 타당. 테스트는 **필수 key 존재 + 타입** 만 검증하고 **추가 key는 tolerate** 하는 방침으로 1차 구현. PR 리뷰 시 확정.

**OPEN-C**: `OUTPUT_STYLE_SCHEMA_ERROR` 같은 오류 코드가 spec.md §5.2 REQ-007/013 에 정의되어 있으나, 실제로 이 문자열이 어디서 emit되어야 하는가 (Go test message만? `make build` 출력? 둘 다?).

**Plan 제안**: Wave 1 범위에서는 **Go test 에러 메시지에 문자열 그대로 포함** (테스트 실패 시 CI 로그에서 검색 가능). `make build` 레벨 통합은 `EXT-002` 또는 별도 Hook에서 수행. acceptance.md에 "CI rejection via go test failure message" 로 명시.

**OPEN-D**: REQ-WF006-006/015 (project > user precedence) 는 Claude Code 세션의 실제 동작이라 Go 테스트로 검증 불가. "문서화 + EXT-002에서 loader 구현 시 검증" 으로 분할한 결정이 타당한가.

**Plan 제안**: 타당. acceptance.md AC-05/06 은 "문서 수용 + EXT-002 deferral" 로 표시. `settings-management.md` 예시에 precedence 테이블 포함 (docs-level acceptance).

**OPEN-E**: REQ-WF006-009 default "MoAI" 이 현재 `.claude/settings.json` 및 `settings.json.tmpl` 에 이미 명시적으로 `"outputStyle": "MoAI"` 로 설정되어 있다. "아무 설정도 없을 때" (REQ-009 "no explicit outputStyle")와 현재 상태(명시적 "MoAI") 는 구별 가능한가.

**Plan 제안**: 스펙 문구("no explicit outputStyle is set in any settings file")와 현재 상태("명시적 MoAI")는 외부 관찰상 동일한 효과를 낸다. 문서에 "명시 생략 시 fallback = MoAI, 템플릿 기본값은 명시적 MoAI"로 병기. 감사 테스트는 `settings.json.tmpl` 의 `"outputStyle": "MoAI"` 라인 존재를 보조 검증 (선택).

---

End of Plan.
