# SPEC-006: Utils 모듈 테스트 작성

## @TAG BLOCK
```
# @CODE:TEST-UTILS-001 | Chain: @SPEC:QUAL-006 -> @SPEC:TEST-006 -> @CODE:TEST-006 -> @TEST:UTIL-006
# Related: @CODE:VALID-002:API, @CODE:INPUT-001, LOG-002
```

## 개요

**목적**: src/utils/ 모듈의 핵심 유틸리티 함수에 대한 TDD 기반 테스트 작성
**범위**: 우선순위 높은 3개 파일 (input-validator.ts, errors.ts, winston-logger.ts)
**현재 상태**: 8개 파일 중 1개만 테스트 존재 (12.5% 커버리지)
**목표**: 최소 50% 커버리지 달성, 핵심 유틸리티 신뢰성 보장

## Environment (환경 및 가정사항)

### 현재 환경
- **테스트 프레임워크**: Vitest (TypeScript 기반 TDD)
- **프로젝트 구조**: `src/utils/` 디렉토리 내 8개 유틸리티 파일
- **기존 테스트**: `src/__tests__/utils/path-validator.test.ts` (1개만 존재)
- **테스트 위치**: `src/__tests__/utils/` 또는 소스 파일과 동일 디렉토리

### 기술 스택
- **언어**: TypeScript (strict mode)
- **런타임**: Node.js
- **테스트 도구**: Vitest, @vitest/coverage-v8
- **모킹**: Vitest의 vi.mock(), vi.fn()

### 가정사항
- Vitest 설정이 완료되어 있음
- 소스 코드는 변경하지 않음 (테스트만 작성)
- 테스트는 독립적이고 격리된 환경에서 실행됨
- 파일 시스템 테스트는 임시 디렉토리 사용

## Assumptions (전제 조건)

### 우선순위 선정 근거
1. **input-validator.ts** (458 LOC)
   - 보안에 중요 (@CODE:INPUT-VALIDATION-001)
   - 가장 복잡한 로직 (Path traversal, 위험 패턴 검증)
   - 사용자 입력 검증의 핵심

2. **errors.ts** (151 LOC)
   - 에러 처리의 기반
   - 타입 안전성 보장 (@CODE:TYPE-SAFETY-001)
   - 모든 모듈에서 사용되는 공통 유틸리티

3. **winston-logger.ts** (349 LOC)
   - 로깅 시스템의 핵심 (@CODE:LOG-002)
   - 민감 정보 마스킹 로직 검증 필요
   - 파일 시스템 의존성 테스트

### 나머지 파일은 차후 SPEC으로 분리
- **banner.ts**: 단순 출력 함수 (낮은 우선순위)
- **i18n.ts**: 국제화 지원 (선택적 기능)
- **package-root.ts**: 패키지 경로 탐색
- **version.ts**: 버전 정보 출력
- **path-validator.ts**: 이미 테스트 존재 (검증 완료)

## Requirements (기능 요구사항)

### EARS 방식 요구사항 정의

#### Ubiquitous Requirements (기본 요구사항)
- 시스템은 input-validator.ts의 모든 public 메서드에 대한 테스트를 제공해야 한다
- 시스템은 errors.ts의 모든 커스텀 에러 클래스와 헬퍼 함수에 대한 테스트를 제공해야 한다
- 시스템은 winston-logger.ts의 핵심 로깅 기능에 대한 테스트를 제공해야 한다
- 시스템은 Vitest 프레임워크를 사용해야 한다
- 시스템은 최소 85% 테스트 커버리지를 달성해야 한다

#### Event-driven Requirements (이벤트 기반)
- WHEN 잘못된 프로젝트명이 입력되면, validateProjectName이 적절한 ValidationResult를 반환하는지 테스트해야 한다
- WHEN 위험한 경로 패턴이 입력되면, validatePath가 에러를 반환하는지 테스트해야 한다
- WHEN 커스텀 에러가 throw되면, instanceof 체크가 올바르게 동작하는지 테스트해야 한다
- WHEN 민감 정보가 포함된 로그가 기록되면, 마스킹이 올바르게 적용되는지 테스트해야 한다
- WHEN 로그 디렉토리가 존재하지 않으면, winston-logger가 자동으로 생성하는지 테스트해야 한다

#### State-driven Requirements (상태 기반)
- WHILE verbose 모드가 활성화되어 있을 때, 로거가 debug 레벨 로그를 출력하는지 테스트해야 한다
- WHILE 파일 로깅이 비활성화되어 있을 때, 콘솔만 사용하는지 테스트해야 한다

#### Optional Features (선택적 기능)
- WHERE 파일 시스템 접근이 필요한 경우, 임시 디렉토리를 사용할 수 있다
- WHERE 실제 파일 생성이 부담되는 경우, mock을 사용할 수 있다

#### Constraints (제약사항)
- IF 테스트가 실패하면, 빌드를 중단해야 한다
- 각 테스트 파일은 300 LOC를 초과하지 않아야 한다
- 테스트는 독립적이고 순서에 무관하게 실행되어야 한다
- 테스트는 외부 의존성 없이 실행 가능해야 한다
- 로거 테스트는 실제 파일을 남기지 않아야 한다 (cleanup 필수)

## Specifications (상세 명세)

### 1. input-validator.test.ts

#### 테스트 대상 메서드
1. **validateProjectName()**
   - 유효한 프로젝트명 검증
   - 길이 제한 검증 (minLength, maxLength)
   - 공백 검증 (allowSpaces 옵션)
   - 특수문자 검증 (allowSpecialChars 옵션)
   - 위험 패턴 검증 (path traversal, control chars, Windows 예약어)
   - sanitizedValue 생성 검증

2. **validatePath()**
   - 경로 길이 검증 (MAX_PATH 260자)
   - 위험 패턴 검증 (시스템 디렉토리, 실행 파일)
   - 경로 깊이 검증 (maxDepth)
   - 파일 존재 검증 (mustExist, mustBeDirectory, mustBeFile)
   - 확장자 검증 (allowedExtensions)
   - Path traversal 방지 검증

3. **validateTemplateType()**
   - 허용된 템플릿 타입 검증 (standard, minimal, advanced, custom)
   - 대소문자 무관 검증
   - 빈 문자열 처리

4. **validateBranchName()**
   - Git 브랜치 명명 규칙 검증
   - 금지된 문자 검증 (control chars, reserved chars)
   - 예약어 검증 (HEAD, master, origin)
   - 길이 제한 (1-250자)

5. **validateCommandOptions()**
   - 옵션 키 형식 검증
   - 타입별 값 검증 (string, boolean, number)
   - 위험 문자열 검증 (script injection)
   - sanitizedValue 생성

#### 테스트 커버리지 목표
- Statement Coverage: 90% 이상
- Branch Coverage: 85% 이상
- 모든 public 메서드: 100%
- 모든 에러 케이스: 100%

### 2. errors.test.ts

#### 테스트 대상
1. **커스텀 에러 클래스**
   - ValidationError 생성 및 속성 검증
   - InstallationError 생성 및 속성 검증
   - TemplateError 생성 및 속성 검증
   - ResourceError 생성 및 속성 검증
   - PhaseError 생성 및 속성 검증

2. **타입 가드 함수**
   - isValidationError() 정확성
   - isInstallationError() 정확성
   - isTemplateError() 정확성
   - isResourceError() 정확성
   - isPhaseError() 정확성
   - 다른 타입 에러 구분 검증

3. **헬퍼 함수**
   - toError() - Error 객체 변환
   - toError() - 문자열 변환
   - toError() - unknown 타입 처리
   - getErrorMessage() - Error 메시지 추출
   - getErrorMessage() - 문자열 처리

#### 테스트 커버리지 목표
- Statement Coverage: 100%
- Branch Coverage: 100%
- 모든 에러 클래스: 100%
- 모든 타입 가드: 100%

### 3. winston-logger.test.ts

#### 테스트 대상
1. **로거 초기화**
   - 기본 옵션 설정
   - 커스텀 옵션 설정
   - 콘솔 전용 모드
   - 파일 로깅 모드
   - 로그 디렉토리 자동 생성

2. **로그 레벨 메서드**
   - debug() 메서드
   - info() 메서드
   - warn() 메서드
   - error() 메서드
   - verbose() 메서드 (verbose 모드 전용)

3. **민감 정보 마스킹**
   - password, token, apiKey 등 필드 마스킹
   - 메시지 문자열 내 패턴 마스킹
   - 중첩 객체 마스킹
   - 배열 처리

4. **Verbose 모드**
   - setVerbose(true) 동작
   - setVerbose(false) 동작
   - isVerbose() 상태 확인
   - verbose 전용 메시지 필터링

5. **TAG 추적성**
   - logWithTag() 메타데이터 포함
   - TAG 정보 보존

#### 테스트 전략
- **Mock 사용**: Winston transport를 mock으로 대체
- **파일 시스템**: 임시 디렉토리 사용 (beforeEach/afterEach cleanup)
- **비동기 처리**: async/await 패턴
- **의존성 주입**: LoggerOptions로 mock transport 주입

#### 테스트 커버리지 목표
- Statement Coverage: 85% 이상
- Branch Coverage: 80% 이상
- 민감 정보 마스킹: 100%
- 핵심 로깅 메서드: 100%

## Traceability (추적성)

### @TAG Chain

```
@SPEC:QUAL-006 (품질 요구사항: Utils 테스트)
  ↓
@SPEC:TEST-006 (설계: TDD 기반 3개 파일 테스트)
  ↓
@CODE:TEST-006 (작업: input-validator, errors, winston-logger 테스트)
  ↓
@TEST:UTIL-006 (검증: 85% 커버리지 달성)
```

### Related TAGs
- `@CODE:TEST-UTILS-001`: Utils 모듈 테스트 기능
- `@CODE:VALID-002:API`: Input validation API 테스트
- `@CODE:INPUT-001`: 보안 입력 검증 테스트
- `LOG-002`: Winston 로거 테스트

### 기존 TAG 재사용
- `@CODE:UTIL-005`: input-validator 소스 코드 TAG
- `@CODE:TYPE-SAFETY-001`: errors 소스 코드 TAG
- `@CODE:LOG-002`: winston-logger 소스 코드 TAG

## 성공 기준

### 테스트 커버리지
- **전체 Utils 모듈**: 50% 이상 (현재 12.5% → 목표 50%+)
- **input-validator.ts**: 90% 이상
- **errors.ts**: 100%
- **winston-logger.ts**: 85% 이상

### 품질 게이트
- 모든 테스트 통과
- 빌드 성공
- ESLint/Biome 검사 통과
- TypeScript strict mode 컴파일 성공

### 기능 검증
- 모든 public API 테스트 존재
- 에러 케이스 100% 커버
- 보안 검증 로직 100% 커버
- 민감 정보 마스킹 100% 커버

### 문서화
- 각 테스트에 Given-When-Then 주석
- @TAG 체인 명시
- describe/test 이름이 테스트 의도를 명확히 표현

## 제약 조건

### 코드 품질
- 테스트 파일당 300 LOC 이하
- 테스트 함수당 50 LOC 이하
- 복잡도 10 이하
- 매개변수 5개 이하

### TDD 원칙
- Red-Green-Refactor 사이클 준수
- 테스트 우선 작성 (필요시 소스 코드 리뷰 후)
- 독립적이고 격리된 테스트
- 결정적 테스트 (매번 동일한 결과)

### 성능
- 전체 테스트 실행 시간 < 10초
- 개별 테스트 < 100ms
- 파일 시스템 cleanup 완료

### 보안
- 실제 비밀 정보 사용 금지
- 임시 파일 정리 보장
- 테스트 격리 보장

## 다음 단계

이 SPEC이 완료되면:
1. `/alfred:2-build SPEC-006` 실행하여 TDD 구현
2. 테스트 커버리지 확인 및 리포트 생성
3. `/alfred:3-sync` 실행하여 문서 동기화
4. 나머지 utils 파일 (banner, i18n 등)은 SPEC-007로 분리

---

**작성일**: 2025-10-01
**버전**: 1.0.0
**상태**: Draft
