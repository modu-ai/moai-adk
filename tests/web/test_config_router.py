"""Config Router Tests

TDD RED phase: Write failing tests first for Config API endpoints.

Tests cover:
- GET /api/config/schema - Get tab schema v3.0.0
- GET /api/config - Get current configuration
- POST /api/config - Save configuration with backup
- POST /api/config/validate - Validate configuration changes
"""

from pathlib import Path
from typing import Any
from unittest.mock import MagicMock

import pytest
import yaml
from fastapi.testclient import TestClient


@pytest.fixture
def client(tmp_path: Path) -> TestClient:
    """Create test client with temp config directory."""
    from fastapi import FastAPI

    import moai_adk.web.routers.config as config_module

    original_dir = config_module.CONFIG_DIR
    test_config_dir = tmp_path / ".moai" / "config"
    test_config_dir.mkdir(parents=True, exist_ok=True)

    # Set CONFIG_DIR before creating app
    config_module.CONFIG_DIR = test_config_dir

    # Create minimal app with just config router
    app = FastAPI()
    app.include_router(config_module.router, prefix="/api", tags=["config"])

    test_client = TestClient(app)

    yield test_client

    # Restore original directory after test
    config_module.CONFIG_DIR = original_dir


@pytest.fixture
def mock_config_manager():
    """Mock ConfigurationManager for testing."""
    manager = MagicMock()
    manager.load.return_value = {
        "user": {"name": "Test User"},
        "language": {"conversation_language": "en"},
        "project": {"name": "test-project"},
        "git_strategy": {"mode": "personal"},
        "constitution": {"test_coverage_target": 90},
    }
    manager.save.return_value = True
    return manager


class TestGetSchemaEndpoint:
    """Tests for GET /api/schema endpoint."""

    def test_get_schema_returns_tab_schema_v3(self, client: TestClient):
        """Should return tab schema v3.0.0 with 3 tabs."""
        response = client.get("/api/schema")

        assert response.status_code == 200
        data = response.json()
        assert data["version"] == "3.0.0"
        assert len(data["tabs"]) == 3

    def test_schema_has_tab1_quick_start(self, client: TestClient):
        """Should include Tab 1: Quick Start with 4 batches."""
        response = client.get("/api/schema")
        data = response.json()

        tab1 = data["tabs"][0]
        assert tab1["id"] == "tab_1_quick_start"
        assert tab1["label"] == "Essential Setup"
        assert len(tab1["batches"]) == 4

    def test_schema_has_tab2_documentation(self, client: TestClient):
        """Should include Tab 2: Documentation with 2 batches."""
        response = client.get("/api/schema")
        data = response.json()

        tab2 = data["tabs"][1]
        assert tab2["id"] == "tab_2_documentation"
        assert tab2["label"] == "Documentation"
        assert len(tab2["batches"]) == 2

    def test_schema_has_tab3_git_automation(self, client: TestClient):
        """Should include Tab 3: Git Automation with 2 conditional batches."""
        response = client.get("/api/schema")
        data = response.json()

        tab3 = data["tabs"][2]
        assert tab3["id"] == "tab_3_git_automation"
        assert tab3["label"] == "Git"
        assert len(tab3["batches"]) == 2

    def test_schema_question_types_are_valid(self, client: TestClient):
        """Should only have valid question types."""
        response = client.get("/api/schema")
        data = response.json()

        valid_types = {"text_input", "select_single", "number_input"}

        for tab in data["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    assert question["type"] in valid_types


class TestGetConfigEndpoint:
    """Tests for GET /api/config endpoint."""

    def test_get_config_returns_current_config(self, client: TestClient, tmp_path: Path):
        """Should return current configuration."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)

        config_file = config_dir / "config.yaml"
        config_file.write_text(
            yaml.dump(
                {
                    "user": {"name": "Test User"},
                    "language": {"conversation_language": "ko"},
                    "project": {"name": "my-project"},
                }
            )
        )

        # Test with direct function call instead of HTTP
        import moai_adk.web.routers.config as config_module

        original_dir = config_module.CONFIG_DIR
        config_module.CONFIG_DIR = config_dir

        try:
            manager = config_module.get_config_manager()
            config = manager.load()

            assert config["user"]["name"] == "Test User"
            assert config["language"]["conversation_language"] == "ko"
            assert config["project"]["name"] == "my-project"
        finally:
            config_module.CONFIG_DIR = original_dir

    def test_get_config_returns_defaults_when_no_file(self, client: TestClient):
        """Should return default/empty config when file doesn't exist."""
        response = client.get("/api/config")

        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, dict)

    def test_get_config_applies_smart_defaults(self, client: TestClient):
        """Should apply smart defaults to missing fields."""
        response = client.get("/api/config")

        assert response.status_code == 200
        data = response.json()

        # Should have basic structure with smart defaults applied
        assert "user" in data or "language" in data or "project" in data
        # When config file exists, returns loaded data; otherwise returns empty structure


class TestSaveConfigEndpoint:
    """Tests for POST /api/config endpoint."""

    def test_save_config_stores_to_file(self, client: TestClient, tmp_path: Path):
        """Should save configuration to file."""
        config_data = {
            "user": {"name": "New User"},
            "language": {"conversation_language": "en"},
            "project": {"name": "new-project", "documentation_mode": "skip"},
            "git_strategy": {"mode": "personal"},
            "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
        }

        response = client.post("/api/config", json=config_data)

        # Debug: print response
        if response.status_code != 200:
            print(f"Response: {response.status_code}")
            print(f"Detail: {response.json()}")

        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True

    def test_save_config_creates_backup(self, client: TestClient, tmp_path: Path):
        """Should create backup before overwriting."""
        config_dir = tmp_path / ".moai" / "config"
        config_file = config_dir / "config.yaml"

        # Create original config file
        original_config = {
            "user": {"name": "Original User"},
            "language": {"conversation_language": "en"},
        }
        config_file.write_text(yaml.dump(original_config))

        new_config = {
            "user": {"name": "New User"},
            "language": {"conversation_language": "en", "agent_prompt_language": "en"},
            "project": {"name": "new-project", "documentation_mode": "skip"},
            "git_strategy": {"mode": "personal"},
            "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
        }

        # Call the HTTP endpoint which includes backup logic
        response = client.post("/api/config", json=new_config)

        assert response.status_code == 200

        # Check backup file exists (config.yaml.backup, not config.backup)
        backup_file = config_dir / "config.yaml.backup"

        assert backup_file.exists(), f"Backup file not found at {backup_file}"
        assert backup_file.stat().st_size > 0, "Backup file is empty"

        # Verify backup content matches original
        with open(backup_file, "r") as f:
            backup_content = yaml.safe_load(f)
        assert backup_content["user"]["name"] == "Original User"
        assert backup_content["language"]["conversation_language"] == "en"

    def test_save_config_validates_required_fields(self, client: TestClient):
        """Should return error for missing required fields."""
        invalid_config = {"user": {"name": "Test"}}

        response = client.post("/api/config", json=invalid_config)

        assert response.status_code == 422
        data = response.json()
        assert "detail" in data


class TestValidateConfigEndpoint:
    """Tests for POST /api/validate endpoint."""

    def test_validate_returns_valid_for_complete_config(self, client: TestClient):
        """Should validate as complete for proper configuration."""
        config = {
            "user": {"name": "Test User"},
            "language": {"conversation_language": "en", "agent_prompt_language": "en"},
            "project": {"name": "test-project", "documentation_mode": "skip"},
            "git_strategy": {"mode": "personal"},
            "constitution": {"test_coverage_target": 90, "enforce_tdd": True},
        }

        response = client.post("/api/validate", json=config)

        assert response.status_code == 200
        data = response.json()
        assert data["valid"] is True

    def test_validate_returns_invalid_for_missing_fields(self, client: TestClient):
        """Should validate as invalid for incomplete configuration."""
        incomplete_config = {"user": {"name": "Test"}}

        response = client.post("/api/validate", json=incomplete_config)

        assert response.status_code == 200
        data = response.json()
        assert data["valid"] is False
        assert len(data["missing_fields"]) > 0

    def test_validate_lists_all_required_fields(self, client: TestClient):
        """Should list all missing required fields."""
        empty_config: dict[str, Any] = {}

        response = client.post("/api/validate", json=empty_config)

        assert response.status_code == 200
        data = response.json()
        assert "missing_fields" in data
        assert isinstance(data["missing_fields"], list)


class TestConditionalBatches:
    """Tests for conditional batch visibility."""

    def test_personal_mode_shows_personal_batches(self, client: TestClient):
        """Should show personal batches when mode is personal."""
        response = client.get("/api/schema")

        assert response.status_code == 200
        # Schema includes conditional info
        data = response.json()
        tab3 = data["tabs"][2]
        personal_batch = tab3["batches"][0]
        assert personal_batch["id"] == "batch_3_1_personal"

    def test_conditional_batches_have_show_if(self, client: TestClient):
        """Should include show_if condition for conditional batches."""
        response = client.get("/api/schema")

        assert response.status_code == 200
        data = response.json()

        # Check that conditional batches have show_if
        tab3 = data["tabs"][2]  # Git tab
        personal_batch = tab3["batches"][0]
        # Personal batch should have show_if
        assert "show_if" in personal_batch
        assert personal_batch["show_if"] is not None
