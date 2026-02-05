"""
MoAI Foundation module - Core foundation-level implementations.
Includes: EARS methodology, programming language ecosystem, testing framework.
"""

from .ears import EARSAnalyzer, EARSParser, EARSValidator
from .langs import (
    AntiPatternDetector,
    EcosystemAnalyzer,
    FrameworkRecommender,
    LanguageInfo,
    LanguageVersionManager,
    Pattern,
    PatternAnalyzer,
    PerformanceOptimizer,
    TestingStrategy,
    TestingStrategyAdvisor,
)

__all__ = [
    # EARS
    "EARSParser",
    "EARSValidator",
    "EARSAnalyzer",
    # Language Ecosystem
    "LanguageVersionManager",
    "FrameworkRecommender",
    "PatternAnalyzer",
    "AntiPatternDetector",
    "EcosystemAnalyzer",
    "PerformanceOptimizer",
    "TestingStrategyAdvisor",
    # Data structures
    "LanguageInfo",
    "Pattern",
    "TestingStrategy",
]
