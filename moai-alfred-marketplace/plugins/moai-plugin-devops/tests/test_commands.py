"""
DevOps Plugin Command Tests

@TEST:DEVOPS-PLUGIN-001 - DevOps Plugin Command Execution
Tests for `/setup-docker`, `/setup-ci`, `/setup-k8s` command functionality

@CODE:DEVOPS-TESTS-SUITE-001:TEST
"""

import pytest
import tempfile
import yaml
from pathlib import Path


class TestSetupDockerCommand:
    """Test cases for /setup-docker command"""

    @pytest.fixture
    def temp_project_dir(self) -> Path:
        """Create temporary project directory"""
        with tempfile.TemporaryDirectory() as tmpdir:
            yield Path(tmpdir)

    def test_setup_docker_basic(self, temp_project_dir):
        """
        GIVEN: User invokes /setup-docker
        WHEN: Project type is valid
        THEN: Docker configuration created
        """
        # @CODE:DEVOPS-DOCKER-BASIC-001:TEST
        from devops_plugin.commands import setup_docker

        result = setup_docker.execute(
            project_name="app",
            output_dir=temp_project_dir,
            app_type="python"
        )

        assert result.success is True
        assert (temp_project_dir / "app" / "Dockerfile").exists()

    def test_setup_docker_with_compose(self, temp_project_dir):
        """
        GIVEN: /setup-docker with --compose option
        WHEN: Docker Compose enabled
        THEN: docker-compose.yml created
        """
        # @CODE:DEVOPS-DOCKER-COMPOSE-001:TEST
        from devops_plugin.commands import setup_docker

        result = setup_docker.execute(
            project_name="app",
            output_dir=temp_project_dir,
            app_type="python",
            include_compose=True
        )

        assert result.success is True
        assert (temp_project_dir / "app" / "docker-compose.yml").exists()

    def test_setup_docker_invalid_type(self, temp_project_dir):
        """
        GIVEN: /setup-docker with invalid app type
        WHEN: App type not supported
        THEN: Raises ValueError
        """
        # @CODE:DEVOPS-DOCKER-INVALID-001:TEST
        from devops_plugin.commands import setup_docker

        with pytest.raises(ValueError):
            setup_docker.execute(
                project_name="app",
                output_dir=temp_project_dir,
                app_type="unsupported"
            )


class TestSetupCICommand:
    """Test cases for /setup-ci command"""

    def test_setup_ci_github_actions(self, tmp_path):
        """
        GIVEN: User invokes /setup-ci with github-actions
        WHEN: GitHub Actions selected
        THEN: Workflow files created
        """
        # @CODE:DEVOPS-CI-GITHUB-001:TEST
        from devops_plugin.commands import setup_ci

        result = setup_ci.execute(
            project_name="app",
            output_dir=tmp_path,
            ci_platform="github-actions"
        )

        assert result.success is True
        assert (tmp_path / "app" / ".github").exists()

    def test_setup_ci_gitlab(self, tmp_path):
        """
        GIVEN: /setup-ci with gitlab-ci
        WHEN: GitLab CI selected
        THEN: .gitlab-ci.yml created
        """
        # @CODE:DEVOPS-CI-GITLAB-001:TEST
        from devops_plugin.commands import setup_ci

        result = setup_ci.execute(
            project_name="app",
            output_dir=tmp_path,
            ci_platform="gitlab-ci"
        )

        assert result.success is True

    def test_setup_ci_with_tests(self, tmp_path):
        """
        GIVEN: /setup-ci with --tests option
        WHEN: Test execution enabled
        THEN: Test stage included in CI
        """
        # @CODE:DEVOPS-CI-TESTS-001:TEST
        from devops_plugin.commands import setup_ci

        result = setup_ci.execute(
            project_name="app",
            output_dir=tmp_path,
            ci_platform="github-actions",
            include_tests=True
        )

        assert result.success is True

    def test_setup_ci_invalid_platform(self, tmp_path):
        """
        GIVEN: /setup-ci with invalid platform
        WHEN: Platform not supported
        THEN: Raises ValueError
        """
        # @CODE:DEVOPS-CI-INVALID-001:TEST
        from devops_plugin.commands import setup_ci

        with pytest.raises(ValueError):
            setup_ci.execute(
                project_name="app",
                output_dir=tmp_path,
                ci_platform="unsupported"
            )


class TestSetupK8sCommand:
    """Test cases for /setup-k8s command"""

    def test_setup_k8s_basic(self, tmp_path):
        """
        GIVEN: User invokes /setup-k8s
        WHEN: Kubernetes setup requested
        THEN: K8s manifests created
        """
        # @CODE:DEVOPS-K8S-BASIC-001:TEST
        from devops_plugin.commands import setup_k8s

        result = setup_k8s.execute(
            project_name="app",
            output_dir=tmp_path,
            replicas=3
        )

        assert result.success is True
        assert (tmp_path / "app" / "k8s").exists()

    def test_setup_k8s_with_ingress(self, tmp_path):
        """
        GIVEN: /setup-k8s with --ingress option
        WHEN: Ingress enabled
        THEN: Ingress manifest created
        """
        # @CODE:DEVOPS-K8S-INGRESS-001:TEST
        from devops_plugin.commands import setup_k8s

        result = setup_k8s.execute(
            project_name="app",
            output_dir=tmp_path,
            replicas=3,
            include_ingress=True
        )

        assert result.success is True
        k8s_dir = tmp_path / "app" / "k8s"
        files = list(k8s_dir.glob("*.yaml")) if k8s_dir.exists() else []
        assert len(files) >= 1

    def test_setup_k8s_invalid_replicas(self, tmp_path):
        """
        GIVEN: /setup-k8s with invalid replica count
        WHEN: Replicas < 1
        THEN: Raises ValueError
        """
        # @CODE:DEVOPS-K8S-INVALID-001:TEST
        from devops_plugin.commands import setup_k8s

        with pytest.raises(ValueError):
            setup_k8s.execute(
                project_name="app",
                output_dir=tmp_path,
                replicas=0
            )


class TestDevOpsPluginIntegration:
    """Integration tests for DevOps Plugin"""

    def test_devops_complete_workflow(self, tmp_path):
        """
        GIVEN: Complete DevOps setup workflow
        WHEN: docker → ci → k8s
        THEN: All steps complete successfully
        """
        from devops_plugin.commands import setup_docker, setup_ci, setup_k8s

        # Step 1: Setup Docker
        result1 = setup_docker.execute(
            project_name="app",
            output_dir=tmp_path,
            app_type="python",
            include_compose=True
        )
        assert result1.success is True

        # Step 2: Setup CI
        result2 = setup_ci.execute(
            project_name="app",
            output_dir=tmp_path,
            ci_platform="github-actions",
            include_tests=True
        )
        assert result2.success is True

        # Step 3: Setup K8s
        result3 = setup_k8s.execute(
            project_name="app",
            output_dir=tmp_path,
            replicas=3,
            include_ingress=True
        )
        assert result3.success is True


class TestDevOpsPluginPerformance:
    """Performance tests for DevOps Plugin"""

    def test_setup_docker_completes_quickly(self, tmp_path):
        """
        GIVEN: /setup-docker command
        WHEN: Docker setup execution
        THEN: Completes within 5 seconds
        """
        import time
        from devops_plugin.commands import setup_docker

        start = time.time()
        setup_docker.execute(
            project_name="app",
            output_dir=tmp_path,
            app_type="python"
        )
        elapsed = time.time() - start

        assert elapsed < 5.0, f"Command took {elapsed}s, expected < 5s"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
