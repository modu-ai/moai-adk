# Extensions and Customization Guide

Comprehensive guide to extending MoAI-ADK with custom Skills, Agents, Commands, and Hooks.

## Table of Contents

- [Extension Architecture](#extension-architecture)
- [Creating Custom Skills](#creating-custom-skills)
- [Creating Custom Agents](#creating-custom-agents)
- [Creating Custom Commands](#creating-custom-commands)
- [Creating Custom Hooks](#creating-custom-hooks)
- [Integration Patterns](#integration-patterns)
- [Real-World Examples](#real-world-examples)
- [Best Practices](#best-practices)

______________________________________________________________________

## Extension Architecture

### Four Extension Points

```
MoAI-ADK Extension Points
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Commands
   Purpose: User-facing workflow orchestration
   Location: .claude/commands/
   Format: Markdown (.md)
   Example: /custom-deploy, /custom-analyze

2. Agents
   Purpose: Domain-specific reasoning
   Location: .claude/agents/
   Format: Markdown (.md) or Python (.py)
   Example: ml-expert, data-analyst

3. Skills
   Purpose: Reusable knowledge capsules
   Location: .claude/skills/
   Format: Markdown (.md)
   Example: moai-custom-mlops

4. Hooks
   Purpose: Event-driven automation
   Location: .claude/hooks/
   Format: Shell (.sh) or Python (.py)
   Example: pre-commit, post-deploy
```

______________________________________________________________________

## Creating Custom Skills

### Skill Anatomy

A Skill consists of:
- **Knowledge**: Core information (<500 words)
- **Examples**: Practical use cases
- **References**: Links to detailed docs

### Directory Structure

```
.claude/skills/
└── moai-custom-skillname/
    ├── skill.yaml          # Metadata
    ├── README.md           # Main documentation
    ├── prompt.md           # Core prompt
    ├── examples.md         # Usage examples
    └── resources/
        ├── templates/      # Code templates
        ├── checklists/     # Validation checklists
        └── references/     # External links
```

### Example: Document Validator Skill

**Step 1: Create skill.yaml**

```yaml
# .claude/skills/moai-custom-doc-validator/skill.yaml
name: moai-custom-doc-validator
version: 1.0.0
description: Validates markdown documentation quality
author: team
tags:
  - documentation
  - validation
  - quality
tier: custom
freedom_level: medium
triggers:
  - document validation
  - quality check
  - markdown lint
dependencies:
  - moai-foundation-tags
  - moai-foundation-best-practices
```

**Step 2: Create prompt.md**

```markdown
# Document Validator Skill

You are an expert markdown documentation validator.

## Validation Criteria

### 1. Structure (Weight: 30%)
- Single H1 heading
- Logical heading hierarchy (no skipped levels)
- Table of contents for documents >300 lines

### 2. Content Quality (Weight: 40%)
- Clear, concise writing
- No broken links
- Code blocks have language specified
- Examples are runnable

### 3. Formatting (Weight: 20%)
- Consistent list markers (- or *)
- Proper table formatting
- No trailing whitespace
- UTF-8 encoding

### 4. Compliance (Weight: 10%)
- Follows project style guide
- Includes required sections
- TAG references valid

## Validation Process

1. Scan document structure
2. Check all links (internal and external)
3. Validate code blocks
4. Verify formatting consistency
5. Generate quality score (0-10)

## Output Format

```
Document Validation Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

File: path/to/document.md
Quality Score: 8.5/10

Structure: 9/10
  ✓ Single H1 heading
  ✓ Logical hierarchy
  ⚠ Missing table of contents

Content: 8/10
  ✓ Clear writing
  ✗ 2 broken links
  ✓ All code blocks have language

Formatting: 9/10
  ✓ Consistent list markers
  ✓ Valid tables
  ✓ No trailing whitespace

Compliance: 8/10
  ✓ Follows style guide
  ✓ Required sections present
  ⚠ 1 invalid TAG reference

Priority Fixes:
  1. Fix broken links (lines 45, 67)
  2. Add table of contents
  3. Fix TAG reference (line 123)
```
```

**Step 3: Create README.md**

```markdown
# Custom Document Validator

Automatically validates markdown documentation quality.

## Installation

```bash
# Skill is auto-discovered from .claude/skills/
# No installation needed
```

## Usage

```
Skill("moai-custom-doc-validator")
```

## Use Cases

1. **Pre-commit validation**: Check docs before commit
2. **CI/CD integration**: Automated quality gate
3. **Documentation audit**: Periodic quality review

## Configuration

```json
{
  "custom_skills": {
    "moai-custom-doc-validator": {
      "enabled": true,
      "quality_threshold": 8.0,
      "auto_fix": false
    }
  }
}
```

## Examples

### Validate Single File

```python
# Alfred automatically uses skill when you say:
"Validate the quality of docs/guide.md"
```

### Validate All Documentation

```python
"Run document validation on all markdown files in docs/"
```

### Generate Quality Report

```python
"Create quality report for documentation"
```
```

**Step 4: Create examples.md**

```markdown
# Examples

## Example 1: Basic Validation

Input:
```
Validate docs/getting-started.md
```

Output:
```
Quality Score: 9.2/10
✓ Excellent structure
✓ All links valid
⚠ Minor formatting issues
```

## Example 2: Batch Validation

Input:
```
Validate all files in docs/guides/
```

Output:
```
Validated 12 files
Average Quality: 8.7/10
3 files need attention:
  - docs/guides/advanced.md (7.5/10)
  - docs/guides/deployment.md (7.8/10)
  - docs/guides/troubleshooting.md (8.0/10)
```

## Example 3: CI/CD Integration

```yaml
# .github/workflows/docs-validation.yml
name: Documentation Quality
on: [pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Validate docs
        run: |
          # Alfred validates via custom skill
          moai-adk validate-docs
```
```

### Registering the Skill

```json
// .moai/config.json
{
  "custom_skills": {
    "moai-custom-doc-validator": {
      "enabled": true,
      "version": "1.0.0",
      "auto_load": false  // Load on-demand only
    }
  }
}
```

______________________________________________________________________

## Creating Custom Agents

### Agent Architecture

An agent has:
- **Identity**: Name, role, specialty
- **Tools**: Available functions (Read, Bash, etc.)
- **Decision Logic**: How it makes choices
- **Collaboration**: How it works with other agents

### Example: ML Model Validator Agent

**File: `.claude/agents/ml-validator.md`**

```markdown
# ML Model Validator Agent

## Identity

- **Name**: ml-validator
- **Role**: Machine learning model validation expert
- **Specialty**: Model performance analysis, overfitting detection, optimization
- **Model**: Sonnet (complex reasoning required)
- **Freedom Level**: Medium (semi-autonomous decisions)

## Responsibilities

1. Analyze model training results
2. Detect overfitting/underfitting
3. Recommend optimization strategies
4. Validate production readiness

## Tools

### Primary Tools
- **Read**: Load metrics files (CSV, JSON)
- **Bash**: Run evaluation scripts
- **Task**: Delegate to backend-expert for infrastructure

### Optional Tools
- **WebFetch**: Lookup latest research papers
- **Grep**: Search logs for anomalies

## Decision Framework

### Phase 1: Data Analysis

```python
metrics = load_metrics("model_results.json")

# Key metrics
train_accuracy = metrics['train_accuracy']
val_accuracy = metrics['val_accuracy']
test_accuracy = metrics['test_accuracy']
training_time = metrics['training_time']
```

### Phase 2: Overfitting Detection

```python
if train_accuracy - val_accuracy > 0.1:
    verdict = "OVERFITTING"
    recommendations = [
        "Add dropout layers",
        "Increase training data",
        "Apply L2 regularization"
    ]
elif val_accuracy - train_accuracy > 0.05:
    verdict = "UNDERFITTING"
    recommendations = [
        "Increase model capacity",
        "Reduce regularization",
        "Train for more epochs"
    ]
else:
    verdict = "BALANCED"
```

### Phase 3: Production Readiness

```python
checklist = {
    "accuracy": test_accuracy >= 0.85,
    "generalization": abs(train_accuracy - test_accuracy) < 0.05,
    "inference_time": inference_time < 100,  # ms
    "model_size": model_size < 100  # MB
}

ready = all(checklist.values())
```

## Activation Triggers

- User mentions: "ml", "model", "training", "validation"
- SPEC contains: machine learning keywords
- Explicit call: `Task(subagent_type="ml-validator")`

## Integration with Other Agents

```python
# Collaborate with backend-expert for deployment
if ready_for_production:
    Task(
        description="Deploy ML model to production",
        subagent_type="backend-expert",
        context={
            "model_path": "models/final_model.pkl",
            "metrics": metrics
        }
    )
```

## Example Session

User: "Validate my ML model results"

Agent:
1. Read "model_results.json"
2. Analyze metrics
3. Detect overfitting (train: 0.95, val: 0.82)
4. Generate report:

```
ML Model Validation Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Model: ResNet50
Dataset: ImageNet-subset

Metrics:
  Training Accuracy:    95.2%
  Validation Accuracy:  82.1%
  Test Accuracy:        81.8%
  Training Time:        2h 15m

Verdict: OVERFITTING DETECTED

Gap Analysis:
  Train-Val Gap: 13.1% (⚠ HIGH)
  Val-Test Gap:  0.3% (✓ GOOD)

Recommendations:
  1. Add dropout (rate=0.3) after dense layers
  2. Apply L2 regularization (lambda=0.001)
  3. Increase training data by 30%
  4. Implement early stopping (patience=5)

Production Readiness: ❌ NOT READY
  ✓ Accuracy meets target (>80%)
  ✗ Overfitting too high
  ✓ Inference time acceptable
  ✓ Model size acceptable

Next Steps:
  1. Apply recommendations
  2. Retrain model
  3. Re-validate
```

## Skill Dependencies

```python
# Load relevant skills
Skill("moai-domain-ml")
Skill("moai-lang-python")
Skill("moai-essentials-optimization")
```
```

______________________________________________________________________

## Creating Custom Commands

### Command Structure

Commands orchestrate multiple agents to complete workflows.

### Example: /train-model Command

**File: `.claude/commands/train-model.md`**

```markdown
# /train-model

Train and validate a machine learning model following MoAI-ADK TDD principles.

## Syntax

```
/train-model <spec-id> [options]
```

## Parameters

- `<spec-id>`: SPEC document ID (e.g., SPEC-001)
- `--dataset <path>`: Path to dataset (required)
- `--model <type>`: Model architecture (required)
- `--epochs <n>`: Number of training epochs (default: 10)
- `--lr <rate>`: Learning rate (default: 0.001)
- `--validate`: Run validation after training (default: true)

## Workflow

### Phase 1: Preparation

1. Load SPEC document
2. Verify dataset exists
3. Check dependencies installed
4. Initialize environment

### Phase 2: TDD Cycle

**RED Phase**:
```python
# Create test first
def test_model_accuracy():
    model = train_model(dataset, epochs=10)
    accuracy = evaluate(model, test_data)
    assert accuracy >= 0.85  # From SPEC requirement
```

**GREEN Phase**:
```python
# Minimal implementation
def train_model(dataset, epochs):
    model = create_model()
    model.fit(dataset, epochs=epochs)
    return model
```

**REFACTOR Phase**:
```python
# Optimize implementation
def train_model(dataset, epochs, callbacks=None):
    model = create_optimized_model()
    model.fit(
        dataset,
        epochs=epochs,
        callbacks=callbacks or get_default_callbacks(),
        validation_split=0.2
    )
    return model
```

### Phase 3: Validation

1. Run ml-validator agent
2. Check overfitting
3. Evaluate production readiness
4. Generate report

### Phase 4: Documentation

1. Update model card
2. Record metrics
3. Commit results
4. Create PR (if team mode)

## Example Usage

### Basic Training

```
/train-model SPEC-042 --dataset data/train.csv --model resnet50
```

Expected output:
```
Training Model (SPEC-042)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Phase 1: Preparation ✓
  Dataset: data/train.csv (10,000 samples)
  Model: ResNet50
  
Phase 2: TDD Cycle
  RED: Tests created ✓
  GREEN: Training complete ✓
    - Training accuracy: 92.5%
    - Validation accuracy: 88.2%
  REFACTOR: Optimizations applied ✓

Phase 3: Validation
  ml-validator report generated
  Overfitting: None detected ✓
  Production ready: YES ✓

Phase 4: Documentation
  Model card updated ✓
  Metrics saved ✓
  Git commit created ✓

Training complete: 15 min 32 sec
```

### Advanced Training

```
/train-model SPEC-042 \
  --dataset data/train.csv \
  --model bert-base \
  --epochs 20 \
  --lr 0.0001 \
  --validate
```

## Integration

This command activates:
- **ml-validator** agent (validation)
- **backend-expert** agent (deployment prep)
- **git-manager** agent (version control)

## Error Handling

Common errors:
- Dataset not found → Prompt user for correct path
- Insufficient GPU memory → Suggest smaller batch size
- Model doesn't converge → Recommend hyperparameter tuning
```

______________________________________________________________________

## Creating Custom Hooks

### Hook Types

1. **SessionStart**: Initialize session context
2. **PreToolUse**: Validate before tool execution
3. **PostToolUse**: Clean up after tool execution
4. **Notification**: Alert on important events

### Example: Security Pre-Commit Hook

**File: `.claude/hooks/security-precommit.sh`**

```bash
#!/bin/bash
# Security checks before git commit

set -e

echo "Running security pre-commit checks..."

# 1. Secret scanning
echo "1/4 Scanning for secrets..."
if command -v detect-secrets &> /dev/null; then
    detect-secrets scan --baseline .secrets.baseline
    if [ $? -ne 0 ]; then
        echo "❌ Secret detected! Commit blocked."
        exit 1
    fi
    echo "✓ No secrets found"
else
    echo "⚠ detect-secrets not installed (skipping)"
fi

# 2. Dependency vulnerability scan
echo "2/4 Checking dependencies..."
if command -v pip-audit &> /dev/null; then
    pip-audit --desc --requirement requirements.txt
    if [ $? -ne 0 ]; then
        echo "⚠ Vulnerabilities found (warning only)"
    else
        echo "✓ Dependencies secure"
    fi
else
    echo "⚠ pip-audit not installed (skipping)"
fi

# 3. SAST (Static Application Security Testing)
echo "3/4 Running SAST..."
if command -v bandit &> /dev/null; then
    bandit -r src/ -ll  # Low severity and above
    if [ $? -ne 0 ]; then
        echo "❌ Security issues found! Commit blocked."
        exit 1
    fi
    echo "✓ SAST passed"
else
    echo "⚠ bandit not installed (skipping)"
fi

# 4. Check for .env files
echo "4/4 Checking for .env files..."
if git diff --cached --name-only | grep -E '\.env$|\.env\.'; then
    echo "❌ .env file detected in commit! Blocked."
    echo "Hint: Add to .gitignore"
    exit 1
fi
echo "✓ No .env files"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All security checks passed"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
```

**File: `.claude/hooks/auto-format.py`**

```python
#!/usr/bin/env python3
"""
Auto-format Python files after creation
Hook Type: PostToolUse
Trigger: Write tool creates .py file
"""

import subprocess
import sys
from pathlib import Path

def format_file(filepath):
    """Format Python file with black and ruff"""
    
    # Black formatting
    result = subprocess.run(
        ['black', filepath],
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0:
        print(f"⚠ Black formatting failed: {result.stderr}")
        return False
    
    # Ruff linting
    result = subprocess.run(
        ['ruff', 'check', '--fix', filepath],
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0:
        print(f"⚠ Ruff linting found issues: {result.stderr}")
        # Don't fail, just warn
    
    print(f"✓ Formatted: {filepath}")
    return True

def main():
    # Get file path from hook arguments
    if len(sys.argv) < 2:
        print("Usage: auto-format.py <filepath>")
        sys.exit(1)
    
    filepath = Path(sys.argv[1])
    
    # Only format Python files
    if filepath.suffix != '.py':
        sys.exit(0)
    
    # Skip test files (optional)
    if 'test_' in filepath.name:
        print(f"⊘ Skipping test file: {filepath}")
        sys.exit(0)
    
    # Format
    if format_file(filepath):
        print("✅ Auto-formatting complete")
    else:
        print("⚠ Auto-formatting had issues")
    
    sys.exit(0)

if __name__ == '__main__':
    main()
```

### Registering Hooks

```json
// .moai/config.json
{
  "hooks": {
    "pre_commit": ".claude/hooks/security-precommit.sh",
    "post_write_py": ".claude/hooks/auto-format.py"
  }
}
```

______________________________________________________________________

## Integration Patterns

### Pattern 1: Command → Agent → Skill

```
User: /train-model SPEC-042 --dataset data.csv
        ↓
Command (train-model.md)
        ↓ Task(subagent_type="ml-validator")
ml-validator Agent
        ↓ Skill("moai-domain-ml")
Skill (ML domain knowledge)
```

### Pattern 2: Hook → Validation → Action

```
Git Commit Triggered
        ↓
PreCommit Hook (security-precommit.sh)
        ↓ Run checks
Checks Pass/Fail
        ↓
Allow/Block Commit
```

### Pattern 3: Multi-Agent Collaboration

```
User: "Deploy ML model to production"
        ↓
Alfred (Orchestrator)
        ├─→ ml-validator (validate model)
        ├─→ backend-expert (deploy infrastructure)
        └─→ security-expert (security review)
        ↓
All agents collaborate
        ↓
Deployment Complete
```

______________________________________________________________________

## Real-World Examples

### Example 1: CI/CD Automation Extension

**Goal**: Automate deployment pipeline

**Components**:
1. Custom command: `/deploy`
2. Custom agent: `deployment-expert`
3. Custom hook: `post-deployment-verify`

**Implementation**:

```markdown
# .claude/commands/deploy.md

/deploy <environment> [--skip-tests]

Workflow:
1. Run all tests (unless --skip-tests)
2. Build Docker image
3. Push to registry
4. Deploy to k8s cluster
5. Run smoke tests
6. Send notification
```

### Example 2: Data Pipeline Extension

**Goal**: Automate ETL workflows

**Components**:
1. Custom skill: `moai-custom-etl`
2. Custom agent: `data-engineer`
3. Custom command: `/run-pipeline`

**Features**:
- Data validation
- Transformation logic
- Error handling
- Monitoring integration

______________________________________________________________________

## Best Practices

### Skill Development

```
✅ DO:
- Keep skills focused (<500 words core content)
- Provide clear examples
- Document dependencies
- Version control
- Test with real use cases

❌ DON'T:
- Create monolithic skills
- Duplicate existing knowledge
- Hardcode paths/values
- Skip documentation
- Ignore Progressive Disclosure principle
```

### Agent Development

```
✅ DO:
- Define clear responsibilities
- Specify activation triggers
- Document decision logic
- Enable collaboration
- Handle errors gracefully

❌ DON'T:
- Overlap with existing agents
- Create circular dependencies
- Ignore tool permissions
- Skip error handling
- Make assumptions without validation
```

### Command Development

```
✅ DO:
- Use descriptive names
- Validate parameters
- Provide helpful error messages
- Support dry-run mode
- Document examples

❌ DON'T:
- Skip parameter validation
- Hardcode configuration
- Ignore user feedback
- Create destructive commands without confirmation
- Bypass safety checks
```

### Hook Development

```
✅ DO:
- Keep execution fast (<100ms)
- Fail gracefully
- Log actions
- Make idempotent
- Test edge cases

❌ DON'T:
- Block for long operations
- Modify files unexpectedly
- Ignore errors silently
- Create race conditions
- Skip testing
```

______________________________________________________________________

## Testing Extensions

### Testing Custom Skills

```bash
# Manual test
Skill("moai-custom-doc-validator")

# Verify skill loads correctly
# Verify examples work
# Verify integration with agents
```

### Testing Custom Agents

```python
# Unit test agent logic
def test_ml_validator():
    metrics = {
        'train_accuracy': 0.95,
        'val_accuracy': 0.82
    }
    verdict = ml_validator.analyze(metrics)
    assert verdict == "OVERFITTING"
```

### Testing Custom Commands

```bash
# Dry run
/train-model SPEC-042 --dataset data.csv --dry-run

# Verify workflow steps
# Check error handling
# Validate output format
```

### Testing Custom Hooks

```bash
# Test hook directly
bash .claude/hooks/security-precommit.sh

# Verify all checks run
# Test failure scenarios
# Check performance (<100ms)
```

______________________________________________________________________

## Troubleshooting Extensions

### Common Issues

**Skill not loading**:
```bash
# Check skill.yaml syntax
cat .claude/skills/my-skill/skill.yaml | jq .

# Verify directory structure
ls -la .claude/skills/my-skill/
```

**Agent not activating**:
```bash
# Check activation triggers
grep -r "activation" .claude/agents/

# Test explicit activation
Task(subagent_type="my-agent")
```

**Command not recognized**:
```bash
# Verify command file
ls .claude/commands/my-command.md

# Restart Claude Code session
```

**Hook not triggering**:
```bash
# Check hook registration
cat .moai/config.json | jq '.hooks'

# Test hook manually
bash .claude/hooks/my-hook.sh
```

______________________________________________________________________

**Next Steps**:
- [Architecture Guide](architecture.md) - Understand system design
- [Security Guide](security.md) - Secure your extensions
- [Performance Guide](performance.md) - Optimize extensions
