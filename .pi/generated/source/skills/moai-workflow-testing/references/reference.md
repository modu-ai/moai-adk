# moai-workflow-testing Reference

Progressive Disclosure Level 2: Extended documentation for testing, debugging, and quality workflows.

---

## API Reference

### Core Classes

DevelopmentWorkflow:
- Purpose: Unified development lifecycle management
- Initialization: `DevelopmentWorkflow(project_path, config)`
- Primary Methods:
  - `execute_complete_workflow()` - Run full development cycle
  - `execute_complete_cycle()` - Async complete cycle execution
  - `run_code_review()` - Execute code review stage
  - `run_full_test_suite()` - Execute all tests
  - `run_performance_tests()` - Execute performance tests
  - `run_security_analysis()` - Execute security analysis

AIDebugger:
- Purpose: AI-powered intelligent debugging
- Initialization: `AIDebugger(context7_client)`
- Key Methods:
  - `debug_with_context7_patterns(exception, context, project_path)` - Debug with AI assistance
  - `classify_error(exception)` - Classify error type
  - `suggest_solutions(analysis)` - Generate fix suggestions
  - `apply_fix(solution)` - Apply recommended fix

AIRefactorer:
- Purpose: Intelligent code refactoring
- Initialization: `AIRefactorer(context7_client)`
- Key Methods:
  - `refactor_with_intelligence(project_path)` - Analyze and suggest refactoring
  - `analyze_technical_debt(project_path)` - Calculate debt score
  - `generate_refactor_plan(opportunities)` - Create refactoring roadmap
  - `apply_safe_refactor(plan)` - Apply safe refactoring operations

PerformanceProfiler:
- Purpose: Real-time performance analysis
- Initialization: `PerformanceProfiler(context7_client)`
- Key Methods:
  - `start_profiling(profile_types)` - Begin profiling session
  - `stop_profiling()` - End profiling and get results
  - `detect_bottlenecks(results)` - Identify performance issues
  - `suggest_optimizations(bottlenecks)` - Generate optimization suggestions

TDDManager:
- Purpose: Test-driven development cycle management
- Initialization: `TDDManager(project_path, context7_client)`
- Key Methods:
  - `run_full_tdd_cycle(specification, target_function)` - Complete RED-GREEN-REFACTOR
  - `generate_test(spec)` - Generate test from specification
  - `run_tests()` - Execute test suite
  - `calculate_coverage()` - Get test coverage metrics

AutomatedCodeReviewer:
- Purpose: AI-powered code review with TRUST 5
- Initialization: `AutomatedCodeReviewer(context7_client)`
- Key Methods:
  - `review_codebase(project_path)` - Full codebase review
  - `calculate_trust_score(review_data)` - Calculate TRUST 5 score
  - `identify_critical_issues(review_data)` - Find critical problems
  - `generate_review_report(review_data)` - Create detailed report

---

## Configuration Options

### WorkflowConfig Schema

```python
@dataclass
class WorkflowConfig:
    # Module enablement
    enable_debugging: bool = True
    enable_refactoring: bool = True
    enable_profiling: bool = True
    enable_tdd: bool = True
    enable_code_review: bool = True

    # Context7 integration
    context7_client: Optional[Context7Client] = None

    # Quality thresholds
    min_trust_score: float = 0.85
    max_critical_issues: int = 0
    required_coverage: float = 0.80

    # Performance thresholds
    max_response_time_ms: int = 100
    max_memory_usage_mb: int = 512
    min_throughput_rps: int = 1000

    # Profiling options
    profile_types: List[str] = field(default_factory=lambda: ['cpu', 'memory'])

    # Output options
    output_format: str = 'json'  # json, html, markdown
    verbose: bool = False
```

### TestSpecification Schema

```python
@dataclass
class TestSpecification:
    name: str                           # Test name (test_*)
    description: str                    # What is being tested
    test_type: TestType                 # UNIT, INTEGRATION, E2E
    requirements: List[str]             # Test requirements
    acceptance_criteria: List[str]      # Pass/fail criteria
    fixtures: Optional[List[str]] = None  # Required fixtures
    mocks: Optional[List[str]] = None   # Required mocks
    timeout_seconds: int = 30           # Test timeout
```

### Quality Gates Configuration

```yaml
quality_gates:
  code_review:
    min_trust_score: 0.85          # Minimum TRUST 5 score
    max_critical_issues: 0         # Maximum critical issues
    max_high_issues: 5             # Maximum high-severity issues
    required_sections:
      - security
      - performance
      - maintainability

  testing:
    required_coverage: 0.80        # Minimum test coverage
    max_test_failures: 0           # Maximum allowed failures
    timeout_multiplier: 2.0        # CI timeout multiplier

  performance:
    response_time_p99_ms: 200      # 99th percentile response time
    memory_growth_threshold: 0.1   # Max memory growth percentage
    cpu_threshold: 0.8             # Max CPU utilization

  security:
    max_vulnerabilities: 0         # Maximum security vulnerabilities
    allowed_severity: ["low"]      # Allowed vulnerability severities
```

---

## Integration Patterns

### Pattern 1: CI/CD Pipeline Integration

```python
# Complete CI/CD integration with quality gates
from moai_workflow_testing import DevelopmentWorkflow, WorkflowConfig

async def ci_pipeline(commit_hash: str, project_path: str):
    """Run full CI pipeline with all quality gates."""

    config = WorkflowConfig(
        enable_debugging=False,  # Skip debugging in CI
        enable_refactoring=False,  # Skip refactoring suggestions
        min_trust_score=0.85,
        required_coverage=0.80
    )

    workflow = DevelopmentWorkflow(project_path, config)

    # Stage 1: Code Review
    review_result = await workflow.run_code_review()
    if review_result.trust_score < config.min_trust_score:
        return {"status": "failed", "stage": "code_review", "details": review_result}

    # Stage 2: Testing
    test_result = await workflow.run_full_test_suite()
    if not test_result.all_passed:
        return {"status": "failed", "stage": "testing", "details": test_result}

    # Stage 3: Performance
    perf_result = await workflow.run_performance_tests()
    if not perf_result.meets_thresholds:
        return {"status": "failed", "stage": "performance", "details": perf_result}

    # Stage 4: Security
    security_result = await workflow.run_security_analysis()
    if security_result.has_critical:
        return {"status": "failed", "stage": "security", "details": security_result}

    return {"status": "passed", "commit": commit_hash}
```

### Pattern 2: TDD Workflow with Context7

```python
# Enhanced TDD with Context7 best practices
from moai_workflow_testing import TDDManager, TestSpecification, TestType

async def implement_feature_tdd(feature_spec: dict, context7_client):
    """Implement feature using TDD with Context7 patterns."""

    ddd = DDDManager("/project/src", context7_client=context7_client)

    # Create test specification from feature
    test_spec = TestSpecification(
        name=f"test_{feature_spec['name']}",
        description=feature_spec['description'],
        test_type=TestType.UNIT,
        requirements=feature_spec['requirements'],
        acceptance_criteria=feature_spec['acceptance_criteria']
    )

    # RED: Generate failing test
    failing_test = await ddd.generate_test(test_spec)
    assert await ddd.run_tests() == False  # Should fail

    # GREEN: Implement minimum code
    implementation = await ddd.generate_implementation(test_spec)
    assert await ddd.run_tests() == True  # Should pass

    # REFACTOR: Optimize with Context7 patterns
    refactored = await ddd.refactor_with_patterns(implementation)
    assert await ddd.run_tests() == True  # Still passes

    # Coverage check
    coverage = await ddd.calculate_coverage()
    assert coverage >= 0.85

    return {
        "test": failing_test,
        "implementation": refactored,
        "coverage": coverage
    }
```

### Pattern 3: AI-Powered Debugging Session

```python
# Intelligent debugging with automatic fix application
from moai_workflow_testing import AIDebugger

async def debug_production_error(exception, context, context7_client):
    """Debug production error with AI assistance."""

    debugger = AIDebugger(context7_client=context7_client)

    # Analyze error with Context7 patterns
    analysis = await debugger.debug_with_context7_patterns(
        exception=exception,
        context=context,
        project_path="/project/src"
    )

    print(f"Error Classification: {analysis.classification}")
    print(f"Root Cause: {analysis.root_cause}")
    print(f"Solutions Found: {len(analysis.solutions)}")

    # Apply recommended fix
    if analysis.solutions:
        best_solution = analysis.solutions[0]
        if best_solution.confidence > 0.9:
            result = await debugger.apply_fix(best_solution)
            print(f"Fix Applied: {result.success}")
            return result

    return analysis
```

### Pattern 4: Performance Optimization Workflow

```python
# Comprehensive performance optimization
from moai_workflow_testing import PerformanceProfiler

async def optimize_critical_path(function_name: str, context7_client):
    """Profile and optimize critical code path."""

    profiler = PerformanceProfiler(context7_client=context7_client)

    # Multi-dimensional profiling
    profiler.start_profiling(['cpu', 'memory', 'line', 'io'])

    # Execute target function
    result = await execute_function(function_name)

    # Collect and analyze results
    profile_results = profiler.stop_profiling()

    # Detect bottlenecks
    bottlenecks = await profiler.detect_bottlenecks(profile_results)

    # Generate optimizations
    optimizations = await profiler.suggest_optimizations(bottlenecks)

    # Create optimization report
    report = {
        "function": function_name,
        "original_metrics": profile_results.summary,
        "bottlenecks": [b.to_dict() for b in bottlenecks],
        "optimizations": [o.to_dict() for o in optimizations],
        "estimated_improvement": calculate_improvement(optimizations)
    }

    return report
```

---

## Troubleshooting

### Common Issues

Issue: Context7 connection timeout:
- Cause: MCP server not responding or network issues
- Solution: Check Context7 server status and restart if needed
- Prevention: Implement retry logic with exponential backoff

Issue: Test coverage below threshold:
- Cause: Insufficient test cases or uncovered code paths
- Solution: Use `ddd.identify_uncovered_paths()` to find gaps
- Prevention: Run coverage check before commits

Issue: Performance profiler high overhead:
- Cause: Line profiler active on large files
- Solution: Use targeted profiling on specific functions
- Prevention: Configure `max_line_profile_lines` threshold

Issue: Code review false positives:
- Cause: Overly strict rule configuration
- Solution: Adjust rule severity in `quality_gates` config
- Prevention: Tune rules based on project context

Issue: TDD cycle stuck in RED phase:
- Cause: Specification too complex or ambiguous
- Solution: Break down specification into smaller units
- Prevention: Follow EARS format for clear specifications

### Diagnostic Commands

```bash
# CLI diagnostics
moai-workflow diagnose --full-check
moai-workflow test --dry-run --verbose
moai-workflow profile --list-available
moai-workflow review --show-rules

# Python diagnostics
from moai_workflow_testing import diagnose

# Component health check
report = await diagnose.check_all_components()
print(f"Debugger: {report.debugger_status}")
print(f"Profiler: {report.profiler_status}")
print(f"TDD Manager: {report.tdd_status}")
print(f"Reviewer: {report.reviewer_status}")
```

### Log Locations

- Workflow logs: `.moai/logs/workflow.log`
- Test results: `.moai/test-results/`
- Profile data: `.moai/profiles/`
- Review reports: `.moai/reviews/`
- Debug sessions: `.moai/debug-sessions/`

---

## External Resources

### Official Documentation

- pytest Documentation: https://docs.pytest.org/
- Python Profiling Guide: https://docs.python.org/3/library/profile.html
- Coverage.py: https://coverage.readthedocs.io/

### Analysis Tool References

- pylint: https://pylint.readthedocs.io/
- flake8: https://flake8.pycqa.org/
- bandit: https://bandit.readthedocs.io/
- mypy: https://mypy.readthedocs.io/

### Related Skills

- moai-foundation-core - TRUST 5 framework and SPEC-First TDD
- moai-domain-backend - Backend development workflows
- moai-domain-frontend - Frontend development workflows
- moai-platform-baas - Backend-as-a-Service integration

### Module References

- AI Debugging: `modules/ai-debugging.md`
- Smart Refactoring: `modules/smart-refactoring.md`
- Performance Optimization: `modules/performance-optimization.md`
- DDD with Context7: `modules/ddd-context7.md`
- Automated Code Review: `modules/automated-code-review.md`

### Best Practices

Testing:
- Write tests before implementation (TDD)
- Maintain 85%+ coverage
- Use meaningful test names
- Mock external dependencies
- Run tests in isolation

Debugging:
- Capture full stack traces
- Include reproduction steps
- Document environment details
- Use structured logging
- Apply fixes incrementally

Performance:
- Profile before optimizing
- Focus on bottlenecks
- Measure improvement
- Document baseline metrics
- Test under realistic load

Code Review:
- Follow TRUST 5 guidelines
- Adddess critical issues first
- Document review decisions
- Track technical debt
- Share knowledge through reviews

### Version History

| Version | Date       | Changes                                           |
|---------|------------|---------------------------------------------------|
| 1.0.0   | 2025-11-30 | Initial unified workflow release                  |
| 0.9.0   | 2025-11-25 | Added Context7 integration                        |
| 0.8.0   | 2025-11-20 | Added performance optimization module             |
| 0.7.0   | 2025-11-15 | Added automated code review                       |

---

Status: Reference Documentation Complete
Last Updated: 2025-12-06
Skill Version: 1.0.0
