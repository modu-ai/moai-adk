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
        """Complete MCP server setup with selective server inclusion

        Args:
            selected_servers: List of MCP servers to include (e.g., ["context7", "notion"])

        Returns:
            True if setup successful, False otherwise
        """
        if not selected_servers:
            console.print("ℹ️  No MCP servers selected")
            return True

        console.print(f"🔧 Setting up MCP servers: {', '.join(selected_servers)}...")

        try:
            # Get the package template path
            import moai_adk

            package_path = Path(moai_adk.__file__).parent
            template_mcp_path = package_path / "templates" / ".mcp.json"

            if not template_mcp_path.exists():
                console.print("❌ Template MCP configuration not found")
                return False

            # Read full template
            with open(template_mcp_path, "r") as f:
                full_mcp_config = json.load(f)

            # Filter only selected servers
            filtered_config = {"mcpServers": {}}
            all_servers = full_mcp_config.get("mcpServers", {})

            for server_name in selected_servers:
                if server_name in all_servers:
                    filtered_config["mcpServers"][server_name] = all_servers[server_name]
                else:
                    console.print(f"[yellow]⚠️  Unknown server: {server_name}[/yellow]")

            if not filtered_config["mcpServers"]:
                console.print("❌ No valid MCP servers found in selection")
                return False

            # Adapt for platform
            adapted_config = self._adapt_mcp_config_for_platform(filtered_config)

            # Write adapted config to project
            project_mcp_path = self.project_path / ".mcp.json"
            with open(project_mcp_path, "w") as f:
                json.dump(adapted_config, f, indent=2)

            server_count = len(adapted_config.get("mcpServers", {}))
            console.print(f"✅ MCP configuration created with {server_count} server(s)")

            # Show platform info
            if self.is_windows:
                console.print(
                    "🪟 Windows platform detected - npx commands wrapped with 'cmd /c'"
                )

            # Show configured servers
            server_names = list(adapted_config.get("mcpServers", {}).keys())
            console.print(f"📋 Configured servers: {', '.join(server_names)}")

            return True

        except Exception as e:
            console.print(f"❌ Failed to setup MCP servers: {e}")
            return False
