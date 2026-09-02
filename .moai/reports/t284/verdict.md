# Card t284 — verdict

SPEC-AUDIT-PARTICIPANT-COUNT-001 — convergence participant count (audit_multi).

Branch `WT-audit-participant-count`, final HEAD at verdict time `1ee593747`.
Commit chain on the branch: `a67dcb472` (M1+M2 implementation) → `5713f82fa` (M3, 8 AC
tests) → `252853f8b` (run evidence) → `918d65366` (single sync commit, 3-phase close) →
`299c40f8c` (sync_commit_sha backfill) → `1ee593747` (sync-audit report). Zero pushes —
the integration window is the only public path (lane discipline 2026-09-01).

## Claim

The card's three items are delivered: (1) `audit_multi` results expose
`participant_count` (gate≠off ∧ verdict∈{pass,fail}, REQ-APC-002); (2)
`disagreement_flag=false` is forbidden below 2 participants — the flag is JSON `null`
there, except an observed intra-backend divergence keeps `true` (the D2 decision,
operator-confirmed option (a)); (3) the representative mutant (counts participants,
still emits `false` below 2) is killed by AC-APC-002/003/008, observed RED. The SPEC
closed 3-phase (`completed`, sync commit `918d65366`, sha backfilled in `299c40f8c`).

## Evidence

- Plan: iter1 FAIL 0.75 (D1-D6) → manager-spec repair → iter2 **PASS-WITH-DEBT 0.85**
  (Tier M 0.80; Clarity 0.75 / Completeness 1.0 / Testability 0.75 / Traceability 0.90;
  D1-D6 all resolved) → auditor-prescribed D7 discharge (scoping clauses, v0.2.1, no
  iter3 per auditor). Reports: `plan-audit-iter1.md`, `plan-audit-iter2.md`.
- Run: 8/8 AC PASS (`go test ./internal/cli/ -run 'AC_APC' -count=1` exit 0);
  three-block RED evidence (runtime, compile, mutant) with verbatim logs; builds
  native + `GOOS=windows` exit 0; `golangci-lint` internal/cli 0 issues (baseline
  identical); touched-function coverage 100% (converge, countParticipants,
  runMultiAudit, HandleMultiReviewGate), package 80.2% (pre-existing level).
  Logs: `e1-ac-matrix.log`, `red-stage1-runtime.log`, `red-stage2-compile.log`,
  `mutant-red.log`, `green-full-internal-cli.log`, `e3-cover.out` (all committed).
- Orchestrator trust-but-verify (independent re-execution, not consumed from §E):
  state (3 commits/clean/divergence), AC suite re-run, package suite fresh
  (`-count=1`, 268s, ok), both builds, lint 0, JSON-tag/`--stat` checks, carve-out
  body read (Step 2c semantics verified against the decided design), F-1
  `mcp_build_identity_test.go` repair spot-check — 8/8 PASS.
- Sync: single sync commit `918d65366` (15 files, docs only) — plan §G surfaces in
  4 locales + skill/rule surfaces (three-state semantics), CHANGELOG entry (B12
  three-pass self-test), template mirrors byte-identical + SPEC-ID-free (0 grep
  hits), `make build` + `go build` exit 0. Orchestrator re-verified: CHANGELOG count
  1, frontmatter `completed`, §E.4 placeholder→backfill, mirror parity, build.
- Independent audit: sync-audit **PASS 97.9≈98** (Functionality 98 / Security 100 /
  Craft 96 / Consistency 98), re-measured on `299c40f8c`: AC suite, five regression
  guards, vet, lint, both builds, coverage re-derived from `e3-cover.out`, evidence
  log authenticity (stage-1 RED carries no `participant_count` member — the
  pre-implementation signature), scope cleanliness. Report: `sync-audit.md`.

## Baseline-attribution

All figures measured in this lane's runs on branch `WT-audit-participant-count`:
RED blocks at `53a3fc1dd` + test file (unchanged engine); AC/E-batch at `5713f82fa`;
sync checks at `918d65366`; audit at `299c40f8c`; verdict at `1ee593747`. The audit
chain's artifact coordinates in SPEC §A were measured at `8c1d911df`/`64bba61aa`
(tree-pinned evidence; the tree absorbed `origin/develop` `53a3fc1dd` — card t248's
build-identity fields — after the audit; symbols located by name during run). The
`moai` MCP server on this host runs build `64bba61aa` — none of the verdict-bearing
measurements above ride that server; they are direct `go`/`grep`/`git` invocations.

## Gaps

- No CI verdict exists: the branch is unpushed by lane discipline; the full-suite
  verdict arrives with the integration-window push to `origin/develop` (the lead
  reads it).
- The mutant was not re-executed by the sync-auditor (read-only constraint);
  verified via assertion-shape analysis + log correspondence instead.
- Base-tree package coverage was not separately measured; the delta bound (≤ +0.12pp,
  ~20 statements all covered) covers the claim.
- docs-site hugo build not run (sync edits are prose-only inside existing blocks; no
  frontmatter/mermaid/structure change).
- sync-audit F1 [MINOR]: the CHANGELOG/§E.2 phrase "every function this SPEC added or
  changed … HandleMultiReviewGate" overstates the changed set (it is an exercised
  consumer; `multi_review_gate.go` has zero diff). Auditor ruled no pre-integration
  action needed; recorded here rather than rewritten post-audit.

## Residual-risk

- Non-Go JSON consumers reading `disagreement_flag` as a strict boolean: `null` is
  the designed undetermined signal; no consumer was found in-repo (audited both
  phases), but out-of-tree tooling is unobservable from here.
- `origin/develop` advanced ~45+ commits past the card's base during the run
  (parallel lanes); the merged-tree verdict is CI's at absorption. Absorption +
  merged-tree re-measurement (touched packages) happens inside the integration
  window before merge, per procedure.
- Pre-existing, recorded in spec.md §E: `docs-site/content/ko/advanced/autonomous-loops.md`
  carries a 4-locale parity gap (0 `disagreement` occurrences) that predates this
  card — deliberately untouched, separate card.
