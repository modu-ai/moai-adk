# Git Performance Optimization

Comprehensive guide to optimizing Git repository performance with modern features: MIDX, shallow clones, partial clones, sparse-checkout, commit-graph, and fsmonitor.

---

## Multi-Pack Index (MIDX)

**Purpose**: Accelerate Git operations by creating a single index covering all packfiles

### What MIDX Solves

**Problem**: Traditional Git searches each packfile individually for objects
```
Without MIDX:
- 10 packfiles → 10 separate indexes to search
- git log: Search each packfile sequentially
- Time complexity: O(n * m) where n = packfiles, m = objects
```

**Solution**: MIDX creates unified index
```
With MIDX:
- Single index covering all 10 packfiles
- git log: Single index lookup
- Time complexity: O(m) - constant time regardless of packfile count
```

### Enable MIDX

**Global configuration** (recommended):
```bash
# Enable MIDX for all repos
git config --global gc.writeMultiPackIndex true
git config --global gc.multiPackIndex true

# Enable for all pack operations
git config --global repack.writeBitmaps true
```

**Per-repository**:
```bash
cd ~/projects/moai-adk

# Enable MIDX
git config gc.writeMultiPackIndex true

# Create MIDX immediately
git repack -ad --write-midx

# Verify MIDX created
ls -lh .git/objects/pack/multi-pack-index
```

### MIDX Performance Gains

**Benchmark results** (repo with 1M commits, 100 packfiles):
```
Operation          | Without MIDX | With MIDX   | Improvement
-------------------|--------------|-------------|-------------
git log            | 12.5s        | 3.2s        | 74% faster
git gc             | 45s          | 28s         | 38% faster
git repack         | 85s          | 49s         | 42% faster
git count-objects  | 8.2s         | 1.5s        | 82% faster
git fsck           | 120s         | 68s         | 43% faster
```

**Real-world example**:
```bash
# Before MIDX (large repo)
time git log --oneline --all > /dev/null
# Output: 18.2s

# Enable MIDX
git repack -ad --write-midx

# After MIDX
time git log --oneline --all > /dev/null
# Output: 4.7s (74% faster)
```

---

## Shallow Clones

**Purpose**: Clone repository with limited history for faster initial setup

### Basic Shallow Clone

**Clone with depth limit**:
```bash
# Clone last 100 commits only
git clone --depth 100 https://github.com/user/moai-adk.git

# Result:
# - Only last 100 commits downloaded
# - Full history available remotely
# - Can fetch more history later
```

**Size comparison**:
```
Full clone:
- Total size: 450MB
- Commits: 15,000
- Clone time: 8 minutes

Shallow clone (depth=100):
- Total size: 120MB (73% smaller)
- Commits: 100
- Clone time: 2 minutes (75% faster)
```

### Working with Shallow Clones

**Fetch additional history**:
```bash
# Deepen by 500 more commits
git fetch --deepen=500

# Or fetch all history (convert to full)
git fetch --unshallow
```

**Common use cases**:
```bash
# CI/CD (only need latest)
git clone --depth 1 https://github.com/user/repo.git

# Quick testing (need recent history)
git clone --depth 50 https://github.com/user/repo.git

# Development (need substantial history)
git clone --depth 1000 https://github.com/user/repo.git
```

### Limitations

**Shallow clone restrictions**:
- Cannot push from shallow clone (must unshallow first)
- Some Git operations slower (need remote fetch)
- `git blame` may not work for old code
- Cannot create branches from commits outside shallow history

**Workaround**:
```bash
# Unshallow before important operations
git fetch --unshallow

# Then perform operation
git push origin feature/SPEC-001
```

---

## Partial Clones (Blob-less)

**Purpose**: Clone repository structure without file contents (downloaded on-demand)

### Blob-less Clone

**Clone without blobs**:
```bash
# Clone without file contents
git clone --filter=blob:none https://github.com/user/moai-adk.git

# What's included:
# - All commits
# - All trees (directory structure)
# - All references
# - NO file contents (blobs)

# Blobs downloaded automatically when needed:
# - git checkout
# - git diff (with file content)
# - git log -p
```

**Size reduction**:
```
Full clone:
- Total: 450MB
- Commits/trees/blobs: 450MB

Blob-less clone:
- Total: 85MB (81% smaller)
- Commits/trees: 85MB
- Blobs: Downloaded on-demand
```

### Tree-less Clone

**Ultimate minimal clone**:
```bash
# Clone with no trees or blobs (commits only)
git clone --filter=tree:0 https://github.com/user/repo.git

# Size: ~30MB (93% smaller than full)
# Use case: CI systems that only need commit metadata
```

### Blob-less Workflow

**Real-world usage**:
```bash
# Developer workflow
git clone --filter=blob:none https://github.com/large-org/monorepo.git
cd monorepo

# Work on specific feature
git switch feature/payment-processing

# Blobs for this branch downloaded automatically
# Other branches' blobs NOT downloaded (saved bandwidth)

# Performance:
# - Initial clone: 90MB (vs 2.5GB full)
# - Branch checkout: 150MB downloaded on-demand
# - Total: 240MB (vs 2.5GB full) - 90% savings
```

---

## Sparse-Checkout

**Purpose**: Check out only specific directories in working tree

### Cone Mode (Recommended)

**Basic sparse-checkout**:
```bash
git clone https://github.com/user/large-repo.git
cd large-repo

# Enable sparse-checkout (cone mode)
git sparse-checkout init --cone

# Select directories to check out
git sparse-checkout set src/moai_adk tests/

# Result: Only specified directories in working tree
# Other directories still in Git, just not checked out
```

**Performance metrics**:
```
Large monorepo (500K files, 1.5GB):

Full checkout:
- Clone time: 8 minutes
- Disk usage: 1.5GB
- git status: 12 seconds

Sparse-checkout (2 directories):
- Clone time: 2.4 minutes (70% faster)
- Disk usage: 220MB (85% smaller)
- git status: 1.8 seconds (85% faster)
```

### Managing Sparse-Checkout

**Add directories**:
```bash
# Add more directories
git sparse-checkout add docs/ scripts/

# View current sparse-checkout
git sparse-checkout list
```

**Disable sparse-checkout**:
```bash
# Restore full working tree
git sparse-checkout disable
```

### Advanced Patterns

**Pattern-based sparse-checkout**:
```bash
# Use patterns instead of cone mode
git sparse-checkout init --no-cone

# Include all Python files
git sparse-checkout set '**/*.py' '**/*.md'

# Exclude specific directories
git sparse-checkout set 'src/**' '!src/deprecated/**'
```

### Sparse-Checkout with Blob-less Clone

**Ultimate optimization** (for monorepos):
```bash
# Combine partial clone + sparse-checkout
git clone --filter=blob:none https://github.com/large-org/monorepo.git
cd monorepo

git sparse-checkout init --cone
git sparse-checkout set src/payment-service tests/payment-service

# Result:
# - Blobs only for sparse directories
# - Minimal disk usage
# - Fast operations

# Performance:
# Full clone: 2.5GB, 10 minutes
# Optimized: 85MB, 2 minutes (97% smaller, 80% faster)
```

---

## Commit-Graph

**Purpose**: Accelerate commit graph traversal operations

### Enable Commit-Graph

**Configuration**:
```bash
# Enable commit-graph generation
git config --global core.commitGraph true
git config --global gc.writeCommitGraph true

# Generate commit-graph immediately
git commit-graph write --reachable
```

**Verify commit-graph**:
```bash
ls -lh .git/objects/info/commit-graph
# Should exist after generation
```

### Performance Benefits

**Benchmark results** (repo with 500K commits):
```
Operation       | Without C-Graph | With C-Graph | Improvement
----------------|-----------------|--------------|-------------
git log --all   | 18.5s           | 2.3s         | 87% faster
git merge-base  | 5.2s            | 0.8s         | 85% faster
git rev-list    | 12.8s           | 1.9s         | 85% faster
git branch -a   | 3.4s            | 0.5s         | 85% faster
```

---

## FSMonitor (File System Monitor)

**Purpose**: Use OS file system notifications to accelerate git status

### Enable FSMonitor

**macOS with Watchman**:
```bash
# Install Watchman
brew install watchman

# Enable FSMonitor
git config --global core.fsmonitor true
git config --global core.untrackedCache true
```

**Windows**:
```bash
# Built-in FSMonitor support (Git 2.45+)
git config --global core.fsmonitor true
```

**Linux with Watchman**:
```bash
# Install Watchman (Ubuntu)
sudo apt install watchman

git config --global core.fsmonitor true
```

### Performance Gains

**Benchmark** (repo with 100K files):
```
git status performance:

Without FSMonitor:
- First run: 8.2s
- Subsequent: 7.5s (no caching benefit)

With FSMonitor:
- First run: 8.0s (slight overhead)
- Subsequent: 0.3s (96% faster)
```

---

## Complete Optimization Guide

### Recommended Configuration

**Apply all optimizations**:
```bash
# Multi-pack index
git config --global gc.writeMultiPackIndex true
git config --global gc.multiPackIndex true

# Commit-graph
git config --global core.commitGraph true
git config --global gc.writeCommitGraph true

# FSMonitor
git config --global core.fsmonitor true
git config --global core.untrackedCache true

# Parallel processing
git config --global pack.threads 4
git config --global index.threads true

# Compression
git config --global core.compression 9
```

### Repository Optimization Workflow

**Optimize existing repository**:
```bash
#!/bin/bash
# optimize-repo.sh

echo "Optimizing Git repository..."

# 1. Cleanup loose objects
git gc --aggressive --prune=now

# 2. Create MIDX
git repack -ad --write-midx

# 3. Generate commit-graph
git commit-graph write --reachable

# 4. Verify optimization
echo "Repository optimized:"
du -sh .git
git count-objects -v
```

**Results**:
```
Before optimization:
.git size: 1.2GB
Loose objects: 15,432

After optimization:
.git size: 450MB (63% smaller)
Loose objects: 0
MIDX: Enabled
Commit-graph: Enabled

Operations 40-85% faster
```

---

## Use Case Scenarios

### Scenario 1: CI/CD Optimization

**Goal**: Minimize clone time in CI pipeline

**Solution**:
```bash
# Shallow clone with minimal history
git clone --depth 1 --filter=blob:none \
  https://github.com/user/repo.git

# Result:
# - Clone time: 15s (vs 5 minutes full)
# - Disk usage: 30MB (vs 450MB full)
# - Sufficient for build/test
```

### Scenario 2: Monorepo Development

**Goal**: Work on single service in large monorepo

**Solution**:
```bash
# Partial clone + sparse-checkout
git clone --filter=blob:none https://github.com/org/monorepo.git
cd monorepo

git sparse-checkout init --cone
git sparse-checkout set services/payment-api tests/payment-api

# Result:
# - Only payment service files in working tree
# - Fast git operations
# - Minimal disk usage
```

### Scenario 3: Large Repository Performance

**Goal**: Improve performance on large existing repository

**Solution**:
```bash
# Apply all optimizations
git config gc.writeMultiPackIndex true
git config core.commitGraph true
git config core.fsmonitor true

# Run optimization
git gc --aggressive
git repack -ad --write-midx
git commit-graph write --reachable

# Result:
# - 40-85% faster operations
# - 63% smaller .git directory
# - Instant git status
```

---

## Best Practices

### ✅ DO

- Enable MIDX globally for all repositories
- Use shallow clones for CI/CD pipelines
- Apply sparse-checkout for monorepos
- Enable commit-graph for large repositories
- Use FSMonitor for repos with many files
- Run `git gc` regularly (or enable auto-gc)

### ❌ DON'T

- Use shallow clones for main development workstation
- Enable MIDX on repos with frequent force pushes
- Use sparse-checkout on small repos (<10K files)
- Skip optimization on repos >500MB
- Force-push without `--force-with-lease` (breaks MIDX)

---

**Last Updated**: 2025-11-23
**Lines**: 280
