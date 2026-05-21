---
id: SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001
title: "statusline 테스트 정리 — retired full-mode 5-line layout 가정 제거"
version: "0.1.1"
status: draft
created: 2026-05-21
updated: 2026-05-21
author: manager-spec
priority: P1
phase: "v2.20.0-rc1"
module: "internal/statusline"
lifecycle: spec-anchored
tags: "statusline, tests, cleanup, retired-layout, backward-compat, tier-s"
tier: S
---

# SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001 — statusline 테스트 정리 (retired full-mode)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-21 | manager-spec | 초안 작성. `e71f1aa54` (default layout baseline 정렬) 머지 후 잔존한 6 pre-existing failing tests (`internal/statusline/{builder,renderer}_test.go`)을 정리. 5-line "full" layout retirement (renderer.go:48-50 anchor)이 supersession 했으나 해당 contract을 assert 하던 테스트들이 잔존하여 `go test ./internal/statusline/...` 실패. 본 SPEC은 (a) 진정 retired 가정에 종속된 테스트 DELETE, (b) backward-compat contract (NormalizeMode 모든 mode → default 3-line collapse)을 assert 하도록 REWRITE. 소스 코드(`renderer.go`/`builder.go`/`types.go`) 변경 없음 — 테스트만 정리. Tier S 분류: 영향 파일 2개 (`builder_test.go` + `renderer_test.go`), LOC delta 추정 < 250. |
| 0.1.1 | 2026-05-21 | manager-spec | iter 2 revision (plan-auditor verdict 0.712 → target ≥ 0.85). **D1 fix**: `TestRenderFullV3_Line1_WithPrefixes` removed from cleanup scope — independently verified PASS (L1 prefix assertions layout-independent, identical in default 3-line and retired full 5-line). **D2 fix**: `TestRenderFullV3_WithResetTimes` (renderer_test.go:1321) added to DELETE list — independently verified FAIL (`full mode should have 5 lines, got 3` at line 1342); parentheses reset-time contract already covered by `TestRenderUsageBarWithReset` at renderer_test.go:1356. **D3 fix**: spec.md §1.2 failure-count claim corrected — default `go test` run shows 6 visible failures (short-circuit hides 2 additional); isolated runs expose 8 retired-layout failures + 1 passing layout-independent test = 9 retired-mode-related tests total. **D4 fold**: Tier S 3-artifact deviation rationale added to plan.md §1. **D6 fold**: M1 atomic-MultiEdit requirement added. **D8 fold**: AC-SFC-007 verification command tightened (`HEAD~1` → `main`). Test treatment matrix: 5 DELETE + 4 REWRITE + 1 UNTOUCHED = 9 test functions referenced, 8 failing + 1 passing(retained). |

## 1. Background

### 1.1 출처

본 SPEC은 `e71f1aa54` (feat(statusline): default layout baseline 정렬 — PR/Task default-on + 15 segments 명시, 2026-05-21) commit의 "Pre-existing 6 failures (TestRenderFullV3* + TestIntegration_*) full mode retired path 잔존 — 본 변경 무관 (stash 검증 baseline 동일). 별도 SPEC 후보 (`SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001` 가칭)" 기록에서 위임된 후속 SPEC이다. project memory entry `project_statusline_default_layout_baseline.md` 참조.

### 1.2 다루는 결함

기본 `go test ./internal/statusline/...` 실행 시 **6개 실패가 가시화**되지만, isolated 실행 시 (`go test -run '^TestX$' ./internal/statusline/`) Go의 test short-circuit 동작에 의해 **추가 2개 실패가 드러난다**. 전부 동일 root cause (5-line full layout retirement — §1.3 참조). 총 영향 범위: **8개 failing + 1개 passing(layout-independent) = 9개 retired-mode-관련 테스트**.

#### 1.2.1 Default run에서 가시화되는 6개 실패

| Test | 파일:line | 실패 형태 |
|------|-----------|-----------|
| TestBuilder_SetMode | `builder_test.go:230` | "default and full output should differ" — 둘 다 3-line default 출력 |
| TestIntegration_ModeLineCount | `builder_test.go:589` | AC-V3-02/05 (verbose/full = 5 lines) 케이스 실패 |
| TestIntegration_NoUsageLineCount | `builder_test.go:675` | AC-V3-06 (full + no usage = 5 lines) 케이스 실패 |
| TestIntegration_GradientBar | `builder_test.go:733` | ModeFull로 CW bar 40-block 검증 시 layout 가정 mismatch |
| TestRenderFullV3_FiveLines | `renderer_test.go:922` | `len(lines) != 5` (실제 3) |
| TestRenderFullV3_Lines2To4_SeparateBars | `renderer_test.go:995` | L2=CW only / L3=5H only / L4=7D only 가정이 retired layout |

#### 1.2.2 Isolated run에서만 드러나는 추가 2개 실패

short-circuit으로 default run에서 가려지는 항목:

| Test | 파일:line | 실패 형태 | 검증 |
|------|-----------|-----------|------|
| TestRenderFullV3_Line5_DirBranchGit | `renderer_test.go:1040` | L5 (5번째 line) directory/branch/git layout — retired 5-line full mode에만 존재 | `go test -count=1 -run '^TestRenderFullV3_Line5_DirBranchGit$' ./internal/statusline/` |
| TestRenderFullV3_WithResetTimes | `renderer_test.go:1321` | `len(lines) < 5` (line 1340) — 5-line full layout + L3/L4 `(reset time)` 가정 | `go test -count=1 -run '^TestRenderFullV3_WithResetTimes$' ./internal/statusline/` |

#### 1.2.3 Layout-independent passing test (DELETE 대상 아님)

| Test | 파일:line | 상태 | 처리 방침 |
|------|-----------|------|----------|
| TestRenderFullV3_Line1_WithPrefixes | `renderer_test.go:960` | **PASS** (default run + isolated run 모두) | L1 prefix assertion이 default 3-line과 retired full 5-line 양쪽 모두에서 동일하게 적용되는 layout-independent 검증. **DELETE 대상 아님 — UNTOUCHED 보존.** 검증: `go test -count=1 -run '^TestRenderFullV3_Line1_WithPrefixes$' ./internal/statusline/` → `ok ... internal/statusline 0.469s` |

#### 1.2.4 추가 동일 카테고리 (이미 §1.2.1에 포함됨)

| Test | 파일:line | 비고 |
|------|-----------|------|
| TestRenderFullV3_StyleInL1 | `renderer_test.go:1084` | full mode L1 style merge 가정 — §1.2.1 default run 실패 목록 외 short-circuit 가려진 항목 (isolated run 시 실패 가시화) |

#### 1.2.5 검증 방법

전체 retired-layout 테스트 상태 확인:

```bash
for t in TestBuilder_SetMode TestIntegration_ModeLineCount \
         TestIntegration_NoUsageLineCount TestIntegration_GradientBar \
         TestRenderFullV3_FiveLines TestRenderFullV3_Lines2To4_SeparateBars \
         TestRenderFullV3_Line1_WithPrefixes TestRenderFullV3_Line5_DirBranchGit \
         TestRenderFullV3_StyleInL1 TestRenderFullV3_WithResetTimes; do
  result=$(go test -count=1 -run "^${t}$" ./internal/statusline/ 2>&1 | tail -1)
  echo "${t}: ${result}"
done
# Expected post-cleanup:
#   TestBuilder_SetMode: ok (REWRITTEN)
#   TestIntegration_ModeLineCount: ok (REWRITTEN)
#   TestIntegration_NoUsageLineCount: ok (REWRITTEN)
#   TestIntegration_GradientBar: ok (REWRITTEN)
#   TestRenderFullV3_FiveLines: no tests to run (DELETED)
#   TestRenderFullV3_Lines2To4_SeparateBars: no tests to run (DELETED)
#   TestRenderFullV3_Line1_WithPrefixes: ok (UNTOUCHED, was passing)
#   TestRenderFullV3_Line5_DirBranchGit: no tests to run (DELETED)
#   TestRenderFullV3_StyleInL1: no tests to run (DELETED)
#   TestRenderFullV3_WithResetTimes: no tests to run (DELETED)
```

### 1.3 Root Cause

`internal/statusline/renderer.go:46-65`에서 `Render()`가 **retired 5-line full layout**을 제거했다:

```go
// The mode argument is accepted for backward compatibility but always
// collapses to ModeDefault via NormalizeMode — the 5-line "Full" layout was
// retired.
func (r *Renderer) Render(data *StatusData, mode StatuslineMode) string {
    if data == nil { return "MoAI" }
    _ = NormalizeMode(mode) // legacy-name collapse, kept for symmetry
    result := r.renderDefaultV3(data)
    ...
}
```

→ 모든 `StatuslineMode` (ModeDefault, ModeFull, ModeCompact, ModeMinimal, ModeVerbose)는 NormalizeMode를 거쳐 **동일한 3-line default output**을 생성한다. 5-line layout assertion + L2/L3/L4 separate-bar assertion + full-mode-specific L1 prefix assertion 모두 retired contract을 검증한다.

### 1.4 본 SPEC의 가치 (Why)

- **CI green 회복** — `go test ./internal/statusline/...` 가 PASS로 돌아가 `statusline` 패키지 regression guard 복구.
- **Backward-compat contract 명문화** — 모든 mode → 3-line collapse 동작을 test로 잠가 향후 NormalizeMode 회귀 방지.
- **Dead-test 제거** — 검증 불가능한 retired contract test를 정리하여 false-failure noise 제거.
- **소스 코드 무변경** — `renderer.go` / `builder.go` / `types.go` 의 anchor (REQ-V3-MODE-001/002, `NormalizeMode`, `StatuslineMode` 타입, `Render(mode)` 시그니처) 보존.

## 2. Goals

- **G1** — `internal/statusline/renderer_test.go`에서 retired 5-line layout 검증 테스트 5개 DELETE: `TestRenderFullV3_FiveLines`, `TestRenderFullV3_Lines2To4_SeparateBars`, `TestRenderFullV3_Line5_DirBranchGit`, `TestRenderFullV3_StyleInL1`, `TestRenderFullV3_WithResetTimes`. (`TestRenderFullV3_Line1_WithPrefixes`는 layout-independent PASS 상태 — DELETE 대상에서 제외, UNTOUCHED 보존. §1.2.3 참조.)
- **G2** — `internal/statusline/builder_test.go`의 `TestBuilder_SetMode`를 REWRITE — backward-compat contract 검증 (모든 `StatuslineMode` 변형이 `Render()` 호출 시 동일한 3-line default output 생성).
- **G3** — `internal/statusline/builder_test.go`의 `TestIntegration_ModeLineCount` REWRITE — retired AC-V3-02 (verbose=5 lines), AC-V3-05 (full=5 lines) 케이스 제거; AC-V3-01/03/04 (minimal/compact/default = 3 lines) 케이스 유지 + NormalizeMode collapse 의도 명시.
- **G4** — `internal/statusline/builder_test.go`의 `TestIntegration_NoUsageLineCount` REWRITE — retired AC-V3-06 (full + no usage = 5 lines) sub-test 제거; AC-V3-06b (default + no usage → L2 CW+5H+7D) sub-test 유지.
- **G5** — `internal/statusline/builder_test.go`의 `TestIntegration_GradientBar` REWRITE — `ModeFull` → `ModeDefault`로 변경 (CW bar 40-block 계약은 default L2에서도 유효).
- **G6** — `go test ./internal/statusline/...` exit 0 (모든 테스트 PASS). 영향받지 않은 테스트는 무변경 보존. `TestRenderFullV3_Line1_WithPrefixes`는 PASS 상태 유지 (UNTOUCHED).

## 3. Out of Scope (Exclusions)

### 3.1 Out of Scope — 소스 코드 변경 금지

- 본 SPEC은 `internal/statusline/{renderer,builder,types}.go` 등 **소스 코드를 수정하지 않는다**. `Render(mode StatuslineMode)` 시그니처, `NormalizeMode` 함수, `StatuslineMode` 타입 (ModeDefault/ModeFull/ModeCompact/ModeMinimal/ModeVerbose 상수), REQ-V3-MODE-001/002 anchor — 전부 보존.
- 5-line full layout을 **복구하지 않는다**. retirement는 의도된 supersession이며 본 SPEC의 책임은 그 contract 변화를 테스트에 반영하는 것뿐.

### 3.2 Out of Scope — 무관 영역

- `internal/cli/update.go` `allStatuslineSegments` 배열 (PR/Task segment preset=full/compact/minimal에서 누락) — 별도 SPEC 후보 (`SPEC-V3R5-STATUSLINE-PRESET-FIX-001` 가칭) per `project_statusline_task_pr_segment.md` 기록.
- REQ-SLV-012 PR opt-in default off 정책 supersession (이미 commit `e71f1aa54`에서 처리 완료) — 본 SPEC scope 외.
- `docs-site/content/{ko,en,ja,zh}/advanced/statusline.md` 문서 갱신 — full mode retirement는 이미 commit `cc98e3739` + `e71f1aa54` 시점에서 처리 가정; 추가 docs 갱신 필요 시 별도 SPEC.
- 5 modified PRESERVE (`internal/merge/confirm*.go`, `.claude/settings.json`, `.moai/harness/usage-log.jsonl`) + 7 untracked PRESERVE (`.claude/commands/99-release.md`, `.claude/skills/moai/workflows/release.md`, `docs-site/.moai/`, `internal/cli/init_layout.go`, `internal/cli/wizard/fullscreen.go`, `internal/cli/wizard/review.go`, `internal/hook/.moai/`) — 변경 금지.
- Cross-platform build 변경, lint baseline 변경 — 본 SPEC은 테스트 정리만 수행.

## 4. Requirements (EARS)

### REQ-SFC-001 (Ubiquitous)
The `internal/statusline` test suite **shall** PASS (`go test ./internal/statusline/...` exit code 0) after the cleanup is applied, with all retired 5-line full-layout assertions removed.

### REQ-SFC-002 (State-Driven)
**WHILE** the `Renderer.Render()` contract collapses all `StatuslineMode` variants to `ModeDefault` via `NormalizeMode` (renderer.go:46-65), the test suite **shall** assert that `ModeDefault`, `ModeFull`, `ModeCompact`, `ModeMinimal`, and `ModeVerbose` all produce identical 3-line output for an identical `*StatusData` input.

### REQ-SFC-003 (Unwanted)
The test suite **shall not** contain any assertion that requires the output of `Render(data, ModeFull)` to differ from `Render(data, ModeDefault)`, nor any assertion that `len(strings.Split(output, "\n"))` equals 5 for the current `Renderer.Render()` API.

### REQ-SFC-004 (Unwanted)
The test suite **shall not** contain any assertion that the rendered output splits CW, 5H, and 7D bars onto separate lines (the retired full-layout L2/L3/L4 contract).

### REQ-SFC-005 (Event-Driven)
**WHEN** a test exercises the CW bar 40-block gradient contract (former AC-V3-07), it **shall** use `ModeDefault` (not `ModeFull`) and verify the bar in the default L2 line, where the 40-block format remains canonical per the current layout.

### REQ-SFC-006 (Optional)
**WHERE** a backward-compat regression guard for `NormalizeMode` is useful, the rewritten `TestBuilder_SetMode` **may** include a table-driven sub-test enumerating all known `StatuslineMode` constants and asserting collapse to identical 3-line output.

### REQ-SFC-007 (Ubiquitous)
The cleanup **shall** modify only files under `internal/statusline/*_test.go`. Source files (`renderer.go`, `builder.go`, `types.go`, etc.) and files outside `internal/statusline/` **shall** remain unchanged.

## 5. Constraints

- **C-SFC-001** — Test-only scope. No source code modifications under `internal/statusline/`.
- **C-SFC-002** — PRESERVE all 5 modified files + 7 untracked files listed in working-tree hygiene note (B8). Do not stage or modify them.
- **C-SFC-003** — Late-Branch workflow: plan-phase commits land directly on local `main`. No `plan/SPEC-…` branch creation at this step. No `git push` until sync-phase cherry-pick.
- **C-SFC-004** — Conventional Commits format. Plan commit message: `plan(SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001): test cleanup spec for retired full mode` with `🗿 MoAI <email@mo.ai.kr>` trailer.
- **C-SFC-005** — C-HRA-008 subagent boundary discipline: 0 `AskUserQuestion` / `mcp__askuser` references in delegated agent prompts (test files).
- **C-SFC-006** — `tier: S` recorded in frontmatter (per LEAN workflow SPEC-V3R5-WORKFLOW-LEAN-001). plan-auditor threshold 0.75.
- **C-SFC-007** — REQ↔AC traceability 100%. Every REQ-SFC-NNN maps to at least one AC in `acceptance.md`.

## 6. Risks

- **R-SFC-001 (Low)** — Test rewrite may accidentally relax coverage of `NormalizeMode` collapse. **Mitigation**: REQ-SFC-002 mandates explicit cross-mode assertion in the rewritten `TestBuilder_SetMode`.
- **R-SFC-002 (Low)** — TestIntegration_GradientBar switch from `ModeFull` to `ModeDefault` may shift CW bar position in output (different line index). **Mitigation**: REWRITE retains the substring search (`for _, l := range lines { if strings.Contains(l, "CW:") ... }`) — line-position-agnostic.
- **R-SFC-003 (Low)** — Future re-introduction of a multi-line layout would require re-adding tests. **Mitigation**: REQ-SFC-007 keeps source unchanged; a future SPEC introducing a new layout owns its own tests. Out-of-scope per §3.1.
- **R-SFC-004 (Low)** — `TestIntegration_NoUsageLineCount` AC-V3-06 deletion may reduce regression coverage for "no usage" edge case in (now retired) full mode. **Mitigation**: AC-V3-06b retention covers the default-mode equivalent; default is the only mode rendered post-retirement.

## 7. Verification

See `acceptance.md` for the binary AC matrix (REQ↔AC traceability) and `plan.md` for the test treatment matrix + milestone breakdown.

Quick verification commands:

```bash
# Primary AC (REQ-SFC-001)
go test ./internal/statusline/...

# Source-files-untouched check (REQ-SFC-007)
git diff --name-only main -- internal/statusline/ | grep -v '_test.go$' | wc -l
# Expected: 0

# PRESERVE check (C-SFC-002)
git status --porcelain | grep -E '(merge/confirm|harness/usage-log|settings\.json)' | wc -l
# Expected: matches baseline (no change)
```
