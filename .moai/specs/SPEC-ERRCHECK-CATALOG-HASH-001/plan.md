# SPEC-ERRCHECK-CATALOG-HASH-001 — Implementation Plan

## §A Context

CI (run 34014859880, job Lint, golangci-lint v2.1.6) and a local run (v2.10.1, same committed
`.golangci.yml`) both report exactly one finding:
`internal/template/catalog_tree_hash.go:60:14: Error return value of `fmt.Fprintf` is not checked (errcheck)`.
Local verbatim summary: `1 issues: * errcheck: 1`.

The call writes into `h := sha256.New()` (line 58). `sha256.New` returns `hash.Hash`, whose
`Write` is documented in the Go standard library as never returning an error. The unchecked error
is structurally always nil.

## §B Known Issues

- The finding blocks the CI Lint job on every push until repaired.
- The defect was introduced by the digest-input write; no other call sites in the package carry
  the same shape (measured: the local run reports errcheck count 1 for the whole package).

## §C Pre-flight

- [x] Worktree confirmed: `.claude/worktrees/t489`, branch `WT-errcheck-catalog-hash`, base `615d18c1f`.
- [x] `.golangci.yml` pins `errcheck.check-blank: false` (lines 31-35) — verified in this worktree.
- [x] RED baseline measured by lane-15; artifact path declared as `.moai/reports/t489/red-baseline.md`.

## §D Constraints

- Single-line repair at `internal/template/catalog_tree_hash.go:60`; no other Go source changes.
- Comment citing the `hash.Hash` never-error contract, in English (repo code-comment convention).
- `.golangci.yml` MUST NOT change.
- Verification is the linter only; no new tests (behavior-preserving by construction — REQ-005).
- No full test-suite runs; scoped verification per lane-local verification discipline.
- Evidence under `.moai/reports/t489/`; every commit message carries card id `t489`.

## §E Self-Verification

Run-phase closure evidence (E1-E7) to be produced by manager-develop:

- E1: `golangci-lint run ./internal/template/... --timeout=2m` — errcheck match count 0 (AC-GREEN).
- E2: mutant check — restore the pre-fix blob, re-run, errcheck match count 1 (AC-MUTANT); restore
  the fix afterwards.
- E3: `git log` — baseline commit precedes fix commit (AC-ORDERING).
- E4: scope check — no new test files, no `.golangci.yml` diff, and the diff shape matches the
  sanctioned `_ =` discard form with no `//nolint` directive in
  `git diff 615d18c1f..HEAD -- '*.go'` (AC-SCOPE).
- E5: evidence files present under `.moai/reports/t489/`; every branch commit message contains
  `t489` (AC-EVIDENCE).

## §F Milestones

### M1 — Single-line discard fix (Priority High)

1. Verify the RED baseline artifact (`.moai/reports/t489/red-baseline.md`, lane-provided) is
   committed in a commit that precedes the fix commit (AC-ORDERING). If it is not yet committed,
   commit it first with the card id in the message, before touching Go source.
2. Apply the fix at `internal/template/catalog_tree_hash.go:60`:
   `_ = fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)` plus a brief English comment citing the
   `hash.Hash` never-error contract.
3. Run AC-GREEN (errcheck count 0), then AC-MUTANT (revert blob → count 1 → restore fix), judging
   on match counts, not exit codes.
4. Export evidence to `.moai/reports/t489/` and commit with explicit pathspecs; message carries
   `t489`.

Milestone ordering follows decision-reversibility: the baseline-first commit ordering (AC-ORDERING)
is the only irreversible-shaped decision and leads the milestone; the mechanical edit follows.

## §G Anti-Patterns

- Do NOT propagate or wrap the error (dead code path — REQ-003).
- Do NOT add `//nolint:errcheck` or a `.golangci.yml` exclusion — both silence the finding without
  documenting the reason at the call site.
- Do NOT widen the fix to other files; the measured finding set is exactly one.

## §H Cross-References

- `.claude/rules/moai/core/verification-claim-integrity.md` §2.3 — ordering attribution (AC-ORDERING).
- `.golangci.yml` lines 31-35 — errcheck settings pin (`check-blank: false`) and the documented
  local/CI version-skew rationale.
- `.moai/reports/t489/red-baseline.md` — RED baseline artifact (lane-measured, this run).
