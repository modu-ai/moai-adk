# Debugging Workflows and Implementation

> Module: AI debugging process patterns and implementation workflows
> Complexity: Advanced
> Dependencies: ai-debugging.md overview module

## Core Implementation

### AIDebugger Class Structure

Complete AIDebugger implementation with Context7 integration, pattern matching, and solution generation. For data class definitions (ErrorType, ErrorAnalysis, Solution, DebugAnalysis), see [ai-debugging.md](../ai-debugging.md). The error categories below are generic; map each language's exception types onto them (e.g. ImportError/ModuleNotFoundError, AttributeError, TypeError, ValueError).

```text
class AIDebugger:
    context7
    error_patterns = load_error_patterns()   # generic-category -> {patterns, solutions, topics}
    error_history = {}
    pattern_cache = {}

    load_error_patterns():
        return {
          IMPORT: {
            patterns:  ["cannot resolve module", "no module named", "circular import"],
            solutions: ["Install the missing dependency",
                        "Check the import path / module specifier",
                        "Resolve circular dependencies"],
            context7_topics: ["module resolution", "dependency management"]
          },
          MEMBER_ACCESS: {
            patterns:  ["has no member", "no such attribute", "is not a function"],
            solutions: ["Check the object type and available members",
                        "Verify the module import / member name",
                        "Add the missing member or method"],
            context7_topics: ["attribute access patterns", "introspection techniques"]
          },
          TYPE: {
            patterns:  ["unsupported operand types", "wrong number of arguments",
                        "cannot be converted to", "expected .* got"],
            solutions: ["Check data types before operations",
                        "Verify the function signature",
                        "Add type validation / conversion"],
            context7_topics: ["type system debugging", "type checking best practices"]
          },
          VALUE: {
            patterns:  ["invalid value", "cannot convert", "out of range", "empty"],
            solutions: ["Validate input data format",
                        "Add error handling for conversions",
                        "Check value ranges"],
            context7_topics: ["input validation patterns", "data conversion error handling"]
          }
        }
```

### Main Debugging Method

Complete debug workflow implementation with error classification, pattern matching, solution generation, and prevention strategies:

```text
    debug_with_context7_patterns(error, context, codebase_path):
        # Classify the error
        error_analysis = classify_error_with_ai(error, context)

        # Fetch Context7 patterns if available
        context7_patterns = {}
        if context7 available:
            context7_patterns = get_context7_patterns(error_analysis)

        # Match against known patterns
        pattern_matches = match_error_patterns(error, error_analysis)

        # Generate comprehensive solutions
        solutions = generate_solutions(error_analysis, context7_patterns, pattern_matches, context)

        # Suggest prevention strategies
        prevention = suggest_prevention_strategies(error_analysis, context)

        # Estimate fix time
        fix_time = estimate_fix_time(error_analysis, solutions)

        return DebugAnalysis(
            error_type=error_analysis.type,
            confidence=error_analysis.confidence,
            context7_patterns=context7_patterns,
            solutions=solutions,
            prevention_strategies=prevention,
            related_errors=find_related_errors(error_analysis),
            estimated_fix_time=fix_time)
```

### Error Classification System

AI-enhanced error classification with context awareness using type mapping, message patterns, and contextual analysis. The type-mapping table is populated with the host language's exception names at runtime.

```text
    classify_error_with_ai(error, context):
        error_type_name = classify(error)
        error_message   = text(error)
        error_traceback = format_stack_trace(error)

        classification = classify_by_type_and_message(error_type_name, error_message, context)

        # Frequency tracking
        key = error_type_name + ":" + first 50 chars of error_message
        frequency = error_history[key] + 1
        error_history[key] = frequency

        severity       = assess_severity(error, context, frequency)
        likely_causes  = analyze_likely_causes(error_type_name, error_message, context)
        suggested_fixes= generate_quick_fixes(classification, error_message, context)

        return ErrorAnalysis(
            type=classification,
            confidence=calculate_confidence(classification, error_message),
            message=error_message, traceback=error_traceback,
            context=context, frequency=frequency, severity=severity,
            likely_causes=likely_causes, suggested_fixes=suggested_fixes)

    classify_by_type_and_message(error_type, message, context):
        # Direct mapping from the host language's exception names to generic categories
        type_mapping = <populated per language>   # e.g. ImportError -> IMPORT, KeyError -> KEY
        if error_type in type_mapping: return type_mapping[error_type]

        # Message-keyword classification
        msg = lowercase(message)
        if any of ["connection","timeout","network","http","socket"] in msg: return NETWORK
        if any of ["database","sql","query","cursor"]            in msg: return DATABASE
        if any of ["memory","out of memory","allocation","heap"] in msg: return MEMORY
        if any of ["thread","lock","race condition","concurrent"]in msg: return CONCURRENCY

        # Context-based classification
        if context.operation_type == "database": return DATABASE
        if context.operation_type == "network":  return NETWORK
        return UNKNOWN
```

### Context7 Integration

Automatic documentation retrieval for latest debugging patterns and best practices with intelligent caching:

```text
    get_context7_patterns(error_analysis):
        cache_key = error_analysis.type + "_" + first 30 chars of message
        if cache_key in pattern_cache: return pattern_cache[cache_key]

        queries = build_context7_queries(error_analysis)
        patterns = {}
        if context7 available:
            for (library_id, topic) in queries:
                try:
                    patterns[library_id] = context7.get_library_docs(
                        library_id, topic, tokens=4000)
                except e:
                    log("Context7 query failed for " + library_id + ": " + e)
        pattern_cache[cache_key] = patterns
        return patterns

    build_context7_queries(error_analysis):
        queries = []
        # Base debugging library / docs
        queries.append(("<debugging-library>",
                        "AI debugging patterns " + error_analysis.type + " error analysis"))
        # Language-specific library when the language is known
        lang = error_analysis.context.language
        if lang:
            queries.append(("<" + lang + " stdlib/runtime>",
                            error_analysis.type + " debugging best practices"))
        # Framework-specific query when a framework is named
        framework = error_analysis.context.framework
        if framework:
            queries.append(("<" + framework + ">",
                            framework + " " + error_analysis.type + " troubleshooting"))
        return queries
```

## Advanced Implementation Patterns

### Learning Debugger Extension

Self-improving debugger that learns from successful fixes with pattern recognition and success rate tracking:

```text
class LearningDebugger(AIDebugger):
    learned_patterns  = {}
    successful_fixes  = {}

    record_successful_fix(error_signature, applied_solution):
        successful_fixes.setdefault(error_signature, []).append({
            solution:     applied_solution,
            timestamp:    now_iso(),
            success_rate: 1.0
        })

    get_learned_solutions(error_signature):
        learned = successful_fixes.get(error_signature, [])
        solutions = []
        for fix in learned:
            if fix.success_rate > 0.7:
                solutions.append(Solution(
                    type='learned_pattern',
                    description="Previously successful fix: " + fix.solution,
                    code_example=fix.solution,
                    confidence=fix.success_rate, impact='high'))
        return solutions
```

### Enhanced Context Collection

Comprehensive debug context extraction with stack frame analysis for improved error classification. Use the host language's reflection/introspection API to walk the stack.

```text
collect_debug_context(error, frame_depth = 5):
    context = {
        error_type:    classify(error),
        error_message: text(error),
        timestamp:     now_iso(),
        runtime:       runtime_version(),     # language runtime + version
        stack_trace:   []
    }
    frame = current_stack_frame()
    for _ in 0..frame_depth:
        if frame is none: break
        context.stack_trace.append({
            filename: frame.file,
            function: frame.function,
            lineno:   frame.line,
            locals:   keys(frame.local_variables)
        })
        frame = frame.caller
    return context
```

### Usage Examples

Complete usage examples for common debugging scenarios. The try/catch idiom varies by language; the analysis call shape is identical.

```text
# Basic usage
debugger = AIDebugger(context7_client=context7)
try:
    result = some_risky_operation()
catch e:
    analysis = debugger.debug_with_context7_patterns(
        e,
        { file: current_file, function: "some_risky_operation", language: <lang> },
        "/project/src")
    print("Error type: " + analysis.error_type)
    print("Confidence: " + analysis.confidence)
    print("Solutions found: " + len(analysis.solutions))
    for (i, solution) in enumerate(analysis.solutions, from=1):
        print("  Solution " + i + ": " + solution.description + " (conf " + solution.confidence + ")")

# Advanced usage with custom context
try:
    data = process_user_input(user_data)
catch e:
    analysis = debugger.debug_with_context7_patterns(
        e,
        { file: current_file, function: "process_user_input", language: <lang>,
          framework: <framework>, operation_type: "data_processing",
          user_facing: true, production: false },
        "/project/src")
    print("Prevention strategies:")
    for strategy in analysis.prevention_strategies: print(" - " + strategy)

# Check debug statistics
stats = debugger.debug_statistics()
print("Debugged " + stats.total_errors_analyzed + " errors")
print("Most common: " + first 3 of stats.most_common_errors)
```

## Best Practices

Context Collection: Always provide comprehensive context including file paths, function names, and relevant variables for accurate analysis

Error Categorization: Use specific error types for better pattern matching and solution relevance

Solution Validation: Test proposed solutions in isolated environment before applying to production

Learning Integration: Record successful fixes to improve pattern recognition over time

Performance Monitoring: Track debugging session performance and cache efficiency for optimization

Module Statistics Tracking: Monitor error frequency and patterns to identify systemic issues

## Related Modules

Pattern Matching: [error-analysis.md](./error-analysis.md) - Comprehensive error categorization and solution patterns

Implementation Details: See methods for severity assessment, likely causes analysis, and prevention strategies

---

Module: modules/debugging/debugging-workflows.md
Version: 2.0.0 (Modular Architecture)
Last Updated: 2025-12-07
Lines: 350 (within 500-line limit)
