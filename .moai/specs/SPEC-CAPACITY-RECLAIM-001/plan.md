# SPEC-CAPACITY-RECLAIM-001 — Implementation Plan

## §A Context

Card t1631 (plan phase) at worktree `WT-capacity-reclaim`. The code measurements in this plan are pinned to tree `e2658f30127375e75e0c606d09ca97be2c551cf2` (short `e2658f3012`), the merge of local `main` into the card branch (689 commits absorbed, zero conflicts). Of the paths this plan measures, only `internal/config/defaults.go` and the template `workflow.yaml` differ from the iter3-audited tree `ab9f5a6d3` (`git diff --name-only ab9f5a6d3 e2658f3012 -- internal/` lists 218 files; those two are the only measured ones among them), and every pin below was re-observed on `e2658f3012` in the v0.2.0 revision session. The commit that carries this revision changes only SPEC artifacts. Tier M: 10 requirements (REQ-CAP-001 to REQ-CAP-010) and 13 acceptance criteria (AC-CAP-001 to AC-CAP-013), each counted independently against the Tier M ceilings of 16. The requirements and measured baseline are in `spec.md` §B and §C. Milestones are ordered by decision reversibility: the decision-gated items first, then the baseline and the tracking change, then the profile-cap conversion, then evidence and sync. No time estimates are given; the order is by phase.

Scope note (operator decision, 2026-10-10): the worktree-sweep, worktree-list, and bulk-run milestones of earlier revisions are split to card B, and the doctor read-path to card D; no milestone below touches `internal/cli/worktree` or `internal/cli/doctor_disk.go`.

## §B Milestones (priority-ordered)

### M1 — Priority High: decision-gated items (Q4 ordering guarantee; Q1 split default; SessionStart split test)

Q4 (operator verdict 8a62: the default stands). The rotation alternate is not approved, so no usage-log rotation work exists in this SPEC: no declared sink directory, no displacement, no pass-lock order, and no change to `internal/hygiene/sinks.go` (the owner exceptions at `:90–91` stay, and no registry entry is added). The ordering guarantee of REQ-CAP-004 is held by the existing prune, which archives stale events before it rewrites the log (`internal/harness/retention.go:416→421`). M1 changes neither the prune nor its lock, and does not retire the prune. M1 verifies the guarantee it relies on: write `TestRetentionPruneArchivesBeforeRewrite` in `internal/harness` (AC-CAP-009) — RED is its absence premise (EV-006) recorded before the test exists, the test's mutant observation (a prune reordered to rewrite before it archives fails the byte-identical assertion) is recorded at pre-flight, and the test lands green against the unchanged prune.

Q1 (operator verdict 8a62: split default). The approved split is implemented as follows:

- Change the shipped default of `workflow.hygiene.mode` from `report` to `apply` in `internal/template/templates/.moai/config/sections/workflow.yaml` (line 331; the settings block is `:331–333`). The mode governs both units of the SessionStart auto path, the audit-log rotator and the GC unit (`internal/hook/session_start_hygiene.go:37`, `:49`; `internal/hygiene/engine.go:96–109`). None of the eight registered GC target classes (`internal/hygiene/targets.go:29–36`) is named worktrees, databases, or run records, so the destructive categories stay outside the SessionStart path by construction; the home-cleanup path (`moai clean --home --force`, `internal/cli/clean.go:107`) is not invoked by the SessionStart pass, so backups and releases stay report-only there. Template-First: edit the template, then build and sync per the project template cycle. AC-CAP-008 verifies the flip (RED EV-007: `mode: report` at `e2658f3012`).
- Add `TestSessionStartAutoPathSplit` in `internal/hook` (AC-CAP-012): with `workflow.hygiene.mode: apply`, the SessionStart pass through the `hygieneRunFn` seam (`session_start_hygiene.go:19`) writes `mode: apply` rows for the rotator and GC units, and no row names a path outside `.moai/logs` or `.moai/state` — no destructive category, no backups or releases path; and the profile-cap advisory check (M3's loader and warning) fires in the pass for a fixture home with the key set and stays silent for the absent-key fixture (REQ-CAP-009's hook surface). RED is its absence premise (EV-008) recorded at pre-flight, before the test exists.
- The verdict's usage-log rotation is the retention prune under the Q4 default. The prune runs from the harness observer spawn gate (`internal/cli/hook.go:167–183`, `:827–838`, `:1322–1335`), not from the SessionStart auto path (`internal/hook/session_start_record.go:115–131` prunes session records only). M1 adds no SessionStart trigger for it.
- The CLI `--apply` switch is unchanged (REQ-CAP-006).

Lessons inbox (operator verdict 8a62). `.moai/lessons-inbox.jsonl` is outside this SPEC's rotation scope, and REQ-CAP-005 stays as written. Its rotation moves to card t1653 together with the drain-offset design. M1 does not rotate, register, or untrack the inbox's owner code, and neither item changes `internal/hook/inbox_lifecycle.go` or the default of the CLI `--apply` switch. (The M2 untracking of the inbox file itself is the tracking change, not a rotation.)

### M0 — Priority High: baseline commit before the tracking change (precedes the M2 commit)

- Record the baseline in `progress.md` §E.2: the tracked-bytes total (the pipeline of §C), the byte counts of `.moai/harness/usage-log.jsonl` and `.moai/lessons-inbox.jsonl`, the SHA-256 of the three `.corrupt` archives and of the six monthly `.jsonl.gz` archives (`shasum -a 256`), and the `git ls-files` listing of the eleven files. Each entry carries its command and its verbatim output (`verification-claim-integrity.md` §2).
- Commit this alone as baseline commit B0 (`docs(SPEC-CAPACITY-RECLAIM-001): M2 pre-change baseline (card t1631)`). B0 touches only `progress.md`. The reports directory is not used for the baseline, because the repository ignores it (`.gitignore:235`, `.moai/reports/*`).
- Ordering (verification-claim-integrity §2.3): B0 is an ancestor of the untracking commit B1 in the commit graph. B1 carries the M2 changes. AC-CAP-013 verifies the order.

### M2 — Priority High: template ignore and untracking (Q3 satisfied; the landing follows the card path)

- Add four rules to `internal/template/templates/.gitignore`: `.moai/harness/usage-log.jsonl`, `.moai/lessons-inbox.jsonl`, `.moai/**/*.corrupt*`, and `.moai/harness/learning-history/archive/*.jsonl.gz` (REQ-CAP-001). Template-First: edit the template, then build and sync per the project template cycle.
- Extend `TestTemplateGitignoreIgnoresLocalArtifacts` (`internal/template/gitignore_local_artifacts_test.go:29`) with the cases of AC-CAP-003. Run it once before the template edit (RED) and once after (GREEN).
- Untrack the eleven tracked files with `git rm --cached`: the two state files and the nine files under `.moai/harness/learning-history/archive/` (three `.corrupt`, six `.jsonl.gz`) (REQ-CAP-002, REQ-CAP-003, REQ-CAP-010). Local copies stay in place. Commit as B1, after B0.
- Q3 (decision-index) is satisfied: the per-clone measurement recipe was executed on 2026-10-10 and recorded at `.moai/reports/t1631/q3-clone-measurement.md`, and the operator's narrowed-scope decision (2026-10-10) authorized the untracking. B1 lands through the card's normal landing path (integration merge; the lane does not push).

### M3 — Priority Medium: profile cap warning-only conversion and shipped default (unconditional code)

- Backups (REQ-CAP-008; assumption A12): in category 4 of `scanHomeCleanable` (`internal/cli/clean_home.go:310–331`), sort the `backups/removed-*` directories by modification time, protect the three newest, and select the others only when they are also older than the backup retention window. At SessionStart this category is report-only (operator verdict 8a62).
- Releases (REQ-CAP-008; assumption A10): no change to the rule. `scanReleaseCandidates` already keeps the current release and the `releaseKeep` newest non-current release binaries (`clean_home.go:408–445`; `internal/config/defaults.go:456`). The requirement's wording states that rule. At SessionStart this category is report-only.
- Profile cap (REQ-CAP-009; assumption A11; warning only, operator verdict 8a62): remove the comparison at `clean_home.go:243` as a candidate source, so the cap selects no `projects/` entry for deletion. Add a loader for `state.profile_max_bytes` modeled on `loadHomeRetentionDays` (`clean_home.go:101–121`) without its compiled fallback; an absent key yields no warning. The loaded value feeds a warning in the SessionStart hook's deferred pass, time-boxed as an advisory check (`.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline). The warning names the profile and the key, and nothing is deleted. AC-CAP-011's test `TestProfileCapWarnsWithoutDeleting` covers the loader, the warning, the absent-key case, and no-deletion. `internal/cli/doctor_disk.go` is not touched in this SPEC (card D).
- Shipped default: add `profile_max_bytes: 8589934592` under `state:` in the template `internal/template/templates/.moai/config/sections/state.yaml` (today `state: {}` at line 13), then build and sync per Template-First. Pre-flight confirms that the home-tier file read by `clean_home.go` is seeded from this template (A11).
- Extend `TestSessionStartAutoPathSplit` with its clause 3 (the profile-cap advisory check in the pass) once the loader and warning exist; the extended test passes at this milestone's commit (AC-CAP-012).
- Confirmation surface: the apply run for backups and releases is `moai clean --home --force` (`internal/cli/clean.go:106`; assumption A13).

### M4 — Priority Low: evidence recording and sync

- Record every AC command and its verbatim output in `progress.md` §E.2 during the run, with the exit-capture rule of `acceptance.md` §0 applied as written. Record the RED cells and the mutant observations before the first implementation commit. The sync phase records §E.4.

## §C Pre-flight commands (run phase)

Measurement commands. Pipelines are allowed here because these are measurements, not the single-invocation RED cells of `acceptance.md` §D.0; each is run as one command, and its verbatim output is recorded in `progress.md` §E.2.

```bash
git rev-parse HEAD
git status --short
git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl | xargs -0 wc -c | tail -1
git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl .moai/harness/learning-history/archive
grep -n -E 'usage-log|lessons-inbox' internal/hygiene/sinks.go
grep -rn -E 'DefaultProfileMaxBytes' internal
grep -n -E 'mode: (report|apply)' internal/template/templates/.moai/config/sections/workflow.yaml
grep -rn 'TestRetentionPruneArchivesBeforeRewrite' internal
grep -rn 'TestSessionStartAutoPathSplit' internal
grep -rn 'profile_max_bytes' internal/hook
```

Exit capture. Observed runner contract of this session's audit runner, stated so a cell can satisfy it (verification-completeness.md §2.1): the runner prints a command's stdout verbatim and prints no status line for a successful run; it renders a bare single invocation that exits non-zero as `Exit code N` (observed on a false `git merge-base --is-ancestor` pair); and it renders an empty-output run as `(Bash completed with no output)` with no status line, including for a no-match grep whose documented convention is exit 1 (observed twice). A cell therefore records: the verbatim stdout; the exit code as a runner observation only where the runner printed `Exit code N`; and otherwise the exit value labelled as the command's documented convention, with a note of what the runner did and did not show. No cell wraps its command (no `bash -c`, no `; echo $?`); the single-invocation form of `verification-completeness.md` §2.1 is never modified to capture an exit.

Pre-flight also observes the RED cells that need a fixture (AC-CAP-002) and the test-level RED of AC-CAP-009's mutant, AC-CAP-011, and AC-CAP-012 (the named tests are written, then run against the pre-change tree, and their failing run is that pre-flight record). A criterion whose RED cannot be observed stays unadopted (`acceptance.md` §0).

## §D Technical approach

### §D.1 Retention ownership and the ordering guarantee (M1)

- Under the Q4 default nothing about the usage-log mechanism changes: `internal/harness/retention.go` stays the owner, `sinkRegistry` gains no entry, and the owner exceptions stay (`sinks.go:90–91`).
- The guarantee REQ-CAP-004 states is a property of the existing code: `prune` partitions the log (`retention.go:404`), calls `archiveEvents` (`:416`), and calls `overwriteWithEvents` (`:421`) only after the archive call returns without error. `TestRetentionPruneArchivesBeforeRewrite` pins that property with two observations: (i) a fixture log with stale and fresh events pruned into a writable fixture archive leaves the stale events in `archiveDir/<YYYY-MM>.jsonl.gz` and the rewritten log with only the fresh events; (ii) with the archive write made to fail (a fixture archive path that cannot be written), the log file is byte-identical to before the call — the rewrite never happens on a failed archive. The mutant is a reorder: a prune variant that calls `overwriteWithEvents` before `archiveEvents` fails (ii). The test drives the retention package directly on fixture paths; it does not need a lock fixture, because `pruneLocked` is reached through the same code path the production caller uses.

### §D.2 Template ignore test (AC-CAP-003)

- The existing test writes the embedded template into a `t.TempDir()` repository, runs `git init`, and asks `git check-ignore -q --no-index -- <path>` (`internal/template/gitignore_local_artifacts_test.go:46`, `:85`). That isolates the template from the repository's own `.gitignore`, which in this worktree already matches `.moai/harness/usage-log.jsonl` (line 352) and `.moai/reports/*` (line 235). A probe run in the worktree cannot isolate the template, so it is not used.
- A comment-only edit fails the positive cases, which is the mutant the criterion requires.

### §D.3 Untracking (M2)

- `git rm --cached` removes the index entries and keeps the working copies. The prune and the hooks keep writing the local `usage-log.jsonl` and `lessons-inbox.jsonl`; the local archive files stay as they are. No lock changes.
- The `.corrupt` payloads and the monthly archives stay on disk unchanged (AC-CAP-005 hashes them before and after).

### §D.4 SessionStart split behavior (M1)

- The SessionStart auto path is `runSessionHygiene` (`internal/hook/session_start_hygiene.go:28–59`), invoked best-effort through the `hygieneRunFn` seam (`:19`). It loads `workflow.hygiene` settings, takes the mode from `workflow.hygiene.mode` alone (`:37`), constructs the engine (`:38`), and runs both units (`:49`); every failure is logged and swallowed.
- With `mode: apply`, the rotator writes `mode: apply` rows for each registered sink decision and the GC writes its report (REQ-CAP-007). The test drives the seam or `runSessionHygiene` directly on a fixture project and asserts: apply rows exist for both units, and no row or report names a path outside the project's `.moai/logs` and `.moai/state` — which holds by construction because the GC's closed class registry (`internal/hygiene/targets.go:29–36`) contains no worktree, database, or run-record class and reads no home directory, and the home-cleanup path is reachable only through `moai clean --home` (`internal/cli/clean.go:107`).
- The usage-log prune is not part of this pass: it is the hidden `retention-prune` child spawned by the harness-observe gate (`internal/cli/hook.go:167–183`, `:827–838`), and the SessionStart record path prunes session records only (`internal/hook/session_start_record.go:115–131`).

### §D.5 Home cleanup and profile warning (M3)

- The apply run is `runCleanHome(p, force)` with `force` true (`internal/cli/clean_home.go:530`, reached from `clean.go:106`). Backups, releases, and profile candidates are computed by `scanHomeCleanable` (`clean_home.go:160`), which the new unit tests call directly with a fixture root.
- The profile-cap loader reads `paths.UserConfigSectionsDir()`, the same directory that `loadHomeRetentionDays` reads (`clean_home.go:101`). Tests set `HOME` with `t.Setenv` so that the fixture root is the only root touched. The loader has no compiled fallback: an absent key yields no warning (REQ-CAP-009).
- The warning lives in the SessionStart deferred pass as an advisory check; it names the profile and the key and deletes nothing. `internal/cli/doctor_disk.go` is untouched here (card D).

### §D.6 Lint and gate compatibility

- The new loader and the warning check are covered by new tests in their packages; the existing `internal/hygiene` tests and the registry-completeness guard are unaffected because no registry entry changes. Those tests must pass in the same milestone as the change.
- `progress.md` keeps its top-level `status: draft` line (audit observation O3, not changed). The sync closer rewrites the first `status:` line of `progress.md` when the SPEC closes (`internal/spec/closer.go`, `rewriteProgressStatusCompleted`), so removing the line would change the close path. The spec audit reads `status:` from `spec.md` only. The duplicate key is recorded for the sync phase to reconcile.

## §E Risks and mitigations

| Risk | Mitigation |
|---|---|
| Untracking changes working copies in clones that pull | Q3 measurement recorded (`.moai/reports/t1631/q3-clone-measurement.md`); B1 lands through the card's normal landing path and is not pushed by the lane. Git's pull-time behavior for a locally modified copy is not measured here |
| Git-tracked margin is about 4.70 MB of headroom | Track the total in every run; new tracked harness files require review |
| The archive untracking widens the clone-impact surface beyond the Q3 measurement | The archives are append-only retention output whose authoritative copies live where they were produced (spec §F, A14) |
| Apply-by-default changes every user's auto-path behavior | Decided by verdict 8a62 (Q1 split); the closed GC registry keeps destructive categories out of the SessionStart path by construction |
| Lessons-inbox offset breaks under rotation | REQ-CAP-005 excludes rotation; the rotation itself is card t1653's |
| The profile cap is absent for users who never set the key | No warning exists in that case (REQ-CAP-009); new templates ship the default; the residual is accepted by the warning-only verdict (spec §F) |
| Backups retention deletes a backup younger than the window | Floor reading (A12): removal requires both "beyond the newest three" and "older than the window" |

## §F Exclusions

See `spec.md` §D (Out of Scope).
