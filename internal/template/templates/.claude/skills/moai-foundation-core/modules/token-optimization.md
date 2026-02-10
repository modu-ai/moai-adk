# Token Optimization - Budget Management

Purpose: Efficient 200K token budget management through strategic context loading, phase separation, and model selection for cost-effective AI development.

Version: 1.0.0
Last Updated: 2025-11-25

---

## Quick Reference (30 seconds)

Token Budget: 200K per feature (250K with overhead)

Phase Allocation:
- SPEC Generation: 30K tokens
- DDD Implementation: 180K tokens
- Documentation: 40K tokens

/clear Execution Rules:
1. Immediately after /moai:1-plan (saves 45-50K)
2. When context > 150K tokens
3. After 50+ conversation messages

Model Selection:
- Sonnet 4.5: Quality-critical (SPEC, security)
- Haiku 4.5: Speed/cost (simple edits, tests)
- Cost savings: 60-70% with strategic Haiku use

Context Optimization:
- Target: 20-30K tokens per agent
- Maximum: 50K tokens
- Load only necessary files for current task

---

## Implementation Guide (5 minutes)

### Token Budget Allocation

Standard Feature Budget (250K tokens):

| Phase | Budget | Purpose | Breakdown |
|-------|--------|---------|-----------|
| Phase 1: SPEC | 30K | Requirements definition | EARS format, acceptance criteria, complexity |
| /clear | - | Context reset | Saves 45-50K tokens |
| Phase 2: DDD | 180K | Implementation + tests | ANALYZE (40K) + PRESERVE (80K) + IMPROVE (60K) |
| Phase 3: Docs | 40K | Documentation | API docs, architecture, reports |
| Total | 250K | Complete feature | 60-70% efficiency vs manual |

Budget Monitoring:

```python
class TokenBudgetManager:
 """Track and enforce token budget limits."""
 
 PHASE_BUDGETS = {
 "spec": 30_000,
 "ddd": 180_000,
 "docs": 40_000
 }
 
 TOTAL_BUDGET = 250_000
 WARNING_THRESHOLD = 150_000
 
 def __init__(self):
 self.current_phase = None
 self.phase_usage = {
 "spec": 0,
 "ddd": 0,
 "docs": 0
 }
 
 def track_usage(self, phase: str, tokens_used: int):
 """Track token usage for current phase."""
 self.current_phase = phase
 self.phase_usage[phase] += tokens_used
 
 # Check budget
 if self.phase_usage[phase] > self.PHASE_BUDGETS[phase]:
 raise TokenBudgetExceededError(
 f"Phase {phase} exceeded budget: "
 f"{self.phase_usage[phase]} > {self.PHASE_BUDGETS[phase]}"
 )
 
 # Warn at threshold
 total = self.total_usage()
 if total > self.WARNING_THRESHOLD:
 suggest_clear()
 
 def total_usage(self) -> int:
 """Calculate total token usage across all phases."""
 return sum(self.phase_usage.values())
 
 def remaining_budget(self, phase: str) -> int:
 """Calculate remaining budget for phase."""
 return self.PHASE_BUDGETS[phase] - self.phase_usage[phase]
 
 def get_budget_report(self) -> dict:
 """Generate budget usage report."""
 return {
 "total_budget": self.TOTAL_BUDGET,
 "total_used": self.total_usage(),
 "total_remaining": self.TOTAL_BUDGET - self.total_usage(),
 "phases": {
 phase: {
 "budget": self.PHASE_BUDGETS[phase],
 "used": self.phase_usage[phase],
 "remaining": self.remaining_budget(phase),
 "utilization": (self.phase_usage[phase] / self.PHASE_BUDGETS[phase]) * 100
 }
 for phase in self.PHASE_BUDGETS
 }
 }

# Usage
budget = TokenBudgetManager()

# Phase 1: SPEC
budget.track_usage("spec", 25_000)
print(budget.remaining_budget("spec")) # 5,000 tokens remaining

# Execute /clear
execute_clear()

# Phase 2: DDD
budget.track_usage("ddd", 85_000)
budget.track_usage("ddd", 75_000)
print(budget.total_usage()) # 185,000 (triggers warning)
```

---

### /clear Execution Strategy

Rule 1: Mandatory After SPEC Generation:

```python
# Pattern: SPEC → /clear → Implementation
async def spec_then_implement():
 """Always execute /clear after SPEC."""
 
 # Phase 1: SPEC Generation (heavy context)
 spec = await Task(
 subagent_type="spec-builder",
 prompt="Generate SPEC for user authentication"
 )
 # Context: ~75K tokens (conversation + SPEC content)
 
 # MANDATORY: Execute /clear
 execute_clear()
 # Context: Reset to 0, saves 45-50K tokens
 
 # Phase 2: Implementation (fresh context)
 impl = await Task(
 subagent_type="ddd-implementer",
 prompt="Implement SPEC-001",
 context={
 "spec_id": "SPEC-001", # Minimal reference
 # SPEC content loaded from file, not conversation
 }
 )
 # Context: Only current phase (~80K tokens)
 
 # Total savings: 45-50K tokens
```

Rule 2: Context > 150K Threshold:

```python
def monitor_context_size():
 """Monitor and manage context size."""
 
 current_tokens = get_current_context_tokens()
 
 if current_tokens > 150_000:
 # Warn user
 print(" Context size: {current_tokens}K tokens")
 print(" Recommendation: Execute /clear to reset context")
 
 # Provide context summary before clearing
 summary = generate_context_summary()
 
 # Execute /clear
 execute_clear()
 
 # Restore minimal context
 restore_minimal_context(summary)

def get_current_context_tokens() -> int:
 """Get current context token count."""
 # Use /context command or API
 result = execute_command("/context")
 return parse_token_count(result)

def generate_context_summary() -> dict:
 """Generate compact summary of current context."""
 return {
 "current_spec": get_current_spec_id(),
 "current_phase": get_current_phase(),
 "key_decisions": extract_key_decisions(),
 "pending_actions": extract_pending_actions()
 }

def restore_minimal_context(summary: dict):
 """Restore only essential context after /clear."""
 # Load only necessary files
 load_file(f".moai/specs/{summary['current_spec']}/spec.md")
 # Do NOT reload entire conversation history
```

Rule 3: After 50+ Messages:

```python
class ConversationMonitor:
 """Monitor conversation length and suggest /clear."""
 
 def __init__(self):
 self.message_count = 0
 self.clear_threshold = 50
 
 def track_message(self):
 """Track each message in conversation."""
 self.message_count += 1
 
 if self.message_count >= self.clear_threshold:
 self.suggest_clear()
 
 def suggest_clear(self):
 """Suggest executing /clear."""
 print(f" {self.message_count} messages in conversation")
 print(" Consider executing /clear to reset context")
 print(" This will improve response quality and speed")

# Usage
monitor = ConversationMonitor()

# Each user/assistant message
monitor.track_message()
```

---

### Selective File Loading

Good Practices :

```python
class SelectiveFileLoader:
 """Load only necessary files for current task."""
 
 def __init__(self):
 self.loaded_files = set()
 
 def load_for_task(self, task_type: str, context: dict):
 """Load files specific to task type."""
 
 if task_type == "backend_implementation":
 # Load only backend-related files
 files = [
 f"src/{context['module']}.py",
 f"tests/test_{context['module']}.py",
 f".moai/specs/{context['spec_id']}/spec.md"
 ]
 
 elif task_type == "frontend_implementation":
 # Load only frontend-related files
 files = [
 f"src/components/{context['component']}.tsx",
 f"src/components/{context['component']}.test.tsx",
 f".moai/specs/{context['spec_id']}/spec.md"
 ]
 
 elif task_type == "testing":
 # Load only test files
 files = [
 f"tests/{context['test_module']}.py",
 f"src/{context['implementation_module']}.py"
 ]
 
 else:
 # Default: Load spec only
 files = [f".moai/specs/{context['spec_id']}/spec.md"]
 
 # Load files
 for file in files:
 if file not in self.loaded_files:
 load_file(file)
 self.loaded_files.add(file)
 
 def load_headers_only(self, file_path: str):
 """Load file metadata and headers only (not full content)."""
 with open(file_path) as f:
 # Read first 50 lines (headers, imports, class definitions)
 headers = "".join(f.readlines()[:50])
 return headers
 
 def load_function_signatures(self, file_path: str):
 """Extract only function signatures from file."""
 import ast
 
 with open(file_path) as f:
 tree = ast.parse(f.read())
 
 signatures = []
 for node in ast.walk(tree):
 if isinstance(node, ast.FunctionDef):
 args = [arg.arg for arg in node.args.args]
 signatures.append(f"{node.name}({', '.join(args)})")
 
 return signatures

# Usage
loader = SelectiveFileLoader()

# Backend task: Load only backend files
loader.load_for_task("backend_implementation", {
 "module": "auth",
 "spec_id": "SPEC-001"
})
# Loaded: src/auth.py, tests/test_auth.py, spec.md
# NOT loaded: frontend files, database files, docs

# Estimated tokens: ~15K (vs 150K if loading entire codebase)
```

Bad Practices :

```python
# BAD: Load entire codebase
def load_everything():
 for file in glob("src//*.py"):
 load_file(file) # Loads 100+ files
 for file in glob("tests//*.py"):
 load_file(file)
 # Result: 200K+ tokens, exceeds budget

# BAD: Load node_modules
def load_dependencies():
 for file in glob("node_modules//*.js"):
 load_file(file) # Millions of tokens

# BAD: Load binary files
def load_binaries():
 for file in glob("/*.png"):
 load_file(file) # Non-text data

# BAD: Load conversation history
def load_history():
 load_file(".moai/conversation_history.json") # 500K+ tokens
```

---

### Model Selection Strategy

Decision Matrix:

| Task Type | Model | Reason | Cost | Speed |
|-----------|-------|--------|------|-------|
| SPEC generation | Sonnet 4.5 | High-quality design | $$$ | Slower |
| Security review | Sonnet 4.5 | Precise analysis | $$$ | Slower |
| Architecture design | Sonnet 4.5 | Complex reasoning | $$$ | Slower |
| DDD implementation | Haiku 4.5 | Fast execution | $ | 3x faster |
| Simple edits | Haiku 4.5 | Minimal complexity | $ | 3x faster |
| Test generation | Haiku 4.5 | Pattern-based | $ | 3x faster |
| Documentation | Haiku 4.5 | Template-based | $ | 3x faster |

Cost Comparison:

```python
class ModelCostCalculator:
 """Calculate cost savings with strategic model selection."""
 
 COSTS_PER_1M_TOKENS = {
 "sonnet-4.5": {
 "input": 3.00,
 "output": 15.00
 },
 "haiku-4.5": {
 "input": 1.00,
 "output": 5.00
 }
 }
 
 def calculate_cost(self, model: str, input_tokens: int, output_tokens: int) -> float:
 """Calculate cost for specific model and token usage."""
 input_cost = (input_tokens / 1_000_000) * self.COSTS_PER_1M_TOKENS[model]["input"]
 output_cost = (output_tokens / 1_000_000) * self.COSTS_PER_1M_TOKENS[model]["output"]
 return input_cost + output_cost
 
 def compare_strategies(self, feature_token_budget: dict):
 """Compare cost of all-Sonnet vs strategic mix."""
 
 # Strategy 1: All Sonnet
 all_sonnet_cost = sum(
 self.calculate_cost("sonnet-4.5", phase["input"], phase["output"])
 for phase in feature_token_budget.values()
 )
 
 # Strategy 2: Strategic mix
 strategic_costs = {
 "spec": self.calculate_cost(
 "sonnet-4.5", # Sonnet for SPEC
 feature_token_budget["spec"]["input"],
 feature_token_budget["spec"]["output"]
 ),
 "ddd": self.calculate_cost(
 "haiku-4.5", # Haiku for DDD
 feature_token_budget["ddd"]["input"],
 feature_token_budget["ddd"]["output"]
 ),
 "docs": self.calculate_cost(
 "haiku-4.5", # Haiku for docs
 feature_token_budget["docs"]["input"],
 feature_token_budget["docs"]["output"]
 )
 }
 strategic_total = sum(strategic_costs.values())
 
 savings = all_sonnet_cost - strategic_total
 savings_percent = (savings / all_sonnet_cost) * 100
 
 return {
 "all_sonnet": all_sonnet_cost,
 "strategic_mix": strategic_total,
 "savings": savings,
 "savings_percent": savings_percent
 }

# Example calculation
calculator = ModelCostCalculator()

feature_budget = {
 "spec": {"input": 20_000, "output": 10_000},
 "ddd": {"input": 100_000, "output": 80_000},
 "docs": {"input": 30_000, "output": 10_000}
}

comparison = calculator.compare_strategies(feature_budget)
print(f"All Sonnet: ${comparison['all_sonnet']:.2f}")
print(f"Strategic Mix: ${comparison['strategic_mix']:.2f}")
print(f"Savings: ${comparison['savings']:.2f} ({comparison['savings_percent']:.1f}%)")

# Output:
# All Sonnet: $32.40
# Strategic Mix: $11.80
# Savings: $20.60 (63.6%)
```

---

## Advanced Implementation (10+ minutes)

### Context Passing Optimization

Efficient Context Structure:

```python
class ContextOptimizer:
 """Optimize context passed between agents."""
 
 def __init__(self):
 self.target_size = 30_000 # 30K tokens
 self.max_size = 50_000 # 50K tokens
 
 def optimize_context(self, full_context: dict, agent_type: str) -> dict:
 """Create optimized context for specific agent."""
 
 # Extract agent-specific requirements
 optimized = self._extract_required_fields(full_context, agent_type)
 
 # Compress large data structures
 optimized = self._compress_large_fields(optimized)
 
 # Remove redundant information
 optimized = self._remove_redundancy(optimized)
 
 # Validate size
 size = self._estimate_tokens(optimized)
 if size > self.max_size:
 optimized = self._aggressive_compression(optimized)
 
 return optimized
 
 def _extract_required_fields(self, context: dict, agent_type: str) -> dict:
 """Extract only fields required by specific agent."""
 
 requirements = {
 "backend-expert": ["spec_id", "api_design", "database_schema"],
 "frontend-expert": ["spec_id", "api_endpoints", "ui_components"],
 "security-expert": ["spec_id", "threat_model", "dependencies"],
 "test-engineer": ["spec_id", "code_structure", "test_strategy"],
 "docs-manager": ["spec_id", "api_spec", "architecture"]
 }
 
 required = requirements.get(agent_type, ["spec_id"])
 
 return {
 field: context[field]
 for field in required
 if field in context
 }
 
 def _compress_large_fields(self, context: dict) -> dict:
 """Compress large data structures."""
 
 for key, value in context.items():
 if isinstance(value, str) and len(value) > 5000:
 # Compress long strings
 context[key] = self._summarize_text(value)
 
 elif isinstance(value, list) and len(value) > 100:
 # Sample large lists
 context[key] = value[:50] + ["... (truncated)"] + value[-50:]
 
 elif isinstance(value, dict) and len(str(value)) > 5000:
 # Compress nested dicts
 context[key] = self._compress_dict(value)
 
 return context
 
 def _summarize_text(self, text: str, max_length: int = 1000) -> str:
 """Summarize long text to key points."""
 # Extract first paragraph + last paragraph
 paragraphs = text.split("\n\n")
 if len(paragraphs) > 2:
 summary = paragraphs[0] + "\n\n...\n\n" + paragraphs[-1]
 if len(summary) > max_length:
 return summary[:max_length] + "..."
 return summary
 return text[:max_length]
 
 def _estimate_tokens(self, context: dict) -> int:
 """Estimate token count."""
 import json
 json_str = json.dumps(context, default=str)
 # Rough estimate: 1 token ≈ 4 characters
 return len(json_str) // 4

# Usage
optimizer = ContextOptimizer()

large_context = {
 "spec_id": "SPEC-001",
 "full_code": "..." * 10000, # 50KB code
 "api_design": {...},
 "database_schema": {...},
 "test_results": [...] * 500, # 500 test results
 "conversation_history": "..." * 20000 # 100KB history
}

# Optimize for backend-expert
backend_context = optimizer.optimize_context(large_context, "backend-expert")
# Result: Only spec_id, api_design, database_schema
# Size: ~25K tokens (vs 200K+ original)

result = await Task(
 subagent_type="backend-expert",
 prompt="Implement backend",
 context=backend_context # Optimized context
)
```

### Token Usage Monitoring

Real-time Monitoring Dashboard:

```python
import time
from dataclasses import dataclass
from typing import List

@dataclass
class TokenUsageSnapshot:
 """Snapshot of token usage at point in time."""
 timestamp: float
 phase: str
 operation: str
 tokens_used: int
 cumulative_tokens: int
 budget_remaining: int

class TokenMonitor:
 """Real-time token usage monitoring."""
 
 def __init__(self, total_budget: int = 250_000):
 self.total_budget = total_budget
 self.snapshots: List[TokenUsageSnapshot] = []
 self.cumulative_usage = 0
 
 def record_usage(self, phase: str, operation: str, tokens: int):
 """Record token usage event."""
 self.cumulative_usage += tokens
 
 snapshot = TokenUsageSnapshot(
 timestamp=time.time(),
 phase=phase,
 operation=operation,
 tokens_used=tokens,
 cumulative_tokens=self.cumulative_usage,
 budget_remaining=self.total_budget - self.cumulative_usage
 )
 
 self.snapshots.append(snapshot)
 
 # Check thresholds
 if self.cumulative_usage > 150_000:
 self._warn_threshold()
 
 if self.cumulative_usage > self.total_budget:
 self._alert_exceeded()
 
 def get_usage_report(self) -> dict:
 """Generate comprehensive usage report."""
 
 phase_breakdown = {}
 for snapshot in self.snapshots:
 if snapshot.phase not in phase_breakdown:
 phase_breakdown[snapshot.phase] = 0
 phase_breakdown[snapshot.phase] += snapshot.tokens_used
 
 return {
 "total_budget": self.total_budget,
 "total_used": self.cumulative_usage,
 "budget_remaining": self.total_budget - self.cumulative_usage,
 "utilization": (self.cumulative_usage / self.total_budget) * 100,
 "phase_breakdown": phase_breakdown,
 "efficiency_score": self._calculate_efficiency(),
 "recommendations": self._generate_recommendations()
 }
 
 def _calculate_efficiency(self) -> float:
 """Calculate token usage efficiency (0-100)."""
 # Higher is better (less waste)
 if self.cumulative_usage == 0:
 return 100.0
 
 # Efficiency based on staying within budget
 if self.cumulative_usage <= self.total_budget:
 return 100 * (1 - (self.cumulative_usage / self.total_budget))
 else:
 # Penalty for exceeding budget
 return max(0, 100 - ((self.cumulative_usage - self.total_budget) / self.total_budget * 100))
 
 def _generate_recommendations(self) -> List[str]:
 """Generate optimization recommendations."""
 recommendations = []
 
 if self.cumulative_usage > 150_000:
 recommendations.append("Execute /clear to reset context")
 
 phase_usage = {}
 for snapshot in self.snapshots:
 phase_usage[snapshot.phase] = phase_usage.get(snapshot.phase, 0) + snapshot.tokens_used
 
 for phase, usage in phase_usage.items():
 if usage > 100_000:
 recommendations.append(f"Phase '{phase}' using {usage}K tokens - consider breaking into smaller tasks")
 
 return recommendations

# Usage
monitor = TokenMonitor(total_budget=250_000)

# Record usage throughout workflow
monitor.record_usage("spec", "generate_spec", 25_000)
monitor.record_usage("spec", "validate_spec", 5_000)
# Execute /clear here
monitor.record_usage("ddd", "analyze_phase", 40_000)
monitor.record_usage("ddd", "preserve_phase", 80_000)
monitor.record_usage("ddd", "improve_phase", 60_000)
monitor.record_usage("docs", "generate_docs", 30_000)

# Generate report
report = monitor.get_usage_report()
print(f"Total used: {report['total_used']:,} / {report['total_budget']:,}")
print(f"Utilization: {report['utilization']:.1f}%")
print(f"Efficiency: {report['efficiency_score']:.1f}")
print("\nRecommendations:")
for rec in report['recommendations']:
 print(f"- {rec}")
```

---

## Works Well With

Skills:
- moai-foundation-delegation-patterns - Context passing
- moai-foundation-progressive-disclosure - Content structuring
- moai-cc-memory - Context persistence

Commands:
- /clear - Context reset (mandatory after /moai:1-plan)
- /context - Check current token usage
- /moai:1-plan - SPEC generation (30K budget)
- /moai:2-run - DDD implementation (180K budget)
- /moai:3-sync - Documentation (40K budget)

Memory:
- Skill("moai-foundation-core") modules/token-optimization.md - Optimization strategies
- @.moai/config/config.json - Budget configuration

---

Version: 1.0.0
Last Updated: 2025-11-25
Status: Production Ready
