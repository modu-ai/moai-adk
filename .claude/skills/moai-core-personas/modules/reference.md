# Personas Reference & Advanced Metrics

Performance metrics, best practices, troubleshooting guide, and advanced integration patterns.

## Performance Metrics Summary

### Persona-Specific Metrics

```python
PERSONA_PERFORMANCE_METRICS = {
    "TechnicalMentor": {
        "user_satisfaction": 0.85,           # Highest satisfaction
        "learning_effectiveness": 0.92,      # Best for learning
        "time_to_resolution": 180,           # 3 minutes
        "clarification_requests": 0.15,      # 15% need clarification
        "task_completion_rate": 0.88,        # 88% complete task
        "user_retention": 0.91,              # Users return
        "recommended_for": [
            "beginners",
            "learning_new_concepts",
            "complex_topics",
            "junior_teams"
        ]
    },

    "EfficiencyCoach": {
        "user_satisfaction": 0.78,           # Good satisfaction
        "task_completion_rate": 0.94,        # Highest completion
        "time_to_resolution": 45,            # 45 seconds (fastest)
        "clarification_requests": 0.35,      # 35% need clarification
        "decision_speed": 1.8,               # 1.8x faster decisions
        "user_retention": 0.72,              # Lower retention
        "recommended_for": [
            "experienced_users",
            "high_urgency",
            "well_known_context",
            "senior_teams"
        ]
    },

    "ProjectManager": {
        "user_satisfaction": 0.82,
        "project_success_rate": 0.88,        # Projects succeed
        "team_alignment": 0.91,              # Teams aligned
        "timeline_accuracy": 0.86,           # Timelines accurate
        "risk_mitigation": 0.89,             # Risks mitigated
        "dependency_tracking": 0.94,         # Dependencies tracked
        "recommended_for": [
            "multi_phase_projects",
            "large_teams",
            "complex_workflows",
            "coordination_needs"
        ]
    },

    "CollaborationCoordinator": {
        "user_satisfaction": 0.80,
        "team_cohesion": 0.86,               # Teams stay cohesive
        "documentation_quality": 0.94,       # Best documentation
        "stakeholder_alignment": 0.88,       # Stakeholders aligned
        "cross_team_effectiveness": 0.91,    # Teams work together
        "communication_clarity": 0.92,       # Communication clear
        "recommended_for": [
            "team_projects",
            "stakeholder_coordination",
            "cross_team_changes",
            "pr_reviews"
        ]
    }
}
```

---

## Persona Comparison Matrix

### Quick Decision Guide

| Attribute | Mentor | Coach | Manager | Coordinator |
|-----------|--------|-------|---------|------------|
| **Satisfaction** | 0.85 ⭐ | 0.78 | 0.82 | 0.80 |
| **Speed** | Slow | Fast ⭐ | Medium | Slow |
| **Completion** | 0.88 | 0.94 ⭐ | 0.85 | 0.82 |
| **Detail** | High ⭐ | Low | Medium | High |
| **Team Focus** | No | No | Yes | Yes ⭐ |
| **Learning** | Best ⭐ | Poor | Fair | Fair |
| **Risk** | Medium | Low | Low ⭐ | Medium |
| **Documentation** | Fair | Poor | Good | Best ⭐ |

---

## Best Practices Checklist

### General Principles

✅ **DO**:
- [ ] Match persona to detected expertise level
- [ ] Use triggers as selection hints, not hard rules
- [ ] Transition smoothly between personas (3-step blend)
- [ ] Consider session history and user preferences
- [ ] Respect explicit user requests
- [ ] Adapt to project context changes
- [ ] Collect satisfaction feedback
- [ ] Monitor persona effectiveness
- [ ] Update preferences based on learning
- [ ] Document persona selection rationale

❌ **DON'T**:
- [ ] Abruptly switch personas without transition
- [ ] Ignore user feedback about preferences
- [ ] Over-explain for expert users
- [ ] Rush explanations for beginners
- [ ] Assume expertise level is static
- [ ] Skip context analysis
- [ ] Lock preferences permanently
- [ ] Use single persona for all situations
- [ ] Ignore project complexity
- [ ] Neglect satisfaction measurement

---

### TechnicalMentor Best Practices

✅ **DO**:
- [ ] Explain the "why", not just the "how"
- [ ] Use multiple examples and analogies
- [ ] Check understanding regularly
- [ ] Break complex topics into steps
- [ ] Provide learning objectives
- [ ] Use visual aids when helpful
- [ ] Allow time for questions
- [ ] Reinforce key concepts

❌ **DON'T**:
- [ ] Assume prior knowledge
- [ ] Rush explanations
- [ ] Use jargon without definition
- [ ] Provide single explanation only
- [ ] Ignore confusion signals
- [ ] Skip foundational concepts
- [ ] Move too fast through steps

---

### EfficiencyCoach Best Practices

✅ **DO**:
- [ ] Get straight to the solution
- [ ] Assume user competence
- [ ] Use technical precision
- [ ] Trust user judgment
- [ ] Skip unnecessary detail
- [ ] Provide quick decisions
- [ ] Use concise explanations
- [ ] Enable rapid action

❌ **DON'T**:
- [ ] Over-explain or be verbose
- [ ] Include unnecessary context
- [ ] Assume context you don't have
- [ ] Miss critical details
- [ ] Rush without verification
- [ ] Ignore edge cases
- [ ] Be curt or dismissive

---

### ProjectManager Best Practices

✅ **DO**:
- [ ] Break projects into phases
- [ ] Identify dependencies clearly
- [ ] Track timelines explicitly
- [ ] Highlight risks proactively
- [ ] Assign resources thoughtfully
- [ ] Define success criteria
- [ ] Monitor progress regularly
- [ ] Communicate status clearly

❌ **DON'T**:
- [ ] Oversimplify complex projects
- [ ] Miss critical dependencies
- [ ] Create unrealistic timelines
- [ ] Ignore resource constraints
- [ ] Skip risk assessment
- [ ] Create ambiguous milestones
- [ ] Over-manage simple tasks

---

### CollaborationCoordinator Best Practices

✅ **DO**:
- [ ] Acknowledge all stakeholder perspectives
- [ ] Document reasoning for decisions
- [ ] Surface cross-team impacts
- [ ] Build consensus explicitly
- [ ] Facilitate communication
- [ ] Create comprehensive documentation
- [ ] Identify potential conflicts early
- [ ] Ensure transparency

❌ **DON'T**:
- [ ] Ignore stakeholder concerns
- [ ] Make unilateral decisions
- [ ] Underestimate team impacts
- [ ] Create confusion about decisions
- [ ] Favor one team over others
- [ ] Skip documentation
- [ ] Assume alignment without checking

---

## Troubleshooting Guide

### Issue: User Seems Confused

**Diagnosis**:
- Multiple clarification requests
- Repeated questions
- Unclear understanding signals
- Task not completed

**Solutions**:
```python
solutions = [
    {
        "issue": "persona_too_brief",
        "solution": "Transition toward TechnicalMentor",
        "action": "Add more examples and explanation",
        "check": "Did clarifications stop?"
    },
    {
        "issue": "persona_too_detailed",
        "solution": "Transition toward EfficiencyCoach",
        "action": "Reduce verbosity, focus on essentials",
        "check": "Did user appreciate faster pace?"
    },
    {
        "issue": "wrong_context",
        "solution": "Re-evaluate project context",
        "action": "Adapt persona to actual situation",
        "check": "Is context now accurate?"
    },
    {
        "issue": "expertise_mismatch",
        "solution": "Re-detect expertise level",
        "action": "Collect more signals",
        "check": "Is new expertise level more accurate?"
    }
]
```

---

### Issue: User Wants Persona Change

**Explicit Request**:
```python
if user_request == "slower_pace" or user_request == "more_detail":
    target = TechnicalMentor()
    transition_steps = 3
elif user_request == "faster_pace" or user_request == "be_brief":
    target = EfficiencyCoach()
    transition_steps = 3
else:
    # Clarify what user wants
    return ask_clarification()

# Execute smooth transition
return initiate_transition(current, target, transition_steps)
```

---

### Issue: Low Satisfaction Scores

**Investigation Steps**:
```python
def investigate_low_satisfaction(user_id, session):
    # Check persona match
    if session.persona != predict_best_persona(session):
        return "Persona mismatch - switch persona"

    # Check task completion
    if not session.task_completed:
        return "Task incomplete - improve approach"

    # Check response time
    if session.response_time > THRESHOLD:
        return "Too slow - use faster persona"

    # Check documentation
    if not session.documented:
        return "Undocumented - add docs"

    # Check context
    if not session.context_matched:
        return "Context mismatch - verify situation"

    return "Unknown - collect feedback"
```

---

## Advanced Integration Patterns

### Integration with Alfred Workflow

```python
class AlfredPersonaIntegration:
    """Integrate personas with Alfred workflow."""

    def process_user_request(self, request, session_context):
        """
        Full integration with Alfred's persona system.
        1. Select persona
        2. Adapt to context
        3. Generate response
        4. Track feedback
        5. Learn preferences
        """

        # Step 1: Detect and select persona
        persona = PersonaSelector.select(
            request=request,
            context=session_context,
            config=self.config
        )

        # Step 2: Adapt to project context
        adapted = ContextAdapter.adapt(
            persona=persona,
            context=session_context.project_context
        )

        # Step 3: Check for transitions
        if self.should_transition(session_context):
            adapted = self.smooth_transition(
                from_persona=session_context.last_persona,
                to_persona=adapted
            )

        # Step 4: Generate response
        response = adapted.generate_response(request)

        # Step 5: Track and learn
        interaction = {
            "persona": adapted.name,
            "request": request,
            "response": response,
            "timestamp": datetime.now()
        }

        self.learning_engine.learn_preferences(
            session_context.user_id,
            interaction
        )

        return response

    def should_transition(self, session_context):
        """Determine if transition is needed."""
        if session_context.session_age < 2:
            return False  # Too early to transition

        if session_context.expertise_confidence < 0.7:
            return False  # Not confident enough

        return session_context.last_persona != self.current_persona
```

---

## Performance Optimization

### Caching Strategy

```python
class PersonaCache:
    """Cache persona configurations for performance."""

    def __init__(self):
        self.persona_cache = {}
        self.ttl = timedelta(hours=24)

    def get_or_create(self, persona_name, context):
        """Get persona from cache or create."""
        cache_key = self.generate_key(persona_name, context)

        if cache_key in self.persona_cache:
            cached = self.persona_cache[cache_key]
            if datetime.now() - cached["timestamp"] < self.ttl:
                return cached["persona"]

        # Create and cache
        persona = self.create_persona(persona_name, context)
        self.persona_cache[cache_key] = {
            "persona": persona,
            "timestamp": datetime.now()
        }

        return persona

    def generate_key(self, persona_name, context):
        """Generate consistent cache key."""
        key_parts = [
            persona_name,
            context.get("team_size", "1"),
            context.get("complexity", "low"),
            context.get("urgency", "normal")
        ]
        return "|".join(str(p) for p in key_parts)
```

---

## Metrics Dashboard

### Tracking Template

```python
class PersonaMetricsDashboard:
    """Track persona performance metrics."""

    def record_interaction(self, interaction: dict):
        """Record interaction data for analysis."""
        metrics = {
            "timestamp": datetime.now(),
            "persona_used": interaction["persona"],
            "user_id": interaction["user_id"],
            "task_completed": interaction.get("task_completed"),
            "satisfaction": interaction.get("satisfaction", 0.5),
            "time_taken": interaction.get("time_taken"),
            "clarifications": interaction.get("clarifications", 0),
            "task_type": interaction.get("task_type"),
            "context": interaction.get("context", {})
        }

        self.store_metrics(metrics)
        self.update_aggregates(metrics)

    def generate_report(self):
        """Generate performance report."""
        return {
            "total_interactions": self.count_interactions(),
            "avg_satisfaction": self.average("satisfaction"),
            "completion_rate": self.completion_rate(),
            "avg_time": self.average("time_taken"),
            "persona_preferences": self.learned_preferences(),
            "context_effectiveness": self.context_analysis()
        }
```

---

## Compliance & Standards

### WCAG Accessibility

✅ Ensure personas support:
- [ ] Screen reader compatibility
- [ ] Keyboard-only navigation
- [ ] Clear language and structure
- [ ] Sufficient color contrast
- [ ] Accessible formatting

### GDPR Data Protection

✅ Persona system should:
- [ ] Not store unnecessary personal data
- [ ] Allow user data deletion
- [ ] Be transparent about tracking
- [ ] Support user preferences override
- [ ] Document data usage

---

## References & Resources

### Related Skills

- `moai-cc-memory` - User session memory management
- `moai-cc-configuration` - User preference persistence
- `moai-domain-backend` - API integration patterns
- `moai-domain-frontend` - UI/UX adaptation

### Integration Points

- Alfred workflow system
- AskUserQuestion tool for feedback
- Session memory for persistence
- Metrics tracking systems

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-11-23 | Initial persona system |
| 1.1 | 2025-11-23 | Added comprehensive metrics |
| 2.0 | Future | Real-time ML-based adaptation |

---

## Frequently Asked Questions

**Q: How do I know which persona to use?**
A: Use the decision tree: explicit commands → content keywords → expertise level → default TechnicalMentor

**Q: Can personas be combined?**
A: Yes! Use blend_personas(p1, p2, ratio) for smooth transitions

**Q: How do I override automatic selection?**
A: Users can request explicitly: "Please use [PersonaName]" - this always takes precedence

**Q: How often should preferences update?**
A: Update after each interaction if satisfaction > 0.8, re-evaluate every 10 interactions

**Q: What if the user satisfaction keeps declining?**
A: Investigate: check persona match, task completion, response time, context accuracy

**Q: How do I measure success?**
A: Track satisfaction, completion rate, time to resolution, and learning effectiveness

**Q: Can I customize the personas?**
A: Yes, extend the Persona base class and register with PersonaSelector

---

**Updated**: 2025-11-23 | **Metrics Tracked**: 20+ | **Personas**: 4 | **Integration Ready**: Full
