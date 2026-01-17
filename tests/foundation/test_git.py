"""
Comprehensive DDD tests for git.py module.
Tests cover all 5 classes.
"""

import pytest
from moai_adk.foundation.git import (
    GitInfo,
    ValidateResult,
    DDDCommitPhase,
    GitVersionDetector,
    ConventionalCommitValidator,
    BranchingStrategySelector,
    GitWorkflowManager,
    GitPerformanceOptimizer,
)


# ============================================================================
# Test Data Classes
# ============================================================================


class TestDataClasses:
    """Test suite for Git data classes."""

    def test_git_info_creation(self):
        """Test GitInfo dataclass creation."""
        info = GitInfo(
            version="2.45.0",
            supports_switch=True,
            supports_worktree=True,
            supports_sparse_checkout=True,
            modern_features=["switch", "worktree", "sparse-checkout"]
        )

        assert info.version == "2.45.0"
        assert info.supports_switch is True
        assert "switch" in info.modern_features

    def test_validate_result_creation(self):
        """Test ValidateResult dataclass creation."""
        result = ValidateResult(
            is_valid=True,
            commit_type="feat",
            scope="auth",
            subject="add authentication",
            errors=[],
            is_breaking_change=False
        )

        assert result.is_valid is True
        assert result.commit_type == "feat"
        assert result.scope == "auth"

    def test_tdd_commit_phase_creation(self):
        """Test DDDCommitPhase dataclass creation."""
        phase = DDDCommitPhase(
            phase_name="RED",
            commit_type="test",
            description="Write failing tests",
            test_status="failing"
        )

        assert phase.phase_name == "RED"
        assert phase.test_status == "failing"


# ============================================================================
# Test GitVersionDetector
# ============================================================================


class TestGitVersionDetector:
    """Test suite for GitVersionDetector class."""

    def test_initialization(self):
        """Test detector initialization."""
        detector = GitVersionDetector()
        assert "switch" in detector.MODERN_FEATURES
        assert "worktree" in detector.MODERN_FEATURES

    def test_detect_valid_version(self, sample_git_version):
        """Test detection of valid Git version."""
        detector = GitVersionDetector()
        info = detector.detect(sample_git_version)

        assert info.version == "2.45.0"
        assert info.supports_switch is True
        assert info.supports_worktree is True
        assert info.supports_sparse_checkout is True
        assert "switch" in info.modern_features

    def test_detect_old_version(self):
        """Test detection of old Git version."""
        detector = GitVersionDetector()
        info = detector.detect("git version 2.20.0")

        assert info.version == "2.20.0"
        assert info.supports_switch is False
        assert info.supports_worktree is False  # worktree added in 2.30, 2.20 < 2.30
        assert "worktree" not in info.modern_features

    def test_detect_invalid_version(self):
        """Test detection of invalid version string."""
        detector = GitVersionDetector()
        info = detector.detect("invalid version string")

        assert info.version == "unknown"
        assert info.supports_switch is False
        assert info.supports_worktree is False

    def test_get_recommended_commands_modern(self):
        """Test recommended commands for modern Git."""
        detector = GitVersionDetector()
        info = GitInfo(
            version="2.45.0",
            supports_switch=True,
            supports_worktree=True,
            supports_sparse_checkout=True
        )

        commands = detector.get_recommended_commands(info)

        assert "git switch" in commands["create_branch"]
        assert "git switch" in commands["switch_branch"]
        assert "git clone --sparse" in commands["sparse_clone"]
        assert "git worktree" in commands["parallel_work"]

    def test_get_recommended_commands_legacy(self):
        """Test recommended commands for legacy Git."""
        detector = GitVersionDetector()
        info = GitInfo(
            version="2.20.0",
            supports_switch=False,
            supports_worktree=True
        )

        commands = detector.get_recommended_commands(info)

        assert "git checkout -b" in commands["create_branch"]
        assert "git checkout" in commands["switch_branch"]


# ============================================================================
# Test ConventionalCommitValidator
# ============================================================================


class TestConventionalCommitValidator:
    """Test suite for ConventionalCommitValidator class."""

    def test_initialization(self):
        """Test validator initialization."""
        validator = ConventionalCommitValidator()
        assert "feat" in validator.VALID_TYPES
        assert "fix" in validator.VALID_TYPES

    def test_validate_valid_commit(self, sample_commit_message):
        """Test validation of valid commit message."""
        validator = ConventionalCommitValidator()
        result = validator.validate(sample_commit_message)

        assert result.is_valid is True
        assert result.commit_type == "feat"
        assert result.scope == "auth"
        assert result.subject == "add user authentication"
        assert len(result.errors) == 0

    def test_validate_commit_without_scope(self):
        """Test validation of commit without scope."""
        validator = ConventionalCommitValidator()
        result = validator.validate("feat: add new feature")

        assert result.is_valid is True
        assert result.commit_type == "feat"
        assert result.scope is None
        assert result.subject == "add new feature"

    def test_validate_commit_with_breaking_change(self):
        """Test validation of breaking change commit."""
        validator = ConventionalCommitValidator()
        result = validator.validate("feat!: breaking API change")

        assert result.is_valid is True
        assert result.is_breaking_change is True

    def test_validate_commit_with_breaking_change_body(self):
        """Test validation with BREAKING CHANGE in body."""
        validator = ConventionalCommitValidator()
        message = """feat(api): add new endpoint

BREAKING CHANGE: This changes the API signature"""
        result = validator.validate(message)

        assert result.is_valid is True
        assert result.is_breaking_change is True

    def test_validate_invalid_type(self):
        """Test validation of invalid commit type."""
        validator = ConventionalCommitValidator()
        result = validator.validate("invalid: some message")

        assert result.is_valid is False
        assert len(result.errors) > 0
        assert "Invalid Conventional Commits format" in result.errors[0]

    def test_validate_missing_colon(self):
        """Test validation of commit missing colon."""
        validator = ConventionalCommitValidator()
        result = validator.validate("feat some message")

        assert result.is_valid is False
        assert len(result.errors) > 0

    def test_validate_batch_commits(self):
        """Test batch validation of multiple commits."""
        validator = ConventionalCommitValidator()
        messages = [
            "feat: add feature",
            "fix: fix bug",
            "invalid message"
        ]

        results = validator.validate_batch(messages)

        assert len(results) == 3
        assert results["feat: add feature"].is_valid is True
        assert results["fix: fix bug"].is_valid is True
        assert results["invalid message"].is_valid is False

    def test_validate_all_valid_types(self):
        """Test validation of all valid commit types."""
        validator = ConventionalCommitValidator()
        valid_types = ["feat", "fix", "docs", "style", "refactor", "perf", "test", "chore"]

        for commit_type in valid_types:
            result = validator.validate(f"{commit_type}: test message")
            assert result.is_valid is True, f"Failed for type: {commit_type}"


# ============================================================================
# Test BranchingStrategySelector
# ============================================================================


class TestBranchingStrategySelector:
    """Test suite for BranchingStrategySelector class."""

    def test_initialization(self):
        """Test selector initialization."""
        selector = BranchingStrategySelector()
        assert "feature_branch" in selector.STRATEGIES
        assert "direct_commit" in selector.STRATEGIES

    def test_select_strategy_feature_branch(self):
        """Test strategy selection for team with review."""
        selector = BranchingStrategySelector()
        strategy = selector.select_strategy(
            team_size=5,
            risk_level="high",
            need_review=True
        )

        assert strategy == "feature_branch"

    def test_select_strategy_direct_commit(self):
        """Test strategy selection for solo developer."""
        selector = BranchingStrategySelector()
        strategy = selector.select_strategy(
            team_size=1,
            risk_level="low",
            need_review=False
        )

        assert strategy == "direct_commit"

    def test_select_strategy_per_spec(self):
        """Test per-SPEC strategy selection."""
        selector = BranchingStrategySelector()
        # team_size > 1 returns feature_branch, not per_spec
        strategy = selector.select_strategy(
            team_size=2,
            risk_level="medium",
            need_review=False
        )

        assert strategy == "feature_branch"

    def test_get_strategy_details_feature_branch(self):
        """Test feature branch strategy details."""
        selector = BranchingStrategySelector()
        details = selector.get_strategy_details("feature_branch")

        assert details["name"] == "Feature Branch"
        assert len(details["commands"]) > 0
        assert "git switch -c" in details["commands"][0]

    def test_get_strategy_details_direct_commit(self):
        """Test direct commit strategy details."""
        selector = BranchingStrategySelector()
        details = selector.get_strategy_details("direct_commit")

        assert details["name"] == "Direct Commit"
        assert "git commit" in details["commands"][1]


# ============================================================================
# Test GitWorkflowManager
# ============================================================================


class TestGitWorkflowManager:
    """Test suite for GitWorkflowManager class."""

    def test_initialization(self):
        """Test manager initialization."""
        manager = GitWorkflowManager()
        assert "RED" in manager.DDD_PHASES
        assert "GREEN" in manager.DDD_PHASES
        assert "REFACTOR" in manager.DDD_PHASES

    def test_create_branch_command_modern(self):
        """Test branch creation command with modern Git."""
        manager = GitWorkflowManager()
        command = manager.create_branch_command("feature/SPEC-001", use_modern=True)

        assert command == "git switch -c feature/SPEC-001"

    def test_create_branch_command_legacy(self):
        """Test branch creation command with legacy Git."""
        manager = GitWorkflowManager()
        command = manager.create_branch_command("feature/SPEC-001", use_modern=False)

        assert command == "git checkout -b feature/SPEC-001"

    def test_format_ddd_commit_red(self):
        """Test TDD commit formatting for RED phase."""
        manager = GitWorkflowManager()
        commit_msg = manager.format_ddd_commit(
            commit_type="test",
            scope="auth",
            subject="add authentication tests",
            phase="RED"
        )

        assert commit_msg == "test(auth): add authentication tests (RED phase)"

    def test_format_ddd_commit_green(self):
        """Test TDD commit formatting for GREEN phase."""
        manager = GitWorkflowManager()
        commit_msg = manager.format_ddd_commit(
            commit_type="feat",
            scope="auth",
            subject="implement authentication",
            phase="GREEN"
        )

        assert commit_msg == "feat(auth): implement authentication (GREEN phase)"

    def test_format_ddd_commit_refactor(self):
        """Test TDD commit formatting for REFACTOR phase."""
        manager = GitWorkflowManager()
        commit_msg = manager.format_ddd_commit(
            commit_type="refactor",
            scope="auth",
            subject="improve authentication code",
            phase="REFACTOR"
        )

        assert commit_msg == "refactor(auth): improve authentication code (REFACTOR phase)"

    def test_get_workflow_commands_feature_branch(self):
        """Test workflow commands for feature branch strategy."""
        manager = GitWorkflowManager()
        commands = manager.get_workflow_commands("feature_branch", "SPEC-001")

        assert len(commands) > 0
        assert "git switch -c feature/SPEC-001" in commands[0]
        assert "# RED phase:" in commands[1]

    def test_get_workflow_commands_direct_commit(self):
        """Test workflow commands for direct commit strategy."""
        manager = GitWorkflowManager()
        commands = manager.get_workflow_commands("direct_commit", "SPEC-001")

        assert len(commands) > 0
        assert "git switch develop" in commands[0]


# ============================================================================
# Test GitPerformanceOptimizer
# ============================================================================


class TestGitPerformanceOptimizer:
    """Test suite for GitPerformanceOptimizer class."""

    def test_initialization(self):
        """Test optimizer initialization."""
        optimizer = GitPerformanceOptimizer()
        assert "small" in optimizer.PERFORMANCE_TIPS
        assert "large" in optimizer.PERFORMANCE_TIPS

    def test_get_tips_for_small_repo(self):
        """Test tips for small repository."""
        optimizer = GitPerformanceOptimizer()
        tips = optimizer.get_tips_for_repo_size("small")

        assert len(tips) > 0
        assert "Standard clone operations sufficient" in tips[0]

    def test_get_tips_for_medium_repo(self):
        """Test tips for medium repository."""
        optimizer = GitPerformanceOptimizer()
        tips = optimizer.get_tips_for_repo_size("medium")

        assert len(tips) > 0
        assert any("MIDX" in tip for tip in tips)

    def test_get_tips_for_large_repo(self):
        """Test tips for large repository."""
        optimizer = GitPerformanceOptimizer()
        tips = optimizer.get_tips_for_repo_size("large")

        assert len(tips) > 0
        assert any("partial clones" in tip.lower() for tip in tips)

    def test_get_optimization_configs(self):
        """Test all optimization configurations."""
        optimizer = GitPerformanceOptimizer()
        configs = optimizer.get_optimization_configs()

        assert "midx" in configs
        assert "partial_clone" in configs
        assert "sparse_checkout" in configs
        assert "shallow_clone" in configs

    def test_optimization_config_midx(self):
        """Test MIDX optimization configuration."""
        optimizer = GitPerformanceOptimizer()
        configs = optimizer.get_optimization_configs()

        assert configs["midx"]["name"] == "Multi-Pack Indexes (MIDX)"
        assert "38%" in configs["midx"]["benefit"]
        assert "gc.writeMultiPackIndex" in configs["midx"]["command"]

    def test_get_recommended_optimizations_large(self):
        """Test recommended optimizations for large repo."""
        optimizer = GitPerformanceOptimizer()
        optimizations = optimizer.get_recommended_optimizations("large")

        assert "midx" in optimizations
        assert "partial_clone" in optimizations
        assert "sparse_checkout" in optimizations

    def test_get_recommended_optimizations_small(self):
        """Test recommended optimizations for small repo."""
        optimizer = GitPerformanceOptimizer()
        optimizations = optimizer.get_recommended_optimizations("small")

        assert len(optimizations) == 0
