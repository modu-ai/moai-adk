# SPEC-002 구현 계획: template-manager.ts 리팩토링

## TAG BLOCK

```text
# @SPEC:REFACTOR-002 | Chain: @SPEC:REFACTOR-002 -> @SPEC:REFACTOR-002 -> @CODE:REFACTOR-002 -> @TEST:REFACTOR-002
# Related: @CODE:TEMPLATE-002:API, @CODE:TEMPLATE-002:DATA
```

## 개요

template-manager.ts (609 LOC)를 3개 모듈로 분리하여 각각 300 LOC 이하로 줄이는 리팩토링 작업입니다. TRUST 원칙의 Unified 항목을 준수하며, TDD 방식으로 진행합니다.

---

## 마일스톤 (우선순위 기반)

### 1차 목표: 핵심 모듈 분리
- TemplateValidator 구현 (검증 및 보안)
- TemplateProcessor 구현 (템플릿 변환 및 생성)
- TemplateManager 슬림화 (오케스트레이션)

### 2차 목표: 테스트 커버리지 확보
- 단위 테스트 작성 (각 모듈 85% 이상)
- 통합 테스트 작성 (전체 프로세스)
- 크로스 플랫폼 테스트 (Windows, macOS, Linux)

### 3차 목표: 문서화 및 검증
- API 문서 업데이트
- TAG 체인 검증
- 성능 벤치마크

---

## 기술적 접근 방법

### 1. 의존성 역전 원칙 (Dependency Inversion)

```
TemplateManager (오케스트레이터)
     ↓ (의존)
TemplateProcessor (구현)
     ↓ (의존)
TemplateValidator (기초)
```

**의존성 방향**:
- TemplateManager → TemplateProcessor, TemplateValidator
- TemplateProcessor → TemplateValidator
- TemplateValidator → 독립 (외부 의존성 없음)

### 2. 단일 책임 원칙 (Single Responsibility)

| 모듈 | 책임 | 주요 기능 |
|------|------|-----------|
| **TemplateManager** | 프로세스 조율 | generateProject(), getTemplatesPath() |
| **TemplateProcessor** | 템플릿 처리 | createTemplateData(), generateProjectFiles() |
| **TemplateValidator** | 검증 및 보안 | validateConfig(), ensureDirectory(), validatePath() |

### 3. 인터페이스 분리 (Interface Segregation)

```typescript
// @CODE:TEMPLATE-002:API: 공개 API 인터페이스
export interface ITemplateManager {
  generateProject(config: ProjectConfig, targetPath: string): Promise<InitResult>;
  getTemplatesPath(): string;
}

// @CODE:TEMPLATE-002:API: 템플릿 처리 인터페이스
export interface ITemplateProcessor {
  createTemplateData(config: ProjectConfig): TemplateData;
  generateProjectFiles(projectPath: string, config: ProjectConfig, templateData: TemplateData, result: InitResult): Promise<void>;
  generateMoaiStructure(projectPath: string, templateData: TemplateData, result: InitResult): Promise<void>;
  generateClaudeStructure(projectPath: string, templateData: TemplateData, result: InitResult): Promise<void>;
}

// @CODE:TEMPLATE-002:API: 검증 인터페이스
export interface ITemplateValidator {
  validateConfig(config: ProjectConfig): ValidationResult;
  ensureDirectory(dirPath: string): Promise<void>;
  validatePath(targetPath: string, basePath: string): boolean;
  maskSensitiveData(data: TemplateData): TemplateData;
  hasFeature(config: ProjectConfig, featureName: string): boolean;
}
```

---

## 아키텍처 설계 방향

### 현재 구조 (Before)

```
src/core/project/
└── template-manager.ts (609 LOC)
    ├── generateProject()           - 50 LOC
    ├── createTemplateData()        - 15 LOC
    ├── createBaseStructure()       - 10 LOC
    ├── createProjectFiles()        - 20 LOC
    ├── createPythonFiles()         - 25 LOC
    ├── createNodeJSFiles()         - 35 LOC
    ├── createFrontendFiles()       - 15 LOC
    ├── createMixedProjectFiles()   - 30 LOC
    ├── createMoaiStructure()       - 35 LOC
    ├── createClaudeStructure()     - 60 LOC
    ├── hasFeature()                - 5 LOC
    ├── ensureDirectory()           - 10 LOC
    └── [템플릿 생성 메서드들]      - 300 LOC
```

### 목표 구조 (After)

```
src/core/project/
├── template-manager.ts (~150 LOC)
│   ├── constructor()                - 10 LOC
│   ├── generateProject()            - 80 LOC
│   └── getTemplatesPath()           - 5 LOC
│
├── template-processor.ts (~250 LOC)
│   ├── createTemplateData()         - 15 LOC
│   ├── generateProjectFiles()       - 30 LOC
│   ├── generateMoaiStructure()      - 40 LOC
│   ├── generateClaudeStructure()    - 65 LOC
│   ├── [프로젝트 타입별 생성]       - 100 LOC
│   └── [템플릿 생성 메서드들]       - 150 LOC
│
└── template-validator.ts (~200 LOC)
    ├── validateConfig()             - 50 LOC
    ├── ensureDirectory()            - 15 LOC
    ├── validatePath()               - 40 LOC
    ├── maskSensitiveData()          - 60 LOC
    └── hasFeature()                 - 10 LOC
```

---

## 리스크 및 대응 방안

### 리스크 1: API 호환성 깨짐 (High)

**원인**: 기존 코드에서 TemplateManager를 직접 사용하는 부분

**대응 방안**:
1. 공개 API(`generateProject()`, `getTemplatesPath()`)는 변경하지 않음
2. 내부 구현만 private 메서드로 이동
3. 기존 테스트가 모두 통과하도록 보장

**검증 방법**:
```bash
# 기존 사용처 검색
rg "new TemplateManager" -g"*.ts" -n
rg "\.generateProject\(" -g"*.ts" -n

# 테스트 실행
pnpm test src/core/project/template-manager.test.ts
```

### 리스크 2: 테스트 깨짐 (High)

**원인**: 기존 테스트가 private 메서드에 의존하는 경우

**대응 방안**:
1. 테스트를 공개 API 기반으로 리팩토링
2. 필요 시 각 모듈별 독립적인 테스트 추가
3. 모킹 전략 수립 (TemplateProcessor, TemplateValidator)

**검증 방법**:
```bash
# 테스트 커버리지 확인
pnpm test:coverage

# 최소 85% 유지 확인
```

### 리스크 3: 플랫폼별 동작 차이 (Medium)

**원인**: Windows 경로 처리 (`\` vs `/`)

**대응 방안**:
1. Node.js `path` 모듈 사용 (path.join, path.resolve)
2. 경로 정규화 함수 추가 (TemplateValidator)
3. 크로스 플랫폼 테스트 자동화

**검증 방법**:
```bash
# CI/CD에서 각 플랫폼별 테스트
# Windows: GitHub Actions windows-latest
# macOS: GitHub Actions macos-latest
# Linux: GitHub Actions ubuntu-latest
```

### 리스크 4: Mustache 템플릿 의존성 (Medium)

**원인**: 템플릿 처리 로직이 복잡하고 중요

**대응 방안**:
1. TemplateProcessor에 격리
2. 템플릿 렌더링 단위 테스트 강화
3. 샘플 템플릿으로 통합 테스트

**검증 방법**:
```bash
# 템플릿 렌더링 테스트
pnpm test src/core/project/template-processor.test.ts
```

---

## 구현 순서 (의존성 역순)

### Step 1: TemplateValidator 구현 (기초)
- **의존성**: 없음 (독립 모듈)
- **우선순위**: 최우선 (다른 모듈이 의존)

**작업 항목**:
1. `src/core/project/template-validator.ts` 생성
2. `validateConfig()` 구현 (프로젝트 이름 형식 검증)
3. `ensureDirectory()` 구현 (안전한 디렉토리 생성)
4. `validatePath()` 구현 (Path traversal 방지)
5. `maskSensitiveData()` 구현 (민감 정보 마스킹)
6. `hasFeature()` 구현 (기능 활성화 확인)
7. 단위 테스트 작성 (`template-validator.test.ts`)

**완료 조건**:
- [ ] TemplateValidator 클래스 구현 완료
- [ ] 모든 메서드 단위 테스트 통과
- [ ] 커버리지 85% 이상

---

### Step 2: TemplateProcessor 구현 (중간)
- **의존성**: TemplateValidator
- **우선순위**: 2순위

**작업 항목**:
1. `src/core/project/template-processor.ts` 생성
2. `createTemplateData()` 구현 (ProjectConfig → TemplateData 변환)
3. `generateProjectFiles()` 구현 (프로젝트 타입별 분기)
4. `generateMoaiStructure()` 구현 (.moai 디렉토리 생성)
5. `generateClaudeStructure()` 구현 (.claude 디렉토리 생성)
6. [프로젝트 타입별 메서드] 구현:
   - `createPythonFiles()`
   - `createNodeJSFiles()`
   - `createFrontendFiles()`
   - `createMixedProjectFiles()`
7. [템플릿 생성 메서드] 구현:
   - `generatePyprojectToml()`
   - `generatePackageJson()`
   - `generateTsConfig()`
   - `generateJestConfig()`
   - `generatePytestConfig()`
8. 단위 테스트 작성 (`template-processor.test.ts`)

**완료 조건**:
- [ ] TemplateProcessor 클래스 구현 완료
- [ ] 모든 프로젝트 타입별 파일 생성 테스트 통과
- [ ] 커버리지 85% 이상

---

### Step 3: TemplateManager 리팩토링 (오케스트레이터)
- **의존성**: TemplateProcessor, TemplateValidator
- **우선순위**: 최종 단계

**작업 항목**:
1. `src/core/project/template-manager.ts` 리팩토링
2. TemplateProcessor, TemplateValidator 인스턴스 생성 로직 추가
3. `generateProject()` 메서드 슬림화:
   - 검증 로직 → `validator.validateConfig()`
   - 템플릿 데이터 생성 → `processor.createTemplateData()`
   - 파일 생성 로직 → `processor.generateProjectFiles()`
   - .moai 생성 → `processor.generateMoaiStructure()`
   - .claude 생성 → `processor.generateClaudeStructure()`
4. 기존 private 메서드 제거 (TemplateProcessor, TemplateValidator로 이동)
5. 기존 테스트 실행 확인 (`template-manager.test.ts`)
6. 통합 테스트 작성 (전체 프로세스)

**완료 조건**:
- [ ] TemplateManager 150 LOC 이하
- [ ] 기존 공개 API 100% 호환
- [ ] 모든 기존 테스트 통과
- [ ] 통합 테스트 통과

---

## 테스트 전략

### 1. 단위 테스트 (Unit Test)

**TemplateValidator 테스트**:
```typescript
// @TEST:REFACTOR-002: TemplateValidator 단위 테스트
describe('TemplateValidator', () => {
  test('should validate valid project name', () => {
    const validator = new TemplateValidator();
    const result = validator.validateConfig({ name: 'my-project', ... });
    expect(result.valid).toBe(true);
  });

  test('should reject invalid project name with special chars', () => {
    const validator = new TemplateValidator();
    const result = validator.validateConfig({ name: 'my@project!', ... });
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Invalid project name format');
  });

  test('should mask sensitive data in logs', () => {
    const validator = new TemplateValidator();
    const data = { password: 'secret123', token: 'abc123' };
    const masked = validator.maskSensitiveData(data);
    expect(masked.password).toBe('******');
    expect(masked.token).toBe('******');
  });
});
```

**TemplateProcessor 테스트**:
```typescript
// @TEST:REFACTOR-002: TemplateProcessor 단위 테스트
describe('TemplateProcessor', () => {
  test('should create template data from config', () => {
    const processor = new TemplateProcessor();
    const config: ProjectConfig = { name: 'test-project', type: ProjectType.NODEJS, ... };
    const templateData = processor.createTemplateData(config);
    expect(templateData.projectName).toBe('test-project');
    expect(templateData.projectType).toBe(ProjectType.NODEJS);
  });

  test('should generate Python project files', async () => {
    const processor = new TemplateProcessor();
    const result: InitResult = { createdFiles: [], ... };
    await processor.generateProjectFiles('/tmp/test', config, templateData, result);
    expect(result.createdFiles).toContain('pyproject.toml');
  });
});
```

**TemplateManager 테스트**:
```typescript
// @TEST:REFACTOR-002: TemplateManager 통합 테스트
describe('TemplateManager', () => {
  test('should generate complete project structure', async () => {
    const manager = new TemplateManager();
    const config: ProjectConfig = { name: 'test-project', ... };
    const result = await manager.generateProject(config, '/tmp');
    expect(result.success).toBe(true);
    expect(result.createdFiles.length).toBeGreaterThan(0);
  });
});
```

### 2. 통합 테스트 (Integration Test)

```typescript
// @TEST:REFACTOR-002: 전체 프로세스 통합 테스트
describe('Template Manager Integration', () => {
  test('should generate Python project with all files', async () => {
    const manager = new TemplateManager();
    const config: ProjectConfig = {
      name: 'python-test',
      type: ProjectType.PYTHON,
      features: [{ name: 'pytest', enabled: true }]
    };
    const result = await manager.generateProject(config, tmpDir);

    expect(result.success).toBe(true);
    expect(result.createdFiles).toContain('pyproject.toml');
    expect(result.createdFiles).toContain('pytest.ini');
    expect(result.createdFiles).toContain('.moai/config.json');
  });

  test('should handle missing directory gracefully', async () => {
    const manager = new TemplateManager();
    const config: ProjectConfig = { ... };
    const result = await manager.generateProject(config, '/nonexistent/path');

    expect(result.success).toBe(false);
    expect(result.errors.length).toBeGreaterThan(0);
  });
});
```

### 3. 크로스 플랫폼 테스트

```typescript
// @TEST:REFACTOR-002: 플랫폼별 경로 처리 테스트
describe('Cross-platform Path Handling', () => {
  test('should handle Windows paths correctly', () => {
    const validator = new TemplateValidator();
    const windowsPath = 'C:\\Users\\test\\project';
    expect(() => validator.validatePath(windowsPath, 'C:\\Users\\test')).not.toThrow();
  });

  test('should handle Unix paths correctly', () => {
    const validator = new TemplateValidator();
    const unixPath = '/home/test/project';
    expect(() => validator.validatePath(unixPath, '/home/test')).not.toThrow();
  });
});
```

---

## 마이그레이션 체크리스트

### Phase 1: 준비 (사전 작업)
- [ ] 기존 template-manager.ts 백업
- [ ] 현재 테스트 커버리지 측정
- [ ] 의존성 분석 완료 (사용처 파악)
- [ ] API 계약 문서화

### Phase 2: 구현 (TDD)
- [ ] TemplateValidator 구현 및 테스트
- [ ] TemplateProcessor 구현 및 테스트
- [ ] TemplateManager 리팩토링 및 테스트

### Phase 3: 검증 (통합)
- [ ] 모든 단위 테스트 통과
- [ ] 통합 테스트 통과
- [ ] 크로스 플랫폼 테스트 통과
- [ ] 테스트 커버리지 85% 이상 확인

### Phase 4: 마무리 (문서화)
- [ ] API 문서 업데이트
- [ ] TAG 체인 검증 (`/alfred:3-sync`)
- [ ] CHANGELOG 업데이트
- [ ] 성능 벤치마크 실행

---

## 성능 고려사항

### 1. 파일 I/O 최적화

**현재**:
```typescript
// 순차 처리
await fs.writeFile(file1, content1);
await fs.writeFile(file2, content2);
await fs.writeFile(file3, content3);
```

**개선**:
```typescript
// 병렬 처리
await Promise.all([
  fs.writeFile(file1, content1),
  fs.writeFile(file2, content2),
  fs.writeFile(file3, content3)
]);
```

### 2. 템플릿 캐싱 (선택적)

```typescript
// @CODE:TEMPLATE-002:DATA: 템플릿 캐싱
private templateCache = new Map<string, string>();

private async loadTemplate(name: string): Promise<string> {
  if (this.templateCache.has(name)) {
    return this.templateCache.get(name)!;
  }
  const content = await fs.readFile(`templates/${name}`, 'utf-8');
  this.templateCache.set(name, content);
  return content;
}
```

---

## 롤백 계획

### 문제 발생 시 롤백 절차

1. **기존 코드 복원**:
   ```bash
   git checkout HEAD -- src/core/project/template-manager.ts
   ```

2. **새 파일 제거**:
   ```bash
   rm src/core/project/template-processor.ts
   rm src/core/project/template-validator.ts
   ```

3. **테스트 재실행**:
   ```bash
   pnpm test src/core/project/
   ```

### 롤백 트리거 조건
- 기존 테스트 50% 이상 실패
- 성능 저하 20% 이상
- 크로스 플랫폼 동작 불가

---

## 문서 업데이트 계획

### 1. API 문서
- `docs/api/template-manager.md` (업데이트)
- `docs/api/template-processor.md` (신규)
- `docs/api/template-validator.md` (신규)

### 2. 아키텍처 문서
- `.moai/project/structure.md` (업데이트)
- `docs/architecture/template-system.md` (신규)

### 3. 개발 가이드
- `.moai/memory/development-guide.md` (리팩토링 사례 추가)

---

## 참조 문서

- **SPEC-002 명세**: `spec.md` (이 계획서의 부모 문서)
- **수락 기준**: `acceptance.md` (테스트 시나리오)
- **개발 가이드**: `.moai/memory/development-guide.md`
- **TRUST 원칙**: `CLAUDE.md`

---

**작성일**: 2025-10-01
**버전**: 1.0
**상태**: Draft
**담당**: MoAI-ADK Team
