# SPEC-UPDATE-MIGRATION-FIX-001 — Acceptance Criteria

Two-cell form per `verification-completeness.md` §2: every criterion below
carries a RED-now cell and a green-path cell. Plan-phase evidence lives in
the evidence ledger at the bottom; every table cell that cites a ledger id
inherits the four elements (command, verbatim output, exit code, tree SHA)
recorded there. Canonical baseline tree: 2aab5f797 (worktree
`.moai/worktrees/t1578`, branch WT-update-migration-fixes). RE-CUT NOTE:
the worktree fast-forwarded from 81786284e to 2aab5f797 mid-research
(155 commits); the ledger's per-entry SHA attributions below are
corrected to the tree each command actually read, and EV-6 records the
re-verification batch that re-pins every anchor on the new baseline.

## AC Matrix

| AC ID | Classification | Requirement | Milestone that flips it |
|-------|----------------|-------------|-------------------------|
| AC-UMF-001 | release-blocking | REQ-UMF-001 | M3 |
| AC-UMF-002 | release-blocking | REQ-UMF-001, REQ-UMF-002 | M3 |
| AC-UMF-003 | regression-guard | REQ-UMF-004 | M2 (confirmation; guard already exists) |
| AC-UMF-004 | regression-guard | REQ-UMF-005, REQ-UMF-006 | M2 |
| AC-UMF-005 | plan-gate (documentation state) | spec.md C.1 | satisfied at plan close |

## AC-UMF-001 — The version-match skip path runs the managed-surface integrity probe

**RED-now cell** (measured plan-phase; evidence ledger EV-4):

- Command: `sed -n '/if syncSkipped {/,/^\t}$/p' internal/cli/update.go`
- Observed stdout: the skip-path block prints the retired model-key strip,
  the participation step, and `return nil` — no integrity probe call
  anywhere in the block.
- Exit code: 0 (the observation succeeds; the CAPABILITY is absent).
- Why red for the right reason: the block is the only early return a
  version-matched update takes; no change other than M3's insertion puts a
  probe on this path, so the criterion cannot flip green through unrelated
  work.

**Green path cell**: M3 adds the probe invocation inside this block (and
threads the version-match-vs-cancellation distinction per REQ-UMF-003).
After M3, the same command prints the probe call between the block's
opening line and `return nil`, and
`go test ./internal/cli/ -run TestRunUpdate_VersionMatch_RunsIntegrityProbe -count=1`
prints `ok`. Both outputs recorded in progress.md §E.2 as the flip
witness.

## AC-UMF-002 — The probe reports a damaged representative path by name and never fails the update

**RED-now cell** (measured plan-phase; EV-4, supported by EV-5):

- No probe symbol exists in internal/cli (EV-5: zero call sites), so no
  update run can name a damaged managed path today — the damage class the
  mo.ai.kr run exposed (an empty skill dir) rides through every subsequent
  "Up to date · Skipping sync" update invisibly.
- Command (supporting): `grep -rn "Integrity" internal/cli/update*.go | grep -v _test`
  — expected at RED: no non-test hits carrying the probe's report marker.
- Exit code: 0/1 as grep reports; recorded verbatim in the ledger at M1-c.

**Green path cell**: M3. A fixture whose `.claude/settings.json` is deleted
runs a version-matched update; the run prints one warning naming the
missing path (REQ-UMF-001's greppable line) and exits 0 (REQ-UMF-002).
Pinned by `TestRunUpdate_VersionMatch_RunsIntegrityProbe` and
`TestIntegrityProbe_FailOpen`.

## AC-UMF-003 — Regression-guard: the 9-form root-denial normalization stays covered on the plain update path

**Classification note (undecidable disposition)**: the RED for this
criterion lives in the historical rc.26 binary (the defect as observed on
mo.ai.kr) and CANNOT be re-executed on this tree — the repair (#1792) is
already an ancestor of the base. Per verification-completeness §2.1 this
criterion loses release-blocking eligibility and is classified
regression-guard; it is never recorded as a fix by this SPEC.

**Evidence cell** (not a RED re-execution — the classification record):

- EV-1: the guard test family is green on the base tree (`ok`, 7.715s,
  exit 0), covering all 9 legacy forms through the real `moai update`
  path with implementation-independent specifier lists in the test.
- EV-2: the timeline placing the repair (#1792, 2026-10-07 05:22:36 UTC)
  30 minutes before the production run (05:52:41 UTC, rc.26 binary).

**Green path**: none required; M1-a re-measures EV-1 on the run HEAD and
M2 confirms the pin, both recorded in progress.md §E.2.

## AC-UMF-004 — Regression-guard: sync cycles leave no empty managed skill/agent directories; the installer reports empty directory targets

**Classification note (undecidable disposition)**: the historical RED (the
empty `moai-lane-watchdog` directory on mo.ai.kr) was produced by the
retired plugin-mode exclusion walk and is not re-executable on this tree —
same §2.1 disposition as AC-UMF-003.

**Evidence cell**: EV-3 records the mechanism retirement
(SPEC-USER-ASSET-INSTALL-001, merged into this base). M1-b performs the
current-tree measurement (sync-cycle fixture + installer directory-target
probe) and its output decides the branch: if an empty-directory producer
is measured, M2 escalates to repair (spec.md R1) and this criterion gains
a live RED; otherwise M2 lands the guards.

**Green path**: M2's `TestTemplateSync_LeavesNoEmptyManagedSkillDirs`
(internal/cli, counts its swept set and fails on a zero-count sweep) and
the installer empty-target test (internal/userassets) green; verbatim
output in progress.md §E.2.

## AC-UMF-005 — The settings-purity review item is dispositioned with rationale (plan-gate)

**Verification** (plan-phase, checked by the plan auditor, not a run-phase
flip): `grep -c "OUT OF SCOPE" .moai/specs/SPEC-UPDATE-MIGRATION-FIX-001/spec.md`
returns >= 1 and spec.md Section C.1 carries the rationale. Status at plan
close: satisfied by authoring. This criterion gates the PLAN, not the
release; it carries no RED-now cell because its subject (the disposition
text) is created by this SPEC document itself.

## Quality gates (Definition of Done)

- Scoped test families green: internal/cli update-path family,
  internal/userassets — verbatim output in progress.md §E.2.
- `gofmt -l` clean on every touched file; `golangci-lint run` introduces
  no NEW findings on touched packages (pre-existing baseline reported
  separately).
- Scope proof: `git diff --stat` from the card base shows Go-only changes
  under internal/cli and internal/userassets (no template tree, no
  .claude/rules).
- Every new test asserts on a non-empty swept set (no `[no tests to run]`
  passes).
- Cross-platform build: `GOOS=windows GOARCH=amd64 go build ./...` exit 0
  (the probe touches no syscall surface; measured at M4 anyway).

## Evidence Ledger (plan-phase measurements; canonical baseline 2aab5f797 — per-entry attributions corrected for the mid-research re-cut, see EV-6)

### EV-1 — the 9-form normalization guard is green on the base tree

```
$ go test ./internal/cli/ -run 'TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	7.715s
```
exit code: 0. First measured 2026-10-09; the run straddled the mid-research
re-cut (81786284e → 2aab5f797), so its tree attribution is uncertain —
EV-6 records the authoritative re-run pinned to 2aab5f797 (the test file
and the migration file are byte-unchanged between the two bases, EV-6).

### EV-2 — the repair landed between the observation and the card

```
$ git show 87da06367 -s --format='%ci | %s'
2026-10-07 05:22:36 +0000 | fix(update): migrate legacy root-denial settings rules to canonical forms on update (card t1569 M2) (#1792)
$ head -1 /tmp/moaikr-force-update.log
Current version   moai-adk v3.2.0-rc.26
```
exit codes: 0, 0. The production run's backup dir name
(`.moai-backups/20261007_145241`, log line 56) timestamps the run at
2026-10-07 14:52:41 KST = 05:52:41 UTC — 30 minutes after the #1792 merge,
on a binary (rc.26) built before it.

### EV-3 — the empty-dir suspect mechanism is retired on this tree

```
$ sed -n '5,17p' internal/template/deployer_mode.go
// deployer_mode.go — the deploy-mode surface after SPEC-USER-ASSET-INSTALL-001
// M7. The plugin-mode split is RETIRED with its carrier: the plugin payload
// constant, the mirror policy, the exclusion walk branch, the .mcp.json
// strip filter, the plugin re-home path, and the mirror re-home entry point
// are all gone. What remains:
//
//   - DeployMode — the project's deployment_mode RECORD surface only. The
//     record exists in deployed projects and update never flips it (REQ-018);
//     the deployer itself no longer branches on it and carries a single
//     project payload shape.
//   - isCommonAssetRoot — the REQ-005 walk exclusion: the project payload
//     carries no common skill or agent file in any mode; the four user roots
//     (plus the manifest/backup homes) are the install surface.
```
exit code: 0. `isCommonAssetRoot` (same file, lines 33-45) excludes
`.claude/skills/`, `.claude/agents/moai/`, `.agents/skills/`,
`.codex/agents/moai/` from the project payload in every mode.

### EV-4 — RED anchor for AC-UMF-001/002: the skip path carries no integrity probe

```
$ sed -n '/if syncSkipped {/,/^\t}$/p' internal/cli/update.go
	if syncSkipped {
		// A version-matched update runs no sync and no merge, so the retired
		// per-agent model/effort keys are stripped here, after its own backup.
		// A user-cancelled merge returns the same skipped=true; the helper
		// re-evaluates the version predicate and leaves that case untouched.
		if err := stripRetiredModelConfigOnVersionMatch(cmd, out, "."); err != nil {
			updateLedger.requiref(sevWarn, "retired model-key removal failed: %v", err)
		}
		// Card t1527 D5 + repair round: the deferred render carries the block
		// on this early return too — no explicit call here.
		// SPEC-FEEDBACK-PARTICIPATION-001 (REQ-ANON-004): a version-matched
		// update is still a finished plain template-sync run, so the ask runs
		// here too; its own gates (mode flags, terminal, CI, asked) decide
		// whether anything prompts. A failure warns; it never fails the update.
		if err := runParticipationStep(cmd, out); err != nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Participation ask", "failed", err.Error(), &th))
		}
		runParticipationFlushAtUpdate(cmd.ErrOrStderr())
		return nil
	}
```
exit code: 0. The block runs exactly two side-steps and returns; no content
check exists between the version stamp comparison and the return.
ATTRIBUTION CORRECTION (moving-coordinate): this read was taken AFTER the
mid-research re-cut — the participation step visible in the output is part
of the 81786284e..2aab5f797 delta — so the measurement tree is
2aab5f797, not 81786284e as first recorded. Re-read verbatim-identical on
2aab5f797 (EV-6); the structural RED conclusion is unchanged.

### EV-5 — no probe symbol exists in the update package (supporting)

AUTHORING-DISCIPLINE CORRECTION: this ledger row was initially authored
BEFORE its command was executed — the recorded output was reasoning-derived,
not measured, which violates the claim-integrity rule this ledger exists to
serve. Corrected by running the command for record on HEAD f569be5d8:

```
$ grep -rn "runManagedSurfaceIntegrityProbe" internal/cli/
(no output)
```
exit code: 1 (grep no-match), measured on f569be5d8. The symbol is
introduced by plan.md M3, so absence is structural; this grep is its
mechanical witness, now actually executed.

### EV-7 — second tree movement: f569be5d8 is the plan-artifact commit itself

```
$ git rev-parse --short HEAD
f569be5d8
$ git log --oneline -1
f569be5d8 feat(SPEC-UPDATE-MIGRATION-FIX-001): plan-phase artifacts (M, 6 artifacts)
$ git diff --stat 2aab5f797..HEAD -- internal/cli/update_deny_migration.go internal/cli/update_deny_migration_test.go internal/cli/update.go internal/template/deployer_mode.go internal/userassets/install.go internal/cli/update_clean_install.go
(empty)
```
exit codes: 0, 0, 0. HEAD advanced from 2aab5f797 to f569be5d8 by the
lane's commit of the six plan artifacts — no new main absorption; the
anchor-path delta is empty, so every EV-1..EV-6 measurement on 2aab5f797
remains valid on f569be5d8. Supporting spot check on f569be5d8:
`grep -c "runIntegrityProbe\|IntegrityProbe\|integrity probe"
internal/cli/update.go` = 0.

### EV-6 — re-verification batch on the re-cut baseline 2aab5f797

```
$ git rev-parse --short HEAD
2aab5f797
$ git diff --stat 81786284e..HEAD -- internal/cli/update_deny_migration.go internal/cli/update_deny_migration_test.go internal/cli/update.go internal/template/deployer_mode.go internal/userassets/install.go internal/cli/update_clean_install.go
 internal/cli/update.go               | 57 ++++++++++++++++++++++++------------
 internal/cli/update_clean_install.go |  6 +++-
 2 files changed, 43 insertions(+), 20 deletions(-)
```
exit code: 0. The deny-migration implementation, its test, deployer_mode.go,
and internal/userassets/install.go are byte-unchanged between the two
bases. Fresh re-reads on 2aab5f797: the `if syncSkipped {` block prints
byte-identical to EV-4 (still no integrity probe);
`grep -c legacyRootDenyNormalizeMap internal/cli/update_deny_migration.go`
= 3; call sites at update.go:433 and update_clean_install.go:570 — both
before the version-match short-circuit (~line 611).

```
$ go test ./internal/cli/ -run 'TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	2.866s
```
exit code: 0. The 9-form normalization guard is GREEN on the re-cut
baseline 2aab5f797 — this is the authoritative EV-1 measurement; the
earlier 7.715s run's tree attribution is uncertain (mid-research re-cut)
and is retained only as history.
