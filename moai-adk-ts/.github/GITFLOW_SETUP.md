# 🌊 MoAI-ADK GitFlow 설정 가이드

완전 자동화된 GitFlow + 릴리즈 파이프라인 설정 방법

---

## 📋 목차

1. [개요](#개요)
2. [브랜치 보호 규칙](#브랜치-보호-규칙)
3. [Secrets 설정](#secrets-설정)
4. [워크플로우 동작 방식](#워크플로우-동작-방식)
5. [사용 방법](#사용-방법)
6. [트러블슈팅](#트러블슈팅)

---

## 🎯 개요

### 전체 워크플로우

```
1. feature/SPEC-XXX 개발
   ↓
2. develop 브랜치로 PR & 머지
   ↓
3. /awesome:release-new patch (develop에서)
   ↓ (자동 트리거)
4. develop → main PR 자동 생성
   ↓
5. PR 리뷰 & 승인
   ↓
6. PR 머지 (main으로)
   ↓ (자동 트리거)
7. 자동 릴리즈
   - npm publish
   - GitHub Release
   - Git 태그
```

### 주요 특징

- ✅ **main 직접 push 차단** - 브랜치 보호 규칙
- ✅ **자동 PR 생성** - release: 커밋 시 develop → main PR
- ✅ **자동 릴리즈** - main PR 머지 시 npm 배포 + GitHub Release
- ✅ **품질 게이트** - 모든 단계에서 테스트/린트 자동 실행
- ✅ **롤백 용이** - main의 모든 커밋이 릴리즈

---

## 🛡️ 브랜치 보호 규칙

### 1. GitHub 설정 페이지 이동

```
https://github.com/modu-ai/moai-adk/settings/branches
```

### 2. main 브랜치 보호 규칙

**Add branch protection rule** 클릭 후 다음 설정:

#### 기본 설정
- **Branch name pattern**: `main`

#### 필수 설정 ✅

**Require a pull request before merging**
- ✅ Require approvals: **1** (팀에서 결정)
- ✅ Dismiss stale pull request approvals when new commits are pushed
- ✅ Require review from Code Owners (선택)

**Require status checks to pass before merging**
- ✅ Require branches to be up to date before merging
- **Status checks**:
  - `🗿 MoAI-ADK 파이프라인` (moai-gitflow.yml)
  - 추가 CI 체크 (있는 경우)

**Require conversation resolution before merging**
- ✅ Require conversation resolution

**Do not allow bypassing the above settings**
- ✅ Do not allow bypassing the above settings
- ⚠️ **예외**: 리포지토리 관리자는 긴급 시 우회 가능 (신중하게)

#### 선택 설정 (권장)

**Require linear history**
- ✅ Require linear history (squash merge 강제)

**Require deployments to succeed before merging**
- 필요 시 설정

### 3. develop 브랜치 보호 규칙 (선택)

develop 브랜치도 보호하려면:

- **Branch name pattern**: `develop`
- **Require a pull request before merging**: 1 approval
- **Require status checks**: 동일하게 설정

---

## 🔐 Secrets 설정

### 1. NPM_TOKEN 설정 (필수)

**npm 토큰 생성**:
```bash
npm login
npm token create --read-write
```

**GitHub Secrets 추가**:
1. `https://github.com/modu-ai/moai-adk/settings/secrets/actions` 이동
2. **New repository secret** 클릭
3. Name: `NPM_TOKEN`
4. Value: (위에서 생성한 토큰 붙여넣기)
5. **Add secret** 클릭

### 2. GITHUB_TOKEN (자동 제공)

- GitHub Actions가 자동으로 제공
- 별도 설정 불필요
- GitHub Release 생성, PR 생성 등에 사용

---

## 🔄 워크플로우 동작 방식

### 워크플로우 1: `moai-gitflow.yml`

**트리거**: feature, develop 브랜치 push/PR

**동작**:
- 다중 언어 환경 자동 감지 (TypeScript, Python, Go, Rust, Java, .NET)
- 언어별 테스트 실행
- TAG 시스템 검증
- TRUST 5원칙 검증

### 워크플로우 2: `auto-release-pr.yml`

**트리거**: develop 브랜치에 `release:` 커밋 push

**동작**:
1. VERSION 파일에서 버전 읽기
2. develop → main PR 존재 여부 확인
3. 없으면 PR 자동 생성
4. CHANGELOG 내용 포함

**예시**:
```bash
# develop 브랜치에서
git commit -m "release: v0.2.13 - 새 기능 추가"
git push origin develop

# → 자동으로 develop → main PR 생성됨
```

### 워크플로우 3: `release.yml`

**트리거**: develop → main PR 머지

**동작**:
1. **품질 검증**
   - 테스트 실행
   - 린트 검사
   - 빌드
   - 타입 체크

2. **버전 정보 추출**
   - VERSION 파일 읽기
   - CHANGELOG 해당 섹션 추출

3. **Git 태그 생성**
   - Annotated tag 생성
   - origin에 push

4. **npm 배포**
   - `npm publish --access public`
   - NPM_TOKEN 사용

5. **GitHub Release 생성**
   - 태그 기반 Release 생성
   - CHANGELOG 내용 포함

6. **알림**
   - 성공 메시지 출력
   - 릴리즈 링크 표시

---

## 📖 사용 방법

### 시나리오 1: 새 기능 개발

```bash
# 1. feature 브랜치 생성
git checkout develop
git pull origin develop
git checkout -b feature/SPEC-XXX

# 2. 개발 (TDD)
# ... 코딩 ...

# 3. develop으로 PR 생성 및 머지
git push origin feature/SPEC-XXX
# GitHub에서 PR 생성 → 리뷰 → 머지

# 4. develop에서 릴리즈 준비
git checkout develop
git pull origin develop
/awesome:release-new patch  # Claude Code에서

# 5. 자동으로 develop → main PR 생성됨!
# → GitHub에서 PR 확인 및 리뷰

# 6. PR 머지 → 자동 릴리즈!
```

### 시나리오 2: 핫픽스

```bash
# 1. main에서 hotfix 브랜치 생성
git checkout main
git pull origin main
git checkout -b hotfix/critical-bug

# 2. 버그 수정
# ... 수정 ...

# 3. main으로 직접 PR
git push origin hotfix/critical-bug
# GitHub에서 PR 생성 → 리뷰 → 머지

# 4. main PR 머지 → 자동 릴리즈!

# 5. main의 변경사항을 develop에 백포트
git checkout develop
git pull origin develop
git merge main
git push origin develop
```

### 시나리오 3: 수동 릴리즈 (비상시)

브랜치 보호 규칙을 우회해야 하는 긴급 상황:

```bash
# 1. 리포지토리 관리자 권한 필요
# 2. GitHub Settings → Branches → main → Edit
# 3. "Allow specified actors to bypass required pull requests" 체크
# 4. 관리자 계정 추가
# 5. main에 직접 push (매우 신중하게!)
```

⚠️ **주의**: 수동 릴리즈는 최후의 수단입니다!

---

## 🔍 트러블슈팅

### 문제 1: npm 배포 실패 (401 Unauthorized)

**원인**: NPM_TOKEN이 만료되었거나 잘못 설정됨

**해결**:
```bash
# 새 토큰 생성
npm token create --read-write

# GitHub Secrets 업데이트
# https://github.com/modu-ai/moai-adk/settings/secrets/actions
```

### 문제 2: GitHub Release 생성 실패 (403 Forbidden)

**원인**: GITHUB_TOKEN 권한 부족

**해결**:
1. `https://github.com/modu-ai/moai-adk/settings/actions` 이동
2. **Workflow permissions** 섹션
3. **Read and write permissions** 선택
4. **Save** 클릭

### 문제 3: 자동 PR이 생성되지 않음

**원인**: 커밋 메시지가 `release:`로 시작하지 않음

**해결**:
```bash
# 올바른 커밋 메시지 형식
git commit -m "release: v0.2.13 - 설명"

# 잘못된 예시
git commit -m "Release v0.2.13"  # ❌ 대문자 R
git commit -m "v0.2.13 release"  # ❌ 순서 다름
```

### 문제 4: 품질 검증 실패로 릴리즈 중단

**원인**: 테스트/린트/빌드 실패

**해결**:
```bash
# develop 브랜치에서 로컬로 검증
npm test
npm run lint
npm run build
npm run type-check

# 문제 수정 후 다시 커밋
```

### 문제 5: 태그가 이미 존재함

**원인**: VERSION 파일 버전이 이미 릴리즈된 버전

**해결**:
```bash
# VERSION 파일 업데이트
echo "0.2.14" > VERSION

# package.json도 업데이트
npm version 0.2.14 --no-git-tag-version

# 다시 릴리즈 커밋
```

---

## 🎓 추가 학습 자료

- [GitHub Branch Protection](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
- [GitHub Actions Secrets](https://docs.github.com/en/actions/security-guides/using-secrets-in-github-actions)
- [npm Token Management](https://docs.npmjs.com/about-access-tokens)
- [GitFlow Workflow](https://nvie.com/posts/a-successful-git-branching-model/)

---

## ✅ 체크리스트

릴리즈 자동화 설정 완료 확인:

- [ ] main 브랜치 보호 규칙 설정
- [ ] NPM_TOKEN Secrets 추가
- [ ] GITHUB_TOKEN 권한 설정 (Read and write)
- [ ] `.github/workflows/` 파일 3개 확인
  - [ ] `moai-gitflow.yml`
  - [ ] `auto-release-pr.yml`
  - [ ] `release.yml`
- [ ] 테스트 릴리즈 실행 (develop → main)
- [ ] npm에 패키지 배포 확인
- [ ] GitHub Release 생성 확인

---

**다음**: 첫 번째 자동 릴리즈를 테스트해보세요! 🚀
