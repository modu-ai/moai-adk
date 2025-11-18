# Selection-Based GitHub Flow - Detailed Guide

**Complete guide for Personal Mode (1-2 developers) and Team Mode (3+ developers) Git workflows in MoAI-ADK v0.26.0+.**

> **See also**: CLAUDE.md → "Selection-Based GitHub Flow (간략)" for quick overview

---

## Personal Mode (1-2명 개발자)

### 활성화 및 특징

```json
// config.json에서 활성화
{
  "git_strategy": {
    "personal": { "enabled": true, "base_branch": "main" }
  }
}
```

**특징**:
- **워크플로우**: GitHub Flow (단순하고 빠름)
- **베이스 브랜치**: `main`
- **Feature 브랜치**: `feature/SPEC-XXX` → `main` (직접 merge)
- **릴리스**: main 태그 → CI/CD로 PyPI 자동 배포
- **릴리스 주기**: ~10분 (매우 빠름)
- **코드 리뷰**: 선택사항 (PR 생성 여부 선택)

### 워크플로우 예시

```bash
# Personal Mode 워크플로우
git checkout main
git checkout -b feature/SPEC-001
# ... 개발 및 테스트 ...
git push origin feature/SPEC-001

# Option A: PR 생성 (권장)
gh pr create --title "SPEC-001: Feature Name"

# Option B: 직접 merge
git checkout main
git merge --no-ff feature/SPEC-001

# Tag and deploy
git tag v0.26.1
git push origin --tags
# → CI/CD: PyPI 배포 자동 실행
```

### 장점

- 간단한 Git 구조 (main만 관리)
- 빠른 릴리스 사이클 (10분)
- 병합 충돌 최소화
- 1-2인 팀에 완벽 최적화
- 신속한 개발 가능

---

## Team Mode (3명 이상 개발자)

### 활성화 및 특징

```json
// config.json에서 활성화
{
  "git_strategy": {
    "team": {
      "enabled": true,
      "base_branch": "main",
      "require_review": true,
      "min_reviewers": 1
    }
  }
}
```

**특징**:
- **워크플로우**: GitHub Flow (동일하게 단순함)
- **베이스 브랜치**: `main`
- **Feature 브랜치**: `feature/SPEC-XXX` → `main` (PR + 리뷰 필수)
- **릴리스**: main 태그 → CI/CD로 PyPI 자동 배포
- **릴리스 주기**: ~15-20분 (리뷰 프로세스 포함)
- **코드 리뷰**: 필수 (min_reviewers: 1명 이상)

### 워크플로우 예시

```bash
# Team Mode 워크플로우
git checkout main
git checkout -b feature/SPEC-001
# ... 개발 및 테스트 ...
git push origin feature/SPEC-001

# PR 생성 (필수)
gh pr create --title "SPEC-001: Feature Name"

# Code Review (최소 1명 필수)
# → 팀원이 코드 리뷰 및 승인

# Merge (자동 또는 수동)
gh pr merge feature/SPEC-001

# Tag and deploy
git tag v0.26.1
git push origin --tags
# → CI/CD: PyPI 배포 자동 실행
```

### 장점

- 간단하면서도 체계적인 리뷰 프로세스
- 병렬 개발 지원 (모두 main 기반)
- 저렴한 인지 부하 (main만 관리)
- 빠른 배포 사이클 유지 (15-20분)
- 팀 협업 최적화

---

## 모드 전환

```bash
# Personal → Team 전환:
# config.json에서:
# personal.enabled: true → false
# team.enabled: false → true

# 자동 전환 없음: 명시적으로 설정해야 함
```

---

## Alfred × Selection-Based Workflow 통합

### /alfred 명령어의 동작

모든 Alfred 명령어는 활성화된 모드에 맞춰 작동합니다:

```bash
# /moai:1-plan → 활성화된 모드에 맞는 Branch 생성
# /moai:2-run → GitHub Flow 기반 TDD 구현
# /moai:3-sync → main 기반 sync
```

### 예시: SPEC-AUTH-001 구현

#### Personal Mode
```bash
/moai:1-plan "사용자 인증"
# → feature/SPEC-AUTH-001 (from main) 생성

/moai:2-run SPEC-AUTH-001
# → main에 직접 merge 준비

/moai:3-sync auto
# → main 태그 → PyPI 배포 (10분)
```

#### Team Mode
```bash
/moai:1-plan "사용자 인증"
# → feature/SPEC-AUTH-001 (from main) 생성

/moai:2-run SPEC-AUTH-001
# → main에 PR, 리뷰 대기

/moai:3-sync auto
# → Code Review (최소 1명) → Merge → Tag → Deploy (15-20분)
```

---

## Best Practices

### Personal Mode에서

- ✅ 신속한 개발에 집중
- ✅ 필요시 PR 생성하여 기록 관리
- ✅ 자동 배포로 빠른 피드백 루프

### Team Mode에서

- ✅ 모든 PR은 최소 1명 리뷰 필수
- ✅ 병렬 개발로 팀 생산성 향상
- ✅ main 기반 작업으로 병합 충돌 최소화
- ✅ 자동 배포로 일관된 릴리스

---

**Last Updated**: 2025-11-18
**Format**: Markdown | **Language**: English
**Scope**: Selection-Based GitHub Flow (Personal/Team Mode)
**Version**: v0.26.0+ (GitHub Flow for both modes)
