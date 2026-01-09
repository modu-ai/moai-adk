"""Test web config module

Additional tests to improve coverage.
"""

from pathlib import Path


class TestWebConfig:
    """Test cases for WebConfig"""

    def test_config_default_values(self):
        """Test that WebConfig has correct default values"""
        from moai_adk.web.config import WebConfig

        config = WebConfig()

        assert config.host == "127.0.0.1"
        assert config.port == 8080
        assert config.debug is False

    def test_config_custom_values(self):
        """Test that WebConfig accepts custom values"""
        from moai_adk.web.config import WebConfig

        config = WebConfig(host="0.0.0.0", port=3000, debug=True)

        assert config.host == "0.0.0.0"
        assert config.port == 3000
        assert config.debug is True

    def test_config_database_path_is_path_object(self):
        """Test that database_path is converted to Path"""
        from moai_adk.web.config import WebConfig

        config = WebConfig(database_path="/tmp/test.db")

        assert isinstance(config.database_path, Path)

    def test_config_cors_origins_default(self):
        """Test that cors_origins has default values"""
        from moai_adk.web.config import WebConfig

        config = WebConfig()

        assert isinstance(config.cors_origins, list)
        assert len(config.cors_origins) > 0


class TestServerApp:
    """Test cases for server app creation"""

    def test_create_app_returns_fastapi_instance(self):
        """Test that create_app returns a FastAPI instance"""
        from fastapi import FastAPI

        from moai_adk.web.server import create_app

        app = create_app()

        assert isinstance(app, FastAPI)

    def test_create_app_with_custom_config(self):
        """Test that create_app accepts custom config"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.server import create_app

        config = WebConfig(port=9000)
        app = create_app(config)

        assert app is not None

    def test_app_has_routers_registered(self):
        """Test that app has routers registered"""
        from moai_adk.web.server import create_app

        app = create_app()

        # Check that routes are registered
        routes = [route.path for route in app.routes]

        assert any("/api/health" in route for route in routes)
        assert any("/api/sessions" in route for route in routes)
        assert any("/api/providers" in route for route in routes)
        assert any("/api/specs" in route for route in routes)
