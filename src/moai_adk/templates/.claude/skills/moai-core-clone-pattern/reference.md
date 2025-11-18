# Clone Pattern Reference

## API Specification

### Core Functions

#### `should_create_clone(task) -> bool`
**Description**: Determine if Clone pattern should be applied based on task characteristics

**Parameters**:
- `task` (Task): Task object with the following attributes:
  - `domain` (str): Task domain (e.g., "ui", "backend", "db", "security", "ml")
  - `steps` (int): Number of steps required
  - `files` (int): Number of files affected
  - `parallelizable` (bool): Whether task can be parallelized
  - `uncertainty` (float): Uncertainty level (0.0-1.0)

**Returns**: `bool` - True if Clone pattern should be used

**Logic**:
```python
return (
    task.domain not in ["ui", "backend", "db", "security", "ml"]
    AND (
        task.steps >= 5
        or task.files >= 100
        or task.parallelizable
        or task.uncertainty > 0.5
    )
)
```

#### `create_clone(task_description, context_scope="full", learning_enabled=True) -> CloneInstance`
**Description**: Create an autonomous Alfred Clone instance

**Parameters**:
- `task_description` (str): Clear, specific task description with defined goals
- `context_scope` (str, optional): Context range - "full" or "domain" (default: "full")
- `learning_enabled` (bool, optional): Whether to save learning memory (default: True)

**Returns**: `CloneInstance` - Independent executable clone

**Clone Capabilities**:
- Full project context access
- All MoAI-ADK tools available
- All 55 Skills accessible
- TRUST 5 principles enforced

### Decision Matrix

| Task Type | Domain Expertise | Steps | Files | Parallelizable | Uncertainty | Use Clone |
|-----------|------------------|-------|-------|----------------|-------------|-----------|
| Large Migration | No | 8+ | 50+ | Yes | High | ✅ |
| UI Component Design | Yes | 3 | 5 | No | Low | ❌ (Use Specialist) |
| Import Refactoring | No | 2 | 200+ | Yes | Medium | ✅ |
| Database Optimization | Yes | 6 | 10 | No | Medium | ❌ (Use Specialist) |
| Architecture Exploration | No | 5+ | 20+ | Yes | High | ✅ |

## Integration Points

### 4-Step Workflow Logic
- **Step 1 (Intent Understanding)**: Clone pattern decision made here
- **Step 2 (Plan Creation)**: Task delegated to clone if pattern applies
- **Step 3 (Task Execution)**: Clone executes autonomously
- **Step 4 (Report & Commit)**: Clone reports back to main Alfred

### TRUST 5 Principles
All clones must adhere to TRUST 5 principles:
1. **Test First**: Implement tests before code
2. **Readable**: Maintain clear, documented code
3. **Unified**: Follow project conventions
4. **Secured**: No security vulnerabilities


## Error Handling

### Clone Failure Recovery
1. **Auto-recovery**: Clone attempts self-debugging first
2. **Context preservation**: Failed state saved for analysis
3. **Main Alfred intervention**: Only if auto-recovery fails
4. **Rollback capability**: Always maintain rollback path

### Memory Management
```python
# Clone learning storage structure
{
  "migration": [
    {
      "timestamp": "2025-11-05T10:30:00Z",
      "success": true,
      "approach_used": "incremental_migration",
      "pitfalls_discovered": ["config_conflicts", "hook_compatibility"],
      "optimization_tips": ["pre_validation", "backup_critical_files"]
    }
  ]
}
```

## Performance Considerations

### Resource Usage
- **Token cost**: Clones use tokens like any Claude session
- **Time efficiency**: Parallel processing reduces overall time
- **Memory**: Learning storage is minimal (JSON structure)

### Optimization Tips
1. **Batch similar tasks**: Create clones for related work items
2. **Limit concurrent clones**: 3-4 clones maximum for resource efficiency
3. **Reuse learnings**: Check memory before creating new clones
4. **Cleanup old data**: Remove learnings older than 6 months

## Security & Safety

### Clone Constraints
- Cannot modify clone creation logic itself
- Must maintain TRUST 5 principles
- Cannot access other clone sessions
- Limited to project-scoped actions

### Safety Mechanisms
1. **Context isolation**: Each clone has isolated context
2. **Permission inheritance**: Clones inherit main Alfred's permissions
4. **Emergency stop**: Main Alfred can terminate any clone

## Troubleshooting

### Common Issues

**Issue**: Clone not responding
**Solution**: Check context scope, recreate with broader permissions

**Issue**: Clone making wrong decisions
**Solution**: Review task description clarity, add specific constraints

**Issue**: Parallel clones conflicting
**Solution**: Use non-overlapping file scopes, implement coordination

**Issue**: Learning not saving
**Solution**: Check .moai/memory/ directory permissions

### Debug Commands
```bash
# Check clone learnings
cat .moai/memory/clone-learnings.json

# Verify clone permissions
ls -la .moai/

```
