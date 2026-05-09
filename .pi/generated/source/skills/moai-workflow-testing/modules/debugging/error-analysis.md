# Error Analysis and Solution Patterns

> Module: Comprehensive error categorization and solution patterns
> Complexity: Advanced
> Dependencies: debugging-workflows.md implementation module

## Pattern Matching System

### Error Pattern Matching

Comprehensive pattern matching implementation with regex support:

```python
import re

def _match_error_patterns(
    self, error: Exception, error_analysis: ErrorAnalysis
) -> Dict[str, Any]:
    """Match error against known patterns."""

    error_type = type(error).__name__
    error_message = str(error)

    if error_type in self.error_patterns:
        pattern_data = self.error_patterns[error_type]

        # Try to match regex patterns
        matched_patterns = []
        for pattern in pattern_data['patterns']:
            if re.search(pattern, error_message, re.IGNORECASE):
                matched_patterns.append(pattern)

        return {
            'matched_patterns': matched_patterns,
            'solutions': pattern_data['solutions'],
            'context7_topics': pattern_data['context7_topics']
        }

    return {'matched_patterns': [], 'solutions': [], 'context7_topics': []}
```

### Solution Generation

Multi-source solution generation with confidence scoring:

```python
async def _generate_solutions(
    self, error_analysis: ErrorAnalysis,
    context7_patterns: Dict, pattern_matches: Dict,
    context: Dict
) -> List[Solution]:
    """Generate comprehensive solutions using multiple sources."""

    solutions = []

    # Pattern-based solutions
    for pattern in pattern_matches.get('matched_patterns', []):
        solution = Solution(
            type='pattern_match',
            description=f"Apply known pattern: {pattern}",
            code_example=self._generate_pattern_example(pattern, context),
            confidence=0.85,
            impact='medium',
            dependencies=[]
        )
        solutions.append(solution)

    # Context7-based solutions
    for library_id, docs in context7_patterns.items():
        if docs and 'solutions' in docs:
            for sol in docs['solutions']:
                solution = Solution(
                    type='context7_pattern',
                    description=sol['description'],
                    code_example=sol.get('code_example', ''),
                    confidence=sol.get('confidence', 0.7),
                    impact=sol.get('impact', 'medium'),
                    dependencies=sol.get('dependencies', [])
                )
                solutions.append(solution)

    # AI-generated solutions
    if self.context7 and len(solutions) < 3:
        ai_solutions = await self._generate_ai_solutions(error_analysis, context)
        solutions.extend(ai_solutions)

    # Sort by confidence and impact
    solutions.sort(key=lambda x: (x.confidence, x.impact), reverse=True)
    return solutions[:5]
```

### Code Example Generation

Pattern-based code example generation:

```python
def _generate_pattern_example(self, pattern: str, context: Dict) -> str:
    """Generate code example for a specific error pattern."""

    examples = {
        r"No module named '(.+)'": """
# Install missing package
pip install package_name

# Or add to requirements.txt
echo "package_name" >> requirements.txt
""",
        r"'(.+)' object has no attribute '(.+)'": """
# Check object type before accessing attribute
if hasattr(obj, 'attribute_name'):
    result = obj.attribute_name
else:
    print(f"Object of type {type(obj)} doesn't have attribute 'attribute_name'")
""",
        r" takes \d+ positional arguments but \d+ were given": """
# Check function signature and call with correct arguments
def function_name(arg1, arg2, arg3=None):
    pass

# Correct call
function_name(value1, value2)
""",
        r"invalid literal for int\(\) with base 10": """
# Add error handling for type conversion
try:
    number = int(value)
except ValueError:
    print(f"Cannot convert '{value}' to integer")
    # Handle the error appropriately
""",
    }

    for pattern_key, example in examples.items():
        if pattern_key in pattern:
            return example

    return f"# Implement fix for pattern: {pattern}"
```

## Error Analysis Methods

### Severity Assessment

Comprehensive severity evaluation based on multiple factors:

```python
def _assess_severity(
    self, error: Exception, context: Dict, frequency: int
) -> str:
    """Assess error severity based on context and frequency."""

    # High severity indicators
    if any(keyword in str(error).lower() for keyword in [
        'critical', 'fatal', 'corruption', 'security'
    ]):
        return "critical"

    # Frequency-based severity
    if frequency > 10:
        return "high"
    elif frequency > 3:
        return "medium"

    # Context-based severity
    if context.get('production', False):
        return "high"
    elif context.get('user_facing', False):
        return "medium"

    return "low"
```

### Likely Causes Analysis

Root cause analysis for common error patterns:

```python
def _analyze_likely_causes(
    self, error_type: str, message: str, context: Dict
) -> List[str]:
    """Analyze likely causes of the error."""

    causes = []

    if error_type == "ImportError":
        if "No module named" in message:
            causes.extend([
                "Missing dependency installation",
                "Incorrect import path",
                "Virtual environment not activated"
            ])
        elif "circular import" in message:
            causes.extend([
                "Circular dependency between modules",
                "Improper module structure"
            ])

    elif error_type == "AttributeError":
        causes.extend([
            "Wrong object type being used",
            "Incorrect attribute name",
            "Object not properly initialized"
        ])

    elif error_type == "TypeError":
        causes.extend([
            "Incorrect data types in operation",
            "Function called with wrong argument types",
            "Missing type conversion"
        ])

    return causes
```

### Quick Fix Generation

Rapid fix suggestions for immediate resolution:

```python
def _generate_quick_fixes(
    self, classification: ErrorType, message: str, context: Dict
) -> List[str]:
    """Generate quick fixes for the error."""

    fixes = []

    if classification == ErrorType.IMPORT:
        fixes.extend([
            "Install missing package with pip",
            "Check Python path configuration",
            "Verify module exists in expected location"
        ])

    elif classification == ErrorType.ATTRIBUTE_ERROR:
        fixes.extend([
            "Add hasattr() check before attribute access",
            "Verify object initialization",
            "Check for typos in attribute name"
        ])

    elif classification == ErrorType.TYPE_ERROR:
        fixes.extend([
            "Add type conversion before operation",
            "Check function signature",
            "Use isinstance() for type validation"
        ])

    return fixes
```

## Prevention Strategies

### Type-Specific Prevention

Comprehensive prevention strategies by error type:

```python
def _suggest_prevention_strategies(
    self, error_analysis: ErrorAnalysis, context: Dict
) -> List[str]:
    """Suggest prevention strategies based on error analysis."""

    strategies = []

    # Type-specific prevention
    if error_analysis.type == ErrorType.IMPORT:
        strategies.extend([
            "Add proper dependency management with requirements.txt",
            "Implement module availability checks before imports",
            "Use virtual environments for dependency isolation"
        ])

    elif error_analysis.type == ErrorType.ATTRIBUTE_ERROR:
        strategies.extend([
            "Use hasattr() checks before attribute access",
            "Implement proper object type checking",
            "Add comprehensive unit tests for object interfaces"
        ])

    elif error_analysis.type == ErrorType.TYPE_ERROR:
        strategies.extend([
            "Add type hints and static type checking with mypy",
            "Implement runtime type validation",
            "Use isinstance() checks before operations"
        ])

    elif error_analysis.type == ErrorType.VALUE_ERROR:
        strategies.extend([
            "Add input validation at function boundaries",
            "Implement comprehensive error handling",
            "Use try-except blocks for data conversion"
        ])

    # General prevention strategies
    strategies.extend([
        "Implement comprehensive logging for error tracking",
        "Add automated testing to catch errors early",
        "Use code review process to prevent common issues"
    ])

    return strategies
```

### Related Error Detection

Identify related errors that frequently occur together:

```python
def _find_related_errors(self, error_analysis: ErrorAnalysis) -> List[str]:
    """Find related errors that might occur together."""

    related_map = {
        ErrorType.IMPORT: ["ModuleNotFoundError", "ImportError", "AttributeError"],
        ErrorType.ATTRIBUTE_ERROR: ["TypeError", "KeyError", "ImportError"],
        ErrorType.TYPE_ERROR: ["ValueError", "AttributeError", "TypeError"],
        ErrorType.VALUE_ERROR: ["TypeError", "KeyError", "IndexError"],
        ErrorType.KEY_ERROR: ["AttributeError", "TypeError", "IndexError"],
    }

    return related_map.get(error_analysis.type, ["TypeError", "ValueError"])
```

## Fix Time Estimation

### Time Estimation Algorithm

Predict fix time based on error type and solution confidence:

```python
def _estimate_fix_time(
    self, error_analysis: ErrorAnalysis, solutions: List[Solution]
) -> str:
    """Estimate time required to fix the error."""

    base_times = {
        ErrorType.SYNTAX: "1-5 minutes",
        ErrorType.IMPORT: "2-10 minutes",
        ErrorType.ATTRIBUTE_ERROR: "5-15 minutes",
        ErrorType.TYPE_ERROR: "5-20 minutes",
        ErrorType.VALUE_ERROR: "2-15 minutes",
        ErrorType.KEY_ERROR: "2-10 minutes",
        ErrorType.NETWORK: "10-30 minutes",
        ErrorType.DATABASE: "15-45 minutes",
        ErrorType.MEMORY: "20-60 minutes",
        ErrorType.CONCURRENCY: "30-90 minutes",
        ErrorType.UNKNOWN: "15-60 minutes"
    }

    base_time = base_times.get(error_analysis.type, "10-30 minutes")

    # Adjust based on solution confidence
    if solutions and solutions[0].confidence > 0.9:
        return f"Quick fix: {base_time}"
    elif solutions and solutions[0].confidence > 0.7:
        return f"Standard: {base_time}"
    else:
        return f"Complex: {base_time}"
```

## Statistics and Monitoring

### Debug Statistics

Comprehensive debugging session statistics:

```python
def get_debug_statistics(self) -> Dict[str, Any]:
    """Get debugging session statistics."""
    return {
        'total_errors_analyzed': len(self.error_history),
        'error_types': dict(Counter(key.split(':')[0] for key in self.error_history.keys())),
        'cache_hits': len(self.pattern_cache),
        'most_common_errors': sorted(
            self.error_history.items(),
            key=lambda x: x[1],
            reverse=True
        )[:5]
    }
```

### Error Frequency Tracking

Monitor error occurrence patterns:

```python
def get_error_frequency(self, error: Exception) -> int:
    """Get frequency of this error occurrence."""
    error_key = f"{type(error).__name__}:{str(error)[:50]}"
    return self.error_history.get(error_key, 0)
```

### Cache Management

Optimize Context7 query caching:

```python
def clear_error_history(self):
    """Clear error history for fresh analysis."""
    self.error_history.clear()
    self.pattern_cache.clear()
```

## Confidence Calculation

### Classification Confidence

Calculate confidence in error classification:

```python
def _calculate_confidence(
    self, classification: ErrorType, message: str
) -> float:
    """Calculate confidence in error classification."""

    # High confidence for direct type matches
    if classification != ErrorType.UNKNOWN:
        return 0.85

    # Lower confidence for unknown errors
    return 0.4
```

## Best Practices

Solution Prioritization: Apply solutions with highest confidence scores first and validate each fix in isolation before integration

Pattern Recognition: Track error patterns over time to identify systemic issues requiring architectural improvements

Prevention Strategy Implementation: Prioritize prevention strategies based on error frequency and severity impact

Learning Integration: Record successful fixes to improve pattern recognition and solution accuracy over time

Performance Optimization: Use caching for Context7 queries and implement batch processing for multiple errors

Documentation Updates: Maintain error pattern database with latest solutions and Context7 topics for continuous improvement

---

Module: modules/debugging/error-analysis.md
Version: 2.0.0 (Modular Architecture)
Last Updated: 2025-12-07
Lines: 350 (within 500-line limit)
