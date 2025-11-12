---
id: PLAN-UTILS-001
version: 0.0.1
status: draft
created: 2025-11-13
updated: 2025-11-13
related_spec: SPEC-UTILS-001
---


## 개요

유틸리티 배너 기능을 동적으로 구성 가능하도록 개선하고, 테스트 커버리지를 85% 이상으로 확대하는 실행 계획입니다.

---

## 마일스톤별 구현 단계

### 마일스톤 1: 기반 구조 (Foundation)
**목표**: 동적 배너 구성을 위한 클래스 및 기본 기능 구현

#### 1.1 BannerConfig 클래스 설계 및 구현
- **내용**: 배너 설정을 관리하는 dataclass 작성
- **담당**: tdd-implementer Agent
- **기술**:
  ```python
  from dataclasses import dataclass, field
  from typing import Optional

  @dataclass
  class BannerConfig:
      ascii_art: str
      banner_color: str
      subtitle_style: str
      version_style: str
      enable_emoji: bool
      line_breaks: int
  ```
- **검증**:
  - 모든 필드가 적절한 타입 검증
  - 기본값 설정
  - 유효한 색상 이름 검증

#### 1.2 테마 정의 (Theme Definitions)
- **내용**: 사전 정의된 테마 상수 작성
- **지원 테마**: light, dark, classic, colorful
- **각 테마별 설정**:
  - `LIGHT_THEME`: green 배너, bright_white 텍스트
  - `DARK_THEME`: magenta 배너, dim 텍스트
  - `CLASSIC_THEME`: cyan 배너, dim 텍스트 (현재)
  - `COLORFUL_THEME`: 다채로운 컬러 조합

#### 1.3 기본 유틸리티 함수
- **내용**: 배너 포맷팅 헬퍼 함수 구현
- **함수 목록**:
  - `get_banner_ascii()`: ASCII 배너 텍스트 반환
  - `format_banner_with_version(version: str)`: 버전 포함 포맷팅
  - `is_color_supported()`: 터미널 컬러 지원 확인

### 마일스톤 2: 핵심 기능 (Core Features)
**목표**: 동적 배너 출력 함수 및 기존 함수 통합

#### 2.1 print_banner_with_config 함수 구현
- **내용**: BannerConfig를 사용하는 새로운 배너 출력 함수
- **특징**:
  - 기존 `print_banner()` 호환성 유지
  - 선택적 config 파라미터
  - 런타임 테마 변경 지원
- **구현**:
  ```python
  def print_banner_with_config(
      version: str = "0.3.0",
      config: Optional[BannerConfig] = None
  ) -> None:
      """구성 객체를 사용한 배너 출력"""
  ```

#### 2.2 테마 로드 함수
- **내용**: 문자열 테마명으로 구성 객체 생성
- **함수**:
  ```python
  def get_theme_config(theme_name: str) -> BannerConfig:
      """테마명으로 구성 객체 반환"""
  ```
- **에러 처리**: 잘못된 테마명은 기본값(classic)으로 폴백

#### 2.3 기존 함수 개선
- **내용**: 기존 `print_banner()`, `print_welcome_message()` 유지보수
- **변경사항**:
  - 내부적으로 기본 config 사용
  - 공통 헬퍼 함수 활용 (DRY 원칙)
  - 코드 중복 제거

### 마일스톤 3: 고급 기능 (Advanced Features)
**목표**: 선택적 고급 기능 구현 (Optional)

#### 3.1 배너 캐싱 (Optional)
- **내용**: 포맷된 배너 문자열 캐싱
- **목표**: 반복 출력 시 성능 향상
- **구현**: 메모리 효율적인 캐싱

#### 3.2 커스텀 배너 등록 (Optional)
- **내용**: 사용자 정의 배너 추가 가능
- **API**: `register_custom_banner(name: str, config: BannerConfig)`

#### 3.3 배너 통계 (Optional)
- **내용**: 배너 출력 횟수, 테마별 사용 통계
- **저장소**: 메모리 기반 카운터

### 마일스톤 4: 테스트 커버리지 (Test Coverage)
**목표**: 85% 이상의 라인 커버리지 달성

#### 4.1 기존 테스트 유지 및 개선
- **유지**: TestBannerConstants, TestPrintBanner, TestPrintWelcomeMessage
- **개선**: 모든 테스트에 docstring 추가, 에러 케이스 확대

#### 4.2 새로운 단위 테스트 (10개 이상)

**그룹 A: 구성 관련 테스트 (3개)**
1. `test_banner_config_default_values()` - 기본값 검증
2. `test_banner_config_custom_colors()` - 커스텀 색상 적용
3. `test_banner_config_invalid_color()` - 잘못된 색상 처리

**그룹 B: 테마 관련 테스트 (3개)**
4. `test_get_theme_config_light()` - light 테마
5. `test_get_theme_config_dark()` - dark 테마
6. `test_get_theme_config_invalid_theme()` - 잘못된 테마명

**그룹 C: 버전 처리 테스트 (3개)**
7. `test_format_banner_unicode_version()` - 유니코드 버전
8. `test_format_banner_long_version()` - 긴 버전 문자열
9. `test_format_banner_none_version()` - None 버전 처리

**그룹 D: 고급 기능 테스트 (2개)**
10. `test_is_color_supported()` - 컬러 지원 감지
11. `test_banner_with_emoji_disabled()` - 이모지 비활성화

#### 4.3 통합 테스트 (3개)
12. `test_theme_switching()` - 테마 동적 변경
13. `test_sequential_banner_output()` - 순차 출력 검증
14. `test_error_handling_in_console_print()` - console.print 예외 처리

#### 4.4 테스트 커버리지 목표
- **라인 커버리지**: 최소 85%
- **브랜치 커버리지**: 최소 80%
- **함수 커버리지**: 100%

---

## 기술적 접근 방식

### 설계 원칙

1. **기존 호환성 (Backward Compatibility)**
   - 기존 `print_banner()`, `print_welcome_message()` API 유지
   - 새 기능은 opt-in 방식

2. **단일 책임 원칙 (Single Responsibility)**
   - 각 함수는 명확한 목적을 가짐
   - 헬퍼 함수로 코드 중복 제거

3. **테스트 주도 개발 (TDD)**
   - 각 기능마다 테스트 먼저 작성 (RED)
   - 최소 구현으로 테스트 통과 (GREEN)
   - 리팩토링으로 코드 품질 향상 (REFACTOR)

4. **문서화**
   - 모든 클래스와 함수에 docstring 작성
   - 사용 예제 포함

### 기술 선택

**dataclass 사용 이유**:
- 간단한 데이터 홀더
- 자동 __init__, __repr__, __eq__ 생성
- 타입 힌팅 지원

**함수형 헬퍼 vs 클래스 메서드**:
- 헬퍼 함수 우선 (함수형 스타일)
- 복잡도 높아지면 클래스 메서드 검토

**Mock 사용**:
- unittest.mock으로 console.print() 가로채기
- 실제 터미널 출력 없이 함수 동작 검증

### 아키텍처 구조

```
src/moai_adk/utils/banner.py
├── Constants
│   ├── MOAI_BANNER
│   ├── LIGHT_THEME (new)
│   ├── DARK_THEME (new)
│   ├── CLASSIC_THEME (new)
│   └── COLORFUL_THEME (new)
├── Classes
│   └── BannerConfig (new)
├── Public Functions
│   ├── print_banner() [기존]
│   ├── print_welcome_message() [기존]
│   └── print_banner_with_config() [new]
└── Helper Functions
    ├── get_banner_ascii() [new]
    ├── format_banner_with_version() [new]
    ├── get_theme_config() [new]
    └── is_color_supported() [new]
```

---

## 위험 요소 및 완화 전략

### 위험 1: 기존 코드와의 호환성 破損
**심각도**: 높음
**완화 전략**:
- 모든 새 기능은 선택적(optional)으로 설계
- 기존 함수의 시그니처 변경 금지
- 통합 테스트로 호환성 검증

### 위험 2: 테스트 커버리지 미달성
**심각도**: 중간
**완화 전략**:
- 각 마일스톤마다 커버리지 측정 (pytest-cov)
- 커버리지 < 85%일 경우 즉시 개선
- CI/CD에서 커버리지 게이트 설정

### 위험 3: 터미널 호환성 문제
**심각도**: 중간
**완화 전략**:
- `is_color_supported()` 함수로 사전 감지
- 컬러 미지원 환경에서는 기본 출력
- CI 환경에서 테스트 (Windows, Linux, macOS)

### 위험 4: 성능 저하
**심각도**: 낮음
**완화 전략**:
- 배너 포맷팅 시간 < 100ms 보장
- 선택적 캐싱으로 반복 출력 최적화
- 프로파일링으로 병목 확인

### 위험 5: 유니코드 문제
**심각도**: 낮음
**완화 전략**:
- 모든 테스트에서 유니코드 문자열 포함
- Rich 라이브러리의 유니코드 처리 검증
- 다양한 인코딩 환경에서 테스트

---

## 자원 요구사항

### 개발 자원
- **Python 3.10+**: 타입 힌팅 및 dataclass 지원
- **pytest 8.0+**: 테스트 프레임워크
- **pytest-cov**: 커버리지 측정
- **Rich 13.0+**: 터미널 렌더링

### 개발 시간
- **분석 및 설계**: 주요 결정사항 명확화
- **구현**: RED-GREEN-REFACTOR 주기 반복
- **테스트**: 각 기능마다 단위 테스트 작성
- **통합**: 마일스톤별 통합 테스트
- **문서화**: docstring 및 사용 예제

### 개발 환경
- 로컬 개발 머신
- Git 저장소
- CI/CD 파이프라인 (GitHub Actions)

---

## 예상 타임라인 (Priority 기반)

### 1순위: 핵심 기능 완성
**목표**: BannerConfig, 테마, 새로운 출력 함수 구현
- 필요성: 매우 높음
- 의존성: 없음
- 차단자: 없음

### 2순위: 테스트 커버리지 확대
**목표**: 85% 이상의 커버리지 달성
- 필요성: 높음 (TRUST 5 원칙)
- 의존성: 마일스톤 1 완료
- 차단자: 구현 완료

### 3순위: 문서 및 예제
**목표**: 사용자 가이드 및 API 문서 작성
- 필요성: 중간
- 의존성: 모든 구현 완료
- 차단자: 예제 코드 검증

### 4순위: 고급 기능 (Optional)
**목표**: 캐싱, 커스텀 배너, 통계
- 필요성: 낮음
- 의존성: 핵심 기능 안정화
- 차단자: 성능 요구사항 확인

---

## 성공 기준

### 기능 완성도
- [ ] BannerConfig 클래스 구현 및 테스트
- [ ] 4개 테마 구현 및 동작 검증
- [ ] 기존 함수와의 호환성 유지
- [ ] print_banner_with_config() 구현 및 테스트

### 테스트 품질
- [ ] 총 15개 이상의 새로운 테스트 추가
- [ ] 라인 커버리지 85% 이상
- [ ] 브랜치 커버리지 80% 이상
- [ ] 모든 테스트 패스

### 코드 품질
- [ ] pylint score ≥ 8.0/10
- [ ] black으로 포맷팅 완료
- [ ] mypy type checking 통과
- [ ] 모든 함수에 docstring 작성

### 문서화
- [ ] 각 함수별 docstring 작성
- [ ] 사용 예제 3개 이상 작성
- [ ] 테마별 사용 설명서 작성

---

## 참고사항

- 이 계획은 /alfred:1-plan 명령의 결과물입니다.
- 각 마일스톤은 `/alfred:2-run SPEC-UTILS-001` 명령으로 실행됩니다.
- 완료 후 `/alfred:3-sync auto SPEC-UTILS-001`로 문서화를 자동 동기화합니다.
- 필요에 따라 이 계획은 승인 후 수정될 수 있습니다.
