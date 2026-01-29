# Context & Session Management Examples

Multishot prompting examples demonstrating practical context management patterns covering 200K token budget optimization, session persistence, multi-agent handoffs, and progressive disclosure.

---

## Example 1: Token Budget Monitoring and Management

**Scenario**: Implement real-time token budget monitoring with automatic actions.

**Input**:
```python
# Need: Token budget management that:
# - Tracks usage against 200K limit
# - Warns at critical thresholds
# - Triggers automatic compression
# - Prevents context overflow
```

**Output**:
```python
from dataclasses import dataclass
from typing import Optional, Callable
from datetime import datetime
import logging

logger = logging.getLogger(__name__)

@dataclass
class TokenBudgetState:
    total_budget: int = 200000
    used_tokens: int = 0
    system_reserved: int = 15000  # System prompt, instructions
    emergency_reserve: int = 55000  # Last 27.5% for recovery

    @property
    def available(self) -> int:
        return self.total_budget - self.used_tokens

    @property
    def usage_percent(self) -> float:
        return (self.used_tokens / self.total_budget) * 100

    @property
    def is_critical(self) -> bool:
        return self.usage_percent >= 85

    @property
    def is_warning(self) -> bool:
        return self.usage_percent >= 75


class TokenBudgetManager:
    """Manages 200K token budget with automatic optimization."""

    # Budget allocation (recommended breakdown)
    ALLOCATION = {
        "system": 15000,      # 7.5% - System prompt, instructions
        "conversation": 80000, # 40% - Active conversation
        "reference": 50000,   # 25% - Reference context
        "reserve": 55000      # 27.5% - Emergency reserve
    }

    THRESHOLDS = {
        "normal": 60,
        "warning": 75,
        "critical": 85,
        "emergency": 95
    }

    def __init__(self):
        self.state = TokenBudgetState()
        self.callbacks: dict[str, Callable] = {}
        self.checkpoints: list[dict] = []

    def register_callback(self, event: str, callback: Callable):
        """Register callback for budget events."""
        self.callbacks[event] = callback

    def update_usage(self, tokens_used: int, source: str = "unknown"):
        """Update token usage and trigger appropriate actions."""
        self.state.used_tokens = tokens_used

        logger.info(
            f"Token update: {tokens_used:,}/{self.state.total_budget:,} "
            f"({self.state.usage_percent:.1f}%) from {source}"
        )

        # Check thresholds and trigger actions
        if self.state.usage_percent >= self.THRESHOLDS["emergency"]:
            self._handle_emergency()
        elif self.state.usage_percent >= self.THRESHOLDS["critical"]:
            self._handle_critical()
        elif self.state.usage_percent >= self.THRESHOLDS["warning"]:
            self._handle_warning()

    def _handle_warning(self):
        """Handle warning threshold (75%)."""
        logger.warning(
            f"Token usage at {self.state.usage_percent:.1f}% - "
            "Starting context optimization"
        )

        # Defer non-critical context
        if "on_warning" in self.callbacks:
            self.callbacks["on_warning"](self.state)

    def _handle_critical(self):
        """Handle critical threshold (85%)."""
        logger.error(
            f"Token usage CRITICAL at {self.state.usage_percent:.1f}% - "
            "Triggering context compression"
        )

        # Create checkpoint before compression
        self.create_checkpoint("pre_compression")

        if "on_critical" in self.callbacks:
            self.callbacks["on_critical"](self.state)

    def _handle_emergency(self):
        """Handle emergency threshold (95%)."""
        logger.critical(
            f"Token usage EMERGENCY at {self.state.usage_percent:.1f}% - "
            "Forcing context clear"
        )

        # Force immediate action
        if "on_emergency" in self.callbacks:
            self.callbacks["on_emergency"](self.state)

    def create_checkpoint(self, name: str):
        """Create a checkpoint for potential recovery."""
        checkpoint = {
            "name": name,
            "timestamp": datetime.utcnow().isoformat(),
            "tokens_used": self.state.used_tokens,
            "usage_percent": self.state.usage_percent
        }
        self.checkpoints.append(checkpoint)
        logger.info(f"Checkpoint created: {name}")
        return checkpoint

    def get_optimization_suggestions(self) -> list[str]:
        """Get suggestions for reducing token usage."""
        suggestions = []

        if self.state.usage_percent > 50:
            suggestions.append("Consider using /clear to reset context")

        if self.state.usage_percent > 60:
            suggestions.append("Defer non-critical reference documentation")

        if self.state.usage_percent > 75:
            suggestions.append("Apply progressive summarization to conversation")

        if self.state.usage_percent > 85:
            suggestions.append("URGENT: Execute context compression immediately")

        return suggestions


# Usage example
manager = TokenBudgetManager()

def on_warning(state: TokenBudgetState):
    print(f"Warning: {state.available:,} tokens remaining")

def on_critical(state: TokenBudgetState):
    print("Critical: Compressing context...")
    # Execute compression logic

def on_emergency(state: TokenBudgetState):
    print("Emergency: Forcing /clear command")
    # Execute /clear

manager.register_callback("on_warning", on_warning)
manager.register_callback("on_critical", on_critical)
manager.register_callback("on_emergency", on_emergency)

# Simulate usage updates
manager.update_usage(120000, source="conversation")  # 60% - normal
manager.update_usage(155000, source="file_loading")  # 77.5% - warning triggered
manager.update_usage(175000, source="agent_output")  # 87.5% - critical triggered
```

**Explanation**: This pattern implements comprehensive token budget monitoring with threshold-based callbacks, checkpointing, and optimization suggestions.

---

## Example 2: Session State Persistence

**Scenario**: Implement session state persistence for resumable workflows.

**Input**:
```python
# Need: Session management that:
# - Persists state across interruptions
# - Supports workflow resumption
# - Tracks SPEC progress
# - Handles model switches
```

**Output**:
```python
import json
from pathlib import Path
from dataclasses import dataclass, asdict, field
from datetime import datetime
from typing import Optional, List, Dict, Any
from enum import Enum

class SessionPhase(Enum):
    INIT = "initialization"
    PLANNING = "planning"
    IMPLEMENTATION = "implementation"
    TESTING = "testing"
    DOCUMENTATION = "documentation"
    REVIEW = "review"
    COMPLETED = "completed"


@dataclass
class WorkState:
    current_spec: Optional[str] = None
    phase: SessionPhase = SessionPhase.INIT
    completed_steps: List[str] = field(default_factory=list)
    pending_steps: List[str] = field(default_factory=list)
    artifacts: Dict[str, str] = field(default_factory=dict)


@dataclass
class ContextWindow:
    total: int = 200000
    used: int = 0
    available: int = 200000
    position_percent: float = 0.0


@dataclass
class SessionState:
    session_id: str
    model: str
    created_at: str
    last_updated: str
    context_window: ContextWindow
    work_state: WorkState
    user_context: Dict[str, Any] = field(default_factory=dict)
    persistence: Dict[str, Any] = field(default_factory=dict)


class SessionManager:
    """Manages session state with persistence and recovery."""

    def __init__(self, storage_path: str = ".moai/sessions"):
        self.storage_path = Path(storage_path)
        self.storage_path.mkdir(parents=True, exist_ok=True)
        self.current_session: Optional[SessionState] = None

    def create_session(
        self,
        model: str = "claude-sonnet-4-5-20250929",
        user_context: Optional[Dict] = None
    ) -> SessionState:
        """Create a new session."""
        import uuid

        session_id = f"sess_{uuid.uuid4().hex[:12]}"
        now = datetime.utcnow().isoformat()

        session = SessionState(
            session_id=session_id,
            model=model,
            created_at=now,
            last_updated=now,
            context_window=ContextWindow(),
            work_state=WorkState(),
            user_context=user_context or {},
            persistence={
                "auto_save": True,
                "save_interval_seconds": 60,
                "context_preservation": "critical_only"
            }
        )

        self.current_session = session
        self._save_session(session)
        return session

    def load_session(self, session_id: str) -> Optional[SessionState]:
        """Load an existing session."""
        file_path = self.storage_path / f"{session_id}.json"

        if not file_path.exists():
            return None

        with open(file_path, 'r') as f:
            data = json.load(f)

        # Reconstruct dataclasses
        session = SessionState(
            session_id=data["session_id"],
            model=data["model"],
            created_at=data["created_at"],
            last_updated=data["last_updated"],
            context_window=ContextWindow(**data["context_window"]),
            work_state=WorkState(
                current_spec=data["work_state"]["current_spec"],
                phase=SessionPhase(data["work_state"]["phase"]),
                completed_steps=data["work_state"]["completed_steps"],
                pending_steps=data["work_state"]["pending_steps"],
                artifacts=data["work_state"]["artifacts"]
            ),
            user_context=data.get("user_context", {}),
            persistence=data.get("persistence", {})
        )

        self.current_session = session
        return session

    def update_work_state(
        self,
        spec_id: Optional[str] = None,
        phase: Optional[SessionPhase] = None,
        completed_step: Optional[str] = None,
        artifact: Optional[tuple[str, str]] = None
    ):
        """Update work state with automatic persistence."""
        if not self.current_session:
            raise ValueError("No active session")

        work = self.current_session.work_state

        if spec_id:
            work.current_spec = spec_id
        if phase:
            work.phase = phase
        if completed_step:
            work.completed_steps.append(completed_step)
            if completed_step in work.pending_steps:
                work.pending_steps.remove(completed_step)
        if artifact:
            work.artifacts[artifact[0]] = artifact[1]

        self.current_session.last_updated = datetime.utcnow().isoformat()
        self._save_session(self.current_session)

    def update_context_usage(self, tokens_used: int):
        """Update context window usage."""
        if not self.current_session:
            return

        ctx = self.current_session.context_window
        ctx.used = tokens_used
        ctx.available = ctx.total - tokens_used
        ctx.position_percent = (tokens_used / ctx.total) * 100

        self.current_session.last_updated = datetime.utcnow().isoformat()
        self._save_session(self.current_session)

    def get_resumption_context(self) -> Dict[str, Any]:
        """Get context for resuming interrupted work."""
        if not self.current_session:
            return {}

        return {
            "spec_id": self.current_session.work_state.current_spec,
            "phase": self.current_session.work_state.phase.value,
            "completed": self.current_session.work_state.completed_steps,
            "pending": self.current_session.work_state.pending_steps,
            "last_update": self.current_session.last_updated,
            "context_usage": self.current_session.context_window.position_percent
        }

    def prepare_for_clear(self) -> Dict[str, Any]:
        """Prepare essential context before /clear."""
        if not self.current_session:
            return {}

        # Save current state
        self._save_session(self.current_session)

        # Return minimal context to reload after clear
        return {
            "session_id": self.current_session.session_id,
            "spec_id": self.current_session.work_state.current_spec,
            "phase": self.current_session.work_state.phase.value,
            "reload_files": [
                "CLAUDE.md",
                f".moai/specs/{self.current_session.work_state.current_spec}.md"
            ]
        }

    def _save_session(self, session: SessionState):
        """Save session to storage."""
        file_path = self.storage_path / f"{session.session_id}.json"

        # Convert to dict with enum handling
        data = {
            "session_id": session.session_id,
            "model": session.model,
            "created_at": session.created_at,
            "last_updated": session.last_updated,
            "context_window": asdict(session.context_window),
            "work_state": {
                **asdict(session.work_state),
                "phase": session.work_state.phase.value
            },
            "user_context": session.user_context,
            "persistence": session.persistence
        }

        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2)


# Usage example
manager = SessionManager()

# Create new session
session = manager.create_session(
    model="claude-sonnet-4-5-20250929",
    user_context={"language": "ko", "user_name": "Developer"}
)
print(f"Created session: {session.session_id}")

# Update work progress
manager.update_work_state(
    spec_id="SPEC-001",
    phase=SessionPhase.IMPLEMENTATION
)
manager.update_work_state(completed_step="spec_complete")
manager.update_work_state(completed_step="architecture_defined")
manager.update_work_state(artifact=("api_schema", "schemas/user.json"))

# Before /clear
clear_context = manager.prepare_for_clear()
print(f"Context to reload after clear: {clear_context}")

# After interruption - resume
# resumed = manager.load_session(session.session_id)
# print(f"Resumed at phase: {resumed.work_state.phase.value}")
```

**Explanation**: This pattern provides comprehensive session state management with persistence, work progress tracking, and seamless resumption after interruptions or /clear commands.

---

## Example 3: Multi-Agent Handoff Protocol

**Scenario**: Implement efficient context handoff between agents.

**Input**:
```python
# Need: Multi-agent handoff that:
# - Minimizes token overhead
# - Validates handoff integrity
# - Supports context compression
# - Handles agent compatibility
```

**Output**:
```python
from dataclasses import dataclass, field
from typing import Dict, Any, List, Optional
from datetime import datetime
from enum import Enum
import hashlib

class AgentType(Enum):
    SPEC_BUILDER = "spec-builder"
    DDD_IMPLEMENTER = "ddd-implementer"
    BACKEND_EXPERT = "backend-expert"
    FRONTEND_EXPERT = "frontend-expert"
    DOCS_MANAGER = "docs-manager"
    QUALITY_GATE = "quality-gate"


# Agent compatibility matrix
AGENT_COMPATIBILITY = {
    AgentType.SPEC_BUILDER: [
        AgentType.DDD_IMPLEMENTER,
        AgentType.BACKEND_EXPERT,
        AgentType.FRONTEND_EXPERT
    ],
    AgentType.DDD_IMPLEMENTER: [
        AgentType.QUALITY_GATE,
        AgentType.DOCS_MANAGER
    ],
    AgentType.BACKEND_EXPERT: [
        AgentType.FRONTEND_EXPERT,
        AgentType.QUALITY_GATE,
        AgentType.DOCS_MANAGER
    ],
    AgentType.FRONTEND_EXPERT: [
        AgentType.QUALITY_GATE,
        AgentType.DOCS_MANAGER
    ],
    AgentType.QUALITY_GATE: [
        AgentType.DOCS_MANAGER
    ]
}


@dataclass
class SessionContext:
    session_id: str
    model: str
    context_position: float
    available_tokens: int
    user_language: str = "en"


@dataclass
class TaskContext:
    spec_id: str
    current_phase: str
    completed_steps: List[str]
    next_step: str
    key_artifacts: Dict[str, str] = field(default_factory=dict)


@dataclass
class RecoveryInfo:
    last_checkpoint: str
    recovery_tokens_reserved: int
    session_fork_available: bool = True


@dataclass
class HandoffPackage:
    handoff_id: str
    from_agent: AgentType
    to_agent: AgentType
    session_context: SessionContext
    task_context: TaskContext
    recovery_info: RecoveryInfo
    created_at: str = field(default_factory=lambda: datetime.utcnow().isoformat())
    checksum: str = ""

    def __post_init__(self):
        if not self.checksum:
            self.checksum = self._calculate_checksum()

    def _calculate_checksum(self) -> str:
        """Calculate checksum for integrity verification."""
        content = f"{self.handoff_id}{self.from_agent.value}{self.to_agent.value}"
        content += f"{self.task_context.spec_id}{self.created_at}"
        return hashlib.sha256(content.encode()).hexdigest()[:16]


class HandoffError(Exception):
    """Base exception for handoff errors."""
    pass


class AgentCompatibilityError(HandoffError):
    """Raised when agents cannot cooperate."""
    pass


class TokenBudgetError(HandoffError):
    """Raised when token budget is insufficient."""
    pass


class IntegrityError(HandoffError):
    """Raised when handoff integrity check fails."""
    pass


class HandoffManager:
    """Manages multi-agent handoff with validation."""

    MINIMUM_SAFE_TOKENS = 30000

    def __init__(self):
        self.handoff_history: List[HandoffPackage] = []

    def can_agents_cooperate(
        self,
        from_agent: AgentType,
        to_agent: AgentType
    ) -> bool:
        """Check if agents can cooperate based on compatibility matrix."""
        compatible = AGENT_COMPATIBILITY.get(from_agent, [])
        return to_agent in compatible

    def create_handoff(
        self,
        from_agent: AgentType,
        to_agent: AgentType,
        session_context: SessionContext,
        task_context: TaskContext,
        recovery_info: RecoveryInfo
    ) -> HandoffPackage:
        """Create a validated handoff package."""
        import uuid

        # Validate compatibility
        if not self.can_agents_cooperate(from_agent, to_agent):
            raise AgentCompatibilityError(
                f"Agent {from_agent.value} cannot hand off to {to_agent.value}"
            )

        # Validate token budget
        if session_context.available_tokens < self.MINIMUM_SAFE_TOKENS:
            raise TokenBudgetError(
                f"Insufficient tokens: {session_context.available_tokens} < "
                f"{self.MINIMUM_SAFE_TOKENS} required"
            )

        handoff = HandoffPackage(
            handoff_id=f"hoff_{uuid.uuid4().hex[:8]}",
            from_agent=from_agent,
            to_agent=to_agent,
            session_context=session_context,
            task_context=task_context,
            recovery_info=recovery_info
        )

        self.handoff_history.append(handoff)
        return handoff

    def validate_handoff(self, package: HandoffPackage) -> bool:
        """Validate handoff package integrity."""
        # Verify checksum
        expected_checksum = package._calculate_checksum()
        if package.checksum != expected_checksum:
            raise IntegrityError("Handoff checksum mismatch")

        # Verify agent compatibility
        if not self.can_agents_cooperate(package.from_agent, package.to_agent):
            raise AgentCompatibilityError("Agents cannot cooperate")

        # Verify token budget
        if package.session_context.available_tokens < self.MINIMUM_SAFE_TOKENS:
            # Trigger compression instead of failing
            return self._trigger_context_compression(package)

        return True

    def _trigger_context_compression(self, package: HandoffPackage) -> bool:
        """Compress context when token budget is low."""
        print(f"Compressing context for handoff {package.handoff_id}")

        # Apply progressive summarization
        # In practice, this would compress task_context.key_artifacts

        return True

    def extract_minimal_context(
        self,
        full_context: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Extract only critical context for handoff."""
        priority_fields = [
            "spec_id",
            "current_phase",
            "next_step",
            "critical_decisions",
            "blocking_issues"
        ]

        return {k: v for k, v in full_context.items() if k in priority_fields}

    def get_handoff_summary(self) -> Dict[str, Any]:
        """Get summary of all handoffs in session."""
        return {
            "total_handoffs": len(self.handoff_history),
            "handoffs": [
                {
                    "id": h.handoff_id,
                    "from": h.from_agent.value,
                    "to": h.to_agent.value,
                    "spec": h.task_context.spec_id,
                    "timestamp": h.created_at
                }
                for h in self.handoff_history
            ]
        }


# Usage example
manager = HandoffManager()

# Create handoff from spec-builder to ddd-implementer
session_ctx = SessionContext(
    session_id="sess_abc123",
    model="claude-sonnet-4-5-20250929",
    context_position=42.5,
    available_tokens=115000,
    user_language="ko"
)

task_ctx = TaskContext(
    spec_id="SPEC-001",
    current_phase="planning_complete",
    completed_steps=["requirement_analysis", "spec_creation", "architecture_design"],
    next_step="write_tests",
    key_artifacts={
        "spec_document": ".moai/specs/SPEC-001.md",
        "architecture": ".moai/architecture/SPEC-001.mermaid"
    }
)

recovery = RecoveryInfo(
    last_checkpoint=datetime.utcnow().isoformat(),
    recovery_tokens_reserved=55000,
    session_fork_available=True
)

try:
    handoff = manager.create_handoff(
        from_agent=AgentType.SPEC_BUILDER,
        to_agent=AgentType.DDD_IMPLEMENTER,
        session_context=session_ctx,
        task_context=task_ctx,
        recovery_info=recovery
    )

    print(f"Handoff created: {handoff.handoff_id}")
    print(f"Checksum: {handoff.checksum}")

    # Validate before sending to next agent
    is_valid = manager.validate_handoff(handoff)
    print(f"Handoff valid: {is_valid}")

except HandoffError as e:
    print(f"Handoff failed: {e}")
```

**Explanation**: This pattern implements robust multi-agent handoffs with compatibility checking, token budget validation, integrity verification, and context compression for efficient agent coordination.

---

## Common Patterns

### Pattern 1: Aggressive /clear Strategy

Execute /clear at strategic checkpoints:

```python
class ClearStrategy:
    """Strategy for executing /clear at optimal points."""

    CLEAR_TRIGGERS = {
        "post_spec_creation": True,
        "token_threshold_150k": True,
        "message_count_50": True,
        "phase_transition": True,
        "model_switch": True
    }

    def should_clear(self, context: Dict[str, Any]) -> tuple[bool, str]:
        """Determine if /clear should be executed."""

        # Check token threshold
        if context.get("token_usage", 0) > 150000:
            return True, "Token threshold exceeded (>150K)"

        # Check message count
        if context.get("message_count", 0) > 50:
            return True, "Message count exceeded (>50)"

        # Check phase transition
        if context.get("phase_changed", False):
            return True, "Phase transition detected"

        # Check post-SPEC creation
        if context.get("spec_just_created", False):
            return True, "SPEC creation completed"

        return False, "No clear needed"

    def prepare_clear_context(
        self,
        session: SessionState
    ) -> Dict[str, Any]:
        """Prepare minimal context to preserve across /clear."""
        return {
            "session_id": session.session_id,
            "spec_id": session.work_state.current_spec,
            "phase": session.work_state.phase.value,
            "reload_sequence": [
                "CLAUDE.md",
                f".moai/specs/{session.work_state.current_spec}.md",
                "src/main.py"  # Current working file
            ],
            "preserved_decisions": session.work_state.artifacts.get("decisions", [])
        }
```

### Pattern 2: Progressive Summarization

Compress context while preserving key information:

```python
class ProgressiveSummarizer:
    """Compress context progressively to save tokens."""

    def summarize_conversation(
        self,
        messages: List[Dict],
        target_ratio: float = 0.3
    ) -> str:
        """Summarize conversation to target ratio."""

        # Extract key information
        decisions = self._extract_decisions(messages)
        code_changes = self._extract_code_changes(messages)
        issues = self._extract_issues(messages)

        # Create compressed summary
        summary = f"""
## Conversation Summary

### Key Decisions
{self._format_list(decisions)}

### Code Changes
{self._format_list(code_changes)}

### Open Issues
{self._format_list(issues)}

### Reference
Original conversation: {len(messages)} messages
Compression ratio: {target_ratio * 100:.0f}%
"""
        return summary

    def _extract_decisions(self, messages: List[Dict]) -> List[str]:
        """Extract decision points from conversation."""
        decisions = []
        decision_markers = ["decided", "agreed", "will use", "chosen"]

        for msg in messages:
            content = msg.get("content", "").lower()
            if any(marker in content for marker in decision_markers):
                decisions.append(self._extract_sentence(msg["content"]))

        return decisions[:5]  # Top 5 decisions

    def _extract_code_changes(self, messages: List[Dict]) -> List[str]:
        """Extract code change summaries."""
        changes = []
        for msg in messages:
            if "```" in msg.get("content", ""):
                # Has code block - likely a change
                changes.append(f"Modified: {msg.get('file', 'unknown')}")
        return changes

    def _extract_issues(self, messages: List[Dict]) -> List[str]:
        """Extract open issues."""
        issues = []
        issue_markers = ["todo", "fixme", "issue", "problem", "bug"]

        for msg in messages:
            content = msg.get("content", "").lower()
            if any(marker in content for marker in issue_markers):
                issues.append(self._extract_sentence(msg["content"]))

        return issues

    def _extract_sentence(self, text: str) -> str:
        """Extract first meaningful sentence."""
        sentences = text.split('.')
        return sentences[0][:100] if sentences else text[:100]

    def _format_list(self, items: List[str]) -> str:
        """Format items as bullet list."""
        if not items:
            return "- None"
        return "\n".join(f"- {item}" for item in items)
```

### Pattern 3: Context Tag References

Use efficient references instead of inline content:

```python
class ContextTagManager:
    """Manage context with efficient tag references."""

    def __init__(self):
        self.tags: Dict[str, str] = {}

    def register_tag(self, tag_id: str, content: str) -> str:
        """Register content with a tag reference."""
        self.tags[tag_id] = content
        return f"@{tag_id}"

    def resolve_tag(self, tag_ref: str) -> Optional[str]:
        """Resolve a tag reference to content."""
        if tag_ref.startswith("@"):
            tag_id = tag_ref[1:]
            return self.tags.get(tag_id)
        return None

    def create_minimal_reference(
        self,
        full_context: Dict[str, Any]
    ) -> Dict[str, str]:
        """Create minimal references to full context."""
        references = {}

        for key, value in full_context.items():
            if isinstance(value, str) and len(value) > 200:
                # Store full content, return reference
                tag_id = f"{key.upper()}-001"
                self.register_tag(tag_id, value)
                references[key] = f"@{tag_id}"
            else:
                references[key] = value

        return references


# Usage
tag_manager = ContextTagManager()

# Instead of inline content (high token cost)
# "The user configuration from the previous 20 messages..."

# Use efficient reference (low token cost)
tag_manager.register_tag("CONFIG-001", full_config_content)
reference = "@CONFIG-001"  # 10 tokens vs 500+ tokens
```

---

## Anti-Patterns (Patterns to Avoid)

### Anti-Pattern 1: Ignoring Token Warnings

**Problem**: Continuing work without adddessing token warnings.

```python
# Incorrect approach
if token_usage > 150000:
    print("Warning: High token usage")
    # Continue working anyway - leads to context overflow
    continue_work()
```

**Solution**: Take immediate action on warnings.

```python
# Correct approach
if token_usage > 150000:
    logger.warning("Token warning triggered")
    # Create checkpoint and clear
    checkpoint = save_current_state()
    execute_clear()
    restore_essential_context(checkpoint)
```

### Anti-Pattern 2: Full Context in Handoffs

**Problem**: Passing complete context between agents wastes tokens.

```python
# Incorrect approach
handoff = {
    "full_conversation": all_messages,  # 50K tokens
    "all_files_content": file_contents,  # 100K tokens
    "complete_history": history          # 30K tokens
}
```

**Solution**: Pass only critical context.

```python
# Correct approach
handoff = {
    "spec_id": "SPEC-001",
    "current_phase": "implementation",
    "next_step": "write_tests",
    "key_decisions": ["Use JWT", "PostgreSQL"],
    "file_references": ["@API-SCHEMA", "@DB-MODEL"]
}
```

### Anti-Pattern 3: No Session Persistence

**Problem**: Losing work progress on interruption.

```python
# Incorrect approach
# No state saved - all progress lost on /clear or interruption
work_in_progress = process_spec(spec_id)
# Connection lost - work lost
```

**Solution**: Persist state continuously.

```python
# Correct approach
session = SessionManager()
session.create_checkpoint("pre_processing")

work_in_progress = process_spec(spec_id)
session.update_work_state(completed_step="processing_done")
session.save()  # State preserved

# After interruption
resumed = session.load_session(session_id)
# Continue from checkpoint
```

---

## Workflow Integration

### SPEC-First Workflow with Context Management

```python
# Complete workflow with context optimization

# Phase 1: Planning (uses ~40K tokens)
analysis = Task(subagent_type="spec-builder", prompt="Analyze: user auth")
session.update_work_state(phase=SessionPhase.PLANNING)

# Mandatory /clear after planning (saves 45-50K tokens)
clear_context = session.prepare_for_clear()
execute_clear()
restore_from_checkpoint(clear_context)

# Phase 2: Implementation (fresh 200K budget)
implementation = Task(
    subagent_type="ddd-implementer",
    prompt=f"Implement: {clear_context['spec_id']}"
)
session.update_work_state(phase=SessionPhase.IMPLEMENTATION)

# Monitor and clear if needed
if token_usage > 150000:
    clear_and_resume()

# Phase 3: Documentation
docs = Task(subagent_type="docs-manager", prompt="Generate docs")
session.update_work_state(phase=SessionPhase.DOCUMENTATION)
```

---

*For detailed implementation patterns and module references, see the `modules/` directory.*
