---
id: SPEC-V3R6-CODE-COMMENTS-EN-001
title: "Acceptance criteria — Mass migration of Korean comments to English"
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
type: acceptance
---

# Acceptance Criteria — SPEC-V3R6-CODE-COMMENTS-EN-001

## 1. Acceptance Criteria Matrix

12 binary AC (PASS/FAIL only, no partial credit). REQ ↔ AC traceability 100%.

### 1.1 Out of Scope

본 SPEC의 acceptance verification은 다음 조건 충족 시 "out of scope" 처리한다:

- **EXCL-AC-001**: 한국어 string literal (`".*[가-힣].*"`, `` `.*[가-힣].*` ``) preservation check는 EXCL-CCE-001 적용. AC-CCE-004로 보존만 검증, 영어화 검증은 안 함.
- **EXCL-AC-002**: `.moai/`, `docs-site/`, `.claude/agents|skills|rules|commands|output-styles`, `internal/template/templates/.claude/` 디렉토리는 검증 범위 외 (EXCL-CCE-002..006).
- **EXCL-AC-003**: 사전 존재 baseline test failures (HARNESS-RENAME-001 cascade, WorktreeCreate hook V3R6 unwire 잔존)는 본 SPEC introduced 아니므로 AC-CCE-007에서 baseline 잔존 허용 (EXCL-CCE-008).
- **EXCL-AC-004**: Generated files (`internal/template/embedded.go`, etc.)는 source-of-truth가 아니므로 검증 범위 외 — 본 SPEC은 hand-written source files만 대상.
- **EXCL-AC-005**: Vendor directories (`vendor/`, `node_modules/`) 검증 범위 외.

---

## 2. Acceptance Criteria Detail

### AC-CCE-001 — Go line comments Korean 0 matches

**REQ link**: REQ-CCE-001 (Ubiquitous)

**Description**: `internal/`, `cmd/`, `pkg/` 모든 Go source files (test 포함)에서 한국어 line comment (`//.*[가-힣]`) grep 결과 0건.

**Verification command**:

```bash
grep -rn '//.*[가-힣]' internal/ cmd/ pkg/ --include="*.go" | wc -l
```

**Expected output**: `0`

**Pass condition**: exit 0 AND wc -l output = `0`

**Status**: PASS / FAIL

---

### AC-CCE-002 — Go block comments Korean 0 matches (single-line AND multi-line)

**REQ link**: REQ-CCE-001 (Ubiquitous)

**Description**: Block comment (`/* ... */`) 내부 한국어 0건. **Single-line AND multi-line 모두 검증**.

**Verification command** (canonical, multi-line aware via perl `-0777` slurp + non-greedy regex):

```bash
# Multi-line block comment Korean detection (perl slurp mode):
COUNT_BLOCK_KOREAN=$(find internal/ cmd/ pkg/ -name '*.go' -print0 | xargs -0 perl -0777 -ne '
  while (/\/\*.*?\*\//gs) {
    print "$&\n" if /[\x{AC00}-\x{D7A3}\x{1100}-\x{11FF}\x{3130}-\x{318F}]/;
  }
' | wc -l)

echo "Multi-line block comment Korean count: $COUNT_BLOCK_KOREAN"
test "$COUNT_BLOCK_KOREAN" = "0" && echo "PASS" || echo "FAIL"
```

**Alternative awk-based verification** (for environments without perl unicode support):

```bash
# awk multi-line block comment scanner — accumulates block, then checks Korean
COUNT_BLOCK_KOREAN_AWK=$(find internal/ cmd/ pkg/ -name '*.go' | while read f; do
  awk '
    /\/\*/ { in_block=1; block="" }
    in_block { block = block $0 "\n" }
    /\*\// { in_block=0; if (block ~ /[가-힣]/) print "MATCH:" FILENAME ":" block; block="" }
  ' "$f"
done | wc -l)

echo "awk block comment Korean count: $COUNT_BLOCK_KOREAN_AWK"
```

**Expected output**: `Multi-line block comment Korean count: 0` / `PASS`

**Pass condition**: perl slurp count = `0` (multi-line aware — captures both single-line `/* X */` and multi-line `/* ... \n X \n ... */` patterns).

**Why this matters**: The original `grep -rn '/\*.*[가-힣].*\*/'` only matches block comments where the entire `/* ... */` fits on a single line. Multi-line block comments like:

```go
/*
   한국어 다중 라인
   주석
*/
```

would silently pass single-line grep but fail proper code-quality compliance. Perl's `-0777 -ne` (slurp mode) + non-greedy `.*?` with `/s` flag (single-line dot-matches-newline) correctly handles both forms. This addresses Edge Case 1 (§3.1) as a verified AC, not just an acknowledged corner case.

**Status**: PASS / FAIL

---

### AC-CCE-003 — @MX tag descriptions Korean 0 matches

**REQ link**: REQ-CCE-002 (Ubiquitous)

**Description**: `@MX:NOTE`, `@MX:WARN`, `@MX:REASON`, `@MX:ANCHOR`, `@MX:TODO` tag 뒤 한국어 0건.

**Verification command**:

```bash
grep -rn '@MX:[A-Z]*[: ].*[가-힣]' internal/ cmd/ pkg/ --include="*.go" | wc -l
```

**Expected output**: `0`

**Pass condition**: wc -l output = `0`

**Baseline (pre-migration)**: 139 @MX Korean tags

**Status**: PASS / FAIL

---

### AC-CCE-004 — String literal Korean count preserved (minus Cobra exception delta)

**REQ link**: REQ-CCE-005 (State-Driven), EXCL-CCE-001 (with documented Cobra exception)

**Description**: 본 SPEC은 EXCL-CCE-001로 한국어 string literal을 원칙적으로 보존하지만, **Cobra command 정의의 `Use:/Short:/Long:/Example:` 필드는 N=14개 항목 영어화**한다 (OQ-CCE-001 사용자 결정 2026-05-22 + spec.md §5.1 EXCL-CCE-001 예외 참조). 따라서 종료 시점 string literal Korean count는 시작 시점 count보다 정확히 N=14 작아야 한다.

**Cobra exception delta**: **N=14** (verified inventory 2026-05-23):
- `internal/cli/mx.go`: 2 entries (lines 12, 13)
- `internal/cli/migration.go`: 8 entries (lines 17, 18, 36, 37, 67, 68, 122, 123)
- `internal/cli/glm_tools.go`: 4 entries (lines 67, 68, 87, 95)

상세 enumeration은 research.md §6.1 참조.

**Pre-SPEC baseline (cached)**: research.md §2.2 baseline에서 가져온 인벤토리 시점(2026-05-23) 측정값. Stash+pop 방식은 working tree dirty 상태에서 silently 일부 파일만 회복되는 결함(`git stash pop` 부분 적용 silent skip, lessons memory `CLAUDE.local.md` §23.4)이 있어 **사용 금지**. 대신 baseline value `PRE_COUNT_FROM_BASELINE=2186`를 research.md §2.2에서 직접 참조.

**Verification command**:

```bash
# Step 1: Pre-SPEC count cached from research.md §2.2 baseline (inventory 2026-05-23):
PRE_COUNT_FROM_BASELINE=2186

# Step 2: Cobra exception delta (per OQ-CCE-001 user decision 2026-05-22):
COBRA_EXCEPTION_DELTA=14

# Step 3: Expected post count
EXPECTED_POST=$((PRE_COUNT_FROM_BASELINE - COBRA_EXCEPTION_DELTA))  # = 2172

# Step 4: Post-Wave count (live measurement)
ACTUAL_POST=$(grep -rn '".*[가-힣].*"' internal/ cmd/ pkg/ --include="*.go" | wc -l)

echo "Pre baseline (research.md §2.2): $PRE_COUNT_FROM_BASELINE"
echo "Cobra exception delta (EXCL-CCE-001 exception): -$COBRA_EXCEPTION_DELTA"
echo "Expected post: $EXPECTED_POST"
echo "Actual post: $ACTUAL_POST"
test "$ACTUAL_POST" -eq "$EXPECTED_POST" && echo "PASS" || echo "FAIL"
```

**Expected output**: `Pre baseline (research.md §2.2): 2186` / `Cobra exception delta (EXCL-CCE-001 exception): -14` / `Expected post: 2172` / `Actual post: 2172` / `PASS`

**Pass condition**: `ACTUAL_POST == (PRE_COUNT_FROM_BASELINE - 14)` (exact integer equality)

**Cobra exception delta note**: AC-CCE-010이 본 14개 항목의 영어화를 별도로 검증한다. AC-CCE-004는 보존(=2,172) + 영어화(=14) 합계가 2,186과 일치함을 검증함으로써 EXCL-CCE-001과 그 예외의 일관성을 확인한다.

**Status**: PASS / FAIL

---

### AC-CCE-005 — `go build ./...` exit 0

**REQ link**: REQ-CCE-008 (Ubiquitous), C-CCE-003

**Description**: 모든 패키지 build 성공. 본 SPEC이 Go syntax/dependency를 깨지 않음 검증.

**Verification command**:

```bash
go build ./...
echo "Exit: $?"
```

**Expected output**: `Exit: 0`

**Pass condition**: exit code = 0

**Status**: PASS / FAIL

---

### AC-CCE-006 — Cross-platform build PASS

**REQ link**: REQ-CCE-008 (Ubiquitous), C-CCE-003

**Description**: Windows / Linux / macOS 3종 build PASS.

**Verification command**:

```bash
GOOS=darwin GOARCH=amd64 go build ./... && echo "darwin: PASS" || echo "darwin: FAIL"
GOOS=linux GOARCH=amd64 go build ./...  && echo "linux: PASS"  || echo "linux: FAIL"
GOOS=windows GOARCH=amd64 go build ./... && echo "windows: PASS" || echo "windows: FAIL"
```

**Expected output**: 3종 PASS

**Pass condition**: 3종 모두 exit 0

**Status**: PASS / FAIL

---

### AC-CCE-007 — `go test ./...` baseline preserved

**REQ link**: REQ-CCE-008 (Ubiquitous), EXCL-CCE-008

**Description**: Test pass count delta NEW failures = 0. 사전 존재 baseline residual (HARNESS-RENAME-001 cascade 등) 잔존 허용.

**Verification command**:

```bash
# Pre-SPEC test baseline (recorded in research.md §5.1):
# Baseline: <pre-SPEC pass count>, <pre-SPEC fail count>

# Post-Wave test:
go test ./... 2>&1 | grep -cE "^--- FAIL|^FAIL\s" > /tmp/post-fail
go test ./... 2>&1 | grep -cE "^--- PASS|^PASS\s+ok" > /tmp/post-pass

echo "Post FAILs: $(cat /tmp/post-fail)"
echo "Pre FAILs (baseline): <from research.md>"
# Delta MUST be <= 0 (NEW failures = 0)
```

**Expected**: Post FAILs <= Pre FAILs (baseline)

**Pass condition**: NEW failures (post - pre) = 0

**Baseline 예외 사유**: EXCL-CCE-008 (HARNESS-RENAME-001 cascade 3 FAIL + WorktreeCreate hook V3R6 unwire baseline residual)

**Status**: PASS / FAIL (with baseline residual notation)

---

### AC-CCE-008 — `golangci-lint` NEW issues = 0

**REQ link**: REQ-CCE-008 (Ubiquitous)

**Description**: Linter NEW issues 0건. Pre-existing baseline은 잔존 허용.

**Verification command**:

```bash
# Pre-SPEC lint baseline (recorded in research.md §5.2):
# Baseline: <pre-SPEC issue count>

# Post-Wave lint:
golangci-lint run --timeout=2m 2>&1 | tee /tmp/lint-post.log
POST_ISSUES=$(grep -cE "^[a-zA-Z_/].*:[0-9]+:[0-9]+:" /tmp/lint-post.log)

echo "Post issues: $POST_ISSUES"
echo "Pre baseline: <from research.md>"
# Delta MUST be <= 0
```

**Expected**: POST_ISSUES <= Pre baseline

**Pass condition**: NEW issues = 0

**Status**: PASS / FAIL

---

### AC-CCE-009 — Sample manual review semantic preservation

**REQ link**: REQ-CCE-001/002/003 (Ubiquitous), R-CCE-001 mitigation

**Description**: 5 random touched files에서 reviewer가 영어 주석의 의미가 원본 한국어 의미를 보존하는지 평가.

**Verification procedure** (manual):

```bash
# Wave 완료 후 random sampling
git diff --name-only HEAD~1 HEAD | grep "\.go$" | shuf -n 5 > /tmp/sample-files.txt

for f in $(cat /tmp/sample-files.txt); do
  echo "=== $f ==="
  git diff HEAD~1 HEAD -- "$f" | head -50
done
```

**Pass condition**:
- 5개 모두 reviewer가 "의미 보존됨" 판정
- 한 건이라도 "의미 손실" 판정 시 해당 Wave PR revert + 재실행

**Output**: Reviewer report (semantic preservation matrix per file)

**Status**: PASS / FAIL

---

### AC-CCE-010 — Cobra command descriptions in English (RESOLVED: Option B active)

**REQ link**: REQ-CCE-001 (Ubiquitous), OQ-CCE-001 (RESOLVED 2026-05-22)

**Description**: `internal/cli/**/*.go`의 Cobra command 정의 (`Use:`, `Short:`, `Long:`, `Example:`) **영어 (OQ-CCE-001 Option B 사용자 결정 2026-05-22)**. 본 SPEC의 EXCL-CCE-001 예외로서 영어화 N=14개 항목이 영어화되어야 한다 (research.md §6.1 enumeration 참조).

**OQ-CCE-001 resolution**: Option **B** (영어화) — 사용자 결정 2026-05-22. AC-CCE-010 status now **active** (NOT skipped). 본 SPEC v0.2.0부터 default = Option B.

**Verification command**:

```bash
# Count any remaining Korean in Cobra fields (expect 0 after Wave 2 completion):
COBRA_KO=$(grep -nE 'Use:|Short:|Long:|Example:' internal/cli --include="*.go" -r | grep '[가-힣]' | wc -l)

echo "Cobra fields with Korean: $COBRA_KO (expect 0 after Wave 2)"
test "$COBRA_KO" = "0" && echo "AC-CCE-010 PASS" || echo "AC-CCE-010 FAIL"
```

**Expected output**: `Cobra fields with Korean: 0` / `AC-CCE-010 PASS`

**Pass condition**: `grep ... | grep '[가-힣]' | wc -l` = `0` after Wave 2 完료.

**Pre-Wave 2 baseline**: 14 entries (research.md §6.1 — verified 2026-05-23). Post-Wave 2 expected: 0.

**Note**: This AC complements AC-CCE-004 — AC-CCE-010 verifies the 14 Cobra entries are Englishified (delta direction), AC-CCE-004 verifies the residual string literal Korean count is preserved at 2,172 (= 2,186 − 14).

**Status**: PASS / FAIL (active, no longer SKIPPED)

---

### AC-CCE-011 — SPEC-ID / REQ-ID identifier verbatim preservation

**REQ link**: REQ-CCE-004 (Event-Driven), R-CCE-002 mitigation

**Description**: SPEC-ID (`SPEC-[A-Z0-9-]+`), REQ-ID (`REQ-[A-Z]+-\d+`), AC-ID (`AC-[A-Z]+-\d+`), MEMO-ID 등 identifier count는 변경 전후 동일.

**Verification command**:

```bash
# Pre-SPEC identifier count (recorded in research.md §3):
# PRE_SPEC_COUNT, PRE_REQ_COUNT, PRE_AC_COUNT

# Post-Wave:
POST_SPEC=$(grep -rohnE 'SPEC-[A-Z0-9-]+' internal/ cmd/ pkg/ --include="*.go" | sort -u | wc -l)
POST_REQ=$(grep -rohnE 'REQ-[A-Z]+-[0-9]+' internal/ cmd/ pkg/ --include="*.go" | sort -u | wc -l)
POST_AC=$(grep -rohnE 'AC-[A-Z]+-[0-9]+' internal/ cmd/ pkg/ --include="*.go" | sort -u | wc -l)

echo "Post SPEC-IDs: $POST_SPEC"
echo "Post REQ-IDs: $POST_REQ"
echo "Post AC-IDs: $POST_AC"
echo "Pre baseline: <from research.md §3>"
```

**Pass condition**: POST counts == PRE counts (exact integer equality per category)

**Status**: PASS / FAIL

---

### AC-CCE-012 — git diff scope limited to `*.go` files only

**REQ link**: C-CCE-005, REQ-CCE-008

**Description**: 본 SPEC의 git diff는 `*.go` 확장자만 포함. `go.mod`, `go.sum`, template files, config files 변경 0.

**Verification command** (per Wave):

```bash
git diff --stat origin/main...HEAD | tail -1
# OR per-Wave:
git diff --name-only origin/main...HEAD | grep -v "\.go$" | wc -l
```

**Expected output**: `0` (non-.go files)

**Pass condition**: `git diff --name-only origin/main...HEAD | grep -v "\.go$" | wc -l` = `0`

**Exception**: `.moai/specs/SPEC-V3R6-CODE-COMMENTS-EN-001/*.md` (이 SPEC 본문) 및 `.moai/specs/SPEC-V3R6-CODE-COMMENTS-EN-001/progress.md`는 SPEC 작업 산출물 — 별도 처리 (chore commit 또는 plan-phase commit).

**Status**: PASS / FAIL

---

## 3. Edge Cases

### 3.1 Edge Case 1 — Multi-line block comments (handled by AC-CCE-002 perl slurp scanner)

```go
/*
한국어 주석 다중 라인
계속되는 라인
*/
```

**처리 (RESOLVED)**: AC-CCE-002의 verification command가 perl `-0777 -ne` slurp mode + non-greedy `/\/\*.*?\*\//gs` 패턴으로 multi-line block comment 한국어를 정확히 탐지한다. 별도 manual review 불필요 — binary AC로 처리. awk fallback도 동일 검증 제공.

**검증 명령** (AC-CCE-002 verification command와 동일):

```bash
find internal/ cmd/ pkg/ -name '*.go' -print0 | xargs -0 perl -0777 -ne '
  while (/\/\*.*?\*\//gs) {
    print "$&\n" if /[\x{AC00}-\x{D7A3}\x{1100}-\x{11FF}\x{3130}-\x{318F}]/;
  }
' | wc -l   # expect 0
```

**근거**: 단일 라인 grep `/\*.*[가-힣].*\*/` 은 multi-line block 을 silently 누락하여 정책 위반을 숨길 수 있음. Perl slurp + non-greedy regex 가 multi-line 을 명시적으로 커버.

### 3.2 Edge Case 2 — Embedded Korean in code identifiers (절대 발생 X but)

```go
var 변수명 = ...
func 함수명() { ... }
```

**판정**: 본 프로젝트는 Go OSS 코드 — 한국어 식별자 사용 없음 (확인 필요). 발견 시 BLOCKER 보고 (refactoring 별도 SPEC 필요).

**검증**: `grep -rn 'func [가-힣]\|var [가-힣]\|type [가-힣]' internal/ cmd/ pkg/ --include="*.go" | wc -l` → `0` 기대.

### 3.3 Edge Case 3 — Mixed Korean+English comment

```go
// REQ-CCE-001 처리: 한국어가 영어 식별자와 섞임
```

**처리**: REQ-CCE-007 (Unwanted) 적용 — clarity 우선. 식별자 (REQ-CCE-001) 보존 + 한국어 부분 영어화:

```go
// REQ-CCE-001 handling: Korean mixed with English identifier
```

### 3.4 Edge Case 4 — Korean variable name in comment

```go
// 사용자_이름 변수의 길이 검증
```

→ 변수명이 영어 (`userName`)이면:

```go
// Validate length of userName variable
```

→ 변수명도 한국어이면 Edge Case 2 (BLOCKER).

### 3.5 Edge Case 5 — Comment-only files (no code)

`doc.go` 등 package-level documentation files. Full file가 godoc 주석으로 구성될 수 있음. REQ-CCE-003 적용 — 전체 영어화.

---

## 4. Definition of Done (DoD)

본 SPEC 완료 시 다음 7개 모두 충족:

1. **All 7 Waves merged**: Wave 1-7 PR 모두 main에 머지 완료
2. **Final AC verification PASS**:
   - AC-CCE-001/002/003 global grep 0 matches
   - AC-CCE-004 string literal count 보존
   - AC-CCE-005/006 cross-platform build PASS
   - AC-CCE-007 test baseline 보존
   - AC-CCE-008 lint NEW = 0
   - AC-CCE-009 sample manual review PASS (final)
   - AC-CCE-010 PASS (OQ-CCE-001 Option B 사용자 결정 2026-05-22, 14 entries Englishified)
   - AC-CCE-011 identifier count 보존
   - AC-CCE-012 diff scope `*.go` only
3. **SPEC status updated**: `spec.md` frontmatter `status: draft → completed`, version `0.1.0 → 1.0.0`
4. **Documentation sync**: `CHANGELOG.md`에 본 SPEC 항목 추가 (Tier L 변경사항)
5. **Progress record**: 본 SPEC 디렉토리에 `progress.md` 작성 (각 Wave 결과 + final summary)
6. **MEMORY.md update**: 본 SPEC 완료 시 project memory entry 추가 (paste-ready resume + lessons)
7. **No regressions introduced**:
   - C-HRA-008 subagent boundary 0 violations (sentinel test)
   - @MX:ANCHOR fan_in invariants 유지
   - Coverage delta acceptable (per-package ≥ 85% baseline 보존)

---

## 5. REQ ↔ AC Traceability Matrix

| REQ | Pattern | ACs |
|-----|---------|-----|
| REQ-CCE-001 | Ubiquitous (line/block comments English) | AC-CCE-001, AC-CCE-002, AC-CCE-007, AC-CCE-009, AC-CCE-010 |
| REQ-CCE-002 | Ubiquitous (@MX tag descriptions English) | AC-CCE-003, AC-CCE-009 |
| REQ-CCE-003 | Ubiquitous (godoc English) | AC-CCE-001, AC-CCE-009 |
| REQ-CCE-004 | Event-Driven (identifier verbatim) | AC-CCE-011 |
| REQ-CCE-005 | State-Driven (string literal preservation) | AC-CCE-004 |
| REQ-CCE-006 | Optional (technical term preservation) | AC-CCE-009 (manual review) |
| REQ-CCE-007 | Unwanted (ambiguity → clarity) | AC-CCE-009 (manual review) |
| REQ-CCE-008 | Ubiquitous (Go syntax preservation) | AC-CCE-005, AC-CCE-006, AC-CCE-007, AC-CCE-008, AC-CCE-012 |

**Coverage**: 100% — 모든 REQ가 최소 1개 AC로 검증됨.

---

## 6. Verification Tool Reference

### 6.1 Read-only Parallel Verification Batch

per `.claude/rules/moai/workflow/verification-batch-pattern.md` § 7-item canonical:

```bash
# Group A — Functional
go test ./...
go test -coverprofile=cover.out ./internal/<wave-pkg>/...

# Group B — Boundary
grep -rn 'AskUserQuestion\|mcp__askuser' internal/ | grep -v "_test.go" | grep -v "// "

# Group C — Quality
golangci-lint run --timeout=2m

# Group D — Smoke
go run ./cmd/moai --version

# In-SPEC additions:
grep -rn '//.*[가-힣]' internal/ cmd/ pkg/ --include="*.go" | wc -l            # AC-CCE-001
# AC-CCE-002 multi-line aware (perl slurp):
find internal/ cmd/ pkg/ -name '*.go' -print0 | xargs -0 perl -0777 -ne 'while (/\/\*.*?\*\//gs) { print "$&\n" if /[\x{AC00}-\x{D7A3}\x{1100}-\x{11FF}\x{3130}-\x{318F}]/; }' | wc -l
grep -rn '@MX:[A-Z]*[: ].*[가-힣]' internal/ cmd/ pkg/ --include="*.go" | wc -l # AC-CCE-003
grep -rn '".*[가-힣].*"' internal/ cmd/ pkg/ --include="*.go" | wc -l         # AC-CCE-004 (count for preservation, expected: 2172 = 2186 - 14 Cobra delta)
```

오케스트레이터는 위 명령들을 **single-turn multi-Bash 병렬 호출**로 실행 (per agent-common-protocol §Parallel Execution).

### 6.2 Per-Wave Verification Script Template

```bash
#!/bin/bash
# wave-N-verify.sh
set -e

WAVE_SCOPE="$1"  # e.g., "internal/cli"

echo "=== AC-CCE-001 (Line comments Korean) ==="
LINE_COMMENT_KO=$(grep -rn '//.*[가-힣]' "$WAVE_SCOPE" --include="*.go" 2>/dev/null | wc -l)
echo "Line comment Korean: $LINE_COMMENT_KO (expect 0)"

echo "=== AC-CCE-002 (Block comments Korean — multi-line aware via perl slurp) ==="
BLOCK_COMMENT_KO=$(find "$WAVE_SCOPE" -name '*.go' -print0 2>/dev/null | xargs -0 perl -0777 -ne '
  while (/\/\*.*?\*\//gs) {
    print "$&\n" if /[\x{AC00}-\x{D7A3}\x{1100}-\x{11FF}\x{3130}-\x{318F}]/;
  }
' 2>/dev/null | wc -l)
echo "Block comment Korean (multi-line aware): $BLOCK_COMMENT_KO (expect 0)"

echo "=== AC-CCE-003 (@MX Korean) ==="
MX_KO=$(grep -rn '@MX:[A-Z]*[: ].*[가-힣]' "$WAVE_SCOPE" --include="*.go" 2>/dev/null | wc -l)
echo "@MX Korean: $MX_KO (expect 0)"

test "$LINE_COMMENT_KO" = "0" && test "$BLOCK_COMMENT_KO" = "0" && test "$MX_KO" = "0" || { echo "FAIL"; exit 1; }

echo "=== AC-CCE-005/006 (Build) ==="
go build ./... || { echo "darwin build FAIL"; exit 1; }
GOOS=windows GOARCH=amd64 go build ./... || { echo "windows build FAIL"; exit 1; }
echo "Build PASS"

echo "=== AC-CCE-012 (Diff scope) ==="
NON_GO=$(git diff --name-only origin/main...HEAD | grep -v '\.go$' | grep -v "SPEC-V3R6-CODE-COMMENTS-EN-001" | wc -l)
echo "Non-.go diff (excluding SPEC dir): $NON_GO (expect 0)"
test "$NON_GO" = "0" || { echo "FAIL"; exit 1; }

echo "All AC PASS for Wave: $WAVE_SCOPE"
```

---

Version: 0.2.0
Status: draft
AC count: 12 binary + 5 edge cases + 7-item DoD
Traceability: 100% REQ ↔ AC coverage
AC-CCE-002: multi-line block comment perl `-0777` slurp scanner (handles Edge Case 1)
AC-CCE-004: PRE_COUNT (2186 cached baseline) − 14 Cobra exception delta = expected post 2172
