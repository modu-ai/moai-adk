"""
MoAI Foundation module - Core foundation-level implementations.
Includes: EARS methodology, programming language ecosystem, Git workflows.
"""

from .ears import EARSAnalyzer, EARSParser, EARSValidator
from .git import (
    BranchingStrategySelector,
    ConventionalCommitValidator,
    GitInfo,
    GitPerformanceOptimizer,
    GitVersionDetector,
    GitWorkflowManager,
    TDDCommitPhase,
    ValidateResult,
)
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
    # Git Workflow
    "GitVersionDetector",
    "ConventionalCommitValidator",
    "BranchingStrategySelector",
    "GitWorkflowManager",
    "GitPerformanceOptimizer",
    # Data structures
    "LanguageInfo",
    "Pattern",
    "TestingStrategy",
    "GitInfo",
    "ValidateResult",
    "TDDCommitPhase",
]
