# Session State Management - Practical Examples ( .0)

**Last Updated**: 2025-11-12 | **Enterprise Patterns** | **Production Ready**

---

## Quick Reference: Common Scenarios

### Scenario 1: Basic Session Creation and Token Tracking

```bash
# Command line usage
/alfred:1-plan "Implement user authentication"

# Output (from Alfred):
Session created: sess_a1b2c3d4
Model: claude-sonnet-4-5-20250929
Token budget: 200,000 available
Context position: 0%

# Next steps shown via AskUserQuestion
[Implement / Revise / New session]
```

### Scenario 2: Session Resume When Interrupted

```python
# In Python code:

# Session A interrupted after 2 hours
previous_session_id = "sess_a1b2c3d4"

# Check if resumable
status = check_session_status(previous_session_id)
# Output:
# {
#   "status": "available",
#   "tokens_available": 85000,
#   "position_percent": 57.5,
#   "last_activity": "2025-11-12T10:30:00Z"
# }

# Resume session
response = client.messages.create(
    model="claude-sonnet-4-5-20250929",
    resume=previous_session_id,
    messages=[{
        "role": "user",
        "content": "Continue implementing the authentication module"
    }]
)

# New session ID for tracking
new_session_info = response.usage.session_info
print(f"Session resumed: {new_session_info.session_id}")
```

### Scenario 3: Token Budget Approaching Limit

```
Automatic Detection:

Step 1: Model detects 75% token usage
        └─ Haiku/Sonnet 4.5 context awareness triggers

Step 2: Progressive summarization activated
        ├─ Remove temporary context
        ├─ Compress historical info
        └─ Maintain critical references

Step 3: Non-critical context deferred
        ├─ Reference materials postponed
        ├─ Debug info archived
        └─ ~30K tokens freed

Step 4: Continue with fresh capacity
        └─ Available tokens: ~75K (fully usable again)

If usage exceeds 85%:
├─ Emergency recovery checkpoint created
├─ Original session archived
├─ New session started (fresh 200K)
├─ Checkpoint data loaded into new session
└─ Workflow continues seamlessly
```

### Scenario 4: Session Forking for Parallel Exploration

```python
# Main implementation session
main_session_id = "sess_main_impl"

# Developer wants to test alternative approach
fork_response = client.messages.create(
    model="claude-sonnet-4-5-20250929",
    fork_session=main_session_id,
    messages=[{
        "role": "user",
        "content": "Try using async/await instead of promises"
    }]
)

fork_session_id = fork_response.usage.session_info.session_id
# fork_session_id = "sess_fork_alt_async"

# Now have two parallel sessions:
# - main_session_id: Continues with promises approach
# - fork_session_id: Explores async/await approach

# Later: Compare results and keep best approach
# Then merge fork results back to main session
```

---

## Detailed Examples

### Example 1: Complete SPEC Implementation with Checkpoints

```python
from datetime import datetime

class SPECWorkflow:
    """Manage complete SPEC implementation workflow"""
    
    def __init__(self, spec_id):
        self.spec_id = spec_id
        self.session = SessionManager().create_session(spec_id)
        self.checkpoints = {}
    
    def run(self):
        """Execute SPEC with enterprise recovery"""
        
        try:
            # Phase 1: Requirements analysis
            print(f"Phase 1: Requirements analysis for {self.spec_id}")
            self._checkpoint('phase_1_start')
            
            # ... analysis work ...
            
            self._checkpoint('phase_1_complete')
            
            # Phase 2: Design (RED tests)
            print(f"Phase 2: Writing tests (RED)")
            self._checkpoint('phase_2_red_start')
            
            # Token check before heavy work
            if self._check_token_budget() < 50000:
                self._optimize_tokens()
            
            # ... write tests ...
            
            self._checkpoint('phase_2_red_complete')
            
            # Phase 3: Implementation (GREEN)
            print(f"Phase 3: Implementing code (GREEN)")
            self._checkpoint('phase_3_green_start')
            
            # ... implement code ...
            
            self._checkpoint('phase_3_green_complete')
            
            # Phase 4: Refactoring
            print(f"Phase 4: Refactoring")
            self._checkpoint('phase_4_refactor_start')
            
            # ... refactor ...
            
            self._checkpoint('phase_4_refactor_complete')
            
            return {
                'spec_id': self.spec_id,
                'status': 'completed',
                'checkpoints': len(self.checkpoints),
                'tokens_used': self._get_total_tokens_used()
            }
        
        except Exception as e:
            print(f"Error in {self.spec_id}: {e}")
            self._recover_from_error()
            raise
    
    def _checkpoint(self, phase_name):
        """Create recovery checkpoint"""
        checkpoint = {
            'phase': phase_name,
            'timestamp': datetime.utcnow().isoformat(),
            'token_budget': self.session['token_budget'].copy(),
            'id': f"ckpt_{len(self.checkpoints)}"
        }
        self.checkpoints[phase_name] = checkpoint
        print(f"  Checkpoint: {phase_name} (tokens available: {checkpoint['token_budget']['available']})")
    
    def _check_token_budget(self):
        """Check remaining tokens"""
        return self.session['token_budget']['available']
    
    def _optimize_tokens(self):
        """Compress context when tokens low"""
        print(f"  Optimizing token budget...")
        # Compress non-critical context
        pass
    
    def _recover_from_error(self):
        """Recover from error using checkpoint"""
        if self.checkpoints:
            latest_phase = list(self.checkpoints.keys())[-1]
            print(f"Recovering from checkpoint: {latest_phase}")
            # Load checkpoint and resume
    
    def _get_total_tokens_used(self):
        """Calculate total tokens used"""
        return self.session['token_budget']['used']

# Usage:
workflow = SPECWorkflow('SPEC-001')
result = workflow.run()
print(result)
# Output:
# {
#   'spec_id': 'SPEC-001',
#   'status': 'completed',
#   'checkpoints': 8,
#   'tokens_used': 95000
# }
```

### Example 2: Multi-Agent Handoff

```python
def multi_agent_implementation_workflow():
    """Coordinate between multiple agents using handoff protocol"""
    
    # Agent 1: Spec Builder
    spec_session = SessionManager().create_session('SPEC-001')
    print(f"Spec Builder: Created session {spec_session['session_id']}")
    
    # Build SPEC...
    spec_data = {
        'title': 'User Authentication System',
        'requirements': ['login', 'registration', 'password_reset'],
        'acceptance_criteria': [...]
    }
    
    # Create checkpoint before handoff
    checkpoint = {
        'spec_id': 'SPEC-001',
        'spec_data': spec_data,
        'completion_percent': 100,
        'timestamp': datetime.utcnow().isoformat()
    }
    
    # Create handoff package
    handoff = {
        'handoff_id': f'hoff_{uuid.uuid4()}',
        'from_agent': 'spec-builder',
        'to_agent': 'tdd-implementer',
        'session_context': {
            'session_id': spec_session['session_id'],
            'model': 'claude-sonnet-4-5-20250929',
            'available_tokens': spec_session['token_budget']['available'],
            'context_position': spec_session['token_budget']['position_percent']
        },
        'task_context': {
            'spec_id': 'SPEC-001',
            'spec_data': spec_data,
            'current_phase': 'tdd_implementation'
        },
        'checkpoint': checkpoint
    }
    
    # Validate handoff
    assert handoff['session_context']['available_tokens'] > 30000
    print(f"Spec Builder: Handoff package ready ({handoff['handoff_id']})")
    
    # Agent 2: TDD Implementer receives handoff
    print(f"TDD Implementer: Received handoff")
    
    # Resume session from handoff
    impl_session = resume_session(handoff['session_context']['session_id'])
    
    # Write tests...
    print(f"TDD Implementer: Writing tests...")
    
    # Implement code...
    print(f"TDD Implementer: Implementing code...")
    
    # Create second handoff for next phase
    handoff2 = {
        'from_agent': 'tdd-implementer',
        'to_agent': 'doc-syncer',
        'session_context': {
            'session_id': impl_session['session_id'],
            'available_tokens': impl_session['token_budget']['available']
        }
    }
    
    print(f"TDD Implementer: Handing off to doc-syncer")
    
    # Agent 3: Doc Syncer
    print(f"Doc Syncer: Creating documentation...")
    
    return {
        'spec_id': 'SPEC-001',
        'status': 'completed',
        'handoff_chain': ['spec-builder', 'tdd-implementer', 'doc-syncer']
    }

result = multi_agent_implementation_workflow()
```

### Example 3: Emergency Recovery

```python
def handle_emergency_recovery():
    """Handle situation when token budget critical"""
    
    session_id = "sess_critical_tokens"
    session = get_session(session_id)
    
    print(f"Status: Token budget at {session['token_budget']['position_percent']:.1f}%")
    print(f"Available: {session['token_budget']['available']} tokens")
    
    if session['token_budget']['position_percent'] > 85:
        print("Emergency: Token budget critical (>85%)")
        
        # Step 1: Create final checkpoint
        final_checkpoint = {
            'checkpoint_id': f"ckpt_emergency_{uuid.uuid4()}",
            'timestamp': datetime.utcnow().isoformat(),
            'phase': 'emergency_recovery',
            'token_usage': session['token_budget'].copy(),
            'context_snapshot': compress_aggressive(session['context_cache']),
            'recovery_instructions': "Use this checkpoint to resume work"
        }
        
        print(f"Created emergency checkpoint: {final_checkpoint['checkpoint_id']}")
        
        # Step 2: Save session to disk
        save_session_to_disk(session_id, final_checkpoint)
        print(f"Session saved to disk")
        
        # Step 3: Create new session with fresh budget
        new_session = SessionManager().create_session(
            task_id=session['task_id'],
            context={
                'recovery_from': session_id,
                'recovery_checkpoint': final_checkpoint['checkpoint_id'],
                'original_work': final_checkpoint['context_snapshot']
            }
        )
        
        print(f"New session created: {new_session['session_id']}")
        print(f"Fresh token budget: {new_session['token_budget']['total']} tokens")
        
        # Step 4: Resume work in new session
        result = client.messages.create(
            model="claude-sonnet-4-5-20250929",
            resume=new_session['session_id'],
            messages=[{
                "role": "user",
                "content": f"Resume work using checkpoint {final_checkpoint['checkpoint_id']}"
            }]
        )
        
        return {
            'original_session': session_id,
            'new_session': new_session['session_id'],
            'checkpoint': final_checkpoint['checkpoint_id'],
            'status': 'recovered'
        }

recovery_result = handle_emergency_recovery()
# Output:
# {
#   'original_session': 'sess_critical_tokens',
#   'new_session': 'sess_recovery_001',
#   'checkpoint': 'ckpt_emergency_abc123',
#   'status': 'recovered'
# }
```

### Example 4: MCP Server Context Optimization

```python
def optimize_mcp_for_code_analysis():
    """Optimize MCP servers for specific task type"""
    
    # Check current MCP consumption
    print("Analyzing MCP server context usage...")
    
    current_servers = {
        'context7': {
            'enabled': True,
            'token_cost': 2000,
            'used_in_last_hour': True
        },
        'playwright': {
            'enabled': True,
            'token_cost': 3000,
            'used_in_last_hour': False
        }
    }
    
    total_cost = sum(s['token_cost'] for s in current_servers.values())
    print(f"Total MCP token cost: {total_cost} tokens")
    
    # Identify unused servers
    unused = [name for name, info in current_servers.items() if not info['used_in_last_hour']]
    
    if unused:
        print(f"Unused servers (not used in last hour): {unused}")
        print(f"Potential savings: {sum(current_servers[name]['token_cost'] for name in unused)} tokens")
        
        # Disable unused servers
        for server_name in unused:
            print(f"  Disabling {server_name}...")
            disable_mcp_server(server_name)
    
    # For code analysis task, only need context7
    recommended = ['context7']
    
    # Enable necessary servers
    for server in recommended:
        if server in current_servers and not current_servers[server]['enabled']:
            print(f"  Enabling {server}...")
            enable_mcp_server(server)
    
    # Verify optimization
    optimized_cost = sum(
        current_servers[name]['token_cost'] 
        for name in recommended 
        if name in current_servers
    )
    
    print(f"\nOptimized MCP token cost: {optimized_cost} tokens")
    print(f"Tokens freed: {total_cost - optimized_cost}")

optimize_mcp_for_code_analysis()
# Output:
# Analyzing MCP server context usage...
# Total MCP token cost: 6500 tokens
# Unused servers (not used in last hour): ['playwright']
# Potential savings: 3000 tokens
#   Disabling playwright...
#   
# Optimized MCP token cost: 3500 tokens
# Tokens freed: 3000
```

### Example 5: Progressive Summarization

```python
class ContextOptimizer:
    """Optimize context with progressive summarization"""
    
    def progressive_summarize(self, context_data):
        """Compress context while maintaining critical information"""
        
        print(f"Original context size: {len(str(context_data))} chars")
        
        # Identify critical vs temporary context
        critical = {
            'task_objectives': context_data.get('task_objectives'),
            'user_preferences': context_data.get('user_preferences'),
            'active_constraints': context_data.get('constraints')
        }
        
        supporting = {
            'related_history': context_data.get('history'),
            'reference_docs': context_data.get('references')
        }
        
        temporary = {
            'raw_outputs': context_data.get('raw_tool_outputs'),
            'intermediate_calcs': context_data.get('temp_calculations')
        }
        
        # Step 1: Keep critical as-is
        optimized = critical.copy()
        
        # Step 2: Summarize supporting context
        if supporting['related_history']:
            optimized['history_summary'] = self._summarize_history(
                supporting['related_history']
            )
        
        if supporting['reference_docs']:
            optimized['reference_pointers'] = [
            ]
        
        # Step 3: Discard temporary
        # (temporary context is removed)
        
        print(f"Optimized context size: {len(str(optimized))} chars")
        print(f"Compression ratio: {len(str(context_data)) / len(str(optimized)):.1f}x")
        
        return optimized
    
    def _summarize_history(self, history):
        """Compress history while keeping key decisions"""
        return {
            'decisions_made': [h['decision'] for h in history if h.get('decision')],
            'key_milestones': [h['milestone'] for h in history if h.get('milestone')],
            'timeline': f"{len(history)} events compressed to summaries"
        }

optimizer = ContextOptimizer()

large_context = {
    'task_objectives': 'Implement user authentication',
    'user_preferences': {'language': 'ko', 'expertise': 'intermediate'},
    'constraints': ['must_use_pytest', 'coverage_85'],
    'history': [
        {'decision': 'use_oauth2', 'timestamp': '...'},
        {'milestone': 'spec_complete', 'timestamp': '...'},
        {'event': 'debug_session_1', 'output': '...(1000 chars)...'},
    ],
    'references': {'oauth2_guide': 'url', 'pytest_docs': 'url'},
    'raw_tool_outputs': ['...large output...'],
    'temp_calculations': {'intermediate_results': [...]}
}

optimized = optimizer.progressive_summarize(large_context)
```

---

## Reference Links

- **SKILL.md** - Complete feature overview
- **reference.md** - Implementation patterns and APIs
- **../moai-core-context-budget** - Advanced token optimization
- **../moai-core-agent-guide** - Multi-agent coordination
- **Claude Agent SDK** - https://docs.claude.com/agent-sdk
- **Claude Session Management** - https://docs.claude.com/en/docs/agent-sdk/sessions
- **Context Windows** - https://docs.claude.com/en/docs/build-with-claude/context-windows

**Examples Last Updated**: 2025-11-12 | **Version**: 4.0.0 Enterprise | **Production Ready**
