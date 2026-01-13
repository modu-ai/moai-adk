# type: ignore
"""
Alfred task detector for statusline

"""

import json
import logging
from dataclasses import dataclass
from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional

logger = logging.getLogger(__name__)


@dataclass
class AlfredTask:
    """Alfred task information"""

    command: Optional[str]
    spec_id: Optional[str]
    stage: Optional[str]


class AlfredDetector:
    """Detects active Alfred tasks with 1-second caching"""

    # Configuration
    _CACHE_TTL_SECONDS = 1

    def __init__(self):
        """Initialize Alfred detector"""
        self._cache: Optional[AlfredTask] = None
        self._cache_time: Optional[datetime] = None
        self._cache_ttl = timedelta(seconds=self._CACHE_TTL_SECONDS)
        self._session_state_path = Path.home() / ".moai" / "memory" / "last-session-state.json"

    def detect_active_task(self) -> AlfredTask:
        """
        Detect currently active Alfred task

        Returns:
            AlfredTask with command and spec_id
        """
        # Check cache
        if self._is_cache_valid():
            return self._cache

        # Read and parse session state
        task = self._read_session_state()
        self._update_cache(task)
        return task

    def _read_session_state(self) -> AlfredTask:
        """
        Read Alfred task from session state file

        Returns:
            AlfredTask from file or defaults
        """
        try:
            if not self._session_state_path.exists():
                return self._create_default_task()

            with open(self._session_state_path, "r") as f:
                data = json.load(f)

            active_task = data.get("active_task")
            if active_task is None:
                return self._create_default_task()

            # Parse command and spec_id
            command = active_task.get("command")
            spec_id = active_task.get("spec_id")

            return AlfredTask(
                command=command,
                spec_id=spec_id,
                stage=active_task.get("stage"),
            )

        except Exception as e:
            logger.debug(f"Error reading Alfred task: {e}")
            return self._create_default_task()

    def _is_cache_valid(self) -> bool:
        """Check if task cache is still valid"""
        if self._cache is None or self._cache_time is None:
            return False
        return datetime.now() - self._cache_time < self._cache_ttl

    def _update_cache(self, task: AlfredTask) -> None:
        """Update task cache"""
        self._cache = task
        self._cache_time = datetime.now()

    @staticmethod
    def _create_default_task() -> AlfredTask:
        """Create default task (no active task)"""
        return AlfredTask(
            command=None,
            spec_id=None,
            stage=None,
        )
