#!/usr/bin/env python3
"""
MoAI-ADK License Compliance Checker v0.1.12 - Modularized
Project dependency license scanning and compatibility verification
"""

from .main import main
from .models import LicenseInfo, PackageLicense
from .checker import LicenseChecker

__all__ = ["main", "LicenseInfo", "PackageLicense", "LicenseChecker"]


if __name__ == "__main__":
    main()