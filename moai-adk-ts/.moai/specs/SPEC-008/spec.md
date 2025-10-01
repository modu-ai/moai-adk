# SPEC-008: config-manager.ts 리팩토링

@SPEC:REFACTOR-008 | Chain: @SPEC:REFACTOR-008 -> @SPEC:REFACTOR-008 -> @CODE:REFACTOR-008

## 목표
config-manager.ts (524 LOC)를 단일 책임 원칙에 따라 리팩토링하여 300 LOC 이하의 여러 모듈로 분리

## 현재 문제점 (EARS: Constraints)
- IF 파일 크기가 524 LOC이면, 단일 책임 원칙을 위반하고 있다
- 6개의 public API + 9개의 private 헬퍼가 하나의 클래스에 혼재
- 설정 생성, 파일 작업, 검증 로직이 모두 ConfigManager에 결합됨

## 요구사항 (EARS)

### Ubiquitous Requirements
- 시스템은 100% API 호환성을 유지해야 한다
- 시스템은 각 파일이 300 LOC 이하가 되도록 분리해야 한다

### Event-driven Requirements
- WHEN 설정 파일 생성 요청이 오면, 적절한 빌더를 선택하여 실행해야 한다
- WHEN 백업이 필요하면, 파일 유틸리티가 백업을 처리해야 한다

### Constraints
- 모든 기존 테스트는 수정 없이 통과해야 한다
- 공개 API 시그니처는 변경하지 않아야 한다

## 설계 방안 (@SPEC:REFACTOR-008)

### 1. 모듈 분리 구조
```
config/
├── config-manager.ts (200 LOC) - 메인 파사드
├── builders/
│   ├── claude-settings-builder.ts (100 LOC)
│   ├── moai-config-builder.ts (100 LOC)
│   └── package-json-builder.ts (100 LOC)
└── utils/
    ├── config-file-utils.ts (80 LOC)
    └── config-helpers.ts (80 LOC)
```

### 2. 책임 분리
- **ConfigManager**: 파사드 역할, 빌더 조율
- **SettingsBuilder**: Claude/MoAI/Package 설정 생성
- **ConfigFileUtils**: 파일 입출력, 백업, 검증
- **ConfigHelpers**: 모드별 설정 값 제공

### 3. TDD 접근
- RED: 기존 테스트 실행 (현재 통과 상태 확인)
- GREEN: 리팩토링 후 동일 테스트 통과
- REFACTOR: 중복 코드 제거 및 타입 안전성 강화

## 성공 기준
- [ ] 모든 파일 ≤ 300 LOC
- [ ] 기존 테스트 100% 통과
- [ ] 공개 API 시그니처 유지
- [ ] 새로운 유닛 테스트 추가 (빌더/유틸 각각)
- [ ] 순환 의존성 없음
