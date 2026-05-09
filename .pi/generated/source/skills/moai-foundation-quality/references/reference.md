# Enterprise Code Quality Reference

Complete API reference and technical documentation for the moai-foundation-quality skill.

---

## Table of Contents

1. [Core Classes](#core-classes)
2. [TRUST 5 Framework](#trust-5-framework)
3. [Quality Analysis APIs](#quality-analysis-apis)
4. [Proactive Analysis](#proactive-analysis)
5. [Best Practices Engine](#best-practices-engine)
6. [Configuration Reference](#configuration-reference)
7. [Context7 Integration](#context7-integration)
8. [Quality Metrics](#quality-metrics)
9. [Integration APIs](#integration-apis)
10. [Error Handling](#error-handling)

---

## Core Classes

### QualityOrchestrator

Main orchestrator class for enterprise code quality management.

```python
class QualityOrchestrator:
 """Enterprise quality orchestration engine"""

 def __init__(self, config: Optional[QualityConfig] = None):
 self.config = config or QualityConfig.default()
 self.trust5_validator = TRUST5Validator()
 self.proactive_scanner = ProactiveScanner()
 self.best_practices_engine = BestPracticesEngine()
 self.context7_client = Context7Client()
 self.metrics_collector = QualityMetricsCollector()

 async def analyze_codebase(
 self,
 path: str,
 languages: List[str],
 quality_threshold: float = 0.85,
 focus_areas: Optional[List[str]] = None
 ) -> QualityResult:
 """
 Comprehensive codebase quality analysis

 Args:
 path: Path to codebase directory
 languages: List of programming languages
 quality_threshold: Minimum acceptable quality score
 focus_areas: Specific areas to focus analysis on

 Returns:
 QualityResult: Comprehensive quality analysis results

 Example:
 orchestrator = QualityOrchestrator()
 result = await orchestrator.analyze_codebase(
 path="src/",
 languages=["python", "typescript"],
 quality_threshold=0.90
 )
 """

 async def generate_report(
 self,
 result: QualityResult,
 output_path: str,
 format: str = "html"
 ) -> None:
 """
 Generate comprehensive quality report

 Args:
 result: Quality analysis results
 output_path: Output file path
 format: Report format (html, pdf, json, sarif)
 """
```

### QualityConfig

Configuration class for quality analysis settings.

```python
@dataclass
class QualityConfig:
 """Configuration for quality analysis"""

 # Core settings
 trust5_enabled: bool = True
 proactive_analysis: bool = True
 best_practices_enforcement: bool = True
 context7_integration: bool = True

 # Quality thresholds
 quality_threshold: float = 0.85
 test_coverage_threshold: float = 0.85
 security_threshold: float = 0.90
 complexity_threshold: int = 10

 # Analysis settings
 languages: List[str] = field(default_factory=lambda: ["python", "javascript"])
 focus_areas: List[str] = field(default_factory=lambda: [
 "security", "performance", "maintainability", "testing"
 ])

 # Output settings
 generate_reports: bool = True
 output_formats: List[str] = field(default_factory=lambda: ["html", "json"])
 output_directory: str = ".moai/reports/quality"

 @classmethod
 def default(cls) -> "QualityConfig:
 """Create default configuration"""
 return cls()

 @classmethod
 def from_file(cls, config_path: str) -> "QualityConfig:
 """Load configuration from file"""
 with open(config_path) as f:
 data = yaml.safe_load(f)
 return cls(data)

 def to_dict(self) -> Dict:
 """Convert configuration to dictionary"""
 return asdict(self)
```

---

## TRUST 5 Framework

### TRUST5Validator

Comprehensive validation framework implementing TRUST 5 principles.

```python
class TRUST5Validator:
 """TRUST 5 quality framework validator"""

 VALIDATORS = {
 "testable": TestableValidator(),
 "readable": ReadableValidator(),
 "unified": UnifiedValidator(),
 "secured": SecuredValidator(),
 "trackable": TrackableValidator()
 }

 async def validate(
 self,
 codebase: str,
 thresholds: Optional[Dict[str, float]] = None
 ) -> TRUST5Result:
 """
 Validate codebase against TRUST 5 principles

 Args:
 codebase: Path to codebase
 thresholds: Custom thresholds for each principle

 Returns:
 TRUST5Result: Comprehensive validation results

 Example:
 validator = TRUST5Validator()
 result = await validator.validate(
 codebase="src/",
 thresholds={
 "testable": 0.90,
 "secured": 0.95,
 "overall": 0.85
 }
 )
 """

 async def validate_principle(
 self,
 principle: str,
 codebase: str,
 threshold: float
 ) -> ValidationResult:
 """Validate specific TRUST 5 principle"""
```

### Principle Validators

#### TestableValidator

```python
class TestableValidator:
 """Test-first principle validation"""

 async def validate(
 self,
 codebase: str,
 threshold: float = 0.85
 ) -> ValidationResult:
 """Validate test coverage and quality"""

 async def _analyze_test_coverage(self, codebase: str) -> CoverageResult:
 """Analyze test coverage metrics"""

 async def _analyze_test_quality(self, codebase: str) -> TestQualityResult:
 """Analyze test quality and effectiveness"""

 async def _validate_test_structure(self, codebase: str) -> StructureResult:
 """Validate test structure and organization"""
```

#### SecuredValidator

```python
class SecuredValidator:
 """Security principle validation with OWASP compliance"""

 async def validate(
 self,
 codebase: str,
 threshold: float = 0.90
 ) -> ValidationResult:
 """Validate security compliance"""

 async def _validate_owasp_compliance(self, codebase: str) -> OWASPResult:
 """Validate OWASP Top 10 compliance"""

 async def _scan_dependency_vulnerabilities(self, codebase: str) -> DependencyResult:
 """Scan for dependency vulnerabilities"""

 async def _analyze_code_security(self, codebase: str) -> CodeSecurityResult:
 """Analyze code-level security patterns"""

 def _calculate_security_level(self, score: float) -> str:
 """Calculate security level based on score"""
 if score >= 0.95:
 return "excellent"
 elif score >= 0.85:
 return "good"
 elif score >= 0.70:
 return "adequate"
 else:
 return "inadequate"
```

---

## Quality Analysis APIs

### ProactiveScanner

```python
class ProactiveScanner:
 """Proactive code quality issue detection"""

 SCAN_TYPES = [
 "performance",
 "maintainability",
 "security",
 "duplication",
 "technical_debt",
 "complexity"
 ]

 async def scan(
 self,
 codebase: str,
 focus_areas: List[str],
 severity_filter: Optional[str] = None
 ) -> ProactiveResult:
 """
 Scan codebase for quality issues

 Args:
 codebase: Path to codebase
 focus_areas: Areas to focus on
 severity_filter: Filter by severity level

 Returns:
 ProactiveResult: Scan results with recommendations
 """

 async def scan_performance_issues(self, codebase: str) -> PerformanceResult:
 """Scan for performance-related issues"""

 async def scan_maintainability_issues(self, codebase: str) -> MaintainabilityResult:
 """Scan for maintainability issues"""

 async def scan_security_issues(self, codebase: str) -> SecurityScanResult:
 """Scan for security vulnerabilities"""

 async def scan_code_duplication(self, codebase: str) -> DuplicationResult:
 """Scan for code duplication"""

 async def analyze_technical_debt(self, codebase: str) -> TechnicalDebtResult:
 """Analyze technical debt"""

 async def analyze_complexity(self, codebase: str) -> ComplexityResult:
 """Analyze code complexity"""
```

### QualityMetricsCollector

```python
class QualityMetricsCollector:
 """Comprehensive quality metrics collection"""

 METRIC_CATEGORIES = [
 "code_quality",
 "test_metrics",
 "security_metrics",
 "performance_metrics",
 "maintainability_metrics"
 ]

 async def collect_comprehensive_metrics(
 self,
 codebase: str,
 languages: List[str]
 ) -> QualityMetrics:
 """Collect all quality metrics"""

 async def collect_code_quality_metrics(
 self,
 codebase: str
 ) -> CodeQualityMetrics:
 """Collect code quality specific metrics"""

 async def collect_test_metrics(self, codebase: str) -> TestMetrics:
 """Collect testing metrics"""

 async def collect_security_metrics(self, codebase: str) -> SecurityMetrics:
 """Collect security metrics"""

 def calculate_quality_trend(
 self,
 historical_metrics: List[QualityMetrics]
 ) -> QualityTrend:
 """Calculate quality trends from historical data"""
```

---

## Best Practices Engine

### BestPracticesEngine

```python
class BestPracticesEngine:
 """Context7-powered best practices validation"""

 def __init__(self, context7_client: Context7Client):
 self.context7_client = context7_client
 self.language_rules = self._load_language_rules()
 self.practice_validators = self._initialize_validators()

 async def validate(
 self,
 codebase: str,
 languages: List[str],
 context7_docs: bool = True
 ) -> PracticesResult:
 """
 Validate coding best practices

 Args:
 codebase: Path to codebase
 languages: List of programming languages
 context7_docs: Whether to use Context7 documentation

 Returns:
 PracticesResult: Best practices validation results
 """

 async def validate_language_practices(
 self,
 codebase: str,
 language: str
 ) -> LanguageValidationResult:
 """Validate language-specific best practices"""

 async def validate_cross_language_practices(
 self,
 codebase: str
 ) -> CrossLanguageResult:
 """Validate cross-language consistency"""

 async def _validate_against_latest_standards(
 self,
 codebase: str,
 language: str,
 latest_docs: str
 ) -> LanguageValidationResult:
 """Validate against latest standards from Context7"""
```

### Language-Specific Validators

```python
class PythonPracticesValidator:
 """Python-specific best practices validation"""

 PRACTICES = [
 "pep8_compliance",
 "type_hints",
 "docstring_conventions",
 "import_organization",
 "exception_handling",
 "testing_patterns"
 ]

 async def validate_naming_conventions(
 self,
 codebase: str
 ) -> NamingValidationResult:
 """Validate Python naming conventions"""

 async def validate_docstring_standards(
 self,
 codebase: str
 ) -> DocstringValidationResult:
 """Validate docstring standards"""

 async def validate_type_hints_usage(
 self,
 codebase: str
 ) -> TypeHintsValidationResult:
 """Validate type hints usage"""

class TypeScriptPracticesValidator:
 """TypeScript-specific best practices validation"""

 PRACTICES = [
 "typescript_strict_mode",
 "interface_definitions",
 "error_handling",
 "async_patterns",
 "module_organization"
 ]

 async def validate_typescript_practices(
 self,
 codebase: str
 ) -> TypeScriptValidationResult:
 """Validate TypeScript best practices"""
```

---

## Configuration Reference

### Quality Thresholds

```python
QUALITY_THRESHOLDS = {
 "overall_quality": {
 "excellent": 0.90,
 "good": 0.80,
 "adequate": 0.70,
 "minimum": 0.60
 },
 "test_coverage": {
 "excellent": 0.95,
 "good": 0.85,
 "adequate": 0.75,
 "minimum": 0.65
 },
 "security_score": {
 "excellent": 0.95,
 "good": 0.85,
 "adequate": 0.75,
 "minimum": 0.65
 },
 "maintainability": {
 "excellent": 0.90,
 "good": 0.80,
 "adequate": 0.70,
 "minimum": 0.60
 }
}
```

### Language Configurations

```python
LANGUAGE_CONFIGURATIONS = {
 "python": {
 "style_guide": "pep8",
 "linter": "ruff",
 "formatter": "black",
 "type_checker": "mypy",
 "test_framework": "pytest",
 "security_tools": ["bandit", "safety"],
 "extensions": [".py"],
 "ignore_patterns": ["__pycache__", "*.pyc", ".venv"]
 },
 "typescript": {
 "style_guide": "google",
 "linter": "eslint",
 "formatter": "prettier",
 "type_checker": "tsc",
 "test_framework": "jest",
 "security_tools": ["tslint-security"],
 "extensions": [".ts", ".tsx"],
 "ignore_patterns": ["node_modules", "dist", "build"]
 },
 "javascript": {
 "style_guide": "airbnb",
 "linter": "eslint",
 "formatter": "prettier",
 "test_framework": "jest",
 "security_tools": ["eslint-plugin-security"],
 "extensions": [".js", ".jsx"],
 "ignore_patterns": ["node_modules", "dist", "build"]
 },
 "go": {
 "style_guide": "gofmt",
 "linter": "golangci-lint",
 "formatter": "gofmt",
 "test_framework": "go test",
 "security_tools": ["gosec"],
 "extensions": [".go"],
 "ignore_patterns": ["vendor"]
 }
}
```

### Analysis Focus Areas

```python
FOCUS_AREA_CONFIGURATIONS = {
 "security": {
 "enabled": True,
 "severity_threshold": "medium",
 "tools": ["bandit", "safety", "snyk"],
 "owasp_categories": [
 "A01: Broken Access Control",
 "A02: Cryptographic Failures",
 "A03: Injection",
 "A04: Insecure Design",
 "A05: Security Misconfiguration"
 ]
 },
 "performance": {
 "enabled": True,
 "severity_threshold": "medium",
 "analysis_types": [
 "algorithm_complexity",
 "database_queries",
 "memory_usage",
 "concurrency_issues"
 ]
 },
 "maintainability": {
 "enabled": True,
 "severity_threshold": "low",
 "metrics": [
 "cyclomatic_complexity",
 "code_duplication",
 "method_length",
 "class_size"
 ]
 },
 "testing": {
 "enabled": True,
 "severity_threshold": "medium",
 "requirements": [
 "unit_test_coverage",
 "integration_tests",
 "test_quality",
 "test_structure"
 ]
 }
}
```

---

## Context7 Integration

### Context7Client

```python
class Context7Client:
 """Context7 MCP client for real-time documentation access"""

 def __init__(self, cache_ttl: int = 3600):
 self.cache = {}
 self.cache_ttl = cache_ttl
 self.client = None

 async def resolve_library_id(
 self,
 library_name: str
 ) -> str:
 """
 Resolve library name to Context7 ID

 Args:
 library_name: Name of the library

 Returns:
 str: Context7-compatible library ID

 Example:
 client = Context7Client()
 library_id = await client.resolve_library_id("react")
 # Returns: "/facebook/react"
 """

 async def get_library_docs(
 self,
 context7CompatibleLibraryID: str,
 topic: str = "best-practices",
 tokens: int = 5000
 ) -> str:
 """
 Get latest documentation from Context7

 Args:
 context7CompatibleLibraryID: Library ID from resolve_library_id
 topic: Specific topic to focus on
 tokens: Maximum tokens to retrieve

 Returns:
 str: Documentation content
 """

 async def get_best_practices(
 self,
 language: str,
 framework: Optional[str] = None
 ) -> str:
 """Get latest best practices for language/framework"""

 async def get_security_guidelines(
 self,
 language: str
 ) -> str:
 """Get latest security guidelines"""

 async def get_performance_patterns(
 self,
 language: str
 ) -> str:
 """Get latest performance optimization patterns"""
```

### Library Mappings

```python
CONTEXT7_LIBRARY_MAPPINGS = {
 # Frontend Libraries
 "react": "/facebook/react",
 "vue": "/vuejs/vue",
 "angular": "/angular/angular",
 "typescript": "/microsoft/TypeScript",
 "javascript": "/nodejs/node",
 "eslint": "/eslint/eslint",
 "prettier": "/prettier/prettier",

 # Backend Libraries
 "python": "/python/cpython",
 "fastapi": "/tiangolo/fastapi",
 "django": "/django/django",
 "flask": "/pallets/flask",
 "express": "/expressjs/express",
 "nestjs": "/nestjs/nest",

 # Database Libraries
 "postgresql": "/postgresql/postgres",
 "mongodb": "/mongodb/mongo",
 "redis": "/redis/redis",
 "prisma": "/prisma/prisma",
 "sequelize": "/sequelize/sequelize",

 # Testing Libraries
 "jest": "/facebook/jest",
 "pytest": "/pytest-dev/pytest",
 "mocha": "/mochajs/mocha",
 "junit": "/junit-team/junit5",
 "cypress": "/cypress-io/cypress",

 # Security Libraries
 "owasp": "/owasp/owasp-top-ten",
 "bandit": "/PyCQA/bandit",
 "snyk": "/snyk/snyk",
 "sonarqube": "/SonarSource/sonarqube",

 # Development Tools
 "black": "/psf/black",
 "ruff": "/astral-sh/ruff",
 "mypy": "/python/mypy",
 "pylint": "/pylint-dev/pylint",
 "golangci-lint": "/golangci/golangci-lint"
}
```

---

## Quality Metrics

### Core Quality Metrics

```python
@dataclass
class QualityMetrics:
 """Comprehensive quality metrics"""

 # Overall metrics
 overall_score: float
 quality_grade: str # A, B, C, D, F
 trend_direction: str # improving, declining, stable

 # TRUST 5 metrics
 testable_score: float
 readable_score: float
 unified_score: float
 secured_score: float
 trackable_score: float

 # Coverage metrics
 test_coverage: float
 branch_coverage: float
 line_coverage: float

 # Complexity metrics
 cyclomatic_complexity: float
 cognitive_complexity: float
 maintainability_index: float

 # Security metrics
 security_score: float
 vulnerability_count: int
 critical_vulnerabilities: int

 # Performance metrics
 performance_score: float
 performance_issues: int
 bottlenecks: List[str]

 # Technical debt metrics
 technical_debt_hours: float
 debt_ratio: float
 debt_interest_rate: float

 # Code metrics
 lines_of_code: int
 duplicated_lines: int
 code_duplication_percentage: float

 def calculate_quality_grade(self) -> str:
 """Calculate letter grade from overall score"""
 if self.overall_score >= 0.90:
 return "A"
 elif self.overall_score >= 0.80:
 return "B"
 elif self.overall_score >= 0.70:
 return "C"
 elif self.overall_score >= 0.60:
 return "D"
 else:
 return "F"

 def is_production_ready(self) -> bool:
 """Determine if code is production ready"""
 return (
 self.overall_score >= 0.85 and
 self.test_coverage >= 0.80 and
 self.critical_vulnerabilities == 0 and
 self.security_score >= 0.85
 )
```

### Detailed Metrics Breakdown

```python
@dataclass
class TestMetrics:
 """Detailed testing metrics"""

 # Coverage metrics
 line_coverage: float
 branch_coverage: float
 function_coverage: float
 statement_coverage: float

 # Test quality metrics
 test_count: int
 passing_tests: int
 failing_tests: int
 skipped_tests: int
 test_success_rate: float

 # Test type distribution
 unit_test_count: int
 integration_test_count: int
 end_to_end_test_count: int
 performance_test_count: int

 # Test complexity
 average_test_length: float
 complex_tests: int
 flaky_tests: int

@dataclass
class SecurityMetrics:
 """Detailed security metrics"""

 # Vulnerability metrics
 total_vulnerabilities: int
 critical_vulnerabilities: int
 high_vulnerabilities: int
 medium_vulnerabilities: int
 low_vulnerabilities: int

 # OWASP Top 10 compliance
 owasp_compliance_score: float
 owasp_violations: Dict[str, int]

 # Security practices
 secure_coding_score: float
 dependency_vulnerabilities: int
 hardcoded_secrets: int
 insecure_configs: int

@dataclass
class PerformanceMetrics:
 """Detailed performance metrics"""

 # Algorithm performance
 algorithm_efficiency_score: float
 time_complexity_violations: int
 space_complexity_violations: int

 # Database performance
 query_efficiency_score: float
 n_plus_one_queries: int
 missing_indexes: int
 slow_queries: int

 # Memory performance
 memory_efficiency_score: float
 memory_leaks: int
 high_memory_usage: int

 # Concurrency performance
 concurrency_score: float
 race_conditions: int
 deadlocks: int
 resource_contention: int
```

---

## Integration APIs

### CI/CD Integration

```python
class CIIntegration:
 """CI/CD pipeline integration"""

 async def generate_pipeline_outputs(
 self,
 result: QualityResult,
 output_dir: str,
 formats: List[str]
 ) -> None:
 """Generate outputs for CI/CD pipelines"""

 async def generate_junit_xml(
 self,
 result: QualityResult,
 output_path: str
 ) -> None:
 """Generate JUnit XML for test results"""

 async def generate_sarif_report(
 self,
 result: QualityResult,
 output_path: str
 ) -> None:
 """Generate SARIF report for GitHub Actions"""

 async def generate_quality_badge(
 self,
 score: float,
 output_path: str
 ) -> None:
 """Generate quality badge for README"""

 def get_exit_code(
 self,
 result: QualityResult,
 threshold: float
 ) -> int:
 """Determine appropriate exit code for CI/CD"""
```

### REST API Integration

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

class QualityAnalysisRequest(BaseModel):
 repository_url: str
 branch: str = "main"
 languages: List[str]
 quality_threshold: float = 0.85
 focus_areas: List[str] = []

class QualityAnalysisResponse(BaseModel):
 analysis_id: str
 overall_score: float
 quality_grade: str
 trust5_validation: Dict[str, Any]
 recommendations: List[Dict[str, Any]]
 analysis_duration: float
 timestamp: datetime

app = FastAPI(title="Code Quality Analysis API")

@app.post("/api/quality/analyze", response_model=QualityAnalysisResponse)
async def analyze_quality(request: QualityAnalysisRequest):
 """API endpoint for quality analysis"""

@app.get("/api/quality/result/{analysis_id}")
async def get_analysis_result(analysis_id: str):
 """Get analysis result by ID"""

@app.post("/api/quality/webhook")
async def quality_webhook(payload: Dict):
 """Handle quality alerts via webhook"""
```

### Dashboard Integration

```python
class QualityDashboard:
 """Quality metrics dashboard integration"""

 async def generate_dashboard_data(
 self,
 historical_data: List[QualityMetrics]
 ) -> DashboardData:
 """Generate data for quality dashboard"""

 async def create_executive_summary(
 self,
 results: List[QualityResult]
 ) -> ExecutiveSummary:
 """Create executive summary of quality metrics"""

 async def generate_trend_analysis(
 self,
 time_range: str,
 project_filter: Optional[str] = None
 ) -> TrendAnalysis:
 """Generate quality trend analysis"""

 def create_alert_config(
 self,
 thresholds: Dict[str, float],
 notification_channels: List[str]
 ) -> AlertConfig:
 """Create alert configuration for monitoring"""
```

---

## Error Handling

### Exception Classes

```python
class QualityAnalysisError(Exception):
 """Base exception for quality analysis errors"""
 pass

class Context7IntegrationError(QualityAnalysisError):
 """Context7 integration errors"""
 pass

class QualityThresholdError(QualityAnalysisError):
 """Quality threshold validation errors"""
 pass

class ValidationError(QualityAnalysisError):
 """Validation process errors"""
 pass

class MetricsCollectionError(QualityAnalysisError):
 """Metrics collection errors"""
 pass

class ReportGenerationError(QualityAnalysisError):
 """Report generation errors"""
 pass
```

### Error Recovery

```python
class ErrorRecoveryManager:
 """Error recovery and retry management"""

 def __init__(self, max_retries: int = 3):
 self.max_retries = max_retries
 self.retry_strategies = {
 "context7_timeout": self._retry_context7_request,
 "file_access_error": self._retry_file_operation,
 "network_error": self._retry_network_request,
 "validation_error": self._retry_validation
 }

 async def handle_error(
 self,
 error: Exception,
 context: Dict[str, Any]
 ) -> ErrorRecoveryResult:
 """Handle and recover from errors"""

 async def _retry_context7_request(
 self,
 request_func: Callable,
 *args,
 kwargs
 ) -> Any:
 """Retry Context7 request with exponential backoff"""

 async def _retry_file_operation(
 self,
 file_operation: Callable,
 *args,
 kwargs
 ) -> Any:
 """Retry file operation"""

 def create_fallback_strategy(
 self,
 primary_operation: str
 ) -> Callable:
 """Create fallback strategy for failed operations"""
```

### Logging and Monitoring

```python
class QualityAnalysisLogger:
 """Structured logging for quality analysis"""

 def __init__(self, log_level: str = "INFO"):
 self.logger = self._setup_logger(log_level)

 def log_analysis_start(
 self,
 codebase: str,
 languages: List[str],
 config: QualityConfig
 ) -> None:
 """Log analysis start"""

 def log_analysis_completion(
 self,
 result: QualityResult,
 duration: float
 ) -> None:
 """Log analysis completion"""

 def log_error(
 self,
 error: Exception,
 context: Dict[str, Any]
 ) -> None:
 """Log error with context"""

 def log_performance_metrics(
 self,
 metrics: Dict[str, float]
 ) -> None:
 """Log performance metrics"""

 def log_quality_trend(
 self,
 current_score: float,
 previous_score: float,
 time_period: str
 ) -> None:
 """Log quality trend information"""
```

---

## Data Models

### Result Models

```python
@dataclass
class QualityResult:
 """Comprehensive quality analysis result"""

 trust5_validation: TRUST5Result
 proactive_analysis: ProactiveResult
 best_practices: PracticesResult
 metrics: QualityMetrics
 overall_score: float
 analysis_metadata: AnalysisMetadata

 def is_passed(self, threshold: float = 0.85) -> bool:
 """Check if quality gate passed"""

 def get_recommendations(
 self,
 severity_filter: Optional[str] = None,
 limit: int = 10
 ) -> List[Recommendation]:
 """Get prioritized recommendations"""

@dataclass
class TRUST5Result:
 """TRUST 5 validation result"""

 principles: Dict[str, ValidationResult]
 overall_score: float
 passed: bool
 recommendations: List[str]

@dataclass
class ValidationResult:
 """Individual validation result"""

 score: float
 passed: bool
 details: Dict[str, Any]
 recommendations: List[str]
 severity: str

@dataclass
class ProactiveResult:
 """Proactive analysis result"""

 scan_results: Dict[str, Any]
 recommendations: List[Recommendation]
 priority_issues: List[PriorityIssue]
 estimated_effort: float
```

### Recommendation Models

```python
@dataclass
class Recommendation:
 """Quality improvement recommendation"""

 title: str
 description: str
 category: str
 severity: str
 priority: str
 estimated_effort: float
 estimated_impact: float
 roi_score: float
 files_affected: List[str]
 implementation_steps: List[str]
 prerequisites: List[str]
 references: List[str]

 def to_dict(self) -> Dict[str, Any]:
 """Convert to dictionary for JSON serialization"""

@dataclass
class PriorityIssue:
 """High-priority quality issue"""

 title: str
 file: str
 line: int
 severity: str
 category: str
 description: str
 recommended_fix: str
 estimated_effort: float
 security_impact: bool
 business_impact: str
```

---

## Performance Optimization

### Caching Strategies

```python
class QualityAnalysisCache:
 """Caching for quality analysis results"""

 def __init__(self, cache_backend: str = "memory"):
 self.cache_backend = cache_backend
 self.cache = self._initialize_cache()

 async def get_cached_result(
 self,
 cache_key: str
 ) -> Optional[QualityResult]:
 """Get cached analysis result"""

 async def cache_result(
 self,
 cache_key: str,
 result: QualityResult,
 ttl: int = 3600
 ) -> None:
 """Cache analysis result"""

 def generate_cache_key(
 self,
 codebase: str,
 config: QualityConfig
 ) -> str:
 """Generate cache key for analysis"""

 async def invalidate_cache(
 self,
 pattern: str
 ) -> None:
 """Invalidate cache entries matching pattern"""
```

### Parallel Processing

```python
class ParallelQualityAnalyzer:
 """Parallel quality analysis for large codebases"""

 def __init__(self, max_workers: int = 4):
 self.max_workers = max_workers
 self.executor = ThreadPoolExecutor(max_workers=max_workers)

 async def analyze_parallel(
 self,
 codebase: str,
 config: QualityConfig
 ) -> QualityResult:
 """Analyze codebase in parallel"""

 async def _analyze_directory_parallel(
 self,
 directory: str,
 config: QualityConfig
 ) -> List[AnalysisResult]:
 """Analyze directory contents in parallel"""

 def _split_analysis_tasks(
 self,
 codebase: str,
 config: QualityConfig
 ) -> List[AnalysisTask]:
 """Split analysis into parallel tasks"""
```

This comprehensive reference documentation provides complete API coverage for the moai-foundation-quality skill, including all classes, methods, configuration options, and integration patterns needed for enterprise code quality management.
