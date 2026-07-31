# SPEC-UPDATE-CI-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set).
- Baseline tree: `d5336214e` on `plan/epic-update-config-audit` (merged with `origin/main`).
  Authored at worktree `9426bf49b`, whose merge-base with `origin/main` was `76d9a8f3b`; the local
  branch `main` at `1d4e4f7da` was a stale divergent branch, not this branch's base.
- Findings F1-F5 each re-verified against this tree while authoring. Four drifts recorded
  (spec.md §A.6): the `ci.yml:222` job attribution (test-integration, not build); the `internal/cli`
  75.6% / 75.7% two-tree coverage split — **RESOLVED** by the `origin/main` merge, which collapsed
  the two trees into one baseline measuring 75.7%; the neutrality workflow running three isolated
  test targets rather than one; and two independent `C<N>` class-numbering schemes across the two
  guard files.
- F3 re-measured at `d5336214e`: `go test -cover ./internal/cli/... ./internal/config/...` —
  `internal/cli` 75.7%, `internal/config` 80.5%, `internal/cli/specid` 58.3%, six update
  subpackages 88.6-97.5%. No `codecov.yml` / `.codecov.yml` at repo root.
- F4 independently re-measured: `go tool cover -func` over `internal/cli/update/backup/` — all five
  `merge.go` functions at 100.0%, package total 88.6%.
- F5 pattern set verified against implementation: `internal_content_leak_test.go:171` C1 =
  `SPEC-(V3R[2-6]|AGENCY|WORKTREE)-`, `:282` C1c = `SPEC-(DB-SYNC-RELOC|PROJECT-DB-HINT)-[0-9]{3}`;
  `template_neutrality_audit_test.go:168` C6 = `PR #[0-9]+`. All three §A.5 shapes escape.
- SPEC ID regex self-check executed as Bash: `PASS`.
- Requirements: 18 REQ + 4 NFR across §B.1-§B.6. Acceptance criteria: 23 (14 behavioural,
  9 presence-only, every presence criterion paired). Falsification procedures: 5.
- Status: `draft`. Awaiting plan-audit and Implementation Kickoff Approval.

### Plan-audit revision (v0.2.0) — D1-D11 resolved, D12-D16 deferred

Verdict being corrected: **FAIL, 0.63** (Clarity 0.60 / Completeness 0.90 / Testability 0.45 /
Traceability 0.75). Must-Pass 7/7 passed; the failure was entirely in verification design. Two
audit-confirmed strengths were preserved: `-run` selector vacuity remains 0, and every recorded
numeric baseline still reproduces.

| Defect | Severity | Resolution |
|---|---|---|
| D1 | critical | §C-1 rebuilt on a scratch worktree — `go test -overlay` was measured not to reach runtime `os.ReadFile` |
| D2 | critical | §C-2 retargeted to a branch that exists (`merge.go:87`); the old `sed '/IsZero()/,+2d'` matched 0 lines |
| D3 | critical | §C-3/-4/-5 mutation steps converted from shell comments to executable commands; every §C procedure gained a `diff -q` no-op guard |
| D4 | critical | zero-value skip re-attributed from `backup/merge.go` to `internal/config/merge.go` (`MergeAll:149`, `isZero:200`) |
| D5 | major | M4 expected-failure set re-measured — **corrected against the audit** (below) |
| D6 | major | REQ-UCG-004 / plan M2 / AC-006 reconciled on the `internal/cli/update` prefix |
| D7 | major | AC-015/-019/-023 diff base `main` → `d5336214e` (`main` is not an ancestor of HEAD) |
| D8 | major | AC-010's EOF-running `sed` range → bounded `awk` extractor with positive and negative controls |
| D9 | major | tolerance made mechanically discriminable: per-candidate assertion table + 0.1pp bump |
| D10 | major | AC-017 branch-conditioned; it was vacuous on the enumeration branch |
| D11 | major | AC-020 scoped off class-definition lines and historical SPEC artifacts (103 → 77) |

Deferred to a later revision, unchanged here: **D12** (NFR-UCG-003 unbound, NFR-UCG-002 half-bound),
**D13** (AC-007's `gh run list --limit 1` does not pin the target PR), **D14** (AC-012 depends on a
future merged PR), **D15** (`SPEC-V3R6-CI-PR-SPEEDUP-001` absent from `.moai/specs/`), **D16**
(spec.md §A.1 cites `ci.yml:184` where the `if:` is at `:200`).

**One audit finding was itself incomplete.** The audit rejected M4's original premise correctly —
the three zero-value rows do *not* fail — but named `nested_old_only` as the single real failing
case. Re-running the matrix against `d5336214e` shows **two** rows fail: `user_only_key` is dropped
for the same reason, because `DeepMerge3Way` iterates only `newMap` (`merge.go:53`) and never visits
a key absent from the new template. Adopting the audit's single-case remediation verbatim would have
left `user_only_key` declared a hard assertion, to fail unexpectedly at run time. The corrected SKIP
set is exactly `{user_only_key, nested_old_only}`.

Every falsification procedure in this revision was executed or its mechanism verified before being
adopted: the overlay/runtime-read limitation, the scratch-worktree lifecycle, the `merge.go:87`
anchored mutation, the `awk` job-block deleter, the `sed` pattern rewrite, and the bounded filter
extractor were each run and their output recorded in the artifacts.

### Epic run order (dependency sequencing)

`SPEC-UPDATE-CI-GUARD-001` declares `depends_on: []` — no hard dependency edge — so its run-phase
`Depends_on Pre-flight Check` is trivially satisfied. Every SPEC in this Epic is currently `draft`,
which is the expected state for an Epic whose members have not yet run; it is a sequencing fact, not
a per-SPEC defect.

| Order | SPEC | Constraint |
|---|---|---|
| 1 | `SPEC-UPDATE-REINSTALL-LOOP-002` | Epic entry point |
| 2 | `SPEC-UPDATE-DATA-SURVIVAL-001` | `depends_on: [E1]` |
| 3 | `SPEC-CONFIG-TIER-PERSIST-001` | `depends_on: [E2]` |
| 4 | `SPEC-CONFIG-KEY-HONESTY-001` | `depends_on: [E3]` |
| 5+ | **this SPEC**, `SPEC-UPDATE-YAML-PRESERVE-001`, `SPEC-UPDATE-DOC-DRIFT-001` | no `depends_on` edge |

Two **soft** ordering constraints bind this SPEC without being frontmatter dependencies, and both
are handled inside the SPEC rather than by a dependency edge:

- **M5 versus `SPEC-CONFIG-KEY-HONESTY-001`.** The extended neutrality patterns will match the three
  leaks that SPEC's REQ-CKH-012 removes. Landing M5 first would turn the trunk red. plan.md §F M5
  resolves this by shipping the new classes in the advisory-WARN tier and promoting them to
  binary-FAIL once the shipped tree is clean — so this SPEC's completion does not depend on another
  SPEC's run-phase.
- **M4 versus `SPEC-UPDATE-YAML-PRESERVE-001`.** The guard's two known-defective rows are committed
  as explicitly-expected failures naming that SPEC, flipping to hard assertions when it lands. The
  guard is deliverable before the fix.

Do **not** invoke `/moai run` on this SPEC with `--ignore-deps`. The dependency is satisfied by the
sequencing above; a bypass would substitute an override for an ordering that costs nothing to honour.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
