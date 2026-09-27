# Progress — SPEC-STATUSLINE-LANDED-LABEL-001 (card t1281)

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier S; acceptance.md included by lane request).
- Base: b59a5d69c on WT-statusline-landed-label.
- Open items: see the plan-phase blocker report returned to the lead (docs-site scope, guard M1/M2 exclusion).

## §E.2 Run-phase Evidence

Run by manager-develop (cycle_type=tdd), card t1281, on `WT-statusline-landed-label` from base `53d549b81`. Raw outputs persisted under `.moai/state/verify/t1281/` (local, untracked).

### RED (captured before any implementation change)

Stage 1 — AC-SLL-001 alone, using only pre-existing symbols, against the old `countNamed` / `--format=%B` code. Behavioral failure, not a compile error:

```
$ go test ./internal/statusline/ -run '^TestRefreshLandedCounts_NonAttributingMentionDoesNotCount$' -count=1
--- FAIL: TestRefreshLandedCounts_NonAttributingMentionDoesNotCount (0.09s)
    landed_test.go:573: non-attributing mention: {Landed:1 Ref:origin/main Measured:true FetchedAt:1790502302 Available:true}, want measured landed=0 (old criterion would give 1)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/statusline	0.489s
FAIL
```

Stage 2 — the remaining AC tests, which reference the new criterion field, criterion constant, and glyph constant:

```
$ go test ./internal/statusline/ -run '^Test(Refresh|Resolve|Maybe|Renderer_Landed|LandedGlyph|LandedScan|LandedRender)' -count=1
internal/statusline/landed_test.go:127:3: unknown field Criterion in struct literal of type LandedCounts
internal/statusline/landed_test.go:127:14: undefined: landedCriterion
internal/statusline/landed_test.go:155:28: undefined: landedGlyph
...
FAIL	github.com/modu-ai/moai-adk/internal/statusline [build failed]
```

### GREEN and gates

```
$ go test ./internal/statusline/... ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/statusline	20.428s
ok  	github.com/modu-ai/moai-adk/internal/kanban	185.537s

$ golangci-lint version
golangci-lint has version v2.1.6 built with go1.26.8 ...
$ golangci-lint run ./internal/statusline/...
0 issues.

$ go vet ./internal/statusline/...
(no output, exit 0)

$ go test ./internal/statusline/ -count=1 -coverprofile=... ; go tool cover -func=...
landed.go  Known 100.0% | resolveLandedCounts 100.0% | maybeRefreshLandedCounts 66.7% (exec spawn path, blocked under go test — pre-existing)
landed.go  landedScanRunner 100.0% | RefreshLandedCounts 91.3% | pickedCards 87.5% | writeLandedCache 70.0%
total: (statements) 90.7%
```

### AC matrix

| AC | Evidence | Status |
|---|---|---|
| AC-SLL-001 | `TestRefreshLandedCounts_NonAttributingMentionDoesNotCount` — RED above (old criterion: `Landed:1`), GREEN in the package run; the test asserts the `\bt101\b` control matches the same subject | PASS |
| AC-SLL-002 | `TestRefreshLandedCounts_AttributedSubjectCounts` (ct = added_at+60 and ct = added_at, the `>=` boundary) | PASS |
| AC-SLL-003 | `TestRefreshLandedCounts_GenerationBoundary/commit_predates_the_card` | PASS |
| AC-SLL-003b | `TestRefreshLandedCounts_GenerationBoundary/empty_added_at_fails_closed` | PASS |
| AC-SLL-004 | `TestRefreshLandedCounts_OneInvocationRegardlessOfCardCount` (n=1 and n=50: calls == 1; argv == `kanban.LandedScanArgs(kanban.LandedRefFor(root))`) | PASS |
| AC-SLL-004b | `TestRefreshLandedCounts_WritesTheCriterion` (raw JSON `criterion` == `subject-attribution/v1`) | PASS |
| AC-SLL-005 | `TestResolveLandedCounts_OldCriterionIsUnknown` (old schema and foreign criterion both unknown; no `⚑`/`✓`; unannotated pair) | PASS |
| AC-SLL-005b | `TestMaybeRefreshLandedCounts_OldCriterionIsStale` (fresh old-schema cache: 1 spawn attempt) | PASS |
| AC-SLL-006 | `TestRefreshLandedCounts_OldCriterionNeverSurvivesAFailure` | PASS |
| AC-SLL-007 | `TestRefreshLandedCounts_FailedQueryKeepsThePriorMeasurement/{runner_error,malformed_stream_(no_separator)}` (landed 3 kept, fetched_at advanced) | PASS |
| AC-SLL-007b | `TestRefreshLandedCounts_NeverMeasuredStaysUnknownOnFailure` (retained) | PASS |
| AC-SLL-007c | `TestRefreshLandedCounts_EmptyQueueNeedsNoQuery` (0 git calls; observed zero under current criterion) | PASS |
| AC-SLL-008 / 008b | `TestRenderer_LandedAnnotation` (`🔄 TODO: 76/4 ⚑48`, `⚑0`; no `✓` in any case) | PASS |
| AC-SLL-009 | `TestRenderer_LandedNeverSubtracts` | PASS |
| AC-SLL-010 | `TestLandedGlyph_SingleCellLocaleNeutral` (one rune U+2691; width 1 with EastAsianWidth false and true) | PASS |
| AC-SLL-011 | `grep -c '✓'` → ko 0 / en 0 / ja 0 / zh 0; `grep -c '⚑N'` → ko 3 / en 2 / ja 3 / zh 3 | PASS |
| AC-SLL-012 | `grep -c -- '--format=%B' internal/statusline/landed.go` → 0 | PASS |
| REQ-SLL-004 render clause | `TestLandedRenderPath_SpawnsNoGit` (retained, now with a current-criterion cache) | PASS |

Additional: `TestLandedScanRunner_RefusesNonGit` pins the adapter guard from plan §H (non-git command refused; git passes through; runner called once).

### Scope notes

- `kanban` package not modified. `countNamed` and `TestCountNamed_WordBoundaryCriterion` removed (criterion retired; no other caller — grep over `internal/statusline` found none).
- `internal/template/templates/**`: no `✓N` occurrence (grep, 0 files) — no template change.
- The cli caller (`internal/cli/statusline.go` → `RefreshLandedCounts`) is unchanged; signature preserved.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-27
run_commit_sha: 7803fa4c0
run_status: complete
ac_pass_count: 18
ac_fail_count: 0
new_warnings_or_lints_introduced: 0
golangci_lint_version: v2.1.6
cross_platform_build: "not measured locally — CI matrix is the verdict"
total_run_phase_files: 9
m1_to_mN_commit_strategy: "single run-phase commit (code + tests + 4-locale docs + SPEC status)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-27
sync_commit_sha: 93f0bb814
sync_status: complete
b12_self_test_a: "grep -c SPEC-STATUSLINE-LANDED-LABEL-001 CHANGELOG.md -> 0 before emission"
b12_self_test_b: "distinct AC IDs in acceptance.md = 12 (AC-SLL-001..012); entry states 12 IDs / 18 matrix rows"
b12_self_test_c: "ls internal/statusline/landed.go internal/statusline/landed_test.go .moai/reports/t1281/verdict.md -> all exist"
changelog_entry_position: "[Unreleased] / ### Changed, first entry"
frontmatter_status_transitions:
  spec.md: "in-progress -> completed (implemented merged into this sync commit); updated already 2026-09-27"
  plan.md: "no status field"
  acceptance.md: "no status field"
verdict: .moai/reports/t1281/verdict.md
retests_rerun_in_sync: false   # §E.2 outputs transcribed, not re-executed
```
