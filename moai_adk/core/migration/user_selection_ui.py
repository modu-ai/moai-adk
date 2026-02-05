"""User Selection UI for MoAI-ADK Custom Element Restoration

This module provides enhanced CLI-based user interface for selecting custom elements
to restore from backup during MoAI-ADK updates. It features:

- Interactive checkbox selection with arrow key navigation (preferred)
- Fallback to numbered selection when curses is not available
- Category grouping (Agents, Commands, Skills, Hooks)
- Proper element preservation (unselected elements remain intact)
"""

from pathlib import Path
from typing import Any, Dict, List, Optional

from .custom_element_scanner import create_custom_element_scanner


class UserSelectionUI:
    """Enhanced CLI-based user interface for selecting custom elements to restore.

    This class provides an improved CLI interface with interactive checkbox selection
    when available, falling back to numbered selection when curses is not supported.
    It properly preserves unselected elements and includes all categories.
    """

    def __init__(self, project_path: str | Path):
        """Initialize the user selection UI.

        Args:
            project_path: Path to the MoAI-ADK project directory
        """
        self.project_path = Path(project_path).resolve()
        self.scanner = create_custom_element_scanner(self.project_path)

        # Try to import interactive UI, fall back to basic UI if not available
        try:
            from .interactive_checkbox_ui import create_interactive_checkbox_ui

            self.interactive_ui = create_interactive_checkbox_ui(self.project_path)
            self.use_interactive = True
        except Exception:
            self.interactive_ui = None
            self.use_interactive = False

    def prompt_user_selection(self, backup_available: bool = True) -> Optional[List[str]]:
        """Prompt user to select custom elements for restoration.

        Args:
            backup_available: Whether backup is available for restoration

        Returns:
            List of selected element paths, or None if no selection made
            or no elements available

        Example:
            >>> ui = UserSelectionUI("/project")
            >>> selected = ui.prompt_user_selection()
            >>> print(f"Selected {len(selected)} elements")
        """
        # Try interactive UI first
        if self.use_interactive and self.interactive_ui:
            try:
                return self.interactive_ui.prompt_user_selection(backup_available)
            except Exception as e:
                print(f"⚠️ Interactive UI failed: {e}")
                print("Falling back to basic selection mode...")

        # Fallback to basic selection
        return self._basic_selection_mode(backup_available)

    def _basic_selection_mode(self, backup_available: bool) -> Optional[List[str]]:
        """Basic numbered selection mode (fallback when interactive UI not available).

        Args:
            backup_available: Whether backup is available for restoration

        Returns:
            List of selected element paths, or None if cancelled
        """
        # Get custom elements organized by category
        elements_by_category = self._get_elements_by_category()

        # No custom elements found
        if not any(elements_by_category.values()):
            print("\n✅ No custom elements found in project.")
            print("   All elements are part of the official MoAI-ADK template.")
            return None

        # Display elements by category
        print("\n" + "=" * 60)
        print("🔍 Custom Elements Detected (Basic Mode)")
        print("=" * 60)
        print("These elements are not part of the official MoAI-ADK template:")
        print()

        element_index = 1
        element_map = {}  # Map index to element path

        for category, elements in elements_by_category.items():
            if not elements:
                continue

            print(f"\n📁 {category} ({len(elements)}):")
            for element in elements:
                status = "✓ Available" if backup_available else "⚠ No backup"
                print(f"  {element_index:2d}. {element['name']:<35} {status}")
                element_map[element_index] = element["path"]
                element_index += 1

        # Show selection instructions
        print()
        if not backup_available:
            print("⚠️ No backup available. Cannot restore custom elements.")
            print("💡 Run 'moai-adk update' without --force to create a backup first.")
            return None

        print("💡 Selection Instructions:")
        print("   • Enter numbers separated by commas (e.g., 1,3,5)")
        print("   • Use 'all' to select all elements")
        print("   • Press Enter with empty input to cancel")

        # Get user input
        try:
            user_input = input("\nSelect elements to restore: ").strip()
        except (KeyboardInterrupt, EOFError):
            print("\n⚠️ Selection cancelled.")
            return None

        # Process user input
        if not user_input:
            print("No elements selected.")
            return None

        # Check for "all" shortcut
        if user_input.lower() == "all":
            all_elements = list(element_map.values())
            # Ask for confirmation when selecting all elements
            if self.confirm_selection(all_elements):
                return all_elements
            else:
                return None

        # Parse numbered selection
        selected_elements = []
        try:
            indices = [int(x.strip()) for x in user_input.split(",") if x.strip()]
            for idx in indices:
                if idx in element_map:
                    selected_elements.append(element_map[idx])
                else:
                    print(f"⚠️ Invalid number: {idx}")
        except ValueError:
            print("⚠️ Invalid input. Please enter numbers separated by commas.")
            return None

        if not selected_elements:
            print("No valid selections made.")
            return None

        # Ask for confirmation
        if self.confirm_selection(selected_elements):
            return selected_elements
        else:
            return None

    def _get_elements_by_category(self) -> Dict[str, List[Dict]]:
        """Organize custom elements by category.

        Returns:
            Dictionary mapping category names to lists of elements
        """
        custom_elements = self.scanner.scan_custom_elements()
        organized: Dict[str, List[Dict[str, Any]]] = {
            "Agents": [],
            "Commands": [],
            "Skills": [],
            "Hooks": [],
        }

        # Add skills (which are directories)
        if "skills" in custom_elements:
            for skill in custom_elements["skills"]:
                organized["Skills"].append({"name": skill.name, "path": str(skill.path), "type": "skill"})

        # Add file-based elements
        for element_type in ["agents", "commands", "hooks"]:
            if element_type in custom_elements:
                category_name = element_type.capitalize()
                for element_path in custom_elements[element_type]:
                    element_name = Path(element_path).name
                    organized[category_name].append(
                        {
                            "name": element_name,
                            "path": str(element_path),
                            "type": element_type.rstrip("s"),  # Remove plural 's'
                        }
                    )

        return organized

    def _parse_selection(self, user_input: str, custom_elements: List[Dict[str, str]]) -> List[str]:
        """Parse user selection input with enhanced support for multiple separators.

        Args:
            user_input: User's input string (can use commas, spaces, semicolons, etc.)
            custom_elements: List of available custom elements

        Returns:
            List of selected element paths
        """
        selected_paths = []

        # Enhanced parsing: handle multiple separator types (comma, space, semicolon, backslash)
        import re

        # Split by common separators and clean up
        selections = re.split(r"[,\s;\\]+", user_input.strip())
        selections = [s.strip() for s in selections if s.strip()]

        for selection in selections:
            try:
                # Parse as number
                index = int(selection)
                # Find element with this index (1-based)
                for element in custom_elements:
                    if element["index"] == index:
                        selected_paths.append(element["path"])
                        break
                else:
                    print(f"⚠️ Invalid selection: {selection} (not in available list)")
            except ValueError:
                # Try to parse as element name (for enhanced UX)
                found = False
                for element in custom_elements:
                    if (
                        selection.lower() in element["name"].lower()
                        or selection.lower() in element["display_name"].lower()
                    ):
                        selected_paths.append(element["path"])
                        found = True
                        break

                if not found:
                    print(f"⚠️ Invalid selection: {selection} (not a number or matching element name)")

        return selected_paths

    def confirm_selection(self, selected_elements: List[str]) -> bool:
        """Confirm user's selection before proceeding with restoration.

        Args:
            selected_elements: List of selected element paths

        Returns:
            True if user confirms, False otherwise
        """
        print("\n📋 Selection Summary:")
        print("-" * 40)
        for i, element_path in enumerate(selected_elements, 1):
            element_name = Path(element_path).name
            element_type = self._get_element_type(element_path)
            print(f"  {i}. {element_name} ({element_type})")

        print("-" * 40)
        print(f"Total elements selected: {len(selected_elements)}")

        try:
            confirm = input("\nConfirm restoration? (y/N): ").strip().lower()
            return confirm in ["y", "yes", ""]
        except (KeyboardInterrupt, EOFError):
            print("\n⚠️ Restoration cancelled.")
            return False

    def _get_element_type(self, element_path: str) -> str:
        """Get element type from path.

        Args:
            element_path: Path to element

        Returns:
            Element type string (agent, command, skill, hook)
        """
        path = Path(element_path)
        parts = path.parts

        if "agents" in parts:
            return "agent"
        elif "commands" in parts:
            return "command"
        elif "skills" in parts:
            return "skill"
        elif "hooks" in parts:
            return "hook"
        else:
            return "unknown"


def display_selection_instructions():
    """Display instructions for using the enhanced selection interface."""
    print(
        """
📖 Enhanced Selection Instructions:
  • Enter numbers separated by spaces, commas, semicolons (e.g., 1 3,4 or 1;3\\4)
  • Use 'all' to select all elements
  • Try partial names for convenience (e.g., 'yoda' or 'agent')
  • Press Enter with empty input to cancel
  • Use Ctrl+C to interrupt selection

📂 Element Types:
  • agent: Custom agent for specific tasks
  • command: Custom slash command for workflows
  • skill: Custom skill with enhanced capabilities
  • hook: Custom hook for system integration

💡 Pro Tips:
  • Mixed separators work: "1,3;5\\7"
  • Partial name matching: "my-agen" matches "my-agent"
  • Case insensitive: "YODA" matches "yoda"
    """
    )


def create_user_selection_ui(project_path: str | Path) -> UserSelectionUI:
    """Factory function to create a UserSelectionUI.

    Args:
        project_path: Path to the MoAI-ADK project directory

    Returns:
        Configured UserSelectionUI instance

    Example:
        >>> ui = create_user_selection_ui("/path/to/project")
        >>> selected = ui.prompt_user_selection()
        >>> if selected:
        >>>     ui.confirm_selection(selected)
    """
    return UserSelectionUI(Path(project_path).resolve())
