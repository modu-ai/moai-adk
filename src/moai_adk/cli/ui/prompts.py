"""MoAI-ADK CLI Prompt Components

Modern interactive prompts using InquirerPy with fuzzy search support.
"""

from typing import Any, Callable, Dict, List, Optional, Sequence, Union

from InquirerPy import inquirer
from InquirerPy.base.control import Choice, Separator
from InquirerPy.utils import get_style
from InquirerPy.validator import EmptyInputValidator

from .theme import MOAI_THEME, SYMBOLS

# Pre-create InquirerPy compatible style
_INQUIRER_STYLE = get_style(MOAI_THEME, style_override=False)


def fuzzy_checkbox(
    message: str,
    choices: Sequence[Union[str, Dict[str, Any], Choice, Separator]],
    default: Optional[List[Any]] = None,
    instruction: str = "[Space] Toggle  [Tab] Toggle All  [Enter] Confirm  [/] Search",
    multiselect: bool = True,
    marker: str = SYMBOLS["checkbox_selected"],  # type: ignore[assignment]
    marker_pl: str = " ",
    border: bool = True,
    height: Optional[int] = None,
    max_height: int = 15,
    validate: Optional[Callable[[List[Any]], bool]] = None,
    invalid_message: str = "Please select at least one item",
    keybindings: Optional[Dict[str, List[Dict[str, str]]]] = None,
) -> Optional[List[Any]]:
    """Display a fuzzy searchable checkbox prompt.

    Allows users to search through choices while selecting multiple items.
    Perfect for selecting files from a large list.

    Args:
        message: Question to display
        choices: List of choices (strings, dicts, or Choice objects)
        default: Pre-selected values
        instruction: Help text shown below prompt
        multiselect: Allow multiple selections (always True for checkbox)
        marker: Symbol for selected items
        marker_pl: Placeholder for unselected items
        border: Show border around choices
        height: Fixed height (None for auto)
        max_height: Maximum height when auto
        validate: Custom validation function
        invalid_message: Message when validation fails
        keybindings: Custom key bindings

    Returns:
        List of selected values, or None if cancelled

    Example:
        >>> files = fuzzy_checkbox(
        ...     "Select files to restore:",
        ...     choices=[
        ...         {"name": "commands/auth.md", "value": "commands/auth.md"},
        ...         {"name": "agents/backend.md", "value": "agents/backend.md"},
        ...     ]
        ... )
    """
    # Convert simple strings to Choice objects
    processed_choices = _process_choices(list(choices), default)

    # Set up default keybindings if not provided
    if keybindings is None:
        keybindings = {
            "toggle": [{"key": "space"}],
            "toggle-all": [{"key": "tab"}],
            "toggle-all-true": [{"key": "a"}],
            "toggle-all-false": [{"key": "n"}],
        }

    try:
        result = inquirer.fuzzy(
            message=message,
            choices=processed_choices,
            multiselect=multiselect,
            marker=marker,
            marker_pl=marker_pl,
            border=border,
            info=True,
            match_exact=False,
            instruction=instruction,
            long_instruction="",
            max_height=max_height if height is None else height,
            validate=validate,
            invalid_message=invalid_message,
            keybindings=keybindings,  # type: ignore[arg-type]
            style=_INQUIRER_STYLE,
            qmark=SYMBOLS["pointer"],
            amark=SYMBOLS["success"],
            transformer=lambda x: f"{len(x)} selected" if isinstance(x, list) else x,
        ).execute()

        return result

    except KeyboardInterrupt:
        return None


def fuzzy_select(
    message: str,
    choices: Sequence[Union[str, Dict[str, Any], Choice, Separator]],
    default: Optional[Any] = None,
    instruction: str = "[↑↓] Navigate  [Enter] Select  [/] Search",
    border: bool = True,
    max_height: int = 10,
) -> Optional[Any]:
    """Display a fuzzy searchable single-select prompt.

    Args:
        message: Question to display
        choices: List of choices
        default: Pre-selected value
        instruction: Help text
        border: Show border around choices
        max_height: Maximum display height

    Returns:
        Selected value, or None if cancelled
    """
    processed_choices = _process_choices(list(choices), [default] if default else None)

    try:
        result = inquirer.fuzzy(
            message=message,
            choices=processed_choices,
            multiselect=False,
            border=border,
            info=True,
            match_exact=False,
            instruction=instruction,
            max_height=max_height,
            style=_INQUIRER_STYLE,
            qmark=SYMBOLS["pointer"],
            amark=SYMBOLS["success"],
        ).execute()

        return result

    except KeyboardInterrupt:
        return None


def styled_checkbox(
    message: str,
    choices: Sequence[Union[str, Dict[str, Any], Choice, Separator]],
    default: Optional[List[Any]] = None,
    instruction: str = "[Space] Toggle  [a] All  [n] None  [Enter] Confirm",
    cycle: bool = True,
) -> Optional[List[Any]]:
    """Display a styled checkbox prompt without fuzzy search.

    Use this for shorter lists where fuzzy search isn't needed.

    Args:
        message: Question to display
        choices: List of choices
        default: Pre-selected values
        instruction: Help text
        cycle: Whether to cycle through choices

    Returns:
        List of selected values, or None if cancelled
    """
    processed_choices = _process_choices(list(choices), default)

    try:
        result = inquirer.checkbox(
            message=message,
            choices=processed_choices,
            cycle=cycle,
            instruction=instruction,
            style=_INQUIRER_STYLE,
            qmark=SYMBOLS["pointer"],
            amark=SYMBOLS["success"],
            enabled_symbol=SYMBOLS["checkbox_selected"],
            disabled_symbol=SYMBOLS["checkbox_unselected"],
            transformer=lambda x: f"{len(x)} selected" if isinstance(x, list) else x,
        ).execute()

        return result

    except KeyboardInterrupt:
        return None


def styled_select(
    message: str,
    choices: Sequence[Union[str, Dict[str, Any], Choice, Separator]],
    default: Optional[Any] = None,
    instruction: str = "[↑↓] Navigate  [Enter] Select",
    cycle: bool = True,
) -> Optional[Any]:
    """Display a styled single-select prompt.

    Args:
        message: Question to display
        choices: List of choices
        default: Pre-selected value
        instruction: Help text
        cycle: Whether to cycle through choices

    Returns:
        Selected value, or None if cancelled
    """
    processed_choices = _process_choices(list(choices), [default] if default else None)

    try:
        result = inquirer.select(
            message=message,
            choices=processed_choices,
            default=default,
            cycle=cycle,
            instruction=instruction,
            style=_INQUIRER_STYLE,
            qmark=SYMBOLS["pointer"],
            amark=SYMBOLS["success"],
            pointer=SYMBOLS["pointer"],
        ).execute()

        return result

    except KeyboardInterrupt:
        return None


def styled_input(
    message: str,
    default: str = "",
    validate: Optional[Callable[[str], bool]] = None,
    invalid_message: str = "Invalid input",
    required: bool = True,
    instruction: str = "",
) -> Optional[str]:
    """Display a styled text input prompt.

    Args:
        message: Question to display
        default: Default value
        validate: Custom validation function
        invalid_message: Message when validation fails
        required: Whether input is required
        instruction: Help text

    Returns:
        User input, or None if cancelled
    """
    validators = []
    if required:
        validators.append(EmptyInputValidator("This field is required"))

    try:
        result = inquirer.text(
            message=message,
            default=default,
            validate=validate,
            invalid_message=invalid_message,
            instruction=instruction,
            style=_INQUIRER_STYLE,
            qmark=SYMBOLS["pointer"],
            amark=SYMBOLS["success"],
        ).execute()

        return result

    except KeyboardInterrupt:
        return None


def styled_confirm(
    message: str,
    default: bool = True,
    instruction: str = "[y/n]",
) -> Optional[bool]:
    """Display a styled confirmation prompt.

    Args:
        message: Question to display
        default: Default value (True = yes, False = no)
        instruction: Help text

    Returns:
        True/False, or None if cancelled
    """
    try:
        result = inquirer.confirm(
            message=message,
            default=default,
            instruction=instruction,
            style=_INQUIRER_STYLE,
            qmark=SYMBOLS["pointer"],
            amark=SYMBOLS["success"],
        ).execute()

        return result

    except KeyboardInterrupt:
        return None


def _process_choices(
    choices: List[Union[str, Dict[str, Any], Choice, Separator]],
    defaults: Optional[List[Any]] = None,
) -> List[Union[Choice, Separator]]:
    """Process and normalize choices for InquirerPy prompts.

    Args:
        choices: Raw choices list
        defaults: Values that should be pre-selected

    Returns:
        List of Choice/Separator objects
    """
    defaults = defaults or []
    processed: List[Union[Choice, Separator]] = []

    for choice in choices:
        if isinstance(choice, (Choice, Separator)):
            # Already a Choice or Separator object
            processed.append(choice)

        elif isinstance(choice, dict):
            # Dictionary with name/value
            name = choice.get("name", choice.get("title", ""))
            value = choice.get("value", name)
            enabled = choice.get("enabled", value in defaults)
            disabled = choice.get("disabled", False)

            if disabled:
                # Create a separator for disabled items
                processed.append(Separator(line=name))
            else:
                processed.append(Choice(value=value, name=name, enabled=enabled))

        elif isinstance(choice, str):
            # Simple string
            enabled = choice in defaults
            processed.append(Choice(value=choice, name=choice, enabled=enabled))

    return processed


def create_grouped_choices(
    groups: Dict[str, List[Dict[str, Any]]],
    defaults: Optional[List[Any]] = None,
) -> List[Union[Choice, Separator]]:
    """Create grouped choices with separators for each category.

    Args:
        groups: Dictionary mapping category names to lists of choices
        defaults: Values that should be pre-selected

    Returns:
        List of choices with category separators

    Example:
        >>> choices = create_grouped_choices({
        ...     "Commands": [
        ...         {"name": "auth.md", "value": "commands/auth.md"},
        ...         {"name": "deploy.md", "value": "commands/deploy.md"},
        ...     ],
        ...     "Agents": [
        ...         {"name": "backend.md", "value": "agents/backend.md"},
        ...     ],
        ... })
    """
    defaults = defaults or []
    result: List[Union[Choice, Separator]] = []

    for category, items in groups.items():
        if items:  # Only add category if it has items
            # Add category separator
            result.append(Separator(line=f"── {category} ──"))

            # Add items in this category
            for item in items:
                name = item.get("name", item.get("title", ""))
                value = item.get("value", name)
                enabled = item.get("enabled", value in defaults)

                result.append(Choice(value=value, name=f"  {name}", enabled=enabled))

    return result
