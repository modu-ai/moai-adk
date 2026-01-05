# Advanced Agent Patterns - Anthropic Engineering Insights

Sources:
- https://www.anthropic.com/engineering/effective-harnesses-for-long-running-agents
- https://www.anthropic.com/engineering/advanced-tool-use
- https://www.anthropic.com/engineering/code-execution-with-mcp
- https://www.anthropic.com/engineering/building-agents-with-the-claude-agent-sdk
- https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents
- https://www.anthropic.com/engineering/writing-tools-for-agents
- https://www.anthropic.com/engineering/multi-agent-research-system
- https://www.anthropic.com/engineering/claude-code-best-practices
- https://www.anthropic.com/engineering/claude-think-tool
- https://www.anthropic.com/engineering/building-effective-agents
- https://www.anthropic.com/engineering/contextual-retrieval

Updated: 2026-01-06

## Long-Running Agent Architecture

### Two-Agent Pattern

For complex, multi-session tasks, use a two-agent system:

Initializer Agent (runs once):
- Sets up project structure and environment
- Creates feature registry tracking completion status
- Establishes progress documentation patterns
- Generates initialization scripts for future sessions

Executor Agent (runs repeatedly):
- Consumes environment created by initializer
- Works on single features per session
- Updates progress documentation
- Maintains feature registry state

### Feature Registry Pattern

Maintain a JSON file tracking all functionality:

```json
{
  "features": [
    {"id": "auth-login", "status": "complete", "tested": true},
    {"id": "auth-logout", "status": "in-progress", "tested": false},
    {"id": "user-profile", "status": "pending", "tested": false}
  ]
}
```

This enables:
- Clear work boundaries per session
- Progress tracking across sessions
- Prioritization of incomplete features

### Progress Documentation

Create persistent progress logs (claude-progress.txt):
- Summary of completed work
- Current feature status
- Blockers and decisions made
- Next steps for future sessions

Commit progress with git for history preservation.

### Session Initialization Protocol

At start of each session:
1. Verify correct directory
2. Review progress logs
3. Select priority feature from registry
4. Test existing baseline functionality
5. Begin focused work on single feature

## Dynamic Tool Discovery

### Tool Search Pattern

For large tool libraries, implement discovery mechanism:

Benefits:
- 85% reduction in token consumption
- Tools loaded only when needed
- Reduced context pollution

Implementation approach:
- Register tools with metadata including name, description, and keywords
- Provide search tool that queries registry
- Use defer_loading parameter to hide tools until searched
- Agent searches for relevant tools before use

### Programmatic Tool Orchestration

For complex multi-step workflows:

Benefits:
- 37% token reduction on complex tasks
- Elimination of repeated inference passes
- Parallel operation execution

Pattern:
- Agent generates code orchestrating multiple tool calls
- Code executes in sandbox environment
- Results returned to agent in single response

### Usage Examples for Tool Clarity

JSON schemas alone are insufficient. Provide 3-5 concrete examples:

Minimal invocation: Required parameters only
Partial invocation: Common optional parameters
Complete invocation: All parameters with edge cases

Examples teach API conventions without token overhead.

## Code Execution Efficiency

### Data Processing in Sandbox

Process data before model sees results:

Benefits:
- 98.7% token reduction possible (150K to 2K tokens)
- Deterministic operations executed reliably
- Complex transformations handled efficiently

Pattern:
- Agent writes filtering and aggregation code
- Code executes in sandboxed environment
- Only relevant results returned to model
- Intermediate results persisted for resumable workflows

### Reusable Skills Pattern

Save working code as functions:
- Extract successful patterns into reusable modules
- Reference modules in future sessions
- Build library of proven implementations

## Multi-Agent Coordination

### Orchestrator-Worker Architecture

Lead Agent (higher capability model):
- Analyzes incoming queries
- Decomposes into parallel subtasks
- Spawns specialized worker agents
- Synthesizes results into final output

Worker Agents (cost-effective models):
- Execute specific, focused tasks
- Return condensed summaries (1K-2K tokens)
- Operate with isolated context windows
- Use specialized prompts and tool access

### Hierarchical Communication

Lead to workers:
- Clear task boundaries
- Specific output format requirements
- Guidance on tools and sources
- Prevention of duplicate work

Workers to lead:
- Condensed findings summary
- Source attribution
- Quality indicators
- Error or blocker reports

### Scaling Rules

Simple queries: Single agent with 3-10 tool calls
Complex research: 10+ workers with parallel execution
State persistence: Prevent disruption during updates
Error resilience: Adapt when tools fail rather than restart

## Context Engineering

### Core Principle

Find the smallest possible set of high-signal tokens that maximize likelihood of desired outcome. Treat context as finite, precious resource.

### Information Prioritization

LLMs lose focus as context grows (context rot). Every token depletes attention budget.

Strategies:
- Place critical information at start and end of context
- Use clear section markers (XML tags or Markdown headers)
- Remove redundant or low-signal content
- Summarize when precision not required

### Context Compaction

For long-running tasks:
- Summarize conversation history automatically
- Reinitiate with compressed context
- Preserve architectural decisions and key findings
- Maintain external memory files outside context window

### Just-In-Time Retrieval

Maintain lightweight identifiers and load data dynamically:
- Store file paths, URLs, and IDs
- Load content only when needed
- Combine upfront retrieval for speed with autonomous exploration
- Progressive disclosure mirrors human cognition

## Tool Design Best Practices

### Consolidation Over Proliferation

Combine related functionality into single tools:

Instead of: list_users, list_events, create_event, delete_event
Use: manage_events with action parameter

Benefits:
- Reduced tool selection complexity
- Clearer mental model for agent
- Lower probability of incorrect tool choice

### Context-Aware Responses

Return high-signal information:
- Use natural language names rather than cryptic IDs
- Include relevant metadata in responses
- Format for agent consumption, not human reading

### Parameter Specification

Clear parameter naming:
- user_id not user
- start_date not start
- include_archived not archived

Enable response format control:
- Optional enum for concise or detailed responses
- Agent specifies verbosity based on task needs

### Error Handling

Replace opaque error codes with instructive feedback:
- Explain what went wrong
- Suggest correct usage
- Provide examples of valid parameters
- Encourage token-efficient strategies

### Poka-Yoke Design

Make incorrect usage harder than correct usage:
- Validate parameters before execution
- Return helpful errors for invalid combinations
- Design APIs that guide toward success

## Think Tool Integration

### When to Use Think Tool

High-value scenarios:
- Processing complex tool outputs before proceeding
- Compliance verification with detailed guidelines
- Sequential decision-making where errors are consequential
- Multi-step domains requiring careful consideration

### Performance Characteristics

Measured improvements:
- Airline domain: 54% relative improvement with targeted examples
- Retail scenarios: 81.2% pass-rate
- SWE-bench: 1.6% average improvement

### Implementation Strategy

Pair with optimized domain-specific prompts
Place comprehensive instructions in system prompts
Avoid for non-sequential or simple tasks
Use for reflecting on tool outputs mid-response

## Verification Patterns

### Quality Assurance Approaches

Code verification: Linting and static analysis most effective
Visual feedback: Screenshot outputs for UI tasks
LLM judgment: Fuzzy criteria evaluation (tone, quality)
Human evaluation: Edge cases automation misses

### Diagnostic Questions

When agents underperform:

Missing context? Restructure search APIs for discoverability
Repeated failures? Add formal validation rules in tool definitions
Error-prone approach? Provide alternative tools enabling different strategies
Variable performance? Build representative test sets for programmatic evaluation

## Workflow Pattern: Explore-Plan-Code-Commit

### Phase 1: Explore

Start with exploration without coding:
- Read files to understand structure
- Identify relevant components
- Map dependencies and interfaces

### Phase 2: Plan

Use extended thinking prompts:
- Outline approach before implementation
- Consider alternatives and tradeoffs
- Define clear success criteria

### Phase 3: Code

Implement iteratively:
- Small, testable changes
- Verify each step before proceeding
- Handle edge cases explicitly

### Phase 4: Commit

Meaningful commits:
- Descriptive messages explaining why
- Logical groupings of related changes
- Clean history for future reference

## Hybrid Context Retrieval

### Combined Approach

Semantic embeddings: Capture meaning relationships
BM25 keyword search: Handle exact phrases and error codes

### Context Prepending

Enrich chunks with metadata before encoding:
- Transform isolated statements into fully-contextualized information
- Include surrounding context and relationships
- Improves retrieval precision by 49-67%

### Configuration

Optimal settings from research:
- Top-20 chunks outperform smaller selections
- Domain-specific prompts improve quality
- Reranking adds significant precision gains

## Security Considerations

### Credential Handling

Web-based execution:
- Credentials never enter sandbox
- Proxy services handle authenticated operations
- Branch-level restrictions enforced externally

### Sandboxing Architecture

Dual-layer protection:
- Filesystem isolation: Read/write boundaries
- Network isolation: Domain allowlists via proxy

OS-level enforcement using kernel security features.

### Permission Boundaries

84% reduction in permission prompts through:
- Defined operation boundaries
- Automatic allowlisting of safe operations
- Clear separation of privileged actions
