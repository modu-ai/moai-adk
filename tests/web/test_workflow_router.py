"""Tests for Workflow Router

TDD RED phase: Tests for workflow REST API endpoints.
Tests should FAIL initially until router is implemented.
"""

import pytest


class TestWorkflowRouterStart:
    """Test POST /api/workflow/start endpoint."""

    @pytest.mark.asyncio
    async def test_start_workflow_success(self):
        """Test starting a workflow successfully."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            response = client.post(
                "/api/workflow/start",
                json={"features": ["user authentication", "dashboard"]},
            )

            assert response.status_code == 201
            data = response.json()
            assert "workflow_id" in data
            assert data["status"] == "in_progress"
            assert data["phase"] == "configuration"

    @pytest.mark.asyncio
    async def test_start_workflow_with_options(self):
        """Test starting workflow with all options."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            response = client.post(
                "/api/workflow/start",
                json={
                    "features": ["api"],
                    "use_worktree": True,
                    "parallel_workers": 3,
                    "create_branch": True,
                    "create_pr": True,
                    "auto_merge": False,
                    "model": "opus",
                },
            )

            assert response.status_code == 201
            data = response.json()
            assert data["config"]["use_worktree"] is True
            assert data["config"]["parallel_workers"] == 3

    @pytest.mark.asyncio
    async def test_start_workflow_empty_features(self):
        """Test starting workflow with empty features returns 422."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            response = client.post(
                "/api/workflow/start",
                json={"features": []},
            )

            assert response.status_code == 422


class TestWorkflowRouterStatus:
    """Test GET /api/workflow/{id}/status endpoint."""

    @pytest.mark.asyncio
    async def test_get_workflow_status(self):
        """Test getting workflow status."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            # First create a workflow
            create_response = client.post(
                "/api/workflow/start",
                json={"features": ["test"]},
            )
            workflow_id = create_response.json()["workflow_id"]

            # Then get its status
            response = client.get(f"/api/workflow/{workflow_id}/status")

            assert response.status_code == 200
            data = response.json()
            assert data["workflow_id"] == workflow_id

    @pytest.mark.asyncio
    async def test_get_workflow_status_not_found(self):
        """Test getting status for non-existent workflow."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            response = client.get("/api/workflow/non-existent/status")

            assert response.status_code == 404


class TestWorkflowRouterCheckpoint:
    """Test checkpoint approval endpoints."""

    @pytest.mark.asyncio
    async def test_approve_checkpoint(self):
        """Test approving a checkpoint."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.models.workflow import CheckpointData, WorkflowPhase
        from moai_adk.web.routers.workflow import get_orchestrator, router
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        # Create orchestrator with checkpoint
        orchestrator = WorkflowOrchestrator()
        app.dependency_overrides[get_orchestrator] = lambda: orchestrator

        with TestClient(app) as client:
            # Create workflow and add checkpoint
            create_response = client.post(
                "/api/workflow/start",
                json={"features": ["test"]},
            )
            workflow_id = create_response.json()["workflow_id"]

            # Directly add checkpoint to internal dict (synchronous)
            checkpoint = CheckpointData(
                workflow_id=workflow_id,
                phase=WorkflowPhase.PLANNING,
                message="Review SPECs",
                requires_approval=True,
            )
            orchestrator.checkpoint_manager._checkpoints[workflow_id] = checkpoint

            # Approve checkpoint
            response = client.post(f"/api/workflow/{workflow_id}/checkpoint/approve")

            assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_approve_checkpoint_not_found(self):
        """Test approving non-existent checkpoint."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            response = client.post("/api/workflow/non-existent/checkpoint/approve")

            assert response.status_code == 404

    @pytest.mark.asyncio
    async def test_reject_checkpoint(self):
        """Test rejecting a checkpoint."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.models.workflow import CheckpointData, WorkflowPhase
        from moai_adk.web.routers.workflow import get_orchestrator, router
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        orchestrator = WorkflowOrchestrator()
        app.dependency_overrides[get_orchestrator] = lambda: orchestrator

        with TestClient(app) as client:
            create_response = client.post(
                "/api/workflow/start",
                json={"features": ["test"]},
            )
            workflow_id = create_response.json()["workflow_id"]

            # Directly add checkpoint to internal dict (synchronous)
            checkpoint = CheckpointData(
                workflow_id=workflow_id,
                phase=WorkflowPhase.PLANNING,
                message="Review SPECs",
                requires_approval=True,
            )
            orchestrator.checkpoint_manager._checkpoints[workflow_id] = checkpoint

            response = client.post(
                f"/api/workflow/{workflow_id}/checkpoint/reject",
                json={"reason": "Not ready"},
            )

            assert response.status_code == 200


class TestWorkflowRouterCancel:
    """Test DELETE /api/workflow/{id} endpoint."""

    @pytest.mark.asyncio
    async def test_cancel_workflow(self):
        """Test cancelling a workflow."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            # Create workflow first
            create_response = client.post(
                "/api/workflow/start",
                json={"features": ["test"]},
            )
            workflow_id = create_response.json()["workflow_id"]

            # Cancel it
            response = client.delete(f"/api/workflow/{workflow_id}")

            assert response.status_code == 200
            data = response.json()
            assert data["cancelled"] is True

    @pytest.mark.asyncio
    async def test_cancel_workflow_not_found(self):
        """Test cancelling non-existent workflow."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            response = client.delete("/api/workflow/non-existent")

            assert response.status_code == 404


class TestWorkflowRouterList:
    """Test GET /api/workflow/list endpoint."""

    @pytest.mark.asyncio
    async def test_list_workflows(self):
        """Test listing active workflows."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            # Create a few workflows
            client.post("/api/workflow/start", json={"features": ["feat1"]})
            client.post("/api/workflow/start", json={"features": ["feat2"]})

            # List them
            response = client.get("/api/workflow/list")

            assert response.status_code == 200
            data = response.json()
            assert "workflows" in data
            assert len(data["workflows"]) >= 2


class TestWorkflowRouterAdvance:
    """Test POST /api/workflow/{id}/advance endpoint."""

    @pytest.mark.asyncio
    async def test_advance_workflow_phase(self):
        """Test advancing workflow to next phase."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            # Create workflow
            create_response = client.post(
                "/api/workflow/start",
                json={"features": ["test"]},
            )
            workflow_id = create_response.json()["workflow_id"]

            # Advance phase
            response = client.post(f"/api/workflow/{workflow_id}/advance")

            assert response.status_code == 200
            data = response.json()
            assert data["phase"] == "planning"  # Next phase after configuration

    @pytest.mark.asyncio
    async def test_advance_workflow_not_found(self):
        """Test advancing non-existent workflow."""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        from moai_adk.web.routers.workflow import router

        app = FastAPI()
        app.include_router(router, prefix="/api/workflow")

        with TestClient(app) as client:
            response = client.post("/api/workflow/non-existent/advance")

            assert response.status_code == 404
