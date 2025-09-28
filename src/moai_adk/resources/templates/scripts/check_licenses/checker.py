#!/usr/bin/env python3
"""
License checker class for check_licenses module
"""

import json
from pathlib import Path
from typing import List, Dict, Any
from .models import PackageLicense
from .database import init_license_database
from .scanner import scan_python_packages
from .analyzer import analyze_license_compliance


class LicenseChecker:
    """License compatibility checker"""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_dir = project_root / ".moai"

        # License database
        self.license_db = init_license_database()

        # Policy settings
        self.policy = self.load_license_policy()

        # Results storage
        self.scan_results: List[PackageLicense] = []
        self.violations: List[str] = []
        self.warnings: List[str] = []

    def load_license_policy(self) -> Dict[str, Any]:
        """Load license policy configuration"""
        default_policy = {
            "allowed_risk_levels": ["low", "medium"],
            "blocked_licenses": ["AGPL-3.0"],
            "require_review": ["GPL-3.0"],
            "auto_approve": ["MIT", "Apache-2.0", "BSD-3-Clause"]
        }

        # Load from .moai/config.json
        config_file = self.moai_dir / "config.json"
        if config_file.exists():
            try:
                moai_config = json.loads(config_file.read_text())
                license_config = moai_config.get("license_policy", {})
                default_policy.update(license_config)
            except Exception:
                pass  # Use defaults

        return default_policy

    def run_license_check(self) -> List[PackageLicense]:
        """Run comprehensive license check"""
        print("🔒 Starting license compliance check...")

        # Scan packages
        self.scan_results = scan_python_packages(self.project_root)
        print(f"  Found {len(self.scan_results)} packages")

        # Analyze compliance
        analyze_license_compliance(self.scan_results, self.license_db, self.policy)

        # Check for violations
        self.check_violations()

        return self.scan_results

    def check_violations(self) -> None:
        """Check for license violations"""
        for package in self.scan_results:
            if package.status == "non-compliant":
                self.violations.append(f"Package {package.package} has non-compliant license: {package.license}")
            elif package.status == "needs-review":
                self.warnings.append(f"Package {package.package} needs license review: {package.license}")

    def has_violations(self) -> bool:
        """Check if there are any violations"""
        return len(self.violations) > 0