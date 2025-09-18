"""
Package Optimization System

@REQ:OPT-CORE-001 - 패키지 크기 최적화 시스템
@DESIGN:PKG-ARCH-001 - 클린 아키텍처 기반 패키지 최적화
"""

__version__ = "0.1.0"
__author__ = "MoAI-ADK"

from .core.package_optimizer import PackageOptimizer
from .core.duplicate_remover import DuplicateRemover
from .core.metrics_tracker import MetricsTracker

__all__ = [
    "PackageOptimizer",
    "DuplicateRemover",
    "MetricsTracker"
]