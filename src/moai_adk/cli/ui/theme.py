"""MoAI-ADK CLI Theme Configuration

Consistent styling for all CLI components using InquirerPy custom styles.
"""

from typing import Any, Dict, Optional

from rich.console import Console

_console = Console()

# Claude Code official colors + MoAI theme
MOAI_COLORS = {
    "primary": "#DA7756",  # Claude Code terra cotta - main brand color
    "secondary": "#10B981",  # Green - success/selected
    "accent": "#DA7756",  # Claude Code terra cotta - highlights
    "info": "#3B82F6",  # Blue - information
    "warning": "#F59E0B",  # Amber - warnings
    "error": "#EF4444",  # Red - errors
    "muted": "#6B7280",  # Gray - secondary text
    "text": "#E5E7EB",  # Light gray - primary text
    "claude_terra": "#DA7756",  # Claude Code official terra cotta
    "claude_black": "#000000",  # Claude Code official black
}

# InquirerPy style configuration
MOAI_THEME: Dict[str, str] = {
    # Question styling
    "questionmark": f"fg:{MOAI_COLORS['primary']} bold",
    "question": "bold",
    "answered_question": "",
    "answer": f"fg:{MOAI_COLORS['secondary']} bold",
    "input": f"fg:{MOAI_COLORS['text']}",
    "instruction": f"fg:{MOAI_COLORS['muted']}",
    # Pointer and marker styling
    "pointer": f"fg:{MOAI_COLORS['primary']} bold",
    "checkbox": f"fg:{MOAI_COLORS['primary']}",
    "selected": f"fg:{MOAI_COLORS['secondary']}",
    "separator": f"fg:{MOAI_COLORS['muted']}",
    # Fuzzy prompt styling
    "fuzzy_prompt": f"fg:{MOAI_COLORS['primary']}",
    "fuzzy_info": f"fg:{MOAI_COLORS['muted']}",
    "fuzzy_border": f"fg:{MOAI_COLORS['muted']}",
    "fuzzy_match": f"fg:{MOAI_COLORS['accent']} bold",
    # Validator styling
    "validator": "",
    "marker": f"fg:{MOAI_COLORS['secondary']}",
    "long_instruction": f"fg:{MOAI_COLORS['muted']}",
    # Skipped question styling
    "skipped": f"fg:{MOAI_COLORS['muted']}",
}

# Symbols for consistent UI (string symbols only for InquirerPy)
SYMBOLS: Dict[str, str] = {
    "checkbox_selected": "●",  # Filled circle for selected
    "checkbox_unselected": "○",  # Empty circle for unselected
    "pointer": "❯",  # Arrow pointer
    "success": "✓",
    "error": "✗",
    "warning": "⚠",
    "info": "ℹ",
}

# Spinner frames for Rich progress indicators
SPINNER_FRAMES = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"]


def get_styled_choice(
    title: str,
    value: Any,
    enabled: bool = False,
    description: Optional[str] = None,
) -> Dict[str, Any]:
    """Create a styled choice for InquirerPy prompts.

    Args:
        title: Display text for the choice
        value: Value returned when selected
        enabled: Whether pre-selected (for checkbox)
        description: Optional description shown below title

    Returns:
        Dictionary suitable for InquirerPy Choice
    """
    choice: Dict[str, Any] = {
        "name": title,
        "value": value,
        "enabled": enabled,
    }

    if description:
        # Append description in muted color
        choice["name"] = f"{title}\n  {description}"

    return choice


def print_styled(
    message: str,
    style: str = "text",
    bold: bool = False,
    newline: bool = True,
) -> None:
    """Print styled text using MoAI theme colors.

    Args:
        message: Text to print
        style: One of 'primary', 'secondary', 'accent', 'info', 'warning', 'error', 'muted', 'text'
        bold: Whether to make text bold
        newline: Whether to add newline after message
    """
    color = MOAI_COLORS.get(style, MOAI_COLORS["text"])
    rich_style = f"[{color}]"
    if bold:
        rich_style = f"[bold {color}]"

    _console.print(f"{rich_style}{message}[/]", end="\n" if newline else "")


def get_category_separator(category: str) -> Dict[str, Any]:
    """Create a styled category separator for grouping choices.

    Args:
        category: Category name to display

    Returns:
        Separator dictionary for InquirerPy
    """
    return {"value": None, "name": f"── {category} ──", "disabled": True}
