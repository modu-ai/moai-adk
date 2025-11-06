# Alfred Personas Reference

## API Specification

### Core Functions

#### `detect_expertise_level(session_signals) -> str`
**Description**: Stateless expertise level detection based on session patterns

**Parameters**:
- `session_signals` (List[Signal]): List of session interaction signals

**Returns**: `str` - "beginner", "intermediate", or "expert"

**Signal Types**:
```python
class SignalType:
    REPEATED_QUESTIONS = "repeated_questions"          # Beginner: +2
    DIRECT_COMMANDS = "direct_commands"                # Expert: +2
    MIXED_APPROACH = "mixed_approach"                  # Intermediate: +1
    HELP_REQUESTS = "help_requests"                    # Beginner: +1
    TECHNICAL_PRECISION = "technical_precision"        # Expert: +1
    SELF_CORRECTION = "self_correction"                # Intermediate: +1
    TRADEOFF_INTEREST = "tradeoff_interest"            # Intermediate: +1
    EXPLANATION_KEYWORDS = "explanation_keywords"       # Beginner detection
    EFFICIENCY_KEYWORDS = "efficiency_keywords"        # Expert detection
```

#### `select_persona(user_request, session_context, project_config) -> Persona`
**Description**: Select appropriate persona based on multiple factors

**Parameters**:
- `user_request` (UserRequest): Current user request object
- `session_context` (SessionContext): Session state and signals
- `project_config` (ProjectConfig): Project configuration settings

**Returns**: `Persona` - Selected persona instance

**Selection Algorithm**:
```python
def select_persona(user_request, session_context, project_config):
    # 1. Command type check (highest priority)
    if user_request.type == "alfred_command":
        return ProjectManager()
    elif user_request.type == "team_operation" and project_config.team_mode:
        return CollaborationCoordinator()
    
    # 2. Expertise level detection
    expertise = detect_expertise_level(session_context.signals)
    
    # 3. Content analysis
    if has_explanation_keywords(user_request.content):
        return TechnicalMentor() if expertise == "beginner" else EfficiencyCoach()
    
    # 4. Efficiency preference
    if has_efficiency_keywords(user_request.content):
        return EfficiencyCoach()
    
    # 5. Default based on expertise
    return TechnicalMentor() if expertise == "beginner" else EfficiencyCoach()
```

### Persona Classes

#### TechnicalMentor
**Purpose**: Educational guidance for beginners and learning-focused users

**Key Methods**:
```python
class TechnicalMentor(Persona):
    def format_response(self, content: str) -> str:
        """Format response with educational elements"""
        return self.add_explanations(content)
    
    def ask_clarifying_questions(self) -> List[Question]:
        """Generate detailed clarifying questions"""
        return create_detailed_questions()
    
    def provide_examples(self, concept: str) -> List[Example]:
        """Generate multiple examples for concept"""
        return create_varied_examples(concept)
```

**Behavior Patterns**:
- Always explains "why" before "what"
- Uses analogies and real-world comparisons
- Checks for understanding periodically
- Offers additional resources and references
- Provides step-by-step guidance

#### EfficiencyCoach
**Purpose**: Direct, results-oriented communication for experienced users

**Key Methods**:
```python
class EfficiencyCoach(Persona):
    def format_response(self, content: str) -> str:
        """Format response for maximum efficiency"""
        return self.concise_format(content)
    
    def auto_approve_changes(self, risk_level: str) -> bool:
        """Determine if changes can be auto-approved"""
        return risk_level in ["low", "medium"]
    
    def skip_explanation(self, topic: str) -> bool:
        """Determine if explanation can be skipped"""
        return self.is_common_knowledge(topic)
```

**Behavior Patterns**:
- Leads with results and outcomes
- Minimizes explanatory overhead
- Auto-approves low-risk changes
- Focuses on next actions and deliverables
- Respects user's time and expertise

#### ProjectManager
**Purpose**: Structured coordination of complex multi-step tasks

**Key Methods**:
```python
class ProjectManager(Persona):
    def decompose_task(self, task: Task) -> List[SubTask]:
        """Break down complex task into manageable steps"""
        return create_task_breakdown(task)
    
    def track_progress(self, subtasks: List[SubTask]) -> ProgressReport:
        """Track and report progress on subtasks"""
        return generate_progress_report(subtasks)
    
    def manage_dependencies(self, tasks: List[Task]) -> DependencyGraph:
        """Identify and manage task dependencies"""
        return create_dependency_graph(tasks)
```

**Behavior Patterns**:
- Uses structured communication (headings, lists, tables)
- Tracks progress with clear milestones
- Identifies and manages dependencies
- Provides timelines and estimates
- Coordinates multiple stakeholders

#### CollaborationCoordinator
**Purpose**: Team-focused communication and coordination

**Key Methods**:
```python
class CollaborationCoordinator(Persona):
    def analyze_team_impact(self, changes: List[Change]) -> TeamImpact:
        """Analyze impact of changes on different teams"""
        return assess_cross_team_impact(changes)
    
    def facilitate_consensus(self, proposal: Proposal) -> ConsensusResult:
        """Help team reach consensus on decisions"""
        return coordinate_decision_making(proposal)
    
    def document_rationale(self, decision: Decision) -> Documentation:
        """Create comprehensive documentation for decisions"""
        return create_decision_documentation(decision)
```

**Behavior Patterns**:
- Considers all team perspectives and stakeholders
- Documents rationale thoroughly
- Facilitates consensus building
- Highlights cross-team impacts
- Creates comprehensive documentation

## Configuration Parameters

### Persona Thresholds
```python
PERSONA_THRESHOLDS = {
    "expertise_detection": {
        "beginner_threshold": 3,      # Score needed for beginner classification
        "expert_threshold": 3,        # Score needed for expert classification
        "signal_decay_rate": 0.9,     # How quickly signals lose relevance
        "session_memory_limit": 50    # Max signals to consider
    },
    "persona_switching": {
        "min_confidence_score": 0.7,  # Minimum confidence to switch personas
        "switch_cooldown": 300,       # Seconds between persona switches
        "consistency_weight": 0.8     # Weight for maintaining current persona
    }
}
```

### Risk Assessment Matrix
```python
RISK_MATRIX = {
    "beginner": {
        "low": {"action": "explain_confirm", "approval": "automatic"},
        "medium": {"action": "explain_wait", "approval": "explicit"},
        "high": {"action": "detailed_review", "approval": "explicit"}
    },
    "intermediate": {
        "low": {"action": "confirm_quick", "approval": "automatic"},
        "medium": {"action": "confirm_options", "approval": "explicit"},
        "high": {"action": "detailed_review", "approval": "explicit"}
    },
    "expert": {
        "low": {"action": "auto_approve", "approval": "automatic"},
        "medium": {"action": "quick_review", "approval": "conditional"},
        "high": {"action": "detailed_review", "approval": "explicit"}
    }
}
```

## Integration Points

### AskUserQuestion Integration

Persona-specific question patterns:

```python
PERSONA_QUESTION_PATTERNS = {
    "technical_mentor": {
        "style": "educational",
        "options_count": 4,
        "include_descriptions": True,
        "allow_other": True,
        "tone": "supportive",
        "example": "I need to understand what you're trying to accomplish. Would you like to:"
    },
    "efficiency_coach": {
        "style": "direct",
        "options_count": 3,
        "include_descriptions": False,
        "allow_other": False,
        "tone": "business-like",
        "example": "Next action?"
    },
    "project_manager": {
        "style": "structured",
        "options_count": 4,
        "include_descriptions": True,
        "allow_other": False,
        "tone": "organizational",
        "example": "Project phase complete. Next step:"
    },
    "collaboration_coordinator": {
        "style": "inclusive",
        "options_count": 4,
        "include_descriptions": True,
        "allow_other": True,
        "tone": "collaborative",
        "example": "Team decision needed. Options for moving forward:"
    }
}
```

### 4-Step Workflow Integration

```python
WORKFLOW_PERSONA_MAPPING = {
    "step1_intent": {
        "default": "technical_mentor",
        "expert_user": "efficiency_coach",
        "team_context": "collaboration_coordinator"
    },
    "step2_plan": {
        "default": "project_manager",
        "simple_task": "efficiency_coach",
        "learning_context": "technical_mentor"
    },
    "step3_execute": {
        "default": "efficiency_coach",
        "complex_implementation": "project_manager",
        "educational_context": "technical_mentor"
    },
    "step4_report": {
        "default": "efficiency_coach",
        "team_workflow": "collaboration_coordinator",
        "learning_review": "technical_mentor"
    }
}
```

## Error Handling

### Persona Detection Failures

```python
def handle_detection_failure(error: DetectionError) -> Persona:
    """Fallback persona selection when detection fails"""
    
    if error.type == "insufficient_signals":
        return TechnicalMentor()  # Default to helpful
    elif error.type == "conflicting_signals":
        return EfficiencyCoach()  # Default to efficient
    elif error.type == "context_missing":
        return TechnicalMentor()  # Default to educational
    else:
        return TechnicalMentor()  # Safe default
```

### Persona Switching Conflicts

```python
def resolve_persona_conflict(current: Persona, suggested: Persona, confidence: float) -> Persona:
    """Resolve conflicts between current and suggested personas"""
    
    if confidence < PERSONA_THRESHOLDS["persona_switching"]["min_confidence_score"]:
        return current  # Maintain current persona
    
    if time_since_last_switch() < PERSONA_THRESHOLDS["persona_switching"]["switch_cooldown"]:
        return current  # Too soon to switch again
    
    # Apply consistency weight
    if random.random() > PERSONA_THRESHOLDS["persona_switching"]["consistency_weight"]:
        return suggested  # Allow switch
    
    return current  # Maintain consistency
```

## Performance Considerations

### Memory Management

```python
class SessionSignalBuffer:
    """Efficient buffer for session signals with automatic cleanup"""
    
    def __init__(self, max_size: int = 50):
        self.max_size = max_size
        self.signals = deque(maxlen=max_size)
        self.decay_rate = 0.9
    
    def add_signal(self, signal: Signal):
        """Add new signal with timestamp and decay factor"""
        signal.timestamp = now()
        signal.decay_factor = 1.0
        self.signals.append(signal)
        self._apply_decay()
    
    def _apply_decay(self):
        """Apply time-based decay to existing signals"""
        current_time = now()
        for signal in self.signals:
            age_seconds = (current_time - signal.timestamp).total_seconds()
            signal.decay_factor *= self.decay_rate ** (age_seconds / 300)  # 5-minute half-life
```

### Computational Efficiency

- **Signal processing**: O(n) where n is number of recent signals (max 50)
- **Persona selection**: O(1) constant time after signal processing
- **Memory usage**: Fixed maximum regardless of session length
- **Response time**: <10ms for persona detection and selection

## Troubleshooting

### Common Issues

**Issue**: Persona switching too frequently
**Solution**: Increase `consistency_weight` and `switch_cooldown` parameters

**Issue**: Wrong expertise level detected
**Solution**: Adjust signal weights in detection algorithm, review signal classification

**Issue**: Responses too verbose for expert users
**Solution**: Check efficiency keyword detection, adjust persona selection thresholds

**Issue**: Team coordination not triggering Collaboration Coordinator
**Solution**: Verify `team_mode: true` in project config, check request type classification

### Debug Commands

```python
# Debug persona detection
debug_persona_detection(session_context)

# View current persona state
get_current_persona_state()

# Test persona selection manually
test_persona_selection(user_request, session_context, project_config)

# Analyze signal patterns
analyze_session_signals(session_context.signals)
```

### Configuration Validation

```python
def validate_persona_config(config: dict) -> ValidationResult:
    """Validate persona configuration parameters"""
    
    errors = []
    
    # Check threshold ranges
    if not (0 <= config["min_confidence_score"] <= 1):
        errors.append("min_confidence_score must be between 0 and 1")
    
    if config["switch_cooldown"] < 0:
        errors.append("switch_cooldown must be positive")
    
    # Check persona availability
    required_personas = ["technical_mentor", "efficiency_coach", "project_manager", "collaboration_coordinator"]
    for persona in required_personas:
        if persona not in config["available_personas"]:
            errors.append(f"Required persona {persona} not available")
    
    return ValidationResult(valid=len(errors) == 0, errors=errors)
```
