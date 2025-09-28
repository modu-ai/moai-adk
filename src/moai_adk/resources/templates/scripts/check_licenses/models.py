#!/usr/bin/env python3
"""
Data models for check_licenses module
"""

from dataclasses import dataclass
from typing import List, Dict, Optional


@dataclass
class LicenseInfo:
    """License information structure"""
    name: str
    spdx_id: Optional[str]
    category: str  # permissive, copyleft, proprietary, unknown
    risk_level: str  # low, medium, high, critical
    restrictions: List[str]
    compatibility: Dict[str, bool]  # Compatibility with other licenses


@dataclass
class PackageLicense:
    """Package license information"""
    package: str
    version: str
    license: str
    license_info: Optional[LicenseInfo]
    source: str  # requirements.txt, package.json, etc.
    status: str  # compliant, non-compliant, needs-review