---
name: moai-alfred-code-reviewer
version: 2.0.0
created: 2025-10-01
updated: 2025-11-11
status: active
description: "Systematic code review guidance and automation. Apply TRUST 5 principles, check code quality, validate SOLID principles, identify security issues, and ensure maintainability. Enhanced with research capabilities for code quality analysis and optimization. Use when conducting code reviews, setting review standards, or implementing review automation."
keywords: ['code-review', 'quality', 'automation', 'research', 'analysis', 'optimization']
allowed-tools: "Read, Write, Edit, Glob, Bash, TodoWrite"
tags: [quality-assurance, code-review, automation, research, analysis, optimization]
---

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Quality |
| Auto-load | When conducting code reviews or quality checks |

## What It Does

체계적인 코드 리뷰 프로세스와 자동화 가이드를 제공합니다. TRUST 5 원칙 적용, 코드 품질 검증, SOLID 원칙 준수, 보안 이슈 식별, 유지보수성 보장을 다룹니다.

## When to Use

- 코드 리뷰를 수행할 때
- 리뷰 표준과 가이드라인을 설정할 때
- 코드 품질 자동화를 구현할 때
- 팀의 코드 리뷰 문화를 개선할 때


# Systematic Code Review with TRUST 5 Principles

Code review is a quality assurance process that ensures code meets standards, follows best practices, and maintains long-term maintainability. This skill provides systematic guidance for conducting thorough, effective code reviews.

## Code Review Framework

### The TRUST 5 Review Framework

| Principle | Review Focus | Key Questions | Tools |
|-----------|--------------|---------------|-------|
| **T** - Test First | Test coverage & quality | Are tests comprehensive? Do they test edge cases? | pytest coverage, jest --coverage |
| **R** - Readable | Code clarity & maintainability | Is code self-documenting? Are names meaningful? | linters, code formatters |
| **U** - Unified | Consistency & standards | Does it follow team patterns? Is it cohesive? | style guides, architectural patterns |
| **S** - Secured | Security & vulnerabilities | Are inputs validated? Are secrets handled properly? | security scanners, static analysis |
| **T** - Trackable | Documentation & traceability | Is code linked to requirements? Are changes documented? | @TAG system, git history |

## High-Freedom: Review Strategy & Philosophy

### When to Review vs When to Trust

```markdown
## Review Decision Matrix

| Change Type | Review Level | Automation | Focus Areas |
|-------------|--------------|------------|-------------|
| Critical security | 🔴 Mandatory | Full scan | Vulnerabilities, input validation |
| Core architecture | 🟡 Deep review | Partial | Design patterns, scalability |
| Bug fixes | 🟢 Standard | Automated | Root cause, test coverage |
| Documentation | 🟢 Light | Basic | Accuracy, completeness |
| Configuration | 🟢 Automated | Full | Security, best practices |
```

### Review Culture Principles

✅ **Psychological Safety**: Reviews are about code, not people  
✅ **Learning Opportunity**: Reviews transfer knowledge and standards  
✅ **Constructive Feedback**: Focus on improvement, not criticism  
✅ **Consistent Standards**: Apply same criteria to all code  
✅ **Efficient Process**: Automated checks first, human review for value-add  

### Review Anti-patterns to Avoid

❌ **Nitpicking**: Focus on style over substance  
❌ **Authoritarian**: "Do it this way because I said so"  
❌ **Incomplete**: "Looks good" without specific feedback  
❌ **Delayed**: Reviews blocking progress for days  
❌ **Inconsistent**: Different standards for different people  

## Medium-Freedom: Review Process & Checklists

### Structured Review Process

```pseudocode
## Review Workflow

1. PRE-REVIEW AUTOMATION (2-5 min)
   a. Run linters and formatters
   b. Execute test suite with coverage
   c. Scan for security vulnerabilities
   d. Check for @TAG compliance
   
2. CODE COMPREHENSION (5-10 min)
   a. Read commit message and PR description
   b. Understand the problem being solved
   c. Identify affected components
   d. Review test changes first
   
3. DETAILED REVIEW (10-20 min)
   a. Apply TRUST 5 framework systematically
   b. Check architectural consistency
   c. Validate error handling
   d. Assess performance implications
   
4. FEEDBACK SYNTHESIS (5 min)
   a. Categorize issues: Must-fix, Should-fix, Nice-to-have
   b. Provide specific, actionable feedback
   c. Explain reasoning behind suggestions
   d. Offer to discuss complex changes
```

### Code Review Checklist Template

```markdown
## Code Review Checklist

### 🧪 Test Coverage (T)
- [ ] New features have corresponding tests
- [ ] Test coverage ≥ 85% (or team standard)
- [ ] Edge cases and error conditions tested
- [ ] Integration tests included where appropriate
- [ ] Tests are readable and maintainable

### 📖 Readability (R)
- [ ] Function and variable names are descriptive
- [ ] Complex logic is commented or extracted
- [ ] File length ≤ 300 LOC (or team standard)
- [ ] Function length ≤ 50 LOC (or team standard)
- [ ] No magic numbers or hardcoded values

### 🔗 Unity (U)
- [ ] Follows established team patterns
- [ ] Consistent with existing codebase style
- [ ] Uses shared utilities and libraries
- [ ] Architecture aligns with project structure
- [ ] Imports and dependencies are organized

### 🔒 Security (S)
- [ ] Input validation for all user inputs
- [ ] No hardcoded secrets or credentials
- [ ] Proper error handling without information leakage
- [ ] Authentication and authorization checked
- [ ] SQL injection and XSS protection in place

### 🏷️ Traceability (T)
- [ ] Code changes linked to SPEC or issue
- [ ] @TAG references are correct and complete
- [ ] Commit message is clear and descriptive
- [ ] Documentation updated as needed
- [ ] Breaking changes are documented
```

## Low-Freedom: Automated Review Scripts

### Pre-commit Review Automation

```bash
#!/bin/bash
# .claude/skills/moai-alfred-code-reviewer/scripts/pre-review-check.sh

set -e

echo "🔍 Running automated code review checks..."

# Test Coverage Check
echo "📊 Checking test coverage..."
python -m pytest --cov=src --cov-fail-under=85 --cov-report=term-missing

# Code Quality Checks
echo "🧹 Running linters..."
python -m ruff check src/ --show-source
python -m mypy src/ --strict

# Security Scanning
echo "🔒 Scanning for security issues..."
python -m bandit -r src/ -f json -o bandit-report.json

echo "✅ All automated checks passed!"
```

## Integration with Development Workflow

### Pre-commit Integration

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: security-review
        name: Security Review
        entry: .claude/skills/moai-alfred-code-reviewer/scripts/security_review.py
        language: script
        args: [src/]
        pass_filenames: false
```

## Review Best Practices Summary

### For Reviewers
✅ **Start with understanding**: Read the PR description first  
✅ **Automate first**: Let tools catch the obvious issues  
✅ **Focus on value**: Spend time on architectural and security concerns  
✅ **Be specific**: Provide exact locations and suggestions  
✅ **Explain why**: Help the author understand the reasoning  

### For Authors
✅ **Self-review first**: Run all automated checks before submitting  
✅ **Write clear descriptions**: Explain what and why  
✅ **Keep PRs small**: Large changes are harder to review effectively  
✅ **Respond promptly**: Address feedback in a timely manner  
✅ **Learn from feedback**: Use reviews as learning opportunities  

---

## Research Integration

This skill is enhanced with advanced research capabilities to optimize code review processes and quality analysis.

### Research Areas

**Code Quality Pattern Research**:
- **Issue Pattern Analysis**: Research and analyze common code issues across different programming languages and frameworks
- **Code Smell Detection**: Research and develop pattern recognition for code quality issues
- **Best Practice Effectiveness**: Research and analyze the impact of different review practices on code quality
- **Language-Specific Optimization**: Research best practices for different programming languages and frameworks

**Review Process Optimization Research**:
- **Review Efficiency Studies**: Research and analyze the time efficiency of different review approaches
- **Automation Effectiveness**: Research the impact of automated tools on review quality and speed
- **Reviewer Training Analysis**: Research and document learning patterns for new reviewers
- **Review Coverage Optimization**: Research optimal review coverage strategies for different project types

**Security and Quality Research**:
- **Vulnerability Pattern Analysis**: Research and analyze security vulnerability patterns in code
- **Code Complexity Research**: Investigate the relationship between code complexity and maintainability
- **Test Coverage Impact**: Research the correlation between test coverage and code quality
- **Code Review ROI Analysis**: Research and measure the return on investment for different review approaches

### Research Methodology

**Quality Pattern Analysis**:
```python
# Code Quality Pattern Research
def analyze_quality_patterns():
    """Research and document code quality patterns"""

    patterns = {
        'common_vulnerabilities': [],
        'code_smells_frequency': [],
        'best_practice_adoption': [],
        'language_specific_issues': []
    }

    # Analyze code review data
    for review_data in collect_review_history():
        # Identify patterns
        # Research effectiveness of different approaches
        # Document findings for improvement
        pass
```

**Review Process Research**:
- Research and document optimal review team sizes
- Analyze review efficiency across different project types
- Research the impact of review timing on quality
- Develop benchmarking metrics for review effectiveness

**Automation Research**:
- Research and test new automation tools and techniques
- Analyze the balance between automated and manual review
- Research AI-assisted review capabilities
- Develop continuous improvement processes for automation

---

**Reference**: Code Review Best Practices, TRUST 5 Principles
**Version**: 2.0.0
