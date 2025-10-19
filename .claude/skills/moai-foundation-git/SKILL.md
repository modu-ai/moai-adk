---
name: moai-foundation-git
description: Git workflow automation (branching, TDD commits, PR management)
allowed-tools:
- Read
- Bash
- Write
- Edit
- TodoWrite
---

# Alfred Git Workflow

## What it does

Automates Git operations following MoAI-ADK conventions: branch creation, locale-based TDD commits, Draft PR creation, and PR Ready transition.

## When to use

- "브랜치 생성", "PR 만들어줘", "커밋 생성"
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

## Examples

### Example 1: Create feature branch
User: "/alfred:1-plan JWT 인증"
Claude: (creates `feature/SPEC-AUTH-001` branch and Draft PR)

### Example 2: TDD commit
User: "/alfred:2-run AUTH-001"
Claude: (commits with locale-specific format: 🔴 RED, 🟢 GREEN, ♻️ REFACTOR)

### Example 3: Finalize PR
User: "/alfred:3-sync"
Claude: (transitions PR to Ready state with quality report)