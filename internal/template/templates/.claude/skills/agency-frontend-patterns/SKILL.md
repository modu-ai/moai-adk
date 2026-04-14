---
name: agency-frontend-patterns
description: >
  Frontend development patterns for AI Agency projects covering tech stack
  preferences, component architecture, file structure, and coding conventions.
  Self-evolves to match user's preferred frameworks and coding style.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, Bash, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "agency"
  status: "active"
  updated: "2026-04-02"
  evolution_count: "0"
  confidence_score: "0.00"
  base_context: "tech-preferences.md"
  dependencies: "agency-evaluation-criteria"
  tags: "frontend, react, nextjs, tailwind, components, architecture, code-patterns"
  related-skills: "agency-design-system, agency-evaluation-criteria"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["frontend", "react", "next.js", "component", "tailwind", "code", "implementation", "build"]
  agents: ["builder"]
  phases: ["run"]
---

# Agency Frontend Patterns Skill

Governs all code implementation decisions for AI Agency projects. Ensures consistent tech stack usage, component architecture, and coding conventions aligned with user preferences.

---

## Static Zone

### Identity

**Purpose**: Define and enforce frontend development patterns, file structure conventions, and code quality standards for agency project implementation.

**Input Contract**:
- Design specification (from designer output: `design-spec.md`)
- Copy document (from copywriter output: `copy.md`)
- Tech preferences (from `.agency/context/tech-preferences.md`)

**Output Contract**:
- Production-ready code files following design system
- Component architecture with proper separation of concerns
- Responsive implementation with mobile-first approach
- Accessible markup meeting WCAG 2.1 AA

**Owner**: `.agency/context/tech-preferences.md`

### Core Principles

> Derived from Brand Context. Never auto-modified. Manual editing only.

1. NEVER modify copy text from copywriter output — implement exactly as provided
2. Follow design system tokens exactly — no ad-hoc values
3. TDD approach: write tests first, then implement (RED-GREEN-REFACTOR)
4. Accessibility is not optional — semantic HTML, ARIA labels, keyboard navigation
5. Performance budget: Lighthouse >= 80 on all metrics

---

## Dynamic Zone

> Rules, Anti-Patterns, and Heuristics below evolve via user feedback.

### Rules

(No rules yet. Rules will be added as the system learns from user feedback.)

### Anti-Patterns

(No anti-patterns yet.)

### Heuristics

(No heuristics yet.)

---

## Evolution Log

- v1.0.0: Initial creation (Static Zone from Brand Context, empty Dynamic Zone)

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I will use whatever framework I know best" | The tech stack preference is a documented context. Using an unspecified framework creates maintenance burden for the project owner. |
| "File structure does not matter, I will organize later" | File structure establishes conventions that every future file follows. Reorganizing later means moving and updating every import. |
| "This coding convention is just a style preference" | Conventions reduce cognitive load for everyone who reads the code. Inconsistency forces readers to learn multiple patterns. |
| "Component architecture is overkill for a small page" | Small pages grow. Architecture established early scales; architecture added late requires refactoring. |
| "I will figure out the component boundary during implementation" | Component boundaries define reusability and testability. Figuring them out during implementation produces coupled, monolithic components. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Framework or library used that does not match `.agency/context/tech-preferences`
- File structure deviates from the documented convention without justification
- Components co-locate unrelated concerns (data fetching, presentation, business logic in one file)
- No component boundary planning before implementation
- Coding style inconsistent within the same file (mixed naming conventions, mixed patterns)

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Tech stack matches `.agency/context/tech-preferences` (show comparison)
- [ ] File structure follows documented convention (show directory tree)
- [ ] Components separate concerns (data, presentation, logic in distinct layers)
- [ ] Component boundaries planned and documented before implementation
- [ ] Coding conventions consistent across all files (naming, imports, exports)

<!-- moai:evolvable-end -->
