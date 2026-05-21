# Implementation Plan — SPEC-V3R5-STATUSLINE-STDINFIELDS-001

> Tier S LEAN minimal form (per WORKFLOW-LEAN-001). Single milestone with self-contained steps. Section A-E delegation template OPTIONAL — minimal prompt acceptable.

## Scope Summary

- 11 REQs (REQ-SSE-001..011), 11 inline binary ACs in spec.md §3
- 4 Out-of-Scope sub-sections (h3 per spec-lint B6 rule)
- 1 module touched primarily: `internal/statusline/` (4 Go files)
- 3 HARD rule files updated: `context-window-management.md`, `session-handoff.md`, `zone-registry.md`
- Estimated LOC: ~150 new + ~10 modified = ~160 LOC (well within Tier S < 300 LOC budget)
- Estimated files affected: 4 source + 3 rules + 1 test = 8 files (above Tier S < 5 guidance but justified by HARD rule mirroring — 3 of 8 are HARD rule sync, which is mechanical text replacement). Tier S judgment retained.

## Brownfield Strategy

| Strategy | Applied | Rationale |
|----------|---------|-----------|
| PRESERVE | `renderPRSegment`, `isPREnabled`, `prReviewStateColor`, `PRInfo`, `RepoInfo`, `EffortInfo`, `ThinkingInfo`, `WorkspaceInfo` | All struct fields and PR rendering logic land unchanged. AC-SSE-007 enforces no-regression. |
| EXTEND | `StdinData` (+1 field), `StatusData` (+2 fields if pass-through needed), renderer.go (+3 functions) | New code parallels existing `renderPRSegment` pattern. Minimal coupling. |
| REPLACE | None | Pure additive change. |

## Milestones

### M1 — Type mapping (REQ-SSE-003, REQ-SSE-011)

**Files modified**: `internal/statusline/types.go`

**Steps**:
1. Add `ExceedsLongTokens bool \`json:"exceeds_200k_tokens"\`` to `StdinData` struct (line ~70, between `Version` and `PR`).
2. Add `SegmentRepo = "repo"`, `SegmentLongContext = "long_context"`, `SegmentHandoffGuide = "handoff_guide"` to segment key constants (line ~270).
3. Bulk-replace `v2.1.122` → `v2.1.139` in lines 68, 69, 100, 103, 109, 111, 230, 231 (Effort/Thinking comments + StatusData docstrings).
4. Add `ExceedsLongTokens bool` pass-through to `StatusData` if renderer needs it (decide during impl; may instead read from collected stdin directly — implementer's choice).

**Acceptance**: AC-SSE-003, AC-SSE-011 PASS via grep verification.

### M2 — Renderer functions (REQ-SSE-001, REQ-SSE-002, REQ-SSE-004, REQ-SSE-005, REQ-SSE-006)

**Files modified**: `internal/statusline/renderer.go`, `internal/statusline/builder.go` (segment dispatch)

**Steps**:
1. Add `renderRepoSegment(data *StatusData) string` — returns `fmt.Sprintf("%s/%s", data.Workspace.Repo.Owner, data.Workspace.Repo.Name)` when conditions met (nil checks per REQ-SSE-001), else "".
2. Add `isRepoEnabled() bool` predicate following `isPREnabled` default-on pattern.
3. Add `renderLongContextMarker(data *StatusData) string` — returns "⚠️ long" (theme-styled via `r.theme.Warning()` when `!r.noColor`) when `data.ExceedsLongTokens == true`, else "".
4. Add `isLongContextEnabled() bool` predicate (default-on).
5. Add `shouldShowHandoffGuide(data *StatusData) bool` — implements REQ-SSE-005 threshold logic:
   - Nil-guard `data.ContextWindow` and `data.ContextWindow.UsedPercentage`
   - If `ContextWindowSize == 1000000`: return `*UsedPercentage >= 50.0`
   - Else (200K or 0): return `*UsedPercentage >= 90.0`
6. Add `renderHandoffGuideSegment(data *StatusData) string` — returns `"📋 /clear"` (theme-styled when color enabled) when `shouldShowHandoffGuide` true AND `isHandoffGuideEnabled` true, else "".
7. Add `isHandoffGuideEnabled() bool` predicate (default-on).
8. Wire all three new segment functions into the segment dispatch order in `Build` (or wherever existing `renderPRSegment` is dispatched). Recommended order (between PR and effort_thinking, near end of right-hand cluster):
   - ... → repo → pr → long_context → handoff_guide → effort_thinking → ...

**Acceptance**: AC-SSE-001, AC-SSE-002, AC-SSE-004, AC-SSE-005, AC-SSE-006 PASS via fixture tests + grep.

### M3 — HARD rule threshold sync (REQ-SSE-008, REQ-SSE-009, REQ-SSE-010)

**Files modified**: 3 rule files. NO source code changes in this milestone.

**Steps**:
1. `.claude/rules/moai/workflow/context-window-management.md`:
   - Line 17 table row: `| Opus 4.7 (1M) | 1,000,000 tokens | **75%** | ~750,000 tokens |` → `| Opus 4.7 (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |`
   - Line 25 prose: `(75% on 1M models, 90% on 200K models)` → `(50% on 1M models, 90% on 200K models)`
   - Line 41 prose: `(75% on 1M, 90% on 200K)` → `(50% on 1M, 90% on 200K)`
   - Line 78 cross-ref: `(1M = 75%, 200K = 90%)` → `(1M = 50%, 200K = 90%)`
2. `.claude/rules/moai/workflow/session-handoff.md`:
   - Line 22 Trigger #1 table cell: `**1M context model (Opus 4.7): 75%** (~750,000 tokens)` → `**1M context model (Opus 4.7): 50%** (~500,000 tokens)`
   - Line 232 cross-ref: `(1M = 75%, 200K = 90%)` → `(1M = 50%, 200K = 90%)`
3. `.claude/rules/moai/core/zone-registry.md`:
   - CONST-V3R5-022 clause: replace verbatim `1M context (Opus 4.7) = 75%` substring → `1M context (Opus 4.7) = 50%`
   - CONST-V3R5-025 clause: same replacement on its quoted threshold reference
   - Leave 200K (Sonnet) = 90% unchanged in both entries

**Acceptance**: AC-SSE-008, AC-SSE-009, AC-SSE-010 PASS via grep verification.

### M4 — Test coverage (≥ 85% target per TRUST 5)

**Files added**: `internal/statusline/renderer_repo_test.go`, `internal/statusline/renderer_long_context_test.go`, `internal/statusline/renderer_handoff_guide_test.go` (one test file per new segment for clarity), OR fold into existing `renderer_test.go` if conventions prefer.

**Steps**:
1. Table-driven test for `renderRepoSegment` (5 cases): valid repo, nil Workspace, nil Repo, empty Owner, empty Name.
2. Table-driven test for `renderLongContextMarker` (2 cases): true / false.
3. Table-driven test for `shouldShowHandoffGuide` (6 cases): 1M@50%, 1M@49%, 1M@51%, 200K@90%, 200K@89%, nil ContextWindow.
4. Table-driven test for `renderHandoffGuideSegment` (2 cases: should-show + opt-out via disabled config).
5. Integration test parsing full stdin JSON fixture containing all three new fields and asserting all three segments render.

**Acceptance**: `go test -cover ./internal/statusline/...` reports ≥ 85% package coverage. `go test -race ./internal/statusline/...` PASS.

### M5 — Self-verification batch (read-only parallel — per agent-common-protocol.md §Parallel Execution)

The manager-develop sub-agent on completion executes the 7-item read-only verification batch in a single turn (multi-Bash parallel):

```bash
go test ./internal/statusline/...                                    # M2/M4 PASS
go test -coverprofile=cover.out ./internal/statusline/...            # ≥ 85% coverage
grep -rn 'AskUserQuestion' internal/statusline/ | grep -v "_test.go" # C-HRA-008: 0 matches
grep -E "1M.*75|750,?000" .claude/rules/moai/workflow/context-window-management.md  # AC-SSE-008: 0
grep -E "1M.*50|500,?000" .claude/rules/moai/workflow/context-window-management.md  # AC-SSE-008: ≥2
go vet ./internal/statusline/...                                     # clean
golangci-lint run --timeout=2m ./internal/statusline/...             # NEW = 0
```

Additionally for cross-platform sanity (`statusline` package is pure Go, no syscall — no Windows build tag work required, but verify):

```bash
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
```

## Technical Approach

### Default-on convention rationale

The PR segment in `commit e71f1aa54` established the default-on convention (unset config key → enabled). This SPEC follows the same pattern for `repo`, `long_context`, `handoff_guide`. Rationale: v2.1.145+/v2.1.146+ stdin payload carries these fields whenever available, so the user benefits from immediate visibility without YAML config edits. Opt-out remains possible via explicit `segments.repo: false` in `.moai/config/sections/statusline.yaml`.

### Color theme integration

Both `renderLongContextMarker` and `renderHandoffGuideSegment` use `r.theme.Warning()` (yellow) by default. The `⚠️` glyph is already used by other Warning-class segments (no new icon vocabulary). When `r.noColor == true`, the segments render plain text without ANSI escapes.

### Threshold tightening — why 50%

| Tokens used | 75% (old) | 50% (new) | Headroom |
|-------------|-----------|-----------|----------|
| 750K | At threshold | Far above | Old: 0 / New: -250K (handoff already due) |
| 500K | Below | At threshold | New: 500K headroom (matches 200K model behavior at 90% / 180K used = 20K headroom but proportionally similar) |
| 250K | Below | Below | Both below — normal operation |

The 50% choice gives Opus 4.7 1M sessions an absolute headroom of ~500K tokens before SSE stall risk, which empirically (per W3 + ATOMIC-WRITE-001 incident memory) is sufficient. The 75% choice gave only 250K headroom, observed insufficient.

### Tier S vs Tier M judgment justification

- LOC: ~160 (Tier S threshold < 300) ✅
- Files: 8 (Tier S threshold < 5 ❌) — exceeded by 3, but exceeded files are all HARD rule mirror sync (mechanical), not implementation files. The implementation files (4) fit Tier S.
- Domain: Statusline rendering + HARD rule mirroring (no constitutional change)
- Risk: Low (additive change, regression guard on PR segment)

Net: Tier S with 1-line judgment note in spec.md frontmatter (`tier: S`). plan-auditor first-pass score regression on Tier-up suggestion would tier-up to M; otherwise S stands.

## Commit Strategy (Late-Branch, per LATE-BRANCH-001)

This plan-phase produces ONE commit on `main` directly. NO feature branch creation in plan-phase. Commit message (Conventional Commits + `🗿 MoAI <email@mo.ai.kr>` trailer):

```
plan(SPEC-V3R5-STATUSLINE-STDINFIELDS-001): statusline stdin enrichment + 1M handoff threshold tightening (Tier S)

- Add workspace.repo segment renderer (gap in v2.1.146 stdin enrichment)
- Add exceeds_200k_tokens mapping + Layer 1 visual marker
- Add Layer 2 handoff_guide segment (1M ≥50%, 200K ≥90%)
- Tighten 1M context handoff threshold: 75% → 50% (SSE stall margin)
- Mirror policy change to session-handoff.md + zone-registry.md
- Side-fix: correct types.go v2.1.122 → v2.1.139 comments

🗿 MoAI <email@mo.ai.kr>
```

PR cherry-pick happens at sync-phase, not here.

## Risk Register (inherited from spec.md §4)

See spec.md §4 (Risks R-SSE-001..003). No new risks identified during plan-phase.

## Self-Verification Checklist (plan-phase exit criteria)

- [x] spec.md present (11 REQs + 11 inline ACs + 4 Out of Scope h3 sub-sections + 3 Risks + 5 Constraints + 3 Dependencies)
- [x] plan.md present (5 milestones + brownfield strategy + technical approach + commit strategy)
- [x] Tier S 2-artifact form (no acceptance.md per WORKFLOW-LEAN-001)
- [x] All 11 REQs have binary AC inline
- [x] All AC commands are single-shell-line verifiable
- [x] PRESERVE list explicit (PR segment regression guard)
- [x] Frontmatter canonical schema: 12 required fields + `tier: S` optional field present
- [x] No `created_at`/`updated_at` snake_case (uses `created`/`updated` per schema)
- [x] `tags` as comma-separated string (not array)
- [x] HARD rule mirror plan complete (3 files: context-window-management, session-handoff, zone-registry)
- [x] 50% threshold rationale documented (technical approach section)

Ready for plan-auditor invocation at Tier S threshold = 0.75.
