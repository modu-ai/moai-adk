# SPEC-CAPACITY-RECLAIM-001 — Implementation Plan

## §A Context

Card t1631 (plan phase) at worktree `WT-capacity-reclaim`. The code measurements in this plan are pinned to the audited tree `40cc1031c2b18781377f1bdf2c590cd39f08244e`, the commit on which plan-audit iteration 1 ran. The commit that carries this revision changes only SPEC artifacts, so the code paths measured here are unchanged by it. Tier M: 14 requirements (REQ-CAP-001 to REQ-CAP-014) and 16 acceptance criteria (AC-CAP-001 to AC-CAP-016), each counted independently against the Tier M ceilings of 16. The requirements and measured baseline are in `spec.md` §B and §C. Milestones are ordered by decision reversibility: decision-gated items first, then the baseline and the tracking change, then the code predicates, then the operator-gated bulk run, then the home cleanup. No time estimates are given; the order is by phase.

## §B Milestones (priority-ordered)

### M1 — Priority High: decision-gated items (Q4 default; Q1 split)

Q4 (operator verdict 8a62: the default stands). The rotation alternate is not approved, so none of the usage-log rotation steps of earlier revisions is built under this SPEC: no declared sink directory, no displacement, no pass-lock order, and no change to `internal/hygiene/sinks.go` (the owner exception at `:91` stays, and no registry entry is added). The ordering guarantee of REQ-CAP-004 is held by the existing prune, which archives stale events before it rewrites the log (`internal/harness/retention.go:415–423`). M1 changes neither the prune nor its lock, and does not retire the prune.

Q1 (operator verdict 8a62: split default). Applies only when decision Q1 carries an approved verdict (REQ-CAP-006; AC-CAP-009). The approved split is implemented as follows:

- Change the shipped default of `workflow.hygiene.mode` from `report` to `apply` in `internal/template/templates/.moai/config/sections/workflow.yaml` (line 318). The mode governs both units of the auto path, the audit-log rotator and the GC unit (`internal/hygiene/engine.go:96–108`). None of the eight registered GC target classes (`internal/hygiene/targets.go:29–36`) is named worktrees, databases, or run records, so the destructive categories stay outside the SessionStart path by construction. Pre-flight confirms that the GC residue classes are the category the verdict calls cache.
- The verdict's usage-log rotation is the retention prune under the Q4 default. The prune runs from the harness observer spawn gate (`internal/cli/hook.go:169–174`, `:1323–1332`), not from the SessionStart auto path (`internal/hook/session_start_record.go:115–131` prunes session records only). M1 adds no SessionStart trigger for it.
- The CLI `--apply` switch is unchanged (REQ-CAP-006).

Lessons inbox (operator verdict 8a62). `.moai/lessons-inbox.jsonl` is outside this SPEC's rotation scope, and REQ-CAP-005 stays as written. Its rotation moves to card t1653 together with the drain-offset design. M1 does not rotate, register, or untrack the inbox, and neither item changes `internal/hook/inbox_lifecycle.go` or the default of the CLI `--apply` switch.

### M0 — Priority High: baseline commit before the tracking change (precedes the M2 commit)

- Record the baseline in `progress.md` §E.2: the tracked-bytes total (the pipeline of §C), the byte counts of `.moai/harness/usage-log.jsonl` and `.moai/lessons-inbox.jsonl`, the SHA-256 of the three `.corrupt` archives (`shasum -a 256`), and the `git ls-files` listing of the five files. Each entry carries its command and its verbatim output (`verification-claim-integrity.md` §2).
- Commit this alone as baseline commit B0 (`docs(SPEC-CAPACITY-RECLAIM-001): M2 pre-change baseline (card t1631)`). B0 touches only `progress.md`. The reports directory is not used for the baseline, because the repository ignores it (`.gitignore:235`, `.moai/reports/*`).
- Ordering (verification-claim-integrity §2.3): B0 is an ancestor of the untracking commit B1 in the commit graph. B1 carries the M2 changes. AC-CAP-016 verifies the order.
- If M1 applies (Q4 approved), take B0 after M1, so that the displacement is part of the baseline.

### M2 — Priority High: template ignore and untracking (Q3 gates the push)

- Add three rules to `internal/template/templates/.gitignore`: `.moai/harness/usage-log.jsonl`, `.moai/lessons-inbox.jsonl`, and `.moai/**/*.corrupt*` (REQ-CAP-001). Template-First: edit the template, then build and sync according to the project template cycle.
- Extend `TestTemplateGitignoreIgnoresLocalArtifacts` (`internal/template/gitignore_local_artifacts_test.go`) with the cases of AC-CAP-004. Run it once before the template edit (RED) and once after (GREEN).
- Untrack the five files with `git rm --cached`: the two state files and the three `.corrupt` archives (REQ-CAP-002, REQ-CAP-003). Local copies stay in place. Commit as B1, after B0.
- Gate: decision Q3 must carry an operator verdict before B1 is pushed (plan §E, `decision-index.md` Q3). Without a verdict B1 stays a local commit.

### M3 — Priority High: sweep predicates and report fields; `moai worktree list` (unconditional)

Sweep (`internal/cli/worktree/sweep.go`):

- Add the fields `owner`, `owner_state`, and `unpushed` to `sweepVerdict` (`sweep.go:72–96`), to the text report (`renderSweepReport`, `sweep.go:593–622`), and to the JSON record (`renderSweepJSON`, `sweep.go:584–588`).
- Add the unpushed predicate immediately after the landing predicate (after `sweep.go:480`). The count is `git rev-list --count <tip> --not --remotes`. A count above zero sets the verdict to PRESERVE with cause `unpushed-commits`. A count git cannot produce preserves the tree.
- Add the owner predicate as the last predicate: after the process-cwd predicate (after `sweep.go:518`) and before the DISPOSE verdict (`sweep.go:522`). Owner state `done` is required; any other state preserves the tree with cause `owner-unconfirmed`; an unreadable factory store preserves it with cause `owner-undetermined` (§B.2 of `spec.md`).
- The owner source is a new read-only lookup over the card store, keyed by `worktree_path`. It follows the binding order of `PreviousCardWorktree` (`internal/homestate/card_worktree.go:73–90`), and it opens the store with `OpenFactoryReadonly` (`internal/homestate/factory.go:213`) on `FactoryDBPath` (`internal/homestate/paths.go:345`). Expose it through a package-level seam in the style of the existing ones (`sweep.go:102`, `:108`, `:147`, `:195`), so that tests can stub it.
- Canonicalize both sides of the path comparison, as the existing containment check does (`internal/cli/worktree/hoist.go:119`).
- Add the new causes (`owner-unconfirmed`, `owner-undetermined`, `unpushed-commits`) to the `cause=` vocabulary (`sweep.go:48–64`).
- `applySweepVerdicts` (`sweep.go:634–702`) needs no change. It removes only DISPOSE records (`sweep.go:638–640`), and the new predicates reach it through `classifySweepVerdicts` (`sweep.go:379–403`), the single evaluation shared by the report, `--json`, and `--yes` (REQ-WS-012).

`moai worktree list`:

- A new `list` subcommand under `internal/cli/worktree` (REQ-CAP-013). Table output and `--json` output with the keys `path`, `branch`, `size_kib`. The source is `WorktreeProvider.List()` (`sweep.go:337`). Sizes are KiB, the unit of `du -sk`. A detached record carries an empty `branch`.
- AC-CAP-012 and AC-CAP-014 invoke the binary built from this tree by path (`verification-claim-integrity.md` §2.2), not an installed `moai`.

### M4 — Priority Medium: bulk run with evidence rows (Q2-gated execution)

- Run only after the candidate report from M3 has been reviewed and decision Q2 carries an operator verdict (REQ-CAP-012, REQ-CAP-014). Under the default or a rejection, M4 does not run (`progress.md` and `decision-index.md` Q2 branch).
- The `--yes` run writes one evidence row for each disposal before the removal calls at `sweep.go:676` (L1) and `sweep.go:684` (L2), inside the tier switch (`sweep.go:673–693`). The evidence path is fixed in pre-flight and lies under `.moai/state/`, not `.moai/logs/`, so the sink-registry guard does not classify it as an unregistered logs writer.
- Record each removed path, its verdict fields, and the evidence row's position before the removal in `progress.md` §E.2.

### M5 — Priority Medium: backups floor, profile warning, and shipped default (unconditional code; the apply run is operator-invoked with `--force`)

- Backups (REQ-CAP-008; assumption A13): in category 4 of `scanHomeCleanable` (`internal/cli/clean_home.go:310–331`), sort the `backups/removed-*` directories by modification time, protect the three newest, and select the others only when they are also older than the backup retention window. At SessionStart this category is report-only (operator verdict 8a62).
- Releases (REQ-CAP-008; assumption A10): no change to the rule. `scanReleaseCandidates` already keeps the current release and the `releaseKeep` newest non-current release binaries (`clean_home.go:291–293`, `:408–445`; `internal/config/defaults.go:451`). The requirement's wording states that rule. At SessionStart this category is report-only.
- Profile cap (REQ-CAP-009; assumption A11; warning only, operator verdict 8a62): remove the comparison at `clean_home.go:243` as a candidate source, so the cap selects no `projects/` entry for deletion. Add a loader for `state.profile_max_bytes` modeled on `loadHomeRetentionDays` (`clean_home.go:101–121`) without its compiled fallback; an absent key yields no warning. The loaded value feeds a warning in `moai doctor` and in the SessionStart hook, time-boxed as an advisory check (`.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline). The warning names the profile and the key, and nothing is deleted. Delete `DefaultProfileMaxBytes` (`internal/config/defaults.go:455`) when the pre-flight grep (plan §C) shows no reader in `internal`.
- Shipped default: add `profile_max_bytes: 8589934592` under `state:` in the template `internal/template/templates/.moai/config/sections/state.yaml` (today `state: {}`), then build and sync per Template-First. Pre-flight confirms that the home-tier file read by `clean_home.go` is seeded from this template (A11).
- Confirmation surface: the apply run for backups and releases is `moai clean --home --force` (`internal/cli/clean.go:106`; assumption A18).

### M6 — Priority Low: evidence recording and sync

- Record every AC command and its verbatim output in `progress.md` §E.2 during the run. Record the four-element RED cells, with the exit code captured by the runner, before the first implementation commit. The sync phase records §E.4.

## §C Pre-flight commands (run phase)

Measurement commands. Pipelines are allowed here because these are measurements, not the single-invocation RED cells of `acceptance.md` §D.0; each is run as one command, and its verbatim output is recorded in `progress.md` §E.2.

```bash
git rev-parse HEAD
git status --short
git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl | xargs -0 wc -c | tail -1
git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*'
grep -n -E 'usage-log|lessons-inbox' internal/hygiene/sinks.go
grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go
grep -rn -E 'DefaultProfileMaxBytes' internal --include='*.go'
grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go
grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'
git worktree list --porcelain | grep -c '^worktree '
```

Exit codes. The plan session could not capture the exit code of a command that prints nothing. Pre-flight records the exit code for each RED command as its own field (`verification-completeness.md` §2.1).

Pre-flight also observes the RED cells that need a fixture (AC-CAP-003, AC-CAP-007, AC-CAP-010) and the fixture-free RED of AC-CAP-004 (the extended test, before the template edit). A criterion whose RED cannot be observed stays unadopted (`acceptance.md` §0).

## §D Technical approach

### §D.1 Sink rotation and serialization (M1)

- The rotator keeps one pass lock per pass and one rotation decision per sink. Usage-log gains a declared directory, a pass-lock order (pass lock, then prune-state lock), and the same inspection rules as the prune.
- The displaced chunk is discarded at the next rotation (`rotate.go:241–245`). This is the Q4 consequence: events in the chunk that the prune has not yet archived do not reach the monthly archive. The plan does not change it; A17 and the Q4 row state it.
- Appenders take no lock, as for every registered sink (`internal/hygiene/sinks.go:169`, `appendText`). An event written to the displaced inode during a rotation lands in the chunk. This is the residual window that SPEC-MOAI-HYGIENE-001 already accepts for every sink.

### §D.2 Template ignore test (AC-CAP-004)

- The existing test writes the embedded template into a `t.TempDir()` repository, runs `git init`, and asks `git check-ignore -q --no-index -- <path>` (`internal/template/gitignore_local_artifacts_test.go:46–97`). That isolates the template from the repository's own `.gitignore`, which in this worktree already matches `.moai/harness/usage-log.jsonl` (line 352) and `.moai/reports/*` (line 235). A probe run in the worktree cannot isolate the template, so it is not used.
- A comment-only edit fails the positive cases, which is the mutant the criterion requires.

### §D.3 Untracking (M2)

- `git rm --cached` removes the index entries and keeps the working copies. The prune and the hooks keep writing the local `usage-log.jsonl` and `lessons-inbox.jsonl`. No lock changes.
- The `.corrupt` archives stay on disk unchanged (AC-CAP-006 hashes them before and after).

### §D.4 Sweep predicates (M3, M4)

- One evaluation feeds the report, `--json`, and `--yes` (`classifySweepVerdicts`, `sweep.go:379–403`). Adding predicates to `sweepEvaluate` therefore changes all three paths at once, and no path can carry a stale predicate.
- The predicate order is: anchored (live session, existing), detached HEAD, on-base, landed (existing), unpushed (new), dirty, ignored, process cwd (existing), owner (new). Each new predicate answers "not checked" for every tree preserved earlier in the chain, as the existing chain does.
- The evidence row of M4 is written before the removal, in the same loop that removes the tree.

### §D.5 Home cleanup (M5)

- The apply run is `runCleanHome(p, force)` with `force` true (`internal/cli/clean_home.go:530–585`, reached from `clean.go:94`). Backups, releases, and profile candidates are computed by `scanHomeCleanable` (`clean_home.go:160`), which the new unit tests call directly with a fixture root.
- The profile-cap loader reads `paths.UserConfigSectionsDir()`, the same directory that `loadHomeRetentionDays` reads. Tests set `HOME` with `t.Setenv` so that the fixture root is the only root touched.

### §D.6 Lint and gate compatibility

- The new `openStateFile` reuse and the new declared directory are covered by the existing `internal/hygiene` tests and by the registry-completeness guard. Those tests must pass in the same milestone as the change.
- `progress.md` keeps its top-level `status: draft` line (audit observation O3, not changed). The sync closer rewrites the first `status:` line of `progress.md` when the SPEC closes (`internal/spec/closer.go:478–490`, `rewriteProgressStatusCompleted`), so removing the line would change the close path. The spec audit reads `status:` from `spec.md` only (`internal/spec/audit.go:100`). The duplicate key is recorded for the sync phase to reconcile.

## §E Risks and mitigations

| Risk | Mitigation |
|---|---|
| Untracking changes working copies in clones that pull | Decision Q3 (per-clone recipe in its row); B1 is not pushed before the verdict. Git's pull-time behavior for a locally modified copy is not measured here |
| Git-tracked margin is about 1.49 MB | Track the total in every run; new tracked harness files require review |
| Bulk sweep disposes unpushed or unowned work | REQ-CAP-011 (owner and unpushed predicates in the shared evaluation); REQ-CAP-012 (evidence before each disposal); decision Q2 |
| Apply-by-default changes every user's auto-path behavior | Decision Q1 FOUNDER; no default flip without verdict |
| Lessons-inbox offset breaks under rotation | REQ-CAP-005 excludes it |
| Reversing a completed SPEC's ownership (HYGIENE-001 §D) for usage-log | Decision Q4 FOUNDER; no registry or owner-exception change without verdict |
| The 30-day prune and the rotation interleave | Serialized on the prune-state lock (M1); lock order stated; prune not retired |
| A displaced chunk discards events the prune has not archived | Q4 row states the consequence; A17 |
| The profile cap is absent for users who never set the key | Notice names the key (A11); the template-shipping question is recorded for the operator |
| Backups retention deletes a backup younger than the window | Floor reading (A13): removal requires both "beyond the newest three" and "older than the window" |

## §F Exclusions

See `spec.md` §D (Out of Scope).
