---
id: INSTALL-001
version: 0.1.0
status: draft
created: 2025-10-06
updated: 2025-10-06
author: @Goos
reference: .moai/reports/moai-adk-redesign-masterplan.md
labels:
  - cli
  - prompts
  - install
  - ux
  - developer-experience
priority: high
---

# @SPEC:INSTALL-001: Install Prompts Redesign - Developer Name, Git Mandatory & PR Automation

## HISTORY

### v0.1.0 (2025-10-06)
- **INITIAL**: Install Prompts Redesign 명세 작성
- **AUTHOR**: @Goos
- **SCOPE**:
  - 개발자 이름 프롬프트 추가 (Git 커밋 서명용)
  - Git 필수화 (Personal/Team 모드 모두)
  - SPEC Workflow 필수화 (Team 모드)
  - Auto PR/Draft PR 사용자 선택 프롬프트
  - Alfred 대화형 페르소나 강화 (Progressive Disclosure)
- **CONTEXT**:
  - 현재 설치 프롬프트가 Git/SPEC 필수 정보를 누락하여 사용자 경험 저하
  - 개발자 이름 미수집으로 Git 커밋 서명 불완전
  - Auto PR/Draft PR이 하드코딩되어 사용자 선택권 부재

---

## Environment (환경 및 전제조건)

### 실행 환경
- **프로젝트**: MoAI-ADK (TypeScript 기반 CLI 도구)
- **CLI 프레임워크**: Commander.js
- **프롬프트 라이브러리**: Inquirer.js 또는 동급 대화형 프롬프트 도구
- **타겟 사용자**: Personal 모드 (솔로 개발자) + Team 모드 (협업 팀)

### 기술 스택
- **언어**: TypeScript
- **런타임**: Node.js ≥ 18.0.0
- **의존성**:
  - `inquirer`: 대화형 프롬프트
  - `chalk`: 콘솔 색상 출력
  - `commander`: CLI 명령어 파싱
  - `fs-extra`: 파일 시스템 조작

### 제약사항
- **하위 호환성**: 기존 `.moai/config.json` 구조 유지 (새로운 필드 추가만 허용)
- **다국어 지원**: `locale: ko` 우선 (향후 다국어 확장 가능하도록 설계)
- **Progressive Disclosure**: 복잡한 옵션은 단계별로 노출 (인지 부담 최소화)

---

## Assumptions (가정사항)

1. **Git 필수 가정**:
   - Personal 모드: Git 체크포인트 자동화 필수
   - Team 모드: GitFlow 전략 필수
   - 설치 시 Git 미설치 사용자는 에러 안내 후 중단

2. **SPEC Workflow 가정**:
   - Team 모드는 `/alfred:1-spec` → 브랜치 → PR 흐름 필수
   - Personal 모드는 SPEC 선택적 (간단한 프로토타입용)

3. **개발자 이름 가정**:
   - Git 전역 설정 `user.name` 우선 사용
   - 미설정 시 프롬프트로 수집 → `.moai/config.json`과 Git 전역 설정에 반영

4. **Auto PR/Draft PR 가정**:
   - 기본값: `auto_pr: true`, `draft_pr: true` (안전한 기본값)
   - 사용자가 설치 시 선택 가능 (Team 모드만)

5. **Alfred 페르소나 가정**:
   - 친절하지만 간결한 대화 스타일
   - 전문 용어는 괄호 안에 설명 (예: "GitFlow (Git 브랜치 전략)")
   - 에러 메시지는 해결 방법 제시

---

## Requirements (EARS 요구사항)

### Ubiquitous Requirements (기본 기능)

**UR-001**: 시스템은 설치 시 개발자 이름을 수집해야 한다
- **입력**: Git `user.name` 또는 프롬프트 입력
- **저장**: `.moai/config.json` (`developer.name`)
- **용도**: Git 커밋 서명 `Co-Authored-By: {name} <email>`

**UR-002**: 시스템은 Git을 필수 요구사항으로 설정해야 한다 (Personal/Team 모드 공통)
- **검증**: 설치 시 `git --version` 확인
- **실패 시**: Git 설치 안내 메시지 출력 후 중단
- **이유**: Personal(체크포인트), Team(GitFlow) 모두 Git 필수

**UR-003**: 시스템은 SPEC Workflow를 필수로 설정해야 한다 (Team 모드만)
- **검증**: Team 모드 선택 시 SPEC 워크플로우 자동 활성화
- **제약**: Personal 모드는 SPEC 선택적 (프롬프트 추가)

---

### Event-driven Requirements (이벤트 기반)

**ER-001**: WHEN 사용자가 Team 모드를 선택하면, 시스템은 Auto PR 사용 여부를 물어봐야 한다
- **조건**: 설치 프롬프트에서 `mode: team` 선택
- **프롬프트**: "자동으로 PR을 생성할까요? (Auto PR)"
  - Yes → `git_strategy.team.auto_pr: true`
  - No → `git_strategy.team.auto_pr: false`
- **기본값**: `true` (안전한 기본값)

**ER-002**: WHEN 사용자가 Auto PR을 활성화하면, 시스템은 Draft PR 사용 여부를 물어봐야 한다
- **조건**: `auto_pr: true` 선택 후
- **프롬프트**: "PR을 Draft 상태로 생성할까요? (Draft PR)"
  - Yes → `git_strategy.team.draft_pr: true`
  - No → `git_strategy.team.draft_pr: false`
- **기본값**: `true` (검토 후 Ready 전환 권장)

**ER-003**: WHEN Git이 설치되지 않았으면, 시스템은 친절한 에러 메시지를 출력하고 설치를 중단해야 한다
- **조건**: `git --version` 실패
- **에러 메시지**:
  ```
  ❌ Git이 설치되지 않았습니다.

  MoAI-ADK는 Git을 필수로 사용합니다.
  다음 방법으로 Git을 설치해주세요:

  macOS: brew install git
  Ubuntu: sudo apt-get install git
  Windows: https://git-scm.com/download/win

  설치 후 다시 시도해주세요.
  ```

**ER-004**: WHEN 개발자 이름 프롬프트 표시 시, 시스템은 Git `user.name`을 기본값으로 제안해야 한다
- **조건**: Git 전역 설정에 `user.name` 존재
- **프롬프트**: "개발자 이름을 입력해주세요: (기본값: {git_user_name})"
- **동작**:
  - Enter → Git `user.name` 사용
  - 직접 입력 → 새 이름 사용 + `.moai/config.json` 저장

**ER-005**: WHEN Personal 모드 선택 시, 시스템은 SPEC Workflow 사용 여부를 물어봐야 한다
- **조건**: `mode: personal` 선택
- **프롬프트**: "SPEC-First Workflow를 사용할까요? (권장)"
  - Yes → `constitution.enforce_spec: true`
  - No → `constitution.enforce_spec: false`
- **기본값**: `true` (권장 설정)

**ER-006**: WHEN 사용자가 모든 프롬프트에 응답하면, 시스템은 Alfred 페르소나로 환영 메시지를 출력해야 한다
- **조건**: 설치 완료 후
- **메시지**:
  ```
  ✅ MoAI-ADK 설치가 완료되었습니다!

  🤖 AI-Agent Alfred가 당신의 개발을 도와드리겠습니다.

  다음 명령어로 시작하세요:
  /alfred:8-project  # 프로젝트 초기화
  /alfred:1-spec     # 첫 SPEC 작성

  질문이 있으시면 언제든 @agent-debug-helper를 호출하세요.
  ```

---

### State-driven Requirements (상태 기반)

**SR-001**: WHILE 설치 프롬프트 진행 중일 때, 시스템은 Progressive Disclosure 원칙을 따라야 한다
- **상태**: 프롬프트 단계별 진행 중
- **동작**:
  1. 필수 정보 먼저 (모드 선택, Git 검증, 개발자 이름)
  2. 모드별 추가 정보 (Auto PR, Draft PR는 Team 모드만)
  3. 선택적 정보 마지막 (SPEC Workflow는 Personal 모드만)
- **이유**: 인지 부담 최소화, 중요한 결정 먼저

**SR-002**: WHILE Team 모드일 때, SPEC Workflow는 항상 활성화되어야 한다
- **상태**: `mode: team`
- **동작**: `constitution.enforce_spec: true` 강제 설정
- **제약**: 사용자가 비활성화 불가 (Team 협업 필수 요소)

**SR-003**: WHILE 프롬프트 출력 중일 때, Alfred 페르소나 톤을 유지해야 한다
- **상태**: 모든 프롬프트 메시지
- **톤 가이드**:
  - 존댓말 사용 ("입력해주세요", "선택해주세요")
  - 이유 설명 제공 ("Git은 코드 버전 관리에 필수입니다")
  - 에러 시 해결책 제시 ("다음 방법으로 설치하세요")
  - 이모지 적절히 사용 (✅ ❌ 🤖)

---

### Optional Features (선택적 기능)

**OF-001**: WHERE 개발자 이메일 수집이 필요하면, 시스템은 Git `user.email`을 함께 수집할 수 있다
- **조건**: 향후 확장 (v0.2.0 이후)
- **구현**: `developer.email` 필드 추가
- **우선순위**: Low (현재는 Git 전역 설정 사용)

**OF-002**: WHERE 다국어 지원이 필요하면, 프롬프트 메시지를 i18n 구조로 관리할 수 있다
- **조건**: 향후 확장 (v0.3.0 이후)
- **구현**: `src/i18n/prompts/{locale}.json` 파일
- **우선순위**: Low (현재는 한국어 하드코딩)

---

### Constraints (제약사항)

**C-001**: IF `.moai/config.json`이 이미 존재하면, 시스템은 덮어쓰기 경고를 표시해야 한다
- **조건**: 재설치 시도
- **동작**: "기존 설정을 덮어쓸까요? (데이터 손실 주의)" 프롬프트
- **옵션**:
  - Yes → 백업 후 덮어쓰기 (`.moai/config.json.backup`)
  - No → 설치 중단

**C-002**: IF 하위 호환성을 유지해야 하면, 새로운 필드는 선택적이어야 한다
- **조건**: 기존 프로젝트 마이그레이션
- **동작**: `developer` 필드 누락 시 경고만 출력 (설치 차단 X)
- **마이그레이션**: 향후 `/alfred:upgrade` 명령 제공

**C-003**: IF Personal 모드에서 SPEC을 비활성화하면, `/alfred:1-spec` 명령은 경고를 출력해야 한다
- **조건**: `mode: personal`, `enforce_spec: false`
- **경고**: "SPEC Workflow가 비활성화되어 있습니다. 활성화하려면 `.moai/config.json`에서 `enforce_spec: true`로 변경하세요."
- **동작**: 명령 실행은 허용 (제약 아님)

---

## Specifications (상세 명세)

### 1. 새로운 프롬프트 흐름

#### 단계 1: 모드 선택 (기존 유지)
```typescript
{
  type: 'list',
  name: 'mode',
  message: '프로젝트 모드를 선택하세요:',
  choices: [
    { name: 'Personal - 혼자 개발하는 프로젝트', value: 'personal' },
    { name: 'Team - 팀으로 협업하는 프로젝트', value: 'team' }
  ]
}
```

#### 단계 2: Git 검증 (NEW)
```typescript
// Git 설치 확인
const gitVersion = await execCommand('git --version');
if (!gitVersion) {
  console.error(`
❌ Git이 설치되지 않았습니다.

MoAI-ADK는 Git을 필수로 사용합니다.
다음 방법으로 Git을 설치해주세요:

macOS: brew install git
Ubuntu: sudo apt-get install git
Windows: https://git-scm.com/download/win

설치 후 다시 시도해주세요.
  `);
  process.exit(1);
}
```

#### 단계 3: 개발자 이름 수집 (NEW)
```typescript
// Git user.name 조회
const gitUserName = await execCommand('git config --global user.name');

const { developerName } = await inquirer.prompt([
  {
    type: 'input',
    name: 'developerName',
    message: '개발자 이름을 입력해주세요:',
    default: gitUserName || '',
    validate: (input) => input.trim() !== '' || '개발자 이름은 필수입니다.'
  }
]);

// config.json에 저장
config.developer = {
  name: developerName,
  timestamp: new Date().toISOString()
};
```

#### 단계 4: SPEC Workflow 선택 (Personal 모드만, NEW)
```typescript
if (mode === 'personal') {
  const { enforceSpec } = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'enforceSpec',
      message: 'SPEC-First Workflow를 사용할까요? (권장)',
      default: true
    }
  ]);

  config.constitution.enforce_spec = enforceSpec;
} else {
  // Team 모드는 SPEC 강제 활성화
  config.constitution.enforce_spec = true;
}
```

#### 단계 5: Auto PR 선택 (Team 모드만, NEW)
```typescript
if (mode === 'team') {
  const { autoPR } = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'autoPR',
      message: '자동으로 PR을 생성할까요? (Auto PR)',
      default: true
    }
  ]);

  config.git_strategy.team.auto_pr = autoPR;
}
```

#### 단계 6: Draft PR 선택 (Team 모드 + Auto PR 활성화 시만, NEW)
```typescript
if (mode === 'team' && config.git_strategy.team.auto_pr) {
  const { draftPR } = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'draftPR',
      message: 'PR을 Draft 상태로 생성할까요? (검토 후 Ready 전환)',
      default: true
    }
  ]);

  config.git_strategy.team.draft_pr = draftPR;
}
```

#### 단계 7: 환영 메시지 (NEW)
```typescript
console.log(`
✅ MoAI-ADK 설치가 완료되었습니다!

🤖 AI-Agent Alfred가 ${developerName}님의 개발을 도와드리겠습니다.

다음 명령어로 시작하세요:
/alfred:8-project  # 프로젝트 초기화
/alfred:1-spec     # 첫 SPEC 작성

질문이 있으시면 언제든 @agent-debug-helper를 호출하세요.
`);
```

---

### 2. 수정 대상 파일 10개 상세

#### 핵심 수정 파일 (7개)

1. **`src/cli/prompts/init/phase1-basic.ts`** (Phase 1: 기본 정보)
   - **수정 내용**: Git 검증 로직 추가
   - **함수**: `validateGitInstallation()` 추가
   - **변경 라인**: ~30줄 추가

2. **`src/cli/prompts/init/phase2-developer.ts`** (Phase 2: 개발자 정보, NEW)
   - **생성 여부**: 신규 파일
   - **역할**: 개발자 이름 수집
   - **함수**: `collectDeveloperInfo()`
   - **라인 수**: ~50줄

3. **`src/cli/prompts/init/phase3-mode.ts`** (Phase 3: 모드 선택)
   - **수정 내용**: SPEC Workflow 프롬프트 추가 (Personal 모드만)
   - **변경 라인**: ~20줄 추가

4. **`src/cli/prompts/init/phase4-git.ts`** (Phase 4: Git 전략, RENAME)
   - **기존 이름**: `phase3-git.ts`
   - **수정 내용**: Auto PR/Draft PR 프롬프트 추가
   - **변경 라인**: ~40줄 추가

5. **`src/cli/prompts/init/phase5-welcome.ts`** (Phase 5: 환영 메시지, NEW)
   - **생성 여부**: 신규 파일
   - **역할**: Alfred 페르소나 환영 메시지 출력
   - **함수**: `displayWelcomeMessage(config)`
   - **라인 수**: ~30줄

6. **`src/cli/prompts/init/index.ts`** (Phase 통합)
   - **수정 내용**: Phase 2, 5 추가, Phase 순서 재조정
   - **변경 라인**: ~15줄 수정

7. **`src/core/installer/phase-executor.ts`** (Phase 실행기)
   - **수정 내용**: Phase 2, 5 실행 로직 추가
   - **변경 라인**: ~20줄 수정

#### 검증 파일 (3개)

8. **`tests/cli/prompts/init/phase2-developer.test.ts`** (NEW)
   - **테스트 내용**:
     - Git `user.name` 조회 테스트
     - 개발자 이름 입력 검증
     - config.json 저장 확인

9. **`tests/cli/prompts/init/phase4-git.test.ts`** (기존 파일 확장)
   - **추가 테스트**:
     - Auto PR 프롬프트 테스트
     - Draft PR 프롬프트 테스트 (Auto PR 활성화 시만)

10. **`tests/core/installer/integration.test.ts`** (통합 테스트)
    - **추가 시나리오**:
      - Personal 모드 + SPEC 비활성화
      - Team 모드 + Auto PR 비활성화
      - Git 미설치 에러 처리

---

### 3. `.moai/config.json` 스키마 변경

#### 추가 필드: `developer` (NEW)
```json
{
  "developer": {
    "name": "홍길동",
    "timestamp": "2025-10-06T12:00:00.000Z"
  }
}
```

#### 추가 필드: `constitution.enforce_spec` (NEW)
```json
{
  "constitution": {
    "enforce_tdd": true,
    "enforce_spec": true,  // NEW
    "require_tags": true,
    "test_coverage_target": 85
  }
}
```

#### 기존 필드 유지: `git_strategy.team.auto_pr`, `draft_pr`
```json
{
  "git_strategy": {
    "team": {
      "auto_pr": true,      // 기존 하드코딩 → 프롬프트로 선택
      "draft_pr": true      // 기존 하드코딩 → 프롬프트로 선택
    }
  }
}
```

---

### 4. Alfred 페르소나 대화 스타일 가이드

#### 프롬프트 메시지 톤
- **존댓말 사용**: "입력해주세요", "선택해주세요"
- **이유 설명**: "Git은 코드 버전 관리에 필수입니다"
- **에러 해결책**: "다음 방법으로 설치하세요:"
- **이모지 사용**: ✅ (성공), ❌ (에러), 🤖 (Alfred)

#### 환영 메시지 예시
```
✅ MoAI-ADK 설치가 완료되었습니다!

🤖 AI-Agent Alfred가 홍길동님의 개발을 도와드리겠습니다.

다음 명령어로 시작하세요:
/alfred:8-project  # 프로젝트 초기화
/alfred:1-spec     # 첫 SPEC 작성

질문이 있으시면 언제든 @agent-debug-helper를 호출하세요.
```

#### Git 미설치 에러 메시지 예시
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

### 5. 마이그레이션 전략 (기존 프로젝트)

#### 시나리오 1: v0.0.x → v0.1.0 (본 SPEC)
**문제**: 기존 프로젝트에 `developer` 필드 누락

**해결책**:
```typescript
// src/core/config/validator.ts
function validateConfig(config: Config): ValidationResult {
  const warnings = [];

  if (!config.developer) {
    warnings.push('developer 필드가 없습니다. Git 커밋 서명이 불완전할 수 있습니다.');
  }

  if (config.constitution.enforce_spec === undefined) {
    warnings.push('enforce_spec 필드가 없습니다. 기본값 true로 설정합니다.');
    config.constitution.enforce_spec = true;
  }

  return { valid: true, warnings };
}
```

#### 시나리오 2: 수동 마이그레이션 가이드
**문서**: `.moai/memory/migration-guide.md`

**내용**:
```markdown
# v0.0.x → v0.1.0 마이그레이션 가이드

## 1. 개발자 정보 추가
`.moai/config.json`에 다음 필드를 추가하세요:
\`\`\`json
{
  "developer": {
    "name": "본인 이름",
    "timestamp": "2025-10-06T12:00:00.000Z"
  }
}
\`\`\`

## 2. SPEC 강제 여부 설정
\`\`\`json
{
  "constitution": {
    "enforce_spec": true
  }
}
\`\`\`
```

---

### 6. 성능 최적화 전략

#### Git 명령 캐싱
```typescript
// 반복 호출 방지
let cachedGitUserName: string | null = null;

async function getGitUserName(): Promise<string | null> {
  if (cachedGitUserName !== null) {
    return cachedGitUserName;
  }

  cachedGitUserName = await execCommand('git config --global user.name');
  return cachedGitUserName;
}
```

#### 프롬프트 건너뛰기 (CI/CD 환경)
```typescript
// 환경 변수로 자동 설정
if (process.env.CI) {
  return {
    mode: process.env.MOAI_MODE || 'personal',
    developerName: process.env.MOAI_DEVELOPER_NAME || 'CI Bot',
    autoPR: process.env.MOAI_AUTO_PR === 'true'
  };
}
```

---

## Acceptance Criteria (수락 기준)

### AC1: Git 필수 검증
```gherkin
GIVEN 사용자가 MoAI-ADK 설치를 시작할 때
WHEN Git이 설치되지 않았으면
THEN "❌ Git이 설치되지 않았습니다" 에러 메시지를 출력해야 한다
AND 설치 안내 (macOS/Ubuntu/Windows)를 제공해야 한다
AND 설치를 중단해야 한다
```

### AC2: 개발자 이름 수집
```gherkin
GIVEN 사용자가 설치 프롬프트를 진행할 때
WHEN "개발자 이름을 입력해주세요" 프롬프트가 표시되면
THEN Git user.name을 기본값으로 제안해야 한다
AND 사용자 입력값을 .moai/config.json의 developer.name에 저장해야 한다
AND 빈 값은 허용하지 않아야 한다
```

### AC3: SPEC Workflow 선택 (Personal 모드)
```gherkin
GIVEN 사용자가 Personal 모드를 선택했을 때
WHEN "SPEC-First Workflow를 사용할까요?" 프롬프트가 표시되면
THEN 기본값은 true여야 한다
AND Yes 선택 시 constitution.enforce_spec: true로 저장해야 한다
AND No 선택 시 constitution.enforce_spec: false로 저장해야 한다
```

### AC4: Auto PR 선택 (Team 모드)
```gherkin
GIVEN 사용자가 Team 모드를 선택했을 때
WHEN "자동으로 PR을 생성할까요?" 프롬프트가 표시되면
THEN 기본값은 true여야 한다
AND Yes 선택 시 git_strategy.team.auto_pr: true로 저장해야 한다
AND No 선택 시 git_strategy.team.auto_pr: false로 저장해야 한다
```

### AC5: Draft PR 선택 (Team 모드 + Auto PR 활성화)
```gherkin
GIVEN 사용자가 Team 모드 + Auto PR을 활성화했을 때
WHEN "PR을 Draft 상태로 생성할까요?" 프롬프트가 표시되면
THEN 기본값은 true여야 한다
AND Yes 선택 시 git_strategy.team.draft_pr: true로 저장해야 한다
AND Auto PR이 비활성화되면 이 프롬프트를 표시하지 않아야 한다
```

### AC6: Alfred 환영 메시지
```gherkin
GIVEN 모든 프롬프트가 완료되었을 때
WHEN 설치가 성공하면
THEN "✅ MoAI-ADK 설치가 완료되었습니다!" 메시지를 출력해야 한다
AND "🤖 AI-Agent Alfred가 {name}님의 개발을 도와드리겠습니다" 메시지를 출력해야 한다
AND 다음 단계 안내 (/alfred:8-project, /alfred:1-spec)를 제공해야 한다
```

### AC7: Progressive Disclosure 준수
```gherkin
GIVEN 설치 프롬프트 흐름을 진행할 때
WHEN 단계별로 질문하면
THEN 순서는 "모드 선택 → Git 검증 → 개발자 이름 → 모드별 옵션" 순이어야 한다
AND Team 모드 전용 옵션은 Personal 모드에서 표시되지 않아야 한다
AND Draft PR 프롬프트는 Auto PR 활성화 시에만 표시되어야 한다
```

### AC8: 하위 호환성 유지
```gherkin
GIVEN 기존 v0.0.x 프로젝트를 v0.1.0으로 업그레이드할 때
WHEN .moai/config.json에 developer 필드가 없으면
THEN 경고만 출력하고 설치를 차단하지 않아야 한다
AND enforce_spec 필드가 없으면 기본값 true로 설정해야 한다
AND 기존 auto_pr, draft_pr 값은 유지되어야 한다
```

---

## Traceability (@TAG 체인)

### TAG 체인 구조
```
@SPEC:INSTALL-001 (본 문서)
  ↓
@TEST:INSTALL-001
  ├─ tests/cli/prompts/init/phase2-developer.test.ts
  ├─ tests/cli/prompts/init/phase4-git.test.ts
  └─ tests/core/installer/integration.test.ts
  ↓
@CODE:INSTALL-001
  ├─ src/cli/prompts/init/phase1-basic.ts (Git 검증)
  ├─ src/cli/prompts/init/phase2-developer.ts (개발자 정보, NEW)
  ├─ src/cli/prompts/init/phase3-mode.ts (SPEC 선택)
  ├─ src/cli/prompts/init/phase4-git.ts (Auto/Draft PR)
  ├─ src/cli/prompts/init/phase5-welcome.ts (환영 메시지, NEW)
  ├─ src/cli/prompts/init/index.ts (Phase 통합)
  └─ src/core/installer/phase-executor.ts (Phase 실행기)
  ↓
@DOC:INSTALL-001 (본 SPEC 문서 + 마이그레이션 가이드)
```

### 검증 명령어
```bash
# SPEC 문서 확인
rg '@SPEC:INSTALL-001' -n .moai/specs/

# 테스트 코드 확인
rg '@TEST:INSTALL-001' -n tests/

# 구현 코드 확인
rg '@CODE:INSTALL-001' -n src/

# 전체 TAG 체인 검증
rg '@(SPEC|TEST|CODE|DOC):INSTALL-001' -n
```

---

## 다음 단계

### 구현 단계 (우선순위 순)
1. **Phase 2 생성**: `phase2-developer.ts` + 테스트
2. **Phase 1 수정**: Git 검증 로직 추가
3. **Phase 3 수정**: SPEC Workflow 프롬프트 (Personal 모드)
4. **Phase 4 수정**: Auto PR/Draft PR 프롬프트 (Team 모드)
5. **Phase 5 생성**: `phase5-welcome.ts` (Alfred 환영 메시지)
6. **통합 테스트**: 전체 프롬프트 흐름 E2E 테스트

### 검증 단계
1. Personal 모드 + SPEC 활성화 시나리오
2. Personal 모드 + SPEC 비활성화 시나리오
3. Team 모드 + Auto PR 활성화 시나리오
4. Team 모드 + Auto PR 비활성화 시나리오
5. Git 미설치 에러 처리

### 동기화 단계
1. `/alfred:2-build SPEC-INSTALL-001` 실행 (TDD 구현)
2. `/alfred:3-sync` 실행 (문서 동기화 + TAG 체인 검증)
3. Living Document 생성

---

_이 문서는 SPEC-First TDD 방법론에 따라 작성되었습니다._
_Alfred 페르소나 강화를 통한 개발자 경험(DX) 개선이 목표입니다._
