# SPEC-006 수락 기준

## @TAG BLOCK
```
# @TEST:UTIL-006 | Chain: @SPEC:QUAL-006 -> @SPEC:TEST-006 -> @CODE:TEST-006 -> @TEST:UTIL-006
# Related: @CODE:VALID-002:API, @CODE:INPUT-001, LOG-002
```

## 개요

본 문서는 SPEC-006 "Utils 모듈 테스트 작성"의 상세한 수락 기준을 정의합니다.
Given-When-Then 형식의 테스트 시나리오를 통해 검증 가능한 조건을 명시합니다.

## 1. errors.test.ts 수락 기준

### 1.1 커스텀 에러 클래스 생성

#### AC-001: ValidationError 생성 및 속성 검증
**Given**: ValidationError를 생성할 때
**When**: 메시지와 옵션을 전달하면
**Then**:
- error.name이 'ValidationError'이어야 함
- error.message가 전달된 메시지와 일치해야 함
- error.pattern이 옵션에서 설정한 값과 일치해야 함
- error.vulnerabilities가 배열로 존재해야 함
- error instanceof ValidationError가 true여야 함
- error instanceof Error가 true여야 함

```typescript
test('should create ValidationError with properties', () => {
  // Given
  const message = 'Invalid input detected';
  const options = {
    pattern: '../traversal',
    vulnerabilities: ['path-traversal', 'injection'],
    context: { input: 'test' },
  };

  // When
  const error = new ValidationError(message, options);

  // Then
  expect(error.name).toBe('ValidationError');
  expect(error.message).toBe(message);
  expect(error.pattern).toBe(options.pattern);
  expect(error.vulnerabilities).toEqual(options.vulnerabilities);
  expect(error.context).toEqual(options.context);
  expect(error instanceof ValidationError).toBe(true);
  expect(error instanceof Error).toBe(true);
});
```

#### AC-002: InstallationError 생성 검증
**Given**: InstallationError를 생성할 때
**When**: 에러 객체와 메타데이터를 전달하면
**Then**:
- error.name이 'InstallationError'여야 함
- error.error가 원본 에러 객체여야 함
- error.projectPath가 설정되어야 함
- error.phase가 설정되어야 함

#### AC-003: TemplateError 생성 검증
**Given**: TemplateError를 생성할 때
**When**: 템플릿 경로를 전달하면
**Then**:
- error.name이 'TemplateError'여야 함
- error.templatePath가 설정되어야 함

#### AC-004: ResourceError 생성 검증
**Given**: ResourceError를 생성할 때
**When**: 리소스 경로를 전달하면
**Then**:
- error.name이 'ResourceError'여야 함
- error.resourcePath가 설정되어야 함

#### AC-005: PhaseError 생성 검증
**Given**: PhaseError를 생성할 때
**When**: 단계 정보를 전달하면
**Then**:
- error.name이 'PhaseError'여야 함
- error.phase가 설정되어야 함

### 1.2 타입 가드 함수

#### AC-006: isValidationError 정확성
**Given**: 다양한 타입의 에러가 있을 때
**When**: isValidationError()로 검증하면
**Then**:
- ValidationError에 대해서만 true를 반환해야 함
- 다른 커스텀 에러는 false를 반환해야 함
- 일반 Error는 false를 반환해야 함
- null/undefined는 false를 반환해야 함

```typescript
test('should correctly identify ValidationError', () => {
  // Given
  const validationError = new ValidationError('test');
  const installationError = new InstallationError('test');
  const genericError = new Error('test');

  // When/Then
  expect(isValidationError(validationError)).toBe(true);
  expect(isValidationError(installationError)).toBe(false);
  expect(isValidationError(genericError)).toBe(false);
  expect(isValidationError(null)).toBe(false);
  expect(isValidationError(undefined)).toBe(false);
});
```

#### AC-007 ~ AC-010: 나머지 타입 가드
- isInstallationError()
- isTemplateError()
- isResourceError()
- isPhaseError()

각각 동일한 패턴으로 정확성 검증

### 1.3 헬퍼 함수

#### AC-011: toError() - Error 객체 변환
**Given**: 이미 Error 객체일 때
**When**: toError()를 호출하면
**Then**: 동일한 Error 객체를 반환해야 함

#### AC-012: toError() - 문자열 변환
**Given**: 문자열이 전달될 때
**When**: toError()를 호출하면
**Then**: 문자열을 메시지로 하는 Error 객체를 생성해야 함

#### AC-013: toError() - unknown 타입 처리
**Given**: 숫자, 객체, null 등이 전달될 때
**When**: toError()를 호출하면
**Then**: String() 변환 후 Error 객체를 생성해야 함

#### AC-014: getErrorMessage() - Error 메시지 추출
**Given**: Error 객체가 있을 때
**When**: getErrorMessage()를 호출하면
**Then**: error.message를 반환해야 함

#### AC-015: getErrorMessage() - 문자열 처리
**Given**: 문자열이 전달될 때
**When**: getErrorMessage()를 호출하면
**Then**: 동일한 문자열을 반환해야 함

### 커버리지 검증
- [ ] Statement Coverage: 100%
- [ ] Branch Coverage: 100%
- [ ] Function Coverage: 100%
- [ ] Line Coverage: 100%

---

## 2. input-validator.test.ts 수락 기준

### 2.1 validateProjectName() 테스트

#### AC-101: 유효한 프로젝트명 검증
**Given**: 유효한 프로젝트명 "my-project"가 있을 때
**When**: validateProjectName()을 호출하면
**Then**:
- result.isValid가 true여야 함
- result.errors가 빈 배열이어야 함
- result.sanitizedValue가 입력값과 일치해야 함

```typescript
test('should accept valid project name', () => {
  // Given
  const validName = 'my-project';

  // When
  const result = validateProjectName(validName);

  // Then
  expect(result.isValid).toBe(true);
  expect(result.errors).toHaveLength(0);
  expect(result.sanitizedValue).toBe(validName);
});
```

#### AC-102: 최소 길이 위반 검증
**Given**: 너무 짧은 프로젝트명 ""가 있을 때
**When**: validateProjectName()을 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "must be at least" 메시지가 포함되어야 함

#### AC-103: 최대 길이 위반 검증
**Given**: 51자 이상의 프로젝트명이 있을 때
**When**: validateProjectName()을 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "must not exceed" 메시지가 포함되어야 함

#### AC-104: 공백 금지 검증
**Given**: 공백이 포함된 "my project"가 있을 때
**When**: allowSpaces 없이 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "cannot contain spaces" 메시지가 포함되어야 함

#### AC-105: 공백 허용 검증
**Given**: 공백이 포함된 "my project"가 있을 때
**When**: allowSpaces: true로 호출하면
**Then**:
- result.isValid가 true여야 함

#### AC-106: 특수문자 금지 검증
**Given**: 특수문자가 포함된 "my@project"가 있을 때
**When**: allowSpecialChars 없이 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "cannot contain special characters" 메시지가 포함되어야 함

#### AC-107: Path Traversal 방지
**Given**: "../dangerous"와 같은 위험 패턴이 있을 때
**When**: validateProjectName()을 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "invalid characters or patterns" 메시지가 포함되어야 함

#### AC-108: Windows 예약어 검증
**Given**: "CON", "PRN" 등 Windows 예약어가 있을 때
**When**: validateProjectName()을 호출하면
**Then**:
- result.isValid가 false여야 함

#### AC-109: sanitizedValue 생성
**Given**: 특수문자가 포함된 "my@project!"가 있을 때
**When**: validateProjectName()을 호출하면
**Then**:
- result.sanitizedValue가 "my-project-"로 변환되어야 함

### 2.2 validatePath() 테스트

#### AC-110: 유효한 경로 검증
**Given**: 유효한 경로 "/valid/path"가 있을 때
**When**: validatePath()를 호출하면
**Then**:
- result.isValid가 true여야 함
- result.sanitizedValue가 정규화된 경로여야 함

```typescript
test('should accept valid path', async () => {
  // Given
  const validPath = tempDir; // 실제 존재하는 임시 디렉토리

  // When
  const result = await validatePath(validPath);

  // Then
  expect(result.isValid).toBe(true);
  expect(result.sanitizedValue).toBeDefined();
});
```

#### AC-111: 경로 길이 제한 검증
**Given**: 260자를 초과하는 경로가 있을 때
**When**: validatePath()를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "too long" 메시지가 포함되어야 함

#### AC-112: 위험 패턴 검증
**Given**: "/etc/passwd"와 같은 시스템 경로가 있을 때
**When**: validatePath()를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "dangerous patterns" 메시지가 포함되어야 함

#### AC-113: 경로 깊이 검증
**Given**: maxDepth를 초과하는 경로가 있을 때
**When**: validatePath(path, { maxDepth: 3 })을 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "depth exceeds" 메시지가 포함되어야 함

#### AC-114: mustExist 검증
**Given**: 존재하지 않는 경로가 있을 때
**When**: validatePath(path, { mustExist: true })를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "does not exist" 메시지가 포함되어야 함

#### AC-115: mustBeDirectory 검증
**Given**: 파일 경로가 있을 때
**When**: validatePath(filePath, { mustBeDirectory: true })를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "must be a directory" 메시지가 포함되어야 함

#### AC-116: mustBeFile 검증
**Given**: 디렉토리 경로가 있을 때
**When**: validatePath(dirPath, { mustBeFile: true })를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "must be a file" 메시지가 포함되어야 함

#### AC-117: allowedExtensions 검증
**Given**: .txt 파일 경로가 있을 때
**When**: validatePath(path, { allowedExtensions: ['.md', '.json'] })를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "extension not allowed" 메시지가 포함되어야 함

#### AC-118: Path Traversal 방지
**Given**: "../../../etc/passwd"와 같은 경로가 있을 때
**When**: validatePath()를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "path traversal" 메시지가 포함되어야 함

#### AC-119: 정규화된 경로 반환
**Given**: 상대 경로 "./test"가 있을 때
**When**: validatePath()를 호출하면
**Then**:
- result.sanitizedValue가 절대 경로로 변환되어야 함

### 2.3 validateTemplateType() 테스트

#### AC-120: 유효한 템플릿 타입
**Given**: "standard", "minimal", "advanced", "custom" 중 하나가 있을 때
**When**: validateTemplateType()을 호출하면
**Then**:
- result.isValid가 true여야 함
- result.sanitizedValue가 소문자로 변환되어야 함

#### AC-121: 대소문자 무관 검증
**Given**: "STANDARD", "Standard" 등이 있을 때
**When**: validateTemplateType()을 호출하면
**Then**:
- result.isValid가 true여야 함
- result.sanitizedValue가 "standard"여야 함

#### AC-122: 허용되지 않는 타입
**Given**: "invalid-type"이 있을 때
**When**: validateTemplateType()을 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "Invalid template type" 메시지가 포함되어야 함

#### AC-123: 빈 문자열 처리
**Given**: ""가 전달될 때
**When**: validateTemplateType()을 호출하면
**Then**:
- result.isValid가 false여야 함

### 2.4 validateBranchName() 테스트

#### AC-124: 유효한 브랜치명
**Given**: "feature/login-page"가 있을 때
**When**: validateBranchName()을 호출하면
**Then**:
- result.isValid가 true여야 함

#### AC-125: 길이 제한 검증
**Given**: 250자를 초과하는 브랜치명이 있을 때
**When**: validateBranchName()을 호출하면
**Then**:
- result.isValid가 false여야 함

#### AC-126: 더블 닷 금지
**Given**: "feature..login"이 있을 때
**When**: validateBranchName()을 호출하면
**Then**:
- result.isValid가 false여야 함

#### AC-127: 시작/끝 문자 검증
**Given**: ".feature" 또는 "feature-"가 있을 때
**When**: validateBranchName()을 호출하면
**Then**:
- result.isValid가 false여야 함

#### AC-128: 예약어 검증
**Given**: "HEAD", "master", "origin"이 있을 때
**When**: validateBranchName()을 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "reserved" 메시지가 포함되어야 함

#### AC-129: 금지된 문자 검증
**Given**: "feature:login" 또는 "feature*login"이 있을 때
**When**: validateBranchName()을 호출하면
**Then**:
- result.isValid가 false여야 함

#### AC-130: 공백 금지
**Given**: "feature login"이 있을 때
**When**: validateBranchName()을 호출하면
**Then**:
- result.isValid가 false여야 함

### 2.5 validateCommandOptions() 테스트

#### AC-131: 유효한 옵션 검증
**Given**: { verbose: true, output: "json", timeout: 5000 }이 있을 때
**When**: validateCommandOptions()를 호출하면
**Then**:
- result.isValid가 true여야 함
- result.sanitizedValue에 모든 옵션이 포함되어야 함

#### AC-132: 잘못된 옵션 키
**Given**: { "123invalid": "value" }가 있을 때
**When**: validateCommandOptions()를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "Invalid option key" 메시지가 포함되어야 함

#### AC-133: 너무 긴 문자열 값
**Given**: 1000자를 초과하는 문자열 값이 있을 때
**When**: validateCommandOptions()를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "too long" 메시지가 포함되어야 함

#### AC-134: 위험 문자열 검증
**Given**: "javascript:alert(1)" 같은 값이 있을 때
**When**: validateCommandOptions()를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "dangerous content" 메시지가 포함되어야 함

#### AC-135: 지원되는 타입
**Given**: string, boolean, number 타입이 있을 때
**When**: validateCommandOptions()를 호출하면
**Then**:
- 모든 타입이 sanitizedValue에 포함되어야 함

#### AC-136: 지원되지 않는 타입
**Given**: { obj: { nested: "value" } }가 있을 때
**When**: validateCommandOptions()를 호출하면
**Then**:
- result.isValid가 false여야 함
- result.errors에 "unsupported type" 메시지가 포함되어야 함

### 커버리지 검증
- [ ] Statement Coverage: 90% 이상
- [ ] Branch Coverage: 85% 이상
- [ ] Function Coverage: 100%
- [ ] 모든 public API 테스트 존재

---

## 3. winston-logger.test.ts 수락 기준

### 3.1 로거 초기화

#### AC-201: 기본 옵션으로 초기화
**Given**: 옵션 없이 로거를 생성할 때
**When**: new MoaiLogger()를 호출하면
**Then**:
- logger 인스턴스가 생성되어야 함
- 기본 로그 레벨이 설정되어야 함
- 콘솔 transport가 활성화되어야 함

```typescript
test('should create logger with default options', () => {
  // Given/When
  const logger = new MoaiLogger();

  // Then
  expect(logger).toBeDefined();
  expect(logger).toBeInstanceOf(MoaiLogger);
});
```

#### AC-202: 커스텀 옵션으로 초기화
**Given**: { level: 'debug', enableConsole: false }가 있을 때
**When**: new MoaiLogger(options)를 호출하면
**Then**:
- logger.level이 'debug'여야 함
- 콘솔 transport가 비활성화되어야 함

#### AC-203: Mock transport 주입
**Given**: 테스트용 mock transport가 있을 때
**When**: LoggerOptions.transports로 주입하면
**Then**:
- mock transport가 사용되어야 함
- 실제 파일이 생성되지 않아야 함

#### AC-204: 로그 디렉토리 자동 생성
**Given**: 존재하지 않는 로그 디렉토리가 있을 때
**When**: enableFile: true로 로거를 생성하면
**Then**:
- 로그 디렉토리가 자동으로 생성되어야 함

#### AC-205: 파일 로깅 실패 시 fallback
**Given**: 파일 쓰기 권한이 없을 때
**When**: enableFile: true로 로거를 생성하면
**Then**:
- 에러를 throw하지 않아야 함
- 콘솔 전용으로 동작해야 함
- stderr에 경고 메시지가 출력되어야 함

### 3.2 로그 레벨 메서드

#### AC-206: debug() 메서드
**Given**: debug 레벨 로거가 있을 때
**When**: logger.debug('test message')를 호출하면
**Then**:
- 로그가 기록되어야 함
- level이 'debug'여야 함

#### AC-207: info() 메서드
**Given**: 로거가 있을 때
**When**: logger.info('info message', { key: 'value' })를 호출하면
**Then**:
- 메시지와 메타데이터가 기록되어야 함
- level이 'info'여야 함

#### AC-208: warn() 메서드
**Given**: 로거가 있을 때
**When**: logger.warn('warning')을 호출하면
**Then**:
- level이 'warn'이어야 함

#### AC-209: error() 메서드 - Error 객체
**Given**: Error 객체가 있을 때
**When**: logger.error('error', error)를 호출하면
**Then**:
- error.name, error.message, error.stack이 메타데이터에 포함되어야 함

#### AC-210: error() 메서드 - unknown 타입
**Given**: Error가 아닌 값이 있을 때
**When**: logger.error('error', unknownValue)를 호출하면
**Then**:
- 메타데이터에 에러 정보가 포함되어야 함

### 3.3 민감 정보 마스킹

#### AC-211: 필드명 기반 마스킹
**Given**: { username: 'test', password: 'secret' }가 있을 때
**When**: logger.info('login', metadata)를 호출하면
**Then**:
- password 필드가 '***REDACTED***'로 마스킹되어야 함
- username은 그대로 유지되어야 함

```typescript
test('should mask sensitive fields', () => {
  // Given
  const mockTransport = {
    log: vi.fn(),
  };
  const logger = new MoaiLogger({
    transports: [mockTransport],
  });
  const sensitiveData = {
    username: 'test',
    password: 'secret123',
    apiKey: 'key-abc-123',
  };

  // When
  logger.info('User data', sensitiveData);

  // Then
  expect(mockTransport.log).toHaveBeenCalled();
  const loggedData = mockTransport.log.mock.calls[0][0];
  expect(loggedData.password).toBe('***REDACTED***');
  expect(loggedData.apiKey).toBe('***REDACTED***');
  expect(loggedData.username).toBe('test');
});
```

#### AC-212: 대소문자 무관 마스킹
**Given**: { Password: 'test', API_KEY: 'key' }가 있을 때
**When**: 로그를 기록하면
**Then**:
- 대소문자에 관계없이 마스킹되어야 함

#### AC-213: 메시지 문자열 패턴 마스킹
**Given**: "password=secret123 token=abc"가 있을 때
**When**: logger.info(message)를 호출하면
**Then**:
- "password=***redacted*** token=***redacted***"로 변환되어야 함

#### AC-214: 중첩 객체 마스킹
**Given**: { user: { password: 'secret' } }가 있을 때
**When**: 로그를 기록하면
**Then**:
- 중첩된 password도 마스킹되어야 함

#### AC-215: Bearer 토큰 마스킹
**Given**: "Authorization: Bearer abc123"가 있을 때
**When**: 로그를 기록하면
**Then**:
- Bearer 토큰이 마스킹되어야 함

#### AC-216: 다양한 민감 패턴
**Given**: "api_key=test secret=value"가 있을 때
**When**: 로그를 기록하면
**Then**:
- 모든 민감 패턴이 마스킹되어야 함

#### AC-217: 배열 처리
**Given**: { tokens: ['token1', 'token2'] }가 있을 때
**When**: 로그를 기록하면
**Then**:
- 배열은 그대로 유지되어야 함 (필드명만 검사)

#### AC-218: null/undefined 처리
**Given**: { password: null, token: undefined }가 있을 때
**When**: 로그를 기록하면
**Then**:
- 에러가 발생하지 않아야 함

### 3.4 Verbose 모드

#### AC-219: Verbose 모드 활성화
**Given**: 로거가 있을 때
**When**: logger.setVerbose(true)를 호출하면
**Then**:
- logger.isVerbose()가 true를 반환해야 함
- logger.level이 'debug'로 변경되어야 함

#### AC-220: Verbose 모드 비활성화
**Given**: Verbose 모드가 활성화된 로거가 있을 때
**When**: logger.setVerbose(false)를 호출하면
**Then**:
- logger.isVerbose()가 false를 반환해야 함
- logger.level이 원래 레벨로 복구되어야 함

#### AC-221: Verbose 전용 메시지
**Given**: Verbose 모드가 활성화되어 있을 때
**When**: logger.verbose('debug info')를 호출하면
**Then**:
- 로그가 기록되어야 함

#### AC-222: Verbose 모드 꺼진 상태
**Given**: Verbose 모드가 비활성화되어 있을 때
**When**: logger.verbose('debug info')를 호출하면
**Then**:
- 로그가 기록되지 않아야 함

### 3.5 TAG 추적성

#### AC-223: logWithTag() 메서드
**Given**: TAG 정보가 있을 때
**When**: logger.logWithTag('info', '@SPEC:TEST-001', 'message')를 호출하면
**Then**:
- 메타데이터에 tag 필드가 포함되어야 함
- tag 값이 '@SPEC:TEST-001'이어야 함

#### AC-224: TAG 정보 보존
**Given**: 추가 메타데이터가 있을 때
**When**: logWithTag()를 호출하면
**Then**:
- TAG와 기존 메타데이터가 모두 보존되어야 함

### 3.6 사용자 메시지 출력

#### AC-225: log() 메서드
**Given**: 로거가 있을 때
**When**: logger.log('User message')를 호출하면
**Then**:
- console.log가 호출되어야 함
- Winston을 거치지 않아야 함

#### AC-226: success() 메서드
**Given**: 로거가 있을 때
**When**: logger.success('Success!')를 호출하면
**Then**:
- console.log가 호출되어야 함

#### AC-227: errorMessage() 메서드
**Given**: 로거가 있을 때
**When**: logger.errorMessage('Error!')를 호출하면
**Then**:
- console.error가 호출되어야 함

### 커버리지 검증
- [ ] Statement Coverage: 85% 이상
- [ ] Branch Coverage: 80% 이상
- [ ] 민감 정보 마스킹: 100%
- [ ] 핵심 로깅 메서드: 100%

---

## 4. 전체 통합 검증

### 4.1 테스트 실행

#### AC-301: 모든 테스트 통과
**Given**: 3개의 테스트 파일이 작성되어 있을 때
**When**: `npm test`를 실행하면
**Then**:
- 모든 테스트가 통과해야 함
- 실패한 테스트가 0개여야 함

#### AC-302: 테스트 격리
**Given**: 테스트 파일들이 있을 때
**When**: 순서를 바꿔 실행하면
**Then**:
- 동일한 결과가 나와야 함
- 테스트 간 상태 공유가 없어야 함

#### AC-303: 성능 기준
**Given**: 테스트 스위트가 있을 때
**When**: 실행하면
**Then**:
- 전체 실행 시간이 10초 미만이어야 함
- 개별 테스트가 100ms 미만이어야 함

### 4.2 커버리지 검증

#### AC-304: errors.ts 커버리지
**Given**: errors.test.ts가 작성되어 있을 때
**When**: `npm run test:coverage`를 실행하면
**Then**:
- errors.ts의 Statement Coverage가 100%여야 함
- Branch Coverage가 100%여야 함

#### AC-305: input-validator.ts 커버리지
**Given**: input-validator.test.ts가 작성되어 있을 때
**When**: 커버리지를 확인하면
**Then**:
- Statement Coverage가 90% 이상이어야 함
- Branch Coverage가 85% 이상이어야 함

#### AC-306: winston-logger.ts 커버리지
**Given**: winston-logger.test.ts가 작성되어 있을 때
**When**: 커버리지를 확인하면
**Then**:
- Statement Coverage가 85% 이상이어야 함
- Branch Coverage가 80% 이상이어야 함

#### AC-307: 전체 utils 모듈 커버리지
**Given**: 3개의 테스트가 작성되어 있을 때
**When**: 전체 커버리지를 확인하면
**Then**:
- Utils 모듈 전체 커버리지가 50% 이상이어야 함

### 4.3 품질 게이트

#### AC-308: ESLint 검사
**Given**: 테스트 코드가 작성되어 있을 때
**When**: `npm run lint`를 실행하면
**Then**:
- 린트 에러가 없어야 함

#### AC-309: TypeScript 컴파일
**Given**: 테스트 코드가 작성되어 있을 때
**When**: `npm run type-check`를 실행하면
**Then**:
- 타입 에러가 없어야 함

#### AC-310: 빌드 성공
**Given**: 모든 테스트가 작성되어 있을 때
**When**: `npm run build`를 실행하면
**Then**:
- 빌드가 성공해야 함

### 4.4 @TAG 추적성

#### AC-311: TAG BLOCK 존재
**Given**: 3개의 테스트 파일이 있을 때
**When**: 파일 상단을 확인하면
**Then**:
- 모든 파일에 TAG BLOCK이 존재해야 함
- @TEST TAG가 포함되어야 함

#### AC-312: TAG 체인 검증
**Given**: 테스트 파일들이 작성되어 있을 때
**When**: `rg '@TEST:UTIL-006' -n`을 실행하면
**Then**:
- 모든 테스트 파일에서 TAG가 발견되어야 함

#### AC-313: 소스 코드 TAG 연결
**Given**: 테스트와 소스 파일이 있을 때
**When**: TAG 체인을 추적하면
**Then**:
- @SPEC → @SPEC → @CODE → @TEST 체인이 완전해야 함

## 완료 조건 (Definition of Done)

### 필수 조건
- [ ] 3개 테스트 파일 작성 완료 (errors, input-validator, winston-logger)
- [ ] 모든 AC 시나리오 구현 및 검증
- [ ] 모든 테스트 통과 (실패 0개)
- [ ] 커버리지 목표 달성 (errors: 100%, input-validator: 90%+, winston-logger: 85%+)
- [ ] 품질 게이트 통과 (ESLint, TypeScript, Build)

### 문서화
- [ ] @TAG BLOCK 모든 파일에 존재
- [ ] Given-When-Then 주석 작성
- [ ] describe/test 이름 명확화

### 코드 품질
- [ ] 테스트 파일당 300 LOC 이하
- [ ] 테스트 함수당 50 LOC 이하
- [ ] 복잡도 10 이하

### 실행 성능
- [ ] 전체 테스트 < 10초
- [ ] 개별 테스트 < 100ms
- [ ] 파일 시스템 cleanup 완료

---

**작성일**: 2025-10-01
**버전**: 1.0.0
**상태**: Draft
