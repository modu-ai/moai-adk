---
id: SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001
title: "Docs Pre-Rewrite Sweep E — stale slash citations in the FAQ, feedback, foundations, and verification-claim pages (card t1725)"
version: "0.2.0"
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
| 2026-10-19 | 0.2.0 | manager-spec | Plan-revision round 1 after plan audit iteration 1 returned FAIL 0.69 (`.moai/reports/t1725/plan-audit-verdict.md`, receipt rcpt-93c8df0b6680dfc7dac83fb9; defects D1 to D12). REQ-PRC-001-004 reconciled with the edit plan: the closing pattern, not resolution, defines the flagged set (D2). Sub-requirement IDs renamed to the form the traceability verb reads, `REQ-SWEEPE-001` to `REQ-SWEEPE-006` (D8). Scope grows from ten to twelve page files: the ja and zh foundations index rows take the Q17 counterpart edit (D12). Replacement lines are pinned verbatim in § 4 (D4). Acceptance criteria re-classed so every release-blocking criterion has a RED-now cell (D1), made independent of QE-2 (D3), and extended with a whole-tree scope fence (D5) and a verb-resolution criterion (D6). Rows QE-4 and QE-5 added; QE-1 to QE-3 normalised. Status stays `draft`. |

## 1. Overview (WHY)

The umbrella's closing check for REQ-PRC-001 (E-04) flags slash forms it counts as unregistered. Ten page files under `docs-site/content/` carry 18 such lines (acceptance.md § D, S-01): the `/moai init` misconception bullet in the four FAQ pages, the `/moai fb`, `/moai bug`, and `/moai issue` alias lines in the four moai-feedback pages, one prose phrase "the /moai relationship" in the English foundations index, and one retired-feature narrative citing `/moai brain` and `moai brain` in the English verification-claim-integrity page. This child fixes those lines, applies the Q17 counterpart edit to the ja and zh foundations index rows, and defines the example marker that operator verdict Q24 (decision d-20261014T060405Z-0fcd) requires. The marker definition is this child's (umbrella defect B03).

**Which test governs.** The unit of work is a line the closing pattern flags with no example marker (the flagged set, § 2) — not only a line whose citation fails to resolve. Operator verdict Q24 (N2) says an unmarked flagged line is fixed. The umbrella calls the flagged forms "unregistered", which is looser than router resolution: `/moai init` and `/moai fb` do resolve through router aliases, yet both are flagged and both are edited. Resolution therefore decides what a replacement may cite; the closing pattern decides which lines are edited (decision QE-3: the umbrella pattern stays unchanged).

Measured findings that shape the fix (all at 74ad197380a0; the docs-site, `.claude`, `internal`, `cmd`, and `scripts` trees are unchanged between that commit and the plan HEAD bb7d928f68, acceptance.md S-00):

- The English FAQ line 24 (`/moai init`) and the English verification-claim-integrity line 84 (`/moai brain`) carry no example marker (S-04: no `<!--` comment exists in either file; S-03: no `slash-example` marker in any page file). Under Q24 both are fixed, not exempt.
- Resolution is wider than the commands directory. The router `.claude/skills/moai/SKILL.md` Priority 1 (lines 66 to 82, identical in `internal/template/templates/.claude/skills/moai/SKILL.md`) lists `feedback (aliases: fb)` and `project (aliases: init)` (S-08). So `/moai fb` and `/moai init` resolve to a workflow although neither is a file in `.claude/commands/moai/`. This makes the FAQ sentence "Typing `/moai init` in Claude Code does nothing" false, not merely stale. `/moai bug`, `/moai issue`, `/moai relationship`, and `/moai brain` have no Priority 1 match.
- `moai brain` is not a CLI verb: the tree build exits 1 with `Unknown command "brain" for "moai"` (S-06d). The umbrella families E-01 to E-03 do not include it, so the umbrella count under-reports this page.
- The CLI families E-01, E-02, and E-03 match nothing in the page files (S-05). Every other inline `moai <word>` token in the page files (`init`, `version`, `update`, `profile setup`, `plan`, `feedback scrub`, `feedback queue`, `spec audit`) exists in the tree build (S-06b, S-06c, S-06d).
- The ja and zh foundations index rows translate the same clause as the English row ("と /moai の関係", "与 /moai 的关系", acceptance.md S-07); the ko row has no `/moai` clause. They are not closing-pattern matches (decision QE-4).
- The FAQ page also says "Running `moai plan` in the terminal does nothing". `moai plan` is a CLI verb (`Plan-Phase CLI utilities`, S-06d), so that sentence is an adjacent false claim, but it is not a flagged line (decision QE-5).

## 2. Definitions

- **Flagged-line files** (ten): `docs-site/content/{en,ja,ko,zh}/getting-started/faq.md`; `docs-site/content/{en,ja,ko,zh}/utility-commands/moai-feedback.md`; `docs-site/content/en/claude-code/foundations/_index.md`; `docs-site/content/en/core-concepts/verification-claim-integrity.md`.
- **Counterpart files** (two): `docs-site/content/{ja,zh}/claude-code/foundations/_index.md`. They carry the translation of an edited English line without carrying a flagged line.
- **Page files** (twelve, the scope): the flagged-line files and the counterpart files.
- **Closing pattern** (the umbrella's E-04 family): `/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)`. It is stricter than resolution: it also matches the router aliases `fb` and `init`.
- **Flagged set**: every line of the page files that the closing pattern matches and that carries no example marker. Resolution of the cited form through a router alias does not remove a line from the flagged set.
- **Q17 counterpart line**: the line of a counterpart file that translates an edited flagged line. At this plan's measurement that is the ja and zh foundations index row (S-07).
- **Resolving slash form** (what a replacement may cite): `/moai:<name>` where `<name>.md` is a file in `.claude/commands/moai/`; or `/moai <word>` where `<word>` is a Priority 1 name or alias in the router; or `/moai "<free text>"` (default pipeline, router Priority 4). **Resolving CLI form**: a verb or sub-verb accepted by the tree build of `moai`.
- **Example marker** (defined here, REQ-SWEEPE-001): an HTML comment `<!-- slash-example: <reason> -->` on the flagged line or the line immediately above it, with a non-empty reason. Nothing wider (a heading or section does not count). An empty reason does not exempt. Same grammar and scope as the `moving-ref-ok` marker already in the tree (`internal/spec/lint_movingref.go`), and Hugo renders HTML comments invisibly here (`docs-site/hugo.toml` `markup.goldmark.renderer.unsafe = true`).

## 3. Requirements (GEARS)

ID scheme: `REQ-PRC-001` is the umbrella pointer, kept under its umbrella ID. The five sub-requirements and the replacement-content requirement carry the child-local prefix `REQ-SWEEPE-`, which the traceability verb reads as separate requirements (a numeric tail after `PRC-001` would be read as `REQ-PRC-001` itself). Each `REQ-SWEEPE-` item refines the umbrella's REQ-PRC-001 page part.

- **REQ-PRC-001** (Ubiquitous; page part, umbrella text kept): The page files of this child **shall** cite only `moai` CLI verbs and slash forms that resolve on the tree, as defined in § 2.
- **REQ-SWEEPE-001** (Ubiquitous): The example marker **shall** be the HTML comment `<!-- slash-example: <reason> -->` with a non-empty reason, scoped to the flagged line or the line immediately above, and a line is exempt from the sweep only when it carries that marker (Q24).
- **REQ-SWEEPE-002** (Event-driven): When a line of the page files belongs to the flagged set, the sweep **shall** replace the citation with the current equivalent where one exists, and otherwise delete it (Q13, operator ruling d-20261019T015908Z-220a).
- **REQ-SWEEPE-003** (Ubiquitous): Every locale (ko, en, ja, zh) whose page carries a flagged line **shall** receive the same fix in the same change, and every counterpart line **shall** receive the matching wording change in the same change (Q17, operator ruling d-20261019T015908Z-220a); the two English-only pages are edited alone because the other locales carry no flagged line there (acceptance S-07).
- **REQ-SWEEPE-004** (Ubiquitous): The sweep **shall** change only the flagged-set lines and the Q17 counterpart lines of the page files, **shall** leave every other line of the page files byte-identical, and **shall** change no file outside the page files other than this SPEC directory and `.moai/reports/t1725/`.
- **REQ-SWEEPE-005** (Ubiquitous): After the sweep, the strict docs i18n check **shall** still pass (parity, frontmatter, H1, glossary).
- **REQ-SWEEPE-006** (Ubiquitous): Each locale's FAQ replacement bullet **shall** state that `moai init` is a terminal command that sets up a project and that, inside Claude Code, project documentation is generated with `/moai project`, and **shall not** state that `/moai init` or `/moai project` does nothing; the exact lines are pinned in § 4.

## 4. Edit plan by line (planner defaults, all within the Q13 ruling)

The replacement text below is the plan's authoritative text. Cells S-12a, S-14, and S-15 match it exactly. A run-phase deviation changes both this table and those cells, so it is a blocker report to the orchestrator, not a local rewording.

| File(s) | Line (en) | Current | Fix |
|---------|-----------|---------|-----|
| `getting-started/faq.md` x4 (en 24, ja 24, ko 28, zh 24) | `/moai init` "does nothing" | flagged; the claim is false (router alias `init` to `project`) | Replace the line with the pinned FAQ line below (one line in, one line out) |
| `utility-commands/moai-feedback.md` x4 (lines 31 to 34) | `# Short aliases` + `/moai fb`, `/moai bug`, `/moai issue` | flagged; `bug` and `issue` do not resolve, `fb` resolves only via the router | Replace the comment and three lines with a comment meaning "Direct slash command" (en `# Direct slash command`, ko `# 직접 슬래시 명령`, ja `# 直接スラッシュコマンド`, zh `# 直接斜杠命令`) and the line `> /moai:feedback` (registered, `.claude/commands/moai/feedback.md`) |
| `claude-code/foundations/_index.md` en (39), ja (40), zh (40) | "and the /moai relationship" and its translations | en flagged by the closing pattern; ja and zh are Q17 counterpart rows | en: "and where `/moai` sits" (mirrors the section "Where the /moai Command Sits" of the linked page); ja: `/moai の関係` becomes `/moai の位置づけ`; zh: `/moai 的关系` becomes `/moai 的定位` |
| `core-concepts/verification-claim-integrity.md` en (84) | `/moai brain` command and `moai brain` CLI in the producer list | flagged; both retired, no current equivalent | Delete those two list items only (Q13: no equivalent exists, so delete); the pinned retained fragment below must read exactly as shown |

Pinned FAQ lines (one per locale, whole line):

```text
en: - `moai init` is a terminal command that sets up a project; inside Claude Code, generate project documentation with `/moai project`.
ja: - `moai init` はプロジェクトを設定するターミナルコマンドで、Claude Code の中では `/moai project` でプロジェクト文書を生成します。
ko: - `moai init`은 프로젝트를 설정하는 터미널 명령이고, Claude Code 안에서는 `/moai project`로 프로젝트 문서를 생성합니다.
zh: - `moai init` 是用于设置项目的终端命令；在 Claude Code 中，用 `/moai project` 生成项目文档。
```

Pinned retained fragment of the verification-claim-integrity line 84 (the rest of the paragraph stays as it is):

```text
every one was already gone: `workflows/brain.md`, the `manager-brain` agent, the `/moai project --from-brain` flag, the `templates/.moai/brain/` scaffold, and the docs-site pages in all four locales.
```

The sibling FAQ bullet ("Running `moai plan` in the terminal does nothing — `/moai plan` is only valid inside Claude Code.", en line 23 and its translations) stays untouched (QE-5).

## 5. Interface to the GATE child

The GATE citation checker (SPEC-V3R6-DOCS-PREREWRITE-GATE-001) must honour the marker grammar of § 2. This child adds zero markers at the planner defaults, so the interface is dormant; it becomes live only if the operator chooses a marker for line 84 (decision QE-2). GATE-001 and README-001 have no directory in `.moai/specs` at this plan's HEAD (their cards are not yet planned), so the marker interface could not be read from them; the grammar here is the one the umbrella assigns to this child.

## 6. Decisions

Rows use the decision-index shape and are mirrored field for field in `decision-index.md` (acceptance.md gap G-5). A `DEFAULT-APPLIED` verdict records the planner default (the published Default rule) and is recorded on the decision board by the leader before Kickoff.

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261019T015908Z-220a (decision board, `moai decision read --scope card:t1725`)
Why unresolved: the umbrella row (copy). Replace where a current document or command exists, otherwise delete.
Class: implementation-level
Child: README and SWEEPA to SWEEPF (this child applies it to its page files)
Operator verdict: replace the stale citation with the current equivalent where one exists, delete where none exists — operator, d-20261019T015908Z-220a (AskUserQuestion, 2026-10-19, scope=standing)

### Q17: Do the same edits land in en, ja, ko, and zh in the same change?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261019T015908Z-220a
Why unresolved: the umbrella row (copy).
Class: implementation-level
Child: cross-cutting (this child applies it to its page files)
Operator verdict: all four locales (ko, en, ja, zh) change in the same card and the same change — operator, d-20261019T015908Z-220a (AskUserQuestion, 2026-10-19, scope=standing)

### Q24: Which narrative examples are exempt from the stale-citation check (N2)?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261014T060405Z-0fcd
Why unresolved: the full row stays in the umbrella; this child records the marker definition (REQ-SWEEPE-001) and the measured marker state (neither line carries one).
Class: implementation-level
Child: SWEEPE
Operator verdict: exempt only narrative examples that carry an explicit example marker; lines without a marker are fixed — operator, d-20261014T060405Z-0fcd (2026-10-14T06:04:05Z)

### QE-1: How are the three feedback alias lines handled?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261019T015908Z-220a (standing ruling, Q13, decision board). The ruling sets the rule; the row records the residual wording choice
Why unresolved: `fb` resolves only as a router alias; `bug` and `issue` do not resolve. The ruling requires replace-where-an-equivalent-exists; `/moai:feedback` is the registered equivalent for all three. Choice left: delete instead of replace.
Class: implementation-level
Default: replace with `> /moai:feedback` (rule: the Q13 ruling)
Alternate: delete the lines and keep only the standard form
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:27:27Z manager-spec (planner, revision 1), rule: the Q13 ruling d-20261019T015908Z-220a

### QE-2: Does verification-claim-integrity line 84 stay fixed, or gain an example marker?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the line sits under the heading "Worked Example — Retention-Claim Hazard" and its text names retired commands as history. The Q24 verdict exempts only lines that carry a marker; adding one in this change would exempt the line the verdict says to fix. Whether a marker added by the sweep counts is the operator's reading.
Class: implementation-level
Default: fix without a marker, deleting the `/moai brain` and `moai brain` list items (rule: the option the Q24 verdict and Q13 ruling both name)
Alternate: keep the paragraph and add `<!-- slash-example: retired-feature incident narrative -->` on the line above line 84. The alternate carries its own criteria (acceptance.md § H): the release-blocking set of this SPEC at the default does not depend on it.
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:27:27Z manager-spec (planner, revision 1), rule: the option the Q24 verdict and Q13 ruling both name

### QE-3: Is the closing check stricter than resolution?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the umbrella pattern also matches the router aliases `fb` and `init`, which do resolve; the umbrella's word "unregistered" is looser than router resolution (acceptance.md S-08). Keeping the pattern forces removal of the `fb` and `init` citations, and Q24 (N2) requires an unmarked flagged line to be fixed. A carve-out for router aliases is an umbrella amendment, not this child's.
Class: implementation-level
Default: keep the umbrella pattern unchanged (rule: preserves the umbrella's closing check)
Alternate: amend the umbrella pattern to exclude router aliases, then keep `/moai fb`. That is an umbrella amendment; this SPEC's criteria a, b, d, and i would be re-derived and re-audited at that point.
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:27:27Z manager-spec (planner, revision 1), rule: preserves the umbrella's closing check

### QE-4: Do the ja and zh foundations index rows change with the en row?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261019T015908Z-220a (Q17, decision board)
Why unresolved: the ja and zh rows translate the clause "and the /moai relationship" and are not closing-pattern matches; their linked pages' section headings still say "関係" and "关系". The Q17 ruling asks that locales carrying the same sentence change in the same change.
Class: implementation-level
Default: change both rows to the "position" wording (`/moai の位置づけ`, `/moai 的定位`) (rule: the Q17 ruling)
Alternate: leave the ja and zh rows unchanged (ten page files); they would then read as a translation of the old English wording and the three locales would no longer say the same thing
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:27:27Z manager-spec (planner, revision 1), rule: the Q17 ruling d-20261019T015908Z-220a

### QE-5: Is the sibling FAQ bullet about `moai plan` in scope?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: "Running `moai plan` in the terminal does nothing" is an adjacent false claim (`moai plan` is a CLI verb, S-06d), but the closing pattern does not flag it and the Q13 ruling concerns stale citations.
Class: implementation-level
Default: out of scope; the line stays byte-identical and is recorded as a follow-up observation (rule: the card scope is the line flagged by the closing pattern)
Alternate: correct it in this change; that adds a line to the flagged set by decision, a new pinned text, and criteria of its own
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:27:27Z manager-spec (planner, revision 1), rule: smaller footprint, undo is one revert

## 7. Out of Scope

### Out of Scope — other children's files

- Any page or README outside the twelve page files; the other sweeps (SWEEPA to SWEEPD, SWEEPF), README, and FLAG own those.
- The `--from-brain` flag and every other flag citation (decision Q18, FLAG child).

### Out of Scope — tooling and code

- Writing the citation checker (GATE child) and any change under `internal/` or `cmd/`.
- Editing `.claude/skills/moai/SKILL.md` or the commands directory.

### Out of Scope — rewrite content

- New prose beyond the replacement lines of § 4, and the v3.2 docs rewrite.
- Translating pages that carry no stale form; the ko foundations page has no `/moai` clause, and the ja, ko, and zh verification-claim-integrity pages carry no retired-feature text.

### Out of Scope — adjacent false claims

- The FAQ bullet "Running `moai plan` in the terminal does nothing" (en line 23 and its translations): not a line the closing pattern flags; recorded as a follow-up observation (decision QE-5).
- Any other prose accuracy question in the page files that the closing pattern does not flag.
