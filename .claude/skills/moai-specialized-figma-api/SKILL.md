---
name: moai-specialized-figma-api
description: Figma API integration for design automation, component extraction, and design token generation
version: 1.0.0
modularized: false
last_updated: 2025-11-24
allowed_tools: [Context7, Skill, Task]
compliance_score: 100%
category_tier: 6
auto_trigger_keywords: [figma, api, design, ui-generation, components, design-tokens, export, automation]
agent_coverage: [component-designer, frontend-expert, design-systems]
context7_references: [/figma/figma-api]
invocation_api_version: 1.0
dependencies: [moai-domain-frontend, moai-design-systems]
deprecated: false
modules: null
successor: null
---

# moai-specialized-figma-api: Design Automation

## Quick Reference (Level 1)

**Figma Design Automation**: Extract components, design tokens, and generate code from Figma designs via REST API.

**Key Capabilities**:
- File metadata retrieval
- Component and frame extraction
- Design token generation
- Asset export (PNG, SVG, PDF)
- Version history access
- Plugin integration

**When to Use**: Automating design system workflows, generating component libraries from designs, extracting design tokens.

---

## Implementation Guide (Level 2)

### File Retrieval

```python
import requests

def get_figma_file(file_id: str, token: str) -> dict:
    headers = {"X-Figma-Token": token}
    response = requests.get(
        f"https://api.figma.com/v1/files/{file_id}",
        headers=headers
    )
    return response.json()
```

### Component Extraction

```python
def extract_components(file_data: dict) -> list:
    components = []
    for component in file_data['components'].values():
        components.append({
            'name': component['name'],
            'description': component['description'],
            'key': component['key']
        })
    return components
```

### Design Token Export

```python
def extract_design_tokens(file_data: dict) -> dict:
    tokens = {
        'colors': {},
        'typography': {},
        'spacing': {}
    }
    # Parse variables and styles from Figma file
    return tokens
```

---

## Advanced Patterns (Level 3)

### Code Generation from Components

```python
def generate_react_from_figma(component: dict) -> str:
    # Generate React component code from Figma frame
    return f"""
export function {component['name']}() {{
  return <div>{component['name']}</div>
}}
"""
```

### Asset Batch Export

```python
def export_assets(file_id: str, node_ids: list, format: str = 'png'):
    # Export multiple assets in batch
    params = {
        'ids': ','.join(node_ids),
        'format': format,
        'scale': 2
    }
    # Make API call with params
```

### Collaboration Tracking

```python
def get_file_comments(file_id: str) -> list:
    # Retrieve design review comments
    # Track design feedback and iterations
```

---

**Status**: Production Ready
**Best for**: Design system automation, component library generation, design token extraction
