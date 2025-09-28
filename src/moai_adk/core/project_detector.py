"""
@FEATURE:PROJECT-001 Project type detection for MoAI-ADK

Handles project type, language, and framework detection based on file analysis.
Extracted from system_manager.py for TRUST compliance (≤300 LOC).
"""

import json
from pathlib import Path
from typing import Any

from ..utils.logger import get_logger

logger = get_logger(__name__)


class ProjectDetector:
    """@TASK:PROJECT-DETECTOR-001 Detects project type, language, and frameworks."""

    def detect_project_type(self, project_path) -> dict[str, Any]:
        """
        Detect project type based on existing files.

        Args:
            project_path: Path to project directory

        Returns:
            dict: Detected project information
        """
        project_path = Path(project_path)
        detected = {
            "type": "unknown",
            "language": "unknown",
            "frameworks": [],
            "build_tools": [],
            "files_found": [],
        }

        logger.info(f"Detecting project type in: {project_path}")

        # Check for various project files
        project_files = {
            "package.json": {"type": "nodejs", "language": "javascript"},
            "requirements.txt": {"type": "python", "language": "python"},
            "pyproject.toml": {"type": "python", "language": "python"},
            "Cargo.toml": {"type": "rust", "language": "rust"},
            "go.mod": {"type": "go", "language": "go"},
            "pom.xml": {"type": "java", "language": "java"},
            "build.gradle": {"type": "java", "language": "java"},
            "Gemfile": {"type": "ruby", "language": "ruby"},
            "composer.json": {"type": "php", "language": "php"},
        }

        for file_name, info in project_files.items():
            file_path = project_path / file_name
            if file_path.exists():
                detected["files_found"].append(file_name)
                detected["type"] = info["type"]
                detected["language"] = info["language"]
                logger.info(f"Found {file_name}, detected as {info['language']} project")

        # Detect frameworks and build tools
        if (project_path / "package.json").exists():
            package_analysis = self._analyze_package_json(project_path / "package.json")
            detected.update(package_analysis)

        logger.info(f"Project detection completed: {detected}")
        return detected

    def _analyze_package_json(self, package_json_path) -> dict[str, Any]:
        """Analyze package.json for frameworks and dependencies."""
        try:
            with open(package_json_path, encoding="utf-8") as f:
                package_data = json.load(f)

            frameworks = []
            build_tools = []

            # Check dependencies and devDependencies
            all_deps = {}
            all_deps.update(package_data.get("dependencies", {}))
            all_deps.update(package_data.get("devDependencies", {}))

            # Detect frameworks
            framework_indicators = {
                "react": ["react", "@types/react"],
                "vue": ["vue", "@vue/cli"],
                "angular": ["@angular/core", "@angular/cli"],
                "svelte": ["svelte"],
                "nextjs": ["next"],
                "nuxtjs": ["nuxt"],
                "express": ["express"],
                "fastify": ["fastify"],
            }

            for framework, indicators in framework_indicators.items():
                if any(indicator in all_deps for indicator in indicators):
                    frameworks.append(framework)
                    logger.info(f"Detected framework: {framework}")

            # Detect build tools
            build_tool_indicators = {
                "webpack": ["webpack"],
                "vite": ["vite"],
                "rollup": ["rollup"],
                "parcel": ["parcel"],
                "typescript": ["typescript", "@types/node"],
            }

            for tool, indicators in build_tool_indicators.items():
                if any(indicator in all_deps for indicator in indicators):
                    build_tools.append(tool)
                    logger.info(f"Detected build tool: {tool}")

            has_scripts = bool(package_data.get("scripts"))
            scripts = list(package_data.get("scripts", {}).keys())

            logger.info(f"Package.json analysis: frameworks={frameworks}, build_tools={build_tools}")

            return {
                "frameworks": frameworks,
                "build_tools": build_tools,
                "has_scripts": has_scripts,
                "scripts": scripts,
            }

        except Exception as e:
            logger.error("Error analyzing package.json: %s", e)
            return {"frameworks": [], "build_tools": []}

    def should_create_package_json(self, config) -> bool:
        """
        Check if package.json should be created based on project configuration.

        Args:
            config: Project configuration

        Returns:
            bool: True if package.json should be created
        """
        # Only create package.json for explicit Node.js/web projects
        should_create = config.runtime.name in ["node", "tsx"] or any(
            tech in config.tech_stack
            for tech in ["nextjs", "react", "vue", "angular", "svelte"]
        )

        logger.info(f"Should create package.json: {should_create} (runtime: {config.runtime.name})")
        return should_create

    def detect_language_from_files(self, project_path) -> str:
        """
        Detect primary language based on file extensions in project.

        Args:
            project_path: Path to project directory

        Returns:
            str: Detected primary language
        """
        project_path = Path(project_path)

        if not project_path.exists():
            logger.warning(f"Project path does not exist: {project_path}")
            return "unknown"

        language_extensions = {
            "python": [".py", ".pyx", ".pyi"],
            "javascript": [".js", ".jsx", ".mjs"],
            "typescript": [".ts", ".tsx"],
            "rust": [".rs"],
            "go": [".go"],
            "java": [".java"],
            "ruby": [".rb"],
            "php": [".php"],
            "cpp": [".cpp", ".cxx", ".cc"],
            "c": [".c"],
        }

        file_counts = {lang: 0 for lang in language_extensions}

        try:
            for file_path in project_path.rglob("*"):
                if file_path.is_file():
                    suffix = file_path.suffix.lower()
                    for lang, extensions in language_extensions.items():
                        if suffix in extensions:
                            file_counts[lang] += 1
        except Exception as e:
            logger.error(f"Error scanning files: {e}")
            return "unknown"

        # Return language with most files
        detected_language = max(file_counts, key=file_counts.get)
        if file_counts[detected_language] > 0:
            logger.info(f"Detected language: {detected_language} ({file_counts[detected_language]} files)")
            return detected_language
        else:
            logger.info("No specific language detected from file extensions")
            return "unknown"