# LSP Server Manager
"""LSP server lifecycle management.

This module manages the lifecycle of Language Server Protocol (LSP) servers,
including starting, stopping, and tracking servers for different languages.
"""

from __future__ import annotations

import asyncio
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any


class ServerNotFoundError(Exception):
    """Error raised when a language server configuration is not found."""

    pass


class ServerStartError(Exception):
    """Error raised when a language server fails to start."""

    pass


@dataclass
class LSPServerConfig:
    """Configuration for an LSP server.

    Attributes:
        language: The language identifier (e.g., 'python', 'typescript').
        command: The command to start the server.
        args: Command-line arguments for the server.
        extensions: Mapping of file extensions to language IDs.
    """

    language: str
    command: str
    args: list[str]
    extensions: dict[str, str]

    @classmethod
    def from_dict(cls, language: str, data: dict[str, Any]) -> LSPServerConfig:
        """Create a config from a dictionary.

        Args:
            language: The language identifier.
            data: Dictionary containing command, args, and extensionToLanguage.

        Returns:
            LSPServerConfig instance.
        """
        return cls(
            language=language,
            command=data["command"],
            args=data.get("args", []),
            extensions=data.get("extensionToLanguage", {}),
        )

    def get_full_command(self) -> list[str]:
        """Get the full command with arguments.

        Returns:
            List containing command and all arguments.
        """
        return [self.command] + self.args


@dataclass
class LSPServer:
    """An LSP server instance.

    Attributes:
        config: The server configuration.
        process: The subprocess running the server (None if not started).
    """

    config: LSPServerConfig
    process: asyncio.subprocess.Process | None = None

    def is_running(self) -> bool:
        """Check if the server is currently running.

        Returns:
            True if the server process exists and hasn't exited.
        """
        if self.process is None:
            return False
        return self.process.returncode is None


class LSPServerManager:
    """Manages LSP server lifecycles.

    This class handles starting, stopping, and tracking LSP servers
    for different programming languages. It loads configuration from
    .lsp.json files and manages server processes.
    """

    def __init__(self) -> None:
        """Initialize the server manager."""
        self.configs: dict[str, LSPServerConfig] = {}
        self.servers: dict[str, LSPServer] = {}

    def load_config(self, config_path: Path) -> None:
        """Load server configurations from a .lsp.json file.

        Args:
            config_path: Path to the .lsp.json file.
        """
        with open(config_path) as f:
            data = json.load(f)

        self._parse_config_data(data)

    def load_config_from_string(self, json_content: str) -> None:
        """Load server configurations from a JSON string.

        Args:
            json_content: JSON string containing the configuration.
        """
        data = json.loads(json_content)
        self._parse_config_data(data)

    def _parse_config_data(self, data: dict[str, Any]) -> None:
        """Parse configuration data from a dictionary.

        Args:
            data: Dictionary containing language configurations.
        """
        # Skip special keys like $schema and _comment
        skip_keys = {"$schema", "_comment"}

        for language, config_data in data.items():
            if language in skip_keys:
                continue

            # Skip if not a valid config dict
            if not isinstance(config_data, dict):
                continue

            # Must have a command
            if "command" not in config_data:
                continue

            self.configs[language] = LSPServerConfig.from_dict(language, config_data)

    def get_language_for_file(self, file_path: str) -> str | None:
        """Get the language for a file based on its extension.

        Args:
            file_path: Path to the file.

        Returns:
            Language identifier or None if not found.
        """
        # Get file extension
        path = Path(file_path)
        ext = path.suffix.lower()

        # Search through all configs to find matching extension
        for language, config in self.configs.items():
            if ext in config.extensions:
                return language

        return None

    def get_server(self, language: str) -> LSPServer | None:
        """Get a running server for a language.

        Args:
            language: The language identifier.

        Returns:
            The LSPServer instance or None if not running.
        """
        return self.servers.get(language)

    async def start_server(self, language: str) -> LSPServer:
        """Start an LSP server for a language.

        Args:
            language: The language identifier.

        Returns:
            The started LSPServer instance.

        Raises:
            ServerNotFoundError: If no configuration exists for the language.
            ServerStartError: If the server fails to start.
        """
        # Check if server is already running
        existing = self.servers.get(language)
        if existing and existing.is_running():
            return existing

        # Get configuration
        config = self.configs.get(language)
        if config is None:
            raise ServerNotFoundError(f"No configuration found for language: {language}")

        # Start the process
        try:
            command = config.get_full_command()
            process = await asyncio.create_subprocess_exec(
                *command,
                stdin=asyncio.subprocess.PIPE,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
            )
        except Exception as e:
            raise ServerStartError(f"Failed to start {language} server: {e}") from e

        # Create and store server
        server = LSPServer(config=config, process=process)
        self.servers[language] = server

        return server

    async def stop_server(self, language: str) -> None:
        """Stop an LSP server for a language.

        Args:
            language: The language identifier.
        """
        server = self.servers.get(language)
        if server is None:
            return

        if server.process is not None and server.is_running():
            server.process.terminate()
            try:
                await asyncio.wait_for(server.process.wait(), timeout=5.0)
            except asyncio.TimeoutError:
                server.process.kill()
                await server.process.wait()

        # Remove from tracking
        self.servers.pop(language, None)

    async def stop_all_servers(self) -> None:
        """Stop all running LSP servers."""
        languages = list(self.servers.keys())
        for language in languages:
            await self.stop_server(language)
