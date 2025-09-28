#!/usr/bin/env python3
# @FEATURE:CLI-ENTRY-011
"""
🗿 MoAI-ADK CLI Main Entry Point

Simple wrapper for the unified CLI entry point.
Supports python -m moai_adk.cli execution.
"""

if __name__ == "__main__":
    from ..cli import main
    main()
