# Git Reference & Troubleshooting

Comprehensive troubleshooting guide, best practices checklist, pre-commit hooks, Git configuration optimization, and common issue resolution for production environments.

---

## Troubleshooting Guide

### Common Issues & Solutions

| Issue | Symptoms | Solution | Prevention |
|-------|----------|----------|------------|
| **Detached HEAD** | `HEAD detached at abc123` | `git switch -c recovery-branch` then `git switch main` | Always create branch before committing |
| **Merge Conflicts** | `CONFLICT (content)` during merge | Resolve conflicts, `git add`, `git commit` | Sync with base branch frequently |
| **Lost Commits** | `git log` doesn't show commit | `git reflog` to find, `git cherry-pick <hash>` | Use branches for all work |
| **Slow git status** | `git status` takes >5 seconds | Enable MIDX: `git config gc.writeMultiPackIndex true` | Enable optimizations early |
| **Large Repository** | Clone takes >10 minutes | Use `--filter=blob:none` + sparse-checkout | Apply optimizations |
| **Accidental main commit** | Committed to main instead of feature | `git reset --soft HEAD~1` then create branch | Use branch protection |
| **PR stuck in draft** | Can't mark PR ready | `gh pr ready <number>` | Check PR status before merging |
| **Permission denied** | Push rejected | Check SSH keys: `ssh -T git@github.com` | Configure SSH properly |
| **Stale branches** | Too many old branches | `gh pr list --state closed --json number \| xargs -I {} gh pr delete {}` | Auto-delete on merge |
| **Force push rejected** | Protected branch | Use `--force-with-lease` or request admin | Never force push to main |

### Advanced Troubleshooting

**Corrupt repository recovery**:
```bash
# Verify repository integrity
git fsck --full

# Fix corrupt objects
git hash-object -w <corrupted-file>

# Recover from backup
git clone --mirror <backup-url> .git
git config --bool core.bare false
git reset --hard
```

**Detached HEAD recovery**:
```bash
# You're here: HEAD detached at abc123
git log --oneline -5  # Find your commits

# Save your work
git switch -c recovery-branch

# Move commits to proper branch
git switch feature/SPEC-001
git cherry-pick abc123  # Your detached commit

# Clean up
git branch -D recovery-branch
```

**Merge conflict resolution**:
```bash
# Conflict during merge
git merge feature/SPEC-001
# CONFLICT in src/auth.py

# View conflict markers
cat src/auth.py
# <<<<<<< HEAD
# Current branch code
# =======
# Incoming branch code
# >>>>>>> feature/SPEC-001

# Resolve (choose strategy)
# Option 1: Accept theirs
git restore --theirs src/auth.py

# Option 2: Accept ours
git restore --ours src/auth.py

# Option 3: Manual resolution
# Edit src/auth.py to keep both changes

# Mark resolved
git add src/auth.py
git commit -m "merge: resolve conflicts in auth.py"
```

**Lost commit recovery**:
```bash
# Find lost commit
git reflog

# Output:
# abc123 HEAD@{0}: commit: feat(auth): add JWT
# def456 HEAD@{1}: reset: moving to HEAD~1  # Lost commit
# ghi789 HEAD@{2}: commit: test(auth): add tests

# Recover lost commit
git cherry-pick def456

# Or create branch from it
git switch -c recovered-feature def456
```

**Undo last commit (not pushed)**:
```bash
# Soft reset (keep changes staged)
git reset --soft HEAD~1

# Mixed reset (keep changes unstaged)
git reset HEAD~1

# Hard reset (discard changes)
git reset --hard HEAD~1
```

**Undo pushed commit**:
```bash
# Create revert commit
git revert HEAD
git push origin feature/SPEC-001

# Or force push (if safe)
git reset --hard HEAD~1
git push --force-with-lease origin feature/SPEC-001
```

---

## Pre-commit Hooks

### Standard Hook Template

**Location**: `.git/hooks/pre-commit`

**Comprehensive hook**:
```bash
#!/bin/bash
# .git/hooks/pre-commit

set -e  # Exit on first error

echo "Running pre-commit checks..."

# 1. Code Formatting
echo "→ Checking code formatting..."
black src/ tests/ --check --quiet || {
  echo "❌ Code formatting failed. Run: black src/ tests/"
  exit 1
}

# 2. Linting
echo "→ Running linter..."
ruff check src/ tests/ --quiet || {
  echo "❌ Linting failed. Run: ruff check --fix src/ tests/"
  exit 1
}

# 3. Type Checking
echo "→ Running type checker..."
mypy src/ --quiet || {
  echo "❌ Type checking failed. Fix type errors in src/"
  exit 1
}

# 4. Tests
echo "→ Running tests..."
pytest tests/ --quiet --tb=short || {
  echo "❌ Tests failed. Fix failing tests before committing."
  exit 1
}

# 5. Security Scan
echo "→ Running security scan..."
bandit -r src/ -ll --quiet || {
  echo "❌ Security issues detected. Review with: bandit -r src/"
  exit 1
}

# 6. Commit Message Format
echo "→ Validating commit message..."
# (Handled by commit-msg hook)

echo "✅ All pre-commit checks passed!"
```

**Make executable**:
```bash
chmod +x .git/hooks/pre-commit
```

### Auto-Formatting Hook

**Auto-fix issues before commit**:
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Format code automatically
black src/ tests/
ruff check --fix src/ tests/

# Stage formatted files
git add src/ tests/

# Run tests
pytest tests/ --quiet || exit 1

echo "✅ Code formatted and tests passed"
```

### Commit Message Hook

**Location**: `.git/hooks/commit-msg`

**Enforce Conventional Commits**:
```bash
#!/bin/bash
# .git/hooks/commit-msg

commit_msg_file="$1"
commit_msg=$(cat "$commit_msg_file")

# Conventional Commits pattern
pattern="^(feat|fix|docs|style|refactor|perf|test|chore)(\([a-z0-9\-]+\))?: .+"

if ! echo "$commit_msg" | grep -qE "$pattern"; then
  echo "❌ Invalid commit message format"
  echo ""
  echo "Format: type(scope): subject"
  echo ""
  echo "Types: feat, fix, docs, style, refactor, perf, test, chore"
  echo "Example: feat(auth): add JWT validation"
  exit 1
fi

# Check subject line length
subject_line=$(echo "$commit_msg" | head -n 1)
if [ ${#subject_line} -gt 72 ]; then
  echo "❌ Subject line too long: ${#subject_line} characters (max: 72)"
  exit 1
fi

echo "✅ Commit message format valid"
```

**Make executable**:
```bash
chmod +x .git/hooks/commit-msg
```

---

## Git Configuration Optimization

### Global Configuration

**Apply recommended settings**:
```bash
#!/bin/bash
# setup-git-config.sh

# User identity
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Default branch
git config --global init.defaultBranch main

# Editor
git config --global core.editor "code --wait"  # VS Code
# OR: git config --global core.editor "vim"

# Performance optimizations
git config --global gc.writeMultiPackIndex true
git config --global gc.multiPackIndex true
git config --global core.commitGraph true
git config --global gc.writeCommitGraph true
git config --global core.fsmonitor true
git config --global core.untrackedCache true
git config --global pack.threads 4
git config --global index.threads true

# Pull strategy
git config --global pull.rebase true  # Rebase on pull

# Push strategy
git config --global push.default simple
git config --global push.autoSetupRemote true

# Aliases
git config --global alias.sw switch
git config --global alias.swc 'switch -c'
git config --global alias.rst restore
git config --global alias.rsts 'restore --staged'
git config --global alias.st status
git config --global alias.br branch
git config --global alias.lg 'log --oneline --graph --all'

# Merge tool
git config --global merge.tool vscode
git config --global mergetool.vscode.cmd 'code --wait $MERGED'

# Diff tool
git config --global diff.tool vscode
git config --global difftool.vscode.cmd 'code --wait --diff $LOCAL $REMOTE'

# Auto-correct typos
git config --global help.autoCorrect 10  # Wait 1 second before correcting

# Colors
git config --global color.ui auto

# Whitespace handling
git config --global core.whitespace trailing-space,space-before-tab
git config --global apply.whitespace fix

echo "✅ Git configuration optimized"
```

### Repository-Specific Configuration

**Configure per-repository**:
```bash
cd ~/projects/moai-adk

# Set repository-specific email
git config user.email "work@company.com"

# Enable pre-commit hooks
git config core.hooksPath .git/hooks

# Repository-specific optimizations
git config gc.writeMultiPackIndex true
git config core.fsmonitor true
```

---

## Best Practices Checklist

### ✅ DO - Daily Workflow

**Branch Management**:
- [ ] Use `git switch` instead of `checkout`
- [ ] Create descriptive branch names: `feature/SPEC-{id}-{description}`
- [ ] Keep branches short-lived (<3 days active work)
- [ ] Sync with base branch daily: `git pull origin develop`
- [ ] Delete merged branches immediately

**Commit Practices**:
- [ ] Follow Conventional Commits 2025 format
- [ ] Write clear, descriptive commit messages (imperative mood)
- [ ] Keep commits atomic (single logical change)
- [ ] Reference SPEC/issue numbers in commits
- [ ] Run tests before committing

**PR Workflow**:
- [ ] Create PR with `gh pr create --generate-description`
- [ ] Request specific reviewers
- [ ] Address all review comments before merge
- [ ] Squash commits on merge for clean history
- [ ] Delete branch after merge

**Code Quality**:
- [ ] Run linter before commit
- [ ] Ensure test coverage ≥85%
- [ ] Update documentation with code changes
- [ ] Verify CI passes before requesting review

### ❌ DON'T - Antipatterns

**Dangerous Operations**:
- [ ] ❌ Force push to main/develop
- [ ] ❌ Commit directly to protected branches
- [ ] ❌ Skip tests before committing
- [ ] ❌ Use generic commit messages ("fix bug", "update code")
- [ ] ❌ Leave long-running branches (>1 week)

**Performance**:
- [ ] ❌ Clone full history for CI/CD
- [ ] ❌ Commit large binary files
- [ ] ❌ Skip repository optimization (repos >500MB)
- [ ] ❌ Ignore slow `git status` (>5 seconds)

**Workflow**:
- [ ] ❌ Mix multiple features in one PR
- [ ] ❌ Merge without review (unless solo project)
- [ ] ❌ Create PRs without tests
- [ ] ❌ Leave stale PRs open (>2 weeks)

---

## Session Recovery

### Save Checkpoint

**Save work-in-progress**:
```bash
# Save current work with descriptive message
git stash push -m "CHECKPOINT: WIP on SPEC-001 JWT validation"

# Or save including untracked files
git stash push --include-untracked -m "CHECKPOINT: Complete auth feature"

# View stash list
git stash list
# Output:
# stash@{0}: On feature/SPEC-001: CHECKPOINT: WIP on SPEC-001
# stash@{1}: On develop: CHECKPOINT: Database migration
```

### Recovery After Crash

**Restore session**:
```bash
# List saved checkpoints
git stash list

# Preview stash contents
git stash show -p stash@{0}

# Apply stash (keep in stash list)
git stash apply stash@{0}

# Or pop stash (remove from list)
git stash pop stash@{0}

# Apply to different branch
git switch feature/SPEC-002
git stash apply stash@{1}
```

**Clean up stash**:
```bash
# Drop specific stash
git stash drop stash@{0}

# Clear all stashes
git stash clear
```

---

## Performance Benchmarks

**Optimized vs Unoptimized Repository**:

| Metric | Unoptimized | Optimized | Improvement |
|--------|-------------|-----------|-------------|
| .git size | 1.2GB | 450MB | 63% smaller |
| git status | 12s | 0.3s | 98% faster |
| git log --all | 18s | 2.3s | 87% faster |
| git gc | 45s | 28s | 38% faster |
| Clone time | 8 min | 2 min | 75% faster |

**Optimization applied**:
- MIDX enabled
- Commit-graph enabled
- FSMonitor enabled
- Sparse-checkout (when applicable)

---

## Quick Reference Commands

**Most Used Operations**:
```bash
# Branch operations
git switch develop                    # Switch to develop
git switch -c feature/SPEC-001       # Create and switch
git switch -                         # Return to previous branch

# File operations
git restore src/auth.py              # Discard changes
git restore --staged src/auth.py     # Unstage file

# Commit operations
git commit -m "feat(auth): add JWT"  # Commit with message
git commit --fixup HEAD              # Fixup previous commit

# PR operations
gh pr create --generate-description  # Create PR with AI
gh pr merge 123 --squash --delete-branch  # Squash and merge

# Optimization
git gc --aggressive                  # Full cleanup
git repack -ad --write-midx         # Create MIDX
```

---

## Emergency Recovery

**Corrupt repository**:
```bash
# 1. Verify corruption
git fsck --full

# 2. Clone backup
git clone --mirror https://github.com/user/repo.git backup.git

# 3. Restore
cd backup.git
git config --bool core.bare false
git reset --hard
```

**Lost work recovery**:
```bash
# Find lost commits
git reflog

# Recover specific commit
git cherry-pick <lost-commit-hash>

# Or create branch
git switch -c recovery-branch <lost-commit-hash>
```

---

**Last Updated**: 2025-11-23
**Lines**: 318
