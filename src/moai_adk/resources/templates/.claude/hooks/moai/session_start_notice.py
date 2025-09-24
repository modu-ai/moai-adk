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
            "spec_progress": self.get_spec_progress(),
        }

    def get_spec_progress(self) -> Dict[str, Any]:
        """Get SPEC progress information"""
        specs_dir = self.project_root / ".moai" / "specs"
        if not specs_dir.exists():
            return {"total": 0, "completed": 0}

        spec_dirs = [d for d in specs_dir.iterdir() if d.is_dir() and d.name.startswith("SPEC-")]
        total_specs = len(spec_dirs)

        # Simple heuristic: completed if has both spec.md and plan.md
        completed = 0
        for spec_dir in spec_dirs:
            if (spec_dir / "spec.md").exists() and (spec_dir / "plan.md").exists():
                completed += 1

        return {"total": total_specs, "completed": completed}

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
                    return config.get("project", {}).get("version", "unknown")
        except Exception:
            pass
        return "unknown"

    def get_current_pipeline_stage(self) -> str:
        """Get current pipeline stage

        @FEATURE:PIPELINE-STAGE-DETECTION
        """
        try:
            if self.moai_config_path.exists():
                with open(self.moai_config_path) as f:
                    config = json.load(f)
                    return config.get("pipeline", {}).get("current_stage", "unknown")
        except Exception:
            pass

        # Fallback heuristic
        specs_dir = self.project_root / ".moai" / "specs"
        if specs_dir.exists() and any(specs_dir.glob("*/spec.md")):
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

            # Enhanced project status display
            spec_progress = status.get("spec_progress", {"total": 0, "completed": 0})

            # Determine if git has uncommitted changes
            import subprocess
            git_status = ""
            try:
                result = subprocess.run(["git", "status", "--porcelain"],
                                      capture_output=True, text=True, timeout=2)
                if result.returncode == 0:
                    changes = len([line for line in result.stdout.strip().split('\n') if line])
                    if changes > 0:
                        git_status = f" ({changes} 변경사항)"
            except:
                pass

            print(f"🗿 MoAI-ADK 프로젝트: {status['project_name']}")
            branch_info = get_current_branch()
            if git_status:
                print(f"🌿 현재 브랜치: {branch_info} ({get_latest_commit()[:7]} {get_commit_message()[:50]}...)")
                print(f"📝 변경사항: {changes}개 파일")
            else:
                print(f"🌿 현재 브랜치: {branch_info} ({get_latest_commit()[:7]} {get_commit_message()[:50]}...)")
            print(f"📝 SPEC 진행률: {spec_progress['completed']}/{spec_progress['total']} (미완료 {spec_progress['total'] - spec_progress['completed']}개)")
            print("✅ 통합 체크포인트 시스템 사용 가능")

        else:
            print("💡 Run `/moai:0-project` to initialize MoAI-ADK")

    except Exception as e:
        # Silent failure to avoid breaking Claude Code session
        pass


def get_current_branch() -> str:
    """Get current git branch name"""
    try:
        import subprocess
        result = subprocess.run(["git", "rev-parse", "--abbrev-ref", "HEAD"],
                              capture_output=True, text=True, timeout=2)
        if result.returncode == 0:
            return result.stdout.strip()
    except:
        pass
    return "unknown"


def get_latest_commit() -> str:
    """Get latest commit hash"""
    try:
        import subprocess
        result = subprocess.run(["git", "rev-parse", "HEAD"],
                              capture_output=True, text=True, timeout=2)
        if result.returncode == 0:
            return result.stdout.strip()
    except:
        pass
    return "unknown"


def get_commit_message() -> str:
    """Get latest commit message"""
    try:
        import subprocess
        result = subprocess.run(["git", "log", "-1", "--pretty=%s"],
                              capture_output=True, text=True, timeout=2)
        if result.returncode == 0:
            return result.stdout.strip()
    except:
        pass
    return "No commit message"


if __name__ == "__main__":
    main()