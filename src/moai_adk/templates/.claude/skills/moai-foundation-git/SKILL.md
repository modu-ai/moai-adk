---
name: moai-foundation-git
description: Enterprise Git 2.47-2.50 Workflow Foundation with GitHub CLI 2.51+, Conventional Commits 2025, intelligent branching strategies
version: 1.1.0
modularized: true
tags:
  - core-concepts
  - principles
  - enterprise
  - git
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**Name**: moai-foundation-git
**Domain**: Version Control & Git Workflows
**Freedom Level**: high
**Target Users**: DevOps engineers, Git administrators, development teams
**Invocation**: Skill("moai-foundation-git")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed guides)
**Last Updated**: 2025-11-23
**Modularized**: true

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Enterprise Git workflow automation with latest Git/GitHub features for SPEC-first TDD development.

**Key Capabilities**:
- Git 2.47-2.50 latest features (worktree, sparse-checkout, switch/restore)
- GitHub CLI 2.51+ AI-powered PR workflows
- Conventional Commits 2025 standard
- Intelligent branching strategies (Feature Branch, Direct Commit, Per-SPEC)
- TDD commit phases (RED → GREEN → REFACTOR)
- Git performance optimization (MIDX, partial clones)
- Session persistence & recovery

**Core Tools**:
- Git 2.47+ with modern commands
- GitHub CLI 2.51+ for automation
- Pre-commit hooks for validation
- Branching strategy patterns

---

## 📚 Core Patterns (5-10 minutes)

### Pattern 1: Git 2.47+ Latest Features

**Key Concept**: Modern Git commands for safer, clearer operations.

**Approach**:
```bash
# Git switch/restore (modern alternative to checkout)
git switch -c feature/SPEC-001        # Create & switch branch (clear)
git restore src/auth.py               # Restore file (clear)
git restore --source=HEAD~1 src/app.py # From specific commit

# git worktree (parallel branch work)
git worktree add ../moai-adk-feature1 feature/SPEC-001
cd ../moai-adk-feature1
# Work on feature in parallel
git worktree remove ../moai-adk-feature1

# git sparse-checkout (monorepo optimization)
git sparse-checkout init --cone
git sparse-checkout set src/moai_adk tests/
# Clone: 70% faster, 85% smaller
```

**Use Case**: Switching branches without confusion, parallel development, large monorepos.

### Pattern 2: GitHub CLI 2.51+ Automation

**Key Concept**: Automate PR workflows with AI-powered descriptions.

**Approach**:
```bash
# Create PR with AI-generated description
gh pr create \
  --base develop \
  --title "feat(auth): implement user authentication" \
  --generate-description

# Draft PR for early feedback
gh pr create --draft --base develop --title "WIP: Auth system"

# Auto-merge when CI passes
gh pr merge 123 --auto --squash --delete-branch

# Bulk operations
gh pr list --author @me --state open
gh pr diff 123
```

**Use Case**: Reducing manual PR documentation, CI-based automatic merging, batch PR operations.

### Pattern 3: Conventional Commits 2025

**Key Concept**: Standardized commit messages for automation & readability.

**Approach**:
```bash
# Format: type(scope): subject
git commit -m "feat(auth): add JWT token validation"
git commit -m "fix(api): handle null user in login"
git commit -m "docs(readme): update installation steps"
git commit -m "refactor(auth): extract password hashing"
git commit -m "perf(db): add indexes for user queries"

# Breaking changes
git commit -m "feat(api)!: change login endpoint to /auth/login

BREAKING CHANGE: Old /login removed, use /auth/login"

# Types: feat, fix, docs, style, refactor, perf, test, chore
```

**Use Case**: Automated changelog generation, CI validation, clear history documentation.

### Pattern 4: Intelligent Branching Strategies

**Key Concept**: Choose workflow based on team/project needs.

**Approach**:
```bash
# Strategy 1: Feature Branch + PR (Teams)
git switch -c feature/SPEC-001
# TDD: RED → GREEN → REFACTOR commits
gh pr create --base develop --generate-description
# After approval: merge via PR

# Strategy 2: Direct Commit (Individual/Fast Track)
git switch develop
# TDD cycle directly on develop
git push origin develop

# Strategy 3: Per-SPEC Choice (Flexible)
# Decide during /moai:1-plan based on SPEC requirements
```

**Use Case**: Team vs. individual development, rapid prototyping vs. formal review processes.

### Pattern 5: Git Performance Optimization

**Key Concept**: Speed up large repositories with modern Git features.

**Approach**:
```bash
# Multi-Pack Indexes (MIDX) - 38% faster operations
git config --global gc.writeMultiPackIndex true
git repack -ad --write-midx

# Shallow Clones - 73% smaller downloads
git clone --depth 100 https://github.com/user/moai-adk.git

# Partial Clones (blob-less) - 81% smaller
git clone --filter=blob:none https://github.com/user/moai-adk.git

# Sparse-checkout - For monorepos
git sparse-checkout set src/moai_adk tests/
```

**Use Case**: CI/CD environments, developers with limited bandwidth, large monorepos.

---

## 📖 Advanced Documentation

This Skill uses Progressive Disclosure. For detailed patterns:

- **[modules/git-latest-features.md](modules/git-latest-features.md)** - Git 2.47-2.50 detailed commands
- **[modules/github-cli-workflows.md](modules/github-cli-workflows.md)** - GitHub CLI 2.51+ automation
- **[modules/conventional-commits.md](modules/conventional-commits.md)** - Commit message standards & validation
- **[modules/branching-strategies.md](modules/branching-strategies.md)** - Workflow patterns & selection
- **[modules/performance-optimization.md](modules/performance-optimization.md)** - MIDX, sparse-checkout, cloning optimization
- **[modules/reference.md](modules/reference.md)** - Troubleshooting, best practices, hooks

---

## 🎯 TDD Workflow Integration

**Step 1**: `/moai:1-plan` creates feature branch
**Step 2**: RED → GREEN → REFACTOR commits with Conventional Commits
**Step 3**: `/moai:3-sync` creates PR with AI description
**Step 4**: Auto-merge on CI pass
**Step 5**: Session recovery via git stash if needed

---

## 🔗 Integration with Other Skills

**Complementary Skills**:
- Skill("moai-core-workflow") - MoAI-ADK orchestration
- Skill("moai-foundation-trust") - TRUST 5 quality gates
- Skill("moai-tdd-implementer") - RED-GREEN-REFACTOR patterns

---

## 📈 Version History

**1.1.0** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure
- ✨ 5 Core Patterns highlighted
- ✨ Modularized advanced content

**1.0.0** (2025-11-22)
- ✨ Git 2.47-2.50 support
- ✨ GitHub CLI 2.51+ features
- ✨ Conventional Commits 2025

---

**Maintained by**: alfred
**Domain**: Version Control
**Generated with**: MoAI-ADK Skill Factory
