---
name: moai-cc-skill-factory
version: 2.2.0
created: 2025-10-22
updated: 2025-11-05
status: active
description: Create and maintain high-quality Claude Code Skills through interactive discovery, web research, and comprehensive description writing standards. Use when building new Skills, researching latest best practices, updating existing Skills, writing skill descriptions with proper trigger keywords, or generating Skill packages backed by official documentation and real-world examples.
keywords: [skill-creation, claude-skills, best-practices, web-research, interactive-discovery, skill-updates, description-writing, metadata-standards]
allowed-tools:
  - Read
  - Write
  - Glob
  - Bash
  - WebFetch
  - WebSearch
---

# Skill Factory - High-Quality Skill Creation System

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 2.1.0 |
| Tier | CC (Claude Code Meta) |
| Auto-load | Proactively when building Skills |
| Purpose | Interactive discovery → Web research → Skill generation |

---

## What It Does

Interactive discovery, web research, continuous updates를 통한 고품질 Claude Code Skill 생성 및 유지관리 시스템을 제공합니다.

**Core capabilities**:
- ✅ Interactive user surveys via `moai-alfred-ask-user-questions`
- ✅ Web research for latest information (WebFetch/WebSearch)
- ✅ Skill analysis and update recommendations
- ✅ Official documentation validation
- ✅ Progressive Disclosure pattern
- ✅ Multi-model compatibility testing

---

## When to Use

- 새로운 Skill 패키지를 생성할 때
- 최신 best practices와 공식 문서를 조사할 때
- 기존 Skill을 분석하고 업데이트할 때
- Skill generation workflow를 orchestrate할 때

---

## Core Principles

### 1. Conciseness Over Completeness
- SKILL.md should stay **under 500 words**
- Trust that Claude is already intelligent—provide guidance, not lectures
- Reference patterns, provide decision trees, link to supporting files

### 2. Appropriate Freedom Levels

| Freedom Level | When to Use | Content Style |
| ------------- | ----------- | ------------- |
| **High** | Flexible, creative work | Principles, trade-offs, considerations |
| **Medium** | Standard patterns exist | Pseudocode, flowcharts, annotated code |
| **Low** | Deterministic, error-prone | Specific scripts with error handling |

### 3. Three-Level Progressive Disclosure

```
Level 1: Metadata (Always Active)
├── name, description, allowed-tools
└── ~100 tokens total
    ↓
Level 2: Instructions (On Demand)
├── SKILL.md main body
└── Loads when Claude recognizes relevance
    ↓
Level 3: Resources (As Needed)
├── reference.md, examples.md
└── Consumed when explicitly referenced
```

---

## Quick Start: Creating a Skill

### 5-Step Process (Total: ~2 hours)

1. **Define Problem** (10 min): What gap does this Skill fill?
2. **Design Metadata** (10 min): Name (gerund + domain), description, allowed-tools
3. **Structure Content** (30 min): High (20%) + Medium (50%) + Low (30%) freedom
4. **Add Examples & References** (30 min): 3-4 examples, detailed reference
5. **Validate & Test** (20 min): CHECKLIST.md, multi-model testing

---

## YAML Metadata Requirements

### Frontmatter Structure

```yaml
---
name: "Skill Name"
description: "What it does and when to use it"
allowed-tools: "Tool1, Tool2, Tool3"
---
```

### Field Specifications

**`name`** (Required):
- Max 64 characters
- Format: Gerund (action verb) + domain
- Example: "Processing CSV Files with Python"

**`description`** (Required):
- Max 1024 characters
- Format: Third person, action-oriented
- Must include: capabilities + trigger scenarios + 3+ keywords
- **Critical**: Use "Use when" pattern with 3-5 trigger keywords
- **For complete description writing guidelines**: See [Description Writing Standards](reference.md#description-writing-standards)

**`allowed-tools`** (Recommended):
- Minimal principle: Only tools actually used
- Format: Comma-separated list
- Example: `"Read, Write, Edit, Bash(git:*)"`

> **Detailed metadata guide**: See [METADATA.md](METADATA.md)

---

## File Organization

### Recommended Structure

```
skill-name/
├── SKILL.md                    # Main instructions (~500 words)
├── reference.md                # Detailed specs, API docs
├── examples.md                 # 3-4 real-world examples
├── CHECKLIST.md                # Quality validation (optional)
├── scripts/                    # Utility scripts
└── templates/                  # Reusable templates
```

### Rules

| Rule | Rationale |
|------|-----------|
| One level deep | Avoids nested discovery |
| Relative paths | Cross-platform compatibility |
| Forward slashes | Unix convention |
| <500 lines per file | Manageable context |

> **Detailed organization guide**: See [STRUCTURE.md](STRUCTURE.md)

---

## Quality Validation

Before publishing, audit your Skill:

### Metadata Completeness
- [ ] `name` ≤ 64 characters, gerund format
- [ ] `description` ≤ 1024 chars, includes 3+ triggers
- [ ] `allowed-tools` is minimal and justified

### Content Quality
- [ ] SKILL.md ≤ 500 words
- [ ] All major concepts have examples
- [ ] Terminology is consistent
- [ ] Progressive Disclosure applied

### Multi-Model Compatibility
- [ ] Haiku can understand concise examples
- [ ] Sonnet exploits full Skill capabilities
- [ ] Opus can extend patterns beyond examples

> **Complete validation checklist**: See [CHECKLIST.md](CHECKLIST.md)

---

## Advanced Features

### Interactive Discovery
Use TUI surveys to clarify vague requirements and capture all user needs.

> **Detailed guide**: [INTERACTIVE-DISCOVERY.md](INTERACTIVE-DISCOVERY.md)

### Web Research
Gather latest information, best practices, and official documentation.

> **Detailed guide**: [WEB-RESEARCH.md](WEB-RESEARCH.md)

### Skill Update Advisor
Analyze existing Skills and propose updates based on latest information.

> **Detailed guide**: [SKILL-UPDATE-ADVISOR.md](SKILL-UPDATE-ADVISOR.md)

---

## Common Failure Modes

| Issue | Root Cause | Fix |
|-------|------------|-----|
| **Not activating** | Description too generic | Add 5+ specific keywords |
| **Haiku ignores it** | Examples too complex | Simplify pseudocode |
| **Over-specifying** | Too much low-freedom content | Increase principles |
| **Scope creep** | Covers too many domains | Split into 2-3 focused Skills |
| **File too large** | SKILL.md > 500 words | Move content to reference.md |

---

## Integration

### With moai-alfred-skill-generator Sub-Agent

```
User: "/alfred:1-plan Create Skill for X"
    ↓
skill-generator Sub-Agent (ANALYZE → DESIGN → ASSURE phases)
    ↓
moai-skill-factory Skill (PRODUCE phase)
    ↓
Complete Skill package created
```

### Tools Used

- **WebFetch**: Fetch official documentation
- **WebSearch**: Search for latest best practices
- **Read/Glob**: Review existing Skills
- **Bash**: Directory creation, file operations

---

## Related Skills

- `moai-alfred-ask-user-questions`: Interactive user surveys (delegated)
- `moai-alfred-skill-generator`: Skill generation orchestrator

---

## Complete Documentation

**Core Guides**:
- [METADATA.md](METADATA.md) — Metadata authoring guide
- [STRUCTURE.md](STRUCTURE.md) — File organization patterns
- [CHECKLIST.md](CHECKLIST.md) — Pre-publication validation

**Advanced Features**:
- [INTERACTIVE-DISCOVERY.md](INTERACTIVE-DISCOVERY.md) — TUI survey patterns
- [WEB-RESEARCH.md](WEB-RESEARCH.md) — Web research strategies
- [SKILL-UPDATE-ADVISOR.md](SKILL-UPDATE-ADVISOR.md) — Skill analysis & updates

**Workflows**:
- [SKILL-FACTORY-WORKFLOW.md](SKILL-FACTORY-WORKFLOW.md) — Complete workflow
- [STEP-BY-STEP-GUIDE.md](STEP-BY-STEP-GUIDE.md) — Practical walkthrough

**References**:
- [reference.md](reference.md) — Detailed API reference
- [examples.md](examples.md) — Real-world case studies
- [EXAMPLES.md](EXAMPLES.md) — Additional examples

---

**Key Differentiator**: User-centric + research-driven + always-current + delegation-first orchestration

---

**End of Skill** | Refactored 2025-10-27
