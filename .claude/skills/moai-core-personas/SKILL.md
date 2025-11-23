---
name: moai-core-personas
description: Adaptive communication patterns and role selection based on user expertise level and request type
version: 1.1.0
modularized: true
tags:
  - architecture
  - personas
  - enterprise
  - framework
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**Name**: moai-core-personas
**Domain**: Adaptive Communication & User Interaction
**Freedom Level**: medium
**Target Users**: Alfred agents, communication coordinators, team leads
**Invocation**: Skill("moai-core-personas")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed guides)
**Last Updated**: 2025-11-23
**Modularized**: true

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Dynamically adapt Alfred's communication style based on user expertise and request context.

**4 Core Personas**:
- 🧑‍🏫 **Technical Mentor** - Detailed, educational (for beginners)
- ⚡ **Efficiency Coach** - Concise, direct (for experts)
- 📋 **Project Manager** - Structured planning (for coordination)
- 🤝 **Collaboration Coordinator** - Team-focused (for groups)

**Key Features**:
- Stateless expertise detection (no session data required)
- Multi-factor persona selection (triggers + context + expertise)
- Gradual persona transitions (smooth style shifts)
- Performance-optimized caching
- Integration with AskUserQuestion

---

## 📚 Core Patterns (5-10 minutes)

### Pattern 1: Expertise Detection

**Key Concept**: Identify user skill level from request patterns.

**Approach**:
```python
def detect_expertise_level(session_signals) -> str:
    beginner_indicators = [
        "repeated_questions", "step_by_step_requests", "why_questions"
    ]
    expert_indicators = [
        "direct_commands", "technical_precision", "efficiency_keywords"
    ]

    beginner_score = sum(1 for signal in session_signals
                        if signal.type in beginner_indicators)
    expert_score = sum(1 for signal in session_signals
                      if signal.type in expert_indicators)

    if beginner_score > expert_score:
        return "beginner"
    elif expert_score > beginner_score:
        return "expert"
    else:
        return "intermediate"
```

**Use Case**: Automatically selecting response complexity for first-time users vs. experienced developers.

### Pattern 2: Multi-Factor Persona Selection

**Key Concept**: Choose persona based on explicit triggers + content + expertise.

**Approach**:
```python
def select_persona(user_request, session_context, project_config):
    # Factor 1: Explicit triggers (highest priority)
    if user_request.type == "alfred_command":
        return ProjectManager()
    elif project_config.get("team_mode", False):
        return CollaborationCoordinator()

    # Factor 2: Content analysis
    if any(kw in user_request.text.lower() for kw in ["how", "why", "explain"]):
        return TechnicalMentor()
    elif any(kw in user_request.text.lower() for kw in ["quick", "fast", "just do"]):
        return EfficiencyCoach()

    # Factor 3: Expertise level
    expertise = detect_expertise_level(session_context.signals)
    if expertise == "beginner":
        return TechnicalMentor()
    elif expertise == "expert":
        return EfficiencyCoach()

    return TechnicalMentor()  # Safe default
```

**Use Case**: Routing user requests to appropriate communication style based on multiple signals.

### Pattern 3: Persona Communication Styles

**Key Concept**: Each persona defines its communication parameters.

**Approach**:
```python
class TechnicalMentor:
    """Educational communication for learners"""
    triggers = ["how", "why", "explain", "help me understand"]
    attributes = {
        "style": "educational",
        "explanation_depth": "thorough",
        "examples": "multiple",
        "pace": "patient",
        "check_understanding": True
    }

class EfficiencyCoach:
    """Direct communication for experts"""
    triggers = ["quick", "fast", "just do it"]
    attributes = {
        "style": "direct",
        "explanation_depth": "minimal",
        "examples": "focused",
        "pace": "rapid",
        "auto_approve": True
    }

class ProjectManager:
    """Structured planning communication"""
    triggers = ["/alfred:", "plan", "coordinate"]
    attributes = {
        "style": "structured",
        "format": "hierarchical",
        "tracking": "detailed",
        "timeline": "included"
    }

class CollaborationCoordinator:
    """Team-focused communication"""
    triggers = ["team", "PR", "review", "collaboration"]
    attributes = {
        "style": "comprehensive",
        "stakeholder_awareness": True,
        "documentation": "thorough"
    }
```

**Use Case**: Standardizing communication patterns across different interaction contexts.

### Pattern 4: Gradual Persona Transitions

**Key Concept**: Smoothly shift communication style as user expertise grows.

**Approach**:
```python
class PersonaTransition:
    def gradual_transition(self, from_persona, to_persona, steps=3):
        transition_steps = []

        for i in range(1, steps + 1):
            blend_ratio = i / steps
            blended_style = self.blend_personas(
                from_persona, to_persona, blend_ratio
            )
            transition_steps.append(blended_style)

        return transition_steps

    def blend_personas(self, persona1, persona2, ratio):
        blended = {}

        for attribute in ["style", "explanation_depth", "pace"]:
            if ratio <= 0.5:
                blended[attribute] = persona1.attributes[attribute]
            else:
                blended[attribute] = persona2.attributes[attribute]

        return blended
```

**Use Case**: Gradually shifting from detailed explanations to concise technical responses as users gain expertise.

### Pattern 5: Context-Aware Adaptation

**Key Concept**: Adjust persona based on project context and constraints.

**Approach**:
```python
def adapt_to_project_context(persona, project_context):
    adapted = copy.deepcopy(persona)

    # High complexity → more detailed
    if project_context.get("complexity") == "high":
        adapted.attributes["detail_level"] = "high"
        adapted.attributes["validation_frequency"] = "high"

    # Large team → more documentation
    if project_context.get("team_size", 0) > 5:
        adapted.attributes["documentation_level"] = "comprehensive"

    # Tight deadline → efficiency focus
    if project_context.get("deadline_pressure"):
        adapted.attributes["efficiency_focus"] = True

    return adapted
```

**Use Case**: Adjusting communication style based on team size, project urgency, and complexity.

---

## 📖 Advanced Documentation

This Skill uses Progressive Disclosure. For detailed patterns:

- **[modules/persona-definitions.md](modules/persona-definitions.md)** - 4 personas with detailed attributes
- **[modules/detection-algorithms.md](modules/detection-algorithms.md)** - Expertise detection & selection logic
- **[modules/transitions-personalization.md](modules/transitions-personalization.md)** - Persona transitions & learning
- **[modules/context-adaptation.md](modules/context-adaptation.md)** - Project context adjustment
- **[modules/integration-patterns.md](modules/integration-patterns.md)** - Integration with Alfred workflow
- **[modules/reference.md](modules/reference.md)** - Metrics, best practices, troubleshooting

---

## 🎯 Persona Selection Workflow

**Step 1**: Detect user expertise level (beginner/intermediate/expert)
**Step 2**: Analyze request content for keywords
**Step 3**: Check explicit triggers (Alfred commands, team mode)
**Step 4**: Adapt to project context (complexity, team size, deadline)
**Step 5**: Transition smoothly if expertise level changes

---

## 🔗 Integration with Other Skills

**Complementary Skills**:
- Skill("moai-core-ask-user-questions") - Adaptive question formulation
- Skill("moai-cc-configuration") - User preference storage
- Skill("moai-core-session-state") - Persona persistence
- Skill("moai-core-workflow") - Workflow orchestration

---

## 📈 Version History

**1.1.0** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure
- ✨ 5 Core Patterns highlighted
- ✨ Modularized advanced content

**1.0.0** (2025-11-13)
- ✨ 4 core personas
- ✨ Expertise detection
- ✨ Persona transitions

---

**Maintained by**: alfred
**Domain**: Adaptive Communication
**Generated with**: MoAI-ADK Skill Factory
