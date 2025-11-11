# # REMOVED_ORPHAN_CODE:TRUST-002:VALIDATOR | SPEC: SPEC-TRUST-001/spec.md
"""Base validator class and validation result"""

from dataclasses import dataclass
from typing import Any


@dataclass
class ValidationResult:
    """Validation result data class"""

    passed: bool
    message: str
    details: str = ""
    metadata: dict[str, Any] | None = None

    def __post_init__(self):
        if self.metadata is None:
            self.metadata = {}
