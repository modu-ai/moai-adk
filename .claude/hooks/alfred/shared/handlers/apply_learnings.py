#!/usr/bin/env python3
# @CODE:HOOKS-CLARITY-CLEAN | SPEC: Apply stored learnings handler implementation
"""Apply Learnings Handler: Apply stored learning patterns to optimize current session

This handler implements the learning application mechanism that uses previously
stored patterns and insights to optimize current Claude Code session behavior.
"""

import json
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Any, NamedTuple

from dataclasses import dataclass


@dataclass
class LearningResult:
    """Result of learning application operation"""
    continue_execution: bool = True
    learnings_applied: int = 0
    optimizations_suggested: int = 0
    message: str = ""
    applied_optimizations: List[Dict[str, Any]] = None

    def __post_init__(self):
        if self.applied_optimizations is None:
            self.applied_optimizations = []


def handle_apply_learnings(session_data: Dict[str, Any]) -> LearningResult:
    """Handle application of stored learnings to current session

    Args:
        session_data: Session start data from Claude Code

    Returns:
        LearningResult: Learning application result
    """
    project_root = Path(session_data.get("projectPath", "."))
    moai_dir = project_root / ".moai"

    if not moai_dir.exists():
        return LearningResult(
            message="No .moai directory found - no learnings to apply"
        )

    # Load learning patterns
    patterns_file = moai_dir / "memory" / "daily-patterns.json"
    clone_learnings_file = moai_dir / "memory" / "clone-learnings.json"

    if not patterns_file.exists() and not clone_learnings_file.exists():
        return LearningResult(
            message="No learning data found - first time setup"
        )

    # Apply learnings from both sources
        applied_optimizations = []
        learnings_applied = 0

    # Apply daily pattern learnings
    if patterns_file.exists():
        daily_optimizations = _apply_daily_patterns(patterns_file)
        applied_optimizations.extend(daily_optimizations)
        learnings_applied += len(daily_optimizations)

    # Apply clone learnings
    if clone_learnings_file.exists():
        clone_optimizations = _apply_clone_learnings(clone_learnings_file, session_data)
        applied_optimizations.extend(clone_optimizations)
        learnings_applied += len(clone_optimizations)

    optimizations_suggested = len([opt for opt in applied_optimizations if opt.get("type") == "suggestion"])

    return LearningResult(
        continue_execution=True,
        learnings_applied=learnings_applied,
        optimizations_suggested=optimizations_suggested,
        message=f"Applied {learnings_applied} learning-based optimizations",
        applied_optimizations=applied_optimizations
    )


def _apply_daily_patterns(patterns_file: Path) -> List[Dict[str, Any]]:
    """Apply daily pattern learnings

    Args:
        patterns_file: Path to daily patterns JSON file

    Returns:
        List[Dict[str, Any]]: Applied optimizations
    """
    try:
        patterns_data = json.loads(patterns_file.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, FileNotFoundError):
        return []

    optimizations = []
    insights = patterns_data.get("insights", {})

    # Apply tool usage optimizations
    most_used_tools = insights.get("most_used_tools", [])
    for tool_info in most_used_tools:
        optimizations.append({
            "type": "tool_optimization",
            "tool": tool_info["tool"],
            "action": "prioritize",
            "reason": f"Most frequently used ({tool_info['count']} times)",
            "priority": "high"
        })

    # Apply error prevention
    common_errors = insights.get("common_errors", [])
    for error_info in common_errors:
        optimizations.append({
            "type": "error_prevention",
            "error": error_info["error"],
            "action": "avoid",
            "reason": f"Recurring error ({error_info['count']} occurrences)",
            "priority": "critical"
        })

    # Apply optimization suggestions
    suggestions = insights.get("optimization_suggestions", [])
    for suggestion in suggestions:
        optimizations.append({
            "type": "suggestion",
            "description": suggestion["description"],
            "priority": suggestion["priority"],
            "action": "consider"
        })

    return optimizations


def _apply_clone_learnings(clone_file: Path, session_data: Dict[str, Any]) -> List[Dict[str, Any]]:
    """Apply clone learning patterns

    Args:
        clone_file: Path to clone learnings JSON file
        session_data: Current session data

    Returns:
        List[Dict[str, Any]]: Applied optimizations
    """
    try:
        clone_data = json.loads(clone_file.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, FileNotFoundError):
        return []

    optimizations = []
    learnings = clone_data.get("learnings", {})

    # Apply relevant learnings based on current context
    current_context = _analyze_session_context(session_data)

    for learning_type, learning_list in learnings.items():
        if learning_type == current_context.get("task_type"):
            for learning in learning_list[-5:]:  # Last 5 learnings
                if learning.get("success_rate", 0) > 0.8:  # Successful learnings only
                    optimizations.append({
                        "type": "clone_learning",
                        "learning_type": learning_type,
                        "approach": learning.get("optimized_approach"),
                        "success_rate": learning.get("success_rate"),
                        "action": "apply_pattern"
                    })

    return optimizations


def _analyze_session_context(session_data: Dict[str, Any]) -> Dict[str, Any]:
    """Analyze current session context to determine relevant learnings

    Args:
        session_data: Session start data

    Returns:
        Dict[str, Any]: Session context analysis
    """
    # Simple context analysis - can be expanded
    project_path = session_data.get("projectPath", "")

    context = {
        "task_type": "general",
        "project_type": "unknown",
        "complexity": "medium"
    }

    # Detect project type from path
    if "python" in project_path.lower() or ".py" in project_path:
        context["project_type"] = "python"
    elif "javascript" in project_path.lower() or "js" in project_path:
        context["project_type"] = "javascript"
    elif "typescript" in project_path.lower() or "ts" in project_path:
        context["project_type"] = "typescript"

    return context


def to_dict(self) -> Dict[str, Any]:
    """Convert LearningResult to dictionary for JSON serialization"""
    return {
        "continue": self.continue_execution,
        "hookSpecificOutput": {
            "learnings_applied": self.learnings_applied,
            "optimizations_suggested": self.optimizations_suggested,
            "message": self.message,
            "applied_optimizations": self.applied_optimizations
        }
    }


# Add to_dict method to LearningResult
LearningResult.to_dict = to_dict