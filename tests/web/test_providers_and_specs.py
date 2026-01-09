"""TAG-006: Test Provider and SPEC API

RED Phase: Tests for verifying provider switch and SPEC status endpoints.
"""

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient


class TestProviderService:
    """Test cases for ProviderService"""

    @pytest.mark.asyncio
    async def test_provider_service_class_exists(self):
        """Test that ProviderService class exists"""
        from moai_adk.web.services.provider_service import ProviderService

        assert ProviderService is not None

    @pytest.mark.asyncio
    async def test_provider_service_can_be_instantiated(self):
        """Test that ProviderService can be instantiated"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        assert service is not None

    @pytest.mark.asyncio
    async def test_get_available_providers_returns_dict(self):
        """Test that get_available_providers returns a dictionary"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        providers = service.get_available_providers()

        assert isinstance(providers, dict)
        assert len(providers) > 0

    @pytest.mark.asyncio
    async def test_claude_provider_is_available(self):
        """Test that Claude provider is available"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        providers = service.get_available_providers()

        assert "claude" in providers

    @pytest.mark.asyncio
    async def test_get_active_provider_returns_string(self):
        """Test that get_active_provider returns a string"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        provider = service.get_active_provider()

        assert isinstance(provider, str)
        assert len(provider) > 0

    @pytest.mark.asyncio
    async def test_get_active_model_returns_string(self):
        """Test that get_active_model returns a string"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        model = service.get_active_model()

        assert isinstance(model, str)
        assert len(model) > 0

    @pytest.mark.asyncio
    async def test_switch_provider_returns_bool(self):
        """Test that switch_provider returns a boolean"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        result = service.switch_provider("claude")

        assert isinstance(result, bool)

    @pytest.mark.asyncio
    async def test_switch_provider_changes_active_provider(self):
        """Test that switch_provider changes the active provider"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        service.switch_provider("openai")

        assert service.get_active_provider() == "openai"

        # Reset to default
        service.reset_to_default()

    @pytest.mark.asyncio
    async def test_switch_provider_invalid_returns_false(self):
        """Test that switch_provider returns False for invalid provider"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        result = service.switch_provider("invalid_provider")

        assert result is False


class TestProvidersAPI:
    """Test cases for Providers REST API"""

    @pytest_asyncio.fixture
    async def client(self, tmp_path):
        """Create a test client with a temporary database"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database
        from moai_adk.web.server import create_app
        from moai_adk.web.services.provider_service import ProviderService

        # Reset provider to default
        ProviderService().reset_to_default()

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        app = create_app(config)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
            yield client

        await close_database()

    @pytest.mark.asyncio
    async def test_list_providers_returns_200(self, client):
        """Test that GET /api/providers returns 200"""
        response = await client.get("/api/providers")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_list_providers_returns_providers_list(self, client):
        """Test that list providers returns providers array"""
        response = await client.get("/api/providers")
        data = response.json()

        assert "providers" in data
        assert "active_provider" in data
        assert len(data["providers"]) > 0

    @pytest.mark.asyncio
    async def test_switch_provider_returns_200(self, client):
        """Test that POST /api/providers/switch returns 200"""
        response = await client.post("/api/providers/switch", json={"provider": "claude"})
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_switch_provider_changes_provider(self, client):
        """Test that switch provider changes the active provider"""
        await client.post("/api/providers/switch", json={"provider": "openai"})

        response = await client.get("/api/providers/current")
        data = response.json()

        assert data["provider"] == "openai"

    @pytest.mark.asyncio
    async def test_switch_provider_invalid_returns_400(self, client):
        """Test that switching to invalid provider returns 400"""
        response = await client.post("/api/providers/switch", json={"provider": "invalid"})
        assert response.status_code == 400

    @pytest.mark.asyncio
    async def test_get_current_provider_returns_200(self, client):
        """Test that GET /api/providers/current returns 200"""
        response = await client.get("/api/providers/current")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_get_current_provider_returns_provider_info(self, client):
        """Test that get current provider returns provider info"""
        response = await client.get("/api/providers/current")
        data = response.json()

        assert "provider" in data
        assert "model" in data


class TestSpecsAPI:
    """Test cases for SPEC Status REST API"""

    @pytest_asyncio.fixture
    async def client(self, tmp_path):
        """Create a test client with a temporary database"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database
        from moai_adk.web.server import create_app

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        app = create_app(config)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
            yield client

        await close_database()

    @pytest.mark.asyncio
    async def test_list_specs_returns_200(self, client):
        """Test that GET /api/specs returns 200"""
        response = await client.get("/api/specs")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_list_specs_returns_specs_array(self, client):
        """Test that list specs returns specs array"""
        response = await client.get("/api/specs")
        data = response.json()

        assert "specs" in data
        assert "total" in data
        assert isinstance(data["specs"], list)

    @pytest.mark.asyncio
    async def test_get_spec_unknown_returns_404(self, client):
        """Test that GET /api/specs/{id} returns 404 for unknown SPEC"""
        response = await client.get("/api/specs/SPEC-UNKNOWN-999")
        assert response.status_code == 404


class TestSpecStatusModel:
    """Test cases for SpecStatus model"""

    @pytest.mark.asyncio
    async def test_spec_status_model_exists(self):
        """Test that SpecStatus model exists"""
        from moai_adk.web.routers.specs import SpecStatus

        assert SpecStatus is not None

    @pytest.mark.asyncio
    async def test_spec_status_has_required_fields(self):
        """Test that SpecStatus has required fields"""
        from moai_adk.web.routers.specs import SpecStatus

        spec = SpecStatus(spec_id="SPEC-001", title="Test SPEC", status="draft")

        assert spec.spec_id == "SPEC-001"
        assert spec.title == "Test SPEC"
        assert spec.status == "draft"
