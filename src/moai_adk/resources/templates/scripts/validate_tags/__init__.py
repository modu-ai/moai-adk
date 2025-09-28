#!/usr/bin/env python3
"""
MoAI-ADK Tag System Validator v0.1.12 - Modularized
16-Core @TAG integrity validation and traceability matrix verification
"""

from .main import main
from .parser import TagReference, TagHealthReport
from .validator import TagValidator

__all__ = ["main", "TagReference", "TagHealthReport", "TagValidator"]


if __name__ == "__main__":
    main()