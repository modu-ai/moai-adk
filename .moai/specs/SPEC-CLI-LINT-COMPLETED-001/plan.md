# Plan — SPEC-CLI-LINT-COMPLETED-001

Implementation plan for adding `completed` to the `terminalStatusEnum` map in `internal/spec/lint.go`.
Tier S minimal — single milestone M1 (one-line map literal extension + characterization test).

## §A. Context

### §A.1 Ground Truth (re-verified by Read on 2026-07-07)

File: `internal/spec/lint.go`. Content-token anchors primary, line numbers secondary (line numbers verified against the working tree at the time of plan-phase authoring).

| Anchor | Line | Content |
|--------|------|---------|
| `terminalStatusEnum` declaration | 959 | `var terminalStatusEnum = map[string]bool{` |
| `superseded` entry | 960 | `"superseded": true,` |
| `archived` entry | 961 | `"archived":   true,` |
| `rejected` entry | 962 | `"rejected":   true, // future-proof: not currently used, kept for future extension` |
| Closing brace | 963 | `}` |
| `StatusGitConsistencyRule` type | 968 | `type StatusGitConsistencyRule struct{}` |
| `Check` entry | 972 | `func (r *StatusGitConsistencyRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {` |
| **Sole consumption** | **984** | `if terminalStatusEnum[fm.Status] { return nil }` |
| Drift detection | 996 | `if fm.Status != gitStatus {` |

Preceding the declaration at lines 957-958 are `@MX` annotations:

```
// @MX:NOTE: [AUTO] terminal lifecycle state — status values where a mismatch with git history is considered normal
// @MX:REASON: resolves Pattern D/E/F/G false positives from SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001; extend this map only when adding future states
```

The `@MX:REASON` explicitly invites extension for future states. `completed` is such a state.

### §A.2 Bug Mechanism

`StatusGitConsistencyRule.Check` (entry at L972) executes the following control flow:

1. L976-979 — empty `id` or `status` early-return.
2. **L984 — terminal-state early-return** (the gate this SPEC extends).
3. L989 — `getGitImpliedStatus(fm.ID)` invocation.
4. L996 — drift check: `if fm.Status != gitStatus { findings = append(...) }`.

A SPEC with `status: completed` reaches step 3 because `completed` is absent from `terminalStatusEnum`. The `getGitImpliedStatus` helper derives status from active-work commits in git history, which naturally ceased at the `implemented` transition (or earlier). The mismatch at L996 fires a `StatusGitConsistency` finding — a false positive.

**Fix**: add `"completed": true,` to the enum. The L984 early-return then short-circuits the drift check for `completed` SPECs, matching the existing pattern for `superseded`/`archived`/`rejected`.

### §A.3 Anchor Stability

- The L984 consumption is the SOLE read of `terminalStatusEnum` in the codebase (verified by file scan — `grep -n 'terminalStatusEnum' internal/spec/`).
- The map is the SOLE gatekeeper for terminal-state drift skipping.
- Adding a key to a Go `map[string]bool` literal is non-breaking — the existing 3 entries and their semantics are untouched. No call-site signature changes. No exported symbol changes.

## §B. Known Issues

None. This is a Type-1 chore (one-line map literal extension + one characterization test).

## §C. Pre-flight

- [x] Read `internal/spec/lint.go` lines 940-1010 (done — see §A.1 anchor table).
- [x] Verify `terminalStatusEnum` has no other consumers (done — sole consumer is L984).
- [x] Read `internal/spec/status_terminal_exemption_test.go` to identify the existing test pattern for `StatusGitConsistencyRule` and the existing Pattern D/E/F/G characterization tests (done at plan-phase iter-2 — 4 top-level functions confirmed, NO existing `rejected` test).
- [ ] Confirm `moai spec lint` is the canonical lint invocation and produces 0 findings on this SPEC after plan-phase authoring (deferred to plan-phase verification step).

## §D. Constraints

- Single-line production-code change in `internal/spec/lint.go` (add one map entry).
- 2 characterization tests in `internal/spec/status_terminal_exemption_test.go` (top-level functions `TestStatusGitConsistency_CompletedIsTerminal_Implemented` Pattern H + `TestStatusGitConsistency_RejectedIsTerminal_Implemented` Pattern I, following the existing `setupGitFixture` + `os.Chdir` fixture pattern).
- No new exported symbols.
- No changes to status enum validation.
- No changes to `StatusGitConsistencyRule.Check` control flow.
- Backward compatible: no behavior change for existing terminal states.

## §E. Self-Verification (Tier S)

Self-verification (E1-E7) is a manager-develop run-phase deliverable per the artifact ownership matrix. This section is intentionally a placeholder at plan-phase. See `acceptance.md` §D for the AC enumeration and `progress.md` §E.1 for the plan-phase audit-ready signal.

## §F. Milestones

### M1. Add enum entry + characterization test

**Scope** (single milestone — Tier S):

1. **Production change** — `internal/spec/lint.go`:
   - Insert `"completed": true,` as a new entry in the `terminalStatusEnum` map literal (between the `"rejected": true,` line at L962 and the closing brace at L963).
   - Update the inline comment on the `rejected` line only if it claims the set is exhaustive (it does not — the comment says "future-proof: not currently used, kept for future extension", which remains accurate for `rejected`; `completed` does not need a parallel comment unless manager-develop judges it helpful).
2. **Characterization test** — `internal/spec/status_terminal_exemption_test.go`:
   - Add a fifth top-level function `TestStatusGitConsistency_CompletedIsTerminal_Implemented` (Pattern H: frontmatter=`completed`, git-implied=`implemented`) asserting 0 findings, alongside the existing Pattern D/E/F/G functions in the same file.
   - Add a sibling `TestStatusGitConsistency_RejectedIsTerminal_Implemented` (Pattern I: frontmatter=`rejected`, git-implied=`implemented`) — currently NO `rejected` test exists (consistent with the L962 inline comment "future-proof: not currently used"). This expands the file from 4 to 6 top-level functions.

**TDD discipline (mandatory)**:

- **RED step**: write the `status: completed` characterization test FIRST. The test MUST use the `setupGitFixture(t, []fixtureCommit{...})` + `os.Chdir(dir)` + `t.Cleanup(...)` pattern (copied from the existing Pattern D/E/F/G tests in the same file) — this is REQUIRED, not optional. Without the git fixture, `getGitImpliedStatus` (`internal/spec/drift.go:194`) errors out → `Check()` L990-993 graceful-skip → 0 findings regardless of fix → AC-LC-003 RED proof becomes vacuous. Run `go test ./internal/spec/...` and confirm the new test FAILS with a non-zero `StatusGitConsistency` finding (the rule reaches L996 and emits drift). Capture the RED output verbatim — this is AC-LC-003 evidence.
- **GREEN step**: add the `"completed": true,` enum entry. Re-run the test and confirm it PASSES (0 findings).
- **REFACTOR step**: not applicable — the change is a one-line map literal extension with no structural improvement opportunity.

**Done criteria**:

- `go test ./internal/spec/...` passes (full package).
- `go test -race ./internal/spec/...` passes (map concurrent-read safety — read-only map, no race expected).
- `golangci-lint run ./internal/spec/...` clean.
- RED output captured pre-fix (AC-LC-003 evidence).
- Existing terminal-state behavior unchanged (AC-LC-002 evidence — existing tests pass without modification, or new tests added for the 3 prior states pass identically).

**Files touched (exhaustive)**:

- `internal/spec/lint.go` — one-line edit.
- `internal/spec/status_terminal_exemption_test.go` — 2 new top-level functions added (`TestStatusGitConsistency_CompletedIsTerminal_Implemented` Pattern H + `TestStatusGitConsistency_RejectedIsTerminal_Implemented` Pattern I), following the existing `setupGitFixture` + `os.Chdir` fixture pattern in the same file.

**Out of M1 scope**:

- No `internal/spec/audit.go` changes.
- No `internal/spec/era.go` changes.
- No CLI subcommand changes.
- No documentation changes (this is below the doc-sync threshold; manager-docs decides at sync-phase whether `CHANGELOG.md` warrants an entry — Tier S chore typically does not).

## §G. Anti-Patterns

- DO NOT modify `StatusGitConsistencyRule.Check` control flow beyond what the enum entry achieves implicitly.
- DO NOT add a `completed` special-case `if` branch — the enum at L984 is the contract; use it.
- DO NOT change severity (warning → error) or `--strict` behavior.
- DO NOT add `completed` to `StatusEnum` validation — it is already an accepted status value per `internal/spec/CLAUDE.md` § Status enum.
- DO NOT touch the `@MX:NOTE` / `@MX:REASON` annotations at L957-958 — they remain accurate (the `@MX:REASON` already invites future extension; this SPEC is precisely such an extension).
- DO NOT batch other terminal-state additions (e.g., speculative future values) into this SPEC.

## §H. Cross-References

- **Parent (provenance)**: `SPEC-CLI-SUBPKG-SPLIT-001` — Phase 1B follow-up; this SPEC is independent and separately closable.
- **Origin of `terminalStatusEnum`**: `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` — cited in the `@MX:REASON` annotation at `lint.go:958`.
- **Lifecycle redesign introducing `completed` as terminal**: `SPEC-V3R6-LIFECYCLE-REDESIGN-001` — retired the separate Mx-phase, folded into the 3-phase plan→run→sync close (the `completed` transition rides the sync commit).
- **Frontmatter schema SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md` (12 canonical fields + 8-value status enum).
- **Status enum + module conventions**: `internal/spec/CLAUDE.md` (8 values: draft, planned, in-progress, implemented, completed, superseded, archived, rejected).
- **Era classification SSOT**: `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (V3R6 = modern era subject to drift detection).
