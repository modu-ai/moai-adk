# Error Analysis and Solution Patterns

> Module: Comprehensive error categorization and solution patterns
> Complexity: Advanced
> Dependencies: debugging-workflows.md implementation module

## Pattern Matching System

### Error Pattern Matching

Comprehensive pattern matching implementation with regex support. The logic is language-neutral; map each language's error type names to the generic categories below.

```text
match_error_patterns(error, error_analysis):
    error_type = classify(error)            # e.g. IMPORT, TYPE, RUNTIME
    error_message = text(error)

    if error_type known in error_patterns:
        pattern_data = error_patterns[error_type]
        matched = []
        for regex in pattern_data.patterns:
            if regex matches error_message (case-insensitive):
                matched.append(regex)
        return {
            matched_patterns: matched,
            solutions:        pattern_data.solutions,
            docs_topics:  pattern_data.docs_topics
        }
    return { matched_patterns: [], solutions: [], docs_topics: [] }
```

### Solution Generation

Multi-source solution generation with confidence scoring:

```text
generate_solutions(error_analysis, docs_patterns, pattern_matches, context):
    solutions = []

    # Pattern-based solutions
    for pattern in pattern_matches.matched_patterns:
        solutions.append(Solution(
            type='pattern_match',
            description="Apply known pattern: " + pattern,
            code_example=generate_pattern_example(pattern, context),
            confidence=0.85, impact='medium'))

    # Documentation-based solutions
    for (library_id, docs) in docs_patterns:
        if docs has 'solutions':
            for sol in docs.solutions:
                solutions.append(Solution(
                    type='docs_pattern', description=sol.description,
                    confidence=sol.confidence or 0.7))

    # AI-generated solutions when few candidates exist
    if docs available and len(solutions) < 3:
        solutions.extend(generate_ai_solutions(error_analysis, context))

    sort solutions descending by (confidence, impact)
    return top 5
```

### Code Example Generation

Pattern-based code example generation. Examples below are illustrative fixes; the exact syntax and tooling differ per language.

```text
generate_pattern_example(pattern, context):
    examples = {
      "missing dependency":   "# Install the missing package with the project's package manager,
                              # then ensure the import path resolves.",
      "no such member":       "# Guard member access: check the type/member exists before use,
                              # or use the language's optional/safe-access idiom.",
      "wrong argument count": "# Reconcile the call site with the function/method signature;
                              # fix the number or shape of arguments passed.",
      "invalid value conversion": "# Validate/parse the value defensively and handle the
                                  # conversion-failure case instead of letting it throw.",
    }
    for (key, example) in examples:
        if pattern references key:
            return example
    return "# Implement fix for pattern: " + pattern
```

## Error Analysis Methods

### Severity Assessment

Comprehensive severity evaluation based on multiple factors:

```text
assess_severity(error, context, frequency):
    msg = lowercase(text(error))
    if any of ["critical", "fatal", "corruption", "security"] in msg:
        return "critical"
    if frequency > 10: return "high"
    if frequency > 3:  return "medium"
    if context.production:  return "high"
    if context.user_facing: return "medium"
    return "low"
```

### Likely causes Analysis

Root cause analysis for common error categories. Names below are generic; map each language's exception types (e.g. ImportError/ModuleNotFoundError, AttributeError, TypeError) onto these categories.

```text
analyze_likely_causes(error_category, message, context):
    causes = []
    if error_category == IMPORT:
        if message contains "not found"/"cannot resolve":
            causes += ["Missing dependency installation",
                       "Incorrect import path / module specifier",
                       "Dependency toolchain / lockfile out of sync"]
        if message contains "circular":
            causes += ["Circular dependency between modules",
                       "Improper module structure"]
    if error_category == MEMBER_ACCESS:
        causes += ["Wrong object type being used",
                   "Incorrect member name",
                   "Object not properly initialized"]
    if error_category == TYPE:
        causes += ["Incorrect data types in operation",
                   "Function called with wrong argument types",
                   "Missing type conversion"]
    return causes
```

### Quick Fix Generation

Rapid fix suggestions for immediate resolution:

```text
generate_quick_fixes(classification, message, context):
    fixes = []
    if classification == IMPORT:
        fixes += ["Install the missing package via the project package manager",
                  "Check the module resolution / include path configuration",
                  "Verify the module exists in the expected location"]
    if classification == MEMBER_ACCESS:
        fixes += ["Guard member access before use",
                  "Verify object initialization",
                  "Check for typos in the member name"]
    if classification == TYPE:
        fixes += ["Add an explicit type conversion before the operation",
                  "Check the function signature",
                  "Validate the runtime type before operating"]
    return fixes
```

## Prevention Strategies

### Type-Specific Prevention

Comprehensive prevention strategies by error type:

```text
suggest_prevention_strategies(error_analysis, context):
    strategies = []
    if error_analysis.type == IMPORT:
        strategies += ["Pin dependencies in a lockfile / manifest",
                       "Implement module availability checks before imports",
                       "Use isolated dependency environments per project"]
    if error_analysis.type == MEMBER_ACCESS:
        strategies += ["Guard or option-type access before member use",
                       "Implement proper object type checking",
                       "Add unit tests for object interfaces"]
    if error_analysis.type == TYPE:
        strategies += ["Adopt static type checking (where the language supports it)",
                       "Add runtime type validation at boundaries",
                       "Validate types before operations"]
    if error_analysis.type == VALUE:
        strategies += ["Add input validation at function boundaries",
                       "Implement comprehensive error handling",
                       "Parse/convert defensively with failure handling"]
    # General
    strategies += ["Implement structured logging for error tracking",
                   "Add automated testing to catch errors early",
                   "Use code review to prevent common issues"]
    return strategies
```

### Related Error Detection

Identify related errors that frequently occur together. The generic categories below stand in for language-specific exception types.

```text
find_related_errors(error_analysis):
    related_map = {
        IMPORT:         [RESOLUTION, RUNTIME, MEMBER_ACCESS],
        MEMBER_ACCESS:  [TYPE, KEY, IMPORT],
        TYPE:           [VALUE, MEMBER_ACCESS, TYPE],
        VALUE:          [TYPE, KEY, INDEX],
        KEY:            [MEMBER_ACCESS, TYPE, INDEX],
    }
    return related_map[error_analysis.type] default [TYPE, VALUE]
```

## Fix Time Estimation

### Time Estimation Algorithm

Predict fix time based on error type and solution confidence. (Estimates are rough guidance, not commitments.)

```text
estimate_fix_time(error_analysis, solutions):
    base_times = {
        SYNTAX:   "1-5 minutes",     IMPORT:   "2-10 minutes",
        MEMBER_ACCESS: "5-15 minutes", TYPE: "5-20 minutes",
        VALUE:    "2-15 minutes",    KEY:      "2-10 minutes",
        NETWORK:  "10-30 minutes",   DATABASE: "15-45 minutes",
        MEMORY:   "20-60 minutes",   CONCURRENCY: "30-90 minutes",
        UNKNOWN:  "15-60 minutes"
    }
    base = base_times[error_analysis.type] default "10-30 minutes"
    if solutions[0].confidence > 0.9: return "Quick fix: " + base
    if solutions[0].confidence > 0.7: return "Standard: " + base
    return "Complex: " + base
```

## Statistics and Monitoring

### Debug Statistics

Comprehensive debugging session statistics:

```text
debug_statistics():
    return {
        total_errors_analyzed: len(error_history),
        error_types:           count by error category,
        cache_hits:            len(pattern_cache),
        most_common_errors:    top 5 by frequency
    }
```

### Error Frequency Tracking

Monitor error occurrence patterns:

```text
error_frequency(error):
    key = category(error) + ":" + first 50 chars of message(error)
    return error_history[key] default 0
```

### Cache Management

Optimize Documentation query caching:

```text
clear_error_history():
    error_history.clear()
    pattern_cache.clear()
```

## Confidence Calculation

### Classification Confidence

Calculate confidence in error classification:

```text
calculate_confidence(classification, message):
    if classification != UNKNOWN:  return 0.85   # direct category match
    return 0.4                                   # unknown — low confidence
```

## Best Practices

Solution Prioritization: Apply solutions with highest confidence scores first and validate each fix in isolation before integration

Pattern Recognition: Track error patterns over time to identify systemic issues requiring architectural improvements

Prevention Strategy Implementation: Prioritize prevention strategies based on error frequency and severity impact

Learning Integration: Record successful fixes to improve pattern recognition and solution accuracy over time

Performance Optimization: Use caching for Documentation queries and implement batch processing for multiple errors

Documentation Updates: Maintain error pattern database with latest solutions and Documentation topics for continuous improvement

---

Module: modules/debugging/error-analysis.md
Version: 2.0.0 (Modular Architecture)
Last Updated: 2025-12-07
Lines: 350 (within 500-line limit)
