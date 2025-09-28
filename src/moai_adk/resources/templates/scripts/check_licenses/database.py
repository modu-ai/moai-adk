#!/usr/bin/env python3
"""
License database for check_licenses module
"""

from typing import Dict
from .models import LicenseInfo


def init_license_database() -> Dict[str, LicenseInfo]:
    """Initialize license information database"""
    return {
        # Permissive Licenses (Low Risk)
        "MIT": LicenseInfo(
            name="MIT License", spdx_id="MIT", category="permissive",
            risk_level="low", restrictions=["include-copyright"],
            compatibility={"MIT": True, "Apache-2.0": True, "BSD": True}
        ),
        "Apache-2.0": LicenseInfo(
            name="Apache License 2.0", spdx_id="Apache-2.0", category="permissive",
            risk_level="low", restrictions=["include-copyright", "include-license"],
            compatibility={"MIT": True, "Apache-2.0": True, "BSD": True}
        ),
        "BSD-3-Clause": LicenseInfo(
            name="BSD 3-Clause License", spdx_id="BSD-3-Clause", category="permissive",
            risk_level="low", restrictions=["include-copyright"],
            compatibility={"MIT": True, "Apache-2.0": True, "BSD": True}
        ),
        # Copyleft Licenses (Higher Risk)
        "GPL-3.0": LicenseInfo(
            name="GNU General Public License v3.0", spdx_id="GPL-3.0",
            category="copyleft", risk_level="high",
            restrictions=["disclose-source", "include-copyright", "same-license"],
            compatibility={"MIT": False, "Apache-2.0": False, "GPL-3.0": True}
        ),
        "AGPL-3.0": LicenseInfo(
            name="GNU Affero General Public License v3.0", spdx_id="AGPL-3.0",
            category="copyleft", risk_level="critical",
            restrictions=["disclose-source", "include-copyright", "same-license", "network-use"],
            compatibility={"MIT": False, "Apache-2.0": False, "AGPL-3.0": True}
        ),
        # Unknown
        "UNKNOWN": LicenseInfo(
            name="Unknown License", spdx_id=None, category="unknown",
            risk_level="medium", restrictions=["review-required"],
            compatibility={}
        )
    }