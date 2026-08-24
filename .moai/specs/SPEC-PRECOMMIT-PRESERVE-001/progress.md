# SPEC-PRECOMMIT-PRESERVE-001 — Progress

Card: t230 · Tier M · Class C · branch `WT-precommit-preserve`

## §E.1 Plan-phase Audit-Ready Signal

- **Artifacts**: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` — created at plan-phase,
  `status: draft`.
- **Baseline tree**: `294b4b6ab` (worktree `.claude/worktrees/t230`, divergence `0 0`).
- **Premises verified in this tree, not inherited on trust** (all `@294b4b6ab`): the missing
  comparison and backup (`hook_install_precommit.go:139-151`), the disguised success line (`:172`),
  the call sites (`update_template_sync.go:575`, `init.go:898`), the parity test
  (`hook_install_precommit_test.go:38`), and the zero-hit `exec` redirection sweep (spec.md §A.4).
- **Citation convention adopted** (spec.md §A.0): every `file:line` names its tree's SHA; the symbol
  is the durable anchor. Adopted after v0.1.0 wrongly declared card t230's `init.go:773` stale — it
  is correct on the primary checkout `@a1b1ca696`, while `:898` is correct here. Retraction recorded
  in spec.md §A.2.
- **Decisions taken**: three-way attribution raised to REQ-PCP-014; digest sidecar over in-body
  stamp, over a full-content snapshot (deferred, not dismissed — it changes no decision logic), and
  over a one-entry previous-digest corpus (rejected on the cheap form's own terms at v0.3.0, and it
  likewise re-opens no decision if later adopted); back-up-and-overwrite over preserve, with the
  silence prohibition (REQ-PCP-006) written so no policy choice can trade it away; and, new at
  v0.3.0, **Decision 3 — release composition** [RETIRED at v0.4.0 with the milestone it sequenced;
  see the trim record below]: M1+M2 ship with the hook body unchanged, M3 follows in a later
  release, bound mechanically by REQ-PCP-015 / AC-PCP-015. All with rejected alternatives stated
  (spec.md §A.5).
- **Counts (v0.3.0, superseded by the v0.4.0 trim below)**: 15 REQ, 15 AC — within the Tier M ceiling of 16/16, with no requirement dropped to
  stay under it. **Falsifiability split (restated at v0.3.0, previously overstated as 11/3)**: ten
  criteria fail against the untouched tree; five are falsifiable only against a named implementation
  mutant — the four behaviour-preserving ones (AC-PCP-008, -012, -013, -015) plus AC-PCP-002, whose
  Given ("bytes match the recorded digest") is unconstructible where no sidecar mechanism exists, so
  it goes red at setup rather than by observing today's behaviour.
- **Every AC carries a mutant and a failing input** (acceptance.md § How to read a criterion). The
  card's headline mutant — a correct, recoverable, entirely silent overwrite — is defeated by
  AC-PCP-004/006/007, which assert the notice text rather than post-run file state.

- **Plan-audit iteration 1** (`.moai/reports/t230/plan-audit.md`): PASS-WITH-DEBT, 0.875 against the
  Tier M threshold 0.80, audited at `7b2f42be0`. Seven defects raised (D1-D2 critical/blocking, D3
  major/blocking, D4-D7 minor/optional). Remediated at v0.3.0 — **all seven closed**, none declined.
  Each of D1/D2/D3 was independently re-measured in this tree before editing rather than accepted on
  the report's word; the measurements are recorded in spec.md §A.5 (D3 release-magnitude table),
  REQ-PCP-004 (D2 writer bindings), and REQ-PCP-010 (D1 precedence). Re-audit pending (iteration 2
  of the Tier M ceiling of 2).

- **Plan-audit iteration 2** (`.moai/reports/t230/plan-audit-2.md`): PASS-WITH-DEBT, 0.8875 against
  the Tier M threshold 0.80, audited at `7b2f42be0`. All three iteration-1 blocking defects (D1-D3)
  confirmed closed. Seven NEW defects: N1-N4 blocking, N5-N7 optional. Verdict on fitness: **M1 and
  M2 fit to enter run-phase; M3 not**.
- **v0.4.0 — scope reduction, not a remediation.** N1 (a release constraint checked at end-of-M2
  where it is green by construction), N2 ("the last released hook body" carrying three referents),
  and N3 (an unresolved yield rule against card t237) share one cause: M3 changes the distributed
  hook body, which forced this SPEC to reason about release composition — a thing this SPEC has no
  mechanism to enforce. Measured in this worktree at `e7fdd4a47`:
  `grep -rn 'git_hooks/pre-commit' .github/workflows/ scripts/` → exit 1, 0 hits, so no release
  tooling reads the pre-commit template at all. (The looser `grep -rn 'git_hooks' …` returns 2 hits,
  both naming `.git_hooks/pre-push` — `scripts/grep-no-release-automation.sh:12` and
  `scripts/ci-mirror/prepush_e2e_test.sh:25` — a different hook, and neither an enforcement of hook
  body composition.) The three
  were dissolved rather than patched, by moving the `pre-commit.local` extension point (REQ-PCP-011,
  REQ-PCP-012) and the composition binding (REQ-PCP-015) to a successor card. N4 was re-evaluated on
  its own terms **after** the trim and survived it — the no-record legacy population is M1's
  concern, not M3's — and is closed by AC-PCP-005 sub-case (c), whose fixture is the `v3.1.2` blob
  read from git rather than the incoming constant. N5 (Go test made the deciding check for
  AC-PCP-004 clause (ii)), N6 (self-counting numeral dropped), N7 (REQ-PCP-010 rephrased onto the
  installer; REQ-PCP-015's half dissolved with the requirement) all closed. None declined.
- **Counts after the trim**: **12 REQ, 12 AC**, 1:1, within the Tier M ceiling of 16/16. Identifiers
  are NOT renumbered — 001-010, 013, 014, with gaps at 011/012/015 — because the two audit reports
  cite the original numbers and silently re-pointing them would falsify the audit trail.
  **Falsifiability split**: nine criteria fail against the untouched tree; three are falsifiable only
  against a named implementation mutant (AC-PCP-008, AC-PCP-013 behaviour-preserving, plus
  AC-PCP-002 whose Given is unconstructible where no sidecar exists).
- **The core mutant re-checked after the AC set changed** (the card's headline case: overwrites
  correctly, backs up correctly, prints nothing): still defeated, and by the same three criteria —
  AC-PCP-004 clause (i), AC-PCP-006, AC-PCP-007 clause (i) — none of which was removed or weakened.
  AC-PCP-004 lost one asserted string (`pre-commit.local`) and retains two; a notice naming a
  facility this SPEC no longer ships would be a false instruction, so dropping it removes an
  assertion that would otherwise have been wrong rather than one that was load-bearing.
- **What the trim leaves unguarded, and what now guards it**: REQ-PCP-015 was also the only
  mechanical check that the hook body stayed byte-identical. Its replacement is not prose — it is
  AC-PCP-005 sub-case (c) (a legacy fixture read from `git show v3.1.2:…`, which goes red the moment
  the body changes) plus AC-PCP-013 (parity, PASS required and SKIP rejected). Both sit inside the
  test suite rather than at a release boundary nothing reads.
- **Plan-audit iteration 3 (`plan-audit-3.md`, PASS-WITH-DEBT 0.8875, flat) — D1-D3 closed at
  v0.5.0.** Each finding was re-measured in this tree before being accepted; none was taken on the
  auditor's word.
  - **D1 (critical)** — re-measured: `gh issue view 1641` → **OPEN**, editing `preCommitHookContent`
    and the template twin, verified patch on `t312-precommit-vet @ b6f478b1a`. The auditor is right
    that the v0.4.0 trim narrowed the composition constraint onto "the successor card" while the
    card about to change the body is t237. Closed by restoring **§A.5 Decision 3**, re-scoped to bind
    the *release* rather than a card, with three named red-moments: the M2 scope gate (green by
    construction — labelled as such, not counted), a **sync-phase DoD item read against the
    integration ref** (acceptance.md §D.3 — flips when a sibling merge lands a body change, so it can
    genuinely fail), and the release-candidate cut, which lies **outside** this SPEC's lifecycle and
    is therefore routed outward as a release-checklist item rather than re-created as an internal
    check. That last point is deliberate: iteration 2's N1 was a constraint checked where it could
    not fail, and inventing a second such check would have repeated it. Override path stated with its
    cost (release-note disclosure + a §D.4 re-pin before the cut). No new requirement: 12 REQ / 12 AC
    unchanged.
  - **D2 (major)** — re-measured rather than assumed: a probe package containing a single `t.Skip`
    test returns `--- SKIP` / `ok` / **exit 0** under `go test … -count=1 -v`, so the old Decides
    genuinely could not distinguish "(c) passed" from "(c) never ran". Closed with the AC-PCP-013
    treatment — `-v` plus the run's skip status inspected, a Then clause that fails on a skipped
    sub-case, and a third named mutant (a (c) that skips whenever the tag lookup fails). CI exposure
    re-measured: `fetch-depth: 0` at `.github/workflows/ci.yml:119,219,316,343,398,455` — six
    occurrences, so tags resolve in CI and the exposure is forks, tarballs and shallow clones.
  - **D3 (major)** — closed by the re-pin route (the first of the auditor's two options), recorded as
    **acceptance.md §D.4**: (c) pins to the most recent released tag whose hook body is byte-identical
    to the incoming body (`v3.1.2` here, re-verified rc 0 this round), the body-changing card re-pins
    it in the same commit, deletion is prohibited, and the interim expectation flip (a `v3.1.2`-era
    hook becoming a genuine backup-and-notice case) is stated. The body-stability signal is
    explicitly relocated to where it can still fail — plan.md §F's M2 scope gate and AC-PCP-013 — so
    the two entangled signals are separable.
  - **D5 (minor)** — closed: the `.git/hooks/` edge row now cites the creating call
    (`os.MkdirAll(hookDir, …)` in `InstallPreCommitHook`, `hook_install_precommit.go:132 @294b4b6ab`,
    re-verified) instead of the path assignment at `:131`.
  - **D4 (minor) — declined.** A third notice element ("this will recur on the next update") is
    defensible, but the notice's strength is its size, and REQ-PCP-006's disclosure floor holds
    without it; the auditor states the decline is equally defensible. **D6 (minor)** — no action
    required by its own terms; the numbering gaps stay accounted for in spec.md §B.
  - `~/go/bin/moai spec lint …/spec.md` → `✓ No findings` after the edits.

## §E.2 Run-phase Evidence

### M1 — the three-way classifier

**Tree**: worktree `.claude/worktrees/t230`, branch `WT-precommit-preserve`, parent `969290823`.
Evidence files under `.moai/state/verify/t230/`.

**Diff surface**: `internal/cli/hook_install_precommit.go` (installer logic) and the new
`internal/cli/hook_install_precommit_attribution_test.go`. `preCommitHookContent` and
`internal/template/templates/.git_hooks/pre-commit` are untouched (spec.md §C.1); no caller changed
(the warning writer of §C.4 belongs to M2, which is where the notice it carries lands).

**Pre-flight** (plan.md §C): call sites unchanged at `update_template_sync.go:575` /
`init.go:898`; `git show v3.1.2:…/.git_hooks/pre-commit | cmp - <same path>` → rc 0
(`preflight` recorded in `.moai/state/verify/t230/v312-hook.txt`);
`go test ./internal/cli/ -run TestPreCommit -count=1` → `ok … 46.621s` before any edit;
`grep -c 'sha256' internal/cli/hook_install_precommit.go` → `0` and
`grep -c 'pre-commit\.bak' …` → `0`, the AC baselines.

**What M1 delivers**: `classifyPreCommitHook` compares three operands — installed bytes, the digest
recorded in `.git/hooks/.moai-pre-commit.sha256`, and the incoming content — and returns a verdict
(`unmodified` / `user-modified`) plus the basis that produced it (`record` /
`undecidable-legacy`). The record is written after every successful hook write. The verdict is
recorded on the installer and **not acted on**: M1 decides, M2 hangs the backup and the notice off
the decision.

| AC | Test | Result | Failing input observed red |
|---|---|---|---|
| AC-PCP-014 | `TestPreCommitThreeWayAttribution` (2 cases) | PASS | two-way implementation → case one FAIL, case two PASS *by luck* (`mutant-a-twoway.txt`) |
| AC-PCP-001 | `TestPreCommitProvenanceRecorded` (2 runs) | PASS | empty-sidecar stub (`mutant-b2-empty-stub.txt`); write-once-never-refresh mutant (`mutant-b-write-once.txt`) |
| AC-PCP-002 | `TestPreCommitVersionBumpIsSilent` | PASS | naive two-way design (`mutant-a-twoway-ac002.txt`) |
| AC-PCP-005 | `TestPreCommitLegacyNoRecord` (a/b/c) | PASS, 0 SKIP | (a) silent-when-unknown mutant (`mutant-c-silent-unknown.txt`); (c) hook-body edit (`mutant-d-body-edit.txt`); skip clause (`mutant-e-bad-tag-fatal.txt` FAIL vs `mutant-e2-skip-reads-green.txt` exit 0) |

**Mutant results, in the terms the criteria set.**

- **The two-way implementation fails exactly where AC-PCP-014 says it must.** Case one (record
  matches installed, incoming differs) went `{class:user-modified basis:record}` against
  `{class:unmodified …}` — a routine version bump misread as a user patch — while case two passed.
  This is the criterion's whole point: a single-case version of it is satisfiable by the design it
  exists to reject.
- **AC-PCP-002 needed strengthening to be falsifiable at M1.** Under the two-way mutant the
  original file-and-output assertions passed, because M1 takes no backup and prints no notice under
  *any* classification. A second arm was added that runs the same Given through the direct
  installer and asserts the verdict; the mutant then fails it
  (`hook_install_precommit_attribution_test.go:187`). Without that arm the criterion observed
  nothing about the noise regression it is named for.
- **AC-PCP-005's sub-case (a) is load-bearing, as written.** The silent-when-unknown mutant passed
  (b) *and* (c) and was caught only by (a).
- **Sub-case (c) discriminates from (b), as claimed.** A one-line edit to `preCommitHookContent`
  left (b) green and turned (c) red — (b) builds its fixture from the incoming content, so it
  proves the code path, not the population.
- **The skip clause is enforced by `t.Fatalf`, not `t.Skip`.** With the pin aimed at a
  non-existent tag the criterion FAILs (exit 1). The `t.Skip` variant of the same failure exits
  **0** with the parent reporting `--- PASS` and the package `ok` — audit finding D2 reproduced
  verbatim, and the reason a skip is treated as a gap here rather than a pass.

**Quality gates**: `go build ./...` exit 0; `go vet ./internal/cli/...` exit 0;
`go test ./internal/cli/ -run TestPreCommit -count=1 -v` → all PASS, **0 SKIP**, including
`TestPreCommitTemplateMatchesConstant` (PASS, not SKIP — the paired-edit guard actually ran).

**Deliberately not delivered in M1** (M2's, per plan.md §F): the backup file, the notice, the
warning-writer split and its two caller lines. `lastProvenanceErr` records a failed sidecar write
rather than discarding it, so M2 has the seam REQ-PCP-010 sub-case (b) needs; M1 does not fail the
caller on it.

### M2 — disclosure: backup and notice

**Tree**: worktree `.claude/worktrees/t230`, branch `WT-precommit-preserve`, implementation commit
`d5c706e25` (parent `f71f63d1b`, the M1 commit). Evidence files under `.moai/state/verify/t230/`.

**Diff surface**: `internal/cli/hook_install_precommit.go` (backup + notice + warning writer +
precedence), the new `internal/cli/hook_install_precommit_disclosure_test.go`, signature updates in
`hook_install_precommit_test.go` / `hook_install_precommit_attribution_test.go`, and exactly one line
at each caller: `update_template_sync.go:575` gains `, errOut`; `init.go:898` gains
`, cmd.ErrOrStderr()`. Nothing else in either caller moved. `preCommitHookContent` and its template
twin are untouched — scope gate `cmp` rc 0 both before the work and after the last mutant revert
(`POST-MUTANT-CMP-RC0`).

**Pre-flight** (plan.md §C, verbatim, `@f71f63d1b`): `git rev-parse --short HEAD` → `f71f63d1b` /
`WT-precommit-preserve`; call sites `update_template_sync.go:575` (passes `out`) and
`init.go:898` (passes `cmd.ErrOrStderr()`), resolved by symbol;
`git show v3.1.2:…/.git_hooks/pre-commit | cmp - <same path>` → rc 0; baseline
`go test ./internal/cli/ -run TestPreCommit -count=1` → `ok … 12.987s`; lint baseline
`golangci-lint run --timeout=2m ./internal/cli/...` → `0 issues.`.

**What M2 delivers**: on a `user-modified` verdict, the pre-run bytes are copied to
`.git/hooks/pre-commit.bak.<20060102T150405Z>` (colon-free UTC, Windows-filename-safe) **before**
the replacement write, via `O_EXCL` so an occupied name is never clobbered (REQ-PCP-003/009).
`installPreCommitHookOptional` gains a warning writer and emits a two-element notice — the backup
path and the fact of replacement — on it (REQ-PCP-004); both callers bind it to the command's
stderr. Failure precedence (REQ-PCP-010): a failed backup write returns the
`errPreCommitBackupFailed` sentinel, which the wrapper turns into a warning while the hook stays
byte-identical to pre-run (no replacement); a failed post-write provenance write warns and keeps the
replacement. The backup-succeeded/write-failed edge (§D.1) names the orphan backup on the warning
writer instead of printing a false replacement notice.

| AC | Test | Result | Failing input observed red |
|---|---|---|---|
| AC-PCP-003 | `TestPreCommitModifiedHookIsBackedUp` | PASS | pre-edit tree: `found 0: []` — today's installer writes no backup (`m2-preedit-red-ac003-ac004ii.txt`) |
| AC-PCP-004 (i) | `TestPreCommitBackupNoticeContent` (exact-output) | PASS | headline silent mutant: warning writer empty (`mutant-m2-silent.txt`) |
| AC-PCP-004 (ii) | `TestPreCommitWarningWriterWiring` — AST wiring test, 3 sub-cases incl. no-other-call-sites | PASS | unchanged callers: `expected 4 arguments … got 3` at BOTH sites, observed before the one-line caller edits (`m2-preedit-red-ac003-ac004ii.txt`) |
| AC-PCP-006 | `TestPreCommitNoSilentReplacement` (one case, both artifacts) | PASS | notice line deleted, backup retained (`mutant-m2-silent.txt`) |
| AC-PCP-007 | `TestPreCommitBackupOutputNotBareSuccess` (clauses i+ii) | PASS | same silent mutant — clause (i) empty writer, clause (ii) names nothing (`mutant-m2-silent.txt`) |
| AC-PCP-008 | `TestPreCommitInstall_PreservesForeignHook` extended with `assertNoBackup` | PASS | marker-less hooks routed into the backup path → `expected no backup file, found […pre-commit.bak…]` (`mutant-m2-ac008-markerless-backup.txt`) |
| AC-PCP-009 | `TestPreCommitBackupNoClobber` (same-second collision via `now` seam) | PASS | fixed-filename clobbering backup → `expected two distinct backups, found 0` (`mutant-m2-ac009-fixed-name.txt`) |
| AC-PCP-010 (a) | `TestPreCommitSupportWriteFailureNonFatal/a_…` | PASS | warn-then-overwrite-anyway mutant (warning + normal return preserved): fails **specifically the POST-STATE clause** — `hook was replaced despite the failed backup (3245 bytes, want the unchanged 78-byte pre-run body)`; sub-case (b) still PASSED under the mutant, confirming (a)'s post-state is the only clause that reaches it (`mutant-m2-ac010a-overwrite-anyway.txt`) |
| AC-PCP-010 (b) | `TestPreCommitSupportWriteFailureNonFatal/b_…` (dir-at-record-path, cross-platform) | PASS | precedence-reordering mutant (provenance written first, hook write skipped on its failure — (a)'s precedence over-applied to the wrong side): `the hook must BE replaced when only the post-write provenance write fails; got 78 bytes`, sub-case (a) still PASS under it (`mutant-m2-ac010b-skip-hook-write.txt`) |
| AC-PCP-013 | `TestPreCommitTemplateMatchesConstant` | PASS, **not SKIP** (`-v` inspected) | one-sided `preCommitHookContent` edit → `template len: 3245, constant len: 3292` (`mutant-m2-ac013-constant-edit.txt`); reverted, cmp re-verified rc 0 |

**How the (a) fixture separates the two writes.** The hooks dir is chmod 0500 while the existing
hook file is 0755: creating a NEW file (the backup) needs dir-write permission and fails, while
truncating the EXISTING hook file needs only file-write permission and succeeds. The mutant red
proves the hook write was physically possible and only the precedence clause prevented it — the
fixture does not pass "by accident of the directory being unwritable".

**Clause (ii) is a Go test, not a grep.** `TestPreCommitWarningWriterWiring` parses both caller
files with go/ast, finds every `installPreCommitHookOptional` call, asserts the warning-writer
argument directly (`cmd.ErrOrStderr()` at init) or resolves it (`errOut` at update, whose single
assignment inside the enclosing function must be `cmd.ErrOrStderr()`; `out` likewise
`cmd.OutOrStdout()` — enclosing-function scoping, because a second unrelated `out :=` lives in a
later function in the same file), and asserts no third call site exists in the package. A
`no_other_call_sites` sub-test scans every non-test `.go` file in `internal/cli`.

**Quality gates** (all `@d5c706e25`, this tree, this run):
- `go test ./internal/cli/ -run TestPreCommit -count=1 -v` → **35 PASS, 0 FAIL, 0 SKIP**, `ok …
  16.651s` (final rerun `m2-ac-green.txt`); `TestPreCommitTemplateMatchesConstant` PASS, not SKIP.
- `go test ./internal/cli/ -count=1` (full package) → `ok … 552.193s` (`m2-cli-full.txt`).
- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (`BUILDS-RC0`).
- `go vet ./internal/cli/...` → exit 0; `golangci-lint run --timeout=2m ./internal/cli/...` →
  `0 issues.` — one NEW errcheck finding (`f.Close` unchecked in the backup error path) surfaced
  during the run and was fixed (`_ = f.Close()`); the re-measure matches the pre-edit baseline of 0.
- Coverage, `-run TestPreCommit -cover`: package 6.7% under the filter (the package is dominated by
  untested-at-this-filter CLI surfaces; the changed file is the lens that matters):
  `InstallPreCommitHook` 89.7%, `backupPreCommitHook` 82.4% (uncovered: O_EXCL write/close
  disk-error branches), `installPreCommitHookOptional` 92.9%, `classifyPreCommitHook` 100%,
  `writePreCommitProvenance` 100% (`m2-coverage.txt`).
- Scope gate (plan.md §F end-of-M2): `cmp` rc 0 pre-edit, rc 0 post-mutant-revert, and the tree at
  `d5c706e25` differs from `f71f63d1b` in installer logic, tests, and two one-line caller edits only.

**Mutant discipline**: every implementation mutant was applied on top of `d5c706e25`, observed red
against its named criterion, then reverted with `git restore` — after each revert `git status
--short` showed only this progress.md and HEAD stayed `d5c706e25`. The pre-edit reds (AC-003,
AC-004 ii) were observed at `f71f63d1b` before any source edit.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_phase: complete (M1 + M2; M3 trimmed at v0.4.0)
run_complete_at: 2026-08-25
m1_commit_sha: f71f63d1b        # the three-way classifier (§E.2 M1)
run_commit_sha: d5c706e25       # M2 implementation + tests (evidence commit follows)
run_status: run complete
ac_pass_count: 12               # all 12 AC green: M1's 4 (014, 001, 002, 005) + M2's 8 (003, 004, 006, 007, 008, 009, 010, 013)
ac_fail_count: 0
ac_pending_count: 0
mutants_observed_red: 15        # M1: 7 (two-way x2, empty-stub, write-once, silent-unknown, body-edit, bad-tag); M2: 8 (silent headline, marker-less backup, fixed-name, overwrite-anyway, precedence-reordering, one-sided constant, + the pre-edit-tree red for AC-003 and AC-004(ii))
skipped_subcases: 0
hook_body_touched: false        # preCommitHookContent and its template twin byte-identical (scope gate rc 0)
callers_touched: true           # exactly the one §C.4-permitted line each, AST-asserted
new_warnings_or_lints_introduced: none   # one transient errcheck finding fixed before the implementation commit
cross_platform_build:
  native: exit 0
  windows_amd64: exit 0
total_run_phase_files: 6        # hook_install_precommit.go + 3 test files + 2 callers (one line each)
m1_to_mN_commit_strategy: per-milestone commits on WT-precommit-preserve (M1 f71f63d1b, M2 d5c706e25)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Recorded by the orchestrator at M2 spawn (2026-08-25). M1 ran serially in the prior session
(9f833cc0) without a §F log; recorded here at the first orchestrator-side run-phase spawn of this
session so the decision is auditable from M2 onward.

- **Input parameters**: tier M · scope 2 source files + 1-2 test files (installer + two one-line
  caller edits) · single domain (Go CLI) · language mix 100% Go · concurrency benefit LOW
  (coding-heavy, per Anthropic's coding-task parallelism caveat) · Agent Teams prereqs not met
  (no operator `--team` request)
- **Mode evaluation**: `direct` — no (Tier M Go implementation is manager-develop's domain);
  `fanout` — no (single-domain coding work, no research fan-out); `sweep` — no (2 files, not a
  high-volume mechanical transform); `agent-team` — no (explicit-request-only, none given);
  `serial` — **selected**
- **Decision**: serial
- **Justification**: One implementation surface (`hook_install_precommit.go`) plus two one-line
  caller edits — a single sequential manager-develop delegation carries the whole milestone; there
  is nothing genuinely parallel to fan out, and coding-heavy work stays sequential per the
  Anthropic multi-agent caveat.
