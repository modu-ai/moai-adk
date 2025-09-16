# Claude Code SDK Documentation

## Overview

The Claude Code SDK provides programmatic access to Claude Code's capabilities, enabling developers to build AI agents with advanced tool access and conversation management. The SDK offers the same powerful features as the interactive CLI in a programmable interface, now enhanced with 2025's latest capabilities including Claude 4.0/4.1 models, enhanced permission systems, and advanced MCP integration.

### Core Capabilities

- **Optimized Claude Integration**: Built-in prompt caching and performance optimizations with Claude 4.0 and 4.1 support
- **Rich Tool Ecosystem**: Access to file operations, code execution, web search, MCP integration, and specialized sub-agents
- **Advanced Permissions**: Fine-grained control with new "ask" mode and enhanced security options
- **Production Essentials**: Improved error handling, session management, cost tracking, and monitoring
- **Multi-Language Support**: TypeScript, Python, and CLI interfaces with comprehensive 2025 feature support

### Available SDK Variants

- **[SDK Overview](sdk-overview.md)**: Comprehensive architecture and usage guide with 2025 features
- **[TypeScript SDK](sdk-typescript.md)**: Native Node.js and web applications with full type safety
- **[Python SDK](sdk-python.md)**: Python applications and data science workflows with async patterns
- **[Headless CLI](sdk-headless.md)**: Command-line automation and scripting with enhanced output formats

## Quick Start

### Installation

#### Prerequisites

- Node.js 18+ (for all SDK variants)
- Python 3.10+ (for Python SDK only)

#### TypeScript/JavaScript

```bash
npm install -g @anthropic-ai/claude-code
```

#### Python

```bash
pip install claude-code-sdk
npm install -g @anthropic-ai/claude-code
```

#### CLI Only

```bash
# Native installer (recommended)
curl -fsSL claude.ai/install.sh | bash

# Or via NPM
npm install -g @anthropic-ai/claude-code
```

### Authentication

Set up authentication before using any SDK:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

**Alternative Providers (2025):**

```bash
# Amazon Bedrock with Vertex AI regions
export CLAUDE_CODE_USE_BEDROCK=1
export VERTEX_REGION_CLAUDE_4_0_OPUS="us-east-1"

# Google Vertex AI with enhanced region support
export CLAUDE_CODE_USE_VERTEX=1
export VERTEX_REGION_CLAUDE_4_1_OPUS="us-central1"
```

## Basic Usage Examples

### TypeScript: Code Review Agent

```typescript
import { query } from '@anthropic-ai/claude-code';

for await (const message of query({
  prompt: 'Use the security-reviewer agent to audit this authentication module',
  options: {
    systemPrompt: 'You coordinate specialized security reviews',
    maxTurns: 5,
    allowedTools: ['Read', 'Grep', 'Task'],
    permissionMode: 'ask', // New 2025 interactive permission mode
    model: 'claude-4-1-opus', // Latest model support
  },
})) {
  if (message.type === 'result') {
    console.log(message.result);
  }
}
```

### Python: Data Analysis Workflow

```python
import asyncio
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

async def analyze_data():
    async with ClaudeSDKClient(
        options=ClaudeCodeOptions(
            system_prompt="You are a data scientist with MCP database access",
            max_turns=10,
            allowed_tools=["Read", "Write", "Bash", "mcp__database__query"],
            permission_mode="ask",  # Enhanced permission control
            model="claude-4-1-opus",
            max_mcp_output_tokens=50000  # New MCP token management
        )
    ) as client:
        await client.query("Analyze sales_data.csv and generate insights using database context")

        async for message in client.receive_response():
            if hasattr(message, 'content'):
                for block in message.content:
                    if hasattr(block, 'text'):
                        print(block.text)

asyncio.run(analyze_data())
```

### CLI: Automated Testing

```bash
# One-shot query with specific tools and new permission mode
claude -p "Run tests and fix any failures using the test-automator agent" \
  --allowedTools "Bash,Read,Edit,Task" \
  --permission-mode ask \
  --model claude-4-1-opus \
  --max-turns 10

# JSON output for scripting with enhanced metadata
claude -p "analyze code complexity and generate report" \
  --output-format json \
  --allowedTools "Read,Grep,Glob" \
  --model claude-4-1-opus
```

## Key Features

### 2025 Enhanced Tool Ecosystem

Claude Code SDK provides access to powerful tools with 2025 enhancements:

| Tool          | Description                  | Permission Required | 2025 Enhancement |
| ------------- | ---------------------------- | ------------------- | ---------------- |
| **Bash**      | Execute shell commands       | Yes                 | Background execution, enhanced safety |
| **Read**      | View file contents           | No                  | Image support, encoding detection |
| **Write**     | Create new files             | Yes                 | Validation, encoding optimization |
| **Edit**      | Targeted file modifications  | Yes                 | Rollback capability, atomic edits |
| **MultiEdit** | Multiple atomic edits        | Yes                 | Enhanced conflict resolution |
| **Glob**      | Pattern-based file discovery | No                  | Performance optimizations |
| **Grep**      | Content search across files  | No                  | Advanced regex support |
| **WebFetch**  | Retrieve web content         | Yes                 | Caching, rate limiting |
| **WebSearch** | Intelligent web searches     | Yes                 | Domain filtering, ranking |
| **Task**      | Delegate to sub-agents       | No                  | Parallel execution, context isolation |
| **MCP Tools** | External integrations        | Yes                 | Token management, API features |

### Enhanced Output Formats (2025)

- **Text**: Human-readable output with improved formatting
- **JSON**: Structured data with enhanced metadata and cost tracking
- **Streaming JSON**: Real-time incremental updates with progress indicators

### Advanced Session Management

- Resume conversations by session ID with enhanced state preservation
- Multi-turn context preservation with cost optimization
- Conversation compaction for long sessions with intelligent summarization
- Session lifecycle events (SessionStart, SessionEnd) for monitoring

### Enhanced Permission Modes (2025)

| Mode | Description | Use Case | 2025 Enhancement |
|------|-------------|----------|------------------|
| `ask` | Request approval for each tool use | Interactive development | New mode with rich prompts |
| `acceptEdits` | Auto-approve file operations, ask for others | Safe development workflow | Enhanced file validation |
| `acceptAll` | Approve all tool uses automatically | Full automation | Audit logging |
| `bypassPermissions` | Skip permission system | Advanced automation | Security warnings |

## Use Cases

### Development Automation (Enhanced 2025)

- **Code Review**: Automated security and quality analysis with specialized sub-agents
- **Testing**: Intelligent test generation and failure fixing using test-automator agent
- **Documentation**: API reference and usage guide generation with enhanced templates
- **Refactoring**: Large-scale code improvements with rollback capabilities

### Data and Analytics

- **Data Analysis**: CSV/JSON processing with MCP database integration
- **Report Generation**: Automated insights and summaries with visual components
- **Model Training**: ML pipeline automation with progress tracking
- **Database Operations**: Query optimization and analysis with enhanced MCP tools

### DevOps and Infrastructure (2025 Enhancements)

- **CI/CD Integration**: Automated pipeline management with GitHub Actions integration
- **Incident Response**: Log analysis and troubleshooting with specialized monitoring agents
- **Monitoring**: Alert analysis and resolution with MCP monitoring tools
- **Deployment**: Automated release management with rollback capabilities

### Business Applications

- **Content Creation**: Documentation and marketing materials with enhanced templates
- **Process Automation**: Workflow optimization with specialized workflow agents
- **Customer Support**: Automated response generation with context preservation
- **Legal and Compliance**: Contract review and analysis with audit logging

## Advanced Features (2025)

### MCP Integration Enhancements

Connect to external tools and data sources with enhanced capabilities:

```typescript
// MCP servers with token management and API integration
const options = {
  allowedTools: ['Read', 'mcp__github__search', 'mcp__slack__send'],
  maxMcpOutputTokens: 50000, // New token management
  mcpServerConfig: {
    github: { apiVersion: 'beta' }, // Beta API features
    database: { connectionPool: true }
  }
};
```

### Sub-Agent System (New 2025)

Delegate specialized tasks to domain-specific agents:

```python
# Sub-agents with parallel execution and context isolation
await client.query("""
Use these specialized agents in parallel:
1. security-reviewer agent to audit authentication code
2. performance-analyzer agent to check database queries
3. documentation-generator agent to update API docs
""")
```

### Streaming Responses with Progress

Real-time feedback for long operations with enhanced progress tracking:

```typescript
for await (const message of query({
  prompt,
  options: {
    stream: true,
    model: 'claude-4-1-opus'
  }
})) {
  switch (message.type) {
    case 'thinking':
      console.log('Progress:', message.content);
      break;
    case 'tool_use':
      console.log(`Using ${message.toolName}:`, message.progress);
      break;
    case 'sub_agent':
      console.log(`Agent ${message.agentName}:`, message.status);
      break;
  }
}
```

### Cost Optimization (2025)

Built-in cost tracking and optimization features:

```typescript
const options = {
  model: 'claude-4-1-opus',
  opusplan: true, // Hybrid mode for cost optimization
  maxTurns: 5, // Limit conversation for cost control
  costTracking: true, // Enable detailed cost monitoring
};
```

## Getting Started

### Choose Your SDK

1. **TypeScript**: For Node.js applications and web development with full type safety
2. **Python**: For data science and Python-based automation with async patterns
3. **CLI**: For shell scripting and CI/CD integration with enhanced output formats

### Next Steps

1. Review the specific SDK documentation for your chosen language
2. Set up authentication and configure Claude 4.x model access
3. Configure appropriate permission modes for your security requirements
4. Start with simple queries and gradually add complexity
5. Implement error handling and cost monitoring
6. Explore MCP integrations and sub-agent capabilities for extended functionality

### 2025 Migration Guide

If upgrading from earlier versions:

1. **Update Models**: Migrate to Claude 4.0/4.1 models for improved performance
2. **Permission Modes**: Review and update permission configurations for new "ask" mode
3. **MCP Integration**: Update MCP server configurations for token management
4. **Session Management**: Implement SessionEnd event handlers for cleanup
5. **Cost Tracking**: Enable new cost monitoring and optimization features

### Additional Resources

- **[SDK Overview](sdk-overview.md)**: Architecture and comprehensive guide with 2025 features
- **[TypeScript SDK](sdk-typescript.md)**: Complete TypeScript/JavaScript reference
- **[Python SDK](sdk-python.md)**: Complete Python reference with async patterns
- **[Headless CLI](sdk-headless.md)**: Command-line automation guide
- **[MCP Integration](mcp.md)**: External tool integration with 2025 enhancements
- **[Sub-Agents Documentation](sub-agents-en.md)**: Specialized agent system
- **[Model Configuration](model-config.md)**: Claude 4.x model setup and optimization
- **[Best Practices](claude-code-best-practices.md)**: Production deployment guidance

This SDK documentation provides comprehensive coverage of Claude Code's 2025 capabilities, enabling developers to build sophisticated AI applications with enhanced security, performance, and functionality.