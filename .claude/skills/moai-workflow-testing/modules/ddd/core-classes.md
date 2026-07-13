# DDD with Documentation - Core Classes

> Sub-module: Core class implementations for DDD workflow management
> Parent: DDD module overview (see the main moai-workflow-testing SKILL)

The sketches below describe the shape of the DDD workflow objects in language-neutral pseudo-code. Implement them in the host project's language using its native enums/records/async idioms; test-runner and coverage commands should use the project's actual toolchain.

## Enumerations

### DDDPhase Enum

```text
enum DDDPhase:
    ANALYZE    # Analyzing existing code and behavior
    PRESERVE   # Creating characterization tests
    IMPROVE    # Refactoring with behavior preservation
    REVIEW     # Validation and documentation
```

### TestType Enum

```text
enum TestType:
    UNIT
    INTEGRATION
    ACCEPTANCE
    PERFORMANCE
    SECURITY
    REGRESSION
    CHARACTERIZATION
```

### TestStatus Enum

```text
enum TestStatus:
    PENDING
    PASSED
    FAILED
    SKIPPED
    ERROR
```

## Data Classes

### TestSpecification

```text
record TestSpecification:
    name:                text
    description:         text
    test_type:           TestType
    requirements:        List<text>
    acceptance_criteria: List<text>
    edge_cases:          List<text>
    mock_requirements:   List<text>     # default []
    fixture_requirements:List<text>     # default []
    timeout:             int?           # default none
    tags:                List<text>     # default []
```

### TestCase

```text
record TestCase:
    id:              text
    name:            text
    file_path:       text
    line_number:     int
    test_type:       TestType
    specification:   TestSpecification
    status:          TestStatus
    execution_time:  float?      # default none
    error_message:   text?       # default none
    code:            text        # default ""
    coverage_impact: float       # default 0.0
```

### DDDSession

```text
record DDDSession:
    id:                  text
    project_path:        text
    current_phase:       DDDPhase
    test_cases:          List<TestCase>
    implementation_files:List<text>
    metrics:             Map<text, Any>
    docs_patterns:   Map<text, Any>
    started_at:          timestamp
    last_activity:       timestamp
```

### DDDCycleResult

```text
record DDDCycleResult:
    session_id:                 text
    test_specification:         TestSpecification
    test_file_path:             text
    implementation_file_path:   text
    analyze_phase_result:       Map<text, Any>
    preserve_phase_result:      Map<text, Any>
    improve_phase_result:       Map<text, Any>
    final_coverage:             float
    total_time:                 float
    docs_patterns_applied:  List<text>
    behavior_preserved:         bool
```

## DDDManager Class

```text
class DDDManager:
    project_path
    docs            # optional client
    docs_integration
    test_generator
    current_session

    start_ddd_session(feature_name):
        session_id = "ddd_" + feature_name + "_" + now_epoch()
        patterns = await docs_integration.load_ddd_patterns()
        session = DDDSession(
            id=session_id,
            project_path=project_path,
            current_phase=DDDPhase.ANALYZE,
            test_cases=[],
            implementation_files=[],
            metrics={ analyze:0, preserve:0, improve:0 },
            docs_patterns=patterns,
            started_at=now(), last_activity=now())
        current_session = session
        return session

    run_full_ddd_cycle(specification, target_function):
        if current_session is none:
            current_session = start_ddd_session("default")
        cycle_start = now()
        patterns_applied = []

        # ANALYZE — understand existing code and behavior
        analyze_result = execute_analyze_phase(specification)
        current_session.metrics.analyze += 1

        # PRESERVE — create characterization tests
        preserve_result = execute_preserve_phase(specification, target_function, analyze_result)
        current_session.metrics.preserve += 1

        # IMPROVE — refactor with behavior preservation
        improve_result = execute_improve_phase(specification, preserve_result)
        current_session.metrics.improve += 1
        patterns_applied.extend(improve_result.patterns_applied)

        coverage = run_coverage_analysis()

        return DDDCycleResult(
            session_id=current_session.id,
            test_specification=specification,
            test_file_path=preserve_result.test_file_path,
            implementation_file_path=improve_result.implementation_file_path,
            analyze_phase_result=analyze_result,
            preserve_phase_result=preserve_result,
            improve_phase_result=improve_result,
            final_coverage=coverage.total_coverage,
            total_time=now() - cycle_start,
            docs_patterns_applied=patterns_applied,
            behavior_preserved=improve_result.behavior_preserved)
```

## Phase Execution Methods

### ANALYZE Phase

```text
execute_analyze_phase(specification):
    current_session.current_phase = DDDPhase.ANALYZE
    code_analysis       = analyze_existing_code(specification)
    behavior_patterns   = identify_behavior_patterns(code_analysis)
    refactoring_targets = identify_refactoring_targets(code_analysis)
    return {
        code_analysis, behavior_patterns, refactoring_targets,
        phase_success: true
    }
```

### PRESERVE Phase

```text
execute_preserve_phase(specification, target_function, analyze_result):
    current_session.current_phase = DDDPhase.PRESERVE

    # Generate characterization tests for existing behavior
    test_code = test_generator.generate_characterization_test(
        specification, analyze_result.behavior_patterns)
    test_file_path = get_test_file_path(specification)
    write_test_file(test_file_path, test_code)

    # Run tests — they should pass (they assert existing behavior)
    test_result = run_tests(test_file_path)

    test_case = TestCase(
        id="tc_" + specification.name,
        name=specification.name,
        file_path=test_file_path, line_number=1,
        test_type=TestType.CHARACTERIZATION,
        specification=specification,
        status=PASSED if test_result.failed == 0 else FAILED,
        code=test_code)
    current_session.test_cases.append(test_case)

    return {
        test_code, test_file_path, test_result, test_case,
        phase_success: test_result.failed == 0   # should pass in PRESERVE
    }
```

### IMPROVE Phase

```text
execute_improve_phase(specification, preserve_result):
    current_session.current_phase = DDDPhase.IMPROVE

    improve_patterns = docs_integration.get_improvement_patterns()
    improvements     = generate_improvements(preserve_result.implementation, improve_patterns)

    patterns_applied = []
    successful       = []
    behavior_preserved = true

    for improvement in improvements:
        improved = apply_improvement(preserve_result.implementation_file_path, improvement)
        if improved.success:
            # Re-run characterization tests to verify behavior preservation
            test_result = run_tests(preserve_result.test_file_path)
            if test_result.failed == 0:
                successful.append(improvement)
                patterns_applied.append(improvement.pattern or "custom")
            else:
                # Rollback failed improvement — behavior not preserved
                rollback_improvement(preserve_result.implementation_file_path)
                behavior_preserved = false

    return {
        improvements_suggested: len(improvements),
        improvements_applied:   len(successful),
        patterns_applied,
        behavior_preserved,
        phase_success: behavior_preserved
    }
```

## Helper Methods

```text
get_test_file_path(specification):
    # Resolve <project>/tests/<test_type>/test_<name><ext> using the
    # host language's conventional test file extension and runner layout.
    test_dir = project_path / "tests" / specification.test_type
    ensure_dir(test_dir)
    return test_dir / test_filename(specification.name)

get_implementation_file_path(target_function):
    src_dir = project_path / "src"
    ensure_dir(src_dir)
    return src_dir / source_filename(target_function)

run_tests(test_path):
    # Invoke the project's test runner (e.g. go test, pytest, jest, cargo test,
    # dotnet test) on test_path and parse passed/failed/error counts.
    output = exec(test_runner_cmd(test_path), cwd=project_path)
    return parse_test_summary(output)

run_coverage_analysis():
    # Invoke the project's coverage tool and read the total coverage figure.
    output = exec(coverage_cmd(), cwd=project_path)
    return { total_coverage: parse_total_coverage(output) default 0.0 }
```

## Related Sub-modules

- [Test Generation](./test-generation.md) - AI-powered test creation

---

Sub-module: `modules/ddd/core-classes.md`
