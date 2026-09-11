# t574 — reproduction (develop tree)

Tree: t574 worktree `WT-temp-roots-ac`, base merge `f295fe698` (HEAD^2 = develop `a4461479e`;
`git merge-base develop HEAD` = `a4461479e`). Prediction pinned first in commit `7da398808`.
Toolchain: go1.26.8.

## Claim

The card reproduces: no acceptance criterion of SPEC-TODO-HOME-TEMP-GUARD-001 (status `completed`)
requires the production temp-root set to contain `/tmp` and `/var/folders`. The guarding test
`TestDefaultTempRoots_Membership` exists and works, but no AC demands it.

## Evidence

- Token census (`grep -c -F`): `Membership` acceptance=0 spec=0 progress=2; `defaultTempRoots`
  acceptance=0 spec=0 progress=3. Controls on the same files: `/var/folders` acceptance=3 spec=7,
  `os.TempDir()` acceptance=2 spec=6.
- AC-THG-006 (`acceptance.md:141-152`): Then = sibling `/tmpfoo` is NOT temporary; And = the temp root
  itself and its children ARE temporary — but its test (`TestTempOrigin_ComponentBoundary`) runs the
  positive half against a **stubbed** root (`stubTempRoots`), and the production-set subtest checks
  only the negative `/tmpfoo`. AC-THG-002 positively pins only the `os.TempDir()` member (symlink
  spelling equivalence).
- Baseline, no mutant (`repro-tests-develop.txt`): Membership + ComponentBoundary PASS, exit 0.
- Mutant M-574 (`defaultTempRoots` -> `{os.TempDir()}`), prediction in `mutant-prediction.md`:
  - `mutant-kanban.txt`: every AC-named kanban test PASS (SymlinkSpellingEquivalence,
    FailsOpenOnUnresolvable, ComponentBoundary, TempOriginRefusesHomeQueue,
    NonTempNonGitKeepsHomeFallback, PureGuardIsSilent, GitBranchUnreachedByGuard,
    PureFallbackWritesNothing); only `TestDefaultTempRoots_Membership` FAIL; exit 1.
  - `mutant-cli.txt`: 3 `TestTempOriginGuidance_*` PASS, exit 0.
  - Mutant reverted; `internal/kanban/temp_origin.go` diff empty.

## Baseline-attribution

This worktree at `7da398808` (+ transient mutant, reverted), go1.26.8, this run.

## Gaps

- Linux cell not measured: there `os.TempDir()` is `/tmp`, so the same mutant keeps `/tmp` and drops
  `/var/folders`; the AC gap is the same but the observed member loss differs.
- AC-THG-007's mutant (constant-false discriminant) was not re-run; it targets a different shape.

## Residual-risk

- Repair is an acceptance.md body amendment on a completed SPEC (manager-spec amendment path,
  HISTORY `## Amendments`). The shape — extend AC-THG-006 with a production-set positive clause vs.
  add a new AC — is a design decision, not taken here.
