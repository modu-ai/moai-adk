# SPEC-CAPACITY-RECLAIM-001 — Acceptance Criteria

This file is the verification layer. Each criterion is Given-When-Then with a binary test and the command that measures it. Baselines are the verbatim outputs recorded in `spec.md` §B.

**Tree pin.** The document-level pin is the audited tree `40cc1031c2b18781377f1bdf2c590cd39f08244e` (short `40cc1031c`) on branch `WT-capacity-reclaim`, observed in the plan session. Every code observation in §D.0 was taken on this pin. The plan commit that carries this file changes only SPEC artifacts. A criterion that needs a fixture run, or a commit that does not exist yet, carries no RED cell and is unadopted until pre-flight observes it.

## §0 Classes, adoption, and command form

Each criterion belongs to one class, which states what its result means for the release gate:

- **RB — release-blocking.** The criterion gates completion. Its RED-now cell is a single-invocation read-only command (no pipe, redirection, `&&`, `;`, or subshell; `verification-completeness.md` §2.1), recorded with its verbatim stdout, its exit code (captured by the runner, not by the command), and the tree SHA. Its green-path cell names the milestone that flips it and the passing output.
- **RG — regression guard.** Green on the pinned tree by design. It must hold at completion. It is reported as "held", never as a RED-to-GREEN transition.
- **MS — measurement.** The command needs a pipeline or a fixture, so it is outside the single-invocation form. It gates completion through its recorded verbatim output: a missing output, or a failing one, blocks the sync-phase close.

**Adoption** (`verification-completeness.md` §2). A criterion is adopted only when its RED cell is observed. A criterion whose RED cannot be observed by pre-flight is not adopted, is not claimed as passing, and blocks the close until a new plan audit re-scopes it. The plan session did not capture exit codes for commands that print nothing, so no RB criterion is adopted at this revision. Pre-flight records the four-element cells (`plan.md` §C).

**Conditional classes.** AC-CAP-007 and AC-CAP-015 are RB only when decisions Q4 and Q2 carry approved verdicts; otherwise their default branches apply (§D.6). AC-CAP-009 is RG for its unapproved branch and is conditional on Q1 for its approved branch.

## §D.0 Adoption status

| AC | Class | Traces to | RED-now at 40cc1031c | Green path | Status |
|---|---|---|---|---|---|
| AC-CAP-001 | MS | REQ-CAP-002, -003 | Tracked harness and lessons-inbox bytes: `128526661 total` (observed; measured with a pipeline, plan §C) | M2 untracks the five files → residual 3,510,917 bytes (arithmetic), re-measured in the green run | Observed; output recorded at M2 |
| AC-CAP-002 | MS | REQ-CAP-014 (Q2) | Not observed (heavy traversal; pre-flight). The two-root proxy of `spec.md` §B is not this criterion's set | M4 disposal of owner-confirmed candidates → total below 10,485,760 KiB | Unadopted until pre-flight |
| AC-CAP-003 | MS | REQ-CAP-007 | Not observed (fixture apply run). The write sites exist at `internal/hygiene/rotate.go:167` and `:184` | Fixture apply run → `"mode":"apply"` rows | Unadopted until pre-flight |
| AC-CAP-004 | RB | REQ-CAP-001 | The extended test must fail before the template edit. The template has no rule for usage-log, lessons-inbox, or corrupt (observed: no match) | M2 template rules → the test passes | Unadopted until the RED run is recorded |
| AC-CAP-005 | RB | REQ-CAP-002, -003 | `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*'` lists five paths (observed): three `.corrupt` archives, usage-log, lessons-inbox | M2 `git rm --cached` → empty stdout | RED observed; exit code recorded at pre-flight |
| AC-CAP-006 | RG | REQ-CAP-002, -003 | Held: both state files exist; archive hashes recorded in B0 | After B1, the same | Guard; baseline recorded at M0 |
| AC-CAP-007 | RB (Q4) | REQ-CAP-004 | `grep -n -E 'usage-log' internal/hygiene/sinks.go` prints line 91 only: the owner exception, no registry entry (observed). Fixture run not observed | After Q4 approval: registry entry, rotation fixture | Unadopted (Q4); RED by source |
| AC-CAP-008 | RG | REQ-CAP-005 | Held: `lessons-inbox.jsonl` is an owner exception (`sinks.go:90`) | Unchanged | Guard |
| AC-CAP-009 | RG; RB on Q1 | REQ-CAP-006 | `grep -n -E 'mode: (report\|apply)' internal/template/templates/.moai/config/sections/workflow.yaml` prints line 318, `mode: report` (observed) | After Q1 approval: `mode: apply` | Guard (unapproved branch held) |
| AC-CAP-010 | MS | REQ-CAP-008 | Not observed (fixture test). RED by source: backups selected by age only (`internal/cli/clean_home.go:310–331`) | M5 backups floor → the test passes | Unadopted until pre-flight |
| AC-CAP-011 | RB | REQ-CAP-009 | `grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go` prints line 243 (observed) | M5 config key → the grep prints nothing for clean_home.go; the fixture test passes | RED observed; exit code recorded at pre-flight |
| AC-CAP-012 | RB | REQ-CAP-010 | `grep -n -i -E 'owner\|unpushed' internal/cli/worktree/sweep.go` prints nothing (observed): the verdict has neither field | M3 fields in the report and in `--json` | RED observed; exit code recorded at pre-flight |
| AC-CAP-013 | RB | REQ-CAP-011 | `grep -n -i owner internal/cli/worktree/sweep.go` prints nothing (observed): no owner predicate. The unpushed half is preserved today by the landing layers and is not separately red | M3 predicates in the shared evaluation | RED observed for the unowned half; exit code recorded at pre-flight |
| AC-CAP-014 | RB | REQ-CAP-013 | `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` prints nothing (observed): no list verb | M3 `list` subcommand | RED observed; exit code recorded at pre-flight |
| AC-CAP-015 | RB (Q2) | REQ-CAP-012 | `grep -n -i owner internal/cli/worktree/sweep.go` prints nothing (observed); the `--yes` loop writes no evidence row (`sweep.go:634–702`) | After Q2 approval: M4 | Unadopted (Q2); RED by source |
| AC-CAP-016 | MS | REQ-CAP-002, -003 (`verification-claim-integrity.md` §2.3) | Not observable at 40cc1031c: no baseline commit exists yet | B0 is an ancestor of B1, and B0 touches only `progress.md` | Unadopted until B0 exists |

## §D.1 Sizing and plan-phase measurements (AC-CAP-001 to AC-CAP-003)

### AC-CAP-001 — Git-tracked harness and lessons-inbox under 5 MB (MS; maps REQ-CAP-002, REQ-CAP-003)

- **Given** the tree after the untracking commit B1,
- **When** `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl | xargs -0 wc -c | tail -1` is run,
- **Then** the printed total is less than 5,000,000 bytes (`spec.md` assumption A5).
- **Baseline (observed at 40cc1031c):** `128526661 total`.
- **Green path:** M2 untracks the five named files (usage-log 88,629,933 B; lessons-inbox 6,043,267 B; the three archives 4,179,609 + 14,398,585 + 11,764,350 B; total 125,015,744 B). The residual is 3,510,917 B by arithmetic on the measured values, re-measured in the green run.

### AC-CAP-002 — Linked worktrees under 10 GiB (MS; maps REQ-CAP-014; conditional on Q2)

- **Given** decision Q2 carries an approved verdict and the bulk run has executed (M4),
- **When** this pipeline is run from the primary checkout, in the background, because it exceeds the 120 s tool limit on the §B tree:
  `git worktree list --porcelain | grep '^worktree ' | tail -n +2 | cut -c10- | tr '\n' '\0' | xargs -0 du -skc`
- **Then** the exit status is 0, no `du:` diagnostic is printed, and the final `total` line is less than 10,485,760 KiB.
- The command takes each path after the nine-character prefix `worktree ` (so spaces in a path are kept), drops the first record (the primary checkout, A16), and returns `xargs`'s status, which is 123 when any `du` invocation fails. It contains no awk program.
- **Baseline:** not observed with this command (pre-flight). The §B two-root sum (92,373,636 KiB) is a proxy and is not this criterion's set (`spec.md` A16).
- **Green path:** M4 disposal of owner-confirmed candidates. Under the default or a rejection of Q2, the N/A branch applies (§D.6).

### AC-CAP-003 — Apply rows present (MS; maps REQ-CAP-007)

- **Given** a fixture project whose `workflow.hygiene.audit_log_max_bytes` is low enough that a fixture sink is at or above it,
- **When** `moai clean --audit-logs --apply` runs once in the fixture, invoked from the tree build by path,
- **Then** `grep -n '"mode":"apply"' .moai/logs/hygiene-audit.jsonl` prints at least one row for each applied action (rows whose outcome is `rotated`, `deleted`, or `error`).
- **Baseline:** not observed (fixture run). The field is pinned: `AuditRow.Mode` (`internal/hygiene/auditlog.go:16`), set to `apply` at `rotate.go:167` and `:184`.

## §D.2 Tracking and template (AC-CAP-004 to AC-CAP-006, AC-CAP-016)

### AC-CAP-004 — Template ignores the state files (RB; maps REQ-CAP-001)

- **Given** the test `TestTemplateGitignoreIgnoresLocalArtifacts` (`internal/template/gitignore_local_artifacts_test.go`), extended with these cases: `.moai/harness/usage-log.jsonl` (ignored), `.moai/lessons-inbox.jsonl` (ignored), `.moai/harness/learning-history/archive/2026-07.jsonl.gz.corrupt` (ignored), and `.moai/harness/learning-history/archive/2026-07.jsonl.gz` (not ignored: the control for the kept archive),
- **When** `go test ./internal/template/ -run '^TestTemplateGitignoreIgnoresLocalArtifacts$' -count=1` is run,
- **Then** the test passes. The test writes the embedded template into a `t.TempDir()` repository and asks `git check-ignore -q --no-index`, so the repository's own `.gitignore` (which already matches usage-log at line 352 and `.moai/reports/*` at line 235 in this worktree) does not answer for the template.
- **Mutant:** a template whose only addition is a comment line that names the file fails the positive cases.
- **RED-now:** the extended cases fail on the current template, which has no matching rule (observed). The failing run is recorded before the template edit.
- **Green path:** M2 adds the three rules, and the same command passes.

### AC-CAP-005 — Repository stops tracking the state files (RB; maps REQ-CAP-002, REQ-CAP-003)

- **Given** the tree after B1,
- **When** `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*'` is run,
- **Then** stdout is empty.
- **RED-now (observed at 40cc1031c):** five paths are listed: `.moai/harness/learning-history/archive/2026-07.jsonl.gz.corrupt`, `.moai/harness/learning-history/archive/2026-08.jsonl.gz.corrupt`, `.moai/harness/learning-history/archive/2026-09.jsonl.gz.corrupt`, `.moai/harness/usage-log.jsonl`, and `.moai/lessons-inbox.jsonl`.
- **Green path:** M2 `git rm --cached` of the five files (B1) → empty stdout.

### AC-CAP-006 — Local copies survive untracking; archive payloads unchanged (RG; maps REQ-CAP-002, REQ-CAP-003)

- **Given** the baseline recorded in B0 (`progress.md` §E.2: the byte counts of the two state files and the SHA-256 of the three archive `.corrupt` files, each with its command and verbatim output),
- **When** `ls -l .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl` and `shasum -a 256` over the three archive `.corrupt` paths are run after B1,
- **Then** all five files exist; each state file is non-empty when its baseline size was non-zero; and each archive's SHA-256 equals its baseline value.
- The state-file size is not required to stay at or above its baseline. The 30-day prune rewrites `usage-log.jsonl` (`internal/harness/retention.go:421`, `overwriteWithEvents`), so a size decrease is expected behavior, not a loss. Deletion, or truncation to zero, is the failure.
- **Classification:** regression guard, green by design. Mutant: deleting the local copies in place of `git rm --cached` turns the existence check red.

### AC-CAP-016 — The baseline commit precedes the untracking commit (MS; maps REQ-CAP-002, REQ-CAP-003; `verification-claim-integrity.md` §2.3)

- **Given** baseline commit B0 (touching only `progress.md`) and untracking commit B1, with their SHAs recorded in `progress.md` §E.2,
- **When** `git merge-base --is-ancestor <B0> <B1>` is run, and then `git show --name-only --format= <B0>`,
- **Then** the first exits 0, and the second lists only `.moai/specs/SPEC-CAPACITY-RECLAIM-001/progress.md`.
- **RED-now:** not observable at 40cc1031c, because no baseline commit exists. The commit graph is the only witness of this order; the commit message and the session record are not.

## §D.3 Sink rotation and hygiene apply (AC-CAP-007 to AC-CAP-009)

### AC-CAP-007 — usage-log rotates in apply mode (RB when Q4 is approved; maps REQ-CAP-004)

- **Given** decision Q4 in `decision-index.md` carries an approved verdict, and a fixture project whose `.moai/harness/usage-log.jsonl` is at or above the fixture threshold. Record its SHA-256 (H0) and size (S0) before the run,
- **When** `moai clean --audit-logs --apply` runs once in the fixture, invoked from the tree build by path,
- **Then** all of the following hold:
  - (a) `.moai/harness/usage-log.jsonl.1` exists, its SHA-256 equals H0, and its size equals S0 (the content was displaced, not truncated);
  - (b) `.moai/harness/usage-log.jsonl` exists with size 0;
  - (c) `.moai/harness/usage-log.jsonl.2` does not exist;
  - (d) `.moai/logs/hygiene-audit.jsonl` holds a row with `"outcome":"rotated"` whose path ends in `harness/usage-log.jsonl`;
  - (e) with an external process holding an exclusive lock on `.moai/harness/usage-log.jsonl.prune-state` (released by the fixture's cleanup), the same command records an outcome other than `rotated` for that file and leaves the primary byte-identical.
- **Mutant:** a run that leaves the file in place and creates an empty `.1` fails (a), by hash mismatch, and fails (d).
- **Default branch (Q4 default or rejected):** this SPEC changes no line of `internal/hygiene/sinks.go`. The guard `grep -n -E 'usage-log' internal/hygiene/sinks.go` still prints the owner-exception line (`sinks.go:91`) and no registry entry.
- **RED-now (observed at 40cc1031c):** `grep -n -E 'usage-log' internal/hygiene/sinks.go` prints line 91 only. The fixture run is not observed; pre-flight records it.

### AC-CAP-008 — lessons-inbox is not rotated (RG; maps REQ-CAP-005)

- **Given** a fixture copy of `.moai/lessons-inbox.jsonl` at or above the fixture threshold,
- **When** `moai clean --audit-logs --apply` runs once in the fixture,
- **Then** the fixture file's size and SHA-256 are unchanged, and no audit row names `lessons-inbox.jsonl`.
- **Classification:** regression guard, green by design (`internal/hygiene/sinks.go:90`, owner exception). Mutant: adding `lessons-inbox.jsonl` to `sinkRegistry` turns it red.

### AC-CAP-009 — Hygiene default follows decision Q1 (RG; RB on the approved branch; maps REQ-CAP-006)

- **Given** decision Q1 in `decision-index.md`,
- **When** `grep -n -E 'mode: (report|apply)' internal/template/templates/.moai/config/sections/workflow.yaml` is run,
- **Then** the printed value is `apply` when Q1 carries an approved verdict, and `report` otherwise.
- **RED-now (observed at 40cc1031c):** `318: mode: report` (the unapproved branch holds). The approved branch is not observable until Q1 carries a verdict.

## §D.4 Archives, profile, and keep rules (AC-CAP-010, AC-CAP-011)

### AC-CAP-010 — Backups keep the three newest; releases keep the current release plus three (MS; maps REQ-CAP-008)

- **Given** a Go test that calls `scanHomeCleanable(root, retention, config.DefaultReleaseKeep, currentVersion, now)` (`internal/cli/clean_home.go:160`) on a fixture home root with: (i) five `backups/removed-*` directories, all older than the retention window, and six release binaries, the current version and five older non-current ones; (ii) five `backups/removed-*` directories, two of them younger than the retention window and three older,
- **When** `go test ./internal/cli/ -run '^TestHomeCleanKeepFloors$' -count=1` is run,
- **Then** in fixture (i) the candidates are exactly the two oldest backups and the two oldest non-current releases, and the current release is not a candidate; in fixture (ii) the candidates are exactly the two older-than-window backups that lie beyond the newest three.
- **Baseline:** RED by source. Backups are selected by age only (`clean_home.go:310–331`), so fixture (i) yields five backup candidates. The release part already matches (`clean_home.go:441–445`).
- **Green path:** M5 backups floor (`spec.md` A13). Releases are unchanged (A10).

### AC-CAP-011 — Profile data held to the configured cap (RB; maps REQ-CAP-009)

- **Given** a fixture home whose profile root has `projects/` entries totalling 6 MiB, and a fixture `state.yaml` in the home config sections with `state.profile_max_bytes: 1048576`; a second fixture with the key absent. The fixtures set `HOME` with `t.Setenv`,
- **When** `go test ./internal/cli/ -run '^TestHomeCleanProfileCapReadsConfig$' -count=1` is run,
- **Then** in fixture 1 the oldest `projects/` entries become candidates until the profile is at or below 1 MiB; in fixture 2 no `projects/` candidate is produced, and a notice names `state.profile_max_bytes`.
- **RED-now (observed at 40cc1031c):** `grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go` prints line 243. Fixture 1 is RED, because a 6 MiB profile is below the compiled 5 GiB and yields no candidate.
- **Mutant:** a loader with a compiled fallback fails fixture 2.
- **Green path:** M5. The grep then prints nothing for `clean_home.go`.

## §D.5 Worktree report, list, and disposal (AC-CAP-012 to AC-CAP-015)

### AC-CAP-012 — The report carries owner, unpushed, and predicate fields; nothing is removed (RB; maps REQ-CAP-010)

- **Given** the tree build with the M3 change, built with `make build` and invoked by path,
- **When** `git worktree list --porcelain | grep -c '^worktree '` is run before and after `./bin/moai worktree sweep --json`, and the records are read,
- **Then** the sweep's exit status is 0, the two counts are equal, and every non-protected record carries the keys `owner`, `owner_state`, `unpushed`, `landed`, `dirty`, `ignored`, `anchored`, `cwd_occupied`, `on_base`, and `verdict`.
- **RED-now (observed at 40cc1031c):** `grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go` prints nothing. The verdict record has neither field (`sweep.go:72–96`).

### AC-CAP-013 — Unpushed, unowned, and live-session trees are preserved in report and `--yes` (RB; maps REQ-CAP-011)

- **Given** fixtures driven through the existing seams (`sweepFetchBase`, `sweepAncestor`, `sweepProcessCWDs`, `sweepDoneCleanup`; `sweep.go:102–195`) and the new owner seam: (U) a tree whose branch has one local commit on no remote-tracking ref; (W) an otherwise disposable tree (landed, clean, unanchored, no process cwd inside) with no card binding; (L) an otherwise disposable tree anchored by a git worktree lock whose reason names a live pid (the test process's own pid); (C) the control, an otherwise disposable tree bound to a card in state `done`,
- **When** `go test ./internal/cli/worktree/ -run '^TestSweepPreservesUnpushedUnownedLive$' -count=1` runs. The test calls the report evaluation, then the `--yes` apply,
- **Then** U, W, and L are `PRESERVE` in the report, and each directory still exists after the `--yes` apply. C is `DISPOSE` in the report and is removed by `--yes` (non-vacuity).
- **Mutant:** removing the owner check from the evaluation makes W `DISPOSE`, and the test fails.
- **RED-now (observed at 40cc1031c; the unowned half):** `grep -n -i owner internal/cli/worktree/sweep.go` prints nothing: no owner predicate exists. The unpushed half is preserved today by the landing layers (`internal/cli/worktree/landing_predicate.go:395`, `:423–431`) and is not separately red.

### AC-CAP-014 — `moai worktree list` matches git (RB; maps REQ-CAP-013)

- **Given** the tree build, invoked by path,
- **When** `./bin/moai worktree list --json` is run and compared with `git worktree list --porcelain`,
- **Then** the array length equals the count of `worktree` lines; the set of `path` values equals the set of `worktree` paths; the set of `branch` values equals the set of `branch` lines (a detached record has an empty `branch`); each object has a numeric `size_kib`; and the table form prints one row per worktree below its header.
- **RED-now (observed at 40cc1031c):** `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` prints nothing: no list verb is registered.

### AC-CAP-015 — Bulk run disposes only owner-confirmed candidates, with an evidence row first (RB when Q2 is approved; maps REQ-CAP-012)

- **Given** decision Q2 carries an approved verdict, and a fixture with two candidates that pass every other predicate: one bound to a card in state `done`, one bound to a card in state `run`,
- **When** `go test ./internal/cli/worktree/ -run '^TestSweepBulkRunEvidenceFirst$' -count=1` runs the `--yes` apply once,
- **Then** only the `done` candidate is removed, and its evidence row is in the evidence file before its directory is removed.
- **Default branch (Q2 default or rejected):** the bulk run is not executed. AC-CAP-015 is N/A and is reported as N/A (§D.6).
- **RED-now (observed at 40cc1031c):** `grep -n -i owner internal/cli/worktree/sweep.go` prints nothing, and the `--yes` loop (`sweep.go:634–702`) writes no evidence row.

## §D.6 Quality gate and Definition of Done

- **Gate (RB).** Every RB criterion whose condition applies passes, with its four-element RED cell (exit code captured by the runner) and its green-path output recorded in `progress.md` §E.2, before the milestone commit that flips it.
- **Gate (RG).** Every RG criterion whose condition applies holds. A held guard is reported as held, never as a RED-to-GREEN transition.
- **Gate (MS).** Every MS criterion whose condition applies has its verbatim output recorded. A missing output, or a failing MS output, blocks the sync-phase close.
- **Decision outcome branches.** Apply the branch that matches each recorded verdict. The other branches are not required:
  - **Q1 approved:** the AC-CAP-009 approved branch holds (`apply`). **Q1 default (`report`) or rejected:** the AC-CAP-009 unapproved branch holds (`report`). Either way Q1 is satisfied.
  - **Q2 approved:** AC-CAP-015 (RB) and AC-CAP-002 (MS) apply, and M4 runs. **Q2 default (reject) or rejected:** AC-CAP-015 is N/A; AC-CAP-002 is N/A (REQ-CAP-014 does not apply, and the size outcome is recorded as measured, not as a pass); no bulk run executes. AC-CAP-013 applies in both cases, because preservation is unconditional.
  - **Q3 approved:** the M2 commit B1 may be pushed. **No verdict or deferred:** B1 stays a local commit; AC-CAP-001, -005, -006, and -016 are recorded locally; the sync phase records the unpushed status and does not record a push.
  - **Q4 approved:** AC-CAP-007 (RB) applies, and the displaced-chunk consequence is recorded with the verdict. **Q4 default (keep owner) or rejected:** the AC-CAP-007 default branch applies (the owner exception and the absence of a registry entry hold).
- **Plan-phase obligation.** Decisions Q1 to Q4 are not decided by this plan. Their verdict fields stay empty until the operator records them.
- **Close.** The sync phase closes `progress.md` §E.4 only after the gates and branches above are recorded.

## §D.7 Edge cases

- A tree bound to two card records (a replaced run) takes the newest binding (`spec.md` A14).
- A zero-candidate report is reported as "0 candidates"; it is not a pass of the disposal path.
- A tree whose factory store cannot be read is preserved (`owner-undetermined`).
- `usage-log.jsonl` grows and may be rewritten by the 30-day prune during a run. AC-CAP-006 therefore checks existence, non-zero length, and archive hashes, never size equality.
- A branch with an unpushed commit and a merged PR is preserved: the merged PR head must equal the local tip (`landing_predicate.go:395`), and the unpushed predicate preserves it independently.
- A zero-match grep in a RED cell is a recorded red, not a pass. The RED cell records stdout (empty) and the exit code.
