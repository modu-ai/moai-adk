---
id: SPEC-V3R5-STATUSLINE-V2145-001
title: "Statusline v2.1.145 alignment — disappearing fix + PR segment"
version: "0.3.0"
status: completed
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.5.0"
module: "internal/statusline"
lifecycle: spec-anchored
tags: "statusline, claude-code, v2.1.145, ux, pr-badge, hotfix"
---

# SPEC-V3R5-STATUSLINE-V2145-001: Statusline v2.1.145 alignment — disappearing fix + PR segment

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-05-20 | 0.1.0 | GOOS Kim | Initial draft — 3 milestones (M1 disappearing hotfix + M2 PR segment + M3 docs-site 4-locale) based on `.moai/research/cc-update-20260520.md` Tier 2 item T2-3 + /97-release-update diagnostic findings |

## 1. Overview

### 1.1 Background

Two independent statusline issues converged during the 2026-05-20 `/97-release-update` diagnostic session and are bundled here as a coordinated alignment SPEC. Both are tied to the Claude Code v2.1.145 upgrade timeline.

**Issue A — Intermittent statusline disappearance.** Production rendered wrapper `.moai/status_line.sh` ships with `DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-1}` (default 1, always-on). This forces a `python3 -m json.tool` fork plus log write on every render. macOS `python3` cold-start adds 50-250 ms per render, which collides with the official Claude Code statusline contract (300 ms debounce; in-flight execution cancelled when a new trigger arrives — code.claude.com/docs/en/statusline). Evidence: `~/.moai/cache/statusline_debug.log` accumulated 2031 lines / 56 KB in production without any explicit `export DEBUG_STATUSLINE=1`, confirming default-on. Secondary defects compound: the template `.tmpl` contains dead code `echo ""` AFTER `exec moai statusline` (exec replaces the shell, echo never runs), and the pre-exec `echo ""` would emit a blank line as the first row of a multi-line statusline once template re-renders to `.moai/status_line.sh`. Tertiary drift: the rendered file hardcodes `/Users/goos/go/bin/moai` which violates CLAUDE.local.md §14 portability rule, while the template correctly uses `{{posixPath .GoBinPath}}`.

**Issue B — v2.1.145 PR field not parsed.** Claude Code v2.1.145 added `pr.number`, `pr.url`, `pr.review_state` to the statusline stdin JSON (release note line: "Status line JSON input now includes GitHub repo and PR information when detected"). Direct verification: 28 occurrences of `workspace.repo.{host,owner,name}` already flowing into the current session's `~/.moai/cache/statusline_debug.log`. The `internal/statusline/types.go` `StdinData` struct does NOT include `pr` or `workspace.repo` fields — `grep -rn -E '"pr"|"repo"|workspace\.repo' internal/statusline/` returns zero matches. moai-adk-go's workflow is PR-centric (every SPEC produces plan-PR + run-PR + sync-PR, often with plan-auditor iterations + admin override), so surfacing `#1023 ⌥pending` in the statusline carries high daily UX value.

### 1.2 Goals

- Eliminate statusline disappearance by removing the always-on debug fork and dead-code echo pairs (M1)
- Adopt v2.1.145 PR fields and add an opt-in `pr` segment with review-state color coding (M2)
- Document the new opt-in segment across all 4 docs-site locales per CLAUDE.local.md §17.3 (M3)
- Maintain zero-regression on the 4 existing always-on segments (model / context / directory / git_status / git_branch + opt-in worktree, effort_thinking, output_style, claude_version, moai_version, session_time, usage_5h, usage_7d, task)

### 1.3 Non-goals

- See §6 Exclusions for the canonical list

## 2. Scope and References

### 2.1 In-scope files

| File | Milestone | Change category |
|------|-----------|-----------------|
| `internal/template/templates/.moai/status_line.sh.tmpl` | M1 | template source — DEBUG default 1→0, remove dead `echo ""` pairs |
| `.moai/status_line.sh` | M1 | rendered wrapper — sync with template; hardcoded path policy decision deferred to plan.md §3 |
| `internal/statusline/types.go` | M2 | extend `StdinData` with `PR`, extend `WorkspaceInfo` with `Repo` and `GitWorktree` (already present) |
| `internal/statusline/builder.go` | M2 | PR segment builder + enable gating |
| `internal/statusline/renderer.go` | M2 | render PR segment with review-state color map + `SegmentPR` constant |
| `internal/statusline/types_test.go` | M2 | unmarshal coverage for v2.1.145-shape JSON |
| `internal/statusline/builder_test.go` | M2 | builder integration coverage |
| `internal/statusline/renderer_test.go` | M2 | renderer table-driven coverage for review-state color codes |
| `internal/template/templates/.moai/config/sections/statusline.yaml.tmpl` (if present) OR `.moai/config/sections/statusline.yaml` schema | M2 | new opt-in `segments.pr: false` default |
| `docs-site/content/ko/advanced/statusline.md` | M3 | canonical Korean source |
| `docs-site/content/en/advanced/statusline.md` | M3 | English translation |
| `docs-site/content/ja/advanced/statusline.md` | M3 | Japanese translation |
| `docs-site/content/zh/advanced/statusline.md` | M3 | Chinese translation |

Verify exact docs-site file paths via Glob during run-phase. If `statusline.md` does not exist under `advanced/`, fall back to `advanced/cli/` or create new files per docs-site conventions.

### 2.2 References

- `.moai/research/cc-update-20260520.md` — release-update research file for v2.1.143..v2.1.145. Tier 2 item T2-3 (Status line JSON input includes GitHub repo + PR info) is the upstream trigger for M2.
- `.moai/cache/cc-release-notes-20260520.txt` — raw CHANGELOG verbatim.
- `code.claude.com/docs/en/statusline` — official statusline JSON schema reference (debounce contract).
- `internal/statusline/types.go:55-71` — current `StdinData` struct (the v2.1.145-shape extension target).
- Prior SPECs (style and ID-convention reference): `SPEC-V3R3-STATUSLINE-FALLBACK-001`, `SPEC-CC2122-STATUSLINE-001`, `SPEC-CC297-001` (worktree segment precedent), `SPEC-STATUSLINE-001`, `SPEC-STATUSLINE-002`.
- CLAUDE.local.md §14 (hardcoding prevention — `$HOME` fallback rule), §17.3 (docs-site same-PR 4-locale obligation).
- Memory: `project_statusline_disappearance_fix.md` (3-root-cause diagnosis precedent from V3R3-FALLBACK).

## 3. Requirements (EARS)

### 3.1 Milestone 1 — Disappearing hotfix (Hotfix)

#### REQ-SLV-001: DEBUG_STATUSLINE default off (Event-driven)

**When** `internal/template/templates/.moai/status_line.sh.tmpl` is rendered to `.moai/status_line.sh`, the system **shall** emit a wrapper where `DEBUG_STATUSLINE` resolves to `0` by default (i.e., `DEBUG_STATUSLINE=${DEBUG_STATUSLINE:-0}`).

#### REQ-SLV-002: Debug log only on explicit opt-in (State-driven)

**While** `DEBUG_STATUSLINE` is unset or set to `0`, the wrapper **shall** skip the `python3 -m json.tool` fork and **shall not** write to `~/.moai/cache/statusline_debug.log`. **While** `DEBUG_STATUSLINE` is set to `1` (explicit opt-in only), the wrapper **shall** continue to log stdin JSON for diagnostic purposes.

#### REQ-SLV-003: Remove dead echo code (Ubiquitous)

The system **shall** remove the `echo ""` line that appears AFTER `exec moai statusline` in `status_line.sh.tmpl` (unreachable post-exec). The system **shall** also remove the `echo ""` line that appears BEFORE `exec moai statusline` (would emit a leading blank line as the first row of a multi-line statusline when template re-renders).

#### REQ-SLV-004: Padding migration to settings.json (Optional)

**Where** the user previously relied on `echo ""` for visual padding around the statusline, the system **shall** document `statusLine.padding: N` in `.claude/settings.json` as the canonical replacement per code.claude.com/docs/en/statusline. This documentation appears in the M3 docs-site updates.

#### REQ-SLV-005: Template-rendered path policy (State-driven)

**While** `.moai/status_line.sh` is rendered from template at `moai init` / `moai update` time, the fallback path entries **shall** use either `{{posixPath .GoBinPath}}` (init-time variable, current behavior) or `$HOME/go/bin/moai` (runtime expansion). Plan.md §3 carries the decision; the SPEC requires only that the rendered output not contain a user-specific absolute path like `/Users/goos/...` that violates CLAUDE.local.md §14.

### 3.2 Milestone 2 — PR segment addition (Feature)

#### REQ-SLV-010: Adopt v2.1.145 PR stdin fields (Event-driven)

**When** Claude Code v2.1.145+ sends stdin JSON containing a `pr` object, `internal/statusline/types.go` `StdinData` **shall** unmarshal it into a `PR *PRInfo` field where `PRInfo` exposes `Number int` (`json:"number"`), `URL string` (`json:"url"`), and `ReviewState string` (`json:"review_state"`).

#### REQ-SLV-011: Adopt workspace.repo stdin fields (Event-driven)

**When** stdin JSON contains a `workspace.repo` object, `WorkspaceInfo` **shall** unmarshal it into a `Repo *RepoInfo` field where `RepoInfo` exposes `Host string` (`json:"host"`), `Owner string` (`json:"owner"`), `Name string` (`json:"name"`). The pre-existing `GitWorktree string` field in `WorkspaceInfo` is preserved unchanged.

#### REQ-SLV-012: PR segment opt-in default off (State-driven)

**While** `.moai/config/sections/statusline.yaml` `segments.pr` is unset or `false`, the renderer **shall not** emit a PR segment, even when stdin contains `pr.*` data. The default value at template scaffold time **shall** be `false` (zero-regression for existing users).

#### REQ-SLV-013: PR segment render format (Event-driven)

**When** `segments.pr: true` AND stdin contains a non-nil `pr` object with `pr.number > 0`, the renderer **shall** emit a segment of the form ` #<number> ⌥<state>` where `<state>` is the lowercase review_state string. Empty `pr.url` is tolerated (segment still renders from `number` alone).

#### REQ-SLV-014: PR segment review-state color coding (State-driven)

**While** rendering the PR segment, the renderer **shall** apply ANSI color codes per the review_state mapping below. Unknown review_state values **shall** fall through to the default (no color) per the raw-passthrough principle already established in REQ-CC2122-004:

| review_state value | Color (semantic) | ANSI hint |
|---|---|---|
| `approved` | green | success |
| `pending` | yellow | warning |
| `changes_requested` | red | error |
| `draft` | gray | muted |
| (other / empty) | default | unstyled |

The exact theme palette resolution (catppuccin-mocha vs others) is delegated to the existing theme.go infrastructure; plan.md §3 enumerates the implementation approach.

#### REQ-SLV-015: PR segment absence handling (Unwanted)

**If** stdin's `pr` object is null, absent, or has `pr.number == 0`, **then** the renderer **shall not** emit a PR segment regardless of the `segments.pr` config value. The system **shall not** emit placeholder text such as `#N/A` or `#0`.

#### REQ-SLV-016: Segment constant + builder wiring (Ubiquitous)

The system **shall** introduce a `SegmentPR = "pr"` constant in `types.go` and a corresponding builder function in `builder.go` that mirrors the existing segment composition pattern (the same shape as `SegmentWorktree`, `SegmentEffortThinking`).

#### REQ-SLV-017: Test coverage for PR segment (Ubiquitous)

The system **shall** include table-driven unit tests under `internal/statusline/` covering:
- Unmarshal of a v2.1.145-shape JSON fixture with `pr.*` and `workspace.repo.*` populated (types_test.go)
- Builder behavior when `segments.pr` is true vs false (builder_test.go)
- Renderer color-code emission for each review_state value, including unknown values (renderer_test.go)
- Absence behavior per REQ-SLV-015 (renderer_test.go)

Coverage target: ≥85% on changed files per moai-foundation-quality TRUST 5 gate.

### 3.3 Milestone 3 — docs-site 4-locale sync (Docs)

#### REQ-SLV-020: Korean canonical update (Ubiquitous)

The system **shall** add a new section to `docs-site/content/ko/advanced/statusline.md` (or equivalent path determined at run-phase) documenting:
- The new `pr` segment opt-in (`segments.pr: true`)
- The `pr.*` stdin field schema (v2.1.145+ Claude Code requirement)
- The review_state color mapping per REQ-SLV-014
- The `statusLine.padding: N` migration guidance per REQ-SLV-004

#### REQ-SLV-021: 4-locale parity (Ubiquitous)

The system **shall** mirror REQ-SLV-020 content across all 4 locales (ko/en/ja/zh) within the same PR per CLAUDE.local.md §17.3. The Korean version is the canonical source; en/ja/zh are translations.

#### REQ-SLV-022: Docs CI compliance (Event-driven)

**When** the M3 PR is opened, `scripts/docs-i18n-check.sh` (or equivalent docs-site CI guard) **shall** PASS. The URL blacklist (`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr`) **shall not** be referenced — canonical host is `adk.mo.ai.kr`.

## 4. Constraints

- **Zero-regression on M1**: The 4 existing always-on segments (model, context, directory, git_status, git_branch) and the 7 opt-in segments (worktree, effort_thinking, output_style, claude_version, moai_version, session_time, usage_5h, usage_7d, task) must render identically before/after M1+M2.
- **Brownfield discipline**: M2 ADDS a new segment; does not modify any existing segment's render contract or color scheme.
- **v2.1.145 dependency**: M2's runtime visibility depends on Claude Code v2.1.145+ stdin. M2's struct extension must remain backward-compatible — when stdin lacks `pr.*` (older Claude Code), `PR` resolves to nil and REQ-SLV-015 absence handling applies.
- **CLAUDE.local.md §14 portability**: Rendered `.moai/status_line.sh` must not contain user-specific absolute paths. Either init-time template variable or `$HOME` runtime expansion is acceptable.
- **CLAUDE.local.md §17.3 same-PR 4-locale obligation**: M3 must ship all 4 locales in the same PR. Korean-only commits are blocked.
- **No time estimates**: per agent-common-protocol §Time Estimation, plan.md uses priority labels and phase ordering only.
- **TRUST 5**: ≥85% coverage on changed files per package (Tested); golangci-lint clean (Readable/Unified); no secret exposure (Secured); Conventional Commits (Trackable).

## 5. Architecture Approach

Light architectural notes; the binding implementation details live in plan.md §3.

- **M1**: 3 file deltas in shell-only files (template + rendered + optional dead-code removal). No Go code change. Mechanism: template edit → `make build` → wrapper template re-renders during user's next `moai update`.
- **M2**: 5-7 file deltas in `internal/statusline/`. Mechanism mirrors the v2.1.122 effort/thinking precedent (`SPEC-CC2122-STATUSLINE-001`) and the v2.1.97 worktree precedent (`SPEC-CC297-001`): extend `StdinData` struct + add a new segment constant + add a render branch in `renderer.go` + table-driven tests.
- **M3**: 4 docs-site files (1 canonical Korean + 3 translations). Mechanism: standard docs-site i18n workflow with `scripts/docs-i18n-check.sh` verification.

## 6. Exclusions (What NOT to Build)

- **EXCL-1**: `statusLine` config mid-session disappearance investigation. This SPEC addresses default-on DEBUG fork as the established root cause of the disappearing-statusline symptom; if symptoms persist post-M1, a separate investigation SPEC will be opened.
- **EXCL-2**: Statusline layout / UI redesign. M2 ADDS a single new segment using existing layout primitives; it does NOT reflow the 3-line layout, change theme, or alter the existing segment order.
- **EXCL-3**: Other v2.1.145 release-note items. Items T1-1 (hook field extensions), T1-2 (permission-prompt bypass closure audit), T1-7 (worktree.bgIsolation docs), T1-8 (stop hook block cap) are bundled into the separate `SPEC-V3R4-CC2X-ADOPT-002` (candidate per `.moai/research/cc-update-20260520.md` §7).
- **EXCL-4**: New display modes. The `default` / `full` mode set is unchanged.
- **EXCL-5**: PR data caching. The PR segment is a pure render of stdin's `pr` object. There is no cache file (analogous to `last-model.txt` from V3R3-FALLBACK) for PR data because PR data is short-lived and reliably re-sent by Claude Code on each render trigger.
- **EXCL-6**: Click-to-open URL behavior. Some terminals support hyperlinks, but emitting OSC 8 escape sequences would be a separate visual feature; this SPEC emits text only.
- **EXCL-7**: `MOAI_LAST_MODEL` style env-var fallback for PR. Not applicable — see EXCL-5.
- **EXCL-8**: Effort/thinking segment unification. The `effort_thinking` combined segment remains as a separate concern.

## 7. Open Questions

OQ-1 (resolved in plan.md §3): Should the rendered `.moai/status_line.sh` use `$HOME/go/bin/moai` (runtime expansion, fully portable) or keep `{{posixPath .GoBinPath}}` (init-time variable, current template behavior)?

OQ-2 (resolved in plan.md §3): Should the PR segment include the URL (`pr.url`) in the output, or number+state only?

OQ-3 (deferred to a future SPEC, see EXCL-1): If statusline disappearance persists after M1 lands, what is the diagnostic protocol — does Claude Code expose a "last statusline trigger error" introspection path?

OQ-4 (resolved in plan.md §3): Where does the `segments.pr` opt-in key live — `.moai/config/sections/statusline.yaml` or a new `statusline.yaml.tmpl`?
