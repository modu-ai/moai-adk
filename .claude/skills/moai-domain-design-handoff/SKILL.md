<!-- Verifies REQ-BRAIN-005: prompt.md is paste-ready (no MoAI tokens) -->
<!-- Verifies REQ-BRAIN-006: Brand voice integrated when present; graceful default when absent -->
<!-- Verifies REQ-BRAIN-009: Phase 7 exit AskUserQuestion with 3 options -->
---
name: moai-domain-design-handoff
description: >
  Claude Design handoff package specialist for /moai brain Phase 7. Assembles 5-file
  handoff bundle (prompt/context/references/acceptance/checklist) for paste-ready
  claude.com Design session. Handles brand-absent fallback and section regeneration.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-05-04"
  modularized: "false"
  tags: "design-handoff, claude-design, prompt-template, brand, acceptance, brain"
  related-skills: "moai-domain-ideation, moai-workflow-brain, moai-workflow-design-import"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["design handoff", "claude design", "prompt template", "brand voice", "handoff package", "brain"]
  agents: ["manager-brain"]
  phases: ["brain"]
---

<!-- @MX:ANCHOR: [AUTO] 5-section prompt.md template structure — canonical definition -->
<!-- @MX:REASON: Consumed by every brain workflow Phase 7 execution (high fan_in). Structural changes affect user trust — prompt.md is pasted directly into external claude.com Design session. -->

# Design Handoff Domain Specialist

Assembles the 5-file Claude Design handoff package for the brain workflow's Phase 7. The package is designed for paste-and-go use in the external claude.com Design product.

## Quick Reference

The handoff package lives at `.moai/brain/IDEA-NNN/claude-design-handoff/`:

| File | Purpose | Paste target |
|------|---------|-------------|
| `prompt.md` | Master prompt — paste directly into claude.com Design | Yes (primary) |
| `context.md` | Extended context for reference during design session | Optional supplement |
| `references.md` | Visual reference URLs and design inspiration sources | Referenced in prompt |
| `acceptance.md` | Design acceptance criteria (WCAG, responsive, brand) | Referenced in prompt |
| `checklist.md` | Pre-paste self-check before using in claude.com Design | Human review tool |

Key guarantees:
- [HARD] `prompt.md` contains NO MoAI-specific tokens (no `SPEC-`, `.moai/`, `manager-`, `IDEA-`)
- [HARD] Brand voice integrated when `.moai/project/brand/brand-voice.md` exists
- [HARD] Brand-absent fallback: `Brand Voice (default — please customize)` placeholder section
- [HARD] Phase 7 exits with AskUserQuestion offering 3 options (a/b/c per REQ-BRAIN-009)
- [HARD] All 5 files produced regardless of brand context availability

---

## Phase 7: Handoff Package Assembly

### Input

- `ideation.md` (Lean Canvas + Evaluation Report from Phases 4 and 5)
- `proposal.md` (product summary and SPEC decomposition from Phase 6)
- Optional: `.moai/project/brand/brand-voice.md` (brand context)
- Optional: `.moai/project/brand/visual-identity.md` (design tokens, colors)

### Step 0: Brand Context Detection

Before writing any file, check brand context:

```
IF .moai/project/brand/brand-voice.md exists AND is non-empty:
  Load brand voice → use in Brand Voice section of prompt.md
  SET brand_present = true
ELSE:
  Use default brand voice placeholder
  SET brand_present = false
  Note: will include AskUserQuestion offer to run brand interview
```

### Step 1: Assemble prompt.md

<!-- @MX:WARN: [AUTO] prompt.md template — output pasted into external claude.com Design session -->
<!-- @MX:REASON: Changes to this template affect what users paste into claude.com. Structural changes can break user's design sessions. Validate against current claude.com Design prompt guidelines before modifying. -->

#### 5-Section Template

`prompt.md` MUST follow this exact 5-section structure:

```markdown
# Design Brief: {product name from proposal.md}

## 1. Goal

{2-3 sentences describing what needs to be designed.}

I need a complete visual design for a {product description} — specifically the {scope: landing page / dashboard / mobile app / web app / etc.}.

The design should communicate: {top 3 value propositions from Lean Canvas UVP block}

Target users: {Customer Segments from Lean Canvas, 1-2 sentences}

---

## 2. References

For visual inspiration and style direction, please study these references:

{List of URLs from references.md — 3-5 URLs to existing products with brief style notes}

Key aesthetic direction:
- {style adjective 1}: {brief explanation}
- {style adjective 2}: {brief explanation}
- {style adjective 3}: {brief explanation}

---

## 3. Brand Voice

{EITHER brand_present branch OR brand_absent branch — see below}

---

## 4. Acceptance Criteria

The design MUST satisfy these non-negotiable requirements:

{Concise list from acceptance.md — typically 5-8 items}

---

## 5. Out of Scope

Do NOT design:

{Explicit exclusions — typically 3-5 items}

```

#### Section 3 — Brand Voice: Two Branches

**Branch A: Brand present** (`brand_present = true`):

```markdown
## 3. Brand Voice

The brand personality is: {from brand-voice.md tone/personality fields}

Voice guidelines:
{Extract 3-5 most actionable brand voice rules from brand-voice.md}

Color palette (from brand identity):
{If visual-identity.md present: list primary colors with hex codes}
{If visual-identity.md absent: "Brand colors TBD — please use {tone}-appropriate palette"}

Typography:
{If visual-identity.md present: font families}
{If visual-identity.md absent: "Typography TBD — sans-serif for readability"}
```

**Branch B: Brand absent** (`brand_present = false`):

```markdown
## 3. Brand Voice (default — please customize)

> NOTE: This project does not yet have a defined brand voice. The placeholders below
> are generic suggestions. Before using this prompt in Claude Design, either:
> (a) Edit this section with your actual brand voice, OR
> (b) Run /moai brain brand-interview (when available) to define brand context

Brand personality: professional, approachable, modern

Voice guidelines:
- Clear and concise language; no jargon
- Action-oriented CTAs
- Friendly but credible tone

Color palette: neutral, modern (grays, whites, one accent color — TBD)
Typography: clean sans-serif (Inter, Geist, or similar)
```

#### Prohibited Content in prompt.md

[HARD] The following MUST NOT appear anywhere in prompt.md:
- References to `SPEC-` identifiers (e.g., `SPEC-AUTH-001`)
- References to `.moai/` paths (e.g., `.moai/brain/`, `.moai/project/`)
- References to agent names (e.g., `manager-brain`, `manager-spec`)
- References to `IDEA-NNN` identifiers
- References to MoAI-specific commands (e.g., `/moai plan`, `/moai run`)
- Internal implementation details (file structures, Go code, database schemas)

The prompt must read as if written by a human product designer with no knowledge of MoAI's internal structure.

### Step 2: Assemble references.md

Populate from research.md's Sources section. Select 3-5 URLs that represent:
1. Existing competitors (to show what the user wants to improve on)
2. Design inspiration from adjacent products (visual quality reference)
3. User experience patterns relevant to the target use case

Format:
```markdown
# Design References

## Competitor Analysis

{URL}: {product name} — {what the design should improve on or learn from}

## Visual Inspiration

{URL}: {product name} — {specific visual quality to emulate: e.g., "clean typography", "card layout", "mobile-first nav"}

## UX Pattern References

{URL}: {product name or pattern} — {specific interaction pattern relevant to the design}
```

If research.md has fewer than 3 URLs (e.g., WebSearch failed), include a note:
```markdown
*Note: Limited references available due to research tool availability. Add 2-3 URLs of products you admire.*
```

### Step 3: Assemble acceptance.md

Design acceptance criteria derived from Lean Canvas + product type:

```markdown
# Design Acceptance Criteria

These criteria must be met for the design to be considered complete.

## Accessibility
- [ ] WCAG 2.1 AA compliance (minimum contrast ratio 4.5:1 for normal text)
- [ ] Interactive elements have visible focus states
- [ ] Alt text descriptions provided for all images and icons

## Responsiveness
- [ ] Mobile-first design (base breakpoint: 375px)
- [ ] Tablet layout defined (768px breakpoint)
- [ ] Desktop layout defined (1280px breakpoint)

## Brand Alignment
- [ ] Color palette consistent with brand voice section of prompt.md
- [ ] Typography consistent and readable
- [ ] Visual hierarchy reflects product priority (UVP communicated first)

## Content Completeness
- [ ] Hero section includes: headline, subheadline, primary CTA
- [ ] Core features visually communicated (minimum 3 features)
- [ ] Social proof element present (testimonial, stat, or logo row)

## Technical Constraints
- [ ] No animations or complex interactions in v1 (static design only)
- [ ] Design system uses reusable components (cards, buttons, inputs)
```

Customize the checklist based on the specific product type identified in proposal.md.

### Step 4: Assemble context.md

Extended context for the design session — NOT pasted into the prompt, kept as reference:

```markdown
# Extended Context: {product name}

> This file supplements prompt.md with additional context for your design session.
> It is NOT meant to be pasted into Claude Design — use prompt.md for that.

## Product Background

{Full Lean Canvas summary from ideation.md — all 9 blocks}

## SPEC Roadmap Context

{List of SPEC decomposition candidates from proposal.md — helps designer understand scope and what is out of scope for v1}

## Research Findings Summary

{Executive summary from research.md — key market insights and competitive dynamics}

## Brand Context

{If brand present: full brand-voice.md content}
{If brand absent: placeholder and instructions to populate .moai/project/brand/}
```

### Step 5: Assemble checklist.md

Human self-check before pasting prompt.md into claude.com Design:

```markdown
# Pre-Paste Checklist

Before pasting prompt.md into Claude Design, verify:

## Content Review
- [ ] Goal section accurately describes what you want designed
- [ ] References section has 3-5 URLs you have checked and are relevant
- [ ] Brand Voice section reflects your actual brand (not placeholder)
- [ ] Acceptance Criteria section reflects your real quality bar

## MoAI-Internal Cleanup (auto-verified)
- [ ] No SPEC- identifiers in prompt.md
- [ ] No .moai/ path references in prompt.md
- [ ] No /moai commands in prompt.md

## Scope Verification
- [ ] Out of Scope section lists things you explicitly do NOT want
- [ ] The "Goal" section is scoped to one page/view (not the entire product)

## Session Readiness
- [ ] You have a claude.com account with Design access
- [ ] You have reviewed IDEA-NNN/proposal.md and know what this design supports
- [ ] You are prepared to provide feedback on the generated design

After design is complete:
- Copy the Claude Design output to a local bundle directory
- Run: /moai design --path A --bundle <path-to-bundle>
```

---

## Phase 7 Exit: AskUserQuestion (REQ-BRAIN-009)

After all 5 files are written, the workflow MUST invoke AskUserQuestion (with ToolSearch preload) presenting 3 options:

```
ToolSearch(query: "select:AskUserQuestion")
AskUserQuestion({
  questions: [{
    question: "핸드오프 패키지가 준비되었습니다. 다음 단계를 선택하세요.",
    header: "Brain Workflow 완료",
    options: [
      {
        label: "/moai project 실행 (권장)",
        description: "IDEA-NNN/proposal.md 기반으로 product.md, structure.md, tech.md 프로젝트 문서 생성. 이후 /moai plan으로 첫 SPEC 작성 가능."
      },
      {
        label: "수동 검토",
        description: "핸드오프 파일을 직접 검토하고 필요한 경우 편집. .moai/brain/IDEA-NNN/ 디렉토리를 확인하세요. 준비가 되면 /moai project --from-brain IDEA-NNN을 실행하세요."
      },
      {
        label: "핸드오프 패키지 재생성",
        description: "prompt.md 또는 다른 파일에 수정이 필요한 경우 어떤 부분을 변경할지 알려주세요. 해당 파일만 재생성합니다."
      }
    ]
  }]
})
```

For non-Korean conversation_language, translate option labels and descriptions accordingly.

---

## Works Well With

- `moai-domain-ideation`: Consumes ideation.md and proposal.md as primary inputs
- `moai-domain-research`: Pulls reference URLs from research.md Sources section
- `moai-workflow-design-import`: Downstream consumer of `claude-design-handoff/` directory after user completes external Claude Design session
- `moai-workflow-brain`: Orchestrates Phase 7 execution with IDEA-NNN directory management

---

## Common Rationalizations

| Rationalization | Reality |
|----------------|---------|
| "Including SPEC-AUTH-001 in prompt.md helps the designer understand scope" | prompt.md is for claude.com Design, not MoAI. SPEC IDs are internal. Use the Out of Scope section to describe scope boundaries in plain English. |
| "I should skip checklist.md — it's obvious" | Checklist.md prevents the most common error: pasting a prompt with placeholder Brand Voice. It takes 30 seconds to complete and saves a bad design session. |
| "references.md is optional if research had no URLs" | references.md is always produced. When URLs are scarce, include a note asking the user to add their own. An empty references file is worse than one with instructions. |
| "If brand is absent, skip the Brand Voice section" | Brand Voice section is always present. Brand-absent path produces an explicit placeholder with instructions — clearer than a missing section. |

## Verification

- [ ] All 5 files produced: prompt.md, context.md, references.md, acceptance.md, checklist.md
- [ ] prompt.md has exactly 5 sections (Goal, References, Brand Voice, Acceptance, Out of Scope)
- [ ] prompt.md contains no SPEC- identifiers
- [ ] prompt.md contains no .moai/ path references
- [ ] prompt.md contains no manager- or /moai references
- [ ] Brand-absent path includes "Brand Voice (default — please customize)" header in prompt.md
- [ ] context.md includes note that it is NOT for pasting into Claude Design
- [ ] Phase 7 exit AskUserQuestion called with exactly 3 options
