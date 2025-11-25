---
name: moai-cc-claude-md
description: Claude Code Markdown integration, documentation generation, and structured content patterns. Use when generating documentation, managing markdown content, or creating structured reports.
version: 1.0.0
modularized: false
tags:
  - enterprise
  - configuration
  - claude
  - claude-code
  - md
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0
**modularized**: false
**last_updated**: 2025-11-22
**compliance_score**: 75%
**auto_trigger_keywords**: cc, moai, claude, md


## Quick Reference (30 seconds)

Claude Code Markdown integration provides document generation, content structuring, and template-based documentation workflows.
It offers powerful documentation patterns for systematic management of project documentation (README, CHANGELOG),
technical documentation (API guides), knowledge bases, reports, and more.

**Core Features**:
- Automatic Markdown content generation and rendering
- Cross-reference and link management system
- Template-based document structuring
- Automatic content validation and quality checks
- Version control and change history tracking

---

## Implementation Guide

### What It Does

Claude Code Markdown integration provides:

**Markdown Content Generation**:
- AI-powered automatic document generation
- Code blocks and syntax highlighting
- Metadata and frontmatter management
- Dynamic content injection

**Document Structuring**:
- Hierarchical document organization
- Automatic table of contents generation
- Inter-section navigation
- Consistent formatting

**Template System**:
- Reusable document templates
- Variable substitution and conditional rendering
- Custom blocks and macros
- Style and theme application

### When to Use

- ✅ Project documentation (README, CONTRIBUTING, CODE_OF_CONDUCT)
- ✅ Technical documentation (API docs, development guides, tutorials)
- ✅ Process documentation (workflows, policies, procedures)
- ✅ Report generation (analysis, status reports, summaries)
- ✅ Knowledge base (FAQ, best practices, pattern libraries)
- ✅ Automated document deployment and publishing

### Core Markdown Patterns

#### 1. Document Structure Pattern
```markdown
# Title (Level 1)
## Subtitle (Level 2)
### Section (Level 3)

- Bullet points
  1. Numbered list
  2. Hierarchical structure

| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
```

#### 2. Cross-Reference Pattern
```markdown
[Link text](../path/to/file.md)
[Internal link](#section-title)
[External link](https://example.com)

[Variable reference]: variable-definition
```

#### 3. Code Block Pattern
````markdown
```python
# Python code example
def function():
    pass
```

```typescript
// TypeScript code example
interface Props {
  name: string;
}
```
````

#### 4. Content Validation Pattern
- Link validity checking
- Code block syntax validation
- Image path verification
- Metadata completeness validation

### Dependencies

- Markdown processing engine (Remark, Marked, Pandoc)
- Content template system
- Document validation framework
- Publishing platform (Nextra, VitePress, Docusaurus)

---

## Works Well With

- `moai-docs-generation` (automatic document generation)
- `moai-docs-validation` (content quality validation)
- `moai-docs-linting` (markdown style checking)
- `moai-cc-commands` (documentation workflow automation)

---

## Advanced Patterns

### 1. Advanced Template System

**Dynamic Content Injection**:
```markdown
<!-- Template Variable -->
{{ projectName }} - {{ version }}
{{ description }}

<!-- Conditional Content -->
{% if environment === 'production' %}
Production specific content
{% endif %}

<!-- Loop Patterns -->
{% for item in items %}
- {{ item.name }}
{% endfor %}
```

### 2. Automatic Documentation Generation Workflow

**Process**:
1. Parse source code/configuration files
2. Extract metadata (JSDoc, type definitions)
3. Merge template and metadata
4. Generate Markdown documentation
5. Automatic validation and deployment

**Example**:
```typescript
// Automatic API documentation generation from TypeScript code
/**
 * @description User creation function
 * @param {string} name - User name
 * @returns {Promise<User>} Created user object
 */
async function createUser(name: string): Promise<User> {
  // API documentation automatically generated
}
```

### 3. Multi-Channel Publishing Pattern

**Publishing Targets**:
- Markdown → HTML (website)
- Markdown → PDF (download)
- Markdown → Slides (presentation)
- Markdown → Email (distribution)
- Markdown → Wiki (organizational documentation)

### 4. Content Version Management

**Change History Tracking**:
- Git-based document version control
- Automatic CHANGELOG generation
- Migration guide provision
- Backward compatibility guarantee

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, markdown patterns
- **v1.0.0** (2025-10-22): Initial markdown integration

---

**End of Skill** | Updated 2025-11-21 | Lines: 180
