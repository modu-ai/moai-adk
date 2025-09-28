#!/usr/bin/env python3
"""
License analysis for check_licenses module
"""

from typing import List, Dict
from .models import PackageLicense, LicenseInfo


def analyze_license_compliance(
    packages: List[PackageLicense],
    license_db: Dict[str, LicenseInfo],
    policy: Dict[str, any]
) -> None:
    """Analyze license compliance for packages"""
    for package in packages:
        # Normalize license name
        normalized_license = normalize_license_name(package.license)

        # Get license info from database
        package.license_info = license_db.get(normalized_license, license_db["UNKNOWN"])

        # Determine compliance status
        package.status = determine_compliance_status(package.license_info, policy)


def normalize_license_name(license_name: str) -> str:
    """Normalize license name for database lookup"""
    if not license_name or license_name.upper() in ["UNKNOWN", "NONE", ""]:
        return "UNKNOWN"

    # Common license name mappings
    mappings = {
        "MIT License": "MIT",
        "Apache Software License": "Apache-2.0",
        "Apache License 2.0": "Apache-2.0",
        "BSD License": "BSD-3-Clause",
        "GNU General Public License v3": "GPL-3.0",
        "GNU Affero General Public License v3": "AGPL-3.0",
    }

    # Try exact match first
    if license_name in mappings:
        return mappings[license_name]

    # Try partial matching
    license_upper = license_name.upper()
    if "MIT" in license_upper:
        return "MIT"
    elif "APACHE" in license_upper:
        return "Apache-2.0"
    elif "BSD" in license_upper:
        return "BSD-3-Clause"
    elif "GPL" in license_upper and "AGPL" not in license_upper:
        return "GPL-3.0"
    elif "AGPL" in license_upper:
        return "AGPL-3.0"

    return "UNKNOWN"


def determine_compliance_status(license_info: LicenseInfo, policy: Dict[str, any]) -> str:
    """Determine compliance status based on policy"""
    risk_level = license_info.risk_level
    allowed_risks = policy.get("allowed_risk_levels", ["low", "medium"])

    if risk_level in allowed_risks:
        return "compliant"
    elif risk_level == "critical":
        return "non-compliant"
    else:
        return "needs-review"