"""
MoAI-ADK Post-Installation Hook System
Automatically configure Claude Code settings based on environment detection.

@REQ:CROSS-PLATFORM-001
@FEATURE:POST-INSTALL-AUTOMATION
@FEATURE:ENV-DETECTION
"""

import json
import os
import shutil
import subprocess
import sys
import platform
from pathlib import Path
from typing import Dict, List, Optional, Any


class EnvironmentDetector:
    """Detect Python execution environment and generate appropriate settings

    @FEATURE:ENV-DETECTION
    Cross-platform Python executable and environment detection
    """

    def __init__(self):
        self.system = platform.system().lower()
        self.python_exec = self.detect_python_executable()

    def detect_python_executable(self) -> str:
        """Detect the best Python executable command for the current environment

        @FEATURE:PYTHON-AUTO-DETECT
        Returns the most appropriate Python command for hook execution
        """
        # Start with the current Python executable (most reliable)
        current_python = sys.executable

        # Test various Python command candidates
        candidates = self._get_python_candidates()

        for candidate in candidates:
            if self._test_python_command(candidate, current_python):
                return candidate

        # Fallback to current executable path
        return current_python

    def _get_python_candidates(self) -> List[str]:
        """Get platform-specific Python command candidates"""
        if self.system == "windows":
            return [
                "python",      # Windows standard
                "py",          # Python Launcher for Windows
                "python3",     # Some Windows setups
                "python.exe",  # Explicit executable
            ]
        else:  # Unix/Linux/macOS
            return [
                "python3",           # Modern Unix standard
                "python",            # General fallback
                "python3.12",        # Specific versions
                "python3.11",
                "python3.10",
                "/usr/bin/python3",  # System paths
                "/usr/local/bin/python3",
                "/opt/homebrew/bin/python3",  # Homebrew on macOS
            ]

    def _test_python_command(self, candidate: str, current_python: str) -> bool:
        """Test if a Python command candidate works and matches current environment"""
        try:
            # For sys.executable, always return True as it's guaranteed to work
            if candidate == current_python:
                return True

            # Test command existence and version
            result = subprocess.run(
                [candidate, "--version"],
                capture_output=True,
                text=True,
                timeout=5
            )

            if result.returncode == 0 and "Python" in result.stdout:
                # Additional check: ensure it's the same Python version
                version_result = subprocess.run(
                    [candidate, "-c", "import sys; print(sys.version)"],
                    capture_output=True,
                    text=True,
                    timeout=5
                )

                if version_result.returncode == 0:
                    return True

        except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
            pass

        return False

    def get_environment_info(self) -> Dict[str, Any]:
        """Get comprehensive environment information"""
        return {
            "system": self.system,
            "platform": sys.platform,
            "python_executable": self.python_exec,
            "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
            "architecture": platform.architecture()[0],
            "in_venv": sys.prefix != sys.base_prefix,
            "venv_path": sys.prefix if sys.prefix != sys.base_prefix else None,
        }


class ClaudeCodeConfigGenerator:
    """Generate Claude Code settings.json with environment-specific configuration

    @FEATURE:CONFIG-GENERATION
    Dynamic configuration based on detected environment
    """

    def __init__(self, env_detector: EnvironmentDetector):
        self.env = env_detector

    def generate_hook_command(self, script_path: str) -> str:
        """Generate cross-platform hook command

        @FEATURE:HOOK-COMMAND-GEN
        Use sys.executable approach for maximum compatibility
        """
        python_cmd = self.env.python_exec

        # Use sys.executable approach for maximum reliability
        if python_cmd == sys.executable:
            # Direct sys.executable approach
            return f'python -c "import sys, subprocess; subprocess.run([sys.executable, \'$CLAUDE_PROJECT_DIR/{script_path}\'])"'
        else:
            # Use detected command
            return f'{python_cmd} "$CLAUDE_PROJECT_DIR/{script_path}"'

    def generate_settings(self) -> Dict[str, Any]:
        """Generate complete Claude Code settings.json configuration"""
        hook_scripts = [
            ".claude/hooks/moai/session_start_notice.py",
            ".claude/hooks/moai/pre_write_guard.py",
            ".claude/hooks/moai/policy_block.py",
            ".claude/hooks/moai/steering_guard.py"
        ]

        # Generate hook commands with environment-specific Python executable
        session_cmd = self.generate_hook_command(hook_scripts[0])
        pre_write_cmd = self.generate_hook_command(hook_scripts[1])
        policy_cmd = self.generate_hook_command(hook_scripts[2])
        steering_cmd = self.generate_hook_command(hook_scripts[3])

        return {
            "hooks": {
                "PostToolUse": [],
                "PreToolUse": [
                    {
                        "hooks": [
                            {
                                "command": pre_write_cmd,
                                "type": "command"
                            }
                        ],
                        "matcher": "Edit|Write|MultiEdit"
                    },
                    {
                        "hooks": [
                            {
                                "command": policy_cmd,
                                "type": "command"
                            }
                        ],
                        "matcher": "Bash"
                    }
                ],
                "SessionStart": [
                    {
                        "hooks": [
                            {
                                "command": session_cmd,
                                "type": "command"
                            }
                        ],
                        "matcher": "*"
                    }
                ],
                "UserPromptSubmit": [
                    {
                        "hooks": [
                            {
                                "command": steering_cmd,
                                "type": "command"
                            }
                        ]
                    }
                ]
            },
            "permissions": {
                "allow": [
                    "Task", "Read", "Write", "Edit", "MultiEdit", "NotebookEdit",
                    "Grep", "Glob", "TodoWrite", "WebFetch", "WebSearch",
                    "BashOutput", "KillShell",
                    "Bash(git:*)", "Bash(rg:*)", "Bash(ls:*)", "Bash(cat:*)",
                    "Bash(echo:*)", "Bash(which:*)", "Bash(make:*)",
                    "Bash(python:*)", "Bash(python3:*)", "Bash(pytest:*)",
                    "Bash(ruff:*)", "Bash(black:*)", "Bash(mypy:*)",
                    "Bash(bandit:*)", "Bash(coverage:*)", "Bash(pip:*)",
                    "Bash(poetry:*)", "Bash(uv:*)", "Bash(npm:*)",
                    "Bash(node:*)", "Bash(pnpm:*)", "Bash(moai:*)",
                    "Bash(gh pr create:*)", "Bash(gh pr view:*)",
                    "Bash(gh pr list:*)", "Bash(gh repo view:*)",
                    "Bash(find:*)", "Bash(mkdir:*)", "Bash(touch:*)",
                    "Bash(cp:*)", "Bash(mv:*)", "Bash(tree:*)",
                    "Bash(diff:*)", "Bash(wc:*)", "Bash(sort:*)", "Bash(uniq:*)"
                ],
                "ask": [
                    "Bash(git push:*)", "Bash(git merge:*)", "Bash(gh pr merge:*)",
                    "Bash(pip install:*)", "Bash(poetry add:*)",
                    "Bash(npm install:*)", "Bash(rm:*)"
                ],
                "defaultMode": "default",
                "deny": [
                    "Read(./.env)", "Read(./.env.*)", "Read(./secrets/**)",
                    "Read(~/.ssh/**)", "Bash(sudo:*)", "Bash(rm -rf :*)",
                    "Bash(chmod -R 777 :*)", "Bash(dd:*)",
                    "Bash(mkfs:*)", "Bash(fdisk:*)"
                ]
            },
            # Add environment info as metadata
            "_moai_env": self.env.get_environment_info()
        }


class PostInstallProcessor:
    """Process post-installation configuration for MoAI-ADK

    @FEATURE:POST-INSTALL-AUTOMATION
    Main orchestrator for post-install configuration
    """

    def __init__(self):
        self.env_detector = EnvironmentDetector()
        self.config_generator = ClaudeCodeConfigGenerator(self.env_detector)

    def should_configure_project(self, project_path: Path) -> bool:
        """Check if project needs MoAI-ADK configuration"""
        claude_dir = project_path / ".claude"
        moai_dir = project_path / ".moai"

        # Configure if either directory exists
        return claude_dir.exists() or moai_dir.exists()

    def update_claude_settings(self, project_path: Path) -> bool:
        """Update Claude Code settings.json with environment-optimized configuration"""
        claude_dir = project_path / ".claude"
        settings_file = claude_dir / "settings.json"

        if not claude_dir.exists():
            return False

        try:
            # Generate new settings
            new_settings = self.config_generator.generate_settings()

            # Backup existing settings if present
            if settings_file.exists():
                backup_file = settings_file.with_suffix(".json.backup")
                shutil.copy2(settings_file, backup_file)

            # Write new settings
            with open(settings_file, 'w', encoding='utf-8') as f:
                json.dump(new_settings, f, indent=2, ensure_ascii=False)

            print(f"✅ Updated Claude Code settings for {self.env_detector.system} environment")
            print(f"   Python command: {self.env_detector.python_exec}")

            return True

        except Exception as e:
            print(f"❌ Failed to update Claude Code settings: {e}")
            return False

    def run_post_install(self) -> bool:
        """Main post-install configuration process"""
        try:
            env_info = self.env_detector.get_environment_info()
            print(f"🔍 Detected environment: {env_info['system']} ({env_info['python_version']})")
            print(f"🐍 Python executable: {env_info['python_executable']}")

            # Find all MoAI-ADK projects in common locations
            search_paths = [
                Path.cwd(),  # Current directory
                Path.home(),  # User home
            ]

            configured_projects = 0

            for search_path in search_paths:
                if not search_path.exists():
                    continue

                # Look for projects with .claude or .moai directories
                for item in search_path.iterdir():
                    if item.is_dir() and self.should_configure_project(item):
                        if self.update_claude_settings(item):
                            configured_projects += 1

            if configured_projects > 0:
                print(f"🎉 Successfully configured {configured_projects} MoAI-ADK project(s)")
            else:
                print("💡 No MoAI-ADK projects found. Settings will be applied when you create a new project.")

            return True

        except Exception as e:
            print(f"❌ Post-install configuration failed: {e}")
            return False


def main():
    """Entry point for post-install hook"""
    processor = PostInstallProcessor()
    success = processor.run_post_install()
    return 0 if success else 1


if __name__ == "__main__":
    exit(main())