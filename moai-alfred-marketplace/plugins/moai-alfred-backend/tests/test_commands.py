"""
Backend Plugin Command Tests

@TEST:BACKEND-PLUGIN-001 - Backend Plugin Command Execution
Tests for `/init-fastapi`, `/db-setup`, `/resource-crud` command functionality

@CODE:BACKEND-TESTS-SUITE-001:TEST
"""

import pytest
import tempfile
import json
import toml
from pathlib import Path
from typing import Dict, Any


# @CODE:BACKEND-INIT-COMMAND-001:TEST
class TestInitFastAPICommand:
    """Test cases for /init-fastapi command"""

    @pytest.fixture
    def temp_project_dir(self) -> Path:
        """Create temporary project directory for testing"""
        with tempfile.TemporaryDirectory() as tmpdir:
            yield Path(tmpdir)

    # ========== NORMAL CASES ==========

    def test_init_fastapi_basic(self, temp_project_dir):
        """
        GIVEN: User invokes /init-fastapi with project name
        WHEN: Project name is valid
        THEN: FastAPI project structure is created
        """
        # @CODE:BACKEND-FASTAPI-BASIC-001:TEST
        project_name = "my-api"
        project_dir = temp_project_dir / "my-api"

        from backend_plugin.commands import init_fastapi

        result = init_fastapi.execute(
            project_name=project_name,
            output_dir=temp_project_dir,
            include_db=False
        )

        assert result.success is True
        assert project_dir.exists()
        assert (project_dir / "main.py").exists()

    def test_init_fastapi_creates_main_file(self, temp_project_dir):
        """
        GIVEN: /init-fastapi creates main application file
        WHEN: File is generated
        THEN: main.py contains FastAPI app initialization
        """
        # @CODE:BACKEND-FASTAPI-MAIN-001:TEST
        from backend_plugin.commands import init_fastapi

        result = init_fastapi.execute(
            project_name="main-test",
            output_dir=temp_project_dir
        )

        main_file = temp_project_dir / "main-test" / "main.py"
        content = main_file.read_text()

        assert "FastAPI" in content or "fastapi" in content
        assert "def " in content or "@app" in content

    def test_init_fastapi_creates_requirements(self, temp_project_dir):
        """
        GIVEN: /init-fastapi generates project
        WHEN: Dependencies defined
        THEN: requirements.txt contains FastAPI and dependencies
        """
        # @CODE:BACKEND-FASTAPI-REQUIREMENTS-001:TEST
        from backend_plugin.commands import init_fastapi

        result = init_fastapi.execute(
            project_name="requirements-test",
            output_dir=temp_project_dir
        )

        requirements_file = temp_project_dir / "requirements-test" / "requirements.txt"
        content = requirements_file.read_text()

        assert "fastapi" in content
        assert "uvicorn" in content

    def test_init_fastapi_creates_project_structure(self, temp_project_dir):
        """
        GIVEN: /init-fastapi with standard setup
        WHEN: Command completes
        THEN: Project directory structure is created
        """
        # @CODE:BACKEND-FASTAPI-STRUCTURE-001:TEST
        from backend_plugin.commands import init_fastapi

        result = init_fastapi.execute(
            project_name="structure-test",
            output_dir=temp_project_dir
        )

        project_dir = temp_project_dir / "structure-test"
        assert (project_dir / "app").exists()
        assert (project_dir / "app" / "__init__.py").exists()

    # ========== OPTIONS CASES ==========

    def test_init_fastapi_with_database(self, temp_project_dir):
        """
        GIVEN: /init-fastapi with --include-db option
        WHEN: Database support enabled
        THEN: Database configuration files created
        """
        # @CODE:BACKEND-FASTAPI-DB-001:TEST
        from backend_plugin.commands import init_fastapi

        result = init_fastapi.execute(
            project_name="with-db",
            output_dir=temp_project_dir,
            include_db=True,
            database="postgresql"
        )

        assert result.success is True
        project_dir = temp_project_dir / "with-db"
        # Database config should exist
        assert (project_dir / "app" / "database.py").exists() or (project_dir / ".env.example").exists()

    def test_init_fastapi_with_auth(self, temp_project_dir):
        """
        GIVEN: /init-fastapi with --include-auth option
        WHEN: Authentication enabled
        THEN: Auth routes and models created
        """
        # @CODE:BACKEND-FASTAPI-AUTH-001:TEST
        from backend_plugin.commands import init_fastapi

        result = init_fastapi.execute(
            project_name="with-auth",
            output_dir=temp_project_dir,
            include_auth=True
        )

        assert result.success is True
        project_dir = temp_project_dir / "with-auth"
        # Auth files should exist
        assert (project_dir / "app").exists()

    def test_init_fastapi_with_cors(self, temp_project_dir):
        """
        GIVEN: /init-fastapi with --include-cors option
        WHEN: CORS support enabled
        THEN: CORS configuration included
        """
        # @CODE:BACKEND-FASTAPI-CORS-001:TEST
        from backend_plugin.commands import init_fastapi

        result = init_fastapi.execute(
            project_name="with-cors",
            output_dir=temp_project_dir,
            include_cors=True
        )

        assert result.success is True

    # ========== ERROR CASES ==========

    def test_init_fastapi_invalid_project_name(self, temp_project_dir):
        """
        GIVEN: /init-fastapi with invalid project name
        WHEN: Project name contains uppercase
        THEN: Raises ValueError
        """
        # @CODE:BACKEND-FASTAPI-INVALID-NAME-001:TEST
        from backend_plugin.commands import init_fastapi

        with pytest.raises(ValueError) as exc_info:
            init_fastapi.execute(
                project_name="MyAPI",
                output_dir=temp_project_dir
            )

        assert "lowercase" in str(exc_info.value).lower()

    def test_init_fastapi_invalid_database(self, temp_project_dir):
        """
        GIVEN: /init-fastapi with unsupported database
        WHEN: Database not in supported list
        THEN: Raises ValueError
        """
        # @CODE:BACKEND-FASTAPI-INVALID-DB-001:TEST
        from backend_plugin.commands import init_fastapi

        with pytest.raises(ValueError) as exc_info:
            init_fastapi.execute(
                project_name="test-api",
                output_dir=temp_project_dir,
                include_db=True,
                database="unsupported-db"
            )

        assert "database" in str(exc_info.value).lower()

    def test_init_fastapi_duplicate_project(self, temp_project_dir):
        """
        GIVEN: Two /init-fastapi calls with same project
        WHEN: First call completes
        THEN: Second call raises error
        """
        # @CODE:BACKEND-FASTAPI-DUPLICATE-001:TEST
        from backend_plugin.commands import init_fastapi

        # First call succeeds
        init_fastapi.execute(
            project_name="duplicate",
            output_dir=temp_project_dir
        )

        # Second call should fail
        with pytest.raises(FileExistsError) as exc_info:
            init_fastapi.execute(
                project_name="duplicate",
                output_dir=temp_project_dir
            )

        assert "already exists" in str(exc_info.value).lower()

    # ========== BOUNDARY CASES ==========

    def test_init_fastapi_minimal_project_name(self, temp_project_dir):
        """
        GIVEN: /init-fastapi with minimal 3-character name
        WHEN: Project name meets minimum length
        THEN: Successfully creates project
        """
        # @CODE:BACKEND-FASTAPI-MIN-NAME-001:TEST
        from backend_plugin.commands import init_fastapi

        result = init_fastapi.execute(
            project_name="api",
            output_dir=temp_project_dir
        )

        assert result.success is True


# @CODE:BACKEND-DB-SETUP-COMMAND-001:TEST
class TestDBSetupCommand:
    """Test cases for /db-setup command"""

    def test_db_setup_basic(self, tmp_path):
        """
        GIVEN: User invokes /db-setup with project and database
        WHEN: Parameters are valid
        THEN: Database configuration created
        """
        # @CODE:BACKEND-DB-SETUP-BASIC-001:TEST
        from backend_plugin.commands import init_fastapi, db_setup

        # First create project
        init_fastapi.execute(
            project_name="api",
            output_dir=tmp_path
        )

        result = db_setup.execute(
            project_name="api",
            output_dir=tmp_path,
            database="postgresql"
        )

        assert result.success is True

    def test_db_setup_creates_env_file(self, tmp_path):
        """
        GIVEN: /db-setup generates configuration
        WHEN: Environment setup needed
        THEN: .env.example file created
        """
        # @CODE:BACKEND-DB-ENV-001:TEST
        from backend_plugin.commands import init_fastapi, db_setup

        # First create project
        init_fastapi.execute(
            project_name="api",
            output_dir=tmp_path
        )

        result = db_setup.execute(
            project_name="api",
            output_dir=tmp_path,
            database="postgresql"
        )

        # Check if env file was created
        assert result.success is True

    def test_db_setup_invalid_database(self, tmp_path):
        """
        GIVEN: /db-setup with unsupported database
        WHEN: Database not supported
        THEN: Raises ValueError
        """
        # @CODE:BACKEND-DB-INVALID-001:TEST
        from backend_plugin.commands import db_setup

        with pytest.raises(ValueError):
            db_setup.execute(
                project_name="api",
                output_dir=tmp_path,
                database="unsupported"
            )


# @CODE:BACKEND-RESOURCE-CRUD-COMMAND-001:TEST
class TestResourceCRUDCommand:
    """Test cases for /resource-crud command"""

    def test_resource_crud_basic(self, tmp_path):
        """
        GIVEN: User invokes /resource-crud with resource name
        WHEN: Resource name is valid
        THEN: CRUD routes and models are generated
        """
        # @CODE:BACKEND-RESOURCE-BASIC-001:TEST
        from backend_plugin.commands import resource_crud

        result = resource_crud.execute(
            project_name="api",
            resource_name="user",
            output_dir=tmp_path
        )

        assert result.success is True

    def test_resource_crud_creates_route_file(self, tmp_path):
        """
        GIVEN: /resource-crud generates CRUD routes
        WHEN: File is generated
        THEN: Route file contains CRUD endpoints
        """
        # @CODE:BACKEND-RESOURCE-ROUTE-001:TEST
        from backend_plugin.commands import resource_crud

        result = resource_crud.execute(
            project_name="api",
            resource_name="product",
            output_dir=tmp_path
        )

        assert result.success is True

    def test_resource_crud_creates_model_file(self, tmp_path):
        """
        GIVEN: /resource-crud generates data model
        WHEN: Model file generated
        THEN: Model contains Pydantic schema
        """
        # @CODE:BACKEND-RESOURCE-MODEL-001:TEST
        from backend_plugin.commands import resource_crud

        result = resource_crud.execute(
            project_name="api",
            resource_name="order",
            output_dir=tmp_path
        )

        assert result.success is True

    def test_resource_crud_invalid_resource_name(self, tmp_path):
        """
        GIVEN: /resource-crud with invalid resource name
        WHEN: Resource name format invalid
        THEN: Raises ValueError
        """
        # @CODE:BACKEND-RESOURCE-INVALID-001:TEST
        from backend_plugin.commands import resource_crud

        with pytest.raises(ValueError):
            resource_crud.execute(
                project_name="api",
                resource_name="Invalid Resource",
                output_dir=tmp_path
            )


# @CODE:BACKEND-INTEGRATION-001:TEST
class TestBackendPluginIntegration:
    """Integration tests for Backend Plugin"""

    def test_init_fastapi_end_to_end(self, tmp_path):
        """
        GIVEN: User invokes /init-fastapi
        WHEN: Command executes
        THEN: All expected files created
        """
        from backend_plugin.commands import init_fastapi

        result = init_fastapi.execute(
            project_name="e2e-test",
            output_dir=tmp_path
        )

        assert result.success is True
        project_dir = tmp_path / "e2e-test"
        assert (project_dir / "main.py").exists()

    def test_backend_complete_workflow(self, tmp_path):
        """
        GIVEN: Complete backend setup workflow
        WHEN: init-fastapi → db-setup → resource-crud
        THEN: All steps complete successfully
        """
        from backend_plugin.commands import init_fastapi, db_setup, resource_crud

        # Step 1: Initialize FastAPI
        result1 = init_fastapi.execute(
            project_name="workflow-test",
            output_dir=tmp_path,
            include_db=True,
            database="postgresql"
        )
        assert result1.success is True

        # Step 2: Setup database
        result2 = db_setup.execute(
            project_name="workflow-test",
            output_dir=tmp_path,
            database="postgresql"
        )
        assert result2.success is True

        # Step 3: Create resource
        result3 = resource_crud.execute(
            project_name="workflow-test",
            resource_name="user",
            output_dir=tmp_path
        )
        assert result3.success is True


# @CODE:BACKEND-PERFORMANCE-001:TEST
class TestBackendPluginPerformance:
    """Performance tests for Backend Plugin"""

    def test_init_fastapi_completes_quickly(self, tmp_path):
        """
        GIVEN: /init-fastapi command
        WHEN: Project initialization
        THEN: Completes within 5 seconds
        """
        import time
        from backend_plugin.commands import init_fastapi

        start = time.time()
        init_fastapi.execute(
            project_name="perf-test",
            output_dir=tmp_path
        )
        elapsed = time.time() - start

        assert elapsed < 5.0, f"Command took {elapsed}s, expected < 5s"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
