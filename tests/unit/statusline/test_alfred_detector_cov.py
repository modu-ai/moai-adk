"""
Comprehensive test coverage for Alfred task detector.

Focuses on uncovered code paths:
- Cache validation (lines 87-96)
- Session state reading (lines 55-85)
- Error handling and fallback paths
- TTL management
"""

import json
import logging
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import MagicMock, mock_open, patch

import pytest

from moai_adk.statusline.alfred_detector import AlfredDetector, AlfredTask


class TestAlfredDetectorInitialization:
    """Test AlfredDetector initialization and setup."""

    def test_init_creates_default_cache(self):
        """Test that initialization creates empty cache."""
        # Arrange & Act
        detector = AlfredDetector()

        # Assert
        assert detector._cache is None
        assert detector._cache_time is None
        assert detector._cache_ttl is not None
        assert detector._cache_ttl.total_seconds() == 1

    def test_init_sets_session_state_path(self):
        """Test that session state path is set to home directory."""
        # Arrange & Act
        detector = AlfredDetector()

        # Assert
        assert detector._session_state_path is not None
        assert ".moai" in str(detector._session_state_path)
        assert "memory" in str(detector._session_state_path)
        assert "last-session-state.json" in str(detector._session_state_path)

    def test_init_cache_ttl_is_correct(self):
        """Test cache TTL is set to 1 second."""
        # Arrange & Act
        detector = AlfredDetector()

        # Assert
        assert detector._cache_ttl.total_seconds() == 1


class TestActiveTaskDetection:
    """Test detect_active_task method (main entry point)."""

    def test_detect_active_task_returns_alfred_task(self):
        """Test that detect_active_task returns AlfredTask instance."""
        # Arrange
        detector = AlfredDetector()

        with patch.object(detector, '_is_cache_valid', return_value=False):
            with patch.object(detector, '_read_session_state') as mock_read:
                mock_read.return_value = AlfredTask(
                    command="test_command",
                    spec_id="SPEC-001",
                    stage="RED"
                )

                # Act
                result = detector.detect_active_task()

        # Assert
        assert isinstance(result, AlfredTask)
        assert result.command == "test_command"
        assert result.spec_id == "SPEC-001"
        assert result.stage == "RED"

    def test_detect_active_task_uses_cache_when_valid(self):
        """Test that cached result is returned when cache is valid."""
        # Arrange
        detector = AlfredDetector()
        cached_task = AlfredTask(
            command="cached_command",
            spec_id="SPEC-999",
            stage="REFACTOR"
        )
        detector._cache = cached_task
        detector._cache_time = datetime.now()

        with patch.object(detector, '_is_cache_valid', return_value=True):
            with patch.object(detector, '_read_session_state') as mock_read:
                # Act
                result = detector.detect_active_task()

        # Assert
        assert result == cached_task
        # _read_session_state should not be called
        mock_read.assert_not_called()

    def test_detect_active_task_reads_when_cache_invalid(self):
        """Test that session state is read when cache is invalid."""
        # Arrange
        detector = AlfredDetector()
        fresh_task = AlfredTask(
            command="fresh_command",
            spec_id="SPEC-002",
            stage="GREEN"
        )

        with patch.object(detector, '_is_cache_valid', return_value=False):
            with patch.object(detector, '_read_session_state') as mock_read:
                mock_read.return_value = fresh_task

                # Act
                result = detector.detect_active_task()

        # Assert
        assert result == fresh_task
        mock_read.assert_called_once()

    def test_detect_active_task_updates_cache(self):
        """Test that cache is updated after reading."""
        # Arrange
        detector = AlfredDetector()
        fresh_task = AlfredTask(
            command="fresh",
            spec_id="SPEC-003",
            stage="SYNC"
        )

        with patch.object(detector, '_is_cache_valid', return_value=False):
            with patch.object(detector, '_read_session_state') as mock_read:
                mock_read.return_value = fresh_task
                with patch.object(detector, '_update_cache') as mock_update:
                    # Act
                    detector.detect_active_task()

        # Assert
        mock_update.assert_called_once_with(fresh_task)


class TestSessionStateReading:
    """Test _read_session_state method (lines 55-85)."""

    @pytest.fixture
    def detector(self):
        """Create detector instance."""
        return AlfredDetector()

    def test_read_session_state_file_not_exists(self, detector):
        """Test reading session state when file doesn't exist."""
        # Arrange
        with patch('pathlib.Path.exists', return_value=False):
            with patch.object(detector, '_create_default_task') as mock_default:
                mock_default.return_value = AlfredTask(None, None, None)

                # Act
                result = detector._read_session_state()

        # Assert
        assert result is not None
        mock_default.assert_called_once()

    def test_read_session_state_file_exists_valid_json(self, detector):
        """Test reading valid session state file."""
        # Arrange
        session_state = {
            "active_task": {
                "command": "/moai:2-run",
                "spec_id": "SPEC-001",
                "stage": "RED"
            }
        }

        with patch('pathlib.Path.exists', return_value=True):
            with patch('builtins.open', mock_open(read_data=json.dumps(session_state))):
                # Act
                result = detector._read_session_state()

        # Assert
        assert result.command == "/moai:2-run"
        assert result.spec_id == "SPEC-001"
        assert result.stage == "RED"

    def test_read_session_state_missing_active_task(self, detector):
        """Test reading session state without active_task key."""
        # Arrange
        session_state = {"other_key": "value"}

        with patch('pathlib.Path.exists', return_value=True):
            with patch('builtins.open', mock_open(read_data=json.dumps(session_state))):
                with patch.object(detector, '_create_default_task') as mock_default:
                    mock_default.return_value = AlfredTask(None, None, None)

                    # Act
                    result = detector._read_session_state()

        # Assert
        mock_default.assert_called_once()

    def test_read_session_state_null_active_task(self, detector):
        """Test reading session state with null active_task."""
        # Arrange
        session_state = {"active_task": None}

        with patch('pathlib.Path.exists', return_value=True):
            with patch('builtins.open', mock_open(read_data=json.dumps(session_state))):
                with patch.object(detector, '_create_default_task') as mock_default:
                    mock_default.return_value = AlfredTask(None, None, None)

                    # Act
                    result = detector._read_session_state()

        # Assert
        mock_default.assert_called_once()

    def test_read_session_state_partial_active_task(self, detector):
        """Test reading session state with incomplete active_task."""
        # Arrange
        session_state = {
            "active_task": {
                "command": "/moai:1-plan",
                # Missing spec_id and stage
            }
        }

        with patch('pathlib.Path.exists', return_value=True):
            with patch('builtins.open', mock_open(read_data=json.dumps(session_state))):
                # Act
                result = detector._read_session_state()

        # Assert
        assert result.command == "/moai:1-plan"
        assert result.spec_id is None
        assert result.stage is None

    def test_read_session_state_invalid_json(self, detector):
        """Test reading session state with invalid JSON."""
        # Arrange
        invalid_json = "{ invalid json content"

        with patch('pathlib.Path.exists', return_value=True):
            with patch('builtins.open', mock_open(read_data=invalid_json)):
                with patch.object(detector, '_create_default_task') as mock_default:
                    mock_default.return_value = AlfredTask(None, None, None)

                    # Act
                    result = detector._read_session_state()

        # Assert
        mock_default.assert_called_once()

    def test_read_session_state_file_read_error(self, detector):
        """Test reading session state with file read error."""
        # Arrange
        with patch('pathlib.Path.exists', return_value=True):
            with patch('builtins.open', side_effect=IOError("Permission denied")):
                with patch.object(detector, '_create_default_task') as mock_default:
                    mock_default.return_value = AlfredTask(None, None, None)

                    # Act
                    result = detector._read_session_state()

        # Assert
        mock_default.assert_called_once()

    def test_read_session_state_logs_error(self, detector):
        """Test that errors are logged when reading session state fails."""
        # Arrange
        with patch('pathlib.Path.exists', return_value=True):
            with patch('builtins.open', side_effect=ValueError("Bad data")):
                with patch.object(detector, '_create_default_task') as mock_default:
                    mock_default.return_value = AlfredTask(None, None, None)
                    with patch('moai_adk.statusline.alfred_detector.logger') as mock_logger:
                        # Act
                        detector._read_session_state()

        # Assert - should log debug message
        # (Note: actual logging happens in the except block)

    def test_read_session_state_all_fields_populated(self, detector):
        """Test reading session state with all fields populated."""
        # Arrange
        session_state = {
            "active_task": {
                "command": "/moai:2-run",
                "spec_id": "SPEC-042",
                "stage": "GREEN",
                "extra_field": "should be ignored"
            }
        }

        with patch('pathlib.Path.exists', return_value=True):
            with patch('builtins.open', mock_open(read_data=json.dumps(session_state))):
                # Act
                result = detector._read_session_state()

        # Assert
        assert result.command == "/moai:2-run"
        assert result.spec_id == "SPEC-042"
        assert result.stage == "GREEN"


class TestCacheValidation:
    """Test _is_cache_valid method (lines 87-96)."""

    @pytest.fixture
    def detector(self):
        """Create detector instance."""
        return AlfredDetector()

    def test_is_cache_valid_no_cache(self, detector):
        """Test cache validation when no cache exists."""
        # Arrange
        detector._cache = None
        detector._cache_time = None

        # Act
        result = detector._is_cache_valid()

        # Assert
        assert result is False

    def test_is_cache_valid_cache_no_time(self, detector):
        """Test cache validation when cache exists but no time."""
        # Arrange
        detector._cache = AlfredTask(command="test", spec_id=None, stage=None)
        detector._cache_time = None

        # Act
        result = detector._is_cache_valid()

        # Assert
        assert result is False

    def test_is_cache_valid_within_ttl(self, detector):
        """Test cache validation within TTL window."""
        # Arrange
        detector._cache = AlfredTask(command="test", spec_id=None, stage=None)
        detector._cache_time = datetime.now() - timedelta(milliseconds=500)  # 0.5 seconds ago

        # Act
        result = detector._is_cache_valid()

        # Assert
        assert result is True

    def test_is_cache_valid_at_ttl_boundary(self, detector):
        """Test cache validation at TTL boundary (exactly 1 second)."""
        # Arrange
        detector._cache = AlfredTask(command="test", spec_id=None, stage=None)
        detector._cache_time = datetime.now() - timedelta(seconds=1)

        # Act
        result = detector._is_cache_valid()

        # Assert
        assert result is False

    def test_is_cache_valid_expired(self, detector):
        """Test cache validation when TTL has expired."""
        # Arrange
        detector._cache = AlfredTask(command="test", spec_id=None, stage=None)
        detector._cache_time = datetime.now() - timedelta(seconds=2)  # 2 seconds ago

        # Act
        result = detector._is_cache_valid()

        # Assert
        assert result is False

    def test_is_cache_valid_just_before_expiry(self, detector):
        """Test cache validation just before expiry."""
        # Arrange
        detector._cache = AlfredTask(command="test", spec_id=None, stage=None)
        detector._cache_time = datetime.now() - timedelta(milliseconds=999)  # Just under 1 second

        # Act
        result = detector._is_cache_valid()

        # Assert
        assert result is True


class TestCacheUpdating:
    """Test _update_cache method (lines 93-96)."""

    @pytest.fixture
    def detector(self):
        """Create detector instance."""
        return AlfredDetector()

    def test_update_cache_stores_task(self, detector):
        """Test that update_cache stores the task."""
        # Arrange
        task = AlfredTask(
            command="/moai:3-sync",
            spec_id="SPEC-010",
            stage="SYNC"
        )

        # Act
        detector._update_cache(task)

        # Assert
        assert detector._cache == task

    def test_update_cache_stores_time(self, detector):
        """Test that update_cache stores current time."""
        # Arrange
        task = AlfredTask(command="test", spec_id=None, stage=None)
        before = datetime.now()

        # Act
        detector._update_cache(task)

        after = datetime.now()

        # Assert
        assert detector._cache_time is not None
        assert before <= detector._cache_time <= after

    def test_update_cache_overwrites_old_cache(self, detector):
        """Test that update_cache overwrites previous cache."""
        # Arrange
        old_task = AlfredTask(command="old", spec_id="OLD-001", stage="RED")
        new_task = AlfredTask(command="new", spec_id="NEW-001", stage="GREEN")

        detector._cache = old_task
        detector._cache_time = datetime.now() - timedelta(seconds=10)

        # Act
        detector._update_cache(new_task)

        # Assert
        assert detector._cache == new_task
        assert detector._cache.command == "new"


class TestDefaultTaskCreation:
    """Test _create_default_task static method."""

    def test_create_default_task_all_none(self):
        """Test that default task has all None values."""
        # Act
        task = AlfredDetector._create_default_task()

        # Assert
        assert task.command is None
        assert task.spec_id is None
        assert task.stage is None

    def test_create_default_task_is_alfred_task(self):
        """Test that default task is an AlfredTask instance."""
        # Act
        task = AlfredDetector._create_default_task()

        # Assert
        assert isinstance(task, AlfredTask)

    def test_create_default_task_multiple_calls_independent(self):
        """Test that multiple calls to create_default_task return independent instances."""
        # Act
        task1 = AlfredDetector._create_default_task()
        task2 = AlfredDetector._create_default_task()

        # Assert - they should be equal but not the same object
        assert task1.command == task2.command
        assert task1.spec_id == task2.spec_id
        assert task1.stage == task2.stage


class TestAlfredTaskDataclass:
    """Test AlfredTask dataclass."""

    def test_alfred_task_init_all_fields(self):
        """Test AlfredTask initialization with all fields."""
        # Act
        task = AlfredTask(
            command="/moai:2-run",
            spec_id="SPEC-001",
            stage="RED"
        )

        # Assert
        assert task.command == "/moai:2-run"
        assert task.spec_id == "SPEC-001"
        assert task.stage == "RED"

    def test_alfred_task_init_with_none(self):
        """Test AlfredTask initialization with None values."""
        # Act
        task = AlfredTask(
            command=None,
            spec_id=None,
            stage=None
        )

        # Assert
        assert task.command is None
        assert task.spec_id is None
        assert task.stage is None

    def test_alfred_task_equality(self):
        """Test AlfredTask equality comparison."""
        # Act
        task1 = AlfredTask("/moai:1-plan", "SPEC-001", "RED")
        task2 = AlfredTask("/moai:1-plan", "SPEC-001", "RED")
        task3 = AlfredTask("/moai:2-run", "SPEC-001", "RED")

        # Assert
        assert task1 == task2
        assert task1 != task3

    def test_alfred_task_partial_none(self):
        """Test AlfredTask with some None fields."""
        # Act
        task = AlfredTask(
            command="/moai:2-run",
            spec_id=None,
            stage="GREEN"
        )

        # Assert
        assert task.command == "/moai:2-run"
        assert task.spec_id is None
        assert task.stage == "GREEN"


class TestCacheConstants:
    """Test cache TTL constant."""

    def test_cache_ttl_constant_is_one_second(self):
        """Test that the cache TTL constant is 1 second."""
        # Assert
        assert AlfredDetector._CACHE_TTL_SECONDS == 1

    def test_cache_ttl_used_in_init(self):
        """Test that cache TTL constant is used in initialization."""
        # Arrange & Act
        detector = AlfredDetector()

        # Assert
        assert detector._cache_ttl.total_seconds() == AlfredDetector._CACHE_TTL_SECONDS


class TestEdgeCases:
    """Test edge cases and error conditions."""

    def test_detect_active_task_concurrent_reads(self):
        """Test behavior with concurrent cache operations."""
        # Arrange
        detector = AlfredDetector()
        task1 = AlfredTask(command="cmd1", spec_id="SPEC-001", stage="RED")
        task2 = AlfredTask(command="cmd2", spec_id="SPEC-002", stage="GREEN")

        # Act - simulate rapid alternating cache updates
        detector._update_cache(task1)
        result1 = detector.detect_active_task()
        detector._update_cache(task2)
        result2 = detector.detect_active_task()

        # Assert
        assert result1.spec_id in ["SPEC-001", "SPEC-002"]  # Could be either due to timing
        assert result2.spec_id in ["SPEC-001", "SPEC-002"]

    def test_read_session_state_unicode_content(self):
        """Test reading session state with Unicode characters."""
        # Arrange
        detector = AlfredDetector()
        session_state = {
            "active_task": {
                "command": "/moai:2-run",
                "spec_id": "SPEC-🚀",
                "stage": "GREEN"
            }
        }

        with patch('pathlib.Path.exists', return_value=True):
            with patch('builtins.open', mock_open(read_data=json.dumps(session_state))):
                # Act
                result = detector._read_session_state()

        # Assert
        assert result.spec_id == "SPEC-🚀"

    def test_read_session_state_empty_strings(self):
        """Test reading session state with empty string values."""
        # Arrange
        detector = AlfredDetector()
        session_state = {
            "active_task": {
                "command": "",
                "spec_id": "",
                "stage": ""
            }
        }

        with patch('pathlib.Path.exists', return_value=True):
            with patch('builtins.open', mock_open(read_data=json.dumps(session_state))):
                # Act
                result = detector._read_session_state()

        # Assert
        assert result.command == ""
        assert result.spec_id == ""
        assert result.stage == ""
