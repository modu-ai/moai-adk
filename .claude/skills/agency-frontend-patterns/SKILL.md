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
