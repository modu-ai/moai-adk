# Troubleshooting Guide

Comprehensive solutions for issues encountered while using MoAI-ADK.

## Quick Navigation

- [Common Error Messages](#common-error-messages) - 20+ error solutions
- [Installation & Setup Issues](#installation-and-setup-issues)
- [Alfred Command Issues](#alfred-command-issues)
- [Development & Testing Issues](#development-and-testing-issues)
- [Git & Deployment Issues](#git-and-deployment-issues)
- [Diagnostic Checklists](#diagnostic-checklists)
- [FAQ](#frequently-asked-questions)

---

## Common Error Messages

### 1. ModuleNotFoundError: No module named 'moai_adk'

**Error Message**:

```
ModuleNotFoundError: No module named 'moai_adk'
```

**Meaning**: Python cannot find the MoAI-ADK package

**Root Causes**:

1. Package not installed in current environment
2. Wrong Python environment activated
3. Installation failed silently
4. Virtual environment not activated

**Solutions**:

```bash
# Solution 1: Install/reinstall package
uv pip install -e ".[dev]"

# Solution 2: Check Python environment
which python3
python3 --version  # Should be 3.13+

# Solution 3: Activate virtual environment
source .venv/bin/activate  # Linux/macOS
.venv\Scripts\activate  # Windows

# Solution 4: Verify installation
python3 -c "import moai_adk; print(moai_adk.__version__)"
```

**Verification**:

```bash
moai-adk --version
# Expected output: moai-adk version 0.21.2
```

**Prevention**:

- Always activate virtual environment before work
- Add activation to shell profile (.bashrc/.zshrc)
- Use `uv` for consistent dependency management

---

### 2. Command not found: /alfred:\*

**Error Message**:

```
Unknown command: /alfred:1-plan
```

**Meaning**: Claude Code doesn't recognize Alfred commands

**Root Causes**:

1. .claude/commands/ directory not loaded
2. Claude Code session not refreshed
3. CLAUDE.md not loaded properly
4. Working directory incorrect

**Solutions**:

```bash
# Solution 1: Refresh Claude Code session
# Exit and restart Claude Code
exit
claude

# Solution 2: Verify command files exist
ls -la .claude/commands/
# Should show: 0-project.md, 1-plan.md, 2-run.md, 3-sync.md

# Solution 3: Check CLAUDE.md
cat CLAUDE.md | grep -i "alfred"

# Solution 4: Reinitialize project
moai-adk init --force
```

**Verification**:

```bash
# Test command recognition
/alfred:0-project
# Should show: Project initialization prompt
```

**Prevention**:

- Keep .claude/ directory in version control
- Never modify command files directly
- Always work from project root

---

### 3. uv: command not found

**Error Message**:

```
bash: uv: command not found
```

**Meaning**: UV package manager not installed or not in PATH

**Root Causes**:

1. UV not installed
2. Installation path not in PATH
3. Shell config not reloaded
4. Installation incomplete

**Solutions**:

```bash
# Solution 1: Install UV
curl -LsSf https://astral.sh/uv/install.sh | sh

# Solution 2: Add to PATH (if needed)
export PATH="$HOME/.cargo/bin:$PATH"

# Solution 3: Reload shell config
source ~/.bashrc  # or ~/.zshrc

# Solution 4: Verify installation
uv --version
```

**Verification**:

```bash
which uv
# Expected: /home/user/.cargo/bin/uv (or similar)

uv pip list
# Should show installed packages
```

**Prevention**:

- Add UV path to shell profile
- Verify installation after system updates

---

### 4. Python version mismatch

**Error Message**:

```
Python 3.9.x is installed, but 3.13+ required
```

**Meaning**: Incompatible Python version

**Root Causes**:

1. Old Python version installed
2. Multiple Python versions on system
3. Virtual environment using wrong version

**Solutions**:

```bash
# Solution 1: Check Python version
python3 --version

# Solution 2: Install Python 3.13+
# macOS (using Homebrew)
brew install python@3.13

# Ubuntu/Debian
sudo add-apt-repository ppa:deadsnakes/ppa
sudo apt update
sudo apt install python3.13

# Windows (download from python.org)

# Solution 3: Create venv with correct version
python3.13 -m venv .venv
source .venv/bin/activate

# Solution 4: Verify
python --version
# Expected: Python 3.13.x
```

**Prevention**:

- Always use latest Python version
- Specify Python version in pyproject.toml

---

### 5. Permission denied: .moai/

**Error Message**:

```
PermissionError: [Errno 13] Permission denied: '.moai/config.json'
```

**Meaning**: Insufficient permissions to access .moai/ directory

**Root Causes**:

1. File ownership issues
2. Directory permissions too restrictive
3. Read-only file system
4. SELinux/AppArmor restrictions

**Solutions**:

```bash
# Solution 1: Fix ownership
sudo chown -R $USER:$USER .moai/

# Solution 2: Fix permissions
chmod -R 755 .moai/

# Solution 3: Check file system
mount | grep "$(pwd)"

# Solution 4: Verify
ls -la .moai/
# Should show: drwxr-xr-x user user
```

**Prevention**:

- Don't use sudo when running moai-adk commands
- Check file permissions before committing

---

### 6. YAML parsing error in config.json

**Error Message**:

```
json.decoder.JSONDecodeError: Expecting ',' delimiter: line 15 column 5
```

**Meaning**: Invalid JSON syntax in config.json

**Root Causes**:

1. Missing comma or bracket
2. Trailing comma (invalid in JSON)
3. Incorrect quotes (single vs double)
4. Comments (not allowed in JSON)

**Solutions**:

```bash
# Solution 1: Validate JSON syntax
cat .moai/config.json | jq .
# If error, shows exact location

# Solution 2: Common fixes
# Remove trailing commas:
{
  "key": "value",  # ❌ Trailing comma
}

# Should be:
{
  "key": "value"   # ✅ No trailing comma
}

# Solution 3: Use double quotes
{
  'key': 'value'  # ❌ Single quotes
}

# Should be:
{
  "key": "value"  # ✅ Double quotes
}

# Solution 4: Remove comments
{
  "key": "value"  // comment  # ❌ Comments not allowed
}

# Should be:
{
  "key": "value"
}
```

**Verification**:

```bash
jq . .moai/config.json
# Should output formatted JSON without errors
```

**Prevention**:

- Always validate JSON after editing
- Use JSON-aware editor (VS Code, etc.)
- Keep backup before editing

---

### 7. SPEC document creation failure

**Error Message**:

```
Error: Failed to create SPEC document at .moai/specs/SPEC-001.md
```

**Meaning**: SPEC file creation failed

**Root Causes**:

1. .moai/specs/ directory doesn't exist
2. File already exists
3. Invalid SPEC ID format
4. Disk space full

**Solutions**:

```bash
# Solution 1: Create specs directory
mkdir -p .moai/specs/

# Solution 2: Check existing files
ls -la .moai/specs/

# Solution 3: Remove duplicate
rm .moai/specs/SPEC-001.md  # If duplicate

# Solution 4: Check disk space
df -h .

# Solution 5: Retry with fresh ID
/alfred:1-plan "feature name"
```

**Prevention**:

- Don't manually edit SPEC files during creation
- Always use /alfred:1-plan for SPEC creation

---

### 8. Test execution failures

**Error Message**:

```
pytest: error: no tests found
```

**Meaning**: pytest cannot find test files

**Root Causes**:

1. Test files not in tests/ directory
2. Test files don't follow naming convention
3. Test functions don't start with test\_
4. pytest configuration issue

**Solutions**:

```bash
# Solution 1: Verify test file location
ls tests/
# Should contain: test_*.py files

# Solution 2: Check test naming
# File: tests/test_feature.py
# Function: def test_feature_works():

# Solution 3: Run with verbose mode
pytest -v

# Solution 4: Check pytest discovery
pytest --collect-only
```

**Prevention**:

- Follow pytest naming conventions
- Use test\_ prefix for all test files and functions

---

### 9. Code formatting errors (black/ruff)

**Error Message**:

```
error: cannot format file.py: Cannot parse: 1:15
```

**Meaning**: Code has syntax errors preventing formatting

**Root Causes**:

1. Python syntax error
2. Unclosed brackets/quotes
3. Invalid indentation
4. Incompatible Python version

**Solutions**:

```bash
# Solution 1: Check syntax
python3 -m py_compile src/file.py

# Solution 2: Find syntax errors
ruff check src/file.py

# Solution 3: Fix common issues
# Unclosed brackets: Check all (, [, {
# Unclosed quotes: Check all ", '
# Indentation: Use 4 spaces consistently

# Solution 4: Format step by step
black --check src/  # Check only
black src/  # Apply formatting
```

**Prevention**:

- Use IDE with syntax checking (VS Code, PyCharm)
- Run formatter before committing

---

### 10. Git merge conflicts

**Error Message**:

```
CONFLICT (content): Merge conflict in src/main.py
```

**Meaning**: Changes conflict between branches

**Root Causes**:

1. Same lines modified in both branches
2. Branch diverged from main
3. Force push overwrote history

**Solutions**:

```bash
# Solution 1: Check conflict status
git status

# Solution 2: View conflicts
git diff

# Solution 3: Resolve manually
# Open conflicted file
# Look for:
Their changes

# Edit to keep desired changes

# Solution 4: Mark resolved
git add src/main.py
git commit -m "Resolve merge conflict"

# Solution 5: Use merge tool
git mergetool
```

**Prevention**:

- Pull frequently from main
- Keep PRs small and focused
- Communicate with team before large changes

---

## Installation and Setup Issues

### Environment Setup Checklist

```
Pre-installation Verification
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Operating System
  [ ] macOS 12+ / Ubuntu 20.04+ / Windows 10+
  [ ] 64-bit architecture

Python Environment
  [ ] Python 3.13+ installed
  [ ] pip updated to latest version
  [ ] Virtual environment support (venv or uv)

Development Tools
  [ ] Git 2.30+ installed
  [ ] UV package manager installed
  [ ] Claude Code installed
  [ ] Code editor (VS Code / PyCharm)

Network & Permissions
  [ ] Internet connection stable
  [ ] PyPI access not blocked
  [ ] User has write permissions
  [ ] No proxy/firewall issues
```

### Fresh Installation Steps

```bash
# Step 1: Verify Python
python3 --version
# Expected: Python 3.13.x

# Step 2: Install UV
curl -LsSf https://astral.sh/uv/install.sh | sh

# Step 3: Clone/Create project
git clone <repo> my-project
cd my-project

# Step 4: Create virtual environment
uv venv

# Step 5: Activate venv
source .venv/bin/activate  # Linux/macOS
.venv\Scripts\activate  # Windows

# Step 6: Install dependencies
uv pip install -e ".[dev]"

# Step 7: Verify installation
moai-adk --version
moai-adk doctor

# Step 8: Initialize project
moai-adk init
```

---

## Alfred Command Issues

### Command Execution Problems

#### /alfred:0-project not working

**Symptoms**:

- Command not recognized
- No response from Alfred
- Initialization incomplete

**Diagnosis**:

```bash
# Check if command file exists
ls .claude/commands/0-project.md

# Verify CLAUDE.md loaded
cat CLAUDE.md | head -20

# Check Claude Code session
# (restart if needed)
```

**Fix**:

```bash
# Reinitialize project
moai-adk init --force

# Verify .claude/ structure
tree .claude/
```

#### /alfred:1-plan fails to create SPEC

**Symptoms**:

- SPEC file not created
- Invalid SPEC format
- Missing sections

**Diagnosis**:

```bash
# Check .moai/specs/ directory
ls -la .moai/specs/

# Verify config.json
cat .moai/config.json | jq .
```

**Fix**:

```bash
# Create specs directory if missing
mkdir -p .moai/specs/

# Try with explicit feature name
/alfred:1-plan "User authentication"

# Check generated SPEC
cat .moai/specs/SPEC-001.md
```

#### /alfred:2-run doesn't execute TDD

**Symptoms**:

- Tests not created
- Implementation without tests
- RED phase skipped

**Diagnosis**:

```bash
# Check SPEC exists
ls .moai/specs/SPEC-*.md

# Verify test directory
ls tests/

# Check pytest configuration
cat pyproject.toml | grep -A 10 pytest
```

**Fix**:

```bash
# Ensure SPEC exists
cat .moai/specs/SPEC-001.md

# Create tests directory
mkdir -p tests/

# Run with explicit SPEC ID
/alfred:2-run SPEC-001

# Monitor TodoWrite progress
# (should show RED → GREEN → REFACTOR)
```

---

## Development and Testing Issues

### Test Coverage Below Target

**Problem**: Coverage < 85% (MoAI-ADK target)

**Diagnosis**:

```bash
# Generate coverage report
pytest --cov=src --cov-report=html tests/

# View detailed report
open htmlcov/index.html

# Find uncovered lines
pytest --cov=src --cov-report=term-missing tests/
```

**Solution**:

```bash
# Identify missing test cases
# Coverage report shows:
src/feature.py   75%   15, 23, 45

# Add tests for lines 15, 23, 45
# File: tests/test_feature.py

def test_edge_case_line_15():
    # Test code for line 15
    pass

def test_error_handling_line_23():
    # Test code for line 23
    pass

# Run coverage again
pytest --cov=src tests/
# Target: >= 85%
```

### Flaky Tests

**Problem**: Tests pass/fail inconsistently

**Common Causes**:

1. Time-dependent logic
2. Random data
3. External dependencies
4. Race conditions

**Solutions**:

```python
# Cause 1: Time-dependent
# ❌ Flaky
def test_timeout():
    start = time.time()
    result = slow_function()
    assert time.time() - start < 1.0  # Flaky!

# ✅ Fixed
def test_timeout():
    with pytest.raises(TimeoutError):
        slow_function(timeout=0.5)

# Cause 2: Random data
# ❌ Flaky
def test_random():
    data = generate_random_data()
    assert len(data) == 10  # May fail randomly

# ✅ Fixed
def test_random():
    random.seed(42)  # Fixed seed
    data = generate_random_data()
    assert len(data) == 10

# Cause 3: External dependencies
# ❌ Flaky
def test_api():
    response = requests.get("https://api.example.com")
    assert response.status_code == 200  # Network dependent

# ✅ Fixed (using mock)
def test_api(mocker):
    mock_response = mocker.Mock()
    mock_response.status_code = 200
    mocker.patch('requests.get', return_value=mock_response)

    response = requests.get("https://api.example.com")
    assert response.status_code == 200
```

### Dependency Conflicts

**Problem**: Package version conflicts

**Diagnosis**:

```bash
# Check dependency tree
uv pip tree

# Find conflicts
uv pip check

# View outdated packages
uv pip list --outdated
```

**Solution**:

```bash
# Update specific package
uv pip install --upgrade package-name

# Update all dependencies
uv pip install --upgrade -r requirements.txt

# Resolve conflicts manually
# Edit pyproject.toml:
[project.dependencies]
package-a = ">=1.0,<2.0"
package-b = ">=2.0,<3.0"

# Reinstall
uv pip install -e ".[dev]"
```

---

## Git and Deployment Issues

### Branch Strategy Violations

**Problem**: PR targeting wrong branch

**Diagnosis**:

```bash
# Check current branch
git branch

# Check remote branches
git branch -r

# Check PR target
gh pr view 123 --json baseRefName
```

**Solution**:

```bash
# Fix PR target (GitFlow)
# PRs should target 'develop', not 'main'

# Change PR base branch
gh pr edit 123 --base develop

# Verify
gh pr view 123
```

### Failed CI/CD Pipeline

**Problem**: GitHub Actions failing

**Common Failures**:

1. **Test failures**:

```bash
# Run tests locally first
pytest tests/

# Check coverage
pytest --cov=src tests/

# Fix failing tests
# Then commit
```

2. **Linting errors**:

```bash
# Run linters locally
black src/ tests/
ruff check src/ tests/

# Fix issues
# Then commit
```

3. **Type checking errors**:

```bash
# Run mypy
mypy src/

# Fix type hints
# Then commit
```

4. **Build errors**:

```bash
# Test build locally
python -m build

# Verify package
twine check dist/*
```

**Prevention**:

```bash
# Pre-commit hook
# .git/hooks/pre-commit

#!/bin/bash
pytest tests/ || exit 1
black --check src/ tests/ || exit 1
ruff check src/ tests/ || exit 1
mypy src/ || exit 1
```

---

## Diagnostic Checklists

### Installation Verification Checklist

```
Installation Validation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

System Check
  [ ] Python 3.13+ installed and accessible
      python3 --version
  [ ] UV installed and in PATH
      uv --version
  [ ] Git configured with user.name and user.email
      git config user.name
      git config user.email

MoAI-ADK Installation
  [ ] Package installed successfully
      moai-adk --version
  [ ] Doctor command passes all checks
      moai-adk doctor
  [ ] Virtual environment activated
      which python (should show .venv path)

Project Structure
  [ ] .moai/ directory exists
      ls .moai/
  [ ] .claude/ directory exists
      ls .claude/
  [ ] CLAUDE.md exists
      cat CLAUDE.md | head -5
  [ ] config.json valid
      jq . .moai/config.json

Claude Code Integration
  [ ] Claude Code installed
      claude --version
  [ ] Commands recognized
      /alfred:0-project (should respond)
  [ ] Session initialized
      (check for welcome message)
```

### SPEC Creation Checklist

```
SPEC Document Validation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Before Creation
  [ ] Feature name clear and descriptive
  [ ] .moai/specs/ directory exists
  [ ] No duplicate SPEC IDs

During Creation (/alfred:1-plan)
  [ ] SPEC ID assigned (SPEC-XXX)
  [ ] Title and overview written
  [ ] EARS statements complete
      [ ] WHEN clause (trigger)
      [ ] IF clause (condition)
      [ ] THEN clause (action)
      [ ] SHALL clause (requirement)
  [ ] Acceptance criteria defined
  [ ] TAG references added

After Creation
  [ ] SPEC file exists (.moai/specs/SPEC-XXX.md)
  [ ] File readable and valid markdown
  [ ] All sections complete
  [ ] Ready for /alfred:2-run
```

### TDD Cycle Checklist

```
RED-GREEN-REFACTOR Validation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

RED Phase
  [ ] Test file created (tests/test_*.py)
  [ ] Test function follows naming (test_*)
  [ ] Test imports work
  [ ] Test fails as expected
      pytest tests/test_feature.py::test_new_feature -v
      (should show FAILED)

GREEN Phase
  [ ] Minimal implementation added
  [ ] All tests pass
      pytest tests/ -v
      (should show PASSED)
  [ ] Coverage meets target (85%+)
      pytest --cov=src tests/

REFACTOR Phase
  [ ] Code improved for readability
  [ ] No duplicate code
  [ ] Comments added where needed
  [ ] Tests still pass
      pytest tests/ -v

Final Validation
  [ ] All tests pass
  [ ] Coverage >= 85%
  [ ] Linting passes
      black --check src/ tests/
      ruff check src/ tests/
  [ ] Type checking passes
      mypy src/
```

### Deployment Readiness Checklist

```
Release Validation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Code Quality
  [ ] All tests pass locally
      pytest tests/
  [ ] Coverage >= 85%
      pytest --cov=src tests/
  [ ] No linting errors
      black src/ tests/ && ruff check src/ tests/
  [ ] Type checking clean
      mypy src/

Documentation
  [ ] README.md updated
  [ ] CHANGELOG.md updated
  [ ] API docs current
  [ ] Migration guide (if needed)

Security
  [ ] No secrets in code
      detect-secrets scan
  [ ] Dependencies scanned
      pip-audit
  [ ] Security headers configured
  [ ] HTTPS enforced

Git Workflow
  [ ] Feature branch clean
  [ ] PR targeting correct branch (develop)
  [ ] CI/CD passing
  [ ] Code review approved
  [ ] Merge conflicts resolved

Deployment
  [ ] Staging deployment successful
  [ ] Integration tests pass
  [ ] Performance acceptable
  [ ] Rollback plan ready
```

---

## Frequently Asked Questions

### General Questions

**Q1: How do I get started with MoAI-ADK?**

A: Follow these steps:

1. Install Python 3.13+
2. Install UV: `curl -LsSf https://astral.sh/uv/install.sh | sh`
3. Create project: `moai-adk init my-project`
4. Follow [Quick Start Guide](getting-started/quick-start.md)

**Q2: What is SPEC-first development?**

A: SPEC-first means:

- Write specifications before code
- Use EARS syntax for clarity
- Link requirements to tests/code via @TAG
- See [SPEC Basics](guides/specs/basics.md)

**Q3: What does Alfred do?**

A: Alfred is a SuperAgent that:

- Orchestrates 19 AI experts
- Automates SPEC → TDD → Sync workflow
- Enforces quality (TRUST 5 principles)
- See [Alfred Guide](guides/alfred/index.md)

### TDD Questions

**Q4: What is RED-GREEN-REFACTOR?**

A: Three-phase TDD cycle:

- **RED**: Write failing test first
- **GREEN**: Minimal code to pass test
- **REFACTOR**: Improve code quality

See [TDD Guide](guides/tdd/index.md)

**Q5: How much test coverage should I have?**

A: MoAI-ADK recommends **85% or higher**. Check with:

```bash
pytest --cov=src --cov-report=term-missing tests/
```

**Q6: Can I skip writing tests?**

A: No. MoAI-ADK enforces TDD:

- Alfred blocks implementation without tests
- Quality gate requires 85%+ coverage
- Tests are linked via @TAG system

### TAG System Questions

**Q7: What is the @TAG system?**

A: Traceability system that links:

- SPEC (requirements) → @TEST (validation)
- @TEST → @CODE (implementation)
- @CODE → @DOC (documentation)

See [TAG Guide](guides/specs/tags.md)

**Q8: How do I add TAGs to my code?**

A:

```python
# @CODE:SPEC-001:AUTH:LOGIN
def login(username, password):
    """Login implementation

    @CODE:SPEC-001:AUTH:LOGIN:VALIDATE
    """
    pass
```

### Git Workflow Questions

**Q9: Should I create PRs to main or develop?**

A: Always target **develop** (GitFlow):

```bash
# Correct
gh pr create --base develop

# Wrong
gh pr create --base main
```

**Q10: How do I resolve merge conflicts?**

A:

```bash
# 1. View conflicts
git status

# 2. Edit conflicted files
# (remove <<<, ===, >>>)

# 3. Mark resolved
git add <file>
git commit

# 4. Push
git push
```

### Configuration Questions

**Q11: Where is configuration stored?**

A: `.moai/config.json`:

```json
{
  "project": {...},
  "language": {...},
  "git_strategy": {...},
  "alfred": {...}
}
```

**Q12: How do I change conversation language?**

A:

```json
{
  "language": {
    "conversation_language": "en" // or "ko", "ja", "zh"
  }
}
```

### Troubleshooting Questions

**Q13: Tests pass locally but fail in CI**

A: Common causes:

1. Python version mismatch (check CI uses 3.13+)
2. Platform differences (Windows vs Linux)
3. Missing environment variables
4. Time zone differences

Fix:

```bash
# Match CI environment locally
python3.13 -m venv .venv
source .venv/bin/activate
uv pip install -e ".[dev]"
pytest tests/
```

**Q14: How do I debug Alfred commands?**

A:

```bash
# Enable verbose mode
export MOAI_ADK_DEBUG=1

# Check Alfred response
/alfred:1-plan "feature name"

# View session log
cat .moai/session.log
```

**Q15: Package installation fails**

A: Try these steps:

```bash
# 1. Clear cache
uv cache clean

# 2. Update UV
curl -LsSf https://astral.sh/uv/install.sh | sh

# 3. Reinstall
rm -rf .venv
uv venv
source .venv/bin/activate
uv pip install -e ".[dev]"
```

---

## Performance Troubleshooting

### Slow Test Execution

**Problem**: Tests take too long

**Diagnosis**:

```bash
# Profile tests
pytest --durations=10 tests/

# Identify slow tests
pytest --durations=0 tests/ | grep -E "s (call|setup|teardown)"
```

**Solutions**:

1. Use pytest markers for slow tests:

```python
@pytest.mark.slow
def test_long_running():
    pass

# Run fast tests only
pytest -m "not slow"
```

2. Parallelize tests:

```bash
pytest -n auto tests/
```

3. Use fixtures efficiently:

```python
@pytest.fixture(scope="module")  # Not "function"
def expensive_fixture():
    # Created once per module
    pass
```

### High Memory Usage

**Problem**: Tests consume too much memory

**Diagnosis**:

```bash
# Monitor memory
pytest --memray tests/
```

**Solutions**:

1. Clean up after tests:

```python
def test_large_data():
    data = create_large_dataset()
    # Use data
    del data  # Explicit cleanup
```

2. Use generators:

```python
# Instead of loading all at once
def load_data():
    for item in data_source:
        yield item
```

---

## Additional Resources

### Community Support

- [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) - Ask questions
- [Issue Tracker](https://github.com/modu-ai/moai-adk/issues) - Report bugs
- [Discord Community](https://discord.gg/moai-adk) - Real-time help

### Documentation

- [Online Docs](https://adk.mo.ai.kr) - Latest documentation
- [API Reference](reference/index.md) - Complete API docs
- [Examples](https://github.com/modu-ai/moai-adk-examples) - Sample projects

### Tools

```bash
# System diagnosis
moai-adk doctor

# Generate support bundle
moai-adk debug-info

# Report issue
/alfred:9-feedback
```

---

**Still stuck?** [Open a GitHub Discussion](https://github.com/modu-ai/moai-adk/discussions) and include:

- Error message (full traceback)
- Steps to reproduce
- Environment info (`moai-adk doctor` output)
- Expected vs actual behavior
