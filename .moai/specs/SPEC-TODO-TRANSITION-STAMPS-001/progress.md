# Progress — SPEC-TODO-TRANSITION-STAMPS-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-TODO-TRANSITION-STAMPS-001
status: draft
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
tier: M
revision: "0.1.1 — plan-audit FAIL 0.80 remediated (D1 path A: verdict+ref+time, no stored SHA; D2 stamps never coexist; D3-D7 addressed)"
schema_ground_truth: live db backlog.db, lane-measured 2026-09-29; DDL re-read in this tree at 9cc3fdc4d
axis_f_conclusion: columns CONFIRM the title-attribution discriminator; storage records answers + ref, never re-derives attribution; SHA re-derived at re-adjudication
known_gaps:
  - .moai/reports/t472/ not present in this tree; axis record cited from SPEC-TODO-LANDING-ATTRIBUTION-001 instead
open_questions: 1 (plan.md §B — t1308/t1311 merge-order notes are bounded; verdict record shape is now pinned by D1 path A, no run-phase decision remains on it)
served_model_gate_exception: operator approval (relayed by lead 2026-09-29) — served_model_gate disabled in this tree only + GLM-served plan-audit adopted; extension of the operator's existing approval to t1310 via the lead's pre-announced extension clause
auditor_serving_correction: the plan-closure report's "auditor-model: opus" named the SPAWN INJECTION value, not the observed serving model; actual serving was glm-5.3-flash (per the gate's served-model declaration). The plan-audit verdict itself is valid and proceeds under the GLM adoption. Run spawns stay model-injected per profile; SERVED_MODEL_VIOLATION re-occurrence must be reported, never routed around.
```

## §E.2 Run-phase Evidence

Run-phase commits on branch `WT-todo-transition-stamps` (base 9d1993ee8):
c86595297 (AC-TST-012 golden baseline, PRE-change shape, separate commit per
verification-claim-integrity §2.3 ordering rule) → 5e834cc7f (M1 schema +
structs; spec.md `draft → in-progress` transition) → a41e43ead (M2 stamping)
→ 6af5880a7 (M3 verdict persistence) → 5fa7dfa45 (M4 history/list exposure)
→ 8e952ef56 (M5 docs, Template-First: template gtd.md + local mirror +
catalog.yaml) → b2516611b (AC-TST-011 selector entry point) → 83a32eb94
(full-suite regression repairs: doc neutrality, audit tab-count, list-json
golden re-capture).

### AC PASS/FAIL matrix (every row = actually-observed command output;
full per-AC transcript in `.moai/reports/t1310/ac-evidence.txt`, this run,
this tree, HEAD at measurement time)

| AC | Status | Verification command (mechanical check) | Actual output (observed) |
|----|--------|------------------------------------------|--------------------------|
| AC-TST-001 | PASS | `go test ./internal/kanban/ -run TestTransitionStampColumns_FreshUpgradedConverge -count=1` | `ok github.com/modu-ai/moai-adk/internal/kanban 0.157s` (fresh vs upgraded column sets identical; old columns prefix in original order; new cols TEXT nullable) |
| AC-TST-002 | PASS | `go test ./internal/cli/ -run TestTransitionStamps_PickedAtStampsAndClears -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 2.782s` (add --pick stamps; unpick clears; re-pick holds second episode's time) |
| AC-TST-003 | PASS | `go test ./internal/cli/ -run TestTransitionStamps_DroppedAtStampsAndClears -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 1.403s` |
| AC-TST-004 | PASS | `go test ./internal/cli/ -run TestTransitionStamps_ArchivePreservesStampsAndStampsArchivedAt -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 1.400s` (+ drop→done edge: dropped_at only, picked_at NULL) |
| AC-TST-005 | PASS | `go test ./internal/cli/ -run TestDoneVerdict_PersistedWithRefAndTime -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 1.525s` (record holds verdict/ref/at; exactly 3 wire keys; no SHA-named key) |
| AC-TST-006 | PASS | `go test ./internal/cli/ -run TestDoneVerdict_NotPersistedWithoutFlag -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 1.447s` (landing_verdict NULL; stdout line reads landing=unknown) |
| AC-TST-007 | PASS | `go test ./internal/cli/ -run TestDoneVerdict_RefusedWithoutRefThroughStoreAPI -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 1.386s` (store refuses empty-ref verdict; record unchanged) |
| AC-TST-008 | PASS | `go test ./internal/cli/ -run 'TestHistoryArchivedRow|TestHistoryListingExposesTimeAxis' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 7.134s` (lookup + listing render picked/dropped/archived + verdict cell, `-` absent, prefix fields byte-identical, text last) |
| AC-TST-009 | PASS | `go test ./internal/cli/ -run 'TestHistoryLiveRow|TestListJSONStampsOmitEmpty' -count=1` | same combined run `ok ... 7.134s` (live picked_at rendered; list --json picked card carries non-empty picked_at, never-picked card omits key) |
| AC-TST-010 | PASS | `go test ./internal/cli/ -run TestDoneVerdict_ReAdjudicationReDerivesSHA -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 1.572s` (stored ref == printed ref=; re-run predicate vs recorded ref reproduces verdict; delivering SHA re-derived = seeded SHA) |
| AC-TST-011 | PASS | `go test ./internal/kanban/ -run SchemaFreeze -count=1 -v` | `--- PASS: TestSchemaFreezeRecordsTransitionStamps (0.01s)` (selector sweeps exactly 1 test; new columns in expectation set; freeze-test diff = column additions only) |
| AC-TST-012 | PASS | `go test ./internal/cli/ -run TestTodoListJSON_GoldenByteIdentity -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 1.054s` (golden captured at PRE-change HEAD c86595297-parent 9d1993ee8 and committed BEFORE implementation commit 5e834cc7f — ordering witnessed by commit graph; post-change render byte-identical) |

### RED evidence (§E8, verbatim pre-GREEN outputs captured in this run)

- M1: freeze + convergence tests observed FAIL against pre-change schema —
  `backlog_schema_freeze_test.go:118 items column tuples = …(no picked_at)
  want …picked_at:TEXT:0:NULL dropped_at:TEXT:0:NULL`; convergence test:
  `items has no picked_at column` plus 4 more absence lines.
- M2: 5 stamping tests observed FAIL on unstamped tree — `picked_at is NULL
  after add --pick, want a timestamp` and siblings.
- M3: persistence tests observed FAIL with the done-verb persistence hunk
  temporarily reverted — `landing_verdict is NULL after a --require-landed
  done, want a record` / `the query verdict is missing beside the operator
  evidence: <nil>`. Note: the store-layer refusal test (AC-TST-007) was
  already green at that point because the store-layer funnel
  (LandingVerdictValue) had landed with the store-layer commit; RED covers
  the verb-persistence behavior.
- M4: 4 surface tests observed FAIL against unextended renderer (line
  carried 5 fields, want extended shape). AC-TST-009's JSON half passed
  pre-GREEN (the omitempty contract was declared at M1) — it pins, not
  proves.

### Package verification (observed outputs, this run, this tree)

- `go test -timeout 30m ./internal/kanban/... -count=1` → `ok … 188.559s`
  (full kanban suite green, run under `moai slot acquire --resource
  kanban-suite`)
- `go test -timeout 30m -cover ./internal/kanban/... -count=1` → `ok …
  181.416s coverage: 85.5% of statements` (≥ 85%)
- `go vet ./internal/kanban/... ./internal/cli/...` → exit 0, no output
- `golangci-lint run --timeout=2m ./internal/kanban/... ./internal/cli/...`
  → `0 issues.`
- `go test -timeout 30m -cover ./internal/cli/... -count=1` →
  `FAIL … coverage: 84.1% of statements` on the root `internal/cli`
  (uncontended run3, all SPEC tests green; the sole failure is the
  pre-existing doctor test, see below). Subpackages all `ok` (wizard 94.0%,
  worktree 87.7%, printer 97.0%, ...). **84.1% vs the 85% AC clause:
  PASS-WITH-DEBT** — the figure is 0.9pp short and the PRE-change baseline
  of this package was never measured (a base-tree worktree is out of lane
  reach), so whether the shortfall predates this SPEC is unmeasured, not
  established either way; the SPEC's own code (stamps, verdict, history
  fields) is exercised by the 19 new test functions.
- internal/cli targeted regression slices (`TestTodoDrop|TestTodoUndrop|
  TestTodoUnpick|TestTodoAdd|TestTodoNext|TestTodoDone|TestTodoList|
  TestTodoExport|TestTodoJSON|TestTodoArchive|TestTodoUndone` → `ok …
  161.314s`; `TestTodoHistory|TestTodoLanding|TestTodoLanded|TestTodoPR|
  TestTodoSurface|TestTodoExport|TestTodoDisclose|TestTodoReadSurface|
  TestTodoJSON|TestTodoWhy` → `ok … 110.652s` / `ok … 57.435s`)
- Full internal/cli suite (three runs, all under `moai slot acquire
  --resource cli-suite`):
  - run1/run2 (before the repair commit): 4 failures — 3 caused by this
    SPEC's deliberate surface changes and FIXED in 83a32eb94
    (template-neutrality guard correctly refused SPEC IDs in the gtd.md
    mirror; audit multiline tab-count 4→6; live-readers list-json golden
    re-captured per the t384 precedent, only that golden changed) — and 1
    PRE-EXISTING: `TestLocalInstructions_UpdateDoctorPreserveFile` fails on
    a constitution-registry DRIFT against CLAUDE.md /
    agent-common-protocol.md, neither of which this branch touches (empty
    `git log 9d1993ee8..HEAD` on both files; the 5 DRIFT errors are
    CONST-V3R2-013/015/016/017/037) — out of run-phase scope, left for the
    owning card.
  - run3 (after 83a32eb94, uncontended — a concurrent foreign lane's
    suite run during run1/run2 was observed via lsof cwd and waited out):
    the only failure is that same pre-existing doctor test.

### Merge-order notes (plan.md §B bounded obligations)

- t1308 (state-CHECK rebuild): different columns; this SPEC's changes are
  pure additive; whichever card lands second absorbs and re-runs the freeze
  test — backstop catches region drift. No coordination device used.
- t1311 (spec_id backfill at pick): shares the pick code path
  (`appendAnalyzedCard` / `next <n>` mutation block). t1311 landing second
  absorbs and resolves the pick-path overlap in its own mutation block; the
  columns and behaviors are disjoint.


## §E.3 Run-phase Audit-Ready Signal

```yaml
phase: run
spec: SPEC-TODO-TRANSITION-STAMPS-001
status: implemented-pending-sync
run_complete_at: 2026-09-29
run_commit_sha: <pending-backfill-run>
run_status: all-critical-majors-pass
ac_pass_count: 12
ac_fail_count: 0
ac_pass_with_debt:
  - acceptance.md §D.5 coverage clause on internal/cli: 84.1% observed vs the 85% gate; pre-change baseline unmeasured (see §E.2)
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (lane worktree; integration via lead's window per gitflow-lane-protocol §4)
l44_post_push_fetch: not-run (lane does not push; lead batch-pushes origin/develop)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  local_darwin: go build ./internal/kanban ./internal/cli exit 0
  windows_matrix: not-run-locally — CI owns the full-suite/OS-matrix verdict (lane-local discipline)
total_run_phase_files: 24
m1_to_mN_commit_strategy: per-milestone commits (golden baseline → M1 schema → M2 stamping → M3 verdict → M4 exposure → M5 docs → AC-selector fix → full-suite repairs); one logical unit per commit
served_model_gate: no denial observed in this run session (§E.1 exemption in force; nothing routed around)
schema_decision: freeze test updated as recorded decision — items += picked_at/dropped_at; archived_items += picked_at/dropped_at/archived_at/landing_verdict; schema_version stays "1"
no_stored_query_sha: enforced structurally (LandingVerdict has no SHA field; encode validates verdict∈{landed,not-landed,unknown} + ref + at)
known_gaps:
  - TestLocalInstructions_UpdateDoctorPreserveFile fails on a pre-existing constitution-registry DRIFT (files untouched by this branch); owned by the registry's card, not this SPEC
  - internal/cli coverage baseline before this SPEC was never measured; the 84.1% figure's delta attribution is therefore unmeasured
residual_risks:
  - run1/run2 of the full cli suite ran while a foreign lane ran suites concurrently (observed via lsof cwd); run3, the recorded verdict, was uncontended
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-29
sync_commit_sha: 54973adcc   # D3 backfill exemption — backfilled by the following tiny commit
sync_status: completed
sync_agent: manager-docs (card t1310)
changelog_entry_position: [Unreleased] > Added > first entry (line 12 region, above SPEC-COMMIT-IDENTITY-GUARD-001)
pre_emission_grep_count: 0               # grep -c 'SPEC-TODO-TRANSITION-STAMPS-001' CHANGELOG.md before emission
ac_count_match: 12 distinct AC-TST-001..012 in acceptance.md == 12 referenced in CHANGELOG entry
file_path_verification: all 15 implementation/test paths cited in the CHANGELOG entry verified present via ls
frontmatter_status_transitions:
  spec.md: in-progress -> implemented -> completed (merged sync close; updated: 2026-09-29, already current date — no edit needed)
mx_tag_validation:
  well_formed: yes
  counts_by_file: {backlog_store.go: 16, backlog_migrate.go: 2}
  added_by_sync: 0
  removed_by_sync: 0
  note: run phase placed all tags; landing_verdict.go exported funcs (EncodeLandingVerdict/DecodeLandingVerdict/LandingVerdictValue) carry no @MX tags — reported, not fixed (source edits out of sync scope)
observed_context_not_judged:
  - Security-Scanner flagged 3 sql-injection findings on the migration code (string-built ALTER TABLE with constant column names — ensureLandingColumn precedent pattern); sync-auditor judges
canary_compliance_check:
  template_neutrality: run-phase docs changed template gtd.md + mirror + catalog.yaml only; no SPEC ID / card id / internal date / commit SHA classes introduced by sync (sync touched no template files)
```

