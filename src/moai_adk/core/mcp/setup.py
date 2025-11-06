# MCP Setup - Simplified version for automatic template copying

import json
from pathlib import Path

from rich.console import Console

console = Console()


class MCPSetupManager:
    """Simplified MCP Setup Manager - copies template configuration"""

    def __init__(self, project_path: Path):
        self.project_path = project_path

    def copy_template_mcp_config(self) -> bool:
        """Copy MCP configuration from package template"""
        try:
            # Get the package template path
            import moai_adk
            package_path = Path(moai_adk.__file__).parent
            template_mcp_path = package_path / "templates" / ".mcp.json"

            if template_mcp_path.exists():
                # Copy template to project
                project_mcp_path = self.project_path / ".mcp.json"

                # Read template
                with open(template_mcp_path, 'r') as f:
                    mcp_config = json.load(f)

                # Write to project
                with open(project_mcp_path, 'w') as f:
                    json.dump(mcp_config, f, indent=2)

                server_names = list(mcp_config.get('mcpServers', {}).keys())
                console.print("✅ MCP configuration copied from template")
                console.print(f"📋 Configured servers: {', '.join(server_names)}")
                return True
            else:
                console.print("❌ Template MCP configuration not found")
                return False

        except Exception as e:
            console.print(f"❌ Failed to copy MCP configuration: {e}")
            return False

    def setup_mcp_servers(self, selected_servers: list[str]) -> bool:
        """Complete MCP server setup process - simplified template copy"""
        if not selected_servers:
            console.print("ℹ️  No MCP servers selected")
            return True

        console.print("🔧 Setting up MCP servers...")
        return self.copy_template_mcp_config()
