# Token Optimization Strategy

## Overview

This document defines token optimization strategies for MoAI-ADK to maintain efficiency within Claude Code's context limits while maximizing development productivity.

## Token Budget Management

### Phase-based Allocation

**SPEC Creation Phase**: 30K tokens maximum
- Load only essential skills (6 maximum)
- Focus on requirement analysis and specification
- Immediate `/clear` execution after completion

**TDD Implementation Phase**: 180K tokens maximum
- RED phase: 60K tokens (test writing)
- GREEN phase: 60K tokens (implementation)
- REFACTOR phase: 60K tokens (optimization)

**Documentation Phase**: 40K tokens maximum
- Final documentation generation
- Quality gate validation
- Release preparation

**Total Budget**: 250K tokens per complete feature cycle

### Context Optimization Rules

**Mandatory Clear Operations**:
- After `/moai:1-plan` completion (saves 45-50K tokens)
- Before starting new complex features
- When context exceeds 150K tokens
- Every 50+ messages in conversation

**Selective Loading Strategy**:
- Load only relevant files for current task
- Use `Task()` context passing between phases
- Avoid loading entire codebase unnecessarily

## Skill Loading Optimization

### Phase-specific Skill Filters

**SPEC Creation Skills** (6 maximum):
```
1. moai-foundation-ears
2. moai-foundation-specs
3. moai-foundation-trust
4. moai-core-context-optimization
5. moai-essentials-review
6. moai-lang-{detected_language}
```

**TDD RED Phase Skills** (6 maximum):
```
1. moai-domain-testing
2. moai-foundation-trust
3. moai-essentials-review
4. moai-core-code-reviewer
5. moai-essentials-debug
6. moai-lang-{implementation_language}
```

**TDD GREEN Phase Skills** (3 maximum):
```
1. moai-lang-{implementation_language}
2. moai-domain-{frontend/backend}
3. moai-essentials-review
```

**TDD REFACTOR Phase Skills** (4 maximum):
```
1. moai-essentials-refactor
2. moai-essentials-review
3. moai-core-code-reviewer
4. moai-essentials-debug
```

### Skill Priority Matrix

**High Priority** (Always Load):
- Foundation skills (ears, specs, trust)
- Language-specific skills
- Domain-relevant skills

**Medium Priority** (Load as Needed):
- Implementation patterns
- Testing frameworks
- Debugging tools

**Low Priority** (Load Rarely):
- Specialized integration skills
- Advanced optimization techniques
- Legacy system patterns

## Context Management Patterns

### JIT (Just-In-Time) Loading

**File Loading Strategy**:
1. Load only entry points initially (main.py, __init__.py)
2. Identify relevant modules based on task
3. Load specific sections only (not entire files)
4. Cache context in `Task()` for reuse
5. Clear irrelevant context between phases

**Directory Scanning**:
```
# Efficient approach
1. Scan directory structure only
2. Load file headers and imports
3. Identify relevant files for current task
4. Load specific file sections
5. Cache for future reuse
```

### Context Passing Patterns

**Between Agents**:
```python
# Efficient context passing
result = await Task(
    subagent_type="api-designer",
    prompt="Design API for user authentication",
    context={
        "spec_id": "SPEC-001",
        "requirements_summary": key_requirements,
        "technical_constraints": constraints
    }
)

# Pass result to next agent
implementation = await Task(
    subagent_type="backend-expert",
    prompt=f"Implement: {result.design}",
    context={"previous_design": result.design}
)
```

**Between Phases**:
- Store only essential results between phases
- Use file-based persistence for large data
- Create checkpoint references for context restoration

## Model Selection Strategy

### Sonnet 4.5 Usage (High-Cost, High-Performance)

**Use Cases**:
- SPEC creation and architecture decisions
- Security reviews and code analysis
- Complex problem-solving
- Multi-agent coordination

**Token Budget**: 50K tokens per session

### Haiku 4.5 Usage (Low-Cost, High-Speed)

**Use Cases**:
- Code exploration and file search
- Simple modifications and fixes
- Test execution and validation
- Status checks and monitoring

**Token Budget**: 20K tokens per session

### Cost Optimization

**Cost per 1K tokens**:
- Sonnet 4.5: $0.003
- Haiku 4.5: $0.0008 (70% cheaper)

**Savings Strategy**:
- Use Haiku for 70% of operations
- Reserve Sonnet for 30% critical operations
- Achieve 60-70% overall cost reduction

## Memory Management

### Session State Persistence

**Store Between Sessions**:
- Project metadata and configuration
- Active task status and checkpoints
- Context optimization settings
- Agent execution history

**Clear Between Sessions**:
- Conversation history
- Temporary file contexts
- Cached code analysis
- Debug information

### Checkpoint Creation

**Create Checkpoints At**:
- Major phase completions
- Before complex operations
- After successful implementations
- Before context clearing

**Checkpoint Content**:
```
{
  "phase": "SPEC_CREATION_COMPLETE",
  "spec_id": "SPEC-001",
  "timestamp": "2025-11-20T07:30:00Z",
  "token_usage": 28500,
  "next_phase": "TDD_IMPLEMENTATION",
  "context_summary": "User authentication system with JWT"
}
```

## Performance Optimization

### Response Time Targets

**Fast Operations** (< 2 seconds):
- File search and navigation
- Simple code generation
- Status checks
- Context loading

**Standard Operations** (2-5 seconds):
- SPEC creation
- Code implementation
- Documentation generation
- Quality validation

**Complex Operations** (5-10 seconds):
- Multi-agent coordination
- Comprehensive analysis
- Complex refactoring
- Integration testing

### Optimization Techniques

**Caching Strategy**:
- Cache frequently accessed file contents
- Store analysis results for reuse
- Maintain skill loading cache
- Preserve agent delegation results

**Parallel Processing**:
- Execute multiple agents concurrently when possible
- Parallel file analysis for large codebases
- Concurrent testing and validation
- Simultaneous documentation generation

## Monitoring and Analytics

### Token Usage Tracking

**Metrics to Monitor**:
- Tokens used per command
- Tokens used per phase
- Context efficiency ratio
- Agent delegation success rate

**Tracking Implementation**:
```python
# Token usage tracking
token_metrics = {
    "command": "/moai:2-run SPEC-001",
    "tokens_used": 45230,
    "context_efficiency": 0.78,
    "agent_delegations": 3,
    "duration_seconds": 12.5
}
```

### Performance Alerts

**Alert Thresholds**:
- Token usage > 80% of phase budget
- Response time > 10 seconds
- Context efficiency < 0.5
- Agent delegation failure rate > 10%

**Alert Actions**:
- Automatic `/clear` execution
- Context optimization recommendations
- Performance improvement suggestions
- Escalation to human operator

## Optimization Checklist

### Pre-execution Validation

**Before Starting Task**:
- [ ] Check available token budget
- [ ] Verify context is clean (execute `/clear` if needed)
- [ ] Load only essential skills
- [ ] Set appropriate model (Sonnet/Haiku)

**During Execution**:
- [ ] Monitor token usage continuously
- [ ] Execute `/clear` at thresholds
- [ ] Use efficient file loading patterns
- [ ] Cache reusable results

**Post-execution Review**:
- [ ] Analyze token usage efficiency
- [ ] Document optimization opportunities
- [ ] Update skill loading patterns
- [ ] Refine context management

### Quality Assurance

**Token Optimization Quality Gates**:
1. Token usage within budget for phase
2. Context efficiency > 0.7
3. Response time within targets
4. Agent delegation success rate > 95%
5. No context overflow occurrences

## Advanced Optimization Strategies

### Context Compression

**Techniques**:
- Summarize long code sections
- Use abstraction for complex concepts
- Create reference pointers instead of full content
- Implement progressive disclosure

### Predictive Loading

**Strategy**:
- Analyze user patterns to predict next actions
- Pre-load likely required skills and context
- Maintain warm cache for frequent operations
- Optimize for common workflows

### Dynamic Budget Adjustment

**Implementation**:
- Adjust token budgets based on task complexity
- Scale context size based on project size
- Adapt model selection based on performance requirements
- Optimize for specific user patterns

## Troubleshooting

### Common Issues

**Token Limit Exceeded**:
- Symptoms: Context truncation, slow responses
- Solution: Execute `/clear` immediately, restart with minimal context

**Poor Performance**:
- Symptoms: Slow response times, high latency
- Solution: Reduce skill loading, use Haiku model, optimize file access

**Context Inefficiency**:
- Symptoms: Low context efficiency ratio (< 0.5)
- Solution: Improve file loading strategy, use JIT loading

### Recovery Procedures

**Emergency Reset**:
1. Execute `/clear` to reset context
2. Restart from last checkpoint
3. Load minimal required context
4. Continue with optimized approach

**Performance Recovery**:
1. Analyze bottleneck in current execution
2. Implement immediate optimization
3. Monitor improvement metrics
4. Document learned patterns

This optimization strategy ensures maximum efficiency while maintaining high-quality development output within Claude Code's constraints.