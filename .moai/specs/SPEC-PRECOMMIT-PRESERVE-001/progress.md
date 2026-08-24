# SPEC-PRECOMMIT-PRESERVE-001 — Progress

Card: t230 · Tier M · Class C · branch `WT-precommit-preserve`

## §E.1 Plan-phase Audit-Ready Signal

- **Artifacts**: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` — created at plan-phase,
  `status: draft`.
- **Baseline tree**: `294b4b6ab` (worktree `.claude/worktrees/t230`, divergence `0 0`).
- **Premises verified in this tree, not inherited on trust**: the missing byte-comparison and backup
  (`hook_install_precommit.go:139-151`), the disguised success line (`:172`), the call sites
  (`update_template_sync.go:575`, `init.go:898`), the parity test
  (`hook_install_precommit_test.go:38`), and the zero-hit `exec` redirection sweep (spec.md §A.4).
- **Decisions taken**: provenance sidecar over in-body stamp; back-up-and-overwrite over preserve
  (spec.md §A.5, both with rejected alternatives stated).
- **Counts**: 13 REQ, 13 AC — within the Tier M ceiling of 16/16. Four AC are labelled
  behaviour-preserving; the remaining nine fail against the untouched tree.

_<pending plan-audit>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
