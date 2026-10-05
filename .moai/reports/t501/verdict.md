# t501 Verdict — codex 미커버 표면 테스트 보강

Card closed through sync in worktree `.claude/worktrees/t501` (branch `WT-codex-uncovered`).
SPEC: `SPEC-CODEX-TEST-GAPS-001` (Tier M, REQ/AC 12/12, status: completed, sync commit `82674b94c`).
Window requested — WT branch unpushed, no CI directly requested (lead batch-pushes develop).

## Claim

1. The card's premise ("6 clusters of uncovered codex surfaces") was mechanically tested before scoping:
   TRUE on the name-citation axis (62/114 unique function names never cited in any `*_test.go`),
   REFUTED on the execution axis (4 of 118 coverprofile function rows at 0.0%; package 80.7%).
2. SPEC-CODEX-TEST-GAPS-001 delivers the genuine residue: 8 test items + a documented-skip record,
   tests only (zero non-test `.go` diffs), each test proven RED under a named mutant before GREEN.
3. After the work: zero 0.0% functions among the 118 baseline rows; all 12 ACs PASS; plan-audit
   iter-2 PASS 1.00; sync-audit PASS 97.9/100 (lens --deep, blocking 0).

## Evidence (commands + verbatim outputs; this lane's own runs unless attributed otherwise)

- Baseline (pre-work): `unset MOAI_KANBAN … && go test -count=1 -coverprofile=/tmp/t501_cover.out ./internal/cli/`
  → `ok github.com/modu-ai/moai-adk/internal/cli 515.612s coverage: 80.7% of statements` rc=0.
  118 rows across the card's 6 files; 4 at 0.0% (`.moai/reports/t501/coverage-perfunc.txt`).
- Post-work full scoped run (manager-develop, tree `1269e6d8d`, background, env-scrubbed,
  `-timeout 700s`): `ok github.com/modu-ai/moai-adk/internal/cli 337.374s coverage: 80.8% of statements` rc=0
  (`.moai/state/verify/t501/cover-after.log`).
- Lane re-measurements on final HEAD `426ab00df`:
  - Union gate (AC-CTG-009): `git diff --name-only 24df2ae45..HEAD` ∪ `git status --short`,
    filtered to non-test `.go` → EMPTY (both axes, measured twice: post-run and post-sync).
  - 0.0% census: `awk '$NF=="0.0%"' .moai/state/verify/t501/cover-after-6files.txt` → exactly 1 row:
    `codex_review_gate.go:183 runCodexReviewGate 0.0%` — outside the 118-row baseline denominator
    (the extract carries 127 rows = 118 + the out-of-scope 7th file's 9). Named 9 functions all
    66.7–100%; the tested `(realCodexConn).pid` is block-level 100% (3/3 blocks; sync-audit measured).
  - Direct test observation: `go test -count=1 -run 'TestTerminateCodexProcess|TestCodexIDMatches|
    TestAwaitCodexResponse|TestCodexCountExecutingImports|TestCodexGatePrintf|
    TestDefaultCodexInitGenerator|TestCodexSessionError|TestRealCodexConnPid' -v ./internal/cli/`
    → 9× `--- PASS` + `ok … 0.968s` (8 tests + 1 helper), observed on final tree.
  - Sync-close-last-write: `git diff --name-only 82674b94c..HEAD | grep '\.go$'` → empty.
- Mutant discipline: 8 mutants (manager-develop) + 1 independent mutant (sync-auditor:
  `realCodexConn.pid` positive branch → `return 0` → RED `mcp_codex_test.go:659` → fully reverted).
  All reverted; `git status --short` clean at every commit point.
- Audits: plan-audit iter-1 FAIL 0.875 (5 blocking instrument defects) → fix round `24df2ae45` →
  iter-2 PASS 1.00 (`.moai/reports/t501/plan-audit-iter2.md`); sync-audit PASS 97.9
  (`.moai/reports/t501/sync-audit.md`).
- Commits on `WT-codex-uncovered` (unpushed): `30bbb1753` plan artifacts · `24df2ae45` iter-1 fix ·
  `cd855f296` iter-2 report · `e652acf40` §F mode log · `ffa06117f` M1 · `1269e6d8d` M2-M7 ·
  `c79428b9c` run close · `82674b94c` sync close · `fb62d22e2` sync-audit report · `426ab00df` SHA backfill.

## Baseline-attribution

Every figure above was measured in this card's worktree on the named tree (plan baseline on
`ace1c5440`, run figures on `1269e6d8d`, lane re-measurements on `426ab00df`), in this run.
No value carried across trees or sessions without the label. Coverage percentages are per-tree
facts: the eventual develop merge tip will re-measure differently as sibling lanes merge.

## Gaps (explicitly NOT observed)

- Remote CI verdict: the lane does not push; the full-suite judgment comes from origin/develop CI
  after the lead's batch push. Local scoped runs (515.6s baseline / 337.4s post) are early signals.
- Deliberate skips (spec.md §D): `writeCodexRequest`/`writeCodexEnvelope` marshal-error arms
  (dead-in-practice: envelopes built from JSON-safe literals); `terminateCodexProcess`
  FindProcess-error arm (platform-unreachable for integer pids on darwin/linux). No tests were
  written for these; their coverage percentage stays below 100 by design.
- Out-of-scope surfaces not touched: everything outside the card's 6 files — including
  `codex_review_gate.go` and `(codexSessionHandle).pid`.

## Residual-risk

- `runCodexReviewGate` (codex_review_gate.go:183) at 0.0% — a 7th codex file the card's scan never
  covered (predates this card, #1430). Follow-up card candidate.
- `(codexSessionHandle).pid` nil-guard arm at 66.7% — hermetically coverable per sync-audit F1;
  follow-up card candidate.
- Package coverage 80.8% < 85% institutional target — inherited baseline, 76 genuine 0.0% functions
  remain across the package; out of this card's scope.
- The post-work full run (337.4s) was faster than baseline (515.6s) — machine-load variance, not a
  code property; do not read speed into the change.

## Card traceability

- Dispatch `card: t501` · branch `WT-codex-uncovered` · card id in every commit message ·
  evidence path `.moai/reports/t501/` (this verdict + 5 supporting files).
- Kickoff: operator approval 2026-09-07 via lead session (lead measured HEAD `cd855f296` pre-approval).
