# Automated Code Review with TRUST 5 Validation

> Module: AI-powered automated code review with TRUST 5 validation framework and comprehensive quality analysis
> Complexity: Advanced
> Time: 35+ minutes
> Dependencies: Python 3.8+, Context7 MCP, ast, pylint, flake8, bandit, mypy

## Core Implementation

### AutomatedCodeReviewer Class

```python
import ast
import subprocess
import json
import asyncio
from typing import Dict, List, Optional, Any, Union, Callable, Set
from dataclasses import dataclass, field
from enum import Enum
import re
import os
import sys
from pathlib import Path
import tempfile
import difflib

class TrustCategory(Enum):
 """TRUST 5 framework categories."""
 TRUTHFULNESS = "truthfulness" # Code correctness and logic accuracy
 RELEVANCE = "relevance" # Code meets requirements and purpose
 USABILITY = "usability" # Code is maintainable and understandable
 SAFETY = "safety" # Code is secure and handles errors properly
 TIMELINESS = "timeliness" # Code meets performance and delivery standards

class Severity(Enum):
 """Issue severity levels."""
 CRITICAL = "critical"
 HIGH = "high"
 MEDIUM = "medium"
 LOW = "low"
 INFO = "info"

class IssueType(Enum):
 """Types of code issues."""
 SYNTAX_ERROR = "syntax_error"
 LOGIC_ERROR = "logic_error"
 SECURITY_VULNERABILITY = "security_vulnerability"
 PERFORMANCE_ISSUE = "performance_issue"
 CODE_SMELL = "code_smell"
 STYLE_VIOLATION = "style_violation"
 DOCUMENTATION_ISSUE = "documentation_issue"
 TESTING_ISSUE = "testing_issue"
 TYPE_ERROR = "type_error"
 IMPORT_ISSUE = "import_issue"

@dataclass
class CodeIssue:
 """Individual code issue found during review."""
 id: str
 category: TrustCategory
 severity: Severity
 issue_type: IssueType
 title: str
 description: str
 file_path: str
 line_number: int
 column_number: int
 code_snippet: str
 suggested_fix: str
 confidence: float # 0.0 to 1.0
 rule_violated: Optional[str] = None
 external_reference: Optional[str] = None
 auto_fixable: bool = False
 fix_diff: Optional[str] = None

@dataclass
class FileReviewResult:
 """Review results for a single file."""
 file_path: str
 issues: List[CodeIssue]
 metrics: Dict[str, Any]
 trust_score: float # 0.0 to 1.0
 category_scores: Dict[TrustCategory, float]
 lines_of_code: int
 complexity_metrics: Dict[str, float]
 review_timestamp: float

@dataclass
class CodeReviewReport:
 """Comprehensive code review report."""
 project_path: str
 files_reviewed: List[FileReviewResult]
 overall_trust_score: float
 overall_category_scores: Dict[TrustCategory, float]
 summary_metrics: Dict[str, Any]
 recommendations: List[str]
 critical_issues: List[CodeIssue]
 review_duration: float
 context7_patterns_used: List[str]

class Context7CodeAnalyzer:
 """Integration with Context7 for code analysis patterns."""

 def __init__(self, context7_client=None):
 self.context7 = context7_client
 self.analysis_patterns = {}
 self.security_patterns = {}
 self.performance_patterns = {}

 async def load_analysis_patterns(self, language: str = "python") -> Dict[str, Any]:
 """Load code analysis patterns from Context7."""

 if not self.context7:
 return self._get_default_analysis_patterns()

 try:
 # Load security analysis patterns
 security_patterns = await self.context7.get_library_docs(
 context7_library_id="/security/semgrep",
 topic="security vulnerability detection patterns 2025",
 tokens=4000
 )
 self.security_patterns = security_patterns

 # Load performance analysis patterns
 performance_patterns = await self.context7.get_library_docs(
 context7_library_id="/performance/python-profiling",
 topic="performance anti-patterns code analysis 2025",
 tokens=3000
 )
 self.performance_patterns = performance_patterns

 # Load code quality patterns
 quality_patterns = await self.context7.get_library_docs(
 context7_library_id="/code-quality/sonarqube",
 topic="code quality best practices smells detection 2025",
 tokens=4000
 )

 # Load TRUST 5 validation patterns
 trust_patterns = await self.context7.get_library_docs(
 context7_library_id="/code-review/trust-framework",
 topic="TRUST 5 code validation framework patterns 2025",
 tokens=3000
 )

 return {
 'security': security_patterns,
 'performance': performance_patterns,
 'quality': quality_patterns,
 'trust': trust_patterns
 }

 except Exception as e:
 print(f"Failed to load Context7 patterns: {e}")
 return self._get_default_analysis_patterns()

 def _get_default_analysis_patterns(self) -> Dict[str, Any]:
 """Get default analysis patterns when Context7 is unavailable."""
 return {
 'security': {
 'sql_injection': [
 r"execute\([^)]*\+[^)]*\)",
 r"format\s*\(",
 r"%\s*[^,]*s"
 ],
 'command_injection': [
 r"os\.system\(",
 r"subprocess\.call\(",
 r"eval\("
 ],
 'path_traversal': [
 r"open\([^)]*\+[^)]*\)",
 r"\.\.\/"
 ]
 },
 'performance': {
 'inefficient_loops': [
 r"for.*in.*range\(len\(",
 r"while.*len\("
 ],
 'memory_leaks': [
 r"global\s+",
 r"\.append\(.*\)\s*\.append\("
 ]
 },
 'quality': {
 'long_functions': {'max_lines': 50},
 'complex_conditionals': {'max_complexity': 10},
 'deep_nesting': {'max_depth': 4}
 }
 }

class StaticAnalysisTools:
 """Wrapper for various static analysis tools."""

 def __init__(self):
 self.tools = {
 'pylint': self._run_pylint,
 'flake8': self._run_flake8,
 'bandit': self._run_bandit,
 'mypy': self._run_mypy
 }

 async def run_all_analyses(self, file_path: str) -> Dict[str, Any]:
 """Run all available static analysis tools."""

 results = {}

 for tool_name, tool_func in self.tools.items():
 try:
 result = await tool_func(file_path)
 results[tool_name] = result
 except Exception as e:
 print(f"Error running {tool_name}: {e}")
 results[tool_name] = {'error': str(e)}

 return results

 async def _run_pylint(self, file_path: str) -> Dict[str, Any]:
 """Run pylint analysis."""
 try:
 result = subprocess.run(
 ['pylint', file_path, '--output-format=json'],
 capture_output=True,
 text=True
 )

 if result.returncode == 0:
 return {'issues': []}

 try:
 issues = json.loads(result.stdout)
 return {'issues': issues, 'summary': self._parse_pylint_summary(result.stderr)}
 except json.JSONDecodeError:
 return {'raw_output': result.stdout, 'raw_errors': result.stderr}

 except FileNotFoundError:
 return {'error': 'pylint not installed'}

 async def _run_flake8(self, file_path: str) -> Dict[str, Any]:
 """Run flake8 analysis."""
 try:
 result = subprocess.run(
 ['flake8', file_path, '--format=json'],
 capture_output=True,
 text=True
 )

 if result.returncode == 0:
 return {'issues': []}

 # Parse flake8 output
 issues = []
 for line in result.stdout.split('\n'):
 if line.strip():
 parts = line.split(':')
 if len(parts) >= 4:
 issues.append({
 'path': parts[0],
 'line': int(parts[1]),
 'column': int(parts[2]),
 'code': parts[3].strip(),
 'message': ':'.join(parts[4:]).strip()
 })

 return {'issues': issues}

 except FileNotFoundError:
 return {'error': 'flake8 not installed'}

 async def _run_bandit(self, file_path: str) -> Dict[str, Any]:
 """Run bandit security analysis."""
 try:
 result = subprocess.run(
 ['bandit', '-f', 'json', file_path],
 capture_output=True,
 text=True
 )

 try:
 bandit_results = json.loads(result.stdout)
 return bandit_results
 except json.JSONDecodeError:
 return {'raw_output': result.stdout}

 except FileNotFoundError:
 return {'error': 'bandit not installed'}

 async def _run_mypy(self, file_path: str) -> Dict[str, Any]:
 """Run mypy type analysis."""
 try:
 result = subprocess.run(
 ['mypy', file_path, '--show-error-codes'],
 capture_output=True,
 text=True
 )

 # Parse mypy output
 issues = []
 for line in result.stdout.split('\n'):
 if ':' in line and 'error:' in line:
 parts = line.split(':', 3)
 if len(parts) >= 4:
 issues.append({
 'path': parts[0],
 'line': int(parts[1]),
 'message': parts[3].strip()
 })

 return {'issues': issues}

 except FileNotFoundError:
 return {'error': 'mypy not installed'}

 def _parse_pylint_summary(self, stderr: str) -> Dict[str, Any]:
 """Parse pylint summary from stderr."""
 summary = {}
 for line in stderr.split('\n'):
 if 'rated at' in line:
 # Extract rating
 match = re.search(r'rated at ([\d.]+)/10', line)
 if match:
 summary['rating'] = float(match.group(1))

 elif any(keyword in line for keyword in ['statements', 'lines', 'functions', 'classes']):
 parts = line.split()
 if len(parts) >= 2:
 summary[parts[0]] = parts[1]

 return summary

class AutomatedCodeReviewer:
 """Main automated code reviewer with TRUST 5 validation."""

 def __init__(self, context7_client=None):
 self.context7 = context7_client
 self.context7_analyzer = Context7CodeAnalyzer(context7_client)
 self.static_analyzer = StaticAnalysisTools()
 self.analysis_patterns = {}
 self.review_history = []

 async def review_codebase(
 self, project_path: str,
 include_patterns: List[str] = None,
 exclude_patterns: List[str] = None
 ) -> CodeReviewReport:
 """Perform comprehensive code review of entire codebase."""

 start_time = time.time()

 # Load analysis patterns
 self.analysis_patterns = await self.context7_analyzer.load_analysis_patterns()

 # Find files to review
 files_to_review = self._find_files_to_review(
 project_path, include_patterns, exclude_patterns
 )

 print(f"Found {len(files_to_review)} files to review")

 # Review each file
 file_results = []
 for file_path in files_to_review:
 print(f"Reviewing {file_path}...")
 file_result = await self.review_single_file(file_path)
 file_results.append(file_result)

 # Generate comprehensive report
 end_time = time.time()
 report = self._generate_comprehensive_report(
 project_path, file_results, end_time - start_time
 )

 return report

 async def review_single_file(self, file_path: str) -> FileReviewResult:
 """Review a single Python file."""

 # Read file content
 try:
 with open(file_path, 'r', encoding='utf-8') as f:
 content = f.read()
 except Exception as e:
 print(f"Error reading {file_path}: {e}")
 return self._create_error_result(file_path, str(e))

 # Parse AST
 try:
 tree = ast.parse(content)
 except SyntaxError as e:
 return self._create_syntax_error_result(file_path, content, e)

 # Run static analyses
 static_results = await self.static_analyzer.run_all_analyses(file_path)

 # Perform Context7-enhanced analysis
 context7_issues = await self._perform_context7_analysis(file_path, content, tree)

 # Perform custom analysis
 custom_issues = await self._perform_custom_analysis(file_path, content, tree)

 # Combine all issues
 all_issues = []
 all_issues.extend(self._convert_static_issues(static_results, file_path))
 all_issues.extend(context7_issues)
 all_issues.extend(custom_issues)

 # Calculate metrics and scores
 metrics = self._calculate_file_metrics(content, tree)
 trust_scores = self._calculate_trust_scores(all_issues, metrics)

 return FileReviewResult(
 file_path=file_path,
 issues=all_issues,
 metrics=metrics,
 trust_score=trust_scores['overall'],
 category_scores=trust_scores['categories'],
 lines_of_code=len(content.split('\n')),
 complexity_metrics=self._calculate_complexity_metrics(content, tree),
 review_timestamp=time.time()
 )

 def _find_files_to_review(
 self, project_path: str,
 include_patterns: List[str] = None,
 exclude_patterns: List[str] = None
 ) -> List[str]:
 """Find Python files to review."""

 if include_patterns is None:
 include_patterns = ['/*.py']

 if exclude_patterns is None:
 exclude_patterns = [
 '/__pycache__/',
 '/venv/',
 '/env/',
 '/node_modules/',
 '/.git/',
 '/migrations/',
 '/tests/'
 ]

 import fnmatch
 from pathlib import Path

 project_root = Path(project_path)
 files = []

 for pattern in include_patterns:
 for file_path in project_root.glob(pattern):
 if file_path.is_file():
 # Check exclude patterns
 excluded = False
 for exclude_pattern in exclude_patterns:
 if fnmatch.fnmatch(str(file_path.relative_to(project_root)), exclude_pattern):
 excluded = True
 break

 if not excluded:
 files.append(str(file_path))

 return sorted(files)

 async def _perform_context7_analysis(
 self, file_path: str, content: str, tree: ast.AST
 ) -> List[CodeIssue]:
 """Perform Context7-enhanced code analysis."""

 issues = []

 # Security analysis
 security_issues = await self._analyze_security_patterns(file_path, content)
 issues.extend(security_issues)

 # Performance analysis
 performance_issues = await self._analyze_performance_patterns(file_path, content)
 issues.extend(performance_issues)

 # Code quality analysis
 quality_issues = await self._analyze_quality_patterns(file_path, tree)
 issues.extend(quality_issues)

 # TRUST 5 analysis
 trust_issues = await self._analyze_trust_patterns(file_path, content, tree)
 issues.extend(trust_issues)

 return issues

 async def _analyze_security_patterns(self, file_path: str, content: str) -> List[CodeIssue]:
 """Analyze security patterns using Context7."""

 issues = []
 security_patterns = self.analysis_patterns.get('security', {})
 lines = content.split('\n')

 for category, patterns in security_patterns.items():
 if isinstance(patterns, list):
 for pattern in patterns:
 try:
 regex = re.compile(pattern, re.IGNORECASE)
 for line_num, line in enumerate(lines, 1):
 if regex.search(line):
 issue = CodeIssue(
 id=f"security_{category}_{line_num}_{len(issues)}",
 category=TrustCategory.SAFETY,
 severity=Severity.HIGH,
 issue_type=IssueType.SECURITY_VULNERABILITY,
 title=f"Security Issue: {category.replace('_', ' ').title()}",
 description=f"Potential {category} vulnerability detected",
 file_path=file_path,
 line_number=line_num,
 column_number=1,
 code_snippet=line.strip(),
 suggested_fix=self._get_security_fix_suggestion(category, line),
 confidence=0.7,
 rule_violated=f"SECURITY_{category.upper()}",
 external_reference=self._get_security_reference(category)
 )
 issues.append(issue)
 except re.error as e:
 print(f"Invalid security pattern {pattern}: {e}")

 return issues

 async def _analyze_performance_patterns(self, file_path: str, content: str) -> List[CodeIssue]:
 """Analyze performance patterns using Context7."""

 issues = []
 performance_patterns = self.analysis_patterns.get('performance', {})
 lines = content.split('\n')

 for category, patterns in performance_patterns.items():
 if isinstance(patterns, list):
 for pattern in patterns:
 try:
 regex = re.compile(pattern)
 for line_num, line in enumerate(lines, 1):
 if regex.search(line):
 issue = CodeIssue(
 id=f"perf_{category}_{line_num}_{len(issues)}",
 category=TrustCategory.TIMELINESS,
 severity=Severity.MEDIUM,
 issue_type=IssueType.PERFORMANCE_ISSUE,
 title=f"Performance Issue: {category.replace('_', ' ').title()}",
 description=f"Performance anti-pattern detected: {category}",
 file_path=file_path,
 line_number=line_num,
 column_number=1,
 code_snippet=line.strip(),
 suggested_fix=self._get_performance_fix_suggestion(category, line),
 confidence=0.6,
 rule_violated=f"PERF_{category.upper()}"
 )
 issues.append(issue)
 except re.error as e:
 print(f"Invalid performance pattern {pattern}: {e}")

 return issues

 async def _analyze_quality_patterns(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
 """Analyze code quality patterns."""

 issues = []
 quality_patterns = self.analysis_patterns.get('quality', {})

 # Analyze function length
 if 'long_functions' in quality_patterns:
 max_lines = quality_patterns['long_functions'].get('max_lines', 50)
 function_issues = self._analyze_function_length(file_path, tree, max_lines)
 issues.extend(function_issues)

 # Analyze complexity
 if 'complex_conditionals' in quality_patterns:
 max_complexity = quality_patterns['complex_conditionals'].get('max_complexity', 10)
 complexity_issues = self._analyze_complexity(file_path, tree, max_complexity)
 issues.extend(complexity_issues)

 # Analyze nesting depth
 if 'deep_nesting' in quality_patterns:
 max_depth = quality_patterns['deep_nesting'].get('max_depth', 4)
 nesting_issues = self._analyze_nesting_depth(file_path, tree, max_depth)
 issues.extend(nesting_issues)

 return issues

 async def _analyze_trust_patterns(
 self, file_path: str, content: str, tree: ast.AST
 ) -> List[CodeIssue]:
 """Analyze TRUST 5 patterns."""

 issues = []

 # Truthfulness: Logic correctness
 truthfulness_issues = self._analyze_truthfulness(file_path, tree)
 issues.extend(truthfulness_issues)

 # Relevance: Requirements fulfillment
 relevance_issues = self._analyze_relevance(file_path, content)
 issues.extend(relevance_issues)

 # Usability: Maintainability
 usability_issues = self._analyze_usability(file_path, content, tree)
 issues.extend(usability_issues)

 # Safety: Error handling
 safety_issues = self._analyze_safety(file_path, tree)
 issues.extend(safety_issues)

 # Timeliness: Performance and standards
 timeliness_issues = self._analyze_timeliness(file_path, content)
 issues.extend(timeliness_issues)

 return issues

 def _analyze_function_length(
 self, file_path: str, tree: ast.AST, max_lines: int
 ) -> List[CodeIssue]:
 """Analyze function length violations."""

 issues = []
 lines = None

 for node in ast.walk(tree):
 if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
 if lines is None:
 with open(file_path, 'r') as f:
 lines = f.readlines()

 # Calculate function length (excluding docstring)
 start_line = node.lineno - 1
 end_line = node.end_lineno - 1 if node.end_lineno else start_line
 func_lines = lines[start_line:end_line + 1]

 # Remove docstring and blank lines
 code_lines = []
 in_docstring = False
 for line in func_lines:
 stripped = line.strip()
 if not in_docstring and ('"""' in line or "'''" in line):
 in_docstring = True
 continue
 if in_docstring and ('"""' in line or "'''" in line):
 in_docstring = False
 continue
 if not in_docstring and stripped and not stripped.startswith('#'):
 code_lines.append(line)

 if len(code_lines) > max_lines:
 issue = CodeIssue(
 id=f"func_length_{node.lineno}",
 category=TrustCategory.USABILITY,
 severity=Severity.MEDIUM,
 issue_type=IssueType.CODE_SMELL,
 title="Long Function",
 description=f"Function '{node.name}' is {len(code_lines)} lines long (max: {max_lines})",
 file_path=file_path,
 line_number=node.lineno,
 column_number=1,
 code_snippet=f"def {node.name}(...): # {len(code_lines)} lines",
 suggested_fix=f"Consider breaking '{node.name}' into smaller functions",
 confidence=0.8,
 rule_violated="FUNC_LENGTH"
 )
 issues.append(issue)

 return issues

 def _analyze_complexity(
 self, file_path: str, tree: ast.AST, max_complexity: int
 ) -> List[CodeIssue]:
 """Analyze cyclomatic complexity."""

 issues = []

 for node in ast.walk(tree):
 if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
 complexity = self._calculate_cyclomatic_complexity(node)

 if complexity > max_complexity:
 issue = CodeIssue(
 id=f"complexity_{node.lineno}",
 category=TrustCategory.USABILITY,
 severity=Severity.HIGH if complexity > max_complexity * 1.5 else Severity.MEDIUM,
 issue_type=IssueType.CODE_SMELL,
 title="High Complexity",
 description=f"Function '{node.name}' has cyclomatic complexity {complexity} (max: {max_complexity})",
 file_path=file_path,
 line_number=node.lineno,
 column_number=1,
 code_snippet=f"def {node.name}(...): # complexity: {complexity}",
 suggested_fix=f"Consider refactoring '{node.name}' to reduce complexity",
 confidence=0.9,
 rule_violated="COMPLEXITY"
 )
 issues.append(issue)

 return issues

 def _calculate_cyclomatic_complexity(self, node: ast.AST) -> int:
 """Calculate cyclomatic complexity for an AST node."""

 complexity = 1 # Base complexity

 for child in ast.walk(node):
 if isinstance(child, (ast.If, ast.While, ast.For, ast.AsyncFor,
 ast.ExceptHandler, ast.With, ast.AsyncWith)):
 complexity += 1
 elif isinstance(child, ast.BoolOp):
 complexity += len(child.values) - 1

 return complexity

 def _analyze_nesting_depth(
 self, file_path: str, tree: ast.AST, max_depth: int
 ) -> List[CodeIssue]:
 """Analyze nesting depth."""

 issues = []

 for node in ast.walk(tree):
 if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
 max_func_depth = self._calculate_nesting_depth(node)

 if max_func_depth > max_depth:
 issue = CodeIssue(
 id=f"nesting_{node.lineno}",
 category=TrustCategory.USABILITY,
 severity=Severity.MEDIUM,
 issue_type=IssueType.CODE_SMELL,
 title="Deep Nesting",
 description=f"Function '{node.name}' has nesting depth {max_func_depth} (max: {max_depth})",
 file_path=file_path,
 line_number=node.lineno,
 column_number=1,
 code_snippet=f"def {node.name}(...): # nesting depth: {max_func_depth}",
 suggested_fix=f"Consider using early returns or extracting functions in '{node.name}'",
 confidence=0.8,
 rule_violated="NESTING_DEPTH"
 )
 issues.append(issue)

 return issues

 def _calculate_nesting_depth(self, node: ast.AST, current_depth: int = 0) -> int:
 """Calculate maximum nesting depth for an AST node."""

 max_depth = current_depth

 for child in ast.walk(node):
 if isinstance(child, (ast.If, ast.While, ast.For, ast.AsyncFor,
 ast.With, ast.AsyncWith, ast.Try)):
 if hasattr(child, 'lineno') and hasattr(node, 'lineno') and child.lineno > node.lineno:
 child_depth = self._calculate_nesting_depth(child, current_depth + 1)
 max_depth = max(max_depth, child_depth)

 return max_depth

 def _analyze_truthfulness(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
 """Analyze code for correctness and logic issues."""

 issues = []

 # Check for unreachable code
 for node in ast.walk(tree):
 if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
 unreachable_issues = self._check_unreachable_code(file_path, node)
 issues.extend(unreachable_issues)

 # Check for logic issues
 logic_issues = self._check_logic_issues(file_path, tree)
 issues.extend(logic_issues)

 return issues

 def _check_unreachable_code(self, file_path: str, func_node: ast.AST) -> List[CodeIssue]:
 """Check for unreachable code after return statements."""

 issues = []

 class UnreachableCodeVisitor(ast.NodeVisitor):
 def __init__(self):
 self.found_return = False
 self.issues = []

 def visit_Return(self, node):
 self.found_return = True
 self.generic_visit(node)

 def generic_visit(self, node):
 if self.found_return and hasattr(node, 'lineno'):
 if isinstance(node, (ast.Expr, ast.Assign, ast.AugAssign)):
 issue = CodeIssue(
 id=f"unreachable_{node.lineno}",
 category=TrustCategory.TRUTHFULNESS,
 severity=Severity.LOW,
 issue_type=IssueType.CODE_SMELL,
 title="Unreachable Code",
 description="Code after return statement is never executed",
 file_path=file_path,
 line_number=node.lineno,
 column_number=1,
 code_snippet=f"# Unreachable code at line {node.lineno}",
 suggested_fix="Remove unreachable code or move before return statement",
 confidence=0.7,
 rule_violated="UNREACHABLE_CODE"
 )
 self.issues.append(issue)

 super().generic_visit(node)

 visitor = UnreachableCodeVisitor()
 visitor.visit(func_node)

 return visitor.issues

 def _check_logic_issues(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
 """Check for common logic issues."""

 issues = []

 # Check for comparison issues
 for node in ast.walk(tree):
 if isinstance(node, ast.Compare):
 comparison_issues = self._check_comparison_issues(file_path, node)
 issues.extend(comparison_issues)

 return issues

 def _check_comparison_issues(self, file_path: str, compare_node: ast.Compare) -> List[CodeIssue]:
 """Check for comparison logic issues."""

 issues = []

 # Check for None comparison
 for op in compare_node.ops:
 if isinstance(op, ast.Eq) or isinstance(op, ast.NotEq):
 for comparator in compare_node.comparators:
 if isinstance(comparator, ast.Constant) and comparator.value is None:
 issue = CodeIssue(
 id=f"none_comparison_{compare_node.lineno}",
 category=TrustCategory.TRUTHFULNESS,
 severity=Severity.LOW,
 issue_type=IssueType.CODE_SMELL,
 title="None Comparison",
 description="Use 'is' or 'is not' for None comparison",
 file_path=file_path,
 line_number=compare_node.lineno,
 column_number=1,
 code_snippet="# Use 'is None' instead of '== None'",
 suggested_fix="Replace '== None' with 'is None' and '!= None' with 'is not None'",
 confidence=0.8,
 rule_violated="NONE_COMPARISON",
 auto_fixable=True
 )
 issues.append(issue)

 return issues

 def _analyze_relevance(self, file_path: str, content: str) -> List[CodeIssue]:
 """Analyze code for relevance and requirements fulfillment."""

 issues = []

 # Check for TODO/FIXME comments
 lines = content.split('\n')
 for line_num, line in enumerate(lines, 1):
 if 'TODO:' in line or 'FIXME:' in line:
 issue = CodeIssue(
 id=f"todo_{line_num}",
 category=TrustCategory.RELEVANCE,
 severity=Severity.LOW,
 issue_type=IssueType.DOCUMENTATION_ISSUE,
 title="Unresolved TODO",
 description=f"TODO/FIXME comment found: {line.strip()}",
 file_path=file_path,
 line_number=line_num,
 column_number=line.find('TODO') if 'TODO' in line else line.find('FIXME'),
 code_snippet=line.strip(),
 suggested_fix="Address the TODO/FIXME item or remove the comment",
 confidence=0.6,
 rule_violated="UNRESOLVED_TODO"
 )
 issues.append(issue)

 return issues

 def _analyze_usability(self, file_path: str, content: str, tree: ast.AST) -> List[CodeIssue]:
 """Analyze code for usability and maintainability."""

 issues = []

 # Check for docstring presence
 for node in ast.walk(tree):
 if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
 if not ast.get_docstring(node):
 issue = CodeIssue(
 id=f"no_docstring_{node.lineno}",
 category=TrustCategory.USABILITY,
 severity=Severity.LOW,
 issue_type=IssueType.DOCUMENTATION_ISSUE,
 title="Missing Docstring",
 description=f"Function '{node.name}' is missing a docstring",
 file_path=file_path,
 line_number=node.lineno,
 column_number=1,
 code_snippet=f"def {node.name}(...):",
 suggested_fix=f"Add a docstring to '{node.name}' explaining its purpose, parameters, and return value",
 confidence=0.7,
 rule_violated="MISSING_DOCSTRING"
 )
 issues.append(issue)

 return issues

 def _analyze_safety(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
 """Analyze code for safety and error handling."""

 issues = []

 # Check for bare except clauses
 for node in ast.walk(tree):
 if isinstance(node, ast.ExceptHandler):
 if node.type is None:
 issue = CodeIssue(
 id=f"bare_except_{node.lineno}",
 category=TrustCategory.SAFETY,
 severity=Severity.MEDIUM,
 issue_type=IssueType.CODE_SMELL,
 title="Bare Except Clause",
 description="Bare except clause can hide unexpected errors",
 file_path=file_path,
 line_number=node.lineno,
 column_number=1,
 code_snippet="except:",
 suggested_fix="Specify exception types or use 'except Exception:' with logging",
 confidence=0.8,
 rule_violated="BARE_EXCEPT"
 )
 issues.append(issue)

 return issues

 def _analyze_timeliness(self, file_path: str, content: str) -> List[CodeIssue]:
 """Analyze code for timeliness and performance."""

 issues = []

 # Check for deprecated imports
 deprecated_imports = {
 'StringIO': 'io.StringIO',
 'cStringIO': 'io.StringIO'
 }

 lines = content.split('\n')
 for line_num, line in enumerate(lines, 1):
 for old_import, new_import in deprecated_imports.items():
 if f"import {old_import}" in line or f"from {old_import}" in line:
 issue = CodeIssue(
 id=f"deprecated_import_{line_num}",
 category=TrustCategory.TIMELINESS,
 severity=Severity.LOW,
 issue_type=IssueType.IMPORT_ISSUE,
 title="Deprecated Import",
 description=f"Using deprecated import '{old_import}', should use '{new_import}'",
 file_path=file_path,
 line_number=line_num,
 column_number=line.find(old_import),
 code_snippet=line.strip(),
 suggested_fix=f"Replace '{old_import}' with '{new_import}'",
 confidence=0.9,
 rule_violated="DEPRECATED_IMPORT",
 auto_fixable=True
 )
 issues.append(issue)

 return issues

 def _convert_static_issues(
 self, static_results: Dict[str, Any], file_path: str
 ) -> List[CodeIssue]:
 """Convert static analysis results to CodeIssue objects."""

 issues = []

 for tool_name, results in static_results.items():
 if 'error' in results:
 continue

 tool_issues = results.get('issues', [])
 for issue_data in tool_issues:
 # Map tool to TRUST category
 category = self._map_tool_to_trust_category(tool_name, issue_data)

 issue = CodeIssue(
 id=f"{tool_name}_{len(issues)}",
 category=category,
 severity=self._map_severity(issue_data.get('severity', 'medium')),
 issue_type=self._map_issue_type(tool_name, issue_data),
 title=f"{tool_name.title()}: {issue_data.get('message', 'Unknown issue')}",
 description=issue_data.get('message', 'Static analysis issue'),
 file_path=file_path,
 line_number=issue_data.get('line', 0),
 column_number=issue_data.get('column', 0),
 code_snippet=issue_data.get('code_snippet', ''),
 suggested_fix=self._get_suggested_fix(tool_name, issue_data),
 confidence=0.8,
 rule_violated=issue_data.get('code', ''),
 external_reference=f"{tool_name} documentation"
 )
 issues.append(issue)

 return issues

 def _map_tool_to_trust_category(self, tool_name: str, issue_data: Dict) -> TrustCategory:
 """Map static analysis tool to TRUST category."""

 if tool_name == 'bandit':
 return TrustCategory.SAFETY
 elif tool_name == 'mypy':
 return TrustCategory.TRUTHFULNESS
 elif tool_name == 'pylint':
 message = issue_data.get('message', '').lower()
 if any(keyword in message for keyword in ['security', 'injection', 'unsafe']):
 return TrustCategory.SAFETY
 elif any(keyword in message for keyword in ['performance', 'inefficient']):
 return TrustCategory.TIMELINESS
 else:
 return TrustCategory.USABILITY
 else:
 return TrustCategory.USABILITY

 def _map_severity(self, severity: str) -> Severity:
 """Map severity string to Severity enum."""

 severity_map = {
 'critical': Severity.CRITICAL,
 'high': Severity.HIGH,
 'medium': Severity.MEDIUM,
 'low': Severity.LOW,
 'info': Severity.INFO
 }

 return severity_map.get(severity.lower(), Severity.MEDIUM)

 def _map_issue_type(self, tool_name: str, issue_data: Dict) -> IssueType:
 """Map tool issue to IssueType enum."""

 if tool_name == 'bandit':
 return IssueType.SECURITY_VULNERABILITY
 elif tool_name == 'mypy':
 return IssueType.TYPE_ERROR
 else:
 message = issue_data.get('message', '').lower()
 if 'security' in message:
 return IssueType.SECURITY_VULNERABILITY
 elif 'performance' in message:
 return IssueType.PERFORMANCE_ISSUE
 elif 'syntax' in message:
 return IssueType.SYNTAX_ERROR
 else:
 return IssueType.CODE_SMELL

 def _get_suggested_fix(self, tool_name: str, issue_data: Dict) -> str:
 """Get suggested fix for tool issue."""

 message = issue_data.get('message', '')

 if 'unused' in message.lower():
 return "Remove unused variable or import"
 elif 'missing docstring' in message.lower():
 return "Add docstring explaining function purpose"
 elif 'too many arguments' in message.lower():
 return "Consider reducing function arguments or using data classes"
 elif else:
 return "Address the linting issue by following best practices"

 def _get_security_fix_suggestion(self, category: str, line: str) -> str:
 """Get security fix suggestion."""

 suggestions = {
 'sql_injection': "Use parameterized queries or ORM to prevent SQL injection",
 'command_injection': "Use subprocess.run with proper argument lists or validate input",
 'path_traversal': "Validate and sanitize file paths, use absolute paths"
 }

 return suggestions.get(category, "Review and fix security vulnerability")

 def _get_performance_fix_suggestion(self, category: str, line: str) -> str:
 """Get performance fix suggestion."""

 suggestions = {
 'inefficient_loops': "Use list comprehensions or generator expressions",
 'memory_leaks': "Review memory usage and ensure proper cleanup"
 }

 return suggestions.get(category, "Optimize code for better performance")

 def _get_security_reference(self, category: str) -> str:
 """Get external security reference."""

 references = {
 'sql_injection': "OWASP SQL Injection Prevention Cheat Sheet",
 'command_injection': "OWASP Command Injection Prevention Cheat Sheet",
 'path_traversal': "OWASP Path Traversal Prevention Cheat Sheet"
 }

 return references.get(category, "OWASP Top 10 Security Risks")

 def _calculate_file_metrics(self, content: str, tree: ast.AST) -> Dict[str, Any]:
 """Calculate comprehensive file metrics."""

 lines = content.split('\n')
 code_lines = [line for line in lines if line.strip() and not line.strip().startswith('#')]

 functions = [node for node in ast.walk(tree) if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))]
 classes = [node for node in ast.walk(tree) if isinstance(node, ast.ClassDef)]

 return {
 'total_lines': len(lines),
 'code_lines': len(code_lines),
 'comment_lines': len(lines) - len(code_lines),
 'functions': len(functions),
 'classes': len(classes),
 'imports': len([node for node in ast.walk(tree) if isinstance(node, (ast.Import, ast.ImportFrom))])
 }

 def _calculate_trust_scores(
 self, issues: List[CodeIssue], metrics: Dict[str, Any]
 ) -> Dict[str, Any]:
 """Calculate TRUST 5 scores."""

 category_scores = {}
 category_weights = {
 TrustCategory.TRUTHFULNESS: 0.25,
 TrustCategory.RELEVANCE: 0.20,
 TrustCategory.USABILITY: 0.25,
 TrustCategory.SAFETY: 0.20,
 TrustCategory.TIMELINESS: 0.10
 }

 # Group issues by category
 issues_by_category = {category: [] for category in TrustCategory}
 for issue in issues:
 issues_by_category[issue.category].append(issue)

 # Calculate scores for each category
 for category in TrustCategory:
 category_issues = issues_by_category[category]

 # Calculate penalty based on severity and number of issues
 penalty = 0.0
 for issue in category_issues:
 severity_penalty = {
 Severity.CRITICAL: 0.5,
 Severity.HIGH: 0.3,
 Severity.MEDIUM: 0.1,
 Severity.LOW: 0.05,
 Severity.INFO: 0.01
 }
 penalty += severity_penalty.get(issue.severity, 0.1) * issue.confidence

 # Apply penalties (max penalty of 1.0)
 score = max(0.0, 1.0 - min(penalty, 1.0))
 category_scores[category] = score

 # Calculate overall score
 overall_score = sum(
 category_scores[cat] * category_weights[cat]
 for cat in TrustCategory
 )

 return {
 'overall': overall_score,
 'categories': category_scores
 }

 def _calculate_complexity_metrics(self, content: str, tree: ast.AST) -> Dict[str, float]:
 """Calculate complexity metrics."""

 total_complexity = 0
 max_function_complexity = 0
 function_count = 0

 for node in ast.walk(tree):
 if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
 complexity = self._calculate_cyclomatic_complexity(node)
 total_complexity += complexity
 max_function_complexity = max(max_function_complexity, complexity)
 function_count += 1

 avg_complexity = total_complexity / max(function_count, 1)

 return {
 'total_complexity': total_complexity,
 'max_function_complexity': max_function_complexity,
 'average_complexity': avg_complexity,
 'function_count': function_count
 }

 def _generate_comprehensive_report(
 self, project_path: str, file_results: List[FileReviewResult], duration: float
 ) -> CodeReviewReport:
 """Generate comprehensive code review report."""

 # Collect all issues
 all_issues = []
 for result in file_results:
 all_issues.extend(result.issues)

 # Calculate overall scores
 overall_category_scores = {}
 for category in TrustCategory:
 scores = [result.category_scores.get(category, 0.0) for result in file_results]
 overall_category_scores[category] = sum(scores) / len(scores) if scores else 0.0

 overall_trust_score = sum(overall_category_scores.values()) / len(overall_category_scores)

 # Get critical issues
 critical_issues = [issue for issue in all_issues if issue.severity == Severity.CRITICAL]

 # Generate recommendations
 recommendations = self._generate_recommendations(overall_category_scores, all_issues)

 # Calculate summary metrics
 summary_metrics = {
 'files_reviewed': len(file_results),
 'total_issues': len(all_issues),
 'critical_issues': len(critical_issues),
 'issues_by_severity': {
 severity.value: len([i for i in all_issues if i.severity == severity])
 for severity in Severity
 },
 'issues_by_category': {
 category.value: len([i for i in all_issues if i.category == category])
 for category in TrustCategory
 },
 'total_lines_of_code': sum(result.lines_of_code for result in file_results),
 'average_trust_score': overall_trust_score,
 'context7_patterns_used': list(self.analysis_patterns.keys())
 }

 return CodeReviewReport(
 project_path=project_path,
 files_reviewed=file_results,
 overall_trust_score=overall_trust_score,
 overall_category_scores=overall_category_scores,
 summary_metrics=summary_metrics,
 recommendations=recommendations,
 critical_issues=critical_issues,
 review_duration=duration,
 context7_patterns_used=list(self.analysis_patterns.keys())
 )

 def _generate_recommendations(
 self, category_scores: Dict[TrustCategory, float], issues: List[CodeIssue]
 ) -> List[str]:
 """Generate actionable recommendations."""

 recommendations = []

 # Category-specific recommendations
 for category, score in category_scores.items():
 if score < 0.7:
 if category == TrustCategory.SAFETY:
 recommendations.append("Address security vulnerabilities immediately - critical safety issues detected")
 elif category == TrustCategory.TRUTHFULNESS:
 recommendations.append("Review code logic and fix correctness issues")
 elif category == TrustCategory.USABILITY:
 recommendations.append("Improve code maintainability by refactoring complex functions")
 elif category == TrustCategory.RELEVANCE:
 recommendations.append("Remove TODO items and improve documentation")
 elif category == TrustCategory.TIMELINESS:
 recommendations.append("Optimize performance issues and update deprecated code")

 # General recommendations
 high_severity_count = len([i for i in issues if i.severity in [Severity.CRITICAL, Severity.HIGH]])
 if high_severity_count > 0:
 recommendations.append(f"Address {high_severity_count} high-priority issues before release")

 auto_fixable_count = len([i for i in issues if i.auto_fixable])
 if auto_fixable_count > 0:
 recommendations.append(f"Use automated fixes for {auto_fixable_count} auto-fixable issues")

 return recommendations

 def _create_error_result(self, file_path: str, error_message: str) -> FileReviewResult:
 """Create error result for file that couldn't be processed."""

 return FileReviewResult(
 file_path=file_path,
 issues=[],
 metrics={'error': error_message},
 trust_score=0.0,
 category_scores={cat: 0.0 for cat in TrustCategory},
 lines_of_code=0,
 complexity_metrics={},
 review_timestamp=time.time()
 )

 def _create_syntax_error_result(
 self, file_path: str, content: str, syntax_error: SyntaxError
 ) -> FileReviewResult:
 """Create result for file with syntax errors."""

 issue = CodeIssue(
 id=f"syntax_error_{syntax_error.lineno}",
 category=TrustCategory.TRUTHFULNESS,
 severity=Severity.CRITICAL,
 issue_type=IssueType.SYNTAX_ERROR,
 title="Syntax Error",
 description=f"Syntax error: {syntax_error.msg}",
 file_path=file_path,
 line_number=syntax_error.lineno,
 column_number=syntax_error.offset or 0,
 code_snippet=content.split('\n')[syntax_error.lineno - 1] if syntax_error.lineno <= len(content.split('\n')) else "",
 suggested_fix="Fix the syntax error",
 confidence=1.0
 )

 return FileReviewResult(
 file_path=file_path,
 issues=[issue],
 metrics={'syntax_error': True},
 trust_score=0.0,
 category_scores={cat: 0.0 for cat in TrustCategory},
 lines_of_code=len(content.split('\n')),
 complexity_metrics={},
 review_timestamp=time.time()
 )

# Usage Examples
"""
# Initialize automated code reviewer
reviewer = AutomatedCodeReviewer(context7_client=context7)

# Review entire codebase
report = await reviewer.review_codebase(
 project_path="/path/to/project",
 include_patterns=["/*.py"],
 exclude_patterns=["/tests/", "/__pycache__/"]
)

print(f"Code Review Results:")
print(f" Overall TRUST Score: {report.overall_trust_score:.2f}")
print(f" Files Reviewed: {report.summary_metrics['files_reviewed']}")
print(f" Total Issues: {report.summary_metrics['total_issues']}")
print(f" Critical Issues: {report.summary_metrics['critical_issues']}")

print(f"\nTRUST 5 Category Scores:")
for category, score in report.overall_category_scores.items():
 print(f" {category.value}: {score:.2f}")

print(f"\nTop Recommendations:")
for i, rec in enumerate(report.recommendations[:5], 1):
 print(f" {i}. {rec}")

print(f"\nCritical Issues:")
for issue in report.critical_issues[:3]:
 print(f" - {issue.title} in {issue.file_path}:{issue.line_number}")
 print(f" {issue.description}")

# Review single file
file_result = await reviewer.review_single_file("/path/to/file.py")
print(f"\nFile Trust Score: {file_result.trust_score:.2f}")
print(f"Issues found: {len(file_result.issues)}")
"""
```

## Advanced Features

### Context7-Enhanced Security Analysis

Advanced Security Pattern Detection:
```python
class EnhancedSecurityAnalyzer:
 """Enhanced security analyzer with Context7 integration."""

 async def analyze_advanced_security_patterns(
 self, file_path: str, content: str
 ) -> List[CodeIssue]:
 """Analyze advanced security patterns using Context7."""

 issues = []

 # Get latest security patterns from Context7
 security_patterns = await self.context7.get_library_docs(
 context7_library_id="/security/owasp",
 topic="advanced security vulnerability patterns 2025",
 tokens=5000
 )

 # Analyze for OWASP Top 10 vulnerabilities
 owasp_issues = await self._analyze_owasp_patterns(file_path, content, security_patterns)
 issues.extend(owasp_issues)

 # Analyze for business logic vulnerabilities
 logic_issues = await self._analyze_business_logic_security(file_path, content)
 issues.extend(logic_issues)

 return issues
```

## Best Practices

1. Comprehensive Coverage: Analyze code across all TRUST 5 dimensions
2. Context Integration: Leverage Context7 for up-to-date security and quality patterns
3. Actionable Feedback: Provide specific, implementable suggestions
4. Severity Prioritization: Focus on critical and high-severity issues first
5. Continuous Integration: Integrate into CI/CD pipeline for automated reviews

---

Module: `modules/automated-code-review.md`
Related: [Smart Refactoring](./smart-refactoring.md) | [Performance Optimization](./performance-optimization.md)
