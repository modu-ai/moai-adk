---
name: moai-specialized-docs
description: Documentation generation and management for comprehensive project documentation
version: 1.0.0
modularized: false
last_updated: 2025-11-24
allowed_tools: [Context7, Skill, Task]
compliance_score: 100%
category_tier: 6
auto_trigger_keywords: [docs, documentation, generation, markdown, api-docs, integration, automation, knowledge]
agent_coverage: [docs-manager, project-manager, test-engineer]
context7_references: [/mkdocs/mkdocs, /mitmwille/sphinx]
invocation_api_version: 1.0
dependencies: [moai-docs-generation, moai-foundation-trust]
deprecated: false
modules: null
successor: null
---

# moai-specialized-docs: Documentation Automation

## Quick Reference (Level 1)

**Documentation Generation**: Automated generation of API docs, changelogs, architecture diagrams, and knowledge bases from code.

**Key Capabilities**:
- API documentation generation
- Code example extraction
- Changelog auto-generation
- Diagram generation (Mermaid, PlantUML)
- Multi-version documentation
- Search index generation
- Static site deployment

**When to Use**: Building comprehensive documentation, maintaining API docs, automating changelog generation.

---

## Implementation Guide (Level 2)

### API Documentation Generation

```python
def generate_api_docs(openapi_spec: dict, output_dir: str):
    # Generate OpenAPI documentation
    docs = {
        'title': openapi_spec['info']['title'],
        'version': openapi_spec['info']['version'],
        'paths': {}
    }
    for path, methods in openapi_spec['paths'].items():
        docs['paths'][path] = {
            'description': methods.get('summary'),
            'methods': list(methods.keys())
        }
    return docs
```

### Changelog Generation

```python
def generate_changelog(commits: list) -> str:
    changelog = "# Changelog\n\n"
    for commit in commits:
        if commit['type'] in ['feat', 'fix', 'docs']:
            changelog += f"- [{commit['type']}] {commit['message']}\n"
    return changelog
```

### Code Example Extraction

```python
def extract_examples(source_path: str, language: str = 'python') -> list:
    examples = []
    # Find code blocks marked with # example tag
    for file in find_files(source_path, language):
        examples.extend(parse_examples(file))
    return examples
```

### Navigation Menu Generation

```python
def generate_toc(markdown_files: list) -> dict:
    toc = {'items': []}
    for file in sorted(markdown_files):
        toc['items'].append({
            'title': extract_title(file),
            'path': file,
            'children': extract_headings(file)
        })
    return toc
```

---

## Advanced Patterns (Level 3)

### Multi-Version Documentation

```python
def generate_versioned_docs(versions: list):
    for version in versions:
        output_dir = f"docs/{version}"
        generate_api_docs(get_openapi(version), output_dir)
        generate_changelog(get_commits(version))
```

### Search Index Generation

```python
def build_search_index(docs_dir: str) -> dict:
    index = {'documents': []}
    for md_file in find_markdown_files(docs_dir):
        doc = {
            'title': extract_title(md_file),
            'content': extract_content(md_file),
            'keywords': extract_keywords(md_file),
            'url': convert_to_url(md_file)
        }
        index['documents'].append(doc)
    return index
```

### Continuous Deployment

```python
def deploy_docs(docs_dir: str, target: str = 'vercel'):
    # Deploy to Vercel, Netlify, GitHub Pages, etc.
    run_build_command(f"mkdocs build -d {docs_dir}")
    deploy_to_platform(docs_dir, target)
```

---

## Documentation Security

### Access Control

```yaml
# mkdocs.yml with authentication
plugins:
  - search
  - auth:
      providers:
        - type: oauth2
          client_id: ${CLIENT_ID}
```

### Sensitive Information Protection
- No API keys in documentation
- Use placeholders: `API_KEY=your_key_here`
- Separate public/private docs
- Version control for confidential docs

### Compliance Patterns
- GDPR-compliant data handling documentation
- Security audit trails for doc changes
- Access logs for sensitive documentation

---

**Status**: Production Ready
**Best for**: Comprehensive API documentation, knowledge base automation, continuous documentation
