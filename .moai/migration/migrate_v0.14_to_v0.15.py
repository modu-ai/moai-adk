#!/usr/bin/env python3
"""
MoAI-ADK Configuration Migration Script
v0.14.0 → v0.15.2 설정 업그레이드

변경사항:
1. conversation_language 위치 이동: project → language 섹션
2. 모든 version 필드 업데이트: 0.15.2
3. env 섹션 제거 (사용 안 함)
4. Hook 설명 필드 추가
5. hooks 설정 섹션 추가 (config.json)
"""

import json
from pathlib import Path
from typing import Any, Dict


def migrate_config_json(config_path: Path) -> None:
    """마이그레이션: .moai/config.json"""
    print(f"Migrating {config_path}...")

    config = json.loads(config_path.read_text(encoding="utf-8"))

    # 1. language 섹션 생성 및 conversation_language 이동
    if "project" in config and "conversation_language" in config["project"]:
        if "language" not in config:
            config["language"] = {}

        config["language"]["conversation_language"] = config["project"].pop("conversation_language")
        config["language"]["conversation_language_name"] = config["project"].pop("conversation_language_name")
        print("  ✅ Moved conversation_language to language section")

    # 2. moai.version 업데이트
    if "moai" in config:
        old_version = config["moai"].get("version")
        config["moai"]["version"] = "0.15.2"
        config["moai"]["update_check_frequency"] = "daily"
        config["moai"]["version_check"] = {
            "enabled": True,
            "cache_ttl_hours": 24
        }
        print(f"  ✅ Updated moai.version: {old_version} → 0.15.2")

    # 3. project.template_version 업데이트
    if "project" in config:
        config["project"]["template_version"] = "0.15.2"
        print("  ✅ Updated project.template_version: 0.15.2")

    # 4. hooks 섹션 추가
    if "hooks" not in config:
        config["hooks"] = {
            "timeout_ms": 5000,
            "graceful_degradation": True,
            "notes": "Hook execution timeout (milliseconds). Set graceful_degradation to true to continue even if a hook fails."
        }
        print("  ✅ Added hooks configuration section")

    # JSON 저장 (원본 포맷 유지)
    config_path.write_text(
        json.dumps(config, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8"
    )
    print(f"  ✅ Saved {config_path}")


def migrate_settings_json(settings_path: Path) -> None:
    """마이그레이션: .claude/settings.json"""
    print(f"Migrating {settings_path}...")

    settings = json.loads(settings_path.read_text(encoding="utf-8"))

    # 1. env 섹션 제거
    if "env" in settings:
        env = settings.pop("env")
        print(f"  ✅ Removed env section: {env}")

    # 2. Hook 설명 필드 추가
    hooks_added = False
    if "hooks" in settings:
        descriptions = {
            "SessionStart": "Display project info and language configuration at session start",
            "PreToolUse": "Create automatic checkpoint before file modifications",
            "UserPromptSubmit": "Load documentation on-demand based on user intent",
            "SessionEnd": "Clean up session state and temporary files",
            "PostToolUse": "Log file changes for audit and tracking"
        }

        for hook_name, description in descriptions.items():
            if hook_name in settings["hooks"]:
                for hook in settings["hooks"][hook_name]:
                    if "hooks" in hook and len(hook["hooks"]) > 0:
                        hook["hooks"][0]["description"] = description
                        hooks_added = True

        if hooks_added:
            print("  ✅ Added Hook descriptions")

    # 3. permissions 최적화 (Git 명령 세분화)
    if "permissions" in settings and "allow" in settings["permissions"]:
        permissions = settings["permissions"]

        # git:* 제거
        if "Bash(git:*)" in permissions["allow"]:
            permissions["allow"].remove("Bash(git:*)")

            # 구체적인 git 읽기 명령 추가
            git_read_commands = [
                "Bash(git status:*)",
                "Bash(git log:*)",
                "Bash(git diff:*)",
                "Bash(git branch:*)",
                "Bash(git show:*)",
                "Bash(git remote:*)",
                "Bash(git tag:*)",
                "Bash(git config:*)"
            ]

            # KillShell 다음에 git 명령 삽입
            kill_shell_index = permissions["allow"].index("Bash(KillShell:*)") if "Bash(KillShell:*)" in permissions["allow"] else 12
            for i, cmd in enumerate(git_read_commands):
                if cmd not in permissions["allow"]:
                    permissions["allow"].insert(kill_shell_index + 1 + i, cmd)

            print("  ✅ Optimized permissions: git:* → specific git commands")

    # JSON 저장
    settings_path.write_text(
        json.dumps(settings, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8"
    )
    print(f"  ✅ Saved {settings_path}")


def main():
    """메인 마이그레이션 함수"""
    print("=" * 60)
    print("MoAI-ADK Configuration Migration v0.14.0 → v0.15.2")
    print("=" * 60)
    print()

    # 파일 경로
    project_root = Path.cwd()
    moai_config = project_root / ".moai" / "config.json"
    claude_settings = project_root / ".claude" / "settings.json"

    # 백업 생성
    print("Creating backups...")
    moai_config_backup = moai_config.with_suffix(".json.backup")
    claude_settings_backup = claude_settings.with_suffix(".json.backup")

    if moai_config.exists():
        moai_config.read_text().encode()
        moai_config_backup.write_text(moai_config.read_text(encoding="utf-8"), encoding="utf-8")
        print(f"  ✅ Backed up {moai_config_backup}")

    if claude_settings.exists():
        claude_settings_backup.write_text(claude_settings.read_text(encoding="utf-8"), encoding="utf-8")
        print(f"  ✅ Backed up {claude_settings_backup}")

    print()

    # 마이그레이션 실행
    try:
        if moai_config.exists():
            migrate_config_json(moai_config)

        if claude_settings.exists():
            migrate_settings_json(claude_settings)

        print()
        print("=" * 60)
        print("✅ Migration completed successfully!")
        print("=" * 60)
        print()
        print("📝 Changes made:")
        print("  1. Moved conversation_language to language section")
        print("  2. Updated all version fields to 0.15.2")
        print("  3. Removed unused env section")
        print("  4. Added Hook descriptions")
        print("  5. Optimized permissions (git:* → specific commands)")
        print("  6. Added hooks timeout configuration")
        print()
        print("💾 Backups created (if needed, restore from .backup files)")
        print()

    except Exception as e:
        print()
        print("=" * 60)
        print(f"❌ Migration failed: {e}")
        print("=" * 60)

        # 백업에서 복원
        if moai_config_backup.exists():
            moai_config.write_text(moai_config_backup.read_text(encoding="utf-8"), encoding="utf-8")
            print(f"  Restored {moai_config} from backup")

        if claude_settings_backup.exists():
            claude_settings.write_text(claude_settings_backup.read_text(encoding="utf-8"), encoding="utf-8")
            print(f"  Restored {claude_settings} from backup")

        raise


if __name__ == "__main__":
    main()
