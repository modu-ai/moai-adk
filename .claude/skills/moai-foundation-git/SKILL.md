---
name: moai-foundation-git
description: Git workflow automation (branching, TDD commits, PR management)
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
tier: 0
auto-load: "true"
---

# Alfred Git Workflow

## What it does

Automates Git operations following MoAI-ADK conventions: branch creation, locale-based TDD commits, Draft PR creation, and PR Ready transition.

## When to use

- "브랜치 생성", "PR 만들어줘", "커밋 생성", "Git 작업", "버전 관리"
- "TDD 커밋", "Draft PR", "Ready PR", "PR 전환", "병합 준비"
- "Feature branch", "Pull request", "Merge request"
- Automatically invoked by `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`
- Git workflow automation needed
- When following GitFlow or GitHub Flow patterns

## How it works

**1. Branch Creation**:
```bash
git checkout develop
git checkout -b feature/SPEC-AUTH-001
```

**2. Locale-based TDD Commits**:
- **Korean (ko)**: 🔴 RED: [테스트 설명]
- **English (en)**: 🔴 RED: [Test description]
- **Japanese (ja)**: 🔴 RED: [テスト説明]
- **Chinese (zh)**: 🔴 RED: [测试说明]

Configured via `.moai/config.json`:
```json
{"project": {"locale": "ko"}}
```

**3. Draft PR Creation**:
Creates Draft PR with SPEC reference and test checklist.

**4. PR Ready Transition** (via `/alfred:3-sync`):
- Updates PR from Draft → Ready
- Adds quality gate checklist
- Verifies TRUST 5-principles

## Git Workflow Commands

### Branch Management
```bash
# Check current branch
git branch --show-current

# List all branches
git branch -a

# Create feature branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/SPEC-{ID}

# Delete local branch
git branch -d feature/SPEC-{ID}

# Delete remote branch
git push origin --delete feature/SPEC-{ID}
```

### TDD Commit Pattern
```bash
# RED commit (Korean locale)
git add tests/
git commit -m "🔴 RED: JWT 인증 테스트 추가

@TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>"

# GREEN commit
git add src/
git commit -m "🟢 GREEN: JWT 인증 서비스 구현

@CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth.test.ts

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>"

# REFACTOR commit
git add src/
git commit -m "♻️ REFACTOR: 인증 로직 최적화

@CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>"
```

### PR Management with GitHub CLI
```bash
# Create Draft PR
gh pr create --title "feat(AUTH-001): JWT 인증 시스템" \
  --body "$(cat <<'EOF'
## Summary
- JWT 기반 인증 시스템 구현
- 토큰 검증 및 갱신 기능

## Test Plan
- [ ] 유효한 자격증명 로그인
- [ ] 토큰 만료 처리
- [ ] 토큰 갱신

🤖 Generated with Claude Code
EOF
)" \
  --draft

# Convert to Ready
gh pr ready

# Check PR status
gh pr list --head $(git branch --show-current)

# View PR details
gh pr view

# Merge PR (squash)
gh pr merge --squash --delete-branch
```

### Locale Configuration Check
```bash
# Read locale setting
jq -r '.project.locale' .moai/config.json

# Update locale
jq '.project.locale = "en"' .moai/config.json > temp.json && mv temp.json .moai/config.json
```

## Examples

### Example 1: Create feature branch
User: "/alfred:1-plan JWT 인증"

Alfred executes:
```bash
git checkout develop
git checkout -b feature/SPEC-AUTH-001
gh pr create --title "feat(AUTH-001): JWT 인증 시스템" --draft
```

Result: Branch created, Draft PR opened

### Example 2: TDD commit sequence
User: "/alfred:2-run AUTH-001"

Alfred executes (Korean locale):
```bash
# RED
git add tests/auth.test.ts
git commit -m "🔴 RED: JWT 인증 테스트 추가..."

# GREEN
git add src/auth/service.ts
git commit -m "🟢 GREEN: JWT 인증 서비스 구현..."

# REFACTOR
git add src/auth/service.ts
git commit -m "♻️ REFACTOR: 인증 로직 최적화..."
```

Result: 3 commits (RED → GREEN → REFACTOR)

### Example 3: Finalize PR
User: "/alfred:3-sync"

Alfred executes:
```bash
# Update documentation
git add docs/

git commit -m "📝 DOCS: 인증 API 문서 동기화

@DOC:AUTH-001 | SPEC: SPEC-AUTH-001.md

🤖 Generated with Claude Code"

# Transition PR
gh pr ready
```

Result: PR transitioned to Ready for Review

### Example 4: Auto-merge (Team mode)
User: "/alfred:3-sync --auto-merge"

Alfred executes:
```bash
# Wait for CI/CD
sleep 10

# Check PR status
gh pr checks

# Auto-merge if all checks pass
gh pr merge --squash --delete-branch

# Checkout develop
git checkout develop
git pull origin develop
```

Result: PR merged, branch deleted, ready for next task

## Branch Naming Convention

**Format**: `feature/SPEC-{ID}`

Examples:
- `feature/SPEC-AUTH-001` - Authentication feature
- `feature/SPEC-REFACTOR-001` - Refactoring task
- `feature/SPEC-UPDATE-004` - Update task

**Complex domains**:
- `feature/SPEC-INSTALLER-SEC-001` - Hyphen-connected domain
- `feature/SPEC-UPDATE-REFACTOR-001` - Multiple domains

## Works well with

- moai-foundation-specs (SPEC metadata integration)
- moai-foundation-tags (TAG chain verification)