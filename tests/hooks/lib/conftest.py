"""Shared fixtures for hooks/moai/lib tests"""

import json

import pytest


@pytest.fixture
def temp_config_dir(tmp_path):
    """Create a temporary config directory"""
    config_dir = tmp_path / ".moai"
    config_dir.mkdir()
    return config_dir


@pytest.fixture
def valid_config_file(temp_config_dir):
    """Create a valid config file"""
    config = {
        "hooks": {
            "timeout_seconds": 5,
            "timeout_ms": 5000,
            "minimum_timeout_seconds": 1,
            "graceful_degradation": True,
            "exit_codes": {"success": 0, "error": 1, "critical_error": 2, "config_error": 3},
            "messages": {
                "timeout": {
                    "post_tool_use": "⚠️ PostToolUse timeout",
                    "session_end": "⚠️ SessionEnd cleanup timeout",
                    "session_start": "⚠️ Session start timeout",
                }
            },
            "cache": {"directory": ".moai/cache", "version_ttl_seconds": 1800, "git_ttl_seconds": 10},
            "git": {"timeout_seconds": 2},
        },
        "language": {"conversation_language": "en"},
    }
    config_file = temp_config_dir / "config.json"
    config_file.write_text(json.dumps(config, indent=2))
    return config_file


@pytest.fixture
def invalid_config_file(temp_config_dir):
    """Create an invalid (malformed) config file"""
    config_file = temp_config_dir / "config.json"
    config_file.write_text("{ invalid json content ]")
    return config_file


@pytest.fixture
def empty_config_file(temp_config_dir):
    """Create an empty config file"""
    config_file = temp_config_dir / "config.json"
    config_file.write_text("")
    return config_file


@pytest.fixture
def sample_json_data():
    """Sample JSON data for testing"""
    return {
        "name": "test",
        "version": "1.0.0",
        "nested": {"key": "value", "deep": {"data": "test"}},
        "list": [1, 2, 3, 4, 5],
    }


@pytest.fixture
def sample_json_file(tmp_path, sample_json_data):
    """Create a sample JSON file"""
    file_path = tmp_path / "sample.json"
    file_path.write_text(json.dumps(sample_json_data, indent=2))
    return file_path


@pytest.fixture
def invalid_json_file(tmp_path):
    """Create an invalid JSON file"""
    file_path = tmp_path / "invalid.json"
    file_path.write_text("{ bad json: }")
    return file_path


@pytest.fixture
def nonexistent_file(tmp_path):
    """Create a path to a nonexistent file"""
    return tmp_path / "nonexistent.json"
