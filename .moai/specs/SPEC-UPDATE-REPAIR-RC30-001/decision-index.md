# SPEC-UPDATE-REPAIR-RC30-001 — Decision index

Stateless on the status axis: the lifecycle lives in spec.md. Each row follows Detect → Explain → Ask. No row carries a recommendation. Authority anchors are committed artifacts; any anchor that cannot be verified routes the row to FOUNDER.

### Q1: Which locale set do the R1 "16 locales" and the R2g "four locale strings" refer to?

Label: FOUNDER
Class: product-level
Authority anchor: none resolves the question. `.claude/rules/moai/development/coding-standards.md` § Programming-Language Neutrality Contract lists 16 programming languages, not locales.
Why unresolved: R1 says "behave the same across the 16 locales". R2g says "the four locale strings stay consistent". The repository's user-facing locales are ko, en, ja, zh (`language.yaml` options list also names es, fr, de).
Default: none (no published rule ranks these options; both change scope).
Operator verdict:

### Q2: Does the profile-reach mechanism install into a profile directory, which REQ-002 forbids, or does the project-side copy stay as the permanent mechanism?

Label: FOUNDER
Class: product-level
Authority anchor: SPEC-USER-ASSET-INSTALL-001, spec.md:237-238 (P6) and spec.md:255 (REQ-002), verified in this run on main.
Why unresolved: The lane addendum rejects "keep project copies" as the steady state and rejects "warning only". Any install into the profile directory conflicts with REQ-002. Amending REQ-002 is an operator decision, not a card decision.
Default: none (no published rule ranks these options).
Operator verdict:

### Q3: Which card owns the init-side presentation of t1527?

Label: FOUNDER
Class: implementation-level
Authority anchor: SPEC-UPDATE-MIGRATION-001 § H (spec.md:228-230) names t1527 as owner of banners, progress, and severity prefixes. SPEC-CLI-TUX-INIT-UPDATE-001 Group D decides the init banner.
Why unresolved: The t1527 card text was not read. Absorbing its init side would touch `internal/cli/init*.go`, which WT-user-asset-bundle also changes.
Default: defer (rule: the option that preserves current behavior; t1527 stays queued unchanged).
Alternate: absorb the init side into this SPEC.
Operator verdict: DEFAULT-APPLIED 2026-10-09T18:19:27Z manager-spec (lane-14, card t1619)

### Q4: Are common skills and agents installed into the shared roots by `moai init` and `moai update`?

Label: DECIDED
Authority anchor: SPEC-USER-ASSET-INSTALL-001 (completed, on main), title line of spec.md: "Install common skills and agents into per-user folders (no plugin carrier) …". Verified in this run.
Why unresolved: Not unresolved. Recorded because D1 asks for it. The operator direction D1 quoted in the lane addendum confirms the model.
Operator verdict:

### Q5: Which version banner does `moai update` print?

Label: DECIDED
Authority anchor: SPEC-CLI-TUX-INIT-UPDATE-001 REQ-TUXIU-015 (spec.md, Group B): the update command renders the identity band "◆ MoAI-ADK <version> <go-runtime> · claude".
Why unresolved: Not unresolved. Decides that the identity band is the banner, so the duplicate "Current version" line at `internal/cli/update.go:200` is the line to remove (U-17).
Operator verdict:

### Q6: Are project-level `.claude/skills` and `.claude/agents` visible to a session launched with `CLAUDE_CONFIG_DIR` set to a profile directory?

Label: EVIDENCE-NEEDED
Authority anchor: none. The audit's empirical check covers user-level skills only (audit §3).
Why unresolved: The removal gate depends on the answer. It needs a runnable measurement (AC-RC30-01), whose command form is not yet measured.
Operator verdict:

### Q7: What width threshold triggers the plain-line fallback for narrow terminals?

Label: FOUNDER
Class: implementation-level
Authority anchor: none in the repository. R2f asks for the fallback, not a number.
Why unresolved: R2f names narrow terminals without a width.
Default: 40 columns (rule: the option with the smaller user-visible surface; a 40-column threshold switches to plain lines on fewer terminals than 60 columns).
Alternate: 60 columns.
Operator verdict: DEFAULT-APPLIED 2026-10-09T18:19:27Z manager-spec (lane-14, card t1619)
