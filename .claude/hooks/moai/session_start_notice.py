#!/usr/bin/env python3
"""
MoAI-ADK Session Start Notice Hook - Optimized v0.2.0
Minimal session start notification with core functionality preserved.

@REQ:HOOK-SESSION-START-001
@FEATURE:SESSION-NOTICE-OPT
@TEST:UNIT-SESSION-SIZE
"""

import json
import os
from pathlib import Path
from typing import Dict, Any, Optional


class SessionNotifier:
    """Optimized MoAI-ADK session start notification system

    @FEATURE:SESSION-NOTICE-OPT
    Reduced from 2,133 to ~200 lines while preserving core functions:
    - MoAI project initialization detection
    - Development guide violation checking
    - Critical configuration missing alerts
    """

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_config_path = project_root / ".moai" / "config.json"

    def get_project_status(self) -> Dict[str, Any]:
        """Get essential project status information

        @FEATURE:PROJECT-STATUS-OPT
        Returns minimal status needed by agents and tests
        """
        return {
            "project_name": self.project_root.name,
            "moai_version": self.get_moai_version(),
            "initialized": self.is_moai_project(),
            "constitution_status": self.check_constitution_status(),
            "pipeline_stage": self.get_current_pipeline_stage(),
        }

    def is_moai_project(self) -> bool:
        """Check if MoAI project is initialized

        @FEATURE:MOAI-PROJECT-DETECTION
        Essential for project status detection
        """
        required_paths = [
            ".moai",
            ".claude/commands/moai",
        ]
        return all((self.project_root / path).exists() for path in required_paths)

    def check_constitution_status(self) -> Optional[Dict[str, Any]]:
        """Check development guide violations

        @FEATURE:DEV-GUIDE-VIOLATIONS
        Essential for maintaining development standards
        """
        if not self.is_moai_project():
            return {"status": "not_initialized", "violations": []}

        # Check for critical missing files
        critical_files = [
            ".moai/memory/development-guide.md",
            "CLAUDE.md"
        ]

        violations = []
        for file_path in critical_files:
            if not (self.project_root / file_path).exists():
                violations.append(f"Missing critical file: {file_path}")

        return {
            "status": "ok" if not violations else "violations_found",
            "violations": violations
        }

    def get_moai_version(self) -> str:
        """Get MoAI-ADK version"""
        try:
            if self.moai_config_path.exists():
                with open(self.moai_config_path) as f:
                    config = json.load(f)
                    return config.get("version", "unknown")
        except Exception:
            pass
        return "unknown"

    def get_current_pipeline_stage(self) -> str:
        """Get current pipeline stage

        @FEATURE:PIPELINE-STAGE-DETECTION
        """
        # Simple heuristic based on directory contents
        specs_dir = self.project_root / ".moai" / "specs"
        if specs_dir.exists() and any(specs_dir.glob("*.md")):
            return "implementation"
        elif self.is_moai_project():
            return "specification"
        else:
            return "initialization"


def main():
    """Main entry point for Claude Code hook system"""
    try:
        project_root = Path(os.getcwd())
        notifier = SessionNotifier(project_root)

        if notifier.is_moai_project():
            status = notifier.get_project_status()
            constitution = status.get("constitution_status", {})

            if constitution.get("violations"):
                print("⚠️  Development guide violations detected:")
                for violation in constitution["violations"]:
                    print(f"   • {violation}")
                print()

            print(f"📋 MoAI Project: {status['project_name']}")
            print(f"🔧 Version: {status['moai_version']}")
            print(f"📍 Stage: {status['pipeline_stage']}")

        else:
            print("💡 Run `/moai:0-project` to initialize MoAI-ADK")

    except Exception as e:
        # Silent failure to avoid breaking Claude Code session
        pass


if __name__ == "__main__":
    main()