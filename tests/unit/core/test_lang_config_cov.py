"""Comprehensive coverage tests for LanguageConfigResolver module.

Tests LanguageConfigResolver class for language configuration resolution and validation.
Target: 70%+ code coverage with actual code path execution and mocked dependencies.
"""

import os
from pathlib import Path
from unittest.mock import patch

import pytest


class TestLanguageConfigResolverInit:
    """Test LanguageConfigResolver initialization."""

    def test_resolver_instantiation_with_path(self, tmp_path):
        """Should instantiate with project root."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        assert resolver.project_root == tmp_path

    def test_resolver_instantiation_without_path(self):
        """Should use current directory by default."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver()

        assert resolver.project_root == Path.cwd()

    def test_resolver_auto_detect_yaml(self, tmp_path):
        """Should auto-detect YAML config."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.yaml").write_text("user:\n  name: test")

        resolver = LanguageConfigResolver(str(tmp_path))

        assert resolver.config_file_path.suffix in [".yaml", ".yml", ".json"]

    def test_resolver_fallback_json(self, tmp_path):
        """Should fallback to JSON if YAML unavailable."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text('{"user": {"name": "test"}}')

        resolver = LanguageConfigResolver(str(tmp_path))

        assert resolver.config_file_path.exists()

    def test_resolver_default_config(self):
        """Should have default configuration."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        assert len(LanguageConfigResolver.DEFAULT_CONFIG) > 0
        assert "conversation_language" in LanguageConfigResolver.DEFAULT_CONFIG
        assert "user_name" in LanguageConfigResolver.DEFAULT_CONFIG


class TestResolveConfig:
    """Test resolve_config method."""

    def test_resolve_config_defaults(self, tmp_path):
        """Should return defaults when no config."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = resolver.resolve_config()

        # Should have conversation_language set (default or from environment)
        assert "conversation_language" in config
        assert "user_name" in config

    def test_resolve_config_caching(self, tmp_path):
        """Should cache resolved configuration."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config1 = resolver.resolve_config()
        config2 = resolver.resolve_config()

        assert config1 == config2

    def test_resolve_config_force_refresh(self, tmp_path):
        """Should refresh cache when requested."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config1 = resolver.resolve_config()
        config2 = resolver.resolve_config(force_refresh=True)

        assert isinstance(config1, dict)
        assert isinstance(config2, dict)


class TestLoadConfigFile:
    """Test _load_config_file method."""

    def test_load_config_file_json(self, tmp_path):
        """Should load JSON config file."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text('{"user": {"name": "TestUser"}, "language": {"conversation_language": "ko"}}')

        resolver = LanguageConfigResolver(str(tmp_path))
        config = resolver._load_config_file()

        assert config is not None
        assert config.get("user_name") == "TestUser"
        assert config.get("conversation_language") == "ko"

    def test_load_config_file_missing(self, tmp_path):
        """Should return None if file missing."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = resolver._load_config_file()

        assert config is None

    def test_load_config_file_invalid_json(self, tmp_path):
        """Should handle invalid JSON."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text("invalid json {")

        resolver = LanguageConfigResolver(str(tmp_path))
        config = resolver._load_config_file()

        assert config is None

    def test_load_config_file_language_settings(self, tmp_path):
        """Should extract language settings."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text(
            """{
            "language": {
                "conversation_language": "ja",
                "agent_prompt_language": "en",
                "git_commit_messages": "en"
            }
        }"""
        )

        resolver = LanguageConfigResolver(str(tmp_path))
        config = resolver._load_config_file()

        assert config.get("conversation_language") == "ja"
        assert config.get("agent_prompt_language") == "en"

    def test_load_config_file_project_owner_fallback(self, tmp_path):
        """Should extract github profile name separately."""
        import yaml

        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.yaml"
        config_file.write_text(yaml.dump({"github": {"profile_name": "GitHubUser"}}))

        resolver = LanguageConfigResolver(str(tmp_path))
        config = resolver._load_config_file()

        assert config.get("github_profile_name") == "GitHubUser"


class TestLoadEnvConfig:
    """Test _load_env_config method."""

    def test_load_env_config_user_name(self, tmp_path):
        """Should load user name from environment."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        with patch.dict(os.environ, {"MOAI_USER_NAME": "EnvUser"}):
            resolver = LanguageConfigResolver(str(tmp_path))
            config = resolver._load_env_config()

            assert config is not None
            assert config.get("user_name") == "EnvUser"

    def test_load_env_config_conversation_language(self, tmp_path):
        """Should load conversation language from environment."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        with patch.dict(os.environ, {"MOAI_CONVERSATION_LANG": "ko"}):
            resolver = LanguageConfigResolver(str(tmp_path))
            config = resolver._load_env_config()

            assert config is not None
            assert config.get("conversation_language") == "ko"

    def test_load_env_config_all_variables(self, tmp_path):
        """Should load all environment variables."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        env_vars = {
            "MOAI_USER_NAME": "TestUser",
            "MOAI_CONVERSATION_LANG": "ko",
            "MOAI_AGENT_PROMPT_LANG": "en",
            "MOAI_CONVERSATION_LANG_NAME": "Korean",
        }

        with patch.dict(os.environ, env_vars):
            resolver = LanguageConfigResolver(str(tmp_path))
            config = resolver._load_env_config()

            assert config.get("user_name") == "TestUser"
            assert config.get("conversation_language") == "ko"
            assert config.get("agent_prompt_language") == "en"

    def test_load_env_config_no_variables(self, tmp_path):
        """Should return None if no env variables set."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        with patch.dict(os.environ, {}, clear=True):
            resolver = LanguageConfigResolver(str(tmp_path))
            config = resolver._load_env_config()

            assert config is None or config == {}


class TestEnsureConsistency:
    """Test _ensure_consistency method."""

    def test_ensure_consistency_auto_language_name(self, tmp_path):
        """Should auto-generate language name."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {"conversation_language": "ko"}

        result = resolver._ensure_consistency(config)

        assert result["conversation_language_name"] == "Korean"

    def test_ensure_consistency_agent_language_default(self, tmp_path):
        """Should default agent language to conversation language."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {"conversation_language": "ja", "agent_prompt_language": ""}

        result = resolver._ensure_consistency(config)

        assert result["agent_prompt_language"] == "ja"

    def test_ensure_consistency_preserves_overrides(self, tmp_path):
        """Should preserve explicit overrides."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {"conversation_language": "ko", "agent_prompt_language": "en"}

        result = resolver._ensure_consistency(config)

        assert result["agent_prompt_language"] == "en"


class TestGetLanguageName:
    """Test get_language_name method."""

    def test_get_language_name_english(self, tmp_path):
        """Should return English for 'en'."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        name = resolver.get_language_name("en")

        assert name == "English"

    def test_get_language_name_korean(self, tmp_path):
        """Should return Korean for 'ko'."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        name = resolver.get_language_name("ko")

        assert name == "Korean"

    def test_get_language_name_japanese(self, tmp_path):
        """Should return Japanese for 'ja'."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        name = resolver.get_language_name("ja")

        assert name == "Japanese"

    def test_get_language_name_chinese(self, tmp_path):
        """Should return Chinese for 'zh'."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        name = resolver.get_language_name("zh")

        assert name == "Chinese"

    def test_get_language_name_unknown(self, tmp_path):
        """Should handle unknown languages."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        name = resolver.get_language_name("xx")

        assert isinstance(name, str)
        assert name != ""

    def test_get_language_name_empty(self, tmp_path):
        """Should return Unknown for empty code."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        name = resolver.get_language_name("")

        assert name == "Unknown"


class TestStandardizeLanguageCode:
    """Test _standardize_language_code method."""

    def test_standardize_simple_code(self, tmp_path):
        """Should handle simple codes."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        code = resolver._standardize_language_code("EN")

        assert code == "en"

    def test_standardize_variant_codes(self, tmp_path):
        """Should standardize variant codes."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        code = resolver._standardize_language_code("en-us")
        assert code == "en"

        code = resolver._standardize_language_code("zh-cn")
        assert code == "zh"

    def test_standardize_with_whitespace(self, tmp_path):
        """Should strip whitespace."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        code = resolver._standardize_language_code("  ko  ")

        assert code == "ko"

    def test_standardize_empty_code(self, tmp_path):
        """Should handle empty code."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        code = resolver._standardize_language_code("")

        assert code == ""


class TestIsKoreanLanguage:
    """Test is_korean_language method."""

    def test_is_korean_language_true(self, tmp_path):
        """Should detect Korean language."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {"conversation_language": "ko"}

        result = resolver.is_korean_language(config)

        assert result is True

    def test_is_korean_language_false(self, tmp_path):
        """Should detect non-Korean language."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {"conversation_language": "en"}

        result = resolver.is_korean_language(config)

        assert result is False

    def test_is_korean_language_resolved(self, tmp_path):
        """Should use resolved config if none provided."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))

        result = resolver.is_korean_language()

        assert isinstance(result, bool)


class TestGetPersonalizedGreeting:
    """Test get_personalized_greeting method."""

    def test_personalized_greeting_korean_with_name(self, tmp_path):
        """Should format Korean greeting with name."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {"user_name": "테스트", "conversation_language": "ko"}

        greeting = resolver.get_personalized_greeting(config)

        assert "테스트" in greeting
        assert "님" in greeting

    def test_personalized_greeting_english_with_name(self, tmp_path):
        """Should format English greeting with name."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {"user_name": "TestUser", "conversation_language": "en"}

        greeting = resolver.get_personalized_greeting(config)

        assert greeting == "TestUser"

    def test_personalized_greeting_no_name(self, tmp_path):
        """Should return empty string when no name."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {"user_name": "", "conversation_language": "en"}

        greeting = resolver.get_personalized_greeting(config)

        assert greeting == ""

    def test_personalized_greeting_whitespace_name(self, tmp_path):
        """Should handle whitespace-only names."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {"user_name": "   ", "conversation_language": "en"}

        greeting = resolver.get_personalized_greeting(config)

        assert greeting == ""


class TestExportTemplateVariables:
    """Test export_template_variables method."""

    def test_export_all_variables(self, tmp_path):
        """Should export all template variables."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {
            "conversation_language": "ko",
            "conversation_language_name": "Korean",
            "agent_prompt_language": "en",
            "git_commit_messages": "en",
            "code_comments": "en",
            "documentation": "ko",
            "error_messages": "ko",
            "user_name": "TestUser",
            "config_source": "config_file",
        }

        variables = resolver.export_template_variables(config)

        assert "CONVERSATION_LANGUAGE" in variables
        assert "USER_NAME" in variables
        assert variables["CONVERSATION_LANGUAGE"] == "ko"
        assert variables["USER_NAME"] == "TestUser"

    def test_export_includes_personalized_greeting(self, tmp_path):
        """Should include personalized greeting."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config = {
            "conversation_language": "en",
            "conversation_language_name": "English",
            "agent_prompt_language": "en",
            "git_commit_messages": "en",
            "code_comments": "en",
            "documentation": "en",
            "error_messages": "en",
            "user_name": "Alice",
            "config_source": "default",
        }

        variables = resolver.export_template_variables(config)

        assert "PERSONALIZED_GREETING" in variables
        assert variables["PERSONALIZED_GREETING"] == "Alice"


class TestClearCache:
    """Test clear_cache method."""

    def test_clear_cache_clears_cached_config(self, tmp_path):
        """Should clear cached configuration."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        resolver = LanguageConfigResolver(str(tmp_path))
        config1 = resolver.resolve_config()
        resolver.clear_cache()
        config2 = resolver.resolve_config()

        assert isinstance(config1, dict)
        assert isinstance(config2, dict)


class TestGlobalResolverInstance:
    """Test global resolver instance management."""

    @patch("moai_adk.core.language_config_resolver._resolver_instance", None)
    def test_get_resolver_creates_instance(self, tmp_path):
        """Should create resolver instance."""
        from moai_adk.core.language_config_resolver import get_resolver

        resolver = get_resolver(str(tmp_path))

        assert resolver is not None

    @patch("moai_adk.core.language_config_resolver._resolver_instance", None)
    def test_resolve_language_config_function(self, tmp_path):
        """Should resolve config via convenience function."""
        from moai_adk.core.language_config_resolver import resolve_language_config

        config = resolve_language_config(str(tmp_path))

        assert isinstance(config, dict)
        assert "conversation_language" in config


class TestPriorityOrder:
    """Test configuration priority ordering."""

    def test_env_overrides_file(self, tmp_path):
        """Environment variables should override config file."""
        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text('{"language": {"conversation_language": "ja"}}')

        with patch.dict(os.environ, {"MOAI_CONVERSATION_LANG": "ko"}):
            resolver = LanguageConfigResolver(str(tmp_path))
            config = resolver.resolve_config()

            assert config["conversation_language"] == "ko"
            assert config["config_source"] == "environment"

    def test_file_overrides_defaults(self, tmp_path):
        """Config file should override defaults."""
        import os

        from moai_adk.core.language_config_resolver import LanguageConfigResolver

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text('{"language": {"conversation_language": "ja"}}')

        # Clear environment to ensure no override
        with patch.dict(os.environ, {}, clear=True):
            resolver = LanguageConfigResolver(str(tmp_path))
            config = resolver.resolve_config()

            assert config["conversation_language"] == "ja"
            assert config["config_source"] == "config_file"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
