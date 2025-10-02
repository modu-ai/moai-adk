# SPEC-002: template-manager.ts 리팩토링

## TAG BLOCK

```text
# @CODE:REFACTOR-002 | Chain: @SPEC:REFACTOR-002 -> @SPEC:REFACTOR-002 -> @CODE:REFACTOR-002 -> @TEST:REFACTOR-002
# Related: @CODE:TEMPLATE-002:API, @CODE:TEMPLATE-002:DATA
```

## 현재 상황 분석

### 문제점
- **파일**: `src/core/project/template-manager.ts`
- **현재 LOC**: 609 라인
- **권장 기준**: 300 LOC 이하
- **초과율**: 203% (309 라인 초과)
- **역할**: 프로젝트 템플릿 생성 및 변환 관리

### 복잡도 분석
1. **템플릿 생성 로직**: Python, Node.js, TypeScript, Frontend, Mixed 프로젝트 타입별 파일 생성 (250+ LOC)
2. **파일 생성 메서드**: 각 프로젝트 타입별 설정 파일 생성 (150+ LOC)
3. **MoAI 구조 생성**: .moai 디렉토리 및 Claude 통합 구조 생성 (100+ LOC)
4. **템플릿 데이터 변환**: Mustache 템플릿 변환 및 환경 변수 처리 (100+ LOC)

## Environment (환경 및 전제사항)

### 기술 환경
- **언어**: TypeScript (Node.js 20+)
- **의존성**: Mustache 템플릿 엔진, Node.js fs/promises
- **플랫폼**: Windows, macOS, Linux (크로스 플랫폼)
- **패키지**: `@/types/project` 타입 정의

### 현재 아키텍처
```
src/core/project/
└── template-manager.ts (609 LOC) ❌ 300 LOC 초과
```

### 목표 아키텍처
```
src/core/project/
├── template-manager.ts      (~150 LOC) - 오케스트레이션
├── template-processor.ts    (~250 LOC) - 템플릿 변환 및 생성
└── template-validator.ts    (~200 LOC) - 검증 및 보안
```

## Assumptions (전제 조건)

1. **기능 보존**: 기존 템플릿 생성 로직의 정확성을 100% 유지
2. **API 호환성**: TemplateManager 공개 API는 변경하지 않음
3. **테스트 커버리지**: 기존 테스트는 수정 없이 통과해야 함
4. **크로스 플랫폼**: Windows/macOS/Linux 호환성 유지

## Requirements (EARS 방식 요구사항)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 template-manager.ts를 3개 모듈로 분리해야 한다
- 시스템은 각 모듈이 300 LOC 이하를 준수해야 한다
- 시스템은 기존 공개 API(generateProject, getTemplatesPath)를 유지해야 한다
- 시스템은 Mustache 템플릿 엔진 의존성을 유지해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 템플릿 생성이 실패하면, 시스템은 명확한 에러 메시지와 원인을 반환해야 한다
- WHEN 잘못된 프로젝트 이름이 입력되면, 시스템은 검증 에러를 즉시 반환해야 한다
- WHEN 파일 생성 중 권한 오류가 발생하면, 시스템은 생성된 파일 목록과 실패 원인을 보고해야 한다

### State-driven Requirements (상태 기반)
- WHILE 개발 모드일 때, 시스템은 템플릿 변환 디버그 정보를 로깅해야 한다
- WHILE 프로덕션 모드일 때, 시스템은 최소한의 에러 정보만 노출해야 한다

### Optional Features (선택적 기능)
- WHERE 템플릿 캐싱이 활성화되면, 시스템은 반복적인 템플릿 로드를 최적화할 수 있다
- WHERE 커스텀 템플릿 경로가 제공되면, 시스템은 기본 템플릿 대신 사용할 수 있다

### Constraints (제약사항)
- IF 프로젝트 이름에 특수문자가 포함되면, 시스템은 생성을 거부해야 한다
- IF 환경 변수에 비밀 정보가 포함되면, 시스템은 로그에 마스킹해야 한다
- 각 모듈은 300 LOC를 초과하지 않아야 한다
- 모든 파일 작업은 async/await로 구현해야 한다
- 테스트 커버리지는 85% 이상을 유지해야 한다

## Specifications (상세 명세)

### 1. template-manager.ts (오케스트레이터, ~150 LOC)

**책임**: 템플릿 생성 프로세스 전체 조율

```typescript
// @CODE:REFACTOR-002 | Chain: @SPEC:REFACTOR-002 -> @SPEC:REFACTOR-002 -> @CODE:REFACTOR-002 -> @TEST:REFACTOR-002
// Related: @CODE:TEMPLATE-002:API

export class TemplateManager {
  private readonly templatesPath: string;
  private readonly processor: TemplateProcessor;
  private readonly validator: TemplateValidator;

  constructor(templatesPath?: string) {
    // 템플릿 경로 초기화
    // TemplateProcessor, TemplateValidator 인스턴스 생성
  }

  public async generateProject(
    config: ProjectConfig,
    targetPath: string
  ): Promise<InitResult> {
    // 1. 검증 (validator.validateConfig)
    // 2. 디렉토리 생성 (validator.ensureDirectory)
    // 3. 템플릿 데이터 생성 (processor.createTemplateData)
    // 4. 프로젝트 파일 생성 (processor.generateProjectFiles)
    // 5. .moai 구조 생성 (processor.generateMoaiStructure)
    // 6. .claude 구조 생성 (선택적)
    // 7. 결과 반환
  }

  public getTemplatesPath(): string {
    return this.templatesPath;
  }
}
```

**주요 메서드**:
- `generateProject()`: 전체 프로세스 오케스트레이션
- `getTemplatesPath()`: 템플릿 경로 반환

**LOC 목표**: ~150 라인

---

### 2. template-processor.ts (템플릿 변환 및 생성, ~250 LOC)

**책임**: 템플릿 렌더링 및 프로젝트 파일 생성

```typescript
// @CODE:REFACTOR-002 | Chain: @SPEC:REFACTOR-002 -> @SPEC:REFACTOR-002 -> @CODE:REFACTOR-002 -> @TEST:REFACTOR-002
// Related: @CODE:TEMPLATE-002:API, @CODE:TEMPLATE-002:DATA

export class TemplateProcessor {
  /**
   * ProjectConfig를 TemplateData로 변환
   * @CODE:TEMPLATE-002:API
   */
  public createTemplateData(config: ProjectConfig): TemplateData {
    // 프로젝트 설정을 템플릿 데이터로 변환
  }

  /**
   * 프로젝트 타입별 파일 생성
   * @CODE:TEMPLATE-002:API
   */
  public async generateProjectFiles(
    projectPath: string,
    config: ProjectConfig,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // Python, Node.js, TypeScript, Frontend, Mixed 프로젝트 타입별 분기
  }

  /**
   * .moai 디렉토리 구조 생성
   * @CODE:TEMPLATE-002:DATA
   */
  public async generateMoaiStructure(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // .moai/config.json, project/*.md, reports/*.md 생성
  }

  /**
   * .claude 디렉토리 구조 생성
   * @CODE:TEMPLATE-002:DATA
   */
  public async generateClaudeStructure(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // .claude/agents, commands, hooks 생성
  }

  // 내부 메서드 (private)
  private async createPythonFiles(...): Promise<void>
  private async createNodeJSFiles(...): Promise<void>
  private async createFrontendFiles(...): Promise<void>
  private async createMixedProjectFiles(...): Promise<void>

  // 템플릿 생성 메서드
  private generatePyprojectToml(data: TemplateData): string
  private generatePackageJson(data: TemplateData): any
  private generateTsConfig(): any
  private generateJestConfig(): string
  private generatePytestConfig(): string
}
```

**주요 메서드**:
- `createTemplateData()`: 설정 → 템플릿 데이터 변환
- `generateProjectFiles()`: 프로젝트 타입별 파일 생성
- `generateMoaiStructure()`: .moai 구조 생성
- `generateClaudeStructure()`: .claude 구조 생성

**LOC 목표**: ~250 라인

---

### 3. template-validator.ts (검증 및 보안, ~200 LOC)

**책임**: 입력 검증, 보안 검사, 파일 시스템 안전성 확인

```typescript
// @CODE:REFACTOR-002 | Chain: @SPEC:REFACTOR-002 -> @SPEC:REFACTOR-002 -> @CODE:REFACTOR-002 -> @TEST:REFACTOR-002
// Related: @CODE:TEMPLATE-002:API, @CODE:TEMPLATE-002:DATA

export class TemplateValidator {
  /**
   * 프로젝트 설정 검증
   * @CODE:TEMPLATE-002:API
   */
  public validateConfig(config: ProjectConfig): ValidationResult {
    // 프로젝트 이름 형식 검증 (^[a-zA-Z0-9-_]+$)
    // 필수 필드 검증
    // 프로젝트 타입 유효성 검증
  }

  /**
   * 디렉토리 존재 확인 및 생성
   * @CODE:TEMPLATE-002:DATA
   */
  public async ensureDirectory(dirPath: string): Promise<void> {
    // 디렉토리 존재 여부 확인
    // 없으면 재귀적으로 생성
  }

  /**
   * 파일 경로 보안 검증
   * @CODE:TEMPLATE-002:API
   */
  public validatePath(targetPath: string, basePath: string): boolean {
    // Path traversal 공격 방지
    // 절대 경로 검증
    // 심볼릭 링크 검증
  }

  /**
   * 환경 변수 마스킹 (보안)
   * @CODE:TEMPLATE-002:API
   */
  public maskSensitiveData(data: TemplateData): TemplateData {
    // 비밀 정보 패턴 감지 (password, token, secret 등)
    // 로그 출력 시 마스킹
  }

  /**
   * 기능 활성화 여부 확인
   * @CODE:TEMPLATE-002:API
   */
  public hasFeature(config: ProjectConfig, featureName: string): boolean {
    // config.features 배열에서 특정 기능 활성화 여부 확인
  }
}

export interface ValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
}
```

**주요 메서드**:
- `validateConfig()`: 프로젝트 설정 검증
- `ensureDirectory()`: 안전한 디렉토리 생성
- `validatePath()`: 경로 보안 검증
- `maskSensitiveData()`: 민감 정보 마스킹
- `hasFeature()`: 기능 활성화 확인

**LOC 목표**: ~200 라인

---

## Traceability (추적성)

### TAG 체인

**Primary Chain**:
```
@SPEC:REFACTOR-002 (이 문서)
  ↓
@SPEC:REFACTOR-002 (설계 문서)
  ↓
@CODE:REFACTOR-002 (구현 작업)
  ↓
@TEST:REFACTOR-002 (테스트 케이스)
```

**Implementation Tags**:
- `@CODE:REFACTOR-002`: 리팩토링 기능 전체
- `@CODE:TEMPLATE-002:API`: 템플릿 관리 API
- `@CODE:TEMPLATE-002:DATA`: 템플릿 데이터 구조

### 영향받는 파일
- `src/core/project/template-manager.ts` (분리)
- `src/core/project/template-processor.ts` (신규)
- `src/core/project/template-validator.ts` (신규)
- `tests/core/project/template-manager.test.ts` (업데이트)
- `tests/core/project/template-processor.test.ts` (신규)
- `tests/core/project/template-validator.test.ts` (신규)

## Success Criteria (성공 기준)

### 기능 요구사항
- [ ] 기존 `generateProject()` API 동작 100% 동일
- [ ] 모든 프로젝트 타입(Python, Node.js, TypeScript, Frontend, Mixed) 정상 생성
- [ ] .moai 및 .claude 구조 정확히 생성
- [ ] 크로스 플랫폼(Windows, macOS, Linux) 동작 확인

### 품질 요구사항
- [ ] template-manager.ts: 150 LOC 이하
- [ ] template-processor.ts: 250 LOC 이하
- [ ] template-validator.ts: 200 LOC 이하
- [ ] 테스트 커버리지 85% 이상 유지
- [ ] 모든 기존 테스트 통과

### 보안 요구사항
- [ ] 프로젝트 이름 형식 검증 (특수문자 차단)
- [ ] Path traversal 공격 방지
- [ ] 환경 변수 민감 정보 마스킹
- [ ] 파일 권한 오류 안전 처리

## Risk Analysis (리스크 분석)

### 고위험 (High Risk)
1. **API 호환성 깨짐**: 기존 코드에서 TemplateManager를 사용하는 부분 영향
   - **완화 방안**: 공개 API 변경 금지, 내부 구현만 분리

2. **테스트 실패**: 기존 테스트가 내부 구현에 의존하는 경우
   - **완화 방안**: 테스트 리팩토링, 모킹 전략 수립

### 중위험 (Medium Risk)
1. **플랫폼별 동작 차이**: Windows 경로 처리 이슈
   - **완화 방안**: path 모듈 사용, 크로스 플랫폼 테스트

2. **Mustache 템플릿 의존성**: 템플릿 처리 로직 복잡도
   - **완화 방안**: TemplateProcessor에 격리, 단위 테스트 강화

### 저위험 (Low Risk)
1. **성능 저하**: 모듈 분리로 인한 오버헤드
   - **완화 방안**: 벤치마크 테스트, 필요 시 최적화

## Next Steps (다음 단계)

### Phase 1: 분석 및 설계 (완료)
- [x] 현재 코드 분석
- [x] 모듈 분리 전략 수립
- [x] SPEC 문서 작성

### Phase 2: 구현 준비
- [ ] 기존 테스트 케이스 분석
- [ ] API 계약 문서화
- [ ] 모킹 전략 수립

### Phase 3: TDD 구현 (`/alfred:2-build`)
- [ ] template-validator.ts 구현 (테스트 우선)
- [ ] template-processor.ts 구현 (테스트 우선)
- [ ] template-manager.ts 리팩토링 (테스트 우선)

### Phase 4: 검증 및 문서화 (`/alfred:3-sync`)
- [ ] 크로스 플랫폼 통합 테스트
- [ ] 성능 벤치마크
- [ ] 문서 동기화

## References (참조 문서)

- **개발 가이드**: `.moai/memory/development-guide.md`
- **프로젝트 구조**: `.moai/project/structure.md`
- **TRUST 원칙**: CLAUDE.md
- **기존 코드**: `src/core/project/template-manager.ts`
- **타입 정의**: `src/types/project.ts`

---

**작성일**: 2025-10-01
**버전**: 1.0
**상태**: Draft
**담당**: MoAI-ADK Team
