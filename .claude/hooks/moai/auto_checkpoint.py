#!/usr/bin/env python3
"""
Auto Checkpoint System for MoAI-ADK Personal Mode

Automatically creates checkpoints every 5 minutes for personal mode projects.
Integrates with git-manager agent for safe development workflow.
"""

import json
import sys
import time
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Dict


class AutoCheckpointManager:
    """Manages automatic checkpoint creation for personal mode development."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_dir = project_root / ".moai"
        self.config_file = self.moai_dir / "config.json"
        self.checkpoints_dir = self.moai_dir / "checkpoints"
        self.metadata_file = self.checkpoints_dir / "metadata.json"
        self.last_checkpoint_file = self.checkpoints_dir / ".last_checkpoint"

        # Ensure directories exist
        self.checkpoints_dir.mkdir(parents=True, exist_ok=True)

    def load_config(self) -> Dict:
        """Load project configuration."""
        if not self.config_file.exists():
            return {}

        try:
            with open(self.config_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            print(f"⚠️ Failed to load config: {e}")
            return {}

    def is_personal_mode(self) -> bool:
        """Check if project is in personal mode."""
        config = self.load_config()
        return config.get("project", {}).get("mode") == "personal"

    def is_auto_checkpoint_enabled(self) -> bool:
        """Check if auto checkpoint is enabled."""
        config = self.load_config()
        return config.get("git_strategy", {}).get("personal", {}).get("auto_checkpoint", False)

    def get_checkpoint_interval(self) -> int:
        """Get checkpoint interval in seconds."""
        config = self.load_config()
        return config.get("git_strategy", {}).get("personal", {}).get("checkpoint_interval", 300)

    def should_create_checkpoint(self) -> bool:
        """Determine if a checkpoint should be created."""
        # Check mode and settings
        if not self.is_personal_mode() or not self.is_auto_checkpoint_enabled():
            return False

        # Check if we're in a git repository
        if not self.is_git_repository():
            return False

        # Check if there are changes to commit
        if not self.has_uncommitted_changes():
            return False

        # Check time since last checkpoint
        return self.time_since_last_checkpoint() >= self.get_checkpoint_interval()

    def is_git_repository(self) -> bool:
        """Check if current directory is a git repository."""
        try:
            result = subprocess.run(
                ["git", "rev-parse", "--is-inside-work-tree"],
                cwd=self.project_root,
                capture_output=True,
                text=True
            )
            return result.returncode == 0
        except Exception:
            return False

    def has_uncommitted_changes(self) -> bool:
        """Check if there are uncommitted changes."""
        try:
            result = subprocess.run(
                ["git", "status", "--porcelain"],
                cwd=self.project_root,
                capture_output=True,
                text=True
            )
            return bool(result.stdout.strip())
        except Exception:
            return False

    def time_since_last_checkpoint(self) -> float:
        """Get time since last checkpoint in seconds."""
        if not self.last_checkpoint_file.exists():
            return float('inf')

        try:
            with open(self.last_checkpoint_file, 'r') as f:
                last_time = float(f.read().strip())
            return time.time() - last_time
        except Exception:
            return float('inf')

    def generate_checkpoint_id(self) -> str:
        """Generate a unique checkpoint ID."""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        return f"checkpoint_{timestamp}"

    def create_checkpoint(self, message: str = "") -> bool:
        """Create a new checkpoint."""
        try:
            checkpoint_id = self.generate_checkpoint_id()
            timestamp = datetime.now()

            # Add all changes
            subprocess.run(
                ["git", "add", "-A"],
                cwd=self.project_root,
                check=True
            )

            # Create commit
            commit_message = f"🔄 Auto-checkpoint: {timestamp.strftime('%H:%M:%S')}"
            if message:
                commit_message += f"\n\n{message}"

            subprocess.run(
                ["git", "commit", "-m", commit_message],
                cwd=self.project_root,
                check=True
            )

            # Get commit hash
            result = subprocess.run(
                ["git", "rev-parse", "HEAD"],
                cwd=self.project_root,
                capture_output=True,
                text=True,
                check=True
            )
            commit_hash = result.stdout.strip()

            # Create backup branch
            subprocess.run(
                ["git", "branch", checkpoint_id, "HEAD"],
                cwd=self.project_root,
                check=True
            )

            # Save checkpoint metadata
            self.save_checkpoint_metadata(checkpoint_id, commit_hash, message)

            # Update last checkpoint time
            with open(self.last_checkpoint_file, 'w') as f:
                f.write(str(time.time()))

            print(f"💾 Auto-checkpoint created: {checkpoint_id}")
            return True

        except subprocess.CalledProcessError as e:
            print(f"❌ Git command failed: {e}")
            return False
        except Exception as e:
            print(f"❌ Checkpoint creation failed: {e}")
            return False

    def save_checkpoint_metadata(self, checkpoint_id: str, commit_hash: str, message: str):
        """Save checkpoint metadata to JSON file."""
        try:
            # Load existing metadata
            metadata = {"checkpoints": []}
            if self.metadata_file.exists():
                with open(self.metadata_file, 'r', encoding='utf-8') as f:
                    metadata = json.load(f)

            # Get current branch
            result = subprocess.run(
                ["git", "branch", "--show-current"],
                cwd=self.project_root,
                capture_output=True,
                text=True
            )
            current_branch = result.stdout.strip() if result.returncode == 0 else "unknown"

            # Count changed files
            result = subprocess.run(
                ["git", "diff", "--name-only", "HEAD~1"],
                cwd=self.project_root,
                capture_output=True,
                text=True
            )
            files_changed = len(result.stdout.strip().split('\n')) if result.stdout.strip() else 0

            # Add new checkpoint
            new_checkpoint = {
                "id": checkpoint_id,
                "timestamp": datetime.now().isoformat(),
                "branch": current_branch,
                "commit": commit_hash,
                "type": "auto",
                "message": message,
                "files_changed": files_changed,
                "mode": "personal"
            }

            metadata["checkpoints"].append(new_checkpoint)

            # Keep only recent checkpoints (configurable limit)
            config = self.load_config()
            max_checkpoints = config.get("git_strategy", {}).get("personal", {}).get("max_checkpoints", 50)
            metadata["checkpoints"] = metadata["checkpoints"][-max_checkpoints:]

            # Save metadata
            with open(self.metadata_file, 'w', encoding='utf-8') as f:
                json.dump(metadata, f, indent=2, ensure_ascii=False)

        except Exception as e:
            print(f"⚠️ Failed to save checkpoint metadata: {e}")

    def cleanup_old_checkpoints(self):
        """Clean up old checkpoints based on configuration."""
        try:
            config = self.load_config()
            cleanup_days = config.get("git_strategy", {}).get("personal", {}).get("cleanup_days", 7)
            cutoff_time = datetime.now().timestamp() - (cleanup_days * 24 * 3600)

            # Load metadata
            if not self.metadata_file.exists():
                return

            with open(self.metadata_file, 'r', encoding='utf-8') as f:
                metadata = json.load(f)

            # Find old checkpoints
            old_checkpoints = []
            for checkpoint in metadata["checkpoints"]:
                checkpoint_time = datetime.fromisoformat(checkpoint["timestamp"]).timestamp()
                if checkpoint_time < cutoff_time:
                    old_checkpoints.append(checkpoint)

            # Delete old checkpoint branches
            for checkpoint in old_checkpoints:
                try:
                    subprocess.run(
                        ["git", "branch", "-D", checkpoint["id"]],
                        cwd=self.project_root,
                        capture_output=True
                    )
                    print(f"🗑️ Cleaned up old checkpoint: {checkpoint['id']}")
                except Exception:
                    pass

            # Update metadata
            metadata["checkpoints"] = [
                cp for cp in metadata["checkpoints"]
                if cp not in old_checkpoints
            ]

            with open(self.metadata_file, 'w', encoding='utf-8') as f:
                json.dump(metadata, f, indent=2, ensure_ascii=False)

            if old_checkpoints:
                print(f"✅ Cleaned up {len(old_checkpoints)} old checkpoints")

        except Exception as e:
            print(f"⚠️ Checkpoint cleanup failed: {e}")

    def run_once(self) -> bool:
        """Run checkpoint check once."""
        if self.should_create_checkpoint():
            return self.create_checkpoint("Auto-generated checkpoint")
        return False

    def run_daemon(self, interval: int = 60):
        """Run as a daemon process checking every interval seconds."""
        print(f"🔄 Auto-checkpoint daemon started (checking every {interval}s)")

        try:
            while True:
                if self.should_create_checkpoint():
                    self.create_checkpoint("Auto-generated checkpoint")

                # Cleanup old checkpoints occasionally
                if int(time.time()) % 3600 == 0:  # Every hour
                    self.cleanup_old_checkpoints()

                time.sleep(interval)

        except KeyboardInterrupt:
            print("\n⏹️ Auto-checkpoint daemon stopped")
        except Exception as e:
            print(f"❌ Daemon error: {e}")


def main():
    """Main entry point for auto checkpoint system."""
    if len(sys.argv) < 2:
        print("Usage: auto_checkpoint.py <project_root> [--daemon] [--once]")
        sys.exit(1)

    project_root = Path(sys.argv[1]).resolve()
    if not project_root.exists():
        print(f"❌ Project root does not exist: {project_root}")
        sys.exit(1)

    manager = AutoCheckpointManager(project_root)

    # Check if personal mode
    if not manager.is_personal_mode():
        print("ℹ️ Auto-checkpoint only works in personal mode")
        sys.exit(0)

    if "--daemon" in sys.argv:
        # Run as daemon
        interval = 60  # Check every minute
        manager.run_daemon(interval)
    elif "--once" in sys.argv:
        # Run once
        if manager.run_once():
            print("✅ Checkpoint created")
        else:
            print("ℹ️ No checkpoint needed")
    else:
        # Default: run once
        if manager.run_once():
            print("✅ Checkpoint created")


if __name__ == "__main__":
    main()
