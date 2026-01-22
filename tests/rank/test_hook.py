"""Tests for rank.hook module."""

import json
import os
import tempfile
from pathlib import Path
from unittest.mock import Mock, patch

from moai_adk.rank.client import SessionSubmission
from moai_adk.rank.config import RankConfig
from moai_adk.rank.hook import (
    TokenUsage,
    _is_duplicate_error,
    add_project_exclusion,
    calculate_cost,
    compute_anonymous_project_id,
    create_global_hook_script,
    create_hook_script,
    get_model_pricing,
    is_project_excluded,
    load_rank_config,
    parse_session_data,
    parse_transcript_for_usage,
    parse_transcript_to_submission,
    remove_project_exclusion,
    save_rank_config,
    submit_session_hook,
)


class TestTokenUsage:
    """Test TokenUsage dataclass."""

    def test_token_usage_creation(self):
        """Test creating TokenUsage with all fields."""
        usage = TokenUsage(
            input_tokens=1000,
            output_tokens=500,
            cache_creation_tokens=100,
            cache_read_tokens=50,
            model_name="claude-opus-4-5",
            cost_usd=0.05,
            started_at="2024-01-01T00:00:00Z",
            duration_seconds=3600,
            turn_count=10,
            tool_usage={"Read": 5, "Write": 3},
            model_usage={"claude-opus-4-5": {"input": 1000, "output": 500}},
            code_metrics={"linesAdded": 100, "linesDeleted": 20},
        )
        assert usage.input_tokens == 1000
        assert usage.output_tokens == 500
        assert usage.model_name == "claude-opus-4-5"
        assert usage.cost_usd == 0.05

    def test_token_usage_defaults(self):
        """Test TokenUsage with default values."""
        usage = TokenUsage()
        assert usage.input_tokens == 0
        assert usage.output_tokens == 0
        assert usage.model_name is None
        assert usage.cost_usd == 0.0


class TestModelPricing:
    """Test model pricing functions."""

    def test_get_pricing_opus_4_5(self):
        """Test pricing for Claude Opus 4.5."""
        pricing = get_model_pricing("claude-opus-4-5-20251101")
        assert pricing["input"] == 5.00
        assert pricing["output"] == 25.00

    def test_get_pricing_sonnet_4_5(self):
        """Test pricing for Claude Sonnet 4.5."""
        pricing = get_model_pricing("claude-sonnet-4-5-20251022")
        assert pricing["input"] == 3.00
        assert pricing["output"] == 15.00

    def test_get_pricing_haiku_4_5(self):
        """Test pricing for Claude Haiku 4.5."""
        pricing = get_model_pricing("claude-haiku-4-5-20251022")
        assert pricing["input"] == 1.00
        assert pricing["output"] == 5.00

    def test_get_pricing_default(self):
        """Test default pricing for unknown models."""
        pricing = get_model_pricing("unknown-model")
        assert pricing["input"] == 3.00
        assert pricing["output"] == 15.00

    def test_get_pricing_none(self):
        """Test pricing for None model name."""
        pricing = get_model_pricing(None)
        assert pricing["input"] == 3.00
        assert pricing["output"] == 15.00

    def test_get_pricing_pattern_matching_opus(self):
        """Test pattern matching for Opus family."""
        pricing = get_model_pricing("some-opus-4-5-model")
        assert pricing["input"] == 5.00

    def test_get_pricing_pattern_matching_sonnet(self):
        """Test pattern matching for Sonnet family."""
        pricing = get_model_pricing("some-sonnet-4-model")
        assert pricing["input"] == 3.00

    def test_get_pricing_pattern_matching_haiku(self):
        """Test pattern matching for Haiku family."""
        pricing = get_model_pricing("some-haiku-model")
        assert pricing["input"] == 0.80


class TestCalculateCost:
    """Test cost calculation."""

    def test_calculate_cost_opus(self):
        """Test cost calculation for Opus."""
        usage = TokenUsage(
            input_tokens=1_000_000,
            output_tokens=500_000,
            cache_creation_tokens=100_000,
            cache_read_tokens=50_000,
            model_name="claude-opus-4-5-20251101",
        )
        cost = calculate_cost(usage)
        # Input: 1M * $5 = $5
        # Output: 0.5M * $25 = $12.50
        # Cache creation: 0.1M * $6.25 = $0.625
        # Cache read: 0.05M * $0.50 = $0.025
        # Total: ~$18.15
        assert abs(cost - 18.15) < 0.01

    def test_calculate_cost_sonnet(self):
        """Test cost calculation for Sonnet."""
        usage = TokenUsage(
            input_tokens=1_000_000,
            output_tokens=1_000_000,
            model_name="claude-sonnet-4-5-20251022",
        )
        cost = calculate_cost(usage)
        # Input: 1M * $3 = $3
        # Output: 1M * $15 = $15
        # Total: $18
        assert abs(cost - 18.0) < 0.01

    def test_calculate_cost_zero_tokens(self):
        """Test cost calculation with zero tokens."""
        usage = TokenUsage(model_name="claude-sonnet-4-5")
        cost = calculate_cost(usage)
        assert cost == 0.0


class TestComputeAnonymousProjectId:
    """Test anonymous project ID computation."""

    def test_compute_project_id(self):
        """Test project ID hash computation."""
        project_id = compute_anonymous_project_id("/Users/test/project")
        assert isinstance(project_id, str)
        assert len(project_id) == 16

    def test_compute_project_id_normalized(self):
        """Test that paths are normalized."""
        id1 = compute_anonymous_project_id("/Users/test/project")
        id2 = compute_anonymous_project_id("/Users/test/project/")
        assert id1 == id2

    def test_compute_project_id_expands_home(self):
        """Test that ~ is expanded."""
        with patch.dict(os.environ, {"HOME": "/home/test"}):
            id1 = compute_anonymous_project_id("~/project")
            id2 = compute_anonymous_project_id("/home/test/project")
            assert id1 == id2


class TestParseTranscriptForUsage:
    """Test transcript parsing."""

    def test_parse_nonexistent_file(self):
        """Test parsing nonexistent file returns None."""
        result = parse_transcript_for_usage("/nonexistent/file.jsonl")
        assert result is None

    def test_parse_valid_transcript(self):
        """Test parsing valid transcript file."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".jsonl", delete=False) as f:
            transcript_data = [
                {"timestamp": "2024-01-01T00:00:00Z", "type": "user", "message": {"model": "claude-opus-4-5"}},
                {
                    "timestamp": "2024-01-01T00:01:00Z",
                    "type": "assistant",
                    "message": {
                        "model": "claude-opus-4-5",
                        "usage": {
                            "input_tokens": 1000,
                            "output_tokens": 500,
                            "cache_creation_input_tokens": 100,
                            "cache_read_input_tokens": 50,
                        },
                        "content": [{"type": "tool_use", "name": "Read", "input": {"file_path": "/test/file.py"}}],
                    },
                },
            ]
            for line in transcript_data:
                f.write(json.dumps(line) + "\n")
            f.flush()
            temp_path = f.name

        try:
            result = parse_transcript_for_usage(temp_path)
            assert result is not None
            assert result.input_tokens == 1000
            assert result.output_tokens == 500
            assert result.cache_creation_tokens == 100
            assert result.cache_read_tokens == 50
            assert result.model_name == "claude-opus-4-5"
            assert result.turn_count == 1
            assert result.tool_usage == {"Read": 1}
        finally:
            os.unlink(temp_path)

    def test_parse_transcript_with_edit_tool(self):
        """Test parsing transcript with Edit tool tracks code metrics."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".jsonl", delete=False) as f:
            transcript_data = [
                {
                    "timestamp": "2024-01-01T00:00:00Z",
                    "type": "assistant",
                    "message": {
                        "model": "claude-sonnet-4",
                        "usage": {"input_tokens": 100, "output_tokens": 50},
                        "content": [
                            {
                                "type": "tool_use",
                                "name": "Edit",
                                "input": {
                                    "file_path": "/test.py",
                                    "old_string": "line1\nline2",
                                    "new_string": "line1\nline2\nline3",
                                },
                            }
                        ],
                    },
                },
            ]
            for line in transcript_data:
                f.write(json.dumps(line) + "\n")
            f.flush()
            temp_path = f.name

        try:
            result = parse_transcript_for_usage(temp_path)
            assert result is not None
            assert result.code_metrics is not None
            assert result.code_metrics["linesAdded"] == 3
            assert result.code_metrics["linesDeleted"] == 2
            assert result.code_metrics["filesModified"] == 1
        finally:
            os.unlink(temp_path)

    def test_parse_transcript_with_multiedit_tool(self):
        """Test parsing transcript with MultiEdit tool."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".jsonl", delete=False) as f:
            transcript_data = [
                {
                    "timestamp": "2024-01-01T00:00:00Z",
                    "type": "assistant",
                    "message": {
                        "model": "claude-sonnet-4",
                        "usage": {"input_tokens": 100, "output_tokens": 50},
                        "content": [
                            {
                                "type": "tool_use",
                                "name": "MultiEdit",
                                "input": {
                                    "file_path": "/test.py",
                                    "edits": [
                                        {"old_string": "line1", "new_string": "lineA"},
                                        {"old_string": "line2\nline3", "new_string": "lineB"},
                                    ],
                                },
                            }
                        ],
                    },
                },
            ]
            for line in transcript_data:
                f.write(json.dumps(line) + "\n")
            f.flush()
            temp_path = f.name

        try:
            result = parse_transcript_for_usage(temp_path)
            assert result is not None
            assert result.code_metrics is not None
            assert result.code_metrics["linesAdded"] == 2
            assert result.code_metrics["linesDeleted"] == 3
        finally:
            os.unlink(temp_path)

    def test_parse_transcript_no_tokens(self):
        """Test parsing transcript with no token usage returns None."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".jsonl", delete=False) as f:
            transcript_data = [
                {"timestamp": "2024-01-01T00:00:00Z", "type": "user", "message": {}},
            ]
            for line in transcript_data:
                f.write(json.dumps(line) + "\n")
            f.flush()
            temp_path = f.name

        try:
            result = parse_transcript_for_usage(temp_path)
            assert result is None
        finally:
            os.unlink(temp_path)

    def test_parse_transcript_invalid_json(self):
        """Test parsing transcript with invalid JSON lines."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".jsonl", delete=False) as f:
            f.write('{"valid": "json"}\n')
            f.write("invalid json line\n")
            f.write('{"also": "valid"}\n')
            f.flush()
            temp_path = f.name

        try:
            result = parse_transcript_for_usage(temp_path)
            # Should skip invalid lines
            assert result is None  # No usage data in valid lines
        finally:
            os.unlink(temp_path)


class TestParseSessionData:
    """Test session data parsing."""

    def test_parse_session_data_valid(self):
        """Test parsing valid session data."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".jsonl", delete=False) as f:
            transcript_data = [
                {"timestamp": "2024-01-01T00:00:00Z", "type": "user", "message": {"model": "claude-opus-4-5"}},
                {
                    "timestamp": "2024-01-01T00:01:00Z",
                    "type": "assistant",
                    "message": {
                        "model": "claude-opus-4-5",
                        "usage": {"input_tokens": 1000, "output_tokens": 500},
                    },
                },
            ]
            for line in transcript_data:
                f.write(json.dumps(line) + "\n")
            f.flush()
            temp_path = f.name

        try:
            session_data = {
                "session_id": "test_session",
                "transcript_path": temp_path,
                "cwd": "/Users/test/project",
            }

            result = parse_session_data(session_data)
            assert result is not None
            assert result["input_tokens"] == 1000
            assert result["output_tokens"] == 500
            assert result["session_id"] == "test_session"
            assert result["model_name"] == "claude-opus-4-5"
            assert "ended_at" in result
        finally:
            os.unlink(temp_path)

    def test_parse_session_data_missing_transcript(self):
        """Test parsing session data without transcript path."""
        session_data = {"session_id": "test"}
        result = parse_session_data(session_data)
        assert result is None

    def test_parse_session_data_zero_tokens(self):
        """Test parsing session data with zero tokens returns None."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".jsonl", delete=False) as f:
            transcript_data = [
                {"timestamp": "2024-01-01T00:00:00Z", "type": "user", "message": {}},
            ]
            for line in transcript_data:
                f.write(json.dumps(line) + "\n")
            f.flush()
            temp_path = f.name

        try:
            session_data = {
                "session_id": "test",
                "transcript_path": temp_path,
                "cwd": "/test",
            }

            result = parse_session_data(session_data)
            assert result is None
        finally:
            os.unlink(temp_path)


class TestParseTranscriptToSubmission:
    """Test parsing transcript to SessionSubmission."""

    def test_parse_transcript_to_submission(self):
        """Test converting transcript to SessionSubmission."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".jsonl", delete=False) as f:
            # Create a transcript file in a temporary directory structure
            transcript_data = [
                {"timestamp": "2024-01-01T00:00:00Z", "type": "user", "message": {"model": "claude-opus-4-5"}},
                {
                    "timestamp": "2024-01-01T00:01:00Z",
                    "type": "assistant",
                    "message": {
                        "model": "claude-opus-4-5",
                        "usage": {"input_tokens": 1000, "output_tokens": 500},
                    },
                },
            ]
            for line in transcript_data:
                f.write(json.dumps(line) + "\n")
            f.flush()
            temp_path = Path(f.name)

        try:
            result = parse_transcript_to_submission(temp_path)
            assert result is not None
            assert isinstance(result, SessionSubmission)
            assert result.input_tokens == 1000
            assert result.output_tokens == 500
            assert result.model_name == "claude-opus-4-5"
            assert result.session_hash is not None
        finally:
            os.unlink(temp_path)

    def test_parse_transcript_to_submission_no_tokens(self):
        """Test parsing transcript with no tokens returns None."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".jsonl", delete=False) as f:
            f.write('{"timestamp": "2024-01-01T00:00:00Z", "type": "user"}\n')
            f.flush()
            temp_path = Path(f.name)

        try:
            result = parse_transcript_to_submission(temp_path)
            assert result is None
        finally:
            os.unlink(temp_path)


class TestSubmitSessionHook:
    """Test session submission hook."""

    def test_submit_session_not_registered(self):
        """Test hook returns not registered when no credentials."""
        with patch.object(RankConfig, "has_credentials", return_value=False):
            session_data = {"session_id": "test"}
            result = submit_session_hook(session_data)
            assert result["success"] is False
            assert "Not registered" in result["message"]

    @patch.object(RankConfig, "has_credentials", return_value=True)
    @patch("moai_adk.rank.hook.RankClient")
    def test_submit_session_success(self, mock_client_class, mock_has_creds):
        """Test successful session submission."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".jsonl", delete=False) as f:
            transcript_data = [
                {"timestamp": "2024-01-01T00:00:00Z", "type": "user", "message": {"model": "claude-opus-4-5"}},
                {
                    "timestamp": "2024-01-01T00:01:00Z",
                    "type": "assistant",
                    "message": {
                        "model": "claude-opus-4-5",
                        "usage": {"input_tokens": 1000, "output_tokens": 500},
                    },
                },
            ]
            for line in transcript_data:
                f.write(json.dumps(line) + "\n")
            f.flush()
            temp_path = f.name

        try:
            mock_client = Mock()
            mock_client.compute_session_hash.return_value = "test_hash"
            mock_client.submit_session.return_value = {"sessionId": "sess_123", "message": "Submitted"}
            mock_client_class.return_value = mock_client

            session_data = {
                "session_id": "test",
                "transcript_path": temp_path,
                "cwd": "/test",
            }

            result = submit_session_hook(session_data)
            assert result["success"] is True
            assert result["session_id"] == "sess_123"
        finally:
            os.unlink(temp_path)

    def test_submit_session_no_token_usage(self):
        """Test hook with no token usage."""
        with patch.object(RankConfig, "has_credentials", return_value=True):
            session_data = {
                "session_id": "test",
                "transcript_path": "/nonexistent",
            }
            result = submit_session_hook(session_data)
            assert result["success"] is False
            assert "No token usage" in result["message"]


class TestHookScriptGeneration:
    """Test hook script generation."""

    def test_create_hook_script(self):
        """Test hook script generation."""
        script = create_hook_script()
        assert "submit_session_hook" in script
        assert "sys.stdin.read" in script
        assert "json.loads" in script

    def test_create_global_hook_script(self):
        """Test global hook script generation."""
        script = create_global_hook_script()
        assert "submit_session_hook" in script
        assert "is_project_excluded" in script


class TestRankConfig:
    """Test rank configuration functions."""

    def test_load_rank_config_default(self):
        """Test loading default config when file doesn't exist."""
        with patch("moai_adk.rank.hook.Path.home", return_value=Path("/nonexistent")):
            config = load_rank_config()
            assert config["enabled"] is True
            assert config["exclude_projects"] == []

    def test_load_rank_config_from_file(self):
        """Test loading config from file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_dir = Path(tmpdir) / ".moai" / "rank"
            config_dir.mkdir(parents=True)
            config_file = config_dir / "config.yaml"

            config_file.write_text("rank:\n  enabled: false\n  exclude_projects:\n    - /test/project\n")

            with patch("moai_adk.rank.hook.Path.home", return_value=Path(tmpdir)):
                config = load_rank_config()
                assert config["enabled"] is False
                assert "/test/project" in config["exclude_projects"]

    def test_is_project_excluded_default(self):
        """Test project exclusion with default config."""
        with patch("moai_adk.rank.hook.Path.home", return_value=Path("/nonexistent")):
            assert is_project_excluded("/test/project") is False

    def test_is_project_excluded_disabled(self):
        """Test project exclusion when disabled."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_dir = Path(tmpdir) / ".moai" / "rank"
            config_dir.mkdir(parents=True)
            config_file = config_dir / "config.yaml"
            config_file.write_text("rank:\n  enabled: false\n")

            with patch("moai_adk.rank.hook.Path.home", return_value=Path(tmpdir)):
                assert is_project_excluded("/test/project") is True

    def test_is_project_excluded_pattern(self):
        """Test project exclusion with pattern."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_dir = Path(tmpdir) / ".moai" / "rank"
            config_dir.mkdir(parents=True)
            config_file = config_dir / "config.yaml"
            config_file.write_text("rank:\n  exclude_projects:\n    - /test/*\n")

            with patch("moai_adk.rank.hook.Path.home", return_value=Path(tmpdir)):
                assert is_project_excluded("/test/project") is True

    def test_save_rank_config(self):
        """Test saving rank config."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.rank.hook.Path.home", return_value=Path(tmpdir)):
                result = save_rank_config({"enabled": False, "exclude_projects": []})
                assert result is True

                config_file = Path(tmpdir) / ".moai" / "rank" / "config.yaml"
                assert config_file.exists()

    def test_add_project_exclusion(self):
        """Test adding project to exclusion list."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.rank.hook.Path.home", return_value=Path(tmpdir)):
                result = add_project_exclusion("/test/project")
                assert result is True

                config = load_rank_config()
                assert "/test/project" in config["exclude_projects"]

    def test_remove_project_exclusion(self):
        """Test removing project from exclusion list."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.rank.hook.Path.home", return_value=Path(tmpdir)):
                # First add it
                add_project_exclusion("/test/project")

                # Then remove it
                result = remove_project_exclusion("/test/project")
                assert result is True

                config = load_rank_config()
                assert "/test/project" not in config["exclude_projects"]

class TestIsDuplicateError:
    """Test _is_duplicate_error function."""

    def test_duplicate_keyword(self):
        """Test detection of 'duplicate' keyword."""
        assert _is_duplicate_error("duplicate session")
        assert _is_duplicate_error("Duplicate entry")
        assert _is_duplicate_error("DUPLICATE")

    def test_already_keyword(self):
        """Test detection of 'already' keyword."""
        assert _is_duplicate_error("already recorded")
        assert _is_duplicate_error("Already exists")
        assert _is_duplicate_error("ALREADY")

    def test_exists_keyword(self):
        """Test detection of 'exists' keyword."""
        assert _is_duplicate_error("session exists")
        assert _is_duplicate_error("Already exists")

    def test_conflict_keyword(self):
        """Test detection of 'conflict' keyword."""
        assert _is_duplicate_error("conflict detected")
        assert _is_duplicate_error("Conflict error")

    def test_previously_recorded_keyword(self):
        """Test detection of 'previously recorded' keyword."""
        assert _is_duplicate_error("previously recorded")
        assert _is_duplicate_error("Previously recorded session")

    def test_multilingual_korean(self):
        """Test detection of Korean '중복' keyword."""
        assert _is_duplicate_error("중복된 세션")
        assert _is_duplicate_error("Duplicate: 중복")

    def test_multilingual_chinese(self):
        """Test detection of Chinese '重複' keyword."""
        assert _is_duplicate_error("重複的会话")
        assert _is_duplicate_error("Duplicate: 重複")

    def test_empty_string(self):
        """Test handling of empty string."""
        assert not _is_duplicate_error("")
        assert not _is_duplicate_error(None)

    def test_non_duplicate_errors(self):
        """Test that non-duplicate errors return False."""
        assert not _is_duplicate_error("network error")
        assert not _is_duplicate_error("validation failed")
        assert not _is_duplicate_error("server unavailable")
        assert not _is_duplicate_error("rate limit exceeded")
