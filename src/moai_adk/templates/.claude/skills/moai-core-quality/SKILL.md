---
name: moai-core-quality
description: Enterprise code quality orchestrator with TRUST 5 validation, proactive analysis, and automated best practices enforcement
allowed-tools: Read, Write, Edit, Bash, Grep, Glob, TodoWrite, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
---

# Enterprise Code Quality Orchestrator 🔍

Enterprise-grade code quality management system that combines systematic code review, proactive improvement suggestions, and automated best practices enforcement. Provides comprehensive quality assurance through TRUST 5 framework validation with Context7 integration for real-time best practices.

## Quick Reference (30 seconds)

**Core Capabilities**:
- **TRUST 5 Validation**: Testable, Readable, Unified, Secured, Trackable quality gates
- **Proactive Analysis**: Automated issue detection and improvement suggestions
- **Best Practices Enforcement**: Context7-powered real-time standards validation
- **Multi-Language Support**: 25+ programming languages with specialized rules
- **Enterprise Integration**: CI/CD pipelines, quality metrics, reporting

**Key Patterns**:
1. **Quality Gate Pipeline** → Automated validation with configurable thresholds
2. **Proactive Scanner** → Continuous analysis with improvement recommendations
3. **Best Practices Engine** → Context7-driven standards enforcement
4. **Quality Metrics Dashboard** → Comprehensive reporting and trend analysis

**When to Use**:
- Code review automation and quality gate enforcement
- Proactive code quality improvement and technical debt reduction
- Enterprise coding standards enforcement and compliance validation
- CI/CD pipeline integration with automated quality checks

## Implementation Guide

### Getting Started

**Basic Quality Validation**:
```python
# Initialize quality orchestrator
quality_orchestrator = QualityOrchestrator(
    trust5_enabled=True,
    proactive_analysis=True,
    best_practices_enforcement=True,
    context7_integration=True
)

# Run comprehensive quality analysis
result = await quality_orchestrator.analyze_codebase(
    path="src/",
    languages=["python", "javascript", "typescript"],
    quality_threshold=0.85
)

# Quality gate validation with TRUST 5
quality_gate = QualityGate()
validation_result = await quality_gate.validate_trust5(
    codebase_path="src/",
    test_coverage_threshold=0.90,
    complexity_threshold=10
)
```

**Proactive Quality Analysis**:
```python
# Initialize proactive scanner
proactive_scanner = ProactiveQualityScanner(
    context7_client=context7_client,
    rule_engine=BestPracticesEngine()
)

# Scan for improvement opportunities
improvements = await proactive_scanner.scan_codebase(
    path="src/",
    scan_types=["security", "performance", "maintainability", "testing"]
)

# Generate improvement recommendations
recommendations = await proactive_scanner.generate_recommendations(
    issues=improvements,
    priority="high",
    auto_fix=True
)
```

### Core Components

#### 1. Quality Orchestration Engine

```python
class QualityOrchestrator:
    """Enterprise quality orchestration with TRUST 5 framework"""

    def __init__(self, config: QualityConfig):
        self.trust5_validator = TRUST5Validator()
        self.proactive_scanner = ProactiveScanner()
        self.best_practices_engine = BestPracticesEngine()
        self.context7_client = Context7Client()
        self.metrics_collector = QualityMetricsCollector()

    async def analyze_codebase(self, request: QualityAnalysisRequest) -> QualityResult:
        """Comprehensive codebase quality analysis"""

        # Phase 1: TRUST 5 Validation
        trust5_result = await self.trust5_validator.validate(
            codebase=request.path,
            thresholds=request.quality_thresholds
        )

        # Phase 2: Proactive Analysis
        proactive_result = await self.proactive_scanner.scan(
            codebase=request.path,
            focus_areas=request.focus_areas
        )

        # Phase 3: Best Practices Check
        practices_result = await self.best_practices_engine.validate(
            codebase=request.path,
            languages=request.languages,
            context7_docs=True
        )

        # Phase 4: Metrics Collection
        metrics = await self.metrics_collector.collect_comprehensive_metrics(
            codebase=request.path,
            analysis_results=[trust5_result, proactive_result, practices_result]
        )

        return QualityResult(
            trust5_validation=trust5_result,
            proactive_analysis=proactive_result,
            best_practices=practices_result,
            metrics=metrics,
            overall_score=self._calculate_overall_quality_score([
                trust5_result, proactive_result, practices_result
            ])
        )
```

#### 2. TRUST 5 Validation Framework

```python
class TRUST5Validator:
    """Comprehensive TRUST 5 quality framework validation"""

    VALIDATORS = {
        "testable": TestableValidator(),
        "readable": ReadableValidator(),
        "unified": UnifiedValidator(),
        "secured": SecuredValidator(),
        "trackable": TrackableValidator()
    }

    async def validate(self, codebase: str, thresholds: Dict[str, float]) -> TRUST5Result:
        """Execute complete TRUST 5 validation"""

        results = {}

        for principle, validator in self.VALIDATORS.items():
            result = await validator.validate(
                codebase=codebase,
                threshold=thresholds.get(principle, 0.8)
            )
            results[principle] = result

        # Calculate overall TRUST 5 score
        overall_score = sum(r.score for r in results.values()) / len(results)

        return TRUST5Result(
            principles=results,
            overall_score=overall_score,
            passed=overall_score >= thresholds.get("overall", 0.85),
            recommendations=self._generate_trust5_recommendations(results)
        )

class TestableValidator:
    """Test-first principle validation"""

    async def validate(self, codebase: str, threshold: float) -> ValidationResult:
        """Validate test coverage and quality"""

        # Check test coverage
        coverage_result = await self._analyze_test_coverage(codebase)

        # Validate test quality
        test_quality = await self._analyze_test_quality(codebase)

        # Check test structure
        test_structure = await self._validate_test_structure(codebase)

        score = (coverage_result.score * 0.5 +
                test_quality.score * 0.3 +
                test_structure.score * 0.2)

        return ValidationResult(
            score=score,
            passed=score >= threshold,
            details={
                "coverage": coverage_result,
                "quality": test_quality,
                "structure": test_structure
            },
            recommendations=self._generate_testing_recommendations(
                coverage_result, test_quality, test_structure
            )
        )

class SecuredValidator:
    """Security principle validation with OWASP compliance"""

    async def validate(self, codebase: str, threshold: float) -> ValidationResult:
        """Validate security compliance and vulnerabilities"""

        # OWASP Top 10 validation
        owasp_result = await self._validate_owasp_compliance(codebase)

        # Security best practices
        security_practices = await self._validate_security_practices(codebase)

        # Dependency vulnerability scan
        dependency_scan = await self._scan_dependency_vulnerabilities(codebase)

        # Code security patterns
        code_security = await self._analyze_code_security(codebase)

        score = (owasp_result.score * 0.4 +
                security_practices.score * 0.3 +
                dependency_scan.score * 0.2 +
                code_security.score * 0.1)

        return ValidationResult(
            score=score,
            passed=score >= threshold,
            details={
                "owasp": owasp_result,
                "practices": security_practices,
                "dependencies": dependency_scan,
                "code_patterns": code_security
            },
            security_level=self._calculate_security_level(score),
            recommendations=self._generate_security_recommendations(
                owasp_result, security_practices, dependency_scan, code_security
            )
        )
```

#### 3. Proactive Quality Scanner

```python
class ProactiveQualityScanner:
    """Proactive code quality issue detection and analysis"""

    def __init__(self, context7_client: Context7Client):
        self.context7_client = context7_client
        self.issue_detectors = self._initialize_detectors()
        self.pattern_analyzer = CodePatternAnalyzer()

    async def scan(self, codebase: str, focus_areas: List[str]) -> ProactiveResult:
        """Comprehensive proactive quality scanning"""

        scan_results = {}

        # Performance analysis
        if "performance" in focus_areas:
            scan_results["performance"] = await self._scan_performance_issues(codebase)

        # Maintainability analysis
        if "maintainability" in focus_areas:
            scan_results["maintainability"] = await self._scan_maintainability_issues(codebase)

        # Security vulnerabilities
        if "security" in focus_areas:
            scan_results["security"] = await self._scan_security_issues(codebase)

        # Code duplication
        if "duplication" in focus_areas:
            scan_results["duplication"] = await self._scan_code_duplication(codebase)

        # Technical debt
        if "technical_debt" in focus_areas:
            scan_results["technical_debt"] = await self._analyze_technical_debt(codebase)

        # Code complexity
        if "complexity" in focus_areas:
            scan_results["complexity"] = await self._analyze_complexity(codebase)

        # Generate improvement recommendations
        recommendations = await self._generate_improvement_recommendations(scan_results)

        return ProactiveResult(
            scan_results=scan_results,
            recommendations=recommendations,
            priority_issues=self._identify_priority_issues(scan_results),
            estimated_effort=self._calculate_improvement_effort(recommendations)
        )

    async def _scan_performance_issues(self, codebase: str) -> PerformanceResult:
        """Scan for performance-related issues"""

        issues = []

        # Get language-specific performance patterns from Context7
        for language in self._detect_languages(codebase):
            try:
                # Resolve library ID
                library_id = await self.context7_client.resolve_library_id(language)

                # Get performance best practices
                perf_docs = await self.context7_client.get_library_docs(
                    context7CompatibleLibraryID=library_id,
                    topic="performance",
                    tokens=3000
                )

                # Analyze code against performance patterns
                language_issues = await self._analyze_performance_patterns(
                    codebase, language, perf_docs
                )
                issues.extend(language_issues)

            except Exception as e:
                logger.warning(f"Failed to get performance docs for {language}: {e}")

        # Common performance issues
        common_issues = await self._detect_common_performance_issues(codebase)
        issues.extend(common_issues)

        return PerformanceResult(
            issues=issues,
            score=self._calculate_performance_score(issues),
            hotspots=self._identify_performance_hotspots(issues),
            optimizations=self._suggest_optimizations(issues)
        )
```

#### 4. Best Practices Engine with Context7

```python
class BestPracticesEngine:
    """Context7-powered best practices validation and enforcement"""

    def __init__(self, context7_client: Context7Client):
        self.context7_client = context7_client
        self.language_rules = self._load_language_rules()
        self.practice_validators = self._initialize_validators()

    async def validate(self, codebase: str, languages: List[str], context7_docs: bool = True) -> PracticesResult:
        """Validate coding best practices with real-time documentation"""

        validation_results = {}

        for language in languages:
            # Get latest best practices from Context7
            if context7_docs:
                try:
                    library_id = await self.context7_client.resolve_library_id(language)
                    latest_docs = await self.context7_client.get_library_docs(
                        context7CompatibleLibraryID=library_id,
                        topic="best-practices",
                        tokens=5000
                    )

                    # Validate against latest standards
                    validation_result = await self._validate_against_latest_standards(
                        codebase, language, latest_docs
                    )

                except Exception as e:
                    logger.warning(f"Failed to get Context7 docs for {language}: {e}")
                    # Fallback to cached rules
                    validation_result = await self._validate_with_cached_rules(
                        codebase, language
                    )
            else:
                validation_result = await self._validate_with_cached_rules(
                    codebase, language
                )

            validation_results[language] = validation_result

        # Cross-language best practices
        cross_language_result = await self._validate_cross_language_practices(codebase)

        # Calculate overall practices score
        overall_score = sum(
            result.score for result in validation_results.values()
        ) / len(validation_results)

        return PracticesResult(
            language_results=validation_results,
            cross_language_practices=cross_language_result,
            overall_score=overall_score,
            compliance_level=self._determine_compliance_level(overall_score),
            improvement_roadmap=self._create_improvement_roadmap(validation_results)
        )

    async def _validate_against_latest_standards(
        self,
        codebase: str,
        language: str,
        latest_docs: str
    ) -> LanguageValidationResult:
        """Validate code against latest language standards from Context7"""

        # Extract best practices from documentation
        best_practices = await self._extract_best_practices_from_docs(latest_docs)

        # Validate naming conventions
        naming_result = await self._validate_naming_conventions(
            codebase, language, best_practices.get("naming", {})
        )

        # Validate code structure
        structure_result = await self._validate_code_structure(
            codebase, language, best_practices.get("structure", {})
        )

        # Validate error handling
        error_handling_result = await self._validate_error_handling(
            codebase, language, best_practices.get("error_handling", {})
        )

        # Validate documentation
        documentation_result = await self._validate_documentation(
            codebase, language, best_practices.get("documentation", {})
        )

        # Validate testing patterns
        testing_result = await self._validate_testing_patterns(
            codebase, language, best_practices.get("testing", {})
        )

        # Calculate language-specific score
        language_score = (
            naming_result.score * 0.2 +
            structure_result.score * 0.3 +
            error_handling_result.score * 0.2 +
            documentation_result.score * 0.15 +
            testing_result.score * 0.15
        )

        return LanguageValidationResult(
            language=language,
            score=language_score,
            validations={
                "naming": naming_result,
                "structure": structure_result,
                "error_handling": error_handling_result,
                "documentation": documentation_result,
                "testing": testing_result
            },
            best_practices_version=await self._get_docs_version(latest_docs),
            recommendations=self._generate_language_recommendations(
                naming_result, structure_result, error_handling_result,
                documentation_result, testing_result
            )
        )
```

### Configuration and Customization

**Quality Configuration**:
```yaml
# quality-config.yaml
quality_orchestration:
  trust5_framework:
    enabled: true
    thresholds:
      overall: 0.85
      testable: 0.90
      readable: 0.80
      unified: 0.85
      secured: 0.90
      trackable: 0.80

  proactive_analysis:
    enabled: true
    scan_frequency: "daily"
    focus_areas:
      - "performance"
      - "security"
      - "maintainability"
      - "technical_debt"

    auto_fix:
      enabled: true
      severity_threshold: "medium"
      confirmation_required: true

  best_practices:
    enabled: true
    context7_integration: true
    auto_update_standards: true
    compliance_target: 0.85

    language_rules:
      python:
        style_guide: "pep8"
        formatter: "black"
        linter: "ruff"
        type_checker: "mypy"

      javascript:
        style_guide: "airbnb"
        formatter: "prettier"
        linter: "eslint"

      typescript:
        style_guide: "google"
        formatter: "prettier"
        linter: "eslint"

  reporting:
    enabled: true
    metrics_retention_days: 90
    trend_analysis: true
    executive_dashboard: true

    notifications:
      quality_degradation: true
      security_vulnerabilities: true
      technical_debt_increase: true
```

**Integration Examples**:
```python
# CI/CD Pipeline Integration
async def quality_gate_pipeline():
    """Integrate quality validation into CI/CD pipeline"""

    # Initialize quality orchestrator
    quality_orchestrator = QualityOrchestrator.from_config("quality-config.yaml")

    # Run comprehensive quality analysis
    quality_result = await quality_orchestrator.analyze_codebase(
        path="src/",
        languages=["python", "typescript"],
        quality_threshold=0.85
    )

    # Quality gate validation
    if not quality_result.trust5_validation.passed:
        print("❌ Quality gate failed!")
        print(f"Overall score: {quality_result.overall_score:.2f}")

        # Print failed principles
        for principle, result in quality_result.trust5_validation.principles.items():
            if not result.passed:
                print(f"  {principle}: {result.score:.2f} (threshold: 0.80)")

        # Exit with error code
        sys.exit(1)

    # Check for critical security issues
    critical_issues = [
        issue for issue in quality_result.proactive_analysis.recommendations
        if issue.severity == "critical" and issue.category == "security"
    ]

    if critical_issues:
        print(f"❌ Found {len(critical_issues)} critical security issues!")
        for issue in critical_issues:
            print(f"  - {issue.description}")
        sys.exit(1)

    print("✅ Quality gate passed!")
    print(f"Overall quality score: {quality_result.overall_score:.2f}")

    # Generate quality report
    await generate_quality_report(quality_result, output_path="quality-report.json")

# GitHub Actions Integration
async def github_actions_quality_check():
    """Quality check for GitHub Actions workflow"""

    # Parse inputs
    github_token = os.getenv("GITHUB_TOKEN")
    repo_path = os.getenv("GITHUB_WORKSPACE", ".")
    pr_number = os.getenv("PR_NUMBER")

    # Run quality analysis
    quality_orchestrator = QualityOrchestrator()
    quality_result = await quality_orchestrator.analyze_codebase(
        path=repo_path,
        languages=["python", "javascript", "typescript"]
    )

    # Post comment on PR if quality issues found
    if pr_number and quality_result.overall_score < 0.85:
        comment = generate_pr_quality_comment(quality_result)
        await post_github_comment(github_token, pr_number, comment)

    # Set output for GitHub Actions
    print(f"::set-output name=quality_score::{quality_result.overall_score}")
    print(f"::set-output name=quality_passed::{quality_result.trust5_validation.passed}")
```

## Advanced Patterns

### 1. Custom Quality Rules

```python
class CustomQualityRule:
    """Define custom quality validation rules"""

    def __init__(self, name: str, validator: Callable, severity: str = "medium"):
        self.name = name
        self.validator = validator
        self.severity = severity

    async def validate(self, codebase: str) -> RuleResult:
        """Execute custom rule validation"""
        try:
            result = await self.validator(codebase)
            return RuleResult(
                rule_name=self.name,
                passed=result.passed,
                severity=self.severity,
                details=result.details,
                recommendations=result.recommendations
            )
        except Exception as e:
            return RuleResult(
                rule_name=self.name,
                passed=False,
                severity="error",
                details={"error": str(e)},
                recommendations=["Fix rule implementation"]
            )

# Usage example
async def custom_naming_convention_rule(codebase: str):
    """Custom rule: Enforce project-specific naming conventions"""

    # Define project-specific patterns
    patterns = {
        "api_endpoints": r"^[a-z]+_[a-z]+$",  # get_user, create_order
        "database_models": r"^[A-Z][a-zA-Z]*Model$",  # UserModel, OrderModel
        "utility_functions": r"^util_[a-z_]+$"  # util_format_date, util_validate_email
    }

    violations = []

    for pattern_name, pattern in patterns.items():
        # Scan codebase for violations
        pattern_violations = await scan_for_pattern_violations(codebase, pattern, pattern_name)
        violations.extend(pattern_violations)

    return RuleValidationResult(
        passed=len(violations) == 0,
        details={"violations": violations, "total_violations": len(violations)},
        recommendations=[f"Fix {len(violations)} naming convention violations"]
    )

# Register custom rule
custom_rule = CustomQualityRule(
    name="project_naming_conventions",
    validator=custom_naming_convention_rule,
    severity="medium"
)

quality_orchestrator.register_custom_rule(custom_rule)
```

### 2. Machine Learning Quality Prediction

```python
class QualityPredictionEngine:
    """ML-powered quality issue prediction"""

    def __init__(self, model_path: str):
        self.model = self._load_model(model_path)
        self.feature_extractor = CodeFeatureExtractor()

    async def predict_quality_issues(self, codebase: str) -> PredictionResult:
        """Predict potential quality issues using ML"""

        # Extract code features
        features = await self.feature_extractor.extract_features(codebase)

        # Make predictions
        predictions = self.model.predict(features)

        # Analyze prediction confidence
        confidence_scores = self.model.predict_proba(features)

        # Group predictions by issue type
        issue_predictions = self._group_predictions_by_type(
            predictions, confidence_scores
        )

        return PredictionResult(
            predictions=issue_predictions,
            confidence_scores=confidence_scores,
            high_risk_areas=self._identify_high_risk_areas(issue_predictions),
            prevention_recommendations=self._generate_prevention_recommendations(
                issue_predictions
            )
        )

    def _group_predictions_by_type(
        self,
        predictions: np.ndarray,
        confidence_scores: np.ndarray
    ) -> Dict[str, List[Prediction]]:
        """Group ML predictions by issue type"""

        issue_types = ["security", "performance", "maintainability", "reliability"]
        grouped_predictions = {issue_type: [] for issue_type in issue_types}

        for i, prediction in enumerate(predictions):
            issue_type = issue_types[prediction]
            confidence = confidence_scores[i][prediction]

            if confidence > 0.7:  # High confidence predictions only
                grouped_predictions[issue_type].append(
                    Prediction(
                        issue_type=issue_type,
                        confidence=confidence,
                        location=self._get_prediction_location(i),
                        severity=self._determine_prediction_severity(confidence)
                    )
                )

        return grouped_predictions
```

### 3. Real-time Quality Monitoring

```python
class RealTimeQualityMonitor:
    """Real-time code quality monitoring and alerting"""

    def __init__(self, webhook_url: str, notification_config: Dict):
        self.webhook_url = webhook_url
        self.notification_config = notification_config
        self.quality_history = deque(maxlen=1000)
        self.alert_thresholds = notification_config.get("thresholds", {})

    async def monitor_quality_changes(self, codebase: str):
        """Continuously monitor quality changes"""

        while True:
            # Get current quality metrics
            current_metrics = await self._get_current_quality_metrics(codebase)

            # Compare with historical data
            if self.quality_history:
                previous_metrics = self.quality_history[-1]
                quality_change = self._calculate_quality_change(
                    previous_metrics, current_metrics
                )

                # Check for quality degradation
                if quality_change < -self.alert_thresholds.get("degradation", 0.1):
                    await self._send_quality_alert(
                        alert_type="quality_degradation",
                        metrics=current_metrics,
                        change=quality_change
                    )

                # Check for security vulnerabilities
                if current_metrics.security_score < self.alert_thresholds.get("security", 0.8):
                    await self._send_security_alert(
                        security_score=current_metrics.security_score,
                        vulnerabilities=current_metrics.security_issues
                    )

            # Store metrics
            self.quality_history.append(current_metrics)

            # Generate quality trend report
            if len(self.quality_history) % 10 == 0:
                await self._generate_trend_report()

            # Wait for next check
            await asyncio.sleep(self.notification_config.get("check_interval", 300))  # 5 minutes

    async def _send_quality_alert(self, alert_type: str, metrics: Dict, change: float):
        """Send quality alert via webhook"""

        alert_payload = {
            "alert_type": alert_type,
            "timestamp": datetime.utcnow().isoformat(),
            "metrics": metrics,
            "quality_change": change,
            "severity": self._determine_alert_severity(change)
        }

        async with aiohttp.ClientSession() as session:
            async with session.post(self.webhook_url, json=alert_payload) as response:
                if response.status == 200:
                    logger.info(f"Quality alert sent: {alert_type}")
                else:
                    logger.error(f"Failed to send quality alert: {response.status}")
```

### 4. Cross-Project Quality Benchmarking

```python
class QualityBenchmarking:
    """Cross-project quality benchmarking and comparison"""

    def __init__(self, benchmark_database: str):
        self.benchmark_db = benchmark_database
        self.comparison_metrics = [
            "code_coverage",
            "security_score",
            "maintainability_index",
            "technical_debt_ratio",
            "duplicate_code_percentage"
        ]

    async def benchmark_project(self, project_path: str, project_metadata: Dict) -> BenchmarkResult:
        """Benchmark project quality against similar projects"""

        # Analyze current project
        current_metrics = await self._analyze_project_quality(project_path)

        # Find comparable projects from database
        comparable_projects = await self._find_comparable_projects(project_metadata)

        # Calculate percentiles and rankings
        benchmark_comparison = await self._calculate_benchmark_comparison(
            current_metrics, comparable_projects
        )

        # Generate improvement recommendations based on top performers
        improvement_recommendations = await self._generate_benchmark_recommendations(
            current_metrics, benchmark_comparison
        )

        return BenchmarkResult(
            project_metrics=current_metrics,
            benchmark_comparison=benchmark_comparison,
            industry_percentiles=benchmark_comparison.percentiles,
            improvement_roadmap=improvement_recommendations,
            competitive_analysis=self._analyze_competitive_position(
                current_metrics, benchmark_comparison
            )
        )

    async def _find_comparable_projects(self, project_metadata: Dict) -> List[ProjectMetrics]:
        """Find projects with similar characteristics for comparison"""

        query = {
            "language": project_metadata.get("language"),
            "project_type": project_metadata.get("type", "web_application"),
            "team_size_range": self._get_team_size_range(project_metadata.get("team_size", 5)),
            "industry": project_metadata.get("industry", "technology")
        }

        # Query benchmark database
        comparable_projects = await self.benchmark_db.find_projects(query)

        return comparable_projects[:50]  # Limit to top 50 comparable projects
```

## Context7 Library Mappings

**Essential Mappings for Quality Analysis**:

```python
QUALITY_LIBRARY_MAPPINGS = {
    # Static Analysis Tools
    "eslint": "/eslint/eslint",
    "prettier": "/prettier/prettier",
    "black": "/psf/black",
    "ruff": "/astral-sh/ruff",
    "mypy": "/python/mypy",
    "pylint": "/pylint-dev/pylint",
    "sonarqube": "/SonarSource/sonarqube",

    # Testing Frameworks
    "jest": "/facebook/jest",
    "pytest": "/pytest-dev/pytest",
    "mocha": "/mochajs/mocha",
    "junit": "/junit-team/junit5",

    # Security Tools
    "bandit": "/PyCQA/bandit",
    "snyk": "/snyk/snyk",
    "owasp-zap": "/zaproxy/zaproxy",

    # Performance Tools
    "lighthouse": "/GoogleChrome/lighthouse",
    "py-spy": "/benfred/py-spy",

    # Documentation Standards
    "openapi": "/OAI/OpenAPI-Specification",
    "sphinx": "/sphinx-doc/sphinx",
    "jsdoc": "/jsdoc/jsdoc"
}
```

## Integration Patterns

**Quality-as-Service Integration**:
```python
# REST API for quality analysis
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(title="Code Quality Analysis API")

class QualityAnalysisRequest(BaseModel):
    repository_url: str
    languages: List[str]
    quality_threshold: float = 0.85

class QualityAnalysisResponse(BaseModel):
    analysis_id: str
    overall_score: float
    trust5_validation: Dict
    recommendations: List[Dict]
    analysis_completed_at: datetime

@app.post("/api/quality/analyze", response_model=QualityAnalysisResponse)
async def analyze_quality(request: QualityAnalysisRequest):
    """API endpoint for quality analysis"""

    try:
        # Clone and analyze repository
        with tempfile.TemporaryDirectory() as temp_dir:
            await clone_repository(request.repository_url, temp_dir)

            quality_orchestrator = QualityOrchestrator()
            quality_result = await quality_orchestrator.analyze_codebase(
                path=temp_dir,
                languages=request.languages,
                quality_threshold=request.quality_threshold
            )

        return QualityAnalysisResponse(
            analysis_id=str(uuid.uuid4()),
            overall_score=quality_result.overall_score,
            trust5_validation=quality_result.trust5_validation.dict(),
            recommendations=[rec.dict() for rec in quality_result.proactive_analysis.recommendations],
            analysis_completed_at=datetime.utcnow()
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
```

---

## Quick Reference Summary

**Core Capabilities**: TRUST 5 validation, proactive scanning, Context7-powered best practices, multi-language support, enterprise integration

**Key Classes**: `QualityOrchestrator`, `TRUST5Validator`, `ProactiveQualityScanner`, `BestPracticesEngine`, `QualityMetricsCollector`

**Essential Methods**: `analyze_codebase()`, `validate_trust5()`, `scan_for_issues()`, `validate_best_practices()`, `generate_quality_report()`

**Integration Ready**: CI/CD pipelines, GitHub Actions, REST APIs, real-time monitoring, cross-project benchmarking

**Enterprise Features**: Custom rules, ML prediction, real-time monitoring, benchmarking, comprehensive reporting

**Quality Standards**: OWASP compliance, TRUST 5 framework, Context7 integration, automated improvement recommendations