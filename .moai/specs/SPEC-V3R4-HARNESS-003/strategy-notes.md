# Strategy Synthesis — SPEC-V3R4-HARNESS-003

> Produced by manager-strategy (Phase 1 of /moai run) on plan PR `d32bb674a` merged main.
> Run-phase audit (iter 3 @ 0.965) PASSED with zero must-pass defects.
> User-confirmed envelope: harness=thorough, mode=autopilot, Wave A standalone → user checkpoint → Waves B-E continuous.

---

## 1. Plan Validation

The plan is **self-consistent and execution-ready**. spec-compact.md (18 REQ + 14 AC), plan.md (4-wave decomposition with §2.3 algorithm pseudocode and §5 MX-12-action plan), and tasks.md (20 atomic tasks T-A1..T-D6 with explicit DoD and dependency graph) all converge.

Three minor drift observations (non-blocking; document as assumptions, do NOT fix):

- **Wave count drift**: plan.md §1 ("four Waves A-D") vs user framing ("5 Waves A→E"). Reconciliation: **Wave E = T-D6 docs split-off** (CHANGELOG + RELEASE-NOTES), leaving Wave D pure-code (T-D1..T-D5). Administrative split only — no code impact.
- **tasks.md §3.1 "5 tasks" vs actual 4**: User prompt mentions 5 Wave-A tasks; tasks.md lists 4 (T-A1..T-A4). Treat user count as approximation; **Wave A = 4 tasks**.
- **AC-HRN-CLS-014 sub-case 6**: T-D1.5 + T-D5's `TestInvalidConfigYamlTypeMismatchFailsSafe` correctly cover loader-path type errors. No drift.

Code surface confirmed via grep:
- `buildPatternKey` (learner.go:99) ✓
- `AggregatePatterns` (L46) ✓
- `ClassifyTier` (L113) ✓
- `confidenceThreshold = 0.70` (L17) ✓
- HARNESS-002 fields in `types.go`: PromptPreview (L114), PromptContent (L119), AgentName (L88), AgentType (L91), PromptLang (L109) ✓
- No `classifier_*.go` files exist yet — clean slate for new code.

---

## 2. Wave A Execution Strategy (USER CHECKPOINT WAVE)

**Goal**: Establish golden Stage-1 baseline + insert no-op Stage-2 seam in `AggregatePatterns` with zero behavior change.

### 2.1 Task List

| # | Task | File | Op |
|---|------|------|-----|
| 1 | T-A1 | `internal/harness/testdata/stage1_baseline.jsonl` | NEW (~150KB, 1000 JSONL lines: 10 combos × 100) |
| 2 | T-A2 | `internal/harness/testdata/stage1_baseline_patterns.json` | NEW (~5KB, indented JSON, 10 keys) |
| 3 | T-A3 | `internal/harness/learner.go` | MOD (~15-25 LOC appended after L91 `scanner.Err()` check) |
| 4 | T-A4 | `internal/harness/learner_test.go` | MOD (~30 LOC added) |

Wave A also introduces stubs (so build never breaks at intermediate state):
- **Stub** `ClassifierConfig` struct in `internal/harness/classifier_cluster.go` with `Stage2Enabled bool` (defaults false) — fleshed out in Wave C (T-C1).
- **Stub** `clusterSingletons(...)` returning input unchanged — implemented in Wave C (T-C2).

### 2.2 TDD Cycle Sketch

| Cycle | Step | Detail |
|-------|------|--------|
| Pre-RED | Fixtures | T-A1 writes JSONL via test helper. T-A2 invokes **pre-modification** `AggregatePatterns` and `json.MarshalIndent` to capture truth BEFORE any learner.go change. |
| RED | T-A4 first | `TestStage1BackwardCompat_StageDisabled`: `reflect.DeepEqual(actual, golden)` + `os.Stat(audit-log) == IsNotExist`. RED because stubs don't exist yet (compile error). |
| GREEN | T-A3 + stubs | Add `ClassifierConfig` + `clusterSingletons` no-op stub. Modify `AggregatePatterns` to add trailing call site gated by `cfg.Stage2Enabled`. Test passes (stub returns input → DeepEqual succeeds). |
| REFACTOR | MX | Add @MX:NOTE to seam citing REQ-HRN-CLS-001. Verify `git diff` shows additions only AFTER scanner loop body, never inside it. |

Golden fixture generation: implement T-A2 as either (a) `_test.go` gated by `MOAI_REGEN_GOLDEN=1` env var, or (b) separate `testdata/_generate.go` with `//go:build ignore`.

### 2.3 AC Coverage (Wave A)

| AC | Pattern of Evidence |
|----|---------------------|
| **AC-HRN-CLS-001** (primary) | `TestStage1BackwardCompat_StageDisabled` proves byte-identical map + zero audit log. |
| AC-HRN-CLS-004 (implicit) | Same test exercises `stage_2_enabled: false` path. |

### 2.4 Risks (Wave A-specific)

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Golden fixture non-determinism (timestamps, map iteration) | Medium | High | Synthetic Events with `time.Time{}` zero value. Sort keys before `json.MarshalIndent`. Verify across 5 `go test -count=5` runs. |
| Stage-2 seam placed inside scanner loop | Low | High | T-A3 DoD: `git diff` shows additions **only AFTER `scanner.Err()` at L91**. |
| Circular import between learner.go and classifier_cluster.go | Low | Medium | Both in `package harness` — same package, no import cycle. |
| `os.Stat` audit-log path resolution differs by cwd | Medium | Low | Test uses `filepath.Join(t.TempDir(), ".moai/harness/cluster-merges.jsonl")` and passes path explicitly to `clusterSingletons`. |
| `AggregatePatterns` signature change breaks existing tests | Low | Medium | T-A3 preserves `AggregatePatterns(logPath string) (map[string]*Pattern, error)` signature. Config loading happens inside function. |

### 2.5 Exit Criterion (Wave A complete)

All commands must succeed:

```bash
# 1. Compilation
go build ./internal/harness/...

# 2. New test passes
go test -run TestStage1BackwardCompat_StageDisabled ./internal/harness/ -v

# 3. Full harness suite (no regressions)
go test ./internal/harness/...

# 4. Vet + race
go vet ./internal/harness/...
go test -race ./internal/harness/ -short

# 5. Static guards
test $(grep -c 'classifier_simhash\|classifier_cluster' internal/cli/hook.go internal/harness/observer.go) -eq 0
test $(grep -c PromptContent internal/harness/classifier_cluster.go 2>/dev/null) -eq 0

# 6. Golden fixture sanity
test $(wc -l < internal/harness/testdata/stage1_baseline.jsonl) -eq 1000
test $(jq 'keys | length' internal/harness/testdata/stage1_baseline_patterns.json) -eq 10
```

Assertion checklist:
- `learner.go` diff shows additions only after `scanner.Err()` block.
- `classifier_cluster.go` exists with stubs.
- `learner_test.go` has new test (~30 LOC).
- Two testdata files exist.
- `cluster-merges.jsonl` does NOT exist anywhere.

**User checkpoint**: Report PR diff summary → await user "proceed" before Waves B-E.

---

## 3. Waves B-E Condensed Strategy

### 3.1 Wave B (SimHash) — P0

**Tasks**: T-B1 (SimHash64+Hamming+tokenize, ~80 LOC), T-B2 (table tests, ~120 LOC), T-B3 (closed-switch buildFeatureString excluding PromptContent, ~30 LOC), T-B4 (PII grep guard, ~15 LOC).

**Files**: `classifier_simhash.go` (NEW), `classifier_simhash_test.go` (NEW).

**AC**: AC-HRN-CLS-002 + partial AC-HRN-CLS-009.

**Risks**: SimHash math error (mitigated by hand-computed table cases); hash choice settled (FNV-1a 64-bit stdlib).

### 3.2 Wave C (Cluster + Audit) — P0

**Tasks**: T-C1 (full ClassifierConfig + Validate), T-C2 (algorithm: EventType partition + Union-Find Hamming, ~200 LOC), T-C3 (appendClusterMergeAudit with hamming_distances cap-at-20 + truncated flag, ~60 LOC), T-C4 (findEventByKey, ~25 LOC), T-C5 (5 integration tests AC-002/003/004/007/013, ~250 LOC + 2 fixtures), T-C6 (audit schema test AC-008, ~80 LOC).

**Files**: `classifier_cluster.go` (extends Wave A stub), `classifier_cluster_test.go` (NEW), `classifier_cluster_audit_test.go` (NEW), 2 new testdata JSONLs.

**AC**: AC-HRN-CLS-002, -003, -004, -007, -008, -013.

**Cross-wave dep**: Wave B (SimHash64) + Wave A (stubs).

**Risks**: Union-Find correctness (hand-crafted fixtures); hamming_distances cap-at-20 with `truncated: true` flag (REQ-HRN-CLS-013 D3.2); lex-min subject canonicalization in emitMergedPattern.

### 3.3 Wave D (Bench + Config + Regression) — P0 + P1

**Tasks**: T-D1 (harness.yaml learning.classifier block + loader with T-D1.5 yaml type-error recovery), T-D2 (BenchmarkClusterSingletons1k informational, P1), T-D3 (PII guards AC-009), T-D4 (3 regression tests: Promotion AC-011 + FROZEN AC-006 + rate-limit AC-010), T-D5 (6-sub-case TestInvalidConfigFailsSafeToStage1 + TestInvalidConfigYamlTypeMismatchFailsSafe AC-014).

**Files**: `harness.yaml` (MOD), `loader.go` (MOD), 5 new test files, 2 new testdata files.

**AC**: AC-HRN-CLS-005, -006, -009, -010, -011, -014.

**Cross-wave dep**: Wave C deliverables + Wave B PII grep targets.

**Risks**: yaml type-mismatch loader path (T-D1.5 highest risk — must catch yaml.TypeError BEFORE Validate()); perf miss informational only; legacy_promotions.jsonl must have pre-V3R4-003 3-field keys.

### 3.4 Wave E (Docs) — P1

**Tasks**: T-D6 (CHANGELOG v2.22.0 "Added" + RELEASE-NOTES-v2.22.0.md with 4 subsections: what/opt-in/tuning/privacy-BC).

**Files**: `CHANGELOG.md` (MOD), `.moai/release/RELEASE-NOTES-v2.22.0.md` (NEW).

**Risks**: Trivial.

### 3.5 Cross-Wave Dependencies + Quality Checkpoints

Critical path: T-A1 → T-A2 → T-A3 → T-B1 → T-B3 → T-C1 → T-C2 → T-C3 → T-C5 → T-D6.

Parallelizable: T-B2/B3/B4 (after T-B1); T-C3/T-C4 (after T-C2); T-C5/T-C6 (after T-C3+T-C4); T-D2/T-D3/T-D4 (after Wave C).

Quality checkpoints (harness=thorough):

| Boundary | Gate |
|----------|------|
| End-of-A | Stage-1 BC test green + zero diff in scanner loop → user checkpoint |
| End-of-B | All SimHash tests green + PII grep returns 0 |
| End-of-C | 6 cluster tests + audit schema green → evaluator-active sprint contract on REQ-HRN-CLS-013 cap-at-20 |
| End-of-D | All 14 AC tests + benchmark logs + 0 PromptContent grep → evaluator-active iter 2 PASS (≥ 0.80) |
| End-of-E | CHANGELOG merged + RELEASE-NOTES exists |

---

## 4. Risk Register (Top 5)

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|------------|--------|------------|
| 1 | Golden fixture non-determinism makes AC-HRN-CLS-001 flaky | Medium | High | Synthetic Events `time.Time{}` zero + sorted map keys + 5-count verification |
| 2 | PII leak via PromptContent into classifier artifact | Low | Critical | 3-layer defense: closed-switch buildFeatureString, static grep, runtime sentinel test |
| 3 | FROZEN zone modification (observer.go/types.go/frozen_guard.go/safety/hook.go) | Low | Critical | Static `git diff` guards + T-D4 runtime assertion + evaluator-active independent verify |
| 4 | p99 perf budget miss (>25ms) | Medium | Low | Informational only (NOT CI-blocking). Mitigation: build key→Event map once in clusterSingletons. Follow-up SPEC if needed. |
| 5 | Config loader yaml type-mismatch propagates error instead of falling back to defaults | Medium | High | T-D1.5 catches yaml.TypeError BEFORE Validate(). T-D5 sub-case 6 covers float `3.5` + string `'3.5'`. Stderr must match `hamming_threshold` AND (`type` OR `yaml`). |

---

## 5. Effort Estimate (Priority labels, no time)

| Wave | Priority | Relative LOC | Tasks | Notes |
|------|----------|--------------|-------|-------|
| A | P0 | ~100 LOC + 150KB fixture | 4 | Load-bearing seam + golden baseline. User checkpoint after. |
| B | P0 | ~230 LOC | 4 | Pure-Go SimHash. Test-heavy. |
| C | P0 | ~600 LOC + 2 fixtures | 6 | Heaviest wave; algorithm correctness load-bearing. |
| D | P0 (+ P1 bench) | ~400 LOC + 2 fixtures | 5 | Config + perf + 5 regression guards. |
| E | P1 | ~30 LOC | 1 | Doc-only. |

Total: ~1,360 LOC + ~5 fixtures across 13 new + 3 modified files. P0=18 tasks, P1=2 tasks.

---

## 6. MX Tag Plan Confirmation

plan.md §5 has 12 actions; all align with Waves. **5 P1 blocking tags** (must exist before Phase 3):

1. `SimHash64` ANCHOR (T-B1)
2. `clusterSingletons` ANCHOR (T-C2)
3. `clusterSingletons` WARN on O(s²) (T-C2)
4. `buildFeatureString` NOTE — PII guard (T-B3)
5. `appendClusterMergeAudit` NOTE — audit log target (T-C3)

3 P2 NOTE tags (`tokenize`, `buildRepresentativeKey`, `Validate`) are recommended, not blocking.

Phase 2.9 quality gate verifies presence; Phase 3.5 MX-Injection applies.

---

## 7. Execution Recommendation

**Verdict: PROCEED**.

Rationale:
- Plan execution-ready (audit iter 3 @ 0.965, zero must-pass defects).
- Code surface verified via direct grep — all symbols at documented line numbers.
- Wave A correctly scoped (4 tasks, ~100 LOC + golden fixture) and introduces stubs that unblock Waves B-D without intermediate build break.
- Dependencies well-ordered; build green at every Wave boundary.
- Top-5 risks have clear mitigations.
- TDD discipline preserved: T-A4 RED → stub-driven GREEN → MX REFACTOR.
- User checkpoint gates expansion to Waves B-E; user retains veto.

Decision: launch **Wave A immediately**; pause after Wave A for user checkpoint; on approval, run B → C → D → E continuously.

---

## 8. Deferred Items

| Item | Destination |
|------|-------------|
| Reflexion self-critique | SPEC-V3R4-HARNESS-004 |
| Principle-based scoring | SPEC-V3R4-HARNESS-005 |
| Multi-objective + auto-rollback | SPEC-V3R4-HARNESS-006 |
| Voyager skill library | SPEC-V3R4-HARNESS-007 |
| Cross-project federation | SPEC-V3R4-HARNESS-008 |
| TF-IDF / n-gram alternative | Follow-up if SimHash FP unacceptable |
| Embedding model inference | Not planned (research §2.2 ruled out) |
| LSH / bit-bucket index | Follow-up perf SPEC if 25ms miss in prod |
| CLI verb to inspect clusters | Follow-up tooling SPEC |
| Historical tier-promotions.jsonl re-aggregate | Follow-up migration SPEC |
| Online/incremental clustering | Not planned (REQ-HRN-CLS-011: strictly batch) |
| findEventByKey optimization (key→Event map) | T-C4 implementation discretion |
| Concurrent AggregatePatterns safety beyond POSIX | Accepted via existing O_APPEND atomicity |

---

End of strategy-notes.md. Owned by manager-strategy. Downstream consumers: manager-tdd (Wave A→E execution), evaluator-active (Wave C/D quality checkpoints), manager-docs (Wave E sync).
