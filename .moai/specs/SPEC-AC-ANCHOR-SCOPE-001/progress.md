# SPEC-AC-ANCHOR-SCOPE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-14
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md
evidence: .moai/reports/t747/anchor-scope-measurement.md (+ probe/ frozen artifacts)
spec_id_regex_check: PASS (verbatim, Bash-run at plan phase)
id_uniqueness: no existing SPEC-AC-ANCHOR-* directory at authoring time
plan_audit: iter1 PASS 0.81 (threshold 0.80) — findings D1-D7 applied this commit; report at .moai/reports/t747/plan-audit.md

## §E.2 Run-phase Evidence

_<run-phase, 2026-09-14, branch WT-ac-anchor-scope, absorb base 188ece2f9 — owned by manager-develop>_

### Milestones

| M | Deliverable | Commit |
|---|---|---|
| M1 | Probe promoted in-tree as two-column corpus test (`internal/spec/zz_t747_anchor_probe_test.go`); frozen pre-repair baseline anchor copy; in-run before-column; control set derived in-run | d5ac26636 |
| M2 | RED fixtures (ce4446694) → narrow-axis declaration-named region selection (01bbb6360) | ce4446694, 01bbb6360 |
| M3 | Declaration-aware terminal fallback (plan §C(ii)(a)) | b82f6ecf9 |
| M4 | Probe instrument dual-shape (live column on parse region); control both shapes; t565 contract update | 2960af5d5 |
| M5 | Doc comments on `findACSectionStart`/`acRegionDeclarationRe`/`sectionHoldsACDeclaration` (landed with M2/M3 commits); vet/lint/gofmt close | 01bbb6360, b82f6ecf9, 2960af5d5 |

### E8 — RED failure output (TDD invariant i, verbatim, pre-GREEN)

Command: `go test ./internal/spec/ -run 'TestFindACSectionStart_' -v -count=1` — verbatim output at `.moai/state/verify/t747/red-m2-unit.txt`. Failing subtests (7):

```
--- FAIL: TestFindACSectionStart_NarrowAxis_DeclarationNamedRegion/numeric_sub_IDs_under_REQ_headings
    parser_anchor_scope_test.go:46: findACSectionStart = -1, want 7
--- FAIL: TestFindACSectionStart_NarrowAxis_DeclarationNamedRegion/level-3_declarations_inside_level-2_section
    parser_anchor_scope_test.go:46: findACSectionStart = -1, want 3
--- FAIL: TestFindACSectionStart_NarrowAxis_ParseableDeclarationsParse
--- FAIL: TestFindACSectionStart_LooseAxis_DeclarationRegionOverEmptyVocabulary/empty_acceptance_section,_declarations_under_later_heading
    parser_anchor_scope_test.go:46: findACSectionStart = 3, want 5
--- FAIL: TestFindACSectionStart_LooseAxis_DeclarationRegionOverEmptyVocabulary/empty_acceptance_section,_declarations_under_earlier_heading
    parser_anchor_scope_test.go:46: findACSectionStart = 7, want 3
--- FAIL: TestFindACSectionStart_NegativeMarker_NeverNamesRegion/out-of-scope_section_holding_declarations_never_anchors
    parser_anchor_scope_test.go:46: findACSectionStart = -1, want 7
```

Preservation/prose-bound guards were green by design pre-GREEN (they pin the no-regression contract).

### E1 — AC PASS/FAIL Matrix (all evidence in-run, this tree, HEAD 2960af5d5)

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-747-001 | PASS | `T747_PROBE_OUT=…/.moai/state/verify/t747/probe-after2 go test ./internal/spec/ -run TestT747AnchorScope -v -count=1` | `NARROW axis (decls, no anchor): before=14 after=11 repaired=3` — 3 repaired (SPEC-AC-COLLECTOR-ANCHOR-001, SPEC-CC297-001, SPEC-STATUS-AUTO-001); 11 justified PROSE-SHAPED-NO-COLON-FORM dispositions (`defect-disposition.txt`); unjustified residual 0 |
| AC-747-002 | PASS | same run | `EMPTY axis (anchored, 0 in-section decls): before=9 after=3 repaired=6` — 6 repaired (DESIGN-ATTACH-001, HANDOFF-MSGMODE-001, HOOK-EVENT-REGISTRY-001, V3R5-WORKFLOW-LEAN-001, V3R6-HOOK-OBSERVE-OPT-IN-001, V3R6-I18N-VALIDATOR-BUDGET-001); 3 justified residuals: GLM-EFFORT-MAX-001 ANCHOR-PARSEABLE-NONBULLET (metric artifact — the anchor reads a genuinely parseable non-bullet section), V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001 + _archive/SPEC-DESIGN-CONST-AMEND-001 ANCHORED-REGION-NO-DECLRE (declarations are colon-less bullets; recovery = line-grammar axis, out of scope per §D); unjustified residual 0 |
| AC-747-003 | PASS | same run | `CONTROL set (decl-bearing, non-defect) = 129, strict-shape deltas = 0, parse-shape deltas = 0` — control set derived in-run from the frozen-column defect union; byte-identical on both comparison shapes; justified-delta list empty |
| AC-747-004 | PASS | same run | `NEWLY-INCLUDED declarations = 48` — all 48 attributed narrow-defect (47: AC-COLLECTOR 3, CC297 19, STATUS-AUTO 25) or empty-defect (1: I18N-VALIDATOR-BUDGET); 0 from the 129 control files; prose layer (in=10/out=1 specimens) untouched |
| AC-747-005 | PASS | same run (single run, before+after columns) | denominator `spec.md read = 861` (frozen filelist 860 + this SPEC's own spec.md — corpus self-modification, t528 discipline); narrow/empty reported as before/after columns from one run |
| AC-747-006 | PASS | `git diff --name-only 188ece2f9..HEAD -- .moai/reports/t747/probe/` → 0 lines; `T528_PROBE_OUT=… go test ./internal/spec/ -run TestT528Anchor -v -count=1` | frozen before-images unmodified; `accepted by FROZEN baseline anchor = 216` (t528 PRESERVE); declRe byte-untouched; baseline column reproduces `IN-SECTION: before(frozen-strict)=1240` and both FROZEN-CROSSCHECK lines `MATCH (14 files)` / `MATCH (9 files)` |
| AC-747-007 | PASS | `git diff --stat 188ece2f9..HEAD -- internal/spec/lint_coverage_sibling.go` → empty | sibling file untouched by any commit; full-package suite green unmodified (existing sibling tests pass) |

### Verification commands + verbatim outputs

- Full affected package: `go test ./internal/spec/ -count=1 -timeout 600s` → `ok github.com/modu-ai/moai-adk/internal/spec 99.574s`
- `go vet ./internal/spec/` → exit 0
- `golangci-lint run ./internal/spec/...` → `0 issues.` (t706 lesson honored — lint measured on this branch)
- `gofmt -l internal/spec/` → empty (clean)
- New anchor logic per-function coverage (coverprofile over anchor test set): `findACSectionStart 100.0%`, `sectionHoldsACDeclaration 100.0%`, `isNegativeSectionHeading 100.0%`, `isACSectionHeading 85.7%`
- `GOOS=windows GOARCH=amd64 go build ./...` → OK
- Probe artifacts (before/after) under `.moai/state/verify/t747/` (probe-before, probe-after2, red-m2-unit.txt, t528-after.txt)

### Per-file dispositions (headline record)

- Narrow 14 = 3 REPAIRED + 11 PROSE-SHAPED-NO-COLON-FORM (declRe hits are mapping tables, discussion notes, colon-less lists — e.g. SPEC-CODEX-EVENT-COVERAGE-001's `AC-CEV-001 → REQ-…` table, SPEC-V3R6-CODERABBIT-ADOPTION-001's colon-less AC list; the corpus declaration notion (discriminator B) over-matches prose here, exactly the mixed-document layer the guard protects).
- Empty 9 = 6 REPAIRED + 3 justified (see AC-747-002 row).
- Note: 6 of the 9 empty-anchor entries and the B4 divergence: the plan-phase probe's strict scan (break at any `##` prefix) cannot see `###`-nested declarations that the parse's anchor-level region includes. The promoted probe measures the live column on the parse's region (AC-747-001/002's region definition) and keeps the strict shape for the frozen baseline column, which still reproduces 1240/14/9 exactly.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-14
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 4  # lint_coverage_sibling.go; zz_t528_anchor_probe_test.go; .moai/reports/t528/**; .moai/reports/t747/probe/* — all diff-empty vs 188ece2f9
l44_pre_commit_fetch: not-run (worktree card branch off develop; HEAD re-read before every commit instead of a shared-branch fetch)
l44_post_push_fetch: not-applicable (lane does not push; develop push is the lead's batch act)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: PASS
cross_platform_build.windows: PASS (GOOS=windows GOARCH=amd64 go build ./...)
total_run_phase_files: 7
m1_to_mN_commit_strategy: one commit per milestone (M1; M2-RED; M2-GREEN; M3; M4+M5; evidence)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
