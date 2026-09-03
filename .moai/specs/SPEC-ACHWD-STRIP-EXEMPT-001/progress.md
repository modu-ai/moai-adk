# SPEC-ACHWD-STRIP-EXEMPT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Plan authority: `SPEC-HOOK-WIRING-DRIFT-001` plan.md §I (iteration-2 patch
  `8eb5b9102`).
- Plan audit: PASS-WITH-DEBT 0.80 (Tier M threshold met marginally), verdict
  `.moai/reports/t469/plan-audit.md` @ `7a8ae6744`; D1–D5 patched, confirming
  check textual only.
- Implementation Kickoff Approval: operator approval 2026-09-03, relayed by
  the lead; run phase GO.
- Pre-run absorb: local develop `6765a75c0` merged into the card branch
  (merge `a122e7568`, one conflict in CHANGELOG.md, both t216/t456 entries
  kept). All run work is on top of `a122e7568`.
- Scope statement: this run's file deltas are `.moai/specs/**` only, so the
  verification scope (file-delta packages ∪ reverse-dep packages via
  `go list -deps`) contains **no Go packages** — no package tests, no
  `go vet`, and no build are in scope. The §I.4 perl command and the spec
  lint are the entire check set. No local full suite is run.
- A5 verification outputs (measured 2026-09-03 on the amended tree, worktree
  `t469`):

  **A5a — the §I.4 A4 strip-aware mirror check:**

  ```
  $ perl -e 'local $/; my $rc=0; for my $f (@ARGV){ ... } exit $rc' \
    .claude/rules/moai/development/hook-independence.md \
    .claude/rules/moai/core/agent-common-protocol.md \
    .claude/rules/moai/core/agent-common-protocol-reference.md
  (no output)
  A5-mirror-check-exit=0
  ```

  **A5b — the spec lint:**

  ```
  $ go run ./cmd/moai spec lint .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md
  SEVERITY  CODE                FILE                                            LINE  MESSAGE
  --------  ----                ----                                            ----  -------
  WARNING   ModalityMalformed   .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  121   REQ REQ-HWD-001: EARS modality violation — SHALL missing or format mismatch
  WARNING   ModalityMalformed   .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  132   REQ REQ-HWD-003: EARS modality violation — SHALL missing or format mismatch
  WARNING   ModalityMalformed   .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  158   REQ REQ-HWD-005: EARS modality violation — SHALL missing or format mismatch
  WARNING   ModalityMalformed   .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  183   REQ REQ-HWD-009: EARS modality violation — SHALL missing or format mismatch
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  121   REQ REQ-HWD-001 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  125   REQ REQ-HWD-002 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  132   REQ REQ-HWD-003 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  138   REQ REQ-HWD-004 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  158   REQ REQ-HWD-005 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  165   REQ REQ-HWD-006 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  171   REQ REQ-HWD-007 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  177   REQ REQ-HWD-008 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  183   REQ REQ-HWD-009 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  187   REQ REQ-HWD-010 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  191   REQ REQ-HWD-011 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  197   REQ REQ-HWD-012 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  210   REQ REQ-HWD-013 is not referenced by any AC
  WARNING   CoverageIncomplete  .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md  225   REQ REQ-HWD-014 is not referenced by any AC

  0 error(s), 18 warning(s)
  A5-spec-lint-exit=0
  ```

  **Warning-composition note (AC-ASE-003's expectation was "18 pre-existing
  CoverageIncomplete"):** the total is 18 as predicted and 0 errors as
  required, but the composition is 4 `ModalityMalformed` + 14
  `CoverageIncomplete`. Verified pre-existing, not amendment-caused: the four
  ModalityMalformed lines are REQ-HWD-001/003/005/009 — wording the amendment
  never touched — and `git diff 3a3f51e83 6765a75c0 -- internal/spec/lint.go
  internal/spec/audit.go` is empty (the pre-run absorb did not change the
  linter), so the same composition existed at audit time; the audit verdict's
  "(all pre-existing CoverageIncomplete)" parenthetical was an imprecise
  summary of the same 18. All 18 `CoverageIncomplete` findings are the known
  cross-file spec.md↔acceptance.md resolution limitation the audit named.

- Payload application status: A1 (frontmatter + HISTORY row + Amendments
  sub-section), A2 (REQ-HWD-013 reword), A3 (§F Out of Scope H3), A4
  (AC-HWD-015 rewrite) — applied verbatim from §I.4. Wrapper status flipped
  `draft → in-progress` on this commit.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_commit_sha: 17b447240   # backfilled here; the sync commit cannot cite its
                             # own SHA (SHA placeholder backfill exemption,
                             # spec-frontmatter-schema.md § D3)
sync_complete_at: 2026-09-03
measured_at_head: 820db6cf9
```

What sync did (single sync commit, both closes riding it):

1. **Wrapper close** — `SPEC-ACHWD-STRIP-EXEMPT-001` frontmatter
   `in-progress → completed` (3-phase close convention: the `completed`
   transition rides the sync commit; manager-docs owns it per the Status
   Transition Ownership Matrix). No body content changed.
2. **Amendment re-close** — `SPEC-HOOK-WIRING-DRIFT-001` frontmatter
   `in-progress → completed` (v0.4.0 re-close, the path declared by the
   `re_close_path` row of its `## Amendments` sub-section: the SPEC returns to
   `completed` riding this card's sync commit). No body content and no other
   frontmatter field touched — the `updated:` date was already refreshed by
   the amendment commit and remains 2026-09-03.
3. **CHANGELOG** — one `[Unreleased]` Added row for this wrapper SPEC,
   adjacent to (above) the t216 row it amends; no separate row for the
   re-close (the amendment is named inside the wrapper row).

Verification (re-run at sync close, this tree):

- `go run ./cmd/moai spec lint .moai/specs/SPEC-ACHWD-STRIP-EXEMPT-001/spec.md`
  — output in the sync commit's evidence below; 0 errors expected.
- `go run ./cmd/moai spec lint .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md`
  — 0 errors, 18 warnings (4 `ModalityMalformed` + 14 `CoverageIncomplete`),
  both classes pre-existing per §E.1's composition note; unchanged from the
  run-phase A5b measurement.
- §I.4 strip-aware mirror check re-run on the final tree — exit 0 expected
  (verbatim output in `.moai/reports/t469/verdict.md`).
- Sync-auditor review is dispatched separately by the lane after this commit;
  the verdict at `.moai/reports/t469/verdict.md` carries a placeholder row
  for the auditor's co-authored score.

## Carried Debt (explicit carry-forward per lead instruction)

Source: `SPEC-HOOK-WIRING-DRIFT-001` plan.md §I.5 (residual risks recorded at
plan phase; plan-audit PASS-WITH-DEBT 0.80 was boundary-value — these are the
debts that survive the amendment).

| # | Debt | Owner |
|---|---|---|
| 1 | Normalization over-absorption at parenthetical granularity — a local-side and a template-side parenthetical containing a forbidden token can differ in non-token content invisibly to the mirror check | Neutrality doctrine (`.moai/docs/template-internal-isolation-doctrine.md` §25) / future fleet-wide strip-aware SPEC |
| 2 | Internal-date class is NOT normalized (no agreed regex) — a future date-mandated template strip will report a neutrality-mandated divergence as MISMATCH, the same false-FAIL shape one class over | Neutrality doctrine / future fleet-wide SPEC |
| 3 | The (ii-bare) absorption gap — a bare forbidden token inserted into the template copy passes the mirror check (exit 0) and is closed only by AC-HWD-016's template-side neutrality scan; either criterion alone does not close the space | AC-HWD-016 (standing compensating control) + future fleet-wide SPEC |
