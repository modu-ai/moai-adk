# SPEC-INSTALL-001 구현 계획

## 1. 개요

### 목표
설치 프롬프트 재설계를 통해 개발자 경험(DX) 개선 및 Git/SPEC Workflow 필수화

### 범위
- **Phase 재구성**: 7개 Phase → 5개 Phase로 재조정
- **신규 Phase**: Phase 2 (개발자 정보), Phase 5 (환영 메시지)
- **수정 Phase**: Phase 1 (Git 검증), Phase 3 (SPEC 선택), Phase 4 (Auto/Draft PR)
- **테스트**: 단위 테스트 + 통합 테스트 (TDD 방식)

### 의존성
- **Inquirer.js**: 대화형 프롬프트 라이브러리
- **Chalk**: 콘솔 색상 출력
- **Commander.js**: CLI 명령어 파싱
- **fs-extra**: 파일 시스템 조작

---

## 2. 아키텍처 설계

### 2.1 프롬프트 흐름 재구성

#### 기존 흐름 (v0.0.x)
```
Phase 1: 모드 선택 (Personal/Team)
  ↓
Phase 2: Git 전략 설정 (하드코딩)
  ↓
Phase 3: 태그 설정
  ↓
Phase 4: 파이프라인 설정
```

**문제점**:
- Git 필수 검증 누락
- 개발자 이름 미수집 (커밋 서명 불완전)
- Auto PR/Draft PR 하드코딩 (사용자 선택권 부재)
- SPEC Workflow 강제 활성화 (Personal 모드 유연성 부족)

#### 신규 흐름 (v0.1.0, 본 SPEC)
```
Phase 1: Git 검증 + 모드 선택
  ↓
Phase 2: 개발자 정보 수집 (NEW)
  ↓
Phase 3: SPEC Workflow 선택 (Personal 모드만, NEW)
  ↓
Phase 4: Auto PR/Draft PR 선택 (Team 모드만, NEW)
  ↓
Phase 5: Alfred 환영 메시지 (NEW)
```

**개선점**:
- Git 필수 검증 사전 실행
- 개발자 이름 수집 → 커밋 서명 완전성
- Auto PR/Draft PR 사용자 선택
- Progressive Disclosure 적용 (인지 부담 최소화)

### 2.2 Phase별 책임 분리

| Phase | 파일명 | 역할 | 주요 함수 | 라인 수 |
|-------|--------|------|----------|--------|
| **Phase 1** | `phase1-basic.ts` | Git 검증 + 모드 선택 | `validateGitInstallation()` | ~80줄 |
| **Phase 2** | `phase2-developer.ts` | 개발자 정보 수집 | `collectDeveloperInfo()` | ~50줄 |
| **Phase 3** | `phase3-mode.ts` | SPEC Workflow 선택 | `selectSpecWorkflow()` | ~60줄 |
| **Phase 4** | `phase4-git.ts` | Auto PR/Draft PR 선택 | `configurePRStrategy()` | ~70줄 |
| **Phase 5** | `phase5-welcome.ts` | 환영 메시지 출력 | `displayWelcomeMessage()` | ~30줄 |
| **통합** | `index.ts` | Phase 조율 | `executeInstallPrompts()` | ~100줄 |

### 2.3 데이터 흐름 (Context 객체)

```typescript
interface InstallContext {
  // Phase 1 출력
  mode: 'personal' | 'team';
  gitVersion: string;

  // Phase 2 출력
  developerName: string;
  gitUserName: string | null;

  // Phase 3 출력 (Personal 모드만)
  enforceSpec?: boolean;

  // Phase 4 출력 (Team 모드만)
  autoPR?: boolean;
  draftPR?: boolean;

  // 최종 config.json 생성
  config: Config;
}
```

**흐름**:
1. Phase 1 → `context.mode`, `context.gitVersion` 생성
2. Phase 2 → `context.developerName` 추가
3. Phase 3/4 → 모드별 옵션 추가
4. Phase 5 → `context.config` 생성 후 저장

### 2.4 에러 처리 전략

#### Git 미설치 (Phase 1)
```typescript
try {
  const gitVersion = await execCommand('git --version');
  context.gitVersion = gitVersion;
} catch (error) {
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

#### 개발자 이름 미입력 (Phase 2)
```typescript
{
  type: 'input',
  name: 'developerName',
  message: '개발자 이름을 입력해주세요:',
  validate: (input) => {
    if (input.trim() === '') {
      return '❌ 개발자 이름은 필수입니다.';
    }
    return true;
  }
}
```

#### config.json 덮어쓰기 경고
```typescript
if (fs.existsSync('.moai/config.json')) {
  const { overwrite } = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'overwrite',
      message: '⚠️  기존 설정을 덮어쓸까요? (데이터 손실 주의)',
      default: false
    }
  ]);

  if (!overwrite) {
    console.log('설치를 중단합니다.');
    process.exit(0);
  }

  // 백업 생성
  fs.copyFileSync('.moai/config.json', '.moai/config.json.backup');
}
```

---

## 3. 구현 우선순위

### 1차 목표: Phase 2 개발자 정보 수집 (핵심 기능)
**목표**: Git 커밋 서명 완전성 확보

**작업**:
- `phase2-developer.ts` 생성
- Git `user.name` 조회 로직
- Inquirer 프롬프트 구현
- config.json에 `developer.name` 저장

**수락 기준**:
- Git user.name을 기본값으로 제안
- 빈 값 입력 시 에러 메시지 출력
- config.json에 저장 확인

### 2차 목표: Phase 1 Git 검증 추가
**목표**: Git 미설치 사용자 조기 차단

**작업**:
- `validateGitInstallation()` 함수 추가
- `git --version` 실행 및 파싱
- 에러 메시지 (macOS/Ubuntu/Windows 안내)

**수락 기준**:
- Git 미설치 시 친절한 에러 메시지 출력
- 설치 안내 후 `process.exit(1)`
- Git 설치 시 버전 정보 캐싱

### 3차 목표: Phase 3/4 모드별 프롬프트 추가
**목표**: SPEC Workflow, Auto PR, Draft PR 사용자 선택

**작업**:
- Phase 3: `selectSpecWorkflow()` (Personal 모드만)
- Phase 4: `configurePRStrategy()` (Team 모드만)
- 조건부 프롬프트 로직 (Progressive Disclosure)

**수락 기준**:
- Personal 모드는 SPEC Workflow 선택 프롬프트 표시
- Team 모드는 Auto PR/Draft PR 선택 프롬프트 표시
- Draft PR은 Auto PR 활성화 시에만 표시

### 4차 목표: Phase 5 Alfred 환영 메시지
**목표**: Alfred 페르소나 강화 및 다음 단계 안내

**작업**:
- `displayWelcomeMessage()` 함수 생성
- 개발자 이름 포함 메시지
- 다음 명령어 안내 (/alfred:8-project, /alfred:1-spec)

**수락 기준**:
- "🤖 AI-Agent Alfred가 {name}님의 개발을 도와드리겠습니다" 출력
- 다음 단계 명령어 표시
- 이모지 적절히 사용

---

## 4. 기술적 접근 방법

### 4.1 Phase별 구현 상세

#### Phase 1: Git 검증 + 모드 선택
**파일**: `src/cli/prompts/init/phase1-basic.ts`

**기존 코드** (예상):
```typescript
export async function executePhase1(): Promise<Phase1Result> {
  const { mode } = await inquirer.prompt([
    {
      type: 'list',
      name: 'mode',
      message: '프로젝트 모드를 선택하세요:',
      choices: [
        { name: 'Personal - 혼자 개발하는 프로젝트', value: 'personal' },
        { name: 'Team - 팀으로 협업하는 프로젝트', value: 'team' }
      ]
    }
  ]);

  return { mode };
}
```

**신규 코드** (Git 검증 추가):
```typescript
// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: tests/cli/prompts/init/phase1-basic.test.ts
import { execCommand } from '../../../utils/exec';

async function validateGitInstallation(): Promise<string> {
  try {
    const gitVersion = await execCommand('git --version');
    return gitVersion.trim();
  } catch (error) {
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
}

export async function executePhase1(): Promise<Phase1Result> {
  // Git 검증 (사전 실행)
  const gitVersion = await validateGitInstallation();
  console.log(`✅ Git 버전: ${gitVersion}`);

  // 모드 선택
  const { mode } = await inquirer.prompt([
    {
      type: 'list',
      name: 'mode',
      message: '프로젝트 모드를 선택하세요:',
      choices: [
        { name: 'Personal - 혼자 개발하는 프로젝트', value: 'personal' },
        { name: 'Team - 팀으로 협업하는 프로젝트', value: 'team' }
      ]
    }
  ]);

  return { mode, gitVersion };
}
```

**변경 라인**: ~30줄 추가

---

#### Phase 2: 개발자 정보 수집 (NEW)
**파일**: `src/cli/prompts/init/phase2-developer.ts`

**신규 코드**:
```typescript
// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: tests/cli/prompts/init/phase2-developer.test.ts
import { execCommand } from '../../../utils/exec';
import inquirer from 'inquirer';

export interface Phase2Result {
  developerName: string;
  gitUserName: string | null;
}

async function getGitUserName(): Promise<string | null> {
  try {
    const userName = await execCommand('git config --global user.name');
    return userName.trim() || null;
  } catch {
    return null;
  }
}

export async function executePhase2(): Promise<Phase2Result> {
  const gitUserName = await getGitUserName();

  const { developerName } = await inquirer.prompt([
    {
      type: 'input',
      name: 'developerName',
      message: '개발자 이름을 입력해주세요:',
      default: gitUserName || '',
      validate: (input) => {
        if (input.trim() === '') {
          return '❌ 개발자 이름은 필수입니다.';
        }
        return true;
      }
    }
  ]);

  return { developerName, gitUserName };
}
```

**라인 수**: ~50줄

---

#### Phase 3: SPEC Workflow 선택 (Personal 모드만)
**파일**: `src/cli/prompts/init/phase3-mode.ts`

**기존 코드** (예상):
```typescript
export async function executePhase3(context: InstallContext): Promise<Phase3Result> {
  // 기존 로직 (태그 설정 등)
}
```

**신규 코드** (SPEC 선택 추가):
```typescript
// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: tests/cli/prompts/init/phase3-mode.test.ts
export async function executePhase3(context: InstallContext): Promise<Phase3Result> {
  let enforceSpec = true; // 기본값

  if (context.mode === 'personal') {
    const { useSpec } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'useSpec',
        message: 'SPEC-First Workflow를 사용할까요? (권장)',
        default: true
      }
    ]);

    enforceSpec = useSpec;
  } else {
    // Team 모드는 SPEC 강제 활성화
    enforceSpec = true;
    console.log('ℹ️  Team 모드는 SPEC-First Workflow가 필수입니다.');
  }

  return { enforceSpec };
}
```

**변경 라인**: ~20줄 추가

---

#### Phase 4: Auto PR/Draft PR 선택 (Team 모드만)
**파일**: `src/cli/prompts/init/phase4-git.ts`

**기존 코드** (예상):
```typescript
export async function executePhase4(context: InstallContext): Promise<Phase4Result> {
  // 기존 Git 전략 하드코딩
  return {
    autoPR: true,
    draftPR: true
  };
}
```

**신규 코드** (사용자 선택 추가):
```typescript
// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: tests/cli/prompts/init/phase4-git.test.ts
export async function executePhase4(context: InstallContext): Promise<Phase4Result> {
  if (context.mode === 'personal') {
    // Personal 모드는 Auto PR/Draft PR 미사용
    return { autoPR: false, draftPR: false };
  }

  // Team 모드: Auto PR 선택
  const { autoPR } = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'autoPR',
      message: '자동으로 PR을 생성할까요? (Auto PR)',
      default: true
    }
  ]);

  let draftPR = true; // 기본값

  if (autoPR) {
    // Auto PR 활성화 시에만 Draft PR 선택 제공
    const { useDraft } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'useDraft',
        message: 'PR을 Draft 상태로 생성할까요? (검토 후 Ready 전환)',
        default: true
      }
    ]);

    draftPR = useDraft;
  } else {
    // Auto PR 비활성화 시 Draft PR 무의미
    draftPR = false;
  }

  return { autoPR, draftPR };
}
```

**변경 라인**: ~40줄 추가

---

#### Phase 5: Alfred 환영 메시지 (NEW)
**파일**: `src/cli/prompts/init/phase5-welcome.ts`

**신규 코드**:
```typescript
// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: tests/cli/prompts/init/phase5-welcome.test.ts
import chalk from 'chalk';

export interface Phase5Input {
  developerName: string;
  mode: 'personal' | 'team';
}

export function displayWelcomeMessage(input: Phase5Input): void {
  console.log(chalk.green(`
✅ MoAI-ADK 설치가 완료되었습니다!

🤖 AI-Agent Alfred가 ${input.developerName}님의 개발을 도와드리겠습니다.

다음 명령어로 시작하세요:
  ${chalk.cyan('/alfred:8-project')}  # 프로젝트 초기화
  ${chalk.cyan('/alfred:1-spec')}     # 첫 SPEC 작성

질문이 있으시면 언제든 ${chalk.cyan('@agent-debug-helper')}를 호출하세요.
  `));
}
```

**라인 수**: ~30줄

---

### 4.2 통합 로직 (index.ts)

**파일**: `src/cli/prompts/init/index.ts`

**신규 코드**:
```typescript
// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md
import { executePhase1 } from './phase1-basic';
import { executePhase2 } from './phase2-developer';
import { executePhase3 } from './phase3-mode';
import { executePhase4 } from './phase4-git';
import { displayWelcomeMessage } from './phase5-welcome';
import { Config } from '../../../types/config';

export async function executeInstallPrompts(): Promise<Config> {
  const context: InstallContext = {};

  // Phase 1: Git 검증 + 모드 선택
  const phase1 = await executePhase1();
  context.mode = phase1.mode;
  context.gitVersion = phase1.gitVersion;

  // Phase 2: 개발자 정보 수집
  const phase2 = await executePhase2();
  context.developerName = phase2.developerName;

  // Phase 3: SPEC Workflow 선택 (Personal 모드만)
  const phase3 = await executePhase3(context);
  context.enforceSpec = phase3.enforceSpec;

  // Phase 4: Auto PR/Draft PR 선택 (Team 모드만)
  const phase4 = await executePhase4(context);
  context.autoPR = phase4.autoPR;
  context.draftPR = phase4.draftPR;

  // config.json 생성
  const config: Config = buildConfig(context);

  // config.json 저장
  await saveConfig(config);

  // Phase 5: Alfred 환영 메시지
  displayWelcomeMessage({
    developerName: context.developerName,
    mode: context.mode
  });

  return config;
}

function buildConfig(context: InstallContext): Config {
  return {
    project: {
      name: 'MoAI-ADK',
      mode: context.mode,
      locale: 'ko'
    },
    developer: {
      name: context.developerName,
      timestamp: new Date().toISOString()
    },
    constitution: {
      enforce_tdd: true,
      enforce_spec: context.enforceSpec ?? true,
      require_tags: true
    },
    git_strategy: {
      personal: {
        auto_checkpoint: true,
        auto_commit: true
      },
      team: {
        auto_pr: context.autoPR ?? false,
        draft_pr: context.draftPR ?? false,
        use_gitflow: true
      }
    }
  };
}
```

**변경 라인**: ~50줄 수정

---

### 4.3 테스트 전략 (TDD)

#### 단위 테스트 예시 (Phase 2)
**파일**: `tests/cli/prompts/init/phase2-developer.test.ts`

```typescript
// @TEST:INSTALL-001 | SPEC: SPEC-INSTALL-001.md
import { executePhase2 } from '../../../../src/cli/prompts/init/phase2-developer';
import { execCommand } from '../../../../src/utils/exec';
import inquirer from 'inquirer';

jest.mock('../../../../src/utils/exec');
jest.mock('inquirer');

describe('Phase 2: 개발자 정보 수집', () => {
  it('Git user.name을 기본값으로 제안해야 한다', async () => {
    // Given
    (execCommand as jest.Mock).mockResolvedValue('홍길동');
    (inquirer.prompt as jest.Mock).mockResolvedValue({ developerName: '홍길동' });

    // When
    const result = await executePhase2();

    // Then
    expect(result.developerName).toBe('홍길동');
    expect(result.gitUserName).toBe('홍길동');
    expect(inquirer.prompt).toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({ default: '홍길동' })
      ])
    );
  });

  it('Git user.name이 없으면 빈 문자열을 기본값으로 사용해야 한다', async () => {
    // Given
    (execCommand as jest.Mock).mockRejectedValue(new Error('Not found'));
    (inquirer.prompt as jest.Mock).mockResolvedValue({ developerName: '김철수' });

    // When
    const result = await executePhase2();

    // Then
    expect(result.gitUserName).toBeNull();
    expect(inquirer.prompt).toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({ default: '' })
      ])
    );
  });

  it('빈 값 입력 시 에러 메시지를 반환해야 한다', async () => {
    // Given
    const promptConfig = (inquirer.prompt as jest.Mock).mock.calls[0]?.[0]?.[0];
    const validate = promptConfig?.validate;

    // When
    const result = validate('   ');

    // Then
    expect(result).toBe('❌ 개발자 이름은 필수입니다.');
  });
});
```

#### 통합 테스트 예시
**파일**: `tests/core/installer/integration.test.ts`

```typescript
// @TEST:INSTALL-001 | SPEC: SPEC-INSTALL-001.md
describe('설치 프롬프트 통합 테스트', () => {
  it('Personal 모드 + SPEC 활성화 시나리오', async () => {
    // Given
    mockInquirerPrompts({
      mode: 'personal',
      developerName: '홍길동',
      enforceSpec: true
    });

    // When
    const config = await executeInstallPrompts();

    // Then
    expect(config.project.mode).toBe('personal');
    expect(config.developer.name).toBe('홍길동');
    expect(config.constitution.enforce_spec).toBe(true);
    expect(config.git_strategy.team.auto_pr).toBe(false); // Personal 모드는 비활성화
  });

  it('Team 모드 + Auto PR 비활성화 시나리오', async () => {
    // Given
    mockInquirerPrompts({
      mode: 'team',
      developerName: '김철수',
      autoPR: false
    });

    // When
    const config = await executeInstallPrompts();

    // Then
    expect(config.project.mode).toBe('team');
    expect(config.constitution.enforce_spec).toBe(true); // Team 모드는 강제 활성화
    expect(config.git_strategy.team.auto_pr).toBe(false);
    expect(config.git_strategy.team.draft_pr).toBe(false); // Auto PR 비활성화 시 무의미
  });

  it('Git 미설치 시 에러 처리', async () => {
    // Given
    (execCommand as jest.Mock).mockRejectedValue(new Error('Command not found'));

    // When & Then
    await expect(executeInstallPrompts()).rejects.toThrow();
    expect(console.error).toHaveBeenCalledWith(
      expect.stringContaining('❌ Git이 설치되지 않았습니다')
    );
  });
});
```

---

## 5. 리스크 및 대응 방안

### 리스크 1: Git 전역 설정 읽기 실패
**원인**: Git이 설치되었지만 `user.name` 미설정

**영향**: 개발자 이름 프롬프트 기본값 없음

**대응 방안**:
- `getGitUserName()` 함수에서 `try-catch` 사용
- 실패 시 `null` 반환 → 빈 문자열 기본값
- 사용자 입력 강제 (validate 함수로 빈 값 차단)

### 리스크 2: 기존 프로젝트 마이그레이션
**원인**: v0.0.x → v0.1.0 업그레이드 시 `developer` 필드 누락

**영향**: Git 커밋 서명 불완전

**대응 방안**:
- Config 검증 함수에서 경고만 출력 (설치 차단 X)
- 향후 `/alfred:upgrade` 명령 제공 (수동 마이그레이션)
- `developer` 필드 누락 시 Git `user.name` 폴백

### 리스크 3: Progressive Disclosure 미준수
**원인**: 모든 프롬프트를 한 번에 표시

**영향**: 인지 부담 증가 → 사용자 이탈

**대응 방안**:
- Phase 단위로 순차 실행 (조건부 로직 엄격 적용)
- Draft PR 프롬프트는 Auto PR 활성화 시에만 표시
- 각 Phase마다 콘솔 출력으로 진행 상황 표시

---

## 6. 타임라인

### 구현 단계별 작업 순서

#### 1단계: Phase 2 (개발자 정보)
**작업**:
- `phase2-developer.ts` 생성
- `getGitUserName()` 함수 구현
- Inquirer 프롬프트 구현
- 단위 테스트 작성

#### 2단계: Phase 1 (Git 검증)
**작업**:
- `validateGitInstallation()` 함수 추가
- 에러 메시지 구현
- 버전 캐싱 로직

#### 3단계: Phase 3/4 (모드별 프롬프트)
**작업**:
- Phase 3: SPEC Workflow 선택 (Personal 모드)
- Phase 4: Auto PR/Draft PR 선택 (Team 모드)
- 조건부 로직 구현

#### 4단계: Phase 5 (환영 메시지)
**작업**:
- `displayWelcomeMessage()` 함수 생성
- Chalk 색상 적용
- 다음 단계 안내

#### 5단계: 통합 및 테스트
**작업**:
- `index.ts` Phase 조율 로직
- `phase-executor.ts` 실행기 업데이트
- 통합 테스트 (E2E 시나리오)

#### 6단계: 문서 동기화
**작업**:
- `/alfred:3-sync` 실행
- TAG 체인 검증
- Living Document 생성

---

## 다음 단계

### 즉시 시작
1. `/alfred:2-build SPEC-INSTALL-001` 실행 (TDD 구현 시작)
2. Phase 2 (개발자 정보) 우선 구현

### 검증 준비
1. Personal 모드 + SPEC 활성화 시나리오 테스트
2. Team 모드 + Auto PR 비활성화 시나리오 테스트
3. Git 미설치 에러 처리 확인

### 마무리
1. `/alfred:3-sync` 실행 (문서 동기화)
2. 통합 테스트 결과 확인
3. Living Document 생성 및 검토

---

_이 계획서는 SPEC-INSTALL-001 구현의 청사진입니다._
_우선순위 기반 마일스톤으로 효율적인 개발을 목표로 합니다._
