"""TAG-002, TAG-004, TAG-006: Test Cost Tracking

TDD tests for Cost Tracking models and service.
RED Phase: Tests written first to define expected behavior.
"""

from datetime import date, datetime

import pytest


class TestTokenUsageModel:
    """Test cases for TokenUsage model"""

    @pytest.mark.asyncio
    async def test_token_usage_class_exists(self):
        """Test that TokenUsage class exists"""
        from moai_adk.web.models.cost import TokenUsage

        assert TokenUsage is not None

    @pytest.mark.asyncio
    async def test_token_usage_has_required_fields(self):
        """Test that TokenUsage has required fields"""
        from moai_adk.web.models.cost import TokenUsage

        usage = TokenUsage(input_tokens=1000, output_tokens=500)

        assert usage.input_tokens == 1000
        assert usage.output_tokens == 500

    @pytest.mark.asyncio
    async def test_token_usage_total_tokens(self):
        """Test that TokenUsage calculates total tokens"""
        from moai_adk.web.models.cost import TokenUsage

        usage = TokenUsage(input_tokens=1000, output_tokens=500)

        assert usage.total_tokens == 1500


class TestCostRecordModel:
    """Test cases for CostRecord model"""

    @pytest.mark.asyncio
    async def test_cost_record_class_exists(self):
        """Test that CostRecord class exists"""
        from moai_adk.web.models.cost import CostRecord

        assert CostRecord is not None

    @pytest.mark.asyncio
    async def test_cost_record_has_required_fields(self):
        """Test that CostRecord has required fields"""
        from moai_adk.web.models.cost import CostRecord, TokenUsage

        record = CostRecord(
            id="test-123",
            session_id="session-456",
            model_id="claude-opus-4-5",
            provider="claude",
            usage=TokenUsage(input_tokens=1000, output_tokens=500),
            cost_usd=0.0525,
            created_at=datetime.now(),
        )

        assert record.id == "test-123"
        assert record.session_id == "session-456"
        assert record.model_id == "claude-opus-4-5"
        assert record.provider == "claude"
        assert record.usage.input_tokens == 1000
        assert record.usage.output_tokens == 500
        assert record.cost_usd == 0.0525

    @pytest.mark.asyncio
    async def test_cost_record_auto_generates_id(self):
        """Test that CostRecord auto-generates ID if not provided"""
        from moai_adk.web.models.cost import CostRecord, TokenUsage

        record = CostRecord(
            session_id="session-456",
            model_id="claude-opus-4-5",
            provider="claude",
            usage=TokenUsage(input_tokens=1000, output_tokens=500),
            cost_usd=0.0525,
        )

        assert record.id is not None
        assert len(record.id) > 0

    @pytest.mark.asyncio
    async def test_cost_record_auto_sets_created_at(self):
        """Test that CostRecord auto-sets created_at if not provided"""
        from moai_adk.web.models.cost import CostRecord, TokenUsage

        record = CostRecord(
            session_id="session-456",
            model_id="claude-opus-4-5",
            provider="claude",
            usage=TokenUsage(input_tokens=1000, output_tokens=500),
            cost_usd=0.0525,
        )

        assert record.created_at is not None


class TestCostSummaryModel:
    """Test cases for CostSummary model"""

    @pytest.mark.asyncio
    async def test_cost_summary_class_exists(self):
        """Test that CostSummary class exists"""
        from moai_adk.web.models.cost import CostSummary

        assert CostSummary is not None

    @pytest.mark.asyncio
    async def test_cost_summary_has_required_fields(self):
        """Test that CostSummary has required fields"""
        from moai_adk.web.models.cost import CostSummary, TokenUsage

        summary = CostSummary(
            total_cost=0.10,
            by_provider={"claude": 0.08, "glm": 0.02},
            by_model={"claude-opus-4-5": 0.08, "glm-4.7": 0.02},
            token_counts=TokenUsage(input_tokens=5000, output_tokens=2500),
        )

        assert summary.total_cost == 0.10
        assert summary.by_provider["claude"] == 0.08
        assert summary.by_model["claude-opus-4-5"] == 0.08
        assert summary.token_counts.total_tokens == 7500


class TestCostTrackerService:
    """Test cases for CostTracker service"""

    @pytest.mark.asyncio
    async def test_cost_tracker_class_exists(self):
        """Test that CostTracker class exists"""
        from moai_adk.web.services.cost_tracker import CostTracker

        assert CostTracker is not None

    @pytest.mark.asyncio
    async def test_cost_tracker_can_be_instantiated(self):
        """Test that CostTracker can be instantiated"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        assert tracker is not None

    @pytest.mark.asyncio
    async def test_cost_tracker_has_records_list(self):
        """Test that CostTracker has records list"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        assert hasattr(tracker, "records")
        assert isinstance(tracker.records, list)

    @pytest.mark.asyncio
    async def test_record_usage_returns_cost_record(self):
        """Test that record_usage returns CostRecord"""
        from moai_adk.web.models.cost import CostRecord
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        record = tracker.record_usage(
            session_id="session-123",
            model_id="claude-opus-4-5",
            provider="claude",
            input_tokens=1000,
            output_tokens=500,
        )

        assert isinstance(record, CostRecord)

    @pytest.mark.asyncio
    async def test_record_usage_calculates_cost(self):
        """Test that record_usage calculates cost correctly"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        record = tracker.record_usage(
            session_id="session-123",
            model_id="claude-opus-4-5",
            provider="claude",
            input_tokens=1000,
            output_tokens=1000,
        )

        # Claude Opus pricing: $0.015/1K input, $0.075/1K output
        # Expected: (1000/1000 * 0.015) + (1000/1000 * 0.075) = 0.015 + 0.075 = 0.09
        assert record.cost_usd == pytest.approx(0.09, rel=0.01)

    @pytest.mark.asyncio
    async def test_record_usage_glm_calculates_cost(self):
        """Test that record_usage calculates GLM cost correctly"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        record = tracker.record_usage(
            session_id="session-123",
            model_id="glm-4.7",
            provider="glm",
            input_tokens=1000,
            output_tokens=1000,
        )

        # GLM pricing: $0.002/1K input, $0.014/1K output
        # Expected: (1000/1000 * 0.002) + (1000/1000 * 0.014) = 0.002 + 0.014 = 0.016
        assert record.cost_usd == pytest.approx(0.016, rel=0.01)

    @pytest.mark.asyncio
    async def test_record_usage_stores_in_records(self):
        """Test that record_usage stores record in records list"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        initial_count = len(tracker.records)

        tracker.record_usage(
            session_id="session-123",
            model_id="claude-opus-4-5",
            provider="claude",
            input_tokens=1000,
            output_tokens=500,
        )

        assert len(tracker.records) == initial_count + 1

    @pytest.mark.asyncio
    async def test_get_session_cost_returns_cost_summary(self):
        """Test that get_session_cost returns CostSummary"""
        from moai_adk.web.models.cost import CostSummary
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        tracker.record_usage(
            session_id="session-123",
            model_id="claude-opus-4-5",
            provider="claude",
            input_tokens=1000,
            output_tokens=500,
        )

        summary = tracker.get_session_cost("session-123")

        assert isinstance(summary, CostSummary)

    @pytest.mark.asyncio
    async def test_get_session_cost_filters_by_session(self):
        """Test that get_session_cost filters by session ID"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        tracker.record_usage(
            session_id="session-123",
            model_id="claude-opus-4-5",
            provider="claude",
            input_tokens=1000,
            output_tokens=500,
        )
        tracker.record_usage(
            session_id="session-456",
            model_id="claude-opus-4-5",
            provider="claude",
            input_tokens=2000,
            output_tokens=1000,
        )

        summary = tracker.get_session_cost("session-123")

        assert summary.token_counts.input_tokens == 1000
        assert summary.token_counts.output_tokens == 500

    @pytest.mark.asyncio
    async def test_get_daily_cost_returns_cost_summary(self):
        """Test that get_daily_cost returns CostSummary"""
        from moai_adk.web.models.cost import CostSummary
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        tracker.record_usage(
            session_id="session-123",
            model_id="claude-opus-4-5",
            provider="claude",
            input_tokens=1000,
            output_tokens=500,
        )

        summary = tracker.get_daily_cost(date.today())

        assert isinstance(summary, CostSummary)

    @pytest.mark.asyncio
    async def test_get_total_cost_returns_cost_summary(self):
        """Test that get_total_cost returns CostSummary"""
        from moai_adk.web.models.cost import CostSummary
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        summary = tracker.get_total_cost()

        assert isinstance(summary, CostSummary)

    @pytest.mark.asyncio
    async def test_get_total_cost_aggregates_all_records(self):
        """Test that get_total_cost aggregates all records"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        # Clear any existing records
        tracker.records.clear()

        tracker.record_usage(
            session_id="session-1",
            model_id="claude-opus-4-5",
            provider="claude",
            input_tokens=1000,
            output_tokens=500,
        )
        tracker.record_usage(
            session_id="session-2",
            model_id="glm-4.7",
            provider="glm",
            input_tokens=2000,
            output_tokens=1000,
        )

        summary = tracker.get_total_cost()

        assert summary.token_counts.input_tokens == 3000
        assert summary.token_counts.output_tokens == 1500

    @pytest.mark.asyncio
    async def test_estimate_cost_returns_float(self):
        """Test that estimate_cost returns float"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        estimate = tracker.estimate_cost(ModelTier.PLANNER, estimated_tokens=1000)

        assert isinstance(estimate, float)

    @pytest.mark.asyncio
    async def test_estimate_cost_planner_tier(self):
        """Test that estimate_cost for PLANNER tier uses Opus pricing"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        # Estimate for 1000 tokens (assuming 50/50 input/output split)
        estimate = tracker.estimate_cost(ModelTier.PLANNER, estimated_tokens=1000)

        # Opus pricing: ~$0.015 input + ~$0.075 output per 1K
        # For 500 input + 500 output: (0.5 * 0.015) + (0.5 * 0.075) = 0.0075 + 0.0375 = 0.045
        assert estimate > 0

    @pytest.mark.asyncio
    async def test_estimate_cost_implementer_tier(self):
        """Test that estimate_cost for IMPLEMENTER tier uses GLM pricing"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        estimate = tracker.estimate_cost(ModelTier.IMPLEMENTER, estimated_tokens=1000)

        # GLM pricing: ~$0.002 input + ~$0.014 output per 1K
        assert estimate > 0

    @pytest.mark.asyncio
    async def test_estimate_cost_implementer_cheaper_than_planner(self):
        """Test that IMPLEMENTER tier is cheaper than PLANNER tier"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        planner_cost = tracker.estimate_cost(ModelTier.PLANNER, estimated_tokens=1000)
        implementer_cost = tracker.estimate_cost(ModelTier.IMPLEMENTER, estimated_tokens=1000)

        assert implementer_cost < planner_cost


class TestCostTrackerPricing:
    """Test cases for pricing calculations"""

    @pytest.mark.asyncio
    async def test_pricing_config_exists(self):
        """Test that MODEL_PRICING config exists"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        assert hasattr(tracker, "MODEL_PRICING")

    @pytest.mark.asyncio
    async def test_pricing_has_claude_opus(self):
        """Test that pricing has Claude Opus"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        assert "claude-opus-4-5" in tracker.MODEL_PRICING

    @pytest.mark.asyncio
    async def test_pricing_has_glm(self):
        """Test that pricing has GLM"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        assert "glm-4.7" in tracker.MODEL_PRICING

    @pytest.mark.asyncio
    async def test_pricing_claude_opus_values(self):
        """Test Claude Opus pricing values"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        pricing = tracker.MODEL_PRICING["claude-opus-4-5"]

        assert pricing["input"] == 0.015
        assert pricing["output"] == 0.075

    @pytest.mark.asyncio
    async def test_pricing_glm_values(self):
        """Test GLM pricing values"""
        from moai_adk.web.services.cost_tracker import CostTracker

        tracker = CostTracker()
        pricing = tracker.MODEL_PRICING["glm-4.7"]

        assert pricing["input"] == 0.002
        assert pricing["output"] == 0.014


# TAG-006: Cost Tracking REST API Tests
import pytest_asyncio
from httpx import ASGITransport, AsyncClient


class TestCostTrackingAPI:
    """Test cases for Cost Tracking REST API"""

    @pytest_asyncio.fixture
    async def client(self, tmp_path):
        """Create a test client with a temporary database"""
        from moai_adk.web.routers.cost import get_cost_tracker

        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database
        from moai_adk.web.server import create_app

        # Clear existing records
        get_cost_tracker().clear_records()

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        app = create_app(config)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
            yield client

        await close_database()

    @pytest.mark.asyncio
    async def test_record_usage_endpoint_returns_201(self, client):
        """Test that POST /api/cost/record returns 201"""
        response = await client.post(
            "/api/cost/record",
            json={
                "session_id": "session-123",
                "model_id": "claude-opus-4-5",
                "provider": "claude",
                "input_tokens": 1000,
                "output_tokens": 500,
            },
        )
        assert response.status_code == 201

    @pytest.mark.asyncio
    async def test_record_usage_returns_cost_record(self, client):
        """Test that record usage returns cost record data"""
        response = await client.post(
            "/api/cost/record",
            json={
                "session_id": "session-123",
                "model_id": "claude-opus-4-5",
                "provider": "claude",
                "input_tokens": 1000,
                "output_tokens": 500,
            },
        )
        data = response.json()

        assert "id" in data
        assert "session_id" in data
        assert "cost_usd" in data
        assert data["session_id"] == "session-123"

    @pytest.mark.asyncio
    async def test_get_session_cost_returns_200(self, client):
        """Test that GET /api/cost/session/{session_id} returns 200"""
        # First record some usage
        await client.post(
            "/api/cost/record",
            json={
                "session_id": "session-123",
                "model_id": "claude-opus-4-5",
                "provider": "claude",
                "input_tokens": 1000,
                "output_tokens": 500,
            },
        )

        response = await client.get("/api/cost/session/session-123")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_get_session_cost_returns_summary(self, client):
        """Test that session cost returns cost summary"""
        await client.post(
            "/api/cost/record",
            json={
                "session_id": "session-123",
                "model_id": "claude-opus-4-5",
                "provider": "claude",
                "input_tokens": 1000,
                "output_tokens": 500,
            },
        )

        response = await client.get("/api/cost/session/session-123")
        data = response.json()

        assert "total_cost" in data
        assert "by_provider" in data
        assert "by_model" in data
        assert "token_counts" in data

    @pytest.mark.asyncio
    async def test_get_daily_cost_returns_200(self, client):
        """Test that GET /api/cost/daily returns 200"""
        response = await client.get("/api/cost/daily")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_get_daily_cost_with_date_param(self, client):
        """Test that daily cost accepts date parameter"""
        response = await client.get("/api/cost/daily", params={"date": "2024-01-15"})
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_get_total_cost_returns_200(self, client):
        """Test that GET /api/cost/total returns 200"""
        response = await client.get("/api/cost/total")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_get_total_cost_returns_summary(self, client):
        """Test that total cost returns cost summary"""
        response = await client.get("/api/cost/total")
        data = response.json()

        assert "total_cost" in data
        assert "by_provider" in data
        assert "by_model" in data

    @pytest.mark.asyncio
    async def test_estimate_cost_returns_200(self, client):
        """Test that GET /api/cost/estimate returns 200"""
        response = await client.get(
            "/api/cost/estimate",
            params={"tier": "planner", "estimated_tokens": 1000},
        )
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_estimate_cost_returns_estimate(self, client):
        """Test that estimate cost returns estimate data"""
        response = await client.get(
            "/api/cost/estimate",
            params={"tier": "planner", "estimated_tokens": 1000},
        )
        data = response.json()

        assert "tier" in data
        assert "estimated_tokens" in data
        assert "estimated_cost" in data
        assert data["estimated_cost"] > 0

    @pytest.mark.asyncio
    async def test_estimate_cost_invalid_tier_returns_400(self, client):
        """Test that estimating for invalid tier returns 400"""
        response = await client.get(
            "/api/cost/estimate",
            params={"tier": "invalid_tier", "estimated_tokens": 1000},
        )
        assert response.status_code == 400
