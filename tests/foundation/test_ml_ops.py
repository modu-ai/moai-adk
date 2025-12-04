"""
Comprehensive test suite for MLOps module.

Tests cover ML pipeline orchestration, model version management,
data pipeline construction, model deployment, drift detection,
performance optimization, and metrics collection with 90%+ coverage goal.
"""

from datetime import datetime

import pytest

from src.moai_adk.foundation.ml_ops import (
    DataPipelineBuilder,
    DriftDetectionMonitor,
    MLOpsMetricsCollector,
    MLPipelineOrchestrator,
    ModelDeploymentPlanner,
    ModelVersionManager,
    PerformanceOptimizer,
)

# ============================================================================
# MLPipelineOrchestrator Tests
# ============================================================================


class TestMLPipelineOrchestrator:
    """Test ML pipeline orchestration functionality."""

    @pytest.fixture
    def orchestrator(self):
        """Create MLPipelineOrchestrator instance for testing."""
        return MLPipelineOrchestrator()

    def test_orchestrator_initialization(self, orchestrator):
        """Test MLPipelineOrchestrator initialization."""
        assert orchestrator.supported_platforms == ["mlflow", "kubeflow", "airflow"]
        assert orchestrator.execution_history == []

    def test_orchestrate_mlflow_pipeline_basic(self, orchestrator):
        """Test basic MLflow pipeline orchestration."""
        config = orchestrator.orchestrate_mlflow_pipeline(
            experiment_name="test_exp",
            run_name="test_run",
            tracking_uri="http://localhost:5000",
        )

        assert config["type"] == "mlflow"
        assert config["experiment_name"] == "test_exp"
        assert config["run_name"] == "test_run"
        assert config["tracking_uri"] == "http://localhost:5000"
        assert config["parameters"]["log_params"] is True
        assert config["parameters"]["log_metrics"] is True
        assert config["parameters"]["log_artifacts"] is True
        assert config["parameters"]["auto_log"] is False

    def test_orchestrate_mlflow_pipeline_with_tags(self, orchestrator):
        """Test MLflow pipeline with custom tags."""
        custom_tags = {
            "framework": "tensorflow",
            "environment": "prod",
            "team": "mlops",
        }
        config = orchestrator.orchestrate_mlflow_pipeline(
            experiment_name="exp_001",
            run_name="run_001",
            tracking_uri="http://mlflow:5000",
            tags=custom_tags,
        )

        assert config["tags"] == custom_tags
        assert config["tags"]["framework"] == "tensorflow"
        assert config["tags"]["environment"] == "prod"

    def test_orchestrate_mlflow_pipeline_default_tags(self, orchestrator):
        """Test MLflow pipeline with default tags."""
        config = orchestrator.orchestrate_mlflow_pipeline(
            experiment_name="exp_002",
            run_name="run_002",
            tracking_uri="http://localhost:5000",
        )

        assert config["tags"]["framework"] == "pytorch"
        assert config["tags"]["environment"] == "dev"

    def test_orchestrate_mlflow_pipeline_metrics(self, orchestrator):
        """Test MLflow pipeline metrics initialization."""
        config = orchestrator.orchestrate_mlflow_pipeline(
            experiment_name="metrics_test",
            run_name="test_run",
            tracking_uri="http://localhost:5000",
        )

        assert "metrics" in config
        assert "accuracy" in config["metrics"]
        assert "precision" in config["metrics"]
        assert "recall" in config["metrics"]
        assert "f1_score" in config["metrics"]
        # All metrics should be None initially
        for metric_value in config["metrics"].values():
            assert metric_value is None

    def test_orchestrate_mlflow_pipeline_artifacts(self, orchestrator):
        """Test MLflow pipeline artifact paths."""
        experiment_name = "artifact_test"
        run_name = "run_001"
        config = orchestrator.orchestrate_mlflow_pipeline(
            experiment_name=experiment_name,
            run_name=run_name,
            tracking_uri="http://localhost:5000",
        )

        assert "artifacts" in config
        assert config["artifacts"]["model_path"] == f"/tmp/models/{experiment_name}"
        assert config["artifacts"]["logs_path"] == f"/tmp/logs/{run_name}"

    def test_orchestrate_mlflow_pipeline_timestamp(self, orchestrator):
        """Test MLflow pipeline includes creation timestamp."""
        config = orchestrator.orchestrate_mlflow_pipeline(
            experiment_name="time_test",
            run_name="run_001",
            tracking_uri="http://localhost:5000",
        )

        assert "created_at" in config
        # Verify it's a valid ISO 8601 timestamp
        created_at = datetime.fromisoformat(config["created_at"])
        assert created_at is not None

    def test_orchestrate_kubeflow_pipeline_basic(self, orchestrator):
        """Test basic Kubeflow pipeline orchestration."""
        components = ["preprocessing", "training", "evaluation"]
        spec = orchestrator.orchestrate_kubeflow_pipeline(
            pipeline_name="kf_pipeline",
            namespace="default",
            components=components,
        )

        assert spec["type"] == "kubeflow"
        assert spec["pipeline_name"] == "kf_pipeline"
        assert spec["namespace"] == "default"
        assert spec["components"] == components

    def test_orchestrate_kubeflow_pipeline_parameters(self, orchestrator):
        """Test Kubeflow pipeline with custom parameters."""
        custom_params = {"batch_size": 64, "epochs": 20, "learning_rate": 0.001}
        spec = orchestrator.orchestrate_kubeflow_pipeline(
            pipeline_name="kf_pipeline",
            namespace="ml",
            components=["train"],
            parameters=custom_params,
        )

        assert spec["parameters"] == custom_params

    def test_orchestrate_kubeflow_pipeline_default_parameters(self, orchestrator):
        """Test Kubeflow pipeline with default parameters."""
        spec = orchestrator.orchestrate_kubeflow_pipeline(
            pipeline_name="kf_pipeline",
            namespace="default",
            components=["train"],
        )

        assert spec["parameters"]["batch_size"] == 32
        assert spec["parameters"]["epochs"] == 10

    def test_orchestrate_kubeflow_pipeline_resources(self, orchestrator):
        """Test Kubeflow pipeline resource specifications."""
        spec = orchestrator.orchestrate_kubeflow_pipeline(
            pipeline_name="kf_pipeline",
            namespace="default",
            components=["train"],
        )

        assert "resources" in spec
        assert spec["resources"]["cpu"] == "2"
        assert spec["resources"]["memory"] == "4Gi"
        assert spec["resources"]["gpu"] == "1"

    def test_orchestrate_kubeflow_pipeline_volumes(self, orchestrator):
        """Test Kubeflow pipeline volumes configuration."""
        spec = orchestrator.orchestrate_kubeflow_pipeline(
            pipeline_name="kf_pipeline",
            namespace="default",
            components=["train"],
        )

        assert "volumes" in spec
        assert len(spec["volumes"]) == 2
        assert spec["volumes"][0]["name"] == "data"
        assert spec["volumes"][1]["name"] == "models"

    def test_orchestrate_airflow_dag_basic(self, orchestrator):
        """Test basic Airflow DAG orchestration."""
        tasks = ["extract", "transform", "load"]
        dag_config = orchestrator.orchestrate_airflow_dags(
            dag_id="ml_pipeline",
            schedule_interval="@daily",
            tasks=tasks,
        )

        assert dag_config["type"] == "airflow"
        assert dag_config["dag_id"] == "ml_pipeline"
        assert dag_config["schedule_interval"] == "@daily"
        assert dag_config["tasks"] == tasks

    def test_orchestrate_airflow_dag_auto_dependencies(self, orchestrator):
        """Test Airflow DAG auto-generates linear dependencies."""
        tasks = ["task1", "task2", "task3"]
        dag_config = orchestrator.orchestrate_airflow_dags(
            dag_id="ml_pipeline",
            schedule_interval="@daily",
            tasks=tasks,
        )

        assert "dependencies" in dag_config
        assert dag_config["dependencies"]["task1"] == ["task2"]
        assert dag_config["dependencies"]["task2"] == ["task3"]
        # task3 should not have dependencies
        assert "task3" not in dag_config["dependencies"] or dag_config["dependencies"].get("task3") == []

    def test_orchestrate_airflow_dag_custom_dependencies(self, orchestrator):
        """Test Airflow DAG with custom dependencies."""
        tasks = ["task1", "task2", "task3"]
        custom_deps = {"task1": ["task2", "task3"], "task2": ["task3"]}
        dag_config = orchestrator.orchestrate_airflow_dags(
            dag_id="ml_pipeline",
            schedule_interval="@daily",
            tasks=tasks,
            dependencies=custom_deps,
        )

        assert dag_config["dependencies"] == custom_deps

    def test_orchestrate_airflow_dag_default_args(self, orchestrator):
        """Test Airflow DAG default arguments."""
        dag_config = orchestrator.orchestrate_airflow_dags(
            dag_id="ml_pipeline",
            schedule_interval="@daily",
            tasks=["task1"],
        )

        assert "default_args" in dag_config
        assert dag_config["default_args"]["owner"] == "mlops_team"
        assert dag_config["default_args"]["retries"] == 3
        assert dag_config["default_args"]["retry_delay"] == "5m"

    def test_orchestrate_airflow_dag_execution_settings(self, orchestrator):
        """Test Airflow DAG execution settings."""
        dag_config = orchestrator.orchestrate_airflow_dags(
            dag_id="ml_pipeline",
            schedule_interval="@daily",
            tasks=["task1"],
        )

        assert dag_config["catchup"] is False
        assert dag_config["max_active_runs"] == 1

    def test_track_pipeline_execution(self, orchestrator):
        """Test pipeline execution tracking."""
        metrics = orchestrator.track_pipeline_execution(
            pipeline_id="pipe_001",
            status="completed",
            start_time="2025-11-26T10:00:00",
            end_time="2025-11-26T11:00:00",
        )

        assert metrics["pipeline_id"] == "pipe_001"
        assert metrics["status"] == "completed"
        assert metrics["start_time"] == "2025-11-26T10:00:00"
        assert metrics["end_time"] == "2025-11-26T11:00:00"

    def test_track_pipeline_execution_auto_end_time(self, orchestrator):
        """Test pipeline execution with auto-generated end time."""
        metrics = orchestrator.track_pipeline_execution(
            pipeline_id="pipe_002",
            status="running",
            start_time="2025-11-26T10:00:00",
        )

        assert metrics["pipeline_id"] == "pipe_002"
        assert metrics["status"] == "running"
        assert metrics["end_time"] is not None
        # Verify end_time is ISO format
        datetime.fromisoformat(metrics["end_time"])

    def test_track_pipeline_execution_metrics(self, orchestrator):
        """Test pipeline execution metrics collection."""
        metrics = orchestrator.track_pipeline_execution(
            pipeline_id="pipe_003",
            status="completed",
            start_time="2025-11-26T10:00:00",
        )

        assert "execution_metrics" in metrics
        assert metrics["execution_metrics"]["tasks_completed"] == 0
        assert metrics["execution_metrics"]["tasks_failed"] == 0
        assert metrics["execution_metrics"]["tasks_pending"] == 0

    def test_track_pipeline_execution_resource_usage(self, orchestrator):
        """Test pipeline execution resource usage tracking."""
        metrics = orchestrator.track_pipeline_execution(
            pipeline_id="pipe_004",
            status="completed",
            start_time="2025-11-26T10:00:00",
        )

        assert "resource_usage" in metrics
        assert metrics["resource_usage"]["cpu_hours"] == 0.0
        assert metrics["resource_usage"]["memory_gb_hours"] == 0.0
        assert metrics["resource_usage"]["gpu_hours"] == 0.0

    def test_track_pipeline_execution_history(self, orchestrator):
        """Test pipeline execution history tracking."""
        assert len(orchestrator.execution_history) == 0

        orchestrator.track_pipeline_execution(
            pipeline_id="pipe_005",
            status="completed",
            start_time="2025-11-26T10:00:00",
        )
        assert len(orchestrator.execution_history) == 1

        orchestrator.track_pipeline_execution(
            pipeline_id="pipe_006",
            status="failed",
            start_time="2025-11-26T11:00:00",
        )
        assert len(orchestrator.execution_history) == 2


# ============================================================================
# ModelVersionManager Tests
# ============================================================================


class TestModelVersionManager:
    """Test model version management functionality."""

    @pytest.fixture
    def manager(self):
        """Create ModelVersionManager instance for testing."""
        return ModelVersionManager()

    def test_manager_initialization(self, manager):
        """Test ModelVersionManager initialization."""
        assert manager.registry == {}
        assert manager.lineage_graph == {}

    def test_register_model_version_basic(self, manager):
        """Test basic model version registration."""
        result = manager.register_model_version(
            model_name="test_model",
            version="v1.0.0",
            registry_uri="s3://models/test_model",
        )

        assert result["model_name"] == "test_model"
        assert result["version"] == "v1.0.0"
        assert result["registry_uri"] == "s3://models/test_model"
        assert "version_id" in result
        assert "created_at" in result

    def test_register_model_version_with_metadata(self, manager):
        """Test model version registration with metadata."""
        metadata = {"framework": "pytorch", "accuracy": 0.95, "f1_score": 0.93}
        result = manager.register_model_version(
            model_name="classifier",
            version="v1.1.0",
            registry_uri="s3://models/classifier",
            metadata=metadata,
        )

        assert result["metadata"] == metadata
        assert result["metadata"]["framework"] == "pytorch"

    def test_register_model_version_default_metadata(self, manager):
        """Test model version registration without metadata."""
        result = manager.register_model_version(
            model_name="model",
            version="v1.0.0",
            registry_uri="s3://models/model",
        )

        assert result["metadata"] == {}

    def test_register_model_version_stage(self, manager):
        """Test model version stage is set to staging."""
        result = manager.register_model_version(
            model_name="model",
            version="v1.0.0",
            registry_uri="s3://models/model",
        )

        assert result["stage"] == "staging"

    def test_register_model_version_tags(self, manager):
        """Test model version tags initialization."""
        result = manager.register_model_version(
            model_name="model",
            version="v1.0.0",
            registry_uri="s3://models/model",
        )

        assert result["tags"] == []

    def test_register_model_version_unique_id(self, manager):
        """Test each version gets unique ID."""
        result1 = manager.register_model_version(
            model_name="model",
            version="v1.0.0",
            registry_uri="s3://models/model",
        )
        result2 = manager.register_model_version(
            model_name="model",
            version="v1.1.0",
            registry_uri="s3://models/model",
        )

        assert result1["version_id"] != result2["version_id"]

    def test_register_model_version_stored_in_registry(self, manager):
        """Test model version is stored in registry."""
        manager.register_model_version(
            model_name="model1",
            version="v1.0.0",
            registry_uri="s3://models/model1",
        )

        assert "model1" in manager.registry
        assert "v1.0.0" in manager.registry["model1"]

    def test_register_multiple_versions_same_model(self, manager):
        """Test registering multiple versions of same model."""
        manager.register_model_version(
            model_name="model",
            version="v1.0.0",
            registry_uri="s3://models/model",
        )
        manager.register_model_version(
            model_name="model",
            version="v1.1.0",
            registry_uri="s3://models/model",
        )
        manager.register_model_version(
            model_name="model",
            version="v2.0.0",
            registry_uri="s3://models/model",
        )

        assert len(manager.registry["model"]) == 3

    def test_track_model_lineage_basic(self, manager):
        """Test basic model lineage tracking."""
        lineage = manager.track_model_lineage(model_id="model_001")

        assert lineage["model_id"] == "model_001"
        assert "data_sources" in lineage
        assert "training_runs" in lineage
        assert "parent_models" in lineage
        assert "deployment_history" in lineage

    def test_track_model_lineage_include_data_sources(self, manager):
        """Test model lineage with data sources."""
        lineage = manager.track_model_lineage(model_id="model_001", include_data_sources=True)

        assert lineage["data_sources"] is not None
        assert len(lineage["data_sources"]) == 2
        assert lineage["data_sources"][0]["dataset_id"] == "train_001"
        assert lineage["data_sources"][1]["dataset_id"] == "valid_001"

    def test_track_model_lineage_exclude_data_sources(self, manager):
        """Test model lineage without data sources."""
        lineage = manager.track_model_lineage(model_id="model_001", include_data_sources=False)

        assert lineage["data_sources"] is None

    def test_track_model_lineage_include_training_runs(self, manager):
        """Test model lineage with training runs."""
        lineage = manager.track_model_lineage(model_id="model_001", include_training_runs=True)

        assert lineage["training_runs"] is not None
        assert len(lineage["training_runs"]) == 1
        assert lineage["training_runs"][0]["run_id"] == "run_001"
        assert lineage["training_runs"][0]["framework"] == "pytorch"

    def test_track_model_lineage_exclude_training_runs(self, manager):
        """Test model lineage without training runs."""
        lineage = manager.track_model_lineage(model_id="model_001", include_training_runs=False)

        assert lineage["training_runs"] is None

    def test_track_model_lineage_graph(self, manager):
        """Test model lineage graph structure."""
        lineage = manager.track_model_lineage(model_id="model_001")

        assert "lineage_graph" in lineage
        assert "nodes" in lineage["lineage_graph"]
        assert "edges" in lineage["lineage_graph"]
        assert len(lineage["lineage_graph"]["nodes"]) == 3
        assert len(lineage["lineage_graph"]["edges"]) == 2

    def test_track_model_lineage_stored(self, manager):
        """Test model lineage is stored in graph."""
        manager.track_model_lineage(model_id="model_001")

        assert "model_001" in manager.lineage_graph
        assert manager.lineage_graph["model_001"]["model_id"] == "model_001"

    def test_manage_model_artifacts_weights(self, manager):
        """Test model artifact management for weights."""
        artifacts = manager.manage_model_artifacts(
            model_id="model_001",
            artifact_types=["weights"],
            storage_backend="s3",
        )

        assert "weights" in artifacts
        assert artifacts["weights"]["path"] == "s3://models/model_001/weights.pkl"
        assert "size_mb" in artifacts["weights"]
        assert "checksum" in artifacts["weights"]

    def test_manage_model_artifacts_config(self, manager):
        """Test model artifact management for config."""
        artifacts = manager.manage_model_artifacts(
            model_id="model_001",
            artifact_types=["config"],
            storage_backend="s3",
        )

        assert "config" in artifacts
        assert artifacts["config"]["path"] == "s3://models/model_001/config.json"
        assert artifacts["config"]["version"] == "v1.0"

    def test_manage_model_artifacts_metadata(self, manager):
        """Test model artifact management for metadata."""
        artifacts = manager.manage_model_artifacts(
            model_id="model_001",
            artifact_types=["metadata"],
            storage_backend="s3",
        )

        assert "metadata" in artifacts
        assert artifacts["metadata"]["path"] == "s3://models/model_001/metadata.yaml"
        assert "created_at" in artifacts["metadata"]

    def test_manage_model_artifacts_all_types(self, manager):
        """Test model artifact management for all types."""
        artifacts = manager.manage_model_artifacts(
            model_id="model_001",
            artifact_types=["weights", "config", "metadata"],
            storage_backend="s3",
        )

        assert "weights" in artifacts
        assert "config" in artifacts
        assert "metadata" in artifacts

    def test_manage_model_artifacts_storage_backends(self, manager):
        """Test model artifacts with different storage backends."""
        # Test S3
        artifacts_s3 = manager.manage_model_artifacts(
            model_id="model_001",
            artifact_types=["weights"],
            storage_backend="s3",
        )
        assert artifacts_s3["storage_backend"] == "s3"

        # Test GCS
        artifacts_gcs = manager.manage_model_artifacts(
            model_id="model_001",
            artifact_types=["weights"],
            storage_backend="gcs",
        )
        assert artifacts_gcs["storage_backend"] == "gcs"

        # Test Azure
        artifacts_azure = manager.manage_model_artifacts(
            model_id="model_001",
            artifact_types=["weights"],
            storage_backend="azure",
        )
        assert artifacts_azure["storage_backend"] == "azure"

    def test_manage_model_artifacts_checksum(self, manager):
        """Test model artifacts have consistent checksum."""
        model_id = "model_001"
        artifacts1 = manager.manage_model_artifacts(
            model_id=model_id,
            artifact_types=["weights"],
        )
        artifacts2 = manager.manage_model_artifacts(
            model_id=model_id,
            artifact_types=["weights"],
        )

        assert artifacts1["weights"]["checksum"] == artifacts2["weights"]["checksum"]


# ============================================================================
# DataPipelineBuilder Tests
# ============================================================================


class TestDataPipelineBuilder:
    """Test data pipeline construction functionality."""

    @pytest.fixture
    def builder(self):
        """Create DataPipelineBuilder instance for testing."""
        return DataPipelineBuilder()

    def test_builder_initialization(self, builder):
        """Test DataPipelineBuilder initialization."""
        assert builder.pipelines == {}

    def test_build_feature_pipeline_basic(self, builder):
        """Test basic feature pipeline construction."""
        features = ["age", "income", "credit_score"]
        transformations = ["normalize", "encode_categorical"]

        pipeline = builder.build_feature_pipeline(
            features=features,
            transformations=transformations,
        )

        assert pipeline["features"] == features
        assert pipeline["transformations"] == transformations
        assert "pipeline_steps" in pipeline

    def test_build_feature_pipeline_steps_count(self, builder):
        """Test feature pipeline generates correct number of steps."""
        features = ["age", "income"]
        transformations = ["normalize", "scale", "encode"]

        pipeline = builder.build_feature_pipeline(
            features=features,
            transformations=transformations,
        )

        assert len(pipeline["pipeline_steps"]) == 3

    def test_build_feature_pipeline_step_structure(self, builder):
        """Test feature pipeline step structure."""
        features = ["age"]
        transformations = ["normalize"]

        pipeline = builder.build_feature_pipeline(
            features=features,
            transformations=transformations,
        )

        step = pipeline["pipeline_steps"][0]
        assert step["step_id"] == 1
        assert step["transformation"] == "normalize"
        assert "input_features" in step
        assert "output_features" in step

    def test_build_feature_pipeline_create_interactions(self, builder):
        """Test feature pipeline interaction feature creation."""
        features = ["age", "income"]
        transformations = ["create_interactions"]

        pipeline = builder.build_feature_pipeline(
            features=features,
            transformations=transformations,
        )

        step = pipeline["pipeline_steps"][0]
        # Should have original features plus interaction features
        assert "age_income" in step["output_features"]
        assert len(step["output_features"]) > len(features)

    def test_build_feature_pipeline_multiple_features_interactions(self, builder):
        """Test interaction creation with multiple features."""
        features = ["a", "b", "c"]
        transformations = ["create_interactions"]

        pipeline = builder.build_feature_pipeline(
            features=features,
            transformations=transformations,
        )

        step = pipeline["pipeline_steps"][0]
        output_features = step["output_features"]
        # Check for all interaction pairs
        assert "a_b" in output_features
        assert "a_c" in output_features
        assert "b_c" in output_features

    def test_validate_data_quality_passed_checks(self, builder):
        """Test data quality validation with passing checks."""
        validation = builder.validate_data_quality(
            dataset_id="dataset_001",
            checks=["missing_values", "schema_compliance"],
        )

        assert validation["dataset_id"] == "dataset_001"
        assert len(validation["passed_checks"]) == 2
        assert "missing_values" in validation["passed_checks"]
        assert "schema_compliance" in validation["passed_checks"]

    def test_validate_data_quality_violations(self, builder):
        """Test data quality validation with violations."""
        validation = builder.validate_data_quality(
            dataset_id="dataset_001",
            checks=["missing_values", "outliers"],
        )

        assert len(validation["violations"]) == 1
        assert validation["violations"][0]["check"] == "outliers"
        assert validation["violations"][0]["severity"] == "warning"

    def test_validate_data_quality_score(self, builder):
        """Test data quality score calculation."""
        validation = builder.validate_data_quality(
            dataset_id="dataset_001",
            checks=["missing_values", "schema_compliance", "outliers"],
        )

        # 2 passed, 1 violation = 2/3 = 0.666...
        assert validation["quality_score"] == pytest.approx(0.666, rel=0.01)

    def test_validate_data_quality_empty_checks(self, builder):
        """Test data quality validation with empty checks."""
        validation = builder.validate_data_quality(
            dataset_id="dataset_001",
            checks=[],
        )

        assert validation["quality_score"] == 0.0

    def test_handle_missing_values_drop_high_ratio(self, builder):
        """Test missing value handling with high missing ratio."""
        strategy = builder.handle_missing_values(
            column_name="column1",
            missing_ratio=0.7,
            data_type="numeric",
        )

        assert strategy["method"] == "drop"
        assert strategy["missing_ratio"] == 0.7

    def test_handle_missing_values_numeric_median(self, builder):
        """Test missing value handling for numeric with median."""
        strategy = builder.handle_missing_values(
            column_name="numeric_col",
            missing_ratio=0.2,
            data_type="numeric",
        )

        assert strategy["method"] == "median"

    def test_handle_missing_values_numeric_mean(self, builder):
        """Test missing value handling for numeric with mean."""
        strategy = builder.handle_missing_values(
            column_name="numeric_col",
            missing_ratio=0.05,
            data_type="numeric",
        )

        assert strategy["method"] == "mean"

    def test_handle_missing_values_categorical(self, builder):
        """Test missing value handling for categorical."""
        strategy = builder.handle_missing_values(
            column_name="category_col",
            missing_ratio=0.1,
            data_type="categorical",
        )

        assert strategy["method"] == "mode"

    def test_handle_missing_values_datetime(self, builder):
        """Test missing value handling for datetime."""
        strategy = builder.handle_missing_values(
            column_name="date_col",
            missing_ratio=0.1,
            data_type="datetime",
        )

        assert strategy["method"] == "forward_fill"

    def test_handle_missing_values_parameters(self, builder):
        """Test missing value handling parameters."""
        strategy = builder.handle_missing_values(
            column_name="col",
            missing_ratio=0.7,
            data_type="numeric",
        )

        assert "parameters" in strategy
        assert "fill_value" in strategy["parameters"]
        assert "strategy" in strategy["parameters"]

    def test_get_data_statistics(self, builder):
        """Test data statistics retrieval."""
        stats = builder.get_data_statistics(dataset_id="dataset_001")

        assert stats["dataset_id"] == "dataset_001"
        assert stats["row_count"] == 100000
        assert stats["column_count"] == 25
        assert stats["missing_count"] == 1500
        assert stats["mean"] == 45.2
        assert stats["std"] == 12.8


# ============================================================================
# ModelDeploymentPlanner Tests
# ============================================================================


class TestModelDeploymentPlanner:
    """Test model deployment planning functionality."""

    @pytest.fixture
    def planner(self):
        """Create ModelDeploymentPlanner instance for testing."""
        return ModelDeploymentPlanner()

    def test_planner_initialization(self, planner):
        """Test ModelDeploymentPlanner initialization."""
        assert planner.deployments == {}

    def test_plan_ray_serve_deployment_basic(self, planner):
        """Test basic Ray Serve deployment planning."""
        config = planner.plan_ray_serve_deployment(
            model_name="model_001",
            replicas=3,
            route_prefix="/predict",
        )

        assert config["type"] == "ray_serve"
        assert config["model_name"] == "model_001"
        assert config["replicas"] == 3
        assert config["route_prefix"] == "/predict"

    def test_plan_ray_serve_deployment_autoscaling(self, planner):
        """Test Ray Serve deployment with autoscaling."""
        config = planner.plan_ray_serve_deployment(
            model_name="model_001",
            replicas=3,
            route_prefix="/predict",
            autoscaling=True,
        )

        assert config["autoscaling"]["enabled"] is True
        assert config["autoscaling"]["min_replicas"] == 1
        assert config["autoscaling"]["max_replicas"] == 10

    def test_plan_ray_serve_deployment_no_autoscaling(self, planner):
        """Test Ray Serve deployment without autoscaling."""
        config = planner.plan_ray_serve_deployment(
            model_name="model_001",
            replicas=3,
            route_prefix="/predict",
            autoscaling=False,
        )

        assert config["autoscaling"]["enabled"] is False

    def test_plan_ray_serve_deployment_config(self, planner):
        """Test Ray Serve deployment configuration."""
        config = planner.plan_ray_serve_deployment(
            model_name="model_001",
            replicas=5,
            route_prefix="/predict",
        )

        assert "deployment_config" in config
        assert config["deployment_config"]["num_replicas"] == 5
        assert config["deployment_config"]["max_concurrent_queries"] == 100

    def test_plan_ray_serve_deployment_ray_actor_options(self, planner):
        """Test Ray Serve Ray actor options."""
        config = planner.plan_ray_serve_deployment(
            model_name="model_001",
            replicas=3,
            route_prefix="/predict",
        )

        assert "ray_actor_options" in config["deployment_config"]
        assert config["deployment_config"]["ray_actor_options"]["num_cpus"] == 2
        assert config["deployment_config"]["ray_actor_options"]["num_gpus"] == 0.5

    def test_plan_ray_serve_deployment_stored(self, planner):
        """Test Ray Serve deployment is stored."""
        planner.plan_ray_serve_deployment(
            model_name="model_001",
            replicas=3,
            route_prefix="/predict",
        )

        assert "model_001" in planner.deployments

    def test_plan_kserve_deployment_basic(self, planner):
        """Test basic KServe deployment planning."""
        spec = planner.plan_kserve_deployment(
            model_name="model_001",
            framework="pytorch",
            storage_uri="s3://models/model_001",
        )

        assert spec["type"] == "kserve"
        assert spec["model_name"] == "model_001"
        assert spec["framework"] == "pytorch"
        assert spec["storage_uri"] == "s3://models/model_001"

    def test_plan_kserve_deployment_predictor(self, planner):
        """Test KServe deployment predictor configuration."""
        spec = planner.plan_kserve_deployment(
            model_name="model_001",
            framework="tensorflow",
            storage_uri="s3://models/model_001",
        )

        assert "predictor" in spec
        assert spec["predictor"]["model_format"] == "tensorflow"
        assert spec["predictor"]["protocol"] == "v2"
        assert spec["predictor"]["runtime"] == "tensorflow-serving"

    def test_plan_kserve_deployment_resources(self, planner):
        """Test KServe deployment resource specifications."""
        spec = planner.plan_kserve_deployment(
            model_name="model_001",
            framework="sklearn",
            storage_uri="s3://models/model_001",
        )

        assert "resources" in spec
        assert spec["resources"]["requests"]["cpu"] == "1"
        assert spec["resources"]["requests"]["memory"] == "2Gi"
        assert spec["resources"]["limits"]["cpu"] == "2"
        assert spec["resources"]["limits"]["memory"] == "4Gi"

    def test_plan_kserve_deployment_scaling(self, planner):
        """Test KServe deployment scaling configuration."""
        spec = planner.plan_kserve_deployment(
            model_name="model_001",
            framework="pytorch",
            storage_uri="s3://models/model_001",
        )

        assert spec["scaling"]["minReplicas"] == 1
        assert spec["scaling"]["maxReplicas"] == 5

    def test_containerize_model_dockerfile(self, planner):
        """Test model containerization generates Dockerfile."""
        container = planner.containerize_model(
            model_path="model.pkl",
            base_image="python:3.10-slim",
            requirements=["numpy", "torch"],
        )

        assert "dockerfile_content" in container
        assert "FROM python:3.10-slim" in container["dockerfile_content"]
        assert "COPY model.pkl /app/model.pkl" in container["dockerfile_content"]

    def test_containerize_model_config(self, planner):
        """Test model containerization configuration."""
        requirements = ["numpy", "scikit-learn", "pandas"]
        container = planner.containerize_model(
            model_path="model.pkl",
            base_image="python:3.10",
            requirements=requirements,
        )

        assert container["dockerfile_path"] == "/tmp/Dockerfile"
        assert container["base_image"] == "python:3.10"
        assert container["requirements"] == requirements

    def test_containerize_model_build_config(self, planner):
        """Test model containerization build configuration."""
        container = planner.containerize_model(
            model_path="model.pkl",
            base_image="python:3.10-slim",
            requirements=["numpy"],
        )

        assert "build_config" in container
        assert container["build_config"]["tag"] == "model:latest"
        assert container["build_config"]["platform"] == "linux/amd64"
        assert container["build_config"]["no_cache"] is False

    def test_create_inference_endpoint_basic(self, planner):
        """Test basic inference endpoint creation."""
        endpoint = planner.create_inference_endpoint(
            model_id="model_001",
            endpoint_name="fraud_detector",
        )

        assert endpoint["endpoint_name"] == "fraud_detector"
        assert endpoint["model_id"] == "model_001"
        assert "endpoint_url" in endpoint
        assert "health_check_url" in endpoint

    def test_create_inference_endpoint_auth_enabled(self, planner):
        """Test inference endpoint with authentication."""
        endpoint = planner.create_inference_endpoint(
            model_id="model_001",
            endpoint_name="fraud_detector",
            auth_enabled=True,
        )

        assert endpoint["auth_enabled"] is True
        assert endpoint["auth_token"] is not None
        assert len(endpoint["auth_token"]) == 32

    def test_create_inference_endpoint_auth_disabled(self, planner):
        """Test inference endpoint without authentication."""
        endpoint = planner.create_inference_endpoint(
            model_id="model_001",
            endpoint_name="fraud_detector",
            auth_enabled=False,
        )

        assert endpoint["auth_enabled"] is False
        assert endpoint["auth_token"] is None

    def test_create_inference_endpoint_methods(self, planner):
        """Test inference endpoint HTTP methods."""
        endpoint = planner.create_inference_endpoint(
            model_id="model_001",
            endpoint_name="detector",
        )

        assert "methods" in endpoint
        assert "POST" in endpoint["methods"]

    def test_create_inference_endpoint_rate_limit(self, planner):
        """Test inference endpoint rate limiting."""
        endpoint = planner.create_inference_endpoint(
            model_id="model_001",
            endpoint_name="detector",
        )

        assert "rate_limit" in endpoint
        assert endpoint["rate_limit"]["requests_per_minute"] == 1000

    def test_create_inference_endpoint_url_format(self, planner):
        """Test inference endpoint URL format."""
        endpoint = planner.create_inference_endpoint(
            model_id="model_001",
            endpoint_name="fraud_detector",
        )

        assert endpoint["endpoint_url"] == "https://api.example.com/fraud_detector"
        assert endpoint["health_check_url"] == "https://api.example.com/fraud_detector/health"


# ============================================================================
# DriftDetectionMonitor Tests
# ============================================================================


class TestDriftDetectionMonitor:
    """Test drift detection monitoring functionality."""

    @pytest.fixture
    def monitor(self):
        """Create DriftDetectionMonitor instance for testing."""
        return DriftDetectionMonitor(threshold=0.1)

    def test_monitor_initialization(self):
        """Test DriftDetectionMonitor initialization with default threshold."""
        monitor = DriftDetectionMonitor()
        assert monitor.threshold == 0.1
        assert monitor.drift_history == []

    def test_monitor_initialization_custom_threshold(self):
        """Test DriftDetectionMonitor with custom threshold."""
        monitor = DriftDetectionMonitor(threshold=0.2)
        assert monitor.threshold == 0.2

    def test_detect_data_drift_no_drift(self, monitor):
        """Test data drift detection with no drift."""
        result = monitor.detect_data_drift(
            reference_data_id="train_2024",
            current_data_id="prod_2025_01",
            features=["age", "income"],
        )

        assert result["drift_detected"] is False
        assert result["drift_score"] == 0.08
        assert result["threshold"] == 0.1
        assert result["drifted_features"] == []

    def test_detect_data_drift_features_monitored(self, monitor):
        """Test data drift detection tracks monitored features."""
        features = ["age", "income", "credit_score"]
        result = monitor.detect_data_drift(
            reference_data_id="train_2024",
            current_data_id="prod_2025_01",
            features=features,
        )

        assert result["features_monitored"] == features

    def test_detect_data_drift_timestamp(self, monitor):
        """Test data drift detection includes timestamp."""
        result = monitor.detect_data_drift(
            reference_data_id="train_2024",
            current_data_id="prod_2025_01",
            features=["age"],
        )

        assert "timestamp" in result
        datetime.fromisoformat(result["timestamp"])

    def test_detect_data_drift_history(self, monitor):
        """Test data drift history is tracked."""
        monitor.detect_data_drift(
            reference_data_id="train_2024",
            current_data_id="prod_2025_01",
            features=["age"],
        )
        assert len(monitor.drift_history) == 1

        monitor.detect_data_drift(
            reference_data_id="train_2024",
            current_data_id="prod_2025_02",
            features=["age"],
        )
        assert len(monitor.drift_history) == 2

    def test_detect_model_drift_no_degradation(self, monitor):
        """Test model drift detection with no degradation."""
        baseline = {"accuracy": 0.95, "f1_score": 0.93}
        current = {"accuracy": 0.94, "f1_score": 0.92}

        result = monitor.detect_model_drift(
            model_id="model_001",
            baseline_metrics=baseline,
            current_metrics=current,
        )

        assert result["model_id"] == "model_001"
        assert result["alert"] is False
        assert len(result["degraded_metrics"]) == 0

    def test_detect_model_drift_with_degradation(self, monitor):
        """Test model drift detection with metric degradation."""
        baseline = {"accuracy": 0.95}
        current = {"accuracy": 0.90}  # 5.26% degradation (0.05/0.95)

        result = monitor.detect_model_drift(
            model_id="model_001",
            baseline_metrics=baseline,
            current_metrics=current,
        )

        assert result["alert"] is True  # 5.26% > 5% threshold

    def test_detect_model_drift_significant_degradation(self, monitor):
        """Test model drift detection with significant degradation."""
        baseline = {"accuracy": 0.95, "f1_score": 0.90}
        current = {"accuracy": 0.88, "f1_score": 0.82}  # >5% degradation

        result = monitor.detect_model_drift(
            model_id="model_001",
            baseline_metrics=baseline,
            current_metrics=current,
        )

        assert result["alert"] is True
        assert len(result["degraded_metrics"]) > 0

    def test_detect_model_drift_missing_metrics(self, monitor):
        """Test model drift detection with missing current metrics."""
        baseline = {"accuracy": 0.95, "precision": 0.92}
        current = {"accuracy": 0.90}  # Missing precision

        result = monitor.detect_model_drift(
            model_id="model_001",
            baseline_metrics=baseline,
            current_metrics=current,
        )

        # Should handle missing metrics gracefully
        assert "timestamp" in result

    def test_detect_model_drift_timestamp(self, monitor):
        """Test model drift detection includes timestamp."""
        baseline = {"accuracy": 0.95}
        current = {"accuracy": 0.94}

        result = monitor.detect_model_drift(
            model_id="model_001",
            baseline_metrics=baseline,
            current_metrics=current,
        )

        assert "timestamp" in result
        datetime.fromisoformat(result["timestamp"])

    def test_detect_concept_drift_basic(self, monitor):
        """Test basic concept drift detection."""
        result = monitor.detect_concept_drift(
            model_id="model_001",
            prediction_window="7d",
        )

        assert result["model_id"] == "model_001"
        assert result["prediction_window"] == "7d"
        assert result["detected"] is False

    def test_detect_concept_drift_custom_threshold(self, monitor):
        """Test concept drift detection with custom threshold."""
        result = monitor.detect_concept_drift(
            model_id="model_001",
            prediction_window="14d",
            threshold=0.2,
        )

        assert result["threshold"] == 0.2

    def test_detect_concept_drift_confidence(self, monitor):
        """Test concept drift detection includes confidence score."""
        result = monitor.detect_concept_drift(
            model_id="model_001",
        )

        assert "confidence" in result
        assert 0 <= result["confidence"] <= 1

    def test_detect_concept_drift_explanation(self, monitor):
        """Test concept drift detection includes explanation."""
        result = monitor.detect_concept_drift(
            model_id="model_001",
        )

        assert "explanation" in result

    def test_detect_concept_drift_timestamp(self, monitor):
        """Test concept drift detection includes timestamp."""
        result = monitor.detect_concept_drift(
            model_id="model_001",
        )

        assert "timestamp" in result
        datetime.fromisoformat(result["timestamp"])

    def test_get_drift_metrics(self, monitor):
        """Test drift metrics retrieval."""
        metrics = monitor.get_drift_metrics()

        assert metrics["drift_score"] == 0.08
        assert metrics["threshold"] == 0.1
        assert metrics["status"] == "healthy"
        assert "total_checks" in metrics
        assert "drift_incidents" in metrics

    def test_get_drift_metrics_after_detection(self, monitor):
        """Test drift metrics after detection."""
        monitor.detect_data_drift(
            reference_data_id="train_2024",
            current_data_id="prod_2025_01",
            features=["age"],
        )

        metrics = monitor.get_drift_metrics()
        assert metrics["total_checks"] == 1


# ============================================================================
# PerformanceOptimizer Tests
# ============================================================================


class TestPerformanceOptimizer:
    """Test model performance optimization functionality."""

    @pytest.fixture
    def optimizer(self):
        """Create PerformanceOptimizer instance for testing."""
        return PerformanceOptimizer()

    def test_optimizer_initialization(self, optimizer):
        """Test PerformanceOptimizer initialization."""
        assert optimizer.optimization_history == []

    def test_quantize_model_int8(self, optimizer):
        """Test model quantization to int8."""
        result = optimizer.quantize_model(
            model_path="/models/large_model.pkl",
            precision="int8",
        )

        assert result["precision"] == "int8"
        assert result["original_size_mb"] == 600.0
        assert result["quantized_size_mb"] == 150.0
        assert result["size_reduction"] == 0.75

    def test_quantize_model_int4(self, optimizer):
        """Test model quantization to int4."""
        result = optimizer.quantize_model(
            model_path="/models/model.pkl",
            precision="int4",
        )

        assert result["precision"] == "int4"
        assert result["quantized_size_mb"] == 75.0
        assert result["size_reduction"] == 0.875

    def test_quantize_model_float16(self, optimizer):
        """Test model quantization to float16."""
        result = optimizer.quantize_model(
            model_path="/models/model.pkl",
            precision="float16",
        )

        assert result["precision"] == "float16"
        assert result["quantized_size_mb"] == 300.0
        assert result["size_reduction"] == 0.5

    def test_quantize_model_accuracy_impact(self, optimizer):
        """Test model quantization accuracy impact."""
        result = optimizer.quantize_model(
            model_path="/models/model.pkl",
            precision="int8",
        )

        assert "accuracy_impact" in result
        assert result["accuracy_impact"] == -0.01

    def test_quantize_model_output_path(self, optimizer):
        """Test model quantization output path."""
        result = optimizer.quantize_model(
            model_path="/models/model.pkl",
            precision="int8",
        )

        assert result["quantized_model"] == "/models/model.pkl.quantized"

    def test_prune_model_basic(self, optimizer):
        """Test basic model pruning."""
        result = optimizer.prune_model(
            model_path="/models/model.pkl",
            sparsity_target=0.5,
        )

        assert result["sparsity_ratio"] == 0.5
        assert result["pruned_model"] == "/models/model.pkl.pruned"

    def test_prune_model_parameters_removed(self, optimizer):
        """Test model pruning calculates removed parameters."""
        result = optimizer.prune_model(
            model_path="/models/model.pkl",
            sparsity_target=0.5,
        )

        assert result["parameters_removed"] == 5000000  # 50% of 10M

    def test_prune_model_inference_speedup(self, optimizer):
        """Test model pruning includes inference speedup."""
        result = optimizer.prune_model(
            model_path="/models/model.pkl",
            sparsity_target=0.5,
        )

        assert "inference_speedup" in result
        assert result["inference_speedup"] == 1.5

    def test_prune_model_high_sparsity(self, optimizer):
        """Test model pruning with high sparsity."""
        result = optimizer.prune_model(
            model_path="/models/model.pkl",
            sparsity_target=0.9,
        )

        assert result["parameters_removed"] == 9000000

    def test_distill_model_basic(self, optimizer):
        """Test basic model distillation."""
        result = optimizer.distill_model(
            teacher_model_path="/models/teacher.pkl",
            student_architecture="mobilenet",
        )

        assert result["teacher_model"] == "/models/teacher.pkl"
        assert result["student_architecture"] == "mobilenet"

    def test_distill_model_output(self, optimizer):
        """Test distilled model output path."""
        result = optimizer.distill_model(
            teacher_model_path="/models/teacher.pkl",
            student_architecture="mobilenet",
        )

        assert result["student_model"] == "/models/teacher.pkl.distilled"

    def test_distill_model_size_reduction(self, optimizer):
        """Test distillation size reduction."""
        result = optimizer.distill_model(
            teacher_model_path="/models/teacher.pkl",
            student_architecture="mobilenet",
        )

        assert result["size_reduction"] == 0.9  # 10x smaller

    def test_distill_model_accuracy_retention(self, optimizer):
        """Test distillation accuracy retention."""
        result = optimizer.distill_model(
            teacher_model_path="/models/teacher.pkl",
            student_architecture="mobilenet",
        )

        assert result["accuracy_retention"] == 0.95

    def test_optimize_inference_latency_basic(self, optimizer):
        """Test basic inference latency optimization."""
        strategy = optimizer.optimize_inference_latency(
            model_path="/models/model.pkl",
            target_latency_ms=100.0,
        )

        assert strategy["model_path"] == "/models/model.pkl"
        assert strategy["target_latency_ms"] == 100.0

    def test_optimize_inference_latency_optimizations(self, optimizer):
        """Test inference latency optimization techniques."""
        strategy = optimizer.optimize_inference_latency(
            model_path="/models/model.pkl",
        )

        assert "optimizations" in strategy
        assert "batch_inference" in strategy["optimizations"]
        assert "caching" in strategy["optimizations"]
        assert "gpu_acceleration" in strategy["optimizations"]

    def test_optimize_inference_latency_estimated_latency(self, optimizer):
        """Test inference latency estimated improvement."""
        strategy = optimizer.optimize_inference_latency(
            model_path="/models/model.pkl",
            target_latency_ms=100.0,
        )

        assert strategy["estimated_latency_ms"] == 80.0

    def test_optimize_inference_latency_throughput(self, optimizer):
        """Test inference latency throughput improvement."""
        strategy = optimizer.optimize_inference_latency(
            model_path="/models/model.pkl",
        )

        assert "throughput_improvement" in strategy
        assert strategy["throughput_improvement"] == 2.5


# ============================================================================
# MLOpsMetricsCollector Tests
# ============================================================================


class TestMLOpsMetricsCollector:
    """Test MLOps metrics collection functionality."""

    @pytest.fixture
    def collector(self):
        """Create MLOpsMetricsCollector instance for testing."""
        return MLOpsMetricsCollector()

    def test_collector_initialization(self, collector):
        """Test MLOpsMetricsCollector initialization."""
        assert collector.metrics_history == []

    def test_collect_training_metrics_default(self, collector):
        """Test collecting training metrics with default parameters."""
        metrics = collector.collect_training_metrics()

        assert metrics["run_id"] == "run_001"
        assert metrics["epoch"] == 10
        assert "accuracy" in metrics
        assert "loss" in metrics

    def test_collect_training_metrics_custom(self, collector):
        """Test collecting training metrics with custom parameters."""
        metrics = collector.collect_training_metrics(
            run_id="custom_run",
            epoch=25,
        )

        assert metrics["run_id"] == "custom_run"
        assert metrics["epoch"] == 25

    def test_collect_training_metrics_values(self, collector):
        """Test training metrics values."""
        metrics = collector.collect_training_metrics()

        assert metrics["accuracy"] == 0.94
        assert metrics["loss"] == 0.12
        assert metrics["f1_score"] == 0.93
        assert metrics["precision"] == 0.92
        assert metrics["recall"] == 0.94
        assert metrics["auc_roc"] == 0.96

    def test_collect_training_metrics_training_time(self, collector):
        """Test training metrics includes training time."""
        metrics = collector.collect_training_metrics()

        assert "training_time_seconds" in metrics
        assert metrics["training_time_seconds"] == 3600.0

    def test_collect_training_metrics_timestamp(self, collector):
        """Test training metrics includes timestamp."""
        metrics = collector.collect_training_metrics()

        assert "timestamp" in metrics
        datetime.fromisoformat(metrics["timestamp"])

    def test_collect_training_metrics_history(self, collector):
        """Test training metrics history tracking."""
        collector.collect_training_metrics(run_id="run_001")
        assert len(collector.metrics_history) == 1

        collector.collect_training_metrics(run_id="run_002")
        assert len(collector.metrics_history) == 2

    def test_collect_inference_metrics_default(self, collector):
        """Test collecting inference metrics with default parameters."""
        metrics = collector.collect_inference_metrics()

        assert metrics["model_id"] == "model_001"
        assert "latency_p50_ms" in metrics
        assert "throughput_qps" in metrics

    def test_collect_inference_metrics_custom(self, collector):
        """Test collecting inference metrics with custom model ID."""
        metrics = collector.collect_inference_metrics(model_id="custom_model")

        assert metrics["model_id"] == "custom_model"

    def test_collect_inference_metrics_latency(self, collector):
        """Test inference metrics latency percentiles."""
        metrics = collector.collect_inference_metrics()

        assert metrics["latency_p50_ms"] == 45.0
        assert metrics["latency_p95_ms"] == 85.0
        assert metrics["latency_p99_ms"] == 120.0

    def test_collect_inference_metrics_throughput(self, collector):
        """Test inference metrics throughput."""
        metrics = collector.collect_inference_metrics()

        assert metrics["throughput_qps"] == 500.0

    def test_collect_inference_metrics_error_rate(self, collector):
        """Test inference metrics error rate."""
        metrics = collector.collect_inference_metrics()

        assert metrics["error_rate"] == 0.001
        assert metrics["success_rate"] == 0.999

    def test_collect_inference_metrics_timestamp(self, collector):
        """Test inference metrics includes timestamp."""
        metrics = collector.collect_inference_metrics()

        assert "timestamp" in metrics
        datetime.fromisoformat(metrics["timestamp"])

    def test_track_model_performance_basic(self, collector):
        """Test model performance tracking."""
        timeline = collector.track_model_performance(model_id="model_001")

        assert isinstance(timeline, list)
        assert len(timeline) > 0

    def test_track_model_performance_default_window(self, collector):
        """Test model performance tracking with default time window."""
        timeline = collector.track_model_performance(model_id="model_001")

        # Default window is 7d
        for entry in timeline:
            assert "timestamp" in entry
            assert "accuracy" in entry
            assert "latency_ms" in entry

    def test_track_model_performance_custom_window(self, collector):
        """Test model performance tracking with custom time window."""
        timeline = collector.track_model_performance(
            model_id="model_001",
            time_window="30d",
        )

        assert isinstance(timeline, list)

    def test_track_model_performance_entries(self, collector):
        """Test model performance timeline entries."""
        timeline = collector.track_model_performance(model_id="model_001")

        assert timeline[0]["accuracy"] == 0.94
        assert timeline[0]["latency_ms"] == 50.0
        assert timeline[1]["accuracy"] == 0.93
        assert timeline[1]["latency_ms"] == 55.0

    def test_get_mlops_health_status(self, collector):
        """Test MLOps health status retrieval."""
        status = collector.get_mlops_health_status()

        assert "overall_score" in status
        assert "status" in status
        assert "components" in status

    def test_get_mlops_health_status_overall(self, collector):
        """Test MLOps health status overall score."""
        status = collector.get_mlops_health_status()

        assert status["overall_score"] == 0.95
        assert status["status"] == "healthy"

    def test_get_mlops_health_status_components(self, collector):
        """Test MLOps health status components."""
        status = collector.get_mlops_health_status()

        components = status["components"]
        assert "data_pipeline" in components
        assert "model_training" in components
        assert "model_serving" in components
        assert "monitoring" in components

    def test_get_mlops_health_status_component_details(self, collector):
        """Test MLOps health status component details."""
        status = collector.get_mlops_health_status()

        data_pipeline = status["components"]["data_pipeline"]
        assert data_pipeline["status"] == "healthy"
        assert data_pipeline["score"] == 0.98

    def test_get_mlops_health_status_alerts(self, collector):
        """Test MLOps health status alerts."""
        status = collector.get_mlops_health_status()

        assert "alerts" in status
        assert isinstance(status["alerts"], list)

    def test_get_mlops_health_status_recommendations(self, collector):
        """Test MLOps health status recommendations."""
        status = collector.get_mlops_health_status()

        assert "recommendations" in status
        assert isinstance(status["recommendations"], list)

    def test_get_mlops_health_status_timestamp(self, collector):
        """Test MLOps health status includes timestamp."""
        status = collector.get_mlops_health_status()

        assert "timestamp" in status
        datetime.fromisoformat(status["timestamp"])


# ============================================================================
# Integration Tests
# ============================================================================


class TestMLOpsIntegration:
    """Integration tests for MLOps module."""

    def test_full_mlops_workflow(self):
        """Test complete MLOps workflow."""
        # Pipeline orchestration
        orchestrator = MLPipelineOrchestrator()
        mlflow_config = orchestrator.orchestrate_mlflow_pipeline(
            experiment_name="fraud_detection",
            run_name="exp_001",
            tracking_uri="http://mlflow:5000",
        )
        assert mlflow_config["experiment_name"] == "fraud_detection"

        # Model registration
        manager = ModelVersionManager()
        version = manager.register_model_version(
            model_name="fraud_classifier",
            version="v1.0.0",
            registry_uri="s3://models/fraud_classifier",
            metadata={"accuracy": 0.95},
        )
        assert version["model_name"] == "fraud_classifier"

        # Model deployment
        planner = ModelDeploymentPlanner()
        deployment = planner.plan_ray_serve_deployment(
            model_name="fraud_classifier",
            replicas=3,
            route_prefix="/predict",
        )
        assert deployment["type"] == "ray_serve"

        # Monitoring
        monitor = DriftDetectionMonitor()
        drift_result = monitor.detect_data_drift(
            reference_data_id="train_2024",
            current_data_id="prod_2025",
            features=["transaction_amount", "frequency"],
        )
        assert "drift_detected" in drift_result

    def test_data_to_deployment_pipeline(self):
        """Test data pipeline to deployment workflow."""
        # Data pipeline
        builder = DataPipelineBuilder()
        pipeline = builder.build_feature_pipeline(
            features=["age", "income", "credit_score"],
            transformations=["normalize", "encode_categorical"],
        )
        assert len(pipeline["pipeline_steps"]) == 2

        # Validate quality
        validation = builder.validate_data_quality(
            dataset_id="dataset_001",
            checks=["missing_values", "schema_compliance"],
        )
        assert validation["quality_score"] == 1.0

        # Deploy model
        planner = ModelDeploymentPlanner()
        endpoint = planner.create_inference_endpoint(
            model_id="fraud_classifier",
            endpoint_name="fraud_api",
            auth_enabled=True,
        )
        assert endpoint["auth_token"] is not None

    def test_monitoring_and_optimization_workflow(self):
        """Test monitoring and optimization workflow."""
        # Monitor drift
        monitor = DriftDetectionMonitor()
        model_drift = monitor.detect_model_drift(
            model_id="classifier",
            baseline_metrics={"accuracy": 0.95},
            current_metrics={"accuracy": 0.93},
        )
        assert model_drift["alert"] is False

        # Optimize performance
        optimizer = PerformanceOptimizer()
        quantized = optimizer.quantize_model(
            model_path="/models/classifier.pkl",
            precision="int8",
        )
        assert quantized["size_reduction"] == 0.75

        # Collect metrics
        collector = MLOpsMetricsCollector()
        health = collector.get_mlops_health_status()
        assert health["status"] == "healthy"
