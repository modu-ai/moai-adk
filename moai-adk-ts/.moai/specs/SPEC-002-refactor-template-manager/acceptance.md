# SPEC-002 수락 기준: template-manager.ts 리팩토링

## TAG BLOCK

```text
# @TEST:REFACTOR-002 | Chain: @SPEC:REFACTOR-002 -> @SPEC:REFACTOR-002 -> @CODE:REFACTOR-002 -> @TEST:REFACTOR-002
# Related: @CODE:TEMPLATE-002:API, @CODE:TEMPLATE-002:DATA
```

## 개요

이 문서는 SPEC-002 리팩토링 작업의 수락 기준과 테스트 시나리오를 정의합니다. 모든 시나리오는 Given-When-Then 형식으로 작성되었습니다.

---

## 1. 기능 수락 기준

### AC-1: 프로젝트 생성 API 호환성

**Given-When-Then 시나리오**:

```gherkin
Scenario: Python 프로젝트 생성
  Given 유효한 Python 프로젝트 설정이 주어졌을 때
  And TemplateManager 인스턴스가 생성되었을 때
  When generateProject() 메서드를 호출하면
  Then 프로젝트 디렉토리가 생성되어야 한다
  And pyproject.toml 파일이 생성되어야 한다
  And src/__init__.py 파일이 생성되어야 한다
  And .moai/config.json 파일이 생성되어야 한다
  And InitResult.success가 true여야 한다

Scenario: Node.js 프로젝트 생성
  Given 유효한 Node.js 프로젝트 설정이 주어졌을 때
  And TemplateManager 인스턴스가 생성되었을 때
  When generateProject() 메서드를 호출하면
  Then 프로젝트 디렉토리가 생성되어야 한다
  And package.json 파일이 생성되어야 한다
  And .moai/config.json 파일이 생성되어야 한다
  And InitResult.success가 true여야 한다

Scenario: TypeScript 프로젝트 생성
  Given 유효한 TypeScript 프로젝트 설정이 주어졌을 때
  And TypeScript 기능이 활성화되었을 때
  When generateProject() 메서드를 호출하면
  Then package.json 파일이 생성되어야 한다
  And tsconfig.json 파일이 생성되어야 한다
  And InitResult.success가 true여야 한다

Scenario: Frontend 프로젝트 생성
  Given 유효한 Frontend 프로젝트 설정이 주어졌을 때
  When generateProject() 메서드를 호출하면
  Then package.json 파일이 생성되어야 한다
  And public/ 디렉토리가 생성되어야 한다
  And src/components/ 디렉토리가 생성되어야 한다
  And InitResult.success가 true여야 한다

Scenario: Mixed 프로젝트 생성
  Given 유효한 Mixed(Full-stack) 프로젝트 설정이 주어졌을 때
  When generateProject() 메서드를 호출하면
  Then backend/ 디렉토리가 생성되어야 한다
  And frontend/ 디렉토리가 생성되어야 한다
  And backend/pyproject.toml 파일이 생성되어야 한다
  And frontend/package.json 파일이 생성되어야 한다
  And InitResult.success가 true여야 한다
```

**자동화 테스트**:
```typescript
// @TEST:REFACTOR-002: AC-1 테스트
describe('AC-1: Project Generation API Compatibility', () => {
  const tmpDir = '/tmp/moai-test-' + Date.now();

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  test('should generate Python project successfully', async () => {
    // Given
    const manager = new TemplateManager();
    const config: ProjectConfig = {
      name: 'python-test',
      type: ProjectType.PYTHON,
      features: [{ name: 'pytest', enabled: true }]
    };

    // When
    const result = await manager.generateProject(config, tmpDir);

    // Then
    expect(result.success).toBe(true);
    expect(result.createdFiles).toContain('pyproject.toml');
    expect(result.createdFiles).toContain('src/__init__.py');
    expect(result.createdFiles).toContain('.moai/config.json');
  });

  test('should generate Node.js project successfully', async () => {
    // Given
    const manager = new TemplateManager();
    const config: ProjectConfig = {
      name: 'nodejs-test',
      type: ProjectType.NODEJS
    };

    // When
    const result = await manager.generateProject(config, tmpDir);

    // Then
    expect(result.success).toBe(true);
    expect(result.createdFiles).toContain('package.json');
    expect(result.createdFiles).toContain('.moai/config.json');
  });

  test('should generate TypeScript project successfully', async () => {
    // Given
    const manager = new TemplateManager();
    const config: ProjectConfig = {
      name: 'typescript-test',
      type: ProjectType.TYPESCRIPT,
      features: [{ name: 'typescript', enabled: true }]
    };

    // When
    const result = await manager.generateProject(config, tmpDir);

    // Then
    expect(result.success).toBe(true);
    expect(result.createdFiles).toContain('package.json');
    expect(result.createdFiles).toContain('tsconfig.json');
  });
});
```

---

### AC-2: .moai 구조 생성

**Given-When-Then 시나리오**:

```gherkin
Scenario: .moai 디렉토리 구조 생성
  Given 유효한 프로젝트 설정이 주어졌을 때
  When generateProject() 메서드를 호출하면
  Then .moai/ 디렉토리가 생성되어야 한다
  And .moai/config.json 파일이 생성되어야 한다
  And .moai/project/product.md 파일이 생성되어야 한다
  And .moai/project/structure.md 파일이 생성되어야 한다
  And .moai/project/tech.md 파일이 생성되어야 한다
  And .moai/reports/sync-report.md 파일이 생성되어야 한다

Scenario: .moai/config.json 내용 검증
  Given 유효한 프로젝트 설정이 주어졌을 때
  When .moai/config.json 파일을 읽으면
  Then project.name이 올바르게 설정되어야 한다
  And project.type이 올바르게 설정되어야 한다
  And constitution.enforce_tdd가 true여야 한다
  And constitution.test_coverage_target이 85여야 한다
```

**자동화 테스트**:
```typescript
// @TEST:REFACTOR-002: AC-2 테스트
describe('AC-2: .moai Structure Generation', () => {
  test('should create complete .moai directory structure', async () => {
    // Given
    const manager = new TemplateManager();
    const config: ProjectConfig = { name: 'test-project', type: ProjectType.NODEJS };

    // When
    const result = await manager.generateProject(config, tmpDir);

    // Then
    expect(result.createdFiles).toContain('.moai/config.json');
    expect(result.createdFiles).toContain('.moai/project/product.md');
    expect(result.createdFiles).toContain('.moai/project/structure.md');
    expect(result.createdFiles).toContain('.moai/project/tech.md');
    expect(result.createdFiles).toContain('.moai/reports/sync-report.md');
  });

  test('should generate valid .moai/config.json content', async () => {
    // Given
    const manager = new TemplateManager();
    const config: ProjectConfig = {
      name: 'test-project',
      type: ProjectType.TYPESCRIPT
    };

    // When
    await manager.generateProject(config, tmpDir);
    const configPath = path.join(tmpDir, 'test-project', '.moai', 'config.json');
    const configContent = await fs.readFile(configPath, 'utf-8');
    const moaiConfig = JSON.parse(configContent);

    // Then
    expect(moaiConfig.project.name).toBe('test-project');
    expect(moaiConfig.project.type).toBe(ProjectType.TYPESCRIPT);
    expect(moaiConfig.constitution.enforce_tdd).toBe(true);
    expect(moaiConfig.constitution.test_coverage_target).toBe(85);
  });
});
```

---

### AC-3: Claude 통합 구조 생성 (선택적)

**Given-When-Then 시나리오**:

```gherkin
Scenario: Claude 통합 기능 활성화 시 .claude 구조 생성
  Given 유효한 프로젝트 설정이 주어졌을 때
  And claude-integration 기능이 활성화되었을 때
  When generateProject() 메서드를 호출하면
  Then .claude/agents/moai/ 디렉토리가 생성되어야 한다
  And .claude/commands/moai/ 디렉토리가 생성되어야 한다
  And .claude/hooks/moai/ 디렉토리가 생성되어야 한다
  And .claude/agents/moai/spec-builder.md 파일이 생성되어야 한다
  And .claude/agents/moai/code-builder.md 파일이 생성되어야 한다
  And .claude/agents/moai/doc-syncer.md 파일이 생성되어야 한다
  And .claude/commands/moai/8-project.md 파일이 생성되어야 한다
  And .claude/hooks/moai/pre-commit.py 파일이 생성되어야 한다

Scenario: Claude 통합 기능 비활성화 시 .claude 구조 미생성
  Given 유효한 프로젝트 설정이 주어졌을 때
  And claude-integration 기능이 비활성화되었을 때
  When generateProject() 메서드를 호출하면
  Then .claude/ 디렉토리가 생성되지 않아야 한다
```

**자동화 테스트**:
```typescript
// @TEST:REFACTOR-002: AC-3 테스트
describe('AC-3: Claude Integration Structure', () => {
  test('should create .claude structure when feature is enabled', async () => {
    // Given
    const manager = new TemplateManager();
    const config: ProjectConfig = {
      name: 'test-project',
      type: ProjectType.NODEJS,
      features: [{ name: 'claude-integration', enabled: true }]
    };

    // When
    const result = await manager.generateProject(config, tmpDir);

    // Then
    expect(result.createdFiles).toContain('.claude/agents/moai/spec-builder.md');
    expect(result.createdFiles).toContain('.claude/agents/moai/code-builder.md');
    expect(result.createdFiles).toContain('.claude/agents/moai/doc-syncer.md');
    expect(result.createdFiles).toContain('.claude/commands/moai/8-project.md');
    expect(result.createdFiles).toContain('.claude/hooks/moai/pre-commit.py');
  });

  test('should not create .claude structure when feature is disabled', async () => {
    // Given
    const manager = new TemplateManager();
    const config: ProjectConfig = {
      name: 'test-project',
      type: ProjectType.NODEJS,
      features: []
    };

    // When
    const result = await manager.generateProject(config, tmpDir);

    // Then
    const claudeFiles = result.createdFiles.filter(f => f.startsWith('.claude/'));
    expect(claudeFiles.length).toBe(0);
  });
});
```

---

## 2. 품질 수락 기준

### AC-4: 코드 복잡도 제약

**Given-When-Then 시나리오**:

```gherkin
Scenario: template-manager.ts LOC 제약 준수
  Given 리팩토링이 완료되었을 때
  When template-manager.ts 파일을 검사하면
  Then 파일 크기가 150 LOC 이하여야 한다

Scenario: template-processor.ts LOC 제약 준수
  Given 리팩토링이 완료되었을 때
  When template-processor.ts 파일을 검사하면
  Then 파일 크기가 250 LOC 이하여야 한다

Scenario: template-validator.ts LOC 제약 준수
  Given 리팩토링이 완료되었을 때
  When template-validator.ts 파일을 검사하면
  Then 파일 크기가 200 LOC 이하여야 한다
```

**자동화 테스트**:
```bash
# @TEST:REFACTOR-002: AC-4 스크립트
#!/bin/bash

# LOC 측정 (주석 및 공백 제외)
check_loc() {
  local file=$1
  local max_loc=$2
  local actual_loc=$(grep -v '^\s*$' "$file" | grep -v '^\s*//' | wc -l)

  echo "File: $file"
  echo "Actual LOC: $actual_loc"
  echo "Max LOC: $max_loc"

  if [ $actual_loc -gt $max_loc ]; then
    echo "FAIL: LOC exceeds limit"
    exit 1
  else
    echo "PASS: LOC within limit"
  fi
}

check_loc "src/core/project/template-manager.ts" 150
check_loc "src/core/project/template-processor.ts" 250
check_loc "src/core/project/template-validator.ts" 200
```

---

### AC-5: 테스트 커버리지

**Given-When-Then 시나리오**:

```gherkin
Scenario: 전체 테스트 커버리지 85% 이상
  Given 모든 단위 테스트가 작성되었을 때
  When pnpm test:coverage를 실행하면
  Then 전체 커버리지가 85% 이상이어야 한다

Scenario: 각 모듈별 커버리지 85% 이상
  Given 각 모듈의 단위 테스트가 작성되었을 때
  When 모듈별 커버리지를 측정하면
  Then template-manager.ts 커버리지가 85% 이상이어야 한다
  And template-processor.ts 커버리지가 85% 이상이어야 한다
  And template-validator.ts 커버리지가 85% 이상이어야 한다
```

**자동화 테스트**:
```bash
# @TEST:REFACTOR-002: AC-5 스크립트
#!/bin/bash

# Jest/Vitest 커버리지 임계값 설정
pnpm test:coverage --coverageThreshold='{"global":{"branches":85,"functions":85,"lines":85,"statements":85}}'

if [ $? -ne 0 ]; then
  echo "FAIL: Test coverage below 85%"
  exit 1
else
  echo "PASS: Test coverage meets requirement"
fi
```

---

## 3. 보안 수락 기준

### AC-6: 입력 검증

**Given-When-Then 시나리오**:

```gherkin
Scenario: 유효한 프로젝트 이름 수락
  Given TemplateValidator 인스턴스가 생성되었을 때
  When 유효한 프로젝트 이름(영문자, 숫자, -, _)을 검증하면
  Then valid가 true여야 한다
  And errors 배열이 비어있어야 한다

Scenario: 특수문자 포함 프로젝트 이름 거부
  Given TemplateValidator 인스턴스가 생성되었을 때
  When 특수문자가 포함된 프로젝트 이름을 검증하면
  Then valid가 false여야 한다
  And errors 배열에 "Invalid project name format"이 포함되어야 한다

Scenario: 빈 프로젝트 이름 거부
  Given TemplateValidator 인스턴스가 생성되었을 때
  When 빈 문자열 프로젝트 이름을 검증하면
  Then valid가 false여야 한다
  And errors 배열에 "Project name is required"가 포함되어야 한다
```

**자동화 테스트**:
```typescript
// @TEST:REFACTOR-002: AC-6 테스트
describe('AC-6: Input Validation', () => {
  test('should accept valid project name', () => {
    // Given
    const validator = new TemplateValidator();
    const config: ProjectConfig = { name: 'my-project-123', type: ProjectType.NODEJS };

    // When
    const result = validator.validateConfig(config);

    // Then
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  test('should reject project name with special characters', () => {
    // Given
    const validator = new TemplateValidator();
    const config: ProjectConfig = { name: 'my@project!', type: ProjectType.NODEJS };

    // When
    const result = validator.validateConfig(config);

    // Then
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Invalid project name format');
  });

  test('should reject empty project name', () => {
    // Given
    const validator = new TemplateValidator();
    const config: ProjectConfig = { name: '', type: ProjectType.NODEJS };

    // When
    const result = validator.validateConfig(config);

    // Then
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Project name is required');
  });
});
```

---

### AC-7: Path Traversal 방지

**Given-When-Then 시나리오**:

```gherkin
Scenario: 안전한 경로 수락
  Given TemplateValidator 인스턴스가 생성되었을 때
  When basePath 내부의 targetPath를 검증하면
  Then true를 반환해야 한다

Scenario: Path Traversal 공격 시도 거부
  Given TemplateValidator 인스턴스가 생성되었을 때
  When "../../../etc/passwd" 같은 상위 경로 참조를 검증하면
  Then false를 반환해야 한다

Scenario: 절대 경로 외부 접근 시도 거부
  Given TemplateValidator 인스턴스가 생성되었을 때
  When basePath 외부의 절대 경로를 검증하면
  Then false를 반환해야 한다
```

**자동화 테스트**:
```typescript
// @TEST:REFACTOR-002: AC-7 테스트
describe('AC-7: Path Traversal Prevention', () => {
  test('should accept safe path within base', () => {
    // Given
    const validator = new TemplateValidator();
    const basePath = '/home/user/projects';
    const targetPath = '/home/user/projects/my-project';

    // When
    const result = validator.validatePath(targetPath, basePath);

    // Then
    expect(result).toBe(true);
  });

  test('should reject path traversal attempt', () => {
    // Given
    const validator = new TemplateValidator();
    const basePath = '/home/user/projects';
    const targetPath = '/home/user/projects/../../../etc/passwd';

    // When
    const result = validator.validatePath(targetPath, basePath);

    // Then
    expect(result).toBe(false);
  });

  test('should reject absolute path outside base', () => {
    // Given
    const validator = new TemplateValidator();
    const basePath = '/home/user/projects';
    const targetPath = '/etc/passwd';

    // When
    const result = validator.validatePath(targetPath, basePath);

    // Then
    expect(result).toBe(false);
  });
});
```

---

### AC-8: 민감 정보 마스킹

**Given-When-Then 시나리오**:

```gherkin
Scenario: 환경 변수에서 password 마스킹
  Given TemplateValidator 인스턴스가 생성되었을 때
  And 환경 변수에 "password" 키가 포함되었을 때
  When maskSensitiveData()를 호출하면
  Then "password" 값이 "******"로 마스킹되어야 한다

Scenario: 환경 변수에서 token 마스킹
  Given TemplateValidator 인스턴스가 생성되었을 때
  And 환경 변수에 "token" 또는 "api_key" 키가 포함되었을 때
  When maskSensitiveData()를 호출하면
  Then 해당 값들이 "******"로 마스킹되어야 한다

Scenario: 일반 환경 변수는 마스킹하지 않음
  Given TemplateValidator 인스턴스가 생성되었을 때
  And 환경 변수에 "projectName", "version" 같은 일반 키가 포함되었을 때
  When maskSensitiveData()를 호출하면
  Then 해당 값들은 마스킹되지 않아야 한다
```

**자동화 테스트**:
```typescript
// @TEST:REFACTOR-002: AC-8 테스트
describe('AC-8: Sensitive Data Masking', () => {
  test('should mask password field', () => {
    // Given
    const validator = new TemplateValidator();
    const data = { projectName: 'test', password: 'secret123' };

    // When
    const masked = validator.maskSensitiveData(data);

    // Then
    expect(masked.projectName).toBe('test');
    expect(masked.password).toBe('******');
  });

  test('should mask token and api_key fields', () => {
    // Given
    const validator = new TemplateValidator();
    const data = { token: 'abc123', api_key: 'xyz789', version: '1.0.0' };

    // When
    const masked = validator.maskSensitiveData(data);

    // Then
    expect(masked.token).toBe('******');
    expect(masked.api_key).toBe('******');
    expect(masked.version).toBe('1.0.0');
  });
});
```

---

## 4. 크로스 플랫폼 수락 기준

### AC-9: Windows 플랫폼 지원

**Given-When-Then 시나리오**:

```gherkin
Scenario: Windows 경로 처리
  Given Windows 환경에서 실행될 때
  When "C:\\Users\\test\\project" 경로를 처리하면
  Then 올바르게 디렉토리가 생성되어야 한다

Scenario: Windows 줄바꿈 처리
  Given Windows 환경에서 실행될 때
  When 파일을 생성하면
  Then CRLF 줄바꿈이 올바르게 처리되어야 한다
```

**자동화 테스트** (GitHub Actions):
```yaml
# @TEST:REFACTOR-002: AC-9 CI 설정
name: Windows Tests
on: [push, pull_request]
jobs:
  windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: pnpm install
      - run: pnpm test src/core/project/
```

---

### AC-10: macOS 플랫폼 지원

**Given-When-Then 시나리오**:

```gherkin
Scenario: macOS 경로 처리
  Given macOS 환경에서 실행될 때
  When "/Users/test/project" 경로를 처리하면
  Then 올바르게 디렉토리가 생성되어야 한다

Scenario: macOS 권한 처리
  Given macOS 환경에서 실행될 때
  When 실행 가능한 파일(.py, .sh)을 생성하면
  Then 올바른 권한(chmod +x)이 설정되어야 한다
```

**자동화 테스트** (GitHub Actions):
```yaml
# @TEST:REFACTOR-002: AC-10 CI 설정
name: macOS Tests
on: [push, pull_request]
jobs:
  macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: pnpm install
      - run: pnpm test src/core/project/
```

---

### AC-11: Linux 플랫폼 지원

**Given-When-Then 시나리오**:

```gherkin
Scenario: Linux 경로 처리
  Given Linux 환경에서 실행될 때
  When "/home/test/project" 경로를 처리하면
  Then 올바르게 디렉토리가 생성되어야 한다

Scenario: Linux 심볼릭 링크 처리
  Given Linux 환경에서 실행될 때
  When 심볼릭 링크가 포함된 경로를 처리하면
  Then 올바르게 해석되어야 한다
```

**자동화 테스트** (GitHub Actions):
```yaml
# @TEST:REFACTOR-002: AC-11 CI 설정
name: Linux Tests
on: [push, pull_request]
jobs:
  linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: pnpm install
      - run: pnpm test src/core/project/
```

---

## 5. 성능 수락 기준

### AC-12: 생성 시간 성능

**Given-When-Then 시나리오**:

```gherkin
Scenario: Python 프로젝트 생성 성능
  Given 유효한 Python 프로젝트 설정이 주어졌을 때
  When generateProject()를 실행하면
  Then 5초 이내에 완료되어야 한다

Scenario: Mixed 프로젝트 생성 성능
  Given 유효한 Mixed 프로젝트 설정이 주어졌을 때
  When generateProject()를 실행하면
  Then 10초 이내에 완료되어야 한다
```

**자동화 테스트**:
```typescript
// @TEST:REFACTOR-002: AC-12 테스트
describe('AC-12: Generation Performance', () => {
  test('should generate Python project within 5 seconds', async () => {
    // Given
    const manager = new TemplateManager();
    const config: ProjectConfig = { name: 'test-project', type: ProjectType.PYTHON };
    const startTime = Date.now();

    // When
    await manager.generateProject(config, tmpDir);
    const duration = Date.now() - startTime;

    // Then
    expect(duration).toBeLessThan(5000); // 5초 이내
  });

  test('should generate Mixed project within 10 seconds', async () => {
    // Given
    const manager = new TemplateManager();
    const config: ProjectConfig = {
      name: 'test-project',
      type: ProjectType.MIXED,
      features: [
        { name: 'backend-python', enabled: true },
        { name: 'frontend-typescript', enabled: true }
      ]
    };
    const startTime = Date.now();

    // When
    await manager.generateProject(config, tmpDir);
    const duration = Date.now() - startTime;

    // Then
    expect(duration).toBeLessThan(10000); // 10초 이내
  });
});
```

---

## 6. Definition of Done (완료 조건)

### Phase 1: 구현 완료
- [ ] TemplateValidator 클래스 구현 완료 (200 LOC 이하)
- [ ] TemplateProcessor 클래스 구현 완료 (250 LOC 이하)
- [ ] TemplateManager 리팩토링 완료 (150 LOC 이하)

### Phase 2: 테스트 완료
- [ ] 모든 단위 테스트 통과 (AC-1 ~ AC-12)
- [ ] 테스트 커버리지 85% 이상 (AC-5)
- [ ] 크로스 플랫폼 테스트 통과 (AC-9 ~ AC-11)

### Phase 3: 보안 검증 완료
- [ ] 입력 검증 테스트 통과 (AC-6)
- [ ] Path Traversal 방지 테스트 통과 (AC-7)
- [ ] 민감 정보 마스킹 테스트 통과 (AC-8)

### Phase 4: 문서화 완료
- [ ] API 문서 업데이트
- [ ] TAG 체인 검증 완료 (`/moai:3-sync`)
- [ ] CHANGELOG 업데이트
- [ ] 아키텍처 다이어그램 업데이트

### Phase 5: 배포 준비
- [ ] 기존 코드와 100% 호환 확인
- [ ] 성능 벤치마크 통과 (AC-12)
- [ ] 코드 리뷰 완료
- [ ] PR 승인 및 머지

---

## 7. 회귀 테스트 (Regression Tests)

### 기존 기능 보존 확인

```gherkin
Scenario: 기존 사용처 영향 없음
  Given 기존 코드베이스에서 TemplateManager를 사용하는 모든 파일을 찾았을 때
  When 리팩토링 후 해당 파일들의 테스트를 실행하면
  Then 모든 테스트가 통과해야 한다
```

**자동화 테스트**:
```bash
# @TEST:REFACTOR-002: 회귀 테스트
#!/bin/bash

# 기존 사용처 찾기
echo "Finding TemplateManager usage..."
rg "new TemplateManager" -g"*.ts" -l

# 전체 테스트 실행
echo "Running all tests..."
pnpm test

if [ $? -ne 0 ]; then
  echo "FAIL: Regression detected"
  exit 1
else
  echo "PASS: No regression"
fi
```

---

## 참조 문서

- **SPEC-002 명세**: `spec.md`
- **구현 계획**: `plan.md`
- **개발 가이드**: `.moai/memory/development-guide.md`

---

**작성일**: 2025-10-01
**버전**: 1.0
**상태**: Draft
**담당**: MoAI-ADK Team
