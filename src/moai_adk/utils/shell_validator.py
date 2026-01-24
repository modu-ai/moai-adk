"""Shell environment validator for MoAI-ADK

MoAI-ADK supports PowerShell only on Windows.
WSL (Windows Subsystem for Linux) and Command Prompt (cmd.exe) are not supported.
"""

import os
import platform
from typing import Tuple


def is_wsl() -> bool:
    """Check if running in WSL (Windows Subsystem for Linux)

    Returns:
        True if running in WSL, False otherwise
    """
    # WSL sets these environment variables
    return "WSL_DISTRO_NAME" in os.environ or "WSLENV" in os.environ or "WSL_INTEROP" in os.environ


def is_cmd() -> bool:
    """Check if running in Command Prompt (not PowerShell)

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
    """Validate if current shell environment is supported

    Returns:
        Tuple of (is_supported, error_message)
        - is_supported: True if shell is supported, False otherwise
        - error_message: Empty string if supported, error message otherwise
    """
    if is_wsl():
        return (
            False,
            "WSL (Windows Subsystem for Linux) is not supported.\n"
            "Please use PowerShell or Windows Terminal with PowerShell on Windows.",
        )

    if is_cmd():
        return (
            False,
            "Command Prompt (cmd.exe) is not supported.\n"
            "Please use PowerShell or Windows Terminal with PowerShell on Windows.",
        )

    return True, ""


def get_shell_info() -> str:
    """Get current shell environment information for debugging

    Returns:
        String describing the current shell environment
    """
    info_parts = [f"Platform: {platform.system()}"]

    if platform.system() == "Windows":
        if is_wsl():
            info_parts.append("Shell: WSL (not supported)")
        elif is_cmd():
            info_parts.append("Shell: Command Prompt (not supported)")
        elif "PSModulePath" in os.environ:
            info_parts.append("Shell: PowerShell (supported)")
        else:
            info_parts.append("Shell: Unknown Windows shell")
    else:
        shell = os.environ.get("SHELL", "unknown")
        info_parts.append(f"Shell: {shell}")

    return " | ".join(info_parts)
