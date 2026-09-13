# SPEC Review Report: SPEC-GATEWAY-WEDGE-REROOT-001 (card t700) — Delta Re-audit

Iteration: **2** (Tier M ceiling = 2 per `harness.plan_audit_tier_ceilings` S=1/M=2/L=3; the dispatch's "2/3" label is the legacy Tier L number. A PASS here needs no further iteration.)
Verdict: **PASS**
Overall Score: **1.00** (Tier M PASS threshold 0.80 — cleared with margin; iter-1 was 0.6875 COND-FAIL, no score regression)

Scope: delta re-audit over iter-1 defects D1–D7 only, plus the absorbed-tree pin re-verification the lead directed, the mechanical lint re-run, and the operator-directed t707 consistency check. Reasoning context from the SPEC author was not supplied; audit performed on the artifacts only (M1 Context Isolation).

Baseline-attribution: worktree `.claude/worktrees/t700`, branch `WT-wedge-reroot-policy`, tip `a86ff2e3c` (= merge of local develop `d416f8162`; SPEC artifacts + iter-1 report committed at `631704655`, spec.md v0.1.1 carries the fix HISTORY row). This run, this tree.

## Regression Check (Iteration 2) — D1–D7 disposition

- **D1 [RESOLVED]**: AC-WRR-014 maps REQ-WRR-008 (BLOCKER, baseline-relative `.Fork(` call-site grep, baseline recorded in progress.md §E.2); AC-WRR-015 maps REQ-WRR-009 (MINOR, named path `.moai/docs/gateway-wedge-recovery.md`, per-element greps; DoD 1 updated in step). AC-WRR-008's requirement cell now uses full ID tokens ("REQ-WRR-003, REQ-WRR-006"). Mechanical: `go run ./cmd/moai spec lint --strict` → **"No findings"** (iter-1: 3 × `CoverageIncomplete`). Verified live.
- **D2 [RESOLVED]**: AC-WRR-012 (acceptance.md:34, §C:53) is now file-level zero-diff over FIVE files — `core.go`, `receipt_history.go`, `request.go`, `family.go`, `gateway_factory.go` — plus zero paths under `internal/gateway/receipt/`, merge-base re-derived at measurement time (gitflow-lane-protocol §8 form), with `TestReceiptHistoryRejectsMidPairTruncationBeforeHistory` named as the wire-level behavioral net. plan.md §D PRESERVE adds `request.go` (entire file, the production authorization call site). Residual editorial lag recorded as N2 below (optional, no gate impact — M4 explicitly delegates to AC-WRR-012, which §E E1 declares the SSOT).
- **D3 [RESOLVED]**: REQ-WRR-007 (spec.md:81) reworded to the §3.2 prefix-property form with the client-side indistinguishability of trailing forged/stripped boundaries stated explicitly and the safeguards named (REQ-WRR-005 + REQ-WRR-003-3); §3.4 gains the trailing-forged/stripped-tail row (spec.md:118); new AC-WRR-016 (BLOCKER: removed content never accepted by any replay, remainder accepted) + §D edge case 4. The iter-1 contradiction is gone; the requirement set is now internally consistent with REQ-003/004/005.
- **D4 [RESOLVED]**: REQ-WRR-003-2 (spec.md:71) requires a durable attempt marker (attempted flag + preimage digest) on the launcher conversation record, fresh-process bound stated; REQ-WRR-006 reads the marker (spec.md:79); AC-WRR-008 covers BOTH same-process and fresh-process re-invocation (acceptance.md:30/:49); aside-before-replace explicit in REQ-WRR-003-3 (spec.md:72). Concurrent-invocation locking remains correctly out (optional hardening per iter-1 scoping).
- **D5 [RESOLVED]**: plan.md §A pins local `develop` as the ONLY sanctioned absorb target (§11 citation) with the iter-1 SHAs marked "attribution, never pins"; §A 1a adds the ancestry gate (`git merge-base --is-ancestor <t697-merge-SHA> HEAD`, blocker on non-zero) — verified live today: `f45c2dddf` IS an ancestor of HEAD `a86ff2e3c`. §C replaces the broken `go doc receiptHistory` form (unexported — iter-1 measured exit 1) with declaration greps; every grep pattern in §C resolves on the absorbed tree (verified this run), and the compile-net statement covers the remaining pins. AC-WRR-013 downgraded MINOR — now consistent with the severity legend and M5's Priority Low; DoD 1 updated.
- **D6 [RESOLVED]**: AC-WRR-006 (acceptance.md:28) names `TestReceiptHistoryRejectsForeignItemReplay` and its `CauseLineage` section, and says no standalone lineage test exists. The cited lines :187–190 remain **accurate on the absorbed tree** (measured: `CauseLineage` at receipt_history_cause_test.go:187–188; the file was untouched by the absorb — all six cited tests sit at their iter-1 line numbers).
- **D7 [RESOLVED]**: AC-WRR-010 (acceptance.md:32/:51) adds the structural no-store-handle assertion (no store argument, no `OpenStore` reference in the recovery path's files) with the digest demoted to complement; §C states the rationale (digest cannot see write-then-restore; the structural assertion covers it).

**Unresolved count: 0.** Per the Retry Loop Contract, no unresolved prior-iteration defect remains.

## Must-Pass Re-check (touched surfaces)

- MP-1: REQ-WRR-001..009 unchanged in numbering — PASS.
- MP-2 (requirement layer): revised REQ-003-2/3, 006, 007 retain their GEARS shapes (Where / When / Ubiquitous-with-shall); the added sentences are explanatory, normative cores intact — PASS.
- MP-3: v0.1.1 quoted semver, dates ISO, all 12 canonical fields present; `spec_audit` (project_root = this worktree) → V3R6, `modern_era_clean: 1`, 0 drift findings — PASS.
- MP-4: N/A (single-language Go SPEC).
- MP-5/D7 re-check after the ~200-commit absorb: `SPEC-MOAI-GATEWAY-001` status advanced `draft` → **`implemented`** — still not in {retired, superseded, archived}; no reconciliation obligation — PASS.
- MP-6/D8: no `syscall` introduced by the fix pass — PASS (auto).
- MP-7: no `[NEEDS CLARIFICATION]` markers in the revised artifacts — PASS.

## Absorbed-Tree Pin Re-verification (lead-directed)

**All 16 symbol pins resolve on `a86ff2e3c`.** receipt_history.go: `historyReplayGuidance`:40 (string byte-unchanged), `HistoryReplayError`:45, `NewReceiptHistory`:86, `NewGPTSubscriptionReceiptHistory`:93, `receiptHistory.Check`:197, `replayCause`:223, `Publish`:239, `checkObserved`:262. core.go: `Candidate`:40, `Observation`:49, `Manifest.Check`:99, `Manifest.Fork`:130. family.go: `Manager.Fork`:259. gateway_factory.go: `authorizeGatewayNativeReceipt`:59, `newGatewayHandlerFactory`:80. Launcher: `families.Fork(` gateway_session.go:204. Line numbers SHIFTED versus iter-1 (e.g., Check 172→197) — the symbols-never-lines pinning worked exactly as designed.

**t697 overlap re-measured on the landed merge**: `git diff 44e56d017 f45c2dddf --name-only -- internal/gateway/ internal/cli/` = 6 files (`anthropic.go`, `openai.go`, `upstream.go` + 3 test files) — **zero overlap** with the five preserved files.

**Cited tests**: all 6 (5 cited at iter-1 + the behavioral net `TestReceiptHistoryRejectsMidPairTruncationBeforeHistory`) exist in `receipt_history_cause_test.go` at unchanged lines. **Characterization suite GREEN on the absorbed tree**: `go test ./internal/gateway/translate/ -run TestReceiptHistory -count=1` → `ok ... 0.809s` (the package now also compiles t707's `replay_roundtrip_test.go` — the compile net for the unexported pins is live).

**AC-WRR-014 live baseline**: `grep -rn "\.Fork(" internal/ --include="*.go" | grep -v _test` → exactly **1** production call site (gateway_session.go:204). Recorded for the run-phase §E.2 baseline.

## Category Scores (0.0–1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.0 | 1.0 | REQ-007 rewrite resolves the iter-1 two-readings defect; the client-undetectability of the trigger condition is now a stated, safeguarded property rather than an ambiguity |
| Completeness | 1.0 | 1.0 | §3.4 disposition row added; all sections present; AC matrix at 16/16 (Tier M ceiling, within budget) |
| Testability | 1.0 | 1.0 | All 16 ACs binary with executable verbs; AC-012 file-level + named command; AC-014 baseline-relative grep (live-measured 1); AC-010 structural + complement; AC-013 legend-consistent. Residual keyword-language nit on AC-015 recorded as N3 (outcomes converge for any executor; not an interpretation fork) |
| Traceability | 1.0 | 1.0 | All 9 REQs covered; full ID tokens; mechanical lint = 0 findings (verbatim: "✓ No findings — all SPEC documents are valid") |

## Defects Found (new, informing — none expands this iteration's fix scope)

N1. AC-WRR-014's parenthetical enumeration (acceptance.md:36/:55) says the expected set is "the Manager.Fork definition + the launcher caller", but the grep pattern `\.Fork(` cannot match the definition (`func (m *Manager) Fork(` has no dot before Fork) — live measurement: 1 hit, the caller only. Severity: minor — Class: optional — Required fix: correct the parenthetical ("the launcher caller; the definition itself does not match the dot-prefixed pattern") or widen the pattern. No false gate: the baseline-relative check governs.
N2. plan.md §F M4 prose (:99) still lists 3 of the 5 preserved files, lagging AC-WRR-012's five-file set. Severity: minor — Class: optional — Required fix: sync the M4 sentence (M4 already defers to AC-WRR-012, so no gate impact).
N3. AC-WRR-015's per-element greps do not specify keyword language while the deliverable doc is ko (`documentation: ko`). Severity: minor — Class: optional — Required fix: name bilingual keywords in M6.
N4. Cosmetic: plan.md §C commands do not explicitly cover the exported pins `Candidate`/`Observation` (compile net covers them); AC-WRR-001's row retains the combined token "REQ-WRR-001/007" (lint-clean because both REQs have standalone references elsewhere). Severity: cosmetic — Class: optional.

## t707 Consistency Check (operator-directed)

**No contradiction.** t707 cleared the receipt serialization seam (publish↔check byte-match maintained across all 8 reproducible shapes; test-only change, no production code) — this leaves the SPEC's premise (validator sound; the wedge is a client-side suffix divergence) untouched. t707's live CauseChain 400s are a DIFFERENT shape — transcript-vs-request byte divergence immediately after an Edit tool, divergence from boundary[0] — and the SPEC explicitly does not claim that shape recoverable: REQ-WRR-007 confines recovery to the trailing-unpublished-turn wedge and rejects all other shapes; REQ-WRR-005's misdiagnosis guard ("the chain class does not uniquely identify the wedge shape; a human confirms") is empirically STRENGTHENED by t707, which gives the chain class a second known underlying shape.

Notes (adjacency, not findings): (a) the M6 operator doc (AC-015) would be enriched by one sentence noting that a chain-classified 400 can also be the t707 Edit-adjacent shape and that a single-shot recovery will not fix it (bounded termination per REQ-WRR-006) — content enrichment, orchestrator discretion; (b) t707's `replay_roundtrip_test.go` now sits in `internal/gateway/translate/` as part of the suite compile net — test-only, no preserve-set conflict; (c) t708 (envelope repair) and t838 (live instrumentation) share the gateway surface but overlap no SPEC claim, and t707's `observability.patch` is uncommitted, so no tree interaction with this card's zero-diff lock.

## Recommendation

Proceed. The SPEC is plan-phase complete at iteration 2 of the Tier M ceiling: all seven iter-1 defects resolved with mechanical evidence, all 16 pins verified on the absorbed tree, lint clean, suite green, and no contradiction with the t707 verdict. Optional polish (N1–N4) may ride any future artifact touch at the orchestrator's discretion — none gates run entry.

Downstream obligations unchanged: Implementation Kickoff Approval remains mandatory before run-phase entry; at run entry, M1 re-measures the absorb (local `develop` is the only sanctioned target) and re-runs the §A 1a ancestry gate plus the §C pin re-verification against the then-current tree — the values recorded in this report are attribution for iteration 2, never pins.

Cross-model attribution for this iteration: single-backend claude verdict (the iter-1 convergence already folded codex's adversarial findings into D1–D7; the delta scope contained no new security-surface question warranting a re-fan-out). Mechanical evidence this run: spec lint 0 findings; spec_audit 0 drift; characterization suite `ok 0.809s`; ancestry gate exit 0; 16/16 pins; 6/6 cited tests; t697 overlap 0; AC-WRR-014 baseline 1.
