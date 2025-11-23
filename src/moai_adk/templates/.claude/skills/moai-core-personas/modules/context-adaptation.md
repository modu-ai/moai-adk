# Context-Aware Persona Adaptation

Adapt personas dynamically based on project context, team structure, and situational factors.

## Project Context Factors

### Context Dimensions

```python
project_context = {
    # Complexity metrics
    "codebase_size": "small|medium|large|enterprise",
    "architecture_complexity": "low|medium|high",

    # Team metrics
    "team_size": 1-100+,
    "team_experience": "junior|mixed|senior",
    "team_distribution": "local|remote|hybrid",

    # Timeline metrics
    "deadline_pressure": False|True,
    "timeline_urgency": "planned|normal|urgent|emergency",

    # Risk metrics
    "risk_level": "low|medium|high|critical",
    "breaking_changes": False|True,

    # Project type
    "project_type": "mvp|production|migration|refactor",
    "domain": "backend|frontend|devops|ml|web3"
}
```

### Adaptation Rules

```python
def adapt_to_project_context(persona, project_context: dict):
    """
    Adapt persona based on project context.
    Returns adjusted persona with context-aware attributes.
    """

    adapted = copy.deepcopy(persona)

    # Complexity-based adaptation
    complexity = project_context.get("architecture_complexity", "low")
    if complexity == "high":
        # High complexity needs more detail
        adapted.attributes["detail_level"] = "high"
        adapted.attributes["validation_freq"] = "high"
        adapted.attributes["review_thoroughness"] = "comprehensive"
        # Prefer TechnicalMentor for complex projects
        if adapted.name == "EfficiencyCoach":
            adapted = blend(TechnicalMentor(0.7), EfficiencyCoach(0.3))

    # Team size-based adaptation
    team_size = project_context.get("team_size", 1)
    if team_size > 5:
        # Larger teams need coordination
        adapted.attributes["doc_level"] = "comprehensive"
        adapted.attributes["stakeholder_awareness"] = True
        adapted.attributes["communication_breadth"] = "extensive"
        # Prefer CollaborationCoordinator for large teams
        if adapted.name != "CollaborationCoordinator":
            adapted = blend(adapted, CollaborationCoordinator(0.5), 0.7)

    # Deadline pressure adaptation
    if project_context.get("deadline_pressure"):
        # Tight deadlines need efficiency
        adapted.attributes["efficiency_focus"] = True
        adapted.attributes["brevity"] = True
        adapted.attributes["quick_decisions"] = True
        adapted.attributes["skip_non_essentials"] = True
        # Shift toward EfficiencyCoach
        if adapted.name == "TechnicalMentor":
            adapted = blend(TechnicalMentor(0.5), EfficiencyCoach(0.5))

    # Risk level adaptation
    risk = project_context.get("risk_level", "low")
    if risk == "high" or risk == "critical":
        # High risk needs validation
        adapted.attributes["validation_checks"] = "thorough"
        adapted.attributes["error_checking"] = "paranoid"
        adapted.attributes["review_depth"] = "comprehensive"
        adapted.attributes["approval_required"] = True
        # Prefer TechnicalMentor or ProjectManager
        adapted.attributes["cautious"] = True

    # Experience level adaptation
    exp = project_context.get("team_experience", "mixed")
    if exp == "junior":
        # Junior team needs guidance
        adapted.attributes["explanation_depth"] = "thorough"
        adapted.attributes["mentoring"] = True
        # Use TechnicalMentor
    elif exp == "senior":
        # Senior team can handle brevity
        adapted.attributes["explanation_depth"] = "minimal"
        adapted.attributes["assume_knowledge"] = True
        # Use EfficiencyCoach

    return adapted
```

---

## Situational Context Handling

### Urgency Levels

```python
URGENCY_LEVELS = {
    "planned": {
        "response_time_target": "180 seconds",
        "detail_level": "thorough",
        "collaboration": "extensive",
        "documentation": "comprehensive",
        "persona_preference": "TechnicalMentor|ProjectManager"
    },
    "normal": {
        "response_time_target": "120 seconds",
        "detail_level": "balanced",
        "collaboration": "moderate",
        "documentation": "standard",
        "persona_preference": "Any"
    },
    "urgent": {
        "response_time_target": "60 seconds",
        "detail_level": "minimal",
        "collaboration": "focused",
        "documentation": "brief",
        "persona_preference": "EfficiencyCoach"
    },
    "emergency": {
        "response_time_target": "15 seconds",
        "detail_level": "critical_only",
        "collaboration": "minimal",
        "documentation": "none",
        "persona_preference": "EfficiencyCoach"
    }
}
```

### Risk-Based Adaptation

```python
def adapt_to_risk_level(persona, risk_level):
    """Adjust persona based on risk."""

    adapted = copy.deepcopy(persona)

    if risk_level == "low":
        # Can be brief and efficient
        adapted.attributes["thoroughness"] = "standard"
        adapted.attributes["validation"] = "basic"

    elif risk_level == "medium":
        # Need moderate validation
        adapted.attributes["thoroughness"] = "moderate"
        adapted.attributes["validation"] = "moderate"
        adapted.attributes["review_required"] = True

    elif risk_level == "high":
        # Need comprehensive validation
        adapted.attributes["thoroughness"] = "comprehensive"
        adapted.attributes["validation"] = "thorough"
        adapted.attributes["peer_review"] = True
        adapted.attributes["testing_required"] = True

    elif risk_level == "critical":
        # Need maximum caution
        adapted.attributes["thoroughness"] = "paranoid"
        adapted.attributes["validation"] = "extensive"
        adapted.attributes["multi_review"] = True
        adapted.attributes["testing_required"] = "100% coverage"
        adapted.attributes["approval_layers"] = 3

    return adapted
```

---

## Real-World Scenario Examples

### Scenario 1: Large Team, High Complexity

```python
context = {
    "team_size": 15,
    "codebase_size": "enterprise",
    "architecture_complexity": "high",
    "deadline_pressure": True,
    "risk_level": "high"
}

# Select base persona
base = TechnicalMentor()

# Adapt to context
adapted = adapt_to_project_context(base, context)

# Result: Blended persona
# 60% TechnicalMentor (detailed)
# 30% ProjectManager (organized)
# 10% CollaborationCoordinator (team-aware)

# Response style:
# - Detailed explanations (team learning)
# - Structured format (coordination)
# - Comprehensive docs (large team)
# - Validation-focused (high risk)
```

**Example response**:
```
Feature Implementation Plan:

📋 Architecture Overview:
[Detailed explanation of design patterns]

👥 Team Assignments:
- Backend (5 engineers): API design
- Frontend (4 engineers): Component integration
- DevOps (2 engineers): Deployment pipeline

⚠️ Risk Assessment:
- Migration risk: HIGH
- Mitigation: Phased rollout, comprehensive testing
- Testing requirement: 95%+ coverage

📅 Timeline:
- Phase 1: Design review (4h)
- Phase 2: Implementation (16h)
- Phase 3: Testing (8h)
- Phase 4: Deployment (4h)
Total: 32 hours, 1 week
```

### Scenario 2: MVP, Tight Deadline

```python
context = {
    "team_size": 3,
    "project_type": "mvp",
    "deadline_pressure": True,
    "timeline_urgency": "urgent",
    "risk_level": "low",
    "team_experience": "senior"
}

# Select and adapt
base = EfficiencyCoach()
adapted = adapt_to_project_context(base, context)

# Result: 100% EfficiencyCoach
# - Minimal explanation
# - Direct action
# - Assume knowledge
# - Fast decisions
```

**Example response**:
```
MVP Build: 3 days

Day 1:
✅ Core API endpoints (4h)
✅ Database schema (2h)

Day 2:
✅ Frontend components (6h)
✅ Integration tests (2h)

Day 3:
✅ Polish & deploy (4h)
✅ Monitoring setup (2h)

Go live: Friday EOD
```

### Scenario 3: Refactoring, Risk Mitigation

```python
context = {
    "team_size": 8,
    "project_type": "refactor",
    "risk_level": "high",
    "breaking_changes": True,
    "deadline_pressure": False,
    "timeline_urgency": "planned"
}

# Select and adapt
base = ProjectManager()
adapted = adapt_to_project_context(base, context)

# Result: 100% ProjectManager (with risk adaptation)
# - Detailed planning
# - Comprehensive testing
# - Multi-phase rollout
# - Risk mitigation strategies
```

---

## Context Matrix

### Quick Reference

| Context | Prefer | Detail | Pace | Validation |
|---------|--------|--------|------|-----------|
| **Large team** | Coordinator | High | Moderate | Comprehensive |
| **Complex arch** | Mentor | High | Moderate | Thorough |
| **Tight deadline** | Coach | Low | Fast | Basic |
| **MVP** | Coach | Low | Fast | Minimal |
| **High risk** | Manager | High | Moderate | Paranoid |
| **Critical** | Manager | Maximum | Careful | Extensive |
| **Senior team** | Coach | Low | Fast | Standard |
| **Junior team** | Mentor | High | Patient | Thorough |

---

## Dynamic Adaptation Example

```python
class ContextAwarePersona:
    """Dynamically adapt persona as context changes."""

    def __init__(self):
        self.current_context = {}
        self.current_persona = None

    def update_context(self, new_context: dict):
        """Update project context and re-adapt persona."""

        context_changed = False
        for key, value in new_context.items():
            if self.current_context.get(key) != value:
                context_changed = True
                self.current_context[key] = value

        if context_changed:
            # Re-evaluate and adapt
            self.adapt_persona()

    def adapt_persona(self):
        """Evaluate context and select/adapt best persona."""

        # Evaluate all factors
        team_size = self.current_context.get("team_size", 1)
        complexity = self.current_context.get("architecture_complexity", "low")
        urgency = self.current_context.get("timeline_urgency", "normal")
        risk = self.current_context.get("risk_level", "low")

        # Determine best base persona
        if team_size > 5:
            base = CollaborationCoordinator()
        elif urgency == "urgent" or urgency == "emergency":
            base = EfficiencyCoach()
        elif complexity == "high" or risk == "high":
            base = TechnicalMentor()
        else:
            base = TechnicalMentor()  # Safe default

        # Apply context adaptations
        self.current_persona = adapt_to_project_context(base, self.current_context)

    def get_persona(self):
        """Return current adapted persona."""
        return self.current_persona
```

---

## Best Practices

✅ **DO**:
- Evaluate full context before selecting persona
- Re-evaluate when context changes
- Adapt gradually as complexity increases
- Document context assumptions
- Allow explicit context override
- Monitor adaptation effectiveness

❌ **DON'T**:
- Ignore project context
- Use fixed personas for all situations
- Assume team expertise
- Skip risk assessment
- Make assumptions about deadlines
- Ignore breaking changes

---

**Updated**: 2025-11-23 | **Context Factors**: 10+ | **Adaptation Rules**: Dynamic
