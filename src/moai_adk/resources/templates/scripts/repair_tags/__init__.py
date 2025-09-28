#!/usr/bin/env python3
# @TASK:TAG-REPAIR-011
"""
MoAI-ADK TAG Auto-Repair System

Modularized TAG repair system following TRUST principles.
Each module has a single responsibility and ≤ 50 LOC.
"""

from .core import TagRepairer
from .scanner import TagScanner
from .analyzer import TagAnalyzer
from .generator import RepairGenerator
from .templates import TemplateCreator
from .updater import IndexUpdater

__all__ = [
    "TagRepairer",
    "TagScanner",
    "TagAnalyzer",
    "RepairGenerator",
    "TemplateCreator",
    "IndexUpdater",
]