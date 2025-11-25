# Claude Code Core Integration Examples

## Quick Start Examples

### Example 1: Basic Skill Creation
```bash
# Create new skill using template
mkdir .claude/skills/my-domain-skill
cp moai-cc-core/templates/skill-template.md .claude/skills/my-domain-skill/SKILL.md

# Validate skill structure
python .claude/skills/moai-cc-core/scripts/skill-validator.py --skill-path .claude/skills/my-domain-skill
```

### Example 2: Memory Optimization
```bash
# Analyze current memory usage
python .claude/skills/moai-cc-core/scripts/memory-optimizer.py --analyze

# Optimize and cleanup memory
python .claude/skills/moai-cc-core/scripts/memory-optimizer.py --optimize --aggressive
```

### Example 3: Configuration Management
```bash
# Switch to personal GitHub workflow
python .claude/skills/moai-cc-core/scripts/config-manager.py --git-mode personal

# Create production environment config
python .claude/skills/moai-cc-core/scripts/config-manager.py --create-env production
```

## Command System Examples

### Example 4: Advanced Command with Context
```markdown
---
name: moai:deploy-feature
description: "Deploy feature with comprehensive validation and rollback support"
argument-hint: "feature-name --env staging|production"
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
model: "sonnet"
skills:
  - moai-cc-core
  - moai-foundation-trust
---

# Feature Deployment Command

## 📋 Pre-execution Context

!git status --porcelain
!git log --oneline -5
!npm test  # Validate build

## 📁 Essential Files

@.moai/config/config.json
@package.json
@docker-compose.yml

---

# Implementation

## 🔍 Pre-deployment Validation
1. **Code Quality**: Run linting and tests
2. **Security Scan**: Check for vulnerabilities
3. **Performance Test**: Validate performance metrics
4. **Documentation**: Ensure docs are updated

## 🚀 Deployment Process
1. **Environment Validation**: Confirm target environment
2. **Backup Creation**: Create rollback point
3. **Feature Deployment**: Deploy with monitoring
4. **Health Check**: Validate deployment success
5. **Rollback if Needed**: Automated rollback on failure

## 📊 Deployment Report
- Deployment ID: auto-generated
- Status: Success/Failure/Rolled Back
- Metrics: Performance, uptime, error rates
- Rollback Available: Yes/No
```

### Example 5: Command with Model Optimization
```markdown
---
name: moai:generate-report
description: "Generate comprehensive project status report"
model: "haiku"  # 70% cost savings for template-based task
---

# Report Generation

## 📊 Auto-generated Report Sections
1. **Project Status**: Current health and metrics
2. **Test Coverage**: Coverage reports and gaps
3. **Security Scan**: Vulnerability assessment
4. **Performance Metrics**: Resource usage and trends
5. **Documentation Status**: API docs and guides completeness

## 🎯 Output Format
- **Console**: Summary overview
- **File**: Detailed JSON report
- **Visualization**: Charts and graphs (if data available)
```

## Skill Development Examples

### Example 6: Domain-Specific Skill Structure
```markdown
---
name: moai-domain-ecommerce
description: "E-commerce domain expertise with payment processing, inventory management, and customer analytics patterns. Use when building e-commerce platforms, payment integrations, or retail analytics systems."
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
version: 1.0.0
modularized: true
tags:
  - ecommerce
  - payment
  - inventory
  - analytics
---

# E-commerce Domain Expertise

## Quick Reference (30 seconds)

**Enterprise E-commerce Platform Architecture** - Specialized expertise for building scalable e-commerce solutions with payment processing, inventory management, customer analytics, and omnichannel retail patterns.

**Core Capabilities**:
- ✅ Payment processing with multiple providers (Stripe, PayPal, Square)
- ✅ Inventory management with real-time synchronization
- ✅ Customer analytics and behavior tracking
- ✅ Order fulfillment and shipping automation
- ✅ Tax calculation and compliance management

**When to Use**:
- Building e-commerce platforms or marketplaces
- Integrating payment processing systems
- Implementing inventory and order management
- Creating customer analytics and personalization
- Setting up omnichannel retail experiences
```

### Example 7: Skill with Progressive Disclosure
```markdown
---
name: moai-domain-ml-ops
description: "Machine Learning Operations expertise with model deployment, monitoring, and lifecycle management. Use when implementing ML workflows, model serving, or MLOps pipelines."
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
modularized: true
---

# MLOps Domain Expertise

## Quick Reference (30 seconds)

**Enterprise MLOps Platform** - ML workflow orchestration with model deployment, monitoring, and lifecycle management for production machine learning systems.

**Core Capabilities**:
- ✅ Model deployment and versioning (MLflow integration)
- ✅ Real-time monitoring and drift detection
- ✅ Automated retraining and A/B testing
- ✅ Feature engineering and data pipelines
- ✅ Model governance and compliance tracking

## Implementation Guide

### What It Does

**Model Deployment**:
- Containerized model serving with Docker/Kubernetes
- API gateway integration with load balancing
- Blue-green deployment strategies
- Automated rollback capabilities

**Monitoring & Observability**:
- Real-time performance metrics and alerting
- Data drift detection and model degradation
- Explainability and fairness monitoring
- Resource usage and cost tracking

**Lifecycle Management**:
- Automated model retraining pipelines
- Continuous integration for ML models
- Model registry and versioning
- Experiment tracking and reproducibility
```

## Configuration Management Examples

### Example 8: Multi-Environment Configuration
```json
{
  "project": {
    "name": "E-commerce Platform",
    "environment": "production"
  },
  "git_strategy": {
    "mode": "team",
    "branch_creation": {
      "prompt_always": false,
      "auto_enabled": true
    }
  },
  "database": {
    "primary": {
      "host": "{{DB_HOST}}",
      "port": 5432,
      "name": "ecommerce_prod"
    },
    "cache": {
      "redis_url": "{{REDIS_URL}}"
    }
  },
  "payment": {
    "providers": ["stripe", "paypal"],
    "webhook_secret": "{{PAYMENT_WEBHOOK_SECRET}}"
  },
  "monitoring": {
    "enabled": true,
    "alert_webhook": "{{ALERT_WEBHOOK_URL}}"
  }
}
```

### Example 9: Git Strategy Configuration
```json
{
  "git_strategy": {
    "mode": "team",
    "environment": "github",
    "workflow": "github-flow",
    "branch_creation": {
      "prompt_always": false,
      "auto_enabled": true
    },
    "automation": {
      "auto_branch": true,
      "auto_commit": true,
      "auto_push": true,
      "auto_pr": true,
      "draft_pr": true
    },
    "team": {
      "required_reviews": 2,
      "branch_protection": true,
      "auto_merge": false,
      "status_checks": ["test", "security-scan", "performance-test"]
    }
  }
}
```

## Memory Management Examples

### Example 10: Context Optimization Strategy
```python
# Phase 1: SPEC Generation (30K tokens)
spec_context = {
    "essential_files": [
        ".moai/config/config.json",
        "CLAUDE.md",
        "README.md"
    ],
    "token_budget": 30000,
    "compression": True
}

# Execute SPEC generation
result = await Task("spec-builder", "Generate SPEC", spec_context)

# Clear context before next phase
await execute("/clear")

# Phase 2: Implementation (180K tokens)
impl_context = {
    "spec_content": result["spec"],
    "implementation_files": ["src/", "tests/"],
    "token_budget": 180000,
    "progressive_loading": True
}
```

### Example 11: Working Memory Management
```python
class WorkingMemoryManager:
    def __init__(self, max_tokens: int = 50000):
        self.max_tokens = max_tokens
        self.active_context = {}
        self.priority_queue = []

    def add_context(self, key: str, content: str, priority: int = 5):
        """Add content to working memory with priority."""
        tokens = self.estimate_tokens(content)

        if self.get_current_usage() + tokens > self.max_tokens:
            self.evict_low_priority()

        self.active_context[key] = {
            "content": content,
            "tokens": tokens,
            "priority": priority,
            "last_accessed": datetime.now()
        }

    def get_essential_context(self) -> str:
        """Get highest priority context within token limit."""
        sorted_items = sorted(
            self.active_context.items(),
            key=lambda x: x[1]["priority"],
            reverse=True
        )

        context_parts = []
        total_tokens = 0

        for key, item in sorted_items:
            if total_tokens + item["tokens"] <= self.max_tokens:
                context_parts.append(item["content"])
                total_tokens += item["tokens"]
            else:
                break

        return "\n\n".join(context_parts)
```

## Hooks System Examples

### Example 12: Pre-Execution Security Hook
```python
#!/usr/bin/env python3
"""
Security validation hook for Bash commands
"""

import subprocess
import re
from typing import List, Tuple

class SecurityHook:
    def __init__(self):
        self.forbidden_commands = [
            "rm -rf",
            "sudo rm",
            "chmod 777",
            "dd if=",
            "mkfs",
            "shutdown",
            "reboot"
        ]
        self.sensitive_paths = [
            "/etc/",
            "/usr/bin/",
            "/bin/",
            "~/.ssh/",
            ".env*"
        ]

    def validate_command(self, command: str) -> Tuple[bool, List[str]]:
        """Validate bash command for security issues."""
        issues = []

        # Check forbidden commands
        for forbidden in self.forbidden_commands:
            if forbidden in command:
                issues.append(f"Forbidden command detected: {forbidden}")

        # Check sensitive path access
        for path in self.sensitive_paths:
            if re.search(rf'\s+{path}.*\s', command):
                issues.append(f"Sensitive path access: {path}")

        # Check for sudo usage
        if "sudo" in command and not self.is_sudo_allowed(command):
            issues.append("Unauthorized sudo usage detected")

        return len(issues) == 0, issues

    def is_sudo_allowed(self, command: str) -> bool:
        """Check if sudo usage is allowed in this context."""
        allowed_sudo_commands = [
            "sudo apt-get",
            "sudo yum",
            "sudo pip",
            "sudo docker"
        ]
        return any(allowed in command for allowed in allowed_sudo_commands)

# Hook execution
if __name__ == "__main__":
    import sys
    command = " ".join(sys.argv[1:])

    hook = SecurityHook()
    is_valid, issues = hook.validate_command(command)

    if not is_valid:
        print("SECURITY VIOLATION:")
        for issue in issues:
            print(f"  - {issue}")
        sys.exit(1)
    else:
        print("SECURITY CHECK PASSED")
        sys.exit(0)
```

### Example 13: Post-Execution Code Optimization Hook
```python
#!/usr/bin/env python3
"""
Code optimization hook for Edit/Write operations
"""

import ast
import subprocess
from pathlib import Path
from typing import Dict, List

class CodeOptimizer:
    def __init__(self):
        self.optimizations = [
            "formatting",
            "imports",
            "complexity",
            "security"
        ]

    def optimize_file(self, file_path: Path) -> Dict[str, any]:
        """Optimize a code file and return results."""
        results = {
            "file": str(file_path),
            "optimizations": {},
            "issues": []
        }

        try:
            content = file_path.read_text()

            # Apply optimizations
            if self.format_code(file_path):
                results["optimizations"]["formatting"] = "Applied"

            if self.optimize_imports(file_path):
                results["optimizations"]["imports"] = "Optimized"

            # Analyze complexity
            complexity = self.analyze_complexity(content)
            results["optimizations"]["complexity"] = complexity

            # Security scan
            security_issues = self.security_scan(content)
            if security_issues:
                results["issues"].extend(security_issues)

        except Exception as e:
            results["issues"].append(f"Optimization failed: {e}")

        return results

    def format_code(self, file_path: Path) -> bool:
        """Format code using appropriate tool."""
        if file_path.suffix == ".py":
            try:
                subprocess.run(["black", str(file_path)], check=True, capture_output=True)
                subprocess.run(["isort", str(file_path)], check=True, capture_output=True)
                return True
            except subprocess.CalledProcessError:
                return False
        elif file_path.suffix in [".js", ".ts", ".jsx", ".tsx"]:
            try:
                subprocess.run(["prettier", "--write", str(file_path)], check=True, capture_output=True)
                return True
            except subprocess.CalledProcessError:
                return False
        return False

    def optimize_imports(self, file_path: Path) -> bool:
        """Optimize imports in Python files."""
        if file_path.suffix == ".py":
            try:
                # Remove unused imports
                subprocess.run(["autoflake", "--in-place", "--remove-unused-variables", "--remove-all-unused-imports", str(file_path)],
                             check=True, capture_output=True)
                return True
            except subprocess.CalledProcessError:
                return False
        return False

    def analyze_complexity(self, content: str) -> Dict[str, int]:
        """Analyze code complexity."""
        try:
            tree = ast.parse(content)
            complexity = {
                "functions": 0,
                "classes": 0,
                "max_complexity": 0,
                "lines_of_code": len(content.split('\n'))
            }

            for node in ast.walk(tree):
                if isinstance(node, ast.FunctionDef):
                    complexity["functions"] += 1
                    func_complexity = self._calculate_complexity(node)
                    complexity["max_complexity"] = max(complexity["max_complexity"], func_complexity)
                elif isinstance(node, ast.ClassDef):
                    complexity["classes"] += 1

            return complexity
        except SyntaxError:
            return {"error": "Invalid syntax"}

    def _calculate_complexity(self, node: ast.FunctionDef) -> int:
        """Calculate cyclomatic complexity."""
        complexity = 1
        for child in ast.walk(node):
            if isinstance(child, (ast.If, ast.While, ast.For, ast.Try)):
                complexity += 1
        return complexity

    def security_scan(self, content: str) -> List[str]:
        """Basic security scan for common issues."""
        issues = []

        # Check for hardcoded secrets
        import re
        secret_patterns = [
            r'password\s*=\s*["\'][^"\']+["\']',
            r'api_key\s*=\s*["\'][^"\']+["\']',
            r'secret\s*=\s*["\'][^"\']+["\']'
        ]

        for pattern in secret_patterns:
            if re.search(pattern, content, re.IGNORECASE):
                issues.append("Potential hardcoded secret detected")

        # Check for eval usage
        if re.search(r'\beval\s*\(', content):
            issues.append("Usage of eval() detected (potential security risk)")

        return issues

# Hook execution
if __name__ == "__main__":
    import sys
    file_path = Path(sys.argv[1])

    optimizer = CodeOptimizer()
    result = optimizer.optimize_file(file_path)

    print(f"OPTIMIZATION RESULTS for {result['file']}:")
    if result['optimizations']:
        print("Optimizations applied:")
        for opt_type, status in result['optimizations'].items():
            print(f"  - {opt_type}: {status}")

    if result['issues']:
        print("Issues found:")
        for issue in result['issues']:
            print(f"  - {issue}")
```

## Integration Examples

### Example 14: Complete Workflow Integration
```bash
# 1. Initialize project with Claude Code core
/moai:0-project

# 2. Generate SPEC for new feature
/moai:1-plan "Implement payment processing with Stripe integration"

# 3. Clear context (mandatory after SPEC generation)
/clear

# 4. Implement using TDD
/moai:2-run SPEC-001

# 5. Validate skill quality
python .claude/skills/moai-cc-core/scripts/skill-validator.py --validate-all

# 6. Optimize memory usage
python .claude/skills/moai-cc-core/scripts/memory-optimizer.py --optimize

# 7. Sync documentation
/moai:3-sync SPEC-001

# 8. Collect feedback
/moai:9-feedback

# 9. Deploy with validation
/moai:deploy-feature payment-processing --env staging
```

### Example 15: Multi-Team Workflow
```json
{
  "git_strategy": {
    "mode": "team",
    "team": {
      "required_reviews": 2,
      "auto_assign_reviewers": ["backend-lead", "security-lead"],
      "status_checks": {
        "required": ["test-suite", "security-scan", "performance-test"],
        "timeout_minutes": 30
      },
      "merge_strategy": "squash",
      "delete_branch_after_merge": true
    }
  },
  "quality_gates": {
    "test_coverage": 90,
    "security_scan": "strict",
    "performance_baseline": "enforced"
  },
  "notifications": {
    "slack_webhook": "{{SLACK_WEBHOOK_URL}}",
    "email_on_failure": true,
    "review_reminder_hours": 24
  }
}
```

---

**Examples Count**: 15 practical examples covering all core integration patterns
**Last Updated**: 2025-11-24
**Complexity**: Basic to advanced enterprise scenarios