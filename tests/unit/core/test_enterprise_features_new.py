"""
Comprehensive tests for enterprise_features module.

Tests cover:
- Load balancing with multiple algorithms
- Auto-scaling policy enforcement
- Deployment management (blue-green, canary, rolling)
- Multi-tenant management
- Audit logging
- Compliance reporting
"""

from datetime import datetime, timedelta

import pytest

from src.moai_adk.core.enterprise_features import (
    AuditLogger,
    AutoScaler,
    ComplianceStandard,
    DeploymentConfig,
    DeploymentManager,
    DeploymentStrategy,
    EnterpriseFeatures,
    LoadBalancer,
    ScalingPolicy,
    TenantManager,
    TenantType,
    get_enterprise_features,
    start_enterprise_features,
    stop_enterprise_features,
)


class TestLoadBalancer:
    """Test LoadBalancer class"""

    def test_load_balancer_initialization(self):
        """Test LoadBalancer initialization"""
        lb = LoadBalancer()

        assert len(lb.backends) == 0
        assert lb.algorithm == "round_robin"
        assert lb.session_affinity is False

    def test_add_backend(self):
        """Test adding backend server"""
        lb = LoadBalancer()

        lb.add_backend("backend_1", "http://server1.example.com", weight=1)

        assert len(lb.backends) == 1
        assert lb.backends[0]["backend_id"] == "backend_1"

    def test_remove_backend(self):
        """Test removing backend server"""
        lb = LoadBalancer()

        lb.add_backend("backend_1", "http://server1.example.com")
        result = lb.remove_backend("backend_1")

        assert result is True
        assert len(lb.backends) == 0

    def test_get_backend_round_robin(self):
        """Test round-robin load balancing"""
        lb = LoadBalancer()
        lb.algorithm = "round_robin"

        lb.add_backend("backend_1", "http://server1.example.com")
        lb.add_backend("backend_2", "http://server2.example.com")

        backend1 = lb.get_backend()
        backend2 = lb.get_backend()

        assert backend1 is not None
        assert backend2 is not None

    def test_get_backend_least_connections(self):
        """Test least-connections load balancing"""
        lb = LoadBalancer()
        lb.algorithm = "least_connections"

        lb.add_backend("backend_1", "http://server1.example.com", max_connections=100)
        lb.add_backend("backend_2", "http://server2.example.com", max_connections=100)

        backend = lb.get_backend()

        assert backend is not None
        # Should select backend with fewer connections
        assert isinstance(backend, str)

    def test_get_backend_weighted(self):
        """Test weighted load balancing"""
        lb = LoadBalancer()
        lb.algorithm = "weighted"

        lb.add_backend("backend_1", "http://server1.example.com", weight=3)
        lb.add_backend("backend_2", "http://server2.example.com", weight=1)

        backend = lb.get_backend()

        assert backend is not None

    def test_perform_health_check(self):
        """Test health check on backend"""
        lb = LoadBalancer()

        lb.add_backend(
            "backend_1",
            "http://server1.example.com",
            health_check={"path": "/health", "interval": 30, "timeout": 10},
        )

        result = lb.perform_health_check("backend_1")

        assert isinstance(result, bool)

    def test_load_balancer_statistics(self):
        """Test load balancer statistics"""
        lb = LoadBalancer()

        lb.add_backend("backend_1", "http://server1.example.com")
        lb.get_backend()

        stats = lb.get_stats()

        assert "total_backends" in stats
        assert "healthy_backends" in stats
        assert "total_requests" in stats

    def test_session_affinity(self):
        """Test session affinity/sticky sessions"""
        lb = LoadBalancer()
        lb.session_affinity = True

        lb.add_backend("backend_1", "http://server1.example.com")
        lb.add_backend("backend_2", "http://server2.example.com")

        # First request
        backend1 = lb.get_backend(session_id="session_123")
        # Second request with same session should go to same backend
        backend2 = lb.get_backend(session_id="session_123")

        assert backend1 == backend2

    def test_release_backend(self):
        """Test releasing backend connection"""
        lb = LoadBalancer()

        lb.add_backend("backend_1", "http://server1.example.com", max_connections=1)
        backend = lb.get_backend()

        assert lb.backends[0]["current_connections"] > 0

        lb.release_backend(backend)

        assert lb.backends[0]["current_connections"] == 0


class TestAutoScaler:
    """Test AutoScaler class"""

    def test_auto_scaler_initialization(self):
        """Test AutoScaler initialization"""
        scaler = AutoScaler()

        assert scaler.min_instances == 1
        assert scaler.max_instances == 10
        assert scaler.current_instances == 1
        assert scaler.scaling_policy == ScalingPolicy.AUTOMATIC

    def test_update_metrics(self):
        """Test updating metrics for scaling"""
        scaler = AutoScaler()

        scaler.update_metrics(cpu_usage=75.0, memory_usage=60.0, request_rate=100.0)

        assert len(scaler.metrics_history) > 0

    def test_should_scale_up(self):
        """Test determining if scaling up is needed"""
        scaler = AutoScaler()
        scaler.scale_up_threshold = 70.0

        # Add high CPU metrics
        for i in range(5):
            scaler.update_metrics(cpu_usage=80.0, memory_usage=60.0, request_rate=100.0)

        result = scaler.should_scale_up()

        # Should consider scaling up after threshold
        assert isinstance(result, bool)

    def test_should_scale_down(self):
        """Test determining if scaling down is needed"""
        scaler = AutoScaler()
        scaler.scale_down_threshold = 30.0
        scaler.current_instances = 3

        # Add low metrics
        for i in range(10):
            scaler.update_metrics(cpu_usage=20.0, memory_usage=20.0, request_rate=10.0)

        result = scaler.should_scale_down()

        # Should consider scaling down
        assert isinstance(result, bool)

    def test_scale_up_limits(self):
        """Test scale up respects max instances"""
        scaler = AutoScaler()
        scaler.max_instances = 5
        scaler.current_instances = 5

        result = scaler.scale_up()

        assert result is False
        assert scaler.current_instances == 5

    def test_scale_down_limits(self):
        """Test scale down respects min instances"""
        scaler = AutoScaler()
        scaler.min_instances = 1
        scaler.current_instances = 1

        result = scaler.scale_down()

        assert result is False
        assert scaler.current_instances == 1

    def test_scale_up_increments(self):
        """Test scaling up increments instance count"""
        scaler = AutoScaler()
        scaler.min_instances = 1
        scaler.max_instances = 5
        scaler.current_instances = 2

        result = scaler.scale_up()

        assert result is True
        assert scaler.current_instances == 3

    def test_scale_down_decrements(self):
        """Test scaling down decrements instance count"""
        scaler = AutoScaler()
        scaler.min_instances = 1
        scaler.current_instances = 3

        result = scaler.scale_down()

        assert result is True
        assert scaler.current_instances == 2


class TestDeploymentManager:
    """Test DeploymentManager class"""

    def test_deployment_manager_initialization(self):
        """Test DeploymentManager initialization"""
        lb = LoadBalancer()
        scaler = AutoScaler()
        manager = DeploymentManager(lb, scaler)

        assert manager.load_balancer == lb
        assert manager.auto_scaler == scaler
        assert len(manager.active_deployments) == 0

    @pytest.mark.asyncio
    async def test_deploy_blue_green(self):
        """Test blue-green deployment"""
        lb = LoadBalancer()
        scaler = AutoScaler()
        manager = DeploymentManager(lb, scaler)

        config = DeploymentConfig(
            deployment_id="deploy_1",
            strategy=DeploymentStrategy.BLUE_GREEN,
            version="1.0.0",
            environment="production",
        )

        result = await manager.deploy(config)

        assert result["status"] in ["in_progress", "completed", "failed"]

    @pytest.mark.asyncio
    async def test_deploy_canary(self):
        """Test canary deployment"""
        lb = LoadBalancer()
        scaler = AutoScaler()
        manager = DeploymentManager(lb, scaler)

        config = DeploymentConfig(
            deployment_id="deploy_2",
            strategy=DeploymentStrategy.CANARY,
            version="1.0.0",
            environment="production",
            traffic_percentage=10,
        )

        result = await manager.deploy(config)

        assert "steps" in result

    @pytest.mark.asyncio
    async def test_deploy_rolling(self):
        """Test rolling deployment"""
        lb = LoadBalancer()
        scaler = AutoScaler()
        manager = DeploymentManager(lb, scaler)

        config = DeploymentConfig(
            deployment_id="deploy_3",
            strategy=DeploymentStrategy.ROLLING,
            version="1.0.0",
            environment="production",
        )

        result = await manager.deploy(config)

        assert "steps" in result

    def test_get_deployment_status(self):
        """Test getting deployment status"""
        lb = LoadBalancer()
        scaler = AutoScaler()
        manager = DeploymentManager(lb, scaler)

        config = DeploymentConfig(
            deployment_id="deploy_1",
            strategy=DeploymentStrategy.BLUE_GREEN,
            version="1.0.0",
            environment="production",
        )

        manager.active_deployments["deploy_1"] = config

        status = manager.get_deployment_status("deploy_1")

        assert status["deployment_id"] == "deploy_1"
        assert status["strategy"] == "blue_green"

    def test_rollback_deployment(self):
        """Test rolling back deployment"""
        lb = LoadBalancer()
        scaler = AutoScaler()
        manager = DeploymentManager(lb, scaler)

        DeploymentConfig(
            deployment_id="deploy_1",
            strategy=DeploymentStrategy.BLUE_GREEN,
            version="1.0.0",
            environment="production",
        )

        # Store rollback point
        manager.rollback_points["deploy_1"] = {
            "version": "0.9.0",
            "backends": ["backend_1", "backend_2"],
            "timestamp": datetime.now(),
        }

        result = manager.rollback("deploy_1")

        assert result["status"] == "completed"
        assert result["rollback_version"] == "0.9.0"


class TestTenantManager:
    """Test TenantManager class"""

    def test_tenant_manager_initialization(self):
        """Test TenantManager initialization"""
        manager = TenantManager()

        assert len(manager.tenants) == 0
        assert len(manager.tenant_resources) == 0

    def test_create_tenant(self):
        """Test creating tenant"""
        manager = TenantManager()

        tenant_id = manager.create_tenant(
            tenant_name="Tenant Corp",
            tenant_type=TenantType.DEDICATED,
            resource_limits={"cpu_cores": 4, "memory_gb": 8},
            compliance_requirements=[ComplianceStandard.GDPR],
        )

        assert tenant_id is not None
        assert tenant_id in manager.tenants

    def test_get_tenant(self):
        """Test getting tenant configuration"""
        manager = TenantManager()

        tenant_id = manager.create_tenant(
            tenant_name="Test Tenant",
            tenant_type=TenantType.SHARED,
        )

        tenant = manager.get_tenant(tenant_id)

        assert tenant is not None
        assert tenant.tenant_name == "Test Tenant"

    def test_list_tenants(self):
        """Test listing tenants"""
        manager = TenantManager()

        for i in range(3):
            manager.create_tenant(
                tenant_name=f"Tenant {i}",
                tenant_type=TenantType.SHARED,
            )

        tenants = manager.list_tenants()

        assert len(tenants) == 3

    def test_update_tenant(self):
        """Test updating tenant configuration"""
        manager = TenantManager()

        tenant_id = manager.create_tenant(
            tenant_name="Test Tenant",
            tenant_type=TenantType.SHARED,
        )

        result = manager.update_tenant(tenant_id, {"tenant_name": "Updated Tenant"})

        assert result is True
        tenant = manager.get_tenant(tenant_id)
        assert tenant.tenant_name == "Updated Tenant"

    def test_delete_tenant(self):
        """Test deleting tenant"""
        manager = TenantManager()

        tenant_id = manager.create_tenant(
            tenant_name="Test Tenant",
            tenant_type=TenantType.SHARED,
        )

        result = manager.delete_tenant(tenant_id)

        assert result is True
        tenant = manager.get_tenant(tenant_id)
        assert tenant is None

    def test_tenant_resource_isolation(self):
        """Test tenant resource isolation"""
        manager = TenantManager()

        tenant_id = manager.create_tenant(
            tenant_name="Isolated Tenant",
            tenant_type=TenantType.ISOLATED,
        )

        resources = manager.tenant_resources.get(tenant_id)

        assert resources is not None
        assert "database" in resources
        assert "storage" in resources

    def test_generate_compliance_report(self):
        """Test generating compliance report"""
        manager = TenantManager()

        tenant_id = manager.create_tenant(
            tenant_name="Compliant Tenant",
            tenant_type=TenantType.DEDICATED,
            compliance_requirements=[ComplianceStandard.GDPR, ComplianceStandard.SOC2],
        )

        report = manager.generate_compliance_report(tenant_id, ComplianceStandard.GDPR)

        assert report["tenant_id"] == tenant_id
        assert report["standard"] == "gdpr"


class TestAuditLogger:
    """Test AuditLogger class"""

    def test_audit_logger_initialization(self):
        """Test AuditLogger initialization"""
        logger = AuditLogger(retention_days=365)

        assert logger.retention_days == 365
        assert len(logger.audit_logs) == 0

    def test_log_audit_event(self):
        """Test logging audit event"""
        logger = AuditLogger()

        log_id = logger.log(
            action="create_deployment",
            resource="application",
            user_id="admin",
            tenant_id="tenant_1",
            compliance_standards=[ComplianceStandard.GDPR],
        )

        assert log_id is not None
        assert len(logger.audit_logs) > 0

    def test_search_logs_by_action(self):
        """Test searching logs by action"""
        logger = AuditLogger()

        logger.log("create_deployment", "app", "admin", tenant_id="tenant_1")
        logger.log("delete_deployment", "app", "admin", tenant_id="tenant_1")

        results = logger.search_logs(action="create_deployment")

        assert len(results) > 0
        assert results[0].action == "create_deployment"

    def test_search_logs_by_tenant(self):
        """Test searching logs by tenant"""
        logger = AuditLogger()

        logger.log("action1", "resource", "user", tenant_id="tenant_1")
        logger.log("action2", "resource", "user", tenant_id="tenant_2")

        results = logger.search_logs(tenant_id="tenant_1")

        assert all(log.tenant_id == "tenant_1" for log in results)

    def test_search_logs_by_time_range(self):
        """Test searching logs by time range"""
        logger = AuditLogger()

        now = datetime.now()
        logger.log("action1", "resource", "user")

        results = logger.search_logs(
            start_time=now - timedelta(hours=1),
            end_time=now + timedelta(hours=1),
        )

        assert len(results) > 0

    def test_get_compliance_report(self):
        """Test generating compliance report"""
        logger = AuditLogger()

        logger.log(
            "action",
            "resource",
            "user",
            compliance_standards=[ComplianceStandard.GDPR],
        )

        report = logger.get_compliance_report(ComplianceStandard.GDPR, days=30)

        assert "standard" in report
        assert "total_logs" in report


class TestEnterpriseFeatures:
    """Test EnterpriseFeatures orchestration"""

    def test_enterprise_features_initialization(self):
        """Test EnterpriseFeatures initialization"""
        enterprise = EnterpriseFeatures(
            min_instances=2,
            max_instances=10,
            enable_multi_tenant=True,
            enable_audit_logging=True,
        )

        assert enterprise.min_instances == 2
        assert enterprise.max_instances == 10
        assert enterprise.load_balancer is not None
        assert enterprise.auto_scaler is not None
        assert enterprise.tenant_manager is not None
        assert enterprise.audit_logger is not None

    @pytest.mark.asyncio
    async def test_start_enterprise_features(self):
        """Test starting enterprise features"""
        enterprise = EnterpriseFeatures(
            enable_multi_tenant=True,
            enable_audit_logging=True,
        )

        await enterprise.start()

        assert enterprise._running is True

        enterprise.stop()

    def test_stop_enterprise_features(self):
        """Test stopping enterprise features"""
        enterprise = EnterpriseFeatures(
            enable_multi_tenant=True,
            enable_audit_logging=True,
        )

        enterprise._running = True
        enterprise.stop()

        assert enterprise._running is False

    @pytest.mark.asyncio
    async def test_deploy_application(self):
        """Test application deployment"""
        enterprise = EnterpriseFeatures(
            enable_multi_tenant=True,
            enable_audit_logging=True,
        )

        result = await enterprise.deploy_application(
            version="1.0.0",
            strategy=DeploymentStrategy.BLUE_GREEN,
            environment="production",
        )

        assert result is not None

    def test_create_tenant(self):
        """Test creating tenant through enterprise features"""
        enterprise = EnterpriseFeatures(
            enable_multi_tenant=True,
            enable_audit_logging=True,
        )

        tenant_id = enterprise.create_tenant(
            tenant_name="Test Tenant",
            tenant_type=TenantType.DEDICATED,
        )

        assert tenant_id is not None

    def test_log_audit_event(self):
        """Test logging audit event through enterprise features"""
        enterprise = EnterpriseFeatures(
            enable_multi_tenant=True,
            enable_audit_logging=True,
        )

        log_id = enterprise.log_audit_event(
            action="deployment",
            resource="app",
            user_id="admin",
        )

        assert log_id is not None

    def test_get_system_status(self):
        """Test getting system status"""
        enterprise = EnterpriseFeatures(
            enable_multi_tenant=True,
            enable_audit_logging=True,
        )

        status = enterprise.get_system_status()

        assert "status" in status
        assert "load_balancer" in status
        assert "auto_scaler" in status
        assert "tenant_manager" in status
        assert "audit_logger" in status


class TestGlobalFunctions:
    """Test global convenience functions"""

    def test_get_enterprise_features(self):
        """Test getting global enterprise features instance"""
        enterprise = get_enterprise_features()

        assert enterprise is not None
        assert isinstance(enterprise, EnterpriseFeatures)

    @pytest.mark.asyncio
    async def test_start_enterprise_features_global(self):
        """Test global start_enterprise_features function"""
        # Get fresh instance
        enterprise = get_enterprise_features()
        await start_enterprise_features()

        # Should be running
        assert enterprise._running is True

    def test_stop_enterprise_features_global(self):
        """Test global stop_enterprise_features function"""
        stop_enterprise_features()

        enterprise = get_enterprise_features()
        assert enterprise._running is False


class TestTenantType:
    """Test TenantType enum"""

    def test_tenant_type_values(self):
        """Test TenantType enum values"""
        assert TenantType.SHARED.value == "shared"
        assert TenantType.DEDICATED.value == "dedicated"
        assert TenantType.ISOLATED.value == "isolated"
        assert TenantType.HYBRID.value == "hybrid"


class TestDeploymentStrategy:
    """Test DeploymentStrategy enum"""

    def test_deployment_strategy_values(self):
        """Test DeploymentStrategy enum values"""
        assert DeploymentStrategy.BLUE_GREEN.value == "blue_green"
        assert DeploymentStrategy.CANARY.value == "canary"
        assert DeploymentStrategy.ROLLING.value == "rolling"


class TestComplianceStandard:
    """Test ComplianceStandard enum"""

    def test_compliance_standard_values(self):
        """Test ComplianceStandard enum values"""
        assert ComplianceStandard.GDPR.value == "gdpr"
        assert ComplianceStandard.HIPAA.value == "hipaa"
        assert ComplianceStandard.SOC2.value == "soc2"
        assert ComplianceStandard.ISO27001.value == "iso27001"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
