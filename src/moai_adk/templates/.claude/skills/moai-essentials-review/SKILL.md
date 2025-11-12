---
name: "moai-essentials-review"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-12
status: stable
description: Enterprise comprehensive code review automation with AI-powered quality analysis, TRUST 5 enforcement, multi-language support, Context7 integration, security scanning, performance analysis, test coverage validation, and automated review feedback generation
keywords: ['code-review', 'quality-analysis', 'TRUST-5-validation', 'security-scanning', 'performance-analysis', 'test-coverage', 'ai-review', 'context7-integration', 'review-automation', 'enterprise-quality']
allowed-tools: 
  - Read
  - Write
  - Edit
  - Glob
  - Bash
  - AskUserQuestion
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
  - WebFetch
---

# Enterprise Code Review Automation v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-essentials-review |
| **Version** | 4.0.0 Enterprise (2025-11-12) |
| **Core Framework** | TRUST 5 principles automation |
| **AI Integration** | ✅ Context7 MCP, AI quality analysis |
| **Auto-load** | On code commit or PR creation |
| **Languages** | 25+ languages with specialized analysis |
| **Lines of Content** | 880+ with 16+ production examples |
| **Progressive Disclosure** | 3-level (automation, analysis, advanced) |

---

## What It Does

Automates comprehensive code review process with AI-powered quality checks, TRUST 5 principle validation, security vulnerability detection, performance analysis, test coverage verification, and detailed review feedback generation.

---

## 3-Phase Automated Review

### Phase 1: Automated Checks (5 minutes)

```
Syntax & Linting:
  ✓ Run linters (pylint, eslint, golint, etc.)
  ✓ Check code formatting (black, prettier, gofmt)
  ✓ Type checking (mypy, TypeScript, go vet)

Security Scanning:
  ✓ Dependency vulnerabilities (safety, npm audit, cargo audit)
  ✓ Credential detection (git-secrets, detect-secrets)
  ✓ OWASP Top 10 checks

Test Coverage:
  ✓ Coverage ≥85%
  ✓ Critical paths covered
  ✓ Edge cases tested
```

### Phase 2: AI Quality Analysis (15 minutes)

```
TRUST 5 Validation:
  ✓ T - Tests present and comprehensive
  ✓ R - Code readable and maintainable
  ✓ U - Unified with codebase patterns
  ✓ S - Security best practices

Design Analysis:
  ✓ SOLID principles
  ✓ Design patterns appropriate
  ✓ Scalability concerns
  ✓ Performance implications
```

### Phase 3: Human Review (20 minutes)

```
Architectural Review:
  ✓ Does solution fit architecture?
  ✓ Any alternatives considered?
  ✓ Trade-offs documented?

Business Logic:
  ✓ Does it solve the problem?
  ✓ Any edge cases missed?
  ✓ User experience impact?

Documentation:
  ✓ README updated
  ✓ API docs current
  ✓ Examples provided
```

---

## AI-Powered Quality Checks

### Code Quality Metrics

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

### TRUST 5 Automated Checks

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
```

---

## Security Vulnerability Detection

```
Critical Checks:
  ✓ Hardcoded credentials (API keys, passwords)
  ✓ SQL injection vectors
  ✓ XSS vulnerabilities
  ✓ CSRF token absence
  ✓ Unsafe deserialization
  ✓ Privilege escalation paths

High Priority:
  ✓ Missing input validation
  ✓ Weak cryptography
  ✓ Insecure randomness
  ✓ Race conditions
  ✓ Dependency vulnerabilities

Medium Priority:
  ✓ Missing error messages
  ✓ Insufficient logging
  ✓ Memory leaks
  ✓ Resource exhaustion risks
```

---

## Performance Analysis

```
Detection Patterns:
  ✓ O(n²) algorithms in O(n) context
  ✓ Unnecessary file I/O in loops
  ✓ Blocking operations in async code
  ✓ Memory allocations in hot paths
  ✓ Inefficient string concatenation
  ✓ Database queries without indexing

Optimization Suggestions:
  ✓ Use more efficient algorithm
  ✓ Cache results
  ✓ Batch operations
  ✓ Use async/await properly
  ✓ Index database columns
```

---

## Automated Review Report

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

---

## Integration with Context7

**Live Security Patterns**: Get latest vulnerability detection from official databases  
**Performance Optimization**: Context7 provides version-specific optimization patterns  
**Language Updates**: Context7 includes latest language/framework best practices  

---

## Best Practices

### DO
- ✅ Run automated checks before human review
- ✅ Provide specific, actionable feedback
- ✅ Explain WHY improvements are needed
- ✅ Link to official documentation
- ✅ Flag security issues immediately
- ✅ Enforce TRUST 5 consistently
- ✅ Update based on new findings
- ✅ Track metrics over time

### DON'T
- ❌ Block on automated issues alone (let linters handle)
- ❌ Miss security vulnerabilities
- ❌ Accept coverage <85%
- ❌ Ignore deprecated patterns
- ❌ Skip performance analysis
- ❌ Approve without TRUST 5 validation
- ❌ Add comments that code already explains

---

## Related Skills

- `moai-alfred-code-reviewer` (Manual review guidance)
- `moai-essentials-debug` (Debugging techniques)

---

**For detailed analysis guidelines**: [reference.md](reference.md)  
**For real-world examples**: [examples.md](examples.md)  
**Last Updated**: 2025-11-12  
**Status**: Production Ready (Enterprise v4.0.0)
