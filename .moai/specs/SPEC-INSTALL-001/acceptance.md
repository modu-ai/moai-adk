# SPEC-INSTALL-001 인수 기준

## 1. Acceptance Criteria

### AC-001: 개발자 이름 프롬프트

#### 시나리오 1: Git user.name 기본값 제안
```gherkin
GIVEN Git 전역 설정에 user.name이 "홍길동"으로 설정되어 있을 때
WHEN "개발자 이름을 입력해주세요" 프롬프트가 표시되면
THEN 기본값으로 "홍길동"이 제안되어야 한다
AND 사용자가 Enter 키를 누르면 "홍길동"을 사용해야 한다
AND .moai/config.json의 developer.name에 "홍길동"이 저장되어야 한다
```

**검증 방법**:
```bash
# 사전 조건
git config --global user.name "홍길동"

# 실행
npm run install

# 검증
cat .moai/config.json | jq '.developer.name'
# 예상 출력: "홍길동"
```

---

#### 시나리오 2: Git user.name 미설정 시 수동 입력
```gherkin
GIVEN Git 전역 설정에 user.name이 없을 때
WHEN "개발자 이름을 입력해주세요" 프롬프트가 표시되면
THEN 기본값으로 빈 문자열이 제안되어야 한다
AND 사용자가 "김철수"를 입력하면 "김철수"가 저장되어야 한다
AND .moai/config.json의 developer.name에 "김철수"가 저장되어야 한다
```

**검증 방법**:
```bash
# 사전 조건
git config --global --unset user.name

# 실행 (수동 입력: "김철수")
npm run install

# 검증
cat .moai/config.json | jq '.developer.name'
# 예상 출력: "김철수"
```

---

#### 시나리오 3: 빈 값 입력 시 에러
```gherkin
GIVEN 개발자 이름 프롬프트가 표시되었을 때
WHEN 사용자가 빈 값(공백만 포함)을 입력하면
THEN "❌ 개발자 이름은 필수입니다" 에러 메시지를 출력해야 한다
AND 프롬프트를 다시 표시해야 한다
AND 유효한 값이 입력될 때까지 반복해야 한다
```

**검증 방법**:
- 단위 테스트: `phase2-developer.test.ts`
- 수동 테스트: 설치 시 공백만 입력 시도

---

### AC-002: Git 필수화

#### 시나리오 1: Git 설치 확인 (정상)
```gherkin
GIVEN Git이 설치되어 있고 버전이 2.30.0일 때
WHEN 설치 프롬프트를 시작하면
THEN "✅ Git 버전: git version 2.30.0" 메시지를 출력해야 한다
AND 다음 Phase(모드 선택)로 진행해야 한다
AND Git 버전 정보를 context에 저장해야 한다
```

**검증 방법**:
```bash
# 사전 조건
git --version
# 출력: git version 2.30.0

# 실행
npm run install

# 예상 출력
# ✅ Git 버전: git version 2.30.0
# 프로젝트 모드를 선택하세요:
```

---

#### 시나리오 2: Git 미설치 시 에러
```gherkin
GIVEN Git이 설치되지 않았을 때
WHEN 설치 프롬프트를 시작하면
THEN "❌ Git이 설치되지 않았습니다" 에러 메시지를 출력해야 한다
AND macOS/Ubuntu/Windows 설치 방법을 안내해야 한다
AND process.exit(1)로 설치를 중단해야 한다
AND 다음 Phase로 진행하지 않아야 한다
```

**검증 방법**:
- 단위 테스트: Git 명령 Mock으로 에러 시뮬레이션
- 수동 테스트: Git을 PATH에서 제거 후 설치 시도

**에러 메시지 예상**:
```
❌ Git이 설치되지 않았습니다.

MoAI-ADK는 Git을 필수로 사용합니다.
다음 방법으로 Git을 설치해주세요:

macOS: brew install git
Ubuntu: sudo apt-get install git
Windows: https://git-scm.com/download/win

설치 후 다시 시도해주세요.
```

---

### AC-003: SPEC Workflow 필수 (Team 모드)

#### 시나리오 1: Team 모드는 SPEC 강제 활성화
```gherkin
GIVEN 사용자가 Team 모드를 선택했을 때
WHEN SPEC Workflow 설정 단계에 도달하면
THEN 프롬프트를 표시하지 않고 enforce_spec: true로 설정해야 한다
AND "ℹ️  Team 모드는 SPEC-First Workflow가 필수입니다" 안내 메시지를 출력해야 한다
AND .moai/config.json의 constitution.enforce_spec이 true여야 한다
```

**검증 방법**:
```bash
# 실행 (Team 모드 선택)
npm run install

# 검증
cat .moai/config.json | jq '.constitution.enforce_spec'
# 예상 출력: true
```

---

#### 시나리오 2: Personal 모드는 SPEC 선택적
```gherkin
GIVEN 사용자가 Personal 모드를 선택했을 때
WHEN "SPEC-First Workflow를 사용할까요?" 프롬프트가 표시되면
THEN 기본값은 true(권장)여야 한다
AND 사용자가 No를 선택하면 enforce_spec: false로 설정해야 한다
AND 사용자가 Yes를 선택하면 enforce_spec: true로 설정해야 한다
```

**검증 방법**:
```bash
# 실행 (Personal 모드 + No 선택)
npm run install

# 검증
cat .moai/config.json | jq '.constitution.enforce_spec'
# 예상 출력: false
```

---

### AC-004: Auto PR 프롬프트 (Team 모드)

#### 시나리오 1: Auto PR 활성화
```gherkin
GIVEN 사용자가 Team 모드를 선택했을 때
WHEN "자동으로 PR을 생성할까요?" 프롬프트가 표시되면
THEN 기본값은 true여야 한다
AND 사용자가 Yes를 선택하면 git_strategy.team.auto_pr이 true여야 한다
AND 다음에 Draft PR 프롬프트가 표시되어야 한다
```

**검증 방법**:
```bash
# 실행 (Team 모드 + Auto PR Yes)
npm run install

# 검증
cat .moai/config.json | jq '.git_strategy.team.auto_pr'
# 예상 출력: true
```

---

#### 시나리오 2: Auto PR 비활성화
```gherkin
GIVEN 사용자가 Team 모드를 선택했을 때
WHEN "자동으로 PR을 생성할까요?" 프롬프트에서 No를 선택하면
THEN git_strategy.team.auto_pr이 false여야 한다
AND Draft PR 프롬프트를 표시하지 않아야 한다
AND git_strategy.team.draft_pr이 false여야 한다 (무의미)
```

**검증 방법**:
```bash
# 실행 (Team 모드 + Auto PR No)
npm run install

# 검증
cat .moai/config.json | jq '.git_strategy.team | {auto_pr, draft_pr}'
# 예상 출력: { "auto_pr": false, "draft_pr": false }
```

---

#### 시나리오 3: Personal 모드는 Auto PR 표시 안 함
```gherkin
GIVEN 사용자가 Personal 모드를 선택했을 때
WHEN Git 전략 설정 단계에 도달하면
THEN Auto PR/Draft PR 프롬프트를 표시하지 않아야 한다
AND git_strategy.team.auto_pr이 false여야 한다
AND git_strategy.team.draft_pr이 false여야 한다
```

**검증 방법**:
- 통합 테스트: Personal 모드 시나리오 확인

---

### AC-005: Draft PR 프롬프트 (Team 모드 + Auto PR 활성화)

#### 시나리오 1: Draft PR 활성화
```gherkin
GIVEN 사용자가 Team 모드를 선택하고 Auto PR을 활성화했을 때
WHEN "PR을 Draft 상태로 생성할까요?" 프롬프트가 표시되면
THEN 기본값은 true여야 한다
AND 사용자가 Yes를 선택하면 git_strategy.team.draft_pr이 true여야 한다
AND "검토 후 Ready 전환" 설명이 포함되어야 한다
```

**검증 방법**:
```bash
# 실행 (Team 모드 + Auto PR Yes + Draft PR Yes)
npm run install

# 검증
cat .moai/config.json | jq '.git_strategy.team.draft_pr'
# 예상 출력: true
```

---

#### 시나리오 2: Draft PR 비활성화
```gherkin
GIVEN 사용자가 Team 모드를 선택하고 Auto PR을 활성화했을 때
WHEN "PR을 Draft 상태로 생성할까요?" 프롬프트에서 No를 선택하면
THEN git_strategy.team.draft_pr이 false여야 한다
AND PR이 Ready 상태로 즉시 생성되어야 한다
```

**검증 방법**:
```bash
# 실행 (Team 모드 + Auto PR Yes + Draft PR No)
npm run install

# 검증
cat .moai/config.json | jq '.git_strategy.team.draft_pr'
# 예상 출력: false
```

---

#### 시나리오 3: Auto PR 비활성화 시 Draft PR 표시 안 함
```gherkin
GIVEN 사용자가 Team 모드를 선택하고 Auto PR을 비활성화했을 때
WHEN Git 전략 설정이 완료되면
THEN Draft PR 프롬프트를 표시하지 않아야 한다
AND git_strategy.team.draft_pr이 false여야 한다
```

**검증 방법**:
- 단위 테스트: `phase4-git.test.ts`에서 조건부 로직 확인

---

### AC-006: Alfred 대화형 페르소나

#### 시나리오 1: 환영 메시지 출력
```gherkin
GIVEN 모든 프롬프트가 완료되고 config.json이 저장되었을 때
WHEN Phase 5 (환영 메시지)가 실행되면
THEN "✅ MoAI-ADK 설치가 완료되었습니다!" 메시지를 출력해야 한다
AND "🤖 AI-Agent Alfred가 {name}님의 개발을 도와드리겠습니다" 메시지를 출력해야 한다
AND {name}은 Phase 2에서 입력한 개발자 이름이어야 한다
AND 다음 명령어 안내를 제공해야 한다
  - /alfred:8-project (프로젝트 초기화)
  - /alfred:1-spec (첫 SPEC 작성)
AND "@agent-debug-helper 호출 안내"를 포함해야 한다
```

**검증 방법**:
```bash
# 실행 (개발자 이름: "홍길동")
npm run install

# 예상 출력
# ✅ MoAI-ADK 설치가 완료되었습니다!
#
# 🤖 AI-Agent Alfred가 홍길동님의 개발을 도와드리겠습니다.
#
# 다음 명령어로 시작하세요:
#   /alfred:8-project  # 프로젝트 초기화
#   /alfred:1-spec     # 첫 SPEC 작성
#
# 질문이 있으시면 언제든 @agent-debug-helper를 호출하세요.
```

---

#### 시나리오 2: Progressive Disclosure 준수
```gherkin
GIVEN 설치 프롬프트가 진행될 때
WHEN 각 Phase를 실행하면
THEN Phase 순서는 다음과 같아야 한다
  1. Git 검증 + 모드 선택
  2. 개발자 정보 수집
  3. SPEC Workflow 선택 (Personal 모드만)
  4. Auto PR/Draft PR 선택 (Team 모드만)
  5. 환영 메시지
AND Personal 모드에서 Team 전용 프롬프트를 표시하지 않아야 한다
AND Team 모드에서 Personal 전용 프롬프트를 표시하지 않아야 한다
AND Draft PR은 Auto PR 활성화 시에만 표시되어야 한다
```

**검증 방법**:
- 통합 테스트: 모드별 시나리오 실행 후 프롬프트 순서 확인

---

#### 시나리오 3: Alfred 톤 일관성
```gherkin
GIVEN 모든 프롬프트 메시지가 출력될 때
WHEN 메시지를 확인하면
THEN 존댓말을 사용해야 한다 ("입력해주세요", "선택해주세요")
AND 이유 설명을 제공해야 한다 ("Git은 코드 버전 관리에 필수입니다")
AND 에러 시 해결책을 제시해야 한다 ("다음 방법으로 설치하세요")
AND 적절한 이모지를 사용해야 한다 (✅, ❌, 🤖, ℹ️)
```

**검증 방법**:
- 코드 리뷰: 모든 프롬프트 메시지 톤 검토
- 수동 테스트: 설치 과정 전체 진행 후 UX 평가

---

## 2. 품질 게이트 (TRUST 원칙)

### Test First (테스트 우선)

#### 단위 테스트 커버리지
- **목표**: ≥85%
- **대상 파일**:
  - `phase1-basic.ts`
  - `phase2-developer.ts`
  - `phase3-mode.ts`
  - `phase4-git.ts`
  - `phase5-welcome.ts`
- **검증**: `npm run test:coverage`

#### 테스트 케이스 우선순위
1. **High**: Git 검증, 개발자 이름 수집, Progressive Disclosure
2. **Medium**: Auto PR/Draft PR 조건부 로직
3. **Low**: 환영 메시지 포맷

---

### Readable (가독성)

#### 코드 복잡도 제한
- **함수 LOC**: ≤50줄
- **매개변수**: ≤5개
- **복잡도**: ≤10

#### 명명 규칙
- **Phase 함수**: `executePhaseN()` (N은 1~5)
- **검증 함수**: `validate*()` (예: `validateGitInstallation()`)
- **수집 함수**: `collect*()` (예: `collectDeveloperInfo()`)

#### 주석 가이드
- **TAG 주석**: 모든 파일 상단에 `@CODE:INSTALL-001` 포함
- **함수 주석**: JSDoc 스타일 (역할, 파라미터, 반환값)
- **복잡한 로직**: 인라인 주석으로 이유 설명

---

### Unified (통일성)

#### TypeScript 타입 안전성
- **모든 변수**: 명시적 타입 선언
- **함수 반환값**: 인터페이스 정의 (`Phase1Result`, `Phase2Result` 등)
- **Strict Mode**: `tsconfig.json`의 `strict: true` 유지

#### 에러 처리 일관성
```typescript
// 패턴 1: Git 명령 실패
try {
  const result = await execCommand('git --version');
} catch (error) {
  console.error('❌ Git이 설치되지 않았습니다.');
  process.exit(1);
}

// 패턴 2: Inquirer 검증
validate: (input) => {
  if (input.trim() === '') {
    return '❌ 필드는 필수입니다.';
  }
  return true;
}
```

---

### Secured (보안)

#### 입력 검증
- **개발자 이름**: XSS 방지 (특수문자 제한 또는 이스케이프)
- **경로 접근**: `.moai/config.json` 경로 외부 접근 차단
- **명령 실행**: `execCommand()` 함수에서 커맨드 인젝션 방지

#### 민감 정보 보호
- **개발자 이메일**: 향후 구현 시 평문 저장 금지 (Git 전역 설정 참조만)
- **Git 토큰**: config.json에 저장 금지

---

### Trackable (추적성)

#### TAG 체인 완전성
```bash
# SPEC 문서 확인
rg '@SPEC:INSTALL-001' -n .moai/specs/

# 테스트 코드 확인
rg '@TEST:INSTALL-001' -n tests/

# 구현 코드 확인
rg '@CODE:INSTALL-001' -n src/

# 고아 TAG 검증
rg '@(SPEC|TEST|CODE):INSTALL-001' -n
```

#### Git 커밋 메시지
- **형식**: `feat(prompts): {변경 내용}`
- **서명**: `Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>`
- **TAG 참조**: 커밋 본문에 `@CODE:INSTALL-001` 포함

---

## 3. 검증 방법

### 3.1 자동 테스트

#### 단위 테스트 실행
```bash
# Phase 2 테스트
npm run test tests/cli/prompts/init/phase2-developer.test.ts

# Phase 4 테스트
npm run test tests/cli/prompts/init/phase4-git.test.ts

# 전체 테스트
npm run test
```

#### 커버리지 확인
```bash
npm run test:coverage

# 예상 결과
# File                     | % Stmts | % Branch | % Funcs | % Lines
# phase1-basic.ts          |     90  |      85  |     100 |      90
# phase2-developer.ts      |     95  |      90  |     100 |      95
# phase3-mode.ts           |     90  |      80  |     100 |      90
# phase4-git.ts            |     95  |      90  |     100 |      95
# phase5-welcome.ts        |     100 |     100  |     100 |     100
```

---

### 3.2 통합 테스트

#### E2E 시나리오 실행
```bash
# Personal 모드 + SPEC 활성화
npm run test:e2e -- --scenario=personal-spec-enabled

# Team 모드 + Auto PR 비활성화
npm run test:e2e -- --scenario=team-no-autopr

# Git 미설치 에러
npm run test:e2e -- --scenario=git-not-installed
```

---

### 3.3 수동 테스트

#### 체크리스트 (QA)

**시나리오 1: Personal 모드 (정상 흐름)**
- [ ] Git 설치 확인 메시지 출력
- [ ] 모드 선택 프롬프트 표시
- [ ] 개발자 이름 프롬프트 (Git user.name 기본값)
- [ ] SPEC Workflow 선택 프롬프트 표시
- [ ] Auto PR/Draft PR 프롬프트 표시 안 함
- [ ] 환영 메시지 출력 (개발자 이름 포함)
- [ ] `.moai/config.json` 생성 확인

**시나리오 2: Team 모드 (정상 흐름)**
- [ ] Git 설치 확인 메시지 출력
- [ ] 모드 선택 프롬프트 표시
- [ ] 개발자 이름 프롬프트
- [ ] SPEC Workflow 프롬프트 표시 안 함 (강제 활성화 안내)
- [ ] Auto PR 프롬프트 표시
- [ ] Draft PR 프롬프트 표시 (Auto PR Yes 시만)
- [ ] 환영 메시지 출력
- [ ] `.moai/config.json` 생성 확인

**시나리오 3: Git 미설치 (에러 흐름)**
- [ ] Git 미설치 에러 메시지 출력
- [ ] macOS/Ubuntu/Windows 설치 안내 표시
- [ ] `process.exit(1)` 실행 확인
- [ ] 다음 프롬프트 표시 안 함

**시나리오 4: 개발자 이름 빈 값 (검증 흐름)**
- [ ] 빈 값 입력 시 "❌ 개발자 이름은 필수입니다" 메시지
- [ ] 프롬프트 재표시
- [ ] 유효한 값 입력까지 반복

---

### 3.4 성능 테스트

#### 설치 시간 측정
```bash
# 목표: ≤5초 (프롬프트 입력 시간 제외)
time npm run install
```

#### Git 명령 캐싱 확인
- `getGitUserName()` 함수가 중복 호출되지 않는지 검증
- Phase 1에서 Git 버전 정보 캐싱 확인

---

## 4. Definition of Done (완료 조건)

### 코드 완료 기준
- [ ] 모든 Phase (1~5) 구현 완료
- [ ] TAG 주석 (`@CODE:INSTALL-001`) 모든 파일에 포함
- [ ] TypeScript 빌드 에러 0건
- [ ] ESLint 경고 0건

### 테스트 완료 기준
- [ ] 단위 테스트 커버리지 ≥85%
- [ ] 통합 테스트 모든 시나리오 통과
- [ ] E2E 테스트 (Personal/Team 모드) 통과
- [ ] Git 미설치 에러 처리 검증

### 문서 완료 기준
- [ ] `spec.md` EARS 구문 준수
- [ ] `plan.md` 구현 계획 상세 작성
- [ ] `acceptance.md` (본 문서) Given-When-Then 시나리오 작성
- [ ] TAG 체인 검증 (`rg '@(SPEC|TEST|CODE):INSTALL-001' -n`)

### 사용자 경험 완료 기준
- [ ] Alfred 톤 일관성 (존댓말, 이유 설명, 해결책)
- [ ] Progressive Disclosure 준수 (모드별 조건부 프롬프트)
- [ ] 에러 메시지 친절성 (Git 미설치, 빈 값 입력)
- [ ] 환영 메시지 출력 (개발자 이름 포함)

### 배포 완료 기준
- [ ] `/alfred:3-sync` 실행 (문서 동기화)
- [ ] Living Document 생성 확인
- [ ] PR 생성 (Draft → Ready 전환)
- [ ] 코드 리뷰 승인 (1명 이상)

---

## 5. 회귀 테스트 (Regression Test)

### 기존 기능 영향 분석

#### 영향 받는 기능
1. **프로젝트 초기화** (`/alfred:8-project`)
   - `.moai/config.json` 스키마 변경 (developer 필드 추가)
   - 하위 호환성 검증 필요

2. **Git 커밋 서명**
   - `developer.name` 필드 활용 여부 확인
   - `Co-Authored-By` 메시지 생성 로직 업데이트

3. **SPEC 강제 여부** (`/alfred:1-spec`)
   - `constitution.enforce_spec` 필드 참조
   - Personal 모드에서 경고 메시지 표시 (비활성화 시)

#### 회귀 테스트 시나리오
```bash
# 시나리오 1: 기존 프로젝트 호환성
# Given: v0.0.x config.json 존재
# When: v0.1.0 설치
# Then: 경고만 출력, 설치 차단 X

# 시나리오 2: Git 커밋 서명 검증
# Given: developer.name 설정됨
# When: git commit 생성
# Then: Co-Authored-By에 developer.name 포함

# 시나리오 3: SPEC 경고 메시지
# Given: Personal 모드 + enforce_spec: false
# When: /alfred:1-spec 실행
# Then: 경고 메시지 출력, 실행은 허용
```

---

## 6. 롤백 계획

### 롤백 트리거
- **치명적 버그**: Git 검증 로직 오작동으로 정상 환경에서 설치 차단
- **성능 저하**: 설치 시간 >10초 (목표의 2배 초과)
- **데이터 손실**: config.json 덮어쓰기 시 백업 미생성

### 롤백 절차
1. **즉시 조치**:
   ```bash
   git revert {commit_hash}
   npm run build
   npm publish --tag rollback
   ```

2. **사용자 안내**:
   - GitHub Release Notes에 롤백 사유 게시
   - npm 패키지 v0.0.x로 다운그레이드 안내

3. **버그 수정**:
   - 핫픽스 브랜치 생성
   - 긴급 패치 버전 릴리스 (v0.1.1)

---

## 7. 다음 단계

### 즉시 실행
1. `/alfred:2-build SPEC-INSTALL-001` → TDD 구현 시작
2. Phase 2 (개발자 정보) 우선 개발
3. 단위 테스트 작성 및 실행

### 검증 단계
1. Personal 모드 E2E 테스트
2. Team 모드 E2E 테스트
3. Git 미설치 에러 처리 확인
4. 회귀 테스트 실행

### 마무리
1. `/alfred:3-sync` 실행 (문서 동기화)
2. TAG 체인 검증
3. Living Document 생성
4. PR 생성 및 코드 리뷰

---

_이 인수 기준은 SPEC-INSTALL-001의 품질 게이트입니다._
_모든 시나리오가 통과되어야 Definition of Done이 충족됩니다._
