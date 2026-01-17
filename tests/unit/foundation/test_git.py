"""Tests for moai_adk.foundation.git module."""

from moai_adk.foundation.git import (
    BranchingStrategySelector,
    ConventionalCommitValidator,
    GitInfo,
    GitPerformanceOptimizer,
    GitVersionDetector,
    GitWorkflowManager,
    DDDCommitPhase,
    ValidateResult,
)


class TestGitInfo:
    """Test GitInfo dataclass."""

    def test_git_info_init(self):
        """Test GitInfo initialization."""
        info = GitInfo(version="2.40.0")
        assert info.version == "2.40.0"
        assert info.supports_switch is False
        assert info.supports_worktree is False

    def test_git_info_with_features(self):
        """Test GitInfo with features."""
        info = GitInfo(
            version="2.40.0",
            supports_switch=True,
            supports_worktree=True,
            modern_features=["switch", "worktree"],
        )
        assert info.supports_switch is True
        assert info.supports_worktree is True


class TestValidateResult:
    """Test ValidateResult dataclass."""

    def test_validate_result_valid(self):
        """Test ValidateResult with valid commit."""
        result = ValidateResult(
            is_valid=True,
            commit_type="feat",
            scope="auth",
            subject="add login",
        )
        assert result.is_valid is True
        assert result.commit_type == "feat"

    def test_validate_result_invalid(self):
        """Test ValidateResult with invalid commit."""
        result = ValidateResult(
            is_valid=False,
            errors=["Invalid format"],
        )
        assert result.is_valid is False
        assert len(result.errors) > 0


class TestDDDCommitPhase:
    """Test DDDCommitPhase dataclass."""

    def test_tdd_commit_phase_red(self):
        """Test DDDCommitPhase for RED phase."""
        phase = DDDCommitPhase(
            phase_name="RED",
            commit_type="test",
            description="Write failing tests",
            test_status="failing",
        )
        assert phase.phase_name == "RED"
        assert phase.commit_type == "test"


class TestGitVersionDetector:
    """Test GitVersionDetector class."""

    def test_detect_git_2_40(self):
        """Test detect method with Git 2.40."""
        detector = GitVersionDetector()
        result = detector.detect("git version 2.40.0")
        assert result.version == "2.40.0"
        assert result.supports_switch is True
        assert result.supports_worktree is True

    def test_detect_git_2_30(self):
        """Test detect method with Git 2.30."""
        detector = GitVersionDetector()
        result = detector.detect("git version 2.30.0")
        assert result.version == "2.30.0"
        assert result.supports_worktree is True

    def test_detect_invalid_version_string(self):
        """Test detect with invalid version string."""
        detector = GitVersionDetector()
        result = detector.detect("invalid version")
        assert result.version == "unknown"

    def test_get_recommended_commands_modern_git(self):
        """Test get_recommended_commands with modern Git."""
        detector = GitVersionDetector()
        git_info = GitInfo(
            version="2.40.0",
            supports_switch=True,
            supports_worktree=True,
        )
        commands = detector.get_recommended_commands(git_info)
        assert "create_branch" in commands
        assert "git switch -c" in commands["create_branch"]

    def test_get_recommended_commands_legacy_git(self):
        """Test get_recommended_commands with legacy Git."""
        detector = GitVersionDetector()
        git_info = GitInfo(version="2.20.0")
        commands = detector.get_recommended_commands(git_info)
        assert "git checkout -b" in commands["create_branch"]


class TestConventionalCommitValidator:
    """Test ConventionalCommitValidator class."""

    def test_validate_valid_commit(self):
        """Test validate with valid commit message."""
        validator = ConventionalCommitValidator()
        result = validator.validate("feat(auth): add login functionality")
        assert result.is_valid is True
        assert result.commit_type == "feat"
        assert result.scope == "auth"

    def test_validate_commit_with_breaking_change(self):
        """Test validate with breaking change."""
        validator = ConventionalCommitValidator()
        result = validator.validate("feat!: change API")
        assert result.is_valid is True
        assert result.is_breaking_change is True

    def test_validate_fix_commit(self):
        """Test validate with fix type."""
        validator = ConventionalCommitValidator()
        result = validator.validate("fix(database): correct connection")
        assert result.is_valid is True
        assert result.commit_type == "fix"

    def test_validate_invalid_format(self):
        """Test validate with invalid format."""
        validator = ConventionalCommitValidator()
        result = validator.validate("invalid format message")
        assert result.is_valid is False
        assert len(result.errors) > 0

    def test_validate_batch(self):
        """Test validate_batch method."""
        validator = ConventionalCommitValidator()
        messages = [
            "feat(auth): add login",
            "fix(db): correct query",
        ]
        results = validator.validate_batch(messages)
        assert isinstance(results, dict)
        assert len(results) == 2

    def test_validate_all_commit_types(self):
        """Test validate with all valid commit types."""
        validator = ConventionalCommitValidator()
        types = ["feat", "fix", "docs", "style", "refactor", "perf", "test", "chore"]
        for commit_type in types:
            result = validator.validate(f"{commit_type}: test message")
            assert result.is_valid is True


class TestBranchingStrategySelector:
    """Test BranchingStrategySelector class."""

    def test_select_strategy_team(self):
        """Test select_strategy for team environment."""
        selector = BranchingStrategySelector()
        strategy = selector.select_strategy(team_size=3, risk_level="high", need_review=True)
        assert strategy == "feature_branch"

    def test_select_strategy_solo_low_risk(self):
        """Test select_strategy for solo developer."""
        selector = BranchingStrategySelector()
        strategy = selector.select_strategy(team_size=1, risk_level="low", need_review=False)
        assert strategy == "direct_commit"

    def test_select_strategy_per_spec(self):
        """Test select_strategy defaults to per_spec."""
        selector = BranchingStrategySelector()
        strategy = selector.select_strategy(team_size=1, risk_level="medium", need_review=False)
        assert strategy in ["direct_commit", "per_spec"]

    def test_get_strategy_details(self):
        """Test get_strategy_details method."""
        selector = BranchingStrategySelector()
        details = selector.get_strategy_details("feature_branch")
        assert "name" in details
        assert "description" in details
        assert "commands" in details


class TestGitWorkflowManager:
    """Test GitWorkflowManager class."""

    def test_create_branch_command_modern(self):
        """Test create_branch_command with modern Git."""
        manager = GitWorkflowManager()
        command = manager.create_branch_command("feature/test", use_modern=True)
        assert "git switch -c" in command
        assert "feature/test" in command

    def test_create_branch_command_legacy(self):
        """Test create_branch_command with legacy Git."""
        manager = GitWorkflowManager()
        command = manager.create_branch_command("feature/test", use_modern=False)
        assert "git checkout -b" in command

    def test_format_ddd_commit_red(self):
        """Test format_ddd_commit for RED phase."""
        manager = GitWorkflowManager()
        commit = manager.format_ddd_commit("test", "auth", "add login tests", "RED")
        assert "test(auth): add login tests" in commit
        assert "RED phase" in commit

    def test_format_ddd_commit_green(self):
        """Test format_ddd_commit for GREEN phase."""
        manager = GitWorkflowManager()
        commit = manager.format_ddd_commit("feat", "auth", "implement login", "GREEN")
        assert "feat(auth): implement login" in commit
        assert "GREEN phase" in commit

    def test_format_ddd_commit_refactor(self):
        """Test format_ddd_commit for REFACTOR phase."""
        manager = GitWorkflowManager()
        commit = manager.format_ddd_commit("refactor", "auth", "improve performance", "REFACTOR")
        assert "refactor(auth): improve performance" in commit
        assert "REFACTOR phase" in commit

    def test_get_workflow_commands_feature_branch(self):
        """Test get_workflow_commands for feature_branch strategy."""
        manager = GitWorkflowManager()
        commands = manager.get_workflow_commands("feature_branch", "SPEC-001")
        assert isinstance(commands, list)
        assert len(commands) > 0
        assert any("SPEC-001" in cmd for cmd in commands)

    def test_get_workflow_commands_direct_commit(self):
        """Test get_workflow_commands for direct_commit strategy."""
        manager = GitWorkflowManager()
        commands = manager.get_workflow_commands("direct_commit", "SPEC-001")
        assert isinstance(commands, list)
        assert any("develop" in cmd for cmd in commands)


class TestGitPerformanceOptimizer:
    """Test GitPerformanceOptimizer class."""

    def test_get_tips_for_repo_size_small(self):
        """Test get_tips_for_repo_size for small repos."""
        optimizer = GitPerformanceOptimizer()
        tips = optimizer.get_tips_for_repo_size("small")
        assert isinstance(tips, list)

    def test_get_tips_for_repo_size_large(self):
        """Test get_tips_for_repo_size for large repos."""
        optimizer = GitPerformanceOptimizer()
        tips = optimizer.get_tips_for_repo_size("large")
        assert isinstance(tips, list)
        assert len(tips) > 0

    def test_get_optimization_configs(self):
        """Test get_optimization_configs method."""
        optimizer = GitPerformanceOptimizer()
        configs = optimizer.get_optimization_configs()
        assert isinstance(configs, dict)
        assert "midx" in configs
        assert "partial_clone" in configs

    def test_get_recommended_optimizations_large(self):
        """Test get_recommended_optimizations for large repos."""
        optimizer = GitPerformanceOptimizer()
        optimizations = optimizer.get_recommended_optimizations("large")
        assert isinstance(optimizations, list)
        assert "midx" in optimizations

    def test_get_recommended_optimizations_medium(self):
        """Test get_recommended_optimizations for medium repos."""
        optimizer = GitPerformanceOptimizer()
        optimizations = optimizer.get_recommended_optimizations("medium")
        assert isinstance(optimizations, list)

    def test_get_recommended_optimizations_small(self):
        """Test get_recommended_optimizations for small repos."""
        optimizer = GitPerformanceOptimizer()
        optimizations = optimizer.get_recommended_optimizations("small")
        assert isinstance(optimizations, list)
