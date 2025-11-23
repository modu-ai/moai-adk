# Persona Transitions & Learning

Smooth transitions between personas and user preference learning mechanisms for personalized communication.

## Gradual Persona Transitions

### Why Gradual Transitions Matter

Abrupt persona switches confuse users. Smooth transitions provide continuity while respecting changing preferences.

**Transition Scenarios**:
```
Beginner → Expert (as user learns)
TechnicalMentor → EfficiencyCoach (after understanding concepts)
Individual → Team (when context shifts)
Planned → Urgent (deadline pressure increases)
```

### Transition Implementation

```python
class PersonaTransition:
    """Manage smooth transitions between communication styles."""

    def gradual_transition(self, from_persona, to_persona, steps=3):
        """
        Create smooth transition from one persona to another.

        Args:
            from_persona: Starting persona
            to_persona: Target persona
            steps: Number of transition steps (default: 3)

        Returns:
            List of blended personas for gradual transition
        """
        transition_steps = []

        for i in range(1, steps + 1):
            # Calculate blend ratio (0.0 = from_persona, 1.0 = to_persona)
            blend_ratio = i / steps
            blended = self.blend_personas(
                from_persona, to_persona, blend_ratio
            )
            transition_steps.append(blended)

        return transition_steps

    def blend_personas(self, p1, p2, ratio):
        """
        Blend two personas at given ratio.

        ratio 0.0: 100% p1, 0% p2
        ratio 0.5: 50% p1, 50% p2
        ratio 1.0: 0% p1, 100% p2
        """
        blended = {}

        # Transition key attributes
        for attr in ["style", "explanation_depth", "pace", "brevity"]:
            if ratio <= 0.5:
                # Favor first persona
                blended[attr] = p1.attributes.get(attr)
            else:
                # Favor second persona
                blended[attr] = p2.attributes.get(attr)

        # Blend numeric attributes (0-1 scale)
        for attr in ["detail_level", "learning_effectiveness"]:
            p1_val = p1.attributes.get(attr, 0.5)
            p2_val = p2.attributes.get(attr, 0.5)
            blended[attr] = p1_val * (1 - ratio) + p2_val * ratio

        return blended
```

### Transition Example: Beginner to Expert

```python
# Session 1: User is clearly beginner
mentor = TechnicalMentor()
response1 = mentor.respond("How do I implement SPEC?")
# Output: Detailed 5-minute explanation with examples

# After 3-4 interactions, signals shift
signals = [
    "technical_precision",
    "efficiency_keywords",
    "edge_case_handling"
]
expertise = detect_expertise_level(signals)
# Result: "intermediate" (65% confidence)

# Start gradual transition
transition = PersonaTransition()
steps = transition.gradual_transition(TechnicalMentor(), EfficiencyCoach())

# Step 1: 70% Mentor, 30% Coach
response2 = steps[0].respond("How do I optimize caching?")
# Output: Explanation with efficiency focus

# Step 2: 50/50 blend
response3 = steps[1].respond("Implement caching pattern")
# Output: Brief explanation with focus

# Step 3: 30% Mentor, 70% Coach
response4 = steps[2].respond("Cache optimization")
# Output: Minimal explanation, solution-focused

# After confidence increases to 90%
coach = EfficiencyCoach()
response5 = coach.respond("Optimize cache")
# Output: One-line response, assume knowledge
```

---

## User Preference Learning

### Learning Engine

```python
class PersonalizationEngine:
    """Learn user preferences and adapt persona selection."""

    def __init__(self):
        self.user_preferences = {}    # user_id → preferences
        self.interaction_history = {} # user_id → interactions
        self.satisfaction_feedback = {} # user_id → satisfaction scores

    def learn_preferences(self, user_id: str, interaction_data: dict):
        """
        Learn user preferences from interaction data.

        interaction_data includes:
        - persona_used: Which persona was selected
        - satisfaction: User satisfaction score (0-1)
        - task_completion: Whether task was completed
        - time_taken: How long interaction took
        - user_feedback: Explicit user preferences
        """

        if user_id not in self.user_preferences:
            self.user_preferences[user_id] = {
                "preferred_persona": None,
                "preferred_style": None,
                "explanation_preference": None,
                "pace_preference": None,
                "learning_effective": False
            }

        # High satisfaction = user liked this persona
        if interaction_data.get("satisfaction", 0) > 0.8:
            persona = interaction_data.get("persona_used")
            self.user_preferences[user_id]["preferred_persona"] = persona

            # Extract preferences from interaction
            style = interaction_data.get("style")
            if style:
                self.user_preferences[user_id]["preferred_style"] = style

            depth = interaction_data.get("explanation_depth")
            if depth:
                self.user_preferences[user_id]["explanation_preference"] = depth

            pace = interaction_data.get("pace")
            if pace:
                self.user_preferences[user_id]["pace_preference"] = pace

        # Task completion indicates effective persona choice
        if interaction_data.get("task_completed"):
            self.user_preferences[user_id]["learning_effective"] = True

        # Store for history
        if user_id not in self.interaction_history:
            self.interaction_history[user_id] = []

        self.interaction_history[user_id].append(interaction_data)

    def get_personalized_persona(self, user_id: str, base_persona):
        """
        Get persona adjusted for user preferences.

        Returns personalized version of base persona if preferences exist,
        otherwise returns unmodified base persona.
        """

        if user_id not in self.user_preferences:
            return base_persona

        prefs = self.user_preferences[user_id]

        # If user has explicit preferred persona, use it
        if prefs.get("preferred_persona"):
            return self.get_persona_by_name(prefs["preferred_persona"])

        # If no explicit preference, apply preference adjustments
        if prefs.get("preferred_style"):
            return self.apply_style_preference(base_persona, prefs["preferred_style"])

        return base_persona

    def apply_style_preference(self, persona, style):
        """Adjust persona to match user's style preference."""

        adjusted = copy.deepcopy(persona)
        adjusted.attributes["style"] = style

        if style == "brief":
            adjusted.attributes["explanation_depth"] = "minimal"
            adjusted.attributes["brevity"] = True
        elif style == "detailed":
            adjusted.attributes["explanation_depth"] = "thorough"
            adjusted.attributes["brevity"] = False

        return adjusted
```

### Learning Examples

**Example 1: User consistently prefers TechnicalMentor**

```python
interactions = [
    {
        "persona_used": "TechnicalMentor",
        "satisfaction": 0.9,  # High satisfaction
        "task_completed": True,
        "feedback": "Clear examples were helpful"
    },
    {
        "persona_used": "EfficiencyCoach",
        "satisfaction": 0.6,  # Lower satisfaction
        "task_completed": False,
        "feedback": "Too brief, confused me"
    },
    {
        "persona_used": "TechnicalMentor",
        "satisfaction": 0.88,
        "task_completed": True
    }
]

for interaction in interactions:
    engine.learn_preferences("user123", interaction)

# Result
prefs = engine.user_preferences["user123"]
# preferred_persona: TechnicalMentor
# preferred_style: "educational"
```

**Example 2: Preference changes over time**

```python
# Early sessions: Prefers detailed explanations
early_interactions = [
    {"persona_used": "TechnicalMentor", "satisfaction": 0.85},
    {"persona_used": "TechnicalMentor", "satisfaction": 0.87}
]

# Later sessions: Prefers brief, efficient responses
later_interactions = [
    {"persona_used": "EfficiencyCoach", "satisfaction": 0.92},
    {"persona_used": "EfficiencyCoach", "satisfaction": 0.88},
    {"persona_used": "EfficiencyCoach", "satisfaction": 0.91}
]

# Preference shifts to EfficiencyCoach
# Learning system detects this and adapts
```

---

## Feedback Integration

### Explicit User Feedback

```python
class FeedbackCollector:
    """Collect explicit user feedback about persona preferences."""

    def request_persona_feedback(self, user_id, persona_used, session_id):
        """Request user feedback after interaction."""

        return {
            "question": f"How was the {persona_used.name} communication style?",
            "options": [
                "Too detailed (prefer briefer)",
                "Too brief (prefer more detail)",
                "Perfect - just right",
                "Wrong persona entirely"
            ]
        }

    def process_feedback(self, user_id, feedback, persona_used):
        """Process feedback and update preferences."""

        if feedback == "Too detailed (prefer briefer)":
            update_preference(user_id, "explanation_depth", "minimal")
            suggest_persona(EfficiencyCoach)

        elif feedback == "Too brief (prefer more detail)":
            update_preference(user_id, "explanation_depth", "thorough")
            suggest_persona(TechnicalMentor)

        elif feedback == "Perfect - just right":
            reinforce_persona(user_id, persona_used)

        elif feedback == "Wrong persona entirely":
            learn_mismatch(user_id, persona_used)
```

---

## Satisfaction Metrics

### Measuring Success

```python
def calculate_satisfaction_score(interaction) -> float:
    """
    Calculate satisfaction from multiple signals.
    Returns 0-1 score.
    """

    score = 0.0
    weights = {
        "task_completed": 0.4,        # 40% weight
        "no_clarification_needed": 0.3, # 30% weight
        "response_time": 0.15,        # 15% weight
        "user_feedback": 0.15         # 15% weight
    }

    # Task completion
    if interaction.get("task_completed"):
        score += 1.0 * weights["task_completed"]
    else:
        score += 0.3 * weights["task_completed"]

    # Needed clarification
    if not interaction.get("clarification_requests"):
        score += 1.0 * weights["no_clarification_needed"]
    else:
        score += 0.5 * weights["no_clarification_needed"]

    # Response time (target 45-180s)
    time_taken = interaction.get("time_taken", 90)
    if 45 <= time_taken <= 180:
        score += 1.0 * weights["response_time"]
    else:
        score += 0.7 * weights["response_time"]

    # Explicit feedback
    feedback = interaction.get("user_feedback", 0)
    score += feedback * weights["user_feedback"]

    return score
```

---

## Transition Strategies Summary

| Strategy | Use Case | Speed | Smoothness |
|----------|----------|-------|-----------|
| **Abrupt** | Explicit request, urgent need | Fast | Low |
| **3-Step Gradual** | Normal expertise growth | Moderate | High |
| **Adaptive** | Real-time feedback | Varies | Very High |
| **Preference-Based** | Learned user patterns | N/A | High |

---

## Best Practices

✅ **DO**:
- Start transitions when confidence > 0.70
- Collect satisfaction feedback regularly
- Allow 3-5 interactions before concluding preference
- Override learned preferences on explicit request
- Update preferences as user expertise grows
- Blend personas during uncertain transitions

❌ **DON'T**:
- Change persona abruptly
- Lock into learned preferences
- Ignore explicit user feedback
- Assume preferences are permanent
- Transition without enough signals
- Skip satisfaction measurement

---

**Updated**: 2025-11-23 | **Learning Enabled**: Yes | **Transition Steps**: Configurable (1-5)
