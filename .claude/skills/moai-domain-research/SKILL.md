---
name: moai-domain-research
description: >
  Market and ecosystem research specialist for /moai brain Phase 3. Executes parallel
  WebSearch + Context7 queries, handles tool failures gracefully, and produces structured
  research.md artifacts with cited sources and research limitations.

when_to_use: >
  Use for /moai brain Phase 3 market and ecosystem research: parallel
  WebSearch and Context7 queries, competitive-landscape analysis, source
  citation, and structured research.md artifacts with graceful
  tool-failure handling.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-05-04"
  modularized: "false"
  tags: "research, web-search, context7, market-analysis, parallel-tools, brain"
  related-skills: "moai-domain-ideation, moai-foundation-thinking"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

<!-- Verifies: Parallel WebSearch + Context7 in single message for Phase 3 -->

# Research Domain Specialist

Parallel research executor for the brain workflow's Phase 3. Issues WebSearch and Context7 tool calls simultaneously (single-message parallel call pattern per Anthropic best practice), handles partial failures gracefully, and assembles a structured `research.md` artifact.

## Quick Reference

Core responsibilities:
- Execute WebSearch + Context7 in parallel (single message, multiple tool calls)
- Handle tool failures gracefully (partial-result tolerance)
- Produce `research.md` with cited sources and explicit Research Limitations section
- Stay language/technology neutral

Key guarantees:
- [HARD] Tool calls are issued in parallel (single Claude message), not sequentially
- [HARD] Research failure does NOT abort Phase 3 — partial results are acceptable
- [HARD] Every source in research.md has a citation (URL or tool reference)
- [HARD] Research Limitations section present when any tool call fails or returns empty

---

## Phase 3: Research Execution

### Input

- Clarity-scored idea with user context from Phase 1
- Diverged concept map from Phase 2 (in-memory)
- Optional: existing `.moai/project/tech.md` for tech-stack context (read-only)

### Step 1: Query Design

Before issuing tool calls, design 2-4 targeted queries for each tool type:

**WebSearch query design principles**:
- Search for: existing solutions, market size, user pain points, competitive landscape
- Include both broad queries ("habit tracking apps market") and targeted queries ("habit tracking for seniors accessibility challenges")
- One query per major angle from Phase 2 diverge (use top 3 angles)
- Avoid technology-specific queries unless the user explicitly constrained to a tech stack

**Context7 query design principles**:
- Search for: relevant libraries, frameworks, or platforms in the solution space
- Focus on ECOSYSTEM tools (not specific language implementations) — e.g., query "habit tracking SDK" not "habit tracking React library"
- Use `resolve-library-id` first, then `get-library-docs` for top matches

### Step 2: Parallel Tool Calls

[HARD] Issue ALL prepared tool calls in a SINGLE Claude message. This is the parallel tool call pattern documented in Anthropic's tool-use documentation.

Pattern (pseudocode — actual tool syntax per Claude Code):
```
[Single message containing multiple tool_use blocks]
  WebSearch("habit tracking apps market size")
  WebSearch("senior citizen mobile app accessibility best practices")
  WebSearch("habit formation psychology research")
  mcp__context7__resolve-library-id("habit tracking")
```

The single-message parallel call is 50-70% faster than sequential calls and is the canonical pattern for independent tool calls.

### Step 3: Failure Handling

After tool calls return, assess results:

| Scenario | Behavior |
|----------|----------|
| All tools succeed | Proceed with full results |
| WebSearch fails, Context7 succeeds | Continue — note WebSearch failure in Research Limitations |
| Context7 fails, WebSearch succeeds | Continue — note Context7 failure in Research Limitations |
| Both fail | Continue with empty sources — add prominent Research Limitations note |
| Partial WebSearch results (some queries empty) | Use available results — note missing queries in Research Limitations |

[HARD] Do NOT abort Phase 3 under any tool failure scenario. A research.md with only a Research Limitations section is valid output.

### Step 4: Source Processing

For each successful WebSearch result:
1. Extract URL, title, and a 1-2 sentence summary of relevance
2. Categorize: market_data, user_research, competitor, technical_ecosystem, case_study
3. Discard results that are clearly off-topic (wrong domain, wrong problem space)

For each Context7 library result:
1. Note library name, version, and primary capability
2. Extract key features relevant to the idea
3. Note the ecosystem (not language-specific) context

### Step 5: Synthesis

After processing sources, synthesize findings into 3-5 thematic areas:
1. **Market landscape**: Size, growth, existing players, gaps
2. **User needs**: Pain points, use cases, validated problems
3. **Technical ecosystem**: Available tools, standards, building blocks (language-neutral)
4. **Risk signals**: Competitive threats, regulatory concerns, technical complexity
5. **Opportunities**: Unaddressed needs, timing factors, differentiation angles

### Output Format

Write `research.md` to `.moai/brain/IDEA-NNN/`:

```markdown
# Research: {idea summary}
*Phase 3 — Brain Workflow | Date: {date} | Idea: IDEA-NNN*

## Executive Summary

{2-3 sentences: what was learned and what it implies for the idea}

## Market Landscape

{Findings about existing solutions, market size, competitive dynamics}

Sources:
- [{source title}]({URL}): {1-sentence relevance}
- ...

## User Needs

{Validated user problems, use cases, and success patterns from research}

Sources:
- [{source title}]({URL}): {1-sentence relevance}

## Technical Ecosystem

{Language-neutral overview of available tools, platforms, and standards relevant to the idea}

Sources:
- [{source title/library}]({URL or context7 reference}): {1-sentence relevance}

## Risk Signals

{Competitive threats, known failure patterns, regulatory or technical risks}

## Opportunities

{Gaps in existing solutions, timing factors, differentiation levers}

## Sources Summary

| Source | Type | Relevance |
|--------|------|-----------|
| {title} | {market_data/user_research/competitor/technical_ecosystem/case_study} | {brief note} |
...
Total sources: {N}

## Research Limitations

{Present ONLY if any tool call failed or returned empty. Omit this section if all tools succeeded.}

{Examples:}
- WebSearch was unavailable during this session. Market data may be incomplete.
- Context7 returned no results for "habit tracking". Technical ecosystem section is based on WebSearch only.
- WebSearch query "{query}" returned zero results. Competitive landscape may have gaps.
```

---

## Technology Neutrality in Research

[HARD] The Technical Ecosystem section must describe capabilities and tools at the ecosystem level, not language level:

- Correct: "Push notification platforms (Firebase, OneSignal, APNS/FCM) support both mobile and web targets"
- Wrong: "Firebase Cloud Messaging SDK for React Native is the standard approach"

The user will choose their tech stack during `/moai project` and `/moai plan`. Research should inform the choice, not make it.

---

## Parallel Call Evidence

When the session transcript is inspected, Phase 3 research MUST show multiple tool_use blocks in a single assistant turn. This is verifiable evidence that parallel calls were issued:

```
Turn N (assistant):
  <tool_use id="a1">WebSearch("query 1")</tool_use>
  <tool_use id="a2">WebSearch("query 2")</tool_use>
  <tool_use id="a3">mcp__context7__resolve-library-id("library")</tool_use>
```

Sequential calls (one tool per turn) violate the parallel-tool-call requirement and should be avoided.

---

## GLM-Backend Note

When the session runs on the GLM backend (`moai glm` / `moai cg` GLM panes), `WebSearch` and `WebFetch` route to the z.ai MCP tools (`mcp__web_search_prime__webSearchPrime` / `mcp__web_reader__webReader`) instead of the built-in tools — see `.claude/rules/moai/core/glm-web-tooling.md` for the HARD routing table.

---

## Works Well With

- `moai-domain-ideation`: Research findings feed into Phase 4 Converge context for more grounded Lean Canvas
- `moai-workflow-brain`: Orchestrates Phase 3 execution with proper IDEA-NNN directory management
- `moai-foundation-thinking`: Critical evaluation in Phase 5 uses research findings as evidence
- `/deep-research <question>`: The bundled deep-research workflow is a heavier multi-source path than this skill's own parallel WebSearch + Context7 — it fans out web searches, cross-checks sources against each other, votes on contested claims, and returns a cited report. It requires the WebSearch tool; a workflow run spends meaningfully more tokens than a single-pass search; and the AskUserQuestion boundary holds — the orchestrator collects and refines the research question before launch, never mid-run.

---

## Common Rationalizations

| Rationalization | Reality |
|----------------|---------|
| "Sequential calls are safer for avoiding rate limits" | Parallel calls are the Anthropic-recommended pattern. Rate limits are handled at the API layer, not by serializing tool calls. |
| "I should abort if WebSearch fails because research will be incomplete" | Partial research is better than no research. A Research Limitations section communicates the gap clearly. |
| "Context7 libraries are language-specific, so I should filter by the user's language" | Research describes the ecosystem — all relevant tools regardless of implementation language. Tech selection is deferred to /moai plan. |
| "2 WebSearch queries are enough" | Use 2-4 per major angle. Undersampling misses competitive landscape gaps. |

## Verification

- [ ] Tool calls appear in parallel (multiple tool_use blocks in single assistant turn)
- [ ] research.md was produced regardless of tool failure scenarios
- [ ] All cited sources have URLs or tool references
- [ ] Technical Ecosystem section contains no language/framework-specific prescriptions
- [ ] Research Limitations section present if any tool call failed
- [ ] Executive Summary in research.md has at least 2 sentences
