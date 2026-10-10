# SPEC-CAPACITY-RECLAIM-001 — Acceptance Criteria

This file is the verification layer. Each criterion is Given-When-Then with a binary test and the command that measures it. Baselines are the verbatim outputs recorded in `spec.md` §B.

**Tree pin.** The document-level pin is the audited tree `40cc1031c2b18781377f1bdf2c590cd39f08244e` (short `40cc1031c`) on branch `WT-capacity-reclaim`, observed in the plan session. Every code observation in §D.0 was taken on this pin. The plan commit that carries this file changes only SPEC artifacts. A criterion that needs a fixture run, or a commit that does not exist yet, carries no RED cell and is unadopted until pre-flight observes it.

## §0 Classes, adoption, command form, and evidence carriers

Each criterion belongs to one class, which states what its result means for the release gate:

- **RB — release-blocking.** The criterion gates completion. Its RED-now cell carries four elements together: the command (one read-only invocation: no pipe, redirection, `&&`, `;`, or subshell; `verification-completeness.md` §2.1), its verbatim stdout, its exit code captured by the runner (not by the command), and the tree SHA. Its green-path cell names the milestone that flips it and the passing output.
- **RG — regression guard.** Green on the pinned tree by design. It must hold at completion. It is reported as "held", never as a RED-to-GREEN transition.
- **MS — measurement.** The observation is the verbatim output, recorded in `progress.md` §E.2. A missing output, or a failing one, blocks the sync-phase close. MS covers a command that needs a pipeline or a fixture, and a plain git read whose exit status the runner cannot capture (G-EXIT).

**Exit capture (G-EXIT).** The worktree guard refuses shell text that names git inside a wrapper, so a plain git command surfaces only its stdout. For grep cells the runner captures the exit status with `bash -c '<command>; echo EXIT=$?'`, which the guard accepts when the text names no git; the cell's command remains the single grep. A plain git cell has no runner-captured exit status and is therefore classed MS.

**Evidence carriers.** The RED cells are the fenced entries EV-* below: command, verbatim stdout, exit status, and tree SHA. The §D.0 table cites them by id. Every entry is pinned to `5d36bf91b15c9869118298613df77209f8a3893b`, the HEAD at measurement. The commits after `40cc1031c` change only `.moai/` paths (`git diff --stat 40cc1031c 5d36bf91b`), so the code paths that the §D.0 observations rest on are unchanged; the document pin at line 5 is superseded for these entries.

```text
EV-004 (AC-CAP-004, RB, premise cell)
  command: grep -n -E 'usage-log|lessons|corrupt' internal/template/templates/.gitignore
  stdout:  (empty)
  exit:    1 (runner-captured)
  tree:    5d36bf91b15c9869118298613df77209f8a3893b

EV-005 (AC-CAP-005, MS)
  command: git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*'
  stdout:  .moai/harness/learning-history/archive/2026-07.jsonl.gz.corrupt
           .moai/harness/learning-history/archive/2026-08.jsonl.gz.corrupt
           .moai/harness/learning-history/archive/2026-09.jsonl.gz.corrupt
           .moai/harness/usage-log.jsonl
           .moai/lessons-inbox.jsonl
  exit:    not runner-captured (G-EXIT); no stderr printed
  tree:    5d36bf91b15c9869118298613df77209f8a3893b

EV-011 (AC-CAP-011, RB)
  command: grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go
  stdout:  243:				if profileBytes <= config.DefaultProfileMaxBytes {
  exit:    0 (runner-captured)
  tree:    5d36bf91b15c9869118298613df77209f8a3893b

EV-012 (AC-CAP-012, RB)
  command: grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go
  stdout:  (empty)
  exit:    1 (runner-captured)
  tree:    5d36bf91b15c9869118298613df77209f8a3893b

EV-013 (AC-CAP-013, RB)
  command: grep -n -i owner internal/cli/worktree/sweep.go
  stdout:  (empty)
  exit:    1 (runner-captured)
  tree:    5d36bf91b15c9869118298613df77209f8a3893b

EV-014 (AC-CAP-014, RB)
  command: grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'
  stdout:  (empty)
  exit:    1 (runner-captured)
  tree:    5d36bf91b15c9869118298613df77209f8a3893b

EV-016a (AC-CAP-016, MS, reversed pair: not in ancestor order)
  command: git merge-base 5d36bf91b15c9869118298613df77209f8a3893b 2aab5f797b75983e132af451da68f69e3426557b
  stdout:  2aab5f797b75983e132af451da68f69e3426557b
  exit:    not runner-captured (G-EXIT); the stdout is the observation
  tree:    5d36bf91b15c9869118298613df77209f8a3893b

EV-016b (AC-CAP-016, MS, control pair in ancestor order)
  command: git merge-base 2aab5f797b75983e132af451da68f69e3426557b 5d36bf91b15c9869118298613df77209f8a3893b
  stdout:  2aab5f797b75983e132af451da68f69e3426557b
  exit:    not runner-captured (G-EXIT); the stdout is the observation
  tree:    5d36bf91b15c9869118298613df77209f8a3893b

EV-016c (AC-CAP-016, MS, content of a non-progress-only commit)
  command: git show --name-only --format= 40cc1031c2b18781377f1bdf2c590cd39f08244e
  stdout:  .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md
           .moai/specs/SPEC-CAPACITY-RECLAIM-001/decision-index.md
           .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md
           .moai/specs/SPEC-CAPACITY-RECLAIM-001/progress.md
           .moai/specs/SPEC-CAPACITY-RECLAIM-001/spec.md
  exit:    not runner-captured (G-EXIT)
  tree:    5d36bf91b15c9869118298613df77209f8a3893b
```

**Adoption at this revision.** RB with four-element cells: AC-CAP-004 (premise cell EV-004), AC-CAP-011 (EV-011), AC-CAP-012 (EV-012), AC-CAP-013 (EV-013), and AC-CAP-014 (EV-014). MS with verbatim output: AC-CAP-005 (EV-005) and AC-CAP-016 (EV-016a to EV-016c). A test that a criterion names but that does not exist yet is never a RED cell: `go test -run` with no match prints `[no tests to run]` and exits 0 (`verification-completeness.md` §1.1). For such a criterion the cell observes the premise the test checks, and the test's own failing run is recorded at pre-flight, before its milestone commit. The extended cases of AC-CAP-004 are written in M2, so their failing run is that pre-flight record.

**Conditional classes under verdict 8a62.** AC-CAP-002 and AC-CAP-015 are N/A: decision Q2 is rejected and no bulk run executes under this SPEC. AC-CAP-013 applies in full. AC-CAP-009 is conditional on Q1, and its approved branch (`apply`, the split default) is the one to verify. AC-CAP-007 is held on its default branch (decision Q4 default; `sinks.go` unchanged). The green paths of AC-CAP-001 and AC-CAP-005 include the lessons-inbox untracking, which is held under decision Q3, so they hold only after a Q3 verdict and the M2 commit. AC-CAP-011 applies to the profile-cap warning as set in A11.

**Decision outcome branches (this revision).** Q1: approved split (REQ-CAP-006). Q2: rejected (REQ-CAP-012 and REQ-CAP-014 are not exercised; no bulk run). Q3: no verdict (B1 stays a local commit). Q4: default (REQ-CAP-004; the AC-CAP-007 default branch). A11: warning only (REQ-CAP-009).

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
- **RED-now (EV-004, premise cell):** no template rule matches the three state-file patterns, so the extended cases cannot pass. `grep -n -E 'usage-log|lessons|corrupt' internal/template/templates/.gitignore` prints nothing and exits 1 (EV-004). The isolated `git check-ignore --no-index` probe itself was not run: a plain git command against a scratch repository outside the worktree is refused by the worktree guard (Gap G-PROBE; `verification-claim-integrity.md` §3.1). The extended test's own failing run is recorded at pre-flight, before the template edit.
- **Green path:** M2 adds the three rules. The same grep then prints the three rule lines, and the same test command passes.

### AC-CAP-005 — Repository stops tracking the state files (MS; maps REQ-CAP-002, REQ-CAP-003)

- **Given** the tree after B1,
- **When** `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*'` is run,
- **Then** stdout is empty.
- **RED-now (EV-005, observed at 5d36bf91b):** the five paths of EV-005 are listed, with no stderr. The class is MS because a plain git exit status is not runner-captured (G-EXIT).
- **Green path:** M2 `git rm --cached` of the five files (B1) → empty stdout. The lessons-inbox part of this path is held under decision Q3; it holds only after a Q3 verdict.

### AC-CAP-006 — Local copies survive untracking; archive payloads unchanged (RG; maps REQ-CAP-002, REQ-CAP-003)

- **Given** the baseline recorded in B0 (`progress.md` §E.2: the byte counts of the two state files and the SHA-256 of the three archive `.corrupt` files, each with its command and verbatim output),
- **When** `ls -l .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl` and `shasum -a 256` over the three archive `.corrupt` paths are run after B1,
- **Then** all five files exist; each state file is non-empty when its baseline size was non-zero; and each archive's SHA-256 equals its baseline value.
- The state-file size is not required to stay at or above its baseline. The 30-day prune rewrites `usage-log.jsonl` (`internal/harness/retention.go:421`, `overwriteWithEvents`), so a size decrease is expected behavior, not a loss. Deletion, or truncation to zero, is the failure.
- **Classification:** regression guard, green by design. Mutant: deleting the local copies in place of `git rm --cached` turns the existence check red.

### AC-CAP-016 — The baseline commit precedes the untracking commit (MS; maps REQ-CAP-002, REQ-CAP-003; `verification-claim-integrity.md` §2.3)

- **Given** baseline commit B0 (touching only `progress.md`) and untracking commit B1, with their SHAs recorded in `progress.md` §E.2,
- **When** `git merge-base <B0> <B1>` is run, and then `git show --name-only --format= <B0>`,
- **Then** the printed merge base equals the full SHA of B0, so B0 is an ancestor of B1 (equivalent to `git merge-base --is-ancestor <B0> <B1>` exiting 0; printed here because a plain git exit status is not runner-captured, G-EXIT), and the second command lists only `.moai/specs/SPEC-CAPACITY-RECLAIM-001/progress.md`.
- **RED-now, observed on known-bad inputs (no B0 or B1 exists yet):**
  - EV-016a, reversed pair: `git merge-base 5d36bf91b15c9869118298613df77209f8a3893b 2aab5f797b75983e132af451da68f69e3426557b` prints `2aab5f797b75983e132af451da68f69e3426557b`, which is not its first argument, so the first commit is not an ancestor of the second.
  - EV-016b, control pair in the required order: `git merge-base 2aab5f797b75983e132af451da68f69e3426557b 5d36bf91b15c9869118298613df77209f8a3893b` prints its first argument, so the predicate holds.
  - EV-016c, content: `git show --name-only --format= 40cc1031c2b18781377f1bdf2c590cd39f08244e` lists five SPEC files, so the content clause is false for a commit that is not progress-only.
  The predicate can therefore fail (EV-016a, EV-016c) and can pass (EV-016b). The commit graph is the only witness of this order; the commit message and the session record are not.
- **Green path:** after B0 and B1 exist, the commands on B0 and B1 print B0 and list only `progress.md`; the observed output is recorded at the M2 commit.

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

### AC-CAP-011 — Profile cap warns and deletes nothing; no compiled cap (RB; maps REQ-CAP-009)

- **Given** a fixture home whose profile root has `projects/` entries totalling 6 MiB, and a fixture `state.yaml` in the home config sections with `state.profile_max_bytes: 1048576`; a second fixture with the key absent. The fixtures set `HOME` with `t.Setenv`,
- **When** `go test ./internal/cli/ -run '^TestProfileCapWarnsWithoutDeleting$' -count=1` is run,
- **Then** in fixture 1 the profile-cap warning names the profile and `state.profile_max_bytes`, and the home cleanup produces no `projects/` candidate on account of the cap; in fixture 2 no profile-cap warning is emitted and no `projects/` candidate is produced. Neither fixture uses a compiled value.
- **RED-now (EV-011):** `grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go` prints line 243 and exits 0 (EV-011): a compiled cap still selects a deletion candidate. The warning path does not exist yet, so its test-level failing run is recorded at pre-flight.
- **Mutant:** a loader with a compiled fallback fails fixture 2; a cap that selects a candidate fails fixture 1.
- **Green path:** M5. Then `grep -rn -E 'DefaultProfileMaxBytes' internal --include='*.go'` prints nothing, and the test passes.

## §D.5 Worktree report, list, and disposal (AC-CAP-012 to AC-CAP-015)

### AC-CAP-012 — The report carries owner, unpushed, and last-activity fields with checked values; nothing is removed (RB; maps REQ-CAP-010)

- **Given** the tree build with the M3 change, built with `make build` and invoked by path,
- **When** `git worktree list --porcelain | grep -c '^worktree '` is run before and after `./bin/moai worktree sweep --json` and the records are read,
- **Then** the sweep exits 0, the two counts are equal, and every non-protected record carries the keys `path`, `branch`, `owner`, `owner_state`, `unpushed`, `last_activity`, `landed`, `dirty`, `ignored`, `anchored`, `cwd_occupied`, `on_base`, and `verdict`. The values are checked, not only the keys: `owner` is the bound card's label or JSON `null` (an empty string fails); `owner_state` is the bound card's state, `unbound`, or `undetermined` (any other value fails); `unpushed` is a non-negative integer equal to `git rev-list --count <tip> --not --remotes` for the record's branch, or `null` (a string fails); `last_activity` is an ISO-8601 timestamp equal to `git log -1 --format=%cI <tip>`. On the fixture set of AC-CAP-013, the values equal the known fixture values: C has `owner_state` `done`, W has `owner` `null` and `owner_state` `unbound`, and U has `unpushed` 1. A card in state `run` bound to a further fixture tree gives `owner_state` `run`, because the state is read from the card record and not inferred.
- **RED-now (EV-012):** `grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go` prints nothing and exits 1 (EV-012). The verdict record has neither field: its JSON keys are `path`, `branch`, `tier`, `verdict`, `reason`, `landed`, `dirty`, `ignored`, `anchored`, `cwd_occupied`, and `on_base` (`sweep.go:72–96`).
- **Green path:** M3 (report fields; the verdict and the `--yes` set are unchanged).

### AC-CAP-013 — The report marks unpushed and unowned candidates; `--yes` removes exactly the pre-change DISPOSE set (RB; maps REQ-CAP-011)

- **Given** fixtures driven through the existing seams (`sweepFetchBase`, `sweepAncestor`, `sweepProcessCWDs`, `sweepDoneCleanup`; `sweep.go:102–195`) and the advisory owner seam: (U) a tree whose branch has one local commit on no remote-tracking ref; (W) an otherwise disposable tree (landed, clean, unanchored, no process cwd inside) with no card binding; (L) an otherwise disposable tree anchored by a git worktree lock whose reason names a live pid (the test process's own pid); (C) the control, an otherwise disposable tree bound to a card in state `done`; and a golden set of the verdicts that the pre-change evaluation returns for U, W, L, and C,
- **When** `go test ./internal/cli/worktree/ -run '^TestSweepReportAdvisoryNoDisposalChange$' -count=1` runs. The test reads the report evaluation, then runs the `--yes` apply once,
- **Then** (1) the report marks U with advisory cause `unpushed-commits` (count 1) and W with advisory cause `owner-unconfirmed` (`owner_state` `unbound`); L keeps its existing cause `anchored`; C shows `owner_state` `done`; (2) every verdict equals the golden set (U `PRESERVE`, W `DISPOSE`, L `PRESERVE`, C `DISPOSE`); (3) after `--yes`, the removed directories are exactly the golden DISPOSE set (W and C), and U and L still exist. No fixture's disposal differs from the pre-change evaluation.
- **Mutant:** adding the owner predicate to the verdict evaluation makes W `PRESERVE`; steps (2) and (3) fail.
- **RED-now (EV-013):** `grep -n -i owner internal/cli/worktree/sweep.go` prints nothing and exits 1 (EV-013): the report has no owner field and no advisory cause. The existing predicates for U (the landing layers, `internal/cli/worktree/landing_predicate.go:395`, `:423–431`) and for L (`anchored`) are not separately red.
- **Green path:** M3 (fields and advisory causes in the report; the verdict and the `--yes` set unchanged).

### AC-CAP-014 — `moai worktree list` matches git per path (RB; maps REQ-CAP-013)

- **Given** the tree build, invoked by path, and a fixture worktree of known size (a file of N KiB written by the fixture),
- **When** `./bin/moai worktree list --json` is run and compared with `git worktree list --porcelain`, and `du -sk` is run on the fixture worktree in the same test,
- **Then** the array length equals the count of `worktree` lines; for each record the `path` equals a `worktree` path, and the record's `branch` equals that worktree's `branch` line with the `refs/heads/` prefix removed (a detached record has `""` in both; the check is per path, so a swapped pair fails); each `size_kib` is a JSON integer ≥ 0 (a string or a fraction fails), and the fixture record's `size_kib` equals its `du -sk` value; and the table form prints one row per worktree below its header.
- **Mutant:** swapping the branches of two records, or setting every `size_kib` to -1, fails the per-path branch check or the integer check.
- **RED-now (EV-014):** `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` prints nothing and exits 1 (EV-014): no list verb is registered.
- **Green path:** M3 (`list` subcommand); the test passes.

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
