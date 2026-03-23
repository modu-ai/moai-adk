# Research: interface-design Plugin Integration

## Date: 2026-03-23
## Source: Team moai-plan-interface-design (researcher, analyst, architect)

---

## 1. Plugin Analysis (github.com/Dammyjay93/interface-design)

- Stars: 4,198 | Language: Shell | Type: Claude Code Plugin
- Core philosophy: "Sameness Is Failure" — every interface must emerge from specific context

### Key Concepts
1. **Intent-First Process**: Before any design, answer: Who? What? Feel?
2. **Product Domain Exploration**: 5+ domain concepts, 5+ color world, 1 signature element, defaults to avoid
3. **Design Memory** (`.interface-design/system.md`): Persists design decisions across sessions
4. **Commands**: /init, /extract, /audit, /critique
5. **Craft Foundations**: Subtle layering, surface elevation, token architecture

### Plugin File Structure
- `.claude/skills/interface-design/SKILL.md` (~400 lines)
- `.claude/commands/init.md`, `extract.md`, `audit.md`, `critique.md`
- `reference/examples/system-precision.md`, `system-warmth.md`
- `reference/system-template.md`

---

## 2. Current MoAI Design Architecture

### Existing Components
| Component | File | Status |
|-----------|------|--------|
| expert-frontend agent | `.claude/agents/moai/expert-frontend.md` | Has Pencil MCP tools, no design philosophy |
| team-designer agent | `.claude/agents/moai/team-designer.md` | Minimal (49 lines), no Intent-First |
| moai-design-tools skill | `.claude/skills/moai-design-tools/SKILL.md` | Figma/Pencil mechanics only |
| moai-domain-uiux skill | `.claude/skills/moai-domain-uiux/SKILL.md` | Tokens/WCAG/icons, Anti-AI Slop in modules/ |
| plan.md workflow | `.claude/skills/moai/workflows/plan.md` | No design direction phase |

### Gap Analysis
| Gap | Impact |
|-----|--------|
| No design memory system | Design decisions lost between sessions |
| No Intent-First process | Generic UI output regardless of product domain |
| No Design Direction phase in plan | Design not captured in SPEC documents |
| No design audit/critique workflow | No quality validation for design consistency |
| Preset-only customization (Nova default) | All projects look the same |

### Integration Points Identified
1. `plan.md` line ~136: Insert Design Direction sub-phase after Phase 0.5
2. `expert-frontend.md` skills list: Add new design-craft skill
3. `team-designer.md` skills list: Add new design-craft skill
4. New `.moai/design/system.md` stub in templates

---

## 3. Architecture Decisions

### AD1: New `moai-design-craft` skill (no overlap with existing)
### AD2: Design memory at `.moai/design/system.md` (MoAI namespace)
### AD3: Commands integrated into existing workflows (no new command files)
### AD4: +1 line skill addition to expert-frontend and team-designer
### AD5: 5 new files + 2 modified files in template source

See SPEC-DESIGN-001/spec.md for full details.
