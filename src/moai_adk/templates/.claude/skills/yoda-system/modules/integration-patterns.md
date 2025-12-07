# Integration Patterns Module

Integration patterns for Context7 MCP, document conversion tools, and Notion publishing.

## Pattern 1: Context7 MCP Integration

**Purpose**: Automatically inject latest library documentation into templates

**When to use**:
- Creating content about libraries/frameworks (React, Python, etc.)
- Ensuring documentation stays current
- Adding official API references
- Referencing latest best practices

**Implementation Pattern**:

```javascript
// In your agent/command that populates templates:

// Step 1: Resolve library identifier
const libraryId = await mcp__context7__resolve-library-id("React");
// Returns: "/facebook/react"

// Step 2: Fetch documentation
const docs = await mcp__context7__get-library_docs("/facebook/react", {
  topic: "hooks",
  tokens: 5000  // Maximum documentation tokens
});

// Step 3: Inject into template
const content = education_template
  .replace("{{CONTEXT7_REACT_HOOKS}}", docs)
  .replace("{{LIBRARY_NAME}}", "React");

// Step 4: Save populated content
save_file(".moai/yoda/output/react-hooks-guide.md", content);
```

**Template Placeholder Format**:

```markdown
# In education.md:
[Official Documentation]: {{CONTEXT7_LIBRARY_TOPIC}}

# In presentation.md:
See [official {{LIBRARY_NAME}} documentation]({{CONTEXT7_DOCS_LINK}})

# In workshop.md:
Latest API Reference: {{CONTEXT7_API_REFERENCE}}
```

**Benefits**:
- Always uses latest documentation (no stale content)
- No manual updates needed
- Reduces documentation maintenance burden
- Improves content accuracy and authority
- Seamless integration with generation pipeline

**MCP Tool Integration**:
```python
from mcp__context7__resolve_library_id import resolve_library_id
from mcp__context7__get_library_docs import get_library_docs

# Automatic discovery
lib_id = resolve_library_id("Python")  # Returns "/python/docs"
docs = get_library_docs("/python/docs", topic="asyncio")
```

---

## Pattern 2: Document Conversion Integration

**Purpose**: Convert markdown templates to multiple output formats using external tools

**When to use**:
- Converting markdown to PDF
- Creating PowerPoint presentations from slides
- Generating Word documents
- Batch format conversion

**Implementation Pattern**:

```bash
# In your agent/command:

# Step 1: Populate template
# (Generate markdown_content from template)

# Step 2: Use pandoc for conversion
# Convert to PDF:
pandoc -s input.md -o output.pdf \
  --metadata title="Python Async Programming" \
  --metadata author="Alice"

# Convert to DOCX:
pandoc -s input.md -o output.docx \
  --reference-doc=template.docx

# Convert to PPTX (if pandoc supports):
pandoc -s input.md -o output.pptx \
  --slide-level=2

# Alternative: Use wkhtmltopdf for PDF
wkhtmltopdf input.html output.pdf
```

**Template Format Requirements**:

For each conversion type, ensure proper formatting:

**Markdown for PDF**:
```markdown
# Heading 1 (becomes page title)
## Heading 2 (becomes section)
```

**Markdown for PPTX** (presentation.md):
```markdown
---
slide: 1
title: "Slide Title"
layout: content
---

## Slide Content
[Content for conversion]
```

**Markdown for DOCX** (workshop.md):
```markdown
# Main Title
## Section Title
### Subsection
[Regular markdown with proper structure]
```

**Supported Conversions**:
- education.md → PDF ✓, DOCX ✓
- presentation.md → PPTX ✓, PDF ✓
- workshop.md → DOCX ✓, PDF ✓

**Recommended Conversion Tools**:
- **pandoc**: Universal document converter (recommended)
- **wkhtmltopdf**: HTML to PDF converter
- **LibreOffice**: CLI conversion for DOCX/PPTX
- **Marp**: Markdown to presentation (alternative)

**Benefits**:
- Leverages proven external tools
- Wide format support
- Professional output quality
- Multi-format support from single markdown
- Community-maintained and well-documented

---

## Pattern 3: Notion Publishing Integration (Optional)

**Purpose**: Automatically publish generated content to Notion

**When to use**:
- Publishing to team knowledge bases
- Creating public documentation sites
- Integrating with project management
- Sharing with non-technical stakeholders

**Implementation Pattern**:

```javascript
// In your agent/command (requires --notion-enhanced flag):

// Step 1: Generate content
const markdown_content = populate_template("education.md", {
  topic: "React Hooks",
  difficulty: "intermediate"
});

// Step 2: Convert to Notion format
const notion_blocks = markdown_to_notion_blocks(markdown_content);

// Step 3: Publish to Notion using moai-integration-mcp
const notion_result = await Skill("moai-integration-mcp").publish({
  database: "Lectures",
  blocks: notion_blocks,
  metadata: {
    title: "React Hooks Guide",
    instructor: "Alice",
    audience: "Mid-level developers",
    difficulty: "Intermediate",
    tags: ["React", "Hooks", "Frontend"],
    category: "Education"
  },
  publish: true  // Make public
});

// Step 4: Save link for reference
save_file(".moai/yoda/output/react-hooks-notion-link.txt",
  notion_result.public_url);
```

**Notion Database Schema**:

```yaml
Properties:
  - title: Text "React Hooks Guide"
  - instructor: Person "Alice"
  - audience: Select "Mid-level Developers"
  - difficulty: Select "Intermediate"
  - category: Select "Education"
  - tags: Multi-select ["React", "Hooks"]
  - created_date: Date (auto-filled)
  - published_url: URL (auto-filled)
  - cover: File (optional preview image)
```

**Benefits**:
- Automatic knowledge base updates
- Team collaboration on documentation
- Version control through Notion
- Public sharing capability
- SEO-friendly documentation sites

**Prerequisites**:
- Notion MCP configured in .claude/mcp.json
- Notion API key configured
- Target database created
- Proper access permissions
