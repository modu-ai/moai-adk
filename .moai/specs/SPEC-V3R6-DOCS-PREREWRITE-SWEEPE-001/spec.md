---
id: SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001
title: "Docs Pre-Rewrite Sweep E — stale slash citations in the FAQ, feedback, foundations, and verification-claim pages (card t1725)"
version: "0.1.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, cleanup, stale-citation, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, SPEC-V3R6-DOCS-PREREWRITE-GATE-001, SPEC-V3R6-DOCS-PREREWRITE-README-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001 — Sweep E (child of the docs pre-rewrite umbrella)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-19 | 0.1.0 | manager-spec | Initial plan-phase authoring for card t1725 (Class C, docs-only), child SWEEPE of SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (REQ-PRC-001 page part). Every count was measured on tree commit 74ad197380a0 (full `74ad197380a0711cdb9816d4e37c94d29af730ea`). Operator ruling d-20261019T015908Z-220a (Q13, Q17) is applied. Status `draft`. |

## 1. Overview (WHY)

The umbrella's closing check for REQ-PRC-001 (E-04) counts slash forms that the commands surface does not register. Ten page files under `docs-site/content/` carry 18 such lines (acceptance.md §C, S-01): the `/moai init` misconception bullet in the four FAQ pages, the `/moai fb`, `/moai bug`, and `/moai issue` alias lines in the four moai-feedback pages, one prose phrase "the /moai relationship" in the English foundations index, and one retired-feature narrative citing `/moai brain` and `moai brain` in the English verification-claim-integrity page. This child fixes those lines and defines the example marker that operator verdict Q24 (decision d-20261014T060405Z-0fcd) requires. The marker definition is this child's (umbrella defect B03).

Measured findings that shape the fix (all at 74ad197380a0):

- The English FAQ line 24 (`/moai init`) and the English verification-claim-integrity line 84 (`/moai brain`) carry no example marker (S-04: no `<!--` comment exists in either file; S-03: no `slash-example` marker in any of the ten files). Under Q24 both are fixed, not exempt.
- Resolution is wider than the commands directory. The router `.claude/skills/moai/SKILL.md` Priority 1 (lines 66 to 82, identical in `internal/template/templates/.claude/skills/moai/SKILL.md`) lists `feedback (aliases: fb)` and `project (aliases: init)`. So `/moai fb` and `/moai init` resolve to a workflow although neither is a file in `.claude/commands/moai/`. This makes the FAQ sentence "Typing `/moai init` in Claude Code does nothing" false, not merely stale. `/moai bug`, `/moai issue`, `/moai relationship`, and `/moai brain` have no Priority 1 match; the two feedback aliases `bug` and `issue` are therefore wrong documentation (natural-language routing sends bug language to `fix`, SKILL.md line 99).
- `moai brain` is not a CLI verb: the tree build exits 1 with `Unknown command "brain" for "moai"` (S-06). The umbrella families E-01 to E-03 do not include it, so the umbrella count under-reports this page.
- The CLI families E-01, E-02, and E-03 match nothing in the ten files (S-05). Every other inline-code CLI verb in the ten files (`moai init`, `version`, `update`, `profile setup`, `plan`, `feedback scrub`, `feedback queue`, `spec audit`) exists in the tree build.

## 2. Definitions

- **Ten page files** (the scope): `docs-site/content/{en,ja,ko,zh}/getting-started/faq.md`; `docs-site/content/{en,ja,ko,zh}/utility-commands/moai-feedback.md`; `docs-site/content/en/claude-code/foundations/_index.md`; `docs-site/content/en/core-concepts/verification-claim-integrity.md`.
- **Resolving slash form**: `/moai:<name>` where `<name>.md` is a file in `.claude/commands/moai/`; or `/moai <word>` where `<word>` is a Priority 1 name or alias in the router; or `/moai "<free text>"` (default pipeline, router Priority 4). **Resolving CLI form**: a verb or sub-verb accepted by the tree build of `moai`.
- **Closing pattern** (the umbrella's E-04 family): `/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)`. It is stricter than resolution: it also matches the router aliases `fb` and `init`. Edits satisfy both.
- **Example marker** (defined here, REQ-PRC-001-001): an HTML comment `<!-- slash-example: <reason> -->` on the flagged line or the line immediately above it, with a non-empty reason. Nothing wider (a heading or section does not count). An empty reason does not exempt. Same grammar and scope as the `moving-ref-ok` marker already in the tree (`internal/spec/lint_movingref.go`), and Hugo renders HTML comments invisibly here (`docs-site/hugo.toml` `markup.goldmark.renderer.unsafe = true`).

## 3. Requirements (GEARS)

- **REQ-PRC-001** (Ubiquitous; page part, inherits the umbrella text): The ten page files **shall** cite only `moai` CLI verbs and slash forms that resolve on the tree, as defined in § Definitions.
- **REQ-PRC-001-001** (Ubiquitous): The example marker **shall** be the HTML comment `<!-- slash-example: <reason> -->` with a non-empty reason, scoped to the flagged line or the line immediately above, and a line is exempt from REQ-PRC-001 only when it carries that marker (Q24).
- **REQ-PRC-001-002** (Event-driven): When a line in the ten page files cites an unresolved form and carries no example marker, the sweep **shall** replace the citation with the current equivalent where one exists, and otherwise delete it (Q13, operator ruling d-20261019T015908Z-220a).
- **REQ-PRC-001-003** (Ubiquitous): Every locale (ko, en, ja, zh) that carries a stale citation named in REQ-PRC-001-002 **shall** receive the same fix in the same change (Q17, operator ruling d-20261019T015908Z-220a); the two English-only pages are edited alone because the other three locales carry no stale form (acceptance S-07).
- **REQ-PRC-001-004** (Unwanted): The sweep **shall not** edit a line whose citations all resolve, and **shall not** edit a file outside the ten page files.
- **REQ-PRC-001-005** (Ubiquitous): After the sweep, the strict docs i18n check **shall** still pass (parity, frontmatter, H1, glossary).

## 4. Edit plan by line (planner defaults, all within the Q13 ruling)

| File(s) | Line (en) | Current | Fix |
|---------|-----------|---------|-----|
| `getting-started/faq.md` x4 (en 24, ja 24, ko 28, zh 24) | `/moai init` "does nothing" | false (router alias `init` to `project`) | Replace: `moai init` is a terminal command that sets up a project; inside Claude Code, project documentation is generated with `/moai project` |
| `utility-commands/moai-feedback.md` x4 (lines 31 to 34) | `# Short aliases` + `/moai fb`, `/moai bug`, `/moai issue` | `bug` and `issue` do not resolve; `fb` resolves only via the router | Replace the comment and three lines with `# Direct slash command` and `> /moai:feedback` (registered, `.claude/commands/moai/feedback.md`) |
| `claude-code/foundations/_index.md` (en 39) | "and the /moai relationship" | prose, matches the closing pattern | Replace with "and where `/moai` sits" (mirrors the section "Where the /moai Command Sits" of the linked page) |
| `core-concepts/verification-claim-integrity.md` (en 84) | `/moai brain` command and `moai brain` CLI in the producer list | both retired; no current equivalent | Delete those two list items only; the rest of the paragraph stays (Q13: no equivalent exists, so delete) |

## 5. Interface to the GATE child

The GATE citation checker (SPEC-V3R6-DOCS-PREREWRITE-GATE-001) must honour the marker grammar of § Definitions. This child adds zero markers at the planner defaults, so the interface is dormant; it becomes live only if the operator chooses a marker for line 84 (decision QE-2).

## 6. Decisions

Rows use the decision-index shape. The dispatch limited this card to three SPEC files, so the rows live here rather than in a separate `decision-index.md` (gap G-5 in acceptance.md).

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261019T015908Z-220a (decision board, `moai decision read --scope card:t1725`)
Why unresolved: the umbrella row (copy). Replace where a current document or command exists, otherwise delete.
Class: implementation-level
Child: README and SWEEPA to SWEEPF (this child applies it to its ten files)
Operator verdict: replace the stale citation with the current equivalent where one exists, delete where none exists — operator, d-20261019T015908Z-220a (AskUserQuestion, 2026-10-19, scope=standing)

### Q17: Do the same edits land in en, ja, ko, and zh in the same change?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261019T015908Z-220a
Why unresolved: the umbrella row (copy).
Class: implementation-level
Child: cross-cutting (this child applies it to its ten files)
Operator verdict: all four locales (ko, en, ja, zh) change in the same card and the same change — operator, d-20261019T015908Z-220a (AskUserQuestion, 2026-10-19, scope=standing)

### Q24: Which narrative examples are exempt from the stale-citation check (N2)?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261014T060405Z-0fcd
Why unresolved: the full row stays in the umbrella; this child records the marker definition (REQ-PRC-001-001) and the measured marker state (neither line carries one).
Class: implementation-level
Child: SWEEPE
Operator verdict: exempt only narrative examples that carry an explicit example marker; lines without a marker are fixed — operator, d-20261014T060405Z-0fcd (2026-10-14T06:04:05Z)

### QE-1: How are the three feedback alias lines handled?
Label: DECIDED (by Q13 above; the row records the residual wording choice)
Authority anchor: operator record d-20261019T015908Z-220a (standing ruling, Q13)
Why unresolved: `fb` resolves only as a router alias; `bug` and `issue` do not resolve. The ruling requires replace-where-an-equivalent-exists; `/moai:feedback` is the registered equivalent for all three. Choice left: delete instead of replace.
Class: implementation-level
Default: replace with `> /moai:feedback` (rule: the Q13 ruling)
Alternate: delete the lines and keep only the standard form
Operator verdict: DEFAULT-APPLIED by Q13 ruling d-20261019T015908Z-220a

### QE-2: Does verification-claim-integrity line 84 stay fixed, or gain an example marker?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the line sits under the heading "Worked Example — Retention-Claim Hazard" and its text names retired commands as history. The Q24 verdict exempts only lines that carry a marker; adding one in this change would exempt the line the verdict says to fix. Whether a marker added by the sweep counts is the operator's reading.
Class: implementation-level
Default: fix without a marker, deleting the `/moai brain` and `moai brain` list items (rule: the option the Q24 verdict and Q13 ruling both name)
Alternate: keep the paragraph and add `<!-- slash-example: retired-feature incident narrative -->`
Operator verdict:

### QE-3: Is the closing check stricter than resolution?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the umbrella pattern also matches the router aliases `fb` and `init`, which do resolve. Keeping the pattern forces removal of the `fb` and `init` citations. A carve-out for router aliases is an umbrella amendment, not this child's.
Class: implementation-level
Default: keep the umbrella pattern unchanged (rule: preserves the umbrella's closing check)
Alternate: amend the umbrella pattern to exclude router aliases, then keep `/moai fb`
Operator verdict:

## 7. Out of Scope

### Out of Scope — other children's files

- Any page or README outside the ten page files; the other sweeps (SWEEPA to SWEEPD, SWEEPF), README, and FLAG own those.
- The `--from-brain` flag and every other flag citation (decision Q18, FLAG child).

### Out of Scope — tooling and code

- Writing the citation checker (GATE child) and any change under `internal/` or `cmd/`.
- Editing `.claude/skills/moai/SKILL.md` or the commands directory.

### Out of Scope — rewrite content

- New prose beyond the replacement lines of § 4, and the v3.2 docs rewrite.
- Translating pages that carry no stale form; the ja, ko, and zh foundations pages already lack the stale phrase.
