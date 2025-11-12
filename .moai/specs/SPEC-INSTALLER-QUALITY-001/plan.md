
## 1. Phase 1: 상수 추출 및 정의

### 1.1 installer-constants.ts 생성
- **파일 경로**: `moai-adk-ts/src/core/installer/installer-constants.ts`
- **테스트 파일**: `moai-adk-ts/tests/core/installer/installer-constants.test.ts`

### 1.2 구현 내용
```typescript

export const TIMEOUTS = {
  DEPENDENCY_INSTALL: 5 * 60 * 1000,
  TEMPLATE_COPY: 30 * 1000,
  CLEANUP: 5 * 1000,
} as const;

export const RETRY_LIMITS = {
  PACKAGE_INSTALL: 3,
  NETWORK_REQUEST: 5,
} as const;

export const DEFAULT_DIRECTORIES = {
  SPEC_DIR: '.moai/specs',
  MEMORY_DIR: '.moai/memory',
  PRODUCT_DIR: '.moai/product',
  TEMPLATE_DIR: '.moai/templates',
} as const;

export const PACKAGE_MANAGERS = {
  NPM: 'npm',
  PNPM: 'pnpm',
  YARN: 'yarn',
  BUN: 'bun',
} as const;
```

### 1.3 매직 넘버 식별
```bash
# 숫자 리터럴 찾기
rg '\b[0-9]+\b' moai-adk-ts/src/core/installer/ -n

# 문자열 리터럴 중복 찾기
rg '".moai' moai-adk-ts/src/core/installer/ -n
rg '"npm"' moai-adk-ts/src/core/installer/ -n
```

## 2. Phase 2: 에러 클래스 계층 구조 구현

### 2.1 installation-error.ts 생성
- **파일 경로**: `moai-adk-ts/src/core/errors/installation-error.ts`
- **테스트 파일**: `moai-adk-ts/tests/core/errors/installation-error.test.ts`

### 2.2 에러 클래스 목록
```typescript

export class InstallationError extends Error {
  constructor(message: string, public readonly cause?: Error);
}

export class DependencyInstallationError extends InstallationError {}
export class TemplateInstallationError extends InstallationError {}
export class ConfigurationError extends InstallationError {}
export class ValidationError extends InstallationError {}
export class FileSystemError extends InstallationError {}
export class NetworkError extends InstallationError {}
```

### 2.3 기존 에러 마이그레이션
```bash
# 기존 Error throw 찾기
rg 'throw new Error' moai-adk-ts/src/core/installer/ -n

# 각 파일별로 적절한 InstallationError 하위 클래스로 변경
```

## 3. Phase 3: DI 패턴 통일

### 3.1 DI 리팩토링 순서 (의존성 역순)
1. **Leaf 클래스** (의존성 없음):
   - package-manager.ts
   - context-manager.ts

2. **중간 클래스**:
   - dependency-installer.ts (PackageManager 의존)
   - template-installer.ts
   - typescript-setup.ts

3. **상위 클래스**:
   - phase-executor.ts (모든 installer 의존)
   - update-executor.ts

4. **최상위 클래스**:
   - installer-core.ts (PhaseExecutor 의존)
   - update-manager.ts

### 3.2 DI 변환 템플릿

#### Before
```typescript
export class PhaseExecutor {
  private dependencyInstaller = new DependencyInstaller();

  async execute() {
    await this.dependencyInstaller.install();
  }
}
```

#### After
```typescript
export class PhaseExecutor {
  constructor(
    private readonly dependencyInstaller: DependencyInstaller,
    private readonly logger: Logger = new ConsoleLogger()
  ) {}

  async execute() {
    await this.dependencyInstaller.install();
  }
}
```

### 3.3 Factory 패턴 도입
```typescript

export class InstallerFactory {
  static create(): InstallerCore {
    // Leaf dependencies
    const packageManager = new PackageManager();
    const contextManager = new ContextManager();
    const logger = new ConsoleLogger();

    // Mid-level dependencies
    const dependencyInstaller = new DependencyInstaller(packageManager, logger);
    const templateInstaller = new TemplateInstaller(logger);
    const typescriptSetup = new TypeScriptSetup(logger);

    // High-level dependencies
    const phaseExecutor = new PhaseExecutor(
      dependencyInstaller,
      templateInstaller,
      typescriptSetup,
      logger
    );

    // Top-level
    return new InstallerCore(phaseExecutor, contextManager, logger);
  }
}
```

## 4. Phase 4: TAG 형식 통일

### 4.1 TAG 스캔 및 검증
```bash
# 현재 TAG 형식 확인

# 통일된 형식이 아닌 TAG 찾기
```

### 4.2 TAG 자동 업데이트 스크립트
```typescript
// scripts/fix-tags.ts

import { readFileSync, writeFileSync } from 'fs';
import { globSync } from 'glob';

const files = globSync('moai-adk-ts/src/core/installer/**/*.ts');

for (const file of files) {
  let content = readFileSync(file, 'utf-8');

  content = content.replace(
  );

  writeFileSync(file, content);
}
```

### 4.3 수동 검토 필요 파일
- installer-core.ts → SPEC-INSTALLER-QUALITY-001.md
- phase-executor.ts → SPEC-INSTALLER-QUALITY-001.md
- 기타 파일들 → 각 파일의 주 기능에 맞는 SPEC 참조

## 5. Phase 5: 크로스 플랫폼 개선

### 5.1 cross-spawn 의존성 추가
```bash
cd moai-adk-ts
pnpm add cross-spawn
pnpm add -D @types/cross-spawn
```

### 5.2 package-manager.ts 수정
```typescript
// Before
import { execSync } from 'child_process';

detectPackageManager(): string {
  if (execSync('which pnpm')) return 'pnpm';
  if (execSync('which yarn')) return 'yarn';
  return 'npm';
}

// After
import spawn from 'cross-spawn';

detectPackageManager(): string {
  if (this.hasCommand('pnpm')) return 'pnpm';
  if (this.hasCommand('yarn')) return 'yarn';
  return 'npm';
}

private hasCommand(cmd: string): boolean {
  const result = spawn.sync(cmd, ['--version'], {
    stdio: 'ignore',
    windowsHide: true
  });
  return result.status === 0;
}
```

### 5.3 영향받는 파일
- package-manager.ts
- dependency-installer.ts
- typescript-setup.ts

## 6. Phase 6: 테스트 작성

### 6.1 새로운 테스트 파일
- installer-constants.test.ts
- installation-error.test.ts
- installer-factory.test.ts

### 6.2 기존 테스트 업데이트
- 모든 installer 테스트에 DI 모킹 추가
- 에러 클래스 변경에 따른 테스트 수정
- 크로스 플랫폼 테스트 추가

### 6.3 테스트 우선순위
1. **Critical**: installer-constants, installation-error
2. **High**: DI 패턴 변경된 클래스들
3. **Medium**: TAG 형식 검증
4. **Low**: 크로스 플랫폼 통합 테스트

## 7. Phase 7: 문서화 및 마이그레이션 가이드

### 7.1 MIGRATION.md 작성
```markdown
# Installer 패키지 품질 개선 마이그레이션 가이드

## 변경 사항

### 1. 에러 처리
- Before: `throw new Error('...')`
- After: `throw new DependencyInstallationError('...', cause)`

### 2. DI 패턴
- Before: `new DependencyInstaller()`
- After: `InstallerFactory.create()` 또는 생성자 주입

### 3. 상수 사용
- Before: `setTimeout(..., 5000)`
- After: `setTimeout(..., TIMEOUTS.CLEANUP)`
```

### 7.2 README.md 업데이트
- 새로운 에러 클래스 문서화
- DI 사용 예제 추가
- 상수 목록 문서화

## 8. 체크리스트

### Phase 1: 상수
- [ ] installer-constants.ts 생성
- [ ] 매직 넘버 식별 및 추출
- [ ] 기존 코드 상수 참조로 변경

### Phase 2: 에러
- [ ] InstallationError 계층 구조 구현
- [ ] 모든 throw new Error 마이그레이션
- [ ] 에러 테스트 작성

### Phase 3: DI
- [ ] Leaf 클래스 DI 적용
- [ ] 중간 클래스 DI 적용
- [ ] 상위 클래스 DI 적용
- [ ] InstallerFactory 구현

### Phase 4: TAG
- [ ] TAG 형식 통일 스크립트 실행
- [ ] 수동 검토 및 수정
- [ ] TAG 검증 테스트 추가

### Phase 5: 크로스 플랫폼
- [ ] cross-spawn 의존성 추가
- [ ] which 명령 대체
- [ ] Windows 환경 테스트

### Phase 6: 테스트
- [ ] 새 테스트 파일 작성
- [ ] 기존 테스트 업데이트
- [ ] 커버리지 85% 달성

### Phase 7: 문서화
- [ ] MIGRATION.md 작성
- [ ] README.md 업데이트
- [ ] 코드 주석 개선

## 9. 일정

- **Phase 1**: 0.5일 (상수)
- **Phase 2**: 1일 (에러)
- **Phase 3**: 2일 (DI)
- **Phase 4**: 0.5일 (TAG)
- **Phase 5**: 1일 (크로스 플랫폼)
- **Phase 6**: 1.5일 (테스트)
- **Phase 7**: 0.5일 (문서화)
- **Total**: 7일

## 10. 롤아웃 전략

### 10.1 Stage 1: Non-breaking Changes
- 상수 추출
- TAG 형식 통일
- 에러 클래스 추가 (기존 Error도 유지)

### 10.2 Stage 2: DI Migration
- Factory 패턴 도입
- 점진적 DI 적용
- 하위 호환성 유지

### 10.3 Stage 3: Breaking Changes
- 기존 Error throw 제거
- 직접 생성자 호출 금지
- 전체 테스트 검증
