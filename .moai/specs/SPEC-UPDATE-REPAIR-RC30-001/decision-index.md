# SPEC-UPDATE-REPAIR-RC30-001 — Decision index

Stateless on the status axis: the lifecycle lives in spec.md. Each row follows Detect → Explain → Ask. No row carries a recommendation. No row carries an applied default. Authority anchors are committed artifacts; an anchor that cannot be verified routes the row to FOUNDER.

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

### Q2: Which mechanism reaches the shared assets for a profile session (CLAUDE_CONFIG_DIR)? — PENDING, operator decision

Label: FOUNDER
Class: product-level
Status: PENDING. No option is selected, and no default is applied. REQ-RC30-002 reads "the mechanism selected under operator decision Q2 (pending)".
Authority anchor: SPEC-USER-ASSET-INSTALL-001 spec.md lines 237-238 (P6) and line 255 (REQ-002), re-read on main in this run. The rest of REQ-002, after line 255, is unread.
Why unresolved: The options change what `moai init` and `moai update` write, or what a profile session reads. The choice belongs to the operator.
Options (factual; no recommendation):
- **(a)** A reach mechanism that writes nothing into a profile directory. None is established by evidence in this card. Audit §3 records none. No evidenced option is stated.
- **(b)** A link or copy of the common assets inside the profile directory. It requires amending REQ-002 (spec.md line 255) to permit a link or copy in a profile directory. P6 (spec.md lines 237-238) would be restated to match. Amendment scope: the line-255 clause only. Follow-on: if (b) is chosen, manager-spec performs the amendment in place in this card after WT-user-asset-bundle (t1591) is absorbed at run start (M1). The leader coordinates that with the other lane.
- **(c)** Keep project-side copies when a profile session is detected. Consequence against D1: D1 names the shared roots as the copy target for `moai init` and `moai update`. Under (c), profile-detected projects hold a second copy of the common skills and agents inside the project tree, outside the shared-root model. That copy is refreshed only if U-12 is implemented, so the two copies can diverge.
Default: none (no published rule ranks these options).
Operator verdict:

### Q3: Queue overlaps of t1527, t1594, and t1612 with this SPEC

Label: FOUNDER
Class: implementation-level
Status: PENDING. Overlap facts are recorded in spec.md § F with the disposition undecided. No Default is recorded.
Authority anchor: SPEC-UPDATE-MIGRATION-001 § H (spec.md lines 228-230) names t1527 as owner of banners, progress rendering, and severity prefixes. The card texts for t1527, t1594, and t1612 were not read in this run.
Operator verdict:

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

Label: EVIDENCE-NEEDED
Authority anchor: none. Audit §3 covers user-level skills only.
Why unresolved: The removal gate depends on the answer. It needs the runnable measurement AC-RC30-01, which is not yet run (its RED cell is pending).
Operator verdict:

### Q7: What width threshold triggers the plain-line fallback for narrow terminals? — PENDING, leader or operator decision (no applied default)

Label: FOUNDER
Class: implementation-level
Status: PENDING. No default is applied. The width threshold stays unset until this row is answered. The question is raised through AskUserQuestion (askuser-protocol.md § Recommendation Placement Principles, condition 3: never a silently applied default).
Authority anchor: none in the repository. R2f asks for the fallback, not a number.
Why unresolved: R2f names narrow terminals without a width.
Options (factual; no recommendation):
- **(i)** 40 columns: plain lines are used on terminals narrower than 40 columns.
- **(ii)** 60 columns: plain lines are used on terminals narrower than 60 columns.
- **(iii)** Another width, stated by the operator or leader.
Default: none. The published rule is not applied here because this row is not closed by it (see Q7 status).
Operator verdict:

### B1: OPEN LEADER DECISION — Tier L acceptance-criteria budget (26 of 25)

Label: FOUNDER
Class: implementation-level
Status: OPEN LEADER DECISION. The SPEC keeps 26 criteria in the text until the leader decides. Exceeding the ceiling is a split signal under spec-workflow.md § SPEC Complexity Tier, and the budget is not relaxed silently.
Authority anchor: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (REQ/AC budget table: Tier L ceiling 25).
Options (factual; no recommendation):
- **(a) Split.** Criteria that would move to a sibling SPEC: AC-RC30-25 (TUX golden and summary, U-26 to U-29) and AC-RC30-26 (integrity probe, U-31, t1578 scope). The parent keeps 24 criteria, under the ceiling. Effect on coverage: REQ-RC30-014, -016, -018, -019, and -020 are mapped only to AC-RC30-25, and REQ-RC30-021 only to AC-RC30-26, so these six requirements have no criterion in the parent (REQ-RC30-017 keeps AC-RC30-09). For the six to stay covered, AC-RC30-25 and AC-RC30-26 must both stay in the parent. A split that keeps both and moves one other criterion reaches 25; moved alone, any of AC-RC30-01, -02, -03, -04, -06, -09, -10, -11, -13, -19, -22, -23, or -24 leaves every requirement covered. Sibling scope: the moved items keep their IDs and disposition.
- **(b) Merge.** Merge AC-RC30-14 (skip decision, U-13) and AC-RC30-15 (stamp skew, U-14) into one criterion with two named sub-cases. Coverage kept: U-13 and U-14 both remain judged, by two sub-cases of one test. The parent keeps 25 criteria.
Default: none.
Operator verdict:
