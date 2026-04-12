# SPEC-SKILL-ENHANCE-001: Skill Anti-Rationalization, Red Flags, and Verification

## Meta

- **Status**: Draft
- **Wave**: 1 (parallel with CORE-BEHAV-001, TELEMETRY-001)
- **Created**: 2026-04-11
- **Origin**: addyosmani/agent-skills analysis (P0.1, P0.2, P0.3, P2.3)
- **Blocked By**: SPEC-EVO-001 (needs evolvable zone markers)

## Objective

Add three structural sections to ALL 41 moai/agency SKILL.md files, adopting the proven anti-rationalization prompt engineering pattern from addyosmani/agent-skills. Each addition is wrapped in evolvable zone markers so it can evolve independently of template updates.

## Background

Analysis of addyosmani/agent-skills revealed that every skill contains:
1. **Common Rationalizations table**: Excuses agents use to skip steps + factual rebuttals
2. **Red Flags section**: Observable symptoms that the skill is being violated
3. **Verification checklist**: Concrete evidence checkboxes for "done" criteria

None of the current 41 moai-adk skills have any of these sections. This is a pure additive change (no existing content modified) that enhances prompt engineering quality.

Additionally, named engineering principles (Beyonce Rule, Chesterton's Fence, Hyrum's Law, Rule of 500) will be cited in relevant skills for improved teachability.

## Requirements (EARS Format)

### R1: Anti-Rationalization Tables [UBIQ]

Every SKILL.md SHALL contain a `## Common Rationalizations` section with a markdown table of 5-7 rows in `| Rationalization | Reality |` format.

**Acceptance Criteria:**
- [ ] All 41 SKILL.md files have `## Common Rationalizations` section
- [ ] Each table has minimum 5, maximum 7 rows
- [ ] Rationalizations are SPECIFIC to the skill's domain (not generic)
- [ ] Reality column provides factual rebuttals (not opinions)
- [ ] Section is wrapped in `<!-- moai:evolvable-start id="rationalizations" -->` markers
- [ ] Content is in English (code_comments language rule)

### R2: Red Flags Sections [UBIQ]

Every SKILL.md SHALL contain a `## Red Flags` section with observable violation signals.

**Acceptance Criteria:**
- [ ] All 41 SKILL.md files have `## Red Flags` section
- [ ] Each section has 4-6 bullet points
- [ ] Flags are OBSERVABLE symptoms (not abstract principles)
- [ ] Format: `- [Observable behavior that indicates skill violation]`
- [ ] Section is wrapped in `<!-- moai:evolvable-start id="red-flags" -->` markers

### R3: Verification Checklists [UBIQ]

Every SKILL.md SHALL contain a `## Verification` section with evidence-based checkboxes.

**Acceptance Criteria:**
- [ ] All 41 SKILL.md files have `## Verification` section
- [ ] Each section has 4-8 checkbox items (`- [ ]` format)
- [ ] Every checkbox requires EVIDENCE (test output, file existence, command result)
- [ ] No checkbox uses subjective language ("seems right", "looks good")
- [ ] Section is wrapped in `<!-- moai:evolvable-start id="verification" -->` markers

### R4: Named Principles Citations [OPT]

WHERE a skill's domain overlaps with established engineering principles, the SKILL.md SHOULD cite the principle by name with brief context.

**Principle-to-Skill Mapping:**
| Principle | Source | Target Skills |
|-----------|--------|---------------|
| Beyonce Rule | Google SWE Book | moai-workflow-tdd, moai-workflow-testing |
| Chesterton's Fence | G.K. Chesterton | moai-workflow-ddd, moai-foundation-quality |
| Hyrum's Law | Google | moai-domain-backend, moai-ref-api-patterns |
| Rule of 500 | Google SWE Book | moai-tool-ast-grep, moai-workflow-loop |
| Shift Left | DevOps | moai-foundation-quality, moai-workflow-testing |
| DAMP over DRY | Google Testing | moai-workflow-tdd, moai-ref-testing-pyramid |

**Acceptance Criteria:**
- [ ] Each mapped skill includes principle citation in relevant section
- [ ] Citation format: `**[Principle Name]**: [One-sentence explanation + applicability]`
- [ ] Citations placed inline in existing process sections (not a separate section)

### R5: Template-First Rule [UBIQ]

ALL changes SHALL be made in `internal/template/templates/` first, then `make build` to regenerate embedded files.

**Acceptance Criteria:**
- [ ] Every modified SKILL.md has its template counterpart updated
- [ ] `make build` succeeds without errors after all changes
- [ ] Local copies (`.claude/skills/`) match template output
- [ ] `go test ./internal/template/...` passes

## Target Files (41 skills)

### Workflow Skills (17)
moai/SKILL.md, moai-workflow-{ddd,tdd,loop,spec,project,templates,testing,research,thinking,worktree,jit-docs}/SKILL.md, moai-workflow-{plan,run,sync,fix,e2e}.md (inline workflows, different path)

### Foundation Skills (6)
moai-foundation-{cc,core,context,philosopher,quality,thinking}/SKILL.md

### Domain Skills (4)
moai-domain-{backend,frontend,database,uiux}/SKILL.md

### Reference Skills (4)
moai-ref-{api-patterns,git-workflow,owasp-checklist,react-patterns,testing-pyramid}/SKILL.md

### Library/Platform/Tool/Design Skills (10)
moai-library-{mermaid,nextra,shadcn}, moai-platform-{auth,chrome-extension,database-cloud,deployment}, moai-tool-{ast-grep,svg}, moai-framework-electron, moai-formats-data, moai-docs-generation, moai-design-{craft,tools}

### Agency Skills (6)
agency/SKILL.md, agency-{client-interview,copywriting,design-system,evaluation-criteria,frontend-patterns}/SKILL.md

## Section Placement

New sections are appended BEFORE the closing `---` or at the end of the SKILL.md body:

```markdown
[...existing skill content...]

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| ... | ... |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- ...

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] ...

<!-- moai:evolvable-end -->
```

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Generic rationalizations across skills | Low prompt effectiveness | Review each skill's domain; use skill-specific language |
| Token overhead (3 sections x 41 skills) | ~200 tokens per skill at Level 2 load | Progressive disclosure already limits loaded skills |
| Evolvable markers in templates confuse merge | Merge errors | SPEC-EVO-001 provides marker-aware merge first |

## Execution Strategy

Given 41 files, batch by category:
1. Workflow skills first (highest impact, most used)
2. Foundation + domain skills
3. Reference + library/platform skills
4. Agency skills

Each batch: edit template → `make build` → verify local copy → run tests.

## Dependencies

- SPEC-EVO-001: Evolvable zone markers and merge support must exist before adding markers to skills

## Non-Goals

- Modifying existing skill process/workflow content
- Adding new skills (that's SPEC-REFLECT-001's output)
- Changing YAML frontmatter or allowed-tools
