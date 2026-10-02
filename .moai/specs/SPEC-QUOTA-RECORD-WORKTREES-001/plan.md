# plan.md — SPEC-QUOTA-RECORD-WORKTREES-001 (card t1442)

Tier M (11 REQ / 13 AC, ceilings 16 / 16; the leader's verdict on D4 added one REQ and one AC and the tier is unchanged; the file-count estimate sits at the top of the Tier M band, see spec.md Tier paragraph). Sections are ordered by decision reversibility: the decisions most likely to change first (§B), the evidence step the leader asked for (§C), then the file map and milestones, with the mechanical steps last. No time estimates; priority and ordering only.

## §A Context

- Work tree: `.moai/worktrees/t1442`, branch `WT-quota-read-worktree-records`, planned at HEAD `284e09c44023598affe486f17701717ca173e6ca` (develop tip at planning time).
- Methodology: TDD, the run-phase default (assumption U4: `constitution.development_mode` in `quality.yaml` was not read in this phase); every AC is decided by a Go test written RED first. The run phase starts only after the plan→run Kickoff gate is met (autonomous form: plan-audit PASS at the Tier M threshold 0.80 and unchanged artifact hashes, `auto-semantics.md` §9.1).
- Artifacts (Tier M, 3 plus 2): `spec.md` (11 REQ, decisions D1-D8), `plan.md`, `acceptance.md` (13 AC, RED-now ledger), `decision-index.md` (decision rows, no recommendations), `progress.md` (§E skeleton).
- Depends on SPEC-QUOTA-AWARE-SCHEDULING-001 (completed). It is not edited; its tests stay green without an edit (REQ-QWR-010, AC-QWR-003, AC-QWR-012).

## §B Decisions most likely to change (review these first)

All five were resolved at spec 0.2.0 by the decision oracle (Jev); the leader's verdict (0.3.0) then ACCEPTED D4 and D2 provisionally. Those two stay the ones most likely to change; review them first.

| Decision | Status and default | Why it may change | Cost of changing it later |
|---|---|---|---|
| D4 bound and order | **ACCEPTED provisionally by leader verdict** (oracle leaned 64 by newest metadata, confidence 0.22, LOW): 128 directories, name order, 4 KiB per `gitdir`, all unmeasured; **the bound is the configuration key `workflow.quota_gate.max_scan_dirs`** (default 128, range 1-1024; REQ-QWR-011) | the 128 is unmeasured and the oracle's lean differs | none in code: an operator changes the key; changing the default is the Go default, the template, the local twin, and the AC-QWR-013 expectation; the order key is one comparison |
| D2 seam shape | **ACCEPTED provisionally by leader verdict** (oracle 0.71, confidence 0.42): seam type unchanged, new thin production value deriving the root from `<root>/.moai/state` and reading `max_scan_dirs` through `config.LoadQuotaGate(root)`. **If deriving the root from the path shape proves unsafe in the run phase, switch to a type change of the seam and report to the leader** (the type change edits `factory_quota.go` and the two predecessor helpers `factory_quota_test.go:246-270`, which AC-QWR-003 and REQ-QWR-010 forbid editing, so it needs the leader's decision and a spec revision first; the second `workflow.yaml` parse per evaluation is a second reason to take that fallback) | the path-shape derivation is a convention, not a type guarantee | a signature change touches `factory_quota.go` and `factory_quota_test.go:246-270` |
| D1 which directories | **RESOLVED** (oracle 1.00): (A) primary + all linked worktrees from `<common>/worktrees/*/gitdir` | the card text says "linked worktrees (or the session's own project dir)"; (B) is cheaper and sees fewer readings | the enumerator is one function; swapping changes `quota_dirs.go` and the fixtures of AC-QWR-001 only |
| D5 observability | **RESOLVED** (oracle 0.97): no new output; predecessor goldens and tests untouched | an operator may want to see where a reading came from | additive follow-up with its own AC; would edit the predecessor goldens if done now |
| D6 which command is measured | **RESOLVED** (oracle 0.83): `moai factory status` always (read-only, before and after); `moai factory next` additionally only when a lease is genuinely due anyway | the leader may require `next` itself | the §C protocol is command-agnostic; only the "measured command" line changes |

## §C The real-lane measurement (leader requirement; protocol QWR-M0)

Run phase, milestone M0, its own commit before any implementation commit (REQ-QWR-009, AC-QWR-011). Evidence cell: `progress.md` §E.2, heading `QWR-M0`.

**Why `moai factory next` is not run as-is.** It leases a card when it succeeds (`factory_card.go:662-675`), an unwanted side effect for a measurement. A throwaway queue is possible only by pointing `CLAUDE_PROJECT_DIR` at a throwaway git repository with the lane environment variables set (`factory_card.go:50-52` admission; `session.go:262-280` root resolution), which measures that repository's own record directory and so cannot answer where a REAL lane's record lives; and the command path resolves the root through the adopting form (`todo.go:74-76`) whose side effects were not audited. Hence the default is the read-only `moai factory status` (`Short: "Report factory card records (read-only; an expired lease is shown, never returned)"`, `factory_card.go:1393`), which evaluates the same pressure function: `factoryQuotaStatusBlock(root)` (`factory_card.go:1439`) calls `factoryQuotaEvaluate(root)` (`factory_quota.go:248`), the function `next` reaches through the latch (`factory_quota.go:136`). `moai factory next --help` was not run in this phase; the flag set was read from source (`--wait`, `--wait-bound`, `--run`; no root flag, `factory_card.go:691-695`).

**Protocol** (each numbered step is a separate tool call so that statusline renders can fall between them; a lane session, not a subagent):

1. `date +%Y-%m-%dT%H:%M:%S%z` → T0.
2. `moai session current` → the lane's session id `SID`; `moai version` → the installed build's commit, and `git merge-base --is-ancestor <that commit> HEAD` (exit 0 = the installed build may lag the tree, `verification-claim-integrity.md` §2.2).
3. Locate every copy: `find <primary>/.moai/worktrees <primary>/.claude/worktrees -maxdepth 6 -path "*/.moai/state/context-usage/SID.json"`.
4. `ls -la <primary>/.moai/state/context-usage/SID.json` and `ls -la` of each copy step 3 found, one call (the lane's current worktree and any other).
5. `cat` the freshest copy (its `captured_at`, `writer_pid`, `schema_version`, and whether a window field exists).
6. The measured command: `moai factory status` (always). When the lane is entitled to a lease anyway at this point of its card, also run `moai factory next` exactly as the lane protocol does (the measurement then piggybacks on a real lease; no artificial lease is created). The lane must be anchored at the parent checkout for `next` (`factory_card.go:645`); record whether it was (`git rev-parse --show-toplevel`).
7. `date +%Y-%m-%dT%H:%M:%S%z` → T1; repeat steps 3-4; then send one assistant message (letting a statusline render land) and repeat 3-4 once more → T2.
8. Write into `progress.md` §E.2 `QWR-M0`: T0/T1/T2, the installed build and its ancestry result, the verbatim `ls -la` and `cat` outputs, the command measured and which variant, the anchor of the lane at step 6, and the statements below. Commit it alone.

**What the observation can establish:** (i) which directory holds the lane's own record at the moment of the call and how old it is, against T0/T1; (ii) whether copies of the same session id remain in other directories and how stale they are (the shape the gate's freshest-wins rule must cope with); (iii) for a lane anchored at the parent checkout when it runs `next`, whether its freshest record is in the primary directory, the premise the predecessor audit left inferred (F1, "NOT observed"); (iv) the same variant answers for `status`, `--auto`, and the acquire warning, which share the evaluation.

**What it cannot establish:** (a) any window value — the installed binary writes schema 2 with no windows (measured below), so no real hold decision can be observed until the rebuilt binary is installed; (b) that the command writes or refreshes a record — no factory verb writes records, the statusline does on a render, so an unchanged mtime across the call shows only that no render fell inside the interval; (c) the primary-anchored leader's render cadence, hence whether the primary directory stays under `max_age`; (d) the `--wait` loop or the latch (the status variant never enters them; AC-QWR-008 covers them in fixtures); (e) cost: the step is not a timing measurement (a separate timing observation of the evaluation is recorded as a data point, D4).

**Recorded data points already made (not re-executed by the run phase, and not a substitute for QWR-M0):**

- **DP1** (2026-10-02, reported by the card's lane after `ExitWorktree` returned the session to the primary checkout): the session's copy in the primary directory started updating again (mtime 18:28) while the copy in the card worktree stayed frozen at its last in-worktree write (18:21).
- **DP2** (this plan phase, `date` 2026-10-02T18:32:26+0900, `ls -la` of session `2da35a68-1196-4183-b6e5-a50fc9b6d901` in four directories): `.claude/worktrees/develop` 18:26, primary 18:28, `.moai/worktrees/t1347` 18:21, `.moai/worktrees/t1442` 18:30 (the lane entered this tree after 18:28); the t1442 copy reads `"schema_version": 2`, `"captured_at": "2026-10-02T18:30:11.548855+09:00"`, no window field. Mtimes are minute-resolution as `ls -la` prints them.
- **DP3** (this plan phase): `find <primary>/.moai/state/context-usage -name "*.json" -mmin -60` printed 4 files; the same `find` over the two worktree roots (`-maxdepth 6`) printed 19 files in 17 directories; record files on disk: 1043 in the primary directory, 624 in worktree directories.
- **DP4** (predecessor evidence, `sync-audit.md` §6 F, 17:46-18:00 the same day): the primary copy of `2da35a68-…` frozen at 14:28 against 17:46 in `.moai/worktrees/t1347`.

## §D File map (indicative; confirmed in the run phase)

| File | Change | Milestone |
|---|---|---|
| `.moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001/progress.md` | the `QWR-M0` evidence cell, alone in its commit | M0 |
| `internal/statusline/quota_dirs_test.go` (new) | `TestQWR_AC001..006`, `TestQWR_AC010` (statusline decided), written RED first | M1 |
| `internal/cli/factory_quota_worktrees_test.go` (new) | `TestQWR_AC007..009`, `TestQWR_AC010b`, `TestQWR_AC013b`, written RED first; reuses the predecessor's `qasFixture` helpers without editing them | M1 |
| `internal/config/quota_gate_scan_dirs_test.go` (new) | `TestQWR_AC013_MaxScanDirsConfigKey`, written RED first (the file name is the one AC-QWR-012 allows; `workflow_quota_gate_test.go` of the predecessor is not edited) | M1 |
| `internal/config/types.go` | add `MaxScanDirs int \`yaml:"max_scan_dirs"\`` to `QuotaGateConfig` (`:775`) | M2a |
| `internal/config/defaults.go` | add the default constant 128 beside `DefaultQuotaGate*` (`:89-92`) and seed it in `NewDefaultWorkflowConfig` (`:1196-1201`) | M2a |
| `internal/config/loader_quota_gate.go` | add `MaxScanDirs` to `QuotaGateSettings`, the range constants 1 and 1024, the default in `DefaultQuotaGate`, and the range check in `resolveQuotaGate` (replace out-of-range by the default, as the other keys) | M2a |
| `internal/config/cache.go` | `configCacheSchemaVersion` 11 to 12 with a comment line in the established style (`:44-47`) | M2a |
| `internal/config/testdata/shipped_key_inventory.yaml` | row `workflow.quota_gate.max_scan_dirs`, class W, evidence reader, between `max_age` and `release_margin_pct` (`:2406-2410`) | M2a |
| `internal/template/templates/.moai/config/sections/workflow.yaml` | key `max_scan_dirs: 128` in the `quota_gate` block (`:201-206`) and a comment line in the block comment (`:182-200`) saying the value is unmeasured and that more worktrees than the key need a larger value; `make build` regenerates the embedded templates | M2a |
| `.moai/config/sections/workflow.yaml` | the same key in the local twin (`:223-228`) | M2a |
| `internal/statusline/quota.go` | extract the per-directory scan loop; `AggregateQuota` becomes a one-element call into the multi-directory reading; capture-time and mtime checks keep their text | M2 |
| `internal/statusline/quota_dirs.go` (new) | the file-read enumerator (primary state dir first, then one state dir per usable `<common>/worktrees/<entry>/gitdir`, name order, bound passed in as an argument, 4 KiB per file); a function deriving the repository root from a `<root>/.moai/state` path | M1 (signature-only stub), M2 (body) |
| `internal/cli/factory_quota.go` | the production value of the seam `factoryQuotaAggregate` becomes a thin function of the unchanged type that reads `config.LoadQuotaGate(root).MaxScanDirs` and calls the multi-directory reading; `factoryQuotaEvaluate` and its MX anchor unchanged | M3 |
| `CHANGELOG.md` | a new entry; the "알려진 한계" limit sentence of the predecessor is superseded in the new entry (the predecessor entry is not edited by this SPEC) | sync |

PRESERVE (not touched, AC-QWR-012): `internal/stateanchor/**`, `internal/statusline/state_anchor.go`, `internal/statusline/context_usage.go`, `internal/statusline/quota_test.go`, `internal/cli/factory_quota_test.go`, `internal/config/workflow_quota_gate_test.go`, `.moai/specs/SPEC-QUOTA-AWARE-SCHEDULING-001/**`; inside `internal/config` only the five files in the rows above change, and the five predecessor `quota_gate` keys, their defaults, and their ranges are untouched (only additions to `defaults.go` and `loader_quota_gate.go`). The one config key and the cache schema bump are REQ-QWR-011 (leader verdict on D4); the earlier "no new config key" default is overridden for this key alone.

## §E Milestones

- **M0 — Baseline first: the real-lane measurement (REQ-QWR-009, AC-QWR-011).** Run the §C protocol, commit `progress.md` with `QWR-M0` alone, before any other run-phase commit. The phase `Evidence` also records the one timing observation of the current evaluation (D4) here, so it is measured on the pre-change code. Evidence cell: `QWR-M0` plus the commit SHA B.
- **M1 — RED tests (all test-decided ACs).** Write the three new test files (statusline, cli, config). The new entry points do not exist yet, and a compile error (`undefined: …`) is not an accepted RED reason, so the M1 commit also carries a signature-only stub in `quota_dirs.go` (the enumerator returns no worktree directories; the multi-directory reading reads only the first directory) that changes no behaviour of any existing function. Run each anchored selector once and record the verbatim failing assertion (`progress.md` §E.2, "E8"): the worktree-only reading is not seen. This is the first implementation-file commit, so M0's commit B must precede it (AC-QWR-011).
- **M2a — The configuration key (REQ-QWR-011; AC-QWR-013 config half).** Add the field, default, range check, cache bump, inventory row, template block with the "unmeasured" comment, and the local twin (file map rows above); `make build`; run `TestQWR_AC013_MaxScanDirsConfigKey`, the shipped-key guard `TestShippedConfigKeysHaveReaders`, and the predecessor's `TestQAS_AC007_ConfigDefaultsMirrorTemplate` (green, unedited). This is a data-model change and is first of the mechanical steps because M2 and M3 consume the value.
- **M2 — Statusline reach (REQ-QWR-001..005, -008; AC-QWR-001..006, -010).** Extract the scan loop, add the multi-directory reading and `quota_dirs.go` (the bound is an argument; the statusline package imports `config` already, `quota.go:9`, but does not call `LoadQuotaGate`); run the statusline selectors and the full `TestQAS_` statusline/config set; confirm the `quota*.go` sweep includes the new file. Decision order inside the milestone: the bound and error handling last.
- **M3 — The gate sees it (REQ-QWR-006, -007; AC-QWR-007..009, -010b, -013b).** Replace the production value of the seam (root from the path shape, bound from `config.LoadQuotaGate(root).MaxScanDirs`); run the anchored cli selectors (scrubbed form, under a `moai slot go-test-cli-<resource>` lease named for the shared target, not the card) and the six old cli guards of AC-QWR-003 and the two of AC-QWR-009. Check at this milestone whether the path-shape derivation is safe and the second `workflow.yaml` parse per evaluation is acceptable; if either is not, take the D2 fallback (type change) and report to the leader before editing any predecessor file.
- **M4 — Closure evidence (AC-QWR-011, -012; §G of acceptance.md).** `go vet`, windows build, lint baseline, `moai spec lint`; the one-shot git commands of AC-QWR-011/-012 with their positive controls, recorded in `progress.md` §E.2/§E.3.
- **M5 — Sync and scoped re-audit (sync phase).** One sync commit (CHANGELOG entry; `status` transition by manager-docs per the ownership matrix) and a sync-auditor run scoped to this SPEC's changed files, its verdict read from `.moai/reports/t1442/` before the card advances. Mechanical.

## §F Self-verification (run phase reports each, per verification-claim-integrity §3)

Command and verbatim output for: the anchored selectors of AC-QWR-001..010 (non-empty swept count each); the old `TestQAS_` deciders; `go vet ./internal/statusline ./internal/cli`; `GOOS=windows GOARCH=amd64 go build ./...`; the AC-QWR-011/-012 git commands with positive controls; the `QWR-M0` commit SHA relative to the first implementation commit. Gaps are named, never silent. The judging build of any tool measurement is named next to the tree HEAD (§2.2).

## §G Unverified assumptions (not established in this plan phase)

- **U1** The on-disk layout `<common>/worktrees/<entry>/gitdir` (one line, an absolute path to the worktree's `.git` file) is stable across git versions. Measured only on the installed git against 27 entries today; `git --version` was not recorded. The `real_git_worktree_layout` subtest is the guard.
- **U2** Windows `gitdir` content (`C:/…`) and absolute-path handling were not examined.
- **U3** The per-call cost: about 1667 file stats (1043 + 624 measured by `find`) versus 1043 today, every 5 s per waiting lane; no timing was measured. The 128-directory bound (now the configurable `max_scan_dirs`) is therefore unmeasured; so is its upper limit of 1024. `find` used `-maxdepth 6`, which could miss deeper trees.
- **U4** The development mode (TDD) was not read from `quality.yaml`.
- **U5** `moai factory status` is read-only: taken from its `Short` text and a reading of `factory_card.go:1396-1439`; `homestate.OpenFactory` (schema DDL, migrations, WAL pragma) when the database exists was not audited as a pure read. It is a routine operator command; the protocol's purpose is record directories, not the factory database.
- **U6** A lane running `moai factory next` is anchored at the parent checkout at that moment: required by `factoryAssertParentCheckout` and consistent with DP1-DP2, never observed for a lane in the act of running it (QWR-M0 step 6 records the anchor).
- **U7** `resolveTodoQueueRoot` returns the primary checkout in production (read: `todo.go:63-76` comments; `kanban.ResolveTodoQueueRootAdopting` itself was not read). A root whose `.git` is a file therefore does not occur in production (D8).
- **U8** The seam production value derives the repository root from the exact string `filepath.Join(root, ".moai", "state")` (`factory_quota.go:93`); a differently shaped state directory path would read the one directory only (fail open).
- **U9** The installed `moai` build (`v3.2.0-rc.25`, commit `802a72235`, built 2026-10-02T08:00:14Z) is an ancestor of this tree's HEAD (`git merge-base --is-ancestor 802a72235 HEAD` exit 0), so `moai spec lint` / `moai spec audit` in this phase ran on a build that may predate lint rules added since; the verbatim output is reported with that attribution.
- **U10** `locked` markers mean "do not prune" and say nothing about the worktree being live; this is git behaviour from knowledge, not verified here. A locked worktree is read (AC-QWR-004).
- **U11** A stray record directory `<root>/.claude/worktrees/.moai/state/context-usage/` exists on disk (seen in a listing, stale); no investigation of how it was created (writer domain, out of scope).
- **U12** The CI runner has `git` on `PATH` for `real_git_worktree_layout` (the workflow file was grepped for the test command only).
- **U13** Adding a field to `QuotaGateConfig` keeps every predecessor config and cli test green without an edit (`workflow_quota_gate_test.go` and the `TestQAS_` set were not read in full; `TestQAS_AC007`'s subtests `defaults_equal_template` and `cache_schema_bumped` are expected to survive because the former compares the default with the template and the latter asserts a version above 10). Not run; the run phase's M2a is the check.
- **U14** The file-count and line estimate (about 15 files, 400-800 lines) is a guess from the file map, not a measurement; if it exceeds the Tier M file band the leader decides on re-tiering (the REQ/AC counts, 11 / 13, are the binding caps).
- **U15** The valid range 1-1024 is a plan choice with no measurement behind it; the upper bound exists only so a mistyped value cannot make the scan unbounded.
- **U16** `config.LoadQuotaGate(root)` called a second time per evaluation costs one more `workflow.yaml` parse; no timing was measured, and whether the loader caches is not known (`loadYAMLFile` was not read).
- **U17** The embedded-template regeneration (`make build`) is assumed to be the established step for a template `workflow.yaml` edit, as the predecessor's plan states; it was not run here.

## §H Risks and anti-patterns

- **Reading a worktree that is not checked out any more.** Mitigated by the missing-target skip (REQ-QWR-003).
- **Quietly widening scope.** The three neighbours F3-F5 stay out except the test-only F3 pin (D7); the writer anchor is not touched (REQ-QWR-010).
- **A refactor that changes the single-directory result.** AC-QWR-003 compares the old function against the new reading on the same fixtures and the old `TestQAS_` suite runs unedited.
- **Time-based or machine-load-based assertions.** None: the bound and the layouts are count-based; the timing observation is a data point.
- **Measuring `next` by leasing.** Prohibited as a default (§C).

## §I Cross-references

`spec.md` (REQ-QWR-001..010, D1-D8), `acceptance.md` (AC-QWR-001..012, §E ledger), `decision-index.md`; predecessor `.moai/specs/SPEC-QUOTA-AWARE-SCHEDULING-001/` (progress.md §E.4 debt D1, §J), `.moai/reports/t1347/sync-audit.md` (F1, F2), `.claude/rules/moai/development/verification-completeness.md` §1.3 and §2, `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 and §2.3.
