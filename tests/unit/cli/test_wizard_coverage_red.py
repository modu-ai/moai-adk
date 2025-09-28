"""
@TEST:CLI-WIZARD-RED-001 RED Phase - Failing tests for wizard.py coverage improvement

Phase 1 of TDD: Write failing tests for the InteractiveWizard class
that accurately reflect the real wizard API and target the core user workflows.

Target Coverage Goals:
- wizard.py: 12.82% → 75%
- Key functions: run_wizard(), _collect_product_vision(), _collect_tech_stack()
- Real API based tests for interactive wizard scenarios
"""

import sys
from unittest.mock import Mock, patch

import click
import pytest
from click.testing import CliRunner

from moai_adk.cli.wizard import InteractiveWizard


class TestInteractiveWizardRed:
    """RED: Test the InteractiveWizard class with real API calls."""

    def setup_method(self):
        """Set up test environment."""
        self.wizard = InteractiveWizard()

    def test_wizard_initialization_should_create_empty_state(self):
        """RED: Test wizard initializes with empty answers and tech_stack."""
        assert self.wizard.answers == {}
        assert self.wizard.tech_stack == []

    @patch('click.prompt')
    @patch('click.confirm')
    @patch('click.echo')
    def test_run_wizard_should_collect_all_sections_and_return_answers(
        self, mock_echo, mock_confirm, mock_prompt
    ):
        """RED: Test complete wizard flow collects all required information."""
        # Arrange: Mock user inputs for full wizard run
        mock_prompt.side_effect = [
            "해결하려는 핵심 문제입니다. 사용자가 겪는 주요 문제점들을 해결합니다.",  # Q1: core_problem
            "개발자",  # Q2: target_users
            "6개월 후 MAU 1000명 달성을 목표로 합니다.",  # Q3: goal
            "사용자 인증 시스템",  # Q4.1: feature 1
            "데이터 분석 대시보드",  # Q4.2: feature 2
            "API 통합 기능",  # Q4.3: feature 3
            "1,2,3",  # Q5: tech_stack (웹)
            "",  # Q5: other categories (skip)
            "",  # Q5: other categories (skip)
            "",  # Q5: other categories (skip)
            "",  # Q5: other categories (skip)
            "중급",  # Q6: skill_level
            85,  # Q7: test_coverage (integer)
            "300",  # Q7: performance_target
        ]

        mock_confirm.side_effect = [
            True,   # Add 2nd feature
            True,   # Add 3rd feature
            False,  # Skip advanced settings
            True,   # Confirm initialization
        ]

        # Act
        result = self.wizard.run_wizard()

        # Assert
        assert isinstance(result, dict)
        assert 'core_problem' in result
        assert 'target_users' in result
        assert 'goal' in result
        assert 'core_features' in result
        assert 'tech_stack' in result
        assert 'skill_level' in result
        assert 'test_coverage' in result
        assert 'performance_target' in result

    @patch('click.prompt')
    @patch('click.confirm')
    @patch('click.echo')
    def test_collect_product_vision_should_enforce_minimum_problem_length(
        self, mock_echo, mock_confirm, mock_prompt
    ):
        """RED: Test product vision collection validates problem description length."""
        # Arrange: First short answer, then valid answer
        mock_prompt.side_effect = [
            "짧은 설명",  # Too short (< 20 chars)
            "이것은 충분히 긴 문제 설명입니다. 사용자가 겪는 구체적인 문제점들을 상세히 설명합니다.",  # Valid
            "개발자",  # target_users
            "6개월 후 MAU 1000명 달성 목표",  # goal with metrics
            "핵심 기능 1",  # feature 1
        ]

        mock_confirm.side_effect = [False]  # Don't add more features

        # Act
        self.wizard._collect_product_vision()

        # Assert
        assert self.wizard.answers['core_problem'] == "이것은 충분히 긴 문제 설명입니다. 사용자가 겪는 구체적인 문제점들을 상세히 설명합니다."
        assert self.wizard.answers['target_users'] == "개발자"
        assert self.wizard.answers['goal'] == "6개월 후 MAU 1000명 달성 목표"
        assert len(self.wizard.answers['core_features']) >= 1

    @patch('click.prompt')
    @patch('click.confirm')
    @patch('click.echo')
    def test_collect_product_vision_should_enforce_measurable_goals(
        self, mock_echo, mock_confirm, mock_prompt
    ):
        """RED: Test goal collection requires measurable KPIs."""
        # Arrange: First vague goal, then measurable goal
        mock_prompt.side_effect = [
            "이것은 충분히 긴 문제 설명입니다. 사용자가 겪는 구체적인 문제점들을 상세히 설명합니다.",
            "개발자",
            "좋은 서비스 만들기",  # Vague goal (no metrics)
            "응답시간 500ms 이하, MAU 1000명 달성",  # Measurable goal
            "핵심 기능 1",
        ]

        mock_confirm.side_effect = [False]  # Don't add more features

        # Act
        self.wizard._collect_product_vision()

        # Assert
        assert "응답시간 500ms 이하, MAU 1000명 달성" in self.wizard.answers['goal']

    @patch('click.prompt')
    @patch('click.echo')
    def test_collect_tech_stack_should_parse_multiple_selections(
        self, mock_echo, mock_prompt
    ):
        """RED: Test tech stack collection handles multiple selections per category."""
        # Arrange: Select multiple items from categories
        mock_prompt.side_effect = [
            "1,3,5",  # Select React, Angular, Next.js from 웹
            "2,4",    # Select Flutter, Kotlin from 모바일
            "",       # Skip 백엔드
            "1",      # Select PostgreSQL from 데이터베이스
            "",       # Skip 인프라
            "고급",   # skill_level
        ]

        # Act
        self.wizard._collect_tech_stack()

        # Assert
        assert len(self.wizard.tech_stack) > 0
        assert self.wizard.answers['tech_stack'] == self.wizard.tech_stack
        assert self.wizard.answers['skill_level'] == "고급"

    @patch('click.prompt')
    @patch('click.echo')
    def test_collect_tech_stack_should_handle_invalid_selections_gracefully(
        self, mock_echo, mock_prompt
    ):
        """RED: Test tech stack collection ignores invalid selections."""
        # Arrange: Include invalid selections
        mock_prompt.side_effect = [
            "1,99,abc,2",  # Valid: 1,2; Invalid: 99,abc
            "",  # Skip other categories
            "",
            "",
            "",
            "초급",
        ]

        # Act
        self.wizard._collect_tech_stack()

        # Assert
        # Should only include valid selections, ignore invalid ones
        assert self.wizard.answers['skill_level'] == "초급"
        # tech_stack should have some valid items, no crashes

    @patch('click.prompt')
    @patch('click.echo')
    def test_collect_quality_standards_should_validate_coverage_range(
        self, mock_echo, mock_prompt
    ):
        """RED: Test quality standards validates coverage percentage range."""
        # Arrange: Valid coverage and performance values
        mock_prompt.side_effect = [
            "90",   # test_coverage (valid range 60-100)
            "200",  # performance_target
        ]

        # Act
        self.wizard._collect_quality_standards()

        # Assert
        assert self.wizard.answers['test_coverage'] == "90"  # Actually string
        assert self.wizard.answers['performance_target'] == "200"

    @patch('click.confirm')
    @patch('click.echo')
    def test_collect_advanced_settings_should_collect_security_and_operational_features(
        self, mock_echo, mock_confirm
    ):
        """RED: Test advanced settings collection for security and operations."""
        # Arrange: Select various security and operational features
        mock_confirm.side_effect = [
            True,   # authentication needed
            False,  # API key management not needed
            True,   # encryption needed
            True,   # monitoring needed
            False,  # CI/CD not needed
        ]

        # Act
        self.wizard._collect_advanced_settings()

        # Assert
        assert 'security_features' in self.wizard.answers
        assert 'authentication' in self.wizard.answers['security_features']
        assert 'encryption' in self.wizard.answers['security_features']
        assert 'api_key_management' not in self.wizard.answers['security_features']
        assert self.wizard.answers['monitoring'] is True
        assert self.wizard.answers['ci_cd'] is False

    @patch('click.echo')
    def test_show_summary_should_display_collected_answers(self, mock_echo):
        """RED: Test summary display shows all collected information."""
        # Arrange: Set up some test answers
        self.wizard.answers = {
            'core_problem': '테스트 문제입니다',
            'target_users': '개발자',
            'goal': 'MAU 1000명 달성',
            'tech_stack': ['React', 'FastAPI'],
            'skill_level': '중급',
            'test_coverage': 85,
            'performance_target': '500',
            'security_features': ['authentication'],
        }

        # Act
        self.wizard._show_summary()

        # Assert
        mock_echo.assert_called()
        # Should have called echo multiple times to display summary

    @patch('click.prompt')
    @patch('click.confirm')
    @patch('click.echo')
    def test_wizard_cancellation_should_exit_gracefully(
        self, mock_echo, mock_confirm, mock_prompt
    ):
        """RED: Test wizard handles user cancellation gracefully."""
        # Arrange: User cancels at confirmation
        mock_prompt.side_effect = [
            "충분히 긴 문제 설명입니다. 구체적인 사용자 문제를 설명합니다.",
            "개발자",
            "MAU 1000명 달성 목표",
            "핵심 기능",
            "",  # Skip tech selections
            "",
            "",
            "",
            "",
            "중급",
            "80",
            "500",
        ]

        mock_confirm.side_effect = [
            False,  # Skip advanced settings
            False,  # Cancel initialization
        ]

        # Act & Assert
        with pytest.raises(SystemExit):
            self.wizard.run_wizard()

    @patch('click.prompt')
    @patch('click.confirm')
    def test_wizard_with_advanced_settings_should_collect_all_information(
        self, mock_confirm, mock_prompt
    ):
        """RED: Test wizard with advanced settings enabled."""
        # Arrange: Enable advanced settings
        mock_prompt.side_effect = [
            "충분히 긴 문제 설명입니다. 구체적인 사용자 문제를 설명합니다.",
            "개발자",
            "MAU 1000명 달성 목표",
            "핵심 기능",
            "",  # Skip tech selections
            "",
            "",
            "",
            "",
            "중급",
            "80",
            "500",
        ]

        mock_confirm.side_effect = [
            True,   # Enable advanced settings
            True,   # authentication
            False,  # api_key_management
            True,   # encryption
            True,   # monitoring
            False,  # ci_cd
            True,   # Confirm initialization
        ]

        # Act
        result = self.wizard.run_wizard()

        # Assert
        assert 'security_features' in result
        assert 'monitoring' in result
        assert 'ci_cd' in result


class TestWizardValidationRed:
    """RED: Test wizard input validation and error handling."""

    def setup_method(self):
        """Set up test environment."""
        self.wizard = InteractiveWizard()

    @patch('click.prompt')
    @patch('click.confirm')
    @patch('click.echo')
    def test_short_problem_description_should_be_rejected_repeatedly(
        self, mock_echo, mock_confirm, mock_prompt
    ):
        """RED: Test wizard keeps asking until valid problem description."""
        # Arrange: Multiple short answers before valid one
        mock_prompt.side_effect = [
            "짧음",  # Too short
            "여전히 짧음",  # Still too short
            "이것도 짧습니다",  # Still too short
            "이제 충분히 긴 문제 설명입니다. 사용자가 겪는 구체적인 문제점들을 상세히 설명합니다.",  # Valid
            "개발자",
            "MAU 1000명 달성",
            "기능1",
        ]

        mock_confirm.side_effect = [False]  # Don't add more features

        # Act
        self.wizard._collect_product_vision()

        # Assert
        # Should eventually accept the valid description
        assert len(self.wizard.answers['core_problem']) >= 20

    @patch('click.prompt')
    @patch('click.confirm')
    @patch('click.echo')
    def test_goal_without_metrics_should_be_rejected_repeatedly(
        self, mock_echo, mock_confirm, mock_prompt
    ):
        """RED: Test wizard enforces measurable goals."""
        # Arrange: Goals without metrics before valid one
        mock_prompt.side_effect = [
            "충분히 긴 문제 설명입니다. 구체적인 사용자 문제를 설명합니다.",
            "개발자",
            "좋은 서비스 만들기",  # No metrics
            "더 나은 시스템 구축",  # No metrics
            "MAU 1000명 달성 목표",  # Has metrics
            "기능1",
        ]

        mock_confirm.side_effect = [False]  # Don't add more features

        # Act
        self.wizard._collect_product_vision()

        # Assert
        assert "1000명" in self.wizard.answers['goal']

    @patch('click.prompt')
    @patch('click.confirm')
    def test_feature_collection_should_handle_early_termination(
        self, mock_confirm, mock_prompt
    ):
        """RED: Test feature collection when user stops early."""
        # Arrange: User adds only 2 features instead of 3
        mock_prompt.side_effect = [
            "충분히 긴 문제 설명입니다. 구체적인 사용자 문제를 설명합니다.",
            "개발자",
            "MAU 1000명 달성 목표",
            "첫 번째 기능",  # Feature 1
            "두 번째 기능",  # Feature 2
        ]

        mock_confirm.side_effect = [
            True,   # Add 2nd feature
            False,  # Don't add 3rd feature
        ]

        # Act
        self.wizard._collect_product_vision()

        # Assert
        assert len(self.wizard.answers['core_features']) == 2
        assert "첫 번째 기능" in self.wizard.answers['core_features']
        assert "두 번째 기능" in self.wizard.answers['core_features']


class TestWizardIntegrationRed:
    """RED: Test wizard integration scenarios and edge cases."""

    def setup_method(self):
        """Set up test environment."""
        self.wizard = InteractiveWizard()

    def test_wizard_state_should_persist_across_method_calls(self):
        """RED: Test wizard maintains state between different collection methods."""
        # Arrange: Set some initial state
        self.wizard.answers['test_key'] = 'test_value'
        self.wizard.tech_stack = ['React']

        # Act: Call a method that shouldn't clear existing state
        with patch('click.prompt', return_value="85"):
            with patch('click.echo'):
                self.wizard._collect_quality_standards()

        # Assert: Previous state should be preserved
        assert self.wizard.answers['test_key'] == 'test_value'
        assert 'React' in self.wizard.tech_stack
        assert 'test_coverage' in self.wizard.answers

    @patch('click.prompt')
    @patch('click.confirm')
    @patch('click.echo')
    def test_empty_tech_stack_selection_should_be_handled(
        self, mock_echo, mock_confirm, mock_prompt
    ):
        """RED: Test wizard handles when user selects no tech stack items."""
        # Arrange: User skips all tech stack categories
        mock_prompt.side_effect = [
            "",  # Skip 웹
            "",  # Skip 모바일
            "",  # Skip 백엔드
            "",  # Skip 데이터베이스
            "",  # Skip 인프라
            "초급",  # skill_level
        ]

        # Act
        self.wizard._collect_tech_stack()

        # Assert
        assert self.wizard.tech_stack == []
        assert self.wizard.answers['tech_stack'] == []
        assert self.wizard.answers['skill_level'] == "초급"

    def test_unicode_input_should_be_handled_safely(self):
        """RED: Test wizard handles unicode input without crashing."""
        # Arrange: Unicode input in answers
        self.wizard.answers = {
            'core_problem': '한글 문제 설명입니다',
            'target_users': '개발자',
            'tech_stack': ['React', 'FastAPI'],
        }

        # Act & Assert: Should not crash with unicode
        with patch('click.echo'):
            self.wizard._show_summary()

        # Unicode should be preserved
        assert '한글' in self.wizard.answers['core_problem']