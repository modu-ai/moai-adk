"""
MoAI Menu Project - Language Initializer Module

Comprehensive language detection, configuration, and localization system.
Integrates patterns from moai-project-language-initializer skill with
advanced multilingual support and locale management.
"""

import json
import locale
import os
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional


class LanguageInitializer:
    """Comprehensive language and localization management system."""

    # Language configuration mappings
    LANGUAGE_CONFIG = {
        "en": {
            "name": "English",
            "native_name": "English",
            "code": "en",
            "locale": "en_US.UTF-8",
            "agent_prompt_language": "english",
            "documentation_language": "en",
            "rtl": False,
            "date_format": "%Y-%m-%d",
            "time_format": "%H:%M:%S",
        },
        "ko": {
            "name": "Korean",
            "native_name": "한국어",
            "code": "ko",
            "locale": "ko_KR.UTF-8",
            "agent_prompt_language": "localized",
            "documentation_language": "ko",
            "rtl": False,
            "date_format": "%Y년 %m월 %d일",
            "time_format": "%H:%M:%S",
        },
        "ja": {
            "name": "Japanese",
            "native_name": "日本語",
            "code": "ja",
            "locale": "ja_JP.UTF-8",
            "agent_prompt_language": "localized",
            "documentation_language": "ja",
            "rtl": False,
            "date_format": "%Y年%m月%d日",
            "time_format": "%H:%M:%S",
        },
        "zh": {
            "name": "Chinese",
            "native_name": "中文",
            "code": "zh",
            "locale": "zh_CN.UTF-8",
            "agent_prompt_language": "localized",
            "documentation_language": "zh",
            "rtl": False,
            "date_format": "%Y年%m月%d日",
            "time_format": "%H:%M:%S",
        },
    }

    # Domain-specific language mappings
    DOMAIN_LANGUAGES = {
        "backend": ["en", "ko"],
        "frontend": ["en", "ko", "ja"],
        "mobile": ["en", "ko", "ja", "zh"],
        "devops": ["en"],
        "data_science": ["en", "ko", "zh"],
        "ai_ml": ["en", "ko", "ja", "zh"],
        "security": ["en"],
        "testing": ["en", "ko"],
        "documentation": ["en", "ko", "ja", "zh"],
    }

    def __init__(self, project_root: str, config: Dict[str, Any]):
        self.project_root = Path(project_root)
        self.config = config
        self.config_path = self.project_root / ".moai/config/config.yaml"
        self._ensure_config_exists()

    def _ensure_config_exists(self):
        """Ensure config file exists with basic structure."""
        self.config_path.parent.mkdir(parents=True, exist_ok=True)

        if not self.config_path.exists():
            default_config = {
                "language": {
                    "conversation_language": "en",
                    "conversation_language_name": "English",
                    "agent_prompt_language": "english",
                    "documentation_language": "en",
                },
                "user": {
                    "name": "Developer",
                    "selected_at": datetime.now().isoformat(),
                },
                "project": {"name": "My Project", "type": "web_application"},
            }
            self.config_path.write_text(
                json.dumps(default_config, indent=2, ensure_ascii=False),
                encoding="utf-8",
            )

    def detect_project_language(self) -> str:
        """
        Auto-detect project language based on project structure and content.

        Returns:
            Detected language code (en, ko, ja, zh)
        """

        # Method 1: Check existing config
        if self.config_path.exists():
            existing_config = json.loads(self.config_path.read_text(encoding="utf-8"))
            existing_lang = existing_config.get("language", {}).get("conversation_language")
            if existing_lang:
                return existing_lang

        # Method 2: Analyze file content
        detected_language = self._analyze_file_content()
        if detected_language:
            return detected_language

        # Method 3: Check system locale
        system_language = self._get_system_language()
        if system_language in self.LANGUAGE_CONFIG:
            return system_language

        # Method 4: Default to English
        return "en"

    def _analyze_file_content(self) -> Optional[str]:
        """Analyze project files to detect primary language."""

        # Check for language indicators in common files
        indicators = {
            "ko": ["한국어", "한글", "ko_KR", ".ko."],
            "ja": ["日本語", "ja_JP", ".ja."],
            "zh": ["中文", "简体中文", "zh_CN", ".zh."],
            "en": ["English", "en_US", ".en."],
        }

        # Search in README files
        for readme_file in self.project_root.glob("README*"):
            if readme_file.is_file():
                content = readme_file.read_text(encoding="utf-8", errors="ignore")
                for lang_code, lang_indicators in indicators.items():
                    for indicator in lang_indicators:
                        if indicator in content:
                            return lang_code

        # Search in package.json
        package_json = self.project_root / "package.json"
        if package_json.exists():
            try:
                package_data = json.loads(package_json.read_text(encoding="utf-8"))
                # Check for language-specific fields
                if "name" in package_data:
                    name = package_data["name"]
                    if any(char in name for char in "가나다라마바사"):
                        return "ko"
                    elif any(char in name for char in "あいうえお"):
                        return "ja"
                    elif any(char in name for char in "中文测试"):
                        return "zh"
            except (json.JSONDecodeError, UnicodeDecodeError):
                pass

        # Search in source code comments
        src_dir = self.project_root / "src"
        if src_dir.exists():
            for file_path in src_dir.rglob("*"):
                if file_path.is_file() and file_path.suffix in [
                    ".py",
                    ".js",
                    ".ts",
                    ".java",
                    ".cpp",
                ]:
                    try:
                        content = file_path.read_text(encoding="utf-8", errors="ignore")
                        for lang_code, lang_indicators in indicators.items():
                            for indicator in lang_indicators:
                                if indicator in content:
                                    return lang_code
                    except UnicodeDecodeError:
                        continue

        return None

    def _get_system_language(self) -> str:
        """Get system default language."""

        try:
            # Try to get system locale
            system_locale = locale.getdefaultlocale()
            if system_locale and system_locale[0]:
                lang_code = system_locale[0].split("_")[0]
                if lang_code in self.LANGUAGE_CONFIG:
                    return lang_code
        except (ValueError, IndexError):
            pass

        # Try environment variables
        env_lang = os.environ.get("LANG", "").split(".")[0].split("_")[0]
        if env_lang in self.LANGUAGE_CONFIG:
            return env_lang

        return "en"

    def initialize_language_configuration(
        self, language: str = None, user_name: str = None, domains: List[str] = None
    ) -> Dict[str, Any]:
        """
        Initialize comprehensive language configuration.

        Args:
            language: Target language code (auto-detected if not provided)
            user_name: User name for personalization
            domains: List of selected domains

        Returns:
            Configuration results and updated settings
        """

        # Auto-detect language if not provided
        if not language:
            language = self.detect_project_language()

        # Validate language
        if language not in self.LANGUAGE_CONFIG:
            language = "en"  # Fallback

        lang_config = self.LANGUAGE_CONFIG[language]

        # Load existing config
        current_config = {}
        if self.config_path.exists():
            current_config = json.loads(self.config_path.read_text(encoding="utf-8"))

        # Update language settings
        current_config.setdefault("language", {}).update(
            {
                "conversation_language": language,
                "conversation_language_name": lang_config["native_name"],
                "agent_prompt_language": lang_config["agent_prompt_language"],
                "documentation_language": lang_config["documentation_language"],
                "locale": lang_config["locale"],
                "rtl": lang_config["rtl"],
                "date_format": lang_config["date_format"],
                "time_format": lang_config["time_format"],
                "initialized_at": datetime.now().isoformat(),
            }
        )

        # Update user settings
        if user_name:
            current_config.setdefault("user", {}).update({"name": user_name, "selected_at": datetime.now().isoformat()})

        # Update domain settings if provided
        if domains:
            # Validate domains for selected language
            valid_domains = []
            for domain in domains:
                if language in self.DOMAIN_LANGUAGES.get(domain, ["en"]):
                    valid_domains.append(domain)

            current_config.setdefault("project", {}).update(
                {
                    "selected_domains": valid_domains,
                    "domain_selection_date": datetime.now().isoformat(),
                }
            )

        # Generate localized templates
        templates = self._generate_localized_templates(language)

        # Save updated configuration
        self.config_path.write_text(json.dumps(current_config, indent=2, ensure_ascii=False), encoding="utf-8")

        return {
            "language": language,
            "language_config": lang_config,
            "updated_config": current_config,
            "templates": templates,
            "domains_supported": self._get_supported_domains(language),
            "token_cost_impact": self._calculate_token_impact(language),
        }

    def _generate_localized_templates(self, language: str) -> Dict[str, str]:
        """Generate language-specific templates and prompts."""

        templates = {
            "welcome_message": "",
            "error_messages": {},
            "success_messages": {},
            "prompts": {},
        }

        if language == "ko":
            templates.update(
                {
                    "welcome_message": "MoAI 프로젝트에 오신 것을 환영합니다!",
                    "error_messages": {
                        "config_load_error": "구성 파일을 로드할 수 없습니다.",
                        "invalid_language": "지원되지 않는 언어입니다.",
                        "file_not_found": "파일을 찾을 수 없습니다.",
                    },
                    "success_messages": {
                        "config_saved": "구성이 성공적으로 저장되었습니다.",
                        "project_initialized": "프로젝트가 초기화되었습니다.",
                        "language_set": "언어가 한국어로 설정되었습니다.",
                    },
                    "prompts": {
                        "select_language": "사용할 언어를 선택해주세요:",
                        "enter_name": "사용자 이름을 입력해주세요:",
                        "select_domains": "작업할 도메인을 선택해주세요:",
                    },
                }
            )
        elif language == "ja":
            templates.update(
                {
                    "welcome_message": "MoAIプロジェクトへようこそ！",
                    "error_messages": {
                        "config_load_error": "設定ファイルを読み込めませんでした。",
                        "invalid_language": "サポートされていない言語です。",
                        "file_not_found": "ファイルが見つかりません。",
                    },
                    "success_messages": {
                        "config_saved": "設定が正常に保存されました。",
                        "project_initialized": "プロジェクトが初期化されました。",
                        "language_set": "言語が日本語に設定されました。",
                    },
                    "prompts": {
                        "select_language": "使用する言語を選択してください：",
                        "enter_name": "ユーザー名を入力してください：",
                        "select_domains": "作業するドメインを選択してください：",
                    },
                }
            )
        elif language == "zh":
            templates.update(
                {
                    "welcome_message": "欢迎使用 MoAI 项目！",
                    "error_messages": {
                        "config_load_error": "无法加载配置文件。",
                        "invalid_language": "不支持的语言。",
                        "file_not_found": "找不到文件。",
                    },
                    "success_messages": {
                        "config_saved": "配置已成功保存。",
                        "project_initialized": "项目已初始化。",
                        "language_set": "语言已设置为中文。",
                    },
                    "prompts": {
                        "select_language": "请选择使用的语言：",
                        "enter_name": "请输入用户名：",
                        "select_domains": "请选择工作领域：",
                    },
                }
            )
        else:  # English (default)
            templates.update(
                {
                    "welcome_message": "Welcome to MoAI Project!",
                    "error_messages": {
                        "config_load_error": "Cannot load configuration file.",
                        "invalid_language": "Unsupported language.",
                        "file_not_found": "File not found.",
                    },
                    "success_messages": {
                        "config_saved": "Configuration saved successfully.",
                        "project_initialized": "Project initialized successfully.",
                        "language_set": "Language set to English.",
                    },
                    "prompts": {
                        "select_language": "Please select your preferred language:",
                        "enter_name": "Please enter your name:",
                        "select_domains": "Please select your work domains:",
                    },
                }
            )

        return templates

    def _get_supported_domains(self, language: str) -> List[str]:
        """Get list of domains supported for the given language."""

        supported = []
        for domain, langs in self.DOMAIN_LANGUAGES.items():
            if language in langs:
                supported.append(domain)

        return sorted(supported)

    def _calculate_token_impact(self, language: str) -> Dict[str, Any]:
        """Calculate token cost impact for the selected language."""

        impact = {
            "language": language,
            "conversation_overhead": 0,
            "agent_prompt_overhead": 0,
            "documentation_overhead": 0,
            "total_overhead_percentage": 0,
            "recommendations": [],
        }

        if language == "en":
            impact.update(
                {
                    "conversation_overhead": 0,
                    "agent_prompt_overhead": 0,
                    "documentation_overhead": 0,
                    "total_overhead_percentage": 0,
                    "recommendations": [
                        "Most token-efficient choice",
                        "Widely supported documentation and libraries",
                    ],
                }
            )
        else:
            # Non-English languages typically require 15-25% more tokens
            overhead_percentage = 20

            impact.update(
                {
                    "conversation_overhead": overhead_percentage,
                    "agent_prompt_overhead": (
                        15
                        if self.LANGUAGE_CONFIG[language]["agent_prompt_language"] == "english"
                        else overhead_percentage
                    ),
                    "documentation_overhead": overhead_percentage,
                    "total_overhead_percentage": overhead_percentage,
                    "recommendations": [
                        f"Expect ~{overhead_percentage}% increase in token usage",
                        "Consider using English for agent prompts to reduce costs",
                        "Localized prompts provide better user experience",
                    ],
                }
            )

        return impact

    def localize_agent_prompts(self, prompt_template: str, language: str = None) -> str:
        """
        Localize agent prompts based on language configuration.

        Args:
            prompt_template: Base prompt template
            language: Target language (uses config default if not provided)

        Returns:
            Localized prompt string
        """

        if not language:
            language = self.detect_project_language()

        # Load current configuration
        if self.config_path.exists():
            config = json.loads(self.config_path.read_text(encoding="utf-8"))
            agent_lang = config.get("language", {}).get("agent_prompt_language", "english")
        else:
            agent_lang = "english"

        # If agent prompts should be in English, return as-is
        if agent_lang == "english":
            return prompt_template

        # Apply language-specific localizations
        localized_prompt = prompt_template

        if language == "ko":
            # Add Korean-specific instructions
            korean_additions = (
                "\n\nIMPORTANT: Please respond in Korean (한국어) "
                "when interacting with the user. Use formal polite form (합니다/입니다 style)."
            )
            localized_prompt += korean_additions

        elif language == "ja":
            # Add Japanese-specific instructions
            japanese_additions = (
                "\n\nIMPORTANT: Please respond in Japanese (日本語) "
                "when interacting with the user. Use polite form (です/ます style)."
            )
            localized_prompt += japanese_additions

        elif language == "zh":
            # Add Chinese-specific instructions
            chinese_additions = (
                "\n\nIMPORTANT: Please respond in Chinese (中文) when interacting with the user. Use standard Mandarin."
            )
            localized_prompt += chinese_additions

        return localized_prompt

    def create_multilingual_documentation_structure(self, language: str = None) -> Dict[str, Any]:
        """
        Create documentation structure for multilingual support.

        Args:
            language: Primary language for documentation

        Returns:
            Documentation structure configuration
        """

        if not language:
            language = self.detect_project_language()

        docs_root = self.project_root / "docs"

        # Create language-specific directories
        lang_dirs = {
            "primary": docs_root / language,
            "en": docs_root / "en",  # Always include English as fallback
        }

        # Create additional language directories if needed
        for lang_code in ["ko", "ja", "zh"]:
            if lang_code != language:
                lang_dirs[lang_code] = docs_root / lang_code

        # Ensure directories exist
        for lang_dir in lang_dirs.values():
            lang_dir.mkdir(parents=True, exist_ok=True)

        # Create language index
        language_index = {
            "primary_language": language,
            "supported_languages": list(lang_dirs.keys()),
            "directory_structure": {k: str(v) for k, v in lang_dirs.items()},
            "fallback_language": "en",
            "auto_redirect": True,
        }

        # Save language index
        index_path = docs_root / "languages.json"
        index_path.write_text(json.dumps(language_index, indent=2, ensure_ascii=False), encoding="utf-8")

        # Create .htaccess for language redirection (if web docs)
        htaccess_content = f"""
# Language negotiation
Options +MultiViews
AddLanguage ko .ko
AddLanguage ja .ja
AddLanguage zh .zh
AddLanguage en .en

# Default language
DefaultLanguage {language}

# Force language for specific files
<Files "README">
    ForceType text/html
    Header set Content-Language {language}
</Files>
"""

        htaccess_path = docs_root / ".htaccess"
        htaccess_path.write_text(htaccess_content.strip(), encoding="utf-8")

        return {
            "created_directories": [str(d) for d in lang_dirs.values()],
            "language_index": language_index,
            "primary_language": language,
            "total_languages": len(lang_dirs),
        }

    def validate_language_configuration(self) -> Dict[str, Any]:
        """
        Validate current language configuration.

        Returns:
            Validation results with recommendations
        """

        validation_result = {
            "valid": True,
            "errors": [],
            "warnings": [],
            "recommendations": [],
            "current_config": {},
        }

        if not self.config_path.exists():
            validation_result["errors"].append("Configuration file does not exist")
            validation_result["valid"] = False
            return validation_result

        try:
            current_config = json.loads(self.config_path.read_text(encoding="utf-8"))
            validation_result["current_config"] = current_config

            # Validate language section
            language_config = current_config.get("language", {})

            # Check required fields
            required_fields = [
                "conversation_language",
                "conversation_language_name",
                "agent_prompt_language",
                "documentation_language",
            ]

            for field in required_fields:
                if field not in language_config:
                    validation_result["errors"].append(f"Missing required field: language.{field}")
                    validation_result["valid"] = False

            # Validate language codes
            conv_lang = language_config.get("conversation_language")
            if conv_lang and conv_lang not in self.LANGUAGE_CONFIG:
                validation_result["errors"].append(f"Invalid conversation language: {conv_lang}")
                validation_result["valid"] = False

            # Check agent prompt language consistency
            agent_lang = language_config.get("agent_prompt_language")
            if agent_lang and agent_lang not in ["english", "localized"]:
                validation_result["warnings"].append(f"Unusual agent prompt language: {agent_lang}")

            # Check for optimization opportunities
            if conv_lang != "en" and agent_lang == "localized":
                validation_result["recommendations"].append(
                    "Consider using English for agent prompts to reduce token costs by ~15%"
                )

            # Validate user section
            user_config = current_config.get("user", {})
            if "name" not in user_config:
                validation_result["recommendations"].append("Set user name for personalized experience")

        except json.JSONDecodeError as e:
            validation_result["errors"].append(f"Invalid JSON in configuration: {e}")
            validation_result["valid"] = False

        return validation_result

    def update_language_settings(self, updates: Dict[str, Any]) -> Dict[str, Any]:
        """
        Update specific language settings.

        Args:
            updates: Dictionary of settings to update

        Returns:
            Update results
        """

        if not self.config_path.exists():
            return {"success": False, "error": "Configuration file does not exist"}

        try:
            # Load current config
            current_config = json.loads(self.config_path.read_text(encoding="utf-8"))

            # Apply updates
            for key_path, value in updates.items():
                keys = key_path.split(".")
                target = current_config

                for key in keys[:-1]:
                    target = target.setdefault(key, {})

                target[keys[-1]] = value

            # Save updated config
            self.config_path.write_text(
                json.dumps(current_config, indent=2, ensure_ascii=False),
                encoding="utf-8",
            )

            return {
                "success": True,
                "updated_config": current_config,
                "updates_applied": updates,
            }

        except Exception as e:
            return {"success": False, "error": str(e)}

    def get_language_status(self) -> Dict[str, Any]:
        """
        Get comprehensive language configuration status.

        Returns:
            Language status and metrics
        """

        status = {
            "detected_language": self.detect_project_language(),
            "configured_language": None,
            "system_language": self._get_system_language(),
            "supported_languages": list(self.LANGUAGE_CONFIG.keys()),
            "configuration_exists": self.config_path.exists(),
            "token_impact": {},
            "domain_support": {},
        }

        # Load current configuration if exists
        if self.config_path.exists():
            try:
                config = json.loads(self.config_path.read_text(encoding="utf-8"))
                lang_config = config.get("language", {})
                status["configured_language"] = lang_config.get("conversation_language")

                # Calculate token impact
                if status["configured_language"]:
                    status["token_impact"] = self._calculate_token_impact(status["configured_language"])

                # Get domain support
                if status["configured_language"]:
                    status["domain_support"] = {
                        domain: status["configured_language"] in langs
                        for domain, langs in self.DOMAIN_LANGUAGES.items()
                    }

            except json.JSONDecodeError:
                status["configuration_error"] = "Invalid JSON in configuration file"

        return status
