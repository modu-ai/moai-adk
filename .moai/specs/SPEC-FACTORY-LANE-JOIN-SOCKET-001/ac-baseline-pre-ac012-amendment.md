# AC baseline snapshot — pre-AC-012-amendment state

Captured 2026-09-30 before the mid-run AC-012 wording amendment (D-NEW-1 re-delegation,
card t1330). This snapshot rides the SAME commit as the amendment per the
acceptance-amendment discipline (precedent: cards t1122 / t1148 / t1154).

- pre-amendment tree: `1b151622c`
- pre-amendment acceptance.md blob: `546ed8e29b0412e26e639e3a645dec88a4b768a8`
  (recoverable: `git show 546ed8e29b0412e26e639e3a645dec88a4b768a8`)
- amendment: AC-012's Then clause only — the literal "(generation 1)" hook-bind pin replaced
  by the measured bind ("pending+1 — launch-pending 1 → bind 2"); no other AC touched.
- measurement basis: dev-t1330's evidence note in progress.md §E.2 (referenced, not duplicated).
- AC set unchanged: AC-001..AC-017, same ids, same subjects, same requirement mapping.

## Pre-amendment AC-012 (verbatim)

### AC-012 — Lane env + hook bind into the resumed run

- **Given** a join performed through discovery (AC-001 shape),
- **When** the child session's env and its SessionStart hook bind are examined,
- **Then** the env carries `MOAI_KANBAN_ID` = the resumed run id and `MOAI_KANBAN_LEAD_NAME` =
  the target leader's name, and the hook binds the lane peer (generation 1) into that run's
  broker — the run identity and broker bind an ordinary join produces, with the leader name
  added by this SPEC on the discovery path (ordinary joins export no leader name and are
  unchanged).

## Pre-amendment §B matrix (ids + subjects)

AC-001 defect contrast / AC-002 zero-leader refusal / AC-003 explicit-id skips discovery /
AC-004 multi-leader ambiguity / AC-005 socket-not-liveness / AC-006 absent-row resume /
AC-007 retired-row resume + leader stamp / AC-008 resume event / AC-009 --lead default+target /
AC-010 legacy spelling refused / AC-011 --lead no-match refuses / AC-012 lane env + hook bind /
AC-013 mirror parity / AC-014 parse invariants / AC-015 auto-assignment / AC-016 docs+help /
AC-017 identity-proof decline branches.
