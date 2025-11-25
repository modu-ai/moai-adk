# Quality Validation Examples

**Purpose**: Real-world examples and usage patterns for moai-quality-validation skill
**Target**: Developers and quality assurance teams
**Last Updated**: 2025-11-25
**Version**: 1.0.0

## Quick Usage Examples

### Basic Quality Validation

```python
# Simple quality check
Task(
    subagent_type="quality-gate",
    prompt="Validate project quality",
    context={
        "validation_scope": ["security", "testing", "performance"],
        "target_coverage": 90,
        "framework": "TRUST 5"
    }
)

# Comprehensive validation with Context7
Task(
    subagent_type="moai-quality-validation",
    prompt="Full quality validation with latest best practices",
    context={
        "include_all_validators": True,
        "context7_enabled": True,
        "generate_report": True
    }
)
```

### Security Validation Examples

#### OWASP Top 10 Security Scan

```python
# Security validation for Python web application
Task(
    subagent_type="security-expert",
    prompt="Run comprehensive security validation using moai-quality-validation",
    context={
        "project_path": "/path/to/django-app",
        "security_checks": [
            "owasp_validation",
            "dependency_scanning", 
            "secret_detection",
            "gdpr_compliance"
        ],
        "severity_threshold": "medium",
        "generate_report": True
    }
)

# Expected output:
# 📊 Security Validation Results
# Found 12 security issues:
#   🚨 CRITICAL: 2 issues
#   ⚠️  HIGH: 5 issues
#   ℹ️  MEDIUM: 5 issues
```

#### GDPR Compliance Check

```python
# GDPR compliance validation
Task(
    subagent_type="quality-gate",
    prompt="Validate GDPR compliance using moai-quality-validation",
    context={
        "compliance_standards": ["gdpr"],
        "data_protection_checks": True,
        "consent_mechanisms": True,
        "right_to_be_forgotten": True,
        "data_portability": True
    }
)
```

### Testing Validation Examples

#### TDD Compliance Validation

```python
# TDD compliance check
Task(
    subagent_type="test-engineer",
    prompt="Validate TDD compliance and test quality",
    context={
        "project_path": "/path/to/project",
        "tdd_checks": {
            "test_to_source_ratio": True,
            "missing_tests": True,
            "test_quality": True,
            "coverage_analysis": True
        },
        "coverage_threshold": 85,
        "framework": "pytest"
    }
)

# Expected output:
# 🧪 Testing Validation Results
# 📊 Coverage: 78.3% lines, 65.2% branches
# 🔄 TDD Compliance: 65.7%, Ratio: 0.4
# ⚠️  Total Issues: 8
# 💡 Recommendations:
#   • Increase line coverage from 78.3% to 85%+ by adding comprehensive unit tests
#   • Improve TDD compliance by creating test files for 12 missing source modules
```

#### Test Quality Analysis

```python
# Comprehensive test quality validation
Task(
    subagent_type="quality-gate",
    prompt="Analyze test quality and framework usage",
    context={
        "validation_type": "testing",
        "framework_analysis": {
            "pytest": {"enabled": True, "check_fixtures": True},
            "unittest": {"enabled": True}
        },
        "quality_checks": {
            "assertions": True,
            "docstrings": True,
            "test_isolation": True,
            "setup_teardown": True
        }
    }
)
```

### Performance Validation Examples

#### Scalene Profiling Integration

```python
# Performance profiling with Scalene
Task(
    subagent_type="performance-engineer",
    prompt="Run performance validation with Scalene profiling",
    context={
        "project_path": "/path/to/project",
        "profiling": {
            "enable_scalene": True,
            "cpu_analysis": True,
            "memory_analysis": True,
            "algorithmic_complexity": True
        },
        "benchmarks": True,
        "optimization_suggestions": True
    }
)

# Expected output:
# ⚡ Performance Validation Results
# 🔍 Code Performance Issues: 6
#    High Impact: 3
# 🔄 High Complexity Functions: 4
# 💡 Performance Recommendations:
#   • Address 3 high-impact performance bottlenecks in code structure and algorithms
#   • Optimize 4 high-complexity functions for better algorithmic efficiency
```

#### Core Web Vitals Validation

```python
# Web performance validation
Task(
    subagent_type="performance-engineer",
    prompt="Validate Core Web Vitals and web performance",
    context={
        "web_performance": {
            "core_web_vitals": True,
            "lighthouse_analysis": True,
            "optimization_suggestions": True
        },
        "thresholds": {
            "lcp": 2.5,
            "fid": 100,
            "cls": 0.1
        }
    }
)
```

## Advanced Usage Patterns

### CI/CD Integration

#### GitHub Actions Workflow

```yaml
# .github/workflows/quality-validation.yml
name: Quality Validation

on: [push, pull_request]

jobs:
  quality-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      
      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          pip install pytest-cov bandit safety scalene
      
      - name: Run Quality Validation
        run: |
          python -c "
          from moai_quality_validation import QualityValidationPipeline
          import asyncio
          
          async def main():
              pipeline = QualityValidationPipeline()
              results = await pipeline.run_validation_pipeline(
                  Path('.'), 
                  {
                      'include_context7': True,
                      'enable_ml_predictions': True,
                      'generate_detailed_report': True
                  }
              )
              
              # Check quality gates
              if results['metrics']['secured']['score'] < 0.9:
                  print('❌ Security validation failed')
                  exit(1)
              
              if results['metrics']['testable']['score'] < 0.85:
                  print('❌ Testing validation failed')
                  exit(1)
              
              print('✅ Quality validation passed')
          
          asyncio.run(main())
          "
      
      - name: Upload Quality Report
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: quality-report
          path: quality_validation_report.json
```

#### Pre-commit Hook

```bash
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: quality-validation
        name: Quality Validation
        entry: python
        language: system
        args:
          - -c
          - |
            import asyncio
            from pathlib import Path
            from moai_quality_validation import validate_project_quality
            
            async def main():
                result = await validate_project_quality(
                    Path('.'),
                    {'quick_mode': True}
                )
                
                if result['overall_score'] < 0.7:
                    print('❌ Quality validation failed')
                    print(f'Score: {result["overall_score"]:.1%}')
                    exit(1)
                else:
                    print(f'✅ Quality validation passed: {result["overall_score"]:.1%}')
            
            asyncio.run(main())
        files: \.py$
```

### Context7 Integration Examples

#### Real-time Best Practices

```python
# Quality validation with Context7 integration
Task(
    subagent_type="quality-gate",
    prompt="Validate code using latest best practices via Context7",
    context={
        "context7_enabled": True,
        "libraries_to_check": [
            "django-security",
            "pytest-testing", 
            "fastapi-performance",
            "flask-best-practices"
        ],
        "validation_scope": "comprehensive",
        "apply_latest_standards": True
    }
)
```

#### Library-specific Validation

```python
# Framework-specific validation with Context7
Task(
    subagent_type="quality-gate",
    prompt="Validate Django app security and performance",
    context={
        "framework": "django",
        "context7_libraries": [
            "django-security",
            "django-testing",
            "django-performance"
        ],
        "django_specific_checks": {
            "middleware_security": True,
            "settings_validation": True,
            "orm_optimization": True
        }
    }
)
```

## Configuration Examples

### Project Configuration

```python
# quality_config.py
QUALITY_VALIDATION_CONFIG = {
    "project_type": "web_application",
    "framework": "django",
    "validation_enabled": {
        "security": True,
        "testing": True,
        "performance": True,
        "code_review": True
    },
    "thresholds": {
        "security_score": 0.95,
        "test_coverage": 85.0,
        "performance_score": 0.80,
        "code_quality": 0.85
    },
    "context7": {
        "enabled": True,
        "auto_update": True,
        "libraries": [
            "owasp",
            "django-security",
            "pytest",
            "performance-optimization"
        ]
    },
    "notifications": {
        "slack_webhook": "https://hooks.slack.com/services/...",
        "email_recipients": ["team@company.com"],
        "on_failure_only": True
    }
}
```

### Custom Validation Rules

```python
# custom_validations.py
from moai_quality_validation import ValidationRule, ValidationResult

class CustomSecurityRule(ValidationRule):
    """Custom security validation rule."""
    
    def validate(self, code_content: str) -> List[ValidationResult]:
        issues = []
        
        # Custom validation logic
        if "DEBUG = True" in code_content and "production" in code_content:
            issues.append(ValidationResult(
                category="security",
                level="critical",
                message="Debug mode enabled in production configuration",
                location="settings",
                recommendation="Disable debug mode in production"
            ))
        
        return issues

# Register custom rule
validation_engine = QualityValidationEngine()
validation_engine.register_custom_rule(CustomSecurityRule())
```

## Integration Examples

### with MoAI-ADK Commands

```bash
# Using moai-cc-commands integration
/moai:quality-validate --scope=security,testing --threshold=high

# Generate comprehensive quality report
/moai:quality-report --format=json --output=quality_report.json

# Run specific validation
/moai:quality-validate --module=security --framework=django
```

### with MoAI Agent Factory

```python
# Create specialized quality validation agent
Task(
    subagent_type="agent-factory",
    prompt="Create specialized quality validation agent for our Django project",
    context={
        "agent_type": "quality-validator",
        "specialization": "django-security-testing",
        "skills": ["moai-quality-validation"],
        "configuration": {
            "focus_areas": ["owasp", "gdpr", "performance"],
            "reporting_format": "json",
            "integration": "github_actions"
        }
    }
)
```

## Real-world Scenarios

### E-commerce Platform Validation

```python
# E-commerce quality validation
Task(
    subagent_type="quality-gate",
    prompt="Validate e-commerce platform for production readiness",
    context={
        "project_type": "ecommerce",
        "critical_checks": {
            "payment_security": True,  # PCI-DSS compliance
            "gdpr_compliance": True,
            "performance_under_load": True,
            "transaction_integrity": True
        },
        "load_testing": {
            "concurrent_users": 1000,
            "response_time_threshold": 2.0,
            "error_rate_threshold": 0.01
        },
        "security_focus": [
            "owasp_a02",  # Cryptographic failures
            "owasp_a03",  # Injection
            "owasp_a06"   # Vulnerable components
        ]
    }
)
```

### Healthcare Application Validation

```python
# Healthcare application validation with HIPAA
Task(
    subagent_type="security-expert",
    prompt="Validate healthcare application for HIPAA compliance",
    context={
        "compliance_standards": ["hipaa", "gdpr"],
        "data_protection": {
            "phi_encryption": True,
            "audit_logging": True,
            "access_controls": True,
            "data_retention": True
        },
        "security_level": "high",
        "validation_scope": "comprehensive"
    }
)
```

### Financial Services Validation

```python
# Financial services validation
Task(
    subagent_type="quality-gate",
    prompt="Validate financial application for regulatory compliance",
    context={
        "compliance_standards": ["pci-dss", "sox", "gdpr"],
        "security_focus": {
            "financial_data_protection": True,
            "transaction_security": True,
            "audit_trail": True,
            "encryption_at_rest": True,
            "encryption_in_transit": True
        },
        "performance_requirements": {
            "transaction_throughput": 1000,  # transactions/second
            "response_time_p99": 500,        # milliseconds
            "availability": 99.99            # percentage
        }
    }
)
```

## Troubleshooting Examples

### Common Issues and Solutions

#### Scalene Not Available

```python
# Fallback when Scalene is not available
Task(
    subagent_type="performance-engineer",
    prompt="Run performance validation with fallback profiling",
    context={
        "profiling": {
            "enable_scalene": False,
            "fallback_profiling": True,
            "builtin_profiling": ["cProfile", "memory_profiler"]
        },
        "optimization_focus": ["algorithmic", "memory", "cpu"]
    }
)
```

#### Context7 Integration Issues

```python
# Quality validation without Context7
Task(
    subagent_type="quality-gate",
    prompt="Run quality validation with local best practices",
    context={
        "context7_enabled": False,
        "local_standards": True,
        "validation_scope": "basic",
        "fallback_rules": True
    }
)
```

#### Memory Usage Optimization

```python
# Memory-efficient validation for large projects
Task(
    subagent_type="quality-gate",
    prompt="Run memory-efficient quality validation",
    context={
        "memory_optimization": {
            "chunk_size": 1000,
            "parallel_processing": False,
            "cache_results": True,
            "exclude_patterns": ["node_modules", ".git", "__pycache__"]
        },
        "validation_scope": "critical_only"
    }
)
```

## Output Examples

### Quality Report Structure

```json
{
  "validation_timestamp": "2025-11-25T10:30:00Z",
  "project_path": "/path/to/project",
  "overall_score": 0.87,
  "metrics": {
    "testable": {
      "score": 0.82,
      "issues": [
        {
          "level": "warning",
          "message": "Test coverage 78.3% below 85% threshold",
          "recommendation": "Add comprehensive unit tests"
        }
      ]
    },
    "readable": {
      "score": 0.91,
      "issues": []
    },
    "unified": {
      "score": 0.85,
      "issues": [
        {
          "level": "info", 
          "message": "Inconsistent code style detected",
          "recommendation": "Use automated code formatter"
        }
      ]
    },
    "secured": {
      "score": 0.95,
      "issues": [
        {
          "level": "critical",
          "message": "Hardcoded secret detected",
          "recommendation": "Use environment variables"
        }
      ]
    },
    "trackable": {
      "score": 0.88,
      "issues": []
    }
  },
  "recommendations": [
    "Address critical security vulnerabilities immediately",
    "Increase test coverage to meet 85% threshold",
    "Fix style consistency issues with automated formatter"
  ],
  "ci_cd_gate_passed": false,
  "blocker_issues": 1
}
```

### Summary Dashboard

```
🏆 QUALITY VALIDATION DASHBOARD
═══════════════════════════════════════
📊 Overall Score:        ████████░░ 87%
🔒 Security Score:       ██████████ 95%
🧪 Test Coverage:        ████████░░ 78%
⚡ Performance Score:     ███████░░░ 68%
📖 Code Quality:         █████████░ 88%
🔄 TDD Compliance:        ███████░░░ 72%

⚠️  CRITICAL ISSUES:     1
⚠️  HIGH PRIORITY:       3
ℹ️  MEDIUM PRIORITY:      5
💡  SUGGESTIONS:          8

🚪 CI/CD GATE: ❌ FAILED (1 blocker)
⏰ Validation Time: 2m 34s
📅 Next Review: 2025-11-26
```

---

**File**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-quality-validation/examples.md`
**Purpose**: Real-world examples and usage patterns
**Status**: Production Ready
**Examples Coverage**: Security, Testing, Performance, CI/CD, Context7 Integration
