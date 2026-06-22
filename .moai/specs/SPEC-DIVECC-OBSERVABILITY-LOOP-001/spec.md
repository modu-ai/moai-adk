---
id: SPEC-DIVECC-OBSERVABILITY-LOOP-001
title: "Failure-Signature Clustering Engine (Observation arm of the harness-learning loop)"
version: "0.1.0"
status: completed
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/harness, internal/cli"
lifecycle: spec-anchored
tags: "harness, observability, clustering, failure-signature, dogfooding, divecc"
era: V3R6
tier: M
depends_on: [SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001]
---

# SPEC-DIVECC-OBSERVABILITY-LOOP-001 — Failure-Signature Clustering Engine

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-22 | manager-spec | Initial plan-phase draft. Candidate N4 of Epic Dive-into-CC. Premise VERIFIED at authoring (evidence in §B; reproduced by 2 read-only Explore agents this plan-phase). Scope deliberately conservative per user decision Q1/Q2 — read-only clustering + surfacing only. |

---

## §A. Background

### A.1 Epic provenance (Dive-into-CC dogfooding)

This is candidate **N4** of **Epic Dive-into-CC** (see `ROADMAP.md` in `SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/`). The Epic applies findings from an academic reverse-engineering analysis of Claude Code to moai-adk's own harness — a self-improvement (dogfooding) exercise. The source is one body of work on two surfaces: arXiv:2604.14228 "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (Liu, Zhao, Shang, Shen, 2026, cs.SE) and its companion repository github.com/VILA-Lab/Dive-into-Claude-Code.

The paper open-direction that motivates N4 (ROADMAP §N4): **observability closes the loop** — the harness-learning loop is `trace → eval → cluster → policy → repair`. moai-adk currently does failure clustering MANUALLY via 56 `feedback_*.md` lesson files; there is no structured, automated failure-signature clustering over the events the harness already captures.

### A.2 The 5-step loop — `cluster` is the only missing step (premise VERIFIED)

The premise was grounded by read-only inspection of the moai-adk tree at plan-phase and is reproduced in §B as observed fact, not hypothesis, per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 (a defect/gap claim is valid only when the domain's tooling — here `grep`/`Read` over `internal/harness/` — was actually run and its output observed). The five steps of the loop, and their status in moai-adk's tree:

| Step | Status | File:line evidence |
|------|--------|--------------------|
| trace (capture) | **EXISTS** | `internal/harness/observer.go:45,97` (`RecordEvent` / `RecordExtendedEvent`) → `.moai/harness/usage-log.jsonl`. Event schema `internal/harness/types.go:74-204`, `apply_outcome` event type `types.go:58-63`. |
| eval | **EXISTS** | `internal/harness/outcome.go` (`RecordOutcome`), `regression_gate.go` (`MetricTriple`: tests/coverage/lint), `scorer*.go`, `rubric.go`. |
| **cluster** | **MISSING** | `internal/harness/outcome.go:13-16` states verbatim that failure-signature clustering is "downstream and out of scope" of the capture surface. No clustering code exists today. |
| policy (1:1) | **EXISTS** | `internal/harness/proposalgen/mapper.go` (1:1 promotion→candidate). |
| repair | **EXISTS** | `internal/harness/applier.go` + safety L1-L5 + lineage + `Apply`/`Rollback`. |

`cluster` is the sole missing step. N4 fills it — and ONLY it.

### A.3 What this SPEC is (and is not)

This SPEC delivers a **read-only failure-signature clustering engine** plus an **observability surface** (a deterministic report artifact + a CLI read surface). Concretely:

1. A NEW read-only clustering component (`internal/harness/cluster/`) that ingests `apply_outcome` / failure / regression events from `usage-log.jsonl` (and optionally `manifest.jsonl`) and groups them into failure signatures.
2. **Deterministic** signature extraction (NO ML): a signature key derived from fields present on the `apply_outcome` event (sorted `OutcomeRegressed` dimension set + `OutcomeVerdict` + `OutcomeDecision`; NOT `pattern_key`, which is absent from the event and one-way hashed into `OutcomeProposalID` — see REQ-OBL-005). Determinism is required so the engine is Go-test-verifiable (table-driven, `t.TempDir` fixtures).
3. An observability surface: a deterministic report artifact under `.moai/harness/learning-history/` AND a CLI read surface (`moai harness clusters`).
4. Strictly **read-only** w.r.t. the proposal/apply path: the clusterer reads JSONL inputs and writes ONLY its own report file. It never writes back into the proposal/apply path and never alters any Apply decision.

This SPEC is **NOT** a root-cause proposal generator, **NOT** a change to the 1:1 promotion logic, **NOT** an autoApply path, and **NOT** a migration of the 56 manual lessons. Those exclusions are the explicit boundary set by the user (Q1/Q2) and are enumerated in §C.

---

## §B. Grounding facts (premise VERIFIED — reproducible evidence)

> Per `verification-claim-integrity.md` §1.1 surface 3, each fact below names the command run and the output observed at plan-phase. A reviewer can re-run each and reproduce it. These are the substrate the clusterer ingests.

**Fact B1 — clustering is explicitly deferred in-tree.** `internal/harness/outcome.go:13-16` states: the captured outcome record makes the Apply outcome "OBSERVABLE for downstream Phase5 analysis (failure-signature clustering, canary-effectiveness) — those consumers are downstream and out of scope." → the clusterer is the named-but-absent downstream consumer.

**Fact B2 — the `apply_outcome` Event fields the clusterer ingests (and what is NOT present).** `internal/harness/outcome.go:27-47` defines `OutcomeRecord { Verdict ("kept"|"rolled-back"), Decision, ProposalID, Baseline MetricTriple, Candidate MetricTriple, Regressed []string }`. `RecordOutcome` (`outcome.go:57-73`) maps it onto an `apply_outcome` `Event` carrying — as additive omitempty fields on the `Event` struct (`types.go:151-176`) — `OutcomeVerdict` (`outcome_verdict`), `OutcomeDecision` (`outcome_decision`), `OutcomeProposalID` (`outcome_proposal_id`), the baseline+candidate triples, and `OutcomeRegressed []string` (`outcome_regressed`), plus `Subject = "apply:" + rec.ProposalID` (`outcome.go:60`). The load-bearing signature inputs are the fields actually present on the `apply_outcome` event: the `OutcomeRegressed []string` dimension set, the `OutcomeVerdict`, the `OutcomeDecision`, and the `OutcomeProposalID`. **There is NO `pattern_key` field on the `apply_outcome` event** (verified: `grep -n 'pattern_key\|PatternKey' internal/harness/outcome.go` → 0 matches). `pattern_key` lives only on `Promotion` / `ProposalCandidate` records in `tier-promotions.jsonl` (`types.go:280`, `proposalgen/types.go:36`), and `ProposalID = "PROPOSAL-<date>-<sha256(pattern_key)[:8]>"` (`proposalgen/mapper.go:97,102-107`) — `pattern_key` is one-way hashed into the ID and is NOT recoverable from an `apply_outcome` event. `LineageEntry` in `manifest.jsonl` (`types.go:483-502`) likewise has no `pattern_key`. Therefore the signature key is derived ONLY from fields present on the `apply_outcome` event itself (REQ-OBL-005).

**Fact B3 — the `apply_outcome` event type + schema.** `internal/harness/types.go:58-63` defines `EventTypeApplyOutcome EventType = "apply_outcome"`. `LogSchemaVersion = "v2.1"` (`types.go:25`). The Event struct (`types.go:74-204`) carries the additive omitempty `Outcome*` fields the clusterer reads.

**Fact B4 — the data-source paths (all already produced by the harness).**
- `.moai/harness/usage-log.jsonl` — `apply_outcome` (+ other) events; written by `internal/harness/observer.go:13` (`Observer.logPath`).
- `.moai/harness/learning-history/manifest.jsonl` — M6 lineage (`internal/harness/lineage.go`, `applier.go:41`).
- `.moai/harness/learning-history/tier-promotions.jsonl` — promotion records (`internal/harness/learner.go:20,142`).

**Fact B5 — the existing read-only CLI aggregator pattern + its live registration site.** `internal/cli/harness.go:117-183` (`newHarnessStatusCmd` / `runHarnessStatus`) is a read-only aggregator that calls `resolveProjectRoot(cmd)` then `harness.AggregatePatterns(logPath)` and prints a summary. The new `moai harness clusters` subcommand mirrors this exact shape (resolve root → read JSONL → aggregate → print), and per `internal/cli/CLAUDE.md` writes machine-readable output to stdout under `--json` and human output otherwise. The LIVE user-facing harness command tree is `newHarnessRouterCmd` (`internal/cli/harness_route.go:59`), registered in `rootCmd` at `internal/cli/root.go:104`; `clusters` is registered there alongside `status` (`harness_route.go:99`). NOTE: `newHarnessCmd` (`harness.go:63`) is a deprecation-marker tree that is NOT added to `rootCmd` (verified: `grep -rn 'AddCommand(newHarnessCmd' internal/cli/*.go` → 0 matches) — do NOT register `clusters` there.

**Fact B6 — the existing CLI path-resolution convention the clusters command reuses.** The harness CLI resolves the project root via the shared `resolveProjectRoot(cmd)` helper (`internal/cli/harness.go:95-111`). That helper reads the `--project-root` flag (including the inherited parent flag) and, when the flag is empty, falls back to `os.Getwd()` (`harness.go:103-105`). No harness CLI command reads `$CLAUDE_PROJECT_DIR` (verified: `grep -rn 'CLAUDE_PROJECT_DIR\|EnvClaudeProjectDir' internal/cli/harness*.go` → 0 matches; the `EnvClaudeProjectDir` constant at `internal/config/envkeys.go:80` is wired into the hook / session / pre-push paths — `internal/hook/path_resolve.go`, `internal/cli/hook_pre_push.go` — NOT into the harness CLI). The `runHarnessStatus` aggregator (`harness.go:128-129`) is the canonical example: it calls `resolveProjectRoot(cmd)` then `filepath.Join(root, harnessDefaultLogPath)`. The `clusters` command reuses this SAME helper — it introduces NO new divergent path-resolution path (REQ-OBL-011).

---

## §C. Out of Scope

The exclusions below are out of scope for this SPEC. Each is expressed as an `### Out of Scope — <topic>` H3 sub-heading with bullet items, satisfying the `OutOfScopeRule` (`MissingExclusions`) lint. The first three encode the Tier-L alternative the user explicitly REJECTED (Q1/Q2); N4 is observation/surfacing ONLY.

### Out of Scope — Root-cause proposal generation (the rejected Tier-L scope)

- Generating N:M root-cause proposals from clusters. The clusterer surfaces failure signatures as observability; it does NOT synthesize fix proposals.
- Changing the existing 1:1 promotion→candidate logic in `internal/harness/proposalgen/` (`mapper.go`). That path is PRESERVED unchanged.
- Any write-back from the clusterer into the proposal/apply path. The clusterer is read-only w.r.t. that path; it writes only its own report file.

### Out of Scope — Apply / safety pipeline / autonomy

- Modifying `internal/harness/applier.go` (`Apply` / `Rollback`) or any apply-decision logic.
- Modifying the safety pipeline L1-L5 decision logic.
- Changing the `autoApply` default (false, REQ-HL-005 — preserved per user decision Q2) or adding any auto-apply path. N4 preserves the L5 human gate verbatim.
- Modifying `internal/harness/regression_gate.go` (`MetricTriple` is read as input; the gate logic is untouched).

### Out of Scope — Other dormant surfaces and the manual lesson store

- Touching `internal/evolution/learning.go` (a separate dormant surface — clean boundary, left untouched).
- Migrating, replacing, or rewriting the 56 manual `feedback_*.md` lesson files. N4 AUGMENTS observability; it does not migrate the human-curated memory store. The two coexist.

### Out of Scope — Machine learning / non-deterministic clustering

- ML-based, embedding-based, or any non-deterministic clustering. The signature extraction is a deterministic key derivation (no model, no randomness) so the engine is Go-test-verifiable. A clusterer that produced different groupings on different runs would be unverifiable and is excluded.

### Out of Scope — Other Epic Dive-into-CC candidates

- Candidates N2, N3, N5, N6, N7 of Epic Dive-into-CC. Each is its own SPEC; N4 covers only the failure-signature clustering engine + its observability surface.

---

## §D. Requirements (GEARS)

> GEARS notation per `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format. Subjects are generalized (the clusterer, the report, the CLI surface) rather than the hardcoded "the system".

### D.1 Ingestion requirements

- **REQ-OBL-001 (Ubiquitous)**: The clusterer **shall** read `apply_outcome`, failure, and regression events from `.moai/harness/usage-log.jsonl` as its primary input.

- **REQ-OBL-002 (Where)**: **Where** `.moai/harness/learning-history/manifest.jsonl` is present, the clusterer **shall** read it as an optional supplementary input (lineage correlation by `proposal_id`); **where** it is absent, the clusterer **shall** proceed on `usage-log.jsonl` alone without error.

- **REQ-OBL-003 (When)**: **When** the clusterer encounters a malformed or non-`apply_outcome` JSONL line, it **shall** skip that line and continue (fail-open ingestion — a single corrupt line does not abort the run).

- **REQ-OBL-004 (When)**: **When** the input log file is absent or empty, the clusterer **shall** produce an empty cluster set (zero clusters) and exit successfully, NOT an error.

### D.2 Deterministic clustering requirements

- **REQ-OBL-005 (Ubiquitous)**: The clusterer **shall** derive a failure signature key for each ingested failure/rolled-back event deterministically from fields that are actually present on the `apply_outcome` event (`types.go:151-176`): the sorted `OutcomeRegressed` dimension set + the `OutcomeVerdict` + the `OutcomeDecision`, using NO machine learning and NO randomness. The signature key **shall not** depend on `pattern_key`, because `pattern_key` is absent from the `apply_outcome` event and is one-way hashed (`sha256`) into `OutcomeProposalID` (`proposalgen/mapper.go:97,102-107`), so it is not recoverable at the clusterer's ingestion surface. (When a future SPEC needs `pattern_key`-level grouping, it would have to make `tier-promotions.jsonl` a REQUIRED join input keyed on the `PatternKey`→`ProposalID` derivation — explicitly out of scope here, see §C.)

- **REQ-OBL-006 (Ubiquitous)**: The clusterer **shall** group events sharing an identical signature key into one `FailureCluster` carrying: the signature key, the member count, references to the member events, the representative regressed-dimension set, and the first-seen / last-seen timestamps.

- **REQ-OBL-007 (Ubiquitous)**: Given the same input log, the clusterer **shall** produce byte-identical cluster output across repeated runs (deterministic ordering — clusters sorted by a stable key), so the engine is verifiable by table-driven Go tests.

- **REQ-OBL-008 (While)**: **While** an event carries the verdict `"kept"` (a non-failure outcome), the clusterer **shall not** include it in any failure cluster (only failure/`rolled-back` signatures are clustered).

### D.3 Observability surface requirements

- **REQ-OBL-009 (Ubiquitous)**: The clusterer **shall** emit a deterministic report artifact under `.moai/harness/learning-history/` enumerating each `FailureCluster` (signature, count, representative dimensions, first/last seen).

- **REQ-OBL-010 (Ubiquitous)**: A CLI read surface (`moai harness clusters`) **shall** exist that reads the input log, computes the clusters, and prints them — human-readable text by default and machine-readable JSON under `--json` (per the stdout/stderr stream discipline in `internal/cli/CLAUDE.md`).

- **REQ-OBL-011 (Where)**: **Where** the CLI read surface resolves the input/report paths, it **shall** reuse the existing shared `resolveProjectRoot(cmd)` helper (`internal/cli/harness.go:95`) — the SAME root-resolution that `runHarnessStatus` uses (`--project-root` flag with `os.Getwd()` fallback) — and **shall not** introduce a NEW, divergent path-resolution path. (This ALIGNS with the established harness CLI convention rather than diverging from it; no harness command reads `$CLAUDE_PROJECT_DIR` today, per Fact B6, so requiring a bespoke `$CLAUDE_PROJECT_DIR`-only path here would contradict the codebase.)

- **REQ-OBL-012 (When)**: **When** there are zero failure clusters, the CLI read surface **shall** report an empty result (exit 0), NOT an error.

### D.4 Read-only boundary requirements

- **REQ-OBL-013 (Ubiquitous)**: The clusterer **shall** treat `usage-log.jsonl`, `manifest.jsonl`, and `tier-promotions.jsonl` as read-only inputs and **shall** write ONLY its own report artifact under `.moai/harness/learning-history/`.

- **REQ-OBL-014 (Ubiquitous)**: The clusterer **shall not** modify `internal/harness/applier.go`, `internal/harness/proposalgen/`, `internal/harness/regression_gate.go`, the safety L1-L5 logic, the `autoApply` default, or `internal/evolution/learning.go` — it never feeds back into the proposal/apply path.

- **REQ-OBL-015 (Ubiquitous)**: The `internal/harness/cluster/` package **shall not** call `AskUserQuestion` or `mcp__askuser__*` (subagent boundary C-HRA-008 / REQ-PGN-012), verified by a `subagent_boundary` grep gate.

---

## §E. Acceptance Criteria (summary)

Full Given-When-Then acceptance criteria are in `acceptance.md`. Summary of the binding gates:

- AC-OBL-001: the clusterer ingests `apply_outcome` events and groups them by a deterministic signature key derived from fields present on the event (sorted `outcome_regressed` + `outcome_verdict` + `outcome_decision`; NOT `pattern_key`).
- AC-OBL-002: repeated runs over the same input produce byte-identical cluster output (determinism).
- AC-OBL-003: `kept` outcomes are excluded; only failure/`rolled-back` signatures cluster.
- AC-OBL-004: `moai harness clusters` is registered in the LIVE `rootCmd` tree (`--help` exits 0), prints clusters, supports `--json`, reuses `resolveProjectRoot` (no new divergent path), and reports empty (exit 0) on no clusters.
- AC-OBL-005: the clusterer writes ONLY its report under `.moai/harness/learning-history/` — zero diff to applier/proposalgen/regression_gate/`safety/`/autoApply-default/evolution (read-only boundary, full REQ-OBL-014 surface).
- AC-OBL-006: `internal/harness/cluster/` has no `AskUserQuestion` call (subagent boundary grep = 0).
- AC-OBL-007: cross-platform build green (`GOOS=windows GOARCH=amd64 go build ./...`).
- AC-OBL-008: new package coverage ≥ 85%.
- AC-OBL-009: the report artifact is emitted deterministically under `.moai/harness/learning-history/` (byte-identical repeated runs).
- AC-OBL-010: ingestion edge cases — malformed line skipped (fail-open), optional manifest absent tolerated, empty/absent log → 0 clusters + success (REQ-OBL-002/003/004).

---

## §F. Cross-References

- `ROADMAP.md` (in `SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/`) — Epic Dive-into-CC, candidate N4.
- `internal/harness/outcome.go:13-16` — the in-tree statement that failure-signature clustering is the named-but-absent downstream consumer (premise origin).
- `internal/harness/types.go:58-63,142-176` — the `apply_outcome` event type + `Outcome*` field schema the clusterer ingests.
- `internal/harness/observer.go` — the capture surface that produces `usage-log.jsonl`.
- `internal/cli/harness.go:117-183` — the `harness status` read-only aggregator pattern the CLI surface mirrors; `harness.go:95-111` (`resolveProjectRoot`) — the shared root-resolution helper the `clusters` command reuses (REQ-OBL-011).
- `internal/cli/harness_route.go:59,99` (`newHarnessRouterCmd`) + `internal/cli/root.go:104` — the LIVE user-facing harness command tree (registered in `rootCmd`) where `clusters` is registered alongside `status`. (`newHarnessCmd` in `harness.go:63` is a deprecation-marker tree NOT wired to `rootCmd`.)
- `internal/harness/proposalgen/mapper.go:97,102-107` — the `ProposalID = "PROPOSAL-<date>-<sha256(pattern_key)[:8]>"` one-way derivation that makes `pattern_key` unrecoverable at the clusterer's surface (REQ-OBL-005 grounding).
- `internal/cli/CLAUDE.md` — subagent boundary (C-HRA-008), stdout/stderr stream discipline, absolute-path resolution.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — the defect/gap-claim grounding invariant binding §A.2 / §B.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` B3/B11 (subagent boundary) — run-phase known-issue constraints.
