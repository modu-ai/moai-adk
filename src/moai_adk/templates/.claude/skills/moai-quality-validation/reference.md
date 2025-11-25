# Quality Validation Reference

**Purpose**: Complete API reference and technical documentation for moai-quality-validation
**Target**: Advanced users and integration developers
**Last Updated**: 2025-11-25
**Version**: 1.0.0

## Core API Reference

### Main Validation Engine

#### QualityValidationPipeline

```python
class QualityValidationPipeline:
    """Main quality validation pipeline orchestrator."""
    
    def __init__(self, config: Optional[Dict] = None):
        """
        Initialize quality validation pipeline.
        
        Args:
            config: Configuration dictionary with validation settings
        """
    
    async def run_validation_pipeline(
        self, 
        project_path: Path, 
        options: Optional[Dict] = None
    ) -> Dict[str, Any]:
        """
        Run complete quality validation pipeline.
        
        Args:
            project_path: Path to project directory
            options: Validation options and settings
            
        Returns:
            Dictionary containing validation results, metrics, and recommendations
        """
```

#### CoreValidationEngine

```python
class CoreValidationEngine:
    """Core TRUST 5 validation engine with Context7 integration."""
    
    def __init__(self, config: Optional[Dict] = None):
        """Initialize core validation engine."""
    
    async def validate_project(
        self, 
        project_path: Path, 
        options: Optional[Dict] = None
    ) -> Dict[str, QualityMetric]:
        """
        Validate entire project against TRUST 5 framework.
        
        Args:
            project_path: Path to project directory
            options: Validation options
            
        Returns:
            Dictionary of quality metrics by dimension (testable, readable, unified, secured, trackable)
        """
    
    async def validate_trust5_dimension(
        self, 
        dimension: QualityDimension,
        project_path: Path
    ) -> QualityMetric:
        """
        Validate specific TRUST 5 dimension.
        
        Args:
            dimension: QualityDimension to validate
            project_path: Path to project directory
            
        Returns:
            QualityMetric with validation results
        """
```

### Security Validation API

#### OWASPValidator

```python
class OWASPValidator:
    """Comprehensive OWASP Top 10 2021 validation."""
    
    def __init__(self):
        """Initialize OWASP validator with vulnerability patterns."""
    
    async def validate_project_security(self, project_path: Path) -> List[SecurityVulnerability]:
        """
        Comprehensive security validation of the entire project.
        
        Args:
            project_path: Path to project directory
            
        Returns:
            List of SecurityVulnerability objects with detailed issue information
        """
    
    async def validate_owasp_category(
        self, 
        category: OWASPCategory,
        project_path: Path
    ) -> List[SecurityVulnerability]:
        """
        Validate specific OWASP Top 10 category.
        
        Args:
            category: OWASPCategory to validate
            project_path: Path to project directory
            
        Returns:
            List of vulnerabilities in the specified category
        """
    
    async def scan_for_secrets(self, project_path: Path) -> List[SecurityVulnerability]:
        """
        Scan project for hardcoded secrets and sensitive information.
        
        Args:
            project_path: Path to project directory
            
        Returns:
            List of secret-related vulnerabilities
        """
```

#### SecurityVulnerability Data Structure

```python
@dataclass
class SecurityVulnerability:
    """Security vulnerability with full context."""
    owasp_category: OWASPCategory     # OWASP Top 10 category
    cwe_id: Optional[str]            # CWE identifier
    severity: SeverityLevel          # critical, high, medium, low
    title: str                       # Brief vulnerability title
    description: str                 # Detailed description
    location: str                    # File:line reference
    code_snippet: Optional[str]      # Vulnerable code
    remediation: str                 # How to fix
    references: List[str]            # Additional resources
    impact_score: float              # Business impact (0.0-1.0)
    exploitability: float            # Ease of exploitation (0.0-1.0)
    cvss_score: Optional[float] = None  # CVSS v3.1 score if available
```

### Testing Validation API

#### TestingValidator

```python
class TestingValidator:
    """Comprehensive testing validation engine."""
    
    def __init__(self, config: Optional[Dict] = None):
        """Initialize testing validator with framework support."""
    
    async def validate_project_testing(self, project_path: Path) -> Dict[str, Any]:
        """
        Comprehensive testing validation of the entire project.
        
        Args:
            project_path: Path to project directory
            
        Returns:
            Dictionary containing testing validation results
        """
    
    async def validate_tdd_compliance(self, project_path: Path) -> TDDComplianceMetrics:
        """
        Validate Test-Driven Development compliance.
        
        Args:
            project_path: Path to project directory
            
        Returns:
            TDDComplianceMetrics with detailed TDD analysis
        """
    
    async def run_coverage_analysis(self, project_path: Path) -> TestCoverageMetrics:
        """
        Run comprehensive test coverage analysis.
        
        Args:
            project_path: Path to project directory
            
        Returns:
            TestCoverageMetrics with detailed coverage information
        """
```

#### Test Coverage Metrics

```python
@dataclass
class TestCoverageMetrics:
    """Comprehensive test coverage metrics."""
    line_coverage: float              # Line coverage percentage
    branch_coverage: float            # Branch coverage percentage
    function_coverage: float          # Function coverage percentage
    statement_coverage: float         # Statement coverage percentage
    path_coverage: Optional[float]    # Path coverage if available
    missing_lines: List[str]          # Uncovered line ranges
    uncovered_files: List[str]        # Files with no coverage
    coverage_by_file: Dict[str, float] # Per-file coverage
```

### Performance Validation API

#### PerformanceValidator

```python
class PerformanceValidator:
    """Comprehensive performance validation engine."""
    
    def __init__(self, config: Optional[Dict] = None):
        """Initialize performance validator with profiling tools."""
    
    async def validate_project_performance(self, project_path: Path) -> Dict[str, Any]:
        """
        Comprehensive performance validation of the entire project.
        
        Args:
            project_path: Path to project directory
            
        Returns:
            Dictionary containing performance validation results
        """
    
    async def profile_function(
        self, 
        func_name: str, 
        func_callable, 
        *args, **kwargs
    ) -> Dict[str, Any]:
        """
        Profile a specific function using Scalene or fallback tools.
        
        Args:
            func_name: Name of function to profile
            func_callable: Function object to profile
            *args, **kwargs: Arguments to pass to function
            
        Returns:
            Dictionary containing profiling results and metrics
        """
    
    async def analyze_algorithmic_complexity(
        self, 
        project_path: Path
    ) -> Dict[str, Any]:
        """
        Analyze algorithmic complexity across the project.
        
        Args:
            project_path: Path to project directory
            
        Returns:
            Dictionary containing complexity analysis and recommendations
        """
```

#### ScaleneProfiler Integration

```python
class ScaleneProfiler:
    """Scalene profiler integration for advanced performance analysis."""
    
    def __init__(self):
        """Initialize Scalene profiler with availability check."""
    
    async def profile_function(
        self, 
        func_name: str, 
        func_callable, 
        *args, **kwargs
    ) -> Dict[str, Any]:
        """
        Profile function using Scalene or fallback to built-in profilers.
        
        Args:
            func_name: Name of function to profile
            func_callable: Function object to profile
            *args, **kwargs: Function arguments
            
        Returns:
            Dictionary with profiling metrics and analysis
        """
```

## Configuration Reference

### Configuration Schema

```python
QUALITY_VALIDATION_CONFIG_SCHEMA = {
    "type": "object",
    "properties": {
        "project_type": {
            "type": "string",
            "enum": ["web_application", "api", "library", "cli_tool", "desktop_app"]
        },
        "framework": {
            "type": "string",
            "enum": ["django", "flask", "fastapi", "react", "vue", "angular", "none"]
        },
        "validation_enabled": {
            "type": "object",
            "properties": {
                "security": {"type": "boolean"},
                "testing": {"type": "boolean"},
                "performance": {"type": "boolean"},
                "code_review": {"type": "boolean"}
            }
        },
        "thresholds": {
            "type": "object",
            "properties": {
                "security_score": {"type": "number", "minimum": 0, "maximum": 1},
                "test_coverage": {"type": "number", "minimum": 0, "maximum": 100},
                "performance_score": {"type": "number", "minimum": 0, "maximum": 1},
                "code_quality": {"type": "number", "minimum": 0, "maximum": 1}
            }
        },
        "context7": {
            "type": "object",
            "properties": {
                "enabled": {"type": "boolean"},
                "auto_update": {"type": "boolean"},
                "libraries": {
                    "type": "array",
                    "items": {"type": "string"}
                }
            }
        },
        "notifications": {
            "type": "object",
            "properties": {
                "slack_webhook": {"type": "string", "format": "uri"},
                "email_recipients": {
                    "type": "array",
                    "items": {"type": "string", "format": "email"}
                },
                "on_failure_only": {"type": "boolean"}
            }
        }
    }
}
```

### Default Configuration

```python
DEFAULT_CONFIG = {
    "project_type": "web_application",
    "framework": "none",
    "validation_enabled": {
        "security": True,
        "testing": True,
        "performance": True,
        "code_review": True
    },
    "thresholds": {
        "security_score": 0.90,
        "test_coverage": 85.0,
        "performance_score": 0.80,
        "code_quality": 0.85
    },
    "trust5_thresholds": {
        "testable": 0.85,
        "readable": 0.80,
        "unified": 0.80,
        "secured": 0.90,
        "trackable": 0.80
    },
    "context7": {
        "enabled": True,
        "auto_update": True,
        "cache_ttl": 3600,
        "max_retries": 3
    },
    "security": {
        "owasp_enabled": True,
        "dependency_scanning": True,
        "secret_detection": True,
        "compliance_standards": ["gdpr"]
    },
    "testing": {
        "frameworks": {
            "pytest": {"enabled": True, "version": ">=6.0"},
            "unittest": {"enabled": True}
        },
        "coverage_thresholds": {
            "line_coverage": 85.0,
            "branch_coverage": 80.0,
            "function_coverage": 90.0
        }
    },
    "performance": {
        "enable_scalene": True,
        "fallback_profiling": True,
        "max_profile_time": 60,
        "web_vitals": {
            "lcp": 2.5,    # seconds
            "fid": 100,   # milliseconds
            "cls": 0.1    # score
        }
    }
}
```

## Context7 Integration Reference

### Library Resolution

```python
async def resolve_quality_library(library_name: str) -> str:
    """
    Resolve library name to Context7 compatible ID.
    
    Args:
        library_name: Name of library to resolve
        
    Returns:
        Context7 compatible library ID
    """
    
    library_mappings = {
        "owasp": "/owasp/Top10",
        "django-security": "/django/django",
        "pytest": "/pytest-dev/pytest",
        "performance": "/python/performance",
        "testing": "/python/testing"
    }
    
    return library_mappings.get(library_name, library_name)

# Usage
library_id = await resolve_quality_library("owasp")
docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID=library_id,
    topic="best-practices"
)
```

### Real-time Best Practices Integration

```python
class Context7Integration:
    """Context7 integration for real-time best practices."""
    
    def __init__(self):
        self.cache = {}
        self.rate_limiter = RateLimiter()
    
    async def get_latest_standards(
        self, 
        domain: str, 
        topic: str = "best-practices"
    ) -> Dict:
        """
        Get latest standards via Context7.
        
        Args:
            domain: Domain (e.g., "security", "testing", "performance")
            topic: Specific topic within domain
            
        Returns:
            Dictionary containing latest standards and practices
        """
        
        cache_key = f"{domain}:{topic}"
        
        if cache_key in self.cache:
            return self.cache[cache_key]
        
        # Resolve library ID
        library_id = await mcp__context7__resolve_library_id(f"{domain}-{topic}")
        
        # Get documentation
        docs = await mcp__context7__get_library_docs(
            context7CompatibleLibraryID=library_id,
            topic=topic
        )
        
        # Cache results
        self.cache[cache_key] = docs
        
        return docs
```

## Extension Points

### Custom Validation Rules

```python
class CustomValidationRule:
    """Base class for custom validation rules."""
    
    def __init__(self, name: str, description: str):
        """Initialize custom validation rule."""
        self.name = name
        self.description = description
    
    async def validate(
        self, 
        project_path: Path, 
        context: Optional[Dict] = None
    ) -> List[ValidationIssue]:
        """
        Validate project against custom rule.
        
        Args:
            project_path: Path to project directory
            context: Additional context for validation
            
        Returns:
            List of validation issues found
        """
        raise NotImplementedError

# Example custom rule
class NoHardcodedUrlsRule(CustomValidationRule):
    """Custom rule to detect hardcoded URLs."""
    
    def __init__(self):
        super().__init__(
            "no_hardcoded_urls",
            "Detect and flag hardcoded URLs in configuration"
        )
    
    async def validate(
        self, 
        project_path: Path, 
        context: Optional[Dict] = None
    ) -> List[ValidationIssue]:
        issues = []
        
        url_pattern = r'https?://[^\s"\'<>]+'
        
        for file_path in project_path.rglob("*.py"):
            if file_path.name.startswith("test_"):
                continue
                
            with open(file_path, 'r') as f:
                content = f.read()
            
            matches = re.finditer(url_pattern, content)
            for match in matches:
                line_num = content[:match.start()].count('\n') + 1
                issues.append(ValidationIssue(
                    category="configuration",
                    level="warning",
                    title="Hardcoded URL detected",
                    description=f"Hardcoded URL found: {match.group()}",
                    location=f"{file_path}:{line_num}",
                    recommendation="Move URL to environment variables or configuration file"
                ))
        
        return issues

# Register custom rule
validation_engine = CoreValidationEngine()
validation_engine.register_custom_rule(NoHardcodedUrlsRule())
```

### Custom Framework Analyzers

```python
class CustomFrameworkAnalyzer:
    """Base class for framework-specific analyzers."""
    
    def __init__(self, framework_name: str):
        """Initialize framework analyzer."""
        self.framework_name = framework_name
    
    async def analyze_project(
        self, 
        project_path: Path
    ) -> Dict[str, Any]:
        """
        Analyze project for framework-specific patterns and issues.
        
        Args:
            project_path: Path to project directory
            
        Returns:
            Dictionary containing framework analysis results
        """
        raise NotImplementedError

# Example: Flask analyzer
class FlaskFrameworkAnalyzer(CustomFrameworkAnalyzer):
    """Flask-specific framework analyzer."""
    
    def __init__(self):
        super().__init__("flask")
    
    async def analyze_project(
        self, 
        project_path: Path
    ) -> Dict[str, Any]:
        results = {
            "flask_app_detected": False,
            "security_issues": [],
            "performance_issues": [],
            "best_practices": []
        }
        
        # Detect Flask app
        for py_file in project_path.rglob("*.py"):
            with open(py_file, 'r') as f:
                content = f.read()
            
            if "from flask import" in content or "import flask" in content:
                results["flask_app_detected"] = True
                
                # Check for common Flask issues
                if "app.run(debug=True)" in content:
                    results["security_issues"].append({
                        "issue": "Debug mode enabled",
                        "location": str(py_file),
                        "recommendation": "Disable debug mode in production"
                    })
                
                break
        
        return results
```

## Performance Metrics Reference

### Benchmark Metrics

```python
PERFORMANCE_BENCHMARKS = {
    "small_project": {
        "validation_time": 30,      # seconds
        "memory_usage": 50,         # MB
        "cpu_usage": 15,            # percentage
        "files_analyzed": 100
    },
    "medium_project": {
        "validation_time": 120,     # seconds
        "memory_usage": 200,        # MB
        "cpu_usage": 45,            # percentage
        "files_analyzed": 1000
    },
    "large_project": {
        "validation_time": 300,     # seconds
        "memory_usage": 500,        # MB
        "cpu_usage": 80,            # percentage
        "files_analyzed": 5000
    }
}
```

### Quality Score Calculations

```python
def calculate_quality_score(
    metrics: Dict[str, QualityMetric],
    weights: Optional[Dict[str, float]] = None
) -> float:
    """
    Calculate overall quality score from TRUST 5 metrics.
    
    Args:
        metrics: Dictionary of QualityMetric objects
        weights: Optional weights for each dimension (defaults to equal weighting)
        
    Returns:
        Overall quality score (0.0-1.0)
    """
    
    if weights is None:
        weights = {
            "testable": 0.25,
            "readable": 0.20,
            "unified": 0.20,
            "secured": 0.25,
            "trackable": 0.10
        }
    
    weighted_sum = sum(
        metrics[dim].score * weight 
        for dim, weight in weights.items()
        if dim in metrics
    )
    
    total_weight = sum(
        weight for dim, weight in weights.items() 
        if dim in metrics
    )
    
    return weighted_sum / max(1, total_weight)
```

## Error Handling Reference

### Exception Types

```python
class QualityValidationError(Exception):
    """Base exception for quality validation errors."""
    pass

class SecurityValidationError(QualityValidationError):
    """Security-specific validation errors."""
    pass

class TestingValidationError(QualityValidationError):
    """Testing-specific validation errors."""
    pass

class PerformanceValidationError(QualityValidationError):
    """Performance-specific validation errors."""
    pass

class Context7IntegrationError(QualityValidationError):
    """Context7 integration errors."""
    pass
```

### Error Recovery Strategies

```python
async def validate_with_fallback(
    primary_validator: callable,
    fallback_validator: callable,
    *args, **kwargs
) -> Dict[str, Any]:
    """
    Attempt validation with primary method, fallback to alternative on failure.
    
    Args:
        primary_validator: Primary validation method
        fallback_validator: Fallback validation method
        *args, **kwargs: Arguments to pass to validators
        
    Returns:
        Validation results from successful validator
    """
    
    try:
        return await primary_validator(*args, **kwargs)
    except Exception as primary_error:
        try:
            return await fallback_validator(*args, **kwargs)
        except Exception as fallback_error:
            raise QualityValidationError(
                f"Both primary and fallback validators failed: "
                f"Primary: {primary_error}, Fallback: {fallback_error}"
            )
```

## Integration Examples

### with Claude Code Commands

```python
# Command integration for moai-cc-commands
@claude_command("quality-validate")
async def quality_validate_command(args: List[str]) -> str:
    """
    Claude Code command for quality validation.
    
    Args:
        args: Command line arguments
        
    Returns:
        Formatted validation results
    """
    
    project_path = Path.cwd()
    
    # Parse arguments
    options = parse_command_args(args)
    
    # Run validation
    validator = QualityValidationPipeline()
    results = await validator.run_validation_pipeline(project_path, options)
    
    # Format results
    return format_validation_results(results)

def parse_command_args(args: List[str]) -> Dict[str, Any]:
    """Parse command line arguments for quality validation."""
    
    options = {}
    
    # Simple argument parsing
    for i, arg in enumerate(args):
        if arg.startswith("--scope="):
            options["validation_scope"] = arg.split("=")[1].split(",")
        elif arg.startswith("--threshold="):
            options["severity_threshold"] = arg.split("=")[1]
        elif arg == "--context7":
            options["context7_enabled"] = True
    
    return options
```

### with MoAI Agent Factory

```python
# Agent factory integration
@agent_factory_method
async def create_quality_validator_agent(config: Dict[str, Any]) -> Dict[str, Any]:
    """
    Create specialized quality validation agent.
    
    Args:
        config: Agent configuration
        
    Returns:
        Agent configuration dictionary
    """
    
    agent_config = {
        "agent_type": "quality-validator",
        "specialization": config.get("specialization", "general"),
        "skills": ["moai-quality-validation"],
        "capabilities": [
            "security_validation",
            "testing_validation", 
            "performance_validation",
            "code_review"
        ],
        "configuration": {
            "project_type": config.get("project_type", "web_application"),
            "framework": config.get("framework", "none"),
            "validation_scope": config.get("validation_scope", "comprehensive"),
            "reporting_format": config.get("reporting_format", "json")
        }
    }
    
    return agent_config
```

---

**File**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-quality-validation/reference.md`
**Purpose**: Complete API reference and technical documentation
**Status**: Production Ready
**API Coverage**: Core validation, security, testing, performance, Context7 integration
