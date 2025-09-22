#!/usr/bin/env python3
"""
File Watcher for MoAI-ADK Auto-Checkpoint System v0.1.0

Monitors file changes and triggers checkpoint creation when needed.
Uses efficient file system events to minimize resource usage.

@REQ:FILE-WATCHER-001
@FEATURE:FILE-MONITORING-001
@API:FILESYSTEM-EVENTS-001
@DESIGN:WATCHDOG-INTEGRATION-001
@TECH:EVENT-DRIVEN-CHECKPOINT-001
"""

import os
import sys
import time
import threading
from pathlib import Path
from typing import Set, Optional

try:
    from watchdog.observers import Observer
    from watchdog.events import FileSystemEventHandler
    WATCHDOG_AVAILABLE = True
except ImportError:
    WATCHDOG_AVAILABLE = False


class MoAIFileWatcher(FileSystemEventHandler):
    """File system event handler for MoAI-ADK projects.

    @FEATURE:FILE-MONITORING-001
    @API:FS-EVENT-HANDLER-001
    """

    def __init__(self, checkpoint_manager):
        super().__init__()
        self.checkpoint_manager = checkpoint_manager
        self.changed_files: Set[str] = set()
        self.last_event_time = 0
        self.debounce_delay = 2  # seconds
        self.checkpoint_delay = 5  # seconds after last change
        self.timer: Optional[threading.Timer] = None

        # File patterns to watch
        self.watch_patterns = {
            '.py', '.js', '.ts', '.java', '.cpp', '.c', '.h',
            '.md', '.txt', '.json', '.yml', '.yaml',
            '.html', '.css', '.scss', '.vue', '.jsx', '.tsx'
        }

        # Patterns to ignore
        self.ignore_patterns = {
            '.git', '__pycache__', '.pytest_cache', 'node_modules',
            '.vscode', '.idea', '.DS_Store', '.coverage',
            '*.pyc', '*.pyo', '*.log', '*.tmp'
        }

    def should_watch_file(self, file_path: str) -> bool:
        """Determine if a file should be watched for changes."""
        path = Path(file_path)

        # Check if file is in ignored directories
        for part in path.parts:
            if any(pattern in part for pattern in self.ignore_patterns):
                return False

        # Check file extension
        if path.suffix.lower() in self.watch_patterns:
            return True

        # Check if it's a SPEC file
        if 'spec' in path.name.lower() or 'SPEC' in path.name:
            return True

        return False

    def on_modified(self, event):
        """Handle file modification events."""
        if event.is_directory:
            return

        if not self.should_watch_file(event.src_path):
            return

        self.changed_files.add(event.src_path)
        self.last_event_time = time.time()

        # Cancel previous timer
        if self.timer:
            self.timer.cancel()

        # Set new timer
        self.timer = threading.Timer(self.checkpoint_delay, self.trigger_checkpoint)
        self.timer.start()

    def on_created(self, event):
        """Handle file creation events."""
        if not event.is_directory and self.should_watch_file(event.src_path):
            self.on_modified(event)

    def trigger_checkpoint(self):
        """Trigger checkpoint creation after file changes."""
        if not self.changed_files:
            return

        try:
            # Create descriptive message
            file_count = len(self.changed_files)
            if file_count == 1:
                file_name = Path(list(self.changed_files)[0]).name
                message = f"Modified: {file_name}"
            else:
                message = f"Modified {file_count} files"

            # Create checkpoint
            success = self.checkpoint_manager.create_checkpoint(message)

            if success:
                print(f"📝 Checkpoint triggered by file changes: {message}")
            else:
                print("⚠️ Checkpoint creation failed")

        except Exception as e:
            print(f"❌ Error triggering checkpoint: {e}")
        finally:
            # Clear changed files
            self.changed_files.clear()


class AutoCheckpointWatcher:
    """Main file watcher for auto-checkpoint system."""

    def __init__(self, project_root: Path):
        self.project_root = project_root

        # Import checkpoint manager from this hooks module directory
        hooks_dir = Path(__file__).resolve().parent
        sys.path.insert(0, str(hooks_dir))
        from auto_checkpoint import AutoCheckpointManager

        self.checkpoint_manager = AutoCheckpointManager(project_root)
        self.observer = None
        self.handler = None

    def start(self):
        """Start file watching."""
        if not WATCHDOG_AVAILABLE:
            print("❌ watchdog library not available. Install with: pip install watchdog")
            return False

        if not self.checkpoint_manager.is_personal_mode():
            print("ℹ️ File watcher only works in personal mode")
            return False

        if not self.checkpoint_manager.is_auto_checkpoint_enabled():
            print("ℹ️ Auto-checkpoint is disabled")
            return False

        try:
            self.handler = MoAIFileWatcher(self.checkpoint_manager)
            self.observer = Observer()

            # Watch project root recursively
            self.observer.schedule(
                self.handler,
                str(self.project_root),
                recursive=True
            )

            self.observer.start()
            print(f"👁️ File watcher started for: {self.project_root}")
            print("   Monitoring code files for automatic checkpoints...")
            return True

        except Exception as e:
            print(f"❌ Failed to start file watcher: {e}")
            return False

    def stop(self):
        """Stop file watching."""
        if self.observer:
            self.observer.stop()
            self.observer.join()
            print("⏹️ File watcher stopped")

    def run(self):
        """Run file watcher until interrupted."""
        if not self.start():
            return

        try:
            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            print("\n🛑 Stopping file watcher...")
        finally:
            self.stop()


class FileWatcherDaemon:
    """Daemon process for file watcher."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.pid_file = project_root / ".moai" / "file_watcher.pid"
        self.log_file = project_root / ".moai" / "file_watcher.log"

    def start_daemon(self):
        """Start the file watcher as a daemon process."""
        if self.is_running():
            print("⚠️ File watcher daemon is already running")
            return False

        try:
            import subprocess

            # Start daemon process
            process = subprocess.Popen([
                sys.executable, __file__, str(self.project_root), "--daemon"
            ], stdout=subprocess.PIPE, stderr=subprocess.PIPE)

            # Save PID
            with open(self.pid_file, 'w') as f:
                f.write(str(process.pid))

            print(f"🚀 File watcher daemon started (PID: {process.pid})")
            return True

        except Exception as e:
            print(f"❌ Failed to start daemon: {e}")
            return False

    def stop_daemon(self):
        """Stop the file watcher daemon."""
        if not self.is_running():
            print("ℹ️ File watcher daemon is not running")
            return False

        try:
            import signal

            with open(self.pid_file, 'r') as f:
                pid = int(f.read().strip())

            os.kill(pid, signal.SIGTERM)
            self.pid_file.unlink(missing_ok=True)

            print(f"⏹️ File watcher daemon stopped (PID: {pid})")
            return True

        except (FileNotFoundError, ProcessLookupError):
            # Process already dead, clean up
            self.pid_file.unlink(missing_ok=True)
            print("✅ Cleaned up stale daemon process")
            return True
        except Exception as e:
            print(f"❌ Failed to stop daemon: {e}")
            return False

    def is_running(self) -> bool:
        """Check if daemon is running."""
        if not self.pid_file.exists():
            return False

        try:
            with open(self.pid_file, 'r') as f:
                pid = int(f.read().strip())

            # Check if process exists
            os.kill(pid, 0)
            return True

        except (FileNotFoundError, ProcessLookupError, ValueError):
            # Process doesn't exist, clean up
            self.pid_file.unlink(missing_ok=True)
            return False

    def get_status(self):
        """Get daemon status."""
        if self.is_running():
            with open(self.pid_file, 'r') as f:
                pid = int(f.read().strip())
            print(f"✅ File watcher daemon is running (PID: {pid})")
        else:
            print("❌ File watcher daemon is not running")


def main():
    """Main entry point for file watcher."""
    if len(sys.argv) < 2:
        print("Usage: file_watcher.py <project_root> [--daemon|--start|--stop|--status]")
        sys.exit(1)

    project_root = Path(sys.argv[1]).resolve()
    if not project_root.exists():
        print(f"❌ Project root does not exist: {project_root}")
        sys.exit(1)

    daemon = FileWatcherDaemon(project_root)

    if "--daemon" in sys.argv:
        # Run as daemon (called by start_daemon)
        watcher = AutoCheckpointWatcher(project_root)
        watcher.run()
    elif "--start" in sys.argv:
        # Start daemon
        daemon.start_daemon()
    elif "--stop" in sys.argv:
        # Stop daemon
        daemon.stop_daemon()
    elif "--status" in sys.argv:
        # Show status
        daemon.get_status()
    else:
        # Run in foreground
        watcher = AutoCheckpointWatcher(project_root)
        watcher.run()


if __name__ == "__main__":
    main()
