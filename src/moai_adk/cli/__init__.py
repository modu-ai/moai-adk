"""
CLI interface modules for MoAI-ADK.

This package contains command-line interface components:
- commands: CLI command implementations
- helpers: CLI utility functions
- banner: Banner display functionality
- wizard: Interactive setup wizard
"""

from .commands import cli
from .wizard import InteractiveWizard
from .__main__ import main

# For backward compatibility
CLICommands = cli

__all__ = [
    'cli',
    'CLICommands',  # Backward compatibility
    'InteractiveWizard',
    'main'
]