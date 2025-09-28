#!/usr/bin/env python3
# @TASK:VALIDATE-STRUCTURE-001
"""
Structure Validation Module

Validates MoAI-specific directory structure and configuration files.
Focuses on MoAI framework compliance verification.
"""

from pathlib import Path
from typing import Dict
import json
from ...utils.logger import get_logger

logger = get_logger(__name__)


def validate_moai_structure(project_path: Path) -> Dict[str, bool]:
    """Validate MoAI-specific directory structure."""
    moai_dir = project_path / ".moai"
    checks = {}

    # Check MoAI directory exists
    checks["has_moai_dir"] = moai_dir.exists() and moai_dir.is_dir()

    if not checks["has_moai_dir"]:
        logger.warning("No .moai directory found")
        return {"structure_valid": False, **checks}

    # Check essential MoAI subdirectories
    essential_moai_dirs = [
        "project",
        "specs",
        "memory",
        "indexes",
        "scripts",
    ]

    for dir_name in essential_moai_dirs:
        dir_path = moai_dir / dir_name
        checks[f"has_{dir_name}_dir"] = dir_path.exists() and dir_path.is_dir()

    # Check essential MoAI files
    essential_files = [
        "config.json",
        "project/product.md",
        "project/structure.md",
        "project/tech.md",
    ]

    for file_name in essential_files:
        file_path = moai_dir / file_name
        check_name = f"has_{file_name.replace('/', '_').replace('.', '_')}"
        checks[check_name] = file_path.exists() and file_path.is_file()

    # Validate config.json format
    config_file = moai_dir / "config.json"
    if config_file.exists():
        try:
            with open(config_file, 'r', encoding='utf-8') as f:
                config_data = json.load(f)
                checks["config_valid"] = isinstance(config_data, dict) and "version" in config_data
        except (json.JSONDecodeError, PermissionError):
            checks["config_valid"] = False
    else:
        checks["config_valid"] = False

    # Calculate overall structure validity
    passed = sum(1 for v in checks.values() if v and isinstance(v, bool))
    total = len([k for k, v in checks.items() if isinstance(v, bool)])
    checks["structure_valid"] = passed >= (total * 0.8)  # 80% pass rate

    logger.info(f"MoAI structure validation: {passed}/{total} checks passed")
    return checks