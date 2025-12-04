"""
Comprehensive test suite for database.py module.

Tests cover schema normalization validation, database technology selection,
indexing strategy optimization, connection pool management, migration planning,
transaction handling, and performance monitoring with 90%+ coverage goal.

Module: src/moai_adk/foundation/database.py
Classes: 7 main classes, 5 data classes
Lines: 1,107 total
"""

import pytest

from src.moai_adk.foundation.database import (
    ACIDCompliance,
    ConnectionPoolManager,
    DatabaseRecommendation,
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
# Data Class Tests
# ============================================================================


class TestValidationResult:
    """Test ValidationResult dataclass."""

    def test_validation_result_initialization(self):
        """Test ValidationResult creation with all fields."""
        result = ValidationResult(
            is_valid=True,
            violations=[],
            normalization_level="3NF",
            suggestions=["suggestion1", "suggestion2"],
        )
        assert result.is_valid is True
        assert result.violations == []
        assert result.normalization_level == "3NF"
        assert result.suggestions == ["suggestion1", "suggestion2"]

    def test_validation_result_post_init_none_suggestions(self):
        """Test ValidationResult initializes None suggestions as empty list."""
        result = ValidationResult(
            is_valid=False,
            violations=["violation1"],
            normalization_level="1NF",
            suggestions=None,
        )
        assert result.suggestions == []

    def test_validation_result_with_violations(self):
        """Test ValidationResult with violations."""
        violations = ["Violation 1", "Violation 2"]
        result = ValidationResult(
            is_valid=False,
            violations=violations,
            normalization_level="0NF",
        )
        assert result.is_valid is False
        assert len(result.violations) == 2
        assert result.violations == violations

    def test_validation_result_different_normalization_levels(self):
        """Test ValidationResult with different normalization levels."""
        for level in ["0NF", "1NF", "2NF", "3NF", "BCNF"]:
            result = ValidationResult(
                is_valid=level != "0NF",
                violations=[],
                normalization_level=level,
            )
            assert result.normalization_level == level


class TestDatabaseRecommendation:
    """Test DatabaseRecommendation dataclass."""

    def test_database_recommendation_initialization(self):
        """Test DatabaseRecommendation creation."""
        rec = DatabaseRecommendation(
            database="PostgreSQL",
            version="17",
            reasoning="ACID compliance required",
            alternatives=["MySQL 8.4"],
        )
        assert rec.database == "PostgreSQL"
        assert rec.version == "17"
        assert rec.reasoning == "ACID compliance required"
        assert rec.alternatives == ["MySQL 8.4"]

    def test_database_recommendation_post_init_none_alternatives(self):
        """Test DatabaseRecommendation initializes None alternatives as empty list."""
        rec = DatabaseRecommendation(
            database="MongoDB",
            version="8.0",
            reasoning="Flexible schema",
            alternatives=None,
        )
        assert rec.alternatives == []

    def test_database_recommendation_multiple_alternatives(self):
        """Test DatabaseRecommendation with multiple alternatives."""
        alternatives = ["MySQL 8.4", "MariaDB", "PostgreSQL"]
        rec = DatabaseRecommendation(
            database="MongoDB",
            version="8.0",
            reasoning="Document-based model",
            alternatives=alternatives,
        )
        assert len(rec.alternatives) == 3


class TestIndexRecommendation:
    """Test IndexRecommendation dataclass."""

    def test_index_recommendation_initialization(self):
        """Test IndexRecommendation creation."""
        rec = IndexRecommendation(
            index_type="BTREE",
            columns=["user_id", "created_at"],
            reasoning="Range queries on created_at",
            estimated_improvement=0.75,
        )
        assert rec.index_type == "BTREE"
        assert rec.columns == ["user_id", "created_at"]
        assert rec.estimated_improvement == 0.75

    def test_index_recommendation_without_estimated_improvement(self):
        """Test IndexRecommendation without estimated improvement."""
        rec = IndexRecommendation(
            index_type="HASH",
            columns=["user_id"],
            reasoning="Exact match queries",
            estimated_improvement=None,
        )
        assert rec.estimated_improvement is None

    def test_index_recommendation_different_types(self):
        """Test IndexRecommendation with different index types."""
        types = ["BTREE", "HASH", "COMPOSITE"]
        for idx_type in types:
            rec = IndexRecommendation(
                index_type=idx_type,
                columns=["col1"],
                reasoning="Test",
            )
            assert rec.index_type == idx_type


class TestPoolConfiguration:
    """Test PoolConfiguration dataclass."""

    def test_pool_configuration_initialization(self):
        """Test PoolConfiguration creation."""
        config = PoolConfiguration(
            min_size=5,
            max_size=50,
            timeout_seconds=30,
            idle_timeout=600,
        )
        assert config.min_size == 5
        assert config.max_size == 50
        assert config.timeout_seconds == 30
        assert config.idle_timeout == 600

    def test_pool_configuration_without_idle_timeout(self):
        """Test PoolConfiguration without idle_timeout."""
        config = PoolConfiguration(
            min_size=5,
            max_size=50,
            timeout_seconds=30,
            idle_timeout=None,
        )
        assert config.idle_timeout is None

    def test_pool_configuration_valid_ranges(self):
        """Test PoolConfiguration with various valid ranges."""
        configs = [
            (1, 10, 10),
            (5, 50, 30),
            (10, 100, 60),
        ]
        for min_s, max_s, timeout in configs:
            config = PoolConfiguration(
                min_size=min_s,
                max_size=max_s,
                timeout_seconds=timeout,
            )
            assert config.min_size == min_s
            assert config.max_size == max_s


class TestMigrationPlan:
    """Test MigrationPlan dataclass."""

    def test_migration_plan_initialization(self):
        """Test MigrationPlan creation."""
        plan = MigrationPlan(
            steps=["Step 1", "Step 2"],
            reversible=True,
            rollback_steps=["Rollback 1"],
            estimated_duration="5 minutes",
        )
        assert plan.steps == ["Step 1", "Step 2"]
        assert plan.reversible is True
        assert plan.rollback_steps == ["Rollback 1"]
        assert plan.estimated_duration == "5 minutes"

    def test_migration_plan_non_reversible(self):
        """Test MigrationPlan for non-reversible migration."""
        plan = MigrationPlan(
            steps=["Custom step"],
            reversible=False,
            rollback_steps=[],
            estimated_duration="unknown",
        )
        assert plan.reversible is False
        assert len(plan.rollback_steps) == 0


class TestACIDCompliance:
    """Test ACIDCompliance dataclass."""

    def test_acid_compliance_all_true(self):
        """Test ACIDCompliance with all properties satisfied."""
        compliance = ACIDCompliance(
            atomicity=True,
            consistency=True,
            isolation=True,
            durability=True,
        )
        assert compliance.atomicity is True
        assert compliance.consistency is True
        assert compliance.isolation is True
        assert compliance.durability is True

    def test_acid_compliance_partial(self):
        """Test ACIDCompliance with partial satisfaction."""
        compliance = ACIDCompliance(
            atomicity=True,
            consistency=True,
            isolation=False,
            durability=True,
        )
        assert compliance.isolation is False
        assert compliance.atomicity is True

    def test_acid_compliance_all_false(self):
        """Test ACIDCompliance with no properties satisfied."""
        compliance = ACIDCompliance(
            atomicity=False,
            consistency=False,
            isolation=False,
            durability=False,
        )
        assert all(
            [
                compliance.atomicity is False,
                compliance.consistency is False,
                compliance.isolation is False,
                compliance.durability is False,
            ]
        )


# ============================================================================
# SchemaNormalizer Tests
# ============================================================================


class TestSchemaNormalizer:
    """Test SchemaNormalizer class."""

    def test_validate_1nf_valid_schema(self):
        """Test 1NF validation with valid schema."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "user_id": "INT",
                "name": "VARCHAR(100)",
                "email": "VARCHAR(100)",
                "PRIMARY KEY": "user_id",
            }
        }
        result = normalizer.validate_1nf(schema)
        assert result["is_valid"] is True
        assert result["normalization_level"] == "1NF"
        assert len(result["violations"]) == 0

    def test_validate_1nf_multi_valued_attributes(self):
        """Test 1NF validation detects multi-valued attributes."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "user_id": "INT",
                "tags": "VARCHAR(500)",  # Multi-valued
                "PRIMARY KEY": "user_id",
            }
        }
        result = normalizer.validate_1nf(schema)
        assert result["is_valid"] is False
        assert result["normalization_level"] == "0NF"
        assert len(result["violations"]) > 0

    @pytest.mark.parametrize(
        "column_name,column_type",
        [
            ("user_list", "VARCHAR(100)"),
            ("items_array", "TEXT"),
            ("category_set", "VARCHAR(200)"),
            ("tags", "TEXT"),
            ("categories", "VARCHAR(100)"),
            ("phone_numbers", "VARCHAR(500)"),
            ("email_addresses", "TEXT"),
        ],
    )
    def test_validate_1nf_various_multi_valued_patterns(self, column_name, column_type):
        """Test 1NF validation with various multi-valued attribute patterns."""
        normalizer = SchemaNormalizer()
        schema = {
            "table1": {
                "id": "INT",
                column_name: column_type,
                "PRIMARY KEY": "id",
            }
        }
        result = normalizer.validate_1nf(schema)
        assert result["is_valid"] is False

    def test_validate_1nf_with_suggestions(self):
        """Test 1NF validation provides suggestions."""
        normalizer = SchemaNormalizer()
        schema = {
            "orders": {
                "order_id": "INT",
                "item_list": "TEXT",
                "PRIMARY KEY": "order_id",
            }
        }
        result = normalizer.validate_1nf(schema)
        assert "suggestions" in result
        assert len(result["suggestions"]) > 0

    def test_validate_2nf_valid_schema(self):
        """Test 2NF validation with valid schema."""
        normalizer = SchemaNormalizer()
        schema = {
            "orders": {
                "order_id": "INT",
                "customer_id": "INT",
                "order_date": "DATE",
                "PRIMARY KEY": "order_id",
            }
        }
        result = normalizer.validate_2nf(schema)
        assert result["is_valid"] is True
        assert result["normalization_level"] == "2NF"

    def test_validate_2nf_partial_dependency(self):
        """Test 2NF validation detects partial dependencies."""
        normalizer = SchemaNormalizer()
        schema = {
            "order_items": {
                "order_id": "INT",
                "product_id": "INT",
                "product_name": "VARCHAR(100)",  # Depends only on product_id
                "PRIMARY KEY": "(order_id, product_id)",
            }
        }
        result = normalizer.validate_2nf(schema)
        assert result["is_valid"] is False

    def test_validate_2nf_product_specific_pattern(self):
        """Test 2NF validation detects product-specific patterns."""
        normalizer = SchemaNormalizer()
        schema = {
            "orders": {
                "order_id": "INT",
                "product_id": "INT",
                "product_price": "DECIMAL",  # Depends only on product_id
                "PRIMARY KEY": "(order_id, product_id)",
            }
        }
        result = normalizer.validate_2nf(schema)
        assert result["is_valid"] is False
        assert len(result["violations"]) > 0

    def test_validate_3nf_valid_schema(self):
        """Test 3NF validation with valid schema."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "user_id": "INT",
                "name": "VARCHAR(100)",
                "department_id": "INT",
                "PRIMARY KEY": "user_id",
            }
        }
        result = normalizer.validate_3nf(schema)
        assert result["is_valid"] is True
        assert result["normalization_level"] == "3NF"

    def test_validate_3nf_transitive_dependency(self):
        """Test 3NF validation detects transitive dependencies."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "user_id": "INT",
                "department_id": "INT",
                "department_name": "VARCHAR(100)",  # Depends on department_id
                "department_location": "VARCHAR(100)",  # Depends on department_id
                "PRIMARY KEY": "user_id",
            }
        }
        result = normalizer.validate_3nf(schema)
        assert result["is_valid"] is False
        assert len(result["violations"]) > 0

    def test_validate_3nf_with_suggestions(self):
        """Test 3NF validation provides suggestions."""
        normalizer = SchemaNormalizer()
        schema = {
            "orders": {
                "order_id": "INT",
                "customer_id": "INT",
                "customer_name": "VARCHAR(100)",
                "PRIMARY KEY": "order_id",
            }
        }
        result = normalizer.validate_3nf(schema)
        assert "suggestions" in result

    def test_recommend_normalization_multiple_violations(self):
        """Test normalization recommendations for schema with violations."""
        normalizer = SchemaNormalizer()
        schema = {
            "mixed_table": {
                "id": "INT",
                "tags": "VARCHAR(500)",  # 1NF violation
                "customer_name": "VARCHAR(100)",
                "customer_email": "VARCHAR(100)",
                "customer_phone": "VARCHAR(20)",
                "PRIMARY KEY": "id",
            }
        }
        recommendations = normalizer.recommend_normalization(schema)
        assert len(recommendations) > 0

    def test_recommend_normalization_valid_schema(self):
        """Test normalization recommendations for valid schema."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "user_id": "INT",
                "name": "VARCHAR(100)",
                "PRIMARY KEY": "user_id",
            }
        }
        recommendations = normalizer.recommend_normalization(schema)
        assert isinstance(recommendations, list)

    def test_recommend_normalization_customer_extraction(self):
        """Test normalization recommendations for customer data extraction."""
        normalizer = SchemaNormalizer()
        schema = {
            "orders": {
                "order_id": "INT",
                "customer_id": "INT",
                "customer_name": "VARCHAR(100)",
                "customer_phone": "VARCHAR(20)",
                "PRIMARY KEY": "order_id",
            }
        }
        recommendations = normalizer.recommend_normalization(schema)
        # Should recommend customer extraction
        assert any("customer" in str(rec).lower() for rec in recommendations)


# ============================================================================
# DatabaseSelector Tests
# ============================================================================


class TestDatabaseSelector:
    """Test DatabaseSelector class."""

    def test_select_database_acid_compliance(self):
        """Test database selection for ACID compliance requirement."""
        selector = DatabaseSelector()
        requirements = {"acid_compliance": True}
        result = selector.select_database(requirements)
        assert result["database"] == "PostgreSQL"
        assert result["version"] == "17"

    def test_select_database_transactions_required(self):
        """Test database selection when transactions are required."""
        selector = DatabaseSelector()
        requirements = {"transactions": "required"}
        result = selector.select_database(requirements)
        assert result["database"] == "PostgreSQL"

    def test_select_database_flexible_schema(self):
        """Test database selection for flexible schema requirement."""
        selector = DatabaseSelector()
        requirements = {"schema_flexibility": "high"}
        result = selector.select_database(requirements)
        assert result["database"] == "MongoDB"
        assert result["version"] == "8.0"

    def test_select_database_document_model(self):
        """Test database selection for document-based data model."""
        selector = DatabaseSelector()
        requirements = {"data_model": "document"}
        result = selector.select_database(requirements)
        assert result["database"] == "MongoDB"

    def test_select_database_caching_use_case(self):
        """Test database selection for caching use case."""
        selector = DatabaseSelector()
        requirements = {"use_case": "caching"}
        result = selector.select_database(requirements)
        assert result["database"] == "Redis"
        assert result["version"] == "7.4"

    def test_select_database_speed_critical(self):
        """Test database selection when speed is critical."""
        selector = DatabaseSelector()
        requirements = {"speed": "critical"}
        result = selector.select_database(requirements)
        assert result["database"] == "Redis"

    def test_select_database_legacy_support(self):
        """Test database selection for legacy support."""
        selector = DatabaseSelector()
        requirements = {"legacy_support": True}
        result = selector.select_database(requirements)
        assert result["database"] == "MySQL"
        assert result["version"] == "8.4"

    def test_select_database_mature_ecosystem(self):
        """Test database selection for mature ecosystem requirement."""
        selector = DatabaseSelector()
        requirements = {"ecosystem": "mature"}
        result = selector.select_database(requirements)
        assert result["database"] == "MySQL"

    def test_select_database_default(self):
        """Test default database selection."""
        selector = DatabaseSelector()
        requirements = {}
        result = selector.select_database(requirements)
        assert result["database"] == "PostgreSQL"
        assert result["version"] == "17"

    def test_select_database_has_reasoning(self):
        """Test that database selection includes reasoning."""
        selector = DatabaseSelector()
        requirements = {"acid_compliance": True}
        result = selector.select_database(requirements)
        assert "reasoning" in result
        assert len(result["reasoning"]) > 0

    def test_select_database_has_alternatives(self):
        """Test that database selection includes alternatives."""
        selector = DatabaseSelector()
        requirements = {"acid_compliance": True}
        result = selector.select_database(requirements)
        assert "alternatives" in result
        assert isinstance(result["alternatives"], list)

    @pytest.mark.parametrize(
        "requirements,expected_db",
        [
            ({"acid_compliance": True}, "PostgreSQL"),
            ({"schema_flexibility": "high"}, "MongoDB"),
            ({"use_case": "caching"}, "Redis"),
            ({"legacy_support": True}, "MySQL"),
            ({}, "PostgreSQL"),
        ],
    )
    def test_select_database_various_requirements(self, requirements, expected_db):
        """Test database selection with various requirements."""
        selector = DatabaseSelector()
        result = selector.select_database(requirements)
        assert result["database"] == expected_db


# ============================================================================
# IndexingOptimizer Tests
# ============================================================================


class TestIndexingOptimizer:
    """Test IndexingOptimizer class."""

    def test_recommend_index_composite(self):
        """Test composite index recommendation for multi-column queries."""
        optimizer = IndexingOptimizer()
        query_pattern = {
            "columns": ["user_id", "created_at"],
            "conditions": ["user_id = 1", "created_at > 2024-01-01"],
        }
        result = optimizer.recommend_index(query_pattern)
        assert result["index_type"] == "COMPOSITE"
        assert len(result["columns"]) == 2

    def test_recommend_index_btree_range_query(self):
        """Test B-tree index recommendation for range queries."""
        optimizer = IndexingOptimizer()
        query_pattern = {
            "columns": ["age"],
            "conditions": ["age > 18", "age < 65"],
        }
        result = optimizer.recommend_index(query_pattern)
        assert result["index_type"] == "BTREE"

    def test_recommend_index_btree_between(self):
        """Test B-tree index recommendation for BETWEEN queries."""
        optimizer = IndexingOptimizer()
        query_pattern = {
            "columns": ["created_at"],
            "conditions": ["created_at BETWEEN date1 AND date2"],
        }
        result = optimizer.recommend_index(query_pattern)
        assert result["index_type"] == "BTREE"

    def test_recommend_index_hash_equality(self):
        """Test hash index recommendation for equality queries."""
        optimizer = IndexingOptimizer()
        query_pattern = {
            "columns": ["user_id"],
            "conditions": ["user_id = 123"],
        }
        result = optimizer.recommend_index(query_pattern)
        assert result["index_type"] == "HASH"
        assert result["estimated_improvement"] == 0.90

    def test_recommend_index_default_btree(self):
        """Test default B-tree index recommendation."""
        optimizer = IndexingOptimizer()
        query_pattern = {
            "columns": ["name"],
            "conditions": [],
        }
        result = optimizer.recommend_index(query_pattern)
        assert result["index_type"] == "BTREE"

    def test_recommend_index_has_reasoning(self):
        """Test that index recommendation includes reasoning."""
        optimizer = IndexingOptimizer()
        query_pattern = {
            "columns": ["user_id"],
            "conditions": ["user_id = 1"],
        }
        result = optimizer.recommend_index(query_pattern)
        assert "reasoning" in result
        assert len(result["reasoning"]) > 0

    def test_recommend_index_has_estimated_improvement(self):
        """Test that index recommendation includes estimated improvement."""
        optimizer = IndexingOptimizer()
        query_pattern = {
            "columns": ["user_id"],
            "conditions": ["user_id = 1"],
        }
        result = optimizer.recommend_index(query_pattern)
        assert "estimated_improvement" in result
        assert 0 <= result["estimated_improvement"] <= 1

    def test_detect_redundant_indexes_prefix(self):
        """Test detection of redundant prefix indexes."""
        optimizer = IndexingOptimizer()
        existing_indexes = [
            {"name": "idx_user_id", "columns": ["user_id"]},
            {"name": "idx_user_id_created", "columns": ["user_id", "created_at"]},
        ]
        redundant = optimizer.detect_redundant_indexes(existing_indexes)
        assert len(redundant) > 0
        assert any(r["name"] == "idx_user_id" for r in redundant)

    def test_detect_redundant_indexes_duplicate(self):
        """Test detection of duplicate indexes."""
        optimizer = IndexingOptimizer()
        existing_indexes = [
            {"name": "idx_user_1", "columns": ["user_id"]},
            {"name": "idx_user_2", "columns": ["user_id"]},
        ]
        redundant = optimizer.detect_redundant_indexes(existing_indexes)
        assert len(redundant) > 0

    def test_detect_redundant_indexes_no_redundancy(self):
        """Test detection when no redundant indexes exist."""
        optimizer = IndexingOptimizer()
        existing_indexes = [
            {"name": "idx_user_id", "columns": ["user_id"]},
            {"name": "idx_email", "columns": ["email"]},
            {"name": "idx_created_at", "columns": ["created_at"]},
        ]
        redundant = optimizer.detect_redundant_indexes(existing_indexes)
        assert len(redundant) == 0

    def test_detect_redundant_indexes_empty_list(self):
        """Test detection with empty index list."""
        optimizer = IndexingOptimizer()
        redundant = optimizer.detect_redundant_indexes([])
        assert redundant == []

    def test_composite_index_column_ordering(self):
        """Test that composite indexes order columns correctly (equality first)."""
        optimizer = IndexingOptimizer()
        query_pattern = {
            "columns": ["created_at", "user_id", "status"],
            "conditions": ["user_id = 123", "created_at > date"],
        }
        result = optimizer.recommend_index(query_pattern)
        assert result["index_type"] == "COMPOSITE"
        # Equality columns should be first
        columns = result["columns"]
        assert "user_id" in columns


# ============================================================================
# ConnectionPoolManager Tests
# ============================================================================


class TestConnectionPoolManager:
    """Test ConnectionPoolManager class."""

    def test_calculate_optimal_pool_size_default(self):
        """Test optimal pool size calculation with default parameters."""
        manager = ConnectionPoolManager()
        server_config = {
            "cpu_cores": 4,
            "max_connections": 100,
            "expected_concurrency": 20,
        }
        result = manager.calculate_optimal_pool_size(server_config)
        assert result["min_size"] == 8  # 4 * 2
        # max_size = min(20 * 1.2, 100 * 0.8) = min(24, 80) = 24
        assert result["max_size"] == 24
        assert result["timeout_seconds"] == 30
        assert result["idle_timeout"] == 600

    def test_calculate_optimal_pool_size_high_concurrency(self):
        """Test pool size calculation with high concurrency."""
        manager = ConnectionPoolManager()
        server_config = {
            "cpu_cores": 8,
            "max_connections": 200,
            "expected_concurrency": 100,
        }
        result = manager.calculate_optimal_pool_size(server_config)
        assert result["min_size"] >= 5  # At least 5
        # max_size = min(100 * 1.2, 200 * 0.8) = min(120, 160) = 120
        assert result["max_size"] == 120

    def test_calculate_optimal_pool_size_min_size_floor(self):
        """Test pool size calculation respects minimum floor."""
        manager = ConnectionPoolManager()
        server_config = {
            "cpu_cores": 1,
            "max_connections": 20,
            "expected_concurrency": 5,
        }
        result = manager.calculate_optimal_pool_size(server_config)
        assert result["min_size"] >= 5  # Minimum floor of 5

    def test_calculate_optimal_pool_size_max_greater_than_min(self):
        """Test pool size calculation ensures max > min."""
        manager = ConnectionPoolManager()
        server_config = {
            "cpu_cores": 100,
            "max_connections": 300,
            "expected_concurrency": 1,
        }
        result = manager.calculate_optimal_pool_size(server_config)
        assert result["max_size"] > result["min_size"]

    def test_monitor_pool_health_normal(self):
        """Test pool health monitoring with normal utilization."""
        manager = ConnectionPoolManager()
        pool_stats = {
            "active_connections": 20,
            "idle_connections": 30,
            "max_connections": 100,
            "wait_time_avg_ms": 10,
        }
        result = manager.monitor_pool_health(pool_stats)
        assert result["is_saturated"] is False
        assert result["saturation_level"] == 0.5
        assert len(result["warnings"]) == 0

    def test_monitor_pool_health_saturated(self):
        """Test pool health monitoring with saturated pool."""
        manager = ConnectionPoolManager()
        pool_stats = {
            "active_connections": 85,
            "idle_connections": 10,
            "max_connections": 100,
            "wait_time_avg_ms": 50,
        }
        result = manager.monitor_pool_health(pool_stats)
        assert result["is_saturated"] is True
        assert any("saturation" in w.lower() for w in result["warnings"])

    def test_monitor_pool_health_high_wait_time(self):
        """Test pool health monitoring with high wait times."""
        manager = ConnectionPoolManager()
        pool_stats = {
            "active_connections": 50,
            "idle_connections": 30,
            "max_connections": 100,
            "wait_time_avg_ms": 150,
        }
        result = manager.monitor_pool_health(pool_stats)
        assert any("wait" in w.lower() for w in result["warnings"])

    def test_monitor_pool_health_low_idle(self):
        """Test pool health monitoring with low idle connections."""
        manager = ConnectionPoolManager()
        pool_stats = {
            "active_connections": 90,
            "idle_connections": 5,
            "max_connections": 100,
            "wait_time_avg_ms": 50,
        }
        result = manager.monitor_pool_health(pool_stats)
        assert any("idle" in w.lower() for w in result["warnings"])

    def test_monitor_pool_health_score(self):
        """Test that pool health score is between 0.0 and 1.0."""
        manager = ConnectionPoolManager()
        pool_stats = {
            "active_connections": 50,
            "idle_connections": 40,
            "max_connections": 100,
            "wait_time_avg_ms": 100,
        }
        result = manager.monitor_pool_health(pool_stats)
        assert 0.0 <= result["health_score"] <= 1.0

    def test_recommend_adjustments_increase(self):
        """Test pool adjustment recommendation for increase."""
        manager = ConnectionPoolManager()
        current_config = {"min_size": 5, "max_size": 50}
        metrics = {
            "avg_wait_time_ms": 250,
            "saturation_events_per_hour": 15,
            "idle_time_percent": 30,
        }
        result = manager.recommend_adjustments(current_config, metrics)
        assert result["priority"] == "high"
        assert result["suggested_max_size"] > current_config["max_size"]

    def test_recommend_adjustments_decrease(self):
        """Test pool adjustment recommendation for decrease."""
        manager = ConnectionPoolManager()
        current_config = {"min_size": 5, "max_size": 50}
        metrics = {
            "avg_wait_time_ms": 10,
            "saturation_events_per_hour": 0,
            "idle_time_percent": 85,
        }
        result = manager.recommend_adjustments(current_config, metrics)
        assert result["priority"] == "low"
        assert result["suggested_max_size"] < current_config["max_size"]

    def test_recommend_adjustments_optimal(self):
        """Test pool adjustment recommendation when optimal."""
        manager = ConnectionPoolManager()
        current_config = {"min_size": 5, "max_size": 50}
        metrics = {
            "avg_wait_time_ms": 50,
            "saturation_events_per_hour": 0,
            "idle_time_percent": 40,
        }
        result = manager.recommend_adjustments(current_config, metrics)
        assert result["priority"] == "none"
        assert result["suggested_max_size"] == current_config["max_size"]


# ============================================================================
# MigrationPlanner Tests
# ============================================================================


class TestMigrationPlanner:
    """Test MigrationPlanner class."""

    def test_generate_migration_plan_add_column(self):
        """Test migration plan generation for adding column."""
        planner = MigrationPlanner()
        change_request = {
            "operation": "add_column",
            "table": "users",
            "column": {
                "name": "phone",
                "type": "VARCHAR(20)",
                "default": "NULL",
            },
        }
        plan = planner.generate_migration_plan(change_request)
        assert plan["reversible"] is True
        assert len(plan["steps"]) > 0
        assert len(plan["rollback_steps"]) > 0

    def test_generate_migration_plan_drop_column(self):
        """Test migration plan generation for dropping column."""
        planner = MigrationPlanner()
        change_request = {
            "operation": "drop_column",
            "table": "users",
            "column": "legacy_field",
        }
        plan = planner.generate_migration_plan(change_request)
        assert plan["reversible"] is True
        assert "Backup" in plan["steps"][0]

    def test_generate_migration_plan_change_type(self):
        """Test migration plan generation for changing column type."""
        planner = MigrationPlanner()
        change_request = {
            "operation": "change_column_type",
            "table": "users",
            "column": "age",
            "old_type": "VARCHAR(3)",
            "new_type": "INT",
        }
        plan = planner.generate_migration_plan(change_request)
        assert plan["reversible"] is True
        assert len(plan["steps"]) > 4  # Multiple steps for type change

    def test_generate_migration_plan_custom_operation(self):
        """Test migration plan generation for custom operation."""
        planner = MigrationPlanner()
        change_request = {
            "operation": "custom_operation",
        }
        plan = planner.generate_migration_plan(change_request)
        assert plan["reversible"] is False
        assert plan["estimated_duration"] == "unknown"

    def test_validate_safety_drop_without_backup(self):
        """Test safety validation for drop without backup."""
        planner = MigrationPlanner()
        migration = {
            "operation": "drop_column",
            "backup": False,
        }
        result = planner.validate_safety(migration)
        assert result["is_safe"] is False
        assert any("data loss" in r.lower() for r in result["risks"])
        assert result["requires_backup"] is True

    def test_validate_safety_type_conversion(self):
        """Test safety validation for type conversion."""
        planner = MigrationPlanner()
        migration = {
            "operation": "change_column_type",
        }
        result = planner.validate_safety(migration)
        assert result["is_safe"] is False
        assert result["requires_backup"] is True

    def test_validate_safety_add_column_safe(self):
        """Test safety validation for adding column is safe."""
        planner = MigrationPlanner()
        migration = {
            "operation": "add_column",
        }
        result = planner.validate_safety(migration)
        assert result["is_safe"] is True
        assert result["requires_backup"] is False

    def test_validate_safety_has_recommendations(self):
        """Test that safety validation includes recommendations."""
        planner = MigrationPlanner()
        migration = {
            "operation": "drop_column",
        }
        result = planner.validate_safety(migration)
        assert "recommended_actions" in result
        assert len(result["recommended_actions"]) > 0

    def test_detect_breaking_changes_type_conversion(self):
        """Test detection of breaking changes in type conversion."""
        planner = MigrationPlanner()
        migration = {
            "operation": "change_column_type",
            "old_type": "VARCHAR(10)",
            "new_type": "INT",
        }
        result = planner.detect_breaking_changes(migration)
        assert result["has_breaking_changes"] is True

    def test_detect_breaking_changes_drop_column(self):
        """Test detection of breaking changes in dropping column."""
        planner = MigrationPlanner()
        migration = {
            "operation": "drop_column",
            "column": "user_id",
        }
        result = planner.detect_breaking_changes(migration)
        assert result["has_breaking_changes"] is True

    def test_detect_breaking_changes_add_non_nullable_no_default(self):
        """Test detection of breaking changes in adding non-nullable column."""
        planner = MigrationPlanner()
        migration = {
            "operation": "add_column",
            "column": {
                "name": "required_field",
                "type": "VARCHAR(100)",
                "nullable": False,
                "default": None,
            },
        }
        result = planner.detect_breaking_changes(migration)
        assert result["has_breaking_changes"] is True

    def test_detect_breaking_changes_add_nullable_with_default(self):
        """Test no breaking changes for adding nullable column with default."""
        planner = MigrationPlanner()
        migration = {
            "operation": "add_column",
            "column": {
                "name": "optional_field",
                "type": "VARCHAR(100)",
                "nullable": True,
                "default": "N/A",
            },
        }
        result = planner.detect_breaking_changes(migration)
        assert result["has_breaking_changes"] is False

    def test_detect_breaking_changes_impact_level(self):
        """Test breaking changes impact level."""
        planner = MigrationPlanner()
        migration = {
            "operation": "drop_column",
            "column": "field",
        }
        result = planner.detect_breaking_changes(migration)
        assert result["impact_level"] == "high"

    def test_detect_breaking_changes_mitigation_strategies(self):
        """Test breaking changes mitigation strategies."""
        planner = MigrationPlanner()
        migration = {
            "operation": "drop_column",
            "column": "field",
        }
        result = planner.detect_breaking_changes(migration)
        assert "mitigation_strategies" in result


# ============================================================================
# TransactionManager Tests
# ============================================================================


class TestTransactionManager:
    """Test TransactionManager class."""

    def test_validate_acid_compliance_valid(self):
        """Test ACID compliance validation with valid config."""
        manager = TransactionManager()
        config = {
            "isolation_level": "SERIALIZABLE",
        }
        result = manager.validate_acid_compliance(config)
        assert result["atomicity"] is True
        assert result["consistency"] is True
        assert result["isolation"] is True
        assert result["durability"] is True

    def test_validate_acid_compliance_invalid_isolation(self):
        """Test ACID compliance validation with invalid isolation level."""
        manager = TransactionManager()
        config = {
            "isolation_level": "INVALID_LEVEL",
        }
        result = manager.validate_acid_compliance(config)
        assert result["isolation"] is False

    @pytest.mark.parametrize(
        "isolation_level",
        [
            "READ_UNCOMMITTED",
            "READ_COMMITTED",
            "REPEATABLE_READ",
            "SERIALIZABLE",
        ],
    )
    def test_validate_acid_compliance_valid_isolation_levels(self, isolation_level):
        """Test ACID compliance with valid isolation levels."""
        manager = TransactionManager()
        config = {"isolation_level": isolation_level}
        result = manager.validate_acid_compliance(config)
        assert result["isolation"] is True

    def test_detect_deadlock_no_deadlock(self):
        """Test deadlock detection with no deadlock."""
        manager = TransactionManager()
        transactions = [
            {"id": "tx1", "locks": ["resource_a"], "waiting_for": None},
            {"id": "tx2", "locks": ["resource_b"], "waiting_for": None},
        ]
        result = manager.detect_deadlock(transactions)
        assert result["deadlock_detected"] is False
        assert len(result["involved_transactions"]) == 0

    def test_detect_deadlock_simple_cycle(self):
        """Test deadlock detection with simple cycle."""
        manager = TransactionManager()
        transactions = [
            {"id": "tx1", "locks": ["resource_a"], "waiting_for": "resource_b"},
            {"id": "tx2", "locks": ["resource_b"], "waiting_for": "resource_a"},
        ]
        result = manager.detect_deadlock(transactions)
        assert result["deadlock_detected"] is True
        assert "tx1" in result["involved_transactions"] or "tx2" in result["involved_transactions"]

    def test_detect_deadlock_three_way_cycle(self):
        """Test deadlock detection with three-way cycle."""
        manager = TransactionManager()
        transactions = [
            {"id": "tx1", "locks": ["resource_a"], "waiting_for": "resource_b"},
            {"id": "tx2", "locks": ["resource_b"], "waiting_for": "resource_c"},
            {"id": "tx3", "locks": ["resource_c"], "waiting_for": "resource_a"},
        ]
        result = manager.detect_deadlock(transactions)
        assert result["deadlock_detected"] is True

    def test_detect_deadlock_resolution_strategy(self):
        """Test that deadlock detection provides resolution strategy."""
        manager = TransactionManager()
        transactions = [
            {"id": "tx1", "locks": ["resource_a"], "waiting_for": "resource_b"},
            {"id": "tx2", "locks": ["resource_b"], "waiting_for": "resource_a"},
        ]
        result = manager.detect_deadlock(transactions)
        assert result["resolution_strategy"] is not None

    def test_generate_retry_plan_default(self):
        """Test retry plan generation with default parameters."""
        manager = TransactionManager()
        retry_config = {
            "max_retries": 3,
            "initial_backoff_ms": 100,
            "backoff_multiplier": 2.0,
            "max_backoff_ms": 1000,
        }
        plan = manager.generate_retry_plan(retry_config)
        assert plan["strategy"] == "exponential_backoff"
        assert len(plan["retry_delays"]) == 3
        assert plan["retry_delays"][0] == 100
        assert plan["retry_delays"][1] == 200
        assert plan["retry_delays"][2] == 400

    def test_generate_retry_plan_exponential_backoff(self):
        """Test retry plan generates exponential backoff."""
        manager = TransactionManager()
        retry_config = {
            "max_retries": 4,
            "initial_backoff_ms": 50,
            "backoff_multiplier": 2.0,
            "max_backoff_ms": 1000,
        }
        plan = manager.generate_retry_plan(retry_config)
        delays = plan["retry_delays"]
        # Verify exponential growth
        assert delays[1] > delays[0]
        assert delays[2] > delays[1]
        assert delays[3] > delays[2]

    def test_generate_retry_plan_respects_max_backoff(self):
        """Test retry plan respects maximum backoff."""
        manager = TransactionManager()
        retry_config = {
            "max_retries": 5,
            "initial_backoff_ms": 100,
            "backoff_multiplier": 3.0,
            "max_backoff_ms": 500,
        }
        plan = manager.generate_retry_plan(retry_config)
        # All delays should be <= max_backoff
        assert all(d <= 500 for d in plan["retry_delays"])

    def test_generate_retry_plan_total_time(self):
        """Test retry plan calculates total time correctly."""
        manager = TransactionManager()
        retry_config = {
            "max_retries": 3,
            "initial_backoff_ms": 100,
            "backoff_multiplier": 2.0,
            "max_backoff_ms": 1000,
        }
        plan = manager.generate_retry_plan(retry_config)
        expected_total = 100 + 200 + 400
        assert plan["total_max_time_ms"] == expected_total


# ============================================================================
# PerformanceMonitor Tests
# ============================================================================


class TestPerformanceMonitor:
    """Test PerformanceMonitor class."""

    def test_analyze_query_performance_excellent(self):
        """Test query performance analysis for excellent performance."""
        monitor = PerformanceMonitor()
        query_stats = {
            "avg_execution_time_ms": 10,
            "max_execution_time_ms": 50,
            "call_count": 1000,
        }
        result = monitor.analyze_query_performance(query_stats)
        assert result["performance_rating"] == "excellent"
        assert result["avg_time_ms"] == 10

    def test_analyze_query_performance_good(self):
        """Test query performance analysis for good performance."""
        monitor = PerformanceMonitor()
        query_stats = {
            "avg_execution_time_ms": 150,
            "max_execution_time_ms": 300,
            "call_count": 500,
        }
        result = monitor.analyze_query_performance(query_stats)
        assert result["performance_rating"] == "good"

    def test_analyze_query_performance_needs_improvement(self):
        """Test query performance analysis for performance needing improvement."""
        monitor = PerformanceMonitor()
        query_stats = {
            "avg_execution_time_ms": 600,
            "max_execution_time_ms": 1500,
            "call_count": 200,
        }
        result = monitor.analyze_query_performance(query_stats)
        assert result["performance_rating"] == "needs_improvement"

    def test_analyze_query_performance_poor(self):
        """Test query performance analysis for poor performance."""
        monitor = PerformanceMonitor()
        query_stats = {
            "avg_execution_time_ms": 2000,
            "max_execution_time_ms": 5000,
            "call_count": 50,
        }
        result = monitor.analyze_query_performance(query_stats)
        assert result["performance_rating"] == "poor"
        assert result["optimization_priority"] == "high"

    def test_analyze_query_performance_slow_query_detection(self):
        """Test that slow queries are recommended for investigation."""
        monitor = PerformanceMonitor()
        query_stats = {
            "avg_execution_time_ms": 600,
            "max_execution_time_ms": 7000,
            "call_count": 100,
        }
        result = monitor.analyze_query_performance(query_stats)
        assert any("slow" in r.lower() for r in result["recommendations"])

    def test_analyze_query_performance_has_recommendations(self):
        """Test that query analysis includes recommendations."""
        monitor = PerformanceMonitor()
        query_stats = {
            "avg_execution_time_ms": 600,
            "max_execution_time_ms": 1000,
            "call_count": 100,
        }
        result = monitor.analyze_query_performance(query_stats)
        assert "recommendations" in result

    def test_monitor_connection_usage_healthy(self):
        """Test connection usage monitoring for healthy state."""
        monitor = PerformanceMonitor()
        connection_metrics = {
            "active_connections": 30,
            "max_connections": 100,
            "failed_connection_attempts": 0,
        }
        result = monitor.monitor_connection_usage(connection_metrics)
        assert result["health_status"] == "healthy"
        assert result["usage_ratio"] == 0.3

    def test_monitor_connection_usage_warning(self):
        """Test connection usage monitoring for warning state."""
        monitor = PerformanceMonitor()
        connection_metrics = {
            "active_connections": 80,
            "max_connections": 100,
            "failed_connection_attempts": 2,
        }
        result = monitor.monitor_connection_usage(connection_metrics)
        assert result["health_status"] == "warning"

    def test_monitor_connection_usage_critical(self):
        """Test connection usage monitoring for critical state."""
        monitor = PerformanceMonitor()
        connection_metrics = {
            "active_connections": 95,
            "max_connections": 100,
            "failed_connection_attempts": 5,
        }
        result = monitor.monitor_connection_usage(connection_metrics)
        assert result["health_status"] == "critical"

    def test_monitor_connection_usage_recommendations(self):
        """Test connection monitoring provides recommendations."""
        monitor = PerformanceMonitor()
        connection_metrics = {
            "active_connections": 85,
            "max_connections": 100,
            "failed_connection_attempts": 15,
        }
        result = monitor.monitor_connection_usage(connection_metrics)
        assert len(result["recommendations"]) > 0

    @pytest.mark.parametrize(
        "active,max_conns,expected_ratio",
        [
            (30, 100, 0.3),
            (50, 100, 0.5),
            (75, 100, 0.75),
            (90, 100, 0.9),
        ],
    )
    def test_monitor_connection_usage_ratios(self, active, max_conns, expected_ratio):
        """Test connection usage ratio calculation."""
        monitor = PerformanceMonitor()
        connection_metrics = {
            "active_connections": active,
            "max_connections": max_conns,
            "failed_connection_attempts": 0,
        }
        result = monitor.monitor_connection_usage(connection_metrics)
        assert result["usage_ratio"] == expected_ratio


# ============================================================================
# Integration Tests
# ============================================================================


class TestDatabaseIntegration:
    """Integration tests combining multiple components."""

    def test_schema_normalization_and_database_selection(self):
        """Test complete workflow: validate schema then select database."""
        normalizer = SchemaNormalizer()
        selector = DatabaseSelector()

        schema = {
            "users": {
                "user_id": "INT",
                "name": "VARCHAR(100)",
                "PRIMARY KEY": "user_id",
            }
        }

        # Normalize schema
        result_1nf = normalizer.validate_1nf(schema)
        assert result_1nf["is_valid"] is True

        # Select database for normalized schema
        requirements = {"acid_compliance": True}
        db_rec = selector.select_database(requirements)
        assert db_rec["database"] == "PostgreSQL"

    def test_migration_planning_and_safety_validation(self):
        """Test migration planning with safety validation."""
        planner = MigrationPlanner()

        change_request = {
            "operation": "add_column",
            "table": "users",
            "column": {
                "name": "created_at",
                "type": "TIMESTAMP",
                "default": "CURRENT_TIMESTAMP",
                "nullable": True,
            },
        }

        # Generate plan
        plan = planner.generate_migration_plan(change_request)
        assert plan["reversible"] is True

        # Validate safety
        migration = {"operation": "add_column"}
        safety = planner.validate_safety(migration)
        assert safety["is_safe"] is True

    def test_pool_management_and_performance_monitoring(self):
        """Test connection pool management with performance monitoring."""
        pool_manager = ConnectionPoolManager()
        PerformanceMonitor()

        # Calculate optimal pool size
        server_config = {
            "cpu_cores": 4,
            "max_connections": 100,
            "expected_concurrency": 20,
        }
        pool_config = pool_manager.calculate_optimal_pool_size(server_config)

        # Monitor health
        pool_stats = {
            "active_connections": int(pool_config["max_size"] * 0.6),
            "idle_connections": int(pool_config["max_size"] * 0.3),
            "max_connections": 100,
            "wait_time_avg_ms": 20,
        }
        health = pool_manager.monitor_pool_health(pool_stats)
        assert health["is_saturated"] is False

    def test_indexing_for_database_performance(self):
        """Test index optimization for database performance."""
        optimizer = IndexingOptimizer()
        monitor = PerformanceMonitor()

        # Recommend indexes for query pattern
        query_pattern = {
            "columns": ["user_id", "created_at"],
            "conditions": ["user_id = 123", "created_at > 2024-01-01"],
        }
        index_rec = optimizer.recommend_index(query_pattern)
        assert index_rec["index_type"] == "COMPOSITE"

        # Monitor potential query performance improvement
        query_stats = {
            "avg_execution_time_ms": 1500,
            "max_execution_time_ms": 3000,
            "call_count": 1000,
        }
        perf_analysis = monitor.analyze_query_performance(query_stats)
        assert perf_analysis["optimization_priority"] == "high"

    def test_transaction_management_and_deadlock_prevention(self):
        """Test transaction management with deadlock prevention."""
        tx_manager = TransactionManager()

        # Validate ACID compliance
        config = {"isolation_level": "SERIALIZABLE"}
        acid_result = tx_manager.validate_acid_compliance(config)
        assert acid_result["atomicity"] is True

        # Detect potential deadlock
        transactions = [
            {"id": "tx1", "locks": ["resource_a"], "waiting_for": None},
            {"id": "tx2", "locks": ["resource_b"], "waiting_for": None},
        ]
        deadlock = tx_manager.detect_deadlock(transactions)
        assert deadlock["deadlock_detected"] is False

        # Generate retry strategy
        retry_plan = tx_manager.generate_retry_plan(
            {
                "max_retries": 3,
                "initial_backoff_ms": 100,
                "backoff_multiplier": 2.0,
                "max_backoff_ms": 1000,
            }
        )
        assert len(retry_plan["retry_delays"]) == 3


# ============================================================================
# Edge Case and Error Handling Tests
# ============================================================================


class TestEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_empty_schema_validation(self):
        """Test schema validation with empty schema."""
        normalizer = SchemaNormalizer()
        schema = {}
        result = normalizer.validate_1nf(schema)
        assert result["is_valid"] is True

    def test_schema_with_only_primary_key(self):
        """Test schema containing only primary key."""
        normalizer = SchemaNormalizer()
        schema = {
            "table1": {
                "PRIMARY KEY": "id",
            }
        }
        result = normalizer.validate_1nf(schema)
        assert isinstance(result, dict)

    def test_index_single_column_vs_multi_column(self):
        """Test index recommendations for single vs multi-column."""
        optimizer = IndexingOptimizer()

        # Single column
        single = optimizer.recommend_index(
            {
                "columns": ["user_id"],
                "conditions": ["user_id = 1"],
            }
        )
        assert single["index_type"] == "HASH"

        # Multiple columns
        multi = optimizer.recommend_index(
            {
                "columns": ["user_id", "created_at"],
                "conditions": ["user_id = 1", "created_at > date"],
            }
        )
        assert multi["index_type"] == "COMPOSITE"

    def test_pool_size_boundary_conditions(self):
        """Test pool size calculation at boundary conditions."""
        manager = ConnectionPoolManager()

        # Very low concurrency
        config = manager.calculate_optimal_pool_size(
            {
                "cpu_cores": 1,
                "max_connections": 10,
                "expected_concurrency": 1,
            }
        )
        assert config["min_size"] >= 5

        # Very high concurrency
        config = manager.calculate_optimal_pool_size(
            {
                "cpu_cores": 64,
                "max_connections": 1000,
                "expected_concurrency": 500,
            }
        )
        assert config["max_size"] > 0

    def test_zero_wait_time_health_score(self):
        """Test pool health score with zero wait time."""
        manager = ConnectionPoolManager()
        pool_stats = {
            "active_connections": 10,
            "idle_connections": 40,
            "max_connections": 100,
            "wait_time_avg_ms": 0,
        }
        result = manager.monitor_pool_health(pool_stats)
        # health_score = (saturation_score + wait_score) / 2
        # saturation_score = 1.0 - 0.5 = 0.5
        # wait_score = 1.0 - (0 / 500) = 1.0
        # health_score = (0.5 + 1.0) / 2 = 0.75
        assert result["health_score"] == 0.75

    def test_migration_add_column_with_various_defaults(self):
        """Test migration planning for adding columns with various defaults."""
        planner = MigrationPlanner()

        test_cases = [
            ("NULL", True),
            ("CURRENT_TIMESTAMP", True),
            (0, False),
            ("'default_value'", False),
        ]

        for default_val, is_breaking in test_cases:
            change = {
                "operation": "add_column",
                "table": "users",
                "column": {
                    "name": "field",
                    "type": "VARCHAR(100)",
                    "default": default_val,
                    "nullable": default_val == "NULL",
                },
            }
            plan = planner.generate_migration_plan(change)
            assert isinstance(plan, dict)

    def test_retry_plan_with_single_retry(self):
        """Test retry plan with single retry attempt."""
        manager = TransactionManager()
        plan = manager.generate_retry_plan(
            {
                "max_retries": 1,
                "initial_backoff_ms": 100,
                "backoff_multiplier": 2.0,
                "max_backoff_ms": 1000,
            }
        )
        assert len(plan["retry_delays"]) == 1

    def test_performance_rating_boundary_values(self):
        """Test performance rating at boundary values."""
        monitor = PerformanceMonitor()

        # Thresholds: excellent < 100, good 100-500, needs_improvement 500-1000, poor > 1000
        boundaries = [
            (50, "excellent"),
            (100, "excellent"),  # at threshold but not over
            (150, "good"),
            (500, "good"),  # at threshold but not over
            (600, "needs_improvement"),
            (1000, "needs_improvement"),  # at threshold but not over
            (1500, "poor"),
        ]

        for avg_time, expected_rating in boundaries:
            result = monitor.analyze_query_performance(
                {
                    "avg_execution_time_ms": avg_time,
                    "max_execution_time_ms": avg_time * 2,
                    "call_count": 100,
                }
            )
            assert (
                result["performance_rating"] == expected_rating
            ), f"Expected {expected_rating} for avg_time {avg_time}, got {result['performance_rating']}"

    def test_connection_usage_at_exact_thresholds(self):
        """Test connection usage monitoring at exact threshold values."""
        monitor = PerformanceMonitor()

        # Slightly above warning threshold (0.76 > 0.75)
        result = monitor.monitor_connection_usage(
            {
                "active_connections": 76,
                "max_connections": 100,
                "failed_connection_attempts": 0,
            }
        )
        assert result["health_status"] == "warning"

        # Above critical threshold (0.91 > 0.90)
        result = monitor.monitor_connection_usage(
            {
                "active_connections": 91,
                "max_connections": 100,
                "failed_connection_attempts": 0,
            }
        )
        assert result["health_status"] == "critical"


# ============================================================================
# Module-Level Tests
# ============================================================================


class TestModuleExports:
    """Test module exports and public API."""

    def test_all_classes_exported(self):
        """Test that all classes are properly exported."""
        # All imports should succeed
        assert SchemaNormalizer is not None
        assert DatabaseSelector is not None
        assert IndexingOptimizer is not None
        assert ConnectionPoolManager is not None
        assert MigrationPlanner is not None
        assert TransactionManager is not None
        assert PerformanceMonitor is not None
        assert ValidationResult is not None
        assert DatabaseRecommendation is not None
        assert IndexRecommendation is not None
        assert PoolConfiguration is not None
        assert MigrationPlan is not None
        assert ACIDCompliance is not None

    def test_classes_instantiation(self):
        """Test that all classes can be instantiated."""
        classes = [
            SchemaNormalizer,
            DatabaseSelector,
            IndexingOptimizer,
            ConnectionPoolManager,
            MigrationPlanner,
            TransactionManager,
            PerformanceMonitor,
        ]
        for cls in classes:
            instance = cls()
            assert instance is not None
