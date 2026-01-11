# SPEC-RALPH-001: 인수 조건

## TAG BLOCK

```yaml
SPEC-ID: SPEC-RALPH-001
Title: MoAI Ralph Engine - 인수 조건
Created: 2026-01-09
Status: Planned
Related-Files:
  - spec.md
  - plan.md
```

---

## 1. 인수 조건 개요

### 1.1 품질 게이트 기준

| 기준 | 목표값 | 측정 방법 |
|------|--------|----------|
| 테스트 커버리지 | ≥ 85% | pytest --cov |
| LSP 응답 시간 | ≤ 5초 | 타임아웃 테스트 |
| AST 스캔 시간 | ≤ 30초 | 벤치마크 테스트 |
| 메모리 사용량 | ≤ 500MB | 프로파일링 |
| 린트 경고 | 0개 | ruff check |
| 타입 오류 | 0개 | mypy/pyright |

### 1.2 Definition of Done

- [ ] 모든 EARS 요구사항에 대한 테스트 존재
- [ ] 모든 테스트 통과 (pytest)
- [ ] 코드 커버리지 ≥ 85%
- [ ] 린트 및 타입 체크 통과
- [ ] API 문서 완성
- [ ] 사용자 가이드 작성
- [ ] 코드 리뷰 승인

---

## 2. LSP 통합 레이어 테스트 시나리오

### 2.1 AC-LSP-001: LSP 서버 시작

```gherkin
Feature: LSP 서버 생명주기 관리

  Scenario: Python 파일 열 때 pyright 서버 시작
    Given LSP 설정 파일 ".lsp.json"이 존재한다
    And pyright-langserver가 시스템에 설치되어 있다
    When 사용자가 ".py" 확장자 파일을 연다
    Then pyright-langserver 프로세스가 시작되어야 한다
    And 서버 상태가 "running"이어야 한다

  Scenario: LSP 서버 미설치 시 graceful degradation
    Given LSP 설정 파일 ".lsp.json"이 존재한다
    And pyright-langserver가 시스템에 설치되어 있지 않다
    When 사용자가 ".py" 확장자 파일을 연다
    Then 시스템은 경고 메시지를 로깅해야 한다
    And 다른 기능은 정상 동작해야 한다

  Scenario: 유휴 서버 자동 종료
    Given Python LSP 서버가 실행 중이다
    And 열려 있는 Python 파일이 없다
    When 유휴 시간이 5분을 초과한다
    Then LSP 서버가 자동으로 종료되어야 한다
```

### 2.2 AC-LSP-002: 진단 정보 조회

```gherkin
Feature: LSP 진단 정보 조회

  Scenario: 타입 오류 감지
    Given Python LSP 서버가 실행 중이다
    And 다음 코드가 있는 파일이 있다:
      """
      def add(a: int, b: int) -> int:
          return a + b

      result: str = add(1, 2)  # Type error
      """
    When 진단 정보를 요청한다
    Then 진단 결과에 타입 오류가 포함되어야 한다
    And 오류 메시지에 "str"과 "int" 불일치가 명시되어야 한다
    And 오류 위치가 4번째 줄이어야 한다

  Scenario: 경고 감지
    Given TypeScript LSP 서버가 실행 중이다
    And 다음 코드가 있는 파일이 있다:
      """
      const unusedVariable = 42;
      console.log("hello");
      """
    When 진단 정보를 요청한다
    Then 진단 결과에 "unused variable" 경고가 포함되어야 한다

  Scenario: 진단 타임아웃 처리
    Given LSP 서버가 응답하지 않는 상태이다
    When 진단 정보를 요청한다
    And 5초가 경과한다
    Then 타임아웃 오류가 반환되어야 한다
    And 시스템은 정상 동작을 계속해야 한다
```

### 2.3 AC-LSP-003: 심볼 참조 검색

```gherkin
Feature: 심볼 참조 검색

  Scenario: 함수 참조 검색
    Given 다음과 같은 프로젝트 구조가 있다:
      | 파일 | 내용 |
      | main.py | from utils import helper; helper() |
      | utils.py | def helper(): pass |
    And Python LSP 서버가 실행 중이다
    When "utils.py"의 "helper" 함수에서 참조를 검색한다
    Then 결과에 "main.py"의 참조가 포함되어야 한다
    And 총 참조 수는 2개이어야 한다 (정의 1 + 사용 1)

  Scenario: 클래스 참조 검색
    Given 여러 파일에서 사용되는 클래스가 있다
    When 클래스 정의에서 참조를 검색한다
    Then 모든 import 및 사용 위치가 반환되어야 한다
```

### 2.4 AC-LSP-004: 심볼 이름 변경

```gherkin
Feature: 안전한 심볼 이름 변경

  Scenario: 함수 이름 변경
    Given 다음과 같은 프로젝트 구조가 있다:
      | 파일 | 내용 |
      | main.py | from utils import old_name; old_name() |
      | utils.py | def old_name(): pass |
    When "old_name"을 "new_name"으로 변경을 요청한다
    Then 변경 프리뷰에 2개 파일이 포함되어야 한다
    And "utils.py"에서 함수 정의가 변경되어야 한다
    And "main.py"에서 import와 호출이 변경되어야 한다

  Scenario: 이름 충돌 감지
    Given 변경하려는 새 이름이 이미 존재한다
    When 이름 변경을 요청한다
    Then 충돌 경고가 반환되어야 한다
    And 변경이 적용되지 않아야 한다
```

---

## 3. AST-grep 강화 레이어 테스트 시나리오

### 3.1 AC-AST-001: 단일 파일 스캔

```gherkin
Feature: AST-grep 단일 파일 스캔

  Scenario: 보안 취약점 탐지
    Given 다음 코드가 있는 Python 파일이 있다:
      """
      cursor.execute(f"SELECT * FROM users WHERE id = {user_id}")
      """
    When AST-grep 스캔을 실행한다
    Then "sql-injection-risk" 규칙 위반이 감지되어야 한다
    And 심각도가 "error"이어야 한다
    And 수정 제안이 포함되어야 한다

  Scenario: 품질 이슈 탐지
    Given 다음 코드가 있는 JavaScript 파일이 있다:
      """
      var oldStyle = "value";
      """
    When AST-grep 스캔을 실행한다
    Then "convert-var-to-const" 규칙 위반이 감지되어야 한다
    And 심각도가 "warning"이어야 한다

  Scenario: 지원되지 않는 파일 형식
    Given ".txt" 확장자 파일이 있다
    When AST-grep 스캔을 실행한다
    Then 스캔이 스킵되어야 한다
    And 적절한 메시지가 반환되어야 한다
```

### 3.2 AC-AST-002: 프로젝트 전체 스캔

```gherkin
Feature: AST-grep 프로젝트 스캔

  Scenario: 전체 프로젝트 스캔
    Given 다음 구조의 프로젝트가 있다:
      | 파일 | 이슈 수 |
      | src/main.py | 2 |
      | src/utils.py | 1 |
      | tests/test_main.py | 0 |
    When 프로젝트 전체 스캔을 실행한다
    Then 총 3개의 이슈가 감지되어야 한다
    And 파일별 이슈 수가 정확해야 한다
    And 심각도별 요약이 포함되어야 한다

  Scenario: 대규모 프로젝트 스캔 성능
    Given 1000개 이상의 파일이 있는 프로젝트가 있다
    When 프로젝트 전체 스캔을 실행한다
    Then 스캔이 30초 이내에 완료되어야 한다

  Scenario: 특정 디렉토리 제외
    Given 스캔 설정에서 "node_modules"가 제외되어 있다
    When 프로젝트 스캔을 실행한다
    Then "node_modules" 내 파일은 스캔되지 않아야 한다
```

### 3.3 AC-AST-003: 패턴 검색

```gherkin
Feature: 커스텀 패턴 검색

  Scenario: 함수 호출 패턴 검색
    Given 프로젝트에 여러 console.log 호출이 있다
    When 패턴 "console.log($MSG)"로 검색한다
    Then 모든 console.log 호출이 반환되어야 한다
    And 각 결과에 파일 경로와 위치가 포함되어야 한다

  Scenario: 가변 인자 패턴 검색
    Given 다양한 인자 수의 함수 호출이 있다
    When 패턴 "myFunc($$$ARGS)"로 검색한다
    Then 인자 수에 관계없이 모든 호출이 반환되어야 한다
```

### 3.4 AC-AST-004: 패턴 변환

```gherkin
Feature: 패턴 기반 코드 변환

  Scenario: Dry-run 모드 변환
    Given 다음 코드가 있는 파일이 있다:
      """
      console.log("debug");
      console.log("test");
      """
    When 패턴 "console.log($MSG)"를 "logger.debug($MSG)"로 dry-run 변환한다
    Then 2개의 변환 프리뷰가 반환되어야 한다
    And 원본 파일은 변경되지 않아야 한다

  Scenario: 실제 변환 적용
    Given dry-run 변환이 확인되었다
    When 변환을 실제로 적용한다
    Then 파일이 업데이트되어야 한다
    And 모든 console.log가 logger.debug로 변경되어야 한다

  Scenario: 변환 실패 시 롤백
    Given 변환 중 일부 파일에서 오류가 발생한다
    When 변환을 적용한다
    Then 모든 변경이 롤백되어야 한다
    And 원본 파일이 유지되어야 한다
```

---

## 4. 루프 컨트롤러 테스트 시나리오

### 4.1 AC-LOOP-001: 루프 시작

```gherkin
Feature: Ralph 루프 초기화

  Scenario: 새 루프 시작
    Given 활성 루프가 없다
    When promise "모든 타입 오류 해결"로 루프를 시작한다
    Then 새 루프 ID가 생성되어야 한다
    And 루프 상태가 "RUNNING"이어야 한다
    And 현재 반복 횟수가 0이어야 한다

  Scenario: 기존 루프 존재 시 시작
    Given 활성 루프가 이미 존재한다
    When 새 루프 시작을 요청한다
    Then 기존 루프 정보와 함께 오류가 반환되어야 한다
    And 기존 루프를 취소하거나 완료할 것을 안내해야 한다
```

### 4.2 AC-LOOP-002: 완료 조건 검사

```gherkin
Feature: 루프 완료 조건 검사

  Scenario: 모든 오류 해결됨
    Given promise가 "모든 LSP 오류 해결"이다
    And LSP 진단 결과에 오류가 없다
    When 완료 조건을 검사한다
    Then 완료 상태가 반환되어야 한다
    And 루프 상태가 "COMPLETED"로 변경되어야 한다

  Scenario: 오류가 남아있음
    Given promise가 "모든 LSP 오류 해결"이다
    And LSP 진단 결과에 2개의 오류가 있다
    When 완료 조건을 검사한다
    Then 미완료 상태가 반환되어야 한다
    And 남은 오류 정보가 포함되어야 한다

  Scenario: 부분 완료 (경고만 남음)
    Given promise가 "모든 오류 해결 (경고 허용)"이다
    And LSP 진단 결과에 오류는 없고 경고만 있다
    When 완료 조건을 검사한다
    Then 완료 상태가 반환되어야 한다
```

### 4.3 AC-LOOP-003: 피드백 루프 실행

```gherkin
Feature: LSP + AST 피드백 루프

  Scenario: 정상 피드백 생성
    Given 루프가 활성 상태이다
    And 파일에 LSP 오류 2개, AST 경고 1개가 있다
    When 피드백 루프를 실행한다
    Then LSP 진단이 수집되어야 한다
    And AST 스캔이 실행되어야 한다
    And 통합 피드백이 생성되어야 한다
    And 피드백에 우선순위 정렬된 이슈가 포함되어야 한다

  Scenario: 피드백 히스토리 기록
    Given 루프가 3번째 반복 중이다
    When 피드백 루프를 실행한다
    Then 현재 진단 상태가 히스토리에 추가되어야 한다
    And 이전 반복과 비교 정보가 포함되어야 한다
```

### 4.4 AC-LOOP-004: 루프 취소

```gherkin
Feature: 루프 취소

  Scenario: 사용자 취소
    Given 활성 루프가 있다
    When /cancel-loop 명령을 실행한다
    Then 루프 상태가 "CANCELLED"로 변경되어야 한다
    And 현재까지의 히스토리가 저장되어야 한다
    And 정리 작업이 수행되어야 한다

  Scenario: 취소할 루프 없음
    Given 활성 루프가 없다
    When /cancel-loop 명령을 실행한다
    Then "활성 루프가 없습니다" 메시지가 반환되어야 한다
```

### 4.5 AC-LOOP-005: 최대 반복 제한

```gherkin
Feature: 최대 반복 제한

  Scenario: 최대 반복 도달
    Given 루프가 9번째 반복 중이다 (max_iterations=10)
    And 오류가 여전히 존재한다
    When 피드백 루프를 실행한다
    And 완료 조건 검사가 실패한다
    Then 루프 상태가 "FAILED"로 변경되어야 한다
    And "최대 반복 횟수 초과" 메시지가 포함되어야 한다
    And 남은 이슈 요약이 포함되어야 한다

  Scenario: 사용자 정의 최대 반복
    Given 사용자가 max_iterations=20으로 설정했다
    When 루프를 시작한다
    Then 최대 20회까지 반복 가능해야 한다
```

---

## 5. 명령어 테스트 시나리오

### 5.1 AC-CMD-001: /alfred 명령

```gherkin
Feature: /alfred 원클릭 자동화

  Scenario: 기본 실행 (브랜치/PR 없음)
    Given 프로젝트가 초기화되어 있다
    And git.auto_branch와 git.auto_pr가 false이다
    When /alfred "사용자 인증 시스템" 명령을 실행한다
    Then /moai:1-plan이 실행되어야 한다
    And SPEC 생성 후 /moai:2-run이 실행되어야 한다
    And 구현 완료 후 /moai:3-sync가 실행되어야 한다
    And 현재 브랜치에서 작업이 완료되어야 한다

  Scenario: 브랜치 생성 옵션 사용
    Given git.auto_branch가 true이다
    When /alfred "로그인 기능" --branch 명령을 실행한다
    Then 새 브랜치 "feature/SPEC-XXX"가 생성되어야 한다
    And 해당 브랜치에서 작업이 진행되어야 한다

  Scenario: PR 생성 옵션 사용
    Given git.auto_pr가 true이다
    When /alfred "결제 시스템" --pr 명령을 실행한다
    Then 작업 완료 후 Draft PR이 생성되어야 한다
    And PR 설명에 SPEC 요약이 포함되어야 한다

  Scenario: 중간 실패 시 처리
    Given /moai:2-run 중 테스트 실패가 발생한다
    When /alfred을 실행한다
    Then 실패 지점에서 중단되어야 한다
    And 실패 정보가 명확히 보고되어야 한다
    And 복구 옵션이 제시되어야 한다
```

### 5.2 AC-CMD-002: /moai-loop 명령

```gherkin
Feature: /moai-loop Ralph 피드백 루프

  Scenario: 기본 루프 시작
    Given 활성 루프가 없다
    When /moai-loop "모든 타입 오류 해결" 명령을 실행한다
    Then 루프가 시작되어야 한다
    And 첫 번째 피드백이 생성되어야 한다
    And Claude가 수정 작업을 시작해야 한다

  Scenario: 커스텀 최대 반복
    Given 복잡한 리팩토링 작업이 필요하다
    When /moai-loop "API 마이그레이션 완료" --max-iterations 20 명령을 실행한다
    Then 최대 반복 횟수가 20으로 설정되어야 한다

  Scenario: 루프 완료까지 실행
    Given 해결 가능한 오류만 있다
    When /moai-loop "모든 린트 오류 해결" 명령을 실행한다
    Then Claude가 오류를 수정해야 한다
    And 모든 오류 해결 시 루프가 완료되어야 한다
    And 완료 요약이 표시되어야 한다
```

### 5.3 AC-CMD-003: /moai-fix 명령

```gherkin
Feature: /moai-fix 자동 수정

  Scenario: LSP 오류 자동 수정
    Given 파일에 자동 수정 가능한 LSP 오류가 있다
    When /moai-fix 명령을 실행한다
    Then LSP가 제공하는 Quick Fix가 적용되어야 한다
    And 수정된 항목 요약이 표시되어야 한다

  Scenario: AST-grep 경고 자동 수정
    Given 파일에 수정 제안이 있는 AST-grep 경고가 있다
    When /moai-fix 명령을 실행한다
    Then 제안된 수정이 적용되어야 한다
    And 변경 전/후 diff가 표시되어야 한다

  Scenario: 수정할 항목 없음
    Given 모든 진단이 깨끗하다
    When /moai-fix 명령을 실행한다
    Then "수정할 항목이 없습니다" 메시지가 표시되어야 한다
```

### 5.4 AC-CMD-004: /cancel-loop 명령

```gherkin
Feature: /cancel-loop 루프 취소

  Scenario: 활성 루프 취소
    Given 루프가 5번째 반복 중이다
    When /cancel-loop 명령을 실행한다
    Then 루프가 즉시 취소되어야 한다
    And 진행 상황 요약이 표시되어야 한다
    And 재시작 옵션이 안내되어야 한다

  Scenario: 활성 루프 없이 취소 시도
    Given 활성 루프가 없다
    When /cancel-loop 명령을 실행한다
    Then "취소할 활성 루프가 없습니다" 메시지가 표시되어야 한다
```

---

## 6. 훅 테스트 시나리오

### 6.1 AC-HOOK-001: PostToolUse LSP 진단

```gherkin
Feature: PostToolUse LSP 진단 훅

  Scenario: Write 후 진단 실행
    Given Python 파일에 타입 오류가 있다
    When Write 도구가 해당 파일을 수정한다
    Then post_tool__lsp_diagnostic.py 훅이 트리거되어야 한다
    And LSP 진단이 실행되어야 한다
    And 진단 결과가 additionalContext에 포함되어야 한다

  Scenario: Edit 후 진단 실행
    Given TypeScript 파일에 타입 오류가 있다
    When Edit 도구가 해당 파일을 수정한다
    Then LSP 진단이 실행되어야 한다
    And 결과가 Claude에 피드백되어야 한다

  Scenario: 지원되지 않는 언어 파일
    Given Markdown 파일이 수정되었다
    When Write 도구가 완료된다
    Then LSP 진단이 스킵되어야 한다
    And 훅이 정상 종료되어야 한다

  Scenario: LSP 서버 미설치 시 처리
    Given Go LSP 서버(gopls)가 설치되어 있지 않다
    When Go 파일이 수정된다
    Then 경고 메시지가 로깅되어야 한다
    And 훅이 exit code 0으로 종료되어야 한다
```

### 6.2 AC-HOOK-002: Stop 훅 루프 컨트롤

```gherkin
Feature: Stop 훅 Ralph 루프 컨트롤

  Scenario: 루프 미완료 - 추가 작업 요청
    Given Ralph 루프가 활성 상태이다
    And LSP 진단에 오류가 남아있다
    When Claude 응답이 완료된다
    Then stop__loop_controller.py 훅이 트리거되어야 한다
    And 완료 조건 검사가 실패해야 한다
    And decision이 "BLOCK"이어야 한다
    And follow_up_prompt에 다음 작업이 포함되어야 한다

  Scenario: 루프 완료 - 종료 허용
    Given Ralph 루프가 활성 상태이다
    And 모든 오류가 해결되었다
    When Claude 응답이 완료된다
    Then 완료 조건 검사가 성공해야 한다
    And decision이 "ALLOW"이어야 한다
    And 루프가 "COMPLETED" 상태로 변경되어야 한다

  Scenario: 루프 비활성 시 훅 패스
    Given 활성 루프가 없다
    When Claude 응답이 완료된다
    Then 훅이 즉시 종료되어야 한다
    And decision이 "ALLOW"이어야 한다
```

---

## 7. 통합 테스트 시나리오

### 7.1 AC-INT-001: 전체 플로우 통합

```gherkin
Feature: Ralph Engine 전체 플로우

  Scenario: LSP + AST-grep + Loop 통합 동작
    Given Python 프로젝트가 있다
    And 파일에 타입 오류 2개와 보안 취약점 1개가 있다
    When /moai-loop "모든 이슈 해결" 명령을 실행한다
    Then LSP 진단이 수집되어야 한다
    And AST-grep 스캔이 실행되어야 한다
    And Claude가 통합 피드백을 받아야 한다
    And Claude가 이슈를 수정해야 한다
    And 각 수정 후 PostToolUse 훅이 실행되어야 한다
    And 모든 이슈 해결 시 루프가 완료되어야 한다

  Scenario: 다중 언어 프로젝트 지원
    Given Python과 TypeScript 파일이 혼재된 프로젝트가 있다
    When /moai-loop "모든 타입 오류 해결" 명령을 실행한다
    Then 각 언어에 맞는 LSP 서버가 사용되어야 한다
    And 모든 언어의 진단이 통합되어야 한다
```

### 7.2 AC-INT-002: 오류 복구 시나리오

```gherkin
Feature: 오류 복구

  Scenario: LSP 서버 크래시 복구
    Given LSP 서버가 예기치 않게 종료되었다
    When 다음 진단 요청이 발생한다
    Then 서버가 자동으로 재시작되어야 한다
    And 진단이 정상 수행되어야 한다

  Scenario: AST-grep CLI 실패 복구
    Given AST-grep 스캔 중 오류가 발생한다
    When 루프가 계속 진행된다
    Then 해당 반복의 AST 결과가 스킵되어야 한다
    And LSP 결과만으로 피드백이 생성되어야 한다
    And 경고가 로깅되어야 한다

  Scenario: 네트워크 타임아웃 처리
    Given LSP 요청이 5초를 초과한다
    When 타임아웃이 발생한다
    Then 적절한 오류 메시지가 반환되어야 한다
    And 시스템이 정상 상태를 유지해야 한다
```

---

## 8. 성능 테스트 시나리오

### 8.1 AC-PERF-001: 응답 시간

```gherkin
Feature: 성능 요구사항

  Scenario: LSP 진단 응답 시간
    Given 1000줄 Python 파일이 있다
    When 진단을 요청한다
    Then 응답이 5초 이내에 반환되어야 한다

  Scenario: AST-grep 스캔 시간
    Given 500개 파일이 있는 프로젝트가 있다
    When 프로젝트 전체 스캔을 실행한다
    Then 스캔이 30초 이내에 완료되어야 한다

  Scenario: 루프 반복 오버헤드
    Given 루프가 활성 상태이다
    When 단일 반복이 실행된다
    Then 진단 + 피드백 생성이 10초 이내에 완료되어야 한다
```

### 8.2 AC-PERF-002: 리소스 사용량

```gherkin
Feature: 리소스 제한

  Scenario: 메모리 사용량 제한
    Given 5개의 LSP 서버가 동시에 실행 중이다
    When 메모리 사용량을 측정한다
    Then 총 메모리가 500MB를 초과하지 않아야 한다

  Scenario: CPU 사용량 안정성
    Given 프로젝트 스캔이 진행 중이다
    When CPU 사용량을 모니터링한다
    Then 평균 CPU 사용량이 50%를 초과하지 않아야 한다
```

---

## 9. 검증 방법

### 9.1 자동화 테스트

```yaml
테스트 실행 명령:
  유닛 테스트: pytest tests/ -v --cov=src/moai_adk
  통합 테스트: pytest tests/integration/ -v
  성능 테스트: pytest tests/performance/ -v --benchmark-only

커버리지 리포트:
  명령: pytest --cov=src/moai_adk --cov-report=html
  기준: 최소 85% 라인 커버리지

린트 검사:
  명령: ruff check src/
  기준: 경고 0개

타입 검사:
  명령: mypy src/moai_adk
  기준: 오류 0개
```

### 9.2 수동 검증

| 검증 항목 | 검증 방법 | 예상 결과 |
|----------|----------|----------|
| LSP 서버 시작 | Python 파일 열기 | pyright 프로세스 확인 |
| 진단 표시 | 타입 오류 코드 작성 | 오류 피드백 수신 |
| 루프 동작 | /moai-loop 실행 | 자동 수정 관찰 |
| 취소 동작 | /cancel-loop 실행 | 즉시 중단 확인 |

---

**문서 버전**: 1.0.0
**최종 수정**: 2026-01-09
**작성자**: workflow-spec agent
