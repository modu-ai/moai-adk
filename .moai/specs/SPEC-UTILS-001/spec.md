---
id: UTILS-001
version: 0.0.1
status: draft
created: 2025-11-13
updated: 2025-11-13
author: @Goos
priority: high
category: utilities
labels:
  - banner
  - utilities
  - dynamic-configuration
  - theme-support
  - test-coverage
related_specs:
  - SPEC-CLI-001
scope:
  packages:
    - src/moai_adk/utils/
    - tests/unit/
  files:
    - src/moai_adk/utils/banner.py
    - tests/unit/test_utils_banner.py
---

# @SPEC:UTILS-001: 유틸리티 배너 기능 강화 및 테스트

> 배너 시스템을 동적으로 구성 가능하도록 개선하고 테스트 커버리지를 85% 이상으로 확대하는 SPEC

## 히스토리

### v0.0.1 (2025-11-13)
- **초기 생성**: 배너 시스템 기능 강화 및 테스트 커버리지 확대 SPEC 초기 작성
- **작성자**: @Goos
- **범위**: 동적 배너 구성, 테마 지원, 테스트 커버리지 강화
- **배경**: 기존 배너 기능의 제한적 커스터마이징과 불완전한 테스트 커버리지 개선 필요

---

## 환경 (Environment)

### 실행 환경
- **프로젝트**: MoAI-ADK 유틸리티 모듈 (Python 기반)
- **핵심 파일**: `src/moai_adk/utils/banner.py`
- **테스트 파일**: `tests/unit/test_utils_banner.py`
- **현재 버전**: MoAI-ADK v0.22.5
- **의존성**: Rich Console (>=13.0.0) for 터미널 출력

### 기술 스택
- **언어**: Python 3.10+
- **출력 라이브러리**: Rich (colored terminal output)
- **테스트 프레임워크**: pytest (>=8.0.0)
- **목(Mock) 도구**: unittest.mock
- **코드 품질**: pylint, black, mypy

### 제약사항
- **터미널 호환성**: 모든 주요 터미널 지원 (색상 지원)
- **성능**: 배너 출력 시간 < 100ms
- **테스트 커버리지**: 최소 85% (라인 커버리지)
- **의존성**: 외부 의존성 최소화 (Rich만 사용)

---

## 가정사항 (Assumptions)

### 기술적 가정
1. Rich 라이브러리가 안정적으로 컬러 렌더링을 지원함
2. pytest mocking이 console.print() 함수를 정확히 가로챌 수 있음
3. 동적 테마 구성이 YAML 또는 딕셔너리로 정의 가능함
4. 배너 및 메시지는 문자열 기반으로 구성됨

### 사용자 가정
1. MoAI-ADK 초기화 시 배너가 표시됨
2. 사용자는 배너의 외형을 커스터마이징할 수 없음 (현재)
3. 개발자는 프로그래매틱하게 배너 동작을 제어할 수 있어야 함

### 프로젝트 가정
1. 기존 배너 기능(`print_banner`, `print_welcome_message`)은 유지보수됨
2. 새로운 기능은 기존 API와 호환성을 유지해야 함
3. 테스트는 모든 기능에 대해 작성됨

---

## 요구사항 (EARS 구조)

### 기본 요구사항 (Ubiquitous)

**REQ-001**: 시스템은 정적 배너 구성을 유지해야 한다
```
Given: 어떤 배너 인스턴스든
When: print_banner() 호출됨
Then: 다음을 보장해야 함
  • ASCII 아트 배너를 표시
  • 버전 정보 포함
  • 청록색(cyan) 컬러 적용
  • 메시지 텍스트 정확함
```

**REQ-002**: 시스템은 컬러 출력을 지원해야 한다
```
Given: Rich console 초기화됨
When: 배너 출력 함수 호출
Then: 다음 컬러 적용
  • 배너: cyan
  • 자막: dim
  • 버전: dim
```

**REQ-003**: 시스템은 모든 배너 함수의 동작을 테스트 가능해야 한다
```
Given: 테스트 환경
When: 배너 함수 호출
Then: 다음 검증 가능
  • console.print 호출 횟수
  • 각 호출의 인자 값
  • 컬러 마크업 포함 여부
```

### 이벤트 기반 요구사항 (Event-driven)

**REQ-004**: WHEN 프로그램이 시작되면, 배너가 자동으로 출력되어야 한다
```
Given: MoAI-ADK CLI 초기화
When: 프로그램 시작
Then: 다음 순서로 실행
  1. print_banner(version) 호출 → 배너 표시
  2. print_welcome_message() 호출 → 환영 메시지 표시
  3. 사용자 입력 대기
```

**REQ-005**: WHEN 커스텀 버전을 지정하면, 해당 버전이 표시되어야 한다
```
Given: 커스텀 버전 문자열 (예: "1.0.0-beta")
When: print_banner(version="1.0.0-beta") 호출
Then: 다음 수행
  • "Version: 1.0.0-beta" 출력
  • 나머지 배너는 동일하게 표시
```

### 상태 기반 요구사항 (State-driven)

**REQ-006**: WHILE 배너 모듈이 활성화되어 있을 때, console 객체는 전역으로 유지되어야 한다
```
Given: Rich console 객체 생성됨
When: 여러 함수가 순차적으로 호출
Then: 모든 출력이 동일한 console을 통해 처리됨
```

### 선택적 요구사항 (Optional)

**REQ-007**: IF 동적 테마 구성이 필요하면, 색상 및 스타일을 파라미터로 설정할 수 있어야 한다
```
Given: BannerConfig 객체
When: 테마 설정 제공
Then: 다음 구성 가능
  • 배너 배경색 (cyan → green, magenta 등)
  • 텍스트 스타일 (bold, dim, italic)
  • 줄바꿈 개수 (현재 2개)
```

**REQ-008**: IF 여러 배너 유형이 필요하면, 시스템은 다양한 배너 템플릿을 지원해야 한다
```
Given: 다양한 배너 디자인 요구
When: 템플릿 시스템 활성화
Then: 지원
  • 기본 배너 (현재)
  • 미니 배너 (한 줄)
  • 상세 배너 (추가 정보 포함)
```

### 부정적 요구사항 (Unwanted Behaviors)

**REQ-009**: IF 잘못된 버전 문자열이 제공되면, 시스템은 정확히 처리해야 한다
```
Given: 비정상적인 버전 입력 (None, 빈 문자열, 유니코드)
When: print_banner(version=...) 호출
Then: 다음 수행
  • 기본값으로 폴백 또는 입력값 그대로 표시
  • 예외 발생 금지
  • 출력이 깨지지 않아야 함
```

**REQ-010**: IF console.print가 실패하면, 시스템은 우아하게 처리해야 한다
```
Given: console.print 호출 중 예외 발생
When: 출력 실패
Then: 다음 수행
  • 예외를 적절히 처리
  • 에러 메시지 출력 또는 로깅
  • 프로그램 중단 금지
```

---

## 명세사항 (Specifications)

### 배너 모듈 개선

#### 1. 동적 배너 구성 (Dynamic Banner Configuration)

배너의 색상과 스타일을 동적으로 설정할 수 있도록 `BannerConfig` 클래스를 추가한다.

**변경 사항**:
```python
@dataclass
class BannerConfig:
    """배너 구성 설정"""
    ascii_art: str = MOAI_BANNER
    banner_color: str = "cyan"
    subtitle_style: str = "dim"
    version_style: str = "dim"
    enable_emoji: bool = True
    line_breaks: int = 2

def print_banner_with_config(version: str = "0.3.0",
                             config: BannerConfig = None) -> None:
    """구성 객체를 사용한 배너 출력"""
```

#### 2. 테마 지원 (Theme Support)

사전 정의된 테마를 지원하여 배너의 외형을 쉽게 변경할 수 있게 한다.

**지원 테마**:
- `light`: 밝은 색상 (green, bright_white)
- `dark`: 어두운 색상 (magenta, dim)
- `classic`: 기존 배너 (cyan, dim)
- `colorful`: 다채로운 (yellow → cyan → green)

#### 3. 배너 포맷팅 유틸리티

배너 텍스트 및 구조를 다루는 유틸리티 함수들을 추가한다.

**추가 함수**:
```python
def get_banner_ascii() -> str:
    """ASCII 배너 텍스트 반환"""

def format_banner_with_version(version: str) -> str:
    """버전과 함께 포맷된 배너 문자열 반환"""

def is_color_supported() -> bool:
    """터미널이 컬러를 지원하는지 확인"""
```

### 테스트 커버리지 확대

#### 1. 기존 테스트 유지
- `TestBannerConstants`: 배너 상수 검증
- `TestPrintBanner`: print_banner() 기본 기능
- `TestPrintWelcomeMessage`: print_welcome_message() 기본 기능

#### 2. 새로운 테스트 추가 (최소 10개 이상)

**구성 관련 테스트**:
- BannerConfig 기본값 검증
- 커스텀 컬러 설정 적용
- 테마별 구성 검증

**동작 테스트**:
- 빈 배너 처리
- 유니코드 문자 포함 버전
- 매우 긴 버전 문자열
- None 버전 처리

**에러 처리 테스트**:
- 잘못된 컬러 이름 처리
- console.print 실패 시뮬레이션
- 메모리 제약 상황 시뮬레이션

**통합 테스트**:
- 순차적 배너 출력 검증
- 테마 변경 후 재출력
- 다양한 터미널 환경 시뮬레이션

---

## 추적성 (@TAG 참조)

| TAG | 파일 | 설명 |
|-----|------|------|
| `@SPEC:UTILS-001` | `.moai/specs/SPEC-UTILS-001/spec.md` | 본 문서 |
| `@CODE:UTILS-001` | `src/moai_adk/utils/banner.py` | 배너 구현 코드 |
| `@TEST:UTILS-001` | `tests/unit/test_utils_banner.py` | 배너 테스트 |
| `@DOC:UTILS-001` | `docs/utils/banner.md` | 배너 사용자 문서 |

---

## 참고사항

- 본 SPEC은 /alfred:1-plan 명령으로 초기 생성되었습니다.
- 배너 모듈은 MoAI-ADK CLI의 초기 로딩 단계에서 호출됩니다.
- 테스트 커버리지 목표는 프로젝트의 TRUST 5 원칙 중 "T(Test First)" 원칙을 따릅니다.
