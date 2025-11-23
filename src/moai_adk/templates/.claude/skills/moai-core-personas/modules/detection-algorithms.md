# Expertise Detection & Persona Selection

Algorithms for detecting user expertise level and selecting appropriate persona dynamically.

## Expertise Level Detection

### Signal Categories

**Beginner Indicators**:

```python
beginner_signals = [
    "repeated_questions",          # Asking same thing multiple times
    "step_by_step_requests",       # Wants detailed walkthroughs
    "why_questions",               # "Why does this work?"
    "help_requests",               # Explicit help seeking
    "clarification_requests",      # "Can you explain differently?"
    "basic_terminology",           # Using general vs. technical terms
    "unknown_features",            # Unaware of available features
    "error_confusion",             # Confused by error messages
    "syntax_questions",            # Basic language questions
    "tutorial_seeking"             # "Show me how to..."
]
```

**Expert Indicators**:

```python
expert_signals = [
    "direct_commands",             # "Just do X"
    "technical_precision",         # Uses exact terminology
    "efficiency_keywords",         # "quick", "fast", "optimize"
    "command_line_usage",          # Uses CLI directly
    "code_review_comments",        # Reviews/critiques code
    "architecture_discussion",     # Discusses design patterns
    "performance_queries",         # Asks about optimization
    "edge_case_handling",          # Considers corner cases
    "technology_comparison",       # Compares tools/approaches
    "internal_api_knowledge"       # Knows undocumented features
]
```

**Intermediate Indicators**:

```python
intermediate_signals = [
    "mixed_signal_set",            # Both beginner and expert signals
    "contextual_learning",         # Learns quickly within context
    "specific_questions",          # Knows what to ask
    "problem_solving",             # Tries before asking
    "framework_knowledge",         # Understands project structure
]
```

### Detection Algorithm

```python
def detect_expertise_level(session_signals: list) -> dict:
    """
    Detect user expertise level from session signals.

    Returns:
        {
            "level": "beginner|intermediate|expert",
            "confidence": 0.0-1.0,
            "beginner_score": float,
            "expert_score": float
        }
    """

    # Count signal types
    beginner_count = sum(1 for s in session_signals
                        if s.type in beginner_signals)
    expert_count = sum(1 for s in session_signals
                      if s.type in expert_signals)
    intermediate_count = sum(1 for s in session_signals
                            if s.type in intermediate_signals)

    total = len(session_signals)

    # Calculate scores
    beginner_score = beginner_count / total if total > 0 else 0.0
    expert_score = expert_count / total if total > 0 else 0.0
    intermediate_score = intermediate_count / total if total > 0 else 0.0

    # Determine level
    max_score = max(beginner_score, expert_score, intermediate_score)
    confidence = max_score

    if beginner_score > expert_score:
        level = "beginner"
    elif expert_score > beginner_score:
        level = "expert"
    else:
        level = "intermediate"

    return {
        "level": level,
        "confidence": confidence,
        "beginner_score": beginner_score,
        "expert_score": expert_score,
        "intermediate_score": intermediate_score
    }
```

### Example Scenarios

**Scenario 1: Beginner Detection**

```python
signals = [
    Signal("how do I start?", type="step_by_step_requests"),
    Signal("what's a SPEC?", type="basic_terminology"),
    Signal("can you explain that?", type="clarification_requests"),
    Signal("help please", type="help_requests")
]

result = detect_expertise_level(signals)
# Output:
# {
#     "level": "beginner",
#     "confidence": 0.95,
#     "beginner_score": 0.95,
#     "expert_score": 0.05
# }
```

**Scenario 2: Expert Detection**

```python
signals = [
    Signal("Refactor with monoidal pattern", type="architecture_discussion"),
    Signal("Optimize cache hit ratio", type="performance_queries"),
    Signal("Check edge case with null values", type="edge_case_handling"),
    Signal("Implement using TDD", type="technical_precision")
]

result = detect_expertise_level(signals)
# Output:
# {
#     "level": "expert",
#     "confidence": 0.92,
#     "beginner_score": 0.05,
#     "expert_score": 0.92
# }
```

---

## Multi-Factor Persona Selection

### Selection Priority Hierarchy

```python
def select_persona(user_request, session_context, project_config) -> Persona:
    """
    Select persona using multi-factor priority system.

    Priority 1: Explicit triggers (highest weight)
    Priority 2: Content keywords (high weight)
    Priority 3: Expertise detection (medium weight)
    Priority 4: Default fallback (lowest weight)
    """

    # Priority 1: Explicit command type
    if user_request.type == "alfred_command":
        # /moai: commands = ProjectManager
        return ProjectManager()

    if project_config.get("team_mode", False):
        # Team mode = CollaborationCoordinator
        return CollaborationCoordinator()

    # Priority 2: Content keyword detection
    request_text = user_request.text.lower()

    mentor_keywords = ["how", "why", "explain", "help me understand",
                       "step by step", "teach me", "what's"]
    if any(kw in request_text for kw in mentor_keywords):
        return TechnicalMentor()

    coach_keywords = ["quick", "fast", "just do it", "no fluff",
                      "direct", "tldr", "immediately"]
    if any(kw in request_text for kw in coach_keywords):
        return EfficiencyCoach()

    # Priority 3: Expertise level detection
    expertise = detect_expertise_level(session_context.signals)

    if expertise["confidence"] < 0.6:
        # Low confidence = safe default
        return TechnicalMentor()

    if expertise["level"] == "beginner":
        return TechnicalMentor()
    elif expertise["level"] == "expert":
        return EfficiencyCoach()
    elif expertise["level"] == "intermediate":
        return TechnicalMentor()  # Conservative default

    # Priority 4: Default fallback
    return TechnicalMentor()  # Safe default
```

### Confidence Thresholds

```python
CONFIDENCE_THRESHOLDS = {
    "high": 0.85,        # >85% = confident decision
    "medium": 0.65,      # 65-85% = moderate confidence
    "low": 0.5,          # <50% = use default
    "uncertain": 0.3     # <30% = needs more signals
}

# Example usage
if expertise["confidence"] > CONFIDENCE_THRESHOLDS["high"]:
    # Use detected persona confidently
    return select_by_expertise(expertise["level"])
elif expertise["confidence"] > CONFIDENCE_THRESHOLDS["medium"]:
    # Blend detected with default
    return blend_personas(detected, TechnicalMentor, 0.6)
else:
    # Use safe default
    return TechnicalMentor()
```

---

## Signal Collection

### Signal Sources

```python
class SignalCollector:
    """Collect signals from user interactions."""

    def collect_signals(self, session) -> list:
        """Gather signals from multiple sources."""

        signals = []

        # From request text
        signals.extend(self.extract_text_signals(session.request.text))

        # From interaction history
        signals.extend(self.extract_history_signals(session.history))

        # From command usage
        signals.extend(self.extract_command_signals(session.commands))

        # From code patterns
        signals.extend(self.extract_code_signals(session.code))

        # From response satisfaction
        signals.extend(self.extract_satisfaction_signals(session.feedback))

        return signals

    def extract_text_signals(self, text: str) -> list:
        """Extract signals from request text."""

        signals = []
        text_lower = text.lower()

        # Check for beginner patterns
        for keyword in beginner_signals:
            if keyword in text_lower:
                signals.append(Signal(text, type=keyword))

        # Check for expert patterns
        for keyword in expert_signals:
            if keyword in text_lower:
                signals.append(Signal(text, type=keyword))

        return signals

    def extract_history_signals(self, history: list) -> list:
        """Extract signals from interaction patterns."""

        signals = []

        # Repeated questions = beginner signal
        unique_questions = len(set(h.question for h in history))
        total_questions = len(history)

        if unique_questions / total_questions < 0.6:
            signals.append(Signal("repeated_questions",
                                 type="repeated_questions"))

        # Rapid learning = expert signal
        if self.learning_speed(history) > 0.8:
            signals.append(Signal("rapid_learning",
                                 type="technical_precision"))

        return signals
```

---

## Real-Time Adjustment

### Continuous Learning

```python
class ExpertiseTracker:
    """Track expertise changes in real-time."""

    def __init__(self, user_id):
        self.user_id = user_id
        self.expertise_history = []
        self.signal_window = 20  # Track last 20 signals

    def update_expertise(self, new_signal):
        """Update expertise estimate with new signal."""

        self.expertise_history.append(new_signal)

        # Keep only recent signals
        recent_signals = self.expertise_history[-self.signal_window:]

        # Recalculate expertise
        expertise = detect_expertise_level(recent_signals)

        # Check for significant change
        previous_expertise = self.expertise_history[-2] if len(
            self.expertise_history) > 1 else expertise

        if expertise["level"] != previous_expertise["level"]:
            # Expertise level changed
            return {
                "changed": True,
                "from": previous_expertise["level"],
                "to": expertise["level"],
                "confidence": expertise["confidence"]
            }

        return {"changed": False}
```

---

## Selection Factors Summary

| Priority | Factor            | Weight | Examples                |
| -------- | ----------------- | ------ | ----------------------- |
| 1        | Explicit commands | 0.50   | /moai:, team_mode       |
| 2        | Content keywords  | 0.30   | how, why, quick, fast   |
| 3        | Expertise level   | 0.15   | Beginner/expert signals |
| 4        | Session history   | 0.05   | Previous preferences    |

---

## Best Practices

✅ **DO**:

- Collect signals from multiple sources
- Use confidence thresholds before deciding
- Allow for expertise level changes
- Override automatic detection if user prefers
- Log detection decisions for analysis
- Adjust thresholds based on performance metrics

❌ **DON'T**:

- Over-weight single signals
- Lock into initial expertise level
- Ignore explicit user requests
- Use only text keywords
- Assume static expertise

---

**Updated**: 2025-11-23 | **Algorithm Version**: 2.0 | **Signal Types**: 20+
