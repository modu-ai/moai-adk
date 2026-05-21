# Implementation Plan — SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001

## 1. Tier Classification

- **Tier**: S (Simple)
- **LOC delta estimate**: < 250 (DELETE 5 test functions ≈ -200 LOC, REWRITE 4 test functions ≈ +60/-100 LOC, net ≈ -240; `Line1_WithPrefixes` removed from cleanup scope per iter 2 D1 fix)
- **Files affected**: 2 (`internal/statusline/builder_test.go`, `internal/statusline/renderer_test.go`)
- **plan-auditor threshold**: 0.75

### 1.1 Tier S 3-artifact deviation note (iter 2 D4 fold)

Tier S canonical artifact set per `spec-workflow.md` § SPEC Complexity Tier은 **2 files** (`spec.md` + `plan.md` with AC inline in spec.md §3). 본 SPEC은 **3 files** (`spec.md` + `plan.md` + `acceptance.md`)로 deviate함.

**사유**:
1. REQ↔AC binary verification matrix (7 ACs + 3 ECs + 5 Quality Gates)가 `plan.md` scope 외부에서 별도 검증 surface를 가지는 것이 LEAN dogfooding 측면에서 유리.
2. Definition of Done 체크리스트 (8 items)가 `acceptance.md`에 위치할 때 run-phase 위임 시 self-verification matrix로 직접 참조 가능.
3. 9개 test function (5 DELETE + 4 REWRITE + 1 UNTOUCHED) 분류가 `plan.md` Test Treatment Matrix와 `acceptance.md` 검증 명령 사이의 traceability 강화.
4. Tier M 선례 SPEC-V3R5-DOCS-SECURITY-001과 일관된 3-artifact pattern (memory project entry `project_v3r5_docs_security_001_run_pass.md` 참조).

**Threshold**: Tier S 0.75 유지 (artifact 수 deviation은 quality gate에 영향 없음 — plan-auditor PASS threshold는 tier 기반).

### 1.2 Delegation prompt format

- **Section A-E template**: OPTIONAL for Tier S — minimal delegation form used (Goal / Deliverables / Constraints / Self-verification). 본 SPEC §5 Delegation Strategy 참조.

## 2. Test Treatment Matrix

| # | Test | File:Line | Action | Rationale |
|---|------|-----------|--------|-----------|
| 1 | TestRenderFullV3_FiveLines | renderer_test.go:922 | **DELETE** | Asserts `len(lines) != 5` for `ModeFull`. Retired layout per renderer.go:48-50. |
| 2 | TestRenderFullV3_Lines2To4_SeparateBars | renderer_test.go:995 | **DELETE** | Asserts L2=CW only / L3=5H only / L4=7D only separate-line bars. Retired. |
| 3 | TestRenderFullV3_Line5_DirBranchGit | renderer_test.go:1040 | **DELETE** | Asserts L5 (5th line) directory/branch/git layout — only exists in retired 5-line full mode. |
| 4 | TestRenderFullV3_StyleInL1 | renderer_test.go:1084 | **DELETE** | Asserts L1 style merge specifically for full mode (`Output style merged into L1 — L6 removed`). Retired. |
| 5 | TestRenderFullV3_WithResetTimes | renderer_test.go:1321 | **DELETE** | Asserts `len(lines) >= 5` at line 1340 + L3 `(reset time)` + L4 `(reset time)` parentheses contract specific to retired 5-line full mode. Independently verified FAIL (`full mode should have 5 lines, got 3`). Parentheses reset-time formatting contract already covered by `TestRenderUsageBarWithReset` at renderer_test.go:1356 (directly tests `renderUsageBarWithReset` function); default mode L2 uses `(rolling)` annotation, not `(reset time)`, so REWRITE would test unrelated semantic. (iter 2 D2 fix) |
| 6 | TestBuilder_SetMode | builder_test.go:230 | **REWRITE** | Currently asserts `default != full`. Rewrite as: NormalizeMode collapses all StatuslineMode variants to identical 3-line output. Add table-driven sub-tests over ModeDefault/ModeFull/ModeCompact/ModeMinimal/ModeVerbose. |
| 7 | TestIntegration_ModeLineCount | builder_test.go:589 | **REWRITE** | Drop AC-V3-02 (verbose=5) + AC-V3-05 (full=5) cases. Keep AC-V3-01/03/04 (minimal/compact/default = 3 lines) and unify expected line count to 3 across all modes. Rename test cases to reflect collapse contract. |
| 8 | TestIntegration_NoUsageLineCount | builder_test.go:675 | **REWRITE** | Delete AC-V3-06 sub-test (`full + no usage = 5 lines`). Keep AC-V3-06b sub-test (`default + no usage → L2 contains CW+5H+7D`) unchanged. |
| 9 | TestIntegration_GradientBar | builder_test.go:733 | **REWRITE** | Change `Mode: ModeFull` → `Mode: ModeDefault`. Update comment from `full mode CW bar` → `default mode CW bar (L2)`. CW 40-block contract preserved in default layout. |
| — | TestRenderFullV3_Line1_WithPrefixes | renderer_test.go:960 | **UNTOUCHED** | L1 prefix assertions (`v2.1.50`, `🗿 v2.8.0`, session time format) are layout-independent — identical in default 3-line mode and retired full 5-line mode. Independently verified PASS in both default run and isolated run. No DELETE / no REWRITE / no source change. (iter 2 D1 fix — was previously listed for DELETE in iter 1) |

### Summary
- **DELETE**: 5 tests (all in `renderer_test.go`: FiveLines, Lines2To4_SeparateBars, Line5_DirBranchGit, StyleInL1, **WithResetTimes**)
- **REWRITE**: 4 tests (all in `builder_test.go`: TestBuilder_SetMode, TestIntegration_ModeLineCount, TestIntegration_NoUsageLineCount, TestIntegration_GradientBar)
- **UNTOUCHED**: 1 test (`TestRenderFullV3_Line1_WithPrefixes` in renderer_test.go — PASS, layout-independent)
- **TOUCH**: 0 source files (REQ-SFC-007)
- **Total functions referenced**: 10 (9 modified + 1 retained as-is)

## 3. Milestones

### M1 — renderer_test.go cleanup (DELETE 5)

**Files**: `internal/statusline/renderer_test.go`

> ⚠️ **Atomic execution required (iter 2 D6 fold)**: M1의 5개 DELETE는 **단일 MultiEdit 호출**로 일괄 처리. 이유: 순차 Edit 시 각 deletion이 후속 line offset을 drift 시켜 후속 step의 line number가 어긋남 (예: 첫 deletion 후 line 995의 `Lines2To4_SeparateBars`가 line 957 등으로 이동). MultiEdit는 모든 변경을 file-snapshot 기준으로 일괄 적용하므로 drift-free.

**Preserved (not deleted)**: `TestRenderFullV3_Line1_WithPrefixes` (renderer_test.go:960) — layout-independent PASS test. M1 시 명시적으로 보존 (iter 2 D1 fix).

**Actions** (single MultiEdit batch):
1. Delete `TestRenderFullV3_FiveLines` (lines 922-958)
2. Delete `TestRenderFullV3_Lines2To4_SeparateBars` (lines 995-1038) — `Line1_WithPrefixes` (lines 960-993) 사이에 있으나 **보존**
3. Delete `TestRenderFullV3_Line5_DirBranchGit` (lines 1040-1082)
4. Delete `TestRenderFullV3_StyleInL1` (lines 1084 onward until next `func TestX`)
5. Delete `TestRenderFullV3_WithResetTimes` (lines 1321-1354) — iter 2 D2 newly added
6. Verify the "Cycle 3: renderFullV3 tests" section header comment block — if it now only annotates `Line1_WithPrefixes`, update the comment to reflect single retained test; do not remove if `Line1_WithPrefixes` remains in scope.

**Verification**:
```bash
# Only Line1_WithPrefixes should remain
go test -count=1 -run '^TestRenderFullV3' ./internal/statusline/ -v 2>&1 | grep -E '^=== RUN|^--- '
# Expected: only TestRenderFullV3_Line1_WithPrefixes appears, with --- PASS

go test ./internal/statusline/...
# Expected: pre-existing renderer_test.go failures resolved (all 5 DELETED tests no longer FAIL)
```

### M2 — builder_test.go REWRITE (4 tests)

**Files**: `internal/statusline/builder_test.go`

**M2.1 — TestBuilder_SetMode REWRITE** (line 230):

Replace the body's "default vs full should differ" assertion with a table-driven sub-test:

```go
func TestBuilder_SetMode(t *testing.T) {
    clearGLMEnv(t)

    input := &StdinData{
        Model:         &ModelInfo{Name: "claude-sonnet-4-20250514"},
        ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
    }

    // 모든 StatuslineMode 변형은 NormalizeMode를 통해 default 3-line layout으로 collapse 한다.
    // renderer.go:46-65 anchor — full layout retirement으로 인한 backward-compat contract.
    modes := []StatuslineMode{
        ModeDefault, ModeFull, ModeCompact, ModeMinimal, ModeVerbose,
    }

    var baseline string
    for i, mode := range modes {
        t.Run(string(mode), func(t *testing.T) {
            builder := New(Options{
                GitProvider: &mockGitProvider{
                    data: &GitStatusData{Branch: "main", Modified: 2, Available: true},
                },
                Mode:    mode,
                NoColor: true,
            })
            got, err := builder.Build(context.Background(), makeStdinJSON(input))
            if err != nil {
                t.Fatalf("Build error for mode=%s: %v", mode, err)
            }
            if i == 0 {
                baseline = got
                // baseline은 3 lines (default layout)
                if lines := countLines(got); lines != 3 {
                    t.Errorf("baseline mode=%s should produce 3 lines, got %d", mode, lines)
                }
                return
            }
            if got != baseline {
                t.Errorf("mode=%s output should collapse to default baseline\nbaseline:\n%s\ngot:\n%s",
                    mode, baseline, got)
            }
        })
    }
}
```

**M2.2 — TestIntegration_ModeLineCount REWRITE** (line 589):

Replace the 5-case table with a 5-case table where all cases expect `minLines=3, maxLines=3`:

```go
// AC-SFC-001/002: all StatuslineMode variants collapse to 3-line default layout
// via NormalizeMode (renderer.go:46-65 anchor — 5-line full layout retired).
func TestIntegration_ModeLineCount(t *testing.T) {
    tests := []struct {
        name        string
        mode        StatuslineMode
        withUsage   bool
        description string
    }{
        {"minimal→default 3 lines", "minimal", true, "minimal mode collapses to default 3-line"},
        {"verbose→default 3 lines", "verbose", true, "verbose mode collapses to default 3-line"},
        {"compact→default 3 lines", ModeCompact, true, "compact mode collapses to default 3-line"},
        {"default exactly 3 lines", ModeDefault, true, "default mode produces 3-line layout"},
        {"full→default 3 lines", ModeFull, true, "full mode collapses to default 3-line"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var usageProv *mockUsageProvider
            if tt.withUsage {
                usageProv = realisticUsage(45.0, 82.0)
            }
            builder := New(Options{
                GitProvider:    realisticGit(),
                UpdateProvider: &mockUpdateProvider{data: &VersionData{Current: "2.8.0", Available: true}},
                UsageProvider:  usageProv,
                Mode:           tt.mode,
                NoColor:        true,
            })
            got, err := builder.Build(context.Background(), makeStdinJSON(realisticInput()))
            if err != nil {
                t.Fatalf("Build error: %v", err)
            }
            if lines := countLines(got); lines != 3 {
                t.Errorf("%s\nline count: got=%d, want=3\noutput:\n%s",
                    tt.description, lines, got)
            }
        })
    }
}
```

**M2.3 — TestIntegration_NoUsageLineCount REWRITE** (line 675):

Delete the `AC-V3-06: full + no usage → 5 lines (5H/7D 0%)` sub-test (lines 677-703). Keep the `AC-V3-06b: default + no usage → L2 CW+5H+7D` sub-test (lines 706 onward) unchanged. Update the function-level comment to reflect new scope:

```go
// TestIntegration_NoUsageLineCount verifies that even when usage data is nil,
// the default layout's L2 line still shows CW + 5H(0%) + 7D(0%) bars.
// (Retired AC-V3-06 full-mode 5-line variant deleted per SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001.)
func TestIntegration_NoUsageLineCount(t *testing.T) {
    // AC-V3-06b: default + no usage → L2 CW+5H+7D
    t.Run("default + no usage → L2 CW+5H+7D", func(t *testing.T) {
        // ... existing AC-V3-06b body preserved unchanged ...
    })
}
```

**M2.4 — TestIntegration_GradientBar REWRITE** (line 733):

In the `AC-V3-07: 60% → 24 of 40 CW blocks filled` sub-test (lines 736-779), change:
- `Mode: ModeFull` → `Mode: ModeDefault`
- Update inline comment: `// full mode: CW(40) + 5H(40, 0%) + 7D(40, 0%) = 120 blocks` → `// default mode L2: CW(40) bar; 60% → 24 filled, 16 empty`
- The CW line extraction loop + `cwFilled == 24` assertion is layout-independent and remains correct.

**Verification (M2)**:
```bash
go test -v -run 'TestBuilder_SetMode|TestIntegration_ModeLineCount|TestIntegration_NoUsageLineCount|TestIntegration_GradientBar' \
  ./internal/statusline/...
# Expected: ok, all sub-tests PASS
```

### M3 — Full validation

```bash
# AC-SFC-001
go test ./internal/statusline/...                          # exit 0

# EC-SFC-002
go test -race ./internal/statusline/...                    # exit 0

# EC-SFC-003
GOOS=windows GOARCH=amd64 go build ./internal/statusline/... # exit 0

# AC-SFC-007
git diff --name-only main -- internal/statusline/ | grep -v '_test\.go$' | wc -l
# Expected: 0

# C-SFC-002 PRESERVE
git status --porcelain | grep -cE '(merge/confirm|harness/usage-log|settings\.json|99-release|workflows/release|init_layout|wizard/fullscreen|wizard/review|hook/\.moai|docs-site/\.moai)'
# Expected: matches pre-cleanup baseline (no change)
```

## 4. Implementation Order

Sequential (single-file edits, no parallelism needed for Tier S):

1. **M1** — Delete 5 functions from `renderer_test.go` in a **single MultiEdit call** (iter 2 D6 atomic-MultiEdit requirement — see M1 § for rationale). Preserves `Line1_WithPrefixes` (UNTOUCHED).
2. **M2.1** — Rewrite `TestBuilder_SetMode` in `builder_test.go`.
3. **M2.2** — Rewrite `TestIntegration_ModeLineCount` in `builder_test.go`.
4. **M2.3** — Rewrite `TestIntegration_NoUsageLineCount` (delete AC-V3-06 sub-test, preserve AC-V3-06b).
5. **M2.4** — Rewrite `TestIntegration_GradientBar` (`ModeFull` → `ModeDefault`).
6. **M3** — Full validation batch (parallelizable via single-turn multi-Bash per `verification-batch-pattern.md`).

## 5. Delegation Strategy

Tier S allows **minimal delegation prompt** form per `manager-develop-prompt-template.md` § Applicability. Minimal prompt structure for run-phase:

```
Goal: Apply SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001 test cleanup per plan.md M1-M3.

Deliverables:
- internal/statusline/renderer_test.go (5 DELETEs per M1)
- internal/statusline/builder_test.go (4 REWRITEs per M2.1-M2.4)
- 0 source file changes outside *_test.go

Constraints:
- PRESERVE: 5 modified + 7 untracked working-tree files (do not stage/modify)
- Late-Branch: commit to local main only (no push, no branch)
- Conventional Commits + 🗿 MoAI trailer
- C-HRA-008: 0 AskUserQuestion references in test files

Self-verification (AC PASS/FAIL matrix):
1) go test ./internal/statusline/... → exit 0 (AC-SFC-001)
2) git diff --name-only ... | grep -v _test.go | wc -l → 0 (AC-SFC-007)
3) grep -c 'len(lines) != 5' internal/statusline/*_test.go → 0 (AC-SFC-003)
4) Cross-mode collapse table in TestBuilder_SetMode (AC-SFC-002/006)
```

Skill: `Skill("moai-workflow-tdd")` — though this is test cleanup not new-feature TDD, the REWRITE pattern (existing test → new assertion → verify) matches RED-GREEN-REFACTOR's REFACTOR phase semantics.

## 6. Risks & Mitigation (cross-reference spec.md §6)

All 4 risks are Low severity. Primary mitigation: REQ-SFC-002 cross-mode assertion + REQ-SFC-005 ModeDefault GradientBar shift + REQ-SFC-007 source-untouched guarantee.

## 7. Out of Scope (Plan-Level)

### 7.1 Out of Scope — deferred follow-ups

- `internal/cli/update.go` `allStatuslineSegments` array fix for preset=full/compact/minimal (PR/Task segment activation) — separate SPEC candidate `SPEC-V3R5-STATUSLINE-PRESET-FIX-001`.
- 5-line full layout re-introduction — not in scope; retirement is intended supersession.
- Performance benchmark of NormalizeMode collapse — current performance is acceptable; benchmark only if regression observed.

## 8. Estimated Effort

- **Wall-time**: ~10-15 minutes (1-pass delegation expected for Tier S)
- **Re-delegation expectation**: 0 (Tier S + clean test treatment matrix + minimal prompt)
- **plan-auditor iterations**: 1 expected (Tier S threshold 0.75, expected score ≥ 0.85 based on clear AC binary verifiability + 100% REQ↔AC traceability)

## 9. Late-Branch Workflow (Plan → Run → Sync)

Per SPEC-V3R5-LATE-BRANCH-001 REQ-LB-005 + `git-strategy.yaml` (`auto_enabled: false`):

1. **Plan-phase (now)**: Write 3 artifacts on local `main`. Plan commit → local `main` (no push).
2. **Run-phase**: M1-M3 edits → commits on local `main` (no push).
3. **Sync-phase / PR**: `git switch -c feat/SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001` from main, cherry-pick relevant commits, push, `gh pr create`.
4. **Post-merge**: Late-branch closure — `git checkout main && git fetch && git reset --hard origin/main && git pull`.

GitHub Issue creation: SKIP (per `feedback_no_github_issue_for_specs.md` policy — default off, opt-in via `/moai plan --issue`).
