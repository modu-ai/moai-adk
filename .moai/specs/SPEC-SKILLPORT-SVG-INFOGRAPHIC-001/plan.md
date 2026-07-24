# Implementation Plan — SPEC-SKILLPORT-SVG-INFOGRAPHIC-001

## §A. Context

Author `moai-domain-svg-infographic` in the MoAI catalog, sibling to `moai-domain-html-report`, borrowing only the functional capability of skillstead's `svg-infographic` skill (an Apache-2.0 project) and authoring the skill clean-room in moai's own voice and structure. Part of the **SKILLPORT** Epic. This is the heaviest of the three SPECs: the skill bundles Node scripts (`render.mjs`, `check-svg.mjs`) and multiple Level-3 reference files, and carries a runtime gate (Node 18+ / headless Chromium) plus a CRITICAL positioning constraint against the existing mermaid pipeline.

Net-new skill authoring (no existing behavior to preserve). Run-phase work is clean-room skill + reference + script authoring, template→local sync, and `make build`.

Clean-room note: ideas/functionality are not copyrightable — only expression is. Because the skill copies none of skillstead's wording, structure, code, or file layout, it is not a derivative work, so Apache-2.0 attribution is neither required nor added. The functional inspiration is recorded here (an internal SPEC file) as design rationale only; the shipped template files carry zero skillstead reference.

## §B. Known Issues / Resolved Decisions

- **Bundled Node scripts (RESOLVED — stdlib-only, neutral, clean-room):** the bundled `scripts/*.mjs` are authored clean-room to be Node-18+-stdlib-only (no npm install, no bundled binary), with English comments and zero internal tokens (no SPEC IDs, REQ/AC tokens, audit citations, internal dates, or commit SHAs). Executable files are the highest neutrality-risk surface, so run-phase MUST grep the scripts for internal tokens before commit. This is the committed default; no clarification is outstanding.
- **Attribution model (RESOLVED — clean-room, no attribution):** the shipped skill files (SKILL.md + references + scripts) carry NO skillstead attribution — no footer provenance line, no `NOTICE` file at either surface (skill directories or repository root), no skillstead-TIED `license:` value. No NOTICE is added because there is no attribution to consolidate. No open questions remain.
- **`license:` frontmatter field (RESOLVED — house convention, ships Apache-2.0):** the skill DOES ship `license: Apache-2.0`. This is moai's OWN license declaration, not attribution: 18 of the 28 existing template skills already carry exactly this field (including the sibling `moai-domain-html-report`), and the repository LICENSE is Apache-2.0. Run-phase MUST include the field; omitting it violates REQ-SVG-016. Do NOT author a bare "no `license:` field" absence check — AC-SVG-016(c) asserts `grep -c '^license: Apache-2.0$'` → exactly 1.

## §C. Pre-flight

- Confirm no skill named `moai-domain-svg-infographic` exists (verified at plan time — absent).
- Confirm `moai-domain-html-report` exists as the sibling + mermaid-pipeline reference point (verified — present).
- Confirm the `moai-domain-*` namespace is template-distributed (verified via skill-authoring.md § Skills Namespace Policy).

## §D. Constraints

- Tool set: Read, Write, Edit, Grep, Glob, Bash (Bash needed to invoke the Node render/lint scripts). No AskUserQuestion, no Agent.
- Node scripts: Node 18+ stdlib only — no npm dependency, no bundled binary.
- Template neutrality: skill body, references, AND script comments carry no internal tokens. CJK font-stack + Korean-budget content is retained (it is rendering capability, not an internal token).
- Progressive disclosure: description Level-1; body Level-2 (~5K); archetypes / authoring geometry / icon set / sketch preset in Level-3 references; scripts under `scripts/`.

## §E. Self-Verification (plan-phase)

- SPEC ID passes the canonical regex → PASS (Bash-executed).
- 12 canonical frontmatter fields present; `created`/`updated`; `tags` CSV string.
- Out-of-Scope section present with `### Out of Scope —` H3 sub-headings + `-` bullets.

## §F. Milestones (ordered by decision-reversibility — highest-change-likelihood first)

### M1 — Mermaid-vs-SVG selection rules + positioning (highest change-likelihood; review focus)
The decision most likely to be contested and the one that prevents dual-maintenance:
- Author the explicit selection-rules section (mermaid for markdown-embedded/frequently-changing/standard-types/4-locale-synced; svg-infographic for static-distribution/freeform/pixel-CJK).
- State that the skill is additive and does NOT replace the mermaid pipeline.
Covers REQ-SVG-012, 013.

### M2 — Skill interface + layout-before-code body (user-facing method)
- Frontmatter `description` + `when_to_use` (Level-1 trigger).
- The layout-before-code workflow (preflight → archetype pick → numeric layout pass → source lint → render → PNG verify).
- CJK-first font stack + Korean 60% budget; geometry-derived centers.
- The runtime prerequisite statement (Node/Chromium needed only for lint+render).
Covers REQ-SVG-002, 004, 005, 006, 011.

### M3 — Runtime gate + graceful degradation (behavior decision)
- Document the render path (`scripts/render.mjs`, 2× PNG, disclosed browser + IHDR verify).
- Document the lint path (`scripts/check-svg.mjs`, hard errors vs low-confidence warnings).
- Document the two degradation branches: no Node → manual checklist (no lint label); no Chromium → SVG-only + explicit limitation.
Covers REQ-SVG-007, 008, 009, 010.

### M4 — Level-3 references + bundled scripts (data/asset shape)
- Author `references/archetypes.md`, `references/authoring.md` (geometry/connector formulas, full icon set, manual render fallback), `references/sketch.md` (opt-in preset).
- Author `scripts/render.mjs` + `scripts/check-svg.mjs` (Node 18+ stdlib only, English comments, template-neutral).
- **Level-2 budget advisory (plan the split up front — measured, not a defect claim):** the sibling `moai-domain-html-report/SKILL.md` measures **22,676 bytes ≈ 5.7K tokens** at the ~4 bytes/token English-markdown ratio — already ABOVE the ~5K Level-2 target. AC-SVG-003's ≤ ~5K budget is therefore AGGRESSIVE for a skill of comparable scope, and this SPEC's scope (layout method + CJK budget + runtime gate + two degradation branches + selection rules) is at least as large. Run-phase MUST plan a heavier Level-3 offload than a naive reading of M4 implies: assume the archetype skeletons, the full geometry/connector formula set, the complete icon set, and the sketch preset ALL live in `references/`, with the body carrying pointers plus only the decision-critical prose. Measure the body (`wc -c`) before declaring AC-SVG-003 green.
- **Filename/layout provenance (not borrowed expression — TWO distinct bases, see spec.md §A.1):** the `scripts/` and `references/` DIRECTORY names follow moai's own `skill-authoring.md` § Skill Directory Layout table, which documents both. No existing moai skill ships a `scripts/` directory, so the basis is that documented rule, not in-catalog precedent. The two script FILENAMES have a different basis: `skill-authoring.md` prescribes no filename convention, so `render.mjs` / `check-svg.mjs` are independently-chosen generic functional descriptors (minimal expressive content), NOT derived from that rule and NOT reproduced from skillstead. Keeping the two bases distinct is what makes REQ-SVG-016's file-layout prohibition consistent with M4's filename mandate.
Covers REQ-SVG-003.

### M5 — Template-First placement + neutrality + clean-room authoring (mechanical, deferred)
- Create in the template path first, sync local, `make build`.
- Author clean-room: draft SKILL.md + references + scripts from this SPEC's functional-capability description ONLY, without consulting skillstead source text; add NO skillstead attribution (no footer provenance line, no `NOTICE` file/entry, no skillstead-TIED `license:` value) while DOING ship the house-convention `license: Apache-2.0` field. Record the clean-room PROCESS attestation in `progress.md` §E.2 (AC-SVG-019).
- Run the mechanical absence-invariant check (AC-SVG-016) across the WHOLE skill tree including `scripts/`: vendor-token grep → 0; no `NOTICE` in either skill directory nor at repo root; `grep -c '^license: Apache-2.0$'` → exactly 1 with no skillstead reference in the frontmatter. **This guard is NOT redundant with the existing CI guard**: `internal/template/internal_content_leak_test.go` `leakClasses` covers internal SPEC-IDs / REQ-AC tokens / audit citations / internal dates / memory paths — the vendor token `skillstead` is in none of them, so a vendor reference in a script comment would pass every existing repo check.
- Verify English body/scripts + template-neutrality CI guard clean (CJK content retained).
Covers REQ-SVG-001, 014, 015, 016, 017.

## §G. Anti-Patterns to avoid

- Positioning svg-infographic as a mermaid replacement (violates the CRITICAL constraint — must be complementary with explicit selection rules).
- Consulting skillstead source text while drafting the SKILL.md, the references, or the Node scripts (violates the clean-room process attestation AC-SVG-019 — draft from this SPEC's functional-capability description only).
- Treating a passing `grep -ril 'skillstead' → 0` as evidence the skill was IMPLEMENTED or that the scripts work. It is an absence-invariant guard only; an empty directory also returns 0. Positive proof lives in the functional ACs (AC-SVG-001, 003-013).
- Reading the 1,536-character listing cap as applying to `description` alone. The platform cap is the COMBINED `description` + `when_to_use` length (REQ-SVG-002), and M2 authors both fields.
- Adding a skillstead attribution footer line, a `NOTICE` file/entry, or a skillstead-TIED `license:` value (removed per the clean-room decision — none is added to the shipped files).
- OMITTING the house-convention `license: Apache-2.0` frontmatter field, or authoring a bare "no `license:` field present" absence check. The field is moai's own license declaration required by REQ-SVG-016; a bare-absence check would false-FAIL a correct implementation (this is the N1 defect fixed at v0.1.3).
- Verifying the AskUserQuestion/Agent prohibition with a whole-frontmatter grep. REQ-SVG-017 binds the `allowed-tools` value only, and a whole-frontmatter grep MEASURED 10 of 28 existing skills as false positives from `description` / `when_to_use` prose. Use the AC-SVG-018(d) exact-line match instead.
- Verifying the English-content AC with a CJK-codepoint scan. CJK font-family names and Korean sample labels are REQUIRED capability (REQ-SVG-006 / AC-SVG-015 / §D.2); a codepoint scan would false-FAIL content another AC mandates.
- Bundling an npm dependency tree or a Chromium binary into the template (violates Node-stdlib-only + neutrality).
- Hand-tuning per-language coordinate offsets instead of deriving from card geometry (violates REQ-SVG-005 and reintroduces the render-fix loop).
- Stripping CJK font-stack / Korean-budget content as if it were an "internal token" (it is rendering capability — retain it).
- Failing loudly when Node/Chromium is absent (violates graceful-degradation REQ-SVG-009/010).

## §H. Cross-References

- `moai-domain-html-report` SKILL.md — the mermaid pipeline this skill is complementary to.
- docs-site `layouts/partials/foot.html` — CDN mermaid loader (positioning reference; not modified).
- `.claude/rules/moai/development/skill-authoring.md` — frontmatter schema, namespace, progressive disclosure, `scripts/` layout.
- `internal/template/CLAUDE.md` — Template-First + embedded FS.
- `CLAUDE.local.md` §15 / §25 — 16-language neutrality + internal-content isolation.
- Sibling Epic SPECs: SPEC-SKILLPORT-CLAIM-CHECK-001, SPEC-SKILLPORT-HUMANIZE-LEDGER-001.
