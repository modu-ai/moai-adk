# Nano-Banana MCP Advanced Examples

## AI Content Generation Workflows

### Documentation Generation Pipeline

Complete documentation generation from specification data:

```python
async def generate_documentation_workflow(
    spec_data: dict,
    output_format: str = "markdown"
) -> dict:
    """Complete documentation generation from specification."""

    # Phase 1: Generate API reference
    api_docs = await mcp_server.invoke_tool("generate_ai_content", {
        "prompt": f"""Create comprehensive API documentation:

        Endpoints:
        {json.dumps(spec_data['api_endpoints'], indent=2)}

        Requirements:
        {json.dumps(spec_data['requirements'], indent=2)}
        """,
        "model": "claude-3-5-sonnet",
        "max_tokens": 3000,
        "temperature": 0.3
    })

    # Phase 2: Generate usage examples
    examples = await mcp_server.invoke_tool("generate_ai_content", {
        "prompt": f"""Create code examples for API:

        {api_docs['content']}

        Include: Python, TypeScript, curl examples
        """,
        "model": "claude-3-5-sonnet",
        "max_tokens": 2000,
        "temperature": 0.5
    })

    # Phase 3: Generate tutorials
    tutorials = await mcp_server.invoke_tool("generate_ai_content", {
        "prompt": f"""Create step-by-step tutorials:

        API Documentation: {api_docs['content']}
        Examples: {examples['content']}

        Target audience: Developers (beginner to intermediate)
        """,
        "model": "claude-3-5-sonnet",
        "max_tokens": 4000,
        "temperature": 0.7
    })

    # Phase 4: Generate diagrams descriptions
    diagrams = await mcp_server.invoke_tool("generate_ai_content", {
        "prompt": f"""Create diagram descriptions for:

        {spec_data['description']}

        Include: Architecture, sequence, flow diagrams
        Format: Mermaid diagram syntax
        """,
        "model": "claude-3-5-sonnet",
        "max_tokens": 1500,
        "temperature": 0.4
    })

    return {
        "api_reference": api_docs['content'],
        "code_examples": examples['content'],
        "tutorials": tutorials['content'],
        "diagrams": diagrams['content'],
        "format": output_format,
        "tokens_used": (
            api_docs['usage']['total_tokens'] +
            examples['usage']['total_tokens'] +
            tutorials['usage']['total_tokens'] +
            diagrams['usage']['total_tokens']
        )
    }
```

### Multi-Language Content Generation

Generate content in multiple languages:

```python
async def multilingual_content_workflow(
    base_content: str,
    target_languages: list
) -> dict:
    """Generate content in multiple languages."""

    translations = {}

    for language in target_languages:
        # Generate localized content
        localized = await mcp_server.invoke_tool("generate_ai_content", {
            "prompt": f"""Translate and culturally adapt content:

            Source Content:
            {base_content}

            Target Language: {language}

            Requirements:
            - Preserve technical accuracy
            - Adapt cultural references
            - Maintain formatting
            - Keep code examples in English
            """,
            "model": "claude-3-5-sonnet",
            "max_tokens": len(base_content) * 2,
            "temperature": 0.5
        })

        translations[language] = localized['content']

    return {
        "base_content": base_content,
        "translations": translations,
        "languages": target_languages,
        "total_tokens": sum(t['usage']['total_tokens'] for t in translations.values())
    }
```

## Image Generation Workflows

### Design Asset Creation

Generate complete set of design assets:

```python
async def generate_design_assets(
    asset_specifications: list,
    style_guide: dict
) -> dict:
    """Generate complete set of design assets."""

    generated_assets = []

    for spec in asset_specifications:
        # Create base image
        image = await mcp_server.invoke_tool("generate_image", {
            "prompt": f"""Create {spec['type']} image:

            Description: {spec['description']}
            Style: {style_guide['visual_style']}
            Colors: {', '.join(style_guide['color_palette'])}
            Size: {spec['dimensions']}
            """,
            "size": spec['dimensions'],
            "style": style_guide['visual_style'],
            "quality": "high"
        })

        # Generate variations if requested
        variations = []
        if spec.get('create_variations', False):
            for i in range(spec['variation_count']):
                variant = await mcp_server.invoke_tool("generate_image_variation", {
                    "source_image": image['url'],
                    "variation_strength": 0.3 + (i * 0.2),
                    "preserve_style": True
                })
                variations.append(variant)

        generated_assets.append({
            "type": spec['type'],
            "base_image": image,
            "variations": variations,
            "metadata": {
                "created_at": datetime.now().isoformat(),
                "style_guide": style_guide['name'],
                "specification": spec
            }
        })

    return {
        "assets": generated_assets,
        "style_guide": style_guide,
        "total_count": len(generated_assets),
        "total_variations": sum(len(a['variations']) for a in generated_assets)
    }
```

## Text Analysis Workflows

### Content Quality Assessment

Comprehensive content quality analysis:

```python
async def assess_content_quality(
    content: str,
    quality_criteria: dict
) -> dict:
    """Comprehensive content quality analysis."""

    # Phase 1: Readability analysis
    readability = await mcp_server.invoke_tool("analyze_with_ai", {
        "content": content,
        "analysis_type": "readability",
        "include_metrics": True
    })

    # Phase 2: Technical accuracy
    accuracy = await mcp_server.invoke_tool("analyze_with_ai", {
        "content": content,
        "analysis_type": "technical_accuracy",
        "criteria": quality_criteria.get('technical_standards', {})
    })

    # Phase 3: Completeness check
    completeness = await mcp_server.invoke_tool("analyze_with_ai", {
        "content": content,
        "analysis_type": "completeness",
        "required_sections": quality_criteria.get('required_sections', [])
    })

    # Phase 4: Style consistency
    style = await mcp_server.invoke_tool("analyze_with_ai", {
        "content": content,
        "analysis_type": "style_consistency",
        "style_guide": quality_criteria.get('style_guide', {})
    })

    # Calculate overall quality score
    quality_score = calculate_quality_score({
        'readability': readability['score'],
        'accuracy': accuracy['score'],
        'completeness': completeness['score'],
        'style': style['score']
    })

    return {
        "quality_score": quality_score,
        "readability": readability,
        "technical_accuracy": accuracy,
        "completeness": completeness,
        "style_consistency": style,
        "recommendations": generate_improvement_recommendations(
            readability, accuracy, completeness, style
        )
    }
```

## Batch Processing

### Automated Content Pipeline

Process multiple content items in parallel:

```python
async def batch_content_generation(
    input_items: list,
    processing_config: dict
) -> dict:
    """Process multiple content items in parallel."""

    # Configure batch processing
    batch_size = processing_config.get('batch_size', 5)
    max_retries = processing_config.get('max_retries', 3)

    results = []
    errors = []

    # Process in batches
    for i in range(0, len(input_items), batch_size):
        batch = input_items[i:i + batch_size]

        # Process batch items concurrently
        batch_results = await asyncio.gather(
            *[
                process_single_item(item, processing_config, max_retries)
                for item in batch
            ],
            return_exceptions=True
        )

        # Separate successful results from errors
        for item, result in zip(batch, batch_results):
            if isinstance(result, Exception):
                errors.append({
                    'item': item,
                    'error': str(result),
                    'timestamp': datetime.now().isoformat()
                })
            else:
                results.append(result)

        # Rate limiting between batches
        if i + batch_size < len(input_items):
            await asyncio.sleep(processing_config.get('batch_delay', 1.0))

    return {
        "total_items": len(input_items),
        "successful": len(results),
        "failed": len(errors),
        "results": results,
        "errors": errors,
        "processing_time": calculate_processing_time()
    }

async def process_single_item(
    item: dict,
    config: dict,
    max_retries: int
) -> dict:
    """Process single item with retry logic."""

    for attempt in range(max_retries):
        try:
            result = await mcp_server.invoke_tool("generate_ai_content", {
                "prompt": item['prompt'],
                "model": config.get('model', 'claude-3-5-sonnet'),
                "max_tokens": config.get('max_tokens', 2000),
                "temperature": config.get('temperature', 0.7)
            })

            return {
                'item_id': item['id'],
                'content': result['content'],
                'tokens_used': result['usage']['total_tokens'],
                'attempt': attempt + 1
            }

        except Exception as e:
            if attempt == max_retries - 1:
                raise
            await asyncio.sleep(2 ** attempt)  # Exponential backoff
```

## Usage Examples

### Content Generation

Generate API documentation:

```python
# API documentation
api_docs = await mcp_server.invoke_tool("generate_ai_content", {
    "prompt": """Create comprehensive REST API documentation:

    Endpoints:
    - POST /api/auth/login - User authentication
    - GET /api/users/:id - Retrieve user profile
    - PUT /api/users/:id - Update user profile

    Include: Request/response examples, error codes, authentication
    """,
    "model": "claude-3-5-sonnet",
    "max_tokens": 3000,
    "temperature": 0.3
})

# Code examples
code_examples = await mcp_server.invoke_tool("generate_ai_content", {
    "prompt": """Create code examples for JWT authentication:

    Languages: Python, TypeScript
    Include: Token generation, validation, refresh flow
    """,
    "model": "claude-3-5-sonnet",
    "max_tokens": 2000,
    "temperature": 0.5
})
```

### Image Generation

Create visual assets:

```python
# Hero image
hero_image = await mcp_server.invoke_tool("generate_image", {
    "prompt": "Modern SaaS dashboard hero image, clean UI, gradient background",
    "size": "1920x1080",
    "style": "digital_art",
    "quality": "high"
})

# Icon set
icons = []
icon_descriptions = [
    "User profile icon, minimal line art",
    "Settings gear icon, modern design",
    "Notification bell icon, clean style"
]

for description in icon_descriptions:
    icon = await mcp_server.invoke_tool("generate_image", {
        "prompt": description,
        "size": "256x256",
        "style": "vector",
        "background": "transparent"
    })
    icons.append(icon)
```

### Text Analysis

Analyze content:

```python
# Sentiment analysis
sentiment = await mcp_server.invoke_tool("analyze_with_ai", {
    "content": user_feedback,
    "analysis_type": "sentiment",
    "include_scores": True
})

# Summarization
summary = await mcp_server.invoke_tool("analyze_with_ai", {
    "content": long_document,
    "analysis_type": "summary",
    "max_length": 200,
    "include_key_points": True
})

# Entity extraction
entities = await mcp_server.invoke_tool("analyze_with_ai", {
    "content": technical_document,
    "analysis_type": "entity_extraction",
    "entity_types": ["technology", "framework", "tool", "concept"]
})
```
