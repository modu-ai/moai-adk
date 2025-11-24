"""
MoAI Foundation module - Core foundation-level implementations.
Includes: EARS methodology, programming language ecosystem, TRUST principles.
"""

from .ears import EARSParser, EARSValidator, EARSAnalyzer
from .langs import (
    LanguageVersionManager,
    FrameworkRecommender,
    PatternAnalyzer,
    AntiPatternDetector,
    EcosystemAnalyzer,
    PerformanceOptimizer,
    TestingStrategyAdvisor,
    LanguageInfo,
    Pattern,
    TestingStrategy,
)

__all__ = [
    # EARS
    'EARSParser',
    'EARSValidator',
    'EARSAnalyzer',
    # Language Ecosystem
    'LanguageVersionManager',
    'FrameworkRecommender',
    'PatternAnalyzer',
    'AntiPatternDetector',
    'EcosystemAnalyzer',
    'PerformanceOptimizer',
    'TestingStrategyAdvisor',
    # Data structures
    'LanguageInfo',
    'Pattern',
    'TestingStrategy',
]
