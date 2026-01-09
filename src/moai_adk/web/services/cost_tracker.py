"""Cost Tracker Service

Service for tracking token usage and calculating costs
across different AI providers and models.
"""

from collections import defaultdict
from datetime import date
from typing import Dict, List

from moai_adk.web.models.cost import CostRecord, CostSummary, TokenUsage
from moai_adk.web.models.model_config import ModelTier


class CostLimitWarning:
    """Cost limit warning data"""

    def __init__(
        self,
        current_cost: float,
        limit: float,
        percentage_used: float,
        is_exceeded: bool,
    ):
        self.current_cost = current_cost
        self.limit = limit
        self.percentage_used = percentage_used
        self.is_exceeded = is_exceeded

    def to_dict(self) -> Dict:
        return {
            "current_cost": self.current_cost,
            "limit": self.limit,
            "percentage_used": self.percentage_used,
            "is_exceeded": self.is_exceeded,
        }


class CostTracker:
    """Tracks token usage and costs per model

    Provides methods for recording usage, calculating costs,
    and generating cost summaries.

    Attributes:
        records: List of all cost records
        MODEL_PRICING: Pricing configuration per model
        daily_limit_usd: Daily spending limit in USD (default: $10.00)
        warning_threshold: Percentage at which to warn (default: 0.8 = 80%)
    """

    # Pricing per 1K tokens for each model
    MODEL_PRICING: Dict[str, Dict[str, float]] = {
        "claude-opus-4-5": {
            "input": 0.015,
            "output": 0.075,
        },
        "glm-4.7": {
            "input": 0.002,
            "output": 0.014,
        },
    }

    # Mapping from tier to model for cost estimation
    TIER_MODEL_MAP: Dict[ModelTier, str] = {
        ModelTier.PLANNER: "claude-opus-4-5",
        ModelTier.IMPLEMENTER: "glm-4.7",
    }

    # Default daily limit ($10)
    DEFAULT_DAILY_LIMIT_USD: float = 10.0

    # Warning threshold (80%)
    DEFAULT_WARNING_THRESHOLD: float = 0.8

    def __init__(
        self,
        daily_limit_usd: float = DEFAULT_DAILY_LIMIT_USD,
        warning_threshold: float = DEFAULT_WARNING_THRESHOLD,
    ):
        """Initialize the cost tracker

        Args:
            daily_limit_usd: Daily spending limit in USD
            warning_threshold: Percentage (0-1) at which to warn
        """
        self.records: List[CostRecord] = []
        self.daily_limit_usd = daily_limit_usd
        self.warning_threshold = warning_threshold

    def _calculate_cost(self, model_id: str, input_tokens: int, output_tokens: int) -> float:
        """Calculate cost for token usage

        Args:
            model_id: The model identifier
            input_tokens: Number of input tokens
            output_tokens: Number of output tokens

        Returns:
            Cost in USD
        """
        if model_id not in self.MODEL_PRICING:
            # Default to GLM pricing for unknown models
            pricing = self.MODEL_PRICING["glm-4.7"]
        else:
            pricing = self.MODEL_PRICING[model_id]

        input_cost = (input_tokens / 1000) * pricing["input"]
        output_cost = (output_tokens / 1000) * pricing["output"]

        return input_cost + output_cost

    def record_usage(
        self,
        session_id: str,
        model_id: str,
        provider: str,
        input_tokens: int,
        output_tokens: int,
    ) -> CostRecord:
        """Record token usage and calculate cost

        Args:
            session_id: Session identifier
            model_id: Model identifier
            provider: AI provider name
            input_tokens: Number of input tokens
            output_tokens: Number of output tokens

        Returns:
            CostRecord with calculated cost
        """
        cost = self._calculate_cost(model_id, input_tokens, output_tokens)

        record = CostRecord(
            session_id=session_id,
            model_id=model_id,
            provider=provider,
            usage=TokenUsage(input_tokens=input_tokens, output_tokens=output_tokens),
            cost_usd=cost,
        )

        self.records.append(record)
        return record

    def _aggregate_records(self, records: List[CostRecord]) -> CostSummary:
        """Aggregate a list of records into a summary

        Args:
            records: List of cost records to aggregate

        Returns:
            CostSummary with aggregated data
        """
        total_cost = 0.0
        by_provider: Dict[str, float] = defaultdict(float)
        by_model: Dict[str, float] = defaultdict(float)
        total_input = 0
        total_output = 0

        for record in records:
            total_cost += record.cost_usd
            by_provider[record.provider] += record.cost_usd
            by_model[record.model_id] += record.cost_usd
            total_input += record.usage.input_tokens
            total_output += record.usage.output_tokens

        return CostSummary(
            total_cost=total_cost,
            by_provider=dict(by_provider),
            by_model=dict(by_model),
            token_counts=TokenUsage(input_tokens=total_input, output_tokens=total_output),
        )

    def get_session_cost(self, session_id: str) -> CostSummary:
        """Get cost summary for a specific session

        Args:
            session_id: Session identifier

        Returns:
            CostSummary for the session
        """
        session_records = [r for r in self.records if r.session_id == session_id]
        return self._aggregate_records(session_records)

    def get_daily_cost(self, target_date: date) -> CostSummary:
        """Get cost summary for a specific date

        Args:
            target_date: Date to get costs for

        Returns:
            CostSummary for the date
        """
        daily_records = [r for r in self.records if r.created_at.date() == target_date]
        return self._aggregate_records(daily_records)

    def get_total_cost(self) -> CostSummary:
        """Get total cost summary across all records

        Returns:
            CostSummary for all records
        """
        return self._aggregate_records(self.records)

    def estimate_cost(self, tier: ModelTier, estimated_tokens: int) -> float:
        """Estimate cost for a task based on tier

        Assumes a 50/50 split between input and output tokens.

        Args:
            tier: The model tier to estimate for
            estimated_tokens: Total estimated tokens

        Returns:
            Estimated cost in USD
        """
        model_id = self.TIER_MODEL_MAP[tier]
        # Assume 50/50 split between input and output
        input_tokens = estimated_tokens // 2
        output_tokens = estimated_tokens - input_tokens

        return self._calculate_cost(model_id, input_tokens, output_tokens)

    def clear_records(self) -> None:
        """Clear all records"""
        self.records.clear()

    def check_daily_limit(self, target_date: date | None = None) -> CostLimitWarning:
        """Check if daily cost limit is approaching or exceeded

        Args:
            target_date: Date to check (defaults to today)

        Returns:
            CostLimitWarning with limit status
        """
        if target_date is None:
            target_date = date.today()

        daily_summary = self.get_daily_cost(target_date)
        current_cost = daily_summary.total_cost
        percentage_used = current_cost / self.daily_limit_usd if self.daily_limit_usd > 0 else 0

        return CostLimitWarning(
            current_cost=current_cost,
            limit=self.daily_limit_usd,
            percentage_used=percentage_used,
            is_exceeded=current_cost >= self.daily_limit_usd,
        )

    def is_limit_warning(self, target_date: date | None = None) -> bool:
        """Check if usage has reached warning threshold

        Args:
            target_date: Date to check (defaults to today)

        Returns:
            True if warning threshold reached
        """
        warning = self.check_daily_limit(target_date)
        return warning.percentage_used >= self.warning_threshold

    def is_limit_exceeded(self, target_date: date | None = None) -> bool:
        """Check if daily limit has been exceeded

        Args:
            target_date: Date to check (defaults to today)

        Returns:
            True if limit exceeded
        """
        warning = self.check_daily_limit(target_date)
        return warning.is_exceeded

    def set_daily_limit(self, limit_usd: float) -> None:
        """Set a new daily spending limit

        Args:
            limit_usd: New daily limit in USD
        """
        self.daily_limit_usd = limit_usd

    def set_warning_threshold(self, threshold: float) -> None:
        """Set a new warning threshold

        Args:
            threshold: New threshold as percentage (0-1)
        """
        self.warning_threshold = max(0.0, min(1.0, threshold))
