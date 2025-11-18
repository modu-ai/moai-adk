# Branch Protection Configuration for Hybrid Personal-Pro Workflow

## Overview
MoAI-ADK uses a Hybrid Personal-Pro Workflow:
- **Personal Mode** (1-2 developers): GitHub Flow (feature → main)
- **Team Mode** (3+ developers): Git-Flow (feature → develop → main)

## GitHub Branch Protection Rules

### develop Branch (팀 모드 대비)

**설정 위치**: Settings → Branches → Branch protection rules → Add rule

**규칙**:
```
Branch name pattern: develop

✓ Require a pull request before merging
  - ✓ Require approval reviews before merging (1명)
  - ✓ Require status checks to pass before merging
    - Contexts: CI, code-quality, test
  - ✓ Require branches to be up to date before merging

✓ Include administrators
  - ✓ Allow force pushes: Disabled
  - ✓ Allow deletions: Disabled
```

**목적**:
- develop에 직접 커밋 방지 (feature 브랜치 강제)
- 팀 모드 전환 시 자동 적용
- 현재 개인 모드에서는 비활성화 (주석 처리)

### main Branch (항상 보호)

**규칙**:
```
Branch name pattern: main

✓ Require a pull request before merging
  - ✓ Require approval reviews before merging (1명)
  - ✓ Require status checks to pass before merging
    - Contexts: CI, code-quality, test, build
  - ✓ Require branches to be up to date before merging
  - ✓ Require code reviews before merging

✓ Require status checks to pass before merging
✓ Require conversations to be resolved before merging
✓ Include administrators
  - ✓ Allow force pushes: Disabled
  - ✓ Allow deletions: Disabled
```

**목적**:
- main 브랜치 안정성 보장
- 모든 SPEC이 CI/CD 통과 후 머지
- 프로덕션 배포 안정성

## 설정 방법

### Option A: GitHub CLI (권장)
```bash
# develop 브랜치 보호 규칙 설정
gh api repos/modu-ai/moai-adk/branches/develop/protection \
  -X PUT \
  --input branch-protection-develop.json

# main 브랜치 보호 규칙 설정
gh api repos/modu-ai/moai-adk/branches/main/protection \
  -X PUT \
  --input branch-protection-main.json
```

### Option B: GitHub Web UI
1. Repository Settings
2. Branches → Branch protection rules
3. 규칙 생성 및 설정

## 워크플로우 동작

### 개인 모드 (현재)
```
feature/SPEC-001 → main (CI/CD 자동 검증)
                    ↓
                  Tag → PyPI 배포
```

### 팀 모드 (3명 이상 시 자동 전환)
```
feature/SPEC-001 → develop (PR 필수)
                    ↓
                  develop → main (PR + 리뷰 필수)
                    ↓
                  Tag → PyPI 배포
```

## Local Git Hook

`.git/hooks/pre-commit`에서 develop 브랜치 직접 커밋을 감지하고 경고합니다:

```
⚠️  경고: develop 브랜치에 직접 커밋하려고 합니다.

개인 모드: feature/SPEC-XXX → main
팀 모드: feature/SPEC-XXX → develop → main

계속하시겠습니까? (y/n):
```

## 자동 모드 감지

`.moai/config/config.json`의 `git_strategy.team.auto_switch_threshold` 설정:

```json
{
  "git_strategy": {
    "team": {
      "auto_switch_threshold": 3  // 3명 이상이면 팀 모드 활성화
    }
  }
}
```

Alfred가 Git 기여자 수를 감지하여 자동으로 모드를 전환합니다.

---

**Last Updated**: 2025-11-18
**Version**: 0.25.11
**Workflow**: Hybrid Personal-Pro
