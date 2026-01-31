"""Shell environment validator for MoAI-ADK

MoAI-ADK supports:
- PowerShell on Windows
- Bash/Zsh on Linux and macOS
- Bash on WSL (Windows Subsystem for Linux)

Command Prompt (cmd.exe) is not supported.
"""

import os
import platform
import subprocess
from dataclasses import dataclass
from pathlib import Path
from typing import Tuple


def is_wsl() -> bool:
    """Check if running in WSL (Windows Subsystem for Linux).

    Detects WSL via environment variables set by WSL runtime.
    Works for both WSL 1 and WSL 2.

    Returns:
        True if running in WSL, False otherwise
    """
    return "WSL_DISTRO_NAME" in os.environ or "WSLENV" in os.environ or "WSL_INTEROP" in os.environ


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


# =============================================================================
# PATH Diagnostics for WSL2/Linux
# =============================================================================


@dataclass
class PathDiagnostics:
    """PATH configuration diagnostics result."""

    local_bin_in_path: bool
    local_bin_exists: bool
    shell_type: str  # bash, zsh, fish, unknown
    config_file: str  # ~/.bashrc, ~/.profile, ~/.zshrc, etc.
    is_login_shell_env: bool  # Whether current env uses login shell
    in_login_shell_path: bool  # Whether ~/.local/bin is in login shell PATH
    current_path: str
    recommended_fix: str
    auto_fixable: bool
    in_system_path: bool = True  # macOS: whether ~/.local/bin is in /etc/paths.d/


def get_user_shell() -> str:
    """Get the user's default shell.

    Returns:
        Shell name: 'bash', 'zsh', 'fish', or 'unknown'
    """
    shell_path = os.environ.get("SHELL", "")
    if "zsh" in shell_path:
        return "zsh"
    elif "bash" in shell_path:
        return "bash"
    elif "fish" in shell_path:
        return "fish"
    return "unknown"


def get_shell_config_file(shell: str, prefer_login: bool = True) -> str:
    """Get the appropriate shell configuration file.

    Args:
        shell: Shell type ('bash', 'zsh', 'fish')
        prefer_login: Whether to prefer login shell config files (for WSL)

    Returns:
        Path to the shell configuration file
    """
    home = Path.home()

    if shell == "zsh":
        # zsh loads: .zshenv (always) -> .zprofile (login) -> .zshrc (interactive)
        if prefer_login:
            # For WSL/Linux, .zshenv is the most reliable
            zshenv = home / ".zshenv"
            if zshenv.exists():
                return str(zshenv)
            return str(home / ".zshrc")
        return str(home / ".zshrc")

    elif shell == "bash":
        # bash loads: .bash_profile or .profile (login) -> .bashrc (interactive)
        # WSL uses login shell, so .profile is more reliable
        if prefer_login:
            profile = home / ".profile"
            if profile.exists():
                return str(profile)
            bash_profile = home / ".bash_profile"
            if bash_profile.exists():
                return str(bash_profile)
        return str(home / ".bashrc")

    elif shell == "fish":
        return str(home / ".config" / "fish" / "config.fish")

    # Default to .profile for unknown shells
    return str(home / ".profile")


def check_local_bin_in_path() -> bool:
    """Check if ~/.local/bin is in the current PATH.

    Returns:
        True if ~/.local/bin is in PATH
    """
    path = os.environ.get("PATH", "")
    local_bin = str(Path.home() / ".local" / "bin")

    # Check both with and without trailing slash
    path_parts = path.split(os.pathsep)
    for part in path_parts:
        if part.rstrip("/\\") == local_bin.rstrip("/\\"):
            return True
    return False


def check_local_bin_in_login_shell() -> bool:
    """Check if ~/.local/bin is in PATH when using login shell.

    This runs a login shell to get the actual PATH.

    Returns:
        True if ~/.local/bin is in login shell PATH
    """
    shell = get_user_shell()

    try:
        if shell == "bash":
            result = subprocess.run(
                ["bash", "-l", "-c", "echo $PATH"],
                capture_output=True,
                text=True,
                timeout=5,
            )
        elif shell == "zsh":
            result = subprocess.run(
                ["zsh", "-l", "-c", "echo $PATH"],
                capture_output=True,
                text=True,
                timeout=5,
            )
        else:
            return False

        if result.returncode != 0:
            return False

        path = result.stdout.strip()
        local_bin = str(Path.home() / ".local" / "bin")

        path_parts = path.split(":")
        for part in path_parts:
            if part.rstrip("/") == local_bin.rstrip("/"):
                return True
        return False

    except (subprocess.TimeoutExpired, FileNotFoundError, Exception):
        return False


def check_macos_system_path() -> bool:
    """Check if ~/.local/bin is in macOS system PATH via /etc/paths or /etc/paths.d/.

    On macOS, GUI apps (VS Code, Cursor) inherit PATH from launchd which reads
    /etc/paths and /etc/paths.d/*. Shell config files (.zshrc, .zshenv) are NOT
    read by GUI apps. Claude Code's Bun process uses this system PATH, so if
    ~/.local/bin is missing, it shows a false positive PATH warning.

    Returns:
        True if ~/.local/bin is in system PATH files, or if not macOS
    """
    if platform.system() != "Darwin":
        return True

    local_bin = str(Path.home() / ".local" / "bin")

    # Check /etc/paths
    etc_paths = Path("/etc/paths")
    if etc_paths.exists():
        try:
            for line in etc_paths.read_text().splitlines():
                if line.strip() == local_bin:
                    return True
        except (PermissionError, OSError):
            pass

    # Check /etc/paths.d/*
    paths_d = Path("/etc/paths.d")
    if paths_d.exists():
        try:
            for f in paths_d.iterdir():
                if f.is_file():
                    try:
                        for line in f.read_text().splitlines():
                            if line.strip() == local_bin:
                                return True
                    except (PermissionError, OSError):
                        pass
        except (PermissionError, OSError):
            pass

    return False


def fix_macos_system_path() -> tuple[bool, str]:
    """Create /etc/paths.d/local-bin on macOS for GUI app PATH access.

    GUI apps (VS Code, Cursor) inherit PATH from launchd which reads
    /etc/paths.d/. Without this, Claude Code's process PATH is missing
    ~/.local/bin, causing a false positive diagnostic warning.

    Returns:
        Tuple of (success, message)
    """
    if platform.system() != "Darwin":
        return (False, "Not macOS")

    if check_macos_system_path():
        return (False, "/etc/paths.d/ already includes ~/.local/bin")

    local_bin = str(Path.home() / ".local" / "bin")

    try:
        result = subprocess.run(
            ["sudo", "tee", "/etc/paths.d/local-bin"],
            input=local_bin + "\n",
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode == 0:
            return (True, "Created /etc/paths.d/local-bin. Restart VS Code/terminal to apply.")
        return (False, f"Failed: {result.stderr.strip()}")
    except subprocess.TimeoutExpired:
        return (False, "Timeout (sudo may need password)")
    except Exception as e:
        return (False, str(e))


def diagnose_path() -> PathDiagnostics:
    """Diagnose PATH configuration issues.

    This is the main diagnostic function that checks:
    1. Whether ~/.local/bin exists
    2. Whether it's in the current PATH
    3. Whether it's in the login shell PATH
    4. Recommends the appropriate fix

    Returns:
        PathDiagnostics with detailed information
    """
    home = Path.home()
    local_bin = home / ".local" / "bin"

    shell = get_user_shell()
    config_file = get_shell_config_file(shell, prefer_login=is_wsl())

    # Check current state
    local_bin_exists = local_bin.exists()
    in_current_path = check_local_bin_in_path()

    # For WSL/Linux, also check login shell
    is_login = is_wsl() or platform.system() == "Linux"
    in_login_shell = check_local_bin_in_login_shell() if is_login else in_current_path

    # macOS: check system PATH (/etc/paths.d/) for GUI app compatibility
    in_sys_path = check_macos_system_path()

    # Determine recommended fix
    recommended_fix = ""
    auto_fixable = False

    if not in_current_path:
        if is_wsl():
            # WSL uses login shell, so need to configure .profile or .zshenv
            if shell == "bash":
                # .bashrc might have it but .profile doesn't source it
                recommended_fix = "echo 'export PATH=\"$HOME/.local/bin:$PATH\"' >> ~/.profile && source ~/.profile"
                config_file = str(home / ".profile")
            elif shell == "zsh":
                recommended_fix = "echo 'export PATH=\"$HOME/.local/bin:$PATH\"' >> ~/.zshenv && source ~/.zshenv"
                config_file = str(home / ".zshenv")
            else:
                recommended_fix = "echo 'export PATH=\"$HOME/.local/bin:$PATH\"' >> ~/.profile && source ~/.profile"
            auto_fixable = True
        else:
            # Regular Linux/macOS
            recommended_fix = f"echo 'export PATH=\"$HOME/.local/bin:$PATH\"' >> {config_file} && source {config_file}"
            auto_fixable = True
    elif not in_sys_path and platform.system() == "Darwin":
        # Shell PATH is fine but macOS system PATH is missing
        local_bin_str = str(local_bin)
        recommended_fix = f"sudo sh -c 'echo \"{local_bin_str}\" > /etc/paths.d/local-bin'"
        auto_fixable = True

    return PathDiagnostics(
        local_bin_in_path=in_current_path,
        local_bin_exists=local_bin_exists,
        shell_type=shell,
        config_file=config_file,
        is_login_shell_env=is_login,
        in_login_shell_path=in_login_shell,
        current_path=os.environ.get("PATH", ""),
        recommended_fix=recommended_fix,
        auto_fixable=auto_fixable,
        in_system_path=in_sys_path,
    )


def auto_fix_path(config_file: str | None = None) -> tuple[bool, str]:
    """Automatically add ~/.local/bin to PATH in shell config.

    Args:
        config_file: Optional override for config file path

    Returns:
        Tuple of (success, message)
    """
    shell = get_user_shell()
    if config_file is None:
        config_file = get_shell_config_file(shell, prefer_login=is_wsl())

    config_path = Path(config_file)
    export_line = 'export PATH="$HOME/.local/bin:$PATH"'

    try:
        # Check if already configured
        if config_path.exists():
            content = config_path.read_text()
            if ".local/bin" in content:
                return (False, f"PATH already configured in {config_file}")

        # Ensure parent directory exists
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Append export line
        with open(config_path, "a") as f:
            f.write(f"\n# Added by MoAI-ADK\n{export_line}\n")

        return (True, f"Added PATH to {config_file}. Run 'source {config_file}' or restart terminal.")

    except PermissionError:
        return (False, f"Permission denied: Cannot write to {config_file}")
    except Exception as e:
        return (False, f"Failed to update {config_file}: {e}")


def get_path_diagnostic_report() -> str:
    """Get a formatted diagnostic report for PATH issues.

    Returns:
        Formatted string report
    """
    diag = diagnose_path()

    lines = [
        "=== PATH Diagnostics ===",
        f"Platform: {platform.system()} {'(WSL)' if is_wsl() else ''}",
        f"Shell: {diag.shell_type}",
        f"Config file: {diag.config_file}",
        f"~/.local/bin exists: {'Yes' if diag.local_bin_exists else 'No'}",
        f"~/.local/bin in PATH: {'Yes' if diag.local_bin_in_path else 'No'}",
    ]

    # macOS: show system PATH status
    if platform.system() == "Darwin":
        lines.append(f"~/.local/bin in /etc/paths.d/: {'Yes' if diag.in_system_path else 'No'}")

    if not diag.local_bin_in_path or (not diag.in_system_path and platform.system() == "Darwin"):
        lines.append("")
        lines.append("=== Problem Detected ===")

        if not diag.local_bin_in_path:
            lines.append("~/.local/bin is NOT in your PATH.")
            lines.append("This can cause MCP servers and CLI tools to fail.")

        if not diag.in_system_path and platform.system() == "Darwin":
            lines.append("~/.local/bin is NOT in /etc/paths.d/ (macOS system PATH).")
            lines.append("GUI apps (VS Code, Cursor) will not see ~/.local/bin.")

        if diag.recommended_fix:
            lines.append("")
            lines.append("=== Recommended Fix ===")
            lines.append(diag.recommended_fix)

        if is_wsl() and diag.shell_type == "bash":
            lines.extend(
                [
                    "",
                    "=== WSL Explanation ===",
                    "WSL uses login shell which reads ~/.profile, not ~/.bashrc.",
                    "Even if ~/.bashrc has PATH configured, it may not be loaded.",
                    "The fix adds PATH to ~/.profile which is always sourced in WSL.",
                ]
            )

    return "\n".join(lines)
