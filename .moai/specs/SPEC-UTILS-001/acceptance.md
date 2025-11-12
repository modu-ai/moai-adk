---
id: ACCEPTANCE-UTILS-001
version: 0.0.1
status: draft
created: 2025-11-13
updated: 2025-11-13
related_spec: SPEC-UTILS-001
---


## 개요

배너 기능 강화 및 테스트 커버리지 확대에 대한 인수 기준을 정의합니다. 모든 시나리오는 Given-When-Then 형식으로 작성되었습니다.

---

## 시나리오 1: 기본 배너 출력 (Backward Compatibility)

### 시나리오 설명
기존 `print_banner()` 함수가 호환성을 유지하며 정상 작동함을 검증합니다.

### Given-When-Then

**Given**: MoAI-ADK 프로그램이 시작되고, 기본 버전(0.3.0)이 설정됨
**When**: `print_banner()`를 명시적으로 호출
**Then**:
- [ ] console.print() 메서드가 정확히 3회 호출됨 (배너, 자막, 버전)
- [ ] 첫 번째 호출: MOAI_BANNER 상수 포함, cyan 색상 마크업 포함
- [ ] 두 번째 호출: "Modu-AI's Agentic Development Kit" 텍스트 포함
- [ ] 세 번째 호출: "Version: 0.3.0" 텍스트 포함
- [ ] 모든 출력이 Rich console 형식으로 처리됨

### 테스트 코드 구조
```python
def test_print_banner_backward_compatibility():
    """기존 print_banner() 함수의 호환성 검증"""
    with patch("moai_adk.utils.banner.console.print") as mock_print:
        # 실행
        print_banner()
        # 검증
        assert mock_print.call_count == 3
        # 각 호출 내용 검증
```

### 성공 기준
- ✓ 호출 횟수: 3회
- ✓ 색상 마크업: [cyan], [/cyan]
- ✓ 텍스트 정확성: MOAI_BANNER, Alfred, Version

---

## 시나리오 2: 커스텀 버전 출력

### 시나리오 설명
사용자가 커스텀 버전을 지정했을 때, 해당 버전이 정확히 출력됨을 검증합니다.

### Given-When-Then

**Given**: 커스텀 버전 문자열 "1.2.3-beta"가 준비됨
**When**: `print_banner(version="1.2.3-beta")`를 호출
**Then**:
- [ ] 세 번째 호출의 인자에 "1.2.3-beta" 문자열이 포함됨
- [ ] "Version: 1.2.3-beta" 형식으로 출력됨
- [ ] 나머지 배너는 동일하게 표시됨
- [ ] 예외 발생 없음

### 테스트 코드 구조
```python
def test_print_banner_custom_version():
    """커스텀 버전 출력 검증"""
    with patch("moai_adk.utils.banner.console.print") as mock_print:
        # 실행
        print_banner(version="1.2.3-beta")
        # 검증
        calls = [str(call) for call in mock_print.call_args_list]
        assert any("1.2.3-beta" in str(call) for call in calls)
```

### 성공 기준
- ✓ 커스텀 버전 포함: "1.2.3-beta"
- ✓ 포맷: "Version: {version}"
- ✓ 호출 횟수: 3회

---

## 시나리오 3: BannerConfig 기본값 설정

### 시나리오 설명
BannerConfig 클래스가 올바른 기본값을 가지고 있음을 검증합니다.

### Given-When-Then

**Given**: BannerConfig 클래스가 정의됨
**When**: `config = BannerConfig()`로 기본값으로 인스턴스 생성
**Then**:
- [ ] ascii_art == MOAI_BANNER
- [ ] banner_color == "cyan"
- [ ] subtitle_style == "dim"
- [ ] version_style == "dim"
- [ ] enable_emoji == True
- [ ] line_breaks == 2
- [ ] 모든 필드가 예상 타입 (str, int, bool)

### 테스트 코드 구조
```python
def test_banner_config_default_values():
    """BannerConfig 기본값 검증"""
    config = BannerConfig()
    assert config.banner_color == "cyan"
    assert config.enable_emoji == True
    assert config.line_breaks == 2
```

### 성공 기준
- ✓ 6개 필드 모두 올바른 기본값
- ✓ 타입 검증 통과
- ✓ dataclass 메서드 (\_\_eq\_\_, \_\_repr\_\_) 작동

---

## 시나리오 4: 커스텀 컬러 설정

### 시나리오 설명
사용자가 배너 색상을 커스텀으로 설정했을 때 정상 작동함을 검증합니다.

### Given-When-Then

**Given**: BannerConfig 인스턴스와 커스텀 색상 "green"이 준비됨
**When**:
```python
config = BannerConfig(banner_color="green", subtitle_style="bold")
print_banner_with_config(config=config)
```
**Then**:
- [ ] [green] 색상 마크업이 배너에 적용됨
- [ ] [bold] 스타일이 자막에 적용됨
- [ ] 나머지 필드는 기본값 사용
- [ ] console.print() 호출에서 색상 마크업 확인

### 테스트 코드 구조
```python
def test_banner_config_custom_colors():
    """커스텀 색상 적용 검증"""
    config = BannerConfig(banner_color="green")
    with patch("moai_adk.utils.banner.console.print") as mock_print:
        print_banner_with_config(config=config)
        # 첫 호출에서 [green] 포함 확인
        first_call = str(mock_print.call_args_list[0])
        assert "[green]" in first_call
```

### 성공 기준
- ✓ 커스텀 색상 적용: "green"
- ✓ 마크업 형식: [color]...[/color]
- ✓ 다른 필드는 기본값 유지

---

## 시나리오 5: Light 테마 로드

### 시나리오 설명
`get_theme_config("light")`가 light 테마 구성을 정확히 반환함을 검증합니다.

### Given-When-Then

**Given**: 테마 로드 함수가 구현됨
**When**: `config = get_theme_config("light")`를 호출
**Then**:
- [ ] banner_color == "green"
- [ ] version_style == "bright_white"
- [ ] enable_emoji == True
- [ ] 반환된 객체는 BannerConfig 인스턴스
- [ ] 객체의 모든 필드가 valid한 값

### 테스트 코드 구조
```python
def test_get_theme_config_light():
    """Light 테마 구성 로드"""
    config = get_theme_config("light")
    assert config.banner_color == "green"
    assert isinstance(config, BannerConfig)
```

### 성공 기준
- ✓ 테마명: "light"
- ✓ 색상: green
- ✓ 반환 타입: BannerConfig

---

## 시나리오 6: Dark 테마 로드

### 시나리오 설명
`get_theme_config("dark")`가 dark 테마 구성을 정확히 반환함을 검증합니다.

### Given-When-Then

**Given**: 테마 로드 함수가 구현됨
**When**: `config = get_theme_config("dark")`를 호출
**Then**:
- [ ] banner_color == "magenta"
- [ ] subtitle_style == "dim"
- [ ] version_style == "dim"
- [ ] 반환된 객체는 BannerConfig 인스턴스
- [ ] 모든 필드가 유효함

### 테스트 코드 구조
```python
def test_get_theme_config_dark():
    """Dark 테마 구성 로드"""
    config = get_theme_config("dark")
    assert config.banner_color == "magenta"
    assert config.subtitle_style == "dim"
```

### 성공 기준
- ✓ 테마명: "dark"
- ✓ 색상: magenta
- ✓ 반환 타입: BannerConfig

---

## 시나리오 7: 잘못된 테마명 처리

### 시나리오 설명
유효하지 않은 테마명이 입력되었을 때 기본값으로 폴백함을 검증합니다.

### Given-When-Then

**Given**: 존재하지 않는 테마명 "invalid_theme"이 준비됨
**When**: `config = get_theme_config("invalid_theme")`를 호출
**Then**:
- [ ] 예외 발생 없음 (graceful handling)
- [ ] 기본 테마(classic) 구성이 반환됨 또는 ValueError 발생
- [ ] 명확한 에러 메시지 포함
- [ ] 테마 리스트 제공 또는 로깅

### 테스트 코드 구조
```python
def test_get_theme_config_invalid_theme():
    """잘못된 테마명 처리"""
    # 방법 1: 기본값으로 폴백
    config = get_theme_config("invalid_theme")
    assert config.banner_color == "cyan"  # classic 테마

    # 방법 2: ValueError 발생
    with pytest.raises(ValueError):
        get_theme_config("nonexistent")
```

### 성공 기준
- ✓ 예외 안전성
- ✓ 기본값 폴백 또는 명확한 에러
- ✓ 로그 또는 메시지 출력

---

## 시나리오 8: 유니코드 버전 처리

### 시나리오 설명
유니코드 문자를 포함한 버전 문자열이 정상 처리됨을 검증합니다.

### Given-When-Then

**Given**: 유니코드 버전 문자열 "v1.0.0-한글-🎉"이 준비됨
**When**: `format_banner_with_version("v1.0.0-한글-🎉")`를 호출
**Then**:
- [ ] 유니코드 문자가 손실되지 않음
- [ ] 반환된 문자열에 모든 문자 포함
- [ ] 인코딩 에러 발생 없음
- [ ] 터미널에서 올바르게 표시됨

### 테스트 코드 구조
```python
def test_format_banner_unicode_version():
    """유니코드 버전 처리"""
    version = "v1.0.0-한글-🎉"
    result = format_banner_with_version(version)
    assert version in result
    assert "한글" in result
    assert "🎉" in result
```

### 성공 기준
- ✓ 유니코드 보존: 한글, 이모지
- ✓ 인코딩: UTF-8
- ✓ 문자열 완전성

---

## 시나리오 9: 긴 버전 문자열 처리

### 시나리오 설명
매우 긴 버전 문자열(예: 100자)이 정상 처리되고 배너가 깨지지 않음을 검증합니다.

### Given-When-Then

**Given**: 100자 이상의 긴 버전 문자열이 준비됨
**When**: `print_banner(version="v1.0.0-very-long-version-string-" * 5)`를 호출
**Then**:
- [ ] 예외 발생 없음
- [ ] 긴 문자열이 그대로 출력됨 또는 잘림 처리
- [ ] 배너 구조가 손상되지 않음
- [ ] 줄 바꿈이 적절히 처리됨

### 테스트 코드 구조
```python
def test_format_banner_long_version():
    """긴 버전 문자열 처리"""
    long_version = "v1.0.0-" + "x" * 100
    with patch("moai_adk.utils.banner.console.print") as mock_print:
        print_banner(version=long_version)
        # 출력이 깨지지 않았는지 확인
        assert mock_print.call_count == 3
```

### 성공 기준
- ✓ 예외 안전성
- ✓ 출력 포맷 유지
- ✓ 배너 구조 보존

---

## 시나리오 10: None 버전 처리

### 시나리오 설명
None 또는 빈 문자열이 버전으로 전달되었을 때 정상 처리됨을 검증합니다.

### Given-When-Then

**Given**: None 또는 빈 문자열이 버전 파라미터로 준비됨
**When**: `print_banner(version=None)` 또는 `print_banner(version="")`를 호출
**Then**:
- [ ] 예외 발생 없음
- [ ] 기본값(0.3.0)으로 폴백 또는 "Version: " 출력
- [ ] 배너가 출력됨
- [ ] console.print() 호출 횟수는 3회

### 테스트 코드 구조
```python
def test_format_banner_none_version():
    """None 버전 처리"""
    with patch("moai_adk.utils.banner.console.print") as mock_print:
        # None 처리
        print_banner(version=None)
        assert mock_print.call_count == 3

        # 빈 문자열 처리
        print_banner(version="")
        assert mock_print.call_count == 3
```

### 성공 기준
- ✓ None 안전성
- ✓ 빈 문자열 처리
- ✓ 기본값 폴백

---

## 시나리오 11: 컬러 지원 감지

### 시나리오 설명
`is_color_supported()` 함수가 터미널의 컬러 지원 여부를 정확히 감지함을 검증합니다.

### Given-When-Then

**Given**: 다양한 터미널 환경이 시뮬레이션됨 (CI, 로컬, no-color)
**When**: `is_color_supported()`를 호출
**Then**:
- [ ] 반환값은 boolean (True 또는 False)
- [ ] 환경 변수 NO_COLOR 감지
- [ ] CI 환경에서도 동작
- [ ] 예외 발생 없음

### 테스트 코드 구조
```python
def test_is_color_supported():
    """컬러 지원 감지"""
    result = is_color_supported()
    assert isinstance(result, bool)

    # NO_COLOR 환경 변수 감지
    with patch.dict(os.environ, {"NO_COLOR": "1"}):
        assert is_color_supported() == False
```

### 성공 기준
- ✓ 반환 타입: boolean
- ✓ NO_COLOR 감지
- ✓ 환경별 동작

---

## 시나리오 12: 이모지 비활성화

### 시나리오 설명
`enable_emoji=False`로 설정했을 때 이모지가 출력되지 않음을 검증합니다.

### Given-When-Then

**Given**: BannerConfig의 enable_emoji가 False로 설정됨
**When**: `print_banner_with_config(config=BannerConfig(enable_emoji=False))`를 호출
**Then**:
- [ ] 배너 출력에서 이모지 제거됨
- [ ] 텍스트만 포함됨 (또는 텍스트 대체 포함)
- [ ] console.print() 호출에서 이모지 미포함
- [ ] 출력이 ASCII 전용

### 테스트 코드 구조
```python
def test_banner_with_emoji_disabled():
    """이모지 비활성화 검증"""
    config = BannerConfig(enable_emoji=False)
    with patch("moai_adk.utils.banner.console.print") as mock_print:
        print_banner_with_config(config=config)
        # 이모지 미포함 확인
        calls = [str(call) for call in mock_print.call_args_list]
        all_output = " ".join(calls)
        assert "🎩" not in all_output  # Alfred emoji 확인
```

### 성공 기준
- ✓ 이모지 제거 또는 대체
- ✓ ASCII 전용 출력
- ✓ 가독성 유지

---

## 시나리오 13: 테마 동적 변경

### 시나리오 설명
런타임에 테마를 변경하여 서로 다른 배너가 출력됨을 검증합니다.

### Given-When-Then

**Given**: 세 개의 다른 테마(light, dark, classic)가 준비됨
**When**: 순서대로
```python
for theme in ["light", "dark", "classic"]:
    config = get_theme_config(theme)
    print_banner_with_config(config=config)
```
**Then**:
- [ ] 각 테마별로 다른 색상이 적용됨
- [ ] console.print() 호출이 각각 이루어짐 (9회 = 3회 × 3테마)
- [ ] 각 호출의 색상 마크업이 다름
- [ ] 메모리 누수 없음

### 테스트 코드 구조
```python
def test_theme_switching():
    """테마 동적 변경"""
    with patch("moai_adk.utils.banner.console.print") as mock_print:
        for theme in ["light", "dark", "classic"]:
            config = get_theme_config(theme)
            print_banner_with_config(config=config)
        # 9회 호출 (3 함수 × 3 테마)
        assert mock_print.call_count == 9
```

### 성공 기준
- ✓ 테마별 다른 색상
- ✓ 순차적 호출
- ✓ 메모리 안정성

---

## 시나리오 14: 순차적 배너 출력

### 시나리오 설명
프로그램 시작 시 배너와 환영 메시지가 순차적으로 출력됨을 검증합니다.

### Given-When-Then

**Given**: MoAI-ADK 초기화 프로세스가 시작됨
**When**:
```python
print_banner(version="0.22.5")
print_welcome_message()
```
를 순차적으로 호출
**Then**:
- [ ] print_banner() 호출 후 print_welcome_message() 호출
- [ ] console.print() 총 5회 이상 호출 (배너 3회 + 환영 2회+)
- [ ] 출력 순서가 맞음
- [ ] 각 메시지가 완전함

### 테스트 코드 구조
```python
def test_sequential_banner_output():
    """순차적 배너 출력 검증"""
    with patch("moai_adk.utils.banner.console.print") as mock_print:
        print_banner(version="0.22.5")
        print_welcome_message()
        # 총 5회 이상
        assert mock_print.call_count >= 5
```

### 성공 기준
- ✓ 호출 순서 유지
- ✓ 호출 횟수 정확
- ✓ 메시지 완전성

---

## 시나리오 15: console.print 예외 처리

### 시나리오 설명
console.print() 호출 중 예외 발생 시 프로그램이 안전하게 처리됨을 검증합니다.

### Given-When-Then

**Given**: console.print()가 임의의 지점에서 예외를 발생시키도록 mock 설정
**When**: `print_banner()` 호출
**Then**:
- [ ] 예외가 발생하지 않거나 적절히 처리됨
- [ ] 프로그램이 중단되지 않음
- [ ] 에러 로그가 기록됨 (선택적)
- [ ] 사용자에게 명확한 피드백 제공

### 테스트 코드 구조
```python
def test_error_handling_in_console_print():
    """console.print 예외 처리"""
    with patch("moai_adk.utils.banner.console.print") as mock_print:
        mock_print.side_effect = IOError("Terminal error")

        # 방법 1: 예외 발생
        with pytest.raises(IOError):
            print_banner()

        # 방법 2: 우아한 처리 (try-except)
        # 예외가 발생하지 않아야 함
        # print_banner()  # should not raise
```

### 성공 기준
- ✓ 예외 안전성
- ✓ 프로그램 안정성
- ✓ 에러 메시지 명확함

---

## 품질 게이트 (Quality Gates)

### 테스트 커버리지
- **라인 커버리지**: ≥ 85%
- **브랜치 커버리지**: ≥ 80%
- **함수 커버리지**: 100%

### 코드 품질
- **pylint 점수**: ≥ 8.0/10
- **black 포맷팅**: 통과
- **mypy 타입 체킹**: 통과 (no errors)
- **docstring**: 모든 함수/클래스 작성

### 테스트 실행
```bash
# 테스트 커버리지 측정
pytest tests/unit/test_utils_banner.py --cov=src/moai_adk/utils/banner --cov-report=html

# 코드 품질 검사
pylint src/moai_adk/utils/banner.py
mypy src/moai_adk/utils/banner.py

# 테스트 실행
pytest tests/unit/test_utils_banner.py -v
```

### 성공 조건
- [ ] 모든 15개 시나리오 테스트 PASS
- [ ] 커버리지 ≥ 85%
- [ ] pylint ≥ 8.0/10
- [ ] 0 개의 mypy 에러
- [ ] 모든 함수에 docstring

---

## 성공 기준 체크리스트

### 기능 완성도
- [ ] BannerConfig 클래스 구현
- [ ] 4개 테마 상수 정의
- [ ] get_theme_config() 함수 구현
- [ ] format_banner_with_version() 함수 구현
- [ ] is_color_supported() 함수 구현
- [ ] print_banner_with_config() 함수 구현

### 테스트 작성
- [ ] 15개 이상의 시나리오 기반 테스트 작성
- [ ] Given-When-Then 포맷 준수
- [ ] 모든 테스트 PASS
- [ ] 엣지 케이스 포함 (None, 빈 문자열, 유니코드, 긴 문자열)
- [ ] 에러 처리 테스트 포함

### 문서화
- [ ] 각 함수의 docstring 작성
- [ ] 파라미터 및 반환값 문서화
- [ ] 사용 예제 3개 이상 포함

### 품질 보증
- [ ] 라인 커버리지 85% 이상
- [ ] pylint 점수 8.0 이상
- [ ] mypy 타입 체크 통과
- [ ] black 포맷팅 통과
- [ ] 모든 테스트 GREEN

---

## 참고사항

- 이 문서는 /alfred:1-plan 명령으로 생성된 SPEC의 일부입니다.
- 각 시나리오는 독립적으로 테스트 가능합니다.
- 테스트는 TDD(Test-Driven Development) 방식으로 작성됩니다:
  - RED: 테스트 작성 후 실패 확인
  - GREEN: 최소 구현으로 테스트 통과
  - REFACTOR: 코드 개선 및 최적화
- 모든 테스트는 pytest 프레임워크를 사용합니다.
- mock 라이브러리로 console.print() 동작을 제어합니다.
