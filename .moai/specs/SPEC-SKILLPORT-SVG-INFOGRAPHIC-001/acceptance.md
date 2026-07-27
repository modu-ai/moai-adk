# Acceptance Criteria — SPEC-SKILLPORT-SVG-INFOGRAPHIC-001

## §D. Acceptance Criteria Matrix

| AC ID | REQ | Criterion | Verification |
|-------|-----|-----------|--------------|
| AC-SVG-001 | REQ-SVG-001 | Skill exists template-first + local-synced | `test -f internal/template/templates/.claude/skills/moai-domain-svg-infographic/SKILL.md` AND local copy present |
| AC-SVG-002 | REQ-SVG-002 | Level-1 listing: scope named + combined char count measured | BOTH halves. (a) SCOPE: the `description` + `when_to_use` text names the SVG-infographic / static-diagram trigger scope (reviewer reads and confirms the trigger is unambiguous). (b) MEASURED CAP: extract the `description` and `when_to_use` field values from the template SKILL.md frontmatter, concatenate, and measure — combined length **≤ 1,536 characters** (the platform `maxSkillDescriptionChars` ceiling, which applies to the two fields TOGETHER). Report the measured number; an estimate FAILS this AC. Design target for reference: ~100 tokens ≈ 400 characters combined. |
| AC-SVG-003 | REQ-SVG-003 | Progressive disclosure structure | SKILL.md body within the Level-2 budget of **≤ ~5K tokens** (measured, not estimated); `references/` + `scripts/` present with the heavy archetype / geometry / icon-set / sketch content in Level-3 |
| AC-SVG-004 | REQ-SVG-004 | Layout-before-code enforced | body requires numeric layout + containment before SVG authoring |
| AC-SVG-005 | REQ-SVG-005 | Geometry-derived centers | body derives centers from card geometry, not per-language offsets |
| AC-SVG-006 | REQ-SVG-006 | CJK 60% budget | body states CJK-first font stack + Korean ~60% line budget |
| AC-SVG-007 | REQ-SVG-007 | Render path documented | body cites `scripts/render.mjs`, 2× PNG, disclosed browser + IHDR verify |
| AC-SVG-008 | REQ-SVG-008 | Lint path documented | body cites `scripts/check-svg.mjs`, hard errors vs warnings |
| AC-SVG-009 | REQ-SVG-009 | No-Node degradation | body: manual checklist substitutes, no machine-lint label |
| AC-SVG-010 | REQ-SVG-010 | No-Chromium degradation | body: SVG-only delivery + explicit limitation |
| AC-SVG-011 | REQ-SVG-011 | Prerequisite clarity | body states Node/Chromium needed only for lint+render, not authoring |
| AC-SVG-012 | REQ-SVG-012 | Mermaid selection rules present | body enumerates the mermaid-vs-svg selection rules |
| AC-SVG-013 | REQ-SVG-013 | No mermaid replacement | body states additive/complementary; mermaid pipeline untouched |
| AC-SVG-014 | REQ-SVG-014 | English body + scripts (CJK-capability carve-out) | No non-English PROSE or CODE COMMENTS in SKILL.md / references / scripts. **Carve-out (identical scope split to REQ-SVG-015's):** CJK font-family names and Korean example labels are diagram-rendering CAPABILITY, not prose — REQ-SVG-006 REQUIRES the CJK-first font stack and the Korean ~60% line budget, AC-SVG-015 requires that content to be present, and §D.2 references Korean sample labels. Their presence MUST NOT fail this AC. Verification is therefore NOT a CJK-codepoint scan (which would false-FAIL content another AC mandates): the reviewer reads the prose and comment lines and confirms they are English, while font-stack names and sample labels remain intact. |
| AC-SVG-015 | REQ-SVG-015 | Template neutrality (CI clean, CJK retained) | neutrality guards pass; CJK font/budget content present |
| AC-SVG-016 | REQ-SVG-016 | Clean-room — MECHANICAL half (absence invariant + house-convention license presence) | ALL THREE clauses (a)-(c) must return the stated result; clause (c) carries two sub-checks (c1, c2), both required. (a) vendor-token absence across the WHOLE skill tree including `scripts/` and `references/`: `grep -ril 'skillstead' internal/template/templates/.claude/skills/moai-domain-svg-infographic/ .claude/skills/moai-domain-svg-infographic/` → **0 matches**; (b) no NOTICE at EITHER surface REQ-SVG-016 names — the two skill directories AND the repository root: `test ! -e internal/template/templates/.claude/skills/moai-domain-svg-infographic/NOTICE && test ! -e .claude/skills/moai-domain-svg-infographic/NOTICE && test ! -e NOTICE` → exit 0 (the repo root carries no NOTICE today — verified at plan time — so this check starts green and guards against one being introduced); (c) **house-convention license PRESENT, no skillstead-tied value** — BOTH sub-checks: (c1) `grep -c '^license: Apache-2.0$' internal/template/templates/.claude/skills/moai-domain-svg-infographic/SKILL.md` → **exactly 1**; (c2) the frontmatter carries no skillstead reference — `awk '/^---$/{c++;next} c==1 && tolower($0) ~ /skillstead/ {n++} END{print n+0}' internal/template/templates/.claude/skills/moai-domain-svg-infographic/SKILL.md` → prints **0** (a clause-local, narrower restatement of (a); written pipe-free so the command survives verbatim copy out of this table cell). Rationale: `license: Apache-2.0` is moai's OWN license declaration per house convention — 18 of the 28 existing template skills carry exactly this field, including the sibling `moai-domain-html-report`, and the repository LICENSE is Apache-2.0 — so it is NOT skillstead attribution. A bare `grep -n '^license:' → 0 matches` assertion is **PROHIBITED** here: it contradicts REQ-SVG-016's affirmative clause and would false-FAIL a correct implementation. See the absence-invariant caveat below the matrix. |
| AC-SVG-017 | REQ-SVG-001 | Embedded FS rebuilt | `make build` succeeds; skill + scripts embedded |
| AC-SVG-018 | REQ-SVG-017 | Frontmatter conventions — full clause coverage | ALL FOUR clauses of REQ-SVG-017 verified in the template SKILL.md frontmatter: (a) `allowed-tools` is a comma-separated STRING (not a YAML array) listing exactly Read, Write, Edit, Grep, Glob, Bash; (b) if a `skills:` key is present it is a YAML ARRAY (`- item` block or `[a, b]` flow form), never a CSV string — if the key is absent, this clause is vacuously satisfied and MUST be reported as "absent, N/A" rather than silently passed; (c) every value under `metadata:` is a QUOTED string; (d) the `allowed-tools` value lists neither AskUserQuestion nor Agent, verified by an EXACT-LINE match scoped to that one field — `awk '/^---$/{c++;next} c==1 && /^allowed-tools:/' internal/template/templates/.claude/skills/moai-domain-svg-infographic/SKILL.md` → prints exactly `allowed-tools: Read, Write, Edit, Grep, Glob, Bash` and nothing else. An exact match on the mandated tool list subsumes the negative check (a line equal to that string cannot contain either token) while adding the (a) tool-set assertion for free. **A whole-frontmatter `grep -c 'AskUserQuestion\|Agent'` → 0 is PROHIBITED as the verification here**: it is out of scope for REQ-SVG-017 (which binds the `allowed-tools` value only) and it MEASURED 10 of the 28 existing template skills as false positives, matching prose in `description` / `when_to_use` such as `Agent(general-purpose)`. The unresolved frontmatter-block path placeholder that previously stood in this clause is likewise replaced by the concrete file path in the command above. |
| AC-SVG-019 | REQ-SVG-016 | Clean-room — JUDGMENT half (process attestation) | The implementing agent records a clean-room PROCESS attestation in `progress.md` §E.2 stating that the SKILL.md body, the `references/` files, AND the bundled `scripts/*.mjs` were drafted **from the functional-capability description in this SPEC only**, and that skillstead source text was NOT consulted while drafting. A reviewer separately confirms the shipped structure follows moai skill conventions (`.claude/rules/moai/development/skill-authoring.md`), including that the `scripts/` + `references/` layout derives from that rule rather than from prior art (see spec.md §A.1). This is a PROCESS attestation, NOT a textual-diff check: this SPEC deliberately records no skillstead source text, so a "no verbatim/near-verbatim wording" comparison has no reference corpus and is not executable. |

### Absence-invariant caveat (binds AC-SVG-016)

A passing `grep -ril 'skillstead' … → 0 matches` is an **absence-invariant guard only**. It proves that no attribution token leaked; it is **NOT** evidence that the skill was implemented, nor that the scripts work — an empty directory also returns 0 matches. Positive proof of implementation lives in the functional ACs (AC-SVG-001, AC-SVG-003 through AC-SVG-013) and must be evaluated independently.

Why this guard is needed at all: the existing template-neutrality CI guard does **not** cover it. `internal/template/internal_content_leak_test.go` `leakClasses` enforces internal SPEC-ID prefixes, REQ/AC token prefixes, audit citations, internal dates, and memory/archive paths — the vendor token `skillstead` is in **none** of those classes. A leaked attribution footer or a vendor reference in a script comment would pass every existing repo check, which is precisely why this AC carries its own explicit grep across the whole skill tree.

## §D.1 Given-When-Then scenarios

### Scenario 1 — layout-before-code (core path, runtime present)
- **Given** a request for a 4-box architecture flow infographic in Korean, with Node 18+ and headless Chromium available,
- **When** the skill runs its workflow,
- **Then** it computes the grid arithmetic and verifies containment before writing SVG, edits Korean copy to the ~60% budget, lints via `check-svg.mjs` (no hard errors), and renders a 2× PNG via `render.mjs` with disclosed browser version and IHDR-verified dimensions.

### Scenario 2 — mermaid selection boundary
- **Given** a request "add a flowchart to the README that we'll update every release",
- **When** the skill's selection rules are consulted,
- **Then** the skill routes the request to mermaid (markdown-embedded, frequently-changing, 4-locale-synced) rather than producing an SVG, demonstrating the anti-dual-maintenance rule.

### Scenario 3 — graceful degradation (no Chromium)
- **Given** a diagram request in an environment with Node 18+ but no headless Chromium,
- **When** the skill reaches the render step,
- **Then** it lints the SVG, delivers the editable SVG only, and states the render limitation explicitly (no fabricated PNG, no crash).

### Scenario 4 — graceful degradation (no Node)
- **Given** an environment without Node 18+,
- **When** the skill authors a diagram,
- **Then** it produces the editable SVG using the manual source checklist, applies no machine-lint label, and states the limitation.

## §D.2 Edge cases

- Text overflow flagged by `check-svg.mjs` as a low-confidence warning: the skill surfaces it as verify-in-PNG guidance, not a hard failure.
- Sketch preset requested: the opt-in `references/sketch.md` preset layers a hand-drawn surface over the SAME computed layout (layout-before-code still applies).
- Very long Korean labels: copy is edited to fit the 60% budget BEFORE SVG authoring, not truncated after.

## §D.3 Definition of Done

- [ ] All 19 AC rows pass.
- [ ] Bundled scripts run under Node 18+ with stdlib only (no npm install).
- [ ] `make build` succeeds; skill + `scripts/` + `references/` embedded and discoverable.
- [ ] Template-neutrality CI guard clean (internal tokens absent; CJK content retained).
- [ ] Selection-rules section present and unambiguous; mermaid pipeline unchanged.
- [ ] Clean-room MECHANICAL half green (AC-SVG-016): vendor-token grep across the whole skill tree → 0; no `NOTICE` in either skill directory nor at the repository root; house-convention `license: Apache-2.0` present exactly once; no skillstead reference anywhere in the frontmatter.
- [ ] `allowed-tools` exact-line match green (AC-SVG-018 d): `allowed-tools: Read, Write, Edit, Grep, Glob, Bash` — neither AskUserQuestion nor Agent listed.
- [ ] Clean-room JUDGMENT half green (AC-SVG-019): process attestation recorded in progress.md §E.2, covering SKILL.md + references + scripts.
- [ ] Level-1 combined `description` + `when_to_use` character count MEASURED and ≤ 1,536 (AC-SVG-002).

## §D.4 Quality gate

- Progressive-disclosure budget respected. Level-1: the COMBINED `description` + `when_to_use` text stays within the 1,536-character platform ceiling (hard), targeting the ~100-token metadata budget (≈ 400 characters, design target). Level-2: SKILL.md body ≤ ~5K tokens. Level-3: heavy geometry / icon-set / sketch content offloaded to `references/`.
- Level-2 sizing advisory: the sibling `moai-domain-html-report` SKILL.md is ~22.7 KB ≈ 5.7K tokens — already ABOVE the ~5K Level-2 target — so the ≤ ~5K budget is aggressive for a skill of comparable scope. Plan the Level-3 split up front (see plan.md §F M4); do not discover the overflow after the body is drafted.
- No dual-maintenance surface introduced against the mermaid pipeline.
- Scripts template-neutral, English-commented, and clean-room authored.
