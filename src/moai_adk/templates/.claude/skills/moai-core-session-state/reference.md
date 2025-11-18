# Session State Management - Enterprise Reference ( .0)

**Last Updated**: 2025-11-12 | Version: 4.0.0 Enterprise | Status: Production Ready

---

## Table of Contents

1. Session Management Patterns
2. Token Budget Tracking Implementation
3. Recovery Procedures
4. Advanced Coordination Patterns
5. November 2025 API Reference
6. Production Troubleshooting
7. Performance Metrics
8. Complete Code Examples

---

## 1. Session Management Patterns

### Pattern 1: Session Lifecycle Management

```python
from datetime import datetime, timedelta
import json
import uuid

class SessionManager:
    """Manages session lifecycle with token budget tracking"""
    
    def __init__(self, model="claude-sonnet-4-5-20250929"):
        self.model = model
        self.sessions = {}
        self.checkpoints = {}
    
    def create_session(self, task_id, context=None):
        """Create new session with enterprise features"""
        session = {
            'session_id': f'sess_{uuid.uuid4()}',
            'model': self.model,
            'task_id': task_id,
            'created_at': datetime.utcnow().isoformat(),
            'token_budget': {
                'total': 200000,
                'used': 0,
                'available': 200000,
                'position_percent': 0
            },
            'context_cache': context or {},
            'checkpoints': [],
            'status': 'active'
        }
        self.sessions[session['session_id']] = session
        return session
    
    def track_token_usage(self, session_id, tokens_used):
        """Update token usage with enterprise tracking"""
        session = self.sessions[session_id]
        session['token_budget']['used'] += tokens_used
        session['token_budget']['available'] -= tokens_used
        position = (session['token_budget']['used'] / session['token_budget']['total']) * 100
        session['token_budget']['position_percent'] = position
        
        # Check thresholds
        if position > 85:
            self._trigger_emergency_recovery(session_id)
        elif position > 75:
            self._trigger_progressive_summarization(session_id)
        
        return session['token_budget']
    
    def create_checkpoint(self, session_id, phase, context_snapshot):
        """Create recovery checkpoint"""
        checkpoint = {
            'checkpoint_id': f'ckpt_{uuid.uuid4()}',
            'session_id': session_id,
            'timestamp': datetime.utcnow().isoformat(),
            'phase': phase,
            'token_usage': self.sessions[session_id]['token_budget'].copy(),
            'context_snapshot': context_snapshot
        }
        checkpoint_id = checkpoint['checkpoint_id']
        self.checkpoints[checkpoint_id] = checkpoint
        self.sessions[session_id]['checkpoints'].append(checkpoint_id)
        return checkpoint
    
    def resume_session(self, session_id):
        """Resume session with context restoration"""
        session = self.sessions.get(session_id)
        if not session:
            raise ValueError(f"Session {session_id} not found")
        
        # Restore context
        if session['token_budget']['available'] < 30000:
            self._trigger_context_compression(session_id)
        
        session['status'] = 'resumed'
        session['last_resumed'] = datetime.utcnow().isoformat()
        return session
    
    def fork_session(self, session_id):
        """Fork session for parallel exploration"""
        original = self.sessions[session_id]
        fork = self.create_session(f"fork_{original['task_id']}")
        fork['parent_session_id'] = session_id
        fork['fork_checkpoint'] = original['checkpoints'][-1] if original['checkpoints'] else None
        return fork
    
    def _trigger_progressive_summarization(self, session_id):
        """Compress non-critical context at 75% threshold"""
        session = self.sessions[session_id]
        # Remove temporary context
        if 'temp_context' in session['context_cache']:
            del session['context_cache']['temp_context']
    
    def _trigger_emergency_recovery(self, session_id):
        """Emergency recovery at 85% threshold"""
        session = self.sessions[session_id]
        session['status'] = 'emergency_recovery'
        # Archive to disk
        self._archive_session(session_id)
    
    def _trigger_context_compression(self, session_id):
        """Compress all non-critical context"""
        pass
    
    def _archive_session(self, session_id):
        """Archive session to disk"""
        pass
```

### Pattern 2: Handoff Protocol

```python
def create_handoff_package(from_agent, to_agent, session_id, task_context):
    """Create enterprise handoff package with token validation"""
    
    session = get_session(session_id)
    
    # Validate token budget for handoff
    min_required = 30000
    if session['token_budget']['available'] < min_required:
        trigger_context_compression(session_id)
    
    handoff = {
        'handoff_id': f'hoff_{uuid.uuid4()}',
        'timestamp': datetime.utcnow().isoformat(),
        'from_agent': from_agent,
        'to_agent': to_agent,
        'session_context': {
            'session_id': session_id,
            'model': session['model'],
            'context_position': session['token_budget']['position_percent'],
            'available_tokens': session['token_budget']['available'],
            'status': session['status']
        },
        'task_context': task_context,
        'context_snapshot': compress_context(session['context_cache']),
        'recovery_info': {
            'last_checkpoint': session['checkpoints'][-1] if session['checkpoints'] else None,
            'recovery_tokens_reserved': 55000,
            'fork_available': True
        }
    }
    
    # Validate handoff
    validate_handoff(handoff)
    return handoff

def validate_handoff(handoff_package):
    """Validate handoff completeness"""
    
    required = ['handoff_id', 'from_agent', 'to_agent', 'session_context', 'task_context']
    for field in required:
        assert field in handoff_package, f"Missing: {field}"
    
    # Token budget check
    available = handoff_package['session_context']['available_tokens']
    assert available > 30000, "Insufficient token budget for handoff"
    
    # Agent compatibility check
    assert can_agents_cooperate(
        handoff_package['from_agent'],
        handoff_package['to_agent']
    ), "Agents cannot cooperate"
```

---

## 2. Token Budget Tracking Implementation

### Enterprise Token Budget Callback

```python
class TokenBudgetManager:
    """Enterprise token budget management with callbacks"""
    
    def __init__(self):
        self.callbacks = []
        self.usage_history = []
    
    def register_callback(self, callback, threshold_percent):
        """Register callback at token threshold"""
        self.callbacks.append({
            'callback': callback,
            'threshold': threshold_percent
        })
    
    def update_usage(self, session_id, tokens_used):
        """Update usage and trigger callbacks"""
        session = get_session(session_id)
        session['token_budget']['used'] += tokens_used
        available = session['token_budget']['total'] - session['token_budget']['used']
        percent = (session['token_budget']['used'] / session['token_budget']['total']) * 100
        
        # Log usage
        self.usage_history.append({
            'session_id': session_id,
            'timestamp': datetime.utcnow().isoformat(),
            'used': session['token_budget']['used'],
            'available': available,
            'percent': percent
        })
        
        # Trigger callbacks
        for cb_info in self.callbacks:
            if percent >= cb_info['threshold']:
                cb_info['callback'](session_id, percent, available)
    
    def get_usage_report(self, session_id, num_entries=10):
        """Get token usage history"""
        return [
            h for h in self.usage_history 
            if h['session_id'] == session_id
        ][-num_entries:]

# Register callbacks
manager = TokenBudgetManager()

def on_75_percent(session_id, percent, available):
    print(f"Token warning: {percent:.1f}% used ({available} remaining)")
    defer_non_critical_context(session_id)

def on_85_percent(session_id, percent, available):
    print(f"Critical: {percent:.1f}% used - triggering emergency recovery")
    trigger_emergency_recovery(session_id)

manager.register_callback(on_75_percent, 75)
manager.register_callback(on_85_percent, 85)
```

### MCP Server Context Optimization

```python
def analyze_mcp_consumption():
    """Analyze MCP server token consumption"""
    
    # Typical MCP server tool definition costs
    mcp_costs = {
        'context7': 2000,          # ~2K tokens
        'playwright': 3000,        # ~3K tokens
    }
    
    active_servers = get_active_mcp_servers()
    total_cost = sum(mcp_costs.get(s, 0) for s in active_servers)
    
    print(f"Active MCP Servers: {active_servers}")
    print(f"Total tool definition cost: {total_cost} tokens")
    
    # Recommendation
    if total_cost > 5000:
        print("Warning: MCP servers consuming >5K tokens")
        print("Consider disabling unused servers before critical tasks")

def optimize_mcp_for_task(task_type):
    """Optimize MCP servers for specific task"""
    
    recommended_servers = {
        'code_analysis': ['context7'],
        'web_interaction': ['playwright'],
        'general': []
    }
    
    servers = recommended_servers.get(task_type, [])
    
    # Disable unnecessary servers
    current = get_active_mcp_servers()
    for server in current:
        if server not in servers:
            disable_mcp_server(server)
    
    # Enable required servers
    for server in servers:
        if server not in current:
            enable_mcp_server(server)
    
    return get_active_mcp_servers()
```

---

## 3. Recovery Procedures

### Complete Session Recovery

```python
def recover_from_checkpoint(checkpoint_id):
    """Complete recovery from checkpoint"""
    
    # Load checkpoint
    checkpoint = load_checkpoint(checkpoint_id)
    session_id = checkpoint['session_id']
    
    print(f"Recovering from checkpoint: {checkpoint_id}")
    print(f"Phase: {checkpoint['phase']}")
    print(f"Token usage at checkpoint: {checkpoint['token_usage']}")
    
    # Validate token budget for recovery
    available = checkpoint['token_usage']['available']
    if available < 30000:
        print("Warning: Low token budget at recovery point")
        compress_context_aggressive(session_id)
    
    # Restore session state
    session = get_session(session_id)
    session['status'] = 'recovering'
    
    # Restore context from snapshot
    session['context_cache'] = decompress_context(
        checkpoint['context_snapshot']
    )
    
    # Use Claude Agent SDK to resume
    response = resume_session_with_sdk(
        session_id=session_id,
        model=session['model'],
        context=session['context_cache']
    )
    
    session['status'] = 'active'
    session['last_recovery'] = datetime.utcnow().isoformat()
    
    return response

def emergency_recovery(session_id):
    """Emergency recovery when approaching token limit"""
    
    session = get_session(session_id)
    
    # Create final checkpoint
    final_checkpoint = create_checkpoint(
        session_id=session_id,
        phase='emergency_recovery',
        context_snapshot=compress_aggressive(session['context_cache'])
    )
    
    # Archive to disk
    save_session_to_disk(session_id)
    
    # Create new session with fresh token budget
    new_session = create_session(
        task_id=session['task_id'],
        context=final_checkpoint['context_snapshot']
    )
    
    print(f"Emergency recovery completed:")
    print(f"  Original session: {session_id}")
    print(f"  New session: {new_session['session_id']}")
    print(f"  Checkpoint: {final_checkpoint['checkpoint_id']}")
    
    return new_session
```

---

## 4. Advanced Coordination Patterns

### Multi-Session Orchestration

```python
class MultiSessionOrchestrator:
    """Coordinate multiple parallel sessions"""
    
    def __init__(self, max_sessions=3):
        self.max_sessions = max_sessions
        self.active_sessions = {}
        self.session_graph = {}
    
    def create_parallel_sessions(self, specs):
        """Create multiple sessions for parallel SPEC implementation"""
        
        sessions = []
        for spec_id in specs:
            session = create_session(
                task_id=spec_id,
                context={'spec_id': spec_id}
            )
            self.active_sessions[session['session_id']] = session
            sessions.append(session)
        
        # Build dependency graph
        self.session_graph = build_dependency_graph(specs)
        
        return sessions
    
    def fork_session_for_exploration(self, session_id):
        """Fork session for alternative implementation"""
        
        original = self.active_sessions[session_id]
        fork = fork_session(session_id)
        
        # Track fork relationship
        self.session_graph[fork['session_id']] = {
            'parent': session_id,
            'type': 'exploration'
        }
        
        return fork
    
    def merge_sessions(self, fork_session_id, keep_session_id):
        """Merge fork results back to original"""
        
        fork = self.active_sessions[fork_session_id]
        keep = self.active_sessions[keep_session_id]
        
        # Merge context
        keep['context_cache'].update(fork['context_cache'])
        
        # Archive fork
        archive_session(fork_session_id)
        del self.active_sessions[fork_session_id]
        
        return keep
    
    def orchestrate_workflow(self, workflow_spec):
        """Execute complex multi-session workflow"""
        
        # Phase 1: Parallel spec implementations
        sessions = self.create_parallel_sessions(workflow_spec['specs'])
        
        # Phase 2: Monitor token budgets
        for session in sessions:
            monitor_token_budget(session['session_id'])
        
        # Phase 3: Conditional forks for exploration
        for session in sessions:
            if should_explore_alternative(session):
                fork = self.fork_session_for_exploration(session['session_id'])
        
        # Phase 4: Merge results
        # ... merge logic
```

---

## 5. November 2025 API Reference

### Session Management API (Agent SDK)

```python
# Create new session
session = client.sessions.create(
    model="claude-sonnet-4-5-20250929",
    system_prompt="...",
    context_awareness=True  #  feature
)

# Send message and capture session ID
response = client.messages.create(
    model="claude-sonnet-4-5-20250929",
    messages=[{"role": "user", "content": "..."}],
    context_budget_aware=True  # NEW: 
)

# Resume session
response = client.messages.create(
    model="claude-sonnet-4-5-20250929",
    resume=session_id,  # Use saved session ID
    messages=[{"role": "user", "content": "..."}]
)

# Fork session for parallel exploration
fork_response = client.messages.create(
    model="claude-sonnet-4-5-20250929",
    fork_session=session_id,  # Creates new session from checkpoint
    messages=[{"role": "user", "content": "..."}]
)
```

### Token Budget API (Haiku/Sonnet 4.5)

```python
# Get current token budget (NEW in )
context_info = response.usage.context_info  # NEW field

print(f"Tokens used: {context_info.input_tokens}")
print(f"Output tokens: {context_info.output_tokens}")
print(f"Context position: {context_info.context_position_percent}")  # NEW
print(f"Available tokens: {context_info.available_tokens}")  # NEW

# Register callback for budget changes
def on_budget_change(budget_info):
    if budget_info.percent_used > 75:
        print(f"Warning: {budget_info.percent_used}% token budget used")

register_token_budget_callback(on_budget_change)
```

---

## 6. Production Troubleshooting

### Token Budget Exceeded

```
Symptom: Model stops responding mid-conversation
Action:
1. Check context position: /alfred:debug --show-token-budget
2. Identify heavy consumers: /context
3. Disable unnecessary MCP servers
4. Create new session with fresh 200K budget
5. Resume from last checkpoint if available
```

### Session Recovery Failed

```
Symptom: Cannot resume session, "session not found"
Action:
1. Check .moai/sessions/ directory for session files
2. List available checkpoints: /alfred:debug --list-checkpoints
3. Restore from latest checkpoint using recover_from_checkpoint()
4. If no checkpoints, start new session
```

### Memory File Inconsistency

```
Symptom: Different information in session files
Action:
1. Run sync_memory_files() to reconcile
2. Check token-usage.log for consistency
3. Verify context-cache.json integrity
4. Restore from backup if corrupted
```

---

## 7. Performance Metrics

### Key Metrics to Monitor

| Metric | Target | Action at Threshold |
|--------|--------|-------------------|
| Token usage % | <60% | Normal operation |
| Token usage % | 60-75% | Enable progressive disclosure |
| Token usage % | 75-85% | Defer non-critical context |
| Token usage % | >85% | Emergency recovery |
| Session duration | <30 min | Normal |
| Session duration | 30-60 min | Create checkpoint every 10 min |
| Session duration | >60 min | Consider new session |
| Handoff success % | >99% | Monitor for issues |
| Recovery time | <5 min | Acceptable |
| Recovery time | >10 min | Optimize checkpoint strategy |

---

## 8. Complete Code Examples

### End-to-End Session Workflow

```python
from datetime import datetime
import json

def complete_workflow_example():
    """Complete session workflow with recovery"""
    
    # Step 1: Create session
    session = SessionManager().create_session('SPEC-001')
    session_id = session['session_id']
    print(f"Created session: {session_id}")
    
    # Step 2: Execute task with token tracking
    token_manager = TokenBudgetManager()
    token_manager.register_callback(on_75_percent, 75)
    token_manager.register_callback(on_85_percent, 85)
    
    # Simulate token usage
    for i in range(5):
        tokens_used = 15000  # ~15K per iteration
        token_manager.update_usage(session_id, tokens_used)
        
        # Create checkpoint before major operations
        if i % 2 == 0:
            checkpoint = create_checkpoint(
                session_id,
                phase=f'iteration_{i}',
                context_snapshot=get_context_snapshot(session_id)
            )
            print(f"Checkpoint created: {checkpoint['checkpoint_id']}")
    
    # Step 3: Handoff to another agent
    handoff = create_handoff_package(
        from_agent='spec-builder',
        to_agent='tdd-implementer',
        session_id=session_id,
        task_context={'phase': 'implementation'}
    )
    print(f"Handoff package: {handoff['handoff_id']}")
    
    # Step 4: Resume after handoff
    resumed_session = resume_session(session_id)
    print(f"Session resumed: {resumed_session['status']}")
    
    # Step 5: Handle recovery if needed
    usage_report = token_manager.get_usage_report(session_id)
    latest = usage_report[-1]
    
    if latest['percent'] > 80:
        print("Token budget critical - initiating recovery")
        new_session = emergency_recovery(session_id)
        print(f"Recovery completed: {new_session['session_id']}")
    
    return {
        'original_session': session_id,
        'total_tokens_used': latest['used'],
        'status': 'completed'
    }

# Run workflow
result = complete_workflow_example()
print(json.dumps(result, indent=2))
```

---

**Reference Last Updated**: 2025-11-12 | **Version**: 4.0.0 Enterprise | **Models**: Sonnet 4.5, Haiku 4.5

For additional guidance, see:
- SKILL.md - Complete feature overview
- examples.md - Practical usage patterns
- ../moai-core-context-budget - Token optimization deep dive
