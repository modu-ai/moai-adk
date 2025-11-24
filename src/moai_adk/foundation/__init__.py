"""
MoAI Foundation module - Core foundation-level implementations.
Includes: EARS methodology, programming language ecosystem, Git workflows.
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
from .git import (
    GitVersionDetector,
    ConventionalCommitValidator,
    BranchingStrategySelector,
    GitWorkflowManager,
    GitPerformanceOptimizer,
    GitInfo,
    ValidateResult,
    TDDCommitPhase,
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
    # Git Workflow
    'GitVersionDetector',
    'ConventionalCommitValidator',
    'BranchingStrategySelector',
    'GitWorkflowManager',
    'GitPerformanceOptimizer',
    # Data structures
    'LanguageInfo',
    'Pattern',
    'TestingStrategy',
    'GitInfo',
    'ValidateResult',
    'TDDCommitPhase',
]
