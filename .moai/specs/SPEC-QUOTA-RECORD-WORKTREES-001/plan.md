# plan.md — SPEC-QUOTA-RECORD-WORKTREES-001 (card t1442)

Tier M (10 REQ / 12 AC, ceilings 16 / 16). Sections are ordered by decision reversibility: the decisions most likely to change first (§B), the evidence step the leader asked for (§C), then the file map and milestones, with the mechanical steps last. No time estimates; priority and ordering only.

## §A Context

- Work tree: `.moai/worktrees/t1442`, branch `WT-quota-read-worktree-records`, planned at HEAD `284e09c44023598affe486f17701717ca173e6ca` (develop tip at planning time).
- Methodology: TDD, the run-phase default (assumption U4: `constitution.development_mode` in `quality.yaml` was not read in this phase); every AC is decided by a Go test written RED first. The run phase starts only after the plan→run Kickoff gate is met (autonomous form: plan-audit PASS at the Tier M threshold 0.80 and unchanged artifact hashes, `auto-semantics.md` §9.1).
- Artifacts (Tier M, 3 plus 2): `spec.md` (10 REQ, decisions D1-D8), `plan.md`, `acceptance.md` (12 AC, RED-now ledger), `decision-index.md` (decision rows, no recommendations), `progress.md` (§E skeleton).
- Depends on SPEC-QUOTA-AWARE-SCHEDULING-001 (completed). It is not edited; its tests stay green without an edit (REQ-QWR-010, AC-QWR-003, AC-QWR-012).

## §B Decisions most likely to change (review these first)

| Decision | Default | Why it may change | Cost of changing it later |
|---|---|---|---|
| D1 which directories (`[DECISION-OPEN]`) | (A) primary + all linked worktrees from `<common>/worktrees/*/gitdir` | the card text says "linked worktrees (or the session's own project dir)"; (B) is cheaper and sees fewer readings | the enumerator is one function; swapping A for B or C changes `quota_dirs.go` and the fixtures of AC-QWR-001 only |
| D4 bound and order (`[DECISION-OPEN]`) | 128 entries, name order, 4 KiB per `gitdir` | the 128 is unmeasured | one constant and the AC-QWR-005 fixture sizes |
| D5 observability (`[DECISION-OPEN]`) | no new output | an operator may want to see where a reading came from | additive follow-up with its own AC; would edit the predecessor goldens if done now |
| D6 which command is measured (`[DECISION-OPEN]`) | `moai factory status`, plus `next` when a real lease is due | the leader may require `next` itself | the §C protocol is command-agnostic; only the "measured command" line changes |
| D2 seam shape | seam type unchanged, new thin production value deriving the root from `<root>/.moai/state` | changing the seam type is cleaner but edits two predecessor test helpers | a signature change touches `factory_quota.go` and `factory_quota_test.go:246-270` |

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
| `internal/cli/factory_quota_worktrees_test.go` (new) | `TestQWR_AC007..009`, `TestQWR_AC010b`, written RED first; reuses the predecessor's `qasFixture` helpers without editing them | M1 |
| `internal/statusline/quota.go` | extract the per-directory scan loop; `AggregateQuota` becomes a one-element call into the multi-directory reading; capture-time and mtime checks keep their text | M2 |
| `internal/statusline/quota_dirs.go` (new) | the file-read enumerator (primary state dir first, then one state dir per usable `<common>/worktrees/<entry>/gitdir`, name order, bound 128, 4 KiB per file); a function deriving the repository root from a `<root>/.moai/state` path | M1 (signature-only stub), M2 (body) |
| `internal/cli/factory_quota.go` | the production value of the seam `factoryQuotaAggregate` becomes a thin function of the unchanged type calling the multi-directory reading; `factoryQuotaEvaluate` and its MX anchor unchanged | M3 |
| `CHANGELOG.md` | a new entry; the "알려진 한계" limit sentence of the predecessor is superseded in the new entry (the predecessor entry is not edited by this SPEC) | sync |

PRESERVE (not touched, AC-QWR-012): `internal/stateanchor/**`, `internal/statusline/state_anchor.go`, `internal/statusline/context_usage.go`, `internal/config/**`, `internal/statusline/quota_test.go`, `internal/cli/factory_quota_test.go`, `.moai/specs/SPEC-QUOTA-AWARE-SCHEDULING-001/**`. No new template or config key, so no config cache schema bump (REQ-QWR-010, D3).

## §E Milestones

- **M0 — Baseline first: the real-lane measurement (REQ-QWR-009, AC-QWR-011).** Run the §C protocol, commit `progress.md` with `QWR-M0` alone, before any other run-phase commit. The phase `Evidence` also records the one timing observation of the current evaluation (D4) here, so it is measured on the pre-change code. Evidence cell: `QWR-M0` plus the commit SHA B.
- **M1 — RED tests (all test-decided ACs).** Write the two new test files. The new entry points do not exist yet, and a compile error (`undefined: …`) is not an accepted RED reason, so the M1 commit also carries a signature-only stub in `quota_dirs.go` (the enumerator returns no worktree directories; the multi-directory reading reads only the first directory) that changes no behaviour of any existing function. Run each anchored selector once and record the verbatim failing assertion (`progress.md` §E.2, "E8"): the worktree-only reading is not seen. This is the first implementation-file commit, so M0's commit B must precede it (AC-QWR-011).
- **M2 — Statusline reach (REQ-QWR-001..005, -008; AC-QWR-001..006, -010).** Extract the scan loop, add the multi-directory reading and `quota_dirs.go`; run the statusline selectors and the full `TestQAS_` statusline/config set; confirm the `quota*.go` sweep includes the new file. Decision order inside the milestone: the bound and error handling last.
- **M3 — The gate sees it (REQ-QWR-006, -007; AC-QWR-007..009, -010b).** Replace the production value of the seam; run the anchored cli selectors (scrubbed form, under a `moai slot go-test-cli-<resource>` lease named for the shared target, not the card) and the six old cli guards of AC-QWR-003 and the two of AC-QWR-009.
- **M4 — Closure evidence (AC-QWR-011, -012; §G of acceptance.md).** `go vet`, windows build, lint baseline, `moai spec lint`; the one-shot git commands of AC-QWR-011/-012 with their positive controls, recorded in `progress.md` §E.2/§E.3.
- **M5 — Sync and scoped re-audit (sync phase).** One sync commit (CHANGELOG entry; `status` transition by manager-docs per the ownership matrix) and a sync-auditor run scoped to this SPEC's changed files, its verdict read from `.moai/reports/t1442/` before the card advances. Mechanical.

## §F Self-verification (run phase reports each, per verification-claim-integrity §3)

Command and verbatim output for: the anchored selectors of AC-QWR-001..010 (non-empty swept count each); the old `TestQAS_` deciders; `go vet ./internal/statusline ./internal/cli`; `GOOS=windows GOARCH=amd64 go build ./...`; the AC-QWR-011/-012 git commands with positive controls; the `QWR-M0` commit SHA relative to the first implementation commit. Gaps are named, never silent. The judging build of any tool measurement is named next to the tree HEAD (§2.2).

## §G Unverified assumptions (not established in this plan phase)

- **U1** The on-disk layout `<common>/worktrees/<entry>/gitdir` (one line, an absolute path to the worktree's `.git` file) is stable across git versions. Measured only on the installed git against 27 entries today; `git --version` was not recorded. The `real_git_worktree_layout` subtest is the guard.
- **U2** Windows `gitdir` content (`C:/…`) and absolute-path handling were not examined.
- **U3** The per-call cost: about 1667 file stats (1043 + 624 measured by `find`) versus 1043 today, every 5 s per waiting lane; no timing was measured. The 128-entry bound is therefore unmeasured. `find` used `-maxdepth 6`, which could miss deeper trees.
- **U4** The development mode (TDD) was not read from `quality.yaml`.
- **U5** `moai factory status` is read-only: taken from its `Short` text and a reading of `factory_card.go:1396-1439`; `homestate.OpenFactory` (schema DDL, migrations, WAL pragma) when the database exists was not audited as a pure read. It is a routine operator command; the protocol's purpose is record directories, not the factory database.
- **U6** A lane running `moai factory next` is anchored at the parent checkout at that moment: required by `factoryAssertParentCheckout` and consistent with DP1-DP2, never observed for a lane in the act of running it (QWR-M0 step 6 records the anchor).
- **U7** `resolveTodoQueueRoot` returns the primary checkout in production (read: `todo.go:63-76` comments; `kanban.ResolveTodoQueueRootAdopting` itself was not read). A root whose `.git` is a file therefore does not occur in production (D8).
- **U8** The seam production value derives the repository root from the exact string `filepath.Join(root, ".moai", "state")` (`factory_quota.go:93`); a differently shaped state directory path would read the one directory only (fail open).
- **U9** The installed `moai` build (`v3.2.0-rc.25`, commit `802a72235`, built 2026-10-02T08:00:14Z) is an ancestor of this tree's HEAD (`git merge-base --is-ancestor 802a72235 HEAD` exit 0), so `moai spec lint` / `moai spec audit` in this phase ran on a build that may predate lint rules added since; the verbatim output is reported with that attribution.
- **U10** `locked` markers mean "do not prune" and say nothing about the worktree being live; this is git behaviour from knowledge, not verified here. A locked worktree is read (AC-QWR-004).
- **U11** A stray record directory `<root>/.claude/worktrees/.moai/state/context-usage/` exists on disk (seen in a listing, stale); no investigation of how it was created (writer domain, out of scope).
- **U12** The CI runner has `git` on `PATH` for `real_git_worktree_layout` (the workflow file was grepped for the test command only).

## §H Risks and anti-patterns

- **Reading a worktree that is not checked out any more.** Mitigated by the missing-target skip (REQ-QWR-003).
- **Quietly widening scope.** The three neighbours F3-F5 stay out except the test-only F3 pin (D7); the writer anchor is not touched (REQ-QWR-010).
- **A refactor that changes the single-directory result.** AC-QWR-003 compares the old function against the new reading on the same fixtures and the old `TestQAS_` suite runs unedited.
- **Time-based or machine-load-based assertions.** None: the bound and the layouts are count-based; the timing observation is a data point.
- **Measuring `next` by leasing.** Prohibited as a default (§C).

## §I Cross-references

`spec.md` (REQ-QWR-001..010, D1-D8), `acceptance.md` (AC-QWR-001..012, §E ledger), `decision-index.md`; predecessor `.moai/specs/SPEC-QUOTA-AWARE-SCHEDULING-001/` (progress.md §E.4 debt D1, §J), `.moai/reports/t1347/sync-audit.md` (F1, F2), `.claude/rules/moai/development/verification-completeness.md` §1.3 and §2, `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 and §2.3.
