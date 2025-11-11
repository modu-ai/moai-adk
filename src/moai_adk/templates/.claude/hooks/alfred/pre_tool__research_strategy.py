#!/usr/bin/env python3
# @CODE:HOOK-RESEARCH-002 | @SPEC:HOOK-RESEARCH-STRATEGY-001 | @TEST: tests/hooks/test_research_strategy.py
"""Real-time Research Strategy Selection Hook

Automatically selects and optimizes research strategy based on task type at PreToolUse stage.
Sets up research context and allocates resources before task execution.

Features:
- Research strategy selection based on task type analysis
- Automatic research context configuration
- Resource optimization and memory management
- Parallel research processing preparation
- JIT loading of research knowledge

Usage:
    python3 pre_tool__research_strategy.py <tool_name> <tool_args_json>
"""

import os

import json
import sys
import time
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

# Add module path
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

from moai_adk.core.tags.validator import CentralValidationResult, CentralValidator, ValidationConfig
from moai_adk.statusline.version_reader import VersionReader

# Local hook configuration functions
def get_graceful_degradation() -> bool:
    """Return graceful degradation mode configuration"""
    return os.environ.get("MOAI_GRACEFUL_DEGRADATION", "true").lower() == "true"

def load_hook_timeout() -> int:
    """Load hook timeout (in milliseconds)"""
    try:
        return int(os.environ.get("MOAI_HOOK_TIMEOUT_MS", "30000"))
    except ValueError:
        return 30000


def load_research_config() -> Dict[str, Any]:
    """Load research configuration

    Returns:
        Research configuration dictionary
    """
    try:
        config_file = Path(".moai/config/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                config = json.load(f)
                return config.get("tags", {}).get("research_tags", {})
    except Exception:
        pass

    return {
        "auto_discovery": True,
        "pattern_matching": True,
        "cross_reference": True,
        "knowledge_graph": True,
        "research_categories": ["RESEARCH", "ANALYSIS", "KNOWLEDGE", "INSIGHT"],
        "research_patterns": {
            "RESEARCH": ["@RESEARCH:", "research", "investigate", "analyze"],
            "ANALYSIS": ["@ANALYSIS:", "analysis", "evaluate", "assess"],
            "KNOWLEDGE": ["@KNOWLEDGE:", "knowledge", "learn", "pattern"],
            "INSIGHT": ["@INSIGHT:", "insight", "innovate", "optimize"]
        }
    }


def classify_tool_type(tool_name: str, tool_args: Dict[str, Any]) -> str:
    """Classify tool type

    Args:
        tool_name: Tool name
        tool_args: Tool arguments

    Returns:
        Classified tool type
    """
    # Code operation tools
    code_tools = {"Edit", "Write", "MultiEdit", "Read", "Grep", "Glob"}

    # Test-related tools
    test_tools = {"Bash"}

    # Documentation tools
    doc_tools = {"NotebookEdit"}

    # Exploration and analysis tools
    explore_tools = {"Task", "Explore", "Plan", "WebFetch", "WebSearch"}

    # Project management tools
    project_tools = {"Bash(git:*)", "Bash(gh:*)"}

    if tool_name in code_tools:
        return "code"
    elif tool_name in test_tools:
        return "test"
    elif tool_name in doc_tools:
        return "documentation"
    elif tool_name in explore_tools:
        return "exploration"
    elif tool_name in project_tools:
        return "project"
    else:
        return "general"


def get_research_strategies_for_tool(tool_type: str) -> Dict[str, Any]:
    """Select appropriate research strategy for tool type

    Args:
        tool_type: Tool type

    Returns:
        Selected research strategy
    """
    config = load_research_config()

    # Strategy mapping (tool type -> suitable strategy)
    strategy_mapping = {
        "code": {
            "primary_strategies": [
                "pattern_recognition",
                "systematic_elimination",
                "first_principles"
            ],
            "secondary_strategies": [
                "root_cause_analysis",
                "cross_domain_analysis"
            ],
            "focus_areas": [
                "code_quality",
                "architecture_consistency",
                "performance_optimization"
            ],
            "knowledge_categories": ["KNOWLEDGE", "INSIGHT"]
        },
        "test": {
            "primary_strategies": [
                "probabilistic_thinking",
                "systematic_elimination",
                "resource_optimization"
            ],
            "secondary_strategies": [
                "pattern_recognition",
                "continuous_learning"
            ],
            "focus_areas": [
                "test_coverage",
                "edge_case_coverage",
                "performance_testing"
            ],
            "knowledge_categories": ["ANALYSIS", "RESEARCH"]
        },
        "documentation": {
            "primary_strategies": [
                "pattern_recognition",
                "cross_domain_analysis",
                "first_principles"
            ],
            "secondary_strategies": [
                "knowledge_graph_building",
                "continuous_learning"
            ],
            "focus_areas": [
                "clarity",
                "completeness",
                "consistency"
            ],
            "knowledge_categories": ["KNOWLEDGE", "INSIGHT"]
        },
        "exploration": {
            "primary_strategies": [
                "cross_domain_analysis",
                "pattern_recognition",
                "first_principles"
            ],
            "secondary_strategies": [
                "probabilistic_thinking",
                "continuous_learning"
            ],
            "focus_areas": [
                "comprehensive_analysis",
                "deep_understanding",
                "insight_generation"
            ],
            "knowledge_categories": ["RESEARCH", "ANALYSIS", "KNOWLEDGE", "INSIGHT"]
        },
        "project": {
            "primary_strategies": [
                "resource_optimization",
                "systematic_elimination",
                "probabilistic_thinking"
            ],
            "secondary_strategies": [
                "cross_domain_analysis",
                "continuous_learning"
            ],
            "focus_areas": [
                "efficiency",
                "scalability",
                "maintainability"
            ],
            "knowledge_categories": ["ANALYSIS", "INSIGHT"]
        },
        "general": {
            "primary_strategies": [
                "pattern_recognition",
                "resource_optimization",
                "continuous_learning"
            ],
            "secondary_strategies": [
                "systematic_elimination",
                "probabilistic_thinking"
            ],
            "focus_areas": [
                "general_optimization",
                "best_practice_application"
            ],
            "knowledge_categories": ["RESEARCH", "KNOWLEDGE"]
        }
    }

    return strategy_mapping.get(tool_type, strategy_mapping["general"])


def load_jit_knowledge(tool_type: str, focus_areas: List[str]) -> Dict[str, Any]:
    """Just-In-Time knowledge loading

    Args:
        tool_type: Tool type
        focus_areas: Focus areas

    Returns:
        JIT loaded knowledge
    """
    knowledge_base = {}
    knowledge_dir = Path(".moai/research/knowledge/")

    if not knowledge_dir.exists():
        return knowledge_base

    # Load knowledge files matching tool type and focus areas
    relevant_files = []

    # Matching file patterns
    patterns = [
        f"{tool_type}_*.json",
        f"*.json"
    ]

    for pattern in patterns:
        try:
            for file_path in knowledge_dir.glob(pattern):
                if file_path.is_file():
                    relevant_files.append(file_path)
        except Exception:
            continue

    # Load files
    for file_path in relevant_files:
        try:
            import json as json_module
            with open(file_path, 'r', encoding='utf-8') as f:
                knowledge_data = json_module.load(f)

                # Filter only knowledge related to focus areas
                if focus_areas:
                    relevant_knowledge = {}
                    for area in focus_areas:
                        if area.lower() in str(file_path).lower() or area.lower() in json.dumps(knowledge_data).lower():
                            relevant_knowledge.update(knowledge_data)
                    knowledge_base[file_path.stem] = relevant_knowledge
                else:
                    knowledge_base[file_path.stem] = knowledge_data

        except Exception:
            continue

    return knowledge_base


def optimize_resources(tool_type: str, strategies: List[str]) -> Dict[str, Any]:
    """Resource optimization configuration

    Args:
        tool_type: Tool type
        strategies: Selected strategies

    Returns:
        Optimized resource configuration
    """
    # Resource allocation based on tool type
    resource_configs = {
        "code": {
            "memory_limit": "512MB",
            "timeout_seconds": 45,
            "priority": "high",
            "parallel_processing": False,
            "cache_enabled": True
        },
        "test": {
            "memory_limit": "256MB",
            "timeout_seconds": 30,
            "priority": "medium",
            "parallel_processing": True,
            "cache_enabled": True
        },
        "documentation": {
            "memory_limit": "128MB",
            "timeout_seconds": 20,
            "priority": "low",
            "parallel_processing": False,
            "cache_enabled": False
        },
        "exploration": {
            "memory_limit": "1GB",
            "timeout_seconds": 60,
            "priority": "high",
            "parallel_processing": True,
            "cache_enabled": True
        },
        "project": {
            "memory_limit": "256MB",
            "timeout_seconds": 30,
            "priority": "medium",
            "parallel_processing": False,
            "cache_enabled": True
        },
        "general": {
            "memory_limit": "256MB",
            "timeout_seconds": 25,
            "priority": "medium",
            "parallel_processing": False,
            "cache_enabled": True
        }
    }

    config = resource_configs.get(tool_type, resource_configs["general"])

    # Additional optimization based on number of strategies
    if len(strategies) > 3:
        config["memory_limit"] = f"{int(config['memory_limit'].replace('MB', '')) * 1.5:.0f}MB"
        config["parallel_processing"] = True

    return config


def create_research_context(tool_name: str, tool_args: Dict[str, Any],
                          strategy_config: Dict[str, Any],
                          knowledge_base: Dict[str, Any],
                          resource_config: Dict[str, Any]) -> Dict[str, Any]:
    """Create research context

    Args:
        tool_name: Tool name
        tool_args: Tool arguments
        strategy_config: Strategy configuration
        knowledge_base: Knowledge base
        resource_config: Resource configuration

    Returns:
        Research context
    """
    tool_type = classify_tool_type(tool_name, tool_args)

    return {
        "tool_context": {
            "name": tool_name,
            "type": tool_type,
            "args": tool_args
        },
        "research_strategy": strategy_config,
        "knowledge_base": knowledge_base,
        "resource_config": resource_config,
        "session_id": f"research_{int(time.time())}",
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
        "optimization_applied": True
    }


def format_strategy_message(context: Dict[str, Any]) -> str:
    """Generate strategy selection message

    Args:
        context: Research context

    Returns:
        Formatted message
    """
    strategy_names = context["research_strategy"]["primary_strategies"]
    focus_areas = context["research_strategy"]["focus_areas"]
    resource_config = context["resource_config"]

    message = [
        f"🔬 Research strategy selection complete",
        f"📝 Tool: {context['tool_context']['name']} ({context['tool_context']['type']})",
        f"⚡ Activated strategies: {', '.join(strategy_names)}",
        f"🎯 Focus areas: {', '.join(focus_areas)}",
        f"💾 Resource configuration: {resource_config['memory_limit']}, timeout: {resource_config['timeout_seconds']}s"
    ]

    if context["knowledge_base"]:
        message.append(f"📚 JIT knowledge loading: {len(context['knowledge_base'])} items")

    return "\n".join(message)


def main() -> None:
    """Main function"""
    try:
        # Load configuration
        timeout_seconds = load_hook_timeout() / 1000
        graceful_degradation = get_graceful_degradation()

        # Parse arguments
        if len(sys.argv) < 3:
            if os.environ.get("MOAI_SILENT_RESEARCH", "false").lower() != "true":
                print(json.dumps({
                    "research_strategy_selected": False,
                    "error": "Invalid arguments. Usage: python3 pre_tool__research_strategy.py <tool_name> <tool_args_json>"
                }))
            sys.exit(0)

        tool_name = sys.argv[1]
        try:
            tool_args = json.loads(sys.argv[2])
        except json.JSONDecodeError:
            if os.environ.get("MOAI_SILENT_RESEARCH", "false").lower() != "true":
                print(json.dumps({
                    "research_strategy_selected": False,
                    "error": "Invalid tool_args JSON"
                }))
            sys.exit(0)

        # Record start time
        start_time = time.time()

        # Classify tool type
        tool_type = classify_tool_type(tool_name, tool_args)

        # Select research strategy
        strategy_config = get_research_strategies_for_tool(tool_type)

        # Extract focus areas
        focus_areas = strategy_config.get("focus_areas", [])

        # JIT knowledge loading
        knowledge_base = load_jit_knowledge(tool_type, focus_areas)

        # Resource optimization
        resource_config = optimize_resources(tool_type, strategy_config["primary_strategies"])

        # Create research context
        research_context = create_research_context(
            tool_name, tool_args, strategy_config,
            knowledge_base, resource_config
        )

        # Timeout check
        if time.time() - start_time > timeout_seconds:
            timeout_response = {
                "research_strategy_selected": False,
                "error": "Research strategy selection timeout",
                "message": "Research strategy selection failed - proceeding with default strategy",
                "graceful_degradation": graceful_degradation
            }
            if os.environ.get("MOAI_SILENT_RESEARCH", "false").lower() != "true":
                print(json.dumps(timeout_response, ensure_ascii=False))
            sys.exit(0)

        # Calculate execution time
        execution_time_ms = (time.time() - start_time) * 1000

        # Final response
        response = {
            **research_context,
            "execution_time_ms": execution_time_ms,
            "research_strategy_selected": True,
            "message": format_strategy_message(research_context)
        }

        # Performance warning
        timeout_warning_ms = timeout_seconds * 1000 * 0.8
        if execution_time_ms > timeout_warning_ms:
            response["performance_warning"] = f"Strategy selection time exceeded 80% of timeout ({execution_time_ms:.0f}ms / {timeout_warning_ms:.0f}ms)"

        if os.environ.get("MOAI_SILENT_RESEARCH", "false").lower() != "true":
            print(json.dumps(response, ensure_ascii=False, indent=2))

    except Exception as e:
        # Exception handling
        error_response = {
            "research_strategy_selected": False,
            "error": f"Research strategy selection error: {str(e)}",
            "message": "Research strategy selection failed - proceeding with default strategy"
        }

        if graceful_degradation:
            error_response["graceful_degradation"] = True

        if os.environ.get("MOAI_SILENT_RESEARCH", "false").lower() != "true":
            print(json.dumps(error_response, ensure_ascii=False))
        sys.exit(1)


if __name__ == "__main__":
    main()