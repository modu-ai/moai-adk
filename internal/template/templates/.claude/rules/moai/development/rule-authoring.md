---
description: "Always-loaded surface recurrence control — statement duty on creating a new slot file and on any single edit growing one by more than 1,000 bytes"
paths: "**/.claude/rules/**,**/CLAUDE.md,**/.claude/output-styles/**,**/MEMORY.md"
---

# Rule Authoring — Always-Loaded Cost Discipline

> Loading scope: this rule keys the four guard slots themselves. Any edit that creates or grows an always-loaded file is the moment this discipline attaches — including edits to this repository's `CLAUDE.md`, its output styles, and `MEMORY.md`.

## The surface and its four slots

Claude Code re-injects the always-loaded surface on every turn, and `/clear` re-pays all of it at once. The surface is made of four slots:

1. Rule files under `.claude/rules/` that lack a top-level `paths:` frontmatter key
2. `CLAUDE.md`
3. Output styles under `.claude/output-styles/`
4. `MEMORY.md` (head)

Rule files are not the whole surface. `CLAUDE.md` and the output styles together can outweigh every rule file, and the largest single contributor is an output style, not a rule. A duty that binds rule files only leaves the biggest share of the surface unwatched — every clause below therefore binds all four slots.

## Why a cleanup is not enough

A one-time diet does not stay done. In the repository this rule set ships from, the always-loaded rule surface grew roughly 4x in the three months after the previous reduction — and about half of that growth was existing files expanding in place, not new files arriving. Blocking new files alone blocks half the recurrence; the duty below covers both modes.

## The statement duty

[ZONE:Evolvable] [HARD] The duty has four parts, binding all four slots:

- **(a) New file.** Creating a new always-loaded file — a rule file without top-level `paths:`, or a slot file such as `CLAUDE.md` or an output style — requires stating the file's byte size and a cost justification in the change description (commit body, or PR description where the change rides a PR).
- **(b) Growth.** A single edit that grows an existing always-loaded file by more than 1,000 bytes requires the same statement, sized to the growth rather than the whole file.
- **(c) Non-invoking cost.** Whatever statement (a) or (b) requires must address the cost paid by sessions that never need the file. Every session loads the surface; sessions that never invoke this feature pay for it anyway — every turn, and again after every `/clear`. No current always-loaded file's justification covers this cost: `session-handoff.md` is the closest precedent (it names the re-paid prefix as a cost), and `native-idiom-and-register.md`'s "zero burden for English sessions" is true for behavior and false for context bytes. A justification that does not name what the never-needing session pays is incomplete.
- **(d) Scope first.** Before either duty fires, ask whether the content can live under a `paths:` scope instead — a reference companion keyed to the files that need it costs the surface nothing. Only content that genuinely must be always-loaded proceeds to the statement.

## Threshold calibration

The threshold is **1,000 bytes, single-edit basis**. Typical edit shapes:

| Edit shape | Typical size | Duty fires |
|---|---|---|
| Typo fix | ~100 B | no |
| One-line addition | ~200 B | no |
| A paragraph | ~800 B | no |
| A new HARD clause | ~1,200 B | yes |
| A new `##` section | ~2,500 B | yes |

There is deliberately **no cumulative-delta secondary trigger**: growth arriving as many sub-1,000-byte edits accumulates unwatched. That blind spot is an accepted residual risk, not an oversight to patch later — tracking per-file drift on every small edit would cost more attention than the growth it catches.

## Self-compliance

This rule carries a top-level `paths:` key, so it is excluded from the always-loaded surface it polices. The recurrence control must not itself be recurrence.

---

Version: 1.0.0
Classification: Evolvable operational rule — binds edits to the four always-loaded slots; changes no gate semantics.
