# Implementation Plan — SPEC-V3R4-HARNESS-003

This document is the Wave-level implementation plan for the Embedding-Cluster Classifier (Tier-2 Pattern Aggregation Upgrade) SPEC. All priorities use P0/P1/P2/P3 labels per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation; no time estimates are used.

---

## 1. Overview

This SPEC closes the design-implementation gap between the rich multi-event observation surface delivered by `SPEC-V3R4-HARNESS-002` (12 omitempty fields on the `Event` struct: PromptPreview, PromptLang, AgentName, AgentType, plus 8 more) and the exact-match-only Stage-1 pattern aggregator at `internal/harness/learner.go:46-101`. Per the research.md §9 recommendation (Option D+A Hybrid Two-Stage with SimHash), V3R4-003 adds a Stage 2 layer that fingerprints Stage-1 singleton patterns over their semantic feature tuple, clusters them by Hamming distance ≤ threshold, and emits merged Pattern records eligible for higher-tier promotion. Stage 1 is preserved byte-identical; the entire upgrade is opt-in via `learning.classifier.stage_2_enabled: false` default. The 5-Layer Safety pipeline, FROZEN zone protection, the 7-day Tier-4 rate-limit floor, the orchestrator-only AskUserQuestion contract, and the 4-tier ladder thresholds 1/3/5/10 are all preserved unchanged from `SPEC-V3R4-HARNESS-001`. The 64-byte PromptPreview cap, opt-in Strategy A/B/C, and PII fail-safe semantics from `SPEC-V3R4-HARNESS-002` are honored; `PromptContent` (Strategy C full text) is forbidden from any classifier artifact (REQ-HRN-CLS-014).

The plan decomposes into four Waves (A through D). Wave A refactors the existing Stage-1 path to expose an aggregator-injection seam without changing behavior. Wave B introduces the new SimHash module. Wave C adds the Hamming-distance singleton clustering layer and audit log. Wave D wires the config extension and end-to-end test. Each Wave is independently PR-able; this plan-phase PR (manager-spec) only ships the five SPEC artifacts.

| Wave | Title | Priority | Owning Run-Phase Agent | Acceptance Gate |
|------|-------|----------|------------------------|-----------------|
| Wave A | Stage-1 refactor (aggregator interface seam) | P0 | `manager-tdd` (or `manager-develop` depending on quality.yaml) | AC-HRN-CLS-001 passes — Stage 1 byte-identical with stage_2_enabled=false. Golden baseline test in place. |
| Wave B | Stage-2 SimHash module | P0 | `manager-tdd` (delegating to `expert-backend`) | AC-HRN-CLS-002 fingerprint test passes; table-driven SimHash unit tests cover determinism, order-insensitivity, range. |
| Wave C | Stage-2 clustering + audit log | P0 | `manager-tdd` | AC-HRN-CLS-002/003/004/007/008 pass — clustering produces correct merged records; audit log schema matches REQ-HRN-CLS-013. |
| Wave D | Config wiring + end-to-end test + performance budget | P1 | `manager-tdd` (delegating to `expert-devops` for config) | AC-HRN-CLS-005 (perf), AC-HRN-CLS-013 (config override), AC-HRN-CLS-014 (fail-safe), AC-HRN-CLS-009 (PII grep guard) pass. |

All four Waves are documented here as the run-phase roadmap. This `/moai plan` invocation produces the SPEC artifacts only; run-phase work happens in `/moai run SPEC-V3R4-HARNESS-003` after this plan PR merges.

---

## 2. Architecture Overview

### 2.1 Current State (pre-V3R4-003)

```
+-----------------------------------------------------------------+
| usage-log.jsonl (Observer writes events; HARNESS-001 schema     |
| + HARNESS-002 12 omitempty fields)                              |
+--------------------------+--------------------------------------+
                           |
                           v
+----------------------------------------------------------------+
| AggregatePatterns (learner.go:46)                              |
| - Reads JSONL line-by-line                                      |
| - For each Event: key = buildPatternKey(et, subject, ctxHash)   |
| - Increments Count if key exists; else creates Pattern          |
| - Returns map[string]*Pattern (three-field keys)                |
+--------------------------+--------------------------------------+
                           |
                           v
+----------------------------------------------------------------+
| ClassifyTier (learner.go:113)                                  |
| - For each pattern: applies thresholds [1, 3, 5, 10]            |
| - Confidence < 0.70 → TierObservation                           |
| - Count >= 10 → TierAutoUpdate, etc.                            |
+--------------------------+--------------------------------------+
                           |
                           v
+----------------------------------------------------------------+
| WritePromotion (learner.go:142)                                |
| - Appends Promotion record to tier-promotions.jsonl             |
+----------------------------------------------------------------+

Signal waste: HARNESS-002 fields populated but ignored by classifier
(research.md §1.4 pathology P3).
```

### 2.2 Target State (post-V3R4-003)

```
+-----------------------------------------------------------------+
| usage-log.jsonl (unchanged from HARNESS-002)                    |
+--------------------------+--------------------------------------+
                           |
                           v
+----------------------------------------------------------------+
| AggregatePatterns (learner.go:46) — EXTENDED                   |
| - Stage 1: existing logic, unchanged                            |
|   ↓                                                             |
| - Stage 2 (only if stage_2_enabled): NEW                        |
|   - Read singletons (Count=1) from Stage-1 output               |
|   - For each: fingerprint via classifier_simhash.go             |
|   - Cluster via classifier_cluster.go (pairwise Hamming ≤ τ)    |
|   - Emit merged Pattern records (two-field keys)                |
|   - Append audit log to .moai/harness/cluster-merges.jsonl      |
+--------------------------+--------------------------------------+
                           |
                           v
+----------------------------------------------------------------+
| ClassifyTier (learner.go:113) — UNCHANGED                      |
| - Now applied to BOTH Stage-1 singletons AND Stage-2 merges    |
+--------------------------+--------------------------------------+
                           |
                           v
+----------------------------------------------------------------+
| WritePromotion (learner.go:142) — UNCHANGED                    |
| - Promotion records may now carry 2-field OR 3-field keys       |
+----------------------------------------------------------------+

  +----------------------------------------------+
  | Config gate (harness.yaml)                   |
  | learning.enabled: true (REQ-HRN-FND-009)     |
  | learning.classifier.stage_2_enabled: bool   |
  | learning.classifier.similarity_algorithm:    |
  |   simhash | none                              |
  | learning.classifier.hamming_threshold: int   |
  | learning.classifier.cluster_min_size: int    |
  | learning.classifier.feature_fields: [...]    |
  +----------------------------------------------+

  +----------------------------------------------+
  | 5-Layer Safety — UNCHANGED                   |
  | L1 Frozen Guard / L2 Canary / L3             |
  | Contradiction / L4 Rate Limit (≤1/week) /    |
  | L5 Human Oversight (AskUserQuestion)         |
  +----------------------------------------------+
```

Key invariants:
- `buildPatternKey` (learner.go:99) unchanged.
- `ClassifyTier` (learner.go:113) unchanged.
- `Pattern`, `Promotion`, `Proposal`, `Tier`, `EventType` structs unchanged.
- 5-Layer Safety pipeline input contract unchanged.
- Observer hot path unchanged.
- New code lives in: `classifier_simhash.go` (NEW), `classifier_cluster.go` (NEW), modifications to `learner.go::AggregatePatterns` (additive Stage-2 invocation only).

### 2.3 Stage 2 Algorithm Flow

```
Input: map[string]*Pattern from Stage 1 (three-field keys)
       AND replay of usage-log.jsonl to recover singleton source events

Step 1 (Singleton identification):
  singletons := []*SingletonInput{}
  for each Pattern p in stage1Map:
    if p.Count == 1:
      // Recover the source Event from usage-log.jsonl (re-read & match)
      evt := findEventByKey(logPath, p.Key)
      singletons = append(singletons, &SingletonInput{
        Pattern:  p,
        Event:    evt,  // carries PromptPreview, PromptLang, AgentName, AgentType
      })

Step 2 (Partition by EventType — EDGE-008):
  groups := map[EventType][]*SingletonInput{}
  for _, s := range singletons:
    groups[s.Event.EventType] = append(groups[...], s)

Step 3 (For each EventType group, fingerprint):
  for et, group := range groups:
    for _, s := range group:
      featureStr := buildFeatureString(s.Event, configFeatureFields)
      // featureStr = "subject:/moai plan|prompt_preview:hello|prompt_lang:en|..."
      // Excludes PromptContent (REQ-HRN-CLS-014)
      s.Fingerprint = SimHash64(featureStr)

Step 4 (Pairwise Hamming clustering — Union-Find or naive):
  clusters := []Cluster{}
  for et, group := range groups:
    uf := NewUnionFind(len(group))
    for i := 0; i < len(group); i++:
      for j := i+1; j < len(group); j++:
        d := Hamming(group[i].Fingerprint, group[j].Fingerprint)
        if d <= configHammingThreshold:
          uf.Union(i, j)
    // Extract clusters of size >= configMinSize
    for root, members := range uf.Components():
      if len(members) >= configMinSize:
        clusters = append(clusters, Cluster{EventType: et, Members: members, ...})

Step 5 (Emit merged Pattern + audit log):
  for _, c := range clusters:
    canonicalSubject := lexMin(c.MemberSubjects())
    merged := &Pattern{
      Key:         fmt.Sprintf("%s:%s", c.EventType, canonicalSubject),
      EventType:   c.EventType,
      Subject:     canonicalSubject,
      ContextHash: "",  // session-agnostic
      Count:       sum(c.MemberCounts()),
      Confidence:  mean(c.MemberConfidences()),
    }
    stage1Map[merged.Key] = merged

    // Emit audit log
    appendClusterMergeAudit(.moai/harness/cluster-merges.jsonl, c, merged)

Output: stage1Map (now contains Stage-1 singletons + Stage-2 merges)
```

Complexity:
- Step 1: O(n + k) where n = events in log, k = unique patterns
- Step 2: O(k)
- Step 3: O(s * m) where s = singletons, m = avg feature string length (per research.md §2.1 ≈ 20µs/pattern)
- Step 4: O(s² / |EventTypes|) — partitioning reduces worst-case quadratic by typical 3-4x factor
- Step 5: O(c) where c = number of clusters
- Total batch cost: ~25ms for s=50 singletons (research.md §3.3 estimate)

### 2.4 File-Level Changes (Run-Phase Reference)

| File | Operation | Wave | Notes |
|------|-----------|------|-------|
| `.moai/specs/SPEC-V3R4-HARNESS-003/research.md` | Pre-existing | Plan | Phase 0 deliverable (read-only input). |
| `.moai/specs/SPEC-V3R4-HARNESS-003/spec.md` | Created | Plan | This SPEC's main document. |
| `.moai/specs/SPEC-V3R4-HARNESS-003/plan.md` | Created | Plan | This file. |
| `.moai/specs/SPEC-V3R4-HARNESS-003/acceptance.md` | Created | Plan | AC definitions. |
| `.moai/specs/SPEC-V3R4-HARNESS-003/tasks.md` | Created | Plan | Task breakdown. |
| `.moai/specs/SPEC-V3R4-HARNESS-003/spec-compact.md` | Created | Plan | Auto-generated compact form. |
| `internal/harness/classifier_simhash.go` | Created | Wave B | SimHash fingerprint module. New file. |
| `internal/harness/classifier_simhash_test.go` | Created | Wave B | Table-driven SimHash unit tests. New file. |
| `internal/harness/classifier_cluster.go` | Created | Wave C | Stage-2 clustering + audit log. New file. |
| `internal/harness/classifier_cluster_test.go` | Created | Wave C | Stage-2 unit + integration tests. New file. |
| `internal/harness/classifier_cluster_audit_test.go` | Created | Wave C | Cluster-merges.jsonl schema test. New file. |
| `internal/harness/classifier_cluster_bench_test.go` | Created | Wave D | p99 ≤ 25ms benchmark. New file. |
| `internal/harness/classifier_pii_test.go` | Created | Wave D | AC-HRN-CLS-009 PII guard test. New file. |
| `internal/harness/classifier_schema_regression_test.go` | Created | Wave D | AC-HRN-CLS-011 Promotion schema preservation test. New file. |
| `internal/harness/classifier_frozen_guard_regression_test.go` | Created | Wave D | AC-HRN-CLS-006 FROZEN guard non-regression test. New file. |
| `internal/harness/classifier_rate_limit_test.go` | Created | Wave D | AC-HRN-CLS-010 rate-limit non-regression test. New file. |
| `internal/harness/learner.go` | Modified | Wave A + C | Refactor: extract Stage-2 hook in `AggregatePatterns` (additive call site, no behavior change when stage_2_enabled=false). |
| `internal/harness/learner_test.go` | Modified | Wave A | Add `TestStage1BackwardCompat_StageDisabled` (AC-HRN-CLS-001 golden baseline). |
| `internal/harness/types.go` | Unchanged | — | NO field additions/modifications. `Pattern`, `Promotion`, `Proposal`, `Tier`, `EventType` all unchanged. |
| `.moai/config/sections/harness.yaml` | Modified | Wave D | Additive: `learning.classifier` sub-block (5 keys). Default `stage_2_enabled: false`. |
| `internal/config/...` (config loader) | Modified | Wave D | Parse new `learning.classifier` block; validate `hamming_threshold`, `cluster_min_size`, `similarity_algorithm`; fail-safe to defaults. |
| `testdata/stage1_baseline.jsonl` | Created | Wave A | Golden input fixture (1000 events, 10 distinct 3-field-key combos × 100 repetitions). |
| `testdata/stage1_baseline_patterns.json` | Created | Wave A | Golden output fixture (Stage-1 pattern map for comparison). |
| `testdata/stage2_similar_10.jsonl` | Created | Wave C | Fixture: 10 similar singletons. |
| `testdata/stage2_dissimilar_10.jsonl` | Created | Wave C | Fixture: 10 dissimilar singletons. |
| `testdata/stage2_perf_1k.jsonl` | Created | Wave D | Fixture: 1000 events for benchmark. |
| `testdata/legacy_promotions.jsonl` | Created | Wave D | Fixture: 3 pre-V3R4-003 three-field-key promotions. |
| `go.mod` / `go.sum` | Optionally Modified | Wave B | If using `glaslos/go-simhash` library; otherwise unchanged (embedded SimHash). |
| `CHANGELOG.md` | Modified | Wave D | Add v2.22.0 entry under "Added" subsection. |
| `.moai/release/RELEASE-NOTES-v2.22.0.md` | Created (or appended) | Wave D | Stage-2 opt-in migration story; recommended `hamming_threshold` tuning guidance. |

**Files explicitly NOT modified in this PR** (per REQ-HRN-CLS-009, REQ-HRN-CLS-011 contracts):
- `internal/harness/observer.go` (Observer hot path preserved)
- `internal/harness/frozen_guard.go` (FROZEN guard preserved)
- `internal/harness/safety/pipeline.go` (5-Layer pipeline preserved)
- `internal/harness/safety/canary.go`, `safety/rate_limit.go`, `safety/oversight.go`, `safety/contradiction.go` (preserved)
- `internal/cli/hook.go` (observer hook handlers preserved per REQ-HRN-CLS-011)
- `.claude/rules/moai/design/constitution.md` (FROZEN)
- `.claude/rules/moai/core/agent-common-protocol.md` (FROZEN)
- `.claude/rules/moai/core/askuser-protocol.md` (FROZEN)
- Any path under `.claude/agents/moai/**`, `.claude/skills/moai-*/**`, `.claude/rules/moai/**`, or `.moai/project/brand/**` (FROZEN per REQ-HRN-FND-006)

---

## 3. Wave Decomposition (Run-Phase Roadmap)

### 3.1 Wave A — Stage-1 Refactor (Aggregator Interface Seam)

**Goal**: Refactor `AggregatePatterns` at `internal/harness/learner.go:46` to expose a clean seam where Stage 2 can plug in, without changing any observable behavior. Establish the golden baseline test fixture (`stage1_baseline.jsonl` + `stage1_baseline_patterns.json`) that AC-HRN-CLS-001 will compare against forever.

**Owner**: `manager-tdd` (or `manager-develop` per quality.yaml development_mode)

**Inputs**:
- This SPEC's REQ-HRN-CLS-001 (Stage-1 backward compatibility)
- REQ-HRN-CLS-004 (Stage-2 disabled equivalence)
- Existing `internal/harness/learner.go` Stage-1 implementation

**Tasks** (T-A1 through T-A4, see tasks.md):
- T-A1: Generate the golden baseline fixture `testdata/stage1_baseline.jsonl` (1000 events; 10 distinct (event_type, subject, context_hash) combos × 100 each, matching existing `learner_test.go:64-68` pattern). RED test.
- T-A2: Generate the golden output fixture `testdata/stage1_baseline_patterns.json` (Stage-1 pattern map serialized via `json.MarshalIndent`).
- T-A3: Refactor `AggregatePatterns` to introduce a deferred Stage-2 call site at the END of the function (after the existing scan/aggregate loop completes), gated by config check. When `stage_2_enabled` is false (default), the call site is a no-op (no SimHash, no allocations, no audit log emission). When true, it invokes the (yet-to-exist in Wave A) `clusterSingletons` function from `classifier_cluster.go`. Wave A leaves `clusterSingletons` as a stub returning the input map unchanged.
- T-A4: Add `TestStage1BackwardCompat_StageDisabled` to `learner_test.go` asserting `reflect.DeepEqual(AggregatePatterns("testdata/stage1_baseline.jsonl"), golden)` returns true.

**Acceptance** (Wave A complete when):
- AC-HRN-CLS-001 passes — Stage-1 byte-identical pattern map with `stage_2_enabled: false`.
- Golden baseline test green (RED → GREEN transition complete).
- `go test ./internal/harness/...` no regressions.

**Risk**:
- Refactor introduces subtle behavior change. Mitigation: golden baseline test catches any diff. The refactor is purely a deferred call site addition; the existing scanner loop body is unchanged.

---

### 3.2 Wave B — Stage-2 SimHash Module

**Goal**: Implement `internal/harness/classifier_simhash.go` exposing `SimHash64(featureStr string) uint64` and `Hamming(a, b uint64) int`. Author table-driven tests covering determinism, order-insensitivity, range, and edge cases (empty input, single-char input).

**Owner**: `manager-tdd` (delegating to `expert-backend` for Go SimHash implementation)

**Inputs**:
- This SPEC's REQ-HRN-CLS-002 (SimHash fingerprint generation)
- Research.md §2.1 (algorithm mechanics, library options)
- **Decision (Wave B-finalized)**: Embed a minimal SimHash implementation locally (no external dependencies). Rationale: aligns with project's zero-dep + pure-Go preference (research.md §9). The reference implementation `glaslos/go-simhash` (~150 LOC core) serves as a design guide; our local implementation will be derived but independent.

**Tasks** (T-B1 through T-B4, see tasks.md):
- T-B1: Implement `internal/harness/classifier_simhash.go` with `SimHash64`, `Hamming`, and an internal `tokenize(s string) []string` helper using simple whitespace + lowercase normalization. RED unit tests first.
- T-B2: Add `internal/harness/classifier_simhash_test.go` with table-driven cases: determinism (same input → same output), order-insensitivity (token order swap → same fingerprint OR adjusted via canonicalization), known-input fingerprints (3-5 hand-computed examples), empty/short-input handling.
- T-B3: Add a `featureFields` enumeration constant in `classifier_simhash.go` declaring the canonical feature order: `["subject", "prompt_preview", "prompt_lang", "agent_name", "agent_type"]`. Add `buildFeatureString(evt *Event, fields []string) string` that concatenates fields with a deterministic separator (e.g., `"|"`) and explicitly does NOT include `prompt_content` (PII rule REQ-HRN-CLS-014).
- T-B4: Static guard: ensure `grep -nE 'PromptContent' internal/harness/classifier_simhash.go` returns 0 matches. CI assertion.

**Acceptance** (Wave B complete when):
- `go test -run TestSimHash64 ./internal/harness/...` passes (all table cases).
- `Hamming(SimHash64("hello world"), SimHash64("hello world")) == 0`.
- `Hamming(SimHash64("hello world"), SimHash64("xyz unrelated")) > 20` (sanity check).
- PII grep guard returns 0 matches in classifier_simhash.go.

**Risk**:
- SimHash implementation bug producing non-deterministic output. Mitigation: table-driven tests with hand-computed expected fingerprints. Determinism is a hard requirement.
- Hash function choice (we recommend FNV-1a 64-bit for token hashing, then weighted bit-set aggregation per standard SimHash algorithm). Tradeoff documented in tasks.md T-B1.

---

### 3.3 Wave C — Stage-2 Clustering + Audit Log

**Goal**: Implement `internal/harness/classifier_cluster.go` exposing `clusterSingletons(input map[string]*Pattern, evtLookup func(key string) *Event, cfg ClassifierConfig) (map[string]*Pattern, error)` that performs the Stage 2 algorithm flow described in §2.3.

**Owner**: `manager-tdd`

**Inputs**:
- This SPEC's REQ-HRN-CLS-003 (Hamming clustering), REQ-HRN-CLS-005 (cluster merge emission), REQ-HRN-CLS-012 (confidence floor), REQ-HRN-CLS-013 (audit log)
- Wave B's `SimHash64` + `Hamming` + `buildFeatureString` functions

**Tasks** (T-C1 through T-C6, see tasks.md):
- T-C1: Define `ClassifierConfig` struct in `classifier_cluster.go` with fields `Stage2Enabled bool`, `SimilarityAlgorithm string`, `HammingThreshold int`, `ClusterMinSize int`, `FeatureFields []string`. Add `DefaultClassifierConfig()` returning the documented defaults.
- T-C2: Implement `clusterSingletons` function with the §2.3 algorithm flow: singleton identification, EventType partitioning, fingerprinting, Union-Find Hamming clustering, merged Pattern emission, audit log emission.
- T-C3: Implement `appendClusterMergeAudit(path string, cluster Cluster, merged *Pattern) error` that writes a JSONL line per REQ-HRN-CLS-013 schema (`ts`, `member_keys`, `member_counts`, `hamming_distances` upper-triangle, `merged_key`, `merged_count`, `confidence`). Use `O_APPEND|O_CREATE|O_WRONLY` mode 0o644.
- T-C4: Implement `findEventByKey(logPath, key string) *Event` that re-reads `usage-log.jsonl` and returns the FIRST event whose three-field key matches. Used to recover singleton source events for fingerprinting. (Alternative: pass events through from Wave A refactor; choice deferred to implementation discretion.)
- T-C5: Add `internal/harness/classifier_cluster_test.go` with `TestStage2Aggregates10SimilarSingletons` (AC-HRN-CLS-002), `TestStage2DissimilarSingletonsStaySeparate` (AC-HRN-CLS-003), `TestClusterMergeEmissionShape` (AC-HRN-CLS-004), `TestConfidenceFloorForcesObservation` (AC-HRN-CLS-007), `TestStage2HonorsTierThresholdsOverride` (AC-HRN-CLS-013).
- T-C6: Add `internal/harness/classifier_cluster_audit_test.go` with `TestClusterMergeAuditLogSchema` (AC-HRN-CLS-008) verifying JSONL schema.

**Acceptance** (Wave C complete when):
- AC-HRN-CLS-002, AC-HRN-CLS-003, AC-HRN-CLS-004, AC-HRN-CLS-007, AC-HRN-CLS-008, AC-HRN-CLS-013 all pass.
- `go test -run TestStage2 ./internal/harness/...` passes (all sub-cases).
- `.moai/harness/cluster-merges.jsonl` schema verified via JSON unmarshal in tests.

**Risk**:
- Union-Find implementation bug producing wrong cluster boundaries. Mitigation: hand-crafted small fixtures (3-5 singletons) with known expected clusters; assertion of cluster membership in test.
- Cluster-merges.jsonl write race condition under concurrent AggregatePatterns. Mitigation: per EDGE-009, the existing `O_APPEND` semantics handle concurrency. Test uses `t.Parallel()` to exercise.
- Confidence-floor logic placement. Must be applied AFTER mean computation, BEFORE ClassifyTier consults thresholds. Mitigation: AC-HRN-CLS-007 test asserts the exact behavior.

---

### 3.4 Wave D — Config Wiring + End-to-End Test + Performance Budget

**Goal**: Extend `.moai/config/sections/harness.yaml` schema with the `learning.classifier` sub-block; wire config parsing through to `clusterSingletons`; implement fail-safe behavior for invalid config; verify performance budget; verify PII guard; verify FROZEN guard non-regression; verify rate-limit non-regression; verify Promotion schema preservation.

**Owner**: `manager-tdd` (delegating to `expert-devops` for config schema validation)

**Inputs**:
- This SPEC's REQ-HRN-CLS-010 (perf budget), REQ-HRN-CLS-014 (PII), REQ-HRN-CLS-015 (rate limit), REQ-HRN-CLS-017 (tier_thresholds), REQ-HRN-CLS-018 (invalid config fail-safe)
- Wave A/B/C deliverables

**Tasks** (T-D1 through T-D5, see tasks.md):
- T-D1: Extend `.moai/config/sections/harness.yaml` with `learning.classifier` block (5 keys with defaults). Update `internal/config/loader.go` (or equivalent) to parse + validate. Validation rules: `hamming_threshold` in [0, 64]; `cluster_min_size` ≥ 2; `similarity_algorithm` in {simhash, none}; on any failure, log warning to stderr and fall back to defaults (REQ-HRN-CLS-018).
- T-D2: Add `internal/harness/classifier_cluster_bench_test.go::BenchmarkClusterSingletons1k` with fixture `testdata/stage2_perf_1k.jsonl`. Report p50/p95/p99 wall-clock.
- T-D3: Add `internal/harness/classifier_pii_test.go::TestStage2NeverIncludesPromptContent` (AC-HRN-CLS-009). Setup writes a usage-log entry with `PromptContent` containing sentinel string; assert sentinel does NOT appear in cluster-merges.jsonl. Plus CI guard: `grep -nE 'PromptContent' internal/harness/classifier_*.go` returns 0.
- T-D4: Add three regression-guard tests:
  - `classifier_schema_regression_test.go::TestPromotionAndProposalSchemaPreserved` (AC-HRN-CLS-011)
  - `classifier_frozen_guard_regression_test.go::TestFrozenGuardUnaffectedByStage2` (AC-HRN-CLS-006)
  - `classifier_rate_limit_test.go::TestStage2DoesNotInflateRateLimit` (AC-HRN-CLS-010)
- T-D5: Add `classifier_cluster_test.go::TestInvalidConfigFailsSafeToStage1` (AC-HRN-CLS-014) iterating over 5 invalid-config cases (negative threshold, oversized threshold, min_size=1, unknown algorithm, garbage algorithm string).
- T-D6: Update `CHANGELOG.md` v2.22.0 "Added" subsection. Create `.moai/release/RELEASE-NOTES-v2.22.0.md` with Stage-2 opt-in migration story.

**Acceptance** (Wave D complete when):
- AC-HRN-CLS-005, AC-HRN-CLS-006, AC-HRN-CLS-009, AC-HRN-CLS-010, AC-HRN-CLS-011, AC-HRN-CLS-013, AC-HRN-CLS-014 all pass.
- `go test ./internal/harness/...` no regressions across the existing test suite.
- Benchmark reports p99 ≤ 25ms (informational gate).
- PII grep guard returns 0 matches across all new classifier files.

**Risk**:
- Config parser changes affect unrelated tests. Mitigation: additive schema only; existing keys (`enabled`, `auto_apply`, `log_retention_days`, `rate_limit`, `tier_thresholds`) unchanged.
- Performance budget miss on slower hardware. Mitigation: budget is informational; CI does not block on it. Tasks.md T-D2 documents remediation path (follow-up optimization SPEC).
- Promotion schema regression test depends on accurate fixture. Mitigation: T-D4 task includes pre-V3R4-003 three-field-key fixture generation explicitly.

---

## 4. Test Strategy

### 4.1 Unit Tests (Wave A-D)

| Test File | Tests | Linked AC |
|-----------|-------|-----------|
| `learner_test.go` (modified) | `TestStage1BackwardCompat_StageDisabled` | AC-HRN-CLS-001 |
| `classifier_simhash_test.go` (new) | `TestSimHash64_Determinism`, `TestSimHash64_OrderInsensitivity`, `TestSimHash64_KnownInputs`, `TestSimHash64_EmptyInput`, `TestHamming_KnownPairs`, `TestBuildFeatureString_ExcludesPromptContent` | AC-HRN-CLS-002, AC-HRN-CLS-009 |
| `classifier_cluster_test.go` (new) | `TestStage2Aggregates10SimilarSingletons`, `TestStage2DissimilarSingletonsStaySeparate`, `TestClusterMergeEmissionShape`, `TestConfidenceFloorForcesObservation`, `TestStage2HonorsTierThresholdsOverride`, `TestInvalidConfigFailsSafeToStage1` (table-driven, 5 cases) | AC-HRN-CLS-002, AC-HRN-CLS-003, AC-HRN-CLS-004, AC-HRN-CLS-007, AC-HRN-CLS-013, AC-HRN-CLS-014 |
| `classifier_cluster_audit_test.go` (new) | `TestClusterMergeAuditLogSchema`, `TestAuditLogAppendOnly` | AC-HRN-CLS-008 |
| `classifier_pii_test.go` (new) | `TestStage2NeverIncludesPromptContent` | AC-HRN-CLS-009 |
| `classifier_schema_regression_test.go` (new) | `TestPromotionAndProposalSchemaPreserved` | AC-HRN-CLS-011 |
| `classifier_frozen_guard_regression_test.go` (new) | `TestFrozenGuardUnaffectedByStage2` | AC-HRN-CLS-006 |
| `classifier_rate_limit_test.go` (new) | `TestStage2DoesNotInflateRateLimit` | AC-HRN-CLS-010 |
| `classifier_cluster_bench_test.go` (new) | `BenchmarkClusterSingletons1k` | AC-HRN-CLS-005 |

### 4.2 Static Verification (CI Guards)

- `grep -nE 'PromptContent' internal/harness/classifier_simhash.go internal/harness/classifier_cluster.go` returns zero matches (PII rule per REQ-HRN-CLS-014).
- `grep -nE 'classifier_simhash|classifier_cluster|clusterSingletons' internal/cli/hook.go internal/harness/observer.go` returns zero matches (Observer hot-path preservation per REQ-HRN-CLS-011).
- `git diff main -- internal/harness/frozen_guard.go` returns zero non-comment changes (FROZEN guard preservation per REQ-HRN-CLS-009).
- `git diff main -- internal/harness/safety/pipeline.go` returns zero changes (5-Layer Safety preservation per REQ-HRN-CLS-008).
- `git diff main -- internal/harness/types.go` shows additions only in constants/comments, NOT in `Pattern`, `Promotion`, `Proposal`, `Tier`, `EventType` struct fields (REQ-HRN-CLS-006, -007, -008).
- `git diff main -- .claude/rules/moai/design/constitution.md` returns zero changes.

### 4.3 Integration Tests (Manual + Automated)

- Run a synthetic 1000-event log through the full pipeline (Observer → AggregatePatterns Stage 1 → Stage 2 → ClassifyTier → WritePromotion). Verify the produced `tier-promotions.jsonl` contains both 3-field and 2-field PatternKey values, all parseable into the existing `Promotion` struct.
- Simulate FROZEN zone violation: synthesize a Phase 4 coordinator call with a target path under `.claude/skills/moai-`; verify `EnsureAllowed` returns `*FrozenViolationError` (FROZEN guard unchanged from pre-V3R4-003).
- Simulate rate-limit pressure: 12 Tier-4-eligible merged Patterns in a single batch; verify only 1 AskUserQuestion invocation per 7-day window.

### 4.4 Plan-Auditor Run

- `Agent(subagent_type: "plan-auditor")` invoked at Phase 2.5 of this `/moai plan` session with the 5 SPEC artifacts as input.
- Pass threshold: 0.80 minimum. Below that, iterative drafts (max 3 iterations).
- If iteration 3 still fails, escalate findings to orchestrator as blocker report.

---

## 5. MX Tag Plan (Phase 3.5 finalized)

This section captures the canonical MX annotation targets for SPEC-V3R4-HARNESS-003. Orchestrator's Phase 3.5 MX-Injection during run-phase MUST apply these tags; the Phase 2.9 quality gate verifies their presence.

### 5.1 New annotation targets (classifier_simhash.go + classifier_cluster.go)

| Target | Tag | Priority | Reason |
|--------|-----|----------|--------|
| `internal/harness/classifier_simhash.go::SimHash64` | `@MX:ANCHOR` | P1 | Public function used by classifier_cluster.go and tests (fan_in ≥ 2 expected). |
| `internal/harness/classifier_simhash.go::buildFeatureString` | `@MX:NOTE` | P1 | **PII guard** — closed switch over `featureFields` literal explicitly excludes `prompt_content`; REQ-HRN-CLS-014 / AC-HRN-CLS-009 contract surface. |
| `internal/harness/classifier_simhash.go::tokenize` | `@MX:NOTE` | P2 | Deterministic Unicode-aware tokenizer; output ordering MUST be stable to keep SimHash deterministic (REQ-HRN-CLS-002). |
| `internal/harness/classifier_cluster.go::clusterSingletons` | `@MX:ANCHOR` | P1 | Stage-2 entry point invoked by learner.go::AggregatePatterns when `learning.classifier.stage_2_enabled=true` (REQ-HRN-CLS-003). |
| `internal/harness/classifier_cluster.go::clusterSingletons` | `@MX:WARN` | P1 | **O(s²) Hamming-distance computation** over singleton set; s capped at 50 by `cluster_singleton_cap` to preserve 25 ms p99 budget (REQ-HRN-CLS-010 / AC-HRN-CLS-005). |
| `internal/harness/classifier_cluster.go::buildRepresentativeKey` | `@MX:NOTE` | P2 | Synthetic Stage-2 key uses `event_type:subject` (two-field) format per `types.go:196` design intent; aligns Promotion schema with classifier output (P4 fix). |
| `internal/harness/classifier_cluster.go::appendClusterMergeAudit` | `@MX:NOTE` | P1 | Audit log emission to `.moai/harness/cluster-merges.jsonl`; PII-sensitive — must never write `prompt_content` (REQ-HRN-CLS-014); cap `hamming_distances` at 20 (REQ-HRN-CLS-013 / AC-HRN-CLS-008). |
| `internal/harness/classifier_cluster.go::ClassifierConfig.Validate` | `@MX:NOTE` | P2 | Fail-safe config parser: in-range invalid + `yaml.Unmarshal` type errors fall back to `DefaultClassifierConfig()` and emit WARN (REQ-HRN-CLS-018 / AC-HRN-CLS-014). |

### 5.2 Existing tag extensions (learner.go + types.go)

| Target | Existing Tag | Action | Reason |
|--------|--------------|--------|--------|
| `internal/harness/learner.go::AggregatePatterns` (line 44) | `@MX:ANCHOR` | **EXTEND** `@MX:REASON` | Add `classifier_cluster.go` to fan_in list once Wave A merges; new fan_in becomes 4 (existing 3 + cluster invocation). |
| `internal/harness/learner.go::ClassifyTier` (line 22 group) | `@MX:ANCHOR` | **NO CHANGE** | Confidence floor logic unchanged (REQ-HRN-CLS-012); same callers. |
| `internal/harness/types.go::Pattern` (line 129) | `@MX:ANCHOR` | **EXTEND** `@MX:REASON` | Add `classifier_cluster.go` as new fan_in member (cluster output materializes as Pattern). |
| `internal/harness/types.go::Tier` enum (line 129) | `@MX:ANCHOR` | **NO CHANGE** | Schema preserved byte-identical (REQ-HRN-CLS-006). |
| `internal/harness/types.go::Promotion` (line 191 group) | (no tag) | **ADD** `@MX:NOTE` | Document line 196 design intent: `PatternKey` is `event_type:subject` (two fields, no context_hash); Stage-2 representative keys honor this; Stage-1 keys continue three-field for BC. |

### 5.3 Config + audit log artifacts (non-code annotations)

| Target | Tag | Reason |
|--------|-----|--------|
| `.moai/config/sections/harness.yaml` `learning.classifier.*` block (NEW) | `@MX:NOTE` (in plan.md) | Five tunables: `stage_2_enabled`, `hamming_threshold` (default 3), `cluster_min_size`, `cluster_singleton_cap`, `feature_fields`. Validation in `ClassifierConfig.Validate()`. |
| `.moai/harness/cluster-merges.jsonl` (NEW audit artifact) | `@MX:NOTE` (in plan.md) | One JSONL line per merge event; capped fields per REQ-HRN-CLS-013 D3.2 mitigation. |

### 5.4 Coverage summary

- New ANCHOR tags: 3 (SimHash64, clusterSingletons, learner.AggregatePatterns extension)
- New WARN tags: 1 (clusterSingletons O(s²) hot path)
- New NOTE tags: 6 (buildFeatureString, tokenize, buildRepresentativeKey, appendClusterMergeAudit, Validate, Promotion design intent)
- Existing tag extensions: 2 (REASON list expansion for AggregatePatterns + Pattern struct)
- Total annotation actions in run-phase Phase 3.5: 12

This coverage satisfies the implicit MX requirement: every function with fan_in ≥ 2 (REQ-HRN-CLS-002 callers + AC-005 benchmark + AC-009 PII test) has @MX:ANCHOR; every PII boundary has @MX:NOTE; the single O(s²) hot path has @MX:WARN.

---

## 6. Risks (Top 5)

| Risk | Likelihood | Severity | Mitigation |
|------|------------|----------|------------|
| False-positive cluster merges (semantically distinct patterns aggregated together) | Medium | Medium | `cluster_min_size ≥ 3` requires 3-way agreement; default `hamming_threshold: 3` aligns with research.md §8.3 recommended range (2-5 in 64-bit space); audit log enables post-hoc review and config tightening (looser values 5-12 acceptable). |
| PII leak via `PromptContent` accidentally entering classifier artifact | Low | High | REQ-HRN-CLS-014 forbids `PromptContent` use; CI grep guard returns 0 matches across classifier files; AC-HRN-CLS-009 test asserts sentinel string absence. |
| Performance regression beyond 25ms p99 budget | Low | Medium | Benchmark `BenchmarkClusterSingletons1k`; budget informational; remediation path documented (follow-up optimization SPEC). Worst-case O(s²) for s=50 singletons is well under budget. |
| Stage-1 backward-compat test (AC-HRN-CLS-001) breaks after unrelated refactor | Low | Medium | Golden baseline file (`testdata/stage1_baseline_patterns.json`) is the contract; regenerate only if Stage-1 change is intentional and documented in a separate SPEC. |
| SimHash library `glaslos/go-simhash` becomes unmaintained | Low | Low | Choose embedded implementation in Wave B (recommended; ~80 LOC overhead); eliminates external dependency risk entirely. REQ-HRN-CLS-018 fail-open path also covers library load failure. |

(See spec.md §8 for the full 7-risk table.)

---

## 7. Dependencies

### 7.1 Inbound Dependencies (this SPEC depends on)

- `SPEC-V3R4-HARNESS-001` (foundation) — establishes 4-tier ladder (REQ-HRN-FND-011), `learning.enabled` gate (REQ-HRN-FND-009), 5-Layer Safety (REQ-HRN-FND-005), FROZEN zone (REQ-HRN-FND-006), 7-day Tier-4 rate-limit floor (REQ-HRN-FND-012), orchestrator-only AskUserQuestion contract (REQ-HRN-FND-015).
- `SPEC-V3R4-HARNESS-002` (observation surface) — delivers 12 omitempty `Event` fields (PromptPreview, PromptLang, AgentName, AgentType) consumed by Stage 2 fingerprinting. Reuses REQ-HRN-OBS-012 (Strategy A default), REQ-HRN-OBS-013 (opt-in strategies), REQ-HRN-OBS-014 (fail-open).
- `.claude/rules/moai/design/constitution.md` (FROZEN) — referenced but not modified.

### 7.2 Outbound Dependencies (SPECs that depend on this)

- `SPEC-V3R4-HARNESS-004` (Reflexion self-critique) — consumes Stage-2 merged patterns as richer Tier-2+ candidate stream.
- `SPEC-V3R4-HARNESS-005` (principle-based scoring) — orthogonal; principle scores apply to both Stage-1 and Stage-2 patterns identically.
- `SPEC-V3R4-HARNESS-006` (multi-objective scoring) — orthogonal; multi-objective tuple applies to merged patterns as to original patterns.
- `SPEC-V3R4-HARNESS-007` (Voyager skill library) — orthogonal; embedding-indexed retrieval is a separate axis.
- `SPEC-V3R4-HARNESS-008` (cross-project federation) — orthogonal; federation operates on Promotion records regardless of 2-field or 3-field keys.

### 7.3 Co-temporal Dependencies (none)

This SPEC has no co-temporal dependencies. All downstream SPECs enter plan-phase only after this SPEC merges and the run-phase Wave A-D PRs complete.

---

## 8. Out of Scope (Explicit List)

(Mirrors spec.md §1.3 + §4; reproduced for plan-phase reference.)

1. Reflexion-style self-critique loop → `SPEC-V3R4-HARNESS-004`.
2. Principle-based scoring rubric → `SPEC-V3R4-HARNESS-005`.
3. Multi-objective effectiveness measurement → `SPEC-V3R4-HARNESS-006`.
4. Voyager-style skill library → `SPEC-V3R4-HARNESS-007`.
5. Cross-project lesson federation → `SPEC-V3R4-HARNESS-008`.
6. TF-IDF / character n-gram fallback algorithm — research.md §2.2 confirmed not materially better than Option A; fail-safe is `stage_2_enabled: false`, not a second algorithm.
7. Migration tooling for historical `tier-promotions.jsonl` — old entries remain in place with 3-field keys; coexistence is verified by AC-HRN-CLS-011.
8. Online/incremental clustering at observer record time — Stage 2 is strictly batch.
9. Embedding-model inference (transformer, ONNX, Ollama, external service) — research.md §2.2 ruled out.
10. DBSCAN / HDBSCAN / k-means clustering via sklearn-go — research.md §2.3 ruled out.
11. Modification of `Promotion`, `Proposal`, `Event`, `Pattern`, `Tier`, `EventType` schemas.
12. Modification of `internal/harness/observer.go` or any hook handler.
13. Modification of 5-Layer Safety pipeline layers.
14. Modification of FROZEN files (constitution, agent-common-protocol, askuser-protocol).
15. Modification of `internal/cli/hook.go` (observer hot-path preserved per REQ-HRN-CLS-011).
16. New CLI verb or slash command surface change.
17. New subagent definition.
18. Networking, telemetry, external API calls.

---

## 9. Execution Order (Plan-Phase)

This `/moai plan SPEC-V3R4-HARNESS-003` session executes the following phases:

| Phase | Activity | Status |
|-------|----------|--------|
| Phase 0 | ultrathink deep research (read research.md, HARNESS-001/002 SPECs, code surface) | Complete |
| Phase 1 | Branch setup: `plan/SPEC-V3R4-HARNESS-003` from `origin/main` HEAD | Pending (orchestrator) |
| Phase 2 | Draft `spec.md` with 18 EARS-format REQs | Complete |
| Phase 2.5 | Invoke `plan-auditor` subagent (max 3 iterations) | Pending |
| Phase 2.75 | Pre-commit quality gate (markdown lint, frontmatter validation) | Pending |
| Phase 2.8 | Draft `acceptance.md` with 14 ACs | Complete |
| Phase 2.9 | Draft `plan.md` (this file) | Complete |
| Phase 2.10 | Draft `tasks.md` | Complete |
| Phase 2.11 | Auto-generate `spec-compact.md` from spec.md | Complete |
| Phase 3 | Commit + delegate PR creation to `manager-git` via `Agent()` | Pending |

The run-phase Waves A-D happen in `/moai run SPEC-V3R4-HARNESS-003` after the plan PR merges.

---

## 10. References

### Sibling Artifacts

- `.moai/specs/SPEC-V3R4-HARNESS-003/research.md` — Phase 0 research deliverable (46935 bytes; pathology validation, algorithm survey, performance budget, PII risk, backward-compat, recommendation D+A).
- `.moai/specs/SPEC-V3R4-HARNESS-003/spec.md` — REQ-HRN-CLS-001 through REQ-HRN-CLS-018.
- `.moai/specs/SPEC-V3R4-HARNESS-003/acceptance.md` — AC-HRN-CLS-001 through AC-HRN-CLS-014.
- `.moai/specs/SPEC-V3R4-HARNESS-003/tasks.md` — T-A1 through T-D6 task breakdown.
- `.moai/specs/SPEC-V3R4-HARNESS-003/spec-compact.md` — Compact REQ + AC + exclusions for downstream consumers.

### Upstream SPECs

- `.moai/specs/SPEC-V3R4-HARNESS-001/spec.md` (foundation contracts).
- `.moai/specs/SPEC-V3R4-HARNESS-001/plan.md` (Wave decomposition pattern referenced).
- `.moai/specs/SPEC-V3R4-HARNESS-002/spec.md` (Event field source).
- `.moai/specs/SPEC-V3R4-HARNESS-002/acceptance.md` (AC pattern source).

### Code Surface

- `internal/harness/learner.go:46` (AggregatePatterns), `:99` (buildPatternKey), `:113` (ClassifyTier), `:17` (confidenceThreshold).
- `internal/harness/types.go:53-119` (Event with HARNESS-002 omitempty fields), `:131-145` (Tier enum), `:163-187` (Pattern), `:191-209` (Promotion), `:220-244` (Proposal).
- `internal/harness/observer.go:53` (RecordEvent), `:103` (RecordExtendedEvent).
- `internal/harness/frozen_guard.go:18` (allowedPrefixes), `:27` (frozenPrefixes), `:60` (IsAllowedPath), `:103` (EnsureAllowed).
- `internal/harness/safety/pipeline.go:89` (Evaluate entry point).
- `.moai/config/sections/harness.yaml:115-127` (learning block).

### External References

- SimHash: Manku, Jain, Sarpatwar, VLDB 2007.
- `glaslos/go-simhash`: github.com/glaslos/go-simhash (MIT, pure Go, last commit 2024-01).
- Reflexion: Li et al., arXiv:2303.11366 (downstream SPEC-V3R4-HARNESS-004).
- Voyager: Wang et al., arXiv:2305.16291 (downstream SPEC-V3R4-HARNESS-007).
- Constitutional AI: Bai et al., arXiv:2212.08073 (downstream SPEC-V3R4-HARNESS-005).

---

End of plan.md.
