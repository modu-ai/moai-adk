# SPEC-012 TDD 구현 계획: TypeScript 기반 구축 (Week 1)

> **@TASK:IMPLEMENTATION-PLAN-012** SPEC-012 TypeScript 기반 구축을 위한 TDD 실행 계획
> **@DESIGN:TDD-STRATEGY-012** Red-Green-Refactor 사이클 기반 단계별 접근
> **@PERF:MILESTONE-012** Week 1 Day-by-Day 마일스톤 및 성과 대상

---

## 구현 전략

### TDD 사이클 적용

```
Red → Green → Refactor → Integration
 │      │        │            │
 │      │        │            └─ CI/CD 통합
 │      │        └───────────── 코드 품질 개선
 │      └─────────────────────── 최소 구현
 └───────────────────────────── 실패하는 테스트 작성
```

### 우선순위 및 의존성

1. **Day 1-2**: 프로젝트 초기화 (기반 작업)
2. **Day 3-4**: 시스템 검증 모듈 (핵심 기능)
3. **Day 5-6**: CLI 구조 및 명령어 (사용자 인터페이스)
4. **Day 7**: 통합 테스트 및 검증 (전체 연동)

---

## Day-by-Day 마일스톤

### Day 1: 프로젝트 초기화 (RED 단계)

#### 작업 목표
- npm 프로젝트 생성 및 TypeScript 환경 설정
- 기본 빌드 파이프라인 구축
- 개발 도구 체인 설정

#### TDD 사이클
```typescript
// __tests__/setup.test.ts - RED
describe('Project Setup', () => {
  test('should have valid package.json', () => {
    const pkg = require('../package.json');
    expect(pkg.name).toBe('moai-adk');
    expect(pkg.version).toBe('0.0.1');
  });

  test('should compile TypeScript successfully', () => {
    // 이 테스트는 초기에 실패해야 함
    expect(() => require('../dist/index.js')).not.toThrow();
  });
});
```

#### 예상 작업
1. **프로젝트 생성**:
   ```bash
   mkdir moai-adk && cd moai-adk
   npm init -y
   ```

2. **TypeScript 및 빌드 도구 설치**:
   ```bash
   npm install -D typescript @types/node tsup jest ts-jest @types/jest
   npm install -D eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin
   npm install -D prettier
   ```

3. **런타임 의존성 설치**:
   ```bash
   npm install commander chalk inquirer fs-extra semver ora which
   ```

#### 성공 기준
- [ ] package.json에 올바른 메타데이터 설정
- [ ] tsconfig.json 구성 완료
- [ ] 기본 빌드 스크립트 동작
- [ ] ESLint/Prettier 설정 완료

---

### Day 2: 설정 완성 및 기본 구조 (GREEN 단계)

#### 작업 목표
- Day 1의 실패 테스트를 통과시키는 최소 구현
- 기본 디렉토리 구조 생성
- 기본 TypeScript 대상 파일 생성

#### TDD 사이클
```typescript
// src/index.ts - GREEN
export const version = '0.0.1';
export const description = 'MoAI-ADK: Modu-AI Agentic Development kit';

// src/cli/index.ts - GREEN
import { Command } from 'commander';

const program = new Command();
program
  .name('moai')
  .description('🗿 MoAI-ADK: Modu-AI Agentic Development kit')
  .version('0.0.1');

program.parse();
```

#### 디렉토리 구조 생성
```
src/
├── cli/
│   └── index.ts
├── core/
│   └── system-checker/
│       └── index.ts
├── utils/
│   ├── logger.ts
│   └── version.ts
└── index.ts
```

#### 성공 기준
- [ ] `npm run build` 성공
- [ ] `npm test` 통과
- [ ] `node dist/cli/index.js --version` 동작
- [ ] ESLint/Prettier 검사 통과

---

### Day 3: 시스템 검증 모듈 TDD (RED 단계)

#### 작업 목표
- 시스템 요구사항 자동 검증 모듈의 실패 테스트 작성
- 핵심 인터페이스 설계
- 테스트 시나리오 완성

#### TDD 사이클
```typescript
// __tests__/system-checker/detector.test.ts - RED
describe('SystemDetector', () => {
  let detector: SystemDetector;

  beforeEach(() => {
    detector = new SystemDetector();
  });

  test('should detect installed Node.js', async () => {
    const requirement: SystemRequirement = {
      name: 'Node.js',
      category: 'runtime',
      minVersion: '18.0.0',
      installCommands: { darwin: 'brew install node' },
      checkCommand: 'node --version'
    };

    const result = await detector.checkRequirement(requirement);

    expect(result.installed).toBe(true);
    expect(result.satisfied).toBe(true);
    expect(result.version).toMatch(/\d+\.\d+\.\d+/);
  });

  test('should handle missing tool gracefully', async () => {
    const requirement: SystemRequirement = {
      name: 'NonExistentTool',
      category: 'runtime',
      installCommands: { darwin: 'brew install nonexistent' },
      checkCommand: 'nonexistent --version'
    };

    const result = await detector.checkRequirement(requirement);

    expect(result.installed).toBe(false);
    expect(result.satisfied).toBe(false);
    expect(result.error).toBeDefined();
  });

  test('should validate version requirements', async () => {
    // 이 테스트는 초기에 실패해야 함
    const result = await detector.checkRequirement({
      name: 'Node.js',
      category: 'runtime',
      minVersion: '20.0.0', // 높은 버전 요구
      installCommands: { darwin: 'brew install node' },
      checkCommand: 'node --version'
    });

    expect(result.satisfied).toBe(false); // 이 부분이 실패해야 함
  });
});
```

#### 인터페이스 설계
```typescript
// src/core/system-checker/types.ts
export interface SystemRequirement {
  name: string;
  category: 'runtime' | 'development' | 'optional';
  minVersion?: string;
  installCommands: Record<string, string>;
  checkCommand: string;
  versionCommand?: string;
}

export interface RequirementStatus {
  name: string;
  installed: boolean;
  version?: string;
  satisfied: boolean;
  required?: string;
  error?: string;
}
```

#### 성공 기준
- [ ] 모든 테스트가 예상대로 실패
- [ ] 인터페이스 설계 완료
- [ ] 테스트 커버리지 100% (빈 구현체)

---

### Day 4: 시스템 검증 모듈 구현 (GREEN 단계)

#### 작업 목표
- Day 3의 실패 테스트를 통과시키는 최소 구현
- SystemDetector 클래스 구현
- 버전 추출 로직 구현

#### TDD 사이클
```typescript
// src/core/system-checker/detector.ts - GREEN
import { exec } from 'child_process';
import { promisify } from 'util';
import * as semver from 'semver';

const execAsync = promisify(exec);

export class SystemDetector {
  async checkRequirement(req: SystemRequirement): Promise<RequirementStatus> {
    try {
      const { stdout } = await execAsync(req.checkCommand);
      const version = this.extractVersion(stdout, req.name);

      if (req.minVersion && version) {
        const satisfied = semver.gte(version, req.minVersion);
        return {
          name: req.name,
          installed: true,
          version,
          satisfied,
          required: req.minVersion
        };
      }

      return {
        name: req.name,
        installed: true,
        version,
        satisfied: true
      };
    } catch (error) {
      return {
        name: req.name,
        installed: false,
        satisfied: false,
        error: error.message
      };
    }
  }

  private extractVersion(output: string, toolName: string): string | null {
    const patterns = {
      'Node.js': /v(\d+\.\d+\.\d+)/,
      'Git': /git version (\d+\.\d+\.\d+)/,
      'SQLite3': /(\d+\.\d+\.\d+)/
    };

    const pattern = patterns[toolName];
    if (pattern) {
      const match = output.match(pattern);
      return match ? match[1] : null;
    }
    return null;
  }
}
```

#### 성공 기준
- [ ] 모든 시스템 검증 테스트 통과
- [ ] Node.js/Git 감지 정상 동작
- [ ] 버전 비교 로직 정상 동작
- [ ] 에러 처리 정상 동작

---

### Day 5: CLI 명령어 TDD (RED 단계)

#### 작업 목표
- CLI 명령어 구조의 실패 테스트 작성
- moai init, moai doctor 명령어 인터페이스 설계
- E2E 테스트 시나리오 작성

#### TDD 사이클
```typescript
// __tests__/cli/commands.test.ts - RED
describe('CLI Commands', () => {
  test('moai --version should return version', async () => {
    const { execSync } = require('child_process');
    const output = execSync('node dist/cli/index.js --version', { encoding: 'utf8' });

    expect(output.trim()).toBe('0.0.1');
  });

  test('moai --help should show help', async () => {
    const { execSync } = require('child_process');
    const output = execSync('node dist/cli/index.js --help', { encoding: 'utf8' });

    expect(output).toContain('MoAI-ADK: Modu-AI Agentic Development kit');
    expect(output).toContain('init');
    expect(output).toContain('doctor');
  });

  test('moai init should perform system check', async () => {
    // 이 테스트는 초기에 실패해야 함
    const { execSync } = require('child_process');
    const output = execSync('node dist/cli/index.js init test-project', { encoding: 'utf8' });

    expect(output).toContain('System Requirements Check');
    expect(output).toContain('Project Setup');
  });

  test('moai doctor should show system status', async () => {
    // 이 테스트는 초기에 실패해야 함
    const { execSync } = require('child_process');
    const output = execSync('node dist/cli/index.js doctor', { encoding: 'utf8' });

    expect(output).toContain('System Diagnosis');
    expect(output).toContain('Node.js');
  });
});
```

#### 성공 기준
- [ ] CLI 테스트 예상대로 실패
- [ ] 명령어 인터페이스 설계 완료
- [ ] E2E 테스트 시나리오 완성

---

### Day 6: CLI 명령어 구현 (GREEN 단계)

#### 작업 목표
- Day 5의 실패 테스트를 통과시키는 최소 구현
- moai init, moai doctor 명령어 구현
- 시스템 검증 모듈과 연동

#### TDD 사이클
```typescript
// src/cli/commands/init.ts - GREEN
import chalk from 'chalk';
import { SystemChecker } from '../../core/system-checker';

export async function initCommand(projectName: string): Promise<void> {
  console.log(chalk.cyan('🗿 MoAI-ADK Project Initialization'));
  console.log(chalk.cyan('================================'));

  // Step 1: 시스템 요구사항 검증
  console.log('\n🔍 Step 1: System Requirements Check');
  const systemChecker = new SystemChecker();
  const missingRequirements = await systemChecker.checkAll();

  if (missingRequirements.length > 0) {
    console.log(chalk.yellow('Missing requirements detected'));
    // 자동 설치 로직은 Week 2에서 구현
  }

  // Step 2: 프로젝트 구조 생성 (향후 Week 2에서 구현)
  console.log('\n🚀 Step 2: Project Setup (Coming in Week 2)');
  console.log(chalk.green(`✅ Project "${projectName}" foundation ready!`));
}

// src/cli/commands/doctor.ts - GREEN
import chalk from 'chalk';
import { SystemChecker } from '../../core/system-checker';

export async function doctorCommand(): Promise<void> {
  console.log(chalk.cyan('🩺 MoAI-ADK System Diagnosis'));
  console.log(chalk.cyan('============================='));

  const systemChecker = new SystemChecker();
  const results = await systemChecker.diagnose();

  results.forEach(result => {
    const status = result.satisfied ? '✅' : '❌';
    const version = result.version ? ` v${result.version}` : '';
    const required = result.required ? ` (required: >=${result.required})` : '';

    console.log(`${status} ${result.name}${version}${required}`);

    if (!result.satisfied && result.error) {
      console.log(chalk.red(`   Error: ${result.error}`));
    }
  });
}
```

#### 성공 기준
- [ ] 모든 CLI 테스트 통과
- [ ] `moai --version`, `moai --help` 동작
- [ ] `moai init test-project` 기본 동작
- [ ] `moai doctor` 시스템 진단 동작

---

### Day 7: 통합 테스트 및 검증 (REFACTOR + INTEGRATION)

#### 작업 목표
- 전체 시스템 통합 테스트
- 코드 품질 개선 및 리팩토링
- Week 1 완료 기준 검증
- 성능 및 보안 검사

#### REFACTOR 단계
```typescript
// 코드 품질 개선
// 1. 에러 처리 강화
// 2. 로깅 시스템 추가
// 3. 타입 안전성 강화
// 4. 성능 최적화
```

#### INTEGRATION 단계
```typescript
// __tests__/integration/full-workflow.test.ts
describe('Full Workflow Integration', () => {
  test('complete init workflow', async () => {
    // 1. 시스템 검사
    // 2. CLI 실행
    // 3. 결과 검증
    // 4. 에러 시나리오 테스트
  });

  test('performance benchmarks', async () => {
    // CLI 시작 시간 < 2초
    // 시스템 검사 시간 < 5초
    // 메모리 사용량 < 100MB
  });

  test('security validation', async () => {
    // 명령어 인젝션 방지
    // 입력 검증
    // 공격 벡터 테스트
  });
});
```

#### 성공 기준
- [ ] 전체 테스트 수트 100% 통과
- [ ] 테스트 커버리지 ≥ 80%
- [ ] ESLint 에러 0개
- [ ] 타입 검사 통과
- [ ] 성능 벤치마크 통과
- [ ] 보안 검사 통과

---

## 성능 목표 및 측정

### 성능 지표
```
CLI 시작 시간: < 2초
시스템 검사: < 5초
메모리 사용: < 100MB
빌드 시간: < 30초
테스트 실행: < 60초
```

### 품질 지표
```
코드 커버리지: ≥ 80%
타입 커버리지: 100%
ESLint 에러: 0개
대상 파일 수: < 50개
평균 함수 길이: < 30 LOC
```

---

## 위험 관리 및 대응책

### 기술적 위험
1. **Node.js 버전 호환성**
   - 위험: Claude Code에서 사용하는 Node.js 버전과 불일치
   - 대응: 최소 버전 18.0.0으로 설정, 호환성 테스트

2. **TypeScript 컴파일 오류**
   - 위험: 복잡한 타입 정의로 인한 컴파일 실패
   - 대응: 단계적 타입 도입, strict 모드 점진적 적용

3. **크로스 플랫폼 이슈**
   - 위험: Windows/macOS/Linux 환경별 동작 차이
   - 대응: 각 플랫폼에서 단위 테스트, CI/CD 매트릭스

### 상황별 대응 계획
```
시나리오 1: Day 3-4 지연
→ Day 5-6 작업을 Day 6-7로 연기
→ 핵심 기능 우선 구현

시나리오 2: 성능 목표 미달성
→ 기능 범위 축소
→ Week 2로 원래 기능 이전

시나리오 3: 심각한 블로커 발생
→ Python 기반 대안 접근
→ 하이브리드 접근법 고려
```

---

**Week 1 성공 완료 조건**: `moai --version`, `moai --help`, `moai doctor` 명령어가 정상 동작하고, 시스템 요구사항 자동 검증 모듈이 완성되어 테스트를 통과해야 함.