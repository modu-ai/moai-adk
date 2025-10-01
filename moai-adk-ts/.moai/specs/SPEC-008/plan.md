# SPEC-008 구현 계획

@CODE:REFACTOR-008 | Chain: @SPEC:REFACTOR-008 -> @CODE:REFACTOR-008

## 구현 단계

### Phase 1: 준비 (RED)
1. 기존 테스트 실행 및 통과 확인
2. 현재 API 표면 문서화
3. 의존성 그래프 분석

### Phase 2: 빌더 추출 (GREEN)
1. **claude-settings-builder.ts** 생성
   - `createClaudeSettings()` 로직 이동
   - 관련 헬퍼 메서드 이동 (getEnabledAgents 등)

2. **moai-config-builder.ts** 생성
   - `createMoAIConfig()` 로직 이동
   - MoAIConfig 구조 생성 로직

3. **package-json-builder.ts** 생성
   - `createPackageJson()` 로직 이동
   - 스크립트/의존성 생성 헬퍼 이동

### Phase 3: 유틸리티 분리 (GREEN)
1. **config-file-utils.ts** 생성
   - `backupConfigFile()` 이동
   - `validateConfigFile()` 이동
   - 파일 읽기/쓰기 공통 로직

2. **config-helpers.ts** 생성
   - 모드별 설정 헬퍼 메서드들
   - 재사용 가능한 유틸리티 함수

### Phase 4: ConfigManager 슬림화 (REFACTOR)
1. 파사드 패턴으로 재구성
2. 빌더 인스턴스 관리
3. 공개 API만 유지

### Phase 5: 테스트 보강
1. 각 빌더별 단위 테스트 작성
2. 유틸리티 함수 테스트 작성
3. 통합 테스트 검증

## 예상 산출물
- config-manager.ts (200 LOC)
- builders/ 하위 3개 파일 (각 100 LOC)
- utils/ 하위 2개 파일 (각 80 LOC)
- __tests__/ 하위 새로운 테스트 파일들

## 위험 요소
- 순환 의존성 발생 가능성 → 단방향 의존성 강제
- 타입 임포트 복잡도 증가 → types.ts 중앙 관리
