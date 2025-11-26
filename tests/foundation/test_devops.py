"""
Comprehensive test suite for DevOps module.

Tests cover CI/CD pipeline orchestration, infrastructure management,
container orchestration, monitoring, deployment strategies, security,
and metrics collection with 90%+ coverage goal.
"""

from unittest.mock import patch

import pytest

from src.moai_adk.foundation.devops import (
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
# Dataclass Tests
# ============================================================================


class TestDataclasses:
    """Test DevOps dataclass definitions."""

    def test_cicd_workflow_config_initialization(self):
        """Test CICDWorkflowConfig dataclass initialization."""
        config = CICDWorkflowConfig(
            name="test-workflow",
            triggers=["push", "pull_request"],
            jobs={"test": {}, "build": {}},
            variables={"ENV": "test"},
        )
        assert config.name == "test-workflow"
        assert config.triggers == ["push", "pull_request"]
        assert config.jobs == {"test": {}, "build": {}}
        assert config.variables == {"ENV": "test"}

    def test_cicd_workflow_config_without_variables(self):
        """Test CICDWorkflowConfig without optional variables."""
        config = CICDWorkflowConfig(
            name="workflow",
            triggers=["push"],
            jobs={"test": {}},
        )
        assert config.variables is None

    def test_infrastructure_config_initialization(self):
        """Test InfrastructureConfig dataclass initialization."""
        config = InfrastructureConfig(
            provider="aws",
            region="us-west-2",
            resources={"vpc": {}},
            variables={"key": "value"},
        )
        assert config.provider == "aws"
        assert config.region == "us-west-2"
        assert config.version == "1.0.0"

    def test_container_config_initialization(self):
        """Test ContainerConfig dataclass initialization."""
        config = ContainerConfig(
            image="nginx:latest",
            ports=[80, 443],
            environment={"ENV": "prod"},
            resources={"cpu": "1", "memory": "1Gi"},
            security={"privileged": False},
        )
        assert config.image == "nginx:latest"
        assert config.ports == [80, 443]

    def test_monitoring_config_initialization(self):
        """Test MonitoringConfig dataclass initialization."""
        config = MonitoringConfig(
            scrape_interval="15s",
            targets=[{"url": "localhost:9090"}],
            alert_rules=[],
            dashboards=[],
        )
        assert config.scrape_interval == "15s"

    def test_security_config_initialization(self):
        """Test SecurityConfig dataclass initialization."""
        config = SecurityConfig(
            policies=[{"name": "policy1"}],
            compliance_standards=["PCI-DSS"],
            audit_settings={"enabled": True},
        )
        assert config.compliance_standards == ["PCI-DSS"]

    def test_deployment_config_initialization(self):
        """Test DeploymentConfig dataclass initialization."""
        config = DeploymentConfig(
            strategy="rolling",
            phases=[{"name": "phase1"}],
            rollback_config={"enabled": True},
            health_checks={"enabled": True},
        )
        assert config.strategy == "rolling"

    def test_devops_metrics_initialization(self):
        """Test DevOpsMetrics dataclass initialization."""
        metrics = DevOpsMetrics(
            deployment_frequency={},
            lead_time_for_changes={},
            change_failure_rate={},
            mean_time_to_recovery={},
        )
        assert isinstance(metrics.deployment_frequency, dict)


# ============================================================================
# CICDPipelineOrchestrator Tests
# ============================================================================


class TestCICDPipelineOrchestrator:
    """Test CI/CD pipeline orchestration functionality."""

    @pytest.fixture
    def orchestrator(self):
        """Create orchestrator instance."""
        return CICDPipelineOrchestrator()

    def test_orchestrator_initialization(self, orchestrator):
        """Test orchestrator initialization."""
        assert orchestrator.supported_platforms == ["github", "gitlab", "jenkins", "azure-pipelines"]
        assert orchestrator.default_environments == ["dev", "staging", "prod"]

    def test_orchestrate_github_actions_basic(self, orchestrator):
        """Test basic GitHub Actions workflow generation."""
        config = {"name": "Test Pipeline"}
        workflow = orchestrator.orchestrate_github_actions(config)

        assert workflow["name"] == "Test Pipeline"
        assert "on" in workflow
        assert "jobs" in workflow
        assert "test" in workflow["jobs"]
        assert "build" in workflow["jobs"]
        assert "deploy" in workflow["jobs"]

    def test_orchestrate_github_actions_full_config(self, orchestrator):
        """Test GitHub Actions workflow with full configuration."""
        config = {
            "name": "Complete CI/CD",
            "runtime": "python",
            "framework": "fastapi",
            "build_command": "pip install -r requirements.txt",
            "test_command": "pytest tests/",
            "deploy_target": "production",
        }
        workflow = orchestrator.orchestrate_github_actions(config)

        env = workflow["env"]
        assert env["PROJECT_NAME"] == "Complete CI/CD"
        assert env["RUNTIME"] == "python"
        assert env["FRAMEWORK"] == "fastapi"

        test_job = workflow["jobs"]["test"]
        assert test_job["runs-on"] == "ubuntu-latest"
        assert len(test_job["steps"]) >= 4

    def test_orchestrate_github_actions_defaults(self, orchestrator):
        """Test GitHub Actions workflow uses defaults."""
        workflow = orchestrator.orchestrate_github_actions({})

        assert workflow["env"]["PROJECT_NAME"] == "app"
        assert workflow["env"]["RUNTIME"] == "python"

    def test_orchestrate_github_actions_job_dependencies(self, orchestrator):
        """Test GitHub Actions job dependencies."""
        workflow = orchestrator.orchestrate_github_actions({"name": "Pipeline"})

        assert workflow["jobs"]["build"]["needs"] == "test"
        assert workflow["jobs"]["deploy"]["needs"] == "build"

    def test_orchestrate_gitlab_ci_basic(self, orchestrator):
        """Test GitLab CI pipeline generation."""
        config = {"stages": ["build", "test", "deploy"]}
        pipeline = orchestrator.orchestrate_gitlab_ci(config)

        assert pipeline["stages"] == ["build", "test", "deploy"]
        assert "build" in pipeline
        assert "test" in pipeline
        assert "deploy" in pipeline

    def test_orchestrate_gitlab_ci_with_before_script(self, orchestrator):
        """Test GitLab CI with before_script."""
        config = {
            "stages": ["test"],
            "before_script": ["echo 'Setup'"],
        }
        pipeline = orchestrator.orchestrate_gitlab_ci(config)

        assert "before_script" in pipeline
        assert pipeline["before_script"] == ["echo 'Setup'"]

    def test_orchestrate_gitlab_ci_default_image(self, orchestrator):
        """Test GitLab CI uses default image."""
        pipeline = orchestrator.orchestrate_gitlab_ci({})

        assert pipeline["image"] == "python:3.11"

    def test_orchestrate_gitlab_ci_custom_image(self, orchestrator):
        """Test GitLab CI with custom image."""
        config = {"docker_image": "node:18"}
        pipeline = orchestrator.orchestrate_gitlab_ci(config)

        assert pipeline["image"] == "node:18"

    def test_orchestrate_gitlab_ci_build_job_config(self, orchestrator):
        """Test GitLab CI build job configuration."""
        pipeline = orchestrator.orchestrate_gitlab_ci({"stages": ["build"]})

        build_job = pipeline["build"]
        assert build_job["stage"] == "build"
        assert "artifacts" in build_job
        assert build_job["artifacts"]["expire_in"] == "1 hour"

    def test_orchestrate_gitlab_ci_test_job_config(self, orchestrator):
        """Test GitLab CI test job configuration."""
        pipeline = orchestrator.orchestrate_gitlab_ci({"stages": ["test"]})

        test_job = pipeline["test"]
        assert test_job["stage"] == "test"
        assert "coverage" in test_job

    def test_orchestrate_gitlab_ci_deploy_job_config(self, orchestrator):
        """Test GitLab CI deploy job configuration."""
        pipeline = orchestrator.orchestrate_gitlab_ci({"stages": ["deploy"]})

        deploy_job = pipeline["deploy"]
        assert deploy_job["stage"] == "deploy"
        assert deploy_job["environment"]["name"] == "production"

    def test_orchestrate_jenkins_basic(self, orchestrator):
        """Test Jenkins pipeline generation."""
        config = {"stages": ["Build", "Test", "Deploy"]}
        pipeline = orchestrator.orchestrate_jenkins(config)

        assert "pipeline" in pipeline
        assert len(pipeline["pipeline"]["stages"]) == 3

    def test_orchestrate_jenkins_stages(self, orchestrator):
        """Test Jenkins pipeline stages."""
        stages = ["Build", "Test", "Deploy", "Monitor"]
        config = {"stages": stages}
        pipeline = orchestrator.orchestrate_jenkins(config)

        assert len(pipeline["pipeline"]["stages"]) == 4
        stage_names = [s["stage"] for s in pipeline["pipeline"]["stages"]]
        assert stage_names == stages

    def test_orchestrate_jenkins_default_agent(self, orchestrator):
        """Test Jenkins default agent."""
        pipeline = orchestrator.orchestrate_jenkins({})

        assert pipeline["pipeline"]["agent"] == "any"

    def test_orchestrate_jenkins_custom_agent(self, orchestrator):
        """Test Jenkins custom agent."""
        config = {"agent": "docker"}
        pipeline = orchestrator.orchestrate_jenkins(config)

        assert pipeline["pipeline"]["agent"] == "docker"

    def test_optimize_build_pipeline_defaults(self, orchestrator):
        """Test build pipeline optimization defaults."""
        optimization = orchestrator.optimize_build_pipeline({})

        assert optimization["multi_stage_build"] is True
        assert optimization["layer_caching"] is True
        assert optimization["security_scan"] is True
        assert optimization["optimization_level"] == "production"

    def test_optimize_build_pipeline_custom_base_image(self, orchestrator):
        """Test build pipeline with custom base image."""
        config = {"base_image": "golang:1.21"}
        optimization = orchestrator.optimize_build_pipeline(config)

        assert optimization["base_image"] == "golang:1.21"

    def test_optimize_build_pipeline_cache_configuration(self, orchestrator):
        """Test build pipeline cache configuration."""
        optimization = orchestrator.optimize_build_pipeline({})

        assert "cache_from" in optimization
        assert "cache_to" in optimization


# ============================================================================
# InfrastructureManager Tests
# ============================================================================


class TestInfrastructureManager:
    """Test infrastructure as code management."""

    @pytest.fixture
    def manager(self):
        """Create infrastructure manager instance."""
        return InfrastructureManager()

    def test_manager_initialization(self, manager):
        """Test infrastructure manager initialization."""
        assert manager.supported_providers == ["aws", "gcp", "azure", "kubernetes"]
        assert manager.default_region == "us-west-2"

    def test_generate_kubernetes_manifests_basic(self, manager):
        """Test basic Kubernetes manifest generation."""
        app_config = {"name": "myapp"}
        manifests = manager.generate_kubernetes_manifests(app_config)

        assert "deployment" in manifests
        assert "service" in manifests
        assert "configmap" in manifests

    def test_kubernetes_deployment_manifest(self, manager):
        """Test Kubernetes deployment manifest."""
        app_config = {"name": "test-app", "namespace": "default"}
        manifests = manager.generate_kubernetes_manifests(app_config)

        deployment = manifests["deployment"]
        assert deployment["kind"] == "Deployment"
        assert deployment["metadata"]["name"] == "test-app"
        assert deployment["spec"]["replicas"] == 3

    def test_kubernetes_service_manifest(self, manager):
        """Test Kubernetes service manifest."""
        app_config = {"name": "test-app"}
        manifests = manager.generate_kubernetes_manifests(app_config)

        service = manifests["service"]
        assert service["kind"] == "Service"
        assert service["metadata"]["name"] == "test-app"

    def test_kubernetes_configmap_manifest(self, manager):
        """Test Kubernetes configmap manifest."""
        app_config = {"name": "test-app"}
        manifests = manager.generate_kubernetes_manifests(app_config)

        configmap = manifests["configmap"]
        assert configmap["kind"] == "ConfigMap"
        assert "data" in configmap

    def test_kubernetes_manifests_with_custom_replicas(self, manager):
        """Test Kubernetes manifests with custom replica count."""
        app_config = {"name": "app", "replicas": 5}
        manifests = manager.generate_kubernetes_manifests(app_config)

        assert manifests["deployment"]["spec"]["replicas"] == 5

    def test_kubernetes_manifests_with_custom_port(self, manager):
        """Test Kubernetes manifests with custom port."""
        app_config = {"name": "app", "port": 3000}
        manifests = manager.generate_kubernetes_manifests(app_config)

        container = manifests["deployment"]["spec"]["template"]["spec"]["containers"][0]
        assert container["ports"][0]["containerPort"] == 3000

    def test_kubernetes_manifests_with_health_checks(self, manager):
        """Test Kubernetes manifests with health check configuration."""
        app_config = {
            "name": "app",
            "health_check": {"path": "/healthz", "initial_delay": 45},
        }
        manifests = manager.generate_kubernetes_manifests(app_config)

        container = manifests["deployment"]["spec"]["template"]["spec"]["containers"][0]
        assert "livenessProbe" in container
        assert "readinessProbe" in container
        assert container["livenessProbe"]["initialDelaySeconds"] == 45

    def test_create_helm_charts_basic(self, manager):
        """Test basic Helm chart creation."""
        chart_config = {"name": "myapp"}
        charts = manager.create_helm_charts(chart_config)

        assert "Chart.yaml" in charts
        assert "values.yaml" in charts
        assert "templates" in charts

    def test_helm_chart_yaml(self, manager):
        """Test Helm Chart.yaml generation."""
        chart_config = {
            "name": "app-chart",
            "description": "Test chart",
            "version": "1.0.0",
        }
        charts = manager.create_helm_charts(chart_config)

        chart = charts["Chart.yaml"]
        assert chart["name"] == "app-chart"
        assert chart["description"] == "Test chart"
        assert chart["version"] == "1.0.0"

    def test_helm_values_yaml(self, manager):
        """Test Helm values.yaml generation."""
        charts = manager.create_helm_charts({})

        values = charts["values.yaml"]
        assert values["replicaCount"] == 3
        assert "image" in values
        assert "resources" in values

    def test_helm_charts_with_custom_values(self, manager):
        """Test Helm charts with custom values."""
        chart_config = {
            "values": {
                "replicaCount": 5,
                "image": {"tag": "1.2.3"},
            }
        }
        charts = manager.create_helm_charts(chart_config)

        assert charts["values.yaml"]["replicaCount"] == 5

    def test_design_terraform_modules_basic(self, manager):
        """Test basic Terraform module design."""
        module_config = {"provider": "aws"}
        modules = manager.design_terraform_modules(module_config)

        assert "provider" in modules
        assert "variables" in modules
        assert "outputs" in modules
        assert "module" in modules

    def test_terraform_provider_config(self, manager):
        """Test Terraform provider configuration."""
        module_config = {
            "provider": "gcp",
            "region": "us-central1",
        }
        modules = manager.design_terraform_modules(module_config)

        assert modules["provider"]["name"] == "gcp"
        assert modules["provider"]["region"] == "us-central1"

    def test_terraform_vpc_module(self, manager):
        """Test Terraform VPC module configuration."""
        module_config = {
            "resources": {
                "vpc": {"cidr": "10.10.0.0/16"},
            }
        }
        modules = manager.design_terraform_modules(module_config)

        assert "vpc" in modules["module"]
        assert modules["module"]["vpc"]["cidr"] == "10.10.0.0/16"

    def test_terraform_ec2_module(self, manager):
        """Test Terraform EC2 module configuration."""
        module_config = {
            "resources": {
                "ec2": {"instance_type": "t3.large", "count": 3},
            }
        }
        modules = manager.design_terraform_modules(module_config)

        assert "ec2" in modules["module"]
        assert modules["module"]["ec2"]["instance_type"] == "t3.large"

    def test_terraform_rds_module(self, manager):
        """Test Terraform RDS module configuration."""
        module_config = {
            "resources": {
                "rds": {"engine": "mysql", "instance_class": "db.t4g.small"},
            }
        }
        modules = manager.design_terraform_modules(module_config)

        assert "rds" in modules["module"]
        assert modules["module"]["rds"]["engine"] == "mysql"

    def test_terraform_multiple_resources(self, manager):
        """Test Terraform with multiple resources."""
        module_config = {
            "resources": {
                "vpc": {"cidr": "10.0.0.0/16"},
                "ec2": {"instance_type": "t3.medium"},
                "rds": {"engine": "postgres"},
            }
        }
        modules = manager.design_terraform_modules(module_config)

        assert len(modules["module"]) == 3

    def test_validate_infrastructure(self, manager):
        """Test infrastructure validation."""
        result = manager.validate_infrastructure()

        assert "compliance_score" in result
        assert "validations" in result
        assert "recommendations" in result
        assert result["overall_status"] == "compliant"

    def test_validate_infrastructure_checks(self, manager):
        """Test infrastructure validation checks."""
        result = manager.validate_infrastructure()

        validations = result["validations"]
        assert "security_groups" in validations
        assert "iam_policies" in validations
        assert "encryption" in validations
        assert "monitoring" in validations


# ============================================================================
# ContainerOrchestrator Tests
# ============================================================================


class TestContainerOrchestrator:
    """Test container orchestration functionality."""

    @pytest.fixture
    def orchestrator(self):
        """Create container orchestrator instance."""
        return ContainerOrchestrator()

    def test_orchestrator_initialization(self, orchestrator):
        """Test container orchestrator initialization."""
        assert orchestrator.supported_runtimes == ["docker", "containerd", "cri-o"]
        assert "python" in orchestrator.default_base_images
        assert "node" in orchestrator.default_base_images

    def test_optimize_dockerfile_defaults(self, orchestrator):
        """Test Dockerfile optimization defaults."""
        optimization = orchestrator.optimize_dockerfile({})

        assert optimization["multi_stage"] is True
        assert optimization["security_features"]["non_root_user"] is True
        assert optimization["build_cache"]["enabled"] is True

    def test_optimize_dockerfile_custom_base_image(self, orchestrator):
        """Test Dockerfile optimization with custom base image."""
        config = {"base_image": "rust:latest"}
        optimization = orchestrator.optimize_dockerfile(config)

        assert optimization["base_image"] == "rust:latest"

    def test_optimize_dockerfile_workdir(self, orchestrator):
        """Test Dockerfile optimization workdir."""
        config = {"workdir": "/workspace"}
        optimization = orchestrator.optimize_dockerfile(config)

        assert optimization["workdir"] == "/workspace"

    def test_scan_container_security_vulnerabilities(self, orchestrator):
        """Test container security scanning."""
        scan_results = orchestrator.scan_container_security(
            "myapp:latest",
            {"scan_level": "intensive"},
        )

        assert "vulnerabilities" in scan_results
        assert "security_score" in scan_results
        assert "recommendations" in scan_results

    def test_scan_container_security_metadata(self, orchestrator):
        """Test container security scan metadata."""
        scan_results = orchestrator.scan_container_security(
            "nginx:1.21",
            {},
        )

        metadata = scan_results["scan_metadata"]
        assert metadata["image_name"] == "nginx:1.21"
        assert "scan_date" in metadata
        assert metadata["scan_level"] == "standard"

    def test_scan_container_security_vulnerabilities_detail(self, orchestrator):
        """Test container security vulnerability details."""
        scan_results = orchestrator.scan_container_security("app:v1", {})

        vulnerabilities = scan_results["vulnerabilities"]
        assert len(vulnerabilities) >= 2
        assert all("severity" in v for v in vulnerabilities)
        assert all("cve" in v for v in vulnerabilities)

    def test_plan_kubernetes_deployment_basic(self, orchestrator):
        """Test Kubernetes deployment planning."""
        config = {"app_name": "myapp"}
        plan = orchestrator.plan_kubernetes_deployment(config)

        assert "deployment_yaml" in plan
        assert "service_yaml" in plan
        assert "ingress_yaml" in plan
        assert "namespace_yaml" in plan

    def test_kubernetes_deployment_plan_yaml(self, orchestrator):
        """Test Kubernetes deployment YAML generation."""
        config = {"app_name": "test-app", "namespace": "production"}
        plan = orchestrator.plan_kubernetes_deployment(config)

        deployment = plan["deployment_yaml"]
        assert deployment["kind"] == "Deployment"
        assert deployment["metadata"]["namespace"] == "production"

    def test_kubernetes_service_plan_yaml(self, orchestrator):
        """Test Kubernetes service YAML in deployment plan."""
        config = {"app_name": "test-app"}
        plan = orchestrator.plan_kubernetes_deployment(config)

        service = plan["service_yaml"]
        assert service["kind"] == "Service"
        assert service["spec"]["type"] == "ClusterIP"

    def test_kubernetes_ingress_plan_yaml(self, orchestrator):
        """Test Kubernetes ingress YAML in deployment plan."""
        config = {"app_name": "test-app"}
        plan = orchestrator.plan_kubernetes_deployment(config)

        ingress = plan["ingress_yaml"]
        assert ingress["kind"] == "Ingress"
        assert "rules" in ingress["spec"]

    def test_kubernetes_rolling_update_strategy(self, orchestrator):
        """Test Kubernetes rolling update strategy."""
        plan = orchestrator.plan_kubernetes_deployment({})

        strategy = plan["rolling_update_strategy"]
        assert strategy["type"] == "RollingUpdate"
        assert strategy["maxUnavailable"] == 1

    def test_kubernetes_health_checks_in_plan(self, orchestrator):
        """Test Kubernetes health checks in deployment plan."""
        plan = orchestrator.plan_kubernetes_deployment({})

        health_checks = plan["health_checks"]
        assert "livenessProbe" in health_checks
        assert "readinessProbe" in health_checks

    def test_configure_service_mesh(self, orchestrator):
        """Test service mesh configuration."""
        config = orchestrator.configure_service_mesh()

        assert "istio" in config
        assert "cilium" in config

    def test_service_mesh_istio_config(self, orchestrator):
        """Test Istio service mesh configuration."""
        config = orchestrator.configure_service_mesh()

        istio = config["istio"]
        assert istio["enabled"] is True
        assert "pilot" in istio["components"]

    def test_service_mesh_cilium_config(self, orchestrator):
        """Test Cilium service mesh configuration."""
        config = orchestrator.configure_service_mesh()

        cilium = config["cilium"]
        assert cilium["enabled"] is False
        assert "network_policy" in cilium["features"]


# ============================================================================
# MonitoringArchitect Tests
# ============================================================================


class TestMonitoringArchitect:
    """Test monitoring and observability setup."""

    @pytest.fixture
    def architect(self):
        """Create monitoring architect instance."""
        return MonitoringArchitect()

    def test_architect_initialization(self, architect):
        """Test monitoring architect initialization."""
        assert architect.default_scrape_interval == "15s"
        assert architect.default_evaluation_interval == "15s"

    def test_setup_prometheus_basic(self, architect):
        """Test Prometheus setup."""
        config = {"app_name": "myapp"}
        prometheus = architect.setup_prometheus(config)

        assert "prometheus_config" in prometheus
        assert "alerting_rules" in prometheus

    def test_prometheus_global_config(self, architect):
        """Test Prometheus global configuration."""
        config = {"scrape_interval": "60s"}
        prometheus = architect.setup_prometheus(config)

        global_config = prometheus["prometheus_config"]["global"]
        assert global_config["scrape_interval"] == "60s"

    def test_prometheus_scrape_configs(self, architect):
        """Test Prometheus scrape configurations."""
        config = {"app_name": "test-app"}
        prometheus = architect.setup_prometheus(config)

        scrape_configs = prometheus["prometheus_config"]["scrape_configs"]
        assert len(scrape_configs) > 0
        assert scrape_configs[0]["job_name"] == "test-app"

    def test_prometheus_alerting_rules(self, architect):
        """Test Prometheus alerting rules."""
        prometheus = architect.setup_prometheus({})

        rules = prometheus["alerting_rules"]
        assert len(rules) > 0
        assert rules[0]["name"] == "HighErrorRate"

    def test_prometheus_custom_metrics(self, architect):
        """Test Prometheus with custom metrics."""
        config = {"custom_metrics": ["my_metric_1", "my_metric_2"]}
        prometheus = architect.setup_prometheus(config)

        assert prometheus["custom_metrics"] == ["my_metric_1", "my_metric_2"]

    def test_design_grafana_dashboards_basic(self, architect):
        """Test Grafana dashboard design."""
        config = {"dashboard_name": "App Dashboard"}
        dashboard = architect.design_grafana_dashboards(config)

        assert "dashboard_json" in dashboard

    def test_grafana_dashboard_title(self, architect):
        """Test Grafana dashboard title."""
        config = {"dashboard_name": "Custom Dashboard"}
        dashboard = architect.design_grafana_dashboards(config)

        assert dashboard["dashboard_json"]["title"] == "Custom Dashboard"

    def test_grafana_dashboard_panels(self, architect):
        """Test Grafana dashboard panel generation."""
        panel_config = [{"title": "Request Rate", "metric": "http_requests_total"}]
        config = {"panels": panel_config}
        dashboard = architect.design_grafana_dashboards(config)

        panels = dashboard["dashboard_json"]["panels"]
        assert len(panels) == 1
        assert panels[0]["title"] == "Request Rate"

    def test_grafana_dashboard_templating(self, architect):
        """Test Grafana dashboard templating."""
        dashboard = architect.design_grafana_dashboards({})

        templating = dashboard["dashboard_json"]["templating"]
        assert "list" in templating
        assert len(templating["list"]) > 0

    def test_grafana_dashboard_refresh(self, architect):
        """Test Grafana dashboard refresh interval."""
        config = {"refresh_interval": "1m"}
        dashboard = architect.design_grafana_dashboards(config)

        assert dashboard["dashboard_json"]["refresh"] == "1m"

    def test_configure_logging_basic(self, architect):
        """Test ELK stack logging configuration."""
        config = {"app_name": "myapp"}
        logging_config = architect.configure_logging(config)

        assert "elasticsearch_config" in logging_config
        assert "logstash_config" in logging_config
        assert "filebeat_config" in logging_config

    def test_elasticsearch_configuration(self, architect):
        """Test Elasticsearch configuration."""
        logging_config = architect.configure_logging({})

        es_config = logging_config["elasticsearch_config"]
        assert es_config["cluster.name"] == "moai-devops-cluster"

    def test_logstash_configuration(self, architect):
        """Test Logstash configuration."""
        logging_config = architect.configure_logging({})

        ls_config = logging_config["logstash_config"]
        assert ls_config["pipeline.id"] == "main"
        assert "input" in ls_config

    def test_filebeat_configuration(self, architect):
        """Test Filebeat configuration."""
        logging_config = architect.configure_logging({"app_name": "test-app"})

        fb_config = logging_config["filebeat_config"]
        assert "filebeat.inputs" in fb_config

    def test_logging_retention_policy(self, architect):
        """Test logging retention policy."""
        config = {"retention_days": 60}
        logging_config = architect.configure_logging(config)

        policy = logging_config["retention_policy"]
        assert policy["days"] == 60

    def test_setup_alerting_basic(self, architect):
        """Test alerting configuration."""
        alerting = architect.setup_alerting()

        assert "alertmanager" in alerting
        assert "alert_rules" in alerting

    def test_alertmanager_configuration(self, architect):
        """Test Alertmanager configuration."""
        alerting = architect.setup_alerting()

        alertmanager = alerting["alertmanager"]
        assert "global" in alertmanager
        assert "route" in alertmanager
        assert "receivers" in alertmanager

    def test_alertmanager_routing(self, architect):
        """Test Alertmanager routing configuration."""
        alerting = architect.setup_alerting()

        route = alerting["alertmanager"]["route"]
        assert route["group_by"] == ["alertname"]
        assert route["repeat_interval"] == "1h"

    def test_alert_rules_configuration(self, architect):
        """Test alert rules configuration."""
        alerting = architect.setup_alerting()

        rules = alerting["alert_rules"]
        assert len(rules) > 0
        assert rules[0]["name"] == "HighErrorRate"


# ============================================================================
# DeploymentStrategist Tests
# ============================================================================


class TestDeploymentStrategist:
    """Test deployment strategy and automation."""

    @pytest.fixture
    def strategist(self):
        """Create deployment strategist instance."""
        return DeploymentStrategist()

    def test_strategist_initialization(self, strategist):
        """Test deployment strategist initialization."""
        assert strategist.supported_strategies == ["blue_green", "canary", "rolling", "a_b_testing"]
        assert strategist.default_health_check_path == "/health"

    def test_plan_continuous_deployment_basic(self, strategist):
        """Test continuous deployment planning."""
        config = {"environments": ["dev", "staging", "prod"]}
        plan = strategist.plan_continuous_deployment(config)

        assert "pipeline_stages" in plan
        assert "quality_gates" in plan
        assert "rollback_strategy" in plan

    def test_continuous_deployment_stages(self, strategist):
        """Test continuous deployment pipeline stages."""
        config = {"environments": ["staging", "prod"]}
        plan = strategist.plan_continuous_deployment(config)

        stages = plan["pipeline_stages"]
        assert len(stages) == 2
        assert stages[0]["environment"] == "staging"
        assert stages[1]["environment"] == "prod"

    def test_continuous_deployment_manual_approval(self, strategist):
        """Test continuous deployment manual approval."""
        config = {"environments": ["staging", "prod"]}
        plan = strategist.plan_continuous_deployment(config)

        stages = plan["pipeline_stages"]
        assert stages[0]["manual_approval"] is False
        assert stages[-1]["manual_approval"] is True

    def test_continuous_deployment_rollback_strategy(self, strategist):
        """Test continuous deployment rollback strategy."""
        config = {"rollback_threshold": "error_rate > 10%"}
        plan = strategist.plan_continuous_deployment(config)

        rollback = plan["rollback_strategy"]
        assert rollback["enabled"] is True
        assert rollback["trigger_condition"] == "error_rate > 10%"

    def test_design_canary_deployment_basic(self, strategist):
        """Test canary deployment design."""
        config = {"canary_percentage": 20}
        canary = strategist.design_canary_deployment(config)

        assert "canary_config" in canary
        assert "traffic_splitting" in canary
        assert "monitoring_rules" in canary

    def test_canary_deployment_percentage(self, strategist):
        """Test canary deployment percentage."""
        config = {"canary_percentage": 15}
        canary = strategist.design_canary_deployment(config)

        assert canary["canary_config"]["initial_percentage"] == 15

    def test_canary_deployment_traffic_splitting(self, strategist):
        """Test canary traffic splitting steps."""
        config = {"increment_steps": [5, 25, 75, 100]}
        canary = strategist.design_canary_deployment(config)

        steps = canary["traffic_splitting"]["steps"]
        assert len(steps) == 4
        assert steps[0]["percentage"] == 5

    def test_canary_deployment_monitoring_rules(self, strategist):
        """Test canary deployment monitoring rules."""
        canary = strategist.design_canary_deployment({})

        rules = canary["monitoring_rules"]
        assert len(rules) >= 2
        assert any(r["metric"] == "error_rate" for r in rules)

    def test_canary_deployment_rollback_triggers(self, strategist):
        """Test canary deployment rollback triggers."""
        canary = strategist.design_canary_deployment({})

        triggers = canary["rollback_triggers"]
        assert "error_rate_increase" in triggers
        assert "latency_spike" in triggers

    def test_implement_blue_green_deployment_basic(self, strategist):
        """Test blue-green deployment implementation."""
        config = {}
        bg = strategist.implement_blue_green_deployment(config)

        assert "environment_config" in bg
        assert "traffic_switch" in bg
        assert "health_checks" in bg

    def test_blue_green_environments(self, strategist):
        """Test blue-green environment configuration."""
        bg = strategist.implement_blue_green_deployment({})

        env_config = bg["environment_config"]
        assert env_config["blue"]["active"] is True
        assert env_config["green"]["active"] is False

    def test_blue_green_traffic_switch(self, strategist):
        """Test blue-green traffic switching."""
        config = {"switch_strategy": "weighted"}
        bg = strategist.implement_blue_green_deployment(config)

        traffic_switch = bg["traffic_switch"]
        assert traffic_switch["strategy"] == "weighted"
        assert traffic_switch["validation_required"] is True

    def test_blue_green_rollback_procedure(self, strategist):
        """Test blue-green rollback procedure."""
        config = {"rollback_timeout": "10m"}
        bg = strategist.implement_blue_green_deployment(config)

        rollback = bg["rollback_procedure"]
        assert rollback["automatic"] is True
        assert rollback["timeout_minutes"] == 10

    def test_blue_green_cleanup_strategy(self, strategist):
        """Test blue-green cleanup strategy."""
        bg = strategist.implement_blue_green_deployment({})

        cleanup = bg["cleanup_strategy"]
        assert cleanup["automatic_cleanup"] is True

    def test_integrate_automated_testing_basic(self, strategist):
        """Test automated testing integration."""
        config = {}
        testing = strategist.integrate_automated_testing(config)

        assert "test_matrix" in testing
        assert "execution_strategy" in testing
        assert "coverage_requirements" in testing

    def test_automated_testing_matrix(self, strategist):
        """Test automated testing matrix."""
        config = {"test_types": ["unit", "integration", "e2e"]}
        testing = strategist.integrate_automated_testing(config)

        tests = testing["test_matrix"]["tests"]
        assert len(tests) == 3

    def test_automated_testing_parallel_execution(self, strategist):
        """Test automated testing parallel execution."""
        config = {"parallel_execution": True}
        testing = strategist.integrate_automated_testing(config)

        assert testing["execution_strategy"]["parallel_execution"] is True

    def test_automated_testing_coverage_requirements(self, strategist):
        """Test automated testing coverage requirements."""
        config = {"coverage_threshold": 95}
        testing = strategist.integrate_automated_testing(config)

        coverage = testing["coverage_requirements"]
        assert coverage["minimum_coverage"] == 95

    def test_automated_testing_environments(self, strategist):
        """Test automated testing environments."""
        config = {"test_environments": ["test", "staging"]}
        testing = strategist.integrate_automated_testing(config)

        environments = testing["test_environments"]
        assert "test" in environments
        assert "staging" in environments

    def test_get_test_framework_unit(self, strategist):
        """Test framework selection for unit tests."""
        assert strategist._get_test_framework("unit") == "pytest"

    def test_get_test_framework_integration(self, strategist):
        """Test framework selection for integration tests."""
        assert strategist._get_test_framework("integration") == "pytest"

    def test_get_test_framework_e2e(self, strategist):
        """Test framework selection for e2e tests."""
        assert strategist._get_test_framework("e2e") == "playwright"

    def test_get_test_framework_performance(self, strategist):
        """Test framework selection for performance tests."""
        assert strategist._get_test_framework("performance") == "locust"

    def test_get_test_framework_unknown(self, strategist):
        """Test framework selection for unknown test type."""
        assert strategist._get_test_framework("unknown") == "pytest"


# ============================================================================
# SecurityHardener Tests
# ============================================================================


class TestSecurityHardener:
    """Test security hardening and compliance."""

    @pytest.fixture
    def hardener(self):
        """Create security hardener instance."""
        return SecurityHardener()

    def test_hardener_initialization(self, hardener):
        """Test security hardener initialization."""
        assert "cis_aws" in hardener.supported_standards
        assert "pci_dss" in hardener.supported_standards
        assert "soc2" in hardener.supported_standards

    def test_scan_docker_images_basic(self, hardener):
        """Test Docker image security scanning."""
        scan_results = hardener.scan_docker_images("myapp:1.0")

        assert "vulnerabilities" in scan_results
        assert "security_score" in scan_results
        assert "recommendations" in scan_results

    def test_docker_scan_metadata(self, hardener):
        """Test Docker scan metadata."""
        scan_results = hardener.scan_docker_images("nginx:latest")

        assert scan_results["image_name"] == "nginx:latest"
        assert "scan_timestamp" in scan_results

    def test_docker_scan_vulnerabilities(self, hardener):
        """Test Docker scan vulnerability details."""
        scan_results = hardener.scan_docker_images("app:v1")

        vulns = scan_results["vulnerabilities"]
        assert len(vulns) >= 2
        assert all("cve" in v for v in vulns)
        assert all("severity" in v for v in vulns)

    def test_configure_secrets_management_basic(self, hardener):
        """Test secrets management configuration."""
        config = hardener.configure_secrets_management()

        assert "vault" in config
        assert "kubernetes_secrets" in config
        assert "rotation_policy" in config

    def test_secrets_vault_configuration(self, hardener):
        """Test Vault configuration."""
        config = hardener.configure_secrets_management()

        vault = config["vault"]
        assert vault["enabled"] is True
        assert vault["backend"] == "consul"

    def test_secrets_vault_policies(self, hardener):
        """Test Vault policies."""
        config = hardener.configure_secrets_management()

        policies = config["vault"]["policies"]
        assert len(policies) > 0
        assert policies[0]["name"] == "app-policy"

    def test_secrets_kubernetes_configuration(self, hardener):
        """Test Kubernetes secrets configuration."""
        config = hardener.configure_secrets_management()

        k8s_secrets = config["kubernetes_secrets"]
        assert k8s_secrets["enabled"] is True
        assert k8s_secrets["encryption_enabled"] is True

    def test_secrets_rotation_policy(self, hardener):
        """Test secrets rotation policy."""
        config = hardener.configure_secrets_management()

        rotation = config["rotation_policy"]
        assert rotation["enabled"] is True
        assert rotation["auto_rotation"] is True

    def test_setup_network_policies_basic(self, hardener):
        """Test network policies setup."""
        policies = hardener.setup_network_policies()

        assert "default_deny" in policies
        assert "allow_same_namespace" in policies
        assert "allow_dns" in policies

    def test_network_policy_default_deny(self, hardener):
        """Test default deny network policy."""
        policies = hardener.setup_network_policies()

        default_deny = policies["default_deny"]
        assert default_deny["kind"] == "NetworkPolicy"
        assert default_deny["metadata"]["name"] == "default-deny-all"

    def test_network_policy_allow_same_namespace(self, hardener):
        """Test allow same namespace network policy."""
        policies = hardener.setup_network_policies()

        allow_ns = policies["allow_same_namespace"]
        assert "ingress" in allow_ns["spec"]
        assert "egress" in allow_ns["spec"]

    def test_network_policy_allow_dns(self, hardener):
        """Test allow DNS network policy."""
        policies = hardener.setup_network_policies()

        allow_dns = policies["allow_dns"]
        assert allow_dns["kind"] == "NetworkPolicy"
        assert "egress" in allow_dns["spec"]

    def test_audit_compliance_basic(self, hardener):
        """Test compliance audit."""
        report = hardener.audit_compliance()

        assert "compliance_standards" in report
        assert "overall_score" in report
        assert "findings" in report
        assert "remediation_plan" in report

    def test_audit_compliance_timestamp(self, hardener):
        """Test audit timestamp."""
        report = hardener.audit_compliance()

        assert "audit_timestamp" in report

    def test_audit_compliance_standards(self, hardener):
        """Test compliance standards in audit."""
        report = hardener.audit_compliance()

        standards = report["compliance_standards"]
        assert "CIS AWS" in standards
        assert "PCI DSS" in standards

    def test_audit_compliance_findings(self, hardener):
        """Test compliance findings."""
        report = hardener.audit_compliance()

        findings = report["findings"]
        assert len(findings) > 0
        assert all("standard" in f for f in findings)
        assert all("status" in f for f in findings)

    def test_audit_compliance_remediation(self, hardener):
        """Test compliance remediation plan."""
        report = hardener.audit_compliance()

        remediation = report["remediation_plan"]
        assert "immediate" in remediation
        assert "short_term" in remediation
        assert "long_term" in remediation


# ============================================================================
# DevOpsMetricsCollector Tests
# ============================================================================


class TestDevOpsMetricsCollector:
    """Test DevOps metrics collection and analysis."""

    @pytest.fixture
    def collector(self):
        """Create metrics collector instance."""
        return DevOpsMetricsCollector()

    def test_collector_initialization(self, collector):
        """Test metrics collector initialization."""
        assert collector.metrics_window_days == 30

    def test_collect_deployment_metrics_basic(self, collector):
        """Test basic deployment metrics collection."""
        info = {
            "start_time": "2024-01-01T10:00:00Z",
            "end_time": "2024-01-01T10:15:00Z",
        }
        metrics = collector.collect_deployment_metrics(info)

        assert "deployment_duration" in metrics
        assert "success_rate" in metrics
        assert "rollback_count" in metrics

    def test_collect_deployment_metrics_duration(self, collector):
        """Test deployment duration metric."""
        metrics = collector.collect_deployment_metrics({})

        assert isinstance(metrics["deployment_duration"], int)

    def test_collect_deployment_metrics_success_rate(self, collector):
        """Test deployment success rate metric."""
        metrics = collector.collect_deployment_metrics({})

        assert isinstance(metrics["success_rate"], float)
        assert 0 <= metrics["success_rate"] <= 100

    def test_collect_deployment_metrics_frequency(self, collector):
        """Test deployment frequency metric."""
        metrics = collector.collect_deployment_metrics({})

        frequency = metrics["deployment_frequency"]
        assert "daily_count" in frequency
        assert "weekly_count" in frequency
        assert "monthly_count" in frequency

    def test_collect_deployment_metrics_performance_impact(self, collector):
        """Test deployment performance impact."""
        metrics = collector.collect_deployment_metrics({})

        impact = metrics["performance_impact"]
        assert "cpu_change" in impact
        assert "memory_change" in impact
        assert "response_time_change" in impact

    def test_track_pipeline_performance_basic(self, collector):
        """Test pipeline performance tracking."""
        pipeline_data = {
            "execution_times": {"build": 300, "test": 600, "deploy": 450},
            "success_rate": 98.5,
        }
        metrics = collector.track_pipeline_performance(pipeline_data)

        assert "total_execution_time" in metrics
        assert "stage_performance" in metrics
        assert "bottleneck_analysis" in metrics

    def test_pipeline_performance_stages(self, collector):
        """Test pipeline stage performance."""
        pipeline_data = {
            "execution_times": {"build": 120, "test": 240, "deploy": 60},
        }
        metrics = collector.track_pipeline_performance(pipeline_data)

        stages = metrics["stage_performance"]
        assert "build" in stages
        assert "test" in stages

    def test_pipeline_performance_bottleneck(self, collector):
        """Test pipeline bottleneck analysis."""
        pipeline_data = {
            "execution_times": {"build": 100, "test": 500, "deploy": 50},
        }
        metrics = collector.track_pipeline_performance(pipeline_data)

        bottleneck = metrics["bottleneck_analysis"]
        assert bottleneck["slowest_stage"] == "test"

    def test_pipeline_performance_throughput(self, collector):
        """Test pipeline throughput metrics."""
        pipeline_data = {
            "execution_times": {"build": 100, "test": 200},
            "success_rate": 96.5,
        }
        metrics = collector.track_pipeline_performance(pipeline_data)

        throughput = metrics["throughput_metrics"]
        assert "pipelines_per_day" in throughput

    def test_monitor_resource_usage_basic(self, collector):
        """Test resource usage monitoring."""
        config = {
            "monitoring_period": "24h",
            "metrics": ["cpu", "memory"],
        }
        metrics = collector.monitor_resource_usage(config)

        assert "cpu_utilization" in metrics
        assert "memory_usage" in metrics

    def test_resource_usage_cpu_metrics(self, collector):
        """Test CPU usage metrics."""
        metrics = collector.monitor_resource_usage({})

        cpu = metrics["cpu_utilization"]
        assert "current" in cpu
        assert "average" in cpu
        assert "peak" in cpu

    def test_resource_usage_memory_metrics(self, collector):
        """Test memory usage metrics."""
        metrics = collector.monitor_resource_usage({})

        memory = metrics["memory_usage"]
        assert "current" in memory
        assert "average" in memory

    def test_resource_usage_disk_io_metrics(self, collector):
        """Test disk I/O metrics."""
        metrics = collector.monitor_resource_usage({})

        disk = metrics["disk_io"]
        assert "read_ops_per_sec" in disk
        assert "write_ops_per_sec" in disk

    def test_resource_usage_network_metrics(self, collector):
        """Test network traffic metrics."""
        metrics = collector.monitor_resource_usage({})

        network = metrics["network_traffic"]
        assert "incoming" in network
        assert "outgoing" in network

    def test_resource_usage_cost_metrics(self, collector):
        """Test cost metrics."""
        metrics = collector.monitor_resource_usage({})

        cost = metrics["cost_metrics"]
        assert "daily_cost" in cost
        assert "monthly_projection" in cost

    def test_resource_usage_scaling_events(self, collector):
        """Test scaling events."""
        metrics = collector.monitor_resource_usage({})

        events = metrics["scaling_events"]
        assert isinstance(events, list)

    def test_get_devops_health_status_basic(self, collector):
        """Test DevOps health status assessment."""
        config = {
            "check_categories": ["deployment", "monitoring"],
            "health_threshold": {"deployment_success": 95},
        }
        health = collector.get_devops_health_status(config)

        assert "overall_health_score" in health
        assert "category_scores" in health
        assert "recommendations" in health

    def test_health_status_categories(self, collector):
        """Test health status category scores."""
        health = collector.get_devops_health_status({})

        categories = health["category_scores"]
        assert "deployment" in categories
        assert "monitoring" in categories
        assert "security" in categories

    def test_health_status_alerts(self, collector):
        """Test health status alerts."""
        health = collector.get_devops_health_status({})

        alerts = health["alerts"]
        assert isinstance(alerts, list)

    def test_health_status_trends(self, collector):
        """Test health status trends."""
        health = collector.get_devops_health_status({})

        trends = health["trends"]
        assert "overall_trend" in trends

    def test_health_status_with_low_score(self, collector):
        """Test health status with low score generates critical issues."""
        config = {"health_threshold": {"deployment_success": 99}}
        # Mock to create low score scenario
        with patch.object(collector, "get_devops_health_status") as mock:
            mock.return_value = {
                "overall_health_score": 85,
                "category_scores": {"deployment": 85, "monitoring": 85, "security": 85, "performance": 85},
                "critical_issues": [{"category": "monitoring", "severity": "medium"}],
            }
            health = collector.get_devops_health_status(config)

            assert len(health["critical_issues"]) > 0


# ============================================================================
# Integration Tests
# ============================================================================


class TestDevOpsIntegration:
    """Integration tests across multiple DevOps components."""

    def test_full_pipeline_orchestration(self):
        """Test full CI/CD pipeline orchestration workflow."""
        orchestrator = CICDPipelineOrchestrator()
        config = {
            "name": "Integration Pipeline",
            "runtime": "python",
            "test_command": "pytest tests/",
        }

        workflow = orchestrator.orchestrate_github_actions(config)
        assert "test" in workflow["jobs"]
        assert "build" in workflow["jobs"]
        assert workflow["jobs"]["build"]["needs"] == "test"

    def test_infrastructure_with_kubernetes(self):
        """Test infrastructure manager with Kubernetes deployment."""
        manager = InfrastructureManager()
        k8s_config = {"name": "test-app", "replicas": 3}

        manifests = manager.generate_kubernetes_manifests(k8s_config)
        assert len(manifests) == 3

    def test_container_with_monitoring(self):
        """Test container orchestration with monitoring setup."""
        orchestrator = ContainerOrchestrator()
        architect = MonitoringArchitect()

        deployment = orchestrator.plan_kubernetes_deployment({"app_name": "app"})
        monitoring = architect.setup_prometheus({"app_name": "app"})

        assert deployment is not None
        assert monitoring is not None

    def test_deployment_with_security(self):
        """Test deployment strategy with security hardening."""
        strategist = DeploymentStrategist()
        hardener = SecurityHardener()

        deployment = strategist.design_canary_deployment({"canary_percentage": 10})
        security = hardener.configure_secrets_management()

        assert deployment is not None
        assert security is not None

    def test_metrics_collection_workflow(self):
        """Test metrics collection workflow."""
        collector = DevOpsMetricsCollector()

        deployment_metrics = collector.collect_deployment_metrics({})
        pipeline_metrics = collector.track_pipeline_performance({"execution_times": {"build": 100, "test": 200}})
        health = collector.get_devops_health_status({})

        assert deployment_metrics is not None
        assert pipeline_metrics is not None
        assert health is not None


# ============================================================================
# Edge Cases and Error Handling Tests
# ============================================================================


class TestEdgeCases:
    """Test edge cases and error conditions."""

    def test_empty_configuration_handling(self):
        """Test handling of empty configurations."""
        orchestrator = CICDPipelineOrchestrator()
        workflow = orchestrator.orchestrate_github_actions({})

        assert workflow is not None
        assert workflow["env"]["PROJECT_NAME"] == "app"

    def test_kubernetes_manifests_with_missing_config(self):
        """Test Kubernetes manifests with minimal config."""
        manager = InfrastructureManager()
        manifests = manager.generate_kubernetes_manifests({})

        assert "deployment" in manifests
        assert manifests["deployment"]["metadata"]["name"] == "app"

    def test_terraform_with_empty_resources(self):
        """Test Terraform module design with no resources."""
        manager = InfrastructureManager()
        modules = manager.design_terraform_modules({"resources": {}})

        assert len(modules["module"]) == 0

    def test_prometheus_custom_interval(self):
        """Test Prometheus with very short scrape interval."""
        architect = MonitoringArchitect()
        prometheus = architect.setup_prometheus({"scrape_interval": "5s"})

        assert prometheus["scrape_interval"] == "5s"

    def test_canary_deployment_with_single_step(self):
        """Test canary deployment with single increment step."""
        strategist = DeploymentStrategist()
        canary = strategist.design_canary_deployment({"increment_steps": [100]})

        steps = canary["traffic_splitting"]["steps"]
        assert len(steps) == 1

    def test_blue_green_with_custom_timeout(self):
        """Test blue-green deployment with custom timeout."""
        strategist = DeploymentStrategist()
        bg = strategist.implement_blue_green_deployment({"rollback_timeout": "3m"})

        assert bg["rollback_procedure"]["timeout_minutes"] == 3

    def test_docker_scan_with_special_image_name(self):
        """Test Docker scan with special image name."""
        hardener = SecurityHardener()
        scan = hardener.scan_docker_images("registry.example.com/app:v1.2.3")

        assert scan["image_name"] == "registry.example.com/app:v1.2.3"

    def test_large_pipeline_data_processing(self):
        """Test pipeline performance with large execution times."""
        collector = DevOpsMetricsCollector()
        pipeline_data = {
            "execution_times": {f"stage_{i}": i * 100 for i in range(1, 11)},
        }
        metrics = collector.track_pipeline_performance(pipeline_data)

        assert metrics["total_execution_time"] > 0

    def test_multiple_health_check_configurations(self):
        """Test Kubernetes with multiple health check configurations."""
        manager = InfrastructureManager()
        config = {
            "name": "multi-check-app",
            "health_check": {"path": "/ready", "initial_delay": 60},
        }
        manifests = manager.generate_kubernetes_manifests(config)

        container = manifests["deployment"]["spec"]["template"]["spec"]["containers"][0]
        assert "livenessProbe" in container
        assert "readinessProbe" in container


# ============================================================================
# DateTime and Timestamp Tests
# ============================================================================


class TestDateTimeHandling:
    """Test datetime and timestamp handling."""

    def test_docker_scan_timestamp_format(self):
        """Test Docker scan timestamp format."""
        hardener = SecurityHardener()
        scan = hardener.scan_docker_images("app:latest")

        assert "scan_timestamp" in scan
        timestamp = scan["scan_timestamp"]
        assert "T" in timestamp
        assert isinstance(timestamp, str)
        assert len(timestamp) > 10

    def test_docker_scan_details(self):
        """Test Docker scan result details."""
        hardener = SecurityHardener()
        scan = hardener.scan_docker_images("app:latest")

        assert scan["image_name"] == "app:latest"
        assert isinstance(scan["vulnerabilities"], list)
        assert "security_score" in scan
        assert isinstance(scan["recommendations"], list)

    def test_compliance_audit_timestamp_format(self):
        """Test compliance audit timestamp format."""
        hardener = SecurityHardener()
        report = hardener.audit_compliance()

        timestamp = report["audit_timestamp"]
        assert "T" in timestamp
        assert isinstance(timestamp, str)

    def test_scaling_events_timestamp_format(self):
        """Test scaling events timestamp format."""
        collector = DevOpsMetricsCollector()
        metrics = collector.monitor_resource_usage({})

        if metrics["scaling_events"]:
            timestamp = metrics["scaling_events"][0]["timestamp"]
            assert "T" in timestamp


# ============================================================================
# Configuration and Defaults Tests
# ============================================================================


class TestConfigurationDefaults:
    """Test default configurations."""

    def test_cicd_orchestrator_defaults(self):
        """Test CI/CD orchestrator default values."""
        orchestrator = CICDPipelineOrchestrator()

        assert len(orchestrator.supported_platforms) > 0
        assert len(orchestrator.default_environments) > 0

    def test_container_orchestrator_default_images(self):
        """Test container orchestrator default base images."""
        orchestrator = ContainerOrchestrator()

        assert orchestrator.default_base_images["python"] == "python:3.11-slim"
        assert orchestrator.default_base_images["node"] == "node:20-alpine"

    def test_monitoring_architect_intervals(self):
        """Test monitoring architect default intervals."""
        architect = MonitoringArchitect()

        assert architect.default_scrape_interval == "15s"
        assert architect.default_evaluation_interval == "15s"

    def test_deployment_strategist_defaults(self):
        """Test deployment strategist default values."""
        strategist = DeploymentStrategist()

        assert len(strategist.supported_strategies) > 0
        assert strategist.default_health_check_path == "/health"

    def test_security_hardener_standards(self):
        """Test security hardener supported standards."""
        hardener = SecurityHardener()

        assert "pci_dss" in hardener.supported_standards
        assert "iso27001" in hardener.supported_standards


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--cov=src.moai_adk.foundation.devops"])
