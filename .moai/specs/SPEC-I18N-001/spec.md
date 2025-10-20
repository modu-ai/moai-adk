---
id: I18N-001
version: 0.0.1
status: draft
created: 2025-10-20
updated: 2025-10-20
author: @Goos
priority: high
category: feature
labels:
  - i18n
  - template
  - multilingual
scope:
  packages:
    - src/moai_adk/templates/
    - src/moai_adk/core/init/
    - src/moai_adk/core/template/
---

# @SPEC:I18N-001: 다국어 템플릿 시스템 (한/영)

## HISTORY

### v0.0.1 (2025-10-20)
- **INITIAL**: 2개 언어(한국어/영어) 템플릿 시스템 명세 작성
- **AUTHOR**: @Goos
- **REASON**: 템플릿만 다국어화하여 프로젝트 초기화 시 선택한 언어로 문서 제공
- **SCOPE**: CLI는 영어 유지, 템플릿만 한/영 분리, README.md 한국어 메인

---

## 1. 개요 (Overview)

### 목적 (Purpose)

MoAI-ADK 템플릿을 한국어와 영어로 분리하여, 사용자가 프로젝트 초기화 시 원하는 언어의 템플릿을 선택할 수 있도록 합니다.

**핵심 원칙**:
- **CLI/코드는 영어 유지**: 모든 명령어, 코드, 로그는 영어로 유지
- **템플릿만 다국어화**: `.moai/`, `.claude/` 등 프로젝트 템플릿 파일만 다국어 제공
- **간단한 구조**: 2개 언어만 지원 (한국어, 영어)
- **사용자 선택**: `moai-adk init` 실행 시 언어 선택

### 범위 (Scope)

**포함 사항**:
- ✅ 템플릿 디렉토리 분리 (`.claude/`, `.claude-en/`, `.claude-ko/`)
- ✅ init.py 수정 (locale 선택 ko/en만)
- ✅ processor.py 수정 (언어별 템플릿 복사)
- ✅ README.md 구조 변경 (메인: 한국어, 영어는 README.en.md)

**제외 사항**:
- ❌ CLI 메시지 다국어화 (영어 고정)
- ❌ 코드 주석 번역 (영어 유지)
- ❌ 런타임 다국어 전환 (향후 확장)
- ❌ 3개 이상 언어 지원 (v0.1.0 범위 외)

---

## 2. EARS 요구사항 (Requirements)

### Ubiquitous Requirements (기본 요구사항)

1. **시스템은 2개 언어(한국어, 영어) 템플릿을 제공해야 한다**
   - 한국어 템플릿: `.claude-ko/`
   - 영어 템플릿: `.claude-en/`
   - 구조는 동일하고 내용만 번역

2. **시스템은 CLI 명령어 및 코드를 영어로 유지해야 한다**
   - 모든 명령어: `moai-adk init`, `moai-adk doctor` 등
   - 모든 로그 및 에러 메시지: 영어
   - 코드 내부 주석: 영어

3. **시스템은 README.md를 한국어로 작성하고 영어는 별도 파일로 제공해야 한다**
   - 메인: `README.md` (한국어)
   - 영어: `README.en.md` (링크 제공)

### Event-driven Requirements (이벤트 기반)

1. **WHEN 사용자가 `moai-adk init` 실행 시, 시스템은 언어 선택 프롬프트를 표시해야 한다**
   - 선택지: "Korean (한국어)" / "English"
   - 기본값: Korean (한국어)

2. **WHEN 사용자가 언어를 선택하면, 시스템은 해당 템플릿 디렉토리를 복사해야 한다**
   - 한국어 선택: `.claude-ko/` → `.claude/`
   - 영어 선택: `.claude-en/` → `.claude/`

3. **WHEN 템플릿이 복사되면, 시스템은 `.moai/config.json`에 locale 값을 저장해야 한다**
   - `{ "project": { "locale": "ko" | "en" } }`

### State-driven Requirements (상태 기반)

1. **WHILE 템플릿 복사 중일 때, 시스템은 영어 메시지만 출력해야 한다**
   - "Copying template files..."
   - "Template initialization complete"

2. **WHILE CLI가 실행 중일 때, 시스템은 항상 영어 메시지를 사용해야 한다**
   - 로그, 에러, 경고 모두 영어

### Constraints (제약사항)

1. **IF 지원되지 않는 locale이 제공되면, 시스템은 영어(en)로 대체해야 한다**
   - 예: `locale: "ja"` → 영어로 대체 + 경고 메시지

2. **템플릿 구조는 동일해야 한다**
   - `.claude-ko/`와 `.claude-en/`는 파일 이름, 디렉토리 구조가 동일
   - 내용만 번역

3. **CLI 명령어는 변경하지 않아야 한다**
   - `moai-adk init`, `moai-adk doctor` 등 모두 영어 유지

---

## 3. 아키텍처 (Architecture)

### 템플릿 디렉토리 구조

```
src/moai_adk/templates/
├── .claude/              # ❌ 제거: 기존 통합 템플릿
├── .claude-en/           # ✅ 신규: 영어 템플릿
│   ├── commands/
│   │   ├── 1-spec.md
│   │   ├── 2-build.md
│   │   └── 3-sync.md
│   ├── agents/
│   │   ├── spec-builder.md
│   │   ├── code-builder.md
│   │   └── doc-syncer.md
│   └── README.md         # English agent docs
│
├── .claude-ko/           # ✅ 신규: 한국어 템플릿 (기존 .claude/ 이동)
│   ├── commands/
│   │   ├── 1-spec.md
│   │   ├── 2-build.md
│   │   └── 3-sync.md
│   ├── agents/
│   │   ├── spec-builder.md
│   │   ├── code-builder.md
│   │   └── doc-syncer.md
│   └── README.md         # 한국어 에이전트 문서
│
└── .moai/
    ├── config.json       # locale: "ko" | "en" 추가
    ├── project/
    │   ├── product.md    # 템플릿 (변수 치환)
    │   ├── structure.md
    │   └── tech.md
    └── memory/
        ├── development-guide.md
        └── spec-metadata.md
```

### 워크플로우

```
사용자: moai-adk init
    ↓
CLI: "Select language: [Korean (한국어) / English]"
    ↓
사용자 선택: "Korean (한국어)"
    ↓
TemplateProcessor:
    1. 선택: .claude-ko/
    2. 복사: .claude-ko/ → .claude/
    3. 저장: .moai/config.json { locale: "ko" }
    ↓
CLI: "✅ Project initialized with Korean template"
```

---

## 4. 구현 계획 (Implementation Plan)

### Phase 1: 템플릿 분리

**작업 항목**:
1. 기존 `.claude/` → `.claude-ko/`로 이동
2. `.claude-en/` 생성 및 영어 번역
3. 템플릿 구조 동일성 검증

**검증 기준**:
- 디렉토리 구조 일치 확인
- 파일 개수 일치 확인
- 필수 파일 존재 확인

### Phase 2: init.py 수정

**작업 항목**:
1. locale 선택 프롬프트 추가 (ko/en만)
2. 선택된 locale 저장 로직 추가
3. 기본값: "ko"

**변경 파일**:
- `src/moai_adk/core/init/commands.py`

**검증 기준**:
- 언어 선택 프롬프트 표시 확인
- 선택 후 `.moai/config.json` 저장 확인

### Phase 3: processor.py 수정

**작업 항목**:
1. locale 기반 템플릿 경로 선택 로직 추가
2. `.claude-{locale}/` → `.claude/` 복사
3. 에러 처리 (템플릿 누락 시)

**변경 파일**:
- `src/moai_adk/core/template/processor.py`

**검증 기준**:
- 한국어 선택 시 `.claude-ko/` 복사 확인
- 영어 선택 시 `.claude-en/` 복사 확인
- 템플릿 누락 시 명확한 에러 메시지 확인

### Phase 4: README.md 구조 변경

**작업 항목**:
1. 기존 `README.md` → 한국어로 작성
2. 영어 번역 → `README.en.md`
3. 메인 README에 영어 링크 추가

**검증 기준**:
- README.md 한국어 완성도 확인
- README.en.md 영어 완성도 확인
- 링크 유효성 확인

---

## 5. 인터페이스 설계 (Interface Design)

### 5.1 CLI 언어 선택 프롬프트

```bash
$ moai-adk init

? Select your preferred language for project templates:
  > Korean (한국어)
    English

# 선택 후
✅ Project initialized with Korean template
✅ Template files copied to .claude/
```

### 5.2 config.json 스키마

```json
{
  "project": {
    "name": "my-project",
    "description": "Project description",
    "version": "0.1.0",
    "mode": "personal",
    "locale": "ko"  // ← NEW: "ko" | "en"
  }
}
```

### 5.3 TemplateProcessor API 변경

```python
# src/moai_adk/core/template/processor.py

class TemplateProcessor:
    def copy_claude_template(self, locale: str = "ko") -> None:
        """
        Copy Claude Code template based on locale.

        Args:
            locale: Language code ("ko" or "en")

        Raises:
            FileNotFoundError: If template directory not found
        """
        # Select template directory
        if locale not in ["ko", "en"]:
            logger.warning(f"Unsupported locale '{locale}', falling back to 'en'")
            locale = "en"

        template_dir = self.template_root / f".claude-{locale}"
        if not template_dir.exists():
            raise FileNotFoundError(f"Template directory not found: {template_dir}")

        # Copy to .claude/
        dest_dir = self.project_root / ".claude"
        shutil.copytree(template_dir, dest_dir, dirs_exist_ok=True)
        logger.info(f"✅ Template copied: .claude-{locale}/ → .claude/")
```

---

## 6. 테스트 전략 (Test Strategy)

### 6.1 단위 테스트

```python
# tests/test_i18n_template.py

# @TEST:I18N-001 | SPEC: SPEC-I18N-001.md

import pytest
from moai_adk.core.template.processor import TemplateProcessor

def test_copy_claude_template_korean(tmp_path):
    """한국어 템플릿 복사 테스트"""
    processor = TemplateProcessor(tmp_path)
    processor.copy_claude_template("ko")

    # 검증
    claude_dir = tmp_path / ".claude"
    assert claude_dir.exists()
    assert (claude_dir / "commands" / "1-spec.md").exists()

def test_copy_claude_template_english(tmp_path):
    """영어 템플릿 복사 테스트"""
    processor = TemplateProcessor(tmp_path)
    processor.copy_claude_template("en")

    # 검증
    claude_dir = tmp_path / ".claude"
    assert claude_dir.exists()
    assert (claude_dir / "commands" / "1-spec.md").exists()

def test_copy_claude_template_fallback_to_english(tmp_path):
    """지원되지 않는 locale은 영어로 대체"""
    processor = TemplateProcessor(tmp_path)
    processor.copy_claude_template("ja")  # 일본어 (미지원)

    # 검증: 영어 템플릿 복사됨
    claude_dir = tmp_path / ".claude"
    assert claude_dir.exists()
```

### 6.2 통합 테스트

```python
# tests/integration/test_init_i18n.py

# @TEST:I18N-001 | SPEC: SPEC-I18N-001.md

def test_init_with_korean_locale(cli_runner, tmp_path):
    """moai-adk init 한국어 선택 통합 테스트"""
    result = cli_runner.invoke(["init", "--locale", "ko"], cwd=str(tmp_path))

    # 검증
    assert result.exit_code == 0
    assert "Korean template" in result.output

    config = (tmp_path / ".moai" / "config.json").read_text()
    assert '"locale": "ko"' in config

def test_init_with_english_locale(cli_runner, tmp_path):
    """moai-adk init 영어 선택 통합 테스트"""
    result = cli_runner.invoke(["init", "--locale", "en"], cwd=str(tmp_path))

    # 검증
    assert result.exit_code == 0
    assert "English template" in result.output

    config = (tmp_path / ".moai" / "config.json").read_text()
    assert '"locale": "en"' in config
```

---

## 7. 마이그레이션 계획 (Migration Plan)

### 기존 코드 영향 (Impact)

**변경 필요**:
- ✅ `src/moai_adk/core/init/commands.py` (locale 선택 프롬프트)
- ✅ `src/moai_adk/core/template/processor.py` (템플릿 경로 로직)
- ✅ `src/moai_adk/templates/.claude/` → `src/moai_adk/templates/.claude-ko/`

**변경 불필요**:
- ✅ CLI 명령어 (영어 유지)
- ✅ 핵심 비즈니스 로직
- ✅ 테스트 프레임워크

### 마이그레이션 단계

1. **Phase 1**: 템플릿 분리 (기존 사용자 영향 없음)
2. **Phase 2**: init.py 수정 (새 프로젝트만 영향)
3. **Phase 3**: processor.py 수정 (새 프로젝트만 영향)
4. **Phase 4**: README.md 구조 변경 (문서만)

---

## 8. 제약사항 및 의존성 (Constraints and Dependencies)

### 의존성 (Dependencies)

- **Python 표준 라이브러리**: shutil, pathlib
- **외부 의존성**: 없음

### 제약사항 (Constraints)

- **지원 언어**: 한국어(ko), 영어(en) 2개만
- **CLI 언어**: 영어 고정
- **템플릿 구조**: 양쪽 동일 유지
- **호환성**: Python 3.10 이상

---

## 9. 성공 지표 (Success Criteria)

### 필수 조건 (Must Have)

- ✅ 2개 언어 템플릿 완성 (구조 동일)
- ✅ init 실행 시 언어 선택 프롬프트 표시
- ✅ 선택한 언어 템플릿 정상 복사
- ✅ CLI 메시지 영어 유지
- ✅ README.md 한국어 메인, 영어 링크 제공

### 측정 지표 (Metrics)

- **번역 커버리지**: 템플릿 파일 100% 번역
- **CLI 영어 유지**: 모든 명령어 및 로그 영어
- **사용자 만족도**: GitHub Issues 피드백 긍정적 평가

---

## 10. 향후 확장 (Future Enhancements)

### v0.1.0 (현재 SPEC)
- ✅ 2개 언어 지원 (ko, en)
- ✅ 템플릿만 다국어화

### v0.2.0 (향후 계획)
- 🔮 3개 이상 언어 추가 (ja, zh 등)
- 🔮 CLI 메시지 다국어화
- 🔮 런타임 locale 전환 API

### v1.0.0 (장기 비전)
- 🔮 커뮤니티 번역 플랫폼
- 🔮 자동 번역 도구 (AI 기반)

---

## 참고 문서 (References)

- [MoAI-ADK 프로젝트 구조](../../README.md)
- [TemplateProcessor API](../../src/moai_adk/core/template/processor.py)
- [init 명령어 구현](../../src/moai_adk/core/init/commands.py)

---

_이 SPEC은 `/alfred:2-build I18N-001` 명령으로 TDD 구현을 시작할 수 있습니다._
