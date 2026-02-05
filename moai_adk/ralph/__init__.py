# Ralph Engine Module
"""Ralph Engine - Intelligent code quality assurance system for MoAI-ADK.

This module provides the unified Ralph Engine that integrates:
1. LSP (Language Server Protocol) for real-time diagnostics
2. AST-grep for structural pattern matching and security scanning
3. Loop Controller for autonomous feedback loops

The Ralph Engine is the main entry point for intelligent code quality
operations, combining all three technologies into a cohesive system.

Main Components:
    - RalphEngine: Main facade class integrating all components

Usage:
    from moai_adk.ralph import RalphEngine

    engine = RalphEngine(project_root="/path/to/project")

    # Diagnose a file (LSP + AST-grep combined)
    result = await engine.diagnose_file("src/main.py")

    # Start an autonomous feedback loop
    state = engine.start_feedback_loop(promise="Fix all errors")

    # Cancel the loop if needed
    engine.cancel_loop(state.loop_id)

    # Cleanup
    await engine.shutdown()
"""

from moai_adk.ralph.engine import RalphEngine

__all__ = [
    "RalphEngine",
]
