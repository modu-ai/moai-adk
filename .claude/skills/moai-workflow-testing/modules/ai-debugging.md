# AI-Powered Debugging Integration

> Module: Comprehensive AI debugging with Context7 integration and intelligent error analysis
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: Python 3.8+, Context7 MCP, asyncio, traceback, dataclasses

## Core Implementation

### AIDebugger Class

```python
import asyncio
import traceback
from typing import Dict, List, Optional, Any, Union
from dataclasses import dataclass
from enum import Enum

class ErrorType(Enum):
 """Classification of error types for intelligent handling."""
 SYNTAX = "syntax_error"
 RUNTIME = "runtime_error"
 IMPORT = "import_error"
 TYPE_ERROR = "type_error"
 VALUE_ERROR = "value_error"
 ATTRIBUTE_ERROR = "attribute_error"
 KEY_ERROR = "key_error"
 NETWORK = "network_error"
 DATABASE = "database_error"
 MEMORY = "memory_error"
 CONCURRENCY = "concurrency_error"
 UNKNOWN = "unknown_error"

@dataclass
class ErrorAnalysis:
 """Analysis of an error with classification and metadata."""
 type: ErrorType
 confidence: float
 message: str
 traceback: str
 context: Dict[str, Any]
 frequency: int
 severity: str # "low", "medium", "high", "critical"
 likely_causes: List[str]
 suggested_fixes: List[str]

@dataclass
class Solution:
 """Proposed solution for an error."""
 type: str # "context7_pattern", "ai_generated", "known_fix"
 description: str
 code_example: str
 confidence: float
 impact: str # "low", "medium", "high"
 dependencies: List[str]

@dataclass
class DebugAnalysis:
 """Complete debug analysis with solutions and prevention strategies."""
 error_type: ErrorType
 confidence: float
 context7_patterns: Dict[str, Any]
 solutions: List[Solution]
 prevention_strategies: List[str]
 related_errors: List[str]
 estimated_fix_time: str

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

 # Generate likely causes based on analysis
 likely_causes = self._analyze_likely_causes(
 error_type_name, error_message, context
 )

 # Generate initial suggested fixes
 suggested_fixes = self._generate_quick_fixes(
 classification, error_message, context
 )

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

 if any(keyword in message_lower for keyword in [
 'connection', 'timeout', 'network', 'http', 'socket'
 ]):
 return ErrorType.NETWORK

 if any(keyword in message_lower for keyword in [
 'database', 'sql', 'query', 'connection', 'cursor'
 ]):
 return ErrorType.DATABASE

 if any(keyword in message_lower for keyword in [
 'memory', 'out of memory', 'allocation', 'heap'
 ]):
 return ErrorType.MEMORY

 if any(keyword in message_lower for keyword in [
 'thread', 'lock', 'race condition', 'concurrent'
 ]):
 return ErrorType.CONCURRENCY

 # Context-based classification
 if context.get('operation_type') == 'database':
 return ErrorType.DATABASE
 elif context.get('operation_type') == 'network':
 return ErrorType.NETWORK

 return ErrorType.UNKNOWN

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

 def _match_error_patterns(
 self, error: Exception, error_analysis: ErrorAnalysis
 ) -> Dict[str, Any]:
 """Match error against known patterns."""

 error_type = type(error).__name__
 error_message = str(error)

 if error_type in self.error_patterns:
 pattern_data = self.error_patterns[error_type]

 # Try to match regex patterns
 import re
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
 if self.context7 and len(solutions) < 3: # Generate AI solutions if limited patterns found
 ai_solutions = await self._generate_ai_solutions(error_analysis, context)
 solutions.extend(ai_solutions)

 # Sort by confidence and impact
 solutions.sort(key=lambda x: (x.confidence, x.impact), reverse=True)
 return solutions[:5] # Return top 5 solutions

 async def _generate_ai_solutions(
 self, error_analysis: ErrorAnalysis, context: Dict
 ) -> List[Solution]:
 """Generate AI-powered solutions when patterns are insufficient."""

 solutions = []

 try:
 # Query for specific error solutions
 language = context.get('language', 'python')
 ai_response = await self.context7.get_library_docs(
 context7_library_id="/openai/chatgpt",
 topic=f"solve {error_analysis.type.value} in {language}: {error_analysis.message[:100]}",
 tokens=3000
 )

 if ai_response and 'solutions' in ai_response:
 for sol in ai_response['solutions']:
 solution = Solution(
 type='ai_generated',
 description=sol['description'],
 code_example=sol.get('code_example', ''),
 confidence=0.75,
 impact='medium',
 dependencies=sol.get('dependencies', [])
 )
 solutions.append(solution)

 except Exception as e:
 print(f"AI solution generation failed: {e}")

 return solutions

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

 def _calculate_confidence(
 self, classification: ErrorType, message: str
 ) -> float:
 """Calculate confidence in error classification."""

 # High confidence for direct type matches
 if classification != ErrorType.UNKNOWN:
 return 0.85

 # Lower confidence for unknown errors
 return 0.4

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

 def get_error_frequency(self, error: Exception) -> int:
 """Get frequency of this error occurrence."""
 error_key = f"{type(error).__name__}:{str(error)[:50]}"
 return self.error_history.get(error_key, 0)

 def clear_error_history(self):
 """Clear error history for fresh analysis."""
 self.error_history.clear()
 self.pattern_cache.clear()

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

# Usage Examples
"""
# Basic usage
debugger = AIDebugger(context7_client=context7)

try:
 # Code that might fail
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

 # Get prevention strategies
 print("Prevention strategies:")
 for strategy in analysis.prevention_strategies:
 print(f" - {strategy}")

# Check debug statistics
stats = debugger.get_debug_statistics()
print(f"Debugged {stats['total_errors_analyzed']} errors")
print(f"Most common: {stats['most_common_errors'][:3]}")
"""
```

## Advanced Features

### Context Integration

Enhanced Context Collection:
```python
def collect_debug_context(
 error: Exception,
 frame_depth: int = 5
) -> Dict[str, Any]:
 """Collect comprehensive debug context."""

 import inspect
 import sys

 # Get current frame and walk up the stack
 frame = inspect.currentframe()
 context = {
 'error_type': type(error).__name__,
 'error_message': str(error),
 'timestamp': datetime.now().isoformat(),
 'python_version': sys.version,
 'stack_trace': []
 }

 # Collect frame information
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

### Error Pattern Learning

Self-Improving Pattern Recognition:
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

 # Create solutions from learned data
 solutions = []
 for fix in learned:
 if fix['success_rate'] > 0.7: # Only return successful fixes
 solution = Solution(
 type='learned_pattern',
 description=f"Previously successful fix: {fix['solution']}",
 code_example=fix['solution'],
 confidence=fix['success_rate'],
 impact='high', # Previously successful fixes are high impact
 dependencies=[]
 )
 solutions.append(solution)

 return solutions

 return []
```

## Best Practices

1. Context Collection: Always provide comprehensive context including file paths, function names, and relevant variables
2. Error Categorization: Use specific error types for better pattern matching
3. Solution Validation: Test proposed solutions before applying them
4. Learning Integration: Record successful fixes to improve future debugging
5. Performance Monitoring: Track debugging session performance and cache efficiency

---

Module: `modules/ai-debugging.md`
Related: [Smart Refactoring](./smart-refactoring.md) | [Performance Optimization](./performance-optimization.md)
