# SPEC-WEB-CONSOLE-018 — Progress

Card t1079. Baseline of record: 67.9% @ `WT-web-coverage 0314801c2` (2026-09-22, `go test ./internal/web/... -cover`).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-WEB-CONSOLE-018
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
baseline_of_record:
  figure: 67.9%
  tree: 0314801c2
  command: go test ./internal/web/... -cover
  measured_at: 2026-09-22
  profile: .moai/state/verify/t1079/cover-baseline.out
historical_figure_provenance: SPEC-WEB-CONSOLE-017 sync-audit F2 @ 9d4a20eae (t1051, cited for provenance only)
target: 85% (quality.yaml test_coverage_target)
arithmetic: 9414 stmts / 3023 zero / need +1611 covered
```

## §E.2 Run-phase Evidence

### M1 — baseline re-pin (card t1079, AC-001)

- Tree: `WT-web-coverage @ 853e0f5df` (plan commit included; parent `0314801c2`). Measured 2026-09-22.
- Command (env-scrubbed, single invocation, `-count=1` to defeat cache):

  ```
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/web/ -coverprofile=/tmp/t1079-baseline.out
  ```

- Verbatim output:

  ```
  ok  	github.com/modu-ai/moai-adk/internal/web	25.523s	coverage: 67.9% of statements
  ```

- Exit code: 0.
- Divergence vs pinned baseline of record (67.9% @ `0314801c2`): **0.0pt — within the 1pt gate; no re-baseline** (REQ-001 / AC-001 satisfied).
- Per-function baseline profile: `/tmp/t1079-baseline.out` (regenerable with the command above). Function-level confirmation of spec §A.2 shape observed via `go tool cover -func`: `Monitor 28.7%`, `gauge 30.4%`, `spark 0.0%`, `specDetail 0.0%`, `buildAttention 45.5%`, `loadGoals 19.0%`, `loadVerify 20.6%`, `panel 51.5%`, `saveCluster 56.2%`, `Kanban 54.7%`.

### M2–M7 — test groups landed (AC-010..AC-015)

Each group: new `_test.go` file, defect answer in the file header, green on landing (`go test -count=1 ./internal/web/ -run '<pattern>'`), `go vet` clean.

| M | TG | File | Tests | Deciding output |
|---|----|------|-------|-----------------|
| M2 | TG-1 | `internal/web/monitor_states_test.go` | 5 (`TestMonitor*`) | `ok github.com/modu-ai/moai-adk/internal/web 0.559s` |
| M3 | TG-2 | `internal/web/screen_states_test.go` | 11 (`TestKanban*`, `TestTodo*`, `TestSpecs*`, `TestCloseDebt*`, `TestSpecRow*`, `TestSpecDetail*`, `TestOverview*`) | `ok ... 0.853s` |
| M4 | TG-3 | `internal/web/fieldsets_states_test.go` | 17 (`TestSchema*`, `TestSettingsPanel*`, `TestFieldset*`, `TestPermissionOption*`, `TestAgentFMRow*`, `TestMCPToggle*`, `TestCodex*`, `TestGlmKey*`) | `ok ... 0.607s` |
| M5 | TG-4 | `internal/web/shell_chrome_states_test.go` | 15 (`TestShell*`, `TestNavRow*`, `TestTopbar*`, `TestLiveState*`, `TestSaveCluster*`, `TestProfilePop*`, `TestPanelMeta*`, `TestNoteBanner*`, `TestGauge*`, `TestStageMark*`, `TestStateMark*`, `TestBackendBadge*`, `TestBadgeKinds*`, `TestLaneUnresolved*`) | `ok ... 0.535s` |
| M6 | TG-5 | `internal/web/page_chrome_states_test.go` | 9 (`TestProfileModifiable*`, `TestProfileActionRoutes*`, `TestIconSVG*`, `TestIconAt*`, `TestBanner*`, `TestSaveAction*`, `TestSettingsPage*`) | `ok ... 0.623s` |
| M7 | TG-6 | `internal/web/viewmodel_ops_edges_test.go` | 16 (`TestSessionState*`, `TestEstimateStage*`, `TestTelemetryCells*`, `TestPipelineColumns*`, `TestHumanSince*`, `TestClampPct*`, `TestBuildAttention*`, `TestChainCardID*`, `TestCloseDebtRows*`, `TestMustFixFindings*`, `TestBuildFilters*`, `TestLoadGoals*`, `TestLoadVerify*`, `TestBuildMonitor*`, `TestBuildOverview*`, `TestLoadVerifyColonKey*`) | `ok ... 0.664s` |

- One observed RED was recorded during M2 (first assertion of the spark cell count mis-counted the `spark__b`/`spark__b--on` substring overlap — an assertion-precision failure, corrected; all groups pin existing behavior, no product change was derived so no feature-RED exists by design). Landing-run REDs in M5/M7 were likewise assertion-vs-contract corrections (rail owns 3 nav rows, not 6; `loadVerify` colon-key divergence — below).
- Helper reuse (REQ-004 / AC-020): `renderTempl` (templ_helpers_test.go), `mustRender`-style direct renders, `newTestApp`/`serveGet`, `writeBoardSpec` (board_test.go), `tg3Errs`-style per-file fixtures only. No new harness, no shared-state files touched.

### M8 — coverage verdict (AC-030 / AC-031)

- Command (same form as M1, single invocation):

  ```
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/web/ -coverprofile=/tmp/t1079-verdict.out
  ```

- Verbatim output:

  ```
  ok  	github.com/modu-ai/moai-adk/internal/web	25.115s	coverage: 74.6% of statements
  ```

- **Verdict: 74.6% — SHORT of the 85% target** (REQ-012 shortfall path; no filler tests added).
- Full package suite at verdict: `go test -count=1 ./internal/web/...` → `ok ... 25.359s` (exit 0). `go vet ./internal/web/...` → exit 0. `golangci-lint run ./internal/web/...` → `0 issues.`

#### Per-file residual table (verdict profile, spec §A.2 shape)

| File | Stmts | Zero | File coverage | Δ zero vs baseline |
|---|---|---|---|---|
| fieldsets_templ.go | 2,821 | 842 | 70.2% | −50 |
| screens_templ.go | 1,940 | 578 | 70.2% | −310 |
| widgets_templ.go | 665 | 176 | 73.5% | −129 |
| shell_templ.go | 771 | 218 | 71.7% | −39 |
| page_templ.go | 791 | 231 | 70.8% | **0** |
| root_templ.go | 329 | 89 | 72.9% | −9 |
| viewmodel_ops.go | 289 | 20 | 93.1% | −69 |
| fieldsets_codex_templ.go | 268 | 76 | 71.6% | −8 |
| icons_templ.go | 170 | 44 | 74.1% | −16 |
| browser.go + assets.go (excluded, spec §E) | 16 | 6 | — | 0 (excluded) |
| all other non-templ files (schemaform, agentfm, events, handlers, screens, profile_crud, …) | 1,354 | 108 | — | −5 |
| **package** | **9,414** | **2,388** | **74.6%** | **−635** |

- §D.13 indirect verification: every group's target file shows non-zero zero-statement reduction **except page_templ.go (TG-5's page-level surface)**. Investigation (per §D.13, performed): page_templ's residual is 227 zero blocks / 231 stmts, of which 224 are 1-statement codegen error-propagation blocks (`if templ err != nil { return err }`) and 3–4 are 2-stmt child-`Render` error propagations — the reachable branches (langSelect/optSelect/toggle/numberField clean+errored) were already covered by the pre-existing `templ_helpers_test.go` parity tests, so TG-5 added zero conversion there by construction, not by omission.

#### REQ-012 shortfall report — why 85% is unreachable within this SPEC's constraints

Classification of the 2,388 residual zero statements (computed from `/tmp/t1079-verdict.out`; block-shape attribution corrected by sync-audit independent measurement — 2,238×1-stmt + 144×2-stmt + 6×3-stmt):

1. **2,238 statements — 1-statement error-propagation blocks** (`if err != nil { return err }`): **2,148 templ codegen** (after every `WriteString` / `ResolveAttributeValue`) + **90 hand-written** (viewmodel_ops 20, agentfm 13, schemaform 11, projectconfig 8, server 7, events 7, handlers 4, …). Rendering into a healthy in-memory buffer cannot fail the templ ones; converting them requires a failing-writer/child-component injection seam, i.e. product-code changes forbidden by spec §C, or error-path tests that exist only to execute the branch — coverage-filling under spec REQ-003. Several of the 90 hand-written blocks are likewise fault-injection-only.
2. **144 statements — 2-statement blocks**: **100 templ child-`Render` error propagations** (e.g. `fieldsets_templ.go:179`, `screens_templ.go:50`) + **44 non-templ**. Same reachability wall as (1): a child render cannot be made to fail without an injected failing component (a seam), or a test whose sole purpose is the error branch (filler).
3. **6 statements — 3-stmt blocks.** Beyond the error-propagation classes, the genuinely reachable residual is larger than the run-phase estimate: sync-audit sample-verified behaviorally-reachable branches in the non-templ surface (e.g. `handlers.go:240` 404 branch, `handlers.go:250` invalid-profile 400, `glmkey.go:120` non-loopback 403, `profile_crud` reserved-name 400, schemaform/projectconfig input branches) — named files alone carry ~74 zero stmts.

**Conclusion (carrying the sync-audit F1 measured correction, 2026-09-22)**: 85% would require converting 976 residual statements; the unreachable error-propagation pool is **≈2,300–2,320 of the 2,388** (not ≥2,358 as first stated — the initial bucket attribution over-counted templ and understated reachable non-templ surface), and the honest ceiling of behavioral render assertions at this codegen shape is in the **75.0–76.1% range** (74.9% if only the templ-codegen residual converts; 76.1% only if every non-templ residual converts, several of which are fault-injection-only). **No reading of the range reaches 85%**, so the verdict-driving conclusion is unchanged: the gap is structural (generated-code error propagation), not a test-authoring shortfall. Decision requested per REQ-012: accept the documented shortfall, OR re-plan with a product-side scope (e.g. templ codegen error-block filtering in coverage policy, or injection seams — both out of this SPEC's tests-only envelope).

#### Discovered defect (recorded, NOT fixed — tests-only scope)

**`loadVerify` (viewmodel_ops.go) never surfaces colon-key verify snapshots.** `loadVerify` re-derives each key from the snapshot *filename*, but `verify.Save` sanitizes the canonical `HEAD:<sha>` key into a `HEAD-<sha>` filename; `verify.Load`'s stored-key match (`Snapshot.Key != key → treated absent`) then rejects every row. The Monitor verification panel reads empty while snapshots exist. Pinned as a deliberate tripwire `TestLoadVerifyColonKeyDivergence` (viewmodel_ops_edges_test.go) — the fix flips that test in the same commit. Fixing it is a product-code change outside spec §C; surfaced here for the orchestrator to dispatch (candidate follow-up card).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: 81a43dd9a   # last TG milestone commit; M8 evidence commit follows (see git log)
run_status: complete-with-shortfall   # 74.6% vs 85% target — REQ-012 shortfall report in §E.2
ac_pass_count: 10   # AC-001, AC-010..AC-015, AC-020, AC-030, AC-040, AC-041
ac_fail_count: 1    # AC-031 (85% not met) — resolved as the REQ-012 documented-shortfall branch; close landed at 7bf28b3a8 and the sync-audit (2026-09-22, PASS-WITH-DEBT 82/100) adjudicated the branch legitimate — this §E.3 "pending" note closed per sync-audit F3
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a   # worktree lane; no push from the lane (lead-batched)
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0   # golangci-lint run ./internal/web/... → 0 issues; go vet → exit 0
cross_platform_build:
  status: n/a   # tests-only SPEC; no product-code build surface (plan.md §E E2)
total_run_phase_files: 6   # new _test.go files only
m1_to_mN_commit_strategy: one commit per milestone (M1..M8), conventional commits, SPEC-ID + card id in subject
coverage:
  baseline: 67.9% @ 853e0f5df
  verdict: 74.6%
  target: 85%
  shortfall_path: REQ-012 report (progress.md §E.2) — no filler tests landed (REQ-003/REQ-012)
discovered_defects:
  - loadVerify colon-key divergence (viewmodel_ops.go) — pinned by TestLoadVerifyColonKeyDivergence; product fix out of tests-only scope; orchestrator dispatch requested
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-22
sync_commit_sha: "7bf28b3a8"   # backfilled per the D3 exemption (placeholder was pending-backfill-sync)
sync_status: complete-with-documented-shortfall   # 74.6% vs 85% target; REQ-012 shortfall report in §E.2 stands
what_sync_changed:
  - CHANGELOG.md [Unreleased] > Added — SPEC-WEB-CONSOLE-018 close entry (tests-only coverage reinforcement + documented shortfall + discovered loadVerify defect recorded as known limitation)
  - spec.md frontmatter status: in-progress → completed (single sync commit, 3-phase close; status + updated only, zero body edits)
  - progress.md §E.4 (this signal)
b12_self_test_a: pass   # grep -c 'SPEC-WEB-CONSOLE-018' CHANGELOG.md → 0 pre-emission (no duplicate)
b12_self_test_b: pass   # 12 distinct AC identifiers in acceptance.md (AC-001, AC-010..015, AC-020, AC-030, AC-031, AC-040, AC-041); entry carries the same 12
b12_self_test_c: pass   # all 6 claimed _test.go paths verified via git diff --name-only 9da176ac2^..8748aa5bc
changelog_entry_position: "[Unreleased] > Added (top; newest-first per house convention)"
readme_docs_site: not-applicable   # internal tests-only, no user-facing surface change
mx_scan: "no tags added — sync surfaces are markdown-only (CHANGELOG, §E.4, spec.md frontmatter); zero exported functions touched; run landed _test.go files only (product files untouched by rule)"
frontmatter_status_transitions:
  from: in-progress
  via: implemented
  to: completed
  carrier: single sync commit (3-phase close — no separate Mx commit)
verification:
  post_commit_scoped: "go test -count=1 ./internal/web/ -cover must still print 'coverage: 74.6% of statements' (docs must not move it); git status --short clean"
  full_suite: not-run-locally-by-rule   # lead-batched push; CI on origin/develop is the full verdict
```
