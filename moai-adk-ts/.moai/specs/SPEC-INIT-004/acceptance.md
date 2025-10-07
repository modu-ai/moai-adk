# SPEC-INIT-004 인수 기준 (Acceptance Criteria)

## 개요

SPEC-INIT-004의 구현 완료를 검증하기 위한 인수 기준을 정의합니다.

**검증 방법**: Given-When-Then 시나리오 기반 테스트

**품질 게이트**: TRUST 5원칙 준수

---

## 시나리오 1: .git 없음 + 자동 초기화

### Given (전제 조건)
- 프로젝트 디렉토리에 `.git` 폴더가 없는 상태
- Git이 설치되어 있음 (git --version 성공)

### When (실행)
```bash
cd /tmp/test-project
moai init my-project
```

### Then (기대 결과)
1. 자동으로 `git init`이 실행됨
2. "✅ Git 저장소 초기화 완료" 메시지 표시
3. `.git` 폴더 생성 확인
4. GitHub 사용 여부 질문 (Personal 모드는 선택, Team 모드는 필수)

### 검증 방법
```bash
# .git 폴더 존재 확인
test -d .git && echo "✅ PASS" || echo "❌ FAIL"

# Git 저장소 검증
git rev-parse --is-inside-work-tree
# 출력: true
```

---

## 시나리오 2: .git 있음 + GitHub 자동 감지

### Given (전제 조건)
- `.git` 폴더가 존재하는 상태
- GitHub remote가 설정되어 있음
  ```bash
  git remote add origin https://github.com/user/repo.git
  ```
- 커밋이 10개 이상 존재

### When (실행)
```bash
moai init .
```

### Then (기대 결과)
1. 기존 저장소 정보 수집 및 표시
   ```
   기존 Git 저장소 발견:
     • 커밋: 15개
     • 브랜치: main
     • 원격: origin
   ```
2. GitHub 저장소 자동 감지
   ```
   ✅ GitHub 저장소 자동 감지: https://github.com/user/repo.git
   ```
3. `.moai/config.json`에 자동 저장
4. "기존 저장소를 유지하시겠습니까? (Y/n)" 질문
5. 기본값 선택 시 (Enter) 기존 저장소 유지

### 검증 방법
```bash
# GitHub URL 자동 저장 확인
cat .moai/config.json | jq -r '.git_strategy.github_repo'
# 출력: https://github.com/user/repo.git

# 기존 커밋 보존 확인
git log --oneline | wc -l
# 출력: 15 (변경 없음)
```

---

## 시나리오 3: 언어 선택 ko, en만

### Given (전제 조건)
- moai init 실행 중 언어 선택 단계

### When (실행)
- 언어 선택 프롬프트 표시

### Then (기대 결과)
1. "한국어 (ko)"와 "English (en)" 2개만 선택 가능
2. "日本語 (ja)", "中文 (zh)"는 선택지에 없음
3. 선택된 언어가 `.moai/config.json`의 `locale` 필드에 저장

### 검증 방법
```bash
# locale 필드 검증
cat .moai/config.json | jq -r '.locale'
# 출력: "ko" 또는 "en" (ja, zh는 불가)
```

---

## 시나리오 4: .git 있음 + GitHub 없음 (Personal 모드)

### Given (전제 조건)
- `.git` 폴더가 존재하지만 GitHub remote가 없는 상태
  ```bash
  git remote -v  # 출력 없음
  ```
- Personal 모드 선택

### When (실행)
```bash
moai init . --mode personal
```

### Then (기대 결과)
1. 기존 저장소 정보 수집 (GitHub 없음 표시)
   ```
   기존 Git 저장소 발견:
     • 커밋: 5개
     • 브랜치: main
     • 원격: 없음
   ```
2. "GitHub 저장소를 사용하시겠습니까? (y/N)" 질문
3. 기본값 선택 시 (Enter) GitHub 없이 진행
4. 초기화 완료

### 검증 방법
```bash
# config.json에 github_repo 없음 확인
cat .moai/config.json | jq -r '.git_strategy.github_repo // "null"'
# 출력: "null" (없음)

# Personal 모드 확인
cat .moai/config.json | jq -r '.mode'
# 출력: "personal"
```

---

## 시나리오 5: .git 없음 + Team 모드 (GitHub 필수)

### Given (전제 조건)
- `.git` 폴더가 없는 상태
- Team 모드 선택

### When (실행)
```bash
moai init my-project --mode team
```

### Then (기대 결과)
1. 자동으로 `git init` 실행
2. "Team 모드에서는 GitHub가 필수입니다" 안내
3. "GitHub 저장소 URL을 입력하세요:" 프롬프트
4. URL 유효성 검증 (GitHub 패턴)
   - ✅ 유효: `https://github.com/user/repo`
   - ❌ 무효: `https://gitlab.com/user/repo` (재입력 요청)
5. `.moai/config.json`에 저장

### 검증 방법
```bash
# Team 모드 확인
cat .moai/config.json | jq -r '.mode'
# 출력: "team"

# GitHub URL 저장 확인
cat .moai/config.json | jq -r '.git_strategy.github_repo'
# 출력: https://github.com/user/repo
```

---

## 시나리오 6: 비대화형 모드 (--auto-git, --locale)

### Given (전제 조건)
- CI/CD 환경 (비대화형 모드)
- TTY 없음 (process.stdin.isTTY === false)

### When (실행)
```bash
moai init my-project --auto-git --locale ko --yes
```

### Then (기대 결과)
1. Git 관련 질문 모두 건너뜀
2. `.git` 없으면 자동 초기화
3. `.git` 있으면 유지
4. 언어 선택 프롬프트 건너뜀 (ko 자동 설정)
5. 초기화 완료

### 검증 방법
```bash
# locale 확인
cat my-project/.moai/config.json | jq -r '.locale'
# 출력: "ko"

# Git 초기화 확인
test -d my-project/.git && echo "✅ PASS" || echo "❌ FAIL"
```

---

## 시나리오 7: .git 삭제 + 백업

### Given (전제 조건)
- `.git` 폴더가 존재하는 상태
- 커밋이 20개 이상 존재

### When (실행)
```bash
moai init .
# "기존 저장소를 유지하시겠습니까? (Y/n)" → n 입력
# "정말로 삭제하시겠습니까? (y/N)" → y 입력
```

### Then (기대 결과)
1. `.git-backup-{timestamp}/` 디렉토리에 백업
2. 기존 `.git` 폴더 삭제
3. 새로 `git init` 실행
4. "✅ Git 저장소가 재초기화되었습니다" 메시지

### 검증 방법
```bash
# 백업 존재 확인
ls -d .git-backup-* && echo "✅ PASS" || echo "❌ FAIL"

# 새 Git 저장소 확인
git log --oneline | wc -l
# 출력: 0 (새 저장소)
```

---

## 시나리오 8: GitHub URL 변경 옵션

### Given (전제 조건)
- `.git` 폴더가 존재하고 GitHub remote가 설정된 상태
  ```bash
  git remote add origin https://github.com/old/repo.git
  ```

### When (실행)
```bash
moai init .
# "감지된 GitHub 저장소: https://github.com/old/repo.git"
# "다른 저장소를 사용하시겠습니까? (y/N)" → y 입력
# "GitHub 저장소 URL을 입력하세요:" → https://github.com/new/repo.git 입력
```

### Then (기대 결과)
1. 새 GitHub URL로 변경
2. `.moai/config.json`에 새 URL 저장

### 검증 방법
```bash
# config.json 확인
cat .moai/config.json | jq -r '.git_strategy.github_repo'
# 출력: https://github.com/new/repo.git
```

---

## TRUST 5원칙 검증

### T - Test First

**단위 테스트**:
- ✅ `tests/utils/git-detector.test.ts` (20개 테스트)
- ✅ `tests/cli/prompts/init/definitions.test.ts` (5개 테스트)

**통합 테스트**:
- ✅ `tests/cli/commands/init/git-workflow.test.ts` (15개 테스트)

**테스트 커버리지**:
- ✅ ≥85% (목표)

---

### R - Readable

**코드 품질**:
- ✅ 함수당 ≤50 LOC
- ✅ 파일당 ≤300 LOC
- ✅ 복잡도 ≤10
- ✅ 의도 드러내는 함수명

**예시**:
```typescript
// ✅ 좋은 예
async function detectGitStatus(cwd: string): Promise<GitStatus>

// ❌ 나쁜 예
async function getGit(path: string)
```

---

### U - Unified

**타입 안전성**:
- ✅ TypeScript strict 모드
- ✅ 모든 함수에 타입 정의
- ✅ 런타임 타입 검증 (zod 등)

**일관된 에러 처리**:
```typescript
try {
  await autoInitGit(cwd);
} catch (error) {
  if (error instanceof GitNotInstalledError) {
    logger.error('Git이 설치되지 않았습니다.');
  } else {
    throw error;
  }
}
```

---

### S - Secured

**입력 검증**:
- ✅ GitHub URL 정규식 검증
- ✅ locale 값 검증 (ko|en만)
- ✅ 디렉토리 경로 sanitization

**Git 명령어 주입 방지**:
```typescript
// ✅ 좋은 예
await simpleGit(cwd).init();

// ❌ 나쁜 예
await exec(`cd ${cwd} && git init`);  // 경로 주입 위험
```

---

### T - Trackable

**@TAG 체인**:
- ✅ @SPEC:INIT-004
- ✅ @TEST:INIT-004 (tests/cli/commands/init/git-workflow.test.ts)
- ✅ @CODE:INIT-004 (src/utils/git-detector.ts, src/cli/commands/init/*.ts)

**TAG 무결성**:
```bash
# TAG 체인 검증
rg '@(SPEC|TEST|CODE):INIT-004' -n
# 모든 TAG가 연결되어 있어야 함
```

---

## 성능 기준

**초기화 속도**:
- ✅ 3분 → 1분 (67% 개선)

**응답 시간**:
- ✅ Git 감지: <500ms
- ✅ GitHub URL 추출: <100ms
- ✅ 자동 `git init`: <1s

---

## 접근성 기준

**언어 지원**:
- ✅ 한국어(ko): 모든 메시지 번역
- ✅ 영어(en): 모든 메시지 번역

**에러 메시지**:
- ✅ 명확하고 실행 가능한 안내
- ✅ 예시 포함

---

## 완료 조건 (Definition of Done)

**필수 조건**:
- [ ] 시나리오 1~8 모두 통과
- [ ] TRUST 5원칙 검증 완료
- [ ] 테스트 커버리지 ≥85%
- [ ] @TAG 체인 무결성 확인
- [ ] 문서화 완료
- [ ] PR 리뷰 승인 (Team 모드)

**선택 조건**:
- [ ] 성능 기준 달성 (초기화 1분 이내)
- [ ] 사용자 피드백 긍정적 (≥70%)
- [ ] 크리티컬 버그 0건
