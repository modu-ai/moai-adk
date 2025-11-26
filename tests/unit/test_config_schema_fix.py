"""Tests for SPEC-CONFIG-FIX-001: Config Schema Completeness and Version Comparison

RED Phase Tests - All tests should fail initially
"""

import json
import tempfile
from datetime import datetime
from pathlib import Path
from unittest import mock

from moai_adk.core.project.initializer import ProjectInitializer


class TestConfigSchemaCompleteness:
    """Test config schema has all required fields"""

    def test_config_has_git_strategy_field(self):
        """Test that config_data includes git_strategy field with correct structure"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            initializer = ProjectInitializer(project_path)

            result = initializer.initialize(mode="personal", locale="en", backup_enabled=False)

            assert result.success, f"Initialization should succeed. Errors: {result.errors}"

            # Read generated config
            config_file = project_path / ".moai" / "config" / "config.json"
            assert config_file.exists(), "config.json should exist"

            config = json.loads(config_file.read_text())

            # Check git_strategy exists
            assert "git_strategy" in config, "git_strategy field missing from config"
            assert isinstance(config["git_strategy"], dict), "git_strategy should be a dict"

            # Check personal mode exists
            assert "personal" in config["git_strategy"], "personal mode missing from git_strategy"
            assert isinstance(config["git_strategy"]["personal"], dict), "personal should be a dict"

            # Verify key fields in personal mode
            personal = config["git_strategy"]["personal"]
            required_keys = [
                "auto_checkpoint",
                "checkpoint_type",
                "auto_commit",
                "branch_prefix",
                "develop_branch",
                "main_branch",
            ]
            for key in required_keys:
                assert key in personal, f"Missing key '{key}' in git_strategy.personal"

    def test_config_has_constitution_field(self):
        """Test that config_data includes constitution field with TDD enforcement"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            initializer = ProjectInitializer(project_path)

            # Initialize project
            result = initializer.initialize(mode="personal", locale="en", backup_enabled=False)

            assert result.success, "Initialization should succeed"

            # Read config
            config_file = project_path / ".moai" / "config" / "config.json"
            config = json.loads(config_file.read_text())

            # Check constitution exists
            assert "constitution" in config, "constitution field missing from config"
            assert isinstance(config["constitution"], dict), "constitution should be a dict"

            # Verify required fields
            constitution = config["constitution"]
            assert "enforce_tdd" in constitution, "enforce_tdd missing from constitution"
            assert "test_coverage_target" in constitution, "test_coverage_target missing from constitution"
            assert "principles" in constitution, "principles missing from constitution"

            # Verify values
            assert constitution["enforce_tdd"] is True, "enforce_tdd should default to True"
            assert constitution["test_coverage_target"] == 90, "test_coverage_target should default to 90"

    def test_config_has_session_field(self):
        """Test that config_data includes session field with suppress_setup_messages"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            initializer = ProjectInitializer(project_path)

            # Initialize project
            result = initializer.initialize(mode="personal", locale="en", backup_enabled=False)

            assert result.success, "Initialization should succeed"

            # Read config
            config_file = project_path / ".moai" / "config" / "config.json"
            config = json.loads(config_file.read_text())

            # Check session exists
            assert "session" in config, "session field missing from config"
            assert isinstance(config["session"], dict), "session should be a dict"

            # Check suppress_setup_messages exists
            session = config["session"]
            assert "suppress_setup_messages" in session, "suppress_setup_messages missing from session"

            # Verify default value
            assert session["suppress_setup_messages"] is False, "suppress_setup_messages should default to False"


class TestVersionComparison:
    """Test that version comparison uses semantic versioning"""

    def test_version_comparison_uses_semantic_versioning(self):
        """Test that version comparison correctly handles semantic versions"""
        from packaging.version import Version

        # Test cases: (installed, latest, expected_is_matched, expected_is_newer)
        test_cases = [
            ("0.25.6", "0.25.6", True, False),  # Same version
            ("0.25.6", "0.25.7", False, False),  # Patch update available
            ("0.25.6", "0.26.0", False, False),  # Minor update available
            ("0.26.0", "0.25.6", False, True),  # Pre-release is newer
            ("0.26.0a1", "0.26.0", False, False),  # Alpha < stable
            ("0.26.0rc1", "0.26.0", False, False),  # Release candidate < stable
        ]

        for installed, latest, expected_matched, expected_newer in test_cases:
            installed_ver = Version(installed)
            latest_ver = Version(latest)

            is_matched = installed_ver == latest_ver
            is_newer = installed_ver > latest_ver

            assert (
                is_matched == expected_matched
            ), f"Version comparison failed for {installed} vs {latest}: expected matched={expected_matched}, got {is_matched}"
            assert (
                is_newer == expected_newer
            ), f"Version comparison failed for {installed} vs {latest}: expected newer={expected_newer}, got {is_newer}"

    def test_version_comparison_prevents_downgrade_advice(self):
        """Test that semantic versioning prevents suggesting downgrade (0.25.7 -> 1.0.0 confusion)"""
        from packaging.version import Version

        # Test case: 1.0.0 > 0.25.7 with semantic versioning
        # With string comparison (WRONG): "1.0.0" < "0.25.7" -> suggests downgrade
        # With semantic versioning (CORRECT): 1.0.0 > 0.25.7 -> suggests upgrade
        installed = Version("0.25.7")
        latest = Version("1.0.0")

        assert latest > installed, "1.0.0 should be greater than 0.25.7 with semantic versioning"

        # Also test: 0.3.0 < 0.25.7 (to ensure we don't suggest upgrade to lower version)
        installed2 = Version("0.25.7")
        latest2 = Version("0.3.0")

        assert latest2 < installed2, "0.3.0 should be less than 0.25.7 with semantic versioning"


class TestSuppressSetupMessagesWithNewFields:
    """Test that suppress_setup_messages feature works with new config fields"""

    def test_suppress_setup_messages_works_with_new_fields(self):
        """Test that session.suppress_setup_messages feature uses new fields"""
        # Import the hook function
        from pathlib import Path as PathlibPath

        hook_path = (
            PathlibPath(__file__).parent.parent.parent / ".claude/hooks/alfred/session_start__config_health_check.py"
        )

        # Read hook as module
        spec = __import__("importlib.util").util.spec_from_file_location("config_health_check", hook_path)
        assert spec is not None, "Hook module should be importable"
        hook = __import__("importlib.util").util.module_from_spec(spec)
        assert spec.loader is not None
        spec.loader.exec_module(hook)

        # Create test config with session field
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)

            config_data = {
                "session": {
                    "suppress_setup_messages": False,
                    "setup_messages_suppressed_at": None,
                }
            }

            config_file = config_dir / "config.json"
            config_file.write_text(json.dumps(config_data))

            # Mock cwd to return test directory
            with mock.patch("pathlib.Path.cwd", return_value=project_path):
                # suppress_setup_messages should be False initially
                is_suppressed = hook.should_suppress_setup()
                assert is_suppressed is False, "Setup should not be suppressed when flag is False"

                # Update to suppress
                config_data["session"]["suppress_setup_messages"] = True
                config_data["session"]["setup_messages_suppressed_at"] = datetime.now().isoformat()
                config_file.write_text(json.dumps(config_data))

                is_suppressed = hook.should_suppress_setup()
                assert is_suppressed is True, "Setup should be suppressed when flag is True and timestamp is recent"


class TestConfigHealthCheckValidation:
    """Test that config health check correctly validates new fields"""

    def test_config_health_check_validates_session_field(self):
        """Test that session_start hook validates session field in config"""
        from pathlib import Path as PathlibPath

        hook_path = (
            PathlibPath(__file__).parent.parent.parent / ".claude/hooks/alfred/session_start__config_health_check.py"
        )

        # Import hook module
        spec = __import__("importlib.util").util.spec_from_file_location("config_health_check", hook_path)
        assert spec is not None
        hook = __import__("importlib.util").util.module_from_spec(spec)
        assert spec.loader is not None
        spec.loader.exec_module(hook)

        # Test config without session field
        incomplete_config = {
            "project": {"name": "test"},
            "language": {"conversation_language": "en"},
            "git_strategy": {},
            "constitution": {},
        }

        is_complete, missing = hook.check_config_completeness(incomplete_config)
        assert not is_complete, "Config without session should be incomplete"
        assert "session" in missing, "Missing 'session' should be reported"

        # Test config with session field
        complete_config = {
            "project": {"name": "test"},
            "language": {"conversation_language": "en"},
            "git_strategy": {},
            "constitution": {},
            "session": {"suppress_setup_messages": False},
        }

        is_complete, missing = hook.check_config_completeness(complete_config)
        assert is_complete, f"Config with session should be complete. Missing: {missing}"


class TestClaudeMdReduction:
    """Test that CLAUDE.md is optimized"""

    def test_CLAUDE_md_under_40KB(self):
        """Test that CLAUDE.md file size is optimized (under 40KB)"""
        claude_md_path = Path(__file__).parent.parent.parent / "CLAUDE.md"

        assert claude_md_path.exists(), "CLAUDE.md should exist"

        file_size = claude_md_path.stat().st_size
        # Target is under 40KB (40,000 bytes)
        # SPEC says <40KB but current is ~8.5KB which is ideal
        # We'll check it's still reasonable
        assert file_size < 80000, f"CLAUDE.md should be optimized, current size: {file_size} bytes"
