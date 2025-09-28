# SPEC-012 수락 기준: Python → TypeScript 완전 포팅 (5주)

> **@TEST:COMPLETE-MIGRATION-ACCEPTANCE-012** Python 완전 제거 + TypeScript 단독 실행 검증 기준
> **@FEATURE:FULL-PORTING-SCENARIOS-012** 완전 전환에 대한 Given-When-Then 시나리오
> **@QUALITY:MIGRATION-DOD-012** 5주 완전 포팅 완료를 위한 정량적 기준

---

## 완전 포팅 수락 기준 개요

### 포팅 완성도 요구사항 (100% 필수)

| 요구사항 ID | 설명 | 포팅 기준 | 수락 기준 |
|------------|------|----------|----------|
| **AC-012-01** | Python 코드 완전 제거 | 0% 잔존 | 모든 .py 파일 삭제 완료 |
| **AC-012-02** | TypeScript 기능 동등성 | 100% 포팅 | 모든 Python 기능 동일 구현 |
| **AC-012-03** | CLI 명령어 완전 전환 | pip → npm | `npm install -g moai-adk` 단독 설치 |
| **AC-012-04** | 훅 시스템 완전 대체 | 7개 훅 전환 | Python 훅 0개, TypeScript 훅 7개 |
| **AC-012-05** | TAG 시스템 포팅 | SQLite → better-sqlite3 | 데이터 마이그레이션 100% |

### 성능 개선 요구사항 (정량적 목표)

| 지표 | Python 기준 | TypeScript 목표 | 측정 방법 |
|------|-------------|----------------|----------|
| **스캔 성능** | 1.1초 | ≤ 0.8초 (27% 개선) | TAG 스캔 벤치마크 |
| **메모리 효율** | 174MB | ≤ 80MB (54% 절약) | Node.js 프로세스 모니터링 |
| **설치 시간** | 30-60초 | ≤ 30초 | `time npm install -g moai-adk` |
| **패키지 크기** | 15MB | ≤ 10MB | npm 패키지 분석 |
| **설치 성공률** | 95% | ≥ 98% | 다양한 환경 테스트 |

---

## Given-When-Then 완전 포팅 시나리오

### AC-012-01: Python 코드 완전 제거 검증

#### 시나리오 1.1: Python 파일 제거 확인
```gherkin
Given 기존 Python MoAI-ADK 프로젝트가 존재함
When TypeScript 포팅이 완료됨
Then 모든 .py 파일이 삭제되었음
And pip 의존성이 완전히 제거되었음
And Python 관련 설정 파일이 모두 삭제되었음
```

**구현 검증**:
```typescript
test('should have zero Python files remaining', async () => {
  const pythonFiles = await glob('**/*.py', { ignore: ['node_modules/**', 'venv/**'] });

  expect(pythonFiles).toHaveLength(0);

  // pip 의존성 확인
  expect(fs.existsSync('requirements.txt')).toBe(false);
  expect(fs.existsSync('setup.py')).toBe(false);
  expect(fs.existsSync('pyproject.toml')).toBe(false);

  // Python 실행 명령어가 없는지 확인
  const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'));
  const scripts = Object.values(packageJson.scripts || {}).join(' ');
  expect(scripts).not.toContain('python');
  expect(scripts).not.toContain('pip');
});
```

#### 시나리오 1.2: TypeScript 단독 실행 확인
```gherkin
Given Python 환경이 없는 시스템
When npm install -g moai-adk를 실행함
Then 설치가 성공적으로 완료됨
And 모든 CLI 명령어가 정상 동작함
And Python 의존성 없이 완전히 동작함
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

## 완전 포팅 Definition of Done (DoD) 체크리스트

### 포팅 완성도 체크리스트 (100% 필수)

#### Python 코드 완전 제거
- [ ] **Python 파일 0% 잔존**: 모든 .py 파일 삭제 완료
- [ ] **pip 의존성 제거**: requirements.txt, setup.py, pyproject.toml 삭제
- [ ] **Python 명령어 제거**: package.json scripts에서 python/pip 명령어 완전 제거
- [ ] **환경 변수 정리**: Python 관련 환경 변수 모두 제거
- [ ] **CI/CD 수정**: GitHub Actions에서 Python 설치 단계 제거
- [ ] **문서 업데이트**: README.md에서 Python 설치 안내 완전 제거

#### TypeScript 기능 동등성 (70개+ 모듈)
- [ ] **CLI 모듈**: Python click → TypeScript commander.js 완전 포팅
- [ ] **Core 모듈**: Git, TAG, 설치 시스템 완전 포팅
- [ ] **Install 모듈**: 설치 오케스트레이션 완전 포팅
- [ ] **훅 시스템**: 7개 Python 훅 → TypeScript 완전 전환
- [ ] **유틸리티**: 로깅, 보안, 파일 작업 완전 포팅
- [ ] **기능 검증**: 모든 Python 기능과 100% 동일 동작

#### npm 패키지 완전 전환
- [ ] **단독 설치**: `npm install -g moai-adk`만으로 완전 설치
- [ ] **Python 의존성 0**: Python 환경 없이 완전 동작
- [ ] **패키지 최적화**: npm 패키지 크기 ≤ 10MB
- [ ] **전역 명령어**: `moai` 명령어 전역 접근 가능
- [ ] **배포 검증**: npm 레지스트리 정식 배포 완료

### 성능 목표 달성 체크리스트

#### 스캔 성능 개선 (27% 목표)
- [ ] **기준 측정**: Python 버전 TAG 스캔 1.1초 확인
- [ ] **목표 달성**: TypeScript 버전 TAG 스캔 ≤ 0.8초
- [ ] **벤치마크**: 4,686개 파일 스캔 성능 측정
- [ ] **최적화**: 비동기 I/O 및 병렬 처리 적용
- [ ] **검증**: 지속적 성능 모니터링 시스템 구축

#### 메모리 효율 개선 (54% 목표)
- [ ] **기준 측정**: Python 버전 메모리 사용량 174MB 확인
- [ ] **목표 달성**: TypeScript 버전 메모리 사용량 ≤ 80MB
- [ ] **프로파일링**: V8 힙 사용량 최적화
- [ ] **누수 방지**: 메모리 누수 감지 및 방지
- [ ] **지속 모니터링**: 장기간 실행 시 메모리 안정성 확인

#### 설치 시간 안정화
- [ ] **Python 대비**: pip 설치 30-60초 → npm 설치 ≤ 30초
- [ ] **의존성 최적화**: 불필요한 의존성 제거
- [ ] **캐시 활용**: npm 캐시 최적화
- [ ] **설치 성공률**: 95% → 98% 개선

### 품질 보장 체크리스트

#### 타입 안전성 (100% 필수)
- [ ] **TypeScript strict**: 100% strict 모드, 타입 오류 0개
- [ ] **타입 커버리지**: 모든 함수 및 변수 타입 정의
- [ ] **런타임 검증**: zod 등을 통한 입력 타입 검증
- [ ] **컴파일 타임**: 모든 잠재적 오류 사전 감지
- [ ] **IDE 지원**: 완전한 타입 힌트 및 자동완성

#### 테스트 완성도
- [ ] **기능 동등성**: Python 버전과 100% 동일 결과 테스트
- [ ] **커버리지**: ≥ 85% 테스트 커버리지 유지
- [ ] **성능 테스트**: 모든 성능 목표 달성 검증
- [ ] **호환성 테스트**: 기존 프로젝트 100% 호환
- [ ] **회귀 테스트**: 기존 기능 보전 확인

#### 크로스 플랫폼 호환성
- [ ] **Windows**: PowerShell 환경 완전 지원
- [ ] **macOS**: Zsh/Bash 환경 완전 지원
- [ ] **Linux**: 주요 배포판 완전 지원
- [ ] **Node.js**: 18.x, 20.x 버전 호환성
- [ ] **CI/CD**: 매트릭스 테스트 모든 환경 통과

### 생태계 통합 체크리스트

#### Claude Code 완전 통합
- [ ] **훅 시스템**: 7개 TypeScript 훅 Claude Code 완전 통합
- [ ] **에이전트**: 6개 에이전트 TypeScript 버전 지원
- [ ] **명령어**: 5개 워크플로우 명령어 TypeScript 구현
- [ ] **설정**: .claude/ 디렉토리 구조 100% 호환
- [ ] **권한**: Claude Code 권한 시스템 완전 호환

#### 기존 프로젝트 호환성
- [ ] **구조 호환**: .moai/ 디렉토리 구조 100% 유지
- [ ] **설정 호환**: 기존 설정 파일 완전 호환
- [ ] **데이터 마이그레이션**: TAG 데이터베이스 무손실 이전
- [ ] **워크플로우**: 기존 Git 워크플로우 완전 보존
- [ ] **사용자 경험**: Python 버전과 동일한 UX

#### 배포 및 지원
- [ ] **npm 정식 배포**: moai-adk@1.0.0 공개 릴리스
- [ ] **Python deprecation**: pip 패키지 폐기 예고 공지
- [ ] **마이그레이션 가이드**: 상세한 전환 가이드 제공
- [ ] **사용자 지원**: 기존 사용자 전환 지원 체계
- [ ] **문서 완성**: API 문서, 사용 가이드 완전 정비

---

**5주 완전 포팅 최종 수락 조건**:
- Python 코드 0% 잔존 (완전 제거)
- TypeScript 기능 100% 동등성 (모든 Python 기능 구현)
- 성능 목표 100% 달성 (0.8초, 80MB)
- 기존 프로젝트 100% 호환성 (무중단 전환)
- npm 패키지 정식 배포 완료