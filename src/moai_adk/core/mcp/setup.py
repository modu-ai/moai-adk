# MCP Setup - Cross-platform npx execution with Windows support

import json
import platform
from pathlib import Path

from rich.console import Console

console = Console()


class MCPSetupManager:
    """Cross-platform MCP Setup Manager with Windows npx support"""

    def __init__(self, project_path: Path):
        self.project_path = project_path
        self.is_windows = platform.system().lower() == "windows"

    def _adapt_command_for_platform(self, command: str) -> str:
        """Adapt command for Windows compatibility.

        Args:
            command: Original command (e.g., "npx")

        Returns:
            Platform-adapted command (e.g., "cmd /c npx" on Windows)
        """
        if self.is_windows and command == "npx":
            return "cmd /c npx"
        return command

    def _adapt_mcp_config_for_platform(self, mcp_config: dict) -> dict:
        """Adapt MCP server commands for the current platform.

        Args:
            mcp_config: Original MCP configuration

        Returns:
            Platform-adapted MCP configuration
        """
        adapted_config = mcp_config.copy()

        if "mcpServers" in adapted_config:
            for server_name, server_config in adapted_config["mcpServers"].items():
                if "command" in server_config:
                    original_command = server_config["command"]
                    adapted_command = self._adapt_command_for_platform(original_command)

                    if adapted_command != original_command:
                        # Need to split command and args for Windows
                        if self.is_windows and original_command == "npx":
                            # Convert "command": "npx", "args": ["-y", "pkg"]
                            # to "command": "cmd", "args": ["/c", "npx", "-y", "pkg"]
                            server_config["command"] = "cmd"
                            server_config["args"] = ["/c", "npx"] + server_config.get(
                                "args", []
                            )
                        else:
                            server_config["command"] = adapted_command

        return adapted_config

    def copy_template_mcp_config(self) -> bool:
        """Copy MCP configuration from package template with platform adaptation"""
        try:
            # Get the package template path
            import moai_adk

            package_path = Path(moai_adk.__file__).parent
            template_mcp_path = package_path / "templates" / ".mcp.json"

            if template_mcp_path.exists():
                # Copy template to project
                project_mcp_path = self.project_path / ".mcp.json"

                # Read template
                with open(template_mcp_path, "r") as f:
                    mcp_config = json.load(f)

                # Adapt for platform
                adapted_config = self._adapt_mcp_config_for_platform(mcp_config)

                # Write adapted config to project
                with open(project_mcp_path, "w") as f:
                    json.dump(adapted_config, f, indent=2)

                server_names = list(adapted_config.get("mcpServers", {}).keys())
                console.print("✅ MCP configuration copied and adapted for platform")

                # Show platform info
                if self.is_windows:
                    console.print(
                        "🪟 Windows platform detected - npx commands wrapped with 'cmd /c'"
                    )

                console.print(f"📋 Configured servers: {', '.join(server_names)}")
                return True
            else:
                console.print("❌ Template MCP configuration not found")
                return False

        except Exception as e:
            console.print(f"❌ Failed to copy MCP configuration: {e}")
            return False

    def setup_mcp_servers(self, selected_servers: list[str]) -> bool:
        """Complete MCP server setup process with platform adaptation"""
        if not selected_servers:
            console.print("ℹ️  No MCP servers selected")
            return True

        console.print("🔧 Setting up MCP servers...")
        return self.copy_template_mcp_config()
