---
id: SPEC-WT-DOCTRINE-CONFLICT-001
acceptance_version: "0.1.0"
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
---

# Acceptance — SPEC-WT-DOCTRINE-CONFLICT-001

## §A Verification Model

Docs-only SPEC: `go test` proves nothing about these edits and is explicitly NOT a gate here. The regression guard is the grep-based parity census (plan.md §E V1-V3 / P1-P3 / B1-B2). Every AC below is binary-testable by a runnable command with a stated expected output.

## §D AC Matrix

### AC-WDC-001 — Item (1): sanctioned creation recipe replaces bare `git worktree add` (blocks)

- **Given** `session-handoff-examples.md` (local copy AND template copy `internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md`) at line 180,
- **When** the run-phase edit lands,
- **Then** `grep -rn "git worktree add -b feat" <local> <template>` returns no output (exit 1), `grep -c "moai worktree new SPEC-X-001" <local>` and `<template>` each return `1`, and `grep -c 'lets the resume line read `moai cc -w SPEC-X-001`'` each return `1` (naming advice preserved).

### AC-WDC-002 — Item (2): `git -C` deprecation rescoped to native-tool harnesses (blocks)

- **Given** `worktree-integration.md:227` (local AND template) carrying the unconditional DEPRECATED sentence,
- **When** the run-phase edit lands,
- **Then** `grep -c "on harnesses that carry a native current-session entry tool" <local>` and `<template>` each return `1`; `grep -c "is not deprecated there" <local>` and `<template>` each return `1` (the Codex carve-out); and the hazard rationale survives: `grep -c 'CWD isolation' <local>` still matches the `Agent(isolation: "worktree")` rationale sentence.

### AC-WDC-003 — Item (3): non-Claude alternative pointer added (blocks)

- **Given** `worktree-integration.md:220` (local AND template) ending with the Claude-exclusivity sentence,
- **When** the run-phase edit lands,
- **Then** `grep -c "On a harness without these runtime tools" <local>` and `<template>` each return `1`, and the appended sentence references only already-documented verbs — `grep -oE "moai (cc|glm|codex) -w|moai worktree new" ` over the appended region yields matches from that closed set only.

### AC-WDC-004 — Per-item scoped landing (blocks)

- **Given** the three items are judged separately by the card,
- **When** the edits land,
- **Then** `git diff` over the two workflow files shows exactly three content changes (one per item, applied to both copies = 6 line-regions), each matching exactly one AC-WDC-001/002/003 check; no hunk rewrites an untargeted sentence.

### AC-WDC-005 — Copy parity (blocks)

- **Given** `session-handoff-examples.md` was byte-identical (local==template) before the edit,
- **When** the edit lands,
- **Then** `diff <local> <template>` is empty (rc 0) — full-file parity restored identically; and for `worktree-integration.md`, `diff <(sed -n '215,235p' <local>) <(sed -n '215,235p' <template>)` is empty (rc 0) — the edited-region text is byte-identical across copies.

### AC-WDC-006 — No new divergence (blocks)

- **Given** the pre-edit baseline census (plan.md §C.2) counts exactly 3 diff hunk headers across 2 logical regions (~610-612 and ~653-665 — intentional, cards t1067/t1069) (`worktree-integration.md` local vs template),
- **When** the edit lands,
- **Then** the same census command still returns exactly `3` hunk headers across the same 2 logical regions (the intentional t1067/t1069 divergence, untouched) — NO NEW hunk headers may appear in the edited regions (~218-232).

### AC-WDC-007 — Boundary fence (blocks)

- **Given** the card forbids code changes, `AGENTS.md` edits, and out-of-region edits,
- **When** the run-phase completes,
- **Then** `git status --porcelain -- AGENTS.md internal/template/templates/AGENTS.md.tmpl` is empty; `git status --porcelain -- '*.go'` is empty; and `git status --porcelain` lists only the two workflow files (both copies) plus `.moai/specs/SPEC-WT-DOCTRINE-CONFLICT-001/`.

### AC-WDC-008 — Template neutrality (blocks)

- **Given** template copies must stay free of internal development state (CLAUDE.local.md §2.1 forbidden classes),
- **When** the template-copy text is inspected,
- **Then** the edited regions contain no card ids (`grep -E "t10[0-9]{2}"` over edited template regions → no match), no internal dates (`grep -E "2026-09" ` → no match in edited regions), no SPEC IDs of this card, and no macOS-bias absolute paths.

### AC-WDC-009 — Residual record present (blocks)

- **Given** card [HARD] requires naming what remains unfixed,
- **When** the SPEC artifacts are read,
- **Then** plan.md §B carries the four-item residual inventory (AGENTS.md/t1071 scope; absent native entry tool for non-Claude harnesses; intentional ~610/~653 divergence; same-class bare-`git worktree add` recipes remaining at the 4 docs-site worktree guides, template `main-checkout-branch-guard.md:38`, and dev-only `hns-release-specialist.md:122`) and spec.md REQ-WDC-007 binds it.

## §D.1 Severity

| AC | Severity | Rationale |
|----|----------|-----------|
| 001-003 | blocks | the three fixes ARE the card |
| 004-006 | blocks | per-item separation + copy parity are the card's core discipline |
| 007-009 | blocks | boundary violations make the fix a regression source |

## §D.5 Traceability

- AC-WDC-001 (maps REQ-WDC-001) — item (1) sanctioned recipe + preserved naming advice.
- AC-WDC-002 (maps REQ-WDC-002, REQ-WDC-003) — item (2) harness-scoped deprecation + non-native carve-out.
- AC-WDC-003 (maps REQ-WDC-004) — item (3) non-Claude alternative pointer.
- AC-WDC-004 (maps REQ-WDC-005) — per-item scoped landing, no lumped rewrite.
- AC-WDC-005, AC-WDC-006, AC-WDC-007, AC-WDC-008 (maps REQ-WDC-006) — copy parity, no-new-divergence, boundary fence, template neutrality.
- AC-WDC-009 (maps REQ-WDC-007) — residual record present.

## §D.6 Indirect Verification

None required — every AC is directly observable by grep/diff/git-status on the changed files. No behavior-level claims are made by a docs-only change.

## §D.7 Closure Gates

1. All ACs observed PASS in one batched read-only verification turn; verbatim outputs recorded in progress.md §E.2.
2. SPEC artifacts committed on `WT-doctrine-conflict` (branch re-read immediately before the commit; explicit pathspec staging).
3. Spec lint result recorded in progress.md §E.1.
