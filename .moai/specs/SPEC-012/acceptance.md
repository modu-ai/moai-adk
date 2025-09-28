# SPEC-012 수락 기준: TypeScript 기반 구축 (Week 1)

> **@TEST:ACCEPTANCE-CRITERIA-012** SPEC-012 수락을 위한 상세 검증 기준
> **@FEATURE:VALIDATION-SCENARIOS-012** Given-When-Then 형식의 테스트 시나리오
> **@QUALITY:DEFINITION-OF-DONE-012** Week 1 완료를 위한 정량적 기준

---

## 수락 기준 개요

### 기능적 요구사항

| 요구사항 ID | 설명 | 우선순위 | 수락 기준 |
|------------|------|---------|----------|
| **AC-012-01** | 시스템 요구사항 자동 검증 | 높음 | Node.js, Git, SQLite3 자동 감지 |
| **AC-012-02** | CLI 기본 명령어 동작 | 높음 | --version, --help, doctor 동작 |
| **AC-012-03** | TypeScript 빌드 시스템 | 중간 | tsup 빌드 성공, ESM/CJS 지원 |
| **AC-012-04** | 테스트 인프라 | 중간 | Jest 테스트 수트 실행 성공 |
| **AC-012-05** | 크로스 플랫폼 지원 | 낮음 | Windows/macOS/Linux 호환성 |

### 비기능적 요구사항

| 지표 | 예상값 | 수락 기준 | 측정 방법 |
|------|-------|----------|----------|
| **성능** | CLI 시작 < 2초 | ≤ 3초 | `time moai --version` |
| **품질** | 커버리지 ≥ 80% | ≥ 75% | `npm run test:coverage` |
| **보안** | 에러 0개 | 0개 | ESLint + 보안 감사 |
| **유지보수** | 타입 커버리지 100% | 100% | TypeScript strict 모드 |

---

## Given-When-Then 테스트 시나리오

### AC-012-01: 시스템 요구사항 자동 검증

#### 시나리오 1.1: Node.js 감지 성공
```gherkin
Given Node.js 18.17.0이 설치되어 있음
When SystemDetector가 Node.js 요구사항을 검사함
Then 설치된 것으로 표시되고 버전이 올바르게 추출됨
And minVersion 18.0.0 요구사항을 만족함
```

**구현 검증**:
```typescript
test('should detect installed Node.js', async () => {
  const detector = new SystemDetector();
  const requirement = {
    name: 'Node.js',
    category: 'runtime' as const,
    minVersion: '18.0.0',
    installCommands: { darwin: 'brew install node' },
    checkCommand: 'node --version'
  };

  const result = await detector.checkRequirement(requirement);

  expect(result.installed).toBe(true);
  expect(result.satisfied).toBe(true);
  expect(result.version).toMatch(/\d+\.\d+\.\d+/);
  expect(semver.gte(result.version!, '18.0.0')).toBe(true);
});
```

#### 시나리오 1.2: 미설치 도구 처리
```gherkin
Given NonExistentTool이 시스템에 없음
When SystemDetector가 해당 도구를 검사함
Then 설치되지 않은 것으로 표시됨
And 적절한 에러 메시지가 제공됨
And satisfied가 false로 설정됨
```

**구현 검증**:
```typescript
test('should handle missing tool gracefully', async () => {
  const detector = new SystemDetector();
  const requirement = {
    name: 'NonExistentTool',
    category: 'runtime' as const,
    installCommands: { darwin: 'brew install nonexistent' },
    checkCommand: 'nonexistent --version'
  };

  const result = await detector.checkRequirement(requirement);

  expect(result.installed).toBe(false);
  expect(result.satisfied).toBe(false);
  expect(result.error).toBeDefined();
  expect(result.error).toContain('not found');
});
```

#### 시나리오 1.3: 버전 요구사항 검증
```gherkin
Given Node.js 16.20.0이 설치되어 있음
When 최소 버전 18.0.0 요구사항을 검사함
Then 설치된 것으로 표시되지만 satisfied가 false
그리고 required 필드에 최소 버전 명시됨
```

**구현 검증**:
```typescript
test('should validate version requirements', async () => {
  // 모킹으로 낮은 버전 시뮤레이션
  jest.spyOn(require('child_process'), 'exec').mockImplementation(
    (cmd, callback) => callback(null, { stdout: 'v16.20.0' })
  );

  const detector = new SystemDetector();
  const requirement = {
    name: 'Node.js',
    category: 'runtime' as const,
    minVersion: '18.0.0',
    installCommands: { darwin: 'brew install node' },
    checkCommand: 'node --version'
  };

  const result = await detector.checkRequirement(requirement);

  expect(result.installed).toBe(true);
  expect(result.satisfied).toBe(false);
  expect(result.required).toBe('18.0.0');
  expect(result.version).toBe('16.20.0');
});
```

### AC-012-02: CLI 기본 명령어 동작

#### 시나리오 2.1: 버전 정보 표시
```gherkin
Given MoAI-ADK TypeScript CLI가 빌드되어 있음
When 사용자가 'moai --version' 명령어를 실행함
Then 버전 '0.0.1'이 출력됨
And exit code가 0임
And 실행 시간이 2초 이내임
```

**구현 검증**:
```typescript
test('moai --version should return version', async () => {
  const startTime = Date.now();
  const { execSync } = require('child_process');
  const output = execSync('node dist/cli/index.js --version', {
    encoding: 'utf8',
    stdio: 'pipe'
  });
  const executionTime = Date.now() - startTime;

  expect(output.trim()).toBe('0.0.1');
  expect(executionTime).toBeLessThan(2000);
});
```

#### 시나리오 2.2: 도움말 표시
```gherkin
Given MoAI-ADK CLI가 빌드되어 있음
When 사용자가 'moai --help' 명령어를 실행함
Then 프로젝트 설명이 표시됨
And 사용 가능한 명령어 목록이 표시됨
And init, doctor 명령어가 포함됨
```

**구현 검증**:
```typescript
test('moai --help should show comprehensive help', async () => {
  const { execSync } = require('child_process');
  const output = execSync('node dist/cli/index.js --help', { encoding: 'utf8' });

  expect(output).toContain('MoAI-ADK: Modu-AI Agentic Development kit');
  expect(output).toContain('init');
  expect(output).toContain('doctor');
  expect(output).toContain('Options:');
  expect(output).toContain('Commands:');
});
```

#### 시나리오 2.3: doctor 명령어 동작
```gherkin
Given 시스템에 Node.js와 Git이 설치되어 있음
When 사용자가 'moai doctor' 명령어를 실행함
Then '시스템 진단' 헤더가 표시됨
And Node.js 상태가 체크마크와 함께 표시됨
And Git 상태가 체크마크와 함께 표시됨
And 전체 실행 시간이 5초 이내임
```

**구현 검증**:
```typescript
test('moai doctor should show system status', async () => {
  const startTime = Date.now();
  const { execSync } = require('child_process');
  const output = execSync('node dist/cli/index.js doctor', { encoding: 'utf8' });
  const executionTime = Date.now() - startTime;

  expect(output).toContain('MoAI-ADK System Diagnosis');
  expect(output).toContain('Node.js');
  expect(output).toMatch(/[✅❌]/); // 체크마크 존재
  expect(executionTime).toBeLessThan(5000);
});
```

### AC-012-03: TypeScript 빌드 시스템

#### 시나리오 3.1: 성공적인 빌드
```gherkin
Given TypeScript 소스 코드가 src/ 디렉토리에 존재함
When 'npm run build' 명령어를 실행함
Then dist/ 디렉토리에 JavaScript 파일이 생성됨
And .d.ts 타입 정의 파일이 생성됨
And 빌드 시간이 30초 이내임
And 빌드 오류가 0개임
```

**구현 검증**:
```typescript
test('should build successfully', async () => {
  const startTime = Date.now();
  const { execSync } = require('child_process');
  
  // 빌드 실행
  const buildResult = execSync('npm run build', { encoding: 'utf8' });
  const buildTime = Date.now() - startTime;

  // 결과 검증
  expect(buildTime).toBeLessThan(30000);
  expect(fs.existsSync('dist/index.js')).toBe(true);
  expect(fs.existsSync('dist/index.d.ts')).toBe(true);
  expect(fs.existsSync('dist/cli/index.js')).toBe(true);
  expect(buildResult).not.toContain('error');
});
```

#### 시나리오 3.2: 타입 검사 통과
```gherkin
Given TypeScript strict 모드가 활성화되어 있음
When 'npx tsc --noEmit' 명령어를 실행함
Then 타입 오류가 0개임
And 타입 경고가 0개임
And 모든 소스 파일이 올바른 타입을 가짐
```

**구현 검증**:
```typescript
test('should pass type checking', async () => {
  const { execSync } = require('child_process');
  
  try {
    const typeCheckResult = execSync('npx tsc --noEmit', { encoding: 'utf8' });
    // tsc가 성공하면 아무 출력이 없음
    expect(typeCheckResult.trim()).toBe('');
  } catch (error) {
    // 타입 오류 발생 시 테스트 실패
    fail(`Type checking failed: ${error.stdout}`);
  }
});
```

### AC-012-04: 테스트 인프라

#### 시나리오 4.1: Jest 테스트 실행
```gherkin
Given Jest 테스트 수트가 __tests__/ 디렉토리에 존재함
When 'npm test' 명령어를 실행함
Then 모든 테스트가 통과함
And 커버리지가 75% 이상임
And 테스트 실행 시간이 60초 이내임
```

**구현 검증**:
```typescript
test('should run all tests successfully', async () => {
  const startTime = Date.now();
  const { execSync } = require('child_process');
  
  const testResult = execSync('npm test', { encoding: 'utf8' });
  const testTime = Date.now() - startTime;

  expect(testResult).toContain('All tests passed');
  expect(testResult).not.toContain('FAIL');
  expect(testTime).toBeLessThan(60000);
  
  // 커버리지 확인
  const coverageResult = execSync('npm run test:coverage', { encoding: 'utf8' });
  expect(coverageResult).toMatch(/All files\s+\|\s+\d{2,3}(\.\d+)?\s+\|/);
  
  // 75% 이상 커버리지 확인
  const coverageMatch = coverageResult.match(/All files\s+\|\s+(\d+(\.\d+)?)\s+\|/);
  const coverage = parseFloat(coverageMatch[1]);
  expect(coverage).toBeGreaterThanOrEqual(75);
});
```

#### 시나리오 4.2: 유닛 테스트 검증
```gherkin
Given SystemDetector 클래스가 구현되어 있음
When 유닛 테스트를 실행함
Then 각 메소드가 독립적으로 테스트됨
And 모킹을 사용한 격리된 테스트가 가능
그리고 테스트 간 의존성이 없음
```

**구현 검증**:
```typescript
describe('SystemDetector Unit Tests', () => {
  let detector: SystemDetector;
  
  beforeEach(() => {
    detector = new SystemDetector();
    jest.clearAllMocks();
  });

  test('should extract Node.js version correctly', () => {
    const output = 'v18.17.0\n';
    const version = detector['extractVersion'](output, 'Node.js');
    
    expect(version).toBe('18.17.0');
  });

  test('should extract Git version correctly', () => {
    const output = 'git version 2.40.0\n';
    const version = detector['extractVersion'](output, 'Git');
    
    expect(version).toBe('2.40.0');
  });

  test('should handle malformed version output', () => {
    const output = 'some unexpected output';
    const version = detector['extractVersion'](output, 'Node.js');
    
    expect(version).toBeNull();
  });
});
```

### AC-012-05: 크로스 플랫폼 지원

#### 시나리오 5.1: 플랫폼 별 빌드 성공
```gherkin
Given TypeScript 코드가 작성되어 있음
When Windows, macOS, Linux 환경에서 각각 빌드를 실행함
Then 모든 플랫폼에서 빌드가 성공함
And 동일한 결과물이 생성됨
And 플랫폼별 특수 처리가 정상 동작함
```

**구현 검증** (테스트 매트릭스 예시):
```yaml
# .github/workflows/test.yml
name: Cross-platform Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        node-version: [18, 20]
    
    steps:
    - uses: actions/checkout@v3
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: ${{ matrix.node-version }}
    - run: npm ci
    - run: npm run build
    - run: npm test
    - run: npm run test:coverage
```

---

## 성능 벤치마크

### 성능 테스트 시나리오

#### 벤치마크 1: CLI 시작 시간
```gherkin
Given MoAI-ADK CLI가 빌드되어 있음
When 'moai --version' 명령어를 10번 연속 실행함
Then 평균 실행 시간이 2초 이내임
And 최대 실행 시간이 3초 이내임
And 메모리 사용량이 100MB 이내임
```

**구현 검증**:
```typescript
test('CLI startup performance benchmark', async () => {
  const iterations = 10;
  const executionTimes: number[] = [];
  
  for (let i = 0; i < iterations; i++) {
    const startTime = Date.now();
    execSync('node dist/cli/index.js --version', { stdio: 'pipe' });
    const executionTime = Date.now() - startTime;
    executionTimes.push(executionTime);
  }
  
  const averageTime = executionTimes.reduce((a, b) => a + b) / iterations;
  const maxTime = Math.max(...executionTimes);
  
  expect(averageTime).toBeLessThan(2000);
  expect(maxTime).toBeLessThan(3000);
});
```

#### 벤치마크 2: 시스템 검사 성능
```gherkin
Given 여러 시스템 도구가 설치되어 있음
When SystemDetector가 전체 시스템 검사를 수행함
Then 검사 시간이 5초 이내임
And 병렬 처리로 인한 성능 개선이 확인됨
And 메모리 누수가 없음
```

**구현 검증**:
```typescript
test('system check performance benchmark', async () => {
  const detector = new SystemDetector();
  const requirements = SYSTEM_REQUIREMENTS;
  
  const startTime = Date.now();
  const startMemory = process.memoryUsage().heapUsed;
  
  // 병렬 처리로 전체 시스템 검사
  const results = await Promise.all(
    requirements.map(req => detector.checkRequirement(req))
  );
  
  const endTime = Date.now();
  const endMemory = process.memoryUsage().heapUsed;
  const memoryDelta = endMemory - startMemory;
  
  expect(endTime - startTime).toBeLessThan(5000);
  expect(results.length).toBe(requirements.length);
  expect(memoryDelta).toBeLessThan(50 * 1024 * 1024); // 50MB 이내
});
```

---

## 리그레션 테스트

### 리그레션 방지 시나리오

#### 리그레션 1: 기존 기능 보전
```gherkin
Given 이전 버전에서 정상 동작하던 기능들
When 새로운 TypeScript 버전을 빌드함
Then 기존 모든 기능이 동일하게 동작함
And API 인터페이스가 변경되지 않음
And 성능이 저하되지 않음
```

#### 리그레션 2: 에러 처리 안정성
```gherkin
Given 잘못된 입력이 제공됨
When 시스템이 에러를 처리함
Then 애플리케이션이 비정상 종료되지 않음
And 적절한 에러 메시지가 표시됨
And 로그가 올바르게 기록됨
```

---

## Definition of Done (DoD) 체크리스트

### 기능 완성 체크리스트

#### 시스템 요구사항 자동 검증 모듈
- [ ] SystemDetector 클래스 구현 완료
- [ ] Node.js, Git, SQLite3 감지 기능 동작
- [ ] 버전 비교 로직 정상 동작
- [ ] 에러 처리 시나리오 커버
- [ ] 관련 유닛 테스트 모두 통과
- [ ] TypeScript 타입 안전성 확보

#### CLI 기본 명령어
- [ ] `moai --version` 명령어 동작
- [ ] `moai --help` 명령어 동작
- [ ] `moai doctor` 명령어 동작
- [ ] `moai init` 기본 구조 동작
- [ ] E2E 테스트 모두 통과
- [ ] CLI 성능 벤치마크 통과

#### 빌드 시스템
- [ ] TypeScript 컴파일 성공
- [ ] tsup 빌드 구성 완료
- [ ] ESM/CJS 듀얼 지원 확인
- [ ] 타입 정의 파일 생성
- [ ] 소스맵 생성 확인
- [ ] 빌드 성능 미치 30초 이내

### 품질 체크리스트

#### 테스트 인프라
- [ ] Jest 테스트 수트 실행 성공
- [ ] 커버리지 75% 이상 달성
- [ ] 유닛 테스트 간 독립성 보장
- [ ] 테스트 실행 시간 60초 이내
- [ ] 리그레션 테스트 통과

#### 코드 품질
- [ ] ESLint 에러 0개
- [ ] Prettier 포맷팅 규칙 준수
- [ ] TypeScript strict 모드 통과
- [ ] 타입 커버리지 100%
- [ ] 하드코딩된 값 없이 설정 대상화

#### 문서화
- [ ] README.md 기본 구조 완료
- [ ] API 대상 인터페이스 문서화
- [ ] 사용자 설치 가이드 작성
- [ ] 테스트 실행 방법 안내
- [ ] 랜전 메시지 및 오류 코드 정리

### 성능 체크리스트

#### 성능 벤치마크
- [ ] CLI 시작 시간 < 2초 (평균)
- [ ] CLI 시작 시간 < 3초 (최대)
- [ ] 시스템 검사 시간 < 5초
- [ ] 메모리 사용량 < 100MB
- [ ] 빌드 시간 < 30초
- [ ] 테스트 실행 시간 < 60초

#### 리소스 사용량
- [ ] npm 패키지 크기 < 10MB
- [ ] 런타임 메모리 사용량 측정
- [ ] CPU 사용량 모니터링
- [ ] 디스크 I/O 병목 현상 없음

### 호환성 체크리스트

#### 크로스 플랫폼
- [ ] Windows 10+ 호환성 테스트
- [ ] macOS 11+ 호환성 테스트
- [ ] Ubuntu 20.04+ 호환성 테스트
- [ ] Node.js 18.x, 20.x 호환성
- [ ] CI/CD 파이프라인 매트릭스 테스트

#### 상호 운용성
- [ ] Claude Code 환경에서 정상 동작
- [ ] npm 전역 설치 가능
- [ ] npx로 임시 실행 가능
- [ ] 기존 Node.js 프로젝트와 충돌 없음

---

**최종 수락 조건**: 위의 모든 체크리스트 항목이 100% 완료되고, Week 1 마지막 날 전체 통합 테스트가 성공적으로 통과해야 함.