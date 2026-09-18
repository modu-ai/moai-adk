---
id: SPEC-OUTPUT-STYLE-SWITCH-PATH-001
title: "Output-style switch guidance names the /output-style slash command alongside /config"
version: "0.1.1"
status: completed
created: 2026-09-18
updated: 2026-09-18
author: lane
priority: P3
phase: "v3.2.0 target"
module: ".claude/output-styles/moai"
lifecycle: spec-anchored
tags: "t906, output-style, persona, template-mirror, docs"
tier: M
---

# SPEC-OUTPUT-STYLE-SWITCH-PATH-001

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-09-18 | 0.1.0 | Plan-phase authoring. Card t906, scoped by the lead to the switch-guidance sites only; docs-site excluded and recorded as a proposed follow-up. |
| 2026-09-18 | 0.1.1 | Clarification resolved before Implementation Kickoff Approval: the lead ruled option (가) — `moai.md` gains no switch guidance and both copies stay byte-unchanged; the card changes four files. The sibling-asymmetry question split to card t936. No requirement, criterion, or file-scope changed — the ruling confirmed the authored scope. |

Lineage. Card **t878** measured the `/output-style` slash command live on Claude
Code 2.1.275 and recorded, as its finding F3, that the three MoAI persona
output-style files were an intentional follow-up rather than part of that card's
scope. This SPEC is that follow-up.

This SPEC covers ONE gap: the persona files teach a single route to switching
output styles, and it is not the route the runtime documents.

## §A. Context — measured

Measuring tree: this worktree at base `9dcbc3dbe` (branch
`WT-output-style-switch-path`). Every figure below was measured in this tree, in
this run; none is carried over from the card text without re-measurement.

### A.1 The surface

Six files carry the persona bodies — three working copies and three template
mirrors:

| Working copy | Template mirror |
|---|---|
| `.claude/output-styles/moai/moai.md` | `internal/template/templates/.claude/output-styles/moai/moai.md` |
| `.claude/output-styles/moai/moai-easy.md` | `internal/template/templates/.claude/output-styles/moai/moai-easy.md` |
| `.claude/output-styles/moai/moai-learn.md` | `internal/template/templates/.claude/output-styles/moai/moai-learn.md` |

`cmp` reports all three pairs **identical**. This is a property of this file
group, not of the repository: the working-copy and template trees diverge
deliberately elsewhere, so nothing here generalizes to other mirrored paths. The
consequence for this SPEC is narrow and load-bearing: each edit is applied to
both sides with the same content, and the identity is re-measured afterwards
(AC-OSP-003).

### A.2 The gap

Anchored sweep, run over all six files:

```
grep -rn -E '/output-style([^-a-zA-Z]|$)' \
  .claude/output-styles/moai/ internal/template/templates/.claude/output-styles/moai/
```

Result: **zero hits**. No persona file names the `/output-style` slash command.

**Corrected premise — the earlier reading was a false positive.** An unanchored
`grep -c '/output-style'` reports **2** hits in `moai.md`. Both are the substring
of a file path, at lines 245 and 268:

```
.claude/rules/moai/core/output-style-localization-catalogue.md
```

Neither is a slash-command reference. The unanchored count is therefore not
evidence that `moai.md` mentions the command, and the anchored form is the only
one this SPEC's criteria use. The false positive is recorded here rather than
silently dropped, because the same unanchored pattern will be reached for again
by the next reader of these files. Both counts are re-measured after the change
(AC-OSP-010), so the distinction stays visible rather than becoming folklore.

### A.3 Where switch guidance actually lives

`/config` appears **7 times in each of the three files**, but most occurrences are
the configuration path `.moai/config/sections/language.yaml` and have nothing to
do with switching styles. Separating the two:

| File | `/config` total | of which `.moai/config/sections/language.yaml` | switch-guidance sites | Lines |
|---|---|---|---|---|
| `moai-easy.md` | 7 | 3 | **4** | 21, 493, 496, 533 |
| `moai-learn.md` | 7 | 6 | **1** | 35 |
| `moai.md` | 7 | 7 | **0** | — |

**`moai.md` gives no output-style switch guidance at all.** This is an *absence*,
not a wrong statement: there is no incorrect instruction in `moai.md` to correct,
and consequently no "existing `/config` path" there to add a second route
alongside. Whether the professional persona should gain switch guidance it has
never carried was the one open scope question of this plan phase; **the lead
settled it on 2026-09-18 — `moai.md` is untouched, the card changes four files**
(`plan.md` §B.2 carries the ruling and its reasoning). The asymmetry the
measurement exposed — of the three personas, only `moai.md` has no exit route —
is a separate design question, split to card **t936** rather than dismissed.

The five switch-guidance sites are therefore the whole editable population, and
they live in four of the six files.

### A.4 The command is live

Probed in this session, from an isolated scratchpad directory, on Claude Code
**2.1.276**:

```
$ claude --bare -p "/output-style"
Output style: default

Available styles:
- default (current)
- Proactive: …
…
Usage: /output-style <style>
```

exit 0. The command exists and reports its own usage form.

**Two gaps in this probe, named rather than papered over.** It confirms the
command is *recognized* and prints its usage; it does NOT confirm that
`/output-style <style>` successfully *switches* a style, and it does not confirm
in-TUI typed behaviour — neither was re-probed on 2.1.276. The scratchpad also
carries no MoAI styles, so the listing shows built-in styles only; it is evidence
about the command, not about the MoAI style names. Card t878 measured the same
command on 2.1.275 (`.moai/reports/t878/verdict.md`, primary checkout, present
and readable). These residuals bound what this SPEC may claim: it adds a
documented route the runtime advertises, and it does not assert an end-to-end
switch it did not observe.

### A.5 docs-site — measured, and deliberately untouched

An anchored `/output-style` sweep over `docs-site/content/` returns **zero hits**
in all four locales. `claude-code/foundations/interactive-mode.md` exists in
en / ja / ko / zh and carries an output-style section, but its body covers only
`/config` theme and display options and the `/btw` / `/recap` extras — it
instructs no style switch by either route.

So docs-site currently teaches *neither* route. Adding the command there would be
a new user-facing surface in four locales, not a correction. It is out of scope
(§D) and recorded as a proposed follow-up card; the measured spread stays at
zero, by decision rather than by oversight.

## §B. Requirements

### B.1 The addition

- **REQ-OSP-001** — Each of the five measured output-style switch-guidance sites shall name **both** routes: the `/config` → Output style navigation path and the `/output-style` slash command.
- **REQ-OSP-002** — The addition shall be **additive**. Where a site names `/config` today, that reference shall survive the change; the slash command shall not replace the navigation path at any site.
- **REQ-OSP-003** — The `/output-style` reference shall be written in a form the anchored locator `/output-style([^-a-zA-Z]|$)` matches, so the command is findable as a command rather than as a path fragment.

### B.2 Register

- **REQ-OSP-004** — The text added at each site shall be written in that file's persona register: professional for `moai`, colloquial and warm for `moai-easy`, Socratic-tutor for `moai-learn`.
- **REQ-OSP-005** — The sentence added to `moai-easy.md` and the sentence added to `moai-learn.md` shall not be byte-identical to one another. One sentence pasted into every persona satisfies REQ-OSP-001 while violating REQ-OSP-004, and this requirement is what makes that failure mechanically visible.

### B.3 Boundaries

- **REQ-OSP-006** — Occurrences of `/config` that form the configuration path `.moai/config/sections/language.yaml` shall not be modified. The per-file count of that path shall be unchanged by this SPEC.
- **REQ-OSP-007** — Each working copy and its template mirror shall remain byte-identical after the change, carrying the same added content on both sides.
- **REQ-OSP-008** — The text added to the three template mirrors shall carry no content class the template-neutrality guards reject: no `/Users/` path, no `CLAUDE.local.md` reference, no `PR #N` reference, no SPEC ID, no REQ token, no internal date, no commit SHA.
- **REQ-OSP-009** — No `/moai` slash command, agent definition, or skill body shall be modified. This SPEC changes documentation content only; no runtime behaviour changes.

### B.4 Embed coupling

- **REQ-OSP-010** — Editing `internal/template/templates/**` changes what the binary embeds through `//go:embed`, so an embed refresh is **owed** by the change. That refresh is the **batch-close** step — a single `make build` plus `moai doctor --check "Agent Emit Embed"` run once by the batch lead — and shall NOT be performed by this SPEC's run phase. No criterion of this SPEC shall require a freshly built binary.

### B.5 Evidence

- **REQ-OSP-011** — The implementation shall be accompanied by the anchored per-file counts on both trees, the three `cmp` results, the unchanged `/config` and `language.yaml` counts, the unchanged-file diff for `moai.md`, the two isolated neutrality-guard runs, and the changed-file list. A count asserted rather than measured does not discharge this requirement.

## §C. Scope

In scope: the five switch-guidance sites at `moai-easy.md` lines 21, 493, 496,
533 and `moai-learn.md` line 35, edited identically in both the working copy and
the template mirror — **four files changed**, all six bound by the identity
constraint.

Files expected to change:

```
.claude/output-styles/moai/moai-easy.md
.claude/output-styles/moai/moai-learn.md
internal/template/templates/.claude/output-styles/moai/moai-easy.md
internal/template/templates/.claude/output-styles/moai/moai-learn.md
```

Both `moai.md` copies are byte-unchanged, verified by diff rather than by
intention (AC-OSP-006).

## §D. Out of Scope

The following are deliberately out of scope for this SPEC.

### Out of Scope — adding switch guidance to moai.md

- `moai.md` carries **zero** switch-guidance sites (§A.3). Adding one would introduce user-facing guidance the professional persona has never carried, which is a content decision rather than the correction this card was opened for. The exclusion is **settled, not pending**: the lead ruled on 2026-09-18 that a file is not opened to fix a defect it does not have — the same predicate that cut docs-site out below — so `moai.md` stays untouched and this card changes four files. Reasoning and consequences: `plan.md` §B.2.
- The sibling-asymmetry question the measurement raised — only `moai.md` of the three personas has no exit route — is carried by card **t936**, which must first decide whether that silence is a defect or a deliberate design choice; if deliberate, recording that fact is t936's deliverable. It is kept out of this SPEC so a design judgment is not adjudicated inside a four-file correction diff.
- Both `moai.md` copies are asserted byte-unchanged, so an incidental edit fails a criterion instead of passing unnoticed.

### Out of Scope — docs-site

- No page in `docs-site/content/` teaches an output-style switch by either route (§A.5). Adding the command there is a **new** four-locale user-facing surface governed by the 4-locale same-PR obligation, not a correction of a wrong statement. Recorded as a **proposed follow-up card**: "add output-style switch guidance to `claude-code/foundations/interactive-mode.md` in all four locales". This SPEC does not act on it, and the measured docs-site spread stays at zero.

### Out of Scope — replacing the /config route

- Substituting `/output-style` for `/config` → Output style is rejected (REQ-OSP-002). Both routes are live; the navigation path is discoverable without knowing a command name, and removing it would trade one incomplete instruction for another.

### Out of Scope — the embed refresh and any binary-dependent check

- `make build` and `moai doctor --check "Agent Emit Embed"` belong to the batch-close step, not to this card (REQ-OSP-010). No criterion here is decided by a built binary; acceptance rests on source-level predicates only.

### Out of Scope — runtime behaviour and adjacent surfaces

- No `/moai` command, agent, skill, hook, or rule file is touched (REQ-OSP-009).
- Verifying that `/output-style <style>` actually completes a switch, or that the MoAI style names appear in its listing, is **not** claimed by this SPEC — neither was observed (§A.4).
- Declined forms, named rather than left implicit: writing the command as `output-style` without the leading slash (the anchored locator would not match it, REQ-OSP-003); writing it inside a file path where it becomes the §A.2 false positive again; and adding the command to `moai.md`'s §9 configuration prose, which is a `language.yaml` reference site and not switch guidance (REQ-OSP-006).

## §E. Constraints

| # | Constraint | Enforced by |
|---|---|---|
| C1 | All three working-copy / template-mirror pairs byte-identical after the change | REQ-OSP-007, AC-OSP-003 |
| C2 | `/config` per-file count unchanged at 7 / 7 / 7 | REQ-OSP-002, AC-OSP-004 |
| C3 | `.moai/config/sections/language.yaml` per-file count unchanged at 7 / 3 / 6 | REQ-OSP-006, AC-OSP-005 |
| C4 | Both `moai.md` copies byte-unchanged | §D, AC-OSP-006 |
| C5 | Template text carries no rejected neutrality class | REQ-OSP-008, AC-OSP-007 |
| C6 | No binary-dependent criterion; embed refresh deferred to batch close | REQ-OSP-010, §D |
| C7 | Every criterion asserting a **zero** or an **unchanged** count names the non-empty-sweep control that makes it mean something | AC-OSP-001, AC-OSP-006, AC-OSP-010 |

C7 is scoped to zero-and-unchanged assertions deliberately. Only those can be
satisfied by a sweep that never ran — a mistyped path, a wrong working directory,
or a pattern matching nothing all yield the same clean result as success. Each
such criterion therefore carries a companion measurement that must be **non-zero**
in the same run: AC-OSP-001's own before/after counts witness the sweep reached
the files, AC-OSP-006 pairs its empty diff with a non-empty diff on the files
that did change, and AC-OSP-010's unanchored count of `2` is itself the witness
for the anchored `0`. A criterion asserting a specific non-zero count
(AC-OSP-002, AC-OSP-004, AC-OSP-005) is self-witnessing and needs no companion.
