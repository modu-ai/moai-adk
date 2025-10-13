# @CODE:CORE-TEMPLATE-001 | SPEC: SPEC-CORE-TEMPLATE-001.md
"""Template 모듈

Jinja2 템플릿 프로세싱 및 config.json 관리 기능을 제공합니다.

Modules:
    - config: ConfigManager (config.json 읽기/쓰기/업데이트)
    - processor: TemplateProcessor (Jinja2 템플릿 렌더링)
    - languages: 20개 언어 템플릿 매핑

Examples:
    >>> from moai_adk.core.template import ConfigManager, TemplateProcessor
    >>> manager = ConfigManager()
    >>> config = manager.load()
    >>> processor = TemplateProcessor("templates")
    >>> result = processor.render("test.j2", {"name": "World"})
"""

from moai_adk.core.template.config import ConfigManager
from moai_adk.core.template.languages import (
    LANGUAGE_TEMPLATES,
    get_language_template,
)
from moai_adk.core.template.processor import TemplateProcessor

__all__ = [
    "ConfigManager",
    "TemplateProcessor",
    "LANGUAGE_TEMPLATES",
    "get_language_template",
]
