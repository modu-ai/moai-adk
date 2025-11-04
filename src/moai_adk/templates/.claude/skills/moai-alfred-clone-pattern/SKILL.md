---
name: moai-alfred-clone-pattern
version: 1.0.0
created: 2025-11-05
updated: 2025-11-05
status: active
description: Master-Clone pattern implementation guide for complex multi-step tasks with full project context
keywords: ['clone-pattern', 'master-clone', 'delegation', 'multi-step', 'parallel-processing']
allowed-tools:
  - Read
  - Bash
  - Task
---

# Clone Pattern Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-clone-pattern |
| **Version** | 1.0.0 (2025-11-05) |
| **Allowed tools** | Read, Bash, Task |
| **Auto-load** | On demand when complex tasks detected |
| **Tier** | Alfred |

---

## What It Does

Provides comprehensive guidance for Alfred's Master-Clone pattern - a delegation mechanism where Alfred creates autonomous clones to handle complex multi-step tasks that don't require domain-specific expertise but benefit from full project context and parallel processing capabilities.

## When to Use

**Use Clone Pattern when:**
- Task requires 5+ steps OR affects 100+ files
- No domain-specific expertise needed (UI, Backend, DB, Security, ML)
- Task is complex with high uncertainty
- Parallel processing would be beneficial
- Full project context is required for optimal decisions

**Examples:**
- Large-scale migrations (v0.14.0 → v0.15.2)
- Refactoring across many files (100+ imports changes)
- Parallel exploration/evaluation tasks
- Complex architecture restructuring

## Key Concepts

### Master-Clone Architecture
```
Main Alfred Session
    │
    ├─ Intent analysis
    ├─ Task classification (domain/complexity)
    │
    └─ Clone creation (if applicable)
        │
        └─ Clone Instance
            ├─ Full project context
            ├─ All tool access permissions
            ├─ All Skills loaded
            ├─ Specific task description only
            └─ Autonomous execution & learning
```

### Clone vs Specialist Selection

| Decision Factor | Clone Pattern | Lead-Specialist Pattern |
|-----------------|---------------|-------------------------|
| Domain expertise needed | ❌ No | ✅ Yes |
| Context scope | Full project | Domain only |
| Autonomy level | Complete autonomous | Follows instructions |
| Parallel execution | ✅ Possible | ❌ Sequential only |
| Learning capability | Self-memory storage | Feedback-based |
| Best for | Long multi-step tasks | Specialized tasks |

## Implementation Rules

### Rule 1: Clone Creation Conditions
```python
def should_create_clone(task) -> bool:
    """Determine if Clone pattern should be applied"""
    return (
        # No domain specialization needed AND
        task.domain not in ["ui", "backend", "db", "security", "ml"]
        
        # AND meets one of these criteria:
        AND (
            task.steps >= 5                    # 5+ steps
            or task.files >= 100               # 100+ files
            or task.parallelizable             # Can be parallelized
            or task.uncertainty > 0.5          # High uncertainty
        )
    )
```

### Rule 2: Clone Creation Method
```python
def create_clone(
    task_description: str,
    context_scope: str = "full",
    learning_enabled: bool = True
) -> CloneInstance:
    """Create Alfred Clone instance
    
    Args:
        task_description: Specific task description (clear goals)
        context_scope: Context range ("full" | "domain")
        learning_enabled: Whether to save learning memory
        
    Returns:
        Independent executable Clone instance
    """
    clone = Task(
        subagent_type="general-purpose",
        description=f"Clone: {task_description}",
        prompt=f"""
You are an Alfred Clone with full MoAI-ADK capabilities.

TASK: {task_description}

CONTEXT:
- Full project context loaded
- All .moai/ configuration available
- All 55 Skills accessible
- Same tools as Main Alfred
- Same TRUST 5 principles enforced

EXECUTION:
1. Plan your approach
2. Execute with transparency
3. Document decisions via @TAG
4. Create PR if modifications needed
5. Log learnings to clone-memory

SUCCESS CRITERIA:
- TRUST 5 principles maintained
- @TAG chain integrity preserved
- All tests passing
- PR ready for review

You have full autonomy. Main Alfred will review your output only.
"""
    )
    return clone
```

## Usage Examples

### Example 1: Large-Scale Migration
```python
# Alfred's analysis
task = UserRequest(
    type="migration",
    scope="large-scale",
    steps=8,  # > 5 steps
    domains=["config", "hooks", "permissions"],
    uncertainty="high"  # New structure transition
)

# Apply Clone pattern
if should_create_clone(task):
    clone = create_clone(
        "Migrate v0.14.0 config structure to v0.15.2"
    )
    clone.execute()
```

### Example 2: Parallel Processing Task
```python
# Alfred's analysis
task = UserRequest(
    type="exploration_evaluation",
    items=["UI/UX redesign", "Backend optimization", "DB migration"],
    independence="high"  # Each item independent
)

# Apply Clone pattern for parallel execution
if task.independence > 0.7:
    clones = [
        create_clone(f"Evaluate: {item}")
        for item in task.items
    ]
    results = parallel_execute(clones)
```

## Learning System

Clones save learnings to improve future similar tasks:

```python
def save_learning(task_type: str, learnings: dict):
    """Save Clone learning to memory"""
    memory_file = Path(".moai/memory/clone-learnings.json")
    
    learnings_db = json.loads(memory_file.read_text())
    learnings_db[task_type].append({
        "timestamp": now(),
        "success": True/False,
        "approach_used": "...",
        "pitfalls_discovered": [...],
        "optimization_tips": [...]
    })
    
    memory_file.write_text(json.dumps(learnings_db, indent=2))
```

## Benefits

1. **Context Preservation**: Full project context vs domain-only
2. **Maximum Autonomy**: Goal-oriented vs instruction-oriented
3. **Parallel Scalability**: Multiple clones can run simultaneously
4. **Self-Learning**: Accumulated experience for future tasks

## Integration with Alfred Workflow

- Activated during 4-Step Workflow Logic (Step 1: Intent Understanding)
- Integrates with TRUST 5 principles
- Maintains @TAG chain integrity
- Works seamlessly with existing GitFlow

## References

- CLAUDE.md: Alfred's Hybrid Architecture
- Skill("moai-alfred-workflow"): 4-Step Workflow Logic
- Skill("moai-alfred-agent-guide"): 19 team members details
