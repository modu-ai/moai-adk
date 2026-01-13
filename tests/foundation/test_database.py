"""
Comprehensive TDD tests for database.py module.
Tests cover all 7 classes.
"""

import pytest
from typing import Any, Dict

from moai_adk.foundation.database import (
    SchemaNormalizer,
    DatabaseSelector,
    IndexingOptimizer,
    ConnectionPoolManager,
    MigrationPlanner,
    TransactionManager,
    PerformanceMonitor,
    ValidationResult,
    IndexRecommendation,
    PoolConfiguration,
)


# ============================================================================
# Test SchemaNormalizer
# ============================================================================


class TestSchemaNormalizer:
    """Test suite for SchemaNormalizer class."""

    def test_initialization(self):
        """Test normalizer initialization."""
        normalizer = SchemaNormalizer()
        assert normalizer.schemas == {}
        assert normalizer.NORMALIZATION_LEVELS == ["1NF", "2NF", "3NF", "BCNF"]

    def test_normalize_1nf_valid(self, sample_database_schema):
        """Test 1NF normalization with valid schema."""
        normalizer = SchemaNormalizer()
        result = normalizer.normalize_1nf(sample_database_schema)

        assert result["level"] == "1NF"
        assert result["atomic_values"] is True
        assert result["no_repeating_groups"] is True
        assert result["errors"] is None

    def test_normalize_1nf_with_repeating_groups(self):
        """Test 1NF normalization failure with repeating groups."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "id": "INTEGER",
                "name": "VARCHAR(100)",
                "emails": "VARCHAR(255)[]"  # Array - repeating group
            }
        }
        result = normalizer.normalize_1nf(schema)

        assert result["level"] == "1NF"
        assert result["atomic_values"] is False
        assert "repeating groups" in result["errors"][0].lower()

    def test_normalize_2nf_valid(self, sample_database_schema):
        """Test 2NF normalization with valid schema."""
        normalizer = SchemaNormalizer()
        result = normalizer.normalize_2nf(sample_database_schema)

        assert result["level"] == "2NF"
        assert result["no_partial_dependencies"] is True

    def test_normalize_2nf_with_partial_dependencies(self):
        """Test 2NF normalization failure with partial dependencies."""
        normalizer = SchemaNormalizer()
        schema = {
            "order_items": {
                "order_id": "INTEGER",  # Part of composite key
                "product_id": "INTEGER",  # Part of composite key
                "product_name": "VARCHAR(100)",  # Depends only on product_id
                "quantity": "INTEGER"
            }
        }
        result = normalizer.normalize_2nf(schema)

        assert result["level"] == "2NF"
        assert result["no_partial_dependencies"] is False
        assert "partial dependency" in result["errors"][0].lower()

    def test_normalize_3nf_valid(self, sample_database_schema):
        """Test 3NF normalization with valid schema."""
        normalizer = SchemaNormalizer()
        result = normalizer.normalize_3nf(sample_database_schema)

        assert result["level"] == "3NF"
        assert result["no_transitive_dependencies"] is True

    def test_normalize_3nf_with_transitive_dependencies(self):
        """Test 3NF normalization failure with transitive dependencies."""
        normalizer = SchemaNormalizer()
        schema = {
            "users": {
                "id": "INTEGER PRIMARY KEY",
                "name": "VARCHAR(100)",
                "city": "VARCHAR(100)",
                "state": "VARCHAR(50)",
                "zip": "VARCHAR(10)"  # Transitive: id -> city -> zip
            }
        }
        result = normalizer.normalize_3nf(schema)

        assert result["level"] == "3NF"
        assert result["no_transitive_dependencies"] is False

    def test_get_normalization_recommendations(self, sample_database_schema):
        """Test normalization recommendations."""
        normalizer = SchemaNormalizer()
        recommendations = normalizer.get_normalization_recommendations(sample_database_schema)

        assert isinstance(recommendations, list)
        assert len(recommendations) > 0


# ============================================================================
# Test DatabaseSelector
# ============================================================================


class TestDatabaseSelector:
    """Test suite for DatabaseSelector class."""

    def test_initialization(self):
        """Test selector initialization."""
        selector = DatabaseSelector()
        assert selector.recommendations == {}

    def test_select_database_for_read_heavy(self):
        """Test database selection for read-heavy workload."""
        selector = DatabaseSelector()
        result = selector.select_database(
            workload_type="read_heavy",
            data_volume="medium",
            consistency_requirement="eventual"
        )

        assert result["database_type"] == "PostgreSQL"
        assert result["recommended"] is True
        assert "read_heavy" in result["workload_type"]

    def test_select_database_for_write_heavy(self):
        """Test database selection for write-heavy workload."""
        selector = DatabaseSelector()
        result = selector.select_database(
            workload_type="write_heavy",
            data_volume="large",
            consistency_requirement="strong"
        )

        assert result["database_type"] in ["PostgreSQL", "MySQL"]
        assert result["recommended"] is True

    def test_select_database_for_analytical(self):
        """Test database selection for analytical workload."""
        selector = DatabaseSelector()
        result = selector.select_database(
            workload_type="analytical",
            data_volume="large",
            consistency_requirement="eventual"
        )

        assert result["database_type"] in ["ClickHouse", "BigQuery"]
        assert result["recommended"] is True

    def test_select_database_for_time_series(self):
        """Test database selection for time-series data."""
        selector = DatabaseSelector()
        result = selector.select_database(
            workload_type="time_series",
            data_volume="large",
            consistency_requirement="eventual"
        )

        assert "InfluxDB" in result["database_type"] or "TimescaleDB" in result["database_type"]

    def test_select_database_for_graph(self):
        """Test database selection for graph data."""
        selector = DatabaseSelector()
        result = selector.select_database(
            workload_type="graph",
            data_volume="medium",
            consistency_requirement="strong"
        )

        assert "Neo4j" in result["database_type"]

    def test_get_database_comparison(self):
        """Test database comparison matrix."""
        selector = DatabaseSelector()
        comparison = selector.get_database_comparison()

        assert "PostgreSQL" in comparison
        assert "MySQL" in comparison
        assert "MongoDB" in comparison
        assert comparison["PostgreSQL"]["type"] == "Relational"


# ============================================================================
# Test IndexingOptimizer
# ============================================================================


class TestIndexingOptimizer:
    """Test suite for IndexingOptimizer class."""

    def test_initialization(self):
        """Test optimizer initialization."""
        optimizer = IndexingOptimizer()
        assert optimizer.indexes == {}

    def test_recommend_indexes_for_primary_key(self, sample_database_schema):
        """Test index recommendation for primary key."""
        optimizer = IndexingOptimizer()
        result = optimizer.recommend_indexes(sample_database_schema)

        assert len(result) > 0
        assert any(idx["type"] == "PRIMARY" for idx in result)

    def test_recommend_indexes_for_foreign_keys(self):
        """Test index recommendation for foreign keys."""
        optimizer = IndexingOptimizer()
        schema = {
            "orders": {
                "id": "INTEGER PRIMARY KEY",
                "user_id": "INTEGER",  # Foreign key
                "product_id": "INTEGER"  # Foreign key
            }
        }
        result = optimizer.recommend_indexes(schema)

        assert any(idx["column"] == "user_id" and idx["type"] == "INDEX" for idx in result)

    def test_recommend_indexes_for_frequent_queries(self):
        """Test index recommendation for frequent queries."""
        optimizer = IndexingOptimizer()
        schema = {
            "users": {
                "id": "INTEGER PRIMARY KEY",
                "email": "VARCHAR(255)",
                "name": "VARCHAR(100)"
            }
        }
        query_patterns = ["SELECT * FROM users WHERE email = ?", "SELECT * FROM users WHERE name = ?"]
        result = optimizer.recommend_indexes(schema, query_patterns)

        assert any(idx["column"] == "email" for idx in result)

    def test_create_composite_index(self):
        """Test composite index creation."""
        optimizer = IndexingOptimizer()
        index = optimizer.create_composite_index(
            table="orders",
            columns=["user_id", "created_at"],
            unique=False
        )

        assert index["table"] == "orders"
        assert index["columns"] == ["user_id", "created_at"]
        assert index["type"] == "COMPOSITE"
        assert index["unique"] is False

    def test_create_unique_index(self):
        """Test unique index creation."""
        optimizer = IndexingOptimizer()
        index = optimizer.create_unique_index(
            table="users",
            column="email"
        )

        assert index["table"] == "users"
        assert index["column"] == "email"
        assert index["type"] == "UNIQUE"
        assert index["unique"] is True


# ============================================================================
# Test ConnectionPoolManager
# ============================================================================


class TestConnectionPoolManager:
    """Test suite for ConnectionPoolManager class."""

    def test_initialization(self):
        """Test manager initialization."""
        manager = ConnectionPoolManager()
        assert manager.pools == {}
        assert manager.DEFAULT_POOL_SIZE == 10

    def test_create_pool_config_default(self):
        """Test pool configuration with defaults."""
        manager = ConnectionPoolManager()
        config = manager.create_pool_config("test_db")

        assert config["database"] == "test_db"
        assert config["pool_size"] == 10
        assert config["max_overflow"] == 5
        assert config["timeout"] == 30
        assert config["recycle"] == 3600

    def test_create_pool_config_custom(self):
        """Test pool configuration with custom values."""
        manager = ConnectionPoolManager()
        config = manager.create_pool_config(
            database="test_db",
            pool_size=20,
            max_overflow=10,
            timeout=60,
            recycle=7200
        )

        assert config["pool_size"] == 20
        assert config["max_overflow"] == 10
        assert config["timeout"] == 60
        assert config["recycle"] == 7200

    def test_get_pool_status_empty(self):
        """Test pool status with no pools."""
        manager = ConnectionPoolManager()
        status = manager.get_pool_status("test_db")

        assert status["active_connections"] == 0
        assert status["idle_connections"] == 0
        assert status["pool_size"] == 0

    def test_optimize_pool_size_for_read_heavy(self):
        """Test pool size optimization for read-heavy workload."""
        manager = ConnectionPoolManager()
        optimized = manager.optimize_pool_size(
            database="test_db",
            workload_type="read_heavy",
            concurrent_users=100
        )

        assert optimized["workload_type"] == "read_heavy"
        assert optimized["recommended_pool_size"] > 10
        assert optimized["concurrent_users"] == 100

    def test_optimize_pool_size_for_write_heavy(self):
        """Test pool size optimization for write-heavy workload."""
        manager = ConnectionPoolManager()
        optimized = manager.optimize_pool_size(
            database="test_db",
            workload_type="write_heavy",
            concurrent_users=50
        )

        assert optimized["workload_type"] == "write_heavy"
        assert optimized["recommended_pool_size"] > 10


# ============================================================================
# Test MigrationPlanner
# ============================================================================


class TestMigrationPlanner:
    """Test suite for MigrationPlanner class."""

    def test_initialization(self):
        """Test planner initialization."""
        planner = MigrationPlanner()
        assert planner.migrations == {}
        assert planner.applied_migrations == []

    def test_create_migration_basic(self):
        """Test basic migration creation."""
        planner = MigrationPlanner()
        migration = planner.create_migration(
            name="add_users_table",
            operations=["CREATE TABLE users (id INTEGER PRIMARY KEY, name VARCHAR(100))"]
        )

        assert migration["name"] == "add_users_table"
        assert migration["version"] is not None
        assert len(migration["operations"]) == 1
        assert "created_at" in migration

    def test_create_migration_with_dependencies(self):
        """Test migration creation with dependencies."""
        planner = MigrationPlanner()
        migration = planner.create_migration(
            name="add_posts_table",
            operations=["CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER)"],
            dependencies=["add_users_table"]
        )

        assert "add_users_table" in migration["dependencies"]

    def test_validate_migration_safe(self):
        """Test safe migration validation."""
        planner = MigrationPlanner()
        migration = {
            "name": "add_column",
            "operations": ["ALTER TABLE users ADD COLUMN email VARCHAR(255)"]
        }

        result = planner.validate_migration(migration)

        assert result["safe"] is True
        assert result["destructive"] is False

    def test_validate_migration_destructive(self):
        """Test destructive migration validation."""
        planner = MigrationPlanner()
        migration = {
            "name": "drop_column",
            "operations": ["ALTER TABLE users DROP COLUMN email"]
        }

        result = planner.validate_migration(migration)

        assert result["safe"] is False
        assert result["destructive"] is True
        assert "DROP" in result["warnings"][0]

    def test_rollback_migration(self):
        """Test migration rollback."""
        planner = MigrationPlanner()
        migration = {
            "name": "add_column",
            "operations": ["ALTER TABLE users ADD COLUMN email VARCHAR(255)"],
            "rollback_operations": ["ALTER TABLE users DROP COLUMN email"]
        }

        rollback = planner.rollback_migration(migration)

        assert rollback["operations"] == migration["rollback_operations"]
        assert "rolled_back_at" in rollback


# ============================================================================
# Test TransactionManager
# ============================================================================


class TestTransactionManager:
    """Test suite for TransactionManager class."""

    def test_initialization(self):
        """Test manager initialization."""
        manager = TransactionManager()
        assert manager.transactions == {}
        assert manager.ISOLATION_LEVELS == ["READ_UNCOMMITTED", "READ_COMMITTED", "REPEATABLE_READ", "SERIALIZABLE"]

    def test_begin_transaction(self):
        """Test transaction begin."""
        manager = TransactionManager()
        tx_id = manager.begin_transaction(isolation_level="READ_COMMITTED")

        assert tx_id is not None
        assert tx_id in manager.transactions
        assert manager.transactions[tx_id]["status"] == "active"

    def test_begin_transaction_default_isolation(self):
        """Test transaction begin with default isolation level."""
        manager = TransactionManager()
        tx_id = manager.begin_transaction()

        assert manager.transactions[tx_id]["isolation_level"] == "READ_COMMITTED"

    def test_commit_transaction(self):
        """Test transaction commit."""
        manager = TransactionManager()
        tx_id = manager.begin_transaction()
        result = manager.commit_transaction(tx_id)

        assert result["success"] is True
        assert result["transaction_id"] == tx_id
        assert manager.transactions[tx_id]["status"] == "committed"

    def test_rollback_transaction(self):
        """Test transaction rollback."""
        manager = TransactionManager()
        tx_id = manager.begin_transaction()
        result = manager.rollback_transaction(tx_id)

        assert result["success"] is True
        assert result["transaction_id"] == tx_id
        assert manager.transactions[tx_id]["status"] == "rolled_back"

    def test_validate_transaction_acyd_compliance(self):
        """Test ACID compliance validation."""
        manager = TransactionManager()
        tx_id = manager.begin_transaction(isolation_level="SERIALIZABLE")

        result = manager.validate_transaction(tx_id)

        assert result["acid_compliant"] is True
        assert result["atomicity"] is True
        assert result["consistency"] is True
        assert result["isolation"] is True
        assert result["durability"] is True


# ============================================================================
# Test PerformanceMonitor
# ============================================================================


class TestPerformanceMonitor:
    """Test suite for PerformanceMonitor class."""

    def test_initialization(self):
        """Test monitor initialization."""
        monitor = PerformanceMonitor()
        assert monitor.metrics == []
        assert monitor.slow_queries == []

    def test_monitor_query_performance(self):
        """Test query performance monitoring."""
        monitor = PerformanceMonitor()
        result = monitor.monitor_query(
            query="SELECT * FROM users",
            execution_time_ms=150.5,
            rows_returned=1000
        )

        assert result["query"] == "SELECT * FROM users"
        assert result["execution_time_ms"] == 150.5
        assert result["rows_returned"] == 1000
        assert result["performance"] == "acceptable"
        assert "timestamp" in result

    def test_monitor_query_performance_slow(self):
        """Test slow query detection."""
        monitor = PerformanceMonitor()
        result = monitor.monitor_query(
            query="SELECT * FROM large_table",
            execution_time_ms=5500,  # 5.5 seconds - slow
            rows_returned=10000
        )

        assert result["performance"] == "slow"
        assert result["slow_query"] is True

    def test_get_slow_queries(self):
        """Test slow queries retrieval."""
        monitor = PerformanceMonitor()
        monitor.monitor_query("SELECT 1", 100, 1)
        monitor.monitor_query("SELECT * FROM large_table", 5500, 10000)

        slow_queries = monitor.get_slow_queries(threshold_ms=1000)

        assert len(slow_queries) == 1
        assert slow_queries[0]["execution_time_ms"] == 5500

    def test_get_performance_summary(self):
        """Test performance summary."""
        monitor = PerformanceMonitor()
        monitor.monitor_query("SELECT 1", 50, 1)
        monitor.monitor_query("SELECT 2", 75, 1)

        summary = monitor.get_performance_summary()

        assert summary["total_queries"] == 2
        assert summary["average_time_ms"] > 0
        assert summary["slow_query_count"] == 0

    def test_identify_missing_indexes(self):
        """Test missing index identification."""
        monitor = PerformanceMonitor()
        # Simulate slow query with table scan
        monitor.monitor_query(
            query="SELECT * FROM users WHERE email = 'test@example.com'",
            execution_time_ms=3000,
            rows_returned=1,
            rows_scanned=100000
        )

        recommendations = monitor.identify_missing_indexes()

        assert len(recommendations) > 0
        assert "users" in recommendations[0]["table"]


# ============================================================================
# Test Data Classes
# ============================================================================


class TestDataClasses:
    """Test suite for database data classes."""

    def test_table_schema_creation(self):
        """Test ValidationResult dataclass creation."""
        schema = ValidationResult(
            table_name="users",
            columns={"id": "INTEGER", "name": "VARCHAR(100)"},
            primary_key="id"
        )

        assert schema.table_name == "users"
        assert schema.columns["id"] == "INTEGER"
        assert schema.primary_key == "id"

    def test_index_config_creation(self):
        """Test IndexRecommendation dataclass creation."""
        index = IndexRecommendation(
            table="users",
            column="email",
            index_type="UNIQUE",
            unique=True
        )

        assert index.table == "users"
        assert index.column == "email"
        assert index.unique is True

    def test_pool_config_creation(self):
        """Test PoolConfiguration dataclass creation."""
        config = PoolConfiguration(
            database="test_db",
            pool_size=20,
            max_overflow=10,
            timeout=30
        )

        assert config.database == "test_db"
        assert config.pool_size == 20
        assert config.max_overflow == 10
