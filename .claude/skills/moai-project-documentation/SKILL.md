---
name: moai-project-documentation
description: Enhanced project documentation with AI-powered features. Modular guide for Product.md, Structure.md, Tech.md creation
version: 1.0.0
modularized: true
allowed-tools:
  - Read
last_updated: 2025-11-22
compliance_score: 85
auto_trigger_keywords:
  - documentation
  - project
category_tier: 1
---

## Quick Reference (30 seconds)

# moai-project-documentation

**Project Documentation**

> **Primary Agent**: alfred  
> **Secondary Agents**: none  
> **Version**: 4.0.0  
> **Keywords**: project, documentation, git, frontend, kubernetes

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

#### Purpose

Guide interactive creation of three core project documentation files (product.md, structure.md, tech.md) based on project type and user input. Provides templates, examples, checklists, and best practices for each project type (Web App, Mobile App, CLI Tool, Library, Data Science).

#### Core Modules

This skill is modularized for optimal loading:

**Module 1: Project Types & Product.md** (`SKILL-types.md`)
- Project type selection (Web App, Mobile App, CLI, Library, Data Science)
- Product.md writing guide by project type
- User personas and success metrics

**Module 2: Structure.md & Tech.md** (`SKILL-structure-tech.md`)
- System architecture patterns by project type
- Technology stack examples
- Quality gates and deployment strategies

**Module 3: Checklists & Examples** (`SKILL-checklists.md`)
- Writing checklists for all three documents
- Common mistakes to avoid
- Real-world examples by project type

---

### Level 2: Practical Implementation (Common Patterns)

#### Metadata

- **Name**: moai-project-documentation
- **Domain**: Project Documentation & Planning
- **Freedom Level**: high
- **Target Users**: Project owners, architects, tech leads
- **Invocation**: Skill("moai-project-documentation")
- **Progressive Disclosure**: Metadata → Content (full guide) → Resources (examples)

#### Usage Pattern

```
1. Identify project type (Web/Mobile/CLI/Library/DataScience)
2. Load SKILL-types.md for Product.md guidance
3. Load SKILL-structure-tech.md for architecture & tech stack
4. Load SKILL-checklists.md for validation & examples
5. Generate customized documentation
```

#### Quick Decision Tree

```
Start
  ├─ Web Application? → SKILL-types.md (Web App section)
  ├─ Mobile App? → SKILL-types.md (Mobile App section)
  ├─ CLI Tool? → SKILL-types.md (CLI section)
  ├─ Library/SDK? → SKILL-types.md (Library section)
  └─ Data Science? → SKILL-types.md (Data Science section)
```

---

### Level 3: Advanced Patterns (Expert Reference)

#### Best Practices Checklist

**Must-Have:**
- ✅ Project type clearly identified before document generation
- ✅ All three documents (Product.md, Structure.md, Tech.md) consistent
- ✅ HISTORY section initialized with v0.1.0
- ✅ Quality gates specific to project type

**Recommended:**
- ✅ Use type-specific templates (Web vs CLI vs Library)
- ✅ Include measurable success metrics
- ✅ Document architectural trade-offs
- ✅ Specify version ranges for dependencies

**Security:**
- 🔒 Never commit credentials in documentation
- 🔒 Document secret management strategy in Tech.md
- 🔒 Include security policy and incident response

---

## 📚 Official References

**Primary Documentation:**
- [SKILL-types.md](/moai-project-documentation/SKILL-types.md) – Project types & Product.md guide
- [SKILL-structure-tech.md](/moai-project-documentation/SKILL-structure-tech.md) – Architecture & tech stack patterns
- [SKILL-checklists.md](/moai-project-documentation/SKILL-checklists.md) – Validation checklists & examples

**Best Practices:**
- See SKILL-checklists.md for common mistakes and corrections
- Each project type has specific templates and examples

---

## 📈 Version History

**4.0.0** (2025-11-12)
- ✨ Modular structure with 3 sub-skills
- ✨ Enhanced Progressive Disclosure
- ✨ Project type-specific guidance
- ✨ Comprehensive checklists and examples
- ✨ Mobile application patterns added
- ✨ Data Science project templates

---

**Generated with**: MoAI-ADK Skill Factory    
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (alfred)

---

## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("moai-foundation-specs") – SPEC format understanding
- Skill("moai-docs-generation") – Documentation generation

**Complementary Skills:**
- Skill("moai-docs-unified") – Documentation standards
- Skill("moai-git-flow") – Version control integration

**Next Steps:**
- After documentation: Use Skill("moai-foundation-trust") for quality gates
- For deployment: Use Skill("moai-devops-kubernetes") or Skill("moai-devops-docker")

---

**End of Skill** | Updated 2025-11-12