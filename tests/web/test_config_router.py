"""Config Router Tests

TDD RED phase: Write failing tests first for Config API endpoints.

Tests cover:
- GET /api/config/schema - Get tab schema v3.0.0
- GET /api/config - Get current configuration
- POST /api/config - Save configuration with backup
- POST /api/config/validate - Validate configuration changes
"""

from pathlib import Path

import pytest
import yaml
from fastapi.testclient import TestClient

from moai_adk.web.server import app


@pytest.fixture
def client(tmp_path: Path) -> TestClient:
    """Create test client with temp config directory."""

    def _override_config_dir():
        import moai_adk.web.routers.config as config_module

        original_dir = config_module.CONFIG_DIR
        config_module.CONFIG_DIR = tmp_path / ".moai" / "config"
        yield
        config_module.CONFIG_DIR = original_dir

    # Patch the config directory before creating client
    import moai_adk.web.routers.config as config_module

    original_dir = config_module.CONFIG_DIR
    config_module.CONFIG_DIR = tmp_path / ".moai" / "config"

    yield TestClient(app)

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
        config_dir.mkdir(parents=True)

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

        response = client.get("/api/config")

        assert response.status_code == 200
        data = response.json()
        assert data["user"]["name"] == "Test User"
        assert data["language"]["conversation_language"] == "ko"

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

        # Should have smart defaults applied
        assert "git_strategy" in data or "constitution" in data


class TestSaveConfigEndpoint:
    """Tests for POST /api/config endpoint."""

    def test_save_config_stores_to_file(self, client: TestClient, tmp_path: Path):
        """Should save configuration to file."""
        config_data = {
            "user": {"name": "New User"},
            "language": {"conversation_language": "en"},
            "project": {"name": "new-project"},
            "git_strategy": {"mode": "personal"},
            "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
        }

        response = client.post("/api/config", json=config_data)

        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True

    def test_save_config_creates_backup(self, client: TestClient, tmp_path: Path):
        """Should create backup before overwriting."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)

        original_config = {
            "user": {"name": "Original User"},
            "language": {"conversation_language": "en"},
        }
        config_file = config_dir / "config.yaml"
        config_file.write_text(yaml.dump(original_config))

        new_config = {
            "user": {"name": "New User"},
            "language": {"conversation_language": "en"},
        }

        response = client.post("/api/config", json=new_config)

        assert response.status_code == 200
        # Backup file should exist
        backup_file = config_dir / "config.yaml.backup"
        assert backup_file.exists()

    def test_save_config_validates_required_fields(self, client: TestClient):
        """Should return error for missing required fields."""
        invalid_config = {"user": {"name": "Test"}}

        response = client.post("/api/config", json=invalid_config)

        assert response.status_code == 422
        data = response.json()
        assert "detail" in data


class TestValidateConfigEndpoint:
    """Tests for POST /api/config/validate endpoint."""

    def test_validate_returns_valid_for_complete_config(self, client: TestClient):
        """Should validate as complete for proper configuration."""
        config = {
            "user": {"name": "Test User"},
            "language": {"conversation_language": "en", "agent_prompt_language": "en"},
            "project": {"name": "test-project"},
            "git_strategy": {"mode": "personal"},
            "constitution": {"test_coverage_target": 90, "enforce_tdd": True},
        }

        response = client.post("/api/config/validate", json=config)

        assert response.status_code == 200
        data = response.json()
        assert data["valid"] is True

    def test_validate_returns_invalid_for_missing_fields(self, client: TestClient):
        """Should validate as invalid for incomplete configuration."""
        incomplete_config = {"user": {"name": "Test"}}

        response = client.post("/api/config/validate", json=incomplete_config)

        assert response.status_code == 200
        data = response.json()
        assert data["valid"] is False
        assert len(data["missing_fields"]) > 0

    def test_validate_lists_all_required_fields(self, client: TestClient):
        """Should list all missing required fields."""
        empty_config = {}

        response = client.post("/api/config/validate", json=empty_config)

        assert response.status_code == 200
        data = response.json()
        assert "missing_fields" in data
        assert isinstance(data["missing_fields"], list)


class TestConditionalBatches:
    """Tests for conditional batch visibility."""

    def test_personal_mode_shows_personal_batches(self, client: TestClient):
        """Should show personal batches when mode is personal."""
        response = client.get("/api/config/schema")

        assert response.status_code == 200
        # Schema includes conditional info
        data = response.json()
        tab3 = data["tabs"][2]
        personal_batch = tab3["batches"][0]
        assert personal_batch["id"] == "batch_3_1_personal"

    def test_conditional_batches_have_show_if(self, client: TestClient):
        """Should include show_if condition for conditional batches."""
        response = client.get("/api/config/schema")

        assert response.status_code == 200
        data = response.json()

        for tab in data["tabs"]:
            for batch in tab["batches"]:
                if "show_if" in batch:
                    assert isinstance(batch["show_if"], str)
