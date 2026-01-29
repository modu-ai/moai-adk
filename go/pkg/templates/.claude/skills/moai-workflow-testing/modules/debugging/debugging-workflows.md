# Debugging Workflows and Implementation

> Module: AI debugging process patterns and implementation workflows
> Complexity: Advanced
> Dependencies: ai-debugging.md overview module

## Core Implementation

### AIDebugger Class Structure

Complete AIDebugger implementation with Context7 integration, pattern matching, and solution generation. For data class definitions (ErrorType, ErrorAnalysis, Solution, DebugAnalysis), see [ai-debugging.md](../ai-debugging.md).

```python
class AIDebugger:
    """AI-powered debugging with Context7 integration."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.error_patterns = self._load_error_patterns()
        self.error_history = {}
        self.pattern_cache = {}

    def _load_error_patterns(self) -> Dict[str, Any]:
        """Load comprehensive error patterns database."""
        return {
            'ImportError': {
                'patterns': [
                    r"No module named '(.+)'",
                    r"cannot import name '(.+)' from '(.+)'",
                    r"circular import"
                ],
                'solutions': [
                    'Install missing package',
                    'Check import path',
                    'Resolve circular dependencies'
                ],
                'context7_topics': [
                    'python import system best practices',
                    'module resolution troubleshooting',
                    'dependency management'
                ]
            },
            'AttributeError': {
                'patterns': [
                    r"'(.+)' object has no attribute '(.+)'",
                    r"module '(.+)' has no attribute '(.+)'"
                ],
                'solutions': [
                    'Check object type and available attributes',
                    'Verify module import',
                    'Add missing attribute or method'
                ],
                'context7_topics': [
                    'python attribute access patterns',
                    'object-oriented debugging',
                    'introspection techniques'
                ]
            },
            'TypeError': {
                'patterns': [
                    r" unsupported operand type\(s\) for",
                    r" takes \d+ positional arguments but \d+ were given",
                    r" must be str, not .+"
                ],
                'solutions': [
                    'Check data types before operations',
                    'Verify function signatures',
                    'Add type validation'
                ],
                'context7_topics': [
                    'python type system debugging',
                    'function signature validation',
                    'type checking best practices'
                ]
            },
            'ValueError': {
                'patterns': [
                    r"invalid literal for int\(\) with base 10",
                    r"cannot convert",
                    r"empty set"
                ],
                'solutions': [
                    'Validate input data format',
                    'Add error handling for conversions',
                    'Check value ranges'
                ],
                'context7_topics': [
                    'input validation patterns',
                    'data conversion error handling',
                    'value range checking'
                ]
            }
        }
```

### Main Debugging Method

Complete debug workflow implementation with error classification, pattern matching, solution generation, and prevention strategies:

```python
    async def debug_with_context7_patterns(
        self, error: Exception, context: Dict, codebase_path: str
    ) -> DebugAnalysis:
        """Debug using AI pattern recognition and Context7 best practices."""

        # Classify error with high accuracy
        error_analysis = await self._classify_error_with_ai(error, context)

        # Get Context7 patterns if available
        context7_patterns = {}
        if self.context7:
            context7_patterns = await self._get_context7_patterns(error_analysis)

        # Match against known patterns
        pattern_matches = self._match_error_patterns(error, error_analysis)

        # Generate comprehensive solutions
        solutions = await self._generate_solutions(
            error_analysis, context7_patterns, pattern_matches, context
        )

        # Suggest prevention strategies
        prevention = self._suggest_prevention_strategies(error_analysis, context)

        # Estimate fix time based on complexity
        fix_time = self._estimate_fix_time(error_analysis, solutions)

        return DebugAnalysis(
            error_type=error_analysis.type,
            confidence=error_analysis.confidence,
            context7_patterns=context7_patterns,
            solutions=solutions,
            prevention_strategies=prevention,
            related_errors=self._find_related_errors(error_analysis),
            estimated_fix_time=fix_time
        )
```

### Error Classification System

AI-enhanced error classification with context awareness using type mapping, message patterns, and contextual analysis:

```python
    async def _classify_error_with_ai(
        self, error: Exception, context: Dict
    ) -> ErrorAnalysis:
        """Classify error using AI-enhanced pattern recognition."""

        error_type_name = type(error).__name__
        error_message = str(error)
        error_traceback = traceback.format_exc()

        # Enhanced classification with context awareness
        classification = self._classify_by_type_and_message(
            error_type_name, error_message, context
        )

        # Analyze frequency and severity
        error_key = f"{error_type_name}:{error_message[:50]}"
        frequency = self.error_history.get(error_key, 0) + 1
        self.error_history[error_key] = frequency

        severity = self._assess_severity(error, context, frequency)

        # Generate likely causes and suggested fixes
        likely_causes = self._analyze_likely_causes(error_type_name, error_message, context)
        suggested_fixes = self._generate_quick_fixes(classification, error_message, context)

        return ErrorAnalysis(
            type=classification,
            confidence=self._calculate_confidence(classification, error_message),
            message=error_message,
            traceback=error_traceback,
            context=context,
            frequency=frequency,
            severity=severity,
            likely_causes=likely_causes,
            suggested_fixes=suggested_fixes
        )

    def _classify_by_type_and_message(
        self, error_type: str, message: str, context: Dict
    ) -> ErrorType:
        """Enhanced error classification using multiple heuristics."""

        # Direct type mapping
        type_mapping = {
            'ImportError': ErrorType.IMPORT,
            'ModuleNotFoundError': ErrorType.IMPORT,
            'AttributeError': ErrorType.ATTRIBUTE_ERROR,
            'KeyError': ErrorType.KEY_ERROR,
            'TypeError': ErrorType.TYPE_ERROR,
            'ValueError': ErrorType.VALUE_ERROR,
            'SyntaxError': ErrorType.SYNTAX,
            'IndentationError': ErrorType.SYNTAX,
        }

        if error_type in type_mapping:
            return type_mapping[error_type]

        # Message-based classification
        message_lower = message.lower()

        network_keywords = ['connection', 'timeout', 'network', 'http', 'socket']
        database_keywords = ['database', 'sql', 'query', 'connection', 'cursor']
        memory_keywords = ['memory', 'out of memory', 'allocation', 'heap']
        concurrency_keywords = ['thread', 'lock', 'race condition', 'concurrent']

        if any(keyword in message_lower for keyword in network_keywords):
            return ErrorType.NETWORK
        if any(keyword in message_lower for keyword in database_keywords):
            return ErrorType.DATABASE
        if any(keyword in message_lower for keyword in memory_keywords):
            return ErrorType.MEMORY
        if any(keyword in message_lower for keyword in concurrency_keywords):
            return ErrorType.CONCURRENCY

        # Context-based classification
        if context.get('operation_type') == 'database':
            return ErrorType.DATABASE
        elif context.get('operation_type') == 'network':
            return ErrorType.NETWORK

        return ErrorType.UNKNOWN
```

### Context7 Integration

Automatic documentation retrieval for latest debugging patterns and best practices with intelligent caching:

```python
    async def _get_context7_patterns(
        self, error_analysis: ErrorAnalysis
    ) -> Dict[str, Any]:
        """Get latest debugging patterns from Context7."""

        cache_key = f"{error_analysis.type.value}_{error_analysis.message[:30]}"
        if cache_key in self.pattern_cache:
            return self.pattern_cache[cache_key]

        # Determine appropriate Context7 libraries based on error type
        context7_queries = self._build_context7_queries(error_analysis)

        patterns = {}
        if self.context7:
            for library_id, topic in context7_queries:
                try:
                    result = await self.context7.get_library_docs(
                        context7_library_id=library_id,
                        topic=topic,
                        tokens=4000
                    )
                    patterns[library_id] = result
                except Exception as e:
                    print(f"Context7 query failed for {library_id}: {e}")

        # Cache results
        self.pattern_cache[cache_key] = patterns
        return patterns

    def _build_context7_queries(self, error_analysis: ErrorAnalysis) -> List[tuple]:
        """Build Context7 queries based on error analysis."""

        queries = []

        # Base debugging library
        queries.append(("/microsoft/debugpy",
                        f"AI debugging patterns {error_analysis.type.value} error analysis 2025"))

        # Language-specific libraries
        if error_analysis.context.get('language') == 'python':
            queries.append(("/python/cpython",
                            f"{error_analysis.type.value} debugging best practices"))

        # Framework-specific queries
        framework = error_analysis.context.get('framework')
        if framework:
            queries.append((f"/{framework}/{framework}",
                            f"{framework} {error_analysis.type.value} troubleshooting"))

        return queries
```

## Advanced Implementation Patterns

### Learning Debugger Extension

Self-improving debugger that learns from successful fixes with pattern recognition and success rate tracking:

```python
class LearningDebugger(AIDebugger):
    """Debugger that learns from fixed errors."""

    def __init__(self, context7_client=None):
        super().__init__(context7_client)
        self.learned_patterns = {}
        self.successful_fixes = {}

    def record_successful_fix(
        self, error_signature: str, applied_solution: str
    ):
        """Record successful fix for future reference."""
        if error_signature not in self.successful_fixes:
            self.successful_fixes[error_signature] = []

        self.successful_fixes[error_signature].append({
            'solution': applied_solution,
            'timestamp': datetime.now().isoformat(),
            'success_rate': 1.0
        })

    def get_learned_solutions(self, error_signature: str) -> List[Solution]:
        """Get solutions learned from previous fixes."""
        if error_signature in self.successful_fixes:
            learned = self.successful_fixes[error_signature]
            solutions = []
            for fix in learned:
                if fix['success_rate'] > 0.7:
                    solution = Solution(
                        type='learned_pattern',
                        description=f"Previously successful fix: {fix['solution']}",
                        code_example=fix['solution'],
                        confidence=fix['success_rate'],
                        impact='high',
                        dependencies=[]
                    )
                    solutions.append(solution)
            return solutions
        return []
```

### Enhanced Context Collection

Comprehensive debug context extraction with stack frame analysis for improved error classification:

```python
def collect_debug_context(
    error: Exception,
    frame_depth: int = 5
) -> Dict[str, Any]:
    """Collect comprehensive debug context."""
    import inspect
    import sys
    from datetime import datetime

    frame = inspect.currentframe()
    context = {
        'error_type': type(error).__name__,
        'error_message': str(error),
        'timestamp': datetime.now().isoformat(),
        'python_version': sys.version,
        'stack_trace': []
    }

    for _ in range(frame_depth):
        if frame:
            frame_info = {
                'filename': frame.f_code.co_filename,
                'function': frame.f_code.co_name,
                'lineno': frame.f_lineno,
                'locals': list(frame.f_locals.keys())
            }
            context['stack_trace'].append(frame_info)
            frame = frame.f_back

    return context
```

### Usage Examples

Complete usage examples for common debugging scenarios:

```python
# Basic usage
debugger = AIDebugger(context7_client=context7)

try:
    result = some_risky_operation()
except Exception as e:
    analysis = await debugger.debug_with_context7_patterns(
        e,
        {'file': __file__, 'function': 'some_risky_operation', 'language': 'python'},
        '/project/src'
    )

    print(f"Error type: {analysis.error_type}")
    print(f"Confidence: {analysis.confidence}")
    print(f"Solutions found: {len(analysis.solutions)}")

    for i, solution in enumerate(analysis.solutions, 1):
        print(f"\nSolution {i}:")
        print(f" Description: {solution.description}")
        print(f" Confidence: {solution.confidence}")
        print(f" Impact: {solution.impact}")
        if solution.code_example:
            print(f" Example:\n{solution.code_example}")

# Advanced usage with custom context
try:
    data = process_user_input(user_data)
except Exception as e:
    analysis = await debugger.debug_with_context7_patterns(
        e,
        {
            'file': __file__,
            'function': 'process_user_input',
            'language': 'python',
            'framework': 'django',
            'operation_type': 'data_processing',
            'user_facing': True,
            'production': False
        },
        '/project/src'
    )

    print("Prevention strategies:")
    for strategy in analysis.prevention_strategies:
        print(f" - {strategy}")

# Check debug statistics
stats = debugger.get_debug_statistics()
print(f"Debugged {stats['total_errors_analyzed']} errors")
print(f"Most common: {stats['most_common_errors'][:3]}")
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
