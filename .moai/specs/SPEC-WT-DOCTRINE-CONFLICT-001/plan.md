---
id: SPEC-WT-DOCTRINE-CONFLICT-001
plan_version: "0.1.0"
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
---

# Implementation Plan — SPEC-WT-DOCTRINE-CONFLICT-001

## §A Context

Docs-only card t1072: three doctrine lines diverge in three different directions, each with its own measured evidence and fix direction (spec.md §A). Zero code changes; Tier S. The fix target is TWO files, each edited in BOTH copies (local `.claude/rules/moai/workflow/` + template `internal/template/templates/.claude/rules/moai/workflow/`), with byte-identical neutral replacement text.

## §B Known Issues and Residual Inventory (measured on base 7f86971fc)

Fixed by this SPEC (three, judged separately):

| # | Location | Defect | Fix direction | Priority |
|---|----------|--------|---------------|----------|
| (1) | `session-handoff-examples.md:180` local+template | Parenthetical presents bare `git worktree add -b ...` as creation recipe; collides with root `AGENTS.md` and `worktree-integration.md:54` [HARD]. Ordering probe: recipe `37abf402f` (2026-07-27) predates prohibition `fd3ac06a8` (2026-08-22) — prohibition is the later rule | Rewrite parenthetical to `moai worktree new SPEC-X-001`; preserve naming advice | 2 |
| (2) | `worktree-integration.md:227` local+template | `git -C <path>` DEPRECATED unconditionally, yet it is the only current-session entry means for Codex (no `EnterWorktree` tool; new-session verb `moai codex -w` verified at `internal/cli/codex_launcher.go:47`; `AGENTS.md` itself mandates `git -C` for driving) | Rescope deprecation to harnesses WITH a native entry tool; add non-native carve-out; preserve hazard rationale | **1 (highest)** |
| (3) | `worktree-integration.md:220` local+template | "These are Claude Code runtime tools" is the doctrine's only exclusivity sentence, no alternative offered | Append harness-neutral alternative pointer, coherent with the (2) rescoping | 3 |

Remains UNFIXED after this SPEC (residual record, card [HARD]):

- `AGENTS.md` / `internal/template/templates/AGENTS.md.tmpl` stay untouched — the ability-binding table row and the `moai codex` verbs-table listing belong to card t1071.
- A native current-session entry tool for non-Claude harnesses remains absent; factory-lane worktree handoff is a separate queued card.
- The intentional local-vs-template divergence at `worktree-integration.md` ~610-612 / ~653-658 (cards t1067/t1069 provenance) is left as-is.
- The same-CLASS bare-`git worktree add` creation recipe REMAINS unfixed at (measured by grep on this tree): the 4 docs-site worktree guides `docs-site/content/{en,ja,ko,zh}/worktree/guide.md` (`en:119`, `ja:114`, `ko:140`, `zh:111` — user-facing docs, separate sync surface), template `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md:38` (`git worktree add -b <branch> <worktree-path> origin/main`), and dev-only `.claude/agents/harness/hns-release-specialist.md:122` (`git worktree add -b release/vX.Y.Z <worktree-path> origin/develop`; same file also carries `git worktree add --detach` at :233). Card scope — the three cited lines in §B's fixed table — is UNCHANGED; these are residual inventory, not fix scope.

## §C Pre-flight (Run-phase entry checklist)

1. [ ] Confirm tree and branch: `git -C <worktree> rev-parse --short HEAD` and `git branch --show-current` → `WT-doctrine-conflict`; re-read immediately before each commit.
2. [ ] Confirm baseline hunk census (the invariant the edits must preserve): `diff .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md | grep -c '^[0-9]'` → **3** — 3 diff hunk headers across 2 logical regions (~610-612 and ~653-665), intentional divergence, cards t1067/t1069.
3. [ ] Confirm `session-handoff-examples.md` full-file parity: `diff` of the local vs template copy → empty.
4. [ ] Capture verbatim current text of the three target sentences (spec.md §A quotes) so the run-phase edits are exact-string replacements, not fuzzy rewrites.

## §D Constraints

- [HARD] Zero code changes. Changed paths are limited to: the two workflow files (both copies) + `.moai/specs/SPEC-WT-DOCTRINE-CONFLICT-001/**`.
- [HARD] No edits to `AGENTS.md` or `internal/template/templates/AGENTS.md.tmpl`.
- [HARD] `worktree-integration.md` hunks stay line-scoped to ~218-232; nothing near line 52 (sibling card) or ~610/~653.
- [HARD] Replacement text is byte-identical in local and template copies and neutrality-clean (no card ids, no internal dates, no macOS-bias paths, no SPEC IDs of this card).
- [HARD] No new verbs invented — only verbs already documented at `worktree-integration.md:44-54`: `moai worktree new`, `moai cc -w`, `moai glm -w`, `moai codex -w`, `EnterWorktree`, `ExitWorktree`, plus the `git -C <path>` form.
- Edit order to keep line numbers stable: `worktree-integration.md` **:227 first**, then **:220** (the :220 append shifts the :227 line), then `session-handoff-examples.md` **:180**. After each edit, re-locate the next target by content, not by remembered line number.

## §E Self-Verification (planned run-phase evidence)

`go test` proves nothing about a docs-only change — the regression guard here IS the grep-based parity census below. Each check is runnable and names its expected output:

| Check | Command (from worktree root) | Expected |
|-------|------------------------------|----------|
| V1 bare recipe gone (both copies) | `grep -rn "git worktree add -b feat" .claude/rules/moai/workflow/session-handoff-examples.md internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md` | no output, exit 1 |
| V1b sanctioned verb present (both copies) | `grep -c "moai worktree new SPEC-X-001" <local>` and `<template>` | `1` each |
| V1c naming advice preserved | `grep -c 'lets the resume line read `moai cc -w SPEC-X-001`' <local>` and `<template>` | `1` each |
| V2 scoping present (both copies) | `grep -c "on harnesses that carry a native current-session entry tool" <local>` and `<template>` | `1` each |
| V2b carve-out present (both copies) | `grep -c "is not deprecated there" <local>` and `<template>` | `1` each |
| V3 alternative present (both copies) | `grep -c "On a harness without these runtime tools" <local>` and `<template>` | `1` each |
| P1 session-handoff full parity restored | `diff .claude/rules/moai/workflow/session-handoff-examples.md internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md` | empty, rc 0 |
| P2 worktree-integration edited-region parity | `diff <(sed -n '215,235p' .claude/rules/moai/workflow/worktree-integration.md) <(sed -n '215,235p' internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md)` | empty, rc 0 |
| P3 no NEW divergence | baseline §C.2 census re-run | still exactly **3** hunk headers across the same 2 logical regions (~610-612, ~653-665) |
| B1 AGENTS.md untouched | `git status --porcelain -- AGENTS.md internal/template/templates/AGENTS.md.tmpl` | empty |
| B2 zero code changes | `git status --porcelain` | only the two workflow files (both copies) + the SPEC dir |

Closure gate: V1-V3, P1-P3, B1-B2 all observed PASS in one batched read-only verification run; outputs recorded verbatim in progress.md §E.2.

## §F Milestones (priority-based, no time estimates)

Commit granularity decision: **a single commit carrying three clearly separable hunks** (plus the SPEC artifacts commit). Rationale: the card forbids LUMPING THE JUDGMENT, not sharing the commit — each edit remains a separate scoped Edit call with its own acceptance check (AC-WDC-001..003 verified independently), and the single-commit form keeps the release-branch diff reviewable as one doctrine-repair unit. Each hunk is independently revertable by path+hunk, satisfying the per-item separation requirement of REQ-WDC-005.

### M1 — Wording decisions land first (Priority: Critical; highest change-likelihood decisions)

The three replacement texts are the review-critical surface. They are frozen in spec.md §B and restated here for the reviewer:

- **(1) :180** — replace the parenthetical only:
  - BEFORE: `` Naming the worktree after the SPEC ID at creation time (`git worktree add -b feat/SPEC-X-001 .claude/worktrees/SPEC-X-001 origin/main`) lets the resume line read `moai cc -w SPEC-X-001`. ``
  - AFTER: `` Naming the worktree after the SPEC ID at creation time (`moai worktree new SPEC-X-001` — the sanctioned L1 creation verb; it creates the tree and returns its absolute path without entering it) lets the resume line read `moai cc -w SPEC-X-001`. ``
- **(2) :227** — replace the first sentence's scope and append the carve-out:
  - BEFORE (first sentence): `` The shell-`cd` form (`cd <path> && <launcher>`), the `git -C <path>` form, and the subshell-`cd` form (`(cd <path> && ...)`) are DEPRECATED for orchestrator-emitted current-session worktree entry guidance. ``
  - AFTER (first sentence): same list, with "are DEPRECATED for orchestrator-emitted current-session worktree entry guidance" → "are DEPRECATED for orchestrator-emitted current-session worktree entry guidance **on harnesses that carry a native current-session entry tool** (Claude Code: `EnterWorktree`)".
  - APPEND (new final sentence): `` On a harness without a native current-session entry tool (Codex), `git -C <path>` remains the documented means of driving a worktree from the current session — the root `AGENTS.md` worktrees contract mandates exactly that form — and is not deprecated there. ``
  - The hazard rationale (CWD-isolation break + mid-run correction incidents) is preserved unchanged.
- **(3) :220** — append after "...`isolation: worktree` frontmatter.":
  - `` On a harness without these runtime tools, the harness-neutral equivalents are the launcher `-w` forms for new-session entry (`moai cc -w <name>`, `moai glm -w <name>`, `moai codex -w <name>`; after `moai worktree new <name>` for a fresh L1 tree) and, for current-session entry where no runtime tool exists, `git -C <path>` per the entry-guidance scoping below. ``

**Exit criteria**: M1 texts reviewed against REQ-WDC-001..004 and template neutrality (§D) before any file is touched.

### M2 — Apply edits, both copies (Priority: High)

- [ ] `worktree-integration.md` :227 (local), then the identical Edit on the template copy.
- [ ] `worktree-integration.md` :220 append (local), then template copy.
- [ ] `session-handoff-examples.md` :180 (local), then template copy.
- [ ] Each edit is a separate, content-anchored Edit call (no fuzzy rewrite, no reflow of neighboring lines).

**Exit criteria**: all six edits landed; `git diff --stat` shows exactly the six regions (2 files x 2 lines each in worktree-integration; 1 line in session-handoff-examples, both copies) plus the SPEC dir.

### M3 — Verification batch + commit (Priority: High)

- [ ] Run §E checks V1-V3, P1-P3, B1-B2 in ONE batched read-only turn; record verbatim outputs.
- [ ] Re-read `git branch --show-current` (must be `WT-doctrine-conflict`) immediately before committing.
- [ ] Stage by explicit pathspec (never `git add -A`): the two workflow files (both copies), then the SPEC dir. Conventional Commit, message carries card id `t1072`, ends with attribution `🗿 MoAI`.

**Exit criteria**: commit landed on `WT-doctrine-conflict`; §E closure gate observed.

## §G Anti-Patterns (explicitly forbidden)

- Rewriting the whole `worktree-integration.md` entry-guidance section instead of the two line-scoped sentences.
- "Fixing" the ~610/~653 divergence while in the file (drive-by) — that divergence is intentional (t1067/t1069).
- Touching `AGENTS.md` to "keep the contract consistent" — t1071 owns it.
- Introducing card ids, dates, or provenance into the template copies (template content neutrality, CLAUDE.local.md §2.1).
- Inventing a new verb (`moai worktree enter` does not exist) where the doctrine already documents the real fallback (`git -C <path>`).
- Declaring the fix done because `go test` passes — go test measures nothing here; the §E grep census is the guard.

## §H Cross-references

- spec.md §A (measured premise + ordering probes) and §B (frozen wording).
- acceptance.md §D AC matrix (per-item Given-When-Then + parity/boundary ACs).
- Root `AGENTS.md` worktrees contract; `worktree-integration.md:44-54` (sanctioned verbs); `internal/cli/codex_launcher.go:47` (Codex launcher usage string).
- Card t1071 (AGENTS.md scope), cards t1067/t1069 (intentional divergence), factory-lane handoff card (queued).
