---
id: I18N-001
version: 0.1.0
status: active
created: 2025-10-20
updated: 2025-10-20
author: @Goos
priority: high
category: feature
labels:
  - i18n
  - internationalization
  - multilingual
  - localization
scope:
  packages:
    - src/moai_adk/i18n
    - .claude/hooks/alfred
    - src/moai_adk/cli
    - docs
---

# @SPEC:I18N-001: 5개 언어 다국어 지원 시스템 (i18n)

## HISTORY

### v0.1.0 (2025-10-20)
- **COMPLETED**: 5개 언어 다국어 지원 시스템 TDD 구현 완료
- **AUTHOR**: @Goos
- **TEST_COVERAGE**: test_i18n.py + integration tests (85%+)
- **RELATED**:
  - Hook 메시지 적용: `.claude/hooks/alfred/handlers/session.py`
  - CLI 메시지 적용: `src/moai_adk/cli/commands/init.py`
  - i18n 로더: `src/moai_adk/i18n.py`
  - README 다국어: `README.{ko,ja,zh,th}.md`
- **CHANGES**:
  - 버전 관리 SSOT 원칙 적용 (pyproject.toml ← 단일 진실의 출처)
  - `src/moai_adk/core/template/config.py`, `utils/banner.py` 동적 버전 로딩

### v0.0.1 (2025-10-20)
- **INITIAL**: 5개 언어(ko, en, ja, zh, th) 다국어 지원 시스템 명세 작성
- **AUTHOR**: @Goos
- **REASON**: 글로벌 오픈소스 프로젝트로 확장하기 위한 다국어 지원 필요
- **SCOPE**: Hook 메시지, CLI 출력, 문서, Git 커밋, TodoWrite 메시지

---

## 1. 개요 (Overview)

### 목적 (Purpose)

MoAI-ADK를 **글로벌 오픈소스 프로젝트**로 확장하기 위해 5개 언어를 지원하는 다국어 시스템을 구축합니다.

**지원 언어**:
- 🇰🇷 한국어 (ko) - 기본값
- 🇺🇸 영어 (en) - 글로벌 표준
- 🇯🇵 일본어 (ja) - 아시아 시장
- 🇨🇳 중국어 간체 (zh) - 중화권 시장
- 🇹🇭 태국어 (th) - 동남아시아 시장

### 범위 (Scope)

**포함 사항**:
- ✅ i18n 메시지 파일 (docs/i18n/{ko,en,ja,zh,th}.json)
- ✅ i18n 로더 모듈 (src/moai_adk/i18n.py)
- ✅ Hook 메시지 다국어화 (SessionStart, Checkpoint, Context)
- ✅ CLI 출력 메시지 다국어화
- ✅ 템플릿 문서 locale 변수 처리
- ✅ README 다국어 분리 (README.{locale}.md)

**제외 사항**:
- ❌ 코드 내부 변수명/함수명 (영어 고정)
- ❌ 자동 번역 도구 (수동 번역)
- ❌ 실시간 언어 전환 UI (향후 확장)

---

## 2. EARS 요구사항 (Requirements)

### Ubiquitous Requirements (기본 요구사항)

1. **시스템은 5개 언어(ko, en, ja, zh, th)를 지원해야 한다**
   - 각 언어별 메시지 파일을 제공해야 한다
   - locale 설정에 따라 자동으로 메시지를 선택해야 한다

2. **시스템은 `.moai/config.json`의 `project.locale` 값을 기반으로 언어를 결정해야 한다**
   - locale 값이 없으면 "ko"를 기본값으로 사용해야 한다
   - 지원하지 않는 locale은 "en"으로 대체해야 한다

3. **시스템은 i18n 메시지 로더를 제공해야 한다**
   - LRU 캐시를 사용하여 성능을 최적화해야 한다
   - JSON 파싱 오류 시 명확한 에러 메시지를 제공해야 한다

### Event-driven Requirements (이벤트 기반)

1. **WHEN SessionStart 이벤트가 발생하면, 시스템은 locale 기반 세션 시작 메시지를 표시해야 한다**
   - "🚀 MoAI-ADK 세션 시작" (ko)
   - "🚀 MoAI-ADK Session Started" (en)
   - "🚀 MoAI-ADKセッション開始" (ja)
   - "🚀 MoAI-ADK 会话开始" (zh)
   - "🚀 MoAI-ADK เซสชันเริ่มต้น" (th)

2. **WHEN Checkpoint가 생성되면, 시스템은 locale 기반 생성 메시지를 표시해야 한다**
   - "🛡️ 체크포인트 생성: {name}" (ko)
   - "🛡️ Checkpoint created: {name}" (en)
   - "🛡️ チェックポイント作成: {name}" (ja)
   - "🛡️ 检查点已创建: {name}" (zh)
   - "🛡️ สร้างจุดตรวจสอบ: {name}" (th)

3. **WHEN CLI 명령어가 실행되면, 시스템은 locale 기반 출력 메시지를 표시해야 한다**
   - 성공 메시지, 오류 메시지, 도움말 메시지 모두 locale 적용

4. **WHEN TodoWrite가 호출되면, 시스템은 locale 기반 작업 설명을 제공해야 한다**
   - `content`와 `activeForm` 모두 locale 기반 메시지 사용

### State-driven Requirements (상태 기반)

1. **WHILE 프로젝트가 초기화되지 않았을 때, 시스템은 locale 선택 프롬프트를 표시해야 한다**
   - `moai-adk init` 실행 시 5개 언어 중 선택 가능
   - 선택한 locale은 `.moai/config.json`에 저장

2. **WHILE 개발 모드일 때, 시스템은 누락된 번역 키를 경고해야 한다**
   - 디버그 로그에 "Missing translation: {key}" 출력
   - 누락된 키는 영어 메시지로 대체

### Optional Features (선택적 기능)

1. **WHERE 커뮤니티 기여가 있으면, 시스템은 추가 언어를 쉽게 확장할 수 있어야 한다**
   - 새 언어 추가는 JSON 파일 하나만 추가하면 됨
   - 번역 가이드 문서 제공 (CONTRIBUTING.md)

2. **WHERE 사용자가 요청하면, 시스템은 실시간으로 locale을 변경할 수 있어야 한다**
   - `moai-adk config set locale en` 명령어
   - 다음 세션부터 적용

### Constraints (제약사항)

1. **IF locale이 지원되지 않으면, 시스템은 영어(en)로 대체해야 한다**
   - 대체 시 경고 메시지 표시: "Locale 'xx' not supported, falling back to English"

2. **메시지 파일 크기는 50KB를 초과하지 않아야 한다**
   - 너무 긴 메시지는 별도 문서로 분리

3. **i18n 로더는 100ms 이내에 메시지를 로드해야 한다**
   - LRU 캐시 활용으로 성능 보장

---

## 3. 아키텍처 (Architecture)

### 언어 계층 구조 (Language Layers)

```
┌─────────────────────────────────────────┐
│ Layer 1: Code (영어 고정)               │
│  - 변수명, 함수명, 클래스명             │
│  - Docstring: 영어 + 한국어 병기         │
└─────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────┐
│ Layer 2: Documentation (다국어)         │
│  - README.{locale}.md                    │
│  - CLAUDE.{locale}.md (템플릿)          │
└─────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────┐
│ Layer 3: User-Facing (locale 기반)     │
│  - Hook 메시지 (SessionStart, etc)      │
│  - CLI 출력 메시지                       │
│  - Git 커밋 메시지                       │
│  - TodoWrite 메시지                      │
└─────────────────────────────────────────┘
```

### 파일 구조 (File Structure)

```
MoAI-ADK/
├── docs/
│   ├── i18n/
│   │   ├── en.json          # 영어 메시지
│   │   ├── ko.json          # 한국어 메시지
│   │   ├── ja.json          # 일본어 메시지
│   │   ├── zh.json          # 중국어 메시지
│   │   └── th.json          # 태국어 메시지
│   │
│   ├── README.md            # 영어 (기본)
│   ├── README.ko.md         # 한국어
│   ├── README.ja.md         # 일본어
│   ├── README.zh.md         # 중국어
│   └── README.th.md         # 태국어
│
├── src/
│   └── moai_adk/
│       ├── i18n.py          # i18n 로더 모듈 (NEW)
│       └── cli/
│           └── main.py      # CLI 메시지 i18n 적용
│
├── .claude/
│   └── hooks/
│       └── alfred/
│           ├── handlers/
│           │   ├── session.py      # SessionStart i18n
│           │   ├── checkpoint.py   # Checkpoint i18n
│           │   └── context.py      # Context i18n
│           └── core/
│               └── project.py      # locale 읽기 유틸
│
└── src/moai_adk/templates/
    ├── CLAUDE.md            # locale 변수 템플릿
    └── .moai/
        ├── config.json      # project.locale 필드
        └── memory/
            └── development-guide.md  # locale 변수 템플릿
```

---

## 4. 인터페이스 설계 (Interface Design)

### 4.1 i18n 로더 API

```python
# src/moai_adk/i18n.py

from functools import lru_cache
from pathlib import Path
import json

@lru_cache(maxsize=5)
def load_messages(locale: str = "ko") -> dict:
    """
    Load i18n messages for the specified locale.

    지정된 locale의 i18n 메시지를 로드합니다.

    Args:
        locale: Language code (ko, en, ja, zh, th)

    Returns:
        Dictionary of translated messages

    Raises:
        FileNotFoundError: If message file not found
        json.JSONDecodeError: If JSON parsing fails

    Example:
        >>> messages = load_messages("en")
        >>> messages["session_start"]
        "🚀 MoAI-ADK Session Started"
    """
    # Supported locales
    supported = ["ko", "en", "ja", "zh", "th"]
    if locale not in supported:
        locale = "en"  # Fallback to English

    # Load message file
    i18n_dir = Path(__file__).parent.parent / "docs" / "i18n"
    message_file = i18n_dir / f"{locale}.json"

    if not message_file.exists():
        raise FileNotFoundError(f"Message file not found: {message_file}")

    return json.loads(message_file.read_text(encoding="utf-8"))

def t(key: str, locale: str = "ko", **kwargs) -> str:
    """
    Translate a message key.

    메시지 키를 번역합니다.

    Args:
        key: Message key (e.g., "session_start")
        locale: Language code
        **kwargs: Format variables for message interpolation

    Returns:
        Translated and formatted message

    Example:
        >>> t("session_start", "en")
        "🚀 MoAI-ADK Session Started"

        >>> t("checkpoint_created", "ko", name="before-merge")
        "🛡️ 체크포인트 생성: before-merge"
    """
    messages = load_messages(locale)
    message = messages.get(key, key)  # Fallback to key itself

    # Format with variables
    if kwargs:
        return message.format(**kwargs)
    return message

def get_supported_locales() -> list[str]:
    """
    Get list of supported locales.

    지원되는 locale 목록을 반환합니다.

    Returns:
        List of locale codes

    Example:
        >>> get_supported_locales()
        ['ko', 'en', 'ja', 'zh', 'th']
    """
    return ["ko", "en", "ja", "zh", "th"]
```

### 4.2 메시지 파일 스키마

```json
// docs/i18n/en.json (Example)
{
  // Session messages
  "session_start": "🚀 MoAI-ADK Session Started",
  "language": "Language",
  "branch": "Branch",
  "changes": "Changes",
  "spec_progress": "SPEC Progress",
  "checkpoints": "Checkpoints",
  "restore_hint": "Restore: /alfred:0-project restore",

  // Checkpoint messages
  "checkpoint_created": "🛡️ Checkpoint created: {name}",
  "checkpoint_operation": "Operation",
  "checkpoint_list": "Available checkpoints",

  // Context messages
  "context_loaded": "📎 Loaded {count} context file(s)",
  "context_recommendation": "💡 Recommended documents",

  // Todo messages
  "todo": {
    "analyzing_docs": "Analyzing project documents",
    "proposing_specs": "Proposing SPEC candidates",
    "writing_spec": "Writing SPEC document",
    "creating_branch": "Creating Git branch",
    "creating_pr": "Creating Draft PR",
    "running_tests": "Running tests",
    "implementing_code": "Implementing code",
    "syncing_docs": "Syncing documentation"
  },

  // Error messages
  "error": {
    "no_git": "❌ Not a Git repository",
    "no_config": "❌ .moai/config.json not found",
    "permission_denied": "❌ Permission denied: {path}",
    "locale_not_supported": "⚠️ Locale '{locale}' not supported, falling back to English",
    "missing_translation": "⚠️ Missing translation: {key}"
  },

  // CLI messages
  "cli": {
    "init_start": "Initializing MoAI-ADK project...",
    "init_success": "✅ Project '{name}' initialized successfully",
    "init_select_locale": "Select your preferred language:",
    "doctor_checking": "Checking system requirements...",
    "doctor_passed": "✅ All checks passed",
    "status_summary": "Project Status Summary"
  }
}
```

---

## 5. 구현 계획 (Implementation Plan)

### Phase 1: 기반 구축 (Foundation)

**작업 항목**:
1. ✅ i18n 메시지 파일 생성 (docs/i18n/{ko,en,ja,zh,th}.json)
2. ✅ i18n 로더 모듈 구현 (src/moai_adk/i18n.py)
3. ✅ 단위 테스트 작성 (tests/test_i18n.py)

**검증 기준**:
- 모든 locale에서 메시지 로드 성공
- LRU 캐시 동작 확인
- Fallback 로직 검증 (지원되지 않는 locale → en)

### Phase 2: Hook 메시지 적용 (Hook Integration)

**작업 항목**:
1. ✅ SessionStart 핸들러 i18n 적용 (.claude/hooks/alfred/handlers/session.py)
2. ✅ Checkpoint 핸들러 i18n 적용 (.claude/hooks/alfred/handlers/checkpoint.py)
3. ✅ Context 핸들러 i18n 적용 (.claude/hooks/alfred/handlers/context.py)

**검증 기준**:
- 각 locale에서 Hook 메시지 정상 출력
- locale 변경 후 메시지 변경 확인

### Phase 3: CLI 메시지 적용 (CLI Integration)

**작업 항목**:
1. ✅ `moai-adk init` 명령어 i18n 적용
2. ✅ `moai-adk doctor` 명령어 i18n 적용
3. ✅ `moai-adk status` 명령어 i18n 적용

**검증 기준**:
- CLI 출력 메시지 locale 적용 확인
- 에러 메시지 locale 적용 확인

### Phase 4: 문서 다국어화 (Documentation)

**작업 항목**:
1. ✅ README.md (영어) 작성
2. ✅ README.{ko,ja,zh,th}.md 작성
3. ✅ 템플릿 CLAUDE.md locale 변수 처리
4. ✅ development-guide.md locale 변수 처리

**검증 기준**:
- 각 언어별 README 품질 확인
- 링크 유효성 검증
- 템플릿 변수 치환 확인

### Phase 5: 통합 테스트 (Integration Test)

**작업 항목**:
1. ✅ 전체 워크플로우 테스트 (각 locale별)
2. ✅ locale 전환 테스트
3. ✅ 에러 처리 테스트

---

## 6. 테스트 전략 (Test Strategy)

### 6.1 단위 테스트 (Unit Tests)

```python
# tests/test_i18n.py

# @TEST:I18N-001 | SPEC: SPEC-I18N-001.md

import pytest
from moai_adk.i18n import load_messages, t, get_supported_locales

def test_load_messages_ko():
    """한국어 메시지 로드 테스트"""
    messages = load_messages("ko")
    assert messages["session_start"] == "🚀 MoAI-ADK 세션 시작"

def test_load_messages_en():
    """영어 메시지 로드 테스트"""
    messages = load_messages("en")
    assert messages["session_start"] == "🚀 MoAI-ADK Session Started"

def test_load_messages_unsupported_locale():
    """지원되지 않는 locale은 영어로 대체"""
    messages = load_messages("fr")  # 프랑스어 (미지원)
    assert messages["session_start"] == "🚀 MoAI-ADK Session Started"

def test_translate_with_format():
    """변수 포함 번역 테스트"""
    result = t("checkpoint_created", "ko", name="before-merge")
    assert "before-merge" in result
    assert "체크포인트" in result

def test_get_supported_locales():
    """지원 언어 목록 테스트"""
    locales = get_supported_locales()
    assert locales == ["ko", "en", "ja", "zh", "th"]
```

### 6.2 통합 테스트 (Integration Tests)

```python
# tests/integration/test_i18n_hooks.py

# @TEST:I18N-001 | SPEC: SPEC-I18N-001.md

def test_session_start_hook_ko(tmp_path):
    """SessionStart Hook 한국어 메시지 테스트"""
    # Setup
    create_test_project(tmp_path, locale="ko")

    # Execute
    result = run_hook("SessionStart", cwd=str(tmp_path))

    # Verify
    assert "세션 시작" in result.systemMessage
    assert "개발 언어" in result.systemMessage

def test_session_start_hook_en(tmp_path):
    """SessionStart Hook 영어 메시지 테스트"""
    create_test_project(tmp_path, locale="en")
    result = run_hook("SessionStart", cwd=str(tmp_path))
    assert "Session Started" in result.systemMessage
    assert "Language" in result.systemMessage
```

---

## 7. 문서화 (Documentation)

### 7.1 사용자 가이드

**위치**: `docs/guides/i18n-guide.md`

**내용**:
- 언어 변경 방법
- 지원 언어 목록
- 번역 기여 방법
- 문제 해결 (Troubleshooting)

### 7.2 개발자 가이드

**위치**: `docs/guides/i18n-development.md`

**내용**:
- i18n 로더 API 사용법
- 새 메시지 추가 방법
- 새 언어 추가 방법
- 테스트 방법

### 7.3 번역 기여 가이드

**위치**: `CONTRIBUTING.md` (i18n 섹션)

**내용**:
- 번역 표준 및 규칙
- Pull Request 프로세스
- 번역 검토 기준
- 번역 품질 보증

---

## 8. 제약사항 및 의존성 (Constraints and Dependencies)

### 의존성 (Dependencies)

- **Python 표준 라이브러리**: json, pathlib, functools
- **외부 의존성**: 없음 (pure Python)

### 제약사항 (Constraints)

- **성능**: 메시지 로드 ≤100ms (LRU 캐시로 보장)
- **메모리**: 메시지 파일 캐시 ≤500KB (5개 locale)
- **파일 크기**: 각 메시지 파일 ≤50KB
- **호환성**: Python 3.10 이상

---

## 9. 보안 고려사항 (Security Considerations)

### 입력 검증 (Input Validation)

- locale 값은 화이트리스트 검증 (ko, en, ja, zh, th만 허용)
- JSON 파싱 오류 시 안전한 fallback

### 파일 접근 제어 (File Access Control)

- 메시지 파일은 읽기 전용으로 접근
- 사용자 입력으로 파일 경로 조작 불가

---

## 10. 마이그레이션 계획 (Migration Plan)

### 기존 코드 영향 (Impact)

**변경 필요**:
- ✅ Hook 핸들러 (session.py, checkpoint.py, context.py)
- ✅ CLI 명령어 (main.py)

**변경 불필요**:
- ✅ 핵심 비즈니스 로직 (installer, template processor 등)
- ✅ 테스트 코드 (새 테스트만 추가)

### 마이그레이션 단계

1. **Phase 1**: i18n 시스템 구축 (영향 없음)
2. **Phase 2**: Hook 메시지 전환 (점진적 적용)
3. **Phase 3**: CLI 메시지 전환 (점진적 적용)
4. **Phase 4**: 문서 분리 (새 파일 생성)

---

## 11. 성공 지표 (Success Criteria)

### 필수 조건 (Must Have)

- ✅ 5개 언어 메시지 파일 완성도 100%
- ✅ 모든 Hook 메시지 i18n 적용
- ✅ 모든 CLI 메시지 i18n 적용
- ✅ README 5개 언어 버전 완성
- ✅ 단위 테스트 커버리지 ≥85%

### 측정 지표 (Metrics)

- **번역 커버리지**: 메시지 키 100% 번역 (5개 언어)
- **성능**: 메시지 로드 시간 ≤100ms
- **사용자 만족도**: GitHub Discussions 피드백 긍정적 평가 ≥80%

---

## 12. 향후 확장 (Future Enhancements)

### v0.1.0 (현재 SPEC)
- ✅ 5개 언어 지원 (ko, en, ja, zh, th)
- ✅ 정적 메시지 파일

### v0.2.0 (향후 계획)
- 🔮 동적 언어 전환 API
- 🔮 번역 자동화 도구 (AI 기반)
- 🔮 웹 기반 번역 관리 도구

### v1.0.0 (장기 비전)
- 🔮 10개 이상 언어 지원
- 🔮 커뮤니티 번역 플랫폼
- 🔮 실시간 번역 품질 검증

---

## 참고 문서 (References)

- [Anthropic Claude Code Documentation](https://docs.claude.com/)
- [i18n Best Practices](https://www.w3.org/International/questions/qa-i18n)
- [Python gettext Documentation](https://docs.python.org/3/library/gettext.html)
- [JSON Schema for i18n](https://json-schema.org/)

---

_이 SPEC은 `/alfred:2-build I18N-001` 명령으로 TDD 구현을 시작할 수 있습니다._
