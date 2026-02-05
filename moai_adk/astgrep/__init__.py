# AST-grep analyzer package
"""AST-grep based code analyzer for MoAI-ADK.

This package provides AST-based code analysis capabilities using ast-grep (sg CLI).
It supports security scanning, pattern searching, and code transformation.

Main Components:
    - ASTMatch: Represents a match found by AST-grep
    - ScanConfig: Configuration for scanning operations
    - ScanResult: Result of scanning a single file
    - ProjectScanResult: Result of scanning a project
    - ReplaceResult: Result of pattern replacement operations
    - Rule: AST-grep rule definition
    - RuleLoader: Loads rules from files
    - MoAIASTGrepAnalyzer: Main analyzer class
"""

from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer
from moai_adk.astgrep.models import (
    ASTMatch,
    ProjectScanResult,
    ReplaceResult,
    ScanConfig,
    ScanResult,
)
from moai_adk.astgrep.rules import Rule, RuleLoader

__all__ = [
    "ASTMatch",
    "MoAIASTGrepAnalyzer",
    "ProjectScanResult",
    "ReplaceResult",
    "Rule",
    "RuleLoader",
    "ScanConfig",
    "ScanResult",
]
