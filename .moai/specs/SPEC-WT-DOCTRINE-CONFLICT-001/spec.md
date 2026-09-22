---
id: SPEC-WT-DOCTRINE-CONFLICT-001
title: "Resolve the three-line worktree doctrine contradiction (creation recipe, unconditional git -C deprecation, Claude-exclusive runtime-tool sentence)"
version: "0.1.0"
status: in-progress
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "worktree, doctrine-conflict, documentation-only, cross-harness, template-parity, git-c-scoping"
tier: S
related_specs: [SPEC-WT-DOC-001, SPEC-GITFLOW-DOCTRINE-ALIGN-001] # SPEC-WT-DOC-001 is archived — reference retained for lineage only; this SPEC has no active dependency on it (see Cross-references)
---

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-09-22 | manager-spec | Initial creation — plan-phase artifacts for card t1072 (Class C, Tier S, docs-only). Scope fixed by the lane dispatch: judge and land three independently-diverging doctrine lines as three separately-scoped edits. Defect inventory measured on this tree (base 7f86971fc); see plan.md §B. |
| 2026-09-22 | manager-spec | Plan-audit iter-1 delta fixes (card t1072): D1 archived-SPEC provenance annotation (SPEC-WT-DOC-001); D2 census reworded to exact form `git worktree add -b feat/SPEC` (2 occurrences) + same-class residual inventory extended to REQ-WDC-007(d)/plan.md §B; D3 baseline census corrected to 3 diff hunk headers across 2 logical regions (~610-612, ~653-665); D5 REQ-WDC-004 citation range `:44-48` → `:44-54`. |
| 2026-09-22 | manager-spec | D6 residual inventory extension per plan-audit iter2 — lead-approved, AC semantics unchanged: REQ-WDC-007(d), plan.md §B, and AC-WDC-009's surface list extended with the same-CLASS docs-site occurrences (`worktree/faq.md` ×4 locales + `cli-reference/worktree.md` ×4 locales, 8 lines measured by grep — the audit prose's "10" is contradicted by its own 8-line evidence listing and by direct measurement). RECORDED, not fixed — fix scope unchanged. |

## §A Context and Problem

The worktree doctrine carries a three-line internal contradiction. Each line diverges in a different direction, so each is judged and repaired separately — the card forbids a lumped fix.

### §A.1 Item (1) — stale creation recipe (session-handoff-examples.md:180)

`.claude/rules/moai/workflow/session-handoff-examples.md` line 180 currently reads (local copy; template copy `internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md` is byte-identical):

> `-w <name>` takes the **worktree name**, not a branch name and not a SPEC ID; it resolves to `.claude/worktrees/<name>/`. An existing worktree of that name is **reused, not recreated**, which is what makes this a valid re-entry path. Naming the worktree after the SPEC ID at creation time (`git worktree add -b feat/SPEC-X-001 .claude/worktrees/SPEC-X-001 origin/main`) lets the resume line read `moai cc -w SPEC-X-001`.

The parenthetical presents bare `git worktree add -b ...` as the creation recipe. This collides with the root `AGENTS.md` worktrees contract ("never create one with a bare `git worktree add`") and with the same tree's `worktree-integration.md:54` [HARD]: "never use bare `git worktree add`".

**Ordering probe (decisive, measured on base 7f86971fc):** the recipe landed in `37abf402f` (2026-07-27, #1161); the `AGENTS.md` prohibition landed in `fd3ac06a8` (2026-08-22, t82). At `fd3ac06a8^` the recipe already existed (then at line 173). The prohibition is therefore the LATER rule; line 180's recipe is a stale leftover the later rule overturned without updating. **Fix direction: update line 180 to the sanctioned creation verb; the naming advice itself (worktree named after the SPEC ID so the resume line reads `moai cc -w SPEC-X-001`) remains valid and is preserved verbatim.**

**Census (measured):** the recipe's exact form `git worktree add -b feat/SPEC` occurs in exactly 2 places — the local and template copies of this one file, `session-handoff-examples.md:180` each (grep census). The broader same-CLASS bare-`git worktree add` recipe has additional occurrences OUTSIDE this card's fix scope — recorded in REQ-WDC-007's residual inventory (plan.md §B).

### §A.2 Item (2) — unconditional `git -C` deprecation (worktree-integration.md:227)

`.claude/rules/moai/workflow/worktree-integration.md` line 227 currently reads:

> The shell-`cd` form (`cd <path> && <launcher>`), the `git -C <path>` form, and the subshell-`cd` form (`(cd <path> && ...)`) are DEPRECATED for orchestrator-emitted current-session worktree entry guidance. They break `Agent(isolation: "worktree")` CWD isolation (the agent's CWD is the worktree root; a `cd /absolute/path` bypasses it) and were the root cause of prior incidents where a sub-agent used `git -C` instead of `EnterWorktree` and was corrected mid-run.

`git -C <path>` is marked DEPRECATED unconditionally, but it is the ONLY remaining current-session entry means for Codex sessions: Codex has no `EnterWorktree` runtime tool, its native entry verb is the new-session launcher `moai codex -w <worktree>` (usage string verified at `internal/cli/codex_launcher.go:47`), and the root `AGENTS.md` worktrees contract itself MANDATES `git -C <path>` as the way to drive a worktree. This is the highest-priority item of the three: marking an existing, operative path as discouraged is more quietly harmful than documenting a capability gap.

### §A.3 Item (3) — Claude-exclusive runtime-tool sentence (worktree-integration.md:220)

`worktree-integration.md` line 220 currently ends with:

> These are Claude Code runtime tools — MoAI does not mandate their use; they are the interactive counterpart to the launcher `-w` flag and `isolation: worktree` frontmatter.

This is the doctrine's ONLY sentence stating Claude-exclusivity of the runtime tools, and it offers no alternative for other harnesses.

### §A.4 Parity and divergence baseline (measured on base 7f86971fc)

- `session-handoff-examples.md`: local copy == template copy, byte-identical (`diff` rc 0). Any edit MUST be applied identically to both copies.
- `worktree-integration.md`: local copy DIFFERS from the template copy at exactly two regions — hunks at lines ~610-612 and ~653-658 (intentional divergence: cards t1067/t1069 carry internal provenance in the local copy; the template copy is neutralized). These regions are OUT OF SCOPE; the edits must not touch them and must not introduce any NEW divergence — the replacement text is written identically (neutral, provenance-free) into BOTH copies.

## §B Requirements (GEARS)

### REQ-WDC-001 — Sanctioned creation recipe (Event-driven)

**When** a reader applies the worktree-creation recipe at `session-handoff-examples.md:180`, the document shall present the sanctioned L1 creation verb — `moai worktree new SPEC-X-001` (creates the tree, returns its absolute path, never enters it) — in place of bare `git worktree add -b ...`, in BOTH the local and the template copy, and shall preserve the naming-advice sentence verbatim in meaning ("lets the resume line read `moai cc -w SPEC-X-001`").

### REQ-WDC-002 — Harness-scoped `git -C` deprecation (State-driven)

**While** a harness carries a native current-session worktree entry tool (Claude Code: `EnterWorktree`), the entry-guidance deprecation at `worktree-integration.md:227` shall apply to the shell-`cd`, `git -C <path>`, and subshell-`cd` forms with the original hazard rationale preserved verbatim in meaning (`Agent(isolation: "worktree")` CWD isolation bypass; the mid-run correction incidents).

### REQ-WDC-003 — Non-native-harness carve-out (Where)

**Where** a harness has NO native current-session worktree entry tool (Codex), the doctrine shall state that `git -C <path>` remains the documented means of driving a worktree from the current session — per the root `AGENTS.md` worktrees contract — and is not deprecated there. No new verbs are invented; the carve-out references only the already-documented `git -C <path>` form.

### REQ-WDC-004 — Non-Claude alternative pointer (Ubiquitous)

The runtime-tool sentence at `worktree-integration.md:220` shall name the harness-neutral alternative: the launcher `-w` forms for new-session entry (`moai cc -w`, `moai glm -w`, `moai codex -w` — already documented at `worktree-integration.md:44-54`) and, for current-session entry where no runtime tool exists, the `git -C <path>` fallback as scoped by REQ-WDC-003.

### REQ-WDC-005 — Per-item scoped landing (Unwanted)

The system shall land each item as its own scoped edit; the three edits shall not be merged into one rewrite of either file. Commit granularity: a single commit carrying three clearly separable hunks is the planned form (decision and rationale in plan.md §F).

### REQ-WDC-006 — Boundary fence (Unwanted)

The change shall NOT modify: any Go code; `AGENTS.md` or `internal/template/templates/AGENTS.md.tmpl`; any region of `worktree-integration.md` outside the two cited sentences' immediate context (a sibling card edits the same file near line 52 — hunks stay line-scoped to ~218-232); and shall introduce NO new local-vs-template divergence in the edited regions. The template-copy text of every edited region shall stay neutrality-clean (no card ids, no internal dates, no macOS-bias paths, no SPEC IDs of this card).

### REQ-WDC-007 — Residual record (Ubiquitous)

The SPEC shall record what REMAINS unfixed after the three edits (plan.md §B): (a) `AGENTS.md` / `AGENTS.md.tmpl` stay untouched — the ability-binding table row and the `moai codex` verbs-table listing are card t1071's scope; (b) a native current-session entry tool for non-Claude harnesses remains absent (factory-lane worktree handoff is a separate queued card); (c) the intentional local-vs-template divergence at `worktree-integration.md` ~610/~653 is left as-is; (d) after the three in-scope edits, the same-CLASS bare-`git worktree add` creation recipe REMAINS unfixed at: the 4 docs-site worktree guides `docs-site/content/{en,ja,ko,zh}/worktree/guide.md` (`en:119`, `ja:114`, `ko:140`, `zh:111` — user-facing docs, separate sync surface), the same-CLASS occurrences at `docs-site/content/{en,ja,ko,zh}/worktree/faq.md` (`en:671`, `ja:662`, `ko:653`, `zh:660` — recovery recipe) and `docs-site/content/{en,ja,ko,zh}/cli-reference/worktree.md` (`en:18`, `ja:18`, `ko:20`, `zh:18` — `moai cc -w` alternative "or `git worktree add`"; 8 additional lines, measured by grep), template `main-checkout-branch-guard.md:38` (`git worktree add -b <branch> <worktree-path> origin/main`), and dev-only `hns-release-specialist.md:122` (`git worktree add -b release/vX.Y.Z <worktree-path> origin/develop`; same file also carries a bare `git worktree add --detach` at :233) — card scope (the three cited lines) is UNCHANGED; these are residual inventory, not fix scope.

## §C Verifiable Scope Summary

| # | Target | Edit | Both copies |
|---|--------|------|-------------|
| 1 | `session-handoff-examples.md:180` (local + template) | Replace the parenthetical creation recipe with `moai worktree new SPEC-X-001`; preserve naming advice | identical text |
| 2 | `worktree-integration.md:227` (local + template) | Rescope the DEPRECATED sentence to native-tool-carrying harnesses; add the non-native carve-out | identical text |
| 3 | `worktree-integration.md:220` (local + template) | Append the non-Claude alternative pointer sentence | identical text |

## Out of Scope

### Out of Scope — Files owned by other cards or layers

- `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl` — card t1071's scope (ability-binding table row, `moai codex` verbs-table listing).
- Any Go source under `internal/`, `pkg/`, `cmd/` — docs-only card, zero code changes.
- `worktree-integration.md` regions outside ~218-232, in particular line ~52 (sibling card edits there) and the intentionally divergent hunks at ~610-612 / ~653-658 (cards t1067/t1069 provenance — left as-is).

### Out of Scope — Capability work beyond doctrine repair

- A native current-session worktree entry tool for non-Claude harnesses (factory-lane worktree handoff is a separate queued card).
- Any change to the launcher `-w` flag semantics, the L1/L2 lifecycle, or `moai worktree new` behavior.
- Rewriting the t1067/t1069 intentional divergence into a neutral shared form.

## Cross-references

- `.claude/rules/moai/workflow/worktree-integration.md` — items (2) and (3); sanctioned verbs at :44-54.
- `.claude/rules/moai/workflow/session-handoff-examples.md` — item (1).
- Root `AGENTS.md` worktrees contract — the later rule that item (1) trails, and the `git -C` driving mandate item (2) carves out for.
- SPEC-WT-DOC-001 (archived — reference retained for lineage only; this SPEC has no active dependency on it) — the worktree shared-state doctrine this SPEC edits adjacent to.
- SPEC-GITFLOW-DOCTRINE-ALIGN-001 — precedent card for a three-item docs-doctrine alignment with the same per-item judgment discipline.
