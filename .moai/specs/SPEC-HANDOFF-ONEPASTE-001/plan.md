---
id: SPEC-HANDOFF-ONEPASTE-001
title: "Session Handoff 1-Paste — implementation plan"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow + .moai/config/sections"
lifecycle: spec-anchored
tags: "handoff, session, auto-resume, doctrine, config, goal-first"
---

# SPEC-HANDOFF-ONEPASTE-001 — Implementation Plan (Tier M)

## §A Context

100% doctrine + config SPEC. Wires the already-verified auto-resume consumption infrastructure (SPEC-HANDOFF-AUTORESUME-001, origin `caf0146c4`) to an emission obligation in the session-handoff SSOT, and flips the local `handoff.yaml` to `mode: auto`. Zero Go changes (REQ-OP-015). All normative content lives in spec.md §C; this plan sequences the edits.

Development mode: doctrine/config edits follow the DDD PRESERVE discipline at the document layer — baseline greps recorded before each edit, delta greps after (see acceptance.md §D.3 baseline-delta pattern).

## §B Known Issues / Constraints Inherited

- **Shared checkout has unrelated uncommitted changes** (statusline, manager-docs.md, llm.yaml, template catalog, etc.). ALL commits in every phase MUST be pathspec-scoped (`git add <paths> && git commit -- <paths>`); `git add -A` is prohibited for this SPEC line.
- **session-handoff.md is large (~48KB live)** and carries a parity sentinel with moai.md §8. Any M1 edit that touches skeleton/label/localization areas triggers the sentinel's parity obligation — M1 deliberately avoids those blocks (REQ-OP-014).
- **Template mirrors are 3 files + 1 no-change file.** The template tree is a SUBSET of live in general, but all 3 target doctrine files ARE mirrored (verified at plan-phase: `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`, `.../goal-directive.md`, `.../output-styles/moai/moai.md` all exist; template `handoff.yaml` line 9 = `mode: manual`). Per-target presence is confirmed; no MIRRORED-vs-LIVE-ONLY split needed for this SPEC.
- **A3 verify-before-author**: the auto-flow doctrine section must describe what `handoff_inject_render.go` actually renders (directive-restoration guidance wording), not an assumed rendering. Run-phase reads that file before authoring the flow section.

## §C Pre-flight (run-phase entry checks)

Execute as a single-turn read-only batch before the first M1 edit:

1. `git log --oneline -1` → record base SHA; `git status --porcelain .claude/rules/moai/workflow/session-handoff.md .claude/output-styles/moai/moai.md .claude/rules/moai/workflow/goal-directive.md .moai/config/sections/handoff.yaml` → all four MUST be clean (no unrelated in-flight edits on target files).
2. Anchor greps (content-token anchors, not line numbers):
   - `grep -n "## Post-Paste /goal Follow-up Block" .claude/rules/moai/workflow/session-handoff.md` → 1 match
   - `grep -n "## Auto-Memory Integration" .claude/rules/moai/workflow/session-handoff.md` → 1 match
   - `grep -n "### Session Handoff \[HARD\]" .claude/output-styles/moai/moai.md` → 1 match
   - `grep -n "## MoAI Integration Notes" .claude/rules/moai/workflow/goal-directive.md` → 1 match
3. Baseline counts for delta ACs (record verbatim into progress.md §E.2):
   - `grep -c "moai handoff save" .claude/rules/moai/workflow/session-handoff.md` (expected baseline 0)
   - `grep -c "moai handoff save" .claude/output-styles/moai/moai.md` (expected baseline 0)
   - `grep -c "Implementation Kickoff Approval" .claude/rules/moai/workflow/session-handoff.md` (baseline N₁ — delta ≥ 1 after M1)
4. `Read internal/hook/handoff_inject_render.go` → capture the actual injected-guidance wording for the flow section (A3).
5. Template mirror presence re-verify: `ls internal/template/templates/.claude/rules/moai/workflow/{session-handoff,goal-directive}.md internal/template/templates/.claude/output-styles/moai/moai.md` + `grep -n "mode: manual" internal/template/templates/.moai/config/sections/handoff.yaml`.

## §D Constraints

- REQ-OP-014: do NOT edit inside § Cut-line Marker Specification, the 6-block skeleton fence, § Localization Table, or moai.md §8 translation tables. New sections are additive.
- REQ-OP-015: `git diff --name-only <base>..HEAD -- '*.go'` must stay empty across the whole SPEC line.
- Template neutrality (REQ-OP-011): mirrors carry the normative rules with NO `SPEC-HANDOFF-ONEPASTE` token, NO internal dates, NO "82%/25KB memory measurement" anecdote. The live SSOT MAY carry the measurement (live tree is not neutrality-bound).
- `make build` required after template edits (embedded via `//go:embed`); commit the rebuilt state per repo convention.
- Commit discipline: one commit per milestone, pathspec-scoped, subjects `feat(SPEC-HANDOFF-ONEPASTE-001): M1 ...` / `M2 ...` (first run-phase commit also carries the `draft → in-progress` frontmatter transition per the ownership matrix).

## §E Self-Verification

Run-phase completion is gated by acceptance.md §D (AC-OP-001..016). The canonical verification batch (single-turn multi-Bash):

1. All grep ACs (AC-OP-001..005, 007..010, 012..014) — see acceptance.md §D.3 for exact commands
2. `moai spec lint .moai/specs/SPEC-HANDOFF-ONEPASTE-001/spec.md` → "No findings"
3. `git diff --name-only <base>..HEAD -- '*.go' | wc -l` → 0 (AC-OP-011)
4. `go build ./...` after `make build` (template embed integrity — build only; no Go source change)
5. `git show --stat` per milestone commit → file lists match §F change list exactly (scope audit)

## §F Milestones

### M1 — Doctrine SSOT + render surface + cross-refs (live tree only)

| # | File | Change |
|---|------|--------|
| 1 | `.claude/rules/moai/workflow/session-handoff.md` | (a) NEW [HARD] emission-obligation clause (REQ-OP-001/002) placed adjacent to § When To Generate; (b) NEW § Auto-Injected Resume Flow (mode=auto) section (REQ-OP-003/004/012 — one-message flow, goal-first variant, ultrathink guidance + A2 caveat, Kickoff-gate invariant, fail-open/manual-reversion note per REQ-OP-013); (c) § Auto-Memory Integration pruning revision (REQ-OP-006); (d) § Post-Paste /goal Follow-up Block re-labeled as the mode=manual fallback path (retained verbatim otherwise, REQ-OP-005); (e) Paste-Time Activation Matrix accuracy check (REQ-OP-005 — additive note only if needed); (f) SSOT-side drift sentinel updated (REQ-OP-007) |
| 2 | `.claude/output-styles/moai/moai.md` | §8 Session Handoff block: compact emission-obligation + auto-flow pointer at SSOT parity; render-side drift sentinel updated (REQ-OP-007) |
| 3 | `.claude/rules/moai/workflow/goal-directive.md` | § MoAI Integration Notes: new bullet cross-referencing the mode=auto injected-context goal-first path, pointing at session-handoff.md as SSOT (REQ-OP-008) |

Commit: `feat(SPEC-HANDOFF-ONEPASTE-001): M1 doctrine SSOT + render surface + goal-directive cross-ref` — pathspec: the 3 files above + spec frontmatter transition + progress.md §E.2.

### M2 — Config flip + template mirrors + build

| # | File | Change |
|---|------|--------|
| 1 | `.moai/config/sections/handoff.yaml` | `mode: manual` → `mode: auto` (REQ-OP-009; comment updated to note local dev opt-in) |
| 2 | `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` | Neutral mirror of M1 normative rules (REQ-OP-011) |
| 3 | `internal/template/templates/.claude/output-styles/moai/moai.md` | Neutral mirror of M1 §8 parity block (REQ-OP-011) |
| 4 | `internal/template/templates/.claude/rules/moai/workflow/goal-directive.md` | Neutral mirror of M1 cross-ref bullet (REQ-OP-011) |
| 5 | `internal/template/templates/.moai/config/sections/handoff.yaml` | **NO CHANGE** — verified by AC-OP-004 (REQ-OP-010) |
| — | `make build` | regenerate embedded templates; `go build ./...` green |

Commit: `feat(SPEC-HANDOFF-ONEPASTE-001): M2 config flip (local auto) + neutral template mirrors` — pathspec-scoped.

## §G Anti-Patterns (do NOT)

- `git add -A` on this shared checkout (unrelated in-flight files would leak into SPEC commits).
- Editing inside the cut-line/6-block/localization specification blocks (REQ-OP-014 violation; also trips the parity sentinel obligations beyond scope).
- Copying the live SSOT verbatim into the template mirror (leaks SPEC IDs/dates/anecdotes — REQ-OP-011 violation; CI guard `template-neutrality-check.yaml` fires on path change).
- Deleting or weakening the two-step `/goal` follow-up mechanism (it is the manual fallback, not dead code — REQ-OP-005).
- Claiming the goal-first flow restores xhigh effort (A2 caveat: effort keywords inside slash args not documented to fire).
- Touching any `.go` file, including "trivial comment sync" (REQ-OP-015 is absolute).
- Flipping the template handoff.yaml to auto "for consistency" (REQ-OP-010 explicitly forbids).

## §H Risk Table

| Risk | Severity | Mitigation |
|------|----------|------------|
| Injected-body staleness — saved preconditions no longer hold when injected after a long gap | Medium | Block 4 preconditions are self-verifying commands the resumed session runs first (unchanged discipline); the injector's auto-mode stale TTL (REQ-AUTORESUME-019) removes old records before injection; doctrine keeps precondition verification mandatory in the auto flow |
| Parallel-session `pending.json` race — two sessions `/clear` concurrently | Low | **Already handled in Go**: claim-then-inject atomic rename (`handoff_inject.go` `claimAndInject`, REQ-AUTORESUME-012/013 — rename success gates injection; any rename errno skips fail-open). Doctrine cites this; no new mechanism |
| Render-surface drift (moai.md §8 vs SSOT) | Medium | Both drift sentinels updated in the same M1 commit; AC-OP-002 parity grep on both surfaces |
| Template neutrality leak (SPEC ID / date / anecdote in mirror) | Medium | AC-OP-005 zero-match grep for the SPEC token in `internal/template/templates/`; CI guard `template-neutrality-check.yaml`; §G anti-pattern named |
| `additionalContext` 10,000-char cap overflow on oversized bodies | Low | A1: Diet Constraints bound the 6-block body far below the cap; cap noted in the flow section; render layer is closed verified Go (out of scope) |
| Memory pruning loses recall signal | Low | Pruning is SHOULD + one-line summary retained + `consumed/` audit path named in the memory entry; MEMORY.md index unaffected; forward-looking only |
| `moai` CLI absent in an environment that adopted the doctrine | Low | REQ-OP-002 fail-open — paste-ready surface unchanged; manual path fully functional |
| Doctrine describes injected guidance inaccurately | Medium | A3 verify-before-author: pre-flight step 4 reads `handoff_inject_render.go` before the flow section is written |

## §I Open Questions (plan-auditor attention)

1. **AC-OP-013 mechanism**: structural-unchanged verification is proposed as a diff-hunk zero-match grep (`git diff -U0 <base>.. -- session-handoff.md | grep -E '^[-+].*✂'` etc.). Confirm this is acceptable as semi-mechanical, or require a stricter section-hash comparison.
2. **moai.md §8 depth**: proposed as a compact pointer + emission clause (render-surface pattern), NOT a full duplicate of the auto-flow section. Confirm parity means "clause present + pointer", not full duplication.
3. **REQ-OP-006 scope**: pruning directive is forward-looking only. Confirm no retroactive memory-file rewrite is expected in this SPEC's run/sync phases.
