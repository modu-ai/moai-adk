---
id: SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001
title: "Docs Pre-Rewrite Cleanup — index over eleven child SPECs (card t1655)"
version: "0.3.1"
status: draft
created: 2026-10-14
updated: 2026-10-14
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "README.md, README.ko.md, README.ja.md, README.zh.md, docs-site, .github/workflows, assets/images"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, ci, cleanup, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001, SPEC-DOCS-SITE-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 — Docs Pre-Rewrite Cleanup (index)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-14 | 0.1.0 | manager-spec | Initial plan-phase authoring for card t1655 (Class C), which is stage 5-0 ("지금 가능한 정리", the cleanup that can land before the rewrite) of the v3.2 upgrade plan. Every count below was measured on tree commit f35103480f (tree ca8fc594f) in this planning pass. The card's figures (12 stale commands, 8 CG pages) were re-measured, not carried over. |
| 2026-10-14 | 0.1.1 | manager-spec | Revision round after the plan-audit verdict FAIL 0.63 (`.moai/reports/t1655/plan-audit-verdict.md`, audited at f35103480f; defect numbers below follow that verdict). **Fixed:** D1 (every RED-now cell re-authored as one read-only invocation, pinned to f35103480f, with full verbatim stdout and a separate exit-code field; the wrapper exception removed; plan.md §A.3 rewritten); D2 (E-01 and E-03 restated as CLI forms that exclude the slash forms and the retained retirement notice; E-03 re-measured at 8 lines; E-01 at 60 lines in 48 files); D3 (`--factory-run` and `--clear-policy` removed from AC-PRC-001; flags are governed by Q18); D4 (M2 names the edit sites: the pull-request strictness step at `docs-i18n-check.yml` lines 76–77 and the expression at line 85; not yet edited); D7 (acceptance §D covers all 15 ACs, and M6 uses the same set); D9 (plan §B M0 lists the open rows each milestone depends on; M1 and M3 are gated behind M0); D10 (planner values moved to `Planner default applied:`; `Operator verdict:` empty in all 22 rows; Q15 and Q16 reopened with class justification); D11 (Q17 and gap G14 corrected: the four-locale obligation is at `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65, outside the authority register); D12 (done checks added: REQ-PRC-002 via the M2 pull-request step in AC-PRC-002, REQ-PRC-004 link presence in AC-PRC-004, REQ-PRC-007 Hextra check in AC-PRC-007); D13 (REQ-PRC-006 conditional on Q3; no budget stated); D14 (cited-verbs defined as a counted metric with a fixed output format and producer inputs, plan §B M1; CI wiring in M2); D15 (binary rebuilt from the Makefile `build` recipe at f35103480f; stamp confirmed in the version output; cited in acceptance §C); D16 (AC-PRC-005 rescoped to the paths searched, and the tracked historical report named); D17 (stage-0 non-overlap reworded as not yet checked); D18 (REQ-PRC-003 conditional on Q8). **Fixed, optional:** O1 (priority P1, as the card title records); O2 (REQ-PRC-005, -009, -011 restated in GEARS Ubiquitous form); O3 (Q9 premise restated); O4 (Q11 probed set named); O5 (Nextra-named defined in REQ-PRC-007); O8 (plan §B M1: the producer reads COMMANDS blocks, not exit codes); O9 (implementation detail removed from REQ-PRC-012; trailing rationale removed from REQ-PRC-002); O10 (the 0.1.0 row's statement that the card's figures were re-measured is withdrawn; the card's '12' has no counting method in the tree, and the Overview no longer presents it as a measured count); O12 (wrapper exception removed under D1). **Found in re-measurement, fixed within the same rewrite:** N1 (E-02's `web console` alternative matched UI names; the family is narrowed to CLI forms and the two moai-todo example pages are exempt, 42 → 24 lines); N3 (REQ-PRC-004, -005, -007, and -013 presupposed planner-default options of Q1, Q2, Q4, and Q10; they are now conditional, the same class as D18); N4 (Q14's premise was wrong: the moai-codemaps page cites the registered slash command, not the absent CLI verb). **Left open:** D5 (decision Q21: remove the two external changelog rules, or amend REQ-PRC-010); D6 (decision Q22: move REQ-PRC-015 to an operator-held section, or amend the Out of Scope exclusion); D8 (decision Q20: the tier; see the note below); O6 and O7 (cross-SPEC effects and the orphan README_TRANSLATE edit; both need a decision of their own); O9 for REQ-PRC-015 (the trailing operator-held note is kept until Q22 is answered); O11 (the `related_specs` schema row belongs to the schema owner, outside this SPEC's paths); N2 (E-04's narrative and error examples need an exemption decision). Status stays `draft`. The plan is pending re-audit. |
| 2026-10-14 | 0.2.0 | manager-spec | Revision to the umbrella. Three operator records apply: `d-20261014T055241Z-1bc6` (Q20 = (c), split into child SPECs, tier M kept); `d-20261014T060405Z-0fcd` (the eleven child SPECs are approved as they stand; the parent stays the umbrella for REQ-PRC-015, Q12, Q22, the evidence ledger, and M6; N2 is decided as Q24; the leader's B04 ruling adds Q23 and an AC-PRC-011 extension, both owned by REDIR); and `d-20261014T061812Z-f5ca` ("allow": manager-spec may revise this parent into the umbrella, and the lane may commit it). The child cards are issued as t1717 (GATE), t1718 (REDIR), t1719 (RESID), t1720 (README), t1721 (SWEEPA), t1722 (SWEEPB), t1723 (SWEEPC), t1724 (SWEEPD), t1725 (SWEEPE), t1726 (FLAG), and t1727 (SWEEPF). REQ-PRC-001 to REQ-PRC-014 move under their owner children with their texts unchanged; REQ-PRC-015 stays in the parent under an operator-held heading. No requirement or acceptance criterion identifier is renumbered. The 0.1.1 evidence-ledger label for E-04 is corrected from 23 lines to 22 lines. Status stays `draft`. |
| 2026-10-14 | 0.2.1 | manager-spec | Corrections after the operator-approved umbrella revision, per leader ruling d-20261014T070143Z-75ac: G21 sentence tidied; E-04 counted as 22 files and 45 matching lines; Q2 reference scope stated with E-09b and E-09b2 both cited. No requirement text changed; the E-04 count in the AC-PRC-001 partial-criteria bullet changed. |
| 2026-10-14 | 0.2.2 | manager-spec | Re-pin of the RED-now cells to 8816de4dae (the card branch after absorbing main 763193861b), outputs re-recorded verbatim. Supporting, verification, and protocol cells stay at f35103480f with labels; a drift table records the four items that differ (E-05b, E-27b, E-00, E-06 BuildID). E-09b2 and E-28 exclude this SPEC's own directory, per leader ruling d-20261014T072226Z-9b67 (expected E-09b2 = 1, E-28 = 0). Documentation corrections: the E-12 header (18 lines), the E-19b reading and timing-line notes, the Q2 wording, and the 0.2.1 history wording. No requirement text changed. |
| 2026-10-14 | 0.3.0 | manager-spec | Structural revision: route 2 of operator decision `d-20261014T114834Z-e039` (plan-audit iteration 3, FAIL 0.56 at bdb7b764, MP-8 failed, 15 blocking findings; `.moai/reports/t1655/plan-audit-verdict-iter3.md`). This SPEC becomes an index. REQ-PRC-001 to REQ-PRC-014 and AC-PRC-001 to AC-PRC-014 become pointers to the owning children, which carry the text, the criteria, and the RED-now cells. REQ-PRC-015, AC-PRC-015, the evidence ledger (acceptance §C and the section A drift table), decision rows Q12, Q20, Q22, and Q24, and M6 stay here. M6 re-runs each child's committed criteria at the child's commit SHA, and the import rule (plan.md §B.3) imports them by path and SHA, never by copied text. Decision rows Q1 to Q11, Q13 to Q19, Q21, and Q23 move to their owning children as stubs; Q21 is decided inside REDIR and Q7 inside RESID. Q22's owner is proposed as GATE, pending leader confirmation. Findings addressed here: B01 and MP-8 (AC-PRC-012 and AC-PRC-006 leave the release-blocking set of this index; cell E-16e replaces the missing block), B02 (import rule), B03 (AC-PRC-001 gate column and closing check carry Q17 and Q24), B07 (labels of cells that are not red), B08 (references to removed plan sections), B09 (one completion definition; REQ-PRC-015 external closure reported apart), B10 (AC-DVR-019b in Dependencies), and B12 (the pin and verbatim-output statements, the Q24 readings, and the distinct-form count of 18). B04, B05, B06, B11, B13, B14, and B15 move with the children that own them (`.moai/reports/t1655/index-mapping.md`). The E-04 header already reads 22 lines, so no change was needed there. Pre-revision text: commit bdb7b764. Status stays `draft`. |
| 2026-10-14 | 0.3.1 | manager-spec | Q22 owner GATE leader-confirmed (card:t1655, 2026-10-14); stale pending and conditional wording cleared in decision-index.md, spec.md (child table), and acceptance.md G22. No requirement, criterion, or decision answer changed. |

**Tier note (decision Q20, resolved).** Q20 = (c), split into child SPECs by `d-20261014T055241Z-1bc6`. The index keeps tier M. Each child is tier M, with 5 to 14 files, under the 15-file cap. No `design.md` or `research.md` is authored in this revision.

## Overview (WHY)

The v3.2 documentation rewrite (plan stage 5) opens with a cleanup that does not depend on the rewrite itself: stale command references, a stale version badge, a duplicated CG retirement passage, unreferenced and oversized images, Nextra-era residue, dangling redirects, and CI jobs that do not block. The card (t1655) names a completion verdict of "cited-verbs 0" and "CI 3 jobs blocking". This SPEC states what must be true for that verdict and records the RED-now state of each criterion.

As of revision 0.3.0 this SPEC is an index over eleven child SPECs (see § Child SPECs). REQ-PRC-001 to REQ-PRC-014 and AC-PRC-001 to AC-PRC-014 are pointers, and the children carry their text. This index keeps REQ-PRC-015, AC-PRC-015, the evidence ledger, the decision rows Q12, Q20, and Q24, and the closing measurement (M6). The full Q22 row also stays here; Q22 has owner GATE (leader-confirmed). The file-level work moves to the children. The measured-state bullets below are the index's baseline evidence, unchanged from revision 0.1.1.

Measured state on tree f35103480f (full evidence in acceptance.md §C):

- The docs i18n check passes in strict mode (exit 0, 0 errors, 0 warnings; E-16d). Its pull-request job is advisory: the strictness step for pull requests writes `strict=false` (`.github/workflows/docs-i18n-check.yml:77`), so the `continue-on-error` expression at line 85 evaluates to true for pull requests (E-16a), and the job is not among the branch's required status checks (E-21, external).
- `.github/workflows` contains no Hugo build step and no link-check tool or job (E-18a, E-19a).
- No producer for the cited-verbs figure exists in the tracked tree (E-28). The card's figure has no counting method in the tree (gap G-I).
- The public Latest release is v3.1.2 (E-20c, external). The four README badges and `docs-site/hugo.toml` show v3.1.3 (E-20a, E-20b), which is a Draft on the public release list.
- 18 distinct stale command forms are cited (E-01 to E-04, after the exclusions in acceptance §C): 9 absent CLI verbs or sub-verbs (`cg`, `gpt`, `audit`, `codemaps`, `model`, `codex job`, `codex task`, `codex setup`, `spec list`), 2 Mermaid labels that cite `moai run SPEC-ID` and `moai sync SPEC-ID` as CLI verbs, and 7 slash forms that the commands surface does not register (`design`, `issue`, `bug`, `fb`, `init`, `brain`, `relationship`). The `todo rename` and `todo pick` lines on the four moai-todo pages are a natural-language input and an error example, not citations (E-02t), so they are not counted. The card's "12" is carried from the card; its counting method is not in the tree, so this SPEC does not use it as a measured count.
- Two flags cited in the docs (`--factory-run`, `--clear-policy`) are defined and parsed in the Go source (E-05b: 101 lines in 24 files). They are absent from the help text of `factory`, `goal`, `codex`, and `cc` (E-25a). They are real flags, not stale citations; they are outside AC-PRC-001 and governed by decision Q18.

The RED-now values in these bullets are identical at 8816de4dae (revision 0.2.2); the drift table in acceptance section A lists the items that differ.

## Scope (WHAT)

The seven card items map to REQ-PRC-001 through REQ-PRC-015:

1. Stale command references (낡은 명령 12건, the card's label): REQ-PRC-001, REQ-PRC-002.
2. Badge version (배지 버전): REQ-PRC-003.
3. CG passage 8 pages to 1 (CG 문단 8쪽→1쪽): REQ-PRC-004.
4. Image 5.8 MB and OG re-encoding (이미지 5.8MB·OG 재인코딩): REQ-PRC-005, REQ-PRC-006.
5. Nextra residue, old README, llms.txt (Nextra 잔재·옛 README·llms.txt): REQ-PRC-007, REQ-PRC-008, REQ-PRC-009.
6. Redirect repair (리다이렉트 수리): REQ-PRC-010, REQ-PRC-011.
7. i18n CI strict, Hugo build, link-check jobs (i18n CI strict·Hugo 빌드·링크 검사 잡): REQ-PRC-012, REQ-PRC-013, REQ-PRC-014, REQ-PRC-015.

Paths the child SPECs may change (the union; this index changes none of them): `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`; `docs-site/content/**`; `docs-site/vercel.json`; `docs-site/hugo.toml`; `docs-site/static/llms.txt`; `docs-site/README.md`; `docs-site/scripts/**`; `scripts/convert-nextra-to-hextra/**` (one Go file, deleted if Q4 records deletion); `.github/workflows/docs-i18n-check.yml` and any new workflow file for the Hugo and link-check jobs; `assets/images/**`; `.claude/skills/hns-oss-docs-*/SKILL.md`; `.moai/docs/docs-site-i18n-rules.md`. No path under `internal/` or `cmd/` is in scope.

Under revision 0.3.0 this index changes none of the paths listed above. Each path is owned by the child SPEC that changes it (see § Child SPECs).

## Requirements (GEARS)

Revision 0.3.0 (route 2, operator decision `d-20261014T114834Z-e039`) makes REQ-PRC-001 to REQ-PRC-014 pointers. The GEARS text of each requirement is carried by its owning child SPEC under the same REQ number, and the child's spec.md governs it. This index keeps no copy of that text. The pre-revision wording is at commit bdb7b764 (`.moai/reports/t1655/index-mapping.md` §4 names the paths). REQ-PRC-015 stays in this index. The FLAG child has no requirement here; it writes its own REQ.

### Pointer entries (text carried by the owning child)

#### Stale command references (card item 1)

- **REQ-PRC-001** (pointer; Ubiquitous in the children): The four README files and every page under `docs-site/content/` cite only `moai` verbs and `/moai` slash commands that resolve on the tree. Owners, each for its own files: README (SPEC-V3R6-DOCS-PREREWRITE-README-001), REQ-PRC-001 README part, for the four README files; SWEEPA (SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001), SWEEPB (SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001), SWEEPC (SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001), SWEEPD (SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001), SWEEPE (SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001), and SWEEPF (SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001), REQ-PRC-001 page part, for their page files. The flag clause is not part of this requirement; FLAG (SPEC-V3R6-DOCS-PREREWRITE-FLAG-001) writes its own. The locale scope (Q17), the example-marker exemption (Q24), and the Q14 marker option are applied by the owning children in their own text.
- **REQ-PRC-002** (pointer; Event-driven in the child): owner GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001), REQ-PRC-002.

#### Version badge (card item 2)

- **REQ-PRC-003** (pointer; conditional on Q8 in the child): owner README (SPEC-V3R6-DOCS-PREREWRITE-README-001), REQ-PRC-003.

#### CG retirement passage (card item 3)

- **REQ-PRC-004** (pointer; conditional on Q1 in the child): owner README (SPEC-V3R6-DOCS-PREREWRITE-README-001), REQ-PRC-004.

#### Assets (card item 4)

- **REQ-PRC-005** (pointer; conditional on Q2 in the child): owner RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), REQ-PRC-005.
- **REQ-PRC-006** (pointer; conditional on Q3 in the child): owner RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), REQ-PRC-006.

#### Nextra residue, old README, llms.txt (card item 5)

- **REQ-PRC-007** (pointer; conditional on Q4 in the child): owner RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), REQ-PRC-007.
- **REQ-PRC-008** (pointer): owner RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), REQ-PRC-008.
- **REQ-PRC-009** (pointer; conditional on Q7 in the child): owner RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001), REQ-PRC-009. Q7 is decided inside RESID, and its cross-SPEC effect is recorded under Dependencies.

#### Redirects (card item 6)

- **REQ-PRC-010** (pointer): owner REDIR (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001), REQ-PRC-010.
- **REQ-PRC-011** (pointer): owner REDIR (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001), REQ-PRC-011, which covers the moai-mx redirect pair under Q23 (leader ruling B04, `d-20261014T060405Z-0fcd`).

#### CI gates (card item 7)

- **REQ-PRC-012** (pointer; conditional on Q9 in the child once the owning child applies the B11 fix): owner GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001), REQ-PRC-012.
- **REQ-PRC-013** (pointer; conditional on Q10 in the child): owner GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001), REQ-PRC-013.
- **REQ-PRC-014** (pointer): owner GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001), REQ-PRC-014.

### Operator-held requirement (not moved; closure is operator-held)

- **REQ-PRC-015** (Ubiquitous): The three jobs named in REQ-PRC-012 through REQ-PRC-014 **shall** be listed as required status checks on the `main` branch protection. This requirement is enforced outside the repository, and its closure is an operator-held action (see Out of Scope and decision Q22).
  - Owner: parent (this index), operator-held (Q12). Q22 has owner GATE (leader-confirmed). Its acceptance criterion AC-PRC-015 stays in acceptance.md as a regression-guard.

## Child SPECs

The index's file-level work is split into eleven child SPECs by decision Q20 = (c) (`d-20261014T055241Z-1bc6`). Each child is tier M, with 5 to 14 files, under the 15-file cap. The cards are already issued under the card ids below, and the child IDs are proposed; the leader confirms the final IDs. The children land independently, in any order. The closing run (M6, plan.md §B) is the one step that depends on another child: it needs the GATE child's citation checker to print `cited-verbs: 0` with exit 0.

| TAG | Proposed child ID | Card | REQ | AC | Decision rows | Files | After-relations |
|-----|-------------------|------|-----|----|---------------|-------|-----------------|
| GATE | SPEC-V3R6-DOCS-PREREWRITE-GATE-001 | t1717 | REQ-PRC-002, REQ-PRC-012, REQ-PRC-013, REQ-PRC-014 | AC-PRC-002, AC-PRC-012, AC-PRC-013, AC-PRC-014 | Q9, Q10, Q11 (own); Q5 and Q18 (inputs); Q22 with owner GATE (leader-confirmed) | 9 (10 if the link-check tool needs extra configuration) | None; no file overlap. Its citation checker is the M6 input. Before it starts AC-PRC-012, the child corrects that criterion's RED cell (audit defect B01). |
| REDIR | SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 | t1718 | REQ-PRC-010, REQ-PRC-011 | AC-PRC-010, AC-PRC-011 (AC-PRC-011 extended to the moai-mx pair under Q23) | Q15, Q16, Q21, Q23 | 5 | None. |
| RESID | SPEC-V3R6-DOCS-PREREWRITE-RESID-001 | t1719 | REQ-PRC-005, REQ-PRC-006, REQ-PRC-007, REQ-PRC-008, REQ-PRC-009 | AC-PRC-005, AC-PRC-006, AC-PRC-007, AC-PRC-008, AC-PRC-009 | Q2, Q3, Q4, Q6, Q7, Q19 | 14 | None. Before any deletion, Q2's basis is corrected for the tracked historical report (audit B05, E-09b2). It splits into 3a (Q2, Q4, Q6, Q19) and 3b (Q3, Q7) if Q3 or Q7 is still open. |
| README | SPEC-V3R6-DOCS-PREREWRITE-README-001 | t1720 | REQ-PRC-001 (README part), REQ-PRC-003, REQ-PRC-004 | AC-PRC-001 (README part), AC-PRC-003, AC-PRC-004 | Q1, Q8, Q13, Q17 | 13 | Recommended first. It shares ko files with SWEEPB, SWEEPC, SWEEPD, and FLAG, on different lines. |
| SWEEPA | SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001 | t1721 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q17 | 12 | Shares factory-mode.md with FLAG, on different lines. |
| SWEEPB | SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001 | t1722 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q17 | 12 | Recommended after README (shared ko files, different lines). Not required. |
| SWEEPC | SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001 | t1723 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q17 | 12 | Recommended after README (shared ko files, different lines). Not required. |
| SWEEPD | SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001 | t1724 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q17 | 12 | Recommended after README (shared ko prompt-caching file, different lines). Not required. |
| SWEEPE | SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001 | t1725 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q17, Q24 | 10 | None. Under Q24, its two narrative lines are checked for an example marker when it starts. |
| FLAG | SPEC-V3R6-DOCS-PREREWRITE-FLAG-001 | t1726 | new REQ, written by the child | new AC, written by the child | Q18 (GATE takes it as an input) | 8 (the four README files are shared with README) | Shares factory-mode.md with SWEEPA and the README files with README, on different lines. The help-text option of Q18 is outside this SPEC's scope. |
| SWEEPF | SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 | t1727 | REQ-PRC-001 (page part) | AC-PRC-001 (page part) | Q13, Q14, Q17 | 12 | None. The structure curator is the single author of `_meta.yaml`, so the child confirms who edits that file before it starts. |

## Out of Scope

The items below are out of scope for this SPEC. Each one is named so that a later change cannot claim it silently. Child-level scope lives in the child SPECs listed in § Child SPECs, not in this index.

### Out of Scope — docs rewrite content

- The 300-line README rewrite, the eight-section docs-site information architecture, and the generated reference pages belong to plan stages 5-1 through 5-3.
- New prose for pages that are not touched by a REQ above.

### Out of Scope — CLI and Go code

- Any change under `internal/` or `cmd/`, including restoring the retired `moai cg` verb or adding the absent verbs and sub-verbs.
- The cited-verbs producer, when written, lives under `scripts/` or as a workflow step. It shall not add Go packages under `internal/`.

### Out of Scope — branch protection and external settings

- Changing the GitHub branch protection of `main`. REQ-PRC-015 is recorded as a gap with an observed external state. The change itself is an operator-held action on a shared external system. Decision Q22 records the conflict between this exclusion and REQ-PRC-015 and is open.

### Out of Scope — version and release

- Changing the project version, tagging a release, publishing a GitHub release, and any Vercel deployment setting.

### Out of Scope — locale expansion

- Translating new prose into en, ja, or zh. Only the mechanical edits that REQ-PRC-001 to REQ-PRC-011 require are in scope in those locales, and their locale scope is an open decision (decision Q17 in decision-index.md).

## Dependencies and Parallel Work

- No `depends_on`. The card runs in parallel with stage 0 (3.2-0-1 to 3.2-0-5). The overlap of the path set above with the stage-0 paths named in the upgrade plan (`internal/` permissions, test isolation, MCP, and the update path) is not yet checked. The check runs at dispatch, per plan.md §D.
- The CG retirement SPEC, `SPEC-MOAI-CG-RETIRE-001`, is still `status: draft` and targets `v3.3.0`. It is not an authority for this SPEC, and this SPEC does not depend on it.
- The eleven child SPECs land independently, in any order, and each carries its own plan audit (decision `d-20261014T114834Z-e039`). The closing run is this index's M6 (plan.md §B.2), and it runs after every child has landed.
- Completed SPEC `SPEC-V3R6-DOCS-V3-REBUILD-001` (status `completed`) carries AC-DVR-019b (its acceptance.md line 97 at commit bdb7b764): `grep -l 'adk.mo.ai.kr' docs-site/static/robots.txt docs-site/static/llms.txt` must confirm that both files carry the canonical domain, and `docs-site/public/sitemap.xml` must exist after build. Decision Q7 (decided inside RESID) therefore has a cross-SPEC effect. Under the removal branch, the check cannot pass for `llms.txt`, so that SPEC needs an amendment (`amendment_of:` and a HISTORY `## Amendments` row, per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix) that re-scopes the check to `robots.txt`. Under the regenerate branch, the regenerated file must keep the canonical domain, and no amendment is needed. RESID's Q7 row names this path; this index does not decide Q7.
