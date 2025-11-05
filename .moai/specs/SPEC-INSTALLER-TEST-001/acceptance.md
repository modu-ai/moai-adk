# @SPEC:INSTALLER-TEST-001: Acceptance Criteria

## AC-001: 전체 커버리지 85% 달성

**GIVEN** Installer 패키지의 12개 미테스트 파일이 존재할 때
**WHEN** 전체 테스트를 실행하고 커버리지를 측정하면
**THEN**
- 전체 라인 커버리지가 85% 이상이어야 한다
- 브랜치 커버리지가 80% 이상이어야 한다
- 함수 커버리지가 90% 이상이어야 한다

## AC-002: 각 파일별 커버리지 85% 달성

**GIVEN** 특정 파일(예: installer-core.ts)에 대한 테스트가 작성되었을 때
**WHEN** 해당 파일의 커버리지를 측정하면
**THEN**
- 해당 파일의 라인 커버리지가 85% 이상이어야 한다
- 모든 public 메서드가 테스트되어야 한다
- 주요 에러 케이스가 테스트되어야 한다

## AC-003: Vitest 프레임워크 사용

**GIVEN** 새로운 테스트 파일을 작성할 때
**WHEN** 테스트 코드를 작성하면
**THEN**
- Vitest의 describe, it, expect를 사용해야 한다
- vi.mock()을 사용하여 의존성을 모킹해야 한다
- async/await 패턴을 사용하여 비동기 테스트를 작성해야 한다

## AC-004: TAG 포함

**GIVEN** 테스트 파일을 작성할 때
**WHEN** 파일 상단에 주석을 추가하면
**THEN**
- `@TEST:INSTALLER-TEST-001` TAG가 포함되어야 한다
- SPEC 파일 경로가 명시되어야 한다

## AC-005: TDD Red-Green-Refactor 사이클

**GIVEN** 새로운 기능을 구현할 때
**WHEN** TDD 사이클을 따르면
**THEN**
- Red: 실패하는 테스트를 먼저 작성해야 한다
- Green: 최소한의 코드로 테스트를 통과시켜야 한다
- Refactor: 코드를 개선하고 커버리지를 확인해야 한다

## AC-006: 모킹을 통한 격리된 테스트

**GIVEN** 의존성이 있는 클래스를 테스트할 때
**WHEN** 테스트를 작성하면
**THEN**
- vi.mock()을 사용하여 의존성을 모킹해야 한다
- 각 테스트는 독립적으로 실행되어야 한다
- 외부 의존성 없이 테스트가 통과해야 한다

## AC-007: 에러 케이스 테스트

**GIVEN** 에러 처리 로직이 있는 코드를 테스트할 때
**WHEN** 에러 케이스 테스트를 작성하면
**THEN**
- 각 InstallationError 타입에 대한 테스트가 존재해야 한다
- 에러 메시지가 정확히 검증되어야 한다
- 에러 발생 시 정리 로직이 실행되는지 확인해야 한다

## AC-008: 파일 시스템 격리 테스트

**GIVEN** 파일 시스템 작업을 수행하는 코드를 테스트할 때
**WHEN** 테스트를 실행하면
**THEN**
- 임시 디렉토리를 사용하여 테스트해야 한다
- 테스트 후 임시 파일이 정리되어야 한다
- 실제 프로젝트 파일에 영향을 주지 않아야 한다

## AC-009: CI 파이프라인 커버리지 게이트

**GIVEN** CI 파이프라인에서 테스트를 실행할 때
**WHEN** 커버리지가 85% 미만이면
**THEN**
- CI 파이프라인이 실패해야 한다
- 커버리지 리포트가 생성되어야 한다
- 어떤 파일의 커버리지가 낮은지 명시되어야 한다

## AC-010: Phase별 테스트 완료

**GIVEN** Phase 1 (Critical Files) 테스트를 완료할 때
**WHEN** 커버리지를 확인하면
**THEN**
- installer-core.ts 커버리지 ≥ 85%
- phase-executor.ts 커버리지 ≥ 85%
- update-executor.ts 커버리지 ≥ 85%

**GIVEN** Phase 2 (High Priority Files) 테스트를 완료할 때
**WHEN** 커버리지를 확인하면
**THEN**
- dependency-installer.ts 커버리지 ≥ 85%
- package-manager.ts 커버리지 ≥ 85%
- template-installer.ts 커버리지 ≥ 85%

**GIVEN** Phase 3 (Medium Priority Files) 테스트를 완료할 때
**WHEN** 커버리지를 확인하면
**THEN**
- context-manager.ts 커버리지 ≥ 85%
- fallback-builder.ts 커버리지 ≥ 85%
- typescript-setup.ts 커버리지 ≥ 85%

**GIVEN** Phase 4 (Low Priority Files) 테스트를 완료할 때
**WHEN** 커버리지를 확인하면
**THEN**
- pre-install.ts 커버리지 ≥ 85%
- post-install.ts 커버리지 ≥ 85%
- update-manager.ts 커버리지 ≥ 85%

## AC-011: 테스트 문서화

**GIVEN** 모든 테스트를 완료할 때
**WHEN** README.md를 업데이트하면
**THEN**
- 테스트 실행 방법이 문서화되어야 한다
- 커버리지 확인 방법이 문서화되어야 한다
- 주요 테스트 케이스가 설명되어야 한다
