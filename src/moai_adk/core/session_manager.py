"""
Session Manager for MoAI-ADK Agent Orchestration

Manages sub-agent session IDs and resume logic based on official Claude Code documentation.
Provides session tracking, result storage, and resume decision making.

Official Documentation Reference:
https://code.claude.com/docs/en/sub-agents

Key Principles:
- Sub-agents operate in isolated context windows
- No direct agent-to-agent communication
- Results flow through main conversation thread (Alfred)
- Resume preserves full conversation history
- Each execution gets unique agentId
"""

import json
import logging
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)


class SessionManager:
    """
    Manages sub-agent session IDs and resume logic.

    Based on official Claude Code sub-agent pattern:
    - Each agent execution gets unique agentId
    - Resume parameter inherits full conversation history
    - Session transcripts stored in agent-{agentId}.jsonl

    Attributes:
        _sessions: Mapping of agent_name to current agentId
        _results: Storage of agent execution results (agentId → result data)
        _chains: Workflow chains tracking (chain_name → [agentIds])
        _session_file: Persistent storage location
        _transcript_dir: Directory for conversation transcripts
    """

    def __init__(
        self,
        session_file: Optional[Path] = None,
        transcript_dir: Optional[Path] = None,
    ):
        """
        Initialize SessionManager.

        Args:
            session_file: Path to session storage JSON file
                         (default: .moai/memory/agent-sessions.json)
            transcript_dir: Directory for agent transcripts
                           (default: .moai/logs/agent-transcripts/)
        """
        # Default paths
        project_root = Path.cwd()
        self._session_file = session_file or project_root / ".moai" / "memory" / "agent-sessions.json"
        self._transcript_dir = transcript_dir or project_root / ".moai" / "logs" / "agent-transcripts"

        # Ensure directories exist
        self._session_file.parent.mkdir(parents=True, exist_ok=True)
        self._transcript_dir.mkdir(parents=True, exist_ok=True)

        # In-memory storage
        self._sessions: Dict[str, str] = {}  # agent_name → current agentId
        self._results: Dict[str, Any] = {}  # agentId → result data
        self._chains: Dict[str, List[str]] = {}  # chain_name → [agentIds]
        self._metadata: Dict[str, Dict[str, Any]] = {}  # agentId → metadata

        # Load existing sessions
        self._load_sessions()

    def _load_sessions(self) -> None:
        """Load session data from persistent storage."""
        if self._session_file.exists():
            try:
                with open(self._session_file, "r", encoding="utf-8") as f:
                    data = json.load(f)
                    self._sessions = data.get("sessions", {})
                    self._chains = data.get("chains", {})
                    self._metadata = data.get("metadata", {})
                logger.info(f"Loaded {len(self._sessions)} sessions from {self._session_file}")
            except json.JSONDecodeError as e:
                logger.warning(f"Failed to load sessions: {e}")
                self._sessions = {}
                self._chains = {}
                self._metadata = {}

    def _save_sessions(self) -> None:
        """Save session data to persistent storage."""
        data = {
            "sessions": self._sessions,
            "chains": self._chains,
            "metadata": self._metadata,
            "last_updated": datetime.now().isoformat(),
        }

        try:
            with open(self._session_file, "w", encoding="utf-8") as f:
                json.dump(data, f, indent=2, ensure_ascii=False)
            logger.debug(f"Saved sessions to {self._session_file}")
        except IOError as e:
            logger.error(f"Failed to save sessions: {e}")

    def register_agent_result(
        self,
        agent_name: str,
        agent_id: str,
        result: Any,
        chain_id: Optional[str] = None,
    ) -> None:
        """
        Register agent execution result in main context.

        This method implements the official pattern:
        "Results flow through main conversation thread"

        Args:
            agent_name: Name of the agent (e.g., "tdd-implementer")
            agent_id: Unique agentId returned from Task() execution
            result: Result data from agent execution
            chain_id: Optional workflow chain identifier (e.g., "SPEC-AUTH-001-implementation")
        """
        # Store agent ID mapping
        self._sessions[agent_name] = agent_id

        # Store result data
        self._results[agent_id] = {
            "agent_name": agent_name,
            "result": result,
            "timestamp": datetime.now().isoformat(),
            "chain_id": chain_id,
        }

        # Track in workflow chain
        if chain_id:
            if chain_id not in self._chains:
                self._chains[chain_id] = []
            self._chains[chain_id].append(agent_id)

        # Store metadata
        self._metadata[agent_id] = {
            "agent_name": agent_name,
            "created_at": datetime.now().isoformat(),
            "chain_id": chain_id,
            "resume_count": 0,
        }

        # Persist to disk
        self._save_sessions()

        logger.info(f"Registered agent result: {agent_name} (agentId: {agent_id[:8]}..., chain: {chain_id})")

    def get_resume_id(
        self,
        agent_name: str,
        chain_id: Optional[str] = None,
    ) -> Optional[str]:
        """
        Get agentId to resume if continuing same work.

        Official pattern:
        - resume parameter preserves full conversation history
        - Same agent can continue work with context

        Args:
            agent_name: Name of the agent to resume
            chain_id: Optional workflow chain to resume

        Returns:
            agentId to resume, or None if should start new session
        """
        # Check if agent has previous session
        if agent_name not in self._sessions:
            logger.debug(f"No previous session for {agent_name}")
            return None

        agent_id = self._sessions[agent_name]

        # Validate chain_id if provided
        if chain_id:
            metadata = self._metadata.get(agent_id, {})
            if metadata.get("chain_id") != chain_id:
                logger.debug(f"Chain mismatch: {agent_name} was in {metadata.get('chain_id')}, requested {chain_id}")
                return None

        logger.info(f"Resume ID for {agent_name}: {agent_id[:8]}...")
        return agent_id

    def should_resume(
        self,
        agent_name: str,
        current_task: str,
        previous_task: Optional[str] = None,
    ) -> bool:
        """
        Determine if resume or new invocation is appropriate.

        Decision logic based on official best practices:
        - Resume: Same agent, continuing previous task, context continuity needed
        - New: Different agent, independent task, context switch

        Args:
            agent_name: Name of the agent
            current_task: Description of current task
            previous_task: Description of previous task (if any)

        Returns:
            True if should resume, False if should start new session
        """
        # No previous session → new
        if agent_name not in self._sessions:
            return False

        # No previous task information → new
        if not previous_task:
            return False

        # Check resume count (prevent infinite loops)
        agent_id = self._sessions[agent_name]
        metadata = self._metadata.get(agent_id, {})
        resume_count = metadata.get("resume_count", 0)

        if resume_count >= 5:  # Max resume depth from config
            logger.warning(f"{agent_name} has been resumed {resume_count} times, starting new session")
            return False

        # Heuristic: Check if tasks are related
        # (This can be enhanced with semantic similarity)
        task_keywords_match = any(
            keyword in current_task.lower() for keyword in previous_task.lower().split() if len(keyword) > 4
        )

        if task_keywords_match:
            logger.info(f"Tasks appear related, resuming {agent_name}")
            return True

        logger.info(f"Tasks appear independent, starting new session for {agent_name}")
        return False

    def increment_resume_count(self, agent_id: str) -> None:
        """
        Increment resume count for an agent session.

        Args:
            agent_id: Agent session ID
        """
        if agent_id in self._metadata:
            self._metadata[agent_id]["resume_count"] += 1
            self._metadata[agent_id]["last_resumed_at"] = datetime.now().isoformat()
            self._save_sessions()

    def get_agent_result(self, agent_id: str) -> Optional[Any]:
        """
        Retrieve stored result for an agent execution.

        Args:
            agent_id: Agent session ID

        Returns:
            Stored result data, or None if not found
        """
        result_data = self._results.get(agent_id)
        if result_data:
            return result_data["result"]
        return None

    def get_chain_results(self, chain_id: str) -> List[Dict[str, Any]]:
        """
        Get all agent results in a workflow chain.

        Args:
            chain_id: Workflow chain identifier

        Returns:
            List of result dictionaries in execution order
        """
        if chain_id not in self._chains:
            return []

        agent_ids = self._chains[chain_id]
        results = []

        for agent_id in agent_ids:
            if agent_id in self._results:
                results.append(self._results[agent_id])

        return results

    def get_chain_summary(self, chain_id: str) -> Dict[str, Any]:
        """
        Get summary of a workflow chain.

        Args:
            chain_id: Workflow chain identifier

        Returns:
            Summary dictionary with agent names, timestamps, etc.
        """
        results = self.get_chain_results(chain_id)

        if not results:
            return {"chain_id": chain_id, "status": "not_found"}

        return {
            "chain_id": chain_id,
            "agent_count": len(results),
            "agents": [r["agent_name"] for r in results],
            "started_at": results[0]["timestamp"] if results else None,
            "completed_at": results[-1]["timestamp"] if results else None,
            "status": "completed",
        }

    def clear_agent_session(self, agent_name: str) -> None:
        """
        Clear session data for a specific agent.

        Use when you want to force a new session for an agent.

        Args:
            agent_name: Name of the agent
        """
        if agent_name in self._sessions:
            agent_id = self._sessions[agent_name]
            del self._sessions[agent_name]

            if agent_id in self._results:
                del self._results[agent_id]

            if agent_id in self._metadata:
                del self._metadata[agent_id]

            self._save_sessions()
            logger.info(f"Cleared session for {agent_name}")

    def clear_chain(self, chain_id: str) -> None:
        """
        Clear all sessions in a workflow chain.

        Args:
            chain_id: Workflow chain identifier
        """
        if chain_id in self._chains:
            agent_ids = self._chains[chain_id]

            for agent_id in agent_ids:
                if agent_id in self._results:
                    del self._results[agent_id]
                if agent_id in self._metadata:
                    del self._metadata[agent_id]

            del self._chains[chain_id]
            self._save_sessions()
            logger.info(f"Cleared chain: {chain_id}")

    def get_all_sessions(self) -> Dict[str, Any]:
        """
        Get all active sessions.

        Returns:
            Dictionary with all session data
        """
        return {
            "sessions": self._sessions,
            "chains": list(self._chains.keys()),
            "total_results": len(self._results),
        }

    def export_transcript(self, agent_id: str) -> Optional[Path]:
        """
        Get path to agent conversation transcript.

        Official pattern:
        - Transcripts stored in agent-{agentId}.jsonl
        - Contains full conversation history

        Args:
            agent_id: Agent session ID

        Returns:
            Path to transcript file, or None if not found
        """
        transcript_file = self._transcript_dir / f"agent-{agent_id}.jsonl"

        if transcript_file.exists():
            return transcript_file

        logger.warning(f"Transcript not found for agentId: {agent_id}")
        return None

    def create_chain(
        self,
        chain_id: str,
        agent_sequence: List[str],
        metadata: Optional[Dict[str, Any]] = None,
    ) -> None:
        """
        Create a new workflow chain.

        Args:
            chain_id: Unique chain identifier (e.g., "SPEC-AUTH-001-implementation")
            agent_sequence: Expected agent execution order
            metadata: Optional metadata for the chain
        """
        self._chains[chain_id] = []

        chain_metadata = {
            "created_at": datetime.now().isoformat(),
            "expected_sequence": agent_sequence,
            "metadata": metadata or {},
        }

        # Store in a separate chains metadata file
        chains_file = self._session_file.parent / "workflow-chains.json"

        if chains_file.exists():
            with open(chains_file, "r", encoding="utf-8") as f:
                chains_data = json.load(f)
        else:
            chains_data = {}

        chains_data[chain_id] = chain_metadata

        with open(chains_file, "w", encoding="utf-8") as f:
            json.dump(chains_data, f, indent=2, ensure_ascii=False)

        logger.info(f"Created workflow chain: {chain_id} with {len(agent_sequence)} agents")


# Global instance (singleton pattern)
_session_manager_instance: Optional[SessionManager] = None


def get_session_manager() -> SessionManager:
    """
    Get global SessionManager instance (singleton).

    Returns:
        SessionManager instance
    """
    global _session_manager_instance

    if _session_manager_instance is None:
        _session_manager_instance = SessionManager()

    return _session_manager_instance


# Convenience functions for direct use


def register_agent(
    agent_name: str,
    agent_id: str,
    result: Any,
    chain_id: Optional[str] = None,
) -> None:
    """
    Convenience function to register agent result.

    Args:
        agent_name: Name of the agent
        agent_id: Unique agentId from Task() execution
        result: Result data
        chain_id: Optional workflow chain identifier
    """
    manager = get_session_manager()
    manager.register_agent_result(agent_name, agent_id, result, chain_id)


def get_resume_id(agent_name: str, chain_id: Optional[str] = None) -> Optional[str]:
    """
    Convenience function to get resume ID.

    Args:
        agent_name: Name of the agent
        chain_id: Optional workflow chain

    Returns:
        agentId to resume, or None
    """
    manager = get_session_manager()
    return manager.get_resume_id(agent_name, chain_id)


def should_resume(
    agent_name: str,
    current_task: str,
    previous_task: Optional[str] = None,
) -> bool:
    """
    Convenience function to check if should resume.

    Args:
        agent_name: Name of the agent
        current_task: Description of current task
        previous_task: Description of previous task

    Returns:
        True if should resume
    """
    manager = get_session_manager()
    return manager.should_resume(agent_name, current_task, previous_task)


# Example usage for documentation
if __name__ == "__main__":
    """
    Example usage of SessionManager.

    This demonstrates the official Claude Code sub-agent patterns.
    """

    # Initialize manager
    manager = SessionManager()

    # Example 1: Linear Chain (spec-builder → implementation-planner)
    print("=== Example 1: Linear Chain ===")

    # Create workflow chain
    manager.create_chain(
        chain_id="SPEC-AUTH-001-planning",
        agent_sequence=["spec-builder", "implementation-planner"],
        metadata={"spec_id": "SPEC-AUTH-001", "feature": "User Authentication"},
    )

    # Simulate spec-builder execution
    spec_result = {
        "spec_id": "SPEC-AUTH-001",
        "files_created": [".moai/specs/SPEC-AUTH-001/spec.md"],
        "status": "success",
    }

    manager.register_agent_result(
        agent_name="spec-builder",
        agent_id="spec-abc123",
        result=spec_result,
        chain_id="SPEC-AUTH-001-planning",
    )

    # Simulate implementation-planner execution
    plan_result = {
        "dependencies": {"fastapi": ">=0.118.3"},
        "status": "success",
    }

    manager.register_agent_result(
        agent_name="implementation-planner",
        agent_id="plan-def456",
        result=plan_result,
        chain_id="SPEC-AUTH-001-planning",
    )

    # Get chain summary
    summary = manager.get_chain_summary("SPEC-AUTH-001-planning")
    print(f"Chain summary: {json.dumps(summary, indent=2)}")

    # Example 2: Resume Pattern (tdd-implementer continues work)
    print("\n=== Example 2: Resume Pattern ===")

    manager.create_chain(
        chain_id="SPEC-AUTH-001-implementation",
        agent_sequence=["tdd-implementer"],
    )

    # First execution: Implementation phase 1
    implementation_001_result = {
        "phase": "phase_1",
        "tests_created": ["tests/test_registration.py"],
        "code_created": ["src/auth/registration.py"],
        "status": "success",
    }

    manager.register_agent_result(
        agent_name="tdd-implementer",
        agent_id="tdd-ghi789",
        result=implementation_001_result,
        chain_id="SPEC-AUTH-001-implementation",
    )

    # Get resume ID for continuing work
    resume_id = manager.get_resume_id(
        agent_name="tdd-implementer",
        chain_id="SPEC-AUTH-001-implementation",
    )

    print(f"Resume ID for tdd-implementer: {resume_id}")

    # Should resume? (continuing user auth flow)
    should_resume_decision = manager.should_resume(
        agent_name="tdd-implementer",
        current_task="Implement user login endpoint",
        previous_task="Implement user registration endpoint",
    )

    print(f"Should resume? {should_resume_decision}")

    if should_resume_decision and resume_id:
        print(f"✅ Resume with agentId: {resume_id}")
        manager.increment_resume_count(resume_id)
    else:
        print("❌ Start new session")

    # Example 3: Parallel Analysis
    print("\n=== Example 3: Parallel Analysis ===")

    manager.create_chain(
        chain_id="SPEC-AUTH-001-review",
        agent_sequence=["backend-expert", "security-expert", "frontend-expert"],
        metadata={"review_type": "expert_consultation"},
    )

    # All experts run independently (no resume)
    experts_results = {
        "backend-expert": {
            "recommendations": ["Use JWT for auth"],
            "agent_id": "backend-jkl012",
        },
        "security-expert": {
            "vulnerabilities": ["Rate limiting needed"],
            "agent_id": "security-mno345",
        },
        "frontend-expert": {
            "ui_concerns": ["Token refresh flow"],
            "agent_id": "frontend-pqr678",
        },
    }

    for expert_name, data in experts_results.items():
        agent_id_value: str = data["agent_id"]  # type: ignore[assignment]
        manager.register_agent_result(
            agent_name=expert_name,
            agent_id=agent_id_value,
            result={k: v for k, v in data.items() if k != "agent_id"},
            chain_id="SPEC-AUTH-001-review",
        )

    # Get all review results
    review_results = manager.get_chain_results("SPEC-AUTH-001-review")
    print(f"Expert reviews: {len(review_results)} experts")

    for result in review_results:
        print(f"  - {result['agent_name']}: {list(result['result'].keys())}")

    # Get all sessions
    print("\n=== All Sessions ===")
    all_sessions = manager.get_all_sessions()
    print(json.dumps(all_sessions, indent=2))
