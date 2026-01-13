"""
Git Workflow Management Implementation (moai-foundation-git).

Provides:
- GitVersionDetector: Detect Git version and modern features
- ConventionalCommitValidator: Validate Conventional Commits format
- BranchingStrategySelector: Select appropriate branching strategy
- GitWorkflowManager: Manage Git workflow operations
- GitPerformanceOptimizer: Performance optimization recommendations

Reference: moai-foundation-git for enterprise Git workflows
"""

import re
from dataclasses import dataclass, field
from typing import Dict, List, Optional


@dataclass
class GitInfo:
    """Git version and feature information."""

    version: str
    supports_switch: bool = False
    supports_worktree: bool = False
    supports_sparse_checkout: bool = False
    modern_features: List[str] = field(default_factory=list)


@dataclass
class ValidateResult:
    """Conventional Commits validation result."""

    is_valid: bool
    commit_type: Optional[str] = None
    scope: Optional[str] = None
    subject: Optional[str] = None
    errors: List[str] = field(default_factory=list)
    is_breaking_change: bool = False
    suggestions: List[str] = field(default_factory=list)


@dataclass
class TDDCommitPhase:
    """TDD commit phase information."""

    phase_name: str  # "RED", "GREEN", or "REFACTOR"
    commit_type: str  # "test", "feat", "refactor"
    description: str
    test_status: str  # "failing", "passing", or "optimizing"


class GitVersionDetector:
    """Detects Git version and supported modern features."""

    MODERN_FEATURES = {
        "switch": (2, 40),  # git switch - alternative to checkout
        "restore": (2, 40),  # git restore - alternative to checkout
        "worktree": (2, 30),  # git worktree - parallel development
        "sparse-checkout": (2, 25),  # git sparse-checkout - monorepo optimization
        "partial-clone": (2, 24),  # --filter for partial clones
    }

    def detect(self, version_string: str) -> GitInfo:
        """Detect Git version from version string."""
        match = re.search(r"git version ([\d.]+)", version_string)
        if not match:
            return GitInfo(version="unknown")

        version = match.group(1)
        parts = [int(x) for x in version.split(".")[:2]]  # major.minor
        major = parts[0]
        minor = parts[1] if len(parts) > 1 else 0

        # Determine supported features
        modern_features = []
        supports_switch = False
        supports_worktree = False
        supports_sparse = False

        for feature, (req_major, req_minor) in self.MODERN_FEATURES.items():
            if major > req_major or (major == req_major and minor >= req_minor):
                modern_features.append(feature)
                if feature == "switch":
                    supports_switch = True
                elif feature == "worktree":
                    supports_worktree = True
                elif feature == "sparse-checkout":
                    supports_sparse = True

        return GitInfo(
            version=version,
            supports_switch=supports_switch,
            supports_worktree=supports_worktree,
            supports_sparse_checkout=supports_sparse,
            modern_features=modern_features,
        )

    def get_recommended_commands(self, git_info: GitInfo) -> Dict[str, str]:
        """Get recommended commands based on Git version."""
        commands = {}

        if git_info.supports_switch:
            commands["create_branch"] = "git switch -c branch-name"
            commands["switch_branch"] = "git switch branch-name"
        else:
            commands["create_branch"] = "git checkout -b branch-name"
            commands["switch_branch"] = "git checkout branch-name"

        if git_info.supports_sparse_checkout:
            commands["sparse_clone"] = "git clone --sparse repo.git"

        if git_info.supports_worktree:
            commands["parallel_work"] = "git worktree add ../dir branch-name"

        return commands


class ConventionalCommitValidator:
    """Validates Conventional Commits 2025 format."""

    VALID_TYPES = {
        "feat": "New feature",
        "fix": "Bug fix",
        "docs": "Documentation",
        "style": "Code style (no logic change)",
        "refactor": "Code refactoring",
        "perf": "Performance improvement",
        "test": "Test updates",
        "chore": "Build/dependencies/tooling",
    }

    def validate(self, message: str) -> ValidateResult:
        """Validate Conventional Commits format."""
        lines = message.strip().split("\n")
        first_line = lines[0]

        # Check for breaking change indicator
        is_breaking = "!" in first_line
        if any("BREAKING CHANGE:" in line for line in lines):
            is_breaking = True

        # Validate format: type(scope): subject
        pattern = r"^(feat|fix|docs|style|refactor|perf|test|chore)(\(.+?\))?(!)?:\s+.+$"
        match = re.match(pattern, first_line)

        if not match:
            return ValidateResult(
                is_valid=False,
                errors=[
                    "Invalid Conventional Commits format",
                    "Expected: type(scope): subject",
                    f"Got: {first_line[:50]}...",
                ],
                suggestions=[
                    "Use one of: feat, fix, docs, style, refactor, perf, test, chore",
                    "Add scope in parentheses: feat(auth): ...",
                    "Use lowercase with clear description",
                ],
            )

        # Extract components
        type_match = re.match(r"^(\w+)(\((.+?)\))?(!)?", first_line)
        if not type_match:
            return ValidateResult(is_valid=False, errors=["Could not parse commit message"])

        commit_type = type_match.group(1)
        scope = type_match.group(3) if type_match.group(3) else None

        # Extract subject (everything after ': ')
        subject_match = re.search(r":\s+(.+)$", first_line)
        subject = subject_match.group(1) if subject_match else None

        return ValidateResult(
            is_valid=True,
            commit_type=commit_type,
            scope=scope,
            subject=subject,
            is_breaking_change=is_breaking,
        )

    def validate_batch(self, messages: List[str]) -> Dict[str, ValidateResult]:
        """Validate multiple commit messages."""
        return {msg: self.validate(msg) for msg in messages}


class BranchingStrategySelector:
    """Selects appropriate branching strategy for project."""

    STRATEGIES = {
        "feature_branch": {
            "name": "Feature Branch",
            "description": "Create feature branches, use PRs for review",
            "best_for": ["Teams", "High-risk changes", "Code review required"],
            "commands": [
                "git switch -c feature/SPEC-001",
                "gh pr create --base develop",
                "gh pr merge --squash --auto",
            ],
        },
        "direct_commit": {
            "name": "Direct Commit",
            "description": "Commit directly to main/develop branch",
            "best_for": ["Solo developers", "Low-risk changes", "Rapid prototyping"],
            "commands": [
                "git switch develop",
                "git commit -m 'feat: ...'",
                "git push origin develop",
            ],
        },
        "per_spec": {
            "name": "Per-SPEC Choice",
            "description": "Decide strategy for each SPEC",
            "best_for": ["Hybrid workflows", "Mixed team projects"],
            "commands": ["Flexible based on SPEC requirements"],
        },
    }

    def select_strategy(self, team_size: int, risk_level: str, need_review: bool) -> str:
        """Select branching strategy based on parameters."""
        # Feature branch for teams, risky changes, or review requirements
        if team_size > 1 or risk_level == "high" or need_review:
            return "feature_branch"

        # Direct commit for solo developers with low risk
        if team_size == 1 and risk_level in ["low", "medium"] and not need_review:
            return "direct_commit"

        # Default to flexible per-SPEC choice
        return "per_spec"

    def get_strategy_details(self, strategy: str) -> Dict:
        """Get details about a strategy."""
        return self.STRATEGIES.get(strategy, {})


class GitWorkflowManager:
    """Manages Git workflow operations for SPEC-first TDD."""

    TDD_PHASES = {
        "RED": {
            "commit_type": "test",
            "description": "Write failing tests",
            "tests_status": "failing",
        },
        "GREEN": {
            "commit_type": "feat",
            "description": "Implement to pass tests",
            "tests_status": "passing",
        },
        "REFACTOR": {
            "commit_type": "refactor",
            "description": "Optimize implementation",
            "tests_status": "passing",
        },
    }

    def create_branch_command(self, branch_name: str, use_modern: bool = True) -> str:
        """Create branch creation command."""
        if use_modern:
            return f"git switch -c {branch_name}"
        return f"git checkout -b {branch_name}"

    def format_tdd_commit(self, commit_type: str, scope: str, subject: str, phase: str) -> str:
        """Format TDD phase commit message."""
        base_msg = f"{commit_type}({scope}): {subject}"

        # Add phase indicator
        phase_indicators = {
            "RED": "(RED phase)",
            "GREEN": "(GREEN phase)",
            "REFACTOR": "(REFACTOR phase)",
        }

        phase_indicator = phase_indicators.get(phase, f"({phase})")
        return f"{base_msg} {phase_indicator}"

    def get_workflow_commands(self, strategy: str, spec_id: str) -> List[str]:
        """Get workflow commands for strategy."""
        commands = []

        if strategy == "feature_branch":
            commands = [
                f"git switch -c feature/{spec_id}",
                "# RED phase: git commit -m 'test(...): add tests'",
                "# GREEN phase: git commit -m 'feat(...): implement'",
                "# REFACTOR phase: git commit -m 'refactor(...): optimize'",
                "gh pr create --base develop --generate-description",
                f"gh pr merge {spec_id} --auto --squash",
            ]
        elif strategy == "direct_commit":
            commands = [
                "git switch develop",
                "# RED phase: git commit -m 'test(...): add tests'",
                "# GREEN phase: git commit -m 'feat(...): implement'",
                "# REFACTOR phase: git commit -m 'refactor(...): optimize'",
                "git push origin develop",
            ]

        return commands


class GitPerformanceOptimizer:
    """Provides Git performance optimization recommendations."""

    PERFORMANCE_TIPS = {
        "small": [
            "Standard clone operations sufficient",
            "Keep working directory clean",
            "Regular garbage collection",
        ],
        "medium": [
            "Enable MIDX for faster operations",
            "Use shallow clones for CI/CD",
            "Consider sparse-checkout for selective directories",
        ],
        "large": [
            "Enable MIDX: git config --global gc.writeMultiPackIndex true",
            "Use partial clones: git clone --filter=blob:none (81% smaller)",
            "Use sparse-checkout for monorepo (70% faster clone)",
            "Implement shallow clones for CI/CD: git clone --depth 100",
            "Run: git repack -ad --write-midx for 38% performance improvement",
        ],
    }

    OPTIMIZATION_CONFIGS = {
        "midx": {
            "name": "Multi-Pack Indexes (MIDX)",
            "benefit": "38% faster repository operations",
            "command": "git config --global gc.writeMultiPackIndex true",
        },
        "partial_clone": {
            "name": "Partial Clones (Blob-less)",
            "benefit": "81% smaller downloads",
            "command": "git clone --filter=blob:none https://repo.git",
        },
        "sparse_checkout": {
            "name": "Sparse Checkout",
            "benefit": "70% faster clone for monorepos",
            "command": "git clone --sparse && git sparse-checkout init --cone",
        },
        "shallow_clone": {
            "name": "Shallow Clones",
            "benefit": "73% smaller initial download",
            "command": "git clone --depth 100 https://repo.git",
        },
    }

    def get_tips_for_repo_size(self, size: str) -> List[str]:
        """Get performance tips for repository size."""
        return self.PERFORMANCE_TIPS.get(size, [])

    def get_optimization_configs(self) -> Dict:
        """Get all optimization configurations."""
        return self.OPTIMIZATION_CONFIGS

    def get_recommended_optimizations(self, repo_size: str) -> List[str]:
        """Get recommended optimizations for repo size."""
        if repo_size == "large":
            return ["midx", "partial_clone", "sparse_checkout"]
        elif repo_size == "medium":
            return ["midx", "shallow_clone"]
        return []


# Export public API
__all__ = [
    "GitInfo",
    "ValidateResult",
    "TDDCommitPhase",
    "GitVersionDetector",
    "ConventionalCommitValidator",
    "BranchingStrategySelector",
    "GitWorkflowManager",
    "GitPerformanceOptimizer",
]
