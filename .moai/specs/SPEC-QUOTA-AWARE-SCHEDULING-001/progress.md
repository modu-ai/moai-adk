# Progress — SPEC-QUOTA-AWARE-SCHEDULING-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L), plus decision-index.md
- plan_complete_at: 2026-10-02
- open_decisions: none — DO-1..DO-12 resolved (oracle; leader verdict on DO-3, DO-7, DO-8; DO-12 `write_backend_at_claim` verified then applied at spec 0.4.0)
- plan_audit_verdict: iteration 2 PASS 0.87 (Tier L threshold 0.85; `.moai/reports/t1347/plan-audit-iter2.md`, audited_sha 7fe1ee49d, local-only). Iteration 1 was FAIL 0.74 (`plan-audit-iter1.md`); the revision at spec 0.5.0 closed D1-D24. Ten MINOR findings (N1-N10) are carried as known plan debt in section F below.
- open_decisions_note: DO-13 (first-exhausted time surfaced in the status block) is PROVISIONAL, escalated to the leader

## §E.2 Run-phase Evidence

_pending run-phase_

### M0 — Baseline goldens (REQ-QAS-016, AC-QAS-015; card t1347)

Commit order witness: goldens commit `c2ae5236af96e8e76fb5c27639c22380bdec5dfa` (parent `3dc73dc103b51d46f1cab6faf43f33e692075a95`) touches only the three golden files; this evidence commit follows it. Measured on tree `bad02c7c047cf9e5a99e069dfa505589c2971a95` (HEAD `3dc73dc10`), branch `WT-quota-aware-scheduling`, toolchain `go version go1.26.8 darwin/arm64`. `moai version` printed `v3.2.0-rc.24 ... gc50da9c2f` (judging build for the slot verbs only; no golden was produced by the installed binary — each was produced by the `go test` toolchain from this tree).

**Pre-change proof (measurement integrity).** `git diff --name-only c50da9c2f..HEAD -- internal` at HEAD `3dc73dc10`, before any generator ran: printed nothing (empty = `internal/` code byte-identical to `c50da9c2f`, the tree the installed build names). `grep -rln "TestQAS_" internal` printed nothing.

**Step 0 — RED-selector record (every anchored selector prints the empty-sweep token on the unmodified tree).** Each row is one invocation of the form `<command>; echo "rc=$?"` with `go -C <card tree>` (the shell cwd was the SPEC directory); `-v -count=1`.

| Package | Selector (anchored, `-run`) | Verbatim empty-sweep output | rc |
|---|---|---|---|
| `./internal/statusline` | `^(TestQAS_AC001_RecordCarriesSuppliedWindowsOnly\|TestQAS_AC002_PreviousSchemaReadsAsNoWindows\|TestQAS_AC002b_WindowlessRecordBytesMatchBaseline\|TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop\|TestQAS_AC004_ExhaustedAtStickyUntilRollover\|TestQAS_AC005_AggregateFreshestBoundariesSkewAndRollover\|TestQAS_AC006_FailOpenOnAbsentOrUnreadable\|TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly)$` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/statusline	0.380s [no tests to run]` | 0 |
| `./internal/config` | `^TestQAS_AC007_ConfigDefaultsMirrorTemplate$` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/config	0.338s [no tests to run]` | 0 |
| `./internal/cli` (slot `go-test-cli` held) | `^(TestQAS_AC006b_NextLeasesWhenQuotaDataAbsent\|TestQAS_AC008_ClaudeLaneHeldAtThreshold\|TestQAS_AC008b_MCPFactoryNextHeld\|TestQAS_AC009_HoldLineCarriesResetTime\|TestQAS_AC010_WaitLatchReleasesOnlyBelowMarginOrResetOrUnknown\|TestQAS_AC011_NonClaudeBackendsNeverHeld\|TestQAS_AC011b_StageAndCompleteIgnoreQuotaHold\|TestQAS_AC012_StatusQuotaBlockOnlyWhenGateEnabledWithData\|TestQAS_AC013_AcquireWarnsNeverRefuses\|TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly\|TestQAS_AC017_SharedPressureEvaluationAndSurfaces\|TestQAS_AC018_LaneInventoryCandidates\|TestQAS_AC019_AutoAndStatusRecommendNonClaudeLanes\|TestQAS_AC020_NoNonClaudeLaneWarnsOnly\|TestQAS_AC021_SteeringChangesNothing\|TestQAS_AC022_GateOffOrPressureOffOutputUnchanged\|TestQAS_AC023b_LaunchersPassTheirBackendToTheClaim)$` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	1.223s [no tests to run]` | 0 |
| `./internal/kanban` (slot held) | `^TestQAS_AC023_ClaimRecordsBackend$` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/kanban	0.406s [no tests to run]` | 0 |

Classification: all four are the intended RED of a test that does not exist yet (`[no tests to run]` with rc 0 is NOT a pass: verification-completeness §1.1; the swept count is zero). Selector coverage: AC-001..-014, -017..-023 (21 Go deciders; AC-QAS-014 is named in both `./internal/statusline` and `./internal/cli`; AC-QAS-015 and -016 are decided by git commands, not Go tests). The first statusline attempt without `-C` failed `[setup failed]` (cwd was the SPEC directory); the row above is the re-run. Slot lease: `moai slot acquire --resource go-test-cli --max-duration 20m` printed `slot go-test-cli acquired by 2da35a68-1196-4183-b6e5-a50fc9b6d901 until 2026-10-02T07:22:41Z` rc=0; `moai slot release --resource go-test-cli` printed `slot go-test-cli released (was 2da35a68-1196-4183-b6e5-a50fc9b6d901)` rc=0 (released before this report).

**Goldens (measured by a temporary generator test `TestZZQASBaselineGen` per package, deleted before staging; two consecutive runs produced byte-identical files).** `internal/cli` generator run as `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_FACTORY_WORKERS MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH MOAI_LAUNCH_PROVIDER && go -C <card tree> test -v -run '^TestZZQASBaselineGen$' -count=1 ./internal/cli` (the first attempt under the lane environment failed: `todo add: moai add: refused — lane boundary`, which is the lane environment falsifying the fixture, not a golden); the statusline generator ran WITHOUT a scrub (not measured whether any lane variable could reach it; the golden bytes carry no env-derived value, and the file was byte-identical across two runs).

`shasum -a 256` (run after the second generator run, on the committed files):

```
46db7e4a8922c16a2566d2b23a91ceac7dcb7f212d635773bb7f8fce153f3726  internal/statusline/testdata/qas_baseline_windowless_record.golden.json
d52210a78575b65cbc2a7bfc929cf394610eb31ae65eb357c44b11fcedb7ce09  internal/cli/testdata/qas_baseline_factory_status.golden.json
570dc05140007a68d165c4b077c5496cf5c3d03876fcc790b9301ddbc7f7bc65  internal/cli/testdata/qas_baseline_todo_auto.golden.txt
```

1. `qas_baseline_windowless_record.golden.json` (AC-QAS-002b). The fixed input is `writeContextUsage(<TempDir>, "qas-baseline-session", 4242, MemoryData{Available: true, ContextWindowSize: 200000, TokensUsed: 50000}, handoffStageNone, "Opus 4.8", "high")`; the bytes are the on-disk file the writer produced (MarshalIndent, two-space indent, no trailing newline), with exactly one edit: the `captured_at` value was replaced by the constant `2026-10-02T00:00:00Z` (the writer stamps `time.Now()`, so the value cannot be a fixed byte string otherwise). `writer_pid` is the fixed argument 4242; `schema_version` is the pre-change `2`. Per AC-QAS-002b the later test normalizes only `captured_at`, `writer_pid`, and `schema_version` before comparing, so those three fields carry no identity here. Bytes: keys in order `schema_version, session_id, writer_pid, captured_at, context_window_size, tokens_used, raw_pct, stage, band, model, effort`; `raw_pct` is `25`, `stage` `none`, `band` `standard`.
2. `qas_baseline_factory_status.golden.json` (AC-QAS-012, -022). One file keyed `empty` and `one_card`, each value the exact `moai factory status --run run-cli --json` output (cobra `newFactoryCommand`, `runFactory` helper) re-indented inside the wrapper. Extraction rule (self-checked in the generator, no mismatch): `json.Unmarshal` the file into `map[string]json.RawMessage`, `json.Indent(&buf, raw, "", "  ")`, append one `\n`; the result equals the command's stdout byte for byte. Fixtures: `empty` = `fcFixture(t)` only (no factory database: output `{"run":"run-cli","cards":[],"unavailable":[]}`); `one_card` = `fcFixture`, `fcQueue(store, queued)`, `fcClassify("t1", high, false, serial)`, `sdRegisterLane("lane-1")`, `fcPlace(Card{CardID:"t1", State:leased, LeaseHolder:"lane-1", LeaseExpiresAt:"2026-09-26T10:00:00Z", Stage:run})` under the fixed clock `fcNow` = `2026-09-26T09:00:00Z`. No volatile field exists in this output (no temp path, no wall-clock value), so nothing is normalized. The `one_card` fixture differs from `TestFactoryStatusShowsHolderModePriority` in one respect: one card instead of two, and `LeaseExpiresAt` set so `lease_expired` is `false`.
3. `qas_baseline_todo_auto.golden.txt` (AC-QAS-022). `runAutoCycle` over two queued cards `t1` "card a" and `t2` "card b" (added in that order through `store.Add`, not through a map), `autoOptions{wait: 5m, sessionID: "operator-session-fixture", jev: stub returning "jev: stubbed (qas baseline fixture)", liveness: autoTestLiveness(root, "t1", true, true, nil), sleep: writes <root>/.moai/reports/<id>/evidence.md for card tick, now: unix 0 + tick minutes}` and nil `landed` / `jevRank` (so the ranking stage prints `selection: source=fallback reason=jev-disabled`, `ranked t1 t2`, and the landed-signal non-finding). Both cards complete (`done t1`, `done t2`, two `/clear guidance` blocks). The only normalization: every occurrence of the per-run `t.TempDir()` root in the evidence path is rendered as the literal `<ROOT>`; the later test applies the same `strings.ReplaceAll(out, root, "<ROOT>")`. 28 lines.

**AC-QAS-015 baseline fact (the commit-graph part).** Each B_i is the single commit `c2ae5236af96e8e76fb5c27639c22380bdec5dfa` (all three goldens in one commit, so B_1 = B_2 = B_3); it touches no file under `internal/` that is neither a `_test.go` nor under `testdata/`, so every future implementation commit I follows it. The decider evaluation (`git log --reverse ... -- internal ':(exclude)*_test.go' ':(exclude)*/testdata/*'` and the `git rev-list --count B_i..I` / `I..B_i` pair) is deferred to the close of the run phase: I does not exist yet.

**Gaps.** (a) AC-QAS-015's ancestor test cannot be evaluated until the first implementation commit exists; the recorded fact is only that B_i precedes it by construction. (b) The generator source is not committed (it was a throwaway); the fixtures are reproducible from the recipe above and from the cited existing helpers, and a regenerated file must match the shasum above. (c) The status and `--auto` goldens come from the cobra command and `runAutoCycle` run through the test harness, not from the installed `moai` binary (which would read the real home and queue). (d) `captured_at` in golden 1 is a substituted constant, not a measured value. (e) A stderr line `WARN config sections directory not found, using defaults` appeared during the status generator; it is a log line, not part of any golden. (f) Lane environment: the `internal/cli` generator result under the lane environment was a failure (see above); only the scrubbed run produced goldens. (g) No `-race`, no `go test ./...`, no lint or Windows build was run: no Go source of this repository changed, so neither applies.

**Residual risk.** A later test that rebuilds a fixture even slightly differently (a different lane label, lease time, card text, or root normalization token) will mismatch the golden for a reason that is not a regression; the recipe above is the single source for the fixtures. The `one_card` status fixture was chosen by this milestone; if a later AC needs a card with a decision or contract field populated, it needs its own golden rather than an edit of this one (an edit would not be pre-change).

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection

### Kickoff gate (plan→run) — autonomous form, `.claude/rules/moai/workflow/auto-semantics.md` §9.1

| Condition (§9.1) | Observed | Evidence |
|---|---|---|
| Independent plan-audit verdict is PASS | PASS, 0.87 against the Tier L threshold 0.85 (iteration 1 was FAIL 0.74; the score rose, no regression stop) | `.moai/reports/t1347/plan-audit-iter2.md` (`verdict: PASS`, `audited_sha: 7fe1ee49d62e67e86ae45e139187a679d67873ca`, read from the file by the lane) |
| Plan phase records audit-ready | `plan_status: audit-ready` | §E.1 above |
| Plan-artifact hashes unchanged since the verdict | equal | `shasum -a 256` of the five plan artifacts re-measured by the lane after the verdict; all five equal the hashes in the iteration 2 report (spec ec919773, plan 6050abc6, acceptance 803e5e69, design bf4d616e, research 99f141e1); `git diff --stat 7fe1ee49d..HEAD` empty; progress.md is not a hash subject |
| No blocker open | none | no BLOCKER or MAJOR finding in iteration 2; DO-1..DO-13 resolved (DO-3 and DO-13 leader-confirmed provisional) |
| Keep-set case (environment-impossible / operator-held / irreversible external-shared) | none applies | nothing is pushed, no PR, no external shared system touched; the operator-held items (push, release) stay with the leader |

```text
decision record: decided_by=claude-code lane-9 orchestrator (factory lane, Kickoff autonomous transition) evidence_refs=.moai/reports/t1347/plan-audit-iter2.md(verdict=PASS score=0.87 audited_sha=7fe1ee49d),.moai/specs/SPEC-QUOTA-AWARE-SCHEDULING-001/progress.md#E.1,commit 7fe1ee49d,sha256 spec=ec919773 plan=6050abc6 acceptance=803e5e69 ladder_path=gate-row plan→run Kickoff (AUTONOMOUS, auto-semantics §9.1)
```

Recorded 2026-10-02T06:59:41Z. The decision board under the moai home (auto-semantics §11) was NOT written: `moai factory decide` records only a human decider (its help text: `--decider ... (F1 accepts only human)`), and this card was dispatched directly by the leader without a factory record. This record lives here and in the card's local evidence files; a reader must treat it as self-attested (auto-semantics §10).

Whether to spend a third and final audit on the ten minor findings was put to the decision oracle (Jev): `proceed_record_debt` (probability 0.86, confidence 0.73). Proceeding keeps the audited hash valid; any edit to a plan artifact would void the verdict and the margin over the threshold is 0.02.

### Known plan debt carried into the run delegation (iteration 2 findings N1-N10, all MINOR)

The implementer may not edit the SPEC body; each item is handed over with the rule that a blocker report goes to the lane if one blocks a milestone.

| ID | Debt | Treatment in run |
|----|------|------------------|
| N1 | AC-014 minimum swept count (3) cannot pass per package before M5, while plan M2/M3 list it | AC-014's count assertion is evaluated at M5 (its own green-path sentence says so); M2/M3 run only the per-file part |
| N2 | plan.md:102 and research.md:19 say "integer-rounded"; REQ-003, design 2 and AC-003 say truncated (`int()`, context_usage.go:272) | follow the REQ/AC: truncate |
| N3 | plan.md:114 reads as if the pressure function takes a caller argument; REQ-017 and AC-017 say it takes none | follow REQ-017: the shared function takes no caller input, the lane gate applies the Claude-caller predicate to its result; AC-017 varies the environment, not an argument |
| N4 | E9's CLI half is a pipeline; AC-022's RED row measures a different directory | add a conforming single-invocation RED row in the M0 evidence if cheap, else record as a gap |
| N5 | AC-005 fixtures are one-directional (a "minimum over fresh" mutant survives); the text-mode status baseline is missing (M0 captures `--json` only) | add a fixture where the newer record is the higher one; capture a text-mode golden in M0 if the SPEC permits, else report the gap |
| N6 | REQ-005 lists only percentage and reset time; REQ-013 also needs capture time and exhausted time; REQ-004 reset-while-at-100% case unspecified | implement REQ-013's fields; on a changed reset time re-observe the exhausted time from the current capture when the window is still at or above 100% |
| N7 | short-form IDs in prose produce 14 ORPHAN lines in the traceability verb (not a coverage gap) | none (doc-only) |
| N8 | AC-018 does not say the fixture registry lacks the `legacy_workers_imported` meta row; release rule for two held windows is unstated | fixture inserts lane rows with plain SQL into a database lacking the marker; treat release as "all held windows released" |
| N9 | the `SaveFactoryRegistry` round-trip of `backend` has no AC | add one subtest alongside AC-023 |
| N10 | AC-016's hunk check reads only the hunk's old-side start line | in the implementation of the check also require the hunk end inside the allowed range |

Also carried from the iteration 2 gaps: the pinned `modernc.org/sqlite v1.57.0` was never exercised; a system `sqlite3` 3.54.0 `-readonly` open of a no-sidecar WAL database failed with `unable to open database file (14)`. At M5 start the implementer measures the real driver with a scratch program outside the tree (a cleanly closed WAL database, `mode=ro`, read one row, attempt a write, record the directory listing before and after) and reports the observed behaviour; if the quiescent case fails, the inventory returns no candidates (fail-open) and the finding goes to the lane.

### Mode evaluation

Input parameters: tier L (REQ 23, AC 23); files affected more than 15 across three packages (`internal/statusline`, `internal/cli`, `internal/kanban`, plus `internal/config` and template mirrors); domains: Go source, tests and goldens, config defaults, template mirror, one rule-adjacent doc line; language mix Go + YAML; concurrency benefit LOW — milestones M0-M6 are ordered (M0 commits the baseline goldens before any implementation commit, M1-M3 build the record, aggregator and gate, M4 the claim write, M5 the steering surfaces that read it). Agent Teams: not requested.

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | non-trivial, spans code, tests, config and template |
| serial | **yes** | coding-heavy, ordered milestones, one writer per tree; the default fallback |
| fanout | no | research is finished; the remaining work is implementation |
| sweep | no | not a uniform mechanical transform |

Decision: serial

Boundary case: the Tier L entry predicate for `manager-lead` (at least 3 milestones and at least 10 files) is met, but a lane session holds standing spawn authority at depth 1 only and agents it spawns are leaf workers that must not spawn further agents, so `manager-lead` (the Agent-carrying coordinator) cannot be used from this lane. The lane drives the milestones itself, one `manager-develop` leaf per milestone, in order; the lane does not write code.

## §J Lane Decision Log (card t1347, lane-9)

Rule (leader dispatch, operator instruction): in-flight decisions are asked of Jev (TypeSafe System One, `jev-1.13.0`) through the lane's dispatch script; question, answer and confidence are recorded here. Confidence < 0.5 or a hard-to-reverse external action goes to the leader. The oracle's answer is advisory evidence; where a fact it was given could be checked in code, the lane checked it first (DO-5: exit status 3 is `factoryNextNoCardExit`, `internal/cli/factory_card.go:181` (comment at `:179-180`), "3, so a supervising launcher can distinguish it from failure").

| ID | Question (short) | Oracle answer | Confidence | Disposition |
|----|------------------|---------------|-----------:|-------------|
| DO-1 | Stamp the hook-observed 429 turn-end into the record? | not_now | 1.00 | applied |
| DO-2 | Distinguish two Claude accounts? | reset_time_only | 0.97 | applied |
| DO-3 | Default thresholds (5h / 7d / margin / max age) | 90 / 95 / 5 / 30m (P 0.54) | 0.32 LOW | leader: accepted as provisional; all four exposed as config keys; "unmeasured defaults"; re-tune after first measurement is a listed follow-up |
| DO-4 | Template default for `workflow.quota_gate.enabled` | false in template, true in local config | 0.99 | applied |
| DO-5 | Exit status for the hold | reuse status 3 (P 0.90) | 0.81 | applied (SPEC default was 4; changed) |
| DO-6 | Gate arm (a), a card already assigned to the lane? | exempt started cards | 0.99 | applied (SPEC default was gate-all; changed) |
| DO-7 | `integration acquire` under pressure | warn_only (P 0.54 vs 0.46) | 0.07 LOW | leader: warn-only, never blocks, FINAL |
| DO-8 | Steer `--auto` / leader beyond a status block? | steer_auto (P 0.66) | 0.32 LOW | leader: REJECTED the lane's status-only provisional; steering IN SCOPE in its minimal form (recommend non-Claude lanes, Claude lanes hold with status 3, warning-only when no non-Claude lane, no forced re-dispatch); SPEC reclassified Tier M to Tier L |
| DO-9 | Stale high reading before reset | treat_unknown (P 0.93) | 0.86 | applied |
| DO-10 | Claude-lane predicate source | launch provider first, kanban backend fallback | 0.96 | applied |
| DO-11 | Record carrier | per-session record | 1.00 | applied |
| DO-12 | Where to read a lane's backend | write_backend_at_claim (P 0.78) | 0.55 | applied after the lane required feasibility verification (U18); verified by manager-spec and re-checked by the lane: `workers.backend TEXT NOT NULL DEFAULT ''` at `internal/homestate/factory.go:27`; claim inserts at `internal/kanban/factory_slots.go:106` and `:329` omit it |

Leader messages: plan commit and the three low-confidence items reported to the leader; verdict received for DO-3 / DO-7 / DO-8 (above). No hard-to-reverse external action has been taken (no push, PR or delete).

Plan artifacts: `400b5f986` (initial), `fdfb2d0c8` (oracle resolutions + leader verdict, Tier L), `ca57dc350` (DO-12). Tier L: REQ 23 of 25, AC 23 of 25.

### Plan-audit iteration 1 (FAIL 0.74) — decisions taken while closing D1-D24

Report `.moai/reports/t1347/plan-audit-iter1.md` is local-only by operator directive (`.gitignore` `.moai/reports/*`, `.moai/docs/audit-artifact-convention.md` Committing section); it is not committed. The leader asked for a commit, the lane declined on that directive, and the leader withdrew the request after reading both sources and recomputing the sha256.

| ID | Question (short) | Oracle answer | Confidence | Disposition |
|----|------------------|---------------|-----------:|-------------|
| D1 | `factory status` quota block vs byte-identical output when pressure is off | block only when the gate is enabled (P 0.94) | 0.88 | applied (REQ-013, REQ-022) |
| D3 | A held wait whose reading ages out mid-wait | release when unknown (P 0.80) | 0.59 | applied (REQ-011, AC-010 `unknown_mid_wait_releases`) |
| D7 | Lane inventory read of the registry | genuinely read-only open (P 0.82) | 0.63 | applied after feasibility was measured by manager-spec on `modernc.org/sqlite v1.57.0` (plan.md D.1: absent file errors without creating anything; `mode=ro` reads a closed WAL database and sees live WAL rows; `immutable=1` rejected because it missed uncheckpointed rows); the lane re-read the existing `mode=ro` pattern the SPEC reuses, `openSQLiteReadOnly` at `internal/discovery/factory_discovery.go:353-362` (fail-open: a WAL database whose recovery needs the write lock contributes no candidates). The lane did not re-run the scratch measurements |
| D12 | Surface the first-exhausted time (otherwise dead data) | surface (P 0.52 vs 0.48) | 0.03 LOW | leader: CONFIRMED surface in the status block (DO-13) |

Revision commits: `0e12b4cfa` (D1-D24 closed, spec 0.5.0), `be14684ab` (drops a suffixed AC identifier that made the commit guard count 24). Lane re-check: tree clean, `moai spec lint --strict` no findings, REQ 23 and AC 23 by grep.
