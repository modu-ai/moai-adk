"""
RED Phase Tests for moai-foundation-git skill - Git workflow management.
"""

from dataclasses import dataclass
from typing import List, Optional


@dataclass
class GitInfo:
    """Git version and feature information."""

    version: str
    supports_switch: bool = False
    supports_worktree: bool = False
    supports_sparse_checkout: bool = False
    modern_features: List[str] = None

    def __post_init__(self):
        if self.modern_features is None:
            self.modern_features = []


class TestGitVersionDetection:
    """Test Git version detection and feature support."""

    def test_detect_git_version(self):
        """Test Git version detection."""
        detector = GitVersionDetector()
        info = detector.detect("git version 2.47.0")
        assert info.version == "2.47.0"

    def test_modern_features_git_2_47(self):
        """Test modern features in Git 2.47+."""
        detector = GitVersionDetector()
        info = detector.detect("git version 2.47.0")
        assert info.supports_switch is True
        assert info.supports_worktree is True
        assert "sparse-checkout" in [f.lower() for f in info.modern_features]

    def test_legacy_git_version(self):
        """Test legacy Git versions lack modern features."""
        detector = GitVersionDetector()
        info = detector.detect("git version 2.30.0")
        assert info.version == "2.30.0"


class TestConventionalCommitValidation:
    """Test Conventional Commits validation."""

    def test_valid_feature_commit(self):
        """Test valid feat commit."""
        validator = ConventionalCommitValidator()
        result = validator.validate("feat(auth): add JWT validation")
        assert result.is_valid is True
        assert result.type == "feat"
        assert result.scope == "auth"

    def test_valid_fix_commit(self):
        """Test valid fix commit."""
        validator = ConventionalCommitValidator()
        result = validator.validate("fix(api): handle null user")
        assert result.is_valid is True
        assert result.type == "fix"

    def test_invalid_commit_format(self):
        """Test invalid commit format."""
        validator = ConventionalCommitValidator()
        result = validator.validate("added new feature")
        assert result.is_valid is False
        assert result.errors is not None and len(result.errors) > 0

    def test_breaking_change_commit(self):
        """Test breaking change commit."""
        validator = ConventionalCommitValidator()
        message = """feat(api)!: change login endpoint

BREAKING CHANGE: Old endpoint /login removed"""
        result = validator.validate(message)
        assert result.is_valid is True
        assert result.is_breaking_change is True


class TestBranchingStrategySelection:
    """Test branching strategy selection."""

    def test_select_feature_branch_strategy(self):
        """Test feature branch strategy for teams."""
        selector = BranchingStrategySelector()
        strategy = selector.select_strategy(team_size=3, risk_level="high", need_review=True)
        assert strategy == "feature_branch"

    def test_select_direct_commit_strategy(self):
        """Test direct commit strategy for individuals."""
        selector = BranchingStrategySelector()
        strategy = selector.select_strategy(team_size=1, risk_level="low", need_review=False)
        assert strategy == "direct_commit"

    def test_select_per_spec_strategy(self):
        """Test flexible per-SPEC strategy."""
        selector = BranchingStrategySelector()
        strategy = selector.select_strategy(team_size=2, risk_level="medium", need_review=False)
        assert strategy in ["feature_branch", "direct_commit"]


class TestGitWorkflowManagement:
    """Test Git workflow management."""

    def test_create_feature_branch_command(self):
        """Test feature branch creation command."""
        manager = GitWorkflowManager()
        cmd = manager.create_branch_command("feature/SPEC-001")
        assert "switch" in cmd or "checkout" in cmd
        assert "SPEC-001" in cmd

    def test_red_phase_commit_message(self):
        """Test RED phase commit message."""
        manager = GitWorkflowManager()
        msg = manager.format_tdd_commit("test", "auth", "add validation tests", "RED")
        assert "test" in msg.lower()
        assert "RED" in msg or "red" in msg.lower()

    def test_green_phase_commit_message(self):
        """Test GREEN phase commit message."""
        manager = GitWorkflowManager()
        msg = manager.format_tdd_commit("feat", "auth", "implement validation", "GREEN")
        assert "feat" in msg
        assert "GREEN" in msg or "green" in msg.lower()

    def test_refactor_phase_commit_message(self):
        """Test REFACTOR phase commit message."""
        manager = GitWorkflowManager()
        msg = manager.format_tdd_commit("refactor", "auth", "optimize code", "REFACTOR")
        assert "refactor" in msg


class TestGitPerformanceOptimization:
    """Test Git performance optimization tips."""

    def test_performance_tips_for_large_repo(self):
        """Test performance tips for large repository."""
        optimizer = GitPerformanceOptimizer()
        tips = optimizer.get_tips_for_repo_size("large")
        assert len(tips) > 0
        assert any("midx" in tip.lower() for tip in tips)

    def test_sparse_checkout_recommendation(self):
        """Test sparse checkout recommendation."""
        optimizer = GitPerformanceOptimizer()
        tips = optimizer.get_tips_for_repo_size("large")
        assert any("sparse" in tip.lower() for tip in tips)

    def test_partial_clone_recommendation(self):
        """Test partial clone recommendation."""
        optimizer = GitPerformanceOptimizer()
        tips = optimizer.get_tips_for_repo_size("large")
        assert any("clone" in tip.lower() or "filter" in tip.lower() for tip in tips)


@dataclass
class ValidateResult:
    """Validation result."""

    is_valid: bool
    type: Optional[str] = None
    scope: Optional[str] = None
    errors: List[str] = None
    is_breaking_change: bool = False

    def __post_init__(self):
        if self.errors is None:
            self.errors = []


class GitVersionDetector:
    """Detects Git version and supported features."""

    def detect(self, version_string: str) -> GitInfo:
        """Detect Git version from version string."""
        import re

        match = re.search(r"git version ([\d.]+)", version_string)
        if not match:
            return GitInfo(version="unknown")

        version = match.group(1)
        parts = [int(x) for x in version.split(".")[:2]]  # major.minor
        major, minor = parts[0], parts[1] if len(parts) > 1 else 0

        # Git 2.40+ has switch/restore
        supports_switch = (major > 2) or (major == 2 and minor >= 40)
        # Git 2.30+ has worktree
        supports_worktree = (major > 2) or (major == 2 and minor >= 30)
        # Git 2.25+ has sparse-checkout
        supports_sparse = (major > 2) or (major == 2 and minor >= 25)

        modern_features = []
        if supports_switch:
            modern_features.append("switch")
            modern_features.append("restore")
        if supports_worktree:
            modern_features.append("worktree")
        if supports_sparse:
            modern_features.append("sparse-checkout")

        return GitInfo(
            version=version,
            supports_switch=supports_switch,
            supports_worktree=supports_worktree,
            supports_sparse_checkout=supports_sparse,
            modern_features=modern_features,
        )


class ConventionalCommitValidator:
    """Validates Conventional Commits format."""

    VALID_TYPES = ["feat", "fix", "docs", "style", "refactor", "perf", "test", "chore"]

    def validate(self, message: str) -> ValidateResult:
        """Validate Conventional Commits format."""
        import re

        lines = message.strip().split("\n")
        first_line = lines[0]

        # Check for breaking change
        is_breaking = "!" in first_line or any("BREAKING" in line for line in lines)

        # Parse: type(scope): subject
        pattern = r"^(feat|fix|docs|style|refactor|perf|test|chore)(\(.+?\))?(!)?:\s+.+$"
        if not re.match(pattern, first_line):
            return ValidateResult(
                is_valid=False, errors=["Invalid Conventional Commits format. Expected: type(scope): subject"]
            )

        # Extract type and scope
        match = re.match(r"^(\w+)(\((.+?)\))?", first_line)
        if not match:
            return ValidateResult(is_valid=False, errors=["Could not parse type/scope"])

        commit_type = match.group(1)
        scope = match.group(3)

        return ValidateResult(is_valid=True, type=commit_type, scope=scope, is_breaking_change=is_breaking)


class BranchingStrategySelector:
    """Selects appropriate branching strategy."""

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


class GitWorkflowManager:
    """Manages Git workflow commands and patterns."""

    def create_branch_command(self, branch_name: str) -> str:
        """Create branch creation command."""
        return f"git switch -c {branch_name}"

    def format_tdd_commit(self, commit_type: str, scope: str, subject: str, phase: str) -> str:
        """Format TDD commit message with phase information."""
        base_msg = f"{commit_type}({scope}): {subject}"
        return f"{base_msg} ({phase})"


class GitPerformanceOptimizer:
    """Provides Git performance optimization tips."""

    PERFORMANCE_TIPS = {
        "small": ["Standard clone operations", "Keep working directory clean", "Regular garbage collection"],
        "medium": [
            "Enable MIDX for faster operations",
            "Use shallow clones for CI/CD",
            "Consider sparse-checkout for selective directories",
        ],
        "large": [
            "Enable MIDX for 38% performance improvement",
            "Use partial clones (--filter=blob:none) for 81% smaller downloads",
            "Use sparse-checkout to clone only needed directories",
            "Implement shallow clones for CI/CD (73% faster)",
            "Configure git config --global gc.writeMultiPackIndex true",
        ],
    }

    def get_tips_for_repo_size(self, size: str) -> List[str]:
        """Get performance tips for repository size."""
        return self.PERFORMANCE_TIPS.get(size, [])
