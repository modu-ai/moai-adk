# SPEC-007: Installer 클래스 리팩토링

## @SPEC:REFACTOR-007 | Chain: @SPEC:REFACTOR-007 -> @SPEC:REFACTOR-007 -> @CODE:REFACTOR-007 -> @TEST:REFACTOR-007

## 개요

**목적**: PackageManagerInstaller 클래스(399 LOC)를 단일 책임 원칙에 따라 3개의 클래스로 분리하여 유지보수성과 테스트 용이성을 향상시킨다.

**배경**: 현재 installer.ts는 명령어 생성, package.json 관리, 패키지 설치 오케스트레이션을 모두 담당하고 있어 복잡도가 높고 테스트가 어렵다.

**성공 기준**:
- installer.ts ≤ 150 LOC
- 3개 클래스로 명확한 책임 분리
- 기존 테스트 327줄 모두 통과
- 테스트 커버리지 ≥ 85%
- @TAG 체인 무결성 유지

## EARS 요구사항

### Ubiquitous Requirements (기본 요구사항)

- 시스템은 CommandBuilder 클래스를 제공하여 패키지 매니저별 명령어 문자열을 생성해야 한다
- 시스템은 PackageJsonBuilder 클래스를 제공하여 package.json 구성을 생성/관리해야 한다
- 시스템은 PackageManagerInstaller 클래스를 제공하여 패키지 설치 오케스트레이션을 담당해야 한다
- 각 클래스는 단일 책임 원칙을 준수하여 하나의 명확한 역할만 수행해야 한다

### Event-driven Requirements (이벤트 기반)

- WHEN 패키지 설치가 요청되면, PackageManagerInstaller는 CommandBuilder로 명령어를 생성하고 실행해야 한다
- WHEN package.json 생성이 요청되면, PackageJsonBuilder는 프로젝트 구성에 맞는 설정을 생성해야 한다
- WHEN 의존성 추가가 요청되면, PackageJsonBuilder는 기존 설정을 유지하면서 새 의존성을 병합해야 한다
- WHEN 명령어 생성 중 오류가 발생하면, 시스템은 명확한 오류 메시지를 반환해야 한다

### State-driven Requirements (상태 기반)

- WHILE 개발 의존성 모드일 때, CommandBuilder는 --save-dev 또는 --dev 플래그를 포함해야 한다
- WHILE 글로벌 설치 모드일 때, CommandBuilder는 --global 플래그를 포함해야 한다
- WHILE TypeScript가 활성화된 상태일 때, PackageJsonBuilder는 TypeScript 관련 의존성과 스크립트를 포함해야 한다
- WHILE 테스팅 프레임워크가 지정된 상태일 때, PackageJsonBuilder는 해당 프레임워크 의존성을 포함해야 한다

### Optional Features (선택적 기능)

- WHERE 작업 디렉토리가 지정되면, 시스템은 해당 디렉토리에서 명령을 실행할 수 있다
- WHERE 타임아웃이 지정되면, 시스템은 해당 시간 제한을 적용할 수 있다

### Constraints (제약사항)

- IF 지원하지 않는 패키지 매니저가 지정되면, 시스템은 명확한 오류를 발생시켜야 한다
- 각 클래스 파일은 300 LOC를 초과하지 않아야 한다
- 각 메서드는 50 LOC를 초과하지 않아야 한다
- 메서드 매개변수는 5개를 초과하지 않아야 한다
- 순환 복잡도는 10을 초과하지 않아야 한다

## 리팩토링 목표

### 현재 상태 (Before)

```
installer.ts (399 LOC)
├── installPackages()           # 패키지 설치 + 명령어 생성
├── generatePackageJson()       # package.json 생성
├── addDependencies()           # 의존성 추가
├── addDevDependencies()        # 개발 의존성 추가
├── initializeProject()         # 프로젝트 초기화
├── buildInstallCommand()       # 설치 명령어 생성
├── generateScripts()           # scripts 섹션 생성
├── getRunCommand()             # 실행 명령어 조회
├── getTestCommand()            # 테스트 명령어 조회
├── getInitCommand()            # 초기화 명령어 조회
└── getPackageManagerEngine()   # 엔진 요구사항 조회
```

**문제점**:
- 단일 클래스가 3가지 책임을 동시에 처리
- 명령어 생성 로직이 여러 private 메서드에 분산
- package.json 생성과 의존성 관리가 혼재
- 테스트 작성 시 모든 책임을 동시에 고려 필요

### 목표 상태 (After)

```
1. CommandBuilder (~100 LOC)
   ├── buildInstallCommand()
   ├── buildRunCommand()
   ├── buildTestCommand()
   ├── buildInitCommand()
   └── getPackageManagerEngine()

2. PackageJsonBuilder (~120 LOC)
   ├── generatePackageJson()
   ├── generateScripts()
   ├── addDependencies()
   └── addDevDependencies()

3. PackageManagerInstaller (~150 LOC)
   ├── installPackages()
   └── initializeProject()
```

**개선점**:
- 명확한 책임 분리 (명령어 생성 / 설정 관리 / 실행 오케스트레이션)
- 각 클래스를 독립적으로 테스트 가능
- 의존성 주입을 통한 느슨한 결합
- 확장성 향상 (새로운 패키지 매니저 추가 시 CommandBuilder만 수정)

## 클래스별 상세 설계

### 1. CommandBuilder (신규)

**책임**: 패키지 매니저별 명령어 문자열 생성

**Public API**:
```typescript
class CommandBuilder {
  buildInstallCommand(
    packages: string[],
    options: PackageInstallOptions
  ): string;

  buildRunCommand(packageManagerType: PackageManagerType): string;

  buildTestCommand(
    packageManagerType: PackageManagerType,
    testingFramework?: string
  ): string;

  buildInitCommand(packageManagerType: PackageManagerType): string;

  getPackageManagerEngine(
    packageManagerType: PackageManagerType
  ): Record<string, string>;
}
```

**주요 로직**:
- npm/yarn/pnpm 명령어 패턴 매핑
- 플래그 조합 (--save-dev, --global 등)
- 명령어 문자열 조립

### 2. PackageJsonBuilder (신규)

**책임**: package.json 구성 생성 및 의존성 관리

**Public API**:
```typescript
class PackageJsonBuilder {
  private commandBuilder: CommandBuilder;

  constructor(commandBuilder: CommandBuilder);

  generatePackageJson(
    projectConfig: Partial<PackageJsonConfig>,
    packageManagerType: PackageManagerType,
    includeTypeScript?: boolean,
    testingFramework?: string
  ): PackageJsonConfig;

  generateScripts(
    packageManagerType: PackageManagerType,
    includeTypeScript: boolean,
    testingFramework?: string
  ): Record<string, string>;

  addDependencies(
    existingPackageJson: Partial<PackageJsonConfig>,
    newDependencies: Record<string, string>
  ): PackageJsonConfig;

  addDevDependencies(
    existingPackageJson: Partial<PackageJsonConfig>,
    newDevDependencies: Record<string, string>
  ): PackageJsonConfig;
}
```

**주요 로직**:
- 기본 package.json 템플릿 생성
- TypeScript/테스팅 프레임워크 의존성 추가
- scripts 섹션 생성 (CommandBuilder 활용)
- 의존성 병합

### 3. PackageManagerInstaller (리팩토링)

**책임**: 패키지 설치 및 프로젝트 초기화 오케스트레이션

**Public API**:
```typescript
class PackageManagerInstaller {
  private commandBuilder: CommandBuilder;

  constructor(commandBuilder?: CommandBuilder);

  async installPackages(
    packages: string[],
    options: PackageInstallOptions
  ): Promise<InstallResult>;

  async initializeProject(
    projectPath: string,
    packageManagerType: PackageManagerType
  ): Promise<InitResult>;
}
```

**주요 로직**:
- CommandBuilder로 명령어 생성
- execa를 통한 명령어 실행
- 결과 처리 및 오류 핸들링
- 타임아웃 관리

## 의존성 관계

```
PackageManagerInstaller (오케스트레이션)
    ↓ 사용
CommandBuilder (명령어 생성)

PackageJsonBuilder (설정 생성)
    ↓ 사용
CommandBuilder (명령어 조회)
```

**특징**:
- CommandBuilder는 의존성 없음 (순수 함수형)
- PackageJsonBuilder는 CommandBuilder 의존
- PackageManagerInstaller는 CommandBuilder 의존
- 순환 의존성 없음

## 호환성 전략

### 기존 API 보장

**installer.ts 공개 API 유지**:
```typescript
// Before (현재)
const installer = new PackageManagerInstaller();
await installer.installPackages(packages, options);
const packageJson = installer.generatePackageJson(config, type);

// After (리팩토링 후)
const installer = new PackageManagerInstaller(); // 동일한 API
await installer.installPackages(packages, options); // 동일한 API
const builder = new PackageJsonBuilder(new CommandBuilder());
const packageJson = builder.generatePackageJson(config, type); // 새 클래스 사용
```

**변경 최소화**:
- PackageManagerInstaller의 `installPackages()`, `initializeProject()` API 유지
- `generatePackageJson()` 등은 PackageJsonBuilder로 이전
- 기존 테스트는 import 경로만 수정하면 대부분 통과

### 마이그레이션 가이드

1. **Phase 1**: 새 클래스 추가 (CommandBuilder, PackageJsonBuilder)
2. **Phase 2**: PackageManagerInstaller에 새 클래스 통합
3. **Phase 3**: 불필요한 private 메서드 제거
4. **Phase 4**: 테스트 업데이트 및 검증

## 테스트 전략

### 테스트 구조 변경

**현재 (327 LOC, 단일 파일)**:
```
installer.test.ts
├── Package Installation (56 LOC)
├── Package.json Generation (83 LOC)
├── Dependency Management (63 LOC)
└── Project Initialization (48 LOC)
```

**목표 (3개 파일로 분리)**:
```
1. command-builder.test.ts (~100 LOC)
   ├── Install Command Building
   ├── Run Command Building
   ├── Test Command Building
   ├── Init Command Building
   └── Engine Requirements

2. package-json-builder.test.ts (~120 LOC)
   ├── Package.json Generation
   ├── Scripts Generation
   ├── Dependency Management
   └── TypeScript/Testing Integration

3. installer.test.ts (~107 LOC) - 기존 파일 슬림화
   ├── Package Installation
   └── Project Initialization
```

### TDD 순서

**Bottom-up 접근**:
1. **CommandBuilder TDD** (의존성 없음)
   - RED: 명령어 생성 테스트 작성
   - GREEN: CommandBuilder 구현
   - REFACTOR: 중복 제거, 패턴 개선

2. **PackageJsonBuilder TDD** (CommandBuilder 의존)
   - RED: package.json 생성 테스트 작성
   - GREEN: PackageJsonBuilder 구현
   - REFACTOR: 구조 개선

3. **PackageManagerInstaller 리팩토링** (CommandBuilder 의존)
   - RED: 기존 테스트 업데이트
   - GREEN: 의존성 주입 방식으로 변경
   - REFACTOR: 불필요한 코드 제거

## 품질 게이트

### 코드 품질

- [ ] 각 클래스 ≤ 300 LOC
- [ ] 각 메서드 ≤ 50 LOC
- [ ] 매개변수 ≤ 5개
- [ ] 순환 복잡도 ≤ 10
- [ ] 중복 코드 없음

### 테스트 품질

- [ ] 테스트 커버리지 ≥ 85%
- [ ] 모든 기존 테스트 통과
- [ ] 각 클래스별 독립적 테스트 작성
- [ ] 모의 객체 사용 최소화

### 아키텍처 품질

- [ ] 단일 책임 원칙 준수
- [ ] 의존성 방향 명확 (순환 없음)
- [ ] 공개 API 호환성 유지
- [ ] @TAG 체인 무결성 유지

## 위험 요소 및 완화 방안

### 위험 1: 기존 API 호환성 깨짐

**위험도**: 중간
**영향**: 다른 모듈에서 installer.ts 사용 시 오류 발생

**완화 방안**:
- PackageManagerInstaller의 핵심 API 유지
- 기존 테스트를 먼저 통과시킨 후 새 기능 추가
- 단계적 마이그레이션 (한 번에 하나씩)

### 위험 2: 테스트 복잡도 증가

**위험도**: 낮음
**영향**: 테스트 유지보수 비용 증가

**완화 방안**:
- 의존성 주입으로 모의 객체 사용 최소화
- 각 클래스별 독립적 테스트 작성
- Given-When-Then 패턴 일관성 유지

### 위험 3: LOC 목표 미달성

**위험도**: 낮음
**영향**: 리팩토링 효과 감소

**완화 방안**:
- 엄격한 LOC 제한 준수
- 중복 코드 철저히 제거
- 필요 시 추가 헬퍼 함수 분리

## 구현 순서

### Phase 1: 기반 클래스 구현 (CommandBuilder)

1. `command-builder.ts` 생성
2. TDD로 명령어 생성 로직 구현
3. 테스트 작성 및 검증

**예상 시간**: 1시간
**LOC**: ~100 LOC (클래스), ~100 LOC (테스트)

### Phase 2: 설정 관리 클래스 구현 (PackageJsonBuilder)

1. `package-json-builder.ts` 생성
2. CommandBuilder 의존성 주입
3. TDD로 package.json 생성 로직 구현
4. 테스트 작성 및 검증

**예상 시간**: 1.5시간
**LOC**: ~120 LOC (클래스), ~120 LOC (테스트)

### Phase 3: 기존 클래스 리팩토링 (PackageManagerInstaller)

1. CommandBuilder 의존성 주입
2. package.json 관련 메서드 제거
3. 불필요한 private 메서드 제거
4. 기존 테스트 업데이트

**예상 시간**: 1시간
**LOC**: ~150 LOC (클래스), ~107 LOC (테스트)

### Phase 4: 통합 테스트 및 검증

1. 전체 테스트 스위트 실행
2. 커버리지 검증
3. 코드 품질 체크
4. @TAG 체인 검증

**예상 시간**: 0.5시간

## 성공 지표

### 정량적 지표

- installer.ts: 399 LOC → 150 LOC (-62%)
- 평균 메서드 크기: 감소
- 테스트 파일 수: 1개 → 3개
- 테스트 커버리지: ≥ 85%
- 순환 복잡도: ≤ 10

### 정성적 지표

- 각 클래스의 역할이 명확함
- 테스트 작성이 용이함
- 새로운 패키지 매니저 추가 시 변경 범위가 제한적임
- 코드 리뷰 시 이해가 쉬움

## @TAG 체인

```
@SPEC:REFACTOR-007 (이 문서)
    ↓
@SPEC:REFACTOR-007 (plan.md)
    ↓
@CODE:REFACTOR-007 (구현 작업)
    ↓
@TEST:REFACTOR-007 (테스트 코드)
    ↓
@CODE:REFACTOR-007 (프로덕션 코드)
```

## 참고 자료

- **TRUST 원칙**: `.moai/memory/development-guide.md`
- **기존 코드**: `src/core/package-manager/installer.ts`
- **기존 테스트**: `src/__tests__/core/package-manager/installer.test.ts`
- **타입 정의**: `src/types/package-manager.ts`

---

_이 SPEC은 `/moai:2-build` 단계에서 TDD 구현의 기준이 됩니다._
