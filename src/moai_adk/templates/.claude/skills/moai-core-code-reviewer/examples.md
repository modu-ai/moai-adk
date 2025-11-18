# Code Review Examples

## Example 1: Basic Code Review Process

```python
# Before invoking the skill
# You have a pull request that needs review

# Invoke the skill
Skill("moai-core-code-reviewer")

# The skill provides:
# 1. Automated review checklist
# 2. TRUST 5 framework guidance
# 3. Security review scripts
# 4. Review comment templates

# Run automated checks
bash .claude/skills/moai-core-code-reviewer/scripts/pre-review-check.sh

# Review output example:
# 🔍 Running automated code review checks...
# 📊 Checking test coverage...
# ✅ Test coverage meets requirements (92%)
# 🧹 Running ruff linter...
# ✅ Ruff checks passed
# 🔍 Running type checks...
# ✅ Type checks passed
# 🔒 Scanning for security issues...
# ✅ Security scan completed
# 🎉 Automated review checks completed!
```

## Example 2: Security-focused Review

```python
# When reviewing code with security implications
Skill("moai-core-code-reviewer")

# Focus on security checklist items:
# - Input validation for all user inputs
# - No hardcoded secrets or credentials  
# - Proper error handling without information leakage
# - Authentication and authorization checked
# - SQL injection and XSS protection in place

# Example security issues found:
# ❌ Potential SQL injection in user_query() function
# ❌ Hardcoded API key in config.py
# ❌ Debug mode enabled in production
```

## Example 3: Integration with GitHub Actions

```yaml
# .github/workflows/code-review.yml
name: Automated Code Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  automated-review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Code Review Skill
        run: |
          # Skill provides automation scripts
          bash .claude/skills/moai-core-code-reviewer/scripts/pre-review-check.sh
      
      - name: Security Review
        run: |
          python .claude/skills/moai-core-code-reviewer/scripts/security_review.py src/
```

## Example 4: Team Review Standards

```python
# Setting up team review standards
Skill("moai-core-code-reviewer")

# Provides templates for:
# 1. Review checklist customization
# 2. Comment templates by severity
# 3. Review metrics tracking
# 4. Quality gates automation

# Custom checklist example:
## Frontend Review Standards
- [ ] Component follows design system
- [ ] Accessibility (WCAG 2.1) compliance
- [ ] Performance metrics met (Lighthouse score > 90)
- [ ] Responsive design tested
- [ ] Cross-browser compatibility verified
```

## Example 5: Review Quality Metrics

```python
# Track review quality over time
Skill("moai-core-code-reviewer")

# Provides metrics calculation:
# - Review coverage percentage
# - Average review time
# - Issue detection rate
# - Reviewer consistency

# Example metrics dashboard:
Review Coverage: 95% ✅
Average Review Time: 4.2 hours
Critical Issues Found: 12
Quality Score: 8.7/10
```
