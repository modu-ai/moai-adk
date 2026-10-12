# SPEC-UPDATE-REPAIR-RC30-001 — Decision index

Stateless on the status axis: the lifecycle lives in spec.md. Each row follows Detect → Explain → Ask. No row carries a recommendation. One row carries an applied default: Q7 (60 columns, kept by planning-session board ruling d-20261010T101935Z-4bc4). No other row carries an applied default. Authority anchors are committed artifacts; an anchor that cannot be verified routes the row to FOUNDER.

### Q1: Which set does R1's "16" name, and which set does R2(g)'s "four locale strings" name?

Status: RESOLVED from disk (POLICY-COVERED). No operator question.
Authority anchor: `.claude/rules/moai/development/coding-standards.md` § Programming-Language Neutrality Contract (16 languages), lines 23, 25, and 35-37. Verified by grep in this run; verbatim output in research.md § 2 E-LOCALE.
Resolution:
- R1 "16" names the 16 supported programming languages (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift). It is a neutrality concern (coding-standards.md lines 25, 35-36).
- R2(g) "four" names the four user-facing locales ko, en, ja, zh. It is a translation concern (coding-standards.md lines 36-37).
- The two sets **differ**. They are disjoint (no member of the 16 is ko, en, ja, or zh) and govern different axes.
- Fact: `.moai/config/sections/language.yaml` line 5 lists seven conversation-language options (ko, en, ja, zh, es, fr, de). REQ-RC30-019 covers the four locales; es, fr, and de are excluded (spec.md § I).
Classification note: the anchor is a policy-layer rule file, not a constitution clause. The row is POLICY-COVERED because the rule text names both sets and their axes.
Operator verdict:

### Q2: Which mechanism reaches the shared assets for a profile session (CLAUDE_CONFIG_DIR)? — DECIDED by operator decision d-20261010T060848Z-dc3b

Label: FOUNDER
Class: product-level
Status: DECIDED by operator decision d-20261010T060848Z-dc3b: option (c) with two conditions, and option (b) is not adopted. Condition 1 is a provenance mark on the kept project copy (the shipped hash or an equivalent marker) so the 3.2-2-5 doctor can identify and reclaim it; condition 2 is a card-record statement that the structural fix moves to 3.2-2-5 (t1642) and 3.2-2-6 (t1643).
Residual Gap: the mo.ai.kr profile was not measured. The measurement was on the moai-adk profile lane session; its report is untracked, is cited by path only, and is not an authority anchor: .moai/reports/t1619/profile-skill-load-probe.md.
Authority anchor: SPEC-USER-ASSET-INSTALL-001 spec.md lines 237-238 (P6) and line 255 (REQ-002), re-read on main in this run. The rest of REQ-002, after line 255, is unread.
Why unresolved: The options change what `moai init` and `moai update` write, or what a profile session reads. The choice belongs to the operator.
Options (factual; no recommendation):
- **(a)** A reach mechanism that writes nothing into a profile directory. None is established by evidence in this card. Audit §3 records none. No evidenced option is stated.
- **(b)** A link or copy of the common assets inside the profile directory. It requires amending REQ-002 (spec.md line 255) to permit a link or copy in a profile directory. P6 (spec.md lines 237-238) would be restated to match. Amendment scope: the line-255 clause only. Follow-on: if (b) is chosen, manager-spec performs the amendment in place in this card after WT-user-asset-bundle (t1591) is absorbed (done; spec.md § G.2). The leader coordinates that with the other lane.
- **(c)** Keep project-side copies when a profile session is detected. Consequence against D1: D1 names the shared roots as the copy target for `moai init` and `moai update`. Under (c), profile-detected projects hold a second copy of the common skills and agents inside the project tree, outside the shared-root model. That copy is refreshed only if U-12 is implemented, so the two copies can diverge. Revision 0.1.4 replaces the "detected" predicate with the reach judgment of AC-RC30-28 (`moai doctor` after `moai update`).
Default: none (no published rule ranks these options).
Gated U-01 in SPEC A before this decision: AC-RC30-01 and AC-RC30-02 moved to card t1668 (d-20261009T235135Z-5e35) and remain there. U-01 proceeds under option (c), judged by AC-RC30-27 (keep) and AC-RC30-28 (reach).
Operator verdict: "(c) + 두 조건" (verbatim operator selection), board record d-20261010T060848Z-dc3b.

### Q3: Queue overlaps of t1527, t1594, and t1612 with this SPEC

Label: FOUNDER
Class: implementation-level
Status: PENDING. Overlap facts are recorded in spec.md § F with the disposition undecided. No Default is recorded.
Authority anchor: SPEC-UPDATE-MIGRATION-001 § H (spec.md lines 228-230) names t1527 as owner of banners, progress rendering, and severity prefixes. The card texts for t1527, t1594, and t1612 were not read in this run.
Operator verdict: "(i) record overlaps only, SPEC A unchanged (overlap owners t1527/t1612->t1654, t1594->t1630; all three absorbed-closed in queue)" (operator verdict 2026-10-11 (AskUserQuestion, 영실이 session)), board record d-20261010T223148Z-1e76.

### Q4: Are common skills and agents installed into the shared roots by moai init and moai update?

Label: DECIDED
Authority anchor: SPEC-USER-ASSET-INSTALL-001 (completed, on main), title line: "Install common skills and agents into per-user folders (no plugin carrier) …". Verified in this run.
Why unresolved: Not unresolved. Recorded because D1 asks for it. The operator direction D1 in the lane addendum confirms the model.
Operator verdict:

### Q5: Which version banner does moai update print?

Label: DECIDED
Authority anchor: SPEC-CLI-TUX-INIT-UPDATE-001 REQ-TUXIU-015 (spec.md, Group B): the update command renders the identity band.
Why unresolved: Not unresolved. The band is the banner, so the duplicate "Current version" line at `internal/cli/update.go:200` is the line to remove (U-17).
Operator verdict:

### Q6: Are project-level .claude/skills and .claude/agents visible to a session launched with CLAUDE_CONFIG_DIR set to a profile directory?

Label: EVIDENCE-NEEDED (partly measured; not decided)
Authority anchor: none. Audit §3 covers user-level skills only.
Measured so far (evidence, not an authority anchor): project skills 3 of 3 visible on the moai-adk profile, recorded in decision d-20261010T060848Z-dc3b. Project agents and the mo.ai.kr profile were not measured. Q6 is not decided beyond the skills measurement.
Why unresolved: The removal gate depends on the answer. The runnable measurement AC-RC30-01 is moved to card t1668 and is not yet run; the reach judgment for the profile state is AC-RC30-28, whose doctor output is Gap G-28.
Operator verdict:

### Q7: What width threshold triggers the plain-line fallback for narrow terminals? — APPLIED DEFAULT: 60 columns

Label: FOUNDER
Class: implementation-level
Status: APPLIED DEFAULT. This is the one Q7 status, and spec.md § J and REQ-RC30-018 state the same value: 60 columns, kept by planning-session board ruling d-20261010T101935Z-4bc4. The operator may override it (option (i) or (iii)); until an override is recorded, the width threshold is 60 columns.
Authority anchor: none in the repository. R2f asks for the fallback, not a number. The board ruling is cited as stated; its verbatim line is not pinned in this revision.
Why unresolved: R2f names narrow terminals without a width. The choice is the board's, not a repository rule.
Options (factual; no recommendation):
- **(i)** 40 columns: plain lines are used on terminals narrower than 40 columns.
- **(ii)** 60 columns: plain lines are used on terminals narrower than 60 columns.
- **(iii)** Another width, stated by the operator or leader.
Default: 60 columns (applied; board ruling d-20261010T101935Z-4bc4). Note: the published Default rule (the option with the smaller user-visible surface) would rank 40 columns; the 60-column default is a board ruling, not an application of that rule. Alternate: 40 columns.
Operator verdict: "(ii) keep 60 columns" (operator verdict 2026-10-11 (AskUserQuestion, 영실이 session)), board record d-20261010T223148Z-1e76.

### Q8: How do the removal items U-02 and U-10 proceed, given that AC-RC30-01 needs a profile session?

Label: FOUNDER
Class: product-level
Status: PENDING. AC-RC30-01 cannot be run in this card. The removal switch stays off in this card, and U-02 and U-10 are moved to card t1668 (split, d-20261009T235135Z-5e35); Q8 gates card t1668 only. The leader's split that moved U-02 and U-10 out of SPEC A (d-20261009T235135Z-5e35) is the scope the operator approved for audit (d-20261010T000832Z-efde); the operator verdict on Q8 itself is not recorded, and its line stays blank.
Authority anchor: acceptance.md, AC-RC30-01 (judging command requires `CLAUDE_CONFIG_DIR` and a profile session); spec.md § D.2 gate sentence.
Options (factual; no recommendation):
- **(i)** An operator-side run of AC-RC30-01 on a profile session, recorded in progress.md, which adopts the criterion. The switch is then turned on and U-02 and U-10 proceed (plan.md § B.3, step 4).
- **(ii)** The operator decides that U-02 and U-10 leave this SPEC or are re-scoped. The switch stays off.
Default: none.
Moved with the removal family to card t1668 (d-20261009T235135Z-5e35); no longer gates SPEC A. Status stays PENDING.
Operator verdict: "(ii) U-02/U-10 leave the SPEC family, removal switch stays off (re-scope at t1668 / 3.2-2-5)" (operator verdict 2026-10-11 (AskUserQuestion, 영실이 session)), board record d-20261010T223148Z-1e76.

### Q9: Which scope does AC-RC30-22 take against REQ-UYP-014 for comments on sequence items? — DECIDED by delegated judgment d-20261010T143542Z-5d19

Label: FOUNDER
Class: implementation-level
Status: DECIDED by delegated judgment d-20261010T143542Z-5d19, option (가): keep REQ-UYP-014, with the AC-RC30-22 scope wording that excludes comments on sequence items.
Record: scope card:t1619; kind ruling; decided_by: a delegated judgment under the operator's delegation (board record: "영실이 판단(운영자 위임)"); board read 2026-10-10T14:38:52Z.
Residual Gap: the lane's copy of the board record is git-ignored (`.moai/reports/t1619/progress.md` lines 4104-4110; `.gitignore` line 235). It is cited by path only, is not an authority anchor, and is not in committed history. The authority anchors below are committed artifacts.
Authority anchor: `.moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md` line 158 (REQ-UYP-014: a merged sequence node is replaced wholesale); `.moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/plan.md` lines 121-131 (Decision D4; the chosen policy is at line 125); `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/acceptance.md` line 235 (AC-RC30-22 expected cell); `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/spec.md` line 48 (REQ-RC30-007 exception).
Derivation: REQ-UYP-014 replaces sequences wholesale and does not name comments, so the comment exclusion is derived from that replacement. The exclusion is not a literal clause of REQ-UYP-014.
Why unresolved: AC-RC30-22 as written (all comment lines survive) and REQ-UYP-014 (sequences replaced wholesale) disagree for comments on sequence items. The question is which of the two gives way.
Options: the board record names the chosen option only: (가) keep REQ-UYP-014, with the AC-RC30-22 scope wording that excludes comments on sequence items. The other options are not recorded in the lane's copy, so none is listed here.
Default: none (decided; no default applied).
Applied to: the AC-RC30-22 expected cell (acceptance.md line 235) and the REQ-RC30-007 location note (spec.md line 48), both in this SPEC.
Operator verdict: "(가): keep REQ-UYP-014" (verbatim board selection; a delegated judgment under the operator's delegation), board record d-20261010T143542Z-5d19.

### B1: CLOSED by the leader's split (d-20261009T235135Z-5e35, option 1) — Tier L acceptance-criteria budget (22 of 25)

Label: FOUNDER
Class: implementation-level
Status: DECIDED by the leader: split (d-20261009T235135Z-5e35). SPEC A carries 22 criteria (21 at revision 0.1.3, plus AC-RC30-28 in revision 0.1.4) and 17 requirements. REQ-RC30-003, -004 and -021 are owned by card t1668; REQ-RC30-014, -016, -018, -019 and -020 are owned by card t1669. Options (a) and (b) below are kept as the record; their counts predate the moves and the 0.1.4 criterion.
Authority anchor: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (REQ/AC budget table: Tier L ceiling 25).
Options (factual; no recommendation):
- **(a) Split.** Criteria that would move to a sibling SPEC: AC-RC30-25 (TUX golden and summary, U-26 to U-29) and AC-RC30-26 (integrity probe, U-31, t1578 scope). The parent keeps 24 criteria, under the ceiling. Effect on coverage: REQ-RC30-014, -016, -018, -019, and -020 are mapped only to AC-RC30-25, and REQ-RC30-021 only to AC-RC30-26, so these six requirements have no criterion in the parent (REQ-RC30-017 keeps AC-RC30-09). For the six to stay covered, AC-RC30-25 and AC-RC30-26 must both stay in the parent. A split that keeps both and moves one other criterion reaches 25; moved alone, any of AC-RC30-01, -02, -03, -04, -06, -09, -10, -11, -13, -19, -22, -23, or -24 leaves every requirement covered. Sibling scope: the moved items keep their IDs and disposition.
- **(b) Merge.** Merge AC-RC30-14 (skip decision, U-13) and AC-RC30-15 (stamp skew, U-14) into one criterion with two named sub-cases. Coverage kept: U-13 and U-14 both remain judged, by two sub-cases of one test. The parent keeps 25 criteria.
Default: none.
Operator verdict: "(a) split, leader split 5e35 confirmed" (operator verdict 2026-10-11 (AskUserQuestion, 영실이 session)), board record d-20261010T223148Z-1e76.

### F1: Receiving cards of moved items (tracked record)

Status: tracked record of the placement of items moved out of SPEC A. Moved items keep their IDs. Two receiving cards: card t1668 (the removal family) and card t1669 (the follow-up batch). Both depend on card t1619. The ruling texts quoted in this SPEC use earlier labels for these two cards; the card IDs above are the names used from revision 0.1.4 on. Each row names the ruling or decision that moved it. The git-ignored draft of this list (.moai/reports/t1619/followup-cards-draft.md, excluded by .gitignore line 235) is not a citation target; this section is.

- Card t1669 (ruling d-20261010T040402Z-94b9): AC-RC30-25 with REQ-RC30-014, REQ-RC30-016, REQ-RC30-018, REQ-RC30-019, REQ-RC30-020 and U-26 through U-30. Dispositions before the move (spec.md § C): U-26 to U-29 IN-NOW; U-30 GAP-INVESTIGATE. AC-RC30-25 also judges the count clause of REQ-RC30-017 (its one conflict count equals the Reconciliation conflict count).
- Card t1668 (ruling d-20261009T235135Z-5e35): AC-RC30-01, AC-RC30-02, AC-RC30-03, AC-RC30-11 with REQ-RC30-003, REQ-RC30-004 and U-02, U-10. The removal half of AC-RC30-27 (TestMigrationRemovesSharedCopyWithoutProfileReach) moves with AC-RC30-01.
- Card t1668 (ruling d-20261010T033220Z-3c5f, decision d-20261010T084752Z-d4be): AC-RC30-26 with REQ-RC30-021 and U-31.
- Kept in SPEC A (added in revision 0.1.4): AC-RC30-28, the single reach judge of REQ-RC30-002 (`moai doctor` after `moai update`).
- Gaps (not closed by any placement above):
  - The REQ-RC30-017 count clause: its judge is AC-RC30-25 in card t1669, which SPEC A does not adopt. Until card t1669 adopts it, the clause has no judge in SPEC A (owner: card t1669).
  - The AC-RC30-24 cross-reference to REQ-RC30-003: REQ-RC30-003 is owned by card t1668 under d-20261009T235135Z-5e35; AC-RC30-24 stays in SPEC A as a cross-reference, criterion text unchanged.
  - The leader alternative of moving AC-RC30-24 with REQ-RC30-003, which would leave 21 criteria in SPEC A and is not applied.
