# @SPEC:INSTALLER-QUALITY-001: Acceptance Criteria

## AC-001: 모든 클래스 생성자 기반 DI 사용

**GIVEN** Installer 패키지의 모든 클래스가 존재할 때
**WHEN** 각 클래스의 생성자를 확인하면
**THEN**
- 모든 의존성이 생성자 파라미터로 주입되어야 한다
- 클래스 내부에서 new 키워드로 의존성을 생성하지 않아야 한다
- readonly 키워드가 의존성 필드에 사용되어야 한다

## AC-002: InstallationError 계열 통일

**GIVEN** 에러가 발생할 수 있는 모든 코드가 존재할 때
**WHEN** 에러를 throw하면
**THEN**
- InstallationError 또는 하위 클래스를 사용해야 한다
- 일반 Error를 직접 throw하지 않아야 한다
- 원본 에러가 cause 파라미터로 전달되어야 한다

## AC-003: 에러 타입별 적절한 하위 클래스 사용

**GIVEN** 다양한 종류의 에러가 발생할 때
**WHEN** 각 에러를 throw하면
**THEN**
- 의존성 설치 실패: DependencyInstallationError
- 템플릿 설치 실패: TemplateInstallationError
- 설정 오류: ConfigurationError
- 검증 실패: ValidationError
- 파일 시스템 오류: FileSystemError
- 네트워크 오류: NetworkError

## AC-004: TAG 형식 통일

**GIVEN** 모든 소스 코드 파일이 존재할 때
**WHEN** 파일 상단의 TAG를 확인하면
**THEN**
- `@CODE:ID | SPEC: ... | TEST: ...` 형식이어야 한다
- SPEC 경로가 명시되어야 한다
- TEST 경로가 명시되어야 한다
- ID는 `[A-Z]+-[0-9]{3}` 패턴이어야 한다

## AC-005: TAG 자동 검증

**GIVEN** TAG 검증 스크립트가 실행될 때
**WHEN** 모든 파일을 스캔하면
**THEN**
- 형식에 맞지 않는 TAG가 0개여야 한다
- TAG가 없는 파일이 리포트되어야 한다
- 검증 결과가 출력되어야 한다

## AC-006: 매직 넘버 제거

**GIVEN** 숫자 리터럴이 코드에 사용될 때
**WHEN** 동일한 값이 2회 이상 사용되면
**THEN**
- installer-constants.ts에 명명된 상수가 정의되어야 한다
- 코드에서 해당 상수를 참조해야 한다
- 매직 넘버가 직접 사용되지 않아야 한다

## AC-007: 매직 문자열 제거

**GIVEN** 문자열 리터럴이 코드에 사용될 때
**WHEN** 동일한 값이 2회 이상 사용되면
**THEN**
- installer-constants.ts에 명명된 상수가 정의되어야 한다
- 경로, 패키지 매니저 이름 등이 상수로 정의되어야 한다
- 문자열이 직접 사용되지 않아야 한다

## AC-008: 크로스 플랫폼 명령 실행

**GIVEN** 외부 명령을 실행해야 할 때
**WHEN** Windows 또는 Unix 환경에서 실행하면
**THEN**
- cross-spawn 라이브러리를 사용해야 한다
- which 명령을 직접 사용하지 않아야 한다
- windowsHide 옵션이 설정되어야 한다
- 플랫폼별 분기 없이 동작해야 한다

## AC-009: Windows 환경 지원

**GIVEN** Windows 환경에서 Installer를 실행할 때
**WHEN** 패키지 매니저 감지 및 설치를 수행하면
**THEN**
- npm, pnpm, yarn이 정확히 감지되어야 한다
- 설치가 정상적으로 완료되어야 한다
- 경로 구분자(\)가 올바르게 처리되어야 한다

## AC-010: InstallerFactory 사용

**GIVEN** InstallerCore를 생성해야 할 때
**WHEN** InstallerFactory.create()를 호출하면
**THEN**
- 모든 의존성이 올바른 순서로 주입되어야 한다
- 순환 의존성이 없어야 한다
- 생성된 인스턴스가 즉시 사용 가능해야 한다

## AC-011: 상수 문서화

**GIVEN** installer-constants.ts 파일이 존재할 때
**WHEN** 각 상수를 확인하면
**THEN**
- JSDoc 주석이 포함되어야 한다
- 단위가 명시되어야 한다 (밀리초, 횟수 등)
- 용도가 명확히 설명되어야 한다

## AC-012: 에러 메시지 일관성

**GIVEN** 에러가 발생했을 때
**WHEN** 에러 메시지를 확인하면
**THEN**
- 동일한 형식을 사용해야 한다 (예: "Failed to ...")
- 컨텍스트 정보가 포함되어야 한다
- 해결 방법 힌트가 포함되어야 한다 (선택 사항)

## AC-013: DI 테스트 가능성

**GIVEN** DI가 적용된 클래스를 테스트할 때
**WHEN** 모킹된 의존성을 주입하면
**THEN**
- 테스트가 독립적으로 실행되어야 한다
- 외부 의존성 없이 테스트되어야 한다
- 각 의존성을 쉽게 모킹할 수 있어야 한다

## AC-014: 마이그레이션 가이드 제공

**GIVEN** 기존 코드를 새로운 패턴으로 마이그레이션할 때
**WHEN** MIGRATION.md를 참조하면
**THEN**
- Before/After 예제가 제공되어야 한다
- 단계별 마이그레이션 가이드가 있어야 한다
- Breaking Changes가 명시되어야 한다

## AC-015: 상수 타입 안전성

**GIVEN** installer-constants.ts의 상수를 사용할 때
**WHEN** TypeScript 컴파일을 수행하면
**THEN**
- as const 단언이 사용되어야 한다
- readonly 타입이어야 한다
- 타입 추론이 정확해야 한다

## AC-016: 에러 스택 트레이스 보존

**GIVEN** 에러가 체이닝되었을 때
**WHEN** 에러 스택을 확인하면
**THEN**
- 원본 에러의 스택이 보존되어야 한다
- cause 체인이 추적 가능해야 한다
- 디버깅이 용이해야 한다

## AC-017: Factory 확장성

**GIVEN** 새로운 의존성을 추가해야 할 때
**WHEN** InstallerFactory를 수정하면
**THEN**
- 한 곳에서만 수정이 필요해야 한다
- 기존 코드 변경이 최소화되어야 한다
- 타입 안전성이 유지되어야 한다

## AC-018: 통합 테스트 통과

**GIVEN** 모든 품질 개선이 완료되었을 때
**WHEN** 전체 테스트 스위트를 실행하면
**THEN**
- 모든 단위 테스트가 통과해야 한다
- 통합 테스트가 통과해야 한다
- 테스트 커버리지 ≥ 85%여야 한다

## AC-019: 코드 리뷰 체크리스트

**GIVEN** 코드 리뷰를 수행할 때
**WHEN** 체크리스트를 확인하면
**THEN**
- [ ] DI 패턴 일관성
- [ ] 에러 처리 통일성
- [ ] TAG 형식 준수
- [ ] 매직 넘버/문자열 제거
- [ ] 크로스 플랫폼 지원
- [ ] 테스트 작성
- [ ] 문서화 완료

## AC-020: 성능 저하 없음

**GIVEN** 품질 개선 전후를 비교할 때
**WHEN** 벤치마크를 실행하면
**THEN**
- 설치 시간이 5% 이상 증가하지 않아야 한다
- 메모리 사용량이 10% 이상 증가하지 않아야 한다
- DI 오버헤드가 무시 가능해야 한다
