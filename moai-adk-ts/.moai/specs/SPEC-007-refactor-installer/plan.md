# SPEC-007 구현 계획

## @SPEC:REFACTOR-007 | Chain: @SPEC:REFACTOR-007 -> @SPEC:REFACTOR-007 -> @CODE:REFACTOR-007 -> @TEST:REFACTOR-007

## 개요

이 문서는 SPEC-007 (Installer 리팩토링)의 3-Phase TDD 구현 계획을 정의합니다.

## TDD 구현 전략

### Bottom-up 접근

의존성이 없는 클래스부터 시작하여 점진적으로 통합합니다:

```
Phase 1: CommandBuilder (의존성 없음)
    ↓
Phase 2: PackageJsonBuilder (CommandBuilder 의존)
    ↓
Phase 3: PackageManagerInstaller (CommandBuilder 의존)
```

이 순서는 각 단계에서 이전 단계의 완성된 코드를 활용하여 안정적인 TDD를 보장합니다.

## Phase 1: CommandBuilder 구현

### 🔴 RED - 실패하는 테스트 작성

**파일**: `src/__tests__/core/package-manager/command-builder.test.ts` (신규, ~100 LOC)

**테스트 구조**:
```typescript
describe('CommandBuilder', () => {
  describe('Install Command Building', () => {
    test('should build npm install command');
    test('should build yarn add command');
    test('should build pnpm add command');
    test('should include --save-dev flag for dev dependencies');
    test('should include --global flag for global installation');
  });

  describe('Run Command Building', () => {
    test('should build npm run command');
    test('should build yarn run command');
    test('should build pnpm run command');
  });

  describe('Test Command Building', () => {
    test('should build default test command');
    test('should build jest test command');
  });

  describe('Init Command Building', () => {
    test('should build npm init command');
    test('should build yarn init command');
    test('should build pnpm init command');
  });

  describe('Engine Requirements', () => {
    test('should return npm engine requirement');
    test('should return yarn engine requirement');
    test('should return pnpm engine requirement');
  });

  describe('Error Handling', () => {
    test('should throw error for unsupported package manager');
  });
});
```

**테스트 케이스 예시**:
```typescript
// @TEST:REFACTOR-007-001: npm install 명령어 생성
test('should build npm install command', () => {
  const builder = new CommandBuilder();
  const command = builder.buildInstallCommand(
    ['express', 'lodash'],
    {
      packageManager: PackageManagerType.NPM,
      isDevelopment: false,
    }
  );
  expect(command).toBe('npm install express lodash');
});

// @TEST:REFACTOR-007-002: 개발 의존성 플래그
test('should include --save-dev flag for dev dependencies', () => {
  const builder = new CommandBuilder();
  const command = builder.buildInstallCommand(
    ['typescript'],
    {
      packageManager: PackageManagerType.NPM,
      isDevelopment: true,
    }
  );
  expect(command).toBe('npm install --save-dev typescript');
});
```

**예상 실패 메시지**:
```
FAIL  src/__tests__/core/package-manager/command-builder.test.ts
  ● CommandBuilder › Install Command Building › should build npm install command

    ReferenceError: CommandBuilder is not defined
```

### 🟢 GREEN - 최소 구현

**파일**: `src/core/package-manager/command-builder.ts` (신규, ~100 LOC)

**구현 구조**:
```typescript
// @CODE:REFACTOR-007 | Chain: @TEST:REFACTOR-007 -> @CODE:REFACTOR-007
// Related: @CODE:PKG-002

/**
 * @file Command builder for package managers
 * @author MoAI Team
 * @tags @CODE:COMMAND-BUILDER-001
 */

import {
  type PackageInstallOptions,
  PackageManagerType,
} from '@/types/package-manager';

/**
 * Builds command strings for various package managers
 * @tags @CODE:COMMAND-BUILDER-001:FEATURE
 */
export class CommandBuilder {
  /**
   * Build install command for package manager
   * @tags @CODE:BUILD-INSTALL-CMD-001:API
   */
  public buildInstallCommand(
    packages: string[],
    options: PackageInstallOptions
  ): string {
    const { packageManager, isDevelopment, isGlobal } = options;

    switch (packageManager) {
      case PackageManagerType.NPM:
        return this.buildNpmInstallCommand(packages, isDevelopment, isGlobal);
      case PackageManagerType.YARN:
        return this.buildYarnInstallCommand(packages, isDevelopment, isGlobal);
      case PackageManagerType.PNPM:
        return this.buildPnpmInstallCommand(packages, isDevelopment, isGlobal);
      default:
        throw new Error(`Unsupported package manager: ${packageManager}`);
    }
  }

  /**
   * Build run command for package manager
   * @tags @CODE:BUILD-RUN-CMD-001:API
   */
  public buildRunCommand(packageManagerType: PackageManagerType): string {
    // 구현...
  }

  /**
   * Build test command
   * @tags @CODE:BUILD-TEST-CMD-001:API
   */
  public buildTestCommand(
    packageManagerType: PackageManagerType,
    testingFramework?: string
  ): string {
    // 구현...
  }

  /**
   * Build init command
   * @tags @CODE:BUILD-INIT-CMD-001:API
   */
  public buildInitCommand(packageManagerType: PackageManagerType): string {
    // 구현...
  }

  /**
   * Get package manager engine requirement
   * @tags @CODE:GET-ENGINE-001:API
   */
  public getPackageManagerEngine(
    packageManagerType: PackageManagerType
  ): Record<string, string> {
    // 구현...
  }

  // Private helpers
  private buildNpmInstallCommand(
    packages: string[],
    isDevelopment?: boolean,
    isGlobal?: boolean
  ): string {
    // 구현...
  }

  // ... 기타 private 메서드
}
```

**테스트 통과 확인**:
```bash
pnpm test command-builder.test.ts
# PASS  src/__tests__/core/package-manager/command-builder.test.ts
# Test Suites: 1 passed, 1 total
# Tests:       15 passed, 15 total
```

### 🔄 REFACTOR - 품질 개선

**개선 항목**:
1. **중복 제거**: npm/yarn/pnpm 명령어 생성 로직의 공통 패턴 추출
2. **가독성**: 명령어 조립 로직을 명확한 단계로 분리
3. **타입 안전성**: 패키지 매니저 타입 체크 강화
4. **성능**: 불필요한 문자열 연산 최소화

**리팩토링 예시**:
```typescript
// Before: 중복된 로직
private buildNpmInstallCommand(packages: string[], isDev?: boolean, isGlobal?: boolean): string {
  let cmd = 'npm install';
  if (isDev) cmd += ' --save-dev';
  if (isGlobal) cmd += ' --global';
  return `${cmd} ${packages.join(' ')}`;
}

// After: 공통 패턴 추출
private buildInstallCommandWithFlags(
  baseCommand: string,
  packages: string[],
  flags: string[]
): string {
  const commandParts = [baseCommand, ...flags, ...packages];
  return commandParts.join(' ');
}
```

## Phase 2: PackageJsonBuilder 구현

### 🔴 RED - 실패하는 테스트 작성

**파일**: `src/__tests__/core/package-manager/package-json-builder.test.ts` (신규, ~120 LOC)

**테스트 구조**:
```typescript
describe('PackageJsonBuilder', () => {
  let builder: PackageJsonBuilder;
  let commandBuilder: CommandBuilder;

  beforeEach(() => {
    commandBuilder = new CommandBuilder();
    builder = new PackageJsonBuilder(commandBuilder);
  });

  describe('Package.json Generation', () => {
    test('should generate basic package.json');
    test('should include default fields');
    test('should use provided configuration');
  });

  describe('TypeScript Integration', () => {
    test('should include TypeScript dependencies');
    test('should include TypeScript scripts');
    test('should configure type-check script');
  });

  describe('Testing Framework Integration', () => {
    test('should include Jest dependencies');
    test('should include Jest scripts');
    test('should include TypeScript Jest dependencies when both enabled');
  });

  describe('Scripts Generation', () => {
    test('should generate npm scripts');
    test('should generate yarn scripts');
    test('should generate pnpm scripts');
    test('should include TypeScript dev script');
  });

  describe('Dependency Management', () => {
    test('should add dependencies');
    test('should add dev dependencies');
    test('should merge with existing dependencies');
    test('should preserve existing configuration');
  });
});
```

**테스트 케이스 예시**:
```typescript
// @TEST:REFACTOR-007-101: package.json 기본 생성
test('should generate basic package.json', () => {
  const config = {
    name: 'test-project',
    version: '1.0.0',
  };

  const packageJson = builder.generatePackageJson(
    config,
    PackageManagerType.NPM
  );

  expect(packageJson.name).toBe('test-project');
  expect(packageJson.version).toBe('1.0.0');
  expect(packageJson.scripts).toBeDefined();
  expect(packageJson.engines).toBeDefined();
});

// @TEST:REFACTOR-007-102: TypeScript 의존성 추가
test('should include TypeScript dependencies', () => {
  const config = { name: 'ts-project', version: '1.0.0' };

  const packageJson = builder.generatePackageJson(
    config,
    PackageManagerType.NPM,
    true // includeTypeScript
  );

  expect(packageJson.devDependencies?.typescript).toBeDefined();
  expect(packageJson.devDependencies?.['@types/node']).toBeDefined();
  expect(packageJson.scripts?.build).toContain('tsc');
});
```

### 🟢 GREEN - 최소 구현

**파일**: `src/core/package-manager/package-json-builder.ts` (신규, ~120 LOC)

**구현 구조**:
```typescript
// @CODE:REFACTOR-007 | Chain: @TEST:REFACTOR-007 -> @CODE:REFACTOR-007
// Related: @CODE:PKG-002

/**
 * @file Package.json configuration builder
 * @author MoAI Team
 * @tags @CODE:PACKAGE-JSON-BUILDER-001
 */

import { type PackageJsonConfig, PackageManagerType } from '@/types/package-manager';
import { CommandBuilder } from './command-builder';

/**
 * Builds and manages package.json configurations
 * @tags @CODE:PACKAGE-JSON-BUILDER-001:FEATURE
 */
export class PackageJsonBuilder {
  private commandBuilder: CommandBuilder;

  constructor(commandBuilder: CommandBuilder) {
    this.commandBuilder = commandBuilder;
  }

  /**
   * Generate package.json configuration
   * @tags @CODE:GENERATE-PKG-JSON-001:API
   */
  public generatePackageJson(
    projectConfig: Partial<PackageJsonConfig>,
    packageManagerType: PackageManagerType,
    includeTypeScript: boolean = false,
    testingFramework?: string
  ): PackageJsonConfig {
    const baseConfig = this.createBaseConfig(projectConfig);
    const scripts = this.generateScripts(
      packageManagerType,
      includeTypeScript,
      testingFramework
    );
    const engines = this.buildEngines(packageManagerType);

    let config: PackageJsonConfig = {
      ...baseConfig,
      scripts,
      engines,
    };

    if (includeTypeScript) {
      config = this.addTypeScriptDependencies(config);
    }

    if (testingFramework) {
      config = this.addTestingDependencies(config, testingFramework, includeTypeScript);
    }

    return config;
  }

  /**
   * Generate scripts section
   * @tags @CODE:GENERATE-SCRIPTS-001:API
   */
  public generateScripts(
    packageManagerType: PackageManagerType,
    includeTypeScript: boolean,
    testingFramework?: string
  ): Record<string, string> {
    // 구현...
  }

  /**
   * Add dependencies
   * @tags @CODE:ADD-DEPS-001:API
   */
  public addDependencies(
    existingPackageJson: Partial<PackageJsonConfig>,
    newDependencies: Record<string, string>
  ): PackageJsonConfig {
    // 구현...
  }

  /**
   * Add dev dependencies
   * @tags @CODE:ADD-DEV-DEPS-001:API
   */
  public addDevDependencies(
    existingPackageJson: Partial<PackageJsonConfig>,
    newDevDependencies: Record<string, string>
  ): PackageJsonConfig {
    // 구현...
  }

  // Private helpers
  private createBaseConfig(config: Partial<PackageJsonConfig>): PackageJsonConfig {
    // 구현...
  }

  private buildEngines(packageManagerType: PackageManagerType): Record<string, string> {
    return {
      node: '>=18.0.0',
      ...this.commandBuilder.getPackageManagerEngine(packageManagerType),
    };
  }

  private addTypeScriptDependencies(config: PackageJsonConfig): PackageJsonConfig {
    // 구현...
  }

  private addTestingDependencies(
    config: PackageJsonConfig,
    framework: string,
    includeTypeScript: boolean
  ): PackageJsonConfig {
    // 구현...
  }
}
```

### 🔄 REFACTOR - 품질 개선

**개선 항목**:
1. **의존성 병합 로직 개선**: 깊은 객체 병합 유틸리티 사용
2. **설정 빌더 패턴**: 체이닝 가능한 빌더 메서드 고려
3. **타입 안전성**: 의존성 버전 상수화
4. **테스트 용이성**: 의존성 주입 검증

## Phase 3: PackageManagerInstaller 리팩토링

### 🔴 RED - 기존 테스트 업데이트

**파일**: `src/__tests__/core/package-manager/installer.test.ts` (수정, 327 LOC → ~107 LOC)

**변경 내역**:
1. **제거할 테스트**: package.json 생성 관련 테스트 (PackageJsonBuilder로 이동)
2. **유지할 테스트**: 패키지 설치, 프로젝트 초기화 테스트
3. **수정할 테스트**: CommandBuilder 모의 객체 사용

**테스트 구조**:
```typescript
describe('PackageManagerInstaller', () => {
  let installer: PackageManagerInstaller;
  let mockCommandBuilder: CommandBuilder;

  beforeEach(() => {
    vi.clearAllMocks();
    mockCommandBuilder = new CommandBuilder();
    installer = new PackageManagerInstaller(mockCommandBuilder);
  });

  describe('Package Installation', () => {
    test('should install packages using npm');
    test('should install dev dependencies using yarn');
    test('should install global packages using pnpm');
    test('should handle installation failures');
  });

  describe('Project Initialization', () => {
    test('should initialize project with package manager');
    test('should handle initialization failures');
  });
});
```

**테스트 수정 예시**:
```typescript
// Before: 내부 명령어 생성 로직 테스트
expect(mockExeca).toHaveBeenCalledWith(
  'npm',
  ['install', 'express', 'lodash'],
  expect.any(Object)
);

// After: CommandBuilder 사용 확인
const expectedCommand = mockCommandBuilder.buildInstallCommand(packages, options);
expect(mockExeca).toHaveBeenCalledWith(
  expectedCommand.split(' ')[0],
  expectedCommand.split(' ').slice(1),
  expect.any(Object)
);
```

### 🟢 GREEN - 리팩토링 구현

**파일**: `src/core/package-manager/installer.ts` (수정, 399 LOC → ~150 LOC)

**변경 내역**:
1. **제거**: `generatePackageJson()`, `addDependencies()`, `addDevDependencies()`, `generateScripts()` 메서드
2. **제거**: private 명령어 생성 메서드들 (`buildInstallCommand`, `getRunCommand`, `getTestCommand` 등)
3. **추가**: CommandBuilder 의존성 주입
4. **유지**: `installPackages()`, `initializeProject()` 메서드 (내부 로직만 변경)

**리팩토링 구조**:
```typescript
// @CODE:REFACTOR-007 | Chain: @TEST:REFACTOR-007 -> @CODE:REFACTOR-007
// Related: @CODE:PKG-002

/**
 * @file Package manager installer (Refactored)
 * @author MoAI Team
 * @tags @CODE:PACKAGE-MANAGER-INSTALLER-001
 */

import { execa } from 'execa';
import {
  type PackageInstallOptions,
  PackageManagerType,
} from '@/types/package-manager';
import { CommandBuilder } from './command-builder';

export interface InstallResult {
  success: boolean;
  installedPackages: string[];
  error?: string;
  output?: string;
}

export interface InitResult {
  success: boolean;
  packageJsonPath?: string;
  error?: string;
  output?: string;
}

/**
 * Orchestrates package installation and project initialization
 * @tags @CODE:PACKAGE-MANAGER-INSTALLER-001:FEATURE
 */
export class PackageManagerInstaller {
  private commandBuilder: CommandBuilder;

  constructor(commandBuilder?: CommandBuilder) {
    this.commandBuilder = commandBuilder ?? new CommandBuilder();
  }

  /**
   * Install packages using specified package manager
   * @tags @CODE:INSTALL-PACKAGES-001:API
   */
  public async installPackages(
    packages: string[],
    options: PackageInstallOptions
  ): Promise<InstallResult> {
    try {
      const command = this.commandBuilder.buildInstallCommand(packages, options);
      const [executable, ...args] = command.split(' ');

      const result = await execa(executable!, args, {
        cwd: options.workingDirectory || process.cwd(),
        timeout: 300000,
        reject: false,
      });

      if (result.exitCode === 0) {
        return {
          success: true,
          installedPackages: packages,
          output: result.stdout,
        };
      } else {
        return {
          success: false,
          installedPackages: [],
          error: result.stderr || `Command failed with exit code ${result.exitCode}`,
          output: result.stdout,
        };
      }
    } catch (error) {
      return {
        success: false,
        installedPackages: [],
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  /**
   * Initialize project with package manager
   * @tags @CODE:INIT-PROJECT-001:API
   */
  public async initializeProject(
    projectPath: string,
    packageManagerType: PackageManagerType
  ): Promise<InitResult> {
    try {
      const command = this.commandBuilder.buildInitCommand(packageManagerType);
      const [executable, ...args] = command.split(' ');

      const result = await execa(executable!, args, {
        cwd: projectPath,
        timeout: 30000,
        reject: false,
      });

      if (result.exitCode === 0) {
        return {
          success: true,
          packageJsonPath: `${projectPath}/package.json`,
          output: result.stdout,
        };
      } else {
        return {
          success: false,
          error: result.stderr || `Command failed with exit code ${result.exitCode}`,
          output: result.stdout,
        };
      }
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }
}
```

**LOC 감소**:
- Before: 399 LOC
- After: ~150 LOC
- 감소율: -62%

### 🔄 REFACTOR - 최종 정리

**개선 항목**:
1. **오류 처리 개선**: 공통 오류 처리 로직 추출
2. **타임아웃 설정**: 설정 가능한 타임아웃 옵션 추가
3. **로깅**: 실행 명령어 로깅 추가 (디버깅 용)
4. **타입 안전성**: Result 타입 개선

## Phase 4: 통합 테스트 및 검증

### 통합 테스트 시나리오

**시나리오 1: 전체 워크플로우**
```typescript
describe('Integration: Full Project Setup', () => {
  test('should setup new project with all components', async () => {
    const commandBuilder = new CommandBuilder();
    const packageJsonBuilder = new PackageJsonBuilder(commandBuilder);
    const installer = new PackageManagerInstaller(commandBuilder);

    // 1. package.json 생성
    const packageJson = packageJsonBuilder.generatePackageJson(
      { name: 'test-project', version: '1.0.0' },
      PackageManagerType.NPM,
      true, // TypeScript
      'jest'
    );

    expect(packageJson.name).toBe('test-project');
    expect(packageJson.devDependencies?.typescript).toBeDefined();

    // 2. 프로젝트 초기화 (모의 환경)
    // 3. 패키지 설치 (모의 환경)
  });
});
```

### 커버리지 검증

```bash
pnpm test:coverage

# 목표: 각 파일 ≥ 85% 커버리지
# - command-builder.ts: ≥ 85%
# - package-json-builder.ts: ≥ 85%
# - installer.ts: ≥ 85%
```

### 코드 품질 검증

```bash
# Biome 린트 체크
pnpm lint

# TypeScript 타입 체크
pnpm type-check

# 복잡도 체크 (수동)
# - 각 메서드 복잡도 ≤ 10 확인
```

### @TAG 체인 검증

```bash
# TAG 체인 확인
rg "@CODE:REFACTOR-007" -n src/core/package-manager/
rg "@TEST:REFACTOR-007" -n src/__tests__/core/package-manager/

# 결과 예시:
# src/core/package-manager/command-builder.ts:1:// @CODE:REFACTOR-007
# src/core/package-manager/package-json-builder.ts:1:// @CODE:REFACTOR-007
# src/core/package-manager/installer.ts:1:// @CODE:REFACTOR-007
# src/__tests__/core/package-manager/command-builder.test.ts:4:* @tags @TEST:REFACTOR-007
```

## Git 커밋 전략

### 커밋 단위

각 TDD Phase별로 독립적인 커밋을 생성합니다:

1. **Phase 1 커밋**:
   ```
   feat(refactor): Add CommandBuilder with TDD

   - Add command-builder.ts (~100 LOC)
   - Add command-builder.test.ts (~100 LOC)
   - Red-Green-Refactor cycle completed
   - All tests passing

   @CODE:REFACTOR-007 @TEST:REFACTOR-007-001~015
   ```

2. **Phase 2 커밋**:
   ```
   feat(refactor): Add PackageJsonBuilder with TDD

   - Add package-json-builder.ts (~120 LOC)
   - Add package-json-builder.test.ts (~120 LOC)
   - Integrate with CommandBuilder
   - All tests passing

   @CODE:REFACTOR-007 @TEST:REFACTOR-007-101~120
   ```

3. **Phase 3 커밋**:
   ```
   refactor(core): Slim down PackageManagerInstaller

   - Refactor installer.ts (399 → 150 LOC, -62%)
   - Integrate with CommandBuilder
   - Update installer.test.ts (327 → 107 LOC)
   - All tests passing

   @CODE:REFACTOR-007 @TEST:REFACTOR-007-201~210
   ```

4. **Phase 4 커밋**:
   ```
   test(refactor): Add integration tests and verify quality

   - Add integration test suite
   - Verify coverage ≥ 85%
   - Verify TRUST principles
   - Update documentation

   @CODE:REFACTOR-007 @TEST:REFACTOR-007-301~310
   ```

## 롤백 계획

각 Phase는 독립적이므로 문제 발생 시 이전 Phase로 롤백 가능합니다:

- **Phase 1 실패**: CommandBuilder 커밋만 revert
- **Phase 2 실패**: PackageJsonBuilder 커밋만 revert
- **Phase 3 실패**: Installer 리팩토링 커밋만 revert
- **전체 롤백**: SPEC-007 브랜치 자체 삭제

## 일정 및 체크포인트

| Phase | 예상 시간 | 주요 체크포인트 | 완료 기준 |
|-------|----------|----------------|-----------|
| Phase 1 | 1시간 | CommandBuilder 구현 | 모든 테스트 통과 |
| Phase 2 | 1.5시간 | PackageJsonBuilder 구현 | 모든 테스트 통과 |
| Phase 3 | 1시간 | Installer 리팩토링 | 기존 테스트 통과 |
| Phase 4 | 0.5시간 | 통합 검증 | 커버리지 ≥ 85% |
| **합계** | **4시간** | - | **전체 품질 게이트 통과** |

## 성공 기준 체크리스트

### 코드 품질
- [ ] installer.ts ≤ 150 LOC
- [ ] command-builder.ts ≤ 100 LOC
- [ ] package-json-builder.ts ≤ 120 LOC
- [ ] 각 메서드 ≤ 50 LOC
- [ ] 매개변수 ≤ 5개
- [ ] 순환 복잡도 ≤ 10

### 테스트 품질
- [ ] 전체 테스트 통과 (0 실패)
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 각 클래스별 독립 테스트 존재
- [ ] 통합 테스트 작성

### 아키텍처 품질
- [ ] 단일 책임 원칙 준수
- [ ] 의존성 방향 명확 (순환 없음)
- [ ] 공개 API 호환성 유지
- [ ] @TAG 체인 무결성 유지

### 문서화
- [ ] 각 클래스 JSDoc 주석 작성
- [ ] 각 메서드 @tags 주석 작성
- [ ] SPEC 문서 작성 완료
- [ ] Acceptance 테스트 시나리오 작성

---

_이 계획은 TDD Red-Green-Refactor 사이클을 엄격히 준수합니다._
