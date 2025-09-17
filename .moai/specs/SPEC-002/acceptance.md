# SPEC-002: 수락 기준 (Acceptance Criteria)

> **@REQ:CODE-TAG-002** 연결된 Given-When-Then 형식 테스트 가능 기준

## 📋 개요

**SPEC**: SPEC-002 (코드 TAG 관리 시스템 구축)
**총 시나리오**: 45개
**총 사용자 스토리**: 12개
**테스트 커버리지**: 각 스토리당 평균 3.75개 시나리오

## 🏷️ TAG 자동 생성 및 관리

### US-001: 새 파일 생성 시 자동 TAG 적용

**Scenario 1-1: 새로운 API 모듈 파일 생성**
```gherkin
Given 개발자가 src/moai_adk/api/ 디렉토리에 있고
  And VS Code나 다른 에디터로 새 Python 파일을 생성할 때
When "user_auth.py" 파일을 생성하면
Then 시스템은 자동으로 파일 상단에 "@API:POST-AUTH" TAG를 추가하고
  And TAG 인덱스에 새 항목을 등록하며
  And 개발자에게 "자동 TAG 추가됨" 알림을 표시한다
```

**Scenario 1-2: 핵심 엔진 모듈 파일 생성**
```gherkin
Given 개발자가 src/moai_adk/core/ 디렉토리에 있고
  And 새로운 핵심 기능을 구현하려 할 때
When "orchestrator.py" 파일을 생성하면
Then 시스템은 자동으로 "@FEATURE:ORCHESTRATOR" TAG를 할당하고
  And 관련된 Constitution 원칙 체크리스트를 표시하며
  And .moai/indexes/tags.json에 실시간으로 반영한다
```

**Scenario 1-3: 잘못된 위치의 파일 생성**
```gherkin
Given 개발자가 정의되지 않은 경로에 파일을 생성할 때
When src/moai_adk/unknown/ 디렉토리에 "test.py"를 생성하면
Then 시스템은 기본 "@FEATURE:UNKNOWN" TAG를 제안하고
  And 적절한 디렉토리 이동을 권장하며
  And 수동 TAG 검토가 필요하다는 경고를 표시한다
```

---

### US-002: 기존 파일 TAG 자동 감지 및 제안

**Scenario 2-1: TAG가 없는 기존 파일 감지**
```gherkin
Given src/moai_adk/cli/command.py 파일이 존재하고
  And 해당 파일에 @TAG 주석이 없을 때
When tag-scanner.py 스크립트가 실행되면
Then 시스템은 파일 내용을 분석하여 "@FEATURE:CLI-COMMAND" TAG를 제안하고
  And 제안 근거를 상세히 설명하며
  And 개발자 승인 후 자동으로 파일 상단에 추가한다
```

**Scenario 2-2: 여러 기능이 혼재된 파일 처리**
```gherkin
Given 하나의 파일에 API와 데이터 모델이 함께 있고
  And 파일 크기가 300 LOC를 초과할 때
When TAG 자동 감지가 실행되면
Then 시스템은 파일 분할을 권장하고
  And 각 기능별로 "@API:GET-MIXED", "@DATA:MIXED-MODEL" 태그를 제안하며
  And Constitution 단순성 원칙 위반 경고를 표시한다
```

**Scenario 2-3: 레거시 코드 일괄 분석**
```gherkin
Given src/moai_adk/ 전체 디렉토리에 50개의 Python 파일이 있고
  And 그 중 30개 파일에 TAG가 누락되어 있을 때
When "/moai:6-sync force" 명령을 실행하면
Then 시스템은 5초 이내에 모든 파일을 스캔하고
  And 누락된 TAG 30개에 대한 제안을 생성하며
  And 95% 이상의 정확도로 적절한 TAG를 분류한다
```

**Scenario 2-4: TAG 제안 품질 검증**
```gherkin
Given 자동 TAG 제안 시스템이 활성화되어 있고
  And 테스트용 코드 샘플 100개가 준비되어 있을 때
When 자동 감지를 실행하면
Then 제안된 TAG의 90% 이상이 수동 검토 기준과 일치하고
  And 잘못된 제안은 5% 이하로 유지되며
  And 제안 신뢰도 점수가 80% 이상인 경우에만 자동 적용한다
```

---

### US-003: TAG 형식 자동 검증 및 수정

**Scenario 3-1: 잘못된 TAG 형식 감지**
```gherkin
Given Python 파일에 "@api-user-login" 형식의 TAG가 있고
  And 이것이 16-Core 명명 규칙에 어긋날 때
When pre-commit hook이 실행되면
Then 시스템은 "잘못된 TAG 형식" 오류를 발생시키고
  And "@API:POST-LOGIN" 형식으로 수정을 제안하며
  And commit을 차단하여 일관성을 보장한다
```

**Scenario 3-2: 중복 TAG 식별자 처리**
```gherkin
Given 두 개의 파일에 동일한 "@API:GET-USERS" TAG가 있을 때
When TAG 검증 시스템이 실행되면
Then 시스템은 중복을 감지하고
  And 각각 "@API:GET-USERS", "@API:GET-USER-LIST"로 고유 식별자를 부여하며
  And 충돌 해결 로그를 .moai/logs/tag-conflicts.log에 기록한다
```

**Scenario 3-3: 대소문자 및 특수문자 정규화**
```gherkin
Given "@core:Main_Engine" 같은 비표준 형식의 TAG가 있을 때
When TAG 형식 검증이 실행되면
Then 시스템은 "@FEATURE:MAIN-ENGINE" 표준 형식으로 변환하고
  And 변경 사항을 개발자에게 알리며
  And 자동 수정 여부를 확인 후 적용한다
```

---

## 🔍 TAG 인덱싱 및 검색

### US-004: 실시간 TAG 인덱스 업데이트

**Scenario 4-1: 파일 수정 시 즉시 인덱스 업데이트**
```gherkin
Given src/moai_adk/core/engine.py 파일이 열려 있고
  And 파일 감시 시스템이 활성화되어 있을 때
When 개발자가 "@FEATURE:ENGINE" TAG를 "@FEATURE:ORCHESTRATOR"로 수정하면
Then 100ms 이내에 .moai/indexes/tags.json이 업데이트되고
  And 인덱스 변경 이벤트가 발생하며
  And 관련 추적성 체인이 자동으로 갱신된다
```

**Scenario 4-2: 파일 삭제 시 인덱스 정리**
```gherkin
Given @API:DELETE-USER TAG를 포함한 파일이 있고
  And 해당 파일이 Git에서 삭제될 때
When git commit이 발생하면
Then 시스템은 삭제된 TAG를 인덱스에서 제거하고
  And 관련된 추적성 링크를 정리하며
  And 의존하는 다른 TAG들에게 알림을 발송한다
```

**Scenario 4-3: 대량 파일 변경 시 배치 처리**
```gherkin
Given 100개 파일에 대한 일괄 TAG 변경이 발생하고
  And 시스템 리소스가 제한적일 때
When 배치 인덱싱 모드가 활성화되면
Then 시스템은 5초 이내에 모든 변경 사항을 처리하고
  And 메모리 사용량을 500MB 이하로 유지하며
  And 진행 상황을 실시간으로 표시한다
```

**Scenario 4-4: 인덱스 무결성 검증**
```gherkin
Given 인덱스 파일이 손상되었거나 불일치가 발생했을 때
When tag-validator.py --repair 명령이 실행되면
Then 시스템은 모든 소스 파일을 재스캔하고
  And 정확한 인덱스를 재구성하며
  And 복구 과정과 결과를 상세히 로깅한다
```

---

### US-005: TAG 기반 코드 검색 및 탐색

**Scenario 5-1: 특정 TAG로 파일 검색**
```gherkin
Given 프로젝트에 @API 카테고리 TAG가 15개 있고
  And 개발자가 API 관련 코드를 찾으려 할 때
When "grep -r '@API:' src/" 명령을 실행하면
Then 2초 이내에 모든 API 관련 파일이 나열되고
  And 각 파일의 기능 설명과 함께 표시되며
  And 관련된 테스트 파일도 함께 제안한다
```

**Scenario 5-2: 추적성 체인 탐색**
```gherkin
Given @REQ:USER-LOGIN-001 요구사항 TAG가 있고
  And 이와 연결된 설계, 구현, 테스트 TAG들이 존재할 때
When "python .moai/scripts/trace-chain.py REQ:USER-LOGIN-001" 명령을 실행하면
Then 요구사항부터 테스트까지의 완전한 체인이 표시되고
  And 누락된 연결고리가 있다면 경고를 표시하며
  And 그래픽 형태의 의존성 트리를 생성한다
```

**Scenario 5-3: VS Code 내 TAG 자동완성**
```gherkin
Given VS Code에서 Python 파일을 편집 중이고
  And MoAI TAG 확장이 설치되어 있을 때
When 개발자가 "@" 문자를 입력하면
Then 사용 가능한 TAG 카테고리가 자동완성으로 표시되고
  And 기존 프로젝트의 TAG 패턴을 참고하여 제안하며
  And 실시간으로 TAG 형식을 검증한다
```

---

### US-006: TAG 추적성 체인 시각화

**Scenario 6-1: 요구사항-구현 추적성 시각화**
```gherkin
Given @REQ:USER-LOGIN-001부터 @TEST:UNIT-LOGIN까지의 체인이 완성되어 있고
  And 웹 대시보드가 활성화되어 있을 때
When 프로젝트 매니저가 추적성 페이지에 접근하면
Then 요구사항부터 테스트까지의 플로우차트가 표시되고
  And 각 단계별 완성도가 퍼센트로 표시되며
  And 누락된 링크는 빨간색으로 강조 표시된다
```

**Scenario 6-2: 프로젝트 전체 TAG 맵 생성**
```gherkin
Given 프로젝트에 200개의 TAG가 분산되어 있고
  And 4개 카테고리별로 분류되어 있을 때
When "/moai:6-sync visual" 명령을 실행하면
Then 전체 TAG 구조가 마인드맵 형태로 생성되고
  And 카테고리별 색상으로 구분되며
  And .moai/docs/tag-map.svg 파일로 저장된다
```

**Scenario 6-3: 실시간 추적성 상태 모니터링**
```gherkin
Given 실시간 대시보드가 구동 중이고
  And 개발팀이 활발히 작업 중일 때
When TAG가 생성, 수정, 삭제될 때마다
Then 대시보드의 추적성 지표가 실시간으로 업데이트되고
  And 완성도가 목표 수준 이하로 떨어지면 알림을 발송하며
  And 주간 추적성 리포트를 자동으로 생성한다
```

**Scenario 6-4: 의존성 영향도 분석**
```gherkin
Given @FEATURE:ENGINE TAG를 가진 핵심 모듈이 변경될 예정이고
  And 이 모듈에 의존하는 10개의 다른 모듈이 있을 때
When 영향도 분석 도구를 실행하면
Then 영향받는 모든 모듈과 TAG가 나열되고
  And 예상 테스트 범위와 작업량이 계산되며
  And 변경 계획 수립을 위한 권장 사항을 제공한다
```

---

## 🛡️ 품질 관리 및 검증

### US-007: Git commit 시 TAG 검증

**Scenario 7-1: 새로운 코드에 TAG 누락 검증**
```gherkin
Given 개발자가 새로운 Python 파일을 추가하고
  And 해당 파일에 @TAG 주석이 없을 때
When git commit을 시도하면
Then pre-commit hook이 "TAG 누락" 오류를 발생시키고
  And 적절한 TAG 추가를 요구하며
  And commit을 차단하여 품질을 보장한다
```

**Scenario 7-2: TAG 형식 규칙 위반 감지**
```gherkin
Given 수정된 파일에 "@api-login" 같은 잘못된 형식의 TAG가 있고
  And 16-Core 명명 규칙을 위반할 때
When pre-commit 검증이 실행되면
Then 시스템은 형식 오류를 상세히 보고하고
  And "@API:POST-LOGIN" 형식으로 수정을 제안하며
  And 수정 완료 후에만 commit을 허용한다
```

**Scenario 7-3: 추적성 체인 무결성 검증**
```gherkin
Given @TASK:IMPL-FEATURE TAG가 있는 파일이 커밋되고
  And 연결된 @REQ TAG가 존재하지 않을 때
When 추적성 검증이 실행되면
Then 시스템은 "추적성 체인 끊김" 경고를 발생시키고
  And 누락된 요구사항 태그 생성을 제안하며
  And 불완전한 추적성을 허용할지 개발자에게 확인한다
```

**Scenario 7-4: 중복 TAG 방지 검증**
```gherkin
Given 새로 추가되는 파일에 "@API:GET-USERS" TAG가 있고
  And 동일한 TAG가 다른 파일에 이미 존재할 때
When commit 시점 검증이 실행되면
Then 시스템은 중복을 감지하고 차단하며
  And 자동으로 "@API:GET-USER-LIST" 등 고유 식별자를 제안하고
  And 중복 해결 후에만 commit을 진행한다
```

---

### US-008: TAG 커버리지 모니터링

**Scenario 8-1: 프로젝트 전체 TAG 커버리지 측정**
```gherkin
Given src/moai_adk/ 디렉토리에 100개의 Python 파일이 있고
  And 그 중 95개에 적절한 TAG가 있을 때
When 커버리지 측정 스크립트를 실행하면
Then "TAG 커버리지: 95%" 결과가 표시되고
  And 목표 수준(95% 이상) 달성 여부가 표시되며
  And 누락된 5개 파일의 목록과 권장 TAG가 제안된다
```

**Scenario 8-2: 카테고리별 커버리지 분석**
```gherkin
Given 16-Core TAG 시스템의 4개 카테고리가 있고
  And 각 카테고리별로 파일 분포가 다를 때
When 상세 커버리지 분석을 실행하면
Then SPEC(90%), IMPLEMENTATION(95%), QUALITY(80%) 등
  And 카테고리별 커버리지가 개별적으로 표시되고
  And 부족한 카테고리에 대한 개선 권장 사항이 제공된다
```

**Scenario 8-3: 실시간 커버리지 모니터링**
```gherkin
Given 개발팀이 활발히 작업 중이고
  And 실시간 모니터링 대시보드가 활성화되어 있을 때
When 새 파일이 추가되거나 TAG가 변경되면
Then 커버리지 지표가 즉시 업데이트되고
  And 목표 수준 이하로 떨어지면 Slack 알림을 발송하며
  And 주간 커버리지 트렌드 리포트를 자동 생성한다
```

---

### US-009: 중복 및 충돌 TAG 자동 해결

**Scenario 9-1: 동일 TAG 중복 자동 해결**
```gherkin
Given 두 개의 파일에 "@API:GET-USERS" TAG가 중복되어 있고
  And 자동 해결 시스템이 활성화되어 있을 때
When 중복 검사 스크립트가 실행되면
Then 시스템은 생성 시간 순으로 우선순위를 매기고
  And 나중 파일을 "@API:GET-USER-LIST"로 자동 변경하며
  And 변경 내역을 .moai/logs/tag-resolution.log에 기록한다
```

**Scenario 9-2: 의미적 충돌 TAG 감지**
```gherkin
Given "@API:POST-LOGIN" 과 "@API:POST-AUTH" TAG가 있고
  And 두 기능이 실제로 동일한 로직을 구현할 때
When 의미적 중복 분석이 실행되면
Then 시스템은 코드 유사성을 분석하여 잠재적 중복을 감지하고
  And 개발자에게 코드 통합을 제안하며
  And Constitution 단순성 원칙 준수를 권장한다
```

**Scenario 9-3: TAG 네이밍 충돌 해결**
```gherkin
Given "@FEATURE:ENGINE" 과 "@API:GET-ENGINE" TAG가 있고
  And 서로 다른 엔진 개념을 다룰 때
When 네이밍 충돌 검사가 실행되면
Then 시스템은 더 구체적인 명명을 제안하고
  And "@FEATURE:ORCHESTRATOR", "@API:GET-STATUS" 등으로 구체화하며
  And 기존 코드의 TAG 참조를 자동으로 업데이트한다
```

**Scenario 9-4: 대량 TAG 충돌 일괄 해결**
```gherkin
Given 레거시 프로젝트 마이그레이션으로 50개의 TAG 충돌이 발생하고
  And 수동 해결이 비현실적일 때
When 일괄 해결 도구를 실행하면
Then 시스템은 충돌 패턴을 분석하여 해결 규칙을 학습하고
  And 80% 이상의 충돌을 자동으로 해결하며
  And 나머지 20%는 우선순위를 매겨 수동 검토 대상으로 분류한다
```

---

## 🔧 도구 통합 및 자동화

### US-010: VS Code 확장 기능 통합

**Scenario 10-1: TAG 자동완성 기능**
```gherkin
Given VS Code에서 Python 파일을 편집 중이고
  And MoAI TAG 확장이 설치되어 있을 때
When 개발자가 파일 상단에서 "@"를 입력하면
Then 16-Core TAG 카테고리가 드롭다운으로 표시되고
  And 프로젝트 기존 패턴을 학습한 제안이 나타나며
  And 실시간으로 TAG 형식 검증과 중복 검사가 수행된다
```

**Scenario 10-2: 실시간 TAG 검증**
```gherkin
Given 개발자가 잘못된 형식으로 TAG를 입력하고
  And VS Code 확장이 활성화되어 있을 때
When "@api-user-login" 같은 잘못된 형식을 입력하면
Then 빨간 밑줄로 오류가 표시되고
  And 올바른 형식인 "@API:POST-LOGIN"이 제안되며
  And 자동 수정 옵션이 Quick Fix로 제공된다
```

**Scenario 10-3: TAG 기반 코드 네비게이션**
```gherkin
Given 프로젝트에 관련된 TAG들이 연결되어 있고
  And 개발자가 특정 TAG 위에서 Ctrl+Click을 할 때
When "@REQ:USER-AUTH-001" TAG를 클릭하면
Then 관련된 설계, 구현, 테스트 파일로 빠르게 이동할 수 있고
  And 사이드바에 관련 TAG 목록이 표시되며
  And 추적성 체인을 시각적으로 확인할 수 있다
```

**Scenario 10-4: TAG 기반 파일 탐색**
```gherkin
Given VS Code Command Palette이 열려 있고
  And 개발자가 특정 기능 관련 파일을 찾으려 할 때
When "MoAI: Find by TAG" 명령을 실행하고 "@API"를 입력하면
Then 모든 API 관련 파일이 목록으로 표시되고
  And 파일명과 함께 TAG 설명이 미리보기로 표시되며
  And 선택 시 해당 파일의 TAG 위치로 바로 이동한다
```

---

### US-011: 대량 TAG 마이그레이션 도구

**Scenario 11-1: 레거시 프로젝트 일괄 분석**
```gherkin
Given 기존 프로젝트에 500개의 Python 파일이 있고
  And TAG가 전혀 적용되지 않은 상태일 때
When 마이그레이션 도구를 실행하면
Then 시스템은 10분 이내에 모든 파일을 분석하고
  And 각 파일의 기능을 추론하여 적절한 TAG를 제안하며
  And 신뢰도 점수와 함께 권장 사항을 생성한다
```

**Scenario 11-2: 기존 주석 시스템에서 마이그레이션**
```gherkin
Given 레거시 코드에 "# Module: UserAPI" 같은 기존 주석이 있고
  And 이를 16-Core TAG 시스템으로 변환하려 할 때
When 변환 규칙을 적용하면
Then "# Module: UserAPI"는 "@API:GET-USERS"로 변환되고
  And 기존 주석은 보존하되 새로운 TAG가 추가되며
  And 변환 과정과 결과가 상세히 로깅된다
```

**Scenario 11-3: 점진적 마이그레이션 지원**
```gherkin
Given 대규모 프로젝트에서 한 번에 모든 파일을 변경할 수 없고
  And 점진적 적용이 필요할 때
When 우선순위 기반 마이그레이션을 실행하면
Then 핵심 모듈부터 우선 적용되고
  And 각 단계별로 검증과 테스트가 수행되며
  And 기존 코드와의 호환성이 유지된다
```

**Scenario 11-4: 마이그레이션 품질 검증**
```gherkin
Given 마이그레이션이 완료된 후
  And 자동 적용된 TAG들의 품질을 검증해야 할 때
When 검증 스크립트를 실행하면
Then 적용된 TAG의 90% 이상이 수동 검토 기준과 일치하고
  And 잘못된 분류는 자동으로 감지되어 수정 제안되며
  And 마이그레이션 성공률과 품질 지표가 리포트로 생성된다
```

**Scenario 11-5: 대용량 프로젝트 처리**
```gherkin
Given 10,000개 이상의 파일을 가진 대규모 프로젝트이고
  And 메모리와 처리 시간이 제한적일 때
When 배치 마이그레이션 모드를 실행하면
Then 시스템은 파일을 청크 단위로 나누어 처리하고
  And 메모리 사용량을 1GB 이하로 유지하며
  And 중간 결과를 저장하여 중단 시 재시작이 가능하다
```

---

### US-012: TAG 시스템 성능 최적화

**Scenario 12-1: 대규모 파일 스캔 성능**
```gherkin
Given 프로젝트에 1,000개의 Python 파일이 있고
  And 전체 TAG 인덱싱이 필요할 때
When 성능 최적화된 스캔을 실행하면
Then 5초 이내에 모든 파일 스캔이 완료되고
  And CPU 사용률이 80% 이하로 유지되며
  And 메모리 사용량이 500MB를 초과하지 않는다
```

**Scenario 12-2: 실시간 파일 감시 성능**
```gherkin
Given 개발자가 파일을 활발히 수정하고 있고
  And 실시간 TAG 인덱스 업데이트가 활성화되어 있을 때
When 초당 10개 파일이 변경되면
Then 각 변경 사항이 100ms 이내에 인덱스에 반영되고
  And 시스템 응답성이 저하되지 않으며
  And 배경 작업이 개발 워크플로우를 방해하지 않는다
```

**Scenario 12-3: 메모리 효율성 최적화**
```gherkin
Given 장시간 실행되는 TAG 모니터링 시스템이 있고
  And 메모리 누수를 방지해야 할 때
When 24시간 연속 실행하면
Then 메모리 사용량이 일정 수준으로 유지되고
  And 가비지 컬렉션이 효율적으로 수행되며
  And 시스템 리소스가 안정적으로 관리된다
```

**Scenario 12-4: 네트워크 및 디스크 I/O 최적화**
```gherkin
Given 원격 저장소나 네트워크 드라이브의 파일을 처리하고
  And I/O 지연이 발생할 수 있을 때
When 최적화된 파일 접근 모드를 사용하면
Then 비동기 I/O를 활용하여 처리 속도를 향상시키고
  And 로컬 캐싱으로 반복 접근을 최적화하며
  And 네트워크 장애 시 우아한 실패 처리를 제공한다
```

---

## 📊 전체 성공 지표 요약

### 기능적 성공 지표
- **자동 TAG 적용율**: 95% 이상
- **TAG 형식 정확성**: 99% 이상
- **실시간 인덱스 업데이트**: 100ms 이내
- **추적성 체인 완성도**: 90% 이상

### 성능 지표
- **대량 스캔 성능**: 1000 파일/5초 이내
- **메모리 사용량**: 500MB 이하 유지
- **시스템 응답성**: 개발 워크플로우 무방해
- **중복 해결 정확성**: 95% 이상

### 사용자 경험 지표
- **자동화 만족도**: 4.5/5 이상
- **개발 생산성 향상**: 20% 이상
- **TAG 검색 효율성**: 3배 향상
- **도구 통합 만족도**: 4.0/5 이상

---

> **@REQ:CODE-TAG-002** 태그를 통해 모든 수락 기준이 상위 요구사항과 연결됩니다.
>
> **테스트 전략**: 각 시나리오는 자동화된 테스트로 구현되어 지속적인 품질을 보장합니다.