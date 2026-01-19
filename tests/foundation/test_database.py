"""
Comprehensive TDD tests for database.py module.
Tests cover all 7 classes.
"""


from moai_adk.foundation.database import (
    ACIDCompliance,
    ConnectionPoolManager,
    DatabaseSelector,
    IndexingOptimizer,
    IndexRecommendation,
    MigrationPlan,
    MigrationPlanner,
    PerformanceMonitor,
    PoolConfiguration,
    SchemaNormalizer,
    TransactionManager,
    ValidationResult,
)

# ============================================================================
# Test SchemaNormalizer
# ============================================================================


class TestSchemaNormalizer:
    """Test suite for SchemaNormalizer class."""

    def test_validate_1nf_valid(self):
        """Test 1NF validation with valid schema."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "id": "INTEGER",
                "name": "VARCHAR(100)",
                "email": "VARCHAR(255)",
                "PRIMARY KEY": "id"
            }
        }
        result = normalizer.validate_1nf(schema)

        assert result["is_valid"] is True
        assert result["normalization_level"] == "1NF"
        assert len(result["violations"]) == 0

    def test_validate_1nf_with_repeating_groups(self):
        """Test 1NF validation failure with repeating groups."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "id": "INTEGER",
                "name": "VARCHAR(100)",
                "email_list": "VARCHAR(255)",  # Multi-value attribute
                "PRIMARY KEY": "id"
            }
        }
        result = normalizer.validate_1nf(schema)

        assert result["is_valid"] is False
        assert result["normalization_level"] == "0NF"
        assert len(result["violations"]) > 0
        assert "email_list" in result["violations"][0]

    def test_validate_2nf_valid(self):
        """Test 2NF validation with valid schema."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "id": "INTEGER",
                "name": "VARCHAR(100)",
                "email": "VARCHAR(255)",
                "PRIMARY KEY": "id"
            }
        }
        result = normalizer.validate_2nf(schema)

        assert result["is_valid"] is True
        assert result["normalization_level"] == "2NF"

    def test_validate_2nf_with_partial_dependencies(self):
        """Test 2NF validation failure with partial dependencies."""
        normalizer = SchemaNormalizer()
        schema = {
            "order_items": {
                "order_id": "INTEGER",
                "product_id": "INTEGER",
                "product_name": "VARCHAR(100)",  # Partial dependency on product_id
                "quantity": "INTEGER",
                "PRIMARY KEY": "(order_id, product_id)"
            }
        }
        result = normalizer.validate_2nf(schema)

        assert result["is_valid"] is False
        assert result["normalization_level"] == "1NF"
        assert len(result["violations"]) > 0

    def test_validate_3nf_valid(self):
        """Test 3NF validation with valid schema."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "id": "INTEGER",
                "name": "VARCHAR(100)",
                "email": "VARCHAR(255)",
                "PRIMARY KEY": "id"
            }
        }
        result = normalizer.validate_3nf(schema)

        assert result["is_valid"] is True
        assert result["normalization_level"] == "3NF"

    def test_validate_3nf_with_transitive_dependencies(self):
        """Test 3NF validation failure with transitive dependencies."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "id": "INTEGER",
                "name": "VARCHAR(100)",
                "customer_id": "INTEGER",  # Foreign key
                "customer_name": "VARCHAR(100)",  # Transitive: depends on customer_id
                "customer_email": "VARCHAR(255)",  # Transitive: depends on customer_id
                "PRIMARY KEY": "id"
            }
        }
        result = normalizer.validate_3nf(schema)

        assert result["is_valid"] is False
        assert result["normalization_level"] == "2NF"
        assert len(result["violations"]) > 0

    def test_recommend_normalization(self):
        """Test normalization recommendations."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "id": "INTEGER",
                "name": "VARCHAR(100)",
                "email_list": "VARCHAR(255)",  # Violates 1NF
                "city": "VARCHAR(100)",
                "city_state": "VARCHAR(50)",  # Violates 3NF
                "PRIMARY KEY": "id"
            }
        }
        recommendations = normalizer.recommend_normalization(schema)

        assert isinstance(recommendations, list)
        assert len(recommendations) > 0


# ============================================================================
# Test DatabaseSelector
# ============================================================================


class TestDatabaseSelector:
    """Test suite for DatabaseSelector class."""

    def test_select_database_for_acid_compliance(self):
        """Test database selection for ACID compliance."""
        selector = DatabaseSelector()
        result = selector.select_database(
            requirements={"acid_compliance": True}
        )

        assert result["database"] == "PostgreSQL"
        assert result["version"] == "17"
        assert "ACID" in result["reasoning"]
        assert isinstance(result["alternatives"], list)

    def test_select_database_for_flexible_schema(self):
        """Test database selection for flexible schema."""
        selector = DatabaseSelector()
        result = selector.select_database(
            requirements={"schema_flexibility": "high"}
        )

        assert result["database"] == "MongoDB"
        assert result["version"] == "8.0"
        assert "flexible" in result["reasoning"].lower()

    def test_select_database_for_caching(self):
        """Test database selection for caching."""
        selector = DatabaseSelector()
        result = selector.select_database(
            requirements={"use_case": "caching"}
        )

        assert result["database"] == "Redis"
        assert result["version"] == "7.4"
        assert "cache" in result["reasoning"].lower()

    def test_select_database_for_legacy_support(self):
        """Test database selection for legacy support."""
        selector = DatabaseSelector()
        result = selector.select_database(
            requirements={"legacy_support": True}
        )

        assert result["database"] == "MySQL"
        assert result["version"] == "8.4"
        assert "legacy" in result["reasoning"].lower()

    def test_select_database_default(self):
        """Test default database selection."""
        selector = DatabaseSelector()
        result = selector.select_database(requirements={})

        assert result["database"] == "PostgreSQL"
        assert result["version"] == "17"


# ============================================================================
# Test IndexingOptimizer
# ============================================================================


class TestIndexingOptimizer:
    """Test suite for IndexingOptimizer class."""

    def test_recommend_index_for_equality(self):
        """Test index recommendation for equality queries."""
        optimizer = IndexingOptimizer()
        result = optimizer.recommend_index(
            query_pattern={
                "columns": ["email"],
                "conditions": ["email = 'test@example.com'"]
            }
        )

        assert result["index_type"] == "HASH"
        assert result["columns"] == ["email"]
        assert "equality" in result["reasoning"].lower()

    def test_recommend_index_for_range(self):
        """Test index recommendation for range queries."""
        optimizer = IndexingOptimizer()
        result = optimizer.recommend_index(
            query_pattern={
                "columns": ["created_at"],
                "conditions": ["created_at > '2024-01-01'"]
            }
        )

        assert result["index_type"] == "BTREE"
        assert result["columns"] == ["created_at"]
        assert "range" in result["reasoning"].lower()

    def test_recommend_index_for_composite(self):
        """Test composite index recommendation."""
        optimizer = IndexingOptimizer()
        result = optimizer.recommend_index(
            query_pattern={
                "columns": ["user_id", "created_at"],
                "conditions": ["user_id = 1", "created_at > '2024-01-01'"]
            }
        )

        assert result["index_type"] == "COMPOSITE"
        assert len(result["columns"]) == 2
        assert result["columns"][0] == "user_id"  # Equality column first

    def test_recommend_index_default(self):
        """Test default index recommendation."""
        optimizer = IndexingOptimizer()
        result = optimizer.recommend_index(
            query_pattern={
                "columns": ["name"],
                "conditions": []
            }
        )

        assert result["index_type"] == "BTREE"

    def test_detect_redundant_indexes(self):
        """Test redundant index detection."""
        optimizer = IndexingOptimizer()
        existing_indexes = [
            {"name": "idx_user_email", "columns": ["email"]},
            {"name": "idx_user_email_name", "columns": ["email", "name"]},
            {"name": "idx_user_email_dup", "columns": ["email"]}
        ]
        redundant = optimizer.detect_redundant_indexes(existing_indexes)

        assert len(redundant) > 0
        assert any("Redundant with composite" in r["reason"] for r in redundant)

    def test_detect_redundant_indexes_no_duplicates(self):
        """Test redundant index detection with no duplicates."""
        optimizer = IndexingOptimizer()
        existing_indexes = [
            {"name": "idx_user_email", "columns": ["email"]},
            {"name": "idx_user_name", "columns": ["name"]}
        ]
        redundant = optimizer.detect_redundant_indexes(existing_indexes)

        assert len(redundant) == 0


# ============================================================================
# Test ConnectionPoolManager
# ============================================================================


class TestConnectionPoolManager:
    """Test suite for ConnectionPoolManager class."""

    def test_calculate_optimal_pool_size_default(self):
        """Test pool size calculation with defaults."""
        manager = ConnectionPoolManager()
        config = manager.calculate_optimal_pool_size(server_config={})

        assert "min_size" in config
        assert "max_size" in config
        assert config["min_size"] >= 5
        assert config["max_size"] > config["min_size"]

    def test_calculate_optimal_pool_size_custom(self):
        """Test pool size calculation with custom config."""
        manager = ConnectionPoolManager()
        config = manager.calculate_optimal_pool_size(
            server_config={
                "cpu_cores": 8,
                "max_connections": 200,
                "expected_concurrency": 50
            }
        )

        assert config["min_size"] == 16  # 8 * 2
        assert config["max_size"] <= 160  # 200 * 0.8

    def test_monitor_pool_health_healthy(self):
        """Test pool health monitoring for healthy pool."""
        manager = ConnectionPoolManager()
        stats = {
            "active_connections": 10,
            "idle_connections": 20,
            "max_connections": 100,
            "wait_time_avg_ms": 50
        }
        health = manager.monitor_pool_health(stats)

        assert health["is_saturated"] is False
        assert health["saturation_level"] < 0.5
        assert len(health["warnings"]) == 0

    def test_monitor_pool_health_saturated(self):
        """Test pool health monitoring for saturated pool."""
        manager = ConnectionPoolManager()
        stats = {
            "active_connections": 90,
            "idle_connections": 5,
            "max_connections": 100,
            "wait_time_avg_ms": 150
        }
        health = manager.monitor_pool_health(stats)

        assert health["is_saturated"] is True
        assert health["saturation_level"] >= 0.90
        assert len(health["warnings"]) > 0

    def test_recommend_adjustments_increase(self):
        """Test recommendation to increase pool size."""
        manager = ConnectionPoolManager()
        current_config = {"max_size": 20}
        metrics = {"avg_wait_time_ms": 250, "saturation_events_per_hour": 15}

        recommendation = manager.recommend_adjustments(current_config, metrics)

        assert recommendation["suggested_max_size"] > current_config["max_size"]
        assert recommendation["priority"] == "high"

    def test_recommend_adjustments_decrease(self):
        """Test recommendation to decrease pool size."""
        manager = ConnectionPoolManager()
        current_config = {"max_size": 100}
        metrics = {"avg_wait_time_ms": 10, "idle_time_percent": 85}

        recommendation = manager.recommend_adjustments(current_config, metrics)

        assert recommendation["suggested_max_size"] < current_config["max_size"]
        assert recommendation["priority"] == "low"


# ============================================================================
# Test MigrationPlanner
# ============================================================================


class TestMigrationPlanner:
    """Test suite for MigrationPlanner class."""

    def test_generate_migration_plan_add_column(self):
        """Test migration plan for adding column."""
        planner = MigrationPlanner()
        plan = planner.generate_migration_plan(
            change_request={
                "operation": "add_column",
                "table": "users",
                "column": {
                    "name": "email",
                    "type": "VARCHAR(255)",
                    "default": "''"
                }
            }
        )

        assert plan["reversible"] is True
        assert len(plan["steps"]) > 0
        assert len(plan["rollback_steps"]) > 0
        assert "estimated_duration" in plan

    def test_generate_migration_plan_drop_column(self):
        """Test migration plan for dropping column."""
        planner = MigrationPlanner()
        plan = planner.generate_migration_plan(
            change_request={
                "operation": "drop_column",
                "table": "users",
                "column": "old_field"
            }
        )

        assert plan["reversible"] is True
        assert any("Backup" in step for step in plan["steps"])

    def test_validate_safety_safe_operation(self):
        """Test safety validation for safe operation."""
        planner = MigrationPlanner()
        migration = {
            "operation": "add_column",
            "table": "users",
            "column": {"name": "email", "type": "VARCHAR(255)"}
        }

        safety = planner.validate_safety(migration)

        assert safety["is_safe"] is True
        assert "risks" in safety

    def test_validate_safety_unsafe_operation(self):
        """Test safety validation for unsafe operation."""
        planner = MigrationPlanner()
        migration = {
            "operation": "drop_column",
            "table": "users",
            "column": "email"
        }

        safety = planner.validate_safety(migration)

        assert safety["is_safe"] is False
        assert safety["requires_backup"] is True

    def test_detect_breaking_changes(self):
        """Test breaking changes detection."""
        planner = MigrationPlanner()
        migration = {
            "operation": "change_column_type",
            "column": "age",
            "old_type": "VARCHAR(10)",
            "new_type": "INTEGER"
        }

        analysis = planner.detect_breaking_changes(migration)

        assert analysis["has_breaking_changes"] is True
        assert analysis["impact_level"] == "high"
        assert len(analysis["changes"]) > 0


# ============================================================================
# Test TransactionManager
# ============================================================================


class TestTransactionManager:
    """Test suite for TransactionManager class."""

    def test_validate_acid_compliance_valid(self):
        """Test ACID compliance validation with valid isolation."""
        manager = TransactionManager()
        config = {
            "isolation_level": "READ_COMMITTED"
        }

        compliance = manager.validate_acid_compliance(config)

        assert compliance["atomicity"] is True
        assert compliance["consistency"] is True
        assert compliance["isolation"] is True
        assert compliance["durability"] is True

    def test_validate_acid_compliance_invalid_isolation(self):
        """Test ACID compliance validation with invalid isolation."""
        manager = TransactionManager()
        config = {
            "isolation_level": "INVALID_LEVEL"
        }

        compliance = manager.validate_acid_compliance(config)

        assert compliance["isolation"] is False

    def test_detect_deadlock_with_cycle(self):
        """Test deadlock detection with circular wait."""
        manager = TransactionManager()
        transactions = [
            {"id": "tx1", "locks": ["resource1"], "waiting_for": "resource2"},
            {"id": "tx2", "locks": ["resource2"], "waiting_for": "resource1"}
        ]

        result = manager.detect_deadlock(transactions)

        assert result["deadlock_detected"] is True
        assert len(result["involved_transactions"]) > 0
        assert result["resolution_strategy"] is not None

    def test_detect_deadlock_no_cycle(self):
        """Test deadlock detection without circular wait."""
        manager = TransactionManager()
        transactions = [
            {"id": "tx1", "locks": ["resource1"], "waiting_for": None},
            {"id": "tx2", "locks": ["resource2"], "waiting_for": None}
        ]

        result = manager.detect_deadlock(transactions)

        assert result["deadlock_detected"] is False
        assert len(result["involved_transactions"]) == 0

    def test_generate_retry_plan_default(self):
        """Test retry plan generation with defaults."""
        manager = TransactionManager()
        config = {}

        plan = manager.generate_retry_plan(config)

        assert "retry_delays" in plan
        assert len(plan["retry_delays"]) == 3  # default max_retries
        assert plan["strategy"] == "exponential_backoff"

    def test_generate_retry_plan_custom(self):
        """Test retry plan generation with custom config."""
        manager = TransactionManager()
        config = {
            "max_retries": 5,
            "initial_backoff_ms": 50,
            "backoff_multiplier": 3.0
        }

        plan = manager.generate_retry_plan(config)

        assert len(plan["retry_delays"]) == 5
        assert plan["retry_delays"][0] == 50
        assert plan["total_max_time_ms"] > 0


# ============================================================================
# Test PerformanceMonitor
# ============================================================================


class TestPerformanceMonitor:
    """Test suite for PerformanceMonitor class."""

    def test_analyze_query_performance_excellent(self):
        """Test query performance analysis for excellent performance."""
        monitor = PerformanceMonitor()
        stats = {
            "avg_execution_time_ms": 50,
            "max_execution_time_ms": 100,
            "call_count": 1000
        }

        analysis = monitor.analyze_query_performance(stats)

        assert analysis["performance_rating"] == "excellent"
        assert analysis["avg_time_ms"] == 50
        assert analysis["optimization_priority"] == "low"

    def test_analyze_query_performance_poor(self):
        """Test query performance analysis for poor performance."""
        monitor = PerformanceMonitor()
        stats = {
            "avg_execution_time_ms": 1500,
            "max_execution_time_ms": 5000,
            "call_count": 100
        }

        analysis = monitor.analyze_query_performance(stats)

        assert analysis["performance_rating"] == "poor"
        assert len(analysis["recommendations"]) > 0
        assert analysis["optimization_priority"] == "high"

    def test_analyze_query_performance_needs_improvement(self):
        """Test query performance analysis for needs improvement."""
        monitor = PerformanceMonitor()
        stats = {
            "avg_execution_time_ms": 600,
            "max_execution_time_ms": 1000,
            "call_count": 500
        }

        analysis = monitor.analyze_query_performance(stats)

        assert analysis["performance_rating"] == "needs_improvement"

    def test_monitor_connection_usage_healthy(self):
        """Test connection usage monitoring for healthy state."""
        monitor = PerformanceMonitor()
        metrics = {
            "active_connections": 30,
            "max_connections": 100,
            "failed_connection_attempts": 0
        }

        health = monitor.monitor_connection_usage(metrics)

        assert health["health_status"] == "healthy"
        assert health["usage_ratio"] < 0.5

    def test_monitor_connection_usage_critical(self):
        """Test connection usage monitoring for critical state."""
        monitor = PerformanceMonitor()
        metrics = {
            "active_connections": 95,
            "max_connections": 100,
            "failed_connection_attempts": 5
        }

        health = monitor.monitor_connection_usage(metrics)

        assert health["health_status"] == "critical"
        assert health["usage_ratio"] > 0.90
        assert len(health["recommendations"]) > 0

    def test_monitor_connection_usage_warning(self):
        """Test connection usage monitoring for warning state."""
        monitor = PerformanceMonitor()
        metrics = {
            "active_connections": 78,
            "max_connections": 100,
            "failed_connection_attempts": 2
        }

        health = monitor.monitor_connection_usage(metrics)

        assert health["health_status"] == "warning"
        assert 0.75 < health["usage_ratio"] < 0.90


# ============================================================================
# Test Data Classes
# ============================================================================


class TestDataClasses:
    """Test suite for database data classes."""

    def test_validation_result_creation(self):
        """Test ValidationResult dataclass creation."""
        result = ValidationResult(
            is_valid=True,
            violations=[],
            normalization_level="3NF",
            suggestions=["No issues found"]
        )

        assert result.is_valid is True
        assert result.normalization_level == "3NF"
        assert len(result.violations) == 0

    def test_index_recommendation_creation(self):
        """Test IndexRecommendation dataclass creation."""
        index = IndexRecommendation(
            index_type="BTREE",
            columns=["email"],
            reasoning="B-tree for range queries",
            estimated_improvement=0.75
        )

        assert index.index_type == "BTREE"
        assert index.columns == ["email"]
        assert index.estimated_improvement == 0.75

    def test_pool_configuration_creation(self):
        """Test PoolConfiguration dataclass creation."""
        config = PoolConfiguration(
            min_size=10,
            max_size=50,
            timeout_seconds=30,
            idle_timeout=600
        )

        assert config.min_size == 10
        assert config.max_size == 50
        assert config.idle_timeout == 600

    def test_migration_plan_creation(self):
        """Test MigrationPlan dataclass creation."""
        plan = MigrationPlan(
            steps=["Add column", "Verify"],
            reversible=True,
            rollback_steps=["Drop column"],
            estimated_duration="5 minutes"
        )

        assert plan.reversible is True
        assert len(plan.steps) == 2
        assert len(plan.rollback_steps) == 1

    def test_acid_compliance_creation(self):
        """Test ACIDCompliance dataclass creation."""
        compliance = ACIDCompliance(
            atomicity=True,
            consistency=True,
            isolation=True,
            durability=True
        )

        assert compliance.atomicity is True
        assert compliance.consistency is True
        assert compliance.isolation is True
        assert compliance.durability is True
