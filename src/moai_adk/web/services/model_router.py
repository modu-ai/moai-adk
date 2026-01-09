"""Model Router Service

Service for routing tasks to appropriate model tiers based on task classification.
Implements the Multi-Model Router pattern with tiered model selection.
"""

from typing import Dict

from moai_adk.web.models.model_config import ModelConfig, ModelTier, TaskClassification


class ModelRoutingError(Exception):
    """Error during model routing"""

    def __init__(self, message: str, tier: ModelTier, fallback_used: bool = False):
        self.message = message
        self.tier = tier
        self.fallback_used = fallback_used
        super().__init__(message)


class ModelRouter:
    """Routes tasks to appropriate model tier based on classification

    The router classifies tasks and routes them to either:
    - PLANNER tier (Opus): For architecture, design, complex reasoning
    - IMPLEMENTER tier (GLM): For implementation, coding, testing

    Attributes:
        MODELS: Configuration for each model tier
        PLANNING_KEYWORDS: Keywords that trigger PLANNER tier
        IMPLEMENTATION_KEYWORDS: Keywords that trigger IMPLEMENTER tier
    """

    # Model configurations for each tier
    MODELS: Dict[ModelTier, ModelConfig] = {
        ModelTier.PLANNER: ModelConfig(
            tier=ModelTier.PLANNER,
            model_id="claude-opus-4-5",
            provider="claude",
            cost_per_1k_input=0.015,
            cost_per_1k_output=0.075,
        ),
        ModelTier.IMPLEMENTER: ModelConfig(
            tier=ModelTier.IMPLEMENTER,
            model_id="glm-4.7",
            provider="glm",
            cost_per_1k_input=0.002,
            cost_per_1k_output=0.014,
            base_url="https://open.bigmodel.cn/api/paas/v4",
        ),
    }

    # Keywords that indicate planning/architecture tasks (use PLANNER)
    PLANNING_KEYWORDS = [
        "/moai:1-plan",
        "architecture",
        "design",
        "strategy",
        "complex",
        "decision",
        "trade-off",
        "evaluate",
    ]

    # Keywords that indicate implementation tasks (use IMPLEMENTER)
    IMPLEMENTATION_KEYWORDS = [
        "/moai:2-run",
        "implement",
        "code",
        "test",
        "tdd",
        "refactor",
        "fix",
        "update",
        "create",
    ]

    # Class-level state (shared across instances)
    _active_tier: ModelTier = ModelTier.IMPLEMENTER

    def __init__(self):
        """Initialize the model router"""
        pass

    def classify_task(self, task_description: str) -> TaskClassification:
        """Classify a task and recommend a model tier

        Args:
            task_description: Description of the task to classify

        Returns:
            TaskClassification with recommended tier and reasoning
        """
        task_lower = task_description.lower()

        # Check for planning keywords first (higher priority)
        for keyword in self.PLANNING_KEYWORDS:
            if keyword.lower() in task_lower:
                return TaskClassification(
                    task_type="planning",
                    recommended_tier=ModelTier.PLANNER,
                    reason=f"Task contains planning keyword: {keyword}",
                )

        # Check for implementation keywords
        for keyword in self.IMPLEMENTATION_KEYWORDS:
            if keyword.lower() in task_lower:
                return TaskClassification(
                    task_type="implementation",
                    recommended_tier=ModelTier.IMPLEMENTER,
                    reason=f"Task contains implementation keyword: {keyword}",
                )

        # Default to IMPLEMENTER (cost-effective choice)
        return TaskClassification(
            task_type="general",
            recommended_tier=ModelTier.IMPLEMENTER,
            reason="No specific keywords detected, defaulting to cost-effective tier",
        )

    def get_model_for_tier(self, tier: ModelTier) -> ModelConfig:
        """Get the model configuration for a specific tier

        Args:
            tier: The model tier to get configuration for

        Returns:
            ModelConfig for the specified tier
        """
        return self.MODELS[tier]

    def get_environment_for_model(self, config: ModelConfig) -> Dict[str, str]:
        """Get environment variables for a model configuration

        Args:
            config: The model configuration

        Returns:
            Dictionary of environment variables
        """
        env = {
            "MODEL": config.model_id,
            "PROVIDER": config.provider,
        }

        if config.provider == "claude":
            env["ANTHROPIC_MODEL"] = config.model_id
        elif config.provider == "glm":
            env["GLM_MODEL"] = config.model_id
            if config.base_url:
                env["GLM_BASE_URL"] = config.base_url

        return env

    def switch_to_tier(self, tier: ModelTier) -> None:
        """Switch to a specific model tier

        Args:
            tier: The tier to switch to
        """
        ModelRouter._active_tier = tier

    def get_active_tier(self) -> ModelTier:
        """Get the currently active tier

        Returns:
            The active ModelTier
        """
        return ModelRouter._active_tier

    def reset_to_default(self) -> None:
        """Reset to default tier (IMPLEMENTER)"""
        ModelRouter._active_tier = ModelTier.IMPLEMENTER

    def get_active_model(self) -> ModelConfig:
        """Get the configuration for the currently active tier

        Returns:
            ModelConfig for the active tier
        """
        return self.MODELS[self._active_tier]

    def get_fallback_tier(self, tier: ModelTier) -> ModelTier:
        """Get the fallback tier for a given tier

        PLANNER falls back to IMPLEMENTER (cost-effective)
        IMPLEMENTER has no fallback (already cost-effective)

        Args:
            tier: The original tier

        Returns:
            The fallback tier
        """
        if tier == ModelTier.PLANNER:
            return ModelTier.IMPLEMENTER
        # IMPLEMENTER is already the most cost-effective, no fallback
        return tier

    def route_with_fallback(
        self,
        task_description: str,
        primary_failed: bool = False,
    ) -> tuple[ModelConfig, bool]:
        """Route a task with fallback support

        Args:
            task_description: Description of the task
            primary_failed: Whether the primary model has failed

        Returns:
            Tuple of (ModelConfig, fallback_used)
        """
        classification = self.classify_task(task_description)
        target_tier = classification.recommended_tier

        if primary_failed:
            fallback_tier = self.get_fallback_tier(target_tier)
            if fallback_tier != target_tier:
                return self.MODELS[fallback_tier], True
            # No fallback available, return original
            return self.MODELS[target_tier], False

        return self.MODELS[target_tier], False

    def is_fallback_available(self, tier: ModelTier) -> bool:
        """Check if a fallback is available for a tier

        Args:
            tier: The tier to check

        Returns:
            True if fallback is available
        """
        return tier == ModelTier.PLANNER
