# @DOC:AGENT-GIT-001 | Chain: @SPEC:DOCS-003 -> @DOC:AGENT-001

# git-manager 🚀

**페르소나**: 릴리스 엔지니어
**전문 영역**: Git 워크플로우 및 GitFlow 보호 정책

## 역할

Git 커밋, 브랜치 관리, PR 생성을 자동화하며, GitFlow 정책에 따라 main 브랜치를 보호합니다.

## GitFlow Main 브랜치 보호 정책

### 핵심 규칙

- **develop만 main으로 머지 가능**: Feature 브랜치는 항상 develop으로 PR 생성
- **직접 push 차단**: pre-push hook으로 main 브랜치 직접 push 자동 차단
- **강제 push 불가**: 어떤 경우에도 main 브랜치에 강제 push 불가
- **모든 변경사항 추적 가능**: 모든 main 변경은 develop을 거쳐 이력 남음

### 적용 범위

- Personal/Team 모드 모두 적용
- 로컬 git 명령 자동 차단
- 원격 저장소 보호 규칙과 시너지

## 호출 방법

```bash
@agent-git-manager "커밋 작업"
```

## Main 브랜치 변경 프로세스

### Feature 개발 (일반 개발자)
```
develop (기본)
  ↓
git checkout -b feature/SPEC-{ID}  (develop에서 분기)
  ↓
작업 및 커밋
  ↓
git push origin feature/SPEC-{ID}
  ↓
PR 생성: feature/SPEC-{ID} → develop (main 아님!)
  ↓
develop으로 머지
```

### Release (Release 엔지니어만)
```
develop (안정적 상태)
  ↓
git checkout -b release/v{VERSION}  (선택사항)
  ↓
테스트 및 버그 수정
  ↓
PR 생성: develop → main (develop에서만 가능)
  ↓
main으로 머지 + 태그 생성
```

## 기술 구현

### Git Hooks 자동 보호

**Pre-commit Hook** (`.git/hooks/pre-commit`)
- main/master 브랜치에서의 직접 커밋 차단
- 모든 커밋 시 자동 실행

**Pre-push Hook** (`.git/hooks/pre-push`)
- main 브랜치로의 직접 push 차단
- develop 또는 release/* 브랜치만 main으로 push 허용
- 강제 push 차단
- main 브랜치 삭제 차단

### 예외 상황

**만약 실수로 main 브랜치에 커밋했다면**:
```bash
git reset --soft HEAD~1
git checkout develop
git checkout -b feature/SPEC-{ID}
git add .
git commit -m "..."
git push origin feature/SPEC-{ID}
```

---

**다음**: [debug-helper →](debug-helper.md)
