# 스크린샷 가이드

**Phase 1 완료를 위한 필수 스크린샷 목록**

이 디렉토리에는 초보자 가이드용 스크린샷을 저장합니다.

## 필수 스크린샷 (7개)

### 1. 프로젝트 초기화 (2개)

**screenshot-01-moai-init.png**
- 캡처 명령: `moai-adk init my-project`
- 화면: 프로젝트 초기화 시작 화면
- 포함 요소:
  - 프로젝트 이름 입력 프롬프트
  - 언어 선택 옵션
  - 프레임워크 선택 메뉴

**screenshot-02-init-questions.png**
- 캡처 명령: 초기화 진행 중 대화형 질문
- 화면: 설정 질문 응답 과정
- 포함 요소:
  - conversation_language 선택
  - codebase_language 선택
  - git_strategy 선택 (solo/team)

### 2. SPEC 생성 (2개)

**screenshot-03-spec-creation.png**
- 캡처 명령: `/alfred:1-plan "사용자 인증 기능"`
- 화면: SPEC 생성 프로세스
- 포함 요소:
  - Alfred의 계획 분석
  - spec-builder 에이전트 호출
  - SPEC-001 생성 완료 메시지

**screenshot-04-spec-file.png**
- 캡처 대상: `.moai/specs/SPEC-001-user-auth.md` 파일 내용
- 화면: 생성된 SPEC 문서
- 포함 요소:
  - @SPEC 태그
  - EARS 형식 요구사항
  - 인수 조건
  - 추적성 섹션

### 3. TDD 실행 (3개)

**screenshot-05-red-phase.png**
- 캡처 명령: `/alfred:2-run SPEC-001` (RED 단계)
- 화면: 실패하는 테스트 작성
- 포함 요소:
  - tdd-implementer 에이전트 활성화
  - RED 단계 표시
  - 실패하는 테스트 코드
  - 테스트 실행 결과 (FAILED)

**screenshot-06-green-phase.png**
- 캡처 명령: `/alfred:2-run SPEC-001` (GREEN 단계)
- 화면: 테스트 통과 코드 작성
- 포함 요소:
  - GREEN 단계 표시
  - 최소 구현 코드
  - 테스트 실행 결과 (PASSED)

**screenshot-07-refactor-phase.png**
- 캡처 명령: `/alfred:2-run SPEC-001` (REFACTOR 단계)
- 화면: 코드 개선 및 최종 검증
- 포함 요소:
  - REFACTOR 단계 표시
  - 개선된 코드
  - 전체 테스트 통과 (ALL PASSED)

## 캡처 가이드라인

### 화면 크기
- 해상도: 1920x1080 이상 권장
- 비율: 16:9 또는 4:3

### 포맷
- 파일 형식: PNG (고품질)
- 파일 크기: 최대 2MB

### 주의사항
- 민감한 정보 (API 키, 토큰) 제거
- 터미널 프롬프트는 일반적인 형태로 표시
- 불필요한 배경 창 닫기
- 폰트 크기: 가독성 확보 (14pt 이상)

## 마크다운 임베딩 예시

```markdown
![프로젝트 초기화 과정](../../images/screenshot-01-moai-init.png)

이 화면에서 프로젝트 이름, 언어, 프레임워크를 선택합니다.
```

## 스크린샷 생성 스크립트

실제 환경에서 스크린샷 캡처를 위한 자동화 스크립트는 추후 제공될 예정입니다.

현재는 다음 도구를 사용하여 수동으로 캡처하세요:

- macOS: `Cmd + Shift + 4`
- Windows: `Win + Shift + S`
- Linux: `gnome-screenshot` 또는 `scrot`

______________________________________________________________________

*최종 업데이트: 2025-11-10 | Phase 1 | 상태: 대기*
