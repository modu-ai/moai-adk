---
name: yoda-citation-verification
description: "ZERO-TOLERANCE citation verification for YodA books. 100% URL verification required. Generates verified-citations-cache.json. Version 2.0.0 - Two-Phase Architecture."
allowed-tools: Read, Write, Edit, Bash, Glob, Grep, WebFetch, Task, mcp__context7__resolve-library-id, mcp__context7__get-library-docs, WebSearch
version: 2.0.0
category: verification
updated: 2025-12-01
status: active
---

## Quick Reference (30 seconds)

**Core API**:
```python
# Main verification function
result = verify_citations(
    citations=["https://docs.anthropic.com/claude-code"],
    domain="claude_code"
)

# Generate trusted citations only
citations = generate_trusted_citations("Claude Code agents", count=3)

# Format citation section
formatted = format_citations(citations, style="korean")

# Validate citation section
validation = validate_citation_section(content_with_citations)
```

**Zero-Tolerance Rules**:
- ✅ Only pre-approved trusted sources from database
- ✅ Real-time URL accessibility verification
- ✅ Content relevance validation
- ❌ No creative citation generation
- ❌ No hallucinated sources
- ❌ No unverified URLs

**Integration**: Automatically loaded by yoda-book-author agent for foolproof citation verification.

---

## Implementation Guide

### Core System Architecture

The YodA Citation Verification Skill provides a comprehensive zero-tolerance system for preventing hallucinations in academic and technical writing. It integrates with the trusted citation database and provides real-time verification capabilities.

**Key Components**:
1. **Trusted Database Integration** - Loads from `/moai/utils/trusted-citations/database.json`
2. **Real-time URL Verification** - Validates accessibility and content relevance
3. **Zero-Tolerance Enforcement** - Blocks any non-trusted sources
4. **Context7 Integration** - Latest official documentation verification
5. **Korean Citation Formatting** - Proper formatting for Korean technical writing

### Database Integration

## Real-Time Verification Workflow (Executable)

### Step 1: Load Trusted Database

Execute:
```
database_content = Read(file_path="/Users/goos/MoAI/yoda/.moai/yoda/trusted-citations/database.json")
```

Parse JSON to extract source list:
```python
import json
database = json.loads(database_content)
domain_sources = database.get("domains", {}).get(domain, {}).get("mandatory_sources", [])
```

### Step 2: Verify Each URL

For EACH citation URL, execute WebFetch:

```
result = WebFetch(
    url=citation_url,
    prompt="Check if this page exists and extract the title"
)
```

Check result for:
- HTTP 200 status (page exists)
- Title extraction (content is readable)
- No error messages

### Step 3: Content Relevance Check

Extract keywords from citation description and verify they appear in page content.

Example:
```python
# Extract keywords from description
keywords = citation["description"].lower().split()

# Check if keywords appear in WebFetch result
matches = sum(1 for keyword in keywords if keyword in result.lower())
relevance_score = matches / len(keywords) if keywords else 0.0
```

### Step 4: Generate Verification Record

For each verified citation:
```json
{
  "id": "CC-001",
  "title": "...",
  "url": "https://...",
  "verified_at": "{ISO timestamp}",
  "status": "verified",
  "content_preview": "First 200 chars of page..."
}
```

### Step 5: 100% Enforcement

- IF all citations verified → Generate cache
- IF ANY failed → REJECT entire batch
- NO partial success

**Domain Mapping**:
```python
DOMAIN_MAPPING = {
    "claude-code": "claude_code",
    "claude code": "claude_code",
    "anthropic": "claude_code",
    "python": "python",
    "python3": "python",
    "react": "react",
    "react.js": "react",
    "nextjs": "nextjs",
    "next.js": "nextjs",
    "node": "nodejs",
    "node.js": "nodejs",
    "typescript": "typescript",
    "ts": "typescript",
    "github": "github"
}

def map_topic_to_domain(topic: str) -> str:
    """Map topic keywords to database domains"""
    topic_lower = topic.lower()

    for keyword, domain in DOMAIN_MAPPING.items():
        if keyword in topic_lower:
            return domain

    # Default to general
    return "general"
```

---


---

## Implementation Details

For detailed implementation including Cache Management, Citation ID System, Real-Time Verification, and Context7 Integration, see [Implementation Details](modules/implementation-details.md).

Key Topics:
- Verified Citation Cache Management
- Citation ID System
- Real-Time Verification Workflow
- Zero-Tolerance Enforcement
- Korean Citation Formatting
- Context7 Integration

## Advanced Patterns

For advanced implementation patterns including batch processing, quality metrics, and error handling, see [Advanced Patterns](modules/advanced-patterns.md).

Key Topics:
- Batch Citation Verification with concurrent processing
- Citation Quality Assessment System
- Zero-Tolerance Error Handling

---

## Works Well With

- **yoda-book-author** - Automatic integration for zero-tolerance citation verification
- **mcp-context7** - Latest official documentation verification
- **yoda-manuscript-quality-standards** - Quality assurance for citations
- **moai-playwright-url-verification** - Real-time URL accessibility checking

## Resources

### Database Schema
The trusted citation database structure is defined in `/moai/utils/trusted-citations/database.json` with the following schema:

```json
{
  "version": "1.0.0",
  "domains": {
    "claude_code": {
      "mandatory_sources": [...],
      "forbidden_patterns": [...]
    }
  },
  "forbidden_domains": [...],
  "allowed_domains": [...],
  "verification_rules": {...},
  "fallback_sources": [...]
}
```

### Performance Metrics
- **Single citation verification**: < 1 second
- **Batch verification (10 citations)**: < 5 seconds
- **Cache hit rate**: > 80% for repeated verifications
- **False positive rate**: < 1% (trusted sources only)

### Integration Examples
See `examples/integration/` for complete examples of integrating with yoda-book-author and other agents.

---

**Version**: 1.0.0
**Last Updated**: 2025-11-30
**Integration Status**: ✅ Complete - Ready for production use with yoda-book-author v5.0