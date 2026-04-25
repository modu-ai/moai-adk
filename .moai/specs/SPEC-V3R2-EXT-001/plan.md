# Plan: SPEC-V3R2-EXT-001 — Typed Memory Taxonomy

- SPEC: SPEC-V3R2-EXT-001
- Wave: 1 (Leaf, no dependencies)
- Methodology: TDD (quality.yaml default)
- Created: 2026-04-24
- Author: manager-spec

이 문서는 spec.md의 17개 REQ를 구현하기 위한 phase/작업/롤아웃 계획이다. `.moai/reports/plan-audit/V3R2-triage-audit-2026-04-24.md` PASS 판정을 기준으로 작성한다.

---

## 0. 실행 전 확인 (Precheck)

- [ ] `make build && make install` 최신 바이너리 탑재 (훅 변경 시 재시작 필수 — MEMORY.md Hard Constraint)
- [ ] `go test ./internal/hook/...` 현재 baseline 녹색 확인
- [ ] `.claude/agent-memory/` 디렉터리 18개 에이전트 스캔 (본 SPEC은 read + warn 경로만; 초기 wave는 파일 수정 없음)
- [ ] `dependencies: []` 재확인 — Wave 1 leaf이므로 병렬 시작 가능

---

## 1. Phase Breakdown

| Phase | 목표 | 주요 산출물 | 가드레일 |
|-------|------|-------------|----------|
| P1. Rule 확장 | `moai-memory.md` 4-type 가이드 완성 | 룰 문서 2곳 (template + live) | template-first 원칙 |
| P2. Go 공용 유틸 | memory 파일 파서 + type enum | `internal/hook/memo/taxonomy.go` | Pure + 단위 테스트 |
| P3. SessionStart staleness | 24h mtime wrap + aggregated warn | `session_start.go` 경로 추가 | 기존 흐름 비파괴 |
| P4. PostToolUse audit | MISSING_TYPE / INDEX_OVERFLOW / EXCLUDED_CATEGORY / DUPLICATE | `post_tool.go` memo audit 경로 | non-blocking warning |
| P5. Config + Docs | `workflow.yaml` staleness_hours 키, 예시 문서 | `workflow.yaml`, `moai-memory.md` 예시 | 기본값만 정의 |
| P6. Verification | 전체 테스트 + audit sweep + 회귀 리뷰 | 테스트 리포트 | TRUST 5 gate |

각 Phase는 독립 커밋. Phase 단위로 `go test ./... -race` 녹색 유지.

---

## 2. Files to Create / Modify

### 2.1 신규 생성

| 경로 | 목적 | 예상 LOC |
|------|------|----------|
| `internal/hook/memo/taxonomy.go` | 4-type enum + frontmatter 파서 + staleness/cap 검사 | ~220 |
| `internal/hook/memo/taxonomy_test.go` | 테이블 드리븐 단위 테스트 (enum, parser, audit rules) | ~350 |
| `internal/hook/memo/fixtures/` | 테스트용 샘플 memory 파일 (user/feedback/project/reference + 오류 케이스) | 8 files |
| `internal/hook/memo/staleness.go` | mtime → `<system-reminder>` wrap + aggregation 로직 | ~140 |
| `internal/hook/memo/staleness_test.go` | 24h threshold, aggregation, wrap 포맷 테스트 | ~220 |
| `internal/hook/memo/audit.go` | MISSING_TYPE / INDEX_OVERFLOW / EXCLUDED_CATEGORY / DUPLICATE 검출 | ~180 |
| `internal/hook/memo/audit_test.go` | 각 warning 규칙 테스트 | ~260 |

### 2.2 수정

| 경로 | 수정 내용 | 근거 REQ |
|------|-----------|----------|
| `.claude/rules/moai/workflow/moai-memory.md` | 4-type 가이드, MEMORY.md 200줄 cap, staleness caveat, writing guideline, 예시 추가 | REQ-EXT001-005, 009-012 |
| `internal/template/templates/.claude/rules/moai/workflow/moai-memory.md` | 위와 동일 내용 (템플릿 동기화) | Template-First 원칙 |
| `internal/hook/session_start.go` | memory 로드 시 taxonomy staleness 호출 + aggregated warning 출력 | REQ-EXT001-006, 017 |
| `internal/hook/session_start_test.go` | stale wrap, aggregation 경로 테스트 추가 | REQ-EXT001-006, 017 |
| `internal/hook/post_tool.go` | memory 파일 Write/Edit 시 audit 경로 호출 | REQ-EXT001-007, 008, 015, 016 |
| `internal/hook/post_tool_test.go` | 각 warning 코드 발생 경로 테스트 | 동일 |
| `.moai/config/sections/workflow.yaml` | `memory.staleness_hours: 24`, `memory.index_line_cap: 200` 기본 키 등록 | REQ-EXT001-003, 006 |
| `internal/template/templates/.moai/config/sections/workflow.yaml` | 동일 | Template-First |
| `internal/config/loader.go` 또는 해당 섹션 구조체 | 신규 키 스키마 반영 (optional, 기본값 지정) | 하드코딩 방지 |
| `Makefile` | (필요 시) `make build` 이후 `moai version` echo 가이드 — 변경 없음 | N/A |

### 2.3 건드리지 않음

- `.claude/agent-memory/<agent>/**` 실제 파일: 본 SPEC은 규약/경고만 추가, 기존 메모리 수정/삭제 없음.
- SPEC-V3R2-EXT-002 범위 (Go loader 파싱/거부): 본 SPEC은 enum/스펙만 확정하고 loader는 후속 SPEC.

---

## 3. Go Package Design

### 3.1 패키지 배치

`internal/hook/memo/` 패키지 신설 (기존 `internal/hook/memo/` 디렉터리가 비어 있는지 확인 필요 — 이미 존재 시 서브패키지 `taxonomy` 또는 파일만 추가). 단일 패키지로 staleness/audit/taxonomy를 묶는 이유는 모두 memory 파일 대상 pure function이기 때문이다.

### 3.2 공개 API (안정화 포인트)

```go
// internal/hook/memo/taxonomy.go
package memo

type MemoryType string

const (
    TypeUser      MemoryType = "user"
    TypeFeedback  MemoryType = "feedback"
    TypeProject   MemoryType = "project"
    TypeReference MemoryType = "reference"
)

var ValidTypes = []MemoryType{TypeUser, TypeFeedback, TypeProject, TypeReference}

type Frontmatter struct {
    Name        string
    Description string
    Type        MemoryType
    Raw         map[string]string // parser fallback
}

// ParseFile reads a markdown file and returns its frontmatter
// plus body. Returns ErrNoFrontmatter when absent.
func ParseFile(path string) (Frontmatter, string, error)

// ValidateType returns nil iff t is one of the 4 enum values.
func ValidateType(t MemoryType) error
```

```go
// internal/hook/memo/staleness.go
type StaleReport struct {
    Path     string
    Age      time.Duration
    Wrapped  string // original content wrapped in <system-reminder>
}

// DetectStale returns files whose mtime exceeds thresholdHours.
func DetectStale(dir string, thresholdHours int, now time.Time) ([]StaleReport, error)

// AggregateWarning returns a single warning line for 10+ stale files
// or per-file warnings for < 10. Threshold from REQ-EXT001-017.
func AggregateWarning(reports []StaleReport) string
```

```go
// internal/hook/memo/audit.go
type AuditCode string

const (
    WarnMissingType        AuditCode = "MEMORY_MISSING_TYPE"
    WarnIndexOverflow      AuditCode = "MEMORY_INDEX_OVERFLOW"
    WarnExcludedCategory   AuditCode = "MEMORY_EXCLUDED_CATEGORY"
    WarnDuplicate          AuditCode = "MEMORY_DUPLICATE"
)

type AuditFinding struct {
    Code    AuditCode
    Path    string
    Detail  string
}

func AuditFile(path string) ([]AuditFinding, error)
func AuditIndex(indexPath string, lineCap int) ([]AuditFinding, error)
func AuditDuplicates(dir string) ([]AuditFinding, error)
```

### 3.3 설계 원칙

1. **Pure functions**: 시간/파일 I/O는 명시 인자로 주입 (`now time.Time`, `threshold int`) → 테스트 용이.
2. **non-blocking**: 모든 warning은 exit code 0, stderr로만 출력. `pkg/version` 충돌 방지용 `internal/hook/memo` 네임스페이스 유지.
3. **하드코딩 금지**: 24/200/10 등은 `workflow.yaml` 기본값 + `internal/config/defaults.go` 단일 원천. 코드 literal 금지.
4. **16개 언어 중립**: 구현은 Go지만 룰 문서(`moai-memory.md`)는 특정 언어 편향 금지.

---

## 4. Test Strategy (TDD)

### 4.1 단위 테스트 (packages)

| 대상 | 테스트 파일 | 핵심 케이스 |
|------|-------------|-------------|
| `memo.ParseFile` | `taxonomy_test.go` | 정상 FM / FM 없음 / 잘못된 type / 빈 name·description |
| `memo.ValidateType` | `taxonomy_test.go` | 4 valid + unknown + empty |
| `memo.DetectStale` | `staleness_test.go` | <24h not stale / =24h boundary / >24h stale / dir empty |
| `memo.AggregateWarning` | `staleness_test.go` | 0,1,9,10,11 파일 각각의 출력 포맷 |
| `memo.AuditFile` | `audit_test.go` | MISSING_TYPE / EXCLUDED_CATEGORY (CLAUDE.md mirror) |
| `memo.AuditIndex` | `audit_test.go` | 199/200/201/250 라인 경계 |
| `memo.AuditDuplicates` | `audit_test.go` | 동일 description 2개 / 3개 / 대소문자 차이 |

커버리지 목표: `internal/hook/memo/` 패키지 90% (CLAUDE.local.md §6 Critical packages).

### 4.2 훅 통합 테스트

| 대상 | 테스트 파일 | 시나리오 |
|------|-------------|----------|
| `session_start.go` memory wrap | `session_start_test.go` | fake agent-memory 디렉터리 생성 → 일부만 오래됨 → 최종 payload에 `<system-reminder>` 포함 |
| `session_start.go` aggregation | `session_start_extra_coverage_test.go` | 10+ 스테일 파일 → 단일 aggregated warning 확인 |
| `post_tool.go` audit trigger | `post_tool_test.go` | Write tool이 memory 파일을 수정할 때만 audit 실행, 다른 파일은 skip |

### 4.3 룰 정합성 테스트

- `.claude/rules/moai/workflow/moai-memory.md` 와 `internal/template/templates/...` 동일 내용 확인 테스트: `internal/template/commands_audit_test.go` 패턴을 참고하여 `rules_audit_test.go`(가칭)에 drift 가드 추가 검토(선택). 필수는 아니며 Wave 1 범위 외일 수 있음 — Task 6에서 결정.

### 4.4 테스트 격리 규칙 (CLAUDE.local.md §6)

- [HARD] 임시 memory 파일은 `t.TempDir()` 사용
- [HARD] `t.Setenv("HOME", ...)` 금지
- OTEL env 조작 금지
- 테이블 드리븐 + `t.Run(tt.name, ...)` 일관 적용

---

## 5. Rollout Plan

### 5.1 단계

1. **Phase 1 커밋 (Rule docs)**: 룰 문서 + template 동기화만 (`make build` 포함). 런타임 영향 없음.
2. **Phase 2 커밋 (memo 패키지)**: 공개 API + 단위 테스트. 훅에는 아직 연결 안 함.
3. **Phase 3 커밋 (SessionStart 통합)**: staleness wrap 경로 활성화. `workflow.yaml` 기본값 24h로 출력 최소화.
4. **Phase 4 커밋 (PostToolUse audit)**: non-blocking warning. 기존 훅 흐름 변경 금지.
5. **Phase 5 커밋 (Config + 문서 예시)**: `workflow.yaml` 키 등록 + `moai-memory.md` 예시 삽입.
6. **Phase 6 커밋 (Verification)**: `go test ./... -race`, `go vet`, `golangci-lint run`, `make build`.

각 Phase는 Conventional Commits 형식(`feat(memo): ...`, `docs(memory): ...`).

### 5.2 v3.0 하위 Wave 위치

- Wave 1 leaf — CON-001 / WF-001 / WF-006과 병렬 가능.
- SPEC-V3R2-EXT-002(Go loader)는 본 SPEC 머지 후 시작(REQ-EXT001-013 prerequisite).

### 5.3 모니터링 포인트

- 신규 warning 로그가 실제 세션에 과다 주입되는지: SessionStart 훅 로그에서 `MEMORY_STALE_AGGREGATED` 빈도 확인.
- `MEMORY_MISSING_TYPE` 최초 1주 기간 count 수집 (non-blocking이지만 현황 파악용).

---

## 6. Rollback Plan

### 6.1 기준

- 회귀 징후: `session_start_test.go`/`post_tool_test.go` 녹색 유지 실패, 또는 실제 사용자 세션에서 warning이 실제 동작을 차단하는 경우(본 SPEC은 non-blocking이지만 오탐으로 인한 혼란 포함).

### 6.2 롤백 절차

1. **Phase 4 롤백**: `post_tool.go`에서 memo audit 호출 분기를 env flag `MOAI_MEMORY_AUDIT=0`로 즉시 비활성화 가능 (구현 시 flag 포함).
2. **Phase 3 롤백**: SessionStart의 staleness 호출도 동일한 flag로 비활성화.
3. **Phase 1-2 롤백**: 룰 문서/패키지는 비활성 flag 상태에서도 컴파일/빌드 영향 없음. 필요 시 `git revert <commit>` 단순 되돌림 (CLAUDE.local.md `feedback_revert_runtime_config.md`: `llm.yaml` 등 런타임 config는 보호, `workflow.yaml` 신규 키는 revert 대상에서 제외하고 기본값 유지).
4. **완전 철회**: `go.mod` 변경 없음 → 패키지만 삭제. 테스트/CI green 재확인.

### 6.3 데이터 보존

- 기존 `.claude/agent-memory/**` 파일은 본 SPEC에서 수정/삭제하지 않으므로 롤백 시 사용자 데이터 손실 위험 없음.

---

## 7. Risks & Mitigations (plan-side)

| 리스크 | 완화 |
|--------|------|
| 기존 18개 에이전트 메모리 대부분 type 미선언 → warning 폭주 | SessionStart aggregation(REQ-017) + 초기 wave는 warning-only |
| SessionStart 훅 지연 증가 (파일 많을수록 mtime stat) | `DetectStale`에서 디렉터리 lazy walk + 상한(예: 500 files) |
| `system-reminder` wrap 포맷이 에이전트 프롬프트 해석에 간섭 | 고정 포맷 + `moai-memory.md`에 해석 가이드 삽입 |
| Template drift (live ↔ template) | Phase 1에서 두 경로 동시 수정 + `make build` 필수 |
| 24h boundary가 CI/로컬 시간대 차로 flake | `DetectStale(now time.Time)` 주입형 설계로 테스트 안정 |

---

## 8. Open Questions

**OPEN-1 (non-blocking)**: SessionStart staleness 임계값이 `workflow.yaml` key로 노출되므로 key 이름 확정 필요. 후보: `memory.staleness_hours` (제안) vs `agent_memory.staleness_hours`. 본 Plan에서는 전자로 가정하되 PR 리뷰에서 확정.

**OPEN-2**: `MEMORY_EXCLUDED_CATEGORY` 판정을 위한 "excluded category" 목록(REQ-015)은 spec.md §5.5에 `code patterns, git history, debugging recipes, CLAUDE.md content, ephemeral task state`로 열거되어 있음. 구현 시 이들을 정적 키워드 매칭으로 판정할지, LLM 기반으로 판정할지 결정 필요. Plan은 정적 키워드(예: "```go ", "commit hash", "fix recipe" 등)로 1차 구현 후 false-positive 수집 후 개선 방향. PR 리뷰 시 확정.

**OPEN-3**: `internal/hook/memo/` 기존 디렉터리 구조와 본 SPEC의 신규 파일 충돌 여부. 기존 디렉터리 내용은 별도 Read로 확인 필요. 충돌 시 `internal/hook/memo/taxonomy/` 서브패키지로 분리. Run phase 시작 시 첫 Task에서 확인.

**OPEN-4**: `rules_audit_test.go`(룰 문서 live/template drift 가드) 추가 여부 — Wave 1 범위 외일 수 있음. Run phase Task 6에서 결정.

---

End of Plan.
