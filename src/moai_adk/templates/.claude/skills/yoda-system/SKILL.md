---
name: yoda-system
version: 1.0.0
category: education
status: active
description: Educational lecture material generation templates and best practices with Context7 MCP integration. Three battle-tested templates for education, presentation, and workshop formats with automatic Context7 documentation injection
allowed-tools: Read, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
created: 2025-11-14
updated: 2025-11-27
tags: education, templates, lectures, MCP, Context7
primary-agents: yoda-master, yoda-book-author
dependencies: moai-integration-mcp, yoda-writing-templates
---

# Yoda System: Educational Content Generation Skill

> Master Yoda's wisdom: "바퀴를 재발명하지 말고, 기존의 도구를 현명하게 재사용하자"
> (Don't reinvent the wheel; reuse existing tools wisely)

## What is Yoda System?

The Yoda System provides **three battle-tested templates** for generating educational lecture materials across different formats and audiences. Instead of creating new conversion tools or duplicating functionality, it focuses on template design and integration patterns with Context7 MCP for latest documentation.

### Purpose

- Generate high-quality educational materials automatically
- Support multiple learning styles (theory, visual, hands-on)
- Integrate with latest documentation via Context7 MCP
- Leverage existing document conversion tools
- Maintain backward compatibility with content generation workflows

### What You Get

- **3 customizable templates** (education/presentation/workshop)
  - education.md: 464 lines, theory-focused learning guide
  - presentation.md: 762 lines, 18-slide visual presentation
  - workshop.md: 928 lines, complete hands-on training

- **Integration patterns** for:
  - Context7 MCP (automatic documentation injection)
  - Document conversion tools (external tools like pandoc)
  - Notion MCP (optional publishing via moai-integration-mcp)

- **Best practices** for:
  - Template selection (decision tree)
  - Difficulty level adaptation
  - Audience customization
  - Educational content structure

- **Supporting documentation**:
  - Usage examples (3 real-world scenarios)
  - Troubleshooting guide
  - Quality validation checklist

### Quick Start

```bash
# Load Yoda System Skill
Skill("yoda-system")

# Access templates:
#   - templates/education.md      (📚 Theory-focused learning)
#   - templates/presentation.md   (🎤 Visual presentation)
#   - templates/workshop.md       (🔧 Hands-on practice)

# See examples below for detailed usage patterns
```

---

## Template Selection Guide

### Decision Tree: Which Template?

**Step 1: What is your primary goal?**

```
Decision Point:
├─ Provide sequential learning path?           → education.md ✓
├─ Create visual presentation materials?       → presentation.md ✓
└─ Enable hands-on practical skills?           → workshop.md ✓
```

**Step 2: Who is your audience?**

```
Audience Analysis:
├─ Beginner (< 1 year experience):
│   ├─ Theory preferred?                       → education.md
│   ├─ Guided practice preferred?              → workshop.md
│   └─ Overview presentation?                  → presentation.md
│
├─ Intermediate (1-3 years experience):
│   ├─ Deepening theory knowledge?             → education.md
│   ├─ Real-world case studies?                → presentation.md
│   └─ Hands-on advanced techniques?           → workshop.md
│
└─ Advanced (3+ years experience):
    ├─ Specialized deep-dive?                  → education.md
    ├─ Architecture & design patterns?         → presentation.md
    └─ Performance optimization workshop?      → workshop.md
```

**Step 3: What is the content type?**

```
Content Categorization:
├─ Conceptual/Theory:           education.md ✓
├─ Practical/Hands-on:          workshop.md ✓
├─ Motivational/Overview:       presentation.md ✓
├─ Mixed (theory + practice):   education.md + workshop.md
└─ Mixed (overview + details):  presentation.md + education.md
```

---

## Template Details

For complete template specifications, see:
- [Template Details Module](modules/template-details.md) - Comprehensive documentation for all 3 templates
- [Integration Patterns Module](modules/integration-patterns.md) - Context7, document conversion, Notion integration
- [Examples](examples.md) - Real-world usage scenarios

### Quick Reference

**education.md** (464 lines)
- When: Sequential learning, theory-focused content
- Best for: Online courses, self-paced training
- Output: MD, PDF, DOCX

**presentation.md** (762 lines)
- When: Visual presentation, conference talks
- Best for: Public speaking, executive summaries
- Output: MD, PPTX, PDF

**workshop.md** (928 lines)
- When: Hands-on practice, intensive training
- Best for: Bootcamps, practical skill development
- Output: MD, DOCX, PDF

---

## Integration Patterns

See [Integration Patterns Module](modules/integration-patterns.md) for:
- Context7 MCP integration (latest documentation)
- Document conversion integration (format conversion using external tools)
- Notion publishing integration (knowledge base via moai-integration-mcp)

---

## Best Practices

See [Best Practices Module](modules/best-practices.md) for:
- Template selection strategy
- Difficulty level adaptation
- Audience-specific customization
- Context7 MCP integration strategy
- Quality validation checklist

---

## Usage Examples

See [Examples](examples.md) for:
- Example 1: Generate Theory-Focused Education Material
- Example 2: Generate Presentation with Context7
- Example 3: Generate Hands-On Workshop with Notion

---

## Troubleshooting

### Issue: Template Not Found

**Cause**: Template file path incorrect or skill not loaded

**Solution**:
```bash
# Verify templates exist
ls -la .claude/skills/yoda-system/templates/

# Expected output:
# - education.md (464 lines)
# - presentation.md (762 lines)
# - workshop.md (928 lines)

# Verify skill loads
Skill("yoda-system")
```

**Diagnosis**:
- Check file permissions (should be readable)
- Verify skill is in correct directory
- Check that files are not corrupted

---

### Issue: Context7 MCP Connection Failed

**Cause**: MCP server not configured, running, or authentication failed

**Solution**:
```bash
# Check MCP status
claude mcp serve

# Verify MCP configuration
cat .claude/mcp.json | jq '.mcpServers.context7'

# Test MCP connection
# (If error: skill continues without MCP, but misses documentation injection)

# Fallback: Use template without injected documentation
# - Template still works completely
# - Just less comprehensive than with Context7
# - Can be updated manually later
```

**Recovery Strategy**:
- MCP is optional, not required
- Templates work standalone
- Generate content, update docs manually if needed
- Set up MCP later for future generations

---

### Issue: PDF Conversion Failed

**Cause**: External conversion tools not installed or misconfigured

**Solution**:
```bash
# Install conversion tools
# For pandoc (recommended):
brew install pandoc  # macOS
apt install pandoc   # Linux

# For wkhtmltopdf (alternative):
brew install wkhtmltopdf  # macOS
apt install wkhtmltopdf   # Linux

# Fallback: Use markdown output only
/yoda:generate --topic "..." --output "md"

# Can convert to PDF later using installed tools
```

**Workaround**:
- Generate markdown version
- Use external PDF tools (pandoc, wkhtmltopdf)
- Share markdown directly to team

---

## See Also

- **External Conversion Tools**: pandoc, wkhtmltopdf for PDF/DOCX/PPTX conversion
- **Context7 MCP**: Latest library documentation retrieval
- **moai-integration-mcp**: Notion knowledge base automation
- **/yoda:generate Command**: Main entry point
- **yoda-master.md Agent**: Orchestration and decision tree

---

## Related Skills

- `moai-integration-mcp`: Notion database publishing and automation
- `yoda-writing-templates`: Additional writing templates and patterns

---

## Version History

**v1.0.0** (2025-11-14) - Initial Release
- 3 complete templates (education, presentation, workshop)
- Context7 MCP integration pattern
- Document conversion integration with external tools
- Notion publishing support via moai-integration-mcp
- Comprehensive best practices guide
- 3+ real-world usage examples
- Complete troubleshooting documentation

**Template Details**:
- education.md: 464 lines, 13 code blocks, progressive disclosure
- presentation.md: 762 lines, 18 slides, multiple layouts
- workshop.md: 928 lines, 53 code blocks, 2 labs + team project

---

## Master Yoda's Wisdom

> "바퀴를 재발명하지 말고, 기존의 도구를 현명하게 재사용하자"
> (Don't reinvent the wheel; reuse existing tools wisely)

This skill embodies that principle:

- **Leverages** Context7 MCP for latest documentation
- **Uses** external tools (pandoc, etc.) for format conversions
- **Builds on** proven template structures
- **Scales through** composition, not duplication
- **Maintains** focus on education, not infrastructure

The Yoda System is a **integration orchestrator**, not a **tool builder**.

---

**Skill Summary**:
- **Name**: yoda-system
- **Version**: 1.0.0
- **Status**: Production Ready
- **Type**: Template Skill with MCP Integration
- **Last Updated**: 2025-11-14
- **Core Philosophy**: Reuse, Composition, Scale
- **Maintenance**: Minimal (templates are content, not code)
