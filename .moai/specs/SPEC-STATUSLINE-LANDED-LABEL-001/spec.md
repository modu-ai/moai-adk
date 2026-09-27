---
id: SPEC-STATUSLINE-LANDED-LABEL-001
title: "Statusline landed annotation — subject-attribution criterion, verify-before-done glyph, criterion-versioned cache"
version: "0.1.0"
status: draft
created: 2026-09-27
updated: 2026-09-27
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/statusline"
lifecycle: spec-anchored
tags: "statusline, kanban, todo, landed-annotation, subject-attribution, glyph, cache-schema"
tier: S
---

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-09-27 | manager-spec | Initial creation — plan-phase artifacts for card t1281 (Class C, Tier S). Decision ③ fixed by the lead (reuse kanban's subject-attribution matcher + generation boundary; replace the check mark with a verify-before-done glyph; keep "never subtracts"; version the cache criterion). Problem measurements supplied by the lane; code shape measured on this tree (base b59a5d69c). |

## §A Context and Problem

The statusline TODO segment renders `🔄 TODO: picked/queued ✓N` (`internal/statusline/renderer.go`, backlog segment). `N` comes from `internal/statusline/landed.go`: a detached refresh child runs ONE `git log <ref> --format=%B` and counts every picked card id that appears ANYWHERE in ANY commit message body under `\b<id>\b`. The result is cached at `.moai/state/landed/counts.json` (TTL 10 minutes, stale-while-revalidate, timestamp-first stampede guard).

Two defects, one visible reading:

1. **The criterion over-counts.** A body mention counts, so another card's report commit that merely names an id makes that id "landed". Lane measurement this session: cache `landed=9`, while `moai todo auto-done --dry-run` (subject attribution, ref `origin/develop`) reported `scanned=29 closed=4`.
2. **Even the subject-attributed number is not "done".** The four subject-attributed cards had landed only their plan phase. Auto-done closing plan-only landings is a known trap.
3. **The glyph invites the misread.** Operators read `✓` as "safe to close". The annotation must read "a landing commit exists — verify before done", never "done / closable".

The kanban package already owns the correct, tested machinery: `kanban.LandedScanArgs` / `kanban.ScanLandedSubjects` (one `%H%x00%ct%x00%s` subject stream), `kanban.LandedAttributions` (the single `subjectAttribution` predicate, including the non-landing declaration and the absorb-direction merge rule), and `kanban.AutoDoneSubjectFresh` (the generation boundary against the card's `added_at`). This SPEC aligns the statusline with that machinery rather than adding a second matcher.

### §A.1 Glyph selection

The glyph must be locale-neutral (identical across ko / en / ja / zh, no words), single display width, and text-presentation (not an emoji that terminals may widen).

| Candidate | Code point | East Asian Width | Rationale |
|---|---|---|---|
| **⚑ (chosen)** | U+2691 BLACK FLAG | N (single width) | A flag reads "flagged for attention"; it carries no completion connotation, has no Emoji property (renders as text, one cell), and is visually distinct from the other segment icons. |
| ⇡ | U+21E1 UPWARDS DASHED ARROW | N (single width) | Reads "reached upstream"; the dashed stroke hints "unconfirmed". Rejected: arrows are easily read as a trend (count going up) rather than a state. |
| ⚐ | U+2690 WHITE FLAG | N (single width) | Same semantics as ⚑ with lower visual weight. Rejected: the hollow outline is hard to see at small font sizes and can read as "surrender / nothing". |

Excluded by construction: any check mark (✓ ✔ ☑ — reads "done"), ⚠ U+26A0 and ⤴ U+2934 (carry the Emoji property; width varies by terminal), ◆ U+25C6 and ⊙ U+2299 (East Asian Width "A" — ambiguous, double width in CJK-configured terminals).

## §B Requirements (GEARS)

### REQ-SLL-001 — Subject-attribution criterion (Ubiquitous)

The landed refresh shall count a picked card as landed only when the landed ref's subject stream attributes it through kanban's existing subject-attribution predicate — the same predicate `moai todo auto-done` evaluates — and shall introduce no second card-matching criterion.

### REQ-SLL-002 — Body mentions do not count (Ubiquitous, prohibition)

The landed refresh shall not count a card whose id appears only in a commit message body, or in a subject that the subject-attribution predicate does not attribute to that card.

### REQ-SLL-003 — Generation boundary (Event-driven)

**When** the commit that attributes a picked card has a committer time earlier than that card's `added_at`, or the card's `added_at` cannot be parsed, the landed refresh shall not count that card.

### REQ-SLL-004 — One query per refresh (Ubiquitous)

The landed refresh shall issue exactly one git invocation per refresh regardless of the number of picked cards, using the argument shape kanban's landed scan uses, and the render path shall issue none.

### REQ-SLL-005 — Unknown is never rendered as zero (Event-driven)

**When** the subject query fails or returns a malformed stream, the landed refresh shall keep the previously stored measurement unchanged apart from its timestamp and shall not write a zero; **when** no picked card exists, the refresh shall record an observed zero without querying git.

### REQ-SLL-006 — Criterion-versioned cache (Ubiquitous)

The landed cache shall record an identifier of the counting criterion that produced its number.

### REQ-SLL-007 — Old-criterion caches are unknown (Event-driven)

**When** the reader encounters a cache whose criterion identifier is absent or differs from the current criterion, the reader shall treat the judgment as unknown (render no annotation) and as stale (eligible for an immediate refresh regardless of its timestamp).

### REQ-SLL-008 — Old-criterion numbers never survive a failed refresh (Event-driven)

**When** a refresh fails and the prior cache was written under a different or absent criterion, the landed refresh shall not carry the prior number forward as a measurement under the current criterion.

### REQ-SLL-009 — Verify-before-done glyph (Ubiquitous)

The TODO segment shall render a known landed count as a space, the glyph `⚑` (U+2691), and the count, and shall not render a check mark (`✓`) anywhere in the segment.

### REQ-SLL-010 — Locale-neutral, single-width glyph (Ubiquitous)

The landed glyph shall be the same single code point in every conversation locale, carry no words, and occupy exactly one terminal display cell.

### REQ-SLL-011 — Landed never subtracts from picked (Ubiquitous)

The TODO segment shall render the picked and queued numbers unchanged by the landed count, and the renderer's rationale comment shall state that a card stays picked until auto-done or `moai todo done` actually closes it.

### REQ-SLL-012 — User documentation matches the glyph (Where)

**Where** user-facing documentation describes the landed annotation, it shall describe the `⚑N` glyph and its verify-before-done meaning in every published locale, and shall no longer present `✓N`.

## §C Constraints

- Reuse only: `kanban.LandedScanArgs` / `kanban.ScanLandedSubjects`, `kanban.LandedAttributions`, `kanban.LandedBranchFromRef`, `kanban.AutoDoneSubjectFresh`. The kanban package is not modified unless an adapter need is proven during run.
- The render path stays constant-cost: one small file read, no subprocess.
- The detached-child, TTL, stampede-guard, and fork-bomb-guard shape of `landed.go` is preserved.
- Lint gate: golangci-lint v2.1.6 (CI version).

## §D Exclusions

### Out of Scope — close-policy guards

- Guard M1 (reissued-id collision, `AutoDoneDistinctTexts`) and guard M2 (SPEC-status sync gate) are not applied to the statusline count. The count answers "a landing commit exists", not "this card may close"; a plan-only landing still counts, which is exactly why the glyph reads verify-before-done.
- No change to `moai todo auto-done`, `moai todo pr`, or the subject-attribution predicate itself.

### Out of Scope — layout and other segments

- No change to the picked/queued numbers, the `🔄 TODO:` label, segment ordering, or the `backlog` / `workflow.todo.enabled` switches.
- No change to the TTL value or refresh budget.
- CHANGELOG.md historical entries mentioning `✓N` are not rewritten (the sync phase adds a new entry).

### Out of Scope — per-card breakdown

- The statusline does not list which cards are counted; `moai todo auto-done --dry-run` remains the per-card surface.
