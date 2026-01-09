# Loop Controller Module
"""Ralph-style autonomous feedback loop controller for MoAI-ADK.

This module provides the core loop controller functionality that enables
autonomous feedback loops with LSP diagnostics and AST-grep analysis.

Main components:
- LoopState: Data model for tracking loop execution state
- LoopStorage: Persistent storage for loop states
- FeedbackGenerator: Generate feedback from LSP and AST-grep results
- MoAILoopController: Main controller orchestrating the feedback loop

Usage:
    from moai_adk.loop import MoAILoopController, LoopStorage

    storage = LoopStorage(storage_dir=".moai/loop")
    controller = MoAILoopController(storage=storage)

    # Start a new feedback loop
    state = controller.start_loop(promise="Fix all LSP errors")

    # Run feedback iteration
    result = await controller.run_feedback_loop(state)

    # Check completion
    completion = controller.check_completion(state)
"""

from moai_adk.loop.controller import MoAILoopController
from moai_adk.loop.feedback import FeedbackGenerator
from moai_adk.loop.state import (
    ASTIssueSnapshot,
    CompletionResult,
    DiagnosticSnapshot,
    FeedbackResult,
    LoopState,
    LoopStatus,
)
from moai_adk.loop.storage import LoopStorage

__all__ = [
    # Controller
    "MoAILoopController",
    # State models
    "LoopStatus",
    "DiagnosticSnapshot",
    "ASTIssueSnapshot",
    "LoopState",
    "CompletionResult",
    "FeedbackResult",
    # Supporting classes
    "LoopStorage",
    "FeedbackGenerator",
]
