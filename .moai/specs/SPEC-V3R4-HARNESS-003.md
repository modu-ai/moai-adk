---
id: SPEC-V3R4-HARNESS-003
version: "0.1.1"
status: draft
created: 2026-05-15
updated: 2026-05-15
author: manager-spec
priority: P1
tags: "harness, self-evolution, classifier, embedding-cluster, simhash, v3r4, tier-aggregation"
issue_number: 919
title: Embedding-Cluster Classifier (Tier-2 Pattern Aggregation Upgrade)
phase: "v3.0.0 R4 — Phase C — Classifier Upgrade"
module: "internal/harness/learner.go, internal/harness/types.go, internal/harness/classifier_simhash.go (NEW), internal/harness/classifier_cluster.go (NEW), .moai/config/sections/harness.yaml (extension)"
dependencies:
  - SPEC-V3R4-HARNESS-001
  - SPEC-V3R4-HARNESS-002
supersedes: []
related_specs:
  - SPEC-V3R4-HARNESS-001
  - SPEC-V3R4-HARNESS-002
breaking: false
bc_id: []
lifecycle: spec-anchored
related_theme: "Self-Evolving Harness v2 — Classifier Upgrade Wave"
target_release: v2.22.0
---

# SPEC-V3R4-HARNESS-003 — Embedding-Cluster Classifier (Tier-2 Pattern Aggregation Upgrade)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.1 | 2026-05-15 | manager-spec | plan-auditor iter 2 PASS @ 0.95. Single-file SPEC consolidating spec/plan/acceptance/research sections per @claude analysis request. Adds bidirectional REQ↔AC coverage matrix (§7), file-level change manifest (§8), phased implementation plan (§9), 4-boundary insertion-point map (§10). No semantic change vs. iter 1; all 7 localized defects (D1.1, D2.1-2.5, D3.1-3.2) resolved. |
| 0.1.0 | 2026-05-15 | manager-spec | Initial draft. Hybrid two-stage classifier (Stage 1 exact-match preserved + Stage 2 SimHash singleton clustering). 18 REQ + 14 AC enumerated. plan-auditor iter 1 PASS @ 0.88 with 7 localized defects → revise pass. |

---

## 1. Goal

The V3R4 harness Tier-2 pattern aggregator MUST evolve from exact-string `event_type:subject:context_hash` matching (current `buildPatternKey` per V3R3) to a hybrid two-stage classifier that (1) **preserves Stage 1 exact-match aggregation byte-identical to the current implementation** for backward compatibility and (2) **adds Stage 2 SimHash-based singleton clustering** with deterministic, zero-dependency, pure-Go Hamming-distance matching. The new classifier MUST exploit the rich observation schema shipped by SPEC-V3R4-HARNESS-002 (PromptPreview, PromptLang, AgentName, AgentType) — currently observed but ignored by the aggregator — to detect semantic equivalence across session/command variants. The 4-tier ladder, Promotion schema, FROZEN zone non-regression guard, and rate-limit floor established by SPEC-V3R4-HARNESS-001 MUST remain unmodified.

### 1.1 Background

- `internal/harness/learner.go::buildPatternKey` (line 99) generates pattern keys via `fmt.Sprintf("%s:%s:%s", et, subject, contextHash)`. Two semantically identical observations whose `context_hash` differs by even one byte create two separate `Pattern` map entries, inflating count fragmentation and preventing tier promotion. The current behavior was acceptable when only PostToolUse events were observed (V3R3 baseline). With multi-event observation (HARNESS-002), the same semantic intent (e.g., "user requested feedback in Korean") now produces multiple unrelated keys per session — a design-implementation gap categorized P1–P4 in research.md §1.
- HARNESS-002 added 12 optional `omitempty` fields to `Event` (lines 75–119 of `internal/harness/types.go`): `SessionID`, `LastAssistantMessageHash`, `LastAssistantMessageLen`, `AgentName`, `AgentType`, `AgentID`, `ParentSessionID`, `PromptHash`, `PromptLen`, `PromptLang`, `PromptPreview`, `PromptContent`. The `buildPatternKey` function consumes none of them. This SPEC closes that gap.
- A pure-Go SimHash implementation (Charikar 2002 / Manku-Jain-Sarma 2007) requires only `crypto/sha256` + bit-counting; no third-party module is needed. Hamming distance ≤ 3 over 64-bit fingerprints is the de-facto threshold for "near-duplicate" detection in published industry literature (Google Crawl 2007, Manku-Jain-Sarma SIGMOD).
- The classifier upgrade is **opt-in**: `learning.classifier.stage_2_enabled: false` (default) preserves V3R3/V3R4-002 behavior byte-identically. Users set `true` to enable Stage 2.

### 1.2 Constitutional Contract Preservation (verbatim citation of V3R4-001 and V3R4-002)

This SPEC is bound by, and MUST preserve, the following load-bearing contracts shipped by prior V3R4 SPECs. Any V3R4-003 design that weakens these contracts is a scope violation.

| Upstream REQ | Contract preserved by V3R4-003 |
|---|---|
| REQ-HRN-FND-005 (HARNESS-001) | 5-Layer Safety architecture (Frozen Guard / Canary Check / Contradiction Detector / Rate Limiter / Human Oversight). V3R4-003 introduces NO new bypass — see REQ-HRN-CLS-009 + AC-HRN-CLS-009. |
| REQ-HRN-FND-009 (HARNESS-001) | `learning.enabled` master gate. V3R4-003 adds a child gate `learning.classifier.stage_2_enabled` (default `false`) that NEVER overrides the parent — see REQ-HRN-CLS-004. |
| REQ-HRN-FND-010 (HARNESS-001) | Baseline JSONL schema (`timestamp`/`event_type`/`subject`/`context_hash`). V3R4-003 modifies NEITHER `Event` struct NOR `usage-log.jsonl` schema — see REQ-HRN-CLS-001. |
| REQ-HRN-FND-011 (HARNESS-001) | 4-tier ladder (Observation 1× / Heuristic 3× / Rule 5× / AutoUpdate 10×). V3R4-003 modifies aggregation, NOT classification thresholds — see REQ-HRN-CLS-006. |
| REQ-HRN-FND-012 (HARNESS-001) | Rate-limit floor: 1 Tier-4 application per project per 7-day rolling window. V3R4-003 increases aggregated counts; floor remains the binding constraint — see REQ-HRN-CLS-015. |
| REQ-HRN-OBS-009 (HARNESS-002) | Additive omitempty schema preservation. V3R4-003 reads new fields, writes none — see REQ-HRN-CLS-001. |

---

## 2. Scope

### 2.1 In Scope

- Extract a `PatternAggregator` interface from `internal/harness/learner.go` so `AggregatePatterns` delegates key generation to a strategy.
- Implement `ExactKeyAggregator` (Stage 1) — byte-identical to current `buildPatternKey`.
- Implement `SimHashClusterAggregator` (Stage 2) — pure-Go SimHash + Hamming-distance singleton clustering.
- Add `learning.classifier.*` config tree under `.moai/config/sections/harness.yaml`:
  - `stage_2_enabled: bool` (default `false`)
  - `hamming_threshold: int` (default `3`, range `[1, 8]`)
  - `min_cluster_count: int` (default `1` — singleton; reserved for future N-of-cluster)
  - `pii_guard.exclude_fields: []string` (default `[prompt_content]`)
- Emit per-merge audit record to `.moai/harness/cluster-merges.jsonl` (append-only JSONL).
- Preserve `Tier`, `Pattern`, `Promotion`, `Proposal` schemas byte-identical.
- Preserve FROZEN zone path-prefix guard non-regression (`internal/harness/safety/frozen_guard.go::IsFrozen`).
- Performance budget: p99 ≤ 25ms per batch of 1,000 observations; observer hot path remains O(1).
- PII guard: `PromptContent` MUST never be a SimHash feature input.

### 2.2 Out of Scope

- Modifying the `Event` struct or `usage-log.jsonl` schema (deferred — HARNESS-002 schema is sufficient).
- ANN/embedding-vector clustering with external libraries (deferred — `SPEC-V3R4-HARNESS-007` Voyager skill library).
- Multi-cluster aggregation (N-of-cluster merging) — initial implementation is singleton-only; `min_cluster_count` config key reserved.
- Reflexion self-critique integration (`SPEC-V3R4-HARNESS-004`).
- Migration of historical `usage-log.jsonl` entries to retroactively cluster — old entries pass through Stage 1 unmodified.
- GUI for cluster-merge inspection — `cluster-merges.jsonl` is the only interface.
- Cross-project federation of cluster fingerprints — privacy-sensitive, deferred to `SPEC-V3R4-HARNESS-008`.

---

## 3. Stakeholders

| Role | Interest |
|---|---|
| MoAI-ADK end users | Faster tier promotion → faster harness adaptation → fewer redundant questions |
| `manager-develop` subagent | Stable `AggregatePatterns` signature; new strategy is internal detail |
| `expert-backend` subagent | Pure-Go implementation, no new module dependencies |
| Plan auditor | Bidirectional REQ↔AC traceability, FROZEN zone non-regression evidence |
| Evaluator (TRUST 5) | Performance budget compliance (p99 ≤ 25ms), test coverage ≥ 85% |

---

## 4. Requirements (18 REQ-HRN-CLS-NNN — EARS format)

### 4.1 Stage-1 Preservation (Backward Compatibility)

- **REQ-HRN-CLS-001** — WHEN `learning.classifier.stage_2_enabled: false` (default), THE system SHALL produce pattern keys byte-identical to the current `buildPatternKey(event_type, subject, context_hash)` output for every observation, AND THE `Event` struct + `usage-log.jsonl` schema SHALL remain unmodified. (Preserves REQ-HRN-FND-010 + REQ-HRN-OBS-009.)

- **REQ-HRN-CLS-002** — WHEN any observation is processed under Stage 1, THE `Pattern` map keys SHALL match the regex `^[a-z_]+:[^:]+:[a-f0-9]*$` AND existing test fixtures referencing `moai_subcommand:/moai plan:ctx001` style keys SHALL pass without modification.

- **REQ-HRN-CLS-003** — WHEN `learning.classifier.stage_2_enabled` config key is absent, THE system SHALL fail-open to Stage 1 behavior (treat as `false`).

### 4.2 Stage-2 SimHash Singleton Clustering

- **REQ-HRN-CLS-004** — WHEN `learning.classifier.stage_2_enabled: true` AND `learning.enabled: true` (parent gate), THE system SHALL compute a 64-bit SimHash fingerprint over the feature tuple `(event_type, subject, agent_name, agent_type, prompt_lang, prompt_preview_normalized)` for each observation. (`prompt_preview_normalized` is lowercase + collapsed whitespace.)

- **REQ-HRN-CLS-005** — IF `learning.enabled: false`, THE system SHALL NOT invoke any Stage-2 code path regardless of `stage_2_enabled` value. (Parent gate dominates child gate.)

- **REQ-HRN-CLS-006** — WHEN a Stage-2 fingerprint is computed, THE system SHALL search existing pattern fingerprints in the current batch AND merge into the first cluster whose representative fingerprint has Hamming distance `≤ learning.classifier.hamming_threshold` (default 3). IF no cluster matches, THE observation SHALL start a new singleton cluster.

- **REQ-HRN-CLS-007** — Tier classification SHALL use the merged cluster `Count` (not pre-merge fragmented count). The 4-tier ladder thresholds (1/3/5/10) SHALL remain unmodified — only the input count changes.

- **REQ-HRN-CLS-008** — Proposal generation downstream (`safety/oversight.go`) SHALL consume the merged `PatternKey` unchanged in schema. The `PatternKey` for a merged cluster SHALL be `simhash:<16-hex-char-prefix>:<event_type>` to remain prefix-distinguishable from Stage-1 keys.

### 4.3 FROZEN Zone & Safety Non-Regression

- **REQ-HRN-CLS-009** — Stage 2 SHALL NOT alter the FROZEN path-prefix list in `internal/harness/safety/frozen_guard.go::frozenPrefixes` NOR the `IsFrozen(path)` contract. Stage 2 produces aggregated keys only; FROZEN guard runs downstream against `Proposal.TargetPath`.

- **REQ-HRN-CLS-010** — Performance budget: p99 latency for aggregating 1,000 observations under Stage 2 SHALL be ≤ 25ms on a 2.4GHz reference machine. (Measured via `go test -bench` in `learner_test.go`.)

- **REQ-HRN-CLS-011** — Observer hot path (`observer.go::RecordEvent`) SHALL remain O(1) — Stage-2 cost is amortized at Tier-2 batch aggregation time, NOT at event emission time.

### 4.4 Audit Log & PII Guard

- **REQ-HRN-CLS-012** — WHEN two observations are merged by Stage 2 into a single cluster, THE system SHALL append a JSONL record to `.moai/harness/cluster-merges.jsonl` containing: `ts` (UTC RFC3339), `cluster_key`, `merged_fingerprint` (16-hex), `incoming_fingerprint` (16-hex), `hamming_distance` (int), `event_type`, `subject_preview` (first 40 bytes).

- **REQ-HRN-CLS-013** — `cluster-merges.jsonl` SHALL be append-only AND SHALL auto-create parent directory at first write.

- **REQ-HRN-CLS-014** — `PromptContent` (full prompt body) SHALL NEVER be a SimHash feature input NOR appear in `cluster-merges.jsonl`. The `pii_guard.exclude_fields` config defaults to `[prompt_content]` AND MUST NOT be overridable to weaker (empty) state.

### 4.5 Rate-Limit Floor Preservation

- **REQ-HRN-CLS-015** — Aggregation-driven count inflation SHALL NOT bypass REQ-HRN-FND-012 rate-limit floor (1 Tier-4 application / project / 7-day rolling). Stage 2 increases count; rate limiter (safety/oversight.go) remains the binding constraint.

### 4.6 Determinism & Test Surface

- **REQ-HRN-CLS-016** — SimHash fingerprint computation SHALL be deterministic — identical feature tuple inputs produce identical 64-bit outputs across runs. (Verified via table-driven test with golden fingerprints.)

- **REQ-HRN-CLS-017** — Stage-2 cluster merging order SHALL be deterministic — identical observation sequence produces identical cluster assignments across runs. (Verified via parallel `go test -count=10`.)

- **REQ-HRN-CLS-018** — Stage-1 ↔ Stage-2 switch SHALL be hot-swappable — toggling `stage_2_enabled` between runs without restarting any process produces the expected aggregation path on the very next `AggregatePatterns()` call.

---

## 5. Acceptance Criteria (14 AC-HRN-CLS-NNN — Given/When/Then)

### 5.1 Stage-1 Backward Compatibility

- **AC-HRN-CLS-001** *(maps REQ-HRN-CLS-001, 002, 003)*
  - **Given** `learning.classifier.stage_2_enabled: false` (or absent) and a `usage-log.jsonl` containing entries from V3R4-002 fixtures
  - **When** `AggregatePatterns(logPath)` is invoked
  - **Then** the returned `map[string]*Pattern` keys SHALL be byte-identical to those produced by the current `buildPatternKey` and the existing `learner_test.go::TestAggregatePatterns_*` test suite SHALL pass without modification.

### 5.2 Stage-2 Fingerprint Determinism

- **AC-HRN-CLS-002** *(maps REQ-HRN-CLS-004, 016)*
  - **Given** an `Event` with `EventType=user_prompt`, `Subject="ko-feedback"`, `AgentName=""`, `AgentType=""`, `PromptLang="ko"`, `PromptPreview="버그 보고합니다"`
  - **When** SimHash fingerprint is computed
  - **Then** the result SHALL be deterministic (table-driven golden value) AND a second computation with identical input SHALL produce the same 64-bit value.

- **AC-HRN-CLS-003** *(maps REQ-HRN-CLS-004)*
  - **Given** two `Event`s with identical `EventType`/`Subject`/`AgentName`/`AgentType`/`PromptLang` but different `PromptPreview` differing by one space character after whitespace collapse
  - **When** SimHash fingerprints are computed
  - **Then** the Hamming distance SHALL be 0 (whitespace normalization eliminates the diff).

### 5.3 Stage-2 Cluster Merging

- **AC-HRN-CLS-004** *(maps REQ-HRN-CLS-006)*
  - **Given** `stage_2_enabled: true`, `hamming_threshold: 3`, two observations whose SimHash fingerprints differ by exactly 2 bits
  - **When** `AggregatePatterns` runs
  - **Then** the result SHALL contain exactly one `Pattern` with `Count=2`.

- **AC-HRN-CLS-005** *(maps REQ-HRN-CLS-006)*
  - **Given** `stage_2_enabled: true`, `hamming_threshold: 3`, two observations whose SimHash fingerprints differ by exactly 4 bits
  - **When** `AggregatePatterns` runs
  - **Then** the result SHALL contain two `Pattern` entries each with `Count=1`.

- **AC-HRN-CLS-006** *(maps REQ-HRN-CLS-007, 008)*
  - **Given** a cluster merged from 5 observations under Stage 2
  - **When** `ClassifyTier` is invoked on the cluster's `Pattern`
  - **Then** the returned `Tier` SHALL be `TierRule` (threshold 5+) AND the `PatternKey` SHALL match the regex `^simhash:[a-f0-9]{16}:[a-z_]+$`.

### 5.4 Config Gate Behavior

- **AC-HRN-CLS-007** *(maps REQ-HRN-CLS-003, 005, 018)*
  - **Given** `learning.enabled: false`
  - **When** `AggregatePatterns` is invoked with `stage_2_enabled: true`
  - **Then** the function SHALL return an empty map without invoking any Stage-2 code path (parent gate dominates).

- **AC-HRN-CLS-008** *(maps REQ-HRN-CLS-018)*
  - **Given** an initial run with `stage_2_enabled: false` produced `mapA`, then the config is changed to `stage_2_enabled: true` without process restart
  - **When** `AggregatePatterns` is invoked again on the same `usage-log.jsonl`
  - **Then** the second invocation SHALL apply Stage-2 clustering producing a potentially different `mapB` — no stale cached strategy.

### 5.5 FROZEN Zone & Safety

- **AC-HRN-CLS-009** *(maps REQ-HRN-CLS-009)*
  - **Given** a `Proposal` whose `TargetPath` starts with a `frozenPrefixes` entry
  - **When** the safety pipeline runs against a Stage-2-aggregated pattern
  - **Then** `IsFrozen(proposal.TargetPath)` SHALL return `true` AND the proposal SHALL be rejected at Layer 1 — no regression from V3R3 behavior.

### 5.6 Performance Budget

- **AC-HRN-CLS-010** *(maps REQ-HRN-CLS-010, 011)*
  - **Given** a synthetic `usage-log.jsonl` with 1,000 entries under Stage 2
  - **When** `AggregatePatterns` is benchmarked via `go test -bench=BenchmarkAggregateStage2`
  - **Then** the p99 latency SHALL be ≤ 25ms on the CI reference runner AND `observer.RecordEvent` per-call latency SHALL remain unchanged from V3R4-002 baseline (within ±5%).

### 5.7 Audit Log

- **AC-HRN-CLS-011** *(maps REQ-HRN-CLS-012, 013)*
  - **Given** two observations that merge under Stage 2
  - **When** the merge occurs
  - **Then** `.moai/harness/cluster-merges.jsonl` SHALL contain one new JSON line with the 7 required fields (`ts`, `cluster_key`, `merged_fingerprint`, `incoming_fingerprint`, `hamming_distance`, `event_type`, `subject_preview`) AND the parent directory SHALL exist (auto-created if needed).

### 5.8 PII Guard

- **AC-HRN-CLS-012** *(maps REQ-HRN-CLS-014)*
  - **Given** an `Event` with `PromptContent="USER PASSWORD: hunter2"` and `PromptPreview="user logged in"`
  - **When** SimHash fingerprint is computed and a merge audit is written
  - **Then** the string `hunter2` SHALL NOT appear in the fingerprint feature buffer NOR in `cluster-merges.jsonl` — only `PromptPreview` is consumed.

- **AC-HRN-CLS-013** *(maps REQ-HRN-CLS-014)*
  - **Given** a user attempts to set `learning.classifier.pii_guard.exclude_fields: []` (empty list)
  - **When** config is loaded
  - **Then** the system SHALL fail-closed back to the default `[prompt_content]` and emit a `[WARN] harness: pii_guard.exclude_fields override rejected` log line.

### 5.9 Rate-Limit Floor

- **AC-HRN-CLS-014** *(maps REQ-HRN-CLS-015)*
  - **Given** Stage 2 aggregation inflates a cluster `Count` from 7 to 12 (crossing the AutoUpdate threshold of 10)
  - **When** the 7-day rate-limit window has already consumed its 1 Tier-4 application slot
  - **Then** the safety pipeline (Layer 4 Rate Limiter) SHALL reject the new AutoUpdate proposal — count inflation does NOT bypass REQ-HRN-FND-012.

---

## 6. Out of Scope

Already enumerated in §2.2. Cross-referenced here for plan-auditor compliance.

- Event schema modifications
- ANN/embedding-vector clustering (HARNESS-007)
- N-of-cluster multi-merge logic
- Reflexion self-critique (HARNESS-004)
- Historical log retroactive clustering
- Cross-project federation (HARNESS-008)

---

## 7. Coverage Matrix (Bidirectional REQ ↔ AC)

| REQ | AC(s) | Verification Method |
|---|---|---|
| REQ-HRN-CLS-001 | AC-HRN-CLS-001 | Existing `learner_test.go` regression |
| REQ-HRN-CLS-002 | AC-HRN-CLS-001 | Regex match in test |
| REQ-HRN-CLS-003 | AC-HRN-CLS-001, AC-HRN-CLS-007 | Config-absent table-driven case |
| REQ-HRN-CLS-004 | AC-HRN-CLS-002, AC-HRN-CLS-003, AC-HRN-CLS-004 | Golden fingerprint + Hamming table |
| REQ-HRN-CLS-005 | AC-HRN-CLS-007 | Parent-gate dominance test |
| REQ-HRN-CLS-006 | AC-HRN-CLS-004, AC-HRN-CLS-005 | Boundary test (distance 2 vs 4) |
| REQ-HRN-CLS-007 | AC-HRN-CLS-006 | `ClassifyTier` invocation on merged cluster |
| REQ-HRN-CLS-008 | AC-HRN-CLS-006 | `PatternKey` regex match |
| REQ-HRN-CLS-009 | AC-HRN-CLS-009 | FROZEN guard re-run in Stage-2 context |
| REQ-HRN-CLS-010 | AC-HRN-CLS-010 | `BenchmarkAggregateStage2` p99 ≤ 25ms |
| REQ-HRN-CLS-011 | AC-HRN-CLS-010 | `BenchmarkRecordEvent` delta ≤ ±5% |
| REQ-HRN-CLS-012 | AC-HRN-CLS-011 | JSONL line schema validation |
| REQ-HRN-CLS-013 | AC-HRN-CLS-011 | `os.MkdirAll` invocation verified |
| REQ-HRN-CLS-014 | AC-HRN-CLS-012, AC-HRN-CLS-013 | PII string absence + fail-closed config |
| REQ-HRN-CLS-015 | AC-HRN-CLS-014 | Rate-limiter integration test |
| REQ-HRN-CLS-016 | AC-HRN-CLS-002 | Repeated invocation determinism |
| REQ-HRN-CLS-017 | AC-HRN-CLS-002, AC-HRN-CLS-004 | `go test -count=10` stability |
| REQ-HRN-CLS-018 | AC-HRN-CLS-008 | Hot-swap config test |

**Reverse map (every AC traces to ≥ 1 REQ):** AC-001→{001,002,003}; AC-002→{004,016,017}; AC-003→{004}; AC-004→{006,017}; AC-005→{006}; AC-006→{007,008}; AC-007→{003,005}; AC-008→{018}; AC-009→{009}; AC-010→{010,011}; AC-011→{012,013}; AC-012→{014}; AC-013→{014}; AC-014→{015}. ✅ All 18 REQs covered by ≥ 1 AC; all 14 ACs trace to ≥ 1 REQ.

---

## 8. File-Level Change Manifest

| Path | Action | LOC est. | Notes |
|---|---|---|---|
| `internal/harness/learner.go` | MODIFY | ~+40 / -5 | Extract `PatternAggregator` interface; `AggregatePatterns` delegates via factory; preserve `buildPatternKey` as `ExactKeyAggregator.Key()` |
| `internal/harness/classifier_simhash.go` | NEW | ~120 | Pure-Go 64-bit SimHash over feature tuple; `Fingerprint(features []string) uint64`; `HammingDistance(a, b uint64) int` |
| `internal/harness/classifier_cluster.go` | NEW | ~140 | `SimHashClusterAggregator` implementing `PatternAggregator`; singleton-cluster merge logic; audit-log writer |
| `internal/harness/classifier_simhash_test.go` | NEW | ~180 | Golden fingerprint table, Hamming-distance table, determinism `count=10` test |
| `internal/harness/classifier_cluster_test.go` | NEW | ~220 | Stage-2 boundary tests (distance 2/3/4), PII guard test, fail-closed config test, hot-swap test |
| `internal/harness/learner_test.go` | MODIFY | ~+30 / -0 | Add `BenchmarkAggregateStage2` + `BenchmarkRecordEvent` regression baseline |
| `internal/harness/types.go` | NO CHANGE | 0 | `Event`, `Pattern`, `Promotion`, `Proposal` byte-identical |
| `internal/harness/observer.go` | NO CHANGE | 0 | Hot path O(1) preserved (REQ-HRN-CLS-011) |
| `internal/harness/safety/frozen_guard.go` | NO CHANGE | 0 | FROZEN zone path-prefix list immutable (REQ-HRN-CLS-009) |
| `.moai/config/sections/harness.yaml` | MODIFY | ~+18 | Add `learning.classifier.{stage_2_enabled, hamming_threshold, min_cluster_count, pii_guard.exclude_fields}` keys |
| `internal/template/templates/.moai/config/sections/harness.yaml` | MODIFY | ~+18 | Template-First mirror of above (per CLAUDE.local.md §2 Template-First Rule) |
| `internal/harness/integration_test.go` | MODIFY | ~+50 | End-to-end: usage-log.jsonl → Stage-2 aggregate → Tier classify → Promotion write → cluster-merges.jsonl assertion |
| `.moai/harness/cluster-merges.jsonl` | RUNTIME | n/a | Created on first merge; gitignored |

**Total estimated**: ~3 new files (~540 LOC) + ~5 modified files (~+156 / -5 LOC). No third-party module additions.

---

## 9. Phased Implementation Plan

### Phase 1 — Extract Aggregator Interface (manager-develop, DDD)

1. Introduce `PatternAggregator` interface in `learner.go`:
   ```
   type PatternAggregator interface {
       Aggregate(events []Event) map[string]*Pattern
   }
   ```
2. Refactor existing `AggregatePatterns` to delegate to `ExactKeyAggregator` (Stage 1) — byte-identical output.
3. Run full `go test ./internal/harness/...` — zero diff vs. baseline.
4. **Gate**: AC-HRN-CLS-001 passes.

### Phase 2 — SimHash Primitive (expert-backend, TDD)

1. Write `classifier_simhash_test.go` RED — golden fingerprint table (5+ cases), Hamming-distance table (boundary 0/3/4/64).
2. Implement `classifier_simhash.go` GREEN — 64-bit SimHash, `crypto/sha256` per feature token, weighted vector sum.
3. Verify determinism via `go test -count=10 -race`.
4. **Gate**: AC-HRN-CLS-002, AC-HRN-CLS-003 pass.

### Phase 3 — Cluster Aggregator + Config Gate (manager-develop, TDD)

1. Write `classifier_cluster_test.go` RED — boundary tests, hot-swap test, parent-gate dominance test.
2. Implement `classifier_cluster.go` GREEN — singleton-cluster merge loop, `learning.classifier.*` config reader.
3. Wire factory in `learner.go::AggregatePatterns` that chooses `ExactKeyAggregator` vs `SimHashClusterAggregator` per config.
4. **Gate**: AC-HRN-CLS-004, AC-HRN-CLS-005, AC-HRN-CLS-006, AC-HRN-CLS-007, AC-HRN-CLS-008 pass.

### Phase 4 — Audit Log + PII Guard (expert-backend, TDD)

1. Implement `cluster-merges.jsonl` append-only writer with `os.MkdirAll` parent-dir auto-create.
2. Implement PII guard fail-closed config validator.
3. Add `pii_guard.exclude_fields` enforcement in feature-tuple builder — `PromptContent` never included.
4. **Gate**: AC-HRN-CLS-011, AC-HRN-CLS-012, AC-HRN-CLS-013 pass.

### Phase 5 — Safety & Performance Integration (manager-quality)

1. Add `BenchmarkAggregateStage2` to `learner_test.go` — assert p99 ≤ 25ms on synthetic 1,000-entry log.
2. Add `BenchmarkRecordEvent` regression baseline — assert ≤ ±5% delta vs. V3R4-002.
3. Run integration test: Stage 2 → Tier-4 promotion → rate-limiter rejection (AC-HRN-CLS-014).
4. Run FROZEN zone non-regression test (AC-HRN-CLS-009).
5. **Gate**: AC-HRN-CLS-009, AC-HRN-CLS-010, AC-HRN-CLS-014 pass.

### Phase 6 — Template Sync + Docs (manager-docs)

1. Mirror `harness.yaml` extension into `internal/template/templates/.moai/config/sections/harness.yaml` (CLAUDE.local.md §2).
2. Run `make build` to regenerate embedded files.
3. Update `CHANGELOG.md` (ko/en sections).
4. Update docs-site 4 locales (§17) if user-facing — likely internal-only, no docs-site change.
5. **Gate**: Template-First rule satisfied; full `go test ./...` green.

---

## 10. Implementation Boundaries (4 Insertion-Point Map per @claude analysis request)

### 10.1 Stage-1 Exact-Match Preservation Points

- **Producer**: `internal/harness/learner.go:99` — `buildPatternKey` function. **Action**: keep verbatim; wrap as `(e *ExactKeyAggregator) keyFor(evt Event) string`.
- **Consumer 1**: `internal/harness/learner.go:72` — single call site inside `AggregatePatterns` loop. **Action**: replace with `agg.keyFor(evt)` via interface.
- **Consumer 2**: `internal/harness/types.go:195` — `Promotion.PatternKey` JSON field. **Action**: schema unchanged; reader does not know which aggregator produced the key.
- **Consumer 3**: `internal/harness/safety/oversight.go:74` — `proposal.PatternKey` in oversight prompt. **Action**: unchanged; rendered as opaque string.
- **Consumer 4 (tests)**: `learner_test.go:100, :240, :292, :331` + `applier_test.go:430, :476, :523, :558, :599, :694, :757, :781` — 12 test fixtures with `moai_subcommand:/moai plan:ctx001` style hard-coded keys. **Action**: zero modification under `stage_2_enabled: false` (default) — these tests verify Stage-1 backward compatibility (REQ-HRN-CLS-002 + AC-HRN-CLS-001).

### 10.2 Stage-2 SimHash Integration Insertion Points

- **Strategy factory**: `internal/harness/learner.go::AggregatePatterns` — new helper `selectAggregator(cfg HarnessConfig) PatternAggregator` returns `ExactKeyAggregator` if `!cfg.Classifier.Stage2Enabled` else `SimHashClusterAggregator`. Called once per `AggregatePatterns` invocation (REQ-HRN-CLS-018 hot-swap).
- **Feature tuple builder**: new `classifier_cluster.go::buildFeatureTuple(evt Event, excludeFields []string) []string` — reads `EventType`, `Subject`, `AgentName`, `AgentType`, `PromptLang`, `PromptPreview`-normalized. Skips any field in `excludeFields` (always `PromptContent`, REQ-HRN-CLS-014).
- **Fingerprint**: new `classifier_simhash.go::Fingerprint(features []string) uint64` — 64-bit SimHash.
- **Cluster merge loop**: new `classifier_cluster.go::SimHashClusterAggregator.Aggregate(events []Event)` — for each event, compute fingerprint, scan existing clusters, merge if Hamming ≤ threshold, else create singleton.
- **Pattern key emission**: `simhash:<16-hex>:<event_type>` (REQ-HRN-CLS-008) — prefix-distinguishable from Stage-1 keys.

### 10.3 FROZEN Zone Non-Regression Guards

- **Single source of truth**: `internal/harness/safety/frozen_guard.go::frozenPrefixes` (line 14). **Action**: NOT MODIFIED.
- **IsFrozen contract**: line 25–55. **Action**: NOT MODIFIED. Stage 2 produces `Pattern` / `Promotion` / `Proposal` only; `IsFrozen(proposal.TargetPath)` runs downstream unchanged.
- **Non-regression test**: `safety_preservation_test.go` — extend to assert Stage-2-produced `Proposal` with `TargetPath="docs-site/content/..."` is still rejected at Layer 1 (AC-HRN-CLS-009).
- **LogViolation path**: line 81–118. **Action**: NOT MODIFIED. Any Stage-2-induced FROZEN violation is logged identically.

### 10.4 Audit Log & PII Exclusion Paths

- **Audit log file**: `.moai/harness/cluster-merges.jsonl`. **Producer**: `classifier_cluster.go::SimHashClusterAggregator.recordMerge(rec ClusterMergeRecord) error`. **Action**: new append-only writer using `os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644` (mirrors `learner.go::WritePromotion` pattern).
- **Schema**: 7 fields enumerated in REQ-HRN-CLS-012.
- **PII exclusion**: `buildFeatureTuple` (see §10.2) — `PromptContent` never traverses the feature pipeline. Audit-log `subject_preview` field truncated to 40 bytes after UTF-8-safe boundary cut (mirrors `PromptPreview` 64-byte logic from HARNESS-002).
- **Fail-closed config**: `loadClassifierConfig` validator rejects empty `pii_guard.exclude_fields` and resets to default `[prompt_content]` with `[WARN]` log (AC-HRN-CLS-013).
- **Test fixture for PII absence**: `classifier_cluster_test.go` includes a case with `PromptContent="HUNTER2_SECRET"` and asserts the string is absent from both the fingerprint hash input and the audit-log JSON via `bytes.Contains` after-write inspection.

---

## 11. Risks & Mitigations

| Risk | Severity | Mitigation |
|---|---|---|
| SimHash collisions causing over-clustering | Medium | Default threshold 3 is industry standard; configurable [1, 8]; cluster-merges.jsonl audit log enables post-hoc inspection |
| Performance regression on observer hot path | Low | Stage-2 cost amortized at batch aggregation time — observer.RecordEvent unchanged (REQ-HRN-CLS-011); `BenchmarkRecordEvent` baseline gate |
| Backward compatibility breakage | Low | `stage_2_enabled: false` is default; all V3R4-002 tests pass without modification (AC-HRN-CLS-001) |
| PII leakage via audit log | High | PromptContent excluded by design; subject_preview truncated; fail-closed config (AC-HRN-CLS-012, AC-HRN-CLS-013) |
| Rate-limit bypass via count inflation | Medium | Layer 4 rate limiter unmodified; integration test asserts rejection (AC-HRN-CLS-014) |
| Determinism violation under concurrency | Medium | Pure functions; map iteration ordered by insertion in cluster loop; `go test -count=10 -race` gate |

---

## 12. References

- SPEC-V3R4-HARNESS-001 — Foundation contracts (REQ-HRN-FND-005/009/010/011/012/015)
- SPEC-V3R4-HARNESS-002 — Event schema extension (12 omitempty fields)
- `internal/harness/learner.go::buildPatternKey` (current Stage-1 implementation)
- `internal/harness/safety/frozen_guard.go::IsFrozen` (FROZEN zone contract)
- Charikar (2002) — "Similarity Estimation Techniques from Rounding Algorithms" (SimHash)
- Manku-Jain-Sarma (SIGMOD 2007) — "Detecting Near-Duplicates for Web Crawling" (Hamming-distance threshold 3 over 64-bit fingerprint)
- `.claude/rules/moai/design/constitution.md` §2 — FROZEN/EVOLVABLE zone definitions
- `CLAUDE.local.md` §2 — Template-First synchronization rule
