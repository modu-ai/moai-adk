# MoAI-ADK 핵심 원칙

MoAI-ADK는 다음 5가지 핵심 원칙을 기반으로 설계되었습니다:

## 1. Spec-First: 명세 없이 코드 없음

모든 개발은 반드시 명세서 작성부터 시작합니다.

### 원칙
- 구현 전 명세 작성 의무화
- EARS (Easy Approach to Requirements Syntax) 형식 준수
- [NEEDS CLARIFICATION] 마커를 통한 불명확한 요구사항 식별

### 적용
```bash
/moai:2-spec user-auth "JWT 기반 사용자 인증 시스템"
```

## 2. TDD-First: 테스트 없이 구현 없음

Red-Green-Refactor 사이클을 엄격히 준수합니다.

### 원칙
- 테스트 작성 → 실패 확인 → 구현 → 리팩토링
- 테스트 커버리지 80% 이상 의무화
- contract/ → unit/ → integration/ → e2e/ 순차 테스트

### 적용
```bash
/moai:4-tasks SPEC-001  # TDD 태스크 생성
/moai:5-dev T001        # Red-Green-Refactor 구현
```

## 3. Living Doc: 문서와 코드는 항상 동기화

코드 변경 시 문서도 자동으로 업데이트됩니다.

### 원칙
- 코드↔문서 실시간 동기화
- @TAG 시스템을 통한 추적성 보장
- PostToolUse Hook을 통한 자동 문서 업데이트

### 적용
```bash
/moai:6-sync  # 문서 동기화 실행
```

## 4. Full Traceability: 모든 요구사항은 추적 가능

14-Core @TAG 시스템으로 완전한 추적성을 제공합니다.

### 원칙
- @REQ → @DESIGN → @TASK → @TEST 체인 보장
- 모든 코드는 요구사항으로 역추적 가능
- 추적성 매트릭스 자동 생성

### 적용
- Primary Chain: @REQ → @DESIGN → @TASK → @TEST
- Steering Chain: @VISION → @STRUCT → @TECH → @STACK
- Quality Chain: @PERF → @SEC → @DEBT → @TODO

## 5. YAGNI: 필요한 것만 구현

"You Aren't Gonna Need It" - 과도한 설계를 지양합니다.

### 원칙
- 명세에 정의된 기능만 구현
- Constitution Check를 통한 복잡도 제한
- 단순성(Simplicity) 우선 사고

### 적용
```bash
/moai:3-plan SPEC-001  # Constitution Check 수행
```

## 원칙 적용을 위한 자동화 시스템

### Hook 시스템
- **PreToolUse**: 정책 준수 자동 검증
- **PostToolUse**: 문서 동기화 자동 실행
- **SessionStart**: 프로젝트 상태 알림

### 품질 게이트
- **Constitution Check**: 5원칙 준수 검증
- **추적성 검증**: @TAG 체인 완성도 확인
- **테스트 커버리지**: 80% 임계값 강제

### 자동 동기화
- 코드 변경 시 관련 문서 자동 업데이트
- @TAG 인덱스 실시간 갱신
- 추적성 매트릭스 자동 재생성

## 원칙 위반 시 대응

### 자동 차단
- 명세 없는 코드 작성 시도 차단
- 테스트 없는 구현 시도 차단
- 추적성 체인 누락 시 경고

### 자동 수정
- @TAG 링크 자동 복구
- 문서 동기화 자동 실행
- 누락된 테스트 케이스 자동 식별

이러한 원칙들은 단순한 가이드라인이 아닌, **자동화된 시스템을 통해 강제되는 개발 철학**입니다.