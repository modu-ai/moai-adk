# plan.md — SPEC-STAMP-REACHABILITY-001

Tier M · 3 milestones · execution order M1 → M2 → M3 · delivery via manager-develop (M1/M2/M3), manager-docs (sync). Version 0.1.0 · 2026-08-27.

## §A. Context

Follow-up F5 of card t279, promoted to card t291 (Class C: this lane carries plan→run→sync). Structural defect under squash merges: a codemaps stamp naming a branch-local HEAD is orphaned by the squash and main's graph-freshness check goes not-comparable → red for every inheriting PR. Emergency restamp (t279 M0) repaired the instance; REQ-GFR-014 froze the discipline; nothing enforces it mechanically. This SPEC adds the two selected countermeasures: the CI pre-merge reachability guard (enforcement at the only point that can catch an orphan before merge) and the explicit-commit stamping mode (ergonomics that make the right action easy).

Canonical problem record: `.moai/reports/t279/triage-table.md` §F5. Code/premise readings: research.md of this directory.

## §B. Known Issues

| # | Issue | Bearing on implementation |
|---|---|---|
| B1 | Unconditional HEAD recording in `baseProvenance` (provenance.go:185-203) is the orphan source — no alternative commit recordable today | M1 thread-through; flagless path byte-preserved (REQ-SR-007) |
| B2 | Generic exit-2 verdict lacks attribution ("which defect class?") | Guard annotations name sha + base ref; exit-2 path untouched |
| B3 | `--commit` + dirty-tree combination would silently blank one anchor (`omitempty` fields) | Pre-write rejection (REQ-SR-006), mirroring the atomic-install contract |
| B4 | Committed `provenance.json` `tree_root` names the t279 worktree path — looks alarming, is documented-correct (CR-R2; informational-only for tracked layers) | DO NOT "fix"; any deviation note refers here |

## §C. Pre-flight Checks

1. Worktree identity re-read immediately before any run-phase write (`git rev-parse --show-toplevel` = this worktree; branch `WT-stamp-orphan-guard`).
2. `git merge-base --is-ancestor $(jq -r '.commit_sha' .moai/project/codemaps/provenance.json) origin/main` still rc=0 before merging (guard-green baseline preserved through the delivery).
3. `0d15864ae90b` fixture availability re-checked (`git cat-file -e 0d15864ae90b^{commit}`); if a fetch policy ever pruned it, synthesize an equivalent orphan in the `/tmp` throwaway repo instead.
4. Affected-package test baseline captured BEFORE edits: `go test ./internal/cli/... ./internal/mx/... ./internal/graph/...` (targeted only — never full suite locally).
5. No parallel-session divergence on the shared checkout before first edit (`git rev-list --count --left-right origin/main...HEAD` after `git fetch origin main`; expected `0 N` from base da791eb0a).

## §D. Constraints

All constraints from spec.md §D bind the run phase; the load-bearing four:

1. **No new dependencies**, no schema_version change, no gate.yaml default changes.
2. **Stamp reachability HARD**: never refresh this PR's codemaps stamp against branch-local HEAD. Below the threshold the committed `410da655f…` stamp carries to merge unchanged (expected case). At-or-above the threshold (≥40 changed files, `count >= th.CodemapsChangedFiles`, measured `check.go:214`) NO in-SPEC recovery exists: a merge-base restamp re-counts the same churn as stale by construction, so the pre-merge codemaps red is accepted and carried until a POST-MERGE main restamp — a lead/operator act outside this SPEC's delivery scope.
3. **Targeted tests only** (`./internal/cli/... ./internal/mx/... ./internal/graph/...`); CI on the delivering PR is the full-suite judge. New tests use `t.TempDir()`, no OTEL env in parallel tests, path-free stderr assertions per cli conventions.
4. **Docs-site 4-locale parity** when cli-reference pages change; no template-source impact (workflow file dev-only — measured, research.md §5).

## §E. Self-Verification Plan

Run-phase records into progress.md §E.2 per criterion with command + verbatim output:

- E1 AC matrix — every acceptance.md criterion recorded PASS/FAIL with its verification command output.
- E2 Package gates — targeted `go test` (listed above), `go vet` on touched packages, `gofmt -l` clean on touched files.
- E3 Workflow guard dry-run — every reference-script branch of acceptance.md §A.1 (ancestor GREEN / non-ancestor RED / missing-object RED / no-anchor SKIP / push-event GREEN) plus the mutant-catch cell executed as local shell exactly as the canonical script reads, outputs quoted; AC-SP-011's static anatomy assertions executed against the landed YAML once M2 merges it (the YAML itself ships syntax-validated but unexecuted; first live execution is the delivering PR's own run — expectation stated there, never pre-claimed as observed).
- E4 Boundary grep — no absolute-path leakage in new CLI error strings; subagent boundary intact (no AskUserQuestion in internal/cli).
- E5 Docs parity — all four locale cli-reference files carry the recipe section or none does (same-change-set rule); `hugo -s docs-site --minify --gc` warning-free when touched.

## §F. Milestones

Ordering principle: decisions most likely to change review-first — type/interface surfaces lead, mechanical polish trails.

### M1 — Explicit stamp-commit mode (Priority High)

Data-model/interface decision surface: how the explicit commit threads through `mx.StampCodemaps`.

- Resolve-design point: single-caller signature extension vs options struct vs sibling constructor — pick minimal (deferred to run phase within the constraint set; REQ-SR-005/006/007 are the contract).
- Validation order inside RunE: resolve `<rev>` via `git rev-parse --verify <rev>^{commit}` FIRST; reject unresolved (exit non-zero, path-free stderr); reject `--commit`-with-dirty pre-write; then existing marshal/install/report path unchanged except CommitSHA sourcing.
- Flag definition follows `--root` precedent (`cmd.Flags().StringVar`); Long text gains the recipe line (REQ-SR-010 partially lands here; docs-site completion in M3).
- Unit tests (new): verbatim-sha recording; short-rev resolution; unresolvable rejection; dirty-combination rejection writing NO file (target absent after failed call); flagless characterization (HEAD/dirty anchoring unchanged vs pre-SPEC behavior snapshot).
- Delegation unit: one manager-develop turn, domain backend.

### M2 — CI pre-merge reachability guard (Priority High)

User-facing CI behavior decision surface: annotation format, skip-with-reason shape.

- **Conditioning mechanism — DECIDED (one mechanism only)**: a GitHub-level step-level `if:` splitting PR/push handling is PROHIBITED; the step runs UNCONDITIONALLY on both events and branches internally on base-ref emptiness (`[ -z "${GITHUB_BASE_REF:-}" ]` guarding ONLY the ancestry clause). Rationale: object-presence must be enforced on push events too, which a step-level `if: github.event_name == 'pull_request'` would silently skip (the exact silent-slip-out-of-enforcement shape plan-audit D1 names); one unconditional step keeps anatomy statically assertable (`continue-on-error` absent; emptiness test present at exactly one point). Plumbing pinned: the step declares `env: GITHUB_BASE_REF: ${{ github.base_ref }}` (empty string on push events).
- One new step in `.github/workflows/graph-freshness.yml`, positioned after `actions/checkout@v7`, before setup-go/build, carrying no `continue-on-error`. Canonical script text lives in acceptance.md §A.1; the shipped step body must match it modulo workflow-env lines. Branch contract: read tracked provenance → empty/no-anchor → print skip-with-reason, exit 0 → object missing (`git cat-file -e <sha>^{commit}`) → `::error` naming sha, exit 1 → empty base ref → "object presence verified only" line, exit 0 → ancestry vs `origin/${GITHUB_BASE_REF}` fails → `::error` naming sha + base ref, exit 1 → else pass line.
- Anatomy + push-event verification: AC-SP-011 (GREEN cell, mutant-catch cell, static assertions) and AC-SP-012 (judgeability-inputs inspection) are the matching criteria.
- Step comment carries the incident citation (`0d15864ae90b`, triage §F5) and the squash-merge rationale, matching the workflow's existing comment style.
- Local proof: execute the same shell block three times (fixture matrix of acceptance.md §AC-SP-001..003) and quote outputs; YAML syntax validated (`act` unavailable in this environment — the workflow's own first PR-run is the live witness, so the pull_request job's expected outcome for THIS PR must be predicted as GREEN given premise C2, and checked post-push).
- Delegation unit: folded into the M1 manager-develop delegation's tail OR one follow-up turn (both acceptable; the workflow edit touches no Go code — keep it serially after M1 to hold one-writer-on-file discipline trivial).

### M3 — Regression lock + operator docs + sync prep (Priority Medium)

Mechanical closure: least likely to change.

- Regression characterization: confirm-or-strengthen existing tests covering codemaps endpoint-diff metric, dirty-fingerprint anchoring, described_roots fidelity, threshold resolution, and the not-comparable → exit-2 path (REQ-SR-009). Add coverage ONLY where a gap measurably exists; cite gaps found (expected: none — predecessor SPEC-V3R6-GRAPH-FRESHNESS-002 already hardened these paths; verify, don't assume).
- Operator documentation: 4-locale cli-reference graph page recipe subsection + the incident reference; verify hugo build warning-free.
- Evidence assembly for progress.md §E.1→§E.2 handoff; confirm final stamp state meets REQ-GFR-014 inheritance (premise C2).
- Delegation unit: one manager-develop turn (tests/evidence) + manager-docs owns sync-phase doc polish if any drift appears.

## §G. Anti-Patterns

- Do NOT implement the guard in Go and consume the built binary in CI (couples detector correctness to the audited machinery — research.md §6 decision).
- Do NOT validate `--commit` against origin/main reachability at stamp time (undecidable locally; dishonest promise).
- Do NOT touch `tree_root`, schema_version, described_roots SSOT, or the exit-code constants.
- Do NOT let "clean up provenance while I'm here" leak into M-milestone diffs (scope discipline; committed tree_root oddity is documented, not broken).
- Do NOT run a local full suite to "be thorough" (load discipline; CI owns the suite).
- Do NOT claim the workflow guard works based on YAML inspection alone — quote the three executed branch outputs (verification-claim integrity; the live-CI execution remains predicted until observed).

## §H. Cross-References

- Problem canon: `.moai/reports/t279/triage-table.md` §F5
- Predecessors: SPEC-V3R6-GRAPH-FRESHNESS-001/-002 (audit trails frozen; related_specs only)
- Inherited obligation: predecessor REQ-GFR-014 (final-stamp main-reachability)
- Code anchors: internal/cli/graph_stamp.go:106 (--root precedent) · internal/mx/provenance.go:185-203 (anchor decision) · internal/graph/check.go:209-210 (not-comparable) · internal/cli/graph_check.go:19-25 (exit contract)
