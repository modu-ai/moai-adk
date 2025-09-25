"""@FEATURE:API-GEN-001 API documentation generator

Automatic generation of API documentation from Python source code.
Extracts docstrings and creates structured markdown documentation.

@REQ:API-GEN-001 → @TASK:API-GEN-001
"""

import ast
import logging
from pathlib import Path
from typing import Any, NamedTuple

import yaml

logger = logging.getLogger(__name__)


class ModuleInfo(NamedTuple):
    """Information about a Python module"""

    name: str
    path: Path
    package: str


class ApiGenerator:
    """@FEATURE:API-GEN-002 Automatic API documentation generator"""

    def __init__(self, project_root: str, source_dir: str = "src"):
        """Initialize API generator

        Args:
            project_root: Root directory of the project
            source_dir: Source code directory relative to project root
        """
        self.project_root = Path(project_root)
        self.source_dir = self.project_root / source_dir
        self.docs_dir = self.project_root / "docs"

    def scan_modules(self) -> list[ModuleInfo]:
        """@TASK:API-GEN-002 Scan and find Python modules"""
        modules = []

        if not self.source_dir.exists():
            return modules

        # Find all Python files
        for py_file in self.source_dir.glob("**/*.py"):
            if py_file.name.startswith("__pycache__"):
                continue

            # Calculate module name
            relative_path = py_file.relative_to(self.source_dir)
            module_parts = list(relative_path.parts[:-1])

            if py_file.name == "__init__.py":
                # For __init__.py, use the package name with __init__ suffix
                if module_parts:
                    module_name = ".".join(module_parts) + ".__init__"
                else:
                    module_name = "__init__"  # Root __init__.py
            else:
                # For regular modules, include the file stem
                module_parts.append(py_file.stem)
                module_name = ".".join(module_parts)

            if not module_name:
                continue

            package = module_parts[0] if module_parts else ""

            modules.append(ModuleInfo(name=module_name, path=py_file, package=package))

        return modules

    def parse_module_docs(self, module_path: str) -> dict[str, Any]:
        """@TASK:API-GEN-003 Parse docstrings from a module"""
        doc_info = {"module_doc": "", "functions": {}, "classes": {}}

        try:
            # Convert module path to actual file path
            parts = module_path.replace("src.", "").split(".")
            module_file = self.source_dir
            for part in parts:
                module_file = module_file / part

            # Try as .py file first
            if not module_file.exists():
                module_file = module_file.with_suffix(".py")

            if not module_file.exists():
                # Try as __init__.py in directory
                module_file = module_file.parent / module_file.name / "__init__.py"

            if not module_file.exists():
                return doc_info

            # Parse the Python AST
            with open(module_file, encoding="utf-8") as f:
                tree = ast.parse(f.read())

            # Extract module docstring
            if (
                tree.body
                and isinstance(tree.body[0], ast.Expr)
                and isinstance(tree.body[0].value, ast.Constant)
                and isinstance(tree.body[0].value.value, str)
            ):
                doc_info["module_doc"] = tree.body[0].value.value.strip()

            # Extract functions and classes
            for node in ast.walk(tree):
                if isinstance(node, ast.FunctionDef):
                    doc_info["functions"][node.name] = {
                        "docstring": ast.get_docstring(node) or "",
                        "args": [arg.arg for arg in node.args.args],
                    }
                elif isinstance(node, ast.ClassDef):
                    doc_info["classes"][node.name] = {
                        "docstring": ast.get_docstring(node) or "",
                        "methods": [],
                    }

        except Exception as e:
            logger.warning(f"Failed to parse {module_path}: {e}")

        return doc_info

    def generate_nav_structure(self) -> dict[str, Any]:
        """@TASK:API-GEN-004 Generate navigation structure for API docs"""
        modules = self.scan_modules()
        nav_structure = {}

        for module in modules:
            parts = module.name.split(".")
            current = nav_structure

            for part in parts[:-1]:  # All except last
                if part not in current:
                    current[part] = {}
                current = current[part]

            # Last part is the actual module
            if len(parts) > 0:
                if isinstance(current, dict):
                    current[parts[-1]] = f"api/{module.name}.md"

        return nav_structure

    def generate_api_docs(self, docs_root: str) -> None:
        """@TASK:API-GEN-005 Generate API documentation files"""
        docs_path = Path(docs_root)
        api_dir = docs_path / "api"
        api_dir.mkdir(exist_ok=True)

        modules = self.scan_modules()

        for module in modules:
            doc_info = self.parse_module_docs(module.name)

            # Generate markdown content
            md_content = self._generate_module_markdown(module.name, doc_info)

            # Write to file
            md_file = api_dir / f"{module.name}.md"
            md_file.parent.mkdir(parents=True, exist_ok=True)
            md_file.write_text(md_content)

    def _generate_module_markdown(
        self, module_name: str, doc_info: dict[str, Any]
    ) -> str:
        """Generate markdown content for a module"""
        lines = []

        # Title
        lines.append(f"# {module_name}")
        lines.append("")

        # Module docstring
        if doc_info["module_doc"]:
            lines.append(doc_info["module_doc"])
            lines.append("")

        # Functions
        if doc_info["functions"]:
            lines.append("## Functions")
            lines.append("")

            for func_name, func_info in doc_info["functions"].items():
                lines.append(f"### {func_name}")
                lines.append("")

                if func_info["docstring"]:
                    lines.append(func_info["docstring"])
                    lines.append("")

                # Function signature
                args_str = ", ".join(func_info["args"])
                lines.append("```python")
                lines.append(f"{func_name}({args_str})")
                lines.append("```")
                lines.append("")

        # Classes
        if doc_info["classes"]:
            lines.append("## Classes")
            lines.append("")

            for class_name, class_info in doc_info["classes"].items():
                lines.append(f"### {class_name}")
                lines.append("")

                if class_info["docstring"]:
                    lines.append(class_info["docstring"])
                    lines.append("")

        return "\n".join(lines)

    def update_mkdocs_nav(self, mkdocs_config_path: str) -> None:
        """@TASK:API-GEN-006 Update mkdocs.yml with API navigation"""
        config_path = Path(mkdocs_config_path)

        if not config_path.exists():
            return

        try:
            with open(config_path) as f:
                config = yaml.safe_load(f)

            # Add API Reference section
            nav = config.get("nav", [])

            # Check if API Reference already exists
            api_exists = any(
                isinstance(item, dict) and "API Reference" in item for item in nav
            )

            if not api_exists:
                api_nav = {"API Reference": "api/index.md"}
                nav.append(api_nav)
                config["nav"] = nav

                with open(config_path, "w") as f:
                    yaml.dump(config, f, default_flow_style=False)

        except Exception as e:
            logger.error(f"Failed to update mkdocs nav: {e}")

    def generate_module_index(self) -> str:
        """@TASK:API-GEN-007 Generate module index content"""
        modules = self.scan_modules()

        lines = ["# API Reference", ""]
        lines.append(
            "This section contains the complete API documentation for MoAI-ADK."
        )
        lines.append("")

        if modules:
            # Group by package
            packages = {}
            for module in modules:
                package = module.package
                if package not in packages:
                    packages[package] = []
                packages[package].append(module)

            for package_name, package_modules in packages.items():
                lines.append(f"## {package_name}")
                lines.append("")

                for module in package_modules:
                    lines.append(f"- [{module.name}](api/{module.name}.md)")

                lines.append("")

        return "\n".join(lines)
