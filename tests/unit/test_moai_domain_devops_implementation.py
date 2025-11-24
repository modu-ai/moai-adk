"""
DevOps Implementation Tests

Test suite for DevOps automation capabilities including CI/CD pipelines,
infrastructure as code, container orchestration, monitoring, and security.
Tests cover enterprise DevOps patterns with modern tools and practices.
"""

import pytest
from unittest.mock import Mock, patch, MagicMock
import json
import yaml
from typing import Dict, Any, List, Optional
import tempfile
import os

# Import the DevOps implementation
try:
    from src.moai_adk.foundation.devops import (
        CICDOrchestrator,
        InfrastructureManager,
        ContainerOrchestrator,
        MonitoringSetupManager,
        SecurityComplianceManager,
        DeploymentAutomationEngine,
        DevOpsMetricsCollector,
    )
except ImportError:
    pytest.skip("DevOps implementation not available", allow_module_level=True)


class TestCICDPipelineOrchestration:
    """Test CI/CD pipeline orchestration capabilities"""

    def test_github_actions_workflow(self):
        """Test GitHub Actions workflow generation"""
        orchestrator = CICDOrchestrator()

        # Generate GitHub Actions workflow
        workflow = orchestrator.create_github_actions_workflow(
            project_name="test-app",
            languages=["python", "javascript"],
            environments=["dev", "staging", "prod"]
        )

        # Validate workflow structure
        assert workflow is not None
        assert isinstance(workflow, dict)
        assert "name" in workflow
        assert "on" in workflow
        assert "jobs" in workflow

        # Validate triggers
        assert "push" in workflow["on"]
        assert "pull_request" in workflow["on"]

        # Validate job structure
        assert "test" in workflow["jobs"]
        assert "build" in workflow["jobs"]
        assert "deploy" in workflow["jobs"]

        # Validate test job matrix
        test_job = workflow["jobs"]["test"]
        assert "strategy" in test_job
        assert "matrix" in test_job["strategy"]

    def test_gitlab_ci_pipeline(self):
        """Test GitLab CI/CD pipeline configuration"""
        orchestrator = CICDOrchestrator()

        # Generate GitLab CI configuration
        pipeline = orchestrator.create_gitlab_ci_pipeline(
            stages=["build", "test", "security", "deploy"],
            docker_image="python:3.11-slim"
        )

        # Validate pipeline structure
        assert pipeline is not None
        assert isinstance(pipeline, dict)
        assert "stages" in pipeline
        assert "variables" in pipeline

        # Validate stages
        expected_stages = ["build", "test", "security", "deploy"]
        assert pipeline["stages"] == expected_stages

        # Validate build job
        assert "build" in pipeline
        build_job = pipeline["build"]
        assert "stage" in build_job
        assert "script" in build_job
        assert build_job["stage"] == "build"

    def test_jenkins_pipeline(self):
        """Test Jenkins pipeline automation"""
        orchestrator = CICDOrchestrator()

        # Generate Jenkinsfile
        pipeline = orchestrator.create_jenkins_pipeline(
            repository_url="https://github.com/test/repo.git",
            branches=["main", "develop"],
            build_tool="maven"
        )

        # Validate pipeline structure
        assert pipeline is not None
        assert isinstance(pipeline, str)
        assert "pipeline" in pipeline
        assert "agent" in pipeline
        assert "stages" in pipeline

        # Validate stages
        assert "Checkout" in pipeline
        assert "Build" in pipeline
        assert "Test" in pipeline
        assert "Deploy" in pipeline

    def test_multi_environment_deployment(self):
        """Test multi-environment deployment strategy"""
        orchestrator = CICDOrchestrator()

        # Generate deployment strategy
        strategy = orchestrator.orchestrate_multi_env_deployment(
            environments=["dev", "staging", "prod"],
            promotion_strategy="manual",
            rollback_enabled=True
        )

        # Validate strategy
        assert strategy is not None
        assert isinstance(strategy, dict)
        assert "environments" in strategy
        assert "promotion_rules" in strategy
        assert "rollback_strategy" in strategy

        # Validate environments
        envs = strategy["environments"]
        assert "dev" in envs
        assert "staging" in envs
        assert "prod" in envs

        # Validate rollback configuration
        rollback = strategy["rollback_strategy"]
        assert rollback["enabled"] is True
        assert "automated_triggers" in rollback


class TestInfrastructureAsCode:
    """Test infrastructure as code capabilities"""

    def test_terraform_infrastructure(self):
        """Test Terraform IaC configuration"""
        infra_manager = InfrastructureManager()

        # Generate Terraform configuration
        config = infra_manager.generate_terraform_config(
            provider="aws",
            region="us-west-2",
            resources=["vpc", "ec2", "rds"]
        )

        # Validate configuration
        assert config is not None
        assert isinstance(config, dict)
        assert "provider" in config
        assert "resources" in config
        assert "variables" in config

        # Validate provider configuration
        provider_config = config["provider"]
        assert provider_config["name"] == "aws"
        assert provider_config["region"] == "us-west-2"

        # Validate resources
        resources = config["resources"]
        assert "vpc" in resources
        assert "ec2" in resources
        assert "rds" in resources

    def test_kubernetes_manifests(self):
        """Test Kubernetes manifests generation"""
        infra_manager = InfrastructureManager()

        # Generate Kubernetes manifests
        manifests = infra_manager.create_kubernetes_manifests(
            app_name="test-app",
            image="nginx:latest",
            replicas=3,
            namespace="production"
        )

        # Validate manifests
        assert manifests is not None
        assert isinstance(manifests, dict)
        assert "deployment" in manifests
        assert "service" in manifests
        assert "configmap" in manifests

        # Validate deployment
        deployment = manifests["deployment"]
        assert deployment["metadata"]["name"] == "test-app"
        assert deployment["spec"]["replicas"] == 3
        assert deployment["spec"]["selector"]["matchLabels"]["app"] == "test-app"

        # Validate service
        service = manifests["service"]
        assert service["metadata"]["name"] == "test-app"
        assert service["spec"]["selector"]["app"] == "test-app"

    def test_ansible_configuration(self):
        """Test Ansible automation setup"""
        infra_manager = InfrastructureManager()

        # Generate Ansible playbook
        playbook = infra_manager.setup_ansible_playbooks(
            target_hosts="webservers",
            tasks=["install_nginx", "configure_firewall", "deploy_app"]
        )

        # Validate playbook
        assert playbook is not None
        assert isinstance(playbook, dict)
        assert "hosts" in playbook
        assert "tasks" in playbook
        assert "vars" in playbook

        # Validate tasks
        tasks = playbook["tasks"]
        assert len(tasks) == 3
        assert tasks[0]["name"] == "install_nginx"
        assert tasks[1]["name"] == "configure_firewall"
        assert tasks[2]["name"] == "deploy_app"


class TestContainerOrchestration:
    """Test container orchestration capabilities"""

    def test_docker_multi_stage_build(self):
        """Test Docker multi-stage build optimization"""
        container_orchestrator = ContainerOrchestrator()

        # Generate Dockerfile
        dockerfile_config = container_orchestrator.create_dockerfile(
            base_image="python:3.11-slim",
            app_type="web_service",
            optimization_level="production"
        )

        # Validate Dockerfile configuration
        assert dockerfile_config is not None
        assert isinstance(dockerfile_config, dict)
        assert "dockerfile_path" in dockerfile_config
        assert "build_context" in dockerfile_config
        assert "optimizations" in dockerfile_config

        # Validate optimizations
        optimizations = dockerfile_config["optimizations"]
        assert "multi_stage" in optimizations
        assert "layer_caching" in optimizations
        assert "security_hardening" in optimizations

        # Validate build configuration
        assert "target_platforms" in dockerfile_config
        assert "build_args" in dockerfile_config

    def test_kubernetes_deployment(self):
        """Test Kubernetes deployment strategy"""
        container_orchestrator = ContainerOrchestrator()

        # Generate K8s deployment spec
        deployment = container_orchestrator.deploy_kubernetes_service(
            service_name="api-service",
            container_image="myapp:v1.0.0",
            port=8080,
            replicas=3
        )

        # Validate deployment
        assert deployment is not None
        assert isinstance(deployment, dict)
        assert "apiVersion" in deployment
        assert "kind" in deployment
        assert "metadata" in deployment
        assert "spec" in deployment

        # Validate spec
        spec = deployment["spec"]
        assert spec["replicas"] == 3
        assert "selector" in spec
        assert "template" in spec

        # Validate container configuration
        containers = spec["template"]["spec"]["containers"]
        assert len(containers) >= 1
        container = containers[0]
        assert container["image"] == "myapp:v1.0.0"
        assert container["ports"][0]["containerPort"] == 8080

    def test_helm_chart_configuration(self):
        """Test Helm chart configuration"""
        container_orchestrator = ContainerOrchestrator()

        # Generate Helm chart template
        helm_chart = container_orchestrator.configure_helm_chart(
            chart_name="web-app",
            app_version="1.0.0",
            values={
                "replicaCount": 3,
                "image": {"repository": "nginx", "tag": "latest"},
                "service": {"type": "ClusterIP", "port": 80}
            }
        )

        # Validate Helm chart
        assert helm_chart is not None
        assert isinstance(helm_chart, dict)
        assert "chart_yaml" in helm_chart
        assert "values_yaml" in helm_chart
        assert "templates" in helm_chart

        # Validate Chart.yaml
        chart_yaml = helm_chart["chart_yaml"]
        assert chart_yaml["name"] == "web-app"
        assert chart_yaml["version"] == "0.1.0"  # Default chart version
        assert chart_yaml["appVersion"] == "1.0.0"

        # Validate values.yaml
        values_yaml = helm_chart["values_yaml"]
        assert values_yaml["replicaCount"] == 3
        assert values_yaml["image"]["repository"] == "nginx"
        assert values_yaml["service"]["type"] == "ClusterIP"


class TestMonitoringAndLogging:
    """Test monitoring and logging setup"""

    def test_prometheus_monitoring(self):
        """Test Prometheus monitoring configuration"""
        monitoring_manager = MonitoringSetupManager()

        # Configure Prometheus
        prometheus_config = monitoring_manager.configure_prometheus(
            scrape_interval="15s",
            evaluation_interval="15s",
            targets=[
                {"job": "api", "targets": ["api:8080"]},
                {"job": "web", "targets": ["web:80"]}
            ]
        )

        # Validate configuration
        assert prometheus_config is not None
        assert isinstance(prometheus_config, dict)
        assert "global" in prometheus_config
        assert "scrape_configs" in prometheus_config
        assert "rule_files" in prometheus_config

        # Validate global config
        global_config = prometheus_config["global"]
        assert global_config["scrape_interval"] == "15s"
        assert global_config["evaluation_interval"] == "15s"

        # Validate scrape configs
        scrape_configs = prometheus_config["scrape_configs"]
        assert len(scrape_configs) >= 2

        # Check API job config
        api_job = next((job for job in scrape_configs if job["job_name"] == "api"), None)
        assert api_job is not None
        assert "static_configs" in api_job
        assert api_job["static_configs"][0]["targets"] == ["api:8080"]

    def test_grafana_dashboards(self):
        """Test Grafana dashboard creation"""
        monitoring_manager = MonitoringSetupManager()

        # Create dashboards
        dashboards = monitoring_manager.create_grafana_dashboards(
            dashboard_type="application",
            metrics=["cpu_usage", "memory_usage", "request_rate", "error_rate"]
        )

        # Validate dashboards
        assert dashboards is not None
        assert isinstance(dashboards, list)
        assert len(dashboards) > 0

        # Validate main dashboard
        main_dashboard = dashboards[0]
        assert isinstance(main_dashboard, dict)
        assert "dashboard" in main_dashboard
        assert "title" in main_dashboard["dashboard"]
        assert "panels" in main_dashboard["dashboard"]

        # Validate panels
        panels = main_dashboard["dashboard"]["panels"]
        assert len(panels) >= 4

        # Check for specific metric panels
        panel_titles = [panel.get("title", "") for panel in panels]
        assert any("cpu" in title.lower() for title in panel_titles)
        assert any("memory" in title.lower() for title in panel_titles)
        assert any("request" in title.lower() for title in panel_titles)
        assert any("error" in title.lower() for title in panel_titles)

    def test_elk_stack_logging(self):
        """Test ELK stack configuration"""
        monitoring_manager = MonitoringSetupManager()

        # Setup ELK stack
        elk_config = monitoring_manager.setup_elk_stack(
            elasticsearch_version="8.11.0",
            logstash_pipeline=["input", "filter", "output"],
            kibana_dashboards=["logs", "metrics", "traces"]
        )

        # Validate ELK configuration
        assert elk_config is not None
        assert isinstance(elk_config, dict)
        assert "elasticsearch" in elk_config
        assert "logstash" in elk_config
        assert "kibana" in elk_config

        # Validate Elasticsearch
        elasticsearch = elk_config["elasticsearch"]
        assert "version" in elasticsearch
        assert "cluster" in elasticsearch
        assert elasticsearch["version"] == "8.11.0"

        # Validate Logstash pipeline
        logstash = elk_config["logstash"]
        assert "pipeline" in logstash
        assert "input" in logstash["pipeline"]
        assert "filter" in logstash["pipeline"]
        assert "output" in logstash["pipeline"]

        # Validate Kibana dashboards
        kibana = elk_config["kibana"]
        assert "dashboards" in kibana
        assert len(kibana["dashboards"]) >= 3


class TestSecurityAndCompliance:
    """Test security and compliance management"""

    def test_iam_policies(self):
        """Test IAM policy management"""
        security_manager = SecurityComplianceManager()

        # Create IAM policies
        policies = security_manager.create_iam_policies(
            resource_type="s3_bucket",
            actions=["read", "write"],
            conditions={"ip_address": {"source_ip": ["10.0.0.0/8"]}}
        )

        # Validate policies
        assert policies is not None
        assert isinstance(policies, dict)
        assert "policy_document" in policies
        assert "role_policy" in policies

        # Validate policy document
        policy_doc = policies["policy_document"]
        assert "Version" in policy_doc
        assert "Statement" in policy_doc
        assert len(policy_doc["Statement"]) > 0

        # Validate statement structure
        statement = policy_doc["Statement"][0]
        assert "Effect" in statement
        assert "Action" in statement
        assert "Resource" in statement
        assert "Condition" in statement

        # Validate role policy
        role_policy = policies["role_policy"]
        assert "assume_role_policy" in role_policy
        assert "inline_policies" in role_policy

    def test_network_security(self):
        """Test network security group configuration"""
        security_manager = SecurityComplianceManager()

        # Configure network security
        security_config = security_manager.configure_network_security(
            vpc_cidr="10.0.0.0/16",
            inbound_rules=[
                {"port": 80, "protocol": "tcp", "source": "0.0.0.0/0"},
                {"port": 443, "protocol": "tcp", "source": "0.0.0.0/0"},
                {"port": 22, "protocol": "tcp", "source": "10.0.0.0/8"}
            ],
            outbound_rules=[
                {"port": 0, "protocol": "-1", "destination": "0.0.0.0/0"}
            ]
        )

        # Validate security configuration
        assert security_config is not None
        assert isinstance(security_config, dict)
        assert "security_groups" in security_config
        assert "network_acls" in security_config

        # Validate security groups
        security_groups = security_config["security_groups"]
        assert len(security_groups) > 0

        # Check inbound rules
        sg = security_groups[0]
        assert "inbound_rules" in sg
        assert "outbound_rules" in sg
        assert len(sg["inbound_rules"]) >= 3
        assert len(sg["outbound_rules"]) >= 1

    def test_compliance_validation(self):
        """Test compliance validation framework"""
        security_manager = SecurityComplianceManager()

        # Validate compliance
        compliance_result = security_manager.validate_compliance(
            standards=["cis_aws", "pci_dss", "soc2"],
            resource_type="ec2_instance",
            resource_config={
                "instance_type": "t3.medium",
                "encryption_enabled": True,
                "iam_role_attached": True,
                "security_groups": ["sg-123456"],
                "monitoring_enabled": True
            }
        )

        # Validate compliance result
        assert compliance_result is not None
        assert isinstance(compliance_result, dict)
        assert "overall_score" in compliance_result
        assert "standards_results" in compliance_result
        assert "violations" in compliance_result
        assert "recommendations" in compliance_result

        # Validate overall score
        score = compliance_result["overall_score"]
        assert isinstance(score, (int, float))
        assert 0 <= score <= 100

        # Validate standards results
        standards_results = compliance_result["standards_results"]
        assert "cis_aws" in standards_results
        assert "pci_dss" in standards_results
        assert "soc2" in standards_results

        # Validate violations and recommendations
        violations = compliance_result["violations"]
        recommendations = compliance_result["recommendations"]
        assert isinstance(violations, list)
        assert isinstance(recommendations, list)


class TestDeploymentAutomation:
    """Test deployment automation engine"""

    def test_automate_rollout_strategy(self):
        """Test automated rollout strategy"""
        deployment_engine = DeploymentAutomationEngine()

        # Configure rollout strategy
        rollout_strategy = deployment_engine.automate_rollout_strategy(
            deployment_type="canary",
            rollout_percentage=[10, 25, 50, 100],
            health_check_endpoint="/health",
            rollback_triggers={"error_rate": 0.05, "latency_p95": 1000}
        )

        # Validate rollout strategy
        assert rollout_strategy is not None
        assert isinstance(rollout_strategy, dict)
        assert "phases" in rollout_strategy
        assert "health_checks" in rollout_strategy
        assert "rollback_conditions" in rollout_strategy

        # Validate phases
        phases = rollout_strategy["phases"]
        assert len(phases) == 4
        assert phases[0]["percentage"] == 10
        assert phases[3]["percentage"] == 100

        # Validate health checks
        health_checks = rollout_strategy["health_checks"]
        assert health_checks["endpoint"] == "/health"
        assert "success_threshold" in health_checks
        assert "timeout_seconds" in health_checks

    def test_blue_green_deployment(self):
        """Test blue-green deployment setup"""
        deployment_engine = DeploymentAutomationEngine()

        # Setup blue-green deployment
        bg_config = deployment_engine.setup_blue_green_deployment(
            service_name="api-service",
            load_balancer_type="application",
            traffic_split_percentage=50,
            health_check_path="/health"
        )

        # Validate configuration
        assert bg_config is not None
        assert isinstance(bg_config, dict)
        assert "blue_environment" in bg_config
        assert "green_environment" in bg_config
        assert "traffic_routing" in bg_config

        # Validate environments
        blue_env = bg_config["blue_environment"]
        green_env = bg_config["green_environment"]
        assert blue_env["name"] == "blue"
        assert green_env["name"] == "green"
        assert "deployment_config" in blue_env
        assert "deployment_config" in green_env

        # Validate traffic routing
        traffic_routing = bg_config["traffic_routing"]
        assert traffic_routing["split_percentage"] == 50
        assert "load_balancer" in traffic_routing
        assert "health_check" in traffic_routing["load_balancer"]

    def test_canary_deployment(self):
        """Test canary deployment configuration"""
        deployment_engine = DeploymentAutomationEngine()

        # Configure canary deployment
        canary_config = deployment_engine.configure_canary_deployment(
            service_name="web-app",
            initial_percentage=5,
            increment_percentage=10,
            max_percentage=50,
            analysis_duration_minutes=10
        )

        # Validate canary configuration
        assert canary_config is not None
        assert isinstance(canary_config, dict)
        assert "initial_rollout" in canary_config
        assert "incremental_rollout" in canary_config
        assert "analysis" in canary_config

        # Validate initial rollout
        initial = canary_config["initial_rollout"]
        assert initial["percentage"] == 5

        # Validate incremental rollout
        incremental = canary_config["incremental_rollout"]
        assert incremental["step_percentage"] == 10
        assert incremental["max_percentage"] == 50

        # Validate analysis
        analysis = canary_config["analysis"]
        assert analysis["duration_minutes"] == 10
        assert "metrics" in analysis

    def test_rollback_mechanism(self):
        """Test automated rollback mechanism"""
        deployment_engine = DeploymentAutomationEngine()

        # Configure rollback strategy
        rollback_strategy = deployment_engine.implement_rollback_mechanism(
            triggers=["error_rate_increase", "health_check_failure", "manual_trigger"],
            rollback_timeout_minutes=15,
            data_consistency_check=True
        )

        # Validate rollback strategy
        assert rollback_strategy is not None
        assert isinstance(rollback_strategy, dict)
        assert "trigger_conditions" in rollback_strategy
        assert "rollback_procedure" in rollback_strategy
        assert "validation" in rollback_strategy

        # Validate triggers
        triggers = rollback_strategy["trigger_conditions"]
        assert "error_rate_increase" in triggers
        assert "health_check_failure" in triggers
        assert "manual_trigger" in triggers

        # Validate rollback procedure
        procedure = rollback_strategy["rollback_procedure"]
        assert procedure["timeout_minutes"] == 15
        assert "steps" in procedure

        # Validate validation
        validation = rollback_strategy["validation"]
        assert validation["data_consistency_check"] is True
        assert "health_check_verification" in validation


class TestDevOpsMetricsCollection:
    """Test DevOps metrics collection and analysis"""

    def test_collect_deployment_metrics(self):
        """Test deployment metrics collection"""
        metrics_collector = DevOpsMetricsCollector()

        # Collect deployment metrics
        metrics = metrics_collector.collect_deployment_metrics(
            time_range="last_7_days",
            environments=["prod", "staging"],
            services=["api", "web", "worker"]
        )

        # Validate metrics
        assert metrics is not None
        assert isinstance(metrics, dict)
        assert "deployment_frequency" in metrics
        assert "lead_time_for_changes" in metrics
        assert "change_failure_rate" in metrics
        assert "mean_time_to_recovery" in metrics

        # Validate DORA metrics
        dora_metrics = metrics["dora_metrics"]
        assert "deployment_frequency" in dora_metrics
        assert "lead_time_for_changes" in dora_metrics
        assert "change_failure_rate" in dora_metrics
        assert "mean_time_to_recovery" in dora_metrics

        # Validate specific metric values
        deployment_freq = dora_metrics["deployment_frequency"]
        assert "value" in deployment_freq
        assert "unit" in deployment_freq
        assert "trend" in deployment_freq

    def test_monitor_infrastructure_health(self):
        """Test infrastructure health monitoring"""
        metrics_collector = DevOpsMetricsCollector()

        # Monitor infrastructure health
        health_status = metrics_collector.monitor_infrastructure_health(
            infrastructure_types=["kubernetes", "database", "network"],
            alert_thresholds={"cpu_usage": 80, "memory_usage": 85, "disk_usage": 90}
        )

        # Validate health status
        assert health_status is not None
        assert isinstance(health_status, dict)
        assert "overall_status" in health_status
        assert "component_status" in health_status
        assert "alerts" in health_status

        # Validate component status
        components = health_status["component_status"]
        assert "kubernetes" in components
        assert "database" in components
        assert "network" in components

        # Validate alerts
        alerts = health_status["alerts"]
        assert isinstance(alerts, list)
        for alert in alerts:
            assert "severity" in alert
            assert "component" in alert
            assert "metric" in alert
            assert "value" in alert

    def test_track_deployment_success_rate(self):
        """Test deployment success rate tracking"""
        metrics_collector = DevOpsMetricsCollector()

        # Track deployment success rate
        success_rate_metrics = metrics_collector.track_deployment_success_rate(
            time_period="30_days",
            deployment_stages=["build", "test", "deploy"],
            success_criteria={"all_tests_pass": True, "rollback_not_required": True}
        )

        # Validate success rate metrics
        assert success_rate_metrics is not None
        assert isinstance(success_rate_metrics, dict)
        assert "overall_success_rate" in success_rate_metrics
        assert "stage_success_rates" in success_rate_metrics
        assert "failure_analysis" in success_rate_metrics

        # Validate overall success rate
        overall_rate = success_rate_metrics["overall_success_rate"]
        assert "percentage" in overall_rate
        assert "total_deployments" in overall_rate
        assert "successful_deployments" in overall_rate
        assert "failed_deployments" in overall_rate

        # Validate stage success rates
        stage_rates = success_rate_metrics["stage_success_rates"]
        assert "build" in stage_rates
        assert "test" in stage_rates
        assert "deploy" in stage_rates

        # Validate failure analysis
        failure_analysis = success_rate_metrics["failure_analysis"]
        assert "common_failure_reasons" in failure_analysis
        assert "failure_patterns" in failure_analysis

    def test_generate_devops_dashboard(self):
        """Test DevOps dashboard data generation"""
        metrics_collector = DevOpsMetricsCollector()

        # Generate dashboard data
        dashboard_data = metrics_collector.generate_devops_dashboard(
            dashboard_type="executive",
            time_range="last_30_days",
            widgets=["deployment_trend", "system_health", "team_performance", "cost_analysis"]
        )

        # Validate dashboard data
        assert dashboard_data is not None
        assert isinstance(dashboard_data, dict)
        assert "metadata" in dashboard_data
        assert "widgets" in dashboard_data
        assert "filters" in dashboard_data

        # Validate metadata
        metadata = dashboard_data["metadata"]
        assert "title" in metadata
        assert "time_range" in metadata
        assert "last_updated" in metadata
        assert metadata["dashboard_type"] == "executive"

        # Validate widgets
        widgets = dashboard_data["widgets"]
        assert len(widgets) >= 4

        # Check for required widgets
        widget_types = [widget.get("type", "") for widget in widgets]
        assert "deployment_trend" in widget_types
        assert "system_health" in widget_types
        assert "team_performance" in widget_types
        assert "cost_analysis" in widget_types

        # Validate individual widget structure
        for widget in widgets:
            assert "type" in widget
            assert "title" in widget
            assert "data" in widget
            assert "visualization" in widget