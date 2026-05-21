# Acceptance Criteria — SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001

## REQ ↔ AC Traceability Matrix

| REQ | AC | Verification Method |
|-----|----|---------------------|
| REQ-SFC-001 | AC-SFC-001 | `go test ./internal/statusline/...` exit 0 |
| REQ-SFC-002 | AC-SFC-002 | Inspect rewritten `TestBuilder_SetMode` for cross-mode collapse assertion |
| REQ-SFC-003 | AC-SFC-003 | grep absence of `len(lines) != 5` + `gotDefault != gotFull` style assertions |
| REQ-SFC-004 | AC-SFC-004 | grep absence of `must not contain '5H:'`, `must not contain '7D:'`, separate-line bar assertions |
| REQ-SFC-005 | AC-SFC-005 | Inspect rewritten `TestIntegration_GradientBar` for `ModeDefault` usage |
| REQ-SFC-006 | AC-SFC-006 | Inspect `TestBuilder_SetMode` for table-driven sub-tests covering 5 known StatuslineMode constants |
| REQ-SFC-007 | AC-SFC-007 | `git diff --name-only` shows only `*_test.go` changes under `internal/statusline/` |

## Binary Acceptance Criteria

### AC-SFC-001 — Test suite green

**Given** the cleanup has been applied to `internal/statusline/*_test.go`
**When** `go test ./internal/statusline/...` is executed from the project root
**Then** the command **shall** exit with status code 0 and **shall not** print any `--- FAIL:` lines

Verification command:
```bash
go test ./internal/statusline/... 2>&1 | tee /tmp/sfc-001.log
echo "exit=$?"
grep -c '^--- FAIL:' /tmp/sfc-001.log
# Expected: exit=0, FAIL count=0
```

### AC-SFC-002 — Backward-compat NormalizeMode contract asserted

**Given** the rewritten `TestBuilder_SetMode` in `internal/statusline/builder_test.go`
**When** the test body is inspected
**Then** it **shall** assert that `Builder.Build()` produces identical output for at least two distinct `StatuslineMode` values (e.g., `ModeDefault` and `ModeFull`) when given the same `*StdinData` input

Verification command:
```bash
# Affirmative: the rewritten test must contain a same-output assertion
grep -A 50 'func TestBuilder_SetMode' internal/statusline/builder_test.go | \
  grep -E '(gotDefault == gotFull|gotA == gotB|require\.Equal|assert\.Equal)' | wc -l
# Expected: ≥ 1
```

### AC-SFC-003 — Retired 5-line assertions removed

**Given** the rewritten `internal/statusline/*_test.go` files
**When** the codebase is grep'd
**Then** there **shall be** zero occurrences of `len(lines) != 5` AND zero occurrences of `gotDefault == gotFull` followed by `t.Errorf` complaining about identical output (the *negation* assertion that retired full ≠ default)

Verification command:
```bash
# Both must be 0
grep -n 'len(lines) != 5' internal/statusline/*_test.go | wc -l
grep -B 1 -A 3 'gotDefault == gotFull' internal/statusline/*_test.go | \
  grep -c 'should differ'
# Expected: 0 and 0
```

### AC-SFC-004 — Retired separate-bar layout assertions removed

**Given** the rewritten `internal/statusline/renderer_test.go`
**When** grep'd for retired full-layout assertions
**Then** zero occurrences of `must not contain '5H:'` (the L2-only CW assertion), `must not contain '7D:'` (the L3-only 5H assertion), and `must not contain '5H:' \|\| '7D:'` patterns from the retired layout

Verification command:
```bash
grep -E "must not contain.*(5H|7D|CW)" internal/statusline/renderer_test.go | wc -l
# Expected: 0
```

### AC-SFC-005 — GradientBar test uses ModeDefault

**Given** the rewritten `TestIntegration_GradientBar` in `builder_test.go`
**When** the test body is inspected
**Then** the `Mode` field in the `Options{}` literal **shall** be `ModeDefault` (not `ModeFull`)

Verification command:
```bash
# Extract the TestIntegration_GradientBar body and check Mode setting
awk '/^func TestIntegration_GradientBar/,/^func [A-Z]/' \
  internal/statusline/builder_test.go | \
  grep -E '^\s*Mode:\s*ModeDefault' | wc -l
# Expected: ≥ 1 (at least one Mode: ModeDefault assignment in the test body)
```

### AC-SFC-006 — Cross-mode collapse table

**Given** the rewritten `TestBuilder_SetMode`
**When** the test body is inspected
**Then** it **shall** enumerate at least 4 of the 5 known `StatuslineMode` constants (`ModeDefault`, `ModeFull`, `ModeCompact`, `ModeMinimal`, `ModeVerbose`) and verify all produce the same 3-line output

Verification command:
```bash
# Count distinct Mode constant references inside the test body
awk '/^func TestBuilder_SetMode/,/^func [A-Z]/' \
  internal/statusline/builder_test.go | \
  grep -oE 'Mode(Default|Full|Compact|Minimal|Verbose)' | sort -u | wc -l
# Expected: ≥ 4
```

### AC-SFC-007 — Source files untouched

**Given** the cleanup commit(s)
**When** the diff against `main` is inspected
**Then** all modified files under `internal/statusline/` **shall** end in `_test.go` (no source files modified)

Verification command (iter 2 D8 fix — `HEAD~1` → `main` for multi-commit safety):
```bash
# From the SPEC-related commits (all commits on the feature branch / Late-Branch range)
git diff --name-only main -- 'internal/statusline/*' | \
  grep -v '_test\.go$' | wc -l
# Expected: 0
#
# Rationale: HEAD~1 assumes single-commit revision; main..HEAD captures all
# SPEC-related commits regardless of count. For Late-Branch workflow where
# plan + run commits may both land on main before the feature branch is cut,
# the run-phase post-revision diff against main remains accurate.
```

## Edge Cases

### EC-SFC-001 — Unrelated tests must continue to pass

The 9 retired-mode-related test functions listed in spec.md §1.2 are the only tests touched (5 DELETE + 4 REWRITE). `TestRenderFullV3_Line1_WithPrefixes` (renderer_test.go:960) is **UNTOUCHED** — it was PASS pre-cleanup and remains PASS post-cleanup (layout-independent L1 prefix assertions). All other tests in `internal/statusline/*_test.go` (e.g., `TestBuilder_Build_NoNewline`, `TestIntegration_SessionTime`, `TestIntegration_NoCost`, `TestIntegration_GitAheadBehind`, `TestRenderUsageBarWithReset`, default-mode renderer tests starting at `renderer_test.go:900`) **shall** continue to pass with their current assertions unchanged.

Verification:
```bash
# Count of test functions before and after (excluding deleted)
grep -c '^func Test' internal/statusline/{builder,renderer}_test.go
# Pre-cleanup baseline: record this number before applying changes
# Post-cleanup expectation: (baseline - 5) — exactly 5 functions DELETED from renderer_test.go.
# REWRITEs do not change the function count. Line1_WithPrefixes is preserved (not in DELETE count).

# Affirmative check: Line1_WithPrefixes must still exist post-cleanup
grep -c '^func TestRenderFullV3_Line1_WithPrefixes' internal/statusline/renderer_test.go
# Expected: 1 (UNTOUCHED — preserved per iter 2 D1 fix)

# Affirmative check: all 5 DELETE targets must be removed post-cleanup
for t in TestRenderFullV3_FiveLines TestRenderFullV3_Lines2To4_SeparateBars \
         TestRenderFullV3_Line5_DirBranchGit TestRenderFullV3_StyleInL1 \
         TestRenderFullV3_WithResetTimes; do
  count=$(grep -c "^func ${t}" internal/statusline/renderer_test.go)
  echo "${t}: ${count}"
done
# Expected post-cleanup: all 5 print "${t}: 0"
```

### EC-SFC-002 — Race detector clean

```bash
go test -race ./internal/statusline/...
# Expected: exit 0, no DATA RACE warnings
```

### EC-SFC-003 — Cross-platform parity

```bash
GOOS=windows GOARCH=amd64 go build ./internal/statusline/...
# Expected: exit 0 (test files don't need to run on Windows but must compile)
```

## Definition of Done

- [ ] AC-SFC-001 through AC-SFC-007 all PASS
- [ ] EC-SFC-001 verified (only intended tests affected)
- [ ] EC-SFC-002 verified (race-clean)
- [ ] EC-SFC-003 verified (cross-platform compile)
- [ ] Conventional Commit message + `🗿 MoAI` trailer applied
- [ ] Frontmatter `status: draft → implemented` and `version: 0.1.0 → 0.2.0` on completion
- [ ] No source files modified outside `internal/statusline/*_test.go`
- [ ] 5 modified + 7 untracked PRESERVE files unchanged

## Quality Gates

| Gate | Threshold | Verification |
|------|-----------|--------------|
| Test pass rate | 100% in `internal/statusline/` | `go test ./internal/statusline/...` |
| Lint NEW issues | 0 | `golangci-lint run --timeout=2m internal/statusline/...` (baseline diff) |
| C-HRA-008 | 0 matches | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/statusline/` |
| Coverage delta | ≥ 0 (no regression) | `go test -cover ./internal/statusline/...` vs pre-cleanup baseline |
| Source-untouched | 0 non-test file changes | `git diff --name-only ... \| grep -v _test.go \| wc -l` |
