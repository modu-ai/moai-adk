---
id: SPEC-V3R4-HARNESS-003
version: "0.1.1"
status: draft
created: 2026-05-15
updated: 2026-05-15
author: manager-spec
priority: P1
tags: "harness, self-evolution, classifier, embedding, simhash, clustering, v3r4, phase-c, tier-2-aggregation"
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

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-15 | manager-spec | Initial plan-phase draft. Phase C of the v3.0.0 R4 Self-Evolving Harness program. Introduces a hybrid two-stage classifier (Stage 1 exact-match preserved verbatim; Stage 2 SimHash + Hamming-distance singleton clustering, opt-in via `learning.classifier.stage_2_enabled`) that aggregates semantically similar Tier-1 singletons into Tier-2+ candidates without breaking the SPEC-V3R4-HARNESS-001 frequency-count ladder (REQ-HRN-FND-011). Consumes the Multi-Event observation surface delivered by SPEC-V3R4-HARNESS-002 (PromptPreview, PromptLang, AgentName, AgentType). Preserves the 5-Layer Safety pipeline (REQ-HRN-FND-005), FROZEN zone immutability (REQ-HRN-FND-006), the 7-day Tier-4 rate-limit floor (REQ-HRN-FND-012), and the orchestrator-only AskUserQuestion contract (REQ-HRN-FND-015). Adds an audit log of cluster-merge decisions at `.moai/harness/cluster-merges.jsonl`. Adopts research.md Option D+A recommendation. Non-breaking, schema-additive, fail-open to Stage-1-only behavior on any configuration parse error or library failure. |
| 0.1.1   | 2026-05-15 | manager-spec | plan-audit iter 1 PASS revise pass — D1.1, D2.1-D2.5, D3.1-D3.2 defects addressed. D1.1: corrected Pattern field count six→seven in §8 Risk row 4. D2.1: aligned default `hamming_threshold` from 12 to 3 (research.md §8.3 recommended range 2-5 in 64-bit space) across §1.2(1), §5 REQ-HRN-CLS-003, §6, §8 Risk row 1 (rationale rewritten), and all acceptance.md references. D2.2: AC-HRN-CLS-002 fixture redesigned to share `prompt_preview` across all 10 events for realistic Tier-2 clustering. D2.3: clarified AC-HRN-CLS-002 outcome wording from "Either/AND" to "exactly 11 entries". D2.4: locked REQ-HRN-CLS-013 `hamming_distances` schema to flat upper-triangle row-major array (dropped triangular-matrix alternative). D2.5: added 6th AC-HRN-CLS-014 sub-case for YAML type-mismatch errors. D3.1: aligned plan.md §3.2 SimHash decision to Wave B-finalized "embed locally". D3.2: capped `hamming_distances` at 20 pair distances with `truncated`/`hamming_pair_count` flags to protect downstream JSONL parsers. |

---

## 1. Goal

The V3R4 self-evolving harness classifier surface MUST evolve from the current exact-match-only pattern aggregator (`buildPatternKey(et, subject, context_hash)` at `internal/harness/learner.go:99`) to a hybrid two-stage classifier that preserves Stage 1 verbatim and adds an opt-in Stage 2 layer of SimHash-based singleton clustering. Stage 2 aggregates semantically similar Tier-1 singleton patterns (typically variants of the same intent that differ only in session context or minor lexical form) into virtual merged patterns eligible for higher-tier promotion. The promotion ladder thresholds (1 / 3 / 5 / 10 per `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-011) MUST remain unchanged. The 5-Layer Safety pipeline, the FROZEN zone path-prefix protection, the 7-day Tier-4 rate-limit floor, and the orchestrator-only AskUserQuestion contract all from `SPEC-V3R4-HARNESS-001` MUST remain intact and unmodified by V3R4-003. The extended Event schema delivered by `SPEC-V3R4-HARNESS-002` (PromptPreview, PromptLang, AgentName, AgentType, plus 8 more `omitempty` fields) MUST be the semantic feature source for Stage 2 fingerprinting; PromptContent (opt-in full text, Strategy C) MUST NEVER enter any classifier artifact or fingerprint input.

### 1.1 Background

- `SPEC-V3R4-HARNESS-001` (PR #909/#910/#911, merged commit `bb80ea0f4`) established the V3R4 foundation including the 4-tier observation ladder (Observation 1× / Heuristic 3× / Rule 5× / AutoUpdate 10×, REQ-HRN-FND-011) and explicitly deferred the embedding-cluster classifier upgrade to this SPEC (`.moai/specs/SPEC-V3R4-HARNESS-001/spec.md` §1.3 Non-Goals item 2).
- `SPEC-V3R4-HARNESS-002` (PR #912/#914/#916, merged commit `dbcfcd3cf` on 2026-05-15) extended the observer surface to four event types (PostToolUse, Stop, SubagentStop, UserPromptSubmit) and added 12 `omitempty` fields to the `Event` struct (`internal/harness/types.go:73-119`). These fields — particularly `PromptPreview` (64-byte excerpt, opt-in Strategy B), `PromptLang` (ISO-639-1 heuristic), `AgentName` and `AgentType` (SubagentStop metadata) — are presently CAPTURED but NEVER CONSULTED by the classifier. This SPEC closes that design-implementation gap.
- The `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-V3R4-HARNESS-003/research.md` artifact (Phase 0 deliverable, 46935 bytes) validated four bottleneck pathologies in the current classifier and surveyed four algorithm options (SimHash, embedding, DBSCAN, hybrid). Its §9 recommendation is **Option D+A (Hybrid Two-Stage with SimHash)** on grounds of zero breaking changes, backward compatibility, deterministic output, single optional dependency, ~25ms batch cost (well within budget), and graceful failover via config flag.
- Four pathologies confirmed by research.md §1.2-1.5 via code inspection of `internal/harness/learner.go:46-101`, `internal/harness/types.go:53-119`, and `internal/harness/learner_test.go:53-74`:
  - **P1**: Three-field key includes `context_hash`, preventing cross-session aggregation. Test at `learner_test.go:64-68` validates this exact behavior (10 combos × 100 events → 10 patterns of count=100 each).
  - **P2**: Subject is a free-form string; `/moai run --team` vs `/moai run --solo` produce distinct keys despite identical intent.
  - **P3**: HARNESS-002 added 12 `omitempty` fields but `ClassifyTier(p *Pattern, thresholds []int)` consumes only the aggregated Pattern struct, not the raw Event. PromptPreview, PromptLang, AgentName, AgentType are systematically discarded before classification.
  - **P4**: `types.go:196` documents `PatternKey` as `"event_type:subject"` (two fields) but `buildPatternKey` at `learner.go:99-101` produces three (adds `context_hash`). Implementation is half-executed.

### 1.2 User-Locked Decisions

The following decisions are locked-in via research.md §9 and MUST NOT be re-litigated within this SPEC or any downstream classifier SPEC:

1. **Algorithm**: SimHash (Option A) is the Stage 2 fingerprinting algorithm. Hamming-distance threshold default is `3` of 64 bits (research.md §8.3 recommended range 2-5 for tight clustering in 64-bit fingerprint space), tunable via `learning.classifier.hamming_threshold` (looser values 5-12 acceptable for projects with weaker semantic signal).
2. **Architecture**: Hybrid two-stage (Option D). Stage 1 (existing exact-match aggregator) is preserved BYTE-IDENTICAL when Stage 2 is disabled. Stage 2 is opt-in via `learning.classifier.stage_2_enabled: false` default. Failure of Stage 2 (library load, parse error, panic) gracefully degrades to Stage-1-only mode without raising user-visible errors.
3. **Feature source**: Stage 2 fingerprint input MUST be the tuple `(subject, prompt_preview, prompt_lang, agent_name, agent_type)`. PromptContent (Strategy C opt-in full text from HARNESS-002) MUST NEVER feed into fingerprinting; the SimHash is computed only over fields that are already truncated, hashed, or non-PII at the source.
4. **Audit trail**: Every cluster merge decision MUST be recorded to `.moai/harness/cluster-merges.jsonl` with member pattern keys and pairwise Hamming distances for inspection. Stage 2 has no silent merges.
5. **Performance budget**: Batch run over 1000 events MUST complete the p99 in ≤ 25ms on Apple M-class hardware. The Observer hot path remains O(1) per event; clustering occurs only in the deferred batch invocation (AggregatePatterns + new clusterSingletons step), preserving the ≤ 5ms p99 hot-path budget from REQ-HRN-FND-010.
6. **Rate limit floor preservation**: Stage 2 MUST NOT lower the 1 Tier-4 per project per 7-day rolling window floor from `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-012. Aggregating 100 singletons into one Tier-4-eligible cluster still produces at most one AskUserQuestion per window.

### 1.3 Non-Goals

This SPEC is the CLASSIFIER UPGRADE. The following capabilities are explicitly OUT OF SCOPE and are deferred to the named downstream SPECs:

- Reflexion-style verbal self-critique loop with iteration cap — deferred to `SPEC-V3R4-HARNESS-004`.
- Principle-based scoring rubric against the design constitution — deferred to `SPEC-V3R4-HARNESS-005`.
- Multi-objective effectiveness scoring tuple and auto-rollback-on-regression — deferred to `SPEC-V3R4-HARNESS-006`.
- Voyager-style embedding-indexed skill library with top-K retrieval — deferred to `SPEC-V3R4-HARNESS-007`.
- Cross-project lesson federation with anonymization — deferred to `SPEC-V3R4-HARNESS-008`.
- TF-IDF or character n-gram fallback algorithm when SimHash is unavailable. Research.md §2 surveyed Option B variants and concluded they are not materially better than Option A; this SPEC's fallback path is "Stage-1-only" (`stage_2_enabled: false`), not a second algorithm.
- Migration tooling that re-clusters historical `tier-promotions.jsonl` entries under the new Stage-2 schema. Old promotions remain in place under their original three-field keys; new Stage-2 promotions append two-field keys to the same file.
- Online or incremental clustering at observer-record time. Stage 2 is strictly a batch step inside `AggregatePatterns`; the PostToolUse / Stop / SubagentStop / UserPromptSubmit hot paths are unmodified.
- Embedding-model inference (transformer, ONNX, Ollama). Research.md §2.2 ruled these out on grounds of ecosystem immaturity (no pure-Go transformer inference, cgo C++ dependencies unacceptable) or operational complexity (50-second batch cost via external Ollama).
- DBSCAN / HDBSCAN / k-means clustering via sklearn-go. Research.md §2.3 ruled these out on grounds of dependency footprint (sklearn-go + gonum ~10MB) and lack of semantic advantage over SimHash for lexical features.
- Modification of the `Promotion` JSONL struct schema, the `Proposal` struct, or any 5-Layer Safety pipeline input. All schemas remain unchanged; only `Pattern` aggregation behavior changes.
- Modification of `internal/harness/observer.go` or any hook code path. Stage 2 is invoked inside `AggregatePatterns`, not at record time.

---

## 2. Scope

### 2.1 In Scope

- Add Stage 2 to the learner pipeline via two new files: `internal/harness/classifier_simhash.go` (fingerprint computation, Hamming distance, fingerprint table) and `internal/harness/classifier_cluster.go` (singleton identification, pairwise comparison, cluster merge, audit log emission).
- Extend `AggregatePatterns` at `internal/harness/learner.go:46` with a Stage 2 invocation step that runs after the current Stage 1 aggregation completes. The new step reads the `Pattern` map, identifies count=1 singletons whose underlying Event still carries the HARNESS-002 extended fields (re-reading from the JSONL log to recover the source event tuple per singleton), computes SimHash fingerprints, clusters by Hamming distance ≤ threshold, and emits merged Pattern records with two-field keys (`event_type:subject_canonical`, dropping `context_hash` per research.md §2.4 decision).
- Preserve `buildPatternKey` at `learner.go:99-101` unchanged. Preserve `ClassifyTier` at `learner.go:113-140` unchanged. Preserve the `Pattern` struct at `types.go:163-187` unchanged. Preserve the `Promotion` struct at `types.go:191-209` unchanged. Preserve the `Proposal` struct at `types.go:220-244` unchanged.
- Extend `.moai/config/sections/harness.yaml` under the `learning.classifier` key with five additive fields: `stage_2_enabled` (default `false`), `similarity_algorithm` (default `simhash`, alt `none` to disable Stage 2), `hamming_threshold` (default `3`, research.md §8.3 recommended range 2-5; looser values 5-12 acceptable for weaker semantic signal), `feature_fields` (default `["subject", "prompt_preview", "prompt_lang", "agent_name", "agent_type"]`), `cluster_min_size` (default `3`). The default `stage_2_enabled: false` makes V3R4-003 a behavior-no-op until operators explicitly opt in.
- Preserve the `tier_thresholds: [1, 3, 5, 10]` config key (REQ-HRN-FND-011 contract). Stage 2 merged patterns are classified through the SAME `ClassifyTier` function with the SAME thresholds.
- Reuse the `learning.enabled` gate (REQ-HRN-FND-009) — when learning is disabled, neither Stage 1 nor Stage 2 runs.
- Emit a cluster-merge audit log to `.moai/harness/cluster-merges.jsonl` with JSONL schema: `{ts, member_keys: []string, member_counts: []int, hamming_distances: [][]int, merged_key, merged_count, confidence}`. The audit log is append-only; the file path is NOT under any FROZEN zone prefix.
- Preserve confidence-floor enforcement: clusters with mean confidence < 0.70 (the existing `confidenceThreshold` constant at `learner.go:17`) MUST classify to `TierObservation` regardless of merged count. This is the SAME confidence floor Stage 1 already applies.
- Preserve the 1 Tier-4 per project per 7-day rolling window floor (REQ-HRN-FND-012). Stage 2 MUST NOT inflate Tier-4 proposal volume past this floor even when 100+ singletons cluster into a single qualifying merged pattern.
- Add a benchmark `BenchmarkClusterSingletons1k` in a new `internal/harness/classifier_cluster_bench_test.go` that synthesizes 1000 singleton events, runs the Stage 2 path, and reports p50 / p95 / p99 wall-clock time. The p99 SHALL be ≤ 25ms (informational gate; not a CI-blocking assertion).

### 2.2 Out of Scope

The seven items listed in §1.3 (Non-Goals). In addition:

- Physical modification of the existing `Pattern` struct fields. Stage 2 emits NEW Pattern records (merged synthetic patterns); existing fields are reused as-is.
- Schema-version bump on `Promotion` or `Event`. The schema remains v1; no migration occurs.
- Modification of `.claude/rules/moai/design/constitution.md` (FROZEN file).
- Modification of `.claude/agents/moai/**`, `.claude/skills/moai-*/**`, `.claude/rules/moai/**`, or `.moai/project/brand/**` (FROZEN per REQ-HRN-FND-006).
- Modification of the `internal/harness/observer.go` write path or the four `runHarnessObserve*` CLI handlers in `internal/cli/hook.go`.
- Modification of the 5-Layer Safety pipeline at `internal/harness/safety/pipeline.go` or any of its layers (`frozen_guard.go`, `safety/canary.go`, `safety/rate_limit.go`, `safety/oversight.go`).
- Networking, telemetry upload, embedding-model inference, or any external API call.
- Any CLI verb addition or removal. The slash-command surface (`/moai:harness`) is unchanged.
- Any new subagent definition. V3R4-003 introduces no subagent; the existing `moai-harness-learner` skill body invokes the new Stage 2 path indirectly via `AggregatePatterns`.

---

## 3. Stakeholders

| Role | Interest |
|------|----------|
| MoAI-ADK maintainer | Single source of truth for classifier upgrade; uniform `learning.enabled` gate; opt-in `stage_2_enabled` flag de-risks the rollout; existing Promotion JSONL format unchanged. |
| Solo developer / power user (GOOS-style) | Tier-2+ promotion unlock for previously-stalled singleton patterns; preserved per-project no-op via `stage_2_enabled: false`; PII-safe (PromptContent never feeds the classifier). |
| Privacy-conscious user | Strategy A default from HARNESS-002 is honored; PromptContent never enters fingerprints; cluster-merges.jsonl records pattern keys only, no plaintext prompts. |
| `manager-spec` (this SPEC's author for downstream sessions) | Clean enumeration of REQ IDs and AC matrix for plan-auditor. |
| `plan-auditor` subagent | Single SPEC to audit; clear `dependencies` on HARNESS-001/002; REQ↔AC matrix complete; V3R4-001/002 contract preservation verifiable. |
| `evaluator-active` subagent | Unchanged binding-gate role; richer Tier-2 candidate stream improves classifier accuracy for downstream Tier-4 selection. |
| Downstream SPEC authors (V3R4-004 through V3R4-008) | Stage 2 merged patterns are the input stream for SPEC-V3R4-HARNESS-004 (Reflexion self-critique). The merged Pattern records carry the same `Pattern` struct shape, so downstream consumers see no schema change. |
| `manager-git` subagent | Owns the V3R4-003 plan PR squash-merge; no breaking change ID required (`breaking: false`). |

---

## 4. Exclusions (What NOT to Build)

[HARD] This SPEC explicitly EXCLUDES the following — building any of these within this PR is a scope violation:

1. Any Reflexion-style self-critique loop, episodic memory, or iteration-cap mechanism — deferred to SPEC-V3R4-HARNESS-004.
2. Any principle-based scoring rubric or constitution-parsing logic — deferred to SPEC-V3R4-HARNESS-005.
3. Any multi-objective scoring tuple or auto-rollback-on-regression mechanism — deferred to SPEC-V3R4-HARNESS-006.
4. Any embedding-indexed skill library or top-K retrieval — deferred to SPEC-V3R4-HARNESS-007.
5. Any cross-project lesson sharing, anonymization layer, or federation — deferred to SPEC-V3R4-HARNESS-008.
6. Embedding-model inference (transformer, ONNX, Ollama, or any external service). The classifier remains lexical; SimHash is the only fingerprinting algorithm.
7. DBSCAN / HDBSCAN / k-means / agglomerative clustering via sklearn-go or any external clustering library. Singleton clustering is pairwise Hamming-distance with `cluster_min_size` ≥ 3.
8. TF-IDF, character n-gram embedding, or hash-trick algorithms as fallback. The fallback path on SimHash failure is "Stage-1-only" via `stage_2_enabled: false`, not a second algorithm.
9. Migration tooling that re-aggregates historical `tier-promotions.jsonl` entries under the Stage 2 schema. Old promotions remain in place with their three-field keys.
10. Modification of `internal/harness/observer.go`, the `Observer.RecordEvent` write path, or the `Observer.RecordExtendedEvent` write path. Stage 2 is invoked downstream in `AggregatePatterns`, not at observation time.
11. Modification of any 5-Layer Safety pipeline layer (`safety/pipeline.go`, `frozen_guard.go`, `safety/canary.go`, `safety/rate_limit.go`, `safety/oversight.go`, `safety/contradiction.go`). The pipeline input contract (`Proposal` struct) is unchanged.
12. Modification of the `Promotion` JSONL struct fields, JSON tag names, or `LogSchemaVersion` constant. Schema remains v1.
13. Modification of the `tier_thresholds: [1, 3, 5, 10]` defaults. The 4-tier ladder is preserved per REQ-HRN-FND-011.
14. Modification of `.claude/rules/moai/design/constitution.md` (FROZEN file) or any path under `.claude/agents/moai/**`, `.claude/skills/moai-*/**`, `.claude/rules/moai/**`, or `.moai/project/brand/**` (FROZEN per REQ-HRN-FND-006).
15. New AskUserQuestion call sites in any subagent (preserved REQ-HRN-FND-015 / REQ-HRN-OBS-011).
16. Network call, telemetry upload, external API integration.
17. Persistence of `PromptContent` (Strategy C opt-in full prompt text from HARNESS-002 REQ-HRN-OBS-013) in any classifier-side artifact — fingerprint input, intermediate buffer, cluster audit log, or merged Pattern record. The classifier MUST consume only the truncated `PromptPreview` (64 bytes max from HARNESS-002 REQ-HRN-OBS-013 / AC-HRN-OBS-008.a).

---

## 5. Requirements (EARS format)

Each requirement is identified by `REQ-HRN-CLS-NNN` and uses one of the five EARS patterns: Ubiquitous (system shall always), Event-Driven (when X then system shall Y), State-Driven (while X the system shall Y), Optional Feature (where X exists the system shall Y), Unwanted Behavior (if X then system shall not / shall reject Y).

### REQ-HRN-CLS-001 (Ubiquitous — Stage-1 backward compatibility)

The system **shall** preserve the existing `buildPatternKey(et, subject, contextHash)` function at `internal/harness/learner.go:99` byte-identical, **shall** preserve the existing `AggregatePatterns` Stage-1 aggregation behavior (three-field exact-match key) byte-identical, and **shall** preserve the existing `ClassifyTier` function at `learner.go:113` byte-identical. When `learning.classifier.stage_2_enabled` resolves to `false` (the default), the output pattern map MUST be byte-identical to the pre-V3R4-003 baseline.

### REQ-HRN-CLS-002 (Ubiquitous — Stage-2 SimHash fingerprint generation)

The system **shall** compute a 64-bit SimHash fingerprint over the semantic feature tuple `(subject, prompt_preview, prompt_lang, agent_name, agent_type)` for each Stage-1 singleton pattern (count == 1) when Stage 2 is enabled. The fingerprint **shall** be deterministic (same input → same output every invocation), order-insensitive within the tuple, and computed via a pure-Go SimHash implementation either embedded in `internal/harness/classifier_simhash.go` or imported as a single zero-dependency library (e.g., `glaslos/go-simhash` per research.md §2.1). The fingerprint algorithm seed/version **shall** be a constant in source code; runtime modification is prohibited.

### REQ-HRN-CLS-003 (Ubiquitous — Hamming-distance singleton clustering)

The system **shall** cluster Stage-1 singletons by pairwise Hamming distance over their SimHash fingerprints. Two singletons **shall** be considered cluster-mergeable if and only if their Hamming distance is less than or equal to the configured `learning.classifier.hamming_threshold` (default `3` of 64 bits, research.md §8.3 recommended range 2-5; looser values 5-12 acceptable via config for projects with weaker semantic signal). Clusters of size below `learning.classifier.cluster_min_size` (default 3) **shall** NOT be emitted as merged patterns; they remain as the original Stage-1 singleton records.

### REQ-HRN-CLS-004 (State-Driven — Stage 2 disabled equivalence)

**While** `learning.classifier.stage_2_enabled` resolves to `false` in `.moai/config/sections/harness.yaml`, OR **while** `learning.classifier.similarity_algorithm` resolves to `none`, the classifier **shall** behave byte-identical to the pre-V3R4-003 baseline: only Stage 1 runs, no SimHash computation occurs, no audit log entry is appended to `cluster-merges.jsonl`, and the returned pattern map contains only the three-field-key Stage-1 records.

### REQ-HRN-CLS-005 (Event-Driven — cluster merge emission)

**When** Stage 2 identifies a cluster of N ≥ `cluster_min_size` singletons whose pairwise Hamming distances are all ≤ `hamming_threshold`, the system **shall** emit a single new `Pattern` record into the aggregated map containing:
- `Key`: the two-field canonical form `"{event_type}:{subject_canonical}"` (dropping `context_hash` per research.md §2.4 decision; `subject_canonical` is the cluster representative's subject string, chosen deterministically as the lexicographically-smallest member subject for reproducibility)
- `EventType`: the cluster representative's event type (all cluster members share the same `event_type` by virtue of being Stage-1 singletons whose first key field matches)
- `Subject`: the canonical subject (same as in `Key`)
- `ContextHash`: empty string `""` (Stage-2 merged patterns are session-agnostic by design)
- `Count`: the sum of all cluster members' Count fields (all 1 for singletons, so equals N)
- `Confidence`: the arithmetic mean of cluster members' Confidence values
- `Tier`: assigned via `ClassifyTier` using the same `tier_thresholds` and `confidenceThreshold` (0.70) gate from Stage 1

### REQ-HRN-CLS-006 (Ubiquitous — Tier enum preservation)

The system **shall** preserve the `Tier` enum at `internal/harness/types.go:131-145` byte-identical: the four constants `TierObservation`, `TierHeuristic`, `TierRule`, `TierAutoUpdate` **shall** retain their integer values (`iota + 1` order) and their `String()` return values (`"observation"`, `"heuristic"`, `"rule"`, `"auto_update"`). Stage 2 merged patterns **shall** consume this same enum unchanged.

### REQ-HRN-CLS-007 (Ubiquitous — Promotion JSONL schema preservation)

The system **shall** preserve the `Promotion` struct at `internal/harness/types.go:191-209` byte-identical: field names (`Ts`, `PatternKey`, `FromTier`, `ToTier`, `ObservationCount`, `Confidence`), JSON tag names (`"ts"`, `"pattern_key"`, `"from_tier"`, `"to_tier"`, `"observation_count"`, `"confidence"`), and field types **shall** remain unchanged. Stage 2 merged-pattern promotions **shall** populate `PatternKey` with the two-field canonical key from REQ-HRN-CLS-005 and **shall** populate the remaining fields per the same semantics as Stage-1 promotions. Existing pre-V3R4-003 promotions with three-field `PatternKey` values **shall** remain parseable without migration.

### REQ-HRN-CLS-008 (Ubiquitous — 5-Layer Safety pipeline input contract preservation)

The system **shall** preserve the `Proposal` struct at `internal/harness/types.go:220-244` byte-identical (field names, JSON tags, types unchanged). Stage 2 merged-pattern proposals (generated by downstream Phase 4 coordinator code outside this SPEC's scope) **shall** populate `PatternKey` with the two-field canonical key and consume the existing `safety/pipeline.go::Evaluate` entry point unchanged. The 5-Layer Safety architecture from `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-005 **shall not** be removed, bypassed, or weakened by Stage 2.

### REQ-HRN-CLS-009 (Unwanted — FROZEN zone protection non-regression)

**If** any Stage-2 merged pattern, downstream proposal generation, or audit log emission attempts to write to a path under the FROZEN prefixes `.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/rules/moai/`, or `.moai/project/brand/` (per `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-006), **then** the existing `internal/harness/frozen_guard.go::EnsureAllowed` function **shall** block the write and **shall** emit an entry to `.moai/harness/learning-history/frozen-guard-violations.jsonl`. The classifier upgrade **shall not** introduce any new path-prefix bypass.

### REQ-HRN-CLS-010 (Ubiquitous — batch performance budget)

The system **shall** complete the Stage 2 clustering step in p99 ≤ 25ms wall-clock time over a fixture of 1000 events (typical workload, no retention pruning) on Apple M-class hardware. The benchmark **shall** be `BenchmarkClusterSingletons1k` in `internal/harness/classifier_cluster_bench_test.go`. The 25ms budget is informational (not a CI-blocking assertion, mirroring the AC-HRN-OBS-013 hook-latency-budget pattern); regression beyond 25ms requires a follow-up optimization SPEC.

### REQ-HRN-CLS-011 (Ubiquitous — observer hot path preservation)

The system **shall** preserve the Observer hot path O(1) per-event cost: `internal/harness/observer.go::RecordEvent` and `Observer.RecordExtendedEvent` **shall** remain byte-identical to their post-HARNESS-002 versions. Stage 2 clustering **shall not** be invoked from any hook handler (`runHarnessObserve`, `runHarnessObserveStop`, `runHarnessObserveSubagentStop`, `runHarnessObserveUserPromptSubmit`); it **shall** run only inside the deferred batch invocation of `AggregatePatterns`.

### REQ-HRN-CLS-012 (Ubiquitous — confidence floor preservation)

The system **shall** preserve the `confidenceThreshold` constant (value `0.70`) at `internal/harness/learner.go:17`. Stage 2 merged patterns whose computed mean confidence falls below 0.70 **shall** classify to `TierObservation` regardless of merged count, identical to the Stage-1 behavior at `learner.go:115-117`. The confidence floor **shall** be checked AFTER the mean is computed and BEFORE the `ClassifyTier` count-threshold ladder is consulted.

### REQ-HRN-CLS-013 (Event-Driven — cluster-merge audit log)

**When** Stage 2 emits a merged Pattern per REQ-HRN-CLS-005, the system **shall** append exactly one JSONL line to `.moai/harness/cluster-merges.jsonl` containing:
- `ts`: ISO-8601 UTC timestamp of the merge decision
- `member_keys`: array of the source Stage-1 singleton pattern keys (three-field form)
- `member_counts`: array of the source patterns' Count values (all 1 for singletons; array length equals member count)
- `hamming_distances`: a JSON array of integers representing the upper-triangle of the N×N pairwise distance matrix in row-major order: `[d(0,1), d(0,2), …, d(0,N-1), d(1,2), …, d(N-2,N-1)]` of length N(N-1)/2.
- `hamming_pair_count`: the full pair count `N(N-1)/2` (always present, independent of truncation).
- `truncated`: boolean flag set to `true` if `hamming_distances` was truncated to its cap (see below); set to `false` or omitted otherwise.
- `merged_key`: the emitted two-field canonical key from REQ-HRN-CLS-005
- `merged_count`: the sum (equal to N)
- `confidence`: the merged Pattern's confidence value

To protect downstream JSONL parsers (Go's `bufio.Scanner` default 64KB buffer), the `hamming_distances` array **shall** be capped at the first 20 pair distances; the full pair count is stored separately as `hamming_pair_count`. If the cluster contains N where N(N-1)/2 > 20 (i.e., N ≥ 8), the array is truncated and `truncated: true` is added to the record.

The audit log file **shall** use append-only semantics (`O_APPEND|O_CREATE|O_WRONLY`, file mode `0o644`); its parent directory `.moai/harness/` **shall** be auto-created if absent. The log file path is NOT under any FROZEN prefix.

### REQ-HRN-CLS-014 (Unwanted — PII leak prevention)

**If** any Stage-2 fingerprint computation, cluster comparison, or audit log emission attempts to include the `PromptContent` field (the opt-in Strategy C full prompt text from `SPEC-V3R4-HARNESS-002` REQ-HRN-OBS-013), **then** the system **shall** reject that path and **shall not** include the field in any classifier artifact (fingerprint input, intermediate buffer, cluster audit log, or merged Pattern record). The classifier **shall** consume only the truncated `PromptPreview` (64-byte max per HARNESS-002 AC-HRN-OBS-008.a) and the other low-PII fields (`subject`, `prompt_lang`, `agent_name`, `agent_type`). The fingerprint output is one-way (SimHash is non-invertible); no plaintext PII **shall** be recoverable from any classifier artifact.

### REQ-HRN-CLS-015 (Unwanted — rate-limit floor non-regression)

**If** Stage 2 aggregates 100 or more singleton patterns into a single qualifying Tier-4 candidate cluster within a single batch invocation, **then** the system **shall not** increase the orchestrator's AskUserQuestion frequency past the existing floor of 1 Tier-4 application per project per 7-day rolling window from `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-012. The merged Pattern carries `Count = N` for a single virtual pattern, which by definition consumes at most one AskUserQuestion slot per the existing rate-limit policy in `safety/rate_limit.go`.

### REQ-HRN-CLS-016 (Optional Feature — TF-IDF / n-gram fallback deferral)

**Where** a downstream SPEC introduces a TF-IDF or character n-gram alternative fingerprinting algorithm, the system **shall** route the choice through the `learning.classifier.similarity_algorithm` config key. This SPEC ships only the `simhash` and `none` values for that key; any additional value is rejected and falls back to `simhash` (or `none` if `stage_2_enabled: false`). The TF-IDF / n-gram path is OUT OF SCOPE for V3R4-003.

### REQ-HRN-CLS-017 (Event-Driven — config override for tier_thresholds)

**When** `.moai/config/sections/harness.yaml` contains a `learning.tier_thresholds` array of exactly four positive integers in strictly-ascending order, the system **shall** use those values in place of the default `[1, 3, 5, 10]`. Stage 2 merged patterns **shall** consume the SAME `tier_thresholds` value as Stage 1; no separate threshold ladder is introduced. This requirement is a re-assertion of `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-011 with an explicit config-override surface that already exists in `learning.tier_thresholds`.

### REQ-HRN-CLS-018 (Unwanted — invalid config fails safe to Stage-1-only)

**If** `learning.classifier.hamming_threshold` is set to a value outside the range `[0, 64]`, OR `learning.classifier.cluster_min_size` is set to a value less than 2, OR `learning.classifier.similarity_algorithm` is set to a value not in the set `{simhash, none}`, **then** the system **shall** fall back to Stage-1-only behavior (equivalent to `stage_2_enabled: false`) and **shall** emit a single one-line warning to stderr (via `cmd.ErrOrStderr()` in CLI contexts, or `os.Stderr` in library contexts) identifying the invalid key and the fallback action. The fallback **shall not** raise an error to the user and **shall not** block the AggregatePatterns return path.

---

## 6. Acceptance Coverage Map

The acceptance criteria are defined in `acceptance.md` (sibling file). The coverage map below shows every REQ from §5 mapped to at least one AC. Full Given-When-Then scenarios are in `acceptance.md`.

| AC ID | Covers REQ IDs |
|-------|----------------|
| AC-HRN-CLS-001 | REQ-HRN-CLS-001, REQ-HRN-CLS-004 |
| AC-HRN-CLS-002 | REQ-HRN-CLS-002, REQ-HRN-CLS-003 |
| AC-HRN-CLS-003 | REQ-HRN-CLS-003 |
| AC-HRN-CLS-004 | REQ-HRN-CLS-005, REQ-HRN-CLS-006 |
| AC-HRN-CLS-005 | REQ-HRN-CLS-010 |
| AC-HRN-CLS-006 | REQ-HRN-CLS-009 |
| AC-HRN-CLS-007 | REQ-HRN-CLS-012 |
| AC-HRN-CLS-008 | REQ-HRN-CLS-013 |
| AC-HRN-CLS-009 | REQ-HRN-CLS-014 |
| AC-HRN-CLS-010 | REQ-HRN-CLS-015 |
| AC-HRN-CLS-011 | REQ-HRN-CLS-007, REQ-HRN-CLS-008 |
| AC-HRN-CLS-012 | REQ-HRN-CLS-011 |
| AC-HRN-CLS-013 | REQ-HRN-CLS-017 |
| AC-HRN-CLS-014 | REQ-HRN-CLS-018, REQ-HRN-CLS-016 |

Coverage: 18 REQs ↔ 14 ACs; every REQ appears in at least one AC.

---

## 7. Constraints

[HARD] Language: All SPEC artifact content body in English where possible. Conversation-language Korean is acceptable per `.moai/config/sections/language.yaml` `conversation_language: ko`. EARS keywords (WHEN / WHILE / WHERE / IF / SHALL) remain English. Code identifiers (function names, file paths, REQ IDs, AC IDs, struct field names) remain English.

[HARD] FROZEN zone preservation: `.claude/rules/moai/design/constitution.md` §2 and §5 are NOT modified by this SPEC. The V3R4-001 contracts (REQ-HRN-FND-005, REQ-HRN-FND-006, REQ-HRN-FND-009, REQ-HRN-FND-010, REQ-HRN-FND-011, REQ-HRN-FND-012, REQ-HRN-FND-015) and the V3R4-002 contracts (REQ-HRN-OBS-007, REQ-HRN-OBS-009, REQ-HRN-OBS-010, REQ-HRN-OBS-011, REQ-HRN-OBS-012, REQ-HRN-OBS-014) are referenced verbatim in §1.2 and §5 and **shall not** be weakened.

[HARD] EARS format mandatory for all REQs in §5.

[HARD] No tech-stack implementation assumptions in this spec.md. Requirements describe contracts. Implementation decisions (exact algorithm parameters, library version pinning, file-internal data structures) belong to plan.md.

[HARD] Conventional Commits format for the plan PR commit message: `plan(SPEC-V3R4-HARNESS-003): embedding-cluster classifier plan`.

[HARD] No emojis in user-facing output (per `.claude/rules/moai/development/coding-standards.md` § Content Restrictions).

[HARD] No time estimates or duration predictions in any artifact (per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation). Use priority labels (P0 / P1 / P2 / P3) and phase ordering.

[HARD] Schema additivity rule: all new fields in any struct touched by this SPEC are tagged `omitempty` if added. The four protected schemas (`Pattern`, `Promotion`, `Proposal`, `Event`) are unchanged.

[HARD] Determinism rule: SimHash seed/version constants are compile-time; runtime modification is prohibited. Cluster representative subject selection is lexicographically-smallest (deterministic tie-break). Audit log timestamps use `time.Now().UTC()` and ISO-8601 format identical to existing `Observer` semantics.

[HARD] PII rule: `PromptContent` (opt-in full text) MUST NEVER enter any classifier artifact. Static check: `grep -nE 'evt\.PromptContent|event\.PromptContent|Pattern\.PromptContent' internal/harness/classifier_*.go` MUST return zero matches.

---

## 8. Risks

| Risk | Likelihood | Severity | Mitigation |
|------|------------|----------|------------|
| Stage-2 false-positive clusters merge semantically distinct patterns (e.g., "find files" vs "search code") | Medium | Medium | `cluster_min_size ≥ 3` (default) requires three-way agreement, reducing false positives. Default `hamming_threshold: 3` matches research.md §8.3 recommended range (2-5 in 64-bit space); tunable to looser values (e.g., 5-12) via `harness.yaml` for projects with weaker semantic signal. Audit log enables post-hoc review and config tightening. |
| SimHash library (`glaslos/go-simhash`) becomes unmaintained or has security advisories | Low | Medium | Library is 1KB pure Go with stable API since 2024-01 (research.md §8.1). Alternative: embed a 50-100 line SimHash implementation directly in `classifier_simhash.go` (research.md §2.1 notes this is feasible). REQ-HRN-CLS-018 fail-open path also disables Stage 2 on any library load error. |
| Hamming threshold tuning (default 3 of 64 bits) proves too aggressive or too conservative in real-world usage | Medium | Low | Threshold is config-tunable (`learning.classifier.hamming_threshold`). Pilot in single project for 2-4 weeks before flipping default `stage_2_enabled: true`. Research.md §8.2 documents this risk. |
| Performance regression beyond 25ms p99 budget under unusual workloads (e.g., 10k singletons in single batch) | Low | Medium | REQ-HRN-CLS-010 sets 25ms p99 over 1000 events as informational gate. Pairwise Hamming is O(k²) where k=unique singletons (typically ≤ 50 per batch, research.md §3.3). For 10k singletons the cost rises to ~2.5s but this would be a pathological case; mitigation is to skip clustering when `len(singletons) > MAX_CLUSTER_INPUT` (config-tunable). Out of scope for v0.1.0; deferred to follow-up SPEC if observed. |
| Schema drift if a future SPEC adds new fields to `Pattern` without updating `classifier_cluster.go` merge logic | Medium | Medium | The merged Pattern emission in REQ-HRN-CLS-005 explicitly enumerates all seven Pattern fields (`Key`, `EventType`, `Subject`, `ContextHash`, `Count`, `Confidence`, `Tier`). Future field additions MUST update this enumeration. Plan-auditor for any classifier-related SPEC will catch drift. |
| PII leak via accidental `PromptContent` inclusion in fingerprint or audit log | Low | High | REQ-HRN-CLS-014 forbids `PromptContent` use. Static grep guard in CI: `grep -nE 'PromptContent' internal/harness/classifier_*.go` MUST return zero matches. Code review by plan-auditor. |
| Test fragility: Stage-1 backward-compatibility test (AC-HRN-CLS-001) fails after unrelated refactor | Low | Medium | The test asserts byte-identical pattern map vs HARNESS-002 baseline using a golden fixture file (sibling `testdata/stage1_baseline.json`). Drift detection is automatic; the fix is to regenerate the golden file only if Stage-1 behavior change is intentional and documented in a separate SPEC. |

---

## 9. Dependencies

| SPEC | Relationship | Notes |
|------|--------------|-------|
| `SPEC-V3R4-HARNESS-001` | Hard dependency (foundation) | Establishes the 4-tier ladder (REQ-HRN-FND-011) preserved verbatim, the `learning.enabled` gate (REQ-HRN-FND-009) reused, the 5-Layer Safety (REQ-HRN-FND-005) preserved, the FROZEN zone (REQ-HRN-FND-006) preserved, the 7-day Tier-4 rate-limit floor (REQ-HRN-FND-012) preserved, and the orchestrator-only AskUserQuestion contract (REQ-HRN-FND-015) preserved. |
| `SPEC-V3R4-HARNESS-002` | Hard dependency (observation surface) | Delivers the 12 `omitempty` Event fields (PromptPreview, PromptLang, AgentName, AgentType, etc.) that Stage 2 consumes as fingerprint input. Stage 2 specifically reads `PromptPreview` (64-byte max per AC-HRN-OBS-008.a) and `PromptLang` (ISO-639-1 heuristic per REQ-HRN-OBS-006); `PromptContent` is forbidden. |
| `SPEC-V3R4-HARNESS-004` | Downstream (blocked by this SPEC) | Reflexion self-critique consumes Stage 2 merged patterns as a richer Tier-2+ candidate stream. |
| `SPEC-V3R4-HARNESS-005` through `SPEC-V3R4-HARNESS-008` | Downstream (indirectly blocked) | All build on the classifier upgrade established by this SPEC. |

---

## 10. Glossary

- **SimHash**: Locality-sensitive hashing algorithm (Manku, Jain, Sarpatwar, VLDB 2007). Produces a fixed-width fingerprint (64 bits in this SPEC) such that similar inputs produce fingerprints with low Hamming distance. Deterministic, order-insensitive within the input token set, one-way (non-invertible).
- **Hamming distance**: The number of bit positions at which two fixed-width binary strings differ. For 64-bit SimHash fingerprints, ranges from 0 (identical) to 64 (every bit differs).
- **Singleton cluster**: A Stage-1 pattern with `Count == 1`, meaning the underlying (event_type, subject, context_hash) tuple appeared exactly once in the observation log. Singletons are Stage 2's input domain.
- **Stage 1 / Stage 2**: Stage 1 is the existing exact-match aggregator (`buildPatternKey` three-field key, preserved verbatim by REQ-HRN-CLS-001). Stage 2 is the new SimHash-based singleton clustering layer (REQ-HRN-CLS-002, -003, -005) that runs after Stage 1 inside `AggregatePatterns`.
- **Merged pattern**: A synthetic `Pattern` record emitted by Stage 2 representing a cluster of N ≥ `cluster_min_size` singletons. Its key uses the two-field canonical form `"{event_type}:{subject_canonical}"` (dropping context_hash per research.md §2.4 decision).
- **Confidence floor**: The 0.70 threshold (`confidenceThreshold` constant at `learner.go:17`) below which any pattern (Stage 1 or Stage 2) classifies to `TierObservation` regardless of count.
- **Two-field key vs three-field key**: The current implementation produces `"event_type:subject:context_hash"` (three fields); Stage 2 emits `"event_type:subject"` (two fields, session-agnostic). The documented spec at `types.go:196` already specifies two fields; this SPEC closes the design-implementation gap (research.md §1.5 pathology P4).
- **Stage-1-only mode**: Behavior when `stage_2_enabled: false` (default) OR `similarity_algorithm: none` OR invalid config (REQ-HRN-CLS-018 fail-safe). Output is byte-identical to pre-V3R4-003 baseline.

---

## 11. References

### Code Files (absolute paths, with line ranges)

| File | Lines | Purpose |
|------|-------|---------|
| `internal/harness/learner.go` | 1-177 | Current Stage-1 classifier; `buildPatternKey:99`, `ClassifyTier:113`, `AggregatePatterns:46`, `WritePromotion:142`, `confidenceThreshold:17`. |
| `internal/harness/types.go` | 1-387 | `Event` struct with HARNESS-002 omitempty fields (lines 73-119), `Pattern` (163-187), `Tier` enum (131-145), `Promotion` (191-209), `Proposal` (220-244), `LogSchemaVersion:11`. |
| `internal/harness/observer.go` | 1-196 | Observer hot path (preserved unchanged); `RecordEvent:53`, `RecordExtendedEvent:103`. |
| `internal/harness/safety/pipeline.go` | 1-159 | 5-Layer Safety pipeline (preserved unchanged); `Evaluate:89` entry point. |
| `internal/harness/frozen_guard.go` | 1-113 | L1 Frozen Guard (preserved unchanged); `IsAllowedPath:60`, `EnsureAllowed:103`, `allowedPrefixes:18`, `frozenPrefixes:27`. |
| `internal/harness/learner_test.go` | 1-445 | Existing Stage-1 tests; line 64-68 asserts 10 distinct combos → 10 patterns (the test that locks in three-field key behavior, used as Stage-1 backward-compat baseline). |
| `.moai/config/sections/harness.yaml` | 115-127 | `learning` block with `auto_apply`, `enabled`, `log_retention_days`, `rate_limit`, `tier_thresholds: [1, 3, 5, 10]`. Stage 2 extends this with `learning.classifier` sub-block. |

### Sibling SPEC Artifacts

- `.moai/specs/SPEC-V3R4-HARNESS-003/research.md` (this SPEC's Phase 0 research deliverable, 46935 bytes; §0 Executive Summary, §1.2-1.5 bottleneck pathologies P1-P4, §2 algorithm options A/B/C/D, §3 performance budget, §4 PII risk, §5 backward-compat surface, §6 integration points, §7 config strategy, §8 risks, §9 recommendation Option D+A, §10 references).
- `.moai/specs/SPEC-V3R4-HARNESS-003/plan.md` (this SPEC's implementation plan).
- `.moai/specs/SPEC-V3R4-HARNESS-003/acceptance.md` (this SPEC's AC enumeration).
- `.moai/specs/SPEC-V3R4-HARNESS-003/tasks.md` (this SPEC's task breakdown).
- `.moai/specs/SPEC-V3R4-HARNESS-003/spec-compact.md` (auto-generated compact form for downstream consumers).

### Upstream SPECs (binding contracts)

- `.moai/specs/SPEC-V3R4-HARNESS-001/spec.md` REQ-HRN-FND-005 (5-Layer Safety preservation), REQ-HRN-FND-006 (FROZEN zone path-prefix protection), REQ-HRN-FND-009 (`learning.enabled` gate), REQ-HRN-FND-010 (PostToolUse baseline JSONL schema), REQ-HRN-FND-011 (4-tier ladder thresholds 1/3/5/10), REQ-HRN-FND-012 (Tier-4 rate-limit 1/week), REQ-HRN-FND-015 (subagent AskUserQuestion prohibition).
- `.moai/specs/SPEC-V3R4-HARNESS-002/spec.md` REQ-HRN-OBS-006 (UserPromptSubmit Strategy A default with `prompt_hash`/`prompt_len`/`prompt_lang`), REQ-HRN-OBS-009 (schema additivity preserved), REQ-HRN-OBS-012 (PII privacy Strategy A default), REQ-HRN-OBS-013 (opt-in Strategy B preview / C content / none), REQ-HRN-OBS-014 (invalid PII config fails open to Strategy A).
- `.moai/specs/SPEC-V3R4-HARNESS-002/acceptance.md` AC-HRN-OBS-006 (schema additivity baseline test pattern), AC-HRN-OBS-007 (Strategy A default PII test pattern), AC-HRN-OBS-013 (latency budget pattern).

### External References (from research.md §10)

- **SimHash**: Manku, Jain, Sarpatwar, VLDB 2007 — Fingerprinting for near-duplicate detection.
- **glaslos/go-simhash**: github.com/glaslos/go-simhash — MIT license, pure Go, last commit 2024-01, stable API.
- **Reflexion**: Li et al., arXiv:2303.11366 — Self-critique pattern (downstream SPEC-V3R4-HARNESS-004).
- **Voyager**: Wang et al., arXiv:2305.16291 — Skill library auto-organization (downstream SPEC-V3R4-HARNESS-007).
- **Constitutional AI**: Bai et al., arXiv:2212.08073 — Principle-based scoring (downstream SPEC-V3R4-HARNESS-005).

### Design Constitution (FROZEN — referenced, not modified)

- `.claude/rules/moai/design/constitution.md` §2 (Frozen vs Evolvable Zones), §5 (Safety Architecture — 5 Layers).
- `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary.
- `.claude/rules/moai/core/askuser-protocol.md` — Canonical AskUserQuestion protocol.

---

End of spec.md.
