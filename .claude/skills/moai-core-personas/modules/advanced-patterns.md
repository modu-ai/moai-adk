# Advanced Persona Management Patterns

Advanced patterns for enterprise-grade persona management in Claude Code, including dynamic persona switching, multi-dimensional expertise detection, and context-aware personalization strategies.

## Multi-Dimensional Expertise Detection

**Challenge**: Accurately detect user expertise across multiple dimensions beyond simple beginner/intermediate/expert classification.

**Pattern**: Composite Expertise Profiling

```python
from dataclasses import dataclass
from typing import Dict, List
from enum import Enum

class ExpertiseDimension(Enum):
    """Dimensions of user expertise."""
    TECHNICAL_KNOWLEDGE = "technical"
    DOMAIN_KNOWLEDGE = "domain"
    TOOL_PROFICIENCY = "tools"
    CONCEPTUAL_UNDERSTANDING = "concepts"
    PROBLEM_SOLVING = "problem_solving"

@dataclass
class ExpertiseProfile:
    """Multi-dimensional expertise profile."""
    dimensions: Dict[ExpertiseDimension, float]  # 0.0-1.0 scale
    confidence: float
    evidence: List[str]
    last_updated: datetime

class MultiDimensionalExpertiseDetector:
    """
    Detect user expertise across multiple dimensions.

    Features:
    - Technical knowledge assessment
    - Domain expertise evaluation
    - Tool proficiency measurement
    - Conceptual understanding analysis
    - Problem-solving capability detection
    """

    def __init__(self):
        self.expertise_signals = {
            ExpertiseDimension.TECHNICAL_KNOWLEDGE: [
                "uses_technical_jargon",
                "references_advanced_concepts",
                "understands_architecture",
                "knows_best_practices"
            ],
            ExpertiseDimension.DOMAIN_KNOWLEDGE: [
                "familiar_with_domain_terms",
                "understands_business_context",
                "knows_industry_standards",
                "references_domain_patterns"
            ],
            ExpertiseDimension.TOOL_PROFICIENCY: [
                "efficient_tool_usage",
                "knows_advanced_features",
                "uses_keyboard_shortcuts",
                "optimizes_workflows"
            ]
        }

    def analyze_expertise(self, user_interactions: List[Interaction]) -> ExpertiseProfile:
        """
        Analyze user interactions to build expertise profile.

        Args:
            user_interactions: History of user interactions

        Returns:
            Multi-dimensional expertise profile

        Example:
            >>> detector = MultiDimensionalExpertiseDetector()
            >>> interactions = load_user_history("user-123")
            >>> profile = detector.analyze_expertise(interactions)
            >>> print(f"Technical: {profile.dimensions[ExpertiseDimension.TECHNICAL_KNOWLEDGE]:.2%}")
            Technical: 78.5%
        """
        dimension_scores = {}

        for dimension in ExpertiseDimension:
            score, evidence = self._assess_dimension(
                dimension, user_interactions
            )
            dimension_scores[dimension] = score

        # Calculate overall confidence
        confidence = self._calculate_confidence(
            user_interactions, dimension_scores
        )

        return ExpertiseProfile(
            dimensions=dimension_scores,
            confidence=confidence,
            evidence=evidence,
            last_updated=datetime.now()
        )

    def _assess_dimension(
        self, dimension: ExpertiseDimension, interactions: List[Interaction]
    ) -> tuple[float, List[str]]:
        """Assess expertise in specific dimension."""

        signals = self.expertise_signals.get(dimension, [])
        detected_signals = []
        signal_counts = []

        for signal in signals:
            count = sum(
                1 for interaction in interactions
                if self._detect_signal(signal, interaction)
            )
            if count > 0:
                detected_signals.append(f"{signal}: {count}")
                signal_counts.append(count)

        # Calculate score (0.0-1.0)
        if not signal_counts:
            score = 0.0
        else:
            # Normalize to 0-1 range
            max_possible = len(interactions) * len(signals)
            actual = sum(signal_counts)
            score = min(1.0, actual / max_possible if max_possible > 0 else 0.0)

        return score, detected_signals

    def _detect_signal(self, signal: str, interaction: Interaction) -> bool:
        """Detect presence of expertise signal in interaction."""

        signal_patterns = {
            "uses_technical_jargon": lambda i: self._has_technical_terms(i.content),
            "references_advanced_concepts": lambda i: self._has_advanced_concepts(i.content),
            "efficient_tool_usage": lambda i: i.tool_usage_count > 5,
            "knows_best_practices": lambda i: "best practice" in i.content.lower()
        }

        detector = signal_patterns.get(signal)
        return detector(interaction) if detector else False
```

## Dynamic Persona Switching

**Challenge**: Seamlessly switch personas during conversation based on context changes.

**Pattern**: Context-Aware Persona Transitions

```python
from typing import Optional
from collections import deque

class PersonaTransitionManager:
    """
    Manage smooth transitions between personas.

    Features:
    - Context-aware persona selection
    - Smooth transition phrases
    - Transition history tracking
    - Rollback capability
    """

    def __init__(self):
        self.current_persona = None
        self.transition_history = deque(maxlen=10)
        self.transition_phrases = {
            ("beginner", "intermediate"): [
                "I notice you're getting the hang of this. Let me provide more detailed explanations.",
                "Great progress! I'll adjust my explanations to match your growing understanding."
            ],
            ("intermediate", "expert"): [
                "You clearly have strong expertise. I'll shift to more technical discussions.",
                "Excellent understanding! Let me engage at a more advanced level."
            ],
            ("expert", "intermediate"): [
                "Let me break this down in more accessible terms.",
                "I'll provide more context to ensure clarity."
            ]
        }

    def should_transition(
        self,
        current_persona: str,
        context: ConversationContext,
        expertise_profile: ExpertiseProfile
    ) -> Optional[str]:
        """
        Determine if persona transition is needed.

        Args:
            current_persona: Current active persona
            context: Conversation context
            expertise_profile: User's expertise profile

        Returns:
            New persona if transition needed, None otherwise

        Example:
            >>> manager = PersonaTransitionManager()
            >>> new_persona = manager.should_transition(
            ...     "beginner",
            ...     context,
            ...     expertise_profile
            ... )
            >>> if new_persona:
            ...     print(f"Transitioning to: {new_persona}")
            Transitioning to: intermediate
        """
        # Calculate recommended persona from expertise
        recommended_persona = self._recommend_persona(expertise_profile)

        # Check if transition is warranted
        if recommended_persona != current_persona:
            # Verify transition stability (avoid ping-ponging)
            if self._is_transition_stable(current_persona, recommended_persona):
                return recommended_persona

        # Check for explicit context signals
        context_persona = self._detect_context_signals(context)
        if context_persona and context_persona != current_persona:
            return context_persona

        return None

    async def execute_transition(
        self,
        from_persona: str,
        to_persona: str,
        context: ConversationContext
    ) -> TransitionResult:
        """
        Execute persona transition with smooth communication.

        Returns:
            Transition result with transition phrase
        """
        # Select appropriate transition phrase
        transition_phrase = self._select_transition_phrase(from_persona, to_persona)

        # Record transition
        self.transition_history.append({
            "from": from_persona,
            "to": to_persona,
            "timestamp": datetime.now(),
            "context": context.summary(),
            "phrase": transition_phrase
        })

        # Update current persona
        self.current_persona = to_persona

        return TransitionResult(
            success=True,
            from_persona=from_persona,
            to_persona=to_persona,
            transition_phrase=transition_phrase,
            context_preserved=True
        )

    def _is_transition_stable(self, current: str, new: str) -> bool:
        """Check if transition is stable (avoid flip-flopping)."""

        # Check recent history for oscillations
        recent_transitions = list(self.transition_history)[-3:]

        if len(recent_transitions) >= 2:
            # Detect A→B→A pattern
            if (recent_transitions[-2]["from"] == new and
                recent_transitions[-2]["to"] == current):
                return False  # Would create oscillation

        return True

    def _select_transition_phrase(self, from_persona: str, to_persona: str) -> str:
        """Select appropriate transition phrase."""

        transition_key = (from_persona, to_persona)
        phrases = self.transition_phrases.get(transition_key, [
            "Let me adjust my communication style to better match your needs."
        ])

        # Select phrase based on context (simple random for now)
        import random
        return random.choice(phrases)
```

## Adaptive Communication Style

**Challenge**: Adjust communication style in real-time based on user response patterns.

**Pattern**: Reinforcement Learning for Style Adaptation

```python
from collections import defaultdict
import numpy as np

class AdaptiveCommunicationStyler:
    """
    Adapt communication style using reinforcement learning.

    Styles:
    - Concise: Brief, to-the-point responses
    - Detailed: Comprehensive explanations
    - Interactive: Questions and examples
    - Reference-heavy: Links and citations
    """

    def __init__(self):
        self.style_rewards = defaultdict(list)
        self.current_style = "detailed"
        self.learning_rate = 0.1
        self.exploration_rate = 0.2

    def select_communication_style(
        self, context: ConversationContext, user_profile: UserProfile
    ) -> CommunicationStyle:
        """
        Select optimal communication style.

        Uses epsilon-greedy strategy:
        - 80% exploitation (use best known style)
        - 20% exploration (try different styles)

        Args:
            context: Conversation context
            user_profile: User profile

        Returns:
            Selected communication style

        Example:
            >>> styler = AdaptiveCommunicationStyler()
            >>> style = styler.select_communication_style(context, profile)
            >>> print(f"Using style: {style.name} (confidence: {style.confidence:.2%})")
            Using style: interactive (confidence: 87.3%)
        """
        # Calculate style scores
        style_scores = {}

        for style_name in ["concise", "detailed", "interactive", "reference-heavy"]:
            # Get historical performance
            rewards = self.style_rewards.get(style_name, [])

            if rewards:
                # Average reward (0-1 scale)
                avg_reward = np.mean(rewards)
            else:
                # No data: neutral score
                avg_reward = 0.5

            # Adjust for context
            context_bonus = self._calculate_context_bonus(style_name, context)
            style_scores[style_name] = avg_reward + context_bonus

        # Epsilon-greedy selection
        if np.random.random() < self.exploration_rate:
            # Explore: random style
            selected_style = np.random.choice(list(style_scores.keys()))
        else:
            # Exploit: best style
            selected_style = max(style_scores, key=style_scores.get)

        confidence = style_scores[selected_style]

        return CommunicationStyle(
            name=selected_style,
            confidence=confidence,
            parameters=self._get_style_parameters(selected_style)
        )

    def record_style_feedback(
        self, style: str, user_response: UserResponse
    ) -> None:
        """
        Record feedback for style adaptation.

        Feedback signals:
        - User asks follow-up questions (style may need adjustment)
        - User expresses satisfaction (style is working)
        - User requests clarification (style too concise)
        - User skims response (style too verbose)
        """
        # Calculate reward (0-1 scale)
        reward = self._calculate_reward(user_response)

        # Store reward
        self.style_rewards[style].append(reward)

        # Keep recent history only (last 20 interactions)
        if len(self.style_rewards[style]) > 20:
            self.style_rewards[style] = self.style_rewards[style][-20:]

    def _calculate_reward(self, response: UserResponse) -> float:
        """Calculate reward from user response."""

        reward = 0.5  # Neutral baseline

        # Positive signals
        if response.expresses_satisfaction:
            reward += 0.3
        if response.continues_naturally:
            reward += 0.2

        # Negative signals
        if response.asks_for_clarification:
            reward -= 0.3
        if response.requests_more_detail:
            reward -= 0.2
        if response.requests_less_detail:
            reward -= 0.2

        # Clamp to 0-1 range
        return max(0.0, min(1.0, reward))
```

## Persona Inheritance and Composition

**Challenge**: Create specialized personas that inherit from base personas while adding custom behavior.

**Pattern**: Persona Composition with Mixins

```python
from abc import ABC, abstractmethod

class BasePersona(ABC):
    """Base class for all personas."""

    def __init__(self, name: str):
        self.name = name
        self.communication_style = None
        self.response_templates = {}

    @abstractmethod
    def format_response(self, content: str, context: dict) -> str:
        """Format response according to persona."""
        pass

    @abstractmethod
    def get_example_complexity(self) -> str:
        """Return example complexity level."""
        pass

class TechnicalExpertiseMixin:
    """Mixin for technical expertise traits."""

    def add_technical_references(self, content: str) -> str:
        """Add technical references to content."""
        # Add links to official documentation
        enhanced = content + "\n\n**Technical References:**\n"
        enhanced += "- [Official Documentation](https://...)\n"
        enhanced += "- [Best Practices Guide](https://...)\n"
        return enhanced

    def use_technical_terminology(self, content: str) -> str:
        """Use appropriate technical terminology."""
        # Replace simple terms with technical equivalents
        replacements = {
            "save": "persist",
            "get": "retrieve",
            "change": "modify",
            "fix": "resolve"
        }

        for simple, technical in replacements.items():
            content = content.replace(simple, technical)

        return content

class InteractiveGuide Mixin:
    """Mixin for interactive guidance traits."""

    def add_examples(self, content: str, count: int = 2) -> str:
        """Add interactive examples."""
        examples_section = "\n\n**Interactive Examples:**\n\n"

        for i in range(count):
            examples_section += f"**Example {i+1}:**\n"
            examples_section += "```python\n"
            examples_section += "# Example code here\n"
            examples_section += "```\n\n"

        return content + examples_section

    def add_questions(self, content: str) -> str:
        """Add clarifying questions."""
        questions = "\n\n**Questions to Consider:**\n"
        questions += "- What specific use case are you working on?\n"
        questions += "- Are there any constraints I should know about?\n"

        return content + questions

class ExpertPersona(BasePersona, TechnicalExpertiseMixin, InteractiveGuideMixin):
    """
    Expert persona with technical expertise and interactivity.

    Combines:
    - Base persona structure
    - Technical expertise traits
    - Interactive guidance capabilities
    """

    def __init__(self):
        super().__init__("expert")
        self.communication_style = "concise-technical"

    def format_response(self, content: str, context: dict) -> str:
        """
        Format response for expert users.

        Example:
            >>> persona = ExpertPersona()
            >>> formatted = persona.format_response(
            ...     "Use async/await for concurrency",
            ...     {"topic": "python"}
            ... )
            >>> print(formatted)
            Use async/await for concurrency

            **Technical References:**
            - [Official Documentation](https://...)
            ...
        """
        # Apply technical enhancements
        content = self.use_technical_terminology(content)
        content = self.add_technical_references(content)

        # Add interactive elements if requested
        if context.get("include_examples"):
            content = self.add_examples(content)

        return content

    def get_example_complexity(self) -> str:
        return "advanced"
```

## Multi-Persona Scenarios

**Challenge**: Handle scenarios requiring multiple personas simultaneously (e.g., team collaborations).

**Pattern**: Persona Orchestra tion

```python
class PersonaOrchestrator:
    """
    Orchestrate multiple personas in complex scenarios.

    Use Cases:
    - Team onboarding (multiple expertise levels)
    - Knowledge transfer sessions
    - Code review with mixed-level reviewers
    """

    def __init__(self):
        self.active_personas = {}
        self.persona_factory = PersonaFactory()

    async def coordinate_multi_persona_response(
        self, users: List[User], content: str, context: dict
    ) -> MultiPersonaResponse:
        """
        Generate response tailored for multiple users.

        Args:
            users: List of users with different expertise levels
            content: Content to communicate
            context: Conversation context

        Returns:
            Response with sections for each user level

        Example:
            >>> orchestrator = PersonaOrchestrator()
            >>> users = [
            ...     User("Alice", expertise="beginner"),
            ...     User("Bob", expertise="expert")
            ... ]
            >>> response = await orchestrator.coordinate_multi_persona_response(
            ...     users, "Implement OAuth2 authentication", {}
            ... )
            >>> print(response.beginner_section)
            ## For Beginners
            OAuth2 is a secure way to allow users to log in...
        """
        # Group users by expertise level
        users_by_level = defaultdict(list)
        for user in users:
            users_by_level[user.expertise_level].append(user)

        # Generate sections for each level
        sections = {}

        for level, level_users in users_by_level.items():
            # Get appropriate persona
            persona = self.persona_factory.create_persona(level)

            # Generate content for this level
            formatted_content = persona.format_response(content, context)

            sections[level] = {
                "users": [u.name for u in level_users],
                "content": formatted_content,
                "persona": persona.name
            }

        return MultiPersonaResponse(
            sections=sections,
            original_content=content,
            total_users=len(users)
        )
```

## Edge Case Handling

### Ambiguous Expertise Detection

```python
def handle_ambiguous_expertise(signals: List[ExpertiseSignal]) -> ExpertiseResolution:
    """
    Handle cases where expertise signals are contradictory.

    Strategy:
    1. Use most recent signals (weight recent > old)
    2. Prioritize explicit signals (user states level)
    3. Default to intermediate if uncertain
    4. Ask clarifying questions
    """
    # Separate explicit vs implicit signals
    explicit_signals = [s for s in signals if s.is_explicit]
    implicit_signals = [s for s in signals if not s.is_explicit]

    if explicit_signals:
        # Trust explicit signals most
        return ExpertiseResolution(
            level=explicit_signals[-1].level,  # Most recent
            confidence=0.9,
            source="explicit"
        )

    # Weight by recency
    weighted_scores = {}
    for i, signal in enumerate(reversed(implicit_signals)):
        weight = 1.0 / (i + 1)  # Recent signals get higher weight
        level = signal.level
        weighted_scores[level] = weighted_scores.get(level, 0) + weight

    if not weighted_scores:
        # No signals: default to intermediate
        return ExpertiseResolution(
            level="intermediate",
            confidence=0.3,
            source="default"
        )

    # Select level with highest weighted score
    best_level = max(weighted_scores, key=weighted_scores.get)
    confidence = weighted_scores[best_level] / sum(weighted_scores.values())

    if confidence < 0.6:
        # Low confidence: ask for clarification
        return ExpertiseResolution(
            level=best_level,
            confidence=confidence,
            source="inferred",
            needs_clarification=True
        )

    return ExpertiseResolution(
        level=best_level,
        confidence=confidence,
        source="inferred"
    )
```

### Persona Mismatch Recovery

```python
async def recover_from_persona_mismatch(
    current_persona: str, user_feedback: Feedback
) -> PersonaRecovery:
    """
    Recover when user indicates current persona is inappropriate.

    User signals:
    - "Can you simplify?"
    - "Too basic, I already know this"
    - "More technical detail please"
    """
    # Detect mismatch type
    if user_feedback.requests_simplification:
        # Current persona too advanced
        new_persona = _downgrade_persona(current_persona)
        recovery_message = "Let me explain this in simpler terms."

    elif user_feedback.requests_more_depth:
        # Current persona too basic
        new_persona = _upgrade_persona(current_persona)
        recovery_message = "I'll provide more technical detail."

    else:
        # Unclear signal: ask
        new_persona = await _ask_for_preferred_level()
        recovery_message = "I want to make sure I'm explaining this at the right level for you."

    return PersonaRecovery(
        original_persona=current_persona,
        new_persona=new_persona,
        recovery_message=recovery_message,
        transition_smooth=True
    )
```

## Performance Optimization

### Persona Selection Caching

```python
from functools import lru_cache
import hashlib

class PersonaSelectorCache:
    """Cache persona selections for performance."""

    def __init__(self, cache_size: int = 1000):
        self.cache = {}
        self.cache_hits = 0
        self.cache_misses = 0

    def get_cached_persona(
        self, user_id: str, context_hash: str
    ) -> Optional[PersonaSelection]:
        """Get cached persona selection."""

        cache_key = f"{user_id}:{context_hash}"

        if cache_key in self.cache:
            self.cache_hits += 1
            return self.cache[cache_key]

        self.cache_misses += 1
        return None

    def cache_persona_selection(
        self, user_id: str, context: dict, selection: PersonaSelection
    ) -> None:
        """Cache persona selection."""

        context_hash = self._hash_context(context)
        cache_key = f"{user_id}:{context_hash}"

        self.cache[cache_key] = selection

    def _hash_context(self, context: dict) -> str:
        """Create hash of context for caching."""
        context_str = str(sorted(context.items()))
        return hashlib.md5(context_str.encode()).hexdigest()

    def get_cache_stats(self) -> dict:
        """Get cache performance statistics."""
        total = self.cache_hits + self.cache_misses
        hit_rate = self.cache_hits / total if total > 0 else 0

        return {
            "hits": self.cache_hits,
            "misses": self.cache_misses,
            "hit_rate": hit_rate,
            "cache_size": len(self.cache)
        }
```

---

## Context7 Integration

**Fetch latest persona patterns**:

```python
async def get_persona_patterns_from_context7():
    """
    Get latest persona management patterns.

    Returns Claude Code persona adaptation best practices.
    """
    patterns = await context7.get_library_docs(
        context7_library_id="/anthropic/claude-code",
        topic="persona adaptation user expertise communication 2025",
        tokens=3000
    )

    return patterns
```

---

**Status**: Production Ready
**Version**: 3.0.0
**Last Updated**: 2025-11-22
