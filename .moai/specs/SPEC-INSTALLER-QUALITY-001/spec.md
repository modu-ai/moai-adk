---
id: INSTALLER-QUALITY-001
version: 0.2.0
status: completed
created: 2025-10-06
updated: 2025-10-18
author: @Goos
priority: medium
---

# @SPEC:INSTALLER-QUALITY-001: Code Quality Integrated Improvements

## HISTORY

### v0.2.0 (2025-10-18)
- **CHANGED**: deprecated → completed (TypeScript 프로젝트 아카이브)
- **AUTHOR**: @Goos
- **REASON**: TypeScript 프로젝트에서 구현 완료된 기능, Python 전환으로 deprecated 처리했으나 실제로는 완료된 것으로 간주

### v0.1.0 (2025-10-16)
- **DEPRECATED**: TypeScript 프로젝트용 SPEC, Python 프로젝트에는 적용 불가
- **AUTHOR**: @Goos
- **REASON**: MoAI-ADK가 Python 프로젝트로 전환됨에 따라 TypeScript installer 품질 SPEC 불필요
- **ALTERNATIVE**: Python 프로젝트는 trust_checker.py로 TRUST 5원칙 검증 (85% 커버리지, LOC 제한 등)

### v0.0.1 (2025-10-06)
- **INITIAL**: Installer 패키지 통합 코드 품질 개선 명세 작성 (TypeScript용)
- **AUTHOR**: @Goos
- **SCOPE**: TypeScript installer 패키지 품질 개선

## 1. 개요

### 1.1 목적
Installer 패키지의 코드 품질을 전반적으로 개선하여 일관성, 가독성, 유지보수성을 향상시킨다.

### 1.2 범위
- **DI 패턴 통일**: 모든 클래스에 일관된 의존성 주입 적용
- **에러 처리 일관성**: InstallationError 계층 구조 통일
- **TAG 형식 통일**: `@CODE:ID | SPEC: ... | TEST: ...` 형식
- **크로스 플랫폼 개선**: Windows 'which' 명령 대체
- **매직 넘버 제거**: 하드코딩된 값을 상수로 추출

### 1.3 제외 사항
- 새로운 기능 추가
- 아키텍처 재설계
- 성능 최적화 (별도 SPEC에서 다룸)

## 2. EARS 요구사항

### 2.1 Ubiquitous Requirements

**REQ-QUALITY-001**: 시스템은 모든 클래스에서 생성자 기반 의존성 주입을 사용해야 한다.

**REQ-QUALITY-002**: 시스템은 모든 에러를 InstallationError 또는 하위 클래스로 통일해야 한다.

**REQ-QUALITY-003**: 시스템은 모든 코드 파일에 통일된 TAG 형식을 사용해야 한다.

**REQ-QUALITY-004**: 시스템은 크로스 플랫폼 명령을 위해 cross-spawn 라이브러리를 사용해야 한다.

**REQ-QUALITY-005**: 시스템은 매직 넘버 대신 명명된 상수를 사용해야 한다.

### 2.2 Event-driven Requirements

**REQ-QUALITY-010**: WHEN 새로운 클래스를 추가하면, 시스템은 생성자에서 의존성을 주입받아야 한다.

**REQ-QUALITY-011**: WHEN 에러가 발생하면, 시스템은 적절한 InstallationError 하위 클래스를 사용해야 한다.

**REQ-QUALITY-012**: WHEN 외부 명령을 실행하면, 시스템은 cross-spawn을 사용해야 한다.

### 2.3 State-driven Requirements

**REQ-QUALITY-020**: WHILE 코드 파일이 존재할 때, 시스템은 TAG BLOCK을 포함해야 한다.

**REQ-QUALITY-021**: WHILE 타임아웃 값을 사용할 때, 시스템은 명명된 상수를 참조해야 한다.

### 2.4 Optional Requirements

**REQ-QUALITY-030**: WHERE 복잡한 DI가 필요하면, 시스템은 DI 컨테이너를 도입할 수 있다.

**REQ-QUALITY-031**: WHERE 에러 컨텍스트가 필요하면, 시스템은 structured error를 사용할 수 있다.

### 2.5 Constraints

**REQ-QUALITY-040**: IF 하드코딩된 값이 2회 이상 사용되면, 시스템은 상수로 추출해야 한다.

**REQ-QUALITY-041**: IF 에러를 throw할 때, 시스템은 원본 에러를 cause로 포함해야 한다.

**REQ-QUALITY-042**: IF Windows 환경을 지원하면, 시스템은 플랫폼별 분기 없이 동작해야 한다.

## 3. 기술 상세

### 3.1 DI 패턴 통일

#### Before (Mixed Pattern)
```typescript
export class PhaseExecutor {
  private dependencyInstaller = new DependencyInstaller();
  private templateInstaller = new TemplateInstaller();

  execute(phase: Phase) {
    // ...
  }
}
```

#### After (Constructor Injection)
```typescript
// @CODE:INSTALLER-QUALITY-001 | SPEC: SPEC-INSTALLER-QUALITY-001.md | TEST: tests/core/installer/phase-executor.test.ts

export class PhaseExecutor {
  constructor(
    private readonly dependencyInstaller: DependencyInstaller,
    private readonly templateInstaller: TemplateInstaller,
    private readonly logger: Logger
  ) {}

  execute(phase: Phase) {
    // ...
  }
}
```

### 3.2 에러 처리 통일

#### InstallationError 계층 구조
```typescript
// @CODE:INSTALLER-QUALITY-001 | SPEC: SPEC-INSTALLER-QUALITY-001.md | TEST: tests/core/errors/installation-error.test.ts

export class InstallationError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'InstallationError';
  }
}

export class DependencyInstallationError extends InstallationError {
  constructor(message: string, cause?: Error) {
    super(message, cause);
    this.name = 'DependencyInstallationError';
  }
}

export class TemplateInstallationError extends InstallationError {
  constructor(message: string, cause?: Error) {
    super(message, cause);
    this.name = 'TemplateInstallationError';
  }
}

export class ConfigurationError extends InstallationError {
  constructor(message: string, cause?: Error) {
    super(message, cause);
    this.name = 'ConfigurationError';
  }
}
```

#### 사용 예시
```typescript
try {
  await installDependencies();
} catch (error) {
  throw new DependencyInstallationError(
    'Failed to install npm packages',
    error instanceof Error ? error : undefined
  );
}
```

### 3.3 TAG 형식 통일

#### 표준 TAG 형식
```typescript
// @CODE:INSTALLER-QUALITY-001 | SPEC: SPEC-INSTALLER-QUALITY-001.md | TEST: tests/core/installer/installer-core.test.ts

export class InstallerCore {
  // ...
}
```

#### 일관성 검증 스크립트
```bash
# TAG 형식 검증
rg '@CODE:[A-Z]+-[0-9]{3}' moai-adk-ts/src/ --count
rg '@CODE:[A-Z]+-[0-9]{3} \| SPEC:' moai-adk-ts/src/ --count

# 불일치 찾기
rg '@CODE:[A-Z]+-[0-9]{3}(?! \| SPEC:)' moai-adk-ts/src/
```

### 3.4 크로스 플랫폼 개선

#### Before (Unix-only)
```typescript
import { exec } from 'child_process';

const hasCommand = (cmd: string): boolean => {
  try {
    execSync(`which ${cmd}`); // Windows에서 실패
    return true;
  } catch {
    return false;
  }
};
```

#### After (Cross-platform)
```typescript
import spawn from 'cross-spawn';

const hasCommand = (cmd: string): boolean => {
  const result = spawn.sync(cmd, ['--version'], {
    stdio: 'ignore',
    windowsHide: true
  });
  return result.status === 0;
};
```

### 3.5 매직 넘버 제거

#### Before
```typescript
setTimeout(() => {
  // cleanup
}, 5000); // 5초는 무엇을 의미?

const MAX_RETRIES = 3; // 왜 3번?
```

#### After
```typescript
// @CODE:INSTALLER-QUALITY-001 | SPEC: SPEC-INSTALLER-QUALITY-001.md

export const TIMEOUTS = {
  /** Dependency 설치 타임아웃 (5분) */
  DEPENDENCY_INSTALL: 5 * 60 * 1000,

  /** 템플릿 복사 타임아웃 (30초) */
  TEMPLATE_COPY: 30 * 1000,

  /** 정리 작업 타임아웃 (5초) */
  CLEANUP: 5 * 1000,
} as const;

export const RETRY_LIMITS = {
  /** 패키지 설치 재시도 횟수 */
  PACKAGE_INSTALL: 3,

  /** 네트워크 요청 재시도 횟수 */
  NETWORK_REQUEST: 5,
} as const;

// 사용
setTimeout(cleanup, TIMEOUTS.CLEANUP);
for (let i = 0; i < RETRY_LIMITS.PACKAGE_INSTALL; i++) {
  // ...
}
```

## 4. 영향받는 파일 목록

### 4.1 DI 패턴 통일 (9개 파일)
- installer-core.ts
- phase-executor.ts
- update-executor.ts
- dependency-installer.ts
- template-installer.ts
- package-manager.ts
- context-manager.ts
- typescript-setup.ts
- update-manager.ts

### 4.2 에러 처리 통일 (12개 파일)
- 모든 installer 패키지 파일

### 4.3 TAG 형식 통일 (12개 파일)
- 모든 installer 패키지 파일

### 4.4 크로스 플랫폼 개선 (3개 파일)
- package-manager.ts
- dependency-installer.ts
- typescript-setup.ts

### 4.5 매직 넘버 제거 (전체)
- 새 파일: installer-constants.ts

## 5. 성공 기준

### 5.1 정량적 지표
- [ ] 100% 생성자 기반 DI 사용
- [ ] 100% InstallationError 계열 사용
- [ ] 100% TAG 형식 통일
- [ ] 0개 매직 넘버 (2회 이상 사용된 값)

### 5.2 정성적 지표
- [ ] Windows에서 모든 명령 정상 동작
- [ ] 에러 메시지 일관성 및 명확성
- [ ] 코드 가독성 향상
- [ ] 테스트 가능성 향상 (DI 덕분)

## 6. 참조

### 6.1 관련 SPEC
- `SPEC-REFACTOR-001`: Installer 패키지 리팩토링 (선행 작업)
- `SPEC-INSTALLER-TEST-001`: 테스트 커버리지 (병행 작업)
- `SPEC-INSTALLER-ROLLBACK-001`: 롤백 메커니즘 (병행 작업)

### 6.2 관련 문서
- `.moai/memory/development-guide.md`: 코딩 규칙, TRUST 원칙
- `moai-adk-ts/package.json`: cross-spawn 의존성

### 6.3 관련 TAG
- `@CODE:INSTALLER-QUALITY-001`: 품질 개선 코드
- `@TEST:INSTALLER-QUALITY-001`: 품질 개선 테스트
