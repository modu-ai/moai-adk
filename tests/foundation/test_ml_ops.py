"""
Comprehensive DDD tests for ml_ops.py module.
Tests cover all 7 classes.
"""

from datetime import UTC, datetime

from moai_adk.foundation.ml_ops import (
    DataPipelineBuilder,
    DriftDetectionMonitor,
    MLOpsMetricsCollector,
    MLPipelineOrchestrator,
    ModelDeploymentPlanner,
    ModelVersionManager,
    PerformanceOptimizer,
)

# ============================================================================
# Test MLPipelineOrchestrator
# ============================================================================


class TestMLPipelineOrchestrator:
    """Test suite for MLPipelineOrchestrator class."""

    def test_initialization(self):
        """Test orchestrator initialization."""
        orchestrator = MLPipelineOrchestrator()
        assert "mlflow" in orchestrator.supported_platforms
        assert orchestrator.execution_history == []

    def test_orchestrate_mlflow_pipeline(self, mlflow_config):
        """Test MLflow pipeline orchestration."""
        orchestrator = MLPipelineOrchestrator()
        config = orchestrator.orchestrate_mlflow_pipeline(
            experiment_name="test_experiment",
            run_name="test_run_001",
            tracking_uri="http://localhost:5000",
            tags={"framework": "pytorch"},
        )

        assert config["type"] == "mlflow"
        assert config["experiment_name"] == "test_experiment"
        assert config["tracking_uri"] == "http://localhost:5000"
        assert "created_at" in config

    def test_orchestrate_kubeflow_pipeline(self):
        """Test Kubeflow pipeline orchestration."""
        orchestrator = MLPipelineOrchestrator()
        spec = orchestrator.orchestrate_kubeflow_pipeline(
            pipeline_name="test_pipeline", namespace="kubeflow", components=["preprocess", "train", "evaluate"]
        )

        assert spec["type"] == "kubeflow"
        assert spec["pipeline_name"] == "test_pipeline"
        assert spec["namespace"] == "kubeflow"
        assert len(spec["components"]) == 3

    def test_orchestrate_airflow_dags(self):
        """Test Airflow DAG orchestration."""
        orchestrator = MLPipelineOrchestrator()
        dag_config = orchestrator.orchestrate_airflow_dags(
            dag_id="test_dag", schedule_interval="0 0 * * *", tasks=["task1", "task2", "task3"]
        )

        assert dag_config["type"] == "airflow"
        assert dag_config["dag_id"] == "test_dag"
        assert dag_config["schedule_interval"] == "0 0 * * *"
        assert len(dag_config["tasks"]) == 3

    def test_orchestrate_airflow_dags_with_dependencies(self):
        """Test Airflow DAG with custom dependencies."""
        orchestrator = MLPipelineOrchestrator()
        dag_config = orchestrator.orchestrate_airflow_dags(
            dag_id="test_dag", schedule_interval="daily", tasks=["task1", "task2"], dependencies={"task1": ["task2"]}
        )

        assert dag_config["dependencies"]["task1"] == ["task2"]

    def test_track_pipeline_execution(self):
        """Test pipeline execution tracking."""
        orchestrator = MLPipelineOrchestrator()
        start_time = datetime.now(UTC).isoformat()
        metrics = orchestrator.track_pipeline_execution(
            pipeline_id="pipeline_001", status="completed", start_time=start_time
        )

        assert metrics["pipeline_id"] == "pipeline_001"
        assert metrics["status"] == "completed"
        assert "execution_metrics" in metrics
        assert len(orchestrator.execution_history) == 1


# ============================================================================
# Test ModelVersionManager
# ============================================================================


class TestModelVersionManager:
    """Test suite for ModelVersionManager class."""

    def test_initialization(self):
        """Test manager initialization."""
        manager = ModelVersionManager()
        assert manager.registry == {}
        assert manager.lineage_graph == {}

    def test_register_model_version(self):
        """Test model version registration."""
        manager = ModelVersionManager()
        result = manager.register_model_version(
            model_name="sentiment_model",
            version="v1.0.0",
            registry_uri="s3://models/registry",
            metadata={"framework": "pytorch", "accuracy": 0.94},
        )

        assert result["model_name"] == "sentiment_model"
        assert result["version"] == "v1.0.0"
        assert result["stage"] == "staging"
        assert "version_id" in result
        assert "created_at" in result

    def test_track_model_lineage(self):
        """Test model lineage tracking."""
        manager = ModelVersionManager()
        lineage = manager.track_model_lineage(
            model_id="model_001", include_data_sources=True, include_training_runs=True
        )

        assert lineage["model_id"] == "model_001"
        assert len(lineage["data_sources"]) > 0
        assert len(lineage["training_runs"]) > 0
        assert "lineage_graph" in lineage

    def test_track_model_lineage_no_data(self):
        """Test lineage tracking without data sources."""
        manager = ModelVersionManager()
        lineage = manager.track_model_lineage(
            model_id="model_001", include_data_sources=False, include_training_runs=False
        )

        assert lineage["data_sources"] is None
        assert lineage["training_runs"] is None

    def test_manage_model_artifacts(self):
        """Test model artifact management."""
        manager = ModelVersionManager()
        artifacts = manager.manage_model_artifacts(
            model_id="model_001", artifact_types=["weights", "config", "metadata"], storage_backend="s3"
        )

        assert artifacts["storage_backend"] == "s3"
        assert "weights" in artifacts
        assert "config" in artifacts
        assert "metadata" in artifacts
        assert artifacts["weights"]["path"] == "s3://models/model_001/weights.pkl"


# ============================================================================
# Test DataPipelineBuilder
# ============================================================================


class TestDataPipelineBuilder:
    """Test suite for DataPipelineBuilder class."""

    def test_initialization(self):
        """Test builder initialization."""
        builder = DataPipelineBuilder()
        assert builder.pipelines == {}

    def test_build_feature_pipeline(self):
        """Test feature pipeline building."""
        builder = DataPipelineBuilder()
        pipeline = builder.build_feature_pipeline(
            features=["age", "income", "credit_score"], transformations=["normalize", "encode_categorical"]
        )

        assert pipeline["features"] == ["age", "income", "credit_score"]
        assert pipeline["transformations"] == ["normalize", "encode_categorical"]
        assert len(pipeline["pipeline_steps"]) == 2

    def test_build_feature_pipeline_with_interactions(self):
        """Test feature pipeline with interaction features."""
        builder = DataPipelineBuilder()
        pipeline = builder.build_feature_pipeline(features=["a", "b"], transformations=["create_interactions"])

        assert len(pipeline["pipeline_steps"]) == 1
        # Should create interaction features
        assert "a_b" in pipeline["pipeline_steps"][0]["output_features"]

    def test_validate_data_quality(self):
        """Test data quality validation."""
        builder = DataPipelineBuilder()
        result = builder.validate_data_quality(
            dataset_id="test_dataset", checks=["missing_values", "schema_compliance", "outliers"]
        )

        assert result["dataset_id"] == "test_dataset"
        assert result["quality_score"] > 0
        assert len(result["passed_checks"]) == 2
        assert len(result["violations"]) == 1

    def test_handle_missing_values_drop(self):
        """Test missing value handling with drop strategy."""
        builder = DataPipelineBuilder()
        strategy = builder.handle_missing_values(column_name="test_col", missing_ratio=0.6, data_type="numeric")

        assert strategy["column_name"] == "test_col"
        assert strategy["method"] == "drop"

    def test_handle_missing_values_median(self):
        """Test missing value handling with median strategy."""
        builder = DataPipelineBuilder()
        strategy = builder.handle_missing_values(column_name="test_col", missing_ratio=0.15, data_type="numeric")

        assert strategy["method"] == "median"

    def test_handle_missing_values_mode(self):
        """Test missing value handling with mode strategy."""
        builder = DataPipelineBuilder()
        strategy = builder.handle_missing_values(column_name="test_col", missing_ratio=0.3, data_type="categorical")

        assert strategy["method"] == "mode"

    def test_get_data_statistics(self):
        """Test data statistics retrieval."""
        builder = DataPipelineBuilder()
        stats = builder.get_data_statistics(dataset_id="test_dataset")

        assert stats["dataset_id"] == "test_dataset"
        assert stats["row_count"] == 100000
        assert stats["column_count"] == 25


# ============================================================================
# Test ModelDeploymentPlanner
# ============================================================================


class TestModelDeploymentPlanner:
    """Test suite for ModelDeploymentPlanner class."""

    def test_initialization(self):
        """Test planner initialization."""
        planner = ModelDeploymentPlanner()
        assert planner.deployments == {}

    def test_plan_ray_serve_deployment(self):
        """Test Ray Serve deployment planning."""
        planner = ModelDeploymentPlanner()
        config = planner.plan_ray_serve_deployment(model_name="fraud_detector", replicas=3, route_prefix="/predict")

        assert config["type"] == "ray_serve"
        assert config["model_name"] == "fraud_detector"
        assert config["replicas"] == 3
        assert config["route_prefix"] == "/predict"

    def test_plan_ray_serve_deployment_with_autoscaling(self):
        """Test Ray Serve with autoscaling."""
        planner = ModelDeploymentPlanner()
        config = planner.plan_ray_serve_deployment(
            model_name="test_model", replicas=2, route_prefix="/test", autoscaling=True
        )

        assert config["autoscaling"]["enabled"] is True
        assert config["autoscaling"]["min_replicas"] == 1

    def test_plan_kserve_deployment(self):
        """Test KServe deployment planning."""
        planner = ModelDeploymentPlanner()
        spec = planner.plan_kserve_deployment(
            model_name="sentiment_model", framework="pytorch", storage_uri="s3://models/sentiment"
        )

        assert spec["type"] == "kserve"
        assert spec["framework"] == "pytorch"
        assert spec["storage_uri"] == "s3://models/sentiment"

    def test_containerize_model(self):
        """Test model containerization."""
        planner = ModelDeploymentPlanner()
        container = planner.containerize_model(
            model_path="/models/model.pkl", base_image="python:3.11", requirements=["torch", "numpy"]
        )

        assert container["base_image"] == "python:3.11"
        assert "requirements" in container
        assert "dockerfile_content" in container
        assert "FROM python:3.11" in container["dockerfile_content"]

    def test_create_inference_endpoint(self):
        """Test inference endpoint creation."""
        planner = ModelDeploymentPlanner()
        endpoint = planner.create_inference_endpoint(
            model_id="model_001", endpoint_name="predict-endpoint", auth_enabled=True
        )

        assert endpoint["model_id"] == "model_001"
        assert endpoint["endpoint_name"] == "predict-endpoint"
        assert endpoint["auth_enabled"] is True
        assert endpoint["auth_token"] is not None
        assert endpoint["methods"] == ["POST"]


# ============================================================================
# Test DriftDetectionMonitor
# ============================================================================


class TestDriftDetectionMonitor:
    """Test suite for DriftDetectionMonitor class."""

    def test_initialization(self):
        """Test monitor initialization."""
        monitor = DriftDetectionMonitor(threshold=0.1)
        assert monitor.threshold == 0.1
        assert monitor.drift_history == []

    def test_detect_data_drift_no_drift(self):
        """Test data drift detection without drift."""
        monitor = DriftDetectionMonitor()
        result = monitor.detect_data_drift(
            reference_data_id="train_2024", current_data_id="prod_2025_01", features=["age", "income"]
        )

        assert result["drift_detected"] is False
        assert result["drift_score"] < monitor.threshold
        assert result["threshold"] == 0.1

    def test_detect_model_drift_no_drift(self):
        """Test model drift detection without drift."""
        monitor = DriftDetectionMonitor()
        baseline = {"accuracy": 0.95, "precision": 0.93}
        current = {"accuracy": 0.94, "precision": 0.92}

        result = monitor.detect_model_drift(model_id="model_001", baseline_metrics=baseline, current_metrics=current)

        assert result["model_id"] == "model_001"
        assert result["alert"] is False

    def test_detect_model_drift_with_drift(self):
        """Test model drift detection with drift."""
        monitor = DriftDetectionMonitor()
        baseline = {"accuracy": 0.95, "precision": 0.93}
        current = {"accuracy": 0.85, "precision": 0.80}  # 10%+ degradation

        result = monitor.detect_model_drift(model_id="model_001", baseline_metrics=baseline, current_metrics=current)

        assert result["alert"] is True
        assert len(result["degraded_metrics"]) == 2

    def test_detect_concept_drift(self):
        """Test concept drift detection."""
        monitor = DriftDetectionMonitor()
        result = monitor.detect_concept_drift(model_id="model_001", prediction_window="7d", threshold=0.15)

        assert "detected" in result
        assert result["model_id"] == "model_001"
        assert result["prediction_window"] == "7d"

    def test_get_drift_metrics(self):
        """Test drift metrics retrieval."""
        monitor = DriftDetectionMonitor(threshold=0.05)
        monitor.detect_data_drift("ref", "cur", ["f1"])
        monitor.detect_data_drift("ref", "cur", ["f2"])

        metrics = monitor.get_drift_metrics()

        assert metrics["threshold"] == 0.05
        assert metrics["total_checks"] == 2
        assert metrics["status"] in ["healthy", "warning", "critical"]


# ============================================================================
# Test PerformanceOptimizer
# ============================================================================


class TestPerformanceOptimizer:
    """Test suite for PerformanceOptimizer class."""

    def test_initialization(self):
        """Test optimizer initialization."""
        optimizer = PerformanceOptimizer()
        assert optimizer.optimization_history == []

    def test_quantize_model_int8(self):
        """Test model quantization to int8."""
        optimizer = PerformanceOptimizer()
        result = optimizer.quantize_model(model_path="/models/model.pkl", precision="int8")

        assert result["precision"] == "int8"
        assert result["original_size_mb"] == 600.0
        assert result["quantized_size_mb"] == 150.0  # 4x reduction
        assert result["size_reduction"] == 0.75

    def test_quantize_model_int4(self):
        """Test model quantization to int4."""
        optimizer = PerformanceOptimizer()
        result = optimizer.quantize_model(model_path="/models/model.pkl", precision="int4")

        assert result["precision"] == "int4"
        assert result["quantized_size_mb"] == 75.0  # 8x reduction

    def test_prune_model(self):
        """Test model pruning."""
        optimizer = PerformanceOptimizer()
        result = optimizer.prune_model(model_path="/models/model.pkl", sparsity_target=0.5)

        assert result["sparsity_ratio"] == 0.5
        assert result["parameters_removed"] == 5000000
        assert result["inference_speedup"] == 1.5

    def test_distill_model(self):
        """Test model distillation."""
        optimizer = PerformanceOptimizer()
        result = optimizer.distill_model(teacher_model_path="/models/teacher.pkl", student_architecture="mobilenet")

        assert result["teacher_model"] == "/models/teacher.pkl"
        assert result["student_architecture"] == "mobilenet"
        assert result["size_reduction"] == 0.9
        assert result["accuracy_retention"] == 0.95

    def test_optimize_inference_latency(self):
        """Test inference latency optimization."""
        optimizer = PerformanceOptimizer()
        strategy = optimizer.optimize_inference_latency(model_path="/models/model.pkl", target_latency_ms=100.0)

        assert strategy["model_path"] == "/models/model.pkl"
        assert strategy["target_latency_ms"] == 100.0
        assert len(strategy["optimizations"]) > 0
        assert strategy["estimated_latency_ms"] < 100.0


# ============================================================================
# Test MLOpsMetricsCollector
# ============================================================================


class TestMLOpsMetricsCollector:
    """Test suite for MLOpsMetricsCollector class."""

    def test_initialization(self):
        """Test collector initialization."""
        collector = MLOpsMetricsCollector()
        assert collector.metrics_history == []

    def test_collect_training_metrics(self):
        """Test training metrics collection."""
        collector = MLOpsMetricsCollector()
        metrics = collector.collect_training_metrics(run_id="run_001", epoch=10)

        assert metrics["run_id"] == "run_001"
        assert metrics["epoch"] == 10
        assert metrics["accuracy"] == 0.94
        assert metrics["loss"] == 0.12
        assert "timestamp" in metrics

    def test_collect_inference_metrics(self):
        """Test inference metrics collection."""
        collector = MLOpsMetricsCollector()
        metrics = collector.collect_inference_metrics(model_id="model_001")

        assert metrics["model_id"] == "model_001"
        assert metrics["latency_p50_ms"] == 45.0
        assert metrics["throughput_qps"] == 500.0
        assert "timestamp" in metrics

    def test_track_model_performance(self):
        """Test model performance tracking."""
        collector = MLOpsMetricsCollector()
        timeline = collector.track_model_performance(model_id="model_001", time_window="7d")

        assert isinstance(timeline, list)
        assert len(timeline) > 0
        assert "timestamp" in timeline[0]
        assert "accuracy" in timeline[0]

    def test_get_mlops_health_status(self):
        """Test MLOps health status."""
        collector = MLOpsMetricsCollector()
        health = collector.get_mlops_health_status()

        assert health["status"] in ["healthy", "warning", "unhealthy"]
        assert "components" in health
        assert "data_pipeline" in health["components"]
        assert "model_training" in health["components"]
        assert health["overall_score"] >= 0
        assert health["overall_score"] <= 1.0
