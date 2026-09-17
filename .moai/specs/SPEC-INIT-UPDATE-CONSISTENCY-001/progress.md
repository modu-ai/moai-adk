# progress.md — SPEC-INIT-UPDATE-CONSISTENCY-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-13
plan_audit: iteration 2/2 PASS score 1.0 (2026-09-14) — iter-1 FAIL 0.975 on single blocking D1 (flag-registration site omitted from deletion-target enumeration), D1 closed at 07a8b8999; boundary-estimation nuance recorded as optional D5 (non-blocking)
plan_audit_head: 07a8b8999
plan_audit_report: .moai/reports/t588/plan-audit-iter2.md
tier: M (artifacts: spec.md + plan.md + acceptance.md)
f16_verdict: dissolved-with-evidence (spec.md §3)
f15_verdict: accept-manual-recovery (REQ-ICU-006)
f17_verdict: record-only-durable (REQ-ICU-007)

## §E.2 Run-phase Evidence

Run tree: `.claude/worktrees/t588`, base `ea56ed9b5` → HEAD (M6). One commit per milestone (M1-M6), all with `Authored-By-Agent: manager-develop`.

### AC PASS/FAIL Matrix (acceptance.md §D)

| AC | Status | Verification Command | Actual Output (this run, this tree) |
|----|--------|---------------------|-------------------------------------|
| AC-001 (F8 ghost removal) | PASS | `go test -run 'TestWritePhase1Configs_DoesNotPersistProjectMode\|TestProjectYAMLTemplateCarriesNoModeKey' ./internal/core/project/` + E4 greps | `ok ... internal/core/project 0.485s`; `grep ProjectMode` non-test Go = 0; flag registration = 0; tmpl `mode:` = 0; reader grep config+web = 0 |
| AC-002 (F9 default parity) | PASS | `go test -run 'TestExecutionMode' -count=1 ./internal/config/` | `ok github.com/modu-ai/moai-adk/internal/config 0.408s` (parity test GREEN; membership test unmodified and passing); `defaults.go:890 ExecutionMode: "auto"` |
| AC-003 (F12 pair dedupe) | PASS | `go test -run 'TestManagedRedeployCount' -count=1 ./internal/cli/` | `ok ... internal/cli 0.943s` (4 pairs → 4 (+1 standalone) after RED 9-vs-5; order-independent; standalone file counts 1) |
| AC-004 (F13 root summary) | PASS | `go test -run 'TestRenderUpdateOutcome' -count=1 ./internal/cli/` | `ok ... internal/cli 0.711s` (all-roots named with paths; zero-root adds no rows; single-root legacy rows byte-identical + advisory row) |
| AC-005 (F14 parity guard) | PASS | `go test -run 'TestSchemaParity' -count=1 ./internal/web/` | `ok ... internal/web 0.630s` (35 orphans census-exempt with rationale; codex toggles + MCP counted rendered via dedicated components; stale-exemption reverse guard) |
| AC-006 (F15 advisory) | PASS | same render batch — `TestRenderUpdateOutcome_ConfigOnlyLegacyShape` | advisory names `evaluator-profiles/` + `astgrep-rules/`, "not merge-restored", backup path visible; absent when no config backup |
| AC-007 (F17 record-only) | PASS | `grep -c '@MX:DEBT\|@MX:CEILING\|@MX:UPGRADE' internal/config/manager.go` | `3` (one each) + 6-section scope godoc present (`Scope: Save writes exactly six section files`) |
| AC-008 (F16 regression guard) | PASS | `go test -run 'TestReconfigureMembershipExcludesPage3\|TestAgentWiring' ./internal/cli/wizard/` | `ok ... internal/cli/wizard 0.425s` (AC-WIZ-012a contract intact; no audit-selection question reintroduced) |

### Boundary Cases (acceptance.md §E)

- Standalone `.sh` (non-pair): covered by `TestManagedRedeployCount_PairCountsOnce` (standalone managed file counts 1).
- Zero-root render: `TestRenderUpdateOutcome_NoRootsNoNewRows` — no backup rows, no advisory.
- Legacy projects with residual `mode:` key: no migration; key dissolves on next update redeploy (zero readers) — documented in `initializer_expansion_test.go` fixture comment and the ghost tests.

### RED Evidence (TDD axes)

- M1: `--- FAIL: TestExecutionModeDefaultMatchesTemplate ... compiled default workflow.execution_mode = "team" but template workflow.yaml ships "auto"` (captured before GREEN).
- M3: `--- FAIL: TestManagedRedeployCount_PairCountsOnce ... managedRedeployed = 9, want 5` and `--- FAIL: TestManagedRedeployCount_TmplOnlyEntry ... = 2, want 1` (captured against the original counting logic after a behavior-preserving extraction of `managedRedeployCount`).
- M2 ghost tests: both FAILED pre-removal (`WritePhase1Configs created project.yaml — project.mode is a ghost key`; `project.yaml.tmpl still carries a mode key ... "mode: personal"`).
- M5 census RED: 24-field orphan list captured with the harness exemptions already in place, then resolved via verified census (codex toggles render on MCP tab's codexAuthBlock — false positives removed; 35 true orphans exempted).

### Baseline (Section C, B5)

`go test ./internal/cli/... ./internal/config/... ./internal/core/project/... ./internal/web/...` at `ea56ed9b5`: all packages `ok` EXCEPT `internal/cli` full-package FAIL at 601.658s with zero `--- FAIL` lines — consistent with the default 10m per-binary timeout under parallel load (environmental; reproduced twice). `internal/web` FAILed in the first parallel batch and passed in isolation (`ok ... 24.161s -count=1`). Full-suite verdict belongs to CI on `origin/develop`.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-14
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (lane protocol — card worktree on WT branch; no push by lane)
l44_post_push_fetch: n/a (lane never pushes; lead batch-pushes develop after the integration window)
new_warnings_or_lints_introduced: 0 (golangci-lint affected packages: 39 issues, all pre-existing in untouched files; 0 in files this SPEC modified)
cross_platform_build.darwin: pass (`go build ./...` exit 0)
cross_platform_build.windows: pass (`GOOS=windows go build ./...` exit 0)
coverage_baseline_to_final: internal/config 82.0→82.0; atomicfile 81.8→81.8; toolpolicy 89.1→89.1; core/project 88.8→88.9; web 67.8→67.8; cli (targeted filter) 7.5→7.6 — no decrease (baseline tree: `git archive ea56ed9b5` to /tmp/t588-base, same commands)
total_run_phase_files: 19 files changed, 642 insertions(+), 214 deletions(-)
m1_to_mN_commit_strategy: one commit per milestone (M1 82290a61f, M2 b53c6c9d6, M3 a3ea5101e, M4 35141e3e9, M5 9c30edc89, M6 pending)
f16_gaps_absorption: no wizard web-console promise text found (grep over internal/web + internal/cli re-ask/promise surfaces: 0 relevant hits) — no absorption needed
deviations: M3 extraction note (counting loop extracted verbatim into `managedRedeployCount` as a behavior-preserving prerequisite so the RED was observable against the original logic); M4 archive-drift attribution implemented as a before/after glob diff in the caller rather than a signature change to `archiveLegacySkills` (ANCHOR, fan_in ≥ 3 — untouched); F15 advisory row keyed on config backup presence, which is the exact pre-clean+restore condition at the sole production call site

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-09-14
sync_commit_sha: pending-backfill-sync
sync_status: complete
run_to_sync_transition: in-progress → implemented → completed (3-phase close, single sync commit)
changelog_entry_emitted: yes (1 entry, `[Unreleased]`/`### Changed`, English-only; pre-emission grep count 0 — no duplicate)
b12_self_test_a: pass (grep -c 'SPEC-INIT-UPDATE-CONSISTENCY-001' CHANGELOG.md = 0 pre-emission)
b12_self_test_b: pass (live AC identifiers in acceptance.md §D = 8: AC-001..AC-008; `AC-WIZ-012a` counted as a prior-test reference, not a live AC — CHANGELOG entry references 8)
b12_self_test_c: pass (every file path named in the CHANGELOG entry verified present in this tree)
frontmatter_status_transitions.spec_md: in-progress → completed (sync commit; `updated: 2026-09-14`)
canary_compliance_check.catalog_hashes: clean (orchestrator ran `go run ./internal/template/scripts/gen-catalog-hashes.go --all` this turn — output byte-identical, zero diff; catalog.yaml untouched)
docs_surface_assessment: docs-site 4 locales still mention the removed `--project-mode` flag / `project.mode` key (13 files: `getting-started/cli.md`, `getting-started/init-wizard.md`, `cli-reference/init.md` in en/ja/ko/zh + `en/workflow-commands/moai-project.md`) — out of this sync commit's scope; recorded in the CHANGELOG residual and handed to a separate docs card. No README surface mentions either token (README.ko.md grep 0).
mx_tag_validation: AC-007 annotations verified present in `internal/config/manager.go` (1× `@MX:DEBT`, 1× `@MX:CEILING`, 1× `@MX:UPGRADE`) per §E.2 — no new annotations required by sync phase

## §F Phase 4 Mode Selection

- Inputs: tier M; scope ~10 files across 5 packages (cli, config, core/project, web, template); domains 2 (Go source + embedded template yaml); language mix Go + shell-template + yaml; concurrency benefit LOW (coding-heavy, per-milestone RED-GREEN sequencing with inter-axis contact); agent-team prereqs: not requested
- Mode evaluation: direct — not a trivial fix; fanout — not research-heavy; sweep — not a mechanical-uniform bulk transform; serial — selected
- Decision: serial
- Justification: coding-heavy multi-axis consistency work with per-milestone TDD ordering and inter-axis contact (M4 touches the render area M3's consumer reads); per Anthropic's coding-task parallelism caveat, one manager-develop runs M1-M6 sequentially.
- Plan-audit gate skip: verdict PASS (iteration 2, score 1.0 ≥ Tier M 0.80), artifact-hash unchanged since audited HEAD 07a8b8999 (post-audit diff is progress.md only — not a hashed plan artifact). Skip recorded per the authoritative skip contract; this skip does not bypass Implementation Kickoff Approval, which was separately obtained (operator approval relayed by the lead, 2026-09-14).
