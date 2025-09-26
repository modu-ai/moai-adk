"""@FEATURE:VERSION-001 Version information for MoAI-ADK."""

# Main version information
__version__ = "0.1.28"

# Essential version information only
VERSIONS = {
    "moai_adk": "0.1.28",
    "min_python": (3, 10),
    "core": "0.1.28",
    "templates": "0.1.28",
    "hooks": "0.1.28",
    "agents": "0.1.28",
    "commands": "0.1.28",
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
