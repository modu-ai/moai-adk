# moai-content-markdown-best-practices

Writing high-quality Markdown content with consistent formatting, accessibility, and readability.

## Quick Start

Markdown is the standard for technical writing. Use this skill when writing blog posts, documentation, or content for platforms that support Markdown.

## Core Patterns

### Pattern 1: Markdown Structure

```markdown
# Main Title (H1)
Use only one H1 per document.

## Section Title (H2)
Main sections of your content.

### Subsection (H3)
Breakdown of H2 section.

#### Minor Point (H4)
Use sparingly for detailed breakdown.

## Lists and Organization

### Unordered Lists
- Item 1
- Item 2
  - Nested item 2.1
  - Nested item 2.2
- Item 3

### Ordered Lists
1. First step
2. Second step
3. Third step

### Definition Lists
**Term 1**: Definition and explanation.
**Term 2**: Definition and explanation.

## Code and Formatting

### Inline Code
Use \`code()\` for variable names and functions.

### Code Blocks
\`\`\`language
code here
\`\`\`

### Emphasis
- *Italic* for emphasis
- **Bold** for strong emphasis
- ***Bold italic*** for strongest emphasis

### Blockquotes
> "This is an important quote"
> — Author Name

## Tables

| Feature | Yes | No |
|---------|-----|-----|
| Feature A | ✅ | ❌ |
| Feature B | ✅ | ❌ |

## Links and References

[Link text](https://example.com)
[Link with title](https://example.com "Title")

### Reference links (for cleaner markdown)
[Link text][reference]

[reference]: https://example.com
```

### Pattern 2: Readability Guidelines

```markdown
## Good Markdown Practices

### Line Length
Aim for 80-100 characters per line for readability.

### Whitespace
Use blank lines between sections.
Use consistent indentation (2 or 4 spaces).
Avoid walls of text.

### Headings
- Use heading hierarchy: H1 → H2 → H3
- No skipping levels (don't go H1 → H3)
- Keep headings descriptive but concise
- Use imperative tone when possible

### Lists
- Keep list items parallel in structure
- Use 2-4 items per list ideally
- Don't exceed 1-2 sub-levels
- Use numbers for sequential, bullets for non-sequential

### Code Examples
\`\`\`python
# Always specify language for syntax highlighting
# Use complete, runnable examples
# Add comments for non-obvious code
def example():
    return "Clear example"
\`\`\`

### Emphasis Usage
- **Bold**: Important terms, key concepts
- *Italic*: Emphasis, citations, variable names
- \`Code\`: Function names, commands, paths
```

### Pattern 3: Accessibility

```markdown
## Accessibility Best Practices

### Image Alt Text
![Description of image content](image.jpg)

### Table Headers
| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Data | Data | Data |

### Link Text
✅ Good: [Click here to read our guide on Node.js](link)
❌ Bad: [Click here](link)

### Color-Independent Info
❌ Bad: "The red button on the right"
✅ Good: "The 'Submit' button on the right"

### Heading Hierarchy
✅ Good:
# Main Title
## Section
### Subsection

❌ Bad:
# Title
### Subsection (skips H2)

### Lists Over Paragraphs
Use lists for:
- Sequential instructions
- Multiple related items
- Bullet points and guidelines
```

**When to use**:
- Writing technical content
- Creating documentation
- Publishing blog posts
- Building knowledge bases

**Key benefits**:
- Professional appearance
- Easy to convert to HTML
- Accessible to all readers
- Consistent formatting

## Progressive Disclosure

### Level 1: Basic Markdown
- Headings and paragraphs
- Bold and italic
- Basic lists
- Code blocks

### Level 2: Advanced Markdown
- Tables
- Reference links
- Complex lists
- Blockquotes

### Level 3: Expert Practices
- Custom HTML/CSS
- Nested lists
- Advanced formatting
- Markdown extensions

## Works Well With

- **GitHub**: Markdown rendering
- **Static Generators**: Hugo, Jekyll, Next.js
- **Documentation**: ReadTheDocs, GitBook
- **Blogs**: Medium, Dev.to, Hashnode
- **Note-taking**: Obsidian, Notion

## References

- **CommonMark Spec**: https://spec.commonmark.org/
- **GitHub Markdown**: https://guides.github.com/features/mastering-markdown/
- **Markdown Guide**: https://www.markdownguide.org/
