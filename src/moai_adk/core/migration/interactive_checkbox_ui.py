"""Interactive Checkbox UI for MoAI-ADK Custom Element Restoration

This module provides an enhanced interactive CLI interface with checkbox-style selection,
arrow key navigation, and proper element preservation during updates.

Key Features:
- Arrow key navigation (↑↓) to move between elements
- Spacebar to toggle selection with [x] checkbox markers
- Category grouping (Agents, Commands, Skills, Hooks)
- Preserve unselected elements (fixes disappearing issue)
- Real-time selection status display
- Keyboard shortcuts for convenience
"""

import curses
from pathlib import Path
from typing import Any, Dict, List, Optional, Set

from .custom_element_scanner import create_custom_element_scanner


class InteractiveCheckboxUI:
    """Interactive checkbox-based CLI interface for element selection.

    Provides an enhanced user experience with:
    - Arrow key navigation
    - Checkbox-style selection with [x] markers
    - Category grouping and organization
    - Real-time feedback and status updates
    - Keyboard shortcuts for power users
    """

    def __init__(self, project_path: str | Path):
        """Initialize the interactive checkbox UI.

        Args:
            project_path: Path to the MoAI-ADK project directory
        """
        self.project_path = Path(project_path).resolve()
        self.scanner = create_custom_element_scanner(self.project_path)
        self.selected_indices: Set[int] = set()
        self.current_index = 0

    def prompt_user_selection(self, backup_available: bool = True) -> Optional[List[str]]:
        """Launch interactive checkbox selection interface.

        Args:
            backup_available: Whether backup is available for restoration

        Returns:
            List of selected element paths, or None if cancelled
        """
        # Get custom elements organized by category
        elements_by_category = self._get_elements_by_category()
        flattened_elements = self._flatten_elements(elements_by_category)

        if not flattened_elements:
            print("\n✅ No custom elements found in project.")
            print("   All elements are part of the official MoAI-ADK template.")
            return None

        if not backup_available:
            print("\n⚠️ No backup available. Cannot restore custom elements.")
            print("💡 Run 'moai-adk update' without --force to create a backup first.")
            return None

        # Launch curses interface
        try:
            selected_indices = self._run_curses_interface(flattened_elements, elements_by_category)
        except ImportError:
            # Curses not available, use enhanced fallback immediately
            print("\n📱 Terminal doesn't support interactive mode, using enhanced selection...")
            return self._fallback_selection(flattened_elements)
        except Exception as e:
            # Fallback to simple selection if curses fails
            print(f"\n⚠️ Interactive mode failed: {e}")
            print("Falling back to enhanced selection mode...")
            return self._fallback_selection(flattened_elements)

        if selected_indices is None:
            print("\n⚠️ Selection cancelled.")
            return None

        # Convert indices to element paths
        selected_paths = []
        for idx in selected_indices:
            if 0 <= idx < len(flattened_elements):
                selected_paths.append(flattened_elements[idx]["path"])

        if not selected_paths:
            print("\n⚠️ No elements selected for restoration.")
            return None

        return selected_paths

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
                skill_name = skill.name
                # Add indicator for template vs custom skills
                if hasattr(skill, "is_template") and skill.is_template:
                    skill_name = f"{skill.name} (template)"
                else:
                    skill_name = f"{skill.name} (custom)"

                organized["Skills"].append({"name": skill_name, "path": str(skill.path), "type": "skill"})

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

        # Remove empty categories
        return {k: v for k, v in organized.items() if v}

    def _flatten_elements(self, elements_by_category: Dict[str, List[Dict]]) -> List[Dict]:
        """Flatten categorized elements into a single list with display info.

        Args:
            elements_by_category: Dictionary of categorized elements

        Returns:
            Flattened list with display information
        """
        flattened = []
        for category, elements in elements_by_category.items():
            # Add category header
            flattened.append(
                {
                    "type": "header",
                    "text": f"{category} ({len(elements)})",
                    "category": category,
                }
            )
            # Add elements
            for element in elements:
                flattened.append(
                    {
                        "type": "element",
                        "name": element["name"],
                        "path": element["path"],
                        "category": element["type"],
                    }
                )

        return flattened

    def _run_curses_interface(self, elements: List[Dict], elements_by_category: Dict) -> Optional[Set[int]]:
        """Run the curses-based interactive interface.

        Args:
            elements: Flattened list of elements and headers
            elements_by_category: Categorized elements for reference

        Returns:
            Set of selected indices, or None if cancelled
        """

        def interface(stdscr):
            # Initialize curses
            curses.curs_set(0)  # Hide cursor
            curses.init_pair(1, curses.COLOR_BLACK, curses.COLOR_WHITE)  # Selected
            curses.init_pair(2, curses.COLOR_WHITE, curses.COLOR_BLUE)  # Header
            curses.init_pair(3, curses.COLOR_GREEN, curses.COLOR_BLACK)  # Success
            curses.init_pair(4, curses.COLOR_YELLOW, curses.COLOR_BLACK)  # Warning

            stdscr.clear()
            stdscr.refresh()

            # Main input loop
            while True:
                self._display_interface(stdscr, elements, elements_by_category)
                key = stdscr.getch()

                if key == curses.KEY_UP:
                    self._navigate_up(elements)
                elif key == curses.KEY_DOWN:
                    self._navigate_down(elements)
                elif key == ord(" "):
                    self._toggle_selection(elements)
                elif key == ord("a") or key == ord("A"):
                    self._select_all(elements)
                elif key == ord("n") or key == ord("N"):
                    self._select_none(elements)
                elif key == ord("\n") or key == ord("\r"):
                    # Confirm selection
                    if self._confirm_selection(stdscr, elements):
                        return self.selected_indices
                elif key == ord("q") or key == ord("Q") or key == 27:  # 27 = ESC
                    # Quit
                    return None

        # Run curses interface
        return curses.wrapper(interface)

    def _display_interface(self, stdscr, elements: List[Dict], elements_by_category: Dict) -> None:
        """Display the interactive interface.

        Args:
            stdscr: Curses window
            elements: Flattened list of elements and headers
            elements_by_category: Categorized elements for reference
        """
        stdscr.clear()
        h, w = stdscr.getmaxyx()

        # Title
        title = "🔍 Custom Elements Restoration"
        stdscr.addstr(1, (w - len(title)) // 2, title, curses.A_BOLD)

        # Instructions
        instructions = ["↑↓ Navigate | Space Toggle | A:All N:None | Enter:Confirm | Q/ESC:Cancel"]
        for i, instruction in enumerate(instructions):
            stdscr.addstr(3, 2, instruction)

        # Separator
        stdscr.addstr(4, 0, "─" * w)

        # Display elements with checkboxes
        y_offset = 6
        for i, element in enumerate(elements):
            if y_offset >= h - 3:  # Leave space for status bar
                break

            if element["type"] == "header":
                # Category header
                header_text = f"📁 {element['text']}"
                stdscr.addstr(y_offset, 2, header_text, curses.color_pair(2) | curses.A_BOLD)
                y_offset += 1
            else:
                # Element with checkbox
                is_selected = i in self.selected_indices
                is_current = i == self.current_index

                checkbox = "[x]" if is_selected else "[ ]"
                display_name = element["name"][:40]  # Truncate long names

                line = f"{checkbox} {display_name}"

                if is_current:
                    stdscr.addstr(y_offset, 4, line, curses.color_pair(1) | curses.A_BOLD)
                else:
                    stdscr.addstr(y_offset, 4, line)

                y_offset += 1

        # Status bar
        selected_count = len(
            [i for i, el in enumerate(elements) if el["type"] == "element" and i in self.selected_indices]
        )
        total_count = len([el for el in elements if el["type"] == "element"])
        status = f"Selected: {selected_count}/{total_count} | Use Space to toggle, Enter to confirm"
        stdscr.addstr(h - 2, 2, status, curses.A_REVERSE)

        stdscr.refresh()

    def _navigate_up(self, elements: List[Dict]) -> None:
        """Move cursor up to previous element.

        Args:
            elements: List of elements to navigate
        """
        # Find previous element index
        new_index = self.current_index - 1
        while new_index >= 0:
            if elements[new_index]["type"] == "element":
                self.current_index = new_index
                break
            new_index -= 1

    def _navigate_down(self, elements: List[Dict]) -> None:
        """Move cursor down to next element.

        Args:
            elements: List of elements to navigate
        """
        # Find next element index
        new_index = self.current_index + 1
        while new_index < len(elements):
            if elements[new_index]["type"] == "element":
                self.current_index = new_index
                break
            new_index += 1

    def _toggle_selection(self, elements: List[Dict]) -> None:
        """Toggle selection for current element.

        Args:
            elements: List of elements
        """
        if elements[self.current_index]["type"] == "element":
            if self.current_index in self.selected_indices:
                self.selected_indices.remove(self.current_index)
            else:
                self.selected_indices.add(self.current_index)

    def _select_all(self, elements: List[Dict]) -> None:
        """Select all elements.

        Args:
            elements: List of elements
        """
        for i, element in enumerate(elements):
            if element["type"] == "element":
                self.selected_indices.add(i)

    def _select_none(self, elements: List[Dict]) -> None:
        """Clear selection for all elements.

        Args:
            elements: List of elements
        """
        self.selected_indices.clear()

    def _confirm_selection(self, stdscr, elements: List[Dict]) -> bool:
        """Show confirmation dialog.

        Args:
            stdscr: Curses window
            elements: List of elements

        Returns:
            True if confirmed, False otherwise
        """
        selected_paths = []
        for idx in self.selected_indices:
            if 0 <= idx < len(elements) and elements[idx]["type"] == "element":
                selected_paths.append(elements[idx]["name"])

        if not selected_paths:
            return False

        h, w = stdscr.getmaxyx()

        # Clear screen for confirmation
        stdscr.clear()
        stdscr.addstr(1, (w - 20) // 2, "Confirm Selection", curses.A_BOLD)
        stdscr.addstr(3, 2, f"Selected {len(selected_paths)} elements:")

        for i, path in enumerate(selected_paths[:10]):  # Show max 10 items
            if i >= h - 8:
                break
            stdscr.addstr(5 + i, 4, f"• {path}")

        if len(selected_paths) > 10:
            stdscr.addstr(h - 4, 4, f"... and {len(selected_paths) - 10} more")

        stdscr.addstr(h - 3, 2, "Confirm restoration? (Y/n): ")
        stdscr.refresh()

        key = stdscr.getch()
        return key in [ord("y"), ord("Y"), ord("\n"), ord("\r")]

    def _fallback_selection(self, flattened_elements: List[Dict]) -> Optional[List[str]]:
        """Fallback selection method when curses is not available.

        Args:
            flattened_elements: List of elements to select from

        Returns:
            List of selected paths or None if cancelled
        """
        print("\n" + "=" * 60)
        print("🔍 Custom Elements Detected (Simple Mode)")
        print("=" * 60)
        print("These elements are not part of the official MoAI-ADK template:")
        print()

        element_list = [el for el in flattened_elements if el["type"] == "element"]
        for i, element in enumerate(element_list, 1):
            print(f"  {i:2d}. [{element['category']}] {element['name']}")

        print("\n💡 Selection Instructions:")
        print("   • Enter numbers separated by commas (e.g., 1,3,5)")
        print("   • Use 'all' to select all elements")
        print("   • Press Enter with empty input to cancel")

        try:
            user_input = input("\nSelect elements to restore: ").strip()
        except (KeyboardInterrupt, EOFError):
            print("\n⚠️ Selection cancelled.")
            return None

        if not user_input:
            return None

        if user_input.lower() == "all":
            return [el["path"] for el in element_list]

        # Parse numbers
        selected_paths = []
        try:
            indices = [int(x.strip()) for x in user_input.split(",")]
            for idx in indices:
                if 1 <= idx <= len(element_list):
                    selected_paths.append(element_list[idx - 1]["path"])
                else:
                    print(f"⚠️ Invalid number: {idx}")
        except ValueError:
            print("⚠️ Invalid input. Please enter numbers separated by commas.")
            return None

        return selected_paths if selected_paths else None

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
            return confirm in ["y", "yes"]
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


def create_interactive_checkbox_ui(project_path: str | Path) -> InteractiveCheckboxUI:
    """Factory function to create an InteractiveCheckboxUI.

    Args:
        project_path: Path to the MoAI-ADK project directory

    Returns:
        Configured InteractiveCheckboxUI instance
    """
    return InteractiveCheckboxUI(Path(project_path).resolve())
