"""Cost Tracking Models

Pydantic models for token usage tracking and cost calculation
across different AI providers and models.
"""

from datetime import datetime
from typing import Dict
from uuid import uuid4

from pydantic import BaseModel, Field


class TokenUsage(BaseModel):
    """Token usage statistics

    Attributes:
        input_tokens: Number of input tokens consumed
        output_tokens: Number of output tokens generated
    """

    input_tokens: int
    output_tokens: int

    @property
    def total_tokens(self) -> int:
        """Calculate total tokens used"""
        return self.input_tokens + self.output_tokens


class CostRecord(BaseModel):
    """Record of token usage and associated cost

    Attributes:
        id: Unique identifier for the record
        session_id: Session this usage belongs to
        model_id: The model used
        provider: The AI provider
        usage: Token usage statistics
        cost_usd: Calculated cost in USD
        created_at: Timestamp of the record
    """

    id: str = Field(default_factory=lambda: str(uuid4()))
    session_id: str
    model_id: str
    provider: str
    usage: TokenUsage
    cost_usd: float
    created_at: datetime = Field(default_factory=datetime.now)


class CostSummary(BaseModel):
    """Summary of costs across multiple records

    Attributes:
        total_cost: Total cost in USD
        by_provider: Cost breakdown by provider
        by_model: Cost breakdown by model
        token_counts: Aggregate token usage
    """

    total_cost: float
    by_provider: Dict[str, float]
    by_model: Dict[str, float]
    token_counts: TokenUsage
