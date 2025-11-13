# type: ignore
"""
Git information collector for statusline

"""

import logging
import re
import subprocess
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import Optional

logger = logging.getLogger(__name__)


@dataclass
class GitInfo:
    """Git repository information"""

    branch: str
    staged: int
    modified: int
    untracked: int


class GitCollector:
    """Collects git information from repository with 5-second caching"""

    # Configuration constants
    _CACHE_TTL_SECONDS = 5
    _GIT_COMMAND_TIMEOUT = 2

    # File status prefixes from git status --porcelain
    _STATUS_ADDED = "A"
    _STATUS_MODIFIED = "M"
    _STATUS_UNTRACKED = "??"

    def __init__(self):
        """Initialize git collector with cache"""
        self._cache: Optional[GitInfo] = None
        self._cache_time: Optional[datetime] = None
        self._cache_ttl = timedelta(seconds=self._CACHE_TTL_SECONDS)

    def collect_git_info(self) -> GitInfo:
        """
        Collect git information from the repository

        Returns:
            GitInfo containing branch name and change counts
        """
        # Check cache validity first
        if self._is_cache_valid():
            return self._cache

        # Run git command and parse output
        git_info = self._fetch_git_info()
        self._update_cache(git_info)
        return git_info

    def _fetch_git_info(self) -> GitInfo:
        """
        Fetch git information from command

        Returns:
            GitInfo from command or error defaults
        """
        try:
            result = subprocess.run(
                ["git", "status", "-b", "--porcelain"],
                capture_output=True,
                text=True,
                timeout=self._GIT_COMMAND_TIMEOUT,
            )

            if result.returncode != 0:
                logger.debug(f"Git command failed: {result.stderr}")
                return self._create_error_info()

            return self._parse_git_output(result.stdout)

        except subprocess.TimeoutExpired:
            logger.warning("Git command timed out")
            return self._create_error_info()
        except Exception as e:
            logger.debug(f"Error collecting git info: {e}")
            return self._create_error_info()

    def _is_cache_valid(self) -> bool:
        """Check if cache is still valid"""
        if self._cache is None or self._cache_time is None:
            return False
        return datetime.now() - self._cache_time < self._cache_ttl

    def _update_cache(self, git_info: GitInfo) -> None:
        """Update cache with new git info"""
        self._cache = git_info
        self._cache_time = datetime.now()

    def _parse_git_output(self, output: str) -> GitInfo:
        """
        Parse git status output into GitInfo

        Args:
            output: Output from 'git status -b --porcelain'

        Returns:
            GitInfo with parsed data
        """
        lines = output.strip().split("\n")

        # Extract branch name from first line (## branch_name...)
        branch = self._extract_branch(lines[0] if lines else "")

        # Count file changes from remaining lines
        staged, modified, untracked = self._count_changes(lines[1:])

        return GitInfo(
            branch=branch,
            staged=staged,
            modified=modified,
            untracked=untracked,
        )

    def _count_changes(self, lines: list) -> tuple[int, int, int]:
        """
        Count staged, modified, and untracked files

        Args:
            lines: Lines from git status output (excluding header)

        Returns:
            Tuple of (staged_count, modified_count, untracked_count)
        """
        staged = 0
        modified = 0
        untracked = 0

        for line in lines:
            if not line or len(line) < 2:
                continue

            status = line[:2]

            # Check staged changes (first character)
            if status[0] == self._STATUS_ADDED:
                staged += 1
            elif status[0] == self._STATUS_MODIFIED:
                staged += 1

            # Check unstaged/working directory changes (second character)
            # Detects: M(modified), D(deleted), A(added), R(renamed), C(copied), T(type changed)
            if len(status) > 1 and status[1] not in (" ", "."):
                modified += 1

            # Check untracked files
            if status == self._STATUS_UNTRACKED:
                untracked += 1

        return staged, modified, untracked

    @staticmethod
    def _extract_branch(branch_line: str) -> str:
        """
        Extract branch name from git status -b output

        Format: ## branch_name...origin/branch_name

        Args:
            branch_line: First line from git status -b

        Returns:
            Branch name or "unknown" if parsing fails
        """
        if not branch_line.startswith("##"):
            return "unknown"

        # Use regex to extract branch name before first dot
        match = re.match(r"##\s+([^\s\.]+)", branch_line)
        return match.group(1) if match else "unknown"

    @staticmethod
    def _create_error_info() -> GitInfo:
        """Create error info with default values for graceful degradation"""
        return GitInfo(
            branch="unknown",
            staged=0,
            modified=0,
            untracked=0,
        )
