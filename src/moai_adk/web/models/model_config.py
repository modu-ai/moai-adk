"""Model Configuration Models

Pydantic models for multi-model router configuration
including model tiers, model configurations, and task classifications.
"""

from enum import Enum
from typing import Literal

from pydantic import BaseModel


class ModelTier(str, Enum):
    """Model tier enumeration for task routing

    Defines the available model tiers:
    - PLANNER: High-capability models for complex reasoning (10% of tokens)
    - IMPLEMENTER: Cost-effective models for bulk work (90% of tokens)
    """

    PLANNER = "planner"
    IMPLEMENTER = "implementer"


class ModelConfig(BaseModel):
    """Configuration for a specific AI model

    Attributes:
        tier: The model tier (PLANNER or IMPLEMENTER)
        model_id: Unique identifier for the model
        provider: The AI provider (claude or glm)
        cost_per_1k_input: Cost per 1000 input tokens in USD
        cost_per_1k_output: Cost per 1000 output tokens in USD
        base_url: Optional base URL for API endpoint
    """

    tier: ModelTier
    model_id: str
    provider: Literal["claude", "glm"]
    cost_per_1k_input: float
    cost_per_1k_output: float
    base_url: str | None = None


class TaskClassification(BaseModel):
    """Classification result for a task

    Attributes:
        task_type: The type of task identified
        recommended_tier: The recommended model tier
        reason: Explanation for the recommendation
    """

    task_type: str
    recommended_tier: ModelTier
    reason: str
