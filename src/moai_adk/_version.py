"""Version information for MoAI-ADK."""

# Main version information
__version__ = "0.1.21"

# Essential version information only
VERSIONS = {
    "moai_adk": "0.1.21",
    "min_python": (3, 11),
    "core": "0.1.21",
    "templates": "0.1.21",
    "hooks": "0.1.4",
    "agents": "0.1.5",
    "commands": "0.1.4",
    "tag_system": "16-core",
    "tag_format": "16-core",
    "constitution": "1.0",
    "pipeline": "1.0.0",
}

# Version display formats
VERSION_FORMATS = {
    "full": f"MoAI-ADK v{__version__}",
    "short": f"v{__version__}",
    "banner": f"🗿 MoAI-ADK v{__version__}",
    "claude_md": f"# MoAI-ADK (MoAI Agentic Development Kit) v{__version__}",
}

def get_version(component: str = "moai_adk") -> str:
    """Get version for specific component."""
    return VERSIONS.get(component, __version__)

def get_version_format(format_type: str = "full") -> str:
    """Get formatted version string."""
    return VERSION_FORMATS.get(format_type, VERSION_FORMATS["full"])

def get_all_versions() -> dict:
    """Get all version information."""
    return VERSIONS.copy()

def get_min_python_version() -> tuple:
    """Get minimum Python version requirement."""
    return VERSIONS["min_python"]
