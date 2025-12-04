"""MoAI-ADK CLI UI Components

Modern, user-friendly CLI interface components using InquirerPy and Rich.

Provides:
- Fuzzy search checkbox prompts
- Progress indicators (bars, spinners)
- Styled text input and selection prompts
- Consistent theming across all CLI commands
"""

from .progress import (
    ProgressContext,
    SpinnerContext,
    create_progress_bar,
    create_spinner,
)
from .prompts import (
    fuzzy_checkbox,
    fuzzy_select,
    styled_checkbox,
    styled_confirm,
    styled_input,
    styled_select,
)
from .theme import MOAI_THEME, get_styled_choice

__all__ = [
    # Prompts
    "fuzzy_checkbox",
    "fuzzy_select",
    "styled_checkbox",
    "styled_confirm",
    "styled_input",
    "styled_select",
    # Progress
    "ProgressContext",
    "SpinnerContext",
    "create_progress_bar",
    "create_spinner",
    # Theme
    "MOAI_THEME",
    "get_styled_choice",
]
