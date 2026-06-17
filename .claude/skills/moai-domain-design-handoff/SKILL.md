---
name: moai-domain-design-handoff
description: >
  Claude Design handoff package specialist for /moai brain Phase 7. Assembles 5-file
  handoff bundle (prompt/context/references/acceptance/checklist) for paste-ready
  claude.com Design session. Handles brand-absent fallback and section regeneration.

when_to_use: >
  Use for /moai brain Phase 7 design handoff: assembling the 5-file Claude
  Design package (prompt, context, references, acceptance, checklist),
  brand-voice context, and paste-ready claude.com session bundles.

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
  related-skills: "moai-domain-ideation, moai-workflow-brain, moai-workflow-design"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

<!-- Verifies REQ-BRAIN-005: prompt.md is paste-ready (no MoAI tokens) -->
<!-- Verifies REQ-BRAIN-006: Brand voice integrated when present; graceful default when absent -->
<!-- Verifies REQ-BRAIN-009: Phase 7 exit AskUserQuestion with 3 options -->

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

`prompt.md` MUST follow this exact 5-section structure:

1. **Goal** — 2-3 sentences describing what needs to be designed, target users, top 3 value propositions from Lean Canvas UVP
2. **References** — 3-5 URLs to existing products with style notes, plus key aesthetic direction
3. **Brand Voice** — Two branches (brand_present vs brand_absent), see decision below
4. **Acceptance Criteria** — Concise non-negotiable requirements list (5-8 items)
5. **Out of Scope** — Explicit exclusions (3-5 items)

#### Section 3 — Brand Voice Decision Tree

- Branch A (`brand_present = true`): Extract personality + voice guidelines + color palette + typography from brand-voice.md and visual-identity.md
- Branch B (`brand_present = false`): Emit `## 3. Brand Voice (default — please customize)` header with explicit placeholder + instructions to either edit or run brand interview

See [5-section prompt template + brand branches detail](references/prompt-template.md) for verbatim section templates.

#### Prohibited Content in prompt.md

[HARD] The following MUST NOT appear anywhere in prompt.md:

- References to `SPEC-` identifiers (e.g., `SPEC-AUTH-001`)
- References to `.moai/` paths (e.g., `.moai/brain/`, `.moai/project/`)
- References to agent names (e.g., `manager-brain`, `manager-spec`)
- References to `IDEA-NNN` identifiers
- References to MoAI-specific commands (e.g., `/moai plan`, `/moai run`)
- Internal implementation details (file structures, Go code, database schemas)

The prompt must read as if written by a human product designer with no knowledge of MoAI's internal structure.

### Steps 2-5: Supporting Files

| Step | File | Purpose |
|------|------|---------|
| 2 | references.md | Competitor analysis + visual inspiration + UX pattern references (3-5 URLs from research.md Sources). Falls back to instructional note when URLs are scarce. |
| 3 | acceptance.md | Accessibility (WCAG 2.1 AA), Responsiveness (375/768/1280px), Brand Alignment, Content Completeness, Technical Constraints |
| 4 | context.md | Extended context — NOT for pasting into Claude Design. Full Lean Canvas summary, SPEC roadmap, research findings, brand context |
| 5 | checklist.md | Human self-check before pasting prompt.md: content review, MoAI-internal cleanup (auto-verified), scope verification, session readiness |

See [supporting files templates](references/supporting-files.md) for verbatim references.md, acceptance.md, context.md, and checklist.md templates.

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
- `moai-workflow-design`: Downstream consumer of `claude-design-handoff/` directory after user completes external Claude Design session (Path A handler)
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
