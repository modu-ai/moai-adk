"""
@FEATURE:CLI-001 CLI interface modules for MoAI-ADK

@REQ:CLI-INTERFACE-001 → @DESIGN:CLI-ARCHITECTURE-001 → @TASK:CLI-INTEGRATION-001 → @TEST:CLI-COMMANDS-001

@DESIGN:CLI-ARCHITECTURE-001 Clean CLI architecture with command separation
@TASK:CLI-INTEGRATION-001 Command-line interface integration layer

This package contains command-line interface components:
- @TASK:CLI-COMMANDS-001 CLI command implementations
- @TASK:CLI-HELPERS-001 CLI utility functions
- @TASK:CLI-BANNER-001 Banner display functionality
- @TASK:CLI-WIZARD-001 Interactive setup wizard
"""

from .__main__ import main
from .commands import cli
from .wizard import InteractiveWizard

# For backward compatibility
CLICommands = cli

__all__ = [
    "CLICommands",  # Backward compatibility
    "InteractiveWizard",
    "cli",
    "main",
]
