#!/usr/bin/env python3
"""
🗿 MoAI-ADK Banner Module

Claude Code 스타일의 3D 블록 효과와 Box Drawing 유니코드를 활용한
현대적인 ASCII 아트 배너 모듈
"""

import os
import sys
from typing import List, Tuple

# Version management integration
try:
    from .._version import __version__, get_version
except ImportError:
    try:
        from moai_adk._version import __version__, get_version
    except ImportError:
        __version__ = "0.1.16"
        def get_version(component="moai_adk"): return "0.1.16"


# Color constants for gradient effect
class Colors:
    """ANSI color codes for terminal output."""

    # Reset
    RESET = "\033[0m"

    # Claude AI Official Brand Color - #da7756 (RGB: 218, 119, 86)
    CLAUDE_BRAND = "\033[38;2;218;119;86m"  # True color using RGB values

    # Legacy gradient colors (commented out)
    # COPPER_DARK = "\033[38;5;130m"    # Dark Copper/Brown (외곽 그림자)
    # COPPER = "\033[38;5;166m"         # Copper (주 외곽선)
    # ORANGE_DARK = "\033[38;5;202m"    # Dark Orange (진한 블록)
    # ORANGE = "\033[38;5;208m"         # Orange (중간 블록)
    # SALMON = "\033[38;5;209m"         # Salmon (밝은 블록)
    # PEACH = "\033[38;5;216m"          # Peach (하이라이트)
    # CREAM = "\033[38;5;223m"          # Cream (가장 밝은 부분)

    # Text colors
    SUBTITLE = "\033[38;5;245m"  # Gray for subtitle
    FOOTER = "\033[38;5;240m"  # Darker gray for footer


def supports_color() -> bool:
    """Check if terminal supports color output."""
    return (
        hasattr(sys.stdout, "isatty")
        and sys.stdout.isatty()
        and os.environ.get("TERM") != "dumb"
        and os.environ.get("NO_COLOR") is None
    )


def apply_claude_brand_color(line: str) -> str:
    """Apply Claude AI brand color to ASCII art line."""
    if not supports_color() or not line.strip():
        return line

    colored = []

    for char in line:
        if char.strip():  # Any non-whitespace character - use Claude brand color
            colored.append(f"{Colors.CLAUDE_BRAND}{char}")
        else:  # Whitespace
            colored.append(char)

    return "".join(colored) + Colors.RESET


def get_moai_logo() -> List[str]:
    """Modern MoAI logo using Box Drawing characters with 3D block effect."""
    return [
        "███╗   ███╗          █████╗ ██╗",
        "████╗ ████║ ██████╗ ██╔══██╗██║",
        "██╔████╔██║██╔═══██ ███████║██║",
        "██║╚██╔╝██║██║   ██║██╔══██║██║",
        "██║ ╚═╝ ██║╚██████╔╝██║  ██║██║",
        "╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝",
    ]


def create_banner(version: str = None, show_usage: bool = False) -> str:
    """Create the complete MoAI banner with two-tone color scheme."""
    if version is None:
        version = get_version("moai_adk")
    
    moai_lines = get_moai_logo()

    banner_lines = []

    # Bottom border only - matching MoAI logo width (33 chars)
    if supports_color():
        border = f"{Colors.CLAUDE_BRAND}{'═' * 33}{Colors.RESET}"
    else:
        border = "═" * 33

    banner_lines.append("")  # Single empty line at top

    # MOAI logo with Claude brand color - LEFT ALIGNED
    for line in moai_lines:
        colored_line = apply_claude_brand_color(line)
        banner_lines.append(colored_line)  # No padding - left aligned

    # Reduced spacing - no empty line before border
    banner_lines.append(border)
    banner_lines.append("")

    # Add description line with Claude brand color
    description = "MoAI-ADK: Agentic Development Toolkit for Claude Code 🚀"
    if supports_color():
        banner_lines.append(f"{Colors.CLAUDE_BRAND}{description}{Colors.RESET}")
    else:
        banner_lines.append(description)

    # Add usage info if requested
    if show_usage:
        banner_lines.append("")
        banner_lines.append("Usage: moai <command> [options]")
        banner_lines.append("")
        banner_lines.append("Commands:")
        banner_lines.append(
            "  init <project-name>            Initialize a new Claude Code project"
        )
        banner_lines.append(
            "  update                         Update MoAI-ADK to the latest version"
        )
        banner_lines.append(
            "  doctor                         Diagnose and fix common MoAI-ADK issues"
        )
        banner_lines.append("  help [command]                 Display help for command")
        banner_lines.append("")
        banner_lines.append("Options:")
        banner_lines.append(
            "  -V, --version                  Output the version number"
        )
        banner_lines.append("  -h, --help                     Display help for command")

    # Add footer
    banner_lines.append("")
    footer = "🗿 모두의AI (https://mo.ai.kr)"
    if supports_color():
        banner_lines.append(f"{Colors.FOOTER}{footer}{Colors.RESET}")
    else:
        banner_lines.append(footer)
    banner_lines.append("")

    return "\n".join(banner_lines)


def print_banner(version: str = None) -> None:
    """Print the MoAI-ADK banner to stdout."""
    print(create_banner(version))


if __name__ == "__main__":
    # Test the banner
    print_banner()
