---
id: SPEC-V3R6-CODE-COMMENTS-EN-001
title: "Implementation plan — Mass migration of Korean comments to English"
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
type: plan
---

# Implementation Plan — SPEC-V3R6-CODE-COMMENTS-EN-001

## 1. Approach Overview

### 1.1 Execution Model

**Agent-based per-file batch translation** (사용자 확정, AskUserQuestion 2026-05-23):

- **Tool**: `manager-develop` subagent (Section A-E delegation MANDATORY for Tier L)
- **Cycle**: per-file Read → Translate (LLM semantic) → Edit (Edit tool, NOT sed)
- **Concurrency option**: Agent Teams (5+1+1 pattern, `.claude/rules/moai/workflow/agent-teams-pattern.md`) for wall-time 단축 시 고려

### 1.2 Why Agent-Based (Not Script)

| 옵션 | 장점 | 단점 | 채택 여부 |
|------|------|------|----------|
| sed/awk regex | 빠름 (10분 wall-time) | semantic loss 위험, 문맥 무시 | ❌ 거부 |
| `gettext`-style i18n | 표준 | Go 주석에 부적합, 도구 부재 | ❌ 거부 |
| **Agent translation** | semantic preservation, 문맥 인식 | 3-5h wall-time | ✅ **채택** |
| ML 자동번역 (DeepL/Google) | 빠름 | 코드 컨텍스트 무지, 식별자 손상 | ❌ 거부 |

### 1.3 Tier L Compliance

- **Tier L 기준**: >15 files, >1000 LOC. 본 SPEC = **267 files / 4,246 lines** → **확실히 L**.
- **5-artifact set**: spec / plan / acceptance / design / research (이 SPEC 디렉토리 내)
- **Section A-E delegation MANDATORY** (LEAN Tier S optional 적용 안 됨)
- **plan-auditor invocation 권장** (Tier L threshold 0.85)

---

## 2. Section A-E Delegation Reference

본 SPEC의 run-phase delegation은 [.claude/rules/moai/development/manager-develop-prompt-template.md](.claude/rules/moai/development/manager-develop-prompt-template.md) **Section A-E 5-section structure REQUIRED**.

각 Wave 위임 시 다음 구조 의무:

- **Section A — Context**: 작업 위치, branch HEAD SHA, SPEC 산출물 경로, plan-auditor verdict
- **Section B — Known Issues 자동 주입** (8 categories): B1 cross-platform build tags / B2 cross-SPEC 충돌 / B3 C-HRA-008 subagent boundary / B4 Frontmatter canonical / B5 CI 3-tier / **B6 spec-lint heading h3** / B7 observer.go capture path / **B8 working tree hygiene**
- **Section C — Pre-flight Checks**: branch + baseline 확인, cross-platform build, lint baseline, PRESERVE list
- **Section D — Constraints**: PRESERVE files, 금지 명령 (sed/awk), Conventional Commits, REQ-CCE-004/005/008
- **Section E — Self-Verification Deliverables**: AC matrix, cross-platform build, coverage, subagent boundary grep, lint NEW vs baseline, branch HEAD + push

---

## 3. Wave-Split Decomposition (7 Waves)

### 3.1 Foundation packages (9 files, smallest blast radius first)

**Scope**: `internal/config` (4) + `internal/core` (1) + `internal/hook` (1) + `internal/spec` (3)

**Justification**: 의존성 그래프 leaf 위치 — 본 Wave 결함은 후속 Wave에서 표면화. baseline 안정성 확립.

**Estimated comment count**: ~200-300 lines (small)

**Independent PR**: `feat/SPEC-V3R6-CODE-COMMENTS-EN-001-wave-1`

### 3.2 Wave 2 — CLI surface (25 files, user-visible)

**Scope**: `internal/cli/**/*.go` (non-test only, 25 files)

**Justification**: User-visible code — comment 품질이 외부 contributor 진입 장벽에 직접 영향. Wave 1 안정성 확립 후 진입.

**Estimated comment count**: ~800-1200 lines

**Special attention**: Cobra command `Use:/Short:/Long:/Example:` 필드 (AC-CCE-010) — **RESOLVED: OQ-CCE-001 Option B 사용자 결정 2026-05-22** — N=14 entries Englishified per EXCL-CCE-001 exception (research.md §6.1 enumeration). AC-CCE-010 PASS = 0 Korean Cobra fields after Wave 2; AC-CCE-004 PRE − 14 = 2172 expected string literal count.

**Independent PR**: `feat/SPEC-V3R6-CODE-COMMENTS-EN-001-wave-2`

### 3.3 Wave 3 — Harness + migration (23 files)

**Scope**: `internal/harness` (16) + `internal/migration` (7)

**Justification**: Harness 시스템은 V3R6 HARNESS-RENAME-001 baseline residual 존재 (EXCL-CCE-008). 본 Wave에서 NEW failure introduce 금지 — stash-test 검증 의무.

**Estimated comment count**: ~600-900 lines (mid)

**Independent PR**: `feat/SPEC-V3R6-CODE-COMMENTS-EN-001-wave-3`

### 3.4 Wave 4 — pkg/ + ALL remaining non-test internal/* packages (39 files)

**Scope**: `pkg/` (2 files) + ALL remaining non-test files under `internal/*` NOT covered by Waves 1-3. Based on research.md §1.2 inventory (96 non-test total), this includes ~37 files in the "(other internal/*)" row (e.g., `internal/{worktree, bodp, mx, statusline, doctor, runtime, project, audit, ciwatch, lsp, ...}`).

**Wave count reconciliation**: Wave 1 (9) + Wave 2 (25) + Wave 3 (23) + **Wave 4 (39)** = **96 non-test files** (matches research.md §1.2 total).

**Justification**: Catch-all cleanup wave — encompasses all `internal/*` packages not enumerated in Waves 1-3 plus `pkg/`. This redefinition (broader than original "2-3 files" estimate) ensures **100% coverage of the 96 non-test inventory**. Risk elevated to **Medium** due to the heterogeneous package mix and larger surface area.

**Sub-wave option (4a/4b split)** if Wave 4 PR diff becomes unwieldy (>~30 files / >~1000 lines): split into Wave 4a (~20 files for high-traffic packages like mx, worktree, lsp) and Wave 4b (~19 files for remaining infra packages). Decision deferred to Wave 4 entry — if grep inventory at that point shows >25 files affected, propose split via BLOCKER report.

**Estimated comment count**: ~800-1,200 lines (revised upward from ~50-100 to reflect actual scope).

**Independent PR**: `feat/SPEC-V3R6-CODE-COMMENTS-EN-001-wave-4`

**Risk**: Medium (large heterogeneous diff, potential for sub-split).

### 3.5 Wave 5 — Test files split A (cli + template tests, ~50 files)

**Scope**: `internal/cli/**/*_test.go` (40) + `internal/template/**/*_test.go` (10)

**Justification**: 가장 큰 test surface. Wave 1-4 non-test 완료 후 test의 영어화로 일관성 확보.

**Estimated comment count**: ~1000-1500 lines

**Independent PR**: `feat/SPEC-V3R6-CODE-COMMENTS-EN-001-wave-5`

### 3.6 Wave 6 — Test files split B (harness + lsp + hook tests, ~38 files)

**Scope**: `internal/harness/**/*_test.go` (25) + `internal/lsp/**/*_test.go` (4) + `internal/hook/**/*_test.go` (9)

**Justification**: Harness test = HARNESS-RENAME-001 cascade 영향 잔존 가능. baseline 보존 검증 의무.

**Estimated comment count**: ~600-900 lines

**Independent PR**: `feat/SPEC-V3R6-CODE-COMMENTS-EN-001-wave-6`

### 3.7 Wave 7 — Test files split C (remaining test, ALL not in Waves 5-6)

**Scope**: All remaining `*_test.go` files NOT covered by Waves 5-6. Initial enumeration per research.md §1.2 includes `internal/{migration=5, config=2, core=1, spec=7}/**/*_test.go` (= 15 files), but **expands to ~83 files when the "(other internal/*)" test row (~68 files in packages like `internal/{worktree, bodp, mx, statusline, doctor, runtime, project, audit, ciwatch, ...}/*_test.go`) is included**.

**Test file count reconciliation**: Wave 5 (50) + Wave 6 (38) + **Wave 7 (~83)** = 171 test files (matches research.md §1.2 total).

**Sub-wave option (7a/7b split)** if Wave 7 PR exceeds ~30 files / ~1000 lines: split into Wave 7a (enumerated migration/config/core/spec = 15 files) and Wave 7b (other internal/* test = ~68 files). Decision deferred to Wave 7 entry inventory.

**Justification**: Final cleanup wave. 본 Wave 완료로 AC-CCE-001/002/003 grep 0 matches 달성 + 171 test files 100% 커버.

**Estimated comment count**: ~1,200-1,800 lines (revised upward from ~200-400 to reflect actual scope).

**Independent PR**: `feat/SPEC-V3R6-CODE-COMMENTS-EN-001-wave-7` (or split into 7a/7b).

**Risk**: Medium (large heterogeneous test surface, similar to Wave 4).

### 3.8 Wave Summary Table

| Wave | Scope | Files | ~Lines | Independent PR | Risk |
|------|-------|------:|--------|----------------|------|
| 1 | Foundation (config=4 / core=1 / hook=1 / spec=3) | 9 | ~250 | wave-1 | Low |
| 2 | CLI surface (non-test only) | 25 | ~1000 | wave-2 | Medium (Cobra exception N=14, OQ-CCE-001 resolved Option B) |
| 3 | Harness + migration (non-test only) | 23 | ~750 | wave-3 | Medium (HARNESS-RENAME-001 baseline residual) |
| 4 | pkg/ + ALL remaining internal/* non-test (sub-split 4a/4b possible) | 39 | ~800-1,200 | wave-4 | Medium (heterogeneous packages) |
| 5 | Test A (cli=40 + template=10) | 50 | ~1,250 | wave-5 | Medium (large diff) |
| 6 | Test B (harness=25 + lsp=4 + hook=9) | 38 | ~750 | wave-6 | Medium (baseline residual) |
| 7 | Test C (migration=5 + config=2 + spec=7 + core=1 + remaining test) | 15 | ~300 | wave-7 | Low |
| **Non-test subtotal** | **Waves 1+2+3+4** | **9 + 25 + 23 + 39 = 96** | — | — | — |
| **Test subtotal** | **Waves 5+6+7** | **50 + 38 + 15 = 103** (rounded; see note) | — | — | — |
| **Total** | **All 7 Waves** | **267 files** (96 non-test + 171 test) | ~5,100 lines | **7 PRs** | — |

**Test subtotal reconciliation note**: Waves 5+6+7 sum to 103 (50+38+15), but research.md §1.2 inventory lists 171 test files. The gap (~68 files) corresponds to the "(other internal/*)" test row — packages like `internal/{worktree, bodp, mx, statusline, doctor, runtime, project, audit, ciwatch, ...}/*_test.go`. These are absorbed into Wave 7's "remaining test" scope (currently estimated 15, but will expand to ~83 at Wave 7 entry). **Wave 7 may sub-split (7a/7b) if PR size exceeds ~30 files**, mirroring the Wave 4 sub-split protocol.

**Authoritative reconciliation**: Non-test 96 (exact per research.md §1.2) split 9+25+23+39. Test 171 (exact per research.md §1.2) split 50+38+~83 (Wave 7 expanded at runtime). Total 267 files matches research.md §1.2 inventory.

---

## 4. Per-Wave Execution Protocol

각 Wave는 다음 표준 7-step 사이클을 따른다:

### 4.1 Pre-Wave Setup

1. **Branch creation**: `git switch -c feat/SPEC-V3R6-CODE-COMMENTS-EN-001-wave-N origin/main` (또는 main 직진 Late-Branch 선택)
2. **Baseline capture**:
   - `git log --oneline -1` → HEAD SHA 기록
   - `golangci-lint run --timeout=2m 2>&1 | tail -5` → lint baseline
   - `go test ./...` → test pass count baseline (EXCL-CCE-008 잔존 확인)
3. **PRESERVE list 작성**: 본 Wave 외 파일은 변경 금지

### 4.2 Per-File Translation Cycle

각 in-scope file 대해:

1. **Read** (`Read` tool with absolute path)
2. **Identify Korean spans**: comment lines (`//` or `/* */`) + @MX tag descriptions + godoc
3. **Translate** (Agent semantic translation, REQ-CCE-001/002/003 적용):
   - Preserve identifiers (SPEC-ID, REQ-ID, error codes) per REQ-CCE-004
   - Preserve string literals byte-identical per REQ-CCE-005
   - Preserve technical terms (goroutine, defer, etc.) per REQ-CCE-006
   - Apply ambiguity heuristic per REQ-CCE-007 (clarity > literal)
4. **Edit** (`Edit` tool, NOT Write/sed/awk per C-CCE-001):
   - One edit per Korean span (or MultiEdit for batched same-file edits)
   - Verify no string literal modification
5. **Verify per-file**:
   - `gofmt -l <file>` → empty (no formatting drift)
   - `go vet <file>` → no new vet errors
6. **Commit per package/wave** (NOT per-file — Conventional Commits):
   ```
   feat(comments): translate Korean to English in <package> (Wave N)
   ```

### 4.3 Post-Wave Verification

1. **AC matrix verification** (per AC-CCE-001..012 in [acceptance.md](./acceptance.md)):
   - `grep -rn '//.*[가-힣]' <wave-scope>` → 0 matches (or remaining count for partial waves)
   - `grep -rn '@MX:[A-Z]*[: ].*[가-힣]' <wave-scope>` → 0 matches
   - `go build ./...` exit 0
   - `GOOS=windows GOARCH=amd64 go build ./...` exit 0
   - `go test ./...` → baseline pass count 보존 (NEW failures = 0)
   - `golangci-lint run --timeout=2m` → NEW issues = 0
2. **Cross-platform build**: Windows + Linux 양쪽 PASS
3. **String literal preservation check**:
   ```bash
   # Before Wave
   grep -rn '".*[가-힣].*"' <wave-scope> --include="*.go" | wc -l > /tmp/string-before
   # After Wave
   grep -rn '".*[가-힣].*"' <wave-scope> --include="*.go" | wc -l > /tmp/string-after
   diff /tmp/string-before /tmp/string-after  # MUST be empty
   ```
4. **Sample manual review** (AC-CCE-009): 5 random files semantic preservation 확인
5. **Push + PR creation**: `gh pr create --title "feat(comments-en): Wave N — <scope>" --body ...`

### 4.4 Inter-Wave Coordination

- Wave 1 PR merged → Wave 2 branch from origin/main (latest)
- 연속 Wave 진행 시 PRESERVE list 갱신 (이전 Wave 머지 파일 반영)
- main 베이스라인 drift 방지 — 각 Wave 진입 전 `git fetch origin main` + branch base 정렬

---

## 5. Estimated Wall-Time

(time predictions 금지 정책 per `agent-common-protocol.md` §Time Estimation — priority-based만 사용)

**Sequential Agent execution**:
- Per file: 평균 LLM Read + Translate + Edit cycle
- Per Wave: 9-50 files
- Total: 7 Waves

**Parallel Agent Teams 옵션** (`.claude/rules/moai/workflow/agent-teams-pattern.md` 5+1+1):
- Implementer 5명 × 동시 Wave 처리 시 wall-time 단축
- Tier L threshold 충족 → Agent Teams 권장 (workflow.yaml `team.enabled: true` 전제)
- Fallback: solo mode (single manager-develop sequential)

**Priority labels** (time predictions 대체):
- Priority Critical: Wave 1 (baseline 확립)
- Priority High: Wave 2, 3 (user-visible + baseline residual)
- Priority Medium: Wave 5, 6 (test surface)
- Priority Low: Wave 4, 7 (cleanup)

---

## 6. Risks and Mitigation Strategy

(spec.md §7과 cross-reference)

### 6.1 R-CCE-001: Semantic Drift

**Mitigation**:
- Agent translation (LLM semantic) 사용
- AC-CCE-009 sample manual review (5 random files per Wave)
- Reviewer가 의미 손실 발견 시 해당 Wave PR revert + 재실행

### 6.2 R-CCE-002: Identifier Corruption

**Mitigation**:
- REQ-CCE-004 verbatim rule (Section D constraints)
- AC-CCE-011 검증: `grep -rn '\(SPEC\|REQ\|AC\)-[A-Z0-9-]\+' --include="*.go" | wc -l` before/after count 동일

### 6.3 R-CCE-003: String Literal Corruption

**Mitigation**:
- REQ-CCE-005 byte-identity
- AC-CCE-004 count 검증 (~2,186 lines 보존)
- 매 Wave post-verification에 string-literal count diff 의무

### 6.4 R-CCE-004: Byte-Corruption (tabs/spaces)

**Mitigation**:
- Edit tool 사용 (sed/awk 금지, C-CCE-001)
- `gofmt -l <file>` per-file verification
- `go vet ./...` per-file verification

### 6.5 R-CCE-005: Wave Conflict

**Mitigation**:
- 7-wave **sequential** 진행 (parallel 권장 안 함)
- 각 Wave PR merged 후 다음 Wave branch base 정렬

### 6.6 R-CCE-006: Large Diff Review Burden

**Mitigation**:
- 7-wave 분할 → PR당 평균 ~625 lines (max ~1250 in Wave 5)
- Per-wave AC verification report PR body 의무

### 6.7 R-CCE-007: Test Failure Cascade

**Mitigation**:
- EXCL-CCE-008 baseline 잔존 문서화
- 매 Wave stash-test verification:
  ```bash
  git stash --include-untracked
  go test ./... 2>&1 | grep -E "FAIL|PASS" | wc -l > /tmp/baseline-tests
  git stash pop
  # After Wave changes
  go test ./... 2>&1 | grep -E "FAIL|PASS" | wc -l > /tmp/post-tests
  diff /tmp/baseline-tests /tmp/post-tests  # delta 검증
  ```

---

## 7. Open Questions

(Default proceed 가능, 발견 시 BLOCKER 처리 — 사용자 결정 필요)

### 7.1 OQ-CCE-001: Cobra Command Strings (Wave 2) — **RESOLVED**

`Use:/Short:/Long:/Example:` 필드는 string literal (EXCL-CCE-001 보존 대상)이지만 사용자 documentation 역할.

**사용자 결정 (2026-05-22)**: **Option B 채택 — 영어화**.

근거: `moai <cmd> --help` 출력은 오픈소스 CLI 사용자가 보는 documentation surface이며 `code_comments: en` 정책과 일관성 유지. EXCL-CCE-001 예외로서 N=14 entries Englishified (research.md §6.1 enumeration).

영향:
- AC-CCE-010 active (no longer SKIPPED) — Wave 2 완료 후 `grep ... '[가-힣]' | wc -l` = 0 의무
- AC-CCE-004 baseline: PRE_COUNT_FROM_BASELINE − 14 = 2172 expected post-count
- REQ-CCE-005 carve-out 추가 (spec.md §3.3) — Cobra 14 entries 예외 명시
- EXCL-CCE-001 본문 (spec.md §5.1) 14 entries enumeration 반영

### 7.2 OQ-CCE-002: error.New("한국어") Messages

Error string은 `errors.New("...")` 또는 `fmt.Errorf("...")` 의 string literal. EXCL-CCE-001로 보존되지만, 영어 사용자 노출 시 부적절.

- **Default**: 보존 (EXCL-CCE-001) — 별도 SPEC `SPEC-V3R6-ERROR-MESSAGES-EN-001` (가칭) 후속 처리 권장
- **운영 권장**: 본 SPEC 진행 중 발견되는 error string 한국어를 grep으로 list 작성 → research.md 부록

### 7.3 OQ-CCE-003: Log Messages

`log.Printf("...")`, `slog.Info("...")` 등 log message string. EXCL-CCE-001 보존, 별도 SPEC 후속 권장.

### 7.4 OQ-CCE-004: Wave Order — main 직진 vs feat-branch

CLAUDE.local.md §23 Hybrid Trunk 정책 — Tier S 미만 main 직진 허용. 본 SPEC은 Tier L → feat-branch + 자동 PR 권장.

- **Default**: Wave별 feat-branch + PR
- **Alternative**: main 직진 + auto_pr 활용 (정책 허용)
- **Recommendation**: feat-branch (review 용이성, large diff)

---

## 8. Verification Plan

(상세 AC는 [acceptance.md](./acceptance.md) 참조)

### 8.1 Per-Wave Verification (7회 반복)

1. AC-CCE-001/002/003 grep 0 matches (wave scope)
2. AC-CCE-005/006 cross-platform build
3. AC-CCE-007 test baseline 보존
4. AC-CCE-008 lint NEW = 0
5. AC-CCE-004 string literal count 보존
6. AC-CCE-011 identifier count 보존
7. AC-CCE-012 diff scope `.go` only

### 8.2 Final Verification (Wave 7 완료 후)

1. **Global grep**: `grep -rn '//.*[가-힣]\|/\*.*[가-힣].*\*/\|@MX:[A-Z]*[: ].*[가-힣]' internal/ cmd/ pkg/ --include="*.go"` → 0 matches
2. **AC-CCE-009 sample manual review**: 전체 SPEC 종료 시 5 random files semantic check
3. **AC-CCE-010 Cobra**: `grep -rn 'Use:\|Short:\|Long:' internal/cli --include="*.go" | grep '[가-힣]'` → 0 matches (OQ-CCE-001 결정에 따라)
4. **Full cross-platform**: `GOOS=darwin/linux/windows GOARCH=amd64 go build ./...` 3종 PASS

---

## 9. Self-Verification Deliverables (E1-E7 per Section E)

각 Wave 완료 보고 시 `manager-develop`가 자체 검증 의무:

- **E1**: AC matrix per Wave (in-scope ACs PASS/FAIL with verification commands)
- **E2**: Cross-platform build result (darwin + windows)
- **E3**: Coverage 측정 (per-package ≥ 85% threshold, baseline 비교)
- **E4**: Subagent boundary grep (C-HRA-008): `grep -rn 'AskUserQuestion\|mcp__askuser' <wave-scope> | grep -v "_test.go" | grep -v "// "` → 0 matches (translation이 boundary 위반 도입 안 함을 검증)
- **E5**: Lint status — NEW issues vs pre-existing baseline 구분
- **E6**: Branch HEAD SHA + push 결과 + PR URL
- **E7**: Blocker report (해당 시) — OQ-CCE-001 (Cobra string) 등 발견 시

---

## 10. Cross-references

- [spec.md](./spec.md) — Requirements + AC summary
- [acceptance.md](./acceptance.md) — Binary AC matrix detail (12 ACs)
- [design.md](./design.md) — Translation methodology + Agent batch strategy
- [research.md](./research.md) — Codebase Korean inventory + sample patterns
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E template
- `.claude/rules/moai/workflow/agent-teams-pattern.md` — 5+1+1 parallel option
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — Read-only verification batching

---

Version: 0.2.0
Status: draft
Tier: L (5-artifact set, Section A-E delegation MANDATORY)
Wave count: 7 (Waves 4 and 7 may sub-split into 4a/4b / 7a/7b at runtime)
Reconciliation: Wave 1 (9) + Wave 2 (25) + Wave 3 (23) + Wave 4 (39) = 96 non-test ; Wave 5 (50) + Wave 6 (38) + Wave 7 (~83) = 171 test ; total = 267 files
