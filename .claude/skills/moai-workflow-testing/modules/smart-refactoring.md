# Smart Refactoring with Technical Debt Management

> Module: AI-powered code refactoring with technical debt analysis and safe transformation
> Complexity: Advanced
> Time: 25+ minutes
> Dependencies: Python 3.8+, Rope, AST, Context7 MCP, asyncio, dataclasses

## Core Implementation

### AIRefactorer Class

```python
import ast
import asyncio
from typing import Dict, List, Optional, Any, Union, Tuple
from dataclasses import dataclass
from enum import Enum
import rope.base.project
import rope.base.libutils
from rope.refactor import rename, move, extract, inline

class RefactorType(Enum):
 """Types of refactoring operations."""
 RENAME = "rename"
 EXTRACT_METHOD = "extract_method"
 EXTRACT_VARIABLE = "extract_variable"
 INLINE_VARIABLE = "inline_variable"
 MOVE_MODULE = "move_module"
 REORGANIZE_IMPORTS = "reorganize_imports"
 SIMPLIFY_CONDITIONALS = "simplify_conditionals"
 EXTRACT_CLASS = "extract_class"
 REMOVE_DEAD_CODE = "remove_dead_code"
 OPTIMIZE_IMPORTS = "optimize_imports"

class TechnicalDebtType(Enum):
 """Categories of technical debt."""
 CODE_COMPLEXITY = "code_complexity"
 DUPLICATION = "duplication"
 LONG_METHODS = "long_methods"
 LARGE_CLASSES = "large_classes"
 DEEP_NESTING = "deep_nesting"
 POOR_NAMING = "poor_naming"
 MISSING_TESTS = "missing_tests"
 SECURITY_ISSUES = "security_issues"
 PERFORMANCE_ISSUES = "performance_issues"
 DOCUMENTATION_DEBT = "documentation_debt"

@dataclass
class TechnicalDebtItem:
 """Individual technical debt item with metrics."""
 type: TechnicalDebtType
 file_path: str
 line_number: int
 severity: str # "low", "medium", "high", "critical"
 description: str
 impact: str # "maintainability", "readability", "performance", "security"
 estimated_effort: str # "minutes", "hours", "days"
 code_snippet: str
 suggested_fix: str
 dependencies: List[str]

@dataclass
class RefactorOpportunity:
 """Potential refactoring opportunity with analysis."""
 type: RefactorType
 file_path: str
 line_range: Tuple[int, int]
 confidence: float
 complexity_reduction: float
 risk_level: str # "low", "medium", "high"
 description: str
 before_code: str
 after_code: str
 technical_debt_addresses: List[TechnicalDebtType]

@dataclass
class RefactorPlan:
 """Comprehensive refactoring plan with execution strategy."""
 opportunities: List[RefactorOpportunity]
 technical_debt_items: List[TechnicalDebtItem]
 execution_order: List[int] # Indices of opportunities in execution order
 estimated_time: str
 risk_assessment: str
 prerequisites: List[str]
 rollback_strategy: str

class TechnicalDebtAnalyzer:
 """Analyzes codebase for technical debt patterns."""

 def __init__(self):
 self.debt_patterns = self._load_debt_patterns()
 self.complexity_metrics = {}

 def _load_debt_patterns(self) -> Dict[str, Any]:
 """Load technical debt detection patterns."""
 return {
 TechnicalDebtType.CODE_COMPLEXITY: {
 'thresholds': {
 'cyclomatic_complexity': 10,
 'cognitive_complexity': 15,
 'nesting_depth': 4
 },
 'indicators': [
 'high_cyclomatic_complexity',
 'deep_nesting',
 'multiple_responsibilities'
 ]
 },
 TechnicalDebtType.DUPLICATION: {
 'thresholds': {
 'similarity_threshold': 0.8,
 'min_lines': 5
 },
 'indicators': [
 'similar_code_blocks',
 'repeated_patterns',
 'copied_pasted_code'
 ]
 },
 TechnicalDebtType.LONG_METHODS: {
 'thresholds': {
 'max_lines': 50,
 'max_parameters': 7
 },
 'indicators': [
 'excessive_length',
 'too_many_parameters',
 'multiple_responsibilities'
 ]
 },
 TechnicalDebtType.POOR_NAMING: {
 'patterns': [
 r'^[a-z]$',
 r'^[a-z]{1,2}$',
 r'^[A-Z]+_[A-Z_]+$',
 r'^temp.*',
 r'^tmp.*'
 ]
 }
 }

 async def analyze(self, codebase_path: str) -> List[TechnicalDebtItem]:
 """Analyze codebase for technical debt."""

 debt_items = []

 # Find all Python files
 python_files = self._find_python_files(codebase_path)

 for file_path in python_files:
 file_debt = await self._analyze_file(file_path)
 debt_items.extend(file_debt)

 # Cross-file analysis for duplication
 duplication_debt = await self._analyze_duplication(python_files)
 debt_items.extend(duplication_debt)

 return self._prioritize_debt_items(debt_items)

 def _find_python_files(self, codebase_path: str) -> List[str]:
 """Find all Python files in codebase."""
 import os
 python_files = []

 for root, dirs, files in os.walk(codebase_path):
 # Skip common non-source directories
 dirs[:] = [d for d in dirs if d not in ['__pycache__', '.git', 'venv', 'node_modules']]

 for file in files:
 if file.endswith('.py'):
 python_files.append(os.path.join(root, file))

 return python_files

 async def _analyze_file(self, file_path: str) -> List[TechnicalDebtItem]:
 """Analyze individual file for technical debt."""

 debt_items = []

 try:
 with open(file_path, 'r', encoding='utf-8') as f:
 content = f.read()

 # Parse AST
 tree = ast.parse(content)

 # Analyze complexity
 complexity_debt = self._analyze_complexity(tree, file_path, content)
 debt_items.extend(complexity_debt)

 # Analyze method length
 length_debt = self._analyze_method_length(tree, file_path, content)
 debt_items.extend(length_debt)

 # Analyze naming
 naming_debt = self._analyze_naming(tree, file_path, content)
 debt_items.extend(naming_debt)

 except Exception as e:
 print(f"Error analyzing {file_path}: {e}")

 return debt_items

 def _analyze_complexity(
 self, tree: ast.AST, file_path: str, content: str
 ) -> List[TechnicalDebtItem]:
 """Analyze cyclomatic and cognitive complexity."""

 debt_items = []
 lines = content.split('\n')

 class ComplexityVisitor(ast.NodeVisitor):
 def __init__(self):
 self.complexities = {}
 self.nesting_depths = {}
 self.current_depth = 0

 def visit_FunctionDef(self, node):
 # Calculate cyclomatic complexity
 complexity = 1 # Base complexity
 for child in ast.walk(node):
 if isinstance(child, (ast.If, ast.While, ast.For, ast.AsyncFor)):
 complexity += 1
 elif isinstance(child, ast.ExceptHandler):
 complexity += 1
 elif isinstance(child, ast.With, ast.AsyncWith):
 complexity += 1
 elif isinstance(child, ast.BoolOp):
 complexity += len(child.values) - 1

 # Calculate maximum nesting depth
 self.current_depth = 0
 max_depth = self._calculate_nesting_depth(node)

 self.complexities[node.name] = {
 'complexity': complexity,
 'line': node.lineno,
 'max_depth': max_depth
 }

 self.generic_visit(node)

 def _calculate_nesting_depth(self, node, current_depth=0):
 """Calculate maximum nesting depth."""
 max_depth = current_depth

 for child in ast.walk(node):
 if isinstance(child, (ast.If, ast.While, ast.For, ast.AsyncFor,
 ast.With, ast.AsyncWith, ast.Try)):
 child_depth = self._calculate_nesting_depth(child, current_depth + 1)
 max_depth = max(max_depth, child_depth)

 return max_depth

 visitor = ComplexityVisitor()
 visitor.visit(tree)

 # Create debt items for high complexity
 for func_name, metrics in visitor.complexities.items():
 if metrics['complexity'] > 10: # Cyclomatic complexity threshold
 debt_item = TechnicalDebtItem(
 type=TechnicalDebtType.CODE_COMPLEXITY,
 file_path=file_path,
 line_number=metrics['line'],
 severity=self._assess_complexity_severity(metrics['complexity']),
 description=f"Function '{func_name}' has high cyclomatic complexity: {metrics['complexity']}",
 impact="maintainability",
 estimated_effort="hours",
 code_snippet=lines[metrics['line'] - 1] if metrics['line'] <= len(lines) else "",
 suggested_fix=f"Consider extracting sub-functions or simplifying logic in '{func_name}'",
 dependencies=["unit_tests"]
 )
 debt_items.append(debt_item)

 if metrics['max_depth'] > 4: # Nesting depth threshold
 debt_item = TechnicalDebtItem(
 type=TechnicalDebtType.DEEP_NESTING,
 file_path=file_path,
 line_number=metrics['line'],
 severity=self._assess_nesting_severity(metrics['max_depth']),
 description=f"Function '{func_name}' has deep nesting: {metrics['max_depth']} levels",
 impact="readability",
 estimated_effort="minutes",
 code_snippet=lines[metrics['line'] - 1] if metrics['line'] <= len(lines) else "",
 suggested_fix=f"Consider using early returns or guard clauses in '{func_name}'",
 dependencies=[]
 )
 debt_items.append(debt_item)

 return debt_items

 def _analyze_method_length(
 self, tree: ast.AST, file_path: str, content: str
 ) -> List[TechnicalDebtItem]:
 """Analyze method length violations."""

 debt_items = []
 lines = content.split('\n')

 class MethodLengthVisitor(ast.NodeVisitor):
 def __init__(self):
 self.methods = []

 def visit_FunctionDef(self, node):
 # Calculate method length (excluding docstrings and comments)
 start_line = node.lineno

 # Find end of function
 end_line = start_line
 for child in ast.walk(node):
 if hasattr(child, 'lineno') and child.lineno > end_line:
 end_line = child.lineno

 # Count non-empty, non-comment lines
 method_lines = []
 for i in range(start_line - 1, min(end_line, len(lines))):
 line = lines[i].strip()
 if line and not line.startswith('#') and not line.startswith('"""') and not line.startswith("'''"):
 method_lines.append(line)

 length = len(method_lines)
 param_count = len(node.args.args)

 self.methods.append({
 'name': node.name,
 'line': start_line,
 'length': length,
 'param_count': param_count
 })

 self.generic_visit(node)

 visitor = MethodLengthVisitor()
 visitor.visit(tree)

 # Create debt items for long methods
 for method in visitor.methods:
 if method['length'] > 50: # Lines threshold
 debt_item = TechnicalDebtItem(
 type=TechnicalDebtType.LONG_METHODS,
 file_path=file_path,
 line_number=method['line'],
 severity=self._assess_length_severity(method['length']),
 description=f"Method '{method['name']}' is too long: {method['length']} lines",
 impact="maintainability",
 estimated_effort="hours",
 code_snippet=lines[method['line'] - 1] if method['line'] <= len(lines) else "",
 suggested_fix=f"Consider extracting smaller methods from '{method['name']}'",
 dependencies=["unit_tests"]
 )
 debt_items.append(debt_item)

 if method['param_count'] > 7: # Parameters threshold
 debt_item = TechnicalDebtItem(
 type=TechnicalDebtType.LARGE_CLASSES, # Could also be its own type
 file_path=file_path,
 line_number=method['line'],
 severity="medium",
 description=f"Method '{method['name']}' has too many parameters: {method['param_count']}",
 impact="readability",
 estimated_effort="minutes",
 code_snippet=lines[method['line'] - 1] if method['line'] <= len(lines) else "",
 suggested_fix=f"Consider using parameter objects or configuration dictionaries",
 dependencies=[]
 )
 debt_items.append(debt_item)

 return debt_items

 def _analyze_naming(
 self, tree: ast.AST, file_path: str, content: str
 ) -> List[TechnicalDebtItem]:
 """Analyze naming convention violations."""

 debt_items = []
 lines = content.split('\n')
 patterns = self.debt_patterns[TechnicalDebtType.POOR_NAMING]['patterns']

 class NamingVisitor(ast.NodeVisitor):
 def __init__(self):
 self.variables = []
 self.functions = []
 self.classes = []

 def visit_Name(self, node):
 if isinstance(node.ctx, ast.Store):
 self.variables.append({
 'name': node.id,
 'line': node.lineno
 })
 self.generic_visit(node)

 def visit_FunctionDef(self, node):
 self.functions.append({
 'name': node.name,
 'line': node.lineno
 })
 self.generic_visit(node)

 def visit_ClassDef(self, node):
 self.classes.append({
 'name': node.name,
 'line': node.lineno
 })
 self.generic_visit(node)

 visitor = NamingVisitor()
 visitor.visit(tree)

 import re

 # Check variable names
 for var in visitor.variables:
 for pattern in patterns:
 if re.match(pattern, var['name']):
 debt_item = TechnicalDebtItem(
 type=TechnicalDebtType.POOR_NAMING,
 file_path=file_path,
 line_number=var['line'],
 severity="low",
 description=f"Variable '{var['name']}' has poor naming pattern",
 impact="readability",
 estimated_effort="minutes",
 code_snippet=lines[var['line'] - 1] if var['line'] <= len(lines) else "",
 suggested_fix=f"Rename '{var['name']}' to something more descriptive",
 dependencies=[]
 )
 debt_items.append(debt_item)

 return debt_items

 async def _analyze_duplication(
 self, python_files: List[str]
 ) -> List[TechnicalDebtItem]:
 """Analyze code duplication across files."""

 debt_items = []
 code_blocks = {}

 # Extract code blocks from all files
 for file_path in python_files:
 try:
 with open(file_path, 'r', encoding='utf-8') as f:
 content = f.read()

 # Split into logical blocks (functions, classes, etc.)
 tree = ast.parse(content)
 blocks = self._extract_code_blocks(tree, content)

 for block in blocks:
 block_hash = hash(block['content'])
 if block_hash not in code_blocks:
 code_blocks[block_hash] = []
 code_blocks[block_hash].append({
 'file': file_path,
 'line': block['line'],
 'content': block['content']
 })

 except Exception as e:
 print(f"Error analyzing duplication in {file_path}: {e}")

 # Find duplicated blocks
 for block_hash, occurrences in code_blocks.items():
 if len(occurrences) > 1:
 # This is duplicated code
 for occ in occurrences:
 debt_item = TechnicalDebtItem(
 type=TechnicalDebtType.DUPLICATION,
 file_path=occ['file'],
 line_number=occ['line'],
 severity="medium",
 description=f"Code block duplicated in {len(occurrences)} locations",
 impact="maintainability",
 estimated_effort="hours",
 code_snippet=occ['content'][:100] + "..." if len(occ['content']) > 100 else occ['content'],
 suggested_fix="Extract duplicated code into a shared function or method",
 dependencies=["unit_tests"]
 )
 debt_items.append(debt_item)

 return debt_items

 def _extract_code_blocks(
 self, tree: ast.AST, content: str
 ) -> List[Dict[str, Any]]:
 """Extract logical code blocks from AST."""

 blocks = []
 lines = content.split('\n')

 class BlockExtractor(ast.NodeVisitor):
 def visit_FunctionDef(self, node):
 # Extract function content
 start_line = node.lineno
 end_line = start_line
 for child in ast.walk(node):
 if hasattr(child, 'lineno') and child.lineno > end_line:
 end_line = child.lineno

 if start_line <= len(lines) and end_line <= len(lines):
 block_content = '\n'.join(lines[start_line - 1:end_line])
 blocks.append({
 'type': 'function',
 'name': node.name,
 'line': start_line,
 'content': block_content
 })

 self.generic_visit(node)

 def visit_ClassDef(self, node):
 # Extract class content
 start_line = node.lineno
 end_line = start_line
 for child in ast.walk(node):
 if hasattr(child, 'lineno') and child.lineno > end_line:
 end_line = child.lineno

 if start_line <= len(lines) and end_line <= len(lines):
 block_content = '\n'.join(lines[start_line - 1:end_line])
 blocks.append({
 'type': 'class',
 'name': node.name,
 'line': start_line,
 'content': block_content
 })

 self.generic_visit(node)

 extractor = BlockExtractor()
 extractor.visit(tree)

 return blocks

 def _prioritize_debt_items(
 self, debt_items: List[TechnicalDebtItem]
 ) -> List[TechnicalDebtItem]:
 """Prioritize technical debt items by severity and impact."""

 # Sort by severity (critical > high > medium > low) and impact
 severity_order = {'critical': 4, 'high': 3, 'medium': 2, 'low': 1}
 impact_order = {'security': 4, 'performance': 3, 'maintainability': 2, 'readability': 1}

 return sorted(
 debt_items,
 key=lambda x: (
 severity_order.get(x.severity, 0),
 impact_order.get(x.impact, 0)
 ),
 reverse=True
 )

 def _assess_complexity_severity(self, complexity: int) -> str:
 """Assess severity based on cyclomatic complexity."""
 if complexity >= 20:
 return "critical"
 elif complexity >= 15:
 return "high"
 elif complexity >= 10:
 return "medium"
 else:
 return "low"

 def _assess_nesting_severity(self, depth: int) -> str:
 """Assess severity based on nesting depth."""
 if depth >= 6:
 return "critical"
 elif depth >= 5:
 return "high"
 elif depth >= 4:
 return "medium"
 else:
 return "low"

 def _assess_length_severity(self, length: int) -> str:
 """Assess severity based on method length."""
 if length >= 100:
 return "critical"
 elif length >= 75:
 return "high"
 elif length >= 50:
 return "medium"
 else:
 return "low"

class AIRefactorer:
 """AI-powered refactoring with technical debt management."""

 def __init__(self, context7_client=None):
 self.context7 = context7_client
 self.technical_debt_analyzer = TechnicalDebtAnalyzer()
 self.rope_project = None

 async def refactor_with_intelligence(
 self, codebase_path: str, refactor_options: Dict = None
 ) -> RefactorPlan:
 """AI-driven code transformation with technical debt quantification."""

 # Initialize Rope project
 self.rope_project = rope.base.project.Project(codebase_path)

 # Analyze technical debt
 debt_analysis = await self.technical_debt_analyzer.analyze(codebase_path)

 # Get Context7 refactoring patterns
 context7_patterns = {}
 if self.context7:
 context7_patterns = await self._get_context7_refactoring_patterns()

 # AI analysis of refactoring opportunities
 refactor_opportunities = await self._identify_refactor_opportunities(
 codebase_path, debt_analysis, context7_patterns
 )

 # Generate safe refactor plan using Rope + AI
 refactor_plan = self._create_safe_refactor_plan(
 refactor_opportunities, debt_analysis, context7_patterns
 )

 return refactor_plan

 async def _get_context7_refactoring_patterns(self) -> Dict[str, Any]:
 """Get latest refactoring patterns from Context7."""

 patterns = {}
 if self.context7:
 try:
 # Rope patterns
 rope_patterns = await self.context7.get_library_docs(
 context7_library_id="/python-rope/rope",
 topic="safe refactoring patterns technical debt 2025",
 tokens=4000
 )
 patterns['rope'] = rope_patterns

 # General refactoring best practices
 refactoring_patterns = await self.context7.get_library_docs(
 context7_library_id="/refactoring/guru",
 topic="code refactoring best practices design patterns 2025",
 tokens=3000
 )
 patterns['general'] = refactoring_patterns

 except Exception as e:
 print(f"Failed to get Context7 patterns: {e}")

 return patterns

 async def _identify_refactor_opportunities(
 self, codebase_path: str, debt_items: List[TechnicalDebtItem],
 context7_patterns: Dict[str, Any]
 ) -> List[RefactorOpportunity]:
 """Identify refactoring opportunities using AI analysis."""

 opportunities = []

 # Group debt items by file
 debt_by_file = {}
 for item in debt_items:
 if item.file_path not in debt_by_file:
 debt_by_file[item.file_path] = []
 debt_by_file[item.file_path].append(item)

 # Analyze each file for refactoring opportunities
 for file_path, file_debt in debt_by_file.items():
 file_opportunities = await self._analyze_file_for_refactoring(
 file_path, file_debt, context7_patterns
 )
 opportunities.extend(file_opportunities)

 # Sort by confidence and complexity reduction
 opportunities.sort(
 key=lambda x: (x.confidence, x.complexity_reduction),
 reverse=True
 )

 return opportunities

 async def _analyze_file_for_refactoring(
 self, file_path: str, debt_items: List[TechnicalDebtItem],
 context7_patterns: Dict[str, Any]
 ) -> List[RefactorOpportunity]:
 """Analyze specific file for refactoring opportunities."""

 opportunities = []

 try:
 with open(file_path, 'r', encoding='utf-8') as f:
 content = f.read()

 tree = ast.parse(content)

 # Extract method opportunities
 method_opportunities = self._identify_method_refactoring(
 tree, file_path, debt_items, content
 )
 opportunities.extend(method_opportunities)

 # Extract variable opportunities
 variable_opportunities = self._identify_variable_refactoring(
 tree, file_path, content
 )
 opportunities.extend(variable_opportunities)

 # Extract import opportunities
 import_opportunities = self._identify_import_refactoring(
 tree, file_path, content
 )
 opportunities.extend(import_opportunities)

 except Exception as e:
 print(f"Error analyzing file {file_path}: {e}")

 return opportunities

 def _identify_method_refactoring(
 self, tree: ast.AST, file_path: str, debt_items: List[TechnicalDebtItem],
 content: str
 ) -> List[RefactorOpportunity]:
 """Identify method-level refactoring opportunities."""

 opportunities = []
 lines = content.split('\n')

 class MethodAnalyzer(ast.NodeVisitor):
 def __init__(self):
 self.methods = []
 self.current_class = None

 def visit_ClassDef(self, node):
 self.current_class = node.name
 self.generic_visit(node)
 self.current_class = None

 def visit_FunctionDef(self, node):
 # Calculate method metrics
 start_line = node.lineno
 end_line = start_line

 for child in ast.walk(node):
 if hasattr(child, 'lineno') and child.lineno > end_line:
 end_line = child.lineno

 method_lines = lines[start_line - 1:end_line]
 method_content = '\n'.join(method_lines)

 # Calculate complexity
 complexity = 1 # Base complexity
 for child in ast.walk(node):
 if isinstance(child, (ast.If, ast.While, ast.For, ast.AsyncFor)):
 complexity += 1
 elif isinstance(child, ast.ExceptHandler):
 complexity += 1

 self.methods.append({
 'name': node.name,
 'class': self.current_class,
 'line': start_line,
 'end_line': end_line,
 'content': method_content,
 'complexity': complexity,
 'param_count': len(node.args.args),
 'lines_count': len([l for l in method_lines if l.strip() and not l.strip().startswith('#')])
 })

 self.generic_visit(node)

 analyzer = MethodAnalyzer()
 analyzer.visit(tree)

 # Generate refactoring opportunities
 for method in analyzer.methods:
 # Extract method opportunity
 if method['lines_count'] > 30 or method['complexity'] > 8:
 opportunity = RefactorOpportunity(
 type=RefactorType.EXTRACT_METHOD,
 file_path=file_path,
 line_range=(method['line'], method['end_line']),
 confidence=min(0.9, method['complexity'] / 10),
 complexity_reduction=max(0.3, method['complexity'] / 20),
 risk_level="medium" if method['complexity'] > 15 else "low",
 description=f"Extract parts of method '{method['name']}' to reduce complexity",
 before_code=method['content'][:200] + "..." if len(method['content']) > 200 else method['content'],
 after_code="# Extracted sub-method calls would go here",
 technical_debt_addresses=[TechnicalDebtType.CODE_COMPLEXITY, TechnicalDebtType.LONG_METHODS]
 )
 opportunities.append(opportunity)

 # Extract variable opportunities for complex expressions
 if self._has_complex_expressions(method['content']):
 opportunity = RefactorOpportunity(
 type=RefactorType.EXTRACT_VARIABLE,
 file_path=file_path,
 line_range=(method['line'], method['end_line']),
 confidence=0.7,
 complexity_reduction=0.2,
 risk_level="low",
 description=f"Extract complex expressions in method '{method['name']}' to variables",
 before_code=method['content'][:200] + "..." if len(method['content']) > 200 else method['content'],
 after_code="# Extracted variables would improve readability",
 technical_debt_addresses=[TechnicalDebtType.CODE_COMPLEXITY]
 )
 opportunities.append(opportunity)

 return opportunities

 def _identify_variable_refactoring(
 self, tree: ast.AST, file_path: str, content: str
 ) -> List[RefactorOpportunity]:
 """Identify variable-level refactoring opportunities."""

 opportunities = []

 class VariableAnalyzer(ast.NodeVisitor):
 def __init__(self):
 self.variables = {}
 self.assignments = []

 def visit_Assign(self, node):
 for target in node.targets:
 if isinstance(target, ast.Name):
 var_name = target.id

 if var_name not in self.variables:
 self.variables[var_name] = []

 self.variables[var_name].append({
 'line': node.lineno,
 'type': 'assignment'
 })

 # Check for complex expressions
 if self._is_complex_expression(node.value):
 self.assignments.append({
 'name': var_name,
 'line': node.lineno,
 'value': node.value
 })

 self.generic_visit(node)

 def _is_complex_expression(self, node):
 """Check if expression is complex enough to extract."""
 complexity = 0
 for child in ast.walk(node):
 if isinstance(child, (ast.BinOp, ast.BoolOp, ast.Call)):
 complexity += 1
 elif isinstance(child, ast.Compare):
 complexity += len(child.ops)
 return complexity > 3

 analyzer = VariableAnalyzer()
 analyzer.visit(tree)

 # Generate extraction opportunities
 for assignment in analyzer.assignments:
 opportunity = RefactorOpportunity(
 type=RefactorType.EXTRACT_VARIABLE,
 file_path=file_path,
 line_range=(assignment['line'], assignment['line']),
 confidence=0.8,
 complexity_reduction=0.15,
 risk_level="low",
 description=f"Extract complex expression for variable '{assignment['name']}'",
 before_code=f"complex_assignment = ...",
 after_code=f"extracted_value = calculate_complex_logic()\n{assignment['name']} = extracted_value",
 technical_debt_addresses=[TechnicalDebtType.CODE_COMPLEXITY]
 )
 opportunities.append(opportunity)

 return opportunities

 def _identify_import_refactoring(
 self, tree: ast.AST, file_path: str, content: str
 ) -> List[RefactorOpportunity]:
 """Identify import-level refactoring opportunities."""

 opportunities = []
 lines = content.split('\n')

 class ImportAnalyzer(ast.NodeVisitor):
 def __init__(self):
 self.imports = []
 self.import_lines = set()

 def visit_Import(self, node):
 for alias in node.names:
 self.imports.append({
 'type': 'import',
 'module': alias.name,
 'alias': alias.asname,
 'line': node.lineno
 })
 self.import_lines.add(node.lineno)
 self.generic_visit(node)

 def visit_ImportFrom(self, node):
 for alias in node.names:
 self.imports.append({
 'type': 'from_import',
 'module': node.module,
 'name': alias.name,
 'alias': alias.asname,
 'line': node.lineno
 })
 self.import_lines.add(node.lineno)
 self.generic_visit(node)

 analyzer = ImportAnalyzer()
 analyzer.visit(tree)

 # Check for import optimization opportunities
 if len(analyzer.imports) > 10: # Too many imports
 opportunity = RefactorOpportunity(
 type=RefactorType.REORGANIZE_IMPORTS,
 file_path=file_path,
 line_range=(min(imp['line'] for imp in analyzer.imports),
 max(imp['line'] for imp in analyzer.imports)),
 confidence=0.9,
 complexity_reduction=0.1,
 risk_level="low",
 description="Reorganize and optimize imports structure",
 before_code="# Multiple scattered imports",
 after_code="# Organized imports with proper grouping",
 technical_debt_addresses=[]
 )
 opportunities.append(opportunity)

 return opportunities

 def _has_complex_expressions(self, method_content: str) -> bool:
 """Check if method contains complex expressions that could be extracted."""

 # Simple heuristic for complex expressions
 complexity_indicators = [
 ' and ' in method_content,
 ' or ' in method_content,
 method_content.count('(') > 10,
 method_content.count('.') > 15,
 len(method_content.split('\n')) > 20
 ]

 return sum(complexity_indicators) >= 2

 def _create_safe_refactor_plan(
 self, opportunities: List[RefactorOpportunity],
 debt_items: List[TechnicalDebtItem],
 context7_patterns: Dict[str, Any]
 ) -> RefactorPlan:
 """Create safe refactor plan with execution strategy."""

 # Filter opportunities by confidence and risk
 safe_opportunities = [
 opp for opp in opportunities
 if opp.confidence > 0.6 and opp.risk_level != "high"
 ]

 # Create execution order (low-risk first, then high-impact)
 execution_order = self._create_execution_order(safe_opportunities)

 # Estimate total time
 total_time = self._estimate_refactoring_time(safe_opportunities)

 # Assess overall risk
 risk_assessment = self._assess_overall_risk(safe_opportunities)

 # Identify prerequisites
 prerequisites = self._identify_prerequisites(safe_opportunities)

 # Create rollback strategy
 rollback_strategy = self._create_rollback_strategy(safe_opportunities)

 return RefactorPlan(
 opportunities=safe_opportunities,
 technical_debt_items=debt_items,
 execution_order=execution_order,
 estimated_time=total_time,
 risk_assessment=risk_assessment,
 prerequisites=prerequisites,
 rollback_strategy=rollback_strategy
 )

 def _create_execution_order(
 self, opportunities: List[RefactorOpportunity]
 ) -> List[int]:
 """Create optimal execution order for refactoring operations."""

 # Sort by: risk level (low first), then by confidence (high first), then by impact (high first)
 risk_order = {'low': 1, 'medium': 2, 'high': 3}

 sorted_opportunities = sorted(
 enumerate(opportunities),
 key=lambda x: (
 risk_order.get(x[1].risk_level, 3),
 -x[1].confidence,
 -x[1].complexity_reduction
 )
 )

 return [idx for idx, _ in sorted_opportunities]

 def _estimate_refactoring_time(
 self, opportunities: List[RefactorOpportunity]
 ) -> str:
 """Estimate total time required for refactoring."""

 total_minutes = 0

 for opp in opportunities:
 # Base time by type
 type_times = {
 RefactorType.EXTRACT_METHOD: 30,
 RefactorType.EXTRACT_VARIABLE: 10,
 RefactorType.REORGANIZE_IMPORTS: 15,
 RefactorType.INLINE_VARIABLE: 5,
 RefactorType.RENAME: 20,
 RefactorType.MOVE_MODULE: 45
 }

 base_time = type_times.get(opp.type, 20)

 # Adjust by risk level
 risk_multipliers = {'low': 1.0, 'medium': 1.5, 'high': 2.0}
 risk_multiplier = risk_multipliers.get(opp.risk_level, 1.0)

 # Adjust by complexity reduction
 complexity_multiplier = 1.0 + (opp.complexity_reduction * 0.5)

 total_minutes += base_time * risk_multiplier * complexity_multiplier

 # Convert to human-readable format
 if total_minutes < 60:
 return f"{int(total_minutes)} minutes"
 elif total_minutes < 480: # 8 hours
 hours = total_minutes / 60
 return f"{int(hours)} hours"
 else:
 days = total_minutes / 480
 return f"{int(days)} days"

 def _assess_overall_risk(
 self, opportunities: List[RefactorOpportunity]
 ) -> str:
 """Assess overall risk of refactoring plan."""

 if not opportunities:
 return "no_risk"

 high_risk_count = sum(1 for opp in opportunities if opp.risk_level == "high")
 medium_risk_count = sum(1 for opp in opportunities if opp.risk_level == "medium")

 if high_risk_count > 2:
 return "high"
 elif high_risk_count > 0 or medium_risk_count > 5:
 return "medium"
 else:
 return "low"

 def _identify_prerequisites(
 self, opportunities: List[RefactorOpportunity]
 ) -> List[str]:
 """Identify prerequisites for safe refactoring."""

 prerequisites = [
 "Create comprehensive test suite",
 "Ensure version control is properly configured",
 "Create backup of current codebase"
 ]

 # Add specific prerequisites based on opportunities
 if any(opp.type == RefactorType.MOVE_MODULE for opp in opportunities):
 prerequisites.append("Update import statements in dependent modules")

 if any(opp.type == RefactorType.EXTRACT_METHOD for opp in opportunities):
 prerequisites.append("Verify extracted methods maintain original functionality")

 return prerequisites

 def _create_rollback_strategy(
 self, opportunities: List[RefactorOpportunity]
 ) -> str:
 """Create rollback strategy for refactoring operations."""

 return """
 Rollback Strategy:
 1. Create git commit before each major refactoring step
 2. Run full test suite after each operation
 3. Maintain detailed change log with timestamps
 4. Use git revert for individual operation rollbacks
 5. Automated tests to verify functionality preservation
 """

# Usage Examples
"""
# Initialize refactoring system
refactorer = AIRefactorer(context7_client=context7)

# Analyze and create refactoring plan
refactor_plan = await refactorer.refactor_with_intelligence(
 codebase_path="/project/src",
 refactor_options={
 'max_risk_level': 'medium',
 'include_tests': True,
 'focus_on': ['complexity', 'duplication']
 }
)

print(f"Found {len(refactor_plan.opportunities)} refactoring opportunities")
print(f"Estimated time: {refactor_plan.estimated_time}")
print(f"Risk assessment: {refactor_plan.risk_assessment}")

# Execute refactoring plan
for i, opp_index in enumerate(refactor_plan.execution_order):
 opportunity = refactor_plan.opportunities[opp_index]
 print(f"\nStep {i+1}: {opportunity.description}")
 print(f"Type: {opportunity.type.value}")
 print(f"Risk: {opportunity.risk_level}")
 print(f"Confidence: {opportunity.confidence}")

 # Here you would implement the actual refactoring using Rope
 # This is a simplified example
 if opportunity.type == RefactorType.EXTRACT_METHOD:
 print("Would extract method using Rope...")
 elif opportunity.type == RefactorType.REORGANIZE_IMPORTS:
 print("Would reorganize imports using Rope...")

# After refactoring, verify with tests
print("\nRunning tests to verify refactoring...")
# test_results = run_test_suite()
# print(f"Tests passed: {test_results.passed}/{test_results.total}")
"""
```

## Advanced Features

### Context-Aware Refactoring

Intelligent Refactoring Context:
```python
class ContextAwareRefactorer(AIRefactorer):
 """Refactorer that considers project context and conventions."""

 def __init__(self, context7_client=None):
 super().__init__(context7_client)
 self.project_conventions = {}
 self.api_boundaries = set()

 async def analyze_project_context(self, codebase_path: str):
 """Analyze project-specific context and conventions."""

 # Detect naming conventions
 await self._detect_naming_conventions(codebase_path)

 # Identify API boundaries
 await self._identify_api_boundaries(codebase_path)

 # Analyze architectural patterns
 await self._analyze_architecture_patterns(codebase_path)

 async def _detect_naming_conventions(self, codebase_path: str):
 """Detect project-specific naming conventions."""

 naming_patterns = {
 'variable_names': [],
 'function_names': [],
 'class_names': [],
 'constant_names': []
 }

 python_files = self._find_python_files(codebase_path)

 for file_path in python_files[:50]: # Sample files for analysis
 try:
 with open(file_path, 'r', encoding='utf-8') as f:
 content = f.read()

 tree = ast.parse(content)

 class NamingConventionVisitor(ast.NodeVisitor):
 def visit_Name(self, node):
 if isinstance(node.ctx, ast.Store):
 if node.id.isupper():
 naming_patterns['constant_names'].append(node.id)
 else:
 naming_patterns['variable_names'].append(node.id)
 self.generic_visit(node)

 def visit_FunctionDef(self, node):
 naming_patterns['function_names'].append(node.name)
 self.generic_visit(node)

 def visit_ClassDef(self, node):
 naming_patterns['class_names'].append(node.name)
 self.generic_visit(node)

 visitor = NamingConventionVisitor()
 visitor.visit(tree)

 except Exception as e:
 print(f"Error analyzing {file_path}: {e}")

 # Analyze patterns
 self.project_conventions = self._analyze_naming_patterns(naming_patterns)

 def _analyze_naming_patterns(self, patterns: Dict[str, List[str]]) -> Dict[str, Any]:
 """Analyze naming patterns to extract conventions."""

 conventions = {}

 # Analyze variable naming
 snake_case_vars = sum(1 for name in patterns['variable_names'] if '_' in name)
 camel_case_vars = sum(1 for name in patterns['variable_names'] if name[0].islower() and any(c.isupper() for c in name[1:]))

 if snake_case_vars > camel_case_vars:
 conventions['variable_naming'] = 'snake_case'
 else:
 conventions['variable_naming'] = 'camelCase'

 # Analyze function naming
 snake_case_funcs = sum(1 for name in patterns['function_names'] if '_' in name)
 camel_case_funcs = sum(1 for name in patterns['function_names'] if name[0].islower() and any(c.isupper() for c in name[1:]))

 if snake_case_funcs > camel_case_funcs:
 conventions['function_naming'] = 'snake_case'
 else:
 conventions['function_naming'] = 'camelCase'

 return conventions
```

## Best Practices

1. Incremental Refactoring: Apply changes incrementally with testing at each step
2. Test Coverage: Ensure comprehensive test coverage before major refactoring
3. Version Control: Commit changes before and after each major refactoring step
4. Documentation: Update documentation to reflect refactored code structure
5. Performance Monitoring: Monitor performance impact of refactoring changes

---

Module: `modules/smart-refactoring.md`
Related: [AI Debugging](./ai-debugging.md) | [Performance Optimization](./performance-optimization.md)
