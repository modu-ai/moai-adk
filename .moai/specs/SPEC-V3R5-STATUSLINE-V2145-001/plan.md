# SPEC-V3R5-STATUSLINE-V2145-001 Implementation Plan

## 1. Milestones (priority-ordered)

| Milestone | Priority | Class | Scope | Dependencies |
|-----------|----------|-------|-------|--------------|
| M1 — Disappearing hotfix | High | Hotfix (shell only) | 2 shell wrappers + dead-code removal | none |
| M2 — PR segment addition | Medium | Feature (Go) | `internal/statusline/` types + builder + renderer + tests + config schema | M1 (independent in mechanics; bundled in PR ordering for cohesion) |
| M3 — docs-site 4-locale sync | Medium | Docs | 4 markdown files under `docs-site/content/{ko,en,ja,zh}/advanced/` | M2 (segment must exist before documenting) |

Phase ordering: M1 first (fastest user-visible relief, zero Go code change, low risk). M2 second (introduces v2.1.145 PR field adoption). M3 last (documents the M2 feature for end-users across all 4 locales in the same PR).

All three milestones SHOULD ship in a single PR to minimize CI overhead and to honor CLAUDE.local.md §17.3 same-PR 4-locale obligation for M3. If complexity dictates split, the boundary is M1 ↔ M2-M3 (M2 and M3 stay together).

## 2. File Modification Table

| # | File | Milestone | LOC delta (est.) | Change type | Risk |
|---|------|-----------|------------------|-------------|------|
| 1 | `internal/template/templates/.moai/status_line.sh.tmpl` | M1 | -4 / +1 | template — DEBUG default 0 + remove 2 dead `echo ""` pairs | Low |
| 2 | `.moai/status_line.sh` | M1 | -10 / +2 | rendered wrapper — sync with template (see §3 Decision D1) | Low |
| 3 | `internal/statusline/types.go` | M2 | +18 / -0 | extend `StdinData.PR`, add `PRInfo`, add `RepoInfo`, add `WorkspaceInfo.Repo`, add `SegmentPR` constant | Low (additive) |
| 4 | `internal/statusline/builder.go` | M2 | +10 / -0 | wire PR segment into composition pipeline | Low |
| 5 | `internal/statusline/renderer.go` | M2 | +35 / -0 | new render branch for PR segment + review-state color map | Medium (color theme integration) |
| 6 | `internal/statusline/types_test.go` | M2 | +60 / -0 | unmarshal coverage for v2.1.145 fixture | Low |
| 7 | `internal/statusline/builder_test.go` | M2 | +80 / -0 | builder integration coverage | Low |
| 8 | `internal/statusline/renderer_test.go` | M2 | +120 / -0 | renderer table-driven coverage for 5 review_state values + absence + unknown | Low |
| 9 | `internal/template/templates/.moai/config/sections/statusline.yaml.tmpl` OR `.moai/config/sections/statusline.yaml` schema (see §3 Decision D4) | M2 | +1 / -0 | new `segments.pr: false` default | Low |
| 10 | `docs-site/content/ko/advanced/statusline.md` | M3 | +60 / -0 | new section "PR segment" canonical Korean | Low |
| 11 | `docs-site/content/en/advanced/statusline.md` | M3 | +60 / -0 | translation | Low |
| 12 | `docs-site/content/ja/advanced/statusline.md` | M3 | +60 / -0 | translation | Low |
| 13 | `docs-site/content/zh/advanced/statusline.md` | M3 | +60 / -0 | translation | Low |

Approximate aggregate: ~520 LOC additive, ~14 LOC subtractive. 13 files total. Run-phase will verify exact paths via Glob.

## 3. Design Decisions

### D1 — Rendered `.moai/status_line.sh` hardcoded path policy (resolves OQ-1)

Decision: Use `$HOME/go/bin/moai` runtime expansion in the rendered wrapper, NOT `{{posixPath .GoBinPath}}` init-time variable.

Rationale: CLAUDE.local.md §14 explicitly mandates `$HOME` substitution in shell fallback paths (§14 "[HARD] .sh.tmpl 폴백 경로에 `.HomeDir` 금지"). The template's primary path entry (`{{posixPath .GoBinPath}}/moai`) is acceptable for the init-time captured path, but the rendered file should resolve at runtime. The existing template at `internal/template/templates/.moai/status_line.sh.tmpl` already does this correctly for one fallback (`$HOME/.local/bin/moai`); M1 will mirror this pattern for the Go bin path fallback.

Trade-off accepted: This is a backward-incompatible change for users who have `moai` installed at a non-`$HOME/go/bin/` path AND not in `$PATH` AND not at `$HOME/.local/bin/`. The wrapper still tries `command -v moai` first, so any installation that is on `$PATH` is unaffected.

### D2 — PR segment URL inclusion (resolves OQ-2)

Decision: Emit number + review_state only. Do NOT include `pr.url` in the rendered statusline output.

Rationale:
- Statusline is bandwidth-constrained (3-line display, dozens of competing segments). The URL is verbose (~60 chars for typical GitHub URL).
- The PR number is unique per repo + uniquely identifies the PR in the user's local context (they already know they are in `modu-ai/moai-adk-go`).
- Click-to-open via OSC 8 escape sequences is excluded per EXCL-6.
- The URL is still parsed into `PRInfo.URL` so future SPECs can opt to surface it (e.g., in a hover tooltip if a future Claude Code version adds that capability).

### D3 — Review-state color resolution mechanism

Decision: Add a private helper `prReviewStateColor(state string) Color` in `renderer.go` that maps the review_state string to a theme-aware color value via the existing `theme.go` infrastructure (the same approach `effort_thinking` uses for level-to-color mapping).

The mapping table:

| review_state | Theme color key | Fallback (when theme unavailable) |
|---|---|---|
| `approved` | `Success` (green) | ANSI 32 |
| `pending` | `Warning` (yellow) | ANSI 33 |
| `changes_requested` | `Error` (red) | ANSI 31 |
| `draft` | `Muted` (gray) | ANSI 90 |
| (other / empty) | default | unstyled |

Unknown values pass through unstyled per the raw-passthrough principle established in REQ-CC2122-004.

### D4 — Config schema location (resolves OQ-4)

Decision: Add `segments.pr: false` to BOTH locations:
1. `internal/template/templates/.moai/config/sections/statusline.yaml.tmpl` (if file exists) — for new project scaffolds via `moai init`
2. The Go struct that parses statusline config (likely `internal/config/statusline.go` — verify in run-phase via Grep)

Existing users with a rendered `.moai/config/sections/statusline.yaml` will have the field default to `false` via Go's zero-value semantics, so no `moai update` sync is required. New project scaffolds get the explicit default.

Rationale: This matches the precedent of the existing `worktree: true`, `git_branch: true` keys in `.moai/config/sections/statusline.yaml`. Default `false` ensures zero-regression (M2 is purely opt-in).

### D5 — Pre-exec vs post-exec `echo ""` removal scope

Decision: Remove BOTH `echo ""` lines flanking each `exec moai statusline` call in `status_line.sh.tmpl`. Document `statusLine.padding: N` in M3 docs as the canonical replacement.

Rationale: post-exec `echo ""` is unreachable dead code (exec replaces the shell). pre-exec `echo ""` would emit a blank line as the first line of statusline output — when Claude Code renders a multi-line statusline, this blank line shifts the visible content down by one row and may collide with the 3-line layout contract. The canonical Claude Code mechanism for vertical padding is the `statusLine.padding` settings.json key (code.claude.com/docs/en/statusline).

### D6 — Brownfield strategy for the 4 always-on segments + opt-in segments

Decision: PRESERVE all existing segments exactly. M2 ADDS one new segment (`pr`) and one new struct field (`PR *PRInfo` on `StdinData`). No existing field or constant is renamed, removed, or repurposed.

This honors the moai-foundation-quality TRUST 5 "Consistency" pillar and prevents downstream breakage in `builder.go`'s segment composition order.

### D7 — Conditional PR rendering when stdin lacks pr field

Decision: If stdin's `pr` is nil/absent AND `segments.pr: true`, the renderer emits no PR segment. No placeholder is rendered. This matches the V3R3-FALLBACK precedent for missing `model` data (REQ-SF-002 third bullet "둘 다 없으면 model name 미표시").

## 4. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| R1 — Rendered/template drift (rendered `.moai/status_line.sh` already differs from template; running `moai update` may overwrite user changes) | Medium | Medium | M1 is a 2-file synchronized edit; sync after rendering must produce byte-identical output (verified by AC-SLV-001 / AC-SLV-002 grep). Document in M3 docs that `DEBUG_STATUSLINE=1` is the new opt-in path. |
| R2 — Padding regression (users who relied on pre-exec `echo ""` see statusline now starts higher) | Medium | Low | M3 docs explicitly call out `statusLine.padding: N` as the canonical replacement. The pre-exec `echo ""` was already unreliable (only fired on the template re-render path, not the user-edited rendered path), so most users have not seen the padding. |
| R3 — PR segment render order conflicts with other 3-line layout segments | Low | Medium | Place PR segment in the same row as `git_branch` and `worktree` (line 2 — git/repo context). renderer_test.go MUST include a multi-segment-on-line-2 integration test. |
| R4 — v2.1.145 stdin schema drift (Anthropic changes field names post-release) | Low | High | The struct uses Go's lenient JSON unmarshal (unknown fields tolerated). If Anthropic renames `review_state` → `state`, a small follow-up SPEC adds the alias via custom `UnmarshalJSON`. Same pattern as `ModelInfo.UnmarshalJSON` (lines 120-138 of types.go). |
| R5 — docs-site path mismatch (no `advanced/statusline.md` exists) | Low | Low | Run-phase Glob discovers the canonical path. Plan §2 lists `advanced/statusline.md`; fall back to creating new files if absent. CLAUDE.local.md §17.3 covers the 4-locale obligation regardless. |
| R6 — golangci-lint baseline regression on new test files | Low | Medium | Follow existing `internal/statusline/*_test.go` patterns (table-driven, `httptest`-free, `t.TempDir()` for any fs needs). Run `golangci-lint run` pre-commit per `CLAUDE.local.md` §4. |
| R7 — Coverage drops below 85% threshold on changed files (M2) | Low | High | Plan §2 LOC estimates 525+ test LOC against 63 production LOC (~8:1 ratio) — table-driven tests cover all 5 review_state values plus absence + unknown + URL-empty paths. |

## 5. Traceability Matrix (REQ ↔ AC ↔ File)

| REQ | AC | Primary file(s) | Verification |
|-----|-----|-----------------|--------------|
| REQ-SLV-001 | AC-SLV-001 | `.moai/status_line.sh`, `internal/template/templates/.moai/status_line.sh.tmpl` | grep `DEBUG_STATUSLINE:-0` |
| REQ-SLV-002 | AC-SLV-002 | `.moai/status_line.sh`, template | grep `if [ "${DEBUG_STATUSLINE:-0}" = "1" ]` pattern |
| REQ-SLV-003 | AC-SLV-003 | `internal/template/templates/.moai/status_line.sh.tmpl` | grep absence of `echo ""` around `exec moai statusline` |
| REQ-SLV-004 | AC-SLV-004 | `docs-site/content/{ko,en,ja,zh}/advanced/statusline.md` | grep `statusLine.padding` in all 4 locales |
| REQ-SLV-005 | AC-SLV-005 | `.moai/status_line.sh` | grep absence of `/Users/goos/` |
| REQ-SLV-010 | AC-SLV-010 | `internal/statusline/types.go` | `go vet ./internal/statusline/...`; struct field reflection test |
| REQ-SLV-011 | AC-SLV-011 | `internal/statusline/types.go` | reflection test for `WorkspaceInfo.Repo` |
| REQ-SLV-012 | AC-SLV-012 | `internal/statusline/builder.go`, config schema | `go test -run TestPRSegmentDefaultOff ./internal/statusline/...` |
| REQ-SLV-013 | AC-SLV-013 | `internal/statusline/renderer.go` | `go test -run TestPRSegmentFormat ./internal/statusline/...` |
| REQ-SLV-014 | AC-SLV-014 | `internal/statusline/renderer.go` | `go test -run TestPRSegmentReviewStateColor ./internal/statusline/...` |
| REQ-SLV-015 | AC-SLV-015 | `internal/statusline/renderer.go` | `go test -run TestPRSegmentAbsence ./internal/statusline/...` |
| REQ-SLV-016 | AC-SLV-016 | `internal/statusline/types.go`, `builder.go` | grep `SegmentPR` constant + builder branch |
| REQ-SLV-017 | AC-SLV-017 | `internal/statusline/{types,builder,renderer}_test.go` | `go test -cover ./internal/statusline/...` ≥ 85% |
| REQ-SLV-020 | AC-SLV-020 | `docs-site/content/ko/advanced/statusline.md` | grep `## PR 세그먼트` section heading |
| REQ-SLV-021 | AC-SLV-021 | all 4 docs-site files | grep PR section in en/ja/zh |
| REQ-SLV-022 | AC-SLV-022 | docs-site CI | `scripts/docs-i18n-check.sh` exit 0; URL blacklist grep clean |

Traceability completeness: 17 REQ → 17 AC = 100% (1:1 mapping). All REQs have a primary file and a verification command.

## 6. Dependencies

- **Upstream**: Claude Code v2.1.145 release (already shipped 2026-05-19 per `.moai/research/cc-update-20260520.md` §1 Metadata). No version pin required — feature degrades gracefully on older Claude Code versions per REQ-SLV-015.
- **Internal**: None. M1 is shell-only, M2 follows the v2.1.122 effort/thinking + v2.1.97 worktree precedents, M3 is standard docs-site i18n.
- **Toolchain**: Go 1.23+ (project standard), golangci-lint, `scripts/docs-i18n-check.sh` (for M3), `make build` (for M1 template → embedded.go regeneration).

## 7. Phase Sequencing (run-phase)

1. **M1 Phase A**: Edit `internal/template/templates/.moai/status_line.sh.tmpl` (template source). Run `make build` to regenerate `internal/template/embedded.go`. Verify embedded contents match the new template.
2. **M1 Phase B**: Edit `.moai/status_line.sh` (rendered file) to match the new template byte-for-byte (excluding the `{{posixPath .GoBinPath}}` template substitution which expands to `$HOME/go/bin` per D1).
3. **M2 Phase A**: Extend `internal/statusline/types.go` with `PRInfo`, `RepoInfo`, `WorkspaceInfo.Repo`, `StdinData.PR`, `SegmentPR` constant. Run `go vet` and `go test ./internal/statusline/...` to verify zero-regression on existing tests.
4. **M2 Phase B**: Write `types_test.go` v2.1.145 unmarshal fixture. RED first per moai-workflow-tdd (failing test).
5. **M2 Phase C**: Implement `builder.go` PR segment wiring. Implement `renderer.go` PR render branch with review-state color helper. GREEN.
6. **M2 Phase D**: Write `builder_test.go` and `renderer_test.go` table-driven tests covering all 5 review_state values + absence + unknown + URL-empty. REFACTOR for clarity.
7. **M2 Phase E**: Add `segments.pr: false` default to config schema (template + Go struct).
8. **M3**: Write Korean canonical section. Translate to en/ja/zh. Run `scripts/docs-i18n-check.sh`. Verify URL blacklist clean.
9. **Pre-commit verification batch** (per agent-common-protocol §Parallel Execution): single-turn 7-command verification (go test, coverage, lint, grep for hardcoded paths, grep for AskUserQuestion misuse, CLI smoke, sentinel scan).

## 8. Quality Gates

- **TRUST 5 Tested**: ≥85% coverage on changed files in `internal/statusline/`; new tests fail RED before GREEN.
- **TRUST 5 Readable**: Function/variable naming follows existing patterns (`renderPRSegment`, `isPREnabled`, `prReviewStateColor`).
- **TRUST 5 Unified**: gofmt + goimports clean; golangci-lint baseline does not regress.
- **TRUST 5 Secured**: No secrets in test fixtures; `pr.url` is opaque string (never executed).
- **TRUST 5 Trackable**: Conventional Commits per CLAUDE.local.md §4; commits scoped per milestone (`feat(statusline): ...` for M2, `fix(statusline): ...` for M1, `docs(statusline): ...` for M3).
- **LSP Quality Gate (run-phase)**: zero errors, zero type errors, zero lint errors per `.moai/config/sections/quality.yaml`.
- **docs-site CI (M3)**: `scripts/docs-i18n-check.sh` PASS; URL blacklist clean.

## 9. Brownfield Strategy

This SPEC is a BROWNFIELD enhancement to a stable production statusline package (`internal/statusline/` has ~2.4 K LOC of production code + ~3.0 K LOC of tests across 16 files). The strategy:

- **PRESERVE**: All existing segment constants (`SegmentModel`, `SegmentContext`, `SegmentOutputStyle`, `SegmentDirectory`, `SegmentGitStatus`, `SegmentClaudeVersion`, `SegmentMoaiVersion`, `SegmentGitBranch`, `SegmentSessionTime`, `SegmentUsage5H`, `SegmentUsage7D`, `SegmentTask`, `SegmentWorktree`, `SegmentEffortThinking`). All existing fields on `StdinData` (`HookEventName`, `SessionID`, `TranscriptPath`, `CWD`, `Model`, `Workspace`, `Cost`, `ContextWindow`, `OutputStyle`, `RateLimits`, `Effort`, `Thinking`, `Version`). All existing tests pass unchanged.
- **EXTEND**: Add `PR *PRInfo` to `StdinData`. Add `Repo *RepoInfo` to `WorkspaceInfo`. Add `SegmentPR` constant. Add new render branch in `renderer.go` gated on `r.isSegmentEnabled(SegmentPR)`.
- **DO NOT RENAME**: Existing function and constant names remain untouched.
- **DO NOT REORDER**: Existing segment render order is preserved.
- **DO NOT MODIFY**: `WorkspaceInfo.GitWorktree` field (already exists at line 145 of `types.go`); no semantic change.

Cross-reference: lessons #21 (cross-platform build tag discipline) does not apply here — no syscall package usage. lessons #19 (CI 3-tier awareness) applies to M3 docs-site CI.

## 10. Out of Scope (cross-reference to spec.md §6)

This plan does NOT cover the items in spec.md §6 Exclusions. In particular:
- EXCL-1 statusline disappearance investigation (separate SPEC if M1 does not resolve)
- EXCL-2 layout redesign
- EXCL-3 other v2.1.145 release-note items (bundled into `SPEC-V3R4-CC2X-ADOPT-002` candidate)
- EXCL-5 PR data caching
- EXCL-6 click-to-open URL

## 11. Post-Implementation Review Checklist (per CLAUDE.md §7 Rule 3)

Upon M2 completion, manager-develop's report MUST include:

- **Potential issues**: render-order conflicts with `git_branch`/`worktree` on line 2; review_state value drift if Anthropic adds new values
- **Suggested additional tests**: integration test with full v2.1.145 stdin fixture; benchmark for render latency (PR segment adds 1 strings.Builder write + 1 lookup)
- **Known limitations**: M2 does not detect "out-of-date PR" (e.g., when local branch is behind remote PR HEAD); review_state is point-in-time
- **Recommendations**: capture `pr.url` in `PRInfo` for future OSC 8 hyperlink emit (deferred per EXCL-6)
