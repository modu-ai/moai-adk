# Tasks — SPEC-V3R4-HARNESS-003

This document is the task-level breakdown of the implementation plan for the Embedding-Cluster Classifier (Tier-2 Pattern Aggregation Upgrade). Tasks are organized by Wave (A through D) and use IDs `T-A1` through `T-D6`. Each task lists its linked REQ IDs, linked AC IDs, files affected, effort estimate (S/M/L per task complexity), dependency edge, test obligation per TDD methodology, MX tag implications, and risk/complexity rating.

All priorities use P0/P1/P2/P3 labels per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation; no time estimates are used.

---

## Task Summary

| Wave | Tasks | Priority Distribution | Total |
|------|-------|-----------------------|-------|
| Wave A (Stage-1 refactor + golden baseline) | T-A1, T-A2, T-A3, T-A4 | P0: 3, P1: 1 | 4 |
| Wave B (Stage-2 SimHash module) | T-B1, T-B2, T-B3, T-B4 | P0: 3, P1: 1 | 4 |
| Wave C (Stage-2 clustering + audit log) | T-C1, T-C2, T-C3, T-C4, T-C5, T-C6 | P0: 5, P1: 1 | 6 |
| Wave D (Config + perf + regression guards) | T-D1, T-D2, T-D3, T-D4, T-D5, T-D6 | P0: 4, P1: 2 | 6 |
| **Total** | | **P0: 15, P1: 5** | **20** |

---

## Wave A — Stage-1 Refactor (Aggregator Interface Seam)

### T-A1 (P0) — Generate golden Stage-1 input fixture

**Description**: Create `internal/harness/testdata/stage1_baseline.jsonl` containing 1000 JSONL event lines matching the existing `learner_test.go:53-74` synthetic pattern: 10 distinct `(event_type, subject, context_hash)` combos × 100 each. Each line MUST be a valid `Event` struct (HARNESS-002 schema) with `Timestamp`, `EventType`, `Subject`, `ContextHash`, `TierIncrement: 0`, `SchemaVersion: "v1"`. Optional HARNESS-002 fields may be absent (`omitempty`). Subject pool: `/moai plan`, `/moai run`, `/moai sync`, `manager-spec`, `manager-tdd`, `expert-backend`, `expert-frontend`, `Edit`, `Bash`, `Read`. Context hash pool: `h1` through `h10`.

**Linked REQs**: REQ-HRN-CLS-001, REQ-HRN-CLS-004

**Linked ACs**: AC-HRN-CLS-001

**Files affected**: `internal/harness/testdata/stage1_baseline.jsonl` (new, ~150 KB)

**Effort**: S (synthetic JSONL generator + commit)

**Dependencies**: None (foundational fixture)

**Test obligation**: RED — fixture must exist BEFORE T-A4 test passes. No test failure on fixture alone (it is data).

**MX tag implications**: None (test fixture file).

**Risk / Complexity**: Low.

**Definition of Done**:
- File exists with exactly 1000 JSONL lines.
- `wc -l internal/harness/testdata/stage1_baseline.jsonl` returns `1000`.
- Each line parses into `Event` struct via `json.Unmarshal`.
- The 10 distinct (event_type, subject, context_hash) combos each appear exactly 100 times.

---

### T-A2 (P0) — Generate golden Stage-1 output fixture

**Description**: After T-A1, run the EXISTING (pre-V3R4-003) `AggregatePatterns(testdata/stage1_baseline.jsonl)` and serialize the returned `map[string]*Pattern` to `internal/harness/testdata/stage1_baseline_patterns.json` using `json.MarshalIndent(_, "", "  ")`. This captures the byte-exact pattern map that AC-HRN-CLS-001 compares against forever.

**Linked REQs**: REQ-HRN-CLS-001

**Linked ACs**: AC-HRN-CLS-001

**Files affected**: `internal/harness/testdata/stage1_baseline_patterns.json` (new, ~5 KB)

**Effort**: S (one-shot generator script)

**Dependencies**: Blocked by T-A1.

**Test obligation**: GREEN gate — the file content captures the contract. Subsequent Stage-1 refactors MUST preserve this exact map.

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Definition of Done**:
- File exists with valid JSON (indented).
- Top-level object has exactly 10 keys.
- Each value has `Key`, `EventType`, `Subject`, `ContextHash`, `Count: 100`, `Confidence: 1.0` fields.
- A short comment in the file (or sibling README) documents how to regenerate it (e.g., a `_generate.go` build-tag-gated helper).

---

### T-A3 (P0) — Refactor `AggregatePatterns` to expose Stage-2 call site

**Description**: Modify `internal/harness/learner.go::AggregatePatterns` (around line 91, after the scanner.Err() check) to add a deferred Stage-2 invocation. The new call site MUST:
1. Read classifier config from `.moai/config/sections/harness.yaml` (via existing config loader). If `learning.classifier.stage_2_enabled` is `false` (default) OR absent OR invalid, SKIP Stage 2 (no-op).
2. If enabled, invoke `clusterSingletons(patterns, eventLookupFunc, classifierCfg)` from `classifier_cluster.go` (function will be added in Wave C; in Wave A it is a stub returning input unchanged).
3. Merge the returned map back into `patterns` (or use the returned map directly).
4. Return `patterns, nil` as before.

**The Stage-1 portion (lines 46-90) MUST NOT be modified in any way**. The new code is purely additive at the bottom of the function.

**Linked REQs**: REQ-HRN-CLS-001, REQ-HRN-CLS-004, REQ-HRN-CLS-011

**Linked ACs**: AC-HRN-CLS-001

**Files affected**: `internal/harness/learner.go` (modified; ~15-25 lines added at the end of AggregatePatterns)

**Effort**: M (refactor + config loader integration + stub function signature)

**Dependencies**: Blocked by T-A1, T-A2.

**Test obligation**: RED → GREEN — T-A4 test must initially fail (because Stage-2 stub is a no-op, baseline matches by default) … actually since the stub returns input unchanged, the test should pass IMMEDIATELY when stage_2_enabled=false. This is the GREEN-on-first-run case for Wave A. The proper RED test comes in Wave C when actual clustering kicks in and must NOT trigger for stage_2_enabled=false.

**MX tag implications**: Add `@MX:NOTE` to the new call site referencing SPEC-V3R4-HARNESS-003 REQ-HRN-CLS-001 (Stage-1 preservation contract).

**Risk / Complexity**: Medium. The seam placement is load-bearing for the rest of the SPEC.

**Definition of Done**:
- `git diff main -- internal/harness/learner.go` shows additions only at the END of `AggregatePatterns`, not in the scanner loop body.
- `go test ./internal/harness/...` no regressions (existing tests pass).
- `go vet ./internal/harness/...` passes.
- The new Stage-2 call site is gated by config (no panic if `clusterSingletons` is not yet defined; Wave C delivers it).

---

### T-A4 (P1) — Add `TestStage1BackwardCompat_StageDisabled` test

**Description**: Add a Go test to `internal/harness/learner_test.go` that:
1. Sets up a test fixture pointing to `testdata/stage1_baseline.jsonl`.
2. Configures `learning.classifier.stage_2_enabled: false` (or relies on default absence).
3. Invokes `AggregatePatterns(testdata/stage1_baseline.jsonl)`.
4. Loads `testdata/stage1_baseline_patterns.json` as the expected golden output.
5. Asserts `reflect.DeepEqual(actual, expected)` returns true.
6. Asserts `os.Stat(".moai/harness/cluster-merges.jsonl")` returns `IsNotExist` (no audit log emitted when Stage 2 OFF).

**Linked REQs**: REQ-HRN-CLS-001, REQ-HRN-CLS-004

**Linked ACs**: AC-HRN-CLS-001

**Files affected**: `internal/harness/learner_test.go` (modified; ~30 lines added)

**Effort**: S

**Dependencies**: Blocked by T-A1, T-A2, T-A3.

**Test obligation**: GREEN — test passes immediately after Wave A completes.

**MX tag implications**: None (test file).

**Risk / Complexity**: Low.

**Definition of Done**:
- `go test -run TestStage1BackwardCompat_StageDisabled ./internal/harness/...` passes.
- Test asserts `reflect.DeepEqual` returns true AND `os.Stat` returns `IsNotExist`.
- Test runs in < 1 second.

---

## Wave B — Stage-2 SimHash Module

### T-B1 (P0) — Implement `classifier_simhash.go` with `SimHash64` and `Hamming`

**Description**: Create `internal/harness/classifier_simhash.go` exposing two public functions and one helper:

```go
// Package harness — Stage-2 classifier SimHash fingerprinting module.
// REQ-HRN-CLS-002: Deterministic 64-bit fingerprint over semantic feature tuple.
// REQ-HRN-CLS-014: Excludes PromptContent from feature input (PII rule).
package harness

// SimHash64 computes a deterministic 64-bit SimHash fingerprint of the input.
// Algorithm: tokenize → for each token, hash via FNV-1a 64-bit → for each
// of 64 bit positions, accumulate +1 if bit is set, -1 if unset → final
// fingerprint bit is 1 if accumulator > 0, else 0.
func SimHash64(s string) uint64

// Hamming returns the number of bit positions where a and b differ.
func Hamming(a, b uint64) int

// tokenize splits s on whitespace + ASCII punctuation, lowercases each token,
// returns deduplicated slice (order-insensitive within token set).
func tokenize(s string) []string
```

Implementation guidance:
- Use Go stdlib `hash/fnv` (FNV-1a 64-bit, available in stdlib, zero new dependency).
- Tokenizer: split on whitespace + `[.,;:!?()[\]{}|/-]`; lowercase; deduplicate.
- SimHash core: 64-element int array; for each token's FNV hash, iterate 64 bits; +1 if set, -1 if unset.
- Hamming: `bits.OnesCount64(a ^ b)` from stdlib `math/bits`.

Reject all external SimHash library dependencies (Wave B decision — embed locally per plan.md §3.2 recommendation).

**Linked REQs**: REQ-HRN-CLS-002

**Linked ACs**: AC-HRN-CLS-002

**Files affected**: `internal/harness/classifier_simhash.go` (new, ~80 lines)

**Effort**: M

**Dependencies**: None (Wave B foundation).

**Test obligation**: RED — function exists but tests in T-B2 verify behavior. RED until T-B2 GREEN.

**MX tag implications**: Add `@MX:ANCHOR` to `SimHash64` (expected fan_in ≥ 2: classifier_cluster.go + tests). Add `@MX:NOTE` to file header citing REQ-HRN-CLS-014 (PII rule).

**Risk / Complexity**: Medium. SimHash math must be correct.

**Definition of Done**:
- File exists with the three identifiers.
- `go vet ./internal/harness/...` passes.
- `go build ./internal/harness/...` succeeds.
- File header comment cites SPEC-V3R4-HARNESS-003 REQ-HRN-CLS-002 and REQ-HRN-CLS-014.

---

### T-B2 (P0) — Add `classifier_simhash_test.go` table-driven tests

**Description**: Create `internal/harness/classifier_simhash_test.go` with the following table-driven tests:

```go
func TestSimHash64_Determinism(t *testing.T) {
  // Same input → same output across 100 invocations.
  for i := 0; i < 100; i++ {
    if SimHash64("hello world") != SimHash64("hello world") { t.Fatal(...) }
  }
}

func TestSimHash64_OrderInsensitivity(t *testing.T) {
  // Token reorder → same fingerprint (because tokenizer dedupes/sorts).
  if SimHash64("hello world") != SimHash64("world hello") { t.Errorf(...) }
}

func TestSimHash64_KnownInputs(t *testing.T) {
  // Hand-computed expected fingerprints for 3-5 short inputs.
  cases := []struct{ in string; expected uint64 }{
    {"a", 0x...},  // hand-computed
    {"a b", 0x...},
    {"", 0x...},   // empty input edge case
  }
  // ... iterate
}

func TestSimHash64_EmptyInput(t *testing.T) {
  // Empty string → deterministic value (probably 0 or all-bits-set).
  fp := SimHash64("")
  if SimHash64("") != fp { t.Fatal(...) }
}

func TestHamming_KnownPairs(t *testing.T) {
  cases := []struct{ a, b uint64; expected int }{
    {0, 0, 0},
    {0, 1, 1},
    {0xFFFFFFFFFFFFFFFF, 0, 64},
    {0xAAAAAAAAAAAAAAAA, 0x5555555555555555, 64},  // alternating bits
  }
}
```

**Linked REQs**: REQ-HRN-CLS-002

**Linked ACs**: AC-HRN-CLS-002

**Files affected**: `internal/harness/classifier_simhash_test.go` (new, ~120 lines)

**Effort**: M (table cases + hand-computation of expected fingerprints)

**Dependencies**: Blocked by T-B1.

**Test obligation**: RED → GREEN — tests written first; T-B1 implementation makes them green.

**MX tag implications**: None (test file).

**Risk / Complexity**: Medium. Hand-computing expected SimHash values for known inputs requires careful arithmetic; alternative: capture them once after T-B1 compiles and treat as golden values (acceptable in TDD as long as the algorithm is documented).

**Definition of Done**:
- `go test -run TestSimHash64 ./internal/harness/...` passes (all sub-cases).
- `go test -run TestHamming ./internal/harness/...` passes.
- Coverage of `classifier_simhash.go` ≥ 90%.

---

### T-B3 (P0) — Implement `buildFeatureString` excluding PromptContent

**Description**: Add to `classifier_simhash.go`:

```go
// featureFields is the canonical order of semantic features used for Stage-2
// fingerprinting. REQ-HRN-CLS-014: PromptContent is INTENTIONALLY EXCLUDED.
var featureFields = []string{
  "subject",
  "prompt_preview",
  "prompt_lang",
  "agent_name",
  "agent_type",
}

// buildFeatureString concatenates the selected feature values from evt in
// canonical order, separated by "|", suitable for SimHash64 input. The
// PromptContent field is structurally inaccessible to this function.
func buildFeatureString(evt *Event, fields []string) string {
  // Use a switch over `fields` literal to enforce the closed enumeration.
  // If `fields` contains "prompt_content", the case is unreachable — function
  // never references evt.PromptContent.
  var parts []string
  for _, f := range fields {
    switch f {
    case "subject":         parts = append(parts, "subject:"+evt.Subject)
    case "prompt_preview":  parts = append(parts, "prompt_preview:"+evt.PromptPreview)
    case "prompt_lang":     parts = append(parts, "prompt_lang:"+evt.PromptLang)
    case "agent_name":      parts = append(parts, "agent_name:"+evt.AgentName)
    case "agent_type":      parts = append(parts, "agent_type:"+evt.AgentType)
    // No case for "prompt_content" — REQ-HRN-CLS-014.
    }
  }
  return strings.Join(parts, "|")
}
```

**Linked REQs**: REQ-HRN-CLS-002, REQ-HRN-CLS-014

**Linked ACs**: AC-HRN-CLS-002, AC-HRN-CLS-009

**Files affected**: `internal/harness/classifier_simhash.go` (modified; ~30 lines added)

**Effort**: S

**Dependencies**: Blocked by T-B1.

**Test obligation**: GREEN — covered by T-B2 (TestBuildFeatureString_ExcludesPromptContent) and T-D3 (TestStage2NeverIncludesPromptContent).

**MX tag implications**: `@MX:NOTE` on `buildFeatureString` citing REQ-HRN-CLS-014.

**Risk / Complexity**: Low.

**Definition of Done**:
- Function exists with the closed switch over `fields`.
- `grep -nE 'PromptContent' internal/harness/classifier_simhash.go` returns 0 matches.
- Test case `TestBuildFeatureString_ExcludesPromptContent` in T-B2 verifies that even when `fields = ["prompt_content"]` is passed, the result does NOT contain the PromptContent string from the input Event.

---

### T-B4 (P1) — Add PII grep guard as Go test

**Description**: Add an in-process CI guard test that asserts `grep -nE 'PromptContent' internal/harness/classifier_simhash.go` returns zero matches. Implement as a Go test that reads the file and uses `strings.Contains`:

```go
func TestSimHashFile_HasNoPromptContentReference(t *testing.T) {
  data, err := os.ReadFile("classifier_simhash.go")
  // ... handle err
  if strings.Contains(string(data), "PromptContent") {
    t.Fatalf("REQ-HRN-CLS-014 violation: classifier_simhash.go contains 'PromptContent' reference")
  }
}
```

**Linked REQs**: REQ-HRN-CLS-014

**Linked ACs**: AC-HRN-CLS-009

**Files affected**: `internal/harness/classifier_simhash_test.go` (modified; ~15 lines added)

**Effort**: S

**Dependencies**: Blocked by T-B1, T-B3.

**Test obligation**: GREEN gate.

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Definition of Done**:
- `go test -run TestSimHashFile_HasNoPromptContentReference ./internal/harness/...` passes.
- Test fails with a clear diagnostic if the file ever contains `PromptContent`.

---

## Wave C — Stage-2 Clustering + Audit Log

### T-C1 (P0) — Define `ClassifierConfig` struct + defaults

**Description**: Add to `internal/harness/classifier_cluster.go` (new file):

```go
// ClassifierConfig holds Stage-2 tunable parameters per
// SPEC-V3R4-HARNESS-003 REQ-HRN-CLS-002/003/005/017.
type ClassifierConfig struct {
  Stage2Enabled       bool     // default false
  SimilarityAlgorithm string   // "simhash" | "none"; default "simhash"
  HammingThreshold    int      // [0, 64]; default 3 (research.md §8.3 recommended 2-5; looser 5-12 acceptable)
  ClusterMinSize      int      // >= 2; default 3
  FeatureFields       []string // default ["subject", "prompt_preview", "prompt_lang", "agent_name", "agent_type"]
}

// DefaultClassifierConfig returns the documented default config.
// REQ-HRN-CLS-018: invalid config falls back to this default.
func DefaultClassifierConfig() ClassifierConfig {
  return ClassifierConfig{
    Stage2Enabled:       false,
    SimilarityAlgorithm: "simhash",
    HammingThreshold:    3,
    ClusterMinSize:      3,
    FeatureFields:       []string{"subject", "prompt_preview", "prompt_lang", "agent_name", "agent_type"},
  }
}

// Validate returns nil if all fields are within allowed ranges, or an error
// describing the first violation. The caller MUST fail safe to
// DefaultClassifierConfig() on error per REQ-HRN-CLS-018.
func (c ClassifierConfig) Validate() error {
  if c.HammingThreshold < 0 || c.HammingThreshold > 64 { ... }
  if c.ClusterMinSize < 2 { ... }
  if c.SimilarityAlgorithm != "simhash" && c.SimilarityAlgorithm != "none" { ... }
  return nil
}
```

**Linked REQs**: REQ-HRN-CLS-002, REQ-HRN-CLS-003, REQ-HRN-CLS-016, REQ-HRN-CLS-018

**Linked ACs**: AC-HRN-CLS-014

**Files affected**: `internal/harness/classifier_cluster.go` (new, ~50 lines for this task)

**Effort**: S

**Dependencies**: Blocked by Wave B (uses `featureFields` from classifier_simhash.go).

**Test obligation**: RED → GREEN (tested by T-D5).

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Definition of Done**:
- Struct and constructor exist.
- `Validate()` covers the 3 documented ranges.
- File header cites SPEC-V3R4-HARNESS-003.

---

### T-C2 (P0) — Implement `clusterSingletons` core algorithm

**Description**: Add to `classifier_cluster.go`:

```go
// SingletonInput pairs a singleton Pattern with its source Event for Stage-2
// fingerprinting.
type SingletonInput struct {
  Pattern     *Pattern
  Event       *Event
  Fingerprint uint64
}

// Cluster holds a group of Stage-1 singletons that Stage 2 merges into one
// virtual Pattern.
type Cluster struct {
  EventType        EventType
  Members          []*SingletonInput
  HammingDistances []int // upper-triangle, row-major
  MergedKey        string
  MergedCount      int
  Confidence       float64
}

// clusterSingletons runs Stage 2 of the classifier. It accepts the Stage-1
// pattern map, an event-lookup function (to recover source Events for
// singletons), and a config. Returns an extended pattern map containing the
// original Stage-1 entries plus zero or more Stage-2 merged entries.
//
// Per REQ-HRN-CLS-001 / REQ-HRN-CLS-004: if cfg.Stage2Enabled is false OR
// cfg.SimilarityAlgorithm is "none", returns input map unchanged.
//
// Per REQ-HRN-CLS-018: if cfg.Validate() returns error, logs warning to
// stderr and falls back to Stage-1-only.
//
// @MX:ANCHOR: clusterSingletons is the Stage-2 entry point called from
// learner.go::AggregatePatterns.
func clusterSingletons(
  stage1 map[string]*Pattern,
  eventLookup func(key string) *Event,
  cfg ClassifierConfig,
  auditLogPath string,
) (map[string]*Pattern, error) {
  // 1. Validate config; fail safe to default
  if err := cfg.Validate(); err != nil {
    fmt.Fprintf(os.Stderr, "WARN: classifier config invalid (%v); Stage 2 disabled\n", err)
    return stage1, nil
  }
  if !cfg.Stage2Enabled || cfg.SimilarityAlgorithm == "none" {
    return stage1, nil
  }

  // 2. Identify singletons (Count == 1)
  singletons := []*SingletonInput{}
  for _, p := range stage1 {
    if p.Count == 1 {
      evt := eventLookup(p.Key)
      if evt == nil { continue }
      singletons = append(singletons, &SingletonInput{Pattern: p, Event: evt})
    }
  }
  if len(singletons) < cfg.ClusterMinSize {
    return stage1, nil // EDGE-002, EDGE-003
  }

  // 3. Partition by EventType (EDGE-008)
  groups := map[EventType][]*SingletonInput{}
  for _, s := range singletons {
    groups[s.Event.EventType] = append(groups[s.Event.EventType], s)
  }

  // 4. For each group, fingerprint + cluster
  result := stage1 // start with Stage-1 map; will append merges
  for et, group := range groups {
    if len(group) < cfg.ClusterMinSize { continue }
    for _, s := range group {
      featureStr := buildFeatureString(s.Event, cfg.FeatureFields)
      s.Fingerprint = SimHash64(featureStr)
    }
    clusters := pairwiseHammingCluster(group, cfg.HammingThreshold, cfg.ClusterMinSize)
    for _, c := range clusters {
      c.EventType = et
      merged := emitMergedPattern(&c) // populates MergedKey, MergedCount, Confidence
      result[merged.Key] = merged
      if err := appendClusterMergeAudit(auditLogPath, &c, merged); err != nil {
        fmt.Fprintf(os.Stderr, "WARN: audit log write failed: %v\n", err) // EDGE-006
      }
    }
  }
  return result, nil
}
```

Helper `pairwiseHammingCluster` uses Union-Find over `group` indices; for each pair (i, j) with `Hamming(group[i].Fingerprint, group[j].Fingerprint) <= threshold`, call `uf.Union(i, j)`. Extract components of size ≥ minSize.

Helper `emitMergedPattern` per REQ-HRN-CLS-005: deterministic lex-min subject choice, ContextHash="", Count=sum, Confidence=mean.

**Linked REQs**: REQ-HRN-CLS-002, REQ-HRN-CLS-003, REQ-HRN-CLS-005, REQ-HRN-CLS-012, REQ-HRN-CLS-018

**Linked ACs**: AC-HRN-CLS-002, AC-HRN-CLS-003, AC-HRN-CLS-004, AC-HRN-CLS-007, AC-HRN-CLS-014

**Files affected**: `internal/harness/classifier_cluster.go` (modified; ~200 lines added)

**Effort**: L (algorithm implementation + helpers)

**Dependencies**: Blocked by T-C1, Wave B (SimHash64, buildFeatureString).

**Test obligation**: RED until T-C5 GREEN.

**MX tag implications**: `@MX:ANCHOR` on `clusterSingletons`; `@MX:NOTE` on `emitMergedPattern` citing REQ-HRN-CLS-005.

**Risk / Complexity**: High. Algorithm correctness is load-bearing.

**Definition of Done**:
- All functions exist with documented signatures.
- `go vet ./internal/harness/...` passes.
- `go build ./internal/harness/...` succeeds.
- Confidence floor logic in `emitMergedPattern` defers to ClassifyTier (which already enforces the 0.70 threshold at learner.go:115); no duplicated check needed.

---

### T-C3 (P0) — Implement `appendClusterMergeAudit`

**Description**: Add to `classifier_cluster.go`:

```go
// clusterMergeAudit is the JSONL schema for cluster-merges.jsonl per
// REQ-HRN-CLS-013.
type clusterMergeAudit struct {
  Ts               time.Time `json:"ts"`
  MemberKeys       []string  `json:"member_keys"`
  MemberCounts     []int     `json:"member_counts"`
  HammingDistances []int     `json:"hamming_distances"` // upper-triangle, row-major
  MergedKey        string    `json:"merged_key"`
  MergedCount      int       `json:"merged_count"`
  Confidence       float64   `json:"confidence"`
}

// appendClusterMergeAudit appends one JSONL line to logPath documenting the
// cluster merge decision. Uses O_APPEND|O_CREATE|O_WRONLY semantics; safe
// under concurrent calls per POSIX append atomicity.
//
// Returns nil on success. On disk-full or permission error, returns the
// error so the caller can log non-blocking warning (EDGE-006).
func appendClusterMergeAudit(logPath string, cluster *Cluster, merged *Pattern) error {
  entry := clusterMergeAudit{
    Ts:               time.Now().UTC(),
    MemberKeys:       collectMemberKeys(cluster.Members),
    MemberCounts:     collectMemberCounts(cluster.Members),
    HammingDistances: collectHammingTriangle(cluster.Members),
    MergedKey:        merged.Key,
    MergedCount:      merged.Count,
    Confidence:       merged.Confidence,
  }
  data, err := json.Marshal(entry)
  if err != nil { return err }
  data = append(data, '\n')

  // Auto-create parent dir
  if dir := filepath.Dir(logPath); dir != "." && dir != "" {
    if err := os.MkdirAll(dir, 0o755); err != nil { return err }
  }

  f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
  if err != nil { return err }
  defer f.Close()
  _, err = f.Write(data)
  return err
}
```

The `collectHammingTriangle` helper iterates `(i, j)` for `i < j` and computes `Hamming(members[i].Fingerprint, members[j].Fingerprint)`.

**Linked REQs**: REQ-HRN-CLS-013

**Linked ACs**: AC-HRN-CLS-008

**Files affected**: `internal/harness/classifier_cluster.go` (modified; ~60 lines added)

**Effort**: M

**Dependencies**: Blocked by T-C2.

**Test obligation**: RED until T-C6 GREEN.

**MX tag implications**: `@MX:NOTE` on `appendClusterMergeAudit` (PII-sensitive write target per REQ-HRN-CLS-014; no PromptContent visible in this function's input — Cluster.Members carry Patterns, not raw Events with PromptContent).

**Risk / Complexity**: Medium.

**Definition of Done**:
- Function exists with the documented signature.
- JSONL line for a 4-member cluster has exactly 6 entries in `hamming_distances` (upper-triangle of 4×4 matrix = C(4,2) = 6).
- File mode is 0o644.
- Parent directory auto-created if absent.

---

### T-C4 (P0) — Implement `findEventByKey` event lookup

**Description**: Add to `classifier_cluster.go`:

```go
// findEventByKey re-reads usage-log.jsonl line by line and returns the FIRST
// Event whose three-field buildPatternKey matches `key`. Returns nil if no
// match found.
//
// Used by clusterSingletons to recover source Events for singleton patterns
// (needed for SimHash fingerprinting, which consumes the extended HARNESS-002
// fields not present in the Pattern struct).
//
// Performance: O(n) per call where n = events in log. For batch invocation
// with k singletons this is O(n*k) worst case. Optimization opportunity:
// build a key→Event map once during AggregatePatterns (deferred to follow-up).
func findEventByKey(logPath, key string) *Event {
  f, err := os.Open(logPath)
  if err != nil { return nil }
  defer f.Close()
  scanner := bufio.NewScanner(f)
  for scanner.Scan() {
    var evt Event
    if err := json.Unmarshal(scanner.Bytes(), &evt); err != nil { continue }
    if buildPatternKey(evt.EventType, evt.Subject, evt.ContextHash) == key {
      return &evt
    }
  }
  return nil
}
```

ALTERNATIVE: Refactor T-A3 to pass an event-by-key map through to `clusterSingletons` instead of a lookup function. Decision deferred to implementation discretion based on what minimizes diff churn.

**Linked REQs**: REQ-HRN-CLS-002 (Stage-2 needs source Events for fingerprinting)

**Linked ACs**: AC-HRN-CLS-002

**Files affected**: `internal/harness/classifier_cluster.go` (modified; ~25 lines added)

**Effort**: S

**Dependencies**: Blocked by T-C2.

**Test obligation**: Implicit (covered by T-C5 integration tests).

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Definition of Done**:
- Function exists with documented signature.
- Returns nil on file-not-found, empty file, or no match.

---

### T-C5 (P0) — Add `classifier_cluster_test.go` integration tests

**Description**: Create `internal/harness/classifier_cluster_test.go` with the following tests:

- `TestStage2Aggregates10SimilarSingletons` (AC-HRN-CLS-002):
  - Fixture: `testdata/stage2_similar_10.jsonl` (10 singleton `user_prompt` events with subjects `/moai plan SPEC-001` through `/moai plan SPEC-010`, prompt_lang="en", identical agent_name/type, AND identical `prompt_preview` content across all 10 events — the dominant shared SimHash feature; 64-byte string such as `"plan a new authentication SPEC with EARS requirements..."`).
  - Config: stage_2_enabled=true, hamming_threshold=3, cluster_min_size=3.
  - Assert: pattern map contains exactly 11 entries — 10 Stage-1 singletons + 1 merged entry with Key=`"user_prompt:/moai plan SPEC-001"`, Count=10, Confidence=1.0, Tier=TierAutoUpdate. Pairwise Hamming distance ≤ 3 across all `C(10,2) = 45` cluster pairs.

- `TestStage2DissimilarSingletonsStaySeparate` (AC-HRN-CLS-003):
  - Fixture: `testdata/stage2_dissimilar_10.jsonl` (10 semantically divergent singletons).
  - Assert: pattern map has exactly 10 entries (all original Stage-1 singletons); no entry with ContextHash="".

- `TestClusterMergeEmissionShape` (AC-HRN-CLS-004):
  - Synthetic 5-singleton cluster with Confidence values `[1.0, 1.0, 0.9, 0.8, 1.0]`.
  - Assert: emitted Pattern has Key=lex-min canonical form, Count=5, Confidence=0.94, Tier=TierRule.

- `TestConfidenceFloorForcesObservation` (AC-HRN-CLS-007):
  - Synthetic 10-singleton cluster with mean Confidence=0.69.
  - Assert: Tier=TierObservation despite Count=10.

- `TestStage2HonorsTierThresholdsOverride` (AC-HRN-CLS-013):
  - Config: tier_thresholds=[2,4,8,20].
  - Synthetic merged Pattern with Count=8.
  - Assert: ClassifyTier returns TierRule.

**Linked REQs**: REQ-HRN-CLS-002, REQ-HRN-CLS-003, REQ-HRN-CLS-005, REQ-HRN-CLS-006, REQ-HRN-CLS-012, REQ-HRN-CLS-017

**Linked ACs**: AC-HRN-CLS-002, AC-HRN-CLS-003, AC-HRN-CLS-004, AC-HRN-CLS-007, AC-HRN-CLS-013

**Files affected**: `internal/harness/classifier_cluster_test.go` (new, ~250 lines); `internal/harness/testdata/stage2_similar_10.jsonl` (new); `internal/harness/testdata/stage2_dissimilar_10.jsonl` (new)

**Effort**: L

**Dependencies**: Blocked by T-C1, T-C2, T-C3, T-C4.

**Test obligation**: RED → GREEN (drives Wave C implementation).

**MX tag implications**: None.

**Risk / Complexity**: High. Coverage breadth.

**Definition of Done**:
- All 5 tests pass with `go test -run 'TestStage2|TestClusterMerge|TestConfidenceFloor' ./internal/harness/...`.
- Fixtures committed.
- Coverage of `classifier_cluster.go` ≥ 85%.

---

### T-C6 (P1) — Add `classifier_cluster_audit_test.go`

**Description**: Create `internal/harness/classifier_cluster_audit_test.go` with:

- `TestClusterMergeAuditLogSchema` (AC-HRN-CLS-008):
  - Run Stage 2 on a synthetic 4-member cluster.
  - Read `.moai/harness/cluster-merges.jsonl` (in t.TempDir).
  - Assert: exactly 1 JSONL line; unmarshals into `clusterMergeAudit` struct; `ts` is valid ISO-8601; `member_keys` has 4 elements; `member_counts` has 4 elements all=1; `hamming_distances` has 6 elements (C(4,2)); `merged_count`=4.

- `TestAuditLogAppendOnly`:
  - Run Stage 2 twice in sequence.
  - Assert: file has total line count = sum of merges from both runs.
  - Pre-existing lines from run 1 are byte-identical in file after run 2.

**Linked REQs**: REQ-HRN-CLS-013

**Linked ACs**: AC-HRN-CLS-008

**Files affected**: `internal/harness/classifier_cluster_audit_test.go` (new, ~80 lines)

**Effort**: M

**Dependencies**: Blocked by T-C3.

**Test obligation**: GREEN.

**MX tag implications**: None.

**Risk / Complexity**: Medium.

**Definition of Done**:
- Both tests pass.
- Schema verification covers all 7 documented fields.

---

## Wave D — Config Wiring + End-to-End Test + Performance Budget

### T-D1 (P0) — Extend `harness.yaml` + config loader

**Description**: Edit `.moai/config/sections/harness.yaml` to add `learning.classifier` block:

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
    classifier:
        stage_2_enabled: false              # default OFF; opt-in
        similarity_algorithm: simhash       # or "none" to disable
        hamming_threshold: 3                # bits different; range [0, 64]; research.md §8.3 recommended 2-5; looser 5-12 acceptable
        cluster_min_size: 3                 # >= 2
        feature_fields:                     # PromptContent NEVER appears here
            - subject
            - prompt_preview
            - prompt_lang
            - agent_name
            - agent_type
```

Update the config loader (`internal/config/loader.go` or equivalent) to parse this block into a `ClassifierConfig` struct. On parse error or validation failure (per `ClassifierConfig.Validate()`), log a single stderr warning and fall back to `DefaultClassifierConfig()`.

Sub-tasks:
- T-D1.1 — define `learning.classifier` block in `harness.yaml` with documented defaults.
- T-D1.2 — extend config loader to unmarshal the block into `ClassifierConfig`.
- T-D1.3 — call `ClassifierConfig.Validate()` post-unmarshal; on error, WARN to stderr and overwrite with `DefaultClassifierConfig()`.
- T-D1.4 — emit one-line WARN format matching AC-HRN-CLS-014 stderr regex.
- T-D1.5 — handle `yaml.Unmarshal` type-mismatch errors (e.g., string or float supplied for an int-typed field such as `hamming_threshold: '3.5'` or `hamming_threshold: 3.5`): catch the YAML decoder error BEFORE `Validate()` runs, log a one-line WARN identifying the malformed key, and fall back to `DefaultClassifierConfig()`. Required by AC-HRN-CLS-014 sub-case 6.

**Linked REQs**: REQ-HRN-CLS-018

**Linked ACs**: AC-HRN-CLS-014 (sub-cases 1-6; sub-case 6 specifically blocked by T-D1.5)

**Files affected**: `.moai/config/sections/harness.yaml` (modified; ~8 lines added); `internal/config/loader.go` (modified; ~30 lines added including type-mismatch recovery)

**Effort**: M

**Dependencies**: Blocked by T-C1 (`ClassifierConfig` struct).

**Test obligation**: GREEN — covered by T-D5 (including new sub-case 6 for type mismatch).

**MX tag implications**: None.

**Risk / Complexity**: Medium.

**Definition of Done**:
- YAML file parses without error via existing harness config loader.
- New keys are accessible via the loader API.
- Defaults match `DefaultClassifierConfig()` when keys are absent.
- `yaml.Unmarshal` type errors (string or float for int field) trigger WARN + fallback to defaults, NOT propagated to caller.

---

### T-D2 (P0) — Add `BenchmarkClusterSingletons1k`

**Description**: Create `internal/harness/classifier_cluster_bench_test.go` with:

```go
func BenchmarkClusterSingletons1k(b *testing.B) {
  // Pre-load fixture
  logPath := "testdata/stage2_perf_1k.jsonl"
  stage1, _ := AggregatePatterns(logPath)  // not measured

  cfg := DefaultClassifierConfig()
  cfg.Stage2Enabled = true

  durations := make([]time.Duration, b.N)
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    start := time.Now()
    _, _ = clusterSingletons(stage1, func(k string) *Event { return findEventByKey(logPath, k) }, cfg, b.TempDir()+"/cluster-merges.jsonl")
    durations[i] = time.Since(start)
  }
  b.StopTimer()

  // Compute p50, p95, p99
  sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
  p50 := durations[len(durations)*50/100]
  p95 := durations[len(durations)*95/100]
  p99 := durations[len(durations)*99/100]
  b.Logf("p50=%v p95=%v p99=%v", p50, p95, p99)
  // Informational only; do NOT b.Fatal on p99 > 25ms (AC-HRN-CLS-005 is informational gate).
}
```

Generate `testdata/stage2_perf_1k.jsonl`: 1000 events with mixed pattern repeats and singletons. Singleton count target: 30-50 (typical workload per research.md §3.3).

**Linked REQs**: REQ-HRN-CLS-010

**Linked ACs**: AC-HRN-CLS-005

**Files affected**: `internal/harness/classifier_cluster_bench_test.go` (new, ~80 lines); `internal/harness/testdata/stage2_perf_1k.jsonl` (new, ~150 KB)

**Effort**: M

**Dependencies**: Blocked by Wave C.

**Test obligation**: Benchmark report; informational.

**MX tag implications**: None.

**Risk / Complexity**: Medium.

**Definition of Done**:
- Benchmark compiles and runs.
- Report logs p50/p95/p99 to stdout.
- Documentation in test comment: regenerate fixture via T-D2 helper.

---

### T-D3 (P0) — Add `classifier_pii_test.go`

**Description**: Create `internal/harness/classifier_pii_test.go` with:

```go
func TestStage2NeverIncludesPromptContent(t *testing.T) {
  tmp := t.TempDir()
  logPath := filepath.Join(tmp, "usage-log.jsonl")
  auditPath := filepath.Join(tmp, "cluster-merges.jsonl")

  // Write 5 singleton events with PromptContent containing sentinel
  sentinel := "my-secret-api-key-DO-NOT-LEAK-12345"
  for i := 0; i < 5; i++ {
    evt := Event{
      Timestamp:     time.Now(),
      EventType:     EventTypeUserPrompt,
      Subject:       fmt.Sprintf("prompt-%d", i),
      ContextHash:   fmt.Sprintf("hash-%d", i),
      PromptContent: sentinel,
      PromptPreview: fmt.Sprintf("preview-%d", i),
      PromptLang:    "en",
    }
    appendJSONL(logPath, evt)
  }

  stage1, _ := AggregatePatterns(logPath)
  cfg := DefaultClassifierConfig()
  cfg.Stage2Enabled = true
  cfg.HammingThreshold = 64 // force all to cluster

  _, _ = clusterSingletons(stage1, func(k string) *Event { return findEventByKey(logPath, k) }, cfg, auditPath)

  data, _ := os.ReadFile(auditPath)
  if strings.Contains(string(data), sentinel) {
    t.Fatalf("REQ-HRN-CLS-014 violation: sentinel %q found in cluster-merges.jsonl", sentinel)
  }
}

func TestClusterFile_HasNoPromptContentReference(t *testing.T) {
  // Static guard: classifier source files do NOT reference PromptContent
  for _, f := range []string{"classifier_simhash.go", "classifier_cluster.go"} {
    data, err := os.ReadFile(f)
    if err != nil { t.Fatalf(...) }
    if strings.Contains(string(data), "PromptContent") {
      t.Fatalf("REQ-HRN-CLS-014 violation: %s contains 'PromptContent' reference", f)
    }
  }
}
```

**Linked REQs**: REQ-HRN-CLS-014

**Linked ACs**: AC-HRN-CLS-009

**Files affected**: `internal/harness/classifier_pii_test.go` (new, ~80 lines)

**Effort**: M

**Dependencies**: Blocked by Wave C.

**Test obligation**: GREEN gate.

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Definition of Done**:
- Both tests pass.
- Runtime test confirms sentinel absence in audit log.
- Static test confirms source files do not reference `PromptContent`.

---

### T-D4 (P0) — Add three regression-guard tests

**Description**: Create three test files:

**File 1**: `internal/harness/classifier_schema_regression_test.go::TestPromotionAndProposalSchemaPreserved` (AC-HRN-CLS-011):
- Fixture: `testdata/legacy_promotions.jsonl` (3 pre-V3R4-003 entries with 3-field pattern_keys).
- Assert: each entry unmarshals into `Promotion` struct successfully.
- Plus: simulate a Stage-2 merged-Pattern promotion; assert it writes a 2-field pattern_key entry to the same file; both old and new entries coexist parseably.

**File 2**: `internal/harness/classifier_frozen_guard_regression_test.go::TestFrozenGuardUnaffectedByStage2` (AC-HRN-CLS-006):
- Assert `EnsureAllowed(".claude/skills/moai-meta-harness/SKILL.md")` returns `*FrozenViolationError`.
- Assert `allowedPrefixes` and `frozenPrefixes` byte-identical to pre-V3R4-003 via static const reference check (impossible at runtime; instead, document in comment that `git diff main -- internal/harness/frozen_guard.go` must be empty).

**File 3**: `internal/harness/classifier_rate_limit_test.go::TestStage2DoesNotInflateRateLimit` (AC-HRN-CLS-010):
- Integration scenario: synthesize 100-singleton cluster yielding 1 merged Pattern.
- Pre-seed rate-limit state showing 1 application within last 7 days.
- Assert: rate limiter denies subsequent Tier-4 candidate (existing CheckLimit behavior preserved; not testing CheckLimit itself but verifying integration).

**Linked REQs**: REQ-HRN-CLS-007, REQ-HRN-CLS-008, REQ-HRN-CLS-009, REQ-HRN-CLS-015

**Linked ACs**: AC-HRN-CLS-006, AC-HRN-CLS-010, AC-HRN-CLS-011

**Files affected**: Three new test files (~180 lines total); `internal/harness/testdata/legacy_promotions.jsonl` (new, ~600 bytes)

**Effort**: L

**Dependencies**: Blocked by Wave C.

**Test obligation**: GREEN gates.

**MX tag implications**: None.

**Risk / Complexity**: Medium.

**Definition of Done**:
- All three test files compile and pass.
- Legacy fixture committed.

---

### T-D5 (P0) — Add `TestInvalidConfigFailsSafeToStage1`

**Description**: Add to `classifier_cluster_test.go`:

```go
func TestInvalidConfigFailsSafeToStage1(t *testing.T) {
  cases := []struct{
    name string
    cfg  ClassifierConfig
    expectedWarnSubstring string
  }{
    {"hamming_threshold_negative", ClassifierConfig{Stage2Enabled: true, HammingThreshold: -5, ClusterMinSize: 3, SimilarityAlgorithm: "simhash", FeatureFields: defaultFields}, "hamming_threshold"},
    {"hamming_threshold_oversized", ClassifierConfig{Stage2Enabled: true, HammingThreshold: 99, ClusterMinSize: 3, SimilarityAlgorithm: "simhash", FeatureFields: defaultFields}, "hamming_threshold"},
    {"cluster_min_size_too_low", ClassifierConfig{Stage2Enabled: true, HammingThreshold: 3, ClusterMinSize: 1, SimilarityAlgorithm: "simhash", FeatureFields: defaultFields}, "cluster_min_size"},
    {"algorithm_tfidf", ClassifierConfig{Stage2Enabled: true, HammingThreshold: 3, ClusterMinSize: 3, SimilarityAlgorithm: "tfidf", FeatureFields: defaultFields}, "similarity_algorithm"},
    {"algorithm_garbage", ClassifierConfig{Stage2Enabled: true, HammingThreshold: 3, ClusterMinSize: 3, SimilarityAlgorithm: "garbage_value_xyz", FeatureFields: defaultFields}, "similarity_algorithm"},
  }
  for _, c := range cases {
    t.Run(c.name, func(t *testing.T) {
      var stderr bytes.Buffer
      // ... redirect stderr to buf, run clusterSingletons with c.cfg ...
      // Assert: pattern map equals Stage-1 baseline (no merges)
      // Assert: cluster-merges.jsonl does not exist
      // Assert: stderr contains c.expectedWarnSubstring
      // Assert: no error returned
    })
  }
}

// Sub-case 6 (AC-HRN-CLS-014 sub-case 6): YAML type mismatch — float/string for int field.
// Exercises the LOADER path (yaml.Unmarshal error caught BEFORE Validate()).
func TestInvalidConfigYamlTypeMismatchFailsSafe(t *testing.T) {
  yamlVariants := []struct{
    name string
    yaml string  // raw YAML snippet with type-mismatched hamming_threshold
  }{
    {"float_value",  "learning:\n  classifier:\n    stage_2_enabled: true\n    hamming_threshold: 3.5\n"},
    {"string_value", "learning:\n  classifier:\n    stage_2_enabled: true\n    hamming_threshold: '3.5'\n"},
  }
  for _, v := range yamlVariants {
    t.Run(v.name, func(t *testing.T) {
      // Write v.yaml to a temp harness.yaml; invoke the loader.
      // Assert: loader returns DefaultClassifierConfig() (with HammingThreshold == 3).
      // Assert: stderr contains substring "hamming_threshold" AND ("type" OR "yaml").
      // Assert: loader does NOT propagate the yaml.Unmarshal error to the caller.
    })
  }
}
```

**Linked REQs**: REQ-HRN-CLS-016, REQ-HRN-CLS-018

**Linked ACs**: AC-HRN-CLS-014 (sub-cases 1-5 in `TestInvalidConfigFailsSafeToStage1`; sub-case 6 in `TestInvalidConfigYamlTypeMismatchFailsSafe`)

**Files affected**: `internal/harness/classifier_cluster_test.go` (modified; ~60 lines added) AND `internal/config/loader_test.go` (modified; ~30 lines added for sub-case 6)

**Effort**: M

**Dependencies**: Blocked by T-C2 (clusterSingletons fail-safe path), T-D1 (config defaults), AND T-D1.5 (yaml type-mismatch loader recovery).

**Test obligation**: GREEN gate (6 sub-cases — 5 in `TestInvalidConfigFailsSafeToStage1` + 1 in `TestInvalidConfigYamlTypeMismatchFailsSafe`).

**MX tag implications**: None.

**Risk / Complexity**: Medium.

**Definition of Done**:
- Test passes all 6 sub-cases (5 in-range validation + 1 yaml type mismatch).
- Stderr capture mechanism works portably (use `cmd.ErrOrStderr` injection where applicable).

---

### T-D6 (P1) — Update CHANGELOG.md + release notes

**Description**:
- Edit `CHANGELOG.md` to add a new section under `## [Unreleased]` or `## [v2.22.0]` containing:
  ```
  ### Added
  - Embedding-Cluster Classifier (Tier-2 Pattern Aggregation Upgrade) via SPEC-V3R4-HARNESS-003. Opt-in via `learning.classifier.stage_2_enabled` (default false). SimHash-based singleton clustering with Hamming-distance threshold (default 3 of 64 bits; research.md §8.3 recommended 2-5). Preserves Stage-1 backward compatibility byte-identically; preserves 5-Layer Safety, FROZEN zone, 7-day Tier-4 rate-limit floor. Audit log at `.moai/harness/cluster-merges.jsonl`.
  ```
- Create `.moai/release/RELEASE-NOTES-v2.22.0.md` with a section titled "Embedding-Cluster Classifier (Tier-2 Aggregation Upgrade)" explaining:
  - What it does (semantic clustering of singleton patterns)
  - How to opt in (`stage_2_enabled: true`)
  - Recommended tuning (`hamming_threshold: 3` default per research.md §8.3 range 2-5; raise to 5-12 for projects with weaker semantic signal)
  - Privacy preservation (PromptContent never enters classifier)
  - Backward compatibility (Stage-1-only behavior preserved when flag is false)

**Linked REQs**: REQ-HRN-CLS-001 (backward-compat documentation), REQ-HRN-CLS-014 (PII rule documentation)

**Linked ACs**: (Documentation; no direct AC.)

**Files affected**: `CHANGELOG.md` (modified); `.moai/release/RELEASE-NOTES-v2.22.0.md` (new or appended)

**Effort**: S

**Dependencies**: All Wave A-D tasks essentially complete (documentation reflects final shape).

**Test obligation**: None (documentation).

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Definition of Done**:
- CHANGELOG.md has the entry.
- Release notes file exists with the four documented subsections.
- Both files reference `SPEC-V3R4-HARNESS-003`.

---

## Dependency Graph (Task-Level)

```
Wave A:
  T-A1 (fixture input)
    └─ T-A2 (fixture output)
         └─ T-A3 (refactor AggregatePatterns) ──┐
              └─ T-A4 (Stage-1 backward-compat test) │
                                                      │
Wave B:                                               │
  T-B1 (SimHash64 + Hamming) ──┐                      │
       └─ T-B2 (SimHash tests) │                      │
       └─ T-B3 (buildFeatureString — REQ-HRN-CLS-014) │
       └─ T-B4 (PII grep guard test)                  │
                                                      │
Wave C (depends on Wave A + B):                       │
  T-C1 (ClassifierConfig) ─┐                          │
       └─ T-C2 (clusterSingletons core) ─┐            │
            ├─ T-C3 (appendClusterMergeAudit)         │
            └─ T-C4 (findEventByKey)                  │
                 └─ T-C5 (cluster tests AC-002/003/004/007/013)
                 └─ T-C6 (audit log schema test AC-008)
                                                      │
Wave D (depends on Wave C):                           │
  T-D1 (config loader extension) ─┐                   │
       └─ T-D5 (TestInvalidConfigFailsSafeToStage1)   │
  T-D2 (benchmark AC-005)                             │
  T-D3 (PII guard tests AC-009)                       │
  T-D4 (3 regression guards: AC-006/010/011)          │
  T-D6 (CHANGELOG + release notes) ←─ depends on all above
```

Critical path: T-A1 → T-A2 → T-A3 → T-B1 → T-B3 → T-C1 → T-C2 → T-C3 → T-C5 → T-D6.

Parallelizable: Wave B and Wave A T-A4 are independent. Within Wave C, T-C3/T-C4 are parallel after T-C2. Within Wave D, T-D2/T-D3/T-D4 are parallel after Wave C completes.

---

## Out of Scope (Task-Level)

The following tasks are explicitly NOT in any Wave of this SPEC. They are deferred to downstream SPECs or follow-up cleanup:

| Downstream SPEC / Cleanup | Task Domain |
|---------------------------|-------------|
| SPEC-V3R4-HARNESS-004 | Reflexion self-critique loop integration with Stage-2 merged Patterns. |
| SPEC-V3R4-HARNESS-005 | Principle-based scoring of Stage-2 merged Patterns. |
| SPEC-V3R4-HARNESS-006 | Multi-objective effectiveness scoring tuple, auto-rollback on regression. |
| SPEC-V3R4-HARNESS-007 | Voyager skill library; embedding-indexed retrieval. |
| SPEC-V3R4-HARNESS-008 | Cross-project federation of merged Patterns. |
| Follow-up performance SPEC | Optimization beyond 25ms p99 if observed in production; replace pairwise Hamming with locality-sensitive hashing or bit-bucket index. |
| Follow-up algorithm SPEC | TF-IDF / character n-gram alternative fingerprinting if SimHash produces unacceptable false positives. |
| Follow-up tooling SPEC | CLI verb to inspect cluster-merges.jsonl and disagree with clusters (Phase 4 inspection tool). |
| Follow-up migration SPEC | Re-aggregate historical tier-promotions.jsonl entries under Stage-2 schema (if desired). |
| Post-merge `manager-git` commit | None for this SPEC (no superseded SPECs; HARNESS-001/002 are dependencies, not superseded). |

---

## Run-Phase Entry Point

After this plan PR merges:

1. Execute `/clear` to reset context.
2. Create a SPEC worktree: `moai worktree new SPEC-V3R4-HARNESS-003 --base origin/main`.
3. Inside the worktree, execute `/moai run SPEC-V3R4-HARNESS-003`.
4. `manager-tdd` (or `manager-develop` depending on `quality.yaml` `development_mode`) takes over execution of Waves A-D sequentially.
5. Each Wave merges as a separate squash PR per Enhanced GitHub Flow doctrine (`CLAUDE.local.md` § 18).
6. After all four Waves merge, execute `/moai sync SPEC-V3R4-HARNESS-003` to generate the final documentation sync PR.
7. Update SPEC status: `draft` → `implemented` (sync-phase responsibility).

---

End of tasks.md.
