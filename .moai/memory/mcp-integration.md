# MCP Integration Guide

## Overview

This document defines the integration patterns for the three MCP servers used in MoAI-ADK: Context7, Playwright, and Figma Dev Mode.

## MCP Server Configuration

### Current MCP Servers

Based on `.mcp.json`, the system uses three MCP servers:

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"]
    },
    "playwright": {
      "command": "npx",
      "args": ["-y", "@playwright/mcp@latest"]
    },
    "figma-dev-mode-mcp-server": {
      "type": "sse",
      "url": "http://127.0.0.1:3845/sse"
    }
  }
}
```

## Context7 Integration

### Purpose

Context7 provides access to up-to-date documentation and library references, ensuring agents work with the latest API information.

### Capabilities

**Library Resolution**:
- Resolve library names to Context7-compatible IDs
- Access latest API documentation
- Version-specific documentation retrieval
- Code examples and best practices

**Documentation Access**:
- Official documentation from libraries
- API reference materials
- Implementation guides
- Troubleshooting resources

### Usage Patterns

**Library Resolution**:
```python
# Resolve library to Context7 ID
library_id = await mcp__context7__resolve_library_id("React")
# Returns: "/facebook/react/19.0.0"

# Get latest documentation
docs = await mcp__context7__get_library_docs(library_id)
```

**Agent Integration**:
```python
# Skill with Context7 integration
Skill("moai-lang-react")  # Load React best practices
# Combine with latest API docs
api_docs = await mcp__context7__get_library_docs("/facebook/react/19.0.0")

# Use in agent delegation
result = await Task(
    subagent_type="frontend-expert",
    prompt=f"Implement with latest React 19 features: {api_docs}",
    context={"latest_api": api_docs}
)
```

**Workflow Integration**:
```python
# SPEC creation with latest API knowledge
spec = await Task(
    subagent_type="spec-builder",
    prompt="Create specification for React application",
    context={
        "react_docs": await get_latest_docs("React"),
        "best_practices": Skill("moai-lang-react")
    }
)
```

### Agent Usage Patterns

**For Frontend Development**:
- `frontend-expert`: Always loads Context7 for current React/Vue/Angular versions
- `component-designer`: Uses latest component library documentation
- `ui-ux-expert`: Integrates latest UI framework best practices

**For Backend Development**:
- `backend-expert`: Uses latest framework documentation
- `api-designer`: Integrates current API design standards
- `database-expert`: Uses latest database documentation

## Playwright Integration

### Purpose

Playwright provides browser automation capabilities for E2E testing, UI validation, and web scraping.

### Capabilities

**Browser Automation**:
- Automated browser control
- Page navigation and interaction
- Form filling and submission
- Click and type operations

**Testing and Validation**:
- E2E test execution
- Visual regression testing
- Performance testing
- Accessibility testing

**Data Extraction**:
- Web scraping capabilities
- Screenshot capture
- Page content extraction
- Element text retrieval

### Usage Patterns

**E2E Testing**:
```python
# Start browser context
context = await mcp__playwright_create_context()

# Navigate to page
await mcp__playwright_goto(context, "https://example.com")

# Test functionality
await mcp__playwright_click(context, "button#submit")
await mcp__playwright_fill(context, "input#email", "test@example.com")

# Get page content
content = await mcp__playwright_get_page_content(context)
```

**Agent Integration**:
```python
# Test automation agent
test_result = await Task(
    subagent_type="test-engineer",
    prompt="Execute E2E tests for user authentication",
    context={
        "test_scenarios": authentication_scenarios,
        "browser_automation": "playwright_available"
    }
)

# UI validation agent
ui_validation = await Task(
    subagent_type="ui-ux-expert",
    prompt="Validate UI accessibility and responsiveness",
    context={
        "test_results": test_result,
        "accessibility_tests": "playwright_accessibility"
    }
)
```

### Agent Usage Patterns

**For Quality Assurance**:
- `test-engineer**: Executes comprehensive test suites
- `quality-gate`: Validates test coverage and quality metrics
- `ui-ux-expert`: Performs accessibility and usability testing

**For Development**:
- `frontend-expert`: Validates implementation with browser testing
- `component-designer`: Tests component behavior
- `performance-engineer`: Conducts performance testing

## Figma Dev Mode Integration

### Purpose

Figma Dev Mode provides design system integration, enabling seamless conversion from design to code and component library generation.

### Capabilities

**Design System Access**:
- Design file metadata retrieval
- Component extraction
- Style guide generation
- Design token export

**Design-to-Code Conversion**:
- Component code generation
- CSS extraction from designs
- Responsive design implementation
- Design system synchronization

**Collaboration Features**:
- Real-time design updates
- Version control integration
- Design review workflows
- Developer handoff automation

### Usage Patterns

**Design Access**:
```python
# Get design metadata
metadata = await mcp__figma_get_metadata("file_id")

# Extract variables
variables = await mcp__figma_get_variable_defs()

# Get screenshot
screenshot = await mcp__figma_get_screenshot("component_id")
```

**Agent Integration**:
```python
# Component design from Figma
component = await Task(
    subagent_type="component-designer",
    prompt="Create React components from Figma design",
    context={
        "design_metadata": await get_figma_design(),
        "style_variables": await get_figma_variables(),
        "component_specs": await extract_components()
    }
)

# UI/UX validation with Figma reference
ui_validation = await Task(
    subagent_type="ui-ux-expert",
    prompt="Validate implementation against Figma design",
    context={
        "figma_design": figma_specs,
        "implementation": frontend_code,
        "accessibility_standards": wcag_standards
    }
)
```

### Agent Usage Patterns

**For Design System Development**:
- `component-designer`: Creates components from Figma designs
- `ui-ux-expert`: Validates design implementation
- `frontend-expert`: Implements design-based components

**For Development**:
- `docs-manager`: Generates documentation from design specifications
- `quality-gate`: Ensures design consistency in implementation

## Integration Workflows

### Complete Design-to-Code Workflow

```python
# 1. Design Analysis (Figma)
design_analysis = await Task(
    subagent_type="ui-ux-expert",
    prompt="Analyze Figma design for implementation requirements",
    context={"figma_file": await get_figma_design()}
)

# 2. Component Creation (Figma + Design)
components = await Task(
    subagent_type="component-designer",
    prompt="Create React components from Figma specifications",
    context={
        "figma_design": await get_figma_design(),
        "analysis": design_analysis,
        "react_docs": await get_context7_docs("React")
    }
)

# 3. Implementation (Context7 + React)
implementation = await Task(
    subagent_type="frontend-expert",
    prompt="Implement components with latest React patterns",
    context={
        "components": components,
        "react_19_docs": await get_context7_docs("React", "19.0.0"),
        "best_practices": Skill("moai-lang-react")
    }
)

# 4. Testing (Playwright)
testing = await Task(
    subagent_type="test-engineer",
    prompt="Create comprehensive E2E tests for components",
    context={
        "implementation": implementation,
        "test_scenarios": await generate_test_cases()
    }
)

# 5. Validation (All MCP Servers)
validation = await Task(
    subagent_type="quality-gate",
    prompt="Validate complete implementation",
    context={
        "design_fidelity": await validate_against_figma(),
        "code_quality": await analyze_code_quality(),
        "test_results": testing,
        "accessibility": await run_accessibility_tests()
    }
)
```

### SPEC-to-Implementation Workflow

```python
# 1. SPEC Creation (Context7 for latest standards)
spec = await Task(
    subagent_type="spec-builder",
    prompt="Create specification with latest technology standards",
    context={
        "react_docs": await get_context7_docs("React"),
        "security_standards": await get_context7_docs("OWASP"),
        "api_standards": await get_context7_docs("OpenAPI")
    }
)

# 2. Architecture Design (Context7 + Domain Expertise)
architecture = await Task(
    subagent_type="api-designer",
    prompt="Design system architecture",
    context={
        "specification": spec,
        "microservices_docs": await get_context7_docs("Microservices"),
        "domain_patterns": Skill("moai-domain-microservices")
    }
)

# 3. Implementation (Context7 + Language Skills)
implementation = await Task(
    subagent_type="tdd-implementer",
    prompt="Implement following specification and architecture",
    context={
        "specification": spec,
        "architecture": architecture,
        "latest_patterns": await get_context7_docs("TDD"),
        "language_skills": [Skill("moai-lang-python"), Skill("moai-lang-typescript")]
    }
)

# 4. Testing (Playwright)
testing = await Task(
    subagent_type="test-engineer",
    prompt="Create comprehensive test suite",
    context={
        "implementation": implementation,
        "testing_frameworks": await get_context7_docs("Playwright"),
        "automation_patterns": Skill("moai-domain-testing")
    }
)

# 5. Documentation (All sources)
documentation = await Task(
    subagent_type="docs-manager",
    prompt="Generate comprehensive documentation",
    context={
        "implementation": implementation,
        "api_docs": await generate_api_docs(),
        "testing_docs": testing,
        "design_docs": await extract_from_figma()
    }
)
```

## Error Handling and Recovery

### MCP Server Connection Issues

**Connection Failure Recovery**:
```python
async def handle_mcp_connection(server_name):
    try:
        # Attempt connection
        if server_name == "context7":
            await test_context7_connection()
        elif server_name == "playwright":
            await test_playwright_connection()
        elif server_name == "figma":
            await test_figma_connection()
    except ConnectionError:
        # Fallback to alternative strategies
        await log_mcp_failure(server_name)
        await use_cached_information(server_name)
        await notify_user_of_limitation(server_name)
```

### Graceful Degradation

**When MCP Unavailable**:
- Use cached documentation when Context7 unavailable
- Use manual testing patterns when Playwright unavailable
- Use static design specifications when Figma unavailable
- Maintain workflow continuity with reduced capabilities

## Performance Optimization

### MCP Server Caching

**Context7 Caching**:
```python
# Cache library documentation
doc_cache = {}

async def get_cached_docs(library_name):
    if library_name not in doc_cache:
        library_id = await mcp__context7__resolve_library_id(library_name)
        doc_cache[library_name] = await mcp__context7__get_library_docs(library_id)
    return doc_cache[library_name]
```

**Figma Design Caching**:
```python
# Cache design metadata
design_cache = {}

async def get_cached_design(figma_id):
    if figma_id not in design_cache:
        design_cache[figma_id] = await mcp__figma_get_metadata(figma_id)
    return design_cache[figma_id]
```

### Batch Operations

**Batch Documentation Requests**:
```python
# Request multiple library docs efficiently
libraries = ["React", "Vue", "Angular"]
doc_requests = [
    get_cached_docs(lib) for lib in libraries
]
docs = await asyncio.gather(*doc_requests)
```

**Batch Design Extractions**:
```python
# Extract multiple components efficiently
component_ids = ["comp1", "comp2", "comp3"]
extractions = [
    mcp__figma_get_screenshot(comp_id) for comp_id in component_ids
]
screenshots = await asyncio.gather(*extractions)
```

## Security Considerations

### Access Control

**MCP Server Access**:
- Restrict Figma access to authorized design files
- Limit Playwright to approved testing environments
- Control Context7 access to approved documentation sources

**Data Protection**:
- Never cache sensitive design data
- Validate all Figma file access permissions
- Sanitize Playwright test data
- Protect Context7 API keys and credentials

### Audit Trail

**MCP Usage Logging**:
```python
{
    "timestamp": "2025-11-20T07:30:00Z",
    "mcp_server": "figma-dev-mode-mcp-server",
    "action": "get_metadata",
    "file_id": "figma_file_123",
    "user_agent": "ui-ux-expert",
    "success": true,
    "data_size": "2.3MB"
}
```

This integration guide ensures optimal use of MCP servers while maintaining security, performance, and reliability standards.