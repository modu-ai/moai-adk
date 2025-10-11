# SPEC-INSTALLER-SEC-001 인수 기준

## 📋 개요

- **SPEC ID**: INSTALLER-SEC-001
- **제목**: 템플릿 보안 검증 통합
- **버전**: 0.1.0
- **작성일**: 2025-10-06

---

## ✅ Acceptance Criteria (Given-When-Then)

### AC-001: 위험한 패턴 탐지 - Constructor 접근

```gherkin
Given: 템플릿 파일이 "{{constructor}}" 패턴을 포함할 때
When: copyTemplateFile() 함수를 실행
Then: InstallationError가 발생해야 한다
  And: 에러 메시지에 "Template contains dangerous patterns"가 포함되어야 한다
  And: 에러 컨텍스트에 파일 경로가 포함되어야 한다
```

**테스트 코드**:
```typescript
test('AC-001: Reject template with constructor pattern', async () => {
  // Given
  const templatePath = '/tmp/dangerous-constructor.md';
  const templateContent = 'Project: {{PROJECT_NAME}}\n{{constructor}}';
  await fs.promises.writeFile(templatePath, templateContent);

  // When & Then
  await expect(
    templateProcessor.copyTemplateFile(
      templatePath,
      '/tmp/output.md',
      { PROJECT_NAME: 'test' }
    )
  ).rejects.toThrow('Template contains dangerous patterns');
});
```

---

### AC-002: 위험한 패턴 탐지 - Prototype 접근

```gherkin
Given: 템플릿 파일이 "{{prototype}}" 패턴을 포함할 때
When: copyTemplateFile() 함수를 실행
Then: InstallationError가 발생해야 한다
  And: 템플릿 렌더링이 중단되어야 한다
  And: 출력 파일이 생성되지 않아야 한다
```

**테스트 코드**:
```typescript
test('AC-002: Reject template with prototype pattern', async () => {
  // Given
  const templatePath = '/tmp/dangerous-prototype.md';
  const templateContent = '{{prototype.polluted}}';
  await fs.promises.writeFile(templatePath, templateContent);

  const outputPath = '/tmp/output.md';

  // When
  try {
    await templateProcessor.copyTemplateFile(
      templatePath,
      outputPath,
      { PROJECT_NAME: 'test' }
    );
  } catch (error) {
    // Expected
  }

  // Then
  expect(fs.existsSync(outputPath)).toBe(false);
});
```

---

### AC-003: 위험한 패턴 탐지 - eval 호출

```gherkin
Given: 템플릿 파일이 "eval(" 패턴을 포함할 때
When: copyTemplateFile() 함수를 실행
Then: InstallationError가 발생해야 한다
  And: 보안 로그가 기록되어야 한다
```

**테스트 코드**:
```typescript
test('AC-003: Reject template with eval pattern', async () => {
  // Given
  const templateContent = 'eval(malicious_code)';

  // When & Then
  await expect(
    templateProcessor.copyTemplateFile(
      '/tmp/eval.md',
      '/tmp/out.md',
      {}
    )
  ).rejects.toThrow('dangerous patterns');
});
```

---

### AC-004: 화이트리스트 검증 - 허용된 변수

```gherkin
Given: 템플릿 컨텍스트가 PROJECT_NAME 변수를 포함할 때
  And: PROJECT_NAME이 ALLOWED_CONTEXT_KEYS에 포함되어 있을 때
When: sanitizeTemplateContext() 함수를 실행
Then: PROJECT_NAME 변수가 유지되어야 한다
  And: 경고 메시지가 발생하지 않아야 한다
```

**테스트 코드**:
```typescript
test('AC-004: Allow whitelisted variable PROJECT_NAME', async () => {
  // Given
  const templateContent = 'Project: {{PROJECT_NAME}}';
  const variables = { PROJECT_NAME: 'MyProject' };

  // When
  await templateProcessor.copyTemplateFile(
    '/tmp/safe.md',
    '/tmp/out.md',
    variables
  );

  // Then
  const output = await fs.promises.readFile('/tmp/out.md', 'utf-8');
  expect(output).toContain('Project: MyProject');
});
```

---

### AC-005: 화이트리스트 검증 - 거부된 변수

```gherkin
Given: 템플릿 컨텍스트가 __proto__ 속성을 포함할 때
  And: __proto__가 DANGEROUS_PROPERTIES에 포함되어 있을 때
When: sanitizeTemplateContext() 함수를 실행
Then: __proto__ 속성이 제거되어야 한다
  And: 경고 로그가 기록되어야 한다
  And: removedKeys 배열에 '__proto__'가 포함되어야 한다
```

**테스트 코드**:
```typescript
test('AC-005: Remove dangerous property __proto__', async () => {
  // Given
  const variables = {
    PROJECT_NAME: 'test',
    __proto__: 'malicious'
  };

  const mockLogger = jest.spyOn(logger, 'warn');

  // When
  await templateProcessor.copyTemplateFile(
    '/tmp/template.md',
    '/tmp/out.md',
    variables
  );

  // Then
  expect(mockLogger).toHaveBeenCalledWith(
    expect.stringContaining('Template context sanitized'),
    expect.objectContaining({
      warnings: expect.arrayContaining([
        expect.stringContaining('__proto__')
      ])
    })
  );
});
```

---

### AC-006: 안전한 템플릿 처리

```gherkin
Given: 템플릿 파일이 안전한 Mustache 변수만 포함할 때
  And: 모든 변수가 화이트리스트에 포함되어 있을 때
When: copyTemplateFile() 함수를 실행
Then: 템플릿이 정상적으로 렌더링되어야 한다
  And: 출력 파일이 생성되어야 한다
  And: 에러가 발생하지 않아야 한다
```

**테스트 코드**:
```typescript
test('AC-006: Process safe template successfully', async () => {
  // Given
  const templateContent = `
# {{PROJECT_NAME}}

Version: {{PROJECT_VERSION}}
Description: {{PROJECT_DESCRIPTION}}
`;
  const variables = {
    PROJECT_NAME: 'MoAI-ADK',
    PROJECT_VERSION: '1.0.0',
    PROJECT_DESCRIPTION: 'Test project'
  };

  // When
  await expect(
    templateProcessor.copyTemplateFile(
      '/tmp/safe.md',
      '/tmp/out.md',
      variables
    )
  ).resolves.not.toThrow();

  // Then
  const output = await fs.promises.readFile('/tmp/out.md', 'utf-8');
  expect(output).toContain('# MoAI-ADK');
  expect(output).toContain('Version: 1.0.0');
  expect(output).toContain('Description: Test project');
});
```

---

### AC-007: 디렉토리 레벨 보안 검증

```gherkin
Given: 템플릿 디렉토리 내 한 파일이 위험한 패턴을 포함할 때
When: copyTemplateDirectory() 함수를 실행
Then: 전체 디렉토리 복사가 중단되어야 한다
  And: 에러 메시지에 위험한 파일 경로가 포함되어야 한다
  And: 부분적으로 복사된 파일이 없어야 한다
```

**테스트 코드**:
```typescript
test('AC-007: Abort directory copy on dangerous file', async () => {
  // Given
  const tempDir = '/tmp/templates';
  await fs.promises.mkdir(tempDir, { recursive: true });
  await fs.promises.writeFile(
    `${tempDir}/safe.md`,
    '{{PROJECT_NAME}}'
  );
  await fs.promises.writeFile(
    `${tempDir}/dangerous.md`,
    '{{constructor}}'
  );

  const outputDir = '/tmp/output';

  // When
  await expect(
    templateProcessor.copyTemplateDirectory(
      tempDir,
      outputDir,
      { PROJECT_NAME: 'test' }
    )
  ).rejects.toThrow();

  // Then
  // 부분 복사 확인
  const exists = fs.existsSync(outputDir);
  if (exists) {
    const files = await fs.promises.readdir(outputDir);
    expect(files.length).toBe(0); // 롤백되어야 함
  }
});
```

---

### AC-008: 성능 기준 - 대량 템플릿 처리

```gherkin
Given: 1000개의 안전한 템플릿 파일이 존재할 때
When: 각 파일에 대해 copyTemplateFile()을 실행
Then: 전체 처리 시간이 1초 미만이어야 한다
  And: 메모리 사용량 증가가 10MB 미만이어야 한다
```

**테스트 코드**:
```typescript
test('AC-008: Process 1000 templates in < 1 second', async () => {
  // Given
  const templateCount = 1000;
  const templates = Array.from({ length: templateCount }, (_, i) => ({
    path: `/tmp/template-${i}.md`,
    content: `# {{PROJECT_NAME}}-${i}`
  }));

  // Setup
  for (const template of templates) {
    await fs.promises.writeFile(template.path, template.content);
  }

  // When
  const startTime = Date.now();
  const startMemory = process.memoryUsage().heapUsed;

  for (const template of templates) {
    await templateProcessor.copyTemplateFile(
      template.path,
      `/tmp/out-${template.path}`,
      { PROJECT_NAME: 'test' }
    );
  }

  const duration = Date.now() - startTime;
  const memoryIncrease = process.memoryUsage().heapUsed - startMemory;

  // Then
  expect(duration).toBeLessThan(1000); // < 1 second
  expect(memoryIncrease).toBeLessThan(10 * 1024 * 1024); // < 10MB
});
```

---

### AC-009: 에러 메시지 명확성

```gherkin
Given: 템플릿 보안 검증이 실패했을 때
When: 에러 메시지를 확인
Then: 어떤 파일에서 에러가 발생했는지 명시되어야 한다
  And: 어떤 패턴이 위험한지 명시되어야 한다
  And: 권장 조치사항이 포함되어야 한다
```

**테스트 코드**:
```typescript
test('AC-009: Provide clear error message', async () => {
  // Given
  const dangerousTemplate = '/tmp/bad.md';
  await fs.promises.writeFile(dangerousTemplate, '{{__proto__}}');

  // When
  try {
    await templateProcessor.copyTemplateFile(
      dangerousTemplate,
      '/tmp/out.md',
      {}
    );
    fail('Should have thrown error');
  } catch (error) {
    // Then
    expect(error.message).toContain('dangerous patterns');
    expect(error.message).toContain(dangerousTemplate);
    // 권장 조치사항 확인 (선택)
  }
});
```

---

### AC-010: 화이트리스트 확장 가능성

```gherkin
Given: ALLOWED_CONTEXT_KEYS에 새로운 변수를 추가했을 때
When: 해당 변수를 포함한 템플릿을 처리
Then: 새 변수가 정상적으로 렌더링되어야 한다
  And: 보안 경고가 발생하지 않아야 한다
```

**테스트 코드**:
```typescript
test('AC-010: Support whitelist extension', async () => {
  // Given
  // ALLOWED_CONTEXT_KEYS에 'CUSTOM_VAR' 추가 (전제)
  const variables = {
    PROJECT_NAME: 'test',
    CUSTOM_VAR: 'custom value'
  };

  // When
  await templateProcessor.copyTemplateFile(
    '/tmp/template.md',
    '/tmp/out.md',
    variables
  );

  // Then
  const output = await fs.promises.readFile('/tmp/out.md', 'utf-8');
  expect(output).toContain('custom value');
});
```

---

## 📊 테스트 커버리지 목표

| 메트릭 | 목표 | 측정 방법 |
|--------|------|-----------|
| Statements | ≥ 85% | `vitest --coverage` |
| Branches | ≥ 80% | `vitest --coverage` |
| Functions | ≥ 85% | `vitest --coverage` |
| Lines | ≥ 85% | `vitest --coverage` |

---

## ✅ 전체 체크리스트

### 기능 요구사항
- [ ] AC-001: Constructor 패턴 탐지
- [ ] AC-002: Prototype 패턴 탐지
- [ ] AC-003: eval 패턴 탐지
- [ ] AC-004: 화이트리스트 허용 변수
- [ ] AC-005: 위험한 속성 제거
- [ ] AC-006: 안전한 템플릿 처리
- [ ] AC-007: 디렉토리 레벨 검증
- [ ] AC-008: 성능 기준 (1000 templates < 1s)
- [ ] AC-009: 명확한 에러 메시지
- [ ] AC-010: 화이트리스트 확장 가능

### 비기능 요구사항
- [ ] 테스트 커버리지 85% 이상
- [ ] 모든 테스트 통과
- [ ] LOC 제한 준수 (≤300)
- [ ] 성능 기준 달성
- [ ] 메모리 사용량 < 10MB 증가

### 문서화
- [ ] SPEC 문서 완성
- [ ] 구현 계획서 완성
- [ ] 인수 기준 완성
- [ ] 코드 주석 및 JSDoc
- [ ] CHANGELOG 업데이트

---

## 🎯 인수 승인 기준

**모든 AC-001 ~ AC-010이 통과**하고 **테스트 커버리지 85% 이상**일 때 인수 승인됩니다.

**검증 방법**:
```bash
# 테스트 실행
npm test -- template-security.test.ts

# 커버리지 확인
npm test -- --coverage

# 전체 빌드 확인
npm run build
```

**승인자**: @goos
**승인일**: (구현 완료 후 기입)
