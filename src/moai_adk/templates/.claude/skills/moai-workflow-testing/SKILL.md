---
name: moai-workflow-testing
description: Comprehensive development workflow specialist combining TDD, debugging, performance optimization, code review, and quality assurance into unified development workflows
version: 1.0.0
category: workflow
tags:
 - workflow
 - testing
 - debugging
 - performance
 - quality
 - tdd
 - review
updated: 2025-11-30
status: active
author: MoAI-ADK Team
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
---

# Development Workflow Specialist

## Quick Reference (30 seconds)

Unified Development Workflow - Comprehensive development lifecycle management combining TDD, AI-powered debugging, performance optimization, automated code review, and quality assurance into integrated workflows.

Core Capabilities:
- Test-Driven Development: RED-GREEN-REFACTOR cycle with Context7 patterns
- AI-Powered Debugging: Intelligent error analysis and Context7 best practices
- Performance Optimization: Real-time profiling and bottleneck detection
- Automated Code Review: TRUST 5 validation with AI quality analysis
- Quality Assurance: Comprehensive testing and CI/CD integration
- Workflow Orchestration: End-to-end development process automation

Unified Development Workflow:
```
Debug → Refactor → Optimize → Review → Test → Profile
 ↓ ↓ ↓ ↓ ↓ ↓
AI- AI- AI- AI- AI- AI-
Powered Powered Powered Powered Powered Powered
```

When to Use:
- Complete development lifecycle management
- Enterprise-grade quality assurance
- Multi-language development projects
- Performance-critical applications
- Technical debt reduction initiatives
- Automated testing and CI/CD integration

Quick Start:
```python
# Initialize comprehensive workflow
from moai_workflow_testing import (
 AIProfiler, TDDManager, AutomatedCodeReviewer
)

# Start complete development workflow
workflow = DevelopmentWorkflow(
 project_path="/project/src",
 context7_client=context7
)

# Run full workflow cycle
results = await workflow.execute_complete_cycle()
```

---

## Implementation Guide (5 minutes)

### Core Concepts

Unified Development Philosophy:
- Integrates all aspects of development into cohesive workflow
- AI-powered assistance for complex decision-making
- Context7 integration for industry best practices
- Continuous feedback loops between workflow stages
- Automated quality gates and validation

Workflow Components:
1. AI-Powered Debugging: Intelligent error classification and Context7-based solutions
2. Smart Refactoring: Technical debt analysis with safe automated transformations
3. Performance Optimization: Real-time monitoring with bottleneck detection
4. TDD with Context7: Enhanced test generation and RED-GREEN-REFACTOR cycles
5. Automated Code Review: TRUST 5 framework validation with AI analysis

### Basic Implementation

```python
from moai_workflow_testing import (
 AIDebugger, AIRefactorer, PerformanceProfiler,
 TDDManager, AutomatedCodeReviewer
)

# 1. AI-Powered Debugging
debugger = AIDebugger(context7_client=context7)

try:
 result = some_risky_operation()
except Exception as e:
 analysis = await debugger.debug_with_context7_patterns(
 e, {'file': __file__, 'function': 'main'}, '/project/src'
 )
 print(f"Found {len(analysis.solutions)} solutions")

# 2. Smart Refactoring
refactorer = AIRefactorer(context7_client=context7)
refactor_plan = await refactorer.refactor_with_intelligence('/project/src')
print(f"Found {len(refactor_plan.opportunities)} refactoring opportunities")

# 3. Performance Optimization
profiler = PerformanceProfiler(context7_client=context7)
profiler.start_profiling(['cpu', 'memory', 'line'])

# Run code to profile
result = expensive_function()

profile_results = profiler.stop_profiling()
bottlenecks = await profiler.detect_bottlenecks(profile_results)

# 4. TDD with Context7
tdd_manager = TDDManager('/project/src', context7_client=context7)

test_spec = TestSpecification(
 name="test_user_authentication",
 description="Test user authentication with valid credentials",
 test_type=TestType.UNIT,
 requirements=["Valid email and password required"],
 acceptance_criteria=["Valid credentials return token"]
)

cycle_results = await tdd_manager.run_full_tdd_cycle(
 specification=test_spec,
 target_function="authenticate_user"
)

# 5. Automated Code Review
reviewer = AutomatedCodeReviewer(context7_client=context7)
review_report = await reviewer.review_codebase('/project/src')

print(f"Overall TRUST Score: {review_report.overall_trust_score:.2f}")
print(f"Critical Issues: {len(review_report.critical_issues)}")
```

### Common Use Cases

Enterprise Development Workflow:
```python
# Complete enterprise workflow integration
workflow = EnterpriseWorkflow(
 project_path="/enterprise/app",
 context7_client=context7,
 quality_gates={
 'min_trust_score': 0.85,
 'max_critical_issues': 0,
 'required_coverage': 0.80
 }
)

# Execute workflow with quality validation
results = await workflow.execute_with_validation()
if results.quality_passed:
 print(" Ready for deployment")
else:
 print(" Quality gates not met")
 workflow.show_quality_issues()
```

Performance-Critical Applications:
```python
# Performance-focused workflow
perf_workflow = PerformanceWorkflow(
 project_path="/performance_app",
 context7_client=context7,
 performance_thresholds={
 'max_response_time': 100, # ms
 'max_memory_usage': 512, # MB
 'min_throughput': 1000 # requests/second
 }
)

# Profile and optimize
optimization_results = await perf_workflow.optimize_performance()
print(f"Performance improvement: {optimization_results.improvement_percentage:.1f}%")
```

---

## Advanced Features (10+ minutes)

### Workflow Integration Patterns

Continuous Integration Integration:
```python
# CI/CD pipeline integration
class CIWorkflowIntegrator:
 def __init__(self, workflow_system, ci_config):
 self.workflow = workflow_system
 self.config = ci_config

 async def run_ci_pipeline(self, commit_hash: str):
 """Run complete CI pipeline with workflow validation."""

 # 1. Code quality validation
 review_results = await self.workflow.run_code_review()
 if not self._meets_quality_standards(review_results):
 return self._create_failure_report("Code quality check failed")

 # 2. Testing validation
 test_results = await self.workflow.run_full_test_suite()
 if not test_results.all_tests_passed:
 return self._create_failure_report("Tests failed")

 # 3. Performance validation
 perf_results = await self.workflow.run_performance_tests()
 if not self._meets_performance_standards(perf_results):
 return self._create_failure_report("Performance standards not met")

 # 4. Security validation
 security_results = await self.workflow.run_security_analysis()
 if security_results.critical_vulnerabilities:
 return self._create_failure_report("Security issues found")

 return self._create_success_report(commit_hash)
```

### AI-Enhanced Decision Making

Context7-Powered Workflow Optimization:
```python
class AIWorkflowOptimizer:
 """AI-powered workflow optimization using Context7 patterns."""

 def __init__(self, context7_client):
 self.context7 = context7_client

 async def optimize_workflow_execution(
 self, project_context: Dict
 ) -> Dict[str, Any]:
 """Optimize workflow execution based on project characteristics."""

 # Get Context7 workflow patterns
 patterns = await self.context7.get_library_docs(
 context7_library_id="/workflow/devops",
 topic="optimal development workflow patterns 2025",
 tokens=4000
 )

 # Analyze project characteristics
 project_analysis = self._analyze_project_context(project_context)

 # Generate optimized workflow plan
 optimized_plan = await self._generate_optimized_workflow(
 project_analysis, patterns
 )

 return optimized_plan
```

### Advanced Quality Assurance

Comprehensive Quality Gates:
```python
class QualityGateManager:
 """Manages comprehensive quality gates across workflow stages."""

 def __init__(self, quality_config: Dict[str, Any]):
 self.gates = self._initialize_quality_gates(quality_config)

 async def validate_workflow_stage(
 self, stage: str, artifacts: Dict[str, Any]
 ) -> Dict[str, Any]:
 """Validate quality gates for specific workflow stage."""

 gate_config = self.gates.get(stage, {})
 validation_results = {}

 # Run stage-specific validations
 for gate_name, gate_config in gate_config.items():
 result = await self._run_quality_gate(
 gate_name, artifacts, gate_config
 )
 validation_results[gate_name] = result

 # Calculate overall gate status
 gate_passed = all(
 result['status'] == 'passed'
 for result in validation_results.values()
 )

 return {
 'stage': stage,
 'passed': gate_passed,
 'validations': validation_results,
 'recommendations': self._generate_recommendations(validation_results)
 }
```

---

## Works Well With

- moai-domain-backend - Backend development workflows and API testing
- moai-domain-frontend - Frontend development workflows and UI testing
- moai-foundation-core - Core SPEC system and workflow management
- moai-platform-supabase - Supabase testing patterns
- moai-platform-vercel - Vercel deployment testing
- moai-platform-firebase-auth - Firebase authentication testing
- moai-workflow-project - Project management and documentation workflows

---

## Module References

Core Implementation Modules:
- [`modules/ai-debugging.md`](./modules/ai-debugging.md) - AI-powered debugging with Context7 integration
- [`modules/smart-refactoring.md`](./modules/smart-refactoring.md) - Technical debt analysis and safe refactoring
- [`modules/performance-optimization.md`](./modules/performance-optimization.md) - Real-time profiling and bottleneck detection
- [`modules/tdd-context7.md`](./modules/tdd-context7.md) - TDD cycles with Context7-enhanced test generation
- [`modules/automated-code-review.md`](./modules/automated-code-review.md) - TRUST 5 validation with AI code review

---

## Usage Examples

### CLI Integration
```bash
# Run complete development workflow
moai-workflow execute --project /project/src --mode full

# Run specific workflow components
moai-workflow debug --file app.py --error "AttributeError"
moai-workflow refactor --directory src/ --max-risk medium
moai-workflow profile --target function_name --types cpu,memory
moai-workflow test --spec user_auth.spec --mode tdd
moai-workflow review --project /project/src --trust-score-min 0.8

# Continuous integration
moai-workflow ci --commit abc123 --quality-gates strict
```

### Python API
```python
from moai_workflow_testing import (
 DevelopmentWorkflow, WorkflowConfig
)

# Configure workflow
config = WorkflowConfig(
 enable_debugging=True,
 enable_refactoring=True,
 enable_profiling=True,
 enable_tdd=True,
 enable_code_review=True,
 context7_client=context7
)

# Initialize and run workflow
workflow = DevelopmentWorkflow(
 project_path="/project/src",
 config=config
)

# Execute complete workflow
results = await workflow.execute_complete_workflow()

# Access results by stage
print(f"Debugging solutions found: {len(results.debugging.solutions)}")
print(f"Refactoring opportunities: {len(results.refactoring.opportunities)}")
print(f"Performance bottlenecks: {len(results.profiling.bottlenecks)}")
print(f"Test coverage: {results.tdd.coverage_percentage:.1f}%")
print(f"Code review score: {results.code_review.trust_score:.2f}")
```

---

## Technology Stack

Core Analysis Libraries:
- cProfile: Python profiling and performance analysis
- memory_profiler: Memory usage analysis and optimization
- psutil: System resource monitoring
- line_profiler: Line-by-line performance profiling

Static Analysis Tools:
- pylint: Comprehensive code analysis and quality checks
- flake8: Style guide enforcement and error detection
- bandit: Security vulnerability scanning
- mypy: Static type checking and validation

Testing Frameworks:
- pytest: Advanced testing framework with fixtures and plugins
- unittest: Standard library testing framework
- coverage: Code coverage measurement and analysis

Context7 Integration:
- MCP Protocol: Context7 message passing and communication
- Dynamic Documentation: Real-time access to latest patterns and practices
- AI-Powered Analysis: Enhanced error analysis and solution generation

---

## Integration Examples

### GitHub Actions Integration
```yaml
# .github/workflows/development-workflow.yml
name: Development Workflow

on: [push, pull_request]

jobs:
 workflow:
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v3
 - name: Setup Python
 uses: actions/setup-python@v4
 with:
 python-version: '3.11'

 - name: Run Development Workflow
 run: |
 moai-workflow execute \
 --project . \
 --mode ci \
 --quality-gates strict \
 --output workflow-results.json

 - name: Upload Results
 uses: actions/upload-artifact@v3
 with:
 name: workflow-results
 path: workflow-results.json
```

### Docker Integration
```dockerfile
# Dockerfile for workflow execution
FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .

# Run complete workflow on container start
CMD ["moai-workflow", "execute", "--project", "/app", "--mode", "full"]
```

---

Status: Production Ready
Last Updated: 2025-11-30
Maintained by: MoAI-ADK Development Workflow Team
