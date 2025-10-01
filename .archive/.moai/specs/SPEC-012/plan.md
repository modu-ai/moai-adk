# SPEC-012 TDD 구현 계획: Python → TypeScript 완전 포팅 (5주)

> **@CODE:COMPLETE-PORTING-PLAN-012** Python MoAI-ADK를 TypeScript로 완전 전환하는 TDD 실행 계획
> **@SPEC:MIGRATION-STRATEGY-012** 5주 단계별 포팅 전략 및 Red-Green-Refactor 적용
> **@CODE:FULL-MIGRATION-012** Python 완전 제거 + TypeScript 단독 실행 목표

---

## 완전 포팅 전략

### 포팅 원칙
```
Python 분석 → TypeScript 설계 → TDD 구현 → 기능 검증 → Python 제거
    │              │              │           │             │
    │              │              │           │             └─ 완전 전환
    │              │              │           └─────────────── 동등성 확인
    │              │              └─────────────────────────── Red-Green-Refactor
    │              └─────────────────────────────────────────── 타입 안전 설계
    └─────────────────────────────────────────────────────── 기존 기능 매핑
```

### 5주 포팅 로드맵

1. **Week 1**: TypeScript 기반 구축 + 시스템 검증 (신규 기능)
2. **Week 2**: 설치 시스템 완전 포팅 (Python install/ → TypeScript)
3. **Week 3**: 훅 시스템 완전 전환 (7개 Python 훅 → TypeScript)
4. **Week 4**: 통합 최적화 + 성능 목표 달성 (0.8초, 80MB)
5. **Week 5**: npm 배포 + Python 버전 deprecation

### 포팅 우선순위
1. **핵심 기능**: CLI, 설치, Git 관리, TAG 시스템
2. **품질 기능**: TRUST 검증, 보안 훅, 품질 게이트
3. **확장 기능**: 문서 시스템, 에이전트, 템플릿

---

## Week 1: TypeScript 기반 구축

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

## 전체 5주 포팅 마일스톤

### Week 2: 설치 시스템 완전 포팅
**목표**: Python install/ 모듈 → TypeScript 100% 전환

**주요 작업**:
- InstallationOrchestrator 포팅 (src/moai_adk/install/installer.py → TypeScript)
- Git 관리 시스템 포팅 (src/moai_adk/core/git_manager.py → TypeScript)
- 템플릿 엔진 포팅 (Jinja2 → Mustache.js)
- 크로스 플랫폼 파일 작업 (fs-extra 활용)

### Week 3: 훅 시스템 완전 전환
**목표**: 7개 Python 훅 → TypeScript 완전 대체

**주요 작업**:
- pre_write_guard.py → pre-write-guard.ts
- policy_block.py → policy-block.ts
- steering_guard.py → steering-guard.ts
- session_start.py → session-start.ts
- (추가 3개 훅 포팅)
- Claude Code 인터페이스 호환성 유지

### Week 4: 통합 최적화 + 성능 달성
**목표**: TRUST 원칙 구현 + 성능 목표 달성

**주요 작업**:
- TAG 시스템 성능 최적화 (1.1초 → 0.8초)
- 메모리 사용량 최적화 (174MB → 80MB)
- better-sqlite3 통합 및 최적화
- TRUST 검증 시스템 구현

### Week 5: 배포 및 Python 폐기
**목표**: npm 정식 배포 + Python 버전 deprecation

**주요 작업**:
- npm 패키지 정식 배포 (moai-adk@1.0.0)
- Python 버전 deprecation 공지
- 마이그레이션 가이드 작성
- 사용자 지원 체계 구축

---

## 성능 목표 및 측정 (완전 포팅 기준)

### 포팅 성능 목표
```
스캔 성능: Python 1.1초 → TypeScript 0.8초 (27% 개선)
메모리 사용: Python 174MB → TypeScript 80MB (54% 절약)
설치 시간: pip 30-60초 → npm 30초 이하
패키지 크기: Python 15MB → TypeScript 10MB 이하
설치 성공률: Python 95% → TypeScript 98%
```

### 품질 지표 (완전 전환)
```
Python 코드 잔존: 0% (완전 제거)
TypeScript 구현: 100% (모든 기능 포팅)
타입 커버리지: 100% (strict 모드)
테스트 커버리지: ≥ 85%
크로스 플랫폼: Windows/macOS/Linux 100% 지원
```

---

## 위험 관리 및 대응책 (완전 포팅 관점)

### 포팅 위험
1. **기능 누락 위험**
   - 위험: Python 기능이 TypeScript에서 구현되지 않음
   - 대응: 기능 매핑 체크리스트, 단위 테스트 기반 검증

2. **성능 저하 위험**
   - 위험: TypeScript 버전이 Python보다 느림
   - 대응: 벤치마크 테스트, 프로파일링, 최적화

3. **호환성 깨짐 위험**
   - 위험: 기존 사용자 프로젝트와 호환성 문제
   - 대응: 호환성 테스트 수트, 마이그레이션 도구

### 포팅 실패 시 대응
```
시나리오 1: 중대한 기능 누락 발견
→ 해당 기능 긴급 포팅
→ 배포 일정 조정

시나리오 2: 성능 목표 미달성
→ 성능 집중 최적화 주간
→ 병목 지점 프로파일링 및 개선

시나리오 3: 사용자 호환성 문제
→ 호환성 레이어 구현
→ 점진적 마이그레이션 지원
```

---

**5주 완전 포팅 성공 조건**:
- Python 코드 0% 잔존
- npm install -g moai-adk 단독 설치
- 기존 모든 기능 100% 동작
- 성능 목표 달성 (0.8초, 80MB)
- 기존 프로젝트 완벽 호환