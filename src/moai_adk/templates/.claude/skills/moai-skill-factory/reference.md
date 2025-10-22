# moai-skill-factory - CLI Reference

_Last updated: 2025-10-22_

## Quick Reference

### Skill Creation Workflow

```bash
# Step 1: Interactive Discovery
# Use moai-alfred-tui-survey for user questions

# Step 2: Web Research
# WebSearch: "latest [tool/framework] version 2025"
# WebFetch: Official documentation URLs

# Step 3: Generate Structure
mkdir -p .claude/skills/skill-name
touch .claude/skills/skill-name/{SKILL.md,examples.md,reference.md}

# Step 4: Dual Deployment
cp -r .claude/skills/skill-name src/moai_adk/templates/.claude/skills/
```

### Tool Versions (2025-10-22)

**Skill Factory Stack**:
- Claude Code: 0.6.0
- MoAI-ADK: 0.4.5+
- Python: 3.13+

**Research Tools**:
- WebSearch: Latest available
- WebFetch: Built-in
- Official Docs: Primary source

---

## Skill File Structure (v2.0.0)

```
skill-name/
├── SKILL.md                # Main instructions with v2.0.0 frontmatter
├── examples.md             # 3-5 working examples
├── reference.md            # CLI commands and tool versions
├── scripts/ (optional)     # Utility scripts
└── templates/ (optional)   # Reusable templates
```

### Required Frontmatter (v2.0.0)

```yaml
---
name: skill-name
version: 2.0.0
created: YYYY-MM-DD
updated: YYYY-MM-DD
status: active
description: "Comprehensive description with trigger keywords"
keywords: [keyword1, keyword2, keyword3, ...]
allowed-tools:
  - Read
  - Bash
---
```

---

_For detailed usage, see SKILL.md and EXAMPLES.md_
