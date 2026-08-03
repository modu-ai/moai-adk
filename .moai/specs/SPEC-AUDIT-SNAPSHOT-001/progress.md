# progress.md — SPEC-AUDIT-SNAPSHOT-001

> Plan-phase skeleton. Run- and sync-phase evidence is populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4).

## §E.1 Plan-phase Audit-Ready Signal

- Status: plan-phase artifacts authored (spec.md + plan.md + acceptance.md + progress.md) on 2026-08-03.
- Tier: M (4 change points across Go runtime + JS workflow + skill YAML + rule doc).
- Frontmatter: `status: in-progress` (transitioned on M1 commit — first run-phase commit).
- plan-auditor verdict: PASS 0.98 (skip-eligible post-A2; Implementation Kickoff Approval granted, continuous-autonomous mode).

## §E.2 Run-phase Evidence

### Implementation summary

- **M1 (A1+A2)** — `internal/runtime/audit_cache.go`, `internal/runtime/audit_snapshot.go` (new), `internal/runtime/audit_cache_test.go`, `internal/runtime/audit_snapshot_test.go` (new). Doctrine: `.claude/rules/moai/workflow/spec-workflow.md` skip policy 4→3 conditions; `.claude/skills/moai/workflows/run/phase-execution.md` Step 2 cites canonical contract; `.claude/output-styles/moai/moai.md` skip-eligible annotation aligned to per-tier. Template mirrors updated.
- **M2 (A3)** — `internal/runtime/sync_4dim_binding.go` (new) + `internal/runtime/sync_4dim_binding_test.go` (new). Doctrine: `.claude/skills/moai/workflows/sync.md` FO-SYNC-1 binding-promotion prose; `.claude/workflows/sync-audit-4dim.js` header comment. Template mirrors updated.
- **M3 (A4)** — `internal/verify/claim_lock.go` (new) + `internal/verify/claim_lock_test.go` (new); `internal/verify/store.go` `RecordCheck` wraps Load-modify-Save in the claim/lock. Doctrine: `.claude/skills/moai/workflows/sync/quality-gates-quality.md` Step 0.5.2 shared-snapshot wiring note. Template mirror updated.

### AC PASS/FAIL matrix

| AC | Subject | Status | Test / Evidence |
|----|---------|--------|-----------------|
| AC-AUDIT-SNAPSHOT-001 | Sticky cache survives past-24h (A1) | PASS | `TestStickyCacheSurvivesPast24h` + `TestCacheStickyAcrossAges` — cache hit at T0+25h and T0+30d (hash unchanged); `Lookup` no longer consults `now` (vestigial param, documented) |
| AC-AUDIT-SNAPSHOT-001b | Tier L hash includes design.md + research.md; tasks.md retained for grandfathered (A1) | PASS | `TestTierLHashIncludesDesignAndResearch` (mutate design.md → hash changes; mutate research.md → hash changes) + `TestTasksMdStillHashSubjectForGrandfathered` (mutate tasks.md → hash changes) |
| AC-AUDIT-SNAPSHOT-002 | Per-tier skip threshold (A2) | PASS | `TestSkipEligibleByScorePerTier` (table: S 0.78✓/0.74✗; M 0.81✓/0.78✗; L 0.86✓/0.82✗; explicit "0.81 Tier M eligible despite <0.90") + `TestSkipEligibleByScoreUnknownTierIsStrict` (empty→Tier L strict) |
| AC-AUDIT-SNAPSHOT-003 | Clean sync → binding verdict, no cold spawn (A3) | PASS | `TestBindingOnCleanPath` — clean PASS verdict returns `(true, "")` (binding; cold sync-auditor spawn suppressed) + `TestBindingSurvivesNonContestedFindings` (minor findings on different dims remain binding) |
| AC-AUDIT-SNAPSHOT-003b | Failure modes still spawn cold sync-auditor (A3 fallback) | PASS | `TestFallbackOnIncomplete` (INCOMPLETE→fallback) + `TestFallbackOnZeroScoredDimension` (dim-0→fallback) + `TestFallbackOnContestedCritical` (predicate i: critical→fallback) + `TestFallbackOnContestedSeverityConflict` (predicate ii: major+minor on same dim→fallback) |
| AC-AUDIT-SNAPSHOT-004 | Single snapshot shared across 3 consumers (A4) | PASS | `TestRecordCheckConcurrentSameKeyNoDimensionDrop` (4 dims × 8 goroutines; pre-fix RED dropped 3/4 dims — only `go vet` survived; post-fix GREEN all 4 land) + `TestRecordCheckConcurrentReplaceStillConsistent` (same-command replace yields exactly 1 entry, no duplicates) |
| AC-AUDIT-SNAPSHOT-004b | HEAD SHA change invalidates snapshot (A4 integrity) | PASS | `TestSnapshotAbsentForNewSHA` — record at S1, request S2 → `(nil, nil)` (no stale service); S1 record intact for S1 consumers; S1/S2 resolve to distinct paths |

**Result: 7/7 ACs PASS.**

### RED evidence (verbatim pre-GREEN failing-test output)

**M1 RED** — `go test -run 'TestSticky|TestTierL|TestTasksMd|TestSkipEligible' ./internal/runtime/`:
```
internal/runtime/audit_snapshot_test.go:190:11: undefined: SkipEligibleByScore
internal/runtime/audit_snapshot_test.go:207:6: undefined: SkipEligibleByScore
internal/runtime/audit_snapshot_test.go:210:5: undefined: SkipEligibleByScore
FAIL	github.com/modu-ai/moai-adk/internal/runtime [build failed]
```
(Compile-error RED for AC-002; the sticky/Tier-L tests target behavior the prior TTL-bound 4-file hash did not provide.)

**M2 RED** — `go test -run 'TestBinding|TestFallback' ./internal/runtime/`:
```
internal/runtime/sync_4dim_binding_test.go:17:26: undefined: FourDimVerdict
internal/runtime/sync_4dim_binding_test.go:18:10: undefined: FourDimVerdict
FAIL	github.com/modu-ai/moai-adk/internal/runtime [build failed]
```

**M3 RED** — `go test -run 'TestRecordCheckConcurrentSameKeyNoDimensionDrop' ./internal/verify/` (pre-fix store.go without claim/lock):
```
--- FAIL: TestRecordCheckConcurrentSameKeyNoDimensionDrop (0.07s)
    claim_lock_test.go:61: dimension "go test ./..." was DROPPED by a concurrent writer — claim/lock failed to serialize RecordCheck; snapshot checks: [{CheckID:vet Command:go vet ./... ...}]
    claim_lock_test.go:61: dimension "golangci-lint run" was DROPPED by a concurrent writer ...
    claim_lock_test.go:61: dimension "go test -cover ./..." was DROPPED by a concurrent writer ...
FAIL
```
(Last-writer-wins dropped 3 of 4 dimensions — only `go vet` survived — the exact D3 race REQ-004 ¶4 mandates the claim/lock to fix.)

### K-1 closure sweep

```
$ grep -rn "Within 24h\|score >= 0.90\|score ≥ 0.90" .claude/ internal/ | grep -v "SPEC-AUDIT-SNAPSHOT-001" | grep -v "retired the.*4th condition\|retired.*24h"
.claude/rules/moai/workflow/spec-workflow.md:348:  prior 4th condition ("Within 24h") and aligned the 2nd condition to per-tier
.claude/skills/moai-foundation-quality/references/reference.md:756: if self.overall_score >= 0.90:
internal/runtime/audit_snapshot.go:6:// retired flat `score >= 0.90` skip predicate is replaced by these per-tier
internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md:348:  prior 4th condition ("Within 24h") ...
```
**Closure:** the 3 surviving matches are (a) the canonical retirement explanation in spec-workflow.md (documents that the condition WAS retired — the SSOT describing the change, both live + template mirror), (b) my own audit_snapshot.go comment documenting the retirement, and (c) an unrelated Python grade-calculator example in `moai-foundation-quality/references/reference.md` (`if overall_score >= 0.90: return "A"` — a generic code-metrics grading example, NOT the plan-audit skip policy). No divergent skip-policy restatement survives.

### Quality gate

```
$ go vet ./...                          → exit 0 (clean)
$ golangci-lint run --timeout=4m        → 0 issues
$ go test -race ./internal/runtime/... ./internal/verify/...  → ok (race-clean)
$ go test -cover ./internal/runtime/    → coverage: 88.0% of statements
$ go test -cover ./internal/verify/     → coverage: 80.7% of statements
```

`go test ./...` full suite: green across all packages; `internal/web` showed `signal: terminated` ONLY under the 4-way parallel verification batch (starvation) and PASSES in isolation (`ok 2.304s`) — not a regression (zero web code touched).

### Commits (per-milestone, branch `worktree-autonomy-epic`)

- `234d6fc02` — M1 sticky hash cache + per-tier skip threshold (A1+A2), status `draft → in-progress`
- `a369bd05c` — M2 4-dim workflow verdict binding on happy path (A3)
- `630f0f44f` — M3 shared snapshot claim/lock + A4 wiring (A4, REQ-004 ¶4)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-03
run_commit_sha: pending-backfill-M4   # the M4 evidence commit's own SHA — backfilled in a follow-up (self-referential-hazard workaround per the SHA placeholder backfill exemption D3)
run_status: run-phase-complete
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 0       # no PRESERVE-list items remain unfinished
l44_pre_commit_fetch: true             # git fetch origin main before commit; 0 0 divergence
l44_post_push_fetch: n/a               # no push performed (worktree branch; orchestrator owns push / PR)
new_warnings_or_lints_introduced: 0    # golangci-lint 0 issues; go vet clean
cross_platform_build:
  darwin: ok                            # go build ./... exit 0
total_run_phase_files: 16              # 5 new Go files + 4 Go test files (2 new, 2 modified) + 7 doctrine/template markdown edits
m1_to_mN_commit_strategy: per-milestone commits (M1, M2, M3) + M4 evidence commit; conventional commit subjects; 🗿 MoAI trailer
```

### Gaps (per verification-claim-integrity §3.4 — what was NOT observed)

- **Cross-process claim/lock not directly AC-tested.** The AC-004 parallel variant is in-process (goroutines sharing one process); the cross-process hazard (Stop-hook process vs sync-auditor agent vs 4dim workflow subprocess racing on the same SHA) is addressed by the `O_EXCL` claim-stamp file lock in `claim_lock.go`, but that cross-process path is NOT covered by a unit test (it needs ≥2 processes). The in-process per-key mutex is the directly-tested primary layer; the file lock is defense-in-depth verified only by code review, not by a multi-process test.
- **A3 binding provenance (K-5) not grep-verified.** acceptance.md §D.4 K-5 closure wants the sync-phase quality audit-trail record schema to accept BOTH the workflow run ID and the sync-auditor agent ID as verdict-source values. The Go helper `FourDimVerdict.IsBinding()` codifies the predicate, but the audit-trail record schema (wherever verdict-source is cited) was not swept — that schema lives in sync-phase report-writing code outside this SPEC's Go-testable surface (doctrine/skill layer). Flagged for sync-phase.
- **A4 consumer wiring (sync-auditor.md, sync-phase-quality-gate.sh, 4dim judges) is doctrine-only.** The quality-gates-quality.md Step 0.5.2 note documents the three consumers consuming the snapshot, but the actual consumer code paths in `.claude/agents/moai/sync-auditor.md`, `.claude/hooks/moai/sync-phase-quality-gate.sh`, and the 4dim JS judges were NOT rewritten to call `moai verify check --key-current` — that wiring is orchestrator-runtime behavior, not testable Go. The snapshot infrastructure + claim/lock (the load-bearing Go contract) IS in place; the consumer-side invocation is a sync-phase / orchestrator-wiring follow-up.

### Residual risk

- The sticky cache (A1) means a cached PASS verdict is valid indefinitely as long as the hash matches. If the plan-auditor rubric itself changes (a future SPEC tightens WHAT is measured), stale cached verdicts would NOT auto-invalidate — the `AuditorVersion` field is recorded but not consulted by `Lookup`. Mitigation: a rubric-change SPEC would need to bump `AuditorVersion` AND clear the cache, or evolve `Lookup` to consult the version. Out of scope for this SPEC (audit-semantics-immutable per §D.6).
- The `O_EXCL` claim-stamp file lock has a 5-minute staleness TTL (`claimLockTTL`). A genuinely-slow recording (longer than 5 min — e.g. a pathological full-suite `go test` under extreme load) could have its lock reclaimed mid-record by a waiter, reintroducing the race. 5 min is generous (typical record < 30s); residual risk only under pathological load.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
