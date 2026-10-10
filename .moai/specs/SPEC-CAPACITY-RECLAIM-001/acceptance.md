# SPEC-CAPACITY-RECLAIM-001 — Acceptance Criteria

This file is the verification layer. Each criterion is Given-When-Then with a binary test and the command that measures it. Baselines are the verbatim outputs recorded in `spec.md` §B. A criterion with no observed red on this tree is marked as such.

**Tree pin.** The document-level pin, which binds every criterion that carries no pin of its own, is card worktree HEAD `2aab5f797` (`git -C <worktree> rev-parse --short HEAD`), observed 10-10 on branch `WT-capacity-reclaim`. The observations below were taken on this pin.

**Adoption.** Each criterion needs two cells: a RED-now cell (observed red on the pinned tree, with its command) and a green-path cell (the milestone that flips it, and the passing output). §D.0 states the status of each. A criterion whose RED-now is not yet observed is unadopted until pre-flight observes it. A criterion that is green on the current tree is a regression guard, not a pass. No criterion here is classified release-blocking, so the four-element cell of `verification-completeness.md` §2.1 is not required in this file.

## §D.0 Adoption status (two-cell pair)

"Source" below means the criterion's subject was observed by a read-only search of the source on this tree, because the criterion's own command cannot run on this tree (the verb or fixture does not exist yet). The criterion's own command is run in pre-flight.

| AC | Traces to | RED-now (10-10, tree `2aab5f797`) | Green path (milestone → passing output) | Status |
|---|---|---|---|---|
| AC-CAP-001 | REQ-CAP-002, REQ-CAP-003 | RED: `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl \| xargs -0 wc -c \| tail -1` → `128526661 total` | M2 untracks the five named files (125,015,744 B) → residual 3,510,917 B, re-measured in the green run | Adopted |
| AC-CAP-002 | REQ-CAP-012 (via M3, M4) | Not observed with the criterion's own command (heavy traversal; pre-flight). Proxy only: `du -sk` over the two roots → 92,373,636 KiB | M4 disposal of owner-confirmed candidates → total below 10,485,760 KiB | Unadopted until pre-flight |
| AC-CAP-003 | REQ-CAP-007 | Not observed (needs a fixture apply run). Source shows the writes exist (`rotate.go:167`, `:184`), so the criterion may already hold | Fixture apply run → at least one `"mode":"apply"` row per applied action | Unadopted until pre-flight |
| AC-CAP-004 | REQ-CAP-001 | RED: `grep -n -E 'usage-log\|lessons-inbox\|\.corrupt' internal/template/templates/.gitignore` → no output | M2 adds the three ignore rules → each pattern matches at least one rule | Adopted |
| AC-CAP-005 | REQ-CAP-002, REQ-CAP-003 | RED: `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*' \| wc -l` → `5` (expected `0`) | M2 `git rm --cached` of the five files → `0` | Adopted |
| AC-CAP-006 | REQ-CAP-002, REQ-CAP-003 | Green on this tree by design (both files present, counts unchanged): regression guard | Guard holds after M2. Mutant: a deletion in place of `git rm --cached` turns it red | Regression guard |
| AC-CAP-007 | REQ-CAP-004 (gated by Q4) | RED by source: `usage-log.jsonl` is absent from `sinkRegistry` (`internal/hygiene/sinks.go`, from line 26), so the rotator cannot rotate it | After Q4 approval: register the file, remove the owner exception, fixture run → one primary plus at most one kept rotation | Unadopted (Q4) |
| AC-CAP-008 | REQ-CAP-005 | Green on this tree by design (`lessons-inbox.jsonl` is an owner exception, `sinks.go:90`): regression guard | Guard holds. Mutant: registering the file in `sinkRegistry` turns it red | Regression guard |
| AC-CAP-009 | REQ-CAP-006 (gated by Q1) | Unapproved branch: green by design (`internal/template/templates/.moai/config/sections/workflow.yaml:318` reads `mode: report`): regression guard. The approved branch is not observable until Q1 carries a verdict | After Q1 approval: line 318 reads `mode: apply` | Regression guard (unapproved branch) |
| AC-CAP-010 | REQ-CAP-008 | RED by source: backups are retained by age (`internal/cli/clean_home.go:310–329`), not by count | Keep-3 rule for backups and releases under the apply confirmation; fixture run | Unadopted until pre-flight |
| AC-CAP-011 | REQ-CAP-009 | RED by source: the comparison uses the compiled constant (`clean_home.go:243`, `config.DefaultProfileMaxBytes`) | Configuration key read for the cap; fixture run at or below the cap | Unadopted until pre-flight |
| AC-CAP-012 | REQ-CAP-010 | RED: `grep -n -i -E 'owner\|unpushed' internal/cli/worktree/sweep.go` → no output (the report has neither field) | M3 adds the owner and unpushed-count fields to the default report and to `--json` | Adopted |
| AC-CAP-013 | REQ-CAP-011 | RED for the unowned half: no owner predicate (the grep of AC-CAP-012 → no output). The unpushed half is already preserved by the landing predicate (`sweep.go:405–470`) and is not separately observed | M3 owner predicate: an unconfirmed owner → PRESERVE | Adopted |
| AC-CAP-014 | REQ-CAP-013 | RED: `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` → no output, exit 1 (the verb is absent) | M3 adds `moai worktree list`; its `--json` array length equals `git worktree list --porcelain \| grep -c '^worktree '` | Adopted |
| AC-CAP-015 | REQ-CAP-012 (gated by Q2) | RED by source: the `--yes` path has no owner-confirmation check (`sweep.go:230`; `grep -n -i owner` → no match) | M4 owner gate and an evidence row before each disposal → fixture: only the confirmed candidate is disposed | Unadopted (Q2) |

## §D.1 Plan-phase checks (the three checks the plan row requires)

### AC-CAP-001 — Git-tracked harness and lessons-inbox under 5 MB (REQ-CAP-002, REQ-CAP-003)

- **Given** the worktree at HEAD `2aab5f797`,
- **When** `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl | xargs -0 wc -c | tail -1` is run,
- **Then** the printed total is less than 5,000,000 bytes (spec.md Assumption A5).
- **Baseline (re-observed 10-10):** `128526661 total` — RED on this tree.
- **Green path:** M2 untracking removes the five named files (usage-log, lessons-inbox, three archives; 125,015,744 bytes). The residual of other tracked files is 3,510,917 bytes by arithmetic on the measured values, to be re-measured in the green run.

### AC-CAP-002 — Worktrees under 10 GiB (REQ-CAP-012 via M3, M4)

- **Given** the primary checkout with its registered worktrees,
- **When** `git worktree list --porcelain | awk '/^worktree /{print $2}' | xargs du -sk | awk '{s+=$1} END {print s}'` is run,
- **Then** the printed total is less than 10,485,760 KiB (spec.md Assumption A5).
- **Baseline:** not observed with this command (the traversal is heavy; it is run in pre-flight). The proxy `du -sk` over the two worktree roots printed a sum of 92,373,636 KiB (spec.md §B). That proxy is not this criterion's set. The card's 53.6 GB figure is not adopted.
- **Green path:** M4 disposal of owner-confirmed candidates. The run phase measures the baseline before any disposal, so the pass is not inferred from the card's figure.

### AC-CAP-003 — Apply rows present (REQ-CAP-007)

- **Given** a fixture hygiene state that contains at least one file over the rotation threshold,
- **When** `moai clean --audit-logs --apply` runs once against the fixture,
- **Then** `.moai/logs/hygiene-audit.jsonl` holds at least one row with `"mode":"apply"` per applied action.
- **Baseline:** not observed (needs the fixture run). The field is pinned (spec.md §E A2): `AuditRow.mode`, value `apply`.

## §D.2 Tracking and template

### AC-CAP-004 — Template ignores the state files (REQ-CAP-001)

- **Given** the template `.gitignore`,
- **When** `grep -n -E 'usage-log|lessons-inbox|\.corrupt' internal/template/templates/.gitignore` is run,
- **Then** each of the three patterns matches at least one ignore rule.
- **Baseline (re-observed 10-10):** no output — RED on this tree. A zero-match grep is a failed criterion (§D.7).

### AC-CAP-005 — Repository stops tracking the state files (REQ-CAP-002, REQ-CAP-003)

- **Given** the repository after M2,
- **When** `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*' | wc -l` is run,
- **Then** the output is `0`.
- **Baseline (re-observed 10-10):** `5` — RED on this tree (the two state files and the three archive `.corrupt` files).

### AC-CAP-006 — Local copies survive untracking (REQ-CAP-002, REQ-CAP-003) — regression guard

- **Given** the two state files as they stand immediately before M2 (tracked sizes on 10-10: `usage-log.jsonl` 88,629,933 B; `lessons-inbox.jsonl` 6,043,267 B),
- **When** M2 is complete and `ls -la .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl` is run,
- **Then** each file still exists, and its byte count is at least its count immediately before M2. Both files are appended to by hooks, so growth is expected; a decrease is the failure.
- **Classification:** regression guard. It is green on this tree by design, so it carries no RED-now. A deletion performed in place of `git rm --cached` turns it red.

## §D.3 Sink rotation and apply

### AC-CAP-007 — usage-log rotates in apply mode (REQ-CAP-004, gated by decision Q4)

- **Given** decision Q4 in `decision-index.md` carries an approved verdict, and a fixture `.moai/harness/usage-log.jsonl` at or above the threshold,
- **When** `moai clean --audit-logs --apply` runs once after the registration that Q4 authorizes,
- **Then** the directory holds one primary file and at most one kept rotation (REQ-CAP-004), and no file was truncated in place.
- **Baseline:** RED by source (`usage-log.jsonl` is absent from `sinkRegistry`; `sinks.go:91` is an owner exception). The criterion is not runnable unless Q4 is approved.

### AC-CAP-008 — lessons-inbox is not rotated (REQ-CAP-005) — regression guard

- **Given** a fixture copy of `lessons-inbox.jsonl` (not the live file) at or above the threshold,
- **When** `moai clean --audit-logs --apply` runs once,
- **Then** the fixture file is unchanged in size and content hash.
- **Classification:** regression guard. It is green on this tree by design, because the rotator's registry does not list the file (`sinks.go:90`). Mutant: adding `lessons-inbox.jsonl` to `sinkRegistry` turns it red.

### AC-CAP-009 — Hygiene default follows decision Q1 (REQ-CAP-006)

- **Given** decision Q1 in `decision-index.md` with its operator verdict field,
- **When** the template default of `workflow.hygiene.mode` is read (`internal/template/templates/.moai/config/sections/workflow.yaml`, line 318, under the `hygiene:` key at line 317),
- **Then** the value is `apply` when Q1 carries an approved verdict, and `report` otherwise. The `--apply` switch of the CLI is not part of this check.
- **Baseline (re-observed 10-10):** line 318 reads `mode: report`, which matches the unapproved branch. The regression guard holds; the approved branch is not observable until Q1 carries a verdict.

## §D.4 Archives and profile

### AC-CAP-010 — Backups and releases keep three (REQ-CAP-008)

- **Given** a fixture backups directory with more than three entries, a fixture releases directory with more than three non-current entries, and an apply run that the operator has confirmed,
- **When** the apply runs,
- **Then** exactly the three most recent entries of each directory remain (the treatment of the current release is pinned per spec.md §E A10).
- **Baseline:** RED by source. Backups are retained by age (`internal/cli/clean_home.go:310–329`), not by count, so a fixture with more than three recent entries keeps all of them.

### AC-CAP-011 — Profile data within the configured cap (REQ-CAP-009)

- **Given** a fixture profile data directory above the cap read from configuration, and an apply run that the operator has confirmed,
- **When** the apply runs,
- **Then** the directory size is at or below the cap read from configuration.
- **Baseline:** RED by source. The comparison at `clean_home.go:243` uses the compiled constant `config.DefaultProfileMaxBytes` (5 GiB, `internal/config/defaults.go:455`); no configuration key is read at that site.

## §D.5 Worktree sweep, list, and bulk disposal

### AC-CAP-012 — Report mode disposes nothing and carries the four fields (REQ-CAP-010)

- **Given** the worktree list before the command,
- **When** `moai worktree sweep` runs without `--yes` (the report form; `--json` is the machine form),
- **Then** the exit status is 0, `git worktree list --porcelain | grep -c '^worktree '` gives the same count before and after, and each candidate row carries owner, branch, predicate result, and unpushed-commit count (REQ-CAP-010).
- **Baseline (re-observed 10-10):** `grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go` returns no output, so the report has neither the owner field nor the unpushed-commit count — RED. The report form already exists by default (no `--yes`; `sweep.go:230`), and it shows each predicate's observed value (`sweep.go:591`). The branch field is not separately observed here.

### AC-CAP-013 — Unpushed or unowned worktrees are preserved (REQ-CAP-011)

- **Given** a fixture worktree whose branch has an unpushed commit, and a fixture worktree without a confirmed owner,
- **When** the sweep runs in report mode,
- **Then** both are reported as PRESERVE, and neither is disposed (REQ-CAP-011).
- **Baseline:** RED for the unowned half. No owner predicate exists (`grep -n -i owner internal/cli/worktree/sweep.go` → no output). The unpushed half is already PRESERVE through the landing predicate (`sweep.go:405–470`), which is not separately observed here.

### AC-CAP-014 — `moai worktree list` matches git (REQ-CAP-013)

- **Given** the repository,
- **When** `moai worktree list --json` is run and its array length is compared with `git worktree list --porcelain | grep -c '^worktree '`,
- **Then** the two counts are equal (REQ-CAP-013).
- **Baseline (re-observed 10-10):** `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` returns no output (exit 1). No list verb exists in source, so the criterion's own command cannot succeed on this tree.

### AC-CAP-015 — Bulk sweep disposes only owner-confirmed candidates (REQ-CAP-012, gated by decision Q2)

- **Given** decision Q2 in `decision-index.md` carries an approved verdict, and a fixture with one owner-confirmed candidate and one unconfirmed candidate, both of which pass the predicate chain,
- **When** `moai worktree sweep --yes` runs once,
- **Then** only the confirmed candidate is disposed, and an evidence row for it is written before its disposal (REQ-CAP-012).
- **Baseline:** RED by source. The `--yes` path has no owner-confirmation check (`internal/cli/worktree/sweep.go:230`; `grep -n -i owner` → no output).

## §D.6 Quality gate and Definition of Done

- The bulk disposal run (REQ-CAP-012, AC-CAP-015) runs only when decision Q2 carries an operator verdict, and the AC-CAP-012 and AC-CAP-013 green state is a precondition.
- Every AC names its command; every AC's verbatim output is recorded in `progress.md` §E.2 during run. The source-level observations in §D.0 are re-run with the criterion's own command in pre-flight.
- Regression guards (AC-CAP-006, AC-CAP-008, and the unapproved branch of AC-CAP-009) are reported as guards held, not as RED-to-GREEN transitions.
- Definition of done: AC-CAP-001 through AC-CAP-015 pass with recorded outputs; decisions Q1 through Q4 carry operator verdicts where they gate a change (Q3 gates the push of the M2 commit); and the sync phase closes §E.4.

## §D.7 Edge cases

- A worktree whose branch is merged remotely but has an unpushed local commit is PRESERVE (REQ-CAP-011).
- A zero-candidate sweep is reported as "0 candidates", not as a pass of the disposal path.
- A zero-match grep in AC-CAP-004 is a failed criterion, not a pass.
- A live-appended file (usage-log, lessons-inbox) grows between measurements. AC-CAP-006 therefore uses "at least the count before M2", never equality.
