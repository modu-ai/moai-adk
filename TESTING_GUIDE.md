# Testing Guidelines for MoAI-ADK Development

## ⚠️ IMPORTANT: Prevent Accidental File Modifications

When running tests, **always execute from an isolated directory** to prevent tests from modifying project files like `.claude/settings.json`.

### Recommended Test Execution

```bash
# ✅ CORRECT: Run from isolated directory
cd /tmp/moai-test && pytest /Users/goos/MoAI/MoAI-ADK

# ❌ WRONG: Run from project root (may modify settings.json)
pytest
```

### Why This Matters

Some tests use `Path.cwd()` to access the current working directory. When run from the project root, these tests can:
- Modify `.claude/settings.json` with test data
- Overwrite user configurations
- Cause git diff noise

### Verification

After running tests, check if project files were modified:

```bash
git status
```

If `.claude/settings.json` appears as modified, restore it from git:

```bash
git checkout .claude/settings.json
```

### Continuous Integration

CI/CD pipelines should always run tests from isolated directories:

```yaml
# Example GitHub Actions
- name: Run tests
  run: |
    cd /tmp/moai-test
    pytest $GITHUB_WORKSPACE
```
