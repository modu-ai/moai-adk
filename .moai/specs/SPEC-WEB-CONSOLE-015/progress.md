# SPEC-WEB-CONSOLE-015 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored (Tier L set complete): `spec.md`, `plan.md`, `acceptance.md`, `design.md`,
  `research.md`, plus this `progress.md`.
- Tier: declared `L`; the honest measurement (`plan.md` §B) puts this revision near the M/L
  boundary and recommends M. Operator decision, flagged rather than taken.
- SPEC ID regex check executed, output `PASS`.
- Version `0.2.0`: three-way carve-out. Session telemetry → `SPEC-SESSION-TELEMETRY-001`; record
  keying, lane number, card identifier → `SPEC-KANBAN-RECORD-SESSION-KEY-001`; todo queue →
  `SPEC-WEB-TODO-QUEUE-001`. This SPEC is consumer-only.
- Budget: **12 requirements / 14 acceptance criteria** (ceiling 25 / 25; version 0.1.0 stood at
  25 / 29, which is what forced the split).
- Iteration-2 audit findings closed here: F2 (REQ/AC-WC15-012 deleted), F7 (duplicate-PID rule
  promoted to REQ-WC15-047 + AC-WC15-047), F9 (AC-WC15-002 restated as an executable inventory),
  F11 (AC-WC15-043a's file-creation clause given a directory listing), F12 (note banner brought
  into scope as REQ-WC15-052 + AC-WC15-052 and into the §C.3 surface table), F13 (GEARS form,
  implementation detail moved out of requirement bodies), F8/N2 (version + HISTORY row).
  F1, F3, F4, F5, F6, F10 left with the carve-outs.
- Two claims of version 0.1.0 corrected by measurement: the launcher is not the model/effort
  producer (`spec.md` §A.2) and the lane join does not close on today's data (`spec.md` §A.4).
- Dependencies: `SPEC-SESSION-TELEMETRY-001` and `SPEC-KANBAN-RECORD-SESSION-KEY-001` must both
  land first.
- Status: `draft`. Awaiting a full plan-audit of this revision.

## §E.2 Run-phase Evidence

Run-phase base `40f064c6d`; every measurement below was taken at `f80bad4d0` unless the row states
otherwise. Verification artifacts live under `.moai/state/verify/t207/`.

### Per-criterion matrix (14 criteria)

| Criterion | Command | Observed | Status |
|---|---|---|---|
| AC-WC15-001 | `git diff 40f064c6d..HEAD -- internal/web/events.go internal/web/assets/app.js` | empty diff, exit 0 | PASS (preservation) |
| AC-WC15-002 mech. | the write-surface inventory grep | exactly `profile_crud.go:38` + `:64` | PASS |
| AC-WC15-002 read. | read `internal/web/profile_crud.go:33-64` | both paths derive from `profile.GetProfileDir`; neither from the project state dir. Inventory unchanged from the pre-change baseline, so this reading is unchanged | PASS (declared human reading) |
| AC-WC15-021a | `grep -rnE 'SchemaVersion\|WriterPID\|CapturedAt\|ContextWindowSize\|TokensUsed\|RawPct\|Band\b' internal/web --include='*.go' \| grep -v _test` | two hits, both FIELD READS on the sibling's value (`viewmodel_ops.go:266,267`); zero declarations. `ReadSessionTelemetry` + `SessionTelemetryPath` present at `viewmodel_ops.go:246,250` (≥1). Every `SessionTelemetryRecord` occurrence is `statusline.`-qualified | PASS |
| AC-WC15-021b | `TestChainCellsCarryTelemetryValues` | model/effort/pct = `claude-opus-5` / `xhigh` / 42 where the pre-change tree yielded `""`/`""`/-1 | PASS |
| AC-WC15-023 | `TestChainCellsDoNotBorrowAnotherSessionsValues` | recorded session keeps its values; unrecorded session shows `""`/`""`/-1, never the neighbour's | PASS |
| AC-WC15-043a | `TestFactoryLanesResolveCompleteJoin` + `TestFactoryLanesJoinWritesNothing` | row carries the record's values; state-tree listing (path+size+mtime) identical before and after the join | PASS |
| AC-WC15-043b | `TestFactoryLanesPresentsUnresolvedLanes`, `TestLaneSectionRendersUnresolvedRows` | lane-4 (no session) and lane-6 (no record) both present, numbered, marked; neither carries join values | PASS |
| AC-WC15-044 | `TestLaneSectionRendersCompleteRow` | markup carries `data-lane="2"`, `t207`, `SPEC-EXAMPLE-001`, `state--live`, `stage--active` | PASS |
| AC-WC15-045 | `TestLaneSectionMarksEstimatedStageOnlyWhenEstimated` | estimated row carries `mark.estimated`; the unestimated (unresolved) row carries no stage mark at all | PASS |
| AC-WC15-046 | `TestLaneSectionPresentWithNoRegistry` (absent + malformed) | status 200 AND `kanban.lanes` section present AND `kanban.noLanes` rendered, in both cases | PASS |
| AC-WC15-047 | `TestFactoryLanesDuplicatePIDFactorySide` + `TestFactoryLanesDuplicatePIDSessionSide` | factory side: both lanes unresolved, neither carries the record. Session side: one lane, one pid, two entries → unresolved, no values | PASS (both halves) |
| AC-WC15-050 | `go test ./internal/web/ -run TestI18n` + `TestSPECIntroducedKeysResolveInEveryLocale` | governance suite passes with the allowlist file byte-unchanged; the introduced key set is 11 members, each present AND distinct from `en` in `ko`/`ja`/`zh` | PASS |
| AC-WC15-051 | `TestChainCellsTolerateSchemaV1Record` | within ONE render: the schema-1 row shows the marker for model/effort while still rendering its recorded 20%; the schema-2 row shows real values. Not both-marker | PASS |
| AC-WC15-052 | removal grep + `TestKanbanNoteBannerCorrected` | all three stale strings return zero across both templates; banner present with `kanban.note`; marker carries `data-i18n-title="mark.notRecorded"` and non-empty hover text; both keys resolve in four locales | PASS |

AC-WC15-052 *Content* half (declared human reading): the banner now reads "Stage is estimated from
the heartbeat. A blank model, effort or context cell means this session has no telemetry record
yet." and the marker reads "not recorded — this session has no record carrying this value yet".
Neither names `kanban.Record` or any other producer; each states that a blank cell means the
session has no telemetry record yet.

### Builds, tests, lint

- `go build ./...` → exit 0. `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `go test -count=1 ./internal/web/` → `ok … 4.146s`; with `-cover`, `coverage: 66.8% of statements`.
- `golangci-lint run ./internal/web/...` → `0 issues.`, exit 0 (`.moai/state/verify/t207/lint-head.log`).
- `go test -count=1 -timeout 1200s ./internal/cli/...` → 17 packages `ok`, zero `FAIL` lines,
  `CLI_EXIT=0` (`.moai/state/verify/t207/cli-suite.log`). Run in the background reading the log,
  never as a blocking foreground call: this package measured 546-866s in three runs today.
- Producer packages unmodified: `git diff --stat 40f064c6d..HEAD -- internal/statusline internal/kanban internal/session internal/cli` → empty.

### Gaps (explicitly NOT observed)

- **Coverage regression vs the pre-change tree is not directly measured.** The post-change rate is
  66.8%; the pre-change rate was not measured, because measuring it needs a checkout of `40f064c6d`
  and this run may create no second worktree. What WAS measured instead, at `f80bad4d0`: every
  function this SPEC adds is covered well above the package rate — `loadFactoryLanes` 100%,
  `resolveLane` 95.8%, `unresolvedLane` 100%, `readTelemetry` 85.7%, `telemetryCells` 100%,
  `buildChain` 94.7% — so the added statements cannot have pulled the package rate down. That is a
  bound, not the measurement the Definition of Done asks for.
- **The browser-side render is not observed.** Every render assertion here reads the server's
  markup; no assertion covers `applyI18n` actually resolving `data-i18n-title` in a live browser.
- **The join is not observed against live production state.** All fixtures are `t.TempDir()`
  synthetic; the §A.4 live measurement was not re-taken.

### RED evidence

- M1: `undefined: LaneVM` / `undefined: loadFactoryLanes` (10 sites), `FAIL … [build failed]`.
- M2: `too many arguments in call to buildChain` at 5 sites, `FAIL … [build failed]`.
- M3: `lane row missing "data-lane=\"2\""`, `factory section absent from the markup`,
  `screens.templ still carries the falsified string "are not recorded yet"`, and 25 missing-locale-key
  failures — all before the implementation.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-24
run_base_sha: 40f064c6d
run_commit_sha: f80bad4d0
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 3   # events.go + app.js transport, six nav rows, @missing() role — all preserved
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: pass
  windows_amd64: pass
total_run_phase_files: 8          # 3 new tests, 1 new source, 2 templates + their generated twins, i18n.js, viewmodel_ops.go
m1_to_mN_commit_strategy: one commit per milestone (b35a3d098, e78080458, f80bad4d0); not pushed
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-24
sync_commit_sha: f9e267e80
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-WEB-CONSOLE-015' CHANGELOG.md → 0 (rc 1, a genuine zero-match; no duplicate, emission permitted)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+[a-z]?' acceptance.md | sort -u → 15 tokens, of which AC-WC15-012 occurs once and only as the record of its own deletion (acceptance.md:7). 14 live criteria, matching the §E.2 matrix and the declared budget; the CHANGELOG entry states 14"
b12_self_test_c: "every path named in the CHANGELOG entry verified present with ls before commit (14 paths, all present)"
changelog_entry_position: "[Unreleased] → ### Changed, appended immediately after the SPEC-KANBAN-RECORD-SESSION-KEY-001 entry — the fourth and last of the four siblings synced on this branch"
frontmatter_status_transitions:
  spec.md: "in-progress → implemented → completed (updated: already 2026-08-24, the sync date — unchanged)"
  plan.md: "n/a — carries no frontmatter block"
  acceptance.md: "n/a — carries no frontmatter block"
  progress.md: "n/a — carries no frontmatter block"
  design.md: "not touched — retained reference material outside the Tier M artifact set"
  research.md: "not touched — same; it already carries the note that spec.md supersedes it where they disagree"
canary_compliance_check: "n/a — this SPEC defines no forward-looking policy that its own sync tests"
```

### Gaps carried forward from run-phase — still NOT observed at sync

The run-phase record (§E.2 § Gaps) framed these correctly; they are carried in that framing rather
than softened. None of the three was closed by sync, which touches markdown only.

- **Coverage regression vs the pre-change tree was NOT measured.** The post-change rate is 66.8%
  of statements for `internal/web`; the pre-change rate was never taken. What was measured instead,
  at `f80bad4d0`: every function this SPEC adds sits well above the package rate —
  `loadFactoryLanes` 100%, `resolveLane` 95.8%, `unresolvedLane` 100%, `readTelemetry` 85.7%,
  `telemetryCells` 100%, `buildChain` 94.7%. **Why the substitution**: measuring the pre-change
  rate requires a second worktree checkout of `40f064c6d`, and the lead judged that cost above the
  benefit here and did not require it. The per-function figures bound the regression; they are not
  the measurement the Definition of Done asks for. This is a **gap, not a pass**.
- **The browser-side render is NOT observed.** Every render assertion reads the server's markup;
  nothing covers `applyI18n` resolving `data-i18n-title` in a live browser. The `mark.notRecorded`
  hover text and the four-locale key resolution are verified as data and as markup, never as
  rendered DOM.
- **The join is NOT observed against live production state.** Every fixture is a `t.TempDir()`
  synthetic tree; `spec.md` §A.4's live measurement — the lane whose join completes and returns
  another lane's record — was not re-taken after the change.

### Documentation review — what was searched, found and changed

**Changed: four files, one per locale.** Two sentences on the console guide page were falsified by
this change and were corrected; one paragraph was added because the page's inventory of what the
Kanban screen shows was now incomplete.

Searched — `docs-site/content/{en,ko,ja,zh}/` and the four READMEs — for any description of the
Kanban screen's cells or of the note banner: `kanban`, `not recorded yet`, `left blank`,
`Unrecorded values`, `kanban.Record`, and the four-locale equivalents (`기록되지`, `아직 기록`,
`記録されていない`, `未被记录`), plus `model` / `effort` / `context` / `lane` on the two pages that
describe the console (`advanced/moai-web-console.md`, `cli-reference/web.md`).

| Where | Finding | Action |
|---|---|---|
| `advanced/moai-web-console.md` § Kanban, chain-board paragraph (all 4 locales) | "model, effort and context are not recorded yet, so they are left blank" — falsified: the cells now carry the session's telemetry values | corrected; a blank cell now stated as "this session has no record carrying that value yet" |
| same page, § "Never write down what it does not know", the *Unrecorded values* bullet (all 4 locales) | same falsified premise, restated as a principle | corrected to "values with no record", extended with the unresolved-lane rule |
| same page, § Kanban (all 4 locales) | the page listed the chain board and the SPEC pipeline; the factory-lane list was absent | one paragraph added per locale, naming the row's fields, the unresolved marker and its reason, and the no-lane case |
| `cli-reference/web.md:50` route table, "Chain session board plus the SPEC pipeline" | a one-line route summary, not an enumeration — not falsified by an added section | unchanged |
| README (all 4), "The Kanban screen shows the kanban chain alongside the SPEC pipeline" | same: a summary that names no cell and no banner | unchanged |

The four locale files were edited in the same change and by the same substitution, so the
section-count parity the docs-site i18n rule requires is preserved (each file gained exactly one
paragraph, at the same position). No template mirror exists for `docs-site/`, so `make build` was
not run and no `catalog.yaml` regeneration applies.

No page anywhere described the falsified strings themselves (`grep -rn 'not recorded yet|left
blank|Unrecorded values'` over `docs-site/content` and the four READMEs returned only the two
English lines corrected above; their three locale twins were found through the per-locale phrase
search). Writing a page this SPEC did not require would be a worse outcome than the silence it
would fill, so nothing else was added.
