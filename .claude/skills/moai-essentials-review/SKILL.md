---
name: moai-essentials-review
description: Enterprise comprehensive code review automation with AI-powered quality analysis, TRUST 5 enforcement, multi-language support, Context7 integration, security scanning, performance analysis, test coverage validation, and automated review feedback generation
version: 1.0.0
modularized: true
last_updated: 2025-11-22
compliance_score: 70
auto_trigger_keywords:
  - database
  - essentials
  - review
  - testing
category_tier: 1
---

## Quick Reference

# Enterprise Code Review Automation

**What it does**: Automates comprehensive code review process with AI-powered quality checks, TRUST 5 principle validation, security vulnerability detection, performance analysis, test coverage verification, and automated review feedback generation.

**3-Phase Review Process**:
```
Phase 1: Automated Checks (5 min)
  → Syntax, linting, security scanning, test coverage

Phase 2: AI Quality Analysis (15 min)
  → TRUST 5 validation, design analysis, pattern matching

Phase 3: Human Review (20 min)
  → Architecture alignment, business logic, documentation
```

**TRUST 5 Automated Validation**:
- **T - Test First**: Coverage ≥85%, comprehensive test scenarios
- **R - Readable**: Functions <50 lines, clear naming, low complexity
- **U - Unified**: Consistent patterns across codebase
- **S - Secured**: OWASP compliance, no vulnerabilities
- **T - Trackable**: SPEC-linked, test-traced

**Key Metrics**:
- Cyclomatic complexity: <10
- Test coverage: ≥85%
- Security issues: 0
- Performance concerns: Minimal

**Context7 Integration**: Latest vulnerability patterns, optimization techniques, and language best practices

---

## Implementation Guide

### AI-Powered Quality Checks

```python
class CodeQualityAnalyzer:
    """AI-powered code quality analysis."""
    
    async def analyze(self, code: str) -> QualityReport:
        metrics = {
            "complexity": calculate_cyclomatic(code),      # Should be <10
            "testability": assess_testability(code),        # Should be >0.85
            "maintainability": calculate_maintainability(code),  # Should be >80
            "readability": assess_readability(code),         # Should be clear
            "security_issues": scan_for_vulnerabilities(code),   # Should be 0
            "performance_concerns": detect_patterns(code),   # Should be minimal
        }
        
        return QualityReport(metrics)
```

### Phase 1: Automated Checks (5 minutes)

**Syntax & Linting**:
```bash
# Python
pylint src/ --fail-under=8.0
black --check src/
mypy src/

# JavaScript/TypeScript
eslint src/ --max-warnings 0
prettier --check src/
tsc --noEmit

# Go
golint ./...
gofmt -l .
go vet ./...
```

**Security Scanning**:
```bash
# Dependency vulnerabilities
pip-audit  # Python
npm audit  # JavaScript
cargo audit  # Rust

# Credential detection
git-secrets --scan
detect-secrets scan

# OWASP Top 10 checks
bandit -r src/  # Python
eslint-plugin-security  # JavaScript
```

**Test Coverage**:
```bash
# Coverage validation
pytest --cov=src --cov-report=term --cov-fail-under=85

# Critical path verification
pytest -m critical

# Edge case validation
pytest -m edge_cases
```

### Phase 2: AI Quality Analysis (15 minutes)

**TRUST 5 Validation**:
```
T - Test First:
  ├─ Coverage ≥85%? ✓
  ├─ Happy path covered? ✓
  ├─ Edge cases tested? ✓
  └─ Error scenarios? ✓

R - Readable:
  ├─ Functions <50 lines? ✓
  ├─ Meaningful names? ✓
  ├─ Comments explain WHY? ✓
  └─ Complexity <10? ✓

U - Unified:
  ├─ Follows team patterns? ✓
  ├─ Consistent style? ✓
  ├─ Error handling aligned? ✓
  └─ Logging strategy consistent? ✓

S - Secured:
  ├─ Inputs validated? ✓
  ├─ No hardcoded secrets? ✓
  ├─ SQL injection prevention? ✓
  └─ XSS prevention? ✓

T - Trackable:
  ├─ SPEC referenced? ✓
  ├─ Tests linked? ✓
  └─ Code traced? ✓
```

**Design Analysis**:
```python
def analyze_design_patterns(code: CodeAst) -> DesignAnalysis:
    """Analyze design patterns and SOLID principles."""
    
    checks = {
        "single_responsibility": verify_srp(code),
        "open_closed": verify_ocp(code),
        "liskov_substitution": verify_lsp(code),
        "interface_segregation": verify_isp(code),
        "dependency_inversion": verify_dip(code),
        "scalability": assess_scalability(code),
        "performance": assess_performance_patterns(code)
    }
    
    return DesignAnalysis(checks)
```

### Phase 3: Human Review (20 minutes)

**Architectural Review**:
```markdown
Questions to address:
- Does solution fit existing architecture?
- Were alternative approaches considered?
- Are trade-offs documented?
- Does it introduce technical debt?
- Are integration points clear?
```

**Business Logic Validation**:
```markdown
Critical checks:
- Does it solve the stated problem?
- Are edge cases properly handled?
- What's the user experience impact?
- Are acceptance criteria met?
- Are there performance implications?
```

**Documentation Review**:
```markdown
Required updates:
- README.md reflects changes
- API documentation current
- Examples provided and tested
- Migration guide (if breaking)
- CHANGELOG.md updated
```

### Security Vulnerability Detection

**Critical Checks**:
```yaml
critical:
  - hardcoded_credentials: "API keys, passwords, tokens"
  - sql_injection: "Unparameterized SQL queries"
  - xss_vulnerabilities: "Unescaped user input in HTML"
  - csrf_missing: "No CSRF token validation"
  - unsafe_deserialization: "pickle, eval, exec usage"
  - privilege_escalation: "Unchecked authorization paths"

high_priority:
  - missing_input_validation: "No input sanitization"
  - weak_cryptography: "MD5, SHA1 usage"
  - insecure_randomness: "random() for security"
  - race_conditions: "TOCTOU vulnerabilities"
  - dependency_vulnerabilities: "Known CVEs in deps"

medium_priority:
  - missing_error_messages: "Information disclosure"
  - insufficient_logging: "Security event gaps"
  - memory_leaks: "Resource not freed"
  - resource_exhaustion: "No rate limiting"
```

### Performance Analysis

**Detection Patterns**:
```python
def detect_performance_issues(code: str) -> List[PerformanceIssue]:
    """Identify common performance anti-patterns."""
    
    issues = []
    
    # O(n²) in O(n) context
    if nested_loops_with_membership_test(code):
        issues.append(PerformanceIssue(
            type="algorithmic_complexity",
            message="Nested loops with membership test: O(n²)",
            suggestion="Use set for O(1) membership testing"
        ))
    
    # I/O in loops
    if io_operations_in_loop(code):
        issues.append(PerformanceIssue(
            type="io_in_loop",
            message="File I/O operation inside loop",
            suggestion="Batch I/O operations or use async"
        ))
    
    # Blocking in async context
    if blocking_in_async(code):
        issues.append(PerformanceIssue(
            type="async_blocking",
            message="Blocking operation in async function",
            suggestion="Use await or run_in_executor()"
        ))
    
    return issues
```

### Automated Review Report

**Report Template**:
```markdown
# Code Review Report

## Summary
✅ **Status**: APPROVED (with 2 minor notes)
- Test Coverage: 87% ✓
- Security: ✓ Clean
- Performance: ✓ No concerns
- Design: ✓ Good
- TRUST 5: All checks passed

## TRUST 5 Assessment

### T - Test First: ✓
Coverage: 87% (target ≥85%)
- Happy path: ✓ Covered
- Edge cases: ✓ 5 tests
- Error scenarios: ✓ 3 tests

### R - Readable: ✓
All functions <50 lines, clear names

### U - Unified: ✓
Consistent with team patterns

### S - Secured: ✓
- No credentials: ✓
- Input validation: ✓
- Error messages safe: ✓

### T - Trackable: ✓
- SPEC-042 referenced
- 5 tests linked
- Code linked to PR

## Detailed Findings

### Strengths
1. ✅ Excellent test coverage (87%)
2. ✅ Clean, readable code
3. ✅ Proper error handling
4. ✅ Security best practices followed

### Minor Notes
1. ⚠️ Function `calculate_discount` could use type hints
2. ⚠️ Consider adding cache for frequently called API

### Recommendations
1. Add type hints to improve IDE support
2. Consider Redis caching for API calls

## Approval
✅ **Ready to merge** - All TRUST 5 checks passed
```

### Integration with Context7

**Latest Security Patterns**:
```python
async def fetch_security_patterns():
    """Get latest vulnerability detection patterns."""
    
    docs = await context7.get_library_docs(
        context7_library_id="/owasp/top-ten",
        topic="vulnerability detection patterns 2025",
        tokens=3000
    )
    
    return docs
```

**Performance Optimization**:
```python
async def fetch_performance_patterns(language: str):
    """Get language-specific optimization patterns."""
    
    library_map = {
        "python": "/python/performance-guide",
        "javascript": "/v8/optimization-manual",
        "go": "/golang/perf-best-practices"
    }
    
    docs = await context7.get_library_docs(
        context7_library_id=library_map[language],
        topic="performance optimization patterns",
        tokens=2000
    )
    
    return docs
```

### Best Practices

**DO**:
- ✅ Run automated checks before human review
- ✅ Provide specific, actionable feedback
- ✅ Explain WHY improvements are needed
- ✅ Link to official documentation
- ✅ Flag security issues immediately
- ✅ Enforce TRUST 5 consistently
- ✅ Update based on new findings
- ✅ Track metrics over time

**DON'T**:
- ❌ Block on automated issues alone (let linters handle)
- ❌ Miss security vulnerabilities
- ❌ Accept coverage <85%
- ❌ Ignore deprecated patterns
- ❌ Skip performance analysis
- ❌ Approve without TRUST 5 validation
- ❌ Add comments that code already explains

---

## Advanced Patterns

### Multi-Language Review Automation

**Language-Specific Configuration**:
```yaml
review_config:
  python:
    linters: [pylint, black, mypy]
    security: [bandit, safety]
    complexity_max: 10
    coverage_min: 85
    
  javascript:
    linters: [eslint, prettier, tsc]
    security: [npm-audit, eslint-plugin-security]
    complexity_max: 10
    coverage_min: 85
    
  go:
    linters: [golint, gofmt, go vet]
    security: [gosec]
    complexity_max: 10
    coverage_min: 85
```

### AI-Powered Pattern Recognition

**Advanced Analysis**:
```python
class AdvancedPatternAnalyzer:
    """AI-powered advanced pattern recognition."""
    
    async def analyze_patterns(self, code: str) -> AdvancedAnalysis:
        """Perform deep pattern analysis using AI."""
        
        # Get Context7 patterns
        context7_patterns = await self.context7.get_library_docs(
            context7_library_id="/code-patterns/enterprise",
            topic="design patterns anti-patterns best-practices",
            tokens=5000
        )
        
        # AI pattern matching
        patterns_detected = self.ai_engine.match_patterns(
            code, context7_patterns
        )
        
        # Generate recommendations
        recommendations = self.generate_recommendations(
            patterns_detected, context7_patterns
        )
        
        return AdvancedAnalysis(
            patterns=patterns_detected,
            recommendations=recommendations,
            confidence=self.calculate_confidence(patterns_detected)
        )
```

### Continuous Improvement

**Learning from Reviews**:
```python
class ReviewLearningSystem:
    """Learn from review history to improve."""
    
    def analyze_review_trends(self, reviews: List[Review]) -> Trends:
        """Analyze review patterns over time."""
        
        trends = {
            "common_issues": self.extract_common_issues(reviews),
            "improvement_areas": self.identify_improvement_areas(reviews),
            "best_practices": self.extract_best_practices(reviews),
            "automation_opportunities": self.find_automation_gaps(reviews)
        }
        
        return Trends(trends)
    
    def update_review_rules(self, trends: Trends):
        """Update automated review rules based on trends."""
        
        for issue in trends.common_issues:
            if issue.frequency > 0.7:  # Appears in 70%+ of reviews
                self.add_automated_check(issue)
```

---

## Related Skills

- `moai-alfred-code-reviewer` (Manual review guidance)
- `moai-essentials-debug` (Debugging techniques)
- `moai-foundation-trust` (TRUST 5 principles)
- `moai-essentials-perf` (Performance optimization)
- `moai-security-api` (Security best practices)

---

**Last Updated**: 2025-11-21  
**Status**: Production Ready (Enterprise)