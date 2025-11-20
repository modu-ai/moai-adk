# MoAI-ADK Memory Library

## Overview

This directory contains the complete reference library for MoAI-ADK agents, skills, commands, and execution patterns. All documents are written in English as AI agent instructions.

## Memory Structure

### Core Reference Documents

| Document | Purpose | Size |
|----------|---------|------|
| [agents.md](agents.md) | 35 specialized agents reference | 8.9KB |
| [skills.md](skills.md) | 135 skills catalog and usage | 11.9KB |
| [commands.md](commands.md) | 6 slash commands execution patterns | 5.9KB |
| [delegation-patterns.md](delegation-patterns.md) | Agent delegation patterns and workflows | 7.4KB |
| [token-optimization.md](token-optimization.md) | Token efficiency and context management | 8.0KB |
| [execution-rules.md](execution-rules.md) | Security rules and execution constraints | 10.5KB |
| [mcp-integration.md](mcp-integration.md) | 3 MCP servers integration guide | 16.8KB |

### MCP Server Configuration

Based on `.mcp.json`, the system uses three MCP servers:

1. **Context7** (`@upstash/context7-mcp`)
   - Library documentation and API reference
   - Latest version compatibility checking
   - Code examples and best practices

2. **Playwright** (`@playwright/mcp`)
   - Browser automation and E2E testing
   - UI validation and accessibility testing
   - Screenshot capture and page content extraction

3. **Figma Dev Mode** (SSE connection to `http://127.0.0.1:3845/sse`)
   - Design system integration
   - Component extraction and code generation
   - Design-to-code conversion

## Usage in CLAUDE.md

Import these documents in CLAUDE.md using the following pattern:

```markdown
## Reference Documentation

Agent selection: `@.moai/memory/agents.md`
Skill catalog: `@.moai/memory/skills.md`
Command patterns: `@.moai/memory/commands.md`
Delegation workflows: `@.moai/memory/delegation-patterns.md`
Token optimization: `@.moai/memory/token-optimization.md`
Execution rules: `@.moai/memory/execution-rules.md`
MCP integration: `@.moai/memory/mcp-integration.md`
```

## Document Relationships

```
CLAUDE.md (Main Instructions)
├── agents.md (Agent Selection)
├── skills.md (Knowledge Loading)
├── commands.md (Command Execution)
├── delegation-patterns.md (Workflow Orchestration)
├── token-optimization.md (Efficiency Management)
├── execution-rules.md (Security & Compliance)
└── mcp-integration.md (Extended Capabilities)
```

## Key Principles

### English-Only Instructions
All documents are written in English as direct instructions for AI agents. No code examples, @ symbols, or graphics are included.

### No Redundancy
Each document serves a specific purpose with no overlapping content. Cross-references are used instead of duplication.

### Agent-Centric Focus
All content is designed for AI agent consumption, providing clear, actionable instructions.

### Process-Oriented
Complex procedures are described as step-by-step processes rather than code examples.

## Maintenance

### Updates
- Update agent references when new agents are added
- Refresh skill catalog when new skills are created
- Update command patterns when workflows change
- Refresh MCP integration when servers are modified

### Validation
- Validate all import paths in CLAUDE.md
- Test agent delegation patterns
- Verify MCP server connections
- Check cross-reference consistency

### Version Control
- Track document versions in headers
- Maintain change logs for major updates
- Use semantic versioning for breaking changes
- Tag releases with corresponding CLAUDE.md versions

## Integration Testing

### Basic Functionality
```python
# Test import functionality
def test_memory_imports():
    documents = [
        "agents.md", "skills.md", "commands.md",
        "delegation-patterns.md", "token-optimization.md",
        "execution-rules.md", "mcp-integration.md"
    ]

    for doc in documents:
        assert validate_memory_document(doc), f"Invalid document: {doc}"
```

### Cross-Reference Validation
```python
# Test cross-reference consistency
def test_cross_references():
    # Verify all imports in CLAUDE.md exist
    # Check for broken references
    # Validate content consistency
    pass
```

### Performance Validation
```python
# Test document loading performance
def test_performance():
    for doc in memory_documents:
        load_time = measure_load_time(doc)
        assert load_time < 1.0, f"Slow loading: {doc}"
```

## Best Practices

### Document Creation
- Use clear, concise English
- Provide specific, actionable instructions
- Include agent selection criteria
- Define success criteria

### Document Maintenance
- Keep content focused and specific
- Regular validation of cross-references
- Performance monitoring of document access
- Version control of all changes

### Integration
- Test all import patterns in CLAUDE.md
- Validate agent delegation workflows
- Verify MCP server integration
- Monitor token usage efficiency

## Troubleshooting

### Common Issues

**Missing Documents**:
- Verify file exists in `.moai/memory/`
- Check file permissions
- Validate document format

**Import Failures**:
- Check file path accuracy
- Validate cross-reference syntax
- Test document accessibility

**Performance Issues**:
- Monitor document loading times
- Optimize large document structure
- Implement caching strategies

**Content Inconsistency**:
- Validate cross-references
- Check for conflicting information
- Update related documents

### Recovery Procedures

**Document Corruption**:
- Restore from version control
- Validate document integrity
- Test functionality

**Integration Failures**:
- Restart MCP servers
- Validate configuration files
- Test individual components

This memory library provides the complete knowledge base for MoAI-ADK agents, ensuring consistent, efficient, and reliable operation.