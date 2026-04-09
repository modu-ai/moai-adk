---
name: agency-evaluation-criteria
description: >
  Quality evaluation criteria for AI Agency project output covering design quality,
  originality, completeness, and functionality scoring with weighted dimensions
  and Playwright-based testing requirements.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "agency"
  status: "active"
  updated: "2026-04-02"
  evolution_count: "0"
  confidence_score: "0.00"
  base_context: "quality-standards.md"
  dependencies: ""
  tags: "evaluation, quality, testing, playwright, scoring, criteria, assessment"
  related-skills: "agency-copywriting, agency-design-system, agency-frontend-patterns"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["evaluate", "quality", "score", "test", "review", "assessment", "criteria", "pass", "fail"]
  agents: ["evaluator"]
  phases: ["run"]
---

# Agency Evaluation Criteria Skill

Governs quality assessment of all agency project deliverables. Enforces skeptical evaluation with evidence-based verdicts, weighted scoring dimensions, and automated testing via Playwright.

---

## Static Zone

### Identity

**Purpose**: Define evaluation criteria, scoring weights, pass/fail thresholds, and testing requirements for agency project quality assessment.

**Input Contract**:
- Built application (URL or local path)
- Original copy.md (for copy integrity verification)
- Original design-spec.md (for design compliance verification)
- BRIEF document (for completeness verification)

**Output Contract**:
- `evaluation-report.md` containing:
  - Overall score (0.00 - 1.00) with PASS/FAIL verdict
  - Per-dimension scores with evidence
  - Specific defect list with file:line references
  - Screenshots (desktop + mobile)
  - Improvement recommendations

**Owner**: `.agency/context/quality-standards.md`

### Core Principles

> Derived from Brand Context. Never auto-modified. Manual editing only.

1. Skeptical by default — tuned to find defects, not rationalize acceptance
2. Evidence-based verdicts only — no PASS without concrete proof
3. Copy integrity is non-negotiable — any deviation from original copy = FAIL
4. AI slop detection — purple gradients + white cards + generic icons = FAIL
5. When in doubt, FAIL — false negatives are costlier than false positives

### Default Evaluation Weights

| Dimension | Weight | Description |
|-----------|--------|-------------|
| Design Quality | 30% | Visual consistency, brand alignment, polish |
| Originality | 25% | Not generic/template-like, unique approach |
| Completeness | 25% | All BRIEF sections present, copy accurate |
| Functionality | 20% | Responsive, accessible, all interactions work |

### Hard Thresholds (always FAIL)

- Copy text differs from original copy.md
- AI slop detected (generic purple gradient + white card layout)
- Mobile viewport broken (content overflow or unreadable)
- CTA count > 1 per page (unless BRIEF explicitly specifies)
- Any link returns 404
- Lighthouse Accessibility < 80

### Testing Requirements

#### Phase 1: Visual Verification
- Desktop screenshot (1280x720): full page capture
- Mobile screenshot (375x667): full page capture
- Compare against design-spec.md layout expectations
- Check for AI slop indicators (purple gradients, white cards, generic stock icons)

#### Phase 2: Interaction Testing
- Click test: all buttons, links, CTAs — verify navigation and state changes
- Form test: all input fields with valid/invalid data, verify validation messages
- Scroll test: full page traversal, verify lazy loading and fixed elements
- Keyboard test: Tab navigation order, Enter/Space activation on interactive elements

#### Phase 3: Responsive Verification
- Mobile viewport: no horizontal overflow, readable text (min 16px)
- Touch targets: minimum 44x44px for all interactive elements
- Mobile navigation: hamburger menu functionality, swipe gestures
- Breakpoint transitions: verify layout changes at 768px, 1024px, 1280px

#### Phase 4: Performance Audit
- Lighthouse scores: Performance >= 80, Accessibility >= 90, Best Practices >= 80, SEO >= 80
- Core Web Vitals: LCP < 2.5s, FID < 100ms, CLS < 0.1
- Bundle size check: main bundle < 200KB gzipped (unless justified)

#### Tool Priority
1. claude-in-chrome MCP (preferred): Live browser interaction with screenshots
2. Playwright via Bash (fallback): Headless testing when MCP unavailable
3. Static analysis (minimum): HTML/CSS validation, link checking

---

## Dynamic Zone

> Weights, thresholds, and test scenarios evolve via user feedback.

### Rules

(No rules yet. Rules will be added as the system learns from user feedback.)

### Anti-Patterns

(No anti-patterns yet.)

### Heuristics

(No heuristics yet.)

---

## Evolution Log

- v1.0.0: Initial creation (Static Zone with default weights, empty Dynamic Zone)
