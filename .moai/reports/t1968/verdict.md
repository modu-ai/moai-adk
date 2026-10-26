# t1968 lane verdict — SPEC-FETCH-WEB-001 (moai fetch web)

lane-7 · run tnbn7e · branch WT-jevfetch-a2sys-ai · verdict head ed6227740e · 2026-10-27

## Claim

Card t1968 (jevfetch port, applications 1-4) is COMPLETE and merge-ready: `moai fetch web <url> "<question>"` + MCP `fetch_web` with two-stage Jev scoring and verbatim excerpt output, internal/jev client cache/retry hardening, jevmeasure bench format + baseline compare, and template packaging (skill `moai-fetch-web`, permission `Bash(moai fetch web:*)`, glm-routing row, Codex note) — all 14 acceptance criteria PASS, independent sync audit PASS-WITH-DEBT 0.93, workflow files untouched.

## Evidence (commands + observed outputs, this run, this tree unless noted)

- Plan gate: plan-auditor 4 iterations converged 0.80→0.82→0.87→**0.92 PASS** (rcpt-e08ce4d1c1c9e3843a82299f @f57bbdd789); artifact-hash unchanged at kickoff (62583ecf57).
- Run: manager-develop (effort high, TDD) M1-M7 + evidence — 14/14 AC PASS with attribution triples (SPEC progress.md §E.2), coverage fetchweb 95.5%/jev 90.2%/jevmeasure 93.1%, `GOOS=windows` build exit 0, lint 0 new, spec.md draft→in-progress at M1 (Authored-By-Agent trailer).
- Sync: manager-docs single close commit be1e8b7e2c (in-progress→implemented→completed), B12 self-tests pass, CHANGELOG 14 live ACs, §E.4 backfilled 77027d5e79.
- Card review (advisory, codex): 3 rounds — R1-R8 (8 P2, verbatim-guarantee violations included), N1-N4, C1-C3; **ceiling honored** (no 4th review; C1-C3 fixed post-ceiling with RED-first tests; residual authority passed to the sync audit). 22/22 findings fixed RED-first across 0169d87655-series commits.
- Sync audit (independent, 5 iterations): 0.79 FAIL → 0.86 → 0.90 → 0.92 → **0.93 PASS-WITH-DEBT** (rcpt-2a184bfefac6e098feba3734 @ed6227740e, convergence pass). 24 findings total, all repaired RED-first — including SSRF defense in depth (dial-level IP validation closing DNS-rebinding TOCTOU, redirect policy, proxy disabled), answerless-scoring error contract, charset decision chain with substitution scan, verbatim-emission defects.
- Lane re-observation after every repair round matched the delegate reports (build/packages/graded tests re-run fresh, -race on concurrency fixes).

## Baseline-attribution

All measurements on this card worktree (`.moai/worktrees/t1968`), final evidence tree ed6227740e; per-AC commands and verbatim outputs committed in SPEC progress.md §E.2 (tree b81bf4d1a1 for the run set, re-observed at each repair head by lane batches). Graded tests re-observed by the lane at 11668653dd: `go build ./...` exit 0, fetchweb package `ok 0.318s`, `TestFetchWeb_OffGateZeroNetwork` PASS.

## Gaps

- Full `internal/cli` suite and repository-wide verdict are **not locally green-observed** (10m/25m timeouts under machine load from concurrent suites, zero test FAILs in every partial run) — owned by the integration-branch CI per lane-local discipline; PENDING at verdict time.
- RED-first is commit-claimed for M7 (markdown artifacts) via plan-phase ledger rows — no Go RED run exists for it by nature.
- RED-first for post-close repairs is repairer-claimed with verbatim pre-fix outputs in reports; the sync-auditor independently probe-reproduced the defect classes (its own probe tests), not every single RED run.

## Residual-risk

- D1 [Low][optional]: bare text after an aside-internal heading carries the restored pre-aside heading path (`[Main]`, not `[Main / Side]`) — test-pinned, commit-documented; follow-up card only if path fidelity is ever needed.
- Untested territory (auditor-noted): CDATA sections, frameset documents, proxy-env interactions.
- The card-review ceiling disposition (C1-C3 fixed without a 4th review) is accepted here as recorded; the sync audit's independent probes covered those fixes.
- Post-merge: leader merges PR; lane records merged-pr after observation; sweep after remote landing.
