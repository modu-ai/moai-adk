# Alfred Development Guide - Reference

## SPEC-First Development Methodology

### EARS Requirements Format Reference

**EARS (Easy Approach to Requirements Syntax)** - 5 Pattern System:

#### 1. Ubiquitous Requirements (The System Shall)
**Purpose**: Define baseline system capabilities that are always present
**Format**: `The system shall [capability]`

```markdown
### Examples:
- The system shall provide user authentication via email and password.
- The system shall validate input data according to defined schemas.
- The system shall maintain audit logs for all user actions.
- The system shall support concurrent connections up to 1000 users.
- The system shall enforce role-based access control (RBAC).
```

#### 2. Event-driven Requirements (WHEN)
**Purpose**: Define system behavior in response to specific events
**Format**: `WHEN [event occurs], the system shall [response]`

```markdown
### Examples:
- WHEN a user submits login form, the system shall validate credentials.
- WHEN payment is processed successfully, the system shall send confirmation email.
- WHEN system detects unauthorized access, the system shall log security event.
- WHEN database connection fails, the system shall attempt reconnection.
- WHEN scheduled maintenance window starts, the system shall enable maintenance mode.
```

#### 3. State-driven Requirements (WHILE)
**Purpose**: Define system behavior while in specific states
**Format**: `WHILE [condition exists], the system shall [behavior]`

```markdown
### Examples:
- WHILE user is authenticated, the system shall maintain active session.
- WHILE maintenance mode is enabled, the system shall serve maintenance page.
- WHILE rate limit is exceeded, the system shall return HTTP 429.
- WHILE background job is processing, the system shall show progress indicator.
- WHILE cache is warming, the system shall serve stale data with warning.
```

#### 4. Optional Features (WHERE)
**Purpose**: Define conditional or optional system capabilities
**Format**: `WHERE [condition], the system may [optional capability]`

```markdown
### Examples:
- WHERE multi-factor authentication is enabled, the system may require OTP verification.
- WHERE user has premium subscription, the system may provide advanced analytics.
- WHERE API quota remains available, the system may allow batch operations.
- WHERE experimental features are enabled, the system may show beta functionality.
- WHERE geographic restrictions apply, the system may filter content availability.
```

#### 5. Constraints (IF)
**Purpose**: Define system limitations and conditional behavior
**Format**: `IF [condition], the system shall [enforced behavior]`

```markdown
### Examples:
- IF password is invalid 3 times, the system shall lock account for 1 hour.
- IF file upload exceeds 10MB, the system shall reject with error message.
- IF concurrent requests exceed threshold, the system shall enable throttling.
- IF database transaction fails, the system shall rollback to previous state.
- IF user session expires, the system shall redirect to login page.
```

### SPEC Template Structure

**Complete SPEC Template** (`.moai/specs/SPEC-{ID}/spec.md`):

```yaml
---
# Required Metadata
id: SPEC-{ID}
title: "{Feature Name}"
version: "0.1.0"
status: "active|completed|deprecated"
created: "YYYY-MM-DD"
updated: "YYYY-MM-DD"
priority: "high|medium|low"
author: "alfred"
reviewer: "[reviewer-name]"

# Optional Metadata
estimated_effort: "{hours} hours"
complexity: "simple|moderate|complex"
dependencies: ["SPEC-XXX", "SPEC-YYY"]
related_features: ["{feature-1}", "{feature-2}"]
tags: ["authentication", "security", "api"]
---

# {Feature Title} SPEC

## Overview
[Brief description of the feature and its purpose]

## Requirements

### Ubiquitous Requirements
[Baseline system capabilities]

### Event-driven Requirements  
[Event-response behaviors]

### State-driven Requirements
[State-dependent behaviors]

### Optional Features
[Conditional capabilities]

### Constraints
[Limitations and enforced behaviors]

## Technical Specifications

### API Endpoints
```yaml
GET /api/auth/login:
  description: "Authenticate user with email and password"
  parameters:
    - name: email
      type: string
      required: true
    - name: password
      type: string
      required: true
  responses:
    200: "Authentication successful"
    401: "Invalid credentials"
    429: "Rate limit exceeded"
```

### Database Schema
```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    locked_until TIMESTAMP,
    failed_attempts INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Performance Requirements
- Response time: <200ms for authentication requests
- Concurrent users: Support 1000+ authenticated sessions
- Throughput: 1000 login requests per minute
- Availability: 99.9% uptime during business hours

### Security Requirements
- Password hashing: bcrypt with minimum 12 rounds
- Session management: JWT tokens with 24-hour expiration
- Rate limiting: 5 attempts per minute per IP
- Audit logging: All authentication events logged

## Acceptance Criteria

### Functional Requirements
- [ ] Users can register with email and password
- [ ] Email verification is required before login
- [ ] Passwords are properly hashed and salted
- [ ] Account lockout after 3 failed attempts
- [ ] JWT tokens for session management

### Non-functional Requirements
- [ ] Response time under 200ms
- [ ] All security requirements met
- [ ] Proper error handling and logging
- [ ] Database transactions properly managed
- [ ] Rate limiting implemented

### Integration Requirements
- [ ] Email service integration for verification
- [ ] Logging service integration for audit trails
- [ ] Monitoring service integration for performance
- [ ] Cache service integration for session storage

## Testing Strategy

### Unit Tests
- Test all validation functions
- Test authentication logic
- Test error handling scenarios
- Test edge cases and boundary conditions

### Integration Tests
- Test complete authentication flow
- Test database interactions
- Test email service integration
- Test JWT token generation and validation

### Security Tests
- Test for SQL injection vulnerabilities
- Test for XSS vulnerabilities
- Test for authentication bypass
- Test for rate limiting effectiveness

### Performance Tests
- Load testing with concurrent users
- Stress testing beyond normal capacity
- Response time benchmarking
- Memory usage profiling

## Definition of Done

A feature is complete when:
1. All SPEC requirements are implemented
2. All tests pass (≥85% coverage)
3. Code review completed and approved
4. Documentation updated
5. Security review passed
6. Performance requirements met
7. Acceptance criteria verified

## Risk Assessment

### Technical Risks
- Database performance under load
- Email service reliability
- Session storage scalability
- Token generation security

### Mitigation Strategies
- Implement database connection pooling
- Use multiple email service providers
- Implement distributed session storage
- Use industry-standard JWT libraries

## History

### v0.1.0 (YYYY-MM-DD)
- Initial draft with core authentication requirements
- Defined basic EARS patterns and acceptance criteria

### v0.2.0 (YYYY-MM-DD)
- Added performance requirements
- Updated security specifications
- Added integration requirements
```

## Alfred Command Reference

### Core Commands

#### `/alfred:0-project` - Project Initialization
```bash
# Interactive project setup
/alfred:0-project

# Initialize with specific configuration
/alfred:0-project --mode team --language python --framework django

# Update existing project configuration
/alfred:0-project setting
```

**Actions Performed**:
- Create `.moai/` directory structure
- Generate `.claude/` configuration
- Set up initial project metadata
- Configure Alfred preferences
- Initialize Git hooks (if applicable)

#### `/alfred:1-plan` - SPEC Creation
```bash
# Create new feature SPEC
/alfred:1-plan "User authentication system with email verification"

# Plan from existing requirements
/alfred:1-plan --from .moai/requirements/auth-requirements.md

# Plan with specific complexity
/alfred:1-plan "Payment processing" --complexity high
```

**Actions Performed**:
- Analyze feature requirements
- Create SPEC with EARS format
- Generate acceptance criteria
- Estimate effort and complexity
- Create feature branch

#### `/alfred:2-run` - TDD Implementation
```bash
# Implement specific SPEC
/alfred:2-run SPEC-AUTH-001

# Run with specific focus
/alfred:2-run SPEC-AUTH-001 --focus security

# Continue from specific phase
/alfred:2-run SPEC-AUTH-001 --from refactor
```

**Actions Performed**:
- Load SPEC and context
- Execute RED-GREEN-REFACTOR cycle
- Generate comprehensive tests
- Implement production code
- Validate TRUST principles

#### `/alfred:3-sync` - Documentation Synchronization
```bash
# Automatic sync for SPEC
/alfred:3-sync auto SPEC-AUTH-001

# Sync specific components
/alfred:3-sync docs SPEC-AUTH-001
/alfred:3-sync tests SPEC-AUTH-001
/alfred:3-sync changelog SPEC-AUTH-001

# Generate comprehensive report
/alfred:3-sync report SPEC-AUTH-001
```

**Actions Performed**:
- Update SPEC status and metadata
- Synchronize documentation
- Generate sync report
- Create pull request
- Validate TAG chain integrity

### Advanced Commands

#### Context Management
```bash
# Show current context usage
/alfred:context status

# Optimize context for phase
/alfred:context optimize --phase implementation

# Clear cached context
/alfred:context clear --all
```

#### Quality Gates
```bash
# Run TRUST 5 validation
/alfred:quality trust SPEC-AUTH-001

# Check security compliance
/alfred:quality security SPEC-AUTH-001

# Validate test coverage
/alfred:quality coverage SPEC-AUTH-001
```

#### TAG Management
```bash
# Verify TAG chain integrity
/alfred:tags validate SPEC-AUTH-001

# Find orphaned TAGs
/alfred:tags orphaned

# Generate TAG report
/alfred:tags report --format markdown
```

## Configuration Reference

### Alfred Configuration (`.moai/config/config.json`)

```json
{
  "project": {
    "name": "My Application",
    "owner": "team-name",
    "repository": "https://github.com/team-name/my-app",
    "mode": "team|individual",
    "language": "python|javascript|go|rust|java",
    "framework": "django|flask|express|spring|gin"
  },
  
  "alfred": {
    "persona": "professional|friendly|technical",
    "context_optimization": true,
    "auto_save": true,
    "traceability": {
      "enabled": true,
      "enforce_chains": true,
      "validation_level": "strict|lenient"
    }
  },
  
  "quality": {
    "trust_principles": {
      "enabled": true,
      "min_test_coverage": 85,
      "max_complexity": 10,
      "require_docstrings": true
    },
    "security": {
      "scan_secrets": true,
      "validate_inputs": true,
      "owasp_compliance": true
    },
    "performance": {
      "response_time_limit": 200,
      "memory_limit": "512MB",
      "concurrent_users": 1000
    }
  },
  
  "workflow": {
    "git": {
      "strategy": "gitflow|trunk|feature-branch",
      "auto_commit": true,
      "commit_template": "conventional|custom"
    },
    "review": {
      "require_approval": true,
      "min_reviewers": 1,
      "auto_assign": true
    },
    "deployment": {
      "environments": ["staging", "production"],
      "auto_deploy": false,
      "health_checks": true
    }
  },
  
  "integration": {
    "email": {
      "provider": "sendgrid|aws-ses|mailgun",
      "templates": true,
      "analytics": false
    },
    "monitoring": {
      "provider": "datadog|newrelic|prometheus",
      "error_tracking": true,
      "performance_metrics": true
    },
    "storage": {
      "provider": "aws-s3|gcp-storage|azure-blob",
      "cdn": true,
      "backup": true
    }
  },
  
  "reporting": {
    "enabled": true,
    "auto_generate": true,
    "formats": ["markdown", "json", "html"],
    "recipients": ["team@company.com"],
    "frequency": "daily|weekly|monthly"
  }
}
```

### Git Hooks Configuration

#### Pre-commit Hook (`.git/hooks/pre-commit`)
```bash
#!/bin/bash
# Alfred pre-commit validation

echo "🔍 Running Alfred pre-commit checks..."

# Check for required TAGs
echo "  → Verifying TAG chain integrity..."
if ! rg '@(SPEC|TEST|CODE|DOC):' --stats . > /dev/null 2>&1; then
    echo "  ❌ Missing required TAGs"
    exit 1
fi

# Validate test coverage
echo "  → Checking test coverage..."
coverage=$(python -m pytest --cov=src --cov-report=json | grep -o '"total_coverage": [0-9.]*' | cut -d: -f2 | tr -d ' ')
if (( $(echo "$coverage < 85" | bc -l) )); then
    echo "  ❌ Test coverage ${coverage}% is below 85%"
    exit 1
fi

# Check for secrets
echo "  → Scanning for secrets..."
if rg '(password|secret|key|token)\s*=\s*["\'][^"\']+["\']' --type py src/ > /dev/null 2>&1; then
    echo "  ❌ Potential hardcoded secrets detected"
    exit 1
fi

echo "  ✅ All pre-commit checks passed"
```

#### Pre-push Hook (`.git/hooks/pre-push`)
```bash
#!/bin/bash
# Alfred pre-push quality gates

echo "🚀 Running Alfred quality gates..."

# Run full test suite
echo "  → Running test suite..."
if ! python -m pytest --cov=src --cov-fail-under=85; then
    echo "  ❌ Test suite failed or coverage below 85%"
    exit 1
fi

# Security scan
echo "  → Running security scan..."
if ! bandit -r src/ -f json; then
    echo "  ❌ Security issues detected"
    exit 1
fi

# Code quality check
echo "  → Running code quality checks..."
if ! flake8 src/ --max-complexity=10; then
    echo "  ❌ Code quality issues detected"
    exit 1
fi

echo "  ✅ All quality gates passed"
```

## Integration Patterns

### CI/CD Pipeline Integration

#### GitHub Actions Workflow (`.github/workflows/alfred.yml`)
```yaml
name: Alfred Quality Gates

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      
      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          pip install -r requirements-dev.txt
      
      - name: Validate SPEC compliance
        run: |
          python -m alfred.cli validate-specs
          python -m alfred.cli check-tags
      
      - name: Run TRUST 5 validation
        run: |
          python -m alfred.cli validate-trust
      
      - name: Execute test suite
        run: |
          python -m pytest --cov=src --cov-report=xml
      
      - name: Upload coverage reports
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.xml
      
      - name: Security scan
        run: |
          python -m bandit -r src/ -f json -o bandit-report.json
      
      - name: Quality metrics
        run: |
          python -m flake8 src/ --format=json --output-file=flake8-report.json
          python -m alfred.cli generate-quality-report

  deploy:
    needs: validate
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy to production
        run: |
          python -m alfred.cli deploy --environment production
```

### Monitoring Integration

#### Alfred Metrics Dashboard
```python
# src/monitoring/alfred_metrics.py
from dataclasses import dataclass
from typing import Dict, List
import json
from datetime import datetime

@dataclass
class QualityMetrics:
    spec_compliance: float
    test_coverage: float
    code_quality: float
    security_score: float
    performance_score: float
    
    def to_dict(self) -> Dict:
        return {
            "spec_compliance": self.spec_compliance,
            "test_coverage": self.test_coverage,
            "code_quality": self.code_quality,
            "security_score": self.security_score,
            "performance_score": self.performance_score,
            "overall_score": (self.spec_compliance + self.test_coverage + 
                             self.code_quality + self.security_score + 
                             self.performance_score) / 5
        }

class AlfredMetricsCollector:
    """Collects and reports Alfred development metrics."""
    
    def collect_quality_metrics(self, project_path: str) -> QualityMetrics:
        """Collect comprehensive quality metrics."""
        
        # SPEC compliance metrics
        spec_score = self._calculate_spec_compliance(project_path)
        
        # Test coverage metrics
        coverage_score = self._get_test_coverage(project_path)
        
        # Code quality metrics
        quality_score = self._calculate_code_quality(project_path)
        
        # Security metrics
        security_score = self._calculate_security_score(project_path)
        
        # Performance metrics
        performance_score = self._calculate_performance_score(project_path)
        
        return QualityMetrics(
            spec_compliance=spec_score,
            test_coverage=coverage_score,
            code_quality=quality_score,
            security_score=security_score,
            performance_score=performance_score
        )
    
    def _calculate_spec_compliance(self, project_path: str) -> float:
        """Calculate SPEC compliance percentage."""
        # Count total required vs implemented requirements
        total_requirements = rg 'The system shall|WHEN|WHILE|WHERE|IF' project_path + '/.moai/specs/' | wc -l
        implemented_features = rg '@(TEST|CODE):' project_path + '/src/' | wc -l
        
        if total_requirements == 0:
            return 100.0
            
        return min(100.0, (implemented_features / total_requirements) * 100)
    
    def _get_test_coverage(self, project_path: str) -> float:
        """Extract test coverage from pytest/cov reports."""
        # Implementation depends on your test setup
        # This is a placeholder for actual coverage extraction
        return 92.0  # Example value
    
    def _calculate_code_quality(self, project_path: str) -> float:
        """Calculate code quality based on various metrics."""
        # Combine complexity, maintainability, duplication metrics
        return 88.5  # Example value
    
    def _calculate_security_score(self, project_path: str) -> float:
        """Calculate security compliance score."""
        # Check for vulnerabilities, secrets, etc.
        return 95.0  # Example value
    
    def _calculate_performance_score(self, project_path: str) -> float:
        """Calculate performance metrics."""
        # Response time, throughput, resource usage
        return 90.0  # Example value
    
    def generate_report(self, metrics: QualityMetrics, output_path: str):
        """Generate comprehensive quality report."""
        report = {
            "timestamp": datetime.now().isoformat(),
            "metrics": metrics.to_dict(),
            "trends": self._get_historical_trends(),
            "recommendations": self._generate_recommendations(metrics)
        }
        
        with open(output_path, 'w') as f:
            json.dump(report, f, indent=2)
```

## Troubleshooting Guide

### Common Issues and Solutions

#### TAG Chain Validation Failures
```bash
# Problem: Missing TAG chain
Error: Orphaned CODE:AUTH-001 found - no corresponding SPEC:AUTH-001

# Solution: Create missing SPEC or fix TAG reference
/alfred:1-plan "Complete authentication system" --id AUTH-001

# Verify TAG chain integrity
/alfred:tags validate AUTH-001
```

#### Context Overflow Errors
```bash
# Problem: Context exceeds limit
Error: Context window exceeded (8000 tokens max)

# Solution: Optimize context loading
/alfred:context optimize --phase implementation
/alfred:context clear --cache

# Check context usage
/alfred:context status
```

#### Test Coverage Failures
```bash
# Problem: Test coverage below 85%
Error: Test coverage 78% is below required 85%

# Solution: Identify missing coverage
pytest --cov=src --cov-report=term-missing
python -m pytest --cov=src --cov-report=html

# Generate coverage report
coverage html
# Open htmlcov/index.html to see detailed coverage
```

#### Security Scan Failures
```bash
# Problem: Security vulnerabilities detected
Error: Bandit found 2 high-severity issues

# Solution: Review and fix security issues
bandit -r src/ -f json -o bandit-report.json
cat bandit-report.json | jq '.results[] | select(.issue_severity == "HIGH")'

# Fix issues and re-run
bandit -r src/
```

#### Performance Issues
```bash
# Problem: Slow response times
Error: Average response time 450ms exceeds 200ms limit

# Solution: Profile and optimize
python -m cProfile -o profile.stats src/main.py
python -c "import pstats; p = pstats.Stats('profile.stats'); p.sort_stats('cumulative'); p.print_stats(20)"

# Use performance monitoring tools
pip install py-spy
py-spy top --pid $(pgrep -f "python src/main.py")
```

---

*Complete technical reference for Alfred development*  
*Enterprise-grade specifications and configurations*  
*Best practices and troubleshooting guidance*
