# Plan — SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001

> **Status**: draft (plan-phase only — no code, no commits). `cycle_type: tdd` (RED → GREEN → REFACTOR).
> **Baseline**: HEAD `734ede821` on `main`. `origin/main` is `54d748ddf` (1 commit ahead — gitignore #1499, unrelated). A rebase/ff to `origin/main` happens before run-phase entry on a feature branch; plan artifacts are unaffected.

## §A. Context

This SPEC closes the last hole in the SPEC-V3R6-LIFECYCLE-REDESIGN-001 3-phase lifecycle redesign. The redesign migrated `audit.go` + `migrate_3phase.go` to the 3-marker predicate (`§E.2` + `§E.4` + `sync_commit_sha`) and retired `§E.5`, but `closer.go` was missed: Precondition 2 (`closer.go:696-699`) still requires a `§E.5` section, so 3-phase SPECs (which honor the schema by omitting `§E.5`) are structurally blocked from closing.

The fix is a localized relaxation modeled on the dual-acceptance pattern already shipped in `transitions.go:79-92` (`closeInfixMatch` accepts both the new `3-phase close` and the legacy `4-phase close` infix). We OR-in a 3-phase path into Precondition 2; the legacy `§E.5` path stays accepted.

**Files in scope (run-phase)**: `internal/spec/closer.go`, `internal/spec/closer_test.go`. Nothing else.

## §B. Known Issues (verified this session)

| ID | Site | Issue |
|----|------|-------|
| K1 | `closer.go:696-699` | Precondition 2 hard-requires `state.HasMxSection` → blocks 3-phase SPECs. |
| K2 | `closer.go:571,611` | `HasMxSection` is populated solely from `§E.5`; no 3-marker sibling field exists. |
| K3 | `closer.go:350` | Commit-subject template uses `"4-phase close"`; canonical is `"3-phase close"` (REQ-LR-020). |
| K4 | `closer.go:107` | Doc comment: `"Close orchestrates the atomic 4-phase close transition for a SPEC."` |
| K5 | `closer.go:683` | Doc comment: `"validatePreconditions checks the 4-phase precondition matrix per AC-LSG-006."` |
| K6 | `closer.go:90` | Sentinel-error comment references `§E.5` as the exemplar failure. |
| K7 | `closer_test.go:47-86` | `TestClose_PreconditionMissingMx` asserts the `§E.5`-absent case FAILS — this is the TDD RED→GREEN pivot point. |
| K8 | `closer_test.go` (whole file) | `§E.4` marker count = **0** (grep-verified: `§E.2`=31, `§E.5`=33, `§E.4`=0). The current test suite covers ONLY the `§E.5`-based legacy path; the 3-phase (`§E.4` + `sync_commit_sha`) path has zero test coverage. This SPEC's M1 introduces that coverage for the first time. The `TestClose_PreconditionMissingMx` fixture (L51-58) exemplifies the gap: `§E.2` + `sync_commit_sha` only, no `§E.4` — insufficient under the proposed predicate. |

## §C. Pre-flight (zero-code verification, run before M1)

| Check | Command | Expected |
|-------|---------|----------|
| Baseline green | `go test ./internal/spec/...` | PASS (current suite, including the about-to-be-converted test) |
| Build clean | `go build ./...` | exit 0 |
| Scope guard — era.go untouched baseline | `grep -n '§E.5' internal/spec/era.go` | ≥4 matches (the H-4-legacy predicate stays) |
| Scope guard — transitions.go untouched | `grep -n 'closeInfix3Phase\|closeInfix4Phase' internal/spec/transitions.go` | both present (dual-infix intact) |
| Parent contract | `grep -n 'REQ-LR-020\|AC-LR-012' .moai/specs/SPEC-V3R6-LIFECYCLE-REDESIGN-001/spec.md` | matches confirm parent reconciliation |

## §D. Constraints

- **TDD**: RED first. The conversion of `TestClose_PreconditionMissingMx` (and the new 3-phase PASS test) MUST be written and confirmed FAILING against current `closer.go` BEFORE the Precondition 2 edit. RED evidence (test name + failure excerpt) recorded in §E.1.
- **Backward-compat absolute**: the legacy `§E.5`-present path MUST stay accepted. The relaxation is an OR (`HasMxSection OR (HasSyncSection3Marker AND sync_commit_sha != "")`), not a replacement.
- **Scope local**: do NOT touch `era.go`, `audit.go`, `transitions.go`, `spec-frontmatter-schema.md`. The §E.5 token stays load-bearing in `era.go`.
- **Style**: `fmt.Errorf("...: %w", err)` wrapping; English comments; match surrounding naming; no new exported symbols unless required.
- **No time estimates** — milestones are priority-ordered, not duration-ordered.

## §E. Self-Verification (TDD RED evidence captured here at run-phase)

### §E.1 Plan-phase Audit-Ready Signal

_This section is populated by manager-spec at plan phase. RED evidence (the failing-test excerpt from M1 step 1) is appended by manager-develop at run phase as the first §E.1 entry, because the RED observation requires running the converted test against unmodified `closer.go` — a run-phase act._

_PLAN-PHASE PLACEHOLDER — RED evidence pending M1 step 1 execution._

### §E.2 Run-phase Evidence

_<pending run-phase>_

### §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

### §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F. Milestones (priority-ordered — highest-change-likelihood first)

### M1 — Precondition 2 relaxation + TDD RED→GREEN (Priority High)

The decision most likely to change shape is the Precondition 2 predicate itself, so it goes first.

**Step 1 (RED)**:
- Add a new test `TestClose_ThreePhasePreconditionPass` asserting a 3-phase `progress.md` (`§E.4` + `sync_commit_sha`, `§E.5` absent) passes Precondition 2 (no `"missing §E.5"` failure, close proceeds).
- Convert `TestClose_PreconditionMissingMx` (`closer_test.go:47-86`): the current fixture's `§E.5`-absent case must now PASS (rename to `TestClose_ThreePhasePreconditionPass_Legacy` or split — see Step 3). Add a sibling `TestClose_LegacyMxSectionStillAccepted` asserting a legacy `§E.5`-present `progress.md` still closes.
- Run `go test ./internal/spec/ -run 'TestClose_ThreePhasePreconditionPass|TestClose_LegacyMxSectionStillAccepted' -v` and **confirm FAIL** (the 3-phase case fails because Precondition 2 still requires `§E.5`). Record the RED excerpt in §E.1.

**Step 2 (GREEN)**:
- At `closer.go:696-699`, widen Precondition 2 to dual-acceptance:
  - Legacy path: `state.HasMxSection` → pass.
  - 3-phase path: `§E.4` sync marker present (`hasProgressMarker(body, "§E.4")`) AND `sync_commit_sha` non-empty → pass.
- Consider adding a sibling `closeState` field (e.g. `HasSyncMarker bool` / `SyncMarkerSection bool`) populated at `closer.go:610-613` alongside `HasMxSection`, OR compute inline in `validatePreconditions` from the already-loaded `state.ProgressMDContent`. Inline is simpler (no struct change) — prefer it unless the field is needed elsewhere.
- Re-run the RED tests; confirm GREEN.

**Step 3 (REFACTOR)**:
- Rename / split the converted `TestClose_PreconditionMissingMx` into two intent-clear tests: one for the 3-phase PASS path, one for the legacy `§E.5`-present PASS path. Keep a negative test (`TestClose_PreconditionMissingSync` at L89 — `§E.2` absent) unchanged; it still legitimately fails.
- **The existing fixture MUST add a `## §E.4 Sync-phase Audit-Ready Signal` marker to uphold the PASS assertion** — the current fixture (closer_test.go:51-58) carries only `## §E.2 Sync-phase` + `sync_commit_sha: abc1234`, which is NOT sufficient under the proposed 3-phase predicate (`§E.4` marker AND `sync_commit_sha != ""`). Without the `§E.4` marker the converted test would still fail Precondition 2 for the wrong reason (3-phase path false on `§E.4` absence), producing a broken test. This is plan-auditor D1 (BLOCKER).
- Canonicalize the subject template (`closer.go:350`) to `"3-phase close"` (REQ-LSG-004).
- Canonicalize the doc comments at `closer.go:107` and `closer.go:683` from `"4-phase"` to `"3-phase"` (REQ-LSG-005).
- Update the `closer.go:90` sentinel-error comment to reference the 3-phase precondition shape rather than singling out `§E.5`.

### M2 — Regression + scope-guard verification (Priority High)

- `go build ./...` exit 0 (AC-LSG-005).
- `go test ./internal/spec/...` green (AC-LSG-005).
- `go test ./...` full suite green (AC-LSG-005).
- Scope guard: `grep -c '§E.5' internal/spec/era.go` unchanged from §C baseline (AC-LSG-003).
- Scope guard: `grep -n 'closeInfix3Phase\|closeInfix4Phase' internal/spec/transitions.go` — both still present (AC-LSG-003).
- Run `go test ./internal/spec/ -run 'TestClose_BackfillOnly_CompletedProductionVariantsAreNoOp|TestClose_FullClose_AtomicRollbackOnFailure|TestClose_FullClose_ProducesCommit' -v` to confirm the no-op invariant (AC-LSG-018), the atomic-rollback invariant (AC-LSG-014), and the full-close commit emission all remain green — these are the load-bearing existing behaviors the relaxation must not disturb.

### M3 — DEFERRED validation record (Priority Low)

- Record (do NOT execute) the mo.ai.kr validation command and expected output in `acceptance.md` AC-LSG-007:
  - `make build` reinstall of `~/go/bin/moai`.
  - In the mo.ai.kr checkout: `moai spec close SPEC-DB-CLEANUP-001 --dry-run` → expected precondition PASS (no `"missing §E.5"` failure).
- This milestone produces no code and no commit; it only finalizes the DEFERRED record.

## §G. Anti-Patterns

- **AP-1**: Purging `§E.5` from `era.go` / `transitions.go` / the parser. `§E.5` is load-bearing for legacy era classification (H-4-legacy predicate) and for recognizing historical close commits in git history. The relaxation is local to `closer.go`. (REQ-LSG-003)
- **AP-2**: Replacing the `§E.5` path instead of OR-ing a 3-phase path. Backward-compat MUST be preserved — grandfather-era SPECs that still carry `§E.5` must continue to close. (REQ-LSG-002)
- **AP-3**: Skipping the RED step. `cycle_type=tdd` requires the converted/new test to be confirmed FAILING against current `closer.go` before the Precondition 2 edit. RED evidence is recorded in §E.1. (REQ-LSG-006)
- **AP-4**: Canonicalizing L350 to `"3-phase close"` without also updating L107/L683/L90 — leaves the file's prose self-contradictory. All three comment sites + the subject template go in the same run-phase pass. (REQ-LSG-004/005)
- **AP-5**: Widening `closeState` with a new exported field when an inline check in `validatePreconditions` would do. Prefer the simpler shape (Agent Core Behavior #4 — Enforce Simplicity).

## §H. Cross-References

- **Parent**: `.moai/specs/SPEC-V3R6-LIFECYCLE-REDESIGN-001/` — owns the 3-marker predicate contract, REQ-LR-020/021, AC-LR-012, D4 reconciliation. Authoritative for the `3-phase close` infix convention.
- **Sibling**: `.moai/specs/SPEC-V3R6-LIFECYCLE-SYNC-GATE-001/` — owns AC-LSG-006/014/018/020/022 (the original closer preconditions + truth-table invariants this relaxation must preserve).
- **Schema**: `.claude/rules/moai/development/spec-frontmatter-schema.md` — `§E.5` retirement (folded into `§E.4`), parser-load-bearing legacy recognition.
- **Source**: `internal/spec/closer.go` (L90, L107, L350, L571, L611, L683, L696-699), `internal/spec/transitions.go` (L79-92), `internal/spec/era.go` (L135-167), `internal/spec/audit.go` (L359-384).
