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

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_phase: M1 only (M2 pending)
run_commit_sha: <M1-COMMIT-SHA>
run_status: M1 complete
ac_pass_count: 4          # AC-PCP-014, -001, -002, -005 (M1's four)
ac_fail_count: 0
ac_pending_count: 8       # M2's: -003, -004, -006, -007, -008, -009, -010, -013
mutants_observed_red: 6   # two-way; two-way vs AC-002; empty-stub; write-once; silent-unknown; body-edit; bad-tag
skipped_subcases: 0
hook_body_touched: false  # preCommitHookContent and its template twin byte-identical
callers_touched: false    # the §C.4 warning-writer lines belong to M2
new_warnings_or_lints_introduced: none
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
