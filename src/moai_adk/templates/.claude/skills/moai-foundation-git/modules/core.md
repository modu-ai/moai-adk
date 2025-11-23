# Check MIDX status

git verify-pack -v .git/objects/pack/multi-pack-index

````

**Use case**: MoAI-ADK monorepo optimization

```bash
# Before optimization
$ git gc --aggressive
Counting objects: 250000
Packing objects: 100%
Duration: 45 seconds

# After MIDX optimization
$ git gc --aggressive
Counting objects: 250000
Packing objects: 100%
Duration: 28 seconds  (38% faster)
````

### Feature 2: Branch Base Detection

**What it does**: Identify which branch a commit likely originated from.

```bash
# Old way (complex):
git for-each-ref --format='%(refname:short) %(objectname)'

# New way (Git 2.47+):
git for-each-ref --format='%(if)%(is-base:develop)%(then)Based on develop%(else)Not base%(end)'

# Example output:
refs/heads/feature/SPEC-001  Based on develop
refs/heads/feature/SPEC-002  Based on develop
refs/heads/hotfix/urgent-bug Not base
```

### Feature 3: Experimental Commands (Git 2.48+)

**Git backfill** - Smart partial clone fetching:

```bash
git backfill --lazy
# Fetches only necessary objects for working directory
# Use case: Monorepos with 50K+ files
```

**Git survey** - Identify repository data shape issues:

```bash
git survey
# Output:
# Monorepo efficiency: 87%
# Largest directories: src/legacy (45MB), docs (12MB)
# Recommendation: Consider sparse checkout for legacy
```

## Git 2.48-2.50 Latest Features

| Version  | Feature                       | Benefit                   |
| -------- | ----------------------------- | ------------------------- |
| **2.48** | Experimental backfill         | 30% faster on monorepos   |
| **2.48** | Improved reftable support     | Better concurrent access  |
| **2.49** | Platform compatibility policy | Stable C11 support        |
| **2.49** | VSCode mergetool integration  | Native IDE support        |
| **2.50** | Enhanced ref verification     | Stronger integrity checks |

## GitHub CLI 2.83.0 Integration

### New Features (November 2025)

```bash
# Feature 1: Copilot Agent Support
gh agent-task create --custom-agent my-agent "Review code for security"

# Feature 2: Enhanced Release Management
gh release create v1.0.0 --notes "Release notes" --draft

# Feature 3: Improved PR Automation
gh pr create --title "Feature" --body "Description" --base develop

# Feature 4: Workflow Improvements (up to 10 nested reusable workflows)
gh workflow run ci.yml --ref develop
```

### Common Operations

```bash
# Create draft PR
gh pr create --draft --title "WIP: Feature Name" --base develop

# List open PRs with author
gh pr list --author @me --state open

# Merge PR with squash
gh pr merge 123 --squash --delete-branch

# View PR reviews
gh pr view 123 --json reviews

# Merge multiple related PRs
gh pr merge 123 --squash && gh pr merge 124 --squash
```

## Enterprise Commit Cycle (MoAI-ADK)

### Complete Flow

```
/moai:1-plan "Feature name"
  └─→ Create feature/SPEC-XXX branch
  └─→ Ask: Which workflow? (Feature Branch or Direct)
  └─→ Create SPEC document

/moai:2-run SPEC-XXX
  ├─→ RED phase: Write failing tests
  ├─→ GREEN phase: Implement code
  └─→ REFACTOR phase: Improve code

/moai:3-sync auto SPEC-XXX
  ├─→ Run quality gates (coverage ≥85%)
  ├─→ Create PR (if Feature Branch workflow)
  │   └─→ gh pr create --base develop
  ├─→ Generate documentation
  └─→ Merge to develop (if ready)
      └─→ gh pr merge --squash --delete-branch
```

## Configuration

**Location**: `.moai/config/config.json`

```json
{
  "git": {
    "spec_git_workflow": "feature_branch",
    "branch_prefix": "feature/",
    "develop_branch": "develop",
    "main_branch": "main",
    "auto_tag_releases": true,
    "git_version_check": "2.47.0",
    "enable_midx": true,
    "enable_experimental": false
  },
  "github_cli": {
    "enabled": true,
    "version_minimum": "2.63.0",
    "copilot_agent": false
  }
}
```

**Valid `spec_git_workflow` values**:

- `"feature_branch"` - Always PR (recommended for teams)
- `"develop_direct"` - Always direct commit (fast track)
- `"per_spec"` - Ask user for each SPEC (flexible)

## Quality Gates

**Enforced before merge**:

- ✅ All tests passing (≥85% coverage)
- ✅ Linting/formatting (0 errors)
- ✅ Type checking (100%)
- ✅ TRUST 5 principles validated
- ✅ TAGs integrity verified
- ✅ Security scan passed
- ✅ No hardcoded secrets

## Performance Optimization (Git 2.47+)

### MIDX Benchmark

```
Repository: moai-adk (250K objects, 45 packfiles)

Before MIDX optimization:
- Pack time: 45s
- Repack time: 38s
- Clone time: 12s

After MIDX (Git 2.47+):
- Pack time: 28s (38% faster)
- Repack time: 22s (42% faster)
- Clone time: 9s (25% faster)

Storage overhead: +2% (acceptable tradeoff)
```

### Recommended Settings

```bash
# Enable MIDX for large repos
git config --global gc.writeMultiPackIndex true
git config --global gc.multiPackIndex true

# Use packfiles with bitmap
git config --global repack.writeBitmaps true

# Enable incremental pack files
git config --global feature.experimental true
```

## Best Practices (Enterprise )

✅ **DO**:

- Choose workflow at SPEC creation (align with team)
- Follow TDD commit phases (RED → GREEN → REFACTOR)
- Keep feature branches short-lived (<3 days)
- Squash commits when merging to develop
- Maintain test coverage ≥85%
- Verify PR checks before merge
- Use session persistence for recovery

❌ **DON'T**:

- Skip quality gates based on workflow
- Mix strategies within single feature
- Commit directly to main branch
- Force push to shared branches
- Merge without all checks passing
- Leave long-running feature branches
- Use deprecated Git versions (<2.40)

## Troubleshooting

| Issue               | Solution                                              |
| ------------------- | ----------------------------------------------------- |
| Merge conflicts     | Rebase on develop before merge                        |
| PR stuck in draft   | Use `gh pr ready` to publish                          |
| Tests failing in CI | Run tests locally before push                         |
| Large pack file     | Enable MIDX: `git config gc.writeMultiPackIndex true` |
| Session lost        | Check .moai/sessions/ for recovery checkpoints        |

## Version History

| Version   | Date       | Key Changes                                                 |
| --------- | ---------- | ----------------------------------------------------------- |
| **4.0.0** | 2025-11-12 | Git 2.47-2.50 support, MIDX optimization, Hybrid strategies |
| **2.1.0** | 2025-11-04 | Three workflows (feature_branch, develop_direct, per_spec)  |
| **2.0.0** | 2025-10-22 | Major update with latest tools, TRUST 5 integration         |
