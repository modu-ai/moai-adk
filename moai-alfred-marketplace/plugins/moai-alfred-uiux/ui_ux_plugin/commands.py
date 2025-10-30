"""
UI/UX Plugin Commands - /setup-shadcn-ui command implementation

@CODE:UIUX-INIT-CMD-001:COMMANDS
"""

from dataclasses import dataclass
from pathlib import Path
from typing import Optional, List
import json
from datetime import datetime


# @CODE:UIUX-COMMAND-RESULT-001:RESULT
@dataclass
class CommandResult:
    """Result object for command execution"""
    success: bool
    config_dir: Path
    files_created: List[str]
    message: str
    error: Optional[str] = None


class SetupShadcnUICommand:
    """
    /setup-shadcn-ui command implementation

    Sets up shadcn/ui component library configuration and templates
    """

    # Validation constants
    MIN_PROJECT_NAME_LENGTH = 3
    MAX_PROJECT_NAME_LENGTH = 50
    VALID_FRAMEWORKS = ["nextjs", "react", "vite"]
    AVAILABLE_COMPONENTS = [
        "button", "input", "card", "modal", "dropdown",
        "accordion", "badge", "calendar", "checkbox", "dialog",
        "form", "label", "pagination", "select", "switch",
        "table", "tabs", "textarea", "tooltip"
    ]

    def __init__(self):
        """Initialize UI/UX Plugin command"""
        pass

    def validate_project_name(self, project_name: str) -> bool:
        """
        Validate project name format

        @CODE:UIUX-VALIDATE-NAME-001:VALIDATION
        """
        if not project_name:
            raise ValueError("Project name cannot be empty")

        if len(project_name) < self.MIN_PROJECT_NAME_LENGTH:
            raise ValueError(
                f"Project name must be at least {self.MIN_PROJECT_NAME_LENGTH} characters"
            )

        if len(project_name) > self.MAX_PROJECT_NAME_LENGTH:
            raise ValueError(
                f"Project name cannot exceed {self.MAX_PROJECT_NAME_LENGTH} characters"
            )

        # Check for lowercase letters, numbers, hyphens only
        if not all(c.islower() or c.isdigit() or c == "-" for c in project_name):
            raise ValueError(
                "Project name must contain only lowercase letters, numbers, and hyphens"
            )

        # Cannot start or end with hyphen
        if project_name.startswith("-") or project_name.endswith("-"):
            raise ValueError("Project name cannot start or end with a hyphen")

        # Check for consecutive hyphens
        if "--" in project_name:
            raise ValueError("Project name cannot contain consecutive hyphens")

        return True

    def validate_framework(self, framework: str) -> bool:
        """
        Validate framework name

        @CODE:UIUX-VALIDATE-FRAMEWORK-001:VALIDATION
        """
        if framework not in self.VALID_FRAMEWORKS:
            raise ValueError(
                f"Invalid framework: {framework}\nSupported frameworks: {', '.join(self.VALID_FRAMEWORKS)}"
            )
        return True

    def validate_components(self, components: Optional[List[str]]) -> bool:
        """
        Validate component names

        @CODE:UIUX-VALIDATE-COMPONENTS-001:VALIDATION
        """
        if components is None:
            return True

        for component in components:
            if component not in self.AVAILABLE_COMPONENTS:
                raise ValueError(
                    f"Invalid component: {component}\nSupported components: {', '.join(self.AVAILABLE_COMPONENTS)}"
                )
        return True

    def create_config_directory(self, output_dir: Path, project_name: str) -> Path:
        """
        Create UI/UX configuration directory structure

        @CODE:UIUX-CONFIG-DIR-001:DIRECTORY
        """
        config_dir = output_dir / ".moai" / "ui" / "shadcn"

        if config_dir.exists():
            raise FileExistsError(f"UI/UX configuration already exists: {config_dir}")

        config_dir.mkdir(parents=True, exist_ok=False)
        return config_dir

    def create_components_json(
        self,
        config_dir: Path,
        framework: str,
        components: Optional[List[str]] = None
    ) -> Path:
        """
        Create components.json with shadcn/ui configuration

        @CODE:UIUX-COMPONENTS-JSON-001:CONFIG
        """
        # Default components if not specified
        if components is None:
            components = ["button", "input", "card"]

        # Framework-specific configuration
        framework_config = {
            "nextjs": {
                "framework": "nextjs",
                "aliases": {
                    "@/*": "./*",
                    "@/components/*": "./components/*",
                    "@/lib/*": "./lib/*"
                },
                "baseColor": "slate"
            },
            "react": {
                "framework": "react",
                "aliases": {
                    "@/*": "./src/*",
                    "@/components/*": "./src/components/*",
                    "@/lib/*": "./src/lib/*"
                },
                "baseColor": "slate"
            },
            "vite": {
                "framework": "vite",
                "aliases": {
                    "@/*": "./src/*",
                    "@/components/*": "./src/components/*",
                    "@/lib/*": "./src/lib/*"
                },
                "baseColor": "slate"
            }
        }

        components_config = {
            **framework_config[framework],
            "components": [
                {
                    "name": comp,
                    "installed": True,
                    "version": "1.0.0"
                }
                for comp in components
            ],
            "created": datetime.now().isoformat(),
            "version": "1.0.0-dev"
        }

        components_file = config_dir / "components.json"
        with open(components_file, "w") as f:
            json.dump(components_config, f, indent=2)

        return components_file

    def create_tailwind_config(self, config_dir: Path, dark_mode: bool = False) -> Path:
        """
        Create tailwind.config.js template with shadcn/ui theme

        @CODE:UIUX-TAILWIND-CONFIG-001:CONFIG
        """
        dark_mode_config = """
  darkMode: ["class"],
""" if dark_mode else ""

        tailwind_content = f"""/** @type {{import('tailwindcss').Config}} */
module.exports = {{
  content: [
    "./pages/**/*.{{js,ts,jsx,tsx}}",
    "./components/**/*.{{js,ts,jsx,tsx}}",
    "./app/**/*.{{js,ts,jsx,tsx}}",
  ],{dark_mode_config}
  theme: {{
    extend: {{
      colors: {{
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {{
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        }},
        secondary: {{
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        }},
        destructive: {{
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        }},
        muted: {{
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        }},
        accent: {{
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        }},
        popover: {{
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        }},
        card: {{
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        }},
      }},
      borderRadius: {{
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      }},
    }},
  }},
  plugins: [require("tailwindcss-animate")],
}}
"""

        tailwind_file = config_dir / "tailwind.config.template.js"
        tailwind_file.write_text(tailwind_content)
        return tailwind_file

    def create_component_structure(self, config_dir: Path) -> Path:
        """
        Create component directory structure for shadcn/ui

        @CODE:UIUX-COMPONENT-STRUCTURE-001:STRUCTURE
        """
        components_dir = config_dir / "components"
        components_dir.mkdir(exist_ok=False)

        # Create subdirectories for component organization
        ui_dir = components_dir / "ui"
        ui_dir.mkdir(exist_ok=False)

        # Create component index
        index_content = """\"\"\"
UI Components - shadcn/ui Component Library

Auto-generated component index for shadcn/ui setup.
\"\"\"

# Import all components here
"""
        (components_dir / "__init__.py").write_text(index_content)
        (ui_dir / "__init__.py").write_text('"""UI component library"""')

        return components_dir

    def execute(
        self,
        project_name: str,
        output_dir: Path,
        framework: str = "nextjs",
        components: Optional[List[str]] = None,
        dark_mode: bool = False
    ) -> CommandResult:
        """
        Execute /setup-shadcn-ui command

        @CODE:UIUX-EXECUTE-001:MAIN

        Args:
            project_name: Project name (lowercase, hyphens)
            output_dir: Output directory for configuration
            framework: Framework type (nextjs, react, vite)
            components: List of components to include
            dark_mode: Enable dark mode support

        Returns:
            CommandResult with success/failure status
        """
        # Validation (may raise exceptions)
        self.validate_project_name(project_name)
        self.validate_framework(framework)
        self.validate_components(components)

        try:
            # Create configuration directory
            config_dir = self.create_config_directory(Path(output_dir), project_name)

            # Create files
            files_created = []

            # components.json
            components_file = self.create_components_json(config_dir, framework, components)
            files_created.append(str(components_file.relative_to(output_dir)))

            # tailwind.config.template.js
            tailwind_file = self.create_tailwind_config(config_dir, dark_mode)
            files_created.append(str(tailwind_file.relative_to(output_dir)))

            # Component structure
            components_dir = self.create_component_structure(config_dir)
            files_created.append(str(components_dir.relative_to(output_dir)))

            message = f"✅ UI/UX setup for '{project_name}' completed successfully\n"
            message += f"📁 Location: {config_dir}\n"
            message += f"🎨 Framework: {framework}\n"
            message += f"🧩 Components: {len(components or ['button', 'input', 'card'])} selected\n"
            message += f"🌙 Dark Mode: {'Enabled' if dark_mode else 'Disabled'}\n"
            message += f"📝 Files created: {len(files_created)}"

            return CommandResult(
                success=True,
                config_dir=config_dir,
                files_created=files_created,
                message=message
            )

        except FileExistsError:
            # Re-raise specific validation errors for caller to handle
            raise
        except Exception as e:
            return CommandResult(
                success=False,
                config_dir=None,
                files_created=[],
                message=f"❌ Error setting up UI/UX",
                error=str(e)
            )


# @CODE:UIUX-MODULE-INIT-001:MODULE
# Create module-level command instance
setup_shadcn = SetupShadcnUICommand()
