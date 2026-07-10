# ANALYZE-PRESERVE-IMPROVE DDD Cycle

> Module: Core DDD cycle implementation with Context7 integration
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: the project's test runner, Context7 MCP

## Core DDD Classes

```text
enum DDDPhase:
    ANALYZE    # analyze existing code and behavior
    PRESERVE   # create characterization tests
    IMPROVE    # improve code while keeping tests green
    REVIEW     # review and commit changes

enum TestType:
    UNIT, INTEGRATION, CHARACTERIZATION, ACCEPTANCE, PERFORMANCE, SECURITY, REGRESSION

enum TestStatus:
    PENDING, RUNNING, PASSED, FAILED, SKIPPED, ERROR

record TestSpecification:
    name:                text
    description:         text
    test_type:           TestType
    requirements:        List<text>
    acceptance_criteria: List<text>
    edge_cases:          List<text>
    preconditions:       List<text>      # default []
    postconditions:      List<text>      # default []
    dependencies:        List<text>      # default []
    mock_requirements:   Map<text, Any>  # default {}
    behavior_snapshot:   Map<text, Any>? # default none

record TestCase:
    id:            text
    name:          text
    file_path:     text
    line_number:   int
    specification: TestSpecification
    status:        TestStatus
    execution_time:float
    error_message: text?           # default none
    coverage_data: Map<text, Any>  # default {}

record DDDSession:
    id:                 text
    project_path:       text
    current_phase:      DDDPhase
    test_cases:         List<TestCase>
    start_time:         timestamp
    context7_patterns:  Map<text, Any>   # default {}
    metrics:            Map<text, Any>   # default {}
    behavior_snapshots: Map<text, Any>   # default {}
```

## DDD Manager Implementation

```text
class DDDManager:
    project_path
    context7
    current_session = none
    test_history = []

    start_ddd_session(feature_name, test_types = none):
        if test_types is none:
            test_types = [CHARACTERIZATION, UNIT, INTEGRATION]
        session = DDDSession(
            id="ddd_" + feature_name + "_" + now_epoch(),
            project_path=project_path,
            current_phase=DDDPhase.ANALYZE,
            test_cases=[], start_time=now(),
            context7_patterns={},
            metrics={ tests_written:0, tests_passing:0, tests_failing:0,
                      coverage_percentage:0.0, behaviors_preserved:0 },
            behavior_snapshots={})
        current_session = session
        return session

    run_full_ddd_cycle(specification, target_function = none):
        results = {}
        # ANALYZE
        log("ANALYZE Phase: Understanding existing code and behavior...")
        results.analyze  = run_analyze_phase(target_function)
        current_session.current_phase = DDDPhase.ANALYZE
        # PRESERVE
        log("PRESERVE Phase: Creating characterization tests...")
        results.preserve = run_preserve_phase(specification, results.analyze)
        current_session.current_phase = DDDPhase.PRESERVE
        # IMPROVE
        log("IMPROVE Phase: Refactoring with behavior preservation...")
        results.improve  = run_improve_phase(specification)
        current_session.current_phase = DDDPhase.IMPROVE
        # REVIEW
        log("REVIEW Phase: Final verification...")
        results.review   = { coverage: run_coverage_analysis() }
        current_session.current_phase = DDDPhase.REVIEW
        return results

    run_analyze_phase(target_function = none):
        analysis = { existing_tests: [], code_patterns: [], dependencies: [], behavior_notes: [] }
        # Find existing tests using the host language's test-file naming convention
        # (e.g. **/*_test.go, **/test_*.py, **/*.test.ts, **/tests/*.rs)
        analysis.existing_tests = glob(project_path, "**/" + test_glob())
        if target_function:
            analysis.target = target_function
            analysis.behavior_notes.append("Analyzing behavior of " + target_function)
        return analysis

    run_preserve_phase(specification, analysis):
        preserve = { characterization_tests_created: 0, behaviors_captured: [], test_files: [] }
        for behavior in analysis.behavior_notes default []:
            preserve.behaviors_captured.append(behavior)
            preserve.characterization_tests_created += 1
        # Run the existing suite to establish a green baseline
        preserve.baseline_results = run_tests()
        return preserve

    run_improve_phase(specification):
        improve = { improvements_made: [], tests_still_passing: true, refactoring_notes: [] }
        result = run_tests()
        improve.tests_still_passing = (result.failed == 0)
        if improve.tests_still_passing:
            current_session.metrics.behaviors_preserved += 1
        return improve

    run_tests():
        # Invoke the project's test runner (go test, pytest, jest, cargo test...)
        try:
            output = exec(test_runner_cmd(), cwd=project_path, capture=stdout)
            return parse_test_summary(output) default { passed: 0, failed: 0 }
        except e:
            log("Error running tests: " + e)
            return { error: text(e), passed: 0, failed: 0 }

    run_coverage_analysis():
        try:
            output = exec(coverage_cmd(), cwd=project_path, capture=stdout)
            return { coverage_output: output }
        except e:
            return { error: text(e) }

    get_session_summary():
        if current_session is none: return {}
        duration = now() - current_session.start_time
        return {
            session_id: current_session.id,
            phase: current_session.current_phase,
            duration_seconds: duration,
            metrics: current_session.metrics,
            test_cases_count: len(current_session.test_cases),
            behaviors_preserved: current_session.metrics.behaviors_preserved default 0
        }
```

## Phase-Specific Guidelines

### ANALYZE Phase
- Understand existing code structure and patterns
- Identify current behavior through code reading
- Document dependencies and side effects
- Map test coverage gaps
- Note existing design patterns

### PRESERVE Phase
- Write characterization tests for existing behavior
- Capture current behavior as the "golden standard"
- Ensure tests pass with current implementation
- Document discovered behavior
- Create behavior snapshots for complex outputs

### IMPROVE Phase
- Refactor code while keeping tests green
- Make small, incremental changes
- Run tests after each change
- Maintain behavior preservation
- Apply design patterns appropriately

### REVIEW Phase
- Verify all characterization tests still pass
- Review code quality and documentation
- Check for any behavior changes
- Commit changes with clear messages
- Document improvements made

## Usage Example

```text
# Initialize the DDD manager
ddd_manager = DDDManager(project_path="/path/to/project", context7_client=context7)

# Start a DDD session
session = ddd_manager.start_ddd_session("user_authentication_refactor")

# Create a test specification
test_spec = TestSpecification(
    name="test_user_login_behavior_preservation",
    description="Preserve existing login behavior during refactoring",
    test_type=TestType.CHARACTERIZATION,
    requirements=[
        "Existing login flow must continue to work",
        "Error messages should remain consistent"],
    acceptance_criteria=[
        "Valid credentials return a user token (existing behavior)",
        "Invalid credentials raise the same error messages"],
    edge_cases=[
        "Test with empty email (existing behavior)",
        "Test with empty password (existing behavior)"])

# Run the complete DDD cycle
cycle_results = ddd_manager.run_full_ddd_cycle(
    specification=test_spec, target_function="authenticate_user")

# Get the session summary
summary = ddd_manager.get_session_summary()
print("Session duration: " + summary.duration_seconds + "s")
print("Behaviors preserved: " + summary.behaviors_preserved)
```

---

Related: [Test Generation](./test-generation.md) | [Test Patterns](./test-patterns.md)
