# Token Budget Management

Allocate and manage token budgets across session components.

## Budget Allocation Framework

### Standard Allocation

```python
allocations = {
    "system_context": 0.10,      # 10% - System prompts and configuration
    "working_memory": 0.30,      # 30% - Active context
    "knowledge_base": 0.20,      # 20% - Skills and documentation
    "agent_context": 0.15,       # 15% - Agent states
    "interaction_buffer": 0.25   # 25% - Conversation history
}
```

### Dynamic Adjustment

Allocations adjust based on usage patterns:

```python
def adjust_allocations(usage_history: dict) -> dict:
    '''Adjust allocations based on historical usage.'''
    
    adjusted = allocations.copy()
    
    # If agent context frequently exhausted
    if usage_history["agent_context_exhaustion"] > 3:
        adjusted["agent_context"] = 0.20
        adjusted["interaction_buffer"] = 0.20
    
    # If knowledge base underutilized
    if usage_history["knowledge_base_utilization"] < 0.50:
        adjusted["knowledge_base"] = 0.15
        adjusted["working_memory"] = 0.35
    
    return adjusted
```

---

## Component Cleanup Strategies

### Working Memory Cleanup

```python
def cleanup_working_memory():
    '''Free up working memory tokens.'''
    
    # Compress large inactive contexts
    freed = compress_inactive_contexts()
    
    # Remove old debugging information
    freed += remove_debug_info()
    
    # Archive completed task contexts
    freed += archive_completed_tasks()
    
    return freed
```

### Knowledge Base Cleanup

```python
def cleanup_knowledge_base():
    '''Free up knowledge base tokens.'''
    
    # Unload unused skill modules
    freed = unload_unused_skills()
    
    # Cache documentation excerpts instead of full docs
    freed += cache_documentation()
    
    # Remove duplicate knowledge
    freed += deduplicate_knowledge()
    
    return freed
```

### Interaction Buffer Cleanup

```python
def cleanup_interaction_buffer():
    '''Free up conversation history tokens.'''
    
    # Summarize old interactions
    freed = summarize_old_interactions(
        older_than=timedelta(hours=2),
        keep_full=timedelta(minutes=30)
    )
    
    # Remove redundant exchanges
    freed += remove_redundant_exchanges()
    
    # Archive to long-term storage
    freed += archive_interactions()
    
    return freed
```

---

## Budget Monitoring

### Token Usage Tracking

```python
class TokenBudgetManager:
    def allocate_tokens(self, component: str, requested: int) -> int:
        '''Allocate tokens with cleanup fallback.'''
        
        max_allowed = int(self.total_budget * self.allocations[component])
        available = max_allowed - self.current_usage[component]
        
        if requested <= available:
            self.current_usage[component] += requested
            return requested
        
        # Try cleanup
        freed = self.cleanup_component(component)
        
        # Retry allocation
        return self.allocate_tokens(component, requested)
    
    def get_utilization_report(self) -> dict:
        '''Get current token utilization.'''
        
        return {
            component: {
                "used": self.current_usage[component],
                "allocated": int(self.total_budget * self.allocations[component]),
                "utilization%": 100 * self.current_usage[component] / allocated
            }
            for component in self.allocations
        }
```

---

## Optimization Targets

- System context: 95-99% utilized (fixed)
- Working memory: 70-80% utilized
- Knowledge base: 60-75% utilized
- Agent context: 70-85% utilized
- Interaction buffer: 60-75% utilized

Total target: 75-85% overall utilization
