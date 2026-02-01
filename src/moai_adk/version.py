"""Version information for MoAI-ADK.

Provides version constants for template and MoAI framework.
"""

from importlib.metadata import PackageNotFoundError
from importlib.metadata import version as pkg_version
from pathlib import Path

# MoAI Framework Version
# Fallback version (only used when all other methods fail)
_FALLBACK_VERSION = "1.12.11"

# Template Schema Version
TEMPLATE_VERSION = "3.0.0"


def _get_version_from_pyproject() -> str | None:
    """Try to read version from pyproject.toml for development environment.

    This ensures the correct version is displayed even when the package
    metadata cache is stale (common with editable installs).
    """
    try:
        import tomllib

        # Find pyproject.toml relative to this file
        # Path: version.py -> moai_adk -> src -> project_root
        version_file = Path(__file__).resolve()
        project_root = version_file.parent.parent.parent
        pyproject_path = project_root / "pyproject.toml"

        if pyproject_path.exists():
            with open(pyproject_path, "rb") as f:
                data = tomllib.load(f)
                return data.get("project", {}).get("version")
    except Exception:
        pass
    return None


def _get_moai_version() -> str:
    """Get MoAI version with fallback chain.

    Priority:
    1. pyproject.toml (development environment - always accurate)
    2. importlib.metadata (installed package)
    3. Fallback version (last resort)
    """
    # Priority 1: Read from pyproject.toml (development environment)
    dev_version = _get_version_from_pyproject()
    if dev_version:
        return dev_version

    # Priority 2: Try importlib.metadata (installed package)
    try:
        return pkg_version("moai-adk")
    except PackageNotFoundError:
        pass

    # Priority 3: Fallback version
    return _FALLBACK_VERSION


MOAI_VERSION = _get_moai_version()

__all__ = ["MOAI_VERSION", "TEMPLATE_VERSION"]
