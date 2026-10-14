# SPEC-CAPACITY-RECLAIM-001 — Acceptance Criteria

This file is the verification layer. Each criterion is Given-When-Then with a binary test and the command that measures it. Baselines are the verbatim outputs recorded in `spec.md` §B.

**Tree pin.** The document-level pin is tree `e2658f30127375e75e0c606d09ca97be2c551cf2` (short `e2658f3012`) on branch `WT-capacity-reclaim`, observed in the v0.2.0 revision session after the card worktree absorbed local `main`. Every code observation in §D.0 and every EV entry below was taken on this pin. EV-015 and EV-016 were observed, and EV-008 re-observed, on tree `111493b71465d475a8d399343aa931df6d628e7d` (short `111493b71`), the HEAD at the v0.2.1 repair session; no `internal/` path differs between `e2658f3012` and the tree this revision commits (`git diff --name-only e2658f3012 HEAD -- internal/` is empty), so the code premises are unchanged. Of the code paths this SPEC measures, only `internal/config/defaults.go` and `internal/template/templates/.moai/config/sections/workflow.yaml` differ from the iter3-audited tree `ab9f5a6d3` (`git diff --name-only ab9f5a6d3 e2658f3012 -- internal/`; 218 files changed in the absorb, those two are the only measured ones). The commit that carries this revision changes only SPEC artifacts. A criterion that needs a fixture run, or a commit or test that does not exist yet, carries no RED cell of its own and is unadopted until pre-flight observes it; its premise cell is recorded below.

## §0 Classes, adoption, command form, exit capture, and evidence carriers

Each criterion belongs to one class, which states what its result means for the release gate:

- **RB — release-blocking.** The criterion gates completion. Its RED-now cell carries four elements together: the command (one read-only invocation: no pipe, redirection, `&&`, `;`, or subshell; `verification-completeness.md` §2.1), that command's verbatim stdout, its exit code as its own element, and the tree SHA. Its green-path cell names the milestone that flips it and the passing output.
- **RG — regression guard.** Green on the pinned tree by design. It must hold at completion. It is reported as "held", never as a RED-to-GREEN transition.
- **MS — measurement.** The observation is the verbatim output, recorded in `progress.md` §E.2. A missing output, or a failing one, blocks the sync-phase close. MS covers a command that needs a pipeline or a fixture.

**Exit capture.** Stated from this session's observed runner behavior, so every cell can satisfy it without modifying its command: the runner prints a command's stdout verbatim and prints no status line for a successful run; it renders a bare single invocation that exits non-zero as `Exit code N` (observed in this session on a false `git merge-base --is-ancestor` pair, EV-011); and it renders an empty-output run as `(Bash completed with no output)` with no status line, including for a no-match grep whose documented convention is exit 1 (observed twice in this session, EV-001 and EV-006). A cell therefore records: the verbatim stdout; the exit code as a **runner observation** only where the runner printed `Exit code N`; and otherwise the exit value labelled **convention**, together with what the runner did and did not show. No element is labelled as captured when it was not: the word "runner-captured" appears below only where the runner printed the status. No cell wraps its command to capture an exit (no `bash -c`, no `; echo $?`); the single-invocation form is never modified.

**Evidence carriers.** The RED cells are the fenced entries EV-* below: command, verbatim stdout, exit element, and tree SHA. The §D.0 table cites them by id. Every entry is pinned to `e2658f30127375e75e0c606d09ca97be2c551cf2`, the HEAD at measurement.

```text
EV-001 (AC-CAP-003, RB, premise cell)
  command: grep -n -E 'usage-log|lessons|corrupt' internal/template/templates/.gitignore
  stdout:  (empty)
  exit:    1 by grep's no-match convention. The runner showed "(Bash completed with
           no output)" and no exit-status line (observed twice this session); the
           value 1 is the documented convention, not a runner observation.
  tree:    e2658f30127375e75e0c606d09ca97be2c551cf2

EV-002 (AC-CAP-004, RB)
  command: git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl .moai/harness/learning-history/archive
  stdout:  .moai/harness/learning-history/archive/2026-04.jsonl.gz
           .moai/harness/learning-history/archive/2026-05.jsonl.gz
           .moai/harness/learning-history/archive/2026-06.jsonl.gz
           .moai/harness/learning-history/archive/2026-07.jsonl.gz
           .moai/harness/learning-history/archive/2026-07.jsonl.gz.corrupt
           .moai/harness/learning-history/archive/2026-08.jsonl.gz
           .moai/harness/learning-history/archive/2026-08.jsonl.gz.corrupt
           .moai/harness/learning-history/archive/2026-09.jsonl.gz
           .moai/harness/learning-history/archive/2026-09.jsonl.gz.corrupt
           .moai/harness/usage-log.jsonl
           .moai/lessons-inbox.jsonl
  exit:    0. The run returned its stdout with no error block; this runner prints
           "Exit code N" for a non-zero bare invocation (observed on EV-011), and
           no such block appeared.
  tree:    e2658f30127375e75e0c606d09ca97be2c551cf2

EV-006 (AC-CAP-009, RB, premise cell)
  command: grep -rn 'TestRetentionPruneArchivesBeforeRewrite' internal
  stdout:  (empty)
  exit:    1 by grep's no-match convention; the runner showed "(Bash completed
           with no output)" and no status line.
  tree:    e2658f30127375e75e0c606d09ca97be2c551cf2

EV-007 (AC-CAP-008, RB, approved-branch cell)
  command: grep -n -E 'mode: (report|apply)' internal/template/templates/.moai/config/sections/workflow.yaml
  stdout:  331:        mode: report              # report | apply (auto path only)
  exit:    0. The run returned its stdout with no error block (same runner
           contract as EV-002).
  tree:    e2658f30127375e75e0c606d09ca97be2c551cf2

EV-008 (AC-CAP-012, RB, premise cell)
  command: grep -rn 'TestSessionStartAutoPathSplit' internal
  stdout:  (empty)
  exit:    1 by grep's no-match convention; the runner showed "(Bash completed
           with no output)" and no status line.
  tree:    111493b71465d475a8d399343aa931df6d628e7d
           (re-observed 2026-10-14 in the v0.2.1 repair session; identical
           output at the prior pin e2658f30127375e75e0c606d09ca97be2c551cf2)

EV-009 (AC-CAP-011, RB, premise cell)
  command: grep -rn 'profile_max_bytes' internal/hook
  stdout:  (empty)
  exit:    1 by grep's no-match convention; the runner showed "(Bash completed
           with no output)" and no status line. No profile-cap warning exists in
           the hook path.
  tree:    e2658f30127375e75e0c606d09ca97be2c551cf2

EV-010 (AC-CAP-011, RB)
  command: grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go
  stdout:  243:				if profileBytes <= config.DefaultProfileMaxBytes {
  exit:    0. The run returned its stdout with no error block (same runner
           contract as EV-002).
  tree:    e2658f30127375e75e0c606d09ca97be2c551cf2

EV-011 (exit-capture probe, not a criterion cell)
  command: git merge-base --is-ancestor e2658f30127375e75e0c606d09ca97be2c551cf2 2aab5f797b75983e132af451da68f69e3426557b
  stdout:  (empty)
  exit:    1 — runner observation: the runner printed "Exit code 1" for this bare
           non-zero invocation. This entry exists to witness the runner's
           non-zero-surfacing contract stated in §0.
  tree:    e2658f30127375e75e0c606d09ca97be2c551cf2

EV-012 (AC-CAP-013, MS, reversed pair: not in ancestor order)
  command: git merge-base 5d36bf91b15c9869118298613df77209f8a3893b 2aab5f797b75983e132af451da68f69e3426557b
  stdout:  2aab5f797b75983e132af451da68f69e3426557b
  exit:    0. The run returned its stdout with no error block; the stdout is the
           observation (it is not the first argument, so the first commit is not
           an ancestor of the second).
  tree:    e2658f30127375e75e0c606d09ca97be2c551cf2

EV-013 (AC-CAP-013, MS, control pair in ancestor order)
  command: git merge-base 2aab5f797b75983e132af451da68f69e3426557b 5d36bf91b15c9869118298613df77209f8a3893b
  stdout:  2aab5f797b75983e132af451da68f69e3426557b
  exit:    0. The run returned its stdout with no error block; the stdout equals
           the first argument, so the predicate holds for this pair.
  tree:    e2658f30127375e75e0c606d09ca97be2c551cf2

EV-014 (AC-CAP-013, MS, content of a non-progress-only commit)
  command: git show --name-only --format= 40cc1031c2b18781377f1bdf2c590cd39f08244e
  stdout:  .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md
           .moai/specs/SPEC-CAPACITY-RECLAIM-001/decision-index.md
           .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md
           .moai/specs/SPEC-CAPACITY-RECLAIM-001/progress.md
           .moai/specs/SPEC-CAPACITY-RECLAIM-001/spec.md
  exit:    0. The run returned its stdout with no error block.
  tree:    e2658f30127375e75e0c606d09ca97be2c551cf2

EV-015 (AC-CAP-012, RB, clause-1 seeding premise)
  command: grep -rn 'ClassVerifyScratch' internal/hook
  stdout:  (empty)
  exit:    1 by grep's no-match convention; the runner showed "(Bash completed
           with no output)" and no status line. No hook fixture seeds a `verify`
           residue-class target: the class is enumerated only under
           `internal/hygiene` (targets.go, gc.go).
  tree:    111493b71465d475a8d399343aa931df6d628e7d

EV-016 (AC-CAP-012, RB, decision-gate premise)
  command: sed -n '395,404p' internal/hygiene/gc.go
  stdout:  		for _, d := range report.Decisions {
			rows = append(rows, AuditRow{
				Unit: "gc", Mode: string(ModeApply), Outcome: Outcome(d.Outcome),
				Path: d.Path, Reason: d.Reason, Signal: d.Signal,
			})
		}
	}
	if len(rows) == 0 {
		return
	}
  exit:    0. The run returned its stdout with no error block (same runner
           contract as EV-002). In apply mode the GC writes one `unit: gc` /
           `mode: apply` row per decision and writes nothing at all when no
           decision fires (the `len(rows) == 0` gate), which is why clause (1)
           must seed a target for the GC to decide on.
  tree:    111493b71465d475a8d399343aa931df6d628e7d
```

**Adoption at this revision.** RB with four-element cells: AC-CAP-003 (premise cell EV-001), AC-CAP-004 (EV-002), AC-CAP-008 (approved-branch cell EV-007), AC-CAP-009 (premise cell EV-006), AC-CAP-011 (EV-010, plus premise cell EV-009), and AC-CAP-012 (premise cells EV-008, EV-015, EV-016). MS with verbatim output: AC-CAP-001 (the §B pipeline), AC-CAP-002 (fixture run), AC-CAP-010 (fixture test), and AC-CAP-013 (EV-012 to EV-014 on known-bad inputs). RG: AC-CAP-005, AC-CAP-006, AC-CAP-007. A test that a criterion names but that does not exist yet is never a RED cell by itself: `go test -run` with no match prints `[no tests to run]` and exits 0 (`verification-completeness.md` §1.1). For such a criterion the cell observes the premise the test checks (the absence probes EV-006, EV-008, EV-009, EV-015), the test's mutant observation is recorded at pre-flight, and the test's own failing run against the pre-change tree is the pre-flight record.

**Decision outcome branches (all recorded; this revision carries no open decision that gates a criterion).** Q1: approved split (REQ-CAP-006; the approved branch is the one AC-CAP-008 verifies). Q2: rejected — the bulk-run scope moved to card B, so no criterion of this SPEC depends on it. Q3: the measurement recipe was executed on 2026-10-10 (`.moai/reports/t1631/q3-clone-measurement.md`) and the untracking is authorized by the operator's narrowed-scope decision; B1 lands through the card's normal landing path. Q4: default (REQ-CAP-004; AC-CAP-009 verifies the guarantee on the default branch). A11: warning only (REQ-CAP-009; AC-CAP-011).

**AC renumbering map (v0.1.2 → v0.2.0).** AC-CAP-001→AC-CAP-001; 003→002; 004→003; 005→004 (MS→RB, D3); 006→005; 007→006 (RB→RG, default branch only); 008→007; 009→008 (approved-branch RB cell added, D1); 010→010; 011→011 (narrowed to card A's paths); 016→013. Removed to card B: 002, 012, 013, 014, 015 (D2 dissolved with AC-CAP-014; D5 with AC-CAP-015; D6 with AC-CAP-012). Added: AC-CAP-009 (D8), AC-CAP-012 (D9).

## §D.0 Adoption status

| AC | Class | Traces to | RED-now at e2658f3012 | Green path | Status |
|---|---|---|---|---|---|
| AC-CAP-001 | MS | REQ-CAP-002, -003, -010 | Tracked harness and lessons-inbox bytes: `132585731 total` (observed; measured with a pipeline, plan §C) | M2 untracks the eleven files → residual 300,746 bytes (arithmetic on the §B values), re-measured in the green run | Observed; output recorded at M2 |
| AC-CAP-002 | MS | REQ-CAP-007 | Not observed (fixture apply run). The write sites exist at `internal/hygiene/rotate.go:167` and `:184` | Fixture apply run → `"mode":"apply"` rows | Unadopted until pre-flight |
| AC-CAP-003 | RB | REQ-CAP-001 | The extended test must fail before the template edit. The template has no rule for usage-log, lessons-inbox, corrupt, or the archives (EV-001: no match) | M2 template rules → the test passes | Unadopted until the RED run is recorded |
| AC-CAP-004 | RB | REQ-CAP-002, -003, -010 | The eleven paths of EV-002 are listed | M2 `git rm --cached` → empty stdout | RED observed (EV-002); pre-flight records nothing further |
| AC-CAP-005 | RG | REQ-CAP-002, -003, -010 | Held: all eleven local files exist; the archive hashes are recorded in B0 | After B1, the same | Guard; baseline recorded at M0 |
| AC-CAP-006 | RG | REQ-CAP-004 | Held: `usage-log.jsonl` is an owner exception (`sinks.go:91`) and the registry (`sinks.go:26–50`) holds no entry for it | Unchanged | Guard |
| AC-CAP-007 | RG | REQ-CAP-005 | Held: `lessons-inbox.jsonl` is an owner exception (`sinks.go:90`) | Unchanged | Guard |
| AC-CAP-008 | RB (approved branch) | REQ-CAP-006 | `mode: report` at workflow.yaml:331 (EV-007): the approved branch (`apply`) is red now | M1 flips the template default → the grep prints `mode: apply` | RED observed (EV-007) |
| AC-CAP-009 | RB | REQ-CAP-004 | The named test does not exist (EV-006). The guarantee it pins is source-verifiable today (`retention.go:416→421`) | M1 writes the test; its mutant observation is the pre-flight RED record | Unadopted until the test's failing mutant run is recorded |
| AC-CAP-010 | MS | REQ-CAP-008 | Not observed (fixture test). RED by source: backups selected by age only (`internal/cli/clean_home.go:310–331`) | M3 backups floor → the test passes | Unadopted until pre-flight |
| AC-CAP-011 | RB | REQ-CAP-009 | A compiled cap still selects a deletion candidate (EV-010); no warning path exists (EV-009) | M3 removes the `clean_home.go:243` candidate source and adds the warning → `grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go` prints nothing, and the fixture test passes (the doctor readers at `doctor_disk.go:155`, `:261` remain until card D and are not part of this criterion) | RED observed (EV-010); the test's failing run recorded at pre-flight |
| AC-CAP-012 | RB | REQ-CAP-006, -007 | The named test does not exist (EV-008); the auto path's apply behavior is unverified | M1 writes the test; its failing run is the pre-flight record | Unadopted until the test's failing run is recorded |
| AC-CAP-013 | MS | REQ-CAP-002, -003, -010 (`verification-claim-integrity.md` §2.3) | Not observable on this SPEC's own commits yet (no B0 or B1 exists). Observed on known-bad inputs: EV-012 (reversed pair — the predicate can fail), EV-013 (control pair — it can hold), EV-014 (content — a non-progress-only commit lists five files) | B0 is an ancestor of B1, and B0 touches only `progress.md` | Unadopted until B0 exists |

## §D.1 Sizing and apply rows (AC-CAP-001, AC-CAP-002)

### AC-CAP-001 — Git-tracked harness under 5 MB (MS; maps REQ-CAP-002, REQ-CAP-003, REQ-CAP-010)

- **Given** the tree after the untracking commit B1,
- **When** `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl | xargs -0 wc -c | tail -1` is run,
- **Then** the printed total is less than 5,000,000 bytes (`spec.md` assumption A5).
- **Baseline (observed at e2658f3012):** `132585731 total` across 191 tracked files.
- **Green path:** M2 untracks the eleven named files (usage-log 90,331,575 B; lessons-inbox 6,681,068 B; the nine archive files 35,272,342 B; total 132,284,985 B). The residual is 300,746 B by arithmetic on the measured values, re-measured in the green run. This criterion is the operator's completion criterion 1; it became reachable again on the merged tree only with the archive untracking of REQ-CAP-010 (`spec.md` §A v0.2.0: without it the residual is 5,230,544 B).

### AC-CAP-002 — Apply rows present (MS; maps REQ-CAP-007)

- **Given** a fixture project whose `workflow.hygiene.audit_log_max_bytes` is low enough that a fixture sink is at or above it,
- **When** `moai clean --audit-logs --apply` runs once in the fixture, invoked from the tree build by path,
- **Then** `grep -n '"mode":"apply"' .moai/logs/hygiene-audit.jsonl` prints at least one row for each applied action (rows whose outcome is `rotated`, `deleted`, or `error`).
- **Baseline:** not observed (fixture run). The field is pinned: `AuditRow.Mode` (`internal/hygiene/auditlog.go:16`), set to `apply` at `rotate.go:167` and `:184`.
- This criterion exercises the CLI path; the SessionStart auto path's apply rows are AC-CAP-012's observation (the operator's completion criterion 2 covers both).

## §D.2 Tracking and template (AC-CAP-003 to AC-CAP-005, AC-CAP-013)

### AC-CAP-003 — Template ignores the state files and archives (RB; maps REQ-CAP-001)

- **Given** the test `TestTemplateGitignoreIgnoresLocalArtifacts` (`internal/template/gitignore_local_artifacts_test.go:29`), extended with these cases: `.moai/harness/usage-log.jsonl` (ignored), `.moai/lessons-inbox.jsonl` (ignored), `.moai/harness/learning-history/archive/2026-07.jsonl.gz.corrupt` (ignored), `.moai/harness/learning-history/archive/2026-07.jsonl.gz` (ignored: the archive rule of REQ-CAP-010), and `.moai/harness/learning-history/tier-promotions.jsonl` (not ignored: the control for the file that stays tracked),
- **When** `go test ./internal/template/ -run '^TestTemplateGitignoreIgnoresLocalArtifacts$' -count=1` is run,
- **Then** the test passes. The test writes the embedded template into a `t.TempDir()` repository and asks `git check-ignore -q --no-index`, so the repository's own `.gitignore` (which already matches usage-log at line 352 and `.moai/reports/*` at line 235 in this worktree) does not answer for the template.
- **Mutant:** a template whose only addition is a comment line that names the file fails the positive cases.
- **RED-now (EV-001, premise cell):** no template rule matches the state-file, corrupt, or archive patterns, so the extended cases cannot pass. `grep -n -E 'usage-log|lessons|corrupt' internal/template/templates/.gitignore` prints nothing (EV-001). The isolated `git check-ignore --no-index` probe itself was not run in this session: a plain git command against a scratch repository outside the worktree was refused by the worktree guard in the v0.1.1 session (Gap G-PROBE; `verification-claim-integrity.md` §3.1). The extended test's own failing run is recorded at pre-flight, before the template edit.
- **Green path:** M2 adds the four rules. The same grep then prints the four rule lines, and the same test command passes.

### AC-CAP-004 — Repository stops tracking the eleven files (RB; maps REQ-CAP-002, REQ-CAP-003, REQ-CAP-010)

- **Given** the tree after B1,
- **When** `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl .moai/harness/learning-history/archive` is run,
- **Then** stdout is empty.
- **RED-now (EV-002, observed at e2658f3012):** the eleven paths of EV-002 are listed. The class is RB with a four-element cell: the command is a single invocation, its stdout is verbatim, its exit element is recorded per the §0 rule, and the tree is pinned. (This criterion was class MS in v0.1.2 on the G-EXIT premise; the premise was contradicted by the runner — this runner surfaces a non-zero bare git invocation as `Exit code N`, EV-011 — so the criterion is restored to RB per the iter3 D3 disposition.)
- **Green path:** M2 `git rm --cached` of the eleven files (B1) → empty stdout. The landing follows the card's normal path (Q3 satisfied; the lane does not push).

### AC-CAP-005 — Local copies survive untracking; archive payloads unchanged (RG; maps REQ-CAP-002, REQ-CAP-003, REQ-CAP-010)

- **Given** the baseline recorded in B0 (`progress.md` §E.2: the byte counts of the two state files and the SHA-256 of the three archive `.corrupt` files and of the six monthly `.jsonl.gz` archives, each with its command and verbatim output),
- **When** `ls -l .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl` and `shasum -a 256` over the nine archive paths are run after B1,
- **Then** all eleven files exist; each state file is non-empty when its baseline size was non-zero; and each archive's SHA-256 equals its baseline value.
- The state-file size is not required to stay at or above its baseline. The 30-day prune rewrites `usage-log.jsonl` (`internal/harness/retention.go:421`, `overwriteWithEvents`), so a size decrease is expected behavior, not a loss. Deletion, or truncation to zero, is the failure.
- **Classification:** regression guard, green by design. Mutant: deleting the local copies in place of `git rm --cached` turns the existence check red.

### AC-CAP-013 — The baseline commit precedes the untracking commit (MS; maps REQ-CAP-002, REQ-CAP-003, REQ-CAP-010; `verification-claim-integrity.md` §2.3)

- **Given** baseline commit B0 (touching only `progress.md`) and untracking commit B1, with their SHAs recorded in `progress.md` §E.2,
- **When** `git merge-base <B0> <B1>` is run, and then `git show --name-only --format= <B0>`,
- **Then** the printed merge base equals the full SHA of B0, so B0 is an ancestor of B1 (equivalent to `git merge-base --is-ancestor <B0> <B1>` exiting 0 — the runner observation form of which is EV-011), and the second command lists only `.moai/specs/SPEC-CAPACITY-RECLAIM-001/progress.md`.
- **RED-now, observed on known-bad inputs (no B0 or B1 exists yet):**
  - EV-012, reversed pair: `git merge-base 5d36bf91b15c9869118298613df77209f8a3893b 2aab5f797b75983e132af451da68f69e3426557b` prints `2aab5f797b75983e132af451da68f69e3426557b`, which is not its first argument, so the first commit is not an ancestor of the second.
  - EV-013, control pair in the required order: `git merge-base 2aab5f797b75983e132af451da68f69e3426557b 5d36bf91b15c9869118298613df77209f8a3893b` prints its first argument, so the predicate holds.
  - EV-014, content: `git show --name-only --format= 40cc1031c2b18781377f1bdf2c590cd39f08244e` lists five SPEC files, so the content clause is false for a commit that is not progress-only.
  The predicate can therefore fail (EV-012, EV-014) and can pass (EV-013). The commit graph is the only witness of this order; the commit message and the session record are not.
- **Green path:** after B0 and B1 exist, the commands on B0 and B1 print B0 and list only `progress.md`; the observed output is recorded at the M2 commit.

## §D.3 Retention and hygiene defaults (AC-CAP-006 to AC-CAP-009)

### AC-CAP-006 — usage-log keeps its owner; no registry entry (RG; maps REQ-CAP-004)

- **Given** the tree at any point of this SPEC's run,
- **When** `grep -n -E 'usage-log' internal/hygiene/sinks.go` is run,
- **Then** the only match is the owner-exception line (`sinks.go:91`) and no `sinkRegistry` entry names it (`sinks.go:26–50`).
- **Classification:** regression guard, green by design (Q4 default; `spec.md` REQ-CAP-004). Mutant: adding `usage-log.jsonl` to `sinkRegistry` turns it red.

### AC-CAP-007 — lessons-inbox is not rotated (RG; maps REQ-CAP-005)

- **Given** a fixture copy of `.moai/lessons-inbox.jsonl` at or above the fixture threshold,
- **When** `moai clean --audit-logs --apply` runs once in the fixture, invoked from the tree build by path,
- **Then** the fixture file's size and SHA-256 are unchanged, and no audit row names `lessons-inbox.jsonl`.
- **Classification:** regression guard, green by design (`internal/hygiene/sinks.go:90`, owner exception). Mutant: adding `lessons-inbox.jsonl` to `sinkRegistry` turns it red.

### AC-CAP-008 — Hygiene default follows the approved Q1 split (RB on the approved branch; maps REQ-CAP-006)

- **Given** decision Q1 in `decision-index.md` carrying the approved split verdict (operator verdict 8a62),
- **When** `grep -n -E 'mode: (report|apply)' internal/template/templates/.moai/config/sections/workflow.yaml` is run,
- **Then** the printed value is `apply`.
- **RED-now (EV-007, approved-branch cell, observed at e2658f3012):** the template default is `report` (line 331), so the approved branch is red now. The cell is a four-element ledger entry: single-invocation command, verbatim stdout, exit element per the §0 rule, tree pin. (The v0.1.2 cell carried an escaped-pipe form that printed nothing and no exit element — iter3 D1; this cell replaces it, and the v0.1.2 note that the approved branch was "not observable" is withdrawn: it is observable and red.)
- **Green path:** M1 flips the template default to `apply`; the same grep then prints `mode: apply`. The unapproved branch (`report`) is not exercised: the verdict is recorded.

### AC-CAP-009 — The retention prune archives before it rewrites (RB; maps REQ-CAP-004)

- **Given** the test `TestRetentionPruneArchivesBeforeRewrite` in `internal/harness`, driving the retention prune on a fixture log that holds stale and fresh events, with a fixture `archiveDir`,
- **When** `go test ./internal/harness/ -run '^TestRetentionPruneArchivesBeforeRewrite$' -count=1` is run,
- **Then** (i) after the prune, the stale events are present in `archiveDir/<YYYY-MM>.jsonl.gz` and the rewritten log holds only the fresh events; and (ii) with the archive write made to fail (a fixture archive path that cannot be written), the log file is byte-identical to its pre-call content — the rewrite never happens on a failed archive.
- **Mutant:** a prune variant that calls `overwriteWithEvents` before `archiveEvents` fails (ii): the log would be rewritten despite the failed archive. The mutant's failing run is the test's RED record, taken at pre-flight.
- **RED-now (EV-006, premise cell):** no test by that name exists (`grep -rn 'TestRetentionPruneArchivesBeforeRewrite' internal` prints nothing). The guarantee it pins is source-verifiable today (`retention.go:404` partition → `:416` archive → `:421` rewrite), which is what makes (i) and (ii) satisfiable by the unchanged prune: this is a characterization test of the default branch, and the Q4 default is the branch in force.
- **Green path:** M1 writes the test; it passes against the unchanged prune, and the mutant observation is recorded at pre-flight.

## §D.4 Archives, profile, and keep rules (AC-CAP-010, AC-CAP-011)

### AC-CAP-010 — Backups keep the three newest; releases keep the current release plus three (MS; maps REQ-CAP-008)

- **Given** a Go test that calls `scanHomeCleanable(root, retention, config.DefaultReleaseKeep, currentVersion, now)` (`internal/cli/clean_home.go:160`) on a fixture home root with: (i) five `backups/removed-*` directories, all older than the retention window, and six release binaries, the current version and five older non-current ones; (ii) five `backups/removed-*` directories, two of them younger than the retention window and three older,
- **When** `go test ./internal/cli/ -run '^TestHomeCleanKeepFloors$' -count=1` is run,
- **Then** in fixture (i) the candidates are exactly the two oldest backups and the two oldest non-current releases, and the current release is not a candidate; in fixture (ii) the candidates are exactly the two older-than-window backups that lie beyond the newest three.
- **Baseline:** RED by source. Backups are selected by age only (`clean_home.go:310–331`), so fixture (i) yields five backup candidates. The release part already matches (`clean_home.go:408–445`).
- **Green path:** M3 backups floor (`spec.md` A12). Releases are unchanged (A10).

### AC-CAP-011 — Profile cap warns and deletes nothing on card A's paths (RB; maps REQ-CAP-009)

- **Given** a fixture home whose profile root has `projects/` entries totalling 6 MiB, and a fixture `state.yaml` in the home config sections with `state.profile_max_bytes: 1048576`; a second fixture with the key absent. The fixtures set `HOME` with `t.Setenv`,
- **When** `go test ./internal/cli/ -run '^TestProfileCapWarnsWithoutDeleting$' -count=1` is run,
- **Then** in fixture 1 the profile-cap warning names the profile and `state.profile_max_bytes`, and the home cleanup produces no `projects/` candidate on account of the cap; in fixture 2 no profile-cap warning is emitted and no `projects/` candidate is produced. Neither fixture uses a compiled value. (The hook-side invocation of this advisory check is AC-CAP-012's observation (3): the same warning surfaces from the SessionStart pass.)
- **RED-now (EV-010, EV-009):** `grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go` prints line 243 (EV-010): a compiled cap still selects a deletion candidate. No profile-cap warning exists in the hook path (`grep -rn 'profile_max_bytes' internal/hook` prints nothing, EV-009). The test's own failing run is recorded at pre-flight.
- **Mutant:** a loader with a compiled fallback fails fixture 2; a cap that selects a candidate fails fixture 1.
- **Green path:** M3. Then `grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go` prints nothing (the candidate source is removed), the template ships the default, and the test passes. The doctor readers (`internal/cli/doctor_disk.go:155`, `:261`) are outside this criterion: they remain until card D and are not part of what this criterion measures.

## §D.5 SessionStart split behavior (AC-CAP-012)

### AC-CAP-012 — The SessionStart auto path applies only the non-destructive categories (RB; maps REQ-CAP-006, REQ-CAP-007)

- **Given** the test `TestSessionStartAutoPathSplit` in `internal/hook`, driving the SessionStart hygiene pass through the `hygieneRunFn` seam (`internal/hook/session_start_hygiene.go:19`) on a fixture project whose template `workflow.yaml` carries `mode: apply` and that is seeded with one GC residue-class target of the closed class `verify` (a session-keyed directory under the fixture project's `.moai/state/verify/` holding a datable entry file with an old `recorded_at`, the enumeration shape of `internal/hygiene/targets.go:173–189` with the dating field of `:92`), with a fixture home that sets `state.profile_max_bytes: 1048576` and a profile above it; a second fixture home with the key absent,
- **When** `go test ./internal/hook/ -run '^TestSessionStartAutoPathSplit$' -count=1` is run,
- **Then** (1) `.moai/logs/hygiene-audit.jsonl` gains rows with `"mode":"apply"` from the rotator unit, and at least one row carries `"unit":"gc"` with `"mode":"apply"` for the seeded `verify` target's path — the GC decided on the seeded target and wrote its apply row (it writes apply rows only per decision and writes nothing when none fires, EV-016), so apply rows exist for the non-destructive categories (the operator's completion criterion 2, auto path); (2) no row, report entry, or filesystem change names a path outside the fixture project's `.moai/logs` and `.moai/state`: no destructive category (worktrees, databases, run records) is touched, and no backups or releases path is touched — a fixture home seeded with old `backups/removed-*` directories and non-current releases is byte-identical after the pass; and (3) the profile-cap advisory check fires in the pass: in fixture 1 the pass emits the warning naming the profile and the key and deletes nothing; in fixture 2 it emits no profile-cap warning (REQ-CAP-009's hook surface).
- **Mutant:** a variant of the pass that invokes the home-cleanup path (`runCleanHome`) fails (2) on the fixture home; a registry that gains a home-path class fails (2) on the path scan; a loader with a compiled fallback fails (3) on the absent-key fixture; a GC whose apply path emits no row for the seeded `verify` target fails (1).
- **RED-now (EV-008, premise cell; EV-015, the clause-1 seeding premise; EV-016, the decision-gate premise):** no test by that name exists (`grep -rn 'TestSessionStartAutoPathSplit' internal` prints nothing), and no hook fixture seeds a `verify` scratch target (`grep -rn 'ClassVerifyScratch' internal/hook` prints nothing, EV-015). The behavior is partly source-verifiable today (the closed class registry `internal/hygiene/targets.go:29–36`; the home path reachable only via `moai clean --home`, `clean.go:107`; the GC's decision-gated apply-row write, EV-016) and partly new (the apply rows exist only after the M1 flip, which is AC-CAP-008's flip). The test's own failing run is recorded at pre-flight, before the M1 commit that flips the default.
- **Green path:** M1 writes the test (clauses 1 and 2) and flips the default; they pass. M3 extends the test with clause 3 once the loader and warning exist; the extended test passes at the M3 commit. The test's failing run against the pre-change tree is the pre-flight record.

## §D.6 Quality gate and Definition of Done

- **Gate (RB).** Every RB criterion passes, with its four-element RED cell (exit element per the §0 rule) and its green-path output recorded in `progress.md` §E.2, before the milestone commit that flips it.
- **Gate (RG).** Every RG criterion holds. A held guard is reported as held, never as a RED-to-GREEN transition.
- **Gate (MS).** Every MS criterion has its verbatim output recorded. A missing output, or a failing MS output, blocks the sync-phase close.
- **Decision outcome branches (recorded).** Q1 approved: AC-CAP-008's approved branch holds (`apply` after M1). Q2 rejected: the bulk-run scope is card B's; no criterion here depends on it. Q3: measurement recorded 2026-10-10, untracking authorized by the operator's narrowed-scope decision; B1 lands through the card's normal landing path; AC-CAP-001, -004, -005, and -013 are recorded locally and the sync phase records the landing state. Q4 default: AC-CAP-006 and AC-CAP-009 hold on the default branch; the rotation alternate is not built.
- **Completion criteria mapping.** The card's narrowed PASS definition maps: criterion 1 (git-tracked harness < 5 MB) → AC-CAP-001; criterion 2 (hygiene apply rows for rotation/cache) → AC-CAP-002 (CLI path) and AC-CAP-012 (auto path); criterion 3 (profile cap warning fires, measured) → AC-CAP-011.
- **Close.** The sync phase closes `progress.md` §E.4 only after the gates above are recorded.

## §D.7 Edge cases

- A zero-candidate fixture is not a pass of any apply path.
- `usage-log.jsonl` grows and may be rewritten by the 30-day prune during a run. AC-CAP-005 therefore checks existence, non-zero length, and archive hashes, never size equality.
- The retention prune may create a new monthly archive file while the run is in progress; AC-CAP-001's green re-measurement reads the tracked set after B1, so a new local archive file created after B1 is untracked by the template rule and does not affect the total.
- A zero-match grep in a RED cell is a recorded red, not a pass. The RED cell records stdout (empty), the exit element labelled per the §0 rule, and what the runner showed.
