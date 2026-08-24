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
  v0.3.0, **Decision 3 — release composition**: M1+M2 ship with the hook body unchanged, M3 follows
  in a later release, bound mechanically by REQ-PCP-015 / AC-PCP-015. All with rejected alternatives
  stated (spec.md §A.5).
- **Counts**: 15 REQ, 15 AC — within the Tier M ceiling of 16/16, with no requirement dropped to
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

_<pending plan-audit iteration 2>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
