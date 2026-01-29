---
name: moai-workflow-jit-docs
description: >
  Enhanced Just-In-Time document loading system that intelligently discovers,
  loads, and caches relevant documentation based on user intent and project
  context. Use when users need specific documentation, when working with new
  technologies, when answering domain-specific questions, or when context
  indicates documentation gaps.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob WebFetch WebSearch mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "3.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-08"
  modularized: "false"
  tags: "workflow, documentation, jit-loading, context-aware, caching, discovery"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["documentation", "docs", "API reference", "how to", "implement", "best practices", "technology guide", "framework documentation"]
  phases: ["plan", "run", "sync"]
  agents: ["manager-docs", "manager-spec", "expert-backend", "expert-frontend"]
---

## Quick Reference (30 seconds)

Purpose: Load relevant documentation on-demand based on user intent and context.

Primary Tools:

- WebSearch: Find latest documentation and resources online
- WebFetch: Retrieve specific documentation pages
- Context7 MCP: Access official library documentation (when available)
- Read, Grep, Glob: Search local project documentation

Trigger Patterns:

- User asks specific technical questions
- Technology keywords detected in conversation
- Domain expertise required for task completion
- Implementation guidance needed

## Implementation Guide

### Intent Detection

The system recognizes documentation needs through several patterns:

Question-Based Triggers:

- When users ask specific implementation questions (e.g., "how do I implement JWT authentication?")
- When users seek best practices or optimization guidance
- When troubleshooting questions arise

Technology-Specific Triggers:

- Detection of framework names: FastAPI, React, PostgreSQL, Docker, Kubernetes
- Detection of library names: pytest, TypeScript, GraphQL, Redis
- Detection of tool names: npm, pip, cargo, maven

Domain-Specific Triggers:

- Authentication and authorization topics
- Database and data modeling discussions
- Performance optimization inquiries
- Security-related questions

Pattern-Based Triggers:

- Implementation requests: "implement", "create", "build"
- Architecture discussions: "design", "structure", "pattern"
- Troubleshooting: "debug", "fix", "error", "not working"

### Documentation Sources

The system retrieves documentation from multiple sources in priority order:

Local Project Documentation (Highest Priority):

- Check .moai/docs/ for project-specific documentation
- Check .moai/specs/ for requirements and specifications
- Check README.md for project overview
- Check docs/ directory for comprehensive documentation

Official Documentation Sources:

- Use WebFetch to retrieve official framework documentation
- Use Context7 MCP tools when available for library documentation
- Access technology-specific official websites

Community Resources:

- Use WebSearch to find high-quality tutorials
- Search for Stack Overflow solutions with high vote counts
- Find GitHub discussions for specific issues

Real-Time Web Research:

- Use WebSearch with current year for latest information
- Search for recent best practices and updates
- Find new features and deprecation notices

### Loading Strategies

Intent Analysis Process:

- Identify technologies mentioned in user request
- Determine domain areas relevant to the question
- Classify question type (implementation, troubleshooting, conceptual)
- Assess complexity to determine documentation depth needed

Source Prioritization:

- If local documentation exists: Load project-specific docs first
- If official documentation available: Retrieve authoritative sources
- If implementation examples needed: Search community resources
- If latest information required: Perform web research

Context-Aware Caching:

- Cache retrieved documentation within session
- Maintain relevance based on current conversation context
- Remove outdated content when context shifts
- Prioritize frequently accessed documentation

### Quality Assessment

Content Quality Evaluation:

- Authority: Official sources receive highest trust
- Recency: Content within 12 months preferred for fast-moving technologies
- Completeness: Documentation with examples ranked higher
- Relevance: Match between content and user intent

Relevance Ranking:

- Calculate match between documentation content and user question
- Weight authority (30%), recency (25%), completeness (25%), relevance (20%)
- Return highest-scoring documentation first
- Indicate confidence level in retrieved information

### Practical Workflows

Authentication Implementation Workflow:

- When user asks about authentication: Detect technologies (e.g., FastAPI, JWT)
- Identify domains: authentication, security
- Load FastAPI security documentation via WebFetch
- Search for JWT best practices via WebSearch
- Provide comprehensive guidance with source attribution

Database Optimization Workflow:

- When user asks about query performance: Detect database technology
- Identify domain: performance, optimization
- Load official database documentation
- Search for optimization guides and tutorials
- Provide actionable recommendations with sources

New Technology Adoption Workflow:

- When user introduces unfamiliar technology: Detect technology name
- Load official getting started documentation
- Search for migration guides if applicable
- Find integration patterns with existing stack
- Provide strategic adoption guidance

### Error Handling

Network Failures:

- If web search fails: Fall back to cached content
- If WebFetch fails: Use local documentation if available
- Indicate partial results when some sources unreachable

Content Quality Issues:

- If retrieved content seems outdated: Search for newer sources
- If relevance unclear: Ask user for clarification
- If conflicting information found: Present multiple sources with dates

Relevance Mismatches:

- If initial search yields poor results: Refine search query
- If user context unclear: Request clarification before loading
- If documentation gap exists: Acknowledge limitation

### Performance Optimization

Caching Strategy:

- Maintain session-level cache for frequently accessed docs
- Keep project-specific documentation in memory
- Evict stale content based on access time

Efficient Loading:

- Load documentation only when explicitly needed
- Avoid preloading all possible documentation
- Use targeted searches rather than broad queries

Batch Processing:

- Combine related searches when possible
- Group documentation requests by technology
- Process multiple sources in parallel when appropriate

## Advanced Patterns

Multi-Source Aggregation:

- Combine official documentation with community examples
- Cross-reference multiple authoritative sources
- Synthesize comprehensive answers from diverse materials

Context Persistence:

- Remember documentation loaded earlier in conversation
- Avoid redundant loading of same documentation
- Build cumulative knowledge through session

Proactive Loading:

- Anticipate documentation needs based on conversation flow
- Pre-load related topics when discussing complex features
- Suggest relevant documentation before user asks

---

## Works Well With

Agents:

- workflow-docs: Documentation generation
- core-planner: Documentation planning
- workflow-spec: SPEC documentation

Skills:

- moai-docs-generation: Documentation generation
- moai-workflow-docs: Documentation validation
- moai-library-nextra: Nextra documentation

Commands:

- /moai:3-sync: Documentation synchronization
- /moai:9-feedback: Documentation improvements
