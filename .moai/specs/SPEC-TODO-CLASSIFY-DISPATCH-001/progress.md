# SPEC-TODO-CLASSIFY-DISPATCH-001 — Progress

Tier M · card t1332 · plan-phase artifact set authored 2026-09-29 at HEAD `145c3d98c`
(worktree `.moai/worktrees/t1332`, branch `WT-card-autodispatch`). Status: `in-progress`
(run phase entered 2026-09-29, cycle_type=tdd; run-phase basis HEAD `51f3e9878` = plan
commits `875a4b33b` + clean develop absorb `8fc5a7407` carrying the t1240 surface —
REQ-TCD-013 entry precondition verified, see §E.2).

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifact set complete: `spec.md` (14 REQ, GEARS), `plan.md` (§A-§H, 5 milestones,
  decision records OD-1..3 — OD-1/OD-3 folded as leader rulings 2026-09-29, OD-2 open — plus one
  flagged lead follow-up on the failure-default vs read-default tension), `acceptance.md`
  (AC-TCD-001..014, Given-When-Then, RED-now/green adoption table), this `progress.md`.
- Revision 0.2.0 folded the leader rulings of 2026-09-29: OD-1 serial-card mutual exclusivity
  (pipeline exclusivity rejected), OD-3 serial failure default; provenance recorded in plan.md
  §C (noul OD-1 0.31 / OD-3 0.36). D1 citation-prefix repair applied. Post-repair lint:
  0 findings.
- Revision 0.3.0 resolved the flagged tension by leader ruling (2026-09-29, OD-3 extension):
  REQ-TCD-014's absent-field READ mode default flips parallelizable → serial (priority/blocked
  defaults untouched); plan.md §C records the resolution. No read-default AC added; no existing
  AC asserted the old default. Post-revision lint: 0 findings; REQ/AC 14/14 unchanged.
- Measured surface basis exported: `.moai/reports/t1332/surface-notes.md` (all card premises
  verified; measured corrections recorded — t1240 branch 28 commits ahead of develop, tip
  `d43e50bb3`, `9866ca25e` an ancestor).
- SPEC ID regex pre-write check: PASS (`SPEC-TODO-CLASSIFY-DISPATCH-001`); ID unique in
  `.moai/specs/`.
- Frontmatter validated against the canonical 12-field schema SSOT; `status: draft` set at
  creation per plan-phase ownership.
- Out of Scope section carries five `### Out of Scope — <topic>` H3 sub-headings with `-` bullets
  (t1240 / t1306 / t1261 / Jev / scheduling-intelligence boundaries).
- Known bounded gaps: none. The absorption-order precondition (REQ-TCD-013) is a run-phase entry
  gate by design, not a plan-phase gap.

## §E.2 Run-phase Evidence

Run phase entered 2026-09-29, cycle_type=tdd, lane worker-62 (card t1332).

- **Entry precondition (REQ-TCD-014 / AC-TCD-014)**: verified at entry — this branch's HEAD
  `51f3e9878` is the plan commits plus the clean absorb of develop `8fc5a7407`, which carries the
  t1240 surface (`newFactoryNextCommand` :459 / `newFactoryStageCommand` :547 /
  `newFactoryCompleteCommand` :640 in `internal/cli/factory_card.go`). AC-TCD-014 is now a pinned
  regression test: `internal/cli/factory_entry_precondition_test.go`
  (`TestFactorySelfDispatchSurfacePresent`, `TestAbsorptionPreconditionDocumented`) — both GREEN.
- **M1 — classification data model (GREEN)**: `internal/kanban/classification.go` (closed value
  sets, validation, JSON parse, read-default derivation, SQLite value codec);
  `BacklogItem.Classification *CardClassification json:"classification,omitempty"`
  (backlog_store.go); `classification` TEXT column appended last on items + archived_items via the
  pragma_table_info-gated ADD COLUMN path (backlog_sqlite.go); read/write wiring in
  backlog_migrate.go following the landing/columnExpr contract. RED evidence captured before GREEN
  (compile-failure output on the missing API, recorded in the M1 commit message trail); GREEN:
  - `go test ./internal/kanban/ -run 'TestCardClassification|TestParseCardClassification|TestEffectiveCardClassification|TestBacklogClassification|TestSchemaFreezeCarriesClassification' -count=1` → `ok github.com/modu-ai/moai-adk/internal/kanban 0.418s` (this run, this tree)
  - `go test -timeout 30m ./internal/kanban/ -count=1` → 2 failures BEFORE the freeze-tuple update
    (`TestSchemaFreezeRecordsTransitionStamps`, `TestTransitionStampColumns_FreshUpgradedConverge`
    — both stale pins, updated in the same milestone); full suite re-run pending after M2/M3 land.
  - AC-TCD-013 (jev boundary + positive control): `go test ./internal/cli/ -run
    'TestProductPathsCarryNoJevReference|TestJevBoundaryScanPositiveControl' -count=1` →
    `ok ... 0.779s`. The positive control is a planted fixture (scripts/jev is untracked, so the
    control must be clone-independent). Product-path hits: 0; `_test.go` files excluded per the
    B3 boundary-grep precedent.
- Files touched (M1): internal/kanban/{classification.go,classification_test.go,backlog_store.go,
  backlog_sqlite.go,backlog_migrate.go,backlog_schema_freeze_test.go,backlog_transition_stamps_test.go},
  internal/cli/{product_jev_boundary_test.go,factory_entry_precondition_test.go}.

### M1-M5 commits (branch WT-card-autodispatch)

- M1 `3d89b598d` — classification data model (see above).
- M2 `9896888f9` — CardDecider seam (Default/Static/Unavailable implementations), add-path wiring
  inside the locked write, `--classification-file` validated input (exit-2 refusals; transport
  unavailability degrades to the fail-safe fallback with one notice), SortByClassification +
  QueuedPosition re-establishing sorted order inside every add's locked write (CLI append/pick,
  MCP todo_add, store.Add), printed position = sorted 1-based position; todo surface guard
  declares the one permitted flag addition.
- M3 `de490ea53` — mode-aware factory selection: sorted-order auto-promotion, blocked never
  auto-selected, serial-vs-serial mutual exclusivity with the candidate's own row excluded by
  identity (self-wedge defect caught by the t1240 suite before my own ACs), POSITIVELY enumerated
  terminal set {done, abandoned}, parallelizable concurrency pinned over the version-checked
  store, factoryCardView mode/priority cells (text + --json).
- M4 `9b39ccb88` — `-f` lane auto-dispatch default-on; `--no-auto-dispatch` lane-only opt-out
  (clear-policy precedent); the stamp always overwrites into `MOAI_FACTORY_AUTO_DISPATCH`
  (internal/config constants); the SessionStart lane rule reads the stamp — a manual launch
  carries a lease-verb-free manual rule in all four locales; absence reads as the auto default.
- M5 `bed109bfc` — t1306/t1308 interaction regressions (the serial cycle consumes queue order,
  which is now the classification order, with zero cycle changes; a held card stays invisible to
  selectors while sorting positions it), gtd.md doc parity on both surfaces (+ the neutral
  mirror), stale column-tuple pins updated, `454945e99` catalog.yaml hash cascade, `68c9fbd11`
  AC-TCD-003 positive-control strengthened into the same test.

### Absorbed-suite integration repairs (commit `19b8c02ce`)

The first full internal/cli sweep surfaced six interactions between my
milestones and the absorbed t1240 surface; all repaired in the same run:

1. `factorySerialSlotFree` re-enumerated — the serial slot protects the
   ordering of IMPLEMENTATION work and releases at merge-ready and later
   (+ abandoned). The recorded behavior of the absorbed relaunch-loop suites
   (a lane leases its next card while the previous sits at merge-ready)
   decided the boundary; my first cut ({done, abandoned}) wedged them.
2. Auto-promotion filter rewritten in the positive form the REQ-THS-012
   guard scans for.
3. "lead" removed from the four manual-rule locales (vocabulary guard).
4. Mode-neutral t1240 fixtures (SD AC-008 arm 1; the m6 relaunch stubs —
   the crashed-lane shape) classify their cards parallelizable, intent
   recorded in place. Unclassified reads serial, which would wedge the
   ordering those tests do not test.
5. list-json.txt golden re-captured with a provenance note.
6. The legacy-record contract test admits the declared addition as
   present-or-absent per row (never a third shape).

NOT repaired here — pre-existing at this branch's base `51f3e9878`,
attributed by graph (`git diff 51f3e9878..HEAD` touches none of the files;
fixtures unchanged from base): the launcher trio
(TestCC_FactoryEntryThroughRunCC, TestGLM_FactoryLaneEntry,
TestGLM_FactoryLeadRunIsJoinableByLane — the REQ-SD-005 git-tree refusal
against non-git temp fixtures) and the i18n dictionary pair in
internal/hook + internal/web (TestDataI18nKeysSubsetOfDictionary,
TestI18nKeySetParity). Reported to the lead as baseline debt outside this
SPEC's scope envelope.

### Verification results (this run, this tree — HEAD at measurement noted per line)

- `go vet ./internal/kanban/ ./internal/cli/ ./internal/hook/ ./internal/web/` → clean (0 findings).
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `make build` → exit 0 (agents-emit + commands-emit checks pass; catalog.yaml hash cascade
  committed separately).
- `golangci-lint run` (v2.1.6, the CI version) over internal/kanban, internal/cli, internal/hook,
  internal/web → 0 issues.
- `go test -timeout 30m ./internal/kanban/ -count=1 -cover` → `ok ... 194.293s coverage: 85.2% of
  statements` (this run, this tree at `68c9fbd11`).
- `go test -timeout 30m ./internal/cli/ -count=1 -cover` (final, at `56c3dd61e`) → `coverage:
  77.2% of statements`, failing EXACTLY the three pre-existing launcher tests (above) — every
  test this SPEC added or repaired is GREEN in the same run. New-code coverage measured by
  `go tool cover -func`: todoClassifyInLock / todoApplyClassification 100%,
  todoDeciderFromClassificationFile 90%, factorySerialSlotFree 100%,
  factoryNextSelectAndLease 84.7%. The 77.2% package figure is this 1000+-file package's
  pre-existing baseline; the SPEC's own surface is covered at or above the 85% floor.
- `go test -timeout 30m ./internal/cli/ -count=1 -cover` (first full run, at `19b8c02ce`'s
  parent) had surfaced the six interactions listed above — none remain in the final run.
- `go test -timeout 30m ./internal/hook/` → the two stable pre-existing reds
  (TestContractRoleScopedAllowWithoutLaneMarker, TestHMPSourceGuardGoLiterals) plus one
  intermittent pre-existing load flake (TestScanWriteContentNoConfigNoTempFile — cross-test
  temp-dir observation window; the i18n dictionary pair did not fire in this run, having fired
  in the previous one — intermittent, and none of the four reads files this SPEC touches).
- MX tags: @MX:ANCHOR added on `EffectiveCardClassification` (sole absent-field default seam) and
  `SortByClassification` (sole queue-order restorer), each with @MX:REASON + @MX:SPEC. No tags
  removed; no existing tag touched.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-29
run_commit_sha: 56c3dd61e
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (lane session; no push, per the git-flow lane protocol)
l44_post_push_fetch: n/a (lane session; push is the lead's batch act)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin_arm64: pass
cross_platform_build.windows_amd64: pass
total_run_phase_files: 30
m1_to_mN_commit_strategy: per-milestone commits M1..M5 + catalog cascade + AC-003
  strengthening + absorbed-suite integration repairs + MX tag commit
```

- Known baseline debt carried INTO this branch, not introduced by it (attribution in §E.2):
  TestCC_FactoryEntryThroughRunCC, TestGLM_FactoryLaneEntry,
  TestGLM_FactoryLeadRunIsJoinableByLane, TestDataI18nKeysSubsetOfDictionary,
  TestI18nKeySetParity — all pre-existing at base `51f3e9878`.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
