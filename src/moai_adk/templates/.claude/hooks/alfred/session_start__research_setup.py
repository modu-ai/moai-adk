#!/usr/bin/env python3
# @CODE:HOOK-RESEARCH-001 | @SPEC:HOOK-RESEARCH-SETUP-001 | @TEST: tests/hooks/test_research_setup.py
"""SessionStart Hook: Research Environment Setup

Claude Code Event: SessionStart
Purpose: Automatically setup research environment with knowledge base, strategies, and tools
Execution: Triggered automatically when Claude Code session begins

Research Features:
- Previous research state restoration
- Research environment configuration
- Knowledge base JIT loading
- Research strategy initialization
- Resource optimization setup
"""

import json
import os
import sys
import time
from pathlib import Path
from typing import Any, Dict, List, Optional

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
SHARED_DIR = HOOKS_DIR / "shared"
if str(SHARED_DIR) not in sys.path:
    sys.path.insert(0, str(SHARED_DIR))

# Try to import existing modules, provide fallbacks if not available
try:
    from core.timeout import CrossPlatformTimeout
    from core.timeout import TimeoutError as PlatformTimeoutError
except ImportError:
    # Fallback timeout implementation

    class CrossPlatformTimeout:
        def __init__(self, seconds):
            self.seconds = seconds

        def start(self):
            pass

        def cancel(self):
            pass

    class PlatformTimeoutError(Exception):
        pass


def load_research_config() -> Dict[str, Any]:
    """Load research configuration from .moai/config/config.json

    Returns:
        Research configuration dictionary
    """
    try:
        config_file = Path(".moai/config/config.json")
        if config_file.exists():
            import json as json_module
            with open(config_file, 'r', encoding='utf-8') as f:
                config = json_module.load(f)

            research_config = config.get("research", {})
            research_config["tags"] = config.get("tags", {}).get("research_tags", {})
            return research_config
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


def load_previous_session_state() -> Dict[str, Any]:
    """Load previous session research state

    Returns:
        Previous session state dictionary
    """
    try:
        state_file = Path(".moai/memory/last-session-state.json")
        if state_file.exists():
            import json as json_module
            with open(state_file, 'r', encoding='utf-8') as f:
                state = json_module.load(f)

            return state.get("research_state", {})
    except Exception:
        pass

    return {
        "active_strategies": [],
        "knowledge_base": {},
        "research_history": [],
        "performance_metrics": {}
    }


def setup_research_directories() -> None:
    """Setup research-specific directories"""
    try:
        research_dirs = [
            ".moai/research/",
            ".moai/research/strategies/",
            ".moai/research/knowledge/",
            ".moai/research/analysis/",
            ".moai/research/temp/"
        ]

        for dir_path in research_dirs:
            Path(dir_path).mkdir(parents=True, exist_ok=True)

    except Exception as e:
        print(json.dumps({
            "setup_completed": False,
            "error": f"Directory setup failed: {str(e)}",
            "message": "Continuing without research directories"
        }))


def initialize_research_strategies(research_config: Dict[str, Any]) -> List[str]:
    """Initialize research strategies based on configuration

    Args:
        research_config: Research configuration

    Returns:
        List of initialized strategy names
    """
    strategies = []

    # Core research strategies
    core_strategies = [
        "pattern_recognition",
        "cross_domain_analysis",
        "root_cause_analysis",
        "first_principles",
        "probabilistic_thinking",
        "resource_optimization",
        "continuous_learning",
        "systematic_elimination"
    ]

    # Filter strategies based on configuration
    if research_config.get("auto_discovery", True):
        strategies.extend(core_strategies)

    # Add custom strategies if defined
    custom_strategies = research_config.get("custom_strategies", [])
    strategies.extend(custom_strategies)

    return strategies


def load_knowledge_base() -> Dict[str, Any]:
    """Load knowledge base with JIT loading for performance

    Returns:
        Knowledge base dictionary
    """
    knowledge_base = {}

    try:
        # Load existing knowledge files
        knowledge_dir = Path(".moai/research/knowledge/")
        if knowledge_dir.exists():
            for file_path in knowledge_dir.glob("*.json"):
                import json as json_module
                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        knowledge_data = json_module.load(f)
                        knowledge_base[file_path.stem] = knowledge_data
                except Exception:
                    continue

    except Exception:
        pass

    return knowledge_base


def optimize_research_resources() -> Dict[str, Any]:
    """Setup resource optimization for research

    Returns:
        Resource optimization configuration
    """
    return {
        "memory_limit": "256MB",
        "timeout_seconds": 30,
        "parallel_processing": True,
        "cache_enabled": True,
        "compression_enabled": True,
        "streaming_processing": True
    }


def create_research_environment_setup(research_config: Dict[str, Any]) -> Dict[str, Any]:
    """Create comprehensive research environment setup

    Args:
        research_config: Research configuration

    Returns:
        Complete research environment setup
    """
    start_time = time.time()

    # Load previous session state
    previous_state = load_previous_session_state()

    # Initialize research strategies
    active_strategies = initialize_research_strategies(research_config)

    # Load knowledge base
    knowledge_base = load_knowledge_base()

    # Setup resource optimization
    resource_config = optimize_research_resources()

    # Calculate setup metrics
    setup_time = (time.time() - start_time) * 1000

    return {
        "research_setup_completed": True,
        "setup_time_ms": setup_time,
        "previous_state_restored": len(previous_state) > 0,
        "active_strategies": active_strategies,
        "knowledge_base_size": len(knowledge_base),
        "resource_optimization": resource_config,
        "research_categories": research_config.get("research_categories", []),
        "research_patterns": research_config.get("research_patterns", {}),
        "session_id": f"research_{int(time.time())}",
        "start_timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
    }


def display_research_status(setup_result: Dict[str, Any]) -> None:
    """Display research setup status

    Args:
        setup_result: Research setup result
    """
    print("\n🔬 Research Environment Setup")
    print("=" * 50)

    if setup_result["research_setup_completed"]:
        print(f"✅ Setup completed in {setup_result['setup_time_ms']:.1f}ms")
        print(f"📊 Active strategies: {len(setup_result['active_strategies'])}")
        print(f"📚 Knowledge base: {setup_result['knowledge_base_size']} items")
        print(f"🔧 Resource optimization: {setup_result['resource_optimization']['memory_limit']}")

        if setup_result["previous_state_restored"]:
            print("🔄 Previous session state restored")

        print(f"🆔 Session ID: {setup_result['session_id']}")
    else:
        print("❌ Research setup failed")


def main() -> None:
    """Main function"""
    try:
        # Load research configuration
        research_config = load_research_config()

        # Setup research directories
        setup_research_directories()

        # Create research environment setup
        setup_result = create_research_environment_setup(research_config)

        # Check if silent research mode is enabled (via environment variable)
        silent_research = os.environ.get('MOAI_SILENT_RESEARCH', 'false').lower() == 'true'

        # Only proceed with output if neither quiet nor silent research mode is enabled
        if not silent_research:
            quiet_mode = os.environ.get('MOAI_QUIET_SETUP', 'false').lower() == 'true'

            if not quiet_mode:
                # Display status
                display_research_status(setup_result)

                # Output result as JSON
                print(json.dumps(setup_result, ensure_ascii=False, indent=2))

    except Exception as e:
        # Only show errors if not in silent research mode
        if not os.environ.get('MOAI_SILENT_RESEARCH', 'false').lower() == 'true':
            error_response = {
                "research_setup_completed": False,
                "error": f"Hook execution error: {str(e)}",
                "message": "Research setup failed - continuing without research features"
            }
            print(json.dumps(error_response, ensure_ascii=False))


if __name__ == "__main__":
    main()