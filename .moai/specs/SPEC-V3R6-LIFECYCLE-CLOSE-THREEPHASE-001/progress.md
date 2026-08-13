# Progress — SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001

> Canonical §E section skeleton. Placeholder headings only — §E.2/§E.3/§E.4 are populated by manager-develop (run) and manager-docs (sync). §E.1 is the plan-phase audit-ready signal, populated by manager-develop at run-phase M1 step 1 (the TDD RED observation requires running the converted test against unmodified `closer.go`).

## §E.1 Plan-phase Audit-Ready Signal

_PLAN-PHASE PLACEHOLDER — RED evidence pending M1 step 1 execution (run phase)._

## §E.2 Run-phase Evidence

**Run-phase HEAD**: `617a8418c` (M2 `617a8418c` stacked on M1 `1199e3b58`, on plan-phase `0b932fe13`). Branch `feat/SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001`.

### §E.2.1 AC PASS matrix (manager-develop §E1, coordinator trust-but-verify PASSED)

| AC | Status | Verification command | Observed output |
|----|--------|---------------------|-----------------|
| AC-LSG-001 | PASS | `go test ./internal/spec/ -run TestClose_ThreePhasePreconditionPass -v` | `--- PASS: TestClose_ThreePhasePreconditionPass` (exit 0) |
| AC-LSG-002 | PASS | `go test ./internal/spec/ -run TestClose_LegacyMxSectionStillAccepted -v` | `--- PASS: TestClose_LegacyMxSectionStillAccepted` (exit 0) |
| AC-LSG-003 | PASS | `grep -c '§E.5' internal/spec/era.go` baseline-vs-post-edit | 9 matches unchanged (scope guard — `era.go` NOT touched by run-phase diff) |
| AC-LSG-004 | PASS | `grep -c '4-phase' internal/spec/closer.go` → 0; `grep -n '3-phase close' internal/spec/closer.go` → L352 subject template + L109/L685/L698 doc comments | 0 residual `4-phase`; 7 `3-phase` matches; L352 = `chore(%s): Mx-phase audit-ready signal + 3-phase close` |
| AC-LSG-005 | PASS | `go build ./... && go test ./...` | both exit 0 (full suite green; `era.go`/`audit.go`/`transitions.go` out-of-scope confirmed) |
| AC-LSG-006 | PASS | TDD RED→GREEN pivot (see §E.2.2) | RED captured pre-GREEN; GREEN after Precondition 2 relaxation |
| AC-LSG-007 | DEFERRED | `moai spec close SPEC-DB-CLEANUP-001 --dry-run` (mo.ai.kr checkout, post-merge) | DEFERRED — command + expected output recorded in acceptance.md §C AC-LSG-007; requires `make build` reinstall + checkout switch |

### §E.2.2 E8 RED verbatim (TDD pre-GREEN failure, captured before the Precondition 2 edit)

```
=== RUN   TestClose_ThreePhasePreconditionPass
--- FAIL: TestClose_ThreePhasePreconditionPass (0.00s)
    closer_test.go:67: Precondition 2 should PASS for a 3-phase SPEC; got ErrPreconditionMissing: ... missing §E.5 Mx-phase audit-ready signal in progress.md
```

**GREEN result (after Precondition 2 relaxation to additive OR — legacy §E.5 OR 3-phase §E.4+sync_commit_sha)**: `TestClose_ThreePhasePreconditionPass` + `TestClose_LegacyMxSectionStillAccepted` both PASS. The conversion of `TestClose_PreconditionMissingMx` (which previously asserted the §E.5-absent case FAILS) into `TestClose_ThreePhasePreconditionPass` (3-phase PASS) plus the sibling `TestClose_LegacyMxSectionStillAccepted` (legacy backward-compat PASS) is the REQ-LSG-006 TDD pivot.

### §E.2.3 Coverage + scope guards (this run, this tree, HEAD `617a8418c`)

- **Coverage**: `go test -cover ./internal/spec/...` → **89.1%** (≥85% threshold).
- **golangci-lint**: `golangci-lint run --timeout=2m ./internal/spec/...` → 0 issues.
- **Subagent boundary (C-HRA-008 family)**: `grep -rn 'AskUserQuestion' internal/spec/ | grep -v "_test.go" | grep -v "// "` → 0 matches.
- **Full suite**: `go test ./...` → exit 0.
- **Scope guard — era.go untouched**: `grep -c '§E.5' internal/spec/era.go` = 9 (H-4-legacy predicate intact; baseline preserved).
- **Scope guard — transitions.go dual-infix intact**: `closeInfix3Phase` + `closeInfix4Phase` both present (`internal/spec/transitions.go:80-81`).

### §E.2.4 Files modified (run-phase scope, `internal/spec/` only)

- `internal/spec/closer.go` (+27 lines): Precondition 2 relaxed to additive OR (legacy `§E.5` OR 3-phase `§E.4 + sync_commit_sha`); "4-phase"→"3-phase" canonicalized at L91/109/352/685/698/701.
- `internal/spec/closer_test.go` (+79 lines): `TestClose_ThreePhasePreconditionPass` added (3-phase PASS); `TestClose_LegacyMxSectionStillAccepted` added (legacy backward-compat PASS); `TestClose_PreconditionMissingMx` converted (the §E.5-absent case is now a PASS assertion).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-08-13
run_commit_sha: 617a8418c
ac_must_pass_count: 6
ac_must_pass_passing: 6
ac_deferred_count: 1
ac_deferred_ids: [AC-LSG-007]
coverage_pct: 89.1
lint_status: clean
full_suite: pass
scope_guard_era_go: intact
scope_guard_transitions_go: intact
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-08-13
sync_commit_sha: pending-backfill-sync
changelog_entry_position: Fixed
frontmatter_status_transitions:
  draft_to_in_progress: missed-on-M1 (recorded-as-gap; M1 commit subject matches canonical `fix(SPEC-{ID}): M1 ...` owner pattern but frontmatter status was not advanced to in-progress at M1)
  in_progress_to_implemented_to_completed: carried-on-sync-commit (manager-docs owns the terminal `completed` transition on the single sync commit per the 3-phase close)
canary_compliance_check:
  changelog_duplicate_pre_emission_grep: 0
  ac_count_acceptance_md: 7
  ac_count_changelog_entry: 7
```

**D3 backfill exemption note**: `sync_commit_sha` is self-referential in this commit (a commit cannot know its own SHA until after it lands). Per `.claude/rules/moai/development/spec-frontmatter-schema.md` § SHA placeholder backfill exemption (D3), the placeholder `pending-backfill-sync` is written here and backfilled to the real SHA in a follow-up commit after the sync PR merges.

**Gap — missed `draft → in-progress` ownership step**: spec.md frontmatter was left at `status: draft` by the run phase (M1 did not advance it to `in-progress`). The M1 commit subject `fix(SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001): M1 relax Precondition 2 to 3-phase close` DOES match the canonical `draft → in-progress` owner pattern (`fix(SPEC-{ID}): M1 ...` per the Status Transition Ownership Matrix), so the OwnershipTransitionRule subject-prefix check is satisfied; the gap is the missing frontmatter status advance, not a wrong owner. This sync commit carries the full `draft → in-progress → implemented → completed` transition atomically (manager-docs owns the terminal `completed` on the single sync commit). Run history is NOT rewritten — the missed step is recorded here, not back-dated into the M1 commit.
