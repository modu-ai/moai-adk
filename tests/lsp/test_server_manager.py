# LSP Server Manager Tests - RED Phase
"""Tests for LSP server lifecycle management."""

from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from moai_adk.lsp.server_manager import (
    LSPServer,
    LSPServerConfig,
    LSPServerManager,
    ServerNotFoundError,
)


class TestLSPServerConfig:
    """Tests for LSP server configuration."""

    def test_create_config(self):
        """Config should store command and args."""
        config = LSPServerConfig(
            language="python",
            command="pyright-langserver",
            args=["--stdio"],
            extensions={".py": "python", ".pyi": "python"},
        )
        assert config.language == "python"
        assert config.command == "pyright-langserver"
        assert config.args == ["--stdio"]
        assert config.extensions[".py"] == "python"

    def test_config_from_dict(self):
        """Config should be creatable from dictionary."""
        data = {"command": "gopls", "args": ["serve"], "extensionToLanguage": {".go": "go"}}
        config = LSPServerConfig.from_dict("go", data)
        assert config.language == "go"
        assert config.command == "gopls"
        assert config.args == ["serve"]
        assert config.extensions[".go"] == "go"

    def test_config_get_full_command(self):
        """Config should return full command list."""
        config = LSPServerConfig(language="rust", command="rust-analyzer", args=[], extensions={".rs": "rust"})
        assert config.get_full_command() == ["rust-analyzer"]

    def test_config_with_args(self):
        """Config should include args in full command."""
        config = LSPServerConfig(
            language="typescript",
            command="typescript-language-server",
            args=["--stdio"],
            extensions={".ts": "typescript"},
        )
        assert config.get_full_command() == ["typescript-language-server", "--stdio"]


class TestLSPServer:
    """Tests for LSP server instance."""

    def test_create_server(self):
        """Server should store config and process."""
        config = LSPServerConfig(
            language="python", command="pyright-langserver", args=["--stdio"], extensions={".py": "python"}
        )
        server = LSPServer(config=config, process=None)
        assert server.config == config
        assert server.process is None
        assert server.is_running() is False

    def test_server_is_running_with_process(self):
        """Server should report running if process exists."""
        config = LSPServerConfig(
            language="python", command="pyright-langserver", args=["--stdio"], extensions={".py": "python"}
        )
        mock_process = MagicMock()
        mock_process.returncode = None  # Still running
        server = LSPServer(config=config, process=mock_process)
        assert server.is_running() is True

    def test_server_not_running_if_exited(self):
        """Server should report not running if process exited."""
        config = LSPServerConfig(
            language="python", command="pyright-langserver", args=["--stdio"], extensions={".py": "python"}
        )
        mock_process = MagicMock()
        mock_process.returncode = 0  # Exited
        server = LSPServer(config=config, process=mock_process)
        assert server.is_running() is False


class TestLSPServerManager:
    """Tests for LSP server lifecycle management."""

    def test_create_manager(self):
        """Manager should initialize with empty servers."""
        manager = LSPServerManager()
        assert len(manager.servers) == 0

    def test_load_config_from_file(self):
        """Manager should load config from .lsp.json file."""
        lsp_json = {
            "python": {"command": "pyright-langserver", "args": ["--stdio"], "extensionToLanguage": {".py": "python"}},
            "go": {"command": "gopls", "args": ["serve"], "extensionToLanguage": {".go": "go"}},
        }
        with patch("builtins.open", MagicMock()):
            with patch("json.load", return_value=lsp_json):
                manager = LSPServerManager()
                manager.load_config(Path("/fake/.lsp.json"))

        assert "python" in manager.configs
        assert "go" in manager.configs
        assert manager.configs["python"].command == "pyright-langserver"

    def test_get_language_for_file_python(self):
        """Manager should detect Python files."""
        manager = LSPServerManager()
        manager.configs["python"] = LSPServerConfig(
            language="python",
            command="pyright-langserver",
            args=["--stdio"],
            extensions={".py": "python", ".pyi": "python"},
        )
        assert manager.get_language_for_file("/path/to/file.py") == "python"
        assert manager.get_language_for_file("/path/to/stub.pyi") == "python"

    def test_get_language_for_file_typescript(self):
        """Manager should detect TypeScript files."""
        manager = LSPServerManager()
        manager.configs["typescript"] = LSPServerConfig(
            language="typescript",
            command="typescript-language-server",
            args=["--stdio"],
            extensions={".ts": "typescript", ".tsx": "typescriptreact"},
        )
        assert manager.get_language_for_file("/src/app.ts") == "typescript"
        assert manager.get_language_for_file("/src/component.tsx") == "typescript"

    def test_get_language_for_unknown_file(self):
        """Manager should return None for unknown extensions."""
        manager = LSPServerManager()
        manager.configs["python"] = LSPServerConfig(
            language="python", command="pyright-langserver", args=["--stdio"], extensions={".py": "python"}
        )
        assert manager.get_language_for_file("/path/to/data.json") is None

    def test_get_server_not_started(self):
        """Manager should return None if server not started."""
        manager = LSPServerManager()
        assert manager.get_server("python") is None


@pytest.mark.asyncio
class TestLSPServerManagerAsync:
    """Async tests for LSP server management."""

    async def test_start_server(self):
        """Manager should start LSP server process."""
        manager = LSPServerManager()
        manager.configs["python"] = LSPServerConfig(
            language="python", command="pyright-langserver", args=["--stdio"], extensions={".py": "python"}
        )

        mock_process = AsyncMock()
        mock_process.returncode = None
        mock_process.stdin = MagicMock()
        mock_process.stdout = MagicMock()

        with patch("asyncio.create_subprocess_exec", return_value=mock_process):
            server = await manager.start_server("python")

        assert server is not None
        assert server.config.language == "python"
        assert "python" in manager.servers

    async def test_start_server_unknown_language(self):
        """Manager should raise error for unknown language."""
        manager = LSPServerManager()

        with pytest.raises(ServerNotFoundError):
            await manager.start_server("unknown_language")

    async def test_start_server_already_running(self):
        """Manager should return existing server if already running."""
        manager = LSPServerManager()
        manager.configs["python"] = LSPServerConfig(
            language="python", command="pyright-langserver", args=["--stdio"], extensions={".py": "python"}
        )

        mock_process = MagicMock()
        mock_process.returncode = None
        existing_server = LSPServer(config=manager.configs["python"], process=mock_process)
        manager.servers["python"] = existing_server

        server = await manager.start_server("python")
        assert server is existing_server

    async def test_stop_server(self):
        """Manager should stop running server."""
        manager = LSPServerManager()
        manager.configs["python"] = LSPServerConfig(
            language="python", command="pyright-langserver", args=["--stdio"], extensions={".py": "python"}
        )

        mock_process = AsyncMock()
        mock_process.returncode = None
        mock_process.terminate = MagicMock()
        mock_process.wait = AsyncMock()

        server = LSPServer(config=manager.configs["python"], process=mock_process)
        manager.servers["python"] = server

        await manager.stop_server("python")

        mock_process.terminate.assert_called_once()
        assert "python" not in manager.servers

    async def test_stop_server_not_running(self):
        """Manager should do nothing if server not running."""
        manager = LSPServerManager()
        # Should not raise
        await manager.stop_server("python")

    async def test_stop_all_servers(self):
        """Manager should stop all running servers."""
        manager = LSPServerManager()

        for lang in ["python", "typescript", "go"]:
            manager.configs[lang] = LSPServerConfig(
                language=lang, command=f"{lang}-lsp", args=["--stdio"], extensions={f".{lang[:2]}": lang}
            )
            mock_process = AsyncMock()
            mock_process.returncode = None
            mock_process.terminate = MagicMock()
            mock_process.wait = AsyncMock()
            manager.servers[lang] = LSPServer(config=manager.configs[lang], process=mock_process)

        await manager.stop_all_servers()

        assert len(manager.servers) == 0

    async def test_get_server_returns_running(self):
        """Manager should return running server."""
        manager = LSPServerManager()
        manager.configs["python"] = LSPServerConfig(
            language="python", command="pyright-langserver", args=["--stdio"], extensions={".py": "python"}
        )

        mock_process = MagicMock()
        mock_process.returncode = None
        server = LSPServer(config=manager.configs["python"], process=mock_process)
        manager.servers["python"] = server

        result = manager.get_server("python")
        assert result is server


class TestLSPServerManagerConfigLoading:
    """Tests for loading .lsp.json configuration."""

    def test_load_real_lsp_json_format(self):
        """Manager should parse real .lsp.json format."""
        lsp_json_content = """{
            "$schema": "https://code.claude.com/schemas/lsp.json",
            "_comment": "MoAI-ADK Language Server Protocol Configuration",
            "python": {
                "command": "pyright-langserver",
                "args": ["--stdio"],
                "extensionToLanguage": {
                    ".py": "python",
                    ".pyi": "python"
                }
            },
            "typescript": {
                "command": "typescript-language-server",
                "args": ["--stdio"],
                "extensionToLanguage": {
                    ".ts": "typescript",
                    ".tsx": "typescriptreact"
                }
            }
        }"""

        manager = LSPServerManager()
        manager.load_config_from_string(lsp_json_content)

        assert "python" in manager.configs
        assert "typescript" in manager.configs
        assert "$schema" not in manager.configs
        assert "_comment" not in manager.configs

    def test_skip_non_language_keys(self):
        """Manager should skip $schema and _comment keys."""
        lsp_json = {
            "$schema": "https://example.com/schema",
            "_comment": "Some comment",
            "python": {"command": "pyright", "args": [], "extensionToLanguage": {".py": "python"}},
        }
        with patch("builtins.open", MagicMock()):
            with patch("json.load", return_value=lsp_json):
                manager = LSPServerManager()
                manager.load_config(Path("/fake/.lsp.json"))

        assert len(manager.configs) == 1
        assert "python" in manager.configs
