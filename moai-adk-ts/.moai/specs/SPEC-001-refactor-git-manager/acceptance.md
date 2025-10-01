# SPEC-001: 수락 기준

## @TAG BLOCK

```text
# @TEST:REFACTOR-001 | Chain: @SPEC:REFACTOR-001 -> @SPEC:REFACTOR-001 -> @CODE:REFACTOR-001 -> @TEST:REFACTOR-001
# Related: @CODE:REFACTOR-001
```

## Given-When-Then 테스트 시나리오

### 시나리오 1: GitBranchManager 분리 검증

#### 1.1 브랜치 생성 기본 동작
```gherkin
Given git-branch-manager.ts 파일이 생성되어 있고
  And GitBranchManager 클래스가 정의되어 있고
  And 파일 크기가 300 LOC 이하일 때
When createBranch('feature/test-branch') 메서드를 호출하면
Then 브랜치가 성공적으로 생성되고
  And 브랜치명 검증이 통과되고
  And 초기 커밋이 자동으로 생성된다
```

**검증 방법**:
- 파일 생성 확인: `ls -la src/core/git/git-branch-manager.ts`
- LOC 측정: `wc -l src/core/git/git-branch-manager.ts`
- 단위 테스트: `npm test git-branch-manager.test.ts`

---

#### 1.2 브랜치명 검증 강화
```gherkin
Given 악의적인 브랜치명 입력이 있을 때
  (예: "../../../etc/passwd", "rm -rf /", "'; DROP TABLE users;--")
When createBranch(maliciousBranchName) 메서드를 호출하면
Then InputValidator가 입력을 차단하고
  And GitNamingRules 검증이 실패하고
  And 적절한 에러 메시지를 반환한다
```

**검증 방법**:
- 보안 테스트: `npm test -- --grep "malicious branch name"`
- InputValidator 통합 확인
- GitNamingRules 준수 검증

---

#### 1.3 원격 저장소 연결
```gherkin
Given 유효한 GitHub 저장소 URL이 주어지고
When linkRemoteRepository('https://github.com/user/repo.git') 메서드를 호출하면
Then 원격 저장소가 'origin'으로 등록되고
  And git remote -v 명령으로 확인 가능하다
```

**검증 방법**:
- URL 검증 테스트
- 원격 저장소 추가 확인
- 기존 원격 제거 로직 검증

---

### 시나리오 2: GitCommitManager 분리 검증

#### 2.1 커밋 생성 기본 동작
```gherkin
Given git-commit-manager.ts 파일이 생성되어 있고
  And GitCommitManager 클래스가 정의되어 있고
  And 파일 크기가 300 LOC 이하일 때
When commitChanges('feat: add new feature', ['file1.ts', 'file2.ts']) 메서드를 호출하면
Then 지정된 파일들이 스테이징되고
  And 커밋 템플릿이 적용되고
  And 커밋이 성공적으로 생성된다
```

**검증 방법**:
- 파일 생성 확인: `ls -la src/core/git/git-commit-manager.ts`
- LOC 측정: `wc -l src/core/git/git-commit-manager.ts`
- 단위 테스트: `npm test git-commit-manager.test.ts`

---

#### 2.2 파일 검증
```gherkin
Given 존재하지 않는 파일 경로가 주어지고
When commitChanges('feat: test', ['non-existent-file.ts']) 메서드를 호출하면
Then 파일 검증이 실패하고
  And "File not found: non-existent-file.ts" 에러가 발생한다
```

**검증 방법**:
- 파일 존재 여부 검증 테스트
- 에러 메시지 확인

---

#### 2.3 커밋 템플릿 적용
```gherkin
Given GitConfig에 commitMessageTemplate이 설정되어 있고
When commitChanges('add feature') 메서드를 호출하면
Then 커밋 메시지가 템플릿 형식으로 변환되고
  And Git 이력에 올바른 형식으로 저장된다
```

**검증 방법**:
- 템플릿 적용 테스트
- `git log -1 --pretty=%B` 명령으로 커밋 메시지 확인

---

#### 2.4 체크포인트 생성
```gherkin
Given 변경사항이 스테이징되어 있고
When createCheckpoint('checkpoint before refactoring') 메서드를 호출하면
Then 체크포인트 형식의 커밋이 생성되고
  And 커밋 해시가 반환된다
```

**검증 방법**:
- 체크포인트 커밋 생성 테스트
- 커밋 메시지 형식 검증

---

#### 2.5 푸시 동작
```gherkin
Given 로컬 브랜치에 커밋이 있고
When pushChanges('feature/test-branch') 메서드를 호출하면
Then 변경사항이 원격 저장소로 푸시되고
  And --set-upstream 옵션이 자동으로 설정된다
```

**검증 방법**:
- 푸시 동작 테스트
- 원격 브랜치 확인

---

### 시나리오 3: GitPRManager 분리 검증

#### 3.1 PR 생성 (Team 모드)
```gherkin
Given GitConfig.mode가 'team'이고
  And GitHubIntegration이 초기화되어 있고
When createPullRequest({ title: 'PR Title', body: 'PR Body' }) 메서드를 호출하면
Then GitHub PR이 생성되고
  And PR URL이 반환된다
```

**검증 방법**:
- Team 모드 테스트
- GitHub API 모킹
- PR 생성 로직 검증

---

#### 3.2 Personal 모드 제한
```gherkin
Given GitConfig.mode가 'personal'이고
When createPullRequest({ title: 'PR Title', body: 'PR Body' }) 메서드를 호출하면
Then "Pull request creation is only available in team mode" 에러가 발생한다
```

**검증 방법**:
- Personal 모드 에러 테스트
- 모드 검증 로직 확인

---

#### 3.3 GitHub CLI 통합
```gherkin
Given GitHub CLI(gh)가 설치되어 있고
When isGitHubCliAvailable() 메서드를 호출하면
Then true를 반환한다
```

**검증 방법**:
- GitHub CLI 설치 확인
- 가용성 테스트

---

### 시나리오 4: GitManager 위임 패턴 검증

#### 4.1 기존 API 호환성
```gherkin
Given GitManager가 리팩토링되어 있고
When 기존 테스트 스위트(git-manager.test.ts)를 실행하면
Then 모든 테스트가 통과하고
  And 기존 API 시그니처가 유지되고
  And 동작이 변경되지 않는다
```

**검증 방법**:
- 전체 테스트 실행: `npm test git-manager.test.ts`
- 테스트 커버리지 측정: `npm run test:coverage`

---

#### 4.2 매니저 인스턴스 생성
```gherkin
Given GitManager 생성자가 호출되고
When new GitManager(config, workingDir)가 실행되면
Then branchManager 인스턴스가 생성되고
  And commitManager 인스턴스가 생성되고
  And prManager 인스턴스가 생성되고
  And lockManager 인스턴스가 생성된다
```

**검증 방법**:
- 생성자 테스트
- 인스턴스 존재 확인

---

#### 4.3 위임 메서드 동작
```gherkin
Given GitManager.createBranch('test') 메서드가 호출되고
When 내부적으로 branchManager.createBranch('test')가 호출되면
Then 브랜치가 정상적으로 생성되고
  And GitManager의 에러 처리가 적용된다
```

**검증 방법**:
- 위임 로직 테스트
- 에러 처리 검증

---

### 시나리오 5: Lock 통합 검증

#### 5.1 Lock을 사용한 커밋
```gherkin
Given 동시에 2개의 커밋 요청이 발생하고
When commitWithLock('message1')와 commitWithLock('message2')가 동시 호출되면
Then 첫 번째 커밋이 Lock을 획득하고
  And 두 번째 커밋이 대기하고
  And 순차적으로 커밋이 완료된다
```

**검증 방법**:
- 동시성 테스트
- Lock 타임아웃 검증

---

#### 5.2 Lock을 사용한 브랜치 생성
```gherkin
Given 동시에 2개의 브랜치 생성 요청이 발생하고
When createBranchWithLock('branch1')와 createBranchWithLock('branch2')가 동시 호출되면
Then 순차적으로 브랜치가 생성되고
  And 브랜치 충돌이 발생하지 않는다
```

**검증 방법**:
- 브랜치 생성 동시성 테스트
- Lock 상태 조회

---

### 시나리오 6: 통합 워크플로우 검증

#### 6.1 Personal 모드 전체 워크플로우
```gherkin
Given GitConfig.mode가 'personal'이고
When 다음 작업을 순차적으로 수행하면
  1. initializeRepository(projectPath)
  2. createBranch('feature/test')
  3. commitChanges('feat: test feature')
  4. pushChanges()
Then 모든 작업이 성공하고
  And 원격 저장소에 변경사항이 반영된다
```

**검증 방법**:
- 전체 워크플로우 통합 테스트
- 각 단계별 성공 확인

---

#### 6.2 Team 모드 전체 워크플로우
```gherkin
Given GitConfig.mode가 'team'이고
When 다음 작업을 순차적으로 수행하면
  1. createRepository({ name: 'test-repo', private: true })
  2. createBranch('feature/test')
  3. commitChanges('feat: test feature')
  4. pushChanges()
  5. createPullRequest({ title: 'Test PR', body: 'Test Body' })
Then 모든 작업이 성공하고
  And GitHub PR이 생성되고
  And PR URL이 반환된다
```

**검증 방법**:
- Team 모드 전체 워크플로우 테스트
- GitHub API 모킹

---

## 품질 게이트 기준

### 1. LOC (Lines of Code)

| 파일 | 목표 LOC | 허용 범위 |
|------|----------|-----------|
| git-manager.ts | 150 LOC | ≤ 180 LOC |
| git-branch-manager.ts | 200 LOC | ≤ 250 LOC |
| git-commit-manager.ts | 200 LOC | ≤ 250 LOC |
| git-pr-manager.ts | 150 LOC | ≤ 180 LOC |

**측정 방법**:
```bash
wc -l src/core/git/git-*.ts
```

**실패 조건**:
- 어떤 파일이라도 300 LOC를 초과하면 실패

---

### 2. 함수 복잡도

| 항목 | 목표 | 허용 범위 |
|------|------|-----------|
| 함수당 LOC | ≤ 50 | ≤ 60 |
| 매개변수 개수 | ≤ 5 | ≤ 6 |
| 복잡도(Cyclomatic) | ≤ 10 | ≤ 12 |

**측정 방법**:
- ESLint complexity 규칙
- TypeScript strict 모드

**실패 조건**:
- 복잡도 15 이상인 함수가 1개라도 존재하면 실패

---

### 3. 테스트 커버리지

| 항목 | 목표 | 최소 기준 |
|------|------|-----------|
| Statement | 90% | 85% |
| Branch | 85% | 80% |
| Function | 90% | 85% |
| Line | 90% | 85% |

**측정 방법**:
```bash
npm run test:coverage
```

**실패 조건**:
- 어떤 항목이라도 최소 기준 미달 시 실패

---

### 4. 의존성 분석

| 항목 | 목표 |
|------|------|
| 순환 의존성 | 0개 |
| 미사용 import | 0개 |
| 타입 에러 | 0개 |

**측정 방법**:
```bash
# 순환 의존성 검사
npx madge --circular src/core/git/

# 미사용 import 검사
npx eslint src/core/git/ --fix

# 타입 체크
npx tsc --noEmit
```

**실패 조건**:
- 순환 의존성이 1개라도 존재하면 실패
- 타입 에러가 1개라도 존재하면 실패

---

### 5. 성능 벤치마크

| 항목 | 기준 | 허용 범위 |
|------|------|-----------|
| 리팩토링 전 대비 속도 | 100% | 95% ~ 105% |
| 메모리 사용량 | 100% | 95% ~ 110% |

**측정 방법**:
- 벤치마크 테스트 작성
- 실행 시간 비교

**실패 조건**:
- 성능이 기존 대비 20% 이상 저하되면 실패

---

## 검증 방법 및 도구

### 자동 검증 스크립트

```bash
#!/bin/bash
# .moai/scripts/verify-spec-001.sh

echo "🔍 SPEC-001 검증 시작..."

# 1. LOC 측정
echo "\n📊 LOC 측정..."
wc -l src/core/git/git-manager.ts
wc -l src/core/git/git-branch-manager.ts
wc -l src/core/git/git-commit-manager.ts
wc -l src/core/git/git-pr-manager.ts

# 2. 테스트 실행
echo "\n🧪 테스트 실행..."
npm test src/core/git/__tests__/

# 3. 테스트 커버리지
echo "\n📈 커버리지 측정..."
npm run test:coverage -- src/core/git/

# 4. 순환 의존성 검사
echo "\n🔗 순환 의존성 검사..."
npx madge --circular src/core/git/

# 5. 타입 체크
echo "\n🔍 타입 체크..."
npx tsc --noEmit

# 6. 린터 실행
echo "\n✨ 린터 실행..."
npx eslint src/core/git/ --max-warnings 0

echo "\n✅ SPEC-001 검증 완료!"
```

---

### 수동 검증 체크리스트

#### Phase 1: 파일 생성 확인
- [ ] `src/core/git/git-branch-manager.ts` 파일 존재
- [ ] `src/core/git/git-commit-manager.ts` 파일 존재
- [ ] `src/core/git/git-pr-manager.ts` 파일 존재
- [ ] 각 파일에 @TAG BLOCK 포함
- [ ] 각 파일에 JSDoc 주석 포함

#### Phase 2: LOC 검증
- [ ] git-manager.ts ≤ 180 LOC
- [ ] git-branch-manager.ts ≤ 250 LOC
- [ ] git-commit-manager.ts ≤ 250 LOC
- [ ] git-pr-manager.ts ≤ 180 LOC

#### Phase 3: 테스트 검증
- [ ] 기존 git-manager.test.ts 전체 통과
- [ ] git-branch-manager.test.ts 작성 및 통과
- [ ] git-commit-manager.test.ts 작성 및 통과
- [ ] git-pr-manager.test.ts 작성 및 통과
- [ ] 통합 테스트 통과

#### Phase 4: 품질 검증
- [ ] ESLint 규칙 준수 (0 warnings)
- [ ] TypeScript strict 모드 통과
- [ ] 순환 의존성 없음
- [ ] 테스트 커버리지 ≥ 85%

#### Phase 5: 기능 검증
- [ ] Personal 모드 워크플로우 동작
- [ ] Team 모드 워크플로우 동작
- [ ] Lock 통합 동작
- [ ] GitHub 연동 동작

---

## 완료 조건 (Definition of Done)

### 필수 조건 (Must Have)
- ✅ 모든 파일이 300 LOC 이하
- ✅ 모든 기존 테스트 통과
- ✅ 테스트 커버리지 85% 이상
- ✅ 순환 의존성 없음
- ✅ TypeScript strict 모드 통과
- ✅ ESLint 규칙 준수 (0 warnings)
- ✅ 기존 API 100% 호환

### 선택 조건 (Should Have)
- ✅ 성능 저하 5% 이내
- ✅ 메모리 사용량 10% 이내
- ✅ 코드 리뷰 승인
- ✅ 문서 동기화 완료

### 최종 확인 (Nice to Have)
- ✅ 벤치마크 테스트 통과
- ✅ 성능 프로파일링 완료
- ✅ 아키텍처 문서 업데이트

---

## 다음 단계

리팩토링 완료 및 검증 통과 후:

1. **문서 동기화**: `/moai:3-sync` 실행하여 TAG 검증
2. **Git 작업**: 브랜치 생성 및 커밋
3. **PR 생성**: Draft PR 생성 (Team 모드)
4. **팀 리뷰**: 코드 리뷰 요청
5. **머지**: 승인 후 develop 브랜치로 머지

---

## 참고 자료

- TRUST 원칙: `.moai/memory/development-guide.md`
- 테스트 전략: Vitest 공식 문서
- TDD 사이클: Red-Green-Refactor
- Git Manager 원본: `src/core/git/git-manager.ts`
