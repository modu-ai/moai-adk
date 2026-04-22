---
title: Getting Started
description: Start hybrid design workflow with /moai design command
weight: 20
draft: false
---

# Getting Started

## Prerequisites

- MoAI-ADK project initialized
- `.moai/project/brand/` directory ready or to be created
- Claude Code desktop client v2.1.50 or later

## Brand Context Interview

Running `/moai design` initiates a **brand context interview** first.

### Creates Three Brand Files

The interview generates these files in `.moai/project/brand/`:

1. **brand-voice.md** — Brand tone, terminology, messaging guidelines
2. **visual-identity.md** — Colors, typography, visual language
3. **target-audience.md** — Customer profile and preferences

### Interview Process

1. Run `/moai design` in Claude Code
2. See message: "Incomplete brand context detected"
3. Select brand interview option
4. `manager-spec` agent presents questions
5. Provide free-form answers
6. Three files auto-generated

Example questions:
- "Is your brand tone professional or approachable?"
- "Choose your three primary brand colors"
- "What is your target customer's main pain point?"

## Path Selection

After brand context setup, a path choice UI appears:

### Option 1 (Recommended) — Use Claude Design

Generate design in **Claude.ai/design**, then export as **handoff bundle**

**Requirements:**
- Claude.ai Pro, Max, Team, or Enterprise subscription

**Advantages:**
- Intuitive UI/UX
- Real-time collaboration (Team subscription)
- Multiple input formats supported (text, images, Figma, GitHub repo)

### Option 2 — Code-Based Design

Use **moai-domain-copywriting** and **moai-domain-brand-design** skills

**Requirements:**
- Complete `brand-voice.md` and `visual-identity.md`

**Advantages:**
- No additional subscription required
- Automated design token generation
- Version control friendly

## First Run

```bash
# Run in Claude Code
/moai design
```

Execution sequence:
1. Check for `.agency/` (show migration guide)
2. Verify brand context
3. Interview missing files
4. Show path choice UI
5. Execute chosen path workflow

## Verify Setup

Check generated brand files:

```bash
ls -la .moai/project/brand/
# brand-voice.md
# visual-identity.md
# target-audience.md
```

## Next Steps

- **Path A selected:** See [Claude Design Handoff](./claude-design-handoff.md) guide
- **Path B selected:** See [Code-Based Path](./code-based-path.md) guide

## Troubleshooting

### Update Brand Context

Edit files directly:

```bash
# Edit in your preferred editor
vim .moai/project/brand/brand-voice.md
```

Changes apply automatically on next `/moai design` run.

### Re-run Interview

```bash
# Back up current files, then re-interview
mv .moai/project/brand .moai/project/brand.backup
/moai design
```
