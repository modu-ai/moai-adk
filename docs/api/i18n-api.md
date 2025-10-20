# @DOC:I18N-001: i18n API Reference

> **SPEC**: SPEC-I18N-001.md | **CODE**: src/moai_adk/i18n.py | **TEST**: tests/test_i18n.py

MoAI-ADK의 다국어(i18n) 지원 API 레퍼런스입니다.

---

## 개요 (Overview)

MoAI-ADK는 5개 언어를 지원하는 다국어 시스템을 제공합니다:
- 🇰🇷 한국어 (ko) - 기본값
- 🇺🇸 영어 (en) - 글로벌 표준
- 🇯🇵 일본어 (ja)
- 🇨🇳 중국어 간체 (zh)
- 🇹🇭 태국어 (th)

---

## API Functions

### 1. `load_messages(locale)`

지정된 locale의 i18n 메시지를 로드합니다.

**시그니처**:
```python
@lru_cache(maxsize=5)
def load_messages(locale: str = "ko") -> dict[str, Any]
```

**파라미터**:
- `locale` (str, optional): 언어 코드 (ko, en, ja, zh, th). 기본값: "ko"

**반환값**:
- `dict[str, Any]`: 번역된 메시지 딕셔너리

**예외**:
- `FileNotFoundError`: 메시지 파일을 찾을 수 없을 때
- `json.JSONDecodeError`: JSON 파싱 실패 시

**사용 예시**:
```python
from moai_adk.i18n import load_messages

# 영어 메시지 로드
messages = load_messages("en")
print(messages["session_start"])
# Output: "🚀 MoAI-ADK Session Started"

# 한국어 메시지 로드 (기본값)
messages_ko = load_messages()
print(messages_ko["session_start"])
# Output: "🚀 MoAI-ADK 세션 시작"
```

**특징**:
- LRU 캐시를 사용하여 성능 최적화 (최대 5개 locale 캐싱)
- 지원하지 않는 locale은 자동으로 영어(en)로 대체
- 메시지 파일 위치: `docs/i18n/{locale}.json`

---

### 2. `t(key, locale, **kwargs)`

메시지 키를 번역합니다.

**시그니처**:
```python
def t(key: str, locale: str = "ko", **kwargs: Any) -> str
```

**파라미터**:
- `key` (str): 메시지 키 (예: "session_start", "error.no_git")
- `locale` (str, optional): 언어 코드. 기본값: "ko"
- `**kwargs`: 메시지 포매팅 변수

**반환값**:
- `str`: 번역되고 포매팅된 메시지

**사용 예시**:
```python
from moai_adk.i18n import t

# 기본 번역
print(t("session_start", "en"))
# Output: "🚀 MoAI-ADK Session Started"

# 변수 포함 번역
print(t("checkpoint_created", "ko", name="before-merge"))
# Output: "🛡️ 체크포인트 생성: before-merge"

# 중첩 키 지원 (dot notation)
print(t("error.no_git", "en"))
# Output: "❌ Not a Git repository"

# 다중 변수
print(t("context_loaded", "ja", count=3))
# Output: "📎 3個のコンテキストファイルを読み込みました"
```

**특징**:
- 점 표기법(dot notation)으로 중첩 키 지원 (예: "error.no_git")
- 번역을 찾지 못하면 키 자체를 반환 (fallback)
- Python 표준 `str.format()` 문법 사용
- 포매팅 오류 시 변수 없는 원본 메시지 반환

---

### 3. `get_supported_locales()`

지원되는 locale 목록을 반환합니다.

**시그니처**:
```python
def get_supported_locales() -> list[str]
```

**반환값**:
- `list[str]`: locale 코드 목록

**사용 예시**:
```python
from moai_adk.i18n import get_supported_locales

locales = get_supported_locales()
print(locales)
# Output: ['ko', 'en', 'ja', 'zh', 'th']

# locale 검증
user_locale = "fr"
if user_locale not in get_supported_locales():
    print(f"Locale '{user_locale}' not supported")
```

---

### 4. `get_locale_name(locale)`

locale의 사람이 읽을 수 있는 이름을 반환합니다.

**시그니처**:
```python
def get_locale_name(locale: str) -> str
```

**파라미터**:
- `locale` (str): 언어 코드

**반환값**:
- `str`: locale 이름 (영어 + 원어 병기)

**사용 예시**:
```python
from moai_adk.i18n import get_locale_name

print(get_locale_name("ko"))
# Output: "Korean (한국어)"

print(get_locale_name("ja"))
# Output: "Japanese (日本語)"

print(get_locale_name("zh"))
# Output: "Chinese (中文)"

# UI에서 사용
for locale in get_supported_locales():
    print(f"{locale}: {get_locale_name(locale)}")
# Output:
# ko: Korean (한국어)
# en: English
# ja: Japanese (日本語)
# zh: Chinese (中文)
# th: Thai (ไทย)
```

---

### 5. `validate_locale(locale)`

locale 코드를 검증하고 정규화합니다.

**시그니처**:
```python
def validate_locale(locale: str) -> str
```

**파라미터**:
- `locale` (str): 검증할 언어 코드

**반환값**:
- `str`: 검증된 locale 코드 (또는 fallback)

**사용 예시**:
```python
from moai_adk.i18n import validate_locale

# 유효한 locale
print(validate_locale("en"))
# Output: "en"

# 지원하지 않는 locale (fallback)
print(validate_locale("fr"))
# Output: "en"

# 대소문자 정규화
print(validate_locale("EN"))
# Output: "en"

# 공백 제거
print(validate_locale("  ko  "))
# Output: "ko"

# 사용자 입력 검증
user_input = "JA"
safe_locale = validate_locale(user_input)
messages = load_messages(safe_locale)
```

**특징**:
- 지원하지 않는 locale은 영어(en)로 대체
- 대소문자 정규화 (소문자 변환)
- 앞뒤 공백 제거
- 안전한 locale 값 보장

---

## 상수 (Constants)

### `SUPPORTED_LOCALES`

지원하는 locale 목록입니다.

```python
SUPPORTED_LOCALES = ["ko", "en", "ja", "zh", "th"]
```

### `DEFAULT_LOCALE`

기본 locale입니다.

```python
DEFAULT_LOCALE = "ko"
```

### `FALLBACK_LOCALE`

fallback locale입니다.

```python
FALLBACK_LOCALE = "en"
```

---

## 메시지 파일 구조

메시지 파일은 `docs/i18n/{locale}.json` 경로에 위치합니다.

### 스키마 예시 (en.json)

```json
{
  "session_start": "🚀 MoAI-ADK Session Started",
  "language": "Language",
  "branch": "Branch",
  "changes": "Changes",

  "checkpoint_created": "🛡️ Checkpoint created: {name}",
  "context_loaded": "📎 Loaded {count} context file(s)",

  "todo": {
    "analyzing_docs": "Analyzing project documents",
    "running_tests": "Running tests"
  },

  "error": {
    "no_git": "❌ Not a Git repository",
    "no_config": "❌ .moai/config.json not found",
    "locale_not_supported": "⚠️ Locale '{locale}' not supported"
  },

  "cli": {
    "init_success": "✅ Project '{name}' initialized",
    "doctor_passed": "✅ All checks passed"
  }
}
```

### 중첩 키 접근

```python
# 중첩 키는 점 표기법(dot notation)으로 접근
t("error.no_git", "en")
# ↓
# messages["error"]["no_git"]
# ↓
# "❌ Not a Git repository"
```

---

## 사용 패턴 (Usage Patterns)

### 패턴 1: Hook에서 사용

```python
# .claude/hooks/alfred/handlers/session.py

from moai_adk.i18n import t

def handle_session_start(payload):
    # config.json에서 locale 읽기
    locale = get_project_locale()

    # 번역된 메시지 생성
    message = t("session_start", locale)
    language = t("language", locale)

    return {
        "systemMessage": f"{message}\n{language}: Python"
    }
```

### 패턴 2: CLI에서 사용

```python
# src/moai_adk/cli/commands/init.py

from moai_adk.i18n import t, get_locale_name, get_supported_locales

def init_command():
    # locale 선택 프롬프트
    print("Select your language:")
    for locale in get_supported_locales():
        print(f"  {locale}: {get_locale_name(locale)}")

    user_locale = input("Choice: ")

    # 선택한 locale로 메시지 출력
    print(t("cli.init_start", user_locale))
```

### 패턴 3: TodoWrite에서 사용

```python
# Alfred 커맨드에서 사용

from moai_adk.i18n import t

def alfred_command(locale="ko"):
    todos = [
        {
            "content": t("todo.analyzing_docs", locale),
            "status": "in_progress",
            "activeForm": t("todo.analyzing_docs", locale) + " 중"
        }
    ]
    TodoWrite(todos)
```

---

## 성능 최적화

### LRU 캐시

`load_messages()` 함수는 `@lru_cache(maxsize=5)` 데코레이터를 사용하여 성능을 최적화합니다.

```python
# 첫 번째 호출: 파일 로드 (느림)
messages = load_messages("en")  # ~10ms

# 두 번째 호출: 캐시에서 반환 (빠름)
messages = load_messages("en")  # <1ms

# 캐시 크기: 5개 locale까지 메모리에 보관
# 메모리 사용량: 약 500KB (5개 locale × 100KB)
```

### 권장사항

- 동일한 locale을 반복 사용할 때는 `t()` 함수 사용
- 매번 `load_messages()`를 호출하지 않고 캐시 활용
- 메시지 파일 크기는 50KB 이하로 유지

---

## 에러 처리

### 파일 누락 시

```python
try:
    messages = load_messages("xx")  # 존재하지 않는 locale
except FileNotFoundError as e:
    print(f"Error: {e}")
    # Fallback to default
    messages = load_messages("en")
```

### JSON 파싱 오류 시

```python
# 잘못된 JSON 파일
# docs/i18n/broken.json: { "key": "value" (닫는 괄호 누락)

try:
    messages = load_messages("broken")
except json.JSONDecodeError as e:
    print(f"JSON Error: {e}")
    # Fallback to default
    messages = load_messages("en")
```

### 번역 키 누락 시

```python
# 존재하지 않는 키
result = t("non.existent.key", "en")
print(result)
# Output: "non.existent.key" (키 자체 반환)
```

---

## 테스트

### 단위 테스트 예시

```python
# tests/test_i18n.py

from moai_adk.i18n import load_messages, t, get_supported_locales

def test_load_messages_ko():
    """한국어 메시지 로드 테스트"""
    messages = load_messages("ko")
    assert messages["session_start"] == "🚀 MoAI-ADK 세션 시작"

def test_translate_with_format():
    """변수 포함 번역 테스트"""
    result = t("checkpoint_created", "ko", name="test")
    assert "체크포인트" in result
    assert "test" in result

def test_fallback_locale():
    """지원하지 않는 locale 테스트"""
    messages = load_messages("fr")  # 프랑스어 (미지원)
    assert messages["session_start"] == "🚀 MoAI-ADK Session Started"
```

---

## 관련 문서

- **SPEC**: [SPEC-I18N-001](/.moai/specs/SPEC-I18N-001/spec.md)
- **소스 코드**: [src/moai_adk/i18n.py](../../src/moai_adk/i18n.py)
- **테스트**: [tests/test_i18n.py](../../tests/test_i18n.py)
- **메시지 파일**: [docs/i18n/](../i18n/)

---

**Last Updated**: 2025-10-20
**Version**: v0.1.0
**Status**: Active ✅
