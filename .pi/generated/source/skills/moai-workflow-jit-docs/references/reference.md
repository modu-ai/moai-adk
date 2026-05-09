# moai-workflow-jit-docs Reference

## API Reference

### Intent Analysis API

Intent Detection Functions:
- `analyze_user_intent(user_input, context)`: Determine documentation needs
- `extract_technologies(user_input)`: Identify technology keywords
- `extract_domains(user_input)`: Identify domain keywords
- `classify_question(user_input)`: Categorize question type
- `assess_complexity(user_input)`: Evaluate question complexity
- `determine_urgency(user_input)`: Assess information urgency

Intent Structure:
```python
{
    "technologies": ["FastAPI", "JWT"],
    "domains": ["authentication", "security"],
    "question_type": "implementation",
    "complexity": "medium",
    "urgency": "standard"
}
```

Question Types:
- `implementation`: How to build/create something
- `troubleshooting`: How to fix issues
- `conceptual`: Understanding concepts
- `best_practices`: Recommended approaches
- `comparison`: Evaluating options

### Source Prioritization API

Priority Functions:
- `prioritize_documentation_sources(intent)`: Rank sources by relevance
- `has_local_docs()`: Check for project documentation
- `get_official_docs(technology)`: Get official documentation URL
- `calculate_relevance(content, context)`: Score content relevance

Source Types and Weights:
- `local`: Project documentation (priority 1.0)
- `official`: Official documentation (priority 0.9)
- `web`: Real-time web research (priority 0.8)
- `community`: Community resources (priority 0.7)

### Caching API

DocumentationCache Class:
- `get(key, context)`: Retrieve cached documentation
- `store(key, content, relevance_score)`: Cache with relevance
- `remove(key)`: Remove from cache
- `is_relevant(key, context)`: Check current relevance
- `update_access_time(key)`: Track access patterns

Cache Levels:
- `session`: Current session only
- `project`: Project-specific persistent cache
- `global`: Cross-project shared cache

### Quality Assessment API

Quality Functions:
- `assess_documentation_quality(content)`: Evaluate content quality
- `is_official_source(content)`: Check source authority
- `is_recent(content, months)`: Check content age
- `has_examples(content)`: Check for code examples
- `has_explanations(content)`: Check for explanations

Quality Scoring:
- Authority: 30% (official sources preferred)
- Recency: 25% (recent content preferred)
- Completeness: 25% (examples and explanations)
- Relevance: 20% (match to user context)

---

## Configuration Options

### Source Configuration

Local Documentation Paths:
- `.moai/docs/`: Project-specific documentation
- `.moai/specs/`: Requirements and specifications
- `README.md`: General project information
- `CHANGELOG.md`: Version history
- `docs/`: Comprehensive documentation directory

Official Documentation URLs:
```python
official_docs = {
    "FastAPI": "https://fastapi.tiangolo.com/",
    "React": "https://react.dev/",
    "PostgreSQL": "https://www.postgresql.org/docs/",
    "Docker": "https://docs.docker.com/",
    "Kubernetes": "https://kubernetes.io/docs/",
    "TypeScript": "https://www.typescriptlang.org/docs/"
}
```

### Cache Configuration

Cache Settings:
- `cache_duration_days`: How long to keep cached content
- `max_cache_size_mb`: Maximum cache size
- `auto_cleanup_enabled`: Enable automatic cleanup
- `relevance_threshold`: Minimum relevance to cache

Eviction Policies:
- Remove content older than 30 days
- Keep high-authority sources longer
- Prioritize frequently accessed content
- Remove low-relevance content first

### Intent Detection Settings

Technology Keywords:
- Frameworks: FastAPI, React, Vue, Django, Flask
- Languages: Python, TypeScript, JavaScript, Go, Rust
- Databases: PostgreSQL, MongoDB, Redis, MySQL
- DevOps: Docker, Kubernetes, Terraform, AWS

Domain Keywords:
- Security: authentication, authorization, encryption
- Performance: optimization, caching, scaling
- Database: queries, migrations, indexing
- API: endpoints, REST, GraphQL

---

## Integration Patterns

### Question-Based Loading

Trigger Pattern:
1. User asks specific question
2. Intent analysis identifies needs
3. Sources prioritized by relevance
4. Documentation loaded and cached
5. Enhanced response generated

Example Flow:
```
User: "How do I implement JWT in FastAPI?"

1. Intent: {technologies: [FastAPI, JWT], domains: [auth]}
2. Sources: [FastAPI docs, JWT best practices]
3. Load: Official FastAPI security guide
4. Cache: Store with high relevance score
5. Response: Comprehensive JWT implementation guide
```

### Technology-Specific Loading

Trigger Pattern:
1. Technology keyword detected
2. Official documentation identified
3. Relevant sections loaded
4. Community resources added
5. Latest updates checked

Technology Coverage:
- Language-specific patterns and idioms
- Framework configuration and usage
- Library integration examples
- Version-specific changes

### Domain-Specific Loading

Trigger Pattern:
1. Domain keyword detected
2. Domain expertise documentation loaded
3. Best practices identified
4. Common patterns extracted
5. Troubleshooting guides added

Domain Coverage:
- Authentication and authorization patterns
- Database optimization strategies
- API design best practices
- Performance tuning guides

### Real-Time Web Research

Trigger Pattern:
1. Latest information needed
2. WebSearch executed with current year
3. Results quality-assessed
4. High-quality content extracted
5. Cached for future use

Search Optimization:
```python
WebSearch(f"{query} best practices 2024 2025")
```

---

## Troubleshooting

### Documentation Not Loading

Symptoms: Skill not activating, missing context

Solutions:
1. Verify intent detection triggered
2. Check local documentation paths exist
3. Confirm network connectivity for web sources
4. Review cache for stale entries

### Low Quality Results

Symptoms: Irrelevant or outdated documentation

Solutions:
1. Adjust relevance threshold higher
2. Clear cache to force fresh fetch
3. Prioritize official sources
4. Add year filter to web searches

### Cache Issues

Symptoms: Stale content, cache bloat

Solutions:
1. Clear session cache: Reset current session
2. Clear project cache: Remove project-specific cache
3. Reduce cache duration
4. Lower max cache size

### Intent Misdetection

Symptoms: Wrong documentation loaded

Solutions:
1. User clarification requested
2. Refine keyword extraction
3. Adjust domain mappings
4. Add technology aliases

### Network Failures

Symptoms: Web sources unavailable

Solutions:
1. Fall back to cached content
2. Use local documentation
3. Provide partial results with notice
4. Retry with backoff

---

## External Resources

### Documentation Sources

Official Sources by Category:
- Languages: Official language documentation
- Frameworks: Framework official guides
- Databases: Database documentation
- Cloud: Provider documentation

Community Resources:
- Stack Overflow (highly-voted answers)
- GitHub Discussions (official projects)
- Dev.to (tutorial articles)
- Medium (technical deep-dives)

### Quality Indicators

High-Quality Sources:
- Official documentation (docs.*.com, *.io/docs)
- GitHub repositories with high stars
- Conference talks and official blogs
- Books from recognized publishers

Low-Quality Sources:
- Outdated content (older than 2 years)
- No code examples
- Missing author attribution
- Scraped or aggregated content

### Related Skills

- `moai-docs-generation`: Generate documentation
- `moai-workflow-docs`: Documentation workflows
- `moai-library-nextra`: Nextra documentation sites
- `moai-foundation-context`: Token budget management

### Performance Optimization

Loading Strategies:
- Lazy loading: Load only when needed
- Batch processing: Combine related queries
- Progressive disclosure: Load detail on demand
- Intelligent caching: Predict future needs

Metrics:
- Cache hit rate target: Greater than 70%
- Average load time: Less than 2 seconds
- Relevance score threshold: 0.7 minimum
- Quality score threshold: 0.6 minimum

---

Version: 2.0.0
Last Updated: 2025-12-06
