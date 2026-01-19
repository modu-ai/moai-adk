"""
Comprehensive DDD tests for devops.py module.
Tests cover all 7 classes with corrected method signatures.
"""

from moai_adk.foundation.devops import (
    CICDPipelineOrchestrator,
    CICDWorkflowConfig,
    ContainerConfig,
    ContainerOrchestrator,
    DeploymentConfig,
    DeploymentStrategist,
    DevOpsMetrics,
    DevOpsMetricsCollector,
    InfrastructureConfig,
    InfrastructureManager,
    MonitoringArchitect,
    MonitoringConfig,
    SecurityConfig,
    SecurityHardener,
)

# ============================================================================
# Test CICDPipelineOrchestrator
# ============================================================================


class TestCICDPipelineOrchestrator:
    """Test suite for CICDPipelineOrchestrator class."""

    def test_initialization(self):
        """Test orchestrator initialization."""
        orchestrator = CICDPipelineOrchestrator()
        assert orchestrator.supported_platforms == ["github", "gitlab", "jenkins", "azure-pipelines"]
        assert orchestrator.default_environments == ["dev", "staging", "prod"]

    def test_orchestrate_github_actions_basic(self):
        """Test basic GitHub Actions orchestration."""
        orchestrator = CICDPipelineOrchestrator()
        config = orchestrator.orchestrate_github_actions(
            config={"name": "test-app", "runtime": "python", "framework": "fastapi"}
        )

        assert config["name"] == "test-app"
        assert "jobs" in config
        assert "test" in config["jobs"]
        assert "build" in config["jobs"]
        assert "deploy" in config["jobs"]

    def test_orchestrate_gitlab_ci_with_docker(self):
        """Test GitLab CI orchestration with Docker."""
        orchestrator = CICDPipelineOrchestrator()
        config = orchestrator.orchestrate_gitlab_ci(
            config={"stages": ["build", "test", "deploy"], "docker_image": "python:3.13"}
        )

        assert config["stages"] == ["build", "test", "deploy"]
        assert config["image"] == "python:3.13"
        assert "build" in config
        assert "test" in config

    def test_orchestrate_jenkins_pipeline(self):
        """Test Jenkins pipeline orchestration."""
        orchestrator = CICDPipelineOrchestrator()
        config = orchestrator.orchestrate_jenkins(config={"stages": ["Build", "Test", "Deploy"]})

        assert "pipeline" in config
        assert config["pipeline"]["agent"] == "any"
        assert len(config["pipeline"]["stages"]) == 3

    def test_optimize_build_pipeline(self):
        """Test build pipeline optimization."""
        orchestrator = CICDPipelineOrchestrator()
        config = orchestrator.optimize_build_pipeline(config={"base_image": "python:3.11-slim"})

        assert config["multi_stage_build"] is True
        assert config["layer_caching"] is True
        assert config["base_image"] == "python:3.11-slim"


# ============================================================================
# Test InfrastructureManager
# ============================================================================


class TestInfrastructureManager:
    """Test suite for InfrastructureManager class."""

    def test_initialization(self):
        """Test infrastructure manager initialization."""
        manager = InfrastructureManager()
        assert manager.supported_providers == ["aws", "gcp", "azure", "kubernetes"]
        assert manager.default_region == "us-west-2"

    def test_generate_kubernetes_manifests(self):
        """Test Kubernetes manifests generation."""
        manager = InfrastructureManager()
        manifests = manager.generate_kubernetes_manifests(
            app_config={"name": "test-app", "namespace": "production", "replicas": 5}
        )

        assert "deployment" in manifests
        assert "service" in manifests
        assert "configmap" in manifests
        assert manifests["deployment"]["spec"]["replicas"] == 5

    def test_create_helm_charts(self):
        """Test Helm chart creation."""
        manager = InfrastructureManager()
        charts = manager.create_helm_charts(chart_config={"name": "test-chart", "version": "1.0.0"})

        assert "Chart.yaml" in charts
        assert "values.yaml" in charts
        assert "templates" in charts
        assert charts["Chart.yaml"]["name"] == "test-chart"

    def test_design_terraform_modules(self):
        """Test Terraform module design."""
        manager = InfrastructureManager()
        modules = manager.design_terraform_modules(
            module_config={"provider": "aws", "region": "us-east-1", "resources": {"vpc": {"cidr": "10.0.0.0/16"}}}
        )

        assert "provider" in modules
        assert "variables" in modules
        assert "outputs" in modules
        assert modules["provider"]["region"] == "us-east-1"

    def test_validate_infrastructure(self):
        """Test infrastructure validation."""
        manager = InfrastructureManager()
        result = manager.validate_infrastructure()

        assert "compliance_score" in result
        assert "validations" in result
        assert result["compliance_score"] >= 0


# ============================================================================
# Test ContainerOrchestrator
# ============================================================================


class TestContainerOrchestrator:
    """Test suite for ContainerOrchestrator class."""

    def test_initialization(self):
        """Test container orchestrator initialization."""
        orchestrator = ContainerOrchestrator()
        assert "docker" in orchestrator.supported_runtimes
        assert "python" in orchestrator.default_base_images

    def test_optimize_dockerfile(self):
        """Test Dockerfile optimization."""
        orchestrator = ContainerOrchestrator()
        config = orchestrator.optimize_dockerfile(
            dockerfile_config={"base_image": "python:3.11-slim", "workdir": "/app"}
        )

        assert config["multi_stage"] is True
        assert config["security_features"]["non_root_user"] is True
        assert config["base_image"] == "python:3.11-slim"

    def test_scan_container_security(self):
        """Test container security scanning."""
        orchestrator = ContainerOrchestrator()
        results = orchestrator.scan_container_security(
            image_name="nginx:latest", security_config={"scan_level": "standard"}
        )

        assert "vulnerabilities" in results
        assert "security_score" in results
        assert results["scan_metadata"]["image_name"] == "nginx:latest"

    def test_plan_kubernetes_deployment(self):
        """Test Kubernetes deployment planning."""
        orchestrator = ContainerOrchestrator()
        plan = orchestrator.plan_kubernetes_deployment(deployment_config={"app_name": "test-app", "replicas": 3})

        assert "deployment_yaml" in plan
        assert "service_yaml" in plan
        assert "ingress_yaml" in plan
        assert plan["deployment_yaml"]["spec"]["replicas"] == 3

    def test_configure_service_mesh(self):
        """Test service mesh configuration."""
        orchestrator = ContainerOrchestrator()
        config = orchestrator.configure_service_mesh()

        assert "istio" in config
        assert "cilium" in config
        assert config["istio"]["enabled"] is True


# ============================================================================
# Test MonitoringArchitect
# ============================================================================


class TestMonitoringArchitect:
    """Test suite for MonitoringArchitect class."""

    def test_initialization(self):
        """Test monitoring architect initialization."""
        architect = MonitoringArchitect()
        assert architect.default_scrape_interval == "15s"
        assert architect.default_evaluation_interval == "15s"

    def test_setup_prometheus(self):
        """Test Prometheus setup."""
        architect = MonitoringArchitect()
        config = architect.setup_prometheus(metrics_config={"app_name": "test-app", "scrape_interval": "30s"})

        assert "prometheus_config" in config
        assert "scrape_interval" in config
        assert config["scrape_interval"] == "30s"

    def test_design_grafana_dashboards(self):
        """Test Grafana dashboard design."""
        architect = MonitoringArchitect()
        config = architect.design_grafana_dashboards(
            dashboard_config={
                "dashboard_name": "Test Dashboard",
                "panels": [{"title": "CPU Usage", "metric": "cpu_usage"}],
            }
        )

        assert "dashboard_json" in config
        assert config["dashboard_json"]["title"] == "Test Dashboard"
        assert len(config["dashboard_json"]["panels"]) == 1

    def test_configure_logging(self):
        """Test ELK logging configuration."""
        architect = MonitoringArchitect()
        config = architect.configure_logging(logging_config={"app_name": "test-app", "environment": "production"})

        assert "elasticsearch_config" in config
        assert "logstash_config" in config
        assert "filebeat_config" in config

    def test_setup_alerting(self):
        """Test alerting setup."""
        architect = MonitoringArchitect()
        config = architect.setup_alerting()

        assert "alertmanager" in config
        assert "alert_rules" in config
        assert len(config["alert_rules"]) > 0


# ============================================================================
# Test DeploymentStrategist
# ============================================================================


class TestDeploymentStrategist:
    """Test suite for DeploymentStrategist class."""

    def test_initialization(self):
        """Test deployment strategist initialization."""
        strategist = DeploymentStrategist()
        assert "blue_green" in strategist.supported_strategies
        assert "canary" in strategist.supported_strategies

    def test_plan_continuous_deployment(self):
        """Test continuous deployment planning."""
        strategist = DeploymentStrategist()
        plan = strategist.plan_continuous_deployment(
            cd_config={"environments": ["staging", "production"], "gates": ["tests", "security"]}
        )

        assert "pipeline_stages" in plan
        assert "quality_gates" in plan
        assert len(plan["pipeline_stages"]) == 2

    def test_design_canary_deployment(self):
        """Test canary deployment design."""
        strategist = DeploymentStrategist()
        config = strategist.design_canary_deployment(
            canary_config={"canary_percentage": 20, "monitoring_duration": "15m"}
        )

        assert "canary_config" in config
        assert "traffic_splitting" in config
        assert config["canary_config"]["initial_percentage"] == 20

    def test_implement_blue_green_deployment(self):
        """Test blue-green deployment implementation."""
        strategist = DeploymentStrategist()
        config = strategist.implement_blue_green_deployment(
            bg_config={"blue_environment": "prod-blue", "green_environment": "prod-green"}
        )

        assert "environment_config" in config
        assert "traffic_switch" in config
        assert config["environment_config"]["blue"]["active"] is True

    def test_integrate_automated_testing(self):
        """Test automated testing integration."""
        strategist = DeploymentStrategist()
        config = strategist.integrate_automated_testing(
            testing_config={"test_types": ["unit", "integration"], "parallel_execution": True}
        )

        assert "test_matrix" in config
        assert "coverage_requirements" in config
        assert len(config["test_matrix"]["tests"]) == 2


# ============================================================================
# Test SecurityHardener
# ============================================================================


class TestSecurityHardener:
    """Test suite for SecurityHardener class."""

    def test_initialization(self):
        """Test security hardener initialization."""
        hardener = SecurityHardener()
        assert "cis_aws" in hardener.supported_standards
        assert "pci_dss" in hardener.supported_standards

    def test_scan_docker_images(self):
        """Test Docker image security scanning."""
        hardener = SecurityHardener()
        results = hardener.scan_docker_images(image_name="nginx:latest")

        assert "vulnerabilities" in results
        assert "security_score" in results
        assert results["image_name"] == "nginx:latest"

    def test_configure_secrets_management(self):
        """Test secrets management configuration."""
        hardener = SecurityHardener()
        config = hardener.configure_secrets_management()

        assert "vault" in config
        assert "kubernetes_secrets" in config
        assert "rotation_policy" in config

    def test_setup_network_policies(self):
        """Test network policies setup."""
        hardener = SecurityHardener()
        policies = hardener.setup_network_policies()

        assert "default_deny" in policies
        assert "allow_same_namespace" in policies
        assert "allow_dns" in policies

    def test_audit_compliance(self):
        """Test compliance audit."""
        hardener = SecurityHardener()
        report = hardener.audit_compliance()

        assert "audit_timestamp" in report
        assert "compliance_standards" in report
        assert "overall_score" in report
        assert report["overall_score"] >= 0


# ============================================================================
# Test DevOpsMetricsCollector
# ============================================================================


class TestDevOpsMetricsCollector:
    """Test suite for DevOpsMetricsCollector class."""

    def test_initialization(self):
        """Test metrics collector initialization."""
        collector = DevOpsMetricsCollector()
        assert collector.metrics_window_days == 30

    def test_collect_deployment_metrics(self):
        """Test deployment metrics collection."""
        collector = DevOpsMetricsCollector()
        metrics = collector.collect_deployment_metrics(
            deployment_info={"start_time": "2024-01-01T00:00:00Z", "end_time": "2024-01-01T00:15:00Z"}
        )

        assert "deployment_duration" in metrics
        assert "success_rate" in metrics
        assert "deployment_frequency" in metrics

    def test_track_pipeline_performance(self):
        """Test pipeline performance tracking."""
        collector = DevOpsMetricsCollector()
        metrics = collector.track_pipeline_performance(
            pipeline_data={"execution_times": {"build": 120, "test": 180, "deploy": 60}, "success_rate": 95.5}
        )

        assert "total_execution_time" in metrics
        assert "stage_performance" in metrics
        assert "bottleneck_analysis" in metrics

    def test_monitor_resource_usage(self):
        """Test resource usage monitoring."""
        collector = DevOpsMetricsCollector()
        metrics = collector.monitor_resource_usage(
            resource_config={"monitoring_period": "24h", "metrics": ["cpu", "memory"]}
        )

        assert "cpu_utilization" in metrics
        assert "memory_usage" in metrics
        assert "cost_metrics" in metrics

    def test_get_devops_health_status(self):
        """Test DevOps health status retrieval."""
        collector = DevOpsMetricsCollector()
        status = collector.get_devops_health_status(
            health_config={"health_threshold": {"deployment_success": 95, "uptime": 99.9}}
        )

        assert "overall_health_score" in status
        assert "category_scores" in status
        assert "recommendations" in status
        assert status["overall_health_score"] >= 0


# ============================================================================
# Test Data Classes
# ============================================================================


class TestDataClasses:
    """Test suite for DevOps dataclasses."""

    def test_cicd_workflow_config(self):
        """Test CICDWorkflowConfig dataclass."""
        config = CICDWorkflowConfig(
            name="test-workflow",
            triggers=["push", "pull_request"],
            jobs={"build": {"steps": ["echo 'Building'"]}},
            variables={"ENV": "test"},
        )

        assert config.name == "test-workflow"
        assert len(config.triggers) == 2
        assert config.variables["ENV"] == "test"

    def test_infrastructure_config(self):
        """Test InfrastructureConfig dataclass."""
        config = InfrastructureConfig(
            provider="aws",
            region="us-west-2",
            resources={"vpc": {"cidr": "10.0.0.0/16"}},
            variables={"environment": "production"},
            version="1.0.0",
        )

        assert config.provider == "aws"
        assert config.version == "1.0.0"

    def test_container_config(self):
        """Test ContainerConfig dataclass."""
        config = ContainerConfig(
            image="nginx:latest",
            ports=[80, 443],
            environment={"ENV": "production"},
            resources={"cpu": "500m"},
            security={"non_root": True},
        )

        assert config.image == "nginx:latest"
        assert len(config.ports) == 2

    def test_monitoring_config(self):
        """Test MonitoringConfig dataclass."""
        config = MonitoringConfig(
            scrape_interval="30s",
            targets=[{"job": "app", "target": "localhost:9000"}],
            alert_rules=[{"name": "HighErrorRate", "expr": "rate > 0.1"}],
            dashboards=[{"name": "Main", "panels": []}],
        )

        assert config.scrape_interval == "30s"
        assert len(config.alert_rules) == 1

    def test_security_config(self):
        """Test SecurityConfig dataclass."""
        config = SecurityConfig(
            policies=[{"name": "network-policy", "rules": []}],
            compliance_standards=["cis_aws", "pci_dss"],
            audit_settings={"enabled": True, "frequency": "daily"},
        )

        assert len(config.compliance_standards) == 2
        assert config.audit_settings["enabled"] is True

    def test_deployment_config(self):
        """Test DeploymentConfig dataclass."""
        config = DeploymentConfig(
            strategy="blue_green",
            phases=[{"name": "deploy", "steps": []}],
            rollback_config={"enabled": True, "timeout": "5m"},
            health_checks={"endpoint": "/health", "interval": "30s"},
        )

        assert config.strategy == "blue_green"
        assert config.rollback_config["enabled"] is True

    def test_devops_metrics(self):
        """Test DevOpsMetrics dataclass."""
        metrics = DevOpsMetrics(
            deployment_frequency={"daily": 2.5, "weekly": 17.5},
            lead_time_for_changes={"avg_minutes": 45},
            change_failure_rate={"percentage": 2.5},
            mean_time_to_recovery={"avg_minutes": 15},
        )

        assert metrics.deployment_frequency["daily"] == 2.5
        assert metrics.change_failure_rate["percentage"] == 2.5
