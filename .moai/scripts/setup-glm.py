#!/usr/bin/env python3
"""GLM Configuration Setup Script for MoAI-ADK

Configures GLM (OpenAI-compatible API endpoint) for Claude models integration.

Usage:
    uv run {{PROJECT_DIR}}/.moai/scripts/setup-glm.py <api-token>

Behavior:
    1. Saves API token to .env.glm (keeps token secure, .gitignore protected)
    2. Updates .claude/settings.local.json with GLM model environment variables
    3. Token persists across sessions (no re-entry needed)
    4. Next Claude Code restart loads configuration automatically

Example:
    $ /moai:0-project --glm-on abc123xyz...
    API token saved to ./.env.glm
    GLM model configuration updated in ./.claude/settings.local.json
    GLM configuration complete!
"""

import json
import sys
from pathlib import Path


def setup_glm(api_token: str, project_root: Path | None = None) -> bool:
    """Configure GLM integration in the project

    Args:
        api_token: GLM API authentication token
        project_root: Project root directory (default: current directory)

    Returns:
        True if successful, False otherwise
    """
    if project_root is None:
        project_root = Path.cwd()
    else:
        project_root = Path(project_root).resolve()

    try:
        # 1. Create .env.glm with API token
        env_glm_path = project_root / ".env.glm"
        env_glm_path.write_text(f"ANTHROPIC_AUTH_TOKEN={api_token}\n")

        # Set secure permissions (user read/write only)
        env_glm_path.chmod(0o600)

        print(f"✓ API token saved to: {env_glm_path.relative_to(project_root)}")

        # 2. Add .env.glm to .gitignore if needed
        gitignore_path = project_root / ".gitignore"
        glm_entry = ".env.glm"

        if gitignore_path.exists():
            gitignore_content = gitignore_path.read_text()
            if glm_entry not in gitignore_content:
                gitignore_path.write_text(gitignore_content.rstrip() + f"\n{glm_entry}\n")
        else:
            gitignore_path.write_text(f"{glm_entry}\n")

        # 3. Update .claude/settings.local.json with GLM model configuration
        settings_path = project_root / ".claude" / "settings.local.json"

        if not settings_path.exists():
            print(f"❌ Error: {settings_path} not found")
            print("   Run 'moai-adk init' first to initialize the project")
            return False

        # Load existing settings
        with open(settings_path, "r", encoding="utf-8") as f:
            settings = json.load(f)

        # Add GLM environment variables with authentication token
        settings["env"] = {
            "ANTHROPIC_AUTH_TOKEN": api_token,
            "ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic",
            "ANTHROPIC_DEFAULT_HAIKU_MODEL": "glm-4.5-air",
            "ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
            "ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-4.7",
        }

        # Write back to file
        with open(settings_path, "w", encoding="utf-8") as f:
            json.dump(settings, f, indent=2, ensure_ascii=False)

        print(f"✓ GLM model configuration updated in: {settings_path.relative_to(project_root)}")
        print()
        print("Configured environment variables:")
        print(f"   • ANTHROPIC_AUTH_TOKEN: {api_token[:20]}... (loaded from .env.glm)")
        print("   • ANTHROPIC_BASE_URL: https://api.z.ai/api/anthropic")
        print("   • ANTHROPIC_DEFAULT_HAIKU_MODEL: glm-4.5-air")
        print("   • ANTHROPIC_DEFAULT_SONNET_MODEL: glm-4.7")
        print("   • ANTHROPIC_DEFAULT_OPUS_MODEL: glm-4.7")
        print()
        print("✓ GLM configuration complete!")
        print("Note: Restart Claude Code to load the new configuration automatically.")

        return True

    except json.JSONDecodeError as e:
        print(f"❌ Error: Invalid JSON in {settings_path}")
        print(f"   Details: {e}")
        return False

    except Exception as e:
        print(f"❌ Error during GLM configuration: {e}")
        return False


def main() -> None:
    """Main entry point"""
    if len(sys.argv) < 2:
        print("Usage: uv run {{PROJECT_DIR}}/.moai/scripts/setup-glm.py <api-token>")
        print()
        print("Arguments:")
        print("  api-token   GLM API authentication token")
        print()
        print("Example:")
        print("  uv run {{PROJECT_DIR}}/.moai/scripts/setup-glm.py abc123xyz...")
        sys.exit(1)

    api_token = sys.argv[1]

    # Validate token (basic check)
    if not api_token or len(api_token) < 10:
        print("❌ Error: API token appears to be invalid (too short)")
        sys.exit(1)

    success = setup_glm(api_token)
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
