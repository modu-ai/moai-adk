---
id: SPEC-V3R5-STATUSLINE-STDINFIELDS-001
title: "Statusline stdin schema enrichment (workspace.repo render + exceeds_200k_tokens marker + handoff_guide) + 1M handoff threshold tightening"
version: "0.1.0"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: P1
phase: "v2.20.0-rc1"
module: "internal/statusline"
lifecycle: spec-anchored
tags: "statusline, stdin-schema, context-window, handoff, v2.1.146, threshold-tightening"
tier: S
---

# SPEC-V3R5-STATUSLINE-STDINFIELDS-001 — Statusline stdin schema enrichment + 1M handoff threshold tightening

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial draft — plan-phase. Tier S, LEAN minimal form (spec.md + plan.md). |

## 1. Background

### 1.1 Sources of truth

- **Official Claude Code statusline docs** (verbatim Available data table): `https://code.claude.com/docs/en/statusline`.
  - `workspace.repo.{host,owner,name}` listed as stdin field since v2.1.145.
  - `exceeds_200k_tokens` listed as boolean stdin field since v2.1.139.
- **Anthropic CHANGELOG.md** (raw fetched 2026-05-22):
  - v2.1.146 entry: "Status line JSON input now includes GitHub repo and PR information when detected."
  - v2.1.139 entry: "Status line: stdin JSON now includes `effort.level` and `thinking.enabled`."
- **User's actual stdin payload** (`~/.moai/cache/statusline_debug.log` 2026-05-20 18:27 entry, Claude Code v2.1.145 — confirmed by orchestrator):
  - `workspace.repo.{host,owner,name}` present (arriving but not rendered)
  - `exceeds_200k_tokens: true` present (not mapped in `StdinData`)
  - `pr` absent (user on v2.1.145, not v2.1.146 yet)

### 1.2 Defects covered

| Defect | Evidence | Impact |
|--------|----------|--------|
| D1: `workspace.repo` mapped (`internal/statusline/types.go:160`) but no segment renderer | `grep -n "renderRepo" internal/statusline/renderer.go` → 0 matches | User cannot see "modu-ai/moai-adk" in statusline despite stdin providing it |
| D2: `exceeds_200k_tokens` NOT mapped in `StdinData` | `grep -n "exceeds_200k\|ExceedsLongTokens" internal/statusline/types.go` → 0 matches | Long-context warning marker cannot fire |
| D3: 1M context handoff threshold 75% (`context-window-management.md:17`) too lax | SSE stall incidents at high absolute load (e.g., 750K+ tokens) | Lost work when stream stalls; recovery requires `/clear` mid-task |
| D4: Code comment `internal/statusline/types.go:68-69` cites "v2.1.122+" for `Effort`/`Thinking` — actual CHANGELOG entry is v2.1.139 | CHANGELOG verification 2026-05-22 | Documentation drift; new contributors trust wrong version |

### 1.3 Root cause

- **D1**: SPEC-V3R5-STATUSLINE-V2145-001 (merged `fb3d1e22b`) added the struct mapping but stopped at `RepoInfo` definition; no renderer was wired into the segment list. This SPEC fills the gap.
- **D2**: `exceeds_200k_tokens` was introduced in v2.1.139 alongside `effort.level` / `thinking.enabled`, but only the latter two were mapped in the v3R5 statusline catch-up SPEC. This field was overlooked.
- **D3**: Current 75% threshold was selected as a conservative margin in `SPEC-V3R5-LATE-BRANCH-001` era. Field evidence (intermittent SSE stalls in W3 HARNESS-AUTONOMY-001 + ATOMIC-WRITE-001 sessions, both > 700K tokens) shows the absolute headroom is insufficient. Tightening to 50% gives ~500K absolute ceiling, leaving 500K margin before SSE stall risk.
- **D4**: Side-fix surfaced during source verification for this SPEC. Code comment was copy-pasted at SPEC-V3R5-EFFORT-THINKING-001 time without consulting CHANGELOG.

### 1.4 Why now

- v2.1.146 PR enrichment is downstream of v2.1.145+; users upgrading to v2.1.146 expect the repo segment to be visible. Without REQ-SSE-001..002 this SPEC, the segment remains hidden.
- 1M Opus 4.7 sessions are now the default modality (per `feedback_w3_metaanalysis_lessons.md`). The 75% → 50% tightening shifts the operational point to a regime with proven stability.

## 2. Goals

| Goal | Description |
|------|-------------|
| G1 | Render `workspace.repo.owner/name` as a new statusline segment when stdin provides it. |
| G2 | Map `exceeds_200k_tokens` to `StdinData.ExceedsLongTokens` and surface a Layer 1 visual marker (⚠️). |
| G3 | Add a Layer 2 `handoff_guide` segment that activates when `context_window.used_percentage` crosses the model-class threshold (1M ≥ 50%, 200K ≥ 90%). |
| G4 | Tighten the 1M context handoff threshold from 75% to 50% across the three HARD-rule sources (`context-window-management.md`, `session-handoff.md`, `zone-registry.md`). |
| G5 | Verify PR segment activation prerequisites are unchanged (no regression on `SPEC-V3R5-STATUSLINE-V2145-001` behavior). |
| G6 | Side-fix: correct code comments in `internal/statusline/types.go` to cite v2.1.139 (CHANGELOG-accurate) for `Effort` / `Thinking`. |
| G7 | Plan-auditor PASS ≥ 0.75 (Tier S threshold) on first iteration. |

## 3. EARS Requirements + Inline Acceptance Criteria

> Tier S convention (per WORKFLOW-LEAN-001): AC inline alongside each REQ. Each AC is binary (PASS/FAIL) and independently verifiable via a single shell command.

### REQ-SSE-001 — Repo segment renderer

The statusline renderer SHALL provide a `renderRepoSegment` function that returns `owner/name` (e.g., `modu-ai/moai-adk`) when `data.Workspace.Repo` is non-nil AND both `Owner` and `Name` are non-empty. The segment SHALL return empty string when `data.Workspace` is nil, `data.Workspace.Repo` is nil, or either `Owner` or `Name` is empty.

**AC-SSE-001** (binary): `grep -n "func.*renderRepoSegment\|renderRepoSegment(data" internal/statusline/renderer.go` returns ≥ 2 matches (definition + call site). The function body contains the string literal `"%s/%s"` for `Owner`/`Name` interpolation. **Verification**: `grep -A 15 "func.*renderRepoSegment" internal/statusline/renderer.go | grep -E "Owner|Name|/%s"`.

### REQ-SSE-002 — Repo segment activation

The `repo` segment SHALL be controlled by an `isRepoEnabled()` predicate following the `isPREnabled()` default-on pattern (REQ-SLV-012 supersession established in commit `e71f1aa54`). Default-on for v2.20.0-rc1: unset `segments.repo` key in `statusline.yaml` resolves to enabled.

**AC-SSE-002** (binary): `grep -n "SegmentRepo\s*=\s*\"repo\"" internal/statusline/types.go` returns exactly 1 match. `grep -n "isRepoEnabled" internal/statusline/renderer.go` returns ≥ 1 match. Fixture test renders statusline from stdin containing `{"workspace":{"repo":{"host":"github.com","owner":"modu-ai","name":"moai-adk"}}}` and asserts output substring `"modu-ai/moai-adk"`.

### REQ-SSE-003 — ExceedsLongTokens field mapping

`StdinData` SHALL contain a field `ExceedsLongTokens bool` mapped to JSON key `exceeds_200k_tokens` (Claude Code v2.1.139+). When the field is absent in stdin, the Go default (`false`) SHALL apply.

**AC-SSE-003** (binary): `grep -n "ExceedsLongTokens\s*bool.*exceeds_200k_tokens" internal/statusline/types.go` returns exactly 1 match.

### REQ-SSE-004 — Long-context visual marker (Layer 1)

The statusline renderer SHALL provide a `renderLongContextMarker` function that returns the marker string `⚠️ long` (or `⚠️ long` with theme-aware color) when `data.ExceedsLongTokens` is `true`. The segment SHALL return empty string when `data.ExceedsLongTokens` is `false`. The marker SHALL NOT trigger any handoff semantics — pure visual signal.

**AC-SSE-004** (binary): Fixture test renders statusline from stdin `{"exceeds_200k_tokens":true}` and asserts output contains `"⚠️"` AND `"long"`. Negative fixture stdin `{"exceeds_200k_tokens":false}` asserts output does NOT contain `"⚠️ long"`.

### REQ-SSE-005 — Handoff guide segment activation logic

The renderer SHALL provide a `shouldShowHandoffGuide(data)` predicate that returns `true` when:
- `data.ContextWindow != nil`
- `data.ContextWindow.UsedPercentage != nil`
- For 1M model class (`data.ContextWindow.ContextWindowSize == 1000000`): `*UsedPercentage >= 50.0`
- For 200K model class (`data.ContextWindow.ContextWindowSize == 200000`): `*UsedPercentage >= 90.0`
- Otherwise (window size unknown or 0): default to 200K rule (>=90%) for safety.

**AC-SSE-005** (binary): Table-driven unit test exercises four cases: (1M, 50%, expect true) (1M, 49%, expect false) (200K, 90%, expect true) (200K, 89%, expect false). All four assertions PASS.

### REQ-SSE-006 — Handoff guide segment rendering (Layer 2)

When `shouldShowHandoffGuide` returns `true`, the renderer SHALL emit a paste-ready hint segment of the form `📋 /clear` (or theme-styled equivalent), instructing the user to run `/clear` and paste the resume message generated by the orchestrator per `session-handoff.md` canonical format. The segment SHALL be controlled by `isHandoffGuideEnabled()` with default-on semantics.

**AC-SSE-006** (binary): Fixture test renders statusline from stdin with `{"context_window":{"used_percentage":60,"context_window_size":1000000}}` and asserts output contains `"/clear"` OR `"📋"`.

### REQ-SSE-007 — PR segment regression guard

The existing `renderPRSegment` (introduced in SPEC-V3R5-STATUSLINE-V2145-001, commit `fb3d1e22b`) SHALL continue to render unchanged. No code path added by this SPEC SHALL modify `renderPRSegment`, `isPREnabled`, `prReviewStateColor`, or the `PRInfo` struct.

**AC-SSE-007** (binary): `git diff main -- internal/statusline/renderer.go | grep -E "^[+-].*(renderPRSegment|isPREnabled|prReviewStateColor)" | grep -v "^[+-]---" | grep -v "^[+-]+++"` returns 0 lines (no insertions or deletions on those functions). Existing test `TestRenderPRSegment` (if present) continues to PASS.

### REQ-SSE-008 — 1M context handoff threshold tightening

The HARD rule in `.claude/rules/moai/workflow/context-window-management.md` SHALL be updated so the 1M model-class threshold reads **50%** (operational ceiling ~500,000 tokens), replacing the prior 75% (~750,000 tokens). The 200K model-class threshold (90% / ~180,000 tokens) SHALL remain unchanged.

**AC-SSE-008** (binary): `grep -E "1M.*75|750,?000.*tokens" .claude/rules/moai/workflow/context-window-management.md` returns 0 matches. `grep -E "1M.*50|500,?000.*tokens" .claude/rules/moai/workflow/context-window-management.md` returns ≥ 2 matches (table row + prose reference).

### REQ-SSE-009 — Mirror threshold change to session-handoff.md

The Trigger #1 table in `.claude/rules/moai/workflow/session-handoff.md` SHALL be updated so the "1M context model (Opus 4.7)" row reads **50%** (~500,000 tokens). The cross-reference text at the bottom of the file SHALL be updated to read "1M = 50%, 200K = 90%".

**AC-SSE-009** (binary): `grep -E "1M.*75|750,?000.*tokens" .claude/rules/moai/workflow/session-handoff.md` returns 0 matches. `grep -E "1M.*50|500,?000.*tokens" .claude/rules/moai/workflow/session-handoff.md` returns ≥ 2 matches.

### REQ-SSE-010 — Mirror threshold change to zone-registry.md

Entries CONST-V3R5-022 and CONST-V3R5-025 in `.claude/rules/moai/core/zone-registry.md` (which quote the 75% threshold verbatim) SHALL be updated to quote 50% verbatim for the 1M model class. The 200K (90%) language in the same entries SHALL remain unchanged.

**AC-SSE-010** (binary): `grep -A 2 "id: CONST-V3R5-022" .claude/rules/moai/core/zone-registry.md | grep -E "75%|750" ` returns 0 matches. Same grep on `id: CONST-V3R5-025` returns 0 matches. `grep -A 2 "id: CONST-V3R5-022" .claude/rules/moai/core/zone-registry.md | grep -E "50%"` returns ≥ 1 match.

### REQ-SSE-011 — Code comment version correction (side-fix)

The code comments at `internal/statusline/types.go:68-69` (StdinData fields `Effort` and `Thinking`) SHALL cite "Claude Code v2.1.139+" instead of "v2.1.122+". The comments at lines 100, 103, 109, 111, 230, 231 (EffortInfo struct, ThinkingInfo struct, StatusData struct fields) SHALL be updated to match.

**AC-SSE-011** (binary): `grep -n "v2.1.122" internal/statusline/types.go` returns 0 matches. `grep -n "v2.1.139" internal/statusline/types.go` returns ≥ 6 matches.

## 4. Risks

| ID | Risk | Mitigation |
|----|------|------------|
| R-SSE-001 | Threshold tightening 75% → 50% may surface premature `/clear` recommendations and confuse users mid-task. | Tier S scope — only update the HARD rule text. The orchestrator runtime detection logic adapts automatically via `context-window-management.md` § Detection Heuristics. Documented in PLAN.md M3. |
| R-SSE-002 | New segments (`repo`, `long_context`, `handoff_guide`) may push statusline width past terminal column limits on narrow terminals. | All three new segments are default-on AND graceful-empty (no output when data absent). Width regression is bounded by stdin payload — if user has narrow terminal AND v2.1.146 stdin, mitigation is `statusline.yaml` opt-out. Documented in PLAN.md M2. |
| R-SSE-003 | `git diff` regression guard on `renderPRSegment` (AC-SSE-007) may produce false positives if a parallel session reformats the file. | PRESERVE LIST hygiene enforced by orchestrator. If diff fires falsely, manager-develop revert-and-retry. Acceptable Tier S risk. |

## 5. Out of Scope

### 5.1 Out of Scope — Other v2.1.146 stdin fields

The following fields are present in v2.1.146 stdin payload but NOT mapped by this SPEC. Reserved for follow-up SPECs:
- `session_name` — display string like "moai-adk-go / main"
- `fast_mode` — undocumented boolean (purpose unclear; needs Anthropic clarification before mapping)
- `workspace.added_dirs` — array of additional context directories
- `cost.total_api_duration_ms` — cumulative API duration
- `vim.mode` — Vim editing mode indicator
- `agent.name` — active sub-agent name
- `worktree.{name,path,branch,original_cwd,original_branch}` — worktree metadata (5 sub-fields)

### 5.2 Out of Scope — Task source migration

`internal/statusline/task.go` has a known chicken-half schema/path mismatch (legacy file-based task collector reads from `.moai/state/` while v2.1.145+ provides task source via Stop/SubagentStop hook channel). Fixing this requires a different stdin channel design and is out of scope for the statusline stdin enrichment SPEC. Separate SPEC candidate.

### 5.3 Out of Scope — v2.1.146 upgrade itself

This SPEC documents the upgrade prerequisite but does NOT perform the Claude Code v2.1.146 upgrade on the user's machine. The user upgrades manually via their package manager. The implementation works correctly on v2.1.145 (repo + exceeds_200k_tokens + handoff_guide all functional via stdin schema already present in v2.1.139+/v2.1.145+) and on v2.1.146+ (additional `pr` field handled by pre-existing REQ-SSE-007 regression guard).

### 5.4 Out of Scope — Renderer 5-line "full" layout cleanup

The pre-existing `full` mode rendering issues observed in commit `e71f1aa54` baseline (6 pre-existing test failures in `TestRenderFullV3*` + `TestIntegration_*`) are covered by SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001 (separate SPEC). This SPEC does NOT touch full mode renderer code paths.

## 6. Constraints (inherited)

- C-SSE-001 (inherited from SPEC-V3R5-STATUSLINE-V2145-001): Default-on segment activation MUST follow `isPREnabled()` pattern (unset key → enabled).
- C-SSE-002 (inherited from CLAUDE.md §1 HARD rules): No XML tags in user-facing output. All new segment strings MUST be Markdown-compatible.
- C-SSE-003 (inherited from CLAUDE.md §9): Code comments in `code_comments: ko` per `language.yaml`. Update REQ-SSE-011 corrections accordingly.
- C-SSE-004 (LEAN Tier S): Maximum 2 SPEC artifact files (spec.md + plan.md). NO acceptance.md (AC inlined here per WORKFLOW-LEAN-001).
- C-SSE-005 (Late-Branch policy): plan-phase commits go to main directly; NO feature branch creation here. PR cherry-picked at sync-phase.

## 7. Dependencies

- Depends on SPEC-V3R5-STATUSLINE-V2145-001 (merged `fb3d1e22b`): provides `PRInfo`, `RepoInfo`, `isPREnabled` pattern.
- Depends on SPEC-V3R5-WORKFLOW-LEAN-001 (merged `c0eb30da6`): provides Tier S 2-artifact form + 0.75 threshold + inline AC.
- Depends on SPEC-V3R5-LATE-BRANCH-001 (merged `664cd6eae`): provides Late-Branch workflow for plan-phase main-direct commits.
- No upstream/downstream conflicts.

## 8. References

- Official statusline docs: `https://code.claude.com/docs/en/statusline` (verbatim Available data table)
- Anthropic CHANGELOG.md (raw): v2.1.139 + v2.1.146 entries
- User stdin sample: `~/.moai/cache/statusline_debug.log` 2026-05-20 18:27
- Sibling SPEC: `.moai/specs/SPEC-V3R5-STATUSLINE-V2145-001/` (PR segment precedent)
- Tier S workflow: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
