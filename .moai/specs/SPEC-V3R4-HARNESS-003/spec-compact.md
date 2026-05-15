# SPEC-V3R4-HARNESS-003 — Compact Form

Auto-generated compact form of `spec.md` for downstream consumers (plan-auditor, evaluator-active, downstream SPEC authors). EARS REQs + Given/When/Then AC text + files-to-modify + Exclusions only. Full context in `spec.md`.

---

## Requirements (EARS, REQ-HRN-CLS-NNN)

### REQ-HRN-CLS-001 (Ubiquitous)
The system **shall** preserve `buildPatternKey` at `learner.go:99`, `AggregatePatterns` Stage-1 aggregation, and `ClassifyTier` at `learner.go:113` byte-identical. With `stage_2_enabled: false` (default), output pattern map MUST be byte-identical to pre-V3R4-003 baseline.

### REQ-HRN-CLS-002 (Ubiquitous)
The system **shall** compute a 64-bit SimHash fingerprint over `(subject, prompt_preview, prompt_lang, agent_name, agent_type)` for each Stage-1 singleton when Stage 2 enabled. Fingerprint MUST be deterministic, order-insensitive, pure-Go. Algorithm seed/version is compile-time constant.

### REQ-HRN-CLS-003 (Ubiquitous)
The system **shall** cluster singletons by pairwise Hamming distance ≤ `hamming_threshold` (default `3` of 64 bits, research.md §8.3 recommended range 2-5; looser values 5-12 acceptable via config). Clusters of size < `cluster_min_size` (default 3) **shall not** be emitted.

### REQ-HRN-CLS-004 (State-Driven)
**While** `stage_2_enabled: false` OR `similarity_algorithm: none`, classifier **shall** behave byte-identical to baseline: only Stage 1 runs, no SimHash, no audit log, no merged records.

### REQ-HRN-CLS-005 (Event-Driven)
**When** Stage 2 identifies a cluster of N ≥ `cluster_min_size` with pairwise Hamming all ≤ threshold, the system **shall** emit one merged Pattern: Key=`"{event_type}:{lex-min-subject}"` (2-field, dropping context_hash), EventType=shared, Subject=lex-min, ContextHash=`""`, Count=sum, Confidence=mean, Tier via `ClassifyTier`.

### REQ-HRN-CLS-006 (Ubiquitous)
The system **shall** preserve `Tier` enum at `types.go:131-145` byte-identical: 4 constants, `iota + 1` values, `String()` returns unchanged.

### REQ-HRN-CLS-007 (Ubiquitous)
The system **shall** preserve `Promotion` struct at `types.go:191-209` byte-identical (field names, JSON tags, types). Pre-V3R4-003 entries with 3-field PatternKey remain parseable; no migration.

### REQ-HRN-CLS-008 (Ubiquitous)
The system **shall** preserve `Proposal` struct at `types.go:220-244` byte-identical. 5-Layer Safety from REQ-HRN-FND-005 **shall not** be removed, bypassed, or weakened.

### REQ-HRN-CLS-009 (Unwanted)
**If** Stage-2 emission attempts a write under FROZEN prefixes per REQ-HRN-FND-006, **then** `frozen_guard.go::EnsureAllowed` **shall** block. V3R4-003 introduces no new bypass.

### REQ-HRN-CLS-010 (Ubiquitous)
The system **shall** complete Stage 2 in p99 ≤ 25ms over 1000 events on Apple M-class hardware. Benchmark `BenchmarkClusterSingletons1k`. Informational gate (not CI-blocking).

### REQ-HRN-CLS-011 (Ubiquitous)
The system **shall** preserve Observer hot path O(1): `RecordEvent`/`RecordExtendedEvent` byte-identical. Stage 2 **shall not** be invoked from any hook handler; runs only in `AggregatePatterns`.

### REQ-HRN-CLS-012 (Ubiquitous)
The system **shall** preserve `confidenceThreshold = 0.70` at `learner.go:17`. Stage 2 clusters with mean confidence < 0.70 **shall** classify to `TierObservation` regardless of count.

### REQ-HRN-CLS-013 (Event-Driven)
**When** Stage 2 emits a merge, system **shall** append one JSONL line to `.moai/harness/cluster-merges.jsonl`: `{ts, member_keys, member_counts, hamming_distances, hamming_pair_count, truncated, merged_key, merged_count, confidence}`. `hamming_distances` is a JSON array of ints representing the upper-triangle of the N×N pairwise distance matrix in row-major order `[d(0,1), d(0,2), …, d(N-2,N-1)]` of length N(N-1)/2, capped at 20 elements (sets `truncated: true` and stores full pair count as `hamming_pair_count`) to protect Go `bufio.Scanner` 64KB default buffer. Append-only (`O_APPEND|O_CREATE|O_WRONLY`, 0o644). Parent dir auto-created.

### REQ-HRN-CLS-014 (Unwanted)
**If** Stage-2 fingerprint / cluster comparison / audit log attempts to include `PromptContent`, **then** system **shall** reject. Classifier consumes only `PromptPreview` (64-byte max), `subject`, `prompt_lang`, `agent_name`, `agent_type`. SimHash is one-way; no PII recoverable.

### REQ-HRN-CLS-015 (Unwanted)
**If** Stage 2 aggregates 100+ singletons into one Tier-4 cluster within a batch, **then** system **shall not** increase AskUserQuestion frequency past 1/project/7-day-window floor from REQ-HRN-FND-012.

### REQ-HRN-CLS-016 (Optional Feature)
**Where** a downstream SPEC introduces TF-IDF/n-gram fingerprinting, choice routes through `learning.classifier.similarity_algorithm`. This SPEC ships only `simhash` and `none`; unknown values fall back to `simhash` (or `none` if `stage_2_enabled: false`).

### REQ-HRN-CLS-017 (Event-Driven)
**When** `learning.tier_thresholds` is set to 4 ascending positive ints, system **shall** use those values. Stage 2 consumes SAME `tier_thresholds` as Stage 1. Re-asserts REQ-HRN-FND-011.

### REQ-HRN-CLS-018 (Unwanted)
**If** `hamming_threshold` outside `[0, 64]`, OR `cluster_min_size < 2`, OR `similarity_algorithm` not in `{simhash, none}`, **then** system **shall** fall back to Stage-1-only AND emit one-line stderr warning. Fallback **shall not** raise error.

---

## Acceptance Criteria (Given-When-Then, AC-HRN-CLS-NNN)

### AC-HRN-CLS-001 — Stage-1 Backward Compatibility (Stage 2 OFF)
**Given** `stage_2_enabled: false` + fixture 1000 events (10 combos × 100). **When** `AggregatePatterns` runs. **Then** map byte-identical to golden `stage1_baseline_patterns.json`: 10 entries, 3-field keys, Count=100 each, Confidence=1.0. No Stage-2 records. `cluster-merges.jsonl` absent.

### AC-HRN-CLS-002 — Stage 2 ON: 10-Singleton Aggregation
**Given** `stage_2_enabled: true`, threshold=3, min_size=3, 10 `user_prompt` singletons sharing IDENTICAL 64-byte `prompt_preview` content across all events (realistic Tier-2 scenario; subjects vary only by SPEC ID suffix). **When** runs. **Then** map has EXACTLY 11 entries: 10 Stage-1 singletons (unchanged, Count=1 each) AND 1 Stage-2 merged Pattern (synthetic 2-field key, Count=10, Confidence=1.0, Tier=TierAutoUpdate). Pairwise Hamming ≤ 3 across all `C(10,2) = 45` cluster pairs (asserted via direct SimHash).

### AC-HRN-CLS-003 — Stage 2 ON: Dissimilar Singletons Stay Separate
**Given** Stage-2 on, threshold=3, 10 divergent singletons (pairwise Hamming > 3). **When** runs. **Then** map has exactly 10 three-field-key entries; NO merge. No audit log entry.

### AC-HRN-CLS-004 — Cluster Merge Emission Shape
**Given** 5-singleton cluster, Confidences `[1.0, 1.0, 0.9, 0.8, 1.0]`. **When** merge emits. **Then** Pattern: Key=2-field canonical (lex-min subject), Count=5, Confidence=0.94, Tier=TierRule. Enum unchanged.

### AC-HRN-CLS-005 — Performance: p99 ≤ 25ms over 1000 Events
**Given** V3R4-003 compiled, M-class hardware, fixture 1000 events. **When** `BenchmarkClusterSingletons1k` runs. **Then** p99 ≤ 25ms across 100 invocations. Informational.

### AC-HRN-CLS-006 — FROZEN Zone Non-Regression
**Given** V3R4-003 merged, hypothetical coordinator maps merge to FROZEN path. **When** `EnsureAllowed(".claude/skills/moai-meta-harness/SKILL.md")` invoked. **Then** returns `*FrozenViolationError`. `allowedPrefixes` / `frozenPrefixes` unchanged.

### AC-HRN-CLS-007 — Confidence Floor Forces TierObservation
**Given** 10-singleton cluster, mean Confidence 0.69. **When** merge + ClassifyTier. **Then** Confidence=0.69, Count=10, Tier=TierObservation (short-circuit at learner.go:115-117).

### AC-HRN-CLS-008 — Cluster-Merge Audit Log Schema
**Given** Stage-2 emits one merge for 4-member cluster. **When** audit log writes. **Then** `cluster-merges.jsonl` has exactly 1 JSONL line with: `ts` ISO-8601, `member_keys` (4 strings, 3-field), `member_counts` (4 ints =1), `hamming_distances` (6 ints, flat upper-triangle row-major `[d(0,1), d(0,2), d(0,3), d(1,2), d(1,3), d(2,3)]`, length `C(4,2) = 6`), `hamming_pair_count` (6), `merged_key` 2-field, `merged_count`=4, `confidence` float. For N ≥ 8 clusters where N(N-1)/2 > 20, `hamming_distances` capped at first 20 elements with `truncated: true`. Append-only across runs. EDGE-005 (N=100): `len(member_keys)==100`, `len(hamming_distances)==20`, `truncated==true`, `hamming_pair_count==4950`.

### AC-HRN-CLS-009 — PII: PromptContent Never Enters Classifier
**Given** user opted Strategy C, PromptContent="my-secret-api-key-DO-NOT-LEAK-12345". **When** Stage 2 processes. **Then** sentinel NOT in `cluster-merges.jsonl`, NOT in merged Pattern fields. Static guard: `grep -nE 'PromptContent' internal/harness/classifier_*.go` returns 0.

### AC-HRN-CLS-010 — Rate-Limit Non-Regression
**Given** Stage-2 produces single Tier-4 merge from 100 singletons. **When** orchestrator processes downstream. **Then** ≤ 1 AskUserQuestion per 7-day window per REQ-HRN-FND-012. Count=100 does NOT inflate rate-limit grants.

### AC-HRN-CLS-011 — Promotion + Proposal Schema Regression-Proof
**Given** pre-V3R4-003 `tier-promotions.jsonl` with 3 entries (3-field pattern_keys). **When** test reads via `json.Unmarshal`. **Then** all parse; new Stage-2 entry with 2-field key also parses; both coexist parseably. `Proposal` struct unchanged; `safety/pipeline.go::Evaluate` signature unchanged.

### AC-HRN-CLS-012 — Observer Hot Path Preserved
**Given** V3R4-003 merged, 1000 hook invocations. **When** `BenchmarkObserverRecordEvent` runs. **Then** per-event time within pre-V3R4-003 baseline. Static: `grep -nE 'classifier_simhash|classifier_cluster' internal/cli/hook.go internal/harness/observer.go` returns 0.

### AC-HRN-CLS-013 — tier_thresholds Config Override
**Given** `tier_thresholds: [2, 4, 8, 20]`, Stage-2 merge Count=8. **When** ClassifyTier. **Then** Tier=TierRule (Count=8 ≥ thresholds[2]=8 AND < thresholds[3]=20). Same ClassifyTier for both stages.

### AC-HRN-CLS-014 — Invalid Config Fails Safe to Stage-1-Only
**Given** ONE of 6 sub-cases: (1) `hamming_threshold: -5` | (2) `99` | (3) `cluster_min_size: 1` | (4) `similarity_algorithm: "tfidf"` | (5) `"garbage_value_xyz"` | (6) yaml type mismatch `hamming_threshold: '3.5'` (string) or `3.5` (float) — loader catches `yaml.Unmarshal` type error BEFORE `Validate()`, logs WARN, falls back to defaults (`hamming_threshold: 3`). **When** `AggregatePatterns` invoked with merge-eligible fixture. **Then** pattern map byte-identical to Stage-1 baseline; no merges; no audit entry; one-line stderr warning; function returns `(patterns, nil)`.

---

## Files to Modify

| File | Op | Wave |
|------|----|------|
| `internal/harness/classifier_simhash.go` | NEW | B |
| `internal/harness/classifier_simhash_test.go` | NEW | B |
| `internal/harness/classifier_cluster.go` | NEW | C |
| `internal/harness/classifier_cluster_test.go` | NEW | C |
| `internal/harness/classifier_cluster_audit_test.go` | NEW | C |
| `internal/harness/classifier_cluster_bench_test.go` | NEW | D |
| `internal/harness/classifier_pii_test.go` | NEW | D |
| `internal/harness/classifier_schema_regression_test.go` | NEW | D |
| `internal/harness/classifier_frozen_guard_regression_test.go` | NEW | D |
| `internal/harness/classifier_rate_limit_test.go` | NEW | D |
| `internal/harness/learner.go` | Mod (additive Stage-2 call site at end of AggregatePatterns) | A |
| `internal/harness/learner_test.go` | Mod (golden baseline test) | A |
| `internal/harness/testdata/stage1_baseline.jsonl` | NEW | A |
| `internal/harness/testdata/stage1_baseline_patterns.json` | NEW | A |
| `internal/harness/testdata/stage2_similar_10.jsonl` | NEW | C |
| `internal/harness/testdata/stage2_dissimilar_10.jsonl` | NEW | C |
| `internal/harness/testdata/stage2_perf_1k.jsonl` | NEW | D |
| `internal/harness/testdata/legacy_promotions.jsonl` | NEW | D |
| `.moai/config/sections/harness.yaml` | Mod (`learning.classifier` block) | D |
| `internal/config/loader.go` | Mod (parse + validate) | D |
| `CHANGELOG.md` | Mod | D |
| `.moai/release/RELEASE-NOTES-v2.22.0.md` | NEW | D |

**Files explicitly NOT modified**:
- `internal/harness/observer.go` (hot path preserved per REQ-HRN-CLS-011)
- `internal/harness/types.go` (Pattern, Promotion, Proposal, Tier, EventType unchanged)
- `internal/harness/frozen_guard.go` (per REQ-HRN-CLS-009)
- `internal/harness/safety/pipeline.go` + all 5-Layer Safety layers (per REQ-HRN-CLS-008)
- `internal/cli/hook.go` (per REQ-HRN-CLS-011)
- `.claude/rules/moai/design/constitution.md` (FROZEN)
- `.claude/rules/moai/core/agent-common-protocol.md` (FROZEN)
- `.claude/rules/moai/core/askuser-protocol.md` (FROZEN)
- Any path under `.claude/agents/moai/**`, `.claude/skills/moai-*/**`, `.claude/rules/moai/**`, `.moai/project/brand/**` (FROZEN)

---

## Exclusions (What NOT to Build)

[HARD] Scope violations:

1. Reflexion self-critique → SPEC-V3R4-HARNESS-004.
2. Principle-based scoring → SPEC-V3R4-HARNESS-005.
3. Multi-objective scoring / auto-rollback → SPEC-V3R4-HARNESS-006.
4. Voyager skill library / embedding-indexed retrieval → SPEC-V3R4-HARNESS-007.
5. Cross-project federation → SPEC-V3R4-HARNESS-008.
6. Embedding-model inference (transformer, ONNX, Ollama). SimHash only.
7. DBSCAN / HDBSCAN / k-means / agglomerative clustering via sklearn-go.
8. TF-IDF / character n-gram fallback. Fail-safe = Stage-1-only.
9. Migration tooling for historical `tier-promotions.jsonl`.
10. Modification of `Observer.RecordEvent` / `RecordExtendedEvent` write paths.
11. Modification of any 5-Layer Safety pipeline layer.
12. Modification of `Promotion` JSONL struct or `LogSchemaVersion`.
13. Modification of `tier_thresholds: [1, 3, 5, 10]` defaults.
14. Modification of FROZEN files (constitution, agent-common-protocol, askuser-protocol).
15. New AskUserQuestion call sites in any subagent.
16. Network call, telemetry, external API integration.
17. Persistence of `PromptContent` in any classifier artifact (fingerprint, audit log, merged Pattern). Classifier consumes only `PromptPreview` (64-byte max).

---

End of spec-compact.md.
