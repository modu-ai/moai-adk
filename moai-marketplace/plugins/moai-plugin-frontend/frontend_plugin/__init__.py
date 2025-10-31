"""
Frontend Plugin - React/Vue Component Setup for Alfred Framework

@CODE:FRONTEND-PLUGIN-INIT-001:INIT
"""

__version__ = "1.0.0-dev"
__author__ = "GOOS🪿"
__license__ = "MIT"

from .commands import init_react, setup_state, setup_testing

__all__ = ["init_react", "setup_state", "setup_testing"]
