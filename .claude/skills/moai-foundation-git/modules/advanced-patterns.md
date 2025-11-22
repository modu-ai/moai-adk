# Advanced Git Patterns

**Version**: 4.0.0
**Focus**: Enterprise automation, branch strategies, conflict resolution

---

## Level 3: Advanced Patterns

### Pattern 1: Trunk-Based Development with Feature Flags

**Automation Strategy**:
```bash
# Trunk-based workflow
git checkout main
git pull origin main

# Create short-lived branch (24-48 hours max)
git checkout -b feature/auth-mfa-${date +%s}

# Implement with feature flag
cat > src/auth/flags.py << 'EOF'
ENABLE_MFA = os.getenv("ENABLE_MFA", "false") == "true"

def authenticate(user, password, totp=None):
    if not authenticate_password(user, password):
        return False

    if ENABLE_MFA and not verify_totp(user, totp):
        return False

    return True
EOF

# Commit and push
git add .
git commit -m "feat(auth): Add MFA support (feature-flagged)"
git push origin feature/auth-mfa-${timestamp}

# Create PR -> Review → Merge to main
# Feature flag remains OFF in production
# Gradually roll out by enabling flag in prod
```

### Pattern 2: Git Automation with GitHub Actions

**Automated Release Management**:
```yaml
# .github/workflows/semantic-release.yml
name: Semantic Release

on:
  push:
    branches: [main]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Semantic Release
        uses: cycjimmy/semantic-release-action@v3
        with:
          semantic_version: 19
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          NPM_TOKEN: ${{ secrets.NPM_TOKEN }}

      - name: Tag Validation
        run: |
          git fetch --tags
          # Verify tag matches version
          if [ -z "$(git tag -l 'v*')" ]; then
            echo "No new tag created"
            exit 1
          fi
```

### Pattern 3: Merge Conflict Resolution Strategy

**AI-Assisted Conflict Resolution**:
```bash
# Detect conflicts early
git merge --no-commit --no-ff feature/new-feature

# Use merge drivers for specific file types
git config merge.ours.driver true  # Keep ours for generated files
git config merge.theirs.driver true  # Keep theirs for package.lock

# Conflict analysis
git diff --name-only --diff-filter=U | while read file; do
  echo "=== Conflict in $file ==="
  git diff --ours "$file" | head -20
  echo "---"
  git diff --theirs "$file" | head -20
done

# Resolve with git mergetool (e.g., VS Code)
git mergetool -t vscode

# Complete merge
git commit -m "🔀 merge: Resolve conflicts with feature/new-feature"
```

### Pattern 4: Branch Strategy for Large Teams

**Multi-Branch Strategy** (Git Flow variant):
```bash
# Main branches
- main          # Production-ready (protected)
- develop       # Integration branch
- release/*     # Release branches

# Feature branches
- feature/*     # New features (from develop)
- bugfix/*      # Bug fixes (from develop)
- hotfix/*      # Critical fixes (from main)
- docs/*        # Documentation only

# Integration
feature/auth -> develop -> release/v1.2.0 -> main

# Automation per branch:
# feature/* -> PR review -> merge to develop
# develop (daily) -> release/v1.2.0 (release day)
# release/* -> final QA -> merge to main + tag
# hotfix/* -> PR -> main + develop (both)
```

### Pattern 5: Git Hooks Automation

**Pre-commit Hook** (Prevent bad commits):
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check commit message format
COMMIT_MSG="$(git diff --cached --name-only | head -1)"

# Validate TDD commit format
if ! git diff --cached --diff-filter=ACM | grep -E "🔴|🟢|♻️|🐛|✨|📝|🔒"; then
  echo "❌ Commit message must start with emoji"
  exit 1
fi

# Run tests before commit
npm test -- --testPathPattern="$(git diff --cached --name-only | grep -E '\.test\.(js|ts)$' | head -1)"
if [ $? -ne 0 ]; then
  echo "❌ Tests failed. Fix errors before committing."
  exit 1
fi

echo "✅ Pre-commit checks passed"
```

**Post-merge Hook** (Clean up after merge):
```bash
#!/bin/bash
# .git/hooks/post-merge

# Delete merged local branches
git branch --merged | grep -v "\*\|main\|develop" | xargs git branch -d

# Update dependencies if package.json changed
if git diff HEAD@{1} HEAD --name-only | grep -q "package.json"; then
  echo "Detected package.json changes. Installing..."
  npm install
fi
```

### Pattern 6: Multi-Repository Synchronization

**Monorepo Strategy with Git Subtrees**:
```bash
# Add subtree
git subtree add --prefix packages/auth \
  https://github.com/org/auth.git main

# Update subtree
git subtree pull --prefix packages/auth \
  https://github.com/org/auth.git main

# Push changes back to subtree
git subtree push --prefix packages/auth \
  https://github.com/org/auth.git feature/new-feature
```

### Pattern 7: Git Performance at Scale

**Large Repository Optimization**:
```bash
# Enable sparse checkout (for monorepos)
git config core.sparseCheckout true
echo "apps/web/" > .git/info/sparse-checkout
git checkout main

# Enable partial clone
git clone --filter=blob:none https://repo.git
git clone --filter=tree:0 https://repo.git

# Maintenance
git gc --aggressive  # Run weekly
git repack -a -d --window-memory=256m

# Monitor size
du -sh .git/
git count-objects -v
```

---

**Last Updated**: 2025-11-22
