"""Core module: primary business logic"""

# Export context manager for CLAUDE.md @path imports
from moai_adk.core.context_manager import ClaudeMDImporter, process_claude_md_imports

# Export rules loader for .claude/rules/ directory
from moai_adk.core.rules_loader import Rule, RulesLoader, load_all_rules, load_rules_for_file

__all__ = [
    # Context manager
    "ClaudeMDImporter",
    "process_claude_md_imports",
    # Rules loader
    "Rule",
    "RulesLoader",
    "load_all_rules",
    "load_rules_for_file",
]
