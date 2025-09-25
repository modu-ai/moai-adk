"""
Configuration management for guideline checking.

@FEATURE:QUALITY-CONFIG Configuration system for TRUST 5 principles validation
@DESIGN:SEPARATED-CONFIG-001 Extracted from oversized guideline_checker.py (924 LOC)
"""

import json
from dataclasses import dataclass, asdict
from pathlib import Path
from typing import Dict, Optional, Any

# Optional YAML support (graceful degradation if not available)
try:
    import yaml
    YAML_AVAILABLE = True
except ImportError:
    YAML_AVAILABLE = False
    yaml = None

from .constants import GuidelineLimits
from ...utils.logger import get_logger

logger = get_logger(__name__)


# @DESIGN:CONFIG-002 Configuration management for extensible guidelines
@dataclass
class GuidelineConfig:
    """Configuration for guideline checking with YAML/JSON support."""
    limits: GuidelineLimits
    file_patterns: Dict[str, Any]
    enabled_checks: Dict[str, bool]
    output_format: str = "json"
    parallel_processing: bool = True
    max_workers: Optional[int] = None

    @classmethod
    def from_file(cls, config_path: Path) -> 'GuidelineConfig':
        """
        Load configuration from YAML or JSON file.

        Args:
            config_path: Path to configuration file

        Returns:
            GuidelineConfig instance

        @DESIGN:CONFIG-LOADING-001 File-based configuration loading
        """
        if not config_path.exists():
            raise FileNotFoundError(f"Configuration file not found: {config_path}")

        try:
            with open(config_path, 'r', encoding='utf-8') as file:
                if config_path.suffix in ['.yaml', '.yml']:
                    if not YAML_AVAILABLE:
                        raise ImportError("PyYAML is required for YAML configuration files. "
                                        "Install with: pip install PyYAML")
                    config_data = yaml.safe_load(file)
                elif config_path.suffix == '.json':
                    config_data = json.load(file)
                else:
                    raise ValueError(f"Unsupported config format: {config_path.suffix}")

            # Extract and validate configuration sections
            limits_data = config_data.get('limits', {})
            limits = GuidelineLimits(
                MAX_FUNCTION_LINES=limits_data.get('max_function_lines', 50),
                MAX_FILE_LINES=limits_data.get('max_file_lines', 300),
                MAX_PARAMETERS=limits_data.get('max_parameters', 5),
                MAX_COMPLEXITY=limits_data.get('max_complexity', 10),
                MIN_DOCSTRING_LENGTH=limits_data.get('min_docstring_length', 10),
                MAX_NESTING_DEPTH=limits_data.get('max_nesting_depth', 4)
            )

            return cls(
                limits=limits,
                file_patterns=config_data.get('file_patterns', {}),
                enabled_checks=config_data.get('enabled_checks', {
                    'function_length': True,
                    'file_size': True,
                    'parameter_count': True,
                    'complexity': True
                }),
                output_format=config_data.get('output_format', 'json'),
                parallel_processing=config_data.get('parallel_processing', True),
                max_workers=config_data.get('max_workers')
            )

        except Exception as e:
            logger.error(f"Failed to load configuration from {config_path}: {e}")
            raise

    def to_file(self, config_path: Path) -> None:
        """
        Save configuration to YAML or JSON file.

        Args:
            config_path: Path where to save configuration
        """
        config_data = {
            'limits': asdict(self.limits),
            'file_patterns': self.file_patterns,
            'enabled_checks': self.enabled_checks,
            'output_format': self.output_format,
            'parallel_processing': self.parallel_processing,
            'max_workers': self.max_workers
        }

        try:
            with open(config_path, 'w', encoding='utf-8') as file:
                if config_path.suffix in ['.yaml', '.yml']:
                    if not YAML_AVAILABLE:
                        raise ImportError("PyYAML is required for YAML configuration files. "
                                        "Install with: pip install PyYAML")
                    yaml.safe_dump(config_data, file, default_flow_style=False, indent=2)
                elif config_path.suffix == '.json':
                    json.dump(config_data, file, indent=2)
                else:
                    raise ValueError(f"Unsupported config format: {config_path.suffix}")

            logger.info(f"Configuration saved to {config_path}")

        except Exception as e:
            logger.error(f"Failed to save configuration to {config_path}: {e}")
            raise

    @classmethod
    def create_default(cls) -> 'GuidelineConfig':
        """Create default configuration."""
        return cls(
            limits=GuidelineLimits(),
            file_patterns={
                'include': ['*.py'],
                'exclude': ['*test*', '*__pycache__*', '*.pyc']
            },
            enabled_checks={
                'function_length': True,
                'file_size': True,
                'parameter_count': True,
                'complexity': True
            }
        )