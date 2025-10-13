# SPEC-CORE-GIT-001 구현 계획

## 우선순위별 마일스톤

### 1차 목표: GitManager 클래스 구현
- [ ] git/manager.py 생성
- [ ] is_repo, current_branch, is_dirty 메서드
- [ ] create_branch, commit, push 메서드

### 2차 목표: 브랜치 및 커밋 유틸리티
- [ ] generate_branch_name 함수
- [ ] format_commit_message 함수 (locale 지원)
- [ ] get_repo_status 함수

### 3차 목표: Draft PR 생성
- [ ] create_draft_pr 함수
- [ ] gh CLI 연동
- [ ] PR 템플릿 적용

### 최종 목표: 통합 테스트
- [ ] 실제 Git 저장소에서 테스트
- [ ] 브랜치 생성/전환 검증
- [ ] 커밋 메시지 검증

---

## 기술적 접근 방법

### GitPython vs simple-git 비교

| 작업 | simple-git (TS) | GitPython (Python) |
|------|-----------------|-------------------|
| 초기화 | `simpleGit()` | `Repo('.')` |
| 브랜치 생성 | `.checkoutBranch()` | `.git.checkout('-b')` |
| 커밋 | `.commit()` | `.index.commit()` |
| 푸시 | `.push()` | `.git.push()` |

### gh CLI 활용
- PR 생성: `gh pr create --draft`
- PR 상태 변경: `gh pr ready`
- PR 머지: `gh pr merge --squash`

---

## 아키텍처 설계 방향

### 모듈 구조
```
core/git/
├── manager.py        # GitManager 클래스
├── branch.py         # 브랜치 관리 유틸리티
├── commit.py         # 커밋 메시지 포맷팅
└── pr.py             # PR 생성/관리
```

### 에러 처리
- GitCommandError: Git 명령 실패 시
- InvalidGitRepositoryError: Git 저장소 아님
- GitCommandNotFound: git 명령 없음

---

## 리스크 및 대응 방안

### 리스크 1: GitPython 퍼포먼스
- **문제**: GitPython은 subprocess 기반으로 느릴 수 있음
- **대응**: 빈번한 호출 최소화, 필요 시 pygit2 고려

### 리스크 2: gh CLI 의존성
- **문제**: gh가 설치되지 않을 수 있음
- **대응**: doctor 명령어에서 확인, PyGithub로 fallback

### 리스크 3: locale 미지원
- **문제**: ko, en 외 다른 locale 요청
- **대응**: 기본 en으로 fallback
