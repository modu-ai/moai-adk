# GitFlow Main 브랜치 보호 정책

**문서 ID**: @DOC:GITFLOW-PROTECTION-001
**작성일**: 2025-10-17
**상태**: Active (모든 팀원에게 자동 적용)
**적용 범위**: Personal/Team 모드 모두

---

## 개요

MoAI-ADK는 GitFlow 전략에 따라 main 브랜치를 엄격히 보호합니다. 이 정책은 실수에 의한 main 브랜치 오염을 방지하고, 모든 변경사항의 추적성을 보장합니다.

## 핵심 규칙

### 1. Main 브랜치 접근 제어

| 규칙 | 설명 |
|------|------|
| **Develop만 Merge** | develop 브랜치에서만 main으로 머지 가능 |
| **Feature는 Develop** | Feature 브랜치는 항상 develop에서 분기하고 develop으로 PR 생성 |
| **Release 프로세스** | Release: develop -> main (Release Engineer만 수행) |
| **직접 Push 불가** | 어떤 경우에도 main으로 직접 push 불가 |
| **강제 Push 차단** | Force-push는 pre-push hook으로 차단 |
| **삭제 불가** | Main 브랜치 삭제 시도 자동 차단 |

### 2. Git Workflow

```
┌─────────────────────────────────────────────────────────┐
│                   GITFLOW WORKFLOW                      │
└─────────────────────────────────────────────────────────┘

        develop (기본 브랜치)
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

## 기술 구현

### Pre-commit Hook

**위치**: `.git/hooks/pre-commit`
**기능**: Main 브랜치에서의 커밋 차단

```bash
# main 브랜치에서 커밋 시도 시:
ERROR: Cannot commit directly on main branch

Correct workflow:
  1. git checkout develop
  2. git checkout -b feature/SPEC-{ID}
  3. Make changes and commit
  4. git push origin feature/SPEC-{ID}
  5. Create PR on GitHub
```

### Pre-push Hook

**위치**: `.git/hooks/pre-push`
**기능**: Main 브랜치로의 직접 push 차단

```bash
# main으로 push 시도 시:
ERROR: Main branch is protected by GitFlow policy
Main branch can only be pushed from 'develop' branch.

Your branch: feature/SPEC-123
Target branch: main
```

### Git Config

**설정**: `branch.main.description`
**값**: "Protected main branch - for releases only. All changes must come from develop via PR."

---

## 사용 사례별 워크플로우

### 사례 1: 새 Feature 개발

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
#    (메인 타겟은 develop, main 아님!)

# 7. 코드 리뷰 및 승인 후 머지 (develop으로)
```

### 사례 2: Release (Release Engineer만)

```bash
# 1. develop이 안정적 상태 확인
git checkout develop
git pull origin develop

# 2. 선택사항: release 브랜치 생성
git checkout -b release/v1.0.0

# 3. 버전 번호 업데이트 및 최종 테스트
# ... 테스트 및 수정 ...

# 4. 커밋 (선택사항, release 브랜치 사용 시)
git add .
git commit -m "Release v1.0.0"

# 5. develop으로 전환 (또는 develop이 이미 안정적 상태)
git checkout develop

# 6. GitHub에서 PR 생성: develop -> main
#    (develop에서만 main으로 PR 가능)

# 7. PR 리뷰 및 승인

# 8. Squash & Merge로 main에 머지
# (GitHub UI에서 자동으로 수행)

# 9. 태그 생성
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 사례 3: 실수로 main에서 작업한 경우

```bash
# 만약 main 브랜치에서 변경사항을 만들었다면:

# 1. 커밋되지 않은 변경사항을 Stash
git stash

# 2. develop으로 전환
git checkout develop

# 3. feature 브랜치 생성
git checkout -b feature/SPEC-002-fixes

# 4. Stashed 변경사항 적용
git stash pop

# 5. 정상적인 워크플로우 진행
git add .
git commit -m "..."
git push origin feature/SPEC-002-fixes
# ... PR 생성 및 리뷰 ...
```

---

## 보호 정책 검증 체크리스트

프로젝트에 참여하는 모든 팀원은 다음을 확인해야 합니다:

- [ ] `.git/hooks/pre-commit` 파일이 존재하고 실행 가능 (755 권한)
- [ ] `.git/hooks/pre-push` 파일이 존재하고 실행 가능 (755 권한)
- [ ] `git config branch.main.description` 설정 확인
- [ ] 로컬 develop 브랜치에서 `git branch -vv` 실행 후 설명 표시 여부 확인
- [ ] Feature 브랜치 생성 시 develop에서만 분기 확인
- [ ] PR 생성 시 대상이 develop인지 확인

**검증 명령**:
```bash
# 모든 설정 확인
git config --local --get branch.main.description
ls -la .git/hooks/pre-commit
ls -la .git/hooks/pre-push
```

---

## 문제 해결

### Q: "ERROR: Cannot commit directly on main branch" 오류가 나옵니다

**A**: develop 브랜치로 전환하고 feature 브랜치를 생성하세요.
```bash
git checkout develop
git checkout -b feature/SPEC-XXX-description
```

### Q: "ERROR: Main branch is protected by GitFlow policy" 오류가 나옵니다

**A**: develop 또는 release 브랜치에서 main으로 push하려고 합니다. 이는 정상입니다. GitHub PR을 통해 머지하세요.

### Q: Main으로 강제 push를 하고 싶습니다

**A**: GitFlow 정책상 불가능합니다. 필요한 경우:
1. Release Engineer에게 요청
2. `.git/hooks/pre-push` 임시 비활성화 후 재활성화
3. Main에 되돌릴 커밋 생성: `git revert {hash}`

### Q: Pre-push hook이 실행되지 않습니다

**A**: Hook 파일 권한 확인:
```bash
chmod +x .git/hooks/pre-push
chmod +x .git/hooks/pre-commit
```

---

## FAQ

**Q: develop -> main이 아닌 다른 경로로 머지할 수 있나요?**
A: 불가능합니다. GitFlow 정책은 develop -> main만 허용합니다.

**Q: Release 브랜치 (release/v1.0.0)를 사용해야 하나요?**
A: 선택사항입니다. develop이 안정적이면 develop -> main으로 직접 진행 가능합니다.

**Q: Hotfix는 어떻게 처리하나요?**
A: Hotfix는 현재 정책상 develop -> main 경로를 따릅니다. 긴급한 경우 Release Engineer와 협의하세요.

**Q: Main 브랜치 보호 정책을 우회할 수 있나요?**
A: 기술적으로는 가능하지만 (hook 삭제), GitFlow 규칙상 강력히 권장하지 않습니다. 정책 변경이 필요하면 팀과 협의하세요.

---

## 정책 업데이트 이력

| 날짜 | 내용 | 담당자 |
|------|------|--------|
| 2025-10-17 | 초기 정책 수립 및 Git hooks 구현 | git-manager |

---

**이 정책은 모든 MoAI-ADK 팀원에게 자동으로 적용됩니다.**
**정책 변경이 필요한 경우 팀 리드 또는 Release Engineer와 협의하세요.**
