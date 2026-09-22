# t519 Verdict — codex coverage residual (runCodexReviewGate + codexSessionHandle.pid)

Card closed through sync in worktree `.claude/worktrees/t519` (branch `WT-codex-cover-residual`).
SPEC: `SPEC-CODEX-COVER-RESIDUAL-001` (Tier M — corrected from the dispatch's S at authoring time:
REQ 10 / AC 12 exceed the Tier S 8/8 ceilings; status: completed, sync commit `650cc7d2f`).
Window requested — WT branch unpushed, no upstream, no CI requested (lead batch-pushes develop).

## Claim

1. The card's premise HELD on the execution axis (unlike t501, where it was refuted): on the
   re-measured baseline `runCodexReviewGate` was 0.0% and `(codexSessionHandle).pid` 66.7%.
   Root cause of the 0.0%: `multi_review_gate_wiring_test.go:13` claims a one-for-one mirror of
   codex-review-gate wiring tests that were never written (grep control: 0 vs 3 rows).
2. SPEC-CODEX-COVER-RESIDUAL-001 delivers tests only: one new file
   (`internal/cli/codex_review_gate_wiring_test.go`, 5 tests) + one appended test
   (`internal/cli/mcp_codex_test.go`, `TestCodexSessionHandlePid`). Zero non-test `.go` diffs.
3. After the work: `runCodexReviewGate` 92.3% (= the 12-of-13-statement ceiling; the single
   uncovered statement is the documented skip S1 `if out == nil`, proven unreachable by code
   reading of all seven `HandleCodexReviewGate` returns), `pid` 100.0%. All 12 ACs PASS;
   plan-audit iter-2 PASS 0.9375; sync-audit PASS 96.8 (blocking 0).

## Evidence (commands + verbatim outputs; lane's own runs unless attributed)

- Baseline (lane, tree `bf779ecf2` = local develop tip at entry, t501 `d4162b368` included):
  `unset MOAI_KANBAN … && go test -count=1 -timeout 700s -coverprofile=… ./internal/cli/`
  → `ok  	github.com/modu-ai/moai-adk/internal/cli	473.665s	coverage: 80.9% of statements` rc=0.
  `go tool cover -func` rows: `codex_review_gate.go:183 runCodexReviewGate 0.0%`,
  `mcp_codex.go:701 pid 66.7%` (`.moai/reports/t519/coverage-baseline-targets.txt`).
- Post-work (manager-develop, tree `a26feb41e`; Go content identical through `d1b66d06e` —
  `git diff --name-only a26feb41e..d1b66d06e | grep '\.go$'` empty, re-verified by sync-auditor):
  `ok  	github.com/modu-ai/moai-adk/internal/cli	448.176s	coverage: 81.0% of statements` rc=0;
  rows `runCodexReviewGate 92.3%`, `pid 100.0%` (`.moai/reports/t519/coverage-after-targets.txt`).
- Lane direct observation on `d1b66d06e` (env-scrubbed):
  `go test -count=1 -run 'TestRunCodexReviewGate|TestCodexSessionHandlePid' -v ./internal/cli/`
  → 6× `--- PASS: Test…` + `ok … 2.334s` (`.moai/state/verify/t519/lane-sweep.txt`).
- Union gate on final HEAD `06ad7e54e`: `git diff --name-only bf779ecf2..HEAD | grep '\.go$' |
  grep -v '_test\.go$'` → empty (rc=1); `git status --short` → empty. The two `.go` paths in the
  diff are both `_test.go`.
- AC-CCR-012 (sync close is the last write): `git diff --name-only 650cc7d2f..HEAD | grep '\.go$'`
  → empty (rc=1); the only commit after the close is the docs-only SHA backfill `06ad7e54e`.
- Mutant discipline (`.moai/reports/t519/mutants.md`): 8 applied RED→GREEN (M1, M1b, M2, M3a,
  M3b, M4, M5a panic-trace, M5b) + M6 absence-guard adoption; M3-vac confirmed vacuous by
  measurement; M2-cf counterfactual proves `withChangeDetector(t, true)` is load-bearing;
  sync-auditor's independent M7 (drop the BLOCK reason at codex_review_gate.go:201) →
  `codex_review_gate_wiring_test.go:180: BLOCK must carry a non-empty reason … --- FAIL:`; all
  reverted, production blobs re-hashed identical.
- Audits: plan-audit iter-1 FAIL 0.775 (blocking 4: host-dependent fixture, vacuous mutant,
  matrix/ledger mismatch, seam-file mis-attribution) → fix `1e21635f1` → iter-2 PASS 0.9375
  (2 minor one-word fixes `c6bf21a72`); sync-audit PASS 96.8 (`.moai/reports/t519/sync-audit.md`).
- Quality (manager-develop, final run tree; sync-auditor re-ran): `go vet` rc=0,
  `golangci-lint run ./internal/cli/` → `0 issues.`, `gofmt -l internal/cli/` empty,
  `GOOS=windows GOARCH=amd64 go build ./...` rc=0.
- Commits on `WT-codex-cover-residual` (11 since `bf779ecf2`, unpushed, no upstream):
  `faa76fa7e` plan artifacts · `1e21635f1` iter-1 fixes · `c6bf21a72` iter-2 · `21a52d507` §F mode
  log · `4ed5011c4` M1 · `a26feb41e` M2 · `22915e790` M3 close · `d1b66d06e` docs correction ·
  `b9945bf19` sync-audit report · `650cc7d2f` sync close · `06ad7e54e` SHA backfill.

## Baseline-attribution

Every figure above names the tree it was measured on (`bf779ecf2` baseline; `a26feb41e` post-work
coverage, attributable to `d1b66d06e`/`06ad7e54e` because no `.go` changed after it; lane
re-measurements on `d1b66d06e` and `06ad7e54e`). Coverage is a per-tree fact; the merged develop
tip will be re-measured in the window before the `--no-ff` merge.

## Gaps (explicitly NOT observed)

- Remote CI verdict: the lane does not push; the full-suite judgment comes from develop CI after the
  lead's batch push. Local runs are scoped to `./internal/cli/` (no `./...`, no `-race` on the
  package; sync-auditor ran `-race` on the six-test selector only: `ok … 2.325s`).
- Package coverage stayed 80.9% → 81.0%, below the 85% institutional target (inherited; not an AC).
- Documented skip S1 (`if out == nil`, codex_review_gate.go:198) untested by design.
- Merged-tree re-measurement not yet done (window not yet granted; local develop was `edf782e7e`
  at last read, 42 commits past this branch's base).

## Residual-risk

- S1's unreachability rests on today's `HandleCodexReviewGate` returns; a future seam returning
  (nil, nil) would silently make the arm live with no mechanical guard (sync-audit advisory F2).
- Adjacent under-covered functions in the same file — `reviewableFromPorcelain` 83.3%,
  `readHookInput` 85.7% — untouched by decision; follow-up card candidates.
- Commit `22915e790` body carries the typo `HandleCadexReviewGate` (`--amend` forbidden); code and
  docs spell it correctly.
- E4 boundary grep returns 32 rows in `internal/cli` — all inherited (help prose, lint fixtures,
  the detector itself); non-test diff 0 so none is attributable to this card.
- `moai spec lint <directory>` fails with `ParseFailure: is a directory` (argument form; against
  `spec.md` it reports no findings) — `/moai feedback` candidate, also seen on t500.

## Card traceability

- Dispatch `card: t519` · branch `WT-codex-cover-residual` · card id in every commit message ·
  evidence path `.moai/reports/t519/` (this verdict + coverage-baseline-*, coverage-after-*,
  mutants.md, plan-audit-iter{1,2}.md, sync-audit.md).
- Kickoff: operator approval 2026-09-07 via lead session (autonomous progression; `ac_converge`
  armed at HEAD `21a52d507`).
