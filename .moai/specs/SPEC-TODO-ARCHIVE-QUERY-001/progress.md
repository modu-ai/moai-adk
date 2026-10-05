# Progress — SPEC-TODO-ARCHIVE-QUERY-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (corrected from S at plan-audit iteration 1).
- Artifacts: spec.md, plan.md, acceptance.md, progress.md — the Tier M set
  (spec + plan + acceptance) plus progress.md, which every Tier emits.
- **Tier rationale — the REQ/AC ceiling is the deciding axis.**
  `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier caps
  Tier S at **8 requirements and 8 acceptance criteria**, and states that
  exceeding either ceiling *"is a signal to tier up or to split the SPEC, not to
  relax the budget"*. This SPEC carries **15 and 15** — an 87% overrun of the S
  ceiling, comfortably inside M's 16/16. That axis alone settles it, and it is
  the axis the first classification omitted.
  - Files: **≥ 7 paths**, inside M's 5-15 and outside S's `< 5`:
    `internal/cli/todo_history.go` (new), the registration edit in
    `internal/cli/todo.go`, a `./internal/cli/` test file, a
    `./internal/kanban/` test file (AC-TAQ-012 runs there), both skill documents
    (local + template mirror), and the golden directory
    `internal/cli/testdata/golden/live-readers/` that M0 commits. The first
    classification counted 5 and read the boundary toward S; 5 is M's lower
    bound, and the count was itself short.
  - LOC: still well under 300, which is S guidance. It is the only axis pointing
    at S, and it is guidance rather than a ceiling — it does not outrank two
    ceilings that are.
  - Consequence, stated plainly because it is the part that matters: the
    plan-auditor PASS threshold is **0.80**, not 0.75. The first classification
    resolved a genuine boundary in the direction that lowered this SPEC's own
    bar.
- Premise re-check: `.moai/reports/t394/premise-check.md` — nine measurements
  against `origin/develop@2c18091d1`, all nine of which reproduce under
  independent re-run. A separate claim in that report's § Gaps — a queue-path
  divergence — did NOT reproduce and is withdrawn; see `spec.md` §E and
  `plan.md` §B.2. The dispatching card's premise (that `done` destroys) was
  found false and is not carried into the SPEC.
- Plan-audit iteration 1: FAIL, 0.775 against the 0.80 threshold. Six blocking
  defects (D1-D6) repaired; response at
  `.moai/reports/t394/audit-response-iter1.md`.
- Plan-audit iteration 2: **PASS-WITH-DEBT, 0.875** against the 0.80 threshold
  (monotonic 0.775 → 0.875, no dimension regressed). All eight iteration-1
  closures were independently re-verified and confirmed; all seven must-pass
  criteria pass. Seven new defects raised — four blocking-class (D1-D4), three
  optional (D5-D7). Report at `.moai/reports/t394/plan-audit-iter2.md`; response
  at `.moai/reports/t394/audit-response-iter2.md`. Per the Retry Loop Contract,
  Tier M's ceiling is 2 iterations and iteration 2 passes, so **no iteration 3
  is authorised**; the seven defects were repaired in-place rather than routed
  into another audit round.
- Iteration-2 disposition: **7 of 7 closed** — D1 (deviation ledger corrected
  against this repository's actual `fetch-depth: 0`), D2 (AC-TAQ-011 modify-path
  escape closed by a byte-integrity assertion; overclaiming prose narrowed),
  D3 (singleness asserted, merge-method dependency recorded, develop-drift
  residual stated), D4 + D5 (measurement provenance block; 112/113 reconciled as
  two instants of a live counter), D6 (fixture convention names its two
  exceptions), D7 (constructor name fixed as a contract).
- **Run-phase debt carried forward** — the iteration-2 report's own carry-forward
  note, which the repair does NOT close: `origin/develop` has moved well past
  `2c18091d1` and is still moving. Every `file:line` in these artifacts was
  verified at `2c18091d1` and **none at the current develop head**. `plan.md` §C
  carries the obligation to re-run the pre-flight batch and re-verify every
  citation before the first run-phase edit. This is a gap by construction, not
  an oversight.
- Status: `draft`. No implementation code, tests, commits, or pushes exist.
- **Addendum 2026-09-01 — plan-phase correction round 3 (lead-mandated premise
  repair).** Lead disposition: **proceed, correction-first** — the alternative
  of proceeding without the premise repair was rejected and dropped. Trigger: a
  re-measurement (lead dispatch, `2026-09-01 16:20–16:25 KST`, against the
  primary checkout's live queue and the installed binary) postdating this
  SPEC's provenance snapshots disproved two premises: (1) the installed binary
  is a **develop build** (`v3.1.2-1033-g64bba61aa`) installed `06:37 KST`
  whose `done` **archives** — the artifacts' `v3.1.2`-main-era-deletes claim
  came from reading only `moai version`'s banner and missing the descriptor
  line printed on the same output; (2) the harness-selection card is **not
  lost** — `t393`, live in `items` (state `picked`). Corrections landed in
  spec.md (§A.2 incident note, §A.4 rewrite, `[MEASUREMENT PROVENANCE]` binary
  bullet, snapshot C recording the post-install archive tables and the ~289
  **pre-archive-era estimate**) and plan.md (§B.1 rewrite; §B.3/§B.4 binary
  stamps). acceptance.md swept clean — no AC text depends on the corrected
  premises. Frontmatter 0.2.1 → 0.2.2. No REQ/AC identifiers or text changed
  (the §A.4 figures are non-load-bearing by the SPEC's own design). No audit
  was run by this round — the artifact-hash change makes the run-entry Phase 1
  plan-audit gate re-execute by contract.

- **Run-entry Phase 1 gate (2026-09-01, post-correction): PASS 0.91** against
  the 0.80 threshold (trajectory 0.775 → 0.875 → 0.91, monotonic; dimensions
  0.90/0.90/0.90/0.95). Correction round 3 verified with zero new defects; the
  frozen 289 estimate survived an independent re-measure taken after the queue
  moved again (410 − 108 − 13 = 289; the two new closures decompose exactly).
  Must-pass 7/7 (MP-4 N/A — single-language SPEC). All 22 file:line citations
  re-confirmed at `e8ae9798a`, matching the §E.2 pre-flight record. Auditor
  gaps stated: AC-TAQ-011 clause 2 and the AC-TAQ-014 RED pair deliberately
  deferred to run-phase DoD. Recommendation proceed-to-M0 taken.
  Report: `.moai/reports/t394/plan-audit-run-gate-2026-09-01.md`.

## §E.2 Run-phase Evidence

### Pre-flight — §C batch re-run (orchestrator, run entry 2026-09-01)

Recorded before the first run-phase edit, per `plan.md` §C. All commands run in
this worktree against the primary queue; no edit has been made yet.

- `git fetch origin develop` → entry-time head **e8ae9798a**.
- Branch state re-read immediately before resolution (staleness rule): HEAD
  `2c18091d1`, branch `WT-todo-done-history`, 0 own commits. §C's divergence
  verdict read `N 0`; with zero local commits the only resolution is
  `git merge --ff-only origin/develop`. Post-merge re-read: HEAD `e8ae9798a`,
  `origin/develop...HEAD` = **0 0**. The untracked SPEC artifacts were
  unaffected by the ff (225 tracked files updated; none under `.moai/specs/`
  or `.moai/reports/` for this card).
- Citation re-verification at `e8ae9798a`: **all 22 cited spots resolve;
  0 refreshed, 0 blockers.** The ff diff touches NONE of the 8 cited files, so
  each is byte-identical across `2c18091d1..e8ae9798a`; every anchor was
  additionally re-located by direct read at the new head:
  - `internal/kanban/state_dir.go` — :37 `stateDirName = "todo"`, :43
    `legacyStateDirName`, :79 `func resolveStateDir`, :129
    `func BacklogPathForRoot` — exact.
  - `internal/kanban/backlog_sqlite.go` — :92-94 (IF NOT EXISTS comment at
    :93, in range), :113 (inside the DDL block), :124 `CREATE TABLE IF NOT
    EXISTS archived_items`, :133 `archived_findings` — exact.
  - `internal/kanban/backlog_store.go` — :14-17 (cited comment text at :16, in
    range), :223 `func (r *BacklogRecord) ArchiveCard`, :411 + :552
    (`inspectBacklogLayout` call sites; the cited regions 411-427
    `QueuedCount` and 551-568 `LoadPure` carry the db-preference pattern as
    claimed), :772-778 (body of `normalizeBacklogRecord`, defined at :751) —
    exact.
  - `internal/kanban/backlog_migrate.go` — :102 `e.readArchive(ctx, rec)`,
    :106-110 `readLastSeq` + `rec.LastSeq = lastSeq` — exact. The M2 ordering
    coupling was re-read at :102-110: `readArchive` precedes `readLastSeq`, and
    an archive error returns before `LastSeq` is populated — the M2
    requirement stands as written.
  - `internal/kanban/backlog_migrate.go` :411 `func loadLegacyBacklogJSON`;
    `internal/cli/todo.go` :148-152 (the 16 verb registrations), :310-311 (the
    render loop — `for _, it := range rec.Items` with no state filter), :409
    `rec.ArchiveCard(id)`; `internal/cli/todo_why.go` :35 (`no findings`,
    cited :34-37, in range); `internal/cli/todo_export.go` :75
    (`writeExportAtomic` call) + :100-110 (`discloseArchiveDowngradeCost` at
    :101); `internal/cli/todo_undone.go` :67-68 (the restore assignment, cited
    :68, in range) — all exact.

### M0 — goldens captured and committed (2026-09-01)

- **C = `e7eec6122`** — `chore(SPEC-TODO-ARCHIVE-QUERY-001): M0 capture
  live-reader goldens (t394)`, the only commit adding
  `internal/cli/testdata/golden/live-readers/`. Parent `e8ae9798a`
  (origin/develop head; the pre-verb tree every capture ran against).
- Capture provenance: binary built from tree `e8ae9798a`
  (`go build -o /tmp/t394-m0.HcM9Ew/moai ./cmd/moai`), run against the
  acceptance.md FIXTURE — a fresh `t.TempDir`-equivalent git repo at
  `/tmp/t394-fixture.YDuKsh` seeded through the public CLI: 4 adds →
  `next 2 --spec SPEC-X-001` (t2 picked) → `drop t3 "superseded by the
  parser rewrite"` (t3 dropped) → `relate t1 t3 --relation conflicts --note
  "needs a decision"` → `done t4` (t4 archived). Six readers captured in a
  fixed order: `list`, `list --json`, `next`, `why t3`, `analyze`, and the
  state counts.
- **Normalization rule (applies to `list-json.txt` only)**: RFC3339 UTC
  wall-clock stamps (`added_at`, findings' `at`) are replaced with
  `<RFC3339>` — `sed -E 's/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z/<RFC3339>/g'`
  at capture; the clause-3 test applies the identical Go regexp. All other
  streams are captured verbatim.
- **State-counts capture**: `json.Marshal(kanban.BacklogCountsForRoot(root))`
  measured through a scratch `go run` package (`cmd/m0counts`, removed before
  the commit; `git status --short` showed only the golden dir as new).
  Result: `{"Picked":1,"Queued":1,"Available":true}`.
- Golden content check: `list.txt` carries t1/t2/t3 (queued/picked/dropped)
  and NOT t4 — the archive's invisibility to the cheapest read is itself in
  the golden (REQ-TDG-007 baseline).

### M1 — the output contract and the four fates (2026-09-01)

- `internal/cli/todo_history.go` — `newTodoHistoryCmd` registered at the
  `internal/cli/todo.go` verb block (grew 16 → 17). Read path:
  `LoadPure` (never adopts, never migrates, lock-free — REQ-TAQ-010's read
  posture). Line shape: tab-separated, card text LAST
  (`<id>\tlive\t<state>\t<text>` / `<id>\tarchived\t<state-at-archive>\t<text>`
  / `<id>\tabsent`); bare listing walks the stored archive order backwards
  (newest first), empty archive prints the explicit line.
- **RED-1 (E8, verbatim)** — `go test ./internal/cli/ -run TestTodoHistory
  -count=1` on the pre-verb tree, 6/6 FAIL for the right reason (verb
  absent): 4× `todo: "history" is not a todo verb and "tN" is a card id —
  refusing to create a card named "history tN"` (the mistyped-verb guard,
  which a registered verb bypasses by routing to the subcommand), 2× `todo:
  unknown command "history" for "todo"`. Recorded from this run before any
  implementation existed; GREEN after `todo_history.go` + registration.
- **Incident + repair (primary-queue pollution)**: the FIRST RED attempt ran
  the tests without `todoFixture(t)`; queue-root resolution (t106) walked the
  worktree to the primary checkout and 7 seed cards (t414-t420, all
  `queued`, my fixture texts) landed on the operator's live queue
  (last_seq 413→420). No picks/drops/dones/findings landed on real cards —
  every later step in that run was REFUSED (no backlog item t2/t1). Repair:
  `done t414..t420` via the public CLI (archived, reversible via `undone`);
  verified primary `todo list` grep for the 7 texts = 0 hits. Lead notified.
  Test file fixed: every test now enters through `todoFixture(t)`.
- Delivers REQ-TAQ-001/002/003/005/006/009; closes AC-TAQ-001/002/003/005/
  006/009 (E1 matrix commands in §E.3).

### M2 — the honest limits (2026-09-01)

- REQ-TAQ-004: `absent` for an id at or below the issued-id mark carries a
  stderr qualifier keyed on `rec.LastSeq` — never on archive emptiness.
  Uses the exported `kanban.ParseBacklogSeq` (thin wrapper over
  `parseBacklogSeq`; a second parser would drift from the issued id form).
  Ordering coupling verified: every reachable degraded path completes
  readRecord's `readLastSeq` (backlog_migrate.go:106-110), so the mark is
  present where the note is most needed.
- REQ-TAQ-013: new `kanban.InspectBacklogArchiveVouch(queuePath)`
  (`internal/kanban/backlog_archive_vouch.go`) names which store answers —
  SQLite / legacy backlog.json / none — and whether it can vouch for an
  archive. **Load-bearing ordering**: the probe reads `sqlite_master` from
  its own connection BEFORE any engine open, because the DDL's
  `IF NOT EXISTS` recreates missing archive tables and would erase exactly
  the fact the disclosure reports. The verb probes, then reads via
  `LoadPure`; disclosures are stderr-only.
- **Degraded shape is consumed by the first read**: a dropped-tables
  database gets its tables recreated by the first history invocation's
  engine open (the store's universal open behavior — `list` does the same),
  so each AC scenario runs one fresh surgery per invocation. Recorded in
  the test comment.
- **Milestone-mapping deviation (documented)**: plan §F placed REQ-TAQ-008
  (truncation notice) in M2, but its observable — the withheld count —
  cannot fire before the bound exists (M3). The notice lands WITH `--limit`
  in M3; AC-TAQ-008 closes there. M2 closes AC-TAQ-004 and AC-TAQ-013.
- RED-2 (E8, verbatim): pre-implementation run — kanban
  `undefined: InspectBacklogArchiveVouch / BacklogStoreSQLite /
  BacklogStoreLegacyJSON` (build failed, 4 tests); cli
  `history t3 stderr = "", want the at-or-below-the-mark qualifier`,
  `history t1 stderr = "" / history (listing) stderr = ""`, want the store
  disclosure, `legacy_json_only` stderr empty — 5 assertion failures, all
  "missing disclosure" shaped. GREEN after the probe + wiring; one test
  revision (fresh surgery per invocation) and one ineffassign fix.

### M3 — the bound (2026-09-01)

- `--limit <n>` on the listing: default 20 (`todoHistoryDefaultLimit`),
  0 = unbounded, negative refused with an error. Truncation states the
  withheld count on stderr (`history: N archived entries withheld —
  showing M of T (--limit 0 lists all)`), stdout unaffected — REQ-TAQ-007
  + REQ-TAQ-008 together, per the M2 deviation note.
- RED-3 (E8, verbatim): pre-implementation run — `default listing carries
  25 lines, want the default bound 20`; `history --limit 5: unknown flag:
  --limit` (×2). GREEN after the flag + bound + notice.
- Closes AC-TAQ-007, AC-TAQ-008.

### M4 — the invariant guards (2026-09-01)

Tests: `TestLiveReadersUnchangedByHistoryVerb` (clause 3),
`TestTodoHistoryLeavesStorageByteIdentical` (AC-TAQ-010),
`TestTodoHistoryAddsNoSchemaChange` (AC-TAQ-012, `internal/kanban/
backlog_schema_freeze_test.go`), `TestTodoHistoryNeverPrompts`
(AC-TAQ-014).

- **Clause 1 batch (verbatim results, this worktree @ HEAD `1a5d8664f`)**:
  adding-commit count = **1**; `C = e7eec612213b1a6390ad627005a3066cb09b9a36`;
  `C != HEAD` (HEAD `1a5d8664fc3eff0e03282b941ed80e726778cd61`);
  `merge-base --is-ancestor` exit **0**; `cat-file -e C:internal/cli/
  todo_history.go` exit **128** (absent → `!` passes); `git grep -q
  newTodoHistoryCmd C -- internal/cli/` exit **1** (absent → `!` passes);
  `git diff --exit-code C -- goldens` exit **0** (bytes unmoved).
- **Clause 1 reasoned rows observed RED once** (throwaway repos, per
  acceptance.md's obligation): orphaned-goldens repo — `test -n` exit **1**
  and `merge-base --is-ancestor` exit **1** (unverifiable provenance does
  not read as verified); single-commit repo standing in for a squash —
  `cat-file` exit **0** and `grep -q` exit **0** with `C == HEAD`
  (`test !=` exit 1), i.e. every `!` form correctly FAILS when C's tree
  holds the verb.
- **Clause 1 failure-mode observation on the real tree (guard bites)**:
  mutating `list.txt` by one byte in the working tree turned clause 3 RED
  with an exact got/want diff; `git restore` returned the golden, `git
  status --short` clean, clause 3 GREEN again.
- **Clause 2 regeneration (local DoD gate)**: binary built from C in a
  throwaway detached worktree, run against a fresh FIXTURE with the M0 seed
  sequence — `IDENTICAL list.txt / list-json.txt / next.txt / why_t3.txt /
  analyze.txt` (cmp, with the capture-time RFC3339 normalization on
  list-json). The state-counts stream is exercised by the clause-3
  in-process comparison (`kanban.BacklogCountsForRoot` render). Worktree
  removed after.
- **AC-TAQ-014 plant-and-remove pair**: `bufio.NewReader(os.Stdin)`
  planted at `todo_history.go:80` → the AC grep matched (exit 0 → the `!`
  guard form fails) AND `TestTodoHistoryNeverPrompts` RED naming the exact
  line; removed → `git diff` empty (byte-exact revert), grep exit **1**
  (guard passes), test GREEN. Both observations recorded here.
- AC-TAQ-010 note: the seed's `done` legitimately leaves `backlog.lock`
  (flock Unlock does not unlink on Unix), so the test removes it before
  the reads and asserts history does not recreate it — the mechanical form
  of "no lock artifact remains" on this platform; the SHA-256 map over the
  whole state dir additionally asserts no file changed and none appeared.

### M5 — documentation, both surfaces (2026-09-01)

- `history` row added to `.claude/skills/moai/workflows/todo.md` and its
  template mirror (inserted after the `why` row in each). Mirror text
  carries no SPEC ID, REQ token, date, or SHA.
- `make build` run (Template-First; exit 0, binary rebuilt).
  `internal/template/catalog.yaml` unchanged BY DESIGN — its skill hashes
  cover root SKILL.md only (`gen-catalog-hashes.go` hashes the directory's
  SKILL.md, not workflows/*.md), so this edit does not move it.
- Pre-existing drift observed and left alone (scope discipline): the local
  `todo.md`'s `pr` row is an older rendering than the mirror's (the mirror
  names the six-column shape and the text-LAST convention); reconciling
  template→local propagation is `moai update`'s surface, not this SPEC's.
- AC-TAQ-015: `TestTodoSkillDocumentsHistoryVerb` — both docs ≥ 1 mention,
  neutrality regex clean over the mirror (grep counts: 1 and 1).

### Post-M5 findings and repairs (2026-09-01)

- **Verb-surface zero-delta guard** — the E3 coverage sweep surfaced
  `TestTodoVerbSurfaceZeroDelta` (SPEC-TODO-SQLITE-001 AC-TOSQ-010) RED:
  `history` was an undeclared post-freeze addition. Declared per the
  guard's own convention in `permittedVerbAdditions` (commit `05cddca2b`);
  the guard's mechanism is declaration-with-SPEC, not prohibition.
- **Negative `--limit` refusal re-derived test-first** — the refusal branch
  from M3 had no failing test (Invariant ii). Branch deleted, RED observed
  (`history --limit -1 succeeded`), branch restored, GREEN (commit
  `c39b9790d`).

_<pending run-phase evidence — none; run phase complete>_

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-09-01T13:05:00+09:00"
run_commit_sha: "c39b9790d"
run_status: complete
ac_pass_count: 15
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "see §E.2 M0 — entry-time ff to e8ae9798a; divergence 0 0 re-read before every commit (staleness rule); no push performed (integration is the lead's window)"
l44_post_push_fetch: "n/a — no push in run phase (gitflow: card branch stays local; origin/develop CI is the post-integration verdict)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: "go build ./... → exit 0 (c39b9790d)"
  windows: "GOOS=windows GOARCH=amd64 go build ./... → exit 0 (c39b9790d)"
total_run_phase_files: 13
m1_to_mN_commit_strategy: "per-milestone commits on WT-todo-done-history: e7eec6122 (M0 goldens) → 95105fcad (M1 verb) → 244828510 (M2 disclosures) → 1a5d8664f (M3 limit) → 1b77fd7eb (M4 guards) → 2f92574cd (M5 docs) → 05cddca2b (surface-guard declaration) → c39b9790d (negative-limit RED/GREEN); C = e7eec6122 is a proper ancestor of HEAD (verified)"
coverage_note: "todo_history.go per-function: newTodoHistoryCmd 100%, runTodoHistory 75%, renderTodoHistoryLookup 92.3%, renderTodoHistoryListing 86.7% — uncovered lines are io-write-error returns only; kanban backlog_archive_vouch.go: InspectBacklogArchiveVouch 100%, archiveTablesPresent 86.7%; package totals: internal/cli 80.2%, internal/kanban 86.5% (whole-package baselines, not this SPEC's subject)"
```

Milestone-mapping deviations (both recorded in §E.2): AC-TAQ-008 closed in
M3 with REQ-TAQ-007 (the withheld count cannot fire before the bound
exists); the verb-surface declaration and the negative-limit RED/GREEN are
post-M5 hardening, both cascade follow-ups within the SPEC's scope
envelope.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-09-01T23:27:00+09:00"
sync_commit_sha: "973832f94"
sync_status: complete
changelog_entry_position: "CHANGELOG.md [Unreleased] > ### Added, first bullet (top insertion)"
b12_self_test_a: "pre-emission grep -c 'SPEC-TODO-ARCHIVE-QUERY-001' CHANGELOG.md → 0 (halt condition not met; duplicate-entry halt not triggered)"
b12_self_test_b: "AC count match — acceptance.md carries 15 live identifiers (AC-TAQ-001..015, zero RETIRED/REF markers); the CHANGELOG entry cites 15/15 — equal, and the count is non-zero"
b12_self_test_c: "every path cited in the entry verified present by ls — internal/cli/todo_history.go, internal/cli/todo.go (registration, todo.go:152), internal/cli/todo_history_test.go, internal/kanban/backlog_archive_vouch.go, internal/kanban/backlog_archive_vouch_test.go, internal/kanban/backlog_schema_freeze_test.go, internal/cli/testdata/golden/live-readers/ (6 golden files), .claude/skills/moai/workflows/todo.md, internal/template/templates/.claude/skills/moai/workflows/todo.md"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (the implemented→completed close merged into this single sync commit per the 3-phase close; updated: 2026-09-01 already current — same-day, no change needed)"
  plan_md: "no status: field to transition (ArtifactStatusFieldForbidden, per card t357)"
  acceptance_md: "no status: field to transition"
  progress_md: "no status: field to transition"
canary_compliance_check:
  mx_tag_validation: "grep '@MX' over the five SPEC-authored Go files (todo_history.go, todo.go, todo_history_test.go, backlog_archive_vouch.go, backlog_archive_vouch_test.go) → 0 matches; this SPEC added and removed no MX tags; new files carry SPEC-ID header comments and exported symbols carry godoc — no MX drift"
  spec_lint: "0 error(s), 1 warning(s) — WARNING MovingRefUnpinned at plan.md:94 (plan BODY; reported to the lead per the ownership boundary, not repaired in sync)"
  codemaps: "unchanged by decision — entry-points.md does not enumerate moai todo sub-verbs, and the verb joined an existing AddCommand call site (internal/cli/todo.go:150-152), so the recorded call-site and root-command counts stand; hand-editing a dated generated artifact is ceremony the next /moai codemaps --force would overwrite"
  docs_site: "not touched — adk.mo.ai.kr is owned by a separate harness"
  scoped_smoke: "go test -count=1 ./internal/cli/ -run TestTodoSkillDocumentsHistoryVerb → ok (both skill documents ≥1 mention, template-mirror neutrality regex clean)"
```

## §F Phase 4 Mode Selection

Logged by the orchestrator at run entry (2026-09-01), before the first
run-phase `Agent()` spawn.

**Input parameters**: tier M; scope ≈ 8 paths (new `internal/cli/todo_history.go`,
registration edit in `internal/cli/todo.go`, test files under `internal/cli/`
and `internal/kanban/`, the golden directory
`internal/cli/testdata/golden/live-readers/`, skill doc + template mirror);
domain count = 2 (backend Go + docs/template mirror); language mix = Go +
markdown; concurrency benefit LOW — M0→M1→M2→M3→M4→M5 is a strict dependency
chain (M0 must precede M1 mechanically per AC-TAQ-011; M4's guards assert M0's
commits; M2's disclosures are judgement work, not parallelizable research);
Agent Teams prereqs not requested — the operator selected autonomous serial
progression at the Implementation Kickoff Approval gate.

**Mode evaluation**:

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | multi-file Go implementation, not a typo-scale change |
| serial | **YES** | coding-heavy; strict milestone dependencies; single writer in this worktree |
| fanout | no | no research-heavy multi-domain fan-out; parallel writes would race the shared tree |
| sweep | no | < 30 files; not one uniform mechanical transform |

**Decision**: `serial` — one manager-develop sub-agent carrying M0→M5
sequentially, cycle_type=tdd.

**Justification**: the plan's own ordering decides it. M0 exists to precede M1
mechanically (AC-TAQ-011 clause 1), M4 writes invariant guards over M0's
commits, and M2's three disclosures are operator-honesty judgements. Anthropic's
coding-task parallelism caveat applies; a single write-capable agent in this
worktree avoids file races entirely. `sweep`'s and `fanout`'s entry
preconditions are unmet on their own terms.

**Boundary cases**: none — every branch of the decision tree resolved
decisively.

Gate context at selection time: iter-1 FAIL 0.775 → iter-2 PASS-WITH-DEBT
0.875; correction round 3 changed artifact hashes, so the run-entry Phase 1
plan-audit gate re-executes before M0.
