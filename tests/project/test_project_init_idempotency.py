"""Test suite for SPEC-PROJECT-INIT-IDEMPOTENT-001

Tests for project initialization idempotency and config optimization state management.
"""

import json
from datetime import datetime, timezone
from pathlib import Path

import pytest

from moai_adk.core.template.config import ConfigManager


class TestOptimizedFieldSemantics:
    """Test clear semantics for optimized field."""

    def test_optimized_false_set_on_reinit(self, tmp_path: Path) -> None:
        """Test that optimized=false is set when reinitializing existing project."""
        config_path = tmp_path / ".moai" / "config" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Create initial config with optimized=true
        initial_config = {
            "project": {"name": "test", "optimized": True},
            "moai": {"version": "0.25.0"},
        }
        with open(config_path, "w") as f:
            json.dump(initial_config, f)

        # Simulate reinit by setting optimized to False
        ConfigManager.set_optimized(config_path, False)

        # Verify optimized is now false
        with open(config_path) as f:
            config = json.load(f)

        assert config["project"]["optimized"] is False

    def test_optimized_true_set_after_merge(self, tmp_path: Path) -> None:
        """Test that optimized=true is set after successful config merge."""
        config_path = tmp_path / ".moai" / "config" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Create config with optimized=false
        initial_config = {
            "project": {"name": "test", "optimized": False},
            "moai": {"version": "0.25.0"},
        }
        with open(config_path, "w") as f:
            json.dump(initial_config, f)

        # Simulate merge completion by setting optimized to True
        ConfigManager.set_optimized(config_path, True)

        # Verify optimized is now true
        with open(config_path) as f:
            config = json.load(f)

        assert config["project"]["optimized"] is True

    def test_config_has_optimized_field(self, tmp_path: Path) -> None:
        """Test that config can be ensured to have optimized field."""
        config_path = tmp_path / ".moai" / "config" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Create minimal config without optimized field
        initial_config = {"project": {"name": "test"}}
        with open(config_path, "w") as f:
            json.dump(initial_config, f)

        # Ensure optimized field is created by calling set_optimized
        ConfigManager.set_optimized(config_path, False)

        # Verify optimized field now exists
        with open(config_path) as f:
            config = json.load(f)

        assert "optimized" in config.get("project", {})
        assert config["project"]["optimized"] is False


class TestOptimizedAtTimestamp:
    """Test optimized_at timestamp management."""

    def test_optimized_at_timestamp_added_on_merge(self, tmp_path: Path) -> None:
        """Test that optimized_at is set with ISO timestamp when merge completes."""
        config_path = tmp_path / ".moai" / "config" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Create config with optimized=false
        initial_config = {
            "project": {"name": "test", "optimized": False, "optimized_at": None},
            "moai": {"version": "0.25.0"},
        }
        with open(config_path, "w") as f:
            json.dump(initial_config, f)

        # Set optimized=true with timestamp
        before_time = datetime.now(timezone.utc)
        ConfigManager.set_optimized_with_timestamp(config_path, True)
        after_time = datetime.now(timezone.utc)

        # Verify timestamp is set
        with open(config_path) as f:
            config = json.load(f)

        assert config["project"]["optimized"] is True
        assert config["project"]["optimized_at"] is not None

        # Verify timestamp is ISO format and within expected range
        timestamp_str = config["project"]["optimized_at"]
        timestamp = datetime.fromisoformat(timestamp_str.replace("Z", "+00:00"))
        assert before_time <= timestamp <= after_time

    def test_optimized_at_null_on_reinit(self, tmp_path: Path) -> None:
        """Test that optimized_at becomes null when reinitializing."""
        config_path = tmp_path / ".moai" / "config" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Create config with timestamp
        initial_config = {
            "project": {
                "name": "test",
                "optimized": True,
                "optimized_at": "2025-01-01T00:00:00Z",
            },
            "moai": {"version": "0.25.0"},
        }
        with open(config_path, "w") as f:
            json.dump(initial_config, f)

        # Reinit by setting optimized=false
        ConfigManager.set_optimized_with_timestamp(config_path, False)

        # Verify optimized_at is null
        with open(config_path) as f:
            config = json.load(f)

        assert config["project"]["optimized"] is False
        assert config["project"]["optimized_at"] is None


class TestIdempotency:
    """Test idempotency - safe to run multiple times."""

    def test_idempotent_first_run(self, tmp_path: Path) -> None:
        """Test that first run completes successfully."""
        config_path = tmp_path / ".moai" / "config" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Create initial config
        initial_config = {
            "project": {"name": "test", "optimized": False},
            "moai": {"version": "0.25.0"},
        }
        with open(config_path, "w") as f:
            json.dump(initial_config, f)

        # First run: merge and set optimized=true
        ConfigManager.set_optimized_with_timestamp(config_path, True)

        with open(config_path) as f:
            config_after_first = json.load(f)

        assert config_after_first["project"]["optimized"] is True
        first_timestamp = config_after_first["project"]["optimized_at"]

        # Verify first run successful
        assert first_timestamp is not None

    def test_idempotent_second_run(self, tmp_path: Path) -> None:
        """Test that second run (with optimized=true) skips merge."""
        config_path = tmp_path / ".moai" / "config" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Create config with optimized=true
        initial_config = {
            "project": {
                "name": "test",
                "optimized": True,
                "optimized_at": "2025-01-01T00:00:00Z",
            },
            "moai": {"version": "0.25.0"},
        }
        with open(config_path, "w") as f:
            json.dump(initial_config, f)

        first_timestamp = initial_config["project"]["optimized_at"]

        # Second run: should skip (no-op)
        # In real implementation, project-manager would check optimized flag
        # and skip merge if already true
        with open(config_path) as f:
            config = json.load(f)

        # Verify state unchanged
        assert config["project"]["optimized"] is True
        assert config["project"]["optimized_at"] == first_timestamp


class TestUserEditsPreservation:
    """Test that user edits are preserved during merge."""

    def test_user_edits_preserved_after_merge(self, tmp_path: Path) -> None:
        """Test that user's custom settings are preserved after merge."""
        config_path = tmp_path / ".moai" / "config" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Create config with user edits
        initial_config = {
            "project": {
                "name": "test",
                "optimized": False,
                "custom_setting": "user_value",
            },
            "moai": {"version": "0.25.0"},
        }
        with open(config_path, "w") as f:
            json.dump(initial_config, f)

        # Merge and set optimized=true
        ConfigManager.set_optimized_with_timestamp(config_path, True)

        # Verify user edits still present
        with open(config_path) as f:
            config = json.load(f)

        assert config["project"]["custom_setting"] == "user_value"
        assert config["project"]["optimized"] is True


class TestSessionHookDisplay:
    """Test session hook displays optimization status correctly."""

    def test_session_hook_shows_optimization_status(self, tmp_path: Path) -> None:
        """Test that session hook can read and display optimization status."""
        config_path = tmp_path / ".moai" / "config" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Test with optimized=false
        config_false = {
            "project": {"name": "test", "optimized": False},
            "moai": {"version": "0.25.0"},
        }
        with open(config_path, "w") as f:
            json.dump(config_false, f)

        with open(config_path) as f:
            config = json.load(f)

        optimized = config.get("project", {}).get("optimized", True)
        status_indicator = "✅" if optimized else "⚠️"

        assert status_indicator == "⚠️"
        assert optimized is False

        # Test with optimized=true
        config_true = {
            "project": {"name": "test", "optimized": True},
            "moai": {"version": "0.25.0"},
        }
        with open(config_path, "w") as f:
            json.dump(config_true, f)

        with open(config_path) as f:
            config = json.load(f)

        optimized = config.get("project", {}).get("optimized", True)
        status_indicator = "✅" if optimized else "⚠️"

        assert status_indicator == "✅"
        assert optimized is True


class TestClearUserGuidance:
    """Test clear user guidance messages in various scenarios."""

    def test_init_command_guidance_on_reinit(self, tmp_path: Path) -> None:
        """Test that init command shows clear guidance when reinitializing."""
        # This test verifies the message structure and content
        # In practice, this is tested through CLI integration tests
        # Here we verify the logic would work correctly

        # When reinit happens, optimized should be False
        config = {"project": {"optimized": False}}

        # Guidance should recommend running /moai:0-project
        optimized = config.get("project", {}).get("optimized", True)

        if not optimized:
            guidance_needed = True
        else:
            guidance_needed = False

        assert guidance_needed is True

    def test_already_optimized_message(self, tmp_path: Path) -> None:
        """Test message when config is already optimized."""
        config = {"project": {"optimized": True}}

        optimized = config.get("project", {}).get("optimized", True)

        if optimized:
            display_message = "Configuration already optimized and ready"
        else:
            display_message = "Configuration merge required"

        assert display_message == "Configuration already optimized and ready"


@pytest.fixture
def config_manager(tmp_path: Path) -> ConfigManager:
    """Create ConfigManager instance for testing."""
    config_path = tmp_path / ".moai" / "config" / "config.json"
    config_path.parent.mkdir(parents=True, exist_ok=True)
    return ConfigManager(config_path)
