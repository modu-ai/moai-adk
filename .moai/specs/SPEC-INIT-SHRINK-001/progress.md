# progress.md — SPEC-INIT-SHRINK-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-03T21:51+09:00 (plan-audit iteration 2: PASS-WITH-DEBT 0.925 ≥ Tier L 0.85, .moai/reports/t1438/plan-audit-iter2.md, audited SHA 49557f87c, artifact hash 6b4c9498acb231857779abb209fa48cbbbdb4cba94927b185965c823881d0420)
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md, decision-index.md (Tier L: 5 artifact set + progress + decision index)
req_ac_count: 21 / 21 (Tier L ceilings 25 / 25)
open_decisions: OD-1..OD-8 — all eight SETTLED 2026-10-03 (decision-index.md Operator verdict rows; leader ruling, codex-informed, relayed via lane; plan-audit-iter2 (e) verified all rows carry verdict letters and provenance)
tree_pin: WT-moai-init-slim @ 3f3ebb763 (authoring)
notes: RED-now ledger cells L-01..L-24 measured at the pin tree; the plan-auditor loop records its verdicts under .moai/reports/t1438/ and updates plan_status here.
repair note: the iter2 verdict above was overturned to FAIL 0.90 by the codex cross-model receipt (rcpt-c5721bde339669a03956a37e; D-7 removal executor vs preservation, D-8 dangling-mirror fallback, D-9 unobservable install outcome — all source-verified) — the 2026-10-03 repair commit on WT-moai-init-slim (classified-set removal executor, re-homed mirror fallback, post-install list-surface probe; AC-006/011/013/015 cells re-observed at a1f17b038) is the iteration-3 fix.
iter3: FAIL 0.85 (.moai/reports/t1438/plan-audit-iter3.md, receipt rcpt-10fb620edbc90ece3943e22f) — D-7/D-8/D-9 verified RESOLVED; new blocking findings D-10 (this file's kickoff record cited the invalidated iter2 verdict — fixed in this commit), D-11 (probe cannot separate this-install-success from pre-existing plugin; failed-or-skipped arm names an unobservable trigger), D-12 (REQ-010 foreign definition classifies moai-custom as managed, contradicting REQ-013/AC-013), D-13 (design §2.4 arm-mapping mixes migration/init surfaces unscoped vs REQ-001/009). Retry cap 3/3 reached — leader escalation in flight (2026-10-04).
iter4: FAIL 0.85 (.moai/reports/t1438/plan-audit-iter4.md, receipt rcpt-755bbbcc6218e0306cb7de90, leader-authorized out-of-cap audit) — D-10/D-11/D-12/D-13 verified RESOLVED; the deferral rule did NOT trigger (no critical user-file-loss defect); new blocking findings D-14..D-17 are docs-scale cross-layer sweep residues of the same repairs, enumerated to specific lines (REQ-004 trigger wording; 4 unqualified absent/stale→modified surfaces; design step-1 vs step-4 contradiction on the not-demonstrated arm + an over-broad byte-for-byte promise; whole-directory archiveSkill vs per-file classification). Leader decision pending: docs-scale repair + one more leader-authorized delta audit, or debt acceptance (2026-10-04).
disposition: the leader accepted the debt — PASS-WITH-DEBT over the iter-4 FAIL verdict (verbatim preserved in the verdict file), 2026-10-04, with five binding run-phase conditions (D-14..17 line repair as the first obligation before manager-develop; D-16 promise narrowed to the REQ-017 scope; D-17 archive unit aligned to per-file classification; spec lint 0 after the repair; D-15/D-16 loss paths pinned as RED tests in implementation). Recorded in §F and on the verdict file.
iter5: PASS-WITH-DEBT 0.93 (.moai/reports/t1438/plan-audit-iter5.md, receipt rcpt-91c53431e86cd49b8a2019ab, leader-authorized minimal delta over the 68d9f2a9b repair) — the auditor's own closing verdict line cites the receipt; all five scope items verified resolved; both flagged judgment calls judged coherent; new non-blocking findings D-18 (archive-ledger note claims AC-012 sharpening the commit leaves to the body — one-line note fix) and D-19 (one overstatement clause, leader discretion) recorded as optional debt. Run-phase entry proceeds on this verdict.
repair-4: leader-ruling repair executed 2026-10-04 (mission contract 11c79e1a, manager-spec) — D-11/D-12/D-13 repaired across spec.md/plan.md/acceptance.md/design.md: D-12 foreign class keyed to template carriage (moai-custom and user-created skills foreign, preserved byte-for-byte; REQ-013/AC-013/AC-010/Edge Cases aligned); D-11 probe contract = pre-execution snapshot vs post-execution diff, residual arm renamed `not-demonstrated`, causal trigger wording removed (REQ-005/REQ-015/AC-005/AC-015, design §2.1/§2.3/§2.4/§3, plan M2/M3); D-13 arm mapping scoped per surface (migration vs init; REQ-001/REQ-009 contradiction eliminated). AC-005/AC-015 cells re-observed at tree `5c380a251` (L-06, L-16); decision-index.md untouched. Iteration-4 delta audit authorized by the leader beyond the 3-iteration retry cap — pending.

## §E.2 Run-phase Evidence

manager-develop, dispatched as a general-type worker in the card worktree (branch `WT-moai-init-slim`, base HEAD `0a1e105d8`). Pre-flight measured 2026-10-04: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m` on internal/cli, internal/template, internal/config → `0 issues.` (the lint baseline; NEW-vs-baseline comparisons below cite it); the consumed t1435 surfaces verified by landed name — `internal/cli/plugin_install.go` present, the `--no-plugin` flag wired (`plugin_install.go:183`), `SKIP_PLUGIN` referenced from `internal/config/envkeys.go`, and the doctor's plugin-list read pattern at `internal/cli/doctor_plugin_version.go:197-237`. The anchor probes of the Evidence Ledger re-run at this tree before any test was written: every AC-named selector printed the `[no tests to run]` shape (L-01/L-03/L-07/L-10/L-11/L-14 forms), exit 0 — red by the PASS-line rule.

### M1 — mode record, migration classification, resolution gate (commit `d8e4300e4`)

| Claim | Evidence (command → deciding output, this run, this tree) |
|---|---|
| AC-009 (record + round-trip) | `go test ./internal/cli -run '^TestDeployModeRecordRoundTrip$' -count=1 -v` → `--- PASS: TestDeployModeRecordRoundTrip (0.01s)`; both closed-set values, re-init rewrite, and the update-cycle survival (Backup → `.moai/config` wipe → template redeploy → restore re-assert → re-read `plugin`) |
| AC-010 (classification) | `go test ./internal/cli -run '^TestMigrationClassification$' -count=1 -v` → `--- PASS: TestMigrationClassification (0.00s)`; identical / modified / foreign / absent-record→modified against the real embedded render; counts 1/2/1 |
| AC-013 (M1 half) | `go test ./internal/cli -run '^TestMigrationLeavesForeignFilesUntouched$' -count=1 -v` → `--- PASS: TestMigrationLeavesForeignFilesUntouched (0.00s)`; `moai-custom` (managed-glob hit, no render carriage) classifies foreign, bytes unchanged, symlink never classified |
| D-15 loss path (leader condition 5a) | `TestClassifyMigrationGlobHitForeignIsNeverRemovable` (internal/cli/update) — RED at `0a1e105d8`: build failure naming `undefined: ClassifyMigration` (the machinery did not exist); GREEN at M1: glob-hit foreign names land foreign under BOTH a missing manifest and a template_managed record — never in a removal/archive class. The executor side that consumes only the classified sets is pinned by M3's `TestMigrationRemovesIdenticalDroppedComponents` |
| AC-008 (b) harness | `go test ./internal/cli -run '^TestResolutionGateHarness$' -count=1 -v` → `--- PASS: TestResolutionGateHarness (3.96s)`; the self-test prints `RESULT pass=5 fail=0` incl. `PASS isolation-scrub-negative-control` (a planted unscrubbed variable FAILs its case — the scrub is load-bearing) |
| AC-020 (a) M1 half | `go test ./internal/cli -run '^TestShrinkVerificationNeverReachesRealHome$' -count=1 -v` → `--- PASS`; no script assigns HOME (word-start guard; `CODEX_HOME=` scratch writes excluded), the scrub is live-enumerated over all three variable families, and the default runner refusal is shown live (silent nil under a test binary) |
| Archive layout (M1 deliverable) | `TestMigrationArchiveLayout` — the migration tag is distinct from the legacy `v2.16`; skill-file root preserves the skill-directory layout; standalone-file variant carries the original path |
| Unit table | `go test ./internal/cli/update -run 'TestClassify|TestMigrationRoots|TestMigrationArchiveLayout' -count=1` → ok; config reader/writer tables → ok (internal/config, internal/template) |

### REQ-008 measurement — the recorded verdict (the M2-flip gate)

Final recorded run (2026-10-04, quiet machine, fixture `/tmp/t1438-resolution-fixture`):
`sh scripts/check-bare-name-resolution.sh /tmp/t1438-resolution-fixture` → `RESULT pass=5 fail=1`
(PASS: isolation-scrub, fixture-present, claude-bare-skill, codex-naming, hermeticity-protected-set;
FAIL: claude-command-body marker). Verdict routing:

- **Q1 (bare skill name resolves): YES.** The session loaded the fixture skill from its bare name and
  wrote the marker (`resolved`). Observed PASS in this run and in two earlier runs of the same harness.
- **Q2 (a plugin command body's bare `Skill("moai")` resolves): YES.** The failing case's own session log
  (`claude-command.log`) states: "The `moai` skill loaded and its instruction is to create a file at
  `$MOAI_RESOLUTION_MARKER2`" — the bare name resolved from inside the plugin command body; only the
  marker-write step failed because that session shape carried no file tools (the `--allowedTools` grant
  did not apply through the SDK session shape on this machine). The resolution question the gate exists
  to answer is affirmative; the marker mechanism is recorded as harness debt.
- **Q3 (Codex lists plugin-borne components): YES — as namespaced entries.** The render scan lists
  `resolution-probe:moai-resolution-probe`, `resolution-probe:moai`, and the migrated command skill
  `resolution-probe:source-command-resolution-probe` (R03-codex confirmed at the runtime level).
- **Hermeticity:** the before/after protected-set hash is equal (ambient subtrees pruned per the t1434
  LEAK-FINDING's directory-entry rule); isolation cases incl. the negative control pass.

OD-2 routing (settled (a) + condition): bare names RESOLVE in plugin mode — the mode-aware reference
rewrite ships the bare names the test proves resolvable; no scaffold instruction rewrite is forced by
resolution. OD-6 routing: Codex LISTING is render-level proof only (t1434 G-f); actual execution is not
demonstrated by this measurement, so the no-verification fallback stands — the mirror is deployed with
entries re-homed to real directory copies (never dangling symlinks), and plugin-mode UPDATE runs hold
the mirror stable (MirrorPolicyNone on the update-path deployer; the re-home belongs to the fresh
plugin deploy).

Harness-debt notes (recorded, non-blocking): (i) the marker write is environment-sensitive — the
`--allowedTools` grant does not reliably apply through gateway/SDK session shapes, so the marker can
fail while resolution demonstrably succeeds (the session's own log is the disambiguator — the script
keeps the log on failure for exactly this); (ii) the codex render case carries no bounded wait (the two
claude cases are 420 s-bounded; the codex case is quick in every observed run).

### M2 — the flip (thin default deploy + full local counterpart)

| Claim | Evidence (command → deciding output, this run, this tree) |
|---|---|
| AC-001 (a) init thin set | `go test ./internal/cli -run '^TestDefaultDeploySetExcludesSkillsAndCommands$' -count=1 -v` → `--- PASS`; no `.claude/skills`/`.claude/commands` files, kept components present, record `plugin` |
| AC-001 (b) deployer split | `go test ./internal/template -run '^TestDeployerModeSplitsFileSet$' -count=1 -v` → `--- PASS` (local deploys everything; plugin excludes the two roots from walk AND ListTemplates) |
| AC-002 (a) sources retained | `go test ./internal/template -run '^TestEmbeddedSkillAndCommandSourcesRetained$' -count=1 -v` → `--- PASS`; 38 skill directories / 17 commands; catalog tiers 36/13/1. **Plan-pin correction (recorded):** spec P-02/L-24's "41" was measured through the session shell's `ls` alias (`ls -la`), whose output adds the total line and `.`/`..` to the 38 real directories — `/bin/ls` and `find` agree on 38. The unpolluted count is the pin. |
| AC-002 (b) emit machinery | `make commands-emit-check agents-emit-check` → both `ok`, exit 0 |
| AC-003 `--no-plugin` | `go test ./internal/cli -run '^TestNoPluginPathDeploysFullLocalPayload$' -count=1 -v` → `--- PASS`; full payload incl. moai entry; record `local` |
| AC-004 guidance | `go test ./internal/cli -run '^TestShrinkInitGuidanceOnMissingPlugin$' -count=1 -v` → `--- PASS`; exactly one block naming both recourses; exit unchanged |
| AC-005 MCP policy | `go test ./internal/cli -run '^TestDefaultPathMcpEntryPolicy$' -count=1 -v` → `--- PASS` (three arms: not-demonstrated → entry; confirmed → absent; `--no-plugin` → entry; context7+staggeredStartup preserved) |
| AC-006 mirror | `go test ./internal/template -run '^TestCodexMirrorFollowsDeployMode$' -count=1 -v` → `--- PASS` (local as today; plugin+none → no mirror; plugin+rehome → real directory copies, never symlinks) |
| AC-007 `--all` | `go test ./internal/cli -run '^TestAllFlagDeploysAllTiersLocally$' -count=1 -v` → `--- PASS` (full payload + every optional-pack entry + record `local`) |
| Probe arms | `go test ./internal/cli -run 'TestProbe' -count=1 -v` → all `--- PASS` (opted-out no-probe; env opt-out ≡ flag; confirmed requires absent→present; pre-existing plugin → not-demonstrated; unreadable → not-demonstrated; codex JSON arm) |
| M1-commit regression note | `TestProbeResurrection` (temporary probe, removed) established the AC-019 assertions must count FILES not directory shells — the classified executor removes files; empty directory shells may remain (removal unit = the classified file) |

### M3 — update scope + migration execution

| Claim | Evidence (command → deciding output, this run, this tree) |
|---|---|
| AC-015 trigger arms | `go test ./internal/cli -run '^TestUpdateMigratesLegacyProject$' -count=1 -v` → `--- PASS` (confirmed: record `plugin`, classified files removed, modified archived with the user's bytes, foreign preserved; not-demonstrated: record `local`, nothing removed/archived; opted-out: full local deploy, record `local`) |
| AC-011 classified executor | `go test ./internal/cli -run '^TestMigrationRemovesIdenticalDroppedComponents$' -count=1 -v` → `--- PASS`; counts line printed; no identical-component archive; foreign file survives the same run |
| AC-012 archive-before-removal | `go test ./internal/cli -run '^TestMigrationArchivesModifiedBeforeRemoval$' -count=1 -v` → `--- PASS`; negative control shows the unguarded path losing the file with no archive; per-file archive unit (no directory-level copy) |
| AC-014 idempotence | `go test ./internal/cli -run '^TestMigrationIdempotent$' -count=1 -v` → `--- PASS` (second run `--force` past the version-skip; archive set unchanged; record unchanged) |
| AC-016 thin redeploy | `go test ./internal/cli -run '^TestUpdatePluginModeSkipsDroppedRedeploy$' -count=1 -v` → `--- PASS`; record byte-identical |
| AC-017 local scope | `go test ./internal/cli -run '^TestUpdateLocalModeKeepsFullScope$' -count=1 -v` → `--- PASS` |
| AC-018 no flip | `go test ./internal/cli -run '^TestUpdateNeverFlipsModeRecord$' -count=1 -v` → `--- PASS` (± `--force`; guidance names the init re-entry) |
| AC-019 no resurrection | `go test ./internal/cli -run '^TestUpdateForceDoesNotResurrectDropped$' -count=1 -v` → `--- PASS`; the only remaining skill file is the preserved foreign skill |
| D-16 boundary (leader condition 5b) | `TestNotDemonstratedPreservationEndsAtNextLocalUpdate` → `--- PASS`: run 1 (not-demonstrated) preserves the foreign file; run 2 (record now `local`, full deployer + today's Clean walk) removes it WITH its pre-clean backup copy — the one-run preservation boundary is executable, with the recovery path asserted |
| D-18 note (one-line) | acceptance.md L-13's re-observation note records the archive-unit restatement as a criterion sharpening; the classified-FILE archive unit itself is asserted here by `TestMigrationArchivesModifiedBeforeRemoval` (per-file archives, no directory-level copy) — the body-level note stands corrected by this evidence line |

### M4 — docs + guarded surfaces

| Claim | Evidence (command → deciding output, this run, this tree) |
|---|---|
| AC-021 (a) init docs guard | `go test ./internal/cli -run '^TestInitDocsDescribeThinDeploy$' -count=1 -v` → `--- PASS` (static: `--no-plugin`/`--all` help + success card name the thin deploy and both paths) |
| AC-021 README/docs-site | README.md/README.ko.md/README.ja.md/README.zh.md gain the deploy-mode section naming `--no-plugin` and the plugin carrier; docs-site `cli-reference/init.md` ×4 locales gain the same (static text edits — verified by the M4 grep guard's tokens and reviewed in the diff) |
| Full package suite | `unset <13 MOAI_ lane vars> && go test -timeout 40m -count=1 -skip 'TestHandleCodexReviewGate_LiveCodex' ./internal/cli/` → 4 failures, ALL dispositioned below; every flip-relevant family green |
| Failure dispositions | (1) `TestCodex1718Fixtures_WidenedContent` — pre-existing parse-fixture failure, zero diff reach (`git diff 0a1e105d8 --stat` shows no review-gate file); (2) `TestWSR006_ReviewGateRootMatrix` — pre-existing/environmental (fails isolated at this tree; zero diff reach); (3) `TestUpdateLLMYAMLFirstDeployCalm` — flip guarded-surface, FIXED in this change set (the migration persists the record; the test now asserts template + record line); (4) `TestUpdateMirrorHeal_RestoresPathA` — flip guarded-surface, FIXED (fixture models the local-mode population; the update-path deployer holds the mirror stable via MirrorPolicyNone — re-run `--- PASS` both paths) |
| Lane-env hazard (recorded) | The unscrubbed full-suite run showed 281 failures — every one a `moai`-exec'ing test inheriting THIS lane session's MOAI_FACTORY_*/MOAI_KANBAN_* variables and hitting the lane-boundary guards; the live-enumerated `unset` of the 13 variables turns the family green (sample re-run `--- PASS` ×3). Lane-local verification requires the env-scrubbed form (kanban-dispatch § Verification load is lane-local) |
| pkilled-process note | Per the leader advisory, no internal/cli run in this session died by signal unexplained; the two stopped background suites were stopped deliberately (stale-tree runs superseded by fresher ones) |

### Card-review repairs (codex_review findings 1-6, 2026-10-04)

| Claim | Evidence (command → deciding output, this run, this tree) |
|---|---|
| F1 symlink-guarded archive | RED: `go test ./internal/cli -run TestMigrationArchiveRefusesSymlinkedArchiveDestination -count=1` → `the archive write went through a symlinked archive destination without refusal` / `the confirmed migration did not abort on the symlinked archive destination`; GREEN: `unset <13 MOAI_ lane vars> && go test -count=1 -v -run 'TestMigrationArchiveRefusesSymlinkedArchiveDestination|...' ./internal/cli/` → `--- PASS: TestMigrationArchiveRefusesSymlinkedArchiveDestination (0.01s)` (unit refusal names `.moai/archive`; flow aborts before any removal with record unwritten) |
| F2 migration re-home per design §3 | RED: `the mirror entry is still a symlink — it dangles once .claude/skills is removed`; GREEN: same `-v` run → `--- PASS: TestMigrationRehomesExistingMirrorEntries (0.18s)` (kept link becomes a real directory carrying the embedded bytes; no provisioning; user entry untouched; classified removal still ran) |
| F3 heal applies deploy mode (REQ-019) | RED: `plugin-mode heal resurrected a mirror (REQ-019): <nil>`; GREEN: same `-v` run → `--- PASS: TestMirrorHealRespectsDeployMode (0.01s)` (plugin record heals nothing; local record heals as today) |
| F4 preview shares the run's target computation | RED: `preview lists a file the plugin-mode run preserves:` (moai-custom listed under a plugin record); GREEN: same `-v` run → `--- PASS: TestPreviewManagedCleanup_ModeScoped (0.00s)` (plugin-mode preview omits the preserved file; local-mode preview still lists it); execution and preview both route through `computeRunCleanTargets` |
| F5 non-vacuous protected-set hash | RED: `the protected-set hash negative control did not pass (a modified protected file did not flip the hash — the tamper check is vacuous)`; GREEN: same `-v` run → `--- PASS: TestResolutionGateHarness (3.61s)`; self-check direct: `sh scripts/check-bare-name-resolution.sh --self-check` ×3 → `PASS protected-set-hash` + `PASS protected-set-hash-negative-control` + `RESULT pass=6 fail=0` (contents hashed via shasum over file bytes; measured churn classes pruned: codex sqlite/-wal/-shm stores, models_cache.json, live history jsonl) |
| F6 accurate re-entry guidance | RED: `guidance does not name the working re-entry flags (--no-plugin --force)` + static `init.go guidance block missing "--no-plugin --force"`; GREEN: same `-v` run → `--- PASS: TestShrinkInitGuidanceOnMissingPlugin (0.37s)` + `--- PASS: TestInitDocsDescribeThinDeploy (0.00s)` (guidance names `--no-plugin --force` and states the `.moai-backups/<timestamp>/` disposition + manifest carry-forward) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-10-04T22:40+09:00 (manager-develop, general-type worker, card worktree)
ac_matrix: 21/21 PASS — full command + verbatim-output matrix in §E.2 (M1/M2/M3/M4 tables)
builds: exit 0 ×2 (go build ./... and GOOS=windows GOARCH=amd64 go build ./..., final tree)
lint: 0 issues (baseline 0 — no NEW; golangci-lint --timeout=2m over internal/cli, internal/template, internal/config, internal/core)
boundary_grep: 0 matches (AskUserQuestion|mcp__askuser over all 13 touched non-test files)
coverage: internal/cli 84.9%, internal/template 83.4%, internal/config 83.7%, internal/core/project 89.0%, update 86.1%, deploy 91.6%, merge 93.1%, plan 95.0%, report 92.9%, backup 86.6% (internal/cli measured with -cover -timeout 45m, live-codex test skipped)
coverage_disposition: internal/cli, internal/template, internal/config sit 0.1–1.6pt under the 85% package-aggregate target — the gap is pre-existing untouched surface, the new migration/probe/mode-record code is fully covered by its own tests (Gaps, not a Claim)
suite_failures: full scrubbed run (-timeout 40m/-45m, TestHandleCodexReviewGate_LiveCodex skipped) → 4 failures, all dispositioned in §E.2 M4 table (2 pre-existing zero-diff-reach, 2 flip guarded-surfaces FIXED in-change; re-run PASS)
req008_gate: verdict recorded in §E.2 BEFORE the M2 flip commit (RESULT pass=5 fail=1 — the one fail is the marker-write mechanism; both session logs confirm bare-name resolution itself)
lane_env_hazard: recorded in §E.2 — lane-local runs require the live-enumerated 13-variable MOAI_* scrub
commits: d8e4300e4 M1, fd877e59a M2, 3b04fc13a M3, e22a41bee M4, b84d7baca gate fix (local only — leader batch push)
```


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-04
sync_commit_sha: c1623e65c
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-INIT-SHRINK-001' CHANGELOG.md → 0 (pre-emission clear; post-emission 1, entry appended under [Unreleased] ### Added, last row)"
b12_self_test_b: "MOAI-AC-COUNTER (awk, AC_FILE=.moai/specs/SPEC-INIT-SHRINK-001/acceptance.md) → stdout 21, live=21 excluded=0 ambiguous=0; CHANGELOG entry states 21 acceptance criteria (AC-001..021)"
b12_self_test_c: "ls of every CHANGELOG-named path → all exist: internal/config/deploy_mode.go, internal/template/apply_deploy_mode.go, internal/template/deployer_mode.go, internal/cli/update/migrate_classify.go, internal/cli/update_migrate.go, internal/cli/plugin_probe.go, internal/cli/update_mirror_heal.go, internal/cli/update_template_sync.go, internal/cli/update_dryrun_preview.go, scripts/check-bare-name-resolution.sh, README.md, README.ko.md, README.ja.md, README.zh.md, docs-site/content/{en,ko,ja,zh}/cli-reference/init.md"
changelog_entry_position: "[Unreleased] ### Added, last row (appended to the newest Added block)"
frontmatter_status_transitions:
  - in-progress → completed (spec.md, status: only — updated: 2026-10-04 already current; zero body edits)
mx_tag_deltas: "+2 @MX:ANCHOR (config.ReadDeployMode fan_in=4 NEW; template.ApplyDeployMode fan_in=4 NOTE→ANCHOR+REASON) · 0 stale @MX:TODO found in the new files · owning-test existence re-verified for all 21 ACs incl. AC-021's guard TestInitDocsDescribeThinDeploy"
doc_surface_validation: "README ×4 deploy-mode sections present (line ~282-284 per locale); docs-site cli-reference/init.md ×4 locale parity 5 h2 / 4 h3 each, 'no-plugin' ×2 per page"
backfill_note: "real sync_commit_sha backfilled in a following chore commit per the D3 SHA-placeholder exemption"
```

Sync-phase evidence summary: the sync-audit verdict for this close is written by the auditor to `.moai/reports/t1438/sync-audit.md` (after this commit); the run-phase per-milestone selector evidence this close rests on is §E.2 above.

## §F Phase 4 Mode Selection

Logged by the lane orchestrator (lane-15, card t1438) before the first run-phase Agent() spawn.

Input parameters: tier=L; scope ~15-20 files (internal/cli, internal/config, internal/template, scripts/, README + docs-site init pages); domains=3 (Go source, shell scripts, docs); language mix=go+sh+md; concurrency benefit=LOW (coding-heavy implementation); Agent Teams prerequisites=not requested.

Mode evaluation:

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-file implementation, not trivial |
| serial | YES | coding-heavy Tier L — Anthropic coding-task parallelism caveat; one manager-develop spawn, milestones sequential M1→M4 |
| fanout | no | not research-heavy multi-domain work |
| sweep | no | semantic new-code work, not a mechanical uniform transform |
| agent-team | no | explicit-request-only; not requested |

Decision: serial

Justification: the card is coding-heavy Go/template implementation, where sequential single-specialist delegation is the safe default per the coding-task caveat. The factory in-lane 3-stage model routes this card's run phase directly to manager-develop (factory-dispatch), so the manager-lead coordination threshold is not taken from inside the lane — that surface belongs to the leader session's dispatch cycle.

Plan→run Kickoff gate: NOT MET — blocked pending leader decision (2026-10-04). Iteration-3 audit FAIL 0.85, receipt rcpt-10fb620edbc90ece3943e22f, 4/4 new findings source-verified (D-10..D-13); D-7/D-8/D-9 repairs verified RESOLVED. Retry cap 3/3 reached — escalation to the leader per the retry contract. The MET record below cites the iteration-2 verdict that the cross-model receipt overturned; retained as history, superseded by this line.

Leader ruling received 2026-10-04 (mission contract 11c79e1a): option ① — repair D-11/D-12/D-13, then re-entry. Directions: D-12 — REQ-010's foreign definition explicitly classifies moai-custom AND user-created skills as foreign (preserved); REQ-013/AC-013 aligned. D-11 — the install probe compares a PRE-EXECUTION snapshot against the post-execution state; only a demonstrated this-install success admits plugin; every other outcome keeps local with no removal; unobservable trigger names are removed. D-13 — the migration surface and the init surface are separately scoped. The leader opens ONE iteration-4 delta audit under leader authority beyond the 3-iteration retry cap; a codex receipt is required there too; any critical user-file-loss defect remaining in iteration 4 defers t1438 to post-3.2.

decision record: decided_by=leader (relay 2026-10-04, mission contract 11c79e1a) evidence_refs=.moai/reports/t1438/plan-audit-iter3.md ladder_path=gate-row plan-to-run Kickoff (operator ruling: option 1 repair-then-reentry; iteration-4 audit authorized beyond retry cap)

decision record: decided_by=leader (relay 2026-10-04, mission contract 11c79e1a, convergence rule) evidence_refs=.moai/reports/t1438/plan-audit-iter4.md;receipt=rcpt-755bbbcc6218e0306cb7de90 disposition=PASS-WITH-DEBT (debt acceptance; deferral rule not triggered; no further audit opened) conditions=D-14..17-line-repair-before-manager-develop;D-16-promise-narrowed-to-REQ-017-scope;D-17-archive-unit-aligned-to-per-file-classification;spec-lint-0-after-repair;D-15/16-loss-paths-RED-in-implementation ladder_path=gate-row plan-to-run Kickoff (gate MET under the leader's debt-acceptance disposition, operator form)

Operator post-hoc review item: the iteration-4 audit's opening beyond the 3-iteration retry cap, exercised under leader authority on 2026-10-04, is recorded here for operator post-hoc review. The subsequent debt-acceptance disposition (PASS-WITH-DEBT over the FAIL 0.85 verdict, with the five binding conditions above) is recorded on the same surface for the same review.

Leader routing on the receipt-guard block (2026-10-04): the lane's option ① (auditor re-emits its closing line as PASS-WITH-DEBT) was REJECTED by the leader — rewriting the auditor's machine-read FAIL line would present the leader disposition as an audit PASS and circumvent the guard's purpose; the disposition section on the verdict file stands untouched. The leader instead reversed the no-further-audit decision FOR THIS REASON and authorized ONE minimal delta audit: plan-auditor re-reads only the D-14..D-17 repair portions of commit 68d9f2a9b, with a codex_audit receipt, applying the convergence rule (no critical defect → PASS-WITH-DEBT); a new critical defect there defers the card to post-3.2. The manager-develop spawn follows the auditor's own verdict + receipt closing line.

decision record: decided_by=lane-15 orchestrator (Claude, card t1438) evidence_refs=.moai/reports/t1438/plan-audit-iter2.md;verdict=PASS-WITH-DEBT;score=0.925;tier_threshold=0.85;audited_sha=49557f87c;artifact_hash=6b4c9498acb231857779abb209fa48cbbbdb4cba94927b185965c823881d0420 (recomputed unchanged this run);depends_on=SPEC-PLUGIN-MARKETPLACE-001:completed;open_blockers=0 ladder_path=gate-row plan-to-run Kickoff (AUTONOMOUS, auto-semantics §9.1)

Progression mode: autonomous (default; leader-designated run, 2026-10-03). Audit debt carried into run-phase execution notes: D-2 (capture each criterion's own selector command at its first RED/GREEN observation), D-3 (R-22 selector-name typo — recorded optional debt), D-5 (absorb guarded-surface test updates into the flip's change set, not deferred to M4).

decision record: decided_by=lane-15 orchestrator (Claude, card t1438) evidence_refs=.moai/reports/t1438/plan-audit-iter3.md;verdict=FAIL;score=0.85;receipt=rcpt-10fb620edbc90ece3943e22f;new_blocking=D-10..D-13;prior_repairs=D-7..D-9:RESOLVED;retry_cap=3/3-reached ladder_path=gate-row plan-to-run Kickoff (BLOCKED — leader escalation, retry contract)
