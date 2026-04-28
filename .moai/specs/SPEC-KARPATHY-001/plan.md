---
id: SPEC-KARPATHY-001
version: "0.1.0"
status: planned
created_at: 2026-04-28
updated_at: 2026-04-28
author: manager-spec
priority: High
labels: [karpathy, plan, milestones, templates]
issue_number: null
---

# SPEC-KARPATHY-001 -- Implementation Plan

## 1. Strategy Overview

3-Wave decomposition. Wave 1 creates new files (anti-pattern skill + quick reference rule). Wave 2 amends the constitution (Evolvable zone only). Wave 3 handles Template-First synchronization + regression testing.

| Wave | Goal                                      | Files | Mode          | Isolation |
|------|-------------------------------------------|-------|---------------|-----------|
| Wave 1 | Anti-pattern skill + quick reference rule creation | 2 new | acceptEdits   | None      |
| Wave 2 | Constitution amendments (3 additions)     | 1 amend | acceptEdits | None      |
| Wave 3 | Template mirror sync + build + test       | All   | acceptEdits   | None      |

**Why no team mode?** Deliverable is 3 markdown files (2 new + 1 amendment), file count < 10, domain count = 1 (rules/skills). Team mode threshold (`domains>=3, files>=10`) not met. Sequential sub-agent execution is token-efficient.

---

## 2. Milestones (Priority-based, no time estimates)

### M1 (Priority High) -- Anti-Pattern Reference Skill

- Create `.claude/skills/moai/references/anti-patterns.md`
- 8 anti-pattern categories with concrete wrong/right code examples
- Primary examples in Go, secondary in Python/TypeScript
- Progressive disclosure: Level 1 metadata (~100 tokens), Level 2 full body (~5000 tokens)
- Skill frontmatter configuration:
  - `triggers.keywords: ["anti-pattern", "over-engineering", "refactor", "simplify", "scope", "style drift"]`
  - `triggers.agents: ["expert-backend", "expert-frontend", "manager-quality", "evaluator-active"]`
  - `triggers.phases: ["run", "review"]`
  - `user-invocable: false`
- Each category maps to corresponding Agent Core Behavior

**Exit criteria**: File exists, 8 categories present with Go examples, progressive disclosure frontmatter valid, 16-language neutrality verified.

### M2 (Priority High) -- Quick Reference Rule

- Create `.claude/rules/moai/development/karpathy-quickref.md`
- Paths frontmatter: `**/*.go,**/*.py,**/*.ts,**/*.js,**/*.java,**/*.rs`
- Map 4 Karpathy principles to 6 Agent Core Behaviors
- Concrete checkpoint questions per principle (3-5 each)
- Cross-reference to anti-pattern skill for detailed examples

**Exit criteria**: File exists, paths frontmatter covers 6 languages, 4 principles mapped, checkpoint questions present.

### M3 (Priority High) -- Constitution Amendments

- Amend `.claude/rules/moai/core/moai-constitution.md`
- Target: `<!-- moai:evolvable-start id="agent-core-behaviors" -->` section only
- Amendment A (Behavior 4, Enforce Simplicity): Add quantitative LOC trigger -- max 6 lines
- Amendment B (Behavior 5, Scope Discipline): Add style-matching directive -- max 4 lines
- Amendment C (Behavior 6, Verify Don't Assume): Add goal-to-test pattern -- max 6 lines
- Total addition: max 16 lines (within +20 line constraint)
- Verify frozen zone untouched (Zone Registry CONST-V3R2-025..046)

**Exit criteria**: 3 amendments added, total line count within budget, evolvable-only zone verified, no frozen content modified.

### M4 (Priority Medium) -- Template-First Synchronization

- Mirror new files to `internal/template/templates/`:
  - `internal/template/templates/.claude/skills/moai/references/anti-patterns.md`
  - `internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md`
  - `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` (updated)
- Run `make build` to regenerate `internal/template/embedded.go`
- Run `go test ./internal/template/...` for regression verification
- Byte-identical verification between local and template copies

**Exit criteria**: All files mirrored, `make build` succeeds, all template tests green, byte-identical confirmed.

### M5 (Priority Medium) -- Validation

- Run AC-001 through AC-06 acceptance criteria verification
- 16-language neutrality grep check on all new files
- Constitution evolvable zone grep check (no frozen content modified)
- Progressive disclosure token count verification
- Anti-pattern skill trigger configuration validation

**Exit criteria**: All ACs PASS, neutrality verified, evolvable-only zone confirmed.

---

## 3. Technical Approach

### 3.1 Constitution Amendment Strategy

The constitution uses `<!-- moai:evolvable-start/end -->` markers around the Agent Core Behaviors section (lines 177-268). Amendments are additive -- no existing text is removed or modified. Each amendment appends within the corresponding Behavior subsection:

```
### 4. Enforce Simplicity [HARD]
<existing text preserved>
+ Quantitative trigger: <new text>
```

This ensures:
- Frozen zone (Zone Registry entries CONST-V3R2-025..046) remains untouched
- Evolvable zone markers preserved
- Amendment additions are clearly identifiable

### 3.2 Anti-Pattern Skill Progressive Disclosure

Level 1 (metadata, always loaded):
```yaml
---
name: moai-reference-anti-patterns
description: 8 Karpathy-inspired anti-patterns with wrong/right code examples
triggers:
  keywords: ["anti-pattern", "over-engineering", "refactor", "simplify", "scope", "style drift"]
  agents: ["expert-backend", "expert-frontend", "manager-quality", "evaluator-active"]
  phases: ["run", "review"]
user-invocable: false
progressive_disclosure:
  level1_tokens: 100
  level2_tokens: 5000
---
```

Level 2 (body, loaded on trigger match):
- 8 categories, each with:
  - Category name + mapped Behavior
  - WRONG code example (Go primary)
  - RIGHT code example (Go primary)
  - Optional secondary example (Python or TypeScript)
  - Detection heuristic (keywords/patterns to watch for)

### 3.3 Quick Reference Rule Frontmatter

```yaml
---
description: Karpathy coding principles mapped to MoAI Agent Core Behaviors with checkpoint questions
paths: "**/*.go,**/*.py,**/*.ts,**/*.js,**/*.java,**/*.rs,**/*.c,**/*.cpp,**/*.rb,**/*.php,**/*.kt,**/*.swift,**/*.dart,**/*.ex,**/*.scala,**/*.hs,**/*.zig"
---
```

The `paths` frontmatter ensures the rule loads only when editing code files, keeping token overhead minimal for non-code contexts.

### 3.4 16-Language Neutrality

- Code examples: Go (primary, matching MoAI-ADK's implementation language), Python/TypeScript (secondary)
- Quick reference: No code examples, pure principle-to-behavior mapping
- Anti-pattern descriptions: Language-agnostic terminology
- Trigger keywords: English-only (aligned with MoAI instruction language policy)

---

## 4. Risks and Mitigations

| Risk                                           | Likelihood | Impact | Mitigation                                                        |
|------------------------------------------------|------------|--------|-------------------------------------------------------------------|
| Constitution amendments exceed +20 line budget  | Low        | High   | M3 pre-calculates line count; each amendment has individual max   |
| Anti-pattern skill Level 2 exceeds token budget | Low        | Medium | Keep each category to ~500 tokens; total 8 * 500 = 4000 < 6000   |
| Frozen zone accidentally modified              | Low        | Critical | M3 targets only evolvable zone; grep verification in M5          |
| Template mirror drift                          | Low        | High   | M4 byte-identical diff check                                      |
| Skill triggers too broad (false positive loads)| Medium     | Low    | Use specific trigger keywords + agent types + phase restrictions  |
| Quick reference paths frontmatter too narrow   | Low        | Medium | Cover 6 major languages; additional languages via extension       |
| Anti-pattern examples favor Go over other langs | Medium     | Medium | Secondary examples in Python/TypeScript; descriptions are neutral |

---

## 5. Dependencies

- [BLOCKING] `make build`: requires Go toolchain (already present)
- [BLOCKING] `go test ./internal/template/...`: requires existing test infrastructure
- [SOFT] SPEC-V3R3-PATTERNS-001: Precedent for reference file integration pattern (completed)
- [SOFT] SPEC-V3R2-CON-001: Zone Registry definitions (completed, ensures frozen zone identification)

---

## 6. Open Questions

| ID    | Question                                                      | Resolution Path                                                 |
|-------|---------------------------------------------------------------|-----------------------------------------------------------------|
| OQ-01 | Should anti-pattern skill support Level 3 (bundled) for very large codebases? | M1: Start with 2-level. Level 3 can be added in a follow-up SPEC if needed. |
| OQ-02 | Should quick reference rule cover all 16 languages in paths frontmatter? | M2: Start with 6 most common. Remaining 10 can be added via config. |
| OQ-03 | LOC ratio threshold: is 3x the right multiplier?              | M3: Start with 3x. User can adjust via feedback. Document as tunable. |

---

## 7. Out of Scope (Reaffirmation)

- NOTICE.md or attribution file creation/modification
- New agent definitions
- Breaking changes to existing rules
- Language-specific bias in templates
- Upstream sync automation with forrestchang repo
- IDE plugins or external tooling
- New SPEC workflow phases
- Custom MX tag types
