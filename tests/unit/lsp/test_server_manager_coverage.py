"""Additional coverage tests for LSP server manager.

Tests for lines not covered by existing tests.
"""

import asyncio
from unittest.mock import patch

import pytest

from moai_adk.lsp.server_manager import (
    LSPServerConfig,
    LSPServerManager,
    ServerNotFoundError,
    ServerStartError,
)


class TestLSPServerManagerParseConfig:
    """Test config parsing edge cases."""

    def test_parse_config_skips_non_dict_entries(self):
        """Should skip entries that are not dictionaries."""
        manager = LSPServerManager()

        data = {
            "python": {"command": "pylsp", "args": []},
            "typescript": "not_a_dict",  # This should be skipped
            "javascript": {"command": "ts-language-server", "args": []},
        }

        manager._parse_config_data(data)

        # Should only have valid dict configs
        assert "python" in manager.configs
        assert "javascript" in manager.configs
        assert "typescript" not in manager.configs


class TestLSPServerManagerStartServerErrors:
    """Test server start error handling."""

    @pytest.mark.asyncio
    async def test_start_server_raises_when_config_not_found(self):
        """Should raise ServerNotFoundError for missing config."""
        manager = LSPServerManager()

        with pytest.raises(ServerNotFoundError, match="No configuration found"):
            await manager.start_server("nonexistent_language")

    @pytest.mark.asyncio
    async def test_start_server_raises_on_process_failure(self, tmp_path):
        """Should raise ServerStartError when process fails to start."""
        manager = LSPServerManager()

        # Create a config with a command that doesn't exist
        manager.configs["test"] = LSPServerConfig(
            language="test",
            command="/nonexistent/command/that/does/not/exist",
            args=[],
            extensions={},
        )

        with pytest.raises(ServerStartError, match="Failed to start"):
            await manager.start_server("test")


class TestLSPServerManagerStopServerTimeout:
    """Test server stop with timeout handling."""

    @pytest.mark.asyncio
    async def test_stop_server_kills_process_on_timeout(self):
        """Should kill process if terminate times out."""
        manager = LSPServerManager()

        # Create a mock process that times out during terminate
        class MockProcess:
            def __init__(self):
                self.returncode = None
                self.killed = False

            async def wait(self):
                # Simulate timeout by never completing
                await asyncio.sleep(10)

            def terminate(self):
                pass

            def kill(self):
                self.killed = True

            async def wait_after_kill(self):
                # Complete after kill
                self.returncode = -9
                pass

        # Create server with mock process
        mock_process = MockProcess()
        from moai_adk.lsp.server_manager import LSPServer

        server = LSPServer(
            config=LSPServerConfig(
                language="test",
                command="test",
                args=[],
                extensions={},
            ),
            process=mock_process,
        )
        manager.servers["test"] = server

        # Patch asyncio.wait_for to simulate timeout on first wait,
        # then succeed on the second wait (after kill)
        wait_count = [0]

        async def mock_wait_for(coro, timeout):
            wait_count[0] += 1
            if wait_count[0] == 1:
                # First call (terminate) times out
                raise asyncio.TimeoutError()
            else:
                # Second call (after kill) succeeds
                return await coro

        with patch.object(asyncio, "wait_for", side_effect=mock_wait_for):
            with patch.object(mock_process, "wait", side_effect=[mock_process.wait(), mock_process.wait_after_kill()]):
                await manager.stop_server("test")

        # Process should have been killed
        assert mock_process.killed


class TestLSPServerConfigFromDict:
    """Test LSPServerConfig.from_dict edge cases."""

    def test_from_dict_with_minimal_data(self):
        """Should create config with minimal required fields."""
        data = {"command": "test-command"}

        config = LSPServerConfig.from_dict("test", data)

        assert config.language == "test"
        assert config.command == "test-command"
        assert config.args == []
        assert config.extensions == {}

    def test_from_dict_with_all_fields(self):
        """Should create config with all fields."""
        data = {
            "command": "test-command",
            "args": ["--arg1", "--arg2"],
            "extensionToLanguage": {".py": "python"},
        }

        config = LSPServerConfig.from_dict("test", data)

        assert config.command == "test-command"
        assert config.args == ["--arg1", "--arg2"]
        assert config.extensions == {".py": "python"}
