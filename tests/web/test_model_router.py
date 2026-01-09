"""TAG-001 to TAG-007: Test Multi-Model Router

TDD tests for Multi-Model Router with tier-based task routing.
RED Phase: Tests written first to define expected behavior.
"""

import pytest


class TestModelTierEnum:
    """Test cases for ModelTier enumeration"""

    @pytest.mark.asyncio
    async def test_model_tier_enum_exists(self):
        """Test that ModelTier enum exists"""
        from moai_adk.web.models.model_config import ModelTier

        assert ModelTier is not None

    @pytest.mark.asyncio
    async def test_model_tier_has_planner_value(self):
        """Test that ModelTier has PLANNER value"""
        from moai_adk.web.models.model_config import ModelTier

        assert hasattr(ModelTier, "PLANNER")
        assert ModelTier.PLANNER.value == "planner"

    @pytest.mark.asyncio
    async def test_model_tier_has_implementer_value(self):
        """Test that ModelTier has IMPLEMENTER value"""
        from moai_adk.web.models.model_config import ModelTier

        assert hasattr(ModelTier, "IMPLEMENTER")
        assert ModelTier.IMPLEMENTER.value == "implementer"

    @pytest.mark.asyncio
    async def test_model_tier_is_string_enum(self):
        """Test that ModelTier is a string enum"""
        from moai_adk.web.models.model_config import ModelTier

        # Should be comparable to strings
        assert ModelTier.PLANNER == "planner"
        assert ModelTier.IMPLEMENTER == "implementer"


class TestModelConfig:
    """Test cases for ModelConfig model"""

    @pytest.mark.asyncio
    async def test_model_config_class_exists(self):
        """Test that ModelConfig class exists"""
        from moai_adk.web.models.model_config import ModelConfig

        assert ModelConfig is not None

    @pytest.mark.asyncio
    async def test_model_config_has_required_fields(self):
        """Test that ModelConfig has required fields"""
        from moai_adk.web.models.model_config import ModelConfig, ModelTier

        config = ModelConfig(
            tier=ModelTier.PLANNER,
            model_id="claude-opus-4-5",
            provider="claude",
            cost_per_1k_input=0.015,
            cost_per_1k_output=0.075,
        )

        assert config.tier == ModelTier.PLANNER
        assert config.model_id == "claude-opus-4-5"
        assert config.provider == "claude"
        assert config.cost_per_1k_input == 0.015
        assert config.cost_per_1k_output == 0.075

    @pytest.mark.asyncio
    async def test_model_config_base_url_optional(self):
        """Test that ModelConfig base_url is optional"""
        from moai_adk.web.models.model_config import ModelConfig, ModelTier

        config = ModelConfig(
            tier=ModelTier.IMPLEMENTER,
            model_id="glm-4.7",
            provider="glm",
            cost_per_1k_input=0.002,
            cost_per_1k_output=0.014,
        )

        assert config.base_url is None

    @pytest.mark.asyncio
    async def test_model_config_with_base_url(self):
        """Test that ModelConfig can have base_url"""
        from moai_adk.web.models.model_config import ModelConfig, ModelTier

        config = ModelConfig(
            tier=ModelTier.IMPLEMENTER,
            model_id="glm-4.7",
            provider="glm",
            cost_per_1k_input=0.002,
            cost_per_1k_output=0.014,
            base_url="https://open.bigmodel.cn/api/paas/v4",
        )

        assert config.base_url == "https://open.bigmodel.cn/api/paas/v4"

    @pytest.mark.asyncio
    async def test_model_config_provider_literal_validation(self):
        """Test that ModelConfig provider is validated"""
        from pydantic import ValidationError

        from moai_adk.web.models.model_config import ModelConfig, ModelTier

        with pytest.raises(ValidationError):
            ModelConfig(
                tier=ModelTier.PLANNER,
                model_id="test-model",
                provider="invalid_provider",
                cost_per_1k_input=0.01,
                cost_per_1k_output=0.01,
            )


class TestTaskClassification:
    """Test cases for TaskClassification model"""

    @pytest.mark.asyncio
    async def test_task_classification_class_exists(self):
        """Test that TaskClassification class exists"""
        from moai_adk.web.models.model_config import TaskClassification

        assert TaskClassification is not None

    @pytest.mark.asyncio
    async def test_task_classification_has_required_fields(self):
        """Test that TaskClassification has required fields"""
        from moai_adk.web.models.model_config import ModelTier, TaskClassification

        classification = TaskClassification(
            task_type="architecture",
            recommended_tier=ModelTier.PLANNER,
            reason="Complex reasoning required",
        )

        assert classification.task_type == "architecture"
        assert classification.recommended_tier == ModelTier.PLANNER
        assert classification.reason == "Complex reasoning required"


class TestModelRouterService:
    """Test cases for ModelRouter service"""

    @pytest.mark.asyncio
    async def test_model_router_class_exists(self):
        """Test that ModelRouter class exists"""
        from moai_adk.web.services.model_router import ModelRouter

        assert ModelRouter is not None

    @pytest.mark.asyncio
    async def test_model_router_can_be_instantiated(self):
        """Test that ModelRouter can be instantiated"""
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        assert router is not None

    @pytest.mark.asyncio
    async def test_model_router_has_models_config(self):
        """Test that ModelRouter has MODELS configuration"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        assert hasattr(router, "MODELS")
        assert ModelTier.PLANNER in router.MODELS
        assert ModelTier.IMPLEMENTER in router.MODELS

    @pytest.mark.asyncio
    async def test_classify_task_returns_task_classification(self):
        """Test that classify_task returns TaskClassification"""
        from moai_adk.web.models.model_config import TaskClassification
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        result = router.classify_task("Design the API architecture")

        assert isinstance(result, TaskClassification)

    @pytest.mark.asyncio
    async def test_classify_task_planner_for_architecture(self):
        """Test that architecture tasks use PLANNER tier"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        result = router.classify_task("Design the system architecture")

        assert result.recommended_tier == ModelTier.PLANNER

    @pytest.mark.asyncio
    async def test_classify_task_planner_for_moai_1_plan(self):
        """Test that /moai:1-plan tasks use PLANNER tier"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        result = router.classify_task("/moai:1-plan Create user authentication")

        assert result.recommended_tier == ModelTier.PLANNER

    @pytest.mark.asyncio
    async def test_classify_task_implementer_for_coding(self):
        """Test that coding tasks use IMPLEMENTER tier"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        result = router.classify_task("Implement the login function")

        assert result.recommended_tier == ModelTier.IMPLEMENTER

    @pytest.mark.asyncio
    async def test_classify_task_implementer_for_moai_2_run(self):
        """Test that /moai:2-run tasks use IMPLEMENTER tier"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        result = router.classify_task("/moai:2-run SPEC-001")

        assert result.recommended_tier == ModelTier.IMPLEMENTER

    @pytest.mark.asyncio
    async def test_classify_task_default_to_implementer(self):
        """Test that unknown tasks default to IMPLEMENTER tier"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        result = router.classify_task("Some random task")

        assert result.recommended_tier == ModelTier.IMPLEMENTER

    @pytest.mark.asyncio
    async def test_get_model_for_tier_returns_model_config(self):
        """Test that get_model_for_tier returns ModelConfig"""
        from moai_adk.web.models.model_config import ModelConfig, ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        config = router.get_model_for_tier(ModelTier.PLANNER)

        assert isinstance(config, ModelConfig)

    @pytest.mark.asyncio
    async def test_get_model_for_planner_tier(self):
        """Test that PLANNER tier returns Opus model"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        config = router.get_model_for_tier(ModelTier.PLANNER)

        assert config.provider == "claude"
        assert "opus" in config.model_id.lower()

    @pytest.mark.asyncio
    async def test_get_model_for_implementer_tier(self):
        """Test that IMPLEMENTER tier returns GLM model"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        config = router.get_model_for_tier(ModelTier.IMPLEMENTER)

        assert config.provider == "glm"
        assert "glm" in config.model_id.lower()

    @pytest.mark.asyncio
    async def test_get_environment_for_model_returns_dict(self):
        """Test that get_environment_for_model returns dict"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        config = router.get_model_for_tier(ModelTier.PLANNER)
        env = router.get_environment_for_model(config)

        assert isinstance(env, dict)

    @pytest.mark.asyncio
    async def test_get_environment_has_model_key(self):
        """Test that environment has MODEL key"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        config = router.get_model_for_tier(ModelTier.PLANNER)
        env = router.get_environment_for_model(config)

        assert "ANTHROPIC_MODEL" in env or "MODEL" in env

    @pytest.mark.asyncio
    async def test_switch_to_tier_changes_active_tier(self):
        """Test that switch_to_tier changes the active tier"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        router.switch_to_tier(ModelTier.PLANNER)

        assert router.get_active_tier() == ModelTier.PLANNER

    @pytest.mark.asyncio
    async def test_get_active_tier_default(self):
        """Test that default active tier is IMPLEMENTER"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        # Reset to default
        router.reset_to_default()

        assert router.get_active_tier() == ModelTier.IMPLEMENTER


class TestPlanningKeywords:
    """Test classification for planning keywords"""

    @pytest.mark.parametrize(
        "task_description",
        [
            "/moai:1-plan Create user authentication",
            "architecture review for API design",
            "design the database schema",
            "strategy for microservices migration",
            "complex decision about caching",
            "trade-off analysis between SQL and NoSQL",
            "evaluate different frameworks",
        ],
    )
    @pytest.mark.asyncio
    async def test_planning_keywords_use_planner(self, task_description):
        """Test that planning keywords result in PLANNER tier"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        result = router.classify_task(task_description)

        assert result.recommended_tier == ModelTier.PLANNER


class TestImplementationKeywords:
    """Test classification for implementation keywords"""

    @pytest.mark.parametrize(
        "task_description",
        [
            "/moai:2-run SPEC-001",
            "implement the login function",
            "code the API endpoint",
            "test the user service",
            "TDD for authentication module",
            "refactor the database layer",
            "fix the null pointer bug",
            "update the configuration file",
            "create the model class",
        ],
    )
    @pytest.mark.asyncio
    async def test_implementation_keywords_use_implementer(self, task_description):
        """Test that implementation keywords result in IMPLEMENTER tier"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        result = router.classify_task(task_description)

        assert result.recommended_tier == ModelTier.IMPLEMENTER


# TAG-005: Router REST API Tests
import pytest_asyncio
from httpx import ASGITransport, AsyncClient


class TestModelRouterAPI:
    """Test cases for Model Router REST API"""

    @pytest_asyncio.fixture
    async def client(self, tmp_path):
        """Create a test client with a temporary database"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database
        from moai_adk.web.server import create_app
        from moai_adk.web.services.model_router import ModelRouter

        # Reset router to default
        ModelRouter().reset_to_default()

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        app = create_app(config)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
            yield client

        await close_database()

    @pytest.mark.asyncio
    async def test_classify_task_endpoint_returns_200(self, client):
        """Test that POST /api/router/classify returns 200"""
        response = await client.post(
            "/api/router/classify",
            json={"task_description": "Design the API architecture"},
        )
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_classify_task_endpoint_returns_classification(self, client):
        """Test that classify endpoint returns classification data"""
        response = await client.post(
            "/api/router/classify",
            json={"task_description": "Design the API architecture"},
        )
        data = response.json()

        assert "task_type" in data
        assert "recommended_tier" in data
        assert "reason" in data

    @pytest.mark.asyncio
    async def test_classify_task_planner_via_api(self, client):
        """Test that architecture task is classified as planner via API"""
        response = await client.post(
            "/api/router/classify",
            json={"task_description": "Design the system architecture"},
        )
        data = response.json()

        assert data["recommended_tier"] == "planner"

    @pytest.mark.asyncio
    async def test_classify_task_implementer_via_api(self, client):
        """Test that coding task is classified as implementer via API"""
        response = await client.post(
            "/api/router/classify",
            json={"task_description": "Implement the login function"},
        )
        data = response.json()

        assert data["recommended_tier"] == "implementer"

    @pytest.mark.asyncio
    async def test_get_config_endpoint_returns_200(self, client):
        """Test that GET /api/router/config returns 200"""
        response = await client.get("/api/router/config")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_get_config_endpoint_returns_models(self, client):
        """Test that config endpoint returns model configurations"""
        response = await client.get("/api/router/config")
        data = response.json()

        assert "models" in data
        assert "planner" in data["models"]
        assert "implementer" in data["models"]

    @pytest.mark.asyncio
    async def test_switch_tier_endpoint_returns_200(self, client):
        """Test that POST /api/router/switch returns 200"""
        response = await client.post(
            "/api/router/switch",
            json={"tier": "planner"},
        )
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_switch_tier_changes_active_tier(self, client):
        """Test that switch tier endpoint changes the active tier"""
        await client.post("/api/router/switch", json={"tier": "planner"})

        response = await client.get("/api/router/config")
        data = response.json()

        assert data["active_tier"] == "planner"

    @pytest.mark.asyncio
    async def test_switch_tier_invalid_returns_400(self, client):
        """Test that switching to invalid tier returns 400"""
        response = await client.post(
            "/api/router/switch",
            json={"tier": "invalid_tier"},
        )
        assert response.status_code == 400

    @pytest.mark.asyncio
    async def test_get_environment_endpoint_returns_200(self, client):
        """Test that GET /api/router/environment/{tier} returns 200"""
        response = await client.get("/api/router/environment/planner")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_get_environment_returns_env_vars(self, client):
        """Test that environment endpoint returns environment variables"""
        response = await client.get("/api/router/environment/planner")
        data = response.json()

        assert "environment" in data
        assert "MODEL" in data["environment"] or "ANTHROPIC_MODEL" in data["environment"]

    @pytest.mark.asyncio
    async def test_get_environment_invalid_tier_returns_400(self, client):
        """Test that getting environment for invalid tier returns 400"""
        response = await client.get("/api/router/environment/invalid_tier")
        assert response.status_code == 400


class TestModelRouterFallback:
    """Test cases for model router fallback functionality"""

    @pytest.mark.asyncio
    async def test_get_fallback_tier_planner_to_implementer(self):
        """Test that PLANNER tier falls back to IMPLEMENTER"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        fallback = router.get_fallback_tier(ModelTier.PLANNER)

        assert fallback == ModelTier.IMPLEMENTER

    @pytest.mark.asyncio
    async def test_get_fallback_tier_implementer_no_fallback(self):
        """Test that IMPLEMENTER tier has no fallback (returns itself)"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        fallback = router.get_fallback_tier(ModelTier.IMPLEMENTER)

        assert fallback == ModelTier.IMPLEMENTER

    @pytest.mark.asyncio
    async def test_route_with_fallback_primary_success(self):
        """Test route_with_fallback returns primary model when not failed"""
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        config, fallback_used = router.route_with_fallback(
            "design the API architecture",
            primary_failed=False,
        )

        assert config is not None
        assert fallback_used is False

    @pytest.mark.asyncio
    async def test_route_with_fallback_primary_failed(self):
        """Test route_with_fallback uses fallback when primary fails"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        config, fallback_used = router.route_with_fallback(
            "design the API architecture",  # Would normally use PLANNER
            primary_failed=True,
        )

        assert config is not None
        assert config.tier == ModelTier.IMPLEMENTER
        assert fallback_used is True

    @pytest.mark.asyncio
    async def test_route_with_fallback_implementer_no_change(self):
        """Test that IMPLEMENTER doesn't change even with failure"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        config, fallback_used = router.route_with_fallback(
            "implement the feature",  # Would use IMPLEMENTER
            primary_failed=True,
        )

        assert config is not None
        assert config.tier == ModelTier.IMPLEMENTER
        assert fallback_used is False  # No fallback because already IMPLEMENTER

    @pytest.mark.asyncio
    async def test_is_fallback_available_planner(self):
        """Test that fallback is available for PLANNER"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        assert router.is_fallback_available(ModelTier.PLANNER) is True

    @pytest.mark.asyncio
    async def test_is_fallback_available_implementer(self):
        """Test that fallback is not available for IMPLEMENTER"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRouter

        router = ModelRouter()
        assert router.is_fallback_available(ModelTier.IMPLEMENTER) is False


class TestModelRoutingError:
    """Test cases for ModelRoutingError exception"""

    @pytest.mark.asyncio
    async def test_model_routing_error_exists(self):
        """Test that ModelRoutingError exception exists"""
        from moai_adk.web.services.model_router import ModelRoutingError

        assert ModelRoutingError is not None

    @pytest.mark.asyncio
    async def test_model_routing_error_creation(self):
        """Test creating a ModelRoutingError instance"""
        from moai_adk.web.models.model_config import ModelTier
        from moai_adk.web.services.model_router import ModelRoutingError

        error = ModelRoutingError(
            message="Model unavailable",
            tier=ModelTier.PLANNER,
            fallback_used=True,
        )

        assert error.message == "Model unavailable"
        assert error.tier == ModelTier.PLANNER
        assert error.fallback_used is True

    @pytest.mark.asyncio
    async def test_model_routing_error_is_exception(self):
        """Test that ModelRoutingError is an Exception subclass"""
        from moai_adk.web.services.model_router import ModelRoutingError

        assert issubclass(ModelRoutingError, Exception)
