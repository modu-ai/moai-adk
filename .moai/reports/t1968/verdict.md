# t1968 lane verdict — SPEC-FETCH-WEB-001 (moai fetch web)

lane-7 · run tnbn7e · branch WT-jevfetch-a2sys-ai · verdict head 8ce65a38cd · 2026-10-28

## Claim

Card t1968 (jevfetch port, applications 1-4) is COMPLETE and merge-ready: `moai fetch web <url> "<question>"` + MCP `fetch_web` with two-stage Jev scoring and verbatim excerpt output, internal/jev client cache/retry hardening, jevmeasure bench format + baseline compare, and template packaging (skill `moai-fetch-web`, permission `Bash(moai fetch web:*)`, glm-routing row, Codex note) — all 14 acceptance criteria PASS, independent sync audit PASS (debt-free) 0.95 at iter14 (rcpt-8babf16e21e731df3fe546dd @8ce65a38cd, convergence check ok), workflow files untouched.

## Evidence (commands + observed outputs, this run, this tree unless noted)

- Plan gate: plan-auditor 4 iterations converged 0.80→0.82→0.87→**0.92 PASS** (rcpt-e08ce4d1c1c9e3843a82299f @f57bbdd789); artifact-hash unchanged at kickoff (62583ecf57).
- Run: manager-develop (effort high, TDD) M1-M7 + evidence — 14/14 AC PASS with attribution triples (SPEC progress.md §E.2), coverage fetchweb 95.5%/jev 90.2%/jevmeasure 93.1%, `GOOS=windows` build exit 0, lint 0 new, spec.md draft→in-progress at M1 (Authored-By-Agent trailer).
- Sync: manager-docs single close commit be1e8b7e2c (in-progress→implemented→completed), B12 self-tests pass, CHANGELOG 14 live ACs, §E.4 backfilled 77027d5e79.
- Card review (advisory, codex): 3 rounds — R1-R8 (8 P2, verbatim-guarantee violations included), N1-N4, C1-C3; **ceiling honored** (no 4th review; C1-C3 fixed post-ceiling with RED-first tests; residual authority passed to the sync audit). 22/22 findings fixed RED-first across 0169d87655-series commits.
- Sync audit (independent, 14 iterations): 0.79 FAIL → 0.86 → 0.90 → 0.92 → 0.84 → 0.92 → 0.91 → 0.95 → 0.91 (F1) → 0.95 → 0.95 → 0.95 → **0.95 PASS debt-free at iter14** (rcpt-8babf16e21e731df3fe546dd @8ce65a38cd; `moai verify audit-plan` convergence ok:true). 40+ findings total across the chain, all repaired RED-first; iter14 re-verified the codex round-3 P2 repairs bidirectionally (HEAD overlay 8/8 PASS, pre-repair overlay 7 FAIL + 2 controls).
- Lane re-observation after every repair round matched the delegate reports (build/packages/graded tests re-run fresh, -race on concurrency fixes).

## Baseline-attribution

All measurements on this card worktree (`.moai/worktrees/t1968`), final evidence tree 8ce65a38cd; per-AC commands and verbatim outputs committed in SPEC progress.md §E.2 (tree b81bf4d1a1 for the run set, re-observed at each repair head by lane batches). iter14 audit re-observed fresh at 8ce65a38cd: trio `ok` (cover 95.6/93.7/91.1), cli filtered 8/0, corpus/table 76 subtests 0 FAIL, lint `0 issues.`, gofmt/vet clean, `BUILD_OK`+`WIN_BUILD_OK`, spec-lint clean. The two codex round-3 repair commits landed and lane-verified at d4457e641b (build/vet/gofmt/lint/windows all exit 0, new tests PASS, slot go-test held).

## Gaps

- Full `internal/cli` suite and repository-wide verdict are **not locally green-observed** (10m/25m timeouts under machine load from concurrent suites, zero test FAILs in every partial run) — owned by the integration-branch CI per lane-local discipline; PENDING at verdict time.
- RED-first is commit-claimed for M7 (markdown artifacts) via plan-phase ledger rows — no Go RED run exists for it by nature.
- RED-first for post-close repairs is repairer-claimed with verbatim pre-fix outputs in reports; the sync-auditor independently probe-reproduced the defect classes (its own probe tests), not every single RED run.

## Residual-risk

- D1 [Low][optional]: bare text after an aside-internal heading carries the restored pre-aside heading path (`[Main]`, not `[Main / Side]`) — test-pinned, commit-documented; follow-up card only if path fidelity is ever needed.
- Untested territory (auditor-noted): CDATA sections, frameset documents, proxy-env interactions.
- The card-review ceiling disposition (C1-C3 fixed without a 4th review) is accepted here as recorded; the sync audit's independent probes covered those fixes.
- Post-merge: leader merges PR; lane records merged-pr after observation; sweep after remote landing.
