# Automated Code Review - Core Classes

> Sub-module: Core class implementations for automated code review
> Parent: [Automated Code Review](../automated-code-review.md)

## Enumerations and Data Classes

### TrustCategory Enum

```python
class TrustCategory(Enum):
    """TRUST 5 framework categories."""
    TRUTHFULNESS = "truthfulness"  # Code correctness and logic accuracy
    RELEVANCE = "relevance"        # Code meets requirements and purpose
    USABILITY = "usability"        # Code is maintainable and understandable
    SAFETY = "safety"              # Code is secure and handles errors properly
    TIMELINESS = "timeliness"      # Code meets performance and delivery standards
```

### Severity Enum

```python
class Severity(Enum):
    """Issue severity levels."""
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    INFO = "info"
```

### IssueType Enum

```python
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
```

### CodeIssue Dataclass

```python
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
    confidence: float  # 0.0 to 1.0
    rule_violated: Optional[str] = None
    external_reference: Optional[str] = None
    auto_fixable: bool = False
    fix_diff: Optional[str] = None
```

### FileReviewResult Dataclass

```python
@dataclass
class FileReviewResult:
    """Review results for a single file."""
    file_path: str
    issues: List[CodeIssue]
    metrics: Dict[str, Any]
    trust_score: float  # 0.0 to 1.0
    category_scores: Dict[TrustCategory, float]
    lines_of_code: int
    complexity_metrics: Dict[str, float]
    review_timestamp: float
```

### CodeReviewReport Dataclass

```python
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
```

## AutomatedCodeReviewer Class

```python
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

        # Review each file
        file_results = []
        for file_path in files_to_review:
            file_result = await self.review_single_file(file_path)
            file_results.append(file_result)

        # Generate comprehensive report
        end_time = time.time()
        return self._generate_comprehensive_report(
            project_path, file_results, end_time - start_time
        )

    async def review_single_file(self, file_path: str) -> FileReviewResult:
        """Review a single Python file."""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except Exception as e:
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
            include_patterns = ['**/*.py']

        if exclude_patterns is None:
            exclude_patterns = [
                '/__pycache__/', '/venv/', '/env/',
                '/node_modules/', '/.git/', '/migrations/', '/tests/'
            ]

        from pathlib import Path
        import fnmatch

        project_root = Path(project_path)
        files = []

        for pattern in include_patterns:
            for file_path in project_root.glob(pattern):
                if file_path.is_file():
                    excluded = any(
                        fnmatch.fnmatch(str(file_path.relative_to(project_root)), ep)
                        for ep in exclude_patterns
                    )
                    if not excluded:
                        files.append(str(file_path))

        return sorted(files)
```

## Context7CodeAnalyzer Class

```python
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
            security_patterns = await self.context7.get_library_docs(
                context7_library_id="/security/semgrep",
                topic="security vulnerability detection patterns 2025",
                tokens=4000
            )
            self.security_patterns = security_patterns

            performance_patterns = await self.context7.get_library_docs(
                context7_library_id="/performance/python-profiling",
                topic="performance anti-patterns code analysis 2025",
                tokens=3000
            )
            self.performance_patterns = performance_patterns

            quality_patterns = await self.context7.get_library_docs(
                context7_library_id="/code-quality/sonarqube",
                topic="code quality best practices smells detection 2025",
                tokens=4000
            )

            return {
                'security': security_patterns,
                'performance': performance_patterns,
                'quality': quality_patterns
            }
        except Exception as e:
            print(f"Failed to load Context7 patterns: {e}")
            return self._get_default_analysis_patterns()

    def _get_default_analysis_patterns(self) -> Dict[str, Any]:
        """Get default analysis patterns when Context7 is unavailable."""
        return {
            'security': {
                'sql_injection': [r"execute\([^)]*\+[^)]*\)", r"format\s*\("],
                'command_injection': [r"os\.system\(", r"subprocess\.call\(", r"eval\("],
                'path_traversal': [r"open\([^)]*\+[^)]*\)", r"\.\.\/"]
            },
            'performance': {
                'inefficient_loops': [r"for.*in.*range\(len\(", r"while.*len\("],
                'memory_leaks': [r"global\s+", r"\.append\(.*\)\s*\.append\("]
            },
            'quality': {
                'long_functions': {'max_lines': 50},
                'complex_conditionals': {'max_complexity': 10},
                'deep_nesting': {'max_depth': 4}
            }
        }
```

## Related Sub-modules

- [Analysis Patterns](./analysis-patterns.md) - TRUST 5 analysis implementation
- [Tool Integration](./tool-integration.md) - Static analysis tool wrappers

---

Sub-module: `modules/code-review/core-classes.md`
