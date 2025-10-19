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

- "브랜치 생성", "PR 만들어줘", "커밋 생성", "Git 워크플로우", "TDD 커밋", "풀 리퀘스트"
- "Create branch", "Pull request", "Git workflow", "TDD commits", "Locale-based commits"
- Automatically invoked by `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`
- Git workflow automation needed

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

## Git Workflow Details

**Branch Naming Convention**:
```
feature/SPEC-{ID}-{name}    # New features
bugfix/SPEC-{ID}-{name}     # Bug fixes
hotfix/SPEC-{ID}-{name}     # Critical production fixes
refactor/SPEC-{ID}-{name}   # Code refactoring
```

**TDD Commit Format**:
```
[emoji] [STAGE]: [description]

@TAG:[SPEC-ID]-[STAGE]
```

**Commit Message Examples**:
- **RED (Test)**: 🔴 RED: JWT 토큰 검증 테스트 작성
- **GREEN (Implementation)**: 🟢 GREEN: JWT 토큰 검증 로직 구현
- **REFACTOR (Improvement)**: ♻️ REFACTOR: JWT 검증 코드 최적화

**Draft PR Template**:
```markdown
## Summary
[Brief description of changes]

## SPEC Reference
- SPEC ID: AUTH-001
- SPEC Version: 0.0.1
- SPEC Status: draft

## Test Checklist
- [ ] @TEST:AUTH-001 written
- [ ] All tests passing
- [ ] Coverage ≥85%

## TDD Commits
- 🔴 RED: Add JWT token validation test
- 🟢 GREEN: Implement JWT token validation
- ♻️ REFACTOR: Optimize token validation logic
```

## Examples

### Example 1: Create feature branch with Draft PR (한국어)
User: "/alfred:1-plan JWT 인증 시스템"
Claude:
```bash
$ git checkout develop
$ git pull origin develop
$ git checkout -b feature/SPEC-AUTH-001-jwt

Branch created: feature/SPEC-AUTH-001-jwt

$ gh pr create --draft --title "SPEC-AUTH-001: JWT 인증 시스템" \
  --body "초기 SPEC 작성 완료..."

Draft PR #42 created: feature/SPEC-AUTH-001-jwt → develop
```

### Example 2: TDD Commits (한국어 locale)
User: "/alfred:2-run AUTH-001"
Claude:
```bash
# Stage 1: RED (Test)
$ git add tests/auth/test_service.py
$ git commit -m "🔴 RED: JWT 토큰 검증 테스트 작성

@TAG:AUTH-001-RED"

# Stage 2: GREEN (Implementation)
$ git add src/auth/service.py
$ git commit -m "🟢 GREEN: JWT 토큰 검증 로직 구현

@TAG:AUTH-001-GREEN"

# Stage 3: REFACTOR
$ git add src/auth/service.py
$ git commit -m "♻️ REFACTOR: JWT 검증 코드 최적화

@TAG:AUTH-001-REFACTOR"

Result: 3 commits created with TDD structure
```

### Example 3: Multi-locale support (English)
User: "Set project locale to en"
Claude:
```bash
$ jq '.project.locale = "en"' .moai/config.json > tmp.json && mv tmp.json .moai/config.json

Commits will now use English format:
- 🔴 RED: Add JWT token validation test
- 🟢 GREEN: Implement JWT token validation
- ♻️ REFACTOR: Optimize token validation logic
```

### Example 4: Finalize PR and transition to Ready
User: "/alfred:3-sync"
Claude:
```bash
Verifying TRUST principles...
✅ Test coverage: 89%
✅ Code quality: All constraints met
✅ TAG chain: Complete

$ gh pr ready #42

PR #42 transitioned: Draft → Ready for review
Updated title: "SPEC-AUTH-001: JWT 인증 시스템"
Updated body with quality checklist

Status: Ready for code review
```

### Example 5: Multi-language SPEC with Japanese commits
User: "/alfred:2-run PAYMENT-001" (locale: ja)
Claude:
```bash
🔴 RED: 支払い処理テストの追加

@TAG:PAYMENT-001-RED

🟢 GREEN: 支払い処理ロジックの実装

@TAG:PAYMENT-001-GREEN

♻️ REFACTOR: 支払い処理コードの改善

@TAG:PAYMENT-001-REFACTOR
```

## Git Hooks Integration (Optional)

**Pre-commit Hook**:
- Runs linter, type checker
- Validates commit message format
- Prevents secrets from being committed

**Pre-push Hook**:
- Validates TAG chain completeness
- Verifies TRUST principles
- Checks CI/CD pipeline status

## Branch Cleanup

```bash
# After PR merge
$ git checkout develop
$ git pull origin develop
$ git branch -d feature/SPEC-AUTH-001-jwt

# Force cleanup if needed
$ git branch -D feature/SPEC-AUTH-001-jwt
```

## Keywords

"브랜치 생성", "PR 만들기", "커밋 생성", "git workflow", "TDD commits", "Pull request", "locale-based commits"

## Reference

- Git workflow guide: CLAUDE.md#Git-브랜치-전략
- Commit message standard: CLAUDE.md#Git-커밋-메시지-표준
- Locale configuration: `.moai/config.json`

## Works well with

- moai-foundation-tags (TAG validation)
- moai-foundation-trust (TRUST validation)
- moai-essentials-review (code review)