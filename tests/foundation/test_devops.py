"""
Comprehensive TDD tests for devops.py module.
Tests cover all 7 classes.
"""

import pytest
from datetime import datetime, UTC
from moai_adk.foundation.devops import (
    CICDPipelineOrchestrator,
    InfrastructureManager,
    ContainerOrchestrator,
    MonitoringArchitect,
    DeploymentStrategist,
    SecurityHardener,
    DevOpsMetricsCollector,
    CICDWorkflowConfig,
    DeploymentConfig,
    SecurityConfig,
)


# ============================================================================
# Test CICDPipelineOrchestrator
# ============================================================================


class TestCICDPipelineOrchestrator:
    """Test suite for CICDPipelineOrchestrator class."""

    def test_initialization(self):
        """Test orchestrator initialization."""
        orchestrator = CICDPipelineOrchestrator()
        assert orchestrator.pipelines == {}
        assert "GitHub Actions" in orchestrator.SUPPORTED_PLATFORMS

    def test_orchestrate_github_actions_basic(self):
        """Test basic GitHub Actions orchestration."""
        orchestrator = CICDPipelineOrchestrator()
        config = orchestrator.orchestrate_github_actions(
            repo_name="test-repo",
            stages=["build", "test", "deploy"]
        )

        assert config["platform"] == "GitHub Actions"
        assert config["repo_name"] == "test-repo"
        assert len(config["stages"]) == 3
        assert "yaml_content" in config

    def test_orchestrate_gitlab_ci_with_docker(self):
        """Test GitLab CI orchestration with Docker."""
        orchestrator = CICDPipelineOrchestrator()
        config = orchestrator.orchestrate_gitlab_ci(
            project_name="test-project",
            docker_image="python:3.13",
            stages=["build", "test", "deploy"]
        )

        assert config["platform"] == "GitLab CI"
        assert config["docker_image"] == "python:3.13"
        assert len(config["stages"]) == 3

    def test_orchestrate_jenkins_pipeline(self):
        """Test Jenkins pipeline orchestration."""
        orchestrator = CICDPipelineOrchestrator()
        config = orchestrator.orchestrate_jenkins_pipeline(
            job_name="test-job",
            stages=["build", "test", "deploy"]
        )

        assert config["platform"] == "Jenkins"
        assert config["job_name"] == "test-job"
        assert "jenkinsfile_content" in config

    def test_add_stage_to_pipeline(self):
        """Test adding stage to pipeline."""
        orchestrator = CICDPipelineOrchestrator()
        pipeline = orchestrator.orchestrate_github_actions("test-repo", ["build"])
        updated = orchestrator.add_stage(pipeline, "test")

        assert "test" in [s["name"] for s in updated["stages"]]

    def test_validate_pipeline_valid(self):
        """Test validation of valid pipeline."""
        orchestrator = CICDPipelineOrchestrator()
        pipeline = orchestrator.orchestrate_github_actions("test-repo", ["build", "test"])

        result = orchestrator.validate_pipeline(pipeline)

        assert result["valid"] is True
        assert len(result["errors"]) == 0


# ============================================================================
# Test InfrastructureManager
# ============================================================================


class TestInfrastructureManager:
    """Test suite for InfrastructureManager class."""

    def test_initialization(self):
        """Test manager initialization."""
        manager = InfrastructureManager()
        assert manager.infrastructure == {}

    def test_terraform_basic_configuration(self):
        """Test basic Terraform configuration."""
        manager = InfrastructureManager()
        config = manager.terraform_basic_configuration(
            provider="aws",
            region="us-east-1"
        )

        assert config["provider"] == "aws"
        assert config["region"] == "us-east-1"
        assert "terraform_content" in config

    def test_configure_vpc_aws(self):
        """Test AWS VPC configuration."""
        manager = InfrastructureManager()
        vpc = manager.configure_vpc(
            provider="aws",
            cidr_block="10.0.0.0/16",
            subnet_count=3
        )

        assert vpc["provider"] == "aws"
        assert vpc["cidr_block"] == "10.0.0.0/16"
        assert vpc["subnet_count"] == 3

    def test_configure_load_balancer(self):
        """Test load balancer configuration."""
        manager = InfrastructureManager()
        lb = manager.configure_load_balancer(
            provider="aws",
            lb_type="application",
            target_groups=["web", "api"]
        )

        assert lb["provider"] == "aws"
        assert lb["type"] == "application"
        assert len(lb["target_groups"]) == 2


# ============================================================================
# Test ContainerOrchestrator
# ============================================================================


class TestContainerOrchestrator:
    """Test suite for ContainerOrchestrator class."""

    def test_initialization(self):
        """Test orchestrator initialization."""
        orchestrator = ContainerOrchestrator()
        assert orchestrator.containers == {}
        assert "Kubernetes" in orchestrator.SUPPORTED_PLATFORMS

    def test_deploy_kubernetes_deployment(self):
        """Test Kubernetes deployment configuration."""
        orchestrator = ContainerOrchestrator()
        config = orchestrator.deploy_kubernetes_deployment(
            name="test-app",
            image="test-image:latest",
            replicas=3
        )

        assert config["platform"] == "Kubernetes"
        assert config["name"] == "test-app"
        assert config["replicas"] == 3
        assert "yaml_content" in config

    def test_deploy_kubernetes_service(self):
        """Test Kubernetes service configuration."""
        orchestrator = ContainerOrchestrator()
        config = orchestrator.deploy_kubernetes_service(
            name="test-service",
            selector={"app": "test"},
            port=80
        )

        assert config["name"] == "test-service"
        assert config["port"] == 80

    def test_deploy_docker_compose(self):
        """Test Docker Compose deployment."""
        orchestrator = ContainerOrchestrator()
        services = {
            "web": {"image": "nginx", "ports": ["80:80"]},
            "api": {"image": "api", "ports": ["8000:8000"]}
        }

        config = orchestrator.deploy_docker_compose(services)

        assert "yaml_content" in config
        assert len(config["services"]) == 2

    def test_configure_helm_chart(self):
        """Test Helm chart configuration."""
        orchestrator = ContainerOrchestrator()
        chart = orchestrator.configure_helm_chart(
            name="test-chart",
            version="1.0.0",
            values={"replicas": 3}
        )

        assert chart["name"] == "test-chart"
        assert chart["version"] == "1.0.0"


# ============================================================================
# Test MonitoringArchitect
# ============================================================================


class TestMonitoringArchitect:
    """Test suite for MonitoringArchitect class."""

    def test_initialization(self):
        """Test architect initialization."""
        architect = MonitoringArchitect()
        assert architect.monitors == {}

    def test_setup_prometheus_metrics(self):
        """Test Prometheus metrics setup."""
        architect = MonitoringArchitect()
        config = architect.setup_prometheus_metrics(
            service_name="test-service",
            port=9090,
            metrics=["http_requests", "response_time"]
        )

        assert config["monitoring_system"] == "Prometheus"
        assert config["service_name"] == "test-service"
        assert config["port"] == 9090

    def test_setup_grafana_dashboard(self):
        """Test Grafana dashboard setup."""
        architect = MonitoringArchitect()
        dashboard = architect.setup_grafana_dashboard(
            dashboard_name="test-dashboard",
            panels=[{"title": "Request Rate", "query": "rate(http_requests[5m])"}]
        )

        assert dashboard["dashboard_name"] == "test-dashboard"
        assert len(dashboard["panels"]) == 1

    def test_configure_alerts(self):
        """Test alert configuration."""
        architect = MonitoringArchitect()
        alerts = architect.configure_alerts(
            alert_name="high_error_rate",
            condition="error_rate > 0.05",
            duration="5m"
        )

        assert alerts["alert_name"] == "high_error_rate"
        assert alerts["condition"] == "error_rate > 0.05"
        assert alerts["duration"] == "5m"


# ============================================================================
# Test DeploymentStrategist
# ============================================================================


class TestDeploymentStrategist:
    """Test suite for DeploymentStrategist class."""

    def test_initialization(self):
        """Test strategist initialization."""
        strategist = DeploymentStrategist()
        assert "blue_green" in strategist.DEPLOYMENT_STRATEGIES

    def test_blue_green_deployment(self):
        """Test blue-green deployment configuration."""
        strategist = DeploymentStrategist()
        config = strategist.blue_green_deployment(
            service_name="test-service",
            blue_version="v1.0",
            green_version="v2.0"
        )

        assert config["strategy"] == "blue_green"
        assert config["blue_version"] == "v1.0"
        assert config["green_version"] == "v2.0"

    def test_canary_deployment(self):
        """Test canary deployment configuration."""
        strategist = DeploymentStrategist()
        config = strategist.canary_deployment(
            service_name="test-service",
            stable_version="v1.0",
            canary_version="v2.0",
            canary_percentage=10
        )

        assert config["strategy"] == "canary"
        assert config["canary_percentage"] == 10

    def test_rolling_deployment(self):
        """Test rolling deployment configuration."""
        strategist = DeploymentStrategist()
        config = strategist.rolling_deployment(
            service_name="test-service",
            new_version="v2.0",
            batch_size=2,
            pause_seconds=30
        )

        assert config["strategy"] == "rolling"
        assert config["batch_size"] == 2
        assert config["pause_seconds"] == 30


# ============================================================================
# Test SecurityHardener
# ============================================================================


class TestSecurityHardener:
    """Test suite for SecurityHardener class."""

    def test_initialization(self):
        """Test hardener initialization."""
        hardener = SecurityHardener()
        assert hardener.security_policies == {}

    def test_configure_secrets_management(self):
        """Test secrets management configuration."""
        hardener = SecurityHardener()
        config = hardener.configure_secrets_management(
            provider="aws",
            secret_name="test-secret"
        )

        assert config["provider"] == "aws"
        assert config["secret_name"] == "test-secret"

    def test_enable_network_policies(self):
        """Test network policies enablement."""
        hardener = SecurityHardener()
        policies = hardener.enable_network_policies(
            platform="kubernetes",
            default_action="deny"
        )

        assert policies["platform"] == "kubernetes"
        assert policies["default_action"] == "deny"

    def test_scan_vulnerabilities(self):
        """Test vulnerability scanning."""
        hardener = SecurityHardener()
        report = hardener.scan_vulnerabilities(
            image="test-image:latest",
            severity_threshold="high"
        )

        assert "scan_results" in report
        assert "vulnerabilities" in report

    def test_enforce_pod_security(self):
        """Test pod security enforcement."""
        hardener = SecurityHardener()
        policy = hardener.enforce_pod_security(
            policy_level="restricted",
            run_as_non_root=True,
            drop_capabilities=["ALL"]
        )

        assert policy["policy_level"] == "restricted"
        assert policy["run_as_non_root"] is True


# ============================================================================
# Test DevOpsMetricsCollector
# ============================================================================


class TestDevOpsMetricsCollector:
    """Test suite for DevOpsMetricsCollector class."""

    def test_initialization(self):
        """Test collector initialization."""
        collector = DevOpsMetricsCollector()
        assert collector.metrics == {}

    def test_record_deployment_frequency(self):
        """Test deployment frequency recording."""
        collector = DevOpsMetricsCollector()
        collector.record_deployment(
            service="test-service",
            version="v1.0",
            environment="production"
        )

        deployments = collector.get_deployment_frequency("test-service")

        assert deployments["count"] == 1

    def test_calculate_lead_time(self):
        """Test lead time calculation."""
        collector = DevOpsMetricsCollector()
        lead_time = collector.calculate_lead_time(
            service="test-service",
            commit_date="2025-01-01",
            deploy_date="2025-01-03"
        )

        assert lead_time["days"] == 2

    def test_calculate_change_failure_rate(self):
        """Test change failure rate calculation."""
        collector = DevOpsMetricsCollector()
        collector.record_deployment("test-service", "v1.0", "production", success=True)
        collector.record_deployment("test-service", "v1.1", "production", success=False)

        metrics = collector.get_dora_metrics("test-service")

        assert metrics["change_failure_rate"] == 0.5

    def test_get_dora_metrics_summary(self):
        """Test DORA metrics summary."""
        collector = DevOpsMetricsCollector()
        collector.record_deployment("test-service", "v1.0", "production")
        collector.record_deployment("test-service", "v1.1", "production")

        summary = collector.get_dora_metrics_summary()

        assert "deployment_frequency" in summary
        assert "lead_time" in summary
        assert "change_failure_rate" in summary


# ============================================================================
# Test Data Classes
# ============================================================================


class TestDataClasses:
    """Test suite for DevOps data classes."""

    def test_pipeline_stage_creation(self):
        """Test CICDWorkflowConfig dataclass creation."""
        stage = CICDWorkflowConfig(
            name="test",
            command="pytest",
            timeout=300
        )

        assert stage.name == "test"
        assert stage.timeout == 300

    def test_deployment_config_creation(self):
        """Test DeploymentConfig dataclass creation."""
        config = DeploymentConfig(
            service="test-service",
            version="v1.0",
            replicas=3
        )

        assert config.service == "test-service"
        assert config.replicas == 3

    def test_security_report_creation(self):
        """Test SecurityConfig dataclass creation."""
        report = SecurityConfig(
            scan_date="2025-01-13",
            critical_count=0,
            high_count=1,
            medium_count=3
        )

        assert report.critical_count == 0
        assert report.high_count == 1
