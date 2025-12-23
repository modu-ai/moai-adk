"""
Unit tests for batch_generate.py module.

Tests cover:
- validate_aspect_ratio(): Batch version (silent)
- validate_resolution(): Batch version (silent)
- build_prompt_with_style(): Style prefix application
- calculate_backoff_delay(): Exponential backoff calculation
- load_config_file(): JSON/YAML file loading
- parse_tasks_from_config(): Task parsing from config dict
- create_tasks_from_prompts(): Task creation from CLI prompts

TDD Approach: RED-GREEN-REFACTOR cycle
"""

import json
from pathlib import Path
from unittest.mock import patch, MagicMock

import pytest

# Import functions to test
from batch_generate import (
    validate_aspect_ratio,
    validate_resolution,
    build_prompt_with_style,
    calculate_backoff_delay,
    load_config_file,
    parse_tasks_from_config,
    create_tasks_from_prompts,
    ImageTask,
    SUPPORTED_ASPECT_RATIOS,
    SUPPORTED_RESOLUTIONS,
    BASE_DELAY,
    MAX_DELAY,
)


class TestBatchValidateAspectRatio:
    """Tests for batch version of validate_aspect_ratio()."""

    @pytest.mark.parametrize("ratio", SUPPORTED_ASPECT_RATIOS)
    def test_supported_ratios_return_unchanged(self, ratio):
        """Supported aspect ratios should be returned unchanged."""
        result = validate_aspect_ratio(ratio)
        assert result == ratio

    @pytest.mark.parametrize("invalid_ratio", ["4:4", "invalid", "", "16:10"])
    def test_unsupported_ratios_return_default_silently(self, invalid_ratio, capsys):
        """Unsupported ratios should return default without printing (batch mode)."""
        result = validate_aspect_ratio(invalid_ratio)
        assert result == "16:9"

        # Batch version should be silent (no warnings)
        captured = capsys.readouterr()
        assert captured.out == ""


class TestBatchValidateResolution:
    """Tests for batch version of validate_resolution()."""

    @pytest.mark.parametrize("resolution,expected", [
        ("1K", "1K"),
        ("2K", "2K"),
        ("4K", "4K"),
        ("1k", "1K"),
        ("2k", "2K"),
        ("4k", "4K"),
    ])
    def test_resolution_normalization(self, resolution, expected):
        """Resolutions should be normalized to uppercase."""
        result = validate_resolution(resolution)
        assert result == expected

    @pytest.mark.parametrize("invalid", ["3K", "8K", "HD", "", "invalid"])
    def test_unsupported_resolutions_return_default(self, invalid, capsys):
        """Unsupported resolutions should return default silently."""
        result = validate_resolution(invalid)
        assert result == "2K"

        # Batch version should be silent
        captured = capsys.readouterr()
        assert captured.out == ""


class TestBatchBuildPromptWithStyle:
    """Tests for batch version of build_prompt_with_style()."""

    def test_prompt_without_style(self):
        """Prompt without style should return original."""
        result = build_prompt_with_style("Test prompt")
        assert result == "Test prompt"

    def test_prompt_with_style(self):
        """Prompt with style should prepend style prefix."""
        result = build_prompt_with_style("Test prompt", "photorealistic")
        assert result == "photorealistic: Test prompt"


class TestBatchCalculateBackoffDelay:
    """Tests for batch version of calculate_backoff_delay()."""

    def test_exponential_growth(self):
        """Delay should grow exponentially."""
        delays = [calculate_backoff_delay(i, jitter=False) for i in range(4)]
        assert delays[0] == BASE_DELAY
        assert delays[1] == BASE_DELAY * 2
        assert delays[2] == BASE_DELAY * 4
        assert delays[3] == BASE_DELAY * 8

    def test_max_delay_cap(self):
        """Delay should not exceed MAX_DELAY."""
        delay = calculate_backoff_delay(100, jitter=False)
        assert delay == MAX_DELAY


class TestLoadConfigFile:
    """Tests for load_config_file() function."""

    def test_load_json_config(self, tmp_path):
        """Should load JSON configuration file."""
        config_data = {
            "defaults": {"resolution": "4K"},
            "images": ["prompt1", "prompt2"]
        }
        config_file = tmp_path / "config.json"
        config_file.write_text(json.dumps(config_data), encoding="utf-8")

        result = load_config_file(str(config_file))

        assert result["defaults"]["resolution"] == "4K"
        assert len(result["images"]) == 2

    def test_load_yaml_config(self, tmp_path):
        """Should load YAML configuration file when PyYAML available."""
        yaml_content = """
defaults:
  resolution: 4K
  style: photorealistic
images:
  - Simple prompt
  - prompt: Complex prompt
    resolution: 2K
"""
        config_file = tmp_path / "config.yaml"
        config_file.write_text(yaml_content, encoding="utf-8")

        try:
            result = load_config_file(str(config_file))
            assert result["defaults"]["resolution"] == "4K"
            assert result["defaults"]["style"] == "photorealistic"
        except ImportError:
            pytest.skip("PyYAML not installed")

    def test_file_not_found_raises_error(self, tmp_path):
        """Should raise FileNotFoundError for missing config."""
        with pytest.raises(FileNotFoundError):
            load_config_file(str(tmp_path / "nonexistent.json"))

    def test_invalid_json_raises_error(self, tmp_path):
        """Should raise error for invalid JSON."""
        config_file = tmp_path / "invalid.json"
        config_file.write_text("{ invalid json }", encoding="utf-8")

        with pytest.raises(json.JSONDecodeError):
            load_config_file(str(config_file))


class TestParseTasksFromConfig:
    """Tests for parse_tasks_from_config() function."""

    def test_simple_string_prompts(self, tmp_path, sample_config):
        """Should parse simple string prompts."""
        config = {"images": ["Prompt 1", "Prompt 2", "Prompt 3"]}
        output_dir = str(tmp_path)

        tasks = parse_tasks_from_config(config, output_dir)

        assert len(tasks) == 3
        assert tasks[0].prompt == "Prompt 1"
        assert tasks[0].task_id == 1
        assert "image_001.png" in tasks[0].output_path

    def test_dict_prompts_with_options(self, tmp_path):
        """Should parse dict prompts with custom options."""
        config = {
            "images": [
                {
                    "prompt": "Custom prompt",
                    "filename": "custom.png",
                    "resolution": "4K",
                    "aspect_ratio": "1:1",
                    "style": "watercolor"
                }
            ]
        }
        output_dir = str(tmp_path)

        tasks = parse_tasks_from_config(config, output_dir)

        assert len(tasks) == 1
        assert tasks[0].prompt == "Custom prompt"
        assert tasks[0].resolution == "4K"
        assert tasks[0].aspect_ratio == "1:1"
        assert tasks[0].style == "watercolor"
        assert "custom.png" in tasks[0].output_path

    def test_global_defaults_applied(self, tmp_path):
        """Should apply global defaults to tasks."""
        config = {
            "defaults": {
                "style": "photorealistic",
                "resolution": "4K",
                "aspect_ratio": "21:9"
            },
            "images": ["Simple prompt"]
        }
        output_dir = str(tmp_path)

        tasks = parse_tasks_from_config(config, output_dir)

        assert tasks[0].style == "photorealistic"
        assert tasks[0].resolution == "4K"
        assert tasks[0].aspect_ratio == "21:9"

    def test_task_overrides_defaults(self, tmp_path):
        """Task-specific settings should override defaults."""
        config = {
            "defaults": {"resolution": "2K", "style": "default_style"},
            "images": [
                {
                    "prompt": "Override test",
                    "resolution": "4K",
                    "style": "custom_style"
                }
            ]
        }
        output_dir = str(tmp_path)

        tasks = parse_tasks_from_config(config, output_dir)

        assert tasks[0].resolution == "4K"
        assert tasks[0].style == "custom_style"

    def test_empty_prompts_skipped(self, tmp_path):
        """Empty prompts should be skipped."""
        config = {
            "images": [
                "Valid prompt",
                {"prompt": ""},
                {"description": ""}
            ]
        }
        output_dir = str(tmp_path)

        tasks = parse_tasks_from_config(config, output_dir)

        assert len(tasks) == 1
        assert tasks[0].prompt == "Valid prompt"

    def test_adds_png_extension(self, tmp_path):
        """Should add .png extension if missing."""
        config = {
            "images": [
                {"prompt": "Test", "filename": "no_extension"}
            ]
        }
        output_dir = str(tmp_path)

        tasks = parse_tasks_from_config(config, output_dir)

        assert tasks[0].output_path.endswith(".png")

    def test_metadata_preserved(self, tmp_path):
        """Should preserve metadata from config."""
        config = {
            "images": [
                {
                    "prompt": "Test",
                    "metadata": {"category": "test", "priority": 1}
                }
            ]
        }
        output_dir = str(tmp_path)

        tasks = parse_tasks_from_config(config, output_dir)

        assert tasks[0].metadata["category"] == "test"
        assert tasks[0].metadata["priority"] == 1


class TestCreateTasksFromPrompts:
    """Tests for create_tasks_from_prompts() function."""

    def test_creates_tasks_from_prompts_list(self, tmp_path, sample_prompts):
        """Should create ImageTask for each prompt."""
        output_dir = str(tmp_path)

        tasks = create_tasks_from_prompts(sample_prompts, output_dir)

        assert len(tasks) == len(sample_prompts)
        for i, task in enumerate(tasks):
            assert task.prompt == sample_prompts[i]
            assert task.task_id == i + 1

    def test_applies_shared_settings(self, tmp_path):
        """Should apply shared settings to all tasks."""
        prompts = ["Prompt 1", "Prompt 2"]
        output_dir = str(tmp_path)

        tasks = create_tasks_from_prompts(
            prompts,
            output_dir,
            style="shared_style",
            resolution="4K",
            aspect_ratio="1:1"
        )

        for task in tasks:
            assert task.style == "shared_style"
            assert task.resolution == "4K"
            assert task.aspect_ratio == "1:1"

    def test_generates_sequential_filenames(self, tmp_path):
        """Should generate sequential filenames."""
        prompts = ["A", "B", "C"]
        output_dir = str(tmp_path)

        tasks = create_tasks_from_prompts(prompts, output_dir)

        assert "image_001.png" in tasks[0].output_path
        assert "image_002.png" in tasks[1].output_path
        assert "image_003.png" in tasks[2].output_path

    def test_uses_default_settings_when_not_specified(self, tmp_path):
        """Should use default settings when not specified."""
        prompts = ["Test"]
        output_dir = str(tmp_path)

        tasks = create_tasks_from_prompts(prompts, output_dir)

        assert tasks[0].resolution == "2K"
        assert tasks[0].aspect_ratio == "16:9"
        assert tasks[0].style is None


class TestImageTaskDataclass:
    """Tests for ImageTask dataclass."""

    def test_default_values(self):
        """ImageTask should have correct default values."""
        task = ImageTask(
            prompt="Test prompt",
            output_path="/output/test.png"
        )

        assert task.aspect_ratio == "16:9"
        assert task.resolution == "2K"
        assert task.style is None
        assert task.enable_grounding is False
        assert task.task_id == 0
        assert task.metadata == {}

    def test_custom_values(self):
        """ImageTask should accept custom values."""
        task = ImageTask(
            prompt="Custom",
            output_path="/path/to/file.png",
            aspect_ratio="1:1",
            resolution="4K",
            style="photorealistic",
            enable_grounding=True,
            task_id=42,
            metadata={"key": "value"}
        )

        assert task.prompt == "Custom"
        assert task.aspect_ratio == "1:1"
        assert task.resolution == "4K"
        assert task.style == "photorealistic"
        assert task.enable_grounding is True
        assert task.task_id == 42
        assert task.metadata["key"] == "value"
