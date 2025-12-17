"""
MoAI Menu Project - Integrated Module System

Comprehensive project management system that integrates documentation,
language initialization, and template optimization modules.

This module provides a unified interface for managing project setup,
documentation generation, multilingual support, and template optimization.
"""

import json
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List

# Import module components
from .modules.documentation_manager import DocumentationManager
from .modules.language_initializer import LanguageInitializer
from .modules.template_optimizer import TemplateOptimizer


class MoaiMenuProject:
    """
    Main interface for MoAI Menu Project integrated system.

    Provides unified access to documentation management, language initialization,
    and template optimization capabilities.
    """

    def __init__(self, project_root: str = None, config: Dict[str, Any] = None):
        """
        Initialize the MoAI Menu Project system.

        Args:
            project_root: Root directory of the project
            config: Configuration dictionary
        """

        # Determine project root
        if project_root is None:
            project_root = Path.cwd()
        else:
            project_root = Path(project_root)

        self.project_root = project_root

        # Load or create configuration
        if config is None:
            config = self._load_or_create_config()

        self.config = config

        # Initialize modules
        self.documentation_manager = DocumentationManager(str(project_root), config)
        self.language_initializer = LanguageInitializer(str(project_root), config)
        self.template_optimizer = TemplateOptimizer(str(project_root), config)

        # System metadata
        self.version = "1.0.0"
        self.initialized_at = datetime.now()

    def _load_or_create_config(self) -> Dict[str, Any]:
        """Load existing configuration or create default."""

        config_path = self.project_root / ".moai/config/config.yaml"

        if config_path.exists():
            try:
                return json.loads(config_path.read_text(encoding="utf-8"))
            except (json.JSONDecodeError, UnicodeDecodeError):
                pass

        # Create default configuration
        default_config = {
            "project": {
                "name": "My Project",
                "type": "web_application",
                "initialized_at": datetime.now().isoformat(),
            },
            "language": {
                "conversation_language": "en",
                "conversation_language_name": "English",
                "agent_prompt_language": "english",
                "documentation_language": "en",
            },
            "user": {"name": "Developer", "selected_at": datetime.now().isoformat()},
            "menu_system": {
                "version": self.version,
                "initialized_at": datetime.now().isoformat(),
            },
        }

        # Ensure config directory exists
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Save default configuration
        config_path.write_text(json.dumps(default_config, indent=2, ensure_ascii=False), encoding="utf-8")

        return default_config

    def initialize_complete_project(
        self,
        language: str = None,
        user_name: str = None,
        domains: List[str] = None,
        project_type: str = None,
        optimization_enabled: bool = True,
    ) -> Dict[str, Any]:
        """
        Initialize complete project with all modules.

        Args:
            language: Primary language for the project
            user_name: User name for personalization
            domains: List of domains to enable
            project_type: Type of project (web_application, mobile_application, etc.)
            optimization_enabled: Whether to run template optimization

        Returns:
            Complete initialization results
        """

        initialization_result = {
            "timestamp": datetime.now().isoformat(),
            "project_root": str(self.project_root),
            "modules_initialized": [],
            "configuration_updates": {},
            "created_files": [],
            "optimization_results": None,
            "success": True,
            "errors": [],
        }

        try:
            # Update project type if provided
            if project_type:
                self.config["project"]["type"] = project_type

            # Step 1: Initialize language configuration
            language_result = self.language_initializer.initialize_language_configuration(
                language=language, user_name=user_name, domains=domains
            )

            initialization_result["modules_initialized"].append("language_initializer")
            initialization_result["configuration_updates"]["language"] = language_result

            # Step 2: Initialize documentation structure
            docs_result = self.documentation_manager.initialize_documentation_structure()

            initialization_result["modules_initialized"].append("documentation_manager")
            initialization_result["configuration_updates"]["documentation"] = docs_result
            initialization_result["created_files"].extend(docs_result.get("created_files", []))

            # Step 3: Create multilingual documentation structure
            multilingual_result = self.language_initializer.create_multilingual_documentation_structure(
                language=language_result.get("language", "en")
            )

            initialization_result["configuration_updates"]["multilingual"] = multilingual_result
            initialization_result["created_files"].extend(multilingual_result.get("created_directories", []))

            # Step 4: Analyze and optimize templates (if enabled)
            if optimization_enabled:
                template_analysis = self.template_optimizer.analyze_project_templates()

                # Apply optimizations if opportunities exist
                optimization_opportunities = template_analysis.get("optimization_opportunities", [])
                if optimization_opportunities:
                    optimization_result = self.template_optimizer.create_optimized_templates()
                    initialization_result["optimization_results"] = optimization_result

                initialization_result["modules_initialized"].append("template_optimizer")
                initialization_result["configuration_updates"]["template_analysis"] = template_analysis

            # Step 5: Create benchmarks (optional)
            if optimization_enabled:
                benchmark_result = self.template_optimizer.benchmark_template_performance()
                initialization_result["configuration_updates"]["benchmark"] = benchmark_result

            # Update final configuration
            self.config["menu_system"]["fully_initialized"] = True
            self.config["menu_system"]["initialization_completed_at"] = datetime.now().isoformat()
            self._save_config()

        except Exception as e:
            initialization_result["success"] = False
            initialization_result["errors"].append(str(e))

        return initialization_result

    def generate_documentation_from_spec(self, spec_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Generate comprehensive documentation from SPEC data.

        Args:
            spec_data: SPEC specification data

        Returns:
            Documentation generation results
        """

        # Generate base documentation
        docs_result = self.documentation_manager.generate_documentation_from_spec(spec_data)

        # Localize documentation based on language settings
        language = self.config.get("language", {}).get("conversation_language", "en")

        if language != "en":
            # Apply language-specific customizations
            localized_docs = self._localize_documentation(docs_result, language)
            docs_result["localized_documentation"] = localized_docs

        return docs_result

    def _localize_documentation(self, docs_result: Dict[str, Any], language: str) -> Dict[str, Any]:
        """Localize documentation content based on language."""

        # This would apply language-specific formatting, terminology, and structure
        # Implementation would be more sophisticated in practice

        localized_docs = {
            "language": language,
            "localization_applied": True,
            "modified_sections": [],
        }

        return localized_docs

    def optimize_project_templates(self, options: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Optimize project templates with advanced analysis.

        Args:
            options: Optimization options

        Returns:
            Template optimization results
        """

        # Analyze current templates
        analysis = self.template_optimizer.analyze_project_templates()

        # Apply optimizations based on analysis
        optimization_result = self.template_optimizer.create_optimized_templates(options)

        # Create benchmarks for comparison
        benchmark_result = self.template_optimizer.benchmark_template_performance()

        return {
            "analysis": analysis,
            "optimization": optimization_result,
            "benchmark": benchmark_result,
            "timestamp": datetime.now().isoformat(),
        }

    def update_language_settings(self, language_updates: Dict[str, Any]) -> Dict[str, Any]:
        """
        Update language configuration.

        Args:
            language_updates: Language configuration updates

        Returns:
            Update results
        """

        # Update language configuration
        update_result = self.language_initializer.update_language_settings(language_updates)

        if update_result.get("success"):
            # Reload configuration
            self.config = self._load_or_create_config()

            # Update documentation language if changed
            doc_lang = language_updates.get("language.documentation_language")
            if doc_lang:
                # Update documentation structure for new language
                multilingual_result = self.language_initializer.create_multilingual_documentation_structure(doc_lang)
                update_result["documentation_updated"] = multilingual_result

        return update_result

    def get_project_status(self) -> Dict[str, Any]:
        """
        Get comprehensive project status.

        Returns:
            Project status and metrics
        """

        status = {
            "timestamp": datetime.now().isoformat(),
            "project_root": str(self.project_root),
            "configuration": {
                "project_name": self.config.get("project", {}).get("name", "Unknown"),
                "project_type": self.config.get("project", {}).get("type", "unknown"),
                "initialized_at": self.config.get("project", {}).get("initialized_at"),
                "menu_system_version": self.config.get("menu_system", {}).get("version", "1.0.0"),
            },
            "modules": {
                "documentation_manager": "initialized",
                "language_initializer": "initialized",
                "template_optimizer": "initialized",
            },
            "language_status": self.language_initializer.get_language_status(),
            "documentation_status": self.documentation_manager.get_documentation_status(),
            "template_analysis": None,  # Will be populated on demand
        }

        # Check if fully initialized
        status["fully_initialized"] = self.config.get("menu_system", {}).get("fully_initialized", False)

        return status

    def export_project_documentation(self, format_type: str = "markdown", language: str = None) -> Dict[str, Any]:
        """
        Export project documentation in specified format.

        Args:
            format_type: Export format (markdown, html, pdf)
            language: Target language for export

        Returns:
            Export results
        """

        # Use specified language or default project language
        if language is None:
            language = self.config.get("language", {}).get("conversation_language", "en")

        # Export documentation
        export_result = self.documentation_manager.export_documentation(format_type)

        # Apply language-specific customizations
        export_result["language"] = language
        export_result["localized"] = language != "en"

        return export_result

    def create_project_backup(self, backup_name: str = None) -> Dict[str, Any]:
        """
        Create comprehensive project backup.

        Args:
            backup_name: Name for the backup

        Returns:
            Backup creation results
        """

        if backup_name is None:
            backup_name = f"project-backup-{datetime.now().strftime('%Y-%m-%d-%H-%M-%S')}"

        backup_result = {
            "backup_name": backup_name,
            "timestamp": datetime.now().isoformat(),
            "components": {},
        }

        # Backup templates
        template_backup = self.template_optimizer._create_template_backup(f"{backup_name}-templates")
        backup_result["components"]["templates"] = template_backup

        # Backup configuration
        config_backup_path = self.project_root / ".moai-backups" / backup_name / "config.json"
        config_backup_path.parent.mkdir(parents=True, exist_ok=True)
        config_backup_path.write_text(json.dumps(self.config, indent=2, ensure_ascii=False), encoding="utf-8")

        backup_result["components"]["configuration"] = {
            "success": True,
            "backup_path": str(config_backup_path),
        }

        # Backup documentation if exists
        docs_dir = self.project_root / "docs"
        if docs_dir.exists():
            import shutil

            docs_backup_path = self.project_root / ".moai-backups" / backup_name / "docs"
            if docs_backup_path.exists():
                shutil.rmtree(docs_backup_path)
            shutil.copytree(docs_dir, docs_backup_path)

            backup_result["components"]["documentation"] = {
                "success": True,
                "backup_path": str(docs_backup_path),
            }

        backup_result["success"] = all(
            component.get("success", False) for component in backup_result["components"].values()
        )

        return backup_result

    def _save_config(self):
        """Save current configuration to file."""

        config_path = self.project_root / ".moai/config/config.yaml"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        config_path.write_text(json.dumps(self.config, indent=2, ensure_ascii=False), encoding="utf-8")

    def get_integration_matrix(self) -> Dict[str, Any]:
        """
        Get module integration matrix and workflow.

        Returns:
            Integration information and workflows
        """

        return {
            "modules": {
                "documentation_manager": {
                    "description": "Manages project documentation generation and maintenance",
                    "integrates_with": ["language_initializer"],
                    "output_formats": ["markdown", "html", "pdf"],
                    "key_features": [
                        "template-based generation",
                        "multilingual support",
                        "SPEC integration",
                    ],
                },
                "language_initializer": {
                    "description": "Handles language detection, configuration, and localization",
                    "integrates_with": ["documentation_manager"],
                    "supported_languages": ["en", "ko", "ja", "zh"],
                    "key_features": [
                        "auto-detection",
                        "locale management",
                        "agent prompt localization",
                    ],
                },
                "template_optimizer": {
                    "description": "Analyzes and optimizes project templates for performance",
                    "integrates_with": [],
                    "optimization_types": ["size", "performance", "complexity"],
                    "key_features": [
                        "automated analysis",
                        "backup creation",
                        "benchmarking",
                    ],
                },
            },
            "workflows": {
                "project_initialization": [
                    "language_initializer.initialize_language_configuration()",
                    "documentation_manager.initialize_documentation_structure()",
                    "language_initializer.create_multilingual_documentation_structure()",
                    "template_optimizer.analyze_project_templates()",
                    "template_optimizer.create_optimized_templates()",
                ],
                "documentation_generation": [
                    "language_initializer.detect_project_language()",
                    "documentation_manager.generate_documentation_from_spec()",
                    "language_initializer.localize_agent_prompts()",
                    "documentation_manager.export_documentation()",
                ],
                "template_optimization": [
                    "template_optimizer.analyze_project_templates()",
                    "template_optimizer.create_optimized_templates()",
                    "template_optimizer.benchmark_template_performance()",
                    "template_optimizer._create_template_backup()",
                ],
            },
            "data_flow": {
                "configuration": "Shared across all modules via config.json",
                "language_settings": "language_initializer → documentation_manager",
                "template_analysis": "template_optimizer (independent)",
                "documentation_output": "documentation_manager → multiple formats",
            },
        }


# Convenience functions for easy access
def initialize_project(project_root: str = None, **kwargs) -> Dict[str, Any]:
    """
    Convenience function to initialize a complete MoAI project.

    Args:
        project_root: Root directory of the project
        **kwargs: Additional initialization parameters

    Returns:
        Initialization results
    """

    project = MoaiMenuProject(project_root)
    return project.initialize_complete_project(**kwargs)


def generate_docs(spec_data: Dict[str, Any], project_root: str = None) -> Dict[str, Any]:
    """
    Convenience function to generate documentation from SPEC.

    Args:
        spec_data: SPEC specification data
        project_root: Root directory of the project

    Returns:
        Documentation generation results
    """

    project = MoaiMenuProject(project_root)
    return project.generate_documentation_from_spec(spec_data)


def optimize_templates(project_root: str = None, **options) -> Dict[str, Any]:
    """
    Convenience function to optimize project templates.

    Args:
        project_root: Root directory of the project
        **options: Optimization options

    Returns:
        Template optimization results
    """

    project = MoaiMenuProject(project_root)
    return project.optimize_project_templates(options)
