"""
Configuration management for MoAI-ADK projects.

Handles project configuration, runtime settings, and validation.
"""

from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Dict, List


@dataclass
class RuntimeConfig:
    """Runtime configuration for the project."""
    name: str
    performance: int = 4
    
    def __post_init__(self) -> None:
        """Validate runtime configuration."""
        match self.name:
            case "node" | "python" | "tsx":
                pass  # Valid runtime
            case _:
                raise ValueError(f"Unsupported runtime: {self.name}")

        if not 1 <= self.performance <= 5:
            raise ValueError(f"Performance rating must be 1-5, got: {self.performance}")


@dataclass
class Config:
    """Main configuration for MoAI-ADK projects."""
    name: str
    template: str = "standard"
    runtime: RuntimeConfig = field(default_factory=lambda: RuntimeConfig("python"))
    tech_stack: List[str] = field(default_factory=list)
    path: str = ""
    backup_enabled: bool = False
    skip_install: bool = False
    silent: bool = False
    is_existing_project: bool = False
    force_overwrite: bool = False
    force_copy: bool = False  # Force file copying instead of symlinks
    include_github: bool = True  # Include GitHub workflows
    initialize_git: bool = True  # Initialize Git repository
    created_at: datetime | None = None
    templates_mode: str = "copy"  # 'copy' (default) or 'package' (no _templates copy)

    def __init__(self, name: str, **kwargs):
        """Initialize Config with backward compatibility for project_path parameter."""
        # Handle backward compatibility: project_path -> path
        if 'project_path' in kwargs and 'path' not in kwargs:
            kwargs['path'] = kwargs.pop('project_path')
        elif 'project_path' in kwargs:
            # Both provided, remove project_path
            kwargs.pop('project_path')

        # Set defaults
        self.name = name
        self.template = kwargs.get('template', 'standard')
        self.runtime = kwargs.get('runtime', RuntimeConfig("python"))
        self.tech_stack = kwargs.get('tech_stack', [])
        self.path = kwargs.get('path', "")
        self.backup_enabled = kwargs.get('backup_enabled', False)
        self.skip_install = kwargs.get('skip_install', False)
        self.silent = kwargs.get('silent', False)
        self.is_existing_project = kwargs.get('is_existing_project', False)
        self.force_overwrite = kwargs.get('force_overwrite', False)
        self.force_copy = kwargs.get('force_copy', False)
        self.include_github = kwargs.get('include_github', True)
        self.initialize_git = kwargs.get('initialize_git', True)
        self.created_at = kwargs.get('created_at', None)
        self.templates_mode = kwargs.get('templates_mode', 'copy')

        self.__post_init__()
    
    def __post_init__(self) -> None:
        """Initialize computed fields."""
        if not self.path:
            self.path = str(Path.cwd() / self.name)

        if self.created_at is None:
            self.created_at = datetime.now()

        self._validate()
    
    def _validate(self) -> None:
        """Validate configuration parameters."""
        if not self.name:
            raise ValueError("Project name is required")
        
        # Allow alphanumeric, hyphens, underscores, and dots
        if not self.name.replace("-", "").replace("_", "").replace(".", "").isalnum():
            raise ValueError(f"❌ Project name '{self.name}' contains invalid characters. Use only letters, numbers, dots (.), hyphens (-), and underscores (_)")
        
        if self.template not in ["minimal", "standard", "enterprise"]:
            raise ValueError(f"Unsupported template: {self.template}")
        
        # templates_mode validation
        if str(self.templates_mode).lower() not in ["copy", "package"]:
            raise ValueError("templates_mode must be 'copy' or 'package'")
        
        # Validate tech stack
        valid_tech = {
            "nextjs", "react", "vue", "nuxt", "angular", "svelte",
            "typescript", "javascript", "python",
            "tailwind", "scss", "css",
            "node", "deno", "bun",
            "express", "fastapi", "django", "flask", "spring", "springboot", "spring-boot",
            "postgresql", "mysql", "sqlite", "mongodb",
            "redis", "docker", "kubernetes",
            "rust", "go", "java"
        }
        
        for tech in self.tech_stack:
            if tech.lower() not in valid_tech:
                raise ValueError(f"Unsupported technology: {tech}")
    
    @property
    def project_path(self) -> Path:
        """Get project path as Path object."""
        return Path(self.path)
    
    @property
    def project_type(self) -> str:
        """Determine project type based on tech stack."""
        web_techs = {"nextjs", "react", "vue", "angular", "svelte"}
        api_techs = {"fastapi", "django", "flask"}

        if any(tech in web_techs for tech in self.tech_stack):
            return "web"
        elif any(tech in api_techs for tech in self.tech_stack):
            return "api"
        elif "python" in self.tech_stack:
            return "python"
        else:
            return "web"  # Default
    
    def get_template_context(self) -> Dict[str, str | int | List[str]]:
        """Get template rendering context."""
        tech_stack_str = ", ".join(self.tech_stack) if self.tech_stack else f"{self.runtime.name}, {self.template}"
        
        return {
            "project_name": self.name,
            "project_type": self.project_type,
            "template": self.template,
            "runtime_name": self.runtime.name,
            "runtime_performance": self.runtime.performance,
            "tech_stack": tech_stack_str,
            "tech_stack_list": self.tech_stack,
            "created_date": f"{self.created_at.year}. {self.created_at.month}. {self.created_at.day}." if self.created_at else "",
            "created_year": self.created_at.year if self.created_at else datetime.now().year,
            "version": "1.0",
        }
    
    def to_dict(self) -> Dict:
        """Convert config to dictionary."""
        return {
            "name": self.name,
            "template": self.template,
            "runtime": {
                "name": self.runtime.name,
                "performance": self.runtime.performance,
            },
            "tech_stack": self.tech_stack,
            "project_type": self.project_type,
            "path": self.path,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }
