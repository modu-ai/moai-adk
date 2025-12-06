---
name: moai-workflow-jit-docs
description: Enhanced Just-In-Time document loading system that intelligently discovers, loads, and caches relevant documentation based on user intent and project context. Use when users need specific documentation, when working with new technologies, when answering domain-specific questions, or when context indicates documentation gaps.
version: 2.0.0
modularized: false
category: workflow
allowed-tools: Read, Write, Edit, Grep, Glob
aliases:
 - moai-workflow-jit-docs
---

## Quick Reference (30 seconds)

# Enhanced JIT Documentation Loader

## Implementation Guide

## Intent Detection Patterns

### 1. Question-Based Triggers
```python
# User asks specific questions
"how do I implement JWT in FastAPI?"
"What's the best way to handle database migrations?"
"How does React useEffect work exactly?"
"What are the security implications of JWT?"

# Trigger: Load relevant documentation
Skill("moai-workflow-jit-docs")
# → Loads FastAPI JWT authentication docs
# → Loads database migration best practices
# → Loads React hooks documentation
# → Loads JWT security analysis
```

### 2. Technology-Specific Triggers
```python
# Technology keywords detected
"FastAPI", "React", "PostgreSQL", "Docker", "Kubernetes"
"pytest", "TypeScript", "GraphQL", "Redis", "NGINX"

# Trigger: Load technology-specific documentation
Skill("moai-workflow-jit-docs")
# → Loads official docs, tutorials, best practices
```

### 3. Domain-Specific Triggers
```python
# Domain keywords detected
"authentication", "authorization", "security", "performance"
"database", "api", "frontend", "backend", "devops"

# Trigger: Load domain expertise documentation
Skill("moai-workflow-jit-docs")
# → Loads domain-specific patterns and guidelines
```

### 4. Pattern-Based Triggers
```python
# Implementation patterns
"implement", "create", "build", "design", "architecture"
"best practices", "optimization", "troubleshooting", "debugging"

# Trigger: Load implementation guidance
Skill("moai-workflow-jit-docs")
# → Loads implementation patterns and examples
```

## Documentation Sources

### 1. Local Project Documentation
```bash
# Primary sources (highest priority)
.moai/docs/ # Project-specific docs
.moai/specs/ # Requirements and specifications
README.md # General project information
CHANGELOG.md # Version history
docs/ # Comprehensive documentation
```

### 2. Official Documentation
```python
# Technology-specific official sources
official_docs = {
 "FastAPI": "https://fastapi.tiangolo.com/",
 "React": "https://react.dev/",
 "PostgreSQL": "https://www.postgresql.org/docs/",
 "Docker": "https://docs.docker.com/",
 "Kubernetes": "https://kubernetes.io/docs/"
}
```

### 3. Community Resources
```python
# High-quality community resources
community_resources = {
 "Stack Overflow": "Highly-voted answers",
 "GitHub Discussions": "Official project discussions",
 "Dev.to": "Tutorial articles",
 "Medium": "Technical deep-dives"
}
```

### 4. Real-Time Web Research
```python
# Dynamic content for latest information
def get_latest_info(query):
 return WebSearch(f"{query} best practices 2024 2025")
```

## Loading Strategies

### 1. Intent Analysis
```python
def analyze_user_intent(user_input, context):
 """Determine what documentation is needed"""
 intent = {
 "technologies": extract_technologies(user_input),
 "domains": extract_domains(user_input),
 "question_type": classify_question(user_input),
 "complexity": assess_complexity(user_input),
 "urgency": determine_urgency(user_input)
 }
 return intent
```

### 2. Source Prioritization
```python
def prioritize_documentation_sources(intent):
 """Rank sources based on intent and context"""
 priorities = []

 # 1. Local project docs (always first)
 if has_local_docs():
 priorities.append(("local", 1.0))

 # 2. Official docs (high authority)
 for tech in intent["technologies"]:
 if official_docs.get(tech):
 priorities.append(("official", 0.9))

 # 3. Community resources (practical examples)
 if intent["question_type"] == "implementation":
 priorities.append(("community", 0.7))

 # 4. Web research (latest info)
 if intent["urgency"] == "latest":
 priorities.append(("web", 0.8))

 return sorted(priorities, key=lambda x: x[1], reverse=True)
```

### 3. Intelligent Caching
```python
class DocumentationCache:
 """Smart caching system for documentation"""

 def __init__(self):
 self.cache = {}
 self.relevance_scores = {}
 self.access_times = {}

 def get(self, key, context):
 """Get cached documentation if still relevant"""
 if key in self.cache:
 # Check relevance based on context
 if self.is_relevant(key, context):
 self.update_access_time(key)
 return self.cache[key]
 else:
 # Remove outdated content
 self.remove(key)
 return None

 def store(self, key, content, relevance_score):
 """Store documentation with relevance score"""
 self.cache[key] = content
 self.relevance_scores[key] = relevance_score
 self.access_times[key] = datetime.now()
```

## Quality Assessment

### 1. Content Quality Metrics
```python
def assess_documentation_quality(content):
 """Evaluate documentation quality"""
 quality_score = 0.0

 # Authority (30%)
 if is_official_source(content):
 quality_score += 0.3
 elif is_reputable_community(content):
 quality_score += 0.2

 # Recency (25%)
 if is_recent(content, months=6):
 quality_score += 0.25
 elif is_recent(content, months=12):
 quality_score += 0.15

 # Completeness (25%)
 if has_examples(content):
 quality_score += 0.1
 if has_code_samples(content):
 quality_score += 0.1
 if has_explanations(content):
 quality_score += 0.05

 # Relevance (20%)
 relevance = calculate_relevance(content, user_context)
 quality_score += relevance * 0.2

 return min(quality_score, 1.0)
```

### 2. Relevance Ranking
```python
def rank_documentation_results(results, user_context):
 """Rank documentation by relevance to user context"""
 ranked_results = []

 for result in results:
 relevance_score = calculate_relevance(result, user_context)
 quality_score = assess_documentation_quality(result)

 # Combined score: 70% relevance, 30% quality
 combined_score = (relevance_score * 0.7) + (quality_score * 0.3)

 ranked_results.append({
 "content": result,
 "relevance": relevance_score,
 "quality": quality_score,
 "combined": combined_score
 })

 return sorted(ranked_results, key=lambda x: x["combined"], reverse=True)
```

## Integration Examples

### Example 1: Authentication Implementation
```python
# User asks: "How do I implement JWT authentication in FastAPI?"

# 1. Intent Analysis
intent = {
 "technologies": ["FastAPI", "JWT"],
 "domains": ["authentication", "security"],
 "question_type": "implementation",
 "complexity": "medium"
}

# 2. Documentation Loading
Skill("moai-workflow-jit-docs")

# 3. Loaded Documentation
docs_loaded = [
 "FastAPI Security - OAuth2 with JWT (Official)",
 "JWT Best Practices for APIs (Community)",
 "FastAPI JWT Implementation Tutorial (High-quality)",
 "JWT Security Considerations (Latest research)"
]

# 4. Enhanced Response
Alfred can now provide comprehensive, accurate guidance
```

### Example 2: Database Optimization
```python
# User asks: "My PostgreSQL queries are slow, how can I optimize them?"

# 1. Intent Detection
# Technologies: PostgreSQL
# Domains: database, performance, optimization
# Question Type: troubleshooting

# 2. JIT Documentation Loading
Skill("moai-workflow-jit-docs")

# 3. Loaded Resources
performance_docs = [
 "PostgreSQL Query Optimization Guide (Official)",
 "EXPLAIN ANALYZE Best Practices (Community)",
 "Database Indexing Strategies (Expert article)",
 "PostgreSQL Performance Tuning (Latest version)"
]

# 4. Contextual Response
Alfred provides specific, actionable optimization advice
```

### Example 3: New Technology Adoption
```python
# User wants to implement GraphQL in existing project

# 1. Technology Detection
# New technology: GraphQL
# Context: Existing REST API project
# Need: Migration guidance

# 2. Comprehensive Documentation Loading
Skill("moai-workflow-jit-docs")

# 3. Multi-Source Documentation
graphql_docs = [
 "GraphQL Official Documentation",
 "Apollo Server Best Practices",
 "REST to GraphQL Migration Guide",
 "GraphQL Schema Design Patterns",
 "Performance Optimization for GraphQL"
]

# 4. Strategic Guidance
Alfred provides complete migration strategy
```

## Performance Optimization

### 1. Caching Strategy
```python
# Multi-level caching
cache_levels = {
 "session": {}, # Current session only
 "project": {}, # Project-specific cache
 "global": {} # Cross-project cache
}

# Cache eviction policies
def evict_old_cache():
 """Remove outdated documentation"""
 # Remove content older than 30 days
 # Keep high-authority sources longer
 # Prioritize frequently accessed content
```

### 2. Lazy Loading
```python
def load_documentation_on_demand():
 """Load documentation only when needed"""
 # Don't preload everything
 # Load based on user interaction
 # Cache for future use
 pass
```

### 3. Batch Processing
```python
def batch_web_searches(queries):
 """Combine multiple searches for efficiency"""
 # Group similar queries
 # Use WebSearch for multiple terms
 # Process results in parallel
 pass
```

## Error Handling

### 1. Network Failures
```python
if web_search_fails():
 # Fall back to cached content
 # Use local documentation
 # Provide partial results
```

### 2. Content Quality Issues
```python
if content_quality_low():
 # Try alternative sources
 # Indicate uncertainty
 # Request user clarification
```

### 3. Relevance Mismatches
```python
if relevance_score < threshold:
 # Refine search query
 # Ask user for clarification
 # Broaden search scope
```

## Usage Statistics and Learning

### 1. Track Effectiveness
```python
def track_documentation_usage(doc_id, user_feedback):
 """Learn which documentation is most helpful"""
 stats[doc_id] = {
 "usage_count": stats[doc_id]["usage_count"] + 1,
 "helpful_votes": stats[doc_id]["helpful_votes"] + user_feedback,
 "contexts": stats[doc_id]["contexts"].append(current_context)
 }
```

### 2. Improve Source Selection
```python
def improve_source_ranking():
 """Learn which sources work best for different contexts"""
 # Analyze usage patterns
 # Adjust source priorities
 # Personalize recommendations
```

End of Skill | Intelligent documentation loading for enhanced context and accuracy

## Advanced Patterns

## What It Does

Advanced Just-In-Time documentation loading system that intelligently discovers, retrieves, and caches relevant documentation based on user intent, project context, and knowledge gaps. Enhances Alfred's capabilities by bringing in the right documentation at the right time.

Core capabilities:
- Intent-based document discovery
- Context-aware loading strategies
- Intelligent caching and retrieval
- Multi-source documentation aggregation
- Domain-specific knowledge integration
- Real-time web research for latest information
- Documentation quality assessment
- Automatic relevance ranking

## When to Use

- When user questions indicate knowledge gaps
- When working with unfamiliar technologies or frameworks
- When domain-specific expertise is needed
- When answering technical questions with precision
- When implementing new features or patterns
- When troubleshooting complex issues
- When best practices guidance is required

---

## Works Well With

Agents:
- workflow-docs - Documentation generation
- core-planner - Documentation planning
- workflow-spec - SPEC documentation

Skills:
- moai-docs-generation - Documentation generation
- moai-workflow-docs - Documentation validation
- moai-library-nextra - Nextra documentation

Commands:
- `/moai:3-sync` - Documentation synchronization
- `/moai:9-feedback` - Documentation improvements
