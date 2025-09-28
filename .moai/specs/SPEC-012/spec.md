# SPEC-012: MoAI-ADK TypeScript 기반 구축 (Week 1)

> **@REQ:TS-FOUNDATION-012** Python v0.1.28 기반 MoAI-ADK의 TypeScript 포팅 기반 구축
> **@DESIGN:TS-ARCH-012** 시스템 요구사항 자동 검증 + CLI 구조 설계
> **@TASK:WEEK1-012** 5주 로드맵의 Week 1: 기반 구축 실행

---

## Environment (환경 및 가정사항)

### E1. 기존 Python 기반 현황
- **현재 버전**: MoAI-ADK v0.1.28 (성능 최적화 완료)
- **성능 달성**: 4,686파일 스캔 1.1초, 87.6% 품질 개선 완료
- **코드 구조**: 70개+ 모듈, TRUST 원칙 준수, CLI/Core/Install 3모듈 구조
- **Claude Code 통합**: 7개 에이전트, 5개 명령어, 8개 훅 완성

### E2. 타겟 TypeScript 환경
- **Node.js**: 18.0.0+ (Claude Code 사용자 기본 보유)
- **패키지 관리**: npm 생태계 중심
- **빌드 도구**: tsup (고성능 번들러)
- **타겟 플랫폼**: Windows/macOS/Linux 크로스 플랫폼

### E3. 제약 조건
- **기존 프로젝트 호환성**: `.moai/` 및 `.claude/` 구조 유지
- **16-Core TAG 시스템**: SQLite → better-sqlite3 포팅 필수
- **Claude Code API**: 기존 훅 인터페이스 호환성 유지
- **TRUST 원칙**: Test First 개발 방법론 준수

## Assumptions (전제 조건)

### A1. 개발 환경 가정
- TypeScript 5.0+ 개발 환경 구축 가능
- Jest 테스트 프레임워크 활용
- ESLint + Prettier 코드 품질 도구 통합
- GitHub Actions CI/CD 파이프라인 구축

### A2. 성능 기준점
- **현재 Python 성능**: 1.1초 스캔, 150-174MB 메모리 사용
- **TypeScript 목표**: 0.8초 스캔, 50-80MB 메모리 사용
- **설치 시간**: 30초 이내 유지 (현재 30-60초)

### A3. 포팅 우선순위
1. **1차**: 시스템 요구사항 자동 검증 모듈 (핵심 혁신)
2. **2차**: 기본 CLI 구조 및 명령어 파싱
3. **3차**: 빌드 시스템 및 패키지 구성
4. **4차**: TypeScript 컴파일 및 배포 준비

## Requirements (기능 요구사항)

### R1. 시스템 요구사항 자동 검증 모듈 @REQ:AUTO-VERIFY-012
**핵심 혁신 기능**: 사용자 설치 환경의 필수 도구를 자동으로 감지하고 설치를 제안/실행

#### R1.1 요구사항 정의 시스템
```typescript
interface SystemRequirement {
  name: string;                    // 도구명 (Git, SQLite3, Claude Code 등)
  category: 'runtime' | 'development' | 'optional';
  minVersion?: string;             // 최소 버전 요구사항
  installCommands: Record<string, string>; // 플랫폼별 설치 명령어
  checkCommand: string;            // 설치 확인 명령어
  versionCommand?: string;         // 버전 확인 명령어
}
```

#### R1.2 자동 감지 엔진
- **기능**: Node.js, Git, SQLite3, Claude Code 등 필수 도구 설치 상태 감지
- **버전 검증**: semver 기반 최소 버전 요구사항 검증
- **플랫폼 지원**: darwin/linux/win32 자동 인식
- **오류 처리**: 명령어 실행 실패 시 적절한 에러 메시지 제공

#### R1.3 자동 설치 제안 시스템
- **대화형 UI**: inquirer 기반 사용자 확인 프롬프트
- **플랫폼별 명령어**: brew/apt-get/winget 등 자동 선택
- **설치 진행률**: ora 스피너 및 실시간 피드백
- **설치 검증**: 설치 후 재검사 및 성공/실패 상태 리포트

### R2. 기본 CLI 구조 @REQ:CLI-FOUNDATION-012

#### R2.1 Commander.js 기반 CLI 프레임워크
```bash
moai --version           # 버전 정보
moai --help             # 도움말
moai init <project>     # 시스템 검증 + 프로젝트 초기화
moai doctor             # 시스템 진단
moai status             # 프로젝트 상태 (향후 확장)
moai update             # 업데이트 (향후 확장)
```

#### R2.2 명령어 파싱 및 라우팅
- **동적 import**: 필요시에만 명령어 모듈 로드
- **에러 처리**: 잘못된 명령어 입력 시 적절한 안내
- **도움말 시스템**: 명령어별 상세 사용법 제공
- **진행률 표시**: chalk + ora 조합으로 사용자 피드백

### R3. 빌드 시스템 구축 @REQ:BUILD-SYSTEM-012

#### R3.1 TypeScript 컴파일 환경
- **tsconfig.json**: ES2022 타겟, strict 모드 활성화
- **tsup 빌드**: 고성능 번들링, ESM/CJS 듀얼 지원
- **소스맵**: 디버깅을 위한 소스맵 생성
- **타입 정의**: .d.ts 파일 자동 생성

#### R3.2 개발 도구 통합
- **ESLint**: @typescript-eslint 규칙 적용
- **Prettier**: 코드 포맷팅 자동화
- **Jest**: TypeScript 지원 테스트 환경
- **Husky**: pre-commit 훅 설정 (향후 확장)

### R4. 패키지 구성 및 배포 준비 @REQ:PACKAGE-CONFIG-012

#### R4.1 npm 패키지 설정
```json
{
  "name": "moai-adk",
  "version": "0.0.1",
  "description": "🗿 MoAI-ADK: Modu-AI Agentic Development kit",
  "main": "dist/index.js",
  "bin": {
    "moai": "dist/cli/index.js"
  },
  "engines": {
    "node": ">=18.0.0"
  }
}
```

#### R4.2 의존성 매핑
- **CLI**: commander (Python click 대체)
- **UI**: chalk + ora (Python colorama 대체)
- **파일**: fs-extra (향상된 파일 작업)
- **버전**: semver (버전 비교)
- **프롬프트**: inquirer (대화형 UI)
- **명령어**: which (명령어 존재 확인)

## Specifications (상세 명세)

### S1. 프로젝트 구조 설계

```
moai-adk/
├── package.json              # npm 패키지 설정
├── tsconfig.json            # TypeScript 설정
├── tsup.config.ts           # 빌드 설정
├── jest.config.js           # 테스트 설정
├── .eslintrc.json          # 린트 설정
├── .prettierrc             # 포맷터 설정
├── src/
│   ├── cli/
│   │   ├── index.ts        # CLI 진입점
│   │   ├── commands/
│   │   │   ├── init.ts     # moai init 명령어
│   │   │   └── doctor.ts   # moai doctor 명령어
│   │   └── wizard.ts       # 대화형 설치 마법사
│   ├── core/
│   │   └── system-checker/ # 🆕 시스템 요구사항 검증
│   │       ├── requirements.ts  # 요구사항 정의
│   │       ├── detector.ts      # 설치된 도구 감지
│   │       ├── installer.ts     # 자동 설치 제안/실행
│   │       └── index.ts         # 통합 인터페이스
│   ├── utils/
│   │   ├── logger.ts       # 구조화 로깅
│   │   └── version.ts      # 버전 정보
│   └── index.ts            # 메인 API 진입점
├── __tests__/              # Jest 테스트
│   ├── system-checker/     # 시스템 검증 테스트
│   └── cli/               # CLI 테스트
└── dist/                  # 컴파일된 JavaScript
```

### S2. 시스템 요구사항 검증 상세 설계

#### S2.1 필수 도구 정의
```typescript
export const SYSTEM_REQUIREMENTS: SystemRequirement[] = [
  {
    name: 'Node.js',
    category: 'runtime',
    minVersion: '18.0.0',
    installCommands: {
      darwin: 'brew install node',
      linux: 'sudo apt-get install nodejs npm',
      win32: 'winget install OpenJS.NodeJS'
    },
    checkCommand: 'node --version',
    versionCommand: 'node --version'
  },
  {
    name: 'Git',
    category: 'runtime',
    minVersion: '2.20.0',
    installCommands: {
      darwin: 'brew install git',
      linux: 'sudo apt-get install git',
      win32: 'winget install Git.Git'
    },
    checkCommand: 'git --version',
    versionCommand: 'git --version'
  },
  {
    name: 'SQLite3',
    category: 'runtime',
    minVersion: '3.30.0',
    installCommands: {
      darwin: 'brew install sqlite',
      linux: 'sudo apt-get install sqlite3',
      win32: 'winget install SQLite.SQLite'
    },
    checkCommand: 'sqlite3 --version',
    versionCommand: 'sqlite3 --version'
  }
];
```

#### S2.2 자동 감지 엔진 구현
```typescript
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

### S3. CLI 명령어 상세 설계

#### S3.1 moai init 명령어
```typescript
export async function initCommand(projectName: string): Promise<void> {
  console.log(chalk.cyan('🗿 MoAI-ADK Project Initialization'));
  console.log(chalk.cyan('================================'));

  // Step 1: 시스템 요구사항 검증
  console.log('\n🔍 Step 1: System Requirements Check');
  const systemChecker = new SystemChecker();
  const missingRequirements = await systemChecker.checkAll();

  if (missingRequirements.length > 0) {
    const autoInstaller = new AutoInstaller();
    await autoInstaller.suggestInstallation(missingRequirements);
  }

  // Step 2: 프로젝트 구조 생성 (향후 Week 2에서 구현)
  console.log('\n🚀 Step 2: Project Setup (Coming in Week 2)');
  console.log(chalk.green(`✅ Project "${projectName}" foundation ready!`));
}
```

#### S3.2 moai doctor 명령어
```typescript
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

## Traceability (추적성 태그)

### Primary Chain
- **@REQ:TS-FOUNDATION-012** → **@DESIGN:TS-ARCH-012** → **@TASK:WEEK1-012** → **@TEST:TS-FOUNDATION-012**

### Related Tags
- **@FEATURE:AUTO-VERIFY-012**: 시스템 요구사항 자동 검증 핵심 기능
- **@API:CLI-COMMANDS-012**: CLI 명령어 공개 API 인터페이스
- **@DATA:SYSTEM-REQUIREMENTS-012**: 시스템 요구사항 데이터 모델
- **@PERF:STARTUP-TIME-012**: CLI 시작 시간 최적화
- **@SEC:COMMAND-INJECTION-012**: 명령어 실행 보안 검증
- **@DOCS:CLI-USAGE-012**: CLI 사용법 문서화

---

**완료 조건**: Week 1 종료 시점에 `moai --version`, `moai --help`, `moai doctor` 명령어가 정상 동작하고, 시스템 요구사항 자동 검증 모듈이 완성되어 테스트를 통과해야 함.