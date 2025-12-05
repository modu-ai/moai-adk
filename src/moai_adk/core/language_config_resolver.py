"""
MoAI-ADK Language Configuration Resolver

Centralized system for resolving language configuration from multiple sources:
1. Environment variables (highest priority)
2. Configuration file (.moai/config/config.json)
3. Default values (fallback)

This module provides a unified interface for all language-related configuration
needs across the MoAI-ADK system.
"""

import json
import os
from pathlib import Path
from typing import Any, Dict, Optional

try:
    import yaml

    YAML_AVAILABLE = True
except ImportError:
    YAML_AVAILABLE = False


class LanguageConfigResolver:
    """
    Resolves language configuration from environment variables and config files
    with proper priority handling and validation.
    """

    # Dynamic language name generation - no fixed language map
    # All languages are supported dynamically based on language codes

    # Default configuration
    DEFAULT_CONFIG = {
        "user_name": "",
        "conversation_language": "en",
        "conversation_language_name": "English",
        "agent_prompt_language": "en",
        "git_commit_messages": "en",
        "code_comments": "en",
        "documentation": "en",
        "error_messages": "en",
        "config_source": "default",
    }

    def __init__(self, project_root: Optional[str] = None):
        """
        Initialize the resolver with project root path.

        Args:
            project_root: Root directory of the project. If None, uses current working directory.
        """
        self.project_root = Path(project_root) if project_root else Path.cwd()

        # Auto-detect YAML (preferred) or JSON (fallback)
        yaml_path = self.project_root / ".moai" / "config" / "config.yaml"
        json_path = self.project_root / ".moai" / "config" / "config.json"

        if YAML_AVAILABLE and yaml_path.exists():
            self.config_file_path = yaml_path
        elif json_path.exists():
            self.config_file_path = json_path
        else:
            # Default to YAML for new projects
            self.config_file_path = yaml_path if YAML_AVAILABLE else json_path

        self._cached_config: Optional[Dict[str, Any]] = None

    def resolve_config(self, force_refresh: bool = False) -> Dict[str, Any]:
        """
        Resolve the complete language configuration from all sources.

        Args:
            force_refresh: Force refresh of cached configuration

        Returns:
            Dictionary containing resolved language configuration
        """
        if self._cached_config is None or force_refresh:
            self._cached_config = self._build_config()

        return self._cached_config.copy()

    def _build_config(self) -> Dict[str, Any]:
        """
        Build configuration by merging sources with proper priority.

        Priority Order:
        1. Environment variables (highest)
        2. Configuration file
        3. Default values (lowest)
        """
        # Start with defaults
        config = self.DEFAULT_CONFIG.copy()

        # Load config file if exists
        file_config = self._load_config_file()
        if file_config:
            config.update(file_config)

        # Override with environment variables (highest priority)
        env_config = self._load_env_config()
        if env_config:
            config.update(env_config)
            config["config_source"] = "environment"
        elif file_config:
            config["config_source"] = "config_file"
        else:
            config["config_source"] = "default"

        # Ensure consistency
        config = self._ensure_consistency(config)

        return config

    def _load_config_file(self) -> Optional[Dict[str, Any]]:
        """
        Load configuration from .moai/config/config.yaml or section files.

        Priority:
        1. Section files (.moai/config/sections/*.yaml) - highest
        2. Main config file (.moai/config/config.yaml)
        3. None if neither exists

        Returns:
            Configuration dictionary or None if file doesn't exist
        """
        config = {}

        # Try loading from section files first (higher priority)
        section_config = self._load_section_files()
        if section_config:
            config.update(section_config)

        # Try loading from main config file (fallback)
        main_config = self._load_main_config_file()
        if main_config:
            # Only update fields not already set by section files
            for key, value in main_config.items():
                if key not in config:
                    config[key] = value

        return config if config else None

    def _load_section_files(self) -> Optional[Dict[str, Any]]:
        """
        Load configuration from section files in .moai/config/sections/

        Section files:
        - user.yaml: User identification
        - language.yaml: Language preferences

        Returns:
            Configuration dictionary or None if section files don't exist
        """
        if not YAML_AVAILABLE:
            return None

        sections_dir = self.project_root / ".moai" / "config" / "sections"
        if not sections_dir.exists():
            return None

        config = {}

        try:
            # Load user.yaml
            user_file = sections_dir / "user.yaml"
            if user_file.exists():
                with open(user_file, "r", encoding="utf-8") as f:
                    user_data = yaml.safe_load(f) or {}
                    user_config = user_data.get("user", {})
                    if "name" in user_config:
                        config["user_name"] = user_config["name"]

            # Load language.yaml
            language_file = sections_dir / "language.yaml"
            if language_file.exists():
                with open(language_file, "r", encoding="utf-8") as f:
                    lang_data = yaml.safe_load(f) or {}
                    language_config = lang_data.get("language", {})

                    if "conversation_language" in language_config:
                        config["conversation_language"] = language_config["conversation_language"]
                    if "conversation_language_name" in language_config:
                        config["conversation_language_name"] = language_config["conversation_language_name"]
                    if "agent_prompt_language" in language_config:
                        config["agent_prompt_language"] = language_config["agent_prompt_language"]
                    if "git_commit_messages" in language_config:
                        config["git_commit_messages"] = language_config["git_commit_messages"]
                    if "code_comments" in language_config:
                        config["code_comments"] = language_config["code_comments"]
                    if "documentation" in language_config:
                        config["documentation"] = language_config["documentation"]
                    if "error_messages" in language_config:
                        config["error_messages"] = language_config["error_messages"]

            return config if config else None

        except (yaml.YAMLError, IOError, KeyError) as e:
            print(f"Warning: Failed to load section files: {e}")
            return None

    def _load_main_config_file(self) -> Optional[Dict[str, Any]]:
        """
        Load configuration from .moai/config/config.yaml or config.json file.

        Returns:
            Configuration dictionary or None if file doesn't exist
        """
        try:
            if not self.config_file_path.exists():
                return None

            with open(self.config_file_path, "r", encoding="utf-8") as f:
                if self.config_file_path.suffix in [".yaml", ".yml"]:
                    if not YAML_AVAILABLE:
                        print(f"Warning: PyYAML not available, skipping {self.config_file_path}")
                        return None
                    full_config = yaml.safe_load(f) or {}
                else:
                    full_config = json.load(f)

            # Extract only relevant fields
            config = {}

            # User name
            user_config = full_config.get("user", {})
            if "name" in user_config:
                config["user_name"] = user_config["name"]

            # GitHub profile name (stored separately, not as user name fallback)
            github_config = full_config.get("github", {})
            if "profile_name" in github_config:
                config["github_profile_name"] = github_config["profile_name"]

            # Language settings
            language_config = full_config.get("language", {})
            if "conversation_language" in language_config:
                config["conversation_language"] = language_config["conversation_language"]

            if "conversation_language_name" in language_config:
                config["conversation_language_name"] = language_config["conversation_language_name"]

            if "agent_prompt_language" in language_config:
                config["agent_prompt_language"] = language_config["agent_prompt_language"]

            if "git_commit_messages" in language_config:
                config["git_commit_messages"] = language_config["git_commit_messages"]

            if "code_comments" in language_config:
                config["code_comments"] = language_config["code_comments"]

            if "documentation" in language_config:
                config["documentation"] = language_config["documentation"]

            if "error_messages" in language_config:
                config["error_messages"] = language_config["error_messages"]

            return config

        except (
            json.JSONDecodeError,
            yaml.YAMLError if YAML_AVAILABLE else Exception,
            IOError,
            KeyError,
        ) as e:
            # Log error but don't fail - fall back to defaults
            print(f"Warning: Failed to load config file {self.config_file_path}: {e}")
            return None

    def _load_env_config(self) -> Optional[Dict[str, Any]]:
        """
        Load configuration from environment variables.

        Returns:
            Configuration dictionary or None if no environment variables set
        """
        env_config = {}

        # User name
        user_name = os.getenv("MOAI_USER_NAME")
        if user_name:
            env_config["user_name"] = user_name

        # Conversation language
        conv_lang = os.getenv("MOAI_CONVERSATION_LANG")
        if conv_lang:
            env_config["conversation_language"] = conv_lang

        # Agent prompt language
        agent_lang = os.getenv("MOAI_AGENT_PROMPT_LANG")
        if agent_lang:
            env_config["agent_prompt_language"] = agent_lang

        # Language name (if not provided, will be auto-generated)
        lang_name = os.getenv("MOAI_CONVERSATION_LANG_NAME")
        if lang_name:
            env_config["conversation_language_name"] = lang_name

        return env_config if env_config else None

    def _ensure_consistency(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """
        Ensure configuration consistency and auto-generate derived values.

        Args:
            config: Configuration dictionary to normalize

        Returns:
            Normalized configuration dictionary
        """
        # Always auto-generate language name based on conversation language
        conv_lang = config.get("conversation_language", "en")
        config["conversation_language_name"] = self.get_language_name(conv_lang)

        # Ensure agent prompt language defaults to conversation language
        if not config.get("agent_prompt_language"):
            config["agent_prompt_language"] = conv_lang

        # Validate language codes - all languages are now supported dynamically
        # No validation needed - all language codes are accepted

        return config

    def get_language_name(self, language_code: str) -> str:
        """
        Get display name for a language code dynamically.

        Args:
            language_code: Language code (two-letter or three-letter)

        Returns:
            Display name for the language (English name)
        """
        if not language_code:
            return "Unknown"

        # Common language name mappings for well-known codes
        common_names = {
            "ko": "Korean",
            "en": "English",
            "ja": "Japanese",
            "zh": "Chinese",
            "es": "Spanish",
            "fr": "French",
            "de": "German",
            "ru": "Russian",
            "pt": "Portuguese",
            "it": "Italian",
            "ar": "Arabic",
            "hi": "Hindi",
            "th": "Thai",
            "vi": "Vietnamese",
            "id": "Indonesian",
            "ms": "Malay",
            "tr": "Turkish",
            "pl": "Polish",
            "nl": "Dutch",
            "sv": "Swedish",
            "da": "Danish",
            "no": "Norwegian",
            "fi": "Finnish",
            "he": "Hebrew",
            "el": "Greek",
            "hu": "Hungarian",
            "cs": "Czech",
            "sk": "Slovak",
            "ro": "Romanian",
            "bg": "Bulgarian",
            "hr": "Croatian",
            "sr": "Serbian",
            "sl": "Slovenian",
            "et": "Estonian",
            "lv": "Latvian",
            "lt": "Lithuanian",
            "uk": "Ukrainian",
            "be": "Belarusian",
            "ka": "Georgian",
            "am": "Amharic",
            "sw": "Swahili",
            "zu": "Zulu",
            "af": "Afrikaans",
            "is": "Icelandic",
            "mt": "Maltese",
            "cy": "Welsh",
            "ga": "Irish",
            "gd": "Scottish Gaelic",
            "eu": "Basque",
            "ca": "Catalan",
            "gl": "Galician",
            "mk": "Macedonian",
            "sq": "Albanian",
            "bn": "Bengali",
            "ta": "Tamil",
            "te": "Telugu",
            "ml": "Malayalam",
            "kn": "Kannada",
            "gu": "Gujarati",
            "pa": "Punjabi",
            "mr": "Marathi",
            "ne": "Nepali",
            "si": "Sinhala",
            "my": "Myanmar (Burmese)",
            "km": "Khmer",
            "lo": "Lao",
            "mn": "Mongolian",
            "hy": "Armenian",
            "az": "Azerbaijani",
            "kk": "Kazakh",
            "ky": "Kyrgyz",
            "uz": "Uzbek",
            "tg": "Tajik",
            "tk": "Turkmen",
            "ps": "Pashto",
            "sd": "Sindhi",
            "ur": "Urdu",
        }

        # Normalize language code
        normalized_code = self._standardize_language_code(language_code.lower())

        # Return known name or generate from language code
        if normalized_code in common_names:
            return common_names[normalized_code]
        else:
            # Generate English name from language code (capitalize first letter)
            return normalized_code.title() if normalized_code else "Unknown"

    def _standardize_language_code(self, language_code: str) -> str:
        """
        Standardize language code to common format.

        Args:
            language_code: Input language code

        Returns:
            Standardized language code
        """
        if not language_code:
            return ""

        # Handle common variations
        code = language_code.lower().strip()

        # Convert common variants to standard codes
        standardizations = {
            "zh-cn": "zh",
            "zh-tw": "zh",
            "zh-hk": "zh",
            "en-us": "en",
            "en-gb": "en",
            "ko-kr": "ko",
            "ja-jp": "ja",
            "es-es": "es",
            "es-mx": "es",
            "pt-br": "pt",
            "pt-pt": "pt",
            "fr-fr": "fr",
            "de-de": "de",
            "it-it": "it",
            "ru-ru": "ru",
        }

        return standardizations.get(code, code[:2])  # Return first 2 chars if not in standardizations

    def is_korean_language(self, config: Optional[Dict[str, Any]] = None) -> bool:
        """
        Check if the resolved configuration uses Korean language.

        Args:
            config: Optional configuration dictionary. If None, uses resolved config.

        Returns:
            True if Korean language, False otherwise
        """
        if config is None:
            config = self.resolve_config()

        return config.get("conversation_language") == "ko"

    def get_personalized_greeting(self, config: Optional[Dict[str, Any]] = None) -> str:
        """
        Get personalized greeting based on user name and language.

        Args:
            config: Optional configuration dictionary. If None, uses resolved config.

        Returns:
            Personalized greeting string
        """
        if config is None:
            config = self.resolve_config()

        user_name = config.get("user_name", "").strip()
        is_korean = self.is_korean_language(config)

        if user_name:
            if is_korean:
                return f"{user_name}님"
            else:
                return user_name

        return ""

    def export_template_variables(self, config: Optional[Dict[str, Any]] = None) -> Dict[str, str]:
        """
        Export configuration as template variables for substitution.

        Args:
            config: Optional configuration dictionary. If None, uses resolved config.

        Returns:
            Dictionary of template variables
        """
        if config is None:
            config = self.resolve_config()

        return {
            "CONVERSATION_LANGUAGE": config["conversation_language"],
            "CONVERSATION_LANGUAGE_NAME": config["conversation_language_name"],
            "AGENT_PROMPT_LANGUAGE": config["agent_prompt_language"],
            "GIT_COMMIT_MESSAGES_LANGUAGE": config["git_commit_messages"],
            "CODE_COMMENTS_LANGUAGE": config["code_comments"],
            "DOCUMENTATION_LANGUAGE": config["documentation"],
            "ERROR_MESSAGES_LANGUAGE": config["error_messages"],
            "USER_NAME": config["user_name"],
            "PERSONALIZED_GREETING": self.get_personalized_greeting(config),
            "CONFIG_SOURCE": config["config_source"],
        }

    def clear_cache(self):
        """Clear any cached configuration."""
        self._cached_config = None


# Global resolver instance
_resolver_instance: Optional[LanguageConfigResolver] = None


def get_resolver(project_root: Optional[str] = None) -> LanguageConfigResolver:
    """
    Get the global language configuration resolver instance.

    Args:
        project_root: Project root directory

    Returns:
        LanguageConfigResolver instance
    """
    global _resolver_instance

    if _resolver_instance is None or (project_root and project_root != str(_resolver_instance.project_root)):
        _resolver_instance = LanguageConfigResolver(project_root)

    return _resolver_instance


def resolve_language_config(project_root: Optional[str] = None, force_refresh: bool = False) -> Dict[str, Any]:
    """
    Convenience function to resolve language configuration.

    Args:
        project_root: Project root directory
        force_refresh: Force refresh of cached configuration

    Returns:
        Resolved language configuration
    """
    resolver = get_resolver(project_root)
    return resolver.resolve_config(force_refresh)
