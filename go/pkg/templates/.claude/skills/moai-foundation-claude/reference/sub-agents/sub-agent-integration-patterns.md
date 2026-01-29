# Claude Code Sub-agents Integration Patterns

Comprehensive guide for sub-agent integration, coordination patterns, and workflow orchestration in Claude Code development environments.

Purpose: Integration patterns and best practices for sub-agent coordination
Target: Sub-agent developers and workflow orchestrators
Last Updated: 2025-11-25
Version: 2.0.0

---

## Quick Reference (30 seconds)

Core Patterns: Sequential, Parallel, Conditional delegation. Coordination: Task() API, context passing, error handling. Best Practices: Single responsibility, clear boundaries, structured workflows. Quality Gates: Validation checkpoints, error recovery, result integration.

---

## Integration Patterns Overview

### 1. Sequential Delegation Pattern

Description: Chain multiple sub-agents where each depends on the output of the previous agent.

Use Cases:
- Multi-phase development workflows
- Quality assurance pipelines
- Documentation generation cycles
- Testing and deployment workflows

Implementation:
```python
# Sequential delegation example
def sequential_workflow(user_request):
 """Execute sequential sub-agent workflow."""

 # Phase 1: Specification
 spec_result = Task(
 subagent_type="workflow-spec",
 prompt=f"Create specification for: {user_request}"
 )

 # Phase 2: Implementation (passes spec as context)
 implementation_result = Task(
 subagent_type="workflow-ddd",
 prompt="Implement from specification",
 context={
 "specification": spec_result,
 "requirements": user_request
 }
 )

 # Phase 3: Quality Validation
 quality_result = Task(
 subagent_type="core-quality",
 prompt="Validate implementation quality",
 context={
 "implementation": implementation_result,
 "specification": spec_result
 }
 )

 # Phase 4: Documentation
 docs_result = Task(
 subagent_type="workflow-docs",
 prompt="Generate documentation",
 context={
 "implementation": implementation_result,
 "specification": spec_result,
 "quality_report": quality_result
 }
 )

 return {
 "specification": spec_result,
 "implementation": implementation_result,
 "quality_report": quality_result,
 "documentation": docs_result
 }

# Usage example
result = sequential_workflow("Create user authentication system")
```

Advantages:
- Clear dependency management
- Structured workflow progression
- Easy error tracking and debugging
- Predictable execution order

Considerations:
- Sequential execution time (may be slower)
- Single point of failure
- Limited parallelization opportunities
- Context passing overhead

### 2. Parallel Delegation Pattern

Description: Execute multiple sub-agents simultaneously when tasks are independent.

Use Cases:
- Multi-component development
- Parallel analysis tasks
- Comprehensive testing scenarios
- Independent quality checks

Implementation:
```python
# Parallel delegation example
def parallel_workflow(project_requirements):
 """Execute parallel sub-agent workflow.

 Note: In Claude Code, calling multiple Task() in a single response
 will automatically execute them in parallel (up to 10 concurrent).
 No need for asyncio.gather or Promise.all.
 """

 # Call multiple Task() for automatic parallel execution
 # Claude Code executes up to 10 Tasks concurrently
 frontend_design = Task(
 subagent_type="code-frontend",
 prompt="Design frontend architecture",
 context={"requirements": project_requirements}
 )

 backend_design = Task(
 subagent_type="code-backend",
 prompt="Design backend architecture",
 context={"requirements": project_requirements}
 )

 database_design = Task(
 subagent_type="data-database",
 prompt="Design database schema",
 context={"requirements": project_requirements}
 )

 security_analysis = Task(
 subagent_type="security-expert",
 prompt="Security threat modeling",
 context={"requirements": project_requirements}
 )

 # All 4 Tasks above execute in parallel automatically
 # Results are available when all complete

 # Integration phase (runs after parallel tasks complete)
 integration_result = Task(
 subagent_type="integration-specialist",
 prompt="Integrate component designs",
 context={
 "frontend_design": frontend_design,
 "backend_design": backend_design,
 "database_design": database_design,
 "security_analysis": security_analysis,
 "requirements": project_requirements
 }
 )

 return {
 "frontend": frontend_design,
 "backend": backend_design,
 "database": database_design,
 "security": security_analysis,
 "integration": integration_result
 }

# Usage example
result = await parallel_workflow("E-commerce platform requirements")
```

Advantages:
- Faster execution for independent tasks
- Efficient resource utilization
- Natural parallelism in development workflows
- Better scalability for complex projects

Considerations:
- Complex integration requirements
- Synchronization challenges
- Error handling across multiple agents
- Resource contention issues

### 3. Conditional Delegation Pattern

Description: Route to different sub-agents based on analysis and classification.

Use Cases:
- Error classification and resolution
- Problem type identification
- Specialized task routing
- Dynamic workflow adaptation

Implementation:
```python
# Conditional delegation example
class ConditionalWorkflow:
 def __init__(self):
 self.analysis_agents = {
 "code_analysis": "code-analyst",
 "security_analysis": "security-expert",
 "performance_analysis": "performance-engineer"
 }

 self.resolution_agents = {
 "syntax_error": "format-expert",
 "logic_error": "support-debug",
 "security_vulnerability": "security-expert",
 "performance_issue": "performance-engineer",
 "integration_error": "integration-specialist"
 }

 def analyze_and_resolve(self, error_context):
 """Analyze error and route to appropriate resolution agent."""

 # Phase 1: Analysis
 analysis_result = Task(
 subagent_type="error-analyst",
 prompt="Analyze error and classify problem type",
 context={"error": error_context}
 )

 # Phase 2: Conditional routing
 problem_type = analysis_result.classification
 resolution_agent = self.get_resolution_agent(problem_type)

 # Phase 3: Resolution
 resolution_result = Task(
 subagent_type=resolution_agent,
 prompt=f"Resolve {problem_type} issue",
 context={
 "error": error_context,
 "analysis": analysis_result
 }
 )

 return {
 "analysis": analysis_result,
 "resolution": resolution_result,
 "routing": {
 "problem_type": problem_type,
 "selected_agent": resolution_agent
 }
 }

 def get_resolution_agent(self, problem_type):
 """Select appropriate resolution agent based on problem type."""
 return self.resolution_agents.get(
 problem_type,
 "support-debug" # Default fallback
 )

# Usage example
workflow = ConditionalWorkflow()
result = workflow.analyze_and_resolve({"error": "Null pointer exception in user service"})
```

Advantages:
- Intelligent task routing
- Specialized problem solving
- Efficient resource allocation
- Adaptive workflow behavior

Considerations:
- Complex classification logic
- Error handling in routing
- Agent selection criteria
- Fallback mechanisms

### 4. Orchestrator Pattern

Description: Master agent coordinates multiple sub-agents in complex workflows.

Use Cases:
- Complex project initialization
- Multi-phase development processes
- Comprehensive quality assurance
- End-to-end system deployment

Implementation:
```yaml
---
name: development-orchestrator
description: Orchestrate complete software development workflow from specification to deployment. Use PROACTIVELY for complex multi-component projects requiring coordination across multiple phases and teams.
tools: Read, Write, Edit, Task
model: sonnet
skills: moai-core-workflow, moai-project-manager, moai-foundation-quality
---

# Development Orchestrator

You are a development workflow orchestrator responsible for coordinating multiple sub-agents throughout the complete software development lifecycle.

## Core Responsibilities

Primary Domain: Workflow orchestration and coordination
Key Capabilities: Multi-agent coordination, workflow management, quality assurance, deployment automation
Focus Areas: End-to-end process automation, team coordination, quality assurance

## Orchestration Workflow

### Phase 1: Project Setup
1. Initialize project structure and configuration
2. Set up development environment and tools
3. Establish team workflows and processes
4. Configure quality gates and validation

### Phase 2: Development Coordination
1. Coordinate specification creation with workflow-spec
2. Manage implementation with workflow-ddd
3. Oversee quality validation with core-quality
4. Handle documentation generation with workflow-docs

### Phase 3: Integration and Deployment
1. Coordinate component integration
2. Manage deployment processes with devops-expert
3. Handle testing and validation
4. Monitor production deployment

## Agent Coordination Patterns

### Sequential Dependencies
- Wait for phase completion before proceeding
- Pass results between phases as context
- Handle phase-specific error recovery

### Parallel Execution
- Identify independent tasks for parallel processing
- Coordinate multiple agents simultaneously
- Integrate parallel results for final output

### Quality Assurance
- Validate outputs at each phase
- Implement rollback mechanisms for failures
- Ensure compliance with quality standards
```

Orchestration Implementation:
```python
# Advanced orchestrator implementation
class DevelopmentOrchestrator:
 def __init__(self):
 self.workflow_phases = {
 'specification': {
 'agent': 'workflow-spec',
 'inputs': ['requirements', 'stakeholders'],
 'outputs': ['specification', 'acceptance_criteria'],
 'dependencies': []
 },
 'implementation': {
 'agent': 'workflow-ddd',
 'inputs': ['specification'],
 'outputs': ['code', 'tests'],
 'dependencies': ['specification']
 },
 'validation': {
 'agent': 'core-quality',
 'inputs': ['code', 'tests', 'specification'],
 'outputs': ['quality_report'],
 'dependencies': ['implementation']
 },
 'documentation': {
 'agent': 'workflow-docs',
 'inputs': ['code', 'specification', 'quality_report'],
 'outputs': ['documentation'],
 'dependencies': ['validation']
 }
 }

 self.current_phase = None
 self.phase_results = {}
 self.error_handlers = {}

 def execute_workflow(self, project_request):
 """Execute complete development workflow."""
 try:
 for phase_name, phase_config in self.workflow_phases.items():
 self.current_phase = phase_name

 # Check dependencies
 if not self.validate_dependencies(phase_config['dependencies']):
 raise DependencyError(f"Missing dependencies for {phase_name}")

 # Execute phase
 phase_result = self.execute_phase(phase_name, phase_config, project_request)
 self.phase_results[phase_name] = phase_result

 print(f" Phase {phase_name} completed")

 return self.generate_final_report()

 except Exception as error:
 return self.handle_workflow_error(error)

 def execute_phase(self, phase_name, phase_config, context):
 """Execute a single workflow phase."""
 agent = phase_config['agent']

 # Prepare phase context
 phase_context = {
 'phase_name': phase_name,
 'phase_inputs': self.get_phase_inputs(phase_config['inputs']),
 'workflow_results': self.phase_results,
 'project_context': context
 }

 # Execute agent
 result = Task(
 subagent_type=agent,
 prompt=f"Execute {phase_name} phase",
 context=phase_context
 )

 return result

 def validate_dependencies(self, required_dependencies):
 """Validate that all required dependencies are satisfied."""
 for dependency in required_dependencies:
 if dependency not in self.phase_results:
 return False
 return True
```

## Error Handling and Recovery

### Error Classification

Error Types and Handling Strategies:
```python
# Error handling strategies
class ErrorHandler:
 def __init__(self):
 self.error_strategies = {
 'agent_failure': {
 'strategy': 'retry_with_alternative',
 'max_retries': 3,
 'fallback_agents': {
 'workflow-spec': 'requirements-analyst',
 'workflow-ddd': 'code-developer',
 'core-quality': 'manual-review'
 }
 },
 'dependency_failure': {
 'strategy': 'resolve_dependency',
 'resolution_methods': ['skip_phase', 'manual_intervention', 'alternative_workflow']
 },
 'quality_failure': {
 'strategy': 'fix_and_retry',
 'auto_fix': True,
 'manual_review_required': True
 },
 'timeout_failure': {
 'strategy': 'increase_timeout_or_simplify',
 'timeout_multiplier': 2.0,
 'simplification_level': 'medium'
 }
 }

 def handle_error(self, error, phase_name, context):
 """Handle workflow error with appropriate strategy."""
 error_type = self.classify_error(error)
 strategy = self.error_strategies.get(error_type, {
 'strategy': 'escalate_to_human',
 'escalation_level': 'high'
 })

 if strategy['strategy'] == 'retry_with_alternative':
 return self.retry_with_alternative(error, phase_name, strategy)
 elif strategy['strategy'] == 'resolve_dependency':
 return self.resolve_dependency(error, phase_name, strategy)
 elif strategy['strategy'] == 'fix_and_retry':
 return self.fix_and_retry(error, phase_name, strategy)
 else:
 return self.escalate_to_human(error, phase_name, context)
```

### Recovery Mechanisms

Recovery Patterns:
```python
# Workflow recovery mechanisms
class RecoveryManager:
 def __init__(self):
 self.checkpoints = {}
 self.rollback_state = None
 self.recovery_strategies = {}

 def create_checkpoint(self, phase_name, state):
 """Create workflow checkpoint for recovery."""
 self.checkpoints[phase_name] = {
 'state': state.copy(),
 'timestamp': datetime.now(),
 'dependencies_met': self.validate_current_dependencies()
 }

 def rollback_to_checkpoint(self, target_phase):
 """Rollback workflow to specified checkpoint."""
 if target_phase not in self.checkpoints:
 raise ValueError(f"No checkpoint found for phase: {target_phase}")

 checkpoint = self.checkpoints[target_phase]
 self.rollback_state = checkpoint['state']

 # Reset current phase to checkpoint state
 self.restore_from_checkpoint(checkpoint)

 return {
 'rollback_successful': True,
 'target_phase': target_phase,
 'restored_state': checkpoint['state']
 }

 def restore_from_checkpoint(self, checkpoint):
 """Restore workflow state from checkpoint."""
 # Clear results from phases after checkpoint
 phases_to_clear = [
 phase for phase in self.workflow_phases.keys()
 if self.get_phase_order(phase) > self.get_phase_order(checkpoint['phase'])
 ]

 for phase in phases_to_clear:
 self.phase_results.pop(phase, None)

 # Restore checkpoint state
 self.phase_results.update(checkpoint['state'])
```

## Context Management

### Context Passing Strategies

Optimal Context Patterns:
```python
# Context optimization for agent delegation
class ContextManager:
 def __init__(self):
 self.context_cache = {}
 self.compression_enabled = True
 self.max_context_size = 10000 # characters

 def optimize_context(self, context_data):
 """Optimize context for efficient agent communication."""
 if not context_data:
 return {}

 # Apply context compression for large data
 if self.compression_enabled and len(str(context_data)) > self.max_context_size:
 return self.compress_context(context_data)

 # Filter relevant information
 return self.filter_relevant_context(context_data)

 def filter_relevant_context(self, context):
 """Filter context to include only relevant information."""
 filtered_context = {}

 # Keep essential workflow information
 if 'workflow_results' in context:
 filtered_context['workflow_results'] = {
 phase: self.summarize_results(results)
 for phase, results in context['workflow_results'].items()
 }

 # Keep current phase information
 if 'current_phase' in context:
 filtered_context['current_phase'] = context['current_phase']

 # Keep critical project data
 critical_keys = ['project_id', 'requirements', 'constraints']
 for key in critical_keys:
 if key in context:
 filtered_context[key] = context[key]

 return filtered_context

 def compress_context(self, context_data):
 """Compress large context data."""
 # Implement context compression logic
 return {
 'compressed': True,
 'summary': self.create_context_summary(context_data),
 'key_data': self.extract_key_data(context_data)
 }
```

### Context Validation

Context Quality Assurance:
```python
# Context validation and sanitization
class ContextValidator:
 def __init__(self):
 self.validation_rules = {
 'required_fields': ['project_id', 'phase_name'],
 'max_size': 50000, # characters
 'allowed_types': [str, int, float, bool, dict, list],
 'sanitization_rules': [
 'remove_sensitive_data',
 'validate_structure',
 'check_for_malicious_content'
 ]
 }

 def validate_context(self, context):
 """Validate context data for agent delegation."""
 validation_result = {
 'valid': True,
 'errors': [],
 'warnings': []
 }

 # Check required fields
 for field in self.validation_rules['required_fields']:
 if field not in context:
 validation_result['valid'] = False
 validation_result['errors'].append(f"Missing required field: {field}")

 # Check size limits
 context_size = len(str(context))
 if context_size > self.validation_rules['max_size']:
 validation_result['warnings'].append(
 f"Context size ({context_size}) exceeds recommended limit"
 )

 # Validate data types
 for key, value in context.items():
 if type(value) not in self.validation_rules['allowed_types']:
 validation_result['warnings'].append(
 f"Unexpected type for {key}: {type(value).__name__}"
 )

 # Apply sanitization
 sanitized_context = self.sanitize_context(context)

 return {
 'validation': validation_result,
 'context': sanitized_context
 }
```

## Performance Optimization

### Parallelization Strategies

Agent Parallelization:
```python
# Parallel agent execution optimization
class ParallelExecutor:
 def __init__(self):
 self.max_concurrent_agents = 5
 self.resource_pool = []
 self.execution_queue = []

 async def execute_parallel_agents(self, agent_tasks):
 """Execute multiple agents in parallel with resource management."""
 # Group tasks by resource requirements
 task_groups = self.group_tasks_by_resources(agent_tasks)

 # Execute groups concurrently
 group_results = []
 for group in task_groups:
 if len(group) <= self.max_concurrent_agents:
 # Execute small group directly
 group_result = await self.execute_concurrent_agents(group)
 group_results.extend(group_result)
 else:
 # Split large group into batches
 batches = self.create_batches(group, self.max_concurrent_agents)
 for batch in batches:
 batch_result = await self.execute_concurrent_agents(batch)
 group_results.extend(batch_result)

 return group_results

 def group_tasks_by_resources(self, tasks):
 """Group tasks by resource requirements."""
 groups = {
 'lightweight': [], # Low resource requirements
 'standard': [], # Standard resource needs
 'heavy': [] # High resource requirements
 }

 for task in tasks:
 resource_level = self.assess_resource_requirements(task)
 groups[resource_level].append(task)

 return groups

 def assess_resource_requirements(self, task):
 """Assess resource requirements for agent task."""
 # Simple assessment based on task complexity
 if task.get('complexity') == 'low':
 return 'lightweight'
 elif task.get('complexity') == 'high':
 return 'heavy'
 else:
 return 'standard'
```

### Caching and Optimization

Agent Result Caching:
```python
# Agent result caching for performance
class AgentCache:
 def __init__(self):
 self.cache = {}
 self.cache_ttl = 300 # 5 minutes
 self.max_cache_size = 1000

 def get_cached_result(self, agent_name, task_hash):
 """Get cached result for agent task."""
 cache_key = f"{agent_name}:{task_hash}"

 if cache_key not in self.cache:
 return None

 cached_item = self.cache[cache_key]

 # Check if cache is still valid
 if time.time() - cached_item['timestamp'] > self.cache_ttl:
 del self.cache[cache_key]
 return None

 return cached_item['result']

 def cache_result(self, agent_name, task_hash, result):
 """Cache agent result for future use."""
 cache_key = f"{agent_name}:{task_hash}"

 # Implement cache size limit
 if len(self.cache) >= self.max_cache_size:
 # Remove oldest entry
 oldest_key = min(self.cache.keys(),
 key=lambda k: self.cache[k]['timestamp'])
 del self.cache[oldest_key]

 self.cache[cache_key] = {
 'result': result,
 'timestamp': time.time(),
 'agent': agent_name,
 'task_hash': task_hash
 }

 def generate_task_hash(self, prompt, context):
 """Generate hash for task identification."""
 import hashlib

 # Create consistent hash from prompt and context
 task_data = {
 'prompt': prompt,
 'context_keys': list(context.keys()),
 'context_size': len(str(context))
 }

 task_string = json.dumps(task_data, sort_keys=True)
 return hashlib.md5(task_string.encode()).hexdigest()
```

## Advanced Integration Patterns

### 1. Agent Composition

Composite Agent Pattern:
```yaml
---
name: full-stack-specialist
description: Combine frontend, backend, database, and DevOps expertise for end-to-end application development. Use PROACTIVELY for complete application development requiring multiple domain expertise.
tools: Read, Write, Edit, Bash, Grep, Glob, Task, MultiEdit, WebFetch
model: sonnet
skills: moai-domain-backend, moai-domain-frontend, moai-domain-database, moai-devops-expert
---

# Full-Stack Development Specialist

You are a comprehensive full-stack development specialist with expertise across all application layers.

## Core Responsibilities

Primary Domain: End-to-end application development
Sub-Domains: Frontend, backend, database, DevOps
Integration Strategy: Coordinate specialized agents for domain-specific tasks

## Agent Delegation Patterns

### When to Delegate
- Frontend Complexity: Delegate to code-frontend
- Backend Architecture: Delegate to code-backend
- Database Design: Delegate to data-database
- Security Analysis: Delegate to security-expert
- Performance Optimization: Delegate to performance-engineer

### Delegation Examples
```python
# Full-stack agent delegation examples
def handle_full_stack_request(request):
 """Handle full-stack development request with intelligent delegation."""

 # Analyze request complexity and domains
 domain_analysis = analyze_request_domains(request)

 # Delegate specialized tasks
 results = {}

 if domain_analysis['frontend_required']:
 results['frontend'] = Task(
 subagent_type="code-frontend",
 prompt="Design and implement frontend components",
 context={"request": request, "analysis": domain_analysis}
 )

 if domain_analysis['backend_required']:
 results['backend'] = Task(
 subagent_type="code-backend",
 prompt="Design and implement backend API",
 context={"request": request, "analysis": domain_analysis, "frontend": results.get('frontend')}
 )

 if domain_analysis['database_required']:
 results['database'] = Task(
 subagent_type="data-database",
 prompt="Design database schema and optimization",
 context={"request": request, "analysis": domain_analysis, "frontend": results.get('frontend'), "backend": results.get('backend')}
 )

 # Integrate results
 integration_result = Task(
 subagent_type="integration-specialist",
 prompt="Integrate all components into cohesive application",
 context={"results": results, "request": request}
 )

 return {
 "domain_analysis": domain_analysis,
 "specialized_results": results,
 "integration": integration_result
 }
```

### 2. Adaptive Workflow Agents

Dynamic Agent Selection:
```python
# Adaptive workflow agent that adjusts based on project needs
class AdaptiveWorkflowAgent:
 def __init__(self):
 self.agent_capabilities = {
 'workflow-spec': {
 'complexity_threshold': 7,
 'task_types': ['specification', 'requirements', 'planning']
 },
 'workflow-ddd': {
 'complexity_threshold': 5,
 'task_types': ['implementation', 'development', 'coding']
 },
 'core-quality': {
 'complexity_threshold': 3,
 'task_types': ['validation', 'testing', 'quality']
 }
 }

 self.performance_metrics = {}

 def select_optimal_agent(self, task_request):
 """Select optimal agent based on task characteristics."""
 task_complexity = self.assess_task_complexity(task_request)
 task_type = self.classify_task_type(task_request)

 suitable_agents = []

 for agent_name, capabilities in self.agent_capabilities.items():
 if (task_type in capabilities['task_types'] and
 task_complexity <= capabilities['complexity_threshold']):
 suitable_agents.append({
 'agent': agent_name,
 'match_score': self.calculate_match_score(task_request, capabilities),
 'estimated_performance': self.get_agent_performance(agent_name)
 })

 # Select best agent based on match score and performance
 if suitable_agents:
 return max(suitable_agents, key=lambda x: x['match_score'] * x['estimated_performance'])

 # Fallback to generalist agent
 return {
 'agent': 'general-developer',
 'match_score': 0.5,
 'estimated_performance': 0.7
 }

 def assess_task_complexity(self, task_request):
 """Assess task complexity on scale 1-10."""
 complexity_factors = {
 'stakeholders': len(task_request.get('stakeholders', [])),
 'integrations': len(task_request.get('integrations', [])),
 'requirements': len(task_request.get('requirements', [])),
 'constraints': len(task_request.get('constraints', []))
 }

 # Calculate complexity score
 complexity_score = 0
 for factor, value in complexity_factors.items():
 complexity_score += min(value * 2, 10) # Cap at 10 per factor

 return min(complexity_score, 10)

 def calculate_match_score(self, task_request, agent_capabilities):
 """Calculate how well agent matches task requirements."""
 match_score = 0.0

 # Task type matching
 task_type = self.classify_task_type(task_request)
 if task_type in agent_capabilities['task_types']:
 match_score += 0.4

 # Experience level matching
 required_experience = task_request.get('experience_level', 'intermediate')
 agent_experience = agent_capabilities.get('experience_level', 'intermediate')
 if required_experience == agent_experience:
 match_score += 0.3

 # Tool requirement matching
 required_tools = set(task_request.get('required_tools', []))
 agent_tools = set(agent_capabilities.get('available_tools', []))
 tool_overlap = required_tools.intersection(agent_tools)
 if required_tools:
 match_score += 0.3 * (len(tool_overlap) / len(required_tools))

 return match_score
```

### 3. Learning Agents

Knowledge Accumulation:
```python
# Learning agent that improves from experience
class LearningAgent:
 def __init__(self):
 self.experience_database = {}
 self.success_patterns = {}
 self.failure_patterns = {}
 self.performance_history = []

 def learn_from_execution(self, agent_task, result, performance_metrics):
 """Learn from agent execution outcomes."""
 task_signature = self.create_task_signature(agent_task)

 learning_data = {
 'task': agent_task,
 'result': result,
 'performance': performance_metrics,
 'timestamp': datetime.now()
 }

 # Store experience
 self.experience_database[task_signature] = learning_data

 # Update performance history
 self.performance_history.append({
 'signature': task_signature,
 'performance': performance_metrics,
 'timestamp': datetime.now()
 })

 # Extract patterns
 if performance_metrics['success_rate'] > 0.8:
 self.extract_success_pattern(task_signature, learning_data)
 else:
 self.extract_failure_pattern(task_signature, learning_data)

 def recommend_strategy(self, current_task):
 """Recommend strategy based on learned patterns."""
 task_signature = self.create_task_signature(current_task)

 # Look for similar successful patterns
 similar_successes = self.find_similar_successful_patterns(task_signature)

 if similar_successes:
 best_pattern = max(similar_successes, key=lambda x: x['success_rate'])
 return best_pattern['strategy']

 # Look for failure patterns to avoid
 similar_failures = self.find_similar_failure_patterns(task_signature)
 if similar_failures:
 worst_pattern = max(similar_failures, key=lambda x: x['failure_rate'])
 return self.invert_pattern(worst_pattern['strategy'])

 # Default strategy
 return self.get_default_strategy(current_task)

 def create_task_signature(self, task):
 """Create unique signature for task."""
 signature_data = {
 'agent_type': task.get('agent_type'),
 'task_type': task.get('task_type'),
 'complexity': task.get('complexity'),
 'domain': task.get('domain'),
 'tools_required': sorted(task.get('tools_required', []))
 }

 return json.dumps(signature_data, sort_keys=True)
```

## Quality Assurance Integration

### Multi-Agent Quality Gates

Comprehensive Quality Framework:
```markdown
## Multi-Agent Quality Validation

### 1. Individual Agent Quality Checks
- Each sub-agent validates its own outputs
- Agent-specific quality metrics and standards
- Error handling and recovery validation
- Performance and efficiency assessment

### 2. Integration Quality Validation
- Validate agent communication and data transfer
- Check context passing and transformation accuracy
- Verify workflow integrity and completeness
- Assess overall system performance

### 3. End-to-End Quality Assurance
- Complete workflow testing and validation
- User acceptance criteria verification
- System integration testing
- Performance and scalability validation

### 4. Continuous Quality Improvement
- Monitor agent performance over time
- Identify improvement opportunities
- Update agent configurations and strategies
- Optimize agent selection and delegation patterns
```

Quality Metrics Dashboard:
```python
# Quality metrics tracking for multi-agent systems
class QualityMetricsTracker:
 def __init__(self):
 self.agent_metrics = {}
 self.workflow_metrics = {}
 self.quality_trends = {}

 def track_agent_performance(self, agent_name, execution_data):
 """Track individual agent performance metrics."""
 if agent_name not in self.agent_metrics:
 self.agent_metrics[agent_name] = {
 'executions': 0,
 'successes': 0,
 'failures': 0,
 'average_time': 0,
 'error_types': {}
 }

 metrics = self.agent_metrics[agent_name]
 metrics['executions'] += 1

 if execution_data['success']:
 metrics['successes'] += 1
 else:
 metrics['failures'] += 1
 error_type = execution_data.get('error_type', 'unknown')
 metrics['error_types'][error_type] = metrics['error_types'].get(error_type, 0) + 1

 metrics['average_time'] = self.update_average_time(
 metrics['average_time'],
 execution_data['execution_time'],
 metrics['executions']
 )

 def calculate_quality_score(self, agent_name):
 """Calculate comprehensive quality score for agent."""
 metrics = self.agent_metrics[agent_name]

 if metrics['executions'] == 0:
 return 0.0

 success_rate = metrics['successes'] / metrics['executions']

 # Quality factors
 quality_factors = {
 'success_rate': success_rate * 0.4,
 'performance': max(0, 1 - (metrics['average_time'] / 300)) * 0.3, # 5 minute baseline
 'reliability': min(1.0, 1 - (metrics['failures'] / metrics['executions'])) * 0.3
 }

 quality_score = sum(quality_factors.values())
 return quality_score
```

---

Version: 2.0.0
Compliance: Claude Code Official Standards
Last Updated: 2025-11-25
Integration Patterns: Sequential, Parallel, Conditional, Orchestrator
Advanced Features: Agent Composition, Adaptive Workflows, Learning Agents

Generated with Claude Code using official documentation and best practices.
