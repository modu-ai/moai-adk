# Debugging / Refactoring / Performance Workflows — Step-by-Step

All three workflows share a 6-step pattern: capture → analyze → identify candidates → apply → verify → document.

## Debugging Workflow Process

1. Capture the error with full context: stack trace, environment, recent code changes
2. Classify the error type: syntax, runtime, logic, integration, or performance
3. Analyze the error pattern against known issue databases and best practices
4. Generate solution candidates ranked by likelihood of success
5. Apply the recommended fix and verify resolution
6. Document the issue and solution for future reference

## Refactoring Workflow Process

1. Analyze the target codebase for code smells and technical debt indicators
2. Calculate complexity metrics including cyclomatic complexity and coupling
3. Identify refactoring opportunities with associated risk levels
4. Generate a refactoring plan with prioritized actions
5. Apply refactoring transformations in safe increments
6. Verify behavior preservation through test execution

## Performance Optimization Process

1. Configure profiling for target metrics: CPU, memory, I/O, network
2. Execute profiling runs under representative load conditions
3. Analyze profiling results to identify bottlenecks
4. Generate optimization recommendations with expected impact estimates
5. Apply optimizations in isolation to measure individual effects
6. Validate overall performance improvement against baseline metrics
