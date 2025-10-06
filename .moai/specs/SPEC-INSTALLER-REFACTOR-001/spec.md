---
id: INSTALLER-REFACTOR-001
version: 0.1.0
status: draft
created: 2025-10-06
updated: 2025-10-06
author: @Goos
labels:
  - refactoring
  - quality
  - installer
  - code-structure
priority: high
---

# @SPEC:INSTALLER-REFACTOR-001: LOC 제한 준수 리팩토링

## HISTORY

### v0.1.0 (2025-10-06)
- **INITIAL**: LOC 제한 준수 리팩토링 명세 작성
- **AUTHOR**: @Goos
- **SCOPE**:
  - phase-executor.ts (358 LOC) → <300 LOC 분리
  - template-processor.ts (371 LOC) → <300 LOC 분리
  - 단일 책임 원칙(SRP) 적용
  - 테스트 용이성 향상
- **BACKGROUND**:
  - TRUST 원칙 중 Readable (가독성) 준수 필요
  - 현재 2개 파일이 300 LOC 제한 위반
  - 복잡도 증가로 유지보수 어려움

---

## Environment (환경)

### 시스템 환경
- **Language**: TypeScript 5.x
- **Runtime**: Node.js 18+
- **Testing**: Vitest
- **Linting**: ESLint + Biome

### 위반 파일
1. **phase-executor.ts**: 358 LOC (+58 초과)
2. **template-processor.ts**: 371 LOC (+71 초과)

### 코드 규칙 (TRUST 원칙)
- 파일 ≤ 300 LOC
- 함수 ≤ 50 LOC
- 매개변수 ≤ 5개
- 복잡도 ≤ 10

---

## Assumptions (가정)

1. **단일 책임 원칙**: 각 클래스는 하나의 책임만 가져야 한다
2. **테스트 독립성**: 분리된 클래스는 각각 독립적으로 테스트 가능해야 한다
3. **의존성 주입**: 모든 의존성은 생성자 주입 방식을 사용한다
4. **후방 호환성**: 기존 public API는 유지되어야 한다
5. **성능 영향 최소화**: 리팩토링으로 인한 성능 저하 < 5%

---

## Requirements (요구사항)

### Ubiquitous Requirements (필수 기능)

1. **LOC 제한 준수**
   - 모든 파일은 300 LOC를 초과하지 않아야 한다
   - 모든 함수는 50 LOC를 초과하지 않아야 한다
   - 모든 함수는 5개 이하의 매개변수를 가져야 한다

2. **phase-executor.ts 분리**
   - PhaseExecutor 클래스를 다음으로 분리해야 한다:
     - `PhaseExecutor` (조율자, <200 LOC)
     - `BackupManager` (백업 관리, <100 LOC)
     - `DirectoryBuilder` (디렉토리 생성, <100 LOC)
     - `GitInitializer` (Git 초기화, <100 LOC)

3. **template-processor.ts 분리**
   - TemplateProcessor 클래스를 다음으로 분리해야 한다:
     - `TemplateProcessor` (조율자, <200 LOC)
     - `TemplatePathResolver` (경로 해석, <150 LOC)
     - `TemplateRenderer` (렌더링, <100 LOC)

### Event-driven Requirements (이벤트 기반)

1. **파일 분리 시**
   - WHEN 클래스를 분리하면, 각 클래스는 독립적인 파일로 생성되어야 한다
   - WHEN 새 클래스가 생성되면, 해당 클래스의 테스트 파일도 생성되어야 한다

2. **의존성 주입 시**
   - WHEN 클래스 간 의존성이 필요하면, 생성자 주입을 사용해야 한다
   - WHEN 의존성이 선택적이면, 기본값을 제공해야 한다

3. **리팩토링 완료 시**
   - WHEN 모든 파일이 LOC 제한을 준수하면, 전체 테스트가 통과해야 한다
   - WHEN 리팩토링이 완료되면, 기존 API 호환성이 유지되어야 한다

### State-driven Requirements (상태 기반)

1. **분리된 클래스 상태**
   - WHILE 클래스가 독립적으로 존재할 때, 다른 클래스 없이도 테스트 가능해야 한다
   - WHILE 의존성 주입이 활성화된 상태일 때, mock 객체로 대체 가능해야 한다

### Optional Features (선택 기능)

1. **성능 최적화**
   - WHERE 리팩토링으로 성능이 개선되면, 벤치마크 결과를 문서화할 수 있다
   - WHERE 클래스 분리로 병렬 처리가 가능하면, 성능 향상을 측정할 수 있다

### Constraints (제약사항)

1. **후방 호환성**
   - 기존 public API는 변경하지 않아야 한다
   - 외부에서 사용 중인 메서드 시그니처는 유지되어야 한다

2. **성능 저하 제한**
   - IF 리팩토링으로 성능이 저하되면, 저하율은 5% 미만이어야 한다
   - IF 새로운 클래스 생성 오버헤드가 발생하면, 측정 가능한 범위 내여야 한다

3. **테스트 커버리지 유지**
   - IF 기존 커버리지가 X%였다면, 리팩토링 후에도 X% 이상이어야 한다
   - 새로 생성된 클래스는 85% 이상의 커버리지를 가져야 한다

---

## Technical Design (기술 설계)

### phase-executor.ts 분리 설계

#### Before (358 LOC)
```typescript
// phase-executor.ts (358 LOC - 위반)
export class PhaseExecutor {
  // 백업 관련 (60 LOC)
  private async createBackup(config) { ... }

  // 디렉토리 관련 (30 LOC)
  private async createProjectDirectories(config) { ... }

  // Git 초기화 (50 LOC)
  private async initializeGitRepository(config) { ... }

  // 5개 Phase 실행 메서드 (200+ LOC)
  async executePreparationPhase() { ... }
  async executeDirectoryPhase() { ... }
  async executeResourcePhase() { ... }
  async executeConfigurationPhase() { ... }
  async executeValidationPhase() { ... }
}
```

#### After (분리)
```typescript
// @CODE:INSTALLER-REFACTOR-001:BACKUP
// backup-manager.ts (<100 LOC)
export class BackupManager {
  async createBackup(config: InstallationConfig): Promise<void> {
    // 백업 로직만 집중
  }

  async validateBackupPath(path: string): Promise<boolean> {
    // 백업 경로 검증
  }
}

// @CODE:INSTALLER-REFACTOR-001:DIR
// directory-builder.ts (<100 LOC)
export class DirectoryBuilder {
  async createProjectDirectories(config: InstallationConfig): Promise<void> {
    // 디렉토리 생성 로직만 집중
  }

  private async ensureParentDirectory(path: string): Promise<void> {
    // 부모 디렉토리 확인
  }
}

// @CODE:INSTALLER-REFACTOR-001:GIT
// git-initializer.ts (<100 LOC)
export class GitInitializer {
  async initializeRepository(config: InstallationConfig): Promise<void> {
    // Git 초기화 로직만 집중
  }

  private async createInitialCommit(path: string): Promise<void> {
    // 초기 커밋 생성
  }
}

// @CODE:INSTALLER-REFACTOR-001:EXECUTOR
// phase-executor.ts (<200 LOC - 준수)
export class PhaseExecutor {
  constructor(
    private readonly contextManager: ContextManager,
    private readonly backupManager: BackupManager,      // DI
    private readonly directoryBuilder: DirectoryBuilder, // DI
    private readonly gitInitializer: GitInitializer      // DI
  ) {}

  async executePreparationPhase(...) {
    if (context.config.backupEnabled) {
      await this.backupManager.createBackup(context.config);
    }
    // ...
  }

  async executeDirectoryPhase(...) {
    await this.directoryBuilder.createProjectDirectories(context.config);
    // ...
  }

  // 조율 로직만 집중
}
```

### template-processor.ts 분리 설계

#### Before (371 LOC)
```typescript
// template-processor.ts (371 LOC - 위반)
export class TemplateProcessor {
  // 경로 해석 관련 (180 LOC)
  getTemplatesPath() { ... }
  private tryPackageRelativeTemplates() { ... }
  private tryDevelopmentTemplates() { ... }
  private tryUserNodeModulesTemplates() { ... }
  private tryGlobalInstallTemplates() { ... }

  // 렌더링 관련 (191 LOC)
  createTemplateVariables() { ... }
  copyTemplateDirectory() { ... }
  copyTemplateFile() { ... }
  copyDirectory() { ... }
}
```

#### After (분리)
```typescript
// @CODE:INSTALLER-REFACTOR-001:PATH
// template-path-resolver.ts (<150 LOC)
export class TemplatePathResolver {
  getTemplatesPath(currentFileUrl: string): string {
    // 경로 해석 로직만 집중
  }

  private tryPackageRelativeTemplates(dir: string): string | null {
    // 패키지 상대 경로
  }

  private tryDevelopmentTemplates(dir: string): string | null {
    // 개발 환경 경로
  }

  // ... 기타 경로 해석 메서드
}

// @CODE:INSTALLER-REFACTOR-001:RENDER
// template-renderer.ts (<100 LOC)
export class TemplateRenderer {
  createTemplateVariables(config: InstallationConfig): TemplateData {
    // 변수 생성 로직만 집중
  }

  async renderFile(
    srcPath: string,
    dstPath: string,
    variables: Record<string, any>
  ): Promise<void> {
    // 파일 렌더링 로직만 집중
  }
}

// @CODE:INSTALLER-REFACTOR-001:PROCESSOR
// template-processor.ts (<200 LOC - 준수)
export class TemplateProcessor {
  constructor(
    private readonly pathResolver: TemplatePathResolver,
    private readonly renderer: TemplateRenderer
  ) {}

  getTemplatesPath(): string {
    return this.pathResolver.getTemplatesPath(import.meta.url);
  }

  async copyTemplateFile(src, dst, vars) {
    return this.renderer.renderFile(src, dst, vars);
  }

  // 조율 로직만 집중
}
```

---

## Traceability (@TAG)

- **SPEC**: @SPEC:INSTALLER-REFACTOR-001
- **TEST**:
  - `__tests__/backup-manager.test.ts`
  - `__tests__/directory-builder.test.ts`
  - `__tests__/git-initializer.test.ts`
  - `__tests__/template-path-resolver.test.ts`
  - `__tests__/template-renderer.test.ts`
  - `__tests__/phase-executor.test.ts` (수정)
  - `__tests__/template-processor.test.ts` (수정)
- **CODE**:
  - `src/core/installer/backup-manager.ts` (신규)
  - `src/core/installer/directory-builder.ts` (신규)
  - `src/core/installer/git-initializer.ts` (신규)
  - `src/core/installer/template-path-resolver.ts` (신규)
  - `src/core/installer/template-renderer.ts` (신규)
  - `src/core/installer/phase-executor.ts` (수정)
  - `src/core/installer/template-processor.ts` (수정)
- **DOC**: `.moai/specs/SPEC-INSTALLER-REFACTOR-001/`

---

## Success Criteria (성공 기준)

### LOC 준수
- [ ] phase-executor.ts ≤ 300 LOC
- [ ] template-processor.ts ≤ 300 LOC
- [ ] 모든 신규 파일 ≤ 300 LOC
- [ ] 모든 함수 ≤ 50 LOC

### 코드 품질
- [ ] 복잡도 ≤ 10 (모든 함수)
- [ ] 매개변수 ≤ 5개 (모든 함수)
- [ ] 테스트 커버리지 ≥ 85%
- [ ] ESLint 경고 없음

### 기능 완성
- [ ] 모든 기존 테스트 통과
- [ ] 신규 클래스별 테스트 작성
- [ ] public API 호환성 유지
- [ ] 성능 저하 < 5%

---

## Dependencies (의존성)

### SPEC 의존성
- **독립적**: 다른 SPEC과 의존성 없음
- **병렬 가능**: SPEC-INSTALLER-SEC-001과 병렬 진행 가능

### 기술 의존성
- TypeScript 컴파일러
- Vitest 테스트 프레임워크
- ESLint + Biome

---

## Risk Analysis (리스크 분석)

### 높은 리스크
1. **API 깨짐**: 외부에서 사용 중인 API 변경 가능성
   - **완화**: public API 변경 금지, 내부 구현만 변경

2. **회귀 버그**: 기존 기능 동작 변경 가능성
   - **완화**: 모든 기존 테스트 유지 및 통과 확인

### 중간 리스크
1. **성능 저하**: 클래스 분리로 인한 오버헤드
   - **완화**: 성능 벤치마크 수행, 5% 이내 유지

2. **복잡도 증가**: 파일 수 증가로 인한 복잡도
   - **완화**: 명확한 책임 분리, 문서화 강화

---

## Implementation Notes (구현 참고사항)

### 단계별 구현
1. **Phase 1**: 신규 클래스 파일 생성 (빈 껍데기)
2. **Phase 2**: 로직 이동 (한 번에 하나씩)
3. **Phase 3**: 테스트 작성 (각 클래스별)
4. **Phase 4**: 기존 테스트 수정 (DI 적용)
5. **Phase 5**: 문서화 및 검토

### 리팩토링 원칙
- **작은 단계**: 한 번에 하나의 클래스만 분리
- **테스트 유지**: 각 단계마다 테스트 통과 확인
- **커밋 전략**: 클래스별 독립적인 커밋
- **롤백 가능**: 각 단계는 독립적으로 롤백 가능

### 코드 리뷰 포인트
- 단일 책임 원칙 준수
- 의존성 주입 일관성
- 테스트 커버리지 유지
- 성능 벤치마크 확인
