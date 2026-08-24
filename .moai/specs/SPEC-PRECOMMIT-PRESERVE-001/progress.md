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
  stamp and over a full-content snapshot (the latter deferred, not dismissed — it changes no
  decision logic); back-up-and-overwrite over preserve, with the silence prohibition
  (REQ-PCP-006) written so no policy choice can trade it away. All with rejected alternatives stated
  (spec.md §A.5).
- **Counts**: 14 REQ, 14 AC — within the Tier M ceiling of 16/16, with no requirement dropped to
  stay under it. Three AC are behaviour-preserving; the remaining eleven fail against the untouched
  tree.
- **Every AC carries a mutant and a failing input** (acceptance.md § How to read a criterion). The
  card's headline mutant — a correct, recoverable, entirely silent overwrite — is defeated by
  AC-PCP-004/006/007, which assert the notice text rather than post-run file state.

_<pending plan-audit>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
