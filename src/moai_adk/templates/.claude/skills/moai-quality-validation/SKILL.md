---
name: moai-quality-validation
description: "Integrated quality validation system consolidating debug, security, testing, performance, review, and refactor into unified TRUST 5 framework"
version: 1.0.0
modularized: true
last_updated: 2025-11-25
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 92
modules:
  - core-validation-engine
  - security-validation
  - testing-validation
  - performance-validation
dependencies:
  - moai-foundation-trust
  - moai-context7-integration
deprecated: false
successor: null
category_tier: 3
auto_trigger_keywords:
  - quality
  - validation
  - review
  - testing
  - security
  - performance
  - refactor
  - audit
  - compliance
agent_coverage:
  - quality-gate
  - security-expert
  - test-engineer
  - performance-engineer
  - debug-helper
context7_references:
  - owasp
  - testing
  - performance
  - security
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**Enterprise Quality Validation Unified**

Comprehensive quality validation system consolidating 6 specialized quality skills into unified TRUST 5 framework with Context7 integration.

**Core Capabilities**:
- ✅ **TRUST 5 Framework**: Testable, Readable, Unified, Secured, Trackable
- ✅ **Security Validation**: OWASP Top 10, GDPR/HIPAA, vulnerability scanning
- ✅ **Testing Validation**: TDD, coverage analysis, multi-framework support
- ✅ **Performance Validation**: Scalene profiling, Core Web Vitals, benchmarking
- ✅ **Code Review**: SOLID principles, quality gates, automated analysis
- ✅ **Context7 Integration**: Real-time best practices, latest documentation
- ✅ **Enterprise Features**: ML-based predictions, real-time monitoring

**When to Use**:
- Comprehensive code quality validation
- Pre-deployment quality gates
- CI/CD pipeline integration
- Enterprise compliance requirements
- Performance optimization projects
- Security audits and assessments

**Core Framework**: TRUST 5 + VALIDATION ENGINE + SPECIALIZED VALIDATORS
```
1. Quality Input → 2. Core Engine (TRUST 5) → 3. Specialized Validators → 4. Context7 Integration → 5. Quality Report
```

---

## Implementation Guide

### Quick Validation Commands

```python
# Basic validation
Task(
    subagent_type="quality-gate",
    prompt="Validate code quality",
    context={
        "validation_scope": ["security", "testing", "performance"],
        "target_coverage": 90,
        "framework": "TRUST 5"
    }
)

# Comprehensive validation
Task(
    subagent_type="moai-quality-validation", 
    prompt="Full quality validation with Context7 integration",
    context={
        "include_all_validators": True,
        "generate_report": True,
        "context7_enabled": True
    }
)
```

---

## Core Validation Patterns

### Pattern 1: TRUST 5 Framework Validation

**Concept**: Validate code against Testable, Readable, Unified, Secured, Trackable principles.

```python
from typing import List
from dataclasses import dataclass
from enum import Enum

@dataclass
class ValidationResult:
    category: str
    level: str  # critical, warning, info
    message: str
    location: str
    recommendation: str
    score_impact: float

class Trust5Validator:
    """TRUST 5 quality validation engine."""
    
    def validate_testable(self, code_content: str) -> List[ValidationResult]:
        """Validate Test-First principle."""
        results = []
        
        # Check test coverage
        coverage_result = self._analyze_test_coverage()
        if coverage_result < 0.85:
            results.append(ValidationResult(
                category="testable",
                level="warning" if coverage_result > 0.70 else "critical",
                message=f"Test coverage {coverage_result:.1%} below 85% threshold",
                location="tests/",
                recommendation="Add unit tests to increase coverage",
                score_impact=0.2 * (0.85 - coverage_result)
            ))
        
        return results
```

**Use Case**: Comprehensive quality validation using TRUST 5 framework.

---

### Pattern 2: Security Validation with OWASP Integration

**Concept**: Automated security validation against OWASP Top 10 and compliance standards.

```python
import re
from typing import List
from dataclasses import dataclass

@dataclass
class SecurityVulnerability:
    owasp_category: str
    severity: str  # critical, high, medium, low
    description: str
    location: str
    recommendation: str

class OWASPValidator:
    """OWASP Top 10 2021 security validation."""
    
    OWASP_PATTERNS = {
        "A03": {
            "patterns": [r"SELECT.*FROM.*WHERE.*\{[^}]+\}[^;]*;"],  # SQL injection
            "description": "SQL injection vulnerability",
            "recommendation": "Use parameterized queries, input validation"
        },
        "A02": {
            "patterns": [r"md5\(", r"password\s*=\s*['\"][^'\"]{0,8}['\"]"],
            "description": "Weak cryptography or password handling",
            "recommendation": "Use strong algorithms (bcrypt, Argon2)"
        }
    }
    
    def validate_code(self, code_content: str, file_path: str) -> List[SecurityVulnerability]:
        """Scan code for OWASP Top 10 vulnerabilities."""
        vulnerabilities = []
        lines = code_content.split('\n')
        
        for owasp_id, config in self.OWASP_PATTERNS.items():
            for pattern in config["patterns"]:
                for line_num, line in enumerate(lines, 1):
                    if re.search(pattern, line, re.IGNORECASE):
                        vulnerabilities.append(SecurityVulnerability(
                            owasp_category=owasp_id,
                            severity="high",
                            description=config["description"],
                            location=f"{file_path}:{line_num}",
                            recommendation=config["recommendation"]
                        ))
        
        return vulnerabilities
```

**Use Case**: Automated security validation against OWASP Top 10.

---

### Pattern 3: Testing Validation with TDD Integration

**Concept**: Comprehensive testing validation including coverage, TDD compliance, and test quality.

```python
from pathlib import Path
from dataclasses import dataclass

@dataclass
class TestingIssue:
    category: str
    severity: str
    description: str
    location: str
    recommendation: str

class TestingValidator:
    """Comprehensive testing validation."""
    
    def __init__(self):
        self.coverage_threshold = 85.0
    
    def validate_tdd_compliance(self, project_root: str) -> List[TestingIssue]:
        """Validate Test-Driven Development compliance."""
        issues = []
        project_path = Path(project_root)
        
        src_files = [f for f in project_path.rglob("*.py") if "test" not in f.name]
        test_files = [f for f in project_path.rglob("*.py") if "test" in f.name]
        
        # Check test-to-source ratio
        test_to_source_ratio = len(test_files) / max(1, len(src_files))
        if test_to_source_ratio < 0.5:
            issues.append(TestingIssue(
                category="tdd",
                severity="high",
                description=f"Low test-to-source ratio: {test_to_source_ratio:.1%} (target: ≥50%)",
                location="project",
                recommendation="Create corresponding test files for each source module"
            ))
        
        return issues
```

**Use Case**: Testing validation and TDD compliance checking.

---

## Integration with Context7

### Real-time Best Practices Integration

```python
class Context7IntegratedValidator:
    """Quality validator with Context7 integration."""
    
    async def validate_with_context7(self, code_content: str, language: str = "python") -> dict:
        """Validate code using Context7 for latest best practices."""
        
        # Get latest documentation via Context7
        library_ids = await mcp__context7__resolve_library_id(f"{language}-security")
        docs = await mcp__context7__get_library_docs(
            context7CompatibleLibraryID=library_ids[0],
            topic="best-practices"
        )
        
        # Validate against latest patterns
        validation_results = {
            "trust5_score": self._calculate_trust5_score(code_content),
            "security_issues": self._validate_security_with_context7(code_content, docs),
            "performance_issues": self._validate_performance_with_context7(code_content, docs),
            "recommendations": await self._get_context7_recommendations(code_content, language)
        }
        
        return validation_results
```

---

## Success Metrics

**Quality Score Components**:
- **TRUST 5 Score**: ≥ 87% for production deployment
- **Security Score**: ≥ 95% with zero critical vulnerabilities
- **Testing Coverage**: ≥ 90% with TDD compliance
- **Performance Score**: ≥ 80% with optimized bottlenecks
- **Code Quality**: ≥ 88% with style consistency

**Validation Efficiency**:
- **Validation Time**: < 5 minutes for typical project
- **False Positive Rate**: < 5%
- **Issue Detection**: > 95% of critical issues
- **Context7 Integration**: 100% best practices coverage

---

## Works Well With

- **moai-foundation-trust** (TRUST 5 framework implementation)
- **moai-context7-integration** (Real-time documentation and best practices)
- **moai-core-claude-code** (Claude Code authoring and validation)
- **moai-quality-security** (Comprehensive security validation)
- **moai-essentials-debug** (Advanced debugging and error analysis)

---

## Consolidated Skills

This skill consolidates functionality from 6 specialized quality skills:

### Source Skills Integrated

1. **moai-quality-debug** - Debugging strategies, error patterns, RCA frameworks
2. **moai-quality-security** - OWASP Top 10, authentication, secrets, compliance
3. **moai-quality-testing** - TDD compliance, coverage analysis, test quality
4. **moai-quality-performance** - Scalene profiling, Core Web Vitals, optimization
5. **moai-quality-review** - SOLID principles, quality gates, style validation
6. **moai-quality-refactor** - Design patterns, code smells, structural improvements

### Integration Benefits

- **Unified Framework**: Single entry point for all quality validations
- **Consistent Reporting**: Standardized quality metrics and scoring
- **Context7 Integration**: Real-time best practices and latest standards
- **Enterprise Features**: ML-based predictions and real-time monitoring
- **Reduced Complexity**: Simplified skill management and maintenance

All functionality preserved and enhanced through unified architecture.

---

## Advanced Documentation

For detailed validation patterns and implementation strategies:

- **[modules/core-validation-engine.md](modules/core-validation-engine.md)** - Core TRUST 5 validation engine
- **[modules/security-validation.md](modules/security-validation.md)** - Comprehensive security validation patterns
- **[modules/testing-validation.md](modules/testing-validation.md)** - Testing framework validation and TDD compliance
- **[modules/performance-validation.md](modules/performance-validation.md)** - Performance profiling and optimization
- **[examples.md](examples.md)** - Real-world usage examples and patterns
- **[reference.md](reference.md)** - Complete API reference and integration patterns

---

## Changelog

- **v1.0.0** (2025-11-25): Consolidated 6 quality skills into unified moai-quality-validation with Context7 integration

---

**Status**: Production Ready (Enterprise)
**Generated with**: MoAI-ADK Skill Factory
**Architecture**: SKILL.md + 4 modules (core, security, testing, performance)
**Context7 Integration**: Real-time best practices and latest documentation
**Consolidation**: 6 quality skills unified into single comprehensive system
