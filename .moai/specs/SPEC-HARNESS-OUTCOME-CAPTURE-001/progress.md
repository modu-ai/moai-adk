---
id: SPEC-HARNESS-OUTCOME-CAPTURE-001
title: "Progress — Harness Apply outcome capture"
version: "0.1.0"
status: completed
created: 2026-06-14
updated: 2026-06-14
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness"
lifecycle: spec-anchored
tier: M
tags: "harness, outcome-capture, progress"
era: V3R6
---

## §F.1 Plan-phase

### Tier + artifact set

- **Tier**: M (standard) — 3-artifact set (spec.md + plan.md + acceptance.md) + this progress.md.
- **Rationale**: brownfield extension of the observer/Event (~250-400 LOC, ~4 files: types.go additive fields + new outcome.go + applier.go composition seam + tests). The Apply-seam composition (DD-2) and the representation choice (DD-1) are genuine non-mechanical decisions in a recently-landed tested pipeline, warranting plan.md design documentation. NOT Tier S (under-documents the seam decision). NOT Tier L (no new package, no FROZEN change, no safety layer, one new exported const).

### Roadmap position

- Self-Harness P2/Phase5 FIRST sub-item: the observer OUTCOME-capture enabler.
- Predecessors completed: P0 (LOOP-CLOSURE-001, M6 lineage + M7 human-gate), P1 (REGRESSION-GATE-001, M2-lite gate — supplies the verdict + delta reused here).
- Downstream consumers (D1 clustering, canary-effectiveness, held-out, scorer-loop) are OUT OF SCOPE — each a separate future SPEC.

### Ground-truth corrections applied (authored from verified file reads, not premise)

- `internal/harness/observer.go` EXISTS — `RecordEvent` + `RecordExtendedEvent` append `Event` structs to `usage-log.jsonl`. The new `RecordOutcome` reuses `RecordExtendedEvent` (DD-6).
- `internal/harness/regression_gate.go` (P1) ALREADY landed: `MetricTriple`, `Regressions`, `ApplyRegressionError`, `BaselineStore`, `goMeasurer`, and `applyWithRegressionGate` are in place. The outcome record REUSES these (no re-creation).
- The Apply pipeline ALREADY emits lineage with `Decision ∈ {"approved","rejected","regression-blocked"}`. The outcome capture is a SEPARATE observer write (the lineage manifest is the human-audit record; the observer log is the machine-analysis substrate — B.2).
- **No `lineage_id` field exists** — `LineageEntry` is keyed by `ProposalID` + `Timestamp`. The outcome reuses `proposal_id` as the correlation key; no `lineage_id` is introduced (DD-3 / C11).
- The GAP this SPEC closes: the `Event` struct has NO OUTCOME fields, and there is NO `RecordOutcome` path surfacing an Apply's verdict+delta+proposal_id into `usage-log.jsonl`. Confirmed by grep (no `RecordOutcome`/`Outcome` capture path in `internal/harness/*.go`).

### Design decisions (folded in plan.md §F)

- **DD-1** Representation: additive `omitempty` fields on `Event` + new `EventTypeApplyOutcome` const (follows the established Stop/SubagentStop/UserPromptSubmit omitempty pattern). Alternatives (b) separate log file / (c) extend `LineageEntry` rejected (substrate fragmentation / completed-schema churn).
- **DD-2** Wiring point: compose + emit inside `applyWithRegressionGate`, AFTER the decision, in BOTH kept + rolled-back branches; nil-observer no-op; emit-error wrapping mirrors `writeLineage`.
- **DD-3** Correlation key: reuse `proposal_id`; do NOT add `lineage_id` (ground-truth: none exists; C11 forbids lineage-schema churn).
- **DD-4** Gate-inactive path emits no outcome (no measured triple) — pre-P1 behavior untouched.
- **DD-5** Honest framing (MUST-PASS): capture+persist ONLY; typically Δ=0; no "improves"/"prevents".
- **DD-6** `RecordOutcome` reuses `RecordExtendedEvent` (no new low-level writer).

### Self-assessment

| Metric | Value |
|--------|-------|
| Tier | M |
| REQ count | 13 (REQ-OC-001 … REQ-OC-013) |
| HARD constraints | 12 (C1 … C12) |
| AC count | 12 (AC-OC-001 … AC-OC-012) |
| Files affected (run-phase est.) | ~4: `types.go` (additive), `outcome.go` (new), `applier.go` (seam), `outcome_test.go` (new) + `applier_test.go` (additive tests) |
| LOC est. (run-phase) | ~250-400 |
| Representation choice | DD-1 option (a): additive `omitempty` Event fields + `EventTypeApplyOutcome` |
| Wiring choice | DD-2: in-`applyWithRegressionGate` post-decision emit, both branches |

### Open questions (for plan-auditor / GATE-2)

- **OQ1 (observer injection signature)**: thread the observer through `NewApplierWithRegressionGate`'s existing signature vs add a separate `WithOutcomeObserver(*Observer)` option. Run-phase detail; SPEC requires only that the gate-active production seam can emit and that a nil observer is a safe no-op. No SPEC-level blocker.
- **OQ2 (omitempty zero-drop)**: `int`/`float64` omitempty drops genuine zero triples on a `kept` Δ=0 outcome. plan.md DD-1 argues the `outcome_verdict` field (always non-empty) is the load-bearing signal, so the dropped zeros are acceptable. If a future consumer needs the explicit zeros, it can drop omitempty on those fields then (additive, backward-compatible). Surfaced for plan-auditor judgment — not a blocker.

plan_complete_at: 2026-06-14T00:00:00Z
plan_status: audit-ready

---

## §E.1 — Phase 0.95 Mode Selection

(orchestrator-autonomous; recorded for the run-phase `Agent()` spawn per `.claude/rules/moai/workflow/orchestration-mode-selection.md` §D. GATE-2 cleared; Phase 0.5 plan-auditor iter-2 PASS 0.92.)

- **Input parameters**: tier M · scope ~4 files · domain count 1 (`internal/harness` Go only) · file mix 100% Go · concurrency benefit LOW (coding-heavy single-domain) · Agent Teams prereqs N/A.
- **Mode evaluation**: trivial=no (multi-file semantic) · background=no (writes files) · agent-team=no (<3 domains, coding-heavy) · parallel=no (coding-task parallelism caveat) · **sub-agent=YES** · workflow=no (not ≥30 files / not a single uniform mechanical transform).
- **Decision**: sub-agent (Mode 5).
- **Justification**: coding-heavy Go run-phase, single domain, Tier M ~4 files. Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"), sequential sub-agent is the safe default — one `manager-develop` (cycle_type=tdd) executed M1→M4 RED-GREEN-REFACTOR. Mode 6 rejected (not high-volume mechanical). GATE-2 cleared before this decision.

## §E.2 Run-phase Evidence

Run-phase implemented by manager-develop (cycle_type=tdd, RED-GREEN-REFACTOR). 5 modified/new
files: `internal/harness/types.go` (additive: EventTypeApplyOutcome const + 10 omitempty Outcome*
fields + LogSchemaVersion "v1"→"v2"), `internal/harness/outcome.go` (NEW: OutcomeRecord +
RecordOutcome), `internal/harness/applier.go` (composition seam: outcomeObserver field +
WithOutcomeObserver setter + recordOutcome nil-safe helper + emit at both terminal branches of
applyWithRegressionGate), `internal/harness/outcome_test.go` (NEW: 12 outcome tests),
`internal/harness/observer_test.go` (additive: TestLogSchemaVersion v1→v2 assertion update — the
SSOT schema-version test, an in-scope consequence of the REQ-OC-010 bump).

### AC PASS/FAIL matrix (acceptance.md SSOT — 12 ACs)

| AC | REQ | Status | Verification command | Actual output |
|----|-----|--------|----------------------|---------------|
| AC-OC-001 | REQ-OC-001/004, C8/C9 | PASS | `go test -run 'TestApplyOutcomeEvent_RoundTrip$\|TestApplyOutcomeEvent_OmitemptyOnOtherEvents$'` + const greps | both tests `--- PASS`; EventTypeApplyOutcome="apply_outcome" present; EventTypeMoaiSubcommand untouched |
| AC-OC-002 | REQ-OC-010 | PASS | `grep LogSchemaVersion="v2"` + `go test -run TestApplyOutcomeEvent_SchemaVersionV2$` | LogSchemaVersion="v2"; recorded event carries schema_version "v2" |
| AC-OC-003 | REQ-OC-003, DD-6 | PASS | `go test -run 'TestRecordOutcome_AppendsOneLine$\|TestRecordOutcome_AutoCreatesDir$'` | one apply_outcome line appended; parent dir auto-created |
| AC-OC-004 | REQ-OC-002/005 | PASS | `go test -run TestApply_Outcome_Kept$` | gate returns nil (non-interference); one apply_outcome event verdict="kept" decision="approved" proposal_id match |
| AC-OC-005 | REQ-OC-002/005 | PASS | `go test -run TestApply_Outcome_RolledBack$` | gate returns *ApplyRegressionError (unchanged), file rolled back; one apply_outcome event verdict="rolled-back" decision="regression-blocked" + regressed list |
| AC-OC-006 | REQ-OC-006/009, C10 | PASS | observer-active Kept/RolledBack + P1 `TestApply_Regression_NonRegressing_Keeps$`/`TestApply_Regression_Blocks_RollsBack$` + `TestApplyOutcome_ReaderTolerance$` | non-interference proven with observer ACTIVE; P1 contract GREEN; AggregatePatterns tolerates apply_outcome events |
| AC-OC-007 | REQ-OC-008, DD-4/DD-2 | PASS | `go test -run 'TestApply_Outcome_GateInactive_NoEmit$\|TestApply_Outcome_NilObserver_NoOp$'` | gate-inactive path emits no apply_outcome event; nil observer = safe no-op |
| AC-OC-008 | REQ-OC-007 | PASS | `go test -run TestApply_Outcome_RecordError_DoesNotFlipVerdict$` | kept change NOT undone; wrapped error surfaced; not an ApplyRegressionError |
| AC-OC-009 | REQ-OC-006, C11, DD-3 | PASS | `grep lineage_id` (none) + `git diff --stat lineage.go` (empty) + `go test -run TestApply_Outcome_ProposalIDMatchesLineage$` | no lineage_id field; lineage.go zero diff; outcome_proposal_id == lineage proposal_id |
| AC-OC-010 | REQ-OC-011, C7 | PASS | `go test -run TestSubagentBoundary_NoAskUserQuestion$` + C-HRA-008 grep | no AskUserQuestion/mcp__askuser in harness/hook Go source (CLEAN) |
| AC-OC-011 | REQ-OC-012, C1-C6 | PASS | FROZEN preservation tests + `git diff --stat` on DO-NOT-MODIFY files + auto_apply grep | preservation tests GREEN; frozen_guard×2/tier/scorer/harness.yaml zero diff; auto_apply: false |
| AC-OC-012 | REQ-OC-013, C12 | PASS | spec.md framing greps (capture+persist / Δ=0 / downstream) + no-overclaim scan | all framing tokens present; no "this spec/enabler improves/prevents" overclaim |

### Cross-platform + quality gates

| Gate | Result |
|------|--------|
| `go build ./...` | exit 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| `go test ./internal/harness/...` | all packages `ok` (incl. P1 regression-gate preservation) |
| Coverage `internal/harness` | 87.3% (package), 87.8% total — ≥ 85% target, ≥ baseline |
| Coverage `internal/measure` | 98.0% (unchanged — leaf still untouched, no new imports) |
| `golangci-lint run ./internal/harness/...` | 0 issues (no NEW lint) |
| C-HRA-008 boundary grep | 0 matches (CLEAN) |
| lineage.go git diff | empty (C11 — lineage schema not extended) |
| FROZEN files git diff | empty (frozen_guard×2 / tier / scorer / harness.yaml) |

### Scope discipline (B10)

Modified/new files limited to the declared scope:
`internal/harness/{types.go, outcome.go (new), applier.go, outcome_test.go (new), observer_test.go}`
+ this SPEC's 4 plan-phase artifacts (status draft→in-progress + this §E.2 evidence). The 13 unrelated
pre-existing working-tree entries (settings.json, deployer.go, web-console-handoff/, other SPEC dirs)
were NOT touched.

### Honest framing reaffirmed (DD-5 / REQ-OC-013)

This SPEC delivers capture + persist ONLY. The captured delta is typically Δ=0 for the current
markdown-only harness write surface; the value is making the Apply outcome OBSERVABLE for downstream
Phase5 analysis (clustering / canary-effectiveness — out of scope). The capture is a passive observer
write of an already-decided outcome; it does NOT alter any Apply decision, does NOT "improve" the
harness, and does NOT "prevent" a regression. Non-interference is proven with the observer ACTIVE
(AC-OC-006), not merely against the bare P1 Applier.

## §E.3 Sync-phase Audit-Ready Signal

Sync-phase documentation (CHANGELOG.md + README assessment + frontmatter transitions + this section) completed by manager-docs.

### Deliverables summary

| Artifact | Change | Status |
|----------|--------|--------|
| CHANGELOG.md | Added `SPEC-HARNESS-OUTCOME-CAPTURE-001` entry (350+ words) under `## [Unreleased]` → `### Added` with honest framing (capture+persist enabler, Δ≈0, downstream Phase5 consumers, 12/12 AC PASS, 87.8% coverage, lineage.go 0-diff, C-HRA-008 clean, plan-auditor PASS 0.91) | ✅ DONE |
| README.md | Assessment: NO change required (internal feature, no user-facing surface) | ✅ DONE |
| spec.md frontmatter | status: `in-progress` → `implemented` | ✅ DONE |
| plan.md frontmatter | status: `in-progress` → `implemented` | ✅ DONE |
| acceptance.md frontmatter | status: `in-progress` → `implemented` | ✅ DONE |
| progress.md frontmatter | status: `in-progress` → `implemented` | ✅ DONE |

### B12 CHANGELOG discipline self-tests (pre-commit)

| Test | Command | Expected | Actual | Result |
|------|---------|----------|--------|--------|
| Pre-emission duplicate check | `grep -c 'SPEC-HARNESS-OUTCOME-CAPTURE-001' CHANGELOG.md` | Count ≥ 1 (exactly 1 for first entry) | 1 | ✅ PASS |
| AC count match | AC rows in acceptance.md (`grep -cE '^\| \*\*AC-OC-' acceptance.md`) should equal CHANGELOG narrative | 12 ACs in narrative text | 12 | ✅ PASS |
| File path verification | `ls` for CHANGELOG.md + 4 SPEC files | All exist | All exist | ✅ PASS |

### Sync commit metadata (pending)

```
Subject: docs(SPEC-HARNESS-OUTCOME-CAPTURE-001): sync-phase artifacts
Scope: CHANGELOG.md, spec.md, plan.md, acceptance.md, progress.md (frontmatter status transitions + this section)
Trailers:
  Authored-By-Agent: manager-docs
  🗿 MoAI <email@mo.ai.kr>
```

sync_commit_sha: 7c75af251
sync_complete_at: 2026-06-14T00:00:00Z
sync_status: ready

---

## §E.4 Mx-phase Audit-Ready Signal

4-phase lifecycle close (plan → run → sync → Mx). Status `implemented` → `completed`.

### sync-auditor verdict (Phase 5 quality gate)

- **PASS-WITH-DEBT 0.86** (Tier M threshold 0.80). 4 must-pass all PASS (Functionality 1.00 / Security 1.00 / Craft 0.75 / Consistency 0.75). Report: `.moai/reports/sync-audit/SPEC-HARNESS-OUTCOME-CAPTURE-001-2026-06-14.md`.
- **F1 (MINOR, resolved)**: CHANGELOG coverage label corrected 87.8% → 87.3% package (orch-direct).
- **F3/F4 (informational)**: OQ2 omitempty zero-drop documented (EC-1); `WithOutcomeObserver` setter 0% cov negligible.

### Known debt carried to P2 activation SPEC

- **F2 (SHOULD-FIX, deferred per user decision + sync-auditor recommendation)**: in `applier.go` `applyWithRegressionGate` rolled-back branch, an `outcomeObserver` write failure returns a wrapped `error` before the typed `*ApplyRegressionError`, which would mask the typed error from `errors.As`. **Currently unreachable** — `WithOutcomeObserver` AND `NewApplierWithRegressionGate` both have zero production callers (dormant scaffold). The kept-path equivalent IS tested (`TestApply_Outcome_RecordError_DoesNotFlipVerdict`); the rolled-back path is not. **P2 activation requirement**: when wiring the observer into production, return `*ApplyRegressionError` as the primary error (via `errors.Join` / typed-wrap) in the rolled-back + failing-observer case, and add `TestApply_Outcome_RolledBack_RecordError`.

### Roadmap position (Self-Harness P2/Phase5)

This enabler (observer OUTCOME capture) is the FIRST sub-item of P2. Downstream consumers remain separate future SPECs: D1 failure-signature clustering (C5) · D3 held-out split validation (C7) · D4 scorer-loop wiring. D2 (LLM proposer / K-candidate, C6) permanently DROPPED.

### Commit metadata

```
mx_commit_sha: (to be recorded after `git commit`)
mx_complete_at: 2026-06-14T00:00:00Z
mx_status: completed
```

4-phase close lineage: plan `d59969efe` + `2cb5e3297` / run `7bd2300cd` + `f7fb94082` / sync `7c75af251` + `f3923bec6` / Mx (this commit + backfill)
