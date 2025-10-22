---
name: moai-foundation-git
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: GitFlow automation and PR policy enforcement for MoAI-ADK workflows.
keywords: ['git', 'gitflow', 'pr', 'automation']
allowed-tools:
  - Read
  - Bash
---

# Foundation Git Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-foundation-git |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | Plan/Run/Sync phases when Git operations detected |
| **Tier** | Foundation |

---

## What It Does

GitFlow automation and PR policy enforcement for MoAI-ADK workflows. This Skill provides comprehensive guidance for Git version control, branching strategies, pull request workflows, and integration with the MoAI-ADK SPEC-first TDD development cycle.

**Key capabilities**:
- ✅ GitFlow workflow orchestration (feature → develop → main)
- ✅ Automated branch creation and PR management
- ✅ TDD commit patterns (RED → GREEN → REFACTOR)
- ✅ Draft PR creation and Ready transition
- ✅ TAG-driven commit messages and traceability
- ✅ Git 2.47.0 and GitHub CLI 2.63.0 latest features
- ✅ TRUST 5 principles integration
- ✅ Merge conflict prevention and resolution strategies

---

## When to Use

**Automatic triggers**:
- `/alfred:1-plan` — Create feature branch and Draft PR
- `/alfred:2-run` — TDD commit loop (RED → GREEN → REFACTOR)
- `/alfred:3-sync` — Convert Draft PR to Ready, quality gate enforcement
- Git operation keywords: branch, commit, merge, rebase, push, pull request, PR

**Manual invocation**:
- Design branching strategy for new projects
- Troubleshoot merge conflicts or Git state issues
- Review commit history quality and TAG traceability
- Automate Git workflows with GitHub CLI

---

## Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Git** | 2.47.0 | Core version control | ✅ Current |
| **GitHub CLI (gh)** | 2.63.0 | PR automation, issue tracking | ✅ Current |
| **Git LFS** | 3.5.1 | Large file storage | ✅ Current |

---

## Core Principles

### 1. Code-First Traceability

Every commit must reference @TAG markers to ensure bidirectional traceability between SPEC, TEST, CODE, and DOC.

**Commit Message Template**:
```
<type>(scope): <subject>

- Context of the change
- Additional notes (optional)

Refs: @TAG-ID (if applicable)
```

**Commit Types**:
- `feat`: New feature implementation
- `fix`: Bug fix
- `test`: Test addition or modification (RED stage)
- `refactor`: Code refactoring without behavior change
- `docs`: Documentation updates
- `chore`: Maintenance tasks (dependencies, build config)
- `perf`: Performance improvement
- `style`: Code formatting (no logic change)

### 2. TDD Commit Rhythm

MoAI-ADK enforces a three-stage commit pattern during `/alfred:2-run`:

| Stage | Commit Template | Example |
|-------|----------------|---------|
| **RED** | `test: add failing test for <feature>` | `test: add failing test for JWT token validation` |
| **GREEN** | `feat: implement <feature> to pass tests` | `feat: implement JWT token validation to pass tests` |
| **REFACTOR** | `refactor: clean up <component> without changing behavior` | `refactor: extract JWT validation logic to helper function` |

**Rules**:
- RED commits must show test failures in CI logs
- GREEN commits must pass all tests
- REFACTOR commits must maintain 100% test coverage
- Each stage gets its own commit (no squashing during TDD cycle)

### 3. GitFlow Branch Hierarchy

MoAI-ADK follows a simplified GitFlow model optimized for SPEC-first development:

```
main (production-ready)
  └── develop (integration branch)
        ├── feature/SPEC-001-jwt-auth
        ├── feature/SPEC-002-user-profile
        └── hotfix/AUTH-003-token-expiry
```

**Branch Types**:
- `main`: Production-ready code, tagged releases only
- `develop`: Integration branch, always buildable
- `feature/SPEC-XXX-description`: Feature branches tied to SPEC IDs
- `hotfix/TAG-XXX-description`: Critical fixes branching from main
- `release/vX.Y.Z`: Release preparation branches (optional for larger projects)

**Branch Naming Convention**:
- Format: `<type>/<SPEC-ID>-<short-description>`
- Examples:
  - `feature/SPEC-001-jwt-auth`
  - `hotfix/AUTH-003-token-expiry-fix`
  - `release/v1.0.0`

### 4. Draft → Ready PR Workflow

Alfred automates PR lifecycle management:

```
/alfred:1-plan (Create SPEC)
    ↓
  Create feature branch
    ↓
  Create Draft PR (title: "[DRAFT] SPEC-001: JWT Authentication")
    ↓
/alfred:2-run (TDD implementation)
    ↓
  RED → GREEN → REFACTOR commits
    ↓
/alfred:3-sync (Quality gate enforcement)
    ↓
  Run tests, check coverage ≥85%
    ↓
  Convert Draft PR to Ready for Review
    ↓
  Notify reviewers, await approval
    ↓
  Auto-merge (if enabled) or manual merge
```

**PR Title Conventions**:
- Draft: `[DRAFT] SPEC-XXX: Brief description`
- Ready: `SPEC-XXX: Brief description` (remove [DRAFT])
- Hotfix: `[HOTFIX] TAG-XXX: Brief description`

**PR Description Template**:
```markdown
## Summary
Brief overview of changes (1-3 sentences)

## SPEC Reference
- SPEC ID: SPEC-XXX
- SPEC File: `.moai/specs/SPEC-XXX/spec.md`

## Changes
- [ ] Tests added/updated (RED → GREEN)
- [ ] Implementation complete (passes all tests)
- [ ] Code refactored (clean, maintainable)
- [ ] Documentation updated
- [ ] TAG markers added to code

## Test Coverage
- Before: XX%
- After: YY%
- Delta: +ZZ%

## Checklist
- [ ] All tests passing
- [ ] Coverage ≥85%
- [ ] No linter warnings
- [ ] TAG markers validated
- [ ] Living Docs updated

Refs: @SPEC:XXX, @TEST:XXX, @CODE:XXX
```

---

## Git 2.47.0 Latest Features

### Incremental Multi-Pack Index (MIDX)

Git 2.47.0 includes experimental support for incremental multi-pack indexes, improving performance for large repositories with many pack files.

**Use case**: Repositories with 100+ pack files (common in monorepos or long-lived projects).

**Enable incremental MIDX**:
```bash
git config core.multiPackIndex true
git repack -Ad  # Repack with incremental MIDX
```

**Status check**:
```bash
git multi-pack-index verify
```

**Limitations**:
- Does not yet support multi-pack reachability bitmaps
- Still experimental (use with caution in production)

### Repository Initialization Configuration

New configuration option `init.defaultObjectFormat` allows setting the default object format (SHA-1 or SHA-256) for new repositories.

**Set default to SHA-256 (recommended for new projects)**:
```bash
git config --global init.defaultObjectFormat sha256
```

**Verify configuration**:
```bash
git config --get init.defaultObjectFormat
# Output: sha256
```

**Benefits**:
- SHA-256 provides better collision resistance
- Future-proofs repositories against cryptographic vulnerabilities
- Required for projects with strict security requirements

**Compatibility**:
- SHA-256 repositories require Git 2.29+ on all systems
- GitHub supports SHA-256 repositories (as of 2023)
- Mixing SHA-1 and SHA-256 in the same repository is not supported

### Background Maintenance with --detach

Git 2.47.0 introduces the `--detach` option for `git maintenance`, allowing the entire maintenance process to run in the background.

**Enable background maintenance**:
```bash
git maintenance start --detach
```

**Configure maintenance tasks**:
```bash
git config maintenance.auto true
git config maintenance.strategy incremental
```

**Verify maintenance is running**:
```bash
git maintenance run --task=gc --task=commit-graph
```

**Benefits**:
- Reduces user-facing latency during Git operations
- Automatically optimizes repository over time
- Prevents repository bloat from accumulating

### Reftable Backend Fixes

Git 2.47.0 includes critical fixes for the "reftable" backend, addressing bugs in table compaction.

**Enable reftable backend** (experimental):
```bash
git config core.refStorage reftable
```

**Benefits**:
- Faster reference lookups (10-100x for repositories with many refs)
- Reduced filesystem overhead (single file vs. thousands of ref files)
- Atomic reference updates (prevents corruption during crashes)

**Compatibility**:
- Requires Git 2.47.0+ on all systems accessing the repository
- Not yet supported by all Git hosting platforms
- Consider using only for internal repositories or advanced use cases

### git-refs Command for Server-Side Consistency

New `git-refs(1)` command provides low-level access to references, useful for ensuring repository consistency on servers.

**Check reference consistency**:
```bash
git refs verify
```

**List all references**:
```bash
git refs list
```

**Future integration**:
- Planned integration with `git-fsck(1)` for automated consistency checks
- Useful for CI/CD pipelines and pre-receive hooks

---

## GitHub CLI (gh) Automation

### PR Creation and Management

**Create Draft PR from command line**:
```bash
gh pr create \
  --draft \
  --title "[DRAFT] SPEC-001: JWT Authentication" \
  --body "$(cat <<'EOF'
## Summary
Implement JWT-based authentication system

## SPEC Reference
- SPEC ID: SPEC-001
- SPEC File: `.moai/specs/SPEC-001/spec.md`

## Changes
- [ ] Tests added (RED)
- [ ] Implementation (GREEN)
- [ ] Refactoring (REFACTOR)

Refs: @SPEC:AUTH-001
EOF
)"
```

**Convert Draft to Ready**:
```bash
gh pr ready <PR-number>
```

**Auto-merge after approval**:
```bash
gh pr merge <PR-number> --auto --squash
```

**Batch PR status check**:
```bash
gh pr list --state open --json number,title,isDraft,reviews
```

### Issue and Project Integration

**Link PR to issue**:
```bash
gh pr create --body "Closes #123"
```

**Add PR to project board**:
```bash
gh pr edit <PR-number> --add-project "MoAI Development"
```

**Request reviewers**:
```bash
gh pr edit <PR-number> --add-reviewer @username1,@username2
```

### CI/CD Integration

**Wait for checks to pass before merge**:
```bash
gh pr checks <PR-number> --watch
gh pr merge <PR-number> --auto --squash
```

**Trigger workflow on PR creation**:
```yaml
# .github/workflows/pr-validation.yml
name: PR Validation
on:
  pull_request:
    types: [opened, synchronize, ready_for_review]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Validate SPEC reference
        run: |
          PR_TITLE="${{ github.event.pull_request.title }}"
          if [[ ! "$PR_TITLE" =~ SPEC-[0-9]+ ]]; then
            echo "ERROR: PR title must reference a SPEC ID"
            exit 1
          fi
```

---

## GitFlow Workflow Best Practices

### Feature Branch Workflow

**1. Start new feature**:
```bash
# Create feature branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/SPEC-001-jwt-auth
```

**2. TDD commit cycle** (during `/alfred:2-run`):
```bash
# RED: Write failing test
git add tests/auth/jwt.test.ts
git commit -m "test: add failing test for JWT token validation

- Test validates JWT signature
- Test checks expiration time
- Test verifies token claims

Refs: @TEST:AUTH-001"

# GREEN: Implement feature
git add src/auth/jwt.ts
git commit -m "feat: implement JWT token validation to pass tests

- Parse JWT token
- Verify signature with public key
- Check expiration and claims

Refs: @CODE:AUTH-001, @TEST:AUTH-001"

# REFACTOR: Clean up code
git add src/auth/jwt.ts src/auth/helpers.ts
git commit -m "refactor: extract JWT validation logic to helper function

- Move signature verification to helpers.ts
- Add JSDoc comments
- Improve error messages

Refs: @CODE:AUTH-001"
```

**3. Push and create PR**:
```bash
git push -u origin feature/SPEC-001-jwt-auth
gh pr create --draft --title "[DRAFT] SPEC-001: JWT Authentication"
```

**4. Convert to Ready after `/alfred:3-sync`**:
```bash
gh pr ready <PR-number>
```

### Hotfix Workflow

**1. Create hotfix branch from main**:
```bash
git checkout main
git pull origin main
git checkout -b hotfix/AUTH-003-token-expiry-fix
```

**2. Fix issue and test**:
```bash
# Add test for bug
git add tests/auth/jwt-expiry.test.ts
git commit -m "test: add test for token expiry edge case

- Test token expiration at exact boundary
- Verify clock skew handling

Refs: @TEST:AUTH-003"

# Implement fix
git add src/auth/jwt.ts
git commit -m "fix: handle token expiry edge case correctly

- Add 5-second clock skew tolerance
- Improve expiry boundary checks

Refs: @CODE:AUTH-003, @TEST:AUTH-003"
```

**3. Merge to main and develop**:
```bash
# Create PR targeting main
gh pr create --base main --title "[HOTFIX] AUTH-003: Token expiry fix"

# After merge to main, cherry-pick to develop
git checkout develop
git cherry-pick <hotfix-commit-sha>
git push origin develop
```

### Release Branch Workflow (Optional)

For projects with scheduled releases:

**1. Create release branch**:
```bash
git checkout develop
git pull origin develop
git checkout -b release/v1.0.0
```

**2. Finalize release**:
```bash
# Update version numbers
git add pyproject.toml package.json
git commit -m "chore: bump version to v1.0.0"

# Final testing and bug fixes only (no new features)
```

**3. Merge to main and tag**:
```bash
git checkout main
git merge --no-ff release/v1.0.0
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin main --tags

# Merge back to develop
git checkout develop
git merge --no-ff release/v1.0.0
git push origin develop
```

---

## Merge Conflict Prevention

### Strategies to Minimize Conflicts

**1. Frequent integration**:
```bash
# Rebase feature branch on develop daily
git checkout feature/SPEC-001-jwt-auth
git fetch origin
git rebase origin/develop
```

**2. Small, focused branches**:
- Keep feature branches short-lived (< 3 days)
- Limit scope to single SPEC implementation
- Break large features into multiple SPECs

**3. Avoid overlapping file edits**:
- Coordinate with team on `.moai/specs/` directory
- Use separate modules for concurrent features
- Communicate when editing shared utilities

**4. Use rebase instead of merge for feature branches**:
```bash
# Instead of: git merge develop
git rebase develop
```

### Conflict Resolution Workflow

**When conflict occurs**:
```bash
# 1. Fetch latest changes
git fetch origin

# 2. Attempt rebase
git rebase origin/develop

# 3. If conflicts occur, resolve manually
# Git will pause and show conflicting files
git status

# 4. Edit conflicting files, remove conflict markers
# <<<<<<< HEAD
# your changes
# =======
# their changes
# >>>>>>> branch-name

# 5. Mark as resolved and continue
git add <conflicted-files>
git rebase --continue

# 6. If conflict resolution is too complex, abort and seek help
git rebase --abort
```

**Using merge tools**:
```bash
# Configure merge tool (e.g., VS Code)
git config --global merge.tool vscode
git config --global mergetool.vscode.cmd 'code --wait $MERGED'

# Launch merge tool during conflict
git mergetool
```

**Prevention checklist**:
- [ ] Rebase feature branches daily on develop
- [ ] Keep feature branches short-lived (< 72 hours)
- [ ] Communicate with team about file ownership
- [ ] Use `git pull --rebase` instead of `git pull`
- [ ] Run tests before pushing to catch issues early

---

## TAG-Driven Commit Messages

### Commit Message Anatomy

```
<type>(scope): <subject>          ← 50 chars max, imperative mood
                                  ← blank line
- Context of the change           ← 72 chars per line
- Additional notes (optional)
                                  ← blank line
Refs: @TAG-ID (if applicable)     ← TAG reference for traceability
```

**Examples**:

**Feature implementation**:
```
feat(auth): implement JWT token validation

- Add JWT signature verification using public key
- Validate expiration time with 5-second clock skew
- Extract and verify required claims (sub, exp, iat)

Refs: @CODE:AUTH-001, @TEST:AUTH-001
```

**Test addition (RED stage)**:
```
test(auth): add failing test for JWT token validation

- Test validates JWT signature against public key
- Test checks expiration time boundary conditions
- Test verifies required claims are present

Refs: @TEST:AUTH-001
```

**Refactoring (REFACTOR stage)**:
```
refactor(auth): extract JWT validation logic to helper

- Move signature verification to auth/helpers.ts
- Improve error messages for invalid tokens
- Add JSDoc comments for public functions

Refs: @CODE:AUTH-001
```

**Bug fix**:
```
fix(auth): handle token expiry edge case correctly

- Add 5-second clock skew tolerance to prevent false negatives
- Improve boundary check for exact expiration time
- Update tests to cover edge case

Refs: @CODE:AUTH-003, @TEST:AUTH-003
```

### TAG Reference Format

**Single TAG reference**:
```
Refs: @CODE:AUTH-001
```

**Multiple TAG references**:
```
Refs: @CODE:AUTH-001, @TEST:AUTH-001, @SPEC:AUTH-001
```

**Cross-domain references**:
```
Refs: @CODE:AUTH-001 (implementation), @CODE:USER-005 (integration)
```

### Commit Message Quality Checklist

- [ ] Type is one of: feat, fix, test, refactor, docs, chore, perf, style
- [ ] Scope identifies affected component (e.g., auth, user, api)
- [ ] Subject is imperative mood ("add feature" not "added feature")
- [ ] Subject is ≤50 characters
- [ ] Body provides context (why, not what)
- [ ] Body lines are ≤72 characters
- [ ] TAG references are present and correct
- [ ] Commit passes linting (if using commitlint)

---

## Git Configuration Best Practices

### Global Configuration (User Level)

**Essential settings**:
```bash
# Identity
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Default branch name
git config --global init.defaultBranch develop

# Object format (SHA-256 recommended for new projects)
git config --global init.defaultObjectFormat sha256

# Editor
git config --global core.editor "code --wait"

# Diff and merge tools
git config --global merge.tool vscode
git config --global diff.tool vscode

# Rebase by default when pulling
git config --global pull.rebase true

# Auto-correct typos (wait 2 seconds before executing)
git config --global help.autocorrect 20

# Color output
git config --global color.ui auto
```

**Advanced settings**:
```bash
# Enable reftable backend (experimental, Git 2.47.0+)
git config --global core.refStorage reftable

# Enable background maintenance
git config --global maintenance.auto true
git config --global maintenance.strategy incremental

# Enable multi-pack index
git config --global core.multiPackIndex true

# Prune remote branches on fetch
git config --global fetch.prune true

# GPG signing (if required)
git config --global commit.gpgsign true
git config --global user.signingkey <YOUR-GPG-KEY-ID>
```

### Repository Configuration (Project Level)

**MoAI-ADK specific settings**:
```bash
# Navigate to project root
cd /path/to/project

# Enforce commit message format
git config commit.template .moai/templates/commit-template.txt

# Hook for pre-commit validation
git config core.hooksPath .moai/hooks

# Ignore file permissions (useful on Windows/macOS)
git config core.fileMode false

# Large file storage (if needed)
git config lfs.url "https://github.com/<org>/<repo>.git/info/lfs"
```

**Example commit template** (`.moai/templates/commit-template.txt`):
```
<type>(scope): <subject>

- Context of the change
- Additional notes (optional)

Refs: @TAG-ID (if applicable)

# Type: feat, fix, test, refactor, docs, chore, perf, style
# Scope: component name (auth, user, api, etc.)
# Subject: imperative mood, ≤50 chars
# Body: 72 chars per line
# TAG: @CODE:XXX, @TEST:XXX, @SPEC:XXX, @DOC:XXX
```

---

## GitHub Actions Integration

### Pre-commit Validation Workflow

**`.github/workflows/pre-commit.yml`**:
```yaml
name: Pre-commit Validation

on:
  pull_request:
    types: [opened, synchronize, ready_for_review]

jobs:
  validate-commit-messages:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Fetch all history for commit analysis

      - name: Validate commit messages
        run: |
          # Check all commits in PR
          for commit in $(git rev-list origin/${{ github.base_ref }}..${{ github.sha }}); do
            message=$(git log --format=%B -n 1 $commit)

            # Check for TAG reference
            if ! echo "$message" | grep -qE 'Refs: @(CODE|TEST|SPEC|DOC):'; then
              echo "ERROR: Commit $commit missing TAG reference"
              exit 1
            fi

            # Check commit message format
            if ! echo "$message" | head -1 | grep -qE '^(feat|fix|test|refactor|docs|chore|perf|style)\(.+\): .+$'; then
              echo "ERROR: Commit $commit has invalid format"
              exit 1
            fi
          done

  validate-pr-title:
    runs-on: ubuntu-latest
    steps:
      - name: Validate PR title
        run: |
          PR_TITLE="${{ github.event.pull_request.title }}"

          # Check for SPEC reference
          if [[ ! "$PR_TITLE" =~ SPEC-[0-9]+ ]]; then
            echo "ERROR: PR title must reference a SPEC ID"
            exit 1
          fi
```

### Automated TAG Validation

**`.github/workflows/tag-validation.yml`**:
```yaml
name: TAG Validation

on:
  pull_request:
    types: [opened, synchronize, ready_for_review]

jobs:
  validate-tags:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Validate TAG references
        run: |
          # Extract SPEC ID from PR title
          SPEC_ID=$(echo "${{ github.event.pull_request.title }}" | grep -oE 'SPEC-[0-9]+')

          if [ -z "$SPEC_ID" ]; then
            echo "ERROR: No SPEC ID found in PR title"
            exit 1
          fi

          # Check for corresponding SPEC file
          if [ ! -f ".moai/specs/$SPEC_ID/spec.md" ]; then
            echo "ERROR: SPEC file not found: .moai/specs/$SPEC_ID/spec.md"
            exit 1
          fi

          # Check for TEST TAG references in code
          if ! grep -rn "@CODE:$SPEC_ID" src/; then
            echo "WARNING: No @CODE:$SPEC_ID references found in src/"
          fi

          if ! grep -rn "@TEST:$SPEC_ID" tests/; then
            echo "ERROR: No @TEST:$SPEC_ID references found in tests/"
            exit 1
          fi
```

### Test Coverage Enforcement

**`.github/workflows/coverage.yml`**:
```yaml
name: Test Coverage

on:
  pull_request:
    types: [opened, synchronize, ready_for_review]

jobs:
  check-coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run tests with coverage
        run: |
          # Run tests (example for Node.js with Jest)
          npm test -- --coverage --coverageThreshold='{"global":{"lines":85,"functions":85,"branches":85,"statements":85}}'

      - name: Coverage delta check
        run: |
          # Compare coverage before and after PR
          BASE_COVERAGE=$(git show origin/${{ github.base_ref }}:coverage/coverage-summary.json | jq '.total.lines.pct')
          HEAD_COVERAGE=$(jq '.total.lines.pct' coverage/coverage-summary.json)

          DELTA=$(echo "$HEAD_COVERAGE - $BASE_COVERAGE" | bc)

          if (( $(echo "$DELTA < 0" | bc -l) )); then
            echo "ERROR: Coverage decreased by $DELTA%"
            exit 1
          fi

          echo "✓ Coverage increased by $DELTA% (Base: $BASE_COVERAGE%, Head: $HEAD_COVERAGE%)"
```

---

## Common Git Scenarios

### Scenario 1: Undo Last Commit (Keep Changes)

```bash
# Undo commit but keep changes staged
git reset --soft HEAD~1

# Undo commit and unstage changes
git reset HEAD~1

# Undo commit and discard changes (DANGEROUS)
git reset --hard HEAD~1
```

### Scenario 2: Amend Last Commit

```bash
# Modify last commit message
git commit --amend -m "New commit message"

# Add forgotten file to last commit
git add forgotten-file.ts
git commit --amend --no-edit
```

### Scenario 3: Cherry-Pick Specific Commit

```bash
# Apply commit from another branch
git cherry-pick <commit-sha>

# Cherry-pick without committing (review changes first)
git cherry-pick -n <commit-sha>
```

### Scenario 4: Stash Changes Temporarily

```bash
# Stash all uncommitted changes
git stash save "WIP: working on feature X"

# List stashes
git stash list

# Apply most recent stash
git stash pop

# Apply specific stash
git stash apply stash@{1}

# Delete specific stash
git stash drop stash@{1}
```

### Scenario 5: Revert Merged PR

```bash
# Revert merge commit (creates new commit)
git revert -m 1 <merge-commit-sha>
git push origin develop
```

### Scenario 6: Interactive Rebase for Commit Cleanup

```bash
# Rebase last 3 commits interactively
git rebase -i HEAD~3

# In the editor:
# pick 1234567 feat: implement feature A
# squash 2345678 fix: typo in feature A    ← Squash this commit
# reword 3456789 test: add tests for feature A    ← Change message

# Save and exit editor to apply changes
```

### Scenario 7: Bisect to Find Bug Introduction

```bash
# Start bisect session
git bisect start
git bisect bad HEAD
git bisect good <known-good-commit>

# Git will check out a commit in the middle
# Test the commit and mark as good or bad
git bisect good  # or: git bisect bad

# Repeat until Git identifies the problematic commit
# Git will output: "<commit-sha> is the first bad commit"

# End bisect session
git bisect reset
```

---

## Failure Modes & Recovery

### Failure Mode 1: Detached HEAD State

**Symptoms**:
```bash
git status
# Output: HEAD detached at 1234567
```

**Recovery**:
```bash
# Create branch from current position
git checkout -b recovery-branch

# Or return to previous branch
git checkout -
```

### Failure Mode 2: Merge Conflict During Rebase

**Symptoms**:
```bash
git rebase origin/develop
# Output: CONFLICT (content): Merge conflict in src/auth/jwt.ts
```

**Recovery**:
```bash
# Option 1: Resolve conflict manually
git status
# Edit conflicting files
git add <resolved-files>
git rebase --continue

# Option 2: Abort rebase
git rebase --abort
```

### Failure Mode 3: Accidentally Committed to Wrong Branch

**Symptoms**:
```bash
# Realize you committed to develop instead of feature branch
git log --oneline -3
# Output shows commit on develop that should be on feature branch
```

**Recovery**:
```bash
# Move commit to correct branch
git checkout feature/SPEC-001-jwt-auth
git cherry-pick <commit-sha>

# Remove commit from wrong branch
git checkout develop
git reset --hard HEAD~1
```

### Failure Mode 4: Lost Commits After Reset

**Symptoms**:
```bash
# Ran: git reset --hard HEAD~3
# Now realize you need those commits back
```

**Recovery**:
```bash
# View reflog to find lost commits
git reflog

# Output shows:
# 1234567 HEAD@{0}: reset: moving to HEAD~3
# 2345678 HEAD@{1}: commit: feat: implement feature
# 3456789 HEAD@{2}: commit: test: add tests

# Recover lost commit
git cherry-pick 2345678 3456789
```

### Failure Mode 5: Corrupted Repository

**Symptoms**:
```bash
git status
# Output: error: object file .git/objects/... is empty
```

**Recovery**:
```bash
# Verify repository integrity
git fsck --full

# If recoverable, pull fresh copy of remote
git fetch origin
git reset --hard origin/develop

# If not recoverable, clone fresh repository
cd ..
mv corrupted-repo corrupted-repo.backup
git clone <repository-url>
cd <repository-name>
```

---

## Integration with MoAI-ADK Workflow

### Phase 0: Bootstrap (`/alfred:0-project`)

**Git initialization**:
```bash
# Initialize repository with SHA-256 (if new project)
git init --object-format=sha256

# Create .gitignore
cat > .gitignore <<'EOF'
# Dependencies
node_modules/
__pycache__/
.venv/

# Build artifacts
dist/
build/
*.egg-info/

# IDE
.vscode/
.idea/
*.swp

# MoAI-ADK temporary files
.moai/tmp/

# Environment
.env
.env.local
EOF

# Initial commit
git add .
git commit -m "chore: initialize project with MoAI-ADK structure

- Add .gitignore
- Set up .moai/ directory structure
- Configure project metadata

Refs: @DOC:INIT-001"

# Create develop branch
git checkout -b develop
git push -u origin develop
```

### Phase 1: Plan (`/alfred:1-plan`)

**Git operations**:
```bash
# Create feature branch
git checkout develop
git pull origin develop
git checkout -b feature/SPEC-001-jwt-auth

# Create Draft PR
gh pr create \
  --draft \
  --title "[DRAFT] SPEC-001: JWT Authentication" \
  --body "Implement JWT-based authentication system"

# Push SPEC file
git add .moai/specs/SPEC-001/spec.md
git commit -m "docs: add SPEC-001 for JWT authentication

- Define JWT token structure
- Specify validation requirements
- Outline error handling

Refs: @SPEC:AUTH-001"

git push -u origin feature/SPEC-001-jwt-auth
```

### Phase 2: Run (`/alfred:2-run`)

**TDD commit cycle**:
```bash
# RED: Add failing test
git add tests/auth/jwt.test.ts
git commit -m "test: add failing test for JWT token validation

- Test validates JWT signature
- Test checks expiration time
- Test verifies token claims

Refs: @TEST:AUTH-001"

# GREEN: Implement feature
git add src/auth/jwt.ts
git commit -m "feat: implement JWT token validation to pass tests

- Parse JWT token
- Verify signature with public key
- Check expiration and claims

Refs: @CODE:AUTH-001, @TEST:AUTH-001"

# REFACTOR: Clean up code
git add src/auth/jwt.ts src/auth/helpers.ts
git commit -m "refactor: extract JWT validation logic to helper function

- Move signature verification to helpers.ts
- Add JSDoc comments
- Improve error messages

Refs: @CODE:AUTH-001"

# Push all TDD commits
git push origin feature/SPEC-001-jwt-auth
```

### Phase 3: Sync (`/alfred:3-sync`)

**Quality gate enforcement**:
```bash
# Verify all tests pass
npm test

# Check test coverage ≥85%
npm test -- --coverage

# Validate TAG references
rg '@(CODE|TEST|SPEC):AUTH-001' -n

# Update Living Docs
git add docs/living/AUTH.md
git commit -m "docs: update Living Docs for JWT authentication

- Add JWT authentication section
- Link to SPEC-001 and implementation
- Include usage examples

Refs: @DOC:AUTH-001, @SPEC:AUTH-001"

# Convert Draft PR to Ready
gh pr ready <PR-number>

# Request reviewers
gh pr edit <PR-number> --add-reviewer @reviewer1,@reviewer2

# Push final changes
git push origin feature/SPEC-001-jwt-auth
```

### Phase 4: Merge and Cleanup

**After PR approval**:
```bash
# Merge PR (squash or merge commit based on project policy)
gh pr merge <PR-number> --squash

# Sync local develop branch
git checkout develop
git pull origin develop

# Delete feature branch (local and remote)
git branch -d feature/SPEC-001-jwt-auth
git push origin --delete feature/SPEC-001-jwt-auth

# Tag release (if applicable)
git tag -a v1.0.0 -m "Release v1.0.0: JWT authentication"
git push origin v1.0.0
```

---

## Best Practices Summary

### DO:
- ✅ Use feature branches tied to SPEC IDs
- ✅ Follow TDD commit rhythm (RED → GREEN → REFACTOR)
- ✅ Reference @TAG markers in all commits
- ✅ Create Draft PRs at the start of `/alfred:1-plan`
- ✅ Convert Draft PRs to Ready after `/alfred:3-sync`
- ✅ Rebase feature branches daily on develop
- ✅ Keep feature branches short-lived (< 3 days)
- ✅ Use GitHub CLI for PR automation
- ✅ Configure Git globally with SHA-256 object format
- ✅ Enable background maintenance for large repositories
- ✅ Write descriptive commit messages (why, not what)
- ✅ Validate TAG references before pushing
- ✅ Use commit message templates for consistency

### DON'T:
- ❌ Commit directly to main or develop
- ❌ Squash commits during TDD cycle (wait until PR merge)
- ❌ Skip TAG references in commit messages
- ❌ Create PRs without SPEC references
- ❌ Merge PRs with failing tests or coverage <85%
- ❌ Use `git push --force` on shared branches
- ❌ Ignore merge conflicts (resolve immediately)
- ❌ Commit large binary files without Git LFS
- ❌ Mix unrelated changes in a single commit
- ❌ Use vague commit messages ("fix bug", "update code")
- ❌ Bypass quality gates to "move faster"
- ❌ Delete remote branches before verifying merge

---

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-trust` for quality gates
- Integration with `moai-alfred-git-workflow` for automated Git operations
- Integration with `moai-alfred-tag-scanning` for TAG validation
- GitHub CLI (gh) for PR automation
- Git 2.47.0+ for latest features (SHA-256, reftable, background maintenance)

---

## Works Well With

- `moai-foundation-trust` — TRUST 5 quality gates
- `moai-alfred-git-workflow` — GitFlow automation
- `moai-alfred-tag-scanning` — TAG integrity validation
- `moai-essentials-review` — Code review checklists
- `moai-essentials-debug` — Debugging support

---

## References (Latest Documentation)

**Official Git Documentation**:
- [Git 2.47.0 Release Notes](https://github.com/git/git/blob/master/Documentation/RelNotes/2.47.0.txt)
- [Git User Manual](https://git-scm.com/docs/user-manual)
- [Pro Git Book](https://git-scm.com/book/en/v2)

**GitHub CLI Documentation**:
- [GitHub CLI Manual](https://cli.github.com/manual/)
- [gh pr create](https://cli.github.com/manual/gh_pr_create)
- [gh pr merge](https://cli.github.com/manual/gh_pr_merge)

**GitFlow Resources**:
- [A Successful Git Branching Model](https://nvie.com/posts/a-successful-git-branching-model/) (original GitFlow article)
- [Atlassian GitFlow Workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow)
- [GitKraken Git Best Practices](https://www.gitkraken.com/learn/git/best-practices/git-branch-strategy)

**Modern Git Workflows (2025)**:
- [Git Workflow Best Practices 2025](https://articles.mergify.com/git-workflow-best-practices/)
- [Trunk-Based Development](https://trunkbaseddevelopment.com/)

_Documentation links updated 2025-10-22_

---

## Changelog

- **v2.0.0** (2025-10-22): Major expansion with Git 2.47.0 features, GitHub CLI automation, comprehensive GitFlow guidance, TAG-driven commits, conflict resolution strategies, GitHub Actions integration, and MoAI-ADK workflow integration
- **v1.0.0** (2025-03-29): Initial Skill release with basic GitFlow concepts

---

## License

This Skill is part of the MoAI-ADK project and follows the same license terms.
