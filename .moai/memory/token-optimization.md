# Token Optimization Strategy

Budget and strategy for optimizing Alfred's token usage.

## Token Budget

### Phase-Based Allocation

| Phase | Budget | Purpose |
|-------|--------|---------|
| **SPEC Generation** | 30K | Specification writing and validation |
| **TDD Implementation** | 180K | RED-GREEN-REFACTOR cycle |
| **Documentation** | 40K | API documentation + project report |
| **Total** | 250K | Budget per feature |

---

## `/clear` Execution Rules

**When to execute**:

1. **Immediately after** `/moai:1-plan` completion (mandatory)
   - Saves 45-50K tokens
   - Prepares clean context

2. When context > 150K tokens
   - Prevents overflow
   - Prepares for next phase

3. After 50+ conversation messages
   - Clear accumulated context

---

## Token Usage Monitoring

**Alfred's behavior**:

1. Check current token usage with `/context` command
2. Track usage against phase budget
3. Suggest `/clear` when usage exceeds 150K
4. Report anomalies via `/moai:9-feedback`

---

## File Loading Optimization

### Good Practices ✅

- Load only files needed for current task
- Load file headers first (not full content)
- Selectively load relevant sections
- Cache results

### Bad Practices ❌

- Load entire codebase
- Include node_modules, .git, etc.
- Load images or binary files
- Include historical conversation logs

---

## Model Selection

| Scenario | Choice | Reason |
|----------|--------|--------|
| SPEC generation | Sonnet 4.5 | High-quality design required |
| TDD implementation | Haiku 4.5 | Fast execution, cost savings |
| Security review | Sonnet 4.5 | Precise analysis required |
| Simple edits | Haiku 4.5 | Minimal cost |

**Cost savings**: Haiku 70% cheaper, 60-70% total savings possible

---

## Context Passing Strategy

### Efficient Context Passing

```
Phase1 results → Pass to Phase2 context
   ↓
Extract only needed information
   ↓
Provide to next agent
   ↓
Exclude irrelevant information
```

### Optimal Context Size

- **Minimum**: Only task-essential information
- **Maximum**: Under 50K tokens
- **Target**: 20-30K tokens

---

## Token Saving Tips

1. **Phase separation**: Use `/clear` to separate phases
2. **Selective loading**: Exclude unnecessary files
3. **Result caching**: Preserve reusable information
4. **Concise prompts**: Remove unnecessary explanations
5. **Haiku utilization**: Use Haiku for non-complex tasks

---

## Performance Targets

- SPEC generation: 15-20K tokens
- TDD implementation: 80-100K tokens (per phase)
- Documentation: 20-25K tokens
- **Efficiency**: 60-70% savings vs. manual work

Alfred automatically applies this strategy, and improvements are suggested via `/moai:9-feedback`.
