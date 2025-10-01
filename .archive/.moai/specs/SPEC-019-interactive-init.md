# SPEC-019: 대화형 프로젝트 초기화 (Interactive Init)

## @TAG Catalog

| Chain | TAG | 설명 | 연관 산출물 |
|-------|-----|------|--------------|
| Primary | @SPEC:INTERACTIVE-INIT-019 | 대화형 초기화 요구사항 | SPEC-019 |
| Primary | @SPEC:INTERACTIVE-INIT-019 | 대화형 프롬프트 설계 | 본 문서 |
| Primary | @CODE:INTERACTIVE-INIT-019 | 대화형 로직 구현 작업 | src/cli/commands/init.ts |
| Primary | @TEST:INTERACTIVE-INIT-019 | 대화형 초기화 테스트 | __tests__/cli/commands/init-interactive.test.ts |
| Implementation | @CODE:INTERACTIVE-INIT-019 | 대화형 설치 기능 | src/cli/prompts/ |
| Quality | @DOC:INTERACTIVE-INIT-019 | 사용자 가이드 | docs/cli/interactive-init.md |

## @SPEC:INTERACTIVE-INIT-019 요구사항

### 배경 (Background)

현재 `moai init` 명령은 CLI 옵션(`--team`, `--force` 등)으로만 설정을 받아, 사용자가 모든 옵션을 미리 알아야 하는 문제가 있습니다. 대화형 방식으로 단계별로 선택을 받으면 사용자 경험이 크게 개선됩니다.

### 문제 (Problem)

1. **복잡한 CLI 옵션**: 사용자가 `--help`를 보고 옵션을 일일이 확인해야 함
2. **설정 누락**: 중요한 설정(GitHub URL, 원격 푸시 등)을 놓치기 쉬움
3. **모드별 차이 불명확**: personal과 team 모드의 차이를 사용자가 이해하기 어려움
4. **재설정 어려움**: 한번 설정하면 config.json 수동 편집 필요

### 목표 (Goals)

1. **대화형 설치 경험**: inquirer 기반 단계별 질문 방식
2. **모드별 맞춤 질문**: personal/team 모드에 따라 다른 질문 제시
3. **검증 및 가이드**: 입력값 검증 + 각 선택에 대한 설명 제공
4. **설정 저장**: 모든 선택을 `.moai/config.json`에 구조화하여 저장
5. **재사용 가능**: 저장된 설정으로 `moai update` 시 자동 적용

### 비목표 (Non-Goals)

- GUI 인터페이스 제공 (CLI만 지원)
- 모든 설정의 실시간 변경 (초기화 시점에만 설정)
- GitHub API를 통한 저장소 자동 생성 (수동 생성 전제)

### 제약사항 (Constraints)

- Node.js 18.0+ 환경에서만 동작
- inquirer 12.x 사용 (ESM 지원)
- 비대화형 모드(`--yes`, `--no-interactive`)도 지원 필요

## @SPEC:INTERACTIVE-INIT-019 설계

### EARS 요구사항

#### Ubiquitous Requirements (기본 기능)
- 시스템은 `moai init` 실행 시 대화형 프롬프트를 제공해야 한다
- 시스템은 사용자 입력을 검증하고 `.moai/config.json`에 저장해야 한다

#### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 `moai init`을 실행하면, 시스템은 프로젝트 설정 질문을 순차적으로 표시해야 한다
- WHEN personal 모드가 선택되면, 시스템은 로컬 Git 설정만 요청해야 한다
- WHEN team 모드가 선택되면, 시스템은 GitHub 관련 설정을 추가로 요청해야 한다
- WHEN 잘못된 입력이 주어지면, 시스템은 오류 메시지를 표시하고 재입력을 요청해야 한다

#### State-driven Requirements (상태 기반)
- WHILE 로컬 Git이 활성화되면, 시스템은 초기 커밋을 생성해야 한다
- WHILE GitHub가 활성화되면, 시스템은 원격 저장소 URL을 config.json에 저장해야 한다
- WHILE 원격 푸시가 활성화되면, 시스템은 초기 커밋을 원격으로 푸시해야 한다

#### Optional Features (선택적 기능)
- WHERE `--yes` 옵션이 제공되면, 시스템은 모든 질문에 기본값을 사용할 수 있다
- WHERE GitHub URL이 제공되면, 시스템은 저장소 유효성을 검증할 수 있다

#### Constraints (제약사항)
- IF 네트워크 연결이 없으면, 시스템은 GitHub 검증을 건너뛰어야 한다
- 대화형 프롬프트는 30초 이내에 완료 가능해야 한다

### 대화형 프롬프트 흐름

```
┌─────────────────────────────────────────────────┐
│ 🗿 MoAI-ADK Project Initialization             │
└─────────────────────────────────────────────────┘

❓ Question 1/7
→ Project name: (todo-app) _______________
  📝 This will be used as the folder name and project identifier

❓ Question 2/7
→ Select mode:
  ○ Personal - Local development with .moai/specs/
  ● Team     - GitHub Issues for SPEC management
  📝 Team mode enables GitHub integration and branch workflows

❓ Question 3/7
→ Initialize local Git repository?
  ● Yes
  ○ No
  📝 Required for version control and branch management

❓ Question 4/7 (Team mode only)
→ Use GitHub for remote repository?
  ● Yes
  ○ No
  📝 Enables GitHub Issues for SPEC tracking

❓ Question 5/7 (GitHub enabled)
→ GitHub repository URL: (https://github.com/user/repo) _______________
  📝 Example: https://github.com/username/project-name

❓ Question 6/7 (Team mode)
→ SPEC workflow:
  ● Branch + Merge (GitHub PR workflow)
  ○ Local commits only
  📝 Branch workflow creates feature branches per SPEC

❓ Question 7/7 (GitHub enabled)
→ Auto-push to remote repository?
  ● Yes
  ○ No
  📝 Automatically push commits to GitHub

✅ Configuration complete!
📦 Installing MoAI-ADK...
```

### config.json 스키마 확장

```typescript
interface MoAIConfig {
  version: string;
  mode: 'personal' | 'team';
  projectName: string;
  features: string[];

  // ✨ 새로운 필드들
  git: {
    enabled: boolean;              // 로컬 Git 사용 여부
    autoCommit: boolean;           // 자동 커밋 활성화
    branchPrefix: string;          // 브랜치 접두사 (예: feature/)
    remote?: {                     // 원격 저장소 설정 (선택)
      enabled: boolean;
      url: string;                 // GitHub URL
      autoPush: boolean;           // 자동 푸시 활성화
      defaultBranch: string;       // 기본 브랜치 (예: main)
    };
  };

  spec: {
    storage: 'local' | 'github';   // SPEC 저장 위치
    workflow: 'commit' | 'branch'; // 워크플로우 방식
    localPath: string;             // 로컬 저장 경로 (.moai/specs/)
    github?: {                     // GitHub 설정 (team 모드)
      issueLabels: string[];       // Issue 라벨
      templatePath: string;        // Issue 템플릿 경로
    };
  };

  backup: {
    enabled: boolean;
    retentionDays: number;
  };
}
```

### 구현 단계

#### Phase 1: 프롬프트 정의
```typescript
// src/cli/prompts/init-prompts.ts

export const initQuestions = [
  {
    type: 'input',
    name: 'projectName',
    message: 'Project name:',
    default: 'moai-project',
    validate: (input: string) => InputValidator.validateProjectName(input),
  },
  {
    type: 'list',
    name: 'mode',
    message: 'Select mode:',
    choices: [
      {
        name: 'Personal - Local development with .moai/specs/',
        value: 'personal',
      },
      {
        name: 'Team - GitHub Issues for SPEC management',
        value: 'team',
      },
    ],
    default: 'personal',
  },
  {
    type: 'confirm',
    name: 'gitEnabled',
    message: 'Initialize local Git repository?',
    default: true,
  },
  // ... 더 많은 질문들
];
```

#### Phase 2: 조건부 질문 로직
```typescript
// src/cli/prompts/conditional-prompts.ts

export function getConditionalQuestions(answers: Partial<InitAnswers>) {
  const questions = [];

  if (answers.mode === 'team' && answers.gitEnabled) {
    questions.push({
      type: 'confirm',
      name: 'githubEnabled',
      message: 'Use GitHub for remote repository?',
      default: true,
    });
  }

  if (answers.githubEnabled) {
    questions.push({
      type: 'input',
      name: 'githubUrl',
      message: 'GitHub repository URL:',
      validate: (input: string) => validateGitHubUrl(input),
    });
  }

  return questions;
}
```

#### Phase 3: 설정 빌더
```typescript
// src/cli/config/config-builder.ts

export class ConfigBuilder {
  buildConfig(answers: InitAnswers): MoAIConfig {
    return {
      version: '0.0.1',
      mode: answers.mode,
      projectName: answers.projectName,
      features: [],

      git: {
        enabled: answers.gitEnabled,
        autoCommit: true,
        branchPrefix: answers.mode === 'team' ? 'feature/' : '',
        ...(answers.githubEnabled && {
          remote: {
            enabled: true,
            url: answers.githubUrl,
            autoPush: answers.autoPush,
            defaultBranch: 'main',
          },
        }),
      },

      spec: {
        storage: answers.mode === 'team' ? 'github' : 'local',
        workflow: answers.specWorkflow || 'commit',
        localPath: '.moai/specs/',
        ...(answers.mode === 'team' && {
          github: {
            issueLabels: ['spec', 'requirements'],
            templatePath: '.github/ISSUE_TEMPLATE/spec.md',
          },
        }),
      },

      backup: {
        enabled: true,
        retentionDays: 30,
      },
    };
  }
}
```

### 모드별 동작 차이

| 기능 | Personal 모드 | Team 모드 |
|------|---------------|-----------|
| **SPEC 저장** | `.moai/specs/*.md` | GitHub Issues |
| **Git 워크플로우** | 로컬 커밋 | 브랜치 + PR |
| **원격 푸시** | 선택적 | 권장 (자동) |
| **협업** | 단독 | 팀 (GitHub 기반) |
| **Issue 템플릿** | 없음 | `.github/ISSUE_TEMPLATE/` |

## @CODE:INTERACTIVE-INIT-019 구현 계획

### Task 1: inquirer 통합
- [ ] inquirer 12.x 설치 및 설정
- [ ] TypeScript 타입 정의 추가
- [ ] 기본 프롬프트 플로우 구현

### Task 2: 검증 로직
- [ ] 프로젝트 이름 검증 (InputValidator 확장)
- [ ] GitHub URL 검증 (정규식 + 선택적 API 체크)
- [ ] 경로 검증 (기존 프로젝트 충돌 확인)

### Task 3: 설정 빌더
- [ ] ConfigBuilder 클래스 구현
- [ ] 모드별 기본값 설정
- [ ] config.json 저장 로직

### Task 4: 비대화형 모드
- [ ] `--yes` 옵션 구현 (모든 기본값 사용)
- [ ] `--no-interactive` 옵션 구현
- [ ] CLI 옵션과 대화형 프롬프트 통합

### Task 5: 테스트
- [ ] 대화형 프롬프트 단위 테스트
- [ ] 모드별 통합 테스트
- [ ] 검증 로직 테스트

## @TEST:INTERACTIVE-INIT-019 테스트 시나리오

### 시나리오 1: Personal 모드 + 로컬 Git
```bash
$ moai init

→ Project name: my-app
→ Mode: Personal
→ Initialize Git? Yes
→ Auto-commit? Yes

✅ Created: my-app/
  - .moai/config.json (git.enabled=true, spec.storage=local)
  - .moai/specs/ (빈 폴더)
  - .git/ (초기 커밋 포함)
```

### 시나리오 2: Team 모드 + GitHub
```bash
$ moai init

→ Project name: team-project
→ Mode: Team
→ Initialize Git? Yes
→ Use GitHub? Yes
→ GitHub URL: https://github.com/user/team-project
→ SPEC workflow: Branch + Merge
→ Auto-push? Yes

✅ Created: team-project/
  - .moai/config.json (git.remote.enabled=true, spec.storage=github)
  - .github/ISSUE_TEMPLATE/spec.md
  - .git/ (초기 커밋 + 원격 푸시됨)
```

### 시나리오 3: 비대화형 모드
```bash
$ moai init --yes --mode=team --github-url=https://github.com/user/repo

✅ Created with defaults (no prompts)
```

## 예상 효과

### 사용자 경험 개선
- ✅ 옵션 암기 불필요
- ✅ 단계별 가이드 제공
- ✅ 실수 방지 (검증 로직)
- ✅ 재설정 용이

### 개발 생산성
- ✅ 프로젝트 설정 시간 단축
- ✅ 모드별 최적 설정 자동 적용
- ✅ GitHub 통합 간소화

## 다음 단계

1. `/moai:2-build SPEC-019` 실행하여 TDD 구현
2. Personal/Team 모드별 문서 작성
3. 기존 `moai init` 사용자에게 마이그레이션 가이드 제공

---

_이 SPEC은 v0.1.0에서 구현 예정입니다._