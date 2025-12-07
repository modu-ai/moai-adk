# Integration Patterns Module

Integration patterns for Context7 MCP, Plan Expander, and Document Processing.

## Pattern 1: Context7 MCP + WebSearch Parallel

**Purpose**: Fetch knowledge from multiple sources simultaneously

```python
# Direct integration with Context7 MCP tools
from mcp__context7__resolve_library_id import resolve_library_id
from mcp__context7__get_library_docs import get_library_docs

async def fetch_from_context7(topic: str) -> Dict:
    """Fetch from Context7 MCP."""
    
    # Resolve library ID
    library_id = resolve_library_id(topic)
    
    # Get documentation
    docs = get_library_docs(
        context7_compatible_library_id=library_id,
        topic=topic,
        tokens=5000
    )
    
    return {
        "source": "context7",
        "library_id": library_id,
        "content": docs,
        "relevance": 100  # Official documentation
    }

async def fetch_from_websearch(query: str) -> List[Dict]:
    """Fetch from WebSearch."""
    
    results = WebSearch(query=f"{query} best practices 2025")
    
    return [
        {
            "source": "websearch",
            "url": result['url'],
            "title": result['title'],
            "content": result['snippet'],
            "relevance": calculate_relevance(result, query)
        }
        for result in results
    ]
```

---

## Pattern 2: Plan Expander Integration

**Purpose**: Integrate with yoda-system templates

```python
# Integration with yoda-system
from moai_yoda_system import load_template

async def expand_with_template(
    plan: Dict,
    template_name: str,
    knowledge: Dict
) -> List[Section]:
    """Expand plan using Yoda System template."""
    
    # Load template
    template = load_template(template_name)  # education.md
    
    # Extract section structure from plan
    section_structure = extract_section_structure(plan)
    
    # Expand each section using template + knowledge
    expanded_sections = []
    
    for section in section_structure:
        expanded = apply_template(
            section=section,
            template=template,
            knowledge=knowledge[section['topic']]
        )
        expanded_sections.append(expanded)
    
    return expanded_sections
```

---

## Pattern 3: Document Processing Integration

**Purpose**: Convert generated content to multiple formats

```python
# Use external conversion tools (pandoc, etc.) or custom processors
async def convert_to_formats(
    markdown_content: str,
    output_formats: List[str] = ["pdf", "pptx", "docx"]
) -> Dict[str, bytes]:
    """Convert markdown to multiple formats using external tools."""

    results = {}

    for format in output_formats:
        # Use pandoc or similar tools for conversion
        if format == "pdf":
            converted = convert_markdown_to_pdf(markdown_content)
        elif format == "pptx":
            converted = convert_markdown_to_pptx(markdown_content)
        elif format == "docx":
            converted = convert_markdown_to_docx(markdown_content)

        results[format] = converted

    return results
```
