# @CODE:MCP-SETUP-001 | SPEC: SPEC-MCP-SETUP-001/spec.md
"""MCP (Model Context Protocol) Setup and Configuration

Handles MCP server installation, configuration, and management for MoAI-ADK projects.
"""

import json
import subprocess
from pathlib import Path
from typing import Dict, List

from rich.console import Console
from rich.progress import BarColumn, Progress, SpinnerColumn, TextColumn

console = Console()


class MCPSetupManager:
    """Manages MCP server setup and configuration"""

    def __init__(self, project_path: Path):
        self.project_path = project_path
        self.settings_path = project_path / ".claude" / "settings.json"

    def get_npm_global_path(self) -> str:
        """Get npm global modules path"""
        try:
            result = subprocess.run(
                ["npm", "root", "-g"],
                capture_output=True,
                text=True,
                check=True
            )
            return result.stdout.strip()
        except subprocess.CalledProcessError:
            return "/usr/local/lib/node_modules"

    def check_npm_package_installed(self, package: str) -> bool:
        """Check if npm package is installed globally"""
        try:
            result = subprocess.run(
                ["npm", "list", "-g", package],
                capture_output=True,
                text=True,
                check=True
            )
            return package in result.stdout
        except subprocess.CalledProcessError:
            return False

    def install_mcp_server(self, package: str, display_name: str) -> bool:
        """Install MCP server globally"""
        console.print(f"📦 Installing {display_name}...")

        try:
            subprocess.run(
                ["npm", "install", "-g", package],
                check=True,
                capture_output=True
            )
            console.print(f"✅ {display_name} installed successfully")
            return True
        except subprocess.CalledProcessError as e:
            console.print(f"❌ Failed to install {display_name}: {e}")
            return False

    def detect_mcp_servers(self, servers: List[str]) -> Dict[str, bool]:
        """Detect installation status of MCP servers"""
        status = {}

        package_mapping = {
            "context7": "@upstash/context7-mcp",
            "playwright": "@playwright/mcp",
            "sequential-thinking": "@modelcontextprotocol/server-sequential-thinking"
        }

        for server in servers:
            package = package_mapping.get(server)
            if package:
                status[server] = self.check_npm_package_installed(package)
            else:
                status[server] = False

        return status

    def install_mcp_servers(self, servers: List[str]) -> Dict[str, bool]:
        """Install specified MCP servers"""
        results = {}

        package_mapping = {
            "context7": ("@upstash/context7-mcp", "Context7 MCP"),
            "playwright": ("@playwright/mcp", "Playwright MCP"),
            "sequential-thinking": ("@modelcontextprotocol/server-sequential-thinking", "Sequential Thinking MCP")
        }

        for server in servers:
            if server in package_mapping:
                package, display_name = package_mapping[server]
                results[server] = self.install_mcp_server(package, display_name)
            else:
                console.print(f"⚠️  Unknown MCP server: {server}")
                results[server] = False

        return results

    def generate_mcp_config(self, installed_servers: Dict[str, bool]) -> Dict:
        """Generate MCP configuration following Microsoft MCP standard"""
        config = {"servers": {}}

        # Context7 MCP
        if installed_servers.get("context7", False):
            config["servers"]["context7"] = {
                "type": "stdio",
                "command": "npx",
                "args": [
                    "-y",
                    "@upstash/context7-mcp"
                ],
                "env": {}
            }


        # Playwright MCP
        if installed_servers.get("playwright", False):
            config["servers"]["playwright"] = {
                "type": "stdio",
                "command": "npx",
                "args": [
                    "-y",
                    "@playwright/mcp"
                ],
                "env": {}
            }

        # Sequential Thinking MCP
        if installed_servers.get("sequential-thinking", False):
            config["servers"]["sequential-thinking"] = {
                "type": "stdio",
                "command": "npx",
                "args": [
                    "-y",
                    "@modelcontextprotocol/server-sequential-thinking"
                ],
                "env": {}
            }

        return config

    def backup_settings(self) -> bool:
        """Backup existing settings file (legacy - not used for MCP)"""
        # Settings backup is no longer needed for MCP setup
        # We only create .mcp.json in project root
        return True

    def update_mcp_file(self, mcp_config: Dict) -> bool:
        """Update MCP configuration file following Microsoft MCP standard (legacy - not used)"""
        # This method is no longer used - we only create .mcp.json in project root
        return True

    def create_project_mcp_file(self, mcp_config: Dict) -> bool:
        """Create .mcp.json file in project root for Claude Code compatibility"""
        try:
            # Create .mcp.json path in project root
            mcp_root_path = self.project_path / ".mcp.json"

            # Write MCP configuration to project root
            with open(mcp_root_path, 'w') as f:
                json.dump(mcp_config, f, indent=2)

            console.print("✅ Project MCP file created (.mcp.json)")
            return True

        except Exception as e:
            console.print(f"❌ Failed to create project MCP file: {e}")
            return False

    def update_settings_file(self, mcp_config: Dict) -> bool:
        """Update Claude Code settings file with MCP configuration (legacy - not used)"""
        # This method is no longer used - we only create .mcp.json in project root
        return True

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

    def setup_mcp_servers(self, selected_servers: List[str]) -> bool:
        """Complete MCP server setup process"""
        if not selected_servers:
            console.print("ℹ️  No MCP servers selected")
            return True

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]"),
            BarColumn(),
            console=console
        ) as progress:
            task = progress.add_task("Setting up MCP servers...", total=100)

            # Step 1: Detect installation status
            progress.update(task, completed=10, description="Detecting server installation...")
            installed_status = self.detect_mcp_servers(selected_servers)

            # Step 2: Install missing servers
            missing_servers = [s for s in selected_servers if not installed_status.get(s, False)]
            if missing_servers:
                progress.update(task, completed=30, description="Installing missing servers...")
                install_results = self.install_mcp_servers(missing_servers)

                # Update status with installation results
                for server, result in install_results.items():
                    installed_status[server] = result

            # Step 3: Generate configuration
            progress.update(task, completed=70, description="Generating configuration...")
            mcp_config = self.generate_mcp_config(installed_status)

            # Step 4: Create .mcp.json in project root for Claude Code compatibility only
            progress.update(task, completed=80, description="Creating project MCP file...")
            success = self.create_project_mcp_file(mcp_config)

            # Step 5: Final verification
            progress.update(task, completed=100, description="Final verification...")

            if success:
                console.print("\n🎉 MCP setup completed successfully!")
                console.print(f"📋 Configured servers: {', '.join([s for s, installed in installed_status.items() if installed])}")


                return True
            else:
                console.print("\n❌ MCP setup failed")
                return False

    def get_installed_servers(self) -> List[str]:
        """Get list of currently installed MCP servers"""
        all_servers = ["context7", "playwright", "sequential-thinking"]
        status = self.detect_mcp_servers(all_servers)
        return [server for server, installed in status.items() if installed]

    def verify_mcp_configuration(self) -> Dict[str, bool]:
        """Verify MCP configuration and server availability"""
        mcp_path = self.project_path / ".claude" / "mcp.json"

        if not mcp_path.exists():
            return {"config_exists": False, "servers_configured": 0}

        try:
            with open(mcp_path, 'r') as f:
                config = json.load(f)

            servers = config.get("servers", {})
            installed_servers = self.get_installed_servers()

            verification = {
                "config_exists": True,
                "servers_configured": len(servers),
                "servers_available": len(installed_servers),
                "server_status": {}
            }

            for server_name in servers.keys():
                verification["server_status"][server_name] = server_name in installed_servers

            return verification

        except Exception as e:
            return {"config_exists": False, "error": str(e)}
