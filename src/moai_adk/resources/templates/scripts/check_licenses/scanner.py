#!/usr/bin/env python3
"""
Package license scanner for check_licenses module
"""

import subprocess
import json
from pathlib import Path
from typing import List
from .models import PackageLicense


def scan_python_packages(project_root: Path) -> List[PackageLicense]:
    """Scan Python packages for license information"""
    packages = []

    # Try to use pip-licenses if available
    try:
        result = subprocess.run(
            ["pip-licenses", "--format=json", "--with-license-file"],
            capture_output=True, text=True, cwd=project_root
        )
        if result.returncode == 0:
            data = json.loads(result.stdout)
            for pkg in data:
                packages.append(PackageLicense(
                    package=pkg.get("Name", "unknown"),
                    version=pkg.get("Version", "unknown"),
                    license=pkg.get("License", "UNKNOWN"),
                    license_info=None,  # Will be filled by analyzer
                    source="pip-licenses",
                    status="needs-review"
                ))
    except (subprocess.SubprocessError, json.JSONDecodeError, FileNotFoundError):
        # Fallback to requirements.txt scanning
        packages.extend(scan_requirements_file(project_root))

    return packages


def scan_requirements_file(project_root: Path) -> List[PackageLicense]:
    """Scan requirements.txt for packages"""
    packages = []
    requirements_file = project_root / "requirements.txt"

    if requirements_file.exists():
        content = requirements_file.read_text()
        for line in content.split('\n'):
            line = line.strip()
            if line and not line.startswith('#'):
                # Simple parsing: package==version or package>=version
                if '==' in line:
                    package, version = line.split('==')[0], line.split('==')[1]
                elif '>=' in line:
                    package, version = line.split('>=')[0], f">={line.split('>=')[1]}"
                else:
                    package, version = line, "unknown"

                packages.append(PackageLicense(
                    package=package.strip(),
                    version=version.strip(),
                    license="UNKNOWN",
                    license_info=None,
                    source="requirements.txt",
                    status="needs-review"
                ))

    return packages