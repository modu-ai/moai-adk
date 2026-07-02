# Progress — SPEC-TOKEN-EFFICIENCY-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (2 items / 2 Go domains / ~6–8 files / ~250–500 LOC)
- Artifacts: spec.md + plan.md + acceptance.md + progress.md (created 2026-07-02;
  revised to v0.2.0 2026-07-02 — "~7×" reword carved out)
- SPEC ID pre-write self-check:
  `decomposition: SPEC ✓ | TOKEN ✓ | EFFICIENCY ✓ | 001 ✓ → PASS`
- Frontmatter: 12 canonical fields present (`created`/`updated`/`tags`, no
  snake_case aliases). version bumped 0.1.0 → 0.2.0 (scope reduction).
- Out of Scope: `### Out of Scope — <topic>` H3 sub-headings with `-` bullets
  present in spec.md §A.5 (incl. the carve-out subsection).
- GEARS REQs: REQ-TEF-001…007 (Ubiquitous / Event-driven / Unwanted / Where).
- REQ↔AC bijection: 7 REQ ↔ 7 AC (AC-TEF-001…007), contiguous.
- Scope carve-out (v0.2.0): the "~7×" attribution reword (formerly P0-2 =
  REQ/AC-TEF-005/006/007) is carved out to SPEC-DIVECC-ATTRIBUTION-FIX-001 after
  a plan-audit FAIL (2 BLOCKING defects confined to that item). Surviving items
  renumbered: P0-1 (budget guard, REQ/AC 001-004) + P0-2 (cache-hit-ratio
  statusline, formerly P0-3, REQ/AC 005-007). D4 (MINOR) fixed: AC-TEF-002 no
  longer hardcodes the "13 files ~258KB" snapshot — the Given/assertion is now
  relative (no-`paths:` rule count + 3 fixed surfaces, token-sum ≤ budget).
- Plan-phase verification resolved:
  - P0-2 statusline cache-field availability **VERIFIED** — official schema
    (`context_window.current_usage.cache_read_input_tokens` /
    `cache_creation_input_tokens`) + existing Go struct
    `internal/statusline/types.go` `CurrentUsageInfo` + consumption in
    `memory.go`. Null/`/compact` degradation trigger confirmed in official docs.
- Open plan decisions (deferred to run-phase, recorded in plan.md §D):
  D-1 token-estimation method (recommend char/4), D-2 budget value (recommend
  baseline+15%), D-3 P0-2 display form (recommend cache-hit %).
- plan_status: audit-ready
- _plan-auditor verdict: pending (re-audit after v0.2.0 scope reduction)_

## §E.2 Run-phase Evidence

Run-phase decisions applied (from plan.md §D): D-1 = **char/4** (`estimateTokens(b) = len(b)/4`,
zero-dependency tripwire); D-2 = budget **75000** (measured baseline 64,624 tokens +
~15% = 74,317, rounded to clean constant); D-3 = **cache-hit %** (`cache_read /
(cache_read + cache_creation) * 100`, rendered `💾 N%`).

Files (new/edited within scope):
- NEW `internal/config/token_budget_guard.go` — `estimateTokens`, `AlwaysLoadedTokenBudget=75000`,
  dynamic no-`paths:` enumeration (`alwaysLoadedSurface`/`hasPathsRestriction`/`frontmatterHasPaths`),
  `memoryHead` (200-line/25KB cap), `findRepoRoot`, `measureAlwaysLoaded`.
- NEW `internal/config/token_budget_guard_test.go` — AC-TEF-001..004 + edges.
- EDIT `internal/statusline/types.go` — `SegmentCacheHit` const + `StatusData.CacheUsage` field.
- EDIT `internal/statusline/renderer.go` — `cacheHitPercent` + `renderCacheHit` + `renderInfoLine` hook.
- EDIT `internal/statusline/builder.go` — `collectAll` populates `CacheUsage` from `CurrentUsage`.
- NEW `internal/statusline/cache_hit_test.go` — AC-TEF-005..007 + edges.

### AC PASS/FAIL matrix (§E1)

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-TEF-001 (over-budget fails) | PASS | `go test -run TestAlwaysLoadedTokenBudget_OverBudgetFails ./internal/config/` | `PASS` (over-budget/under-budget sub-tests both assert fire/no-fire) |
| AC-TEF-002 (enumerate + pass-on-tree) | PASS | `go test -run 'TestAlwaysLoadedTokenBudget$|TestAlwaysLoadedSurfaceEnumeration' ./internal/config/` | `PASS` — surface = (no-`paths:` count) + 3 fixed; total 64,624 ≤ 75,000; relative assertion, no literal count |
| AC-TEF-003 (named/documented) | PASS | `grep -n 'estimateTokens\|AlwaysLoadedTokenBudget' internal/config/token_budget_guard.go` | both are named symbols with doc comments (char/4 method + baseline+headroom derivation) |
| AC-TEF-004 (scoped excluded, silent PASS) | PASS | `go test -run 'TestAlwaysLoadedSurfaceEnumeration|TestHasPathsRestriction' ./internal/config/` | `PASS` — `languages/go.md` (paths:-scoped) absent from surface; malformed frontmatter counts as always-loaded |
| AC-TEF-005 (ratio when data present) | PASS | `go test -run 'TestCacheHitPercent|TestRenderCacheHit' ./internal/statusline/` | `PASS` — `(2000,5000)→28%`, rendered `💾 28%` |
| AC-TEF-006 (graceful degrade null+zero) | PASS | `go test -run 'TestCacheHitPercent|TestRenderCacheHit' ./internal/statusline/` | `PASS` — nil usage → ""; cache_creation==0 → ""; both-zero → "" (no 0/0); negative-read denom==0 defense |
| AC-TEF-007 (config-toggleable) | PASS | `go test -run TestCacheHitSegmentToggle ./internal/statusline/` | `PASS` — `{cache_hit:false}` absent; default/`{cache_hit:true}` present (parallel to SegmentEffortThinking) |

Invariants:
- D-2 independent-shippability: P0-1 (`internal/config`) and P0-2 (`internal/statusline`) are separate packages — PASS.
- D-3 P0-1 hermetic: reads repo-relative files only; missing MEMORY.md → 0 tokens (repo-root MEMORY.md absent) — PASS.
- D-4 P0-2 no stdin schema change: reads already-parsed `CurrentUsage`; `.moai/status_line.sh.tmpl` untouched — PASS.
- Scope discipline: `SegmentCacheHit` is a render-time toggle constant (like `SegmentRepo`), NOT added to the 15-key `CanonicalSegments` schema — avoids the settings/profile/web/i18n cascade (B10 PRESERVE).

### §E.2 Design decisions (run-phase)

- **P0-2 toggle placement**: `CanonicalSegments` (preset.go) has NO real consumer (only a comment
  reference in `cli/update.go`); `defaultStatuslineSegments`/settings schema/profile setup each carry
  independent hardcoded 15-key lists. Adding cache_hit to `CanonicalSegments` would NOT surface the
  toggle without cascading edits to all three + web + i18n×4. Minimal choice: render-time toggle via
  `isSegmentEnabled(SegmentCacheHit)` (default-on for unknown key), disable with `cache_hit: false` in
  statusline.yaml. Satisfies AC-TEF-007 fully. No existing render test supplies `cache_creation>0`, so
  default-on introduces no regression.
- **MEMORY.md hermeticity (D-3)**: the always-loaded MEMORY.md is the machine-specific auto-memory copy
  (`~/.claude/projects/{hash}/memory/`), outside the repo. Measuring it would break hermeticity, so the
  guard measures the repo-relative `MEMORY.md` (absent → 0 tokens; head-capped when present). The 3 fixed
  surface slots are always enumerated (count = no-`paths:` rules + 3) even when a file is absent, matching
  AC-TEF-002's "+3 fixed surfaces".

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-02
run_commit_sha: 3e30fef48eb20041c71f14783435a6847f675509
run_status: implemented
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: performed (see §race note)
l44_post_push_fetch: performed
new_warnings_or_lints_introduced: 0  # golangci-lint ./internal/config/... ./internal/statusline/... → 0 issues
cross_platform_build:
  darwin_amd64: exit 0
  windows_amd64: exit 0  # GOOS=windows GOARCH=amd64 go build ./...
coverage_new_files:
  token_budget_guard.go: ">=86.7% (all funcs 86.7-100%)"
  renderer.go_new_funcs: "cacheHitPercent 100%, renderCacheHit 100%"
total_run_phase_files: 6  # 3 new + 3 edited (all within internal/config + internal/statusline)
m1_to_mN_commit_strategy: single consolidated commit (M1+M2+M3, Tier M small scope, main-direct)
full_suite_note: >
  go test ./... shows 1 failing package (internal/template TestSanitizedPairParity/agent-common-protocol.md).
  Attributed to a PARALLEL SESSION's uncommitted work (untracked internal/template/sanitized_pair_parity_test.go
  + committed rule-file mirror divergence) — NOT this SPEC's scope. This SPEC's two packages
  (internal/config, internal/statusline) pass cleanly. Not swept into this commit (specific-path staging).
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-02
sync_commit_sha: 8f626b67f
sync_status: completed
changelog_entry_position: "[Unreleased] ### Added section, SPEC-TOKEN-EFFICIENCY-001"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented"
  updated_field_refresh: "2026-07-02"
b12_self_test_a: "grep -c 'SPEC-TOKEN-EFFICIENCY-001' CHANGELOG.md → 1 (no duplicates from parallel BATCH-SYNC)"
b12_self_test_b: "AC count match: acceptance.md has 7 ACs (AC-TEF-001..007); CHANGELOG cites 7/7 AC PASS"
b12_self_test_c: "ls -1 internal/config/token_budget_guard{,_test}.go internal/statusline/{cache_hit_test.go,renderer.go,builder.go,types.go} → 6 files (all cited in CHANGELOG ✓)"
canary_compliance_check: "PASS (no Mx-deferred items, all ACs closed, CHANGELOG prose conformant)"
```
