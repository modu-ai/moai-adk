# SPEC-007 인수 테스트 시나리오

## @TEST:REFACTOR-007 | Chain: @SPEC:REFACTOR-007 -> @SPEC:REFACTOR-007 -> @CODE:REFACTOR-007 -> @TEST:REFACTOR-007

## 개요

이 문서는 SPEC-007 (Installer 리팩토링)의 인수 기준을 Given-When-Then 시나리오로 정의합니다.

## 시나리오 1: CommandBuilder - 명령어 생성

### 시나리오 1-1: NPM 패키지 설치 명령어 생성

**Given**: CommandBuilder 인스턴스가 생성되어 있고
**And**: 설치할 패키지 목록이 ['express', 'lodash']이고
**And**: 패키지 매니저가 NPM이고
**And**: 개발 의존성 플래그가 false일 때

**When**: buildInstallCommand() 메서드를 호출하면

**Then**: 반환된 명령어는 'npm install express lodash'이어야 함

```typescript
// @TEST:REFACTOR-007-001
test('Scenario 1-1: Generate npm install command', () => {
  // Given
  const builder = new CommandBuilder();
  const packages = ['express', 'lodash'];
  const options: PackageInstallOptions = {
    packageManager: PackageManagerType.NPM,
    isDevelopment: false,
  };

  // When
  const command = builder.buildInstallCommand(packages, options);

  // Then
  expect(command).toBe('npm install express lodash');
});
```

### 시나리오 1-2: Yarn 개발 의존성 설치 명령어 생성

**Given**: CommandBuilder 인스턴스가 생성되어 있고
**And**: 설치할 패키지 목록이 ['@types/node', 'typescript']이고
**And**: 패키지 매니저가 Yarn이고
**And**: 개발 의존성 플래그가 true일 때

**When**: buildInstallCommand() 메서드를 호출하면

**Then**: 반환된 명령어는 'yarn add --dev @types/node typescript'이어야 함

```typescript
// @TEST:REFACTOR-007-002
test('Scenario 1-2: Generate yarn dev dependency command', () => {
  // Given
  const builder = new CommandBuilder();
  const packages = ['@types/node', 'typescript'];
  const options: PackageInstallOptions = {
    packageManager: PackageManagerType.YARN,
    isDevelopment: true,
  };

  // When
  const command = builder.buildInstallCommand(packages, options);

  // Then
  expect(command).toBe('yarn add --dev @types/node typescript');
});
```

### 시나리오 1-3: PNPM 글로벌 설치 명령어 생성

**Given**: CommandBuilder 인스턴스가 생성되어 있고
**And**: 설치할 패키지 목록이 ['typescript']이고
**And**: 패키지 매니저가 PNPM이고
**And**: 글로벌 설치 플래그가 true일 때

**When**: buildInstallCommand() 메서드를 호출하면

**Then**: 반환된 명령어는 'pnpm add --global typescript'이어야 함

```typescript
// @TEST:REFACTOR-007-003
test('Scenario 1-3: Generate pnpm global install command', () => {
  // Given
  const builder = new CommandBuilder();
  const packages = ['typescript'];
  const options: PackageInstallOptions = {
    packageManager: PackageManagerType.PNPM,
    isGlobal: true,
  };

  // When
  const command = builder.buildInstallCommand(packages, options);

  // Then
  expect(command).toBe('pnpm add --global typescript');
});
```

### 시나리오 1-4: 지원하지 않는 패키지 매니저 오류

**Given**: CommandBuilder 인스턴스가 생성되어 있고
**And**: 지원하지 않는 패키지 매니저 타입이 입력될 때

**When**: buildInstallCommand() 메서드를 호출하면

**Then**: 'Unsupported package manager' 오류가 발생해야 함

```typescript
// @TEST:REFACTOR-007-004
test('Scenario 1-4: Throw error for unsupported package manager', () => {
  // Given
  const builder = new CommandBuilder();
  const packages = ['express'];
  const options: PackageInstallOptions = {
    packageManager: 'unsupported' as any,
  };

  // When & Then
  expect(() => builder.buildInstallCommand(packages, options)).toThrow(
    'Unsupported package manager'
  );
});
```

### 시나리오 1-5: 테스트 명령어 생성 (Jest)

**Given**: CommandBuilder 인스턴스가 생성되어 있고
**And**: 패키지 매니저가 NPM이고
**And**: 테스팅 프레임워크가 'jest'일 때

**When**: buildTestCommand() 메서드를 호출하면

**Then**: 반환된 명령어는 'jest'이어야 함

```typescript
// @TEST:REFACTOR-007-005
test('Scenario 1-5: Generate jest test command', () => {
  // Given
  const builder = new CommandBuilder();
  const packageManager = PackageManagerType.NPM;
  const testingFramework = 'jest';

  // When
  const command = builder.buildTestCommand(packageManager, testingFramework);

  // Then
  expect(command).toBe('jest');
});
```

## 시나리오 2: PackageJsonBuilder - 설정 생성

### 시나리오 2-1: 기본 package.json 생성

**Given**: CommandBuilder와 PackageJsonBuilder 인스턴스가 생성되어 있고
**And**: 프로젝트 구성이 { name: 'test-project', version: '1.0.0' }이고
**And**: 패키지 매니저가 NPM일 때

**When**: generatePackageJson() 메서드를 호출하면

**Then**: 반환된 package.json은 다음을 포함해야 함:
- name: 'test-project'
- version: '1.0.0'
- scripts 객체
- engines 객체 (node, npm)
- dependencies 객체
- devDependencies 객체

```typescript
// @TEST:REFACTOR-007-101
test('Scenario 2-1: Generate basic package.json', () => {
  // Given
  const commandBuilder = new CommandBuilder();
  const builder = new PackageJsonBuilder(commandBuilder);
  const projectConfig = {
    name: 'test-project',
    version: '1.0.0',
  };
  const packageManager = PackageManagerType.NPM;

  // When
  const packageJson = builder.generatePackageJson(projectConfig, packageManager);

  // Then
  expect(packageJson.name).toBe('test-project');
  expect(packageJson.version).toBe('1.0.0');
  expect(packageJson.scripts).toBeDefined();
  expect(packageJson.engines).toBeDefined();
  expect(packageJson.engines?.node).toBe('>=18.0.0');
  expect(packageJson.engines?.npm).toBe('>=9.0.0');
  expect(packageJson.dependencies).toBeDefined();
  expect(packageJson.devDependencies).toBeDefined();
});
```

### 시나리오 2-2: TypeScript 프로젝트 package.json 생성

**Given**: CommandBuilder와 PackageJsonBuilder 인스턴스가 생성되어 있고
**And**: 프로젝트 구성이 { name: 'ts-project', version: '1.0.0' }이고
**And**: 패키지 매니저가 NPM이고
**And**: TypeScript가 활성화(true)되어 있을 때

**When**: generatePackageJson() 메서드를 호출하면

**Then**: 반환된 package.json은 다음을 포함해야 함:
- devDependencies에 'typescript', '@types/node'
- scripts에 'build': 'tsc'
- scripts에 'type-check': 'tsc --noEmit'
- scripts에 'dev': 'ts-node src/index.ts'

```typescript
// @TEST:REFACTOR-007-102
test('Scenario 2-2: Generate TypeScript package.json', () => {
  // Given
  const commandBuilder = new CommandBuilder();
  const builder = new PackageJsonBuilder(commandBuilder);
  const projectConfig = {
    name: 'ts-project',
    version: '1.0.0',
  };
  const packageManager = PackageManagerType.NPM;
  const includeTypeScript = true;

  // When
  const packageJson = builder.generatePackageJson(
    projectConfig,
    packageManager,
    includeTypeScript
  );

  // Then
  expect(packageJson.devDependencies?.typescript).toBeDefined();
  expect(packageJson.devDependencies?.['@types/node']).toBeDefined();
  expect(packageJson.scripts?.build).toBe('tsc');
  expect(packageJson.scripts?.['type-check']).toBe('tsc --noEmit');
  expect(packageJson.scripts?.dev).toBe('ts-node src/index.ts');
});
```

### 시나리오 2-3: Jest 테스팅 프레임워크 포함

**Given**: CommandBuilder와 PackageJsonBuilder 인스턴스가 생성되어 있고
**And**: 프로젝트 구성이 { name: 'test-app', version: '1.0.0' }이고
**And**: 패키지 매니저가 NPM이고
**And**: TypeScript가 활성화되어 있고
**And**: 테스팅 프레임워크가 'jest'일 때

**When**: generatePackageJson() 메서드를 호출하면

**Then**: 반환된 package.json은 다음을 포함해야 함:
- devDependencies에 'jest', '@types/jest', 'ts-jest'
- scripts에 'test': 'jest'
- scripts에 'test:watch': 'jest --watch'
- scripts에 'test:coverage': 'jest --coverage'

```typescript
// @TEST:REFACTOR-007-103
test('Scenario 2-3: Include Jest testing framework', () => {
  // Given
  const commandBuilder = new CommandBuilder();
  const builder = new PackageJsonBuilder(commandBuilder);
  const projectConfig = {
    name: 'test-app',
    version: '1.0.0',
  };
  const packageManager = PackageManagerType.NPM;
  const includeTypeScript = true;
  const testingFramework = 'jest';

  // When
  const packageJson = builder.generatePackageJson(
    projectConfig,
    packageManager,
    includeTypeScript,
    testingFramework
  );

  // Then
  expect(packageJson.devDependencies?.jest).toBeDefined();
  expect(packageJson.devDependencies?.['@types/jest']).toBeDefined();
  expect(packageJson.devDependencies?.['ts-jest']).toBeDefined();
  expect(packageJson.scripts?.test).toBe('jest');
  expect(packageJson.scripts?.['test:watch']).toBe('jest --watch');
  expect(packageJson.scripts?.['test:coverage']).toBe('jest --coverage');
});
```

### 시나리오 2-4: 의존성 추가

**Given**: PackageJsonBuilder 인스턴스가 생성되어 있고
**And**: 기존 package.json이 { name: 'app', version: '1.0.0', dependencies: { express: '^4.18.0' } }이고
**And**: 추가할 의존성이 { lodash: '^4.17.21', axios: '^1.5.0' }일 때

**When**: addDependencies() 메서드를 호출하면

**Then**: 반환된 package.json의 dependencies는 다음을 포함해야 함:
- express: '^4.18.0' (기존 유지)
- lodash: '^4.17.21' (신규 추가)
- axios: '^1.5.0' (신규 추가)

```typescript
// @TEST:REFACTOR-007-104
test('Scenario 2-4: Add dependencies to existing package.json', () => {
  // Given
  const commandBuilder = new CommandBuilder();
  const builder = new PackageJsonBuilder(commandBuilder);
  const existingPackageJson = {
    name: 'app',
    version: '1.0.0',
    dependencies: {
      express: '^4.18.0',
    },
  };
  const newDependencies = {
    lodash: '^4.17.21',
    axios: '^1.5.0',
  };

  // When
  const updatedPackageJson = builder.addDependencies(
    existingPackageJson,
    newDependencies
  );

  // Then
  expect(updatedPackageJson.dependencies).toEqual({
    express: '^4.18.0',
    lodash: '^4.17.21',
    axios: '^1.5.0',
  });
});
```

## 시나리오 3: PackageManagerInstaller - 설치 오케스트레이션

### 시나리오 3-1: 패키지 설치 성공

**Given**: CommandBuilder와 PackageManagerInstaller 인스턴스가 생성되어 있고
**And**: 설치할 패키지 목록이 ['express', 'lodash']이고
**And**: 패키지 매니저가 NPM이고
**And**: execa 모의 함수가 성공 응답을 반환하도록 설정되어 있을 때

**When**: installPackages() 메서드를 호출하면

**Then**: 반환된 결과는 다음을 포함해야 함:
- success: true
- installedPackages: ['express', 'lodash']
- output: 'added 2 packages'

```typescript
// @TEST:REFACTOR-007-201
test('Scenario 3-1: Install packages successfully', async () => {
  // Given
  const commandBuilder = new CommandBuilder();
  const installer = new PackageManagerInstaller(commandBuilder);
  const packages = ['express', 'lodash'];
  const options: PackageInstallOptions = {
    packageManager: PackageManagerType.NPM,
    isDevelopment: false,
  };

  mockExeca.mockResolvedValue({
    stdout: 'added 2 packages',
    stderr: '',
    exitCode: 0,
  } as any);

  // When
  const result = await installer.installPackages(packages, options);

  // Then
  expect(result.success).toBe(true);
  expect(result.installedPackages).toEqual(['express', 'lodash']);
  expect(result.output).toBe('added 2 packages');
  expect(mockExeca).toHaveBeenCalledWith(
    'npm',
    ['install', 'express', 'lodash'],
    expect.objectContaining({
      cwd: process.cwd(),
    })
  );
});
```

### 시나리오 3-2: 패키지 설치 실패

**Given**: CommandBuilder와 PackageManagerInstaller 인스턴스가 생성되어 있고
**And**: 설치할 패키지 목록이 ['nonexistent-package']이고
**And**: 패키지 매니저가 NPM이고
**And**: execa 모의 함수가 오류를 발생시키도록 설정되어 있을 때

**When**: installPackages() 메서드를 호출하면

**Then**: 반환된 결과는 다음을 포함해야 함:
- success: false
- installedPackages: []
- error: 'Package not found'

```typescript
// @TEST:REFACTOR-007-202
test('Scenario 3-2: Handle package installation failure', async () => {
  // Given
  const commandBuilder = new CommandBuilder();
  const installer = new PackageManagerInstaller(commandBuilder);
  const packages = ['nonexistent-package'];
  const options: PackageInstallOptions = {
    packageManager: PackageManagerType.NPM,
  };

  mockExeca.mockRejectedValue(new Error('Package not found'));

  // When
  const result = await installer.installPackages(packages, options);

  // Then
  expect(result.success).toBe(false);
  expect(result.installedPackages).toEqual([]);
  expect(result.error).toBe('Package not found');
});
```

### 시나리오 3-3: 프로젝트 초기화 성공

**Given**: CommandBuilder와 PackageManagerInstaller 인스턴스가 생성되어 있고
**And**: 프로젝트 경로가 '/test/new-project'이고
**And**: 패키지 매니저가 Yarn이고
**And**: execa 모의 함수가 성공 응답을 반환하도록 설정되어 있을 때

**When**: initializeProject() 메서드를 호출하면

**Then**: 반환된 결과는 다음을 포함해야 함:
- success: true
- packageJsonPath: '/test/new-project/package.json'
- output: 'success Saved package.json'

```typescript
// @TEST:REFACTOR-007-203
test('Scenario 3-3: Initialize project successfully', async () => {
  // Given
  const commandBuilder = new CommandBuilder();
  const installer = new PackageManagerInstaller(commandBuilder);
  const projectPath = '/test/new-project';
  const packageManager = PackageManagerType.YARN;

  mockExeca.mockResolvedValue({
    stdout: 'success Saved package.json',
    stderr: '',
    exitCode: 0,
  } as any);

  // When
  const result = await installer.initializeProject(projectPath, packageManager);

  // Then
  expect(result.success).toBe(true);
  expect(result.packageJsonPath).toBe('/test/new-project/package.json');
  expect(result.output).toBe('success Saved package.json');
  expect(mockExeca).toHaveBeenCalledWith(
    'yarn',
    ['init', '-y'],
    expect.objectContaining({
      cwd: projectPath,
    })
  );
});
```

## 시나리오 4: 통합 - 전체 워크플로우

### 시나리오 4-1: TypeScript 프로젝트 전체 설정

**Given**: 모든 클래스 인스턴스가 생성되어 있고
**And**: 프로젝트 구성이 { name: 'my-ts-app', version: '1.0.0' }이고
**And**: TypeScript와 Jest를 사용하기로 결정했을 때

**When**: 전체 프로젝트 설정 워크플로우를 실행하면

**Then**:
1. PackageJsonBuilder가 올바른 package.json을 생성해야 함
2. PackageManagerInstaller가 초기화 명령을 성공적으로 실행해야 함
3. PackageManagerInstaller가 의존성 설치 명령을 성공적으로 실행해야 함

```typescript
// @TEST:REFACTOR-007-301
test('Scenario 4-1: Full TypeScript project setup', async () => {
  // Given
  const commandBuilder = new CommandBuilder();
  const packageJsonBuilder = new PackageJsonBuilder(commandBuilder);
  const installer = new PackageManagerInstaller(commandBuilder);

  const projectConfig = {
    name: 'my-ts-app',
    version: '1.0.0',
  };
  const packageManager = PackageManagerType.NPM;
  const includeTypeScript = true;
  const testingFramework = 'jest';

  // When: Generate package.json
  const packageJson = packageJsonBuilder.generatePackageJson(
    projectConfig,
    packageManager,
    includeTypeScript,
    testingFramework
  );

  // Then: Verify package.json structure
  expect(packageJson.name).toBe('my-ts-app');
  expect(packageJson.devDependencies?.typescript).toBeDefined();
  expect(packageJson.devDependencies?.jest).toBeDefined();
  expect(packageJson.scripts?.build).toBe('tsc');
  expect(packageJson.scripts?.test).toBe('jest');

  // And: Verify command builder integration
  const installCommand = commandBuilder.buildInstallCommand(
    ['typescript', 'jest'],
    { packageManager, isDevelopment: true }
  );
  expect(installCommand).toBe('npm install --save-dev typescript jest');
});
```

## 성공 기준 요약

### 기능 요구사항
- [ ] 모든 시나리오 테스트 통과
- [ ] CommandBuilder 14개 테스트 통과
- [ ] PackageJsonBuilder 12개 테스트 통과
- [ ] PackageManagerInstaller 6개 테스트 통과
- [ ] 통합 테스트 1개 통과

### 품질 요구사항
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 모든 테스트 실행 시간 < 5초
- [ ] 0개의 테스트 실패
- [ ] 0개의 ESLint 오류

### 아키텍처 요구사항
- [ ] 단일 책임 원칙 준수
- [ ] 의존성 주입 패턴 적용
- [ ] 순환 의존성 없음
- [ ] @TAG 체인 무결성 유지

---

_이 인수 테스트는 SPEC-007의 완료 기준을 정의합니다._
