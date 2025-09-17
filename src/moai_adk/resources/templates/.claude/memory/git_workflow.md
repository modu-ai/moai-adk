# Git 워크플로우 가이드

> MoAI-ADK 프로젝트의 표준 Git 워크플로우와 명령어

## 🔄 기본 Git 워크플로우

### 1. 저장소 초기화 및 클론
```bash
# 새 저장소 초기화
git init
git remote add origin https://github.com/user/repo.git

# 기존 저장소 클론
git clone https://github.com/user/repo.git
cd repo

# 서브모듈 포함 클론
git clone --recursive https://github.com/user/repo.git
```

### 2. 일상적인 작업 흐름
```bash
# 현재 상태 확인
git status                              # 작업 디렉토리 상태
git diff                                # 변경사항 확인
git diff --cached                       # 스테이지된 변경사항

# 변경사항 스테이징
git add .                               # 모든 변경사항
git add -p                              # 대화형 부분 스테이징
git add src/                            # 특정 디렉토리만

# 커밋
git commit -m "type(scope): description"  # 표준 커밋 메시지
git commit -am "message"                # add + commit 한번에
git commit --amend                      # 마지막 커밋 수정
```

## 📝 커밋 메시지 규칙

### MoAI-ADK 표준 커밋 형식
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### 커밋 타입
- **feat**: 새로운 기능 추가
- **fix**: 버그 수정
- **docs**: 문서 변경
- **style**: 코드 포매팅, 세미콜론 누락 등
- **refactor**: 코드 리팩토링
- **test**: 테스트 추가 또는 수정
- **chore**: 빌드 프로세스 또는 보조 도구 변경

### 예시
```bash
git commit -m "feat(auth): add JWT authentication"
git commit -m "fix(api): resolve timeout issue in user endpoint"
git commit -m "docs(readme): update installation instructions"
git commit -m "test(utils): add unit tests for date helpers"
```

## 🌿 브랜치 관리

### 브랜치 전략 (Git Flow 기반)
```bash
# 메인 브랜치
main/master                             # 프로덕션 브랜치
develop                                 # 개발 통합 브랜치

# 작업 브랜치
feature/feature-name                    # 기능 개발
bugfix/bug-description                  # 버그 수정
hotfix/critical-fix                     # 긴급 수정
release/v1.2.0                          # 릴리스 준비
```

### 브랜치 작업
```bash
# 브랜치 생성 및 전환
git checkout -b feature/user-auth       # 브랜치 생성 + 전환
git switch -c feature/user-auth         # 최신 방식

# 브랜치 전환
git checkout main                       # 브랜치 전환
git switch main                         # 최신 방식

# 브랜치 목록
git branch                              # 로컬 브랜치
git branch -r                           # 리모트 브랜치
git branch -a                           # 모든 브랜치

# 브랜치 삭제
git branch -d feature/completed         # 로컬 브랜치 삭제
git push origin --delete feature/old    # 리모트 브랜치 삭제
```

## 🛡️ 브랜치 보호 규칙(권장)

- 보호 대상: `main`(또는 `master`), `release/*` 브랜치
- 필수 검사: CI 테스트 통과(빌드/린트/테스트), 리뷰 1+ 승인, 상태 체크 필수화
- 강제 푸시 차단: force push/삭제 금지, 병합 전략은 팀 표준(스쿼시/리베이스) 준수
- 코드 오너(Code Owners): 핵심 경로에 코드오너 지정으로 승인 필수화

GitHub 설정 경로: Settings → Branches → Branch protection rules

## 🔄 원격 저장소 작업

### 동기화
```bash
# 가져오기
git fetch origin                        # 변경사항 가져오기 (병합 안함)
git pull origin main                    # 가져오기 + 병합
git pull --rebase origin main          # 리베이스로 가져오기

# 보내기
git push origin feature/branch-name     # 브랜치 푸시
git push -u origin feature/new-branch   # 새 브랜치 푸시 + 업스트림 설정
git push --force-with-lease origin main # 안전한 강제 푸시
```

### 강제 푸시 정책

- 기본 금지. 불가피할 경우에만 `--force-with-lease` 사용하고 사전 공지 필수
- 보호 브랜치에는 허용하지 않음

### 리모트 관리
```bash
# 리모트 확인
git remote -v                           # 리모트 저장소 목록

# 리모트 추가/변경
git remote add upstream https://github.com/original/repo.git
git remote set-url origin https://new-url.git

# 업스트림 동기화
git fetch upstream
git merge upstream/main
```

## 🔏 커밋 서명(Signing) 설정(GPG/SSH)

### SSH 서명(간편)
```bash
git config --global gpg.format ssh
git config --global user.signingkey "$(ssh-add -L | head -1)"
git config --global commit.gpgsign true
```

### GPG 서명(전통)
```bash
gpg --full-generate-key
gpg --list-secret-keys --keyid-format=long
git config --global user.signingkey <KEYID>
git config --global commit.gpgsign true
```

> 리포지토리 “Verified” 배지를 위해 GitHub 계정에 공개 키 등록 필요

## 🔀 병합과 리베이스

### 병합 (Merge)
```bash
# 일반 병합
git checkout main
git merge feature/user-auth

# 병합 커밋 생성하지 않기
git merge --squash feature/user-auth
git commit -m "feat: add user authentication"

# 병합 취소
git merge --abort
```

### 리베이스 (Rebase)
```bash
# 브랜치 리베이스
git checkout feature/user-auth
git rebase main

# 인터랙티브 리베이스 (커밋 정리)
git rebase -i HEAD~3                    # 최근 3개 커밋 정리
git rebase -i main                      # main 브랜치 기준 정리

# 리베이스 중 충돌 해결
git add .                               # 충돌 해결 후
git rebase --continue                   # 리베이스 계속
git rebase --abort                      # 리베이스 취소
```

## 🏷️ 태그와 릴리스

### 태그 생성
```bash
# 태그 생성
git tag v1.0.0                          # 라이트웨이트 태그
git tag -a v1.0.0 -m "Release version 1.0.0"  # 주석 태그

# 태그 푸시
git push origin v1.0.0                  # 특정 태그 푸시
git push origin --tags                  # 모든 태그 푸시

# 태그 삭제
git tag -d v1.0.0                       # 로컬 태그 삭제
git push origin --delete v1.0.0         # 리모트 태그 삭제
```

### 시맨틱 버전 관리
```bash
# 버전 타입
v1.0.0                                  # 초기 릴리스
v1.0.1                                  # 패치 (버그 수정)
v1.1.0                                  # 마이너 (기능 추가)
v2.0.0                                  # 메이저 (Breaking Change)
```

## 🔍 히스토리와 검색

### 로그 확인
```bash
# 기본 로그
git log                                 # 전체 로그
git log --oneline                       # 한 줄로 요약
git log --graph --oneline --all         # 그래프 형태

# 필터링
git log --author="John Doe"             # 작성자별
git log --since="2023-01-01"            # 날짜별
git log --grep="fix"                    # 커밋 메시지 검색
git log -p src/auth.py                  # 특정 파일의 변경사항

# 통계
git log --stat                          # 변경 파일 통계
git shortlog -sn                        # 기여자별 커밋 수
```

### 검색과 비교
```bash
# 변경사항 검색
git grep "TODO"                         # 코드에서 문자열 검색
git log -S "function_name"              # 특정 코드 추가/삭제 검색

# 브랜치 비교
git diff main..feature/branch           # 브랜치 간 차이
git diff --stat main..feature/branch    # 변경 파일 요약

# 특정 커밋 확인
git show commit_hash                    # 커밋 상세 정보
git show --stat commit_hash             # 커밋 파일 변경 통계
```

## 🪝 pre-commit 훅(추천)

로컬에서 빠른 검사로 PR 실패를 예방합니다.

### 예시(.pre-commit-config.yaml)
```yaml
repos:
  - repo: https://github.com/psf/black
    rev: 24.8.0
    hooks:
      - id: black
  - repo: https://github.com/charliermarsh/ruff-pre-commit
    rev: v0.6.8
    hooks:
      - id: ruff
  - repo: https://github.com/pre-commit/mirrors-prettier
    rev: v4.0.0
    hooks:
      - id: prettier
  - repo: https://github.com/pre-commit/mirrors-eslint
    rev: v9.11.1
    hooks:
      - id: eslint
```

```bash
pipx install pre-commit || pip install pre-commit
pre-commit install
```

## 🧾 Pull Request 템플릿 요건(요약)

- 공통 체크리스트는 @.claude/memory/shared_checklists.md를 참고하여 PR 본문/템플릿에 포함
- 태그 매핑(@REQ/@TASK/@TEST), 테스트/커버리지 리포트, 문서/보안 영향, 가정/대안 비교 항목을 반드시 확인

## 🛠️ 유용한 Git 설정

### 전역 설정
```bash
# 사용자 정보
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# 기본 설정
git config --global init.defaultBranch main
git config --global pull.rebase true
git config --global push.default simple

# 에디터 설정
git config --global core.editor "code --wait"  # VS Code
git config --global core.editor "vim"          # Vim
```

### 별칭 설정
```bash
# 유용한 별칭들
git config --global alias.st status
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.cm commit
git config --global alias.lg "log --oneline --graph --all"
git config --global alias.unstage "reset HEAD --"
```

## 🚨 문제 해결

### 일반적인 문제들
```bash
# 실수로 커밋한 경우
git reset --soft HEAD~1                 # 커밋 취소, 변경사항 유지
git reset --hard HEAD~1                 # 커밋 취소, 변경사항 삭제

# 파일 되돌리기
git restore file.txt                    # 작업 디렉토리 변경사항 취소
git restore --staged file.txt           # 스테이지 취소

# 커밋 수정
git commit --amend -m "new message"     # 마지막 커밋 메시지 수정
git commit --amend --no-edit            # 마지막 커밋에 변경사항 추가

# 충돌 해결
git status                              # 충돌 파일 확인
# 파일 수정 후
git add .                               # 해결된 파일 스테이징
git commit                              # 병합 커밋 생성
```

### 응급 상황
```bash
# 작업 임시 저장
git stash                               # 현재 작업 임시 저장
git stash pop                           # 임시 저장된 작업 복원
git stash list                          # 저장된 stash 목록
git stash apply stash@{0}               # 특정 stash 적용

# 실수한 커밋 찾기
git reflog                              # 모든 커밋 기록
git reset --hard commit_hash            # 특정 커밋으로 되돌리기

# 파일 복구
git checkout commit_hash -- file.txt    # 특정 커밋에서 파일 복구
```

## 🔄 MoAI-ADK 표준 워크플로우

### 새 기능 개발
```bash
# 1. 최신 코드 가져오기
git checkout main
git pull origin main

# 2. 기능 브랜치 생성
git checkout -b feature/SPEC-001-user-auth

# 3. 개발 작업 (TDD 사이클)
git add .
git commit -m "test: add user authentication tests"
git commit -m "feat: implement user authentication"

# 4. 정기적으로 main과 동기화
git fetch origin main
git rebase origin/main

# 5. 푸시 및 PR 생성
git push -u origin feature/SPEC-001-user-auth
```

### 버그 수정
```bash
# 1. 핫픽스 브랜치 생성
git checkout main
git checkout -b hotfix/login-timeout

# 2. 수정 및 테스트
git add .
git commit -m "fix: resolve login timeout issue"

# 3. 즉시 배포
git checkout main
git merge hotfix/login-timeout
git tag v1.0.1
git push origin main --tags
```

---

**참고**: 이 워크플로우는 MoAI Constitution의 Versioning 원칙(`@.moai/memory/constitution.md` Article V)과 16-Core TAG 시스템(`@.claude/memory/project_guidelines.md`)을 따릅니다.

> 변경은 ‘작고 안전하게(Small & Safe)’. 필요시 후속 PR로 단계적 진행.
