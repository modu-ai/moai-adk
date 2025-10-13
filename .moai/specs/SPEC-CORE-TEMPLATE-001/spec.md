---
id: CORE-TEMPLATE-001
version: 0.0.1
status: draft
created: 2025-10-13
updated: 2025-10-13
author: @Goos
priority: high
category: feature
labels:
  - template
  - jinja2
  - config
depends_on:
  - PY314-001
scope:
  packages:
    - moai-adk-py/src/moai_adk/core/template/
  files:
    - template/processor.py
    - template/config.py
---

# @SPEC:CORE-TEMPLATE-001: Jinja2 템플릿 및 Config 관리

## HISTORY

### v0.0.1 (2025-10-13)
- **INITIAL**: Jinja2 기반 템플릿 프로세서 및 config.json 관리 시스템 명세 작성
- **AUTHOR**: @Goos
- **REASON**: TypeScript EJS를 Jinja2로 전환

---

## 개요

Jinja2 템플릿 엔진을 사용하여 .moai/ 디렉토리 구조를 생성하고, config.json 파일을 관리한다. 20개 언어별 템플릿을 처리한다.

---

## Environment (환경 및 전제조건)

### 기술 스택
- **템플릿 엔진**: Jinja2 3.1+
- **설정 파일**: JSON (PyYAML로 읽기)
- **템플릿 위치**: templates/.moai/, templates/.claude/

### 기존 시스템
- TypeScript EJS 기반 템플릿
- config.json: mode, locale, git, spec, backup, constitution
- 20개 언어 템플릿 (Python, TypeScript, Java, Go, Rust, 등)

---

## Requirements (요구사항)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 Jinja2 템플릿을 렌더링할 수 있어야 한다
- 시스템은 config.json을 읽고 쓸 수 있어야 한다
- 시스템은 20개 언어 템플릿을 지원해야 한다
- 시스템은 템플릿 변수를 치환할 수 있어야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 템플릿 렌더링 요청이 오면, 시스템은 변수를 치환해야 한다
- WHEN config.json 업데이트 요청이 오면, 시스템은 기존 값을 병합해야 한다
- WHEN 언어별 템플릿 조회 시, 시스템은 해당 언어 파일을 찾아야 한다

### State-driven Requirements (상태 기반)
- WHILE 템플릿 파일이 없을 때, 시스템은 기본 템플릿을 사용해야 한다
- WHILE config.json이 없을 때, 시스템은 기본 설정을 생성해야 한다

### Constraints (제약사항)
- 템플릿 파일은 .j2 확장자를 사용해야 한다
- config.json은 UTF-8 인코딩이어야 한다
- 템플릿 변수는 {{ variable }} 형식이어야 한다

---

## Specifications (상세 명세)

### 1. TemplateProcessor 클래스

```python
# moai_adk/core/template/processor.py
from jinja2 import Environment, FileSystemLoader, select_autoescape
from pathlib import Path

class TemplateProcessor:
    def __init__(self, templates_dir: str = "templates"):
        self.env = Environment(
            loader=FileSystemLoader(templates_dir),
            autoescape=select_autoescape(),
            trim_blocks=True,
            lstrip_blocks=True,
        )

    def render(self, template_path: str, context: dict) -> str:
        """템플릿 렌더링"""
        template = self.env.get_template(template_path)
        return template.render(**context)

    def render_to_file(self, template_path: str, output_path: str, context: dict) -> None:
        """템플릿을 파일로 렌더링"""
        content = self.render(template_path, context)
        Path(output_path).parent.mkdir(parents=True, exist_ok=True)
        Path(output_path).write_text(content, encoding="utf-8")
```

### 2. ConfigManager 클래스

```python
# moai_adk/core/template/config.py
import json
from pathlib import Path
from typing import Any

class ConfigManager:
    DEFAULT_CONFIG = {
        "moai": {"version": "0.3.0"},
        "mode": "personal",
        "projectName": "",
        "features": [],
        "locale": "ko",
        "git": {
            "enabled": True,
            "autoCommit": True,
            "branchPrefix": ""
        },
        "spec": {
            "storage": "local",
            "workflow": "commit",
            "localPath": ".moai/specs/"
        },
        "backup": {
            "enabled": True,
            "retentionDays": 30
        },
        "constitution": {
            "enforce_tdd": True,
            "enforce_spec": False,
            "require_tags": True,
            "test_coverage_target": 85
        }
    }

    def __init__(self, config_path: str = ".moai/config.json"):
        self.config_path = Path(config_path)

    def load(self) -> dict:
        """config.json 읽기"""
        if not self.config_path.exists():
            return self.DEFAULT_CONFIG.copy()

        with open(self.config_path, "r", encoding="utf-8") as f:
            return json.load(f)

    def save(self, config: dict) -> None:
        """config.json 저장"""
        self.config_path.parent.mkdir(parents=True, exist_ok=True)
        with open(self.config_path, "w", encoding="utf-8") as f:
            json.dump(config, f, indent=2, ensure_ascii=False)

    def update(self, updates: dict) -> None:
        """config.json 업데이트 (병합)"""
        config = self.load()
        self._deep_merge(config, updates)
        self.save(config)

    def _deep_merge(self, base: dict, updates: dict) -> None:
        """딕셔너리 깊은 병합"""
        for key, value in updates.items():
            if key in base and isinstance(base[key], dict) and isinstance(value, dict):
                self._deep_merge(base[key], value)
            else:
                base[key] = value
```

### 3. 언어별 템플릿 매핑

```python
LANGUAGE_TEMPLATES = {
    "python": ".moai/project/tech/python.md.j2",
    "typescript": ".moai/project/tech/typescript.md.j2",
    "java": ".moai/project/tech/java.md.j2",
    "go": ".moai/project/tech/go.md.j2",
    "rust": ".moai/project/tech/rust.md.j2",
    # ... 15개 더
}

def get_language_template(language: str) -> str:
    """언어별 템플릿 경로 반환"""
    return LANGUAGE_TEMPLATES.get(language.lower(), ".moai/project/tech/default.md.j2")
```

### 4. 템플릿 변수 예시

```jinja2
{# templates/.moai/config.json.j2 #}
{
  "moai": {
    "version": "{{ version }}"
  },
  "mode": "{{ mode }}",
  "projectName": "{{ project_name }}",
  "locale": "{{ locale }}",
  "git": {
    "enabled": {{ git_enabled | lower }},
    "autoCommit": {{ auto_commit | lower }}
  }
}
```

---

## Traceability (추적성)

- **SPEC ID**: @SPEC:CORE-TEMPLATE-001
- **Depends on**: PY314-001
- **TAG 체인**: @SPEC:CORE-TEMPLATE-001 → @TEST:CORE-TEMPLATE-001 → @CODE:CORE-TEMPLATE-001
