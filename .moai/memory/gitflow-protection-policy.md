# GitFlow 권장 정책 (Advisory Policy)

**문서 ID**: @DOC:GITFLOW-POLICY-001
**작성일**: 2025-10-17
**상태**: Advisory (권장 사항, 강제 아님)
**적용 범위**: Personal/Team 모드 모두

---

## 개요

MoAI-ADK는 GitFlow 전략을 **권장**합니다. 이 정책은 best practice를 제시하지만, 사용자의 판단에 따라 유연하게 적용할 수 있습니다.

## 핵심 권장사항

### 1. Main 브랜치 접근 제어 (권장)

| 권장사항 | 설명 | 적용 방식 |
|---------|------|----------|
| **Develop 기반 Merge** | develop 브랜치에서 main으로 머지 권장 | Advisory ⚠️ |
| **Feature는 Develop** | Feature 브랜치는 develop에서 분기하고 develop으로 PR 생성 | Advisory ⚠️ |
| **Release 프로세스** | Release: develop -> main (Release Engineer 권장) | Advisory ⚠️ |
| **강제 Push** | Force-push 시 경고, 하지만 허용 | Warning ⚠️ |
| **직접 Push** | main 직접 push 시 경고, 하지만 허용 | Warning ⚠️ |

### 2. Git Workflow (권장)

```
┌─────────────────────────────────────────────────────────┐
│                RECOMMENDED GITFLOW                      │
└─────────────────────────────────────────────────────────┘

        develop (권장 기본 브랜치)
          ↑     ↓
    ┌─────────────────┐
    │                 │
    │ (개발자가 작업)  │
    │                 │
    ↓                 ↑
feature/SPEC-{ID}   [PR: feature -> develop]
                     [코드 리뷰 + 승인]
                     [Merge to develop]

    develop (안정적)
         ↓
         │ (Release Manager가 준비)
         ↓
    [PR: develop -> main]
    [CI/CD 검증]
    [태그 생성]
         ↓
       main (릴리스)
```

**유연성**: 프로젝트 상황에 따라 직접 main 푸시도 가능하지만, 위 워크플로우를 권장합니다.

## 기술 구현

### Pre-push Hook (Advisory Mode)

**위치**: `.git/hooks/pre-push`
**기능**: Main 브랜치 작업 시 경고 표시 (차단 안 함)

```bash
# main으로 push 시도 시:
⚠️  ADVISORY: Non-standard GitFlow detected

Current branch: feature/SPEC-123
Target branch: main

Recommended GitFlow workflow:
  1. Work on feature/SPEC-{ID} branch (created from develop)
  2. Push to feature/SPEC-{ID} and create PR to develop
  3. Merge into develop after code review
  4. When develop is stable, create PR from develop to main
  5. Release manager merges develop -> main with tag

✓ Push will proceed (flexibility mode enabled)
```

### 강제 Push Advisory

```bash
⚠️  ADVISORY: Force-push to main branch detected

Recommended approach:
  - Use GitHub PR with proper code review
  - Ensure changes are merged via fast-forward

✓ Push will proceed (flexibility mode enabled)
```

---

## 사용 사례별 워크플로우

### 사례 1: 표준 Feature 개발 (권장)

```bash
# 1. develop에서 최신 코드 받기
git checkout develop
git pull origin develop

# 2. feature 브랜치 생성 (develop에서)
git checkout -b feature/SPEC-001-new-feature

# 3. 작업 진행
# ... 코드 작성 및 테스트 ...

# 4. 커밋
git add .
git commit -m "..."

# 5. Push
git push origin feature/SPEC-001-new-feature

# 6. GitHub에서 PR 생성: feature/SPEC-001-new-feature -> develop

# 7. 코드 리뷰 및 승인 후 머지 (develop으로)
```

### 사례 2: 빠른 Hotfix (유연한 방식)

```bash
# 긴급 수정이 필요한 경우:

# 옵션 1: 권장 방식 (develop 기반)
git checkout develop
git checkout -b hotfix/critical-bug
# ... 수정 ...
git push origin hotfix/critical-bug
# PR 생성: hotfix -> develop -> main

# 옵션 2: 직접 main 수정 (허용되지만 권장하지 않음)
git checkout main
# ... 수정 ...
git commit -m "Fix critical bug"
git push origin main  # ⚠️ Advisory 경고 표시되지만 진행됨
```

### 사례 3: Release (표준 또는 유연)

```bash
# 표준 방식 (권장):
git checkout develop
gh pr create --base main --head develop --title "Release v1.0.0"

# 직접 push 방식 (허용):
git checkout develop
git push origin main  # ⚠️ Advisory 경고 표시되지만 진행됨
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

---

## 정책 모드 비교

### Strict Mode (이전 방식, 현재 비활성)

- ❌ main 직접 push 차단
- ❌ force-push 차단
- ❌ develop 외 브랜치에서 main 머지 차단

### Advisory Mode (현재 활성, v0.3.5+)

- ⚠️ main 직접 push 시 경고 + 허용
- ⚠️ force-push 시 경고 + 허용
- ⚠️ best practice 권장 + 유연성 제공
- ✅ 사용자 판단 존중

---

## 권장 체크리스트

프로젝트에 참여하는 모든 팀원은 다음을 권장합니다:

- [ ] `.git/hooks/pre-push` 파일 존재 및 실행 가능 (755 권한)
- [ ] develop 브랜치에서 feature 분기 (권장)
- [ ] PR 생성 시 대상이 develop (권장)
- [ ] Release는 develop -> main (권장)

**검증 명령**:
```bash
ls -la .git/hooks/pre-push
git branch -vv
```

---

## FAQ

**Q: develop -> main이 아닌 다른 경로로 머지할 수 있나요?**
A: 가능합니다. Advisory 경고가 표시되지만 진행됩니다. 단, develop -> main을 권장합니다.

**Q: Force-push를 할 수 있나요?**
A: 가능합니다. 경고가 표시되지만 허용됩니다. 단, 신중하게 사용하세요.

**Q: Main으로 직접 commit/push할 수 있나요?**
A: 가능합니다. Advisory 경고가 표시되지만 진행됩니다.

**Q: Hook을 완전히 비활성화할 수 있나요?**
A: 가능합니다. `.git/hooks/pre-push` 파일을 삭제하거나 실행 권한을 제거하세요.

**Q: 왜 Advisory Mode로 변경했나요?**
A: 사용자의 유연성과 판단을 존중하면서도 best practice를 권장하기 위해서입니다.

---

## 정책 업데이트 이력

| 날짜 | 내용 | 담당자 |
|------|------|--------|
| 2025-10-17 | 초기 정책 수립 (Strict Mode) | git-manager |
| 2025-10-17 | Advisory Mode로 전환 (경고만, 차단 안 함) | git-manager |

---

**이 정책은 권장사항이며, 사용자의 판단에 따라 유연하게 적용할 수 있습니다.**
**질문이나 제안사항은 팀 리드 또는 Release Engineer와 협의하세요.**
