# AI-Powered Refactoring Workflows

> Sub-module: AI-assisted refactoring workflows with Context7 integration
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: Python 3.8+, Context7 MCP, AST, Rope

## Overview

This module provides advanced AI-assisted refactoring workflows that leverage Context7 MCP for real-time access to latest refactoring patterns, best practices, and framework-specific guidance.

---

## Context-Aware Refactoring

### Project Convention Detection

Detect and respect project-specific conventions for intelligent refactoring:

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

        for file_path in python_files[:50]:  # Sample files for analysis
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
        camel_case_vars = sum(1 for name in patterns['variable_names'] 
                             if name[0].islower() and any(c.isupper() for c in name[1:]))

        if snake_case_vars > camel_case_vars:
            conventions['variable_naming'] = 'snake_case'
        else:
            conventions['variable_naming'] = 'camelCase'

        # Analyze function naming
        snake_case_funcs = sum(1 for name in patterns['function_names'] if '_' in name)
        camel_case_funcs = sum(1 for name in patterns['function_names'] 
                              if name[0].islower() and any(c.isupper() for c in name[1:]))

        if snake_case_funcs > camel_case_funcs:
            conventions['function_naming'] = 'snake_case'
        else:
            conventions['function_naming'] = 'camelCase'

        return conventions
```

### API Boundary Detection

Identify public vs internal APIs for safer refactoring:

```python
async def _identify_api_boundaries(self, codebase_path: str):
    """Identify which modules are public APIs vs internal."""
    
    python_files = self._find_python_files(codebase_path)
    
    for file_path in python_files:
        # Check if file is in public API location
        if 'api' in file_path.split('/') or 'public' in file_path.split('/'):
            self.api_boundaries.add(file_path)
        
        # Check for __all__ exports (indicates public API)
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            if '__all__' in content:
                self.api_boundaries.add(file_path)
                
        except Exception as e:
            print(f"Error checking {file_path}: {e}")
```

---

## Context7 Integration

### Real-Time Pattern Retrieval

Fetch latest refactoring patterns from Context7:

```python
async def get_context7_refactoring_patterns(self) -> Dict[str, Any]:
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
```

### Framework-Specific Patterns

Get refactoring patterns specific to your framework:

```python
async def get_framework_patterns(self, framework: str) -> Dict[str, Any]:
    """Get framework-specific refactoring patterns."""

    framework_patterns = {
        'django': {
            'library_id': '/django/django',
            'topic': 'Django best practices refactoring views models 2025'
        },
        'fastapi': {
            'library_id': '/fastapi/fastapi',
            'topic': 'FastAPI route refactoring dependency injection 2025'
        },
        'flask': {
            'library_id': '/pallets/flask',
            'topic': 'Flask blueprint refactoring patterns 2025'
        }
    }

    if framework not in framework_patterns:
        return {}

    config = framework_patterns[framework]
    
    try:
        patterns = await self.context7.get_library_docs(
            context7_library_id=config['library_id'],
            topic=config['topic'],
            tokens=5000
        )
        return patterns

    except Exception as e:
        print(f"Failed to get {framework} patterns: {e}")
        return {}
```

---

## Intelligent Refactoring Pipeline

### End-to-End Workflow

Complete AI-assisted refactoring pipeline:

```python
async def intelligent_refactoring_pipeline(
    self, 
    codebase_path: str,
    framework: str = None,
    max_risk_level: str = 'medium'
) -> RefactorPlan:
    """Complete AI-assisted refactoring pipeline."""

    # Step 1: Analyze project context
    await self.analyze_project_context(codebase_path)
    
    # Step 2: Analyze technical debt
    debt_items = await self.technical_debt_analyzer.analyze(codebase_path)
    
    # Step 3: Get Context7 patterns
    context7_patterns = await self.get_context7_refactoring_patterns()
    
    if framework:
        framework_patterns = await self.get_framework_patterns(framework)
        context7_patterns['framework'] = framework_patterns
    
    # Step 4: Identify opportunities with AI
    opportunities = await self._identify_refactor_opportunities(
        codebase_path, 
        debt_items, 
        context7_patterns
    )
    
    # Step 5: Filter by risk and conventions
    filtered_opportunities = self._filter_by_conventions_and_risk(
        opportunities,
        max_risk_level
    )
    
    # Step 6: Create safe execution plan
    refactor_plan = self._create_safe_refactor_plan(
        filtered_opportunities,
        debt_items,
        context7_patterns
    )
    
    return refactor_plan

def _filter_by_conventions_and_risk(
    self,
    opportunities: List[RefactorOpportunity],
    max_risk_level: str
) -> List[RefactorOpportunity]:
    """Filter opportunities by project conventions and risk tolerance."""

    filtered = []
    
    risk_order = {'low': 1, 'medium': 2, 'high': 3}
    max_risk_value = risk_order.get(max_risk_level, 2)

    for opp in opportunities:
        # Filter by risk level
        if risk_order.get(opp.risk_level, 3) > max_risk_value:
            continue
        
        # Filter by API boundaries (don't refactor public APIs lightly)
        if opp.file_path in self.api_boundaries and opp.risk_level != 'low':
            continue
        
        filtered.append(opp)

    return filtered
```

---

## Safe Refactoring Execution

### Pre-Refactoring Checklist

Verify conditions before starting refactoring:

```python
async def pre_refactoring_checklist(self, codebase_path: str) -> Dict[str, bool]:
    """Run pre-refactoring safety checks."""

    checks = {
        'has_tests': False,
        'tests_passing': False,
        'has_version_control': False,
        'has_backup': False,
        'coverage_sufficient': False
    }

    # Check for tests
    test_files = self._find_test_files(codebase_path)
    checks['has_tests'] = len(test_files) > 0

    # Run tests if they exist
    if checks['has_tests']:
        test_results = await self._run_tests(codebase_path)
        checks['tests_passing'] = test_results['passed'] == test_results['total']

    # Check for git
    checks['has_version_control'] = os.path.exists(
        os.path.join(codebase_path, '.git')
    )

    # Check test coverage
    if checks['has_tests']:
        coverage = await self._calculate_coverage(codebase_path)
        checks['coverage_sufficient'] = coverage >= 0.8

    return checks

async def _run_tests(self, codebase_path: str) -> Dict[str, int]:
    """Run test suite and return results."""
    
    try:
        import subprocess
        result = subprocess.run(
            ['pytest', '--tb=no', '-q'],
            cwd=codebase_path,
            capture_output=True,
            text=True,
            timeout=300
        )
        
        # Parse output
        output = result.stdout
        if 'passed' in output:
            parts = output.split()
            for i, part in enumerate(parts):
                if part == 'passed' and i > 0:
                    return {'total': int(parts[i-1].split('/')[1]), 'passed': int(parts[i-1].split('/')[0])}

    except Exception as e:
        print(f"Error running tests: {e}")

    return {'total': 0, 'passed': 0}

async def _calculate_coverage(self, codebase_path: str) -> float:
    """Calculate test coverage percentage."""
    
    try:
        import subprocess
        result = subprocess.run(
            ['pytest', '--cov=.', '--cov-report=term-missing'],
            cwd=codebase_path,
            capture_output=True,
            text=True,
            timeout=300
        )
        
        output = result.stdout
        if 'TOTAL' in output:
            for line in output.split('\n'):
                if 'TOTAL' in line:
                    parts = line.split()
                    if len(parts) >= 4 and parts[-1].endswith('%'):
                        return float(parts[-1].rstrip('%')) / 100

    except Exception as e:
        print(f"Error calculating coverage: {e}")

    return 0.0
```

### Incremental Refactoring Execution

Execute refactoring in safe, incremental steps:

```python
async def execute_refactoring_plan(
    self,
    refactor_plan: RefactorPlan,
    codebase_path: str
) -> Dict[str, Any]:
    """Execute refactoring plan with safety checks."""

    results = {
        'successful': [],
        'failed': [],
        'skipped': []
    }

    # Pre-refactoring checks
    checks = await self.pre_refactoring_checklist(codebase_path)
    
    if not all(checks.values()):
        print("Pre-refactoring checks failed:")
        for check, passed in checks.items():
            if not passed:
                print(f"  - {check}: FAILED")
        return results

    # Execute in order
    for i, opp_index in enumerate(refactor_plan.execution_order):
        opportunity = refactor_plan.opportunities[opp_index]
        
        print(f"\n[{i+1}/{len(refactor_plan.execution_order)}] {opportunity.description}")
        
        try:
            # Create git commit before operation
            await self._create_git_commit(
                codebase_path,
                f"Before refactoring: {opportunity.description}"
            )

            # Execute refactoring
            success = await self._execute_single_refactoring(
                opportunity,
                codebase_path
            )

            if success:
                # Run tests to verify
                test_results = await self._run_tests(codebase_path)
                
                if test_results['passed'] == test_results['total']:
                    results['successful'].append(opportunity)
                    
                    # Commit after successful refactoring
                    await self._create_git_commit(
                        codebase_path,
                        f"After refactoring: {opportunity.description}"
                    )
                else:
                    results['failed'].append(opportunity)
                    # Revert on test failure
                    await self._revert_git_commit(codebase_path)
            else:
                results['skipped'].append(opportunity)

        except Exception as e:
            print(f"Error executing refactoring: {e}")
            results['failed'].append(opportunity)
            await self._revert_git_commit(codebase_path)

    return results

async def _execute_single_refactoring(
    self,
    opportunity: RefactorOpportunity,
    codebase_path: str
) -> bool:
    """Execute a single refactoring operation."""

    try:
        if opportunity.type == RefactorType.EXTRACT_METHOD:
            return await self._execute_extract_method(opportunity, codebase_path)
        
        elif opportunity.type == RefactorType.REORGANIZE_IMPORTS:
            return await self._execute_reorganize_imports(opportunity, codebase_path)
        
        elif opportunity.type == RefactorType.RENAME:
            return await self._execute_rename(opportunity, codebase_path)
        
        # Add other refactoring types...
        
        else:
            print(f"Unsupported refactoring type: {opportunity.type}")
            return False

    except Exception as e:
        print(f"Error executing {opportunity.type}: {e}")
        return False

async def _create_git_commit(self, codebase_path: str, message: str):
    """Create git commit with message."""
    
    try:
        import subprocess
        subprocess.run(
            ['git', 'add', '.'],
            cwd=codebase_path,
            capture_output=True
        )
        subprocess.run(
            ['git', 'commit', '-m', message],
            cwd=codebase_path,
            capture_output=True
        )
    except Exception as e:
        print(f"Error creating git commit: {e}")

async def _revert_git_commit(self, codebase_path: str):
    """Revert the most recent git commit."""
    
    try:
        import subprocess
        subprocess.run(
            ['git', 'revert', '--no-commit', 'HEAD'],
            cwd=codebase_path,
            capture_output=True
        )
    except Exception as e:
        print(f"Error reverting git commit: {e}")
```

---

## Post-Refactoring Analysis

### Impact Assessment

Analyze the impact of refactoring changes:

```python
async def post_refactoring_analysis(
    self,
    codebase_path: str,
    execution_results: Dict[str, Any]
) -> Dict[str, Any]:
    """Analyze impact of refactoring changes."""

    analysis = {
        'debt_reduction': 0,
        'complexity_reduction': 0,
        'lines_changed': 0,
        'files_modified': len(execution_results['successful']),
        'test_results': {}
    }

    # Re-analyze technical debt
    new_debt_items = await self.technical_debt_analyzer.analyze(codebase_path)
    
    # Calculate debt reduction
    original_debt_count = 100  # Would need to track original count
    new_debt_count = len(new_debt_items)
    analysis['debt_reduction'] = (original_debt_count - new_debt_count) / original_debt_count

    # Run final tests
    final_test_results = await self._run_tests(codebase_path)
    analysis['test_results'] = final_test_results

    # Count lines changed (from git)
    try:
        import subprocess
        result = subprocess.run(
            ['git', 'diff', '--shortstat'],
            cwd=codebase_path,
            capture_output=True,
            text=True
        )
        output = result.stdout
        if 'file' in output:
            parts = output.split()
            for i, part in enumerate(parts):
                if part == 'insertion(+)' and i > 0:
                    analysis['lines_changed'] += int(parts[i-1])
                elif part == 'deletion(-)' and i > 0:
                    analysis['lines_changed'] += int(parts[i-1])

    except Exception as e:
        print(f"Error counting lines changed: {e}")

    return analysis
```

---

## Best Practices

1. Always run pre-refactoring checks before starting
2. Create git commits before each operation
3. Run tests after every change
4. Revert immediately on test failure
5. Document all changes in commit messages
6. Use Context7 for latest patterns and practices
7. Respect project conventions when refactoring
8. Prioritize low-risk, high-impact changes first
9. Monitor performance impact of changes
10. Keep refactoring sessions short and focused

---

## Resources

### Context7 MCP Libraries

- Python Rope: `/python-rope/rope`
- Refactoring Guru: `/refactoring/guru`
- Django: `/django/django`
- FastAPI: `/fastapi/fastapi`
- Flask: `/pallets/flask`

### Tools

- Rope: Python refactoring library
- Git: Version control and rollback
- Pytest: Testing framework
- Coverage.py: Test coverage measurement

---

Sub-module: `modules/refactoring/ai-workflows.md`
Related: [patterns.md](./patterns.md) | [../smart-refactoring.md](../smart-refactoring.md)
