"""Shell environment validator for MoAI-ADK

MoAI-ADK supports:
- PowerShell on Windows
- Bash/Zsh on Linux and macOS
- Bash on WSL (Windows Subsystem for Linux)

Command Prompt (cmd.exe) is not supported.
"""

import os
import platform
from typing import Tuple


def is_wsl() -> bool:
    """Check if running in WSL (Windows Subsystem for Linux).

    Detects WSL via environment variables set by WSL runtime.
    Works for both WSL 1 and WSL 2.

    Returns:
        True if running in WSL, False otherwise
    """
    return (
        "WSL_DISTRO_NAME" in os.environ
        or "WSLENV" in os.environ
        or "WSL_INTEROP" in os.environ
    )


def is_cmd() -> bool:
    """Check if running in Command Prompt (not PowerShell).

    Returns:
        True if running in cmd.exe, False otherwise
    """
    if platform.system() != "Windows":
        return False

    # PowerShell sets PSModulePath environment variable
    # CMD does not have this variable
    has_ps_module_path = "PSModulePath" in os.environ

    # In CMD, PROMPT is usually set to $P$G
    # In PowerShell, it's more complex or not set
    prompt = os.environ.get("PROMPT", "")
    is_cmd_prompt = prompt == "$P$G"

    # If PSModulePath exists, it's PowerShell
    # If it doesn't exist and PROMPT is $P$G, it's CMD
    return not has_ps_module_path and is_cmd_prompt


def validate_shell() -> Tuple[bool, str]:
    """Validate if current shell environment is supported.

    Returns:
        Tuple of (is_supported, error_message)
        - is_supported: True if shell is supported, False otherwise
        - error_message: Empty string if supported, error message otherwise
    """
    # Detect platform
    current_os = platform.system()

    # WSL is treated as Linux (supported)
    if is_wsl():
        current_os = "Linux"

    # Linux and macOS: bash/zsh supported
    if current_os in ("Linux", "Darwin"):
        return (True, "")

    # Windows: Check for Command Prompt (unsupported)
    if current_os == "Windows":
        if is_cmd():
            return (
                False,
                "Command Prompt (cmd.exe) is not supported.\n"
                "Please use PowerShell or Windows Terminal with PowerShell on Windows.",
            )
        # PowerShell is supported
        return (True, "")

    # Other platforms: assume supported
    return (True, "")


def get_shell_info() -> str:
    """Get current shell environment information for debugging.

    Returns:
        String describing the current shell environment
    """
    info_parts = [f"Platform: {platform.system()}"]

    if platform.system() == "Windows":
        if is_wsl():
            info_parts.append("Shell: WSL Bash (supported)")
        elif is_cmd():
            info_parts.append("Shell: Command Prompt (not supported)")
        elif "PSModulePath" in os.environ:
            info_parts.append("Shell: PowerShell (supported)")
        else:
            info_parts.append("Shell: Unknown Windows shell")
    else:
        if is_wsl():
            info_parts.append("Shell: WSL Bash (supported)")
        else:
            shell = os.environ.get("SHELL", "unknown")
            info_parts.append(f"Shell: {shell}")

    return " | ".join(info_parts)
