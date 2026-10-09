# SPEC-UPDATE-REPAIR-RC30-001 — Decision index

Stateless on the status axis: the lifecycle lives in spec.md. Each row follows Detect → Explain → Ask. No row carries a recommendation. Authority anchors are committed artifacts; any anchor that cannot be verified routes the row to FOUNDER.

### Q1: Which set does R1's "16" name, and which set does R2(g)'s "four locale strings" name?

Label: POLICY-COVERED
Authority anchor: `.claude/rules/moai/development/coding-standards.md` § Programming-Language Neutrality Contract (16 languages), lines 23, 25, and 35-37. Verified by grep in this run; verbatim output in research.md § 2 E-LOCALE.
Resolution from disk (no operator question):
- R1 "16": the 16 supported programming languages go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, and swift. This is a neutrality concern (coding-standards.md line 25 and lines 35-36).
- R2(g) "four": the four user-facing locales ko, en, ja, and zh. This is a translation concern (coding-standards.md lines 36-37).
- Match or differ: **differ**. The two sets are disjoint (no member of the 16 is ko, en, ja, or zh), and they govern different axes.
- Fact: `.moai/config/sections/language.yaml` line 5 lists seven conversation-language options (ko, en, ja, zh, es, fr, de). The rule's "four" is narrower than that option list.
- The R1 brief text reads "16 locales". On disk, the 16 is a count of programming languages, not of locales.
Classification note: the anchor is a policy-layer rule file, not a constitution clause. The row is POLICY-COVERED because the rule text itself names both sets and their axes, so no operator question remains.
Operator verdict:

### Q2: Which mechanism reaches the shared assets for a profile session (CLAUDE_CONFIG_DIR)?

Label: FOUNDER
Class: product-level
Authority anchor: SPEC-USER-ASSET-INSTALL-001 spec.md lines 237-238 (P6) and line 255 (REQ-002), re-read on main in this run. The rest of REQ-002, after line 255, is unread.
Why unresolved: The options below change what `moai init` and `moai update` write, or what a profile session reads. The choice belongs to the operator.
Options (factual; no recommendation):
- **(a)** A reach mechanism that writes nothing into a profile directory. None is established by evidence in this card. Audit §3 records none. Stated only as: no evidenced option.
- **(b)** A link or copy of the common assets inside the profile directory. It requires amending REQ-002 (spec.md line 255) to permit a link or copy in a profile directory. P6 (spec.md lines 237-238) would be restated to match. Amendment scope: the line-255 clause only. Follow-on: if (b) is chosen, manager-spec performs the amendment in place in this card after WT-user-asset-bundle (t1591) is absorbed at run start (M1). The leader coordinates that with the other lane.
- **(c)** Keep project-side copies when a profile session is detected. Consequence against D1: D1 names the shared roots as the copy target for `moai init` and `moai update`. Under (c), profile-detected projects hold a second copy of the common skills and agents inside the project tree, outside the shared-root model. That copy is refreshed only if U-12 is implemented, so the two copies can diverge.
Default: none (no published rule ranks these options).
Operator verdict:

### Q3: Queue overlaps of t1527, t1594, and t1612 with this SPEC

Label: FOUNDER
Class: implementation-level
Authority anchor: SPEC-UPDATE-MIGRATION-001 § H (spec.md lines 228-230) names t1527 as owner of banners, progress rendering, and severity prefixes. The card texts for t1527, t1594, and t1612 were not read in this run.
Why unresolved: Disposition undecided for each card. Recorded as overlap facts only (spec.md § F). No Default is recorded.
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
Why unresolved: The removal gate depends on the answer. It needs the runnable measurement AC-RC30-01, whose command form is not yet measured.
Operator verdict:

### Q7: What width threshold triggers the plain-line fallback for narrow terminals?

Label: FOUNDER
Class: implementation-level
Authority anchor: none in the repository. R2f asks for the fallback, not a number.
Why unresolved: R2f names narrow terminals without a width.
Default: 40 columns (rule: the option with the smaller user-visible surface).
Alternate: 60 columns.
Operator verdict: DEFAULT-APPLIED 2026-10-09T18:19:27Z manager-spec (lane-14, card t1619)
