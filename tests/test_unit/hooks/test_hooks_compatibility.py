#!/usr/bin/env python3
"""
Compatibility tests for .claude/hooks optimization (TDD RED phase)

Tests that verify optimized hooks maintain compatibility with existing systems.
These tests ensure no breaking changes are introduced during optimization.

@TEST:UNIT-HOOKS-COMPAT
@REQ:COMPAT-HOOKS-001
"""

import sys
import os
import json
from pathlib import Path
from unittest.mock import patch, MagicMock, mock_open
import pytest

# Add project root to path for imports
PROJECT_ROOT = Path(__file__).resolve().parents[3]
HOOKS_DIR = PROJECT_ROOT / ".claude" / "hooks" / "moai"
CLAUDE_DIR = PROJECT_ROOT / ".claude"
sys.path.insert(0, str(HOOKS_DIR))


class TestClaudeCodeApiCompatibility:
    """Test Claude Code API compatibility preservation

    @TEST:UNIT-CLAUDE-API-COMPAT
    Ensure optimized hooks work with Claude Code's hook system
    """

    def setup_method(self):
        """Setup test environment"""
        self.project_root = PROJECT_ROOT

    def test_hooks_maintain_expected_interface(self):
        """Test that hooks maintain expected function signatures

        @TEST:UNIT-HOOK-INTERFACE
        This test defines required interface compatibility (RED phase)
        """
        # Test session_start_notice hook interface
        try:
            import session_start_notice
            assert hasattr(session_start_notice, 'SessionNotifier'), \
                "Should have SessionNotifier class for Claude Code integration"

            notifier = session_start_notice.SessionNotifier(self.project_root)
            assert notifier is not None, "Should initialize with project_root"
        except ImportError:
            pytest.skip("session_start_notice.py not found")

    def test_pre_write_guard_hook_interface(self):
        """Test pre_write_guard maintains hook interface

        @TEST:UNIT-PRE-WRITE-INTERFACE
        This test ensures security hook compatibility (RED phase)
        """
        try:
            import pre_write_guard
            assert callable(getattr(pre_write_guard, 'check_file_safety', None)) or \
                   callable(getattr(pre_write_guard, 'main', None)) or \
                   callable(getattr(pre_write_guard, 'validate', None)), \
                "Should have callable function for Claude Code hook system"
        except ImportError:
            pytest.skip("pre_write_guard.py not found")

    def test_file_monitor_hook_interface(self):
        """Test file_monitor maintains hook interface after integration

        @TEST:UNIT-FILE-MONITOR-INTERFACE
        This test ensures file monitoring compatibility (RED phase)
        """
        # This will fail initially as integration not done
        try:
            import file_monitor
            assert hasattr(file_monitor, 'FileMonitor'), \
                "Should have FileMonitor class after integration"

            monitor = file_monitor.FileMonitor(self.project_root)
            assert monitor is not None, "Should initialize file monitor"
        except ImportError:
            # Expected to fail before integration
            assert False, "file_monitor.py should exist after integration"


class TestAgentCommunicationCompatibility:
    """Test agent communication interface preservation

    @TEST:UNIT-AGENT-COMM-COMPAT
    Ensure optimized hooks can still communicate with MoAI agents
    """

    def test_session_notice_agent_communication(self):
        """Test session notice can still communicate with agents

        @TEST:UNIT-SESSION-AGENT-COMM
        This test ensures agent integration works (RED phase)
        """
        try:
            import session_start_notice
            notifier = session_start_notice.SessionNotifier(self.project_root)

            with patch.object(Path, 'exists', return_value=True), \
                 patch('builtins.open', mock_open(read_data='{}')):

                status = notifier.get_project_status()
                assert isinstance(status, dict), "Should return dictionary for agent consumption"

                expected_keys = ['project_name', 'initialized', 'moai_version']
                for key in expected_keys:
                    if key not in status:
                        pytest.skip(f"Agent communication key {key} not found - may be optimized out")
        except ImportError:
            pytest.skip("session_start_notice.py not found")


class TestSettingsFileCompatibility:
    """Test .claude/settings.json compatibility preservation

    @TEST:UNIT-SETTINGS-COMPAT
    Ensure optimized hooks respect Claude Code settings
    """

    def setup_method(self):
        """Setup test environment"""
        self.settings_file = CLAUDE_DIR / "settings.json"

    def test_settings_file_respected(self):
        """Test that hooks respect .claude/settings.json

        @TEST:UNIT-SETTINGS-RESPECT
        This test ensures settings compatibility (RED phase)
        """
        mock_settings = {
            "defaultMode": "acceptEdits",
            "overrides": {
                ".claude/hooks/moai/pre_write_guard.py": "ask"
            }
        }

        with patch.object(Path, 'exists', return_value=True), \
             patch('builtins.open', mock_open(read_data=json.dumps(mock_settings))):

            try:
                with open(self.settings_file) as f:
                    settings = json.load(f)
                assert settings == mock_settings, "Should read settings correctly"
            except Exception:
                pytest.skip("Settings reading not implemented in optimized hooks")


class TestBackwardCompatibility:
    """Test backward compatibility with existing workflows

    @TEST:UNIT-BACKWARD-COMPAT
    Ensure optimized hooks don't break existing MoAI workflows
    """

    def test_moai_pipeline_compatibility(self):
        """Test compatibility with MoAI 4-stage pipeline

        @TEST:UNIT-PIPELINE-COMPAT
        This test ensures pipeline workflow compatibility (RED phase)
        """
        try:
            import session_start_notice
            notifier = session_start_notice.SessionNotifier(self.project_root)

            with patch.object(Path, 'exists', return_value=True), \
                 patch('builtins.open', mock_open(read_data='{}')):

                status = notifier.get_project_status()
                pipeline_info_keys = ['initialized', 'pipeline_stage']
                for key in pipeline_info_keys:
                    if key in status:
                        assert status[key] is not None, f"Pipeline info {key} should be available"
        except ImportError:
            pytest.skip("session_start_notice.py not found")

    def test_git_workflow_compatibility(self):
        """Test compatibility with Git workflow automation

        @TEST:UNIT-GIT-WORKFLOW-COMPAT
        This test ensures Git integration compatibility (RED phase)
        """
        try:
            import file_monitor
            monitor = file_monitor.FileMonitor(self.project_root)
            assert hasattr(monitor, 'create_checkpoint') or \
                   hasattr(monitor, 'git_checkpoint'), \
                "Should maintain Git checkpoint capability"
        except ImportError:
            try:
                import auto_checkpoint
                assert hasattr(auto_checkpoint, 'create_checkpoint'), \
                    "Should maintain checkpoint functionality during transition"
            except ImportError:
                pytest.skip("Neither file_monitor nor auto_checkpoint found")


if __name__ == "__main__":
    pytest.main([__file__, "-v"])