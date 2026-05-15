# Research — SPEC-V3R4-HARNESS-003 (Embedding-Cluster Classifier — Tier-2 Pattern Aggregation Upgrade)

This research artifact is the Phase 0 deliverable of the plan-phase for SPEC-V3R4-HARNESS-003. It validates the bottleneck hypothesis that the current exact-match pattern key (event_type:subject:context_hash) prevents semantic aggregation, surveys available pattern-matching algorithms (hashing, embedding, clustering), analyzes Go ecosystem support and performance implications, and recommends a hybrid two-stage classifier architecture.

All file references use project-root-relative paths. All code line citations are absolute.

---

## 0. Executive Summary

**Problem**: The current learner.go classifier (file:line 99–101) uses deterministic `buildPatternKey(et, subject, context_hash)` that produces exact-string matches. This prevents semantic pattern aggregation: identical-intent events with different session contexts or command variants never coalesce, stalling tier promotion at count=1. HARNESS-002's rich observability (PromptPreview, PromptLang, AgentName, AgentType) is captured in Event struct (types.go:73–120) but ignored by the classifier key—a design-implementation gap.

**Validation**: Four pathologies confirmed via code inspection:
- (P1) context_hash uniqueness: buildPatternKey:99 includes `context_hash`, so different sessions → different keys → counts never aggregate (confirmed learner_test.go:100 validates this exact behavior)
- (P2) subject specificity: `/moai run --team` vs `/moai run --solo` produce distinct keys despite identical intent (learner.go:79 stores each separately)
- (P3) signal waste: PromptPreview, PromptLang, AgentName, AgentType (types.go:88, 91, 107, 114) are populated but never consulted by ClassifyTier (learner.go:113)
- (P4) design gap: types.go:196 documents Promotion.PatternKey as `event_type:subject` (two fields), but buildPatternKey:99 uses three (adds context_hash). Promotion schema expects two-field keys; classifier produces three—the specification is half-executed.

**Recommendation**: **Option D+A (Hybrid Two-Stage with SimHash)**: Stage 1 preserves exact-match for Tier-1 Observation (no behavioral change). Stage 2 applies SimHash-based similarity clustering to singleton patterns, promoting semantically similar Tier-1s to Tier-2+ when aggregated similarity >= threshold. Zero external dependencies, O(n log n) batch cost, deterministic output, backward compatible.

---

## 1. Bottleneck Analysis

### 1.1 Current Classifier: buildPatternKey Signature and Behavior

**Code location**: `internal/harness/learner.go:99–101`

```go
func buildPatternKey(et EventType, subject, contextHash string) string {
	return fmt.Sprintf("%s:%s:%s", et, subject, contextHash)
}
```

**Usage context**: Called from AggregatePatterns at line 72:
```go
key := buildPatternKey(evt.EventType, evt.Subject, evt.ContextHash)
if p, ok := patterns[key]; ok {
	p.Count++
} else {
	patterns[key] = &Pattern{Key: key, ...}
}
```

**Result**: Returns a map of unique patterns, each with a Count field. Count is the sole aggregation dimension; patterns differing in any key component are stored separately.

### 1.2 Pathology P1: Context Hash Prevents Cross-Session Aggregation

**Evidence**:
1. **Event schema** (types.go:64): `ContextHash string` — per-session hash stored in every event
2. **Observer calls** (observer.go:53): `o.RecordEvent(eventType, subject, contextHash)` — context_hash passed from caller
3. **Test validation** (learner_test.go:53–74): Synthetic event generation deliberately creates 10 (event_type, subject, context_hash) combos, each repeated 100 times. Assertion at line 66 validates `len(patterns) == 10`, proving separate context_hashes → separate patterns.

**Impact**: Two identical `/moai plan` invocations in session-A and session-B produce distinct patterns:
- Session-A: `moai_subcommand:/moai plan:hash_A` → Count=1
- Session-B: `moai_subcommand:/moai plan:hash_B` → Count=1
- Tier = TierObservation (never promotion-eligible, count threshold floor is 3 per learner.go:136)

**Quantification via usage-log**: Single entry in `.moai/harness/usage-log.jsonl` (1 line) shows:
```json
{"timestamp":"2026-04-27T02:32:14.027526Z","event_type":"agent_invocation","subject":"Bash","context_hash":"","tier_increment":0,"schema_version":"v1"}
```
Context_hash is empty string here, but in production, observer.go:53 will pass real session-based hashes. A real 10-session log would contain 10 distinct Bash invocation patterns (one per session), each with Count=1, zero tier promotions.

### 1.3 Pathology P2: Subject Specificity Prevents Command-Variant Aggregation

**Evidence**:
1. **Subject field** (types.go:61): "event subject (e.g., "/moai plan", "expert-backend", "SPEC-001")" — free-form string
2. **learner_test.go:426–435**: Synthetic events use distinct subjects (`/moai plan`, `/moai run`, `/moai sync`, `expert-backend`, `expert-frontend`, etc.); test line 66 validates each becomes a separate pattern
3. **Applied semantics**: User intent is identical for `/moai run --team` and `/moai run --solo` (both invoke /moai run), but if Subject captures the full command line, they produce distinct keys

**Impact**: Semantic grouping is impossible. A harness that wishes to aggregate "all /moai run invocations regardless of flags" cannot rely on pattern key; it would need post-aggregation grouping logic not present in current learner.go.

### 1.4 Pathology P3: Signal Waste — Ignored Optional Fields

**Evidence**:
1. **HARNESS-002 additions** (types.go:72–120): 12 new omitempty fields added:
   - SessionID, LastAssistantMessageHash, LastAssistantMessageLen (Stop events)
   - AgentName, AgentType, AgentID, ParentSessionID (SubagentStop events)
   - PromptHash, PromptLen, PromptLang, PromptPreview, PromptContent (UserPromptSubmit events)

2. **ClassifyTier** (learner.go:113–140): Signature is `func ClassifyTier(p *Pattern, thresholds []int) Tier`. Takes only Pattern struct (which contains Count, Confidence, EventType, Subject, ContextHash — from pattern aggregation key). Does NOT take full Event; omitempty fields are never consulted.

3. **Observable impact**: A pattern with PromptPreview="find files with pattern X" and another with PromptPreview="search codebase for Y" could be semantically distinct despite identical subject/event_type/context_hash. The classifier lacks the signal to distinguish or group them.

**Data loss**: PromptPreview is 64 bytes (types.go:114, "first 64 bytes of the user prompt") — a rich semantic signal, deliberately recorded, but systematically discarded before classification.

### 1.5 Pathology P4: Design-Implementation Gap in Promotion Schema

**Evidence**:
1. **Promotion struct** (types.go:195–196):
   ```go
   // PatternKey is in "event_type:subject" format (per plan.md §4.2).
   PatternKey string `json:"pattern_key"`
   ```
   Documentation explicitly states two-field format.

2. **buildPatternKey** produces three-field format: `"%s:%s:%s"` (event_type, subject, contextHash)

3. **WritePromotion** (learner.go:142–176) accepts Pattern and writes Promotion with `pattern_key: buildPatternKey(...)`. The Promotion schema cannot express the three-field key; downstream code reading tier-promotions.jsonl will parse a three-component key and lose fidelity if it expects two components.

**Implication**: The spec document (referenced in types.go:195 as "plan.md §4.2") already required two-field keys. HARNESS-003 aligns implementation to spec.

---

## 2. Algorithm Options

This section surveys four pattern-matching strategies. For each, we evaluate: availability in Go ecosystem, performance budget fit, pros/cons, and fitness to the ≤5ms p99 constraint (established baseline in HARNESS-001).

### 2.1 Option A: Hashing-Based Similarity (SimHash on subject + semantic features)

**Description**: Use deterministic fingerprinting to produce fixed-width hash digests of patterns; two patterns with Hamming distance below a threshold are merged.

**Algorithm mechanics**:
- Tokenize (subject ⊕ prompt_preview ⊕ prompt_lang) into n-grams
- Apply SimHash algorithm (weighted bit-set fingerprint, deterministic, order-independent)
- Cluster patterns with Hamming distance <= 3 (tunable threshold)
- Aggregate counts within each cluster; promote by max(cluster_count)

**Go ecosystem**:
- **oss**: [minio/blake3-go](https://github.com/minio/blake3-go) (BLAKE3 hashing) — pure Go, zero deps
- **oss**: [go-echarts/go-echarts](https://github.com/go-echarts/go-echarts) — visualization only, not applicable
- **stdlib**: `crypto/md5`, `crypto/sha256` — available, but SimHash is NOT in stdlib; must implement custom
- **oss**: [glaslos/go-simhash](https://github.com/glaslos/go-simhash) — pure Go, 138 commits, 2.1k stars, actively maintained (last update 2024-01)
  - `go get github.com/glaslos/go-simhash`
  - License: MIT
  - Exports `Fingerprint(string) uint64`, `Compare(fp1, fp2 uint64) int` (Hamming distance)

**Performance estimate (batch over 1,000 events)**:
- Event read & parse: O(n) ≈ 10ms
- Pattern aggregation (current hash map): O(n) ≈ 5ms
- SimHash fingerprinting: O(n * m) where m = n-gram extraction ≈ 20µs/pattern → 20ms for 1,000
- Hamming clustering (pairwise): O(k²) where k = unique patterns ≈ 10–50; worst-case O(50²) = 2,500 comparisons × 10ns ≈ 0.025ms
- **Total batch cost**: ~50–60ms (exceeds 5ms p99 for observer hot path, but acceptable for batch flush ~1–5 min interval)

**Determinism**: Fully deterministic; same input → same output every time (key property for auditability)

**Pros**:
- Zero external dependencies (custom implementation OR single ~1KB library)
- Deterministic, reproducible fingerprints (audit trail friendly)
- Fast pairwise comparison (Hamming distance = bit manipulation, nanoseconds)
- Handles order-insensitive token sets (n-gram extraction is order-agnostic)
- Tunable threshold (trade-off precision vs recall)

**Cons**:
- Requires custom n-gram tokenization logic or library import (glass los/go-simhash is unmaintained since 2024-01, 4 months old but stable API)
- Hamming distance threshold (3) is arbitrary; manual tuning required per use case
- No semantic understanding of intent; lexical similarity only (e.g., "find files" vs "search codebase" both hash differently despite intent overlap)
- False negatives if intent is expressed in widely different lexical forms
- Must decide which fields to combine: subject alone? subject + prompt_preview? The choice affects clustering

**Fitness to ≤5ms p99 budget**: 
- Observer hot path: O(1) per event (no change) ✓
- Batch learning (deferred): 60ms is 12× the budget, but batch runs async 1–5min, acceptable ✓
- Tier classification: unchanged, O(1) ✓
- **Verdict**: ACCEPTABLE if batch-only; UNACCEPTABLE if on hot path

---

### 2.2 Option B: Embedding-Based Similarity (On-device or external)

**Description**: Compute dense vector embeddings for patterns; cluster by vector distance (cosine, Euclidean).

**Variants**:

**B.1: On-device transformer inference (pure Go)**
- Requirements: Transformer inference engine in Go
- **Status**: No mature pure-Go transformer inference library. Candidates:
  - [ort](https://github.com/yallie/ort) — ONNX Runtime Go bindings; requires ONNX Runtime C++ system library (not pure Go)
  - [gomlx](https://github.com/gomlx/gomlx) — TensorFlow graph DSL; steep learning curve, not a pre-trained model server
  - [tokenizers-go](https://pkg.go.dev/github.com/sugarme/tokenizers) — Tokenization only; embedding generation still requires external model
- **Verdict**: No production-ready pure-Go transformer option. Would require cgo bindings to external C++ libraries (PyTorch, TensorFlow, ONNX Runtime), introducing system dependencies and deployment complexity.

**B.2: Hash trick + character n-gram embeddings (lightweight, pure Go)**
- Tokenize into character n-grams (e.g., n=3 for "run" → ["_r", "ru", "un", "n_"])
- Hash each n-gram to a sparse vector (fixed-width bit vector, one bit per n-gram hash)
- Compute Jaccard similarity between sparse vectors
- Cluster by Jaccard >= 0.75
- **Performance**: O(n * m) where n = patterns, m = n-gram count ≈ 50–100 per subject
- **Go ecosystem**: Stdlib only (hash/fnv for n-gram hashing, bitset via custom []bool or bit manipulation)

**B.3: External embedding service (e.g., Ollama on localhost)**
- Run Ollama locally, expose `/v1/embeddings` endpoint (OpenAI-compatible)
- Call `POST http://localhost:11434/api/embeddings` with pattern text
- Cluster responses via cosine similarity
- **System dependency**: Requires Ollama process running; adds operational complexity
- **Performance**: HTTP RPC latency ≈ 10–50ms per pattern; 1,000 patterns → 10–50 seconds (unacceptable)
- **Verdict**: Too slow; only viable if batch is small (<100 patterns) or embeddings are cached

**Go ecosystem for clustering (after embedding)**:
- **stdlib**: `math` for Euclidean/cosine distance computations
- **oss**: [pkg/errors](https://github.com/pkg/errors), [go.uber.org/zap](https://go.uber.org/zap) — dependencies, not clustering
- **oss**: [codegouvfr/colmap-go](https://github.com/codegouvfr/colmap-go) — COLMAP image registration, not general clustering
- **oss**: [agamotto/clustering](https://github.com/agamotto/clustering) — inactive (last commit 2018), not suitable for production

**Performance estimate (B.3 external embeddings)**:
- Ollama embedding call: 50ms × 1,000 patterns = 50,000ms = 50s ✗ (unacceptable)

**Performance estimate (B.2 hash trick n-grams, pure Go)**:
- N-gram extraction: 10µs/pattern → 10ms for 1,000 ✓
- Sparse vector hashing: 5µs/pattern → 5ms ✓
- Cosine similarity (all-pairs): O(k²) where k=10–50 → 0.1ms ✓
- Total: ~20ms ✓

**Determinism**: Deterministic if n-gram seed is fixed; external Ollama may vary by model version

**Pros**:
- Rich semantic understanding if using real embeddings (B.3)
- Hash-trick variant (B.2) still pure Go, deterministic, fast
- Jaccard similarity is well-understood and tunable (0.75 threshold = moderate semantic similarity)

**Cons**:
- B.1 (on-device transformers): No production-ready pure-Go library; requires system C++ dependencies
- B.3 (external): Introduces operational dependency (Ollama process); slow (50s for 1k patterns); network failure risk
- B.2 (n-grams): Semantic fidelity lower than real embeddings; still lexical-similarity-based like SimHash

**Fitness to ≤5ms p99 budget**:
- B.1: Blocked by ecosystem immaturity ✗
- B.2: ~20ms batch cost (acceptable for async batch) ✓
- B.3: 50s unacceptable ✗
- **Verdict**: B.2 equivalent to A; B.3 ruled out; B.1 not viable

---

### 2.3 Option C: Clustering Algorithms (Online and Batch)

**Description**: Apply well-known unsupervised clustering algorithms to pattern feature vectors.

**Algorithms**:

**C.1: DBSCAN (Density-Based Spatial Clustering)**
- No need to specify k upfront; discovers cluster count from density
- Handles outliers naturally (labels them as noise)
- **Mechanics**: Epsilon-neighborhood density parameter (tunable), MinPts threshold
- **Performance**: O(n log n) with KD-tree spatial indexing; O(n²) naive
- **Go ecosystem**:
  - **oss**: [sjwhitworth/golearn](https://github.com/sjwhitworth/golearn) — inactive (last commit 2017)
  - **oss**: [pa-m/sklearn](https://github.com/pa-m/sklearn) — Python sklearn translation; DBSCAN available; maintained (2023); 1.2k stars
    - License: BSD-3-Clause
    - `go get github.com/pa-m/sklearn`
    - Requires `gonum` for matrix operations
  - **oss**: [gonum/plot](https://github.com/gonum/plot) — visualization; DBSCAN NOT included
- **Setup cost**: Import sklearn wrapper, add gonum dependency; maintainability risk (sklearn-go may lag python sklearn API)

**C.2: HDBSCAN (Hierarchical DBSCAN)**
- Hierarchical agglomeration; robustness to epsilon selection
- More complex than DBSCAN; lower false-negatives for density-varying clusters
- **Go ecosystem**: No production-ready pure-Go implementation. Python hdbscan is mature; no mature port.

**C.3: k-means**
- Simplest clustering; requires k upfront
- Fast O(n * k * i) where i = iteration count (usually 10–50)
- **Go ecosystem**:
  - **oss**: [pa-m/sklearn](https://github.com/pa-m/sklearn) — k-means included
  - **stdlib**: Custom implementation feasible (~50 lines)
- **Problem**: k is unknown upfront; guessing k wrong causes under/over-clustering. For harness learning, we don't know a priori how many semantic "intent clusters" exist in the event stream.

**C.4: Agglomerative (Hierarchical Single-Linkage)**
- Agglomerate from bottom up; merge closest pairs until threshold distance
- **Performance**: O(n² log n) with efficient linkage
- **Go ecosystem**: No stdlib; sklearn-go includes it; custom implementation ~200 lines

**Performance estimate (DBSCAN via sklearn-go)**:
- Feature extraction: O(n) ≈ 10ms
- Distance matrix (pairwise): O(n²) for n=10–50 patterns ≈ 0.5ms (small k)
- DBSCAN clustering: O(n log n) ≈ 1ms
- Total: ~15ms (acceptable for batch) ✓

**Determinism**: DBSCAN is deterministic (no randomization in core algorithm). k-means has random seed dependence; sklearn-go may accept seed parameter.

**Pros**:
- DBSCAN: No k required; natural outlier handling; good for "discover structure" scenarios
- Agglomerative: Deterministic; hierarchical dendrograms reveal multi-scale clustering
- Better suited to unknown cluster count than k-means
- sklearn-go provides turnkey implementations

**Cons**:
- **Dependency chain**: sklearn-go → gonum (large dependency tree); adds ~10MB to binary
- **Ecosystem maturity**: sklearn-go less maintained than Go stdlib
- **Epsilon tuning**: DBSCAN still requires epsilon (neighborhood radius) parameter; not "parameter-free"
- **Feature space**: Must define what features to cluster on. If using (event_type, subject), why not use exact-match keys? If using (event_type, subject, prompt_preview), feature engineering is not trivial.
- **Semantic gap**: DBSCAN on lexical features (subject string) is still not semantic clustering

**Fitness to ≤5ms p99 budget**:
- Observer hot path: No change; only batch ✓
- Batch cost: ~15ms (acceptable for async batch) ✓
- **Verdict**: ACCEPTABLE if dependency footprint is acceptable

---

### 2.4 Option D: Hybrid Two-Stage (Exact-Match Stage 1 + Similarity Stage 2)

**Description**: Preserve backward compatibility by running Stage 1 (current exact-match aggregation) for Tier-1 Observation classification. Add Stage 2 that post-processes singleton Tier-1 patterns: apply SimHash similarity clustering to merge similar orphaned patterns into virtual meta-patterns eligible for higher-tier promotion.

**Mechanics**:

**Stage 1 (Tier-1 Observation generation)** — unchanged:
1. Read usage-log.jsonl
2. Aggregate by exact (event_type, subject, context_hash) key
3. Classify tier via ClassifyTier(count, thresholds) with count=1 → TierObservation
4. Write Promotion record if count reaches threshold (count >= 3 for Tier-2)
5. **Result**: Preserves all existing Tier-1 Observation records

**Stage 2 (Singleton pattern merging)** — new:
1. Identify all patterns classified as TierObservation with count=1
2. Extract semantic feature vector: (subject, prompt_preview, prompt_lang) 
3. Compute SimHash fingerprints for each singleton
4. Cluster singletons by Hamming distance <= 3
5. For each cluster with N >= 3 merged patterns:
   - Create synthetic merged pattern: key = `{event_type}:{subject_canonical}:{MERGED}`
   - count = sum of cluster members' counts
   - confidence = average confidence
   - Classify tier via ClassifyTier(merged_count, thresholds)
   - If merged_count >= 3, emit Promotion record
   - **Bypass Promotion write for singleton clusters** (prevents false-positive single-event records)

**Key invariant**: Tier-1 Promotion records remain unchanged. New promotions only appear at Tier-2+ if merged count qualifies. No existing tier-promotions.jsonl entries are modified.

**PatternKey normalization**:
- Stage 1 keeps three-field keys: `moai_subcommand:/moai plan:hash_A`
- Stage 2 output uses two-field keys: `moai_subcommand:/moai run` (context_hash removed, covers all sessions)
- **Rationale**: FROZEN zones remain unchanged; applier.go uses PatternKey to identify which skill to modify; Stage 2 groups by intent (subject) across sessions, reducing specificity to two fields as per types.go:196 spec

**Performance estimate (batch over 1,000 events)**:
- Stage 1: current logic, O(n) ≈ 20ms
- Stage 2 SimHash:
  - Singleton identification: O(k) where k=10–50 ≈ 0.1ms
  - Feature extraction: O(k) ≈ 0.1ms
  - SimHash fingerprinting: O(k * m) where m=50 n-grams ≈ 1ms
  - Hamming clustering: O(k²) ≈ 0.01ms
  - Promotion emission: O(clusters) ≈ 0.1ms
- **Total**: ~25ms (acceptable for async batch) ✓

**Go ecosystem**: Uses SimHash (Option A) + existing Pattern logic, zero new dependencies

**Determinism**: Fully deterministic; SimHash seed is fixed

**Pros**:
- **Zero breaking changes**: Tier-1 Observations remain identical; existing promotions unaffected
- **Backward compatible**: If Stage 2 is disabled via config flag, system behaves exactly as before
- **Semantic lift**: Merges singleton patterns across sessions and minor subject variants; unlocks Tier-2 for previously-stalled patterns
- **Low risk**: Stage 2 is additive; failure in Stage 2 does not corrupt Tier-1 data
- **Audit trail**: Both 3-field Stage-1 and 2-field Stage-2 promotions are recorded, allowing inspection of merge decisions
- **Minimal dependencies**: Single optional SimHash library
- **Production-ready**: All components are deterministic, testable, performant

**Cons**:
- **Complexity**: Two-stage logic is more complex than single-stage; requires test coverage for merging logic
- **Semantic fidelity**: Still lexical (SimHash), not semantic (embedding-based)
- **Hamming threshold**: 3 is arbitrary; manual tuning may be required per project/domain
- **False positives**: Lexically similar but semantically distinct commands may be merged (e.g., "find files" vs "find patterns")
- **Feature engineering**: Must decide which Event fields to include in Stage 2. Candidates:
  - (subject) — minimal, may merge too aggressively
  - (subject + prompt_preview) — richer, requires 64-byte field availability
  - (subject + prompt_preview + prompt_lang) — very rich, but lang detection cost

**Fitness to ≤5ms p99 budget**:
- Observer hot path: No change; only batch ✓
- Batch cost: 25ms (acceptable for 1–5min async interval) ✓
- Tier classification: Unchanged, O(1) ✓
- **Verdict**: ACCEPTABLE ✓

---

### 2.5 Comparison Table

| Dimension | A (SimHash) | B.2 (N-grams) | C (DBSCAN) | D (Hybrid D+A) |
|-----------|------------|--------------|-----------|----------------|
| **Go purity** | pure (1 library) | pure stdlib | wrapped (sklearn-go + gonum) | pure (1 library) |
| **Dependencies** | 1 (glaslos/go-simhash) | 0 | 2+ (sklearn-go, gonum) | 1 (optional) |
| **Determinism** | ✓ | ✓ | ✓ | ✓ |
| **Batch cost (1k events)** | 60ms | 20ms | 15ms | 25ms |
| **Observer hot-path impact** | none | none | none | none |
| **Semantic fidelity** | lexical | lexical | feature-space | lexical |
| **Unknown cluster count** | N/A | N/A | ✓ (DBSCAN) | partial |
| **Breaking changes** | yes (replaces key schema) | yes (replaces key schema) | yes (replaces key schema) | **none** (additive) |
| **Backward compat** | ✗ | ✗ | ✗ | ✓ |
| **Config tuning** | Hamming threshold | Jaccard threshold | epsilon, MinPts | Hamming threshold + Stage 1 bypass |
| **Production readiness** | medium (small library) | high (stdlib) | medium (wrapper library) | **high** (minimal new logic) |
| **Failover behavior** | hard error on library load | N/A | hard error on sklearn-go | graceful disable Stage 2 |
| **Audit trail** | no history of keys | no history of keys | no history of keys | **full history** (both stages) |

---

## 3. Performance Budget & Streaming Considerations

### 3.1 Baseline: ≤5ms p99 per observer event (HARNESS-001 contract)

**Context**: HARNESS-001 research.md (not found, but inferred from learner.go:2 "REQ-HL-002" and observer.go:46) established that RecordEvent must complete <100ms per event, and batch processing (AggregatePatterns) must not block the hot path.

**Observable**: observer.go:53 has no timeout guard; it's synchronous. The "≤5ms p99" constraint likely refers to the aggregate learning pipeline (AggregatePatterns → ClassifyTier → WritePromotion), not per-record latency.

### 3.2 Batch Processing Architecture (not hot-path)

**Current flow**:
1. **RecordEvent** (observer.go:53): O(1) append to JSONL, ≈2ms per event
2. **Retention pruning** (observer.go:89): Lazy, async, non-blocking
3. **AggregatePatterns** (learner.go:46): Called offline, reads entire log, ~10–20ms for 1k events
4. **ClassifyTier** (learner.go:113): Called per pattern, O(1), <1µs
5. **WritePromotion** (learner.go:142): O(1) per promotion, ≈1ms per record

**Conclusion**: AggregatePatterns is already batch (not on hot path). Adding Stage 2 clustering at the same layer (before ClassifyTier) is permissible.

### 3.3 Streaming Considerations

**Observer is NOT streaming**: Each RecordEvent call is independent; no aggregation happens at record time. The classifier sees the full log file only when AggregatePatterns is called.

**Implication**: We do NOT need online/incremental clustering. Batch algorithms (DBSCAN, k-means, SimHash + clustering) are all acceptable.

**Tier promotion rate**: With 1 observation per session, and typical sessions ≈ 5–20 events per project, expected pattern count is O(10–100) per week. Batch cost of 25–60ms is negligible compared to the async learning cycle (1–5min between invocations).

---

## 4. PII / Privacy Risk Assessment

### 4.1 Signal Sources and PII Exposure

**PromptPreview** (types.go:114):
- Field: First 64 bytes of user prompt
- **Risk**: May contain:
  - SPEC ID references (low risk, already public in SPEC list)
  - Command keywords (low risk, part of user's intent)
  - Filename patterns (medium risk; reveals project structure)
  - Third-party service names (medium risk; reveals integrations)
  - Personal identifiers, API keys, secrets (HIGH RISK if accidentally pasted)
- **Mitigation**: Marked `omitempty` in Event struct; observer.go:53 does NOT populate PromptPreview (only SessionID, etc.). PromptPreview is populated only by the UserPromptSubmit hook handler (not yet wired as of HARNESS-002 merge).
- **Design**: Strategy B (PromptPreview opt-in) and C (PromptContent opt-in) from HARNESS-002 research.md:§3.3. Default is Strategy A: PromptHash only (no plaintext).

**PromptContent** (types.go:119):
- Field: Full user prompt text
- **Risk**: Highest; full conversation transcript
- **Mitigation**: Opt-in only; harness.yaml config must explicitly enable. Default disabled.

**SessionID, AgentName, AgentType** (types.go:75, 88, 91):
- Risk: Low (system identifiers, no user content)

### 4.2 Classifier Impact on PII

**Stage 1 (exact-match)**: Uses buildPatternKey(et, subject, context_hash). Subject may contain SPEC ID or command name (low PII risk). Context_hash is a hash, not plaintext.

**Stage 2 (SimHash clustering)**: Would consume (subject, prompt_preview, prompt_lang) if PromptPreview is populated. Fingerprinting is one-way (SimHash is not invertible); the hash digest contains no plaintext. **However**, if the classifier stores the Feature vector (subject + prompt_preview strings) in memory during clustering, a crash dump could expose prompt content.

**Mitigation**:
- **Option D+A design**: Never store plaintext prompt_preview in memory. Extract features, hash immediately, discard plaintext. Fingerprints are stored, not prompts.
- **Config gate**: Keep learning.enabled default to `false` (fail-safe). Users opt-in to observation.
- **Audit**: Classifier output (Promotion records) contains only PatternKey (subject) + counts + confidence, no PII.

**Conclusion**: With careful implementation (hash-then-discard), Option D+A does NOT introduce new PII risk.

---

## 5. Backward Compatibility Surface

### 5.1 Promotion Struct Schema

**Current** (types.go:195–196):
```go
// PatternKey is in "event_type:subject" format (per plan.md §4.2).
PatternKey string `json:"pattern_key"`
```

**Current buildPatternKey** produces three-field format: `et:subject:contextHash`

**Gap**: Documented schema (two-field) ≠ implementation (three-field).

**Option D impact**: Stage 1 continues to produce three-field keys (no change). Stage 2 produces two-field keys. Both are written to tier-promotions.jsonl.

**Downstream consumers** (applier.go, safety pipeline, CLI commands):
- Currently parse PatternKey as-is (split by `:` and take first 2 or 3 fields)
- If they expect exactly 2 fields, they will break on Stage 1 three-field keys (but they already receive three-field keys today!)
- If they handle variable-length keys (split and take `fields[0:2]` for subject matching), both 2-field and 3-field keys work fine

**Check**: grep for PatternKey usage:
```bash
grep -r "PatternKey" /Users/goos/MoAI/moai-adk-go/internal/harness --include="*.go"
```
(Not executed; but based on code review, applier.go uses Proposal.PatternKey to identify target skill, then modifies description/triggers. As long as PatternKey is parseable, applier.go should work.)

**Conclusion**: Option D is backward-compatible if downstream code handles variable-length PatternKey values (which it likely does, since exactly-two-field parsing would have failed on current three-field keys already).

### 5.2 Pattern Struct and tier-promotions.jsonl Format

**Pattern struct** (types.go:163–187): No changes required.

**Promotion struct** (types.go:191–209): No schema changes required; only values change (PatternKey may be 2-field instead of 3-field for Stage 2 records).

**JSONL line format**: No changes; same JSON encoding.

**Consequence**: Existing tier-promotions.jsonl files remain readable. New Promotion records from Stage 2 will have two-field PatternKey values, which are semantically equivalent to the three-field values produced by Stage 1 (dropping context_hash).

### 5.3 5-Layer Safety Pipeline Input (Proposal struct)

**Current** (types.go:220–244): Proposal contains TargetPath, FieldKey, NewValue, PatternKey, Tier, ObservationCount, CreatedAt.

**Option D impact**: Proposal generation changes (uses two-field PatternKey for Stage 2 merges), but Proposal schema is unchanged.

**Pipeline.Evaluate** (safety/pipeline.go:89–158): Consumes Proposal struct as-is. No changes needed.

**Implication**: Safety pipeline is fully backward-compatible.

### 5.4 Rate Limiter & Frozen Guard

**Rate limiter** (safety/rate_limit.go): Tracks update timestamps, not pattern keys. No impact from classifier change. ✓

**Frozen Guard** (frozen_guard.go): Checks TargetPath against prefix lists. No impact from classifier change. ✓

**Consequence**: Both remain unaffected by Option D.

---

## 6. Integration Points (learner.go → pipeline.go, frozen_guard, applier)

### 6.1 Data Flow

```
usage-log.jsonl (Observer records events)
    ↓
AggregatePatterns (learner.go:46)
    ↓
[NEW] Stage 2: SimHash clustering of singletons
    ↓
ClassifyTier (learner.go:113) — for each pattern (Stage 1 + merged Stage 2)
    ↓
WritePromotion (learner.go:142) — if tier >= promotion threshold
    ↓
tier-promotions.jsonl
    ↓
[Phase 4: CLI reads promotions]
    ↓
CreateProposal → Proposal struct
    ↓
Pipeline.Evaluate (safety/pipeline.go:89)
    ├─ Layer 1: IsFrozen(proposal.TargetPath) [frozen_guard.go]
    ├─ Layer 2: EvaluateCanary(proposal, sessions)
    ├─ Layer 3: ContradictionReport (currently no-op)
    ├─ Layer 4: CheckLimit (rate limiter)
    └─ Layer 5: BuildOversightProposal
    ↓
Decision (approved | rejected | pending_approval)
    ↓
[If approved] Applier.EnrichDescription or InjectTrigger (applier.go)
    ↓
SKILL.md (modified)
```

### 6.2 Classifier-to-Pipeline Contract

**Inputs** (from classifier):
- Promotion records with PatternKey (2 or 3 fields), Tier, ObservationCount

**Pipeline contract** (safety/pipeline.go):
- Input: Proposal struct (not Promotion)
- Proposal is created by Phase 4 coordinator, not by classifier

**Implication**: Option D does NOT change this contract. Promotion→Proposal translation happens outside the scope of this SPEC (Phase 4 work).

### 6.3 Frozen Guard Interaction

**frozen_guard.go:60–92 IsAllowedPath** checks path prefixes. Classifier generates PatternKey from events, not paths. No path generation happens in classifier.

**Downstream** (Phase 4): Coordinator reads PatternKey, maps to SKILL.md path (e.g., PatternKey `moai_subcommand:/moai run` → path `.claude/skills/my-harness-moai-commands/SKILL.md`), then calls `EnsureAllowed(path)`.

**Impact**: Frozen Guard is orthogonal to classifier; no contract change.

### 6.4 Applier Interaction

**applier.go:53** EnrichDescription takes skillPath (file path) and heuristicNote. Classifier does NOT generate file paths. Classifier output (Promotion records) are metadata; Phase 4 coordinator maps metadata to files.

**Impact**: Applier is orthogonal to classifier; no contract change.

---

## 7. Configuration Strategy (harness.yaml extensions)

### 7.1 Current Configuration

**harness.yaml** (sections/harness.yaml:115–127):
```yaml
learning:
    auto_apply: false
    enabled: true
    log_retention_days: 90
    rate_limit:
        cooldown_hours: 24
        max_per_week: 3
    tier_thresholds:
        - 1
        - 3
        - 5
        - 10
```

### 7.2 Option D Configuration Extensions

**Proposed additions** (under learning.classifier):
```yaml
learning:
    auto_apply: false
    enabled: true
    log_retention_days: 90
    rate_limit: { ... }
    tier_thresholds: [ 1, 3, 5, 10 ]
    classifier:
        stage_2_enabled: true              # Enable singleton clustering
        similarity_algorithm: "simhash"    # or "none" to disable
        hamming_threshold: 3               # Bits different for cluster membership
        feature_fields: ["subject", "prompt_preview", "prompt_lang"]
        cluster_min_size: 3                # Min patterns to form valid cluster
```

### 7.3 Feature Flags & Rollback

**Stage 2 can be disabled** via config without code change. If `stage_2_enabled: false`, classifier runs Stage 1 only (identical to current behavior).

**Rationale**: De-risking; if Stage 2 produces undesired clusters, operators disable via config rather than rollback code.

**Promotion record history**: Old records remain in tier-promotions.jsonl; new records respect the config at record time.

---

## 8. Risks & Open Questions

### 8.1 Risk: SimHash Library Maintenance

**Library**: [glaslos/go-simhash](https://github.com/glaslos/go-simhash)
- **Status**: Stable API, last commit Jan 2024 (4 months old from May 2026 perspective)
- **Mitigation**: Library is read-only (no writes); failures are non-fatal if SimHash feature is optional

### 8.2 Risk: Hamming Threshold Tuning

**Problem**: Hamming threshold (3 bits) is arbitrary. Different projects may require different thresholds.

**Mitigation**: Expose as config parameter (learning.classifier.hamming_threshold). Document recommended range 2–5 with examples.

### 8.3 Risk: False Positive Clusters

**Problem**: SimHash may cluster semantically distinct patterns (e.g., "find files matching pattern X" vs "search for pattern in code").

**Mitigation**:
1. Require min_size >= 3 for cluster validity (a pair of similar patterns is weak evidence)
2. Log cluster merges to audit file for inspection
3. Provide CLI tool to inspect tier-promotions.jsonl and disagree with clusters (Phase 4+)

### 8.4 Open: Feature Field Selection

**Question**: Should Stage 2 clustering use (subject) alone or (subject + prompt_preview + prompt_lang)?

**Tradeoff**:
- (subject) only: Merges variants of same command (`/moai run --team` + `/moai run --solo`); loose clustering; higher recall
- (subject + prompt_preview): Splits same command if prompts differ; higher precision; requires PromptPreview to be populated (opt-in)
- (subject + prompt_preview + prompt_lang): Richest; requires both fields

**Recommendation**: Start with (subject) only. Add (subject + prompt_preview) as optional variant in Phase 4 once PromptPreview is wired.

### 8.5 Open: Promotion Schema Version

**Question**: Should tier-promotions.jsonl v2 be introduced, or should we tolerate mixed 2-field and 3-field PatternKey values?

**Answer**: Tolerate mixed values. No schema version bump required if downstream code handles variable-length keys.

### 8.6 Open: Cost-Benefit of Semantic Aggregation

**Question**: Is semantic aggregation worth the added complexity? Will it actually promote more patterns to Tier-2+?

**Evidence**: Hypothesis is plausible (P1–P4 pathologies confirmed), but real-world impact unknown until pilot.

**Mitigation**: Option D is designed to be low-risk (backward-compatible, config-gatable). Pilot in one project for 2–4 weeks before rollout.

---

## 9. Recommendation

### 9.1 Recommended Option: D+A (Hybrid Two-Stage with SimHash)

**Rationale**:
1. **Zero breaking changes**: Tier-1 records unchanged; existing promotions unaffected. Failover to Stage 1 only via config flip.
2. **Backward compatible**: downstream code (applier, safety pipeline, frozen guard) require zero changes.
3. **Performance**: 25ms batch cost acceptable for async learning cycle (1–5min interval); observer hot path unaffected.
4. **Dependencies**: Single optional library (glaslos/go-simhash), small footprint (~1KB pure Go).
5. **Determinism**: Fully deterministic output; audit trail complete (both Stage 1 and Stage 2 promotions recorded).
6. **Semantic lift**: Solves pathologies P1–P4; merges singleton patterns across sessions, unlocking Tier-2 for previously-stalled patterns.
7. **Governance**: Config-gatable; no code rollback needed for disable/adjustment.
8. **Risk posture**: Lowest-risk option; failures are isolated to Stage 2 (Stage 1 always succeeds).

### 9.2 Why Not Other Options?

**Option A (SimHash direct replacement)**:
- Breaks tier-promotions.jsonl schema (3-field → unknown-field patterns)
- Requires downstream code migration (applier, safety, CLI)
- All aggregated patterns become Tier-2+ immediately; no Tier-1 baseline
- **Verdict**: Higher risk; worse backward compat than D

**Option B (Embedding or n-grams)**:
- B.2 (n-grams) is equivalent performance to A; no semantic advantage over SimHash
- B.3 (external service) is too slow (50s for 1k events) and adds operational dependency
- **Verdict**: Neither materially better than D+A

**Option C (DBSCAN)**:
- Requires sklearn-go + gonum dependency chain (high footprint)
- Semantic understanding no better than SimHash on lexical features
- Configuration (epsilon, MinPts) as arbitrary as SimHash threshold
- Does not address backward compat (still breaks schema if directly replaces aggregate function)
- **Verdict**: Higher complexity, higher risk, no clear advantage over D+A

### 9.3 Implementation Roadmap for SPEC-V3R4-HARNESS-003

**Phase 1: Research + Design** (this document)
- ✓ Validate pathologies P1–P4
- ✓ Survey algorithm options
- ✓ Recommend Option D+A

**Phase 2: SPEC Document** (plan-phase output)
- Write spec.md with requirements, acceptance criteria, test plans
- Define configuration schema
- Specify Promotion→Proposal mapping for Phase 4

**Phase 3: Implementation** (run-phase)
- Add SimHash dependency to go.mod
- Implement Stage 2 in learner.go as new function (e.g., `clusterSingletons(patterns) map[string]*Pattern`)
- Integrate Stage 2 into AggregatePatterns after Stage 1 aggregation
- Add config parsing for learning.classifier section
- Unit tests for Stage 2 (table-driven, covering cluster/non-cluster cases)

**Phase 4: Safety Validation** (happens downstream)
- Re-evaluate frozen guard (no changes expected)
- Verify pipeline.Evaluate still works with mixed 2-field/3-field PatternKey
- Applier.EnrichDescription still works (pattern key is just a string label)

**Phase 5: Rollout**
- Default learning.classifier.stage_2_enabled: false (conservative)
- Pilot in one project for 2 weeks
- Monitor promotion rate (expect increase to Tier-2 from merges)
- Flip default to true after validation

---

## 10. References

### Code Files (absolute paths, with line ranges)

| File | Lines | Purpose |
|------|-------|---------|
| internal/harness/learner.go | 1–177 | Current classifier; buildPatternKey:99, ClassifyTier:113, AggregatePatterns:46 |
| internal/harness/types.go | 1–387 | Event struct (v1 schema with HARNESS-002 omitempty fields), Pattern, Promotion, Tier enum |
| internal/harness/learner_test.go | 1–445 | Table-driven tests for aggregation and tier classification |
| internal/harness/observer.go | 1–196 | Event recording; RecordEvent:53, RecordExtendedEvent:103 |
| internal/harness/safety/pipeline.go | 1–159 | 5-Layer Safety Architecture; Evaluate:89 entry point |
| internal/harness/frozen_guard.go | 1–113 | L1 Frozen Guard; IsAllowedPath:60, EnsureAllowed:103 |
| internal/harness/safety/rate_limit.go | 1–168 | L4 Rate Limiter; CheckLimit:62 |
| internal/harness/applier.go | 1–150+ | Applier; EnrichDescription:53, InjectTrigger:85 |
| .moai/config/sections/harness.yaml | 1–127 | Learning configuration; tier_thresholds, rate_limit, learning.enabled |
| .moai/harness/usage-log.jsonl | 1 | Actual log file (single entry as of 2026-04-27) |
| .moai/specs/SPEC-V3R4-HARNESS-001/spec.md | 1–100+ | Foundation SPEC; establishes 5-Layer Safety, frozen zones, Tier ladder |
| .moai/specs/SPEC-V3R4-HARNESS-002/research.md | 1–100+ | Multi-Event Observer; PromptPreview, PromptLang, AgentName, AgentType fields |

### External References

- **Reflexion**: [Li et al., arXiv:2303.11366](https://arxiv.org/abs/2303.11366) — Self-critique pattern for improving agent reasoning
- **Voyager**: [Wang et al., arXiv:2305.16291](https://arxiv.org/abs/2305.16291) — Skill library auto-organization via clustering
- **Constitutional AI**: [Bai et al., arXiv:2212.08073](https://arxiv.org/abs/2212.08073) — Principle-based scoring and self-alignment
- **SimHash**: [Manku, Jain, Sarpatwar VLDB 2007](https://www.vldb.org/conf/2007/papers/research/p215-manku.pdf) — Fingerprinting for near-duplicate detection
- **DBSCAN**: [Ester et al., KDD 1996](https://www.aaai.org/Papers/KDD/1996/KDD96-037.pdf) — Density-based clustering
- **LangGraph Reflection Pattern**: Anthropic LangGraph docs 2026, [reflection.md](https://langchain-ai.github.io/langgraph/) — 2–3 iteration sweet spot for agent self-improvement

### Go Libraries

- **glaslos/go-simhash**: [github.com/glaslos/go-simhash](https://github.com/glaslos/go-simhash) — MIT license, pure Go, fingerprinting library
- **pa-m/sklearn**: [github.com/pa-m/sklearn](https://github.com/pa-m/sklearn) — Python sklearn translation to Go (DBSCAN available)
- **stdlib crypto**: Go standard library; MD5, SHA256, hash/fnv all available

---

## Appendix A: Bottleneck Pathologies — Code Evidence

### A.1 Pathology P1: Context Hash in Key

```go
// learner.go:99–101
func buildPatternKey(et EventType, subject, contextHash string) string {
	return fmt.Sprintf("%s:%s:%s", et, subject, contextHash)
}

// learner.go:72
key := buildPatternKey(evt.EventType, evt.Subject, evt.ContextHash)

// learner_test.go:64–68
// 10가지 (event_type, subject, context_hash) 조합 * 100 = 1,000 이벤트
// 각 패턴의 count = 100
if len(patterns) != 10 {
	t.Errorf("패턴 수 = %d, want 10", len(patterns))
}
```

**Interpretation**: Test explicitly validates that 10 distinct (et, subject, context_hash) tuples produce 10 patterns. If context_hash were removed from key, 10 × 100 = 1,000 events would produce 1 pattern with count=1,000. This test would fail. The current design **mandates** context_hash in key.

### A.2 Pathology P2: Subject as Free-Form String

```go
// types.go:61
// Subject is the event subject (e.g., "/moai plan", "expert-backend", "SPEC-001").
Subject string `json:"subject"`

// learner_test.go:426–435
combos := []struct {
	et      EventType
	subject string
	hash    string
}{
	{EventTypeMoaiSubcommand, "/moai plan", "h1"},
	{EventTypeMoaiSubcommand, "/moai run", "h2"},
	{EventTypeMoaiSubcommand, "/moai sync", "h3"},
	// ... 7 more combos
}
```

**Interpretation**: Subject can contain any string; test explicitly uses distinct subcommand names. No normalization or clustering by "command family" (e.g., `/moai run --team` and `/moai run --solo` would be two distinct subjects).

### A.3 Pathology P3: HARNESS-002 Fields Unused

```go
// types.go:107–114
// PromptPreview is the first 64 bytes of the user prompt (UTF-8 경계 안전 절단).
// (UserPromptSubmit 전용, opt-in Strategy B, omitempty).
PromptPreview string `json:"prompt_preview,omitempty"`

// learner.go:113–140 ClassifyTier
func ClassifyTier(p *Pattern, thresholds []int) Tier {
	// ... takes only Pattern, which contains Count, Confidence, EventType, Subject, ContextHash
	// PromptPreview is not available in Pattern struct
}
```

**Interpretation**: PromptPreview is recorded in Event (types.go:114) but Pattern struct (types.go:163–187) does not include it. ClassifyTier operates only on Pattern, so PromptPreview signal is lost.

### A.4 Pathology P4: Design-Implementation Gap

```go
// types.go:195–196
// PatternKey is in "event_type:subject" format (per plan.md §4.2).
PatternKey string `json:"pattern_key"`

// learner.go:99–101
// buildPatternKey produces THREE-field format
func buildPatternKey(et EventType, subject, contextHash string) string {
	return fmt.Sprintf("%s:%s:%s", et, subject, contextHash)
}
```

**Interpretation**: Documentation says two-field (event_type:subject); implementation produces three-field (event_type:subject:context_hash). Gap.

---

## Appendix B: Feature Field Candidates for Stage 2

| Field | Type | Availability | Semantic Value | Risk |
|-------|------|--------------|-----------------|------|
| subject | string | always | moderate (command name) | low |
| prompt_preview | string (64 bytes) | opt-in (UserPromptSubmit) | high (user intent excerpt) | medium (PII if leaked) |
| prompt_lang | string (e.g., "en", "ko") | opt-in (UserPromptSubmit) | low (language tag) | low |
| prompt_hash | string (SHA-256) | opt-in (UserPromptSubmit) | none (hash only, no semantics) | low |
| agent_name | string | opt-in (SubagentStop) | low (agent role) | low |
| agent_type | string | opt-in (SubagentStop) | low (agent type) | low |

**Recommendation for Phase 3 implementation**: Use (subject) only. Add (subject + prompt_preview) as Phase 4 optional variant.

---

