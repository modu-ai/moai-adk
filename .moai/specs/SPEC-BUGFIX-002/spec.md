---
title: Smart Merge for .github Directory - Preserve User Workflows
id: SPEC-BUGFIX-002
type: bugfix
priority: high
status: completed
affects: template processor, initialization system
discussion: #119
---

# SPEC-BUGFIX-002: Smart Merge for .github Directory

## Problem Statement

**Issue**: `moai-adk init` unconditionally deletes the entire `.github` directory, destroying user's custom GitHub Actions workflows.

**Impact**: 
- Loss of CI/CD pipelines
- Loss of deployment workflows
- Loss of custom automation scripts

**Severity**: High (data loss risk)

## Root Cause

### Affected Code

**File**: `src/moai_adk/core/template/processor.py`
**Method**: `_copy_github()`
**Lines**: 429-431

```python
def _copy_github(self, silent: bool = False) -> None:
    """Copy .github/ directory with variable substitution."""
    src = self.template_root / ".github"
    dst = self.target_path / ".github"

    if not src.exists():
        if not silent:
            console.print("⚠️ .github/ template not found")
        return

    # 🔴 Problem: Unconditionally deletes existing .github directory
    if dst.exists():
        shutil.rmtree(dst)  # ← Deletes all files!

    self._copy_dir_with_substitution(src, dst)
```

### Design Flaw

The current implementation follows a "**force overwrite**" strategy, which is appropriate for MoAI-ADK managed files but **not for user content**.

## Solution

### Approach: Smart Merge Strategy

Apply the same "smart merge" logic used for `.claude/`, `CLAUDE.md`, and `settings.json`.

### Design Principles

1. **MoAI-ADK Managed Files**: Always overwrite to ensure updates
2. **User Files**: Preserve existing files (never delete)
3. **New Files**: Copy from templates if not present
4. **Explicit Namespace**: MoAI-ADK workflows go into `workflows/moai-adk-*.yml`

### Implementation Design

#### 1. Define Managed Files

**MoAI-ADK Managed Workflows** (always overwrite):
```
.github/
└── workflows/
    ├── moai-adk-ci.yml          # MoAI-ADK CI pipeline
    ├── moai-adk-tag-validation.yml  # TAG validation
    └── moai-adk-auto-update.yml     # Auto-update system
```

**User Workflows** (never touch):
```
.github/
├── workflows/
│   ├── deploy.yml               # User's deployment
│   ├── test.yml                 # User's tests
│   └── custom.yml               # User's automation
└── scripts/
    └── deploy.sh                # User's scripts
```

#### 2. Update _copy_github() Method

**File**: `src/moai_adk/core/template/processor.py`

```python
def _copy_github(self, silent: bool = False) -> None:
    """Copy .github/ with smart merge (preserve user workflows).

    Strategy:
    - MoAI-ADK managed files (moai-adk-*.yml) → always overwrite
    - User files → preserve existing
    - New files → copy from template

    @CODE:BUGFIX-GITHUB-MERGE-001
    @SPEC:SPEC-BUGFIX-002
    """
    src = self.template_root / ".github"
    dst = self.target_path / ".github"

    if not src.exists():
        if not silent:
            console.print("⚠️ .github/ template not found")
        return

    # Create .github directory if not exists
    dst.mkdir(parents=True, exist_ok=True)

    # MoAI-ADK managed file patterns (always overwrite)
    managed_patterns = [
        "workflows/moai-adk-*.yml",
        "workflows/moai-adk-*.yaml",
    ]

    all_warnings = []

    for item in src.rglob("*"):
        if not item.is_file():
            continue

        rel_path = item.relative_to(src)
        dst_item = dst / rel_path

        # Check if this is a managed file
        is_managed = any(
            dst_item.match(pattern) for pattern in managed_patterns
        )

        if is_managed:
            # Always overwrite managed files
            dst_item.parent.mkdir(parents=True, exist_ok=True)
            warnings = self._copy_file_with_substitution(item, dst_item)
            all_warnings.extend(warnings)
            if not silent:
                console.print(f"   ✅ Overwritten: .github/{rel_path}")
        elif dst_item.exists():
            # Skip existing user files
            if not silent:
                console.print(f"   ⏭️  Preserved: .github/{rel_path}")
            continue
        else:
            # Copy new files from template
            dst_item.parent.mkdir(parents=True, exist_ok=True)
            warnings = self._copy_file_with_substitution(item, dst_item)
            all_warnings.extend(warnings)
            if not silent:
                console.print(f"   ➕ Created: .github/{rel_path}")

    # Print warnings if any
    if all_warnings and not silent:
        console.print("[yellow]⚠️ Template warnings:[/yellow]")
        for warning in set(all_warnings):  # Deduplicate
            console.print(f"   {warning}")

    if not silent:
        console.print("   ✅ .github/ smart merge complete")
```

#### 3. Update Template Workflow Names

Rename MoAI-ADK workflows to use `moai-adk-` prefix:

**Before**:
```
.github/workflows/ci.yml
.github/workflows/tag-validation.yml
```

**After**:
```
.github/workflows/moai-adk-ci.yml
.github/workflows/moai-adk-tag-validation.yml
```

This ensures clear separation between MoAI-ADK and user workflows.

## Testing Strategy

### Unit Tests

**File**: `tests/unit/test_github_merge.py`

```python
import shutil
from pathlib import Path
import pytest
from moai_adk.core.template.processor import TemplateProcessor


def test_github_smart_merge_preserves_user_workflows(tmp_path):
    """Test that user workflows are preserved during merge."""
    # Setup: existing project with user workflows
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    github_dir = project_dir / ".github" / "workflows"
    github_dir.mkdir(parents=True)

    user_workflow = github_dir / "deploy.yml"
    user_workflow.write_text("name: Deploy\non: push")

    # Initialize template processor
    processor = TemplateProcessor(project_dir)

    # Copy templates (should preserve user workflow)
    processor._copy_github(silent=True)

    # Verify user workflow still exists
    assert user_workflow.exists()
    assert user_workflow.read_text() == "name: Deploy\non: push"


def test_github_smart_merge_updates_managed_files(tmp_path):
    """Test that MoAI-ADK managed files are overwritten."""
    # Setup: existing project with old MoAI-ADK workflow
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    github_dir = project_dir / ".github" / "workflows"
    github_dir.mkdir(parents=True)

    old_workflow = github_dir / "moai-adk-ci.yml"
    old_workflow.write_text("name: Old CI\nversion: 0.9.0")

    # Initialize template processor
    processor = TemplateProcessor(project_dir)

    # Copy templates (should update managed workflow)
    processor._copy_github(silent=True)

    # Verify managed workflow is updated
    assert old_workflow.exists()
    content = old_workflow.read_text()
    assert "version: 0.9.0" not in content  # Old version removed
    assert "MoAI-ADK" in content  # New template content


def test_github_smart_merge_creates_new_files(tmp_path):
    """Test that new files from template are created."""
    # Setup: existing project without .github
    project_dir = tmp_path / "project"
    project_dir.mkdir()

    # Initialize template processor
    processor = TemplateProcessor(project_dir)

    # Copy templates (should create new files)
    processor._copy_github(silent=True)

    # Verify new files are created
    github_dir = project_dir / ".github" / "workflows"
    assert github_dir.exists()
    assert (github_dir / "moai-adk-ci.yml").exists()
```

### Integration Tests

**Test Scenarios**:

1. **Fresh Project**: `.github` doesn't exist
   - Expected: Create all MoAI-ADK workflows

2. **Existing Project with User Workflows**: `.github` has custom workflows
   - Expected: Preserve user workflows, add MoAI-ADK workflows

3. **Reinit Project**: Old MoAI-ADK workflows + user workflows
   - Expected: Update MoAI-ADK workflows, preserve user workflows

4. **Mixed Workflows**: `deploy.yml` (user) + `moai-adk-ci.yml` (managed)
   - Expected: Update only `moai-adk-ci.yml`

### Manual Testing

```bash
# Test 1: Fresh project
cd /tmp/test-project
moai-adk init
ls -la .github/workflows/
# Expected: moai-adk-ci.yml, moai-adk-tag-validation.yml

# Test 2: Add user workflow
echo "name: Deploy" > .github/workflows/deploy.yml
moai-adk init --force
ls -la .github/workflows/
# Expected: deploy.yml (preserved), moai-adk-ci.yml (updated)

# Test 3: Verify user workflow content
cat .github/workflows/deploy.yml
# Expected: "name: Deploy" (unchanged)
```

## Acceptance Criteria

- [ ] User workflows are **never deleted**
- [ ] MoAI-ADK managed workflows are **always updated**
- [ ] New workflows from template are **created if not present**
- [ ] Backup still created before any changes
- [ ] Clear console output showing preserved/updated/created files
- [ ] Unit tests pass for all scenarios
- [ ] Integration tests pass
- [ ] Documentation updated with merge behavior

## Implementation Plan

### Phase 1: Design and Planning (Day 1)
- Define managed file patterns
- Review merge logic design
- Write test cases

### Phase 2: Update Template Workflows (Day 2)
- Rename workflows to use `moai-adk-` prefix
- Update workflow documentation
- Test workflow execution

### Phase 3: Implement Smart Merge (Day 3-4)
- Update `_copy_github()` method
- Add managed file detection
- Implement merge logic

### Phase 4: Testing (Day 5)
- Write unit tests
- Write integration tests
- Manual testing on all scenarios

### Phase 5: Documentation and Release (Day 6-7)
- Update README.md with merge behavior
- Update initialization guide
- Create migration guide for existing projects
- Release v0.11.0

## Migration Guide for Users

### For Existing Projects

When upgrading to v0.11.0:

```bash
# 1. Backup is automatic, but you can manually backup too
cp -r .github ~/github-backup

# 2. Reinitialize (force mode)
moai-adk init --force

# 3. Verify your workflows are preserved
ls -la .github/workflows/
cat .github/workflows/your-workflow.yml

# 4. Check MoAI-ADK workflows are updated
cat .github/workflows/moai-adk-ci.yml
```

### Rollback Procedure

If issues occur:

```bash
# Option 1: Restore from backup
cp -r .moai-backups/{timestamp}/.github .github

# Option 2: Git restore
git checkout HEAD~1 -- .github/
```

## Risks and Mitigations

### Risk 1: Pattern matching fails
**Impact**: User files accidentally overwritten
**Mitigation**: Comprehensive unit tests, explicit allowlist

### Risk 2: New workflow conflicts with user workflows
**Impact**: Workflow name collision
**Mitigation**: Use `moai-adk-` prefix for all managed workflows

### Risk 3: Backup failure
**Impact**: Data loss without recovery
**Mitigation**: Validate backup creation before proceeding

## Related Issues

- Discussion #119: .github deletion on moai-adk init
- Related to SPEC-INIT-003: Backup and merge system

## Success Metrics

- Zero reports of lost user workflows
- No workflow conflicts reported
- CI/CD passes after reinit
- User satisfaction with merge behavior

---

**Author**: debug-helper
**Created**: 2025-10-30
**Target Release**: v0.11.0
