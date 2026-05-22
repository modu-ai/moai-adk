---
id: SPEC-V3R6-CODE-COMMENTS-EN-001
title: "Research — Mass migration of Korean comments to English"
version: "0.2.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: Medium
phase: "v3.0.0"
module: "internal/"
lifecycle: spec-anchored
tags: "code-quality, comments, internationalization, en-migration, mass-migration"
tier: L
type: research
---

# Research — SPEC-V3R6-CODE-COMMENTS-EN-001

## 1. Codebase Inventory (2026-05-23 snapshot)

### 1.1 File counts

```bash
# Non-test files with Korean
find internal cmd pkg -name "*.go" ! -name "*_test.go" | xargs grep -l "[가-힣]" | wc -l
# = 96 files
```

```bash
# Test files with Korean
find internal cmd pkg -name "*_test.go" | xargs grep -l "[가-힣]" | wc -l
# = 171 files
```

**Total in scope**: **96 + 171 = 267 Go files**

### 1.2 Per-package distribution

**Non-test files (96)**:

| Package | Files with Korean |
|---------|-------------------|
| internal/cli | 25 |
| internal/harness | 16 |
| internal/migration | 7 |
| internal/config | 4 |
| internal/spec | 3 |
| internal/core | 1 |
| internal/hook | 1 |
| pkg/ | 2 |
| (other internal/*) | ~37 |

**Test files (171)**:

| Package | Test files with Korean |
|---------|------------------------|
| internal/cli | 40 |
| internal/harness | 25 |
| internal/template | 10 |
| internal/hook | 9 |
| internal/spec | 7 |
| internal/migration | 5 |
| internal/lsp | 4 |
| internal/config | 2 |
| internal/core | 1 |
| (other internal/*) | ~68 |

### 1.3 Wave alignment

본 분포가 plan.md §3 7-wave 분할에 직접 반영됨. 모든 96 non-test + 171 test = 267 files 100% 할당:

**Non-test waves (Waves 1-4 = 96 files):**

- Wave 1 (Foundation, 9 files): config=4 / spec=3 / core=1 / hook=1
- Wave 2 (CLI surface, 25 files): cli=25 (non-test)
- Wave 3 (Harness+Migration, 23 files): harness=16 / migration=7
- Wave 4 (pkg + ALL remaining internal/*, **39 files**): pkg=2 + "(other internal/*)" non-test=~37 (worktree, bodp, mx, statusline, doctor, runtime, project, audit, ciwatch, lsp, etc.)
  - **Reconciliation**: 9 + 25 + 23 + 39 = **96 non-test** ✓

**Test waves (Waves 5-7 = 171 files):**

- Wave 5 (Test A, 50 files): cli test=40 + template test=10
- Wave 6 (Test B, 38 files): harness test=25 + lsp test=4 + hook test=9
- Wave 7 (Test C, **~83 files**): migration test=5 + config test=2 + spec test=7 + core test=1 + "(other internal/*) test"=~68 (covers all remaining test files in packages not enumerated in Waves 5-6)
  - **Reconciliation**: 50 + 38 + 83 = **171 test** ✓

**Total**: 96 (non-test) + 171 (test) = **267 Go files** (matches §1.1 inventory)

**Sub-split protocol**: Waves 4 and 7 may sub-split into 4a/4b or 7a/7b if PR size exceeds ~30 files at runtime inventory check (plan.md §3.4 / §3.7).

---

## 2. Korean Span Inventory

### 2.1 Comment lines

```bash
find internal cmd pkg -name "*.go" | xargs grep -hE "//.*[가-힣]|/\*.*[가-힣]" | wc -l
# = 4,246 lines
```

### 2.2 String literal lines (보존 대상, EXCL-CCE-001)

```bash
find internal cmd pkg -name "*.go" | xargs grep -hE "\".*[가-힣].*\"" | wc -l
# = 2,186 lines
```

**Pre-SPEC baseline (cached for AC-CCE-004 use)**: `PRE_COUNT_FROM_BASELINE=2186` (inventory 2026-05-23).

**Important**: 본 SPEC은 **2,186 string literal Korean lines 중 N=14개 (Cobra exception, OQ-CCE-001 Option B 사용자 결정 2026-05-22)는 영어화하고, 나머지 2,172개는 보존한다**.

**Baseline breakdown table**:

| Row | Category | Count | Treatment by this SPEC |
|-----|----------|------:|------------------------|
| (a) | Total string literal Korean lines (pre-SPEC inventory) | 2,186 | — |
| (b) | Cobra exception delta (EXCL-CCE-001 exception, §6.1) | -14 | Englishify (per AC-CCE-010) |
| (c) | Expected post-SPEC count (a + b) | **2,172** | Preserve (per AC-CCE-004) |

**AC-CCE-004 verification formula**: `ACTUAL_POST == (PRE_COUNT_FROM_BASELINE - 14) == 2172`.

**Cached baseline rationale**: Stash-based pre-SPEC capture (`git stash --include-untracked` + grep + pop) is fragile when working tree is dirty — `git stash pop` 부분 적용 silent skip 결함이 알려져 있음 (CLAUDE.local.md §23.4). 따라서 본 SPEC AC verification 은 stash 방식 대신 본 §2.2 cached baseline (2186) 을 reference 값으로 사용한다.

### 2.3 @MX tag breakdown

```bash
grep -rn "@MX:" internal cmd pkg --include="*.go" | grep "[가-힣]" | wc -l
# = 139 total
```

Tag-level breakdown:

| Tag | Korean count |
|-----|--------------|
| @MX:NOTE | 58 |
| @MX:REASON | 40 |
| @MX:ANCHOR | 28 |
| @MX:WARN | 14 |
| @MX:TODO | 0 (none found with Korean) |
| **Total** | **140** (recount, original 139 was rounded) |

### 2.4 Identifier counts (AC-CCE-011 baseline)

```bash
SPEC_COUNT=$(grep -rohnE 'SPEC-[A-Z0-9-]+' internal/ cmd/ pkg/ --include="*.go" | sort -u | wc -l)
REQ_COUNT=$(grep -rohnE 'REQ-[A-Z]+-[0-9]+' internal/ cmd/ pkg/ --include="*.go" | sort -u | wc -l)
AC_COUNT=$(grep -rohnE 'AC-[A-Z]+-[0-9]+' internal/ cmd/ pkg/ --include="*.go" | sort -u | wc -l)
```

**Baseline (2026-05-23)**:

| Identifier | Unique count |
|-----------|--------------|
| SPEC-IDs | 1,733 |
| REQ-IDs | 1,444 |
| AC-IDs | 830 |

**AC-CCE-011 verification**: 본 SPEC 완료 후 위 3종 count가 동일해야 함.

---

## 3. Top 10 Files by Korean Comment Count

(Wave 분할 우선순위 + 큰 diff 예상 시점 식별)

| Rank | File | Korean comment lines | Wave |
|------|------|---------------------|------|
| 1 | internal/cli/glm_tools_test.go | 147 | Wave 5 (cli test) |
| 2 | internal/mx/resolver_query_test.go | 124 | Wave 6 또는 7 |
| 3 | internal/harness/rubric.go | 98 | Wave 3 (harness) |
| 4 | internal/cli/glm_tools.go | 98 | Wave 2 (cli) |
| 5 | internal/config/types.go | 85 | Wave 1 (foundation) |
| 6 | internal/spec/lint_test.go | 81 | Wave 7 (test) |
| 7 | internal/harness/applier_test.go | 81 | Wave 6 (harness test) |
| 8 | internal/template/agent_frontmatter_audit_test.go | 79 | Wave 5 (template test) |
| 9 | internal/harness/scorer_engine.go | 76 | Wave 3 (harness) |
| 10 | internal/harness/router/router.go | 59 | Wave 3 (harness) |

**Observation**:
- Top 5 files = ~552 Korean lines (13% of total)
- Wave 3 (harness) has 3 top-10 files (largest cluster)
- Wave 5/6 (test files) include 5 top-10 files

---

## 4. Sample Korean Comment Patterns

### 4.1 Type description (godoc)

```go
// Before (Korean):
// LSPReferencesClient는 LSPFanInCounter가 LSP 서버와 통신할 때 사용하는 인터페이스입니다.

// After (English):
// LSPReferencesClient is the interface used by LSPFanInCounter to communicate with the LSP server.
```

### 4.2 Function description

```go
// Before:
// FindReferences는 주어진 파일의 position에서 심볼의 모든 참조 위치를 반환합니다.

// After:
// FindReferences returns all reference locations for the symbol at the given file position.
```

### 4.3 State description

```go
// Before:
// IsAvailable은 LSP 서버가 현재 사용 가능한지 확인합니다.
// nil 클라이언트 또는 서버 미시작 상태에서 false를 반환합니다.

// After:
// IsAvailable checks whether the LSP server is currently available.
// Returns false when the client is nil or the server is not started.
```

### 4.4 @MX:NOTE pattern

```go
// Before:
// @MX:NOTE: [AUTO] LSPReferencesClient — core.Client 의존 없이 mx 패키지 내부에서 LSP 참조 질의를 추상화.

// After:
// @MX:NOTE: [AUTO] LSPReferencesClient — Abstracts LSP reference queries inside the mx package without depending on core.Client.
```

### 4.5 @MX:REASON pattern (fan_in invariant)

```go
// Before:
// @MX:REASON: fan_in >= 3 — Resolver.Resolve(), CLI mx_query.go, M6 sweep test 모두 이 구현체를 사용

// After:
// @MX:REASON: fan_in >= 3 — Resolver.Resolve(), CLI mx_query.go, and M6 sweep test all use this implementation
```

### 4.6 @MX:ANCHOR pattern

```go
// Before:
// @MX:ANCHOR: [AUTO] harness.yaml 전체 스키마의 Go 표현 — LoadHarnessConfig()의 반환 타입

// After:
// @MX:ANCHOR: [AUTO] Go representation of the full harness.yaml schema — return type of LoadHarnessConfig()
```

### 4.7 Mixed identifier + Korean (REQ-CCE-007 ambiguity case)

```go
// Before:
// REQ-DPL-011: 설계 헌법 §2, §3.1-3.3, §5, §11, §12 + 브랜드 컨텍스트 파일 보호.

// After:
// REQ-DPL-011: Design constitution §2, §3.1-3.3, §5, §11, §12 + brand context file protection.
```

Identifier `REQ-DPL-011` verbatim preserved per REQ-CCE-004.

### 4.8 TODO comment

```go
// Before:
StartedAt: nil, // TODO: 추적 추가

// After:
StartedAt: nil, // TODO: add tracking
```

```go
// Before:
// TODO: SPEC-V3R2-SCH-001 validator/v10 통합 후 구현

// After:
// TODO: implement after SPEC-V3R2-SCH-001 validator/v10 integration
```

SPEC-ID `SPEC-V3R2-SCH-001` verbatim preserved.

### 4.9 Block comment (single line)

```go
// Before:
/* 사용자 입력 검증 */

// After:
/* User input validation */
```

### 4.10 Test file comment

```go
// Before:
// TestValidationError_Error: ValidationError 문자열 표현 검증.

// After:
// TestValidationError_Error verifies the string representation of ValidationError.
```

(Godoc convention applied even for test functions.)

---

## 5. Pre-SPEC Baseline (AC-CCE-007/008 reference)

### 5.1 Test baseline

**Pre-existing FAILs (EXCL-CCE-008)**: per project memory `project_v3r6_harness_rename_001_run_complete` and recent SPEC project memories:

- `TestRuleTemplateMirrorDrift` — Template-First Rule drift (10 files, separately documented)
- `TestLateBranchTemplateMirror` — Late-Branch template residual
- `TestSkillsContainPlanAuditGateMarkers` — Plan audit sentinel
- HARNESS-RENAME-001 cascade: 3 test count assertions (33→37 / 52→60 / 52→60)
- WorktreeCreate hook V3R6 unwire baseline (commit `a3239d3de`)

**Baseline count**: ~6-8 pre-existing FAILs (exact count to be measured at Wave 1 entry via `go test ./...`).

**AC-CCE-007 verification**: NEW failures from this SPEC = 0. Baseline 잔존 허용.

### 5.2 Lint baseline

```bash
golangci-lint run --timeout=2m 2>&1 | tail -5
# Pre-SPEC baseline: ~22 issues (project_v3r6_harness_rename_001_sync_complete §Build×5 PASS + Lint 22)
```

**AC-CCE-008 verification**: NEW issues = 0. Pre-existing 22 baseline 잔존 허용.

### 5.3 Build baseline (must be 0)

```bash
go build ./...                       # exit 0 (currently)
GOOS=windows GOARCH=amd64 go build ./...  # exit 0 (currently)
```

AC-CCE-005/006 entry condition: build must already be 0. 본 SPEC introduced syntax error = 0 검증.

---

## 6. Risk Hotspots

### 6.1 Cobra command Korean (Wave 2 special attention) — EXCL-CCE-001 exception, N=14

**OQ-CCE-001 user decision (2026-05-22)**: **Option B 채택 — Cobra `Use:/Short:/Long:/Example:` 필드 영어화 (EXCL-CCE-001 예외).** AC-CCE-010이 영어화를 검증하고, AC-CCE-004가 본 14개 delta를 반영한 string literal count 차감(-14)을 검증한다.

**Exhaustive enumeration (verified 2026-05-23 via `grep -nE 'Use:|Short:|Long:|Example:' internal/cli/{mx,glm_tools,migration}.go | grep '[가-힣]'`)**:

| # | File | Line | Field | Korean text snippet |
|---|------|-----:|-------|---------------------|
| 1 | internal/cli/mx.go | 12 | Short | `"@MX TAG 관리 도구"` |
| 2 | internal/cli/mx.go | 13 | Long | `` `@MX TAG 사이드카 인덱스를 관리하고 조회하는 도구입니다.` `` |
| 3 | internal/cli/migration.go | 17 | Short | `"마이그레이션을 관리합니다 (run, status, rollback)"` |
| 4 | internal/cli/migration.go | 18 | Long | `` `마이그레이션 관리 도구입니다.` `` (multi-line) |
| 5 | internal/cli/migration.go | 36 | Short | `"Pending migrations를 실행합니다"` |
| 6 | internal/cli/migration.go | 37 | Long | `` `현재 프로젝트에서 아직 실행되지 않은 마이그레이션을 순서대로 실행합니다.` `` (multi-line) |
| 7 | internal/cli/migration.go | 67 | Short | `"마이그레이션 상태를 표시합니다"` |
| 8 | internal/cli/migration.go | 68 | Long | `` `현재 프로젝트의 마이그레이션 상태를 표시합니다.` `` (multi-line) |
| 9 | internal/cli/migration.go | 122 | Short | `"특정 버전으로 롤백합니다"` |
| 10 | internal/cli/migration.go | 123 | Long | `` `지정된 버전으로 롤백합니다.` `` (multi-line) |
| 11 | internal/cli/glm_tools.go | 67 | Short | `"Z.AI MCP 서버 도구 관리 (enable/disable)"` |
| 12 | internal/cli/glm_tools.go | 68 | Long | `` `Z.AI MCP 서버를 Claude Code 에 등록하거나 해제합니다.` `` (multi-line) |
| 13 | internal/cli/glm_tools.go | 87 | Short | `"Z.AI MCP 서버를 ~/.claude.json 에 등록"` |
| 14 | internal/cli/glm_tools.go | 95 | Short | `"Z.AI MCP 서버를 ~/.claude.json 에서 해제"` |

**Total: N=14 entries** (mx.go: 2 / migration.go: 8 / glm_tools.go: 4).

**Wave 2 진입 시점 처리**:

- 본 14개 항목은 AC-CCE-010 (Cobra English) 검증 대상 — 영어화 후 grep 0 matches 의무
- AC-CCE-004 (String literal count) baseline은 `PRE_COUNT_FROM_BASELINE - 14 = 2186 - 14 = 2172`
- Multi-line Long: 필드는 raw string (`` ` ` ``) 내부 multi-line이라 grep 시 첫 줄만 매치되지만, 영어화 시 전체 multi-line 한국어를 모두 영어로 치환해야 함
- Wave 2 PR body에 14개 영어화 결과 enumeration 의무 (review 용이성)

### 6.2 Test t.Errorf message Korean (EXCL-CCE-001 scope)

```
internal/cli/mx_query_test.go:49:		t.Errorf("Use: 기대 'query', 실제 %q", cmd.Use)
internal/cli/mx_query_test.go:295:		t.Errorf("Use: 기대 'mx', 실제 %q", cmd.Use)
```

**판정**: `t.Errorf("...")`의 한국어는 string literal → EXCL-CCE-001 적용 → **보존**.

**별도 SPEC 후보**: `SPEC-V3R6-TEST-MESSAGES-EN-001` (가칭) — test error message 영어화는 본 SPEC 범위 외.

### 6.3 @MX:REASON fan_in invariants (C-CCE-004)

```
internal/config/loader.go:270:// @MX:REASON: fan_in >= 3: HarnessRouter.Route, CLI validate, ConfigManager.Reload에서 호출
```

**Constraint**: 번역 후에도 fan_in invariant criterion (`fan_in >= 3`) 보존. 호출자 list 영어화 OK.

```go
// After:
// @MX:REASON: fan_in >= 3 — called by HarnessRouter.Route, CLI validate, and ConfigManager.Reload
```

### 6.4 Korean within Korean text (compound Korean expressions)

Sample:

```go
// "기대 'query', 실제 %q"  (t.Errorf 메시지, EXCL-CCE-001)
```

t.Errorf inside test functions — string literal 보존. 단, 함수 외부 godoc은 영어화:

```go
// Before:
// TestMxCmd_Use_Returns_query: Use 필드가 "query"인지 검증.

// After:
// TestMxCmd_Use_Returns_query verifies that the Use field equals "query".
```

---

## 7. Tooling Used in Inventory

| Tool | Purpose | Notes |
|------|---------|-------|
| `grep -rn '[가-힣]'` | Korean span discovery | Per-line scanner |
| `grep -rnE '/\*.*[가-힣]'` | Block comment discovery | Per-line (multi-line edge case 있음) |
| `grep -rohnE 'SPEC-...'` | Identifier extraction | sort -u for unique count |
| `wc -l` | Count aggregation | |
| `find ... | xargs grep -l` | File-level Korean filter | |
| `awk '/\/\*/,/\*\//'` | Multi-line block comment | Wave 진행 중 발견 시 사용 |

---

## 8. Cross-SPEC Dependencies

### 8.1 No upstream blockers

본 SPEC은 _comment-only_ migration — 코드 동작 변경 없음. 어떤 SPEC도 본 SPEC을 blocker로 갖지 않음.

### 8.2 Concurrent SPECs (잠재적 conflict)

| SPEC | Status | Conflict risk |
|------|--------|---------------|
| HARNESS-RENAME-001 | implemented (merged 2026-05-22) | None (이미 머지됨) |
| TEMPLATE-MIRROR-DRIFT-001 (provisional) | Wave 1 후보 | None (template 디렉토리 외) |
| HARNESS-USER-AREA-RESOLUTION-001 (provisional) | Wave 1 후보 | None (separate scope) |

### 8.3 Downstream SPECs (후속 권장)

본 SPEC 완료 후 권장 후속 SPEC:

- `SPEC-V3R6-TEST-MESSAGES-EN-001` (가칭): test `t.Errorf` Korean message 영어화 (EXCL-CCE-007 처리)
- `SPEC-V3R6-CLI-USER-MESSAGES-EN-001` (가칭): CLI `fmt.Println` / `slog.Info` user-facing 한국어 string 처리 (EXCL-CCE-010, OQ-CCE-002/003)
- `SPEC-V3R6-COBRA-DOCS-EN-001` (가칭): OQ-CCE-001 Option A 채택 시 별도 SPEC

---

## 9. Estimation Confidence

### 9.1 Translation effort estimation

- **Per-comment block average**: 2-4 lines (godoc 평균)
- **Total blocks**: ~4,246 / 3 = ~1,400 translation units
- **Per-block agent time**: ~1-3 seconds (Read+Translate+Edit)
- **Per-Wave time** (without Agent Teams):
  - Wave 1 (~250 lines, ~80 blocks): rough estimate available but no time prediction (per agent-common-protocol §Time Estimation)
  - Wave 5 (~1,250 lines, ~400 blocks): largest

**Confidence**: Medium-High. Agent translation은 deterministic하고 검증 가능 (binary AC).

### 9.2 Wave order confidence

Wave 1-7 순서는 의존성 그래프 + 영향 표면에 따라 정렬. 변경 가능:

- Wave 2 (CLI) 우선화: user-facing impact 큰 패키지
- Wave 3 (Harness) baseline-aware: HARNESS-RENAME-001 cascade 후 진입

**Confidence**: High. plan.md §3 Wave breakdown justification 명시.

---

## 10. Open Decisions (Pre-Run)

(Plan.md §7과 cross-reference)

| Open Question | Default | Reviewer / User Decision Required? |
|---------------|---------|-----------------------------------|
| OQ-CCE-001 Cobra Short/Long Korean | **RESOLVED**: Option B (영어화, AC-CCE-010 active) — verified 14 entries (mx.go 2 / migration.go 8 / glm_tools.go 4) per §6.1 | NO (user-resolved 2026-05-22) |
| OQ-CCE-002 error.New Korean | Default 보존 (EXCL-CCE-001), 별도 SPEC 후속 | NO (default acceptable) |
| OQ-CCE-003 Log message Korean | Default 보존 (EXCL-CCE-001), 별도 SPEC 후속 | NO (default acceptable) |
| OQ-CCE-004 main 직진 vs feat-branch | feat-branch + auto PR | NO (default acceptable, CLAUDE.local.md §23 Tier L 권장) |

---

## 11. Validation Script Sample

### 11.1 Pre-Wave baseline capture

```bash
#!/bin/bash
# pre-wave-baseline.sh — Wave 진입 전 baseline 기록
set -e
WAVE_NUM="$1"
BASELINE_DIR=".moai/research/wave-${WAVE_NUM}-baseline-$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -p "$BASELINE_DIR"

# Korean count baseline
grep -rn '//.*[가-힣]\|/\*.*[가-힣].*\*/' internal/ cmd/ pkg/ --include="*.go" > "$BASELINE_DIR/comment-korean.txt" || true
grep -rn '".*[가-힣].*"' internal/ cmd/ pkg/ --include="*.go" > "$BASELINE_DIR/string-literal-korean.txt" || true
grep -rn '@MX:[A-Z]*[: ].*[가-힣]' internal/ cmd/ pkg/ --include="*.go" > "$BASELINE_DIR/mx-korean.txt" || true

# Identifier baseline
grep -rohnE 'SPEC-[A-Z0-9-]+' internal/ cmd/ pkg/ --include="*.go" | sort -u > "$BASELINE_DIR/spec-ids.txt"
grep -rohnE 'REQ-[A-Z]+-[0-9]+' internal/ cmd/ pkg/ --include="*.go" | sort -u > "$BASELINE_DIR/req-ids.txt"
grep -rohnE 'AC-[A-Z]+-[0-9]+' internal/ cmd/ pkg/ --include="*.go" | sort -u > "$BASELINE_DIR/ac-ids.txt"

# Lint baseline
golangci-lint run --timeout=2m 2>&1 | tee "$BASELINE_DIR/lint.log" || true

# Test baseline
go test ./... 2>&1 | tee "$BASELINE_DIR/test.log" || true

echo "Baseline captured: $BASELINE_DIR"
wc -l "$BASELINE_DIR/comment-korean.txt" "$BASELINE_DIR/string-literal-korean.txt" "$BASELINE_DIR/mx-korean.txt"
wc -l "$BASELINE_DIR/spec-ids.txt" "$BASELINE_DIR/req-ids.txt" "$BASELINE_DIR/ac-ids.txt"
```

### 11.2 Post-Wave verification

```bash
#!/bin/bash
# post-wave-verify.sh — Wave 완료 후 AC matrix 검증
set -e
WAVE_NUM="$1"
WAVE_SCOPE="$2"  # e.g., "internal/cli"

# AC-CCE-001/002/003 (Wave scope grep)
COMMENT_KO=$(grep -rn '//.*[가-힣]\|/\*.*[가-힣].*\*/' "$WAVE_SCOPE" --include="*.go" | wc -l)
MX_KO=$(grep -rn '@MX:[A-Z]*[: ].*[가-힣]' "$WAVE_SCOPE" --include="*.go" | wc -l)
echo "AC-CCE-001/002 Comment Korean in scope: $COMMENT_KO (expect 0)"
echo "AC-CCE-003 @MX Korean in scope: $MX_KO (expect 0)"

# AC-CCE-005/006 Build
go build ./... && echo "AC-CCE-005 darwin: PASS" || { echo "AC-CCE-005 FAIL"; exit 1; }
GOOS=windows GOARCH=amd64 go build ./... && echo "AC-CCE-006 windows: PASS" || { echo "AC-CCE-006 FAIL"; exit 1; }

# AC-CCE-011 Identifier preservation
POST_SPEC=$(grep -rohnE 'SPEC-[A-Z0-9-]+' internal/ cmd/ pkg/ --include="*.go" | sort -u | wc -l)
POST_REQ=$(grep -rohnE 'REQ-[A-Z]+-[0-9]+' internal/ cmd/ pkg/ --include="*.go" | sort -u | wc -l)
POST_AC=$(grep -rohnE 'AC-[A-Z]+-[0-9]+' internal/ cmd/ pkg/ --include="*.go" | sort -u | wc -l)
echo "AC-CCE-011 Post SPEC-IDs: $POST_SPEC (baseline: 1733)"
echo "AC-CCE-011 Post REQ-IDs: $POST_REQ (baseline: 1444)"
echo "AC-CCE-011 Post AC-IDs: $POST_AC (baseline: 830)"
test "$POST_SPEC" = "1733" || { echo "AC-CCE-011 FAIL: SPEC-ID drift"; exit 1; }
test "$POST_REQ" = "1444" || { echo "AC-CCE-011 FAIL: REQ-ID drift"; exit 1; }
test "$POST_AC" = "830" || { echo "AC-CCE-011 FAIL: AC-ID drift"; exit 1; }

# AC-CCE-012 Diff scope
NON_GO=$(git diff --name-only origin/main...HEAD | grep -v '\.go$' | grep -v "SPEC-V3R6-CODE-COMMENTS-EN-001" | wc -l)
echo "AC-CCE-012 Non-.go diff (excluding SPEC dir): $NON_GO (expect 0)"
test "$NON_GO" = "0" || { echo "AC-CCE-012 FAIL: non-.go drift"; exit 1; }

echo "All AC PASS for Wave $WAVE_NUM ($WAVE_SCOPE)"
```

---

## 12. Cross-references

- [spec.md](./spec.md) — Requirements + scope + EARS
- [plan.md](./plan.md) — 7-wave execution plan + Section A-E template
- [acceptance.md](./acceptance.md) — Binary AC matrix
- [design.md](./design.md) — Translation methodology + Agent strategy
- Project memories: `project_v3r6_harness_rename_001_sync_complete`, `project_v3r6_template_mirror_drift_audit_2026_05_22`
- Lessons memories: `lessons #B4 Frontmatter`, `lessons #B6 spec-lint heading`, `lessons #B8 working tree hygiene`

---

Version: 0.2.0
Status: draft
Inventory date: 2026-05-23
Total scope: 267 Go files (96 non-test + 171 test) / 4,246 comment lines / 139 @MX tags / 2,186 string literals (2,172 preserved, 14 Cobra exception Englishified)
Identifier baseline: 1,733 SPEC-IDs / 1,444 REQ-IDs / 830 AC-IDs (must preserve via AC-CCE-011)
Cobra exception (EXCL-CCE-001 exception, §6.1): N=14 enumerated entries (mx.go 2 / migration.go 8 / glm_tools.go 4)
