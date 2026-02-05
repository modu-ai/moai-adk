"""Cross-platform hook context builder for MoAI-ADK

Builds template variable context for settings.json hook commands.
Shared by both `moai init` (phase_executor.py) and `moai update` (update.py)
to ensure consistent behavior across all platforms.

Supported platforms:
- macOS (zsh, bash)
- Linux (bash, zsh)
- WSL (bash, zsh) with Windows path normalization
- Windows PowerShell (cmd.exe PATH expansion)
"""

import os
import platform
import shutil
from dataclasses import dataclass


@dataclass
class HookContext:
    """Platform-specific hook configuration values."""

    project_dir: str
    project_dir_win: str
    project_dir_unix: str
    hook_shell_prefix: str
    hook_shell_suffix: str
    mcp_shell: str
    statusline_command: str
    is_windows: bool
    is_wsl: bool
    detected_shell: str


def _is_wsl() -> bool:
    """Check if running in WSL (Windows Subsystem for Linux).

    Returns:
        True if running in WSL, False otherwise
    """
    return "WSL_DISTRO_NAME" in os.environ or "WSLENV" in os.environ or "WSL_INTEROP" in os.environ


def _detect_shell() -> str:
    """Detect the user's default shell path.

    Priority order:
    1. User's default shell from $SHELL env var
    2. bash (most common login shell)
    3. sh (POSIX standard)
    4. Common hardcoded paths
    5. /bin/sh (ultimate fallback)

    Returns:
        Absolute path to the detected shell executable
    """
    # Get user's default shell from environment
    user_shell = os.environ.get("SHELL", "")

    # Verify user's shell exists and is executable
    if user_shell and os.path.isfile(user_shell) and os.access(user_shell, os.X_OK):
        return user_shell

    # Try bash first (most common, supports -l flag)
    bash_path = shutil.which("bash")
    if bash_path:
        return bash_path

    # Try sh (POSIX standard, supports -l flag)
    sh_path = shutil.which("sh")
    if sh_path:
        return sh_path

    # Last resort - check common paths
    for path in ["/bin/bash", "/bin/sh", "/usr/bin/bash", "/usr/bin/sh"]:
        if os.path.isfile(path) and os.access(path, os.X_OK):
            return path

    # Ultimate fallback (shouldn't happen on Unix systems)
    return "/bin/sh"


# PATH augmentation for macOS/Linux
# Ensures common binary directories (uv, cargo, homebrew) are in PATH
# even if login shell profile doesn't include them.
# This runs AFTER login shell profile (-l) so user PATH takes precedence.
_PATH_AUGMENT = "export PATH=$HOME/.local/bin:$HOME/.cargo/bin:/opt/homebrew/bin:/usr/local/bin:$PATH; "

# WSL path normalization snippet (bash)
# Converts Windows-format CLAUDE_PROJECT_DIR to WSL /mnt/ format
# Example: C:\Users\goos\project → /mnt/c/Users/goos/project
# Handles both backslash (C:\) and forward slash (C:/) separators
_WSL_PATH_NORMALIZE = (
    '_D="$CLAUDE_PROJECT_DIR"; '
    'if [[ "$_D" =~ ^[a-zA-Z]:[\\\\/] ]]; then '
    'CLAUDE_PROJECT_DIR="/mnt/${_D:0:1,,}/${_D:3//\\\\//}"; '
    "export CLAUDE_PROJECT_DIR; "
    "fi; "
)


def build_hook_context() -> HookContext:
    """Build cross-platform hook context for template variable substitution.

    Detects the current platform and builds appropriate values for:
    - PROJECT_DIR: Runtime environment variable reference
    - HOOK_SHELL_PREFIX/SUFFIX: Login shell wrapper for PATH loading
    - STATUSLINE_COMMAND: Platform-appropriate statusline command
    - MCP_SHELL: Shell for MCP server commands

    Platform behavior:
    - Windows: Direct execution, %CLAUDE_PROJECT_DIR% expansion via cmd.exe
    - macOS/Linux: Login shell wrapper with absolute shell path
    - WSL: Login shell wrapper with inline Windows path normalization

    Returns:
        HookContext with all platform-specific values
    """
    is_windows = platform.system() == "Windows"
    wsl = _is_wsl()

    # PROJECT_DIR: Cross-platform forward slash path with trailing separator
    # Standard since v1.8.0 - works on all modern platforms
    if is_windows:
        hook_project_dir = "%CLAUDE_PROJECT_DIR%/"
    else:
        hook_project_dir = "$CLAUDE_PROJECT_DIR/"

    # Deprecated path variables (kept for backward compatibility, removal in v2.0.0)
    hook_project_dir_win = "%CLAUDE_PROJECT_DIR%\\" if is_windows else "$CLAUDE_PROJECT_DIR/"
    hook_project_dir_unix = "%CLAUDE_PROJECT_DIR%/" if is_windows else "$CLAUDE_PROJECT_DIR/"

    # HOOK_SHELL_PREFIX & HOOK_SHELL_SUFFIX
    # Windows: Direct execution (PATH loaded from system environment)
    # Unix: Login shell wrapper to ensure PATH is loaded
    # WSL: Login shell wrapper + inline Windows path normalization
    detected_shell = ""
    if is_windows:
        hook_shell_prefix = ""
        hook_shell_suffix = ""
    else:
        detected_shell = _detect_shell()

        if wsl:
            # WSL: PATH augmentation + path normalization before command execution
            # This handles CLAUDE_PROJECT_DIR set in Windows format (C:\...)
            hook_shell_prefix = f"{detected_shell} -l -c '{_PATH_AUGMENT}{_WSL_PATH_NORMALIZE}"
        else:
            # macOS/Linux: PATH augmentation to ensure uv/cargo binaries are findable
            hook_shell_prefix = f"{detected_shell} -l -c '{_PATH_AUGMENT}"

        hook_shell_suffix = "'"

    # MCP_SHELL: Cross-platform shell for MCP server commands
    if is_windows:
        mcp_shell = "cmd"
    else:
        mcp_shell = "${SHELL:-/bin/bash}"

    # STATUSLINE_COMMAND: Use uv run for reliable tool resolution on all platforms
    # uv run --no-sync ensures moai-adk is found via uv tool management
    statusline_command = "uv run --no-sync moai-adk statusline"

    return HookContext(
        project_dir=hook_project_dir,
        project_dir_win=hook_project_dir_win,
        project_dir_unix=hook_project_dir_unix,
        hook_shell_prefix=hook_shell_prefix,
        hook_shell_suffix=hook_shell_suffix,
        mcp_shell=mcp_shell,
        statusline_command=statusline_command,
        is_windows=is_windows,
        is_wsl=wsl,
        detected_shell=detected_shell,
    )


def build_template_context(hook_ctx: HookContext) -> dict[str, str]:
    """Convert HookContext to template variable dictionary.

    Returns only the hook/path related variables. Caller should merge
    with project-specific variables (version, language, etc.).

    Args:
        hook_ctx: HookContext from build_hook_context()

    Returns:
        Dictionary of template variable names to values
    """
    return {
        "PROJECT_DIR": hook_ctx.project_dir,
        "PROJECT_DIR_WIN": hook_ctx.project_dir_win,
        "PROJECT_DIR_UNIX": hook_ctx.project_dir_unix,
        "HOOK_SHELL_PREFIX": hook_ctx.hook_shell_prefix,
        "HOOK_SHELL_SUFFIX": hook_ctx.hook_shell_suffix,
        "MCP_SHELL": hook_ctx.mcp_shell,
        "STATUSLINE_COMMAND": hook_ctx.statusline_command,
    }
