---
name: moai-ai-nano-banana
description: Nano-Banana AI service integration for content generation, image creation, and AI-powered workflows. Use when integrating AI services for content creation.
version: 1.0.0
category: integration
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
tags:
  - ai
  - content-generation
  - image-generation
  - nano-banana
  - ai-service
related-skills:
  - moai-docs-generation
  - moai-domain-uiux
updated: 2025-12-07
status: active
author: MoAI-ADK Team
---

# Nano-Banana AI Service Integration

## Quick Reference (30 seconds)

Nano-Banana MCP Integration - Specialized MCP connector for AI-powered content generation, image creation, text processing, and multi-modal AI workflows using Nano-Banana AI services.

Core Capabilities:
- AI Content Generation: Text, documentation, code generation
- Image Generation: AI-powered image creation and editing
- Text Analysis: Sentiment analysis, summarization, extraction
- Multi-Modal Operations: Combined text and image workflows
- Workflow Automation: Batch processing and pipelines

When to Use:
- Generating AI-powered content for documentation
- Creating images and visual assets programmatically
- Building automated content pipelines
- Implementing AI analysis workflows
- Developing multi-modal AI applications

---

## Implementation Guide (5 minutes)

### Quick Start Workflow

Nano-Banana MCP Server Setup:
```python
from moai_integration_mcp.nano_banana import NanoBananaMCPServer

# Initialize server with API credentials
mcp_server = NanoBananaMCPServer("nano-banana-server")

# Configure authentication
mcp_server.setup_credentials({
    'api_key': os.getenv('NANO_BANANA_TOKEN'),
    'api_url': 'https://api.nano-banana.ai/v1'
})

# Register AI tools
mcp_server.register_tools()

# Start server
mcp_server.start(port=3001)
```

Basic AI Content Generation:
```bash
# Generate text content
mcp-tools nano_banana generate_content \
  --prompt "Create API documentation for user authentication" \
  --model "claude-3-5-sonnet" \
  --max_tokens 2000

# Create image from description
mcp-tools nano_banana generate_image \
  --prompt "Modern dashboard UI with dark theme" \
  --size "1024x1024" \
  --style "photorealistic"

# Analyze text content
mcp-tools nano_banana analyze_text \
  --input "./docs/content.md" \
  --analysis_type "summary" \
  --include_key_points
```

### Core Operations

Content Generation:
```python
# Generate documentation
result = await mcp_server.invoke_tool("generate_ai_content", {
    "prompt": "Create API documentation for authentication endpoints",
    "model": "claude-3-5-sonnet",
    "max_tokens": 3000,
    "temperature": 0.3
})

# Multi-language content
translations = await mcp_server.invoke_tool("translate_content", {
    "source_content": documentation,
    "target_languages": ["ko", "ja", "zh"],
    "preserve_formatting": True
})
```

Image Creation:
```python
# Generate images
image = await mcp_server.invoke_tool("generate_image", {
    "prompt": "Modern SaaS dashboard hero image",
    "size": "1920x1080",
    "style": "digital_art",
    "quality": "high"
})

# Create variations
variations = await mcp_server.invoke_tool("generate_image_variations", {
    "source_image": image['url'],
    "count": 3,
    "variation_strength": 0.5
})
```

Text Analysis:
```python
# Analyze content
analysis = await mcp_server.invoke_tool("analyze_with_ai", {
    "content": document_text,
    "analysis_type": "comprehensive",
    "include_sentiment": True,
    "include_summary": True,
    "include_entities": True
})

# Quality assessment
quality = await mcp_server.invoke_tool("assess_content_quality", {
    "content": documentation,
    "criteria": {
        "readability": True,
        "technical_accuracy": True,
        "completeness": True
    }
})
```

### Workflow Patterns

Documentation Generation Pipeline:
```python
async def documentation_workflow(spec_data: dict):
    """Complete documentation generation from specification."""

    # Generate API reference
    api_docs = await mcp_server.invoke_tool("generate_ai_content", {
        "prompt": f"Create API documentation: {spec_data['endpoints']}",
        "max_tokens": 3000
    })

    # Generate code examples
    examples = await mcp_server.invoke_tool("generate_ai_content", {
        "prompt": f"Create code examples for: {api_docs['content']}",
        "max_tokens": 2000
    })

    # Generate tutorials
    tutorials = await mcp_server.invoke_tool("generate_ai_content", {
        "prompt": f"Create tutorials for: {examples['content']}",
        "max_tokens": 4000
    })

    return {
        "api_reference": api_docs,
        "code_examples": examples,
        "tutorials": tutorials
    }
```

Batch Processing:
```python
async def batch_process_content(items: list, config: dict):
    """Process multiple items in parallel."""

    results = []
    batch_size = config.get('batch_size', 5)

    for i in range(0, len(items), batch_size):
        batch = items[i:i + batch_size]

        batch_results = await asyncio.gather(*[
            mcp_server.invoke_tool("generate_ai_content", {
                "prompt": item['prompt'],
                "max_tokens": config['max_tokens']
            })
            for item in batch
        ])

        results.extend(batch_results)

        # Rate limiting
        if i + batch_size < len(items):
            await asyncio.sleep(1.0)

    return results
```

---

## Advanced Patterns (10+ minutes)

### Content Generation

Documentation Pipeline:
- Multi-phase documentation generation (API reference, examples, tutorials)
- Template-based content creation with custom prompts
- Multi-language documentation with cultural adaptation
- Automated diagram generation (Mermaid syntax)

See [examples.md](examples.md) for complete implementation.

Image Workflows:
- Design asset generation with style guides
- Image variation generation for A/B testing
- Batch image creation with consistent styling
- Image editing and enhancement workflows

Quality Assurance:
- Content quality assessment with multiple metrics
- Readability and technical accuracy analysis
- Completeness verification against requirements
- Style consistency validation

Batch Operations:
- Parallel content processing with rate limiting
- Retry logic with exponential backoff
- Error handling and recovery
- Progress tracking and monitoring

### Integration Patterns

MCP Tool Registration:
```python
@mcp_server.tool()
async def generate_documentation(
    spec_id: str,
    output_format: str = "markdown"
) -> dict:
    """Generate comprehensive documentation from SPEC."""

    # Load specification
    spec = load_spec(spec_id)

    # Generate documentation
    result = await documentation_workflow(spec)

    return {
        "spec_id": spec_id,
        "documentation": result,
        "format": output_format
    }
```

Error Handling:
```python
async def resilient_generation(prompt: str, max_retries: int = 3):
    """Generate content with retry logic."""

    for attempt in range(max_retries):
        try:
            return await mcp_server.invoke_tool("generate_ai_content", {
                "prompt": prompt,
                "max_tokens": 2000
            })
        except Exception as e:
            if attempt == max_retries - 1:
                raise
            await asyncio.sleep(2 ** attempt)
```

---

## Technology Stack

Core Framework:
- FastMCP (Python MCP server framework)
- AsyncIO for concurrent operations
- HTTPX for HTTP client
- Pydantic for data validation

AI Services:
- Nano-Banana AI API
- Claude models (Sonnet, Opus, Haiku)
- Image generation models
- Text analysis services

Authentication & Security:
- API key management
- Token-based authentication
- Secure credential storage
- Rate limiting and quotas

Error Handling:
- Retry logic with exponential backoff
- Circuit breaker patterns
- Comprehensive error classification
- Monitoring and logging

---

## Configuration

Environment Variables:
```bash
# Nano-Banana API
NANO_BANANA_TOKEN=your_api_key
NANO_BANANA_API_URL=https://api.nano-banana.ai/v1

# Model Configuration
DEFAULT_MODEL=claude-3-5-sonnet
DEFAULT_MAX_TOKENS=2000
DEFAULT_TEMPERATURE=0.7

# Rate Limiting
BATCH_SIZE=5
BATCH_DELAY=1.0
MAX_RETRIES=3
```

MCP Server Configuration:
```json
{
  "nano_banana": {
    "enabled": true,
    "api_url": "https://api.nano-banana.ai/v1",
    "default_model": "claude-3-5-sonnet",
    "rate_limits": {
      "requests_per_minute": 60,
      "tokens_per_minute": 100000
    },
    "retry_config": {
      "max_retries": 3,
      "backoff_factor": 2
    }
  }
}
```

---

## Performance Optimization

Token Management:
- Optimize prompt length for efficiency
- Use appropriate max_tokens settings
- Implement token counting for cost tracking
- Cache frequently requested content

Parallel Processing:
- Batch operations with concurrent execution
- Configure optimal batch sizes
- Implement rate limiting between batches
- Use connection pooling for HTTP requests

Error Recovery:
- Implement exponential backoff for retries
- Use circuit breakers for failing services
- Log errors for debugging and monitoring
- Provide graceful degradation

---

## Usage Examples

Quick Start:
```python
# Initialize server
server = NanoBananaMCPServer("ai-server")
server.setup_credentials({'api_key': os.getenv('NANO_BANANA_TOKEN')})
server.start()

# Generate content
content = await server.invoke_tool("generate_ai_content", {
    "prompt": "Create user guide for REST API",
    "max_tokens": 2000
})

# Generate image
image = await server.invoke_tool("generate_image", {
    "prompt": "Dashboard UI mockup",
    "size": "1024x1024"
})

# Analyze text
analysis = await server.invoke_tool("analyze_with_ai", {
    "content": documentation,
    "analysis_type": "summary"
})
```

For comprehensive examples including documentation generation, image workflows, and batch processing, see [examples.md](examples.md).

---

## Works Well With

Complementary Skills:
- `moai-docs-generation` - Automated documentation workflows
- `moai-domain-uiux` - UI/UX design integration
- `moai-domain-frontend` - Frontend component generation
- `moai-workflow-templates` - Template-based content

Integration Points:
- Documentation generation pipelines
- Design system workflows
- Content management systems
- Multi-language documentation

---

## Resources

Core Files:
- `SKILL.md` - Main skill documentation (this file)
- `examples.md` - Advanced examples and complete workflows
- `modules/` - Implementation modules (if applicable)

External Documentation:
- Nano-Banana API Documentation
- FastMCP Framework Guide
- Claude API Reference
- MCP Protocol Specification

---

Version: 1.0.0
Last Updated: 2025-12-07
Status: Active
